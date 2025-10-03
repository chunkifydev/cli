package chunkify

import (
	"fmt"
	"slices"
	"strings"
)

var validPixFmts = []string{"yuv410p", "yuv411p", "yuv420p", "yuv422p", "yuv440p", "yuv444p", "yuvJ411p", "yuvJ420p", "yuvJ422p", "yuvJ440p", "yuvJ444p", "yuv420p10le", "yuv422p10le", "yuv440p10le", "yuv444p10le", "yuv420p12le", "yuv422p12le", "yuv440p12le", "yuv444p12le", "yuv420p10be", "yuv422p10be", "yuv440p10be", "yuv444p10be", "yuv420p12be", "yuv422p12be", "yuv440p12be", "yuv444p12be"}

func validateCommonVideoFlags() error {
	if width != nil && *width != 0 {
		if *width < 0 || *width > 8192 {
			return fmt.Errorf("--width must be between 0 and 8192")
		}
	}
	if height != nil && *height != 0 {
		if *height < 0 || *height > 8192 {
			return fmt.Errorf("--height must be between 0 and 8192")
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
		if *framerate < 15 || *framerate > 120 {
			return fmt.Errorf("--framerate must be between 15 and 120")
		}
	}
	if gop != nil && *gop != 0 {
		if *gop < 1 || *gop > 300 {
			return fmt.Errorf("--gop must be between 1 and 30")
		}
	}
	if channels != nil && *channels != 0 {
		if slices.Contains([]int64{1, 2, 5, 7}, *channels) {
			return fmt.Errorf("--channels must be one of 1, 2, 5, 7")
		}
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
	if pixfmt != nil && *pixfmt != "" {
		if !slices.Contains(validPixFmts, *pixfmt) {
			return fmt.Errorf("--pixfmt must be one of %s", strings.Join(validPixFmts, ", "))
		}
	}

	return nil
}

func validateH264Flags() error {
	if crf != nil && ((*crf > 0 && *crf < 16) || *crf > 35) {
		return fmt.Errorf("--crf must be between 16 and 35")
	}
	if preset != nil && *preset != "" {
		if !slices.Contains([]string{"ultrafast", "superfast", "veryfast", "faster", "fast", "medium"}, *preset) {
			return fmt.Errorf("--preset must be one of ultrafast, superfast, veryfast, faster, fast, medium, slow, slower, veryslow")
		}
	}
	if profilev != nil && *profilev != "" {
		if !slices.Contains([]string{"baseline", "main", "high", "high10", "high422", "high444"}, *profilev) {
			return fmt.Errorf("--profilev must be one of baseline, main, high, high10, high422, high444")
		}
	}
	if level != nil && *level != 0 {
		if slices.Contains([]int64{10, 11, 12, 13, 20, 21, 22, 30, 31, 32, 40, 41, 42, 50, 51}, *level) {
			return fmt.Errorf("--level must be one of 10, 11, 12, 13, 20, 21, 22, 30, 31, 32, 40, 41, 42, 50, 51")
		}
	}
	if x264KeyInt != nil && *x264KeyInt != 0 {
		if *x264KeyInt < 1 || *x264KeyInt > 300 {
			return fmt.Errorf("--x264keyint must be between 1 and 30")
		}
	}

	return nil
}

func validateH265Flags() error {
	if crf != nil && ((*crf > 0 && *crf < 16) || *crf > 35) {
		return fmt.Errorf("--crf must be between 16 and 35")
	}
	if preset != nil && *preset != "" {
		if !slices.Contains([]string{"ultrafast", "superfast", "veryfast", "faster", "fast", "medium"}, *preset) {
			return fmt.Errorf("--preset must be one of ultrafast, superfast, veryfast, faster, fast, medium, slow, slower, veryslow")
		}
	}
	if profilev != nil && *profilev != "" {
		if !slices.Contains([]string{"main", "main10", "mainstillpicture"}, *profilev) {
			return fmt.Errorf("--profilev must be one of baseline, main, high, high10, high422, high444")
		}
	}
	if level != nil && *level != 0 {
		if slices.Contains([]int64{30, 31, 41}, *level) {
			return fmt.Errorf("--level must be one of 30, 31, 41")
		}
	}
	if x265KeyInt != nil && *x265KeyInt != 0 {
		if *x265KeyInt < 1 || *x265KeyInt > 300 {
			return fmt.Errorf("--x265keyint must be between 1 and 30")
		}
	}

	return nil
}

func validateAv1Flags() error {
	if crf != nil && ((*crf > 0 && *crf < 16) || *crf > 63) {
		return fmt.Errorf("--crf must be between 16 and 63")
	}
	if preset != nil && *preset != "" {
		if !slices.Contains([]string{"6", "7", "8", "9", "10", "11", "12", "13"}, *preset) {
			return fmt.Errorf("--preset must be one of 6, 7, 8, 9, 10, 11, 12, 13")
		}
	}
	if profilev != nil && *profilev != "" {
		if !slices.Contains([]string{"main", "main10", "mainstillpicture"}, *profilev) {
			return fmt.Errorf("--profilev must be one of main, main10, mainstillpicture")
		}
	}
	if level != nil && *level != 0 {
		if slices.Contains([]int64{30, 31, 41}, *level) {
			return fmt.Errorf("--level must be one of 30, 31, 41")
		}
	}
	return nil
}

func validateWebmVp9Flags() error {
	if crf != nil && (*crf < 15 || *crf > 35) {
		return fmt.Errorf("--crf must be between 15 and 35")
	}
	if quality != nil && *quality != "" {
		if !slices.Contains([]string{"good", "best", "realtime"}, *quality) {
			return fmt.Errorf("--quality must be one of good, best, realtime")
		}
	}
	if cpuUsed != nil && *cpuUsed != "" {
		if !slices.Contains([]string{"0", "1", "2", "3", "4", "5", "6", "7", "8"}, *cpuUsed) {
			return fmt.Errorf("--cpu-used must be one of 0, 1, 2, 3, 4, 5, 6, 7, 8")
		}
	}

	return nil
}

func validateJpgFlags() error {
	if interval == nil && *interval != 0 {
		if *interval < 0 || *interval > 60 {
			return fmt.Errorf("--interval must be between 0 and 60")
		}
	}
	return nil
}

func validateHlsFlags() error {
	if videoBitrate == nil || *videoBitrate == 0 {
		return fmt.Errorf("--vb (video bitrate) flag is required when format is hls")
	}
	if audioBitrate == nil || *audioBitrate == 0 {
		return fmt.Errorf("--ab (audio bitrate) flag is required when format is hls")
	}

	if hlsTime != nil && *hlsTime != 0 {
		if *hlsTime < 1 || *hlsTime > 10 {
			return fmt.Errorf("--hls-time must be between 1 and 10")
		}
	}
	if hlsSegmentType != nil && *hlsSegmentType != "" {
		if !slices.Contains([]string{"mpegts", "fmp4"}, *hlsSegmentType) {
			return fmt.Errorf("--hls-segment-type must be one of mpegts, fmp4")
		}
	}
	return nil
}

func validateHlsH264Flags() error {
	if err := validateHlsFlags(); err != nil {
		return err
	}

	return validateH264Flags()
}

func validateHlsH265Flags() error {
	if err := validateHlsFlags(); err != nil {
		return err
	}

	return validateH265Flags()
}

func validateHlsAv1Flags() error {
	if err := validateHlsFlags(); err != nil {
		return err
	}

	return validateAv1Flags()
}
