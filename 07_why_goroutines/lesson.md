# Lesson 7: Why Goroutines Exist — From First Principles

> **Goal:** Before writing a single goroutine, understand WHY they were invented — what problems they solve that Python's threading, multiprocessing, and asyncio cannot.

---

## 7.1 The Problem: Handling 10,000 Concurrent Connections

Imagine you're building a web server (like your FastAPI app). A request comes in, you process it, send a response. Simple. But what happens when **10,000 requests** arrive simultaneously?

You have three classical approaches. Let's examine why each one breaks.

---

## 7.2 Approach 1: One Thread Per Request (Python `threading`)

```
Client1 ──▶ Thread1 ──▶ Process ──▶ Response
Client2 ──▶ Thread2 ──▶ Process ──▶ Response
Client3 ──▶ Thread3 ──▶ Process ──▶ Response
...
Client10000 ──▶ Thread10000 ──▶ ??? 💥
```

**Why this fails:**

| Resource | Per OS Thread | With 10,000 threads |
|----------|--------------|-------------------|
| Stack memory | ~1-8 MB | 10-80 **GB** of RAM |
| Creation time | ~1 ms | 10 seconds just to spawn |
| Context switch cost | ~1-10 μs (kernel mode switch) | CPU spends more time switching than working |

**OS threads are heavy** because:
1. Each thread gets a **fixed-size stack** (typically 1-8 MB) allocated by the OS kernel.
2. Switching between threads requires a **kernel context switch**: save all CPU registers, switch page tables, flush caches. Expensive!
3. The OS scheduler has no idea what your application is doing — it schedules threads blindly using time slices.

**And then there's Python's GIL (Global Interpreter Lock):**
Even if you create 10,000 threads in Python, only ONE thread executes Python bytecode at a time. The GIL is a mutex that protects CPython's memory management. So Python threads give you concurrency (interleaving) but NOT parallelism (simultaneous execution).

```python
# Python threading — looks parallel, but the GIL serializes CPU work
import threading

def cpu_work():
    total = sum(range(10_000_000))  # CPU-bound

# These run ONE AT A TIME due to the GIL!
threads = [threading.Thread(target=cpu_work) for _ in range(4)]
for t in threads: t.start()
for t in threads: t.join()
```

**Analogy:** Imagine a restaurant with 10,000 waiters (threads), but only ONE kitchen door (the GIL). Waiters line up to enter the kitchen one at a time, even though there are 8 stoves (CPU cores) inside. Most waiters are just standing around.

---

## 7.3 Approach 2: Multiple Processes (Python `multiprocessing`)

To escape the GIL, Python uses separate OS processes:

```python
from multiprocessing import Pool

def cpu_work(n):
    return sum(range(n))

with Pool(8) as pool:  # 8 separate Python interpreters
    results = pool.map(cpu_work, [10_000_000] * 8)
```

**This achieves true parallelism, but:**
1. Each process has its own **entire Python interpreter** (~30-50 MB of memory).
2. Communication between processes requires **serialization** (pickle) — slow for large data.
3. Creating a process is ~100x more expensive than creating a thread.
4. You can't share memory easily (no shared variables — need `multiprocessing.Manager`, shared memory, etc.).

**Analogy:** Instead of one restaurant with many waiters, you build 8 **separate restaurants**, each with its own kitchen, ingredients, and staff. True parallelism, but incredibly wasteful.

For 10,000 connections? You can't create 10,000 processes. The OS will kill you.

---

## 7.4 Approach 3: Event Loop (Python `asyncio`)

This is what you know well. Instead of one thread per connection, use ONE thread with an event loop:

```python
import asyncio

async def handle_request(reader, writer):
    data = await reader.read(1024)   # yields control while waiting for I/O
    await asyncio.sleep(0.1)         # simulates async work
    writer.write(b"HTTP/1.1 200 OK\r\n\r\nHello!")
    await writer.drain()
    writer.close()

async def main():
    server = await asyncio.start_server(handle_request, '0.0.0.0', 8080)
    await server.serve_forever()

asyncio.run(main())
```

