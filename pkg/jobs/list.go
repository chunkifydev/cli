package jobs

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

type ListCmd struct {
	Metadata []string

	Params chunkify.JobListParams

	interactive bool
	Data        []chunkify.Job
}

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

	jobs, err := cmd.Config.Client.JobList(r.Params)
	if err != nil {
		return err
	}

	r.Data = jobs.Items
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
		fmt.Println(styles.DefaultText.Render("No job found."))
		return
	}

	fmt.Println(r.jobsTable())

	if len(r.Data) > 1 {
		fmt.Println(styles.Debug.MarginTop(1).Render(fmt.Sprintf("Total: %d\n", len(r.Data))))
	}
}

func (r *ListCmd) jobsTable() *table.Table {
	rightCols := []int{5, 7, 8}
	centerCols := []int{2, 3, 4, 6}

	rows := jobsListToRows(r.Data)

	table := table.New().
		BorderRow(true).
		BorderColumn(false).
		BorderStyle(styles.Border).
		Headers("Date", "Id", "HLS", "Status", "Progress", "Template", "Transcoders", "Time").
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
			case col == 3:
				return styles.Center.Padding(0, 1).Width(14)
			case slices.Contains(rightCols, col):
				return styles.Right.Padding(0, 1)
			case slices.Contains(centerCols, col):
				return styles.Center.Padding(0, 1)
			default:
				return styles.TableSpacing
			}
		}).
		Rows(rows...)

	return table
}

func jobsListToRows(jobs []chunkify.Job) [][]string {
	rows := make([][]string, len(jobs))
	for i, job := range jobs {

		// showing timer in real time while the job is running
		endDate := time.Now()
		if job.Status == "finished" || job.Status == "error" {
			endDate = job.UpdatedAt
		}

		hlsManifestId := ""
		if job.HlsManifestId != nil {
			hlsManifestId = *job.HlsManifestId
		}

		rows[i] = []string{
			job.CreatedAt.Format(time.RFC822),
			styles.Id.Render(job.Id),
			styles.Id.Render(hlsManifestId),
			formatter.JobStatus(job.Status),
			fmt.Sprintf("%.f%%", job.Progress),
			job.Format.Name,
			fmt.Sprintf("%d x %s", job.Transcoder.Quantity, job.Transcoder.Type),
			formatter.TimeDiff(job.StartedAt, endDate),
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

	flags.Int64VarPtr(cmd.Flags(), &req.Params.Offset, "offset", 0, "Offset")
	flags.Int64VarPtr(cmd.Flags(), &req.Params.Limit, "limit", 100, "Limit")

	flags.StringVarPtr(cmd.Flags(), &req.Params.CreatedGte, "created.gte", "", "Created Greater or Equal")
	flags.StringVarPtr(cmd.Flags(), &req.Params.CreatedLte, "created.lte", "", "Created Less or Equal")

	flags.StringVarPtr(cmd.Flags(), &req.Params.Status, "status", "", "Job's status: finished, processing, error")

	flags.StringVarPtr(cmd.Flags(), &req.Params.SourceId, "source-id", "", "List jobs by source Id")

	flags.StringVarPtr(cmd.Flags(), &req.Params.FormatName, "format", "", "List jobs by format name: mp4/x264, mp4/x265, mp4/av1, hls/x264, webm/vp9, jpg")
	flags.StringVarPtr(cmd.Flags(), &req.Params.CreatedSort, "created.sort", "asc", "Created Sort: asc (default), desc")

	flags.StringArrayVar(cmd.Flags(), &req.Metadata, "metadata", nil, "Metadata")

	flags.BoolVarP(cmd.Flags(), &req.interactive, "interactive", "i", false, "Refresh the jobs in real time")

	return cmd
}
