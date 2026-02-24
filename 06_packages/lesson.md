# Lesson 6: Packages, Modules & Visibility

> **Goal:** Understand Go's code organization — how it differs from Python's `import`, `pip`, `venv`, and `__all__`.

---

## 6.1 Python vs Go: Mental Model

| Concept | Python | Go |
|---------|--------|-----|
| A single file | A module (`.py`) | Part of a package (`.go`) |
| A folder of code | A package (`__init__.py`) | A package (all `.go` files in one dir share the same `package` name) |
| Dependency manager | `pip` + `venv` + `requirements.txt` | `go mod` + `go.sum` (built-in, no virtualenv needed) |
| Public/private | Convention (`_prefix` = private, `__all__`) | **Enforced by compiler**: Uppercase = exported, lowercase = unexported |
| Central registry | PyPI (`pip install requests`) | `proxy.golang.org` (`go get github.com/...`) |
| Import path | `from mypackage import MyClass` | `import "github.com/user/repo/package"` |

**Key insight:** In Python, a file IS a module. In Go, a **directory** is a package. All `.go` files in the same directory must declare the same `package` name, and they can access each other's unexported symbols as if they were one file.

---

## 6.2 Package Declaration & Structure

Every `.go` file starts with `package <name>`:

```
myproject/
├── go.mod              ← module declaration (like setup.py + requirements.txt combined)
├── main.go             ← package main (entry point)
├── server/
│   ├── handler.go      ← package server
│   └── middleware.go    ← package server (same package!)
└── utils/
    └── helpers.go      ← package utils
```

**Rules:**
1. All files in `server/` declare `package server`. They can see each other's unexported (lowercase) names.
2. `main.go` imports `server` and `utils` but can only use their **Exported** (uppercase) names.
3. There is no `__init__.py` equivalent. The directory IS the package.

```go
// server/handler.go
package server

// HandleRequest is exported (uppercase H) — other packages can call it
func HandleRequest() string {
    return formatResponse("OK") // can call unexported function in same package
}

// formatResponse is unexported (lowercase f) — only server package can use it
func formatResponse(msg string) string {
    return "Response: " + msg
}
```

```go
// main.go
package main

import (
    "fmt"
    "myproject/server"  // import by module path + package directory
)

func main() {
    fmt.Println(server.HandleRequest())  // ✅ exported
    // fmt.Println(server.formatResponse("hi")) // ❌ COMPILE ERROR — unexported
}
```

> See: [example_package_structure/](example_package_structure/) for a runnable multi-package example.

---

## 6.3 `go.mod` — Your Project's Identity Card

Python:
```
# requirements.txt         # setup.py / pyproject.toml
requests==2.28.0           name="myproject"
flask>=2.0                 version="1.0"
```

Go: Everything is in `go.mod`:
```
module gopractice          // ← your module's import path (like name in setup.py)

go 1.25.0                  // ← minimum Go version

require (
    github.com/gorilla/mux v1.8.0   // ← like requirements.txt entries
)
```

**Key commands:**

| Command | What it does | Python equivalent |
|---------|-------------|-------------------|
| `go mod init <name>` | Create a new module | `pip init` / create `setup.py` |
| `go mod tidy` | Add missing deps, remove unused | `pip freeze > requirements.txt` (but smarter) |
| `go get github.com/pkg` | Add a dependency | `pip install pkg` |
| `go build` | Compile | No direct equivalent (Python is interpreted) |
| `go run main.go` | Compile + run | `python main.py` |

**No virtualenv needed.** Go modules are hermetic — dependencies are cached globally in `$GOPATH/pkg/mod` but each project has its own `go.mod` specifying exact versions. No "activate venv" dance.

---

## 6.4 Visibility: The Uppercase Rule (Enforced, Not Convention)

This is worth repeating because it's so different from Python:

