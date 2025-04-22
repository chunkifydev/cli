package webhooks

import (
	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/spf13/cobra"
)

// GetCmd represents the command for getting webhook information
type GetCmd struct {
	Id   string           `json:"id"` // Id of the webhook to retrieve
	Data chunkify.Webhook // Data contains the retrieved webhook information
}

// Execute retrieves the webhook with the specified ID
func (r *GetCmd) Execute() error {
	source, err := cmd.Config.Client.Webhook(r.Id)
	if err != nil {
		return err
	}

	r.Data = source

	return nil
}

// View displays the webhook information using the list view format
func (r *GetCmd) View() {
	sourceList := ListCmd{Data: []chunkify.Webhook{r.Data}}
	sourceList.View()
}

// newGetCmd creates and returns a new cobra command for getting webhook info
func newGetCmd() *cobra.Command {
	req := GetCmd{}

	cmd := &cobra.Command{
		Use:   "get webhook-id",
		Short: "get info about a webhook",
		Long:  "get info about a webhook",
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
