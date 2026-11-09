package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/enterprise-erp/core/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type JobHandlerFunc func(ctx context.Context, payload json.RawMessage) error

type Job struct {
	ID        uuid.UUID
	JobType   string
	Payload   json.RawMessage
	Attempts  int
	Priority  int
	CreatedAt time.Time
}

type WorkerPool struct {
	db          *database.DB
	workerCount int
	handlers    map[string]JobHandlerFunc
	stopChan    chan struct{}
	wg          sync.WaitGroup
	workerID    string
}

func NewWorkerPool(db *database.DB, workerCount int) *WorkerPool {
	if workerCount <= 0 {
		workerCount = 4
	}
	return &WorkerPool{
		db:          db,
		workerCount: workerCount,
		handlers:    make(map[string]JobHandlerFunc),
		stopChan:    make(chan struct{}),
		workerID:    fmt.Sprintf("worker-%s", uuid.New().String()[:8]),
	}
}

// RegisterHandler registers a processor for a specific job_type
func (wp *WorkerPool) RegisterHandler(jobType string, handler JobHandlerFunc) {
	wp.handlers[jobType] = handler
}

// EnqueueJob pushes a job into the PostgreSQL table
func (wp *WorkerPool) EnqueueJob(ctx context.Context, jobType string, payload interface{}, priority int) (uuid.UUID, error) {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to marshal job payload: %w", err)
	}

	jobID := uuid.New()
	query := `
		INSERT INTO background_jobs (id, job_type, payload, status, priority, run_at, created_at, updated_at)
		VALUES ($1, $2, $3, 'queued', $4, NOW(), NOW(), NOW())
		RETURNING id;
	`
	err = wp.db.Pool.QueryRow(ctx, query, jobID, jobType, bytes, priority).Scan(&jobID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to enqueue job: %w", err)
	}

	return jobID, nil
}

// Start launches worker goroutines that poll with FOR UPDATE SKIP LOCKED
func (wp *WorkerPool) Start() {
	log.Printf("[INFO] Starting PostgreSQL SKIP LOCKED Worker Pool (%d workers, node: %s)...", wp.workerCount, wp.workerID)
	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.runWorker(i)
	}
}

func (wp *WorkerPool) Stop() {
	log.Println("[INFO] Stopping Background Worker Pool...")
	close(wp.stopChan)
	wp.wg.Wait()
	log.Println("[INFO] Background Worker Pool exited safely.")
}

func (wp *WorkerPool) runWorker(id int) {
	defer wp.wg.Done()
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-wp.stopChan:
			return
		case <-ticker.C:
			wp.processNextJob(id)
		}
	}
}

func (wp *WorkerPool) processNextJob(workerNum int) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	tx, err := wp.db.Pool.Begin(ctx)
	if err != nil {
		return
	}
	defer tx.Rollback(ctx)

	// Atomic PostgreSQL Lock-Free Queue Fetch
	query := `
		WITH next_job AS (
			SELECT id
			FROM background_jobs
			WHERE status = 'queued' AND run_at <= NOW()
			ORDER BY priority DESC, created_at ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE background_jobs
		SET status = 'processing', locked_at = NOW(), locked_by = $1, attempts = attempts + 1, updated_at = NOW()
		WHERE id = (SELECT id FROM next_job)
		RETURNING id, job_type, payload, attempts, priority, created_at;
	`

	var job Job
	err = tx.QueryRow(ctx, query, wp.workerID).Scan(
		&job.ID, &job.JobType, &job.Payload, &job.Attempts, &job.Priority, &job.CreatedAt,
	)
	if err != nil {
		if errorsIsNoRows(err) {
			// No jobs queued, return silently
			return
		}
		log.Printf("[ERROR] Worker %d error fetching job: %v", workerNum, err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		return
	}

	log.Printf("[INFO] Worker %d picked up job [%s] ID: %s (attempt %d)", workerNum, job.JobType, job.ID, job.Attempts)

	// Execute handler
	handler, exists := wp.handlers[job.JobType]
	var jobErr error
	if !exists {
		jobErr = fmt.Errorf("no handler registered for job_type: %s", job.JobType)
	} else {
		jobErr = handler(ctx, job.Payload)
	}

	// Update Job Status
	wp.finalizeJob(context.Background(), job.ID, jobErr, job.Attempts)
}

func (wp *WorkerPool) finalizeJob(ctx context.Context, jobID uuid.UUID, jobErr error, attempts int) {
	if jobErr == nil {
		_, _ = wp.db.Pool.Exec(ctx, `
			UPDATE background_jobs
			SET status = 'completed', updated_at = NOW(), error_message = NULL
			WHERE id = $1;
		`, jobID)
		log.Printf("[INFO] Job %s completed successfully.", jobID)
	} else {
		log.Printf("[WARN] Job %s failed: %v", jobID, jobErr)
		status := "queued"
		// Retry up to 5 times with exponential backoff
		if attempts >= 5 {
			status = "failed"
		}
		backoffSeconds := attempts * attempts * 10
		nextRun := time.Now().Add(time.Duration(backoffSeconds) * time.Second)

		_, _ = wp.db.Pool.Exec(ctx, `
			UPDATE background_jobs
			SET status = $1, run_at = $2, error_message = $3, updated_at = NOW()
			WHERE id = $4;
		`, status, nextRun, jobErr.Error(), jobID)
	}
}

func errorsIsNoRows(err error) bool {
	return err == pgx.ErrNoRows
}
