package chunkify

import (
	"fmt"
	"path"
	"slices"
	"strconv"
	"strings"

	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/cli/pkg/formatter"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

const (
	FormatMp4H264 = "mp4_h264"
	FormatMp4H265 = "mp4_h265"
	FormatMp4Av1  = "mp4_av1"
	FormatWebmVp9 = "webm_vp9"
	FormatHlsH264 = "hls_h264"
	FormatHlsH265 = "hls_h265"
	FormatHlsAv1  = "hls_av1"
	FormatJpg     = "jpg"
)

// Transcoder flags
var (
	transcoders    = new(int64)
	transcoderVcpu = new(int64)
)

// Storage flags
var (
	storagePath = new(string)
)

// Common video flags
var (
	resolution   = new(string)
	width        = new(int64)
	height       = new(int64)
	framerate    = new(float64)
	gop          = new(int64)
	channels     = new(int64)
	pixfmt       = new(string)
	disableAudio = new(bool)
	disableVideo = new(bool)
	duration     = new(int64)
	seek         = new(int64)

	maxrate      = new(int64)
	bufsize      = new(int64)
	videoBitrate = new(int64)
	audioBitrate = new(int64)

	maxrateStr      = new(string)
	bufsizeStr      = new(string)
	videoBitrateStr = new(string)
	audioBitrateStr = new(string)
)

// H264, H265 and AV1 flags
var (
	crf        = new(int64)
	preset     = new(string)
	profilev   = new(string)
	level      = new(int64)
	x264KeyInt = new(int64)
	x265KeyInt = new(int64)
)

// VP9 flags
var (
	quality = new(string)
	cpuUsed = new(string)
)

// HLS flags
var (
	hlsManifestId  = new(string)
	hlsTime        = new(int64)
	hlsSegmentType = new(string)
	hlsEnc         = new(bool)
	hlsEncKey      = new(string)
	hlsEncKeyUrl   = new(string)
	hlsEncIv       = new(string)
)

// JPG
var (
	interval = new(int64)
	sprite   = new(bool)
)

// BindFlags attaches root-level flags used by the root command
func BindFlags(app *App, cmd *cobra.Command) {
	app.Command = &ChunkifyCommand{Id: uuid.New().String()}

	cmd.Flags().BoolVar(&app.JSON, "json", false, "Output in JSON format")
	cmd.Flags().StringVarP(&app.Command.Input, "input", "i", "", "Input video to transcode. It can be a file, HTTP URL or source ID (src_*)")
	cmd.Flags().StringVarP(&app.Command.Output, "output", "o", "", "Output file path")
	cmd.Flags().StringVarP(&app.Command.Format, "format", "f", "", "Output format (mp4/h264, mp4/h265, mp4/av1, webm/vp9, hls/h264, hls/h265, hls/av1, jpg)")

	cmd.Flags().Int64Var(transcoders, "transcoders", 0, "Number of transcoders to use")
	cmd.Flags().Int64Var(transcoderVcpu, "vcpu", 0, "vCPU per transcoder (4, 8, or 16)")

	cmd.Flags().StringVar(storagePath, "storage-path", "", "Storage absolute path")

	// Common video settings
	cmd.Flags().StringVarP(resolution, "resolution", "s", "", "Set resolution wxh (0-8192x0-8192)")
	cmd.Flags().Float64VarP(framerate, "framerate", "r", 0, "Set frame rate (15-120)")
	cmd.Flags().Int64VarP(gop, "gop", "g", 0, "Set group of pictures size (1-300)")
	cmd.Flags().Int64Var(channels, "channels", 0, "Set number of audio channels (1, 2, 5, 7)")
	cmd.Flags().StringVar(maxrateStr, "maxrate", "", "Set maximum bitrate in bits per second (100000-50000000). You can use units like K, M (e.g. 1200K, 2M)")
	cmd.Flags().StringVar(bufsizeStr, "bufsize", "", "Set buffer size in bits (100000-50000000). You can use units like K, M (e.g. 1200K, 2M)")
	cmd.Flags().StringVar(pixfmt, "pixfmt", "", "Set pixel format (yuv410p, yuv411p, yuv420p, yuv422p, yuv440p, yuv444p, yuvJ411p, yuvJ420p, yuvJ422p, yuvJ440p, yuvJ444p, yuv420p10le, yuv422p10le, yuv440p10le, yuv444p10le, yuv420p12le, yuv422p12le, yuv440p12le, yuv444p12le, yuv420p10be, yuv422p10be, yuv440p10be, yuv444p10be, yuv420p12be, yuv422p12be, yuv440p12be, yuv444p12be)")
	cmd.Flags().BoolVar(disableAudio, "an", false, "Disable audio")
	cmd.Flags().BoolVar(disableVideo, "vn", false, "Disable video")
	cmd.Flags().Int64VarP(duration, "duration", "t", 0, "Set duration in seconds")
	cmd.Flags().Int64Var(seek, "seek", 0, "Seek to position in seconds")
	cmd.Flags().StringVar(videoBitrateStr, "vb", "", "Set video bitrate in bits per second (100000-50000000). You can use units like K, M (e.g. 1200K, 2M)")
	cmd.Flags().StringVar(audioBitrateStr, "ab", "", "Set audio bitrate in bits per second (32000-512000). You can use units like K, M (e.g. 1200K, 2M)")

	// H264, H265 and AV1 flags
	cmd.Flags().Int64Var(crf, "crf", 0, "Set constant rate factor (H264/H265: 16-35, AV1: 16-63, VP9: 15-35)")
	cmd.Flags().StringVar(preset, "preset", "", "Set encoding preset (H264/H265: ultrafast, superfast, veryfast, faster, fast, medium, AV1: 6-13)")
	cmd.Flags().StringVar(profilev, "profilev", "", "Set video profile (H264: baseline, main, high, high10, high422, high444, H265/AV1: main, main10, mainstillpicture)")
	cmd.Flags().Int64Var(level, "level", 0, "Set encoding level (H264: 10, 11, 12, 13, 20, 21, 22, 30, 31, 32, 40, 41, 42, 50, 51, H265: 30, 31, 41, AV1: 30, 31, 41)")
	cmd.Flags().Int64Var(x264KeyInt, "x264keyint", 0, "H264 - Set x264 keyframe interval")
	cmd.Flags().Int64Var(x265KeyInt, "x265keyint", 0, "H265 - Set x265 keyframe interval")

	// VP9 flags
	cmd.Flags().StringVar(quality, "quality", "", "Set VP9 quality (good, best, realtime)")
	cmd.Flags().StringVar(cpuUsed, "cpu-used", "", "Set VP9 CPU usage (0-8)")

	// HLS flags
	cmd.Flags().StringVar(hlsManifestId, "hls-manifest-id", "", "Set HLS manifest ID")
	cmd.Flags().Int64Var(hlsTime, "hls-time", 0, "Set HLS segment duration in seconds (1-10)")
	cmd.Flags().StringVar(hlsSegmentType, "hls-segment-type", "", "Set HLS segment type (mpegts, fmp4)")
	cmd.Flags().BoolVar(hlsEnc, "hls-enc", false, "Enable HLS encryption")
	cmd.Flags().StringVar(hlsEncKey, "hls-enc-key", "", "Set HLS encryption key")
	cmd.Flags().StringVar(hlsEncKeyUrl, "hls-enc-key-url", "", "Set HLS encryption key URL")
	cmd.Flags().StringVar(hlsEncIv, "hls-enc-iv", "", "Set HLS encryption IV")

	// JPG flags
	cmd.Flags().Int64Var(interval, "interval", 0, "Set frame extraction interval in seconds (1-60)")
	cmd.Flags().BoolVar(sprite, "sprite", false, "Generate sprite sheet")

	cmd.MarkFlagRequired("input")

	cmd.MarkFlagsRequiredTogether("transcoders", "vcpu")

	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if err := setupCommand(app); err != nil {
			return err
		}

		if err := validateTranscodeSettings(app); err != nil {
			return err
		}

		// build job format params according to all format flags
		setJobFormatParams(app)

		return nil
	}
}

