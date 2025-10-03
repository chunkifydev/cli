// Package notifications provides functionality for managing and interacting with notifications
package webhooks

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
	"github.com/chunkifydev/cli/pkg/config"
	"github.com/chunkifydev/cli/pkg/flags"
	"github.com/chunkifydev/cli/pkg/formatter"
	"github.com/chunkifydev/cli/pkg/styles"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// Global variables used across the proxy functionality
var (
	mut                      sync.Mutex              // Mutex for thread-safe access to shared resources
	lastProxiedNotifications []chunkify.Notification // Tracks the 10 last proxied notifications
	logs                     []string                // Stores log messages
)

// Command represents the root notifications command and configuration
type Command struct {
	Command *cobra.Command // The root cobra command for notifications
	Config  *config.Config // Configuration for the notifications command
}

// cmd is a package-level variable holding the current Command instance
var cmd *Command

// NewCommand creates and configures a new notifications root command
func NewCommand(config *config.Config) *Command {
	cmd = &Command{
		Config: config,
		Command: &cobra.Command{
			Use:     "webhooks",
			Short:   "Proxy webhooks to a local URL for local development",
			Long:    "Proxy webhooks to a local URL for local development",
			Example: "chunkify webhooks proxy http://localhost:3000/webhook",
		}}

	// Add all subcommands
	cmd.Command.AddCommand(newProxyCmd()) // Proxy notifications to a local URL
	return cmd
}

// printError formats and prints an error message using the error style
func printError(err error) {
	fmt.Println(styles.Error.Render(err.Error()))
}

// ProxyCmd represents the command for proxying notifications to a local URL
type ProxyCmd struct {
	localUrl      string                  // Target URL to proxy notifications to
	webhookSecret string                  // Key used to sign proxied notifications
	WebhookId     string                  // ID of the webhook receiving notifications
	Events        []string                // List of event types to proxy
	CreatedGte    time.Time               // Filter for notifications created after this time
	Data          []chunkify.Notification // The notifications data
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
			if len(lastProxiedNotifications) > 0 {
				m.cmd.httpProxy(lastProxiedNotifications[len(lastProxiedNotifications)-1])
			}
			return m, tickCmd()
		case "v":
			if len(lastProxiedNotifications) > 0 {
				prettyJson := prettyRenderJSONPayload(lastProxiedNotifications[len(lastProxiedNotifications)-1].Payload)
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
	s += styles.Debug.Render("[R] Replay the last notification\n[V] View last notification payload\n[Q] Exit\n")

	return s
}

// toParams converts ProxyCmd fields to NotificationListParams
func (r *ProxyCmd) toParams() chunkify.NotificationListParams {
	limit := int64(10)
	params := chunkify.NotificationListParams{
		WebhookId: &r.WebhookId,
		Limit:     &limit,
	}

	if len(lastProxiedNotifications) > 0 {
		createdGte := lastProxiedNotifications[len(lastProxiedNotifications)-1].CreatedAt.Format(time.RFC3339)
		params.CreatedGte = &createdGte
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
	if len(r.Events) > 0 {
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

func shouldProxy(notif chunkify.Notification) bool {
	mut.Lock()
	defer mut.Unlock()

	for _, n := range lastProxiedNotifications {
		if n.Id == notif.Id {
			return false
		}
	}

	// Add this notification to the list
	lastProxiedNotifications = append(lastProxiedNotifications, notif)
	if len(lastProxiedNotifications) > 10 {
		lastProxiedNotifications = lastProxiedNotifications[1:]
	}
	return true
}

// httpProxy forwards a notification to the configured local URL
func (r *ProxyCmd) httpProxy(notif chunkify.Notification) {
	if !shouldProxy(notif) {
		return
	}

	buf := bytes.NewBufferString(notif.Payload)
	req, err := http.NewRequest("POST", r.localUrl, buf)
	if err != nil {
		log("Error creating http request:" + err.Error())
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "chunkify-cli/proxy")

	signature := generateSignature(notif.Payload, r.webhookSecret)
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
	var payloadStruct chunkify.NotificationPayload

	if err := json.Unmarshal([]byte(payload), &payloadStruct); err != nil {
		log("Couldn't not pretty render JSON payload: " + err.Error())
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
func createLocaldevWebhook(webhookUrl string) (chunkify.Webhook, error) {
	log(fmt.Sprintf("Setting up localdev webhook for %s", webhookUrl))

	enabled := true
	wh, err := cmd.Config.Client.WebhookCreate(chunkify.WebhookCreateParams{Url: webhookUrl, Events: chunkify.NotificationEventsAll, Enabled: &enabled})
	if err != nil {
		log(styles.Error.Render(fmt.Sprintf("Couldn't create localdev webhook for proxying: %s", err)))
		return chunkify.Webhook{}, err
	}

	return wh, nil
}

// deleteLocalDevWebhook removes the local development webhook
func deleteLocalDevWebhook(webhookId string) error {
	if err := cmd.Config.Client.WebhookDelete(webhookId); err != nil {
		fmt.Printf("Couldn't delete localdev webhook. You need to manually delete it. webhookId: %s, error: %s\n", webhookId, err)
		return err
	}

	return nil
}

// log adds a message to the log list
func log(l string) {
	logs = append(logs, l)
}

// newProxyCmd creates and configures a new cobra command for proxying notifications
func newProxyCmd() *cobra.Command {
	var hostname string
	req := ProxyCmd{}

	cmd := &cobra.Command{
		Use:   "proxy",
		Short: "Proxy notifications to local HTTP URL",
		Long:  `Proxy notifications to local HTTP URL`,
		Args:  cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			log("chunkify proxy\n")
			req.localUrl = args[0]

			if hostname == "" {
				hostname, _ = os.Hostname()
				if hostname == "" {
					hostname = uuid.New().String()
				}
			}

			webhook, err := createLocaldevWebhook(fmt.Sprintf("http://%s.chunkify.local", hostname))
			if err != nil {
				return
			}

			defer deleteLocalDevWebhook(webhook.Id)

			req.WebhookId = webhook.Id
			log(fmt.Sprintf("Secret key: %s\n", req.webhookSecret))

			log(fmt.Sprintf("Start proxying notifications to %s\n\nEvents:\n%s", styles.Important.Render(req.localUrl), strings.Join(req.Events, "\n")))

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

	flags.StringSliceVar(cmd.Flags(), &req.Events, "events", chunkify.NotificationEventsAll, "Proxy all notifications with the given event. By default, all events are proxied. Event can be job.completed, job.failed, upload.completed, upload.failed, upload.expired")
	flags.StringVar(cmd.Flags(), &req.webhookSecret, "webhook-secret", "", "Use your project's webhook secret key to sign the notifications.")
	flags.StringVar(cmd.Flags(), &hostname, "hostname", "", "Use the given hostname for the localdev webhook. If not provided, we use the hostname of the machine. It's purely visual, it will just appear on Chunkify")

	cmd.MarkFlagRequired("webhook-secret")

	return cmd
}