```go
package server

type Server struct {    // Server is exported (other packages can use it)
    Host string         // Host is exported
    port int            // port is unexported — only this package can see it
}

func NewServer() *Server {  // exported "constructor" (Go convention)
    return &Server{
        Host: "localhost",
        port: 8080,           // we can set unexported fields inside the package
    }
}

func (s *Server) Start() {  // exported method
    listen(s.port)           // calling unexported function
}

func listen(port int) {     // unexported helper
    fmt.Printf("Listening on :%d\n", port)
}
```

**Convention — "constructor" functions:** Go has no `__init__`. The convention is a function named `NewXxx` that returns a `*Xxx`:
- `NewServer()` → `*Server`
- `NewRouter(config Config)` → `*Router`

This is like Python's `@classmethod` factory pattern, but it's the *standard* way in Go.

---

## 6.5 Import Paths & Aliases

```go
import (
    "fmt"                          // standard library
    "net/http"                     // standard library, nested
    "github.com/gorilla/mux"      // third-party
    "gopractice/server"            // your own package (module path + dir)
)
```

**Alias imports** (like Python's `import numpy as np`):
```go
import (
    myhttp "net/http"              // alias to avoid conflicts
    _ "github.com/lib/pq"         // blank import: runs init() only (like side-effect imports)
    . "fmt"                        // dot import: Println() instead of fmt.Println() (avoid this)
)
```

**The blank import `_`:** Sometimes a package needs to be imported just for its `init()` side effect (e.g., registering a database driver). This is like Python's `import antigravity` — you import it for what it does on load, not for its exports.

---

## 6.6 `init()` Functions — Package Initialization

```go
package database

import "fmt"

// init() runs automatically when the package is imported.
// Like Python's module-level code that runs on import.
// You can have multiple init() functions even in the same file.
func init() {
    fmt.Println("database package initialized")
    // register drivers, set defaults, validate config, etc.
}
```

**Execution order:**
1. All imported packages' `init()` run first (recursively, depth-first).
2. Package-level variables are initialized.
3. `init()` functions run in source file order.
4. Finally, `main()` executes.

Python equivalent: code at the top level of a `.py` file that runs on `import`.

---

## 6.7 Internal Packages — Go's Access Control

Go has a special convention: any package in a directory named `internal/` can only be imported by code in the parent tree:

```
myproject/
├── server/
│   ├── internal/
│   │   └── auth.go     ← only server/ and its children can import this
│   └── handler.go
└── main.go             ← CANNOT import server/internal/auth
```

Python has no equivalent — there's nothing stopping you from `from mypackage._internal import secret`. Go enforces this at the **compiler** level.

---

## Summary

| Python | Go |
|--------|-----|
| `import module` | `import "path/to/package"` |
| `from pkg import func` | Not possible — always `pkg.Func()` |
| `pip install` + `venv` | `go get` + `go mod tidy` (no venv needed) |
| `_private` convention | `lowercase` = unexported (compiler enforced) |
| `__all__` for public API | Uppercase first letter = exported |
| `__init__.py` | Not needed — directory = package |
| `setup.py` / `pyproject.toml` | `go.mod` |
| Module-level code | `init()` function |

---

## Exercises

### Exercise 1: Create a Multi-Package Project
Create a project with this structure and make it compile:
```
calculator/
├── go.mod
├── main.go          ← uses math and display packages
├── math/
│   └── operations.go  ← exports Add, Subtract; unexported multiply
└── display/
    └── printer.go     ← exports PrettyPrint
```

### Exercise 2: Visibility Bug Hunt
Why won't this compile?
```go
// file: models/user.go
package models

type user struct {
    name string
    Age  int
}

// file: main.go
package main

import "myapp/models"

func main() {
    u := models.user{name: "Alice", Age: 25}
    fmt.Println(u.name)
}
```
Fix it. (Hint: what's exported and what isn't?)

---

> **Phase 1 Complete!** You now have the language foundations. 
> **Next → Phase 2: Concurrency — The Main Event.** Starting with [Lesson 7: Why Goroutines Exist](../07_why_goroutines/lesson.md)
