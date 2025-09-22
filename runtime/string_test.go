package runtime

import (
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// TestStringFunctions tests all string functions using TDD approach
func TestStringFunctions(t *testing.T) {
	functions := GetStringFunctions()
	functionMap := make(map[string]*registry.Function)
	for _, fn := range functions {
		functionMap[fn.Name] = fn
	}

	t.Run("strrpos", func(t *testing.T) {
		fn := functionMap["strrpos"]
		if fn == nil {
			t.Fatal("strrpos function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected interface{}
		}{
			{
				name: "find last occurrence",
				args: []*values.Value{
					values.NewString("hello world hello"),
					values.NewString("hello"),
				},
				expected: int64(12),
			},
			{
				name: "not found",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("xyz"),
				},
				expected: false,
			},
			{
				name: "needle longer than haystack",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("hello world"),
				},
				expected: false,
			},
			{
				name: "empty haystack",
				args: []*values.Value{
					values.NewString(""),
					values.NewString("hello"),
				},
				expected: false,
			},
			{
				name: "empty needle",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString(""),
				},
				expected: int64(5),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				switch expected := tt.expected.(type) {
				case int64:
					if result.Type != values.TypeInt || result.Data.(int64) != expected {
						t.Errorf("Expected %d, got %v", expected, result)
					}
				case bool:
					if result.Type != values.TypeBool || result.Data.(bool) != expected {
						t.Errorf("Expected %v, got %v", expected, result)
					}
				}
			})
		}
	})

	t.Run("stripos", func(t *testing.T) {
		fn := functionMap["stripos"]
		if fn == nil {
			t.Fatal("stripos function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected interface{}
		}{
			{
				name: "case insensitive match",
				args: []*values.Value{
					values.NewString("Hello World"),
					values.NewString("WORLD"),
				},
				expected: int64(6),
			},
			{
				name: "not found",
				args: []*values.Value{
					values.NewString("Hello World"),
					values.NewString("xyz"),
				},
				expected: false,
			},
			{
				name: "case insensitive at start",
				args: []*values.Value{
					values.NewString("HELLO"),
					values.NewString("hello"),
				},
				expected: int64(0),
			},
			{
				name: "empty haystack",
				args: []*values.Value{
					values.NewString(""),
					values.NewString("hello"),
				},
				expected: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				switch expected := tt.expected.(type) {
				case int64:
					if result.Type != values.TypeInt || result.Data.(int64) != expected {
						t.Errorf("Expected %d, got %v", expected, result)
					}
				case bool:
					if result.Type != values.TypeBool || result.Data.(bool) != expected {
						t.Errorf("Expected %v, got %v", expected, result)
					}
				}
			})
		}
	})

	t.Run("substr_count", func(t *testing.T) {
		fn := functionMap["substr_count"]
		if fn == nil {
			t.Fatal("substr_count function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected int64
		}{
			{
				name: "multiple occurrences",
				args: []*values.Value{
					values.NewString("hello world hello"),
					values.NewString("hello"),
				},
				expected: 2,
			},
			{
				name: "overlapping matches",
				args: []*values.Value{
					values.NewString("aaaa"),
					values.NewString("aa"),
				},
				expected: 2,
			},
			{
				name: "not found",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("xyz"),
				},
				expected: 0,
			},
			{
				name: "empty haystack",
				args: []*values.Value{
					values.NewString(""),
					values.NewString("hello"),
				},
				expected: 0,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				if result.Type != values.TypeInt || result.Data.(int64) != tt.expected {
					t.Errorf("Expected %d, got %v", tt.expected, result)
				}
			})
		}
	})

	t.Run("ucfirst", func(t *testing.T) {
		fn := functionMap["ucfirst"]
		if fn == nil {
			t.Fatal("ucfirst function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected string
		}{
			{
				name:     "lowercase first",
				args:     []*values.Value{values.NewString("hello world")},
				expected: "Hello world",
			},
			{
				name:     "already uppercase",
				args:     []*values.Value{values.NewString("HELLO WORLD")},
				expected: "HELLO WORLD",
			},
			{
				name:     "empty string",
				args:     []*values.Value{values.NewString("")},
				expected: "",
			},
			{
				name:     "single character",
				args:     []*values.Value{values.NewString("h")},
				expected: "H",
			},
			{
				name:     "starts with number",
				args:     []*values.Value{values.NewString("123abc")},
				expected: "123abc",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				if result.Type != values.TypeString || result.Data.(string) != tt.expected {
					t.Errorf("Expected %q, got %v", tt.expected, result)
				}
			})
		}
	})

	t.Run("lcfirst", func(t *testing.T) {
		fn := functionMap["lcfirst"]
		if fn == nil {
			t.Fatal("lcfirst function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected string
		}{
			{
				name:     "uppercase first",
				args:     []*values.Value{values.NewString("Hello World")},
				expected: "hello World",
			},
			{
				name:     "all uppercase",
				args:     []*values.Value{values.NewString("HELLO WORLD")},
				expected: "hELLO WORLD",
			},
			{
				name:     "empty string",
				args:     []*values.Value{values.NewString("")},
				expected: "",
			},
			{
				name:     "single character",
				args:     []*values.Value{values.NewString("H")},
				expected: "h",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				if result.Type != values.TypeString || result.Data.(string) != tt.expected {
					t.Errorf("Expected %q, got %v", tt.expected, result)
				}
			})
		}
	})

	t.Run("ucwords", func(t *testing.T) {
		fn := functionMap["ucwords"]
		if fn == nil {
			t.Fatal("ucwords function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected string
		}{
			{
				name:     "multiple words",
				args:     []*values.Value{values.NewString("hello world test")},
				expected: "Hello World Test",
			},
			{
				name:     "already uppercase",
				args:     []*values.Value{values.NewString("HELLO WORLD")},
				expected: "HELLO WORLD",
			},
			{
				name:     "empty string",
				args:     []*values.Value{values.NewString("")},
				expected: "",
			},
			{
				name: "custom delimiters",
				args: []*values.Value{
					values.NewString("hello-world_test"),
					values.NewString("-_"),
				},
				expected: "Hello-World_Test",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				if result.Type != values.TypeString || result.Data.(string) != tt.expected {
					t.Errorf("Expected %q, got %v", tt.expected, result)
				}
			})
		}
	})

	t.Run("str_ireplace", func(t *testing.T) {
		fn := functionMap["str_ireplace"]
		if fn == nil {
			t.Fatal("str_ireplace function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected string
		}{
			{
				name: "case insensitive replace",
				args: []*values.Value{
					values.NewString("HELLO"),
					values.NewString("hi"),
					values.NewString("Hello World"),
				},
				expected: "hi World",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				if result.Type != values.TypeString || result.Data.(string) != tt.expected {
					t.Errorf("Expected %q, got %v", tt.expected, result)
				}
			})
		}
	})

	t.Run("strcmp", func(t *testing.T) {
		fn := functionMap["strcmp"]
		if fn == nil {
			t.Fatal("strcmp function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected int64
		}{
			{
				name: "equal strings",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("hello"),
				},
				expected: 0,
			},
			{
				name: "first less than second",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("world"),
				},
				expected: -1,
			},
			{
				name: "first greater than second",
				args: []*values.Value{
					values.NewString("world"),
					values.NewString("hello"),
				},
				expected: 1,
			},
			{
				name: "case sensitive",
				args: []*values.Value{
					values.NewString("Hello"),
					values.NewString("hello"),
				},
				expected: -1,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				resultInt := result.Data.(int64)
				// Normalize to -1, 0, 1 for comparison
				if resultInt < 0 {
					resultInt = -1
				} else if resultInt > 0 {
					resultInt = 1
				}

				if result.Type != values.TypeInt || resultInt != tt.expected {
					t.Errorf("Expected %d, got %v", tt.expected, result)
				}
			})
		}
	})

	t.Run("strcasecmp", func(t *testing.T) {
		fn := functionMap["strcasecmp"]
		if fn == nil {
			t.Fatal("strcasecmp function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected int64
		}{
			{
				name: "case insensitive equal",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("HELLO"),
				},
				expected: 0,
			},
			{
				name: "case insensitive less",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("world"),
				},
				expected: -1,
			},
			{
				name: "case insensitive greater",
				args: []*values.Value{
					values.NewString("world"),
					values.NewString("hello"),
				},
				expected: 1,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				resultInt := result.Data.(int64)
				// Normalize to -1, 0, 1 for comparison
				if resultInt < 0 {
					resultInt = -1
				} else if resultInt > 0 {
					resultInt = 1
				}

				if result.Type != values.TypeInt || resultInt != tt.expected {
					t.Errorf("Expected %d, got %v", tt.expected, result)
				}
			})
		}
	})

	t.Run("str_pad", func(t *testing.T) {
		fn := functionMap["str_pad"]
		if fn == nil {
			t.Fatal("str_pad function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected string
		}{
			{
				name: "pad right with spaces",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewInt(10),
				},
				expected: "hello     ",
			},
			{
				name: "pad right with custom char",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewInt(10),
					values.NewString("*"),
				},
				expected: "hello*****",
			},
			{
				name: "pad left",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewInt(10),
					values.NewString("*"),
					values.NewInt(0), // STR_PAD_LEFT
				},
				expected: "*****hello",
			},
			{
				name: "pad both sides",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewInt(10),
					values.NewString("*"),
					values.NewInt(2), // STR_PAD_BOTH
				},
				expected: "**hello***",
			},
			{
				name: "no padding needed",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewInt(3),
				},
				expected: "hello",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				if result.Type != values.TypeString || result.Data.(string) != tt.expected {
					t.Errorf("Expected %q, got %q", tt.expected, result.Data.(string))
				}
			})
		}
	})

	t.Run("strrev", func(t *testing.T) {
		fn := functionMap["strrev"]
		if fn == nil {
			t.Fatal("strrev function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected string
		}{
			{
				name:     "reverse normal string",
				args:     []*values.Value{values.NewString("hello")},
				expected: "olleh",
			},
			{
				name:     "empty string",
				args:     []*values.Value{values.NewString("")},
				expected: "",
			},
			{
				name:     "single character",
				args:     []*values.Value{values.NewString("a")},
				expected: "a",
			},
			{
				name:     "numbers",
				args:     []*values.Value{values.NewString("12345")},
				expected: "54321",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				if result.Type != values.TypeString || result.Data.(string) != tt.expected {
					t.Errorf("Expected %q, got %v", tt.expected, result)
				}
			})
		}
	})

	t.Run("strstr", func(t *testing.T) {
		fn := functionMap["strstr"]
		if fn == nil {
			t.Fatal("strstr function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected interface{}
		}{
			{
				name: "find first occurrence",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("world"),
				},
				expected: "world",
			},
			{
				name: "not found",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("xyz"),
				},
				expected: false,
			},
			{
				name: "find character occurrence",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("o"),
				},
				expected: "o world",
			},
			{
				name: "empty needle returns entire string",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString(""),
				},
				expected: "hello world",
			},
			{
				name: "empty haystack",
				args: []*values.Value{
					values.NewString(""),
					values.NewString("hello"),
				},
				expected: false,
			},
			{
				name: "before needle option - basic",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("world"),
					values.NewBool(true),
				},
				expected: "hello ",
			},
			{
				name: "before needle option - character",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("o"),
					values.NewBool(true),
				},
				expected: "hell",
			},
			{
				name: "before needle option - not found",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("xyz"),
					values.NewBool(true),
				},
				expected: false,
			},
			{
				name: "before needle option - at start",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("hello"),
					values.NewBool(true),
				},
				expected: "",
			},
			{
				name: "case sensitive - not found",
				args: []*values.Value{
					values.NewString("Hello World"),
					values.NewString("world"),
				},
				expected: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				switch expected := tt.expected.(type) {
				case string:
					if result.Type != values.TypeString || result.Data.(string) != expected {
						t.Errorf("Expected %q, got %v", expected, result)
					}
				case bool:
					if result.Type != values.TypeBool || result.Data.(bool) != expected {
						t.Errorf("Expected %v, got %v", expected, result)
					}
				}
			})
		}
	})

	t.Run("strrchr", func(t *testing.T) {
		fn := functionMap["strrchr"]
		if fn == nil {
			t.Fatal("strrchr function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected interface{}
		}{
			{
				name: "find last occurrence of character",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("o"),
				},
				expected: "orld",
			},
			{
				name: "find last occurrence of repeating character",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("l"),
				},
				expected: "ld",
			},
			{
				name: "character not found",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("x"),
				},
				expected: false,
			},
			{
				name: "empty needle",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString(""),
				},
				expected: false,
			},
			{
				name: "empty haystack",
				args: []*values.Value{
					values.NewString(""),
					values.NewString("o"),
				},
				expected: false,
			},
			{
				name: "multi-character needle uses first char",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("orld"),
				},
				expected: "orld",
			},
			{
				name: "multi-character needle not found",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("xyz"),
				},
				expected: false,
			},
			{
				name: "path parsing example",
				args: []*values.Value{
					values.NewString("/path/to/file.txt"),
					values.NewString("/"),
				},
				expected: "/file.txt",
			},
			{
				name: "file extension parsing",
				args: []*values.Value{
					values.NewString("/path/to/file.txt"),
					values.NewString("."),
				},
				expected: ".txt",
			},
			{
				name: "case sensitive - lowercase",
				args: []*values.Value{
					values.NewString("Hello World"),
					values.NewString("o"),
				},
				expected: "orld",
			},
			{
				name: "case sensitive - uppercase not found",
				args: []*values.Value{
					values.NewString("Hello World"),
					values.NewString("O"),
				},
				expected: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				switch expected := tt.expected.(type) {
				case string:
					if result.Type != values.TypeString || result.Data.(string) != expected {
						t.Errorf("Expected %q, got %v", expected, result)
					}
				case bool:
					if result.Type != values.TypeBool || result.Data.(bool) != expected {
						t.Errorf("Expected %v, got %v", expected, result)
					}
				}
			})
		}
	})

	t.Run("strtr", func(t *testing.T) {
		fn := functionMap["strtr"]
		if fn == nil {
			t.Fatal("strtr function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected string
		}{
			// Character translation mode (3 arguments)
			{
				name: "character translation basic",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("helo"),
					values.NewString("HELO"),
				},
				expected: "HELLO wOrLd",
			},
			{
				name: "character translation partial",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("hl"),
					values.NewString("HL"),
				},
				expected: "HeLLo worLd",
			},
			{
				name: "character translation no match",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("xyz"),
					values.NewString("ABC"),
				},
				expected: "hello world",
			},
			{
				name: "character translation empty from",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString(""),
					values.NewString("ABC"),
				},
				expected: "hello world",
			},
			{
				name: "character translation empty to",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("helo"),
					values.NewString(""),
				},
				expected: "hello world",
			},
			{
				name: "character translation from longer than to",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("abcdef"),
					values.NewString("xy"),
				},
				expected: "hello",
			},
			{
				name: "empty string character mode",
				args: []*values.Value{
					values.NewString(""),
					values.NewString("abc"),
					values.NewString("xyz"),
				},
				expected: "",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				if result.Type != values.TypeString || result.Data.(string) != tt.expected {
					t.Errorf("Expected %q, got %v", tt.expected, result)
				}
			})
		}
	})

	t.Run("str_split", func(t *testing.T) {
		fn := functionMap["str_split"]
		if fn == nil {
			t.Fatal("str_split function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected []string
		}{
			{
				name: "split into characters",
				args: []*values.Value{
					values.NewString("hello"),
				},
				expected: []string{"h", "e", "l", "l", "o"},
			},
			{
				name: "split empty string",
				args: []*values.Value{
					values.NewString(""),
				},
				expected: []string{},
			},
			{
				name: "split with chunk size 2",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewInt(2),
				},
				expected: []string{"he", "ll", "o"},
			},
			{
				name: "split with chunk size 3",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewInt(3),
				},
				expected: []string{"hel", "lo"},
			},
			{
				name: "split with chunk size larger than string",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewInt(10),
				},
				expected: []string{"hello"},
			},
			{
				name: "split with spaces",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewInt(5),
				},
				expected: []string{"hello", " worl", "d"},
			},
			{
				name: "split with chunk size 1",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewInt(1),
				},
				expected: []string{"h", "e", "l", "l", "o"},
			},
			{
				name: "split empty string with chunk size",
				args: []*values.Value{
					values.NewString(""),
					values.NewInt(2),
				},
				expected: []string{},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				if result.Type != values.TypeArray {
					t.Fatalf("Expected array, got %v", result.Type)
				}

				// Access array using the proper values API
				actualCount := result.ArrayCount()
				if actualCount != len(tt.expected) {
					t.Errorf("Expected array length %d, got %d", len(tt.expected), actualCount)
					return
				}

				for i, expected := range tt.expected {
					keyValue := values.NewInt(int64(i))
					elementValue := result.ArrayGet(keyValue)

					if elementValue.Type != values.TypeString || elementValue.Data.(string) != expected {
						t.Errorf("Array element %d: expected %q, got %v", i, expected, elementValue)
					}
				}
			})
		}
	})

	t.Run("chunk_split", func(t *testing.T) {
		fn := functionMap["chunk_split"]
		if fn == nil {
			t.Fatal("chunk_split function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected string
		}{
			{
				name: "basic chunk split with defaults",
				args: []*values.Value{
					values.NewString("hello"),
				},
				expected: "hello\r\n",
			},
			{
				name: "empty string",
				args: []*values.Value{
					values.NewString(""),
				},
				expected: "\r\n",
			},
			{
				name: "custom chunk length 2",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewInt(2),
				},
				expected: "he\r\nll\r\no\r\n",
			},
			{
				name: "custom chunk length 3",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewInt(3),
				},
				expected: "hel\r\nlo\r\n",
			},
			{
				name: "custom chunk length with spaces",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewInt(5),
				},
				expected: "hello\r\n worl\r\nd\r\n",
			},
			{
				name: "custom ending",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewInt(2),
					values.NewString("-"),
				},
				expected: "he-ll-o-",
			},
			{
				name: "custom ending pipe",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewInt(3),
					values.NewString("|"),
				},
				expected: "hel|lo|",
			},
			{
				name: "chunk size 1",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewInt(1),
				},
				expected: "h\r\ne\r\nl\r\nl\r\no\r\n",
			},
			{
				name: "chunk size larger than string",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewInt(10),
				},
				expected: "hello\r\n",
			},
			{
				name: "empty ending",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewInt(2),
					values.NewString(""),
				},
				expected: "hello",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				if result.Type != values.TypeString || result.Data.(string) != tt.expected {
					t.Errorf("Expected %q, got %v", tt.expected, result)
				}
			})
		}
	})

	t.Run("stristr", func(t *testing.T) {
		fn := functionMap["stristr"]
		if fn == nil {
			t.Fatal("stristr function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected interface{}
		}{
			{
				name: "case insensitive match uppercase needle",
				args: []*values.Value{
					values.NewString("Hello World"),
					values.NewString("WORLD"),
				},
				expected: "World",
			},
			{
				name: "case insensitive match lowercase needle",
				args: []*values.Value{
					values.NewString("Hello World"),
					values.NewString("world"),
				},
				expected: "World",
			},
			{
				name: "not found",
				args: []*values.Value{
					values.NewString("Hello World"),
					values.NewString("xyz"),
				},
				expected: false,
			},
			{
				name: "case insensitive character match uppercase",
				args: []*values.Value{
					values.NewString("Hello World"),
					values.NewString("O"),
				},
				expected: "o World",
			},
			{
				name: "case insensitive character match lowercase",
				args: []*values.Value{
					values.NewString("Hello World"),
					values.NewString("o"),
				},
				expected: "o World",
			},
			{
				name: "empty needle returns entire string",
				args: []*values.Value{
					values.NewString("Hello World"),
					values.NewString(""),
				},
				expected: "Hello World",
			},
			{
				name: "empty haystack",
				args: []*values.Value{
					values.NewString(""),
					values.NewString("hello"),
				},
				expected: false,
			},
			{
				name: "before needle option - basic",
				args: []*values.Value{
					values.NewString("Hello World"),
					values.NewString("WORLD"),
					values.NewBool(true),
				},
				expected: "Hello ",
			},
			{
				name: "before needle option - character",
				args: []*values.Value{
					values.NewString("Hello World"),
					values.NewString("o"),
					values.NewBool(true),
				},
				expected: "Hell",
			},
			{
				name: "before needle option - not found",
				args: []*values.Value{
					values.NewString("Hello World"),
					values.NewString("xyz"),
					values.NewBool(true),
				},
				expected: false,
			},
			{
				name: "before needle option - at start",
				args: []*values.Value{
					values.NewString("Hello World"),
					values.NewString("hello"),
					values.NewBool(true),
				},
				expected: "",
			},
			{
				name: "all uppercase to lowercase",
				args: []*values.Value{
					values.NewString("HELLO WORLD"),
					values.NewString("hello"),
				},
				expected: "HELLO WORLD",
			},
			{
				name: "all lowercase to uppercase",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("HELLO"),
				},
				expected: "hello world",
			},
			{
				name: "mixed case needle and haystack",
				args: []*values.Value{
					values.NewString("HeLLo WoRLd"),
					values.NewString("llo wo"),
				},
				expected: "LLo WoRLd",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				switch expected := tt.expected.(type) {
				case string:
					if result.Type != values.TypeString || result.Data.(string) != expected {
						t.Errorf("Expected %q, got %v", expected, result)
					}
				case bool:
					if result.Type != values.TypeBool || result.Data.(bool) != expected {
						t.Errorf("Expected %v, got %v", expected, result)
					}
				}
			})
		}
	})

	t.Run("strripos", func(t *testing.T) {
		fn := functionMap["strripos"]
		if fn == nil {
			t.Fatal("strripos function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected interface{}
		}{
			{
				name: "case insensitive find last occurrence uppercase",
				args: []*values.Value{
					values.NewString("Hello World Hello"),
					values.NewString("HELLO"),
				},
				expected: int64(12),
			},
			{
				name: "case insensitive find last occurrence lowercase",
				args: []*values.Value{
					values.NewString("Hello World Hello"),
					values.NewString("hello"),
				},
				expected: int64(12),
			},
			{
				name: "not found",
				args: []*values.Value{
					values.NewString("Hello World Hello"),
					values.NewString("xyz"),
				},
				expected: false,
			},
			{
				name: "case insensitive character uppercase",
				args: []*values.Value{
					values.NewString("Hello World Hello"),
					values.NewString("O"),
				},
				expected: int64(16),
			},
			{
				name: "case insensitive character lowercase",
				args: []*values.Value{
					values.NewString("Hello World Hello"),
					values.NewString("o"),
				},
				expected: int64(16),
			},
			{
				name: "case insensitive L uppercase",
				args: []*values.Value{
					values.NewString("Hello World Hello"),
					values.NewString("L"),
				},
				expected: int64(15),
			},
			{
				name: "case insensitive L lowercase",
				args: []*values.Value{
					values.NewString("Hello World Hello"),
					values.NewString("l"),
				},
				expected: int64(15),
			},
			{
				name: "empty needle",
				args: []*values.Value{
					values.NewString("Hello World"),
					values.NewString(""),
				},
				expected: int64(11),
			},
			{
				name: "empty haystack",
				args: []*values.Value{
					values.NewString(""),
					values.NewString("hello"),
				},
				expected: false,
			},
			{
				name: "needle longer than haystack",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("hello world"),
				},
				expected: false,
			},
			{
				name: "all uppercase haystack lowercase needle",
				args: []*values.Value{
					values.NewString("HELLO WORLD HELLO"),
					values.NewString("hello"),
				},
				expected: int64(12),
			},
			{
				name: "all lowercase haystack uppercase needle",
				args: []*values.Value{
					values.NewString("hello world hello"),
					values.NewString("HELLO"),
				},
				expected: int64(12),
			},
			{
				name: "mixed case haystack and needle",
				args: []*values.Value{
					values.NewString("HeLLo WoRLd HeLLo"),
					values.NewString("hello"),
				},
				expected: int64(12),
			},
			{
				name: "mixed case both ways",
				args: []*values.Value{
					values.NewString("HeLLo WoRLd HeLLo"),
					values.NewString("HELLO"),
				},
				expected: int64(12),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				switch expected := tt.expected.(type) {
				case int64:
					if result.Type != values.TypeInt || result.Data.(int64) != expected {
						t.Errorf("Expected %d, got %v", expected, result)
					}
				case bool:
					if result.Type != values.TypeBool || result.Data.(bool) != expected {
						t.Errorf("Expected %v, got %v", expected, result)
					}
				}
			})
		}
	})

	t.Run("substr_replace", func(t *testing.T) {
		fn := functionMap["substr_replace"]
		if fn == nil {
			t.Fatal("substr_replace function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected string
		}{
			{
				name: "basic replacement from start",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("hi"),
					values.NewInt(0),
					values.NewInt(5),
				},
				expected: "hi world",
			},
			{
				name: "basic replacement different text",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("HELLO"),
					values.NewInt(0),
					values.NewInt(5),
				},
				expected: "HELLO world",
			},
			{
				name: "replacement in middle",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("hi"),
					values.NewInt(6),
					values.NewInt(5),
				},
				expected: "hello hi",
			},
			{
				name: "replacement to end without length",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("universe"),
					values.NewInt(6),
				},
				expected: "hello universe",
			},
			{
				name: "negative offset from end",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("hi"),
					values.NewInt(-5),
					values.NewInt(5),
				},
				expected: "hello hi",
			},
			{
				name: "negative offset partial replacement",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("hi"),
					values.NewInt(-5),
					values.NewInt(3),
				},
				expected: "hello hild",
			},
			{
				name: "negative offset to end",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("hi"),
					values.NewInt(-5),
				},
				expected: "hello hi",
			},
			{
				name: "negative length from start",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("hi"),
					values.NewInt(0),
					values.NewInt(-1),
				},
				expected: "hid",
			},
			{
				name: "negative length from middle",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("hi"),
					values.NewInt(2),
					values.NewInt(-2),
				},
				expected: "hehild",
			},
			{
				name: "insert at position zero length",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("hi"),
					values.NewInt(0),
					values.NewInt(0),
				},
				expected: "hihello",
			},
			{
				name: "insert at end zero length",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("hi"),
					values.NewInt(5),
					values.NewInt(0),
				},
				expected: "hellohi",
			},
			{
				name: "delete characters empty replacement",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString(""),
					values.NewInt(2),
					values.NewInt(2),
				},
				expected: "heo",
			},
			{
				name: "insert into empty string",
				args: []*values.Value{
					values.NewString(""),
					values.NewString("hello"),
					values.NewInt(0),
					values.NewInt(0),
				},
				expected: "hello",
			},
			{
				name: "offset beyond string length",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("world"),
					values.NewInt(10),
					values.NewInt(5),
				},
				expected: "helloworld",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				if result.Type != values.TypeString || result.Data.(string) != tt.expected {
					t.Errorf("Expected %q, got %v", tt.expected, result)
				}
			})
		}
	})

	t.Run("strncmp", func(t *testing.T) {
		fn := functionMap["strncmp"]
		if fn == nil {
			t.Fatal("strncmp function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected int64
		}{
			{
				name: "equal strings same length",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("hello"),
					values.NewInt(5),
				},
				expected: 0,
			},
			{
				name: "first string less than second",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("world"),
					values.NewInt(5),
				},
				expected: -1,
			},
			{
				name: "first string greater than second",
				args: []*values.Value{
					values.NewString("world"),
					values.NewString("hello"),
					values.NewInt(5),
				},
				expected: 1,
			},
			{
				name: "equal partial comparison",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("hello world"),
					values.NewInt(5),
				},
				expected: 0,
			},
			{
				name: "reverse equal partial comparison",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("hello"),
					values.NewInt(5),
				},
				expected: 0,
			},
			{
				name: "equal first 3 characters",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("help"),
					values.NewInt(3),
				},
				expected: 0,
			},
			{
				name: "different at 4th character",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("help"),
					values.NewInt(4),
				},
				expected: -1,
			},
			{
				name: "zero length comparison",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("world"),
					values.NewInt(0),
				},
				expected: 0,
			},
			{
				name: "single character comparison less",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("world"),
					values.NewInt(1),
				},
				expected: -1,
			},
			{
				name: "case sensitive Hello vs hello",
				args: []*values.Value{
					values.NewString("Hello"),
					values.NewString("hello"),
					values.NewInt(5),
				},
				expected: -1,
			},
			{
				name: "case sensitive HELLO vs hello",
				args: []*values.Value{
					values.NewString("HELLO"),
					values.NewString("hello"),
					values.NewInt(5),
				},
				expected: -1,
			},
			{
				name: "case sensitive hello vs HELLO",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("HELLO"),
					values.NewInt(5),
				},
				expected: 1,
			},
			{
				name: "both empty strings zero length",
				args: []*values.Value{
					values.NewString(""),
					values.NewString(""),
					values.NewInt(0),
				},
				expected: 0,
			},
			{
				name: "both empty strings non-zero length",
				args: []*values.Value{
					values.NewString(""),
					values.NewString(""),
					values.NewInt(5),
				},
				expected: 0,
			},
			{
				name: "first non-empty second empty",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString(""),
					values.NewInt(5),
				},
				expected: 1,
			},
			{
				name: "first empty second non-empty",
				args: []*values.Value{
					values.NewString(""),
					values.NewString("hello"),
					values.NewInt(5),
				},
				expected: -1,
			},
			{
				name: "single char a vs b",
				args: []*values.Value{
					values.NewString("a"),
					values.NewString("b"),
					values.NewInt(1),
				},
				expected: -1,
			},
			{
				name: "single char b vs a",
				args: []*values.Value{
					values.NewString("b"),
					values.NewString("a"),
					values.NewInt(1),
				},
				expected: 1,
			},
			{
				name: "length larger than both strings equal",
				args: []*values.Value{
					values.NewString("abc"),
					values.NewString("abc"),
					values.NewInt(10),
				},
				expected: 0,
			},
			{
				name: "length larger first shorter",
				args: []*values.Value{
					values.NewString("abc"),
					values.NewString("abcd"),
					values.NewInt(10),
				},
				expected: -1,
			},
			{
				name: "length larger second shorter",
				args: []*values.Value{
					values.NewString("abcd"),
					values.NewString("abc"),
					values.NewInt(10),
				},
				expected: 1,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				resultInt := result.Data.(int64)
				// Normalize to -1, 0, 1 for comparison
				if resultInt < 0 {
					resultInt = -1
				} else if resultInt > 0 {
					resultInt = 1
				}

				if result.Type != values.TypeInt || resultInt != tt.expected {
					t.Errorf("Expected %d, got %v", tt.expected, result)
				}
			})
		}
	})

	t.Run("strncasecmp", func(t *testing.T) {
		fn := functionMap["strncasecmp"]
		if fn == nil {
			t.Fatal("strncasecmp function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected int64
		}{
			{
				name: "identical strings",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("hello"),
					values.NewInt(5),
				},
				expected: 0,
			},
			{
				name: "case insensitive equal Hello vs hello",
				args: []*values.Value{
					values.NewString("Hello"),
					values.NewString("hello"),
					values.NewInt(5),
				},
				expected: 0,
			},
			{
				name: "case insensitive equal HELLO vs hello",
				args: []*values.Value{
					values.NewString("HELLO"),
					values.NewString("hello"),
					values.NewInt(5),
				},
				expected: 0,
			},
			{
				name: "case insensitive equal hello vs HELLO",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("HELLO"),
					values.NewInt(5),
				},
				expected: 0,
			},
			{
				name: "case insensitive hello vs world",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("world"),
					values.NewInt(5),
				},
				expected: -1,
			},
			{
				name: "case insensitive world vs hello",
				args: []*values.Value{
					values.NewString("world"),
					values.NewString("hello"),
					values.NewInt(5),
				},
				expected: 1,
			},
			{
				name: "partial comparison Hello vs Help 3 chars",
				args: []*values.Value{
					values.NewString("Hello"),
					values.NewString("Help"),
					values.NewInt(3),
				},
				expected: 0,
			},
			{
				name: "partial comparison Hello vs Help 4 chars",
				args: []*values.Value{
					values.NewString("Hello"),
					values.NewString("Help"),
					values.NewInt(4),
				},
				expected: -1,
			},
			{
				name: "partial comparison HELLO vs help 3 chars",
				args: []*values.Value{
					values.NewString("HELLO"),
					values.NewString("help"),
					values.NewInt(3),
				},
				expected: 0,
			},
			{
				name: "partial comparison HELLO vs help 4 chars",
				args: []*values.Value{
					values.NewString("HELLO"),
					values.NewString("help"),
					values.NewInt(4),
				},
				expected: -1,
			},
			{
				name: "zero length comparison",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("WORLD"),
					values.NewInt(0),
				},
				expected: 0,
			},
			{
				name: "one char comparison hello vs WORLD",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("WORLD"),
					values.NewInt(1),
				},
				expected: -1,
			},
			{
				name: "both empty zero length",
				args: []*values.Value{
					values.NewString(""),
					values.NewString(""),
					values.NewInt(0),
				},
				expected: 0,
			},
			{
				name: "both empty with length",
				args: []*values.Value{
					values.NewString(""),
					values.NewString(""),
					values.NewInt(5),
				},
				expected: 0,
			},
			{
				name: "first non-empty second empty",
				args: []*values.Value{
					values.NewString("HELLO"),
					values.NewString(""),
					values.NewInt(5),
				},
				expected: 1,
			},
			{
				name: "first empty second non-empty",
				args: []*values.Value{
					values.NewString(""),
					values.NewString("hello"),
					values.NewInt(5),
				},
				expected: -1,
			},
			{
				name: "single char A vs b",
				args: []*values.Value{
					values.NewString("A"),
					values.NewString("b"),
					values.NewInt(1),
				},
				expected: -1,
			},
			{
				name: "single char B vs a",
				args: []*values.Value{
					values.NewString("B"),
					values.NewString("a"),
					values.NewInt(1),
				},
				expected: 1,
			},
			{
				name: "length larger than both strings equal",
				args: []*values.Value{
					values.NewString("ABC"),
					values.NewString("abc"),
					values.NewInt(10),
				},
				expected: 0,
			},
			{
				name: "length larger first shorter",
				args: []*values.Value{
					values.NewString("ABC"),
					values.NewString("abcd"),
					values.NewInt(10),
				},
				expected: -1,
			},
			{
				name: "length larger second shorter",
				args: []*values.Value{
					values.NewString("ABCD"),
					values.NewString("abc"),
					values.NewInt(10),
				},
				expected: 1,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				resultInt := result.Data.(int64)
				// Normalize to -1, 0, 1 for comparison
				if resultInt < 0 {
					resultInt = -1
				} else if resultInt > 0 {
					resultInt = 1
				}

				if result.Type != values.TypeInt || resultInt != tt.expected {
					t.Errorf("Expected %d, got %v", tt.expected, result)
				}
			})
		}
	})

	t.Run("strcasecmp", func(t *testing.T) {
		fn := functionMap["strcasecmp"]
		if fn == nil {
			t.Fatal("strcasecmp function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected int64
		}{
			{
				name: "identical strings",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("hello"),
				},
				expected: 0,
			},
			{
				name: "case insensitive equal Hello vs hello",
				args: []*values.Value{
					values.NewString("Hello"),
					values.NewString("hello"),
				},
				expected: 0,
			},
			{
				name: "case insensitive equal HELLO vs hello",
				args: []*values.Value{
					values.NewString("HELLO"),
					values.NewString("hello"),
				},
				expected: 0,
			},
			{
				name: "case insensitive equal hello vs HELLO",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("HELLO"),
				},
				expected: 0,
			},
			{
				name: "case insensitive hello vs world",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("world"),
				},
				expected: -1,
			},
			{
				name: "case insensitive world vs hello",
				args: []*values.Value{
					values.NewString("world"),
					values.NewString("hello"),
				},
				expected: 1,
			},
			{
				name: "mixed case HeLLo vs hello",
				args: []*values.Value{
					values.NewString("HeLLo"),
					values.NewString("hello"),
				},
				expected: 0,
			},
			{
				name: "mixed case hello vs HeLLo",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("HeLLo"),
				},
				expected: 0,
			},
			{
				name: "mixed case ABC vs abc",
				args: []*values.Value{
					values.NewString("ABC"),
					values.NewString("abc"),
				},
				expected: 0,
			},
			{
				name: "mixed case abc vs ABC",
				args: []*values.Value{
					values.NewString("abc"),
					values.NewString("ABC"),
				},
				expected: 0,
			},
			{
				name: "length difference hello vs hello world",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("hello world"),
				},
				expected: -1,
			},
			{
				name: "length difference hello world vs hello",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("hello"),
				},
				expected: 1,
			},
			{
				name: "length difference abc vs abcd",
				args: []*values.Value{
					values.NewString("abc"),
					values.NewString("abcd"),
				},
				expected: -1,
			},
			{
				name: "length difference abcd vs abc",
				args: []*values.Value{
					values.NewString("abcd"),
					values.NewString("abc"),
				},
				expected: 1,
			},
			{
				name: "both empty",
				args: []*values.Value{
					values.NewString(""),
					values.NewString(""),
				},
				expected: 0,
			},
			{
				name: "first non-empty second empty",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString(""),
				},
				expected: 1,
			},
			{
				name: "first empty second non-empty",
				args: []*values.Value{
					values.NewString(""),
					values.NewString("hello"),
				},
				expected: -1,
			},
			{
				name: "single char a vs b",
				args: []*values.Value{
					values.NewString("a"),
					values.NewString("b"),
				},
				expected: -1,
			},
			{
				name: "single char b vs a",
				args: []*values.Value{
					values.NewString("b"),
					values.NewString("a"),
				},
				expected: 1,
			},
			{
				name: "single char A vs b",
				args: []*values.Value{
					values.NewString("A"),
					values.NewString("b"),
				},
				expected: -1,
			},
			{
				name: "single char B vs a",
				args: []*values.Value{
					values.NewString("B"),
					values.NewString("a"),
				},
				expected: 1,
			},
			{
				name: "numbers and letters hello123 vs HELLO123",
				args: []*values.Value{
					values.NewString("hello123"),
					values.NewString("HELLO123"),
				},
				expected: 0,
			},
			{
				name: "special chars hello! vs HELLO!",
				args: []*values.Value{
					values.NewString("hello!"),
					values.NewString("HELLO!"),
				},
				expected: 0,
			},
			{
				name: "numbers only",
				args: []*values.Value{
					values.NewString("123"),
					values.NewString("123"),
				},
				expected: 0,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				resultInt := result.Data.(int64)
				// Normalize to -1, 0, 1 for comparison
				if resultInt < 0 {
					resultInt = -1
				} else if resultInt > 0 {
					resultInt = 1
				}

				if result.Type != values.TypeInt || resultInt != tt.expected {
					t.Errorf("Expected %d, got %v", tt.expected, result)
				}
			})
		}
	})

	t.Run("str_contains", func(t *testing.T) {
		fn := functionMap["str_contains"]
		if fn == nil {
			t.Fatal("str_contains function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected bool
		}{
			{
				name: "contains hello",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("hello"),
				},
				expected: true,
			},
			{
				name: "contains world",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("world"),
				},
				expected: true,
			},
			{
				name: "contains substring",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("lo wo"),
				},
				expected: true,
			},
			{
				name: "does not contain xyz",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("xyz"),
				},
				expected: false,
			},
			{
				name: "contains empty string",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString(""),
				},
				expected: true,
			},
			{
				name: "case sensitive Hello vs hello",
				args: []*values.Value{
					values.NewString("Hello World"),
					values.NewString("hello"),
				},
				expected: false,
			},
			{
				name: "case sensitive Hello exact match",
				args: []*values.Value{
					values.NewString("Hello World"),
					values.NewString("Hello"),
				},
				expected: true,
			},
			{
				name: "case sensitive HELLO vs hello",
				args: []*values.Value{
					values.NewString("HELLO WORLD"),
					values.NewString("hello"),
				},
				expected: false,
			},
			{
				name: "case sensitive hello vs HELLO",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("HELLO"),
				},
				expected: false,
			},
			{
				name: "both empty strings",
				args: []*values.Value{
					values.NewString(""),
					values.NewString(""),
				},
				expected: true,
			},
			{
				name: "empty haystack non-empty needle",
				args: []*values.Value{
					values.NewString(""),
					values.NewString("hello"),
				},
				expected: false,
			},
			{
				name: "identical strings",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("hello"),
				},
				expected: true,
			},
			{
				name: "needle longer than haystack",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("hello world"),
				},
				expected: false,
			},
			{
				name: "single char h",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("h"),
				},
				expected: true,
			},
			{
				name: "single char o",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("o"),
				},
				expected: true,
			},
			{
				name: "single char x not found",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("x"),
				},
				expected: false,
			},
			{
				name: "single char exact match",
				args: []*values.Value{
					values.NewString("a"),
					values.NewString("a"),
				},
				expected: true,
			},
			{
				name: "single char no match",
				args: []*values.Value{
					values.NewString("a"),
					values.NewString("b"),
				},
				expected: false,
			},
			{
				name: "special char exclamation",
				args: []*values.Value{
					values.NewString("hello!@#"),
					values.NewString("!"),
				},
				expected: true,
			},
			{
				name: "special chars at-hash",
				args: []*values.Value{
					values.NewString("hello!@#"),
					values.NewString("@#"),
				},
				expected: true,
			},
			{
				name: "numbers 123",
				args: []*values.Value{
					values.NewString("hello123"),
					values.NewString("123"),
				},
				expected: true,
			},
			{
				name: "newline character",
				args: []*values.Value{
					values.NewString("hello\nworld"),
					values.NewString("\n"),
				},
				expected: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				if result.Type != values.TypeBool || result.Data.(bool) != tt.expected {
					t.Errorf("Expected %v, got %v", tt.expected, result)
				}
			})
		}
	})

	t.Run("str_starts_with", func(t *testing.T) {
		fn := functionMap["str_starts_with"]
		if fn == nil {
			t.Fatal("str_starts_with function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected bool
		}{
			{
				name: "starts with hello",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("hello"),
				},
				expected: true,
			},
			{
				name: "does not start with world",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("world"),
				},
				expected: false,
			},
			{
				name: "starts with he",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("he"),
				},
				expected: true,
			},
			{
				name: "does not start with xyz",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("xyz"),
				},
				expected: false,
			},
			{
				name: "starts with empty string",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString(""),
				},
				expected: true,
			},
			{
				name: "case sensitive Hello vs hello",
				args: []*values.Value{
					values.NewString("Hello World"),
					values.NewString("hello"),
				},
				expected: false,
			},
			{
				name: "case sensitive Hello exact match",
				args: []*values.Value{
					values.NewString("Hello World"),
					values.NewString("Hello"),
				},
				expected: true,
			},
			{
				name: "case sensitive HELLO vs hello",
				args: []*values.Value{
					values.NewString("HELLO WORLD"),
					values.NewString("hello"),
				},
				expected: false,
			},
			{
				name: "case sensitive hello vs HELLO",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("HELLO"),
				},
				expected: false,
			},
			{
				name: "both empty strings",
				args: []*values.Value{
					values.NewString(""),
					values.NewString(""),
				},
				expected: true,
			},
			{
				name: "empty haystack non-empty needle",
				args: []*values.Value{
					values.NewString(""),
					values.NewString("hello"),
				},
				expected: false,
			},
			{
				name: "identical strings",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("hello"),
				},
				expected: true,
			},
			{
				name: "needle longer than haystack",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("hello world"),
				},
				expected: false,
			},
			{
				name: "single char h",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("h"),
				},
				expected: true,
			},
			{
				name: "single char o not at start",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("o"),
				},
				expected: false,
			},
			{
				name: "single char x not found",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("x"),
				},
				expected: false,
			},
			{
				name: "single char exact match",
				args: []*values.Value{
					values.NewString("a"),
					values.NewString("a"),
				},
				expected: true,
			},
			{
				name: "single char no match",
				args: []*values.Value{
					values.NewString("a"),
					values.NewString("b"),
				},
				expected: false,
			},
			{
				name: "special char exclamation at start",
				args: []*values.Value{
					values.NewString("!hello"),
					values.NewString("!"),
				},
				expected: true,
			},
			{
				name: "special chars at-hash at start",
				args: []*values.Value{
					values.NewString("@#hello"),
					values.NewString("@#"),
				},
				expected: true,
			},
			{
				name: "numbers 123 at start",
				args: []*values.Value{
					values.NewString("123hello"),
					values.NewString("123"),
				},
				expected: true,
			},
			{
				name: "newline character at start",
				args: []*values.Value{
					values.NewString("\nhello"),
					values.NewString("\n"),
				},
				expected: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				if result.Type != values.TypeBool || result.Data.(bool) != tt.expected {
					t.Errorf("Expected %v, got %v", tt.expected, result)
				}
			})
		}
	})

	t.Run("str_ends_with", func(t *testing.T) {
		fn := functionMap["str_ends_with"]
		if fn == nil {
			t.Fatal("str_ends_with function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected bool
		}{
			{
				name: "ends with world",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("world"),
				},
				expected: true,
			},
			{
				name: "does not end with hello",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("hello"),
				},
				expected: false,
			},
			{
				name: "ends with ld",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("ld"),
				},
				expected: true,
			},
			{
				name: "does not end with xyz",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("xyz"),
				},
				expected: false,
			},
			{
				name: "ends with empty string",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString(""),
				},
				expected: true,
			},
			{
				name: "case sensitive World vs world",
				args: []*values.Value{
					values.NewString("Hello World"),
					values.NewString("world"),
				},
				expected: false,
			},
			{
				name: "case sensitive World exact match",
				args: []*values.Value{
					values.NewString("Hello World"),
					values.NewString("World"),
				},
				expected: true,
			},
			{
				name: "case sensitive WORLD vs world",
				args: []*values.Value{
					values.NewString("HELLO WORLD"),
					values.NewString("world"),
				},
				expected: false,
			},
			{
				name: "case sensitive world vs WORLD",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("WORLD"),
				},
				expected: false,
			},
			{
				name: "both empty strings",
				args: []*values.Value{
					values.NewString(""),
					values.NewString(""),
				},
				expected: true,
			},
			{
				name: "empty haystack non-empty needle",
				args: []*values.Value{
					values.NewString(""),
					values.NewString("hello"),
				},
				expected: false,
			},
			{
				name: "identical strings",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("hello"),
				},
				expected: true,
			},
			{
				name: "needle longer than haystack",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("hello world"),
				},
				expected: false,
			},
			{
				name: "single char o at end",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("o"),
				},
				expected: true,
			},
			{
				name: "single char h not at end",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("h"),
				},
				expected: false,
			},
			{
				name: "single char x not found",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("x"),
				},
				expected: false,
			},
			{
				name: "single char exact match",
				args: []*values.Value{
					values.NewString("a"),
					values.NewString("a"),
				},
				expected: true,
			},
			{
				name: "single char no match",
				args: []*values.Value{
					values.NewString("a"),
					values.NewString("b"),
				},
				expected: false,
			},
			{
				name: "special char exclamation at end",
				args: []*values.Value{
					values.NewString("hello!"),
					values.NewString("!"),
				},
				expected: true,
			},
			{
				name: "special chars at-hash at end",
				args: []*values.Value{
					values.NewString("hello@#"),
					values.NewString("@#"),
				},
				expected: true,
			},
			{
				name: "numbers 123 at end",
				args: []*values.Value{
					values.NewString("hello123"),
					values.NewString("123"),
				},
				expected: true,
			},
			{
				name: "newline character at end",
				args: []*values.Value{
					values.NewString("hello\n"),
					values.NewString("\n"),
				},
				expected: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				if result.Type != values.TypeBool || result.Data.(bool) != tt.expected {
					t.Errorf("Expected %v, got %v", tt.expected, result)
				}
			})
		}
	})

	t.Run("strchr", func(t *testing.T) {
		fn := functionMap["strchr"]
		if fn == nil {
			t.Fatal("strchr function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected interface{}
		}{
			{
				name: "basic find wo",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("wo"),
				},
				expected: "world",
			},
			{
				name: "not found xyz",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("xyz"),
				},
				expected: false,
			},
			{
				name: "empty needle",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString(""),
				},
				expected: "hello world",
			},
			{
				name: "before needle true",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("wo"),
					values.NewBool(true),
				},
				expected: "hello ",
			},
			{
				name: "before needle false",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("wo"),
					values.NewBool(false),
				},
				expected: "world",
			},
			{
				name: "both empty",
				args: []*values.Value{
					values.NewString(""),
					values.NewString(""),
				},
				expected: "",
			},
			{
				name: "identical strings",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("hello"),
				},
				expected: "hello",
			},
			{
				name: "needle at start",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("hello"),
				},
				expected: "hello world",
			},
			{
				name: "needle at end",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewString("world"),
				},
				expected: "world",
			},
			{
				name: "single character",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("l"),
				},
				expected: "llo",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				if tt.expected == false {
					if result.Type != values.TypeBool || result.Data.(bool) != false {
						t.Errorf("Expected false, got %v", result)
					}
				} else {
					if result.Type != values.TypeString || result.Data.(string) != tt.expected {
						t.Errorf("Expected %s, got %v", tt.expected, result)
					}
				}
			})
		}
	})

	t.Run("str_word_count", func(t *testing.T) {
		fn := functionMap["str_word_count"]
		if fn == nil {
			t.Fatal("str_word_count function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected interface{}
		}{
			{
				name: "basic count default format",
				args: []*values.Value{
					values.NewString("hello world"),
				},
				expected: int64(2),
			},
			{
				name: "basic count format 0",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewInt(0),
				},
				expected: int64(2),
			},
			{
				name: "words array format 1",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewInt(1),
				},
				expected: []string{"hello", "world"},
			},
			{
				name: "words with positions format 2",
				args: []*values.Value{
					values.NewString("hello world"),
					values.NewInt(2),
				},
				expected: map[int64]string{0: "hello", 6: "world"},
			},
			{
				name: "multiple spaces",
				args: []*values.Value{
					values.NewString("hello  world"),
				},
				expected: int64(2),
			},
			{
				name: "punctuation",
				args: []*values.Value{
					values.NewString("hello, world!"),
				},
				expected: int64(2),
			},
			{
				name: "punctuation format 1",
				args: []*values.Value{
					values.NewString("hello, world!"),
					values.NewInt(1),
				},
				expected: []string{"hello", "world"},
			},
			{
				name: "punctuation format 2",
				args: []*values.Value{
					values.NewString("hello, world!"),
					values.NewInt(2),
				},
				expected: map[int64]string{0: "hello", 7: "world"},
			},
			{
				name: "empty string",
				args: []*values.Value{
					values.NewString(""),
				},
				expected: int64(0),
			},
			{
				name: "whitespace only",
				args: []*values.Value{
					values.NewString("   "),
				},
				expected: int64(0),
			},
			{
				name: "single word",
				args: []*values.Value{
					values.NewString("word"),
				},
				expected: int64(1),
			},
			{
				name: "numbers only",
				args: []*values.Value{
					values.NewString("123"),
				},
				expected: int64(0),
			},
			{
				name: "mixed alphanumeric",
				args: []*values.Value{
					values.NewString("hello123world"),
				},
				expected: int64(2),
			},
			{
				name: "custom charlist dash",
				args: []*values.Value{
					values.NewString("hello-world"),
					values.NewInt(0),
					values.NewString("-"),
				},
				expected: int64(1),
			},
			{
				name: "custom charlist dash format 1",
				args: []*values.Value{
					values.NewString("hello-world"),
					values.NewInt(1),
					values.NewString("-"),
				},
				expected: []string{"hello-world"},
			},
			{
				name: "custom charlist underscore",
				args: []*values.Value{
					values.NewString("hello_world"),
					values.NewInt(1),
					values.NewString("_"),
				},
				expected: []string{"hello_world"},
			},
			{
				name: "alphanumeric separation",
				args: []*values.Value{
					values.NewString("hello123 world456"),
					values.NewInt(1),
				},
				expected: []string{"hello", "world"},
			},
			{
				name: "special chars only",
				args: []*values.Value{
					values.NewString("!@# $%^ &*()"),
				},
				expected: int64(0),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				switch expected := tt.expected.(type) {
				case int64:
					if result.Type != values.TypeInt || result.Data.(int64) != expected {
						t.Errorf("Expected %d, got %v", expected, result)
					}
				case []string:
					if result.Type != values.TypeArray {
						t.Errorf("Expected array, got %v", result.Type)
						return
					}
					arrayData := result.Data.(*values.Array)
					if len(arrayData.Elements) != len(expected) {
						t.Errorf("Expected %d elements, got %d", len(expected), len(arrayData.Elements))
						return
					}
					for i, expectedStr := range expected {
						element, exists := arrayData.Elements[int64(i)]
						if !exists {
							t.Errorf("Expected element at index %d not found", i)
							continue
						}
						if element.Type != values.TypeString || element.Data.(string) != expectedStr {
							t.Errorf("Expected element %d to be %s, got %v", i, expectedStr, element)
						}
					}
				case map[int64]string:
					if result.Type != values.TypeArray {
						t.Errorf("Expected array, got %v", result.Type)
						return
					}
					arrayData := result.Data.(*values.Array)
					if len(arrayData.Elements) != len(expected) {
						t.Errorf("Expected %d elements, got %d", len(expected), len(arrayData.Elements))
						return
					}
					// For format 2, check elements at specific positions
					for key, expectedStr := range expected {
						element, exists := arrayData.Elements[key]
						if !exists {
							t.Errorf("Expected key %d not found in result", key)
							continue
						}
						if element.Type != values.TypeString || element.Data.(string) != expectedStr {
							t.Errorf("Expected element at key %d to be %s, got %v", key, expectedStr, element)
						}
					}
				}
			})
		}
	})

	t.Run("htmlspecialchars", func(t *testing.T) {
		fn := functionMap["htmlspecialchars"]
		if fn == nil {
			t.Fatal("htmlspecialchars function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected string
		}{
			{
				name: "basic XSS prevention",
				args: []*values.Value{
					values.NewString(`<script>alert("XSS")</script>`),
				},
				expected: `&lt;script&gt;alert(&quot;XSS&quot;)&lt;/script&gt;`,
			},
			{
				name: "ampersand conversion",
				args: []*values.Value{
					values.NewString("Hello & World"),
				},
				expected: "Hello &amp; World",
			},
			{
				name: "double quotes default",
				args: []*values.Value{
					values.NewString(`"quoted"`),
				},
				expected: "&quot;quoted&quot;",
			},
			{
				name: "single quotes default",
				args: []*values.Value{
					values.NewString("'single'"),
				},
				expected: "&#039;single&#039;",
			},
			{
				name: "ENT_COMPAT quotes",
				args: []*values.Value{
					values.NewString(`"double" & 'single'`),
					values.NewInt(2), // ENT_COMPAT
				},
				expected: `&quot;double&quot; &amp; 'single'`,
			},
			{
				name: "ENT_QUOTES both quotes",
				args: []*values.Value{
					values.NewString(`"double" & 'single'`),
					values.NewInt(3), // ENT_QUOTES
				},
				expected: `&quot;double&quot; &amp; &#039;single&#039;`,
			},
			{
				name: "ENT_NOQUOTES no quotes",
				args: []*values.Value{
					values.NewString(`"double" & 'single'`),
					values.NewInt(0), // ENT_NOQUOTES
				},
				expected: `"double" &amp; 'single'`,
			},
			{
				name: "double encode true default",
				args: []*values.Value{
					values.NewString("&amp; &lt;"),
				},
				expected: "&amp;amp; &amp;lt;",
			},
			{
				name: "double encode false",
				args: []*values.Value{
					values.NewString("&amp; &lt;"),
					values.NewInt(2), // ENT_COMPAT
					values.NewString("UTF-8"),
					values.NewBool(false),
				},
				expected: "&amp; &lt;",
			},
			{
				name: "empty string",
				args: []*values.Value{
					values.NewString(""),
				},
				expected: "",
			},
			{
				name: "normal text",
				args: []*values.Value{
					values.NewString("normal text"),
				},
				expected: "normal text",
			},
			{
				name: "angle brackets only",
				args: []*values.Value{
					values.NewString("<>"),
				},
				expected: "&lt;&gt;",
			},
			{
				name: "ampersand only",
				args: []*values.Value{
					values.NewString("&"),
				},
				expected: "&amp;",
			},
			{
				name: "complex HTML",
				args: []*values.Value{
					values.NewString(`<tag attr="value"> & text </tag>`),
				},
				expected: `&lt;tag attr=&quot;value&quot;&gt; &amp; text &lt;/tag&gt;`,
			},
			{
				name: "comparison operators",
				args: []*values.Value{
					values.NewString("10 > 5 & 3 < 7"),
				},
				expected: "10 &gt; 5 &amp; 3 &lt; 7",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				if result.Type != values.TypeString || result.Data.(string) != tt.expected {
					t.Errorf("Expected %s, got %v", tt.expected, result)
				}
			})
		}
	})

	t.Run("urlencode", func(t *testing.T) {
		fn := functionMap["urlencode"]
		if fn == nil {
			t.Fatal("urlencode function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected string
		}{
			{
				name: "space to plus",
				args: []*values.Value{
					values.NewString("hello world"),
				},
				expected: "hello+world",
			},
			{
				name: "plus sign encoding",
				args: []*values.Value{
					values.NewString("hello+world"),
				},
				expected: "hello%2Bworld",
			},
			{
				name: "ampersand encoding",
				args: []*values.Value{
					values.NewString("hello&world"),
				},
				expected: "hello%26world",
			},
			{
				name: "equals encoding",
				args: []*values.Value{
					values.NewString("hello=world"),
				},
				expected: "hello%3Dworld",
			},
			{
				name: "question mark encoding",
				args: []*values.Value{
					values.NewString("hello?world"),
				},
				expected: "hello%3Fworld",
			},
			{
				name: "special symbols",
				args: []*values.Value{
					values.NewString("@#$%^&*()"),
				},
				expected: "%40%23%24%25%5E%26%2A%28%29",
			},
			{
				name: "brackets and pipes",
				args: []*values.Value{
					values.NewString("[]{}|\\<>"),
				},
				expected: "%5B%5D%7B%7D%7C%5C%3C%3E",
			},
			{
				name: "quotes",
				args: []*values.Value{
					values.NewString("\"'"),
				},
				expected: "%22%27",
			},
			{
				name: "safe alphanumeric",
				args: []*values.Value{
					values.NewString("abc123"),
				},
				expected: "abc123",
			},
			{
				name: "safe uppercase alphanumeric",
				args: []*values.Value{
					values.NewString("ABC123"),
				},
				expected: "ABC123",
			},
			{
				name: "safe characters with tilde",
				args: []*values.Value{
					values.NewString("-_.~"),
				},
				expected: "-_.%7E",
			},
			{
				name: "UTF-8 accented",
				args: []*values.Value{
					values.NewString("héllo"),
				},
				expected: "h%C3%A9llo",
			},
			{
				name: "UTF-8 Chinese",
				args: []*values.Value{
					values.NewString("你好"),
				},
				expected: "%E4%BD%A0%E5%A5%BD",
			},
			{
				name: "UTF-8 cafe",
				args: []*values.Value{
					values.NewString("café"),
				},
				expected: "caf%C3%A9",
			},
			{
				name: "empty string",
				args: []*values.Value{
					values.NewString(""),
				},
				expected: "",
			},
			{
				name: "single space",
				args: []*values.Value{
					values.NewString(" "),
				},
				expected: "+",
			},
			{
				name: "control characters",
				args: []*values.Value{
					values.NewString("\n\r\t"),
				},
				expected: "%0A%0D%09",
			},
			{
				name: "email address",
				args: []*values.Value{
					values.NewString("user@example.com"),
				},
				expected: "user%40example.com",
			},
			{
				name: "full URL",
				args: []*values.Value{
					values.NewString("http://example.com/path?query=value"),
				},
				expected: "http%3A%2F%2Fexample.com%2Fpath%3Fquery%3Dvalue",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				if result.Type != values.TypeString || result.Data.(string) != tt.expected {
					t.Errorf("Expected %s, got %v", tt.expected, result)
				}
			})
		}
	})

	t.Run("urldecode", func(t *testing.T) {
		fn := functionMap["urldecode"]
		if fn == nil {
			t.Fatal("urldecode function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected string
		}{
			{
				name: "plus to space",
				args: []*values.Value{
					values.NewString("hello+world"),
				},
				expected: "hello world",
			},
			{
				name: "encoded plus sign",
				args: []*values.Value{
					values.NewString("hello%2Bworld"),
				},
				expected: "hello+world",
			},
			{
				name: "encoded ampersand",
				args: []*values.Value{
					values.NewString("hello%26world"),
				},
				expected: "hello&world",
			},
			{
				name: "encoded equals",
				args: []*values.Value{
					values.NewString("hello%3Dworld"),
				},
				expected: "hello=world",
			},
			{
				name: "encoded question mark",
				args: []*values.Value{
					values.NewString("hello%3Fworld"),
				},
				expected: "hello?world",
			},
			{
				name: "special symbols",
				args: []*values.Value{
					values.NewString("%40%23%24%25%5E%26%2A%28%29"),
				},
				expected: "@#$%^&*()",
			},
			{
				name: "brackets and pipes",
				args: []*values.Value{
					values.NewString("%5B%5D%7B%7D%7C%5C%3C%3E"),
				},
				expected: "[]{}|\\<>",
			},
			{
				name: "quotes",
				args: []*values.Value{
					values.NewString("%22%27"),
				},
				expected: "\"'",
			},
			{
				name: "safe alphanumeric unchanged",
				args: []*values.Value{
					values.NewString("abc123"),
				},
				expected: "abc123",
			},
			{
				name: "safe uppercase unchanged",
				args: []*values.Value{
					values.NewString("ABC123"),
				},
				expected: "ABC123",
			},
			{
				name: "safe chars with encoded tilde",
				args: []*values.Value{
					values.NewString("-_.%7E"),
				},
				expected: "-_.~",
			},
			{
				name: "UTF-8 accented",
				args: []*values.Value{
					values.NewString("h%C3%A9llo"),
				},
				expected: "héllo",
			},
			{
				name: "UTF-8 Chinese",
				args: []*values.Value{
					values.NewString("%E4%BD%A0%E5%A5%BD"),
				},
				expected: "你好",
			},
			{
				name: "UTF-8 cafe",
				args: []*values.Value{
					values.NewString("caf%C3%A9"),
				},
				expected: "café",
			},
			{
				name: "empty string",
				args: []*values.Value{
					values.NewString(""),
				},
				expected: "",
			},
			{
				name: "plus only",
				args: []*values.Value{
					values.NewString("+"),
				},
				expected: " ",
			},
			{
				name: "control characters",
				args: []*values.Value{
					values.NewString("%0A%0D%09"),
				},
				expected: "\n\r\t",
			},
			{
				name: "malformed percent only",
				args: []*values.Value{
					values.NewString("%"),
				},
				expected: "%",
			},
			{
				name: "malformed incomplete hex",
				args: []*values.Value{
					values.NewString("%2"),
				},
				expected: "%2",
			},
			{
				name: "malformed invalid hex",
				args: []*values.Value{
					values.NewString("%ZZ"),
				},
				expected: "%ZZ",
			},
			{
				name: "normal with encoded space",
				args: []*values.Value{
					values.NewString("normal%20text"),
				},
				expected: "normal text",
			},
			{
				name: "email address",
				args: []*values.Value{
					values.NewString("user%40example.com"),
				},
				expected: "user@example.com",
			},
			{
				name: "full URL",
				args: []*values.Value{
					values.NewString("http%3A%2F%2Fexample.com%2Fpath%3Fquery%3Dvalue"),
				},
				expected: "http://example.com/path?query=value",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				if result.Type != values.TypeString || result.Data.(string) != tt.expected {
					t.Errorf("Expected %s, got %v", tt.expected, result)
				}
			})
		}
	})

	// Test base64_encode function
	t.Run("base64_encode", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			expected string
		}{
			// Basic tests
			{"basic_hello_world", "hello world", "aGVsbG8gd29ybGQ="},
			{"basic_Hello_World", "Hello World", "SGVsbG8gV29ybGQ="},
			{"basic_numbers", "123456", "MTIzNDU2"},
			{"basic_abc", "abc", "YWJj"},

			// Special characters
			{"special_symbols", "@#$%^&*()", "QCMkJV4mKigp"},
			{"special_extended_symbols", "!@#$%^&*()_+", "IUAjJCVeJiooKV8r"},
			{"special_brackets_pipes", "[]{}|\\<>", "W117fXxcPD4="},
			{"special_quote", "\"", "Ig=="},

			// Binary/control characters
			{"binary_null", "\x00", "AA=="},
			{"binary_max", "\xff", "/w=="},
			{"control_chars", "\n\r\t", "Cg0J"},
			{"binary_sequence", "\x00\x01\x02\x03", "AAECAw=="},

			// UTF-8 characters
			{"utf8_accented", "héllo", "aMOpbGxv"},
			{"utf8_chinese", "你好", "5L2g5aW9"},
			{"utf8_cafe", "café", "Y2Fmw6k="},
			{"utf8_emoji", "🌟", "8J+Mnw=="},

			// Edge cases
			{"empty_string", "", ""},
			{"single_char", "a", "YQ=="},
			{"two_chars", "ab", "YWI="},
			{"three_chars", "abc", "YWJj"},
			{"four_chars", "abcd", "YWJjZA=="},

			// Longer strings
			{"long_sentence", "The quick brown fox jumps over the lazy dog", "VGhlIHF1aWNrIGJyb3duIGZveCBqdW1wcyBvdmVyIHRoZSBsYXp5IGRvZw=="},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				funcs := GetStringFunctions()
				var base64EncodeFunc *registry.Function
				for _, f := range funcs {
					if f.Name == "base64_encode" {
						base64EncodeFunc = f
						break
					}
				}

				if base64EncodeFunc == nil {
					t.Fatal("base64_encode function not found")
				}

				result, err := base64EncodeFunc.Builtin(nil, []*values.Value{values.NewString(tt.input)})

				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				if result.Type != values.TypeString || result.Data.(string) != tt.expected {
					t.Errorf("Expected %s, got %v", tt.expected, result)
				}
			})
		}
	})

	// Test base64_decode function
	t.Run("base64_decode", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			strict   *bool // nil means don't provide strict parameter
			expected interface{} // string for success, bool false for strict mode failure
		}{
			// Basic tests (reverse of base64_encode)
			{"basic_hello_world", "aGVsbG8gd29ybGQ=", nil, "hello world"},
			{"basic_Hello_World", "SGVsbG8gV29ybGQ=", nil, "Hello World"},
			{"basic_numbers", "MTIzNDU2", nil, "123456"},
			{"basic_abc", "YWJj", nil, "abc"},

			// Special characters
			{"special_symbols", "QCMkJV4mKigp", nil, "@#$%^&*()"},
			{"special_extended_symbols", "IUAjJCVeJiooKV8r", nil, "!@#$%^&*()_+"},
			{"special_brackets_pipes", "W117fXxcPD4=", nil, "[]{}|\\<>"},
			{"special_quote", "Ig==", nil, "\""},

			// Binary/control characters
			{"binary_null", "AA==", nil, "\x00"},
			{"binary_max", "/w==", nil, "\xff"},
			{"control_chars", "Cg0J", nil, "\n\r\t"},
			{"binary_sequence", "AAECAw==", nil, "\x00\x01\x02\x03"},

			// UTF-8 characters
			{"utf8_accented", "aMOpbGxv", nil, "héllo"},
			{"utf8_chinese", "5L2g5aW9", nil, "你好"},
			{"utf8_cafe", "Y2Fmw6k=", nil, "café"},
			{"utf8_emoji", "8J+Mnw==", nil, "🌟"},

			// Edge cases
			{"empty_string", "", nil, ""},
			{"single_char", "YQ==", nil, "a"},
			{"two_chars", "YWI=", nil, "ab"},
			{"three_chars", "YWJj", nil, "abc"},
			{"four_chars", "YWJjZA==", nil, "abcd"},

			// Malformed data - non-strict mode (attempts to decode)
			{"malformed_exclamations_non_strict", "!!!", boolPtr(false), ""},
			{"malformed_missing_padding_non_strict", "YWJ", boolPtr(false), "ab"},
			{"malformed_wrong_padding_non_strict", "YWJj=", boolPtr(false), "abc"},

			// Strict mode tests
			{"valid_strict_mode", "YWJj", boolPtr(true), "abc"},
			{"malformed_exclamations_strict", "!!!", boolPtr(true), false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				funcs := GetStringFunctions()
				var base64DecodeFunc *registry.Function
				for _, f := range funcs {
					if f.Name == "base64_decode" {
						base64DecodeFunc = f
						break
					}
				}

				if base64DecodeFunc == nil {
					t.Fatal("base64_decode function not found")
				}

				var args []*values.Value
				args = append(args, values.NewString(tt.input))
				if tt.strict != nil {
					args = append(args, values.NewBool(*tt.strict))
				}

				result, err := base64DecodeFunc.Builtin(nil, args)

				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				// Handle expected false return for strict mode failures
				if expectedBool, ok := tt.expected.(bool); ok && !expectedBool {
					if result.Type != values.TypeBool || result.Data.(bool) != false {
						t.Errorf("Expected false, got %v", result)
					}
				} else {
					expectedStr := tt.expected.(string)
					if result.Type != values.TypeString || result.Data.(string) != expectedStr {
						t.Errorf("Expected %q, got %v", expectedStr, result)
					}
				}
			})
		}
	})

	// Test addslashes function
	t.Run("addslashes", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			expected string
		}{
			// Basic tests
			{"basic_hello_world", "hello world", "hello world"},
			{"single_quote", "O'Reilly", `O\'Reilly`},
			{"double_quote", `He said "Hello"`, `He said \"Hello\"`},
			{"backslash", `Back\slash`, `Back\\slash`},

			// Characters that need escaping
			{"single_quote_only", "'", `\'`},
			{"double_quote_only", `"`, `\"`},
			{"backslash_only", `\`, `\\`},
			{"null_byte", "\x00", `\0`},

			// Multiple escape characters
			{"mixed_escape_chars", "a'b\"c\\d\x00e", `a\'b\"c\\d\0e`},

			// Edge cases
			{"empty_string", "", ""},
			{"normal_text", "normal text", "normal text"},
			{"numbers", "123456", "123456"},
			{"special_symbols", "!@#$%^&*()", "!@#$%^&*()"},

			// UTF-8 characters (should not be escaped)
			{"utf8_accented", "héllo", "héllo"},
			{"utf8_chinese", "你好", "你好"},
			{"utf8_cafe", "café", "café"},
			{"utf8_emoji", "🌟", "🌟"},

			// Mixed content
			{"sql_query", `SELECT * FROM users WHERE name = 'John' AND city = "NYC"`, `SELECT * FROM users WHERE name = \'John\' AND city = \"NYC\"`},
			{"windows_path", `Path: C:\\Windows\System32\`, `Path: C:\\\\Windows\\System32\\`},

			// Control characters (most should NOT be escaped)
			{"newline", "\n", "\n"},
			{"carriage_return", "\r", "\r"},
			{"tab", "\t", "\t"},
			{"null_byte_explicit", "\x00", `\0`},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				funcs := GetStringFunctions()
				var addslashesFunc *registry.Function
				for _, f := range funcs {
					if f.Name == "addslashes" {
						addslashesFunc = f
						break
					}
				}

				if addslashesFunc == nil {
					t.Fatal("addslashes function not found")
				}

				result, err := addslashesFunc.Builtin(nil, []*values.Value{values.NewString(tt.input)})

				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				if result.Type != values.TypeString || result.Data.(string) != tt.expected {
					t.Errorf("Expected %q, got %v", tt.expected, result)
				}
			})
		}
	})

	// Test stripslashes function
	t.Run("stripslashes", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			expected string
		}{
			// Basic tests (reverse of addslashes)
			{"basic_hello_world", "hello world", "hello world"},
			{"single_quote", `O\'Reilly`, "O'Reilly"},
			{"double_quote", `He said \"Hello\"`, `He said "Hello"`},
			{"backslash", `Back\\slash`, `Back\slash`},

			// Characters that need unescaping
			{"single_quote_only", `\'`, "'"},
			{"double_quote_only", `\"`, `"`},
			{"backslash_only", `\\`, `\`},
			{"null_byte", `\0`, "\x00"},

			// Multiple escape characters
			{"mixed_escape_chars", `a\'b\"c\\d\0e`, "a'b\"c\\d\x00e"},

			// Edge cases
			{"empty_string", "", ""},
			{"normal_text", "normal text", "normal text"},
			{"numbers", "123456", "123456"},
			{"special_symbols", "!@#$%^&*()", "!@#$%^&*()"},

			// UTF-8 characters (should not be affected)
			{"utf8_accented", "héllo", "héllo"},
			{"utf8_chinese", "你好", "你好"},
			{"utf8_cafe", "café", "café"},
			{"utf8_emoji", "🌟", "🌟"},

			// Orphaned backslashes (backslash followed by non-special char)
			{"orphaned_backslash_a", `\a`, "a"},
			{"orphaned_backslash_z", `\z`, "z"},
			{"trailing_backslash", `trailing\`, "trailing"},

			// Control characters (only \0 is special, others are orphaned)
			{"newline_escape", `\n`, "n"},
			{"carriage_return_escape", `\r`, "r"},
			{"tab_escape", `\t`, "t"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				funcs := GetStringFunctions()
				var stripslashesFunc *registry.Function
				for _, f := range funcs {
					if f.Name == "stripslashes" {
						stripslashesFunc = f
						break
					}
				}

				if stripslashesFunc == nil {
					t.Fatal("stripslashes function not found")
				}

				result, err := stripslashesFunc.Builtin(nil, []*values.Value{values.NewString(tt.input)})

				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				if result.Type != values.TypeString || result.Data.(string) != tt.expected {
					t.Errorf("Expected %q, got %v", tt.expected, result)
				}
			})
		}
	})

	// Test md5 function
	t.Run("md5", func(t *testing.T) {
		tests := []struct {
			name         string
			input        string
			binary       *bool // nil means don't provide binary parameter
			expected     string
			expectedLen  int
		}{
			// Basic tests
			{"basic_hello_world", "hello world", nil, "5eb63bbbe01eeed093cb22bb8f5acdc3", 32},
			{"basic_Hello_World", "Hello World", nil, "b10a8db164e0754105b7a99be72e3fe5", 32},
			{"basic_numbers", "123456", nil, "e10adc3949ba59abbe56e057f20f883e", 32},
			{"basic_abc", "abc", nil, "900150983cd24fb0d6963f7d28e17f72", 32},

			// Edge cases
			{"empty_string", "", nil, "d41d8cd98f00b204e9800998ecf8427e", 32},
			{"single_char_a", "a", nil, "0cc175b9c0f1b6a831c399e269772661", 32},
			{"single_char_0", "0", nil, "cfcd208495d565ef66e7dff9f98764da", 32},
			{"single_space", " ", nil, "7215ee9c7d9dc229d2921a40e899ec5f", 32},

			// Special characters
			{"special_symbols", "!@#$%^&*()", nil, "05b28d17a7b6e7024b6e5d8cc43a8bf7", 32},
			{"quotes", "\"'", nil, "c61c1aca91758d029b272e56d6c3bb98", 32},
			{"control_chars", "\n\r\t", nil, "34c34c548ec80a813d48a51fc236dc52", 32},

			// UTF-8 characters
			{"utf8_accented", "héllo", nil, "be50e8478cf24ff3595bc7307fb91b50", 32},
			{"utf8_chinese", "你好", nil, "7eca689f0d3389d9dea66ae112e5cfd7", 32},
			{"utf8_cafe", "café", nil, "07117fe4a1ebd544965dc19573183da2", 32},
			{"utf8_emoji", "🌟", nil, "3714c7e811a90743e2121a4d82f796d6", 32},

			// Binary data
			{"binary_null", "\x00", nil, "93b885adfe0da089cdf634904fd59f71", 32},
			{"binary_max", "\xff", nil, "00594fd4f42ba43fc1ca0427a0576295", 32},
			{"binary_sequence", "\x00\x01\x02\x03", nil, "37b59afd592725f9305e484a5d7f5168", 32},

			// Longer strings
			{"long_sentence", "The quick brown fox jumps over the lazy dog", nil, "9e107d9d372bb6826bd81d3542a419d6", 32},

			// Binary output tests
			{"hello_hex_explicit", "hello", boolPtr(false), "5d41402abc4b2a76b9719d911017c592", 32},
			{"hello_binary", "hello", boolPtr(true), "", 16}, // Binary output, check length only
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				funcs := GetStringFunctions()
				var md5Func *registry.Function
				for _, f := range funcs {
					if f.Name == "md5" {
						md5Func = f
						break
					}
				}

				if md5Func == nil {
					t.Fatal("md5 function not found")
				}

				var args []*values.Value
				args = append(args, values.NewString(tt.input))
				if tt.binary != nil {
					args = append(args, values.NewBool(*tt.binary))
				}

				result, err := md5Func.Builtin(nil, args)

				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				if result.Type != values.TypeString {
					t.Errorf("Expected string type, got %v", result.Type)
				}

				resultStr := result.Data.(string)

				// Check length
				if len(resultStr) != tt.expectedLen {
					t.Errorf("Expected length %d, got %d", tt.expectedLen, len(resultStr))
				}

				// For binary output, we only check length since binary data is hard to compare in test
				if tt.binary != nil && *tt.binary {
					// Binary output - just verify length is 16 bytes
					if len(resultStr) != 16 {
						t.Errorf("Expected binary output length 16, got %d", len(resultStr))
					}
				} else {
					// Hex output - compare exact value
					if resultStr != tt.expected {
						t.Errorf("Expected %q, got %q", tt.expected, resultStr)
					}
				}
			})
		}
	})

	// Test sha1 function
	t.Run("sha1", func(t *testing.T) {
		tests := []struct {
			name         string
			input        string
			binary       *bool // nil means don't provide binary parameter
			expected     string
			expectedLen  int
		}{
			// Basic tests
			{"basic_hello_world", "hello world", nil, "2aae6c35c94fcfb415dbe95f408b9ce91ee846ed", 40},
			{"basic_Hello_World", "Hello World", nil, "0a4d55a8d778e5022fab701977c5d840bbc486d0", 40},
			{"basic_numbers", "123456", nil, "7c4a8d09ca3762af61e59520943dc26494f8941b", 40},
			{"basic_abc", "abc", nil, "a9993e364706816aba3e25717850c26c9cd0d89d", 40},

			// Edge cases
			{"empty_string", "", nil, "da39a3ee5e6b4b0d3255bfef95601890afd80709", 40},
			{"single_char_a", "a", nil, "86f7e437faa5a7fce15d1ddcb9eaeaea377667b8", 40},
			{"single_char_0", "0", nil, "b6589fc6ab0dc82cf12099d1c2d40ab994e8410c", 40},
			{"single_space", " ", nil, "b858cb282617fb0956d960215c8e84d1ccf909c6", 40},

			// Special characters
			{"special_symbols", "!@#$%^&*()", nil, "bf24d65c9bb05b9b814a966940bcfa50767c8a8d", 40},
			{"quotes", "\"'", nil, "b5989a085ef2d1a556b947160e395c470c3d5c55", 40},
			{"control_chars", "\n\r\t", nil, "5afecb81bd8cd5fd01c6424742920f10c66cde33", 40},

			// UTF-8 characters
			{"utf8_accented", "héllo", nil, "35b5ea45c5e41f78b46a937cc74d41dfea920890", 40},
			{"utf8_chinese", "你好", nil, "440ee0853ad1e99f962b63e459ef992d7c211722", 40},
			{"utf8_cafe", "café", nil, "f424452a9673918c6f09b0cdd35b20be8e6ae7d7", 40},
			{"utf8_emoji", "🌟", nil, "72f4543105a9b5d7fafc6e50037874b1a5209afa", 40},

			// Binary data
			{"binary_null", "\x00", nil, "5ba93c9db0cff93f52b521d7420e43f6eda2784f", 40},
			{"binary_max", "\xff", nil, "85e53271e14006f0265921d02d4d736cdc580b0b", 40},
			{"binary_sequence", "\x00\x01\x02\x03", nil, "a02a05b025b928c039cf1ae7e8ee04e7c190c0db", 40},

			// Longer strings
			{"long_sentence", "The quick brown fox jumps over the lazy dog", nil, "2fd4e1c67a2d28fced849ee1bb76e7391b93eb12", 40},

			// Binary output tests
			{"hello_hex_explicit", "hello", boolPtr(false), "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d", 40},
			{"hello_binary", "hello", boolPtr(true), "", 20}, // Binary output, check length only
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				funcs := GetStringFunctions()
				var sha1Func *registry.Function
				for _, f := range funcs {
					if f.Name == "sha1" {
						sha1Func = f
						break
					}
				}

				if sha1Func == nil {
					t.Fatal("sha1 function not found")
				}

				var args []*values.Value
				args = append(args, values.NewString(tt.input))
				if tt.binary != nil {
					args = append(args, values.NewBool(*tt.binary))
				}

				result, err := sha1Func.Builtin(nil, args)

				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				if result.Type != values.TypeString {
					t.Errorf("Expected string type, got %v", result.Type)
				}

				resultStr := result.Data.(string)

				// Check length
				if len(resultStr) != tt.expectedLen {
					t.Errorf("Expected length %d, got %d", tt.expectedLen, len(resultStr))
				}

				// For binary output, we only check length since binary data is hard to compare in test
				if tt.binary != nil && *tt.binary {
					// Binary output - just verify length is 20 bytes
					if len(resultStr) != 20 {
						t.Errorf("Expected binary output length 20, got %d", len(resultStr))
					}
				} else {
					// Hex output - compare exact value
					if resultStr != tt.expected {
						t.Errorf("Expected %q, got %q", tt.expected, resultStr)
					}
				}
			})
		}
	})

	t.Run("number_format", func(t *testing.T) {
		tests := []struct {
			name            string
			number          interface{}
			decimals        *int
			decimalPoint    *string
			thousandsSep    *string
			expected        string
		}{
			// Basic functionality
			{"basic float", 1234.5, nil, nil, nil, "1,235"},
			{"large number", 1234567.891, nil, nil, nil, "1,234,568"},
			{"two decimals", 1234.567, intPtr(2), nil, nil, "1,234.57"},
			{"three decimals", 1234567.891, intPtr(3), nil, nil, "1,234,567.891"},

			// Custom separators
			{"custom decimal comma", 1234567.891, intPtr(2), strPtr(","), strPtr("."), "1.234.567,89"},
			{"custom space separator", 1234567.891, intPtr(2), strPtr(","), strPtr(" "), "1 234 567,89"},
			{"no thousands separator", 1234567.891, intPtr(2), strPtr("."), strPtr(""), "1234567.89"},
			{"custom symbols", 12345.67, intPtr(2), strPtr("@"), strPtr("#"), "12#345@67"},

			// Edge cases
			{"zero", 0.0, nil, nil, nil, "0"},
			{"zero with decimals", 0.0, intPtr(2), nil, nil, "0.00"},
			{"negative number", -1234.567, nil, nil, nil, "-1,235"},
			{"negative with decimals", -1234.567, intPtr(2), nil, nil, "-1,234.57"},

			// Small and large numbers
			{"small number no decimals", 0.001, nil, nil, nil, "0"},
			{"small number with decimals", 0.001, intPtr(3), nil, nil, "0.001"},
			{"rounding up", 0.999, nil, nil, nil, "1"},
			{"large number", 999999999.0, nil, nil, nil, "999,999,999"},

			// Precision and rounding
			{"round down", 1.1, intPtr(0), nil, nil, "1"},
			{"round up half", 1.5, intPtr(0), nil, nil, "2"},
			{"round up", 1.9, intPtr(0), nil, nil, "2"},
			{"round up half even", 2.5, intPtr(0), nil, nil, "3"},
			{"pad with zeros", 1.234, intPtr(5), nil, nil, "1.23400"},

			// String inputs
			{"string number", "1234.567", nil, nil, nil, "1,235"},
			{"string with decimals", "1234.567", intPtr(2), nil, nil, "1,234.57"},
			{"padded string", "0001234.567000", intPtr(2), nil, nil, "1,234.57"},

			// Parameter variations
			{"one param", 1234.5, nil, nil, nil, "1,235"},
			{"two params", 1234.5, intPtr(1), nil, nil, "1,234.5"},
		}

		funcs := GetStringFunctions()
		var numberFormatFunc *registry.Function
		for _, f := range funcs {
			if f.Name == "number_format" {
				numberFormatFunc = f
				break
			}
		}

		if numberFormatFunc == nil {
			t.Fatal("number_format function not found")
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var args []*values.Value

				// First argument (number)
				switch v := tt.number.(type) {
				case string:
					args = append(args, values.NewString(v))
				case float64:
					args = append(args, values.NewFloat(v))
				case int:
					args = append(args, values.NewInt(int64(v)))
				}

				// Optional arguments
				if tt.decimals != nil {
					args = append(args, values.NewInt(int64(*tt.decimals)))
				}
				if tt.decimalPoint != nil {
					args = append(args, values.NewString(*tt.decimalPoint))
				}
				if tt.thousandsSep != nil {
					args = append(args, values.NewString(*tt.thousandsSep))
				}

				result, err := numberFormatFunc.Builtin(nil, args)

				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				if result.Type != values.TypeString {
					t.Errorf("Expected string type, got %v", result.Type)
				}

				resultStr := result.Data.(string)
				if resultStr != tt.expected {
					t.Errorf("Expected %q, got %q", tt.expected, resultStr)
				}
			})
		}
	})

	t.Run("htmlentities", func(t *testing.T) {
		tests := []struct {
			name        string
			input       string
			flags       *int
			encoding    *string
			doubleEncode *bool
			expected    string
		}{
			// Basic functionality
			{"basic string", "hello world", nil, nil, nil, "hello world"},
			{"XSS script", "<script>alert(\"XSS\")</script>", nil, nil, nil, "&lt;script&gt;alert(&quot;XSS&quot;)&lt;/script&gt;"},
			{"accented chars", "café & résumé", nil, nil, nil, "caf&eacute; &amp; r&eacute;sum&eacute;"},
			{"copyright symbol", "© 2023 & Company", nil, nil, nil, "&copy; 2023 &amp; Company"},

			// Quote style tests
			{"quotes ENT_COMPAT", "hello \"world\" 'test'", intPtr(2), nil, nil, "hello &quot;world&quot; 'test'"},
			{"quotes ENT_QUOTES", "hello \"world\" 'test'", intPtr(3), nil, nil, "hello &quot;world&quot; &#039;test&#039;"},
			{"quotes ENT_NOQUOTES", "hello \"world\" 'test'", intPtr(0), nil, nil, "hello \"world\" 'test'"},

			// Double encode tests
			{"double encode true", "&lt;test&gt;", intPtr(2), strPtr("UTF-8"), boolPtr(true), "&amp;lt;test&amp;gt;"},
			{"double encode false", "&lt;test&gt;", intPtr(2), strPtr("UTF-8"), boolPtr(false), "&lt;test&gt;"},

			// Special characters
			{"basic special chars", "<>&\"'", nil, nil, nil, "&lt;&gt;&amp;&quot;&#039;"},
			{"Latin-1 supplement", "¡¢£¤¥¦§¨©ª«¬­®¯", nil, nil, nil, "&iexcl;&cent;&pound;&curren;&yen;&brvbar;&sect;&uml;&copy;&ordf;&laquo;&not;&shy;&reg;&macr;"},
			{"uppercase accented", "ÀÁÂÃÄÅÆÇÈÉÊË", nil, nil, nil, "&Agrave;&Aacute;&Acirc;&Atilde;&Auml;&Aring;&AElig;&Ccedil;&Egrave;&Eacute;&Ecirc;&Euml;"},
			{"lowercase accented", "àáâãäåæçèéêë", nil, nil, nil, "&agrave;&aacute;&acirc;&atilde;&auml;&aring;&aelig;&ccedil;&egrave;&eacute;&ecirc;&euml;"},

			// Mathematical symbols
			{"math symbols", "±×÷≤≥≠≈∞", nil, nil, nil, "&plusmn;&times;&divide;&le;&ge;&ne;&asymp;&infin;"},
			{"advanced math", "∀∂∃∇∈∉∋∏∑", nil, nil, nil, "&forall;&part;&exist;&nabla;&isin;&notin;&ni;&prod;&sum;"},

			// Greek letters
			{"greek lowercase", "αβγδεζηθικλμ", nil, nil, nil, "&alpha;&beta;&gamma;&delta;&epsilon;&zeta;&eta;&theta;&iota;&kappa;&lambda;&mu;"},
			{"greek uppercase", "ΑΒΓΔΕΖΗΘΙΚΛΜ", nil, nil, nil, "&Alpha;&Beta;&Gamma;&Delta;&Epsilon;&Zeta;&Eta;&Theta;&Iota;&Kappa;&Lambda;&Mu;"},

			// Edge cases
			{"empty string", "", nil, nil, nil, ""},
			{"single char", "a", nil, nil, nil, "a"},
			{"space", " ", nil, nil, nil, " "},
			{"numbers", "123", nil, nil, nil, "123"},

			// Various symbols
			{"punctuation symbols", "†‡•…‰′″‹›€™", nil, nil, nil, "&dagger;&Dagger;&bull;&hellip;&permil;&prime;&Prime;&lsaquo;&rsaquo;&euro;&trade;"},
			{"card suits", "♠♣♥♦", nil, nil, nil, "&spades;&clubs;&hearts;&diams;"},

			// Mixed content
			{"mixed special", "<>&\"'àáâã", nil, nil, nil, "&lt;&gt;&amp;&quot;&#039;&agrave;&aacute;&acirc;&atilde;"},
		}

		funcs := GetStringFunctions()
		var htmlentitiesFunc *registry.Function
		for _, f := range funcs {
			if f.Name == "htmlentities" {
				htmlentitiesFunc = f
				break
			}
		}

		if htmlentitiesFunc == nil {
			t.Fatal("htmlentities function not found")
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var args []*values.Value
				args = append(args, values.NewString(tt.input))

				if tt.flags != nil {
					args = append(args, values.NewInt(int64(*tt.flags)))
				}
				if tt.encoding != nil {
					args = append(args, values.NewString(*tt.encoding))
				}
				if tt.doubleEncode != nil {
					args = append(args, values.NewBool(*tt.doubleEncode))
				}

				result, err := htmlentitiesFunc.Builtin(nil, args)

				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				if result.Type != values.TypeString {
					t.Errorf("Expected string type, got %v", result.Type)
				}

				resultStr := result.Data.(string)
				if resultStr != tt.expected {
					t.Errorf("Expected %q, got %q", tt.expected, resultStr)
				}
			})
		}
	})

	t.Run("nl2br", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			isXHTML  *bool
			expected string
		}{
			// Basic functionality
			{"no newlines", "hello world", nil, "hello world"},
			{"unix newline", "hello\nworld", nil, "hello<br />\nworld"},
			{"windows newline", "hello\r\nworld", nil, "hello<br />\r\nworld"},
			{"mac newline", "hello\rworld", nil, "hello<br />\rworld"},

			// XHTML parameter tests
			{"unix HTML mode", "hello\nworld", boolPtr(false), "hello<br>\nworld"},
			{"unix XHTML mode", "hello\nworld", boolPtr(true), "hello<br />\nworld"},
			{"windows HTML mode", "hello\r\nworld", boolPtr(false), "hello<br>\r\nworld"},
			{"windows XHTML mode", "hello\r\nworld", boolPtr(true), "hello<br />\r\nworld"},
			{"mac HTML mode", "hello\rworld", boolPtr(false), "hello<br>\rworld"},
			{"mac XHTML mode", "hello\rworld", boolPtr(true), "hello<br />\rworld"},

			// Edge cases
			{"empty string", "", nil, ""},
			{"single unix newline", "\n", nil, "<br />\n"},
			{"single windows newline", "\r\n", nil, "<br />\r\n"},
			{"single mac newline", "\r", nil, "<br />\r"},
			{"single unix newline HTML", "\n", boolPtr(false), "<br>\n"},
			{"single unix newline XHTML", "\n", boolPtr(true), "<br />\n"},

			// Multiple newlines
			{"multiple unix", "line1\nline2\nline3", nil, "line1<br />\nline2<br />\nline3"},
			{"multiple windows", "line1\r\nline2\r\nline3", nil, "line1<br />\r\nline2<br />\r\nline3"},
			{"consecutive unix", "line1\n\nline3", nil, "line1<br />\n<br />\nline3"},
			{"consecutive windows", "line1\r\n\r\nline3", nil, "line1<br />\r\n<br />\r\nline3"},
			{"triple consecutive", "line1\n\n\nline2", nil, "line1<br />\n<br />\n<br />\nline2"},

			// Multiple newlines with XHTML modes
			{"multiple unix HTML", "line1\nline2\nline3", boolPtr(false), "line1<br>\nline2<br>\nline3"},
			{"multiple unix XHTML", "line1\nline2\nline3", boolPtr(true), "line1<br />\nline2<br />\nline3"},
			{"consecutive HTML", "line1\n\nline3", boolPtr(false), "line1<br>\n<br>\nline3"},
			{"consecutive XHTML", "line1\n\nline3", boolPtr(true), "line1<br />\n<br />\nline3"},

			// Mixed newline types
			{"mixed newlines", "unix\nmac\rwindows\r\n", nil, "unix<br />\nmac<br />\rwindows<br />\r\n"},
			{"mixed newlines HTML", "unix\nmac\rwindows\r\n", boolPtr(false), "unix<br>\nmac<br>\rwindows<br>\r\n"},

			// Newlines at boundaries
			{"newline at start", "\nhello", nil, "<br />\nhello"},
			{"newline at end", "hello\n", nil, "hello<br />\n"},
			{"newlines at both ends", "\nhello\n", nil, "<br />\nhello<br />\n"},
			{"windows at boundaries", "\r\nhello\r\n", nil, "<br />\r\nhello<br />\r\n"},

			// Special characters (nl2br doesn't escape HTML)
			{"with ampersand", "hello & world\ntest", nil, "hello & world<br />\ntest"},
			{"with HTML tags", "<script>\nalert(\"test\")\n</script>", nil, "<script><br />\nalert(\"test\")<br />\n</script>"},

			// Real-world scenarios
			{"paragraph text", "First paragraph.\n\nSecond paragraph.", nil, "First paragraph.<br />\n<br />\nSecond paragraph."},
			{"list format", "Item 1\nItem 2\nItem 3", nil, "Item 1<br />\nItem 2<br />\nItem 3"},
		}

		funcs := GetStringFunctions()
		var nl2brFunc *registry.Function
		for _, f := range funcs {
			if f.Name == "nl2br" {
				nl2brFunc = f
				break
			}
		}

		if nl2brFunc == nil {
			t.Fatal("nl2br function not found")
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var args []*values.Value
				args = append(args, values.NewString(tt.input))

				if tt.isXHTML != nil {
					args = append(args, values.NewBool(*tt.isXHTML))
				}

				result, err := nl2brFunc.Builtin(nil, args)

				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				if result.Type != values.TypeString {
					t.Errorf("Expected string type, got %v", result.Type)
				}

				resultStr := result.Data.(string)
				if resultStr != tt.expected {
					t.Errorf("Expected %q, got %q", tt.expected, resultStr)
				}
			})
		}
	})

	t.Run("str_rot13", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			expected string
		}{
			// Basic functionality
			{"basic lowercase", "hello world", "uryyb jbeyq"},
			{"basic uppercase", "HELLO WORLD", "URYYB JBEYQ"},
			{"mixed case", "Hello World", "Uryyb Jbeyq"},
			{"alphabet start", "abc", "nop"},
			{"alphabet end", "xyz", "klm"},
			{"uppercase start", "ABC", "NOP"},
			{"uppercase end", "XYZ", "KLM"},

			// Edge cases
			{"empty string", "", ""},
			{"single a", "a", "n"},
			{"single z", "z", "m"},
			{"single A", "A", "N"},
			{"single Z", "Z", "M"},
			{"middle lowercase", "n", "a"},
			{"middle uppercase", "N", "A"},

			// Numbers and special characters (unchanged)
			{"numbers only", "123", "123"},
			{"special characters", "!@#$%^&*()", "!@#$%^&*()"},
			{"mixed alphanumeric", "hello123world", "uryyb123jbeyq"},
			{"interspersed", "a1b2c3", "n1o2p3"},

			// Full alphabet tests
			{"lowercase alphabet", "abcdefghijklmnopqrstuvwxyz", "nopqrstuvwxyzabcdefghijklm"},
			{"uppercase alphabet", "ABCDEFGHIJKLMNOPQRSTUVWXYZ", "NOPQRSTUVWXYZABCDEFGHIJKLM"},

			// Real-world content
			{"sentence", "The quick brown fox jumps over the lazy dog!", "Gur dhvpx oebja sbk whzcf bire gur ynml qbt!"},
			{"description", "ROT13 is a simple letter substitution cipher.", "EBG13 vf n fvzcyr yrggre fhofgvghgvba pvcure."},

			// Whitespace and punctuation
			{"with spaces", "  spaces  ", "  fcnprf  "},
			{"with newlines", "line1\nline2\ttab", "yvar1\nyvar2\tgno"},
			{"with punctuation", "Hello, World!", "Uryyb, Jbeyq!"},

			// Unicode characters (should remain unchanged)
			{"unicode accented", "café", "pnsé"},
			{"unicode chinese", "你好", "你好"},
			{"unicode mixed", "naïve résumé", "anïir eéfhzé"},

			// Known ROT13 examples
			{"known hello", "hello", "uryyb"},
			{"known world", "world", "jbeyq"},
			{"known test", "test", "grfg"},
			{"known PHP", "PHP", "CUC"},

			// Reversibility examples (applying ROT13 twice)
			{"reversible simple", "uryyb", "hello"},
			{"reversible mixed", "Uryyb Jbeyq", "Hello World"},
			{"reversible sentence", "Gur dhvpx oebja sbk", "The quick brown fox"},

			// Boundary characters
			{"boundary m to z", "m", "z"},
			{"boundary n to a", "n", "a"},
			{"boundary M to Z", "M", "Z"},
			{"boundary N to A", "N", "A"},

			// Complex mixed content
			{"complex mixed", "Test123!@# ROT13", "Grfg123!@# EBG13"},
			{"email format", "test@example.com", "grfg@rknzcyr.pbz"},
			{"url format", "https://example.com/path", "uggcf://rknzcyr.pbz/cngu"},
		}

		funcs := GetStringFunctions()
		var strRot13Func *registry.Function
		for _, f := range funcs {
			if f.Name == "str_rot13" {
				strRot13Func = f
				break
			}
		}

		if strRot13Func == nil {
			t.Fatal("str_rot13 function not found")
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var args []*values.Value
				args = append(args, values.NewString(tt.input))

				result, err := strRot13Func.Builtin(nil, args)

				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				if result.Type != values.TypeString {
					t.Errorf("Expected string type, got %v", result.Type)
				}

				resultStr := result.Data.(string)
				if resultStr != tt.expected {
					t.Errorf("Expected %q, got %q", tt.expected, resultStr)
				}
			})
		}
	})

	// Test ROT13 reversibility property
	t.Run("str_rot13_reversibility", func(t *testing.T) {
		testCases := []string{
			"hello world",
			"HELLO WORLD",
			"Hello World",
			"The quick brown fox",
			"test123!@#",
			"",
			"a",
			"Z",
		}

		funcs := GetStringFunctions()
		var strRot13Func *registry.Function
		for _, f := range funcs {
			if f.Name == "str_rot13" {
				strRot13Func = f
				break
			}
		}

		if strRot13Func == nil {
			t.Fatal("str_rot13 function not found")
		}

		for _, original := range testCases {
			t.Run("reversible_"+original, func(t *testing.T) {
				// Apply ROT13 once
				args1 := []*values.Value{values.NewString(original)}
				result1, err := strRot13Func.Builtin(nil, args1)
				if err != nil {
					t.Fatalf("First ROT13 failed: %v", err)
				}
				encoded := result1.Data.(string)

				// Apply ROT13 again
				args2 := []*values.Value{values.NewString(encoded)}
				result2, err := strRot13Func.Builtin(nil, args2)
				if err != nil {
					t.Fatalf("Second ROT13 failed: %v", err)
				}
				decoded := result2.Data.(string)

				// Should be back to original
				if decoded != original {
					t.Errorf("ROT13 not reversible: original=%q, encoded=%q, decoded=%q", original, encoded, decoded)
				}
			})
		}
	})

	t.Run("wordwrap", func(t *testing.T) {
		tests := []struct {
			name     string
			text     string
			width    *int
			break_   *string
			cut      *bool
			expected string
		}{
			// Basic functionality
			{"basic no wrap", "hello world", nil, nil, nil, "hello world"},
			{"basic wrap at 5", "hello world", intPtr(5), nil, nil, "hello\nworld"},
			{"basic wrap at 10", "hello world", intPtr(10), nil, nil, "hello\nworld"},
			{"multiple wraps", "hello world test", intPtr(8), nil, nil, "hello\nworld\ntest"},

			// Custom line breaks
			{"HTML breaks", "hello world test", intPtr(8), strPtr("<br>"), nil, "hello<br>world<br>test"},
			{"pipe separator", "hello world test", intPtr(8), strPtr("|"), nil, "hello|world|test"},
			{"windows breaks", "hello world test", intPtr(8), strPtr("\r\n"), nil, "hello\r\nworld\r\ntest"},

			// Cut parameter tests
			{"long word no cut", "supercalifragilisticexpialidocious", intPtr(10), nil, nil, "supercalifragilisticexpialidocious"},
			{"long word with cut", "supercalifragilisticexpialidocious", intPtr(10), nil, boolPtr(true), "supercalif\nragilistic\nexpialidoc\nious"},
			{"mixed with cut", "hello supercalifragilisticexpialidocious world", intPtr(10), nil, boolPtr(true), "hello\nsupercalif\nragilistic\nexpialidoc\nious world"},

			// Edge cases
			{"empty string", "", nil, nil, nil, ""},
			{"single character", "a", nil, nil, nil, "a"},
			{"only spaces", "   ", nil, nil, nil, "   "},
			{"width larger than text", "hello", intPtr(100), nil, nil, "hello"},
			{"width smaller than word", "hello", intPtr(1), nil, nil, "hello"},

			// Existing line breaks
			{"text with newlines", "hello\nworld", intPtr(10), nil, nil, "hello\nworld"},
			{"multiple newlines", "hello\nworld\ntest", intPtr(5), nil, nil, "hello\nworld\ntest"},

			// Whitespace handling
			{"multiple spaces", "hello  world", intPtr(8), nil, nil, "hello \nworld"},
			{"various spaces", "hello   world   test", intPtr(8), nil, nil, "hello  \nworld  \ntest"},
			{"leading trailing spaces", " hello world ", intPtr(8), nil, nil, " hello\nworld "},

			// Punctuation
			{"with punctuation", "hello, world! How are you?", intPtr(10), nil, nil, "hello,\nworld! How\nare you?"},
			{"comma separated", "one,two,three,four,five", intPtr(10), nil, nil, "one,two,three,four,five"},

			// Different widths
			{"width 20", "The quick brown fox jumps over the lazy dog", intPtr(20), nil, nil, "The quick brown fox\njumps over the lazy\ndog"},
			{"width 15", "The quick brown fox jumps over the lazy dog", intPtr(15), nil, nil, "The quick brown\nfox jumps over\nthe lazy dog"},
			{"width 10", "The quick brown fox jumps over the lazy dog", intPtr(10), nil, nil, "The quick\nbrown fox\njumps over\nthe lazy\ndog"},
			{"width 5", "The quick brown fox jumps over the lazy dog", intPtr(5), nil, nil, "The\nquick\nbrown\nfox\njumps\nover\nthe\nlazy\ndog"},

			// Special widths
			{"width 0", "hello world", intPtr(0), nil, nil, "hello\nworld"},
			{"width negative", "hello world", intPtr(-1), nil, nil, "hello\nworld"},

			// Special characters
			{"with tab", "hello\tworld", intPtr(10), nil, nil, "hello\tworld"},
			{"with ampersand", "hello & world", intPtr(8), nil, nil, "hello &\nworld"},

			// Long lines with cut
			{"alphabet with cut", "abcdefghijklmnopqrstuvwxyz", intPtr(10), nil, boolPtr(true), "abcdefghij\nklmnopqrst\nuvwxyz"},
			{"mixed with cut 5", "1234567890abcdefghijk", intPtr(5), nil, boolPtr(true), "12345\n67890\nabcde\nfghij\nk"},

			// Real-world scenarios
			{"email line", "This is a long email line that should be wrapped at 72 characters for proper email formatting.", intPtr(72), nil, nil, "This is a long email line that should be wrapped at 72 characters for\nproper email formatting."},
			{"code comment", "// This is a very long code comment that exceeds the usual line length limit", intPtr(50), nil, nil, "// This is a very long code comment that exceeds\nthe usual line length limit"},

			// Cut mode edge cases
			{"cut single char", "a", intPtr(5), nil, boolPtr(true), "a"},
			{"cut exact width", "hello", intPtr(5), nil, boolPtr(true), "hello"},
			{"cut one over", "hello!", intPtr(5), nil, boolPtr(true), "hello\n!"},

			// Multiple parameters combination
			{"all params", "hello world test example", intPtr(8), strPtr(" | "), boolPtr(false), "hello | world | test | example"},
		}

		funcs := GetStringFunctions()
		var wordwrapFunc *registry.Function
		for _, f := range funcs {
			if f.Name == "wordwrap" {
				wordwrapFunc = f
				break
			}
		}

		if wordwrapFunc == nil {
			t.Fatal("wordwrap function not found")
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var args []*values.Value
				args = append(args, values.NewString(tt.text))

				if tt.width != nil {
					args = append(args, values.NewInt(int64(*tt.width)))

					if tt.break_ != nil {
						args = append(args, values.NewString(*tt.break_))

						if tt.cut != nil {
							args = append(args, values.NewBool(*tt.cut))
						}
					} else if tt.cut != nil {
						// If break is nil but cut is not, we need to add default break
						args = append(args, values.NewString("\n"))
						args = append(args, values.NewBool(*tt.cut))
					}
				} else if tt.break_ != nil || tt.cut != nil {
					// Need to add default width
					args = append(args, values.NewInt(75))

					if tt.break_ != nil {
						args = append(args, values.NewString(*tt.break_))

						if tt.cut != nil {
							args = append(args, values.NewBool(*tt.cut))
						}
					} else if tt.cut != nil {
						args = append(args, values.NewString("\n"))
						args = append(args, values.NewBool(*tt.cut))
					}
				}

				result, err := wordwrapFunc.Builtin(nil, args)

				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				if result.Type != values.TypeString {
					t.Errorf("Expected string type, got %v", result.Type)
				}

				resultStr := result.Data.(string)
				if resultStr != tt.expected {
					t.Errorf("Expected %q, got %q", tt.expected, resultStr)
				}
			})
		}
	})
}

// Helper functions for test pointers
func intPtr(i int) *int {
	return &i
}

func strPtr(s string) *string {
	return &s
}

// Helper function to create bool pointer for optional parameters
func boolPtr(b bool) *bool {
	return &b
}