//go:build ignore

package main

import "fmt"

// ============================================
// Non-blocking channel operations using select + default
//
// Without `default`: select blocks until a case is ready.
// With `default`: select returns immediately if no case is ready.
//
// Python equivalent:
//   q.get_nowait()  → raises Empty
//   q.put_nowait()  → raises Full
// ============================================

func main() {
	// ============================================
	// Example 1: Non-blocking receive
	// ============================================
	fmt.Println("=== Example 1: Non-blocking receive ===")

	messages := make(chan string, 5)

	// Try to receive — channel is empty
	select {
	case msg := <-messages:
		fmt.Println("  Got:", msg)
	default:
		fmt.Println("  No message available (non-blocking)")
	}

	// Now put something in and try again
	messages <- "hello"

	select {
	case msg := <-messages:
		fmt.Println("  Got:", msg) // this time it succeeds
	default:
		fmt.Println("  No message available")
	}

	// ============================================
	// Example 2: Non-blocking send
	// ============================================
	fmt.Println("\n=== Example 2: Non-blocking send ===")

	ch := make(chan int, 2) // buffer of 2
	ch <- 1
	ch <- 2 // buffer full

	select {
	case ch <- 3:
		fmt.Println("  Sent 3")
	default:
		fmt.Println("  Channel full! Dropped message 3 (non-blocking)")
	}

	// ============================================
	// Example 3: Try-send pattern for dropping old messages
	// Common in real-time systems: if consumer is slow, drop old data
	// ============================================
	fmt.Println("\n=== Example 3: Drop-on-full pattern ===")

	updates := make(chan string, 1) // tiny buffer

	// Simulate fast producer, slow consumer
	for i := 0; i < 5; i++ {
		msg := fmt.Sprintf("update-%d", i)
		select {
		case updates <- msg:
			fmt.Printf("  Queued: %s\n", msg)
		default:
			// Buffer full — drop the OLD message, send new one
			<-updates // drain old
			updates <- msg
			fmt.Printf("  Replaced old with: %s\n", msg)
		}
	}

	// Consumer gets the latest
	fmt.Printf("  Consumer got: %s\n", <-updates)

	// ============================================
	// Example 4: Polling pattern (check multiple sources without blocking)
	// ============================================
	fmt.Println("\n=== Example 4: Polling ===")

	ch1 := make(chan string, 1)
	ch2 := make(chan string, 1)

	// Only ch1 has data
	ch1 <- "data from source 1"

	// Poll: check which sources have data right now
	for i := 0; i < 3; i++ {
		select {
		case msg := <-ch1:
			fmt.Printf("  Poll %d: ch1 = %s\n", i, msg)
		case msg := <-ch2:
			fmt.Printf("  Poll %d: ch2 = %s\n", i, msg)
		default:
			fmt.Printf("  Poll %d: nothing ready\n", i)
		}
	}
}
