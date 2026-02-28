# Lesson 2: Error Handling — No Exceptions!

> **Goal:** Understand Go's explicit error handling. No `try/except`, no stack traces magically bubbling up. In Go, errors are just **values** you return and check.

---

## 2.1 The Big Difference

| Python | Go |
|--------|-----|
| `try: ... except ValueError as e:` | `if err != nil { ... }` |
| Exceptions fly up the call stack automatically | Errors are **returned**, not thrown |
| `raise ValueError("bad input")` | `return fmt.Errorf("bad input")` |
| You can ignore errors (no one forces you to `try`) | Go **forces** you to handle the return value (linter warns) |
| `except Exception` catches everything | Check `errors.Is()` / `errors.As()` for specific errors |

**Key insight:** In Python, errors are **exceptional events** that interrupt flow. In Go, errors are **ordinary return values** — just another thing a function gives back.

```python
# Python: errors interrupt flow (exceptions)
try:
    result = divide(10, 0)
except ZeroDivisionError as e:
    print(f"Error: {e}")
```

```go
// Go: errors are return values (explicit)
result, err := divide(10, 0)
if err != nil {
    fmt.Println("Error:", err)
}
```

---

## 2.2 The `error` Interface

In Go, `error` is a built-in interface with one method:

```go
type error interface {
    Error() string
}
```

Any type with an `Error() string` method is an error. That's it. No magic, no inheritance.

### Creating Errors

```go
import (
    "errors"
    "fmt"
)

// Method 1: errors.New — simple static error
err := errors.New("something went wrong")

// Method 2: fmt.Errorf — formatted error message (like f-strings)
name := "config.yaml"
err := fmt.Errorf("file not found: %s", name)
```

Python equivalent:
```python
# Python
raise ValueError("something went wrong")
raise FileNotFoundError(f"file not found: {name}")
```

---

## 2.3 The `if err != nil` Pattern

This is the most common Go pattern. You'll write it hundreds of times:

```go
func readConfig(path string) (Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return Config{}, err  // pass error up
    }

    var cfg Config
    err = json.Unmarshal(data, &cfg)
    if err != nil {
        return Config{}, err  // pass error up
    }

    return cfg, nil  // nil means no error (success!)
}
```

**Pattern:** Call a function → check `err != nil` → handle or return → continue if nil.

Python equivalent:
```python
def read_config(path: str) -> Config:
    try:
        data = open(path).read()  # raises FileNotFoundError
    except FileNotFoundError:
        raise  # or handle

    try:
        cfg = json.loads(data)  # raises JSONDecodeError
    except json.JSONDecodeError:
        raise

    return cfg
```

**Why Go developers prefer this:** Every error is handled at the point it occurs. No hidden control flow. No wondering "will this line throw?" You always know.

**Why newcomers hate it:** Verbose! Yes. But explicit beats implicit when debugging production servers at 3 AM.

> See: [example_basic_errors.go](example_basic_errors.go)

---

## 2.4 Wrapping Errors — Adding Context

A bare error like `"file not found"` isn't helpful. Which file? In which function? **Wrapping** adds context:

```go
func loadUserProfile(userID int) (*Profile, error) {
    path := fmt.Sprintf("/data/users/%d.json", userID)
    data, err := os.ReadFile(path)
    if err != nil {
        // Wrap: add context while preserving original error
        return nil, fmt.Errorf("loadUserProfile(%d): %w", userID, err)
    }
    // ...
}
```

The `%w` verb in `fmt.Errorf` **wraps** the original error inside a new error with more context.

Result: `"loadUserProfile(42): open /data/users/42.json: no such file or directory"`

Python equivalent:
```python
try:
    data = open(path).read()
except FileNotFoundError as e:
    raise RuntimeError(f"loadUserProfile({user_id})") from e
    # The `from e` chains the original exception
```

---

## 2.5 Checking Error Types: `errors.Is` and `errors.As`

### `errors.Is` — Check if an error IS a specific error (value comparison)

```go
import (
    "errors"
    "os"
)

_, err := os.Open("nonexistent.txt")
if errors.Is(err, os.ErrNotExist) {
    fmt.Println("File doesn't exist!")
}
```

Python equivalent:
```python
try:
    open("nonexistent.txt")
except FileNotFoundError:  # type matching
    print("File doesn't exist!")
```

### `errors.As` — Check if an error is a specific TYPE (and extract it)

```go
var pathErr *os.PathError
if errors.As(err, &pathErr) {
    fmt.Println("Failed path:", pathErr.Path)
    fmt.Println("Operation:", pathErr.Op)
}
```

Python equivalent:
```python
try:
    ...
except OSError as e:
    print(f"errno: {e.errno}, filename: {e.filename}")
```

**Why not just `==`?** Because wrapped errors contain the original error inside. `errors.Is` unwraps the chain to find a match. Same for `errors.As`.

> See: [example_wrapping.go](example_wrapping.go)

---

## 2.6 Custom Error Types

Since `error` is just an interface, you can make your own:

```go
type ValidationError struct {
    Field   string
    Message string
}

// Implement the error interface
func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error: %s — %s", e.Field, e.Message)
}

// Use it:
func validateAge(age int) error {
    if age < 0 {
        return &ValidationError{Field: "age", Message: "must be non-negative"}
    }
    return nil
}
```

Python equivalent:
```python
class ValidationError(Exception):
    def __init__(self, field: str, message: str):
        self.field = field
        self.message = message
        super().__init__(f"validation error: {field} — {message}")
```

> See: [example_custom_errors.go](example_custom_errors.go)

---

## 2.7 Sentinel Errors — Package-Level Error Constants

Common pattern: define known errors at the package level so callers can check for them.

```go
var (
    ErrNotFound     = errors.New("not found")
    ErrUnauthorized = errors.New("unauthorized")
    ErrInvalidInput = errors.New("invalid input")
)

func getUser(id int) (*User, error) {
    if id <= 0 {
        return nil, ErrInvalidInput
    }
    // ... lookup ...
    return nil, ErrNotFound
}

// Caller:
_, err := getUser(-1)
if errors.Is(err, ErrInvalidInput) {
    fmt.Println("Bad input!")
}
```

Python equivalent:
```python
# Python uses exception subclasses
class NotFoundError(Exception): pass
class UnauthorizedError(Exception): pass

try:
    get_user(-1)
except NotFoundError:
    print("Not found!")
```

---

## 2.8 The `errors` Package Cheat Sheet

| Function | Purpose | Python Equivalent |
|----------|---------|------------------|
| `errors.New("msg")` | Create simple error | `raise Exception("msg")` |
| `fmt.Errorf("context: %w", err)` | Wrap error with context | `raise X() from err` |
| `errors.Is(err, target)` | Check if error matches a value | `except SpecificError:` |
| `errors.As(err, &target)` | Extract typed error | `except OSError as e:` |
| `errors.Unwrap(err)` | Get the wrapped inner error | `e.__cause__` |

---

## Exercises

### Exercise 1: Divide Function
Write `func divide(a, b float64) (float64, error)` that returns an error for division by zero. The caller should check and handle the error.

### Exercise 2: Wrap It
Write a function chain: `readFile → parseJSON → validateConfig`. Each wraps the error from the previous step. Print the final error and verify `errors.Is` can find the original.

### Exercise 3: Custom Error
Create an `HTTPError` struct with `StatusCode int` and `Message string`. Use `errors.As` to extract the status code in the caller.

---

> **Next → [Lesson 3: Pointers](../03_pointers/lesson.md)** — What Python hides from you about memory.
