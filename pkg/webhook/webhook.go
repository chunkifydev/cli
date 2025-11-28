package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"strings"
	"sync"
	"syscall"
	"time"

	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/cli/pkg/config"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// Command represents the root notifications command and configuration
type Command struct {
	Command *cobra.Command // The root cobra command for notifications
	Config  *config.Config // Configuration for the notifications command
}

type ChunkifyClientInterface interface {
	NotificationList(params chunkify.NotificationListParams) ([]chunkify.Notification, error)
	WebhookCreate(params chunkify.WebhookNewParams) (*chunkify.Webhook, error)
	WebhookDelete(webhookId string) error
}

type ChunkifyClient struct {
	Client *chunkify.Client
}

func (c *ChunkifyClient) NotificationList(params chunkify.NotificationListParams) ([]chunkify.Notification, error) {
	res, err := c.Client.Notifications.List(context.Background(), params)
	if err != nil {
		return nil, err
	}
	return res.Data, nil
}

func (c *ChunkifyClient) WebhookCreate(params chunkify.WebhookNewParams) (*chunkify.Webhook, error) {
	return c.Client.Webhooks.New(context.Background(), params)
}

func (c *ChunkifyClient) WebhookDelete(webhookId string) error {
	return c.Client.Webhooks.Delete(context.Background(), webhookId)
}

// NewCommand creates and configures a new notifications root command
func NewCommand(config *config.Config) *Command {
	var hostname string
	req := WebhookProxy{}

	cmd := &Command{
		Config: config,
		Command: &cobra.Command{
			Use:     "listen",
			Short:   "Forward webhook notifications to local HTTP URL",
			Long:    "Forward webhook notifications to local HTTP URL for local development",
			Example: "chunkify listen --forward-to http://localhost:3000/webhooks/chunkify --webhook-secret <ws_secret>",
			Run: func(_ *cobra.Command, args []string) {
				if hostname == "" {
					hostname, _ = os.Hostname()
					if hostname == "" {
						hostname = uuid.New().String()
					}
				}

				webhookUrl := fmt.Sprintf("http://%s.chunkify.local", hostname)

				req.Client = &ChunkifyClient{Client: config.Client}

				webhook, err := req.createLocaldevWebhook(webhookUrl)
				if err != nil {
					fmt.Printf("Error creating localdev webhook: %s\n", err)
					return
				}

				defer req.deleteLocalDevWebhook(webhook.ID)

				req.WebhookId = webhook.ID

				fmt.Printf("  [%s] Start forwarding to %s\n\n  Events:\n  - %s",
					hostname,
					req.localUrl,
					strings.Join(req.Events, "\n  - "))

				fmt.Printf("\n\n  ────────────────────────────────────────────────\n\n")

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				sigChan := make(chan os.Signal, 1)
				signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
				defer signal.Stop(sigChan)

				// Handle Ctrl+C in a separate goroutine
				go func() {
					sig := <-sigChan
					if sig == os.Interrupt {
						fmt.Println("\nCTRL+C received, stopping...")
					}
					cancel()
				}()

				req.Run(ctx)
			},
		},
	}

	allEvents := []string{
		string(chunkify.NotificationEventJobCompleted),
		string(chunkify.NotificationEventJobFailed),
		string(chunkify.NotificationEventUploadCompleted),
		string(chunkify.NotificationEventUploadFailed),
		string(chunkify.NotificationEventUploadExpired),
	}

	cmd.Command.Flags().StringVar(&req.localUrl, "forward-to", "", "The URL to forward webhook notifications to")
	cmd.Command.Flags().StringSliceVar(&req.Events, "events", allEvents, "Proxy all notifications with the given event. By default, all events are proxied. Event can be job.completed, job.failed, upload.completed, upload.failed, upload.expired")
	cmd.Command.Flags().StringVar(&req.webhookSecret, "webhook-secret", "", "Use your project's webhook secret key to sign the notifications.")
	cmd.Command.Flags().StringVar(&hostname, "hostname", "", "Use the given hostname for the localdev webhook. If not provided, we use the hostname of the machine. It's purely visual, it will just appear on Chunkify")

	cmd.Command.MarkFlagRequired("webhook-secret")

	return cmd
}

// WebhookProxy represents the command for proxying notifications to a local URL
type WebhookProxy struct {
	Client                   ChunkifyClientInterface // Client to use to create the webhook
	localUrl                 string                  // Target URL to proxy notifications to
	webhookSecret            string                  // Key used to sign proxied notifications
	WebhookId                string                  // ID of the webhook receiving notifications
	Events                   []string                // List of event types to proxy
	CreatedGte               time.Time               // Filter for notifications created after this time
	mut                      sync.Mutex              // Mutex for thread-safe access to shared resources
	lastProxiedNotifications []chunkify.Notification // Tracks the 10 last proxied notifications
}

