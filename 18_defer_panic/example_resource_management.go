//go:build ignore

package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

// ──────────────────────────────────────────────────────────────
// Real-world resource management with defer
//
// Python equivalent: context managers (with statement)
//
//   with open("f") as f:       →  f, _ := os.Open("f"); defer f.Close()
//   with lock:                 →  mu.Lock(); defer mu.Unlock()
//   with conn.cursor() as c:  →  rows, _ := db.Query(...); defer rows.Close()
//
// Go's defer is MORE FLEXIBLE than Python's with:
// - Works with any function call, not just __enter__/__exit__
// - Multiple defers in one function (LIFO order)
// - Can modify return values (with named returns)
// ──────────────────────────────────────────────────────────────

func main() {
	fmt.Println("═══ 1. File Resource Management ═══")
	fileResourceDemo()

	fmt.Println("\n═══ 2. Mutex Lock/Unlock ═══")
	mutexDemo()

	fmt.Println("\n═══ 3. Multi-Resource Acquisition ═══")
	multiResourceDemo()

	fmt.Println("\n═══ 4. defer with Error Handling ═══")
	errorHandlingDemo()

	fmt.Println("\n═══ 5. Cleanup in Loop (Correct Pattern) ═══")
	cleanupInLoopDemo()

	fmt.Println("\n═══ 6. Custom Resource Manager ═══")
	customResourceManagerDemo()
}

// ──── 1. File Resource Management ───────────────────────────
func fileResourceDemo() {
	// Create a temp file, write to it, read it back
	tmpFile := "defer_example_temp.txt"

	// Write
	err := writeFile(tmpFile, "Hello from defer example!\nLine 2\nLine 3\n")
	if err != nil {
		fmt.Printf("  Write error: %v\n", err)
		return
	}
	fmt.Println("  Written to", tmpFile)

	// Read
	content, err := readFile(tmpFile)
	if err != nil {
		fmt.Printf("  Read error: %v\n", err)
		return
	}
	fmt.Printf("  Read back: %q\n", content)

	// Clean up
	os.Remove(tmpFile)
}

func writeFile(path, data string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}
	defer f.Close() // ← always closes, even if Write fails

	_, err = f.WriteString(data)
	return err
}

func readFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open: %w", err)
	}
	defer f.Close() // ← always closes, even if ReadAll fails

	data, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("read: %w", err)
	}
	return string(data), nil
}

// ──── 2. Mutex Lock/Unlock ──────────────────────────────────
type SafeCounter struct {
	mu    sync.Mutex
	count int
}

func (c *SafeCounter) Increment() {
	c.mu.Lock()
	defer c.mu.Unlock() // guaranteed unlock, even if code panics
	c.count++
}

func (c *SafeCounter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}

func mutexDemo() {
	counter := &SafeCounter{}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Increment()
		}()
	}
	wg.Wait()

	fmt.Printf("  Counter after 100 increments: %d\n", counter.Value())
}

// ──── 3. Multi-Resource Acquisition ─────────────────────────
// When you need multiple resources, each gets its own defer.
// They release in reverse order (LIFO).

type FakeDB struct{ name string }
type FakeCache struct{ name string }
type FakeConn struct{ name string }

func (d *FakeDB) Close() { fmt.Printf("    Closed %s\n", d.name) }
func (c *FakeCache) Close() { fmt.Printf("    Closed %s\n", c.name) }
func (c *FakeConn) Close() { fmt.Printf("    Closed %s\n", c.name) }

func multiResourceDemo() {
	fmt.Println("  Acquiring resources...")

	db := &FakeDB{name: "PostgreSQL"}
	defer db.Close()
	fmt.Println("    Opened PostgreSQL")

	cache := &FakeCache{name: "Redis"}
	defer cache.Close()
	fmt.Println("    Opened Redis")

	conn := &FakeConn{name: "gRPC connection"}
	defer conn.Close()
	fmt.Println("    Opened gRPC connection")

	fmt.Println("  Doing work with all resources...")
	// On return: gRPC closes first, then Redis, then PostgreSQL
}

// ──── 4. defer with Error Handling ──────────────────────────
// Advanced: modify the return error in a deferred function

func processData(data string) (err error) {
	r := strings.NewReader(data)

	// Use a buffer as our "output file"
	var buf bytes.Buffer

	// Simulate writing + potential error
	defer func() {
		// Flush the buffer (like flushing a file)
		flushErr := flush(&buf)
		if err == nil {
			err = flushErr // only overwrite if no prior error
		}
	}()

	_, err = io.Copy(&buf, r)
	return err
}

func flush(buf *bytes.Buffer) error {
	fmt.Printf("  Flushing buffer (%d bytes)\n", buf.Len())
	return nil // simulate successful flush
}

func errorHandlingDemo() {
	err := processData("some important data")
	fmt.Printf("  Error: %v\n", err)
}

// ──── 5. Cleanup in Loops ───────────────────────────────────
// WRONG: defer in a loop stacks all defers until function returns
// RIGHT: wrap each iteration in a function

func cleanupInLoopDemo() {
	files := []string{"file1.txt", "file2.txt", "file3.txt"}

	fmt.Println("  Processing files with proper cleanup:")
	for _, name := range files {
		err := processOneFile(name)
		if err != nil {
			fmt.Printf("    Error processing %s: %v\n", name, err)
		}
	}
}

func processOneFile(name string) error {
	fmt.Printf("    Opening %s\n", name)
	// In real code: f, err := os.Open(name)
	defer fmt.Printf("    Closing %s\n", name)

	fmt.Printf("    Processing %s\n", name)
	return nil
}

// ──── 6. Custom Resource Manager ────────────────────────────
// Building a Python-style context manager in Go

type ResourceManager struct {
	resources []io.Closer
}

func NewResourceManager() *ResourceManager {
	return &ResourceManager{}
}

func (rm *ResourceManager) Add(c io.Closer) {
	rm.resources = append(rm.resources, c)
}

func (rm *ResourceManager) CloseAll() error {
	var firstErr error
	// Close in reverse order (LIFO, like defer)
	for i := len(rm.resources) - 1; i >= 0; i-- {
		if err := rm.resources[i].Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// closerFunc adapts a plain function to io.Closer
type closerFunc func() error

func (f closerFunc) Close() error { return f() }

func customResourceManagerDemo() {
	rm := NewResourceManager()
	defer rm.CloseAll()

	// Add resources
	rm.Add(closerFunc(func() error {
		fmt.Println("    Releasing resource A")
		return nil
	}))
	rm.Add(closerFunc(func() error {
		fmt.Println("    Releasing resource B")
		return nil
	}))
	rm.Add(closerFunc(func() error {
		fmt.Println("    Releasing resource C")
		return nil
	}))

	fmt.Println("  Working with managed resources...")
	fmt.Println("  Done, cleaning up:")
	// CloseAll runs on defer: C, B, A (reverse order)
}
