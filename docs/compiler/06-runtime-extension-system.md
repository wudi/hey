# Runtime Extension System Design

This document explains the design and implementation of the unified runtime extension system in the PHP bytecode compiler, covering built-in registration, extension framework, and VM integration.

## Table of Contents

1. [Problem Overview](#problem-overview)
2. [Architecture Overview](#architecture-overview)
3. [Core Components](#core-components)
4. [Implementation Details](#implementation-details)
5. [Extension Development](#extension-development)
6. [VM Integration](#vm-integration)
7. [Testing and Validation](#testing-and-validation)

## Problem Overview

### The Challenge

Modern programming language runtimes need a systematic way to manage built-in functionality while allowing external extensions. The original PHP parser lacked a unified system for:

1. **Built-in Registration**: Constants, variables, functions, and classes scattered across different locations
2. **Extension Safety**: No protection against overriding core PHP entities
3. **Dependency Management**: No system for extension load ordering and dependencies
4. **Thread Safety**: No concurrent access protection for runtime entities
5. **VM Integration**: Tight coupling between built-ins and execution engine

### Requirements Analysis

Based on research of Python, V8 JavaScript, Ruby, and Go runtime systems, the requirements were:

- **Unified Registry**: Single source of truth for all runtime entities
- **Built-in Protection**: Extensions cannot override core PHP functionality
- **Thread Safety**: Safe concurrent access to registry
- **Clean Architecture**: No circular dependencies between VM and runtime
- **Extensibility**: Framework for external modules with dependency resolution

## Architecture Overview

### High-Level Design

The runtime extension system follows a layered architecture:

```
┌─────────────────────────────────────────────────────────────┐
│                    VM Execution Layer                       │
├─────────────────────────────────────────────────────────────┤
│                   VM Integration Layer                      │
│         (compiler/runtime/integration.go)                  │
├─────────────────────────────────────────────────────────────┤
│                   Extension Framework                       │
│         (compiler/runtime/extension.go)                    │
├─────────────────────────────────────────────────────────────┤
│                    Core Registry                            │
│         (compiler/runtime/registry.go)                     │
├─────────────────────────────────────────────────────────────┤
│                   Bootstrap System                          │
│         (compiler/runtime/bootstrap.go)                    │
└─────────────────────────────────────────────────────────────┘
```

### Design Principles

1. **Research-Informed**: Based on proven patterns from major language runtimes
2. **Interface Abstraction**: Prevents circular dependencies through interfaces
3. **Thread-Safe**: All operations protected by mutex locks
4. **Extensible**: Clean extension API with dependency management
5. **Performance-Oriented**: Efficient lookup and minimal overhead

## Core Components

### RuntimeRegistry - Central Registration System

The `RuntimeRegistry` is the heart of the system, managing all runtime entities:

```go
type RuntimeRegistry struct {
    mu sync.RWMutex
    
    // Core registry maps
    constants map[string]*values.Value
    variables map[string]*values.Value
    functions map[string]*FunctionDescriptor
    classes   map[string]*ClassDescriptor
    extensions map[string]*ExtensionDescriptor
    
    // Built-in protection
    builtinConstants map[string]bool
    builtinVariables map[string]bool
    builtinFunctions map[string]bool
    builtinClasses   map[string]bool
}
```

**Key Features:**
- **Thread-Safe Operations**: All access protected by `sync.RWMutex`
- **Conflict Detection**: Prevents extensions from overriding built-ins
- **Type Safety**: Strongly typed descriptors for all entities
- **Bulk Operations**: Efficient methods for retrieving all entities

### Bootstrap System - Built-in PHP Standard Library

The bootstrap system (`compiler/runtime/bootstrap.go`) provides complete PHP built-in functionality:

#### Constants (100+ entities)
```go
// PHP Version constants
"PHP_VERSION":          values.NewString("8.4.0"),
"PHP_MAJOR_VERSION":    values.NewInt(8),

// Math constants  
"M_PI":                 values.NewFloat(math.Pi),
"M_E":                  values.NewFloat(math.E),

// Error constants
"E_ERROR":              values.NewInt(1),
"E_WARNING":            values.NewInt(2),
```

#### Superglobal Variables
```go
// $_SERVER with realistic server environment
server := createServerArray() // 30+ server variables

// Other superglobals
"_GET":     values.NewArray(),
"_POST":    values.NewArray(),
"_SESSION": values.NewArray(),
"GLOBALS":  values.NewArray(),
```

#### Built-in Functions (40+ implementations)
```go
// String functions
strlen, substr, strpos, strtolower, strtoupper

// Array functions  
count, array_push, array_pop, array_keys, array_values

// Type checking functions
is_string, is_int, is_float, is_bool, is_array, is_null

// Math functions
abs, max, min, round, floor, ceil

// Output functions
var_dump, print_r, echo
```

#### Built-in Classes
```go
// Exception hierarchy
Exception -> InvalidArgumentException
          -> RuntimeException
          -> LogicException

// Core classes
stdClass, DateTime, ReflectionClass
```

### Extension Framework - Plugin Architecture

The extension system provides a complete framework for external modules:

#### Extension Interface
```go
type Extension interface {
    GetName() string
    GetVersion() string
    GetDescription() string
    GetDependencies() []string
    GetLoadOrder() int
    Register(registry *RuntimeRegistry) error
    Unregister(registry *RuntimeRegistry) error
}
```

#### BaseExtension Foundation
```go
type BaseExtension struct {
    name         string
    version      string
    description  string
    dependencies []string
    loadOrder    int
    
    // Tracking for cleanup
    registeredConstants []string
    registeredFunctions []string
    registeredClasses   []string
}
```

#### ExtensionManager
```go
type ExtensionManager struct {
    registry   *RuntimeRegistry
    extensions map[string]Extension
    loadOrder  []Extension
}
```

**Features:**
- **Dependency Resolution**: Automatically resolves and orders dependencies
- **Load Ordering**: Configurable load priority for extensions
- **Circular Dependency Detection**: Prevents invalid dependency chains
- **Automatic Cleanup**: Tracks registered entities for proper unregistration

### VM Integration Layer

The integration layer (`compiler/runtime/integration.go`) provides clean separation:

```go
type VMIntegration struct {
    registry *RuntimeRegistry
}
```

**Responsibilities:**
- **Interface Abstraction**: Breaks circular dependencies through interfaces
- **Context Initialization**: Populates VM execution context with runtime entities
- **Function Routing**: Directs VM function calls to runtime registry
- **State Management**: Manages runtime state lifecycle

## Implementation Details

### Thread Safety Implementation

All registry operations use proper locking:

```go
func (r *RuntimeRegistry) RegisterFunction(descriptor *FunctionDescriptor) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    // Check for conflicts
    if r.builtinFunctions[descriptor.Name] && !descriptor.IsBuiltin {
        return fmt.Errorf("cannot override built-in function: %s", descriptor.Name)
    }
    
    r.functions[descriptor.Name] = descriptor
    if descriptor.IsBuiltin {
        r.builtinFunctions[descriptor.Name] = true
    }
    
    return nil
}
```

### Built-in Protection Mechanism

The system prevents extensions from overriding core PHP functionality:

```go
// Built-in functions are protected
if r.builtinFunctions[name] && !isBuiltin {
    return fmt.Errorf("cannot override built-in function: %s", name)
}

// Extensions are validated before registration
if err := em.validateExtension(ext); err != nil {
    return fmt.Errorf("extension validation failed: %v", err)
}
```

### Function Call Integration

The VM integrates seamlessly with the runtime registry:

```go
// In VM function execution
if runtimeRegistry.GlobalVMIntegration != nil && 
   runtimeRegistry.GlobalVMIntegration.HasFunction(functionName) {
    
    result, err := runtimeRegistry.GlobalVMIntegration.CallFunction(
        ctx, functionName, ctx.CallContext.Arguments)
    if err != nil {
        return err
    }
    
    // Store result and continue execution
    vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), result)
}
```

## Extension Development

### Creating a Custom Extension

#### Step 1: Extend BaseExtension
```go
type MathExtension struct {
    *BaseExtension
}

func NewMathExtension() *MathExtension {
    return &MathExtension{
        BaseExtension: NewBaseExtension("math", "1.0.0", "Extended math functions"),
    }
}
```

#### Step 2: Implement Registration
```go
func (me *MathExtension) Register(registry *RuntimeRegistry) error {
    // Call base registration
    if err := me.BaseExtension.Register(registry); err != nil {
        return err
    }
    
    // Add constants
    me.RegisterConstant(registry, "MATH_E", values.NewFloat(math.E))
    
    // Add functions
    me.RegisterFunction(registry, &FunctionDescriptor{
        Name:    "deg2rad",
        Handler: deg2radHandler,
        Parameters: []ParameterDescriptor{
            {Name: "degrees", Type: "float"},
        },
        MinArgs: 1,
        MaxArgs: 1,
    })
    
    return nil
}
```

#### Step 3: Implement Function Handlers
```go
func deg2radHandler(ctx ExecutionContext, args []*values.Value) (*values.Value, error) {
    if len(args) != 1 {
        return nil, fmt.Errorf("deg2rad() expects 1 parameter, %d given", len(args))
    }
    
    degrees := args[0].ToFloat()
    radians := degrees * (math.Pi / 180.0)
    return values.NewFloat(radians), nil
}
```

### Extension Registration and Usage

```go
// Initialize runtime system
err := runtime.Bootstrap()
if err != nil {
    log.Fatal(err)
}

// Create and register extension
mathExt := NewMathExtension()
err = runtime.GlobalRegistry.RegisterExtension(mathExt)
if err != nil {
    log.Fatal(err)
}

// Extension functions are now available in PHP execution
```

### Dependency Management Example

```go
type JsonExtension struct {
    *BaseExtension
}

func NewJsonExtension() *JsonExtension {
    ext := &JsonExtension{
        BaseExtension: NewBaseExtension("json", "1.0.0", "JSON processing"),
    }
    
    // Set dependencies and load order
    ext.SetDependencies([]string{"core", "string"})
    ext.SetLoadOrder(200) // Load after core extensions
    
    return ext
}
```

## VM Integration

### Initialization Process

The VM integration follows a specific initialization sequence:

```go
// 1. Bootstrap runtime system
err := runtime.Bootstrap()
if err != nil {
    return err
}

// 2. Initialize VM integration
err = runtime.InitializeVMIntegration()
if err != nil {
    return err
}

// 3. Create VM and execution context
vmachine := vm.NewVirtualMachine()
vmCtx := vm.NewExecutionContext()

// 4. Initialize context with runtime entities
variables := runtime.GlobalVMIntegration.GetAllVariables()
for name, value := range variables {
    vmCtx.GlobalVars[name] = value
}
```

### Function Call Resolution

The VM uses a hierarchical lookup system:

```
1. Check runtime registry for built-in/extension functions
2. Check VM context function table for user-defined functions  
3. Return "function not found" error
```

### Interface Abstraction

To prevent circular dependencies, the runtime uses interface abstraction:

```go
// Runtime defines interface
type ExecutionContext interface {
    // Methods needed for function handlers
}

// VM implements interface
type ExecutionContext struct {
    // VM-specific fields
}

// Function handlers use interface
type FunctionHandler func(ctx ExecutionContext, args []*values.Value) (*values.Value, error)
```

## Testing and Validation

### Test Framework Integration

The test framework provides proper runtime initialization:

```go
func executeWithRuntime(t *testing.T, comp *Compiler) error {
    // Initialize runtime if not already done
    if runtime.GlobalRegistry == nil {
        err := runtime.Bootstrap()
        require.NoError(t, err, "Failed to bootstrap runtime")
    }
    
    // Initialize VM integration
    if runtime.GlobalVMIntegration == nil {
        err := runtime.InitializeVMIntegration()
        require.NoError(t, err, "Failed to initialize VM integration")
    }
    
    // Execute with runtime support
    vmachine := vm.NewVirtualMachine()
    vmCtx := vm.NewExecutionContext()
    
    // Initialize global variables from runtime
    variables := runtime.GlobalVMIntegration.GetAllVariables()
    for name, value := range variables {
        vmCtx.GlobalVars[name] = value
    }
    
    return vmachine.Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), 
                           comp.GetFunctions(), comp.GetClasses())
}
```

### Built-in Function Tests

Comprehensive tests validate built-in functionality:

```go
func TestBuiltinFunctions(t *testing.T) {
    testCases := []struct {
        name string
        code string
    }{
        {
            name: "strlen",
            code: `<?php echo strlen("hello");`, // Expected: 5
        },
        {
            name: "count_array", 
            code: `<?php $arr = [1, 2, 3]; echo count($arr);`, // Expected: 3
        },
        {
            name: "is_string",
            code: `<?php $str = "hello"; var_dump(is_string($str));`, // Expected: bool(true)
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Parse and compile
            p := parser.New(lexer.New(tc.code))
            prog := p.ParseProgram()
            
            comp := NewCompiler()
            err := comp.Compile(prog)
            require.NoError(t, err)

            // Execute with runtime
            err = executeWithRuntime(t, comp)
            require.NoError(t, err)
        })
    }
}
```

### Extension Testing

Extension tests validate registration and functionality:

```go
func TestMathExtension(t *testing.T) {
    // Bootstrap runtime
    err := runtime.Bootstrap()
    require.NoError(t, err)
    
    // Register extension
    mathExt := NewMathExtension()
    err = runtime.GlobalRegistry.RegisterExtension(mathExt)
    require.NoError(t, err)
    
    // Test extension function
    code := `<?php echo deg2rad(180);` // Expected: 3.14159...
    // ... execute and validate
}
```

### Performance Validation

The system maintains performance characteristics:

- **Registry Lookup**: O(1) hash table access
- **Function Calls**: Minimal overhead over direct calls
- **Memory Usage**: Efficient shared storage for built-ins
- **Concurrency**: Read-mostly workload with RWMutex optimization

## Conclusion

The runtime extension system provides a comprehensive, research-informed solution for managing PHP built-ins and extensions. Key achievements:

1. **Unified Management**: Single registry for all runtime entities
2. **Built-in Protection**: Safe extension system that cannot break core functionality
3. **Thread Safety**: Proper concurrent access handling
4. **Clean Architecture**: No circular dependencies, proper separation of concerns
5. **Extensibility**: Complete framework for external modules
6. **Performance**: Efficient implementation with minimal overhead

The system successfully integrates with the existing VM while providing a solid foundation for future PHP standard library expansion and external extension development.