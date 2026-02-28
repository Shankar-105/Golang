//go:build ignore

package main

import "fmt"

// ============================================
// Structs: Go's answer to Python classes
// Data containers with named fields
// ============================================

// User is a struct — like a Python dataclass
type User struct {
	Name  string
	Email string
	Age   int
}

// Address shows nested structs
type Address struct {
	Street string
	City   string
	State  string
	Zip    string
}

// Employee shows struct composition (fields of struct types)
type Employee struct {
	User            // embedded — promotes fields (like inheritance)
	Address Address // named field — must access as e.Address.City
	Salary  float64
}

func main() {
	// ============================================
	// Example 1: Creating structs
	// ============================================
	fmt.Println("=== Example 1: Creating structs ===")

	// Using named fields (preferred — like keyword args)
	u1 := User{Name: "Alice", Email: "alice@go.dev", Age: 25}
	fmt.Printf("  u1: %+v\n", u1) // %+v prints field names

	// Zero value — all fields are zero-valued
	var u2 User
	fmt.Printf("  u2 (zero): %+v\n", u2) // {Name: Email: Age:0}

	// Pointer to struct
	u3 := &User{Name: "Charlie", Age: 30}
	fmt.Printf("  u3 (pointer): %+v\n", *u3)

	// Partial initialization — unset fields get zero values
	u4 := User{Name: "Dave"}
	fmt.Printf("  u4 (partial): %+v\n", u4) // Email:"", Age:0

	// ============================================
	// Example 2: Accessing and modifying fields
	// ============================================
	fmt.Println("\n=== Example 2: Field access ===")

	u := User{Name: "Eve", Email: "eve@go.dev", Age: 28}
	fmt.Println("  Name:", u.Name)
	fmt.Println("  Age:", u.Age)

	u.Age = 29 // direct modification
	fmt.Println("  After birthday:", u.Age)

	// Pointer access — Go auto-dereferences
	p := &u
	fmt.Println("  Via pointer:", p.Name) // same as (*p).Name

	// ============================================
	// Example 3: Methods on structs
	// ============================================
	fmt.Println("\n=== Example 3: Methods ===")

	u5 := User{Name: "Frank", Email: "frank@go.dev", Age: 35}
	fmt.Println("  Greeting:", u5.Greet())

	u5.SetAge(36) // pointer receiver — modifies original
	fmt.Println("  After SetAge(36):", u5.Age)

	// ============================================
	// Example 4: Struct comparison
	// ============================================
	fmt.Println("\n=== Example 4: Comparison ===")

	a := User{Name: "Alice", Email: "a@b.c", Age: 25}
	b := User{Name: "Alice", Email: "a@b.c", Age: 25}
	c := User{Name: "Bob", Email: "a@b.c", Age: 25}

	fmt.Println("  a == b:", a == b) // true — all fields match
	fmt.Println("  a == c:", a == c) // false — Name differs

	// ============================================
	// Example 5: Anonymous structs (quick one-offs)
	// ============================================
	fmt.Println("\n=== Example 5: Anonymous structs ===")

	// Useful for quick, one-time data structures
	config := struct {
		Host string
		Port int
	}{
		Host: "localhost",
		Port: 8080,
	}
	fmt.Printf("  Config: %s:%d\n", config.Host, config.Port)

	// ============================================
	// Example 6: Struct embedding (composition)
	// ============================================
	fmt.Println("\n=== Example 6: Embedding ===")

	emp := Employee{
		User:    User{Name: "Grace", Email: "grace@co.dev", Age: 30},
		Address: Address{City: "Seattle", State: "WA"},
		Salary:  95000,
	}

	// Embedded User fields are "promoted" — accessible directly
	fmt.Println("  emp.Name:", emp.Name)       // from User (promoted)
	fmt.Println("  emp.Email:", emp.Email)     // from User (promoted)
	fmt.Println("  emp.Greet():", emp.Greet()) // User's method is promoted too!

	// Address is a named field — not promoted
	fmt.Println("  emp.Address.City:", emp.Address.City)
	fmt.Println("  emp.Salary:", emp.Salary)
}

// Greet is a value receiver method — doesn't modify User
func (u User) Greet() string {
	return fmt.Sprintf("Hi, I'm %s (%d years old)", u.Name, u.Age)
}

// SetAge is a pointer receiver method — modifies the User
func (u *User) SetAge(age int) {
	u.Age = age
}
