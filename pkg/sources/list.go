// Package sources provides functionality for managing and interacting with sources
package sources

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/chunkifydev/cli/pkg/flags"
	"github.com/chunkifydev/cli/pkg/formatter"
	"github.com/chunkifydev/cli/pkg/styles"
	"github.com/spf13/cobra"

	chunkify "github.com/chunkifydev/chunkify-go"
)

// ListCmd represents the command for listing sources with various filter options
type ListCmd struct {
	Id          *string  // Filter by source ID
	Offset      *int64   // Starting offset for pagination
	Limit       *int64   // Maximum number of items to return
	DurationEq  *string  // Filter for exact duration match
	DurationGte *string  // Filter for duration greater than or equal
	DurationGt  *string  // Filter for duration greater than
	DurationLte *string  // Filter for duration less than or equal
	DurationLt  *string  // Filter for duration less than
	CreatedGte  *string  // Filter for creation date greater than or equal
	CreatedLte  *string  // Filter for creation date less than or equal
	WidthEq     *int64   // Filter for exact width match
	WidthGte    *int64   // Filter for width greater than or equal
	WidthGt     *int64   // Filter for width greater than
	WidthLte    *int64   // Filter for width less than or equal
	WidthLt     *int64   // Filter for width less than
	HeightEq    *int64   // Filter for exact height match
	HeightGte   *int64   // Filter for height greater than or equal
	HeightGt    *int64   // Filter for height greater than
	HeightLte   *int64   // Filter for height less than or equal
	HeightLt    *int64   // Filter for height less than
	SizeEq      *string  // Filter for exact size match
	SizeGte     *string  // Filter for size greater than or equal
	SizeGt      *string  // Filter for size greater than
	SizeLte     *string  // Filter for size less than or equal
	SizeLt      *string  // Filter for size less than
	AudioCodec  *string  // Filter by audio codec
	VideoCodec  *string  // Filter by video codec
	Device      *string  // Filter by device
	CreatedSort *string  // Sort direction for creation date
	Metadata    []string // Filter by metadata key-value pairs

	interactive bool              // Enable real-time refresh mode
	Data        []chunkify.Source // The list of sources retrieved
}

// toParams converts the ListCmd fields into API parameters
func (r *ListCmd) toParams() chunkify.SourceListParams {
	params := chunkify.SourceListParams{}

	if r.DurationEq != nil {
		dur, err := time.ParseDuration(*r.DurationEq)
		if err == nil {
			seconds := dur.Seconds()
			params.DurationEq = &seconds
		}
	}
	if r.DurationGte != nil {
		dur, err := time.ParseDuration(*r.DurationGte)
		if err == nil {
			seconds := dur.Seconds()
			params.DurationGte = &seconds
		}
	}
	if r.DurationGt != nil {
		dur, err := time.ParseDuration(*r.DurationGt)
		if err == nil {
			seconds := dur.Seconds()
			params.DurationGt = &seconds
		}
	}
	if r.DurationLte != nil {
		dur, err := time.ParseDuration(*r.DurationLte)
		if err == nil {
			seconds := dur.Seconds()
			params.DurationLte = &seconds
		}
	}
	if r.DurationLt != nil {
		dur, err := time.ParseDuration(*r.DurationLt)
		if err == nil {
			seconds := dur.Seconds()
			params.DurationLt = &seconds
		}
	}
	if r.WidthEq != nil {
		width := int64(*r.WidthEq)
		params.WidthEq = &width
	}
	if r.WidthGte != nil {
		width := int64(*r.WidthGte)
		params.WidthGte = &width
	}
	if r.WidthGt != nil {
		width := int64(*r.WidthGt)
		params.WidthGt = &width
	}
	if r.WidthLte != nil {
		width := int64(*r.WidthLte)
		params.WidthLte = &width
	}
	if r.WidthLt != nil {
		width := int64(*r.WidthLt)
		params.WidthLt = &width
	}
	if r.HeightEq != nil {
		height := int64(*r.HeightEq)
		params.HeightEq = &height
	}
	if r.HeightGte != nil {
		height := int64(*r.HeightGte)
		params.HeightGte = &height
	}
	if r.HeightGt != nil {
		height := int64(*r.HeightGt)
		params.HeightGt = &height
	}
	if r.HeightLte != nil {
		height := int64(*r.HeightLte)
		params.HeightLte = &height
	}
	if r.HeightLt != nil {
		height := int64(*r.HeightLt)
		params.HeightLt = &height
	}
	if r.SizeEq != nil {
		bytes, err := formatter.ParseFileSize(*r.SizeEq)
		if err == nil {
			params.SizeEq = &bytes
		}
	}
	if r.SizeGte != nil {
		bytes, err := formatter.ParseFileSize(*r.SizeGte)
		if err == nil {
			params.SizeGte = &bytes
		}
	}
	if r.SizeGt != nil {
		bytes, err := formatter.ParseFileSize(*r.SizeGt)
		if err == nil {
			params.SizeGt = &bytes
		}
	}
	if r.SizeLte != nil {
		bytes, err := formatter.ParseFileSize(*r.SizeLte)
		if err == nil {
			params.SizeLte = &bytes
		}
	}
	if r.SizeLt != nil {
		bytes, err := formatter.ParseFileSize(*r.SizeLt)
		if err == nil {
			params.SizeLt = &bytes
		}
	}

	var metadata map[string]string
	if len(r.Metadata) > 0 {
		metadata = make(map[string]string)
		for _, m := range r.Metadata {
			parts := strings.SplitN(m, "=", 2)
			if len(parts) == 2 {
				metadata[parts[0]] = parts[1]
			}
		}
	}

	params.Id = r.Id
	params.Metadata = metadata
	params.Offset = r.Offset
	params.Limit = r.Limit
	params.CreatedSort = r.CreatedSort
	params.AudioCodec = r.AudioCodec
	params.VideoCodec = r.VideoCodec
	params.Device = r.Device

	return params
}

