package hooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProcessM3u8(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "m3u8_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name                    string
		downloadedFiles         []string
		basename                string
		oldManifestContent      []byte
		m3u8Content             string
		manifestContent         string
		expectedError           bool
		expectedM3u8Content     string
		expectedManifestContent string
		description             string
	}{
		{
			name: "Successful M3U8 processing",
			downloadedFiles: []string{
				filepath.Join(tempDir, "video.m3u8"),
				filepath.Join(tempDir, "manifest.m3u8"),
				filepath.Join(tempDir, "video.mp4"),
			},
			basename: "job_1234",
			m3u8Content: `#EXTM3U
#EXT-X-VERSION:7
#EXT-X-TARGETDURATION:10
#EXTINF:10.0,
job_1234.mp4`,
			manifestContent: `#EXTM3U
#EXT-X-VERSION:7
#EXT-X-STREAM-INF:BANDWIDTH=1000000,RESOLUTION=1920x1080,CODECS="avc1.640015"
job_1234.m3u8`,
			expectedError: false,
			expectedM3u8Content: `#EXTM3U
#EXT-X-VERSION:7
#EXT-X-TARGETDURATION:10
#EXTINF:10.0,
video.mp4`,
			expectedManifestContent: `#EXTM3U
#EXT-X-VERSION:7
#EXT-X-STREAM-INF:BANDWIDTH=1000000,RESOLUTION=1920x1080,CODECS="avc1.640015"
video.m3u8`,
			description: "Should replace basename with video basename and merge manifest",
		},
		{
			name: "Error missing M3U8 file",
			downloadedFiles: []string{
				filepath.Join(tempDir, "manifest.m3u8"),
				filepath.Join(tempDir, "video.mp4"),
			},
			basename:           "job_1234",
			oldManifestContent: []byte(``),
			manifestContent: `#EXTM3U
#EXT-X-VERSION:7
#EXT-X-STREAM-INF:BANDWIDTH=1000000,RESOLUTION=1920x1080,CODECS="avc1.640015"
job_1234.m3u8`,
			expectedError: true,
			description:   "Should return error when M3U8 file is missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing test files
			for _, file := range tt.downloadedFiles {
				os.Remove(file)
			}

			// Create test files
			for _, file := range tt.downloadedFiles {
				if strings.HasSuffix(file, ".m3u8") {
					var content string
					if strings.HasSuffix(file, "manifest.m3u8") {
						content = tt.manifestContent
					} else {
						content = tt.m3u8Content
					}
					if content != "" {
						err := os.WriteFile(file, []byte(content), 0644)
						if err != nil {
							t.Fatalf("Failed to create M3U8 file: %v", err)
						}
					}
				} else if strings.HasSuffix(file, ".mp4") {
					// Create dummy MP4 file
					err := os.WriteFile(file, []byte("dummy video content"), 0644)
					if err != nil {
						t.Fatalf("Failed to create MP4 file: %v", err)
					}
				}
			}

			// Run the function
			err := ProcessM3u8(tt.downloadedFiles, tt.basename, tt.oldManifestContent)

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

			// Verify the M3U8 file content was modified correctly
			for _, file := range tt.downloadedFiles {
				if strings.HasSuffix(file, ".m3u8") && !strings.HasSuffix(file, "manifest.m3u8") {
					content, err := os.ReadFile(file)
					if err != nil {
						t.Errorf("Failed to read M3U8 file after processing: %v", err)
						return
					}
					actualContent := string(content)
					if actualContent != tt.expectedM3u8Content {
						t.Errorf("M3U8 content mismatch.\nExpected: %q\nActual: %q", tt.expectedM3u8Content, actualContent)
					}
				} else if strings.HasSuffix(file, "manifest.m3u8") {
					content, err := os.ReadFile(file)
					if err != nil {
						t.Errorf("Failed to read manifest file after processing: %v", err)
						return
					}
					actualContent := string(content)
					if actualContent != tt.expectedManifestContent {
						t.Errorf("Manifest content mismatch.\nExpected: %q\nActual: %q", tt.expectedManifestContent, actualContent)
					}
				}
			}
		})
	}
}

func TestMergeManifest_Success(t *testing.T) {
	// Example m3u8 content from the user
	currentManifest := `#EXTM3U
#EXT-X-VERSION:7
#EXT-X-STREAM-INF:BANDWIDTH=800000,RESOLUTION=1280x720,CODECS="avc1.640015"
video_720.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=1000000,RESOLUTION=1920x1080,CODECS="avc1.640015"
job_1234.m3u8
`

	oldManifest := `#EXTM3U
#EXT-X-VERSION:7
#EXT-X-STREAM-INF:BANDWIDTH=1000000,RESOLUTION=1920x1080,CODECS="avc1.640015"
video_1080.m3u8
`

	expected := `#EXTM3U
#EXT-X-VERSION:7
#EXT-X-STREAM-INF:BANDWIDTH=800000,RESOLUTION=1280x720,CODECS="avc1.640015"
video_720.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=1000000,RESOLUTION=1920x1080,CODECS="avc1.640015"
video_1080.m3u8
`

	result := mergeManifest([]byte(currentManifest), []byte(oldManifest))
	actual := string(result)

	if actual != expected {
		t.Errorf("Expected:\n%q\n\nGot:\n%q", expected, actual)
	}
}

func TestMergeManifest_NoMatch(t *testing.T) {
	// Example m3u8 content from the user
	currentManifest := `#EXTM3U
#EXT-X-VERSION:7
#EXT-X-STREAM-INF:BANDWIDTH=800000,RESOLUTION=1280x720,CODECS="avc1.640015"
video_720.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=1000000,RESOLUTION=1920x1080,CODECS="avc1.640015"
job_1234.m3u8
`

	// Different stream info that won't match
	oldManifest := `#EXTM3U
#EXT-X-VERSION:7
#EXT-X-STREAM-INF:BANDWIDTH=200000,RESOLUTION=640x360,CODECS="avc1.640028"
video_360.m3u8
`

	// Should remain unchanged since stream infos don't match
	expected := `#EXTM3U
#EXT-X-VERSION:7
#EXT-X-STREAM-INF:BANDWIDTH=800000,RESOLUTION=1280x720,CODECS="avc1.640015"
video_720.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=1000000,RESOLUTION=1920x1080,CODECS="avc1.640015"
job_1234.m3u8
`

	result := mergeManifest([]byte(currentManifest), []byte(oldManifest))
	actual := string(result)

	if actual != expected {
		t.Errorf("Expected:\n%q\n\nGot:\n%q", expected, actual)
	}
}
