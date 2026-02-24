//go:build ignore

package main

import (
	"fmt"
	"os"
)

// ============================================
// defer: Schedules a function call to run when the enclosing function returns.
//
// Python equivalent: context managers (with statement) + finally blocks.
//
//   Python:                          Go:
//   with open("f.txt") as f:        f, _ := os.Open("f.txt")
//       data = f.read()             defer f.Close()  ← runs on return
//                                   data, _ := io.ReadAll(f)
//
// Key rules:
// 1. Deferred calls execute in LIFO (stack) order.
// 2. Arguments are evaluated IMMEDIATELY (when defer is hit), not when it runs.
// 3. Deferred calls run even if the function panics.
// ============================================

func main() {
	demonstrateLIFO()
	fmt.Println()
	demonstrateArgumentEvaluation()
	fmt.Println()
	demonstratePracticalUse()
	fmt.Println()
	demonstrateLoopTrap()
}

// demonstrateLIFO shows that defers run in Last-In-First-Out order.
func demonstrateLIFO() {
	fmt.Println("=== LIFO Order ===")
	fmt.Println("start")

	defer fmt.Println("deferred 1") // pushed first → runs last
	defer fmt.Println("deferred 2") // pushed second
	defer fmt.Println("deferred 3") // pushed third → runs first

	fmt.Println("end")

	// Output:
	// start
	// end
	// deferred 3
	// deferred 2
	// deferred 1
}

// demonstrateArgumentEvaluation shows that defer captures argument values
// at the time the defer statement is reached, NOT when the deferred function runs.
func demonstrateArgumentEvaluation() {
	fmt.Println("=== Argument Evaluation (tricky!) ===")

	x := 10
	defer fmt.Println("deferred x =", x) // x is captured as 10 RIGHT NOW

	x = 20
	fmt.Println("current x =", x) // prints 20

	// Output:
	// current x = 20
	// deferred x = 10  ← captured the value at defer time!

	// To defer with the "current" value at run time, use a closure:
	y := 100
	defer func() {
		fmt.Println("closure y =", y) // captures the VARIABLE, not value
	}()
	y = 200

	// Output includes:
	// closure y = 200  ← closure sees the final value
}

// demonstratePracticalUse shows real-world defer usage.
func demonstratePracticalUse() {
	fmt.Println("=== Practical: File cleanup ===")

	// Create a temp file
	f, err := os.CreateTemp("", "example-*.txt")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}

	// Schedule cleanup IMMEDIATELY after successful open.
	// This ensures the file is closed no matter how the function exits:
	// - normal return
	// - early return due to error
	// - even a panic
	defer f.Close()
	defer os.Remove(f.Name()) // clean up the temp file too

	// Write some data
	_, err = f.WriteString("Hello, Go!")
	if err != nil {
		fmt.Println("Error writing:", err)
		return // deferred Close() and Remove() still run!
	}

	fmt.Println("Wrote to:", f.Name())
	fmt.Println("File will be closed and removed when this function returns.")
}

// demonstrateLoopTrap shows a common MISTAKE with defer in loops.
func demonstrateLoopTrap() {
	fmt.Println("=== ⚠ Defer in Loops (Common Trap) ===")

	// ❌ BAD: defer in a loop — the deferred calls pile up and only
	// execute when the FUNCTION returns, not when the loop iteration ends!
	//
	// for _, file := range files {
	//     f, _ := os.Open(file)
	//     defer f.Close()  // these all run at function end, not loop end!
	// }
	//
	// If you open 1000 files, ALL 1000 stay open until function returns.

	// ✅ GOOD: wrap the loop body in an anonymous function
	files := []string{"a.txt", "b.txt", "c.txt"}
	for _, file := range files {
		func() { // anonymous function — defer runs when THIS returns
			fmt.Println("Processing:", file)
			// f, err := os.Open(file)
			// if err != nil { return }
			// defer f.Close()  // closes at end of THIS anonymous function
		}()
	}
}
