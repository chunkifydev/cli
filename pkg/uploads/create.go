package uploads

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/cli/pkg/flags"
	"github.com/spf13/cobra"
)

// CreateCmd represents the command for creating a new source
type CreateCmd struct {
	Params   chunkify.UploadCreateParams // Parameters for creating the upload
	metadata string                      `json:"-"` // Raw metadata string to be parsed
	body     io.Reader                   `json:"-"` // Content of the video file to upload
	Data     chunkify.Upload             `json:"-"` // The created upload data
}

// Execute creates a new upload with the specified parameters
func (r *CreateCmd) Execute() error {
	// Parse metadata JSON if provided
	if r.metadata != "" {
		if err := json.Unmarshal([]byte(r.metadata), &r.Params.Metadata); err != nil {
			return fmt.Errorf("invalid JSON format for metadata: %v", err)
		}
	}

	// Create the source via API
	upload, err := cmd.Config.Client.UploadCreate(r.Params)
	if err != nil {
		return err
	}

	r.Data = upload

	if r.body != nil {
		if err := cmd.Config.Client.UploadBlob(r.body, upload); err != nil {
			r.Data.Status = chunkify.UploadStatusFailed
			return err
		}
	}

	return nil
}

// View displays the newly created source information
func (r *CreateCmd) View() {
	uploadList := ListCmd{Data: []chunkify.Upload{r.Data}}
	uploadList.View()
}

// newCreateCmd creates and configures a new cobra command for creating sources
func newCreateCmd() *cobra.Command {
	req := CreateCmd{}
	var path string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new upload",
		Long:  `Create a new upload`,
		Run: func(cmd *cobra.Command, args []string) {
			if path != "" {
				body, err := os.Open(path)
				if err != nil {
					printError(err)
					return
				}
				req.body = body
			}
			if err := req.Execute(); err != nil {
				printError(err)
				return
			}
			req.View()
		},
	}

	// Define flags for upload creation
	flags.StringVar(cmd.Flags(), &req.metadata, "metadata", "", "Optional metadata in JSON format")
	flags.Int64VarPtr(cmd.Flags(), &req.Params.Timeout, "timeout", 1800, "Optional timeout in seconds, default is 1800 seconds")
	flags.StringVar(cmd.Flags(), &path, "path", "", "Path to the video file to upload")

	return cmd
}
