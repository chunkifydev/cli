package webhook

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	chunkify "github.com/chunkifydev/chunkify-go"
)

// Mock client for testing
type MockChunkifyClient struct {
	notifications []chunkify.Notification
	webhooks      map[string]chunkify.Webhook
	createError   error
	deleteError   error
	listError     error
}

func (m *MockChunkifyClient) NotificationList(params chunkify.NotificationListParams) ([]chunkify.Notification, error) {
	if m.listError != nil {
		return nil, m.listError
	}
	return m.notifications, nil
}

func (m *MockChunkifyClient) WebhookCreate(params chunkify.WebhookNewParams) (*chunkify.Webhook, error) {
	if m.createError != nil {
		return nil, m.createError
	}
	webhook := chunkify.Webhook{
		ID:      "wh_webhookid",
		URL:     params.URL,
		Events:  params.Events,
		Enabled: params.Enabled.Value,
	}
	return &webhook, nil
}

func (m *MockChunkifyClient) WebhookDelete(webhookId string) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	delete(m.webhooks, webhookId)
	return nil
}

func TestWebhookProxy_ToParams(t *testing.T) {
	proxy := &WebhookProxy{
		WebhookId: "wh_webhookid",
	}

	params := proxy.toParams()

	if params.WebhookID.Value != "wh_webhookid" {
		t.Errorf("Expected WebhookId to be 'wh_webhookid', got %v", params.WebhookID.Value)
	}

	if params.Limit.Value != 10 {
		t.Errorf("Expected Limit to be 10, got %v", params.Limit.Value)
	}
}

func TestWebhookProxy_ToParams_WithLastNotification(t *testing.T) {
	proxy := &WebhookProxy{
		WebhookId: "wh_webhookid",
		lastProxiedNotifications: []chunkify.Notification{
			{
				ID:        "notf_notifid",
				CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			},
		},
	}

	params := proxy.toParams()

	if params.Created.Gte.Value > 0 {
		t.Error("Expected CreatedGte to be set when lastProxiedNotifications is not empty")
	}

	expectedTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC).Unix()
	if params.Created.Gte.Value != expectedTime {
		t.Errorf("Expected CreatedGte to be %d, got %d", expectedTime, params.Created.Gte.Value)
	}
}

func TestWebhookProxy_Execute(t *testing.T) {
	mockClient := &MockChunkifyClient{
		notifications: []chunkify.Notification{
			{ID: "notif_notifid1", Event: "job.completed"},
			{ID: "notif_notifid2", Event: "job.failed"},
		},
	}

	proxy := &WebhookProxy{
		Client: mockClient,
		Events: []string{"job.completed"},
	}

	notifications, err := proxy.Execute()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(notifications) != 1 {
		t.Errorf("Expected 1 notification, got %d", len(notifications))
	}

	if notifications[0].Event != "job.completed" {
		t.Errorf("Expected event 'job.completed', got %s", notifications[0].Event)
	}
}

func TestWebhookProxy_Execute_AllEvents(t *testing.T) {
	mockClient := &MockChunkifyClient{
		notifications: []chunkify.Notification{
			{ID: "notif_notifid1", Event: "job.completed"},
			{ID: "notif_notifid2", Event: "job.failed"},
		},
	}

	proxy := &WebhookProxy{
		Client: mockClient,
		Events: []string{}, // Empty events means all events
	}

	notifications, err := proxy.Execute()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(notifications) != 2 {
		t.Errorf("Expected 2 notifications, got %d", len(notifications))
	}
}

func TestWebhookProxy_ShouldProxy(t *testing.T) {
	proxy := &WebhookProxy{
		lastProxiedNotifications: []chunkify.Notification{},
	}

	notif := chunkify.Notification{ID: "notf_notifid1"}

	// First time should proxy
	if !proxy.shouldProxy(notif) {
		t.Error("Expected shouldProxy to return true for new notification")
	}

	// Second time should not proxy (already seen)
	if proxy.shouldProxy(notif) {
		t.Error("Expected shouldProxy to return false for already seen notification")
	}
}

