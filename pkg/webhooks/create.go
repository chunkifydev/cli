package webhooks

import (
	"github.com/level63/cli/pkg/api"
	"github.com/spf13/cobra"
)

type CreateCmd struct {
	Url     string      `json:"url"`
	Enabled bool        `json:"enabled"`
	Events  string      `json:"events,omitempty"`
	Data    api.Webhook `json:"-"`
}

func (r *CreateCmd) Execute() error {
	webhook, err := api.ApiRequest[api.Webhook](api.Request{Config: cmd.Config, Path: "/api/webhooks", Method: "POST", Body: r})
	if err != nil {
		return err
	}

	r.Data = webhook

	return nil
}

func (r *CreateCmd) View() {
	webhooksList := ListCmd{Data: []api.Webhook{r.Data}}
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

	cmd.Flags().StringVar(&req.Url, "url", "", "The webhook URL (required)")
	cmd.Flags().BoolVar(&req.Enabled, "enabled", true, "Enable the webhook")
	cmd.Flags().StringVar(&req.Events, "events", "*", "Create a webhook that will trigger for specific events. *, job.* or job.completed")
	cmd.MarkFlagRequired("url")

	return cmd
}
