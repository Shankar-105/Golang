//go:build ignore

package main

import (
	"fmt"
	"sync"
	"time"
)

// ============================================
// Deadlock demonstrations
//
// Go's runtime detects when ALL goroutines are blocked:
//   "fatal error: all goroutines are asleep - deadlock!"
//
// But: if ANY goroutine is running, no detection occurs.
// Partial deadlocks are silent killers.
// ============================================

func main() {
	fmt.Println("=== Deadlock Examples ===")
	fmt.Println("(Each example is run safely — we detect and avoid the actual deadlock)")

	// ============================================
	// Example 1: Unbuffered channel with no reader
	// ============================================
	fmt.Println("\n--- Example 1: Channel deadlock (simulated) ---")
	fmt.Println("  ch := make(chan int)")
	fmt.Println("  ch <- 42  // DEADLOCK: no goroutine to receive!")
	fmt.Println("  Fix: use a goroutine to receive, or use buffered channel")

	// Safe demo:
	ch1 := make(chan int, 1) // buffered! Holds 1 value without blocking
	ch1 <- 42
	fmt.Printf("  Buffered fix: received %d\n", <-ch1)

	// ============================================
	// Example 2: Lock ordering deadlock
	// ============================================
	fmt.Println("\n--- Example 2: Lock ordering deadlock (simulated) ---")

	var muA, muB sync.Mutex

	// Dangerous pattern (DON'T ACTUALLY RUN THIS — it deadlocks!)
	fmt.Println("  Goroutine 1: Lock(A) → Lock(B)")
	fmt.Println("  Goroutine 2: Lock(B) → Lock(A)")
	fmt.Println("  Result: DEADLOCK — each waits for the other's lock")

	// Safe demo: always lock in the same order
	fmt.Println("  Fix: Always lock in order A → B")
	var wg sync.WaitGroup

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			muA.Lock() // always A first
			muB.Lock() // then B
			fmt.Printf("  Goroutine %d: got both locks\n", id)
			muB.Unlock()
			muA.Unlock()
		}(i)
	}
	wg.Wait()

	// ============================================
	// Example 3: Self-deadlock (non-reentrant mutex)
	// ============================================
	fmt.Println("\n--- Example 3: Non-reentrant mutex ---")
	fmt.Println("  mu.Lock()")
	fmt.Println("  mu.Lock()  // DEADLOCK! Go's Mutex is NOT re-entrant")
	fmt.Println("  Python's threading.RLock() IS re-entrant. Go's is NOT.")
	fmt.Println("  Fix: Restructure code so a locked function never calls another locked function")

	// ============================================
	// Example 4: Goroutine waits on itself via channel
	// ============================================
	fmt.Println("\n--- Example 4: Select deadlock ---")

	// This is a subtle deadlock:
	ch := make(chan int)
	timeout := time.After(100 * time.Millisecond)

	select {
	case val := <-ch:
		fmt.Println("  Got:", val)
	case <-timeout:
		fmt.Println("  Timeout prevented deadlock (no sender on ch)")
	}

	// ============================================
	// Example 5: WaitGroup misuse
	// ============================================
	fmt.Println("\n--- Example 5: WaitGroup misuse ---")
	fmt.Println("  wg.Add(1)")
	fmt.Println("  wg.Wait()  // DEADLOCK: nobody calls wg.Done()!")
	fmt.Println("  Fix: ensure every Add has a corresponding Done")

	// Safe demo:
	var wg2 sync.WaitGroup
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		fmt.Println("  Worker completed → wg.Done() called")
	}()
	wg2.Wait()
	fmt.Println("  WaitGroup resolved correctly")

	// ============================================
	// Summary
	// ============================================
	fmt.Println("\n=== Deadlock Prevention Rules ===")
	fmt.Println("  1. Always use buffered channels or launch a goroutine before sending")
	fmt.Println("  2. Lock mutexes in a consistent global order")
	fmt.Println("  3. Never lock a mutex twice (Go's Mutex is NOT re-entrant)")
	fmt.Println("  4. Every wg.Add() must have a wg.Done()")
	fmt.Println("  5. Use timeouts/context to prevent indefinite blocking")
}