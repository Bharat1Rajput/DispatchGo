package models

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusPending    Status = "PENDING"
	StatusRunning    Status = "RUNNING"
	StatusCompleted  Status = "COMPLETED"
	StatusFailed     Status = "FAILED"
	StatusProcessing Status = "PROCESSING"
)

type Task struct {
	ID         uuid.UUID       `json:"id"`
	Type       string          `json:"type"`
	Payload    json.RawMessage `json:"payload"`
	Status     Status          `json:"status"`
	Error      string          `json:"error,omitempty"`
	Retries    int             `json:"retries"`
	MaxRetries int             `json:"max_retries"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

// WebhookPayload defines the structure of the JSON inside Task.Payload.
// This is what the Worker will unpack to know WHERE to send the data.
type WebhookPayload struct {
	URL     string                 `json:"url"`               // Destination URL (e.g. https://api.client.com/hooks)
	Method  string                 `json:"method,omitempty"`  // GET, POST, PUT (Default: POST)
	Event   string                 `json:"event"`             // Event Name (e.g. "payment.success")
	Data    map[string]interface{} `json:"data"`              // The actual business data to send
	Headers map[string]string      `json:"headers,omitempty"` // Custom headers (e.g. Authorization)
	Secret  string                 `json:"secret,omitempty"`  // Used for HMAC signing
}

// Validate ensures the payload has the minimum required fields
func (w *WebhookPayload) Validate() error {
	if w.URL == "" {
		return errors.New("webhook URL is required")
	}
	if w.Event == "" {
		return errors.New("event type is required")
	}
	if w.Data == nil {
		return errors.New("data payload is required")
	}
	return nil
}
