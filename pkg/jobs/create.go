package jobs

import (
	"encoding/json"
	"fmt"

	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/cli/pkg/formatter"
	"github.com/chunkifydev/cli/pkg/sources"
	"github.com/spf13/cobra"
)

// CreateCmd represents the command for creating a new job
type CreateCmd struct {
	Params chunkify.JobCreateParams `json:"-"`

	vcpu         int64        `json:"-"` // Number of vCPUs to use
	metadata     string       `json:"-"` // metadata for the job
	interactive  bool         `json:"-"` // Whether to run in interactive mode
	sourceUrl    string       `json:"-"` // URL of the source media
	Data         chunkify.Job `json:"-"` // Response data from job creation
	videoBitrate string       // Video bitrate setting
	audioBitrate string       // Audio bitrate setting
}

// Valid vCPU values that can be used for transcoding
var validCpus = map[int64]bool{4: true, 8: true, 16: true}

// Execute creates a new job with the configured parameters
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
		// Validate vCPU value
		if !validCpus[r.vcpu] {
			return fmt.Errorf("invalid cpu value: %d. Allowed values are 4, 8, 16", r.vcpu)
		}
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

// View displays the job list, either once or continuously in interactive mode
func (r *CreateCmd) View() {
	jobList := &ListCmd{Params: chunkify.JobListParams{CreatedSort: "asc", SourceId: r.Params.SourceId}, interactive: r.interactive}
	jobList.Execute()
	if r.interactive {
		StartPolling(jobList)
	} else {
		jobList.View()
	}
}

// newCreateCmd creates and returns a new cobra command for job creation
func newCreateCmd() *cobra.Command {
	req := &CreateCmd{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new job",
		Long:  `Create a new job`,
	}

	// Add common flags that will be shared across all subcommands
	cmd.PersistentFlags().StringVar(&req.Params.SourceId, "source-id", "", "The source id (required)")
	cmd.PersistentFlags().StringVar(&req.metadata, "metadata", "", "Optional metadata in JSON format")
	cmd.PersistentFlags().Int64Var(&req.Params.Transcoder.Quantity, "transcoder", 0, "Number of transcoders: 1 to 50 (required if cpu is set)")
	cmd.PersistentFlags().Int64Var(&req.vcpu, "cpu", 0, "Instance vCPU: 4, 8, 16 (required if transcoder is set)")
	cmd.PersistentFlags().StringVar(&req.Params.Storage.Id, "storage", "", "The storage id (default is the project default storage id)")
	cmd.PersistentFlags().StringVar(&req.Params.Storage.Path, "path", "", "The destination path on your storage")
	cmd.PersistentFlags().BoolVarP(&req.interactive, "interactive", "i", false, "Refresh the list in real time")
	cmd.PersistentFlags().StringVar(&req.sourceUrl, "url", "", "Create the job with the given source url")
	cmd.PersistentFlags().StringVar(&req.videoBitrate, "vb", "", "ffmpeg config: VideoBitrate")
	cmd.PersistentFlags().StringVar(&req.audioBitrate, "ab", "", "ffmpeg config: AudioBitrate")

	// Add subcommands for different encoding formats
	cmd.AddCommand(newMp4H264Cmd(req))
	cmd.AddCommand(newMp4H265Cmd(req))
	cmd.AddCommand(newMp4Av1Cmd(req))
	cmd.AddCommand(newWebmVp9Cmd(req))
	cmd.AddCommand(newHlsH264Cmd(req))
	cmd.AddCommand(newHlsH265Cmd(req))
	cmd.AddCommand(newHlsAv1Cmd(req))
	cmd.AddCommand(newJpgCmd(req))

	// Set flag requirements and exclusions
	cmd.MarkFlagsRequiredTogether("transcoder", "cpu")
	cmd.MarkFlagsMutuallyExclusive("source-id", "url")

	return cmd
}

// newMp4H264Cmd creates a new command for x264 encoding
func newMp4H264Cmd(req *CreateCmd) *cobra.Command {
	// Create x264 config
	h264Config := &chunkify.H264{}

	cmd := &cobra.Command{
		Use:   string(chunkify.FormatMp4H264),
		Short: "Create a job with x264 encoding",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse bitrates
			setBitrate(req.videoBitrate, req.audioBitrate, &h264Config.Video)

			// Set the config
			req.Params.Format.Mp4H264 = h264Config
			// Execute the job creation
			if err := req.Execute(); err != nil {
				return err
			}

			req.View()
			return nil
		},
	}

	// Add h264 specific flags
	setH264Flags(cmd, h264Config)

	// Add common video flags
	setCommonVideoFlags(cmd, &h264Config.Video)

	return cmd
}

