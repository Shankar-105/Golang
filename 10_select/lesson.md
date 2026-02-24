# Lesson 10: The `select` Statement — Multiplexing Channels

> **Goal:** Master `select` — Go's way to wait on multiple channel operations simultaneously. Think of it as a `switch` statement for channels, or Python's `asyncio.wait()` on steroids.

---

## 10.1 Why `select` Exists

Imagine your goroutine needs data from multiple sources:
- A channel with database results
- A channel with cache results  
- A timeout signal

In Python asyncio:
```python
done, pending = await asyncio.wait(
    [db_task, cache_task],
    timeout=5.0,
    return_when=asyncio.FIRST_COMPLETED
)
```

In Go:
```go
select {
case result := <-dbChan:
    fmt.Println("DB:", result)
case result := <-cacheChan:
    fmt.Println("Cache:", result)
case <-time.After(5 * time.Second):
    fmt.Println("Timeout!")
}
```

**`select` blocks until ONE of its cases can proceed, then executes that case.** If multiple cases are ready simultaneously, **one is chosen at random** (fair scheduling).

---

## 10.2 Basic `select` Syntax

```go
select {
case val := <-ch1:
    // received val from ch1
case ch2 <- value:
    // sent value to ch2
case val, ok := <-ch3:
    // received from ch3, ok tells if channel is still open
default:
    // runs if no other case is ready (non-blocking)
}
```

**Rules:**
1. Each `case` must be a channel send OR receive.
2. `select` evaluates all cases, and blocks until one is ready.
3. If multiple cases are ready, one is picked **uniformly at random**.
4. `default` makes it non-blocking — if no channel is ready, execute default immediately.
5. An empty `select {}` blocks forever (useful for keeping a server's main goroutine alive).

> See: [example_basic_select.go](example_basic_select.go)

---

## 10.3 Timeout Pattern

The most common `select` pattern — essential for any network code:

```go
// Python equivalent:
//   try:
//       result = await asyncio.wait_for(some_task(), timeout=2.0)
//   except asyncio.TimeoutError:
//       print("timed out!")

select {
case result := <-longOperation():
    fmt.Println("Got result:", result)
case <-time.After(2 * time.Second):
    fmt.Println("Timed out!")
}
```

`time.After(d)` returns a channel that receives a value after duration `d`. So `<-time.After(2*time.Second)` blocks for 2 seconds, then becomes ready — which triggers the timeout case if the main operation hasn't completed.

**Why this is elegant:** No special timeout API. No `asyncio.wait_for`. Timeouts are just channels. Everything in Go composes through channels.

> See: [example_timeout.go](example_timeout.go)

---

## 10.4 Non-Blocking Operations with `default`

Without `default`, `select` blocks. With `default`, it becomes a non-blocking try:

```go
// Non-blocking receive
select {
case msg := <-ch:
    fmt.Println("Received:", msg)
default:
    fmt.Println("No message available (would have blocked)")
}

// Non-blocking send
select {
case ch <- "hello":
    fmt.Println("Sent!")
default:
    fmt.Println("Channel full, dropped message")
}
```

Python equivalent:
```python
try:
    msg = q.get_nowait()  # raises Empty if nothing available
except asyncio.QueueEmpty:
    print("No message available")
```

> See: [example_nonblocking.go](example_nonblocking.go)

---

## 10.5 Select in a Loop — Event Loop Pattern

The most powerful pattern: `for` + `select` = your own event loop.

```go
for {
    select {
    case msg := <-msgChan:
        handleMessage(msg)
    case err := <-errChan:
        handleError(err)
    case <-quit:
        fmt.Println("Shutting down")
        return
    }
}
```

This is conceptually similar to Python's event loop running forever, dispatching events:
```python
while True:
    event = await get_next_event()  # could be msg, error, or quit
    dispatch(event)
```

Except Go's `select` handles multiple event sources natively with type safety.

> See: [example_event_loop.go](example_event_loop.go)

---

## 10.6 `select` with `nil` Channels — Dynamic Enable/Disable

A nil channel **always blocks**. In a `select`, a nil channel's case is **never chosen**. This lets you dynamically enable/disable channels:

```go
var ch1 chan int = nil  // disabled
ch2 := make(chan int)

select {
case v := <-ch1:   // never selected (ch1 is nil)
    fmt.Println(v)
case v := <-ch2:   // this is the only active case
    fmt.Println(v)
}
```

**Use case:** You're reading from two channels, and one finishes before the other. Set the finished one to `nil` so select ignores it:

```go
for ch1 != nil || ch2 != nil {
    select {
    case v, ok := <-ch1:
        if !ok {
            ch1 = nil  // disable this case
            continue
        }
        process(v)
    case v, ok := <-ch2:
        if !ok {
            ch2 = nil  // disable this case
            continue
        }
        process(v)
    }
}
```

Python has no equivalent — you'd need to remove completed tasks from a set.

---

## 10.7 `select` Fairness — Random Choice

When multiple cases are ready, `select` chooses one **uniformly at random**. This prevents starvation:

```go
ch1 := make(chan string, 10)
ch2 := make(chan string, 10)

// Fill both
for i := 0; i < 10; i++ {
    ch1 <- "from ch1"
    ch2 <- "from ch2"
}

// select picks randomly — roughly 50/50
for i := 0; i < 10; i++ {
    select {
    case msg := <-ch1:
        fmt.Println(msg)
    case msg := <-ch2:
        fmt.Println(msg)
    }
}
```

This is important for fairness — without randomness, a busy channel could starve a less busy one.

---

## Exercises

### Exercise 1: First Response Wins
Create 3 goroutines that each "query" a different "server" (simulate with random sleep). Use `select` to get whichever response comes first and ignore the others.

### Exercise 2: Heartbeat Monitor
Write a goroutine that sends a "heartbeat" to a channel every 500ms. In main, use `select` in a loop to:
- Print "alive" when heartbeat arrives
- Print "DEAD — no heartbeat" if no heartbeat for 2 seconds
- Quit after 5 seconds total

### Exercise 3: Convert asyncio.wait
Rewrite this Python code using Go's `select`:
```python
import asyncio

async def fast():
    await asyncio.sleep(0.5)
    return "fast result"

async def slow():
    await asyncio.sleep(2.0)
    return "slow result"

async def main():
    task1 = asyncio.create_task(fast())
    task2 = asyncio.create_task(slow())
    done, pending = await asyncio.wait(
        {task1, task2}, return_when=asyncio.FIRST_COMPLETED
    )
    for task in done:
        print(task.result())
    for task in pending:
        task.cancel()

asyncio.run(main())
```

---

> **Next → [Lesson 11: Sync Primitives](../11_sync_primitives/lesson.md)** — WaitGroup, Mutex, Once and friends.
