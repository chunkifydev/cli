package functions

import (
	"fmt"

	"github.com/level63/cli/pkg/api"
	"github.com/spf13/cobra"
)

type EnableCmd struct {
	Id      string `json:"-"`
	Enabled bool   `json:"enabled"`
}

func (r *EnableCmd) Execute() error {
	_, err := api.ApiRequest[api.EmptyResponse](
		api.Request{
			Config: cmd.Config,
			Path:   fmt.Sprintf("/api/functions/%s", r.Id),
			Method: "PATCH",
			Body:   r,
		})
	if err != nil {
		return err
	}

	return nil
}

func (r *EnableCmd) View() {
	fmt.Println("Function is enabled")
}

func newEnableCmd() *cobra.Command {
	req := EnableCmd{Enabled: true}

	cmd := &cobra.Command{
		Use:   "enable function-id",
		Short: "Enable a function",
		Long:  `Enable a function`,
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
