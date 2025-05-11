// Package notifications provides functionality for managing and interacting with notifications
package notifications

import (
	"fmt"

	"github.com/chunkifydev/cli/pkg/config"
	"github.com/chunkifydev/cli/pkg/styles"
	"github.com/spf13/cobra"
)

// Command represents the root notifications command and configuration
type Command struct {
	Command *cobra.Command // The root cobra command for notifications
	Config  *config.Config // Configuration for the notifications command
}

// cmd is a package-level variable holding the current Command instance
var cmd *Command

// NewCommand creates and configures a new notifications root command
func NewCommand(config *config.Config) *Command {
	cmd = &Command{
		Config: config,
		Command: &cobra.Command{
			Use:   "notifications",
			Short: "Manage your notifications",
			Long:  "Manage your notifications",
		}}

	// Add all subcommands
	cmd.Command.AddCommand(newListCmd())   // List of notifications
	cmd.Command.AddCommand(newGetCmd())    // Get a single notification
	cmd.Command.AddCommand(newCreateCmd()) // Create a new notification
	cmd.Command.AddCommand(newProxyCmd())  // Proxy notifications to a local URL
	cmd.Command.AddCommand(newDeleteCmd()) // Delete a notification
	return cmd
}

// printError formats and prints an error message using the error style
func printError(err error) {
	fmt.Println(styles.Error.Render(err.Error()))
}
