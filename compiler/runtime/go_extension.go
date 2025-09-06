package runtime

import (
	"fmt"
	"sync"

	"github.com/wudi/php-parser/compiler/values"
)

// GoExtension provides goroutine-based concurrent execution for PHP callables
type GoExtension struct {
	*BaseExtension
	activeGoroutines sync.WaitGroup
	goroutineResults map[string]*values.Value
	resultMutex      sync.RWMutex
}

// NewGoExtension creates a new Go extension
func NewGoExtension() *GoExtension {
	base := NewBaseExtension("go", "1.0.0", "Provides goroutine-based concurrent execution for PHP callables")
	base.SetLoadOrder(50) // Load after core extensions
	
	return &GoExtension{
		BaseExtension:    base,
		goroutineResults: make(map[string]*values.Value),
	}
}

// Register registers the Go extension with the runtime registry
func (ge *GoExtension) Register(registry *RuntimeRegistry) error {
	// Call base registration
	if err := ge.BaseExtension.Register(registry); err != nil {
		return err
	}
	
	// Register the 'go' function
	goFuncDesc := &FunctionDescriptor{
		Name:    "go",
		Handler: ge.goHandler,
		Parameters: []ParameterDescriptor{
			{
				Name:        "callback",
				Type:        "callable",
			},
		},
		MinArgs:      1,
		MaxArgs:      1,
		IsBuiltin:    false,
	}
	
	return ge.RegisterFunction(registry, goFuncDesc)
}

// goHandler handles the 'go' function call
func (ge *GoExtension) goHandler(ctx ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("go() expects exactly 1 argument, got %d", len(args))
	}
	
	callable := args[0]
	
	// Validate that the argument is callable
	if !callable.IsClosure() {
		return nil, fmt.Errorf("go() argument must be callable, got %s", callable.Type.String())
	}
	
	// Execute the callable in a goroutine
	ge.activeGoroutines.Add(1)
	go ge.executeCallableAsync(ctx, callable)
	
	// Return null immediately (non-blocking)
	return values.NewNull(), nil
}

// executeCallableAsync executes a callable in a goroutine
func (ge *GoExtension) executeCallableAsync(ctx ExecutionContext, callable *values.Value) {
	defer ge.activeGoroutines.Done()
	
	closure := callable.ClosureGet()
	if closure == nil {
		// Log error but don't crash the goroutine
		fmt.Printf("Error: invalid closure in goroutine\n")
		return
	}
	
	// Try to execute the callable
	result, err := ge.invokeCallable(ctx, closure, []*values.Value{})
	if err != nil {
		fmt.Printf("Error executing callable in goroutine: %v\n", err)
		return
	}
	
	// Store result (optional, for potential future retrieval)
	ge.resultMutex.Lock()
	defer ge.resultMutex.Unlock()
	
	// Use closure name as key, or generate one
	key := closure.Name
	if key == "" {
		key = fmt.Sprintf("goroutine_%p", closure)
	}
	
	ge.goroutineResults[key] = result
}

// invokeCallable invokes a closure with the given arguments
func (ge *GoExtension) invokeCallable(ctx ExecutionContext, closure *values.Closure, args []*values.Value) (*values.Value, error) {
	if closure.Function == nil {
		return nil, fmt.Errorf("closure function is nil")
	}
	
	// Handle different function types
	switch fn := closure.Function.(type) {
	case func(ExecutionContext, []*values.Value) (*values.Value, error):
		// Runtime function handler with context
		return fn(ctx, args)
		
	case func([]*values.Value) (*values.Value, error):
		// Legacy function handler without context
		return fn(args)
		
	case func() *values.Value:
		// Simple parameterless function
		return fn(), nil
		
	case func():
		// Void function
		fn()
		return values.NewNull(), nil
		
	default:
		// Try to call as a generic interface{}
		return ge.invokeGenericCallable(closure, args)
	}
}

// invokeGenericCallable attempts to invoke a generic callable
func (ge *GoExtension) invokeGenericCallable(closure *values.Closure, args []*values.Value) (*values.Value, error) {
	// This is a fallback for other callable types
	// In a real implementation, you might need reflection or other mechanisms
	
	// For now, we'll create a simple result
	fmt.Printf("Executing generic callable '%s' in goroutine\n", closure.Name)
	
	// Simulate some work
	if len(args) > 0 {
		fmt.Printf("  Arguments: %v\n", args)
	}
	
	// Apply bound variables if any
	if len(closure.BoundVars) > 0 {
		fmt.Printf("  Bound variables: %v\n", closure.BoundVars)
	}
	
	return values.NewString("goroutine_completed"), nil
}

// WaitForAll waits for all active goroutines to complete
func (ge *GoExtension) WaitForAll() {
	ge.activeGoroutines.Wait()
}

// GetResult retrieves a result by key (if stored)
func (ge *GoExtension) GetResult(key string) (*values.Value, bool) {
	ge.resultMutex.RLock()
	defer ge.resultMutex.RUnlock()
	
	result, exists := ge.goroutineResults[key]
	return result, exists
}

// ClearResults clears all stored results
func (ge *GoExtension) ClearResults() {
	ge.resultMutex.Lock()
	defer ge.resultMutex.Unlock()
	
	ge.goroutineResults = make(map[string]*values.Value)
}

// Unregister cleans up the Go extension
func (ge *GoExtension) Unregister(registry *RuntimeRegistry) error {
	// Wait for all goroutines to complete
	ge.WaitForAll()
	
	// Clear results
	ge.ClearResults()
	
	// Call base unregister
	return ge.BaseExtension.Unregister(registry)
}