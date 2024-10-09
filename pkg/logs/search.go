package logs

import (
	"encoding/json"
	"fmt"
	"net/url"
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

type SearchCmd struct {
	Services       []string
	Levels         []string
	Msg            string
	JobId          string
	IgnoreProgress bool
	FfmpegDebug    bool
	Limit          int64
	CreatedGte     string
	CreatedLte     string
	CreatedSort    string
	TimeGt         string
	Data           []api.Log
}

func (r *SearchCmd) Execute() error {
	apiReq := api.Request{
		Config: cmd.Config,
		Path:   "/api/logs",
		Method: "GET",
	}

	query := url.Values{}

	if len(r.Services) > 0 {
		for _, service := range r.Services {
			query.Add("services", service)
		}
	}

	if len(r.Levels) > 0 {
		for _, level := range r.Levels {
			query.Add("levels", level)
		}
	}

	if r.Msg != "" {
		query.Add("msg", r.Msg)
	}

	if r.IgnoreProgress {
		query.Add("ignore_progress", "true")
	}

	if r.JobId != "" {
		query.Add("job_id", r.JobId)
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

	if r.TimeGt != "" {
		query.Add("time.gt", r.TimeGt)
	}

	if r.CreatedSort != "" {
		query.Add("created.sort", r.CreatedSort)
	}

	apiReq.QueryParams = query

	logs, err := api.ApiRequest[[]api.Log](apiReq)
	if err != nil {
		return err
	}

	r.Data = logs

	return nil
}

func (r *SearchCmd) View() {
	if cmd.Config.JSON {
		dataBytes, err := json.MarshalIndent(r.Data, "", "  ")
		if err != nil {
			printError(err)
			return
		}
		fmt.Println(string(dataBytes))
		return
	}

	if r.FfmpegDebug {
		for _, log := range r.Data {
			if log.Msg == "ffmpeg output" {
				if stderr, ok := log.Attributes["stderr"].(string); ok {
					printFfmpegDebug(stderr)
					fmt.Println()
					return
				}
			}
		}
		return
	}

	fmt.Println(r.logsTable())
}

func (r *SearchCmd) logsTable() *table.Table {
	rows := logsSearchToRows(r)

	rightCols := []int{0}
	centerCols := []int{2}

	table := table.New().
		BorderRow(false).
		BorderColumn(false).
		Border(lipgloss.HiddenBorder()).
		Headers("Line", "Date", "Job Id", "Service", "Level", "Message").
		StyleFunc(func(row, col int) lipgloss.Style {
			gs := lipgloss.Style{}.Padding(0, 1).MarginBottom(1)
			switch {
			case row == 0:
				gs = gs.Foreground(styles.GrayColor)
			case slices.Contains(rightCols, col):
				gs = gs.AlignHorizontal(lipgloss.Right)
			case slices.Contains(centerCols, col):
				gs = gs.AlignHorizontal(lipgloss.Center)
			}

			return gs
		}).
		Rows(rows...)

	if len(rows) == 0 {
		table.Rows([][]string{{"No log found"}}...)
	}

	return table
}

func logsSearchToRows(r *SearchCmd) [][]string {
	rows := [][]string{}

	for _, log := range r.Data {
		attrsStr := log.AttributesString()

		if log.Level == "DEBUG" && log.Msg == "ffmpeg output" {
			log.Msg = "Check ffmpeg output by running: "
			attrsStr = styles.Hint.Render(fmt.Sprintf("`level63 logs list %s --ffmpeg-debug`", log.JobId))
		} else {
			if len(attrsStr) > 100 {
				attrsStr = attrsStr[:100] + "..."
			}

			attrsStr = styles.Debug.Render(attrsStr)
		}

		if strings.HasPrefix(log.Msg, "Billable time") {
			log.Msg = "✅ " + log.Msg
		}

		if strings.HasPrefix(log.Msg, "Starting transcoder#") {
			log.Msg = "🚀 " + log.Msg
		}

		rows = append(rows, []string{
			fmt.Sprintf("%d", len(rows)+1),
			styles.Debug.Render(log.Time.Format(time.RFC3339)),
			log.JobId,
			formatter.LogService(log.Service),
			formatter.LogLevel(log.Level),
			styles.DefaultText.Render(
				fmt.Sprintf("%s %s", log.Msg, attrsStr),
			),
		})

	}

	return rows
}

func newSearchCmd() *cobra.Command {
	req := SearchCmd{}

	cmd := &cobra.Command{
		Use:   "search",
		Short: "search all logs",
		Long:  `search all logs`,
		Run: func(_ *cobra.Command, args []string) {
			if err := req.Execute(); err != nil {
				printError(err)
				return
			}

			req.View()
		},
	}

	cmd.Flags().StringArrayVar(&req.Services, "service", []string{}, "Search by Service name: manager, transcoder")
	cmd.Flags().StringArrayVar(&req.Levels, "level", []string{}, "Search by log level: INFO, DEBUG, WARN, ERROR")
	cmd.Flags().BoolVar(&req.IgnoreProgress, "ignore-progress", false, "Do not show progress logs")
	cmd.Flags().BoolVar(&req.FfmpegDebug, "ffmpeg-debug", false, "Show ffmpeg stderr for debugging")

	cmd.Flags().StringVar(&req.JobId, "job-id", "", "Search by job id")
	cmd.Flags().StringVar(&req.Msg, "message", "", "Search by log message")
	cmd.Flags().Int64Var(&req.Limit, "limit", 100, "Limit")

	cmd.Flags().StringVar(&req.CreatedGte, "created.gte", "", "Created Greater or Equal")
	cmd.Flags().StringVar(&req.CreatedLte, "created.lte", "", "Created Less or Equal")

	cmd.Flags().StringVar(&req.TimeGt, "time.gt", "", "Time Greater")
	cmd.Flags().StringVar(&req.CreatedSort, "created.sort", "asc", "Created Sort: asc (default), desc")

	return cmd
}
