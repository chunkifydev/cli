package tokens

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
			Use:   "tokens",
			Short: "Manage your Chunkify access tokens",
			Long:  "Manage your Chunkify access tokens",
		}}

	cmd.Command.AddCommand(newCreateCmd())
	cmd.Command.AddCommand(newDeleteCmd())
	cmd.Command.AddCommand(newListCmd())
	return cmd
}

func printError(err error) {
	fmt.Println(styles.Error.Render(err.Error()))
}
