//go:build ignore

package main

import (
	"fmt"
	"math/rand"
	"time"
)

// ============================================
// Timeout patterns using select + time.After
//
// Python equivalent:
//   result = await asyncio.wait_for(task, timeout=2.0)
//
// Go: select between the operation channel and a timeout channel.
// ============================================

func slowDatabaseQuery() <-chan string {
	ch := make(chan string)
	go func() {
		delay := time.Duration(rand.Intn(3000)) * time.Millisecond
		time.Sleep(delay) // random 0-3 second delay
		ch <- fmt.Sprintf("query result (took %v)", delay)
	}()
	return ch
}

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	// ============================================
	// Example 1: Simple timeout
	// ============================================
	fmt.Println("=== Example 1: Simple timeout ===")

	select {
	case result := <-slowDatabaseQuery():
		fmt.Println("  Success:", result)
	case <-time.After(1 * time.Second):
		fmt.Println("  TIMEOUT: query took longer than 1s")
	}

	// ============================================
	// Example 2: Timeout in a loop (per-iteration timeout)
	// ============================================
	fmt.Println("\n=== Example 2: Per-iteration timeout ===")

	ch := make(chan int)
	go func() {
		for i := 0; i < 5; i++ {
			time.Sleep(time.Duration(rand.Intn(800)) * time.Millisecond)
			ch <- i
		}
		close(ch)
	}()

	for i := 0; i < 5; i++ {
		select {
		case val, ok := <-ch:
			if !ok {
				fmt.Println("  Channel closed")
				break
			}
			fmt.Printf("  Received: %d\n", val)
		case <-time.After(500 * time.Millisecond):
			fmt.Println("  Timed out waiting for value")
		}
	}

	// ============================================
	// Example 3: Overall deadline (not per-operation)
	// "Give me results, but stop after 2 seconds total"
	// ============================================
	fmt.Println("\n=== Example 3: Overall deadline ===")

	resultCh := make(chan string)
	deadline := time.After(2 * time.Second) // fires once after 2s

	// Spawn workers that finish at different times
	for i := 1; i <= 5; i++ {
		go func(id int) {
			time.Sleep(time.Duration(id*600) * time.Millisecond)
			resultCh <- fmt.Sprintf("worker-%d done", id)
		}(i)
	}

	collected := 0
	for {
		select {
		case result := <-resultCh:
			collected++
			fmt.Printf("  Collected: %s (%d/5)\n", result, collected)
			if collected == 5 {
				fmt.Println("  All done!")
				return
			}
		case <-deadline:
			fmt.Printf("  DEADLINE HIT: only got %d/5 results in 2s\n", collected)
			return
		}
	}
}
