//go:build ignore

package main

import (
	"fmt"
	"math/rand"
	"time"
)

// ============================================
// for-select loop: the Go "event loop" pattern
//
// This is Go's equivalent of Python's:
//   while True:
//       event = await get_next_event()
//       handle(event)
//
// But with native multi-channel support.
// ============================================

func main() {
	// ============================================
	// Example: Chat server event loop
	// Handles messages, errors, join/leave events, and shutdown
	// ============================================
	fmt.Println("=== Chat Server Event Loop ===")

	messages := make(chan string, 10)
	errors := make(chan error, 5)
	joins := make(chan string, 5)
	quit := make(chan struct{})

	// Simulate events from different sources
	go func() {
		users := []string{"Alice", "Bob", "Charlie"}
		for _, u := range users {
			time.Sleep(100 * time.Millisecond)
			joins <- u
		}
	}()

	go func() {
		msgs := []string{"Hello!", "How are you?", "Go is awesome", "Bye!"}
		for _, m := range msgs {
			time.Sleep(time.Duration(150+rand.Intn(200)) * time.Millisecond)
			messages <- m
		}
	}()

	go func() {
		time.Sleep(600 * time.Millisecond)
		errors <- fmt.Errorf("connection lost for user 42")
	}()

	// Shutdown after 2 seconds
	go func() {
		time.Sleep(2 * time.Second)
		close(quit)
	}()

	// THE EVENT LOOP — for + select
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case msg := <-messages:
			fmt.Printf("  [MSG]  %s\n", msg)

		case err := <-errors:
			fmt.Printf("  [ERR]  %v\n", err)

		case user := <-joins:
			fmt.Printf("  [JOIN] %s joined the chat\n", user)

		case <-ticker.C: // periodic task (like a cron within the loop)
			fmt.Println("  [TICK] Sending keepalive ping...")

		case <-quit:
			fmt.Println("  [QUIT] Shutting down event loop")
			return
		}
	}
}
