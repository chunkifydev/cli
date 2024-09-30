package jobs

import (
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/level63/cli/pkg/api"
	"github.com/level63/cli/pkg/formatter"
	"github.com/level63/cli/pkg/styles"
	"github.com/spf13/cobra"
)

type FunctionsListCmd struct {
	Id      string `json:"id"`
	payload bool
	Data    []api.JobFunction
}

func (r *FunctionsListCmd) Execute() error {
	apiReq := api.Request{
		Config: cmd.Config,
		Path:   fmt.Sprintf("/api/jobs/%s/functions", r.Id),
		Method: "GET",
	}

	functions, err := api.ApiRequest[[]api.JobFunction](apiReq)
	if err != nil {
		return err
	}

	r.Data = functions

	return nil
}

func (r *FunctionsListCmd) View() {
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
		fmt.Println(styles.DefaultText.Render("No function triggered for this job."))
		return
	}

	if r.payload {
		for _, function := range r.Data {
			fmt.Printf("[%s] %s (%s)\n\n%s\n\n",
				styles.DefaultText.Render(function.CreatedAt.Format(time.RFC3339)),
				function.Function.Description,
				formatter.HttpCode(function.ResponseStatusCode),
				styles.Debug.Render(function.Payload),
			)
		}
		return
	}

	fmt.Println(r.filesTable())
	if len(r.Data) > 1 {
		fmt.Println(styles.Debug.MarginTop(1).Render(fmt.Sprintf("Total: %d\n", len(r.Data))))
	}
}

func (r *FunctionsListCmd) filesTable() *table.Table {
	rightCols := []int{}
	centerCols := []int{}

	table := table.New().
		BorderRow(true).
		BorderColumn(false).
		BorderStyle(styles.Border).
		Headers("Date", "Id", "Name", "Status").
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
		Rows(functionsListToRows(r.Data)...)

	return table
}

func functionsListToRows(functions []api.JobFunction) [][]string {
	rows := make([][]string, len(functions))
	for i, function := range functions {
		rows[i] = []string{
			function.CreatedAt.Format(time.RFC822),
			styles.Id.Render(function.Id),
			function.Function.Description,
			formatter.HttpCode(function.ResponseStatusCode),
		}
	}
	return rows
}

func newFunctionsListCmd() *cobra.Command {
	req := FunctionsListCmd{}

	cmd := &cobra.Command{
		Use:   "functions job-id",
		Short: "list all functions triggered",
		Long:  `list all functions triggered after the given job succeeded / failed.`,
		Args:  cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			req.Id = args[0]
			if err := req.Execute(); err != nil {
				printError(err)
				return
			}

			req.View()
		},
	}

	cmd.Flags().BoolVarP(&req.payload, "payload", "p", false, "Return the webhook payload in JSON")

	return cmd
}
