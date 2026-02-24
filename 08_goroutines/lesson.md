# Lesson 8: Goroutines Deep Dive — Writing Concurrent Code

> **Goal:** Go from theory to practice. Write goroutines, understand their lifecycle, avoid common traps, and see exactly how they differ from `asyncio.create_task()`.

---

## 8.1 The `go` Keyword — Launching a Goroutine

In Python asyncio:
```python
task = asyncio.create_task(my_coroutine())   # schedule it
await task                                    # wait for it
```

In Go:
```go
go myFunction()    // that's it. Fire and forget.
```

The `go` keyword:
1. Creates a new goroutine (allocates ~2 KB stack + a `G` struct).
2. Places it on the current P's run queue.
3. **Returns immediately** — the caller does NOT wait.

This is critical: `go` is fire-and-forget. If `main()` returns before the goroutine finishes, **the goroutine is killed**. There's no implicit "wait for all goroutines" (unlike Python's `asyncio.run()` which runs until the main coroutine completes).

> See: [example_basic_goroutine.go](example_basic_goroutine.go)

---

## 8.2 The "Main Exits Too Early" Trap

```go
func main() {
    go fmt.Println("hello from goroutine")
    // main() returns here — goroutine may not have run yet!
}
```

**Output: probably nothing!** The goroutine is scheduled but `main()` exits before it gets a chance to run. The Go runtime shuts down all goroutines when `main()` returns.

Python equivalent trap:
```python
import asyncio

async def main():
    asyncio.create_task(print("hello"))  # scheduled but not awaited
    # main() returns — the task might not run!

asyncio.run(main())  # might print nothing
```

**How to wait for goroutines:**

| Method | When to use |
|--------|-------------|
| `sync.WaitGroup` | When you need to wait for N goroutines to finish |
| Channels | When goroutines need to communicate results |
| `time.Sleep()` | **NEVER in production** — only for quick demos |
| `select {}` | Block forever (useful for long-running servers) |

> See: [example_main_exits.go](example_main_exits.go)

---

## 8.3 Anonymous Goroutines (Closures)

Just like Python lambdas and async function expressions, Go goroutines are often launched as anonymous functions:

```go
// Named function
go processRequest(req)

// Anonymous function (closure)
go func() {
    fmt.Println("I'm an anonymous goroutine!")
}()   // ← note the () — you must CALL the function

// Anonymous with parameters
go func(msg string) {
    fmt.Println(msg)
}("hello")   // ← passing "hello" as the argument
```

**Why pass arguments to anonymous goroutines?** Because of the **closure variable capture trap** (same as Python!):

```go
// ❌ BUG: all goroutines print the same value of i
for i := 0; i < 5; i++ {
    go func() {
        fmt.Println(i)  // captures `i` by reference — will be 5 by the time goroutines run
    }()
}
// Likely output: 5, 5, 5, 5, 5

// ✅ FIX: pass i as a parameter (captures by value)
for i := 0; i < 5; i++ {
    go func(n int) {
        fmt.Println(n)  // n is a copy of i at launch time
    }(i)
}
// Output: 0, 1, 2, 3, 4 (in random order — goroutines are concurrent!)
```

Python has the exact same bug:
```python
# ❌ BUG: same closure problem
for i in range(5):
    asyncio.create_task(lambda: print(i))  # all print 4

# ✅ FIX: default argument captures value
for i in range(5):
    asyncio.create_task(lambda n=i: print(n))
```

> See: [example_closure_trap.go](example_closure_trap.go)

---

## 8.4 Goroutine Lifecycle

```
Created ──▶ Runnable ──▶ Running ──▶ Finished
                │              │
                │              ▼
                │          Blocked (I/O, channel, mutex, sleep)
                │              │
                ◀──────────────┘
                 (unblocked → back to runnable)
```

States explained:
- **Created:** `go func()` was called. A `G` struct is allocated with a 2 KB stack.
- **Runnable:** On a P's run queue, waiting for a turn on an M (OS thread).
- **Running:** Currently executing on an M. Only `GOMAXPROCS` goroutines can be running simultaneously.
- **Blocked:** Waiting for I/O, channel operation, mutex, `time.Sleep`, or syscall. The M is freed to run other goroutines.
- **Finished:** Function returned. The `G` struct is recycled (not garbage collected — put on a free list for reuse).

**Key insight:** When a goroutine blocks (e.g., reading from a channel, doing network I/O), the **M (OS thread) is NOT blocked**. The scheduler detaches the goroutine from the M and assigns another runnable goroutine. This is why Go can handle 100K blocked goroutines on just 8 threads.

