package projects

import (
	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/cli/pkg/flags"
	"github.com/spf13/cobra"
)

// CreateCmd represents the command for creating a new project
type CreateCmd struct {
	Params chunkify.ProjectCreateParams // Parameters to create the project
	Data   chunkify.Project             `json:"-"` // The created project data
}

// Execute creates a new project using the provided parameters
func (r *CreateCmd) Execute() error {
	project, err := cmd.Config.Client.ProjectCreate(r.Params)

	if err != nil {
		return err
	}

	r.Data = project

	return nil
}

// View displays the newly created project data
func (r *CreateCmd) View() {
	projectList := ListCmd{Data: []chunkify.Project{r.Data}}
	projectList.View()
}

// newCreateCmd creates and configures a new cobra command for creating projects
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

	flags.StringVar(cmd.Flags(), &req.Params.Name, "name", "", "The name of your project (required)")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("storage")

	return cmd
}
