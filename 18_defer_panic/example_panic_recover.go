//go:build ignore

package main

import (
	"errors"
	"fmt"
	"runtime/debug"
)

// ──────────────────────────────────────────────────────────────
// panic and recover — Go's mechanism for truly unrecoverable errors
//
// Python comparison:
//   raise Exception("bad")     →  panic("bad")
//   except Exception as e:     →  if r := recover(); r != nil { ... }
//   traceback.print_exc()      →  debug.Stack()
//
// KEY DIFFERENCE: In Python, exceptions are normal control flow.
// In Go, panic is for BUGS and IMPOSSIBLE STATES only.
// Normal errors use the (value, error) return pattern.
// ──────────────────────────────────────────────────────────────

func main() {
	fmt.Println("═══ 1. Basic Panic ═══")
	basicPanicDemo()

	fmt.Println("\n═══ 2. Basic Recover ═══")
	basicRecoverDemo()

	fmt.Println("\n═══ 3. Recover Returns the Panic Value ═══")
	recoverValueDemo()

	fmt.Println("\n═══ 4. Convert Panic to Error ═══")
	convertPanicToError()

	fmt.Println("\n═══ 5. Panic in Goroutines ═══")
	goroutinePanicDemo()

	fmt.Println("\n═══ 6. Stack Trace on Recover ═══")
	stackTraceDemo()

	fmt.Println("\n═══ 7. Re-panicking ═══")
	rePanicDemo()

	fmt.Println("\n═══ 8. Panic with Different Value Types ═══")
	panicTypesDemo()
}

// ──── 1. Basic Panic ────────────────────────────────────────
func basicPanicDemo() {
	// Wrap in a function so the panic doesn't crash our program
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("  Caught panic: %v\n", r)
			}
		}()

		fmt.Println("  Before panic")
		panic("something went wrong")
		// The line below NEVER executes — panic stops normal flow
	}()

	fmt.Println("  Program continues after recovery")
}

// ──── 2. Basic Recover ──────────────────────────────────────
func basicRecoverDemo() {
	fmt.Println("  Calling riskyFunction()...")
	riskyFunction()
	fmt.Println("  Back from riskyFunction, program is fine!")
}

func riskyFunction() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("  Recovered in riskyFunction: %v\n", r)
		}
	}()

	fmt.Println("  About to do something risky...")
	panic("risky operation failed")
}

// ──── 3. Recover Returns the Panic Value ────────────────────
func recoverValueDemo() {
	defer func() {
		r := recover()
		if r == nil {
			fmt.Println("  No panic occurred")
			return
		}

		// r is of type `any` — you need type assertions
		switch v := r.(type) {
		case string:
			fmt.Printf("  Panic string: %q\n", v)
		case error:
			fmt.Printf("  Panic error: %v\n", v)
		case int:
			fmt.Printf("  Panic int: %d\n", v)
		default:
			fmt.Printf("  Panic unknown type: %v\n", v)
		}
	}()

	// Try different panic values:
	panic(errors.New("custom error"))
	// panic("string panic")
	// panic(42)
}

// ──── 4. Convert Panic to Error ─────────────────────────────
// This is the MOST IMPORTANT real-world pattern.
// Use this at API boundaries to turn panics into errors.

func safeDivide(a, b float64) (result float64, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in safeDivide: %v", r)
		}
	}()

	if b == 0 {
		panic("division by zero") // simulating a panic-prone library
	}
	return a / b, nil
}

func convertPanicToError() {
	// Normal case
	result, err := safeDivide(10, 3)
	fmt.Printf("  10/3 = %.2f, err = %v\n", result, err)

	// Panic case → converted to error
	result, err = safeDivide(10, 0)
	fmt.Printf("  10/0 = %.2f, err = %v\n", result, err)

	// The program continues! No crash.
	fmt.Println("  Program still running after division by zero")
}

// ──── 5. Panic in Goroutines ────────────────────────────────
// CRITICAL: A panic in a goroutine kills the ENTIRE PROGRAM
// unless recovered WITHIN that goroutine!

func goroutinePanicDemo() {
	done := make(chan string)

	// ✅ GOOD: recover inside the goroutine
	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Sprintf("goroutine recovered: %v", r)
			}
		}()
		panic("goroutine panic!")
	}()

	msg := <-done
	fmt.Printf("  %s\n", msg)
	fmt.Println("  Main goroutine is fine!")

	// ❌ BAD: recovering in the parent does NOT help
	// go func() {
	//     panic("crash!")  // This kills the whole program!
	// }()
	// recover()  // This does NOTHING for panics in other goroutines
}

// ──── 6. Stack Trace on Recover ─────────────────────────────
func stackTraceDemo() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("  Panic: %v\n", r)
			fmt.Printf("  Stack trace:\n%s\n", debug.Stack())
		}
	}()

	level1()
}

func level1() { level2() }
func level2() { level3() }
func level3() { panic("deep panic") }

// ──── 7. Re-panicking ──────────────────────────────────────
// Sometimes you want to recover, check the panic, and re-panic
// if you can't handle it.

func rePanicDemo() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("  Outer recover caught: %v\n", r)
		}
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				// Only handle string panics
				if _, ok := r.(string); ok {
					fmt.Printf("  Inner recovered string panic: %v\n", r)
					return // handled
				}
				// Re-panic for anything else
				fmt.Printf("  Inner can't handle this, re-panicking: %v\n", r)
				panic(r) // re-panic!
			}
		}()

		panic(42) // int panic — inner handler can't handle it
	}()
}

// ──── 8. Panic with Different Value Types ───────────────────
func panicTypesDemo() {
	tryPanic := func(name string, fn func()) {
		defer func() {
			r := recover()
			fmt.Printf("  %-20s → recovered: %v (type: %T)\n", name, r, r)
		}()
		fn()
	}

	tryPanic("string", func() { panic("oops") })
	tryPanic("error", func() { panic(errors.New("bad")) })
	tryPanic("int", func() { panic(42) })
	tryPanic("struct", func() {
		panic(struct{ Code int; Msg string }{500, "internal error"})
	})
	tryPanic("nil (no panic)", func() { /* no panic */ })
}
