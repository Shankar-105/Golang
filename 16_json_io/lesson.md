# Lesson 16: JSON, I/O & Encoding

## Python → Go Mental Model

| Python | Go |
|--------|-----|
| `json.dumps(obj)` | `json.Marshal(v)` |
| `json.loads(s)` | `json.Unmarshal(data, &v)` |
| `json.dump(obj, f)` | `json.NewEncoder(w).Encode(v)` |
| `json.load(f)` | `json.NewDecoder(r).Decode(&v)` |
| `open("f.txt")` | `os.Open("f.txt")` |
| `with open(...) as f:` | `f, _ := os.Open(...); defer f.Close()` |
| duck-typed file objects | `io.Reader` / `io.Writer` interfaces |
| `io.StringIO` / `io.BytesIO` | `strings.NewReader` / `bytes.Buffer` |

---

## 1. JSON Basics — `encoding/json`

### 1.1 Marshalling (Go struct → JSON bytes)

```go
type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age"`
}

u := User{Name: "Alice", Email: "alice@example.com", Age: 30}
data, err := json.Marshal(u)
// data = []byte(`{"name":"Alice","email":"alice@example.com","age":30}`)
```

**Python equivalent:**
```python
import json
data = json.dumps({"name": "Alice", "email": "alice@example.com", "age": 30})
```

### 1.2 Unmarshalling (JSON bytes → Go struct)

```go
var u User
err := json.Unmarshal([]byte(`{"name":"Bob","email":"bob@ex.com","age":25}`), &u)
// u.Name == "Bob", u.Email == "bob@ex.com", u.Age == 25
```

**Key difference from Python:** You pass a **pointer** (`&u`) so Go can fill in the struct in-place. Python's `json.loads()` returns a new dict — Go mutates the target.

### 1.3 Struct Tags — The Secret Sauce

Struct tags control how JSON fields are named, whether they're omitted when empty, and more:

```go
type Product struct {
    ID          int      `json:"id"`                  // rename to lowercase
    Name        string   `json:"name"`                // rename
    Price       float64  `json:"price"`               // rename
    Description string   `json:"description,omitempty"` // omit if ""
    InternalSKU string   `json:"-"`                   // NEVER include in JSON
    Tags        []string `json:"tags,omitempty"`       // omit if nil/empty
}
```

**Tag cheat sheet:**

| Tag | Effect |
|-----|--------|
| `json:"name"` | Field appears as `"name"` in JSON |
| `json:"name,omitempty"` | Omit field if zero value (0, "", nil, false, empty slice/map) |
| `json:"-"` | Always skip this field |
| `json:",string"` | Encode number/bool as JSON string |
| No tag | Uses field name as-is (capitalized, since exported) |

### 1.4 Pretty Printing

```go
data, err := json.MarshalIndent(user, "", "  ")
// Like Python's json.dumps(obj, indent=2)
```

---

## 2. Working with Dynamic / Unknown JSON

### 2.1 Unmarshalling into `map[string]any`

When you don't know the JSON structure ahead of time (like Python dicts):

```go
var result map[string]any
err := json.Unmarshal(data, &result)
// result["name"] is any — you need type assertions
name := result["name"].(string)
```

### 2.2 Unmarshalling into `[]any`

For JSON arrays:

```go
var items []any
json.Unmarshal([]byte(`[1, "hello", true]`), &items)
```

### 2.3 `json.RawMessage` — Delay Parsing

Sometimes you want to partially parse JSON and handle a field later:

```go
type Event struct {
    Type    string          `json:"type"`
    Payload json.RawMessage `json:"payload"` // kept as raw bytes
}

// Parse the envelope first
var event Event
json.Unmarshal(data, &event)