// Execute retrieves the list of sources from the API
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

// View displays the list of sources in either JSON or table format
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

// sourcesTable creates a formatted table displaying source information
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

// sourcesListToRows converts source data into formatted table rows
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

// newListCmd creates and configures a new cobra command for listing sources
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

	flags.Int64VarPtr(cmd.Flags(), &req.Offset, "offset", 0, "Offset")
	flags.Int64VarPtr(cmd.Flags(), &req.Limit, "limit", 100, "Limit")

	flags.StringVarPtr(cmd.Flags(), &req.Id, "id", "", "Source ID")
	flags.StringVarPtr(cmd.Flags(), &req.DurationEq, "duration.eq", "", "Duration Equals")
	flags.StringVarPtr(cmd.Flags(), &req.DurationGte, "duration.gte", "", "Duration Greater or Equal")
	flags.StringVarPtr(cmd.Flags(), &req.DurationGt, "duration.gt", "", "Duration Greater")
	flags.StringVarPtr(cmd.Flags(), &req.DurationLte, "duration.lte", "", "Duration Less or Equal")
	flags.StringVarPtr(cmd.Flags(), &req.DurationLt, "duration.lt", "", "Duration Less")

	flags.StringVarPtr(cmd.Flags(), &req.CreatedGte, "created.gte", "", "Created Greater or Equal")
	flags.StringVarPtr(cmd.Flags(), &req.CreatedLte, "created.lte", "", "Created Less or Equal")

	flags.Int64VarPtr(cmd.Flags(), &req.WidthEq, "width.eq", -1, "Width Equals")
	flags.Int64VarPtr(cmd.Flags(), &req.WidthGte, "width.gte", -1, "Width Greater or Equal")
	flags.Int64VarPtr(cmd.Flags(), &req.WidthGt, "width.gt", -1, "Width Greater")
	flags.Int64VarPtr(cmd.Flags(), &req.WidthLte, "width.lte", -1, "Width Less or Equal")
	flags.Int64VarPtr(cmd.Flags(), &req.WidthLt, "width.lt", -1, "Width Less")

	flags.Int64VarPtr(cmd.Flags(), &req.HeightEq, "height.eq", -1, "Height Equals")
	flags.Int64VarPtr(cmd.Flags(), &req.HeightGte, "height.gte", -1, "Height Greater or Equal")
	flags.Int64VarPtr(cmd.Flags(), &req.HeightGt, "height.gt", -1, "Height Greater")
	flags.Int64VarPtr(cmd.Flags(), &req.HeightLte, "height.lte", -1, "Height Less or Equal")
	flags.Int64VarPtr(cmd.Flags(), &req.HeightLt, "height.lt", -1, "Height Less")

	flags.StringVarPtr(cmd.Flags(), &req.SizeEq, "size.eq", "", "Size Equals")
	flags.StringVarPtr(cmd.Flags(), &req.SizeGte, "size.gte", "", "Size Greater or Equal")
	flags.StringVarPtr(cmd.Flags(), &req.SizeGt, "size.gt", "", "Size Greater")
	flags.StringVarPtr(cmd.Flags(), &req.SizeLte, "size.lte", "", "Size Less or Equal")
	flags.StringVarPtr(cmd.Flags(), &req.SizeLt, "size.lt", "", "Size Less")

	flags.StringVarPtr(cmd.Flags(), &req.AudioCodec, "audio-codec", "", "Audio Codec")
	flags.StringVarPtr(cmd.Flags(), &req.VideoCodec, "video-codec", "", "Video Codec")
	flags.StringVarPtr(cmd.Flags(), &req.Device, "device", "", "Device")

	flags.StringVarPtr(cmd.Flags(), &req.CreatedSort, "created.sort", "asc", "Created Sort")

	flags.StringArrayVar(cmd.Flags(), &req.Metadata, "metadata", nil, "Metadata")
	cmd.Flags().BoolVarP(&req.interactive, "interactive", "i", false, "Refresh the sources in real time")

	return cmd
}
