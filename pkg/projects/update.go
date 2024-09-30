package projects

import (
	"fmt"

	"github.com/level63/cli/pkg/api"
	"github.com/spf13/cobra"
)

type UpdateCmd struct {
	Id      string
	Name    string `json:"name"`
	Storage string `json:"storage"`
}

func (r *UpdateCmd) Execute() error {
	_, err := api.ApiRequest[api.EmptyResponse](
		api.Request{
			Config: cmd.Config,
			Path:   fmt.Sprintf("/api/projects/%s", r.Id),
			Method: "PATCH",
			Body:   r,
		})

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
			req.Id = args[0]
			if err := req.Execute(); err != nil {
				printError(err)
				return
			}
			req.View()
		},
	}

	cmd.Flags().StringVar(&req.Name, "name", "", "Rename this project")
	cmd.Flags().StringVar(&req.Storage, "storage", "", "Change the default storage for this project")
	cmd.MarkFlagsOneRequired("name", "storage")

	return cmd
}
