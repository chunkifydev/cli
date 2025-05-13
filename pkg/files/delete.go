package files

import (
	"fmt"

	"github.com/spf13/cobra"
)

// DeleteCmd represents the command for deleting a webhook
type DeleteCmd struct {
	Id string // Id of the file to delete
}

// Execute deletes the file with the specified ID
func (r *DeleteCmd) Execute() error {
	err := cmd.Config.Client.FileDelete(r.Id)
	if err != nil {
		return err
	}

	return nil
}

// View displays a confirmation message after successful deletion
func (r *DeleteCmd) View() {
	fmt.Println("File deleted")
}

// newDeleteCmd creates and returns a new cobra command for upload deletion
func newDeleteCmd() *cobra.Command {
	req := DeleteCmd{}

	cmd := &cobra.Command{
		Use:   "delete file-id",
		Short: "Delete a file",
		Long:  `Delete a file`,
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
