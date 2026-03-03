//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
)

// ──────────────────────────────────────────────────────────────
// HTTP Handler Testing with httptest
//
// httptest lets you test HTTP handlers without starting a real server.
// Python equivalent: Flask's test_client() or FastAPI's TestClient.
//
//   # Python (FastAPI)
//   from fastapi.testclient import TestClient
//   client = TestClient(app)
//   response = client.get("/health")
//   assert response.status_code == 200
//
// This is a runnable demo that shows the httptest patterns.
// In real code, these would be in *_test.go files.
// ──────────────────────────────────────────────────────────────

// ──── Handlers to Test ──────────────────────────────────────

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var nextID = 1

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if input.Name == "" || input.Email == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "name and email are required",
		})
		return
	}

	user := User{ID: nextID, Name: input.Name, Email: input.Email}
	nextID++

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func greetHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "World"
	}
	fmt.Fprintf(w, "Hello, %s!", name)
}

func main() {
	fmt.Println("═══ httptest Demo — Testing HTTP Handlers ═══")
	fmt.Println("(These patterns would normally be in *_test.go files)")

	fmt.Println("── 1. Test Health Endpoint ──")
	testHealthEndpoint()

	fmt.Println("\n── 2. Test Create User (Happy Path) ──")
	testCreateUserHappy()

	fmt.Println("\n── 3. Test Create User (Validation Error) ──")
	testCreateUserValidation()

	fmt.Println("\n── 4. Test Create User (Wrong Method) ──")
	testCreateUserWrongMethod()

	fmt.Println("\n── 5. Test Query Parameters ──")
	testQueryParams()

	fmt.Println("\n── 6. Test with Real Server (httptest.NewServer) ──")
	testWithServer()

	fmt.Println("\n── 7. Test Headers ──")
	testHeaders()
}

// ──── 1. Test Health Endpoint ───────────────────────────────
func testHealthEndpoint() {
	// Create a fake request
	req := httptest.NewRequest("GET", "/health", nil)

	// Create a ResponseRecorder (fake ResponseWriter)
	w := httptest.NewRecorder()

	// Call the handler directly
	healthHandler(w, req)

	// Check the result
	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	fmt.Printf("  Status: %d\n", resp.StatusCode)
	fmt.Printf("  Content-Type: %s\n", resp.Header.Get("Content-Type"))
	fmt.Printf("  Body: %s", string(body))

	// In a real test:
	// if resp.StatusCode != http.StatusOK {
	//     t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	// }
}

// ──── 2. Test Create User (Happy Path) ──────────────────────
func testCreateUserHappy() {
	body := `{"name":"Alice","email":"alice@example.com"}`
	req := httptest.NewRequest("POST", "/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	createUserHandler(w, req)

	resp := w.Result()
	respBody, _ := io.ReadAll(resp.Body)

	fmt.Printf("  Status: %d (want 201)\n", resp.StatusCode)
	fmt.Printf("  Body: %s", string(respBody))

	// Decode and verify
	var user User
	json.Unmarshal(respBody, &user)
	fmt.Printf("  Created user ID: %d, Name: %s\n", user.ID, user.Name)
}

// ──── 3. Test Validation Error ──────────────────────────────
func testCreateUserValidation() {
	// Missing email
	body := `{"name":"Bob"}`
	req := httptest.NewRequest("POST", "/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	createUserHandler(w, req)

	resp := w.Result()
	respBody, _ := io.ReadAll(resp.Body)

	fmt.Printf("  Status: %d (want 400)\n", resp.StatusCode)
	fmt.Printf("  Body: %s", string(respBody))
}

// ──── 4. Test Wrong Method ──────────────────────────────────
func testCreateUserWrongMethod() {
	req := httptest.NewRequest("GET", "/users", nil)
	w := httptest.NewRecorder()
	createUserHandler(w, req)

	resp := w.Result()
	fmt.Printf("  Status: %d (want 405)\n", resp.StatusCode)
}

// ──── 5. Test Query Parameters ──────────────────────────────
func testQueryParams() {
	// With name parameter
	req1 := httptest.NewRequest("GET", "/greet?name=Alice", nil)
	w1 := httptest.NewRecorder()
	greetHandler(w1, req1)
	fmt.Printf("  With name: %s\n", w1.Body.String())

	// Without name parameter (default)
	req2 := httptest.NewRequest("GET", "/greet", nil)
	w2 := httptest.NewRecorder()
	greetHandler(w2, req2)
	fmt.Printf("  Without name: %s\n", w2.Body.String())
}

// ──── 6. Test with Real Server ──────────────────────────────
func testWithServer() {
	// httptest.NewServer starts a real HTTP server on a random port.
	// Useful for integration tests where you need actual HTTP communication.

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/greet", greetHandler)

	srv := httptest.NewServer(mux)
	defer srv.Close()

	fmt.Printf("  Test server URL: %s\n", srv.URL)

	// Make real HTTP requests
	resp, err := http.Get(srv.URL + "/greet?name=TestUser")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("  Real HTTP response: %s\n", string(body))
}

// ──── 7. Test Headers ───────────────────────────────────────
func testHeaders() {
	req := httptest.NewRequest("GET", "/health", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("X-Request-ID", "req-123")

	w := httptest.NewRecorder()
	healthHandler(w, req)

	resp := w.Result()
	fmt.Printf("  Response Content-Type: %s\n", resp.Header.Get("Content-Type"))
	fmt.Printf("  Status: %d\n", resp.StatusCode)
	fmt.Println("  (In real tests, you'd verify the handler uses the headers correctly)")
}
