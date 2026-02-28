//go:build ignore

package main

import "fmt"

// ============================================
// Type assertions and type switches
//
// When you have an interface value (like `any`),
// you can check what concrete type is inside.
//
// Python: isinstance(val, int), type(val), match/case
// Go:    val.(int), val.(type)
// ============================================

func main() {
	// ============================================
	// Example 1: Basic type assertion
	// ============================================
	fmt.Println("=== Example 1: Type assertion ===")

	var val any = "hello, world"

	// Safe assertion with ok check (comma-ok pattern)
	s, ok := val.(string)
	fmt.Printf("  val.(string) → %q, ok=%t\n", s, ok)

	n, ok := val.(int)
	fmt.Printf("  val.(int)    → %d, ok=%t\n", n, ok) // 0, false

	// Dangerous assertion (panics if wrong):
	// n := val.(int) // would panic!

	// ============================================
	// Example 2: Type switch
	// ============================================
	fmt.Println("\n=== Example 2: Type switch ===")

	testValues := []any{42, "hello", 3.14, true, nil, []int{1, 2, 3}}

	for _, val := range testValues {
		fmt.Printf("  %-20v → %s\n", val, describe(val))
	}

	// ============================================
	// Example 3: Interface type assertion
	// ============================================
	fmt.Println("\n=== Example 3: Checking interface satisfaction ===")

	animals := []any{
		Dog{Name: "Rex"},
		Cat{Name: "Whiskers"},
		Fish{Name: "Nemo"},
	}

	for _, a := range animals {
		fmt.Printf("  %v: ", a)

		// Check if it satisfies the Speaker interface
		if speaker, ok := a.(Speaker); ok {
			fmt.Printf("can speak → %q", speaker.Speak())
		} else {
			fmt.Printf("cannot speak")
		}

		// Check if it can swim
		if swimmer, ok := a.(Swimmer); ok {
			fmt.Printf(", can swim → %q", swimmer.Swim())
		}
		fmt.Println()
	}

	// ============================================
	// Example 4: Practical example — process mixed data
	// ============================================
	fmt.Println("\n=== Example 4: Processing mixed data ===")

	data := []any{
		map[string]any{"name": "Alice", "age": 25},
		"just a string",
		42,
		[]string{"a", "b", "c"},
	}

	for i, item := range data {
		fmt.Printf("  Item %d: ", i)
		processItem(item)
	}
}

func describe(val any) string {
	switch v := val.(type) {
	case int:
		return fmt.Sprintf("int (value: %d)", v)
	case string:
		return fmt.Sprintf("string (len: %d)", len(v))
	case float64:
		return fmt.Sprintf("float64 (value: %.2f)", v)
	case bool:
		return fmt.Sprintf("bool (value: %t)", v)
	case nil:
		return "nil"
	default:
		return fmt.Sprintf("other (%T)", v)
	}
}

// ---- Interfaces and types for Example 3 ----

type Speaker interface {
	Speak() string
}

type Swimmer interface {
	Swim() string
}

type Dog struct{ Name string }
type Cat struct{ Name string }
type Fish struct{ Name string }

func (d Dog) Speak() string  { return "Woof!" }
func (d Dog) String() string { return fmt.Sprintf("Dog(%s)", d.Name) }

func (c Cat) Speak() string  { return "Meow!" }
func (c Cat) String() string { return fmt.Sprintf("Cat(%s)", c.Name) }

func (f Fish) Swim() string   { return "Swimming..." }
func (f Fish) String() string { return fmt.Sprintf("Fish(%s)", f.Name) }

// ---- Practical processor for Example 4 ----

func processItem(item any) {
	switch v := item.(type) {
	case map[string]any:
		fmt.Printf("Map with %d entries → ", len(v))
		for key, val := range v {
			fmt.Printf("%s=%v ", key, val)
		}
		fmt.Println()
	case string:
		fmt.Printf("String: %q\n", v)
	case int:
		fmt.Printf("Number: %d\n", v)
	case []string:
		fmt.Printf("String slice: %v\n", v)
	default:
		fmt.Printf("Unknown type: %T\n", v)
	}
}
