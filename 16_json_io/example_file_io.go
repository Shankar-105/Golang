//go:build ignore

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// ──────────────────────────────────────────────────────────────
// File I/O in Go vs Python
//
// Python                              Go
// ──────                              ──
// open("f.txt")                       os.Open("f.txt")
// open("f.txt", "w")                  os.Create("f.txt")
// with open(...) as f:                f, _ := os.Open(...); defer f.Close()
// f.read()                            os.ReadFile("f.txt")
// Path("f").write_text("hi")          os.WriteFile("f.txt", data, 0644)
// for line in f:                      scanner.Scan() loop
// ──────────────────────────────────────────────────────────────

func main() {
	fmt.Println("═══ 1. Write & Read Entire File ═══")
	readWriteEntireFile()

	fmt.Println("\n═══ 2. Read Line by Line ═══")
	readLineByLine()

	fmt.Println("\n═══ 3. Buffered Writer ═══")
	bufferedWriterDemo()

	fmt.Println("\n═══ 4. Append to File ═══")
	appendToFile()

	fmt.Println("\n═══ 5. JSON Config File ═══")
	jsonConfigDemo()

	fmt.Println("\n═══ 6. Working with Directories ═══")
	directoryDemo()

	fmt.Println("\n═══ 7. Temp Files ═══")
	tempFileDemo()

	// Clean up example files
	cleanup()
}

// ──── 1. Write & Read Entire File ───────────────────────────
func readWriteEntireFile() {
	filename := "example_output.txt"

	// Write entire file at once
	// Python: Path("example_output.txt").write_text("Hello\nWorld\nGo is great!\n")
	content := []byte("Hello\nWorld\nGo is great!\n")
	err := os.WriteFile(filename, content, 0644) // 0644 = rw-r--r--
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("  Written to", filename)

	// Read entire file at once
	// Python: text = Path("example_output.txt").read_text()
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  Read %d bytes:\n", len(data))
	fmt.Print("  ", strings.ReplaceAll(string(data), "\n", "\n  "))
}

// ──── 2. Read Line by Line ──────────────────────────────────
func readLineByLine() {
	filename := "example_output.txt"

	// Open file for reading
	// Python: with open("example_output.txt") as f:
	f, err := os.Open(filename) // os.Open = read-only
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close() // Go's version of Python's `with` context manager

	// Scan line by line
	// Python: for line in f:
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text() // like Python's line.strip()
		fmt.Printf("  Line %d: %q\n", lineNum, line)
	}

	// Always check for scanner errors
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

// ──── 3. Buffered Writer ────────────────────────────────────
func bufferedWriterDemo() {
	filename := "buffered_output.txt"

	// Create file for writing
	f, err := os.Create(filename) // os.Create = write, truncates if exists
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Wrap in a buffered writer (more efficient for many small writes)
	// Python equivalent: writing with a buffer before flushing
	w := bufio.NewWriter(f)

	for i := 1; i <= 5; i++ {
		fmt.Fprintf(w, "Line %d: This is buffered content\n", i)
	}

	// CRITICAL: Flush the buffer! Without this, data may not be written.
	// Python does this automatically in `with` block or on .close()
	err = w.Flush()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("  Written 5 lines with buffered writer")

	// Verify
	data, _ := os.ReadFile(filename)
	fmt.Printf("  File size: %d bytes\n", len(data))
}

// ──── 4. Append to File ─────────────────────────────────────
func appendToFile() {
	filename := "append_example.txt"

	// Write initial content
	os.WriteFile(filename, []byte("First line\n"), 0644)

	// Open for appending
	// Python: open("file.txt", "a")
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Append lines
	fmt.Fprintln(f, "Second line (appended)")
	fmt.Fprintln(f, "Third line (appended)")

	// Verify
	data, _ := os.ReadFile(filename)
	fmt.Printf("  File contents after append:\n  %s",
		strings.ReplaceAll(string(data), "\n", "\n  "))
	fmt.Println()
}

// ──── 5. JSON Config File ───────────────────────────────────
// A very common real-world pattern: load/save config as JSON.

type AppConfig struct {
	Server   string   `json:"server"`
	Port     int      `json:"port"`
	Debug    bool     `json:"debug"`
	AllowIPs []string `json:"allow_ips"`
}

func jsonConfigDemo() {
	filename := "config.json"

	// Save config
	config := AppConfig{
		Server:   "localhost",
		Port:     8080,
		Debug:    true,
		AllowIPs: []string{"127.0.0.1", "10.0.0.0/8"},
	}

	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	err = enc.Encode(config)
	f.Close()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("  Saved config to", filename)

	// Load config
	f2, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f2.Close()

	var loaded AppConfig
	err = json.NewDecoder(f2).Decode(&loaded)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("  Loaded: server=%s, port=%d, debug=%v\n",
		loaded.Server, loaded.Port, loaded.Debug)
	fmt.Printf("  Allow IPs: %v\n", loaded.AllowIPs)
}

// ──── 6. Working with Directories ───────────────────────────
func directoryDemo() {
	dirName := "example_dir"

	// Create directory (like Python's os.makedirs)
	err := os.MkdirAll(dirName, 0755)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("  Created directory:", dirName)

	// Create some files in it
	for i := 1; i <= 3; i++ {
		path := filepath.Join(dirName, fmt.Sprintf("file%d.txt", i))
		os.WriteFile(path, []byte(fmt.Sprintf("Content of file %d", i)), 0644)
	}

	// List directory contents (like Python's os.listdir)
	entries, err := os.ReadDir(dirName)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("  Directory listing:")
	for _, entry := range entries {
		info, _ := entry.Info()
		fmt.Printf("    %s  (%d bytes, dir=%v)\n",
			entry.Name(), info.Size(), entry.IsDir())
	}

	// Walk a directory tree (like Python's os.walk)
	// filepath.Walk visits every file/dir recursively
	fmt.Println("  Walking directory tree:")
	filepath.Walk(dirName, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		indent := strings.Repeat("  ", strings.Count(path, string(os.PathSeparator)))
		fmt.Printf("    %s%s\n", indent, info.Name())
		return nil
	})

	// Check if file/directory exists (like Python's os.path.exists)
	if _, err := os.Stat(dirName); err == nil {
		fmt.Println("  Directory exists: true")
	}
	if _, err := os.Stat("nonexistent"); os.IsNotExist(err) {
		fmt.Println("  'nonexistent' exists: false")
	}
}

// ──── 7. Temp Files ─────────────────────────────────────────
func tempFileDemo() {
	// Create a temporary file (like Python's tempfile.NamedTemporaryFile)
	tmpFile, err := os.CreateTemp("", "gopractice-*.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpFile.Name()) // clean up when done

	fmt.Println("  Temp file:", tmpFile.Name())

	// Write to temp file
	tmpFile.WriteString("temporary data\n")
	tmpFile.Close()

	// Read it back
	data, _ := os.ReadFile(tmpFile.Name())
	fmt.Printf("  Temp file contents: %q\n", string(data))

	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "gopractice-dir-*")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	fmt.Println("  Temp dir:", tmpDir)
}

// ──── Cleanup ───────────────────────────────────────────────
func cleanup() {
	os.Remove("example_output.txt")
	os.Remove("buffered_output.txt")
	os.Remove("append_example.txt")
	os.Remove("config.json")
	os.RemoveAll("example_dir")
}
