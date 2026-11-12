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
	"strconv"
	"strings"
	"sync"
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
		storageID        string
		storagePath      string
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
		{name: "managed storage override", storageID: "stor_chunkify", putStatus: 200, completionStatus: 204, wantRequests: []string{"create", "put", "complete", "upload", "source"}},
		{name: "external default", storagePath: "incoming/video.mp4", putStatus: 200, completionStatus: 204, wantRequests: []string{"create", "put", "complete", "upload", "source"}},
		{name: "external override", storageID: "stor_external", storagePath: "incoming/video.mp4", putStatus: 200, completionStatus: 204, wantRequests: []string{"create", "put", "complete", "upload", "source"}},
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
				case "GET /sources":
					requests = append(requests, "lookup")
					fmt.Fprint(w, `{"data":[]}`)
				case "GET /uploads/upl_test":
					requests = append(requests, "upload")
					sourceID := "src_test"
					if tt.missingSource {
						sourceID = ""
					}
					json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": "upl_test", "source_id": sourceID}})
				case "GET /sources/src_test":
					requests = append(requests, "source")
					json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": "src_test", "metadata": metadata}})
				case "POST /uploads":
					requests = append(requests, "create")
					if r.Header.Get("Authorization") != "Bearer project-secret" {
						t.Error("session creation must use the project token")
					}
					var params struct {
						Metadata      map[string]string
						Storage       map[string]string
						FileSizeBytes int64 `json:"file_size_bytes"`
					}
					if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
						t.Error(err)
					}
					if params.FileSizeBytes != int64(len(content)) {
						t.Errorf("file_size_bytes = %d, want %d", params.FileSizeBytes, len(content))
					}
					metadata = params.Metadata
					wantStorage := map[string]string{}
					if tt.storageID != "" {
						wantStorage["id"] = tt.storageID
					}
					if tt.storagePath != "" {
						wantStorage["path"] = tt.storagePath
					}
					if len(wantStorage) == 0 {
						if len(params.Storage) != 0 {
							t.Errorf("unexpected storage fields: %v", params.Storage)
						}
					} else if !reflect.DeepEqual(params.Storage, wantStorage) {
						t.Errorf("storage = %v, want %v", params.Storage, wantStorage)
					}
					completionURL := server.URL + "/uploads/completion/" + completionToken
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
						"id": "upl_test", "upload_mode": "single", "upload_url": server.URL + "/storage", "completion_url": completionURL, "expires_at": expiresAt,
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
				case "POST /uploads/completion/" + completionToken:
					requests = append(requests, "complete")
					completionCalls++
					body, _ := io.ReadAll(r.Body)
					if (len(body) != 0 && string(body) != "{}") || r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" {
						t.Errorf("single-PUT completion must send no parts, project token, or cookies: %s", body)
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
			app.Command = &ChunkifyCommand{Input: file, Id: executionID, UploadStorageID: tt.storageID, UploadStoragePath: tt.storagePath}
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

