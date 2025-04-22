package tokens

import (
	"fmt"

	"github.com/chunkifydev/cli/pkg/config"
	"github.com/chunkifydev/cli/pkg/styles"
	"github.com/spf13/cobra"
)

// Command represents the root command for token management
type Command struct {
	Command *cobra.Command // The cobra command instance
	Config  *config.Config // Configuration for the command
}

// Global command instance
var cmd *Command

// NewCommand creates and configures a new root command for token management
func NewCommand(config *config.Config) *Command {
	cmd = &Command{
		Config: config,
		Command: &cobra.Command{
			Use:   "tokens",
			Short: "Manage your Chunkify access tokens",
			Long:  "Manage your Chunkify access tokens",
		}}

	// Add subcommands for token operations
	cmd.Command.AddCommand(newCreateCmd()) // Create a new token
	cmd.Command.AddCommand(newDeleteCmd()) // Delete/revoke a token
	cmd.Command.AddCommand(newListCmd())   // List all tokens
	return cmd
}

// printError formats and prints an error message
func printError(err error) {
	fmt.Println(styles.Error.Render(err.Error()))
}
