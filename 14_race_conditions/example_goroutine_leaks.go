//go:build ignore

package main

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// ============================================
// Goroutine Leaks — when goroutines never exit
//
// Each goroutine uses ~2-8KB memory.
// Leak 1000 per second = OOM in hours/days.
//
// Python equivalent: creating asyncio tasks that never complete
//   tasks = [asyncio.create_task(blocked_coro()) for _ in range(1000)]
//   # tasks pile up in memory
// ============================================

func main() {
	// ============================================
	// Example 1: Goroutine leak — blocked channel
	// ============================================
	fmt.Println("=== Example 1: Leaky generator ===")

	before := runtime.NumGoroutine()
	fmt.Printf("  Goroutines before: %d\n", before)

	// This function creates a goroutine that sends forever
	// If the consumer stops reading, the goroutine is stuck on send
	leakyCh := leakyGenerator()

	// Read only 3 values, then abandon the channel
	for i := 0; i < 3; i++ {
		fmt.Printf("  Got: %d\n", <-leakyCh)
	}
	// We stopped reading! The goroutine inside leakyGenerator is now
	// blocked on `ch <- i` forever. It's leaked!

	time.Sleep(100 * time.Millisecond)
	after := runtime.NumGoroutine()
	fmt.Printf("  Goroutines after: %d (leaked %d!)\n", after, after-before)

	// ============================================
	// Example 2: Fixed with context cancellation
	// ============================================
	fmt.Println("\n=== Example 2: Fixed with context ===")

	before = runtime.NumGoroutine()
	fmt.Printf("  Goroutines before: %d\n", before)

	ctx, cancel := context.WithCancel(context.Background())
	safeCh := safeGenerator(ctx)

	for i := 0; i < 3; i++ {
		fmt.Printf("  Got: %d\n", <-safeCh)
	}
	cancel() // signal the generator to stop

	time.Sleep(100 * time.Millisecond)
	after = runtime.NumGoroutine()
	fmt.Printf("  Goroutines after: %d (no leak!)\n", after)

	// ============================================
	// Example 3: Fixed with done channel
	// ============================================
	fmt.Println("\n=== Example 3: Fixed with done channel ===")

	before = runtime.NumGoroutine()
	done := make(chan struct{})
	doneCh := doneGenerator(done)

	for i := 0; i < 3; i++ {
		fmt.Printf("  Got: %d\n", <-doneCh)
	}
	close(done) // signal stop

	time.Sleep(100 * time.Millisecond)
	after = runtime.NumGoroutine()
	fmt.Printf("  Goroutines after cancel: %d (no leak!)\n", after)

	// ============================================
	// Example 4: Leak detector helper
	// ============================================
	fmt.Println("\n=== Example 4: Goroutine leak detector ===")
	checkLeaks := func(name string, f func()) {
		before := runtime.NumGoroutine()
		f()
		time.Sleep(200 * time.Millisecond)
		after := runtime.NumGoroutine()
		if after > before {
			fmt.Printf("  ⚠ %s: LEAKED %d goroutines (%d → %d)\n", name, after-before, before, after)
		} else {
			fmt.Printf("  ✓ %s: no leak (%d → %d)\n", name, before, after)
		}
	}

	checkLeaks("leaky version", func() {
		ch := leakyGenerator()
		<-ch
		<-ch
	})

	checkLeaks("safe version", func() {
		ctx, cancel := context.WithCancel(context.Background())
		ch := safeGenerator(ctx)
		<-ch
		<-ch
		cancel()
	})
}

// leakyGenerator creates a goroutine that may never exit
func leakyGenerator() <-chan int {
	ch := make(chan int)
	go func() {
		for i := 0; ; i++ {
			ch <- i // blocks forever if consumer stops reading!
		}
	}()
	return ch
}

// safeGenerator uses context for clean cancellation
func safeGenerator(ctx context.Context) <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for i := 0; ; i++ {
			select {
			case <-ctx.Done():
				return // clean exit
			case ch <- i:
			}
		}
	}()
	return ch
}

// doneGenerator uses a done channel for clean cancellation
func doneGenerator(done <-chan struct{}) <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for i := 0; ; i++ {
			select {
			case <-done:
				return
			case ch <- i:
			}
		}
	}()
	return ch
}
