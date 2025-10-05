package chunkify

import (
	"fmt"
	"path"
	"slices"
	"strconv"
	"strings"

	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/cli/pkg/formatter"
	"github.com/spf13/cobra"
)

var chunkifyCmd ChunkifyCommand

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
func BindFlags(cmd *cobra.Command) {
	chunkifyCmd = ChunkifyCommand{}

	cmd.Flags().StringVarP(&chunkifyCmd.Input, "input", "i", "", "Input video to transcode. It can be a file, HTTP URL or source ID (src_*)")
	cmd.Flags().StringVarP(&chunkifyCmd.Output, "output", "o", "", "Output file path")
	cmd.Flags().StringVarP(&chunkifyCmd.Format, "format", "f", "", "Output format (mp4/h264, mp4/h265, mp4/av1, webm/vp9, hls/h264, hls/h265, hls/av1, jpg)")

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
		if chunkifyCmd.Format == "" && chunkifyCmd.Output == "" {
			return nil
		}

		// Set default format based on output file extension
		if chunkifyCmd.Format == "" {
			switch path.Ext(chunkifyCmd.Output) {
			case ".mp4":
				chunkifyCmd.Format = string(chunkify.FormatMp4H264)
			case ".webm":
				chunkifyCmd.Format = string(chunkify.FormatWebmVp9)
			case ".m3u8":
				chunkifyCmd.Format = string(chunkify.FormatHlsH264)
			case ".jpg":
				chunkifyCmd.Format = string(chunkify.FormatJpg)
			default:
				return fmt.Errorf("invalid output file extension: %s. Please provide a valid format with --format", path.Ext(chunkifyCmd.Output))
			}
		}

		if !slices.Contains([]chunkify.FormatName{
			chunkify.FormatMp4H264,
			chunkify.FormatMp4H265,
			chunkify.FormatMp4Av1,
			chunkify.FormatWebmVp9,
			chunkify.FormatHlsH264,
			chunkify.FormatHlsH265,
			chunkify.FormatHlsAv1,
			chunkify.FormatJpg,
		}, chunkify.FormatName(chunkifyCmd.Format)) {
			return fmt.Errorf("invalid format: %s", chunkifyCmd.Format)
		}

		if transcoders != nil && *transcoders > 0 {
			chunkifyCmd.JobTranscoderParams = &chunkify.JobCreateTranscoderParams{
				Quantity: *transcoders,
				Type:     fmt.Sprintf("%dvCPU", *transcoderVcpu),
			}
		}

		if storagePath != nil && *storagePath != "" {
			chunkifyCmd.JobCreateStorageParams = &chunkify.JobCreateStorageParams{
				Path: storagePath,
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

		if err := validateCommonVideoFlags(); err != nil {
			return err
		}

		switch chunkifyCmd.Format {
		case string(chunkify.FormatMp4H264):
			if err := validateH264Flags(); err != nil {
				return err
			}

		case string(chunkify.FormatMp4H265):
			if err := validateH265Flags(); err != nil {
				return err
			}

		case string(chunkify.FormatMp4Av1):
			if err := validateAv1Flags(); err != nil {
				return err
			}

		case string(chunkify.FormatWebmVp9):
			if err := validateWebmVp9Flags(); err != nil {
				return err
			}

		case string(chunkify.FormatJpg):
			if err := validateJpgFlags(); err != nil {
				return err
			}

		case string(chunkify.FormatHlsH264):
			if err := validateHlsH264Flags(); err != nil {
				return err
			}

		case string(chunkify.FormatHlsH265):
			if err := validateHlsH265Flags(); err != nil {
				return err
			}

		case string(chunkify.FormatHlsAv1):
			if err := validateHlsAv1Flags(); err != nil {
				return err
			}

		}

		// build job format params according to all format flags
		setJobFormatParams()

		return nil
	}
}

func setJobFormatParams() {
	chunkifyCmd.JobFormatParams = chunkify.JobCreateFormatParams{}

	videoCommon := &chunkify.Video{
		Width:        width,
		Height:       height,
		Framerate:    framerate,
		Gop:          gop,
		Channels:     channels,
		Maxrate:      maxrate,
		Bufsize:      bufsize,
		DisableAudio: disableAudio,
		DisableVideo: disableVideo,
		Duration:     duration,
		Seek:         seek,
		PixFmt:       pixfmt,
		VideoBitrate: videoBitrate,
		AudioBitrate: audioBitrate,
	}

	switch chunkifyCmd.Format {
	case string(chunkify.FormatMp4H264):
		h264Params := &chunkify.H264{
			Video:      videoCommon,
			Crf:        crf,
			Preset:     preset,
			Profilev:   profilev,
			Level:      level,
			X264KeyInt: x264KeyInt,
		}
		chunkifyCmd.JobFormatParams.Mp4H264 = h264Params
	case string(chunkify.FormatMp4H265):
		h265Params := &chunkify.H265{
			Video:      videoCommon,
			Crf:        crf,
			Preset:     preset,
			Profilev:   profilev,
			Level:      level,
			X265KeyInt: x265KeyInt,
		}
		chunkifyCmd.JobFormatParams.Mp4H265 = h265Params
	case string(chunkify.FormatWebmVp9):
		vp9Params := &chunkify.Vp9{
			Video:   videoCommon,
			Crf:     crf,
			Quality: quality,
			CpuUsed: cpuUsed,
		}
		chunkifyCmd.JobFormatParams.WebmVp9 = vp9Params
	case string(chunkify.FormatMp4Av1):
		av1Params := &chunkify.Av1{
			Video:    videoCommon,
			Crf:      crf,
			Preset:   preset,
			Profilev: profilev,
			Level:    level,
		}
		chunkifyCmd.JobFormatParams.Mp4Av1 = av1Params
	case string(chunkify.FormatHlsH264):
		hlsH264Params := &chunkify.HlsH264{
			Hls: &chunkify.Hls{
				HlsTime:        hlsTime,
				HlsSegmentType: hlsSegmentType,
				HlsEnc:         hlsEnc,
				HlsEncKey:      hlsEncKey,
				HlsEncKeyUrl:   hlsEncKeyUrl,
				HlsEncIv:       hlsEncIv,
				VideoBitrate:   videoBitrate,
				AudioBitrate:   audioBitrate,
			},
			H264: &chunkify.H264{
				Video:      videoCommon,
				Crf:        crf,
				Preset:     preset,
				Profilev:   profilev,
				Level:      level,
				X264KeyInt: x264KeyInt,
			},
		}
		chunkifyCmd.JobFormatParams.HlsH264 = hlsH264Params
	case string(chunkify.FormatHlsH265):
		hlsH265Params := &chunkify.HlsH265{
			Hls: &chunkify.Hls{
				HlsTime:        hlsTime,
				HlsSegmentType: hlsSegmentType,
				HlsEnc:         hlsEnc,
				HlsEncKey:      hlsEncKey,
				HlsEncKeyUrl:   hlsEncKeyUrl,
				HlsEncIv:       hlsEncIv,
			},
			H265: &chunkify.H265{
				Video:      videoCommon,
				Crf:        crf,
				Preset:     preset,
				Profilev:   profilev,
				Level:      level,
				X265KeyInt: x265KeyInt,
			},
		}
		chunkifyCmd.JobFormatParams.HlsH265 = hlsH265Params
	case string(chunkify.FormatHlsAv1):
		hlsAv1Params := &chunkify.HlsAv1{
			Hls: &chunkify.Hls{
				HlsTime:        hlsTime,
				HlsSegmentType: hlsSegmentType,
				HlsEnc:         hlsEnc,
				HlsEncKey:      hlsEncKey,
				HlsEncKeyUrl:   hlsEncKeyUrl,
				HlsEncIv:       hlsEncIv,
			},
			Av1: &chunkify.Av1{
				Video:    videoCommon,
				Crf:      crf,
				Preset:   preset,
				Profilev: profilev,
				Level:    level,
			},
		}
		chunkifyCmd.JobFormatParams.HlsAv1 = hlsAv1Params
	case string(chunkify.FormatJpg):
		jpgParams := &chunkify.Jpg{
			Image: &chunkify.Image{
				Width:    width,
				Height:   height,
				Interval: *interval,
				Sprite:   sprite,
			},
		}
		chunkifyCmd.JobFormatParams.Jpg = jpgParams
	}
}
