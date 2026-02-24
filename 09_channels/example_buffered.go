//go:build ignore

package main

import (
	"fmt"
	"time"
)

// ============================================
// Buffered channels — like asyncio.Queue(maxsize=N)
//
// Key property: sends don't block until buffer is full,
// receives don't block until buffer is empty.
//
// Like a mailbox that can hold N letters. The mail carrier
// (sender) can drop off letters without waiting, unless the
// mailbox is full.
// ============================================

func main() {
	// ============================================
	// Example 1: Basic buffered channel
	// ============================================
	fmt.Println("=== Example 1: Buffered channel basics ===")

	ch := make(chan int, 3) // buffer holds 3 values

	// These don't block because buffer has room
	ch <- 10
	ch <- 20
	ch <- 30
	fmt.Printf("Buffer: len=%d, cap=%d\n", len(ch), cap(ch))

	// Receive in FIFO order
	fmt.Println(<-ch) // 10
	fmt.Println(<-ch) // 20
	fmt.Println(<-ch) // 30
	fmt.Printf("Buffer after drain: len=%d, cap=%d\n", len(ch), cap(ch))

	// ============================================
	// Example 2: Buffered channel as rate limiter
	//
	// Python equivalent:
	//   q = asyncio.Queue(maxsize=3)
	//   # producers slow down when queue is full
	// ============================================
	fmt.Println("\n=== Example 2: Backpressure / rate limiting ===")

	jobs := make(chan int, 3) // only 3 jobs can be queued

	// Fast producer
	go func() {
		for i := 1; i <= 10; i++ {
			fmt.Printf("  [producer] Sending job %d...\n", i)
			jobs <- i // blocks when buffer is full (backpressure!)
			fmt.Printf("  [producer] Job %d sent\n", i)
		}
		close(jobs)
	}()

	// Slow consumer
	for job := range jobs {
		fmt.Printf("  [consumer] Processing job %d\n", job)
		time.Sleep(300 * time.Millisecond) // slow processing
	}

	// ============================================
	// Example 3: Buffered channel as semaphore
	// Limit concurrency to N goroutines at a time.
	//
	// Python equivalent:
	//   sem = asyncio.Semaphore(3)
	//   async with sem:
	//       await do_work()
	// ============================================
	fmt.Println("\n=== Example 3: Semaphore pattern ===")

	sem := make(chan struct{}, 3) // max 3 concurrent workers
	done := make(chan struct{})

	for i := 1; i <= 8; i++ {
		go func(id int) {
			sem <- struct{}{} // acquire (blocks if 3 are already running)
			fmt.Printf("  Worker %d: started (concurrent: %d)\n", id, len(sem))
			time.Sleep(500 * time.Millisecond) // simulate work
			fmt.Printf("  Worker %d: done\n", id)
			<-sem // release

			if id == 8 {
				close(done)
			}
		}(i)
	}

	<-done
	time.Sleep(100 * time.Millisecond) // let stragglers print

	// ============================================
	// Example 4: len() and cap() for channels
	// ============================================
	fmt.Println("\n=== Example 4: Channel inspection ===")

	ch2 := make(chan string, 5)
	ch2 <- "a"
	ch2 <- "b"

	fmt.Printf("len(ch) = %d (items in buffer)\n", len(ch2))
	fmt.Printf("cap(ch) = %d (buffer capacity)\n", cap(ch2))

	// len() on unbuffered channel is always 0:
	unbuf := make(chan int)
	fmt.Printf("Unbuffered: len=%d, cap=%d\n", len(unbuf), cap(unbuf))
}
