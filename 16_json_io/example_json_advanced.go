//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// ──────────────────────────────────────────────────────────────
// Advanced JSON: RawMessage, custom marshal, embedding, ",string"
// ──────────────────────────────────────────────────────────────

// ──── 1. json.RawMessage — Delay Parsing ────────────────────
// Useful for API envelopes where the payload type depends on another field.
// Python equivalent: keeping a sub-dict as-is until you know which model to use.

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"` // raw bytes, not parsed yet
}

type UserCreated struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type OrderPlaced struct {
	OrderID string  `json:"order_id"`
	Total   float64 `json:"total"`
}

func rawMessageDemo() {
	fmt.Println("═══ 1. json.RawMessage — Delay Parsing ═══")

	events := []string{
		`{"type":"user_created","payload":{"name":"Alice","email":"alice@ex.com"}}`,
		`{"type":"order_placed","payload":{"order_id":"ORD-123","total":99.99}}`,
	}

	for _, raw := range events {
		var event Event
		if err := json.Unmarshal([]byte(raw), &event); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Event type: %s\n", event.Type)

		// Now parse the payload based on the type
		switch event.Type {
		case "user_created":
			var uc UserCreated
			json.Unmarshal(event.Payload, &uc)
			fmt.Printf("  → User: %s (%s)\n", uc.Name, uc.Email)

		case "order_placed":
			var op OrderPlaced
			json.Unmarshal(event.Payload, &op)
			fmt.Printf("  → Order: %s, Total: $%.2f\n", op.OrderID, op.Total)
		}
	}
}

// ──── 2. Custom MarshalJSON / UnmarshalJSON ─────────────────
// Implement the json.Marshaler and json.Unmarshaler interfaces
// to control how a type serializes.

// Status is an int enum that serializes as a string.
type Status int

const (
	StatusActive Status = iota
	StatusInactive
	StatusSuspended
)

var statusNames = map[Status]string{
	StatusActive:    "active",
	StatusInactive:  "inactive",
	StatusSuspended: "suspended",
}

var statusValues = map[string]Status{
	"active":    StatusActive,
	"inactive":  StatusInactive,
	"suspended": StatusSuspended,
}

func (s Status) MarshalJSON() ([]byte, error) {
	name, ok := statusNames[s]
	if !ok {
		return nil, fmt.Errorf("unknown status: %d", s)
	}
	return json.Marshal(name) // wraps in quotes → "active"
}

func (s *Status) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return err
	}
	val, ok := statusValues[name]
	if !ok {
		return fmt.Errorf("unknown status: %q", name)
	}
	*s = val
	return nil
}

// Account uses Status — it will serialize as "active"/"inactive"/etc.
type Account struct {
	Name   string `json:"name"`
	Status Status `json:"status"`
}

func customMarshalDemo() {
	fmt.Println("\n═══ 2. Custom MarshalJSON / UnmarshalJSON ═══")

	// Marshal: int enum → string in JSON
	acc := Account{Name: "Alice", Status: StatusActive}
	data, _ := json.MarshalIndent(acc, "", "  ")
	fmt.Println("Marshalled:")
	fmt.Println(string(data))
	// Output: {"name":"Alice","status":"active"}

	// Unmarshal: string in JSON → int enum
	raw := `{"name":"Bob","status":"suspended"}`
	var acc2 Account
	json.Unmarshal([]byte(raw), &acc2)
	fmt.Printf("\nUnmarshalled: %+v (Status int value: %d)\n", acc2, acc2.Status)
}

// ──── 3. Custom Time Format ─────────────────────────────────
// Go's time.Time marshals to RFC3339 by default.
// What if your API uses "2006-01-02" (date only)?

type DateOnly struct {
	time.Time
}

func (d DateOnly) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Format("2006-01-02"))
}

func (d *DateOnly) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	d.Time = t
	return nil
}