// Then parse payload based on type
switch event.Type {
case "user_created":
    var u User
    json.Unmarshal(event.Payload, &u)
case "order_placed":
    var o Order
    json.Unmarshal(event.Payload, &o)
}
```

---

## 3. Streaming JSON — Encoder & Decoder

### 3.1 Why Streaming?

`json.Marshal` / `json.Unmarshal` work on **entire byte slices in memory**. For HTTP bodies, files, or large data, use streaming:

- `json.NewEncoder(w io.Writer)` — writes JSON directly to a writer (HTTP response, file, etc.)
- `json.NewDecoder(r io.Reader)` — reads JSON directly from a reader (HTTP request body, file, etc.)

```go
// Writing JSON to HTTP response (streaming)
func handler(w http.ResponseWriter, r *http.Request) {
    user := User{Name: "Alice", Age: 30}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user) // writes directly to ResponseWriter
}

// Reading JSON from HTTP request body (streaming)
func createUser(w http.ResponseWriter, r *http.Request) {
    var user User
    err := json.NewDecoder(r.Body).Decode(&user) // reads directly from Body
    if err != nil {
        http.Error(w, "invalid JSON", http.StatusBadRequest)
        return
    }
}
```

### 3.2 Encoder vs Marshal — When to Use Which

| Use Case | Use |
|----------|-----|
| Writing JSON to HTTP response | `json.NewEncoder(w).Encode(v)` |
| Reading JSON from HTTP request | `json.NewDecoder(r.Body).Decode(&v)` |
| Converting struct to `[]byte` for storage/logging | `json.Marshal(v)` |
| Parsing a `[]byte` you already have | `json.Unmarshal(data, &v)` |

---

## 4. Custom JSON Marshalling

### 4.1 `json.Marshaler` and `json.Unmarshaler` Interfaces

Implement these to control exactly how a type converts to/from JSON:

```go
type Marshaler interface {
    MarshalJSON() ([]byte, error)
}

type Unmarshaler interface {
    UnmarshalJSON([]byte) error
}
```

Common use cases:
- Custom date formats (Go's `time.Time` uses RFC3339 by default)
- Enum-like types that should serialize as strings
- Wrapping/unwrapping nested structures

```go
type Status int

const (
    Active Status = iota
    Inactive
    Suspended
)

func (s Status) MarshalJSON() ([]byte, error) {
    names := map[Status]string{Active: "active", Inactive: "inactive", Suspended: "suspended"}
    name, ok := names[s]
    if !ok {
        return nil, fmt.Errorf("unknown status: %d", s)
    }
    return json.Marshal(name) // wrap string in quotes
}

func (s *Status) UnmarshalJSON(data []byte) error {
    var name string
    if err := json.Unmarshal(data, &name); err != nil {
        return err
    }
    values := map[string]Status{"active": Active, "inactive": Inactive, "suspended": Suspended}
    val, ok := values[name]
    if !ok {
        return fmt.Errorf("unknown status: %s", name)
    }
    *s = val
    return nil
}
```

---

## 5. The `io.Reader` / `io.Writer` Interfaces

### 5.1 The Core Interfaces

These two interfaces are **the foundation of all I/O in Go**:

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}
```

**Everything** that reads data implements `io.Reader`. **Everything** that writes data implements `io.Writer`:

| Type | `io.Reader`? | `io.Writer`? |
|------|:---:|:---:|
| `os.File` | ✅ | ✅ |
| `http.Request.Body` | ✅ | ❌ |
| `http.ResponseWriter` | ❌ | ✅ |
| `bytes.Buffer` | ✅ | ✅ |
| `strings.NewReader()` | ✅ | ❌ |
| `os.Stdout` | ❌ | ✅ |
| `net.Conn` (TCP) | ✅ | ✅ |

### 5.2 Python Comparison

```python
# Python: file-like objects with .read() and .write()
f = open("data.txt")      # has .read()
buf = io.BytesIO()         # has .read() and .write()
response.content            # has .read() (via .raw)
```

Go's approach is more explicit and composable. Functions accept `io.Reader` or `io.Writer`, so any compatible type works.

### 5.3 `io.Copy` — The Universal Pipe

```go
// Copy from any Reader to any Writer
n, err := io.Copy(dst, src)
```

This works for:
- Downloading a file: `io.Copy(file, resp.Body)`
- Uploading: `io.Copy(httpWriter, file)`
- Piping: `io.Copy(os.Stdout, strings.NewReader("hello"))`

