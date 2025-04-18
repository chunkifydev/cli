package projects

import (
	"fmt"

	"github.com/chunkifydev/cli/pkg/config"
	"github.com/chunkifydev/cli/pkg/styles"
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
			Use:   "projects",
			Short: "Manage your Chunkify projects",
			Long:  "Manage your Chunkify projects",
		}}

	cmd.Command.AddCommand(newCreateCmd())
	cmd.Command.AddCommand(newUpdateCmd())
	cmd.Command.AddCommand(newDeleteCmd())
	cmd.Command.AddCommand(newGetCmd())
	cmd.Command.AddCommand(newListCmd())
	cmd.Command.AddCommand(newSelectCmd(config))
	return cmd
}

func printError(err error) {
	fmt.Println(styles.Error.Render(err.Error()))
}
