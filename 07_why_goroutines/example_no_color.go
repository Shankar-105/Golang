//go:build ignore

package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

// ============================================
// Go has NO "function coloring" problem.
//
// In Python asyncio, you have two worlds:
//   - async def fetch():     # "async" function (blue)
//   - def process():         # "sync" function (red)
//   - You can't call blue from red without asyncio.run()
//   - Once one func is async, everything above it must be async too
//
// In Go, there's no distinction. ANY function can be run as a goroutine.
// The same function works both synchronously and concurrently.
// ============================================

// fetchURL is a regular function — no special syntax.
// In Python, this would HAVE to be `async def` if you wanted it concurrent.
func fetchURL(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Return first 100 chars
	result := string(body)
	if len(result) > 100 {
		result = result[:100]
	}
	return result, nil
}

func main() {
	urls := []string{
		"https://httpbin.org/get",
		"https://httpbin.org/ip",
		"https://httpbin.org/user-agent",
	}

	// ============================================
	// Sequential — same function, no special syntax
	// ============================================
	fmt.Println("=== Sequential (like regular Python) ===")
	for _, url := range urls {
		body, err := fetchURL(url) // called normally
		if err != nil {
			fmt.Printf("Error fetching %s: %v\n", url, err)
			continue
		}
		fmt.Printf("%s → %s...\n", url, strings.TrimSpace(body[:50]))
	}

	// ============================================
	// Concurrent — SAME function, just add `go`
	// ============================================
	fmt.Println("\n=== Concurrent (same function, just `go`) ===")

	var wg sync.WaitGroup
	results := make([]string, len(urls)) // pre-allocate results slice

	for i, url := range urls {
		wg.Add(1)
		go func(idx int, u string) { // launch as goroutine
			defer wg.Done()
			body, err := fetchURL(u) // SAME fetchURL — no "async" version needed!
			if err != nil {
				results[idx] = fmt.Sprintf("Error: %v", err)
				return
			}
			results[idx] = body
		}(i, url)
	}

	wg.Wait()

	for i, r := range results {
		preview := r
		if len(preview) > 50 {
			preview = preview[:50]
		}
		fmt.Printf("%s → %s...\n", urls[i], strings.TrimSpace(preview))
	}

	// ============================================
	// The Python asyncio equivalent would require:
	//
	//   import asyncio, aiohttp
	//
	//   async def fetch_url(url):           # MUST be async
	//       async with aiohttp.ClientSession() as session:  # MUST use async client
	//           async with session.get(url) as resp:         # MUST use async context manager
	//               return await resp.text()                  # MUST await
	//
	//   async def main():                   # MUST be async
	//       tasks = [fetch_url(u) for u in urls]
	//       results = await asyncio.gather(*tasks)  # MUST await gather
	//
	//   asyncio.run(main())                 # special entry point
	//
	// Count the "MUST be async" — that's function coloring.
	// In Go: add `go` in front. Done.
	// ============================================
}
