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
	var metadata [][]string
	if len(r.Metadata) > 0 {
		md := []string{}
		for _, m := range r.Metadata {
			md = append(md, strings.Replace(m, "=", ":", -1))
		}
		metadata = [][]string{md}
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

	cmd.Flags().Int64Var(&req.Params.Offset, "offset", 0, "Offset")
	cmd.Flags().Int64Var(&req.Params.Limit, "limit", 100, "Limit")

	cmd.Flags().StringVar(&req.Params.CreatedGte, "created.gte", "", "Created Greater or Equal")
	cmd.Flags().StringVar(&req.Params.CreatedLte, "created.lte", "", "Created Less or Equal")

	cmd.Flags().StringVar(&req.Params.Status, "status", "", "Job's status: finished, processing, error")

	cmd.Flags().StringVar(&req.Params.SourceId, "source-id", "", "List jobs by source Id")

	cmd.Flags().StringVar(&req.Params.FormatName, "format", "", "List jobs by format name: mp4/x264, mp4/x265, mp4/av1, hls/x264, webm/vp9, jpg")
	cmd.Flags().StringVar(&req.Params.CreatedSort, "created.sort", "asc", "Created Sort: asc (default), desc")

	cmd.Flags().StringArrayVar(&req.Metadata, "metadata", nil, "Metadata")

	cmd.Flags().BoolVarP(&req.interactive, "interactive", "i", false, "Refresh the jobs in real time")

	return cmd
}
