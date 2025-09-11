# Extension Development Guide

This document provides a step-by-step guide for developing extensions for the PHP parser's runtime system, with practical examples and best practices.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Extension Basics](#extension-basics)
3. [Step-by-Step Tutorial](#step-by-step-tutorial)
4. [Advanced Features](#advanced-features)
5. [Best Practices](#best-practices)
6. [Troubleshooting](#troubleshooting)

## Quick Start

### Minimal Extension Example

Here's the simplest possible extension:

```go
package main

import (
    "github.com/wudi/hey/compiler/runtime"
    "github.com/wudi/hey/compiler/values"
)

type HelloExtension struct {
    *runtime.BaseExtension
}

func NewHelloExtension() *HelloExtension {
    return &HelloExtension{
        BaseExtension: runtime.NewBaseExtension("hello", "1.0.0", "Hello World extension"),
    }
}

func (he *HelloExtension) Register(registry *runtime.RuntimeRegistry) error {
    if err := he.BaseExtension.Register(registry); err != nil {
        return err
    }
    
    // Add a simple function
    return he.RegisterFunction(registry, &runtime.FunctionDescriptor{
        Name:    "say_hello",
        Handler: sayHelloHandler,
        MinArgs: 0,
        MaxArgs: 0,
    })
}

func sayHelloHandler(ctx runtime.ExecutionContext, args []*values.Value) (*values.Value, error) {
    return values.NewString("Hello, World!"), nil
}

// Usage:
// runtime.Bootstrap()
// registry.RegisterExtension(NewHelloExtension())
// PHP code: echo say_hello(); // Outputs: Hello, World!
```

## Extension Basics

### Core Concepts

1. **BaseExtension**: Foundation class providing common functionality
2. **Extension Interface**: Contract that all extensions must implement
3. **Descriptors**: Metadata objects describing functions, classes, etc.
4. **Handlers**: Go functions that implement PHP functionality
5. **Registry**: Central system that manages all extensions

### Extension Lifecycle

```
1. Create Extension → 2. Implement Interface → 3. Register with Registry
       ↓                       ↓                        ↓
4. Bootstrap calls Register() → 5. Extension adds entities → 6. Available in PHP
```

## Step-by-Step Tutorial

### Step 1: Create Extension Structure

Create a new Go package for your extension:

```go
package myextension

import (
    "fmt"
    "math"
    
    "github.com/wudi/hey/compiler/runtime"
    "github.com/wudi/hey/compiler/values"
)

// MathPlusExtension provides additional mathematical functions
type MathPlusExtension struct {
    *runtime.BaseExtension
}

func NewMathPlusExtension() *MathPlusExtension {
    ext := &MathPlusExtension{
        BaseExtension: runtime.NewBaseExtension(
            "mathplus",           // Extension name
            "1.0.0",             // Version
            "Extended math functions", // Description
        ),
    }
    
    // Optional: Set dependencies and load order
    ext.SetDependencies([]string{"core"}) // Depends on core
    ext.SetLoadOrder(100)                 // Load after core (order 0)
    
    return ext
}
```

### Step 2: Implement Extension Interface

The `Register` method is where you define what your extension provides:

```go
func (mpe *MathPlusExtension) Register(registry *runtime.RuntimeRegistry) error {
    // Always call parent Register first
    if err := mpe.BaseExtension.Register(registry); err != nil {
        return err
    }
    
    // Register constants
    if err := mpe.registerConstants(registry); err != nil {
        return err
    }
    
    // Register functions
    if err := mpe.registerFunctions(registry); err != nil {
        return err
    }
    
    // Register classes (optional)
    if err := mpe.registerClasses(registry); err != nil {
        return err
    }
    
    return nil
}
```

### Step 3: Add Constants

Constants are simple key-value pairs:

```go
func (mpe *MathPlusExtension) registerConstants(registry *runtime.RuntimeRegistry) error {
    constants := map[string]*values.Value{
        // Mathematical constants
        "MATHPLUS_PI":       values.NewFloat(math.Pi),
        "MATHPLUS_E":        values.NewFloat(math.E),
        "MATHPLUS_PHI":      values.NewFloat(1.618033988749), // Golden ratio
        
        // Configuration constants
        "MATHPLUS_VERSION":  values.NewString("1.0.0"),
        "MATHPLUS_PRECISION": values.NewInt(15),
    }
    
    for name, value := range constants {
        if err := mpe.RegisterConstant(registry, name, value); err != nil {
            return fmt.Errorf("failed to register constant %s: %v", name, err)
        }
    }
    
    return nil
}
```

### Step 4: Add Functions

Functions require descriptors and handler implementations:

```go
func (mpe *MathPlusExtension) registerFunctions(registry *runtime.RuntimeRegistry) error {
    functions := []*runtime.FunctionDescriptor{
        {
            Name:    "deg2rad",
            Handler: mpe.deg2radHandler,
            Parameters: []runtime.ParameterDescriptor{
                {Name: "degrees", Type: "float"},
            },
            MinArgs: 1,
            MaxArgs: 1,
        },
        {
            Name:    "rad2deg", 
            Handler: mpe.rad2degHandler,
            Parameters: []runtime.ParameterDescriptor{
                {Name: "radians", Type: "float"},
            },
            MinArgs: 1,
            MaxArgs: 1,
        },
        {
            Name:    "factorial",
            Handler: mpe.factorialHandler,
            Parameters: []runtime.ParameterDescriptor{
                {Name: "n", Type: "int"},
            },
            MinArgs: 1,
            MaxArgs: 1,
        },
        {
            Name:       "statistics",
            Handler:    mpe.statisticsHandler,
            IsVariadic: true,
            MinArgs:    1,
            MaxArgs:    -1, // Unlimited
        },
    }
    
    for _, desc := range functions {
        if err := mpe.RegisterFunction(registry, desc); err != nil {
            return fmt.Errorf("failed to register function %s: %v", desc.Name, err)
        }
    }
    
    return nil
}
```

### Step 5: Implement Function Handlers

Each function needs a Go implementation:

```go
func (mpe *MathPlusExtension) deg2radHandler(ctx runtime.ExecutionContext, args []*values.Value) (*values.Value, error) {
    // Input validation
    if len(args) != 1 {
        return nil, fmt.Errorf("deg2rad() expects 1 parameter, %d given", len(args))
    }
    
    // Convert to float
    degrees := args[0].ToFloat()
    
    // Perform calculation
    radians := degrees * (math.Pi / 180.0)
    
    // Return result
    return values.NewFloat(radians), nil
}

func (mpe *MathPlusExtension) rad2degHandler(ctx runtime.ExecutionContext, args []*values.Value) (*values.Value, error) {
    if len(args) != 1 {
        return nil, fmt.Errorf("rad2deg() expects 1 parameter, %d given", len(args))
    }
    
    radians := args[0].ToFloat()
    degrees := radians * (180.0 / math.Pi)
    
    return values.NewFloat(degrees), nil
}

func (mpe *MathPlusExtension) factorialHandler(ctx runtime.ExecutionContext, args []*values.Value) (*values.Value, error) {
    if len(args) != 1 {
        return nil, fmt.Errorf("factorial() expects 1 parameter, %d given", len(args))
    }
    
    n := args[0].ToInt()
    
    // Validation
    if n < 0 {
        return nil, fmt.Errorf("factorial() expects non-negative integer, %d given", n)
    }
    
    // Calculate factorial
    var result int64 = 1
    for i := int64(2); i <= n; i++ {
        result *= i
    }
    
    return values.NewInt(result), nil
}

func (mpe *MathPlusExtension) statisticsHandler(ctx runtime.ExecutionContext, args []*values.Value) (*values.Value, error) {
    if len(args) == 0 {
        return nil, fmt.Errorf("statistics() expects at least 1 parameter")
    }
    
    // Convert all arguments to floats
    var numbers []float64
    for i, arg := range args {
        if arg.IsArray() {
            // Handle array input
            for j := 0; j < arg.ArrayCount(); j++ {
                element := arg.ArrayGet(values.NewInt(int64(j)))
                if element != nil {
                    numbers = append(numbers, element.ToFloat())
                }
            }
        } else {
            numbers = append(numbers, arg.ToFloat())
        }
    }
    
    if len(numbers) == 0 {
        return values.NewArray(), nil
    }
    
    // Calculate statistics
    var sum float64
    for _, n := range numbers {
        sum += n
    }
    mean := sum / float64(len(numbers))
    
    // Create result array
    result := values.NewArray()
    result.ArraySet(values.NewString("count"), values.NewInt(int64(len(numbers))))
    result.ArraySet(values.NewString("sum"), values.NewFloat(sum))
    result.ArraySet(values.NewString("mean"), values.NewFloat(mean))
    
    return result, nil
}
```

### Step 6: Add Classes (Optional)

For more complex functionality, you can define classes:

```go
func (mpe *MathPlusExtension) registerClasses(registry *runtime.RuntimeRegistry) error {
    // Define a Vector class
    vectorClass := &runtime.ClassDescriptor{
        Name: "Vector",
        Properties: map[string]*runtime.PropertyDescriptor{
            "x": {
                Name:         "x",
                Type:         "float",
                Visibility:   "public",
                DefaultValue: values.NewFloat(0.0),
            },
            "y": {
                Name:         "y",
                Type:         "float",
                Visibility:   "public",
                DefaultValue: values.NewFloat(0.0),
            },
        },
        Methods: map[string]*runtime.MethodDescriptor{
            "__construct": {
                Name:       "__construct",
                Visibility: "public",
                Parameters: []runtime.ParameterDescriptor{
                    {Name: "x", Type: "float", HasDefault: true, DefaultValue: values.NewFloat(0.0)},
                    {Name: "y", Type: "float", HasDefault: true, DefaultValue: values.NewFloat(0.0)},
                },
            },
            "magnitude": {
                Name:       "magnitude",
                Visibility: "public",
            },
        },
    }
    
    return mpe.RegisterClass(registry, vectorClass)
}
```

### Step 7: Register Extension

Finally, register your extension with the runtime:

```go
// In your main application or test
func main() {
    // Initialize runtime system
    err := runtime.Bootstrap()
    if err != nil {
        log.Fatal("Failed to bootstrap runtime:", err)
    }
    
    // Register your extension
    mathPlusExt := NewMathPlusExtension()
    err = runtime.GlobalRegistry.RegisterExtension(mathPlusExt)
    if err != nil {
        log.Fatal("Failed to register extension:", err)
    }
    
    // Your extension functions are now available in PHP execution
    fmt.Println("MathPlus extension registered successfully!")
}
```

## Advanced Features

### Dependency Management

Extensions can depend on other extensions:

```go
func NewAdvancedExtension() *AdvancedExtension {
    ext := &AdvancedExtension{
        BaseExtension: runtime.NewBaseExtension("advanced", "1.0.0", "Advanced features"),
    }
    
    // This extension depends on mathplus and json extensions
    ext.SetDependencies([]string{"mathplus", "json"})
    ext.SetLoadOrder(200) // Load after dependencies
    
    return ext
}
```

### Parameter Validation

Implement robust parameter validation:

```go
func validateParameters(args []*values.Value, expected []string) error {
    if len(args) != len(expected) {
        return fmt.Errorf("expected %d parameters, got %d", len(expected), len(args))
    }
    
    for i, expectedType := range expected {
        arg := args[i]
        switch expectedType {
        case "string":
            if !arg.IsString() {
                return fmt.Errorf("parameter %d must be string", i+1)
            }
        case "int":
            if !arg.IsInt() {
                return fmt.Errorf("parameter %d must be integer", i+1)
            }
        case "array":
            if !arg.IsArray() {
                return fmt.Errorf("parameter %d must be array", i+1)
            }
        }
    }
    
    return nil
}
```

### Error Handling

Provide meaningful error messages:

```go
func safeHandler(ctx runtime.ExecutionContext, args []*values.Value) (*values.Value, error) {
    // Validate input
    if err := validateParameters(args, []string{"int", "int"}); err != nil {
        return nil, fmt.Errorf("divide: %v", err)
    }
    
    divisor := args[1].ToInt()
    if divisor == 0 {
        return nil, fmt.Errorf("divide: division by zero")
    }
    
    dividend := args[0].ToInt()
    result := dividend / divisor
    
    return values.NewInt(result), nil
}
```

### Working with PHP Arrays

Handle PHP arrays in your extensions:

```go
func arrayProcessHandler(ctx runtime.ExecutionContext, args []*values.Value) (*values.Value, error) {
    if len(args) != 1 || !args[0].IsArray() {
        return nil, fmt.Errorf("expects array parameter")
    }
    
    inputArray := args[0]
    outputArray := values.NewArray()
    
    // Iterate over PHP array
    for i := 0; i < inputArray.ArrayCount(); i++ {
        key := values.NewInt(int64(i))
        value := inputArray.ArrayGet(key)
        
        if value != nil {
            // Process each element
            processedValue := processElement(value)
            outputArray.ArraySet(key, processedValue)
        }
    }
    
    return outputArray, nil
}

func processElement(value *values.Value) *values.Value {
    // Example: double numeric values, uppercase strings
    if value.IsInt() {
        return values.NewInt(value.ToInt() * 2)
    } else if value.IsString() {
        return values.NewString(strings.ToUpper(value.ToString()))
    }
    return value
}
```

## Best Practices

### 1. Naming Conventions

- **Extension Names**: Use lowercase, descriptive names (`mathplus`, `json`, `crypto`)
- **Function Names**: Follow PHP conventions (`snake_case`)
- **Constants**: Use UPPERCASE with extension prefix (`MATHPLUS_PI`)

### 2. Error Handling

- Always validate input parameters
- Provide clear, descriptive error messages
- Use Go's error return pattern consistently

### 3. Performance Considerations

- Avoid expensive operations in function handlers
- Cache computed values when appropriate
- Use efficient algorithms for array processing

### 4. Memory Management

- Don't store references to `*values.Value` objects across calls
- Let the runtime handle memory lifecycle
- Be careful with large data structures

### 5. Testing

Create comprehensive tests for your extensions:

```go
func TestMathPlusExtension(t *testing.T) {
    // Bootstrap runtime
    err := runtime.Bootstrap()
    require.NoError(t, err)
    
    // Register extension
    ext := NewMathPlusExtension()
    err = runtime.GlobalRegistry.RegisterExtension(ext)
    require.NoError(t, err)
    
    testCases := []struct {
        name     string
        code     string
        expected string
    }{
        {
            name:     "deg2rad",
            code:     `<?php echo deg2rad(180);`,
            expected: "3.14159",
        },
        {
            name:     "factorial",
            code:     `<?php echo factorial(5);`,
            expected: "120",
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Test your extension functionality
            result := executeCode(t, tc.code)
            assert.Contains(t, result, tc.expected)
        })
    }
}
```

## Troubleshooting

### Common Issues

#### 1. Extension Not Found
```
Error: extension not registered
```
**Solution**: Make sure you call `RegisterExtension()` before using the extension.

#### 2. Function Override Error
```
Error: cannot override built-in function
```
**Solution**: Built-in functions cannot be overridden. Choose a different name.

#### 3. Dependency Error
```
Error: missing dependency: core
```
**Solution**: Ensure all dependencies are registered before your extension.

#### 4. Circular Dependency
```
Error: circular dependency detected: A -> B -> A
```
**Solution**: Review your dependency chain and break the cycle.

### Debugging Tips

1. **Enable Debug Logging**: Add logging to your handlers to trace execution
2. **Test Incrementally**: Start with simple functions and build complexity
3. **Validate Early**: Check parameters at the start of handlers
4. **Use Test Cases**: Create comprehensive tests for all functionality

### Performance Issues

1. **Slow Function Calls**: Profile your handlers to find bottlenecks
2. **Memory Usage**: Monitor memory allocation in complex operations
3. **Registry Lookup**: The registry uses hash tables for O(1) lookup, but large numbers of entities may impact performance

## Conclusion

The extension system provides a powerful way to extend the PHP parser with custom functionality. By following this guide, you can:

- Create robust, well-tested extensions
- Handle dependencies and load ordering properly
- Implement efficient function handlers
- Follow best practices for maintainable code

Remember to start simple and build complexity incrementally. The runtime system is designed to be extensible while maintaining safety and performance.