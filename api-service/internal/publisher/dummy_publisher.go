package publisher

import (
	"context"
	"log"

	"github.com/Bharat1Rajput/apiService/internal/models"
)

type DummyPublisher struct{}

func NewDummyPublisher() *DummyPublisher {
	return &DummyPublisher{}
}

func (d *DummyPublisher) Publish(ctx context.Context, job models.Job) error {
	log.Println("Received job:", job.ID)
	return nil
}
