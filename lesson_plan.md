# Go Mastery Lesson Plan — From Python to High-Throughput Go

> **Who this is for:** A 2nd-year CS student fluent in Python (asyncio, FastAPI, classes, decorators) who wants to master Go for building high-throughput backend systems.
>
> **End goal:** Build a mini concurrent web server in Go that handles HTTP requests with goroutines — the Go equivalent of your Python asyncio + FastAPI server, but with true parallelism.

---

## Phase 1: Go Through a Python Lens (Language Differences)

| # | Topic | Key Comparisons to Python | Files |
|---|-------|--------------------------|-------|
| 1 | **Syntax & Idioms That Differ** | Declarations, visibility, multiple return values, no classes, packages vs modules | `01_syntax_differences/` |
| 2 | **Error Handling: No Exceptions** | `error` interface vs `try/except/raise`, wrapping errors, `errors.Is/As` vs `isinstance` | `02_error_handling/` |
| 3 | **Pointers: What Python Hides From You** | Everything in Python is a reference; Go gives you the choice. Value vs pointer semantics. | `03_pointers/` |
| 4 | **Structs & Interfaces vs Classes** | No inheritance, composition over inheritance, implicit interfaces vs Python's ABC/duck typing | `04_structs_interfaces/` |
| 5 | **Slices, Maps & Strings** | Slices vs Python lists (capacity, append gotchas), maps vs dicts, strings as byte slices vs Python str | `05_collections/` |
| 6 | **Packages, Modules & Visibility** | `go mod` vs `pip/venv`, uppercase export vs `__all__`, package layout conventions | `06_packages/` |

---

## Phase 2: Concurrency — The Main Event

| # | Topic | Key Comparisons to Python | Files |
|---|-------|--------------------------|-------|
| 7 | **Why Goroutines Exist (First Principles)** | Python's GIL problem, asyncio's cooperative model, OS threads cost. Why Go invented goroutines. | `07_why_goroutines/` |
| 8 | **Goroutines Deep Dive** | M:N scheduler, goroutine stack growth, `go` keyword vs `asyncio.create_task()` vs `threading.Thread` | `08_goroutines/` |
| 9 | **Channels: Communicating Sequential Processes** | Channels vs `asyncio.Queue`, buffered vs unbuffered, directional channels | `09_channels/` |
| 10 | **Select Statement & Multiplexing** | `select` vs `asyncio.gather/wait`, timeout patterns, non-blocking receives | `10_select/` |
| 11 | **Sync Primitives** | `sync.WaitGroup` vs `asyncio.gather`, `sync.Mutex` vs `threading.Lock`, `sync.Once`, `sync.Map` | `11_sync_primitives/` |
| 12 | **Concurrency Patterns** | Fan-in, fan-out, worker pools, pipelines, semaphores — all compared to asyncio equivalents | `12_concurrency_patterns/` |
| 13 | **Context & Cancellation** | `context.Context` vs `asyncio.Task.cancel()`, timeouts, propagation through call chains | `13_context/` |
| 14 | **Edge Cases & Debugging** | Race conditions, deadlocks, starvation, `go run -race`, `pprof` profiling | `14_race_conditions/` |

---

## Phase 3: The Standard Library for Backend Work

| # | Topic | Key Comparisons to Python | Files |
|---|-------|--------------------------|-------|
| 15 | **`net/http` Basics** | Go's `http.ListenAndServe` vs Flask/FastAPI, handlers vs route decorators | `15_http_basics/` |
| 16 | **JSON, I/O & Encoding** | `encoding/json` vs `json` module, struct tags, `io.Reader/Writer` vs Python file objects | `16_json_io/` |
| 17 | **Testing in Go** | `go test` vs `pytest`, table-driven tests, benchmarks, test coverage | `17_testing/` |
| 18 | **`defer`, `panic`, `recover`** | `defer` vs `with`/context managers, `panic` vs raising exceptions, `recover` vs `except` | `18_defer_panic/` |

---

## Phase 4: Capstone — Mini Concurrent Web Server

| # | Topic | What You'll Build | Files |
|---|-------|-------------------|-------|
| 19 | **Design & Architecture** | Blueprint for a concurrent HTTP server: routing, middleware, connection handling | `19_server_design/` |
| 20 | **Implementation** | Full mini web server with goroutine-per-request, graceful shutdown, context propagation | `20_mini_server/` |
| 21 | **Load Testing & Comparison** | Benchmark your Go server vs your Python asyncio server — see the throughput difference | `21_benchmarks/` |

---

## How Each Lesson Works

```
📂 07_why_goroutines/
├── lesson.md          ← Read this first (theory, analogies, diagrams)
├── example_basic.go   ← Run: go run example_basic.go
├── example_race.go    ← Run: go run -race example_race.go
└── exercises.md       ← Your turn: challenges to solve
```

1. **Read** the `lesson.md` — concepts explained from first principles with Python comparisons.
2. **Run** the `.go` files — see the concepts in action.
3. **Do** the exercises — modify code, solve challenges, break things on purpose.
4. **Ask questions** — this is mentorship, not a lecture. Slow down anytime.

---

## Let's Begin → [Lesson 1: Syntax & Idioms That Differ](01_syntax_differences/lesson.md)
