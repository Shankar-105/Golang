//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

// ──────────────────────────────────────────────────────────────
// Custom JSON marshalling — implement json.Marshaler / json.Unmarshaler
//
// You need custom marshalling when:
// 1. You want enums to serialize as strings (not ints)
// 2. You need a custom date format
// 3. You want to flatten/restructure nested data
// 4. You need to handle polymorphic types
//
// Python equivalent: defining __dict__() or using custom json.JSONEncoder
// ──────────────────────────────────────────────────────────────

func main() {
	fmt.Println("═══ 1. Enum as String ═══")
	enumDemo()

	fmt.Println("\n═══ 2. Custom Date Format ═══")
	dateFormatDemo()

	fmt.Println("\n═══ 3. Duration as Human-Readable String ═══")
	durationDemo()

	fmt.Println("\n═══ 4. Flattening Nested JSON ═══")
	flattenDemo()

	fmt.Println("\n═══ 5. Polymorphic JSON (different types in same field) ═══")
	polymorphicDemo()
}

// ════════════════════════════════════════════════════════════════
// 1. Enum as String
// ════════════════════════════════════════════════════════════════

type Priority int

const (
	PriorityLow Priority = iota
	PriorityMedium
	PriorityHigh
	PriorityCritical
)

var priorityNames = map[Priority]string{
	PriorityLow:      "low",
	PriorityMedium:   "medium",
	PriorityHigh:     "high",
	PriorityCritical: "critical",
}

var priorityValues = map[string]Priority{
	"low":      PriorityLow,
	"medium":   PriorityMedium,
	"high":     PriorityHigh,
	"critical": PriorityCritical,
}

func (p Priority) MarshalJSON() ([]byte, error) {
	name, ok := priorityNames[p]
	if !ok {
		return nil, fmt.Errorf("unknown priority: %d", p)
	}
	return json.Marshal(name)
}

func (p *Priority) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	val, ok := priorityValues[strings.ToLower(s)]
	if !ok {
		return fmt.Errorf("unknown priority: %q", s)
	}
	*p = val
	return nil
}

// Stringer interface for fmt.Printf
func (p Priority) String() string {
	return priorityNames[p]
}

type Ticket struct {
	ID       int      `json:"id"`
	Title    string   `json:"title"`
	Priority Priority `json:"priority"`
}

func enumDemo() {
	// Marshal: int enum → string
	ticket := Ticket{ID: 1, Title: "Fix login bug", Priority: PriorityHigh}
	data, _ := json.MarshalIndent(ticket, "", "  ")
	fmt.Println("  Marshalled:")
	fmt.Println(" ", string(data))

	// Unmarshal: string → int enum
	raw := `{"id":2,"title":"Update docs","priority":"low"}`
	var t2 Ticket
	json.Unmarshal([]byte(raw), &t2)
	fmt.Printf("\n  Unmarshalled: ID=%d, Title=%q, Priority=%s (int=%d)\n",
		t2.ID, t2.Title, t2.Priority, t2.Priority)
}

// ════════════════════════════════════════════════════════════════
// 2. Custom Date Format
// ════════════════════════════════════════════════════════════════

// By default, time.Time uses RFC3339 format: "2024-01-15T10:30:00Z"
// But many APIs use different formats.

// DateOnly formats as "2024-01-15"
type DateOnly time.Time

func (d DateOnly) MarshalJSON() ([]byte, error) {
	t := time.Time(d)
	return json.Marshal(t.Format("2006-01-02"))
}

func (d *DateOnly) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return fmt.Errorf("invalid date format %q, expected YYYY-MM-DD", s)
	}
	*d = DateOnly(t)
	return nil
}

// USDate formats as "01/15/2024"
type USDate time.Time

func (d USDate) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(d).Format("01/02/2006"))
}

func (d *USDate) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	t, err := time.Parse("01/02/2006", s)
	if err != nil {
		return err
	}
	*d = USDate(t)
	return nil
}

type Event struct {
	Name      string   `json:"name"`
	StartDate DateOnly `json:"start_date"` // "2024-06-15"
	PurchasedAt USDate `json:"purchased_at"` // "06/15/2024"
}

func dateFormatDemo() {
	event := Event{
		Name:        "Go Conference",
		StartDate:   DateOnly(time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)),
		PurchasedAt: USDate(time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)),
	}

	data, _ := json.MarshalIndent(event, "", "  ")
	fmt.Println("  Marshalled with custom date formats:")
	fmt.Println(" ", string(data))

	// Unmarshal
	raw := `{"name":"Workshop","start_date":"2025-09-20","purchased_at":"09/01/2025"}`
	var e2 Event
	json.Unmarshal([]byte(raw), &e2)
	fmt.Printf("\n  Unmarshalled: %s, starts %s\n",
		e2.Name, time.Time(e2.StartDate).Format("Jan 2, 2006"))
}

// ════════════════════════════════════════════════════════════════
// 3. Duration as Human-Readable String
// ════════════════════════════════════════════════════════════════

// Go's time.Duration marshals as nanoseconds (an int64).
// Let's make it human-readable like "5m30s" or "2h15m".

