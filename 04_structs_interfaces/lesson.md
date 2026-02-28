# Lesson 4: Structs & Interfaces — OOP Without Classes

> **Goal:** Learn Go's take on object-oriented programming. No classes, no inheritance, no `self`. Instead: **structs** (data), **methods** (behavior), **interfaces** (contracts), and **composition** (embedding).

---

## 4.1 The Big Picture

| Python | Go |
|--------|-----|
| `class User:` | `type User struct { ... }` |
| `self.name` | `u.Name` |
| `def method(self):` | `func (u User) method()` |
| `class Admin(User):` (inheritance) | Embedding `User` in `Admin` (composition) |
| `class ABC(metaclass=ABCMeta):` | `type Interface interface { ... }` |
| Duck typing (implicit, runtime) | Interfaces (implicit, **compile-time**) |
| `isinstance(obj, MyClass)` | Type assertion: `v, ok := obj.(MyType)` |

**Go's philosophy:** No inheritance hierarchy. No `class` keyword. Just data (structs) + behavior (methods) + contracts (interfaces).

---

## 4.2 Structs — Go's "Classes" (Without the Baggage)

A struct groups related data together:

```go
type User struct {
    Name  string
    Email string
    Age   int
}
```

Python equivalent:
```python
class User:
    def __init__(self, name: str, email: str, age: int):
        self.name = name
        self.email = email
        self.age = age

# Or with dataclass:
@dataclass
class User:
    name: str
    email: str
    age: int
```

### Creating Structs

```go
// Named fields (like keyword arguments)
u1 := User{Name: "Alice", Email: "alice@go.dev", Age: 25}

// Positional (must include ALL fields in order) — avoid this
u2 := User{"Bob", "bob@go.dev", 30}

// Zero value — all fields get their zero value
var u3 User  // {Name:"", Email:"", Age:0}

// Pointer to a struct
u4 := &User{Name: "Charlie", Age: 22}
```

### Accessing Fields

```go
fmt.Println(u1.Name)    // "Alice"
u1.Age = 26             // direct field access (no getters/setters needed)
```

> See: [example_structs.go](example_structs.go)

---

## 4.3 Methods — Attaching Behavior to Structs

In Go, methods are functions with a **receiver**:

```go
type Rectangle struct {
    Width, Height float64
}

// Method with a value receiver
func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}

// Method with a pointer receiver (can modify)
func (r *Rectangle) Scale(factor float64) {
    r.Width *= factor
    r.Height *= factor
}
```

Python equivalent:
```python
class Rectangle:
    def __init__(self, width, height):
        self.width = width
        self.height = height

    def area(self):           # self = Go's receiver
        return self.width * self.height

    def scale(self, factor):  # always modifies (self is a reference)
        self.width *= factor
        self.height *= factor
```

**Key difference:** In Python, `self` is always a reference. In Go, you choose value receiver (copy) vs pointer receiver (reference).

---

## 4.4 Composition Over Inheritance — Embedding

Go has **no inheritance**. Instead, you **embed** one struct inside another:

```go
// Base behavior
type Animal struct {
    Name string
    Age  int
}

func (a Animal) Speak() string {
    return a.Name + " makes a sound"
}

// "Inherits" Animal by embedding it
type Dog struct {
    Animal      // embedded — Dog gets all Animal fields and methods
    Breed string
}
```

```go
d := Dog{
    Animal: Animal{Name: "Rex", Age: 3},
    Breed:  "Labrador",
}

fmt.Println(d.Name)    // "Rex" — promoted from Animal
fmt.Println(d.Speak()) // "Rex makes a sound" — promoted method
fmt.Println(d.Breed)   // "Labrador" — Dog's own field
```

Python equivalent:
```python
class Animal:
    def __init__(self, name, age):
        self.name = name
        self.age = age

    def speak(self):
        return f"{self.name} makes a sound"

class Dog(Animal):  # inheritance
    def __init__(self, name, age, breed):
        super().__init__(name, age)
        self.breed = breed
```

### Overriding Methods

```go
// Dog can define its own Speak — "overrides" Animal's
func (d Dog) Speak() string {
    return d.Name + " barks!"
}
```

> See: [example_composition.go](example_composition.go)

---

## 4.5 Interfaces — Implicit Contracts

This is Go's most powerful feature. An interface defines **what methods a type must have**, not what it IS.

```go
type Shape interface {
    Area() float64
    Perimeter() float64
}
```

**Any type that has `Area()` and `Perimeter()` methods automatically satisfies `Shape`.**
No `implements` keyword. No registration. Just have the methods.

