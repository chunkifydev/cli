package projects

import (
	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/spf13/cobra"
)

type CreateCmd struct {
	Params chunkify.ProjectCreateParams
	Data   chunkify.Project `json:"-"`
}

func (r *CreateCmd) Execute() error {
	project, err := cmd.Config.Client.ProjectCreate(r.Params)

	if err != nil {
		return err
	}

	r.Data = project

	return nil
}

func (r *CreateCmd) View() {
	projectList := ListCmd{Data: []chunkify.Project{r.Data}}
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

	cmd.Flags().StringVar(&req.Params.Name, "name", "", "The name of your project (required)")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("storage")

	return cmd
}
