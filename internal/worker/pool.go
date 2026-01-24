package worker

import (
	"context"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/Bharat1Rajput/task-management/internal/models"
	"github.com/Bharat1Rajput/task-management/internal/store"
)

// manages a set of worker routines
type Pool struct {
	store       *store.TaskStore // The Task Store (DB)
	processor   *Processor       // The Task Processor
	concurrency int              // Number of concurrent workers
	wg          sync.WaitGroup   // To wait for all workers to finish
}

// NewPool creates a new Worker Pool
func NewPool(store *store.TaskStore, concurrency int) *Pool {
	return &Pool{
		store:       store,
		processor:   NewProcessor(store),
		concurrency: concurrency,
	}
}

func (p *Pool) Start(ctx context.Context) {
	fmt.Printf("Starting Worker Pool with %d workers...\n", p.concurrency)

	for i := 0; i < p.concurrency; i++ {
		p.wg.Add(1)
		go p.workerLoop(ctx, i+1)
	}

	// Wait for shutdown signal
	<-ctx.Done()
	p.wg.Wait()
}

func (p *Pool) workerLoop(ctx context.Context, workerID int) {
	defer p.wg.Done()

	for {
		//  Check for shutdown signal
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Fetch Task from DB
		task, err := p.store.GetNextTask()
		if err != nil {
			log.Printf("[Worker %d] DB Error: %v", workerID, err)
			time.Sleep(1 * time.Second)
			continue
		}

		// If no work, backoff (Sleep)
		if task == nil {
			time.Sleep(1 * time.Second)
			continue
		}

		// Process the Task
		startTime := time.Now()
		err = p.processor.ProcessTask(task)
		duration := time.Since(startTime)

		// Handle Result
		if err != nil {
			log.Printf("[Worker %d] Failed Task %s (%v) in %v", workerID, task.ID, err, duration)
			p.handleFailure(task, err)
		} else {
			log.Printf("[Worker %d] Delivered Task %s in %v", workerID, task.ID, duration)
			p.handleSuccess(task)
		}
	}
}

func (p *Pool) handleSuccess(task *models.Task) {
	// Mark as COMPLETED and clear any error message
	_ = p.store.UpdateStatus(task.ID, models.StatusCompleted, "")
}

func (p *Pool) handleFailure(task *models.Task, err error) {
	if task.Retries < task.MaxRetries {
		// Exponential Backoff Formula: 2^retry_count * 1 second
		// Retry 1: 2s, Retry 2: 4s, Retry 3: 8s...
		backoffSeconds := math.Pow(2, float64(task.Retries+1))
		nextTry := time.Now().Add(time.Duration(backoffSeconds) * time.Second)

		log.Printf(" -> Retrying task %s in %.0f seconds (Attempt %d/%d)",
			task.ID, backoffSeconds, task.Retries+1, task.MaxRetries)

		//Update DB (Increment Retry Count)
		_ = p.store.IncrementRetry(task.ID, nextTry)
	} else {
		log.Printf("Task %s has exhausted all retries. Marking FAILED.", task.ID)
		_ = p.store.UpdateStatus(task.ID, models.StatusFailed, err.Error())
	}
}
