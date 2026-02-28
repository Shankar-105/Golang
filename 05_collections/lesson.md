# Lesson 5: Slices, Maps & Strings — Go's Core Collections

> **Goal:** Master Go's three most-used collection types. They're similar to Python's `list`, `dict`, and `str` — but with important differences in how they work under the hood.

---

## 5.1 Overview: Python vs Go Collections

| Python | Go | Notes |
|--------|-----|-------|
| `list` | `[]T` (slice) | Dynamic arrays, but with capacity |
| `tuple` | `[N]T` (array) | Fixed-size, rarely used directly |
| `dict` | `map[K]V` | Hash map, unordered |
| `set` | *(no built-in)* | Use `map[T]bool` or `map[T]struct{}` |
| `str` | `string` | Immutable, UTF-8 byte sequence |

---

## 5.2 Arrays — Fixed Size (You'll Rarely Use These)

Arrays have a **fixed size** that's part of the type:

```go
var a [5]int                    // [0 0 0 0 0]
b := [3]string{"go", "is", "fun"} // [go is fun]

// Size is part of the type:
// [3]int and [5]int are DIFFERENT types (can't assign one to the other)
```

Python equivalent: No real equivalent. `tuple` is closest (fixed once created), but tuples can hold mixed types.

**You almost always want slices instead.** Arrays exist mainly as the backing store for slices.

---

## 5.3 Slices — Go's "Lists"

A slice is a **dynamic, resizable view** over an underlying array.

```go
// Create slices
nums := []int{1, 2, 3, 4, 5}              // literal
names := make([]string, 0, 10)             // make(type, length, capacity)
var empty []int                             // nil slice (length 0)
```

### Slice Anatomy

A slice is a 3-field struct internally:
```
┌─────────┬────────┬──────────┐
│ pointer │ length │ capacity │
└─────────┴────────┴──────────┘
```

- **Pointer:** address of the first element in the underlying array
- **Length:** how many elements the slice currently holds (`len(s)`)
- **Capacity:** how many elements the underlying array can hold (`cap(s)`)

```go
s := make([]int, 3, 10)
fmt.Println(len(s))  // 3
fmt.Println(cap(s))  // 10
```

### Append — Growing a Slice

```go
s := []int{1, 2, 3}
s = append(s, 4)           // add one element
s = append(s, 5, 6, 7)     // add multiple
s = append(s, []int{8, 9}...) // append another slice (... unpacks it)
```

Python equivalent:
```python
s = [1, 2, 3]
s.append(4)
s.extend([5, 6, 7])
s += [8, 9]
```

**Important:** `append` may or may not create a new underlying array. Always reassign: `s = append(s, x)`.

### Slicing (Sub-slices)

```go
s := []int{0, 1, 2, 3, 4, 5}

a := s[1:4]   // [1 2 3] — elements at index 1, 2, 3
b := s[:3]    // [0 1 2] — from start
c := s[3:]    // [3 4 5] — to end
d := s[:]     // [0 1 2 3 4 5] — full copy... or is it?
```

**GOTCHA:** Sub-slices **share** the underlying array! Modifying one affects the other.

```go
original := []int{1, 2, 3, 4, 5}
sub := original[1:3]  // [2 3]
sub[0] = 999
fmt.Println(original)  // [1 999 3 4 5] — changed!
```

Python equivalent: Python slices create copies. Go slices share memory.
```python
original = [1, 2, 3, 4, 5]
sub = original[1:3]  # [2, 3] — this is a COPY
sub[0] = 999
print(original)  # [1, 2, 3, 4, 5] — unchanged
```

### nil vs Empty Slice

```go
var s1 []int        // nil slice — s1 == nil is true
s2 := []int{}       // empty slice — s2 == nil is false
s3 := make([]int, 0) // empty slice — s3 == nil is false

// All three have length 0 and work with append:
len(s1) // 0
len(s2) // 0
s1 = append(s1, 42) // works fine!
```

**Rule of thumb:** Prefer `var s []int` (nil) unless you need to distinguish nil from empty (like in JSON: `null` vs `[]`).

> See: [example_slices.go](example_slices.go)

---

## 5.4 Maps — Go's "Dicts"

```go
// Create maps
ages := map[string]int{
    "Alice": 25,
    "Bob":   30,
}

// Empty map
scores := make(map[string]int)

// Add / update
ages["Charlie"] = 35
ages["Alice"] = 26

// Delete
delete(ages, "Bob")

// Get value
age := ages["Alice"]  // 26
```

### The Comma-OK Pattern

```go
age, ok := ages["Dave"]
if ok {
    fmt.Println("Dave's age:", age)
} else {
    fmt.Println("Dave not found")
}
```

Python equivalent:
```python
ages = {"Alice": 25, "Bob": 30}

# Get with default
age = ages.get("Dave", None)

# Or try/except
try:
    age = ages["Dave"]
except KeyError:
    print("Dave not found")
```

### Iterating Maps

```go
for key, value := range ages {
    fmt.Printf("  %s: %d\n", key, value)
}
```

**Warning:** Map iteration order is **random** in Go (intentionally). Don't depend on order.

Python equivalent (ordered since 3.7):
```python
for key, value in ages.items():
    print(f"  {key}: {value}")
```

### nil Map Gotcha

```go
var m map[string]int  // nil map — reading returns zero value
fmt.Println(m["key"]) // 0 (no panic)

// But WRITING to a nil map panics!
// m["key"] = 42  // panic: assignment to entry in nil map

// Always initialize:
m = make(map[string]int)  // now safe to write
```

