package files

import (
	"encoding/json"
	"fmt"

	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/spf13/cobra"
)

// GetCmd represents the command for retrieving a single upload
type GetCmd struct {
	Id   string        `json:"id"` // ID of the file to retrieve
	Data chunkify.File // The retrieved file data
}

// Execute fetches a notification by ID from the API
func (r *GetCmd) Execute() error {
	files, err := cmd.Config.Client.File(r.Id)
	if err != nil {
		return err
	}

	r.Data = files
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

	filesList := &ListCmd{Data: []chunkify.File{r.Data}}
	filesList.View()
}

// newGetCmd creates and configures a new cobra command for retrieving files
func newGetCmd() *cobra.Command {
	req := GetCmd{}

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get info about a file",
		Long:  `Get info about a file"`,
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