**How asyncio works internally:**
```
┌─────────────────────────────────────────┐
│           Event Loop (1 thread)          │
│                                          │
│   ┌──────┐  ┌──────┐  ┌──────┐         │
│   │Task A│  │Task B│  │Task C│  ...     │
│   └──┬───┘  └──┬───┘  └──┬───┘         │
│      │         │         │               │
│   await ──► run B ──► await ──► run A    │
│   (yield)           (yield)    (resume)  │
└─────────────────────────────────────────┘
         ONE thread, cooperative switching
```

**What asyncio does well:**
- Handles 10,000+ I/O-bound connections on ONE thread.
- Extremely low overhead per coroutine (~few KB).
- No context switches for I/O waits.

**What asyncio CANNOT do:**
1. **CPU-bound work blocks everything.** If one coroutine does heavy computation without `await`, ALL other coroutines freeze. There's no preemption — it's cooperative.

    ```python
    async def bad_handler():
        # This blocks the ENTIRE event loop for 5 seconds!
        # No other request can be served during this time.
        total = sum(range(100_000_000))  # no await = no yielding
        return total
    ```

2. **Single-threaded.** Cannot use multiple CPU cores. For CPU work, you still need `multiprocessing` or `concurrent.futures.ProcessPoolExecutor`:

    ```python
    import asyncio
    from concurrent.futures import ProcessPoolExecutor

    def cpu_heavy(n):
        return sum(range(n))

    async def handler():
        loop = asyncio.get_event_loop()
        # Escape to a process to do CPU work — clunky!
        result = await loop.run_in_executor(
            ProcessPoolExecutor(), cpu_heavy, 100_000_000
        )
    ```

3. **Function coloring.** Every function in the call chain must be `async`. You can't call an async function from a sync function easily. This "infects" your entire codebase.

    ```python
    # "colored" functions — async and sync don't mix naturally
    async def fetch_data(): ...          # async (blue)
    def process(data): ...               # sync (red)
    async def handler():                 # must be async because it awaits
        data = await fetch_data()        # can only call from async context
        return process(data)
    ```

**Analogy:** asyncio is a single chef in a kitchen who starts boiling pasta, then while waiting, starts chopping vegetables, then checks the pasta... Very efficient for I/O (waiting for things), but if one task requires 10 minutes of continuous chopping (CPU), everything else stops.

---

## 7.5 Go's Solution: Goroutines — The Best of All Worlds

Go looked at all three approaches and said: "What if we make threads SO cheap that you can have millions of them, AND they run on multiple CPU cores?"

That's exactly what goroutines are.

```
┌─────────────── Go Runtime ──────────────┐
│                                          │
│  Goroutine1  Goroutine2  ...  Goroutine_N│
│      │           │                │      │
│      ▼           ▼                ▼      │
│  ┌────────────────────────────────────┐  │
│  │       M:N Scheduler               │  │
│  │  Maps N goroutines to M OS threads │  │
│  └──────┬──────────┬──────────┬───────┘  │
│         ▼          ▼          ▼          │
│     Thread1    Thread2    Thread3        │
│     (Core 1)   (Core 2)   (Core 3)      │
└─────────────────────────────────────────┘
              ▼          ▼          ▼
         ┌────────────────────────────┐
         │     OS Kernel / Hardware    │
         │     (actual CPU cores)      │
         └────────────────────────────┘
```

| Property | OS Thread | asyncio Coroutine | **Goroutine** |
|----------|-----------|-------------------|---------------|
| Memory | 1-8 MB stack | ~few KB | **2 KB initial stack** (grows dynamically!) |
| Creation | ~1 ms | ~1 μs | **~1 μs** |
| Scheduling | OS kernel (preemptive) | User-space (cooperative) | **User-space (preemptive since Go 1.14!)** |
| Parallelism | Yes (but GIL in Python) | No (single-threaded) | **Yes (M:N on real cores)** |
| CPU-bound | Blocked by GIL (Python) | Blocks event loop | **Runs truly parallel** |
| Context switch | Expensive (kernel) | Cheap (user-space) | **Cheap (user-space)** |
| Can have 1M+ | No (RAM/OS limit) | Yes | **Yes** |

