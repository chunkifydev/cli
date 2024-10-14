package tokens

import (
	"encoding/json"
	"fmt"

	"github.com/level63/cli/pkg/api"
	"github.com/spf13/cobra"
)

type CreateCmd struct {
	Name      string    `json:"name,omitempty"`
	Scope     string    `json:"scope"`
	ProjectId string    `json:"project_id,omitempty"`
	Data      api.Token `json:"-"`
}

func (r *CreateCmd) Execute() error {
	project, err := api.ApiRequest[api.Token](api.Request{Config: cmd.Config, Path: "/api/tokens", Method: "POST", Body: r})
	if err != nil {
		return err
	}

	r.Data = project

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

	cmd.Flags().StringVar(&req.Name, "name", "", "The name of your access token")
	cmd.Flags().StringVar(&req.Scope, "scope", "", "The access token scope: account or project (required)")
	cmd.Flags().StringVar(&req.ProjectId, "project-id", "", "The created access token will have permissions to create jobs for the given project id")

	cmd.MarkFlagRequired("scope")
	return cmd
}
