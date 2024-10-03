package projects

import (
	"github.com/level63/cli/pkg/api"
	"github.com/spf13/cobra"
)

type CreateCmd struct {
	Name    string      `json:"name"`
	Storage string      `json:"storage"`
	Data    api.Project `json:"-"`
}

func (r *CreateCmd) Execute() error {
	project, err := api.ApiRequest[api.Project](api.Request{Config: cmd.Config, Path: "/api/projects", Method: "POST", Body: r})
	if err != nil {
		return err
	}

	r.Data = project

	return nil
}

func (r *CreateCmd) View() {
	projectList := ListCmd{Data: []api.Project{r.Data}}
	projectList.View()
}

func newCreateCmd() *cobra.Command {
	req := CreateCmd{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new project",
		Long:  `Create a new project`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := req.Execute(); err != nil {
				printError(err)
				return
			}
			req.View()
		},
	}

	cmd.Flags().StringVar(&req.Name, "name", "", "The name of your project (required)")
	cmd.Flags().StringVar(&req.Storage, "storage", "", "The storage to use for this project (required)")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("storage")

	return cmd
}
