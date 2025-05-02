package webhooks

import (
	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/spf13/cobra"
)

// CreateCmd represents the command for creating a new webhook
type CreateCmd struct {
	Params chunkify.WebhookCreateParams  // Params contains the parameters for creating a webhook
	Data   chunkify.WebhookWithSecretKey `json:"-"` // Data contains the created webhook response including the secret key
}

// Execute creates a new webhook using the provided parameters
func (r *CreateCmd) Execute() error {
	webhook, err := cmd.Config.Client.WebhookCreate(r.Params)
	if err != nil {
		return err
	}

	r.Data = webhook

	return nil
}

// View displays the created webhook in a formatted table
func (r *CreateCmd) View() {
	webhooksList := ListCmd{Data: []chunkify.Webhook{r.Data.Webhook}}
	webhooksList.View()
}

// newCreateCmd creates and returns a new cobra command for webhook creation
func newCreateCmd() *cobra.Command {
	var enabled bool
	req := CreateCmd{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new webhook for your current project",
		Long:  `Create a new webhook for your current project`,
		Run: func(cmd *cobra.Command, args []string) {
			req.Params.Enabled = &enabled
			if err := req.Execute(); err != nil {
				printError(err)
				return
			}
			req.View()
		},
	}

	// Add command flags for configuring the webhook
	cmd.Flags().StringVar(&req.Params.Url, "url", "", "The webhook URL (required)")
	cmd.Flags().BoolVar(&enabled, "enabled", true, "Enable the webhook")
	cmd.Flags().StringVar(&req.Params.Events, "events", "*", "Create a webhook that will trigger for specific events. *, job.* or job.completed")
	cmd.MarkFlagRequired("url")

	return cmd
}
