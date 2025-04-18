package notifications

import (
	"encoding/json"
	"fmt"

	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/spf13/cobra"
)

type GetCmd struct {
	Id      string `json:"id"`
	payload bool

	Data chunkify.Notification
}

func (r *GetCmd) Execute() error {

	notifications, err := cmd.Config.Client.Notification(r.Id)
	if err != nil {
		return err
	}

	r.Data = notifications

	return nil
}

func (r *GetCmd) View() {
	if cmd.Config.JSON {
		dataBytes, err := json.MarshalIndent(r.Data, "", "  ")
		if err != nil {
			printError(err)
			return
		}
		fmt.Println(string(dataBytes))
		return
	}

	if r.payload {
		fmt.Println(r.Data.Payload)
		return
	}

	notificationsList := &ListCmd{Data: []chunkify.Notification{r.Data}}
	notificationsList.View()
}

func newGetCmd() *cobra.Command {
	req := GetCmd{}

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get info about a notification",
		Long:  `Get info about a notification"`,
		Args:  cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			req.Id = args[0]
			if err := req.Execute(); err != nil {
				printError(err)
				return
			}

			req.View()
		},
	}

	cmd.Flags().BoolVarP(&req.payload, "payload", "p", false, "Return the webhook payload in JSON")

	return cmd
}
