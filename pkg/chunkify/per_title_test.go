package chunkify

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPerTitleVideoFormats(t *testing.T) {
	for _, format := range []string{
		FormatMp4H264, FormatMp4H265, FormatMp4Av1, FormatWebmVp9,
		FormatHlsH264, FormatHlsH265, FormatHlsAv1,
	} {
		t.Run(format, func(t *testing.T) {
			for _, enabled := range []bool{true, false} {
				resetGlobalFlags()
				perTitle = &enabled
				// Flag binding initializes CRF to zero, even when it is omitted.
				crf = new(int64)
				app := &App{Command: &ChunkifyCommand{Format: format}}
				if enabled {
					if err := validateTranscodeSettings(app); err != nil {
						t.Fatalf("per-title without manual rate control: %v", err)
					}
				}
				setJobFormatParams(app)
				data, err := json.Marshal(app.Command.JobFormatParams)
				if err != nil {
					t.Fatal(err)
				}
				var payload map[string]any
				if err := json.Unmarshal(data, &payload); err != nil {
					t.Fatal(err)
				}
				value, present := payload["per_title"]
				if enabled && value != true {
					t.Errorf("expected per_title=true, got %s", data)
				}
				if !enabled && present {
					t.Errorf("disabled per_title should be omitted, got %s", data)
				}
			}
		})
	}
	t.Cleanup(resetGlobalFlags)
}

func TestPerTitleValidation(t *testing.T) {
	for _, tt := range []struct {
		name   string
		format string
		setup  func()
		want   string
	}{
		{"crf", FormatMp4H264, func() { v := int64(21); crf = &v }, "--crf"},
		{"video bitrate", FormatMp4H264, func() { v := int64(2000000); videoBitrate = &v }, "--vb"},
		{"maxrate", FormatMp4H264, func() { v := int64(2000000); maxrate = &v }, "--maxrate"},
		{"bufsize", FormatMp4H264, func() { v := int64(2000000); bufsize = &v }, "--bufsize"},
		{"jpg", FormatJpg, func() {}, "only supported for video formats"},
		{"audio bitrate", FormatHlsH264, func() { v := int64(128000); audioBitrate = &v }, ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalFlags()
			enabled := true
			perTitle = &enabled
			tt.setup()
			err := validateTranscodeSettings(&App{Command: &ChunkifyCommand{Format: tt.format}})
			if tt.want == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("expected error containing %q, got %v", tt.want, err)
			}
		})
	}
	t.Cleanup(resetGlobalFlags)
}

func TestHlsRequiresBitrateWithoutPerTitle(t *testing.T) {
	for _, format := range []string{FormatHlsH264, FormatHlsH265, FormatHlsAv1} {
		t.Run(format, func(t *testing.T) {
			for _, value := range []*bool{nil, new(bool)} {
				resetGlobalFlags()
				perTitle = value
				err := validateTranscodeSettings(&App{Command: &ChunkifyCommand{Format: format}})
				if err == nil || !strings.Contains(err.Error(), "required when format is hls") {
					t.Errorf("expected missing bitrate error, got %v", err)
				}
			}
		})
	}
	t.Cleanup(resetGlobalFlags)
}
