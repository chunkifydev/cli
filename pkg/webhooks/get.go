package webhooks

import (
	"fmt"

	"github.com/chunkifydev/cli/pkg/api"
	"github.com/spf13/cobra"
)

type GetCmd struct {
	Id   string `json:"id"`
	Data api.Webhook
}

func (r *GetCmd) Execute() error {
	apiReq := api.Request{
		Config: cmd.Config,
		Path:   fmt.Sprintf("/api/webhooks/%s", r.Id),
		Method: "GET",
	}

	source, err := api.ApiRequest[api.Webhook](apiReq)
	if err != nil {
		return err
	}

	r.Data = source

	return nil
}

func (r *GetCmd) View() {
	sourceList := ListCmd{Data: []api.Webhook{r.Data}}
	sourceList.View()
}

func newGetCmd() *cobra.Command {
	req := GetCmd{}

	cmd := &cobra.Command{
		Use:   "get webhook-id",
		Short: "get info about a webhook",
		Long:  "get info about a webhook",
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
