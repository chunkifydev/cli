package notifications

import (
	"github.com/chunkifydev/cli/pkg/api"
	"github.com/spf13/cobra"
)

type CreateCmd struct {
	JobId     string `json:"job_id"`
	WebhookId string `json:"webhook_id,omitempty"`
	Event     string `json:"event"`

	Data api.Notification `json:"-"`
}

func (r *CreateCmd) Execute() error {
	notification, err := api.ApiRequest[api.Notification](api.Request{Config: cmd.Config, Path: "/api/notifications", Method: "POST", Body: r})
	if err != nil {
		return err
	}

	r.Data = notification

	return nil
}

func (r *CreateCmd) View() {
	notifList := &ListCmd{CreatedSort: "asc", JobId: r.JobId}
	notifList.Execute()

	notifList.View()

}

func newCreateCmd() *cobra.Command {
	req := CreateCmd{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new notification",
		Long:  `Create a new notification`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := req.Execute(); err != nil {
				printError(err)
				return
			}
			req.View()
		},
	}

	cmd.Flags().StringVar(&req.JobId, "job-id", "", "The job id (required)")
	cmd.Flags().StringVar(&req.WebhookId, "webhook-id", "", "The webhook id (required)")
	cmd.Flags().StringVar(&req.Event, "event", "", "The event associated with the notification. Possible values: job.completed (required)")

	cmd.MarkFlagRequired("job-id")
	cmd.MarkFlagRequired("webhook-id")
	cmd.MarkFlagRequired("event")

	return cmd
}
