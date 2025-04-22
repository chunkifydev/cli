package projects

import (
	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/spf13/cobra"
)

// GetCmd represents the command for retrieving project information
type GetCmd struct {
	Id   string           `json:"id"` // The ID of the project to retrieve
	Data chunkify.Project // The retrieved project data
}

// Execute retrieves the project with the specified ID
func (r *GetCmd) Execute() error {
	project, err := cmd.Config.Client.Project(r.Id)
	if err != nil {
		return err
	}

	r.Data = project

	return nil
}

// View displays the retrieved project information
func (r *GetCmd) View() {
	sourceList := ListCmd{Data: []chunkify.Project{r.Data}}
	sourceList.View()
}

// newGetCmd creates and configures a new cobra command for retrieving project information
func newGetCmd() *cobra.Command {
	req := GetCmd{}

	cmd := &cobra.Command{
		Use:   "get project-id",
		Short: "get info about a project",
		Long:  "get info about a project",
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
