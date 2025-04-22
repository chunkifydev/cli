// Package projects provides functionality for managing and interacting with projects
package projects

import (
	"fmt"

	"github.com/chunkifydev/cli/pkg/config"
	"github.com/chunkifydev/cli/pkg/styles"
	"github.com/spf13/cobra"
)

// Command represents the root command for project management
type Command struct {
	Command *cobra.Command // The cobra command instance
	Config  *config.Config // Configuration for the command
}

// cmd is a package-level variable to store the command instance
var cmd *Command

// NewCommand creates and configures a new root command for project management
func NewCommand(config *config.Config) *Command {
	cmd = &Command{
		Config: config,
		Command: &cobra.Command{
			Use:   "projects",
			Short: "Manage your Chunkify projects",
			Long:  "Manage your Chunkify projects",
		}}

	// Add all subcommands
	cmd.Command.AddCommand(newCreateCmd())       // Create a new project
	cmd.Command.AddCommand(newUpdateCmd())       // Update a project
	cmd.Command.AddCommand(newDeleteCmd())       // Delete a project
	cmd.Command.AddCommand(newGetCmd())          // Get a project
	cmd.Command.AddCommand(newListCmd())         // List all projects
	cmd.Command.AddCommand(newSelectCmd(config)) // Select a project
	return cmd
}

// printError formats and prints an error message using the defined style
func printError(err error) {
	fmt.Println(styles.Error.Render(err.Error()))
}
