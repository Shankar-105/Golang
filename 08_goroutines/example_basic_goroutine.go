//go:build ignore

package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	// ============================================
	// Example 1: Simple goroutine — fire and forget
	// Python: asyncio.create_task(say_hello())
	// ============================================
	fmt.Println("=== Example 1: Basic goroutine ===")

	go sayHello("goroutine") // launched — main continues immediately
	sayHello("main")         // runs in main goroutine (synchronous)

	time.Sleep(100 * time.Millisecond) // give goroutine time to run (BAD practice!)

	// ============================================
	// Example 2: Multiple goroutines with WaitGroup
	// Python equivalent:
	//   tasks = [asyncio.create_task(process(i)) for i in range(5)]
	//   await asyncio.gather(*tasks)
	// ============================================
	fmt.Println("\n=== Example 2: WaitGroup (like asyncio.gather) ===")

	var wg sync.WaitGroup

	for i := 1; i <= 5; i++ {
		wg.Add(1) // increment counter BEFORE launching goroutine
		go func(id int) {
			defer wg.Done() // decrement counter when goroutine completes
			fmt.Printf("  Task %d: starting\n", id)
			time.Sleep(time.Duration(id*100) * time.Millisecond)
			fmt.Printf("  Task %d: done\n", id)
		}(i) // pass i as parameter to avoid closure trap
	}

	wg.Wait() // block until counter reaches 0 (all goroutines done)
	fmt.Println("All tasks complete!")

	// ============================================
	// Example 3: Goroutines execute in non-deterministic order
	// Unlike Python asyncio which starts tasks in creation order,
	// goroutine scheduling order is NOT guaranteed.
	// ============================================
	fmt.Println("\n=== Example 3: Non-deterministic ordering ===")

	var wg2 sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg2.Add(1)
		go func(n int) {
			defer wg2.Done()
			fmt.Printf("%d ", n)
		}(i)
	}
	wg2.Wait()
	fmt.Println("\n(order varies each run!)")

	// ============================================
	// Example 4: Returning values from goroutines
	// In asyncio: result = await task
	// In Go: goroutines can't "return" to the caller.
	// You use channels (Lesson 9) or shared state.
	// Here's a simple shared-state approach:
	// ============================================
	fmt.Println("\n=== Example 4: Collecting results ===")

	results := make([]string, 5)
	var wg3 sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg3.Add(1)
		go func(idx int) {
			defer wg3.Done()
			// Each goroutine writes to its own index — no race condition
			results[idx] = fmt.Sprintf("result-%d", idx)
		}(i)
	}

	wg3.Wait()
	fmt.Println("Results:", results)
}

func sayHello(caller string) {
	fmt.Printf("  Hello from %s!\n", caller)
}
