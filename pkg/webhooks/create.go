package webhooks

import (
	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/spf13/cobra"
)

type CreateCmd struct {
	Params chunkify.WebhookCreateParams
	Data   chunkify.WebhookWithSecretKey `json:"-"`
}

func (r *CreateCmd) Execute() error {
	webhook, err := cmd.Config.Client.WebhookCreate(r.Params)
	if err != nil {
		return err
	}

	r.Data = webhook

	return nil
}

func (r *CreateCmd) View() {
	webhooksList := ListCmd{Data: []chunkify.Webhook{r.Data.Webhook}}
	webhooksList.View()
}

func newCreateCmd() *cobra.Command {
	req := CreateCmd{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new webhook for your current project",
		Long:  `Create a new webhook for your current project`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := req.Execute(); err != nil {
				printError(err)
				return
			}
			req.View()
		},
	}

	cmd.Flags().StringVar(&req.Params.Url, "url", "", "The webhook URL (required)")
	cmd.Flags().BoolVar(&req.Params.Enabled, "enabled", true, "Enable the webhook")
	cmd.Flags().StringVar(&req.Params.Events, "events", "*", "Create a webhook that will trigger for specific events. *, job.* or job.completed")
	cmd.MarkFlagRequired("url")

	return cmd
}
