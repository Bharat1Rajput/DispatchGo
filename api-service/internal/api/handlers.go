package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/Bharat1Rajput/apiService/internal/models"
	"github.com/Bharat1Rajput/apiService/internal/publisher"
)

type Handler struct {
	publisher publisher.JobPublisher
}

func NewHandler(pub publisher.JobPublisher) *Handler {
	return &Handler{
		publisher: pub,
	}
}

func (h *Handler) CreateJob(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Payload   string `json:"payload"`
		TargetURL string `json:"target_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	job := models.Job{
		ID:        uuid.New().String(),
		Payload:   req.Payload,
		TargetURL: req.TargetURL,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	if err := h.publisher.Publish(r.Context(), job); err != nil {
		http.Error(w, "failed to create job", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"id": job.ID,
	})
}
