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

// NewTaskStore creates a new instance of the repository
func NewTaskStore(db *sql.DB) *TaskStore {
	return &TaskStore{db: db}
}

// Create saves a new task (Used by the API)
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

// GetNextTask picks the next 'PENDING' task and locks it.
// This is the CRITICAL function for the worker engine.
func (s *TaskStore) GetNextTask() (*models.Task, error) {
	var t models.Task

	// SENIOR LEVEL SQL:
	// 1. SELECT ... FOR UPDATE SKIP LOCKED: Locks the row so other workers skip it.
	// 2. UPDATE ... RETURNING: Changes status to PROCESSING atomically.
	query := `
		UPDATE tasks
		SET status = $1, updated_at = NOW()
		WHERE id = (
			SELECT id FROM tasks
			WHERE status = $2 
            AND created_at < NOW() -- simple check to ensure we don't pick future tasks
			ORDER BY created_at ASC
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		RETURNING id, type, payload, status, retries, max_retries, created_at
	`

	row := s.db.QueryRow(query, models.StatusProcessing, models.StatusPending)

	err := row.Scan(&t.ID, &t.Type, &t.Payload, &t.Status, &t.Retries, &t.MaxRetries, &t.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil // Queue is empty, no work to do
	}
	if err != nil {
		return nil, fmt.Errorf("error fetching next task: %w", err)
	}

	return &t, nil
}

// UpdateStatus is called by the worker when it finishes (Success or Fail)
func (s *TaskStore) UpdateStatus(id uuid.UUID, status models.Status, errorMessage string) error {
	// If failed, we might want to increment retry count (logic can be here or in processor)
	query := `
		UPDATE tasks 
		SET status = $1, error = $2, updated_at = NOW() 
		WHERE id = $3
	`
	_, err := s.db.Exec(query, status, errorMessage, id)
	return err
}

// IncrementRetry is called when a task fails but can be retried
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
