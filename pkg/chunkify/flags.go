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
	FormatM4aAac  = "m4a_aac"
)

// SupportedFormats lists the output formats accepted by the Transcodings API.
var SupportedFormats = []string{FormatMp4H264, FormatMp4H265, FormatMp4Av1, FormatWebmVp9, FormatM4aAac}

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
	perTitle     = new(bool)
	duration     = new(float64)
	seek         = new(float64)

	maxrate         = new(int64)
	bufsize         = new(int64)
	videoBitrate    = new(int64)
	audioBitrate    = new(int64)
	audioSamplerate = new(int64)

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

// BindFlags attaches root-level flags used by the root command
func BindFlags(app *App, cmd *cobra.Command) {
	app.Command = &ChunkifyCommand{Id: uuid.New().String()}

	cmd.Flags().BoolVar(&app.JSON, "json", false, "Output in JSON format")
	cmd.Flags().StringVarP(&app.Command.Input, "input", "i", "", "Input video: local file, HTTP URL, source ID (src_*), or store://object-key")
	cmd.Flags().StringVar(&app.Command.SourceStorageID, "source-storage-id", "", "External storage for store:// input (takes precedence over config storage-id)")
	cmd.Flags().StringVarP(&app.Command.Output, "output", "o", "", "Output file path")
	cmd.Flags().StringVarP(&app.Command.Format, "format", "f", "", "Output format (mp4_h264, mp4_h265, mp4_av1, webm_vp9, m4a_aac)")

	cmd.Flags().StringVar(storagePath, "storage-path", "", "Output object path relative to the storage base prefix")
	cmd.Flags().StringVar(storagePath, "output-storage-path", "", "Output object path relative to the storage base prefix")
	cmd.Flags().MarkDeprecated("storage-path", "use --output-storage-path instead")
	cmd.MarkFlagsMutuallyExclusive("storage-path", "output-storage-path")
	cmd.Flags().StringVar(&app.Command.UploadStorageID, "upload-storage-id", "", "Storage for local uploads (overridden by config storage-id)")
	cmd.Flags().StringVar(&app.Command.OutputStorageID, "output-storage-id", "", "Storage for transcoding outputs (overridden by config storage-id)")
	cmd.Flags().StringVar(&app.Command.UploadStoragePath, "upload-storage-path", "", "Exact object key for uploads to external storage, including the filename")

	// Common video settings
	cmd.Flags().StringVarP(resolution, "resolution", "s", "", "Set resolution wxh (0-8192x0-8192)")
	cmd.Flags().Float64VarP(framerate, "framerate", "r", 0, "Set frame rate (15-120)")
	cmd.Flags().Int64VarP(gop, "gop", "g", 0, "Set group of pictures size (1-300)")
	cmd.Flags().Int64Var(channels, "channels", 0, "Set number of audio channels (1, 2, 5, 7)")
	cmd.Flags().StringVar(maxrateStr, "maxrate", "", "Set maximum bitrate in bits per second (100000-50000000). You can use units like K, M (e.g. 1200K, 2M)")
	cmd.Flags().StringVar(bufsizeStr, "bufsize", "", "Set buffer size in bits (100000-50000000). You can use units like K, M (e.g. 1200K, 2M)")
	cmd.Flags().StringVar(pixfmt, "pixfmt", "", "Set pixel format (H264: yuv420p, yuv422p, yuv444p and their 10le variants, H265: adds 12le variants, AV1: yuv420p, yuv420p10le, VP9: yuv420p, yuv422p, yuv440p, yuv444p and their 10le/12le variants)")
	cmd.Flags().BoolVar(disableAudio, "an", false, "Disable audio")
	cmd.Flags().BoolVar(disableVideo, "vn", false, "Disable video")
	cmd.Flags().BoolVar(perTitle, "per-title", false, "Automatically select video rate-control settings for each source")
	cmd.Flags().Float64VarP(duration, "duration", "t", 0, "Set duration in seconds")
	cmd.Flags().Float64Var(seek, "seek", 0, "Seek to position in seconds")
	cmd.Flags().StringVar(videoBitrateStr, "vb", "", "Set video bitrate in bits per second (100000-50000000). You can use units like K, M (e.g. 1200K, 2M)")
	cmd.Flags().StringVar(audioBitrateStr, "ab", "", "Set audio bitrate in bits per second (32000-512000). You can use units like K, M (e.g. 1200K, 2M)")
	cmd.Flags().Int64Var(audioSamplerate, "ar", 0, "Set audio sample rate in Hz (8000, 11025, 12000, 16000, 22050, 24000, 32000, 44100, 48000)")

	// H264, H265 and AV1 flags
	cmd.Flags().Int64Var(crf, "crf", 0, "Set constant rate factor (H264/H265: 16-35, AV1: 16-63, VP9: 15-35)")
	cmd.Flags().StringVar(preset, "preset", "", "Set encoding preset (H264/H265: ultrafast, superfast, veryfast, faster, fast, medium, AV1: 6-13)")
	cmd.Flags().StringVar(profilev, "profilev", "", "Set video profile (H264: baseline, main, high, high10, high422, high444, H265: main, main10, main12, main422-10, main422-12, main444-8, main444-10, main444-12, AV1: main)")
	cmd.Flags().Int64Var(level, "level", 0, "Set encoding level (H264: 10, 11, 12, 13, 20, 21, 22, 30, 31, 32, 40, 41, 42, 50, 51, H265: 30, 31, 41, AV1: 30, 31, 41)")
	cmd.Flags().Int64Var(x264KeyInt, "x264keyint", 0, "H264 - Set x264 keyframe interval")
	cmd.Flags().Int64Var(x265KeyInt, "x265keyint", 0, "H265 - Set x265 keyframe interval")

	// VP9 flags
	cmd.Flags().StringVar(quality, "quality", "", "Set VP9 quality (good, best, realtime)")
	cmd.Flags().StringVar(cpuUsed, "cpu-used", "", "Set VP9 CPU usage (0-8)")

	cmd.MarkFlagRequired("input")

	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if err := app.validateSourceInput(); err != nil {
			return err
		}
		if err := setupCommand(app); err != nil {
			return err
		}

		if err := validateTranscodeSettings(app); err != nil {
			return err
		}

		// build transcoding format params according to all format flags
		setTranscodingFormatParams(app)

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
		case ".m4a":
			app.Command.Format = FormatM4aAac
		default:
			return fmt.Errorf("invalid output file extension: %s. Please provide a valid format with --format", path.Ext(app.Command.Output))
		}
	}

	// Set the storage path
	if storagePath != nil && *storagePath != "" {
		app.Command.TranscodingStorageParams.Path = chunkify.String(*storagePath)
	}
	if app.Command.OutputStorageID != "" {
		app.Command.TranscodingStorageParams.ID = chunkify.String(app.Command.OutputStorageID)
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
	if !slices.Contains(SupportedFormats, app.Command.Format) {
		return fmt.Errorf("invalid format: %s. Supported formats: %s", app.Command.Format, strings.Join(SupportedFormats, ", "))
	}

	if err := validateCommonVideoFlags(); err != nil {
		return err
	}
	if err := validatePerTitleFlags(app.Command.Format); err != nil {
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

	case FormatM4aAac:
		if err := validateM4aAacFlags(); err != nil {
			return err
		}
	}

	return nil
}