func setupCommand(app *App) error {
	// If no format or output is specified, we don't need to validate anything
	if app.Command.Format == "" && app.Command.Output == "" {
		return nil
	}

	// Set default format based on output file extension
	if app.Command.Format == "" {
		switch path.Ext(app.Command.Output) {
		case ".mp4":
			app.Command.Format = FormatMp4H264
		case ".webm":
			app.Command.Format = FormatWebmVp9
		case ".m3u8":
			app.Command.Format = FormatHlsH264
		case ".jpg":
			app.Command.Format = FormatJpg
		default:
			return fmt.Errorf("invalid output file extension: %s. Please provide a valid format with --format", path.Ext(app.Command.Output))
		}
	}

	// Set the number of transcoders and their type
	if transcoders != nil && *transcoders > 0 {
		app.Command.JobTranscoderParams = chunkify.JobNewParamsTranscoder{
			Quantity: chunkify.Int(*transcoders),
			Type:     fmt.Sprintf("%dvCPU", *transcoderVcpu),
		}
	}

	// Set the storage path
	if storagePath != nil && *storagePath != "" {
		app.Command.JobCreateStorageParams = chunkify.JobNewParamsStorage{
			Path: chunkify.String(*storagePath),
		}
	}

	// shortcut to set width and height from resolution flag
	if resolution != nil && *resolution != "" {
		parts := strings.Split(*resolution, "x")
		if len(parts) != 2 {
			return fmt.Errorf("invalid resolution: %s", *resolution)
		}

		resWidth, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid width: %s", parts[0])
		}
		resHeight, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid height: %s", parts[1])
		}
		width = &resWidth
		height = &resHeight
		resolution = nil
	}

	// Convert bitrate like "1200K", "2M" to bits (int64)
	// for convenience
	if videoBitrateStr != nil && *videoBitrateStr != "" {
		videoBitrateInt, err := formatter.ParseFileSize(*videoBitrateStr)
		if err != nil {
			return fmt.Errorf("invalid video bitrate: %s", *videoBitrateStr)
		}
		videoBitrate = &videoBitrateInt
	}

	if audioBitrateStr != nil && *audioBitrateStr != "" {
		audioBitrateInt, err := formatter.ParseFileSize(*audioBitrateStr)
		if err != nil {
			return fmt.Errorf("invalid audio bitrate: %s", *audioBitrateStr)
		}
		audioBitrate = &audioBitrateInt
	}

	if maxrateStr != nil && *maxrateStr != "" {
		maxrateInt, err := formatter.ParseFileSize(*maxrateStr)
		if err != nil {
			return fmt.Errorf("invalid maximum bitrate: %s", *maxrateStr)
		}
		maxrate = &maxrateInt
	}

	if bufsizeStr != nil && *bufsizeStr != "" {
		bufsizeInt, err := formatter.ParseFileSize(*bufsizeStr)
		if err != nil {
			return fmt.Errorf("invalid buffer size: %s", *bufsizeStr)
		}
		bufsize = &bufsizeInt
	}

	return nil
}

