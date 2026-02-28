//go:build ignore

package main

import (
	"errors"
	"fmt"
	"os"
)

// ============================================
// Error wrapping and unwrapping
//
// Python: raise NewError("context") from original_error
// Go:    fmt.Errorf("context: %w", originalErr)
//
// Key: %w wraps the error. %v just formats the string (no wrapping).
// ============================================

// Sentinel errors — known errors that callers can check
var (
	ErrNotFound = errors.New("not found")
	ErrEmpty    = errors.New("empty input")
)

func main() {
	// ============================================
	// Example 1: Wrapping with fmt.Errorf %w
	// ============================================
	fmt.Println("=== Example 1: Error wrapping ===")

	err := loadConfig("nonexistent.yaml")
	if err != nil {
		fmt.Println("  Error:", err)
		// Output: "loadConfig: readFile: open nonexistent.yaml: ..."

		// Unwrap to check the root cause
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("  Root cause: file doesn't exist!")
		}
	}

	// ============================================
	// Example 2: errors.Is — check if a specific error is in the chain
	// ============================================
	fmt.Println("\n=== Example 2: errors.Is ===")

	err = findUser("")
	fmt.Println("  findUser(\"\"):", err)
	fmt.Println("  Is ErrEmpty?", errors.Is(err, ErrEmpty))       // true
	fmt.Println("  Is ErrNotFound?", errors.Is(err, ErrNotFound)) // false

	err = findUser("ghost")
	fmt.Println("\n  findUser(\"ghost\"):", err)
	fmt.Println("  Is ErrNotFound?", errors.Is(err, ErrNotFound)) // true

	// ============================================
	// Example 3: errors.As — extract a typed error
	// ============================================
	fmt.Println("\n=== Example 3: errors.As ===")

	err = processRequest(-1)
	fmt.Println("  processRequest(-1):", err)

	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		fmt.Printf("  Extracted HTTPError: status=%d, message=%q\n",
			httpErr.StatusCode, httpErr.Message)
	}

	// ============================================
	// Example 4: Error chain visualization
	// ============================================
	fmt.Println("\n=== Example 4: Error chain ===")

	err = level3()
	fmt.Println("  Full error:", err)

	// Unwrap step by step
	fmt.Println("  Unwrap 1:", errors.Unwrap(err))
	fmt.Println("  Unwrap 2:", errors.Unwrap(errors.Unwrap(err)))
}

func loadConfig(path string) error {
	err := readFile(path)
	if err != nil {
		return fmt.Errorf("loadConfig: %w", err) // wrap with context
	}
	return nil
}

func readFile(path string) error {
	_, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("readFile: %w", err) // wrap with context
	}
	return nil
}

func findUser(name string) error {
	if name == "" {
		return fmt.Errorf("findUser: %w", ErrEmpty)
	}
	// Simulate: user not in database
	return fmt.Errorf("findUser(%q): %w", name, ErrNotFound)
}

// HTTPError is a custom error type with status code
type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

func processRequest(userID int) error {
	if userID < 0 {
		return &HTTPError{StatusCode: 400, Message: "invalid user ID"}
	}
	return nil
}

func level1() error {
	return errors.New("disk full")
}

func level2() error {
	err := level1()
	return fmt.Errorf("writing cache: %w", err)
}

func level3() error {
	err := level2()
	return fmt.Errorf("saving user profile: %w", err)
}
