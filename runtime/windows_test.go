package runtime

import (
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestSapiWindowsVt100Support(t *testing.T) {
	functions := GetWindowsFunctions()
	var sapiWindowsVt100Fn *registry.Function

	for _, fn := range functions {
		if fn.Name == "sapi_windows_vt100_support" {
			sapiWindowsVt100Fn = fn
			break
		}
	}

	if sapiWindowsVt100Fn == nil {
		t.Fatal("sapi_windows_vt100_support function not found")
	}

	ctx := &functionTestContext{}

	// Test with no arguments - should return false
	result, err := sapiWindowsVt100Fn.Builtin(ctx, []*values.Value{})
	if err != nil {
		t.Errorf("Expected no error for no arguments, got %v", err)
	}
	if result.Type != values.TypeBool || result.Data.(bool) != false {
		t.Errorf("Expected false for no arguments, got %v", result)
	}

	// Test with null argument - should return false
	result, err = sapiWindowsVt100Fn.Builtin(ctx, []*values.Value{values.NewNull()})
	if err != nil {
		t.Errorf("Expected no error for null argument, got %v", err)
	}
	if result.Type != values.TypeBool || result.Data.(bool) != false {
		t.Errorf("Expected false for null argument, got %v", result)
	}

	// Test with non-resource argument - should return false
	result, err = sapiWindowsVt100Fn.Builtin(ctx, []*values.Value{values.NewString("not_a_resource")})
	if err != nil {
		t.Errorf("Expected no error for non-resource argument, got %v", err)
	}
	if result.Type != values.TypeBool || result.Data.(bool) != false {
		t.Errorf("Expected false for non-resource argument, got %v", result)
	}

	// Test with resource argument - should return false
	resourceValue := &values.Value{Type: values.TypeResource, Data: "mock_resource"}
	result, err = sapiWindowsVt100Fn.Builtin(ctx, []*values.Value{resourceValue})
	if err != nil {
		t.Errorf("Expected no error for resource argument, got %v", err)
	}
	if result.Type != values.TypeBool || result.Data.(bool) != false {
		t.Errorf("Expected false for resource argument, got %v", result)
	}

	// Test with resource and enable parameter - should return false
	result, err = sapiWindowsVt100Fn.Builtin(ctx, []*values.Value{resourceValue, values.NewBool(true)})
	if err != nil {
		t.Errorf("Expected no error for resource and enable arguments, got %v", err)
	}
	if result.Type != values.TypeBool || result.Data.(bool) != false {
		t.Errorf("Expected false for resource and enable arguments, got %v", result)
	}

	// Test with resource and disable parameter - should return false
	result, err = sapiWindowsVt100Fn.Builtin(ctx, []*values.Value{resourceValue, values.NewBool(false)})
	if err != nil {
		t.Errorf("Expected no error for resource and disable arguments, got %v", err)
	}
	if result.Type != values.TypeBool || result.Data.(bool) != false {
		t.Errorf("Expected false for resource and disable arguments, got %v", result)
	}
}