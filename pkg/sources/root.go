package sources

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
			Use:   "sources",
			Short: "Create and get information about source media",
			Long:  "Create and get information about source media",
		}}

	cmd.Command.AddCommand(newCreateCmd())
	cmd.Command.AddCommand(newGetCmd())
	cmd.Command.AddCommand(newListCmd())
	return cmd
}

func printError(err error) {
	fmt.Println(styles.Error.Render(err.Error()))
}
