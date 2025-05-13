// Package files provides functionality for managing and interacting with file media files
package files

import (
	"fmt"

	"github.com/chunkifydev/cli/pkg/config"
	"github.com/chunkifydev/cli/pkg/styles"
	"github.com/spf13/cobra"
)

// Command represents the sources command structure
type Command struct {
	Command *cobra.Command // The cobra command instance
	Config  *config.Config // Configuration for the command
}

// Global command instance
var cmd *Command

// NewCommand creates and configures a new uploads command
func NewCommand(config *config.Config) *Command {
	cmd = &Command{
		Config: config,
		Command: &cobra.Command{
			Use:   "files",
			Short: "Manage files",
			Long:  "Manage files",
		}}

	// Add subcommands for file operations
	cmd.Command.AddCommand(newListCmd())   // List all files
	cmd.Command.AddCommand(newGetCmd())    // Get a file by ID
	cmd.Command.AddCommand(newDeleteCmd()) // Delete a file by ID
	return cmd
}

// printError formats and displays an error message
func printError(err error) {
	fmt.Println(styles.Error.Render(err.Error()))
}
