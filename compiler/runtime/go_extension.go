package runtime

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/wudi/php-parser/compiler/values"
)

// ClosureExecutor interface for executing closures (to avoid import cycles)
type ClosureExecutor interface {
	ExecuteClosure(ctx interface{}, closure *values.Closure, args []*values.Value) (*values.Value, error)
}

// GoExtension provides goroutine-based concurrent execution for PHP callables
type GoExtension struct {
	*BaseExtension
	closureExecutor  ClosureExecutor
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

// SetClosureExecutor sets the closure executor (VM) for proper closure handling
func (ge *GoExtension) SetClosureExecutor(executor ClosureExecutor) {
	ge.closureExecutor = executor
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

// invokeCallable invokes a closure with the given arguments using VM execution when possible
func (ge *GoExtension) invokeCallable(ctx ExecutionContext, closure *values.Closure, args []*values.Value) (*values.Value, error) {
	if closure.Function == nil {
		return nil, fmt.Errorf("closure function is nil")
	}
	
	// Try VM execution first for proper closure handling
	if ge.closureExecutor != nil {
		// Use VM's closure execution
		result, err := ge.closureExecutor.ExecuteClosure(ctx, closure, args)
		if err == nil {
			return result, nil
		}
		
		// If VM execution fails for type mismatch, fall through to other handlers
		fmt.Printf("VM execution failed, falling back to direct handlers: %v\n", err)
	}
	
	// Fallback to direct function type handling
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
		// Try reflection as final fallback
		return ge.invokeGenericCallable(closure, args)
	}
}

// invokeGenericCallable attempts to invoke a generic callable using reflection
func (ge *GoExtension) invokeGenericCallable(closure *values.Closure, args []*values.Value) (*values.Value, error) {
	if closure.Function == nil {
		return nil, fmt.Errorf("closure function is nil")
	}
	
	// Use reflection to analyze the function
	fnValue := reflect.ValueOf(closure.Function)
	fnType := fnValue.Type()
	
	// Ensure it's a function
	if fnType.Kind() != reflect.Func {
		return ge.handleNonFunction(closure, args)
	}
	
	fmt.Printf("Executing generic callable '%s' with reflection\n", closure.Name)
	
	// Analyze function signature
	numIn := fnType.NumIn()
	numOut := fnType.NumOut()
	
	fmt.Printf("  Function signature: %d inputs, %d outputs\n", numIn, numOut)
	
	// Prepare arguments for the call
	callArgs, err := ge.prepareReflectionArgs(fnType, args)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare arguments: %v", err)
	}
	
	// Make the reflection call
	var results []reflect.Value
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("  Panic during reflection call: %v\n", r)
		}
	}()
	
	results = fnValue.Call(callArgs)
	
	// Process results
	return ge.processReflectionResults(results, numOut)
}

// prepareReflectionArgs prepares arguments for reflection call
func (ge *GoExtension) prepareReflectionArgs(fnType reflect.Type, args []*values.Value) ([]reflect.Value, error) {
	numIn := fnType.NumIn()
	callArgs := make([]reflect.Value, 0, numIn)
	
	for i := 0; i < numIn; i++ {
		paramType := fnType.In(i)
		
		// Handle different parameter types
		switch {
		case paramType == reflect.TypeOf((*ExecutionContext)(nil)).Elem():
			// ExecutionContext interface - pass nil for now
			callArgs = append(callArgs, reflect.Zero(paramType))
			
		case paramType == reflect.TypeOf([]*values.Value{}):
			// []*values.Value slice
			callArgs = append(callArgs, reflect.ValueOf(args))
			
		case paramType == reflect.TypeOf(&values.Value{}):
			// Individual *values.Value
			if i-1 < len(args) { // Account for potential context parameter
				callArgs = append(callArgs, reflect.ValueOf(args[i-1]))
			} else {
				callArgs = append(callArgs, reflect.ValueOf(values.NewNull()))
			}
			
		case paramType.Kind() == reflect.String:
			// String parameter
			if len(args) > 0 {
				callArgs = append(callArgs, reflect.ValueOf(args[0].ToString()))
			} else {
				callArgs = append(callArgs, reflect.ValueOf(""))
			}
			
		case paramType.Kind() == reflect.Int64:
			// Int64 parameter
			if len(args) > 0 {
				callArgs = append(callArgs, reflect.ValueOf(args[0].ToInt()))
			} else {
				callArgs = append(callArgs, reflect.ValueOf(int64(0)))
			}
			
		case paramType.Kind() == reflect.Float64:
			// Float64 parameter
			if len(args) > 0 {
				callArgs = append(callArgs, reflect.ValueOf(args[0].ToFloat()))
			} else {
				callArgs = append(callArgs, reflect.ValueOf(float64(0)))
			}
			
		case paramType.Kind() == reflect.Bool:
			// Bool parameter
			if len(args) > 0 {
				callArgs = append(callArgs, reflect.ValueOf(args[0].ToBool()))
			} else {
				callArgs = append(callArgs, reflect.ValueOf(false))
			}
			
		default:
			// Try to pass zero value for unknown types
			callArgs = append(callArgs, reflect.Zero(paramType))
		}
	}
	
	return callArgs, nil
}

