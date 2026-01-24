package store

import (
	"database/sql"
	"testing"
	"time"

	"github.com/Bharat1Rajput/task-management/internal/models"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Test Create method using sqlmock
func TestTaskStore_Create(t *testing.T) {
	// Create a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	// Create the store with the mock DB
	store := NewTaskStore(db)

	//Create test data
	taskID := uuid.New()
	task := &models.Task{
		ID:         taskID,
		Type:       "webhook",
		Payload:    []byte(`{"url": "https://example.com/webhook", "method": "POST", "data": {"key": "value"}}`),
		Status:     models.StatusPending,
		Retries:    0,
		MaxRetries: 5,
	}

	//EXPECT the database to receive
	mock.ExpectExec("INSERT INTO tasks").
		WithArgs(
			task.ID,
			task.Type,
			task.Payload,
			task.Status,
			task.Retries,
			task.MaxRetries,
			sqlmock.AnyArg(), // created_at
			sqlmock.AnyArg(), // updated_at
		).
		WillReturnResult(sqlmock.NewResult(1, 1)) // (lastInsertId, rowsAffected)

	err = store.Create(task)
	assert.NoError(t, err)

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// Test Create with database error
func TestTaskStore_Create_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	store := NewTaskStore(db)

	task := &models.Task{
		ID:         uuid.New(),
		Type:       "send_email",
		Payload:    []byte(`{"to": "user@example.com"}`),
		Status:     models.StatusPending,
		Retries:    0,
		MaxRetries: 5,
	}

	// Simulate a database error
	mock.ExpectExec("INSERT INTO tasks").
		WillReturnError(sql.ErrConnDone) // Connection closed error

	// Execute
	err = store.Create(task)

	// Assert we got an error
	assert.Error(t, err)

	// Verify expectations
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// Test GetNextTask method
func TestTaskStore_GetNextTask(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	store := NewTaskStore(db)

	// Expected task data
	taskID := uuid.New()
	expectedTask := &models.Task{
		ID:         taskID,
		Type:       "send_email",
		Payload:    []byte(`{"to": "user@example.com"}`),
		Status:     models.StatusProcessing,
		Retries:    0,
		MaxRetries: 5,
		CreatedAt:  time.Now(),
	}

	// Define what rows will be returned
	rows := sqlmock.NewRows([]string{
		"id", "type", "payload", "status", "retries", "max_retries", "created_at",
	}).AddRow(
		expectedTask.ID,
		expectedTask.Type,
		expectedTask.Payload,
		expectedTask.Status,
		expectedTask.Retries,
		expectedTask.MaxRetries,
		expectedTask.CreatedAt,
	)

	// Expect the UPDATE query
	mock.ExpectQuery("UPDATE tasks").
		WithArgs(models.StatusProcessing, models.StatusPending).
		WillReturnRows(rows)

	// Execute
	task, err := store.GetNextTask()

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, expectedTask.ID, task.ID)
	assert.Equal(t, expectedTask.Type, task.Type)
	assert.Equal(t, models.StatusProcessing, task.Status)

	// Verify expectations
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// Test GetNextTask when queue is empty
func TestTaskStore_GetNextTask_EmptyQueue(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	store := NewTaskStore(db)

	// Simulate no rows found
	mock.ExpectQuery("UPDATE tasks").
		WithArgs(models.StatusProcessing, models.StatusPending).
		WillReturnError(sql.ErrNoRows)

	// Execute
	task, err := store.GetNextTask()

	// Assert -should return nil task and nil error (empty queue is normal)
	assert.NoError(t, err)
	assert.Nil(t, task)

	// Verify expectations
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// Table-driven test for UpdateStatus
func TestTaskStore_UpdateStatus(t *testing.T) {
	tests := []struct {
		name         string
		taskID       uuid.UUID
		status       models.Status
		errorMessage string
		mockError    error
		wantErr      bool
	}{
		{
			name:         "Success status",
			taskID:       uuid.New(),
			status:       models.StatusCompleted,
			errorMessage: "",
			mockError:    nil,
			wantErr:      false,
		},
		{
			name:         "Failed status with error",
			taskID:       uuid.New(),
			status:       models.StatusFailed,
			errorMessage: "connection timeout",
			mockError:    nil,
			wantErr:      false,
		},
		{
			name:         "Database error",
			taskID:       uuid.New(),
			status:       models.StatusCompleted,
			errorMessage: "",
			mockError:    sql.ErrConnDone,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create mock db: %v", err)
			}
			defer db.Close()

			store := NewTaskStore(db)

			// Setup expectation
			expectation := mock.ExpectExec("UPDATE tasks").
				WithArgs(tt.status, tt.errorMessage, tt.taskID)

			if tt.mockError != nil {
				expectation.WillReturnError(tt.mockError)
			} else {
				expectation.WillReturnResult(sqlmock.NewResult(0, 1))
			}

			// Execute
			err = store.UpdateStatus(tt.taskID, tt.status, tt.errorMessage)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify expectations
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

// Test IncrementRetry
func TestTaskStore_IncrementRetry(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	store := NewTaskStore(db)

	taskID := uuid.New()
	nextRetryTime := time.Now().Add(5 * time.Minute)

	mock.ExpectExec("UPDATE tasks").
		WithArgs(models.StatusPending, taskID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = store.IncrementRetry(taskID, nextRetryTime)

	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
