package chunkify

import (
	"fmt"
	"path"
	"reflect"
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

		setParam(&h264Params.Width, width)
		setParam(&h264Params.Height, height)
		setParam(&h264Params.Framerate, framerate)
		setParam(&h264Params.Gop, gop)
		setParam(&h264Params.Channels, channels)
		setParam(&h264Params.Maxrate, maxrate)
		setParam(&h264Params.Bufsize, bufsize)
		setParam(&h264Params.Pixfmt, pixfmt)
		setParam(&h264Params.DisableAudio, disableAudio)
		setParam(&h264Params.DisableVideo, disableVideo)
		setParam(&h264Params.Duration, duration)
		setParam(&h264Params.Seek, seek)
		setParam(&h264Params.VideoBitrate, videoBitrate)
		setParam(&h264Params.AudioBitrate, audioBitrate)
		setParam(&h264Params.Crf, crf)
		setParam(&h264Params.Preset, preset)
		setParam(&h264Params.Profilev, profilev)
		setParam(&h264Params.Level, level)
		setParam(&h264Params.X264Keyint, x264KeyInt)

		app.Command.JobFormatParams.OfMP4H264 = h264Params
	case string(FormatMp4H265):
		h265Params := &chunkify.MP4H265Param{ID: FormatMp4H265}

		setParam(&h265Params.Width, width)
		setParam(&h265Params.Height, height)
		setParam(&h265Params.Framerate, framerate)
		setParam(&h265Params.Gop, gop)
		setParam(&h265Params.Channels, channels)
		setParam(&h265Params.Maxrate, maxrate)
		setParam(&h265Params.Bufsize, bufsize)
		setParam(&h265Params.Pixfmt, pixfmt)
		setParam(&h265Params.DisableAudio, disableAudio)
		setParam(&h265Params.DisableVideo, disableVideo)
		setParam(&h265Params.Duration, duration)
		setParam(&h265Params.Seek, seek)
		setParam(&h265Params.VideoBitrate, videoBitrate)
		setParam(&h265Params.AudioBitrate, audioBitrate)
		setParam(&h265Params.Crf, crf)
		setParam(&h265Params.Preset, preset)
		setParam(&h265Params.Profilev, profilev)
		setParam(&h265Params.Level, level)
		setParam(&h265Params.X265Keyint, x265KeyInt)

		app.Command.JobFormatParams.OfMP4H265 = h265Params
	case string(FormatMp4Av1):
		av1Params := &chunkify.MP4Av1Param{ID: FormatMp4Av1}

		setParam(&av1Params.Width, width)
		setParam(&av1Params.Height, height)
		setParam(&av1Params.Framerate, framerate)
		setParam(&av1Params.Gop, gop)
		setParam(&av1Params.Channels, channels)
		setParam(&av1Params.Maxrate, maxrate)
		setParam(&av1Params.Bufsize, bufsize)
		setParam(&av1Params.Pixfmt, pixfmt)
		setParam(&av1Params.DisableAudio, disableAudio)
		setParam(&av1Params.DisableVideo, disableVideo)
		setParam(&av1Params.Duration, duration)
		setParam(&av1Params.Seek, seek)
		setParam(&av1Params.VideoBitrate, videoBitrate)
		setParam(&av1Params.AudioBitrate, audioBitrate)
		setParam(&av1Params.Crf, crf)
		setParam(&av1Params.Preset, preset)
		setParam(&av1Params.Profilev, profilev)
		setParam(&av1Params.Level, level)

		app.Command.JobFormatParams.OfMP4Av1 = av1Params
	case string(FormatWebmVp9):
		vp9Params := &chunkify.WebmVp9Param{ID: FormatWebmVp9}

		setParam(&vp9Params.Width, width)
		setParam(&vp9Params.Height, height)
		setParam(&vp9Params.Framerate, framerate)
		setParam(&vp9Params.Gop, gop)
		setParam(&vp9Params.Channels, channels)
		setParam(&vp9Params.Maxrate, maxrate)
		setParam(&vp9Params.Bufsize, bufsize)
		setParam(&vp9Params.Pixfmt, pixfmt)
		setParam(&vp9Params.DisableAudio, disableAudio)
		setParam(&vp9Params.DisableVideo, disableVideo)
		setParam(&vp9Params.Duration, duration)
		setParam(&vp9Params.Seek, seek)
		setParam(&vp9Params.VideoBitrate, videoBitrate)
		setParam(&vp9Params.AudioBitrate, audioBitrate)
		setParam(&vp9Params.Crf, crf)
		setParam(&vp9Params.Quality, quality)
		setParam(&vp9Params.CPUUsed, cpuUsed)

		app.Command.JobFormatParams.OfWebmVp9 = vp9Params
	case string(FormatHlsH264):
		hlsH264Params := &chunkify.HlsH264Param{ID: FormatHlsH264}

		setParam(&hlsH264Params.HlsTime, hlsTime)
		setParam(&hlsH264Params.HlsSegmentType, hlsSegmentType)
		setParam(&hlsH264Params.HlsEnc, hlsEnc)
		setParam(&hlsH264Params.HlsEncKey, hlsEncKey)
		setParam(&hlsH264Params.HlsEncKeyURL, hlsEncKeyUrl)
		setParam(&hlsH264Params.HlsEncIv, hlsEncIv)

		setParam(&hlsH264Params.Width, width)
		setParam(&hlsH264Params.Height, height)
		setParam(&hlsH264Params.Framerate, framerate)
		setParam(&hlsH264Params.Gop, gop)
		setParam(&hlsH264Params.Channels, channels)
		setParam(&hlsH264Params.Maxrate, maxrate)
		setParam(&hlsH264Params.Bufsize, bufsize)
		setParam(&hlsH264Params.Pixfmt, pixfmt)
		setParam(&hlsH264Params.DisableAudio, disableAudio)
		setParam(&hlsH264Params.DisableVideo, disableVideo)
		setParam(&hlsH264Params.Duration, duration)
		setParam(&hlsH264Params.Seek, seek)
		setParam(&hlsH264Params.VideoBitrate, videoBitrate)
		setParam(&hlsH264Params.AudioBitrate, audioBitrate)
		setParam(&hlsH264Params.Crf, crf)
		setParam(&hlsH264Params.Preset, preset)
		setParam(&hlsH264Params.Profilev, profilev)
		setParam(&hlsH264Params.Level, level)
		setParam(&hlsH264Params.X264Keyint, x264KeyInt)

		app.Command.JobFormatParams.OfHlsH264 = hlsH264Params
	case string(FormatHlsH265):
		hlsH265Params := &chunkify.HlsH265Param{ID: FormatHlsH265}

		setParam(&hlsH265Params.HlsTime, hlsTime)
		setParam(&hlsH265Params.HlsSegmentType, hlsSegmentType)
		setParam(&hlsH265Params.HlsEnc, hlsEnc)
		setParam(&hlsH265Params.HlsEncKey, hlsEncKey)
		setParam(&hlsH265Params.HlsEncKeyURL, hlsEncKeyUrl)
		setParam(&hlsH265Params.HlsEncIv, hlsEncIv)

		setParam(&hlsH265Params.Width, width)
		setParam(&hlsH265Params.Height, height)
		setParam(&hlsH265Params.Framerate, framerate)
		setParam(&hlsH265Params.Gop, gop)
		setParam(&hlsH265Params.Channels, channels)
		setParam(&hlsH265Params.Maxrate, maxrate)
		setParam(&hlsH265Params.Bufsize, bufsize)
		setParam(&hlsH265Params.Pixfmt, pixfmt)
		setParam(&hlsH265Params.DisableAudio, disableAudio)
		setParam(&hlsH265Params.DisableVideo, disableVideo)
		setParam(&hlsH265Params.Duration, duration)
		setParam(&hlsH265Params.Seek, seek)
		setParam(&hlsH265Params.VideoBitrate, videoBitrate)
		setParam(&hlsH265Params.AudioBitrate, audioBitrate)
		setParam(&hlsH265Params.Crf, crf)
		setParam(&hlsH265Params.Preset, preset)
		setParam(&hlsH265Params.Profilev, profilev)
		setParam(&hlsH265Params.Level, level)
		setParam(&hlsH265Params.X265Keyint, x265KeyInt)

		app.Command.JobFormatParams.OfHlsH265 = hlsH265Params
	case string(FormatHlsAv1):
		hlsAv1Params := &chunkify.HlsAv1Param{ID: FormatHlsAv1}

		setParam(&hlsAv1Params.HlsTime, hlsTime)
		setParam(&hlsAv1Params.HlsSegmentType, hlsSegmentType)
		setParam(&hlsAv1Params.HlsEnc, hlsEnc)
		setParam(&hlsAv1Params.HlsEncKey, hlsEncKey)
		setParam(&hlsAv1Params.HlsEncKeyURL, hlsEncKeyUrl)
		setParam(&hlsAv1Params.HlsEncIv, hlsEncIv)

		setParam(&hlsAv1Params.Width, width)
		setParam(&hlsAv1Params.Height, height)
		setParam(&hlsAv1Params.Framerate, framerate)
		setParam(&hlsAv1Params.Gop, gop)
		setParam(&hlsAv1Params.Channels, channels)
		setParam(&hlsAv1Params.Maxrate, maxrate)
		setParam(&hlsAv1Params.Bufsize, bufsize)
		setParam(&hlsAv1Params.Pixfmt, pixfmt)
		setParam(&hlsAv1Params.DisableAudio, disableAudio)
		setParam(&hlsAv1Params.DisableVideo, disableVideo)
		setParam(&hlsAv1Params.Duration, duration)
		setParam(&hlsAv1Params.Seek, seek)
		setParam(&hlsAv1Params.VideoBitrate, videoBitrate)
		setParam(&hlsAv1Params.AudioBitrate, audioBitrate)
		setParam(&hlsAv1Params.Crf, crf)
		setParam(&hlsAv1Params.Preset, preset)
		setParam(&hlsAv1Params.Level, level)

		app.Command.JobFormatParams.OfHlsAv1 = hlsAv1Params
	case string(FormatJpg):
		jpgParams := &chunkify.JpgParam{ID: FormatJpg}
		setParam(&jpgParams.Width, width)
		setParam(&jpgParams.Height, height)
		setParam(&jpgParams.Interval, interval)
		setParam(&jpgParams.Sprite, sprite)

		app.Command.JobFormatParams.OfJpg = jpgParams
	}
}

