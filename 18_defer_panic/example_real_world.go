//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"runtime/debug"
	"sync"
	"time"
)

// ──────────────────────────────────────────────────────────────
// Real-world patterns combining defer, panic, and recover
// ──────────────────────────────────────────────────────────────

func main() {
	fmt.Println("═══ 1. HTTP Recovery Middleware ═══")
	httpRecoveryDemo()

	fmt.Println("\n═══ 2. Must-style Initialization ═══")
	mustInitDemo()

	fmt.Println("\n═══ 3. Safe Goroutine Launcher ═══")
	safeGoroutineDemo()

	fmt.Println("\n═══ 4. Transaction Pattern (defer Rollback) ═══")
	transactionDemo()

	fmt.Println("\n═══ 5. Panic-to-Error Wrapper ═══")
	panicToErrorDemo()

	fmt.Println("\n═══ 6. defer for Metrics/Logging ═══")
	metricsDemo()
}

// ════════════════════════════════════════════════════════════════
// 1. HTTP Recovery Middleware
// ════════════════════════════════════════════════════════════════
// The #1 real-world use of recover: preventing one bad request
// from crashing your entire web server.

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				// Log the panic with stack trace
				log.Printf("PANIC recovered: %v\nStack:\n%s", rec, debug.Stack())

				// Return 500 to the client
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "internal server error",
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func httpRecoveryDemo() {
	// Handler that panics (simulating a bug)
	buggyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This panic would normally crash the server!
		panic("nil pointer dereference in handler")
	})

	// Wrap with recovery middleware
	safe := recoveryMiddleware(buggyHandler)

	// Test it
	req := httptest.NewRequest("GET", "/crash", nil)
	w := httptest.NewRecorder()
	safe.ServeHTTP(w, req)

	fmt.Printf("  Status: %d (not crashed!)\n", w.Code)
	fmt.Printf("  Body: %s", w.Body.String())

	// Normal handler still works
	normalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	safeNormal := recoveryMiddleware(normalHandler)

	req2 := httptest.NewRequest("GET", "/health", nil)
	w2 := httptest.NewRecorder()
	safeNormal.ServeHTTP(w2, req2)
	fmt.Printf("  Normal request: %s", w2.Body.String())
}

// ════════════════════════════════════════════════════════════════
// 2. Must-style Initialization
// ════════════════════════════════════════════════════════════════
// Convention: functions prefixed with "Must" panic instead of
// returning errors. Use ONLY for program initialization.

func MustParseJSON(data string) map[string]any {
	var result map[string]any
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		panic(fmt.Sprintf("MustParseJSON: %v", err))
	}
	return result
}

func MustLoadConfig() map[string]any {
	// In real code, this would read from a file
	config := `{
		"host": "localhost",
		"port": 8080,
		"debug": true
	}`
	return MustParseJSON(config)
}

func mustInitDemo() {
	// ✅ Safe to use at program startup
	config := MustLoadConfig()
	fmt.Printf("  Config loaded: host=%s, port=%v\n",
		config["host"], config["port"])

	// ✅ Will panic with clear message for invalid JSON
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("  Caught init panic: %v\n", r)
		}
	}()
	MustParseJSON(`{"invalid json`) // panics
}

// ════════════════════════════════════════════════════════════════
// 3. Safe Goroutine Launcher
// ════════════════════════════════════════════════════════════════
// A panic in a goroutine kills the WHOLE PROGRAM.
// This wrapper prevents that.

func safeGo(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Goroutine panicked: %v\nStack: %s", r, debug.Stack())
			}
		}()
		fn()
	}()
}

func safeGoroutineDemo() {
	var wg sync.WaitGroup

	// Launch goroutines that might panic
	for i := 0; i < 3; i++ {
		wg.Add(1)
		i := i
		safeGo(func() {
			defer wg.Done()
			if i == 1 {
				panic(fmt.Sprintf("goroutine %d panicked!", i))
			}
			fmt.Printf("  Goroutine %d completed safely\n", i)
		})
	}

	wg.Wait()
	fmt.Println("  All goroutines done, program still alive!")
}

