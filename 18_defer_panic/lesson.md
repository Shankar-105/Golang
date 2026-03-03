# Lesson 18: `defer`, `panic`, `recover`

## Python → Go Mental Model

| Python | Go |
|--------|-----|
| `with open("f") as f:` (context manager) | `f, _ := os.Open("f"); defer f.Close()` |
| `try: ... finally: cleanup()` | `defer cleanup()` |
| `raise ValueError("...")` | `panic("...")` (but DON'T use this normally!) |
| `except Exception as e:` | `recover()` inside a deferred function |
| `raise` (re-raise) | `panic(r)` after `recover()` |
| `atexit.register(fn)` | `defer fn()` |
| `contextlib.ExitStack` | Multiple `defer` calls (LIFO order) |

---

## 1. `defer` — Guaranteed Cleanup

### 1.1 The Basics

`defer` schedules a function call to run **when the surrounding function returns** — no matter how it returns (normal return, early return, or even panic).

```go
func readFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close()  // runs when readFile returns, guaranteed

    // ... use f ...
    return nil
}
```

**Python equivalent:**
```python
def read_file(path):
    with open(path) as f:  # __exit__ guarantees cleanup
        # ... use f ...
```

### 1.2 LIFO Order

Multiple `defer` calls execute in **Last-In-First-Out** order (like a stack):

```go
defer fmt.Println("first")   // runs 3rd
defer fmt.Println("second")  // runs 2nd  
defer fmt.Println("third")   // runs 1st
// Output: third, second, first
```

This is intentional — resources should be released in reverse order of acquisition.

### 1.3 `defer` Evaluates Arguments Immediately

```go
x := 10
defer fmt.Println(x)  // captures x=10 NOW, not at execution time
x = 20
// Prints: 10 (not 20!)
```

To defer with the latest value, use a closure:
```go
x := 10
defer func() { fmt.Println(x) }()  // closure captures the variable, not the value
x = 20
// Prints: 20
```

---

## 2. Common `defer` Patterns

### 2.1 File Handling
```go
f, err := os.Open("data.txt")
if err != nil { return err }
defer f.Close()
```

### 2.2 Mutex Locking
```go
mu.Lock()
defer mu.Unlock()
// ... critical section ...
```

### 2.3 HTTP Response Bodies
```go
resp, err := http.Get(url)
if err != nil { return err }
defer resp.Body.Close()  // ALWAYS close the body!
```

### 2.4 Database Transactions
```go
tx, err := db.Begin()
if err != nil { return err }
defer tx.Rollback()  // rollback if not committed

// ... do work ...
return tx.Commit()  // commit; the deferred Rollback is a no-op after Commit
```

### 2.5 Timing Functions
```go
func timeIt(name string) func() {
    start := time.Now()
    return func() {
        fmt.Printf("%s took %v\n", name, time.Since(start))
    }
}

func doWork() {
    defer timeIt("doWork")()  // note the ()() — first call returns the cleanup func
    // ... work ...
}
```

---

## 3. `panic` — When Things Go Truly Wrong

### 3.1 What is `panic`?

`panic` is Go's mechanism for **unrecoverable errors**. It immediately stops normal execution, runs all deferred functions, then crashes the program.

```go
panic("something terrible happened")
panic(fmt.Sprintf("index %d out of range", i))
panic(errors.New("database connection lost"))
```

### 3.2 When to Use `panic` (Rarely!)

**DO panic for:**
- Programming errors (bugs that should never happen in correct code)
- Unrecoverable initialization failures
- Violated invariants / impossible states

**DON'T panic for:**
- File not found → return `error`
- Invalid user input → return `error`
- Network timeout → return `error`
- Any recoverable condition → return `error`

```go
// ✅ Good: panic for programming errors
func MustCompile(pattern string) *regexp.Regexp {
    re, err := regexp.Compile(pattern)
    if err != nil {
        panic("invalid regex: " + err.Error())
    }
    return re
}

// ❌ Bad: don't panic for normal errors
func LoadConfig(path string) (*Config, error) {
    // panic("file not found")  // NEVER do this!
    return nil, fmt.Errorf("config file not found: %s", path)  // return error instead
}
```

### 3.3 Python Comparison

```python
# Python: exceptions for everything
def divide(a, b):
    if b == 0:
        raise ValueError("division by zero")  # normal error control flow

# Go: errors for expected problems, panic only for bugs
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")  // return error, don't panic
    }
    return a / b, nil
}
```

### 3.4 Built-in Panics

Go panics automatically for some runtime errors:
```go
var s []int
_ = s[10]         // panic: runtime error: index out of range

var m map[string]int
m["key"] = 1       // panic: assignment to entry in nil map

var p *int
_ = *p             // panic: runtime error: invalid memory address (nil pointer)
```

---

## 4. `recover` — Catching Panics

### 4.1 How `recover` Works

`recover()` is only useful **inside a deferred function**. It:
1. Stops the panic from propagating
2. Returns the value passed to `panic()`
3. Returns `nil` if there's no panic

```go
func safeOperation() {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("Recovered from panic:", r)
        }
    }()

    // This panic will be caught
    panic("something bad")
}
```

### 4.2 Recovery in HTTP Servers

The most common real-world use: preventing one bad request from crashing the entire server.

```go
func recoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if r := recover(); r != nil {
                log.Printf("PANIC: %v", r)
                http.Error(w, "Internal Server Error", 500)
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

### 4.3 Python Comparison

```python
# Python: try/except
try:
    risky_operation()
except Exception as e:
    print(f"Caught: {e}")

# Go: defer + recover
func safe() {
    defer func() {
        if r := recover(); r != nil {
            fmt.Printf("Caught: %v\n", r)
        }
    }()
    riskyOperation()
}
```

---

## 5. The `defer`/`panic`/`recover` Interaction

```
Function executes normally
    ↓
panic("oh no") is called
    ↓
Normal execution STOPS immediately
    ↓
Deferred functions run in LIFO order
    ↓
If a deferred function calls recover():
    → Panic is absorbed, function returns normally
If NO deferred function calls recover():
    → Panic propagates up the call stack
    → Eventually crashes the program with a stack trace
```

---

## 6. Patterns and Anti-Patterns

### ✅ Do: Convert panic to error
```go
func safeDivide(a, b float64) (result float64, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("panic: %v", r)
        }
    }()
    return a / b, nil
}
```

### ✅ Do: Recover at API boundaries
```go
// At the top of each HTTP handler or goroutine
defer func() {
    if r := recover(); r != nil {
        log.Printf("recovered: %v\nstack: %s", r, debug.Stack())
    }
}()
```

### ❌ Don't: Use panic for control flow
```go
// BAD — don't use panic like Python exceptions
func findItem(items []Item, id int) Item {
    for _, item := range items {
        if item.ID == id {
            return item
        }
    }
    panic("item not found")  // DON'T DO THIS — return (Item{}, error) instead
}
```

### ❌ Don't: Recover and silently ignore
```go
// BAD — swallowing panics hides real bugs
defer func() { recover() }()  // NEVER do this
```

---

## Run the Examples

```bash
go run example_defer.go
go run example_panic_recover.go
go run example_resource_management.go
go run example_real_world.go
```
