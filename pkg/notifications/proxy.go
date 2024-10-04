package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"sync"
	"syscall"
	"time"

	"github.com/level63/cli/pkg/api"
	"github.com/level63/cli/pkg/styles"
	"github.com/level63/cli/pkg/webhooks"
	"github.com/spf13/cobra"
)

var mut sync.Mutex
var lastProxiedTime = time.Now()
var count = 0

type ProxyCmd struct {
	localUrl   string
	WebhookId  string
	FunctionId string
	Type       string
	Events     []string
	CreatedGte time.Time

	Data []api.Notification
}

func (r *ProxyCmd) toQueryMap() map[string]string {
	queryMap := map[string]string{}

	if r.WebhookId != "" {
		queryMap["webhook_id"] = r.WebhookId
	}

	if r.Type != "" {
		queryMap["type"] = r.Type
	}

	queryMap["created.gte"] = lastProxiedTime.Add(1 * time.Second).Format(time.RFC3339)

	return queryMap
}

func (r *ProxyCmd) Execute() error {
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

	// we filter the notifications by the given events
	if len(r.Events) > 0 && !slices.Contains(r.Events, "*") {
		filteredNotifications := []api.Notification{}
		for _, notif := range notifications {
			if slices.Contains(r.Events, notif.Event) {
				filteredNotifications = append(filteredNotifications, notif)
			}
		}
		r.Data = filteredNotifications
		return nil
	}

	r.Data = notifications

	return nil
}

func (r *ProxyCmd) View() {
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
}

func StartPolling(r *ProxyCmd) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	t := time.NewTicker(time.Second * 5)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			if err := r.Execute(); err != nil {
				printError(err)
				return
			}

			if len(r.Data) > 0 {
				for _, notif := range r.Data {
					go r.httpPoxy(notif)
				}
			}
		case sig := <-sigs:
			slog.Info(sig.String())
			return
		}
	}
}

func (r *ProxyCmd) httpPoxy(notif api.Notification) {
	if lastProxiedTime.After(notif.CreatedAt) {
		slog.Debug("Already proxied", "notificationId", notif.Id)
		return
	}

	mut.Lock()
	count++
	lastProxiedTime = notif.CreatedAt
	mut.Unlock()

	slog.Info("Proxying request", "notificationId", notif.Id, "count", count)

	buf := bytes.NewBufferString(notif.Payload)
	req, err := http.NewRequest("POST", r.localUrl, buf)
	if err != nil {
		slog.Error("Error creating http request", "error", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "level63-cli/proxy")

	// Make the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Request error", "error", err)
		return
	}

	slog.Info("Response status code", "code", resp.StatusCode)
}

func createLocaldevWebhook() (api.Webhook, error) {
	slog.Info("Setting up localdev webhook...")
	cmd := &webhooks.CreateCmd{Url: "http://localdev", Events: "*", Enabled: true}
	if err := cmd.Execute(); err != nil {
		slog.Error("Couldn't create localdev webhook for proxying", "error", err)
		return api.Webhook{}, err
	}

	return cmd.Data, nil
}

func deleteLocalDevWebhook(webhookId string) error {
	slog.Info("Cleaning up localdev webhook...")
	cmd := webhooks.DeleteCmd{Id: webhookId}
	if err := cmd.Execute(); err != nil {
		slog.Error("Couldn't delete localdev webhook. You need to manually delete it.", "webhookId", webhookId, "error", err)
		return err
	}

	slog.Info("Done")

	return nil
}

func newProxyCmd() *cobra.Command {
	req := ProxyCmd{}

	cmd := &cobra.Command{
		Use:   "proxy",
		Short: "Proxy notifications to local HTTP URL",
		Long:  `Proxy notifications to local HTTP URL`,
		Args:  cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			req.localUrl = args[0]

			webhook, err := createLocaldevWebhook()
			if err != nil {
				return
			}

			defer deleteLocalDevWebhook(webhook.Id)

			req.WebhookId = webhook.Id

			slog.Info("Start proxying to", "url", req.localUrl)
			StartPolling(&req)
		},
	}

	cmd.Flags().StringVar(&req.Type, "type", "", "Proxy all notifications with the given type. Type can be webhook or function")
	cmd.Flags().StringSliceVar(&req.Events, "event", []string{"*"}, "Proxy all notifications with the given event. Event can be *, job.completed")

	return cmd
}
