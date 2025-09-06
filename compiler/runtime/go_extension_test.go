package runtime

import (
	"testing"
	"time"

	"github.com/wudi/php-parser/compiler/values"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoExtension_Creation(t *testing.T) {
	ext := NewGoExtension()
	
	assert.Equal(t, "go", ext.GetName())
	assert.Equal(t, "1.0.0", ext.GetVersion())
	assert.Contains(t, ext.GetDescription(), "goroutine")
	assert.Equal(t, 50, ext.GetLoadOrder())
}

func TestGoExtension_Registration(t *testing.T) {
	// Initialize runtime
	err := Bootstrap()
	require.NoError(t, err)
	
	// Create and register extension
	ext := NewGoExtension()
	err = ext.Register(GlobalRegistry)
	require.NoError(t, err)
	
	// Check that the 'go' function is registered
	goFunc, exists := GlobalRegistry.GetFunction("go")
	assert.True(t, exists)
	assert.NotNil(t, goFunc)
	assert.Equal(t, "go", goFunc.Name)
}

func TestGoExtension_GoFunction_ValidCallable(t *testing.T) {
	// Setup
	err := Bootstrap()
	require.NoError(t, err)
	
	ext := NewGoExtension()
	err = ext.Register(GlobalRegistry)
	require.NoError(t, err)
	
	// Create a test callable
	executed := false
	testHandler := func(args []*values.Value) (*values.Value, error) {
		executed = true
		return values.NewString("test_result"), nil
	}
	
	testClosure := values.NewClosure(testHandler, nil, "test_closure")
	
	// Get the go function
	goFunc, exists := GlobalRegistry.GetFunction("go")
	require.True(t, exists)
	
	// Call go() with the callable
	result, err := goFunc.Handler(nil, []*values.Value{testClosure})
	require.NoError(t, err)
	
	// Should return null immediately
	assert.True(t, result.IsNull())
	
	// Wait for goroutine to complete
	ext.WaitForAll()
	
	// The handler should have been executed
	assert.True(t, executed)
}

func TestGoExtension_GoFunction_NonCallable(t *testing.T) {
	// Setup
	err := Bootstrap()
	require.NoError(t, err)
	
	ext := NewGoExtension()
	err = ext.Register(GlobalRegistry)
	require.NoError(t, err)
	
	// Get the go function
	goFunc, exists := GlobalRegistry.GetFunction("go")
	require.True(t, exists)
	
	// Try with non-callable value
	nonCallable := values.NewString("not a function")
	result, err := goFunc.Handler(nil, []*values.Value{nonCallable})
	
	// Should return error
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "must be callable")
}

func TestGoExtension_GoFunction_WrongArgumentCount(t *testing.T) {
	// Setup
	err := Bootstrap()
	require.NoError(t, err)
	
	ext := NewGoExtension()
	err = ext.Register(GlobalRegistry)
	require.NoError(t, err)
	
	// Get the go function
	goFunc, exists := GlobalRegistry.GetFunction("go")
	require.True(t, exists)
	
	// Test with no arguments
	result, err := goFunc.Handler(nil, []*values.Value{})
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "expects exactly 1 argument")
	
	// Test with too many arguments
	testClosure := values.NewClosure(func(args []*values.Value) (*values.Value, error) {
		return values.NewNull(), nil
	}, nil, "test")
	
	result, err = goFunc.Handler(nil, []*values.Value{testClosure, testClosure})
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "expects exactly 1 argument")
}

func TestGoExtension_ConcurrentExecution(t *testing.T) {
	// Setup
	err := Bootstrap()
	require.NoError(t, err)
	
	ext := NewGoExtension()
	err = ext.Register(GlobalRegistry)
	require.NoError(t, err)
	
	// Get the go function
	goFunc, exists := GlobalRegistry.GetFunction("go")
	require.True(t, exists)
	
	// Track execution order
	execOrder := make([]int, 0)
	
	// Create multiple callables with different execution times
	for i := 0; i < 3; i++ {
		taskID := i
		testHandler := func(args []*values.Value) (*values.Value, error) {
			// Longer sleep for lower IDs to test concurrency
			time.Sleep(time.Duration(50*(3-taskID)) * time.Millisecond)
			execOrder = append(execOrder, taskID)
			return values.NewInt(int64(taskID)), nil
		}
		
		testClosure := values.NewClosure(testHandler, nil, "concurrent_test")
		
		// Execute concurrently
		result, err := goFunc.Handler(nil, []*values.Value{testClosure})
		require.NoError(t, err)
		assert.True(t, result.IsNull())
	}
	
	// Wait for all to complete
	ext.WaitForAll()
	
	// Should have all 3 executions
	assert.Len(t, execOrder, 3)
}

func TestGoExtension_BoundVariables(t *testing.T) {
	// Setup
	err := Bootstrap()
	require.NoError(t, err)
	
	ext := NewGoExtension()
	err = ext.Register(GlobalRegistry)
	require.NoError(t, err)
	
	// Create callable with bound variables
	boundVars := map[string]*values.Value{
		"message": values.NewString("test message"),
		"number":  values.NewInt(42),
	}
	
	executed := false
	testHandler := func(args []*values.Value) (*values.Value, error) {
		executed = true
		return values.NewString("bound_test_result"), nil
	}
	
	testClosure := values.NewClosure(testHandler, boundVars, "bound_test")
	
	// Get the go function and execute
	goFunc, exists := GlobalRegistry.GetFunction("go")
	require.True(t, exists)
	
	result, err := goFunc.Handler(nil, []*values.Value{testClosure})
	require.NoError(t, err)
	assert.True(t, result.IsNull())
	
	// Wait and verify execution
	ext.WaitForAll()
	assert.True(t, executed)
}

func TestGoExtension_Unregister(t *testing.T) {
	// Setup
	err := Bootstrap()
	require.NoError(t, err)
	
	ext := NewGoExtension()
	err = ext.Register(GlobalRegistry)
	require.NoError(t, err)
	
	// Verify registration
	_, exists := GlobalRegistry.GetFunction("go")
	assert.True(t, exists)
	
	// Unregister
	err = ext.Unregister(GlobalRegistry)
	require.NoError(t, err)
	
	// Verify function is gone
	_, exists = GlobalRegistry.GetFunction("go")
	assert.False(t, exists)
}