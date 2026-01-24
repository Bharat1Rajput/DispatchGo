package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Bharat1Rajput/task-management/internal/models"
	"github.com/Bharat1Rajput/task-management/internal/store"
	"github.com/Bharat1Rajput/task-management/internal/utils"
)

type Processor struct {
	store  *store.TaskStore
	client *http.Client
}

func NewProcessor(store *store.TaskStore) *Processor {
	return &Processor{
		store: store,
		client: &http.Client{
			// If a server takes >10s to reply, we cut it off and count it as a failure.
			Timeout: 10 * time.Second,
		},
	}
}

func (p *Processor) ProcessTask(task *models.Task) error {
	var payload models.WebhookPayload
	if err := json.Unmarshal(task.Payload, &payload); err != nil {
		return fmt.Errorf("invalid payload structure: %w", err)
	}

	if err := payload.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Serialize Data for the Request Body
	reqBody, err := json.Marshal(payload.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal request data: %w", err)
	}

	//Create the HTTP Request
	req, err := http.NewRequest(payload.Method, payload.URL, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create http request: %w", err)
	}

	//Add Headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Go-Dispatch-Engine/1.0")
	req.Header.Set("X-Webhook-ID", task.ID.String())
	req.Header.Set("X-Event", payload.Event)

	// Apply custom headers from payload
	for k, v := range payload.Headers {
		req.Header.Set(k, v)
	}

	// HMAC SECURITY SIGNING
	if payload.Secret != "" {
		signature := utils.GenerateSignature(reqBody, payload.Secret)
		req.Header.Set("X-Signature", signature)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("network delivery failed: %w", err)
	}
	defer resp.Body.Close()

	// We consider 200-299 as success. Everything else is a failure.
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	return fmt.Errorf("recipient returned error status: %d", resp.StatusCode)
}
