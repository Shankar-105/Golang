//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// ============================================
// HTTP Client — making outgoing requests from Go
//
// Python equivalent: requests / httpx library
// Go:    net/http (built-in, no pip install needed!)
//
// Run: go run example_http_client.go
// (This example hits public APIs — needs internet)
// ============================================

func main() {
	// ============================================
	// Example 1: Simple GET request
	// ============================================
	fmt.Println("=== Example 1: Simple GET ===")
	simpleGet()

	// ============================================
	// Example 2: GET with timeout
	// ============================================
	fmt.Println("\n=== Example 2: GET with timeout ===")
	getWithTimeout()

	// ============================================
	// Example 3: Custom request with headers
	// ============================================
	fmt.Println("\n=== Example 3: Custom request ===")
	customRequest()

	// ============================================
	// Example 4: POST with JSON body
	// ============================================
	fmt.Println("\n=== Example 4: POST JSON ===")
	postJSON()

	// ============================================
	// Example 5: Decode JSON response into struct
	// ============================================
	fmt.Println("\n=== Example 5: Decode JSON ===")
	decodeJSON()
}

// Example 1: Simple GET request
// Python: resp = requests.get(url)
func simpleGet() {
	resp, err := http.Get("https://httpbin.org/get")
	if err != nil {
		log.Printf("  Error: %v", err)
		return
	}
	defer resp.Body.Close() // ALWAYS close the body!

	fmt.Printf("  Status: %s\n", resp.Status)
	fmt.Printf("  Status Code: %d\n", resp.StatusCode)

	// Read first 200 bytes of body
	body := make([]byte, 200)
	n, _ := resp.Body.Read(body)
	fmt.Printf("  Body (first %d bytes): %s...\n", n, string(body[:n]))
}

// Example 2: GET with timeout
// Python: requests.get(url, timeout=5)
func getWithTimeout() {
	// Create a client with a timeout
	client := &http.Client{
		Timeout: 5 * time.Second, // fail if response takes > 5s
	}

	resp, err := client.Get("https://httpbin.org/delay/1") // 1 second delay
	if err != nil {
		log.Printf("  Error: %v", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("  Status: %s (within timeout)\n", resp.Status)
}

// Example 3: Custom request with headers
// Python: requests.get(url, headers={"Authorization": "Bearer ..."})
func customRequest() {
	// Build the request manually
	req, err := http.NewRequest("GET", "https://httpbin.org/headers", nil)
	if err != nil {
		log.Printf("  Error creating request: %v", err)
		return
	}

	// Add custom headers
	req.Header.Set("Authorization", "Bearer my-secret-token")
	req.Header.Set("X-Custom-Header", "go-client")
	req.Header.Set("Accept", "application/json")

	// Send the request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req) // Do() sends any request
	if err != nil {
		log.Printf("  Error: %v", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("  Response:\n%s\n", string(body))
}

// Example 4: POST with JSON body
// Python: requests.post(url, json={"key": "value"})
func postJSON() {
	// Build JSON body using a pipe (encoder → reader)
	pr, pw := io.Pipe()
	go func() {
		json.NewEncoder(pw).Encode(map[string]string{
			"name":    "Alice",
			"message": "Hello from Go!",
		})
		pw.Close()
	}()

	resp, err := http.Post("https://httpbin.org/post", "application/json", pr)
	if err != nil {
		log.Printf("  Error: %v", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("  Status: %s\n", resp.Status)

	// Read response
	body, _ := io.ReadAll(resp.Body)
	// Print just the first 300 bytes
	if len(body) > 300 {
		body = body[:300]
	}
	fmt.Printf("  Response: %s...\n", string(body))
}

// Example 5: Decode JSON response into a Go struct
// Python: data = resp.json()
func decodeJSON() {
	// This API returns JSON about an IP address
	type IPInfo struct {
		IP       string `json:"ip"`
		Hostname string `json:"hostname"`
		City     string `json:"city"`
		Region   string `json:"region"`
		Country  string `json:"country"`
		Org      string `json:"org"`
	}

	resp, err := http.Get("https://ipinfo.io/json")
	if err != nil {
		log.Printf("  Error: %v", err)
		return
	}
	defer resp.Body.Close()

	// Decode JSON directly from the response body
	var info IPInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		log.Printf("  Decode error: %v", err)
		return
	}

	fmt.Printf("  Your IP: %s\n", info.IP)
	fmt.Printf("  Country: %s\n", info.Country)
	fmt.Printf("  City:    %s\n", info.City)
	fmt.Printf("  Org:     %s\n", info.Org)
}