### Using Maps as Sets

```go
// Go has no built-in set type. Use map[T]bool or map[T]struct{}
seen := map[string]bool{}
seen["apple"] = true
seen["banana"] = true

if seen["apple"] {
    fmt.Println("already seen apple")
}

// Memory-efficient version with struct{} (zero bytes per entry)
visited := map[string]struct{}{}
visited["page1"] = struct{}{}
if _, ok := visited["page1"]; ok {
    fmt.Println("visited page1")
}
```

> See: [example_maps.go](example_maps.go)

---

## 5.5 Strings — UTF-8 Byte Sequences

In Go, a string is an **immutable sequence of bytes** (usually UTF-8):

```go
s := "Hello, 世界"    // UTF-8 encoded
fmt.Println(len(s))   // 13 — bytes, NOT characters!
```

Python equivalent:
```python
s = "Hello, 世界"
len(s)  # 9 — characters (Python 3 strings are Unicode)
```

### Bytes vs Runes (Characters)

| Concept | Go | Python |
|---------|-----|--------|
| Single byte | `byte` (alias for `uint8`) | Not commonly used separately |
| Unicode code point | `rune` (alias for `int32`) | `str` character |
| String length in bytes | `len(s)` | `len(s.encode('utf-8'))` |
| String length in chars | `utf8.RuneCountInString(s)` | `len(s)` |

```go
s := "café"
fmt.Println(len(s))           // 5 bytes (é is 2 bytes in UTF-8)

// Iterate by byte:
for i := 0; i < len(s); i++ {
    fmt.Printf("%x ", s[i])   // 63 61 66 c3 a9
}

// Iterate by rune (character):
for i, r := range s {
    fmt.Printf("%d:%c ", i, r) // 0:c 1:a 2:f 3:é
}
```

### String Operations

```go
import "strings"

// Common operations — Python equivalents in comments
strings.Contains(s, "sub")       // "sub" in s
strings.HasPrefix(s, "He")       // s.startswith("He")
strings.HasSuffix(s, "ld")       // s.endswith("ld")
strings.ToUpper(s)               // s.upper()
strings.ToLower(s)               // s.lower()
strings.TrimSpace(s)             // s.strip()
strings.Split(s, ",")            // s.split(",")
strings.Join(parts, ",")         // ",".join(parts)
strings.Replace(s, "old", "new", -1)  // s.replace("old", "new")
strings.Count(s, "l")            // s.count("l")
strings.Index(s, "sub")          // s.find("sub") or s.index("sub")
```

### String Concatenation

```go
// Simple concatenation (ok for small amounts)
name := "Hello" + ", " + "World"

// For building strings efficiently (like Python's ''.join or io.StringIO)
import "strings"

var b strings.Builder
for i := 0; i < 1000; i++ {
    b.WriteString("hello ")
}
result := b.String()
```

> See: [example_strings.go](example_strings.go)

---

## 5.6 The `range` Keyword — Iterating Everything

```go
// Slice
for index, value := range nums { ... }
for _, value := range nums { ... }     // ignore index
for index := range nums { ... }        // ignore value

// Map
for key, value := range myMap { ... }
for key := range myMap { ... }         // keys only

// String (iterates runes, not bytes!)
for index, char := range "hello" { ... }

// Channel (we'll cover this in Phase 2)
for msg := range ch { ... }
```

Python equivalent:
```python
for i, v in enumerate(nums): ...
for v in nums: ...
for k, v in my_map.items(): ...
for ch in "hello": ...
```

---

## 5.7 `make` vs Literals — When to Use Which

| Expression | Creates | When to Use |
|-----------|---------|-------------|
| `[]int{1, 2, 3}` | Slice with data | You know the initial values |
| `make([]int, n)` | Slice of length n (zeroed) | You know the size, will fill later |
| `make([]int, 0, cap)` | Empty slice, pre-allocated | You'll append up to `cap` items |
| `map[K]V{"a": 1}` | Map with data | You know initial entries |
| `make(map[K]V)` | Empty map | You'll populate dynamically |
| `make(map[K]V, hint)` | Empty map, size hint | You know approximate size |

---

## 5.8 Quick Reference

```
Slice:  []T                    → dynamic list
        append(s, items...)    → grow
        s[low:high]            → sub-slice (SHARES memory!)
        len(s), cap(s)         → length and capacity
        make([]T, len, cap)    → pre-allocate

Map:    map[K]V                → hash map
        m[key] = value         → set
        value, ok := m[key]    → get + check
        delete(m, key)         → remove
        make(map[K]V)          → initialize

String: string                 → immutable UTF-8 bytes
        len(s)                 → byte count (not char count!)
        []rune(s)              → convert to rune slice for char ops
        strings.Builder        → efficient concatenation
```

---

## Exercises

### Exercise 1: Remove Duplicates
Write `func removeDuplicates(s []int) []int` that returns a new slice with duplicates removed. (Hint: use a map as a set.)

### Exercise 2: Word Counter
Write `func wordCount(text string) map[string]int` that counts word frequencies. (Hint: use `strings.Fields`.)

### Exercise 3: Matrix
Represent a 3x3 matrix as `[][]int`, write functions to print it and transpose it.

### Exercise 4: Reverse String
Write `func reverseString(s string) string` that correctly reverses a string with Unicode characters like "Hello, 世界" → "界世 ,olleH".

---

> **Next → [Lesson 6: Packages](../06_packages/lesson.md)** — Organizing Go code into modules and packages.
