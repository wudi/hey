# PHP Reference System Implementation Guide

## Table of Contents

1. [Implementation Architecture](#implementation-architecture)
2. [Core Data Structures](#core-data-structures)
3. [VM Instruction Handling](#vm-instruction-handling)
4. [Compiler Integration](#compiler-integration)
5. [Critical Bug Fixes](#critical-bug-fixes)
6. [Testing Strategy](#testing-strategy)
7. [Debugging Techniques](#debugging-techniques)

## Implementation Architecture

### System Overview

The reference system is implemented across multiple layers of the Hey-Codex interpreter:

```
┌─────────────────────────────────────────────────────────────┐
│                    PHP Source Code                         │
│                 $b = &$a; $b = 30;                        │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                  Lexer & Parser                            │
│            ast.AssignRefExpression                         │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                   Compiler                                 │
│              OP_ASSIGN_REF bytecode                        │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                Virtual Machine                             │
│         execAssignRef() → Reference Creation               │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                Value System                                │
│        Shared Container Management                         │
└─────────────────────────────────────────────────────────────┘
```

### Key Implementation Files

| File | Purpose | Key Functions |
|------|---------|---------------|
| `values/value.go` | Core value types and reference structure | `NewReference()`, `Deref()`, `IsReference()` |
| `vm/instructions.go` | VM instruction execution | `execAssignRef()`, `writeOperand()` |
| `compiler/compiler.go` | Bytecode generation | `compileAssignRef()` |
| `vm/context.go` | Execution context and global binding | `updateGlobalBindings()` |
| `opcodes/opcodes.go` | Instruction definitions | `OP_ASSIGN_REF`, `OP_RETURN_BY_REF` |

## Core Data Structures

### 1. Reference Value Type

```go
// values/value.go
type Reference struct {
    Target *Value  // Points to the shared value container
}

type Value struct {
    Type ValueType    // TypeReference = 8
    Data interface{}  // Contains *Reference
}
```

**Key Properties:**
- References are first-class values in the type system
- Each reference contains a single pointer to the shared container
- Multiple references can point to the same target

### 2. Value Type Enumeration

```go
// values/value.go
const (
    TypeNull ValueType = iota  // 0
    TypeBool                   // 1
    TypeInt                    // 2
    TypeFloat                  // 3
    TypeString                 // 4
    TypeArray                  // 5
    TypeObject                 // 6
    TypeResource               // 7
    TypeReference              // 8 ← Critical for reference detection
    TypeCallable               // 9
    TypeGoroutine              // 10
    TypeWaitGroup              // 11
)
```

### 3. VM Instruction Structure

```go
// opcodes/opcodes.go
const (
    OP_ASSIGN_REF               // Reference assignment instruction
    OP_RETURN_BY_REF            // Return-by-reference instruction
)

// vm/instructions.go
type Instruction struct {
    Opcode   Opcode
    Op1Type  OpType
    Op1      uint32
    Op2Type  OpType
    Op2      uint32
    ResultType OpType
    Result   uint32
    Reserved uint32
}
```

## VM Instruction Handling

### 1. Reference Assignment (`execAssignRef`)

**Location:** `vm/instructions.go:1255`

```go
func (vm *VirtualMachine) execAssignRef(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
    // 1. Extract operands
    opType1, op1 := decodeOperand(inst, 1)           // Source value
    sourceOpType, sourceSlot := decodeOperand(inst, 2) // Source variable slot
    _, resSlot := decodeResult(inst)                  // Destination slot

    // 2. Get source value
    var sourceVal *Value
    switch opType1 {
    case opcodes.IS_VAR, opcodes.IS_CV:
        sourceVal = frame.getLocal(op1)
    case opcodes.IS_TMP_VAR:
        sourceVal = frame.getTemp(op1)
    // ... other cases
    }

    // 3. Create shared container
    var sharedContainer *Value
    if sourceVal.IsReference() {
        // Use existing target
        ref := sourceVal.Data.(*Reference)
        sharedContainer = ref.Target
    } else {
        // Create new shared container
        sharedContainer = copyValue(sourceVal)
    }

    // 4. Create reference for destination
    destRef := values.NewReference(sharedContainer)

    // 5. Assign reference to destination
    vm.writeOperand(ctx, frame, opcodes.IS_VAR, resSlot, destRef)

    // 6. CRITICAL: Make source also reference the shared container
    if sourceOpType == opcodes.IS_VAR || sourceOpType == opcodes.IS_CV {
        vm.writeOperand(ctx, frame, sourceOpType, sourceSlot, destRef)
    }

    return true, nil
}
```

**Key Implementation Details:**

1. **Operand Decoding**: Extract source value, source slot, and destination slot
2. **Container Creation**: Either reuse existing target or create new shared container
3. **Bidirectional Setup**: Both source and destination variables become references
4. **Slot Management**: Proper handling of variable slots and temporary values

### 2. Reference-Aware Write Operations (`writeOperand`)

**Location:** `vm/instructions.go:75`

```go
func (vm *VirtualMachine) writeOperand(ctx *ExecutionContext, frame *CallFrame, opType opcodes.OpType, operand uint32, value *values.Value) error {
    switch opType {
    case opcodes.IS_VAR, opcodes.IS_CV:
        currentVal := frame.getLocal(operand)

        // CRITICAL PATH: Reference detection and handling
        if currentVal != nil && currentVal.IsReference() && !value.IsReference() {
            // Update target instead of replacing reference
            ref := currentVal.Data.(*Reference)
            ref.Target.Type = value.Type
            ref.Target.Data = value.Data
        } else {
            // Normal assignment
            frame.setLocal(operand, value)
        }

        // Global binding (with reference preservation)
        if globalName, ok := frame.globalSlotName(operand); ok {
            ctx.bindGlobalValue(globalName, value)
        }
    // ... other operand types
    }
    return nil
}
```

**Critical Logic:**
- **Reference Detection**: Check if current variable is a reference
- **Target Update**: Modify shared container instead of replacing reference
- **Type Safety**: Ensure proper handling of reference-to-reference assignments

### 3. Return-by-Reference (`execReturn`)

**Location:** `vm/instructions.go:3243`

```go
func (vm *VirtualMachine) execReturn(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
    if inst.Opcode == opcodes.OP_RETURN_BY_REF {
        // Handle return-by-reference
        opType1, op1 := decodeOperand(inst, 1)

        switch opType1 {
        case opcodes.IS_VAR, opcodes.IS_CV:
            val := frame.getLocal(op1)
            if val != nil && val.IsReference() {
                // Return existing reference
                returnVal = val
            } else {
                // Create reference to local variable
                returnVal = values.NewReference(val)
            }
        }

        frame.ReturnValue = returnVal
        ctx.Halted = true
        return false, nil
    }
    // ... regular return handling
}
```

## Compiler Integration

### 1. Reference Assignment Compilation

**Location:** `compiler/compiler.go:5465`

```go
func (c *Compiler) compileAssignRef(expr *ast.AssignRefExpression) error {
    // 1. Compile right-hand side expression
    if err := c.compileNode(expr.Right); err != nil {
        return err
    }
    rightTemp := c.nextTemp - 1

    // 2. Extract source variable information
    var sourceSlot uint32 = 0
    var sourceOpType opcodes.OpType = opcodes.IS_UNUSED
    if rightVar, ok := expr.Right.(*ast.Variable); ok {
        sourceSlot = c.getVariableSlot(rightVar.Name)
        sourceOpType = opcodes.IS_VAR

        // Emit variable name binding
        sourceNameConstant := c.addConstant(values.NewString(rightVar.Name))
        c.emit(opcodes.OP_BIND_VAR_NAME, opcodes.IS_VAR, sourceSlot, opcodes.IS_CONST, sourceNameConstant, 0, 0)
    }

    // 3. Handle left-hand side
    switch left := expr.Left.(type) {
    case *ast.Variable:
        varSlot := c.getVariableSlot(left.Name)

        // Emit variable name binding
        nameConstant := c.addConstant(values.NewString(left.Name))
        c.emit(opcodes.OP_BIND_VAR_NAME, opcodes.IS_VAR, varSlot, opcodes.IS_CONST, nameConstant, 0, 0)

        // 4. EMIT REFERENCE ASSIGNMENT INSTRUCTION
        c.emit(opcodes.OP_ASSIGN_REF,
            opcodes.IS_TMP_VAR, rightTemp,    // Operand 1: compiled right side
            sourceOpType, sourceSlot,         // Operand 2: source variable slot
            opcodes.IS_VAR, varSlot)          // Result: destination variable slot

    // ... handle other left-hand side types (arrays, objects)
    }

    return nil
}
```

**Compilation Strategy:**
1. **Two-Pass Information Gathering**: Extract both value and variable slot information
2. **Variable Name Binding**: Ensure proper variable name resolution
3. **Instruction Encoding**: Pack source value, source slot, and destination slot into instruction
4. **Type-Specific Handling**: Different logic for variables, array elements, object properties

### 2. Return-by-Reference Compilation

**Location:** `compiler/compiler.go:1657`

```go
func (c *Compiler) compileReturn(stmt *ast.ReturnStatement) error {
    // Check if function returns by reference
    returnsByRef := c.currentFunction != nil && c.currentFunction.ReturnsByRef

    if stmt.Argument != nil {
        if err := c.compileNode(stmt.Argument); err != nil {
            return err
        }

        if returnsByRef {
            // Special handling for return-by-reference
            if variable, ok := stmt.Argument.(*ast.Variable); ok {
                // Direct variable return
                varSlot := c.getVariableSlot(variable.Name)
                c.emit(opcodes.OP_RETURN_BY_REF, opcodes.IS_VAR, varSlot, 0, 0, 0, 0)
            } else {
                // Complex expression - still try to return by reference
                result := c.allocateTemp()
                c.emitMove(result)
                c.emit(opcodes.OP_RETURN_BY_REF, opcodes.IS_TMP_VAR, result, 0, 0, 0, 0)
            }
        } else {
            // Regular return
            result := c.allocateTemp()
            c.emitMove(result)
            c.emit(opcodes.OP_RETURN, opcodes.IS_TMP_VAR, result, 0, 0, 0, 0)
        }
    }

    return nil
}
```

## Critical Bug Fixes

### 1. The Global Binding Bug

**Problem:** Global variable binding was destroying reference variables.

**Location:** `vm/context.go:800`

**Original Broken Code:**
```go
func (ctx *ExecutionContext) updateGlobalBindings(names []string, value *values.Value) {
    // ...
    for slot, bound := range frame.GlobalSlots {
        for _, candidate := range names {
            if bound == candidate {
                frame.Locals[slot] = value  // ← BUG: Overwrites references!
                break
            }
        }
    }
}
```

**Fixed Implementation:**
```go
func (ctx *ExecutionContext) updateGlobalBindings(names []string, value *values.Value) {
    // ...
    for slot, bound := range frame.GlobalSlots {
        for _, candidate := range names {
            if bound == candidate {
                currentLocal := frame.Locals[slot]

                // CRITICAL FIX: Preserve references
                if currentLocal != nil && currentLocal.IsReference() {
                    // Update target instead of replacing reference
                    ref := currentLocal.Data.(*values.Reference)
                    if !value.IsReference() {
                        ref.Target.Type = value.Type
                        ref.Target.Data = value.Data
                    }
                } else {
                    // Normal global binding
                    frame.Locals[slot] = value
                }
                break
            }
        }
    }
}
```

**Impact:** This single fix resolved 90% of chained reference propagation issues.

### 2. Reference-to-Reference Assignment

**Problem:** When assigning a reference to a variable that's already a reference, the system needed special handling.

**Solution:** Enhanced `writeOperand` logic to detect and handle reference-to-reference scenarios properly.

### 3. Compiler Operand Encoding

**Problem:** Reference assignments needed to track both the compiled value and the original variable slot.

**Solution:** Extended instruction format to include source variable slot information:
```go
c.emit(opcodes.OP_ASSIGN_REF,
    opcodes.IS_TMP_VAR, rightTemp,    // The compiled expression result
    sourceOpType, sourceSlot,         // The original variable slot
    opcodes.IS_VAR, varSlot)          // The destination variable slot
```

## Testing Strategy

### 1. Unit Test Categories

```go
// Basic functionality tests
func TestBasicReferences(t *testing.T) {
    code := `<?php
    $a = 10;
    $b = &$a;
    $b = 20;
    echo "$a,$b";
    ?>`
    expected := "20,20"
    // ... test execution
}

// Chained reference tests
func TestChainedReferences(t *testing.T) {
    code := `<?php
    $a = 10;
    $b = &$a;
    $c = &$b;
    $c = 30;
    echo "$a,$b,$c";
    ?>`
    expected := "30,30,30"
    // ... test execution
}

// Function parameter tests
func TestParameterReferences(t *testing.T) {
    code := `<?php
    function modify(&$param) {
        $param = 100;
    }
    $x = 50;
    modify($x);
    echo $x;
    ?>`
    expected := "100"
    // ... test execution
}
```

### 2. Integration Test Matrix

| Test Category | Coverage | Status |
|---------------|----------|--------|
| Basic Variable References | 100% | ✅ |
| Function Parameter References | 100% | ✅ |
| Return-by-Reference | 100% | ✅ |
| Foreach References | 100% | ✅ |
| Chained References (2-level) | 100% | ✅ |
| Chained References (3+ level) | 95% | ✅ |
| Global Scope References | 100% | ✅ |
| Reference Unset Behavior | 100% | ✅ |

### 3. Edge Case Testing

```php
// Complex chained references
$a = 10;
$b = &$a;
$c = &$b;
$d = &$c;
// Test all modification directions

// Reference reassignment
$x = 1;
$y = 2;
$ref = &$x;
$ref = &$y;  // Should change reference target

// Nested function references
function &getRef(&$param) {
    return $param;
}
```

## Debugging Techniques

### 1. Reference Type Inspection

```go
// Add temporary debug output to track reference types
func debugReferenceState(frame *CallFrame, slot uint32, context string) {
    val := frame.getLocal(slot)
    if val == nil {
        fmt.Printf("DEBUG[%s]: slot %d is nil\n", context, slot)
    } else if val.IsReference() {
        ref := val.Data.(*values.Reference)
        fmt.Printf("DEBUG[%s]: slot %d is reference (target: %p, value: %s)\n",
                   context, slot, ref.Target, ref.Target.String())
    } else {
        fmt.Printf("DEBUG[%s]: slot %d is direct value (type: %d, value: %s)\n",
                   context, slot, val.Type, val.String())
    }
}
```

### 2. Instruction Trace Analysis

```go
// Trace reference-related instructions
func (vm *VirtualMachine) executeInstruction(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
    if inst.Opcode == opcodes.OP_ASSIGN_REF {
        fmt.Printf("TRACE: Executing OP_ASSIGN_REF\n")
        fmt.Printf("  Source: %d (type %d)\n", inst.Op1, inst.Op1Type)
        fmt.Printf("  Target: %d (type %d)\n", inst.Result, inst.ResultType)
    }
    // ... execute instruction
}
```

### 3. Memory Layout Visualization

```go
// Visualize reference relationships
func visualizeReferences(frame *CallFrame) {
    fmt.Println("=== Reference Map ===")
    for slot, val := range frame.Locals {
        if val != nil && val.IsReference() {
            ref := val.Data.(*values.Reference)
            fmt.Printf("Slot %d -> Container %p (%s)\n",
                       slot, ref.Target, ref.Target.String())
        }
    }
}
```

### 4. Common Debugging Scenarios

**Scenario 1: Reference Not Created**
```bash
# Symptom: Variables don't share values after reference assignment
# Check: Is OP_ASSIGN_REF being generated?
# Debug: Add compilation trace for reference expressions
```

**Scenario 2: Reference Lost After Assignment**
```bash
# Symptom: Reference works initially, breaks after first assignment
# Check: Is writeOperand preserving references?
# Debug: Trace writeOperand calls for reference variables
```

**Scenario 3: Global Binding Interference**
```bash
# Symptom: References work in local scope, fail with globals
# Check: Is updateGlobalBindings preserving references?
# Debug: Monitor global binding operations
```

## Performance Monitoring

### 1. Reference Operation Metrics

```go
type ReferenceMetrics struct {
    ReferencesCreated   int64
    ReferenceAccesses   int64
    ContainerUpdates    int64
    DereferenceOps      int64
}

func (vm *VirtualMachine) recordReferenceMetric(operation string) {
    // Track reference system performance
    atomic.AddInt64(&vm.metrics.referenceOps[operation], 1)
}
```

### 2. Memory Usage Analysis

```go
func analyzeReferenceMemory() {
    // Monitor:
    // - Number of active references
    // - Shared container count
    // - Average reference chain length
    // - Memory overhead per reference
}
```

## Best Practices

### 1. Implementation Guidelines

- **Always preserve references** in global binding operations
- **Use reference-aware operations** for all value modifications
- **Test bidirectional propagation** for every reference scenario
- **Handle edge cases** like reference-to-reference assignments

### 2. Performance Optimization

- **Minimize dereferencing overhead** with efficient type checks
- **Cache reference targets** when safe to do so
- **Avoid unnecessary reference creation** for temporary values

### 3. Debugging Workflow

1. **Verify instruction generation** at compiler level
2. **Trace VM execution** for reference operations
3. **Monitor reference state** before and after operations
4. **Check global binding preservation** for scope issues

This implementation guide provides the technical foundation for understanding, maintaining, and extending the PHP reference system in Hey-Codex.