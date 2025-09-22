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

	t.Run("levenshtein", func(t *testing.T) {
		// Find the levenshtein function
		var levenshteinFunc *registry.Function
		for _, f := range functions {
			if f.Name == "levenshtein" {
				levenshteinFunc = f
				break
			}
		}

		if levenshteinFunc == nil {
			t.Fatal("levenshtein function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected int64
		}{
			// Basic functionality
			{
				name: "identical strings",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("hello"),
				},
				expected: 0,
			},
			{
				name: "completely different",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("world"),
				},
				expected: 4,
			},
			{
				name: "one character different",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("hallo"),
				},
				expected: 1,
			},
			{
				name: "partial match",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewString("help"),
				},
				expected: 2,
			},
			{
				name: "classic example kitten/sitting",
				args: []*values.Value{
					values.NewString("kitten"),
					values.NewString("sitting"),
				},
				expected: 3,
			},

			// Edge cases
			{
				name: "both empty",
				args: []*values.Value{
					values.NewString(""),
					values.NewString(""),
				},
				expected: 0,
			},
			{
				name: "first empty",
				args: []*values.Value{
					values.NewString(""),
					values.NewString("hello"),
				},
				expected: 5,
			},
			{
				name: "second empty",
				args: []*values.Value{
					values.NewString("world"),
					values.NewString(""),
				},
				expected: 5,
			},
			{
				name: "single character identical",
				args: []*values.Value{
					values.NewString("a"),
					values.NewString("a"),
				},
				expected: 0,
			},
			{
				name: "single character different",
				args: []*values.Value{
					values.NewString("a"),
					values.NewString("b"),
				},
				expected: 1,
			},

			// Case sensitivity
			{
				name: "case different",
				args: []*values.Value{
					values.NewString("Hello"),
					values.NewString("hello"),
				},
				expected: 1,
			},
			{
				name: "all caps vs lowercase",
				args: []*values.Value{
					values.NewString("HELLO"),
					values.NewString("hello"),
				},
				expected: 5,
			},

			// Insertions and deletions
			{
				name: "insertion",
				args: []*values.Value{
					values.NewString("cat"),
					values.NewString("cats"),
				},
				expected: 1,
			},
			{
				name: "deletion",
				args: []*values.Value{
					values.NewString("cats"),
					values.NewString("cat"),
				},
				expected: 1,
			},

			// Same characters different order
			{
				name: "transposed",
				args: []*values.Value{
					values.NewString("ab"),
					values.NewString("ba"),
				},
				expected: 2,
			},

			// Known test vectors
			{
				name: "saturday/sunday",
				args: []*values.Value{
					values.NewString("saturday"),
					values.NewString("sunday"),
				},
				expected: 3,
			},
			{
				name: "abc/def no common",
				args: []*values.Value{
					values.NewString("abc"),
					values.NewString("def"),
				},
				expected: 3,
			},

			// Longer strings
			{
				name: "fox/dog word difference",
				args: []*values.Value{
					values.NewString("The quick brown fox"),
					values.NewString("The quick brown dog"),
				},
				expected: 2,
			},

			// Unicode handling
			{
				name: "accented characters",
				args: []*values.Value{
					values.NewString("café"),
					values.NewString("cafe"),
				},
				expected: 2, // Each accented char counts as 2 bytes
			},
			{
				name: "Chinese characters identical",
				args: []*values.Value{
					values.NewString("你好"),
					values.NewString("你好"),
				},
				expected: 0,
			},
			{
				name: "Chinese characters different",
				args: []*values.Value{
					values.NewString("你好"),
					values.NewString("再见"),
				},
				expected: 6, // Each Chinese char is 3 bytes
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := levenshteinFunc.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("levenshtein() error = %v", err)
				}
				if result.Type != values.TypeInt {
					t.Fatalf("levenshtein() returned %s, want int", result.Type)
				}
				if result.Data.(int64) != tt.expected {
					t.Errorf("levenshtein() = %d, want %d", result.Data.(int64), tt.expected)
				}
			})
		}
	})

	t.Run("hash", func(t *testing.T) {
		// Find the hash function
		var hashFunc *registry.Function
		for _, f := range functions {
			if f.Name == "hash" {
				hashFunc = f
				break
			}
		}

		if hashFunc == nil {
			t.Fatal("hash function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected string
		}{
			// MD5 tests
			{
				name: "md5 hello",
				args: []*values.Value{
					values.NewString("md5"),
					values.NewString("hello"),
				},
				expected: "5d41402abc4b2a76b9719d911017c592",
			},
			{
				name: "md5 empty string",
				args: []*values.Value{
					values.NewString("md5"),
					values.NewString(""),
				},
				expected: "d41d8cd98f00b204e9800998ecf8427e",
			},
			{
				name: "md5 single char",
				args: []*values.Value{
					values.NewString("md5"),
					values.NewString("a"),
				},
				expected: "0cc175b9c0f1b6a831c399e269772661",
			},

			// SHA1 tests
			{
				name: "sha1 hello",
				args: []*values.Value{
					values.NewString("sha1"),
					values.NewString("hello"),
				},
				expected: "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d",
			},
			{
				name: "sha1 empty string",
				args: []*values.Value{
					values.NewString("sha1"),
					values.NewString(""),
				},
				expected: "da39a3ee5e6b4b0d3255bfef95601890afd80709",
			},

			// SHA256 tests
			{
				name: "sha256 hello",
				args: []*values.Value{
					values.NewString("sha256"),
					values.NewString("hello"),
				},
				expected: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
			},
			{
				name: "sha256 empty string",
				args: []*values.Value{
					values.NewString("sha256"),
					values.NewString(""),
				},
				expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			},

			// SHA512 tests
			{
				name: "sha512 hello",
				args: []*values.Value{
					values.NewString("sha512"),
					values.NewString("hello"),
				},
				expected: "9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043",
			},

			// Case sensitivity
			{
				name: "md5 case sensitive",
				args: []*values.Value{
					values.NewString("md5"),
					values.NewString("Hello"),
				},
				expected: "8b1a9953c4611296a827abf8c47804d7",
			},

			// Special characters
			{
				name: "md5 special chars",
				args: []*values.Value{
					values.NewString("md5"),
					values.NewString("!@#$%^&*()"),
				},
				expected: "05b28d17a7b6e7024b6e5d8cc43a8bf7",
			},

			// Unicode characters
			{
				name: "md5 unicode",
				args: []*values.Value{
					values.NewString("md5"),
					values.NewString("café"),
				},
				expected: "07117fe4a1ebd544965dc19573183da2",
			},

			// Longer strings
			{
				name: "md5 pangram",
				args: []*values.Value{
					values.NewString("md5"),
					values.NewString("The quick brown fox jumps over the lazy dog"),
				},
				expected: "9e107d9d372bb6826bd81d3542a419d6",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := hashFunc.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("hash() error = %v", err)
				}
				if result.Type != values.TypeString {
					t.Fatalf("hash() returned %s, want string", result.Type)
				}
				if result.Data.(string) != tt.expected {
					t.Errorf("hash() = %q, want %q", result.Data.(string), tt.expected)
				}
			})
		}

		// Test invalid algorithm
		t.Run("invalid algorithm", func(t *testing.T) {
			_, err := hashFunc.Builtin(nil, []*values.Value{
				values.NewString("invalid_algo"),
				values.NewString("test"),
			})
			if err == nil {
				t.Error("Expected error for invalid algorithm, got nil")
			}
		})
	})

	t.Run("money_format", func(t *testing.T) {
		// Find the money_format function
		var moneyFormatFunc *registry.Function
		for _, f := range functions {
			if f.Name == "money_format" {
				moneyFormatFunc = f
				break
			}
		}

		if moneyFormatFunc == nil {
			t.Fatal("money_format function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected string
		}{
			// Basic formatting
			{
				name: "basic national format",
				args: []*values.Value{
					values.NewString("%.2n"),
					values.NewFloat(1234.56),
				},
				expected: "$1,234.56",
			},
			{
				name: "basic international format",
				args: []*values.Value{
					values.NewString("%.2i"),
					values.NewFloat(1234.56),
				},
				expected: "USD 1,234.56",
			},
			{
				name: "no precision national",
				args: []*values.Value{
					values.NewString("%n"),
					values.NewFloat(1234.56),
				},
				expected: "$1,235",
			},
			{
				name: "no precision international",
				args: []*values.Value{
					values.NewString("%i"),
					values.NewFloat(1234.56),
				},
				expected: "USD 1,235",
			},

			// Different precision
			{
				name: "zero decimal places",
				args: []*values.Value{
					values.NewString("%.0n"),
					values.NewFloat(1234.56),
				},
				expected: "$1,235",
			},
			{
				name: "one decimal place",
				args: []*values.Value{
					values.NewString("%.1n"),
					values.NewFloat(1234.56),
				},
				expected: "$1,234.6",
			},
			{
				name: "three decimal places",
				args: []*values.Value{
					values.NewString("%.3n"),
					values.NewFloat(1234.56),
				},
				expected: "$1,234.560",
			},

			// Negative numbers
			{
				name: "negative amount",
				args: []*values.Value{
					values.NewString("%.2n"),
					values.NewFloat(-1234.56),
				},
				expected: "-$1,234.56",
			},

			// Edge cases
			{
				name: "zero amount",
				args: []*values.Value{
					values.NewString("%.2n"),
					values.NewFloat(0),
				},
				expected: "$0.00",
			},
			{
				name: "one cent",
				args: []*values.Value{
					values.NewString("%.2n"),
					values.NewFloat(0.01),
				},
				expected: "$0.01",
			},
			{
				name: "large number",
				args: []*values.Value{
					values.NewString("%.2n"),
					values.NewFloat(1000000),
				},
				expected: "$1,000,000.00",
			},

			// Width formatting
			{
				name: "right aligned",
				args: []*values.Value{
					values.NewString("%10.2n"),
					values.NewFloat(123.45),
				},
				expected: "   $123.45",
			},
			{
				name: "left aligned",
				args: []*values.Value{
					values.NewString("%-10.2n"),
					values.NewFloat(123.45),
				},
				expected: "$123.45   ",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := moneyFormatFunc.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("money_format() error = %v", err)
				}
				if result.Type != values.TypeString {
					t.Fatalf("money_format() returned %s, want string", result.Type)
				}
				if result.Data.(string) != tt.expected {
					t.Errorf("money_format() = %q, want %q", result.Data.(string), tt.expected)
				}
			})
		}
	})

	t.Run("mb_strlen", func(t *testing.T) {
		// Find the mb_strlen function
		var mbStrlenFunc *registry.Function
		for _, f := range functions {
			if f.Name == "mb_strlen" {
				mbStrlenFunc = f
				break
			}
		}

		if mbStrlenFunc == nil {
			t.Fatal("mb_strlen function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected int64
		}{
			// ASCII strings (same as strlen)
			{
				name: "ascii string",
				args: []*values.Value{
					values.NewString("hello"),
				},
				expected: 5,
			},
			{
				name: "empty string",
				args: []*values.Value{
					values.NewString(""),
				},
				expected: 0,
			},
			{
				name: "single character",
				args: []*values.Value{
					values.NewString("a"),
				},
				expected: 1,
			},
			{
				name: "numbers",
				args: []*values.Value{
					values.NewString("123"),
				},
				expected: 3,
			},

			// UTF-8 characters (different from strlen)
			{
				name: "accented characters",
				args: []*values.Value{
					values.NewString("café"),
				},
				expected: 4, // strlen would be 5
			},
			{
				name: "multiple accents",
				args: []*values.Value{
					values.NewString("naïve"),
				},
				expected: 5, // strlen would be 6
			},
			{
				name: "german umlaut",
				args: []*values.Value{
					values.NewString("München"),
				},
				expected: 7, // strlen would be 8
			},

			// Asian characters
			{
				name: "chinese characters",
				args: []*values.Value{
					values.NewString("你好"),
				},
				expected: 2, // strlen would be 6
			},
			{
				name: "japanese hiragana",
				args: []*values.Value{
					values.NewString("こんにちは"),
				},
				expected: 5, // strlen would be 15
			},

			// Emoji
			{
				name: "single emoji",
				args: []*values.Value{
					values.NewString("🌟"),
				},
				expected: 1, // strlen would be 4
			},
			{
				name: "mixed text and emoji",
				args: []*values.Value{
					values.NewString("Hello 🌍 World"),
				},
				expected: 13, // strlen would be 16
			},

			// Mixed content
			{
				name: "mixed ascii and chinese",
				args: []*values.Value{
					values.NewString("Hello 世界"),
				},
				expected: 8, // strlen would be 12
			},

			// Edge cases
			{
				name: "newline character",
				args: []*values.Value{
					values.NewString("\n"),
				},
				expected: 1,
			},
			{
				name: "tab character",
				args: []*values.Value{
					values.NewString("\t"),
				},
				expected: 1,
			},
			{
				name: "string with newline",
				args: []*values.Value{
					values.NewString("Hello\nWorld"),
				},
				expected: 11,
			},

			// Special Unicode
			{
				name: "copyright symbols",
				args: []*values.Value{
					values.NewString("©®™"),
				},
				expected: 3, // strlen would be 7
			},
			{
				name: "greek letters",
				args: []*values.Value{
					values.NewString("αβγδε"),
				},
				expected: 5, // strlen would be 10
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := mbStrlenFunc.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("mb_strlen() error = %v", err)
				}
				if result.Type != values.TypeInt {
					t.Fatalf("mb_strlen() returned %s, want int", result.Type)
				}
				if result.Data.(int64) != tt.expected {
					t.Errorf("mb_strlen() = %d, want %d", result.Data.(int64), tt.expected)
				}
			})
		}
	})

	t.Run("mb_substr", func(t *testing.T) {
		// Find the mb_substr function
		var mbSubstrFunc *registry.Function
		for _, f := range functions {
			if f.Name == "mb_substr" {
				mbSubstrFunc = f
				break
			}
		}

		if mbSubstrFunc == nil {
			t.Fatal("mb_substr function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected string
		}{
			// Basic functionality
			{
				name: "full string",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewInt(0),
					values.NewInt(5),
				},
				expected: "hello",
			},
			{
				name: "middle portion",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewInt(1),
					values.NewInt(3),
				},
				expected: "ell",
			},
			{
				name: "start portion",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewInt(0),
					values.NewInt(3),
				},
				expected: "hel",
			},
			{
				name: "from position to end (no length)",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewInt(2),
				},
				expected: "llo",
			},

			// UTF-8 characters
			{
				name: "full accented string",
				args: []*values.Value{
					values.NewString("café"),
					values.NewInt(0),
					values.NewInt(4),
				},
				expected: "café",
			},
			{
				name: "middle of accented",
				args: []*values.Value{
					values.NewString("café"),
					values.NewInt(1),
					values.NewInt(2),
				},
				expected: "af",
			},
			{
				name: "last accented char",
				args: []*values.Value{
					values.NewString("café"),
					values.NewInt(3),
					values.NewInt(1),
				},
				expected: "é",
			},

			// Asian characters
			{
				name: "first two chinese",
				args: []*values.Value{
					values.NewString("你好世界"),
					values.NewInt(0),
					values.NewInt(2),
				},
				expected: "你好",
			},
			{
				name: "middle two chinese",
				args: []*values.Value{
					values.NewString("你好世界"),
					values.NewInt(1),
					values.NewInt(2),
				},
				expected: "好世",
			},
			{
				name: "last two chinese (no length)",
				args: []*values.Value{
					values.NewString("你好世界"),
					values.NewInt(2),
				},
				expected: "世界",
			},

			// Emoji
			{
				name: "first emoji",
				args: []*values.Value{
					values.NewString("🌟⭐🌙"),
					values.NewInt(0),
					values.NewInt(1),
				},
				expected: "🌟",
			},
			{
				name: "middle emoji",
				args: []*values.Value{
					values.NewString("🌟⭐🌙"),
					values.NewInt(1),
					values.NewInt(1),
				},
				expected: "⭐",
			},
			{
				name: "emoji in mixed text",
				args: []*values.Value{
					values.NewString("Hello 🌍 World"),
					values.NewInt(6),
					values.NewInt(1),
				},
				expected: "🌍",
			},

			// Negative start positions
			{
				name: "last character",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewInt(-1),
					values.NewInt(1),
				},
				expected: "o",
			},
			{
				name: "last two characters",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewInt(-2),
					values.NewInt(2),
				},
				expected: "lo",
			},
			{
				name: "last chinese char",
				args: []*values.Value{
					values.NewString("你好"),
					values.NewInt(-1),
					values.NewInt(1),
				},
				expected: "好",
			},

			// Zero and negative lengths
			{
				name: "zero length",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewInt(1),
					values.NewInt(0),
				},
				expected: "",
			},
			{
				name: "negative length",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewInt(1),
					values.NewInt(-1),
				},
				expected: "ell",
			},

			// Edge cases
			{
				name: "empty string",
				args: []*values.Value{
					values.NewString(""),
					values.NewInt(0),
					values.NewInt(1),
				},
				expected: "",
			},
			{
				name: "start beyond string",
				args: []*values.Value{
					values.NewString("hello"),
					values.NewInt(10),
					values.NewInt(5),
				},
				expected: "",
			},
			{
				name: "single emoji",
				args: []*values.Value{
					values.NewString("🌟"),
					values.NewInt(0),
					values.NewInt(1),
				},
				expected: "🌟",
			},
			{
				name: "beyond single emoji",
				args: []*values.Value{
					values.NewString("🌟"),
					values.NewInt(1),
					values.NewInt(1),
				},
				expected: "",
			},

			// Mixed content
			{
				name: "ascii plus space",
				args: []*values.Value{
					values.NewString("Hello 世界"),
					values.NewInt(0),
					values.NewInt(6),
				},
				expected: "Hello ",
			},
			{
				name: "chinese portion",
				args: []*values.Value{
					values.NewString("Hello 世界"),
					values.NewInt(6),
					values.NewInt(2),
				},
				expected: "世界",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := mbSubstrFunc.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("mb_substr() error = %v", err)
				}
				if result.Type != values.TypeString {
					t.Fatalf("mb_substr() returned %s, want string", result.Type)
				}
				if result.Data.(string) != tt.expected {
					t.Errorf("mb_substr() = %q, want %q", result.Data.(string), tt.expected)
				}
			})
		}
	})

	t.Run("mb_strtolower", func(t *testing.T) {
		// Find the mb_strtolower function
		var mbStrtolowerFunc *registry.Function
		for _, f := range functions {
			if f.Name == "mb_strtolower" {
				mbStrtolowerFunc = f
				break
			}
		}

		if mbStrtolowerFunc == nil {
			t.Fatal("mb_strtolower function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected string
		}{
			// Basic ASCII tests
			{
				name: "simple uppercase",
				args: []*values.Value{
					values.NewString("HELLO"),
				},
				expected: "hello",
			},
			{
				name: "mixed case",
				args: []*values.Value{
					values.NewString("Hello"),
				},
				expected: "hello",
			},
			{
				name: "already lowercase",
				args: []*values.Value{
					values.NewString("hello"),
				},
				expected: "hello",
			},
			{
				name: "multiple words",
				args: []*values.Value{
					values.NewString("HELLO WORLD"),
				},
				expected: "hello world",
			},
			{
				name: "mixed case words",
				args: []*values.Value{
					values.NewString("HeLLo WoRLd"),
				},
				expected: "hello world",
			},

			// Empty string
			{
				name: "empty string",
				args: []*values.Value{
					values.NewString(""),
				},
				expected: "",
			},

			// Numbers and special characters
			{
				name: "numbers only",
				args: []*values.Value{
					values.NewString("123"),
				},
				expected: "123",
			},
			{
				name: "mixed letters and numbers",
				args: []*values.Value{
					values.NewString("Hello123"),
				},
				expected: "hello123",
			},
			{
				name: "special characters",
				args: []*values.Value{
					values.NewString("HELLO!@#$"),
				},
				expected: "hello!@#$",
			},

			// Accented characters (basic Latin)
			{
				name: "café uppercase",
				args: []*values.Value{
					values.NewString("CAFÉ"),
				},
				expected: "café",
			},
			{
				name: "naïve uppercase",
				args: []*values.Value{
					values.NewString("NAÏVE"),
				},
				expected: "naïve",
			},
			{
				name: "résumé uppercase",
				args: []*values.Value{
					values.NewString("RÉSUMÉ"),
				},
				expected: "résumé",
			},

			// German umlauts
			{
				name: "german umlaut",
				args: []*values.Value{
					values.NewString("MÜNCHEN"),
				},
				expected: "münchen",
			},

			// Nordic characters
			{
				name: "nordic characters",
				args: []*values.Value{
					values.NewString("ØÆÅ"),
				},
				expected: "øæå",
			},

			// Greek alphabet
			{
				name: "greek letters",
				args: []*values.Value{
					values.NewString("ΑΒΓΔ"),
				},
				expected: "αβγδ",
			},

			// Cyrillic
			{
				name: "cyrillic russian",
				args: []*values.Value{
					values.NewString("ПРИВЕТ"),
				},
				expected: "привет",
			},

			// Single characters
			{
				name: "single uppercase A",
				args: []*values.Value{
					values.NewString("A"),
				},
				expected: "a",
			},
			{
				name: "single uppercase Z",
				args: []*values.Value{
					values.NewString("Z"),
				},
				expected: "z",
			},
			{
				name: "single accented",
				args: []*values.Value{
					values.NewString("Ñ"),
				},
				expected: "ñ",
			},

			// Whitespace preservation
			{
				name: "spaces preserved",
				args: []*values.Value{
					values.NewString("  HELLO  "),
				},
				expected: "  hello  ",
			},
			{
				name: "newline preserved",
				args: []*values.Value{
					values.NewString("HELLO\nWORLD"),
				},
				expected: "hello\nworld",
			},
			{
				name: "tab preserved",
				args: []*values.Value{
					values.NewString("HELLO\tWORLD"),
				},
				expected: "hello\tworld",
			},

			// Real world examples
			{
				name: "mixed content with numbers",
				args: []*values.Value{
					values.NewString("Hello 123 WORLD!"),
				},
				expected: "hello 123 world!",
			},
			{
				name: "restaurant example",
				args: []*values.Value{
					values.NewString("Café & Restaurant"),
				},
				expected: "café & restaurant",
			},
			{
				name: "technical terms",
				args: []*values.Value{
					values.NewString("HTML & CSS"),
				},
				expected: "html & css",
			},

			// Longer strings
			{
				name: "pangram",
				args: []*values.Value{
					values.NewString("THE QUICK BROWN FOX JUMPS OVER THE LAZY DOG"),
				},
				expected: "the quick brown fox jumps over the lazy dog",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := mbStrtolowerFunc.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("mb_strtolower() error = %v", err)
				}
				if result.Type != values.TypeString {
					t.Fatalf("mb_strtolower() returned %s, want string", result.Type)
				}
				if result.Data.(string) != tt.expected {
					t.Errorf("mb_strtolower() = %q, want %q", result.Data.(string), tt.expected)
				}
			})
		}
	})

	t.Run("mb_strtoupper", func(t *testing.T) {
		// Find the mb_strtoupper function
		var mbStrtoupperFunc *registry.Function
		for _, f := range functions {
			if f.Name == "mb_strtoupper" {
				mbStrtoupperFunc = f
				break
			}
		}

		if mbStrtoupperFunc == nil {
			t.Fatal("mb_strtoupper function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected string
		}{
			// Basic ASCII tests
			{
				name: "simple lowercase",
				args: []*values.Value{
					values.NewString("hello"),
				},
				expected: "HELLO",
			},
			{
				name: "mixed case",
				args: []*values.Value{
					values.NewString("Hello"),
				},
				expected: "HELLO",
			},
			{
				name: "already uppercase",
				args: []*values.Value{
					values.NewString("HELLO"),
				},
				expected: "HELLO",
			},
			{
				name: "multiple words",
				args: []*values.Value{
					values.NewString("hello world"),
				},
				expected: "HELLO WORLD",
			},
			{
				name: "mixed case words",
				args: []*values.Value{
					values.NewString("heLLo WoRLd"),
				},
				expected: "HELLO WORLD",
			},

			// Empty string
			{
				name: "empty string",
				args: []*values.Value{
					values.NewString(""),
				},
				expected: "",
			},

			// Numbers and special characters
			{
				name: "numbers only",
				args: []*values.Value{
					values.NewString("123"),
				},
				expected: "123",
			},
			{
				name: "mixed letters and numbers",
				args: []*values.Value{
					values.NewString("hello123"),
				},
				expected: "HELLO123",
			},
			{
				name: "special characters",
				args: []*values.Value{
					values.NewString("hello!@#$"),
				},
				expected: "HELLO!@#$",
			},

			// Accented characters (basic Latin)
			{
				name: "café lowercase",
				args: []*values.Value{
					values.NewString("café"),
				},
				expected: "CAFÉ",
			},
			{
				name: "naïve lowercase",
				args: []*values.Value{
					values.NewString("naïve"),
				},
				expected: "NAÏVE",
			},
			{
				name: "résumé lowercase",
				args: []*values.Value{
					values.NewString("résumé"),
				},
				expected: "RÉSUMÉ",
			},

			// German umlauts
			{
				name: "german umlaut",
				args: []*values.Value{
					values.NewString("münchen"),
				},
				expected: "MÜNCHEN",
			},

			// Nordic characters
			{
				name: "nordic characters",
				args: []*values.Value{
					values.NewString("øæå"),
				},
				expected: "ØÆÅ",
			},

			// Greek alphabet
			{
				name: "greek letters",
				args: []*values.Value{
					values.NewString("αβγδ"),
				},
				expected: "ΑΒΓΔ",
			},

			// Cyrillic
			{
				name: "cyrillic russian",
				args: []*values.Value{
					values.NewString("привет"),
				},
				expected: "ПРИВЕТ",
			},

			// Single characters
			{
				name: "single lowercase a",
				args: []*values.Value{
					values.NewString("a"),
				},
				expected: "A",
			},
			{
				name: "single lowercase z",
				args: []*values.Value{
					values.NewString("z"),
				},
				expected: "Z",
			},
			{
				name: "single accented",
				args: []*values.Value{
					values.NewString("ñ"),
				},
				expected: "Ñ",
			},

			// Whitespace preservation
			{
				name: "spaces preserved",
				args: []*values.Value{
					values.NewString("  hello  "),
				},
				expected: "  HELLO  ",
			},
			{
				name: "newline preserved",
				args: []*values.Value{
					values.NewString("hello\nworld"),
				},
				expected: "HELLO\nWORLD",
			},
			{
				name: "tab preserved",
				args: []*values.Value{
					values.NewString("hello\tworld"),
				},
				expected: "HELLO\tWORLD",
			},

			// Real world examples
			{
				name: "mixed content with numbers",
				args: []*values.Value{
					values.NewString("hello 123 world!"),
				},
				expected: "HELLO 123 WORLD!",
			},
			{
				name: "restaurant example",
				args: []*values.Value{
					values.NewString("café & restaurant"),
				},
				expected: "CAFÉ & RESTAURANT",
			},
			{
				name: "technical terms",
				args: []*values.Value{
					values.NewString("html & css"),
				},
				expected: "HTML & CSS",
			},

			// Longer strings
			{
				name: "pangram",
				args: []*values.Value{
					values.NewString("the quick brown fox jumps over the lazy dog"),
				},
				expected: "THE QUICK BROWN FOX JUMPS OVER THE LAZY DOG",
			},

			// Special Unicode case: German sharp s conversion
			{
				name: "german sharp s conversion",
				args: []*values.Value{
					values.NewString("ß"),
				},
				expected: "SS",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := mbStrtoupperFunc.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("mb_strtoupper() error = %v", err)
				}
				if result.Type != values.TypeString {
					t.Fatalf("mb_strtoupper() returned %s, want string", result.Type)
				}
				if result.Data.(string) != tt.expected {
					t.Errorf("mb_strtoupper() = %q, want %q", result.Data.(string), tt.expected)
				}
			})
		}
	})
}
