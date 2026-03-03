# Lesson 15: `net/http` — Building HTTP Servers & Clients

> **Goal:** Learn Go's powerful standard library HTTP package. No frameworks needed — Go's `net/http` is production-grade out of the box. Compare to Python's Flask/FastAPI.

---

## 15.1 The Big Picture

| Python | Go |
|--------|-----|
| `flask.Flask()` / `fastapi.FastAPI()` | `http.NewServeMux()` (or `http.DefaultServeMux`) |
| `@app.get("/path")` decorator | `mux.HandleFunc("GET /path", handler)` |
| `uvicorn.run(app, port=8000)` | `http.ListenAndServe(":8080", mux)` |
| `requests.get(url)` | `http.Get(url)` |
| `request.json()` | `json.NewDecoder(r.Body).Decode(&v)` |
| `return JSONResponse(data)` | `json.NewEncoder(w).Encode(data)` |
| Middleware via decorators | Middleware via handler wrapping |

**Key insight:** In Python you need a framework (Flask, FastAPI, Django). In Go, the standard library `net/http` is all most projects need. It's fast, concurrent by default (goroutine-per-request), and battle-tested at Google scale.

---

## 15.2 The Simplest HTTP Server

```go
package main

import (
    "fmt"
    "net/http"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "Hello, World!")
    })

    fmt.Println("Server running on :8080")
    http.ListenAndServe(":8080", nil) // nil = use DefaultServeMux
}
```

That's it. 10 lines. Run it, visit `http://localhost:8080`, done.

Python equivalent:
```python
from fastapi import FastAPI
import uvicorn

app = FastAPI()

@app.get("/")
async def root():
    return {"message": "Hello, World!"}

uvicorn.run(app, port=8080)
```

**Go's advantage:** Every incoming request is automatically handled in its own **goroutine**. No async/await needed. No event loop. True parallelism on multi-core CPUs.

> See: [example_hello_server.go](example_hello_server.go)

---

## 15.3 Understanding Handlers

A **handler** is anything that implements the `http.Handler` interface:

```go
type Handler interface {
    ServeHTTP(w http.ResponseWriter, r *http.Request)
}
```

Two ways to create handlers:

### Way 1: `http.HandlerFunc` (most common)

```go
func myHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "Hello!")
}

mux.HandleFunc("/hello", myHandler)
```

### Way 2: Struct implementing `http.Handler` (for stateful handlers)

```go
type CounterHandler struct {
    mu    sync.Mutex
    count int
}

func (h *CounterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    h.mu.Lock()
    h.count++
    current := h.count
    h.mu.Unlock()
    fmt.Fprintf(w, "Visit #%d\n", current)
}

mux.Handle("/count", &CounterHandler{})
```

Python equivalent:
```python
# FastAPI — function-based (like HandlerFunc)
@app.get("/hello")
async def hello():
    return "Hello!"

# Flask — class-based (like Handler interface)
class CounterView(MethodView):
    count = 0
    def get(self):
        CounterView.count += 1
        return f"Visit #{CounterView.count}"
```

---

## 15.4 The Request Object: `*http.Request`

The request `r` contains everything about the incoming HTTP request:

```go
func handler(w http.ResponseWriter, r *http.Request) {
    // Method: GET, POST, PUT, DELETE, etc.
    fmt.Println("Method:", r.Method)

    // URL path
    fmt.Println("Path:", r.URL.Path)

    // Query parameters: /search?q=golang&page=2
    q := r.URL.Query().Get("q")        // "golang"
    page := r.URL.Query().Get("page")  // "2" (string!)

    // Headers
    contentType := r.Header.Get("Content-Type")
    userAgent := r.Header.Get("User-Agent")

    // Body (for POST/PUT)
    body, _ := io.ReadAll(r.Body)
    defer r.Body.Close()

    // Form data (POST forms)
    r.ParseForm()
    username := r.FormValue("username")

    // Remote address
    fmt.Println("Client IP:", r.RemoteAddr)

    // Context (for cancellation/timeouts — from Lesson 13!)
    ctx := r.Context()
}
```