func TestWebhookProxy_ShouldProxy_MaxNotifications(t *testing.T) {
	proxy := &WebhookProxy{
		lastProxiedNotifications: []chunkify.Notification{},
	}

	// Add 10 notifications
	for i := 0; i < 10; i++ {
		notif := chunkify.Notification{ID: "notif_notifid" + string(rune(i))}
		proxy.shouldProxy(notif)
	}

	// Add 11th notification - should remove the first one
	notif11 := chunkify.Notification{ID: "notif_notifid11"}
	proxy.shouldProxy(notif11)

	if len(proxy.lastProxiedNotifications) != 10 {
		t.Errorf("Expected 10 notifications in list, got %d", len(proxy.lastProxiedNotifications))
	}

	// First notification should be gone
	notif0 := chunkify.Notification{ID: "notif_notifid0"}
	if !proxy.shouldProxy(notif0) {
		t.Error("Expected shouldProxy to return true for notification that was removed from list")
	}
}

func TestGenerateSignature(t *testing.T) {
	payload := "test payload"
	secret := "test secret"

	signature := generateSignature(payload, secret)

	if signature == "" {
		t.Error("Expected signature to be generated")
	}

	// Test that same input produces same signature
	signature2 := generateSignature(payload, secret)
	if signature != signature2 {
		t.Error("Expected same signature for same input")
	}

	// Test that different payload produces different signature
	payload2 := "different payload"
	signature3 := generateSignature(payload2, secret)
	if signature == signature3 {
		t.Error("Expected different signature for different payload")
	}
}

func TestWebhookProxy_HttpProxy(t *testing.T) {
	// Create a test server that captures the request
	var receivedRequest *http.Request
	var receivedBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedRequest = r
		// Read the body
		buf := make([]byte, 1024)
		n, _ := r.Body.Read(buf)
		receivedBody = string(buf[:n])

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	proxy := &WebhookProxy{
		localUrl:                 server.URL,
		webhookSecret:            "test-secret",
		lastProxiedNotifications: []chunkify.Notification{},
	}

	notif := chunkify.Notification{
		ID:       "notf_test123",
		Event:    "job.completed",
		ObjectID: "job_123",
		Payload:  `{"status": "job.completed"}`,
	}

	// Call httpProxy
	proxy.httpProxy(notif)

	// Verify the request was made correctly
	if receivedRequest == nil {
		t.Fatal("Expected HTTP request to be made")
	}

	// Check method
	if receivedRequest.Method != "POST" {
		t.Errorf("Expected POST method, got %s", receivedRequest.Method)
	}

	// Check headers
	contentType := receivedRequest.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got %s", contentType)
	}

	userAgent := receivedRequest.Header.Get("User-Agent")
	if userAgent != "chunkify-cli/webhook-proxy" {
		t.Errorf("Expected User-Agent 'chunkify-cli/webhook-proxy', got %s", userAgent)
	}

	// Check signature header
	signature := receivedRequest.Header.Get("X-Chunkify-Signature")
	if signature == "" {
		t.Error("Expected X-Chunkify-Signature header to be set")
	}

	// Verify signature is correct
	expectedSignature := generateSignature(notif.Payload, "test-secret")
	if signature != expectedSignature {
		t.Errorf("Expected signature %s, got %s", expectedSignature, signature)
	}

	// Check body
	if receivedBody != notif.Payload {
		t.Errorf("Expected body %s, got %s", notif.Payload, receivedBody)
	}
}

func TestWebhookProxy_HttpProxy_ShouldNotProxy(t *testing.T) {
	// Create a test server that should not be called
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("HTTP request should not have been made")
	}))
	defer server.Close()

	proxy := &WebhookProxy{
		localUrl:      server.URL,
		webhookSecret: "test-secret",
		lastProxiedNotifications: []chunkify.Notification{
			{ID: "notf_test123"}, // Already seen notification
		},
	}

	notif := chunkify.Notification{
		ID:       "notf_test123", // Same ID as already seen
		Event:    "job.completed",
		ObjectID: "job_123",
		Payload:  `{"status": "completed"}`,
	}

	// Call httpProxy - should not make HTTP request
	proxy.httpProxy(notif)

	// The test will fail if the HTTP handler is called due to the t.Error() in the handler
}
