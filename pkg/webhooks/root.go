package webhooks

import (
	"fmt"

	"github.com/chunkifydev/cli/pkg/config"
	"github.com/chunkifydev/cli/pkg/styles"
	"github.com/spf13/cobra"
)

// Command represents the webhooks command and its configuration
type Command struct {
	Command *cobra.Command // The cobra command instance
	Config  *config.Config // Configuration for the command
}

// cmd is the global instance of the webhooks command
var cmd *Command

// NewCommand creates and returns a new webhooks command with the given configuration
func NewCommand(config *config.Config) *Command {
	cmd = &Command{
		Config: config,
		Command: &cobra.Command{
			Use:   "webhooks",
			Short: "Manage your webhooks",
			Long:  "Manage your webhooks",
		}}

	// Add all subcommands for webhook management
	cmd.Command.AddCommand(newCreateCmd())  // Create a new webhook
	cmd.Command.AddCommand(newEnableCmd())  // Enable a webhook
	cmd.Command.AddCommand(newDisableCmd()) // Disable a webhook
	cmd.Command.AddCommand(newDeleteCmd())  // Delete a webhook
	cmd.Command.AddCommand(newGetCmd())     // Get a webhook
	cmd.Command.AddCommand(newListCmd())    // List all webhooks

	return cmd
}

// printError formats and prints an error message using the defined error style
func printError(err error) {
	fmt.Println(styles.Error.Render(err.Error()))
}
