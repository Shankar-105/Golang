//go:build ignore

package main

import (
	"errors"
	"fmt"
)

// ============================================
// Custom error types
//
// Since error is just an interface { Error() string },
// ANY struct can be an error if it has an Error() method.
//
// Python equivalent:
//   class ValidationError(Exception):
//       def __init__(self, field, message):
//           self.field = field
//           self.message = message
// ============================================

// ValidationError represents a field validation failure
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed on %q: %s", e.Field, e.Message)
}

// NotFoundError represents a missing resource
type NotFoundError struct {
	Resource string
	ID       int
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with ID %d not found", e.Resource, e.ID)
}

// ============================================
// Example usage
// ============================================

type User struct {
	Name  string
	Email string
	Age   int
}

func validateUser(u User) error {
	if u.Name == "" {
		return &ValidationError{Field: "name", Message: "cannot be empty"}
	}
	if u.Age < 0 || u.Age > 150 {
		return &ValidationError{Field: "age", Message: fmt.Sprintf("%d is out of range", u.Age)}
	}
	if u.Email == "" {
		return &ValidationError{Field: "email", Message: "cannot be empty"}
	}
	return nil
}

var users = map[int]User{
	1: {Name: "Alice", Email: "alice@go.dev", Age: 25},
	2: {Name: "Bob", Email: "bob@go.dev", Age: 30},
}

func getUser(id int) (User, error) {
	u, ok := users[id]
	if !ok {
		return User{}, &NotFoundError{Resource: "User", ID: id}
	}
	return u, nil
}

func main() {
	// ============================================
	// Example 1: Validation errors
	// ============================================
	fmt.Println("=== Example 1: Validation errors ===")

	testUsers := []User{
		{Name: "Charlie", Email: "charlie@go.dev", Age: 28},
		{Name: "", Email: "anon@go.dev", Age: 20},
		{Name: "Dave", Email: "dave@go.dev", Age: -5},
		{Name: "Eve", Email: "", Age: 22},
	}

	for _, u := range testUsers {
		err := validateUser(u)
		if err != nil {
			// Use errors.As to extract the ValidationError
			var valErr *ValidationError
			if errors.As(err, &valErr) {
				fmt.Printf("  INVALID: field=%q, reason=%q\n", valErr.Field, valErr.Message)
			}
		} else {
			fmt.Printf("  VALID: %s\n", u.Name)
		}
	}

	// ============================================
	// Example 2: Not found errors
	// ============================================
	fmt.Println("\n=== Example 2: Not found errors ===")

	for _, id := range []int{1, 99} {
		user, err := getUser(id)
		if err != nil {
			var nfErr *NotFoundError
			if errors.As(err, &nfErr) {
				fmt.Printf("  %s #%d: NOT FOUND\n", nfErr.Resource, nfErr.ID)
			}
		} else {
			fmt.Printf("  Found: %s (%s)\n", user.Name, user.Email)
		}
	}

	// ============================================
	// Example 3: Switch on error type
	// ============================================
	fmt.Println("\n=== Example 3: Handle different error types ===")

	errs := []error{
		&ValidationError{Field: "email", Message: "invalid format"},
		&NotFoundError{Resource: "Order", ID: 404},
		fmt.Errorf("unknown database error"),
	}

	for _, err := range errs {
		switch {
		case errors.As(err, new(*ValidationError)):
			fmt.Println("  → Validation problem:", err)
		case errors.As(err, new(*NotFoundError)):
			fmt.Println("  → Not found:", err)
		default:
			fmt.Println("  → Unknown error:", err)
		}
	}
}
