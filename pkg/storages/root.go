package storages

import (
	"fmt"

	"github.com/chunkifydev/cli/pkg/config"
	"github.com/chunkifydev/cli/pkg/styles"
	"github.com/spf13/cobra"
)

// Command represents the root command for storage management
type Command struct {
	Command *cobra.Command // The cobra command instance
	Config  *config.Config // Configuration for the command
}

// Global command instance
var cmd *Command

// NewCommand creates and configures a new root command for storage management
func NewCommand(config *config.Config) *Command {
	cmd = &Command{
		Config: config,
		Command: &cobra.Command{
			Use:   "storages",
			Short: "Manage your storages",
			Long:  "Manage your storages",
		}}

	// Add subcommands for storage operations
	cmd.Command.AddCommand(newCreateCmd()) // Create a new storage
	cmd.Command.AddCommand(newDeleteCmd()) // Delete a storage
	cmd.Command.AddCommand(newGetCmd())    // Get a storage
	cmd.Command.AddCommand(newListCmd())   // List all storages
	return cmd
}

// printError formats and prints an error message
func printError(err error) {
	fmt.Println(styles.Error.Render(err.Error()))
}
