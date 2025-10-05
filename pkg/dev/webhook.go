// Package notifications provides functionality for managing and interacting with notifications
package dev

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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
	"github.com/chunkifydev/cli/pkg/formatter"
	"github.com/chunkifydev/cli/pkg/styles"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// Command represents the root notifications command and configuration
type Command struct {
	Command *cobra.Command // The root cobra command for notifications
	Config  *config.Config // Configuration for the notifications command
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

				req.Client = config.Client

				webhook, err := req.createLocaldevWebhook(webhookUrl)
				if err != nil {
					fmt.Printf("Error creating localdev webhook: %s\n", err)
					return
				}

				defer req.deleteLocalDevWebhook(webhook.Id)

				req.WebhookId = webhook.Id

				fmt.Printf("  [%s] Start forwarding to %s\n\n  Events:\n  - %s\n\n  ────────────────────────────────────────────────\n\n", hostname, styles.Important.Render(req.localUrl), strings.Join(req.Events, "\n  - "))

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

	cmd.Command.Flags().StringVar(&req.localUrl, "forward-to", "", "The URL to forward webhook notifications to")
	cmd.Command.Flags().StringSliceVar(&req.Events, "events", chunkify.NotificationEventsAll, "Proxy all notifications with the given event. By default, all events are proxied. Event can be job.completed, job.failed, upload.completed, upload.failed, upload.expired")
	cmd.Command.Flags().StringVar(&req.webhookSecret, "webhook-secret", "", "Use your project's webhook secret key to sign the notifications.")
	cmd.Command.Flags().StringVar(&hostname, "hostname", "", "Use the given hostname for the localdev webhook. If not provided, we use the hostname of the machine. It's purely visual, it will just appear on Chunkify")

	cmd.Command.MarkFlagRequired("webhook-secret")

	return cmd
}

// WebhookProxy represents the command for proxying notifications to a local URL
type WebhookProxy struct {
	Client                   *chunkify.Client        // Client to use to create the webhook
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
		WebhookId: &r.WebhookId,
		Limit:     &limit,
	}

	if len(r.lastProxiedNotifications) > 0 {
		createdGte := r.lastProxiedNotifications[len(r.lastProxiedNotifications)-1].CreatedAt.Format(time.RFC3339)
		params.CreatedGte = &createdGte
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
		for _, notif := range notifications.Items {
			if slices.Contains(r.Events, notif.Event) {
				filteredNotifications = append(filteredNotifications, notif)
			}
		}
		return filteredNotifications, nil
	}

	return notifications.Items, nil
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

	signature := generateSignature(notif.Payload, r.webhookSecret)
	req.Header.Set("X-Chunkify-Signature", signature)

	// Make the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Request error: " + err.Error())
		return
	}

	fmt.Printf("  [%s] %s %s (%s)\n", formatter.HttpCode(resp.StatusCode), notif.Id, notif.Event, notif.ObjectId)
}

// createLocaldevWebhook sets up a webhook for local development
func (r *WebhookProxy) createLocaldevWebhook(webhookUrl string) (chunkify.Webhook, error) {
	enabled := true
	wh, err := r.Client.WebhookCreate(chunkify.WebhookCreateParams{Url: webhookUrl, Events: chunkify.NotificationEventsAll, Enabled: &enabled})
	if err != nil {
		fmt.Println(styles.Error.Render(fmt.Sprintf("Couldn't create localdev webhook for proxying: %s", err)))
		return chunkify.Webhook{}, err
	}

	return wh, nil
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
		if n.Id == notif.Id {
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
func generateSignature(payloadString string, secretKey string) string {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(payloadString))
	return hex.EncodeToString(h.Sum(nil))
}
