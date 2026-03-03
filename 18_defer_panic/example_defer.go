//go:build ignore

package main

import (
	"fmt"
	"time"
)

// ──────────────────────────────────────────────────────────────
// defer Basics — scheduling cleanup functions
//
// Key rules:
// 1. Deferred calls run when the surrounding function returns
// 2. Multiple defers execute in LIFO (stack) order
// 3. Arguments are evaluated immediately (not at execution time)
// 4. Deferred functions can read/modify named return values
// ──────────────────────────────────────────────────────────────

func main() {
	fmt.Println("═══ 1. Basic defer ═══")
	basicDefer()

	fmt.Println("\n═══ 2. LIFO Order ═══")
	lifoOrder()

	fmt.Println("\n═══ 3. Arguments Evaluated Immediately ═══")
	argsEvaluatedImmediately()

	fmt.Println("\n═══ 4. Closure vs Value Capture ═══")
	closureVsValue()

	fmt.Println("\n═══ 5. Defer in Loops ═══")
	deferInLoops()

	fmt.Println("\n═══ 6. Named Returns + Defer ═══")
	result := namedReturnDefer()
	fmt.Println("  Final result:", result)

	fmt.Println("\n═══ 7. Defer with Methods ═══")
	deferWithMethods()

	fmt.Println("\n═══ 8. Timing with Defer ═══")
	timedFunction()
}

// ──── 1. Basic defer ────────────────────────────────────────
func basicDefer() {
	fmt.Println("  start")
	defer fmt.Println("  deferred (runs last)")
	fmt.Println("  middle")
	fmt.Println("  end")
	// Output: start, middle, end, deferred (runs last)
}

// ──── 2. LIFO Order ─────────────────────────────────────────
// Like a stack: last defer in = first to execute
// This is why: you acquire resource A, then B, then C.
// On cleanup: release C first, then B, then A. (Reverse order!)
func lifoOrder() {
	fmt.Println("  Deferring 1, 2, 3...")
	defer fmt.Println("  defer 1 (pushed first → runs last)")
	defer fmt.Println("  defer 2 (pushed second → runs middle)")
	defer fmt.Println("  defer 3 (pushed last → runs first)")
	fmt.Println("  Function body done")
}

// ──── 3. Arguments Evaluated Immediately ────────────────────
func argsEvaluatedImmediately() {
	x := 10
	defer fmt.Printf("  Deferred with x=%d (captured at defer time)\n", x)
	x = 20
	x = 30
	fmt.Printf("  x is now %d\n", x)
	// The deferred call prints x=10, not 30!
}

// ──── 4. Closure vs Value Capture ───────────────────────────
func closureVsValue() {
	// Value capture (argument): captures VALUE at defer time
	x := 100
	defer fmt.Printf("  Value capture: x=%d\n", x) // x=100

	// Closure capture: captures the VARIABLE, reads it at execution time
	defer func() {
		fmt.Printf("  Closure capture: x=%d\n", x) // x=300
	}()

	x = 200
	x = 300
	fmt.Printf("  x is now %d\n", x)
}

// ──── 5. Defer in Loops ─────────────────────────────────────
// WARNING: defers don't run until the function returns!
// In a loop, they all stack up = potential resource leak.
func deferInLoops() {
	// ❌ BAD: defers accumulate for entire function
	fmt.Println("  Defers in loop (all stack up, run at function end):")
	for i := 0; i < 3; i++ {
		defer fmt.Printf("    deferred i=%d\n", i)
	}
	// These ALL run when deferInLoops returns, not at end of each iteration!

	// ✅ GOOD: wrap in a helper function so defer runs each iteration
	fmt.Println("\n  Better pattern: wrap in a function")
	for i := 0; i < 3; i++ {
		func(n int) {
			defer fmt.Printf("    deferred (inner func) n=%d\n", n)
			fmt.Printf("    processing n=%d\n", n)
		}(i)
	}
}

// ──── 6. Named Returns + Defer ──────────────────────────────
// Deferred functions can READ and MODIFY named return values!
// This is a powerful pattern for error handling.
func namedReturnDefer() (result string) {
	result = "original"
	defer func() {
		// This MODIFIES the return value!
		result = result + " (modified by defer)"
	}()
	result = "computed"
	return result
	// Returns: "computed (modified by defer)"
}

// ──── 7. Defer with Methods ─────────────────────────────────
type Resource struct {
	Name string
}

func (r *Resource) Open() {
	fmt.Printf("    Opening %s\n", r.Name)
}

func (r *Resource) Close() {
	fmt.Printf("    Closing %s\n", r.Name)
}

func deferWithMethods() {
	db := &Resource{Name: "database"}
	cache := &Resource{Name: "cache"}
	conn := &Resource{Name: "connection"}

	db.Open()
	defer db.Close()

	cache.Open()
	defer cache.Close()

	conn.Open()
	defer conn.Close()

	fmt.Println("    All resources open, doing work...")
	// Closes in reverse order: connection, cache, database
}

// ──── 8. Timing with Defer ──────────────────────────────────
// A neat trick: use defer to measure function execution time.

func timeTrack(name string) func() {
	start := time.Now()
	fmt.Printf("  [%s] Starting...\n", name)
	return func() {
		fmt.Printf("  [%s] Done in %v\n", name, time.Since(start))
	}
}

func timedFunction() {
	defer timeTrack("timedFunction")()
	// Note: timeTrack runs NOW (captures start time)
	//       the returned func runs LATER (on defer)

	// Simulate work
	time.Sleep(50 * time.Millisecond)
	fmt.Println("  Working...")
}
