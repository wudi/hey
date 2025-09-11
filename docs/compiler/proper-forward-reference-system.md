# Proper Forward Reference System Design

This document explains the design and implementation of the forward reference system for handling jump instructions in the PHP compiler. This system resolves the "invalid jump target" issue that occurs when compiling control flow statements like `if`, `while`, and `ternary` operators.

## Table of Contents

1. [Problem Overview](#problem-overview)
2. [System Architecture](#system-architecture)
3. [Core Components](#core-components)
4. [Implementation Details](#implementation-details)
5. [Jump Resolution Process](#jump-resolution-process)
6. [Usage Examples](#usage-examples)
7. [Testing and Validation](#testing-and-validation)

## Problem Overview

### The Challenge

When compiling control flow statements in PHP, we encounter a fundamental problem: **forward jumps**. Consider this PHP code:

```php
<?php
if ($condition) {
    echo "true";
} else {
    echo "false";  // <- Jump target not yet known during compilation
}
```

During compilation, when we process the `if` statement:

1. We compile the condition
2. We need to emit a `JMPZ` (Jump if Zero) instruction to jump to the `else` block
3. **Problem**: We don't know the instruction address of the `else` block yet!

### The Original Broken Approach

The original implementation had this flawed `addLabel` function:

```go
func (c *Compiler) addLabel(name string) uint32 {
    // WRONG: Returns current position, not future target
    return uint32(len(c.instructions)) // Placeholder
}
```

This caused:
- **Infinite loops**: Jumps targeting themselves
- **Invalid targets**: References to non-existent instructions  
- **Runtime errors**: "invalid jump target" during VM execution

## System Architecture

### High-Level Design

The proper forward reference system uses a **two-phase approach**:

1. **Compilation Phase**: Emit jump instructions with placeholder targets and record forward references
2. **Resolution Phase**: When labels are placed, resolve all recorded forward references

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Compilation   │───▶│   Recording     │───▶│   Resolution    │
│     Phase       │    │Forward References│    │     Phase       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
        │                       │                       │
        ▼                       ▼                       ▼
  Emit JMPZ with          Store forward          Update constant
  placeholder target      jump metadata          with real target
```

### Key Design Principles

1. **Separation of Concerns**: Jump emission and target resolution are separate
2. **Constant Pool Strategy**: Jump targets stored as constants, not direct operands
3. **Metadata Tracking**: Forward jumps tracked with complete resolution information
4. **Batch Resolution**: All forward references resolved when label is placed

## Core Components

### 1. ForwardJump Structure

The `ForwardJump` struct tracks unresolved jump instructions:

```go
type ForwardJump struct {
    instructionIndex int         // Constant pool index (not instruction index!)
    opType           opcodes.OpType // Always IS_CONST for new system
    operand          int         // 0 = constant update, 1/2 = instruction operand
}
```

**Key Design Decision**: `instructionIndex` stores the **constant pool index**, not the bytecode instruction index. This enables updating the constant value during resolution.

### 2. Compiler Extensions

The compiler is extended with forward reference tracking:

```go
type Compiler struct {
    // ... existing fields ...
    forwardJumps    map[string][]ForwardJump // labelName -> []ForwardJump
}
```

**Data Structure Choice**: `map[string][]ForwardJump` allows multiple jumps to the same label (e.g., multiple `break` statements in a loop).

### 3. Specialized Jump Emitters

Two specialized functions handle jump instruction emission:

#### Conditional Jumps (JMPZ)
```go
func (c *Compiler) emitJumpZ(condType opcodes.OpType, cond uint32, labelName string) {
    // Add placeholder constant for jump target
    jumpConstant := c.addConstant(values.NewInt(0)) // Will be updated later
    
    // Emit instruction referencing the constant
    c.emit(opcodes.OP_JMPZ, condType, cond, opcodes.IS_CONST, jumpConstant, 0, 0)
    
    // Record forward reference for later resolution
    jump := ForwardJump{
        instructionIndex: int(jumpConstant), // Store constant index
        opType:           opcodes.IS_CONST,
        operand:          0, // Special marker for constant update
    }
    c.forwardJumps[labelName] = append(c.forwardJumps[labelName], jump)
}
```

#### Unconditional Jumps (JMP)
```go
func (c *Compiler) emitJump(opcode opcodes.Opcode, op1Type opcodes.OpType, op1 uint32, labelName string) {
    // Similar structure to emitJumpZ
    jumpConstant := c.addConstant(values.NewInt(0))
    c.emit(opcode, op1Type, op1, opcodes.IS_CONST, jumpConstant, 0, 0)
    
    // Record forward reference
    jump := ForwardJump{
        instructionIndex: int(jumpConstant),
        opType:           opcodes.IS_CONST,
        operand:          0,
    }
    c.forwardJumps[labelName] = append(c.forwardJumps[labelName], jump)
}
```

## Implementation Details

### Step 1: Jump Instruction Emission

When the compiler encounters a conditional statement:

```go
// Example: compiling if statement
func (c *Compiler) compileIf(stmt *ast.IfStatement) error {
    // Compile condition
    startTemp := c.nextTemp
    err := c.compileNode(stmt.Test)
    condResult := startTemp

    // Generate unique labels
    elseLabel := c.generateLabel() // e.g., "L0"
    endLabel := c.generateLabel()  // e.g., "L1"

    // Emit conditional jump with forward reference
    c.emitJumpZ(opcodes.IS_TMP_VAR, condResult, elseLabel)
    
    // ... compile consequence and alternative ...
}
```

**Generated Bytecode** (before resolution):
```
Instruction[1]: JMPZ temp[1000], constant[5]  // constant[5] = int(0) placeholder
```

**Forward Jump Record**:
```go
forwardJumps["L0"] = []ForwardJump{
    {instructionIndex: 5, opType: IS_CONST, operand: 0}
}
```

### Step 2: Label Placement and Resolution

When the compiler reaches the label placement point:

```go
func (c *Compiler) placeLabel(name string) {
    pos := len(c.instructions) // Current instruction position
    c.labels[name] = pos
    
    // Resolve ALL forward jumps to this label
    if jumps, exists := c.forwardJumps[name]; exists {
        for _, jump := range jumps {
            if jump.operand == 0 {
                // Update constant value (new system)
                constantIndex := jump.instructionIndex
                c.constants[constantIndex] = values.NewInt(int64(pos))
            } else {
                // Update instruction operand (legacy system)
                instruction := &c.instructions[jump.instructionIndex]
                if jump.operand == 1 {
                    instruction.Op1 = uint32(pos)
                } else if jump.operand == 2 {
                    instruction.Op2 = uint32(pos)
                }
            }
        }
        delete(c.forwardJumps, name) // Clean up
    }
}
```

**After Resolution**:
```
constant[5] = int(8)  // Updated to actual instruction address
```

## Jump Resolution Process

### Complete Example Walkthrough

Let's trace through compiling this PHP code:

```php
<?php
if ($a === 5) {
    echo "match";
} else {
    echo "no match";
}
```

#### Phase 1: Compilation

1. **Compile condition**: `$a === 5`
   ```
   [0] FETCH_R var[13] -> temp[1000]     // Load $a
   [1] QM_ASSIGN const[0] -> temp[1001]  // Load 5  
   [2] IS_IDENTICAL temp[1000], temp[1001] -> temp[1002]
   ```

2. **Emit conditional jump**:
   ```go
   elseLabel := c.generateLabel() // "L0"
   c.emitJumpZ(opcodes.IS_TMP_VAR, 1002, elseLabel)
   ```
   
   **Generated**:
   ```
   [3] JMPZ temp[1002], const[1]  // const[1] = int(0) placeholder
   ```
   
   **Forward Reference Recorded**:
   ```go
   forwardJumps["L0"] = [{instructionIndex: 1, operand: 0}]
   ```

3. **Compile true branch**:
   ```
   [4] QM_ASSIGN const[2] -> temp[1003]  // "match"
   [5] ECHO temp[1003]
   ```

4. **Emit unconditional jump to end**:
   ```go
   endLabel := c.generateLabel() // "L1"
   c.emitJump(opcodes.OP_JMP, opcodes.IS_CONST, 0, endLabel)
   ```
   
   **Generated**:
   ```
   [6] JMP const[3] // const[3] = int(0) placeholder
   ```

5. **Place else label and compile false branch**:
   ```go
   c.placeLabel(elseLabel) // Resolves "L0" to instruction 7
   ```
   
   **Resolution**: `const[1] = int(7)`
   
   ```
   [7] QM_ASSIGN const[4] -> temp[1004]  // "no match"
   [8] ECHO temp[1004]
   ```

6. **Place end label**:
   ```go
   c.placeLabel(endLabel) // Resolves "L1" to instruction 9
   ```
   
   **Resolution**: `const[3] = int(9)`

#### Phase 2: Final Bytecode

After resolution, the complete bytecode:

```
[0] FETCH_R var[13] -> temp[1000]
[1] QM_ASSIGN const[0] -> temp[1001]     // const[0] = int(5)
[2] IS_IDENTICAL temp[1000], temp[1001] -> temp[1002]
[3] JMPZ temp[1002], const[1]            // const[1] = int(7) ✓
[4] QM_ASSIGN const[2] -> temp[1003]     // const[2] = "match"
[5] ECHO temp[1003]
[6] JMP const[3]                         // const[3] = int(9) ✓
[7] QM_ASSIGN const[4] -> temp[1004]     // const[4] = "no match"
[8] ECHO temp[1004]
[9] RETURN const[5]                      // const[5] = null
```

**Constant Pool**:
```
[0] = int(5)          // Comparison value
[1] = int(7)          // JMPZ target (else branch)
[2] = string("match") // True branch string
[3] = int(9)          // JMP target (end)
[4] = string("no match") // False branch string
[5] = null            // Return value
```

## Usage Examples

### If Statement Compilation

```go
func (c *Compiler) compileIf(stmt *ast.IfStatement) error {
    // Step 1: Compile condition
    startTemp := c.nextTemp
    err := c.compileNode(stmt.Test)
    if err != nil {
        return err
    }
    condResult := startTemp

    // Step 2: Generate labels
    elseLabel := c.generateLabel()
    endLabel := c.generateLabel()

    // Step 3: Emit forward jump
    c.emitJumpZ(opcodes.IS_TMP_VAR, condResult, elseLabel)

    // Step 4: Compile consequence
    for _, s := range stmt.Consequent {
        err = c.compileNode(s)
        if err != nil {
            return err
        }
    }

    // Step 5: Jump to end
    c.emitJump(opcodes.OP_JMP, opcodes.IS_CONST, 0, endLabel)

    // Step 6: Place else label and compile alternative
    c.placeLabel(elseLabel) // Resolves all jumps to elseLabel
    if len(stmt.Alternate) > 0 {
        for _, s := range stmt.Alternate {
            err = c.compileNode(s)
            if err != nil {
                return err
            }
        }
    }

    // Step 7: Place end label
    c.placeLabel(endLabel) // Resolves all jumps to endLabel

    return nil
}
```

### While Loop Compilation

```go
func (c *Compiler) compileWhile(stmt *ast.WhileStatement) error {
    startLabel := c.generateLabel()
    endLabel := c.generateLabel()

    // Place start label
    c.placeLabel(startLabel) // No forward reference needed

    // Compile condition
    startTemp := c.nextTemp
    err := c.compileNode(stmt.Test)
    condResult := startTemp

    // Forward jump to end if condition false
    c.emitJumpZ(opcodes.IS_TMP_VAR, condResult, endLabel)

    // Compile body
    for _, s := range stmt.Body {
        err = c.compileNode(s)
        if err != nil {
            return err
        }
    }

    // Backward jump to start (no forward reference)
    if pos, exists := c.labels[startLabel]; exists {
        jumpConstant := c.addConstant(values.NewInt(int64(pos)))
        c.emit(opcodes.OP_JMP, opcodes.IS_CONST, 0, opcodes.IS_CONST, jumpConstant, 0, 0)
    }

    // Place end label
    c.placeLabel(endLabel) // Resolves forward jumps

    return nil
}
```

## Testing and Validation

### Unit Test Structure

```go
func TestForwardReferenceSystem(t *testing.T) {
    testCases := []struct {
        name     string
        code     string
        expected string
    }{
        {
            name: "Simple If Statement",
            code: `<?php
                if (true) {
                    echo "TRUE";
                } else {
                    echo "FALSE";
                }`,
            expected: "TRUE",
        },
        {
            name: "Nested If Statements",
            code: `<?php
                if (true) {
                    if (false) {
                        echo "INNER FALSE";
                    } else {
                        echo "INNER TRUE";
                    }
                } else {
                    echo "OUTER FALSE";
                }`,
            expected: "INNER TRUE",
        },
        {
            name: "While Loop",
            code: `<?php
                $i = 0;
                while ($i < 3) {
                    echo $i;
                    $i++;
                }`,
            expected: "012",
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Compile and execute
            output := compileAndExecute(tc.code)
            assert.Equal(t, tc.expected, output)
        })
    }
}
```

### Integration Testing

The system is validated through:

1. **Existing Test Suite**: All 154 existing compiler tests must pass
2. **Control Flow Tests**: Specific tests for if/while/ternary operators
3. **Edge Cases**: Deeply nested conditions, empty blocks, complex expressions
4. **Performance Tests**: Large numbers of forward references

### Debugging and Troubleshooting

Common issues and solutions:

1. **"Invalid jump target" errors**:
   - Check constant pool bounds
   - Verify OpType encoding/decoding
   - Ensure constants are properly updated during resolution

2. **Infinite loops**:
   - Verify labels are placed at correct positions
   - Check for self-referencing jumps
   - Validate forward reference cleanup

3. **Memory leaks**:
   - Ensure `delete(c.forwardJumps, name)` after resolution
   - Clean up temporary constants if needed

## Conclusion

The proper forward reference system provides a robust foundation for compiling control flow statements in the PHP compiler. Key benefits:

- **Correctness**: Eliminates "invalid jump target" errors
- **Maintainability**: Clean separation between compilation and resolution
- **Extensibility**: Easy to add new control flow constructs
- **Performance**: Efficient batch resolution process
- **Debugging**: Clear tracking of forward references

This system enables the PHP compiler to correctly handle complex control flow patterns while maintaining clean, understandable code architecture.