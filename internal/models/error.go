package models

import "errors"

var (
	ErrTaskNotFound       = errors.New("task not found")
	ErrInvalidTaskType    = errors.New("invalid task type")
	ErrInvalidPayload     = errors.New("invalid payload")
	ErrMaxRetriesExceeded = errors.New("max retries exceeded")
)
