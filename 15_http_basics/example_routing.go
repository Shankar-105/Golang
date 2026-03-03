//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// ============================================
// Routing with Go 1.22+ ServeMux
//
// New in Go 1.22:
// - Method matching: "GET /path" only matches GET
// - Path parameters: "/users/{id}" captures "id"
// - Wildcards: "/files/{path...}" captures the rest
//
// Python equivalent:
//   @app.get("/users/{user_id}")
//   async def get_user(user_id: int): ...
//
// Run: go run example_routing.go
// ============================================

// In-memory "database"
var (
	users   = map[string]User{}
	usersMu sync.RWMutex
	nextID  = 1
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	mux := http.NewServeMux()

	// ---- Static routes ----
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "API v1 — try /api/users")
	})

	// ---- RESTful CRUD routes (Go 1.22+ syntax) ----
	mux.HandleFunc("GET /api/users", listUsersHandler)
	mux.HandleFunc("POST /api/users", createUserHandler)
	mux.HandleFunc("GET /api/users/{id}", getUserHandler)
	mux.HandleFunc("PUT /api/users/{id}", updateUserHandler)
	mux.HandleFunc("DELETE /api/users/{id}", deleteUserHandler)

	// ---- Seed some data ----
	seedData()

	fmt.Println("=== Routing Demo (Go 1.22+) ===")
	fmt.Println("Listening on http://localhost:8080")
	fmt.Println("")
	fmt.Println("Try these:")
	fmt.Println("  curl http://localhost:8080/api/users")
	fmt.Println("  curl http://localhost:8080/api/users/1")
	fmt.Println("  curl -X POST http://localhost:8080/api/users \\")
	fmt.Println("       -H 'Content-Type: application/json' \\")
	fmt.Println("       -d '{\"name\":\"Charlie\",\"email\":\"charlie@go.dev\"}'")
	fmt.Println("  curl -X PUT http://localhost:8080/api/users/1 \\")
	fmt.Println("       -H 'Content-Type: application/json' \\")
	fmt.Println("       -d '{\"name\":\"Alice Updated\",\"email\":\"alice2@go.dev\"}'")
	fmt.Println("  curl -X DELETE http://localhost:8080/api/users/2")

	log.Fatal(http.ListenAndServe(":8080", mux))
}

func seedData() {
	usersMu.Lock()
	defer usersMu.Unlock()

	users["1"] = User{ID: 1, Name: "Alice", Email: "alice@go.dev"}
	users["2"] = User{ID: 2, Name: "Bob", Email: "bob@go.dev"}
	nextID = 3
}

// GET /api/users — list all users
func listUsersHandler(w http.ResponseWriter, r *http.Request) {
	usersMu.RLock()
	defer usersMu.RUnlock()

	list := make([]User, 0, len(users))
	for _, u := range users {
		list = append(list, u)
	}

	writeJSON(w, http.StatusOK, list)
}

// POST /api/users — create a new user
func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	if input.Name == "" || input.Email == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "name and email are required",
		})
		return
	}

	usersMu.Lock()
	id := nextID
	nextID++
	idStr := fmt.Sprintf("%d", id)
	u := User{ID: id, Name: input.Name, Email: input.Email}
	users[idStr] = u
	usersMu.Unlock()

	writeJSON(w, http.StatusCreated, u)
}

// GET /api/users/{id} — get a single user
func getUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id") // Go 1.22+ path parameter

	usersMu.RLock()
	u, ok := users[id]
	usersMu.RUnlock()

	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": fmt.Sprintf("user %s not found", id),
		})
		return
	}

	writeJSON(w, http.StatusOK, u)
}

// PUT /api/users/{id} — update a user
func updateUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var input struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	usersMu.Lock()
	u, ok := users[id]
	if !ok {
		usersMu.Unlock()
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": fmt.Sprintf("user %s not found", id),
		})
		return
	}

	if input.Name != "" {
		u.Name = input.Name
	}
	if input.Email != "" {
		u.Email = input.Email
	}
	users[id] = u
	usersMu.Unlock()

	writeJSON(w, http.StatusOK, u)
}

// DELETE /api/users/{id} — delete a user
func deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	usersMu.Lock()
	_, ok := users[id]
	if !ok {
		usersMu.Unlock()
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": fmt.Sprintf("user %s not found", id),
		})
		return
	}
	delete(users, id)
	usersMu.Unlock()

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// writeJSON is a helper to send JSON responses
// You'll write helpers like this in every Go API
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
