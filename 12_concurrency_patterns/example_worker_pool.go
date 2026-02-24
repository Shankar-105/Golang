//go:build ignore

package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ============================================
// Worker Pool Pattern
//
// The most common Go concurrency pattern.
// N workers read from a shared jobs channel.
// Results go to a results channel.
//
// Python equivalent:
//   asyncio.Queue + N worker coroutines
//   or concurrent.futures.ThreadPoolExecutor
// ============================================

func worker(id int, jobs <-chan int, results chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		// Simulate variable processing time
		duration := time.Duration(50+rand.Intn(200)) * time.Millisecond
		time.Sleep(duration)
		results <- fmt.Sprintf("Worker %d completed job %d in %v", id, job, duration)
	}
}

func main() {
	// ============================================
	// Example 1: Basic worker pool — 3 workers, 10 jobs
	// ============================================
	fmt.Println("=== Example 1: Basic Worker Pool ===")

	const numWorkers = 3
	const numJobs = 10

	jobs := make(chan int, numJobs)
	results := make(chan string, numJobs)
	var wg sync.WaitGroup

	// Launch workers
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go worker(w, jobs, results, &wg)
	}

	// Send jobs
	start := time.Now()
	for j := 1; j <= numJobs; j++ {
		jobs <- j
	}
	close(jobs) // signal: no more jobs

	// Wait for all workers to finish, then close results
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	for result := range results {
		fmt.Println(" ", result)
	}
	fmt.Printf("  Total time: %v (much less than serial!)\n", time.Since(start))

	// ============================================
	// Example 2: Worker pool with job struct
	// ============================================
	fmt.Println("\n=== Example 2: Typed Job + Result ===")

	type Job struct {
		ID   int
		Data string
	}

	type Result struct {
		JobID  int
		Output string
	}

	jobCh := make(chan Job, 5)
	resultCh := make(chan Result, 5)
	var wg2 sync.WaitGroup

	// Launch 4 workers
	for w := 0; w < 4; w++ {
		wg2.Add(1)
		go func(workerID int) {
			defer wg2.Done()
			for job := range jobCh {
				// "Process" the job
				output := fmt.Sprintf("processed(%s) by worker-%d", job.Data, workerID)
				resultCh <- Result{JobID: job.ID, Output: output}
			}
		}(w)
	}

	// Send jobs
	inputs := []string{"apple", "banana", "cherry", "date", "elderberry"}
	for i, input := range inputs {
		jobCh <- Job{ID: i, Data: input}
	}
	close(jobCh)

	go func() {
		wg2.Wait()
		close(resultCh)
	}()

	for result := range resultCh {
		fmt.Printf("  Job %d → %s\n", result.JobID, result.Output)
	}

	// ============================================
	// Example 3: Dynamic worker pool — adjust workers at runtime
	// ============================================
	fmt.Println("\n=== Example 3: Worker pool stats ===")
	fmt.Println("  Key insight: workers process jobs from the SAME channel.")
	fmt.Println("  Go's scheduler efficiently distributes work.")
	fmt.Println("  Unlike Python's GIL, workers truly run in parallel on multiple cores.")
	fmt.Println("  With 3 workers and 10 jobs (each ~125ms avg), parallel time ≈ 10/3 * 125ms ≈ 420ms")
	fmt.Println("  Serial time would be 10 * 125ms = 1250ms")
}
