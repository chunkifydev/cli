package webhooks

import (
	"github.com/level63/cli/pkg/api"
	"github.com/spf13/cobra"
)

type CreateCmd struct {
	Url     string `json:"url"`
	Enabled bool   `json:"enabled"`
	Data    api.Webhook
}

func (r *CreateCmd) Execute() error {
	project, err := api.ApiRequest[api.Webhook](api.Request{Config: cmd.Config, Path: "/api/webhooks", Method: "POST", Body: r})
	if err != nil {
		return err
	}

	r.Data = project

	return nil
}

func (r *CreateCmd) View() {
	projectList := ListCmd{Data: []api.Webhook{r.Data}}
	projectList.View()
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

	cmd.Flags().StringVar(&req.Url, "url", "", "The webhook URL (required)")
	cmd.Flags().BoolVar(&req.Enabled, "enabled", true, "Enable the webhook")
	cmd.MarkFlagRequired("url")

	return cmd
}
