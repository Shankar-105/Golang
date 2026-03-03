//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"log"
)

// ──────────────────────────────────────────────────────────────
// Struct tags control JSON field names, omission, etc.
// ──────────────────────────────────────────────────────────────

// User demonstrates basic struct tags.
// Python equivalent: a dataclass with field aliases.
//
//	@dataclass
//	class User:
//	    name: str
//	    email: str
//	    age: int
type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// Product shows omitempty and the "-" skip tag.
type Product struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Price       float64  `json:"price"`
	Description string   `json:"description,omitempty"` // omit when ""
	InternalSKU string   `json:"-"`                     // NEVER in JSON
	Tags        []string `json:"tags,omitempty"`         // omit when nil
}

// Address shows nested structs — they embed naturally in JSON.
type Address struct {
	Street string `json:"street"`
	City   string `json:"city"`
	State  string `json:"state"`
	Zip    string `json:"zip"`
}

// Employee shows nested struct + pointer (nil = omit when omitempty).
type Employee struct {
	Name    string   `json:"name"`
	Role    string   `json:"role"`
	Address *Address `json:"address,omitempty"` // pointer so nil = omitted
}

func main() {
	fmt.Println("═══ 1. Basic Marshal (struct → JSON) ═══")
	marshalBasic()

	fmt.Println("\n═══ 2. Pretty Print with MarshalIndent ═══")
	marshalPretty()

	fmt.Println("\n═══ 3. Basic Unmarshal (JSON → struct) ═══")
	unmarshalBasic()

	fmt.Println("\n═══ 4. omitempty Behavior ═══")
	omitemptyDemo()

	fmt.Println("\n═══ 5. Nested Structs ═══")
	nestedDemo()

	fmt.Println("\n═══ 6. Unmarshal into map[string]any (dynamic JSON) ═══")
	dynamicJSON()

	fmt.Println("\n═══ 7. JSON Arrays ═══")
	jsonArrays()

	fmt.Println("\n═══ 8. Unknown Fields Are Silently Ignored ═══")
	unknownFields()
}

// ──── 1. Basic Marshal ──────────────────────────────────────
func marshalBasic() {
	u := User{
		Name:  "Alice",
		Email: "alice@example.com",
		Age:   30,
	}

	// json.Marshal returns []byte
	data, err := json.Marshal(u)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(data))
	// Output: {"name":"Alice","email":"alice@example.com","age":30}

	// Python equivalent:
	// json.dumps({"name": "Alice", "email": "alice@example.com", "age": 30})
}

// ──── 2. Pretty Print ──────────────────────────────────────
func marshalPretty() {
	u := User{Name: "Bob", Email: "bob@example.com", Age: 25}

	// MarshalIndent adds newlines + indentation
	// Like Python's json.dumps(obj, indent=2)
	data, err := json.MarshalIndent(u, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(data))
	// Output:
	// {
	//   "name": "Bob",
	//   "email": "bob@example.com",
	//   "age": 25
	// }
}

// ──── 3. Basic Unmarshal ────────────────────────────────────
func unmarshalBasic() {
	raw := `{"name":"Charlie","email":"charlie@ex.com","age":35}`

	var u User
	// Note: you MUST pass a pointer (&u) so Unmarshal can fill the struct.
	// Python's json.loads() returns a new dict — Go mutates the target.
	err := json.Unmarshal([]byte(raw), &u)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Name: %s, Email: %s, Age: %d\n", u.Name, u.Email, u.Age)
}

// ──── 4. omitempty Demo ─────────────────────────────────────
func omitemptyDemo() {
	// Full product — all fields present
	full := Product{
		ID:          1,
		Name:        "Laptop",
		Price:       999.99,
		Description: "A nice laptop",
		InternalSKU: "SKU-12345", // this will NEVER appear in JSON
		Tags:        []string{"electronics", "sale"},
	}
	printJSON("Full product", full)

	// Sparse product — zero/nil fields omitted thanks to omitempty
	sparse := Product{
		ID:          2,
		Name:        "Mouse",
		Price:       29.99,
		InternalSKU: "SKU-99999",
		// Description is "" → omitted
		// Tags is nil → omitted
	}
	printJSON("Sparse product", sparse)
}

// ──── 5. Nested Structs ─────────────────────────────────────
func nestedDemo() {
	// Employee WITH address
	withAddr := Employee{
		Name: "Diana",
		Role: "Engineer",
		Address: &Address{
			Street: "123 Main St",
			City:   "Portland",
			State:  "OR",
			Zip:    "97201",
		},
	}
	printJSON("Employee with address", withAddr)

	// Employee WITHOUT address (pointer is nil → omitted)
	noAddr := Employee{
		Name: "Eve",
		Role: "Designer",
		// Address is nil → omitted thanks to omitempty on the pointer
	}
	printJSON("Employee without address", noAddr)
}

// ──── 6. Dynamic JSON (map[string]any) ──────────────────────
func dynamicJSON() {
	// When you don't know the structure ahead of time,
	// unmarshal into map[string]any (like Python dict)
	raw := `{
		"event": "purchase",
		"amount": 49.99,
		"items": ["book", "pen"],
		"customer": {"name": "Frank", "vip": true}
	}`

	var result map[string]any
	err := json.Unmarshal([]byte(raw), &result)
	if err != nil {
		log.Fatal(err)
	}

	// Accessing fields requires type assertions (unlike Python where dict["key"] just works)
	event := result["event"].(string)
	amount := result["amount"].(float64) // JSON numbers are always float64 in any
	fmt.Printf("Event: %s, Amount: $%.2f\n", event, amount)

	// Nested map
	customer := result["customer"].(map[string]any)
	fmt.Printf("Customer: %s, VIP: %v\n", customer["name"], customer["vip"])

	// Slice — JSON arrays become []any
	items := result["items"].([]any)
	for i, item := range items {
		fmt.Printf("  Item %d: %s\n", i, item.(string))
	}
}

// ──── 7. JSON Arrays ────────────────────────────────────────
func jsonArrays() {
	// Marshal a slice of structs → JSON array
	users := []User{
		{Name: "Alice", Email: "alice@ex.com", Age: 30},
		{Name: "Bob", Email: "bob@ex.com", Age: 25},
		{Name: "Charlie", Email: "charlie@ex.com", Age: 35},
	}

	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Marshalled array:")
	fmt.Println(string(data))

	// Unmarshal JSON array → slice of structs
	raw := `[{"name":"X","email":"x@ex.com","age":1},{"name":"Y","email":"y@ex.com","age":2}]`
	var parsed []User
	err = json.Unmarshal([]byte(raw), &parsed)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nUnmarshalled %d users: %+v\n", len(parsed), parsed)
}

// ──── 8. Unknown Fields Are Silently Ignored ────────────────
func unknownFields() {
	// Go ignores JSON keys that don't match any struct field
	// Python: you'd get all keys in the dict regardless
	raw := `{
		"name": "Grace",
		"email": "grace@ex.com",
		"age": 28,
		"nickname": "Gracie",
		"phone": "555-0123",
		"favorite_color": "blue"
	}`

	var u User
	err := json.Unmarshal([]byte(raw), &u)
	if err != nil {
		log.Fatal(err)
	}

	// Only name, email, age were captured — the rest silently dropped
	fmt.Printf("Parsed: %+v\n", u)
	fmt.Println("(nickname, phone, favorite_color were silently ignored)")
}

// ──── Helper ────────────────────────────────────────────────
func printJSON(label string, v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s:\n%s\n\n", label, string(data))
}
