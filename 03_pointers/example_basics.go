//go:build ignore

package main

import "fmt"

// ============================================
// Pointer basics: & (address-of) and * (dereference)
//
// Python: everything is a reference, you never see addresses
// Go:    you explicitly choose value vs pointer
// ============================================

func main() {
	// ============================================
	// Example 1: Creating and using pointers
	// ============================================
	fmt.Println("=== Example 1: Pointer basics ===")

	x := 42
	p := &x // p is a *int (pointer to int), holds the address of x

	fmt.Println("  x  =", x)  // 42
	fmt.Println("  &x =", &x) // 0xc0000b2008 (memory address)
	fmt.Println("  p  =", p)  // same address as &x
	fmt.Println("  *p =", *p) // 42 (dereference: value at address)

	// ============================================
	// Example 2: Modifying through a pointer
	// ============================================
	fmt.Println("\n=== Example 2: Modify via pointer ===")

	fmt.Println("  Before: x =", x)
	*p = 100                        // change the value at the address p points to
	fmt.Println("  After:  x =", x) // 100 — x changed!

	// ============================================
	// Example 3: Pointers share data
	// ============================================
	fmt.Println("\n=== Example 3: Shared data ===")

	a := "hello"
	b := &a
	c := &a // both b and c point to a

	fmt.Println("  a =", a)
	fmt.Println("  *b =", *b)
	fmt.Println("  *c =", *c)

	*b = "world"
	fmt.Println("  After *b = \"world\":")
	fmt.Println("  a =", a)   // "world"
	fmt.Println("  *c =", *c) // "world" — same variable!

	// ============================================
	// Example 4: Value copy vs pointer
	// ============================================
	fmt.Println("\n=== Example 4: Copy vs pointer ===")

	original := 42
	copied := original   // VALUE copy — independent
	pointed := &original // POINTER — shares original

	copied = 999
	fmt.Println("  original =", original) // 42 — unaffected by copy
	fmt.Println("  copied   =", copied)   // 999

	*pointed = 777
	fmt.Println("  original =", original) // 777 — changed via pointer!
	fmt.Println("  pointed  =", *pointed) // 777

	// ============================================
	// Example 5: nil pointer
	// ============================================
	fmt.Println("\n=== Example 5: nil pointer ===")

	var np *int                            // declared but not assigned → nil
	fmt.Println("  np == nil?", np == nil) // true

	// Dereferencing nil would panic:
	// fmt.Println(*np) // runtime error: invalid memory address

	// Safe usage — always check before dereferencing:
	safeDeref(np) // nil — prints warning
	val := 99
	safeDeref(&val) // non-nil — prints value

	// ============================================
	// Example 6: new() function
	// ============================================
	fmt.Println("\n=== Example 6: new() ===")

	ip := new(int)    // allocates an int, returns *int
	sp := new(string) // allocates a string, returns *string

	fmt.Printf("  new(int)    → *ip = %d (zero value)\n", *ip)
	fmt.Printf("  new(string) → *sp = %q (zero value)\n", *sp)

	*ip = 42
	*sp = "hello"
	fmt.Printf("  After assignment: *ip = %d, *sp = %q\n", *ip, *sp)
}

func safeDeref(p *int) {
	if p != nil {
		fmt.Println("  *p =", *p)
	} else {
		fmt.Println("  p is nil, cannot dereference")
	}
}
