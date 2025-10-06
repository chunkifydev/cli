package chunkify

import (
	"testing"

	chunkify "github.com/chunkifydev/chunkify-go"
)

// Helper function to reset all global variables to nil
func resetGlobalFlags() {
	transcoders = nil
	transcoderVcpu = nil
	storagePath = nil
	resolution = nil
	width = nil
	height = nil
	framerate = nil
	gop = nil
	channels = nil
	pixfmt = nil
	disableAudio = nil
	disableVideo = nil
	duration = nil
	seek = nil
	maxrate = nil
	bufsize = nil
	videoBitrate = nil
	audioBitrate = nil
	maxrateStr = nil
	bufsizeStr = nil
	videoBitrateStr = nil
	audioBitrateStr = nil
	crf = nil
	preset = nil
	profilev = nil
	level = nil
	x264KeyInt = nil
	x265KeyInt = nil
	quality = nil
	cpuUsed = nil
	hlsTime = nil
	hlsSegmentType = nil
	interval = nil
}

func TestSetupCommand_NoFormatOrOutput(t *testing.T) {
	resetGlobalFlags()

	app := &App{
		Command: &ChunkifyCommand{
			Format: "",
			Output: "",
		},
	}

	err := setupCommand(app)
	if err != nil {
		t.Errorf("Expected no error when no format or output specified, got: %v", err)
	}
}

func TestSetupCommand_FormatFromOutputExtension(t *testing.T) {
	tests := []struct {
		name           string
		output         string
		expectedFormat string
	}{
		{
			name:           "mp4_extension",
			output:         "video.mp4",
			expectedFormat: string(chunkify.FormatMp4H264),
		},
		{
			name:           "webm_extension",
			output:         "video.webm",
			expectedFormat: string(chunkify.FormatWebmVp9),
		},
		{
			name:           "m3u8_extension",
			output:         "playlist.m3u8",
			expectedFormat: string(chunkify.FormatHlsH264),
		},
		{
			name:           "jpg_extension",
			output:         "image.jpg",
			expectedFormat: string(chunkify.FormatJpg),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalFlags()

			app := &App{
				Command: &ChunkifyCommand{
					Format: "",
					Output: tt.output,
				},
			}

			err := setupCommand(app)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if app.Command.Format != tt.expectedFormat {
				t.Errorf("Expected format %s, got %s", tt.expectedFormat, app.Command.Format)
			}
		})
	}
}

func TestSetupCommand_InvalidOutputExtension(t *testing.T) {
	resetGlobalFlags()

	app := &App{
		Command: &ChunkifyCommand{
			Format: "",
			Output: "video.avi",
		},
	}

	err := setupCommand(app)
	if err == nil {
		t.Error("Expected error for invalid output extension")
	}

	expectedError := "invalid output file extension: .avi. Please provide a valid format with --format"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}

func TestSetupCommand_InvalidFormat(t *testing.T) {
	resetGlobalFlags()

	app := &App{
		Command: &ChunkifyCommand{
			Format: "invalid_format",
			Output: "video.mp4",
		},
	}

	setupCommand(app)
	err := validateTranscodeSettings(app)
	if err == nil {
		t.Errorf("Expected error for invalid format, got: %v", err)
	}

	expectedError := "invalid format: invalid_format"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}

func TestSetupCommand_ValidFormats(t *testing.T) {
	validFormats := []string{
		string(chunkify.FormatMp4H264),
		string(chunkify.FormatMp4H265),
		string(chunkify.FormatMp4Av1),
		string(chunkify.FormatWebmVp9),
		string(chunkify.FormatHlsH264),
		string(chunkify.FormatHlsH265),
		string(chunkify.FormatHlsAv1),
		string(chunkify.FormatJpg),
	}

	for _, format := range validFormats {
		t.Run("format_"+format, func(t *testing.T) {
			resetGlobalFlags()

			app := &App{
				Command: &ChunkifyCommand{
					Format: format,
					Output: "output.mp4",
				},
			}

			err := setupCommand(app)
			if err != nil {
				t.Errorf("Unexpected error for valid format %s: %v", format, err)
			}

			if app.Command.Format != format {
				t.Errorf("Expected format %s, got %s", format, app.Command.Format)
			}
		})
	}
}

