//go:build ignore

package main

import "fmt"

func main() {
	// ============================================
	// 1. var declarations (explicit type)
	// ============================================
	var name string = "Shankar"
	var age int = 21
	var gpa float64 = 3.8
	var enrolled bool = true

	fmt.Println("=== Explicit var declarations ===")
	fmt.Println("Name:", name)
	fmt.Println("Age:", age)
	fmt.Println("GPA:", gpa)
	fmt.Println("Enrolled:", enrolled)

	// ============================================
	// 2. Short declarations with := (type inferred)
	// This is what you'll use 90% of the time inside functions.
	// ============================================
	city := "Bangalore" // compiler infers string
	year := 2026        // compiler infers int
	pi := 3.14159       // compiler infers float64

	fmt.Println("\n=== Short declarations (:=) ===")
	fmt.Println("City:", city)
	fmt.Println("Year:", year)
	fmt.Println("Pi:", pi)

	// ============================================
	// 3. Zero values — every type has a default
	// In Python, uninitialized = NameError.
	// In Go, uninitialized = zero value. Always safe.
	// ============================================
	var count int      // 0
	var message string // "" (empty string, NOT None)
	var active bool    // false
	var rate float64   // 0.0
	var ptr *int       // nil (pointers default to nil)

	fmt.Println("\n=== Zero values ===")
	fmt.Printf("int: %d\n", count)
	fmt.Printf("string: %q\n", message) // %q shows quotes around string
	fmt.Printf("bool: %t\n", active)
	fmt.Printf("float64: %f\n", rate)
	fmt.Printf("pointer: %v\n", ptr)

	// ============================================
	// 4. Multiple variables at once
	// ============================================
	var (
		host  = "localhost"
		port  = 8080
		debug = false
	)
	fmt.Println("\n=== Grouped declarations ===")
	fmt.Printf("Server: %s:%d (debug=%t)\n", host, port, debug)

	// ============================================
	// 5. Constants — like Python's convention of UPPER_CASE, but enforced
	// ============================================
	const MaxRetries = 3
	const BaseURL = "https://api.example.com"
	// MaxRetries = 5  // ← COMPILE ERROR: cannot assign to MaxRetries

	fmt.Println("\n=== Constants ===")
	fmt.Println("Max retries:", MaxRetries)
	fmt.Println("Base URL:", BaseURL)

	// ============================================
	// 6. Type conversions (not casting — Go is explicit)
	// Python: int("42"), float(42), str(42)
	// Go: no implicit conversions, ever.
	// ============================================
	var x int = 42
	var y float64 = float64(x) // must explicitly convert
	var z int = int(y)         // and back

	fmt.Println("\n=== Type conversions ===")
	fmt.Printf("int %d → float64 %f → int %d\n", x, y, z)

	// This would NOT compile:
	// var a int = 42
	// var b float64 = a   // ← ERROR: cannot use a (type int) as type float64
}