// Update handles incoming messages and updates the model state
func (r *WebhookProxy) Run(ctx context.Context) error {
	notificationsChan := make(chan []chunkify.Notification, 100)
	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ticker.C:
			notifications, err := r.Execute()
			if err != nil {
				fmt.Printf("Error fetching notifications: %s\n", err)
			}
			notificationsChan <- notifications
		case notifications := <-notificationsChan:
			for _, notif := range notifications {
				r.httpProxy(notif)
			}
		case <-ctx.Done():
			return ctx.Err()

		}
	}
}

// toParams converts ProxyCmd fields to NotificationListParams
func (r *WebhookProxy) toParams() chunkify.NotificationListParams {
	limit := int64(10)
	params := chunkify.NotificationListParams{
		WebhookID: chunkify.String(r.WebhookId),
		Limit:     chunkify.Int(limit),
	}

	if len(r.lastProxiedNotifications) > 0 {
		createdGte := r.lastProxiedNotifications[len(r.lastProxiedNotifications)-1].CreatedAt.Unix()
		params.Created.Gte = chunkify.Int(createdGte)
	}

	return params
}

// Execute fetches notifications from the API based on the command parameters
func (r *WebhookProxy) Execute() ([]chunkify.Notification, error) {
	notifications, err := r.Client.NotificationList(r.toParams())
	if err != nil {
		return nil, err
	}

	// we filter the notifications by the given events
	if len(r.Events) > 0 {
		filteredNotifications := []chunkify.Notification{}
		for _, notif := range notifications {
			if slices.Contains(r.Events, string(notif.Event)) {
				filteredNotifications = append(filteredNotifications, notif)
			}
		}
		return filteredNotifications, nil
	}

	return notifications, nil
}

// httpProxy forwards a notification to the configured local URL
func (r *WebhookProxy) httpProxy(notif chunkify.Notification) {
	if !r.shouldProxy(notif) {
		return
	}

	buf := bytes.NewBufferString(notif.Payload)
	req, err := http.NewRequest("POST", r.localUrl, buf)
	if err != nil {
		fmt.Printf("Error creating http request:" + err.Error())
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "chunkify-cli/webhook-proxy")

	timestamp := time.Now()
	signature := generateSignature(notif.Id, timestamp, notif.Payload, r.webhookSecret)
	req.Header.Set("webhook-signature", signature)
	req.Header.Set("webhook-id", notif.Id)
	req.Header.Set("webhook-timestamp", fmt.Sprintf("%d", timestamp.Unix()))

	// Make the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Request error: " + err.Error())
		return
	}

	fmt.Printf("  [%d %s] %s %s (%s)\n",
		resp.StatusCode,
		http.StatusText(resp.StatusCode),
		notif.ID,
		notif.Event,
		notif.ObjectID)
}

// createLocaldevWebhook sets up a webhook for local development
func (r *WebhookProxy) createLocaldevWebhook(webhookUrl string) (chunkify.Webhook, error) {
	enabled := true
	wh, err := r.Client.WebhookCreate(chunkify.WebhookNewParams{URL: webhookUrl, Events: r.Events, Enabled: chunkify.Bool(enabled)})
	if err != nil {
		fmt.Printf("Couldn't create localdev webhook for proxying: %s\n", err)
		return chunkify.Webhook{}, err
	}

	return *wh, nil
}

// deleteLocalDevWebhook removes the local development webhook
func (r *WebhookProxy) deleteLocalDevWebhook(webhookId string) error {
	if err := r.Client.WebhookDelete(webhookId); err != nil {
		fmt.Printf("Couldn't delete localdev webhook. You need to manually delete it. webhookId: %s, error: %s\n", webhookId, err)
		return err
	}

	return nil
}

func (r *WebhookProxy) shouldProxy(notif chunkify.Notification) bool {
	r.mut.Lock()
	defer r.mut.Unlock()

	for _, n := range r.lastProxiedNotifications {
		if n.ID == notif.ID {
			return false
		}
	}

	// Add this notification to the list
	r.lastProxiedNotifications = append(r.lastProxiedNotifications, notif)
	if len(r.lastProxiedNotifications) > 10 {
		r.lastProxiedNotifications = r.lastProxiedNotifications[1:]
	}
	return true
}

// generateSignature creates an HMAC signature for the payload using the secret key
// Following the Standard Webhooks specification: signs "msgId.timestamp.payload"
func generateSignature(id string, timestamp time.Time, payloadString string, secretKey string) string {
	// Remove prefix
	secret := strings.TrimPrefix(secretKey, "whsec_")

	secretBytes, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		fmt.Printf("Error decoding secret: %v\n", err)
		return ""
	}

	// Create the string to sign
	msgToSign := fmt.Sprintf("%s.%d.%s", id, timestamp.Unix(), payloadString)

	h := hmac.New(sha256.New, secretBytes)
	h.Write([]byte(msgToSign))

	// Use base64 encoding instead of hex
	sig := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// Return with v1 prefix
	return "v1," + sig
}
