package sources

import (
	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/spf13/cobra"
)

type GetCmd struct {
	Id   string `json:"id"`
	Data chunkify.Source
}

func (r *GetCmd) Execute() error {
	source, err := cmd.Config.Client.Source(r.Id)
	if err != nil {
		return err
	}

	r.Data = source

	return nil
}

func (r *GetCmd) View() {
	sourceList := ListCmd{Data: []chunkify.Source{r.Data}}
	sourceList.View()
}

func newGetCmd() *cobra.Command {
	req := GetCmd{}

	cmd := &cobra.Command{
		Use:   "get source-id",
		Short: "get info about a source",
		Long:  "get info about a source",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			req.Id = args[0]
			if err := req.Execute(); err != nil {
				printError(err)
				return
			}
			req.View()
		},
	}

	return cmd
}
