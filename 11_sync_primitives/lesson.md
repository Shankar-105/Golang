# Lesson 11: Sync Primitives — WaitGroup, Mutex, Once & Friends

> **Goal:** Learn Go's `sync` package — the low-level tools for coordinating goroutines when channels aren't the right fit.

---

## 11.1 Channels vs Sync Primitives — When to Use Which

Go says "share memory by communicating" (channels). But sometimes you DO need shared memory with locks. Here's the decision guide:

| Use **Channels** when... | Use **sync primitives** when... |
|---|---|
| Passing data between goroutines | Protecting access to shared state |
| Signaling events (done, quit) | Simple counter increment |
| Pipeline/producer-consumer | Cache with concurrent reads |
| Orchestrating goroutine workflow | One-time initialization |

**Analogy:** Channels are like sending Amazon packages (data moves between goroutines). Mutexes are like bathroom door locks (only one person uses the shared resource at a time).

Python equivalents:
| Go | Python |
|-----|--------|
| `sync.WaitGroup` | `asyncio.gather()` or `threading.Barrier` |
| `sync.Mutex` | `threading.Lock()` or `asyncio.Lock()` |
| `sync.RWMutex` | `threading.RLock()` (sort of) |
| `sync.Once` | Module-level initialization or `functools.lru_cache` trick |
| `sync.Map` | `dict` (Python's GIL makes dicts thread-safe for basic ops) |
| `sync/atomic` | No direct equivalent (Python's GIL handles it) |

---

## 11.2 `sync.WaitGroup` — Waiting for N Goroutines

You've already seen this, but let's go deep:

```go
var wg sync.WaitGroup

wg.Add(3)    // tell WaitGroup: 3 goroutines are coming
go func() { defer wg.Done(); /* work */ }()  // Done() decrements by 1
go func() { defer wg.Done(); /* work */ }()
go func() { defer wg.Done(); /* work */ }()
wg.Wait()    // blocks until counter reaches 0
```

**How it works internally:**
- `WaitGroup` has an internal atomic counter.
- `Add(n)` increases it by `n`.
- `Done()` decreases it by 1 (it's literally `Add(-1)`).
- `Wait()` blocks until the counter is 0.

**Critical rules:**
1. Call `Add()` **BEFORE** launching the goroutine (in the launching goroutine), not inside.
2. Call `Done()` with `defer` to ensure it runs even if the goroutine panics.
3. Don't copy a WaitGroup after first use — pass it by pointer.

```go
// ❌ BAD: Add inside goroutine — race condition!
for i := 0; i < 5; i++ {
    go func() {
        wg.Add(1)    // might run AFTER wg.Wait() checks!
        defer wg.Done()
    }()
}
wg.Wait()

// ✅ GOOD: Add before goroutine launch
for i := 0; i < 5; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
    }()
}
wg.Wait()
```

Python equivalent:
```python
# asyncio.gather is the closest equivalent
tasks = [asyncio.create_task(work()) for _ in range(5)]
await asyncio.gather(*tasks)
```

> See: [example_waitgroup.go](example_waitgroup.go)

---

## 11.3 `sync.Mutex` — Mutual Exclusion Lock

When multiple goroutines read/write the same variable, you get a **race condition**. A mutex ensures only one goroutine accesses the shared state at a time.

### The Problem
```go
counter := 0
for i := 0; i < 1000; i++ {
    go func() {
        counter++  // DATA RACE! Multiple goroutines read-modify-write simultaneously
    }()
}
// counter is NOT 1000. It's some random smaller number.
```

**Why it's broken:** `counter++` is actually THREE operations:
1. Read counter (value = 5)
2. Add 1 (value = 6) 
3. Write counter (store 6)

If two goroutines do this simultaneously:
```
Goroutine A: read counter → 5
Goroutine B: read counter → 5   (before A writes!)
Goroutine A: write counter ← 6
Goroutine B: write counter ← 6  (overwrites A's work!)
```

Result: Both incremented, but counter only went from 5 to 6. One increment was lost.

### The Fix: Mutex

```go
var mu sync.Mutex
counter := 0

for i := 0; i < 1000; i++ {
    go func() {
        mu.Lock()    // acquire lock — blocks if another goroutine holds it
        counter++     // safe: only one goroutine here at a time
        mu.Unlock()  // release lock — next goroutine can proceed
    }()
}
```

**Always use `defer mu.Unlock()`** to prevent forgetting to unlock:
```go
mu.Lock()
defer mu.Unlock()
// safe code here — even if we panic, unlock runs
```

Python equivalent:
```python
import threading
lock = threading.Lock()

with lock:  # acquire + release automatically
    counter += 1

# Or with asyncio:
async with asyncio.Lock():
    counter += 1
```

Go doesn't have `with` statements for locks, but `defer mu.Unlock()` is the idiomatic equivalent.

> See: [example_mutex.go](example_mutex.go)

---

## 11.4 `sync.RWMutex` — Read-Write Lock

Sometimes many goroutines read shared data, but few write. A regular Mutex blocks ALL access during reads — wasteful. `RWMutex` allows:

- **Multiple concurrent readers** (`RLock`/`RUnlock`)
- **Single exclusive writer** (`Lock`/`Unlock`)

```go
var rwmu sync.RWMutex
cache := make(map[string]string)

// Reader — many can run simultaneously
func get(key string) string {
    rwmu.RLock()              // shared lock — other readers OK
    defer rwmu.RUnlock()
    return cache[key]
}

// Writer — exclusive access, blocks all readers AND writers
func set(key, value string) {
    rwmu.Lock()               // exclusive lock
    defer rwmu.Unlock()
    cache[key] = value
}
```

**When to use RWMutex:** When reads vastly outnumber writes. Example: a config cache read by 100 goroutines but updated once a minute. Using Mutex would make all reads sequential. RWMutex lets them all read simultaneously.

Python's `threading.RLock()` is a re-entrant lock (different concept). Python doesn't have a built-in RWLock, though there are third-party implementations.

> See: [example_rwmutex.go](example_rwmutex.go)

---

## 11.5 `sync.Once` — Run Something Exactly Once

Perfect for lazy initialization, singleton patterns, or one-time setup:

```go
var once sync.Once
var db *Database

func getDB() *Database {
    once.Do(func() {
        // This function runs EXACTLY once, even if 100 goroutines call getDB() simultaneously
        fmt.Println("Initializing database connection...")
        db = connectToDatabase()
    })
    return db
}
```

**How it works:** `once.Do(f)` calls `f` the first time. All subsequent calls are no-ops. If multiple goroutines call `once.Do` simultaneously, only one executes `f`, and the others **wait** until `f` completes. This is thread-safe.

Python equivalent:
```python
# Module-level initialization (runs once on import)
db = connect_to_database()

# Or using a class:
class Singleton:
    _instance = None
    @classmethod
    def get(cls):
        if cls._instance is None:
            cls._instance = connect_to_database()
        return cls._instance
# But the Python version is NOT thread-safe without a lock!
```

> See: [example_once.go](example_once.go)

---

## 11.6 `sync/atomic` — Lock-Free Atomic Operations

For simple counters and flags, atomic operations are faster than mutexes:

```go
import "sync/atomic"

var counter int64

func increment() {
    atomic.AddInt64(&counter, 1) // atomic — no lock needed
}

func getCount() int64 {
    return atomic.LoadInt64(&counter) // atomic read
}
```

**Why atomic is faster:** Mutexes involve goroutine blocking and context switches. Atomic operations use CPU hardware instructions (like `LOCK CMPXCHG` on x86) that complete in a single, indivisible step.

**When to use atomic:** Only for simple operations on single variables (increment, swap, compare-and-swap). For anything complex, use Mutex.

**Python doesn't need this** because the GIL makes simple operations like `counter += 1` atomic on CPython (for built-in types). But this is a CPython implementation detail, not a language guarantee.

> See: [example_atomic.go](example_atomic.go)

---

## 11.7 Protecting a Struct (Thread-Safe Type Pattern)

The idiomatic Go pattern for a thread-safe type: embed a mutex in the struct.

```go
type SafeCounter struct {
    mu    sync.Mutex
    count map[string]int
}

func NewSafeCounter() *SafeCounter {
    return &SafeCounter{count: make(map[string]int)}
}

func (c *SafeCounter) Increment(key string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count[key]++
}

func (c *SafeCounter) Get(key string) int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.count[key]
}
```

Python equivalent:
```python
class SafeCounter:
    def __init__(self):
        self._lock = threading.Lock()
        self._count = {}

    def increment(self, key):
        with self._lock:
            self._count[key] = self._count.get(key, 0) + 1
```

---

## Exercises

### Exercise 1: Fix the Race
This code has a race condition. Fix it using (a) Mutex, (b) atomic, and (c) channels. Measure which is fastest with `time.Since`.
```go
counter := 0
var wg sync.WaitGroup
for i := 0; i < 100000; i++ {
    wg.Add(1)
    go func() { defer wg.Done(); counter++ }()
}
wg.Wait()
fmt.Println(counter) // not 100000!
```

### Exercise 2: Thread-Safe Cache
Build a `Cache` struct with `Get(key)`, `Set(key, value)`, and `Delete(key)` methods. Use `sync.RWMutex` for optimal read performance.

### Exercise 3: Once-Only Logger
Create a logger that initializes its output file exactly once, no matter how many goroutines call `Log()` simultaneously. Use `sync.Once`.

---

> **Next → [Lesson 12: Concurrency Patterns](../12_concurrency_patterns/lesson.md)** — Fan-in, fan-out, worker pools, pipelines.
