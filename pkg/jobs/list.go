/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package jobs

import (
	"encoding/json"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/chunkifydev/cli/pkg/api"
	"github.com/chunkifydev/cli/pkg/formatter"
	"github.com/chunkifydev/cli/pkg/styles"
	"github.com/spf13/cobra"
)

type ListCmd struct {
	Id              string
	Offset          int64
	Limit           int64
	CreatedGte      string
	CreatedLte      string
	CreatedSort     string
	SourceId        string
	Status          string
	TemplateName    string
	TemplateVersion string
	Metadata        []string

	interactive bool
	Data        []api.Job
}

func (r *ListCmd) toQueryMap() url.Values {
	query := url.Values{}

	if r.Id != "" {
		query.Add("id", r.Id)
	}

	if r.Offset != -1 {
		query.Add("offset", fmt.Sprintf("%d", r.Offset))
	}
	if r.Limit != -1 {
		query.Add("limit", fmt.Sprintf("%d", r.Limit))
	}

	if r.CreatedGte != "" {
		query.Add("created.gte", r.CreatedGte)
	}
	if r.CreatedLte != "" {
		query.Add("created.lte", r.CreatedLte)
	}

	if r.CreatedSort != "" {
		query.Add("created.sort", r.CreatedSort)
	}

	if r.Status != "" {
		query.Add("status", r.Status)
	}

	if r.SourceId != "" {
		query.Add("source_id", r.SourceId)
	}

	if r.TemplateName != "" {
		query.Add("template_name", r.TemplateName)
	}

	if r.TemplateVersion != "" {
		query.Add("template_version", r.TemplateVersion)
	}

	if len(r.Metadata) > 0 {
		md := []string{}
		for _, metadata := range r.Metadata {
			md = append(md, strings.Replace(metadata, "=", ":", -1))
		}
		query.Add("metadata", strings.Join(md, ","))
	}

	return query
}

func (r *ListCmd) Execute() error {
	apiReq := api.Request{
		Config:      cmd.Config,
		Path:        "/api/jobs",
		Method:      "GET",
		QueryParams: r.toQueryMap(),
	}

	jobs, err := api.ApiRequest[[]api.Job](apiReq)
	if err != nil {
		return err
	}

	r.Data = jobs

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
		Headers("Date", "Id", "Status", "Progress", "Template", "Transcoders", "Speed", "Time", "Billable").
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
			case col == 2:
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

func jobsListToRows(jobs []api.Job) [][]string {
	rows := make([][]string, len(jobs))
	for i, job := range jobs {

		// showing timer in real time while the job is running
		endDate := time.Now()
		if job.Status == "finished" || job.Status == "error" {
			endDate = job.UpdatedAt
		}

		rows[i] = []string{
			job.CreatedAt.Format(time.RFC822),
			styles.Id.Render(job.Id),
			formatter.JobStatus(job.Status),
			fmt.Sprintf("%.f%%", job.Progress),
			fmt.Sprintf("%s/%s", job.Template.Name, job.Template.Version),
			fmt.Sprintf("%d x %s", job.Transcoder.Quantity, job.Transcoder.Type),
			fmt.Sprintf("%.2fx", job.Transcoder.Speed),
			formatter.TimeDiff(job.StartedAt, endDate),
			formatter.Duration(job.BillableTime),
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

	cmd.Flags().StringVar(&req.CreatedGte, "created.gte", "", "Created Greater or Equal")
	cmd.Flags().StringVar(&req.CreatedLte, "created.lte", "", "Created Less or Equal")

	cmd.Flags().StringVar(&req.Status, "status", "", "Job's status: finished, processing, error")

	cmd.Flags().StringVar(&req.SourceId, "source-id", "", "List jobs by source Id")

	cmd.Flags().StringVar(&req.TemplateName, "template-name", "", "List jobs by template name: mp4, hls, jpg, webm")
	cmd.Flags().StringVar(&req.TemplateVersion, "template-version", "", "List jobs by template version: x264-v1, x265-v1, av1-v1, v1")

	cmd.Flags().StringVar(&req.CreatedSort, "created.sort", "asc", "Created Sort: asc (default), desc")

	cmd.Flags().StringArrayVar(&req.Metadata, "metadata", nil, "Metadata")

	cmd.Flags().BoolVarP(&req.interactive, "interactive", "i", false, "Refresh the jobs in real time")

	return cmd
}
