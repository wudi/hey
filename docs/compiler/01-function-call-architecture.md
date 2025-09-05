# Function Call Architecture Design

This document explains the design and implementation of the function call system in the PHP bytecode compiler, covering function declarations, call initialization, argument passing, and execution.

## Table of Contents

1. [Problem Overview](#problem-overview)
2. [Architecture Overview](#architecture-overview)
3. [Core Components](#core-components)
4. [Implementation Flow](#implementation-flow)
5. [Bytecode Instructions](#bytecode-instructions)
6. [PHP Compatibility](#php-compatibility)

## Problem Overview

### The Challenge

PHP function calls involve several complex operations that must be coordinated:

```php
<?php
function foo($n): array {
    $ret = [];
    for($i=0; $i<$n; $i++) {
        $ret[] = $i;
    }
    return $ret;
}

$result = foo(5); // Function call with parameter passing and return value
```

The challenges include:
- **Function Declaration**: Compiling function definitions and storing them for runtime access
- **Parameter Mapping**: Mapping function arguments to parameter variable slots
- **Execution Context**: Creating isolated execution environments for functions
- **Return Values**: Properly returning values from functions to callers

## Architecture Overview

The function call system follows PHP's Zend VM architecture with these phases:

```
Function Declaration → Function Registration → Call Initialization → 
Argument Passing → Function Execution → Return Value Handling
```

### Core Components

#### 1. Function Declaration (`OP_DECLARE_FUNCTION`)

**Compilation Phase:**
```go
// Store function in compiler's function table
function := &vm.Function{
    Name:         funcName,
    Instructions: compiled_bytecode,
    Constants:    compiled_constants,
    Parameters:   parameter_info,
}
c.functions[funcName] = function
```

**Runtime Phase:**
```go
func (vm *VirtualMachine) executeDeclareFunction(ctx *ExecutionContext, inst *opcodes.Instruction) error {
    // Function already available in ctx.Functions (passed from compiler)
    // This opcode confirms function availability at runtime
}
```

#### 2. Function Call Initialization (`OP_INIT_FCALL`)

```go
func (vm *VirtualMachine) executeInitFunctionCall(ctx *ExecutionContext, inst *opcodes.Instruction) error {
    // Extract function name and argument count
    calleeValue := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
    numArgsValue := vm.getValue(ctx, inst.Op2, opcodes.DecodeOpType2(inst.OpType1))
    
    // Initialize call context
    ctx.CallContext = &CallContext{
        FunctionName: functionName,
        Arguments:    make([]*values.Value, 0, numArgs),
        NumArgs:      numArgs,
    }
}
```

#### 3. Argument Passing (`OP_SEND_VAL`)

```go
func (vm *VirtualMachine) executeSendValue(ctx *ExecutionContext, inst *opcodes.Instruction) error {
    // Get argument value and add to call context
    argValue := vm.getValue(ctx, inst.Op2, opcodes.DecodeOpType2(inst.OpType1))
    ctx.CallContext.Arguments = append(ctx.CallContext.Arguments, argValue)
}
```

#### 4. Function Execution (`OP_DO_FCALL`)

```go
func (vm *VirtualMachine) executeDoFunctionCall(ctx *ExecutionContext, inst *opcodes.Instruction) error {
    // Look up function
    function := ctx.Functions[ctx.CallContext.FunctionName]
    
    // Create new execution context for function
    functionCtx := NewExecutionContext()
    functionCtx.Instructions = function.Instructions
    functionCtx.Constants = function.Constants
    
    // Map arguments to parameter variable slots
    for i, param := range function.Parameters {
        if i < len(ctx.CallContext.Arguments) {
            functionCtx.Variables[uint32(i)] = ctx.CallContext.Arguments[i]
        }
    }
    
    // Execute function
    err := vm.Execute(functionCtx, function.Instructions, function.Constants, ctx.Functions)
    
    // Get return value from function stack
    var result *values.Value
    if len(functionCtx.Stack) > 0 {
        result = functionCtx.Stack[len(functionCtx.Stack)-1]
    } else {
        result = values.NewNull()
    }
    
    // Store result and clear call context
    vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), result)
    ctx.CallContext = nil
}
```

## Implementation Flow

### 1. Compilation Phase

```go
// Function declaration compilation
func (c *Compiler) compileFunctionDeclaration(decl *ast.FunctionDeclaration) error {
    // 1. Create function object
    function := &vm.Function{...}
    
    // 2. Set up function scope with parameters
    c.pushScope(true)
    for _, param := range decl.Parameters.Parameters {
        c.getOrCreateVariable(param.Name) // Assigns slots 0, 1, 2, ...
    }
    
    // 3. Compile function body
    c.compileNode(decl.Body)
    
    // 4. Store compiled function
    c.functions[funcName] = function
    
    // 5. Emit declaration instruction
    c.emit(opcodes.OP_DECLARE_FUNCTION, opcodes.IS_CONST, nameConstant, ...)
}
```

### 2. Function Call Compilation

```go
func (c *Compiler) compileFunctionCall(expr *ast.CallExpression) error {
    // 1. Compile callee (function name)
    c.compileNode(expr.Callee)
    
    // 2. Initialize function call
    c.emit(opcodes.OP_INIT_FCALL, opcodes.IS_TMP_VAR, calleeResult, 
           opcodes.IS_CONST, c.addConstant(values.NewInt(numArgs)), ...)
    
    // 3. Send arguments
    for i, arg := range expr.Arguments.Arguments {
        c.compileNode(arg)
        c.emit(opcodes.OP_SEND_VAL, opcodes.IS_CONST, argNum, 
               opcodes.IS_TMP_VAR, argResult, ...)
    }
    
    // 4. Execute call
    c.emit(opcodes.OP_DO_FCALL, opcodes.IS_TMP_VAR, result, ...)
}
```

## Bytecode Instructions

### PHP Compatibility

Our instructions match PHP's Zend VM opcodes:

| Our Opcode | PHP Zend Opcode | Purpose |
|------------|-----------------|---------|
| `OP_DECLARE_FUNCTION` | `ZEND_DECLARE_FUNCTION` | Register function at runtime |
| `OP_INIT_FCALL` | `ZEND_INIT_FCALL` | Initialize function call |
| `OP_SEND_VAL` | `ZEND_SEND_VAL` | Pass argument value |
| `OP_DO_FCALL` | `ZEND_DO_FCALL` | Execute function call |
| `OP_RETURN` | `ZEND_RETURN` | Return value from function |

### Example Bytecode Sequence

For `foo(5)`:

```
[0000] DECLARE_FUNCTION CONST:0    # Register function "foo"
[0001] INIT_FCALL TMP:100, CONST:1 # Initialize call to "foo", 1 arg
[0002] SEND_VAL CONST:2, TMP:101   # Send argument: 5
[0003] DO_FCALL TMP:102            # Execute call, store result in TMP:102
```

## Key Design Principles

1. **PHP Compatibility**: Follows PHP's exact function call semantics
2. **Isolation**: Each function call gets its own execution context
3. **Parameter Mapping**: Arguments map to parameter variable slots (0, 1, 2, ...)
4. **Return Handling**: Functions can return any PHP value type
5. **Error Handling**: Proper error messages for missing functions or parameters

## Testing and Validation

The implementation is tested with:
- Simple function calls with parameters and return values
- Nested function calls
- Functions returning arrays for use in foreach loops
- Parameter validation and error handling

This architecture ensures full PHP compatibility while maintaining clean separation between compilation and execution phases.