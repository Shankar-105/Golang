//go:build ignore

package main

import "fmt"

// ============================================
// Composition over inheritance
//
// Python: class Dog(Animal) — single/multiple inheritance
// Go:    type Dog struct { Animal } — embedding (composition)
//
// Embedding promotes fields and methods. It LOOKS like inheritance
// but it's actually delegation.
// ============================================

// ---- Base types ----

type Animal struct {
	Name  string
	Age   int
	Sound string
}

func (a Animal) Speak() string {
	return fmt.Sprintf("%s says %s", a.Name, a.Sound)
}

func (a Animal) Info() string {
	return fmt.Sprintf("%s (age %d)", a.Name, a.Age)
}

// ---- Composed types ----

// Dog embeds Animal — gets all its fields and methods
type Dog struct {
	Animal // embedded (promoted)
	Breed  string
	Tricks []string
}

// Dog can "override" methods by defining its own
func (d Dog) Speak() string {
	return fmt.Sprintf("%s barks: Woof! 🐕", d.Name)
}

func (d Dog) ShowTricks() {
	if len(d.Tricks) == 0 {
		fmt.Printf("  %s doesn't know any tricks yet\n", d.Name)
		return
	}
	fmt.Printf("  %s's tricks: ", d.Name)
	for i, t := range d.Tricks {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(t)
	}
	fmt.Println()
}

// Cat embeds Animal
type Cat struct {
	Animal
	Indoor bool
}

func (c Cat) Speak() string {
	return fmt.Sprintf("%s purrs: Meow~ 🐱", c.Name)
}

// ---- Multiple embedding ----

type Logger struct {
	Prefix string
}

func (l Logger) Log(msg string) {
	fmt.Printf("  [%s] %s\n", l.Prefix, msg)
}

// Server embeds multiple types — like multiple "inheritance"
type Server struct {
	Logger
	Host string
	Port int
}

func (s Server) Start() {
	s.Log(fmt.Sprintf("Starting server on %s:%d", s.Host, s.Port))
}

func main() {
	// ============================================
	// Example 1: Basic embedding
	// ============================================
	fmt.Println("=== Example 1: Basic embedding ===")

	dog := Dog{
		Animal: Animal{Name: "Rex", Age: 3, Sound: "Woof"},
		Breed:  "Labrador",
		Tricks: []string{"sit", "shake", "roll over"},
	}

	// Promoted fields from Animal
	fmt.Println("  Name:", dog.Name)   // promoted
	fmt.Println("  Age:", dog.Age)     // promoted
	fmt.Println("  Breed:", dog.Breed) // Dog's own field

	// Can still access Animal explicitly
	fmt.Println("  Animal.Info():", dog.Animal.Info())

	// ============================================
	// Example 2: Method "overriding"
	// ============================================
	fmt.Println("\n=== Example 2: Method overriding ===")

	cat := Cat{
		Animal: Animal{Name: "Whiskers", Age: 5, Sound: "Meow"},
		Indoor: true,
	}

	// Dog and Cat each have their own Speak()
	fmt.Println("  Dog:", dog.Speak()) // Dog's Speak wins
	fmt.Println("  Cat:", cat.Speak()) // Cat's Speak wins

	// Access the "parent" method explicitly
	fmt.Println("  Dog (Animal.Speak):", dog.Animal.Speak())
	fmt.Println("  Cat (Animal.Speak):", cat.Animal.Speak())

	dog.ShowTricks()

	// ============================================
	// Example 3: Multiple embedding
	// ============================================
	fmt.Println("\n=== Example 3: Multiple embedding ===")

	srv := Server{
		Logger: Logger{Prefix: "HTTP"},
		Host:   "localhost",
		Port:   8080,
	}

	srv.Start()                    // Server's method
	srv.Log("Handling request...") // Promoted from Logger

	// ============================================
	// Example 4: Embedding vs named field
	// ============================================
	fmt.Println("\n=== Example 4: Embedding vs named field ===")

	type Config struct {
		Debug bool
	}

	// Embedded — fields promoted
	type App1 struct {
		Config
	}

	// Named field — not promoted
	type App2 struct {
		Cfg Config
	}

	a1 := App1{Config: Config{Debug: true}}
	a2 := App2{Cfg: Config{Debug: true}}

	fmt.Println("  Embedded: a1.Debug =", a1.Debug)      // promoted — direct access
	fmt.Println("  Named: a2.Cfg.Debug =", a2.Cfg.Debug) // must go through Cfg
}
