package jobs

import (
	"encoding/json"
	"fmt"

	"github.com/level63/cli/pkg/api"
	"github.com/level63/cli/pkg/sources"
	"github.com/spf13/cobra"
)

type CreateCmd struct {
	SourceId   string           `json:"source_id"`
	Metadata   map[string]any   `json:"metadata,omitempty"`
	Storage    storageParams    `json:"storage,omitempty"`
	Template   templateParams   `json:"template"`
	Transcoder transcoderParams `json:"transcoder"`

	cpu         int64  `json:"-"`
	config      string `json:"-"`
	metadata    string `json:"-"`
	interactive bool   `json:"-"`
	sourceUrl   string `json:"-"`
	Data        api.Job
}

type templateParams struct {
	Name    string         `json:"name"`
	Version string         `json:"version"`
	Config  map[string]any `json:"config"`
}

type transcoderParams struct {
	Quantity int64  `json:"quantity"`
	Type     string `json:"type"`
}

type storageParams struct {
	Path string `json:"path,omitempty"`
	Name string `json:"name,omitempty"`
}

func (r *CreateCmd) Execute() error {
	if r.metadata != "" {
		if err := json.Unmarshal([]byte(r.metadata), &r.Metadata); err != nil {
			return err
		}
	}

	if r.config != "" {
		if err := json.Unmarshal([]byte(r.config), &r.Template.Config); err != nil {
			return err
		}
	}

	r.Transcoder.Type = fmt.Sprintf("cpu:%d", r.cpu)

	job, err := api.ApiRequest[api.Job](api.Request{Config: cmd.Config, Path: "/api/jobs", Method: "POST", Body: r})
	if err != nil {
		return err
	}

	r.Data = job

	return nil
}

func (r *CreateCmd) View() {
	jobList := &ListCmd{CreatedSort: "asc", SourceId: r.SourceId, interactive: r.interactive}
	jobList.Execute()
	if r.interactive {
		StartPolling(jobList)
	} else {
		jobList.View()
	}
}

func newCreateCmd() *cobra.Command {
	req := CreateCmd{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new job",
		Long:  `Create a new job`,
		Run: func(cmd *cobra.Command, args []string) {
			if req.sourceUrl != "" {
				sourceCreateCmd := &sources.CreateCmd{Url: req.sourceUrl}
				if err := sourceCreateCmd.Execute(); err != nil {
					printError(err)
					return
				}
				req.SourceId = sourceCreateCmd.Data.Id
			}

			if err := req.Execute(); err != nil {
				printError(err)
				return
			}
			req.View()
		},
	}

	cmd.Flags().StringVar(&req.SourceId, "source-id", "", "The source id (required)")
	cmd.Flags().StringVar(&req.metadata, "metadata", "", "Optional metadata. Format is key=value")
	cmd.Flags().StringVar(&req.Template.Name, "template", "mp4", "Template name: mp4, hls, jpg")
	cmd.Flags().StringVar(&req.Template.Version, "version", "x264-v1", "Template version: x264-v1, x265-v1, av1-v1, v1")
	cmd.Flags().StringVar(&req.config, "config", "", "Template config in JSON")

	cmd.Flags().Int64Var(&req.Transcoder.Quantity, "transcoder", 1, "Number of transcoders: 1 to 50")
	cmd.Flags().Int64Var(&req.cpu, "cpu", 4, "Number of CPU per transcoder: 2, 4, 8, 16")

	cmd.Flags().StringVar(&req.Storage.Name, "storage", "", "The storage name (default: your default storage)")
	cmd.Flags().StringVar(&req.Storage.Path, "path", "", "The destination path on your storage")

	cmd.Flags().BoolVarP(&req.interactive, "interactive", "i", false, "Refresh the list in real time")

	cmd.Flags().StringVar(&req.sourceUrl, "source-url", "", "Create the job with the given source url")

	cmd.MarkFlagsMutuallyExclusive("source-id", "source-url")

	return cmd
}