**The magic is the M:N scheduler:**
- **N goroutines** (potentially millions) are multiplexed onto **M OS threads** (typically = number of CPU cores).
- The Go runtime scheduler (not the OS) decides which goroutine runs on which thread.
- When a goroutine does I/O or calls into the runtime, it yields — and another goroutine runs on that thread.
- Since Go 1.14, the scheduler also **preempts** goroutines that run too long (even CPU-bound ones), so no single goroutine can hog a thread.

**Analogy:** Imagine a restaurant with 8 chefs (OS threads, one per CPU core) and 10,000 orders (goroutines). A smart restaurant manager (the Go scheduler) constantly assigns the next order to whichever chef is free. If a chef is waiting for an oven (I/O), the manager gives them another order to prep in the meantime. If a chef has been chopping too long (CPU-bound), the manager taps their shoulder and says "take a break, let someone else chop" (preemption).

Compare to asyncio's restaurant: 1 chef, 10,000 orders, and the chef voluntarily decides when to switch (cooperative). If the chef decides to make a complicated dish without pausing, all other orders wait.

---

## 7.6 Why Not Just Use OS Threads Without the GIL?

Languages like Java and C++ create OS threads without a GIL. Why didn't Go just do that?

Because OS threads are still too expensive for 100,000+ concurrent tasks. A Java server handling 100K connections with one thread each would need ~100 GB of stack memory alone. Java's Project Loom (virtual threads) was actually inspired BY Go's goroutine model — it arrived 12 years later!

Go's insight: put a **user-space scheduler** between your code and the OS. This gives you:
1. **Tiny stacks** (2 KB, grows as needed) — the runtime manages memory, not the OS.
2. **Fast context switches** — no kernel involvement, just swapping a few pointers.
3. **Awareness of I/O** — the scheduler knows when a goroutine is waiting, so it can run others.

---

## 7.7 The GMP Model (Go's Scheduler Internals)

Go's scheduler uses three entities, called the **GMP model**:

```
G = Goroutine      (your concurrent task — lightweight, millions of them)
M = Machine        (an OS thread — typically one per CPU core)
P = Processor      (a scheduling context — holds the run queue)

GOMAXPROCS = number of P's = number of goroutines that can run truly parallel

┌──────────────────────────────────────────────────┐
│                  Go Runtime                       │
│                                                   │
│   P0 (run queue)         P1 (run queue)           │
│   ┌─────────────┐       ┌─────────────┐          │
│   │ G1, G4, G7  │       │ G2, G5, G8  │          │
│   └──────┬──────┘       └──────┬──────┘          │
│          │                     │                  │
│          ▼                     ▼                  │
│      M0 (thread)          M1 (thread)             │
│      ║                    ║                       │
│      ║ currently          ║ currently             │
│      ║ running G1         ║ running G2            │
│      ▼                    ▼                       │
│   OS Core 0            OS Core 1                  │
│                                                   │
│   Global run queue: [G3, G6, G9, ...]             │
│   (overflow from P's local queues)                │
└──────────────────────────────────────────────────┘
```

**How scheduling works step-by-step:**

1. You call `go myFunc()` → runtime creates a **G** (goroutine struct: ~2 KB stack + metadata).
2. The G is placed on the **local run queue** of the current P.
3. When the current G on a P finishes, blocks on I/O, or gets preempted, the P picks the next G from its local queue.
4. If the local queue is empty, the P **steals** work from another P's queue (work stealing!).
5. If ALL local queues are empty, check the **global run queue**.

