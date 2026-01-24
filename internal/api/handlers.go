package api

import (
	"encoding/json"
	"net/http"

	"github.com/Bharat1Rajput/task-management/internal/models"
	"github.com/Bharat1Rajput/task-management/internal/store"
	"github.com/google/uuid"
)

type Handler struct {
	store *store.TaskStore
}

func NewHandler(store *store.TaskStore) *Handler {
	return &Handler{store: store}
}

// CreateTask handles POST /tasks
func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var req models.WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Validate
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//Serialize back to bytes (to store in DB)
	payloadBytes, err := json.Marshal(req)
	if err != nil {
		http.Error(w, "Failed to process payload", http.StatusInternalServerError)
		return
	}

	// Create the Task
	task := &models.Task{
		ID:         uuid.New(),
		Type:       "webhook", // Hardcoded for now
		Payload:    payloadBytes,
		Status:     models.StatusPending,
		Retries:    0,
		MaxRetries: 5, // Default policy
	}

	// Save to DB
	if err := h.store.Create(task); err != nil {
		http.Error(w, "Failed to save task", http.StatusInternalServerError)
		return
	}

	//  Respond immediately (Async!)
	w.WriteHeader(http.StatusAccepted) // 202 Accepted
	response := map[string]string{
		"task_id": task.ID.String(),
		"status":  "queued",
	}
	json.NewEncoder(w).Encode(response)
}
