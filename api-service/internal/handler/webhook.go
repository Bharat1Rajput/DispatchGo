package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Bharat1Rajput/apiService/internal/broker"
	"github.com/Bharat1Rajput/apiService/internal/config"
	"github.com/Bharat1Rajput/apiService/internal/model"
)

type WebhookHandler struct {
	cfg       *config.Config
	publisher broker.Publisher
	logger    *zap.Logger
}

func NewWebhookHandler(cfg *config.Config, pub broker.Publisher, logger *zap.Logger) *WebhookHandler {
	return &WebhookHandler{
		cfg:       cfg,
		publisher: pub,
		logger:    logger,
	}
}

func (h *WebhookHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/webhooks", h.handlePostWebhook)
	return r
}

type webhookRequest struct {
	Payload   string `json:"payload"`
	ClientURL string `json:"client_url"`
}

type webhookResponse struct {
	JobID   string `json:"job_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (h *WebhookHandler) handlePostWebhook(w http.ResponseWriter, r *http.Request) {
	var req webhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	if req.Payload == "" {
		http.Error(w, "payload is required", http.StatusBadRequest)
		return
	}

	u, err := url.Parse(req.ClientURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		http.Error(w, "client_url must be a valid http or https URL", http.StatusBadRequest)
		return
	}

	job := model.WebhookJob{
		ID:        uuid.New().String(),
		Payload:   req.Payload,
		ClientURL: req.ClientURL,
		Status:    model.StatusPending,
		CreatedAt: time.Now().UTC(),
	}

	body, err := json.Marshal(job)
	if err != nil {
		h.logger.Error("handler.webhook: marshal job", zap.Error(err))
		http.Error(w, "failed to enqueue job", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := h.publisher.Publish(ctx, h.cfg.RabbitRoutingKey, body); err != nil {
		h.logger.Error("handler.webhook: publish job", zap.Error(err))
		http.Error(w, "failed to enqueue job", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	resp := webhookResponse{
		JobID:   job.ID,
		Status:  string(job.Status),
		Message: "job accepted and queued for processing",
	}
	_ = json.NewEncoder(w).Encode(resp)
}


