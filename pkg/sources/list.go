package sources

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/level63/cli/pkg/api"
	"github.com/level63/cli/pkg/formatter"
	"github.com/level63/cli/pkg/styles"
	"github.com/spf13/cobra"
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
	Data        []api.Source
}

func (r *ListCmd) toQueryMap() map[string]string {
	queryMap := map[string]string{}

	if r.Offset != -1 {
		queryMap["offset"] = fmt.Sprintf("%d", r.Offset)
	}
	if r.Limit != -1 {
		queryMap["limit"] = fmt.Sprintf("%d", r.Limit)
	}
	if r.DurationEq != "" {
		dur, err := time.ParseDuration(r.DurationEq)
		if err == nil {
			queryMap["duration.eq"] = fmt.Sprintf("%f", dur.Seconds())
		}
	}
	if r.DurationGte != "" {
		dur, err := time.ParseDuration(r.DurationGte)
		if err == nil {
			queryMap["duration.gte"] = fmt.Sprintf("%f", dur.Seconds())
		}
	}
	if r.DurationGt != "" {
		dur, err := time.ParseDuration(r.DurationGt)
		if err == nil {
			queryMap["duration.gt"] = fmt.Sprintf("%f", dur.Seconds())
		}
	}
	if r.DurationLte != "" {
		dur, err := time.ParseDuration(r.DurationLte)
		if err == nil {
			queryMap["duration.lte"] = fmt.Sprintf("%f", dur.Seconds())
		}
	}
	if r.DurationLt != "" {
		dur, err := time.ParseDuration(r.DurationLt)
		if err == nil {
			queryMap["duration.lt"] = fmt.Sprintf("%f", dur.Seconds())
		}
	}
	if r.CreatedGte != "" {
		queryMap["created.gte"] = r.CreatedGte
	}
	if r.CreatedLte != "" {
		queryMap["created.lte"] = r.CreatedLte
	}
	if r.WidthEq != -1 {
		queryMap["width.eq"] = fmt.Sprintf("%d", r.WidthEq)
	}
	if r.WidthGte != -1 {
		queryMap["width.gte"] = fmt.Sprintf("%d", r.WidthGte)
	}
	if r.WidthGt != -1 {
		queryMap["width.gt"] = fmt.Sprintf("%d", r.WidthGt)
	}
	if r.WidthLte != -1 {
		queryMap["width.lte"] = fmt.Sprintf("%d", r.WidthLte)
	}
	if r.WidthLt != -1 {
		queryMap["width.lt"] = fmt.Sprintf("%d", r.WidthLt)
	}
	if r.HeightEq != -1 {
		queryMap["height.eq"] = fmt.Sprintf("%d", r.HeightEq)
	}
	if r.HeightGte != -1 {
		queryMap["height.gte"] = fmt.Sprintf("%d", r.HeightGte)
	}
	if r.HeightGt != -1 {
		queryMap["height.gt"] = fmt.Sprintf("%d", r.HeightGt)
	}
	if r.HeightLte != -1 {
		queryMap["height.lte"] = fmt.Sprintf("%d", r.HeightLte)
	}
	if r.HeightLt != -1 {
		queryMap["height.lt"] = fmt.Sprintf("%d", r.HeightLt)
	}
	if r.SizeEq != "" {
		bytes, err := formatter.ParseFileSize(r.SizeEq)
		if err == nil {
			queryMap["size.eq"] = fmt.Sprintf("%d", bytes)
		}
	}
	if r.SizeGte != "" {
		bytes, err := formatter.ParseFileSize(r.SizeGte)
		if err == nil {
			queryMap["size.gte"] = fmt.Sprintf("%d", bytes)
		}
	}
	if r.SizeGt != "" {
		bytes, err := formatter.ParseFileSize(r.SizeGt)
		if err == nil {
			queryMap["size.gt"] = fmt.Sprintf("%d", bytes)
		}
	}
	if r.SizeLte != "" {
		bytes, err := formatter.ParseFileSize(r.SizeLte)
		if err == nil {
			queryMap["size.lte"] = fmt.Sprintf("%d", bytes)
		}
	}
	if r.SizeLt != "" {
		bytes, err := formatter.ParseFileSize(r.SizeLt)
		if err == nil {
			queryMap["size.lt"] = fmt.Sprintf("%d", bytes)
		}
	}
	if r.AudioCodec != "" {
		queryMap["audio_codec"] = r.AudioCodec
	}
	if r.VideoCodec != "" {
		queryMap["video_codec"] = r.VideoCodec
	}
	if r.Device != "" {
		queryMap["device"] = r.Device
	}
	if r.CreatedSort != "" {
		queryMap["created.sort"] = r.CreatedSort
	}
	if len(r.Metadata) > 0 {
		md := []string{}
		for _, metadata := range r.Metadata {
			md = append(md, strings.Replace(metadata, "=", ":", -1))
		}
		queryMap["metadata"] = strings.Join(md, ",")
	}

	return queryMap
}

func (r *ListCmd) Execute() error {
	apiReq := api.Request{
		Config:      cmd.Config,
		Path:        "/api/sources",
		Method:      "GET",
		QueryParams: r.toQueryMap(),
	}

	sources, err := api.ApiRequest[[]api.Source](apiReq)
	if err != nil {
		return err
	}

	r.Data = sources

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
		Headers("Date", "Id", "Duration", "Size", "WxH", "Video", "Bitrate", "Audio", "Bitrate", "jobs").
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

func sourcesListToRows(sources []api.Source) [][]string {
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
			fmt.Sprintf("%d", len(source.Jobs)),
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
