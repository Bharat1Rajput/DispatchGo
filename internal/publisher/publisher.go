package publisher

import (
	"context"

	"github.com/Bharat1Rajput/task-management/internal/models"
)

type JobPublisher interface {
	Publish(ctx context.Context, job models.Job) error
}
