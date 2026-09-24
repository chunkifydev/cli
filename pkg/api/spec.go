package api

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/chunkifydev/cli/pkg/config"
)

const productionSpecURL = "https://chunkify.dev/docs/openapi.json"
const specCacheAge = time.Hour
const maxSpecBytes = 5 << 20

var pathParameter = regexp.MustCompile(`\{([^}]+)\}`)

type openAPIDocument struct {
	Paths map[string]map[string]json.RawMessage `json:"paths"`
}

type openAPIOperation struct {
	OperationID string          `json:"operationId"`
	Security    json.RawMessage `json:"security"`
	Parameters  []struct {
		Name string `json:"name"`
		In   string `json:"in"`
	} `json:"parameters"`
	RequestBody *struct {
		Required bool                       `json:"required"`
		Content  map[string]json.RawMessage `json:"content"`
	} `json:"requestBody"`
}

func loadConfiguredOperations(ctx context.Context, cfg *config.Config) ([]operation, error) {
	specURL, err := cfg.OpenAPIURL()
	if err != nil {
		return nil, err
	}
	if specURL == "" {
		specURL = productionSpecURL
	}
	cacheRoot, err := os.UserCacheDir()
	if err != nil {
		cacheRoot = os.TempDir()
	}
	return loadOperations(ctx, specURL, filepath.Join(cacheRoot, "chunkify-cli"), &http.Client{Timeout: 10 * time.Second})
}

func loadOperations(ctx context.Context, specURL, cacheDir string, client *http.Client) ([]operation, error) {
	key := sha256.Sum256([]byte(specURL))
	cachePath := filepath.Join(cacheDir, fmt.Sprintf("openapi-%x.json", key[:8]))
	cached, cacheErr := os.ReadFile(cachePath)
	var cachedOperations []operation
	if cacheErr == nil {
		cachedOperations, cacheErr = parseOperations(cached)
	}
	if cacheErr == nil {
		if info, err := os.Stat(cachePath); err == nil && time.Since(info.ModTime()) < specCacheAge {
			return cachedOperations, nil
		}
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, specURL, nil)
	if err == nil {
		request.Header.Set("Accept", "application/json")
		response, requestErr := client.Do(request)
		if requestErr != nil {
			err = requestErr
		} else {
			defer response.Body.Close()
			if response.StatusCode != http.StatusOK {
				err = fmt.Errorf("specification returned HTTP %d", response.StatusCode)
			} else {
				var document []byte
				document, err = io.ReadAll(io.LimitReader(response.Body, maxSpecBytes+1))
				if err == nil && len(document) > maxSpecBytes {
					err = fmt.Errorf("specification exceeds %d bytes", maxSpecBytes)
				}
				if err == nil {
					var operations []operation
					operations, err = parseOperations(document)
					if err == nil {
						cacheSpec(cacheDir, cachePath, document)
						return operations, nil
					}
				}
			}
		}
	}
	if cacheErr == nil {
		return cachedOperations, nil
	}
	return nil, fmt.Errorf("cannot load Chunkify API definition from %s: %w", specURL, err)
}

func cacheSpec(dir, path string, data []byte) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return
	}
	file, err := os.CreateTemp(dir, "openapi-*.tmp")
	if err != nil {
		return
	}
	defer os.Remove(file.Name())
	if err := file.Chmod(0600); err != nil {
		file.Close()
		return
	}
	if _, err := file.Write(data); err != nil {
		file.Close()
		return
	}
	if err := file.Close(); err != nil {
		return
	}
	_ = os.Rename(file.Name(), path)
}

