//go:build ignore

package main

import (
	"fmt"
	"log"
	"net/http"
)

// ============================================
// The simplest HTTP server in Go
//
// Python equivalent:
//   from fastapi import FastAPI
//   app = FastAPI()
//   @app.get("/")
//   async def root(): return {"message": "Hello!"}
//
// Run: go run example_hello_server.go
// Visit: http://localhost:8080
// Stop: Ctrl+C
// ============================================

func main() {
	// Register handlers on the default mux
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/hello", helloHandler)
	http.HandleFunc("/about", aboutHandler)

	fmt.Println("=== Hello Server ===")
	fmt.Println("Listening on http://localhost:8080")
	fmt.Println("Try these URLs:")
	fmt.Println("  http://localhost:8080/")
	fmt.Println("  http://localhost:8080/hello")
	fmt.Println("  http://localhost:8080/hello?name=Alice")
	fmt.Println("  http://localhost:8080/about")
	fmt.Println("Press Ctrl+C to stop")

	// Start the server — this blocks forever (until Ctrl+C)
	// Each incoming request is handled in its own goroutine automatically!
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

// homeHandler handles GET /
func homeHandler(w http.ResponseWriter, r *http.Request) {
	// w = where you write the response (like return in FastAPI)
	// r = the incoming request (like request: Request in FastAPI)
	fmt.Fprintln(w, "Welcome to Go HTTP!")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Available routes:")
	fmt.Fprintln(w, "  /       - this page")
	fmt.Fprintln(w, "  /hello  - greeting (try ?name=YourName)")
	fmt.Fprintln(w, "  /about  - about this server")
}

// helloHandler handles GET /hello?name=xxx
func helloHandler(w http.ResponseWriter, r *http.Request) {
	// Query parameters — like request.query_params in FastAPI
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "World"
	}
	fmt.Fprintf(w, "Hello, %s!\n", name)
}

// aboutHandler handles GET /about
func aboutHandler(w http.ResponseWriter, r *http.Request) {
	// Set a custom header
	w.Header().Set("X-Powered-By", "Go")

	fmt.Fprintln(w, "About This Server")
	fmt.Fprintln(w, "-----------------")
	fmt.Fprintf(w, "Method: %s\n", r.Method)
	fmt.Fprintf(w, "Path: %s\n", r.URL.Path)
	fmt.Fprintf(w, "User-Agent: %s\n", r.Header.Get("User-Agent"))
	fmt.Fprintf(w, "Remote Addr: %s\n", r.RemoteAddr)
}
