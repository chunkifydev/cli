// Package notifications provides functionality for managing and interacting with notifications
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
	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/cli/pkg/formatter"
	"github.com/chunkifydev/cli/pkg/styles"
	"github.com/chunkifydev/cli/pkg/webhooks"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// Global variables used across the proxy functionality
var (
	mut                     sync.Mutex             // Mutex for thread-safe access to shared resources
	lastProxiedNotification *chunkify.Notification // Tracks the most recently proxied notification
	logs                    []string               // Stores log messages
	startTime               time.Time              // Records when proxy started
)

// ProxyCmd represents the command for proxying notifications to a local URL
type ProxyCmd struct {
	localUrl   string                  // Target URL to proxy notifications to
	secretKey  string                  // Key used to sign proxied notifications
	WebhookId  string                  // ID of the webhook receiving notifications
	Events     []string                // List of event types to proxy
	CreatedGte time.Time               // Filter for notifications created after this time
	Data       []chunkify.Notification // The notifications data
}

// WebhookPayload represents the structure of webhook notification payloads
type WebhookPayload struct {
	Event string    `json:"event"` // Type of event
	Date  time.Time `json:"date"`  // When the event occurred
	Data  any       `json:"data"`  // Event-specific data
}

// WebhookJobPayloadData contains the detailed data for a webhook notification
type WebhookJobPayloadData struct {
	JobId    string          `json:"job_id"`    // ID of the associated job
	Status   string          `json:"status"`    // Status of the job
	Metadata any             `json:"metadata"`  // Additional metadata
	SourceId string          `json:"source_id"` // ID of the source
	Error    *string         `json:"error"`     // Error message if any
	Files    []chunkify.File `json:"files"`     // Associated files
}

// WebhookUploadPayloadData contains the detailed data for a webhook notification
type WebhookUploadPayloadData struct {
	UploadId string `json:"upload_id"` // ID of the associated upload
	Status   string `json:"status"`    // Status of the upload
	Metadata any    `json:"metadata"`  // Additional metadata
	SourceId string `json:"source_id"` // ID of the source
}

// model represents the state for the bubbletea TUI
type model struct {
	cmd *ProxyCmd                    // Reference to the proxy command
	ch  chan []chunkify.Notification // Channel for notification updates
}

// tickMsg represents a tick event for periodic updates
type tickMsg time.Time

// listenToNotificationsChan creates a tea.Cmd that listens for notification updates
func listenToNotificationsChan(ch chan []chunkify.Notification) tea.Cmd {
	return func() tea.Msg {
		notifs := <-ch
		return notifs
	}
}

// Init initializes the TUI model
func (m model) Init() tea.Cmd {
	return tea.Batch(tickCmd(), listenToNotificationsChan(m.ch))
}

// Update handles incoming messages and updates the model state
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
			testNotif := generateTestJobNotification()
			m.cmd.Data = []chunkify.Notification{testNotif}
			m.cmd.httpProxy(testNotif)
			return m, tickCmd()
		case "u":
			testNotif := generateTestUploadNotification()
			m.cmd.Data = []chunkify.Notification{testNotif}
			m.cmd.httpProxy(testNotif)
			return m, tickCmd()
		case "v":
			if lastProxiedNotification != nil {
				prettyJson := prettyRenderJSONPayload(lastProxiedNotification.Payload)
				log(styles.Debug.Render(prettyJson) + "\n")
			}
			return m, tickCmd()
		}
	case []chunkify.Notification:
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

// tickCmd creates a tea.Cmd that sends tick events periodically
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// View renders the current state as a string
func (m model) View() string {
	s := strings.Join(logs, "\n")
	s += "\n\n"
	s += styles.Debug.Render("[T] Send a test notification\n[V] View last notification payload\n[R] Replay the last notification\n[Q] Exit\n")

	return s
}

// toParams converts ProxyCmd fields to NotificationListParams
func (r *ProxyCmd) toParams() chunkify.NotificationListParams {
	params := chunkify.NotificationListParams{
		WebhookId: r.WebhookId,
	}

	if lastProxiedNotification != nil {
		params.CreatedGte = lastProxiedNotification.CreatedAt.Add(1 * time.Second).Format(time.RFC3339)
	} else {
		params.CreatedGte = startTime.Format(time.RFC3339)
	}

	return params
}

// Execute fetches notifications from the API based on the command parameters
func (r *ProxyCmd) Execute() error {
	notifications, err := cmd.Config.Client.NotificationList(r.toParams())
	if err != nil {
		return err
	}

	// we filter the notifications by the given events
	if len(r.Events) > 0 && !slices.Contains(r.Events, "*") {
		filteredNotifications := []chunkify.Notification{}
		for _, notif := range notifications.Items {
			if slices.Contains(r.Events, notif.Event) {
				filteredNotifications = append(filteredNotifications, notif)
			}
		}
		r.Data = filteredNotifications
		return nil
	}

	r.Data = notifications.Items

	return nil
}

