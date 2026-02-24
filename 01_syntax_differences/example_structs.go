//go:build ignore

package main

import "fmt"

// ============================================
// Structs + Methods: Go's replacement for Python classes
//
// Python:
//   class Dog:
//       def __init__(self, name, age):
//           self.name = name
//           self.age = age
//       def bark(self):
//           return f"{self.name} says woof!"
//
// Go has NO classes. Instead: struct (data) + methods (behavior).
// ============================================

// Dog is a struct — holds data only, like Python's @dataclass
type Dog struct {
	Name  string // Uppercase = exported (public)
	Age   int
	breed string // lowercase = unexported (private to this package)
}

// Bark is a method with a VALUE receiver.
// (d Dog) means: this method gets a COPY of the Dog.
// Like calling a function and passing a copy of the struct.
//
// Python equivalent: def bark(self) -> str:
// BUT unlike Python's self (which is a reference), this is a copy!
func (d Dog) Bark() string {
	return d.Name + " says woof!"
}

// Info returns a formatted string. Another value receiver method.
func (d Dog) Info() string {
	return fmt.Sprintf("%s is %d years old", d.Name, d.Age)
}

// SetAge is a method with a POINTER receiver.
// (d *Dog) means: this method gets a POINTER to the original Dog.
// It can MODIFY the original, just like Python's self.
func (d *Dog) SetAge(age int) {
	d.Age = age // modifies the ORIGINAL Dog, not a copy
}

// HaveBirthday demonstrates pointer receiver modifying the struct.
func (d *Dog) HaveBirthday() {
	d.Age++
	fmt.Printf("🎂 Happy birthday %s! Now %d years old.\n", d.Name, d.Age)
}

// ============================================
// Struct embedding: Go's version of inheritance (it's composition!)
//
// Python:
//   class ServiceDog(Dog):
//       def __init__(self, name, age, task):
//           super().__init__(name, age)
//           self.task = task
//
// Go: No inheritance. Embed the struct instead.
// ============================================

type ServiceDog struct {
	Dog  // embedded — ServiceDog "has a" Dog (not "is a" Dog)
	Task string
}

func main() {
	// ============================================
	// Creating structs (no __init__, no constructor)
	// ============================================
	fmt.Println("=== Creating structs ===")

	// Named fields (most readable — use this)
	rex := Dog{Name: "Rex", Age: 3}
	fmt.Println(rex.Info())
	fmt.Println(rex.Bark())

	// Positional (all fields required, in order — fragile, avoid unless simple)
	// buddy := Dog{"Buddy", 5, "labrador"}

	// Zero value struct (all fields get zero values)
	var ghost Dog
	fmt.Println("Zero dog:", ghost.Info()) // " is 0 years old"

	// ============================================
	// Value receiver vs Pointer receiver
	// ============================================
	fmt.Println("\n=== Value vs Pointer receiver ===")

	myDog := Dog{Name: "Luna", Age: 2}
	fmt.Println("Before:", myDog.Info())

	// Pointer receiver — modifies the original
	myDog.SetAge(5)
	fmt.Println("After SetAge(5):", myDog.Info())

	myDog.HaveBirthday()
	fmt.Println("After birthday:", myDog.Info())

	// NOTE: Go automatically takes the address when calling pointer methods.
	// You DON'T need to write (&myDog).SetAge(5) — Go does it for you.

	// ============================================
	// Why value vs pointer matters — the copy trap
	// ============================================
	fmt.Println("\n=== The copy trap ===")

	original := Dog{Name: "Max", Age: 4}
	copyDog := original // this is a COPY in Go (Python would make a reference!)

	copyDog.Name = "NotMax"
	fmt.Println("Original:", original.Name) // still "Max"!
	fmt.Println("Copy:", copyDog.Name)      // "NotMax"

	// In Python:
	//   original = Dog("Max", 4)
	//   copy_dog = original  # this is a REFERENCE, not a copy!
	//   copy_dog.name = "NotMax"
	//   print(original.name)  # "NotMax" — surprise!

	// To get Python-like behavior (shared reference), use pointers:
	originalPtr := &Dog{Name: "Max", Age: 4} // & takes the address
	aliasPtr := originalPtr                  // both point to the same Dog

	aliasPtr.Name = "NotMax"
	fmt.Println("\nPointer original:", originalPtr.Name) // "NotMax"
	fmt.Println("Pointer alias:", aliasPtr.Name)         // "NotMax"

	// ============================================
	// Struct embedding (composition)
	// ============================================
	fmt.Println("\n=== Struct embedding ===")

	sd := ServiceDog{
		Dog:  Dog{Name: "Buddy", Age: 5},
		Task: "Guide",
	}

	// Buddy's Dog methods are "promoted" — you can call them directly
	fmt.Println(sd.Bark())        // calls Dog's Bark()
	fmt.Println(sd.Info())        // calls Dog's Info()
	fmt.Println("Task:", sd.Task) // ServiceDog's own field

	// You can also access the embedded struct explicitly:
	fmt.Println("Embedded name:", sd.Dog.Name)
}
