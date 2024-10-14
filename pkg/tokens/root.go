package tokens

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
			Use:   "tokens",
			Short: "Manage your Level63 access tokens",
			Long:  "Manage your Level63 access tokens",
		}}

	cmd.Command.AddCommand(newCreateCmd())
	cmd.Command.AddCommand(newDeleteCmd())
	cmd.Command.AddCommand(newListCmd())
	return cmd
}

func printError(err error) {
	fmt.Println(styles.Error.Render(err.Error()))
}
