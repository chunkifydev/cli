package tokens

import (
	"encoding/json"
	"fmt"

	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/spf13/cobra"
)

type CreateCmd struct {
	Params chunkify.TokenCreateParams
	Data   chunkify.Token `json:"-"`
}

func (r *CreateCmd) Execute() error {
	token, err := cmd.Config.Client.TokenCreate(r.Params)
	if err != nil {
		return err
	}

	r.Data = token

	return nil
}

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

	cmd.Flags().StringVar(&req.Params.Name, "name", "", "The name of your access token")
	cmd.Flags().StringVar(&req.Params.Scope, "scope", "", "The access token scope: team or project (required)")
	cmd.Flags().StringVar(&req.Params.ProjectId, "project-id", "", "The created access token will have permissions to create jobs for the given project slug")

	cmd.MarkFlagRequired("scope")
	return cmd
}
