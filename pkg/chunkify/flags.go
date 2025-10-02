package chunkify

import (
	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/cli/pkg/flags"
	"github.com/spf13/cobra"
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
)

// h264, h265 and av1 flags
var (
	crf        *int64
	preset     *string
	profilev   *string
	level      *int64
	x264KeyInt *int64
	x265KeyInt *int64
)

// jpg
var (
	interval *int64
	sprite   *bool
)

// BindFlags attaches root-level flags used by the root command
func BindFlags(rcmd *cobra.Command) {
	chunkifyCmd = ChunkifyCommand{}

	flags.StringVar(rcmd.Flags(), &chunkifyCmd.Input, "input", "", "Video file or URL to process")
	flags.StringVar(rcmd.Flags(), &chunkifyCmd.Output, "output", "", "Output file or directory")
	flags.StringVar(rcmd.Flags(), &chunkifyCmd.Format, "format", string(chunkify.FormatMp4H264), "chunkify format: mp4_h264, mp4_h265 or mp4_av1")
	flags.Int64VarPtr(rcmd.Flags(), &chunkifyCmd.Transcoders, "transcoders", 0, "chunkify transcoder quantity: Transcoders")
	flags.Int64VarPtr(rcmd.Flags(), &chunkifyCmd.TranscoderVcpu, "vcpu", 8, "chunkify transcoder vCPU: 4, 8 or 16")

	// format settings
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

	flags.Int64VarPtr(rcmd.Flags(), &crf, "crf", 0, "ffmpeg config: Crf")
	flags.StringVarPtr(rcmd.Flags(), &preset, "preset", "", "ffmpeg config: Preset")
	flags.StringVarPtr(rcmd.Flags(), &profilev, "profilev", "", "ffmpeg config: Profilev")
	flags.Int64VarPtr(rcmd.Flags(), &level, "level", 0, "ffmpeg config: Level")
	flags.Int64VarPtr(rcmd.Flags(), &x264KeyInt, "x264keyint", 0, "ffmpeg config: X264KeyInt")
	flags.Int64VarPtr(rcmd.Flags(), &x265KeyInt, "x265keyint", 0, "ffmpeg config: X265KeyInt")

	flags.Int64VarPtr(rcmd.Flags(), &interval, "interval", 0, "ffmpeg config: Interval in seconds (jpg)")
	flags.BoolVarPtr(rcmd.Flags(), &sprite, "sprite", false, "ffmpeg config: Sprite (jpg)")
}
