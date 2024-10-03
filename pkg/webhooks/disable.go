package webhooks

import (
	"fmt"

	"github.com/level63/cli/pkg/api"
	"github.com/spf13/cobra"
)

type DisableCmd struct {
	Id      string `json:"-"`
	Enabled bool   `json:"enabled"`
}

func (r *DisableCmd) Execute() error {
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

func (r *DisableCmd) View() {
	fmt.Println("Webhook is disabled")
}

func newDisableCmd() *cobra.Command {
	req := DisableCmd{Enabled: false}

	cmd := &cobra.Command{
		Use:   "disable webhook-id",
		Short: "Disable a webhok",
		Long:  `Disable a webhok`,
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
