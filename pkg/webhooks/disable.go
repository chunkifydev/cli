package webhooks

import (
	"fmt"

	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/spf13/cobra"
)

type DisableCmd struct {
	Params chunkify.WebhookUpdateParams
}

func (r *DisableCmd) Execute() error {
	err := cmd.Config.Client.WebhookUpdate(r.Params)
	if err != nil {
		return err
	}

	return nil
}

func (r *DisableCmd) View() {
	fmt.Println("Webhook is disabled")
}

func newDisableCmd() *cobra.Command {
	enabled := false
	req := DisableCmd{Params: chunkify.WebhookUpdateParams{Enabled: &enabled}}

	cmd := &cobra.Command{
		Use:   "disable webhook-id",
		Short: "Disable a webhook",
		Long:  `Disable a webhook`,
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
