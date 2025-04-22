// Package storages provides functionality for managing storage configurations
package storages

import (
	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/spf13/cobra"
)

// GetCmd represents the command for retrieving storage information
type GetCmd struct {
	Id   string           // ID of the storage to retrieve
	Data chunkify.Storage // The retrieved storage data
}

// Execute retrieves the storage information for the specified ID
func (r *GetCmd) Execute() error {
	storage, err := cmd.Config.Client.Storage(r.Id)
	if err != nil {
		return err
	}

	r.Data = storage

	return nil
}

// View displays the retrieved storage information
func (r *GetCmd) View() {
	sourceList := ListCmd{Data: []chunkify.Storage{r.Data}}
	sourceList.View()
}

// newGetCmd creates and configures a new cobra command for retrieving storage information
func newGetCmd() *cobra.Command {
	req := GetCmd{}

	cmd := &cobra.Command{
		Use:   "get storage-id",
		Short: "get info about a storage",
		Long:  "get info about a storage",
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
