package chunkify

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

var chunkifyCmd ChunkifyCommand

// Transcoder flags
var (
	transcoders    = new(int64)
	transcoderVcpu = new(int64)
)

// Common video flags
var (
	resolution   = new(string)
	width        = new(int64)
	height       = new(int64)
	framerate    = new(float64)
	gop          = new(int64)
	channels     = new(int64)
	maxrate      = new(int64)
	bufsize      = new(int64)
	pixfmt       = new(string)
	disableAudio = new(bool)
	disableVideo = new(bool)
	duration     = new(int64)
	seek         = new(int64)
	videoBitrate = new(int64)
	audioBitrate = new(int64)
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
func BindFlags(rcmd *cobra.Command) {
	chunkifyCmd = ChunkifyCommand{}

	flag.StringVarP(&chunkifyCmd.Input, "input", "i", "", "Input video to transcode. It can be a file, HTTP URL or source ID (src_*)")
	flag.StringVarP(&chunkifyCmd.Output, "output", "o", "", "Output file path")
	flag.StringVarP(&chunkifyCmd.Format, "format", "f", string(chunkify.FormatMp4H264), "Output format (mp4/h264, mp4/h265, mp4/av1, webm/vp9, hls/h264, hls/h265, hls/av1, jpg)")
	flag.Int64Var(transcoders, "transcoders", 0, "Number of transcoders to use")
	flag.Int64Var(transcoderVcpu, "vcpu", 8, "vCPU per transcoder (4, 8, or 16)")

	// Common video settings
	flag.StringVarP(resolution, "resolution", "s", "", "Set resolution wxh (0-8192x0-8192)")
	flag.Float64VarP(framerate, "framerate", "r", 0, "Set frame rate (15-120)")
	flag.Int64VarP(gop, "gop", "g", 0, "Set group of pictures size (1-300)")
	flag.Int64Var(channels, "channels", 0, "Set number of audio channels (1, 2, 5, 7)")
	flag.Int64Var(maxrate, "maxrate", 0, "Set maximum bitrate in bits per second (100000-50000000)")
	flag.Int64Var(bufsize, "bufsize", 0, "Set buffer size in bits (100000-50000000)")
	flag.StringVar(pixfmt, "pixfmt", "", "Set pixel format (yuv410p, yuv411p, yuv420p, yuv422p, yuv440p, yuv444p, yuvJ411p, yuvJ420p, yuvJ422p, yuvJ440p, yuvJ444p, yuv420p10le, yuv422p10le, yuv440p10le, yuv444p10le, yuv420p12le, yuv422p12le, yuv440p12le, yuv444p12le, yuv420p10be, yuv422p10be, yuv440p10be, yuv444p10be, yuv420p12be, yuv422p12be, yuv440p12be, yuv444p12be)")
	flag.BoolVar(disableAudio, "an", false, "Disable audio")
	flag.BoolVar(disableVideo, "vn", false, "Disable video")
	flag.Int64VarP(duration, "duration", "t", 0, "Set duration in seconds")
	flag.Int64Var(seek, "seek", 0, "Seek to position in seconds")
	flag.Int64Var(videoBitrate, "vb", 0, "Set video bitrate in bits per second (100000-50000000)")
	flag.Int64Var(audioBitrate, "ab", 0, "Set audio bitrate in bits per second (32000-512000)")

	// H264, H265 and AV1 flags
	flag.Int64Var(crf, "crf", 0, "Set constant rate factor (H264/H265: 16-35, AV1: 16-63, VP9: 15-35)")
	flag.StringVar(preset, "preset", "", "Set encoding preset (H264/H265: ultrafast, superfast, veryfast, faster, fast, medium, AV1: 6-13)")
	flag.StringVar(profilev, "profilev", "", "Set video profile (H264: baseline, main, high, high10, high422, high444, H265/AV1: main, main10, mainstillpicture)")
	flag.Int64Var(level, "level", 0, "Set encoding level (H264: 10, 11, 12, 13, 20, 21, 22, 30, 31, 32, 40, 41, 42, 50, 51, H265: 30, 31, 41, AV1: 30, 31, 41)")
	flag.Int64Var(x264KeyInt, "x264keyint", 0, "H264 - Set x264 keyframe interval")
	flag.Int64Var(x265KeyInt, "x265keyint", 0, "H265 - Set x265 keyframe interval")

	// VP9 flags
	flag.StringVar(quality, "quality", "", "Set VP9 quality (good, best, realtime)")
	flag.StringVar(cpuUsed, "cpu-used", "", "Set VP9 CPU usage (0-8)")

	// HLS flags
	flag.StringVar(hlsManifestId, "hls-manifest-id", "", "Set HLS manifest ID")
	flag.Int64Var(hlsTime, "hls-time", 0, "Set HLS segment duration in seconds (1-10)")
	flag.StringVar(hlsSegmentType, "hls-segment-type", "", "Set HLS segment type (mpegts, fmp4)")
	flag.BoolVar(hlsEnc, "hls-enc", false, "Enable HLS encryption")
	flag.StringVar(hlsEncKey, "hls-enc-key", "", "Set HLS encryption key")
	flag.StringVar(hlsEncKeyUrl, "hls-enc-key-url", "", "Set HLS encryption key URL")
	flag.StringVar(hlsEncIv, "hls-enc-iv", "", "Set HLS encryption IV")

	// JPG flags
	flag.Int64Var(interval, "interval", 0, "Set frame extraction interval in seconds (1-60)")
	flag.BoolVar(sprite, "sprite", false, "Generate sprite sheet")

	rcmd.MarkFlagRequired("input")
	rcmd.MarkFlagRequired("output")

	rcmd.MarkFlagsRequiredTogether("transcoders", "vcpu")

	rcmd.PreRunE = func(cmd *cobra.Command, args []string) error {
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
