package chunkify

import (
	"encoding/json"
	"strings"
	"testing"

	chunkify "github.com/chunkifydev/chunkify-go"
)

func TestConfigureStorage(t *testing.T) {
	for _, tt := range []struct {
		name, configured, uploadID, outputID, uploadPath, outputPath string
		wantUploadID, wantOutputID, wantUploadPath, wantOutputPath   string
		wantError                                                    string
	}{
		{name: "project default remains implicit"},
		{name: "manual flags preserved", uploadID: "stor_aws_input", outputID: "stor_aws_output", uploadPath: "in/video.mp4", outputPath: "out/result.mp4", wantUploadID: "stor_aws_input", wantOutputID: "stor_aws_output", wantUploadPath: "in/video.mp4", wantOutputPath: "out/result.mp4"},
		{name: "manual IDs do not generate paths", uploadID: "stor_aws_input", outputID: "stor_aws_output", wantUploadID: "stor_aws_input", wantOutputID: "stor_aws_output"},
		{name: "configured external overrides flags", configured: "stor_aws_config", uploadID: "stor_chunkify_input", outputID: "stor_chunkify_output", wantUploadID: "stor_aws_config", wantOutputID: "stor_aws_config", wantUploadPath: "chunkify-cli/sources/run-123/video.mp4", wantOutputPath: "chunkify-cli/jobs/run-123/result.mp4"},
		{name: "configured managed overrides flags", configured: "stor_chunkify_config", uploadID: "stor_aws_input", outputID: "stor_aws_output", wantUploadID: "stor_chunkify_config", wantOutputID: "stor_chunkify_config"},
		{name: "configured external preserves paths", configured: "stor_aws_config", uploadPath: "custom/input.mp4", outputPath: "custom/output.mp4", wantUploadID: "stor_aws_config", wantOutputID: "stor_aws_config", wantUploadPath: "custom/input.mp4", wantOutputPath: "custom/output.mp4"},
		{name: "managed rejects upload path", configured: "stor_chunkify_config", uploadPath: "video.mp4", wantError: "--upload-storage-path"},
		{name: "managed rejects output path", configured: "stor_chunkify_config", outputPath: "result.mp4", wantError: "--output-storage-path"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			app := &App{Command: &ChunkifyCommand{Id: "run-123", Input: "/tmp/video.mp4", Output: "/tmp/result.mp4", Format: FormatMp4H264, UploadStorageID: tt.uploadID, OutputStorageID: tt.outputID, UploadStoragePath: tt.uploadPath}}
			if tt.outputID != "" {
				app.Command.JobCreateStorageParams.ID = chunkify.String(tt.outputID)
			}
			if tt.outputPath != "" {
				app.Command.JobCreateStorageParams.Path = chunkify.String(tt.outputPath)
			}
			err := app.configureStorage(tt.configured)
			if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Fatalf("expected %q, got %v", tt.wantError, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if app.Command.UploadStorageID != tt.wantUploadID || app.Command.UploadStoragePath != tt.wantUploadPath || app.Command.OutputStorageID != tt.wantOutputID {
				t.Fatalf("unexpected storage config: %+v", app.Command)
			}
			data, err := json.Marshal(app.Command.JobCreateStorageParams)
			if err != nil {
				t.Fatal(err)
			}
			var payload map[string]string
			if err := json.Unmarshal(data, &payload); err != nil {
				t.Fatal(err)
			}
			if payload["id"] != tt.wantOutputID || payload["path"] != tt.wantOutputPath {
				t.Fatalf("unexpected job storage payload: %s", data)
			}
			if tt.wantOutputPath == "" {
				if _, exists := payload["path"]; exists {
					t.Fatalf("empty path must be omitted: %s", data)
				}
			}
		})
	}
}

func TestConfigureStorageSkipsUnusedPaths(t *testing.T) {
	app := &App{Command: &ChunkifyCommand{Input: "src_existing", Format: FormatMp4H264, UploadStoragePath: "unused"}}
	if err := app.configureStorage("stor_chunkify_test"); err != nil {
		t.Fatal(err)
	}
	app = &App{Command: &ChunkifyCommand{Input: "video.mp4", JobCreateStorageParams: chunkify.JobNewParamsStorage{Path: chunkify.String("unused")}}}
	if err := app.configureStorage("stor_chunkify_test"); err != nil {
		t.Fatal(err)
	}
}

func TestStorageObjectPath(t *testing.T) {
	for _, tt := range []struct {
		name, id, explicit, category, filename, want string
		wantError                                    bool
	}{
		{name: "managed ID fallback", id: "stor_chunkify_test", want: ""},
		{name: "managed explicit path", id: "stor_chunkify_test", explicit: "video.mp4", wantError: true},
		{name: "external upload", id: "stor_aws_test", category: "sources", filename: "/tmp/input/video.mp4", want: "chunkify-cli/sources/run-123/video.mp4"},
		{name: "external output", id: "stor_s3_compatible_test", category: "jobs", filename: "/tmp/output/video.mp4", want: "chunkify-cli/jobs/run-123/video.mp4"},
		{name: "explicit path preserved", id: "stor_aws_test", explicit: "/custom//video.mp4", want: "/custom//video.mp4"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			app := &App{Command: &ChunkifyCommand{Id: "run-123"}}
			got, err := app.storageObjectPath(tt.id, tt.explicit, tt.category, tt.filename, "--upload-storage-path")
			if tt.wantError {
				if err == nil || !strings.Contains(err.Error(), "--upload-storage-path") {
					t.Fatalf("expected an actionable flag error, got %v", err)
				}
			} else if err != nil || got != tt.want {
				t.Fatalf("path = %q, error = %v, want %q", got, err, tt.want)
			}
		})
	}
}

func TestGeneratedStoragePathsUseExecutionID(t *testing.T) {
	storage := "stor_aws_test"
	first := &App{Command: &ChunkifyCommand{}}
	second := &App{Command: &ChunkifyCommand{}}
	one, _ := first.storageObjectPath(storage, "", "sources", "video.mp4", "--upload-storage-path")
	retry, _ := first.storageObjectPath(storage, "", "sources", "video.mp4", "--upload-storage-path")
	two, _ := second.storageObjectPath(storage, "", "sources", "video.mp4", "--upload-storage-path")
	if one != retry || one == two || first.Command.Id == "" {
		t.Fatalf("paths must be stable for retries and unique across executions: %q, %q, %q", one, retry, two)
	}
}

func TestJobOutputFilename(t *testing.T) {
	for _, tt := range []struct{ format, output, want string }{
		{FormatMp4H264, "/tmp/custom.mp4", "custom.mp4"},
		{FormatMp4H264, "", "output.mp4"},
		{FormatWebmVp9, "", "output.webm"},
		{FormatHlsH264, "", "output.m3u8"},
		{FormatJpg, "", "output.jpg"},
	} {
		app := &App{Command: &ChunkifyCommand{Format: tt.format, Output: tt.output}}
		if got := app.jobOutputFilename(); got != tt.want {
			t.Errorf("format %s: filename = %q, want %q", tt.format, got, tt.want)
		}
	}
}
