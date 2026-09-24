package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/chunkifydev/chunkify-go/option"
	"github.com/chunkifydev/cli/pkg/config"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

type operation struct {
	Resource     string
	Action       string
	Method       string
	Path         string
	Auth         string
	PathParams   []string
	QueryParams  []string
	Body         bool
	BodyRequired bool
}

func NewCommand(cfg *config.Config) *cobra.Command {
	return newCommand(cfg, func(ctx context.Context) ([]operation, error) {
		return loadConfiguredOperations(ctx, cfg)
	})
}

func newCommand(cfg *config.Config, load func(context.Context) ([]operation, error)) *cobra.Command {
	return &cobra.Command{
		Use:                "api",
		Short:              "Call the Chunkify API and print JSON",
		Long:               "Call any supported Chunkify API resource. Responses are JSON by default and can be piped to jq.",
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			operations, err := load(cmd.Context())
			if err != nil {
				return err
			}
			api := commandForOperations(cfg, operations)
			api.SetArgs(append([]string{"api"}, args...))
			api.SetOut(cmd.OutOrStdout())
			api.SetErr(cmd.ErrOrStderr())
			api.SetIn(cmd.InOrStdin())
			api.SetContext(cmd.Context())
			return api.Execute()
		},
	}
}

func commandForOperations(cfg *config.Config, operations []operation) *cobra.Command {
	root := &cobra.Command{
		Use:           "chunkify",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	cmd := &cobra.Command{
		Use:           "api",
		Short:         "Call the Chunkify API and print JSON",
		Long:          "Call any supported Chunkify API resource. Responses are JSON by default and can be piped to jq.",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.PersistentFlags().StringVar(&cfg.Profile, "profile", cfg.Profile, "Use a specific profile")
	root.AddCommand(cmd)
	resources := map[string]*cobra.Command{}
	for _, op := range operations {
		resource := resources[op.Resource]
		if resource == nil {
			resource = &cobra.Command{Use: op.Resource, Short: "API operations for " + op.Resource}
			resources[op.Resource] = resource
			cmd.AddCommand(resource)
		}
		resource.AddCommand(newActionCommand(cfg, op))
	}
	return root
}

func newActionCommand(cfg *config.Config, op operation) *cobra.Command {
	var dataArg string
	var queryArgs []string
	use := op.Action
	for _, name := range op.PathParams {
		use += " <" + name + ">"
	}
	cmd := &cobra.Command{
		Use:   use,
		Short: op.Method + " /" + op.Path,
		Args:  cobra.ExactArgs(len(op.PathParams)),
		RunE: func(cmd *cobra.Command, args []string) error {
			return execute(cmd, cfg, op, args, dataArg, queryArgs)
		},
	}
	if op.Body {
		cmd.Flags().StringVar(&dataArg, "data", "", "JSON body, @file.json, or @- for standard input")
	}
	if len(op.QueryParams) > 0 {
		cmd.Flags().StringArrayVar(&queryArgs, "query", nil, "Query parameter as key=value; may be repeated")
	}
	return cmd
}

func execute(cmd *cobra.Command, cfg *config.Config, op operation, args []string, dataArg string, queryArgs []string) error {
	if cfg.Client == nil {
		return errors.New("API client is not initialized")
	}
	requestPath, err := fillPath(op.Path, op.PathParams, args)
	if err != nil {
		return err
	}
	query := url.Values{}
	for _, item := range queryArgs {
		key, value, ok := strings.Cut(item, "=")
		if !ok || key == "" {
			return fmt.Errorf("invalid --query %q: expected key=value", item)
		}
		allowed := false
		for _, name := range op.QueryParams {
			if name == key {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("query parameter %q is not supported by %s %s", key, op.Resource, op.Action)
		}
		query.Add(key, value)
	}
	if len(query) > 0 {
		requestPath += "?" + query.Encode()
	}
	var body any
	if cmd.Flags().Changed("data") {
		if !op.Body {
			return fmt.Errorf("%s %s does not accept a JSON body", op.Resource, op.Action)
		}
		value, err := readData(cmd, dataArg)
		if err != nil {
			return err
		}
		body = value
	} else if op.BodyRequired {
		return fmt.Errorf("%s %s requires --data with a JSON body", op.Resource, op.Action)
	}
	var token string
	switch op.Auth {
	case "project":
		if cfg.Token == "" {
			if err := cfg.SetToken(); err != nil {
				if errors.Is(err, keyring.ErrNotFound) {
					return errors.New("project token missing: run `chunkify config token <sk_project_token>` or set CHUNKIFY_TOKEN")
				}
				return fmt.Errorf("cannot load project token: %w", err)
			}
		}
		token = cfg.Token
	case "team":
		if cfg.TeamToken == "" {
			if err := cfg.SetTeamToken(); err != nil {
				if errors.Is(err, keyring.ErrNotFound) {
					return errors.New("team token missing: run `chunkify config team-token <sk_team_token>` or set CHUNKIFY_TEAM_TOKEN")
				}
				return fmt.Errorf("cannot load team token: %w", err)
			}
		}
		token = cfg.TeamToken
	case "none":
	default:
		return fmt.Errorf("unsupported authentication type %q", op.Auth)
	}
	requestOptions := []option.RequestOption{}
	if token != "" {
		requestOptions = append(requestOptions, option.WithHeader("Authorization", "Bearer "+token))
	}
	var response []byte
	if err := cfg.Client.Execute(cmd.Context(), op.Method, requestPath, body, &response, requestOptions...); err != nil {
		return err
	}
	return printResponse(cmd.OutOrStdout(), response)
}

func fillPath(template string, names, args []string) (string, error) {
	if len(names) != len(args) {
		return "", errors.New("wrong number of path arguments")
	}
	for i, name := range names {
		value := args[i]
		if value == "" {
			return "", fmt.Errorf("%s must not be empty", name)
		}
		if strings.Contains(value, "/") && name != "objectPath" {
			return "", fmt.Errorf("%s must not contain '/'", name)
		}
		segments := strings.Split(value, "/")
		for j, segment := range segments {
			if segment == "." || segment == ".." || segment == "" {
				return "", fmt.Errorf("%s contains an invalid path segment", name)
			}
			segments[j] = url.PathEscape(segment)
		}
		template = strings.Replace(template, "{"+name+"}", strings.Join(segments, "/"), 1)
	}
	return template, nil
}

func readData(cmd *cobra.Command, value string) ([]byte, error) {
	var data []byte
	var err error
	if value == "@-" {
		data, err = io.ReadAll(cmd.InOrStdin())
	} else if strings.HasPrefix(value, "@") {
		data, err = os.ReadFile(strings.TrimPrefix(value, "@"))
	} else {
		data = []byte(value)
	}
	if err != nil {
		return nil, err
	}
	if !json.Valid(data) {
		return nil, errors.New("--data must contain valid JSON")
	}
	return data, nil
}

func printResponse(output io.Writer, response []byte) error {
	if len(response) == 0 {
		_, err := io.WriteString(output, "{}\n")
		return err
	}
	trimmed := bytes.TrimSpace(response)
	if json.Valid(trimmed) {
		_, err := fmt.Fprintln(output, string(trimmed))
		return err
	}
	return json.NewEncoder(output).Encode(map[string]string{
		"encoding": "base64",
		"data":     base64.StdEncoding.EncodeToString(response),
	})
}
