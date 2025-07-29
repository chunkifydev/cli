package tokens

import (
	"encoding/json"
	"fmt"

	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/cli/pkg/flags"
	"github.com/spf13/cobra"
)

// CreateCmd represents the command for creating a new access token
type CreateCmd struct {
	Params chunkify.TokenCreateParams // Parameters for token creation
	Data   chunkify.Token             `json:"-"` // The created token data
}

// Execute creates a new access token with the specified parameters
func (r *CreateCmd) Execute() error {
	token, err := cmd.Config.Client.TokenCreate(r.Params)
	if err != nil {
		return err
	}

	r.Data = token

	return nil
}

// View displays the created token information
// If JSON output is enabled, it prints the data in JSON format
// Otherwise, it displays just the token string
func (r *CreateCmd) View() {
	if cmd.Config.JSON {
		dataBytes, err := json.MarshalIndent(r.Data, "", "  ")
		if err != nil {
			printError(err)
			return
		}
		fmt.Println(string(dataBytes))
		return
	}

	fmt.Println(r.Data.Token)
}

// newCreateCmd creates and configures a new cobra command for creating access tokens
func newCreateCmd() *cobra.Command {
	req := CreateCmd{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new token",
		Long:  `Create a new token`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := req.Execute(); err != nil {
				printError(err)
				return
			}
			req.View()
		},
	}

	flags.StringVar(cmd.Flags(), &req.Params.Name, "name", "", "The name of your access token")
	flags.StringVar(cmd.Flags(), &req.Params.Scope, "scope", "", "The access token scope: team or project (required)")
	flags.StringVarPtr(cmd.Flags(), &req.Params.ProjectId, "project-id", "", "The created access token will have permissions to create jobs for the given project slug")

	cmd.MarkFlagRequired("scope")
	return cmd
}
