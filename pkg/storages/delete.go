package storages

import (
	"fmt"

	"github.com/spf13/cobra"
)

type DeleteCmd struct {
	Id string
}

func (r *DeleteCmd) Execute() error {
	err := cmd.Config.Client.StorageDelete(r.Id)
	if err != nil {
		return err
	}

	return nil
}

func (r *DeleteCmd) View() {
	fmt.Println("Storage deleted")
}

func newDeleteCmd() *cobra.Command {
	req := DeleteCmd{}

	cmd := &cobra.Command{
		Use:   "delete storage-id",
		Short: "Delete a storage",
		Long:  `Delete a storage`,
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
