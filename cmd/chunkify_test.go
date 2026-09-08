package cmd

import "testing"

func TestPerTitleHlsCommand(t *testing.T) {
	if err := rootCmd.ParseFlags([]string{
		"-i", "video.mp4", "-f", "hls_h264", "--per-title", "--ab", "128k",
	}); err != nil {
		t.Fatal(err)
	}
	if err := rootCmd.PreRunE(rootCmd, nil); err != nil {
		t.Fatalf("per-title HLS command should accept an audio bitrate without a video bitrate: %v", err)
	}
	enabled, err := rootCmd.Flags().GetBool("per-title")
	if err != nil || !enabled {
		t.Fatalf("expected --per-title to enable optimization, got %v, %v", enabled, err)
	}
}
