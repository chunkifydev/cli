package chunkify

import (
	"fmt"
	"slices"
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

	flags.StringVar(rcmd.Flags(), &chunkifyCmd.Input, "input", "", "Video file or URL to process")
	flags.StringVar(rcmd.Flags(), &chunkifyCmd.Output, "output", "", "Output file")
	flags.StringVar(rcmd.Flags(), &chunkifyCmd.Format, "format", string(chunkify.FormatMp4H264), "chunkify formats: mp4/h264, mp4/h265, mp4/av1, webm/vp9, hls/h264, hls/h265, hls/av1, jpg")
	flags.Int64VarPtr(rcmd.Flags(), &transcoders, "transcoders", 0, "chunkify transcoder quantity: Transcoders")
	flags.Int64VarPtr(rcmd.Flags(), &transcoderVcpu, "vcpu", 8, "chunkify transcoder vCPU: 4, 8 or 16")

	// Common video settings
	flags.Int64VarPtr(rcmd.Flags(), &width, "width", 0, "ffmpeg config: Width")
	flags.Int64VarPtr(rcmd.Flags(), &height, "height", 0, "ffmpeg config: Height")
	flags.Float64VarPtr(rcmd.Flags(), &framerate, "framerate", 0, "ffmpeg config: Framerate")
	flags.Int64VarPtr(rcmd.Flags(), &gop, "gop", 0, "ffmpeg config: Gop")
	flags.Int64VarPtr(rcmd.Flags(), &channels, "channels", 0, "ffmpeg config: Channels")
	flags.Int64VarPtr(rcmd.Flags(), &maxrate, "maxrate", 0, "ffmpeg config: Maxrate")
	flags.Int64VarPtr(rcmd.Flags(), &bufsize, "bufsize", 0, "ffmpeg config: Bufsize")
	flags.StringVarPtr(rcmd.Flags(), &pixfmt, "pixfmt", "", "ffmpeg config: PixFmt")
	flags.BoolVarPtr(rcmd.Flags(), &disableAudio, "an", false, "ffmpeg config: DisableAudio")
	flags.BoolVarPtr(rcmd.Flags(), &disableVideo, "vn", false, "ffmpeg config: DisableVideo")
	flags.Int64VarPtr(rcmd.Flags(), &duration, "duration", 0, "ffmpeg config: Duration")
	flags.Int64VarPtr(rcmd.Flags(), &seek, "seek", 0, "ffmpeg config: Seek")
	flags.Int64VarPtr(rcmd.Flags(), &videoBitrate, "vb", 0, "ffmpeg config: VideoBitrate")
	flags.Int64VarPtr(rcmd.Flags(), &audioBitrate, "ab", 0, "ffmpeg config: AudioBitrate")

	// H264, H265 and AV1 flags
	flags.Int64VarPtr(rcmd.Flags(), &crf, "crf", 0, "ffmpeg config: Crf")
	flags.StringVarPtr(rcmd.Flags(), &preset, "preset", "", "ffmpeg config: Preset")
	flags.StringVarPtr(rcmd.Flags(), &profilev, "profilev", "", "ffmpeg config: Profilev")
	flags.Int64VarPtr(rcmd.Flags(), &level, "level", 0, "ffmpeg config: Level")
	flags.Int64VarPtr(rcmd.Flags(), &x264KeyInt, "x264keyint", 0, "ffmpeg config: X264KeyInt")
	flags.Int64VarPtr(rcmd.Flags(), &x265KeyInt, "x265keyint", 0, "ffmpeg config: X265KeyInt")

	// VP9 flags
	flags.StringVarPtr(rcmd.Flags(), &quality, "quality", "", "ffmpeg config: Quality")
	flags.StringVarPtr(rcmd.Flags(), &cpuUsed, "cpuUsed", "", "ffmpeg config: CpuUsed")

	// HLS flags
	flags.StringVarPtr(rcmd.Flags(), &hlsManifestId, "hls-manifest-id", "", "ffmpeg config: HlsManifestId")
	flags.Int64VarPtr(rcmd.Flags(), &hlsTime, "hlsTime", 0, "ffmpeg config: HlsTime")
	flags.StringVarPtr(rcmd.Flags(), &hlsSegmentType, "hlsSegmentType", "", "ffmpeg config: HlsSegmentType")
	flags.BoolVarPtr(rcmd.Flags(), &hlsEnc, "hlsEnc", false, "ffmpeg config: HlsEnc")
	flags.StringVarPtr(rcmd.Flags(), &hlsEncKey, "hlsEncKey", "", "ffmpeg config: HlsEncKey")
	flags.StringVarPtr(rcmd.Flags(), &hlsEncKeyUrl, "hlsEncKeyUrl", "", "ffmpeg config: HlsEncKeyUrl")
	flags.StringVarPtr(rcmd.Flags(), &hlsEncIv, "hlsEncIv", "", "ffmpeg config: HlsEncIv")

	// JPG flags
	flags.Int64VarPtr(rcmd.Flags(), &interval, "interval", 0, "ffmpeg config: Interval in seconds (jpg)")
	flags.BoolVarPtr(rcmd.Flags(), &sprite, "sprite", false, "ffmpeg config: Sprite (jpg)")

	rcmd.MarkFlagRequired("input")
	rcmd.MarkFlagRequired("output")

	rcmd.MarkFlagsRequiredTogether("transcoders", "vcpu")

	rcmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		// build job format params according to all format flags
		setJobFormatParams()

		if !slices.Contains(validFormats, chunkify.FormatName(chunkifyCmd.Format)) {
			return fmt.Errorf("invalid format: %s", chunkifyCmd.Format)
		}

		if transcoders != nil && *transcoders > 0 {
			chunkifyCmd.JobTranscoderParams = &chunkify.JobCreateTranscoderParams{
				Quantity: *transcoders,
				Type:     fmt.Sprintf("%dvCPU", *transcoderVcpu),
			}
		}

		if chunkifyCmd.Format == string(chunkify.FormatJpg) {
			if interval == nil || *interval == 0 {
				return fmt.Errorf("--interval flag is required when format is jpg")
			}
		}
		if strings.HasPrefix(chunkifyCmd.Format, "hls") {
			if videoBitrate == nil || *videoBitrate == 0 {
				return fmt.Errorf("--vb (video bitrate) flag is required when format is hls")
			}
			if audioBitrate == nil || *audioBitrate == 0 {
				return fmt.Errorf("--ab (audio bitrate) flag is required when format is hls")
			}
		}

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
