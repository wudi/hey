# Integration Overview: Foreach with Function Calls

This document provides an integrated view of how the function call system and foreach iterator system work together to implement complex PHP patterns like `foreach(function_call() as $value)`.

## Table of Contents

1. [Integration Challenge](#integration-challenge)
2. [Complete Flow Analysis](#complete-flow-analysis)
3. [Bytecode Sequence](#bytecode-sequence)
4. [Execution Timeline](#execution-timeline)
5. [System Interactions](#system-interactions)
6. [Testing Strategy](#testing-strategy)

## Integration Challenge

### The Complete Test Case

```php
<?php
function foo($n): array {
    $ret = [];
    for($i=0; $i<$n; $i++) {
        $ret[] = $i;
    }
    return $ret;
}

foreach(foo(5) as $v) {
    echo "$v\n";
}
```

Expected output:
```
0
1
2
3
4
```

This involves the integration of:
- **Function Declaration**: Compiling and registering `foo()`
- **Function Call**: Executing `foo(5)` with parameter passing
- **Array Operations**: Building arrays with `$ret[] = $i`
- **Return Values**: Returning the array from `foo()`
- **Foreach Iteration**: Using the returned array as an iterable

## Complete Flow Analysis

### Phase 1: Compilation

1. **Function Declaration Compilation**:
   ```go
   // Process function foo($n): array
   function := &vm.Function{
       Name: "foo",
       Parameters: []Parameter{{Name: "$n", Type: "array"}},
       Instructions: compiled_body_bytecode,
       Constants: compiled_constants,
   }
   c.functions["foo"] = function
   ```

2. **Foreach Statement Compilation**:
   ```go
   // Compile foreach(foo(5) as $v)
   
   // 2a. Compile the iterable expression: foo(5)
   c.compileNode(callExpression)  // Produces function call bytecode
   iterableTemp := c.nextTemp - 1
   
   // 2b. Initialize foreach iterator
   iteratorTemp := c.allocateTemp()
   c.emit(OP_FE_RESET, IS_TMP_VAR, iterableTemp, IS_TMP_VAR, iteratorTemp)
   
   // 2c. Generate loop with FE_FETCH, variable assignment, body execution
   ```

### Phase 2: Runtime Execution

1. **Function Declaration Registration**:
   ```go
   // OP_DECLARE_FUNCTION: Register foo in VM context
   ctx.Functions["foo"] = compiled_function
   ```

2. **Function Call Execution**:
   ```go
   // OP_INIT_FCALL: Initialize call to foo with 1 argument
   ctx.CallContext = &CallContext{
       FunctionName: "foo",
       Arguments: [],
       NumArgs: 1,
   }
   
   // OP_SEND_VAL: Send argument 5
   ctx.CallContext.Arguments = append(args, values.NewInt(5))
   
   // OP_DO_FCALL: Execute foo(5)
   functionCtx := NewExecutionContext()
   functionCtx.Variables[0] = values.NewInt(5)  // Map $n parameter
   vm.Execute(functionCtx, foo_instructions, foo_constants, functions)
   result = functionCtx.Stack[top]  // Get returned array
   ```

3. **Foreach Iterator Setup**:
   ```go
   // OP_FE_RESET: Initialize iterator with returned array
   iterator := &ForeachIterator{
       Array: result,  // The array returned from foo(5)
       Keys: [0, 1, 2, 3, 4],
       Values: [0, 1, 2, 3, 4],
       Index: 0,
       HasMore: true,
   }
   ctx.ForeachIterators[iteratorID] = iterator
   ```

4. **Foreach Loop Execution**:
   ```go
   // Loop iterations:
   // OP_FE_FETCH: Get next value (0, then 1, then 2, etc.)
   // OP_ASSIGN: Assign to $v variable
   // OP_ECHO: Output value
   // JMP: Loop back for next iteration
   ```

## Bytecode Sequence

### Complete Bytecode for the Test Case

```assembly
# Function Declaration Phase
[0000] DECLARE_FUNCTION CONST:0           # Register "foo"

# Foreach Setup Phase  
[0001] QM_ASSIGN CONST:1, TMP:100         # Load "foo" function name
[0002] INIT_FCALL TMP:100, CONST:2        # Initialize call foo(1 arg)
[0003] QM_ASSIGN CONST:3, TMP:101         # Load argument 5
[0004] SEND_VAL CONST:4, TMP:101          # Send argument 5
[0005] DO_FCALL TMP:102                   # Execute foo(5) → array result

# Foreach Loop Phase
[0006] FE_RESET TMP:102, TMP:103          # Initialize iterator with array
[0007] FE_FETCH TMP:103, TMP:104          # Fetch next value
[0008] IS_IDENTICAL TMP:104, CONST:5, TMP:105  # Check if null (done)
[0009] JMPNZ TMP:105, LABEL:end           # Jump to end if done
[0010] ASSIGN TMP:104, VAR:0              # Assign to $v
[0011] INTERPOLATED_STRING VAR:0, CONST:6, TMP:106  # Create "$v\n"
[0012] ECHO TMP:106                       # Output value
[0013] JMP LABEL:start                    # Loop back
[0014] LABEL:end                          # End of loop

# Function foo() Bytecode (separate context)
[F000] INIT_ARRAY TMP:200                 # $ret = []
[F001] ASSIGN TMP:200, VAR:1              # Store in $ret variable
[F002] QM_ASSIGN CONST:F0, VAR:2          # $i = 0
[F003] FETCH_R VAR:2, TMP:201             # Load $i
[F004] FETCH_R VAR:0, TMP:202             # Load $n
[F005] IS_SMALLER TMP:201, TMP:202, TMP:203  # $i < $n
[F006] JMPZ TMP:203, LABEL:F_end          # Exit loop if false
[F007] FETCH_R VAR:1, TMP:204             # Load $ret
[F008] FETCH_R VAR:2, TMP:205             # Load $i  
[F009] ADD_ARRAY_ELEMENT UNUSED, TMP:205, TMP:204  # $ret[] = $i
[F010] POST_INC VAR:2                     # $i++
[F011] JMP LABEL:F_start                  # Loop back
[F012] LABEL:F_end                        # End for loop
[F013] FETCH_R VAR:1, TMP:206             # Load $ret for return
[F014] RETURN TMP:206                     # Return $ret
```

## Execution Timeline

### Step-by-Step Execution

1. **T1: Function Declaration**
   - `DECLARE_FUNCTION`: Register `foo` in `ctx.Functions`
   - Function bytecode and metadata stored for later use

2. **T2: Function Call Setup**
   - `INIT_FCALL`: Create `CallContext` for `foo` with 1 argument
   - `SEND_VAL`: Add value `5` to `CallContext.Arguments`

3. **T3: Function Execution**
   - `DO_FCALL`: Create new `ExecutionContext` for `foo`
   - Map argument `5` to parameter slot 0 (`$n`)
   - Execute function bytecode:
     - Initialize empty array `$ret`
     - For loop: `$i` from 0 to 4
     - Array push: `$ret[] = $i` for each iteration
     - Return `$ret` containing `[0, 1, 2, 3, 4]`

4. **T4: Foreach Iterator Setup**
   - `FE_RESET`: Extract keys `[0,1,2,3,4]` and values `[0,1,2,3,4]`
   - Store iterator state in `ctx.ForeachIterators`

5. **T5: Foreach Loop Iterations**
   - **Iteration 1**: `FE_FETCH` returns `0`, assign to `$v`, echo "0\n"
   - **Iteration 2**: `FE_FETCH` returns `1`, assign to `$v`, echo "1\n" 
   - **Iteration 3**: `FE_FETCH` returns `2`, assign to `$v`, echo "2\n"
   - **Iteration 4**: `FE_FETCH` returns `3`, assign to `$v`, echo "3\n"
   - **Iteration 5**: `FE_FETCH` returns `4`, assign to `$v`, echo "4\n"
   - **Iteration 6**: `FE_FETCH` returns `null`, loop exits

## System Interactions

### 1. Function System → Foreach System

```go
// Function call result feeds into foreach iterator
functionResult := executeDoFunctionCall(...)  // Returns PHP array
foreachIterator := executeForeachReset(functionResult)  // Creates iterator
```

### 2. Execution Context Isolation

```go
// Main context
mainCtx := NewExecutionContext()
mainCtx.Functions = compiledFunctions

// Function context (isolated)
functionCtx := NewExecutionContext() 
functionCtx.Functions = mainCtx.Functions  // Shared function table
functionCtx.Variables[0] = argument_value  // Isolated variables

// Return to main context
mainCtx.Temporaries[resultSlot] = functionResult
```

### 3. Memory Management

```go
// Function call creates temporary context
functionCtx := NewExecutionContext()
defer functionCtx.cleanup()  // Cleanup after execution

// Iterator state persists in main context
mainCtx.ForeachIterators[iteratorID] = iterator  // Persistent across loop
```

## Testing Strategy

### 1. Component Testing

```go
func TestFunctionDeclaration(t *testing.T) {
    // Test: Function compilation and registration
}

func TestSimpleFunctionCall(t *testing.T) {
    // Test: Function call with parameters and return value
}

func TestSimpleForeach(t *testing.T) {
    // Test: Foreach with static array
}
```

### 2. Integration Testing

```go
func TestForeachWithFunctionCall(t *testing.T) {
    // Test: Complete integration of function call + foreach
    code := `<?php
    function foo($n):array {
        $ret = [];
        for($i=0; $i<$n; $i++) {
            $ret[] = $i;
        }
        return $ret;
    }

    foreach(foo(5) as $v) {
        echo "$v\n";
    }`
    
    // Expected: "0\n1\n2\n3\n4\n"
}
```

### 3. Error Cases

```go
func TestMissingFunction(t *testing.T) {
    // Test: foreach(missing_function() as $v)
}

func TestNonArrayReturn(t *testing.T) {
    // Test: foreach(function_returning_null() as $v)
}
```

## Key Integration Points

1. **Compiler Integration**: Function and foreach compilation must work together
2. **Runtime Integration**: Function results must be compatible with foreach iterators  
3. **Context Integration**: Proper isolation and sharing of execution contexts
4. **Memory Integration**: Efficient cleanup and resource management
5. **Error Integration**: Consistent error handling across systems

This integrated architecture ensures that complex PHP patterns work correctly while maintaining the modularity and testability of individual components.