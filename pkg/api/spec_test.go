package api

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/chunkifydev/cli/pkg/config"
	"github.com/zalando/go-keyring"
)

func TestConfiguredOpenAPIURLIsLoaded(t *testing.T) {
	keyring.MockInit()
	t.Setenv("CHUNKIFY_OPENAPI_URL", "")
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, testSpec)
	}))
	defer server.Close()
	cfg := &config.Config{Profile: "staging"}
	if err := config.Set(cfg.ConfigKey(config.ConfigOpenAPIURLKey), server.URL); err != nil {
		t.Fatal(err)
	}
	operations, err := loadConfiguredOperations(context.Background(), cfg)
	if err != nil || len(operations) != 5 {
		t.Fatalf("configured OpenAPI URL: %d operations, %v", len(operations), err)
	}
}

func TestOpenAPICacheAndStaleFallback(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, testSpec)
	}))
	cacheDir := t.TempDir()
	client := &http.Client{Timeout: time.Second}
	for i := 0; i < 2; i++ {
		operations, err := loadOperations(context.Background(), server.URL, cacheDir, client)
		if err != nil || len(operations) != 5 {
			t.Fatalf("load %d: %d operations, %v", i, len(operations), err)
		}
	}
	if calls != 1 {
		t.Fatalf("fresh cache should avoid a second fetch; got %d requests", calls)
	}
	files, err := filepath.Glob(filepath.Join(cacheDir, "openapi-*.json"))
	if err != nil || len(files) != 1 {
		t.Fatalf("cache files: %v, %v", files, err)
	}
	old := time.Now().Add(-2 * specCacheAge)
	if err := os.Chtimes(files[0], old, old); err != nil {
		t.Fatal(err)
	}
	server.Close()
	operations, err := loadOperations(context.Background(), server.URL, cacheDir, client)
	if err != nil || len(operations) != 5 {
		t.Fatalf("stale cache fallback: %d operations, %v", len(operations), err)
	}
}

func TestOpenAPIRequiresKnownSecurity(t *testing.T) {
	bad := `{"paths":{"/api/projects":{"get":{"operationId":"listProjects"}}}}`
	if _, err := parseOperations([]byte(bad)); err == nil {
		t.Fatal("accepted an operation without a security definition")
	}
}

func TestCommandNamesFollowOperationIDs(t *testing.T) {
	for _, tc := range []struct {
		id, method, path, resource, action string
	}{
		{"listJobs", "get", "/api/jobs", "jobs", "list"},
		{"getJob", "get", "/api/jobs/{jobId}", "jobs", "get"},
		{"cancelJob", "post", "/api/jobs/{jobId}/cancel", "jobs", "cancel"},
		{"getJobFiles", "get", "/api/jobs/{jobId}/files", "job-files", "list"},
		{"getJobLogs", "get", "/api/jobs/{jobId}/logs", "job-logs", "list"},
		{"getJobTranscoders", "get", "/api/jobs/{jobId}/transcoders", "job-transcoders", "list"},
		{"completeUpload", "post", "/api/uploads/completion/{token}", "uploads", "complete"},
		{"createAssetStaticRendition", "post", "/api/assets/{assetId}/static-renditions", "asset-static-renditions", "create"},
		{"deleteAssetStaticRendition", "delete", "/api/assets/{assetId}/static-renditions/{renditionId}", "asset-static-renditions", "delete"},
		{"getAssetPreview", "get", "/api/assets/{assetId}/preview.m3u8", "asset-previews", "get"},
		{"getAssetPreviewObject", "get", "/api/assets/{assetId}/preview/{objectPath}", "asset-preview-objects", "get"},
		{"retryJob", "post", "/api/jobs/{jobId}/retry", "jobs", "retry"},
	} {
		resource, action, err := commandName(tc.id, tc.method, tc.path)
		if err != nil || resource != tc.resource || action != tc.action {
			t.Errorf("%s: got %s %s, %v; want %s %s", tc.id, resource, action, err, tc.resource, tc.action)
		}
	}
}
