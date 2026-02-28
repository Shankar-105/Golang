//go:build ignore

package main

import "fmt"

// ============================================
// Common pointer gotchas for Python developers
// ============================================

func main() {
	// ============================================
	// Gotcha 1: Slices are already reference-like
	// You do NOT need *[]int for most cases
	// ============================================
	fmt.Println("=== Gotcha 1: Slices are reference types ===")

	nums := []int{1, 2, 3}
	modifySlice(nums)
	fmt.Println("  After modifySlice:", nums) // [999 2 3] — modified!
	// Because slices contain a pointer to the underlying array

	// BUT: append might create a new backing array
	nums2 := []int{1, 2, 3}
	appendToSlice(nums2)
	fmt.Println("  After appendToSlice:", nums2) // [1 2 3] — NOT modified!
	// append may allocate a new array, so the original slice header is unchanged

	// Solution: return the new slice
	nums3 := []int{1, 2, 3}
	nums3 = appendAndReturn(nums3)
	fmt.Println("  After appendAndReturn:", nums3) // [1 2 3 4]

	// ============================================
	// Gotcha 2: Maps are already references
	// ============================================
	fmt.Println("\n=== Gotcha 2: Maps are reference types ===")

	scores := map[string]int{"alice": 90}
	addScore(scores, "bob", 85)
	fmt.Println("  After addScore:", scores) // map[alice:90 bob:85] — modified!

	// ============================================
	// Gotcha 3: Struct copies in range loops
	// ============================================
	fmt.Println("\n=== Gotcha 3: Range loop copies ===")

	type Point struct {
		X, Y int
	}

	points := []Point{{1, 2}, {3, 4}, {5, 6}}

	// WRONG: p is a copy, modifications don't stick
	for _, p := range points {
		p.X *= 10 // modifies copy only!
		_ = p
	}
	fmt.Println("  After range (copy):", points) // unchanged

	// RIGHT: use index to modify in place
	for i := range points {
		points[i].X *= 10 // modifies the actual element
	}
	fmt.Println("  After range (index):", points) // X values are 10x

	// ============================================
	// Gotcha 4: Pointer to loop variable
	// ============================================
	fmt.Println("\n=== Gotcha 4: Pointer to loop variable ===")

	names := []string{"Alice", "Bob", "Charlie"}
	ptrs := make([]*string, len(names))

	// Note: In Go 1.22+, loop variables are per-iteration,
	// so this gotcha is fixed. But for older Go or awareness:
	for i, name := range names {
		name := name // re-declare to capture (pre-Go 1.22 fix)
		ptrs[i] = &name
	}

	for _, p := range ptrs {
		fmt.Printf("  %s ", *p)
	}
	fmt.Println()

	// ============================================
	// Gotcha 5: Comparing pointers vs values
	// ============================================
	fmt.Println("\n=== Gotcha 5: Pointer comparison ===")

	a := 42
	b := 42
	pa := &a
	pb := &b

	fmt.Println("  a == b:", a == b)         // true — same value
	fmt.Println("  pa == pb:", pa == pb)     // false — different addresses!
	fmt.Println("  *pa == *pb:", *pa == *pb) // true — same dereferenced value
}

func modifySlice(s []int) {
	s[0] = 999 // modifies underlying array — visible to caller
}

func appendToSlice(s []int) {
	s = append(s, 4) // may create new backing array
	_ = s
}

func appendAndReturn(s []int) []int {
	return append(s, 4) // return the new slice
}

func addScore(m map[string]int, name string, score int) {
	m[name] = score // maps are references — visible to caller
}
