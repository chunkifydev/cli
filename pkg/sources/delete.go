package sources

import (
	"fmt"

	"github.com/spf13/cobra"
)

// DeleteCmd represents the command for deleting a webhook
type DeleteCmd struct {
	Id string // Id of the upload to delete
}

// Execute deletes the source with the specified ID
func (r *DeleteCmd) Execute() error {
	err := cmd.Config.Client.SourceDelete(r.Id)
	if err != nil {
		return err
	}

	return nil
}

// View displays a confirmation message after successful deletion
func (r *DeleteCmd) View() {
	fmt.Println("Source deleted")
}

// newDeleteCmd creates and returns a new cobra command for source deletion
func newDeleteCmd() *cobra.Command {
	req := DeleteCmd{}

	cmd := &cobra.Command{
		Use:   "delete source-id",
		Short: "Delete a source",
		Long:  `Delete a source`,
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
