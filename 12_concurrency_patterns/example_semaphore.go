//go:build ignore

package main

import (
	"fmt"
	"sync"
	"time"
)

// ============================================
// Semaphore Pattern — Limit concurrent goroutines
//
// Use a buffered channel as a semaphore.
// The channel capacity = max concurrent goroutines.
//
// Python equivalent:
//   sem = asyncio.Semaphore(5)
//   async with sem:
//       await do_work()
// ============================================

func main() {
	// ============================================
	// Example 1: Buffered channel as semaphore
	// ============================================
	fmt.Println("=== Example 1: Semaphore (max 3 concurrent) ===")

	const maxConcurrent = 3
	sem := make(chan struct{}, maxConcurrent) // buffered channel = semaphore

	var wg sync.WaitGroup

	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			sem <- struct{}{}        // acquire: blocks when 3 are running
			defer func() { <-sem }() // release when done

			fmt.Printf("  [%s] Worker %d: START (running concurrently: %d)\n",
				time.Now().Format("15:04:05.000"), id, len(sem))
			time.Sleep(500 * time.Millisecond) // simulate work
			fmt.Printf("  [%s] Worker %d: DONE\n",
				time.Now().Format("15:04:05.000"), id)
		}(i)
	}

	wg.Wait()
	fmt.Println("  Notice: at most 3 workers ran at the same time!")

	// ============================================
	// Example 2: Rate limiter — max N operations per second
	// ============================================
	fmt.Println("\n=== Example 2: Rate limiter (5 per second) ===")

	// A ticker-based rate limiter
	rateLimit := time.NewTicker(200 * time.Millisecond) // 5 per second (1000ms / 200ms = 5)
	defer rateLimit.Stop()

	start := time.Now()
	for i := 1; i <= 10; i++ {
		<-rateLimit.C // wait for next tick
		fmt.Printf("  Request %d sent at %v\n", i, time.Since(start).Round(time.Millisecond))
	}
	fmt.Println("  10 requests sent over ~2 seconds (5 per second)")

	// ============================================
	// Example 3: Weighted semaphore
	// Some tasks need more slots than others
	// ============================================
	fmt.Println("\n=== Example 3: Weighted semaphore ===")

	type Task struct {
		Name   string
		Weight int // how many semaphore slots this needs
	}

	tasks := []Task{
		{"light-1", 1},
		{"heavy-1", 3},
		{"light-2", 1},
		{"light-3", 1},
		{"heavy-2", 2},
	}

	capacity := 4
	weightedSem := make(chan struct{}, capacity)
	var wg2 sync.WaitGroup

	for _, task := range tasks {
		wg2.Add(1)
		go func(t Task) {
			defer wg2.Done()

			// Acquire N slots
			for i := 0; i < t.Weight; i++ {
				weightedSem <- struct{}{}
			}

			fmt.Printf("  %s (weight=%d) running, slots used: %d/%d\n",
				t.Name, t.Weight, len(weightedSem), capacity)
			time.Sleep(300 * time.Millisecond)

			// Release N slots
			for i := 0; i < t.Weight; i++ {
				<-weightedSem
			}
		}(task)
	}

	wg2.Wait()
	fmt.Println("  All tasks done!")
}
