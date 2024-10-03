package sources

import (
	"encoding/json"

	"github.com/level63/cli/pkg/api"
	"github.com/spf13/cobra"
)

type CreateCmd struct {
	Url      string         `json:"url"`
	metadata string         `json:"-"`
	Metadata map[string]any `json:"metadata,omitempty"`
	Data     api.Source     `json:"-"`
}

func (r *CreateCmd) Execute() error {
	if r.metadata != "" {
		if err := json.Unmarshal([]byte(r.metadata), &r.Metadata); err != nil {
			return err
		}
	}

	source, err := api.ApiRequest[api.Source](api.Request{Config: cmd.Config, Path: "/api/sources", Method: "POST", Body: r})
	if err != nil {
		return err
	}

	r.Data = source

	return nil
}

func (r *CreateCmd) View() {
	sourceList := ListCmd{Data: []api.Source{r.Data}}
	sourceList.View()
}

func newCreateCmd() *cobra.Command {
	req := CreateCmd{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new source",
		Long:  `Create a new source`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := req.Execute(); err != nil {
				printError(err)
				return
			}
			req.View()
		},
	}

	cmd.Flags().StringVar(&req.Url, "url", "", "The url of the source (required)")
	cmd.Flags().StringVar(&req.metadata, "metadata", "", "Optional metadata. Format is key=value")
	cmd.MarkFlagRequired("url")

	return cmd
}
