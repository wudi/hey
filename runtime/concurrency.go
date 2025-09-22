package runtime

import (
	"fmt"
	"sync"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// Global goroutine manager
var (
	goroutineManager = &GoroutineManager{
		running: make(map[int64]*values.Value),
		mu:      sync.RWMutex{},
	}
)

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
		defer func() {
			gm.mu.Lock()
			delete(gm.running, goroutineData.ID)
			gm.mu.Unlock()

			if r := recover(); r != nil {
				goroutineData.Status = "error"
				goroutineData.Error = fmt.Errorf("goroutine panic: %v", r)
			}
			close(goroutineData.Done)
		}()

		// Execute the closure
		closure := goroutineData.Function

		// For compiled closures, we need to execute them properly
		// This is a simplified version - for now just mark as completed
		if fn, ok := closure.Function.(*registry.Function); ok {
			if fn.Handler != nil {
				result, err := fn.Handler(nil, []*values.Value{})
				if err != nil {
					goroutineData.Status = "error"
					goroutineData.Error = err
				} else {
					goroutineData.Status = "completed"
					goroutineData.Result = result
				}
			} else if fn.Builtin != nil {
				result, err := fn.Builtin(nil, []*values.Value{})
				if err != nil {
					goroutineData.Status = "error"
					goroutineData.Error = err
				} else {
					goroutineData.Status = "completed"
					goroutineData.Result = result
				}
			}
		} else {
			// For bytecode functions, we need VM integration
			// For now, just mark as completed
			goroutineData.Status = "completed"
			goroutineData.Result = values.NewNull()
		}
	}()
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
				useVars := make(map[string]*values.Value)
				for i, arg := range args[1:] {
					useVars[fmt.Sprintf("var_%d", i)] = arg
				}

				// Create goroutine value
				gor := values.NewGoroutine(closure, useVars)

				// Execute immediately in real Go goroutine
				goroutineManager.ExecuteGoroutine(gor)

				return gor, nil
			},
		},
	}
}

// GetConcurrencyClasses returns WaitGroup class
func GetConcurrencyClasses() []*registry.ClassDescriptor {
	// TODO: Implement proper WaitGroup class with method handlers
	// For now, return empty slice - the WaitGroup value methods work directly
	return []*registry.ClassDescriptor{}
}