**Work stealing** is what makes this efficient — if one P has 100 goroutines queued and another P is idle, the idle P steals half the work. No goroutine sits waiting while there's an idle core.

**What happens during I/O:**
1. Goroutine G1 on M0 calls `net.Read()` (a blocking I/O operation).
2. The runtime **parks** G1 (moves it to a waiting list).
3. P0 takes the next goroutine (G4) from its queue and runs it on M0.
4. When the I/O completes (the OS notifies via `epoll`/`kqueue`/`IOCP`), G1 is put back on a run queue.
5. **The thread M0 was NEVER blocked.** It kept doing useful work.

Compare to Python: `await reader.read(1024)` yields to the event loop. Conceptually similar, but on ONLY one thread. Go does this across ALL cores simultaneously.

---

## 7.8 Stack Growth — Why 2 KB Is Enough

OS thread stacks are fixed size (e.g., 8 MB on Linux). If you need more, you crash (stack overflow). If you need less, the memory is wasted.

Goroutine stacks **grow dynamically:**
1. Start at **2 KB** (4000x smaller than an 8 MB thread!).
2. When a function call would exceed the stack, the runtime allocates a new, larger stack (typically 2x).
3. Copies the old stack contents to the new stack.
4. Updates all pointers to point to the new locations.
5. The old stack is freed.

This is called **contiguous stack growth** (replaced the older "segmented stacks" in Go 1.4).

```
Goroutine starts:     [2 KB stack]
Function call chain:  [2 KB stack] → need more → [4 KB stack] → [8 KB stack] → ...
Maximum:              up to 1 GB (configurable with runtime.SetMaxStack)
```

**Why this matters for concurrency:** If each goroutine took 8 MB (like an OS thread), 1 million goroutines would need 8 TB of RAM. At 2 KB each, 1 million goroutines need only ~2 GB. That's why Go can handle massive concurrency.

---

## 7.9 Summary: Why Go > Python for Concurrency

| Scenario | Python Solution | Go Solution |
|----------|----------------|-------------|
| 10K I/O-bound connections | asyncio (1 thread) ✅ | Goroutines (multi-core) ✅✅ |
| CPU-bound parallelism | multiprocessing (heavy) 😐 | Goroutines (same syntax!) ✅✅ |
| Mixed I/O + CPU | asyncio + ProcessPool (complex) 😰 | Goroutines (just `go func()`) ✅✅ |
| 1M concurrent tasks | Possible with asyncio (single core) | Possible with goroutines (all cores) ✅✅ |
| Code complexity | async/await infects everything | No function coloring — same syntax |

**The killer feature:** In Go, the syntax for concurrent I/O and concurrent CPU work is **identical**: `go doSomething()`. There's no "async version" vs "sync version" of functions. The runtime handles everything.

> See: [example_why_goroutines.go](example_why_goroutines.go) — demonstrates goroutine creation and basic comparison
> See: [example_no_color.go](example_no_color.go) — shows Go has no "function coloring" problem

---

## Exercises

### Exercise 1: Mental Model Check
Without running code, answer: if you have 8 CPU cores and launch 100 goroutines doing CPU-heavy work, how many goroutines run *truly simultaneously*? What is `GOMAXPROCS` likely set to? What happens to the other 92?

### Exercise 2: Compare Throughput (Thought Experiment)
You have a web server handling requests. Each request does:
- 10ms of CPU work
- 90ms of I/O (database query)

Calculate the theoretical throughput (requests/sec) for:
1. Python single-threaded asyncio
2. Python multiprocessing with 8 workers
3. Go with `GOMAXPROCS=8`

(Hint: think about what blocks what.)

### Exercise 3: The asyncio Trap
What happens in this Python code? Why is it a problem? How would Go handle the same scenario?
```python
async def handle_request():
    data = heavy_computation()  # 500ms of CPU, no await
    await send_response(data)
```

---

> **Next → [Lesson 8: Goroutines Deep Dive](../08_goroutines/lesson.md)** — Time to write actual goroutine code!
