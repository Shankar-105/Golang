package stringutil

import (
	"testing"
)

// ──────────────────────────────────────────────────────────────
// Benchmarks for stringutil functions.
//
// Run with:
//   go test ./17_testing/stringutil/ -bench=. -benchmem
//   go test ./17_testing/stringutil/ -bench=BenchmarkReverse -benchmem
//   go test ./17_testing/stringutil/ -bench=. -benchtime=5s -benchmem
// ──────────────────────────────────────────────────────────────

// ════════════════════════════════════════════════════════════════
// 1. Basic Benchmark
// ════════════════════════════════════════════════════════════════
//
// Python equivalent:
//   import timeit
//   timeit.timeit(lambda: reverse("hello world"), number=1000000)
//
// Go benchmarks automatically determine the right number of iterations.

func BenchmarkReverse(b *testing.B) {
	// b.N is set by the testing framework to get stable results
	for i := 0; i < b.N; i++ {
		Reverse("Hello, World! This is a benchmark test string.")
	}
}

// ════════════════════════════════════════════════════════════════
// 2. Benchmark with Different Input Sizes
// ════════════════════════════════════════════════════════════════
//
// Use sub-benchmarks to compare performance across input sizes.
// Like pytest-benchmark with parametrize.

func BenchmarkReverseByLength(b *testing.B) {
	inputs := map[string]string{
		"short":  "hello",
		"medium": "The quick brown fox jumps over the lazy dog",
		"long":   "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam.",
	}

	for name, input := range inputs {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				Reverse(input)
			}
		})
	}
}

// ════════════════════════════════════════════════════════════════
// 3. Benchmark IsPalindrome
// ════════════════════════════════════════════════════════════════

func BenchmarkIsPalindrome(b *testing.B) {
	inputs := map[string]string{
		"short_true":  "racecar",
		"short_false": "hello",
		"long_true":   "A man, a plan, a canal: Panama",
		"long_false":  "This is definitely not a palindrome at all",
	}

	for name, input := range inputs {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				IsPalindrome(input)
			}
		})
	}
}

// ════════════════════════════════════════════════════════════════
// 4. Benchmark WordCount
// ════════════════════════════════════════════════════════════════

func BenchmarkWordCount(b *testing.B) {
	input := "The quick brown fox jumps over the lazy dog and then it jumps again"
	for i := 0; i < b.N; i++ {
		WordCount(input)
	}
}

// ════════════════════════════════════════════════════════════════
// 5. Benchmark CountVowels
// ════════════════════════════════════════════════════════════════

func BenchmarkCountVowels(b *testing.B) {
	input := "The quick brown fox jumps over the lazy dog"
	for i := 0; i < b.N; i++ {
		CountVowels(input)
	}
}
