package tokens

import (
	"fmt"

	"github.com/spf13/cobra"
)

// RevokeCmd represents the command for revoking an access token
type RevokeCmd struct {
	Id string // The ID of the token to revoke
}

// Execute revokes the access token with the specified ID
func (r *RevokeCmd) Execute() error {
	err := cmd.Config.Client.TokenRevoke(r.Id)
	if err != nil {
		return err
	}

	return nil
}

// View displays a confirmation message after token revocation
func (r *RevokeCmd) View() {
	fmt.Println("Token revoked")
}

// newDeleteCmd creates and configures a new cobra command for revoking access tokens
func newDeleteCmd() *cobra.Command {
	req := RevokeCmd{}

	cmd := &cobra.Command{
		Use:   "revoke token-id",
		Short: "Revoke a token",
		Long:  `Revoke a token`,
		Args:  cobra.ExactArgs(1), // Requires exactly one argument (token ID)
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
