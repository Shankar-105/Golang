//go:build ignore

package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// ============================================
// Middleware in Go
//
// Middleware = a function that wraps a handler to add behavior
// (logging, auth, CORS, rate limiting, recovery...)
//
// Pattern: func middleware(next http.Handler) http.Handler
//
// Python equivalent:
//   @app.middleware("http")
//   async def logging(request, call_next):
//       response = await call_next(request)
//       return response
//
// Run: go run example_middleware.go
// ============================================

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Home page")
	})

	mux.HandleFunc("GET /slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // simulate slow operation
		fmt.Fprintln(w, "Slow response (2s)")
	})

	mux.HandleFunc("GET /panic", func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong!") // recovery middleware catches this
	})

	mux.HandleFunc("GET /secret", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Top secret data! You are authorized.")
	})

	// ============================================
	// Stack middleware: request → logging → recovery → auth → handler
	// Order matters! Outermost runs first.
	// ============================================
	var handler http.Handler = mux
	handler = authMiddleware(handler)     // check auth (innermost)
	handler = recoveryMiddleware(handler) // catch panics
	handler = loggingMiddleware(handler)  // log requests (outermost)

	fmt.Println("=== Middleware Demo ===")
	fmt.Println("Listening on http://localhost:8080")
	fmt.Println("")
	fmt.Println("Try these:")
	fmt.Println("  curl http://localhost:8080/          → logged, no auth needed for /")
	fmt.Println("  curl http://localhost:8080/slow      → logged with duration")
	fmt.Println("  curl http://localhost:8080/panic     → recovered from panic")
	fmt.Println("  curl http://localhost:8080/secret    → 401 (no token)")
	fmt.Println("  curl http://localhost:8080/secret -H 'Authorization: Bearer my-token'")

	log.Fatal(http.ListenAndServe(":8080", handler))
}

// ============================================
// Middleware 1: Logging
// Logs method, path, status code, and duration
// ============================================
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.statusCode = code
	sr.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the ResponseWriter to capture the status code
		recorder := &statusRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // default
		}

		log.Printf("→ %s %s", r.Method, r.URL.Path)

		next.ServeHTTP(recorder, r) // call the next handler

		log.Printf("← %s %s [%d] (%s)",
			r.Method, r.URL.Path,
			recorder.statusCode,
			time.Since(start).Round(time.Millisecond),
		)
	})
}

// ============================================
// Middleware 2: Recovery (catch panics)
// Like Python's try/except around every request
// ============================================
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC recovered: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// ============================================
// Middleware 3: Simple Auth
// Checks for Authorization header on /secret routes
// ============================================
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only protect /secret paths
		if r.URL.Path == "/secret" {
			token := r.Header.Get("Authorization")
			if token == "" {
				http.Error(w, `{"error":"unauthorized: missing token"}`, http.StatusUnauthorized)
				return
			}
			// In real code, you'd validate the token here
			log.Printf("Auth: token received for %s", r.URL.Path)
		}

		next.ServeHTTP(w, r)
	})
}
