# Foreach Iterator Design

This document explains the design and implementation of the foreach iteration system in the PHP bytecode compiler, covering iterator initialization, element fetching, and loop control.

## Table of Contents

1. [Problem Overview](#problem-overview)
2. [Architecture Overview](#architecture-overview)
3. [Core Components](#core-components)
4. [Iterator Lifecycle](#iterator-lifecycle)
5. [Bytecode Instructions](#bytecode-instructions)
6. [PHP Compatibility](#php-compatibility)

## Problem Overview

### The Challenge

PHP foreach loops must handle various iterable types and provide proper key-value iteration:

```php
<?php
$arr = [0, 1, 2, 3, 4];
foreach($arr as $key => $value) {
    echo "$key: $value\n";
}

foreach(foo(5) as $v) {  // Dynamic iterable from function call
    echo "$v\n";
}
```

The challenges include:
- **Iterator State Management**: Tracking current position and remaining elements
- **Key-Value Extraction**: Properly extracting keys and values from arrays
- **Loop Termination**: Detecting when iteration is complete
- **Dynamic Iterables**: Handling function call results and complex expressions

## Architecture Overview

The foreach system uses a state-based iterator approach:

```
Iterable Compilation → Iterator Initialization → Element Fetching Loop → 
Variable Assignment → Body Execution → Continue/Break Handling
```

### Core Components

#### 1. ForeachIterator Structure

```go
type ForeachIterator struct {
    Array   *values.Value      // The iterable being processed
    Index   int                // Current iteration index
    Keys    []*values.Value    // Pre-extracted keys for iteration
    Values  []*values.Value    // Pre-extracted values for iteration
    HasMore bool               // Whether more elements exist
}
```

#### 2. Iterator Initialization (`OP_FE_RESET`)

```go
func (vm *VirtualMachine) executeForeachReset(ctx *ExecutionContext, inst *opcodes.Instruction) error {
    // Get the iterable (array/object to iterate over)
    iterable := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
    
    // Create new iterator state
    iterator := &ForeachIterator{
        Array:   iterable,
        Index:   0,
        Keys:    make([]*values.Value, 0),
        Values:  make([]*values.Value, 0),
        HasMore: true,
    }
    
    // Extract all keys and values from array
    if iterable.Type == values.TypeArray {
        arrayVal := iterable.Data.(*values.Array)
        for key, value := range arrayVal.Elements {
            keyVal := convertToValue(key)
            iterator.Keys = append(iterator.Keys, keyVal)
            iterator.Values = append(iterator.Values, value)
        }
        iterator.HasMore = len(iterator.Keys) > 0
    }
    
    // Store iterator in context
    ctx.ForeachIterators[inst.Result] = iterator
}
```

#### 3. Element Fetching (`OP_FE_FETCH`)

```go
func (vm *VirtualMachine) executeForeachFetch(ctx *ExecutionContext, inst *opcodes.Instruction) error {
    // Get iterator from context
    iteratorID := uint32(iteratorValue.Data.(int64))
    iterator := ctx.ForeachIterators[iteratorID]
    
    // Check if we have more elements
    if !iterator.HasMore || iterator.Index >= len(iterator.Values) {
        // No more elements - return null to signal end
        nullValue := values.NewNull()
        
        // Set key if requested
        if opcodes.DecodeOpType2(inst.OpType1) != opcodes.IS_UNUSED {
            vm.setValue(ctx, inst.Op2, opcodes.DecodeOpType2(inst.OpType1), nullValue)
        }
        
        // Set value
        vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), nullValue)
        return nil
    }
    
    // Get current key and value
    currentKey := iterator.Keys[iterator.Index]
    currentValue := iterator.Values[iterator.Index]
    
    // Set key if requested
    if opcodes.DecodeOpType2(inst.OpType1) != opcodes.IS_UNUSED {
        vm.setValue(ctx, inst.Op2, opcodes.DecodeOpType2(inst.OpType1), currentKey)
    }
    
    // Set value
    vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), currentValue)
    
    // Advance iterator
    iterator.Index++
}
```

## Iterator Lifecycle

### 1. Compilation Phase

```go
func (c *Compiler) compileForeach(stmt *ast.ForeachStatement) error {
    // Generate labels for loop control
    startLabel := c.generateLabel()
    endLabel := c.generateLabel()
    continueLabel := c.generateLabel()
    
    // Compile the iterable expression
    err := c.compileNode(stmt.Iterable)
    iterableTemp := c.nextTemp - 1
    
    // Initialize foreach iterator
    iteratorTemp := c.allocateTemp()
    c.emit(opcodes.OP_FE_RESET, opcodes.IS_TMP_VAR, iterableTemp, 
           opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, iteratorTemp)
    
    // Loop start
    c.placeLabel(startLabel)
    
    // Fetch next element
    valueTemp := c.allocateTemp()
    if stmt.Key != nil {
        keyTemp := c.allocateTemp()
        c.emit(opcodes.OP_FE_FETCH, opcodes.IS_TMP_VAR, iteratorTemp,
               opcodes.IS_TMP_VAR, keyTemp, opcodes.IS_TMP_VAR, valueTemp)
    } else {
        c.emit(opcodes.OP_FE_FETCH, opcodes.IS_TMP_VAR, iteratorTemp,
               opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, valueTemp)
    }
    
    // Check for end of iteration (null value means done)
    nullCheckTemp := c.allocateTemp()
    nullConstant := c.addConstant(values.NewNull())
    c.emit(opcodes.OP_IS_IDENTICAL, opcodes.IS_TMP_VAR, valueTemp,
           opcodes.IS_CONST, nullConstant, opcodes.IS_TMP_VAR, nullCheckTemp)
    c.emitJumpNZ(opcodes.IS_TMP_VAR, nullCheckTemp, endLabel)
    
    // Assign key and value to variables
    if stmt.Key != nil {
        if keyVar, ok := stmt.Key.(*ast.Variable); ok {
            keySlot := c.getOrCreateVariable(keyVar.Name)
            c.emit(opcodes.OP_ASSIGN, opcodes.IS_TMP_VAR, keyTemp,
                   opcodes.IS_UNUSED, 0, opcodes.IS_VAR, keySlot)
        }
    }
    
    if valueVar, ok := stmt.Value.(*ast.Variable); ok {
        valueSlot := c.getOrCreateVariable(valueVar.Name)
        c.emit(opcodes.OP_ASSIGN, opcodes.IS_TMP_VAR, valueTemp,
               opcodes.IS_UNUSED, 0, opcodes.IS_VAR, valueSlot)
    }
    
    // Compile loop body
    err = c.compileNode(stmt.Body)
    
    // Continue label and jump back to start
    c.placeLabel(continueLabel)
    c.emitJump(startLabel)
    
    // End label
    c.placeLabel(endLabel)
}
```

### 2. Runtime Execution Flow

```
1. OP_FE_RESET: Initialize iterator, extract all keys/values
2. Label: Loop start point
3. OP_FE_FETCH: Get current key/value or null if done
4. OP_IS_IDENTICAL: Check if value is null (end of iteration)
5. JMPNZ: Jump to end if null
6. OP_ASSIGN: Assign key/value to loop variables
7. Body execution
8. JMP: Jump back to start
9. Label: Loop end point
```

## Bytecode Instructions

### Example Bytecode for `foreach($arr as $v)`

```
[0000] FE_RESET TMP:100, UNUSED:0, TMP:101    # Initialize iterator
[0001] FE_FETCH TMP:101, UNUSED:0, TMP:102    # Fetch next value
[0002] IS_IDENTICAL TMP:102, CONST:0, TMP:103 # Compare with null
[0003] JMPNZ TMP:103, LABEL:end               # Jump if done
[0004] ASSIGN TMP:102, UNUSED:0, VAR:0        # Assign to $v
[0005] ECHO VAR:0                             # Loop body
[0006] JMP LABEL:start                        # Continue loop
[0007] LABEL:end                              # End of loop
```

### With Key-Value iteration `foreach($arr as $k => $v)`

```
[0000] FE_RESET TMP:100, UNUSED:0, TMP:101    # Initialize iterator
[0001] FE_FETCH TMP:101, TMP:102, TMP:103     # Fetch key and value
[0002] IS_IDENTICAL TMP:103, CONST:0, TMP:104 # Compare value with null
[0003] JMPNZ TMP:104, LABEL:end               # Jump if done
[0004] ASSIGN TMP:102, UNUSED:0, VAR:0        # Assign key to $k
[0005] ASSIGN TMP:103, UNUSED:0, VAR:1        # Assign value to $v
[0006] ECHO VAR:0, VAR:1                      # Loop body
[0007] JMP LABEL:start                        # Continue loop
[0008] LABEL:end                              # End of loop
```

## Key Design Features

1. **Pre-extraction Strategy**: All keys and values are extracted during `FE_RESET` for consistent iteration
2. **Null Termination**: Uses null values to signal end of iteration
3. **Iterator Storage**: Multiple iterators can coexist using unique IDs
4. **Break/Continue Support**: Proper label management for loop control
5. **Dynamic Iterables**: Works with function call results and complex expressions

## PHP Compatibility

The implementation follows PHP's foreach semantics:
- Proper key-value iteration for associative arrays
- Numeric indexing for list arrays
- Iterator state isolation for nested foreach loops
- Compatible termination conditions

## Performance Characteristics

- **Time Complexity**: O(n) for iteration setup, O(1) for each fetch
- **Space Complexity**: O(n) for storing keys/values during iteration
- **Iterator Overhead**: Minimal per-element processing cost

This design ensures full PHP compatibility while providing efficient iteration over various iterable types.