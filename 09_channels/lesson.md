# Lesson 9: Channels — Communicating Sequential Processes (CSP)

> **Goal:** Understand Go's primary concurrency communication mechanism — channels. Learn why Go says _"Don't communicate by sharing memory; share memory by communicating."_

---

## 9.1 The Problem: How Do Goroutines Talk to Each Other?

In Python asyncio, coroutines communicate through:
1. **Return values:** `result = await task`
2. **`asyncio.Queue`:** Producer/consumer pattern
3. **Shared variables:** (dangerous without locks)

In Go, goroutines are fire-and-forget (`go func()` returns nothing). How do they send results back? **Channels.**

**Why channels instead of shared memory?**

Shared memory + locks:
```
Goroutine A ──write──▶ [shared variable] ◀──read── Goroutine B
                              ↑
                          Need a lock! Easy to forget. Deadlock potential.
```

Channels:
```
Goroutine A ──send──▶ [channel] ──receive──▶ Goroutine B
                         ↑
                   Channel handles synchronization internally.
                   No locks needed in YOUR code.
```

**Analogy:** Shared memory is like a whiteboard in the office — anyone can write on it, but you need rules to prevent people overwriting each other. A channel is like a mailbox — one person puts a letter in, another takes it out. The mailbox itself ensures orderly delivery.

---

## 9.2 Channel Basics

### Creating a Channel

```go
ch := make(chan int)        // unbuffered channel of ints
ch := make(chan string, 10) // buffered channel of strings, capacity 10
ch := make(chan struct{})   // channel of empty structs (for signaling only)
```

Python equivalent:
```python
q = asyncio.Queue()         # unbuffered-ish (but actually has no max by default)
q = asyncio.Queue(maxsize=10)  # buffered with capacity 10
```

### Sending and Receiving

```go
ch <- 42       // send 42 into channel (blocks if channel is full/unbuffered)
value := <-ch  // receive from channel (blocks if channel is empty)
```

Python equivalent:
```python
await q.put(42)       # send (blocks if full)
value = await q.get() # receive (blocks if empty)
```

**Key difference:** Go's channel operations look like operators (`<-`), not method calls. The arrow shows the direction of data flow: `ch <- value` (into channel), `value := <-ch` (from channel).

> See: [example_basic_channel.go](example_basic_channel.go)

---

## 9.3 Unbuffered Channels — Synchronous Handoff

An **unbuffered** channel (`make(chan int)`) has zero capacity. A send blocks until a receiver is ready, and vice versa. It's a direct hand-off.

```
Sender goroutine:                  Receiver goroutine:
ch <- 42                           value := <-ch
   │                                   │
   ├── blocks here ◀──────────────────┤ arrives, takes value
   │                                   │
   ├── unblocks (value was taken) ────▶│ has value = 42
```

**This is synchronous communication.** The sender and receiver must rendezvous — both must be ready at the same time. It's like a relay race baton pass: the runner (sender) can't drop the baton and leave; they must wait for the next runner (receiver) to grab it.

```go
func main() {
    ch := make(chan string)  // unbuffered

    go func() {
        ch <- "hello"  // blocks until someone receives
        fmt.Println("sent!")  // only prints AFTER value was received
    }()

    time.Sleep(1 * time.Second) // receiver arrives late
    msg := <-ch  // now the sender unblocks
    fmt.Println("received:", msg)
}
```

**Deadlock trap with unbuffered channels:**
```go
func main() {
    ch := make(chan int)
    ch <- 42  // DEADLOCK! No other goroutine to receive.
    // fatal error: all goroutines are asleep - deadlock!
}
```

Go's runtime detects this — if ALL goroutines are blocked, it panics with a deadlock error. Very helpful for debugging!

> See: [example_unbuffered.go](example_unbuffered.go)

---

## 9.4 Buffered Channels — Asynchronous Queue

A **buffered** channel (`make(chan int, 5)`) has capacity. Sends don't block until the buffer is full. Receives don't block until the buffer is empty.

```go
ch := make(chan int, 3)  // buffer of 3

ch <- 1  // doesn't block (buffer: [1])
ch <- 2  // doesn't block (buffer: [1, 2])
ch <- 3  // doesn't block (buffer: [1, 2, 3])
// ch <- 4  // BLOCKS — buffer is full, waits for a receive

fmt.Println(<-ch)  // 1 (FIFO — first in, first out)
fmt.Println(<-ch)  // 2
fmt.Println(<-ch)  // 3
```

This is exactly like Python's `asyncio.Queue(maxsize=3)`:
```python
q = asyncio.Queue(maxsize=3)
await q.put(1)  # doesn't block
await q.put(2)  # doesn't block  
await q.put(3)  # doesn't block
# await q.put(4)  # blocks until someone calls q.get()
```

