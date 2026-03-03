//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// ============================================
// Graceful shutdown — production-essential pattern
//
// Problem: If you just kill a server, in-flight requests get dropped.
// Solution: Catch the interrupt signal, stop accepting new connections,
//           wait for existing requests to finish, THEN exit.
//
// Python equivalent:
//   uvicorn has built-in graceful shutdown
//   signal.signal(signal.SIGINT, handler) for custom
//
// Run: go run example_graceful_shutdown.go
// Then press Ctrl+C to see graceful shutdown in action
// ============================================

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello! Server is running.")
	})

	mux.HandleFunc("GET /slow", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Starting slow operation...")
		// Simulate a long operation — try pressing Ctrl+C while this runs
		select {
		case <-time.After(10 * time.Second):
			fmt.Fprintln(w, "Slow operation completed!")
		case <-r.Context().Done():
			// Client disconnected OR server shutting down
			log.Println("Request cancelled:", r.Context().Err())
			return
		}
	})

	// Create server with explicit config
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ---- Start server in a goroutine ----
	go func() {
		log.Println("Server starting on :8080")
		log.Println("Try: curl http://localhost:8080/")
		log.Println("Try: curl http://localhost:8080/slow  (then Ctrl+C the server)")
		log.Println("Press Ctrl+C to initiate graceful shutdown")

		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// ---- Wait for interrupt signal ----
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt) // Ctrl+C
	<-quit // blocks until signal received

	log.Println("")
	log.Println("Shutdown signal received!")
	log.Println("Waiting for in-flight requests to finish (max 30s)...")

	// ---- Graceful shutdown ----
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Forced shutdown: %v", err)
	}

	log.Println("Server shut down gracefully. Goodbye!")
}
