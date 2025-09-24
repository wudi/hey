package runtime

import (
	"sync"
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestIniGet(t *testing.T) {
	// Reset storage for clean tests
	globalIniStorage = nil
	iniStorageOnce = sync.Once{}

	functions := GetIniFunctions()
	var iniGet *registry.Function
	for _, f := range functions {
		if f.Name == "ini_get" {
			iniGet = f
			break
		}
	}
	if iniGet == nil {
		t.Fatal("ini_get function not found")
	}

	tests := []struct {
		name     string
		args     []*values.Value
		expected interface{}
	}{
		{
			name:     "get existing setting",
			args:     []*values.Value{values.NewString("memory_limit")},
			expected: "-1",
		},
		{
			name:     "get existing empty setting",
			args:     []*values.Value{values.NewString("display_errors")},
			expected: "",
		},
		{
			name:     "get nonexistent setting",
			args:     []*values.Value{values.NewString("nonexistent_setting")},
			expected: false,
		},
		{
			name:     "no arguments",
			args:     []*values.Value{},
			expected: false,
		},
		{
			name:     "nil argument",
			args:     []*values.Value{nil},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := iniGet.Builtin(nil, tt.args)
			if err != nil {
				t.Fatalf("ini_get failed: %v", err)
			}

			switch expected := tt.expected.(type) {
			case string:
				if result.ToString() != expected {
					t.Errorf("expected %q, got %q", expected, result.ToString())
				}
			case bool:
				if result.ToBool() != expected {
					t.Errorf("expected %v, got %v", expected, result.ToBool())
				}
			}
		})
	}
}

func TestIniSet(t *testing.T) {
	// Reset storage for clean tests
	globalIniStorage = nil
	iniStorageOnce = sync.Once{}

	functions := GetIniFunctions()
	var iniSet *registry.Function
	var iniGet *registry.Function
	for _, f := range functions {
		if f.Name == "ini_set" {
			iniSet = f
		}
		if f.Name == "ini_get" {
			iniGet = f
		}
	}
	if iniSet == nil || iniGet == nil {
		t.Fatal("ini_set or ini_get function not found")
	}

	tests := []struct {
		name        string
		option      string
		value       string
		expectedOld string
		expectFalse bool
	}{
		{
			name:        "set existing setting",
			option:      "display_errors",
			value:       "1",
			expectedOld: "",
			expectFalse: false,
		},
		{
			name:        "set nonexistent setting",
			option:      "nonexistent_setting",
			value:       "value",
			expectedOld: "",
			expectFalse: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []*values.Value{
				values.NewString(tt.option),
				values.NewString(tt.value),
			}
			result, err := iniSet.Builtin(nil, args)
			if err != nil {
				t.Fatalf("ini_set failed: %v", err)
			}

			if tt.expectFalse {
				if result.ToBool() != false {
					t.Errorf("expected false, got %v", result.ToBool())
				}
			} else {
				if result.ToString() != tt.expectedOld {
					t.Errorf("expected old value %q, got %q", tt.expectedOld, result.ToString())
				}

				// Verify the value was actually set
				getResult, err := iniGet.Builtin(nil, []*values.Value{values.NewString(tt.option)})
				if err != nil {
					t.Fatalf("ini_get verification failed: %v", err)
				}
				if getResult.ToString() != tt.value {
					t.Errorf("value not set correctly: expected %q, got %q", tt.value, getResult.ToString())
				}
			}
		})
	}
}

func TestIniRestore(t *testing.T) {
	// Reset storage for clean tests
	globalIniStorage = nil
	iniStorageOnce = sync.Once{}

	functions := GetIniFunctions()
	var iniSet *registry.Function
	var iniGet *registry.Function
	var iniRestore *registry.Function
	for _, f := range functions {
		if f.Name == "ini_set" {
			iniSet = f
		}
		if f.Name == "ini_get" {
			iniGet = f
		}
		if f.Name == "ini_restore" {
			iniRestore = f
		}
	}
	if iniSet == nil || iniGet == nil || iniRestore == nil {
		t.Fatal("required ini functions not found")
	}

	// Set memory_limit to a new value
	_, err := iniSet.Builtin(nil, []*values.Value{
		values.NewString("memory_limit"),
		values.NewString("256M"),
	})
	if err != nil {
		t.Fatalf("ini_set failed: %v", err)
	}

	// Verify it was set
	result, err := iniGet.Builtin(nil, []*values.Value{values.NewString("memory_limit")})
	if err != nil {
		t.Fatalf("ini_get failed: %v", err)
	}
	if result.ToString() != "256M" {
		t.Errorf("expected 256M, got %q", result.ToString())
	}

	// Restore the original value
	_, err = iniRestore.Builtin(nil, []*values.Value{values.NewString("memory_limit")})
	if err != nil {
		t.Fatalf("ini_restore failed: %v", err)
	}

	// Verify it was restored
	result, err = iniGet.Builtin(nil, []*values.Value{values.NewString("memory_limit")})
	if err != nil {
		t.Fatalf("ini_get after restore failed: %v", err)
	}
	if result.ToString() != "-1" {
		t.Errorf("expected -1 (original value), got %q", result.ToString())
	}
}

