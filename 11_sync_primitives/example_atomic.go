//go:build ignore

package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================
// sync/atomic — lock-free atomic operations
//
// For simple counters and flags, atomics are faster
// than mutexes because they use hardware CPU instructions
// (CAS — Compare-And-Swap) instead of OS-level locks.
// ============================================

func main() {
	// ============================================
	// Example 1: Atomic counter vs Mutex counter
	// ============================================
	fmt.Println("=== Example 1: Atomic vs Mutex performance ===")

	const N = 1_000_000
	const workers = 8

	// Atomic counter
	var atomicCounter int64
	start := time.Now()
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < N/workers; j++ {
				atomic.AddInt64(&atomicCounter, 1)
			}
		}()
	}
	wg.Wait()
	atomicTime := time.Since(start)

	// Mutex counter
	var mutexCounter int64
	var mu sync.Mutex
	start = time.Now()

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < N/workers; j++ {
				mu.Lock()
				mutexCounter++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	mutexTime := time.Since(start)

	fmt.Printf("  Atomic: %v (count=%d)\n", atomicTime, atomicCounter)
	fmt.Printf("  Mutex:  %v (count=%d)\n", mutexTime, mutexCounter)
	fmt.Printf("  Atomic is ~%.1fx faster\n", float64(mutexTime)/float64(atomicTime))

	// ============================================
	// Example 2: Atomic flag (done signal)
	// ============================================
	fmt.Println("\n=== Example 2: Atomic flag ===")

	var done int32 // 0 = not done, 1 = done

	go func() {
		fmt.Println("  Worker: processing...")
		time.Sleep(500 * time.Millisecond)
		atomic.StoreInt32(&done, 1) // set flag atomically
		fmt.Println("  Worker: done!")
	}()

	// Poll for completion (not ideal — channels are better for this)
	for atomic.LoadInt32(&done) == 0 {
		time.Sleep(100 * time.Millisecond)
		fmt.Println("  Main: still waiting...")
	}
	fmt.Println("  Main: worker finished!")

	// ============================================
	// Example 3: Compare-and-Swap (CAS)
	// Fundamental building block of lock-free algorithms
	// ============================================
	fmt.Println("\n=== Example 3: Compare-and-Swap (CAS) ===")

	var value int64 = 42

	// CAS: "If value is 42, set it to 100. Tell me if it worked."
	swapped := atomic.CompareAndSwapInt64(&value, 42, 100)
	fmt.Printf("  CAS(42→100): swapped=%t, value=%d\n", swapped, value)

	// Second CAS: value is now 100, not 42, so this fails
	swapped = atomic.CompareAndSwapInt64(&value, 42, 200)
	fmt.Printf("  CAS(42→200): swapped=%t, value=%d (unchanged)\n", swapped, value)

	// ============================================
	// Example 4: Available atomic operations
	// ============================================
	fmt.Println("\n=== Available atomic operations ===")
	fmt.Println("  atomic.AddInt64(&v, delta)  — add/subtract")
	fmt.Println("  atomic.LoadInt64(&v)        — safe read")
	fmt.Println("  atomic.StoreInt64(&v, new)  — safe write")
	fmt.Println("  atomic.SwapInt64(&v, new)   — swap and return old")
	fmt.Println("  atomic.CompareAndSwapInt64(&v, old, new) — CAS")
	fmt.Println("  Works for int32, int64, uint32, uint64, uintptr, unsafe.Pointer")
}
