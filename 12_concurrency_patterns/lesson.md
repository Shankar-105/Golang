# Lesson 12: Concurrency Patterns — Worker Pools, Pipelines, Fan-In/Fan-Out

> **Goal:** Combine goroutines, channels, and sync primitives into real-world concurrency patterns. These are the blueprints you'll use daily when building backend systems.

---

## 12.1 Why Patterns Matter

You now know goroutines, channels, select, and mutexes. But knowing the bricks doesn't mean you can build a house. This lesson teaches the **architectural patterns** for concurrent Go programs.

Python equivalent: In asyncio, patterns like `asyncio.gather()`, `asyncio.Queue`, and `asyncio.Semaphore` serve similar roles. Go's approach is more explicit and compositional.

---

## 12.2 Pattern 1: Worker Pool (a.k.a. Thread Pool)

**Problem:** You have 10,000 tasks but want at most 10 running concurrently.

**Solution:** Launch N worker goroutines. Feed them tasks via a channel.

```
          ┌─── Worker 1 ───┐
          │                 │
Jobs ──→  ├─── Worker 2 ───┤  ──→ Results
Channel   │                 │      Channel
          ├─── Worker 3 ───┤
          │                 │
          └─── Worker N ───┘
```

```go
func worker(id int, jobs <-chan int, results chan<- int) {
    for job := range jobs {
        fmt.Printf("Worker %d processing job %d\n", id, job)
        time.Sleep(time.Second) // simulate work
        results <- job * 2
    }
}

func main() {
    jobs := make(chan int, 100)
    results := make(chan int, 100)

    // Launch 3 workers
    for w := 1; w <= 3; w++ {
        go worker(w, jobs, results)
    }

    // Send 9 jobs
    for j := 1; j <= 9; j++ {
        jobs <- j
    }
    close(jobs) // signal workers: no more jobs

    // Collect results
    for r := 1; r <= 9; r++ {
        fmt.Println(<-results)
    }
}
```

Python equivalent:
```python
import asyncio

async def worker(name, queue, results):
    while True:
        job = await queue.get()
        result = await process(job)
        results.append(result)
        queue.task_done()

async def main():
    queue = asyncio.Queue()
    results = []
    workers = [asyncio.create_task(worker(f"w{i}", queue, results)) for i in range(3)]
    for job in range(9):
        await queue.put(job)
    await queue.join()
    for w in workers:
        w.cancel()
```

**Key insight:** The worker pool pattern naturally limits concurrency. Only N goroutines run, even with millions of jobs. This prevents resource exhaustion (file handles, DB connections, memory).

> See: [example_worker_pool.go](example_worker_pool.go)

---

## 12.3 Pattern 2: Pipeline

**Problem:** Data passes through multiple processing stages in sequence.

**Solution:** Chain channels — each stage reads from one channel and writes to the next.

```
Source → [Stage 1] → ch1 → [Stage 2] → ch2 → [Stage 3] → Sink
```

```go
// Stage 1: Generate numbers
func generate(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        for _, n := range nums {
            out <- n
        }
        close(out)
    }()
    return out
}

// Stage 2: Square each number
func square(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        for n := range in {
            out <- n * n
        }
        close(out)
    }()
    return out
}

// Stage 3: Filter — keep only even numbers
func filterEven(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        for n := range in {
            if n%2 == 0 {
                out <- n
            }
        }
        close(out)
    }()
    return out
}

func main() {
    // Compose the pipeline
    ch := generate(1, 2, 3, 4, 5, 6)
    ch = square(ch)
    ch = filterEven(ch)

    for result := range ch {
        fmt.Println(result) // 4, 16, 36
    }
}
```

Python equivalent:
```python
# Generator pipeline
def generate(*nums):
    yield from nums

def square(source):
    for n in source:
        yield n * n

def filter_even(source):
    for n in source:
        if n % 2 == 0:
            yield n

for result in filter_even(square(generate(1, 2, 3, 4, 5, 6))):
    print(result)
```

**Key insight:** Each pipeline stage runs in its own goroutine. Data flows through channels. This is clean, composable, and concurrent — each stage processes independently.

> See: [example_pipeline.go](example_pipeline.go)

---

## 12.4 Pattern 3: Fan-Out / Fan-In