func TestIniGetAll(t *testing.T) {
	// Reset storage for clean tests
	globalIniStorage = nil
	iniStorageOnce = sync.Once{}

	functions := GetIniFunctions()
	var iniGetAll *registry.Function
	for _, f := range functions {
		if f.Name == "ini_get_all" {
			iniGetAll = f
			break
		}
	}
	if iniGetAll == nil {
		t.Fatal("ini_get_all function not found")
	}

	// Test getting all settings
	result, err := iniGetAll.Builtin(nil, []*values.Value{})
	if err != nil {
		t.Fatalf("ini_get_all failed: %v", err)
	}

	if !result.IsArray() {
		t.Fatalf("expected array, got %T", result)
	}

	if result.ArrayCount() == 0 {
		t.Errorf("expected non-empty array")
	}

	// Check that memory_limit setting exists and has correct structure
	memoryLimitKey := values.NewString("memory_limit")
	setting := result.ArrayGet(memoryLimitKey)
	if setting.IsNull() {
		t.Errorf("memory_limit setting not found in ini_get_all result")
	} else {
		if !setting.IsArray() {
			t.Errorf("memory_limit setting should be an array")
		} else {
			globalValueKey := values.NewString("global_value")
			localValueKey := values.NewString("local_value")
			accessKey := values.NewString("access")

			globalValue := setting.ArrayGet(globalValueKey)
			localValue := setting.ArrayGet(localValueKey)
			accessValue := setting.ArrayGet(accessKey)

			if globalValue.IsNull() {
				t.Errorf("global_value key missing from memory_limit setting")
			}
			if localValue.IsNull() {
				t.Errorf("local_value key missing from memory_limit setting")
			}
			if accessValue.IsNull() {
				t.Errorf("access key missing from memory_limit setting")
			}
		}
	}

	// Test with invalid extension (should return false)
	result, err = iniGetAll.Builtin(nil, []*values.Value{values.NewString("NonexistentExtension")})
	if err != nil {
		t.Fatalf("ini_get_all with invalid extension failed: %v", err)
	}
	if result.ToBool() != false {
		t.Errorf("expected false for invalid extension, got %v", result.ToBool())
	}
}

func TestIniParseQuantity(t *testing.T) {
	functions := GetIniFunctions()
	var iniParseQuantity *registry.Function
	for _, f := range functions {
		if f.Name == "ini_parse_quantity" {
			iniParseQuantity = f
			break
		}
	}
	if iniParseQuantity == nil {
		t.Fatal("ini_parse_quantity function not found")
	}

	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{"plain number", "1024", 1024},
		{"uppercase K", "2K", 2048},
		{"uppercase M", "4M", 4194304},
		{"uppercase G", "1G", 1073741824},
		{"lowercase k", "512k", 524288},
		{"lowercase m", "128m", 134217728},
		{"lowercase g", "2g", 2147483648},
		{"invalid input", "invalid", 0},
		{"empty string", "", 0},
		{"zero", "0", 0},
		{"with spaces", " 64M ", 67108864},
		{"decimal number", "1.5K", 1536}, // 1.5 * 1024
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := iniParseQuantity.Builtin(nil, []*values.Value{values.NewString(tt.input)})
			if err != nil {
				t.Fatalf("ini_parse_quantity failed: %v", err)
			}

			if result.ToInt() != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result.ToInt())
			}
		})
	}

	// Test with no arguments
	result, err := iniParseQuantity.Builtin(nil, []*values.Value{})
	if err != nil {
		t.Fatalf("ini_parse_quantity with no args failed: %v", err)
	}
	if result.ToInt() != 0 {
		t.Errorf("expected 0 for no args, got %d", result.ToInt())
	}

	// Test with nil argument
	result, err = iniParseQuantity.Builtin(nil, []*values.Value{nil})
	if err != nil {
		t.Fatalf("ini_parse_quantity with nil arg failed: %v", err)
	}
	if result.ToInt() != 0 {
		t.Errorf("expected 0 for nil arg, got %d", result.ToInt())
	}
}