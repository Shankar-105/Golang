//go:build ignore

package main

import (
	"fmt"
	"sync"
	"time"
)

// ============================================
// Unbuffered channels — synchronous rendezvous
//
// Key property: sender blocks until receiver is ready,
// and receiver blocks until sender is ready.
// It's a DIRECT hand-off. Zero buffer capacity.
//
// Like two people exchanging a package: both must be
// at the meeting point at the same time.
// ============================================

func main() {
	// ============================================
	// Example 1: Demonstrating the synchronous nature
	// ============================================
	fmt.Println("=== Example 1: Synchronous hand-off ===")

	ch := make(chan string) // unbuffered!

	go func() {
		fmt.Println("[sender] About to send... (will block until receiver ready)")
		start := time.Now()
		ch <- "the package"
		fmt.Printf("[sender] Sent! Waited %v for receiver\n", time.Since(start))
	}()

	// Deliberately delay the receive to show the sender blocks
	time.Sleep(1 * time.Second)
	fmt.Println("[receiver] Ready to receive now!")
	msg := <-ch
	fmt.Printf("[receiver] Got: %q\n\n", msg)

	// ============================================
	// Example 2: Unbuffered channel as synchronization tool
	// No WaitGroup needed — the channel receive IS the wait.
	// ============================================
	fmt.Println("=== Example 2: Channel as synchronization ===")

	done := make(chan struct{}) // empty struct = zero bytes. Pure signal.

	go func() {
		fmt.Println("  Worker: doing heavy computation...")
		time.Sleep(500 * time.Millisecond)
		fmt.Println("  Worker: done!")
		close(done) // signal completion (closing is idiomatic for "done" signals)
	}()

	<-done // blocks until channel is closed
	fmt.Println("  Main: worker finished!")

	// ============================================
	// Example 3: Ping-pong between two goroutines
	// ============================================
	fmt.Println("=== Example 3: Ping-pong ===")

	ping := make(chan string)
	pong := make(chan string)

	// Player 1: receives on ping, sends on pong
	go func() {
		for msg := range ping {
			fmt.Printf("  Player 1 received: %s\n", msg)
			pong <- "pong"
		}
	}()

	// Main plays as Player 2
	for i := 0; i < 3; i++ {
		ping <- "ping"
		msg := <-pong
		fmt.Printf("  Player 2 received: %s\n", msg)
	}
	close(ping) // tell Player 1 we're done

	// ============================================
	// Example 4: DEADLOCK demonstration
	// ============================================
	fmt.Println("\n=== Example 4: Deadlock scenario (commented out) ===")
	fmt.Println("Uncomment the code below and run to see the deadlock panic:")
	fmt.Println("  ch := make(chan int)")
	fmt.Println("  ch <- 42  // DEADLOCK: no goroutine to receive!")
	fmt.Println("  // fatal error: all goroutines are asleep - deadlock!")

	// Uncomment to see the deadlock:
	// deadlockCh := make(chan int)
	// deadlockCh <- 42

	// ============================================
	// Example 5: Using unbuffered channels for ordered execution
	// ============================================
	fmt.Println("\n=== Example 5: Ordered execution via channels ===")

	step1Done := make(chan struct{})
	step2Done := make(chan struct{})

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-step1Done // wait for step 1
		fmt.Println("  Step 2: processing (after step 1)")
		close(step2Done)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-step2Done // wait for step 2
		fmt.Println("  Step 3: finalizing (after step 2)")
	}()

	// Step 1 runs first
	fmt.Println("  Step 1: initializing")
	time.Sleep(200 * time.Millisecond)
	close(step1Done)

	wg.Wait()
	fmt.Println("  All steps complete in order!")
}
