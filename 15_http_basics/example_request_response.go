//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// ============================================
// Deep dive into Request and Response objects
//
// r *http.Request  = everything about the incoming request
// w http.ResponseWriter = your way to send a response
//
// Run: go run example_request_response.go
// Test with: curl, browser, or httpie
// ============================================

func main() {
	mux := http.NewServeMux()

	// ---- Route registrations ----
	mux.HandleFunc("GET /inspect", inspectHandler)
	mux.HandleFunc("GET /query", queryParamsHandler)
	mux.HandleFunc("GET /headers", headersHandler)
	mux.HandleFunc("POST /echo", echoBodyHandler)
	mux.HandleFunc("POST /form", formHandler)

	// ---- Status code examples ----
	mux.HandleFunc("GET /status/ok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK) // 200 (default, but explicit)
		fmt.Fprintln(w, "OK!")
	})

	mux.HandleFunc("GET /status/created", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated) // 201
		fmt.Fprintln(w, "Resource created!")
	})

	mux.HandleFunc("GET /status/notfound", func(w http.ResponseWriter, r *http.Request) {
		// http.Error is a shorthand for setting status + writing body
		http.Error(w, "the thing you wanted doesn't exist", http.StatusNotFound)
	})

	mux.HandleFunc("GET /status/error", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "something went terribly wrong", http.StatusInternalServerError)
	})

	// ---- JSON response ----
	mux.HandleFunc("GET /json", jsonResponseHandler)

	fmt.Println("=== Request/Response Demo ===")
	fmt.Println("Listening on http://localhost:8080")
	fmt.Println("")
	fmt.Println("Try these:")
	fmt.Println("  curl http://localhost:8080/inspect")
	fmt.Println("  curl 'http://localhost:8080/query?name=Alice&age=25'")
	fmt.Println("  curl http://localhost:8080/headers -H 'X-Custom: hello'")
	fmt.Println("  curl -X POST http://localhost:8080/echo -d 'Hello Go!'")
	fmt.Println("  curl -X POST http://localhost:8080/form -d 'user=alice&pass=secret'")
	fmt.Println("  curl http://localhost:8080/json")
	fmt.Println("  curl http://localhost:8080/status/notfound")

	log.Fatal(http.ListenAndServe(":8080", mux))
}

// inspectHandler — shows all request details
func inspectHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "=== Request Inspector ===")
	fmt.Fprintf(w, "Method:      %s\n", r.Method)
	fmt.Fprintf(w, "URL:         %s\n", r.URL.String())
	fmt.Fprintf(w, "Path:        %s\n", r.URL.Path)
	fmt.Fprintf(w, "RawQuery:    %s\n", r.URL.RawQuery)
	fmt.Fprintf(w, "Proto:       %s\n", r.Proto)
	fmt.Fprintf(w, "Host:        %s\n", r.Host)
	fmt.Fprintf(w, "RemoteAddr:  %s\n", r.RemoteAddr)
	fmt.Fprintf(w, "ContentLen:  %d\n", r.ContentLength)

	fmt.Fprintln(w, "\n--- Headers ---")
	for name, values := range r.Header {
		fmt.Fprintf(w, "  %s: %s\n", name, strings.Join(values, ", "))
	}
}

// queryParamsHandler — demonstrates reading query parameters
// curl 'http://localhost:8080/query?name=Alice&age=25&tags=go&tags=http'
func queryParamsHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query() // returns url.Values (map[string][]string)

	// Get single value (returns "" if missing)
	name := query.Get("name")
	ageStr := query.Get("age")

	// Get multiple values for the same key
	tags := query["tags"] // []string

	// Check if a key exists
	_, hasName := query["name"]

	fmt.Fprintln(w, "=== Query Parameters ===")
	fmt.Fprintf(w, "name (string):  %q\n", name)
	fmt.Fprintf(w, "age (string):   %q\n", ageStr)
	fmt.Fprintf(w, "tags ([]string): %v\n", tags)
	fmt.Fprintf(w, "has 'name' key: %t\n", hasName)

	// Convert string to int (query params are always strings!)
	if ageStr != "" {
		age, err := strconv.Atoi(ageStr)
		if err != nil {
			http.Error(w, "age must be a number", http.StatusBadRequest)
			return
		}
		fmt.Fprintf(w, "age (int):      %d\n", age)
	}
}

// headersHandler — demonstrates reading/writing headers
func headersHandler(w http.ResponseWriter, r *http.Request) {
	// Read request headers
	contentType := r.Header.Get("Content-Type")
	userAgent := r.Header.Get("User-Agent")
	custom := r.Header.Get("X-Custom")

	// Write response headers (MUST be before WriteHeader/Write)
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("X-Server", "GoHTTPDemo/1.0")
	w.Header().Add("X-Multi", "value1")
	w.Header().Add("X-Multi", "value2") // Add allows multiple values

	fmt.Fprintln(w, "=== Headers Demo ===")
	fmt.Fprintf(w, "Content-Type: %q\n", contentType)
	fmt.Fprintf(w, "User-Agent:   %q\n", userAgent)
	fmt.Fprintf(w, "X-Custom:     %q\n", custom)
}

// echoBodyHandler — reads the request body and echoes it back
func echoBodyHandler(w http.ResponseWriter, r *http.Request) {
	// Read the entire body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close() // always close the body

	fmt.Fprintln(w, "=== Echo ===")
	fmt.Fprintf(w, "Body length: %d bytes\n", len(body))
	fmt.Fprintf(w, "Body: %s\n", string(body))
}

// formHandler — reads URL-encoded form data
// curl -X POST http://localhost:8080/form -d 'user=alice&pass=secret'
func formHandler(w http.ResponseWriter, r *http.Request) {
	// ParseForm must be called before FormValue
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form data", http.StatusBadRequest)
		return
	}

	user := r.FormValue("user")
	pass := r.FormValue("pass")

	fmt.Fprintln(w, "=== Form Data ===")
	fmt.Fprintf(w, "user: %q\n", user)
	fmt.Fprintf(w, "pass: %q\n", pass)
}

// jsonResponseHandler — returns a JSON response
func jsonResponseHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"message": "Hello from Go!",
		"status":  "ok",
		"code":    200,
		"items":   []string{"alpha", "beta", "gamma"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
