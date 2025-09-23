package vm

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestExecutionContextConcurrentMapAccess(t *testing.T) {
	ctx := NewExecutionContext()

	// Test concurrent access to GlobalVars
	t.Run("ConcurrentGlobalVars", func(t *testing.T) {
		var wg sync.WaitGroup
		numGoroutines := 50
		numOperations := 100

		// Start multiple goroutines that access GlobalVars concurrently
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < numOperations; j++ {
					key := fmt.Sprintf("var_%d_%d", id, j)
					value := values.NewString(fmt.Sprintf("value_%d_%d", id, j))

					// Set global variable
					ctx.SetGlobalVar(key, value)

					// Get global variable
					if val, ok := ctx.GetGlobalVar(key); ok {
						assert.Equal(t, value.ToString(), val.ToString())
					}
				}
			}(i)
		}

		wg.Wait()
	})

	// Test concurrent access to Variables
	t.Run("ConcurrentVariables", func(t *testing.T) {
		var wg sync.WaitGroup
		numGoroutines := 50
		numOperations := 100

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < numOperations; j++ {
					key := fmt.Sprintf("var_%d_%d", id, j)
					value := values.NewString(fmt.Sprintf("value_%d_%d", id, j))

					ctx.setVariable(key, value)

					if val, ok := ctx.GetVariable(key); ok {
						assert.Equal(t, value.ToString(), val.ToString())
					}
				}
			}(i)
		}

		wg.Wait()
	})

	// Test concurrent access to Temporaries
	t.Run("ConcurrentTemporaries", func(t *testing.T) {
		var wg sync.WaitGroup
		numGoroutines := 50
		numOperations := 100

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < numOperations; j++ {
					slot := uint32(id*1000 + j) // Unique slot per goroutine
					value := values.NewString(fmt.Sprintf("temp_%d_%d", id, j))

					ctx.setTemporary(slot, value)

					if val, ok := ctx.GetTemporary(slot); ok {
						assert.Equal(t, value.ToString(), val.ToString())
					}
				}
			}(i)
		}

		wg.Wait()
	})

	// Test concurrent access to ClassTable
	t.Run("ConcurrentClassTable", func(t *testing.T) {
		var wg sync.WaitGroup
		numGoroutines := 50

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				className := fmt.Sprintf("TestClass_%d", id)

				// This will internally use the ClassTable
				cls := ctx.ensureClass(className)
				assert.NotNil(t, cls)
				assert.Equal(t, className, cls.Name)

				// Try to get the same class again
				cls2, ok := ctx.getClass(className)
				assert.True(t, ok)
				assert.Equal(t, cls.Name, cls2.Name)
			}(i)
		}

		wg.Wait()
	})

	// Test concurrent access to IncludedFiles
	t.Run("ConcurrentIncludedFiles", func(t *testing.T) {
		var wg sync.WaitGroup
		numGoroutines := 50

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				filename := fmt.Sprintf("test_file_%d.php", id)

				// Mark file as included
				ctx.MarkFileIncluded(filename)

				// Check if file is included
				assert.True(t, ctx.IsFileIncluded(filename))
			}(i)
		}

		wg.Wait()
	})
}

func TestVMGoroutineExecutorIsolation(t *testing.T) {
	// Initialize a VM with some global state
	vm := &VirtualMachine{
		lastContext: NewExecutionContext(),
		watchVars:   make(map[string]struct{}),
		profile:     newProfileState(),
		DebugMode:   false,
	}

	// Set up some global state in the main context
	vm.lastContext.SetGlobalVar("shared_var", values.NewString("original_value"))
	vm.lastContext.ensureClass("SharedClass")

	executor := &GoroutineExecutor{vm: vm}

	// Create a simple function that modifies global state
	fn := &registry.Function{
		Name:         "testFunc",
		Parameters:   []*registry.Parameter{},
		Instructions: nil, // Will use builtin
		IsBuiltin:    true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			// Try to modify the shared variable
			if ctx != nil {
				ctx.SetGlobal("shared_var", values.NewString("modified_value"))
			}
			return values.NewString("done"), nil
		},
	}

	var wg sync.WaitGroup
	numGoroutines := 10

	// Execute the function in multiple goroutines simultaneously
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			boundVars := map[string]*values.Value{
				"goroutine_id": values.NewInt(int64(id)),
			}

			result, err := executor.ExecuteFunction(fn, boundVars)
			require.NoError(t, err)
			assert.Equal(t, "done", result.ToString())
		}(i)
	}

	wg.Wait()

	// Verify that the original context's shared variable wasn't modified
	// due to proper isolation
	if val, ok := vm.lastContext.GetGlobalVar("shared_var"); ok {
		assert.Equal(t, "original_value", val.ToString(),
			"Original context should not be modified by goroutines")
	}
}

