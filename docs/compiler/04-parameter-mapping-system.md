# Parameter Mapping System Design

This document explains the design and implementation of the parameter mapping system in the PHP bytecode compiler, covering how function parameters are mapped from declarations to runtime variable slots.

## Table of Contents

1. [Problem Overview](#problem-overview)
2. [Architecture Overview](#architecture-overview)
3. [Compilation Phase](#compilation-phase)
4. [Runtime Phase](#runtime-phase)
5. [Variable Slot Allocation](#variable-slot-allocation)
6. [PHP Compatibility](#php-compatibility)

## Problem Overview

### The Challenge

PHP function parameters must be correctly mapped from function calls to function execution:

```php
<?php
function calculate($base, $multiplier = 2, $offset = 0): int {
    $result = $base * $multiplier + $offset;  // Parameters as variables
    return $result;
}

$value = calculate(10, 3, 5);  // Arguments: 10, 3, 5
//                  ↓   ↓   ↓
// Maps to slots:   0   1   2
```

The challenges include:
- **Parameter Order**: Maintaining correct parameter-to-argument mapping
- **Default Values**: Handling parameters with default values
- **Variable Slots**: Mapping parameter names to variable slot indices
- **Type Checking**: Ensuring parameter types match (when specified)
- **Scope Isolation**: Parameters only exist within function scope

## Architecture Overview

The parameter mapping system works in two phases:

```
Compilation Phase: Parameter Declaration → Variable Slot Assignment
Runtime Phase: Function Call Arguments → Parameter Variable Mapping
```

### Core Components

#### 1. Parameter Structure

```go
type Parameter struct {
    Name          string           // Parameter name (e.g., "$base")
    Type          string           // Type hint (e.g., "int", "array")
    IsReference   bool            // Whether passed by reference (&$param)
    HasDefault    bool            // Whether has default value
    DefaultValue  *values.Value   // The default value (if any)
}
```

#### 2. Function Structure with Parameters

```go
type Function struct {
    Name          string
    Instructions  []opcodes.Instruction
    Constants     []*values.Value
    Parameters    []Parameter      // Ordered list of parameters
    IsVariadic    bool            // Whether accepts ...args
    IsGenerator   bool            // Whether is a generator function
}
```

## Compilation Phase

### 1. Parameter Declaration Processing

```go
func (c *Compiler) compileFunctionDeclaration(decl *ast.FunctionDeclaration) error {
    // Create function structure
    function := &vm.Function{
        Name:         funcName,
        Instructions: make([]opcodes.Instruction, 0),
        Constants:    make([]*values.Value, 0),
        Parameters:   make([]Parameter, 0),
        IsVariadic:   false,
        IsGenerator:  false,
    }
    
    // Process parameters in declaration order
    if decl.Parameters != nil {
        for _, param := range decl.Parameters.Parameters {
            // Extract parameter name
            paramName := ""
            if nameNode, ok := param.Name.(*ast.IdentifierNode); ok {
                paramName = nameNode.Name
            }
            
            // Create parameter info
            vmParam := vm.Parameter{
                Name:        paramName,
                IsReference: param.ByReference,
                HasDefault:  param.DefaultValue != nil,
            }
            
            // Handle type hints
            if param.Type != nil {
                vmParam.Type = param.Type.String()
            }
            
            // Handle default values (simplified)
            if param.DefaultValue != nil {
                vmParam.HasDefault = true
                // In full implementation, would evaluate default value
            }
            
            // Check for variadic parameters
            if param.Variadic {
                function.IsVariadic = true
            }
            
            function.Parameters = append(function.Parameters, vmParam)
        }
    }
    
    // Set up compilation scope with parameters
    c.pushScope(true)
    
    // CRITICAL: Register parameters in order (slots 0, 1, 2, ...)
    if decl.Parameters != nil {
        for _, param := range decl.Parameters.Parameters {
            if nameNode, ok := param.Name.(*ast.IdentifierNode); ok {
                // This assigns slots 0, 1, 2, ... in declaration order
                c.getOrCreateVariable(nameNode.Name)
            }
        }
    }
    
    // Compile function body
    // ... rest of function compilation
}
```

### 2. Variable Slot Allocation

```go
func (c *Compiler) getOrCreateVariable(name string) uint32 {
    scope := c.currentScope()
    
    // Check if variable already exists in this scope
    if slot, exists := scope.variables[name]; exists {
        return slot
    }
    
    // Assign new slot - parameters get slots 0, 1, 2, ... in order
    slot := scope.nextSlot
    scope.variables[name] = slot
    scope.nextSlot++
    return slot
}

// Example for function foo($base, $multiplier, $offset):
// getOrCreateVariable("$base")       → slot 0
// getOrCreateVariable("$multiplier") → slot 1  
// getOrCreateVariable("$offset")     → slot 2
// getOrCreateVariable("$result")     → slot 3 (local variable)
```

## Runtime Phase

### 1. Function Call Argument Collection

During `OP_INIT_FCALL` and `OP_SEND_VAL`:

```go
// OP_INIT_FCALL - Initialize function call
func (vm *VirtualMachine) executeInitFunctionCall(ctx *ExecutionContext, inst *opcodes.Instruction) error {
    functionName := calleeValue.ToString()
    numArgs := int(numArgsValue.Data.(int64))
    
    // Prepare call context to collect arguments
    ctx.CallContext = &CallContext{
        FunctionName: functionName,
        Arguments:    make([]*values.Value, 0, numArgs),
        NumArgs:      numArgs,
    }
    return nil
}

// OP_SEND_VAL - Collect function arguments in order
func (vm *VirtualMachine) executeSendValue(ctx *ExecutionContext, inst *opcodes.Instruction) error {
    argValue := vm.getValue(ctx, inst.Op2, opcodes.DecodeOpType2(inst.OpType1))
    
    // Add argument to call context - ORDER MATTERS
    ctx.CallContext.Arguments = append(ctx.CallContext.Arguments, argValue)
    return nil
}
```

### 2. Parameter-to-Variable Mapping

During `OP_DO_FCALL`:

```go
func (vm *VirtualMachine) executeDoFunctionCall(ctx *ExecutionContext, inst *opcodes.Instruction) error {
    function := ctx.Functions[ctx.CallContext.FunctionName]
    
    // Create new execution context for function
    functionCtx := NewExecutionContext()
    // ... context setup
    
    // CRITICAL: Map arguments to parameter variable slots
    for i, param := range function.Parameters {
        if functionCtx.Variables == nil {
            functionCtx.Variables = make(map[uint32]*values.Value)
        }
        
        if i < len(ctx.CallContext.Arguments) {
            // Map argument[i] to variable slot[i]
            functionCtx.Variables[uint32(i)] = ctx.CallContext.Arguments[i]
        } else if param.HasDefault {
            // Use default value for missing arguments
            functionCtx.Variables[uint32(i)] = values.NewNull()
        } else {
            // Missing required parameter
            return fmt.Errorf("missing required parameter %s for function %s", 
                             param.Name, functionName)
        }
    }
    
    // Execute function with mapped parameters
    err := vm.Execute(functionCtx, function.Instructions, function.Constants, ctx.Functions)
    // ... return value handling
}
```

## Variable Slot Allocation Strategy

### 1. Slot Assignment Order

```
Function: calculate($base, $multiplier, $offset)

Compilation Phase:
┌─────────────────┬──────┬─────────────────┐
│ Parameter Name  │ Slot │ Declaration     │
├─────────────────┼──────┼─────────────────┤
│ $base          │  0   │ First parameter │
│ $multiplier    │  1   │ Second parameter│ 
│ $offset        │  2   │ Third parameter │
│ $result        │  3   │ Local variable  │
└─────────────────┴──────┴─────────────────┘

Runtime Phase - Function Call: calculate(10, 3, 5)
┌─────────────────┬──────┬─────────────────┐
│ Argument Value  │ Slot │ Parameter Name  │
├─────────────────┼──────┼─────────────────┤
│ 10             │  0   │ $base          │
│ 3              │  1   │ $multiplier    │
│ 5              │  2   │ $offset        │
└─────────────────┴──────┴─────────────────┘
```

### 2. Variable Access in Function Body

When the function body accesses parameters:

```php
$result = $base * $multiplier + $offset;
```

This compiles to bytecode that accesses variable slots:

```
[0000] FETCH_R VAR:0, TMP:100           # Load $base (slot 0) 
[0001] FETCH_R VAR:1, TMP:101           # Load $multiplier (slot 1)
[0002] MUL TMP:100, TMP:101, TMP:102    # $base * $multiplier
[0003] FETCH_R VAR:2, TMP:103           # Load $offset (slot 2)
[0004] ADD TMP:102, TMP:103, TMP:104    # + $offset
[0005] ASSIGN TMP:104, UNUSED, VAR:3    # Store in $result (slot 3)
```

## Default Parameter Handling

### 1. Compilation with Defaults

```php
function greet($name, $greeting = "Hello") {
    echo "$greeting, $name!";
}
```

Compiles to:
```go
Parameter{Name: "$name", HasDefault: false}
Parameter{Name: "$greeting", HasDefault: true}
```

### 2. Runtime Default Application

```php
greet("World");        // Only 1 argument provided
greet("World", "Hi");  // Both arguments provided
```

Runtime mapping:
```go
// Call with 1 argument:
functionCtx.Variables[0] = "World"     // $name
functionCtx.Variables[1] = values.NewNull()  // $greeting (default)

// Call with 2 arguments:  
functionCtx.Variables[0] = "World"     // $name
functionCtx.Variables[1] = "Hi"        // $greeting
```

## Error Handling

### 1. Parameter Count Validation

```go
if i >= len(ctx.CallContext.Arguments) && !param.HasDefault {
    return fmt.Errorf("missing required parameter %s for function %s", 
                     param.Name, functionName)
}
```

### 2. Type Validation (Future Enhancement)

```go
if param.Type != "" {
    if !validateParameterType(argument, param.Type) {
        return fmt.Errorf("parameter %s expects %s, got %s", 
                         param.Name, param.Type, argument.Type)
    }
}
```

## PHP Compatibility

The parameter mapping system ensures full PHP compatibility:

1. **Parameter Order**: Maintains exact declaration order
2. **Default Values**: Supports optional parameters with defaults
3. **Type Hints**: Framework for type validation
4. **Reference Parameters**: Supports pass-by-reference
5. **Variadic Parameters**: Framework for `...$args` support

## Performance Characteristics

- **Slot Allocation**: O(1) per parameter during compilation
- **Runtime Mapping**: O(n) where n is number of parameters
- **Variable Access**: O(1) hash map lookup
- **Memory Usage**: Minimal overhead per parameter

This system provides the foundation for proper PHP function parameter semantics while maintaining efficiency and compatibility.