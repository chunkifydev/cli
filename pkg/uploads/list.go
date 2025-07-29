package uploads

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/cli/pkg/flags"
	"github.com/chunkifydev/cli/pkg/formatter"
	"github.com/chunkifydev/cli/pkg/styles"
	"github.com/spf13/cobra"
)

// ListCmd represents the command for listing storage configurations
type ListCmd struct {
	Params   chunkify.UploadListParams
	Metadata []string

	Data []chunkify.Upload // List of uploads
}

// Execute retrieves the list of storage configurations
func (r *ListCmd) Execute() error {
	// Convert metadata to the required format
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
	r.Params.Metadata = metadata

	uploads, err := cmd.Config.Client.UploadList(r.Params)
	if err != nil {
		return err
	}

	r.Data = uploads.Items

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
		fmt.Println(styles.DefaultText.Render("No uploads found."))
		return
	}

	fmt.Println(r.uploadsTable())
	if len(r.Data) > 1 {
		fmt.Println(styles.Debug.MarginTop(1).Render(fmt.Sprintf("Total: %d\n", len(r.Data))))
	}
}

// uploadsTable creates and configures a table for displaying upload information
func (r *ListCmd) uploadsTable() *table.Table {
	rightCols := []int{}
	centerCols := []int{3, 4}

	table := table.New().
		BorderRow(true).
		BorderColumn(false).
		BorderStyle(styles.Border).
		Headers("Date", "Upload ID", "Source ID", "Status", "Expires At").
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
		Rows(uploadsListToRows(r.Data)...)

	return table
}

// uploadsListToRows converts upload data into string rows for table display
func uploadsListToRows(uploads []chunkify.Upload) [][]string {
	rows := make([][]string, len(uploads))
	for i, upload := range uploads {
		sourceId := "N/A"
		if upload.SourceId != nil {
			sourceId = *upload.SourceId
		}
		rows[i] = []string{
			upload.CreatedAt.Format(time.RFC822),
			upload.Id,
			sourceId,
			formatter.UploadStatus(upload.Status),
			upload.ExpiresAt.Format(time.RFC822),
		}
	}
	return rows
}

// newListCmd creates and configures a new cobra command for listing upload configurations
func newListCmd() *cobra.Command {
	req := ListCmd{}

	// Initialize nil pointers
	req.Params.Offset = new(int64)
	req.Params.Limit = new(int64)
	req.Params.CreatedGte = new(string)
	req.Params.CreatedLte = new(string)
	req.Params.Status = new(string)
	req.Params.SourceId = new(string)
	req.Params.CreatedSort = new(string)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "list all uploads",
		Long:  `list all uploads`,
		Run: func(_ *cobra.Command, args []string) {
			if err := req.Execute(); err != nil {
				printError(err)
				return
			}
			req.View()
		},
	}

	flags.Int64VarPtr(cmd.Flags(), &req.Params.Offset, "offset", 0, "Offset")
	flags.Int64VarPtr(cmd.Flags(), &req.Params.Limit, "limit", 100, "Limit")

	flags.StringVarPtr(cmd.Flags(), &req.Params.CreatedGte, "created.gte", "", "Created Greater or Equal")
	flags.StringVarPtr(cmd.Flags(), &req.Params.CreatedLte, "created.lte", "", "Created Less or Equal")

	flags.StringVarPtr(cmd.Flags(), &req.Params.Status, "status", "", "Upload's status: completed, waiting, failed, expired")

	flags.StringVarPtr(cmd.Flags(), &req.Params.SourceId, "source-id", "", "List uploads by source Id")

	flags.StringVarPtr(cmd.Flags(), &req.Params.CreatedSort, "created.sort", "asc", "Created Sort: asc (default), desc")

	flags.StringArrayVar(cmd.Flags(), &req.Metadata, "metadata", nil, "Metadata")

	return cmd
}
