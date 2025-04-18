package notifications

import (
	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/spf13/cobra"
)

type CreateCmd struct {
	Params chunkify.NotificationCreateParams

	Data chunkify.Notification `json:"-"`
}

func (r *CreateCmd) Execute() error {
	notification, err := cmd.Config.Client.NotificationCreate(r.Params)
	if err != nil {
		return err
	}

	r.Data = notification

	return nil
}

func (r *CreateCmd) View() {
	notifList := &ListCmd{Params: chunkify.NotificationListParams{CreatedSort: "asc", JobId: r.Params.JobId}}
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

	cmd.Flags().StringVar(&req.Params.JobId, "job-id", "", "The job id (required)")
	cmd.Flags().StringVar(&req.Params.WebhookId, "webhook-id", "", "The webhook id (required)")
	cmd.Flags().StringVar(&req.Params.Event, "event", "", "The event associated with the notification. Possible values: job.completed (required)")

	cmd.MarkFlagRequired("job-id")
	cmd.MarkFlagRequired("webhook-id")
	cmd.MarkFlagRequired("event")

	return cmd
}
