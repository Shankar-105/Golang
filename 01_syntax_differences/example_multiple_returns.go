//go:build ignore

package main

import (
	"errors"
	"fmt"
	"math"
)

// ============================================
// In Python, you'd raise an exception:
//   def divide(a, b):
//       if b == 0:
//           raise ValueError("division by zero")
//       return a / b
//
// In Go, errors are VALUES — you return them alongside the result.
// This is the #1 most important Go idiom.
// ============================================

// divide returns two values: the result AND an error.
// Convention: error is ALWAYS the last return value.
func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("cannot divide %.2f by zero", a)
	}
	return a / b, nil // nil means "no error" — like Python's None
}

// safeSqrt demonstrates named return values (a Go-specific feature).
// The return variables (result, err) are pre-declared and zero-valued.
func safeSqrt(x float64) (result float64, err error) {
	if x < 0 {
		err = fmt.Errorf("cannot take square root of negative number: %.2f", x)
		return // "naked return" — returns the named variables as-is
	}
	result = math.Sqrt(x)
	return // returns result=sqrt(x), err=nil
}

// parseAge shows how you chain error checks — the Go way.
// Compare to Python where you'd nest try/except blocks.
func parseAge(input string) (int, error) {
	// Simulate parsing — in real code you'd use strconv.Atoi
	if input == "" {
		return 0, errors.New("age cannot be empty")
	}
	if input == "abc" {
		return 0, fmt.Errorf("invalid age: %q is not a number", input)
	}
	return 21, nil // simplified for demo
}

func main() {
	fmt.Println("=== Basic multiple returns ===")

	// Pattern: call, check error, use result
	// You'll write this pattern THOUSANDS of times in Go.
	result, err := divide(10, 3)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("10 / 3 = %.4f\n", result)

	// Error case
	result, err = divide(10, 0) // note: = not :=, because result and err already exist
	if err != nil {
		fmt.Println("Error:", err) // this will print
	}

	fmt.Println("\n=== Named return values ===")

	sqrtResult, sqrtErr := safeSqrt(16)
	if sqrtErr != nil {
		fmt.Println("Error:", sqrtErr)
	} else {
		fmt.Printf("sqrt(16) = %.2f\n", sqrtResult)
	}

	sqrtResult, sqrtErr = safeSqrt(-4)
	if sqrtErr != nil {
		fmt.Println("Error:", sqrtErr) // this will print
	}
	_ = sqrtResult // suppress unused warning for the negative case

	fmt.Println("\n=== Discarding values with _ ===")

	// If you ONLY care about the error (rare, but sometimes):
	_, err = divide(10, 0)
	if err != nil {
		fmt.Println("Got expected error:", err)
	}

	// If you ONLY care about the result (dangerous — you're ignoring errors!):
	val, _ := divide(10, 2) // _ discards the error
	fmt.Println("Result (error ignored):", val)

	fmt.Println("\n=== Chaining error checks ===")

	// In Python you might write:
	//   try:
	//       age = parse_age(input)
	//       result = divide(100, age)
	//   except ValueError as e:
	//       print(e)
	//
	// In Go, each step is explicit:
	age, err := parseAge("21")
	if err != nil {
		fmt.Println("Parse error:", err)
		return
	}

	share, err := divide(100, float64(age)) // explicit type conversion!
	if err != nil {
		fmt.Println("Divide error:", err)
		return
	}

	fmt.Printf("Each person's share: %.2f (age: %d)\n", share, age)
}
