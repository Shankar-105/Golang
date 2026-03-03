// Package stringutil provides string manipulation functions.
// This package exists to demonstrate Go testing patterns.
package stringutil

import (
	"errors"
	"strings"
	"unicode"
)

// Reverse returns the string reversed.
// "hello" → "olleh"
// Handles Unicode correctly (rune-by-rune, not byte-by-byte).
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// IsPalindrome checks if a string reads the same forwards and backwards.
// It ignores case and non-letter characters.
// "Racecar" → true, "A man, a plan, a canal: Panama" → true
func IsPalindrome(s string) bool {
	// Clean: lowercase, letters only
	var cleaned []rune
	for _, r := range s {
		if unicode.IsLetter(r) {
			cleaned = append(cleaned, unicode.ToLower(r))
		}
	}

	for i, j := 0, len(cleaned)-1; i < j; i, j = i+1, j-1 {
		if cleaned[i] != cleaned[j] {
			return false
		}
	}
	return true
}

// WordCount returns the number of words in a string.
// Words are separated by whitespace.
func WordCount(s string) int {
	return len(strings.Fields(s))
}

// Truncate shortens a string to maxLen characters, adding "..." if truncated.
// Returns error if maxLen < 3.
func Truncate(s string, maxLen int) (string, error) {
	if maxLen < 3 {
		return "", errors.New("maxLen must be at least 3")
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s, nil
	}
	return string(runes[:maxLen-3]) + "...", nil
}

// CamelToSnake converts "CamelCase" to "camel_case".
// "HelloWorld" → "hello_world"
// "HTTPServer" → "http_server"
func CamelToSnake(s string) string {
	var result []rune
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// CountVowels counts the number of vowels (a, e, i, o, u) in a string.
func CountVowels(s string) int {
	count := 0
	for _, r := range strings.ToLower(s) {
		switch r {
		case 'a', 'e', 'i', 'o', 'u':
			count++
		}
	}
	return count
}