// newMp4H265Cmd creates a new command for x265 encoding
func newMp4H265Cmd(req *CreateCmd) *cobra.Command {
	// Create x265 config
	h265Config := &chunkify.H265{}

	cmd := &cobra.Command{
		Use:   string(chunkify.FormatMp4H265),
		Short: "Create a job with x265 encoding",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse bitrates
			setBitrate(req.videoBitrate, req.audioBitrate, &h265Config.Video)

			// Set the config
			req.Params.Format.Mp4H265 = h265Config
			// Execute the job creation
			if err := req.Execute(); err != nil {
				return err
			}

			req.View()
			return nil
		},
	}

	// Add x264 specific flags
	setH265Flags(cmd, h265Config)

	// Add common video flags
	setCommonVideoFlags(cmd, &h265Config.Video)

	return cmd
}

// newMp4Av1Cmd creates a new command for AV1 encoding
func newMp4Av1Cmd(req *CreateCmd) *cobra.Command {
	// Create x265 config
	av1Config := &chunkify.Av1{}

	cmd := &cobra.Command{
		Use:   string(chunkify.FormatMp4Av1),
		Short: "Create a job with av1 encoding",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse bitrates
			setBitrate(req.videoBitrate, req.audioBitrate, &av1Config.Video)

			// Set the config
			req.Params.Format.Mp4Av1 = av1Config
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
	setCommonVideoFlags(cmd, &av1Config.Video)

	return cmd
}

// newWebmVp9Cmd creates a new command for VP9 encoding
func newWebmVp9Cmd(req *CreateCmd) *cobra.Command {
	// Create vp9 config
	vp9Config := &chunkify.Vp9{}

	cmd := &cobra.Command{
		Use:   string(chunkify.FormatWebmVp9),
		Short: "Create a job with vp9 encoding",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse bitrates
			setBitrate(req.videoBitrate, req.audioBitrate, &vp9Config.Video)

			// Set the config
			req.Params.Format.WebmVp9 = vp9Config
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
	setCommonVideoFlags(cmd, &vp9Config.Video)

	return cmd
}

// newHlsH264Cmd creates a new command for HLS with x264 encoding
func newHlsH264Cmd(req *CreateCmd) *cobra.Command {
	// Create x265 config
	hlsH264Config := &chunkify.HlsH264{}

	cmd := &cobra.Command{
		Use:   string(chunkify.FormatHlsH264),
		Short: "Create a job with hls x264 encoding",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse bitrates
			setBitrate(req.videoBitrate, req.audioBitrate, &hlsH264Config.Video)

			// Set the config
			req.Params.Format.HlsH264 = hlsH264Config
			// Execute the job creation
			if err := req.Execute(); err != nil {
				return err
			}

			req.View()
			return nil
		},
	}

	// Add HLS specific flags
	setHLSFlags(cmd, &hlsH264Config.Hls)

	// Add h264 specific flags
	setH264Flags(cmd, &hlsH264Config.H264)

	// Add common video flags
	setCommonVideoFlags(cmd, &hlsH264Config.Video)

	return cmd
}

// newHlsH265Cmd creates a new command for HLS with x265 encoding
func newHlsH265Cmd(req *CreateCmd) *cobra.Command {
	// Create h265 config
	hlsH265Config := &chunkify.HlsH265{}

	cmd := &cobra.Command{
		Use:   string(chunkify.FormatHlsH265),
		Short: "Create a job with hls x265 encoding",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse bitrates
			setBitrate(req.videoBitrate, req.audioBitrate, &hlsH265Config.Video)

			// Set the config
			req.Params.Format.HlsH265 = hlsH265Config
			// Execute the job creation
			if err := req.Execute(); err != nil {
				return err
			}

			req.View()
			return nil
		},
	}

	// Add HLS specific flags
	setHLSFlags(cmd, &hlsH265Config.Hls)

	// Add x265 specific flags
	setH265Flags(cmd, &hlsH265Config.H265)

	// Add common video flags
	setCommonVideoFlags(cmd, &hlsH265Config.Video)

	return cmd
}

// newHlsAv1Cmd creates a new command for HLS with AV1 encoding
func newHlsAv1Cmd(req *CreateCmd) *cobra.Command {
	// Create x265 config
	hlsAv1Config := &chunkify.HlsAv1{}

	cmd := &cobra.Command{
		Use:   string(chunkify.FormatHlsAv1),
		Short: "Create a job with hls av1 encoding",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse bitrates
			setBitrate(req.videoBitrate, req.audioBitrate, &hlsAv1Config.Video)

			// Set the config
			req.Params.Format.HlsAv1 = hlsAv1Config
			// Execute the job creation
			if err := req.Execute(); err != nil {
				return err
			}

			req.View()
			return nil
		},
	}

	// Add HLS specific flags
	setHLSFlags(cmd, &hlsAv1Config.Hls)

	// Add av1 specific flags
	setAv1Flags(cmd, &hlsAv1Config.Av1)

	// Add common video flags
	setCommonVideoFlags(cmd, &hlsAv1Config.Video)

	return cmd
}

