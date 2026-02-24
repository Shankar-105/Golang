//go:build ignore

package main

import (
	"fmt"
	"sync"
	"time"
)

// ============================================
// RWMutex: multiple concurrent readers OR one exclusive writer.
//
// Use when reads >> writes. Regular Mutex would serialize all reads.
// RWMutex lets reads happen in parallel.
//
// Python has no built-in RWLock. The closest is threading.RLock()
// (re-entrant lock), which is a different concept.
// ============================================

// Config is a thread-safe configuration store.
// Many goroutines read it; occasionally one writes.
type Config struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewConfig() *Config {
	return &Config{data: make(map[string]string)}
}

// Get uses RLock — multiple goroutines can call simultaneously
func (c *Config) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.data[key]
	return val, ok
}

// Set uses Lock — exclusive access, blocks all readers and writers
func (c *Config) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

func main() {
	cfg := NewConfig()
	cfg.Set("host", "localhost")
	cfg.Set("port", "8080")

	var wg sync.WaitGroup

	// ============================================
	// Spawn 100 readers — they all run concurrently (RLock allows it)
	// ============================================
	fmt.Println("=== RWMutex: 100 readers + 5 writers ===")

	start := time.Now()

	// Readers (fast, concurrent)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				val, _ := cfg.Get("host") // RLock — runs in parallel with other reads
				_ = val
			}
		}(i)
	}

	// Writers (slow, exclusive)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				cfg.Set("host", fmt.Sprintf("server-%d-%d", id, j))
				time.Sleep(1 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)
	fmt.Printf("Completed in %v\n", elapsed)

	val, _ := cfg.Get("host")
	fmt.Printf("Final host value: %s\n", val)

	// ============================================
	// Compare: same thing with plain Mutex (no concurrent reads)
	// ============================================
	fmt.Println("\n=== Comparison: Mutex vs RWMutex ===")
	fmt.Println("With RWMutex: 100 readers run simultaneously")
	fmt.Println("With plain Mutex: 100 readers would be serialized (much slower)")
	fmt.Println("Rule of thumb: use RWMutex when reads outnumber writes 10:1 or more")
}
