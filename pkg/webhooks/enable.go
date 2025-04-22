package webhooks

import (
	"fmt"

	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/spf13/cobra"
)

// EnableCmd represents the command for enabling a webhook
type EnableCmd struct {
	Params chunkify.WebhookUpdateParams // Params contains the parameters for updating a webhook
}

// Execute enables the webhook by setting enabled=true
func (r *EnableCmd) Execute() error {
	err := cmd.Config.Client.WebhookUpdate(r.Params)
	if err != nil {
		return err
	}

	return nil
}

// View displays a confirmation message after successful enabling
func (r *EnableCmd) View() {
	fmt.Println("Webhook is enabled")
}

// newEnableCmd creates and returns a new cobra command for webhook enabling
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
