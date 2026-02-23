package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Bharat1Rajput/task-management/internal/models"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) InsertJob(ctx context.Context, job models.Job) error {
	query := `
		INSERT INTO jobs (id, payload, target_url, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := s.db.ExecContext(
		ctx,
		query,
		job.ID,
		job.Payload,
		job.TargetURL,
		job.Status,
		job.CreatedAt,
	)

	return err
}

func (s *Store) UpdateJobStatus(ctx context.Context, id string, status string) error {
	query := `
		UPDATE jobs
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := s.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("job not found")
	}

	return nil
}

func (s *Store) GetJobByID(ctx context.Context, id string) (*models.Job, error) {
	query := `
		SELECT id, payload, target_url, status, created_at
		FROM jobs
		WHERE id = $1
	`

	row := s.db.QueryRowContext(ctx, query, id)

	job := &models.Job{}
	err := row.Scan(
		&job.ID,
		&job.Payload,
		&job.TargetURL,
		&job.Status,
		&job.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return job, nil
}
