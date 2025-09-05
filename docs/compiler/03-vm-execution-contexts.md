# VM Execution Contexts Design

This document explains the design and implementation of execution contexts in the PHP virtual machine, covering context creation, variable scoping, function calls, and memory management.

## Table of Contents

1. [Problem Overview](#problem-overview)
2. [Architecture Overview](#architecture-overview)
3. [Core Components](#core-components)
4. [Context Lifecycle](#context-lifecycle)
5. [Memory Management](#memory-management)
6. [PHP Compatibility](#php-compatibility)

## Problem Overview

### The Challenge

PHP execution requires isolated contexts for different scopes:

```php
<?php
function outer($x) {
    $local = $x * 2;
    
    function inner($y) {
        $inner_local = $y + 1;  // Isolated from outer scope
        return $inner_local;
    }
    
    $result = inner($local);
    return $result;
}

$global = outer(5);  // Each call needs its own context
```

The challenges include:
- **Scope Isolation**: Function calls must not interfere with each other
- **Variable Management**: Each context needs its own variable storage
- **Stack Management**: Proper stack handling for temporary values
- **Resource Cleanup**: Memory management and context cleanup
- **Call Stack**: Tracking function call hierarchy

## Architecture Overview

The VM uses a hierarchical context system:

```
Global Context
├── Function Context (outer)
│   ├── Local Variables
│   ├── Parameters  
│   └── Function Context (inner)
│       ├── Local Variables
│       └── Parameters
└── Iterator Contexts (foreach loops)
```

### Core Components

#### 1. ExecutionContext Structure

```go
type ExecutionContext struct {
    // Bytecode execution
    Instructions []opcodes.Instruction
    IP           int // Instruction pointer
    
    // Runtime stacks
    Stack         []*values.Value
    SP            int // Stack pointer
    MaxStackSize  int
    
    // Variable storage
    Variables     map[uint32]*values.Value // Variable slots
    Constants     []*values.Value          // Constant pool
    Temporaries   map[uint32]*values.Value // Temporary variables
    
    // Function call stack
    CallStack     []CallFrame
    
    // Global state
    GlobalVars    map[string]*values.Value
    Functions     map[string]*Function
    
    // Loop state
    ForeachIterators map[uint32]*ForeachIterator
    Classes       map[string]*Class
    
    // Function call state
    CallContext   *CallContext // Current function call being prepared
    
    // Error handling
    ExceptionStack    []Exception
    ExceptionHandlers []ExceptionHandler
    CurrentException  *Exception
    
    // Execution control
    Halted        bool
    ExitCode      int
}
```

#### 2. CallFrame Structure

```go
type CallFrame struct {
    Function      *Function
    ReturnIP      int
    Variables     map[uint32]*values.Value
    ThisObject    *values.Value
    Arguments     []*values.Value
}
```

#### 3. CallContext for Function Preparation

```go
type CallContext struct {
    FunctionName string
    Arguments    []*values.Value
    NumArgs      int
}
```

## Context Lifecycle

### 1. Context Creation

```go
func NewExecutionContext() *ExecutionContext {
    return &ExecutionContext{
        Instructions: make([]opcodes.Instruction, 0),
        IP:          0,
        Stack:       make([]*values.Value, 0),
        SP:          0,
        MaxStackSize: 1000,
        Variables:   make(map[uint32]*values.Value),
        Constants:   make([]*values.Value, 0),
        Temporaries: make(map[uint32]*values.Value),
        CallStack:   make([]CallFrame, 0),
        GlobalVars:  make(map[string]*values.Value),
        Functions:   make(map[string]*Function),
        ForeachIterators: make(map[uint32]*ForeachIterator),
        Classes:     make(map[string]*Class),
        CallContext: nil,
        ExceptionStack: make([]Exception, 0),
        ExceptionHandlers: make([]ExceptionHandler, 0),
        CurrentException: nil,
        Halted:      false,
        ExitCode:    0,
    }
}
```

### 2. Function Call Context Creation

```go
func (vm *VirtualMachine) executeDoFunctionCall(ctx *ExecutionContext, inst *opcodes.Instruction) error {
    // Get function from call context
    function := ctx.Functions[ctx.CallContext.FunctionName]
    
    // Create new execution context for function
    functionCtx := NewExecutionContext()
    functionCtx.Instructions = function.Instructions
    functionCtx.Constants = function.Constants
    functionCtx.Functions = ctx.Functions // Share function table
    
    // Set up function parameters - map to variable slots
    for i, param := range function.Parameters {
        if functionCtx.Variables == nil {
            functionCtx.Variables = make(map[uint32]*values.Value)
        }
        
        if i < len(ctx.CallContext.Arguments) {
            // Map argument to parameter variable slot
            functionCtx.Variables[uint32(i)] = ctx.CallContext.Arguments[i]
        } else if param.HasDefault {
            // Use default value
            functionCtx.Variables[uint32(i)] = values.NewNull()
        } else {
            return fmt.Errorf("missing required parameter %s", param.Name)
        }
    }
    
    // Execute the function in isolated context
    err := vm.Execute(functionCtx, function.Instructions, function.Constants, ctx.Functions)
    if err != nil {
        return fmt.Errorf("error executing function: %v", err)
    }
    
    // Get return value from function's stack
    var result *values.Value
    if len(functionCtx.Stack) > 0 {
        result = functionCtx.Stack[len(functionCtx.Stack)-1]
    } else {
        result = values.NewNull()
    }
    
    // Store result in caller's context
    vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), result)
    
    // Clear call context
    ctx.CallContext = nil
    return nil
}
```

### 3. Variable Access Patterns

#### Global Scope Variables
```go
// Global variables stored by name
ctx.GlobalVars["$global_var"] = value
```

#### Function Parameters and Locals
```go
// Function parameters mapped to slots 0, 1, 2, ...
ctx.Variables[0] = parameter1_value  // $n in function foo($n)
ctx.Variables[1] = parameter2_value  // $x in function foo($n, $x)

// Local variables allocated incrementally
ctx.Variables[2] = local_var1        // $ret in function body
ctx.Variables[3] = local_var2        // $i in for loop
```

#### Temporary Variables
```go
// Temporary values during expression evaluation
ctx.Temporaries[1000] = expression_result
ctx.Temporaries[1001] = intermediate_value
```

### 4. Stack Management

```go
func (vm *VirtualMachine) pushValue(ctx *ExecutionContext, value *values.Value) error {
    if ctx.SP >= ctx.MaxStackSize {
        return fmt.Errorf("stack overflow")
    }
    
    ctx.Stack = append(ctx.Stack, value)
    ctx.SP++
    return nil
}

func (vm *VirtualMachine) popValue(ctx *ExecutionContext) (*values.Value, error) {
    if ctx.SP <= 0 {
        return nil, fmt.Errorf("stack underflow")
    }
    
    value := ctx.Stack[ctx.SP-1]
    ctx.Stack = ctx.Stack[:ctx.SP-1]
    ctx.SP--
    return value, nil
}
```

## Memory Management

### 1. Context Isolation

Each function call gets a completely isolated execution context:
- **Variables**: Separate variable slots prevent interference
- **Stack**: Independent stack for temporary values
- **Constants**: Shared constant pool for efficiency
- **Functions**: Shared function table for recursive calls

### 2. Resource Cleanup

```go
func (ctx *ExecutionContext) cleanup() {
    // Clear variable references
    for k := range ctx.Variables {
        delete(ctx.Variables, k)
    }
    
    // Clear temporary values
    for k := range ctx.Temporaries {
        delete(ctx.Temporaries, k)
    }
    
    // Clear stack
    ctx.Stack = ctx.Stack[:0]
    ctx.SP = 0
    
    // Clear foreach iterators
    for k := range ctx.ForeachIterators {
        delete(ctx.ForeachIterators, k)
    }
}
```

### 3. Iterator Context Management

Foreach loops maintain separate iterator state:
```go
// Each foreach gets a unique iterator ID
ctx.ForeachIterators[iteratorID] = &ForeachIterator{
    Array:   iterable,
    Index:   0,
    Keys:    extracted_keys,
    Values:  extracted_values,
    HasMore: true,
}
```

## Value Access Methods

### 1. getValue - Unified Value Access

```go
func (vm *VirtualMachine) getValue(ctx *ExecutionContext, operand uint32, opType opcodes.OpType) *values.Value {
    switch opType {
    case opcodes.IS_CONST:
        if int(operand) < len(ctx.Constants) {
            return ctx.Constants[operand]
        }
        return values.NewNull()
        
    case opcodes.IS_TMP_VAR:
        if val, exists := ctx.Temporaries[operand]; exists {
            return val
        }
        return values.NewNull()
        
    case opcodes.IS_VAR:
        if val, exists := ctx.Variables[operand]; exists {
            return val
        }
        return values.NewNull()
        
    default:
        return values.NewNull()
    }
}
```

### 2. setValue - Unified Value Storage

```go
func (vm *VirtualMachine) setValue(ctx *ExecutionContext, operand uint32, opType opcodes.OpType, value *values.Value) {
    switch opType {
    case opcodes.IS_TMP_VAR:
        if ctx.Temporaries == nil {
            ctx.Temporaries = make(map[uint32]*values.Value)
        }
        ctx.Temporaries[operand] = value
        
    case opcodes.IS_VAR:
        if ctx.Variables == nil {
            ctx.Variables = make(map[uint32]*values.Value)
        }
        ctx.Variables[operand] = value
    }
}
```

## PHP Compatibility

The execution context design ensures PHP compatibility:

1. **Variable Scoping**: Each function has isolated variable space
2. **Parameter Passing**: Arguments properly mapped to parameter slots
3. **Return Values**: Functions can return any PHP value type
4. **Nested Calls**: Recursive and nested function calls work correctly
5. **Iterator Isolation**: Foreach loops maintain separate state

## Performance Characteristics

- **Context Creation**: O(1) for basic context setup
- **Variable Access**: O(1) hash map lookup
- **Stack Operations**: O(1) push/pop operations
- **Memory Usage**: Minimal overhead per context
- **Cleanup**: O(n) where n is number of variables/temporaries

This architecture provides the foundation for proper PHP execution semantics while maintaining performance and memory efficiency.