Python equivalent:
```python
@app.post("/submit")
async def submit(request: Request):
    method = request.method            # r.Method
    path = request.url.path            # r.URL.Path
    q = request.query_params.get("q")  # r.URL.Query().Get("q")
    ct = request.headers["content-type"] # r.Header.Get(...)
    body = await request.body()        # io.ReadAll(r.Body)
    form = await request.form()        # r.ParseForm()
```

---

## 15.5 The Response Writer: `http.ResponseWriter`

The writer `w` is how you send the response:

```go
func handler(w http.ResponseWriter, r *http.Request) {
    // Set headers BEFORE writing body
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("X-Custom-Header", "my-value")

    // Set status code (default is 200 if not called)
    w.WriteHeader(http.StatusCreated) // 201

    // Write body
    w.Write([]byte(`{"status":"created"}`))
    // OR use fmt:
    // fmt.Fprintf(w, `{"status":"created"}`)
}
```

### Common Status Codes

```go
http.StatusOK                  // 200
http.StatusCreated             // 201
http.StatusBadRequest          // 400
http.StatusUnauthorized        // 401
http.StatusForbidden           // 403
http.StatusNotFound            // 404
http.StatusMethodNotAllowed    // 405
http.StatusInternalServerError // 500
```

### Quick Error Responses

```go
// http.Error sets status code + writes message + sets Content-Type to text/plain
http.Error(w, "not found", http.StatusNotFound)
http.Error(w, "bad request: missing name", http.StatusBadRequest)
```

> See: [example_request_response.go](example_request_response.go)

---

## 15.6 Routing with `http.ServeMux` (Go 1.22+)

Go 1.22 added **method-based routing and path parameters** to the standard mux:

```go
mux := http.NewServeMux()

// Static routes
mux.HandleFunc("GET /", homeHandler)
mux.HandleFunc("GET /about", aboutHandler)

// Method-specific routes
mux.HandleFunc("GET /api/users", listUsers)
mux.HandleFunc("POST /api/users", createUser)

// Path parameters (Go 1.22+)
mux.HandleFunc("GET /api/users/{id}", getUser)
mux.HandleFunc("DELETE /api/users/{id}", deleteUser)

// Wildcard (catch all under /files/)
mux.HandleFunc("GET /files/{path...}", serveFile)

http.ListenAndServe(":8080", mux)
```

### Extracting Path Parameters

```go
func getUser(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")  // Go 1.22+
    fmt.Fprintf(w, "Getting user: %s\n", id)
}
```

Python equivalent:
```python
# FastAPI
@app.get("/api/users/{id}")
async def get_user(id: int):
    return {"user_id": id}

# Flask
@app.route("/api/users/<int:id>")
def get_user(id):
    return {"user_id": id}
```

> See: [example_routing.go](example_routing.go)

---

## 15.7 Middleware — Wrapping Handlers

Middleware in Go is a function that takes a handler and returns a new handler:

```go
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        log.Printf("→ %s %s", r.Method, r.URL.Path)

        next.ServeHTTP(w, r) // call the actual handler

        log.Printf("← %s %s (%s)", r.Method, r.URL.Path, time.Since(start))
    })
}
```

### Chaining Middleware

```go
// Apply middleware to the entire mux:
handler := loggingMiddleware(authMiddleware(mux))
http.ListenAndServe(":8080", handler)
```

Python equivalent:
```python
# FastAPI middleware
@app.middleware("http")
async def log_requests(request: Request, call_next):
    start = time.time()
    print(f"→ {request.method} {request.url.path}")
    response = await call_next(request)
    print(f"← {request.method} {request.url.path} ({time.time()-start:.3f}s)")
    return response
```

### Common Middleware Patterns

