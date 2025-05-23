package transcoders

import (
	"fmt"

	"github.com/chunkifydev/cli/pkg/config"
	"github.com/chunkifydev/cli/pkg/styles"
	"github.com/spf13/cobra"
)

// Command wraps a cobra.Command with configuration
type Command struct {
	Command *cobra.Command // The underlying cobra command
	Config  *config.Config // Configuration for the command
}

// Global command instance
var cmd *Command

// NewCommand creates and configures a new jobs command with subcommands
func NewCommand(config *config.Config) *Command {
	cmd = &Command{
		Config: config,
		Command: &cobra.Command{
			Use:   "transcoders",
			Short: "Check the status of the transcoders",
			Long:  "Check the status of the transcoders",
		}}

	// Add all subcommands
	cmd.Command.AddCommand(newListCmd()) // Get transcoder statuses for a specific job

	return cmd
}

// printError formats and prints an error message
func printError(err error) {
	fmt.Println(styles.Error.Render(err.Error()))
}
