package chunkify

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/chunkify-go/option"
)

func TestCreateSourceFromStorage(t *testing.T) {
	for _, tt := range []struct{ name, configured, explicit, wantID string }{
		{"project default", "", "", ""},
		{"configured external", "stor_aws_config", "", "stor_aws_config"},
		{"explicit external", "", "stor_aws_input", "stor_aws_input"},
		{"explicit overrides configured", "stor_aws_config", "stor_aws_input", "stor_aws_input"},
		{"external source with managed output", "stor_chunkify_config", "stor_aws_input", "stor_aws_input"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			const key = "/videos//../café + %2F?#.mp4"
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				if r.Method != http.MethodPost || r.URL.Path != "/api/sources" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				if r.Header.Get("Authorization") != "Bearer project-secret" {
					t.Error("missing project authentication")
				}
				var payload map[string]json.RawMessage
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					t.Error(err)
				}
				if _, exists := payload["url"]; exists {
					t.Error("storage input must omit url")
				}
				var storage map[string]string
				if err := json.Unmarshal(payload["storage"], &storage); err != nil {
					t.Error(err)
				}
				if storage["path"] != key || storage["id"] != tt.wantID {
					t.Errorf("unexpected source storage: %v", storage)
				}
				if tt.wantID == "" {
					if _, exists := storage["id"]; exists {
						t.Error("default storage ID must be omitted")
					}
				}
				var metadata map[string]string
				if err := json.Unmarshal(payload["metadata"], &metadata); err != nil {
					t.Error(err)
				}
				if metadata["origin"] != MetadataOrigin || metadata["cli_execution_id"] != "run-123" {
					t.Errorf("unexpected metadata: %v", metadata)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				io.WriteString(w, `{"data":{"id":"src_stored"}}`)
			}))
			defer server.Close()
			client := chunkify.NewClient(option.WithBaseURL(server.URL), option.WithProjectAccessToken("project-secret"))
			app := NewApp()
			app.Client = &client
			app.Command = &ChunkifyCommand{Id: "run-123", Input: "store://" + key, SourceStorageID: tt.explicit, Format: FormatMp4H264, Output: "output.mp4"}
			if err := app.validateSourceInput(); err != nil {
				t.Fatal(err)
			}
			if err := app.configureStorage(tt.configured); err != nil {
				t.Fatal(err)
			}
			if app.Command.UploadStorageID != "" || app.Command.UploadStoragePath != "" {
				t.Fatal("stored sources must not generate an upload destination")
			}
			source, err := app.CreateSource(context.Background())
			if err != nil || source == nil || source.ID != "src_stored" {
				t.Fatalf("source = %v, error = %v", source, err)
			}
			if requests != 1 {
				t.Errorf("expected one source creation and no upload, got %d requests", requests)
			}
			if status := <-app.Progress.Status; status != ReadingFromStorage {
				t.Errorf("unexpected status %d", status)
			}
			if app.Command.JobCreateStorageParams.ID.Value != tt.configured {
				t.Error("source selection changed output storage")
			}
		})
	}
}

func TestSourceStorageValidation(t *testing.T) {
	for _, tt := range []struct{ name, input, sourceID, uploadID, uploadPath, want string }{
		{"empty key", "store://", "", "", "", "object key"},
		{"too many bytes", "store://" + strings.Repeat("é", 513), "", "", "", "1024"},
		{"invalid UTF8", "store://\xff", "", "", "", "UTF-8"},
		{"managed source", "store://video.mp4", "stor_chunkify_test", "", "", "external storage"},
		{"source flag with local file", "video.mp4", "stor_aws_test", "", "", "requires a store:// input"},
		{"source flag with URL", "https://example.com/video.mp4", "stor_aws_test", "", "", "requires a store:// input"},
		{"source flag with source ID", "src_existing", "stor_aws_test", "", "", "requires a store:// input"},
		{"upload ID with storage source", "store://video.mp4", "", "stor_aws_test", "", "upload storage flags"},
		{"upload path with storage source", "store://video.mp4", "", "", "video.mp4", "upload storage flags"},
		{"maximum key", "store://" + strings.Repeat("é", 512), "", "", "", ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			app := &App{Command: &ChunkifyCommand{Input: tt.input, SourceStorageID: tt.sourceID, UploadStorageID: tt.uploadID, UploadStoragePath: tt.uploadPath}}
			err := app.validateSourceInput()
			if tt.want == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected %q, got %v", tt.want, err)
			}
		})
	}
	app := &App{Command: &ChunkifyCommand{Input: "store://video.mp4"}}
	if err := app.configureStorage("stor_chunkify_config"); err == nil {
		t.Fatal("configured managed source storage must be rejected")
	}
}

func TestSourceStorageErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		io.WriteString(w, `{"error":{"message":"storage access denied"}}`)
	}))
	defer server.Close()
	client := chunkify.NewClient(option.WithBaseURL(server.URL), option.WithProjectAccessToken("project-secret"))
	app := NewApp()
	app.Client = &client
	app.Command = &ChunkifyCommand{Input: "store://video.mp4"}
	if _, err := app.CreateSource(context.Background()); err == nil || !strings.Contains(err.Error(), "storage access denied") {
		t.Fatalf("expected API failure, got %v", err)
	}
	<-app.Progress.Status
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := app.CreateSource(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
}