func TestExecutionContextFrameOperations(t *testing.T) {
	ctx := NewExecutionContext()

	// Test concurrent frame operations
	t.Run("ConcurrentFrameOperations", func(t *testing.T) {
		var wg sync.WaitGroup
		numGoroutines := 50

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				// Create a frame
				frame := newCallFrame(
					fmt.Sprintf("func_%d", id),
					&registry.Function{Name: fmt.Sprintf("func_%d", id)},
					nil,
					nil,
				)

				// Push frame
				ctx.pushFrame(frame)

				// Get current frame
				current := ctx.currentFrame()
				if current != nil {
					assert.Contains(t, current.FunctionName, fmt.Sprintf("func_%d", id))
				}

				// Pop frame
				popped := ctx.popFrame()
				if popped != nil {
					assert.Equal(t, frame.FunctionName, popped.FunctionName)
				}
			}(i)
		}

		wg.Wait()
	})
}

func TestExecutionContextDebugOperations(t *testing.T) {
	ctx := NewExecutionContext()

	// Test concurrent debug operations
	t.Run("ConcurrentDebugOperations", func(t *testing.T) {
		var wg sync.WaitGroup
		numGoroutines := 50
		numRecords := 100

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < numRecords; j++ {
					record := fmt.Sprintf("debug_record_%d_%d", id, j)
					ctx.appendDebugRecord(record)
				}
			}(i)
		}

		wg.Wait()

		// Drain debug records
		records := ctx.drainDebugRecords()

		// Should have all records from all goroutines
		expectedTotal := numGoroutines * numRecords
		assert.Equal(t, expectedTotal, len(records))
	})
}

// Benchmark concurrent map access
func BenchmarkConcurrentMapAccess(b *testing.B) {
	ctx := NewExecutionContext()

	b.Run("GlobalVarsAccess", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := fmt.Sprintf("key_%d", i)
				value := values.NewString(fmt.Sprintf("value_%d", i))
				ctx.SetGlobalVar(key, value)
				ctx.GetGlobalVar(key)
				i++
			}
		})
	})

	b.Run("VariablesAccess", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := fmt.Sprintf("var_%d", i)
				value := values.NewString(fmt.Sprintf("value_%d", i))
				ctx.setVariable(key, value)
				ctx.GetVariable(key)
				i++
			}
		})
	})
}

// Test that helps ensure no deadlocks occur
func TestNoDeadlocks(t *testing.T) {
	ctx := NewExecutionContext()

	// This test runs operations that could potentially deadlock
	// if the locking strategy is incorrect
	done := make(chan bool, 1)
	timeout := time.After(5 * time.Second)

	go func() {
		var wg sync.WaitGroup

		// Mix different types of operations
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				// Global vars
				ctx.SetGlobalVar(fmt.Sprintf("g%d", id), values.NewString("test"))
				ctx.GetGlobalVar(fmt.Sprintf("g%d", id))

				// Variables
				ctx.setVariable(fmt.Sprintf("v%d", id), values.NewString("test"))
				ctx.GetVariable(fmt.Sprintf("v%d", id))

				// Classes
				ctx.ensureClass(fmt.Sprintf("Class%d", id))
				ctx.getClass(fmt.Sprintf("Class%d", id))

				// Files
				ctx.MarkFileIncluded(fmt.Sprintf("file%d.php", id))
				ctx.IsFileIncluded(fmt.Sprintf("file%d.php", id))

				// Temporaries
				ctx.setTemporary(uint32(id), values.NewString("temp"))
				ctx.GetTemporary(uint32(id))
			}(i)
		}

		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		// Test completed successfully
	case <-timeout:
		t.Fatal("Test timed out - potential deadlock detected")
	}
}
