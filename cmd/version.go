package cmd

import (
	"fmt"

	"github.com/chunkifydev/cli/pkg/version"
	"github.com/spf13/cobra"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of chunkify",
	Long:  `Print the version number of chunkify`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("chunkify version %s\n", version.Version)
	},
}
