package files

import (
	"encoding/json"
	"fmt"
	"path"
	"slices"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/cli/pkg/formatter"
	"github.com/chunkifydev/cli/pkg/styles"
	"github.com/spf13/cobra"
)

// ListCmd represents the command for listing files with various filter options
type ListCmd struct {
	Id          string // Filter by file ID
	Offset      int64  // Starting offset for pagination
	Limit       int64  // Maximum number of items to return
	DurationEq  string // Filter for exact duration match
	DurationGte string // Filter for duration greater than or equal
	DurationGt  string // Filter for duration greater than
	DurationLte string // Filter for duration less than or equal
	DurationLt  string // Filter for duration less than
	CreatedGte  string // Filter for creation date greater than or equal
	CreatedLte  string // Filter for creation date less than or equal
	WidthEq     int64  // Filter for exact width match
	WidthGte    int64  // Filter for width greater than or equal
	WidthGt     int64  // Filter for width greater than
	WidthLte    int64  // Filter for width less than or equal
	WidthLt     int64  // Filter for width less than
	HeightEq    int64  // Filter for exact height match
	HeightGte   int64  // Filter for height greater than or equal
	HeightGt    int64  // Filter for height greater than
	HeightLte   int64  // Filter for height less than or equal
	HeightLt    int64  // Filter for height less than
	SizeEq      string // Filter for exact size match
	SizeGte     string // Filter for size greater than or equal
	SizeGt      string // Filter for size greater than
	SizeLte     string // Filter for size less than or equal
	SizeLt      string // Filter for size less than
	AudioCodec  string // Filter by audio codec
	VideoCodec  string // Filter by video codec
	CreatedSort string // Sort direction for creation date
	MimeType    string // Filter by mime type
	JobId       string // Filter by job id
	StorageId   string // Filter by storage id
	PathEq      string // Filter by exact path
	PathILike   string // Filter by matching path

	Data []chunkify.File // The list of sources retrieved
}

// toParams converts the ListCmd fields into API parameters
func (r *ListCmd) toParams() chunkify.FileListParams {
	params := chunkify.FileListParams{}

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
	params.Id = r.Id
	params.Offset = r.Offset
	params.Limit = r.Limit
	params.CreatedSort = r.CreatedSort
	params.AudioCodec = r.AudioCodec
	params.VideoCodec = r.VideoCodec
	params.MimeType = r.MimeType
	params.JobId = r.JobId
	params.StorageId = r.StorageId
	params.PathEq = r.PathEq
	params.PathILike = r.PathILike

	return params
}

// Execute retrieves the list of storage configurations
func (r *ListCmd) Execute() error {
	files, err := cmd.Config.Client.FileList(r.toParams())
	if err != nil {
		return err
	}

	r.Data = files.Items

	return nil
}

// View displays the list of upload configurations
// If JSON output is enabled, it prints the data in JSON format
// Otherwise, it displays the data in a formatted table
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
		fmt.Println(styles.DefaultText.Render("No files found."))
		return
	}

	fmt.Println(r.filesTable())
	if len(r.Data) > 1 {
		fmt.Println(styles.Debug.MarginTop(1).Render(fmt.Sprintf("Total: %d\n", len(r.Data))))
	}
}

// filesTable creates a formatted table displaying file information
func (r *ListCmd) filesTable() *table.Table {
	rightCols := []int{3, 6, 8}
	centerCols := []int{4, 5, 7}

	table := table.New().
		BorderRow(true).
		BorderColumn(false).
		BorderStyle(styles.Border).
		Headers("Date", "File ID", "Filename", "Duration", "Size", "WxH", "Video", "Bitrate", "Audio", "Bitrate", "Job ID").
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
		Rows(filesListToRows(r.Data)...)

	return table
}

// filesListToRows converts file data into formatted table rows
func filesListToRows(files []chunkify.File) [][]string {
	rows := make([][]string, len(files))
	for i, file := range files {
		rows[i] = []string{
			file.CreatedAt.Format(time.RFC822),
			styles.Id.Render(file.Id),
			path.Base(file.Path),
			formatter.Duration(file.Duration),
			formatter.Size(file.Size),
			fmt.Sprintf("%dx%d", file.Width, file.Height),
			styles.Important.Render(file.VideoCodec),
			formatter.Bitrate(file.VideoBitrate),
			file.AudioCodec,
			formatter.Bitrate(file.AudioBitrate),
			styles.Id.Render(file.JobId),
		}
	}
	return rows
}

// newListCmd creates and configures a new cobra command for listing file configurations
func newListCmd() *cobra.Command {
	req := ListCmd{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "list all files",
		Long:  `list all files`,
		Run: func(_ *cobra.Command, args []string) {
			if err := req.Execute(); err != nil {
				printError(err)
				return
			}
			req.View()
		},
	}

	cmd.Flags().Int64Var(&req.Offset, "offset", 0, "Offset")
	cmd.Flags().Int64Var(&req.Limit, "limit", 100, "Limit")

	cmd.Flags().StringVar(&req.Id, "id", "", "File ID")

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

	cmd.Flags().StringVar(&req.CreatedSort, "created.sort", "asc", "Created Sort")

	cmd.Flags().StringVar(&req.MimeType, "mime-type", "", "Mime Type")
	cmd.Flags().StringVar(&req.JobId, "job-id", "", "Job ID")
	cmd.Flags().StringVar(&req.StorageId, "storage-id", "", "Storage ID")
	cmd.Flags().StringVar(&req.PathEq, "path.eq", "", "Path Equals")
	cmd.Flags().StringVar(&req.PathILike, "path.ilike", "", "Path iLike")

	return cmd
}
