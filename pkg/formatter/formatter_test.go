package formatter

import (
	"testing"
	"time"
)

func TestDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration int64
		expected string
	}{
		{
			name:     "seconds_only",
			duration: 30,
			expected: "00:30",
		},
		{
			name:     "minutes_and_seconds",
			duration: 90,
			expected: "01:30",
		},
		{
			name:     "hours_minutes_seconds",
			duration: 3661,
			expected: "01:01:01",
		},
		{
			name:     "zero_seconds",
			duration: 0,
			expected: "00:00",
		},
		{
			name:     "exactly_one_minute",
			duration: 60,
			expected: "01:00",
		},
		{
			name:     "exactly_one_hour",
			duration: 3600,
			expected: "01:00:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Duration(tt.duration)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestSize(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		expected string
	}{
		{
			name:     "bytes",
			size:     512,
			expected: "512B",
		},
		{
			name:     "kilobytes",
			size:     2048,
			expected: "2KB",
		},
		{
			name:     "megabytes",
			size:     2097152,
			expected: "2MB",
		},
		{
			name:     "gigabytes",
			size:     2147483648,
			expected: "2.00GB",
		},
		{
			name:     "zero_bytes",
			size:     0,
			expected: "0B",
		},
		{
			name:     "exactly_one_kb",
			size:     1024,
			expected: "1KB",
		},
		{
			name:     "exactly_one_mb",
			size:     1048576,
			expected: "1MB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Size(tt.size)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestBitrate(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		expected string
	}{
		{
			name:     "too_small",
			size:     512,
			expected: "N/A",
		},
		{
			name:     "kilobytes_per_second",
			size:     2048,
			expected: "2KB/s",
		},
		{
			name:     "megabytes_per_second",
			size:     2097152,
			expected: "2MB/s",
		},
		{
			name:     "exactly_one_kb_per_second",
			size:     1024,
			expected: "1KB/s",
		},
		{
			name:     "exactly_one_mb_per_second",
			size:     1048576,
			expected: "1MB/s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Bitrate(tt.size)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestTimeDiff(t *testing.T) {
	tests := []struct {
		name     string
		start    time.Time
		end      time.Time
		expected string
	}{
		{
			name:     "valid_duration",
			start:    time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			end:      time.Date(2023, 1, 1, 12, 1, 30, 0, time.UTC),
			expected: "01:30",
		},
		{
			name:     "zero_start_time",
			start:    time.Time{},
			end:      time.Date(2023, 1, 1, 12, 1, 30, 0, time.UTC),
			expected: "",
		},
		{
			name:     "zero_end_time",
			start:    time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			end:      time.Time{},
			expected: "",
		},
		{
			name:     "negative_duration",
			start:    time.Date(2023, 1, 1, 12, 1, 0, 0, time.UTC),
			end:      time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			expected: "",
		},
		{
			name:     "same_time",
			start:    time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			end:      time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			expected: "00:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TimeDiff(tt.start, tt.end)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestParseFileSize(t *testing.T) {
	tests := []struct {
		name        string
		sizeStr     string
		expected    int64
		expectError bool
	}{
		{
			name:        "bytes",
			sizeStr:     "512B",
			expected:    512,
			expectError: false,
		},
		{
			name:        "kilobytes",
			sizeStr:     "2KB",
			expected:    2048,
			expectError: false,
		},
		{
			name:        "kilobytes_short",
			sizeStr:     "2K",
			expected:    2048,
			expectError: false,
		},
		{
			name:        "kilobytes_binary",
			sizeStr:     "2KiB",
			expected:    0,
			expectError: true,
		},
		{
			name:        "megabytes",
			sizeStr:     "2MB",
			expected:    2097152,
			expectError: false,
		},
		{
			name:        "megabytes_short",
			sizeStr:     "2M",
			expected:    2097152,
			expectError: false,
		},
		{
			name:        "gigabytes",
			sizeStr:     "1.5GB",
			expected:    1610612736,
			expectError: false,
		},
		{
			name:        "gigabytes_short",
			sizeStr:     "1.5G",
			expected:    1610612736,
			expectError: false,
		},
		{
			name:        "terabytes",
			sizeStr:     "2TB",
			expected:    2199023255552,
			expectError: false,
		},
		{
			name:        "decimal_megabytes",
			sizeStr:     "1.5MB",
			expected:    1572864,
			expectError: false,
		},
		{
			name:        "no_unit",
			sizeStr:     "1024",
			expected:    0,
			expectError: true,
		},
		{
			name:        "invalid_number",
			sizeStr:     "abcMB",
			expected:    0,
			expectError: true,
		},
		{
			name:        "invalid_unit",
			sizeStr:     "1024XB",
			expected:    0,
			expectError: true,
		},
		{
			name:        "empty_string",
			sizeStr:     "",
			expected:    0,
			expectError: true,
		},
		{
			name:        "no_digits",
			sizeStr:     "MB",
			expected:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseFileSize(tt.sizeStr)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %d, got %d", tt.expected, result)
				}
			}
		})
	}
}
