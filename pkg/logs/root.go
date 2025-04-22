package logs

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

// NewCommand creates and configures a new logs command with subcommands
func NewCommand(config *config.Config) *Command {
	cmd = &Command{
		Config: config,
		Command: &cobra.Command{
			Use:   "logs",
			Short: "Retrieve logs",
			Long:  "Retrieve logs",
		}}

	cmd.Command.AddCommand(newListCmd()) // Add list subcommand

	return cmd
}

// printError formats and prints an error message
func printError(err error) {
	fmt.Println(styles.Error.Render(err.Error()))
}
