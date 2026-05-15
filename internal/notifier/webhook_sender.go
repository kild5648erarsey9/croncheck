package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/croncheck/internal/alert"
)

// WebhookSender delivers alert events to an HTTP endpoint via POST.
type WebhookSender struct {
	url    string
	client *http.Client
}

// NewWebhookSender creates a WebhookSender that posts events to the given URL.
func NewWebhookSender(url string) *WebhookSender {
	return &WebhookSender{
		url: url,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Name returns the identifier for this sender.
func (w *WebhookSender) Name() string {
	return "webhook"
}

// Send serialises the event as JSON and POSTs it to the configured URL.
// It returns an error if the request fails or the server responds with a
// non-2xx status code.
func (w *WebhookSender) Send(event alert.Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("webhook: marshal event: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, w.url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("webhook: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: unexpected status %d from %s", resp.StatusCode, w.url)
	}

	return nil
}
