//go:build ignore

package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================
// Race condition demonstrations
// Run with: go run -race example_race_detector.go
//
// The race detector will identify EXACTLY which
// line of code causes the race.
// ============================================

func main() {
	// ============================================
	// Example 1: Simple counter race
	// ============================================
	fmt.Println("=== Example 1: Counter race ===")
	fmt.Println("  Run with: go run -race example_race_detector.go")

	var unsafeCounter int
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				unsafeCounter++ // RACE: read-modify-write is 3 operations
			}
		}()
	}
	wg.Wait()
	fmt.Printf("  Unsafe: %d (expected 10000)\n", unsafeCounter)

	// ============================================
	// Fix A: Mutex
	// ============================================
	fmt.Println("\n=== Fix A: Mutex ===")

	var mu sync.Mutex
	safeCounterMu := 0

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				mu.Lock()
				safeCounterMu++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	fmt.Printf("  Mutex: %d (always 10000)\n", safeCounterMu)

	// ============================================
	// Fix B: Atomic
	// ============================================
	fmt.Println("\n=== Fix B: Atomic ===")

	var safeCounterAtomic int64

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				atomic.AddInt64(&safeCounterAtomic, 1)
			}
		}()
	}
	wg.Wait()
	fmt.Printf("  Atomic: %d (always 10000)\n", safeCounterAtomic)

	// ============================================
	// Fix C: Channel
	// ============================================
	fmt.Println("\n=== Fix C: Channel ===")

	safeCounterCh := 0
	ch := make(chan int, 100)

	// Counter goroutine — single owner of the data
	done := make(chan struct{})
	go func() {
		for inc := range ch {
			safeCounterCh += inc
		}
		close(done)
	}()

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				ch <- 1 // send increment request
			}
		}()
	}
	wg.Wait()
	close(ch) // signal counter goroutine to stop
	<-done    // wait for all increments to be processed

	fmt.Printf("  Channel: %d (always 10000)\n", safeCounterCh)

	// ============================================
	// Performance comparison
	// ============================================
	fmt.Println("\n=== Performance comparison (1M increments) ===")

	const N = 1_000_000

	// Mutex
	start := time.Now()
	safeCounterMu = 0
	for i := 0; i < N; i++ {
		mu.Lock()
		safeCounterMu++
		mu.Unlock()
	}
	fmt.Printf("  Mutex (single goroutine):  %v\n", time.Since(start))

	// Atomic
	start = time.Now()
	safeCounterAtomic = 0
	for i := 0; i < N; i++ {
		atomic.AddInt64(&safeCounterAtomic, 1)
	}
	fmt.Printf("  Atomic (single goroutine): %v\n", time.Since(start))

	fmt.Println("  Note: atomic is faster because it uses CPU hardware instructions")
	fmt.Println("  Mutex involves OS-level synchronization overhead")
}