func setParam(ptr any, val any) {
	if ptr == nil || val == nil {
		return
	}

	// Handle pointer types for val
	var intVal *int64
	var floatVal *float64
	var stringVal *string
	var boolVal *bool

	switch v := val.(type) {
	case *int64:
		if v == nil || *v <= 0 {
			return
		}
		intVal = v
	case *float64:
		if v == nil || *v <= 0 {
			return
		}
		floatVal = v
	case *string:
		if v == nil || *v == "" {
			return
		}
		stringVal = v
	case *bool:
		if v == nil || !*v {
			return
		}
		boolVal = v
	default:
		return
	}

	// Set the value using reflection
	ptrValue := reflect.ValueOf(ptr)
	if ptrValue.Kind() != reflect.Ptr {
		return
	}
	elem := ptrValue.Elem()

	switch {
	case intVal != nil:
		if elem.CanSet() {
			intResult := chunkify.Int(*intVal)
			resultValue := reflect.ValueOf(intResult)
			// Check if result type can be assigned to element type
			if resultValue.Type().AssignableTo(elem.Type()) {
				elem.Set(resultValue)
			} else if elem.Type().Kind() == reflect.Int64 {
				// Direct int64 assignment (like Channels, Level)
				elem.SetInt(*intVal)
			} else if resultValue.Type().ConvertibleTo(elem.Type()) {
				// Try to convert if possible
				elem.Set(resultValue.Convert(elem.Type()))
			}
		}
	case floatVal != nil:
		if elem.CanSet() {
			floatResult := chunkify.Float(*floatVal)
			resultValue := reflect.ValueOf(floatResult)
			if resultValue.Type().AssignableTo(elem.Type()) {
				elem.Set(resultValue)
			} else if resultValue.Type().ConvertibleTo(elem.Type()) {
				elem.Set(resultValue.Convert(elem.Type()))
			}
		}
	case stringVal != nil:
		if elem.CanSet() {
			// For string types, we'll need type-specific converters
			// For now, just set the string directly if the type matches
			if elem.Type().Kind() == reflect.String {
				elem.SetString(*stringVal)
			}
		}
	case boolVal != nil:
		if elem.CanSet() {
			boolResult := chunkify.Bool(true)
			resultValue := reflect.ValueOf(boolResult)
			if resultValue.Type().AssignableTo(elem.Type()) {
				elem.Set(resultValue)
			} else if resultValue.Type().ConvertibleTo(elem.Type()) {
				elem.Set(resultValue.Convert(elem.Type()))
			}
		}
	}
}
