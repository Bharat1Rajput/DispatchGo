package worker

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Bharat1Rajput/workerService/internal/store"
)

type Worker struct {
	store *store.Store
}

func New(store *store.Store) *Worker {
	return &Worker{store: store}
}

func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	log.Println("Worker started")

	for {
		select {
		case <-ctx.Done():
			log.Println("Worker stopped")
			return

		case <-ticker.C:
			w.processJobs(ctx)
		}
	}
}

func (w *Worker) processJobs(ctx context.Context) {
	jobs, err := w.store.FetchPendingJobs(ctx, 10)
	if err != nil {
		log.Println("Error fetching jobs:", err)
		return
	}

	for _, job := range jobs {
		log.Println("Processing job:", job.ID)

		err := w.execute(job.TargetURL)

		if err != nil {
			log.Println("Job failed:", err)
			_ = w.store.UpdateJobStatus(ctx, job.ID, "failed")
		} else {
			_ = w.store.UpdateJobStatus(ctx, job.ID, "completed")
		}
	}
}

func (w *Worker) execute(targetURL string) error {
	resp, err := http.Get(targetURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return err
	}

	return nil
}
