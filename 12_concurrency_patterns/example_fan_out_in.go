//go:build ignore

package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ============================================
// Fan-Out / Fan-In Pattern
//
// Fan-OUT: Multiple goroutines read from the SAME channel.
//          The slow stage gets parallelized.
//
// Fan-IN:  Multiple channels are merged into ONE channel.
//          Downstream sees a single stream.
//
// Python equivalent:
//   tasks = [asyncio.create_task(process(item)) for item in items]
//   results = await asyncio.gather(*tasks)
// ============================================

func main() {
	// ============================================
	// Example 1: Fan-out — parallelize a slow stage
	// ============================================
	fmt.Println("=== Example 1: Fan-out with 4 workers ===")

	// Generate work
	jobs := make(chan int, 20)
	go func() {
		for i := 1; i <= 20; i++ {
			jobs <- i
		}
		close(jobs)
	}()

	// Fan-OUT: 4 workers all read from the same jobs channel
	var workerChannels []<-chan string
	for w := 1; w <= 4; w++ {
		ch := slowProcess(w, jobs) // each returns its own result channel
		workerChannels = append(workerChannels, ch)
	}

	// Fan-IN: merge all worker channels into one
	merged := merge(workerChannels...)

	start := time.Now()
	count := 0
	for result := range merged {
		fmt.Println(" ", result)
		count++
	}
	fmt.Printf("  Processed %d items in %v (4x faster than serial)\n", count, time.Since(start))

	// ============================================
	// Example 2: Fan-out/Fan-in with typed results
	// ============================================
	fmt.Println("\n=== Example 2: Typed fan-in ===")

	type Result struct {
		WorkerID int
		Value    int
	}

	source := make(chan int, 10)
	go func() {
		for i := 1; i <= 10; i++ {
			source <- i
		}
		close(source)
	}()

	resultCh := make(chan Result, 10)
	var wg sync.WaitGroup

	// Fan-out: 3 workers
	for w := 1; w <= 3; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for val := range source {
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
				resultCh <- Result{WorkerID: workerID, Value: val * val}
			}
		}(w)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	for r := range resultCh {
		fmt.Printf("  Worker %d: %d\n", r.WorkerID, r.Value)
	}
}

// slowProcess simulates a slow worker. It reads from jobs and returns a results channel.
func slowProcess(id int, jobs <-chan int) <-chan string {
	results := make(chan string)
	go func() {
		defer close(results)
		for job := range jobs {
			// Simulate slow processing (100-300ms)
			time.Sleep(time.Duration(100+rand.Intn(200)) * time.Millisecond)
			results <- fmt.Sprintf("Worker %d: job %d done", id, job)
		}
	}()
	return results
}

// merge combines multiple channels into one (fan-in).
func merge(channels ...<-chan string) <-chan string {
	var wg sync.WaitGroup
	merged := make(chan string)

	// Start a goroutine for each input channel
	for _, ch := range channels {
		wg.Add(1)
		go func(c <-chan string) {
			defer wg.Done()
			for val := range c {
				merged <- val
			}
		}(ch)
	}

	// Close merged channel when all inputs are drained
	go func() {
		wg.Wait()
		close(merged)
	}()

	return merged
}
