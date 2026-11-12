package chunkify

import (
	"fmt"
	"slices"
	"strings"
)

func validateCommonVideoFlags() error {
	if width != nil && *width != 0 {
		if *width < 0 || *width > 7680 {
			return fmt.Errorf("--resolution width must be between 0 and 7680")
		}
	}
	if height != nil && *height != 0 {
		if *height < 0 || *height > 7680 {
			return fmt.Errorf("--resolution height must be between 0 and 7680")
		}
	}

	if videoBitrate != nil && *videoBitrate != 0 {
		if *videoBitrate < 100000 || *videoBitrate > 50000000 {
			return fmt.Errorf("--vb must be between 100000 and 50000000")
		}
	}
	if audioBitrate != nil && *audioBitrate != 0 {
		if *audioBitrate < 32000 || *audioBitrate > 512000 {
			return fmt.Errorf("--ab must be between 32000 and 512000")
		}
	}

	if framerate != nil && *framerate != 0 {
		if *framerate < 0.001 || *framerate > 240 {
			return fmt.Errorf("--framerate must be between 0.001 and 240")
		}
	}
	if gop != nil && *gop != 0 {
		if *gop < 1 || *gop > 960 {
			return fmt.Errorf("--gop must be between 1 and 960")
		}
	}
	if channels != nil && *channels != 0 {
		if !slices.Contains([]int64{1, 2, 6, 8}, *channels) {
			return fmt.Errorf("--channels must be one of 1, 2, 6, 8")
		}
	}
	if audioSamplerate != nil && *audioSamplerate != 0 {
		if !slices.Contains(validAudioSamplerates, *audioSamplerate) {
			return fmt.Errorf("--ar must be one of 8000, 11025, 12000, 16000, 22050, 24000, 32000, 44100, 48000, 64000, 88200, 96000")
		}
	}
	if duration != nil && *duration < 0 {
		return fmt.Errorf("--duration must be positive")
	}
	if seek != nil && *seek < 0 {
		return fmt.Errorf("--seek must be positive")
	}
	if maxrate != nil && *maxrate != 0 {
		if *maxrate < 100000 || *maxrate > 50000000 {
			return fmt.Errorf("--maxrate must be between 100000 and 50000000")
		}
	}
	if bufsize != nil && *bufsize != 0 {
		if *bufsize < 100000 || *bufsize > 50000000 {
			return fmt.Errorf("--bufsize must be between 100000 and 50000000")
		}
	}
	return nil
}

var validAudioSamplerates = []int64{8000, 11025, 12000, 16000, 22050, 24000, 32000, 44100, 48000, 64000, 88200, 96000}

func validateOneOf(flag string, value *string, valid []string) error {
	if value == nil || *value == "" || slices.Contains(valid, *value) {
		return nil
	}
	return fmt.Errorf("--%s must be one of %s", flag, strings.Join(valid, ", "))
}

func validatePerTitleFlags(format string) error {
	if perTitle == nil || !*perTitle {
		return nil
	}
	if format == FormatM4aAac {
		return fmt.Errorf("--per-title is only supported for video formats")
	}
	for _, flag := range []struct {
		name  string
		value *int64
	}{
		{"crf", crf},
		{"vb", videoBitrate},
		{"maxrate", maxrate},
		{"bufsize", bufsize},
	} {
		if flag.value != nil && *flag.value != 0 {
			return fmt.Errorf("--per-title cannot be used with --%s", flag.name)
		}
	}
	return nil
}

func validateH264Flags() error {
	if crf != nil && ((*crf > 0 && *crf < 16) || *crf > 35) {
		return fmt.Errorf("--crf must be between 16 and 35")
	}
	if err := validateOneOf("preset", preset, x26xPresets); err != nil {
		return err
	}
	if err := validateOneOf("profilev", profilev, []string{"baseline", "main", "high", "high10", "high422", "high444"}); err != nil {
		return err
	}
	if err := validateOneOf("pixfmt", pixfmt, []string{"yuv420p", "yuv422p", "yuv444p", "yuv420p10le", "yuv422p10le", "yuv444p10le"}); err != nil {
		return err
	}

	if level != nil && *level != 0 {
		if !slices.Contains([]int64{10, 11, 12, 13, 20, 21, 22, 30, 31, 32, 40, 41, 42, 50, 51}, *level) {
			return fmt.Errorf("--level must be one of 10, 11, 12, 13, 20, 21, 22, 30, 31, 32, 40, 41, 42, 50, 51")
		}
	}
	if x264KeyInt != nil && *x264KeyInt != 0 {
		if *x264KeyInt < 1 || *x264KeyInt > 960 {
			return fmt.Errorf("--x264keyint must be between 1 and 960")
		}
	}

	return nil
}

