package jobs

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/level63/cli/pkg/api"
	"github.com/level63/cli/pkg/formatter"
	"github.com/level63/cli/pkg/styles"
	"github.com/spf13/cobra"
)

type LogsListCmd struct {
	Id         string `json:"id"`
	Service    string
	Levels     []string
	NoProgress bool
	Tail       bool
	Data       []api.Log
}

func (r *LogsListCmd) Execute() error {
	apiReq := api.Request{
		Config:      cmd.Config,
		Path:        fmt.Sprintf("/api/jobs/%s/logs", r.Id),
		Method:      "GET",
		QueryParams: map[string]string{},
	}

	if r.Service != "" {
		filterService := r.Service
		if strings.HasPrefix(r.Service, "transcoder#") {
			filterService = "transcoder"
		}
		apiReq.QueryParams["service"] = filterService
	}

	logs, err := api.ApiRequest[[]api.Log](apiReq)
	if err != nil {
		return err
	}

	r.Data = logs

	return nil
}

func (r *LogsListCmd) View() {
	if cmd.Config.JSON {
		dataBytes, err := json.MarshalIndent(r.Data, "", "  ")
		if err != nil {
			printError(err)
			return
		}
		fmt.Println(string(dataBytes))
		return
	}

	fmt.Println(r.logsTable())
}

func (r *LogsListCmd) logsTable() *table.Table {
	rows := logsListToRows(r.Data, r.Service, r.Levels, r.NoProgress)

	rightCols := []int{0}
	centerCols := []int{2}

	table := table.New().
		BorderRow(false).
		BorderColumn(false).
		Border(lipgloss.HiddenBorder()).
		Headers("Line", "Date", "Service", "Level", "Message").
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
		// if we are tailing, we don't show this message
		if !r.Tail {
			table.Rows([][]string{{"No log found"}}...)
		}
	}

	return table
}

func logsListToRows(logs []api.Log, filterService string, filterLevels []string, noProgress bool) [][]string {
	rows := [][]string{}
	var duration float64
	var firstTime time.Time

	for _, log := range logs {
		if noProgress && log.Msg == "Progress" {
			continue
		}

		if filterService != "" {
			if filterService == "transcoder" && !strings.HasPrefix(log.Service, "transcoder") {
				continue
			} else if strings.HasPrefix(filterService, "transcoder#") && log.Service != filterService {
				continue
			}

			if filterService != "transcoder" && filterService != log.Service {
				continue
			}
		}

		if len(filterLevels) > 0 {
			if !slices.Contains(filterLevels, log.Level) {
				continue
			}
		}

		attrs := []string{}
		keys := make([]string, 0, len(log.LogAttrs))
		for k := range log.LogAttrs {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := log.LogAttrs[k]
			attrs = append(attrs, fmt.Sprintf("%s=%v", k, v))
		}

		attrsStr := strings.Join(attrs, ", ")
		if len(attrsStr) > 100 {
			attrsStr = attrsStr[:100] + "..."
		}

		if firstTime.IsZero() {
			firstTime = log.Time
		}

		duration = log.Time.Sub(firstTime).Seconds()

		durationStr := fmt.Sprintf("%.1fs", duration)
		if duration > 0 {
			durationStr = "+" + durationStr
		}

		if strings.HasPrefix(log.Msg, "Billable time") {
			log.Msg = "✅ " + log.Msg
		}

		rows = append(rows, []string{
			fmt.Sprintf("%d", len(rows)+1),
			styles.Debug.Render(log.Time.Format(time.RFC3339)) + " " + durationStr,
			formatter.LogService(log.Service),
			formatter.LogLevel(log.Level),
			styles.DefaultText.Render(
				fmt.Sprintf("%s %s", log.Msg, styles.Debug.Render(attrsStr)),
			),
		})

	}

	return rows
}

func newLogsListCmd() *cobra.Command {
	req := LogsListCmd{}

	cmd := &cobra.Command{
		Use:   "logs job-id",
		Short: "list all logs of a job",
		Long:  `list all logs of a job`,
		Args:  cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			req.Id = args[0]
			if err := req.Execute(); err != nil {
				printError(err)
				return
			}

			if !cmd.Config.JSON && req.Tail {
				StartTailing(&req)
			} else {
				req.View()
			}
		},
	}

	cmd.Flags().StringVar(&req.Service, "service", "", "Filter by Service name: manager, transcoder, transcoder#1, notifier")
	cmd.Flags().StringArrayVar(&req.Levels, "level", []string{}, "Filter by log level: INFO, DEBUG, WARN, ERROR")
	cmd.Flags().BoolVar(&req.NoProgress, "no-progress", false, "Do not show progress logs")
	cmd.Flags().BoolVar(&req.Tail, "tail", false, "Tail logs")

	return cmd
}
