package storages

import (
	"fmt"

	"github.com/spf13/cobra"
)

// DeleteCmd represents the command for deleting a storage configuration
type DeleteCmd struct {
	Id string // ID of the storage to delete
}

// Execute deletes the storage with the specified ID
func (r *DeleteCmd) Execute() error {
	err := cmd.Config.Client.StorageDelete(r.Id)
	if err != nil {
		return err
	}

	return nil
}

// View displays a confirmation message after successful deletion
func (r *DeleteCmd) View() {
	fmt.Println("Storage deleted")
}

// newDeleteCmd creates and configures a new cobra command for deleting storage configurations
func newDeleteCmd() *cobra.Command {
	req := DeleteCmd{}

	cmd := &cobra.Command{
		Use:   "delete storage-id",
		Short: "Delete a storage",
		Long:  `Delete a storage`,
		Args:  cobra.ExactArgs(1), // Requires exactly one argument (storage ID)
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