func validateTranscodeSettings(app *App) error {
	// don't validate format and format settings if no format or output is specified
	// it will just upload the file and return the source ID
	if app.Command.Format == "" && app.Command.Output == "" {
		return nil
	}

	// Check if the format is valid
	if !slices.Contains([]string{
		FormatMp4H264,
		FormatMp4H265,
		FormatMp4Av1,
		FormatWebmVp9,
		FormatHlsH264,
		FormatHlsH265,
		FormatHlsAv1,
		FormatJpg,
	}, app.Command.Format) {
		return fmt.Errorf("invalid format: %s", app.Command.Format)
	}

	if err := validateCommonVideoFlags(); err != nil {
		return err
	}

	// validate format settings according to the format
	switch app.Command.Format {
	case FormatMp4H264:
		if err := validateH264Flags(); err != nil {
			return err
		}

	case FormatMp4H265:
		if err := validateH265Flags(); err != nil {
			return err
		}

	case FormatMp4Av1:
		if err := validateAv1Flags(); err != nil {
			return err
		}

	case FormatWebmVp9:
		if err := validateWebmVp9Flags(); err != nil {
			return err
		}

	case FormatJpg:
		if err := validateJpgFlags(); err != nil {
			return err
		}

	case FormatHlsH264:
		if err := validateHlsH264Flags(); err != nil {
			return err
		}

	case FormatHlsH265:
		if err := validateHlsH265Flags(); err != nil {
			return err
		}

	case FormatHlsAv1:
		if err := validateHlsAv1Flags(); err != nil {
			return err
		}

	}

	return nil
}