// ════════════════════════════════════════════════════════════════
// 4. Transaction Pattern
// ════════════════════════════════════════════════════════════════
// defer Rollback + explicit Commit

type FakeTransaction struct {
	committed bool
	operations []string
}

func (tx *FakeTransaction) Execute(op string) error {
	fmt.Printf("    Executing: %s\n", op)
	tx.operations = append(tx.operations, op)
	return nil
}

func (tx *FakeTransaction) Commit() error {
	tx.committed = true
	fmt.Println("    ✓ Transaction committed")
	return nil
}

func (tx *FakeTransaction) Rollback() {
	if tx.committed {
		return // already committed, rollback is a no-op
	}
	fmt.Println("    ✗ Transaction rolled back!")
}

func doTransaction(shouldFail bool) error {
	tx := &FakeTransaction{}
	defer tx.Rollback() // always runs, but no-op if committed

	if err := tx.Execute("INSERT INTO users ..."); err != nil {
		return err
	}

	if shouldFail {
		return fmt.Errorf("simulated failure")
	}

	if err := tx.Execute("UPDATE counters ..."); err != nil {
		return err
	}

	return tx.Commit() // if we reach here, commit
}

func transactionDemo() {
	fmt.Println("  Successful transaction:")
	if err := doTransaction(false); err != nil {
		fmt.Printf("    Error: %v\n", err)
	}

	fmt.Println("\n  Failed transaction:")
	if err := doTransaction(true); err != nil {
		fmt.Printf("    Error: %v\n", err)
	}
}

// ════════════════════════════════════════════════════════════════
// 5. Panic-to-Error Wrapper
// ════════════════════════════════════════════════════════════════
// Generic wrapper to call any function and convert panics to errors.

func catchPanic(fn func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				err = fmt.Errorf("caught panic: %w", v)
			case string:
				err = fmt.Errorf("caught panic: %s", v)
			default:
				err = fmt.Errorf("caught panic: %v", v)
			}
		}
	}()

	fn()
	return nil
}

func panicToErrorDemo() {
	// Panicking function wrapped to return error
	err := catchPanic(func() {
		panic("unexpected nil pointer")
	})
	fmt.Printf("  Panic converted to error: %v\n", err)

	// Non-panicking function
	err = catchPanic(func() {
		fmt.Println("  This function is fine")
	})
	fmt.Printf("  No panic: error = %v\n", err)
}

// ════════════════════════════════════════════════════════════════
// 6. defer for Metrics/Logging
// ════════════════════════════════════════════════════════════════

type RequestMetrics struct {
	mu       sync.Mutex
	requests map[string]time.Duration
}

func NewRequestMetrics() *RequestMetrics {
	return &RequestMetrics{
		requests: make(map[string]time.Duration),
	}
}

func (m *RequestMetrics) Track(endpoint string) func() {
	start := time.Now()
	return func() {
		duration := time.Since(start)
		m.mu.Lock()
		m.requests[endpoint] = duration
		m.mu.Unlock()
		fmt.Printf("    %s took %v\n", endpoint, duration)
	}
}

func (m *RequestMetrics) Report() {
	m.mu.Lock()
	defer m.mu.Unlock()
	fmt.Println("  Metrics report:")
	for endpoint, duration := range m.requests {
		fmt.Printf("    %-20s %v\n", endpoint, duration)
	}
}

func metricsDemo() {
	metrics := NewRequestMetrics()

	// Simulate tracked requests
	func() {
		defer metrics.Track("/api/users")()
		time.Sleep(10 * time.Millisecond)
	}()

	func() {
		defer metrics.Track("/api/orders")()
		time.Sleep(20 * time.Millisecond)
	}()

	func() {
		defer metrics.Track("/api/health")()
		// Very fast
	}()

	metrics.Report()
}
