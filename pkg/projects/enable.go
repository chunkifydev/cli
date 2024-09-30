package projects

import (
	"fmt"

	"github.com/level63/cli/pkg/api"
	"github.com/spf13/cobra"
)

type EnableCmd struct {
	Id    string
	Pause bool `json:"pause"`
}

func (r *EnableCmd) Execute() error {
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

func (r *EnableCmd) View() {
	fmt.Println("Project is enabled")
}

func newEnableCmd() *cobra.Command {
	req := EnableCmd{Pause: false}

	cmd := &cobra.Command{
		Use:   "enable project-id",
		Short: "Enable a project",
		Long:  `Enable a project`,
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
