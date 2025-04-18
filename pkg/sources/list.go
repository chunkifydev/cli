package sources

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/chunkifydev/cli/pkg/formatter"
	"github.com/chunkifydev/cli/pkg/styles"
	"github.com/spf13/cobra"

	chunkify "github.com/chunkifydev/chunkify-go"
)

type ListCmd struct {
	Offset      int64
	Limit       int64
	DurationEq  string
	DurationGte string
	DurationGt  string
	DurationLte string
	DurationLt  string
	CreatedGte  string
	CreatedLte  string
	WidthEq     int64
	WidthGte    int64
	WidthGt     int64
	WidthLte    int64
	WidthLt     int64
	HeightEq    int64
	HeightGte   int64
	HeightGt    int64
	HeightLte   int64
	HeightLt    int64
	SizeEq      string
	SizeGte     string
	SizeGt      string
	SizeLte     string
	SizeLt      string
	AudioCodec  string
	VideoCodec  string
	Device      string
	CreatedSort string
	Metadata    []string

	interactive bool
	Data        []chunkify.Source
}

func (r *ListCmd) toParams() chunkify.SourceListParams {
	params := chunkify.SourceListParams{}

	if r.DurationEq != "" {
		dur, err := time.ParseDuration(r.DurationEq)
		if err == nil {
			params.DurationEq = dur.Seconds()
		}
	}
	if r.DurationGte != "" {
		dur, err := time.ParseDuration(r.DurationGte)
		if err == nil {
			params.DurationGte = dur.Seconds()
		}
	}
	if r.DurationGt != "" {
		dur, err := time.ParseDuration(r.DurationGt)
		if err == nil {
			params.DurationGt = dur.Seconds()
		}
	}
	if r.DurationLte != "" {
		dur, err := time.ParseDuration(r.DurationLte)
		if err == nil {
			params.DurationLte = dur.Seconds()
		}
	}
	if r.DurationLt != "" {
		dur, err := time.ParseDuration(r.DurationLt)
		if err == nil {
			params.DurationLt = dur.Seconds()
		}
	}
	if r.WidthEq != -1 {
		params.WidthEq = r.WidthEq
	}
	if r.WidthGte != -1 {
		params.WidthGte = r.WidthGte
	}
	if r.WidthGt != -1 {
		params.WidthGt = r.WidthGt
	}
	if r.WidthLte != -1 {
		params.WidthLte = r.WidthLte
	}
	if r.WidthLt != -1 {
		params.WidthLt = r.WidthLt
	}
	if r.HeightEq != -1 {
		params.HeightEq = r.HeightEq
	}
	if r.HeightGte != -1 {
		params.HeightGte = r.HeightGte
	}
	if r.HeightGt != -1 {
		params.HeightGt = r.HeightGt
	}
	if r.HeightLte != -1 {
		params.HeightLte = r.HeightLte
	}
	if r.HeightLt != -1 {
		params.HeightLt = r.HeightLt
	}
	if r.SizeEq != "" {
		bytes, err := formatter.ParseFileSize(r.SizeEq)
		if err == nil {
			params.SizeEq = bytes
		}
	}
	if r.SizeGte != "" {
		bytes, err := formatter.ParseFileSize(r.SizeGte)
		if err == nil {
			params.SizeGte = bytes
		}
	}
	if r.SizeGt != "" {
		bytes, err := formatter.ParseFileSize(r.SizeGt)
		if err == nil {
			params.SizeGt = bytes
		}
	}
	if r.SizeLte != "" {
		bytes, err := formatter.ParseFileSize(r.SizeLte)
		if err == nil {
			params.SizeLte = bytes
		}
	}
	if r.SizeLt != "" {
		bytes, err := formatter.ParseFileSize(r.SizeLt)
		if err == nil {
			params.SizeLt = bytes
		}
	}

	var metadata [][]string
	if len(r.Metadata) > 0 {
		md := []string{}
		for _, m := range r.Metadata {
			md = append(md, strings.Replace(m, "=", ":", -1))
		}
		metadata = [][]string{md}
	}

	params.Metadata = metadata
	params.Offset = r.Offset
	params.Limit = r.Limit
	params.CreatedSort = r.CreatedSort
	params.AudioCodec = r.AudioCodec
	params.VideoCodec = r.VideoCodec
	params.Device = r.Device

	return params
}

func (r *ListCmd) Execute() error {
	params := r.toParams()
	// Convert metadata to the required format

	sources, err := cmd.Config.Client.SourceList(params)
	if err != nil {
		return err
	}

	r.Data = sources.Items

	return nil
}

func (r *ListCmd) View() {
	if cmd.Config.JSON {
		dataBytes, err := json.MarshalIndent(r.Data, "", "  ")
		if err != nil {
			printError(err)
			return
		}
		fmt.Println(string(dataBytes))
		return
	}

	if len(r.Data) == 0 {
		fmt.Println(styles.DefaultText.Render("No source found."))
		return
	}

	fmt.Println(r.sourcesTable())
	if len(r.Data) > 1 {
		fmt.Println(styles.Debug.MarginTop(1).Render(fmt.Sprintf("Total: %d\n", len(r.Data))))
	}
}

