//go:build ignore

package main

import (
	"context"
	"fmt"
	"time"
)

// ============================================
// Context timeout patterns for simulated API calls
//
// These patterns are used EVERYWHERE in Go backend code:
// - HTTP handlers
// - Database queries
// - gRPC calls
// - Microservice communication
// ============================================

// simulateAPICall pretends to call an external service
func simulateAPICall(ctx context.Context, service string, delay time.Duration) (string, error) {
	select {
	case <-time.After(delay):
		return fmt.Sprintf("response from %s", service), nil
	case <-ctx.Done():
		return "", fmt.Errorf("%s: %w", service, ctx.Err())
	}
}

func main() {
	// ============================================
	// Example 1: Sequential API calls with shared timeout
	// ============================================
	fmt.Println("=== Example 1: Sequential calls, shared timeout ===")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	start := time.Now()

	// Call 1: fast (200ms)
	r1, err := simulateAPICall(ctx, "user-service", 200*time.Millisecond)
	if err != nil {
		fmt.Printf("  Call 1 failed: %v\n", err)
	} else {
		fmt.Printf("  Call 1: %s (%v)\n", r1, time.Since(start).Round(time.Millisecond))
	}

	// Call 2: medium (400ms) — still within timeout
	r2, err := simulateAPICall(ctx, "order-service", 400*time.Millisecond)
	if err != nil {
		fmt.Printf("  Call 2 failed: %v\n", err)
	} else {
		fmt.Printf("  Call 2: %s (%v)\n", r2, time.Since(start).Round(time.Millisecond))
	}

	// Call 3: slow (600ms) — timeout fires because 200+400+600 > 1000ms
	r3, err := simulateAPICall(ctx, "payment-service", 600*time.Millisecond)
	if err != nil {
		fmt.Printf("  Call 3 failed: %v (%v)\n", err, time.Since(start).Round(time.Millisecond))
	} else {
		fmt.Printf("  Call 3: %s\n", r3)
	}

	// ============================================
	// Example 2: Parallel API calls with context
	// ============================================
	fmt.Println("\n=== Example 2: Parallel calls with timeout ===")

	ctx2, cancel2 := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel2()

	type apiResult struct {
		Service string
		Data    string
		Err     error
	}

	resultCh := make(chan apiResult, 3)

	// Launch 3 parallel calls
	services := []struct {
		name  string
		delay time.Duration
	}{
		{"fast-api", 100 * time.Millisecond},
		{"medium-api", 300 * time.Millisecond},
		{"slow-api", 800 * time.Millisecond}, // will timeout
	}

	for _, svc := range services {
		go func(name string, delay time.Duration) {
			data, err := simulateAPICall(ctx2, name, delay)
			resultCh <- apiResult{Service: name, Data: data, Err: err}
		}(svc.name, svc.delay)
	}

	start2 := time.Now()
	for i := 0; i < 3; i++ {
		r := <-resultCh
		if r.Err != nil {
			fmt.Printf("  %s: FAILED (%v) at %v\n", r.Service, r.Err, time.Since(start2).Round(time.Millisecond))
		} else {
			fmt.Printf("  %s: %s at %v\n", r.Service, r.Data, time.Since(start2).Round(time.Millisecond))
		}
	}

	// ============================================
	// Example 3: Nested timeouts (inner tighter than outer)
	// ============================================
	fmt.Println("\n=== Example 3: Nested timeouts ===")

	outerCtx, outerCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer outerCancel()

	fmt.Println("  Outer timeout: 2s")

	// Inner is tighter — 500ms
	innerCtx, innerCancel := context.WithTimeout(outerCtx, 500*time.Millisecond)
	defer innerCancel()

	fmt.Println("  Inner timeout: 500ms")

	_, err = simulateAPICall(innerCtx, "inner-call", 800*time.Millisecond)
	fmt.Printf("  Inner call result: %v\n", err) // deadline exceeded (500ms)

	// Outer is still alive!
	_, err = simulateAPICall(outerCtx, "outer-call", 300*time.Millisecond)
	if err != nil {
		fmt.Printf("  Outer call: %v\n", err)
	} else {
		fmt.Println("  Outer call: succeeded (outer context still valid)")
	}
}
