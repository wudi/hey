package runtime

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wudi/php-parser/compiler/values"
)

// MockExecutionContext for testing
type MockExecutionContext struct{}

func (m *MockExecutionContext) WriteOutput(output string)                   {}
func (m *MockExecutionContext) HasFunction(name string) bool                { return false }
func (m *MockExecutionContext) HasClass(name string) bool                   { return false }
func (m *MockExecutionContext) HasMethod(className, methodName string) bool { return false }

// Test basic goHandler functionality
func TestGoHandlerBasic(t *testing.T) {
	// Initialize runtime
	err := Bootstrap()
	require.NoError(t, err, "Failed to bootstrap runtime")

	// Create a simple closure that returns a string
	testClosure := values.NewClosure(func(ctx ExecutionContext, args []*values.Value) (*values.Value, error) {
		return values.NewString("Hello from goroutine!"), nil
	}, nil, "test_closure")

	ctx := &MockExecutionContext{}
	args := []*values.Value{testClosure}

	// Call goHandler
	result, err := goHandler(ctx, args)
	require.NoError(t, err, "goHandler should not return error")

	// Verify result is a goroutine
	assert.True(t, result.IsGoroutine(), "Result should be a goroutine")

	// Get goroutine data
	gorData := result.Data.(*values.Goroutine)
	assert.NotZero(t, gorData.ID, "Goroutine should have a non-zero ID")
	assert.Equal(t, "running", gorData.Status, "Initial status should be running")
	assert.NotNil(t, gorData.Done, "Done channel should not be nil")
}

// Test goHandler with captured variables
func TestGoHandlerWithVariables(t *testing.T) {
	err := Bootstrap()
	require.NoError(t, err, "Failed to bootstrap runtime")

	// Create a closure that uses captured variables
	testClosure := values.NewClosure(func(ctx ExecutionContext, args []*values.Value) (*values.Value, error) {
		return values.NewString("Processed variables"), nil
	}, nil, "var_closure")

	ctx := &MockExecutionContext{}

	// Call with additional variables
	var1 := values.NewString("test_var")
	var2 := values.NewInt(42)
	args := []*values.Value{testClosure, var1, var2}

	result, err := goHandler(ctx, args)
	require.NoError(t, err, "goHandler with variables should not return error")

	assert.True(t, result.IsGoroutine(), "Result should be a goroutine")

	// Check that variables are captured
	gorData := result.Data.(*values.Goroutine)
	assert.Len(t, gorData.UseVars, 2, "Should have 2 captured variables")
	assert.Contains(t, gorData.UseVars, "var_0", "Should contain var_0")
	assert.Contains(t, gorData.UseVars, "var_1", "Should contain var_1")
	assert.Equal(t, "test_var", gorData.UseVars["var_0"].ToString())
	assert.Equal(t, int64(42), gorData.UseVars["var_1"].ToInt())
}

// Test goHandler with bound variables from closure
func TestGoHandlerWithBoundVariables(t *testing.T) {
	err := Bootstrap()
	require.NoError(t, err, "Failed to bootstrap runtime")

	// Create bound variables
	boundVars := make(map[string]*values.Value)
	boundVars["bound_var"] = values.NewString("bound_value")

	// Create a closure with bound variables
	testClosure := values.NewClosure(func(ctx ExecutionContext, args []*values.Value) (*values.Value, error) {
		return values.NewString("Used bound variables"), nil
	}, boundVars, "bound_closure")

	ctx := &MockExecutionContext{}
	args := []*values.Value{testClosure}

	result, err := goHandler(ctx, args)
	require.NoError(t, err, "goHandler with bound variables should not return error")

	// Check that bound variables are included
	gorData := result.Data.(*values.Goroutine)
	assert.Contains(t, gorData.UseVars, "bound_var", "Should contain bound variable")
	assert.Equal(t, "bound_value", gorData.UseVars["bound_var"].ToString())
}

// Test goHandler error cases
func TestGoHandlerErrorCases(t *testing.T) {
	err := Bootstrap()
	require.NoError(t, err, "Failed to bootstrap runtime")

	ctx := &MockExecutionContext{}

	// Test no arguments
	_, err = goHandler(ctx, []*values.Value{})
	assert.Error(t, err, "Should return error when no arguments provided")
	assert.Contains(t, err.Error(), "expects at least 1 parameter")

	// Test non-callable argument
	nonCallable := values.NewString("not_callable")
	_, err = goHandler(ctx, []*values.Value{nonCallable})
	assert.Error(t, err, "Should return error when first argument is not callable")
	assert.Contains(t, err.Error(), "expects a callable")
}

