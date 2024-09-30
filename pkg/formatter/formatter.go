package formatter

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/level63/cli/pkg/styles"
)

func Duration(duration int64) string {
	if duration < 60 {
		return fmt.Sprintf("00:%02d", duration)
	}

	if duration < 3600 {
		return fmt.Sprintf("%02d:%02d", duration/60, duration%60)
	}

	return fmt.Sprintf("%02d:%02d:%02d", duration/3600, (duration%3600)/60, duration%60)
}

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

func Bitrate(size int64) string {
	if size < 1024 {
		return "N/A"
	}

	if size < 1024*1024 {
		return fmt.Sprintf("%dKB/s", size/1024)
	}

	return fmt.Sprintf("%.fMB/s", float64(size)/1024/1024)
}

func JobStatus(status string) string {
	switch status {
	case "finished":
		return styles.Important.Render(status)
	case "error":
		return styles.Error.Render(status)
	default:
		return styles.Working.Render(status)
	}
}

func Bool(b bool) string {
	if b {
		return styles.Important.Render("yes")
	}

	return styles.Error.Render("no")
}

func BoolDefaultColor(b bool) string {
	if b {
		return styles.DefaultText.Render("yes")
	}

	return styles.DefaultText.Render("no")
}

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