func setTranscodingFormatParams(app *App) {
	app.Command.TranscodingFormatParams = chunkify.TranscodingNewParamsFormatUnion{}

	switch app.Command.Format {
	case FormatMp4H264:
		p := &chunkify.TranscodingMP4H264Param{}
		setVideoParams(p)
		setParam(&p.Crf, crf)
		setParam(&p.Preset, preset)
		setParam(&p.Profilev, profilev)
		setParam(&p.Level, level)
		setParam(&p.X264Keyint, x264KeyInt)
		app.Command.TranscodingFormatParams.OfMP4H264 = p
	case FormatMp4H265:
		p := &chunkify.TranscodingMP4H265Param{}
		setVideoParams(p)
		setParam(&p.Crf, crf)
		setParam(&p.Preset, preset)
		setParam(&p.Profilev, profilev)
		setParam(&p.Level, level)
		setParam(&p.X265Keyint, x265KeyInt)
		app.Command.TranscodingFormatParams.OfMP4H265 = p
	case FormatMp4Av1:
		p := &chunkify.TranscodingMP4Av1Param{}
		setVideoParams(p)
		setParam(&p.Crf, crf)
		setParam(&p.Preset, preset)
		setParam(&p.Profilev, profilev)
		setParam(&p.Level, level)
		app.Command.TranscodingFormatParams.OfMP4Av1 = p
	case FormatWebmVp9:
		p := &chunkify.TranscodingWebmVp9Param{}
		setVideoParams(p)
		setParam(&p.Crf, crf)
		setParam(&p.Quality, quality)
		setParam(&p.CPUUsed, cpuUsed)
		app.Command.TranscodingFormatParams.OfWebmVp9 = p
	case FormatM4aAac:
		p := &chunkify.TranscodingM4aAacParam{}
		setParam(&p.Channels, channels)
		setParam(&p.Duration, duration)
		setParam(&p.Seek, seek)
		setParam(&p.AudioBitrate, audioBitrate)
		setParam(&p.AudioSamplerate, audioSamplerate)
		app.Command.TranscodingFormatParams.OfM4aAac = p
	}
}

// setVideoParams sets the fields shared by every video format. Each format
// param struct uses the same field names, so they are set by name.
func setVideoParams(p any) {
	elem := reflect.ValueOf(p).Elem()
	for name, val := range map[string]any{
		"Width":           width,
		"Height":          height,
		"Framerate":       framerate,
		"Gop":             gop,
		"Channels":        channels,
		"Maxrate":         maxrate,
		"Bufsize":         bufsize,
		"Pixfmt":          pixfmt,
		"DisableAudio":    disableAudio,
		"DisableVideo":    disableVideo,
		"PerTitle":        perTitle,
		"Duration":        duration,
		"Seek":            seek,
		"VideoBitrate":    videoBitrate,
		"AudioBitrate":    audioBitrate,
		"AudioSamplerate": audioSamplerate,
	} {
		field := elem.FieldByName(name)
		if !field.IsValid() {
			continue
		}
		setParam(field.Addr().Interface(), val)
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
