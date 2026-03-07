package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Bharat1Rajput/workerService/internal/model"
)

type JobRepository interface {
	UpsertProcessing(ctx context.Context, job *model.WebhookJob) error
	MarkSuccess(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id string, errMsg string) error
	IncrementRetry(ctx context.Context, id string) (int, error)
}

type PostgresJobRepository struct {
	db *sql.DB
}

func NewPostgresJobRepository(db *sql.DB) *PostgresJobRepository {
	return &PostgresJobRepository{db: db}
}

func (r *PostgresJobRepository) UpsertProcessing(ctx context.Context, job *model.WebhookJob) error {
	const query = `
		INSERT INTO webhook_jobs (
			id, payload, client_url, status, error, retry_count, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (id) DO NOTHING
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		job.ID,
		job.Payload,
		job.ClientURL,
		model.StatusProcessing,
		"",
		job.RetryCount,
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("repository.job: upsert processing: %w", err)
	}
	return nil
}

func (r *PostgresJobRepository) MarkSuccess(ctx context.Context, id string) error {
	const query = `
		UPDATE webhook_jobs
		SET status = $1,
		    error = '',
		    updated_at = $2
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, model.StatusSuccess, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("repository.job: mark success: %w", err)
	}
	return nil
}

func (r *PostgresJobRepository) MarkFailed(ctx context.Context, id string, errMsg string) error {
	const query = `
		UPDATE webhook_jobs
		SET status = $1,
		    error = $2,
		    updated_at = $3
		WHERE id = $4
	`
	_, err := r.db.ExecContext(ctx, query, model.StatusFailed, errMsg, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("repository.job: mark failed: %w", err)
	}
	return nil
}

func (r *PostgresJobRepository) IncrementRetry(ctx context.Context, id string) (int, error) {
	const query = `
		UPDATE webhook_jobs
		SET retry_count = retry_count + 1,
		    updated_at = $1
		WHERE id = $2
		RETURNING retry_count
	`
	var retryCount int
	if err := r.db.QueryRowContext(ctx, query, time.Now().UTC(), id).Scan(&retryCount); err != nil {
		return 0, fmt.Errorf("repository.job: increment retry: %w", err)
	}
	return retryCount, nil
}

