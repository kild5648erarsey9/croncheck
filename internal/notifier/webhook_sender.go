package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebhookSender posts events as JSON to a remote URL.
type WebhookSender struct {
	url    string
	client *http.Client
}

// NewWebhookSender creates a WebhookSender targeting the given URL.
func NewWebhookSender(url string, timeout time.Duration) *WebhookSender {
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	return &WebhookSender{
		url:    url,
		client: &http.Client{Timeout: timeout},
	}
}

// Name returns the sender identifier.
func (w *WebhookSender) Name() string {
	return "webhook"
}

// Send marshals the event to JSON and POSTs it to the webhook URL.
func (w *WebhookSender) Send(event Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	resp, err := w.client.Post(w.url, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("post webhook: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	return nil
}
