package notifications

import (
	"encoding/json"
	"fmt"
	"net/url"
	"slices"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/chunkifydev/cli/pkg/api"
	"github.com/chunkifydev/cli/pkg/formatter"
	"github.com/chunkifydev/cli/pkg/styles"
	"github.com/spf13/cobra"
)

type ListCmd struct {
	Id          string `json:"id"`
	payload     bool
	JobId       string
	WebhookId   string
	FunctionId  string
	Type        string
	Event       string
	Offset      int64
	Limit       int64
	CreatedGte  string
	CreatedLte  string
	CreatedSort string
	interactive bool

	Data []api.Notification
}

func (r *ListCmd) toQueryMap() url.Values {
	query := url.Values{}

	if r.JobId != "" {
		query.Add("job_id", r.JobId)
	}

	if r.WebhookId != "" {
		query.Add("webhook_id", r.WebhookId)
	}

	if r.Type != "" {
		query.Add("type", r.Type)
	}

	if r.Event != "" {
		query.Add("event", r.Event)
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

	return query
}

func (r *ListCmd) Execute() error {
	apiReq := api.Request{
		Config:      cmd.Config,
		Path:        "/api/notifications",
		Method:      "GET",
		QueryParams: r.toQueryMap(),
	}

	notifications, err := api.ApiRequest[[]api.Notification](apiReq)
	if err != nil {
		return err
	}

	r.Data = notifications

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
		Headers("Date", "Id", "Type", "Event", "Status").
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

func notificationsListToRows(notifications []api.Notification) [][]string {
	rows := make([][]string, len(notifications))
	for i, notif := range notifications {
		rows[i] = []string{
			notif.CreatedAt.Format(time.RFC822),
			styles.Id.Render(notif.Id),
			notif.Type,
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

	cmd.Flags().StringVar(&req.JobId, "job-id", "", "Return all sent notifications for the job Id")
	cmd.Flags().StringVar(&req.Type, "type", "", "Return all sent notifications with the given type. Type can be webhook or function")
	cmd.Flags().StringVar(&req.Event, "event", "", "Return all sent notifications with the given event. Event can be *, job.* or job.completed")
	cmd.Flags().StringVar(&req.WebhookId, "webhook-id", "", "Return all sent notifications for a given webhook Id")
	cmd.Flags().StringVar(&req.FunctionId, "function-id", "", "Return all sent notifications for a given function Id")

	cmd.Flags().Int64Var(&req.Offset, "offset", 0, "Offset")
	cmd.Flags().Int64Var(&req.Limit, "limit", 100, "Limit")

	cmd.Flags().StringVar(&req.CreatedGte, "created.gte", "", "Created Greater or Equal")
	cmd.Flags().StringVar(&req.CreatedLte, "created.lte", "", "Created Less or Equal")

	cmd.Flags().StringVar(&req.CreatedSort, "created.sort", "asc", "Created Sort: asc (default), desc")

	cmd.Flags().BoolVarP(&req.interactive, "interactive", "i", false, "Refresh the jobs in real time")

	return cmd
}
