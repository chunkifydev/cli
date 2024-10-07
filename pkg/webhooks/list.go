package webhooks

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/level63/cli/pkg/api"
	"github.com/level63/cli/pkg/formatter"
	"github.com/level63/cli/pkg/styles"
	"github.com/spf13/cobra"
)

type ListCmd struct {
	Data []api.Webhook
}

func (r *ListCmd) Execute() error {
	apiReq := api.Request{
		Config: cmd.Config,
		Path:   "/api/webhooks",
		Method: "GET",
	}

	webhooks, err := api.ApiRequest[[]api.Webhook](apiReq)
	if err != nil {
		return err
	}

	r.Data = webhooks

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
		fmt.Println(styles.DefaultText.Render("No webhook found."))
		return
	}

	fmt.Println(r.webhooksTable())
	if len(r.Data) > 1 {
		fmt.Println(styles.Debug.MarginTop(1).Render(fmt.Sprintf("Total: %d\n", len(r.Data))))
	}
}

func (r *ListCmd) webhooksTable() *table.Table {
	rightCols := []int{}
	centerCols := []int{2, 3}

	table := table.New().
		BorderRow(true).
		BorderColumn(false).
		BorderStyle(styles.Border).
		Headers("Id", "Url", "Events", "Active").
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
		Rows(webhooksListToRows(r.Data)...)

	return table
}

func webhooksListToRows(webhooks []api.Webhook) [][]string {
	rows := make([][]string, len(webhooks))
	for i, webhook := range webhooks {
		rows[i] = []string{
			styles.Id.Render(webhook.Id),
			webhook.Url,
			strings.Join(webhook.Events, ","),
			formatter.Bool(webhook.Enabled),
		}
	}
	return rows
}

func newListCmd() *cobra.Command {
	req := ListCmd{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "list all webhooks",
		Long:  `list all webhooks`,
		Run: func(_ *cobra.Command, args []string) {
			if err := req.Execute(); err != nil {
				printError(err)
				return
			}
			req.View()
		},
	}

	return cmd
}
