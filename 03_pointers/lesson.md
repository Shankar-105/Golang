# Lesson 3: Pointers — What Python Hides From You

> **Goal:** Understand Go's pointer system. In Python, everything is a reference and you never think about it. In Go, you choose: **copy the value** or **share a pointer to it**.

---

## 3.1 Python vs Go: The Memory Model

In Python, **every variable is a reference** (a pointer under the hood):

```python
a = [1, 2, 3]
b = a          # b points to the SAME list
b.append(4)
print(a)       # [1, 2, 3, 4] — both see the change!
```

In Go, **variables hold the actual value** by default:

```go
a := [3]int{1, 2, 3}  // array (fixed size)
b := a                  // b is a COPY of a
b[0] = 999
fmt.Println(a)          // [1 2 3] — a is unchanged!
```

| Concept | Python | Go |
|---------|--------|-----|
| Assignment | Copies the reference (share) | Copies the value (independent) |
| Mutation | Both variables see changes | Only the copy changes |
| Explicit sharing | Not needed (default) | Use pointers: `&` and `*` |

---

## 3.2 What Is a Pointer?

A **pointer** is a variable that holds the **memory address** of another variable.

```
x := 42

   x          p
┌──────┐    ┌──────┐
│  42  │    │ &x   │ ← p holds the ADDRESS of x
└──────┘    └──────┘
 0x1000      0x2000          p "points to" x
```

### Syntax

```go
x := 42
p := &x    // & = "address of x" — p is a *int (pointer to int)
fmt.Println(p)   // 0xc0000b2008 (memory address)
fmt.Println(*p)  // 42 — * dereferences: "value at that address"

*p = 100         // change the value at the address
fmt.Println(x)   // 100 — x changed because p points to it!
```

### The Two Operators

| Operator | Name | What It Does | Example |
|----------|------|-------------|---------|
| `&` | Address-of | Gets the pointer to a variable | `p := &x` |
| `*` | Dereference | Gets the value a pointer points to | `val := *p` |

And in type declarations:
| Syntax | Meaning |
|--------|---------|
| `*int` | "pointer to an int" |
| `*string` | "pointer to a string" |
| `*User` | "pointer to a User struct" |

> See: [example_basics.go](example_basics.go)

---

## 3.3 Why Do Pointers Exist?

### Reason 1: Avoid Expensive Copies

Structs can be large. Passing by value copies everything:

```go
type BigConfig struct {
    // imagine 50 fields, nested structs, etc.
    Data [1000000]byte
}

func processCopy(c BigConfig) {      // copies 1MB every call!
    // ...
}

func processPointer(c *BigConfig) {  // copies 8 bytes (the pointer)
    // ...
}
```

Python equivalent: not an issue — Python always passes references.

### Reason 2: Functions Need to Modify the Caller's Data

```go
// This does NOT work — n is a copy
func tryToDouble(n int) {
    n = n * 2  // modifies the copy only
}

// This WORKS — p points to the original
func double(p *int) {
    *p = *p * 2  // modifies the original
}

x := 5
tryToDouble(x)
fmt.Println(x)  // 5 — unchanged!

double(&x)
fmt.Println(x)  // 10 — modified!
```

Python equivalent (where this is automatic):
```python
def append_item(lst):
    lst.append(42)  # modifies the original — lists are references

my_list = [1, 2, 3]
append_item(my_list)
print(my_list)  # [1, 2, 3, 42]
```

> See: [example_why_pointers.go](example_why_pointers.go)

---

## 3.4 Pointer Receivers vs Value Receivers

When defining methods on structs, you choose between a **value receiver** and a **pointer receiver**:

```go
type Counter struct {
    Count int
}

// Value receiver — operates on a COPY
func (c Counter) GetCount() int {
    return c.Count  // reading is fine
}

// Pointer receiver — operates on the ORIGINAL
func (c *Counter) Increment() {
    c.Count++  // this modifies the actual struct
}
```

### When to Use Which

| Use Pointer Receiver `(c *T)` | Use Value Receiver `(c T)` |
|-------------------------------|---------------------------|
| Method modifies the struct | Method only reads |
| Struct is large (avoid copy) | Struct is small (int, string) |
| Consistency (if any method is pointer, all should be) | Immutable / stateless |

Python equivalent: In Python, `self` is always a reference, so every method can modify the object. In Go, only pointer receivers can modify.

```python
# Python — self is always a reference
class Counter:
    def __init__(self):
        self.count = 0

    def increment(self):  # always modifies the original
        self.count += 1
```

> See: [example_receivers.go](example_receivers.go)

---

## 3.5 `nil` Pointers — Go's Version of `None`

A pointer that doesn't point to anything is `nil`:

```go
var p *int      // declared but not assigned → nil
fmt.Println(p)  // <nil>

// DANGER: dereferencing nil crashes!
// fmt.Println(*p)  // runtime panic: invalid memory address
```

Python equivalent:
```python
x = None
x.something  # AttributeError: 'NoneType'...
```

**Always check for nil before dereferencing:**

```go
func greet(name *string) string {
    if name == nil {
        return "Hello, stranger!"
    }
    return "Hello, " + *name + "!"
}
```

---

## 3.6 The `new()` Function

`new(T)` allocates memory for type T and returns a pointer to it (zero-valued):

```go
p := new(int)       // *int pointing to 0
fmt.Println(*p)     // 0

s := new(string)    // *string pointing to ""
fmt.Println(*s)     // ""
```

More commonly, you'll use the address-of operator `&` with a literal:

```go
// These are equivalent:
p1 := new(User)
p2 := &User{}       // preferred — more readable, can set fields

p3 := &User{Name: "Alice", Age: 25}  // even better
```

---

## 3.7 Quick Reference

```
x := 42

&x   → pointer to x    (type: *int)
*p   → value at p       (type: int)
*int → type "pointer to int"

Pass by value:   func f(x int)    — gets a copy
Pass by pointer: func f(x *int)   — gets the original (via address)
```

---

## 3.8 Common Gotchas

### Gotcha 1: You Can't Take the Address of a Literal

```go
// This does NOT work:
// p := &42  // ERROR: cannot take address of 42

// Do this instead:
x := 42
p := &x
```

### Gotcha 2: Slices, Maps, and Channels Are ALREADY References

```go
// Slices are already reference types — no pointer needed
func addItem(s []int) []int {
    return append(s, 42)
}

// Maps are already reference types
func addEntry(m map[string]int) {
    m["key"] = 99  // modifies the original map
}
```

This is why slices and maps feel like Python lists and dicts — they already behave like references internally.

> See: [example_gotchas.go](example_gotchas.go)

---

## Exercises

### Exercise 1: Swap
Write `func swap(a, b *int)` that swaps the values of two integers using pointers.

### Exercise 2: Pointer Receiver
Create a `BankAccount` struct with a `Balance` field. Add `Deposit(amount float64)` and `Withdraw(amount float64) error` methods using pointer receivers.

### Exercise 3: Optional Parameters
Write `func greet(name *string) string` that returns "Hello, World!" if name is nil, or "Hello, {name}!" otherwise.

---

> **Next → [Lesson 4: Structs & Interfaces](../04_structs_interfaces/lesson.md)** — How Go does OOP without classes.
