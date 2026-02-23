package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/Bharat1Rajput/workerService/internal/models"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) FetchPendingJobs(ctx context.Context, limit int) ([]models.Job, error) {
	query := `
		SELECT id, payload, target_url, status, retry_count, created_at
		FROM jobs
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT $1
	`

	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []models.Job

	for rows.Next() {
		var job models.Job
		if err := rows.Scan(
			&job.ID,
			&job.Payload,
			&job.TargetURL,
			&job.Status,
			&job.RetryCount,
			&job.CreatedAt,
		); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (s *Store) UpdateJobStatus(ctx context.Context, id string, status string) error {
	query := `
		UPDATE jobs
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := s.db.ExecContext(ctx, query, status, time.Now(), id)
	return err
}
