package projects

import (
	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/spf13/cobra"
)

type GetCmd struct {
	Id   string `json:"id"`
	Data chunkify.Project
}

func (r *GetCmd) Execute() error {
	project, err := cmd.Config.Client.Project(r.Id)
	if err != nil {
		return err
	}

	r.Data = project

	return nil
}

func (r *GetCmd) View() {
	sourceList := ListCmd{Data: []chunkify.Project{r.Data}}
	sourceList.View()
}

func newGetCmd() *cobra.Command {
	req := GetCmd{}

	cmd := &cobra.Command{
		Use:   "get project-id",
		Short: "get info about a project",
		Long:  "get info about a project",
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
