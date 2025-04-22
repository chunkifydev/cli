package logs

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/cli/pkg/formatter"
	"github.com/chunkifydev/cli/pkg/styles"
	"github.com/spf13/cobra"
)

// ListCmd represents the command for listing and filtering logs
type ListCmd struct {
	Params         chunkify.JobListLogsParams // Parameters for filtering logs
	Id             string                     // Job ID to get logs for
	Levels         []string                   // Log levels to filter by
	IgnoreProgress bool                       // Whether to hide progress logs
	Tail           bool                       // Whether to continuously tail logs
	FfmpegDebug    bool                       // Whether to show ffmpeg debug output
	Data           []chunkify.Log             // The log data retrieved
}

// Execute fetches logs based on the command parameters
func (r *ListCmd) Execute() error {
	logs, err := cmd.Config.Client.JobListLogs(r.Params)
	if err != nil {
		return err
	}

	r.Data = logs

	return nil
}

// View displays the logs in the requested format (JSON or table)
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

// printFfmpegDebug formats and prints ffmpeg debug output with error highlighting
func printFfmpegDebug(stderr string) {
	for _, l := range strings.Split(stderr, "\n") {
		if strings.Contains(l, "Error") || strings.Contains(l, "Invalid") || strings.Contains(l, "Unable") || strings.Contains(l, "Undefined") {
			fmt.Println(styles.Error.Render(l))
		} else {
			fmt.Println(styles.Debug.Render(l))
		}
	}
}

// logsTable creates a formatted table of logs
func (r *ListCmd) logsTable() *table.Table {
	rows := logsListToRows(r)

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

// logsListToRows converts log entries into formatted table rows
func logsListToRows(r *ListCmd) [][]string {
	rows := [][]string{}
	var duration float64
	var firstTime time.Time

	for _, log := range r.Data {
		if r.IgnoreProgress && log.Msg == "Progress" {
			continue
		}

		if len(r.Levels) > 0 {
			if !slices.Contains(r.Levels, log.Level) {
				continue
			}
		}

		attrsStr := String(log.Attributes)

		if log.Level == "DEBUG" && log.Msg == "ffmpeg output" {
			log.Msg = "Check ffmpeg output by running: "
			attrsStr = styles.Hint.Render(fmt.Sprintf("`chunkify logs list %s --ffmpeg-debug`", r.Id))
		} else {
			if len(attrsStr) > 100 {
				attrsStr = attrsStr[:100] + "..."
			}

			attrsStr = styles.Debug.Render(attrsStr)
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

		if strings.HasPrefix(log.Msg, "Starting transcoder#") {
			log.Msg = "🚀 " + log.Msg
		}

		rows = append(rows, []string{
			fmt.Sprintf("%d", len(rows)+1),
			styles.Debug.Render(log.Time.Format(time.RFC3339)) + " " + durationStr,
			formatter.LogService(log.Service),
			formatter.LogLevel(log.Level),
			styles.DefaultText.Render(
				fmt.Sprintf("%s %s", log.Msg, attrsStr),
			),
		})

	}

	return rows
}

// newListCmd creates and configures the logs list command
func newListCmd() *cobra.Command {
	req := ListCmd{}

	cmd := &cobra.Command{
		Use:   "list job-id",
		Short: "list all logs of a job",
		Long:  `list all logs of a job`,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return err
			}

			if req.Params.Service == "transcoder" && req.Params.TranscoderId == 0 {
				err := fmt.Errorf("--id (min 1) is required when --service is set to transcoder")
				return err
			}
			return nil
		},
		Run: func(_ *cobra.Command, args []string) {
			req.Params.JobId = args[0]

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

	cmd.Flags().StringVar(&req.Params.Service, "service", "manager", "Filter by Service name: manager or transcoder (required)")
	cmd.Flags().Int64Var(&req.Params.TranscoderId, "id", 0, "Filter by transcoder number (min 1)")
	cmd.Flags().StringArrayVar(&req.Levels, "level", []string{}, "Filter by log level: INFO, DEBUG, WARN, ERROR")
	cmd.Flags().BoolVar(&req.IgnoreProgress, "ignore-progress", false, "Do not show progress logs")
	cmd.Flags().BoolVar(&req.Tail, "tail", false, "Tail logs")
	cmd.Flags().BoolVar(&req.FfmpegDebug, "ffmpeg-debug", false, "Show ffmpeg stderr for debugging")

	return cmd
}

// AttributesString returns a string representation of a log's attributes
func AttributesString(l chunkify.Log) string {
	if l.Attributes == nil {
		return ""
	}
	return String(l.Attributes)
}

// String converts log attributes to a sorted, formatted string
func String(a chunkify.LogAttrs) string {
	attrs := []string{}
	keys := make([]string, 0, len(a))
	for k := range a {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if a[k] != nil {
			attrs = append(attrs, fmt.Sprintf("%s=%v", k, a[k]))
		}
	}

	return strings.Join(attrs, " ")
}
