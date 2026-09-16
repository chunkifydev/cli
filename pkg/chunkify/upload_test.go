package chunkify

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/chunkify-go/option"
)

func TestCreateSourceFromFileCompletesUpload(t *testing.T) {
	for _, tt := range []struct {
		name             string
		putStatus        int
		completionStatus int
		completionURL    string
		missingSource    bool
		alwaysFail       bool
		expired          bool
		wantError        string
		wantRequests     []string
	}{
		{name: "success", putStatus: 200, completionStatus: 204, wantRequests: []string{"lookup", "create", "put", "complete", "upload", "source"}},
		{name: "storage returns 201", putStatus: 201, completionStatus: 204, wantRequests: []string{"lookup", "create", "put", "complete", "upload", "source"}},
		{name: "storage returns 204", putStatus: 204, completionStatus: 204, wantRequests: []string{"lookup", "create", "put", "complete", "upload", "source"}},
		{name: "failed transfer", putStatus: 403, wantError: "error uploading blob", wantRequests: []string{"lookup", "create", "put"}},
		{name: "invalid token", putStatus: 200, completionStatus: 401, wantError: "error completing upload", wantRequests: []string{"lookup", "create", "put", "complete"}},
		{name: "rejected file", putStatus: 200, completionStatus: 400, wantError: "error completing upload", wantRequests: []string{"lookup", "create", "put", "complete"}},
		{name: "forbidden storage", putStatus: 200, completionStatus: 403, wantError: "error completing upload", wantRequests: []string{"lookup", "create", "put", "complete"}},
		{name: "expired completion", putStatus: 200, completionStatus: 410, wantError: "error completing upload", wantRequests: []string{"lookup", "create", "put", "complete"}},
		{name: "retry server error", putStatus: 200, completionStatus: 503, wantRequests: []string{"lookup", "create", "put", "complete", "complete", "upload", "source"}},
		{name: "retry rate limit", putStatus: 200, completionStatus: 429, wantRequests: []string{"lookup", "create", "put", "complete", "complete", "upload", "source"}},
		{name: "missing URL", completionURL: "missing", wantError: "missing upload session URLs", wantRequests: []string{"lookup", "create"}},
		{name: "invalid URL", completionURL: "://invalid", wantError: "invalid upload completion URL", wantRequests: []string{"lookup", "create"}},
		{name: "expired session", expired: true, wantError: "context deadline exceeded", wantRequests: []string{"lookup", "create"}},
		{name: "missing source relationship", putStatus: 200, completionStatus: 204, missingSource: true, wantError: "completed upload has no source", wantRequests: []string{"lookup", "create", "put", "complete", "upload"}},
		{name: "exhausted retries", putStatus: 200, completionStatus: 503, alwaysFail: true, wantError: "error completing upload", wantRequests: []string{"lookup", "create", "put", "complete", "complete", "complete"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			const content = "test video bytes"
			const executionID = "test-execution"
			const completionToken = "private-completion-token"
			var requests []string
			var metadata map[string]string
			completionCalls := 0
			var server *httptest.Server
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch r.Method + " " + r.URL.Path {
				case "GET /api/sources":
					requests = append(requests, "lookup")
					fmt.Fprint(w, `{"data":[]}`)
				case "GET /api/uploads/upl_test":
					requests = append(requests, "upload")
					sourceID := "src_test"
					if tt.missingSource {
						sourceID = ""
					}
					json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": "upl_test", "source_id": sourceID}})
				case "GET /api/sources/src_test":
					requests = append(requests, "source")
					json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": "src_test", "metadata": metadata}})
				case "POST /api/uploads":
					requests = append(requests, "create")
					if r.Header.Get("Authorization") != "Bearer project-secret" {
						t.Error("session creation must use the project token")
					}
					var params struct {
						Metadata map[string]string
						Storage  map[string]string
					}
					if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
						t.Error(err)
					}
					metadata = params.Metadata
					if len(params.Storage) != 0 {
						t.Errorf("unexpected storage fields: %v", params.Storage)
					}
					completionURL := server.URL + "/api/uploads/completion/" + completionToken
					if tt.completionURL == "missing" {
						completionURL = ""
					} else if tt.completionURL != "" {
						completionURL = tt.completionURL
					}
					expiresAt := time.Now().Add(time.Hour)
					if tt.expired {
						expiresAt = time.Now().Add(-time.Second)
					}
					json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{
						"id": "upl_test", "upload_url": server.URL + "/storage", "completion_url": completionURL, "expires_at": expiresAt,
					}})
				case "PUT /storage":
					requests = append(requests, "put")
					body, _ := io.ReadAll(r.Body)
					if string(body) != content || r.ContentLength != int64(len(content)) {
						t.Errorf("incorrect transfer: body=%q length=%d", body, r.ContentLength)
					}
					if r.Header.Get("Authorization") != "" {
						t.Error("storage received project credentials")
					}
					w.WriteHeader(tt.putStatus)
				case "POST /api/uploads/completion/" + completionToken:
					requests = append(requests, "complete")
					completionCalls++
					body, _ := io.ReadAll(r.Body)
					if len(body) != 0 || r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" {
						t.Error("completion must send no body, project token, or cookies")
					}
					if completionCalls > 1 && !tt.alwaysFail {
						w.WriteHeader(204)
						return
					}
					w.Header().Set("Retry-After-Ms", "1")
					w.WriteHeader(tt.completionStatus)
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(404)
				}
			}))
			defer server.Close()

			file := filepath.Join(t.TempDir(), "video.mp4")
			if err := os.WriteFile(file, []byte(content), 0600); err != nil {
				t.Fatal(err)
			}
			client := chunkify.NewClient(option.WithBaseURL(server.URL), option.WithProjectAccessToken("project-secret"))
			app := NewApp()
			app.Client = &client
			app.Command = &ChunkifyCommand{Input: file, Id: executionID}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			source, err := app.CreateSourceFromFile(ctx)
			if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Fatalf("want error containing %q, got %v", tt.wantError, err)
				}
				if strings.Contains(err.Error(), completionToken) {
					t.Error("completion error exposed the private token")
				}
			} else if err != nil || source == nil || source.ID != "src_test" {
				t.Fatalf("want completed source, got %v, %v", source, err)
			}
			if !reflect.DeepEqual(requests, tt.wantRequests) {
				t.Errorf("requests = %v, want %v", requests, tt.wantRequests)
			}
			if metadata["cli_execution_id"] != executionID || metadata["origin"] != MetadataOrigin || metadata["md5"] == "" {
				t.Errorf("upload metadata was not preserved: %v", metadata)
			}
		})
	}
}

func TestUploadBlobAcceptsSuccessfulStorageResponses(t *testing.T) {
	for _, status := range []int{200, 201, 204, 403, 500} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				io.Copy(io.Discard, r.Body)
				w.WriteHeader(status)
			}))
			defer server.Close()
			err := UploadBlob(context.Background(), strings.NewReader("video"), &chunkify.Upload{UploadURL: server.URL})
			if (err == nil) != (status < 300) {
				t.Errorf("status %d: unexpected error %v", status, err)
			}
		})
	}
}

func TestUploadProgressClosesOnInvalidRequest(t *testing.T) {
	progress := make(chan UploadProgress, 1)
	err := UploadBlobWithProgress(context.Background(), strings.NewReader("video"), &chunkify.Upload{UploadURL: "://invalid"}, progress)
	if err == nil {
		t.Fatal("expected invalid URL error")
	}
	select {
	case _, ok := <-progress:
		if ok {
			t.Fatal("progress channel must be closed")
		}
	default:
		t.Fatal("progress channel was left open")
	}
}
