# Lesson 17: Testing in Go

## Python → Go Mental Model

| Python (pytest) | Go (testing) |
|-----------------|-------------|
| `test_*.py` files | `*_test.go` files |
| `def test_something():` | `func TestSomething(t *testing.T)` |
| `assert x == y` | `if x != y { t.Errorf(...) }` |
| `@pytest.mark.parametrize` | Table-driven tests (slice of test cases) |
| `pytest.fixture` | `TestMain(m *testing.M)` or `t.Cleanup()` |
| `pytest.raises(ValueError)` | Check `err != nil` in the test |
| `unittest.mock` | Interfaces + dependency injection |
| `pytest --benchmark` | `func BenchmarkXxx(b *testing.B)` |
| `pytest -v` | `go test -v` |
| `pytest --cov` | `go test -cover` |
| `httptest` (Flask test client) | `httptest.NewRecorder()` |

---

## 1. Testing Fundamentals

### 1.1 The `testing` Package

Go has testing built in — no third-party framework needed (unlike Python where you install pytest).

**Rules:**
1. Test files end in `_test.go`
2. Test functions start with `Test` and take `*testing.T`
3. Run with `go test ./...`

```go
// math_test.go
package mathutil

import "testing"

func TestAdd(t *testing.T) {
    got := Add(2, 3)
    want := 5
    if got != want {
        t.Errorf("Add(2, 3) = %d, want %d", got, want)
    }
}
```

### 1.2 No `assert` — Use `if` Statements

Go doesn't have a built-in `assert`. You use `if` + error reporting:

```go
// t.Error / t.Errorf  — report failure, CONTINUE running the test
// t.Fatal / t.Fatalf  — report failure, STOP this test immediately
// t.Log / t.Logf      — log info (shown only with -v flag)
// t.Skip / t.Skipf    — skip this test

func TestDivide(t *testing.T) {
    result, err := Divide(10, 2)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)  // Fatal = stop
    }
    if result != 5 {
        t.Errorf("Divide(10,2) = %d, want 5", result)  // Error = continue
    }
}
```

**Why no assert?**  
Go's philosophy: explicit is better. The `if` pattern gives you full control over the error message, which makes test failures much easier to debug.

### 1.3 Python Comparison

```python
# Python (pytest)
def test_add():
    assert add(2, 3) == 5

def test_divide():
    assert divide(10, 2) == 5.0
    with pytest.raises(ValueError):
        divide(10, 0)
```

```go
// Go
func TestAdd(t *testing.T) {
    if got := Add(2, 3); got != 5 {
        t.Errorf("Add(2, 3) = %d, want 5", got)
    }
}

func TestDivide(t *testing.T) {
    _, err := Divide(10, 0)
    if err == nil {
        t.Fatal("expected error for division by zero")
    }
}
```

---

## 2. Table-Driven Tests

The **#1 Go testing pattern**. Instead of writing many similar test functions, define a table of inputs and expected outputs.

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {"positive", 2, 3, 5},
        {"zero", 0, 0, 0},
        {"negative", -1, -2, -3},
        {"mixed", -5, 10, 5},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Add(tt.a, tt.b)
            if got != tt.want {
                t.Errorf("Add(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
            }
        })
    }
}
```

**Python equivalent:**
```python
@pytest.mark.parametrize("a,b,want", [
    (2, 3, 5),
    (0, 0, 0),
    (-1, -2, -3),
    (-5, 10, 5),
])
def test_add(a, b, want):
    assert add(a, b) == want
```

### Why `t.Run()`?

- Gives each sub-test a **name** (visible in output)
- Sub-tests can be run individually: `go test -run TestAdd/negative`
- Sub-tests can run in **parallel** with `t.Parallel()`

---

## 3. Testing Errors

```go
func TestDivideErrors(t *testing.T) {
    tests := []struct {
        name    string
        a, b    float64
        wantErr bool
    }{
        {"valid", 10, 2, false},
        {"divide by zero", 10, 0, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := Divide(tt.a, tt.b)
            if (err != nil) != tt.wantErr {
                t.Errorf("Divide(%v, %v) error = %v, wantErr %v",
                    tt.a, tt.b, err, tt.wantErr)
            }
        })
    }
}
```

---

## 4. Test Helpers

### 4.1 `t.Helper()`

Mark a function as a test helper so error line numbers point to the caller:

```go
func assertEqual(t *testing.T, got, want int) {
    t.Helper()  // without this, errors point here instead of the test
    if got != want {
        t.Errorf("got %d, want %d", got, want)
    }
}

