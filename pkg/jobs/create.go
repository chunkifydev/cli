package jobs

import (
	"encoding/json"
	"fmt"

	"github.com/chunkifydev/cli/pkg/api"
	"github.com/chunkifydev/cli/pkg/formatter"
	"github.com/chunkifydev/cli/pkg/sources"
	"github.com/spf13/cobra"
)

type CreateCmd struct {
	SourceId   string           `json:"source_id"`
	Metadata   map[string]any   `json:"metadata,omitempty"`
	Storage    storageParams    `json:"storage,omitempty"`
	Template   templateParams   `json:"template"`
	Transcoder transcoderParams `json:"transcoder"`

	vcpu        int64   `json:"-"`
	metadata    string  `json:"-"`
	interactive bool    `json:"-"`
	sourceUrl   string  `json:"-"`
	Data        api.Job `json:"-"`
}

type templateParams struct {
	Name         string             `json:"name"`
	Config       api.FfmpegTemplate `json:"config"`
	videoBitrate string
	audioBitrate string
}

type transcoderParams struct {
	Quantity int64  `json:"quantity,omitempty"`
	Type     string `json:"type,omitempty"`
}

type storageParams struct {
	Path string `json:"path,omitempty"`
	Name string `json:"name,omitempty"`
}

func (r *CreateCmd) Execute() error {
	if r.Template.videoBitrate != "" {
		vb, err := formatter.ParseFileSize(r.Template.videoBitrate)
		if err == nil {
			r.Template.Config.VideoBitrate = vb
		}
	}

	if r.Template.audioBitrate != "" {
		ab, err := formatter.ParseFileSize(r.Template.audioBitrate)
		if err == nil {
			r.Template.Config.AudioBitrate = ab
		}
	}

	if r.metadata != "" {
		if err := json.Unmarshal([]byte(r.metadata), &r.Metadata); err != nil {
			return err
		}
	}

	if r.vcpu > 0 {
		r.Transcoder.Type = fmt.Sprintf("%dvCPU", r.vcpu)
	}

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
	cmd.Flags().StringVar(&req.Template.Name, "format", "mp4/x264", "Template name: mp4/x264, mp4/x265, mp4/av1, hls/x264, image/jpg")
	cmd.Flags().Int64Var(&req.Transcoder.Quantity, "transcoder", 0, "Number of transcoders: 1 to 50 (required if cpu is set)")
	cmd.Flags().Int64Var(&req.vcpu, "cpu", 0, "Instance vCPU: 2, 4, 8, 16 (required if transcoder is set)")
	cmd.Flags().StringVar(&req.Storage.Name, "storage", "", "The storage name (default: your default storage)")
	cmd.Flags().StringVar(&req.Storage.Path, "path", "", "The destination path on your storage")
	cmd.Flags().BoolVarP(&req.interactive, "interactive", "i", false, "Refresh the list in real time")
	cmd.Flags().StringVar(&req.sourceUrl, "url", "", "Create the job with the given source url")

	// ffmpeg config
	cmd.Flags().StringVar(&req.Template.videoBitrate, "vb", "", "ffmpeg config: VideoBitrate")
	cmd.Flags().StringVar(&req.Template.videoBitrate, "ab", "", "ffmpeg config: AudioBitrate")

	cmd.Flags().Int64Var(&req.Template.Config.Width, "width", 0, "ffmpeg config: Width")
	cmd.Flags().Int64Var(&req.Template.Config.Height, "height", 0, "ffmpeg config: Height")
	cmd.Flags().Float64Var(&req.Template.Config.Framerate, "fps", 0, "ffmpeg config: Framerate")
	cmd.Flags().Int64Var(&req.Template.Config.Gop, "gop", 0, "ffmpeg config: Gop")
	cmd.Flags().Int64Var(&req.Template.Config.Duration, "duration", 0, "ffmpeg config: Duration")
	cmd.Flags().BoolVar(&req.Template.Config.DisableVideo, "vn", false, "ffmpeg config: DisableVideo")
	cmd.Flags().Int64Var(&req.Template.Config.Channels, "channels", 0, "ffmpeg config: Channels")
	cmd.Flags().BoolVar(&req.Template.Config.DisableAudio, "an", false, "ffmpeg config: DisableAudio")
	cmd.Flags().Int64Var(&req.Template.Config.Maxrate, "maxrate", 0, "ffmpeg config: Maxrate")
	cmd.Flags().Int64Var(&req.Template.Config.Bufsize, "bufsize", 0, "ffmpeg config: Bufsize")
	cmd.Flags().StringVar(&req.Template.Config.PixFmt, "pixfmt", "", "ffmpeg config: PixFmt")
	cmd.Flags().Int64Var(&req.Template.Config.Seek, "seek", 0, "ffmpeg config: Seek")
	cmd.Flags().Int64Var(&req.Template.Config.Crf, "crf", 0, "ffmpeg config: Crf")
	cmd.Flags().Int64Var(&req.Template.Config.X264KeyInt, "x264_keyint", 0, "ffmpeg config: X264KeyInt")
	cmd.Flags().Int64Var(&req.Template.Config.X265KeyInt, "x265_keyint", 0, "ffmpeg config: X265KeyInt")
	cmd.Flags().Int64Var(&req.Template.Config.Level, "level", 0, "ffmpeg config: Level")
	cmd.Flags().StringVar(&req.Template.Config.Profilev, "profilev", "", "ffmpeg config: Profilev")
	cmd.Flags().StringVar(&req.Template.Config.Preset, "preset", "", "ffmpeg config: Preset")
	cmd.Flags().Int64Var(&req.Template.Config.HlsTime, "hls_time", 0, "ffmpeg config: HlsTime")
	cmd.Flags().StringVar(&req.Template.Config.HlsSegmentType, "hls_segment_type", "", "ffmpeg config: HlsSegmentType")
	cmd.Flags().BoolVar(&req.Template.Config.HlsEnc, "hls_enc", false, "ffmpeg config: HlsEnc")
	cmd.Flags().StringVar(&req.Template.Config.HlsEncKey, "hls_enc_key", "", "ffmpeg config: HlsEncKey")
	cmd.Flags().StringVar(&req.Template.Config.HlsEncKeyUrl, "hls_enc_key_url", "", "ffmpeg config: HlsEncKeyUrl")
	cmd.Flags().StringVar(&req.Template.Config.HlsEncIv, "hls_enc_iv", "", "ffmpeg config: HlsEncIv")
	cmd.Flags().Int64Var(&req.Template.Config.Interval, "interval", 0, "ffmpeg config: Interval")

	cmd.MarkFlagsRequiredTogether("transcoder", "cpu")
	cmd.MarkFlagsMutuallyExclusive("source-id", "url")

	return cmd
}
