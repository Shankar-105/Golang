//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
)

// ============================================
// Building a complete JSON REST API
//
// This is what you'd build in FastAPI/Flask but in Go.
// Full CRUD with proper error handling, JSON input/output,
// and thread-safe storage using sync.RWMutex.
//
// Run: go run example_json_api.go
// ============================================

// ---- Models ----

type Book struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Year   int    `json:"year,omitempty"` // omitempty = skip if zero
}

type CreateBookRequest struct {
	Title  string `json:"title"`
	Author string `json:"author"`
	Year   int    `json:"year,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// ---- In-memory store (like a simple database) ----

type BookStore struct {
	mu     sync.RWMutex
	books  map[int]Book
	nextID int
}

func NewBookStore() *BookStore {
	return &BookStore{
		books:  make(map[int]Book),
		nextID: 1,
	}
}

func (s *BookStore) GetAll() []Book {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Book, 0, len(s.books))
	for _, b := range s.books {
		result = append(result, b)
	}
	return result
}

func (s *BookStore) GetByID(id int) (Book, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	b, ok := s.books[id]
	return b, ok
}

func (s *BookStore) Create(title, author string, year int) Book {
	s.mu.Lock()
	defer s.mu.Unlock()

	b := Book{ID: s.nextID, Title: title, Author: author, Year: year}
	s.books[s.nextID] = b
	s.nextID++
	return b
}

func (s *BookStore) Update(id int, title, author string, year int) (Book, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, ok := s.books[id]
	if !ok {
		return Book{}, false
	}

	if title != "" {
		b.Title = title
	}
	if author != "" {
		b.Author = author
	}
	if year != 0 {
		b.Year = year
	}
	s.books[id] = b
	return b, true
}

func (s *BookStore) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.books[id]
	if ok {
		delete(s.books, id)
	}
	return ok
}

// ---- HTTP Handlers ----

var store = NewBookStore()

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/books", handleListBooks)
	mux.HandleFunc("POST /api/books", handleCreateBook)
	mux.HandleFunc("GET /api/books/{id}", handleGetBook)
	mux.HandleFunc("PUT /api/books/{id}", handleUpdateBook)
	mux.HandleFunc("DELETE /api/books/{id}", handleDeleteBook)

	// Seed data
	store.Create("The Go Programming Language", "Donovan & Kernighan", 2015)
	store.Create("Concurrency in Go", "Katherine Cox-Buday", 2017)
	store.Create("Learning Go", "Jon Bodner", 2021)

	fmt.Println("=== Bookstore JSON API ===")
	fmt.Println("Listening on http://localhost:8080")
	fmt.Println("")
	fmt.Println("Try these:")
	fmt.Println("  curl http://localhost:8080/api/books")
	fmt.Println("  curl http://localhost:8080/api/books/1")
	fmt.Println("  curl -X POST http://localhost:8080/api/books \\")
	fmt.Println("       -H 'Content-Type: application/json' \\")
	fmt.Println("       -d '{\"title\":\"Black Hat Go\",\"author\":\"Tom Steele\",\"year\":2020}'")
	fmt.Println("  curl -X PUT http://localhost:8080/api/books/1 \\")
	fmt.Println("       -H 'Content-Type: application/json' \\")
	fmt.Println("       -d '{\"title\":\"The Go Programming Language (2nd ed)\"}'")
	fmt.Println("  curl -X DELETE http://localhost:8080/api/books/2")

	log.Fatal(http.ListenAndServe(":8080", mux))
}

func handleListBooks(w http.ResponseWriter, r *http.Request) {
	books := store.GetAll()
	respondJSON(w, http.StatusOK, books)
}

func handleGetBook(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid book ID"})
		return
	}

	book, ok := store.GetByID(id)
	if !ok {
		respondJSON(w, http.StatusNotFound, ErrorResponse{
			Error: fmt.Sprintf("book %d not found", id),
		})
		return
	}

	respondJSON(w, http.StatusOK, book)
}

func handleCreateBook(w http.ResponseWriter, r *http.Request) {
	var req CreateBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid request body",
			Details: err.Error(),
		})
		return
	}

	// Validate
	if req.Title == "" {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "title is required"})
		return
	}
	if req.Author == "" {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "author is required"})
		return
	}

	book := store.Create(req.Title, req.Author, req.Year)
	respondJSON(w, http.StatusCreated, book)
}

func handleUpdateBook(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid book ID"})
		return
	}

	var req CreateBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid request body",
			Details: err.Error(),
		})
		return
	}

	book, ok := store.Update(id, req.Title, req.Author, req.Year)
	if !ok {
		respondJSON(w, http.StatusNotFound, ErrorResponse{
			Error: fmt.Sprintf("book %d not found", id),
		})
		return
	}

	respondJSON(w, http.StatusOK, book)
}

func handleDeleteBook(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid book ID"})
		return
	}

	if !store.Delete(id) {
		respondJSON(w, http.StatusNotFound, ErrorResponse{
			Error: fmt.Sprintf("book %d not found", id),
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// respondJSON encodes any value as JSON and writes it
func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
}
