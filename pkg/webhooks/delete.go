package webhooks

import (
	"fmt"

	"github.com/spf13/cobra"
)

// DeleteCmd represents the command for deleting a webhook
type DeleteCmd struct {
	Id string // Id of the webhook to delete
}

// Execute deletes the webhook with the specified ID
func (r *DeleteCmd) Execute() error {
	err := cmd.Config.Client.WebhookDelete(r.Id)
	if err != nil {
		return err
	}

	return nil
}

// View displays a confirmation message after successful deletion
func (r *DeleteCmd) View() {
	fmt.Println("Webhook deleted")
}

// newDeleteCmd creates and returns a new cobra command for webhook deletion
func newDeleteCmd() *cobra.Command {
	req := DeleteCmd{}

	cmd := &cobra.Command{
		Use:   "delete webhook-id",
		Short: "Delete a webhook",
		Long:  `Delete a webhook`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			req.Id = args[0]
			if err := req.Execute(); err != nil {
				printError(err)
				return
			}
			req.View()
		},
	}

	return cmd
}
