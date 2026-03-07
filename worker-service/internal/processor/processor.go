package processor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/Bharat1Rajput/workerService/internal/config"
	"github.com/Bharat1Rajput/workerService/internal/model"
	"github.com/Bharat1Rajput/workerService/internal/repository"
)

type Processor struct {
	cfg    *config.Config
	repo   repository.JobRepository
	client *http.Client
	logger *zap.Logger
}

func New(cfg *config.Config, repo repository.JobRepository, logger *zap.Logger) *Processor {
	return &Processor{
		cfg:  cfg,
		repo: repo,
		client: &http.Client{
			Timeout: time.Duration(cfg.HTTPClientTimeoutSec) * time.Second,
		},
		logger: logger,
	}
}

func (p *Processor) ProcessJob(ctx context.Context, job *model.WebhookJob) error {
	if err := p.repo.UpsertProcessing(ctx, job); err != nil {
		return err
	}

	attempt := 0
	for {
		attempt++
		err := p.postWebhook(ctx, job)
		if err == nil {
			if err := p.repo.MarkSuccess(ctx, job.ID); err != nil {
				return err
			}
			return nil
		}

		p.logger.Warn("processor: job failed", zap.String("job_id", job.ID), zap.Error(err))

		retryCount, incErr := p.repo.IncrementRetry(ctx, job.ID)
		if incErr != nil {
			return incErr
		}

		if retryCount >= p.cfg.MaxRetries {
			if err := p.repo.MarkFailed(ctx, job.ID, err.Error()); err != nil {
				return err
			}
			return fmt.Errorf("processor: job %s exhausted retries: %w", job.ID, err)
		}

		backoff := time.Duration(p.cfg.BackoffBaseMS) * time.Millisecond * (1 << (retryCount - 1))
		if backoff > 5*time.Minute {
			backoff = 5 * time.Minute
		}
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (p *Processor) postWebhook(ctx context.Context, job *model.WebhookJob) error {
	payload := job.Payload
	buf := bytes.NewBufferString(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, job.ClientURL, buf)
	if err != nil {
		return fmt.Errorf("processor: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Job-Id", job.ID)

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("processor: do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var respBody map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&respBody)
		return fmt.Errorf("processor: non-2xx status %d", resp.StatusCode)
	}

	return nil
}

