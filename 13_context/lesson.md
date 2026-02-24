# Lesson 13: Context & Cancellation — Timeouts, Deadlines, and Propagation

> **Goal:** Master `context.Context`, Go's standard mechanism for cancellation, timeouts, and request-scoped values. Every production Go server uses this.

---

## 13.1 The Problem Context Solves

Imagine an HTTP handler that:
1. Queries a database
2. Calls an external API
3. Processes results

The client disconnects after 2 seconds. Without context, all three operations keep running — wasting resources on work nobody will read.

**`context.Context` is Go's way to say "stop everything related to this request."**

Python equivalent: `asyncio.Task.cancel()` cancels a single task. Context cancels an **entire tree** of operations.

```python
# Python: Cancel one task
task = asyncio.create_task(fetch())
task.cancel()

# Go: Cancel an entire operation tree
ctx, cancel := context.WithCancel(context.Background())
// All goroutines checking ctx.Done() will stop
cancel()
```

---

## 13.2 The `context.Context` Interface

```go
type Context interface {
    Deadline() (deadline time.Time, ok bool) // when will this context expire?
    Done() <-chan struct{}                    // closed when context is cancelled
    Err() error                              // why was it cancelled?
    Value(key any) any                       // request-scoped data
}
```

**Key methods:**
- `ctx.Done()` — returns a channel that closes when the context is cancelled. Use in `select`.
- `ctx.Err()` — returns `context.Canceled` or `context.DeadlineExceeded` after cancellation.

---

## 13.3 Creating Contexts

### Root Contexts
```go
ctx := context.Background() // root context for main, init, tests
ctx := context.TODO()       // placeholder — "I know I need a context but haven't decided which"
```

### Derived Contexts (the important ones)

#### `WithCancel` — Manual cancellation
```go
ctx, cancel := context.WithCancel(parentCtx)
defer cancel() // ALWAYS defer cancel to free resources

go doWork(ctx)

// Later: cancel all work
cancel()
```

#### `WithTimeout` — Auto-cancel after duration
```go
ctx, cancel := context.WithTimeout(parentCtx, 5*time.Second)
defer cancel()

// ctx.Done() fires after 5 seconds
result, err := fetchWithContext(ctx)
```

#### `WithDeadline` — Auto-cancel at specific time
```go
deadline := time.Now().Add(30 * time.Second)
ctx, cancel := context.WithDeadline(parentCtx, deadline)
defer cancel()
```

#### `WithValue` — Attach request-scoped data
```go
ctx = context.WithValue(parentCtx, "userID", 42)

// Later, in a deep function:
userID := ctx.Value("userID").(int) // 42
```

**⚠️ Warning:** `WithValue` is for request-scoped data (trace IDs, auth tokens), NOT for passing function arguments. Don't abuse it.

---

## 13.4 Context Cancellation Tree

Contexts form a tree. Cancelling a parent cancels ALL children:

```
Background()
    └── WithCancel()        ← cancel this...
        ├── WithTimeout()   ← ...cancels this
        │   └── goroutine   ← ...and this
        └── goroutine       ← ...and this
```

This is incredibly powerful for HTTP servers: when a client disconnects, the server cancels the request context, which automatically cancels all database queries, API calls, and goroutines spawned by that request.

---

## 13.5 Using Context in Your Code

### Pattern 1: Check `ctx.Done()` in a loop
```go
func processItems(ctx context.Context, items []int) error {
    for _, item := range items {
        select {
        case <-ctx.Done():
            return ctx.Err() // cancelled or timed out
        default:
            // process item
            process(item)
        }
    }
    return nil
}
```

### Pattern 2: Pass context to blocking operations
```go
func fetchData(ctx context.Context, url string) ([]byte, error) {
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err // returns error if context cancelled
    }
    defer resp.Body.Close()
    return io.ReadAll(resp.Body)
}
```

### Pattern 3: Context in a select with channels
```go
func worker(ctx context.Context, jobs <-chan Job) {
    for {
        select {
        case <-ctx.Done():
            fmt.Println("Worker cancelled:", ctx.Err())
            return
        case job := <-jobs:
            process(job)
        }
    }
}
```

Python equivalent:
```python
async def worker(tasks):
    try:
        while True:
            task = await asyncio.wait_for(queue.get(), timeout=5.0)
            await process(task)
    except asyncio.CancelledError:
        print("Worker cancelled")
        raise  # re-raise to propagate
```

---

## 13.6 `context.WithTimeout` — The Most Common Pattern

In production, almost every network call should have a timeout:

```go
func getUserProfile(userID int) (*Profile, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    // Database query with timeout
    row := db.QueryRowContext(ctx, "SELECT * FROM users WHERE id = ?", userID)

    // If the query takes more than 3 seconds, ctx.Done() fires,
    // and QueryRowContext returns context.DeadlineExceeded
    var p Profile
    if err := row.Scan(&p.Name, &p.Email); err != nil {
        return nil, err
    }
    return &p, nil
}
```

---

## 13.7 Rules and Best Practices

1. **Context is the first parameter.** Always.
   ```go
   // ✅ GOOD
   func DoSomething(ctx context.Context, arg1 string) error

   // ❌ BAD
   func DoSomething(arg1 string, ctx context.Context) error
   ```

2. **Never store context in a struct.** Pass it explicitly.
   ```go
   // ❌ BAD
   type Server struct {
       ctx context.Context // NO!
   }

   // ✅ GOOD: pass to methods
   func (s *Server) Handle(ctx context.Context, req *Request) {}
   ```

3. **Always `defer cancel()`.** Even if the context times out on its own. The cancel function releases resources.

4. **Don't pass `nil` context.** Use `context.TODO()` if unsure.

5. **Context values are for request-scoped data only.** Trace IDs, auth tokens — not business logic.

---

## 13.8 Combining Context with Patterns

### Worker Pool + Context (graceful shutdown)
```go
func workerPool(ctx context.Context, jobs <-chan Job, numWorkers int) {
    var wg sync.WaitGroup
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for {
                select {
                case <-ctx.Done():
                    return // shutdown signal
                case job, ok := <-jobs:
                    if !ok {
                        return // channel closed
                    }
                    process(job)
                }
            }
        }()
    }
    wg.Wait()
}
```

### Pipeline + Context (cancellation propagation)
```go
func generate(ctx context.Context) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for i := 0; ; i++ {
            select {
            case <-ctx.Done():
                return
            case out <- i:
            }
        }
    }()
    return out
}
```

---

## Exercises

### Exercise 1: Timeout Fetch
Write a function `fetchWithTimeout(url string, timeout time.Duration) (string, error)` that uses `context.WithTimeout` to cancel an HTTP request if it takes too long.

### Exercise 2: Graceful Worker Shutdown
Create a worker pool of 5 workers. After 3 seconds, cancel the context and verify all workers stopped cleanly. Print how many jobs each worker completed.

### Exercise 3: Context Value Chain
Create a 3-level context chain: Background → WithValue("requestID", "abc-123") → WithTimeout(2s). Pass it through 3 functions that each read the requestID and check for cancellation.

---

> **Next → [Lesson 14: Race Conditions & Debugging](../14_race_conditions/lesson.md)** — Detecting and fixing concurrency bugs.
