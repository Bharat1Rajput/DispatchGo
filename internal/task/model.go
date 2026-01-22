package task

import (
	"encoding/json"
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
	Retries    int             `json:"retries"`
	MaxRetries int             `json:"max_retries"`
	Error      string          `json:"error,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

type CreateTaskRequest struct {
	Type       string          `json:"type"`
	Payload    json.RawMessage `json:"payload"`
	MaxRetries int             `json:"max_retries,omitempty"`
}

func (r *CreateTaskRequest) Validate() error {
	if r.Type == "" {
		return ErrInvalidTaskType
	}
	if len(r.Payload) == 0 {
		return ErrInvalidPayload
	}
	if r.MaxRetries == 0 {
		r.MaxRetries = 3 // Default
	}
	return nil
}
