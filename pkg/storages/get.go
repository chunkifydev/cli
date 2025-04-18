package storages

import (
	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/spf13/cobra"
)

type GetCmd struct {
	Id   string
	Data chunkify.Storage
}

func (r *GetCmd) Execute() error {
	storage, err := cmd.Config.Client.Storage(r.Id)
	if err != nil {
		return err
	}

	r.Data = storage

	return nil
}

func (r *GetCmd) View() {
	sourceList := ListCmd{Data: []chunkify.Storage{r.Data}}
	sourceList.View()
}

func newGetCmd() *cobra.Command {
	req := GetCmd{}

	cmd := &cobra.Command{
		Use:   "get storage-id",
		Short: "get info about a storage",
		Long:  "get info about a storage",
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