// processReflectionResults processes the results from reflection call
func (ge *GoExtension) processReflectionResults(results []reflect.Value, numOut int) (*values.Value, error) {
	switch numOut {
	case 0:
		// No return value
		fmt.Printf("  Function completed with no return value\n")
		return values.NewNull(), nil
		
	case 1:
		// Single return value
		result := results[0]
		return ge.convertReflectionValueToValue(result)
		
	case 2:
		// Two return values - typically (result, error) pattern
		resultValue := results[0]
		errorValue := results[1]
		
		// Check if second value is an error
		if errorValue.Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			if !errorValue.IsNil() {
				return nil, fmt.Errorf("function returned error: %v", errorValue.Interface())
			}
		}
		
		return ge.convertReflectionValueToValue(resultValue)
		
	default:
		// Multiple return values - return as array
		fmt.Printf("  Function returned %d values, wrapping in array\n", numOut)
		resultArray := values.NewArray()
		
		for i, result := range results {
			convertedValue, err := ge.convertReflectionValueToValue(result)
			if err != nil {
				fmt.Printf("  Warning: failed to convert result %d: %v\n", i, err)
				convertedValue = values.NewString(fmt.Sprintf("unconvertible_result_%d", i))
			}
			resultArray.ArraySet(values.NewInt(int64(i)), convertedValue)
		}
		
		return resultArray, nil
	}
}

// convertReflectionValueToValue converts a reflection value to a PHP Value
func (ge *GoExtension) convertReflectionValueToValue(reflectVal reflect.Value) (*values.Value, error) {
	if !reflectVal.IsValid() {
		return values.NewNull(), nil
	}
	
	// Handle different types
	switch reflectVal.Kind() {
	case reflect.Bool:
		return values.NewBool(reflectVal.Bool()), nil
		
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return values.NewInt(reflectVal.Int()), nil
		
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return values.NewInt(int64(reflectVal.Uint())), nil
		
	case reflect.Float32, reflect.Float64:
		return values.NewFloat(reflectVal.Float()), nil
		
	case reflect.String:
		return values.NewString(reflectVal.String()), nil
		
	case reflect.Ptr:
		// Handle pointer types
		if reflectVal.IsNil() {
			return values.NewNull(), nil
		}
		
		// Check if it's a *values.Value
		if reflectVal.Type() == reflect.TypeOf(&values.Value{}) {
			return reflectVal.Interface().(*values.Value), nil
		}
		
		// Dereference and try again
		return ge.convertReflectionValueToValue(reflectVal.Elem())
		
	case reflect.Interface:
		// Handle interface types
		if reflectVal.IsNil() {
			return values.NewNull(), nil
		}
		
		// Get the concrete value
		return ge.convertReflectionValueToValue(reflectVal.Elem())
		
	case reflect.Slice, reflect.Array:
		// Convert to PHP array
		resultArray := values.NewArray()
		length := reflectVal.Len()
		
		for i := 0; i < length; i++ {
			element := reflectVal.Index(i)
			convertedElement, err := ge.convertReflectionValueToValue(element)
			if err != nil {
				convertedElement = values.NewString(fmt.Sprintf("unconvertible_element_%d", i))
			}
			resultArray.ArraySet(values.NewInt(int64(i)), convertedElement)
		}
		
		return resultArray, nil
		
	case reflect.Map:
		// Convert to PHP array
		resultArray := values.NewArray()
		keys := reflectVal.MapKeys()
		
		for _, key := range keys {
			value := reflectVal.MapIndex(key)
			
			convertedKey, err := ge.convertReflectionValueToValue(key)
			if err != nil {
				continue // Skip unconvertible keys
			}
			
			convertedValue, err := ge.convertReflectionValueToValue(value)
			if err != nil {
				convertedValue = values.NewString("unconvertible_value")
			}
			
			resultArray.ArraySet(convertedKey, convertedValue)
		}
		
		return resultArray, nil
		
	default:
		// For unknown types, convert to string representation
		return values.NewString(fmt.Sprintf("%v", reflectVal.Interface())), nil
	}
}

// handleNonFunction handles non-function callables
func (ge *GoExtension) handleNonFunction(closure *values.Closure, args []*values.Value) (*values.Value, error) {
	fmt.Printf("Handling non-function callable '%s'\n", closure.Name)
	
	// Try to handle as a callable object or other special types
	closureValue := reflect.ValueOf(closure.Function)
	
	switch closureValue.Kind() {
	case reflect.String:
		// String-based callable (function name)
		funcName := closureValue.String()
		fmt.Printf("  String callable: %s\n", funcName)
		
		// This could be used to look up functions by name
		// For now, just return a placeholder result
		return values.NewString(fmt.Sprintf("called_%s", funcName)), nil
		
	case reflect.Map:
		// Map-based callable (could represent a callable array in PHP)
		fmt.Printf("  Map callable with %d entries\n", closureValue.Len())
		return values.NewString("map_callable_result"), nil
		
	default:
		// Unknown type - return string representation
		return values.NewString(fmt.Sprintf("unknown_callable_%s", closureValue.Type().String())), nil
	}
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