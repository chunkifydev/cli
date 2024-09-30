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

type WebhooksListCmd struct {
	Id      string `json:"id"`
	payload bool
	Data    []api.JobWebhook
}

func (r *WebhooksListCmd) Execute() error {
	apiReq := api.Request{
		Config: cmd.Config,
		Path:   fmt.Sprintf("/api/jobs/%s/webhooks", r.Id),
		Method: "GET",
	}

	webhooks, err := api.ApiRequest[[]api.JobWebhook](apiReq)
	if err != nil {
		return err
	}

	r.Data = webhooks

	return nil
}

func (r *WebhooksListCmd) View() {
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
		fmt.Println(styles.DefaultText.Render("No webhook triggered for this job."))
		return
	}

	if r.payload {
		for _, webhook := range r.Data {
			fmt.Printf("[%s] %s (%s)\n\n%s\n\n",
				styles.DefaultText.Render(webhook.CreatedAt.Format(time.RFC3339)),
				webhook.Webhook.Url,
				formatter.HttpCode(webhook.ResponseStatusCode),
				styles.Debug.Render(webhook.Payload),
			)
		}
		return
	}

	fmt.Println(r.filesTable())
	if len(r.Data) > 1 {
		fmt.Println(styles.Debug.MarginTop(1).Render(fmt.Sprintf("Total: %d\n", len(r.Data))))
	}
}

func (r *WebhooksListCmd) filesTable() *table.Table {
	rightCols := []int{}
	centerCols := []int{}

	table := table.New().
		BorderRow(true).
		BorderColumn(false).
		BorderStyle(styles.Border).
		Headers("Date", "Id", "URL", "Status").
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

func webhooksListToRows(webhooks []api.JobWebhook) [][]string {
	rows := make([][]string, len(webhooks))
	for i, webhook := range webhooks {
		rows[i] = []string{
			webhook.CreatedAt.Format(time.RFC822),
			styles.Id.Render(webhook.Id),
			webhook.Webhook.Url,
			formatter.HttpCode(webhook.ResponseStatusCode),
		}
	}
	return rows
}

func newWebhooksListCmd() *cobra.Command {
	req := WebhooksListCmd{}

	cmd := &cobra.Command{
		Use:   "webhooks job-id",
		Short: "list all webhooks triggered",
		Long:  `list all webhooks triggered after the given job succeeded / failed.`,
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
