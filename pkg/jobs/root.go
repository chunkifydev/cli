package jobs

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
			Use:   "jobs",
			Short: "Manage your jobs",
			Long:  "Manage your jobs",
		}}

	// Add all subcommands
	cmd.Command.AddCommand(newCreateCmd())             // Create new jobs
	cmd.Command.AddCommand(newGetCmd())                // Get details of a specific job
	cmd.Command.AddCommand(newListCmd())               // List all jobs
	cmd.Command.AddCommand(newFilesListCmd())          // List files associated with jobs
	cmd.Command.AddCommand(newTranscoderProgressCmd()) // Monitor transcoding progress

	return cmd
}

// printError formats and prints an error message
func printError(err error) {
	fmt.Println(styles.Error.Render(err.Error()))
}
