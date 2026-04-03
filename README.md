# Learn Go for Real-World Backend Concurrency

I am a second-year CS student, and I built this repo as the learning path I wish I had while learning Go for backend systems.

This repo is for everyone who wants to learn Go, from complete beginners to people coming from other languages. A major focus is comparing Go concurrency with Python Event loop mechanism, Why Python Event Loop comparsion only? as i use to deal with python asyncio in my previous projects i used it as a comparsion nothin more than that, so Python knowledge is not required.

## Why I built this

I wanted one place where each topic answers:

1. What is the Go concept?
2. What is the Python asyncio, or Node.js Event Loop equivalent?
3. Where do they differ in real backend systems?

Go has become a strong backend and infra language because it gives very high throughput with a simple runtime model, great tooling, and predictable deployment for services, APIs, and distributed systems.

## Who this is for

- Anyone learning Go from scratch
- Students learning backend engineering fundamentals
- Developers from any language who want practical Go patterns
- Python developers who want clear asyncio-to-Go comparisons

## Learning Path

### 1) Foundations through Python comparisons

- [01_syntax_differences](01_syntax_differences)
- [02_error_handling](02_error_handling)
- [03_pointers](03_pointers)
- [04_structs_interfaces](04_structs_interfaces)
- [05_collections](05_collections)
- [06_packages](06_packages)

### 2) Concurrency core

- [07_why_goroutines](07_why_goroutines)
- [08_goroutines](08_goroutines)
- [09_channels](09_channels)
- [10_select](10_select)
- [11_sync_primitives](11_sync_primitives)
- [12_concurrency_patterns](12_concurrency_patterns)
- [13_context](13_context)
- [14_race_conditions](14_race_conditions)

### 3) Backend practicals

- [15_http_basics](15_http_basics)
- [16_json_io](16_json_io)
- [17_testing](17_testing)
- [18_defer_panic](18_defer_panic)

## How to use this repo

For each module:

1. Read lesson.md first.
2. Run the examples with go run.
3. Break and modify the code to learn behavior.
4. Re-run with race detector where relevant.

## Contributing

If you find mistakes, missing explanations, or better Python-to-Go analogies, or any other Specific Node.js or Java Vthreads to Go comparsions please do open an issue or PR.
