package sources

import (
	"encoding/json"
	"fmt"

	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/spf13/cobra"
)

type CreateCmd struct {
	Params   chunkify.SourceCreateParams
	metadata string          `json:"-"`
	Data     chunkify.Source `json:"-"`
}

func (r *CreateCmd) Execute() error {
	if r.metadata != "" {
		if err := json.Unmarshal([]byte(r.metadata), &r.Params.Metadata); err != nil {
			return fmt.Errorf("invalid JSON format for metadata: %v", err)
		}
	}

	source, err := cmd.Config.Client.SourceCreate(r.Params)
	if err != nil {
		return err
	}

	r.Data = source

	return nil
}

func (r *CreateCmd) View() {
	sourceList := ListCmd{Data: []chunkify.Source{r.Data}}
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

	cmd.Flags().StringVar(&req.Params.Url, "url", "", "The url of the source (required)")
	cmd.Flags().StringVar(&req.metadata, "metadata", "", "Optional metadata in JSON format")
	cmd.MarkFlagRequired("url")

	return cmd
}
