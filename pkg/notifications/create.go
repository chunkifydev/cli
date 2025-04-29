package notifications

import (
	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/spf13/cobra"
)

// CreateCmd represents the command for creating a new notification
type CreateCmd struct {
	Params chunkify.NotificationCreateParams // Parameters to use when creating the notification
	Data   chunkify.Notification             `json:"-"` // The created notification data
}

// Execute creates a new notification using the provided parameters
func (r *CreateCmd) Execute() error {
	notification, err := cmd.Config.Client.NotificationCreate(r.Params)
	if err != nil {
		return err
	}

	r.Data = notification

	return nil
}

// View displays the newly created notification by showing the notifications list
// filtered to the associated job ID
func (r *CreateCmd) View() {
	notifList := &ListCmd{Params: chunkify.NotificationListParams{CreatedSort: "asc", ObjectId: r.Params.ObjectId}}
	notifList.Execute()
	notifList.View()
}

// newCreateCmd creates and configures a new cobra command for notification creation
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

	// Configure required command flags
	cmd.Flags().StringVar(&req.Params.ObjectId, "object-id", "", "The object id (required)")
	cmd.Flags().StringVar(&req.Params.WebhookId, "webhook-id", "", "The webhook id (required)")
	cmd.Flags().StringVar(&req.Params.Event, "event", "", "The event associated with the notification. Possible values: job.completed (required)")

	cmd.MarkFlagRequired("object-id")
	cmd.MarkFlagRequired("webhook-id")
	cmd.MarkFlagRequired("event")

	return cmd
}
