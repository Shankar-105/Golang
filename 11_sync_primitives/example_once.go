//go:build ignore

package main

import (
	"fmt"
	"sync"
)

// ============================================
// sync.Once — execute a function exactly ONCE
// no matter how many goroutines call it.
//
// Thread-safe singleton pattern in Go.
//
// Python doesn't have a stdlib equivalent.
// The closest is module-level initialization or functools.lru_cache.
// ============================================

type Database struct {
	Name string
}

var (
	dbOnce sync.Once
	dbInst *Database
)

func GetDB() *Database {
	dbOnce.Do(func() {
		// This function runs exactly ONCE, even if 100 goroutines
		// call GetDB() simultaneously.
		fmt.Println("  [init] Connecting to database... (this only prints once)")
		dbInst = &Database{Name: "production-db"}
	})
	return dbInst
}

func main() {
	fmt.Println("=== sync.Once: Singleton initialization ===")

	var wg sync.WaitGroup

	// 10 goroutines all try to get the database connection
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			db := GetDB()
			fmt.Printf("  Goroutine %d: got db=%q\n", id, db.Name)
		}(i)
	}

	wg.Wait()
	fmt.Println("  Notice: '[init] Connecting...' printed only ONCE above")

	// ============================================
	// Example 2: Once for expensive computation
	// ============================================
	fmt.Println("\n=== Example 2: Lazy initialization ===")

	var configOnce sync.Once
	var config map[string]string

	loadConfig := func() {
		configOnce.Do(func() {
			fmt.Println("  Loading config from disk (expensive)...")
			config = map[string]string{
				"host": "localhost",
				"port": "8080",
			}
		})
	}

	// Call it multiple times — config loads only once
	loadConfig()
	loadConfig()
	loadConfig()
	fmt.Printf("  Config: %v\n", config)
}
