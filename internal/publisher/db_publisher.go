package publisher

import (
	"context"

	"github.com/Bharat1Rajput/task-management/internal/models"
	"github.com/Bharat1Rajput/task-management/internal/store"
)

type DBPublisher struct {
	store *store.Store
}

func NewDBPublisher(s *store.Store) *DBPublisher {
	return &DBPublisher{store: s}
}

func (d *DBPublisher) Publish(ctx context.Context, job models.Job) error {
	return d.store.InsertJob(ctx, job)
}
