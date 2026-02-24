//go:build ignore

package main

import (
	"fmt"
	"time"
)

func main() {
	// ============================================
	// Example 1: Basic select — first channel wins
	// ============================================
	fmt.Println("=== Example 1: First channel ready wins ===")

	ch1 := make(chan string)
	ch2 := make(chan string)

	go func() {
		time.Sleep(200 * time.Millisecond)
		ch1 <- "result from ch1"
	}()

	go func() {
		time.Sleep(100 * time.Millisecond)
		ch2 <- "result from ch2"
	}()

	// select waits for whichever channel is ready first
	select {
	case msg := <-ch1:
		fmt.Println("  Got:", msg)
	case msg := <-ch2:
		fmt.Println("  Got:", msg) // ch2 is faster → this wins
	}

	// ============================================
	// Example 2: Select with send AND receive cases
	// ============================================
	fmt.Println("\n=== Example 2: Send and receive in select ===")

	outCh := make(chan int, 1)
	inCh := make(chan int, 1)
	inCh <- 42 // pre-load a value

	select {
	case outCh <- 99:
		fmt.Println("  Sent 99 to outCh")
	case val := <-inCh:
		fmt.Printf("  Received %d from inCh\n", val)
	}
	// Both are ready (outCh has room, inCh has value) — one is chosen RANDOMLY

	// ============================================
	// Example 3: Random fairness demonstration
	// ============================================
	fmt.Println("\n=== Example 3: Random fairness ===")

	a := make(chan string, 100)
	b := make(chan string, 100)

	for i := 0; i < 100; i++ {
		a <- "A"
		b <- "B"
	}

	countA, countB := 0, 0
	for i := 0; i < 100; i++ {
		select {
		case <-a:
			countA++
		case <-b:
			countB++
		}
	}
	fmt.Printf("  A chosen: %d times, B chosen: %d times (roughly 50/50)\n", countA, countB)

	// ============================================
	// Example 4: Empty select blocks forever
	// ============================================
	fmt.Println("\n=== Example 4: Empty select ===")
	fmt.Println("  select {} would block forever — used to keep servers alive")
	fmt.Println("  (Not running it here because it would hang!)")
	// select {} // uncomment to see it hang
}
