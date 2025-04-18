package jobs

import (
	"encoding/json"
	"fmt"

	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/cli/pkg/formatter"
	"github.com/chunkifydev/cli/pkg/sources"
	"github.com/spf13/cobra"
)

type CreateCmd struct {
	Params chunkify.JobCreateParams `json:"-"`

	vcpu         int64        `json:"-"`
	metadata     string       `json:"-"`
	interactive  bool         `json:"-"`
	sourceUrl    string       `json:"-"`
	Data         chunkify.Job `json:"-"`
	videoBitrate string
	audioBitrate string
}

func (r *CreateCmd) Execute() error {
	// Handle metadata
	if r.metadata != "" {
		if err := json.Unmarshal([]byte(r.metadata), &r.Params.Metadata); err != nil {
			return err
		}
	}

	// Handle source URL if provided
	if r.sourceUrl != "" {
		sourceCreateCmd := &sources.CreateCmd{
			Params: chunkify.SourceCreateParams{
				Url: r.sourceUrl,
			},
		}
		if err := sourceCreateCmd.Execute(); err != nil {
			return err
		}
		r.Params.SourceId = sourceCreateCmd.Data.Id
	}

	// Set transcoder type if CPU is specified
	if r.vcpu > 0 {
		r.Params.Transcoder.Type = fmt.Sprintf("%dvCPU", r.vcpu)
	}

	// Create the job
	job, err := cmd.Config.Client.JobCreate(r.Params)
	if err != nil {
		return err
	}

	r.Data = job
	return nil
}

func (r *CreateCmd) View() {
	jobList := &ListCmd{Params: chunkify.JobListParams{CreatedSort: "asc", SourceId: r.Params.SourceId}, interactive: r.interactive}
	jobList.Execute()
	if r.interactive {
		StartPolling(jobList)
	} else {
		jobList.View()
	}
}

func newCreateCmd() *cobra.Command {
	req := &CreateCmd{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new job",
		Long:  `Create a new job`,
	}

	// Add common flags that will be shared across all subcommands
	cmd.PersistentFlags().StringVar(&req.Params.SourceId, "source-id", "", "The source id (required)")
	cmd.PersistentFlags().StringVar(&req.metadata, "metadata", "", "Optional metadata. Format is key=value")
	cmd.PersistentFlags().Int64Var(&req.Params.Transcoder.Quantity, "transcoder", 0, "Number of transcoders: 1 to 50 (required if cpu is set)")
	cmd.PersistentFlags().Int64Var(&req.vcpu, "cpu", 0, "Instance vCPU: 4, 8, 16 (required if transcoder is set)")
	cmd.PersistentFlags().StringVar(&req.Params.Storage.Id, "storage", "", "The storage id (default is the project default storage id)")
	cmd.PersistentFlags().StringVar(&req.Params.Storage.Path, "path", "", "The destination path on your storage")
	cmd.PersistentFlags().BoolVarP(&req.interactive, "interactive", "i", false, "Refresh the list in real time")
	cmd.PersistentFlags().StringVar(&req.sourceUrl, "url", "", "Create the job with the given source url")
	cmd.PersistentFlags().StringVar(&req.videoBitrate, "vb", "", "ffmpeg config: VideoBitrate")
	cmd.PersistentFlags().StringVar(&req.audioBitrate, "ab", "", "ffmpeg config: AudioBitrate")

	// Add subcommands
	cmd.AddCommand(newX264Cmd(req))
	cmd.AddCommand(newX265Cmd(req))
	cmd.AddCommand(newAv1Cmd(req))
	cmd.AddCommand(newVp9Cmd(req))
	cmd.AddCommand(newHlsX264Cmd(req))
	cmd.AddCommand(newHlsX265Cmd(req))
	cmd.AddCommand(newHlsAv1Cmd(req))
	cmd.AddCommand(newJpgCmd(req))

	cmd.MarkFlagsRequiredTogether("transcoder", "cpu")
	cmd.MarkFlagsMutuallyExclusive("source-id", "url")

	return cmd
}

func newX264Cmd(req *CreateCmd) *cobra.Command {
	// Create x264 config
	x264Config := &chunkify.FfmpegX264{}

	cmd := &cobra.Command{
		Use:   "mp4/x264",
		Short: "Create a job with x264 encoding",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Set the template name for x264
			req.Params.Format.Name = "mp4/x264"

			// Parse bitrates
			setBitrate(req.videoBitrate, req.audioBitrate, &x264Config.FfmpegVideo)

			// Set the config
			req.Params.Format.Config = x264Config
			// Execute the job creation
			if err := req.Execute(); err != nil {
				return err
			}

			req.View()
			return nil
		},
	}

	// Add x264 specific flags
	setX264Flags(cmd, x264Config)

	// Add common video flags
	setCommonVideoFlags(cmd, &x264Config.FfmpegVideo)
	setCommonFlags(cmd, &x264Config.FfmpegCommon)

	return cmd
}

