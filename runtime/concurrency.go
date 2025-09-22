package runtime

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// copyValue creates a deep copy of a value to avoid race conditions
func copyValue(val *values.Value) *values.Value {
	if val == nil {
		return nil
	}

	// For basic types, create a new value with the same data
	switch val.Type {
	case values.TypeNull:
		return values.NewNull()
	case values.TypeBool:
		return values.NewBool(val.ToBool())
	case values.TypeInt:
		return values.NewInt(val.ToInt())
	case values.TypeFloat:
		return values.NewFloat(val.ToFloat())
	case values.TypeString:
		return values.NewString(val.ToString())
	default:
		// For complex types (objects, arrays, etc.), return the same reference
		// This is a simplified copy - full deep copy would be more complex
		return val
	}
}

// GoroutineExecutor is an interface for executing functions in goroutines
type GoroutineExecutor interface {
	ExecuteFunction(fn *registry.Function, boundVars map[string]*values.Value) (*values.Value, error)
}

// Global executor that will be set by the VM
var globalGoroutineExecutor GoroutineExecutor

// SetGoroutineExecutor sets the global executor for goroutine bytecode execution
func SetGoroutineExecutor(executor GoroutineExecutor) {
	globalGoroutineExecutor = executor
}

// Global goroutine manager and context tracking
var (
	goroutineManager = &GoroutineManager{
		running: make(map[int64]*values.Value),
		mu:      sync.RWMutex{},
	}

	// Context key for current goroutine ID
	currentGoroutineKey = &contextKey{"current_goroutine_id"}

	// Map to track current goroutine ID for each Go goroutine
	currentGoroutineIDs = make(map[int64]int64)
	currentGoroutineIDsMu sync.RWMutex
)

// contextKey is used for goroutine context values
type contextKey struct {
	name string
}

func (k *contextKey) String() string {
	return "concurrency context key " + k.name
}

// setCurrentGoroutineContext stores the current goroutine ID for this Go goroutine
func setCurrentGoroutineContext(ctx context.Context) {
	goroutineID := ctx.Value(currentGoroutineKey).(int64)
	goRoutineID := getGoRoutineID()

	currentGoroutineIDsMu.Lock()
	currentGoroutineIDs[goRoutineID] = goroutineID
	currentGoroutineIDsMu.Unlock()
}

// getCurrentGoroutineID returns the current PHP goroutine ID or error if not in a goroutine
func getCurrentGoroutineID() (int64, error) {
	goRoutineID := getGoRoutineID()

	currentGoroutineIDsMu.RLock()
	goroutineID, exists := currentGoroutineIDs[goRoutineID]
	currentGoroutineIDsMu.RUnlock()

	if !exists {
		return 0, fmt.Errorf("goid() can only be called from within a goroutine started by go()")
	}

	return goroutineID, nil
}

// getGoRoutineID returns the current Go goroutine ID using runtime
func getGoRoutineID() int64 {
	// Use a simple approach: convert current goroutine to a unique ID
	// For better implementation, you could parse runtime.Stack() or use other methods
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, _ := strconv.ParseInt(idField, 10, 64)
	return id
}

// GoroutineManager tracks running goroutines
type GoroutineManager struct {
	running map[int64]*values.Value
	mu      sync.RWMutex
}

// ExecuteGoroutine runs a goroutine with basic execution
func (gm *GoroutineManager) ExecuteGoroutine(gor *values.Value) {
	goroutineData := gor.Data.(*values.Goroutine)

	gm.mu.Lock()
	gm.running[goroutineData.ID] = gor
	gm.mu.Unlock()

	go func() {
		goRoutineID := getGoRoutineID()
		defer func() {
			// Clean up goroutine tracking
			gm.mu.Lock()
			delete(gm.running, goroutineData.ID)
			gm.mu.Unlock()

			// Clean up goroutine ID mapping
			currentGoroutineIDsMu.Lock()
			delete(currentGoroutineIDs, goRoutineID)
			currentGoroutineIDsMu.Unlock()

			if r := recover(); r != nil {
				goroutineData.Status = "error"
				goroutineData.Error = fmt.Errorf("goroutine panic: %v", r)
			}
			close(goroutineData.Done)
		}()

		// Set current goroutine ID in context
		ctx := context.WithValue(context.Background(), currentGoroutineKey, goroutineData.ID)

		// Store context in a goroutine-local way for goid() function access
		setCurrentGoroutineContext(ctx)

		// Execute the closure
		closure := goroutineData.Function

		// Handle different types of closures
		if fn, ok := closure.Function.(*registry.Function); ok {
			if fn.IsBuiltin && fn.Builtin != nil {
				// Built-in function
				result, err := fn.Builtin(nil, []*values.Value{})
				if err != nil {
					goroutineData.Status = "error"
					goroutineData.Error = err
				} else {
					goroutineData.Status = "completed"
					goroutineData.Result = result
				}
			} else if fn.Instructions != nil {
				// User-defined function with bytecode - need VM execution
				err := gm.executeUserFunction(fn, goroutineData)
				if err != nil {
					goroutineData.Status = "error"
					goroutineData.Error = err
				} else {
					goroutineData.Status = "completed"
				}
			} else {
				goroutineData.Status = "error"
				goroutineData.Error = fmt.Errorf("function has no implementation")
			}
		} else {
			// Unknown closure type
			goroutineData.Status = "error"
			goroutineData.Error = fmt.Errorf("unsupported closure type: %T", closure.Function)
		}
	}()
}

