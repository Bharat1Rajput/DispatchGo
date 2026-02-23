package models

import "time"

type Job struct {
	ID        string    `json:"id"`
	Payload   string    `json:"payload"`
	TargetURL string    `json:"target_url"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
