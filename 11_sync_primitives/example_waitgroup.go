//go:build ignore

package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	// ============================================
	// Example 1: Basic WaitGroup usage
	// ============================================
	fmt.Println("=== Example 1: Basic WaitGroup ===")

	var wg sync.WaitGroup

	tasks := []string{"download", "parse", "save", "notify"}

	for _, task := range tasks {
		wg.Add(1) // MUST be before `go`
		go func(name string) {
			defer wg.Done()
			fmt.Printf("  [%s] starting...\n", name)
			time.Sleep(time.Duration(100+len(name)*50) * time.Millisecond)
			fmt.Printf("  [%s] done!\n", name)
		}(task)
	}

	fmt.Println("  Waiting for all tasks...")
	wg.Wait()
	fmt.Println("  All tasks completed!")

	// ============================================
	// Example 2: WaitGroup with error collection
	// WaitGroup doesn't return errors — you need a separate mechanism
	// ============================================
	fmt.Println("\n=== Example 2: WaitGroup + error collection ===")

	var wg2 sync.WaitGroup
	errCh := make(chan error, 5) // buffered channel for errors

	urls := []string{"ok.com", "fail.com", "ok2.com", "fail2.com"}

	for _, url := range urls {
		wg2.Add(1)
		go func(u string) {
			defer wg2.Done()
			// Simulate request
			if u[0:4] == "fail" {
				errCh <- fmt.Errorf("failed to fetch %s", u)
				return
			}
			fmt.Printf("  Fetched %s successfully\n", u)
		}(url)
	}

	// Wait in a goroutine so we can close errCh after all are done
	go func() {
		wg2.Wait()
		close(errCh)
	}()

	// Collect all errors
	for err := range errCh {
		fmt.Printf("  ERROR: %v\n", err)
	}

	// ============================================
	// Example 3: Nested WaitGroups (waiting for groups of groups)
	// ============================================
	fmt.Println("\n=== Example 3: Nested WaitGroups ===")

	var outer sync.WaitGroup
	phases := []string{"init", "process", "cleanup"}

	for _, phase := range phases {
		outer.Add(1)
		go func(p string) {
			defer outer.Done()

			var inner sync.WaitGroup
			fmt.Printf("  Phase '%s': starting 3 sub-tasks\n", p)

			for j := 1; j <= 3; j++ {
				inner.Add(1)
				go func(subID int) {
					defer inner.Done()
					time.Sleep(100 * time.Millisecond)
					fmt.Printf("    Phase '%s', sub-task %d: done\n", p, subID)
				}(j)
			}

			inner.Wait()
			fmt.Printf("  Phase '%s': complete\n", p)
		}(phase)
	}

	outer.Wait()
	fmt.Println("All phases complete!")
}
