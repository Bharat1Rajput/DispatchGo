package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Bharat1Rajput/task-management/internal/models"
	"github.com/google/uuid"
)

type TaskStore struct {
	db *sql.DB
}

// new instance of the repository
func NewTaskStore(db *sql.DB) *TaskStore {
	return &TaskStore{db: db}
}

// Create a new task
func (s *TaskStore) Create(t *models.Task) error {
	// Initialize timestamps in Go so the object stays in sync with the DB
	now := time.Now()
	t.CreatedAt = now
	t.UpdatedAt = now

	query := `
        INSERT INTO tasks (id, type, payload, status, retries, max_retries, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `

	// Executing the query
	_, err := s.db.Exec(query,
		t.ID,
		t.Type,
		t.Payload,
		t.Status,
		t.Retries,
		t.MaxRetries,
		t.CreatedAt,
		t.UpdatedAt,
	)

	return err
}

// picks the next 'PENDING' task and locks it.
func (s *TaskStore) GetNextTask() (*models.Task, error) {
	var t models.Task

	query := `
        UPDATE tasks
        SET status = $1, updated_at = NOW()
        WHERE id = (
            SELECT id FROM tasks
            WHERE status = $2 
            AND retries < max_retries     -- Added: Don't pick up dead tasks
            AND created_at <= NOW() 
            ORDER BY created_at ASC
            LIMIT 1                       -- LIMIT must come before FOR UPDATE
            FOR UPDATE SKIP LOCKED
        )
        RETURNING id, type, payload, status, retries, max_retries, created_at
    `

	// Execute the query. $1 = PROCESSING, $2 = PENDING
	row := s.db.QueryRow(query, models.StatusProcessing, models.StatusPending)

	err := row.Scan(
		&t.ID,
		&t.Type,
		&t.Payload,
		&t.Status,
		&t.Retries,
		&t.MaxRetries,
		&t.CreatedAt,
	)

	// Queue is empty
	if err == sql.ErrNoRows {
		return nil, nil
	}

	//database error
	if err != nil {
		return nil, fmt.Errorf("error fetching next task: %w", err)
	}

	return &t, nil
}

// the worker finishes (Success or Fail)
func (s *TaskStore) UpdateStatus(id uuid.UUID, status models.Status, errorMessage string) error {
	query := `
		UPDATE tasks 
		SET status = $1, error = $2, updated_at = NOW() 
		WHERE id = $3
	`
	_, err := s.db.Exec(query, status, errorMessage, id)
	return err
}

// when a task fails but can be retried
func (s *TaskStore) IncrementRetry(id uuid.UUID, nextRetryTime time.Time) error {
	query := `
        UPDATE tasks
        SET retries = retries + 1, 
            status = $1, 
            updated_at = NOW()
            -- We could also add a 'scheduled_at' column in the future for delayed retries
        WHERE id = $2
    `
	// We set it back to PENDING so it gets picked up again
	_, err := s.db.Exec(query, models.StatusPending, id)
	return err
}
