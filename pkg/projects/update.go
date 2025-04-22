package projects

import (
	"fmt"

	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/spf13/cobra"
)

// UpdateCmd represents the command for updating project details
type UpdateCmd struct {
	Params chunkify.ProjectUpdateParams // Parameters for the update operation
}

// Execute updates the project with the specified parameters
func (r *UpdateCmd) Execute() error {
	err := cmd.Config.Client.ProjectUpdate(r.Params)
	if err != nil {
		return err
	}

	return nil
}

// View displays a confirmation message after successful update
func (r *UpdateCmd) View() {
	fmt.Println("Project updated")
}

// newUpdateCmd creates and configures a new cobra command for updating projects
func newUpdateCmd() *cobra.Command {
	req := UpdateCmd{}

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a project",
		Long:  `Update a project`,
		Args:  cobra.ExactArgs(1), // Requires exactly one argument (project ID)
		Run: func(cmd *cobra.Command, args []string) {
			req.Params.Id = args[0]
			if err := req.Execute(); err != nil {
				printError(err)
				return
			}
			req.View()
		},
	}

	// Define flags for updating project properties
	cmd.Flags().StringVar(&req.Params.Name, "name", "", "Rename this project")
	cmd.Flags().StringVar(&req.Params.StorageId, "storage-id", "", "Change the default storage for this project")
	cmd.MarkFlagsOneRequired("name", "storage-id") // Require at least one flag to be set

	return cmd
}