type HumanDuration time.Duration

func (d HumanDuration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String()) // "5m30s"
}

func (d *HumanDuration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = HumanDuration(dur)
	return nil
}

type Task struct {
	Name     string        `json:"name"`
	Timeout  HumanDuration `json:"timeout"`
	Interval HumanDuration `json:"interval"`
}

func durationDemo() {
	task := Task{
		Name:     "health-check",
		Timeout:  HumanDuration(30 * time.Second),
		Interval: HumanDuration(5 * time.Minute),
	}

	data, _ := json.MarshalIndent(task, "", "  ")
	fmt.Println("  Marshalled:")
	fmt.Println(" ", string(data))
	// {"name":"health-check","timeout":"30s","interval":"5m0s"}

	raw := `{"name":"backup","timeout":"1h30m","interval":"24h0m0s"}`
	var t2 Task
	json.Unmarshal([]byte(raw), &t2)
	fmt.Printf("\n  Unmarshalled: %s (timeout: %v, interval: %v)\n",
		t2.Name, time.Duration(t2.Timeout), time.Duration(t2.Interval))
}

// ════════════════════════════════════════════════════════════════
// 4. Flattening Nested JSON
// ════════════════════════════════════════════════════════════════

// External API returns:   {"data": {"user": {"name": "Alice", "id": 42}}}
// We want to unmarshal directly into a flat struct.

type APIUser struct {
	Name string
	ID   int
}

func (u *APIUser) UnmarshalJSON(data []byte) error {
	// Define the nested structure we expect
	var envelope struct {
		Data struct {
			User struct {
				Name string `json:"name"`
				ID   int    `json:"id"`
			} `json:"user"`
		} `json:"data"`
	}

	if err := json.Unmarshal(data, &envelope); err != nil {
		return err
	}

	u.Name = envelope.Data.User.Name
	u.ID = envelope.Data.User.ID
	return nil
}

func (u APIUser) MarshalJSON() ([]byte, error) {
	// Re-create the nested structure
	return json.Marshal(struct {
		Data struct {
			User struct {
				Name string `json:"name"`
				ID   int    `json:"id"`
			} `json:"user"`
		} `json:"data"`
	}{
		Data: struct {
			User struct {
				Name string `json:"name"`
				ID   int    `json:"id"`
			} `json:"user"`
		}{
			User: struct {
				Name string `json:"name"`
				ID   int    `json:"id"`
			}{Name: u.Name, ID: u.ID},
		},
	})
}

func flattenDemo() {
	raw := `{"data":{"user":{"name":"Alice","id":42}}}`

	var user APIUser
	err := json.Unmarshal([]byte(raw), &user)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  Flattened: Name=%s, ID=%d\n", user.Name, user.ID)

	// Re-marshal back to nested
	data, _ := json.Marshal(user)
	fmt.Printf("  Re-marshalled: %s\n", string(data))
}

// ════════════════════════════════════════════════════════════════
// 5. Polymorphic JSON
// ════════════════════════════════════════════════════════════════

// Different notification types with different payloads.
// Python: you'd use Union[EmailNotif, SMSNotif, PushNotif] with Pydantic.

type Notification struct {
	Type    string `json:"type"`
	Email   *EmailPayload `json:"email,omitempty"`
	SMS     *SMSPayload   `json:"sms,omitempty"`
	Push    *PushPayload  `json:"push,omitempty"`
}

type EmailPayload struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type SMSPayload struct {
	Phone   string `json:"phone"`
	Message string `json:"message"`
}

type PushPayload struct {
	DeviceID string `json:"device_id"`
	Title    string `json:"title"`
	Body     string `json:"body"`
}

func polymorphicDemo() {
	notifications := []Notification{
		{
			Type: "email",
			Email: &EmailPayload{
				To:      "alice@example.com",
				Subject: "Welcome!",
				Body:    "Thanks for signing up.",
			},
		},
		{
			Type: "sms",
			SMS: &SMSPayload{
				Phone:   "+1-555-0123",
				Message: "Your code is 123456",
			},
		},
		{
			Type: "push",
			Push: &PushPayload{
				DeviceID: "device-abc-123",
				Title:    "New message",
				Body:     "You have a new message from Bob",
			},
		},
	}

	for _, n := range notifications {
		data, _ := json.MarshalIndent(n, "  ", "  ")
		fmt.Printf("  %s notification:\n  %s\n\n", n.Type, string(data))
	}

	// Unmarshal and dispatch
	raw := `{"type":"email","email":{"to":"bob@ex.com","subject":"Hi","body":"Hello!"}}`
	var notif Notification
	json.Unmarshal([]byte(raw), &notif)

	switch notif.Type {
	case "email":
		fmt.Printf("  Dispatch email to %s: %s\n", notif.Email.To, notif.Email.Subject)
	case "sms":
		fmt.Printf("  Dispatch SMS to %s\n", notif.SMS.Phone)
	case "push":
		fmt.Printf("  Dispatch push to %s\n", notif.Push.DeviceID)
	}
}