func setJobFormatParams(app *App) {
	app.Command.JobFormatParams = chunkify.JobNewParamsFormatUnion{}

	switch app.Command.Format {
	case string(FormatMp4H264):
		h264Params := &chunkify.MP4H264Param{ID: FormatMp4H264}

		if width != nil && *width > 0 {
			h264Params.Width = chunkify.Int(*width)
		}
		if height != nil && *height > 0 {
			h264Params.Height = chunkify.Int(*height)
		}
		if framerate != nil && *framerate > 0 {
			h264Params.Framerate = chunkify.Float(*framerate)
		}
		if gop != nil && *gop > 0 {
			h264Params.Gop = chunkify.Int(*gop)
		}
		if channels != nil && *channels > 0 {
			h264Params.Channels = *channels
		}
		if maxrate != nil && *maxrate > 0 {
			h264Params.Maxrate = chunkify.Int(*maxrate)
		}
		if bufsize != nil && *bufsize > 0 {
			h264Params.Bufsize = chunkify.Int(*bufsize)
		}
		if pixfmt != nil && *pixfmt != "" {
			h264Params.Pixfmt = chunkify.MP4H264Pixfmt(*pixfmt)
		}
		if disableAudio != nil && *disableAudio {
			h264Params.DisableAudio = chunkify.Bool(true)
		}
		if disableVideo != nil && *disableVideo {
			h264Params.DisableVideo = chunkify.Bool(true)
		}
		if duration != nil && *duration > 0 {
			h264Params.Duration = chunkify.Int(*duration)
		}
		if seek != nil && *seek > 0 {
			h264Params.Seek = chunkify.Int(*seek)
		}
		if videoBitrate != nil && *videoBitrate > 0 {
			h264Params.VideoBitrate = chunkify.Int(*videoBitrate)
		}
		if audioBitrate != nil && *audioBitrate > 0 {
			h264Params.AudioBitrate = chunkify.Int(*audioBitrate)
		}
		if crf != nil && *crf > 0 {
			h264Params.Crf = chunkify.Int(*crf)
		}
		if preset != nil && *preset != "" {
			h264Params.Preset = chunkify.MP4H264Preset(*preset)
		}
		if profilev != nil && *profilev != "" {
			h264Params.Profilev = chunkify.MP4H264Profilev(*profilev)
		}
		if level != nil && *level > 0 {
			h264Params.Level = *level
		}
		if x264KeyInt != nil && *x264KeyInt > 0 {
			h264Params.X264Keyint = chunkify.Int(*x264KeyInt)
		}

		app.Command.JobFormatParams.OfMP4H264 = h264Params
	case string(FormatMp4H265):
		h265Params := &chunkify.MP4H265Param{ID: FormatMp4H265}

		if width != nil && *width > 0 {
			h265Params.Width = chunkify.Int(*width)
		}
		if height != nil && *height > 0 {
			h265Params.Height = chunkify.Int(*height)
		}
		if framerate != nil && *framerate > 0 {
			h265Params.Framerate = chunkify.Float(*framerate)
		}
		if gop != nil && *gop > 0 {
			h265Params.Gop = chunkify.Int(*gop)
		}
		if channels != nil && *channels > 0 {
			h265Params.Channels = *channels
		}
		if maxrate != nil && *maxrate > 0 {
			h265Params.Maxrate = chunkify.Int(*maxrate)
		}
		if bufsize != nil && *bufsize > 0 {
			h265Params.Bufsize = chunkify.Int(*bufsize)
		}
		if pixfmt != nil && *pixfmt != "" {
			h265Params.Pixfmt = chunkify.MP4H265Pixfmt(*pixfmt)
		}
		if disableAudio != nil && *disableAudio {
			h265Params.DisableAudio = chunkify.Bool(true)
		}
		if disableVideo != nil && *disableVideo {
			h265Params.DisableVideo = chunkify.Bool(true)
		}
		if duration != nil && *duration > 0 {
			h265Params.Duration = chunkify.Int(*duration)
		}
		if seek != nil && *seek > 0 {
			h265Params.Seek = chunkify.Int(*seek)
		}
		if videoBitrate != nil && *videoBitrate > 0 {
			h265Params.VideoBitrate = chunkify.Int(*videoBitrate)
		}
		if audioBitrate != nil && *audioBitrate > 0 {
			h265Params.AudioBitrate = chunkify.Int(*audioBitrate)
		}
		if crf != nil && *crf > 0 {
			h265Params.Crf = chunkify.Int(*crf)
		}
		if preset != nil && *preset != "" {
			h265Params.Preset = chunkify.MP4H265Preset(*preset)
		}
		if profilev != nil && *profilev != "" {
			h265Params.Profilev = chunkify.MP4H265Profilev(*profilev)
		}
		if level != nil && *level > 0 {
			h265Params.Level = *level
		}
		if x265KeyInt != nil && *x265KeyInt > 0 {
			h265Params.X265Keyint = chunkify.Int(*x265KeyInt)
		}
		app.Command.JobFormatParams.OfMP4H265 = h265Params
	case string(FormatWebmVp9):
		vp9Params := &chunkify.WebmVp9Param{ID: FormatWebmVp9}

		if width != nil && *width > 0 {
			vp9Params.Width = chunkify.Int(*width)
		}
		if height != nil && *height > 0 {
			vp9Params.Height = chunkify.Int(*height)
		}
		if framerate != nil && *framerate > 0 {
			vp9Params.Framerate = chunkify.Float(*framerate)
		}
		if gop != nil && *gop > 0 {
			vp9Params.Gop = chunkify.Int(*gop)
		}
		if channels != nil && *channels > 0 {
			vp9Params.Channels = *channels
		}
		if maxrate != nil && *maxrate > 0 {
			vp9Params.Maxrate = chunkify.Int(*maxrate)
		}
		if bufsize != nil && *bufsize > 0 {
			vp9Params.Bufsize = chunkify.Int(*bufsize)
		}
		if pixfmt != nil && *pixfmt != "" {
			vp9Params.Pixfmt = chunkify.WebmVp9Pixfmt(*pixfmt)
		}
		if disableAudio != nil && *disableAudio {
			vp9Params.DisableAudio = chunkify.Bool(true)
		}
		if disableVideo != nil && *disableVideo {
			vp9Params.DisableVideo = chunkify.Bool(true)
		}
		if duration != nil && *duration > 0 {
			vp9Params.Duration = chunkify.Int(*duration)
		}
		if seek != nil && *seek > 0 {
			vp9Params.Seek = chunkify.Int(*seek)
		}
		if videoBitrate != nil && *videoBitrate > 0 {
			vp9Params.VideoBitrate = chunkify.Int(*videoBitrate)
		}
		if audioBitrate != nil && *audioBitrate > 0 {
			vp9Params.AudioBitrate = chunkify.Int(*audioBitrate)
		}

		if crf != nil && *crf > 0 {
			vp9Params.Crf = chunkify.Int(*crf)
		}
		if quality != nil && *quality != "" {
			vp9Params.Quality = chunkify.WebmVp9Quality(*quality)
		}
		if cpuUsed != nil && *cpuUsed != "" {
			vp9Params.CPUUsed = chunkify.WebmVp9CPUUsed(*cpuUsed)
		}
		app.Command.JobFormatParams.OfWebmVp9 = vp9Params
	case string(FormatMp4Av1):
		av1Params := &chunkify.MP4Av1Param{ID: FormatMp4Av1}

		if width != nil && *width > 0 {
			av1Params.Width = chunkify.Int(*width)
		}
		if height != nil && *height > 0 {
			av1Params.Height = chunkify.Int(*height)
		}
		if framerate != nil && *framerate > 0 {
			av1Params.Framerate = chunkify.Float(*framerate)
		}
		if gop != nil && *gop > 0 {
			av1Params.Gop = chunkify.Int(*gop)
		}
		if channels != nil && *channels > 0 {
			av1Params.Channels = *channels
		}
		if maxrate != nil && *maxrate > 0 {
			av1Params.Maxrate = chunkify.Int(*maxrate)
		}
		if bufsize != nil && *bufsize > 0 {
			av1Params.Bufsize = chunkify.Int(*bufsize)
		}
		if pixfmt != nil && *pixfmt != "" {
			av1Params.Pixfmt = chunkify.MP4Av1Pixfmt(*pixfmt)
		}
		if disableAudio != nil && *disableAudio {
			av1Params.DisableAudio = chunkify.Bool(true)
		}
		if disableVideo != nil && *disableVideo {
			av1Params.DisableVideo = chunkify.Bool(true)
		}
		if duration != nil && *duration > 0 {
			av1Params.Duration = chunkify.Int(*duration)
		}
		if seek != nil && *seek > 0 {
			av1Params.Seek = chunkify.Int(*seek)
		}
		if videoBitrate != nil && *videoBitrate > 0 {
			av1Params.VideoBitrate = chunkify.Int(*videoBitrate)
		}
		if audioBitrate != nil && *audioBitrate > 0 {
			av1Params.AudioBitrate = chunkify.Int(*audioBitrate)
		}

		if crf != nil && *crf > 0 {
			av1Params.Crf = chunkify.Int(*crf)
		}
		if preset != nil && *preset != "" {
			av1Params.Preset = chunkify.MP4Av1Preset(*preset)
		}
		if profilev != nil && *profilev != "" {
			av1Params.Profilev = chunkify.MP4Av1Profilev(*profilev)
		}
		if level != nil && *level > 0 {
			av1Params.Level = *level
		}

		app.Command.JobFormatParams.OfMP4Av1 = av1Params
	case string(FormatHlsH264):
		hlsH264Params := &chunkify.HlsH264Param{ID: FormatHlsH264}

		if width != nil && *width > 0 {
			hlsH264Params.Width = chunkify.Int(*width)
		}
		if height != nil && *height > 0 {
			hlsH264Params.Height = chunkify.Int(*height)
		}
		if framerate != nil && *framerate > 0 {
			hlsH264Params.Framerate = chunkify.Float(*framerate)
		}
		if gop != nil && *gop > 0 {
			hlsH264Params.Gop = chunkify.Int(*gop)
		}
		if channels != nil && *channels > 0 {
			hlsH264Params.Channels = *channels
		}
		if maxrate != nil && *maxrate > 0 {
			hlsH264Params.Maxrate = chunkify.Int(*maxrate)
		}
		if bufsize != nil && *bufsize > 0 {
			hlsH264Params.Bufsize = chunkify.Int(*bufsize)
		}
		if pixfmt != nil && *pixfmt != "" {
			hlsH264Params.Pixfmt = chunkify.HlsH264Pixfmt(*pixfmt)
		}
		if disableAudio != nil && *disableAudio {
			hlsH264Params.DisableAudio = chunkify.Bool(true)
		}
		if disableVideo != nil && *disableVideo {
			hlsH264Params.DisableVideo = chunkify.Bool(true)
		}
		if duration != nil && *duration > 0 {
			hlsH264Params.Duration = chunkify.Int(*duration)
		}
		if seek != nil && *seek > 0 {
			hlsH264Params.Seek = chunkify.Int(*seek)
		}
		if videoBitrate != nil && *videoBitrate > 0 {
			hlsH264Params.VideoBitrate = chunkify.Int(*videoBitrate)
		}
		if audioBitrate != nil && *audioBitrate > 0 {
			hlsH264Params.AudioBitrate = chunkify.Int(*audioBitrate)
		}

		if hlsTime != nil && *hlsTime > 0 {
			hlsH264Params.HlsTime = chunkify.Int(*hlsTime)
		}
		if hlsSegmentType != nil && *hlsSegmentType != "" {
			hlsH264Params.HlsSegmentType = chunkify.HlsH264HlsSegmentType(*hlsSegmentType)
		}
		if hlsEnc != nil && *hlsEnc {
			hlsH264Params.HlsEnc = chunkify.Bool(true)
		}
		if hlsEncKey != nil && *hlsEncKey != "" {
			hlsH264Params.HlsEncKey = chunkify.String(*hlsEncKey)
		}
		if hlsEncKeyUrl != nil && *hlsEncKeyUrl != "" {
			hlsH264Params.HlsEncKeyURL = chunkify.String(*hlsEncKeyUrl)
		}
		if hlsEncIv != nil && *hlsEncIv != "" {
			hlsH264Params.HlsEncIv = chunkify.String(*hlsEncIv)
		}
		if crf != nil && *crf > 0 {
			hlsH264Params.Crf = chunkify.Int(*crf)
		}
		if preset != nil && *preset != "" {
			hlsH264Params.Preset = chunkify.HlsH264Preset(*preset)
		}
		if profilev != nil && *profilev != "" {
			hlsH264Params.Profilev = chunkify.HlsH264Profilev(*profilev)
		}
		if level != nil && *level > 0 {
			hlsH264Params.Level = *level
		}
		if x264KeyInt != nil && *x264KeyInt > 0 {
			hlsH264Params.X264Keyint = chunkify.Int(*x264KeyInt)
		}

		app.Command.JobFormatParams.OfHlsH264 = hlsH264Params
	case string(FormatHlsH265):
		hlsH265Params := &chunkify.HlsH265Param{ID: FormatHlsH265}
		if width != nil && *width > 0 {
			hlsH265Params.Width = chunkify.Int(*width)
		}
		if height != nil && *height > 0 {
			hlsH265Params.Height = chunkify.Int(*height)
		}
		if framerate != nil && *framerate > 0 {
			hlsH265Params.Framerate = chunkify.Float(*framerate)
		}
		if gop != nil && *gop > 0 {
			hlsH265Params.Gop = chunkify.Int(*gop)
		}
		if channels != nil && *channels > 0 {
			hlsH265Params.Channels = *channels
		}
		if maxrate != nil && *maxrate > 0 {
			hlsH265Params.Maxrate = chunkify.Int(*maxrate)
		}
		if bufsize != nil && *bufsize > 0 {
			hlsH265Params.Bufsize = chunkify.Int(*bufsize)
		}
		if pixfmt != nil && *pixfmt != "" {
			hlsH265Params.Pixfmt = chunkify.HlsH265Pixfmt(*pixfmt)
		}
		if disableAudio != nil && *disableAudio {
			hlsH265Params.DisableAudio = chunkify.Bool(true)
		}
		if disableVideo != nil && *disableVideo {
			hlsH265Params.DisableVideo = chunkify.Bool(true)
		}
		if duration != nil && *duration > 0 {
			hlsH265Params.Duration = chunkify.Int(*duration)
		}
		if seek != nil && *seek > 0 {
			hlsH265Params.Seek = chunkify.Int(*seek)
		}
		if videoBitrate != nil && *videoBitrate > 0 {
			hlsH265Params.VideoBitrate = chunkify.Int(*videoBitrate)
		}
		if audioBitrate != nil && *audioBitrate > 0 {
			hlsH265Params.AudioBitrate = chunkify.Int(*audioBitrate)
		}

		if hlsTime != nil && *hlsTime > 0 {
			hlsH265Params.HlsTime = chunkify.Int(*hlsTime)
		}
		if hlsSegmentType != nil && *hlsSegmentType != "" {
			hlsH265Params.HlsSegmentType = chunkify.HlsH265HlsSegmentType(*hlsSegmentType)
		}
		if hlsEnc != nil && *hlsEnc {
			hlsH265Params.HlsEnc = chunkify.Bool(true)
		}
		if hlsEncKey != nil && *hlsEncKey != "" {
			hlsH265Params.HlsEncKey = chunkify.String(*hlsEncKey)
		}
		if hlsEncKeyUrl != nil && *hlsEncKeyUrl != "" {
			hlsH265Params.HlsEncKeyURL = chunkify.String(*hlsEncKeyUrl)
		}
		if hlsEncIv != nil && *hlsEncIv != "" {
			hlsH265Params.HlsEncIv = chunkify.String(*hlsEncIv)
		}
		if crf != nil && *crf > 0 {
			hlsH265Params.Crf = chunkify.Int(*crf)
		}
		if preset != nil && *preset != "" {
			hlsH265Params.Preset = chunkify.HlsH265Preset(*preset)
		}
		if profilev != nil && *profilev != "" {
			hlsH265Params.Profilev = chunkify.HlsH265Profilev(*profilev)
		}
		if level != nil && *level > 0 {
			hlsH265Params.Level = *level
		}
		if x264KeyInt != nil && *x264KeyInt > 0 {
			hlsH265Params.X265Keyint = chunkify.Int(*x265KeyInt)
		}

		app.Command.JobFormatParams.OfHlsH265 = hlsH265Params
	case string(FormatHlsAv1):
		hlsAv1Params := &chunkify.HlsAv1Param{ID: FormatHlsAv1}
		if width != nil && *width > 0 {
			hlsAv1Params.Width = chunkify.Int(*width)
		}
		if height != nil && *height > 0 {
			hlsAv1Params.Height = chunkify.Int(*height)
		}
		if framerate != nil && *framerate > 0 {
			hlsAv1Params.Framerate = chunkify.Float(*framerate)
		}
		if gop != nil && *gop > 0 {
			hlsAv1Params.Gop = chunkify.Int(*gop)
		}
		if channels != nil && *channels > 0 {
			hlsAv1Params.Channels = *channels
		}
		if maxrate != nil && *maxrate > 0 {
			hlsAv1Params.Maxrate = chunkify.Int(*maxrate)
		}
		if bufsize != nil && *bufsize > 0 {
			hlsAv1Params.Bufsize = chunkify.Int(*bufsize)
		}
		if pixfmt != nil && *pixfmt != "" {
			hlsAv1Params.Pixfmt = chunkify.HlsAv1Pixfmt(*pixfmt)
		}
		if disableAudio != nil && *disableAudio {
			hlsAv1Params.DisableAudio = chunkify.Bool(true)
		}
		if disableVideo != nil && *disableVideo {
			hlsAv1Params.DisableVideo = chunkify.Bool(true)
		}
		if duration != nil && *duration > 0 {
			hlsAv1Params.Duration = chunkify.Int(*duration)
		}
		if seek != nil && *seek > 0 {
			hlsAv1Params.Seek = chunkify.Int(*seek)
		}
		if videoBitrate != nil && *videoBitrate > 0 {
			hlsAv1Params.VideoBitrate = chunkify.Int(*videoBitrate)
		}
		if audioBitrate != nil && *audioBitrate > 0 {
			hlsAv1Params.AudioBitrate = chunkify.Int(*audioBitrate)
		}

		if hlsTime != nil && *hlsTime > 0 {
			hlsAv1Params.HlsTime = chunkify.Int(*hlsTime)
		}
		if hlsSegmentType != nil && *hlsSegmentType != "" {
			hlsAv1Params.HlsSegmentType = chunkify.HlsAv1HlsSegmentType(*hlsSegmentType)
		}
		if hlsEnc != nil && *hlsEnc {
			hlsAv1Params.HlsEnc = chunkify.Bool(true)
		}
		if hlsEncKey != nil && *hlsEncKey != "" {
			hlsAv1Params.HlsEncKey = chunkify.String(*hlsEncKey)
		}
		if hlsEncKeyUrl != nil && *hlsEncKeyUrl != "" {
			hlsAv1Params.HlsEncKeyURL = chunkify.String(*hlsEncKeyUrl)
		}
		if hlsEncIv != nil && *hlsEncIv != "" {
			hlsAv1Params.HlsEncIv = chunkify.String(*hlsEncIv)
		}
		if crf != nil && *crf > 0 {
			hlsAv1Params.Crf = chunkify.Int(*crf)
		}
		if preset != nil && *preset != "" {
			hlsAv1Params.Preset = chunkify.HlsAv1Preset(*preset)
		}
		if profilev != nil && *profilev != "" {
			hlsAv1Params.Profilev = chunkify.HlsAv1Profilev(*profilev)
		}
		if level != nil && *level > 0 {
			hlsAv1Params.Level = *level
		}
		app.Command.JobFormatParams.OfHlsAv1 = hlsAv1Params
	case string(FormatJpg):
		jpgParams := &chunkify.JpgParam{ID: FormatJpg}
		if width != nil && *width > 0 {
			jpgParams.Width = chunkify.Int(*width)
		}
		if height != nil && *height > 0 {
			jpgParams.Height = chunkify.Int(*height)
		}
		if interval != nil && *interval > 0 {
			jpgParams.Interval = *interval
		}
		if sprite != nil && *sprite {
			jpgParams.Sprite = chunkify.Bool(true)
		}
		app.Command.JobFormatParams.OfJpg = jpgParams
	}
}
