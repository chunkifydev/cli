// Package projects provides functionality for managing and interacting with projects
package projects

import (
	"fmt"

	"github.com/spf13/cobra"
)

// DeleteCmd represents the command for deleting a project
type DeleteCmd struct {
	Id string // The ID of the project to delete
}

// Execute deletes the project with the specified ID
func (r *DeleteCmd) Execute() error {
	err := cmd.Config.Client.ProjectDelete(r.Id)
	if err != nil {
		return err
	}

	return nil
}

// View displays a confirmation message after successful deletion
func (r *DeleteCmd) View() {
	fmt.Println("Project deleted")
}

// newDeleteCmd creates and configures a new cobra command for deleting projects
func newDeleteCmd() *cobra.Command {
	req := DeleteCmd{}

	cmd := &cobra.Command{
		Use:   "delete project-id",
		Short: "Delete a project",
		Long:  `Delete a project`,
		Args:  cobra.ExactArgs(1), // Requires exactly one argument (project ID)
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
