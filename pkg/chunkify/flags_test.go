package chunkify

import (
	"encoding/json"
	"reflect"
	"testing"
)

// Helper function to reset all global variables to nil
func resetGlobalFlags() {
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
	perTitle = nil
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
	audioSamplerate = nil
	crf = nil
	preset = nil
	profilev = nil
	level = nil
	x264KeyInt = nil
	x265KeyInt = nil
	quality = nil
	cpuUsed = nil
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

func TestSetupCommandOutputStorage(t *testing.T) {
	resetGlobalFlags()
	t.Cleanup(resetGlobalFlags)
	value := "manual/output.mp4"
	storagePath = &value
	app := &App{Command: &ChunkifyCommand{Format: FormatMp4H264, OutputStorageID: "stor_aws_output"}}
	if err := setupCommand(app); err != nil {
		t.Fatal(err)
	}
	if got := app.Command.TranscodingStorageParams; got.ID.Value != "stor_aws_output" || got.Path.Value != value {
		t.Fatalf("output storage flags were not mapped to job parameters: %+v", got)
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
			expectedFormat: string(FormatMp4H264),
		},
		{
			name:           "webm_extension",
			output:         "video.webm",
			expectedFormat: string(FormatWebmVp9),
		},
		{
			name:           "m4a_extension",
			output:         "audio.m4a",
			expectedFormat: string(FormatM4aAac),
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

	expectedError := "invalid format: invalid_format. Supported formats: mp4_h264, mp4_h265, mp4_av1, webm_vp9, m4a_aac"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}

func TestSetupCommand_ValidFormats(t *testing.T) {
	validFormats := []string{
		FormatMp4H264,
		FormatMp4H265,
		FormatMp4Av1,
		FormatWebmVp9,
		FormatM4aAac,
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
					Format: FormatMp4H264,
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
					Format: FormatMp4H264,
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

func TestSetupCommand_RemovedFormatsFromExtension(t *testing.T) {
	for _, output := range []string{"playlist.m3u8", "image.jpg"} {
		resetGlobalFlags()
		app := &App{Command: &ChunkifyCommand{Output: output}}
		if err := setupCommand(app); err == nil {
			t.Errorf("%s: expected unsupported extension error", output)
		}
	}
}

func TestSetupCommand_RemovedFormats(t *testing.T) {
	for _, format := range []string{"hls_h264", "hls_h265", "hls_av1", "jpg"} {
		resetGlobalFlags()
		app := &App{Command: &ChunkifyCommand{Format: format, Output: "output.mp4"}}
		if err := validateTranscodeSettings(app); err == nil {
			t.Errorf("%s: expected invalid format error", format)
		}
	}
}

func TestSetTranscodingFormatParams(t *testing.T) {
	t.Cleanup(resetGlobalFlags)
	for _, tt := range []struct {
		format string
		want   string
	}{
		{FormatMp4H264, `{"crf":21,"duration":12.5,"preset":"fast","seek":1.5,"width":1280,"x264_keyint":48,"audio_samplerate":44100,"channels":2,"id":"mp4_h264"}`},
		{FormatWebmVp9, `{"crf":21,"duration":12.5,"seek":1.5,"width":1280,"audio_samplerate":44100,"channels":2,"cpu_used":"4","id":"webm_vp9"}`},
		{FormatM4aAac, `{"audio_bitrate":192000,"duration":12.5,"seek":1.5,"audio_samplerate":44100,"channels":2,"id":"m4a_aac"}`},
	} {
		t.Run(tt.format, func(t *testing.T) {
			resetGlobalFlags()
			w, c, k, ch, sr := int64(1280), int64(21), int64(48), int64(2), int64(44100)
			d, sk := 12.5, 1.5
			p, cu := "fast", "4"
			width, crf, x264KeyInt, channels, audioSamplerate = &w, &c, &k, &ch, &sr
			duration, seek = &d, &sk
			preset, cpuUsed = &p, &cu
			if tt.format == FormatM4aAac {
				ab := int64(192000)
				width, crf, preset, audioBitrate = nil, nil, nil, &ab
			} else if tt.format == FormatWebmVp9 {
				preset = nil
			}
			app := &App{Command: &ChunkifyCommand{Format: tt.format}}
			setTranscodingFormatParams(app)
			data, err := json.Marshal(app.Command.TranscodingFormatParams)
			if err != nil {
				t.Fatal(err)
			}
			var got, want map[string]any
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(tt.want), &want); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("payload = %s\nwant      %s", data, tt.want)
			}
		})
	}
}

func TestSetupCommand_StorageParams(t *testing.T) {
	resetGlobalFlags()

	// Set storage path flag
	storageVal := "/path/to/storage"
	storagePath = &storageVal

	app := &App{
		Command: &ChunkifyCommand{
			Format: FormatMp4H264,
			Output: "output.mp4",
		},
	}

	err := setupCommand(app)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if app.Command.TranscodingStorageParams.Path.Value != storageVal {
		t.Errorf("Expected Path %s, got %v", storageVal, app.Command.TranscodingStorageParams.Path.Value)
	}
}
