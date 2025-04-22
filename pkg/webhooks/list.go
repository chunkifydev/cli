package webhooks

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/cli/pkg/formatter"
	"github.com/chunkifydev/cli/pkg/styles"
	"github.com/spf13/cobra"
)

// ListCmd represents the command for listing webhooks
type ListCmd struct {
	Data []chunkify.Webhook // Data contains the list of webhooks
}

// Execute retrieves the list of webhooks
func (r *ListCmd) Execute() error {
	webhooks, err := cmd.Config.Client.WebhookList()
	if err != nil {
		return err
	}

	r.Data = webhooks

	return nil
}

// View displays the webhooks list either in JSON or table format
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

// webhooksTable creates and returns a formatted table of webhooks
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

// webhooksListToRows converts webhook data into string rows for table display
func webhooksListToRows(webhooks []chunkify.Webhook) [][]string {
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

// newListCmd creates and returns a new cobra command for listing webhooks
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
