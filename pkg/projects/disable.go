package projects

import (
	"fmt"

	"github.com/level63/cli/pkg/api"
	"github.com/spf13/cobra"
)

type DisableCmd struct {
	Id    string
	Pause bool `json:"pause"`
}

func (r *DisableCmd) Execute() error {
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

func (r *DisableCmd) View() {
	fmt.Println("Project is disabled")
}

func newDisableCmd() *cobra.Command {
	req := DisableCmd{Pause: true}

	cmd := &cobra.Command{
		Use:   "disable project-id",
		Short: "Disable a project",
		Long:  `Disable a project`,
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
