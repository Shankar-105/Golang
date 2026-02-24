//go:build ignore

package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	// ============================================
	// Example 1: Data race WITHOUT mutex
	// Run with: go run -race example_mutex.go
	// The race detector will flag this!
	// ============================================
	fmt.Println("=== Example 1: Race condition (unsafe) ===")

	unsafeCounter := 0
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			unsafeCounter++ // DATA RACE!
		}()
	}
	wg.Wait()
	fmt.Printf("  Unsafe counter: %d (expected 1000, got less due to race)\n", unsafeCounter)

	// ============================================
	// Example 2: Fixed with Mutex
	// ============================================
	fmt.Println("\n=== Example 2: Fixed with Mutex ===")

	var mu sync.Mutex
	safeCounter := 0
	var wg2 sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			mu.Lock()
			safeCounter++ // protected by mutex
			mu.Unlock()
		}()
	}
	wg2.Wait()
	fmt.Printf("  Safe counter: %d (always 1000)\n", safeCounter)

	// ============================================
	// Example 3: Mutex-protected struct
	// The idiomatic Go pattern: embed mutex in the struct
	// ============================================
	fmt.Println("\n=== Example 3: Thread-safe struct ===")

	bank := NewBankAccount(1000)

	var wg3 sync.WaitGroup

	// 100 goroutines deposit $10 each
	for i := 0; i < 100; i++ {
		wg3.Add(1)
		go func() {
			defer wg3.Done()
			bank.Deposit(10)
		}()
	}

	// 50 goroutines withdraw $5 each
	for i := 0; i < 50; i++ {
		wg3.Add(1)
		go func() {
			defer wg3.Done()
			bank.Withdraw(5)
		}()
	}

	wg3.Wait()
	fmt.Printf("  Balance: $%d (expected: 1000 + 100*10 - 50*5 = $1750)\n", bank.Balance())

	// ============================================
	// Example 4: Deadlock from double-lock (common mistake)
	// ============================================
	fmt.Println("\n=== Example 4: Deadlock warning ===")
	fmt.Println("  A mutex is NOT re-entrant in Go!")
	fmt.Println("  mu.Lock(); mu.Lock() → DEADLOCK (goroutine blocks waiting for itself)")
	fmt.Println("  Python's threading.RLock() IS re-entrant. Go's Mutex is NOT.")
	fmt.Println("  Design your code so a locked function never calls another locked function.")
}

// BankAccount is a thread-safe bank account
type BankAccount struct {
	mu      sync.Mutex
	balance int
}

func NewBankAccount(initial int) *BankAccount {
	return &BankAccount{balance: initial}
}

func (a *BankAccount) Deposit(amount int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.balance += amount
}

func (a *BankAccount) Withdraw(amount int) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.balance < amount {
		return false
	}
	a.balance -= amount
	return true
}

func (a *BankAccount) Balance() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.balance
}

// Timing a mutex vs no-mutex (informal benchmark)
func init() {
	_ = time.Now // suppress unused import warning from time
}
