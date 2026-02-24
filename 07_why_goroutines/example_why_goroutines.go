//go:build ignore

package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// ============================================
// This example demonstrates the basics of goroutines and why they're cheap.
//
// Python equivalent would be:
//   import asyncio
//   async def worker(id):
//       await asyncio.sleep(1)
//       print(f"Worker {id} done")
//   asyncio.run(asyncio.gather(*[worker(i) for i in range(10)]))
//
// But in Go, there's no async/await — just `go` keyword.
// And it runs on MULTIPLE cores, not one.
// ============================================

func worker(id int, wg *sync.WaitGroup) {
	defer wg.Done() // signal completion when this function returns
	fmt.Printf("Worker %d starting on goroutine\n", id)
	time.Sleep(1 * time.Second) // simulate I/O work
	fmt.Printf("Worker %d done\n", id)
}

func main() {
	fmt.Printf("Number of CPU cores: %d\n", runtime.NumCPU())
	fmt.Printf("GOMAXPROCS (threads used): %d\n", runtime.GOMAXPROCS(0))
	fmt.Println()

	// ============================================
	// Launch 10 goroutines — all run concurrently
	// ============================================
	var wg sync.WaitGroup

	start := time.Now()

	for i := 1; i <= 10; i++ {
		wg.Add(1)         // tell WaitGroup we're launching one more goroutine
		go worker(i, &wg) // `go` keyword = launch as goroutine
	}

	// Wait for all goroutines to finish
	// (Without this, main() would exit and kill all goroutines!)
	wg.Wait()

	elapsed := time.Since(start)
	fmt.Printf("\n10 workers completed in %v (not 10 seconds — they ran concurrently!)\n", elapsed)

	// ============================================
	// Show how lightweight goroutines are
	// ============================================
	fmt.Println("\n=== Launching 100,000 goroutines ===")

	var wg2 sync.WaitGroup
	start2 := time.Now()

	for i := 0; i < 100_000; i++ {
		wg2.Add(1)
		go func(id int) {
			defer wg2.Done()
			// Each goroutine does minimal work
			_ = id * id
		}(i)
	}

	wg2.Wait()
	elapsed2 := time.Since(start2)
	fmt.Printf("100,000 goroutines completed in %v\n", elapsed2)
	fmt.Println("Try doing that with Python threads! (Spoiler: you can't)")

	// ============================================
	// Show current goroutine count
	// ============================================
	fmt.Printf("\nActive goroutines right now: %d\n", runtime.NumGoroutine())
}
