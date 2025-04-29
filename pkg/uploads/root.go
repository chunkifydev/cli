// Package uploads provides functionality for managing and interacting with upload media files
package uploads

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
			Use:   "uploads",
			Short: "Create a new upload",
			Long:  "Create a new upload",
		}}

	// Add subcommands for upload operations
	cmd.Command.AddCommand(newCreateCmd()) // Create a new upload
	cmd.Command.AddCommand(newListCmd())   // List all uploads
	cmd.Command.AddCommand(newGetCmd())    // Get an upload by ID
	return cmd
}

// printError formats and displays an error message
func printError(err error) {
	fmt.Println(styles.Error.Render(err.Error()))
}
