package runtime

import (
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// TestUrlPathFunctions tests URL and path parsing functions
func TestUrlPathFunctions(t *testing.T) {
	functions := GetStringFunctions()
	functionMap := make(map[string]*registry.Function)
	for _, fn := range functions {
		functionMap[fn.Name] = fn
	}

	t.Run("parse_url", func(t *testing.T) {
		fn := functionMap["parse_url"]
		if fn == nil {
			t.Fatal("parse_url function not found")
		}

		// Test full URL parsing
		result, err := fn.Builtin(nil, []*values.Value{
			values.NewString("https://user:pass@example.com:8080/path/to/page?query=value#section"),
		})

		if err != nil {
			t.Errorf("parse_url error: %v", err)
			return
		}

		if !result.IsArray() {
			t.Error("parse_url should return array")
			return
		}

		scheme := result.ArrayGet(values.NewString("scheme"))
		if scheme.ToString() != "https" {
			t.Errorf("expected scheme 'https', got '%s'", scheme.ToString())
		}

		host := result.ArrayGet(values.NewString("host"))
		if host.ToString() != "example.com" {
			t.Errorf("expected host 'example.com', got '%s'", host.ToString())
		}

		port := result.ArrayGet(values.NewString("port"))
		if port.ToInt() != 8080 {
			t.Errorf("expected port 8080, got %d", port.ToInt())
		}

		user := result.ArrayGet(values.NewString("user"))
		if user.ToString() != "user" {
			t.Errorf("expected user 'user', got '%s'", user.ToString())
		}

		pass := result.ArrayGet(values.NewString("pass"))
		if pass.ToString() != "pass" {
			t.Errorf("expected pass 'pass', got '%s'", pass.ToString())
		}

		path := result.ArrayGet(values.NewString("path"))
		if path.ToString() != "/path/to/page" {
			t.Errorf("expected path '/path/to/page', got '%s'", path.ToString())
		}

		query := result.ArrayGet(values.NewString("query"))
		if query.ToString() != "query=value" {
			t.Errorf("expected query 'query=value', got '%s'", query.ToString())
		}

		fragment := result.ArrayGet(values.NewString("fragment"))
		if fragment.ToString() != "section" {
			t.Errorf("expected fragment 'section', got '%s'", fragment.ToString())
		}

		// Test component extraction
		hostOnly, err := fn.Builtin(nil, []*values.Value{
			values.NewString("https://example.com:8080/path"),
			values.NewInt(1), // PHP_URL_HOST
		})

		if err != nil {
			t.Errorf("parse_url component error: %v", err)
			return
		}

		if hostOnly.ToString() != "example.com" {
			t.Errorf("expected host component 'example.com', got '%s'", hostOnly.ToString())
		}
	})

	t.Run("pathinfo", func(t *testing.T) {
		fn := functionMap["pathinfo"]
		if fn == nil {
			t.Fatal("pathinfo function not found")
		}

		// Test full path parsing
		result, err := fn.Builtin(nil, []*values.Value{
			values.NewString("/var/www/html/index.php"),
		})

		if err != nil {
			t.Errorf("pathinfo error: %v", err)
			return
		}

		if !result.IsArray() {
			t.Error("pathinfo should return array")
			return
		}

		dirname := result.ArrayGet(values.NewString("dirname"))
		if dirname.ToString() != "/var/www/html" {
			t.Errorf("expected dirname '/var/www/html', got '%s'", dirname.ToString())
		}

		basename := result.ArrayGet(values.NewString("basename"))
		if basename.ToString() != "index.php" {
			t.Errorf("expected basename 'index.php', got '%s'", basename.ToString())
		}

		extension := result.ArrayGet(values.NewString("extension"))
		if extension.ToString() != "php" {
			t.Errorf("expected extension 'php', got '%s'", extension.ToString())
		}

		filename := result.ArrayGet(values.NewString("filename"))
		if filename.ToString() != "index" {
			t.Errorf("expected filename 'index', got '%s'", filename.ToString())
		}

		// Test specific component
		extOnly, err := fn.Builtin(nil, []*values.Value{
			values.NewString("/path/file.txt"),
			values.NewInt(4), // PATHINFO_EXTENSION
		})

		if err != nil {
			t.Errorf("pathinfo component error: %v", err)
			return
		}

		if extOnly.ToString() != "txt" {
			t.Errorf("expected extension 'txt', got '%s'", extOnly.ToString())
		}

		// Test file without extension
		noExt, err := fn.Builtin(nil, []*values.Value{
			values.NewString("/path/file"),
		})

		if err != nil {
			t.Errorf("pathinfo no extension error: %v", err)
			return
		}

		extField := noExt.ArrayGet(values.NewString("extension"))
		if extField != nil && !extField.IsNull() {
			t.Error("expected no extension field for file without extension")
		}
	})

	t.Run("bin2hex", func(t *testing.T) {
		fn := functionMap["bin2hex"]
		if fn == nil {
			t.Fatal("bin2hex function not found")
		}

		tests := []struct {
			input    string
			expected string
		}{
			{"Hello", "48656c6c6f"},
			{"123", "313233"},
			{"", ""},
			{"\x00\x01\x02", "000102"},
		}

		for _, tt := range tests {
			result, err := fn.Builtin(nil, []*values.Value{
				values.NewString(tt.input),
			})

			if err != nil {
				t.Errorf("bin2hex error: %v", err)
				return
			}

			if result.ToString() != tt.expected {
				t.Errorf("bin2hex('%s'): expected '%s', got '%s'", tt.input, tt.expected, result.ToString())
			}
		}
	})

	t.Run("hex2bin", func(t *testing.T) {
		fn := functionMap["hex2bin"]
		if fn == nil {
			t.Fatal("hex2bin function not found")
		}

		tests := []struct {
			input    string
			expected string
			valid    bool
		}{
			{"48656c6c6f", "Hello", true},
			{"313233", "123", true},
			{"", "", true},
			{"000102", "\x00\x01\x02", true},
			{"xyz", "", false}, // Invalid hex
			{"1", "", false},   // Odd length
		}

		for _, tt := range tests {
			result, err := fn.Builtin(nil, []*values.Value{
				values.NewString(tt.input),
			})

			if err != nil {
				t.Errorf("hex2bin error: %v", err)
				return
			}

			if tt.valid {
				if result.ToString() != tt.expected {
					t.Errorf("hex2bin('%s'): expected '%s', got '%s'", tt.input, tt.expected, result.ToString())
				}
			} else {
				if !result.IsBool() || result.ToBool() {
					t.Errorf("hex2bin('%s'): expected false for invalid input", tt.input)
				}
			}
		}
	})
}