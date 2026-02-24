# Lesson 1: Go Syntax & Idioms That Differ From Python

> **Goal:** Get you writing idiomatic Go in 30 minutes by focusing *only* on what's different from Python. No "this is a variable" — we jump straight to the weird parts.

---

## 1.1 Declarations: Explicit Types, Short Syntax

### Python
```python
name = "Shankar"       # type inferred
age: int = 21          # type hint (optional, not enforced at runtime)
```

### Go
```go
var name string = "Shankar"  // explicit type
var age int = 21

// Short declaration (most common inside functions):
name := "Shankar"   // type inferred by compiler (like Python, but enforced)
age := 21
```

**Key differences:**
| | Python | Go |
|---|--------|-----|
| Type enforcement | Runtime (duck typing) | Compile-time (static) |
| `:=` | Doesn't exist | Declares + assigns (only inside functions) |
| `var` | Doesn't exist | Used at package level or when you want to declare without assigning |
| Zero values | `None` or error | Every type has a zero value: `0`, `""`, `false`, `nil` |

**Why zero values matter:** In Python, accessing an uninitialized variable raises `NameError`. In Go, every variable is *always* initialized — to its zero value if you don't give one. This eliminates an entire class of bugs.

```go
var count int      // count is 0, not nil, not undefined
var msg string     // msg is "", not None
var active bool    // active is false
var ptr *int       // ptr is nil (pointers are the exception)
```

> See: [example_declarations.go](example_declarations.go)

---

## 1.2 Multiple Return Values (Go's Superpower for Error Handling)

### Python
```python
def divide(a, b):
    if b == 0:
        raise ValueError("division by zero")
    return a / b
```

### Go
```go
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, fmt.Errorf("division by zero")
    }
    return a / b, nil
}
```

**Why this matters:** Go doesn't have exceptions (`try/except`). Instead, functions return an `error` as a second value. This is *the* most important idiom in Go:

```go
result, err := divide(10, 0)
if err != nil {
    fmt.Println("Error:", err)
    return
}
fmt.Println(result)
```

**Mental model:** Think of every Go function call as Python's:
```python
result, err = divide(10, 0)  # as if every function returns (value, Optional[Error])
if err is not None:
    print("Error:", err)
    return
```

Except in Go this isn't optional — it's the **standard pattern**. You'll write `if err != nil` hundreds of times. This is intentional: Go forces you to handle every error at the call site, unlike Python where exceptions can silently propagate up the stack.

> See: [example_multiple_returns.go](example_multiple_returns.go)

---

## 1.3 Visibility: Uppercase = Public (No `__init__`, No `_private`)

In Python:
```python
class Server:
    def __init__(self):
        self._port = 8080       # "private" by convention (underscore)
        self.host = "localhost"  # public by convention

    def _internal_method(self):  # "private" by convention
        pass
```

In Go, there are **no classes**. But the visibility rule is dead simple:

- **Uppercase first letter → Exported (public):** `Server`, `HandleRequest`, `Port`
- **Lowercase first letter → Unexported (private to package):** `server`, `handleRequest`, `port`

```go
package server

// Exported — other packages can use this
type Server struct {
    Host string   // Exported field
    port int      // unexported — only this package can access
}

// Exported function
func NewServer() *Server {
    return &Server{Host: "localhost", port: 8080}
}

// unexported function
func helper() {
    // only callable within the 'server' package
}
```

**Why this design:** Python's underscore convention is just a *hint* — nothing stops you from accessing `obj._private`. Go enforces visibility at the **compiler level**. If it's lowercase, other packages literally cannot see it. No decorators, no `__all__`, no gentleman's agreements.

---

## 1.4 No Classes — Structs + Methods

This is probably the biggest mental shift from Python.

### Python
```python
class Dog:
    def __init__(self, name: str, age: int):
        self.name = name
        self.age = age

    def bark(self) -> str:
        return f"{self.name} says woof!"

d = Dog("Rex", 3)
print(d.bark())
```

