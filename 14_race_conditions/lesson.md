# Lesson 14: Race Conditions, Deadlocks & Debugging Concurrency

> **Goal:** Learn to detect, diagnose, and fix the three major concurrency bugs: race conditions, deadlocks, and goroutine leaks. Master Go's built-in race detector and profiling tools.

---

## 14.1 The Three Concurrency Bugs

| Bug | What Happens | Symptoms |
|-----|-------------|----------|
| **Race condition** | Two goroutines access shared data, at least one writes | Wrong results, non-deterministic behavior |
| **Deadlock** | Goroutines wait for each other forever | Program hangs, no CPU usage |
| **Goroutine leak** | Goroutines never terminate | Memory grows forever, eventually OOM |

Python developers: the GIL protects you from most race conditions (for CPython built-in types). In Go, **there is no GIL**. You are responsible for synchronization.

---

## 14.2 Race Conditions — The Silent Killer

### What is a race condition?

A race condition occurs when two or more goroutines access shared memory concurrently, and at least one is writing.

```go
// ❌ RACE CONDITION
var counter int

func main() {
    for i := 0; i < 1000; i++ {
        go func() { counter++ }()  // read + modify + write — not atomic!
    }
    time.Sleep(time.Second)
    fmt.Println(counter) // different result every run!
}
```

**Why it's dangerous:** The program might work 99% of the time and fail in production under load. Race conditions are **non-deterministic** — they depend on goroutine scheduling timing.

### Go's Race Detector — Your Best Friend

Go has a built-in race detector. USE IT.

```bash
# Run with race detection
go run -race main.go

# Test with race detection
go test -race ./...

# Build with race detection (for staging)
go build -race -o myapp
```

Output when a race is detected:
```
WARNING: DATA RACE
Read at 0x00c0000b4010 by goroutine 8:
  main.main.func1()
      /path/main.go:12 +0x38

Previous write at 0x00c0000b4010 by goroutine 7:
  main.main.func1()
      /path/main.go:12 +0x4e

Goroutine 8 (running) created at:
  main.main()
      /path/main.go:11 +0x6c
```

The race detector tells you:
1. **What** was accessed concurrently (the memory address)
2. **Where** each goroutine accessed it (file + line number)
3. **When** the goroutines were created

**Performance impact:** Race detector adds ~2-10x slowdown. Use in development and CI, not production.

Python comparison: Python doesn't have a race detector because the GIL prevents most data races. Go doesn't have a GIL, so the race detector is essential.

> See: [example_race_detector.go](example_race_detector.go)

---

## 14.3 Common Race Condition Patterns

### Pattern 1: Unprotected shared variable
```go
// ❌ BAD
var cache map[string]string // maps are NOT goroutine-safe

go func() { cache["a"] = "1" }()
go func() { cache["b"] = "2" }()
// fatal error: concurrent map writes

// ✅ FIX: sync.Mutex
var mu sync.Mutex
mu.Lock(); cache["a"] = "1"; mu.Unlock()
```

### Pattern 2: Check-then-act (TOCTOU)
```go
// ❌ BAD: Time-of-check to time-of-use bug
if balance > amount {
    // Another goroutine could change balance HERE!
    balance -= amount
}

// ✅ FIX: Lock the entire check-then-act
mu.Lock()
defer mu.Unlock()
if balance > amount {
    balance -= amount
}
```

### Pattern 3: Closure variable capture
```go
// ❌ BAD: All goroutines share the same `i`
for i := 0; i < 10; i++ {
    go func() {
        fmt.Println(i) // prints 10 ten times!
    }()
}

// ✅ FIX: Pass as argument
for i := 0; i < 10; i++ {
    go func(n int) {
        fmt.Println(n) // prints 0-9
    }(i)
}
```

---

## 14.4 Deadlocks — The Freeze

### What is a deadlock?
Two or more goroutines are each waiting for something the other holds. Nobody can proceed.

### Go Runtime Detection
Go detects when ALL goroutines are blocked:
```
fatal error: all goroutines are asleep - deadlock!
```

