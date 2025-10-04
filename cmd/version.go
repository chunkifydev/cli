package cmd

import (
	"fmt"

	"github.com/chunkifydev/cli/pkg/version"
	"github.com/spf13/cobra"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Chunkify",
	Long:  `Print the version number of Chunkify`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Chunkify version %s\n", version.Version)
	},
}
