package functions

import (
	"fmt"

	"github.com/chunkifydev/cli/pkg/api"
	"github.com/spf13/cobra"
)

type GetCmd struct {
	Id   string `json:"id"`
	Data api.Function
}

func (r *GetCmd) Execute() error {
	apiReq := api.Request{
		Config: cmd.Config,
		Path:   fmt.Sprintf("/api/functions/%s", r.Id),
		Method: "GET",
	}

	source, err := api.ApiRequest[api.Function](apiReq)
	if err != nil {
		return err
	}

	r.Data = source

	return nil
}

func (r *GetCmd) View() {
	functionsList := ListCmd{Data: []api.Function{r.Data}}
	functionsList.View()
}

func newGetCmd() *cobra.Command {
	req := GetCmd{}

	cmd := &cobra.Command{
		Use:   "get function-id",
		Short: "get info about a function",
		Long:  "get info about a function",
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
