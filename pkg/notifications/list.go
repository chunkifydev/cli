package notifications

import (
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/cli/pkg/formatter"
	"github.com/chunkifydev/cli/pkg/styles"
	"github.com/spf13/cobra"
)

type ListCmd struct {
	Params chunkify.NotificationListParams

	payload     bool
	interactive bool
	Data        []chunkify.Notification
}

func (r *ListCmd) Execute() error {
	notifications, err := cmd.Config.Client.NotificationList(r.Params)
	if err != nil {
		return err
	}

	r.Data = notifications.Items

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
		fmt.Println(styles.DefaultText.Render("No notification found."))
		return
	}

	if r.payload {
		for _, notif := range r.Data {
			fmt.Println(notif.Payload + "\n")
		}
		return
	}

	fmt.Println(r.notificationsTable())
	if len(r.Data) > 1 {
		fmt.Println(styles.Debug.MarginTop(1).Render(fmt.Sprintf("Total: %d\n", len(r.Data))))
	}
}

func (r *ListCmd) notificationsTable() *table.Table {
	rightCols := []int{}
	centerCols := []int{}

	table := table.New().
		BorderRow(true).
		BorderColumn(false).
		BorderStyle(styles.Border).
		Headers("Date", "Id", "Event", "Status").
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
		Rows(notificationsListToRows(r.Data)...)

	return table
}

func notificationsListToRows(notifications []chunkify.Notification) [][]string {
	rows := make([][]string, len(notifications))
	for i, notif := range notifications {
		rows[i] = []string{
			notif.CreatedAt.Format(time.RFC822),
			styles.Id.Render(notif.Id),
			notif.Event,
			formatter.HttpCode(notif.ResponseStatusCode),
		}
	}
	return rows
}

func newListCmd() *cobra.Command {
	req := ListCmd{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "list all sent notifications",
		Long:  `list all sent notifications`,
		Run: func(_ *cobra.Command, args []string) {
			if err := req.Execute(); err != nil {
				printError(err)
				return
			}

			req.View()
		},
	}

	cmd.Flags().BoolVarP(&req.payload, "payload", "p", false, "Return the webhook payload in JSON")

	cmd.Flags().StringVar(&req.Params.ObjectId, "object-id", "", "Return all sent notifications for the object Id")
	cmd.Flags().StringArrayVar(&req.Params.Events, "events", nil, "Return all sent notifications with the given event. Event can be job.completed, job.failed, upload.completed, upload.failed, upload.expired")
	cmd.Flags().StringVar(&req.Params.WebhookId, "webhook-id", "", "Return all sent notifications for a given webhook Id")

	cmd.Flags().Int64Var(&req.Params.Offset, "offset", 0, "Offset")
	cmd.Flags().Int64Var(&req.Params.Limit, "limit", 100, "Limit")

	cmd.Flags().StringVar(&req.Params.CreatedGte, "created.gte", "", "Created Greater or Equal")
	cmd.Flags().StringVar(&req.Params.CreatedLte, "created.lte", "", "Created Less or Equal")

	cmd.Flags().StringVar(&req.Params.CreatedSort, "created.sort", "asc", "Created Sort: asc (default), desc")

	cmd.Flags().BoolVarP(&req.interactive, "interactive", "i", false, "Refresh the jobs in real time")

	return cmd
}