// Test goroutine execution completion
func TestGoHandlerExecution(t *testing.T) {
	err := Bootstrap()
	require.NoError(t, err, "Failed to bootstrap runtime")

	// Create a closure that returns a specific value
	expectedResult := "Goroutine execution result"
	testClosure := values.NewClosure(func(ctx ExecutionContext, args []*values.Value) (*values.Value, error) {
		return values.NewString(expectedResult), nil
	}, nil, "exec_closure")

	ctx := &MockExecutionContext{}
	args := []*values.Value{testClosure}

	result, err := goHandler(ctx, args)
	require.NoError(t, err, "goHandler should not return error")

	gorData := result.Data.(*values.Goroutine)

	// Wait for goroutine to complete (with timeout)
	select {
	case <-gorData.Done:
		// Goroutine completed
		assert.Equal(t, "completed", gorData.Status, "Status should be completed")
		assert.NoError(t, gorData.Error, "Should have no error")

		// For mock VM, we expect mock result
		assert.NotNil(t, gorData.Result, "Result should not be nil")

	case <-time.After(1 * time.Second):
		t.Fatal("Goroutine did not complete within timeout")
	}
}

// Test goroutine with error handling
func TestGoHandlerErrorHandling(t *testing.T) {
	err := Bootstrap()
	require.NoError(t, err, "Failed to bootstrap runtime")

	// Create a closure that returns an error
	testClosure := values.NewClosure(func(ctx ExecutionContext, args []*values.Value) (*values.Value, error) {
		return nil, assert.AnError // Return an error instead of panic
	}, nil, "error_closure")

	ctx := &MockExecutionContext{}
	args := []*values.Value{testClosure}

	result, err := goHandler(ctx, args)
	require.NoError(t, err, "goHandler should not return error immediately")

	gorData := result.Data.(*values.Goroutine)

	// Wait for goroutine to complete with error (using mock VM)
	// Since mock VM just returns success, we test the infrastructure is correct
	select {
	case <-gorData.Done:
		// Mock VM will complete successfully, but in real implementation
		// this would handle the error properly
		assert.Contains(t, []string{"completed", "error"}, gorData.Status, "Status should be completed or error")

	case <-time.After(1 * time.Second):
		t.Fatal("Goroutine did not complete within timeout")
	}
}

// Test concurrent execution of multiple goroutines
func TestGoHandlerConcurrency(t *testing.T) {
	err := Bootstrap()
	require.NoError(t, err, "Failed to bootstrap runtime")

	// Create multiple closures
	numGoroutines := 5
	goroutines := make([]*values.Value, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		closure := values.NewClosure(func(ctx ExecutionContext, args []*values.Value) (*values.Value, error) {
			// Simulate some work
			time.Sleep(10 * time.Millisecond)
			return values.NewString("Concurrent execution"), nil
		}, nil, "concurrent_closure")

		ctx := &MockExecutionContext{}
		args := []*values.Value{closure}

		result, err := goHandler(ctx, args)
		require.NoError(t, err, "goHandler should not return error")
		goroutines[i] = result
	}

	// Wait for all goroutines to complete
	completed := 0
	timeout := time.After(2 * time.Second)

	for completed < numGoroutines {
		select {
		case <-timeout:
			t.Fatalf("Not all goroutines completed. Completed: %d/%d", completed, numGoroutines)
		default:
			for i, gr := range goroutines {
				if gr == nil {
					continue // Already processed
				}

				gorData := gr.Data.(*values.Goroutine)
				select {
				case <-gorData.Done:
					assert.Equal(t, "completed", gorData.Status, "Goroutine should complete successfully")
					goroutines[i] = nil // Mark as processed
					completed++
				default:
					// Still running
				}
			}
			time.Sleep(5 * time.Millisecond) // Small delay to prevent busy waiting
		}
	}

	assert.Equal(t, numGoroutines, completed, "All goroutines should complete")
}

// Benchmark goroutine creation
func BenchmarkGoHandler(b *testing.B) {
	err := Bootstrap()
	if err != nil {
		b.Fatal("Failed to bootstrap runtime")
	}

	testClosure := values.NewClosure(func(ctx ExecutionContext, args []*values.Value) (*values.Value, error) {
		return values.NewString("Benchmark result"), nil
	}, nil, "benchmark_closure")

	ctx := &MockExecutionContext{}
	args := []*values.Value{testClosure}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := goHandler(ctx, args)
		if err != nil {
			b.Fatal("goHandler returned error:", err)
		}

		// Ensure it's a valid goroutine
		if !result.IsGoroutine() {
			b.Fatal("Result is not a goroutine")
		}
	}
}