func TestSetupCommand_ResolutionParsing(t *testing.T) {
	tests := []struct {
		name           string
		resolution     string
		expectedWidth  int64
		expectedHeight int64
		expectError    bool
	}{
		{
			name:           "valid_resolution",
			resolution:     "1920x1080",
			expectedWidth:  1920,
			expectedHeight: 1080,
			expectError:    false,
		},
		{
			name:           "small_resolution",
			resolution:     "640x480",
			expectedWidth:  640,
			expectedHeight: 480,
			expectError:    false,
		},
		{
			name:        "invalid_format",
			resolution:  "1920",
			expectError: true,
		},
		{
			name:        "invalid_width",
			resolution:  "abcx1080",
			expectError: true,
		},
		{
			name:        "invalid_height",
			resolution:  "1920xabc",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalFlags()

			// Set resolution flag
			resolution = &tt.resolution

			app := &App{
				Command: &ChunkifyCommand{
					Format: string(chunkify.FormatMp4H264),
					Output: "output.mp4",
				},
			}

			err := setupCommand(app)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				if width == nil || *width != tt.expectedWidth {
					t.Errorf("Expected width %d, got %v", tt.expectedWidth, width)
				}

				if height == nil || *height != tt.expectedHeight {
					t.Errorf("Expected height %d, got %v", tt.expectedHeight, height)
				}
			}
		})
	}
}

func TestSetupCommand_BitrateParsing(t *testing.T) {
	tests := []struct {
		name        string
		bitrateStr  string
		expectError bool
	}{
		{
			name:        "valid_kilobitrate",
			bitrateStr:  "1200K",
			expectError: false,
		},
		{
			name:        "valid_megabitrate",
			bitrateStr:  "2M",
			expectError: false,
		},
		{
			name:        "invalid_bitrate",
			bitrateStr:  "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalFlags()

			// Set video bitrate string flag
			videoBitrateStr = &tt.bitrateStr

			app := &App{
				Command: &ChunkifyCommand{
					Format: string(chunkify.FormatMp4H264),
					Output: "output.mp4",
				},
			}

			err := setupCommand(app)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				if videoBitrate == nil {
					t.Error("Expected videoBitrate to be set")
				}
			}
		})
	}
}

func TestSetupCommand_TranscoderParams(t *testing.T) {
	resetGlobalFlags()

	// Set transcoder flags
	transcoderVal := int64(2)
	vcpuVal := int64(4)
	transcoders = &transcoderVal
	transcoderVcpu = &vcpuVal

	app := &App{
		Command: &ChunkifyCommand{
			Format: string(chunkify.FormatMp4H264),
			Output: "output.mp4",
		},
	}

	err := setupCommand(app)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if app.Command.JobTranscoderParams == nil {
		t.Error("Expected JobTranscoderParams to be set")
	}

	if app.Command.JobTranscoderParams.Quantity != 2 {
		t.Errorf("Expected Quantity 2, got %d", app.Command.JobTranscoderParams.Quantity)
	}

	if app.Command.JobTranscoderParams.Type != "4vCPU" {
		t.Errorf("Expected Type '4vCPU', got %s", app.Command.JobTranscoderParams.Type)
	}
}

func TestSetupCommand_StorageParams(t *testing.T) {
	resetGlobalFlags()

	// Set storage path flag
	storageVal := "/path/to/storage"
	storagePath = &storageVal

	app := &App{
		Command: &ChunkifyCommand{
			Format: string(chunkify.FormatMp4H264),
			Output: "output.mp4",
		},
	}

	err := setupCommand(app)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if app.Command.JobCreateStorageParams == nil {
		t.Error("Expected JobCreateStorageParams to be set")
	}

	if app.Command.JobCreateStorageParams.Path == nil || *app.Command.JobCreateStorageParams.Path != storageVal {
		t.Errorf("Expected Path %s, got %v", storageVal, app.Command.JobCreateStorageParams.Path)
	}
}
