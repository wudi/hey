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

// GetConcurrencyClasses returns WaitGroup class
func GetConcurrencyClasses() []*registry.ClassDescriptor {
	return []*registry.ClassDescriptor{
		{
			Name:       "WaitGroup",
			Parent:     "",
			Interfaces: []string{},
			Traits:     []string{},
			Methods: map[string]*registry.MethodDescriptor{
				"__construct": {
					Name:           "__construct",
					Visibility:     "public",
					IsStatic:       false,
					IsAbstract:     false,
					IsFinal:        false,
					IsVariadic:     false,
					Parameters:     []*registry.ParameterDescriptor{},
					Implementation: NewBuiltinMethodImpl(createWaitGroupConstructor()),
				},
				"Add": {
					Name:       "Add",
					Visibility: "public",
					IsStatic:   false,
					IsAbstract: false,
					IsFinal:    false,
					IsVariadic: false,
					Parameters: []*registry.ParameterDescriptor{
						{Name: "delta", Type: "int"},
					},
					Implementation: NewBuiltinMethodImpl(createWaitGroupAddMethod()),
				},
				"Done": {
					Name:           "Done",
					Visibility:     "public",
					IsStatic:       false,
					IsAbstract:     false,
					IsFinal:        false,
					IsVariadic:     false,
					Parameters:     []*registry.ParameterDescriptor{},
					Implementation: NewBuiltinMethodImpl(createWaitGroupDoneMethod()),
				},
				"Wait": {
					Name:           "Wait",
					Visibility:     "public",
					IsStatic:       false,
					IsAbstract:     false,
					IsFinal:        false,
					IsVariadic:     false,
					Parameters:     []*registry.ParameterDescriptor{},
					Implementation: NewBuiltinMethodImpl(createWaitGroupWaitMethod()),
				},
			},
			Properties: make(map[string]*registry.PropertyDescriptor),
			Constants:  make(map[string]*registry.ConstantDescriptor),
			IsAbstract: false,
			IsFinal:    false,
		},
	}
}

// Helper functions to create WaitGroup method implementations
func createWaitGroupConstructor() *registry.Function {
	return &registry.Function{
		Name:       "__construct",
		Parameters: []*registry.Parameter{},
		ReturnType: "void",
		IsBuiltin:  true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			return values.NewNull(), nil
		},
	}
}

func createWaitGroupAddMethod() *registry.Function {
	return &registry.Function{
		Name:       "Add",
		Parameters: []*registry.Parameter{{Name: "delta", Type: "int"}},
		ReturnType: "void",
		IsBuiltin:  true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			// This would need proper method context handling
			// For now, this is a placeholder implementation
			return values.NewNull(), nil
		},
	}
}

func createWaitGroupDoneMethod() *registry.Function {
	return &registry.Function{
		Name:       "Done",
		Parameters: []*registry.Parameter{},
		ReturnType: "void",
		IsBuiltin:  true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			// This would need proper method context handling
			// For now, this is a placeholder implementation
			return values.NewNull(), nil
		},
	}
}

func createWaitGroupWaitMethod() *registry.Function {
	return &registry.Function{
		Name:       "Wait",
		Parameters: []*registry.Parameter{},
		ReturnType: "void",
		IsBuiltin:  true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			// This would need proper method context handling
			// For now, this is a placeholder implementation
			return values.NewNull(), nil
		},
	}
}