package model

import "time"

type JobStatus string

const (
	StatusPending    JobStatus = "pending"
	StatusProcessing JobStatus = "processing"
	StatusSuccess    JobStatus = "success"
	StatusFailed     JobStatus = "failed"
)

type WebhookJob struct {
	ID         string    `json:"id"`
	Payload    string    `json:"payload"`
	ClientURL  string    `json:"client_url"`
	Status     JobStatus `json:"status"`
	Error      string    `json:"error"`
	RetryCount int       `json:"retry_count"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

