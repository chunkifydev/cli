package sources

import (
	"encoding/json"
	"fmt"

	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/cli/pkg/flags"
	"github.com/spf13/cobra"
)

// CreateCmd represents the command for creating a new source
type CreateCmd struct {
	Params   chunkify.SourceCreateParams // Parameters for creating the source
	metadata string                      `json:"-"` // Raw metadata string to be parsed
	Data     chunkify.Source             `json:"-"` // The created source data
}

// Execute creates a new source with the specified parameters
func (r *CreateCmd) Execute() error {
	// Parse metadata JSON if provided
	if r.metadata != "" {
		if err := json.Unmarshal([]byte(r.metadata), &r.Params.Metadata); err != nil {
			return fmt.Errorf("invalid JSON format for metadata: %v", err)
		}
	}

	// Create the source via API
	source, err := cmd.Config.Client.SourceCreate(r.Params)
	if err != nil {
		return err
	}

	r.Data = source

	return nil
}

// View displays the newly created source information
func (r *CreateCmd) View() {
	sourceList := ListCmd{Data: []chunkify.Source{r.Data}}
	sourceList.View()
}

// newCreateCmd creates and configures a new cobra command for creating sources
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

	// Define flags for source creation
	flags.StringVar(cmd.Flags(), &req.Params.Url, "url", "", "The url of the source (required)")
	flags.StringVar(cmd.Flags(), &req.metadata, "metadata", "", "Optional metadata in JSON format")
	cmd.MarkFlagRequired("url")

	return cmd
}
