package storages

import (
	"fmt"

	"github.com/level63/cli/pkg/api"
	"github.com/spf13/cobra"
)

type GetCmd struct {
	Id   string
	Data api.Storage
}

func (r *GetCmd) Execute() error {
	apiReq := api.Request{
		Config: cmd.Config,
		Path:   fmt.Sprintf("/api/storage/%s", r.Id),
		Method: "GET",
	}

	source, err := api.ApiRequest[api.Storage](apiReq)
	if err != nil {
		return err
	}

	r.Data = source

	return nil
}

func (r *GetCmd) View() {
	sourceList := ListCmd{Data: []api.Storage{r.Data}}
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
