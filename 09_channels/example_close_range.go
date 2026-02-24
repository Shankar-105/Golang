//go:build ignore

package main

import "fmt"

// ============================================
// Closing channels and range loops.
//
// close(ch) signals: "no more values will be sent."
// Receivers can detect this with the value, ok := <-ch pattern
// or by using for-range which stops on close.
//
// Python has no equivalent — you use sentinel values:
//   await q.put(None)  # "I'm done"
// ============================================

func main() {
	// ============================================
	// Example 1: close() and the ok idiom
	// ============================================
	fmt.Println("=== Example 1: close() and ok idiom ===")

	ch := make(chan int, 5)
	ch <- 10
	ch <- 20
	close(ch)

	// Method 1: check if channel is still open
	val, ok := <-ch
	fmt.Printf("val=%d, ok=%t (channel had data)\n", val, ok)

	val, ok = <-ch
	fmt.Printf("val=%d, ok=%t (channel had data)\n", val, ok)

	val, ok = <-ch
	fmt.Printf("val=%d, ok=%t (channel closed + empty → zero value!)\n", val, ok)

	// ============================================
	// Example 2: range over channel (idiomatic)
	// ============================================
	fmt.Println("\n=== Example 2: range over channel ===")

	numbers := make(chan int)

	go func() {
		for i := 1; i <= 5; i++ {
			numbers <- i * i // send squares
		}
		close(numbers) // MUST close or range blocks forever!
	}()

	// range automatically stops when channel is closed
	for num := range numbers {
		fmt.Printf("  Received: %d\n", num)
	}
	fmt.Println("  Channel drained!")

	// ============================================
	// Example 3: Pipeline with close propagation
	// Each stage closes its output when done.
	// ============================================
	fmt.Println("\n=== Example 3: Pipeline ===")

	// Stage 1: generate
	gen := func(nums ...int) <-chan int {
		out := make(chan int)
		go func() {
			for _, n := range nums {
				out <- n
			}
			close(out)
		}()
		return out
	}

	// Stage 2: square
	sq := func(in <-chan int) <-chan int {
		out := make(chan int)
		go func() {
			for n := range in { // range stops when 'in' is closed
				out <- n * n
			}
			close(out) // propagate close to downstream
		}()
		return out
	}

	// Stage 3: add 1
	inc := func(in <-chan int) <-chan int {
		out := make(chan int)
		go func() {
			for n := range in {
				out <- n + 1
			}
			close(out)
		}()
		return out
	}

	// Connect the pipeline: generate → square → increment
	pipeline := inc(sq(gen(2, 3, 4, 5)))

	for result := range pipeline {
		fmt.Printf("  Result: %d\n", result)
	}
	// 2→4→5, 3→9→10, 4→16→17, 5→25→26

	// ============================================
	// Example 4: Don't close from receiver side!
	// ============================================
	fmt.Println("\n=== Rules for closing ===")
	fmt.Println("1. Only the SENDER should close a channel")
	fmt.Println("2. Sending to a closed channel PANICS")
	fmt.Println("3. Receiving from a closed channel returns zero value + false")
	fmt.Println("4. Closing is optional — channels get garbage collected")
	fmt.Println("5. Close only when receivers need to know 'no more data'")
}