func TestCreateSourceFromFileMultipart(t *testing.T) {
	const partSize = 4
	content := "0123456789" // three parts: 4 + 4 + 2 bytes
	const completionToken = "private-completion-token"
	var mu sync.Mutex
	received := map[int]string{}
	var completion struct {
		Parts []struct {
			Etag       string `json:"etag"`
			PartNumber int64  `json:"part_number"`
		} `json:"parts"`
	}
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == "GET" && r.URL.Path == "/sources":
			fmt.Fprint(w, `{"data":[]}`)
		case r.Method == "POST" && r.URL.Path == "/uploads":
			parts := []map[string]any{}
			for i := 1; i <= 3; i++ {
				parts = append(parts, map[string]any{"part_number": i, "upload_url": fmt.Sprintf("%s/part/%d", server.URL, i)})
			}
			json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{
				"id": "upl_test", "upload_mode": "multipart", "completion_url": server.URL + "/uploads/completion/" + completionToken,
				"expires_at": time.Now().Add(time.Hour),
				"multipart":  map[string]any{"part_size_bytes": partSize, "parts": parts, "expires_at": time.Now().Add(time.Hour)},
			}})
		case r.Method == "PUT" && strings.HasPrefix(r.URL.Path, "/part/"):
			n, _ := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/part/"))
			body, _ := io.ReadAll(r.Body)
			mu.Lock()
			received[n] = string(body)
			mu.Unlock()
			w.Header().Set("ETag", fmt.Sprintf(`"etag-%d"`, n))
		case r.Method == "POST" && r.URL.Path == "/uploads/completion/"+completionToken:
			if err := json.NewDecoder(r.Body).Decode(&completion); err != nil {
				t.Error(err)
			}
			w.WriteHeader(204)
		case r.Method == "GET" && r.URL.Path == "/uploads/upl_test":
			fmt.Fprint(w, `{"data":{"id":"upl_test","source_id":"src_test"}}`)
		case r.Method == "GET" && r.URL.Path == "/sources/src_test":
			fmt.Fprint(w, `{"data":{"id":"src_test"}}`)
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
	app.Command = &ChunkifyCommand{Input: file, Id: "test-execution"}
	source, err := app.CreateSourceFromFile(context.Background())
	if err != nil || source.ID != "src_test" {
		t.Fatalf("want completed source, got %v, %v", source, err)
	}
	if want := map[int]string{1: "0123", 2: "4567", 3: "89"}; !reflect.DeepEqual(received, want) {
		t.Errorf("parts = %v, want %v", received, want)
	}
	if len(completion.Parts) != 3 {
		t.Fatalf("completion parts = %+v", completion.Parts)
	}
	for i, part := range completion.Parts {
		if part.PartNumber != int64(i+1) || part.Etag != fmt.Sprintf(`"etag-%d"`, i+1) {
			t.Errorf("completion part %d = %+v", i, part)
		}
	}
}

func TestUploadMultipartRetriesAndErrors(t *testing.T) {
	multipartRetryDelay = time.Millisecond
	t.Cleanup(func() { multipartRetryDelay = time.Second })

	for _, tt := range []struct {
		name      string
		handler   func(attempt int, w http.ResponseWriter)
		parts     int
		wantError string
		wantCalls int
	}{
		{name: "retries server error", parts: 1, wantCalls: 2, handler: func(attempt int, w http.ResponseWriter) {
			if attempt == 1 {
				w.WriteHeader(503)
				return
			}
			w.Header().Set("ETag", `"ok"`)
		}},
		{name: "does not retry forbidden", parts: 1, wantCalls: 1, wantError: "403", handler: func(int, http.ResponseWriter) {}},
		{name: "missing etag", parts: 1, wantCalls: 1, wantError: "no ETag", handler: func(int, http.ResponseWriter) {}},
		{name: "gives up after retries", parts: 1, wantCalls: multipartAttempts, wantError: "503", handler: func(_ int, w http.ResponseWriter) { w.WriteHeader(503) }},
		{name: "part count mismatch", parts: 2, wantCalls: 0, wantError: "expected 1", handler: func(int, http.ResponseWriter) {}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				io.Copy(io.Discard, r.Body)
				calls++
				if tt.name == "does not retry forbidden" {
					w.WriteHeader(403)
					return
				}
				tt.handler(calls, w)
			}))
			defer server.Close()

			parts := []chunkify.UploadMultipartPart{}
			for i := 1; i <= tt.parts; i++ {
				parts = append(parts, chunkify.UploadMultipartPart{PartNumber: int64(i), UploadURL: server.URL})
			}
			progress := make(chan UploadProgress, 100)
			result, err := UploadMultipartWithProgress(context.Background(), strings.NewReader("data"), 4,
				chunkify.UploadMultipart{PartSizeBytes: 4, Parts: parts}, progress)
			if tt.wantError == "" {
				if err != nil || len(result) != 1 || result[0].Etag != `"ok"` {
					t.Fatalf("unexpected result %+v, %v", result, err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("want error containing %q, got %v", tt.wantError, err)
			}
			if calls != tt.wantCalls {
				t.Errorf("calls = %d, want %d", calls, tt.wantCalls)
			}
			if _, open := <-drain(progress); open {
				t.Error("progress channel must be closed")
			}
		})
	}
}

// drain discards buffered progress and returns the closed channel.
func drain(ch chan UploadProgress) chan UploadProgress {
	for range ch {
	}
	return ch
}