### Go
```go
type Dog struct {
    Name string
    Age  int
}

// Method with a "receiver" — this is how you attach methods to structs
func (d Dog) Bark() string {
    return d.Name + " says woof!"
}

func main() {
    d := Dog{Name: "Rex", Age: 3}
    fmt.Println(d.Bark())
}
```

**Breaking it down:**
- `type Dog struct { ... }` → Like Python's `@dataclass` — defines the shape/fields.
- `func (d Dog) Bark()` → The `(d Dog)` part is the **receiver**. It's like Python's `self`, but declared *before* the function name instead of as a parameter.
- No `__init__`. You create structs with literal syntax: `Dog{Name: "Rex", Age: 3}`.
- No inheritance. Period. Go uses **composition** (embedding structs) and **interfaces** (Lesson 4).

**Value receiver vs pointer receiver** (critical — ties into Lesson 3 on pointers):
```go
// Value receiver — gets a COPY of Dog (like passing by value)
func (d Dog) Bark() string { ... }

// Pointer receiver — gets the actual Dog (like Python's self — can modify it)
func (d *Dog) SetAge(age int) {
    d.Age = age  // modifies the original
}
```

**Mental model:** In Python, `self` is always a reference (pointer). In Go, you choose: value receiver (read-only copy) or pointer receiver (can modify original).

> See: [example_structs.go](example_structs.go)

---

## 1.5 Packages & `main` — Not a Script Language

### Python
```python
# any .py file can be run directly
# server.py
if __name__ == "__main__":
    print("starting server")
```

### Go
```go
// Every Go file belongs to a package
package main  // "main" is special — it's the entry point

import "fmt"  // explicit imports (no implicit builtins)

func main() {  // THE entry point — like if __name__ == "__main__"
    fmt.Println("starting server")
}
```

**Key differences:**
- Every `.go` file starts with `package <name>`.
- Only `package main` with a `func main()` can be run as a program.
- **All imports must be used.** Unused import → compile error. (Python ignores unused imports.)
- **All declared variables must be used.** Unused variable → compile error. Go forces clean code.

This is where Go's philosophy shows: the compiler is strict so your team's code stays clean. Python trusts you; Go doesn't.

---

## 1.6 `for` Is the Only Loop (No `while`)

```go
// Classic for (like Python's for i in range(10))
for i := 0; i < 10; i++ {
    fmt.Println(i)
}

// "While" loop (just for with one condition)
for count < 100 {
    count++
}

// Infinite loop (like while True)
for {
    // break when needed
}

// Range loop (like Python's for item in list)
fruits := []string{"apple", "banana", "cherry"}
for index, fruit := range fruits {
    fmt.Println(index, fruit)
}

// If you don't need the index (like Python's _ convention):
for _, fruit := range fruits {
    fmt.Println(fruit)
}
```

**The `_` blank identifier:** Same idea as Python's `_` in `for _, item in enumerate(list)`. Go actually *enforces* this — if you declare `index` and don't use it, the compiler yells at you. `_` tells the compiler "I know, discard it."

---

## 1.7 `defer` — Like Python's `with` Statement, But More Flexible

### Python
```python
with open("file.txt") as f:
    data = f.read()
# f is automatically closed here
```

### Go
```go
f, err := os.Open("file.txt")
if err != nil {
    log.Fatal(err)
}
defer f.Close()  // will run when this function returns, no matter what

data, err := io.ReadAll(f)
```