// executeUserFunction executes a user-defined PHP function in a goroutine context
func (gm *GoroutineManager) executeUserFunction(fn *registry.Function, goroutineData *values.Goroutine) error {
	if globalGoroutineExecutor == nil {
		return fmt.Errorf("no goroutine executor registered - VM integration not available")
	}

	// Prepare bound variables map from closure
	boundVars := make(map[string]*values.Value)

	// Add bound variables from closure (these are the 'use' variables)
	// Make deep copies to avoid race conditions between goroutines
	if goroutineData.Function.BoundVars != nil {
		for varName, boundVar := range goroutineData.Function.BoundVars {
			// Create a deep copy for this goroutine
			boundVars[varName] = copyValue(boundVar)
		}
	}

	// Add use variables from goroutine (additional variables)
	if goroutineData.UseVars != nil {
		for varName, useVar := range goroutineData.UseVars {
			// Create a deep copy for this goroutine
			boundVars[varName] = copyValue(useVar)
		}
	}

	// Execute the function using the global executor
	result, err := globalGoroutineExecutor.ExecuteFunction(fn, boundVars)
	if err != nil {
		return err
	}

	// Store the result
	goroutineData.Result = result
	return nil
}

// GetConcurrencyFunctions returns concurrency-related PHP functions
func GetConcurrencyFunctions() []*registry.Function {
	return []*registry.Function{
		{
			Name:       "go",
			Parameters: []*registry.Parameter{{Name: "closure", Type: "callable"}},
			ReturnType: "Goroutine",
			IsVariadic: true,
			MinArgs:    1,
			MaxArgs:    -1,
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return nil, fmt.Errorf("go() expects at least one argument")
				}
				closureVal := args[0]
				if closureVal == nil || !closureVal.IsCallable() {
					return nil, fmt.Errorf("go() expects a callable as first argument")
				}

				// Check if closure data is valid before calling ClosureGet()
				if closureVal.Data == nil {
					return nil, fmt.Errorf("go() invalid closure")
				}

				closure := closureVal.ClosureGet()
				if closure == nil {
					return nil, fmt.Errorf("go() invalid closure")
				}

				// Create a copy of the closure for this goroutine to avoid shared state
				closureCopy := &values.Closure{
					Function:  closure.Function, // Function can be shared
					BoundVars: make(map[string]*values.Value), // But BoundVars must be separate
					Name:      closure.Name,
				}

				// Copy the bound variables to the new closure
				if closure.BoundVars != nil {
					for k, v := range closure.BoundVars {
						closureCopy.BoundVars[k] = copyValue(v)
					}
				}

				useVars := make(map[string]*values.Value)
				for i, arg := range args[1:] {
					useVars[fmt.Sprintf("var_%d", i)] = arg
				}

				// Create goroutine value with the isolated closure copy
				gor := values.NewGoroutine(closureCopy, useVars)

				// Execute immediately in real Go goroutine
				goroutineManager.ExecuteGoroutine(gor)

				return gor, nil
			},
		},
		{
			Name:       "goid",
			Parameters: []*registry.Parameter{},
			ReturnType: "int",
			IsVariadic: false,
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) != 0 {
					return nil, fmt.Errorf("goid() expects no arguments")
				}

				// Get current goroutine ID
				goroutineID, err := getCurrentGoroutineID()
				if err != nil {
					return nil, err
				}

				return values.NewInt(goroutineID), nil
			},
		},
	}
}

// GetConcurrencyClasses returns concurrency-related classes
// Note: WaitGroup class is now defined in exception.go
func GetConcurrencyClasses() []*registry.ClassDescriptor {
	return []*registry.ClassDescriptor{
		// Add other concurrency-related classes here if needed
	}
}