type Contract struct {
	Client    string   `json:"client"`
	StartDate DateOnly `json:"start_date"`
	EndDate   DateOnly `json:"end_date"`
}

func customTimeDemo() {
	fmt.Println("\n═══ 3. Custom Time Format ═══")

	c := Contract{
		Client:    "Acme Corp",
		StartDate: DateOnly{time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
		EndDate:   DateOnly{time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)},
	}
	data, _ := json.MarshalIndent(c, "", "  ")
	fmt.Println("Marshalled with date-only format:")
	fmt.Println(string(data))

	// Unmarshal back
	raw := `{"client":"BigCo","start_date":"2025-06-01","end_date":"2025-12-31"}`
	var c2 Contract
	json.Unmarshal([]byte(raw), &c2)
	fmt.Printf("\nUnmarshalled: %s from %s to %s\n",
		c2.Client,
		c2.StartDate.Format("Jan 2, 2006"),
		c2.EndDate.Format("Jan 2, 2006"))
}

// ──── 4. Struct Embedding in JSON ───────────────────────────
// Embedded structs "flatten" their fields into the parent JSON.
// Like Python inheritance where child gets all parent fields.

type BaseModel struct {
	ID        int    `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type Article struct {
	BaseModel        // embedded — fields appear at the top level in JSON
	Title     string `json:"title"`
	Body      string `json:"body"`
	Author    string `json:"author"`
}

func embeddingDemo() {
	fmt.Println("\n═══ 4. Struct Embedding in JSON ═══")

	a := Article{
		BaseModel: BaseModel{
			ID:        42,
			CreatedAt: "2024-01-15T10:00:00Z",
			UpdatedAt: "2024-03-20T14:30:00Z",
		},
		Title:  "Go JSON Guide",
		Body:   "Learn JSON in Go...",
		Author: "Alice",
	}

	data, _ := json.MarshalIndent(a, "", "  ")
	fmt.Println(string(data))
	// Note: id, created_at, updated_at appear at TOP LEVEL, not nested!
}

// ──── 5. The ",string" Tag ──────────────────────────────────
// Some APIs send numbers/bools as strings in JSON.
// The ",string" tag handles conversion automatically.

type APIResponse struct {
	UserID  int     `json:"user_id,string"`  // "123" → 123
	Balance float64 `json:"balance,string"`  // "99.50" → 99.50
	Active  bool    `json:"active,string"`   // "true" → true
}

func stringTagDemo() {
	fmt.Println("\n═══ 5. The \",string\" Tag ═══")

	raw := `{
		"user_id": "42",
		"balance": "1234.56",
		"active": "true"
	}`

	var resp APIResponse
	err := json.Unmarshal([]byte(raw), &resp)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("UserID: %d (type: int)\n", resp.UserID)
	fmt.Printf("Balance: %.2f (type: float64)\n", resp.Balance)
	fmt.Printf("Active: %v (type: bool)\n", resp.Active)

	// Re-marshal → numbers become strings again
	data, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println("\nRe-marshalled:")
	fmt.Println(string(data))
}

// ──── 6. Validating JSON Before Processing ──────────────────
func validationDemo() {
	fmt.Println("\n═══ 6. JSON Validation ═══")

	inputs := []string{
		`{"name":"Alice","age":30}`,       // valid
		`{"name":"Bob","age":}`,            // invalid
		`not json at all`,                  // invalid
		`[1, 2, 3]`,                        // valid JSON array
		`{"nested":{"deep":{"value":42}}}`, // valid complex
	}

	for _, input := range inputs {
		if json.Valid([]byte(input)) {
			fmt.Printf("  ✓ Valid:   %.40s...\n", input)
		} else {
			fmt.Printf("  ✗ Invalid: %.40s...\n", input)
		}
	}
}

func main() {
	rawMessageDemo()
	customMarshalDemo()
	customTimeDemo()
	embeddingDemo()
	stringTagDemo()
	validationDemo()
}
