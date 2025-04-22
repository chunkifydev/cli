// Package sources provides functionality for managing and interacting with source media files
package sources

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

// NewCommand creates and configures a new sources command
func NewCommand(config *config.Config) *Command {
	cmd = &Command{
		Config: config,
		Command: &cobra.Command{
			Use:   "sources",
			Short: "Create and get information about source media",
			Long:  "Create and get information about source media",
		}}

	// Add subcommands for source operations
	cmd.Command.AddCommand(newCreateCmd()) // Create a new source
	cmd.Command.AddCommand(newGetCmd())    // Get information about a source
	cmd.Command.AddCommand(newListCmd())   // List all sources
	return cmd
}

// printError formats and displays an error message
func printError(err error) {
	fmt.Println(styles.Error.Render(err.Error()))
}
