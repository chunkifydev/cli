package webhooks

import (
	"fmt"

	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/spf13/cobra"
)

type EnableCmd struct {
	Params chunkify.WebhookUpdateParams
}

func (r *EnableCmd) Execute() error {
	err := cmd.Config.Client.WebhookUpdate(r.Params)
	if err != nil {
		return err
	}

	return nil
}

func (r *EnableCmd) View() {
	fmt.Println("Webhook is enabled")
}

func newEnableCmd() *cobra.Command {
	enabled := true
	req := EnableCmd{Params: chunkify.WebhookUpdateParams{Enabled: &enabled}}

	cmd := &cobra.Command{
		Use:   "enable webhook-id",
		Short: "Enable a webhook",
		Long:  `Enable a webhook`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			req.Params.WebhookId = args[0]
			if err := req.Execute(); err != nil {
				printError(err)
				return
			}
			req.View()
		},
	}

	return cmd
}