func newX265Cmd(req *CreateCmd) *cobra.Command {
	// Create x265 config
	x265Config := &chunkify.FfmpegX265{}

	cmd := &cobra.Command{
		Use:   "mp4/x265",
		Short: "Create a job with x265 encoding",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Set the template name for x264
			req.Params.Format.Name = "mp4/x265"

			// Parse bitrates
			setBitrate(req.videoBitrate, req.audioBitrate, &x265Config.FfmpegVideo)

			// Set the config
			req.Params.Format.Config = x265Config
			// Execute the job creation
			if err := req.Execute(); err != nil {
				return err
			}

			req.View()
			return nil
		},
	}

	// Add x264 specific flags
	setX265Flags(cmd, x265Config)

	// Add common video flags
	setCommonVideoFlags(cmd, &x265Config.FfmpegVideo)
	setCommonFlags(cmd, &x265Config.FfmpegCommon)

	return cmd
}

func newAv1Cmd(req *CreateCmd) *cobra.Command {
	// Create x265 config
	av1Config := &chunkify.FfmpegAv1{}

	cmd := &cobra.Command{
		Use:   "mp4/av1",
		Short: "Create a job with av1 encoding",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Set the template name for x264
			req.Params.Format.Name = "mp4/av1"

			// Parse bitrates
			setBitrate(req.videoBitrate, req.audioBitrate, &av1Config.FfmpegVideo)

			// Set the config
			req.Params.Format.Config = av1Config
			// Execute the job creation
			if err := req.Execute(); err != nil {
				return err
			}

			req.View()
			return nil
		},
	}

	// Add x264 specific flags
	setAv1Flags(cmd, av1Config)

	// Add common video flags
	setCommonVideoFlags(cmd, &av1Config.FfmpegVideo)
	setCommonFlags(cmd, &av1Config.FfmpegCommon)

	return cmd
}

func newVp9Cmd(req *CreateCmd) *cobra.Command {
	// Create x265 config
	vp9Config := &chunkify.FfmpegVp9{}

	cmd := &cobra.Command{
		Use:   "webm/vp9",
		Short: "Create a job with vp9 encoding",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Set the template name for x264
			req.Params.Format.Name = "webm/vp9"

			// Parse bitrates
			setBitrate(req.videoBitrate, req.audioBitrate, &vp9Config.FfmpegVideo)

			// Set the config
			req.Params.Format.Config = vp9Config
			// Execute the job creation
			if err := req.Execute(); err != nil {
				return err
			}

			req.View()
			return nil
		},
	}

	// Add x264 specific flags
	setVp9Flags(cmd, vp9Config)

	// Add common video flags
	setCommonVideoFlags(cmd, &vp9Config.FfmpegVideo)
	setCommonFlags(cmd, &vp9Config.FfmpegCommon)

	return cmd
}

func newHlsX264Cmd(req *CreateCmd) *cobra.Command {
	// Create x265 config
	hlsX264Config := &chunkify.FfmpegHlsX264{}

	cmd := &cobra.Command{
		Use:   "hls/x264",
		Short: "Create a job with hls x264 encoding",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Set the template name for x264
			req.Params.Format.Name = "hls/x264"

			// Parse bitrates
			setBitrate(req.videoBitrate, req.audioBitrate, &hlsX264Config.FfmpegVideo)

			// Set the config
			req.Params.Format.Config = hlsX264Config
			// Execute the job creation
			if err := req.Execute(); err != nil {
				return err
			}

			req.View()
			return nil
		},
	}

	// Add HLS specific flags
	setHLSFlags(cmd, &hlsX264Config.FfmpegHls)

	// Add x264 specific flags
	setX264Flags(cmd, &hlsX264Config.FfmpegX264)

	// Add common video flags
	setCommonVideoFlags(cmd, &hlsX264Config.FfmpegVideo)
	setCommonFlags(cmd, &hlsX264Config.FfmpegCommon)

	return cmd
}