func TestSomething(t *testing.T) {
    assertEqual(t, Add(1, 2), 3)  // error points HERE, not inside assertEqual
}
```

### 4.2 `t.Cleanup()`

Register cleanup functions (like pytest fixtures with teardown):

```go
func TestWithDB(t *testing.T) {
    db := setupTestDB()
    t.Cleanup(func() {
        db.Close()  // runs after test completes
    })
    // ... test using db ...
}
```

### 4.3 `t.TempDir()`

Get a temp directory that's auto-cleaned:

```go
func TestFileProcessing(t *testing.T) {
    dir := t.TempDir()  // automatically removed after test
    path := filepath.Join(dir, "test.txt")
    os.WriteFile(path, []byte("test data"), 0644)
    // ... test file processing ...
}
```

---

## 5. Parallel Tests

```go
func TestParallel(t *testing.T) {
    tests := []struct{ name string; input int }{
        {"case1", 1}, {"case2", 2}, {"case3", 3},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()  // this sub-test runs concurrently
            // NOTE: capture tt in the closure — Go's range variable trap!
            result := SlowFunction(tt.input)
            if result != tt.input*2 {
                t.Errorf("got %d, want %d", result, tt.input*2)
            }
        })
    }
}
```

---

## 6. HTTP Handler Testing with `httptest`

### 6.1 `httptest.NewRecorder()`

Test HTTP handlers without starting a real server:

```go
func TestHealthHandler(t *testing.T) {
    req := httptest.NewRequest("GET", "/health", nil)
    w := httptest.NewRecorder()

    HealthHandler(w, req)

    resp := w.Result()
    if resp.StatusCode != http.StatusOK {
        t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
    }

    body, _ := io.ReadAll(resp.Body)
    if string(body) != `{"status":"ok"}` {
        t.Errorf("body = %s", string(body))
    }
}
```

### 6.2 `httptest.NewServer()`

Spin up a real test server (useful for integration tests):

```go
func TestAPIIntegration(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(MyHandler))
    defer srv.Close()

    resp, err := http.Get(srv.URL + "/api/users")
    // ... assert response ...
}
```

---

## 7. Benchmarks

```go
func BenchmarkAdd(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Add(2, 3)
    }
}
```

Run with: `go test -bench=. -benchmem`

Output:
```
BenchmarkAdd-8    1000000000    0.25 ns/op    0 B/op    0 allocs/op
```

---

## 8. Test Coverage

```bash
go test -cover ./...                      # show coverage %
go test -coverprofile=coverage.out ./...  # generate coverage file
go tool cover -html=coverage.out          # open HTML report in browser
go tool cover -func=coverage.out          # per-function coverage
```

---

## 9. `TestMain` — Setup/Teardown for Entire Package

```go
func TestMain(m *testing.M) {
    // Setup (runs before ALL tests in package)
    db := setupDatabase()

    // Run all tests
    code := m.Run()

    // Teardown (runs after ALL tests)
    db.Close()

    os.Exit(code)
}
```

Python equivalent:
```python
@pytest.fixture(scope="session")
def db():
    db = setup_database()
    yield db
    db.close()
```

---

## Running the Tests

```bash
# Run all tests in the stringutil package
go test ./17_testing/stringutil/ -v

# Run a specific test
go test ./17_testing/stringutil/ -run TestReverse

# Run benchmarks
go test ./17_testing/stringutil/ -bench=. -benchmem

# Test with coverage
go test ./17_testing/stringutil/ -cover

# Run tests with race detector
go test ./17_testing/stringutil/ -race
```
