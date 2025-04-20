package projects

import (
	"fmt"

	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/spf13/cobra"
)

type UpdateCmd struct {
	Params chunkify.ProjectUpdateParams
}

func (r *UpdateCmd) Execute() error {
	err := cmd.Config.Client.ProjectUpdate(r.Params)
	if err != nil {
		return err
	}

	return nil
}

func (r *UpdateCmd) View() {
	fmt.Println("Project updated")
}

func newUpdateCmd() *cobra.Command {
	req := UpdateCmd{}

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a project",
		Long:  `Update a project`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			req.Params.Id = args[0]
			if err := req.Execute(); err != nil {
				printError(err)
				return
			}
			req.View()
		},
	}

	cmd.Flags().StringVar(&req.Params.Name, "name", "", "Rename this project")
	cmd.Flags().StringVar(&req.Params.StorageId, "storage-id", "", "Change the default storage for this project")
	cmd.MarkFlagsOneRequired("name", "storage-id")

	return cmd
}
