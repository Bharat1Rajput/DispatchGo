package models

import "time"

type Job struct {
	ID         string
	Payload    string
	TargetURL  string
	Status     string
	RetryCount int
	CreatedAt  time.Time
}
