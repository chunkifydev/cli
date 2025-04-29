package uploads

import (
	"encoding/json"
	"fmt"

	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/spf13/cobra"
)

// GetCmd represents the command for retrieving a single upload
type GetCmd struct {
	Id   string          `json:"id"` // ID of the upload to retrieve
	Data chunkify.Upload // The retrieved upload data
}

// Execute fetches a notification by ID from the API
func (r *GetCmd) Execute() error {
	uploads, err := cmd.Config.Client.Upload(r.Id)
	if err != nil {
		return err
	}

	r.Data = uploads
	return nil
}

// View displays the notification data in the requested format
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

	uploadsList := &ListCmd{Data: []chunkify.Upload{r.Data}}
	uploadsList.View()
}

// newGetCmd creates and configures a new cobra command for retrieving uploads
func newGetCmd() *cobra.Command {
	req := GetCmd{}

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get info about an upload",
		Long:  `Get info about an upload"`,
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

	return cmd
}
