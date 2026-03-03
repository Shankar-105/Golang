//go:build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
)

// ──────────────────────────────────────────────────────────────
// Streaming JSON with json.Encoder / json.Decoder
//
// Key insight: json.Marshal/Unmarshal work on []byte in memory.
// Encoder/Decoder work on io.Writer/io.Reader STREAMS.
//
// Use streaming when:
//   - Reading from HTTP request bodies
//   - Writing to HTTP responses
//   - Processing large JSON files
//   - Piping JSON between services
// ──────────────────────────────────────────────────────────────

type LogEntry struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	Service string `json:"service"`
}

func main() {
	fmt.Println("═══ 1. Encoder — Write JSON to a Writer ═══")
	encoderDemo()

	fmt.Println("\n═══ 2. Decoder — Read JSON from a Reader ═══")
	decoderDemo()

	fmt.Println("\n═══ 3. Streaming Multiple Objects (NDJSON) ═══")
	ndjsonDemo()

	fmt.Println("\n═══ 4. Decoder with Token-by-Token Parsing ═══")
	tokenDemo()

	fmt.Println("\n═══ 5. Encoder Options ═══")
	encoderOptionsDemo()

	fmt.Println("\n═══ 6. Simulated HTTP Handler Pattern ═══")
	httpPatternDemo()
}

// ──── 1. Encoder — Write JSON to a Writer ───────────────────
func encoderDemo() {
	// Write JSON directly to a buffer (in real code, this would be
	// an http.ResponseWriter, a file, or any io.Writer)
	var buf bytes.Buffer

	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")

	entry := LogEntry{
		Level:   "info",
		Message: "Server started",
		Service: "api-gateway",
	}

	err := enc.Encode(entry)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("  Encoded to buffer:")
	fmt.Println(" ", buf.String())

	// Note: Encode() adds a trailing newline automatically.
	// This is intentional — it makes NDJSON (newline-delimited JSON) easy.
}

// ──── 2. Decoder — Read JSON from a Reader ──────────────────
func decoderDemo() {
	// Read JSON from a string (simulating an HTTP request body)
	jsonStr := `{"level":"error","message":"Connection failed","service":"db-pool"}`
	reader := strings.NewReader(jsonStr)

	var entry LogEntry
	err := json.NewDecoder(reader).Decode(&entry)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("  Decoded: [%s] %s (%s)\n", entry.Level, entry.Message, entry.Service)
}

// ──── 3. NDJSON — Multiple JSON Objects in a Stream ─────────
// NDJSON = Newline-Delimited JSON. Each line is a complete JSON object.
// Used by: Docker logs, Elasticsearch bulk API, structured logging.
// Python equivalent: reading multiple json.loads() from lines.

func ndjsonDemo() {
	// Simulate a stream of log entries (one JSON per line)
	stream := strings.NewReader(`{"level":"info","message":"Starting up","service":"web"}
{"level":"warn","message":"High memory","service":"web"}
{"level":"error","message":"Connection lost","service":"db"}
{"level":"info","message":"Reconnected","service":"db"}
`)

	// Decoder reads one object at a time from the stream
	dec := json.NewDecoder(stream)

	fmt.Println("  Processing NDJSON stream:")
	for {
		var entry LogEntry
		err := dec.Decode(&entry)
		if err == io.EOF {
			break // stream exhausted
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("    [%-5s] %-20s (service: %s)\n",
			entry.Level, entry.Message, entry.Service)
	}

	// Writing NDJSON
	fmt.Println("\n  Writing NDJSON:")
	var out bytes.Buffer
	enc := json.NewEncoder(&out)

	entries := []LogEntry{
		{Level: "info", Message: "Request received", Service: "api"},
		{Level: "info", Message: "Processing complete", Service: "api"},
	}
	for _, e := range entries {
		enc.Encode(e) // each Encode adds a newline
	}
	fmt.Print("  ", out.String())
}

// ──── 4. Token-by-Token Parsing ─────────────────────────────
// For very large JSON or when you need to extract specific fields
// without parsing the entire document.

func tokenDemo() {
	jsonStr := `{
		"users": [
			{"name": "Alice", "age": 30},
			{"name": "Bob", "age": 25}
		],
		"total": 2
	}`

	dec := json.NewDecoder(strings.NewReader(jsonStr))

	fmt.Println("  Tokens in the JSON:")
	for {
		t, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		switch v := t.(type) {
		case json.Delim:
			fmt.Printf("    Delimiter: %c\n", v)
		case string:
			fmt.Printf("    String:    %q\n", v)
		case float64:
			fmt.Printf("    Number:    %g\n", v)
		case bool:
			fmt.Printf("    Bool:      %v\n", v)
		case nil:
			fmt.Println("    Null")
		}
	}
}

// ──── 5. Encoder Options ────────────────────────────────────
func encoderOptionsDemo() {
	type HTMLContent struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}

	content := HTMLContent{
		Title: "Hello <World> & \"Friends\"",
		Body:  "<p>This has <b>HTML</b> & special chars</p>",
	}

	// Default: HTML-sensitive characters are escaped
	var buf1 bytes.Buffer
	json.NewEncoder(&buf1).Encode(content)
	fmt.Println("  Default (HTML escaped):")
	fmt.Println(" ", buf1.String())

	// SetEscapeHTML(false): don't escape <, >, &
	var buf2 bytes.Buffer
	enc := json.NewEncoder(&buf2)
	enc.SetEscapeHTML(false)
	enc.Encode(content)
	fmt.Println("  SetEscapeHTML(false):")
	fmt.Println(" ", buf2.String())
}

// ──── 6. Simulated HTTP Handler Pattern ─────────────────────
// This is exactly how you'd handle JSON in an HTTP handler.
// (Using buffers instead of real HTTP for this demo)

type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CreateUserResponse struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

func httpPatternDemo() {
	// Simulate request body
	requestBody := strings.NewReader(`{"name":"Alice","email":"alice@example.com"}`)

	// Simulate response writer
	var responseBody bytes.Buffer

	// === This is what your HTTP handler would look like ===

	// 1. Decode request
	var req CreateUserRequest
	if err := json.NewDecoder(requestBody).Decode(&req); err != nil {
		fmt.Println("  Error:", err)
		return
	}
	fmt.Printf("  Received request: name=%s, email=%s\n", req.Name, req.Email)

	// 2. Process (create user, etc.)
	resp := CreateUserResponse{
		ID:      42,
		Name:    req.Name,
		Email:   req.Email,
		Message: "User created successfully",
	}

	// 3. Encode response
	// In real HTTP: w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(&responseBody)
	enc.SetIndent("", "  ")
	if err := enc.Encode(resp); err != nil {
		fmt.Println("  Error:", err)
		return
	}

	fmt.Println("  Response body:")
	fmt.Println(" ", responseBody.String())
}
