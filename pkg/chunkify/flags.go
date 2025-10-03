package chunkify

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/cli/pkg/flags"
	"github.com/spf13/cobra"
)

var chunkifyCmd ChunkifyCommand

var validFormats = []chunkify.FormatName{
	chunkify.FormatMp4H264,
	chunkify.FormatMp4H265,
	chunkify.FormatMp4Av1,
	chunkify.FormatWebmVp9,
	chunkify.FormatHlsH264,
	chunkify.FormatHlsH265,
	chunkify.FormatHlsAv1,
	chunkify.FormatJpg,
}

// Transcoder flags
var (
	transcoders    *int64
	transcoderVcpu *int64
)

// Common video flags
var (
	resolution   *string
	width        *int64
	height       *int64
	framerate    *float64
	gop          *int64
	channels     *int64
	maxrate      *int64
	bufsize      *int64
	pixfmt       *string
	disableAudio *bool
	disableVideo *bool
	duration     *int64
	seek         *int64
	videoBitrate *int64
	audioBitrate *int64
)

// H264, H265 and AV1 flags
var (
	crf        *int64
	preset     *string
	profilev   *string
	level      *int64
	x264KeyInt *int64
	x265KeyInt *int64
)

// VP9 flags
var (
	quality *string
	cpuUsed *string
)

// HLS flags
var (
	hlsManifestId  *string
	hlsTime        *int64
	hlsSegmentType *string
	hlsEnc         *bool
	hlsEncKey      *string
	hlsEncKeyUrl   *string
	hlsEncIv       *string
)

// JPG
var (
	interval *int64
	sprite   *bool
)

// BindFlags attaches root-level flags used by the root command
func BindFlags(rcmd *cobra.Command) {
	chunkifyCmd = ChunkifyCommand{}

	flags.StringVar(rcmd.Flags(), &chunkifyCmd.Input, "input", "", "Input video to transcode. It can be a file, URL or source ID (src_*)")
	flags.StringVar(rcmd.Flags(), &chunkifyCmd.Output, "output", "", "Output file path")
	flags.StringVar(rcmd.Flags(), &chunkifyCmd.Format, "format", string(chunkify.FormatMp4H264), "Output format (mp4/h264, mp4/h265, mp4/av1, webm/vp9, hls/h264, hls/h265, hls/av1, jpg)")
	flags.Int64VarPtr(rcmd.Flags(), &transcoders, "transcoders", 0, "Number of transcoders to use")
	flags.Int64VarPtr(rcmd.Flags(), &transcoderVcpu, "vcpu", 8, "vCPU per transcoder (4, 8, or 16)")

	// Common video settings
	flags.Int64VarPtr(rcmd.Flags(), &width, "width", 0, "Set video width (0-8192)")
	flags.Int64VarPtr(rcmd.Flags(), &height, "height", 0, "Set video height (0-8192)")
	flags.StringVarPtr(rcmd.Flags(), &resolution, "resolution", "", "Set resolution wxh")
	flags.Float64VarPtr(rcmd.Flags(), &framerate, "framerate", 0, "Set frame rate (15-120)")
	flags.Int64VarPtr(rcmd.Flags(), &gop, "gop", 0, "Set group of pictures size (1-300)")
	flags.Int64VarPtr(rcmd.Flags(), &channels, "channels", 0, "Set number of audio channels (1, 2, 5, 7)")
	flags.Int64VarPtr(rcmd.Flags(), &maxrate, "maxrate", 0, "Set maximum bitrate in bits per second (100000-50000000)")
	flags.Int64VarPtr(rcmd.Flags(), &bufsize, "bufsize", 0, "Set buffer size in bits (100000-50000000)")
	flags.StringVarPtr(rcmd.Flags(), &pixfmt, "pixfmt", "", "Set pixel format ("+strings.Join(validPixFmts, ", ")+")")
	flags.BoolVarPtr(rcmd.Flags(), &disableAudio, "an", false, "Disable audio")
	flags.BoolVarPtr(rcmd.Flags(), &disableVideo, "vn", false, "Disable video")
	flags.Int64VarPtr(rcmd.Flags(), &duration, "duration", 0, "Set duration in seconds")
	flags.Int64VarPtr(rcmd.Flags(), &seek, "seek", 0, "Seek to position in seconds")
	flags.Int64VarPtr(rcmd.Flags(), &videoBitrate, "vb", 0, "Set video bitrate in bits per second (100000-50000000)")
	flags.Int64VarPtr(rcmd.Flags(), &audioBitrate, "ab", 0, "Set audio bitrate in bits per second (32000-512000)")

	// H264, H265 and AV1 flags
	flags.Int64VarPtr(rcmd.Flags(), &crf, "crf", 0, "Set constant rate factor (H264/H265: 16-35, AV1: 16-63)")
	flags.StringVarPtr(rcmd.Flags(), &preset, "preset", "", "Set encoding preset (H264/H265: ultrafast, superfast, veryfast, faster, fast, medium, AV1: 6-13)")
	flags.StringVarPtr(rcmd.Flags(), &profilev, "profilev", "", "Set video profile (H264: baseline, main, high, high10, high422, high444, H265/AV1: main, main10, mainstillpicture)")
	flags.Int64VarPtr(rcmd.Flags(), &level, "level", 0, "Set encoding level (H264: 10, 11, 12, 13, 20, 21, 22, 30, 31, 32, 40, 41, 42, 50, 51, H265: 30, 31, 41, AV1: 30, 31, 41)")
	flags.Int64VarPtr(rcmd.Flags(), &x264KeyInt, "x264keyint", 0, "H264 - Set x264 keyframe interval")
	flags.Int64VarPtr(rcmd.Flags(), &x265KeyInt, "x265keyint", 0, "H265 - Set x265 keyframe interval")

	// VP9 flags
	flags.StringVarPtr(rcmd.Flags(), &quality, "quality", "", "Set VP9 quality (good, best, realtime)")
	flags.StringVarPtr(rcmd.Flags(), &cpuUsed, "cpu-used", "", "Set VP9 CPU usage (0-8)")

	// HLS flags
	flags.StringVarPtr(rcmd.Flags(), &hlsManifestId, "hls-manifest-id", "", "Set HLS manifest ID")
	flags.Int64VarPtr(rcmd.Flags(), &hlsTime, "hls-time", 0, "Set HLS segment duration in seconds (1-10)")
	flags.StringVarPtr(rcmd.Flags(), &hlsSegmentType, "hls-segment-type", "", "Set HLS segment type (mpegts, fmp4)")
	flags.BoolVarPtr(rcmd.Flags(), &hlsEnc, "hls-enc", false, "Enable HLS encryption")
	flags.StringVarPtr(rcmd.Flags(), &hlsEncKey, "hls-enc-key", "", "Set HLS encryption key")
	flags.StringVarPtr(rcmd.Flags(), &hlsEncKeyUrl, "hls-enc-key-url", "", "Set HLS encryption key URL")
	flags.StringVarPtr(rcmd.Flags(), &hlsEncIv, "hls-enc-iv", "", "Set HLS encryption IV")

	// JPG flags
	flags.Int64VarPtr(rcmd.Flags(), &interval, "interval", 0, "Set frame extraction interval in seconds (1-60)")
	flags.BoolVarPtr(rcmd.Flags(), &sprite, "sprite", false, "Generate sprite sheet")

	rcmd.MarkFlagRequired("input")
	rcmd.MarkFlagRequired("output")

	rcmd.MarkFlagsRequiredTogether("transcoders", "vcpu")

	rcmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if !slices.Contains(validFormats, chunkify.FormatName(chunkifyCmd.Format)) {
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
