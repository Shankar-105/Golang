//go:build ignore

package main

import (
	"fmt"
	"math"
)

// ============================================
// Interfaces: implicit contracts
//
// Python: ABC with @abstractmethod (explicit)
//         Or duck typing (implicit, runtime)
// Go:    Interfaces (implicit, COMPILE-TIME)
//
// If a type has the right methods → it implements the interface.
// No "implements" keyword needed.
// ============================================

// Shape is an interface — any type with Area() and Perimeter() satisfies it
type Shape interface {
	Area() float64
	Perimeter() float64
}

// Stringer — like Python's __str__
// (This is actually defined in fmt package, but we show it here)
type Describer interface {
	Describe() string
}

// ---- Concrete types that satisfy Shape ----

type Circle struct {
	Radius float64
}

func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

func (c Circle) Perimeter() float64 {
	return 2 * math.Pi * c.Radius
}

func (c Circle) Describe() string {
	return fmt.Sprintf("Circle(r=%.1f)", c.Radius)
}

type Rectangle struct {
	Width, Height float64
}

func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

func (r Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

func (r Rectangle) Describe() string {
	return fmt.Sprintf("Rectangle(%.1f x %.1f)", r.Width, r.Height)
}

type Triangle struct {
	A, B, C float64 // side lengths
}

func (t Triangle) Perimeter() float64 {
	return t.A + t.B + t.C
}

func (t Triangle) Area() float64 {
	// Heron's formula
	s := t.Perimeter() / 2
	return math.Sqrt(s * (s - t.A) * (s - t.B) * (s - t.C))
}

func (t Triangle) Describe() string {
	return fmt.Sprintf("Triangle(%.1f, %.1f, %.1f)", t.A, t.B, t.C)
}

func main() {
	// ============================================
	// Example 1: Using interfaces
	// ============================================
	fmt.Println("=== Example 1: Interface usage ===")

	shapes := []Shape{
		Circle{Radius: 5},
		Rectangle{Width: 3, Height: 4},
		Triangle{A: 3, B: 4, C: 5},
	}

	for _, s := range shapes {
		printShapeInfo(s)
	}

	// ============================================
	// Example 2: Interface as parameter type
	// ============================================
	fmt.Println("\n=== Example 2: Largest shape ===")

	largest := findLargest(shapes)
	fmt.Printf("  Largest shape has area: %.2f\n", largest.Area())

	// ============================================
	// Example 3: Multiple interfaces
	// ============================================
	fmt.Println("\n=== Example 3: Multiple interfaces ===")

	// Circle satisfies both Shape AND Describer
	c := Circle{Radius: 10}
	describeAndMeasure(c)

	// ============================================
	// Example 4: Empty interface (any)
	// ============================================
	fmt.Println("\n=== Example 4: Empty interface (any) ===")

	printAnything(42)
	printAnything("hello")
	printAnything(true)
	printAnything(Circle{Radius: 3})

	// ============================================
	// Example 5: Interface satisfaction check
	// ============================================
	fmt.Println("\n=== Example 5: Compile-time check ===")

	// This is a common Go pattern to verify interface satisfaction
	// at compile time without creating an instance:
	var _ Shape = Circle{}    // compile error if Circle doesn't satisfy Shape
	var _ Shape = Rectangle{} // compile error if Rectangle doesn't satisfy Shape
	var _ Shape = Triangle{}  // compile error if Triangle doesn't satisfy Shape

	fmt.Println("  All types satisfy Shape interface!")
}

// printShapeInfo accepts ANY Shape — Circle, Rectangle, or Triangle
func printShapeInfo(s Shape) {
	fmt.Printf("  Area: %7.2f | Perimeter: %7.2f\n", s.Area(), s.Perimeter())
}

// findLargest returns the shape with the largest area
func findLargest(shapes []Shape) Shape {
	largest := shapes[0]
	for _, s := range shapes[1:] {
		if s.Area() > largest.Area() {
			largest = s
		}
	}
	return largest
}

// ShapeDescriber combines two interfaces
type ShapeDescriber interface {
	Shape
	Describer
}

// describeAndMeasure requires BOTH Shape and Describer methods
func describeAndMeasure(sd ShapeDescriber) {
	fmt.Printf("  %s → Area: %.2f\n", sd.Describe(), sd.Area())
}

// printAnything takes the empty interface — accepts any type
func printAnything(v any) {
	fmt.Printf("  [%T] %v\n", v, v)
}
