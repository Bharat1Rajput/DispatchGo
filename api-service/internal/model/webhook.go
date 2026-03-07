package model

import "time"

type WebhookStatus string

const (
	StatusPending  WebhookStatus = "pending"
	StatusSuccess  WebhookStatus = "success"
	StatusFailed   WebhookStatus = "failed"
	StatusUnknown  WebhookStatus = "unknown"
)

type WebhookJob struct {
	ID        string        `json:"id"`
	Payload   string        `json:"payload"`
	ClientURL string        `json:"client_url"`
	Status    WebhookStatus `json:"status"`
	Error     string        `json:"error,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
}

