//go:build ignore

package main

import (
	"fmt"
	"strings"
	"time"
)

// ============================================
// Pipeline Pattern
//
// Each stage is a goroutine that reads from one channel
// and writes to the next. Data flows through the pipeline.
//
// Python equivalent: generator chaining
//   gen = generate(data)
//   squared = (x*x for x in gen)
//   filtered = (x for x in squared if x > 10)
//
// But Go's version is CONCURRENT — each stage runs simultaneously.
// ============================================

// Stage 1: Generate strings
func generate(words ...string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for _, w := range words {
			out <- w
		}
	}()
	return out
}

// Stage 2: Transform — uppercase
func toUpper(in <-chan string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for s := range in {
			time.Sleep(100 * time.Millisecond) // simulate work
			out <- strings.ToUpper(s)
		}
	}()
	return out
}

// Stage 3: Transform — add prefix
func addPrefix(prefix string, in <-chan string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for s := range in {
			time.Sleep(50 * time.Millisecond)
			out <- prefix + s
		}
	}()
	return out
}

// Stage 4: Filter — keep only strings longer than N
func filterByLength(minLen int, in <-chan string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for s := range in {
			if len(s) >= minLen {
				out <- s
			}
		}
	}()
	return out
}

func main() {
	// ============================================
	// Example 1: Simple 3-stage pipeline
	// ============================================
	fmt.Println("=== Example 1: 3-stage pipeline ===")

	words := generate("hello", "world", "go", "concurrency", "pipeline", "hi")
	uppered := toUpper(words)
	prefixed := addPrefix("[processed] ", uppered)

	start := time.Now()
	for result := range prefixed {
		fmt.Println(" ", result)
	}
	fmt.Printf("  Pipeline time: %v\n", time.Since(start))

	// ============================================
	// Example 2: Pipeline with filter
	// ============================================
	fmt.Println("\n=== Example 2: Pipeline with filter ===")

	ch := generate("go", "rust", "python", "c", "javascript", "typescript")
	ch = toUpper(ch)
	ch = addPrefix("LANG:", ch)
	ch = filterByLength(10, ch) // only keep results >= 10 chars

	for result := range ch {
		fmt.Println(" ", result) // filters out short ones like "LANG:GO", "LANG:C"
	}

	// ============================================
	// Example 3: Number pipeline (generate → square → sum)
	// ============================================
	fmt.Println("\n=== Example 3: Number pipeline ===")

	nums := generateNums(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	squared := squareNums(nums)
	evenOnly := filterEven(squared)

	sum := 0
	for n := range evenOnly {
		sum += n
		fmt.Printf("  Got: %d (running sum: %d)\n", n, sum)
	}
	fmt.Printf("  Sum of even squares: %d\n", sum)
}

func generateNums(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

func squareNums(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			out <- n * n
		}
	}()
	return out
}

func filterEven(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			if n%2 == 0 {
				out <- n
			}
		}
	}()
	return out
}
