//go:build ignore

package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// ============================================
// Strings in Go: immutable UTF-8 byte sequences
//
// Python: str is a sequence of Unicode characters
// Go:    string is a sequence of bytes (usually UTF-8)
//
// Key difference: len("café") is 5 in Go (bytes), 4 in Python (chars)
// ============================================

func main() {
	// ============================================
	// Example 1: Strings are bytes, not characters
	// ============================================
	fmt.Println("=== Example 1: Bytes vs characters ===")

	s := "Hello, 世界!"
	fmt.Println("  String:", s)
	fmt.Println("  len() (bytes):", len(s))                        // 13
	fmt.Println("  RuneCount (chars):", utf8.RuneCountInString(s)) // 10

	// Byte-level view
	fmt.Printf("  Bytes: ")
	for i := 0; i < len(s); i++ {
		fmt.Printf("%02x ", s[i])
	}
	fmt.Println()

	// Rune-level view (characters)
	fmt.Printf("  Runes: ")
	for _, r := range s {
		fmt.Printf("%c ", r)
	}
	fmt.Println()

	// ============================================
	// Example 2: range iterates runes, not bytes
	// ============================================
	fmt.Println("\n=== Example 2: Range over string ===")

	word := "café"
	fmt.Println("  String:", word, "| len:", len(word), "| runes:", utf8.RuneCountInString(word))

	// range gives (byte_index, rune)
	for i, r := range word {
		fmt.Printf("    byte[%d] = '%c' (U+%04X)\n", i, r, r)
	}
	// Notice: index jumps from 3 to 4 because 'é' is 2 bytes

	// ============================================
	// Example 3: Converting between bytes, runes, strings
	// ============================================
	fmt.Println("\n=== Example 3: Conversions ===")

	str := "Go 🚀"

	// String → byte slice
	bytes := []byte(str)
	fmt.Println("  []byte:", bytes)

	// String → rune slice (for character-level operations)
	runes := []rune(str)
	fmt.Println("  []rune:", runes)
	fmt.Println("  Rune count:", len(runes))

	// Rune slice → string
	reversed := reverseString(str)
	fmt.Println("  Reversed:", reversed)

	// ============================================
	// Example 4: strings package (most common operations)
	// ============================================
	fmt.Println("\n=== Example 4: strings package ===")

	s = "  Hello, Go World!  "

	fmt.Println("  Original:", fmt.Sprintf("%q", s))
	fmt.Println("  TrimSpace:", fmt.Sprintf("%q", strings.TrimSpace(s)))
	fmt.Println("  ToUpper:", strings.ToUpper("hello"))
	fmt.Println("  ToLower:", strings.ToLower("HELLO"))
	fmt.Println("  Contains:", strings.Contains("seafood", "foo"))
	fmt.Println("  HasPrefix:", strings.HasPrefix("Hello", "He"))
	fmt.Println("  HasSuffix:", strings.HasSuffix("Hello", "lo"))
	fmt.Println("  Count:", strings.Count("cheese", "e"))
	fmt.Println("  Index:", strings.Index("hello", "ll"))
	fmt.Println("  Replace:", strings.Replace("oink oink", "oink", "moo", -1))

	// Split and Join
	csv := "alice,bob,charlie"
	parts := strings.Split(csv, ",")
	fmt.Println("  Split:", parts)

	joined := strings.Join(parts, " | ")
	fmt.Println("  Join:", joined)

	// Fields — split on whitespace (like Python's split())
	sentence := "  hello   world   go  "
	words := strings.Fields(sentence)
	fmt.Println("  Fields:", words) // ["hello", "world", "go"]

	// ============================================
	// Example 5: Efficient string building
	// ============================================
	fmt.Println("\n=== Example 5: strings.Builder ===")

	// BAD: string concatenation in a loop (creates new string each time)
	// result := ""
	// for i := 0; i < 1000; i++ {
	//     result += "x"  // O(n²) — copies the whole string each time
	// }

	// GOOD: use strings.Builder
	var b strings.Builder
	for i := 0; i < 10; i++ {
		fmt.Fprintf(&b, "item_%d ", i)
	}
	fmt.Println("  Builder:", b.String())

	// Python equivalent:
	// parts = [f"item_{i}" for i in range(10)]
	// result = " ".join(parts)

	// ============================================
	// Example 6: Multi-line strings (raw strings)
	// ============================================
	fmt.Println("\n=== Example 6: Raw strings ===")

	// Backtick strings — raw, no escape processing
	raw := `This is a
multi-line string.
No need to escape \n or "quotes".`
	fmt.Println(raw)

	// Regular strings use double quotes and process escapes
	escaped := "Line 1\nLine 2\tTabbed"
	fmt.Println("\n  Escaped:", escaped)
}

// reverseString correctly reverses a UTF-8 string
func reverseString(s string) string {
	runes := []rune(s) // convert to runes for character-level ops
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
