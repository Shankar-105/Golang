//go:build ignore

package main

import (
	"fmt"
	"time"
)

// ============================================
// Demonstration: main() exits = goroutines die
//
// This is the #1 beginner trap.
// In Python asyncio, asyncio.run(main()) keeps the loop alive
// until main() completes. In Go, when main() returns, EVERYTHING stops.
// ============================================

func backgroundWork(id int) {
	fmt.Printf("Goroutine %d: started\n", id)
	time.Sleep(2 * time.Second)
	fmt.Printf("Goroutine %d: finished\n", id) // may never print!
}

func main() {
	fmt.Println("=== Demo: main exits too early ===")

	for i := 1; i <= 3; i++ {
		go backgroundWork(i)
	}

	// Uncomment ONE of these approaches to fix:

	// Approach 1: time.Sleep — TERRIBLE, only for demos
	// time.Sleep(3 * time.Second)

	// Approach 2: select{} — blocks forever (good for servers)
	// select {}

	// Approach 3: sync.WaitGroup — the right way (see next example)

	fmt.Println("main() is returning... goroutines may be killed!")

	// Try running this: you'll see "started" messages but probably NOT "finished"
	// because main() returns before the 2-second sleep completes.
	//
	// Fix it by uncommenting Approach 1 and running again.
}