func newHlsX265Cmd(req *CreateCmd) *cobra.Command {
	// Create x265 config
	hlsX265Config := &chunkify.FfmpegHlsX265{}

	cmd := &cobra.Command{
		Use:   "hls/x265",
		Short: "Create a job with hls x265 encoding",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Set the template name for x264
			req.Params.Format.Name = "hls/x265"

			// Parse bitrates
			setBitrate(req.videoBitrate, req.audioBitrate, &hlsX265Config.FfmpegVideo)

			// Set the config
			req.Params.Format.Config = hlsX265Config
			// Execute the job creation
			if err := req.Execute(); err != nil {
				return err
			}

			req.View()
			return nil
		},
	}

	// Add HLS specific flags
	setHLSFlags(cmd, &hlsX265Config.FfmpegHls)

	// Add x265 specific flags
	setX265Flags(cmd, &hlsX265Config.FfmpegX265)

	// Add common video flags
	setCommonVideoFlags(cmd, &hlsX265Config.FfmpegVideo)
	setCommonFlags(cmd, &hlsX265Config.FfmpegCommon)

	return cmd
}

func newHlsAv1Cmd(req *CreateCmd) *cobra.Command {
	// Create x265 config
	hlsAv1Config := &chunkify.FfmpegHlsAv1{}

	cmd := &cobra.Command{
		Use:   "hls/av1",
		Short: "Create a job with hls av1 encoding",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Set the template name for x264
			req.Params.Format.Name = "hls/av1"

			// Parse bitrates
			setBitrate(req.videoBitrate, req.audioBitrate, &hlsAv1Config.FfmpegVideo)

			// Set the config
			req.Params.Format.Config = hlsAv1Config
			// Execute the job creation
			if err := req.Execute(); err != nil {
				return err
			}

			req.View()
			return nil
		},
	}

	// Add HLS specific flags
	setHLSFlags(cmd, &hlsAv1Config.FfmpegHls)

	// Add av1 specific flags
	setAv1Flags(cmd, &hlsAv1Config.FfmpegAv1)

	// Add common video flags
	setCommonVideoFlags(cmd, &hlsAv1Config.FfmpegVideo)
	setCommonFlags(cmd, &hlsAv1Config.FfmpegCommon)

	return cmd
}

func newJpgCmd(req *CreateCmd) *cobra.Command {
	// Create x265 config
	jpgConfig := &chunkify.FfmpegJpg{}

	cmd := &cobra.Command{
		Use:   "jpg",
		Short: "Create a job with jpg encoding",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Set the template name for x264
			req.Params.Format.Name = "jpg"

			// Set the config
			req.Params.Format.Config = jpgConfig
			// Execute the job creation
			if err := req.Execute(); err != nil {
				return err
			}

			req.View()
			return nil
		},
	}

	// Add image specific flags
	cmd.Flags().BoolVar(&jpgConfig.Sprite, "sprite", false, "Generate sprite images instead of many single images")
	cmd.Flags().Int64Var(&jpgConfig.Frames, "frames", 0, "Generate only the number of givenframes")
	cmd.Flags().Int64Var(&jpgConfig.Width, "width", 0, "ffmpeg config: Width")
	cmd.Flags().Int64Var(&jpgConfig.Height, "height", 0, "ffmpeg config: Height")
	cmd.Flags().Int64Var(&jpgConfig.Interval, "interval", 0, "ffmpeg config: Interval")
	//cmd.Flags().Int64Var(&jpgConfig.ChunkDuration, "chunk_duration", 0, "ffmpeg config: ChunkDuration")
	// Add common flags
	setCommonFlags(cmd, &jpgConfig.FfmpegCommon)

	return cmd
}

func setCommonVideoFlags(cmd *cobra.Command, videoCommon *chunkify.FfmpegVideo) {

	cmd.Flags().Int64Var(&videoCommon.Width, "width", 0, "ffmpeg config: Width")
	cmd.Flags().Int64Var(&videoCommon.Height, "height", 0, "ffmpeg config: Height")
	cmd.Flags().Float64Var(&videoCommon.Framerate, "framerate", 0, "ffmpeg config: Framerate")
	cmd.Flags().Int64Var(&videoCommon.Gop, "gop", 0, "ffmpeg config: Gop")
	cmd.Flags().Int64Var(&videoCommon.Channels, "channels", 0, "ffmpeg config: Channels")
	cmd.Flags().Int64Var(&videoCommon.Maxrate, "maxrate", 0, "ffmpeg config: Maxrate")
	cmd.Flags().Int64Var(&videoCommon.Bufsize, "bufsize", 0, "ffmpeg config: Bufsize")
	cmd.Flags().StringVar(&videoCommon.PixFmt, "pixfmt", "", "ffmpeg config: PixFmt")
	cmd.Flags().BoolVar(&videoCommon.DisableAudio, "an", false, "ffmpeg config: DisableAudio")
	cmd.Flags().BoolVar(&videoCommon.DisableVideo, "vn", false, "ffmpeg config: DisableVideo")
}

