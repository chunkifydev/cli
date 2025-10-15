package cmd

import (
	"fmt"
	"os"
	"runtime"
	"syscall"

	"github.com/chunkifydev/cli/pkg/version"
	"github.com/spf13/cobra"
)

var CliUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Chunkify CLI to the latest version",
	Long:  `Update Chunkify CLI to the latest version`,
	Run: func(cmd *cobra.Command, args []string) {
		upToDate, latestVersion := version.IsUpToDate()
		if upToDate {
			fmt.Printf("Chunkify CLI %s is up to date\n\n", latestVersion)
			return
		}
		fmt.Printf("Updating Chunkify CLI to version %s...\n\n", latestVersion)
		UpdateCli()
	},
}

func UpdateCli() {
	if runtime.GOOS == "windows" {
		fmt.Printf("Windows is not supported for automatic updates\n")
		return
	}

	// Replace current process with bash executing the curl command
	// This will show all progress and output, then exit
	if err := syscall.Exec("/bin/bash", []string{"bash", "-c", "curl -fsSL https://cli.chunkify.sh | bash"}, os.Environ()); err != nil {
		fmt.Printf("Error updating Chunkify CLI: %s\n", err)
		os.Exit(1)
	}
}
