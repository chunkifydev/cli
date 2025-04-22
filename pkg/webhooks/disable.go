package webhooks

import (
	"fmt"

	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/spf13/cobra"
)

// DisableCmd represents the command for disabling a webhook
type DisableCmd struct {
	Params chunkify.WebhookUpdateParams // Params contains the parameters for updating a webhook
}

// Execute disables the webhook by setting enabled=false
func (r *DisableCmd) Execute() error {
	err := cmd.Config.Client.WebhookUpdate(r.Params)
	if err != nil {
		return err
	}

	return nil
}

// View displays a confirmation message after successful disabling
func (r *DisableCmd) View() {
	fmt.Println("Webhook is disabled")
}

// newDisableCmd creates and returns a new cobra command for webhook disabling
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
