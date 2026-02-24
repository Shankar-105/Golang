//go:build ignore

package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// ============================================
// Context basics: WithCancel, WithTimeout, WithDeadline
//
// Python equivalent:
//   asyncio.wait_for(coro, timeout=5.0)
//   task.cancel()
// ============================================

func main() {
	// ============================================
	// Example 1: context.WithCancel — manual cancellation
	// ============================================
	fmt.Println("=== Example 1: WithCancel ===")

	ctx, cancel := context.WithCancel(context.Background())

	// Worker that checks for cancellation
	go func(ctx context.Context) {
		for i := 1; ; i++ {
			select {
			case <-ctx.Done():
				fmt.Printf("  Worker stopped: %v (after %d iterations)\n", ctx.Err(), i-1)
				return
			default:
				fmt.Printf("  Worker iteration %d\n", i)
				time.Sleep(200 * time.Millisecond)
			}
		}
	}(ctx)

	// Let it run for a bit, then cancel
	time.Sleep(1 * time.Second)
	fmt.Println("  → Calling cancel()...")
	cancel()
	time.Sleep(100 * time.Millisecond) // give worker time to print

	// ============================================
	// Example 2: context.WithTimeout — auto-cancel after duration
	// ============================================
	fmt.Println("\n=== Example 2: WithTimeout ===")

	ctx2, cancel2 := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel2() // always defer cancel, even with timeout

	result, err := slowOperation(ctx2)
	if err != nil {
		fmt.Printf("  Operation failed: %v\n", err)
	} else {
		fmt.Printf("  Result: %s\n", result)
	}

	// ============================================
	// Example 3: context.WithDeadline — cancel at specific time
	// ============================================
	fmt.Println("\n=== Example 3: WithDeadline ===")

	deadline := time.Now().Add(300 * time.Millisecond)
	ctx3, cancel3 := context.WithDeadline(context.Background(), deadline)
	defer cancel3()

	fmt.Printf("  Deadline set to: %v from now\n", time.Until(deadline).Round(time.Millisecond))

	select {
	case <-time.After(1 * time.Second):
		fmt.Println("  Operation completed (shouldn't happen)")
	case <-ctx3.Done():
		fmt.Printf("  Context expired: %v\n", ctx3.Err())
	}

	// ============================================
	// Example 4: context.WithValue — request-scoped data
	// ============================================
	fmt.Println("\n=== Example 4: WithValue ===")

	type contextKey string
	const requestIDKey contextKey = "requestID"
	const userIDKey contextKey = "userID"

	ctx4 := context.WithValue(context.Background(), requestIDKey, "req-abc-123")
	ctx4 = context.WithValue(ctx4, userIDKey, 42)

	handleRequest(ctx4)

	// ============================================
	// Example 5: Cancellation propagates to children
	// ============================================
	fmt.Println("\n=== Example 5: Cancellation tree ===")

	parentCtx, parentCancel := context.WithCancel(context.Background())

	// Child contexts inherit parent's cancellation
	child1Ctx, child1Cancel := context.WithCancel(parentCtx)
	defer child1Cancel()
	child2Ctx, child2Cancel := context.WithTimeout(parentCtx, 10*time.Second)
	defer child2Cancel()

	go watchContext("child1", child1Ctx)
	go watchContext("child2", child2Ctx)

	time.Sleep(200 * time.Millisecond)
	fmt.Println("  → Cancelling PARENT context...")
	parentCancel() // cancels child1 AND child2!
	time.Sleep(100 * time.Millisecond)
	_ = child2Ctx // use variable
}

func slowOperation(ctx context.Context) (string, error) {
	// Simulate a slow operation (200-800ms)
	duration := time.Duration(200+rand.Intn(600)) * time.Millisecond
	fmt.Printf("  slowOperation will take %v\n", duration)

	select {
	case <-time.After(duration):
		return "success!", nil
	case <-ctx.Done():
		return "", ctx.Err() // context.DeadlineExceeded
	}
}

func handleRequest(ctx context.Context) {
	type contextKey string
	const requestIDKey contextKey = "requestID"
	const userIDKey contextKey = "userID"

	reqID := ctx.Value(requestIDKey).(string)
	userID := ctx.Value(userIDKey).(int)
	fmt.Printf("  Handling request %s for user %d\n", reqID, userID)
}

func watchContext(name string, ctx context.Context) {
	<-ctx.Done()
	fmt.Printf("  %s cancelled: %v\n", name, ctx.Err())
}
