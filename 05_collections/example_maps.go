//go:build ignore

package main

import "fmt"

// ============================================
// Maps: Go's hash maps (Python's dict)
//
// Python: dict — ordered (3.7+), any hashable key
// Go:    map[K]V — unordered, key must be comparable
// ============================================

func main() {
	// ============================================
	// Example 1: Creating maps
	// ============================================
	fmt.Println("=== Example 1: Creating maps ===")

	// Literal
	ages := map[string]int{
		"Alice":   25,
		"Bob":     30,
		"Charlie": 35,
	}
	fmt.Println("  Literal:", ages)

	// make
	scores := make(map[string]int)
	scores["math"] = 95
	scores["english"] = 88
	fmt.Println("  make:", scores)

	// nil map (read ok, write panics!)
	var nilMap map[string]int
	fmt.Println("  nil map read:", nilMap["key"]) // 0 (zero value, no panic)
	// nilMap["key"] = 1  // PANIC!

	// ============================================
	// Example 2: CRUD operations
	// ============================================
	fmt.Println("\n=== Example 2: CRUD ===")

	m := map[string]string{
		"name":  "Alice",
		"email": "alice@go.dev",
	}

	// Create / Update
	m["city"] = "Seattle"  // add new key
	m["name"] = "Alice B." // update existing
	fmt.Println("  After add/update:", m)

	// Read
	fmt.Println("  name:", m["name"])

	// Read with existence check (comma-ok pattern)
	email, ok := m["email"]
	fmt.Printf("  email=%q, exists=%t\n", email, ok)

	phone, ok := m["phone"]
	fmt.Printf("  phone=%q, exists=%t\n", phone, ok) // "", false

	// Delete
	delete(m, "email")
	fmt.Println("  After delete email:", m)

	// ============================================
	// Example 3: Iterating maps
	// ============================================
	fmt.Println("\n=== Example 3: Iteration (random order!) ===")

	inventory := map[string]int{
		"apples":  5,
		"bananas": 12,
		"oranges": 8,
		"grapes":  3,
	}

	// key + value
	for item, count := range inventory {
		fmt.Printf("  %s: %d\n", item, count)
	}

	// keys only
	fmt.Print("  Keys: ")
	for item := range inventory {
		fmt.Print(item, " ")
	}
	fmt.Println()

	// ============================================
	// Example 4: Map as a set
	// ============================================
	fmt.Println("\n=== Example 4: Map as set ===")

	// Using map[T]bool
	seen := map[string]bool{}
	words := []string{"hello", "world", "hello", "go", "world", "go", "go"}

	for _, w := range words {
		seen[w] = true
	}

	fmt.Println("  Unique words:")
	for word := range seen {
		fmt.Printf("    %s\n", word)
	}

	// Using map[T]struct{} — zero memory per value
	visited := map[int]struct{}{}
	for _, n := range []int{1, 2, 3, 2, 1, 4, 3} {
		visited[n] = struct{}{}
	}
	fmt.Printf("  Unique numbers: ")
	for n := range visited {
		fmt.Printf("%d ", n)
	}
	fmt.Println()

	// ============================================
	// Example 5: Counting / grouping
	// ============================================
	fmt.Println("\n=== Example 5: Word frequency ===")

	text := []string{"the", "cat", "sat", "on", "the", "mat", "the", "cat"}
	freq := map[string]int{}

	for _, word := range text {
		freq[word]++ // zero value of int is 0, so this works!
	}

	for word, count := range freq {
		fmt.Printf("  %q: %d\n", word, count)
	}

	// ============================================
	// Example 6: Nested maps
	// ============================================
	fmt.Println("\n=== Example 6: Nested maps ===")

	// map of maps — like Python's dict of dicts
	users := map[string]map[string]string{
		"alice": {
			"email": "alice@go.dev",
			"role":  "admin",
		},
		"bob": {
			"email": "bob@go.dev",
			"role":  "user",
		},
	}

	for name, info := range users {
		fmt.Printf("  %s: email=%s, role=%s\n", name, info["email"], info["role"])
	}

	// Adding a new user requires initializing the inner map
	users["charlie"] = map[string]string{
		"email": "charlie@go.dev",
		"role":  "user",
	}
	fmt.Println("  Added charlie:", users["charlie"])
}
