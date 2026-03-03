package stringutil

import (
	"testing"
)

// ──────────────────────────────────────────────────────────────
// Basic tests + table-driven tests for stringutil functions.
//
// Run with:
//   go test ./17_testing/stringutil/ -v
//   go test ./17_testing/stringutil/ -run TestReverse
//   go test ./17_testing/stringutil/ -cover
// ──────────────────────────────────────────────────────────────

// ════════════════════════════════════════════════════════════════
// 1. Basic Test — The Simplest Form
// ════════════════════════════════════════════════════════════════

// Python equivalent:
//   def test_reverse_simple():
//       assert reverse("hello") == "olleh"
func TestReverseSimple(t *testing.T) {
	got := Reverse("hello")
	want := "olleh"
	if got != want {
		t.Errorf("Reverse(%q) = %q, want %q", "hello", got, want)
	}
}

// ════════════════════════════════════════════════════════════════
// 2. Table-Driven Tests — THE Go Pattern
// ════════════════════════════════════════════════════════════════

// Python equivalent:
//   @pytest.mark.parametrize("input,expected", [
//       ("hello", "olleh"),
//       ("", ""),
//       ("a", "a"),
//       ("racecar", "racecar"),
//       ("Hello, 世界", "界世 ,olleH"),
//   ])
//   def test_reverse(input, expected):
//       assert reverse(input) == expected

func TestReverse(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "simple", input: "hello", want: "olleh"},
		{name: "empty string", input: "", want: ""},
		{name: "single char", input: "a", want: "a"},
		{name: "palindrome", input: "racecar", want: "racecar"},
		{name: "unicode", input: "Hello, 世界", want: "界世 ,olleH"},
		{name: "spaces", input: "go is great", want: "taerg si og"},
		{name: "numbers", input: "abc123", want: "321cba"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Reverse(tt.input)
			if got != tt.want {
				t.Errorf("Reverse(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ════════════════════════════════════════════════════════════════
// 3. IsPalindrome — Testing Booleans
// ════════════════════════════════════════════════════════════════

func TestIsPalindrome(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"simple palindrome", "racecar", true},
		{"mixed case", "Racecar", true},
		{"with spaces and punctuation", "A man, a plan, a canal: Panama", true},
		{"not palindrome", "hello", false},
		{"empty string", "", true},
		{"single char", "a", true},
		{"two same chars", "aa", true},
		{"two different chars", "ab", false},
		{"numbers ignored (letters only)", "abc", false},
		{"madam", "Madam", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsPalindrome(tt.input)
			if got != tt.want {
				t.Errorf("IsPalindrome(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// ════════════════════════════════════════════════════════════════
// 4. WordCount — Simple Integer Results
// ════════════════════════════════════════════════════════════════

func TestWordCount(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"simple", "hello world", 2},
		{"empty", "", 0},
		{"single word", "hello", 1},
		{"multiple spaces", "hello   world   foo", 3},
		{"leading/trailing spaces", "  hello world  ", 2},
		{"tabs and newlines", "hello\tworld\nfoo", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WordCount(tt.input)
			if got != tt.want {
				t.Errorf("WordCount(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

// ════════════════════════════════════════════════════════════════
// 5. Truncate — Testing Functions That Return Errors
// ════════════════════════════════════════════════════════════════

// Python equivalent:
//   def test_truncate_valid():
//       assert truncate("Hello, World!", 10) == "Hello, ..."
//
//   def test_truncate_invalid():
//       with pytest.raises(ValueError):
//           truncate("hello", 2)

func TestTruncate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		maxLen  int
		want    string
		wantErr bool
	}{
		{"truncate long string", "Hello, World!", 10, "Hello, ...", false},
		{"string fits", "Hello", 10, "Hello", false},
		{"exact length", "Hello", 5, "Hello", false},
		{"minimum maxLen", "Hello", 3, "...", false},
		{"maxLen too small", "Hello", 2, "", true},
		{"maxLen zero", "Hello", 0, "", true},
		{"empty string", "", 5, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Truncate(tt.input, tt.maxLen)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Fatalf("Truncate(%q, %d) error = %v, wantErr %v",
					tt.input, tt.maxLen, err, tt.wantErr)
			}

			// If we expected an error, don't check the result
			if tt.wantErr {
				return
			}

			// Check result
			if got != tt.want {
				t.Errorf("Truncate(%q, %d) = %q, want %q",
					tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

// ════════════════════════════════════════════════════════════════
// 6. CamelToSnake
// ════════════════════════════════════════════════════════════════

func TestCamelToSnake(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"HelloWorld", "hello_world"},
		{"CamelCase", "camel_case"},
		{"simple", "simple"},
		{"A", "a"},
		{"", ""},
		{"ABC", "a_b_c"},
		{"getHTTPResponse", "get_h_t_t_p_response"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := CamelToSnake(tt.input)
			if got != tt.want {
				t.Errorf("CamelToSnake(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ════════════════════════════════════════════════════════════════
// 7. CountVowels
// ════════════════════════════════════════════════════════════════

func TestCountVowels(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"hello", "hello", 2},
		{"empty", "", 0},
		{"all vowels", "aeiou", 5},
		{"no vowels", "bcdf", 0},
		{"mixed case", "HELLO", 2},
		{"sentence", "the quick brown fox", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CountVowels(tt.input)
			if got != tt.want {
				t.Errorf("CountVowels(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

// ════════════════════════════════════════════════════════════════
// 8. Test Helpers — Reusable Assertion Functions
// ════════════════════════════════════════════════════════════════

// assertEqual is a test helper. t.Helper() makes error messages
// point to the calling test, not this function.
func assertEqual(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestUsingHelper(t *testing.T) {
	assertEqual(t, Reverse("abc"), "cba")
	assertEqual(t, Reverse("xyz"), "zyx")
	assertEqual(t, Reverse(""), "")
}

// ════════════════════════════════════════════════════════════════
// 9. Property: Reverse(Reverse(s)) == s
// ════════════════════════════════════════════════════════════════

func TestReverseReverse(t *testing.T) {
	inputs := []string{"hello", "", "a", "Hello, 世界", "racecar", "12345"}
	for _, s := range inputs {
		t.Run(s, func(t *testing.T) {
			got := Reverse(Reverse(s))
			if got != s {
				t.Errorf("Reverse(Reverse(%q)) = %q, want %q", s, got, s)
			}
		})
	}
}
