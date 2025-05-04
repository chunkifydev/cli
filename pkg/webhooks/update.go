package webhooks

import (
	"fmt"

	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/spf13/cobra"
)

// UpdateCmd represents the command for updating a webhook
type UpdateCmd struct {
	Params chunkify.WebhookUpdateParams // Params contains the parameters for updating a webhook
}

// Execute updates the webhook
func (r *UpdateCmd) Execute() error {
	err := cmd.Config.Client.WebhookUpdate(r.Params)
	if err != nil {
		return err
	}

	return nil
}

// View displays a confirmation message after successful disabling
func (r *UpdateCmd) View() {
	fmt.Println("Webhook is updated")
}

// newUpdateCmd creates and returns a new cobra command for webhook updating
func newUpdateCmd() *cobra.Command {
	var (
		enable  bool
		disable bool
		events  []string
	)
	req := UpdateCmd{Params: chunkify.WebhookUpdateParams{}}

	cmd := &cobra.Command{
		Use:   "update webhook-id",
		Short: "Update a webhook",
		Long:  `Update a webhook`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			req.Params.WebhookId = args[0]
			if cmd.Flags().Changed("enable") {
				enable = true
				req.Params.Enabled = &enable
			}
			if cmd.Flags().Changed("disable") {
				enable = false
				req.Params.Enabled = &enable
			}
			if cmd.Flags().Changed("events") {
				if len(events) > 0 {
					req.Params.Events = &events
				}
			}
			if err := req.Execute(); err != nil {
				printError(err)
				return
			}
			req.View()
		},
	}

	// Add command flags for configuring the webhook
	cmd.Flags().BoolVar(&enable, "enable", true, "Enable the webhook")
	cmd.Flags().BoolVar(&disable, "disable", true, "Disable the webhook")
	cmd.Flags().StringArrayVar(&events, "events", []string{"*"}, "Update webhook events. *, job.*, job.completed, job.failed, upload.* upload.completed, upload.failed, upload.expired")

	cmd.MarkFlagsMutuallyExclusive("enable", "disable")
	cmd.MarkFlagsOneRequired("enable", "disable", "events")

	return cmd
}
