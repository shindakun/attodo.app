package jobs

import (
	"context"
	"log"
	"sync"
	"time"
)

// Job represents a background job that runs periodically
type Job interface {
	// Name returns the job name for logging
	Name() string
	// Run executes the job
	Run(ctx context.Context) error
}

// Runner manages and executes background jobs
type Runner struct {
	jobs     []Job
	interval time.Duration
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// NewRunner creates a new job runner
// interval is how often to run all jobs
func NewRunner(interval time.Duration) *Runner {
	ctx, cancel := context.WithCancel(context.Background())
	return &Runner{
		jobs:     make([]Job, 0),
		interval: interval,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// AddJob registers a job to run
func (r *Runner) AddJob(job Job) {
	r.jobs = append(r.jobs, job)
	log.Printf("Registered job: %s", job.Name())
}

// Start begins running jobs in the background
func (r *Runner) Start() {
	if len(r.jobs) == 0 {
		log.Println("No jobs registered, runner will not start")
		return
	}

	log.Printf("Starting job runner with %d job(s), interval: %s", len(r.jobs), r.interval)

	r.wg.Add(1)
	go r.run()
}

// run is the main job loop
func (r *Runner) run() {
	defer r.wg.Done()

	// Run jobs immediately on start
	r.runAll()

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.runAll()
		case <-r.ctx.Done():
			log.Println("Job runner stopping...")
			return
		}
	}
}

// runAll executes all registered jobs
func (r *Runner) runAll() {
	for _, job := range r.jobs {
		// Run each job in its own goroutine
		go func(j Job) {
			start := time.Now()
			log.Printf("Running job: %s", j.Name())

			if err := j.Run(r.ctx); err != nil {
				log.Printf("Job %s failed: %v", j.Name(), err)
			} else {
				log.Printf("Job %s completed in %s", j.Name(), time.Since(start))
			}
		}(job)
	}
}

// Stop gracefully stops the job runner
func (r *Runner) Stop() {
	log.Println("Stopping job runner...")
	r.cancel()
	r.wg.Wait()
	log.Println("Job runner stopped")
}