```go
// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}

// Auth middleware
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        // validate token...
        next.ServeHTTP(w, r)
    })
}

// Recovery middleware (catch panics)
func recoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("PANIC: %v", err)
                http.Error(w, "internal server error", http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

> See: [example_middleware.go](example_middleware.go)

---

## 15.8 Building a JSON API — Putting It All Together

```go
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func listUsers(w http.ResponseWriter, r *http.Request) {
    users := []User{
        {ID: 1, Name: "Alice", Email: "alice@go.dev"},
        {ID: 2, Name: "Bob", Email: "bob@go.dev"},
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(users)
}

func createUser(w http.ResponseWriter, r *http.Request) {
    var user User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, "invalid JSON", http.StatusBadRequest)
        return
    }
    // save user...

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}
```

> See: [example_json_api.go](example_json_api.go)

---

## 15.9 HTTP Client — Making Requests

```go
// Simple GET
resp, err := http.Get("https://api.example.com/users")
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()

body, _ := io.ReadAll(resp.Body)
fmt.Println(string(body))

// Custom request with headers
client := &http.Client{Timeout: 10 * time.Second}
req, _ := http.NewRequest("GET", "https://api.example.com/users", nil)
req.Header.Set("Authorization", "Bearer my-token")

resp, err = client.Do(req)
```

Python equivalent:
```python
import httpx

# Simple GET
resp = httpx.get("https://api.example.com/users")
print(resp.json())

# With headers
resp = httpx.get(
    "https://api.example.com/users",
    headers={"Authorization": "Bearer my-token"},
    timeout=10,
)
```

> See: [example_http_client.go](example_http_client.go)

---

## 15.10 Graceful Shutdown

Production servers need to finish in-flight requests before stopping:

```go
srv := &http.Server{Addr: ":8080", Handler: mux}

// Start server in a goroutine
go func() {
    if err := srv.ListenAndServe(); err != http.ErrServerClosed {
        log.Fatal(err)
    }
}()

// Wait for interrupt signal
quit := make(chan os.Signal, 1)
signal.Notify(quit, os.Interrupt)
<-quit

// Give in-flight requests 30 seconds to complete
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
srv.Shutdown(ctx)
fmt.Println("Server shut down gracefully")
```

Python equivalent:
```python
# Uvicorn handles SIGINT/SIGTERM gracefully by default
# Or with aiohttp:
runner = web.AppRunner(app)
await runner.setup()
site = web.TCPSite(runner, 'localhost', 8080)
await site.start()
# ... wait for signal ...
await runner.cleanup()
```

> See: [example_graceful_shutdown.go](example_graceful_shutdown.go)

---

## 15.11 Quick Reference

```
Server:
  http.ListenAndServe(":8080", handler)
  http.NewServeMux()

Routing (Go 1.22+):
  mux.HandleFunc("GET /path", fn)
  mux.HandleFunc("GET /users/{id}", fn)
  r.PathValue("id")

Handler signature:
  func(w http.ResponseWriter, r *http.Request)

Response:
  w.Header().Set("Key", "Value")
  w.WriteHeader(statusCode)
  w.Write([]byte(body))
  fmt.Fprintf(w, "template %s", value)
  http.Error(w, message, code)

Request:
  r.Method, r.URL.Path, r.URL.Query()
  r.Header.Get("Key")
  r.Body (io.ReadCloser)
  r.PathValue("name")  // Go 1.22+
  r.Context()

Client:
  http.Get(url)
  http.Post(url, contentType, body)
  client := &http.Client{Timeout: 10*time.Second}
  client.Do(req)

Middleware:
  func mw(next http.Handler) http.Handler { ... }
```

---

## Exercises

### Exercise 1: Health Check API
Build a server with `GET /health` that returns `{"status":"ok","uptime":"5m30s"}`.

### Exercise 2: CRUD API
Build a full CRUD API for a "bookstore": `GET /books`, `POST /books`, `GET /books/{id}`, `PUT /books/{id}`, `DELETE /books/{id}`. Store data in a `map` protected by a `sync.RWMutex`.

### Exercise 3: Middleware Stack
Create logging + auth + recovery middleware. Stack them and test with invalid auth tokens, panicking handlers, etc.

---

> **Next → [Lesson 16: JSON, I/O & Encoding](../16_json_io/lesson.md)** — Marshaling data and streaming I/O.
