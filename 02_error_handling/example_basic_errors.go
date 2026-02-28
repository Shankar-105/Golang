//go:build ignore

package main

import (
	"errors"
	"fmt"
	"math"
	"strconv"
)

// ============================================
// Basic error handling in Go
//
// Python:  try/except with exceptions
// Go:     return values with `if err != nil`
// ============================================

func main() {
	// ============================================
	// Example 1: Simple error return
	// ============================================
	fmt.Println("=== Example 1: Division with error ===")

	result, err := divide(10, 3)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("  10 / 3 = %.2f\n", result)
	}

	result, err = divide(10, 0)
	if err != nil {
		fmt.Println("  10 / 0 =", err) // "cannot divide by zero"
	} else {
		fmt.Printf("  10 / 0 = %.2f\n", result)
	}

	// ============================================
	// Example 2: Multiple error checks in sequence
	// (This is the real Go style — check each step)
	// ============================================
	fmt.Println("\n=== Example 2: Sequential error checks ===")

	input := "42"
	age, err := parseAge(input)
	if err != nil {
		fmt.Println("  Error:", err)
	} else {
		fmt.Printf("  Parsed age: %d\n", age)
	}

	input = "not-a-number"
	age, err = parseAge(input)
	if err != nil {
		fmt.Println("  Error:", err)
	} else {
		fmt.Printf("  Parsed age: %d\n", age)
	}

	input = "-5"
	age, err = parseAge(input)
	if err != nil {
		fmt.Println("  Error:", err)
	} else {
		fmt.Printf("  Parsed age: %d\n", age)
	}

	// ============================================
	// Example 3: errors.New vs fmt.Errorf
	// ============================================
	fmt.Println("\n=== Example 3: Creating errors ===")

	err1 := errors.New("simple error")
	err2 := fmt.Errorf("detailed error: user %q not found", "alice")
	fmt.Println("  errors.New:", err1)
	fmt.Println("  fmt.Errorf:", err2)

	// ============================================
	// Example 4: nil means success
	// ============================================
	fmt.Println("\n=== Example 4: nil = no error ===")

	if err := validatePositive(42); err != nil {
		fmt.Println("  Invalid:", err)
	} else {
		fmt.Println("  42 is valid!")
	}

	if err := validatePositive(-1); err != nil {
		fmt.Println("  Invalid:", err)
	} else {
		fmt.Println("  -1 is valid!")
	}
}

// divide returns an error instead of panicking on division by zero
// Python would raise ZeroDivisionError; Go returns an error value
func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("cannot divide by zero")
	}
	return a / b, nil // nil = no error
}

// parseAge converts a string to an age, validating the result
func parseAge(s string) (int, error) {
	// Step 1: parse the string to int
	n, err := strconv.Atoi(s) // equivalent to Python's int("42")
	if err != nil {
		return 0, fmt.Errorf("parseAge: %q is not a valid number: %w", s, err)
	}

	// Step 2: validate range
	if n < 0 || n > 150 {
		return 0, fmt.Errorf("parseAge: %d is out of range (0-150)", n)
	}

	return n, nil
}

func validatePositive(n int) error {
	if n < 0 {
		return fmt.Errorf("%d is negative", n)
	}
	return nil
}

// Suppress unused import warning
var _ = math.Abs
