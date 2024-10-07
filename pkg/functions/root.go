package functions

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
			Use:   "functions",
			Short: "Manage your functions",
			Long:  "Manage your functions",
		}}

	cmd.Command.AddCommand(newCreateCmd())
	cmd.Command.AddCommand(newEnableCmd())
	cmd.Command.AddCommand(newDisableCmd())
	cmd.Command.AddCommand(newDeleteCmd())
	cmd.Command.AddCommand(newGetCmd())
	cmd.Command.AddCommand(newListCmd())
	cmd.Command.AddCommand(newInvokeCmd())

	return cmd
}

func printError(err error) {
	fmt.Println(styles.Error.Render(err.Error()))
}
