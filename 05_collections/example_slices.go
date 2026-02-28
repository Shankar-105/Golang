//go:build ignore

package main

import "fmt"

// ============================================
// Slices: Go's dynamic lists
//
// Python: list — always dynamic, holds any type
// Go:    []T — dynamic, but typed ([]int, []string, etc.)
// ============================================

func main() {
	// ============================================
	// Example 1: Creating slices
	// ============================================
	fmt.Println("=== Example 1: Creating slices ===")

	// Literal
	nums := []int{10, 20, 30, 40, 50}
	fmt.Println("  Literal:", nums)

	// make(type, length, capacity)
	sized := make([]int, 5)             // [0 0 0 0 0] — length 5, cap 5
	preallocated := make([]int, 0, 100) // [] — length 0, cap 100
	fmt.Println("  make(5):", sized, "len:", len(sized), "cap:", cap(sized))
	fmt.Println("  make(0,100):", preallocated, "len:", len(preallocated), "cap:", cap(preallocated))

	// nil slice
	var nilSlice []int
	fmt.Println("  nil slice:", nilSlice, "== nil?", nilSlice == nil)

	// Empty slice (not nil)
	emptySlice := []int{}
	fmt.Println("  empty slice:", emptySlice, "== nil?", emptySlice == nil)

	// ============================================
	// Example 2: Append
	// ============================================
	fmt.Println("\n=== Example 2: Append ===")

	s := []int{1, 2, 3}
	fmt.Println("  Before:", s, "len:", len(s), "cap:", cap(s))

	s = append(s, 4)       // single element
	s = append(s, 5, 6, 7) // multiple elements
	fmt.Println("  After append:", s, "len:", len(s), "cap:", cap(s))

	// Append another slice using ...
	more := []int{8, 9, 10}
	s = append(s, more...) // ... unpacks the slice
	fmt.Println("  After append slice:", s)

	// Append to nil slice — works fine!
	var ns []int
	ns = append(ns, 1, 2, 3)
	fmt.Println("  Append to nil:", ns)

	// ============================================
	// Example 3: Slicing (sub-slices)
	// ============================================
	fmt.Println("\n=== Example 3: Sub-slices ===")

	data := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	fmt.Println("  data[2:5]:", data[2:5]) // [2 3 4]
	fmt.Println("  data[:3]:", data[:3])   // [0 1 2]
	fmt.Println("  data[7:]:", data[7:])   // [7 8 9]
	fmt.Println("  data[:]:", data[:])     // full slice

	// DANGER: sub-slices share the underlying array!
	sub := data[2:5] // [2 3 4]
	sub[0] = 999
	fmt.Println("  After sub[0] = 999:")
	fmt.Println("    sub:", sub)   // [999 3 4]
	fmt.Println("    data:", data) // [0 1 999 3 4 5 6 7 8 9] — changed!

	// To make an independent copy:
	original := []int{1, 2, 3, 4, 5}
	copied := make([]int, len(original))
	copy(copied, original) // built-in copy function
	copied[0] = 999
	fmt.Println("\n  Original:", original) // [1 2 3 4 5] — safe!
	fmt.Println("  Copied:", copied)       // [999 2 3 4 5]

	// ============================================
	// Example 4: Iterating with range
	// ============================================
	fmt.Println("\n=== Example 4: Range iteration ===")

	fruits := []string{"apple", "banana", "cherry"}

	// Index and value
	fmt.Println("  Index + Value:")
	for i, fruit := range fruits {
		fmt.Printf("    [%d] %s\n", i, fruit)
	}

	// Value only (ignore index with _)
	fmt.Println("  Value only:")
	for _, fruit := range fruits {
		fmt.Printf("    %s\n", fruit)
	}

	// Index only
	fmt.Println("  Index only:")
	for i := range fruits {
		fmt.Printf("    %d\n", i)
	}

	// ============================================
	// Example 5: Common operations
	// ============================================
	fmt.Println("\n=== Example 5: Common operations ===")

	// Check if slice contains an element (no built-in — loop it)
	fmt.Println("  Contains 'banana'?", contains(fruits, "banana"))
	fmt.Println("  Contains 'mango'?", contains(fruits, "mango"))

	// Remove element at index (order-preserving)
	colors := []string{"red", "green", "blue", "yellow"}
	colors = removeAt(colors, 1) // remove "green"
	fmt.Println("  After remove index 1:", colors)

	// Filter
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	evens := filter(numbers, func(n int) bool { return n%2 == 0 })
	fmt.Println("  Evens:", evens)

	// ============================================
	// Example 6: Capacity growth
	// ============================================
	fmt.Println("\n=== Example 6: Capacity growth ===")

	var growing []int
	prevCap := cap(growing)
	for i := 0; i < 20; i++ {
		growing = append(growing, i)
		if cap(growing) != prevCap {
			fmt.Printf("  len=%2d, cap changed: %d → %d\n", len(growing), prevCap, cap(growing))
			prevCap = cap(growing)
		}
	}
}

func contains(s []string, target string) bool {
	for _, v := range s {
		if v == target {
			return true
		}
	}
	return false
}

func removeAt(s []string, i int) []string {
	return append(s[:i], s[i+1:]...)
}

func filter(s []int, predicate func(int) bool) []int {
	var result []int
	for _, v := range s {
		if predicate(v) {
			result = append(result, v)
		}
	}
	return result
}