func (r *ListCmd) sourcesTable() *table.Table {
	rightCols := []int{2, 3, 6, 8, 9}
	centerCols := []int{4, 5, 7}

	table := table.New().
		BorderRow(true).
		BorderColumn(false).
		BorderStyle(styles.Border).
		Headers("Date", "Id", "Duration", "Size", "WxH", "Video", "Bitrate", "Audio", "Bitrate").
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == 0:
				if slices.Contains(rightCols, col) {
					return styles.Right.Padding(0, 1).Foreground(styles.GrayColor)
				}
				if slices.Contains(centerCols, col) {
					return styles.Center.Padding(0, 1).Foreground(styles.GrayColor)
				}

				return styles.Header.Padding(0, 1)
			case slices.Contains(rightCols, col):
				return styles.Right.Padding(0, 1)
			case slices.Contains(centerCols, col):
				return styles.Center.Padding(0, 1)
			default:
				return styles.TableSpacing
			}
		}).
		Rows(sourcesListToRows(r.Data)...)

	return table
}

func sourcesListToRows(sources []chunkify.Source) [][]string {
	rows := make([][]string, len(sources))
	for i, source := range sources {
		rows[i] = []string{
			source.CreatedAt.Format(time.RFC822),
			styles.Id.Render(source.Id),
			formatter.Duration(source.Duration),
			formatter.Size(source.Size),
			fmt.Sprintf("%dx%d", source.Width, source.Height),
			styles.Important.Render(source.VideoCodec),
			formatter.Bitrate(source.VideoBitrate),
			source.AudioCodec,
			formatter.Bitrate(source.AudioBitrate),
		}
	}
	return rows
}

func newListCmd() *cobra.Command {
	req := ListCmd{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "list all sources",
		Long:  `list all sources`,
		Run: func(_ *cobra.Command, args []string) {
			if err := req.Execute(); err != nil {
				printError(err)
				return
			}
			if !cmd.Config.JSON && req.interactive {
				StartPolling(&req)
			} else {
				req.View()
			}
		},
	}

	cmd.Flags().Int64Var(&req.Offset, "offset", 0, "Offset")
	cmd.Flags().Int64Var(&req.Limit, "limit", 100, "Limit")

	cmd.Flags().StringVar(&req.DurationEq, "duration.eq", "", "Duration Equals")
	cmd.Flags().StringVar(&req.DurationGte, "duration.gte", "", "Duration Greater or Equal")
	cmd.Flags().StringVar(&req.DurationGt, "duration.gt", "", "Duration Greater")
	cmd.Flags().StringVar(&req.DurationLte, "duration.lte", "", "Duration Less or Equal")
	cmd.Flags().StringVar(&req.DurationLt, "duration.lt", "", "Duration Less")

	cmd.Flags().StringVar(&req.CreatedGte, "created.gte", "", "Created Greater or Equal")
	cmd.Flags().StringVar(&req.CreatedLte, "created.lte", "", "Created Less or Equal")

	cmd.Flags().Int64Var(&req.WidthEq, "width.eq", -1, "Width Equals")
	cmd.Flags().Int64Var(&req.WidthGte, "width.gte", -1, "Width Greater or Equal")
	cmd.Flags().Int64Var(&req.WidthGt, "width.gt", -1, "Width Greater")
	cmd.Flags().Int64Var(&req.WidthLte, "width.lte", -1, "Width Less or Equal")
	cmd.Flags().Int64Var(&req.WidthLt, "width.lt", -1, "Width Less")

	cmd.Flags().Int64Var(&req.HeightEq, "height.eq", -1, "Height Equals")
	cmd.Flags().Int64Var(&req.HeightGte, "height.gte", -1, "Height Greater or Equal")
	cmd.Flags().Int64Var(&req.HeightGt, "height.gt", -1, "Height Greater")
	cmd.Flags().Int64Var(&req.HeightLte, "height.lte", -1, "Height Less or Equal")
	cmd.Flags().Int64Var(&req.HeightLt, "height.lt", -1, "Height Less")

	cmd.Flags().StringVar(&req.SizeEq, "size.eq", "", "Size Equals")
	cmd.Flags().StringVar(&req.SizeGte, "size.gte", "", "Size Greater or Equal")
	cmd.Flags().StringVar(&req.SizeGt, "size.gt", "", "Size Greater")
	cmd.Flags().StringVar(&req.SizeLte, "size.lte", "", "Size Less or Equal")
	cmd.Flags().StringVar(&req.SizeLt, "size.lt", "", "Size Less")

	cmd.Flags().StringVar(&req.AudioCodec, "audio-codec", "", "Audio Codec")
	cmd.Flags().StringVar(&req.VideoCodec, "video-codec", "", "Video Codec")
	cmd.Flags().StringVar(&req.Device, "device", "", "Device")

	cmd.Flags().StringVar(&req.CreatedSort, "created.sort", "asc", "Created Sort")

	cmd.Flags().StringArrayVar(&req.Metadata, "metadata", nil, "Metadata")
	cmd.Flags().BoolVarP(&req.interactive, "interactive", "i", false, "Refresh the sources in real time")

	return cmd
}
