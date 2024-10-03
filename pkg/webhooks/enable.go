package webhooks

import (
	"fmt"

	"github.com/level63/cli/pkg/api"
	"github.com/spf13/cobra"
)

type EnableCmd struct {
	Id      string `json:"-"`
	Enabled bool   `json:"enabled"`
}

func (r *EnableCmd) Execute() error {
	_, err := api.ApiRequest[api.EmptyResponse](
		api.Request{
			Config: cmd.Config,
			Path:   fmt.Sprintf("/api/webhooks/%s", r.Id),
			Method: "PATCH",
			Body:   r,
		})
	if err != nil {
		return err
	}

	return nil
}

func (r *EnableCmd) View() {
	fmt.Println("Webhook is enabled")
}

func newEnableCmd() *cobra.Command {
	req := EnableCmd{Enabled: true}

	cmd := &cobra.Command{
		Use:   "enable webhook-id",
		Short: "Enable a webhook",
		Long:  `Enable a webhook`,
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
