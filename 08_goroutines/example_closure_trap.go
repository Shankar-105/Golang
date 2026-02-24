//go:build ignore

package main

import (
	"fmt"
	"sync"
)

// ============================================
// The closure variable capture trap.
// This is IDENTICAL to the Python closure bug:
//
// Python bug:
//   funcs = []
//   for i in range(5):
//       funcs.append(lambda: print(i))  # all print 4!
//
// Go bug:
//   for i := 0; i < 5; i++ {
//       go func() { fmt.Println(i) }()  // all print 5!
//   }
// ============================================

func main() {
	var wg sync.WaitGroup

	// ============================================
	// ❌ BUG: closure captures variable `i` by reference
	// ============================================
	fmt.Println("=== ❌ BUG: closure captures variable by reference ===")

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// By the time this goroutine runs, the loop is probably done
			// and `i` is 5 (the loop's exit value).
			fmt.Printf("  i = %d\n", i) // likely prints 5 five times
		}()
	}
	wg.Wait()

	// ============================================
	// ✅ FIX 1: Pass as function argument (captures by value)
	// ============================================
	fmt.Println("\n=== ✅ FIX 1: Pass as argument ===")

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(n int) { // n is a COPY of i at this moment
			defer wg.Done()
			fmt.Printf("  n = %d\n", n)
		}(i) // pass i here — creates a copy
	}
	wg.Wait()

	// ============================================
	// ✅ FIX 2: Shadow variable in loop body
	// Since Go 1.22, loop variables are per-iteration by default!
	// But for older Go versions or clarity, this explicit shadow works:
	// ============================================
	fmt.Println("\n=== ✅ FIX 2: Shadow variable ===")

	for i := 0; i < 5; i++ {
		i := i // create a NEW variable scoped to this iteration
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Printf("  i = %d\n", i) // captures the shadowed copy
		}()
	}
	wg.Wait()

	// ============================================
	// NOTE: Since Go 1.22, loop variables are per-iteration
	// So the original "bug" example actually works correctly in Go 1.22+!
	// But it's still good practice to pass as arguments for clarity.
	// ============================================
	fmt.Println("\n=== Go 1.22+ loop variable semantics ===")
	fmt.Println("In Go 1.22+, loop vars are per-iteration by default.")
	fmt.Println("The 'bug' above may actually work correctly in your Go version!")
	fmt.Println("But passing as argument is still clearer and more portable.")
}