func parseOperations(data []byte) ([]operation, error) {
	var document openAPIDocument
	if err := json.Unmarshal(data, &document); err != nil {
		return nil, fmt.Errorf("invalid OpenAPI JSON: %w", err)
	}
	if len(document.Paths) == 0 {
		return nil, fmt.Errorf("OpenAPI document has no paths")
	}
	var operations []operation
	seen := map[string]bool{}
	for path, methods := range document.Paths {
		for method, raw := range methods {
			if method != "get" && method != "post" && method != "put" && method != "patch" && method != "delete" {
				continue
			}
			var spec openAPIOperation
			if err := json.Unmarshal(raw, &spec); err != nil {
				return nil, fmt.Errorf("invalid %s %s operation: %w", method, path, err)
			}
			auth, err := operationAuth(spec.Security)
			if err != nil {
				return nil, fmt.Errorf("%s %s: %w", method, path, err)
			}
			if spec.RequestBody != nil && len(spec.RequestBody.Content) > 0 && spec.RequestBody.Content["application/json"] == nil {
				return nil, fmt.Errorf("%s %s has an unsupported request body", method, path)
			}
			resource, action, err := commandName(spec.OperationID, method, path)
			if err != nil {
				return nil, fmt.Errorf("%s %s: %w", method, path, err)
			}
			key := resource + " " + action
			if seen[key] {
				return nil, fmt.Errorf("duplicate API command %s", key)
			}
			seen[key] = true
			matches := pathParameter.FindAllStringSubmatch(path, -1)
			pathParams := make([]string, 0, len(matches))
			for _, match := range matches {
				pathParams = append(pathParams, match[1])
			}
			op := operation{
				Resource:   resource,
				Action:     action,
				Method:     strings.ToUpper(method),
				Path:       strings.TrimPrefix(path, "/"),
				Auth:       auth,
				PathParams: pathParams,
			}
			for _, parameter := range spec.Parameters {
				if parameter.In == "query" {
					op.QueryParams = append(op.QueryParams, parameter.Name)
				}
			}
			if spec.RequestBody != nil {
				op.Body = len(spec.RequestBody.Content) > 0
				op.BodyRequired = spec.RequestBody.Required
			}
			operations = append(operations, op)
		}
	}
	sort.Slice(operations, func(i, j int) bool {
		if operations[i].Resource != operations[j].Resource {
			return operations[i].Resource < operations[j].Resource
		}
		return operations[i].Action < operations[j].Action
	})
	return operations, nil
}

func operationAuth(raw json.RawMessage) (string, error) {
	if len(raw) == 0 {
		return "", fmt.Errorf("missing security definition")
	}
	var entries []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &entries); err != nil {
		return "", err
	}
	if len(entries) == 0 {
		return "none", nil
	}
	if len(entries) != 1 || len(entries[0]) != 1 {
		return "", fmt.Errorf("unsupported security definition")
	}
	for name := range entries[0] {
		switch name {
		case "ProjectAccessToken":
			return "project", nil
		case "TeamAccessToken":
			return "team", nil
		}
	}
	return "", fmt.Errorf("unsupported security definition")
}

// OpenAPI operation IDs follow <verb><resource>, such as updateProject or
// getJobFiles. The resource name is plural kebab-case. A GET on a collection
// path is called list even when its operation ID starts with get.
func commandName(operationID, method, path string) (string, string, error) {
	boundary := strings.IndexFunc(operationID, unicode.IsUpper)
	if boundary <= 0 || boundary == len(operationID)-1 {
		return "", "", fmt.Errorf("operationId %q must use <verb><Resource> naming", operationID)
	}
	verb := operationID[:boundary]
	words := splitPascalCase(operationID[boundary:])
	words[len(words)-1] = pluralize(words[len(words)-1])
	resource := strings.Join(words, "-")
	action := verb
	if verb == "get" && method == "get" && isCollectionPath(path, resource) {
		action = "list"
	}
	return resource, action, nil
}

func splitPascalCase(value string) []string {
	runes := []rune(value)
	if len(runes) == 0 {
		return nil
	}
	var words []string
	start := 0
	for i := 1; i < len(runes); i++ {
		if unicode.IsUpper(runes[i]) && (!unicode.IsUpper(runes[i-1]) || (i+1 < len(runes) && unicode.IsLower(runes[i+1]))) {
			words = append(words, strings.ToLower(string(runes[start:i])))
			start = i
		}
	}
	words = append(words, strings.ToLower(string(runes[start:])))
	return words
}

func pluralize(word string) string {
	if strings.HasSuffix(word, "us") {
		return word + "es"
	}
	if strings.HasSuffix(word, "s") {
		return word
	}
	if strings.HasSuffix(word, "y") && len(word) > 1 && !strings.ContainsRune("aeiou", rune(word[len(word)-2])) {
		return word[:len(word)-1] + "ies"
	}
	if strings.HasSuffix(word, "ch") || strings.HasSuffix(word, "sh") || strings.HasSuffix(word, "x") || strings.HasSuffix(word, "z") {
		return word + "es"
	}
	return word + "s"
}

func isCollectionPath(path, resource string) bool {
	last := path[strings.LastIndex(path, "/")+1:]
	return last == resource || strings.HasSuffix(resource, "-"+last)
}
