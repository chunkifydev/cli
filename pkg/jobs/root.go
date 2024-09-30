package jobs

import (
	"fmt"

	"github.com/level63/cli/pkg/config"
	"github.com/level63/cli/pkg/styles"
	"github.com/spf13/cobra"
)

type Command struct {
	Command *cobra.Command
	Config  *config.Config
}

var cmd *Command

func NewCommand(config *config.Config) *Command {
	cmd = &Command{
		Config: config,
		Command: &cobra.Command{
			Use:   "jobs",
			Short: "Manage your jobs",
			Long:  "Manage your jobs",
		}}

	cmd.Command.AddCommand(newCreateCmd())
	cmd.Command.AddCommand(newGetCmd())
	cmd.Command.AddCommand(newListCmd())
	cmd.Command.AddCommand(newFilesListCmd())
	cmd.Command.AddCommand(newWebhooksListCmd())
	cmd.Command.AddCommand(newFunctionsListCmd())
	cmd.Command.AddCommand(newRestartCmd())

	return cmd
}

func printError(err error) {
	fmt.Println(styles.Error.Render(err.Error()))
}
