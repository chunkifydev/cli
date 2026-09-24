package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/chunkify-go/option"
	"github.com/chunkifydev/cli/pkg/config"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

const testSpec = `{"paths":{
  "/api/projects":{"get":{"operationId":"listProjects","security":[{"TeamAccessToken":[]}]}},
  "/api/projects/{projectId}":{"patch":{"operationId":"updateProject","security":[{"TeamAccessToken":[]}],"requestBody":{"required":true,"content":{"application/json":{}}}}},
  "/api/jobs":{"get":{"operationId":"listJobs","security":[{"ProjectAccessToken":[]}],"parameters":[{"in":"query","name":"limit"},{"in":"query","name":"status"}]}},
  "/api/uploads/completion/{token}":{"post":{"operationId":"completeUpload","security":[]}},
  "/api/assets/{assetId}/preview.m3u8":{"get":{"operationId":"getAssetPreview","security":[]}}
}}`

func testCommand(cfg *config.Config) *cobra.Command {
	return newCommand(cfg, func(context.Context) ([]operation, error) { return parseOperations([]byte(testSpec)) })
}

func TestAPIRequestsUseCorrectTokenAndKeepJSON(t *testing.T) {
	type request struct {
		method string
		path   string
		query  string
		auth   string
		body   string
	}
	var received request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		received = request{r.Method, r.URL.Path, r.URL.RawQuery, r.Header.Get("Authorization"), string(body)}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":"success","data":{"id":"project_123"}}`)
	}))
	defer server.Close()
	client := chunkify.NewClient(option.WithBaseURL(server.URL + "/v1/"))
	cfg := &config.Config{Client: &client, TeamToken: "sk_team_example", Token: "sk_project_example"}

	var output bytes.Buffer
	cmd := testCommand(cfg)
	cmd.SetOut(&output)
	cmd.SetArgs([]string{"projects", "update", "my-project", "--data", `{"storage_id":"stor_123"}`})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if received.method != http.MethodPatch || received.path != "/v1/api/projects/my-project" || received.auth != "Bearer sk_team_example" || received.body != `{"storage_id":"stor_123"}` {
		t.Fatalf("wrong team request: %+v", received)
	}
	if output.String() != "{\"status\":\"success\",\"data\":{\"id\":\"project_123\"}}\n" {
		t.Fatalf("response changed: %q", output.String())
	}

	output.Reset()
	cmd = testCommand(cfg)
	cmd.SetOut(&output)
	cmd.SetArgs([]string{"jobs", "list", "--query", "limit=2", "--query", "status=completed"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if received.method != http.MethodGet || received.path != "/v1/api/jobs" || received.query != "limit=2&status=completed" || received.auth != "Bearer sk_project_example" {
		t.Fatalf("wrong project request: %+v", received)
	}
}

func TestAPIEmptyAndNonJSONResponsesStayJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "completion") {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		_, _ = io.WriteString(w, "#EXTM3U\n")
	}))
	defer server.Close()
	client := chunkify.NewClient(option.WithBaseURL(server.URL + "/v1/"))
	cfg := &config.Config{Client: &client}
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"uploads", "complete", "token_123"}, "{}\n"},
		{[]string{"asset-previews", "get", "asset_123"}, "{\"data\":\"I0VYVE0zVQo=\",\"encoding\":\"base64\"}\n"},
	} {
		var output bytes.Buffer
		cmd := testCommand(cfg)
		cmd.SetOut(&output)
		cmd.SetArgs(tc.args)
		if err := cmd.Execute(); err != nil {
			t.Fatal(err)
		}
		if output.String() != tc.want || !json.Valid(bytes.TrimSpace(output.Bytes())) {
			t.Fatalf("response for %v: %q", tc.args, output.String())
		}
	}
}

func TestAPIRequiresOnlyRelevantToken(t *testing.T) {
	keyring.MockInit()
	t.Setenv("CHUNKIFY_TEAM_TOKEN", "")
	t.Setenv("CHUNKIFY_TOKEN", "")
	client := chunkify.NewClient()
	cfg := &config.Config{Client: &client}
	cmd := testCommand(cfg)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"projects", "list"})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "team token missing") {
		t.Fatalf("expected team token error, got %v", err)
	}
	cmd = testCommand(cfg)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"jobs", "list"})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "project token missing") {
		t.Fatalf("expected project token error, got %v", err)
	}
}
