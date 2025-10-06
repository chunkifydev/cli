package chunkify

import (
	"testing"
)

// Helper function to reset all global variables to nil
func resetGlobalVars() {
	width = nil
	height = nil
	framerate = nil
	gop = nil
	channels = nil
	pixfmt = nil
	maxrate = nil
	bufsize = nil
	videoBitrate = nil
	audioBitrate = nil
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

func TestValidateCommonVideoFlags(t *testing.T) {
	tests := []struct {
		name        string
		setup       func()
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_width",
			setup: func() {
				resetGlobalVars()
				val := int64(1920)
				width = &val
			},
			expectError: false,
		},
		{
			name: "invalid_width_negative",
			setup: func() {
				resetGlobalVars()
				val := int64(-100)
				width = &val
			},
			expectError: true,
			errorMsg:    "--resolution width must be between 0 and 8192",
		},
		{
			name: "invalid_width_too_large",
			setup: func() {
				resetGlobalVars()
				val := int64(10000)
				width = &val
			},
			expectError: true,
			errorMsg:    "--resolution width must be between 0 and 8192",
		},
		{
			name: "valid_height",
			setup: func() {
				resetGlobalVars()
				val := int64(1080)
				height = &val
			},
			expectError: false,
		},
		{
			name: "invalid_height_negative",
			setup: func() {
				resetGlobalVars()
				val := int64(-50)
				height = &val
			},
			expectError: true,
			errorMsg:    "--resolution height must be between 0 and 8192",
		},
		{
			name: "valid_video_bitrate",
			setup: func() {
				resetGlobalVars()
				val := int64(2000000)
				videoBitrate = &val
			},
			expectError: false,
		},
		{
			name: "invalid_video_bitrate_too_low",
			setup: func() {
				resetGlobalVars()
				val := int64(50000)
				videoBitrate = &val
			},
			expectError: true,
			errorMsg:    "--vb must be between 100000 and 50000000",
		},
		{
			name: "invalid_video_bitrate_too_high",
			setup: func() {
				resetGlobalVars()
				val := int64(60000000)
				videoBitrate = &val
			},
			expectError: true,
			errorMsg:    "--vb must be between 100000 and 50000000",
		},
		{
			name: "valid_audio_bitrate",
			setup: func() {
				resetGlobalVars()
				val := int64(128000)
				audioBitrate = &val
			},
			expectError: false,
		},
		{
			name: "invalid_audio_bitrate_too_low",
			setup: func() {
				resetGlobalVars()
				val := int64(16000)
				audioBitrate = &val
			},
			expectError: true,
			errorMsg:    "--ab must be between 32000 and 512000",
		},
		{
			name: "invalid_audio_bitrate_too_high",
			setup: func() {
				resetGlobalVars()
				val := int64(600000)
				audioBitrate = &val
			},
			expectError: true,
			errorMsg:    "--ab must be between 32000 and 512000",
		},
		{
			name: "valid_framerate",
			setup: func() {
				resetGlobalVars()
				val := float64(30.0)
				framerate = &val
			},
			expectError: false,
		},
		{
			name: "invalid_framerate_too_low",
			setup: func() {
				resetGlobalVars()
				val := float64(10.0)
				framerate = &val
			},
			expectError: true,
			errorMsg:    "--framerate must be between 15 and 120",
		},
		{
			name: "invalid_framerate_too_high",
			setup: func() {
				resetGlobalVars()
				val := float64(150.0)
				framerate = &val
			},
			expectError: true,
			errorMsg:    "--framerate must be between 15 and 120",
		},
		{
			name: "valid_gop",
			setup: func() {
				resetGlobalVars()
				val := int64(30)
				gop = &val
			},
			expectError: false,
		},
		{
			name: "invalid_gop_too_low",
			setup: func() {
				resetGlobalVars()
				val := int64(-1)
				gop = &val
			},
			expectError: true,
			errorMsg:    "--gop must be between 1 and 30",
		},
		{
			name: "invalid_gop_too_high",
			setup: func() {
				resetGlobalVars()
				val := int64(400)
				gop = &val
			},
			expectError: true,
			errorMsg:    "--gop must be between 1 and 30",
		},
		{
			name: "valid_channels",
			setup: func() {
				resetGlobalVars()
				val := int64(6)
				channels = &val
			},
			expectError: false,
		},
		{
			name: "invalid_channels_valid_value",
			setup: func() {
				resetGlobalVars()
				val := int64(2)
				channels = &val
			},
			expectError: true,
			errorMsg:    "--channels must be one of 1, 2, 5, 7",
		},
		{
			name: "valid_pixfmt",
			setup: func() {
				resetGlobalVars()
				val := "yuv420p"
				pixfmt = &val
			},
			expectError: false,
		},
		{
			name: "invalid_pixfmt",
			setup: func() {
				resetGlobalVars()
				val := "invalid_format"
				pixfmt = &val
			},
			expectError: true,
			errorMsg:    "--pixfmt must be one of yuv410p, yuv411p, yuv420p, yuv422p, yuv440p, yuv444p, yuvJ411p, yuvJ420p, yuvJ422p, yuvJ440p, yuvJ444p, yuv420p10le, yuv422p10le, yuv440p10le, yuv444p10le, yuv420p12le, yuv422p12le, yuv440p12le, yuv444p12le, yuv420p10be, yuv422p10be, yuv440p10be, yuv444p10be, yuv420p12be, yuv422p12be, yuv440p12be, yuv444p12be",
		},
		{
			name: "all_valid_values",
			setup: func() {
				resetGlobalVars()
				widthVal := int64(1920)
				heightVal := int64(1080)
				videoBitrateVal := int64(2000000)
				audioBitrateVal := int64(128000)
				framerateVal := float64(30.0)
				gopVal := int64(30)
				channelsVal := int64(6)
				maxrateVal := int64(2500000)
				bufsizeVal := int64(5000000)
				pixfmtVal := "yuv420p"

				width = &widthVal
				height = &heightVal
				videoBitrate = &videoBitrateVal
				audioBitrate = &audioBitrateVal
				framerate = &framerateVal
				gop = &gopVal
				channels = &channelsVal
				maxrate = &maxrateVal
				bufsize = &bufsizeVal
				pixfmt = &pixfmtVal
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := validateCommonVideoFlags()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error message %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateH264Flags(t *testing.T) {
	tests := []struct {
		name        string
		setup       func()
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_crf",
			setup: func() {
				resetGlobalVars()
				val := int64(23)
				crf = &val
			},
			expectError: false,
		},
		{
			name: "invalid_crf_too_low",
			setup: func() {
				resetGlobalVars()
				val := int64(10)
				crf = &val
			},
			expectError: true,
			errorMsg:    "--crf must be between 16 and 35",
		},
		{
			name: "invalid_crf_too_high",
			setup: func() {
				resetGlobalVars()
				val := int64(40)
				crf = &val
			},
			expectError: true,
			errorMsg:    "--crf must be between 16 and 35",
		},
		{
			name: "valid_preset",
			setup: func() {
				resetGlobalVars()
				val := "fast"
				preset = &val
			},
			expectError: false,
		},
		{
			name: "invalid_preset",
			setup: func() {
				resetGlobalVars()
				val := "invalid"
				preset = &val
			},
			expectError: true,
			errorMsg:    "--preset must be one of ultrafast, superfast, veryfast, faster, fast, medium, slow, slower, veryslow",
		},
		{
			name: "valid_profilev",
			setup: func() {
				resetGlobalVars()
				val := "high"
				profilev = &val
			},
			expectError: false,
		},
		{
			name: "invalid_profilev",
			setup: func() {
				resetGlobalVars()
				val := "invalid"
				profilev = &val
			},
			expectError: true,
			errorMsg:    "--profilev must be one of baseline, main, high, high10, high422, high444",
		},
		{
			name: "valid_level",
			setup: func() {
				resetGlobalVars()
				val := int64(40)
				level = &val
			},
			expectError: false,
		},
		{
			name: "invalid_level_valid_value",
			setup: func() {
				resetGlobalVars()
				val := int64(9)
				level = &val
			},
			expectError: true,
			errorMsg:    "--level must be one of 10, 11, 12, 13, 20, 21, 22, 30, 31, 32, 40, 41, 42, 50, 51",
		},
		{
			name: "valid_x264keyint",
			setup: func() {
				resetGlobalVars()
				val := int64(30)
				x264KeyInt = &val
			},
			expectError: false,
		},
		{
			name: "invalid_x264keyint_too_low",
			setup: func() {
				resetGlobalVars()
				val := int64(-1)
				x264KeyInt = &val
			},
			expectError: true,
			errorMsg:    "--x264keyint must be between 1 and 30",
		},
		{
			name: "invalid_x264keyint_too_high",
			setup: func() {
				resetGlobalVars()
				val := int64(400)
				x264KeyInt = &val
			},
			expectError: true,
			errorMsg:    "--x264keyint must be between 1 and 30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := validateH264Flags()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error message %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateH265Flags(t *testing.T) {
	tests := []struct {
		name        string
		setup       func()
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_h265_profilev",
			setup: func() {
				resetGlobalVars()
				val := "main10"
				profilev = &val
			},
			expectError: false,
		},
		{
			name: "invalid_h265_profilev",
			setup: func() {
				resetGlobalVars()
				val := "baseline"
				profilev = &val
			},
			expectError: true,
			errorMsg:    "--profilev must be one of baseline, main, high, high10, high422, high444",
		},
		{
			name: "valid_h265_level",
			setup: func() {
				resetGlobalVars()
				val := int64(31)
				level = &val
			},
			expectError: false,
		},
		{
			name: "invalid_h265_level_valid_value",
			setup: func() {
				resetGlobalVars()
				val := int64(28)
				level = &val
			},
			expectError: true,
			errorMsg:    "--level must be one of 30, 31, 41",
		},
		{
			name: "valid_x265keyint",
			setup: func() {
				resetGlobalVars()
				val := int64(30)
				x265KeyInt = &val
			},
			expectError: false,
		},
		{
			name: "invalid_x265keyint_too_high",
			setup: func() {
				resetGlobalVars()
				val := int64(400)
				x265KeyInt = &val
			},
			expectError: true,
			errorMsg:    "--x265keyint must be between 1 and 30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := validateH265Flags()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error message %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateAv1Flags(t *testing.T) {
	tests := []struct {
		name        string
		setup       func()
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_av1_crf",
			setup: func() {
				resetGlobalVars()
				val := int64(30)
				crf = &val
			},
			expectError: false,
		},
		{
			name: "invalid_av1_crf_too_high",
			setup: func() {
				resetGlobalVars()
				val := int64(70)
				crf = &val
			},
			expectError: true,
			errorMsg:    "--crf must be between 16 and 63",
		},
		{
			name: "valid_av1_preset",
			setup: func() {
				resetGlobalVars()
				val := "8"
				preset = &val
			},
			expectError: false,
		},
		{
			name: "invalid_av1_preset",
			setup: func() {
				resetGlobalVars()
				val := "5"
				preset = &val
			},
			expectError: true,
			errorMsg:    "--preset must be one of 6, 7, 8, 9, 10, 11, 12, 13",
		},
		{
			name: "valid_av1_profilev",
			setup: func() {
				resetGlobalVars()
				val := "main"
				profilev = &val
			},
			expectError: false,
		},
		{
			name: "invalid_av1_profilev",
			setup: func() {
				resetGlobalVars()
				val := "baseline"
				profilev = &val
			},
			expectError: true,
			errorMsg:    "--profilev must be one of main, main10, mainstillpicture",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := validateAv1Flags()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error message %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateWebmVp9Flags(t *testing.T) {
	tests := []struct {
		name        string
		setup       func()
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_vp9_crf",
			setup: func() {
				resetGlobalVars()
				val := int64(25)
				crf = &val
			},
			expectError: false,
		},
		{
			name: "invalid_vp9_crf_too_low",
			setup: func() {
				resetGlobalVars()
				val := int64(10)
				crf = &val
			},
			expectError: true,
			errorMsg:    "--crf must be between 15 and 35",
		},
		{
			name: "valid_quality",
			setup: func() {
				resetGlobalVars()
				val := "good"
				quality = &val
			},
			expectError: false,
		},
		{
			name: "invalid_quality",
			setup: func() {
				resetGlobalVars()
				val := "invalid"
				quality = &val
			},
			expectError: true,
			errorMsg:    "--quality must be one of good, best, realtime",
		},
		{
			name: "valid_cpu_used",
			setup: func() {
				resetGlobalVars()
				val := "4"
				cpuUsed = &val
			},
			expectError: false,
		},
		{
			name: "invalid_cpu_used",
			setup: func() {
				resetGlobalVars()
				val := "10"
				cpuUsed = &val
			},
			expectError: true,
			errorMsg:    "--cpu-used must be one of 0, 1, 2, 3, 4, 5, 6, 7, 8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := validateWebmVp9Flags()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error message %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateHlsFlags(t *testing.T) {
	tests := []struct {
		name        string
		setup       func()
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_with_video_bitrate",
			setup: func() {
				resetGlobalVars()
				val := int64(2000000)
				videoBitrate = &val
			},
			expectError: false,
		},
		{
			name: "valid_with_audio_bitrate",
			setup: func() {
				resetGlobalVars()
				val := int64(128000)
				audioBitrate = &val
			},
			expectError: false,
		},
		{
			name: "invalid_no_bitrates",
			setup: func() {
				resetGlobalVars()
			},
			expectError: true,
			errorMsg:    "--vb (video bitrate) or --ab (audio bitrate) are required when format is hls",
		},
		{
			name: "valid_hls_time",
			setup: func() {
				resetGlobalVars()
				videoBitrateVal := int64(2000000)
				hlsTimeVal := int64(5)
				videoBitrate = &videoBitrateVal
				hlsTime = &hlsTimeVal
			},
			expectError: false,
		},
		{
			name: "invalid_hls_time_too_low",
			setup: func() {
				resetGlobalVars()
				videoBitrateVal := int64(2000000)
				hlsTimeVal := int64(-1)
				videoBitrate = &videoBitrateVal
				hlsTime = &hlsTimeVal
			},
			expectError: true,
			errorMsg:    "--hls-time must be between 1 and 10",
		},
		{
			name: "invalid_hls_time_too_high",
			setup: func() {
				resetGlobalVars()
				videoBitrateVal := int64(2000000)
				hlsTimeVal := int64(15)
				videoBitrate = &videoBitrateVal
				hlsTime = &hlsTimeVal
			},
			expectError: true,
			errorMsg:    "--hls-time must be between 1 and 10",
		},
		{
			name: "valid_hls_segment_type",
			setup: func() {
				resetGlobalVars()
				videoBitrateVal := int64(2000000)
				segmentTypeVal := "mpegts"
				videoBitrate = &videoBitrateVal
				hlsSegmentType = &segmentTypeVal
			},
			expectError: false,
		},
		{
			name: "invalid_hls_segment_type",
			setup: func() {
				resetGlobalVars()
				videoBitrateVal := int64(2000000)
				segmentTypeVal := "invalid"
				videoBitrate = &videoBitrateVal
				hlsSegmentType = &segmentTypeVal
			},
			expectError: true,
			errorMsg:    "--hls-segment-type must be one of mpegts, fmp4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := validateHlsFlags()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error message %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateJpgFlags(t *testing.T) {
	tests := []struct {
		name        string
		setup       func()
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_interval",
			setup: func() {
				resetGlobalVars()
				val := int64(5)
				interval = &val
			},
			expectError: false,
		},
		{
			name: "invalid_interval_negative",
			setup: func() {
				resetGlobalVars()
				val := int64(-1)
				interval = &val
			},
			expectError: true,
			errorMsg:    "--interval must be between 0 and 60",
		},
		{
			name: "invalid_interval_too_high",
			setup: func() {
				resetGlobalVars()
				val := int64(70)
				interval = &val
			},
			expectError: true,
			errorMsg:    "--interval must be between 0 and 60",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := validateJpgFlags()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error message %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateHlsH264Flags(t *testing.T) {
	tests := []struct {
		name        string
		setup       func()
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_hls_h264",
			setup: func() {
				resetGlobalVars()
				videoBitrateVal := int64(2000000)
				crfVal := int64(23)
				videoBitrate = &videoBitrateVal
				crf = &crfVal
			},
			expectError: false,
		},
		{
			name: "invalid_hls_h264_no_bitrate",
			setup: func() {
				resetGlobalVars()
				crfVal := int64(23)
				crf = &crfVal
			},
			expectError: true,
			errorMsg:    "--vb (video bitrate) or --ab (audio bitrate) are required when format is hls",
		},
		{
			name: "invalid_hls_h264_invalid_crf",
			setup: func() {
				resetGlobalVars()
				videoBitrateVal := int64(2000000)
				crfVal := int64(10)
				videoBitrate = &videoBitrateVal
				crf = &crfVal
			},
			expectError: true,
			errorMsg:    "--crf must be between 16 and 35",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := validateHlsH264Flags()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error message %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}
