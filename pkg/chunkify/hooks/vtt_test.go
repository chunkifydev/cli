package hooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProcessVtt(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "vtt_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name            string
		downloadedFiles []string
		basename        string
		vttContent      string
		expectedError   bool
		expectedContent string
		description     string
	}{
		{
			name: "Successful VTT processing",
			downloadedFiles: []string{
				filepath.Join(tempDir, "images.vtt"),
				filepath.Join(tempDir, "video-frame-00001.jpg"),
				filepath.Join(tempDir, "video-frame-00002.jpg"),
			},
			basename: "job_1234",
			vttContent: `WEBVTT

00:00:00.000 --> 00:00:05.000
job_1234-00001.jpg

00:00:05.000 --> 00:00:10.000
job_1234-00002.jpg`,
			expectedError: false,
			expectedContent: `WEBVTT

00:00:00.000 --> 00:00:05.000
video-frame-00001.jpg

00:00:05.000 --> 00:00:10.000
video-frame-00002.jpg`,
			description: "Should replace basename with image basename in VTT content",
		},
		{
			name: "Error Image basename not found",
			downloadedFiles: []string{
				filepath.Join(tempDir, "images.vtt"),
				filepath.Join(tempDir, "video00001.jpg"),
				filepath.Join(tempDir, "video00002.jpg"),
			},
			basename: "job_1234",
			vttContent: `WEBVTT

00:00:00.000 --> 00:00:05.000
error-00001.jpg

00:00:05.000 --> 00:00:10.000
error-00002.jpg`,
			expectedError: true,
			expectedContent: `WEBVTT

00:00:00.000 --> 00:00:05.000
video-frame-00001.jpg

00:00:05.000 --> 00:00:10.000
video-frame-00002.jpg`,
			description: "Should replace basename with image basename in VTT content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing test files
			for _, file := range tt.downloadedFiles {
				os.Remove(file)
			}

			// Create VTT file if it should exist and has content
			var vttFile string
			for _, file := range tt.downloadedFiles {
				if strings.HasSuffix(file, ".vtt") {
					vttFile = file
					if tt.vttContent != "" {
						err := os.WriteFile(file, []byte(tt.vttContent), 0644)
						if err != nil {
							t.Fatalf("Failed to create VTT file: %v", err)
						}
					}
					break
				}
			}

			// Run the function
			err := ProcessVtt(tt.downloadedFiles, tt.basename)

			// Check for expected error
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify the VTT file content was modified correctly
			if vttFile != "" {
				content, err := os.ReadFile(vttFile)
				if err != nil {
					t.Errorf("Failed to read VTT file after processing: %v", err)
					return
				}

				actualContent := string(content)
				if actualContent != tt.expectedContent {
					t.Errorf("VTT content mismatch.\nExpected: %q\nActual: %q", tt.expectedContent, actualContent)
				}
			}
		})
	}
}
