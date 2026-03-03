//go:build ignore

package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// ──────────────────────────────────────────────────────────────
// io.Reader / io.Writer — the foundation of all I/O in Go
//
// Python equivalent: file-like objects with .read() and .write()
//   f = open("data.txt")       # has .read()
//   buf = io.BytesIO()          # has .read() and .write()
//
// Go's approach is interface-based. Any type that implements:
//   Read(p []byte) (n int, err error)   → is an io.Reader
//   Write(p []byte) (n int, err error)  → is an io.Writer
// ──────────────────────────────────────────────────────────────

func main() {
	fmt.Println("═══ 1. strings.NewReader — string as io.Reader ═══")
	stringsReaderDemo()

	fmt.Println("\n═══ 2. bytes.Buffer — in-memory read/write ═══")
	bytesBufferDemo()

	fmt.Println("\n═══ 3. io.Copy — universal pipe ═══")
	ioCopyDemo()

	fmt.Println("\n═══ 4. io.TeeReader — read + log simultaneously ═══")
	teeReaderDemo()

	fmt.Println("\n═══ 5. io.LimitReader — cap how much you read ═══")
	limitReaderDemo()

	fmt.Println("\n═══ 6. io.MultiReader — concatenate readers ═══")
	multiReaderDemo()

	fmt.Println("\n═══ 7. io.MultiWriter — write to many at once ═══")
	multiWriterDemo()

	fmt.Println("\n═══ 8. io.ReadAll — slurp everything ═══")
	readAllDemo()

	fmt.Println("\n═══ 9. Writing functions that accept io.Reader ═══")
	genericReaderDemo()
}

// ──── 1. strings.NewReader ──────────────────────────────────
func stringsReaderDemo() {
	// strings.NewReader creates an io.Reader from a string.
	// Python equivalent: io.StringIO("hello world")
	r := strings.NewReader("hello world from strings.NewReader")

	// Read into a byte buffer
	buf := make([]byte, 12)
	n, err := r.Read(buf)
	fmt.Printf("Read %d bytes: %q\n", n, string(buf[:n]))

	// Read the rest
	n, err = r.Read(buf)
	fmt.Printf("Read %d bytes: %q\n", n, string(buf[:n]))

	// When no more data, Read returns io.EOF
	n, err = r.Read(buf)
	if err == io.EOF {
		fmt.Println("Reached EOF (end of data)")
	}
	_ = n // suppress unused warning
}

// ──── 2. bytes.Buffer ───────────────────────────────────────
func bytesBufferDemo() {
	// bytes.Buffer implements BOTH io.Reader AND io.Writer.
	// Python equivalent: io.BytesIO()
	var buf bytes.Buffer

	// Write to it
	buf.WriteString("Hello, ")
	buf.WriteString("World!")
	fmt.Println("Buffer contents:", buf.String())

	// Read from it (consuming the data)
	data := make([]byte, 5)
	n, _ := buf.Read(data)
	fmt.Printf("Read %d bytes: %q\n", n, string(data[:n]))
	fmt.Println("Remaining:", buf.String())

	// Reset and reuse
	buf.Reset()
	fmt.Fprintf(&buf, "Formatted: %d + %d = %d", 2, 3, 5)
	fmt.Println("After fmt.Fprintf:", buf.String())
}

// ──── 3. io.Copy ────────────────────────────────────────────
func ioCopyDemo() {
	// io.Copy(dst Writer, src Reader) — copies everything from src to dst.
	// Like a universal pipe. Works with any Reader → any Writer.

	src := strings.NewReader("data flowing through io.Copy\n")

	// Copy to stdout
	fmt.Print("  Output: ")
	n, err := io.Copy(os.Stdout, src)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  (copied %d bytes)\n", n)

	// Copy between buffers
	src2 := strings.NewReader("buffer to buffer")
	var dst bytes.Buffer
	io.Copy(&dst, src2)
	fmt.Println("  Buffer result:", dst.String())
}