// httpProxy forwards a notification to the configured local URL
func (r *ProxyCmd) httpProxy(notif chunkify.Notification) {
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
	req.Header.Set("User-Agent", "chunkify-cli/proxy")

	signature := generateSignature(notif.Payload, r.secretKey)
	req.Header.Set("X-Chunkify-Signature", signature)

	// Make the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log("Request error: " + err.Error())
		return
	}

	log(fmt.Sprintf("[%s] Proxied notification %s (signature: %s)", formatter.HttpCode(resp.StatusCode), notif.Id, signature))
}

// generateSignature creates an HMAC signature for the payload using the secret key
func generateSignature(payloadString string, secretKey string) string {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(payloadString))
	return hex.EncodeToString(h.Sum(nil))
}

// prettyRenderJSONPayload formats a JSON payload string for display
func prettyRenderJSONPayload(payload string) string {
	var payloadStruct WebhookPayload

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

// createLocaldevWebhook sets up a webhook for local development
func createLocaldevWebhook() (chunkify.WebhookWithSecretKey, error) {
	log(fmt.Sprintln("Setting up localdev webhook..."))
	enabled := true
	cmd := &webhooks.CreateCmd{Params: chunkify.WebhookCreateParams{Url: "http://localdev", Events: []string{"*"}, Enabled: &enabled}}
	if err := cmd.Execute(); err != nil {
		log(styles.Error.Render(fmt.Sprintf("Couldn't create localdev webhook for proxying: %s", err)))
		return chunkify.WebhookWithSecretKey{}, err
	}

	return cmd.Data, nil
}

// deleteLocalDevWebhook removes the local development webhook
func deleteLocalDevWebhook(webhookId string) error {
	cmd := webhooks.DeleteCmd{Id: webhookId}
	if err := cmd.Execute(); err != nil {
		fmt.Printf("Couldn't delete localdev webhook. You need to manually delete it. webhookId: %s, error: %s\n", webhookId, err)
		return err
	}

	return nil
}

// generateTestJobNotification creates a sample notification for testing
func generateTestJobNotification() chunkify.Notification {
	jobId := uuid.NewString()
	payload := WebhookPayload{
		Event: "job.completed",
		Date:  time.Now(),
		Data: WebhookJobPayloadData{
			JobId:    jobId,
			Metadata: map[string]any{"VideoId": uuid.NewString()},
			SourceId: uuid.NewString(),
			Error:    nil,
			Files: []chunkify.File{
				{
					Id:        uuid.NewString(),
					JobId:     jobId,
					StorageId: "stor_aws_xxx",
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
		return chunkify.Notification{}
	}

	notif := chunkify.Notification{
		Id:                 uuid.NewString(),
		ObjectId:           jobId,
		CreatedAt:          time.Now(),
		Payload:            string(payloadBytes),
		ResponseStatusCode: 200,
		Event:              "job.completed",
	}

	return notif
}

// generateTestUploadNotification creates a sample notification for testing
func generateTestUploadNotification() chunkify.Notification {
	uploadId := uuid.NewString()
	payload := WebhookPayload{
		Event: "upload.completed",
		Date:  time.Now(),
		Data: WebhookUploadPayloadData{
			UploadId: uploadId,
			Status:   "completed",
			Metadata: map[string]any{"VideoId": uuid.NewString()},
			SourceId: uuid.NewString(),
		},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log("Couldn't not marshal json")
		return chunkify.Notification{}
	}

	notif := chunkify.Notification{
		Id:                 uuid.NewString(),
		ObjectId:           uploadId,
		CreatedAt:          time.Now(),
		Payload:            string(payloadBytes),
		ResponseStatusCode: 200,
		Event:              "upload.completed",
	}

	return notif
}

// log adds a message to the log list
func log(l string) {
	logs = append(logs, l)
}

// newProxyCmd creates and configures a new cobra command for proxying notifications
func newProxyCmd() *cobra.Command {
	req := ProxyCmd{}

	cmd := &cobra.Command{
		Use:   "proxy",
		Short: "Proxy notifications to local HTTP URL",
		Long:  `Proxy notifications to local HTTP URL`,
		Args:  cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			startTime = time.Now()
			log("chunkify proxy\n")
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

			ch := make(chan []chunkify.Notification)
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

	cmd.Flags().StringSliceVar(&req.Events, "event", []string{"*"}, "Proxy all notifications with the given event. Event can be *, job.completed, upload.completed")
	cmd.Flags().StringVar(&req.secretKey, "secret-key", "", "Use the given secret key to sign the notifications. If not provided, a random secret key will be used")

	return cmd
}