But it ONLY detects when the **entire program** is stuck. If even one goroutine is running (like a timer or `main`'s `select{}`), Go won't report the deadlock even though other goroutines are stuck.

### Common Deadlock Patterns

#### Deadlock 1: Unbuffered channel, no reader
```go
ch := make(chan int) // unbuffered
ch <- 42            // DEADLOCK: blocks because nobody is reading
```

#### Deadlock 2: Lock ordering
```go
// Goroutine 1        // Goroutine 2
muA.Lock()           muB.Lock()
muB.Lock() // waits  muA.Lock() // waits
// Both wait forever!

// ✅ FIX: Always lock in the same order
// Both goroutines: muA first, then muB
```

#### Deadlock 3: Self-deadlock (non-reentrant mutex)
```go
mu.Lock()
mu.Lock() // DEADLOCK: Go's Mutex is NOT re-entrant
```

> See: [example_deadlocks.go](example_deadlocks.go)

---

## 14.5 Goroutine Leaks — The Memory Killer

A goroutine leak happens when a goroutine is started but never finishes. Each goroutine uses ~2-8KB of memory, so leaking thousands will eventually crash your program.

### Leak 1: Blocked channel send/receive
```go
func leakyFunc() {
    ch := make(chan int)
    go func() {
        val := <-ch // blocks forever — nobody sends to ch!
    }()
    // Function returns, but goroutine is stuck forever
}
```

### Leak 2: Missing done/cancel signal
```go
// ❌ Goroutine runs forever
go func() {
    for {
        doWork()
        time.Sleep(time.Second)
    }
}()

// ✅ FIX: Use context for cancellation
go func(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            doWork()
            time.Sleep(time.Second)
        }
    }
}(ctx)
```

### Detecting Goroutine Leaks
```go
import "runtime"

// Check goroutine count
before := runtime.NumGoroutine()
doSomething()
after := runtime.NumGoroutine()
if after > before {
    fmt.Printf("WARNING: %d goroutines leaked!\n", after-before)
}
```

> See: [example_goroutine_leaks.go](example_goroutine_leaks.go)

---

## 14.6 Debugging Tools

### Tool 1: `go run -race` (Race Detector)
```bash
go run -race main.go
go test -race -v ./...    # test with race detection
```

### Tool 2: `runtime` Package
```go
fmt.Println("Goroutines:", runtime.NumGoroutine())
fmt.Println("CPUs:", runtime.NumCPU())
runtime.GOMAXPROCS(4) // limit parallel threads
```

### Tool 3: `pprof` — Performance Profiler
```go
import _ "net/http/pprof"

go func() {
    http.ListenAndServe("localhost:6060", nil)
}()
```

Then visit:
- `http://localhost:6060/debug/pprof/goroutine` — goroutine dump
- `http://localhost:6060/debug/pprof/heap` — memory profile
- `http://localhost:6060/debug/pprof/profile?seconds=30` — CPU profile

```bash
# Analyze with go tool
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

### Tool 4: `SIGQUIT` Goroutine Dump
Press `Ctrl+\` (Unix) or send SIGQUIT to get a stack trace of all goroutines. Useful for diagnosing hung programs.

---

## 14.7 Concurrency Safety Checklist

Before shipping concurrent code, verify:

- [ ] **Every shared variable** is protected by a mutex or channel
- [ ] **Every goroutine can terminate** — has a cancellation path (context, done channel)
- [ ] **Channels are properly closed** — only by the sender, after all sends
- [ ] **`go run -race` passes** for all critical paths
- [ ] **No lock ordering bugs** — locks acquired in consistent order
- [ ] **`defer Unlock()`** used everywhere after `Lock()`
- [ ] **Goroutine count doesn't grow** — monitor with `runtime.NumGoroutine()`
- [ ] **Closure variables captured correctly** — passed as arguments or re-assigned

---

## 14.8 Python vs Go — Why Races Are Harder in Go

| | Python | Go |
|---|--------|-----|
| **GIL** | Prevents most data races for built-in types | No GIL — YOU must synchronize |
| **Default behavior** | Concurrent (one thread at a time) | **Parallel** (multiple goroutines simultaneously) |
| **Map/dict safety** | GIL makes `dict[k] = v` atomic | `map[k] = v` CRASHES with concurrent access |
| **Integer increment** | `x += 1` is atomic on CPython (but not guaranteed) | `x++` is NOT atomic — causes data race |
| **Detection** | Hard to trigger races due to GIL | `go run -race` catches races easily |
| **Common issue** | Mainly I/O races (async/file) | Data races everywhere if not careful |

---

## Exercises

### Exercise 1: Find the Bug
This code has a race condition. Run it with `-race` and fix it THREE different ways: (a) mutex, (b) channel, (c) atomic.
```go
var total int
var wg sync.WaitGroup

for i := 0; i < 100; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        for j := 0; j < 1000; j++ {
            total++
        }
    }()
}
wg.Wait()
fmt.Println(total) // should be 100,000
```

### Exercise 2: Diagnose the Deadlock
This program hangs. Find and fix the deadlock:
```go
ch := make(chan int) // unbuffered!
ch <- 1             // ???
fmt.Println(<-ch)
```

### Exercise 3: Fix the Goroutine Leak
This function leaks goroutines. Run it in a loop and watch `runtime.NumGoroutine()` grow. Fix it using context cancellation.
```go
func leaky() <-chan int {
    ch := make(chan int)
    go func() {
        for i := 0; ; i++ {
            ch <- i
        }
    }()
    return ch
}
```

### Exercise 4: Full Audit
Take any example from Lessons 8-12 and run it with `go run -race`. If any race is detected, fix it. Document what you found.

---

> **Phase 2 Complete!** You now understand goroutines, channels, select, sync primitives, concurrency patterns, context, and debugging.
> 
> **Next → Phase 3: [Lesson 15: Building a Concurrent HTTP Server](../15_http_server/lesson.md)** — Apply everything to build a real server.
