//go:build ignore

package main

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================
// GOMAXPROCS demonstration:
// Shows how parallelism changes with GOMAXPROCS setting.
//
// GOMAXPROCS=1 → like asyncio (concurrent, not parallel)
// GOMAXPROCS=N → true parallelism on N cores
// ============================================

func cpuWork(iterations int) {
	total := 0
	for i := 0; i < iterations; i++ {
		total += i * i
	}
	_ = total
}

func benchmark(label string, numWorkers, maxProcs int) time.Duration {
	runtime.GOMAXPROCS(maxProcs)

	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cpuWork(10_000_000) // CPU-heavy work
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)
	fmt.Printf("  %s: %d workers, GOMAXPROCS=%d → %v\n", label, numWorkers, maxProcs, elapsed)
	return elapsed
}

func main() {
	numCPU := runtime.NumCPU()
	fmt.Printf("CPU cores available: %d\n\n", numCPU)

	// ============================================
	// Test 1: CPU-bound work with different GOMAXPROCS
	// ============================================
	fmt.Println("=== CPU-bound work: GOMAXPROCS comparison ===")

	// Single-threaded (like asyncio)
	t1 := benchmark("Single-thread", 8, 1)

	// All cores
	t2 := benchmark("All cores", 8, numCPU)

	speedup := float64(t1) / float64(t2)
	fmt.Printf("\n  Speedup: %.1fx faster with %d cores\n", speedup, numCPU)
	fmt.Println("  (In Python asyncio, CPU-bound work gets NO speedup — it's always 1x)")

	// ============================================
	// Test 2: Show goroutine preemption
	// Even with GOMAXPROCS=1, Go can preempt CPU-bound goroutines
	// (since Go 1.14). asyncio CANNOT do this.
	// ============================================
	fmt.Println("\n=== Preemption test (GOMAXPROCS=1) ===")
	runtime.GOMAXPROCS(1)

	var counter int64
	var wg sync.WaitGroup

	// Launch a CPU-hogging goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100_000_000; i++ {
			atomic.AddInt64(&counter, 1) // busy work
		}
		fmt.Println("  CPU hog: finished")
	}()

	// Launch a quick goroutine — will it run DURING the hog?
	wg.Add(1)
	go func() {
		defer wg.Done()
		// In asyncio, this would NEVER run until the hog yields.
		// In Go 1.14+, preemption allows this to run.
		time.Sleep(10 * time.Millisecond)
		fmt.Println("  Quick task: I got to run! (preemption works)")
	}()

	wg.Wait()
	fmt.Printf("  Counter: %d\n", counter)

	// Reset GOMAXPROCS
	runtime.GOMAXPROCS(numCPU)
}