func validateH265Flags() error {
	if crf != nil && ((*crf > 0 && *crf < 16) || *crf > 35) {
		return fmt.Errorf("--crf must be between 16 and 35")
	}
	if err := validateOneOf("preset", preset, x26xPresets); err != nil {
		return err
	}
	if err := validateOneOf("profilev", profilev, []string{"main", "main10", "main12", "main422-10", "main422-12", "main444-8", "main444-10", "main444-12"}); err != nil {
		return err
	}
	if err := validateOneOf("pixfmt", pixfmt, []string{"yuv420p", "yuv422p", "yuv444p", "yuv420p10le", "yuv422p10le", "yuv444p10le", "yuv420p12le", "yuv422p12le", "yuv444p12le"}); err != nil {
		return err
	}
	if level != nil && *level != 0 {
		if !slices.Contains([]int64{30, 31, 41}, *level) {
			return fmt.Errorf("--level must be one of 30, 31, 41")
		}
	}
	if x265KeyInt != nil && *x265KeyInt != 0 {
		if *x265KeyInt < 1 || *x265KeyInt > 300 {
			return fmt.Errorf("--x265keyint must be between 1 and 300")
		}
	}

	return nil
}

func validateAv1Flags() error {
	if crf != nil && ((*crf > 0 && *crf < 16) || *crf > 63) {
		return fmt.Errorf("--crf must be between 16 and 63")
	}
	if err := validateOneOf("preset", preset, []string{"6", "7", "8", "9", "10", "11", "12", "13"}); err != nil {
		return err
	}
	if err := validateOneOf("profilev", profilev, []string{"main"}); err != nil {
		return err
	}
	if err := validateOneOf("pixfmt", pixfmt, []string{"yuv420p", "yuv420p10le"}); err != nil {
		return err
	}
	if level != nil && *level != 0 {
		if !slices.Contains([]int64{30, 31, 41}, *level) {
			return fmt.Errorf("--level must be one of 30, 31, 41")
		}
	}
	return nil
}

func validateWebmVp9Flags() error {
	if crf != nil && !(perTitle != nil && *perTitle && *crf == 0) && (*crf < 15 || *crf > 35) {
		return fmt.Errorf("--crf must be between 15 and 35")
	}
	if err := validateOneOf("quality", quality, []string{"good", "best", "realtime"}); err != nil {
		return err
	}
	if err := validateOneOf("cpu-used", cpuUsed, []string{"0", "1", "2", "3", "4", "5", "6", "7", "8"}); err != nil {
		return err
	}
	if err := validateOneOf("pixfmt", pixfmt, []string{"yuv420p", "yuv422p", "yuv440p", "yuv444p", "yuv420p10le", "yuv422p10le", "yuv440p10le", "yuv444p10le", "yuv420p12le", "yuv422p12le", "yuv440p12le", "yuv444p12le"}); err != nil {
		return err
	}

	return nil
}

// validateM4aAacFlags rejects video settings, which audio-only output ignores.
func validateM4aAacFlags() error {
	for _, flag := range []struct {
		name string
		set  bool
	}{
		{"resolution", (width != nil && *width != 0) || (height != nil && *height != 0)},
		{"framerate", framerate != nil && *framerate != 0},
		{"gop", gop != nil && *gop != 0},
		{"vb", videoBitrate != nil && *videoBitrate != 0},
		{"maxrate", maxrate != nil && *maxrate != 0},
		{"bufsize", bufsize != nil && *bufsize != 0},
		{"pixfmt", pixfmt != nil && *pixfmt != ""},
		{"crf", crf != nil && *crf != 0},
		{"preset", preset != nil && *preset != ""},
		{"profilev", profilev != nil && *profilev != ""},
		{"level", level != nil && *level != 0},
		{"an", disableAudio != nil && *disableAudio},
		{"vn", disableVideo != nil && *disableVideo},
	} {
		if flag.set {
			return fmt.Errorf("--%s is not supported for %s", flag.name, FormatM4aAac)
		}
	}
	return nil
}

var x26xPresets = []string{"ultrafast", "superfast", "veryfast", "faster", "fast", "medium"}