// ──── 4. io.TeeReader ──────────────────────────────────────
func teeReaderDemo() {
	// TeeReader: everything you read from src ALSO gets written to w.
	// Like Unix `tee` command: cat file | tee copy.txt
	//
	// Real use case: reading an HTTP response body while also
	// logging the raw JSON for debugging.

	original := strings.NewReader("secret message that needs to be logged")
	var logBuf bytes.Buffer

	tee := io.TeeReader(original, &logBuf)

	// Read from tee (this also writes to logBuf)
	data, _ := io.ReadAll(tee)
	fmt.Println("  Read:", string(data))
	fmt.Println("  Log copy:", logBuf.String())
	fmt.Println("  (Both are identical — the tee duplicated the stream)")
}

// ──── 5. io.LimitReader ────────────────────────────────────
func limitReaderDemo() {
	// LimitReader wraps a reader and stops after N bytes.
	// Critical for security: prevent reading a 10GB request body!
	//
	// Real use case:
	//   limited := io.LimitReader(r.Body, 1<<20)  // max 1 MB
	//   json.NewDecoder(limited).Decode(&data)

	big := strings.NewReader("This is a longer string but we only want 10 bytes")
	limited := io.LimitReader(big, 10)

	data, _ := io.ReadAll(limited)
	fmt.Printf("  Limited to 10 bytes: %q\n", string(data))
}

// ──── 6. io.MultiReader ────────────────────────────────────
func multiReaderDemo() {
	// MultiReader concatenates multiple readers into one.
	// Like itertools.chain() for file-like objects in Python.

	header := strings.NewReader("=== HEADER ===\n")
	body := strings.NewReader("This is the body content.\n")
	footer := strings.NewReader("=== FOOTER ===\n")

	combined := io.MultiReader(header, body, footer)

	fmt.Println("  Combined output:")
	io.Copy(os.Stdout, combined)
}

// ──── 7. io.MultiWriter ────────────────────────────────────
func multiWriterDemo() {
	// MultiWriter creates a writer that duplicates writes to all writers.
	// Like Python: writing to multiple files at once.

	var buf1, buf2 bytes.Buffer
	multi := io.MultiWriter(&buf1, &buf2, os.Stdout)

	fmt.Fprint(multi, "  This goes to 3 places at once!\n")
	fmt.Println("  buf1:", buf1.String())
	fmt.Println("  buf2:", buf2.String())
}

// ──── 8. io.ReadAll ─────────────────────────────────────────
func readAllDemo() {
	// ReadAll reads everything from a reader into a byte slice.
	// Python equivalent: f.read() (read entire file)
	//
	// WARNING: Don't use on unbounded data (like HTTP bodies)
	// without a LimitReader first!

	r := strings.NewReader("Read all of this into memory")
	data, err := io.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  Got %d bytes: %s\n", len(data), string(data))
}

// ──── 9. Writing Generic Functions with io.Reader ───────────
// This is the POWER of Go's interfaces. Write a function that takes
// io.Reader, and it works with strings, files, HTTP bodies, buffers...

func countWords(r io.Reader) (int, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return 0, err
	}

	words := 0
	inWord := false
	for _, b := range data {
		if b == ' ' || b == '\n' || b == '\t' {
			inWord = false
		} else if !inWord {
			inWord = true
			words++
		}
	}
	return words, nil
}

func genericReaderDemo() {
	// Same function works with ALL these different sources:

	// From a string
	count, _ := countWords(strings.NewReader("hello world foo bar"))
	fmt.Printf("  String: %d words\n", count)

	// From a buffer
	var buf bytes.Buffer
	buf.WriteString("one two three four five")
	count, _ = countWords(&buf)
	fmt.Printf("  Buffer: %d words\n", count)

	// From a file (if it existed):
	// f, _ := os.Open("data.txt")
	// count, _ = countWords(f)
	//
	// From an HTTP response:
	// count, _ = countWords(resp.Body)
	//
	// The function doesn't care WHERE the data comes from!
	fmt.Println("  (io.Reader makes functions source-agnostic)")
}