**How `defer` works:**
1. When the runtime hits `defer f.Close()`, it **schedules** `f.Close()` to run when the enclosing function returns.
2. Multiple defers execute in **LIFO** (stack) order — last deferred, first executed.
3. Deferred calls run even if the function panics (like Python's `finally`).

**Why not `with`?** Go's `defer` is more general — it works with any function call, not just context managers. You'll use it for:
- Closing files, database connections, HTTP response bodies
- Unlocking mutexes (`defer mu.Unlock()`)
- Timing functions (`defer timeTrack(time.Now(), "myFunc")`)

```go
func doStuff() {
    fmt.Println("start")
    defer fmt.Println("deferred 1")
    defer fmt.Println("deferred 2")
    defer fmt.Println("deferred 3")
    fmt.Println("end")
}
// Output:
// start
// end
// deferred 3  ← LIFO order
// deferred 2
// deferred 1
```

> See: [example_defer.go](example_defer.go)

---

## 1.8 Semicolons, Braces & Formatting — The Compiler Is Your Linter

| | Python | Go |
|---|--------|-----|
| Blocks | Indentation | Curly braces `{}` (always required) |
| Semicolons | Never | Inserted automatically by compiler (don't write them) |
| Formatting | PEP8 (convention) | `gofmt` (enforced, one true style) |
| Opening brace | N/A | **Must** be on same line (or compile error) |

```go
// ✅ Correct — opening brace on same line
if x > 0 {
    fmt.Println("positive")
}

// ❌ COMPILE ERROR — opening brace on next line
if x > 0
{
    fmt.Println("positive")
}
```

**Why?** Go's lexer automatically inserts semicolons at end of lines. If you put `{` on the next line, it inserts a semicolon after `if x > 0`, breaking the syntax. This means there's literally ONE way to format Go code. No tabs-vs-spaces debates.

Run `gofmt` or `goimports` on save (your editor probably already does this) — it auto-formats everything.

---

## 1.9 No Generics-Style Duck Typing (Static Interfaces)

### Python
```python
# Python doesn't care about types — if it quacks, it's a duck
def print_length(thing):
    print(len(thing))  # works for str, list, dict, anything with __len__

print_length("hello")
print_length([1, 2, 3])
```

### Go
```go
// Go needs to know the type at compile time
// But interfaces give you "structural typing" — a compile-time version of duck typing

type HasLength interface {
    Len() int
}

func printLength(thing HasLength) {
    fmt.Println(thing.Len())
}
```

A type satisfies an interface **implicitly** — no `implements` keyword, no `class MyClass(ABC)`. If your struct has the right methods, it satisfies the interface automatically. This is Go's version of duck typing, but checked at compile time.

We'll dive deep into this in Lesson 4. For now, just know: **interfaces are Go's answer to Python's duck typing, but with compile-time safety.**

---

## Summary: Python → Go Mental Translation Table

| Python | Go | Notes |
|--------|-----|-------|
| `x = 5` | `x := 5` | Inside functions only |
| `def func():` | `func func() { }` | Braces, not indentation |
| `class Dog:` | `type Dog struct { }` | No classes, no inheritance |
| `self.name` | `d.Name` (receiver) | Uppercase = exported |
| `raise ValueError(...)` | `return fmt.Errorf(...)` | Errors are values, not exceptions |
| `try/except` | `if err != nil { }` | Must handle every error explicitly |
| `with open(...) as f:` | `defer f.Close()` | Deferred cleanup |
| `from module import X` | `import "package"` | Must use everything you import |
| `_` in `for _, x in ...` | `_` in `for _, x := range ...` | Blank identifier |
| `None` | `nil` | Only for pointers, interfaces, slices, maps, channels, functions |
| `pip install` | `go get` / `go mod tidy` | Module system |
| `if __name__ == "__main__":` | `func main()` in `package main` | Entry point |

---

## Exercises

After reading and running the examples, try these:

### Exercise 1: Multiple Returns
Write a function `safeSqrt(x float64) (float64, error)` that returns an error if `x` is negative, otherwise returns the square root. Handle the error at the call site.

### Exercise 2: Struct + Methods
Create a `Rectangle` struct with `Width` and `Height` fields. Add methods:
- `Area() float64` (value receiver)
- `Scale(factor float64)` (pointer receiver — multiplies both dimensions)

### Exercise 3: Defer Ordering
Write a function with 5 `defer` statements. Predict the output order before running it.

### Exercise 4: Unused Variables
Try declaring a variable you don't use, and an import you don't use. See what the compiler says. (Then fix it.)

---

> **Next up → [Lesson 2: Error Handling — No Exceptions Allowed](../02_error_handling/lesson.md)** 
>
> Questions about anything above? Ask before moving on — this is mentorship, not a race.
