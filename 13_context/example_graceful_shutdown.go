//go:build ignore

package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ============================================
// Context + Worker Pool = Graceful Shutdown
//
// This is how production Go servers work:
// 1. Workers check ctx.Done() in their select loop
// 2. When it's time to shut down, cancel the context
// 3. All workers finish their current job, then exit
//
// Python equivalent:
//   for task in asyncio.all_tasks():
//       task.cancel()
//   await asyncio.gather(*tasks, return_exceptions=True)
// ============================================

type Job struct {
	ID       int
	Duration time.Duration
}

type Result struct {
	JobID    int
	WorkerID int
	Output   string
}

func worker(ctx context.Context, id int, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	completed := 0

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("  Worker %d: shutting down (completed %d jobs, reason: %v)\n",
				id, completed, ctx.Err())
			return

		case job, ok := <-jobs:
			if !ok {
				fmt.Printf("  Worker %d: no more jobs (completed %d)\n", id, completed)
				return
			}

			// Simulate processing with cancellation check
			select {
			case <-ctx.Done():
				fmt.Printf("  Worker %d: cancelled mid-job %d\n", id, job.ID)
				return
			case <-time.After(job.Duration):
				completed++
				results <- Result{
					JobID:    job.ID,
					WorkerID: id,
					Output:   fmt.Sprintf("job-%d processed", job.ID),
				}
			}
		}
	}
}

func main() {
	// ============================================
	// Scenario: Worker pool with 3-second deadline
	// ============================================
	fmt.Println("=== Worker Pool with Context Timeout ===")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	jobs := make(chan Job, 50)
	results := make(chan Result, 50)
	var wg sync.WaitGroup

	// Launch 3 workers
	numWorkers := 3
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go worker(ctx, w, jobs, results, &wg)
	}

	// Send many jobs (more than can finish in 3 seconds)
	fmt.Println("Sending 20 jobs (each takes 500ms)...")
	go func() {
		for i := 1; i <= 20; i++ {
			select {
			case jobs <- Job{ID: i, Duration: 500 * time.Millisecond}:
			case <-ctx.Done():
				fmt.Printf("  Stopped sending at job %d (context cancelled)\n", i)
				close(jobs)
				return
			}
		}
		close(jobs)
	}()

	// Collect results until workers are done
	go func() {
		wg.Wait()
		close(results)
	}()

	start := time.Now()
	totalProcessed := 0
	for r := range results {
		totalProcessed++
		fmt.Printf("  Got result: Worker %d → %s\n", r.WorkerID, r.Output)
	}

	elapsed := time.Since(start)
	fmt.Printf("\nSummary:\n")
	fmt.Printf("  Time elapsed: %v\n", elapsed.Round(time.Millisecond))
	fmt.Printf("  Jobs processed: %d / 20\n", totalProcessed)
	fmt.Printf("  Workers: %d\n", numWorkers)
	fmt.Printf("  Expected: ~%d jobs in 3s (%d workers × 6 jobs each at 500ms)\n",
		numWorkers*6, numWorkers)
}