Exception: **blocking syscalls** (like file I/O on some systems) DO block the M. In that case, the runtime spins up a new M so the P still has a thread. When the syscall completes, the extra M goes to sleep.

---

## 8.5 `GOMAXPROCS` — Controlling Parallelism

```go
import "runtime"

// Get current GOMAXPROCS
fmt.Println(runtime.GOMAXPROCS(0))  // 0 means "just tell me, don't change"

// Set GOMAXPROCS
runtime.GOMAXPROCS(4)  // use 4 OS threads for goroutine scheduling

// Or via environment variable:
// GOMAXPROCS=4 go run main.go
```

**Default:** `runtime.NumCPU()` (number of logical CPU cores). On an 8-core machine, `GOMAXPROCS=8`.

**What it means:**
- `GOMAXPROCS=1`: Only ONE goroutine runs at a time (like asyncio — concurrent but not parallel).
- `GOMAXPROCS=8`: Up to 8 goroutines run simultaneously on 8 threads.
- Setting it higher than `NumCPU()` rarely helps — you'd exceed physical cores.

**Experiment:** Run CPU-heavy work with GOMAXPROCS=1 vs 8 and compare:

```go
runtime.GOMAXPROCS(1)  // single-threaded like asyncio
// vs
runtime.GOMAXPROCS(8)  // true parallelism
```

> See: [example_gomaxprocs.go](example_gomaxprocs.go)

---

## 8.6 Goroutines vs Python Comparison Table

| | Python `threading` | Python `asyncio` | Go goroutines |
|---|---|---|---|
| Launch syntax | `Thread(target=f).start()` | `asyncio.create_task(f())` | `go f()` |
| Wait for completion | `thread.join()` | `await task` | `sync.WaitGroup` or channels |
| Return values | side effects only (or queue) | `await` returns value | channels |
| Error handling | try/except in thread | try/except in coroutine | channel or error return |
| Preemptive? | Yes (OS), but GIL | No (cooperative) | Yes (since Go 1.14) |
| True parallelism? | No (GIL) | No (single thread) | **Yes** |
| Memory per unit | 1-8 MB | ~few KB | ~2 KB (grows) |
| Max practical count | ~1,000 | ~100,000 | ~1,000,000+ |
| Function coloring? | No | Yes (async/sync split) | **No** |

---

## 8.7 Yielding & Preemption

**asyncio (cooperative):**
```python
async def work():
    # Must explicitly yield with `await` 
    # Otherwise blocks the entire event loop
    await asyncio.sleep(0)  # yield point
```

**Go (preemptive since 1.14):**
Before Go 1.14, goroutines only yielded at "safe points" — function calls, channel ops, I/O. A tight CPU loop with no function calls could starve other goroutines.

Since Go 1.14, the runtime uses **asynchronous preemption**: it sends a signal (SIGURG on Unix) to the thread, which interrupts the tight loop and switches goroutines. This means even a `for { x++ }` loop won't starve others.

```go
// Before Go 1.14, this could starve other goroutines:
go func() {
    for {
        // tight loop, no function calls = no preemption point
        // OTHER goroutines couldn't run on this thread!
    }
}()

// Since Go 1.14: the runtime can preempt even this.
// The signal-based mechanism interrupts it periodically.
```

> See: [example_preemption.go](example_preemption.go)

---

## Exercises

### Exercise 1: Goroutine Launch Order
What does this print? Predict before running.
```go
func main() {
    for i := 0; i < 5; i++ {
        go func(n int) {
            fmt.Println(n)
        }(i)
    }
    time.Sleep(time.Second)
}
```
Is the order guaranteed? Why or why not?

### Exercise 2: Rewrite Python asyncio in Go
Convert this Python code to Go using goroutines:
```python
import asyncio

async def fetch(url: str) -> str:
    await asyncio.sleep(1)  # simulate network
    return f"data from {url}"

async def main():
    urls = ["url1", "url2", "url3", "url4", "url5"]
    tasks = [asyncio.create_task(fetch(u)) for u in urls]
    results = await asyncio.gather(*tasks)
    for r in results:
        print(r)

asyncio.run(main())
```

### Exercise 3: Million Goroutines
Write a program that launches 1,000,000 goroutines, each incrementing a shared counter. Use `sync.WaitGroup` to wait. Measure how long it takes. (Warning: the counter will be wrong due to race conditions — we'll fix this in Lesson 11!)

---

> **Next → [Lesson 9: Channels](../09_channels/lesson.md)** — How goroutines communicate safely.