```go
type Circle struct {
    Radius float64
}

func (c Circle) Area() float64 {
    return math.Pi * c.Radius * c.Radius
}

func (c Circle) Perimeter() float64 {
    return 2 * math.Pi * c.Radius
}

// Circle satisfies Shape — no explicit declaration needed!
```

Python equivalent:
```python
from abc import ABC, abstractmethod

class Shape(ABC):
    @abstractmethod
    def area(self) -> float:
        pass

    @abstractmethod
    def perimeter(self) -> float:
        pass

class Circle(Shape):  # must explicitly inherit Shape
    def __init__(self, radius):
        self.radius = radius

    def area(self):
        return math.pi * self.radius ** 2

    def perimeter(self):
        return 2 * math.pi * self.radius
```

### Using Interfaces

```go
func printShapeInfo(s Shape) {
    fmt.Printf("  Area: %.2f, Perimeter: %.2f\n", s.Area(), s.Perimeter())
}

// Both Circle and Rectangle satisfy Shape
printShapeInfo(Circle{Radius: 5})
printShapeInfo(Rectangle{Width: 3, Height: 4})
```

> See: [example_interfaces.go](example_interfaces.go)

---

## 4.6 The Empty Interface: `any` (formerly `interface{}`)

An empty interface has no methods, so **every type satisfies it**:

```go
func printAnything(v any) {
    fmt.Println(v)
}

printAnything(42)
printAnything("hello")
printAnything([]int{1, 2, 3})
```

Python equivalent: no type hint (or `Any` from `typing`):
```python
def print_anything(v):  # accepts anything
    print(v)
```

**Use sparingly.** Prefer specific interfaces when possible.

---

## 4.7 Type Assertions and Type Switches

When you have an `any` or interface value, you can check the underlying type:

### Type Assertion

```go
var val any = "hello"

// Assert it's a string
s, ok := val.(string)
if ok {
    fmt.Println("It's a string:", s)
}

// Dangerous: without ok — panics if wrong type!
// s := val.(string)  // ok if correct, panics if not
```

Python equivalent:
```python
val = "hello"
if isinstance(val, str):
    print("It's a string:", val)
```

### Type Switch

```go
func describe(val any) string {
    switch v := val.(type) {
    case int:
        return fmt.Sprintf("integer: %d", v)
    case string:
        return fmt.Sprintf("string: %q", v)
    case bool:
        return fmt.Sprintf("boolean: %t", v)
    default:
        return fmt.Sprintf("unknown: %v", v)
    }
}
```

Python equivalent:
```python
def describe(val):
    match val:
        case int():
            return f"integer: {val}"
        case str():
            return f"string: {val}"
        case bool():
            return f"boolean: {val}"
        case _:
            return f"unknown: {val}"
```

> See: [example_type_assertions.go](example_type_assertions.go)

---

## 4.8 Common Standard Library Interfaces

Go's stdlib uses interfaces everywhere. Knowing these gives you superpowers:

| Interface | Methods | Python Equivalent |
|-----------|---------|------------------|
| `fmt.Stringer` | `String() string` | `__str__()` |
| `error` | `Error() string` | `Exception` (with `__str__`) |
| `io.Reader` | `Read([]byte) (int, error)` | file-like `.read()` |
| `io.Writer` | `Write([]byte) (int, error)` | file-like `.write()` |
| `sort.Interface` | `Len()`, `Less()`, `Swap()` | `__lt__()` for sorting |

```go
// Implement Stringer to control how your type prints
type User struct {
    Name string
    Age  int
}

func (u User) String() string {
    return fmt.Sprintf("%s (age %d)", u.Name, u.Age)
}

// Now fmt.Println(user) uses your String() method
```

---

## 4.9 Quick Reference

```
Struct:     type T struct { Field Type }     → data container
Method:     func (t T) Name() RetType        → behavior on T
Interface:  type I interface { Method() }    → contract
Embedding:  type B struct { A }              → composition
any:        interface{} / any                → accepts everything
Assert:     v, ok := x.(Type)               → check concrete type
Switch:     switch v := x.(type) { }         → branch on type
```

---

## Exercises

### Exercise 1: Shape System
Create `Circle` and `Rectangle` types that both implement a `Shape` interface with `Area()` and `Perimeter()`. Write a function that takes `[]Shape` and prints info about each.

### Exercise 2: Stringer
Create a `Duration` struct with `Hours`, `Minutes`, `Seconds` int fields. Implement `fmt.Stringer` to print like `"2h 30m 15s"`.

### Exercise 3: Embedding
Create an `Employee` struct that embeds a `Person` struct. Add a `Salary` field. Show that you can access `Person` fields directly on `Employee`.

---

> **Next → [Lesson 5: Collections](../05_collections/lesson.md)** — Slices, maps, and strings — Go's core data structures.
