//go:build ignore

package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// ============================================
// Preemption demo: Since Go 1.14, the runtime can preempt
// goroutines even in tight CPU loops.
//
// Before 1.14: a tight loop with no function calls would
// monopolize its OS thread — other goroutines starved.
//
// After 1.14: the runtime sends SIGURG to the thread,
// interrupting the loop and allowing other goroutines to run.
//
// Compare to asyncio: asyncio NEVER preempts. If a coroutine
// doesn't `await`, the event loop is blocked forever.
//
//   Python (broken):
//     async def hog():
//         while True: pass   # event loop dead, all tasks frozen
//
//   Go (works fine):
//     go func() {
//         for { }   // scheduler still preempts this
//     }()
// ============================================

func main() {
	runtime.GOMAXPROCS(1) // force single thread to show preemption clearly

	var wg sync.WaitGroup

	// Launch a tight-loop goroutine (CPU hog)
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("Hog: starting infinite-ish loop...")
		count := 0
		for i := 0; i < 1_000_000_000; i++ {
			count++ // tight loop — no function calls, no I/O
		}
		fmt.Printf("Hog: done (counted to %d)\n", count)
	}()

	// Launch several other goroutines
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			fmt.Printf("Worker %d: waiting to run...\n", id)
			time.Sleep(10 * time.Millisecond)
			fmt.Printf("Worker %d: got my turn! Preemption works.\n", id)
		}(i)
	}

	wg.Wait()
	fmt.Println("\nAll goroutines completed. Preemption allowed everyone to run!")
	fmt.Println("In Python asyncio, the workers would NEVER have gotten a turn.")
}