---

## 6. File I/O

### 6.1 Reading Files

```go
// Read entire file into memory (like Python's Path("f").read_text())
data, err := os.ReadFile("config.json")

// Read line by line (like Python's `for line in f:`)
f, err := os.Open("data.txt")
if err != nil { log.Fatal(err) }
defer f.Close()

scanner := bufio.NewScanner(f)
for scanner.Scan() {
    line := scanner.Text()
    fmt.Println(line)
}
if err := scanner.Err(); err != nil {
    log.Fatal(err)
}
```

### 6.2 Writing Files

```go
// Write entire file at once (like Python's Path("f").write_text("..."))
err := os.WriteFile("output.txt", []byte("hello\nworld\n"), 0644)

// Write with a buffered writer (better for many small writes)
f, err := os.Create("output.txt")
if err != nil { log.Fatal(err) }
defer f.Close()

w := bufio.NewWriter(f)
fmt.Fprintln(w, "line 1")
fmt.Fprintln(w, "line 2")
w.Flush() // DON'T FORGET — buffered data isn't written until Flush()
```

### 6.3 File Permissions

Go uses Unix-style octal permissions:

| Pattern | Meaning |
|---------|---------|
| `0644` | Owner read/write, group/others read-only |
| `0755` | Owner all, group/others read+execute |
| `0600` | Owner read/write only (secrets!) |

---

## 7. Composing I/O — Advanced Patterns

### 7.1 `io.TeeReader` — Read and Copy Simultaneously

```go
// Like Unix `tee` command — read from src, and everything read also goes to dst
tee := io.TeeReader(resp.Body, logFile)
json.NewDecoder(tee).Decode(&data)
// Now resp.Body was decoded AND a copy was written to logFile
```

### 7.2 `io.LimitReader` — Read at Most N Bytes

```go
// Prevent reading more than 1MB (protect against huge payloads)
limited := io.LimitReader(r.Body, 1<<20) // 1 MB
json.NewDecoder(limited).Decode(&data)
```

### 7.3 `io.Pipe` — In-Memory Pipe

```go
pr, pw := io.Pipe()
go func() {
    json.NewEncoder(pw).Encode(bigStruct)
    pw.Close()
}()
// pr can be used as request body, or read from in another goroutine
```

### 7.4 `io.MultiReader` — Concatenate Readers

```go
header := strings.NewReader("HEADER\n")
body := strings.NewReader("body content")
footer := strings.NewReader("\nFOOTER")
combined := io.MultiReader(header, body, footer)
io.Copy(os.Stdout, combined) // prints: HEADER\nbody content\nFOOTER
```

### 7.5 `io.MultiWriter` — Write to Multiple Destinations

```go
logFile, _ := os.Create("app.log")
multi := io.MultiWriter(os.Stdout, logFile)
fmt.Fprintln(multi, "this goes to both stdout AND the log file")
```

---

## 8. Other Encoding Formats

### 8.1 CSV

```go
import "encoding/csv"

// Writing
w := csv.NewWriter(os.Stdout)
w.Write([]string{"name", "age", "email"})
w.Write([]string{"Alice", "30", "alice@example.com"})
w.Flush()

// Reading
r := csv.NewReader(strings.NewReader(csvData))
records, err := r.ReadAll()
```

### 8.2 XML

```go
import "encoding/xml"

type Person struct {
    XMLName xml.Name `xml:"person"`
    Name    string   `xml:"name"`
    Age     int      `xml:"age"`
}
// Same Marshal/Unmarshal pattern as JSON
```

### 8.3 Base64

```go
import "encoding/base64"

encoded := base64.StdEncoding.EncodeToString([]byte("Hello, World!"))
decoded, err := base64.StdEncoding.DecodeString(encoded)
```

---

## Run the Examples

```bash
go run example_json_basics.go
go run example_json_advanced.go
go run example_io_readers.go
go run example_file_io.go
go run example_streaming_json.go
go run example_custom_marshal.go
```
