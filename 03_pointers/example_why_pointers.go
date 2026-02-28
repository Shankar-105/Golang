//go:build ignore

package main

import "fmt"

// ============================================
// Why pointers exist:
// 1. Functions can modify the caller's data
// 2. Avoid copying large structs
//
// Python: all objects are references — modification "just works"
// Go:    you must explicitly pass a pointer to allow modification
// ============================================

func main() {
	// ============================================
	// Example 1: Without pointer — function can't modify
	// ============================================
	fmt.Println("=== Example 1: Pass by value (copy) ===")

	x := 5
	tryToDouble(x)
	fmt.Println("  After tryToDouble:", x) // 5 — unchanged!

	// ============================================
	// Example 2: With pointer — function CAN modify
	// ============================================
	fmt.Println("\n=== Example 2: Pass by pointer ===")

	y := 5
	double(&y)                        // pass the ADDRESS of y
	fmt.Println("  After double:", y) // 10 — modified!

	// ============================================
	// Example 3: Swap two values
	// ============================================
	fmt.Println("\n=== Example 3: Swap ===")

	a, b := 10, 20
	fmt.Printf("  Before swap: a=%d, b=%d\n", a, b)
	swap(&a, &b)
	fmt.Printf("  After swap:  a=%d, b=%d\n", a, b)

	// ============================================
	// Example 4: Modify a struct (the real use case)
	// ============================================
	fmt.Println("\n=== Example 4: Struct modification ===")

	user := User{Name: "Alice", Age: 25}

	// Pass by value — doesn't change original
	birthdayByValue(user)
	fmt.Println("  After birthdayByValue:", user.Age) // 25

	// Pass by pointer — changes original
	birthdayByPointer(&user)
	fmt.Println("  After birthdayByPointer:", user.Age) // 26

	// ============================================
	// Example 5: Return a pointer from a function
	// ============================================
	fmt.Println("\n=== Example 5: Return pointer ===")

	u := newUser("Bob", 30)
	fmt.Printf("  Created user: %+v\n", *u)
	// Go is smart: it moves this to the heap automatically
	// (Python does this for every object)
}

// tryToDouble receives a COPY — can't change the original
func tryToDouble(n int) {
	n = n * 2 // only changes the local copy
}

// double receives a POINTER — changes the original
func double(n *int) {
	*n = *n * 2 // dereference and modify
}

// swap exchanges two values through pointers
func swap(a, b *int) {
	*a, *b = *b, *a
}

type User struct {
	Name string
	Age  int
}

// birthdayByValue — receives a copy, original unchanged
func birthdayByValue(u User) {
	u.Age++ // modifies the copy only
}

// birthdayByPointer — receives a pointer, original is modified
func birthdayByPointer(u *User) {
	u.Age++ // Go auto-dereferences: same as (*u).Age++
}

// newUser creates a User and returns a pointer to it
// In Python, returning a new object is natural — it's always a reference
// In Go, returning &localVar is safe — Go moves it to the heap
func newUser(name string, age int) *User {
	u := User{Name: name, Age: age}
	return &u // safe! Go handles the memory
}
