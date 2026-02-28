//go:build ignore

package main

import (
	"fmt"
	"time"
)

// ============================================
// Basic channel usage — sending and receiving between goroutines.
//
// Python equivalent:
//   q = asyncio.Queue()
//   await q.put(value)   // send
//   value = await q.get() // receive
//
// Go:
//   ch <- value    // send
//   value := <-ch  // receive
// ============================================

func main() {
	// ============================================
	// Example 1: Simple send and receive
	// ============================================
	fmt.Println("=== Example 1: Basic send/receive ===")

	ch := make(chan string) // unbuffered channel

	// Sender goroutine
	go func() {
		fmt.Println("Sender: about to send")
		ch <- "hello from goroutine!" // blocks until receiver is ready
		fmt.Println("Sender: value was received!")
	}()

	// Receiver (main goroutine)
	time.Sleep(500 * time.Millisecond) // simulate doing other work first
	msg := <-ch                        // receive — unblocks the sender
	fmt.Println("Receiver got:", msg)

	// ============================================
	// Example 2: Goroutine returning a result via channel
	// In Python asyncio: result = await task
	// In Go: result = <-ch
	// ============================================
	fmt.Println("\n=== Example 2: Getting results from goroutines ===")

	resultCh := make(chan int)

	go func() {
		// Simulate computation
		sum := 0
		for i := 1; i <= 100; i++ {
			sum += i
		}
		resultCh <- sum // send result back
	}()

	total := <-resultCh // wait for and receive result
	fmt.Println("Sum of 1..100 =", total)

	// ============================================
	// Example 3: Multiple senders, one receiver
	// ============================================
	fmt.Println("\n=== Example 3: Multiple senders ===")

	ch2 := make(chan string)

	// Launch 3 goroutines, each sending a message
	for i := 1; i <= 3; i++ {
		go func(id int) {
			time.Sleep(time.Duration(id*100) * time.Millisecond)
			ch2 <- fmt.Sprintf("worker %d reporting", id)
			fmt.Printf("Sent to %d\n",i)
		}(i)
	}

	// Receive all 3 messages (order depends on timing)
	for i := 0; i < 3; i++ {
		msg := <-ch2
		fmt.Printf("  Received: %s\n", msg)
	}

	// ============================================
	// Example 4: Channel as function return (generator pattern)
	// Like Python:
	//   def counter(n):
	//       for i in range(n):
	//           yield i
	// ============================================
	fmt.Println("\n=== Example 4: Channel as generator ===")

	for num := range counter(5) {
		fmt.Printf("  Got: %d\n", num)
	}
}

// counter returns a receive-only channel that yields 0..n-1
func counter(n int) <-chan int {
	ch := make(chan int)
	go func() {
		for i := 0; i < n; i++ {
			ch <- i
		}
		close(ch) // signal: no more values
	}()
	return ch
}
