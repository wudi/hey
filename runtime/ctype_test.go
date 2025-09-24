package runtime

import (
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// TestCtypeFunctions tests the character type checking functions
func TestCtypeFunctions(t *testing.T) {
	functions := GetCtypeFunctions()
	functionMap := make(map[string]*registry.Function)
	for _, fn := range functions {
		functionMap[fn.Name] = fn
	}

	t.Run("ctype_alnum", func(t *testing.T) {
		fn := functionMap["ctype_alnum"]
		if fn == nil {
			t.Fatal("ctype_alnum function not found")
		}

		tests := []struct {
			input    string
			expected bool
		}{
			{"Hello123", true},
			{"Hello-123", false},
			{"Hello", true},
			{"123456", true},
			{"", false},
			{"Hello World", false},
			{"Test@123", false},
			{"abc", true},
			{"ABC", true},
		}

		for _, tt := range tests {
			t.Run("input_"+tt.input, func(t *testing.T) {
				result, err := fn.Builtin(nil, []*values.Value{
					values.NewString(tt.input),
				})

				if err != nil {
					t.Errorf("ctype_alnum error: %v", err)
					return
				}

				if result.ToBool() != tt.expected {
					t.Errorf("ctype_alnum('%s'): expected %v, got %v", tt.input, tt.expected, result.ToBool())
				}
			})
		}
	})

	t.Run("ctype_alpha", func(t *testing.T) {
		fn := functionMap["ctype_alpha"]
		if fn == nil {
			t.Fatal("ctype_alpha function not found")
		}

		tests := []struct {
			input    string
			expected bool
		}{
			{"Hello", true},
			{"Hello123", false},
			{"", false},
			{"123", false},
			{"Hello World", false},
			{"abc", true},
			{"ABC", true},
			{"àáâãäå", true},
		}

		for _, tt := range tests {
			t.Run("input_"+tt.input, func(t *testing.T) {
				result, err := fn.Builtin(nil, []*values.Value{
					values.NewString(tt.input),
				})

				if err != nil {
					t.Errorf("ctype_alpha error: %v", err)
					return
				}

				if result.ToBool() != tt.expected {
					t.Errorf("ctype_alpha('%s'): expected %v, got %v", tt.input, tt.expected, result.ToBool())
				}
			})
		}
	})

	t.Run("ctype_digit", func(t *testing.T) {
		fn := functionMap["ctype_digit"]
		if fn == nil {
			t.Fatal("ctype_digit function not found")
		}

		tests := []struct {
			input    string
			expected bool
		}{
			{"12345", true},
			{"123a5", false},
			{"", false},
			{"0", true},
			{"00123", true},
			{"12.34", false},
			{"-123", false},
			{"+123", false},
		}

		for _, tt := range tests {
			t.Run("input_"+tt.input, func(t *testing.T) {
				result, err := fn.Builtin(nil, []*values.Value{
					values.NewString(tt.input),
				})

				if err != nil {
					t.Errorf("ctype_digit error: %v", err)
					return
				}

				if result.ToBool() != tt.expected {
					t.Errorf("ctype_digit('%s'): expected %v, got %v", tt.input, tt.expected, result.ToBool())
				}
			})
		}
	})

	t.Run("ctype_lower", func(t *testing.T) {
		fn := functionMap["ctype_lower"]
		if fn == nil {
			t.Fatal("ctype_lower function not found")
		}

		tests := []struct {
			input    string
			expected bool
		}{
			{"hello", true},
			{"Hello", false},
			{"HELLO", false},
			{"", false},
			{"hello123", false}, // Contains numbers
			{"hello world", false}, // Contains space
			{"abc", true},
		}

		for _, tt := range tests {
			t.Run("input_"+tt.input, func(t *testing.T) {
				result, err := fn.Builtin(nil, []*values.Value{
					values.NewString(tt.input),
				})

				if err != nil {
					t.Errorf("ctype_lower error: %v", err)
					return
				}

				if result.ToBool() != tt.expected {
					t.Errorf("ctype_lower('%s'): expected %v, got %v", tt.input, tt.expected, result.ToBool())
				}
			})
		}
	})

	t.Run("ctype_upper", func(t *testing.T) {
		fn := functionMap["ctype_upper"]
		if fn == nil {
			t.Fatal("ctype_upper function not found")
		}

		tests := []struct {
			input    string
			expected bool
		}{
			{"HELLO", true},
			{"Hello", false},
			{"hello", false},
			{"", false},
			{"HELLO123", false}, // Contains numbers
			{"HELLO WORLD", false}, // Contains space
			{"ABC", true},
		}

		for _, tt := range tests {
			t.Run("input_"+tt.input, func(t *testing.T) {
				result, err := fn.Builtin(nil, []*values.Value{
					values.NewString(tt.input),
				})

				if err != nil {
					t.Errorf("ctype_upper error: %v", err)
					return
				}

				if result.ToBool() != tt.expected {
					t.Errorf("ctype_upper('%s'): expected %v, got %v", tt.input, tt.expected, result.ToBool())
				}
			})
		}
	})

	t.Run("ctype_space", func(t *testing.T) {
		fn := functionMap["ctype_space"]
		if fn == nil {
			t.Fatal("ctype_space function not found")
		}

		tests := []struct {
			input    string
			expected bool
		}{
			{"   ", true},
			{" a ", false},
			{"", false},
			{" ", true},
			{"\t\n\r", true},
			{"\t hello \n", false},
			{"  \t  ", true},
		}

		for _, tt := range tests {
			t.Run("input_"+tt.input, func(t *testing.T) {
				result, err := fn.Builtin(nil, []*values.Value{
					values.NewString(tt.input),
				})

				if err != nil {
					t.Errorf("ctype_space error: %v", err)
					return
				}

				if result.ToBool() != tt.expected {
					t.Errorf("ctype_space('%s'): expected %v, got %v", tt.input, tt.expected, result.ToBool())
				}
			})
		}
	})
}