**When to use buffered vs unbuffered:**
| Use | Unbuffered | Buffered |
|-----|-----------|----------|
| **Synchronization** | ✅ Guarantees sender/receiver are in sync | ❌ No sync guarantee |
| **Speed** | Slower (both must rendezvous) | Faster (sends don't wait if buffer has room) |
| **Backpressure** | Immediate (sender waits every time) | Delayed (sender waits only when full) |
| **Use case** | Coordination, signals, handoffs | Producer-consumer, batching, rate limiting |

> See: [example_buffered.go](example_buffered.go)

---

## 9.5 Channel Direction — Type Safety for Communication

Go lets you restrict a channel to send-only or receive-only in function signatures:

```go
func producer(out chan<- int) {  // can only SEND to out
    out <- 42
    // val := <-out  // COMPILE ERROR — can't receive from send-only channel
}

func consumer(in <-chan int) {   // can only RECEIVE from in
    val := <-in
    fmt.Println(val)
    // in <- 99  // COMPILE ERROR — can't send to receive-only channel
}

func main() {
    ch := make(chan int)  // bidirectional
    go producer(ch)       // implicitly narrows to chan<-
    consumer(ch)          // implicitly narrows to <-chan
}
```

**Why this matters:** It's documentation AND enforcement. When you see `func process(in <-chan Request)`, you know this function only reads from the channel. It can't accidentally send to it. Python has no equivalent — `asyncio.Queue` is always read-write.

---

## 9.6 Closing Channels — Signaling "No More Data"

```go
ch := make(chan int, 5)
ch <- 1
ch <- 2
close(ch)  // signal: no more values will be sent

// Receiving from closed channel:
val, ok := <-ch  // val=1, ok=true (value was available)
val, ok = <-ch   // val=2, ok=true
val, ok = <-ch   // val=0, ok=false (channel is closed and empty)
```

**The `range` loop on channels** — the idiomatic way to drain a channel:
```go
go func() {
    for i := 0; i < 5; i++ {
        ch <- i
    }
    close(ch)  // MUST close, or range will block forever
}()

for val := range ch {  // loops until channel is closed
    fmt.Println(val)
}
```

Python equivalent:
```python
async def producer(q):
    for i in range(5):
        await q.put(i)
    await q.put(None)  # sentinel value — Python has no "close queue"

async def consumer(q):
    while True:
        val = await q.get()
        if val is None:
            break
        print(val)
```

**Key difference:** Go has a built-in `close()` + `range` pattern. Python needs a sentinel value or separate signaling.

**Rules for closing:**
1. Only the **sender** should close a channel. Never the receiver.
2. Sending to a closed channel **panics** (crash).
3. Receiving from a closed channel returns the zero value immediately.
4. You don't HAVE to close channels — they get garbage collected. Close only when receivers need to know "no more data."

> See: [example_close_range.go](example_close_range.go)

---

## 9.7 Channel Patterns

### Pattern 1: Done Channel (Signaling completion)
```go
done := make(chan struct{})  // empty struct = zero memory

go func() {
    // do work...
    close(done)  // signal completion
}()

<-done  // blocks until closed
```

### Pattern 2: Generator (yield values over time)
```go
func fibonacci(n int) <-chan int {
    ch := make(chan int)
    go func() {
        a, b := 0, 1
        for i := 0; i < n; i++ {
            ch <- a
            a, b = b, a+b
        }
        close(ch)
    }()
    return ch  // return receive-only channel
}

for val := range fibonacci(10) {
    fmt.Println(val)
}
```

This is like Python's generators:
```python
def fibonacci(n):
    a, b = 0, 1
    for _ in range(n):
        yield a
        a, b = b, a + b
```

### Pattern 3: Pipeline (chain of processing stages)
```go
// Stage 1: generate numbers
// Stage 2: square them
// Stage 3: print them
// Each stage is a goroutine communicating via channels

nums := generate(1, 2, 3, 4, 5)    // returns <-chan int
squared := square(nums)              // takes <-chan int, returns <-chan int
for val := range squared {
    fmt.Println(val)                 // 1, 4, 9, 16, 25
}
```

> See: [example_patterns.go](example_patterns.go)

---

## 9.8 Common Mistakes

### Mistake 1: Deadlock — Send with no receiver
```go
ch := make(chan int)
ch <- 42  // deadlock: main goroutine blocks, nothing else to receive
```

### Mistake 2: Sending to a closed channel
```go
ch := make(chan int)
close(ch)
ch <- 42  // PANIC: send on closed channel
```

### Mistake 3: Forgetting to close → range blocks forever
```go
ch := make(chan int, 5)
go func() {
    for i := 0; i < 5; i++ {
        ch <- i
    }
    // forgot close(ch)!
}()

for val := range ch {  // blocks forever after receiving 5 values
    fmt.Println(val)
}
```

### Mistake 4: Reading from nil channel blocks forever
```go
var ch chan int  // nil channel (not initialized with make)
val := <-ch     // blocks forever — no deadlock detection for this!
```

---

## Exercises

### Exercise 1: Ping-Pong
Create two goroutines that send a "ball" (integer) back and forth through two channels, incrementing it each time. Stop after 10 volleys. Print the ball value at each hit.

### Exercise 2: Pipeline
Build a 3-stage pipeline with channels:
1. `generate()` → sends numbers 1-100 into a channel
2. `filter(in)` → reads from channel, passes only even numbers
3. `square(in)` → reads from channel, squares each number

Print the results.

### Exercise 3: Convert Python asyncio.Queue
Rewrite this Python code in Go using channels:
```python
import asyncio

async def producer(q):
    for i in range(10):
        await q.put(f"item-{i}")
        await asyncio.sleep(0.1)
    await q.put(None)  # done signal

async def consumer(q):
    while True:
        item = await q.get()
        if item is None:
            break
        print(f"Consumed: {item}")

async def main():
    q = asyncio.Queue(maxsize=3)
    await asyncio.gather(producer(q), consumer(q))

asyncio.run(main())
```

---

> **Next → [Lesson 10: Select Statement](../10_select/lesson.md)** — Multiplexing multiple channels.
