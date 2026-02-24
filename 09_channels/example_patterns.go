//go:build ignore

package main

import (
	"fmt"
	"time"
)

// ============================================
// Common channel patterns
// ============================================

func main() {
	// ============================================
	// Pattern 1: Done channel (signaling completion)
	// Like asyncio.Event() — a one-time signal.
	// ============================================
	fmt.Println("=== Pattern 1: Done channel ===")

	done := make(chan struct{}) // empty struct = 0 bytes. Pure signal.

	go func() {
		fmt.Println("  Background task running...")
		time.Sleep(500 * time.Millisecond)
		fmt.Println("  Background task complete!")
		close(done) // signal completion
	}()

	fmt.Println("  Main waiting...")
	<-done // blocks until closed
	fmt.Println("  Main: continuing after background task")

	// ============================================
	// Pattern 2: Generator (like Python's yield)
	// Function returns a channel that produces values lazily.
	// ============================================
	fmt.Println("\n=== Pattern 2: Generator ===")

	for val := range fibonacci(10) {
		fmt.Printf("  %d ", val)
	}
	fmt.Println()

	// ============================================
	// Pattern 3: Fan-out (one producer, multiple consumers)
	// Distribute work across multiple goroutines.
	//
	// Python equivalent:
	//   workers = [asyncio.create_task(worker(q)) for _ in range(3)]
	// ============================================
	fmt.Println("\n=== Pattern 3: Fan-out ===")

	jobs := make(chan int, 10)
	results := make(chan string, 10)

	// Start 3 worker goroutines
	for w := 1; w <= 3; w++ {
		go worker(w, jobs, results)
	}

	// Send 9 jobs
	for j := 1; j <= 9; j++ {
		jobs <- j
	}
	close(jobs) // no more jobs

	// Collect results
	for i := 0; i < 9; i++ {
		fmt.Printf("  %s\n", <-results)
	}

	// ============================================
	// Pattern 4: Fan-in (multiple producers, one consumer)
	// Merge multiple channels into one.
	//
	// Python equivalent:
	//   async for result in merge_async_iters(iter1, iter2):
	// ============================================
	fmt.Println("\n=== Pattern 4: Fan-in (merge) ===")

	ch1 := emitter("A", 100*time.Millisecond, 3)
	ch2 := emitter("B", 150*time.Millisecond, 3)

	merged := fanIn(ch1, ch2)

	for val := range merged {
		fmt.Printf("  Merged: %s\n", val)
	}

	// ============================================
	// Pattern 5: Timeout using channel + time.After
	// ============================================
	fmt.Println("\n=== Pattern 5: Timeout ===")

	slowCh := make(chan string)
	go func() {
		time.Sleep(2 * time.Second) // simulates slow operation
		slowCh <- "slow result"
	}()

	select {
	case result := <-slowCh:
		fmt.Println("  Got:", result)
	case <-time.After(500 * time.Millisecond):
		fmt.Println("  Timed out! (waited 500ms, task took 2s)")
	}
}

// fibonacci generates the first n Fibonacci numbers into a channel
func fibonacci(n int) <-chan int {
	ch := make(chan int)
	go func() {
		a, b := 0, 1
		for i := 0; i < n; i++ {
			ch <- a
			a, b = b, a+b
		}
		close(ch)
	}()
	return ch
}

// worker processes jobs and sends results
func worker(id int, jobs <-chan int, results chan<- string) {
	for j := range jobs {
		time.Sleep(50 * time.Millisecond) // simulate work
		results <- fmt.Sprintf("worker %d processed job %d", id, j)
	}
}

// emitter sends n labeled messages at the given interval
func emitter(label string, interval time.Duration, count int) <-chan string {
	ch := make(chan string)
	go func() {
		for i := 1; i <= count; i++ {
			time.Sleep(interval)
			ch <- fmt.Sprintf("%s-%d", label, i)
		}
		close(ch)
	}()
	return ch
}

// fanIn merges two channels into one
func fanIn(ch1, ch2 <-chan string) <-chan string {
	merged := make(chan string)
	go func() {
		// Use two goroutines to read from both inputs simultaneously
		done := make(chan struct{})
		go func() {
			for val := range ch1 {
				merged <- val
			}
			done <- struct{}{}
		}()
		go func() {
			for val := range ch2 {
				merged <- val
			}
			done <- struct{}{}
		}()
		<-done // wait for first
		<-done // wait for second
		close(merged)
	}()
	return merged
}
