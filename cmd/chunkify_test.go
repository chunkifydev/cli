package cmd

import (
	"bytes"
	"strings"
	"testing"

	chunkifyCmd "github.com/chunkifydev/cli/pkg/chunkify"
	"github.com/chunkifydev/cli/pkg/config"
)

func TestOutputStorageFlags(t *testing.T) {
	for _, flag := range []string{"--output-storage-path", "--storage-path"} {
		t.Run(flag, func(t *testing.T) {
			command := chunkifyCmd.NewCommand(&config.Config{})
			var output bytes.Buffer
			command.Command.SetOut(&output)
			command.Command.SetErr(&output)
			if err := command.Command.ParseFlags([]string{"-i", "video.mp4", "-o", "result.mp4", "--output-storage-id", "stor_aws_test", flag, "custom/result.mp4"}); err != nil {
				t.Fatal(err)
			}
			if err := command.Command.PreRunE(command.Command, nil); err != nil {
				t.Fatal(err)
			}
			storage := command.App.Command.JobCreateStorageParams
			if storage.ID.Value != "stor_aws_test" || storage.Path.Value != "custom/result.mp4" {
				t.Fatalf("incorrect storage parameters: %+v", storage)
			}
			if flag == "--storage-path" && !strings.Contains(output.String(), "deprecated") {
				t.Fatal("legacy flag must emit a deprecation warning")
			}
		})
	}
}

func TestOutputStoragePathAliasesConflict(t *testing.T) {
	command := chunkifyCmd.NewCommand(&config.Config{})
	command.Command.SetOut(new(bytes.Buffer))
	command.Command.SetErr(new(bytes.Buffer))
	command.Command.SetArgs([]string{"-i", "video.mp4", "-o", "result.mp4", "--storage-path", "one.mp4", "--output-storage-path", "two.mp4"})
	if err := command.Command.Execute(); err == nil || !strings.Contains(err.Error(), "storage-path") {
		t.Fatalf("expected conflicting aliases error, got %v", err)
	}
}

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
