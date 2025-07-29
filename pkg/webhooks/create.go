package webhooks

import (
	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/cli/pkg/flags"
	"github.com/spf13/cobra"
)

// CreateCmd represents the command for creating a new webhook
type CreateCmd struct {
	Params chunkify.WebhookCreateParams // Params contains the parameters for creating a webhook
	Data   chunkify.Webhook             `json:"-"` // Data contains the created webhook response
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
	webhooksList := ListCmd{Data: []chunkify.Webhook{r.Data}}
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
	flags.StringVar(cmd.Flags(), &req.Params.Url, "url", "", "The webhook URL (required)")
	flags.BoolVar(cmd.Flags(), &enabled, "enabled", true, "Enable the webhook")
	flags.StringArrayVar(cmd.Flags(), &req.Params.Events, "events", chunkify.NotificationEventsAll, "Create a webhook that will trigger for specific events. job.completed, job.failed, upload.completed, upload.failed, upload.expired")
	cmd.MarkFlagRequired("url")

	return cmd
}