**Problem:** One stage is slow. You want to parallelize it.

**Fan-out:** Launch multiple goroutines reading from the same channel.
**Fan-in:** Merge multiple channels into one.

```
                    ┌──→ Worker A ──┐
Source ──→ ch ──→   ├──→ Worker B ──┤  ──→ merge() ──→ final channel
                    └──→ Worker C ──┘
```

```go
// Fan-in: merge multiple channels into one
func merge(channels ...<-chan int) <-chan int {
    var wg sync.WaitGroup
    merged := make(chan int)

    for _, ch := range channels {
        wg.Add(1)
        go func(c <-chan int) {
            defer wg.Done()
            for val := range c {
                merged <- val
            }
        }(ch)
    }

    go func() {
        wg.Wait()
        close(merged)
    }()

    return merged
}
```

> See: [example_fan_out_in.go](example_fan_out_in.go)

---

## 12.5 Pattern 4: Semaphore (Bounded Concurrency)

**Problem:** Limit concurrent access to a resource (e.g., max 5 HTTP requests at once).

**Solution:** Use a buffered channel as a semaphore.

```go
sem := make(chan struct{}, 5) // max 5 concurrent

for _, url := range urls {
    sem <- struct{}{} // acquire (blocks when 5 are running)
    go func(u string) {
        defer func() { <-sem }() // release
        fetch(u)
    }(url)
}
```

Python equivalent:
```python
sem = asyncio.Semaphore(5)

async def bounded_fetch(url):
    async with sem:
        return await fetch(url)
```

> See: [example_semaphore.go](example_semaphore.go)

---

## 12.6 Pattern 5: Or-Done Channel

**Problem:** You're ranging over a channel, but you also need to respect cancellation.

```go
// Without orDone — ignores cancellation!
for val := range ch {
    process(val)
}

// With orDone — stops when done is closed
func orDone(done <-chan struct{}, c <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for {
            select {
            case <-done:
                return
            case v, ok := <-c:
                if !ok {
                    return
                }
                select {
                case out <- v:
                case <-done:
                    return
                }
            }
        }
    }()
    return out
}
```

---

## 12.7 Pattern 6: Error Group (golang.org/x/sync/errgroup)

**Problem:** WaitGroup doesn't handle errors. You want to run N goroutines, wait for all, and collect the first error.

```go
import "golang.org/x/sync/errgroup"

g, ctx := errgroup.WithContext(context.Background())

for _, url := range urls {
    url := url // capture
    g.Go(func() error {
        return fetch(ctx, url)
    })
}

if err := g.Wait(); err != nil {
    log.Fatal(err) // first error from any goroutine
}
```

Python equivalent:
```python
results = await asyncio.gather(*tasks, return_exceptions=True)
errors = [r for r in results if isinstance(r, Exception)]
```

This is one of the most useful packages for production Go code. It's in `golang.org/x/sync` (the official extended library).

---

## 12.8 When to Use What

| Pattern | Use Case | Example |
|---------|----------|---------|
| Worker Pool | Bounded parallel processing | HTTP scraper with 10 workers |
| Pipeline | Sequential processing stages | ETL: extract → transform → load |
| Fan-Out/In | Parallelize one slow stage | Multiple DB queries merged |
| Semaphore | Limit concurrent resource use | Max 5 open files |
| Or-Done | Cancellable channel reads | Graceful shutdown |
| ErrGroup | Wait + first error | Parallel API calls |

---

## Exercises

### Exercise 1: Image Processing Pipeline
Build a 3-stage pipeline:
1. **Generate:** produce file paths
2. **Process:** simulate resizing (sleep 200ms)
3. **Save:** simulate writing (sleep 100ms)

Use fan-out on stage 2 (3 workers) since it's the bottleneck.

### Exercise 2: Rate-Limited Scraper
Build a worker pool that fetches URLs but limits to 5 concurrent requests. Use a buffered channel as a semaphore.

### Exercise 3: Error-Collecting Worker Pool
Modify the worker pool to collect errors. If more than 3 errors occur, stop processing remaining jobs.

---

> **Next → [Lesson 13: Context & Cancellation](../13_context/lesson.md)** — The canonical way to cancel, timeout, and propagate deadlines.
