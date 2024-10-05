package notifications

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/level63/cli/pkg/api"
	"github.com/level63/cli/pkg/formatter"
	"github.com/level63/cli/pkg/styles"
	"github.com/level63/cli/pkg/webhooks"
	"github.com/spf13/cobra"
)

var (
	mut                     sync.Mutex
	lastProxiedNotification *api.Notification
	logs                    []string
	startTime               time.Time
)

type ProxyCmd struct {
	localUrl   string
	secretKey  string
	WebhookId  string
	Events     []string
	CreatedGte time.Time

	Data []api.Notification
}

type model struct {
	cmd *ProxyCmd
	ch  chan []api.Notification
}

type tickMsg time.Time

func listenToNotificationsChan(ch chan []api.Notification) tea.Cmd {
	return func() tea.Msg {
		notifs := <-ch
		return notifs
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tickCmd(), listenToNotificationsChan(m.ch))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			if lastProxiedNotification != nil {
				m.cmd.httpProxy(*lastProxiedNotification)
			}
			return m, tickCmd()
		case "t":
			testNotif := generateTestNotification()
			m.cmd.Data = []api.Notification{testNotif}
			m.cmd.httpProxy(testNotif)
			return m, tickCmd()
		case "v":
			if lastProxiedNotification != nil {
				prettyJson := prettyRenderJSONPayload(lastProxiedNotification.Payload)
				log(styles.Debug.Render(prettyJson) + "\n")
			}
			return m, tickCmd()
		}
	case []api.Notification:
		if len(msg) > 0 {
			for _, notif := range msg {
				go m.cmd.httpProxy(notif)
			}
		}
		m.cmd.Data = msg
		return m, listenToNotificationsChan(m.ch)
	case tickMsg:
		if err := m.cmd.Execute(); err != nil {
			printError(err)
			return m, tickCmd()
		}

		if len(m.cmd.Data) > 0 {
			m.ch <- m.cmd.Data
		}

		return m, tickCmd()
	}
	return m, nil
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) View() string {
	s := strings.Join(logs, "\n")
	s += "\n\n"
	s += styles.Debug.Render("[T] Send a test notification\n[V] View last notification payload\n[R] Replay the last notification\n[Q] Exit\n")

	return s
}

func (r *ProxyCmd) toQueryMap() map[string]string {
	queryMap := map[string]string{}

	if r.WebhookId != "" {
		queryMap["webhook_id"] = r.WebhookId
	}

	if lastProxiedNotification != nil {
		queryMap["created.gte"] = lastProxiedNotification.CreatedAt.Add(1 * time.Second).Format(time.RFC3339)
	} else {
		queryMap["created.gte"] = startTime.Format(time.RFC3339)
	}

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

func (r *ProxyCmd) httpProxy(notif api.Notification) {
	if lastProxiedNotification != nil && lastProxiedNotification.CreatedAt.After(notif.CreatedAt) {
		return
	}

	mut.Lock()
	lastProxiedNotification = &notif
	mut.Unlock()

	buf := bytes.NewBufferString(notif.Payload)
	req, err := http.NewRequest("POST", r.localUrl, buf)
	if err != nil {
		log("Error creating http request:" + err.Error())
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "level63-cli/proxy")

	signature := generateSignature(notif.Payload, r.secretKey)
	req.Header.Set("X-Level63-Signature", signature)

	// Make the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log("Request error: " + err.Error())
		return
	}

	log(fmt.Sprintf("[%s] Proxied notification %s (signature: %s)", formatter.HttpCode(resp.StatusCode), notif.Id, signature))
}

func generateSignature(payloadString string, secretKey string) string {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(payloadString))
	return hex.EncodeToString(h.Sum(nil))
}

func prettyRenderJSONPayload(payload string) string {
	var payloadStruct api.WebhookPayload

	if err := json.Unmarshal([]byte(payload), &payloadStruct); err != nil {
		log("Couldn't not pretty render JSON payload")
		return ""
	}

	prettryBytes, err := json.MarshalIndent(payloadStruct, "", "    ")
	if err != nil {
		log("Couldn't not pretty render JSON payload")
		return ""
	}

	return string(prettryBytes)
}

func createLocaldevWebhook() (api.WebhookWithSecretKey, error) {
	log(fmt.Sprintln("Setting up localdev webhook..."))
	cmd := &webhooks.CreateCmd{Url: "http://localdev", Events: "*", Enabled: true}
	if err := cmd.Execute(); err != nil {
		log(styles.Error.Render(fmt.Sprintf("Couldn't create localdev webhook for proxying: %s", err)))
		return api.WebhookWithSecretKey{}, err
	}

	return cmd.Data, nil
}

func deleteLocalDevWebhook(webhookId string) error {
	cmd := webhooks.DeleteCmd{Id: webhookId}
	if err := cmd.Execute(); err != nil {
		fmt.Printf("Couldn't delete localdev webhook. You need to manually delete it. webhookId: %s, error: %s\n", webhookId, err)
		return err
	}

	return nil
}

func generateTestNotification() api.Notification {
	jobId := uuid.NewString()
	payload := api.WebhookPayload{
		Event: "job.completed",
		Date:  time.Now(),
		Data: api.WebhookPayloadData{
			JobId:    jobId,
			Metadata: map[string]any{"VideoId": uuid.NewString()},
			SourceId: uuid.NewString(),
			Error:    nil,
			Files: []api.File{
				{
					Id:        uuid.NewString(),
					JobId:     jobId,
					Storage:   "aws",
					Path:      "/tmp/test.mp4",
					Size:      1024,
					MimeType:  "video/mp4",
					CreatedAt: time.Now(),
					Url:       "http://localhost:8080/tmp/test.mp4",
				},
			},
		},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log("Couldn't not marshal json")
		return api.Notification{}
	}

	notif := api.Notification{
		Id:                 uuid.NewString(),
		JobId:              jobId,
		Type:               "webhook",
		CreatedAt:          time.Now(),
		Attempts:           1,
		Payload:            string(payloadBytes),
		ResponseStatusCode: 200,
		Event:              "job.completed",
	}

	return notif
}

func log(l string) {
	logs = append(logs, l)
}

func newProxyCmd() *cobra.Command {
	req := ProxyCmd{}

	cmd := &cobra.Command{
		Use:   "proxy",
		Short: "Proxy notifications to local HTTP URL",
		Long:  `Proxy notifications to local HTTP URL`,
		Args:  cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			startTime = time.Now()
			log("level63 proxy\n")
			req.localUrl = args[0]

			webhook, err := createLocaldevWebhook()
			if err != nil {
				return
			}

			defer deleteLocalDevWebhook(webhook.Id)

			req.WebhookId = webhook.Id
			if req.secretKey == "" {
				req.secretKey = webhook.SecretKey
			}
			log(fmt.Sprintf("Secret key: %s\n", req.secretKey))

			log(fmt.Sprintf("Start proxying notifications matching '%s' to %s", strings.Join(req.Events, ","), styles.Important.Render(req.localUrl)))

			ch := make(chan []api.Notification)
			m := model{
				cmd: &req,
				ch:  ch,
			}

			p := tea.NewProgram(m)
			if _, err := p.Run(); err != nil {
				fmt.Printf("Alas, there's been an error: %v", err)
				os.Exit(1)
			}

		},
	}

	cmd.Flags().StringSliceVar(&req.Events, "event", []string{"*"}, "Proxy all notifications with the given event. Event can be *, job.completed")
	cmd.Flags().StringVar(&req.secretKey, "secret-key", "", "Use the given secret key to sign the notifications. If not provided, a random secret key will be used")

	return cmd
}