func setBitrate(vb, ab string, videoCommon *chunkify.FfmpegVideo) {
	if vb != "" {
		if vb, err := formatter.ParseFileSize(vb); err == nil {
			videoCommon.VideoBitrate = vb
		}
	}
	if ab != "" {
		if ab, err := formatter.ParseFileSize(ab); err == nil {
			videoCommon.AudioBitrate = ab
		}
	}
}

// Helper function for common settings
func setCommonFlags(cmd *cobra.Command, common *chunkify.FfmpegCommon) {
	cmd.Flags().Int64Var(&common.Duration, "duration", 0, "ffmpeg config: Duration")
	cmd.Flags().Int64Var(&common.Seek, "seek", 0, "ffmpeg config: Seek")
}

// Helper function for X264 codec flags
func setX264Flags(cmd *cobra.Command, x264 *chunkify.FfmpegX264) {
	cmd.Flags().Int64Var(&x264.X264KeyInt, "x264_keyint", 0, "ffmpeg config: X264KeyInt")
	cmd.Flags().Int64Var(&x264.Level, "level", 0, "ffmpeg config: Level")
	cmd.Flags().StringVar(&x264.Profilev, "profilev", "", "ffmpeg config: Profilev")
	cmd.Flags().Int64Var(&x264.Crf, "crf", 0, "ffmpeg config: Crf")
	cmd.Flags().StringVar(&x264.Preset, "preset", "", "ffmpeg config: Preset")

}

// Helper function for X265 codec flags
func setX265Flags(cmd *cobra.Command, x265 *chunkify.FfmpegX265) {
	cmd.Flags().Int64Var(&x265.X265KeyInt, "x265_keyint", 0, "ffmpeg config: X265KeyInt")
	cmd.Flags().Int64Var(&x265.Level, "level", 0, "ffmpeg config: Level")
	cmd.Flags().StringVar(&x265.Profilev, "profilev", "", "ffmpeg config: Profilev")
	cmd.Flags().Int64Var(&x265.Crf, "crf", 0, "ffmpeg config: Crf")
	cmd.Flags().StringVar(&x265.Preset, "preset", "", "ffmpeg config: Preset")
}

func setVp9Flags(cmd *cobra.Command, vp9 *chunkify.FfmpegVp9) {
	cmd.Flags().Int64Var(&vp9.Crf, "crf", 0, "ffmpeg config: Crf")
	cmd.Flags().StringVar(&vp9.CpuUsed, "cpu_used", "", "ffmpeg config: CpuUsed")
	cmd.Flags().StringVar(&vp9.Quality, "quality", "", "ffmpeg config: Quality")
}

// Helper function for AV1 codec flags
func setAv1Flags(cmd *cobra.Command, av1 *chunkify.FfmpegAv1) {
	cmd.Flags().Int64Var(&av1.Level, "level", 0, "ffmpeg config: Level")
	cmd.Flags().StringVar(&av1.Profilev, "profilev", "", "ffmpeg config: Profilev")
	cmd.Flags().Int64Var(&av1.Crf, "crf", 0, "ffmpeg config: Crf")
	cmd.Flags().StringVar(&av1.Preset, "preset", "", "ffmpeg config: Preset")
}

// Helper function for HLS flags
func setHLSFlags(cmd *cobra.Command, hls *chunkify.FfmpegHls) {
	cmd.Flags().Int64Var(&hls.HlsTime, "hls_time", 0, "ffmpeg config: HlsTime")
	cmd.Flags().StringVar(&hls.HlsSegmentType, "hls_segment_type", "", "ffmpeg config: HlsSegmentType")
	cmd.Flags().BoolVar(&hls.HlsEnc, "hls_enc", false, "ffmpeg config: HlsEnc")
	cmd.Flags().StringVar(&hls.HlsEncKey, "hls_enc_key", "", "ffmpeg config: HlsEncKey")
	cmd.Flags().StringVar(&hls.HlsEncKeyUrl, "hls_enc_key_url", "", "ffmpeg config: HlsEncKeyUrl")
	cmd.Flags().StringVar(&hls.HlsEncIv, "hls_enc_iv", "", "ffmpeg config: HlsEncIv")
}
