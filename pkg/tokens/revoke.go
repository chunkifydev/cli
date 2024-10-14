package tokens

import (
	"fmt"

	"github.com/level63/cli/pkg/api"
	"github.com/spf13/cobra"
)

type RevokeCmd struct {
	Id string
}

func (r *RevokeCmd) Execute() error {
	_, err := api.ApiRequest[api.EmptyResponse](
		api.Request{
			Config: cmd.Config,
			Path:   fmt.Sprintf("/api/tokens/%s", r.Id),
			Method: "DELETE",
			Body:   r,
		})

	if err != nil {
		return err
	}

	return nil
}

func (r *RevokeCmd) View() {
	fmt.Println("Token revoked")
}

func newDeleteCmd() *cobra.Command {
	req := RevokeCmd{}

	cmd := &cobra.Command{
		Use:   "revoke token-id",
		Short: "Revoke a token",
		Long:  `Revoke a token`,
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
