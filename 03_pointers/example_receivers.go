//go:build ignore

package main

import "fmt"

// ============================================
// Value receivers vs Pointer receivers
//
// Python: self is always a reference → methods always modify the object
// Go:    you choose:
//   (c Counter)  = value receiver (read-only copy)
//   (c *Counter) = pointer receiver (modifies original)
// ============================================

// Counter is a simple struct to demonstrate receivers
type Counter struct {
	Name  string
	Count int
}

// Value receiver — works on a COPY
// Use for: reading, computing, returning derived values
func (c Counter) GetCount() int {
	return c.Count
}

// Value receiver — modifications are LOST
func (c Counter) IncrementBroken() {
	c.Count++ // modifies the copy, not the original!
}

// Pointer receiver — works on the ORIGINAL
// Use for: mutating state, large structs
func (c *Counter) Increment() {
	c.Count++ // modifies the actual struct
}

func (c *Counter) Reset() {
	c.Count = 0
}

func (c Counter) String() string {
	return fmt.Sprintf("Counter{%s: %d}", c.Name, c.Count)
}

func main() {
	// ============================================
	// Example 1: Pointer receiver modifies
	// ============================================
	fmt.Println("=== Example 1: Pointer receiver ===")

	c := Counter{Name: "clicks", Count: 0}
	fmt.Println("  Start:", c)

	c.Increment()
	c.Increment()
	c.Increment()
	fmt.Println("  After 3 Increment():", c)

	c.Reset()
	fmt.Println("  After Reset():", c)

	// ============================================
	// Example 2: Value receiver does NOT modify
	// ============================================
	fmt.Println("\n=== Example 2: Value receiver (broken) ===")

	c2 := Counter{Name: "attempts", Count: 5}
	c2.IncrementBroken()
	c2.IncrementBroken()
	c2.IncrementBroken()
	fmt.Println("  After 3 IncrementBroken():", c2) // still 5!

	// ============================================
	// Example 3: Real-world example — BankAccount
	// ============================================
	fmt.Println("\n=== Example 3: BankAccount ===")

	account := &BankAccount{Owner: "Alice", Balance: 100.0}
	fmt.Println("  Start:", account)

	account.Deposit(50)
	fmt.Println("  After Deposit(50):", account)

	err := account.Withdraw(30)
	if err != nil {
		fmt.Println("  Error:", err)
	}
	fmt.Println("  After Withdraw(30):", account)

	err = account.Withdraw(200)
	if err != nil {
		fmt.Println("  Withdraw(200) Error:", err)
	}

	// ============================================
	// Example 4: Pointer receiver with a pointer variable
	// Go auto-dereferences for you!
	// ============================================
	fmt.Println("\n=== Example 4: Auto-dereference ===")

	// All of these work the same:
	c3 := Counter{Name: "test", Count: 0}
	c3.Increment() // Go takes &c3 automatically

	c4 := &Counter{Name: "test2", Count: 0}
	c4.Increment() // c4 is already a pointer

	fmt.Printf("  c3 (value var): %s\n", c3)
	fmt.Printf("  c4 (pointer var): %s\n", *c4)
}

// BankAccount demonstrates a practical use of pointer receivers
type BankAccount struct {
	Owner   string
	Balance float64
}

// Deposit adds money — must use pointer receiver to modify balance
func (a *BankAccount) Deposit(amount float64) {
	a.Balance += amount
}

// Withdraw removes money — returns an error if insufficient funds
func (a *BankAccount) Withdraw(amount float64) error {
	if amount > a.Balance {
		return fmt.Errorf("insufficient funds: need %.2f, have %.2f", amount, a.Balance)
	}
	a.Balance -= amount
	return nil
}

// GetBalance is a value receiver — it only reads
func (a BankAccount) GetBalance() float64 {
	return a.Balance
}

func (a BankAccount) String() string {
	return fmt.Sprintf("BankAccount{%s: $%.2f}", a.Owner, a.Balance)
}
