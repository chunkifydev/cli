package formatter

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// Duration formats a duration in seconds into a human readable string (HH:MM:SS)
func Duration(duration int64) string {
	if duration < 60 {
		return fmt.Sprintf("00:%02d", duration)
	}

	if duration < 3600 {
		return fmt.Sprintf("%02d:%02d", duration/60, duration%60)
	}

	return fmt.Sprintf("%02d:%02d:%02d", duration/3600, (duration%3600)/60, duration%60)
}

// Size formats a size in bytes into a human readable string (B, KB, MB, GB)
func Size(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%dB", size)
	}

	if size < 1024*1024 {
		return fmt.Sprintf("%dKB", size/1024)
	}

	if size < 1024*1024*1024 {
		return fmt.Sprintf("%dMB", size/1024/1024)
	}
	return fmt.Sprintf("%.2fGB", float64(size)/1024/1024/1024)
}

// Bitrate formats a bitrate in bytes into a human readable string (KB/s, MB/s)
func Bitrate(size int64) string {
	if size < 1024 {
		return "N/A"
	}

	if size < 1024*1024 {
		return fmt.Sprintf("%dKB/s", size/1024)
	}

	return fmt.Sprintf("%.fMB/s", float64(size)/1024/1024)
}

// TimeDiff calculates and formats the duration between two timestamps
func TimeDiff(start, end time.Time) string {
	if start.IsZero() || end.IsZero() {
		return ""
	}

	duration := end.Sub(start).Seconds()
	if duration < 0 {
		return ""
	}

	return Duration(int64(duration))
}

// ParseFileSize parses a human readable file size string (e.g., "1.5GB") into bytes
// Supports B, KB/K/KiB, MB/M/MiB, GB/G/GiB, TB/T/TiB units
func ParseFileSize(sizeStr string) (int64, error) {
	const (
		_   = iota
		KiB = 1 << (10 * iota)
		MiB
		GiB
		TiB
	)

	multipliers := map[string]int64{
		"B":  1,
		"KB": KiB, "K": KiB, "KiB": KiB,
		"MB": MiB, "M": MiB, "MiB": MiB,
		"GB": GiB, "G": GiB, "GiB": GiB,
		"TB": TiB, "T": TiB, "TiB": TiB,
	}

	var (
		number   string
		suffix   string
		hasDigit bool
	)

	for _, r := range sizeStr {
		if unicode.IsDigit(r) || r == '.' {
			number += string(r)
			hasDigit = true
		} else {
			suffix += string(r)
		}
	}

	if !hasDigit {
		return 0, fmt.Errorf("invalid size: %s", sizeStr)
	}

	value, err := strconv.ParseFloat(number, 64)
	if err != nil {
		return 0, err
	}

	suffix = strings.TrimSpace(strings.ToUpper(suffix))
	multiplier, ok := multipliers[suffix]
	if !ok {
		return 0, fmt.Errorf("unknown size suffix: %s", suffix)
	}

	return int64(value * float64(multiplier)), nil
}
