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
	// Example 3: Nested timeouts (OUTER times out BEFORE inner)
	// ============================================
	fmt.Println("\n=== Example 3: Parent context timeout cancels child ===")

	// OUTER: 300ms (will timeout first!)
	outerCtx, outerCancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer outerCancel()

	fmt.Println("  Outer timeout: 300ms (parent)")

	// INNER: 700ms (longer than outer, but won't matter!)
	innerCtx, innerCancel := context.WithTimeout(outerCtx, 700*time.Millisecond)
	defer innerCancel()

	fmt.Println("  Inner timeout: 700ms (child)")

	start3 := time.Now()

	// Try to call with inner context for 600ms
	// Even though 600ms < 700ms (inner timeout), the outer will cancel at 300ms!
	_, err = simulateAPICall(innerCtx, "inner-call", 600*time.Millisecond)
	elapsed := time.Since(start3).Round(time.Millisecond)

	if err != nil {
		fmt.Printf("  Inner call FAILED at %v: %v\n", elapsed, err)
		fmt.Println("  ⚠️  Inner was cancelled by OUTER timeout (300ms), not its own (700ms)!")
	} else {
		fmt.Println("  Inner call: succeeded")
	}

	fmt.Println("\n  KEY INSIGHT:")
	fmt.Println("  When a parent context times out, ALL child contexts are immediately cancelled.")
	fmt.Println("  The child's longer timeout (700ms) is irrelevant — parent wins at 300ms.")
}