// newJpgCmd creates a new command for JPG encoding
func newJpgCmd(req *CreateCmd) *cobra.Command {
	// Create x265 config
	jpgConfig := &chunkify.Jpg{}

	cmd := &cobra.Command{
		Use:   string(chunkify.FormatJpg),
		Short: "Create a job with jpg encoding",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Set the config
			req.Params.Format.Jpg = jpgConfig
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
	cmd.Flags().Int64Var(&jpgConfig.Frames, "frames", 0, "Generate only the number of given frames")
	cmd.Flags().Int64Var(&jpgConfig.Width, "width", 0, "ffmpeg config: Width")
	cmd.Flags().Int64Var(&jpgConfig.Height, "height", 0, "ffmpeg config: Height")
	cmd.Flags().Int64Var(&jpgConfig.Interval, "interval", 0, "ffmpeg config: Interval")
	return cmd
}

// setCommonVideoFlags adds common video-related flags to the command
func setCommonVideoFlags(cmd *cobra.Command, videoCommon *chunkify.Video) {
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
	cmd.Flags().Int64Var(&videoCommon.Duration, "duration", 0, "ffmpeg config: Duration")
	cmd.Flags().Int64Var(&videoCommon.Seek, "seek", 0, "ffmpeg config: Seek")
}

// setBitrate parses and sets video and audio bitrates
func setBitrate(vb, ab string, videoCommon *chunkify.Video) {
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

// setH264Flags adds x264-specific encoding flags to the command
func setH264Flags(cmd *cobra.Command, h264 *chunkify.H264) {
	cmd.Flags().Int64Var(&h264.X264KeyInt, "x264_keyint", 0, "ffmpeg config: X264KeyInt")
	cmd.Flags().Int64Var(&h264.Level, "level", 0, "ffmpeg config: Level")
	cmd.Flags().StringVar(&h264.Profilev, "profilev", "", "ffmpeg config: Profilev")
	cmd.Flags().Int64Var(&h264.Crf, "crf", 0, "ffmpeg config: Crf")
	cmd.Flags().StringVar(&h264.Preset, "preset", "", "ffmpeg config: Preset")
}

// setH265Flags adds x265-specific encoding flags to the command
func setH265Flags(cmd *cobra.Command, x265 *chunkify.H265) {
	cmd.Flags().Int64Var(&x265.X265KeyInt, "x265_keyint", 0, "ffmpeg config: X265KeyInt")
	cmd.Flags().Int64Var(&x265.Level, "level", 0, "ffmpeg config: Level")
	cmd.Flags().StringVar(&x265.Profilev, "profilev", "", "ffmpeg config: Profilev")
	cmd.Flags().Int64Var(&x265.Crf, "crf", 0, "ffmpeg config: Crf")
	cmd.Flags().StringVar(&x265.Preset, "preset", "", "ffmpeg config: Preset")
}

// setVp9Flags adds VP9-specific encoding flags to the command
func setVp9Flags(cmd *cobra.Command, vp9 *chunkify.Vp9) {
	cmd.Flags().Int64Var(&vp9.Crf, "crf", 0, "ffmpeg config: Crf")
	cmd.Flags().StringVar(&vp9.CpuUsed, "cpu_used", "", "ffmpeg config: CpuUsed")
	cmd.Flags().StringVar(&vp9.Quality, "quality", "", "ffmpeg config: Quality")
}

// setAv1Flags adds AV1-specific encoding flags to the command
func setAv1Flags(cmd *cobra.Command, av1 *chunkify.Av1) {
	cmd.Flags().Int64Var(&av1.Level, "level", 0, "ffmpeg config: Level")
	cmd.Flags().StringVar(&av1.Profilev, "profilev", "", "ffmpeg config: Profilev")
	cmd.Flags().Int64Var(&av1.Crf, "crf", 0, "ffmpeg config: Crf")
	cmd.Flags().StringVar(&av1.Preset, "preset", "", "ffmpeg config: Preset")
}

// setHLSFlags adds HLS-specific encoding flags to the command
func setHLSFlags(cmd *cobra.Command, hls *chunkify.Hls) {
	cmd.Flags().Int64Var(&hls.HlsTime, "hls_time", 0, "ffmpeg config: HlsTime")
	cmd.Flags().StringVar(&hls.HlsSegmentType, "hls_segment_type", "", "ffmpeg config: HlsSegmentType")
	cmd.Flags().BoolVar(&hls.HlsEnc, "hls_enc", false, "ffmpeg config: HlsEnc")
	cmd.Flags().StringVar(&hls.HlsEncKey, "hls_enc_key", "", "ffmpeg config: HlsEncKey")
	cmd.Flags().StringVar(&hls.HlsEncKeyUrl, "hls_enc_key_url", "", "ffmpeg config: HlsEncKeyUrl")
	cmd.Flags().StringVar(&hls.HlsEncIv, "hls_enc_iv", "", "ffmpeg config: HlsEncIv")
}
