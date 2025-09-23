# PHP References Implementation Specification

## Overview

PHP references allow two variables to point to the same data, enabling changes to one variable to affect the other. This document outlines the implementation strategy for supporting PHP references in the Hey-Codex interpreter.

## Current State Analysis

### Existing Infrastructure

1. **AST Support**: Already defined in `/compiler/ast/kind.go`
   - `ASTRef` (kind 281) - For reference expressions
   - `ASTAssignRef` (kind 519) - For reference assignments

2. **Token Support**: Already defined in `/compiler/lexer/token.go`
   - `T_AMPERSAND_FOLLOWED_BY_VAR_OR_VARARG` - For `&$var` or `&...`
   - `T_AMPERSAND_NOT_FOLLOWED_BY_VAR_OR_VARARG` - For bitwise AND

3. **Value System**: Already defined in `/values/value.go`
   - `TypeReference` type
   - `Reference` struct with `Target *Value`
   - `NewReference()` constructor
   - `IsReference()` helper method

4. **Opcode Support**: Already defined in `/opcodes/opcodes.go`
   - `OP_ASSIGN_REF` - For reference assignment

5. **Parameter Support**: Already tracked in AST and registry
   - `ByReference` field in `ParameterNode`
   - `IsReference` field in `registry.Parameter`

### Missing Components

1. **Parser**: No proper handling of `=&` assignment operator
2. **Compiler**: No generation of reference-related opcodes
3. **VM**: No implementation of `OP_ASSIGN_REF` instruction
4. **Function calls**: No handling of reference parameters
5. **Return by reference**: No implementation

## Implementation Plan

### Phase 1: Reference Assignment (`$b = &$a`)

#### Parser Changes
1. Add parsing for `=&` operator in `parseAssignmentExpression`
2. Create `AssignRefNode` when encountering reference assignment
3. Ensure proper precedence handling

#### Compiler Changes
1. Add handling for `ASTAssignRef` node type
2. Generate `OP_ASSIGN_REF` opcode
3. Handle reference tracking in symbol table

#### VM Changes
1. Implement `OP_ASSIGN_REF` instruction
2. Ensure proper reference creation and linking
3. Handle reference updates correctly

### Phase 2: Function Parameter References

#### Parser Changes
1. Already supports `ByReference` flag in parameters
2. Ensure proper parsing of `&$param` syntax

#### Compiler Changes
1. Track reference parameters in function signatures
2. Generate appropriate opcodes for reference parameter passing

#### VM Changes
1. Modify function call handling to pass references
2. Ensure parameter modifications affect original variables
3. Handle reference counting and cleanup

### Phase 3: Return by Reference

#### Parser Changes
1. Already supports `ByReference` flag in function declarations
2. Ensure proper parsing of `function &getName()`

#### Compiler Changes
1. Track return-by-reference in function metadata
2. Generate appropriate return opcodes

#### VM Changes
1. Handle reference returns in `OP_RETURN`
2. Ensure proper reference chain maintenance

### Phase 4: Advanced Features

1. **Global references**: `global $var`
2. **Static references**: `static &$var`
3. **Array element references**: `$ref = &$arr[0]`
4. **Object property references**: `$ref = &$obj->prop`
5. **Foreach with references**: `foreach ($arr as &$val)`

## Reference Semantics

### Core Behavior

1. **Assignment by Reference**
   ```php
   $a = 10;
   $b = &$a;  // $b now references $a
   $b = 20;   // Both $a and $b are now 20
   ```

2. **Reference Chains**
   ```php
   $a = 1;
   $b = &$a;
   $c = &$b;  // All three reference the same value
   ```

3. **Unset Behavior**
   ```php
   $a = 10;
   $b = &$a;
   unset($b); // Only removes $b, $a remains
   ```

### Reference Counting

- Each reference maintains a count of variables pointing to it
- When count reaches 0, the reference can be garbage collected
- Circular references need special handling

## Testing Strategy

### Unit Tests
1. Basic reference assignment
2. Reference chains
3. Function parameter references
4. Return by reference
5. Array element references
6. Foreach with references
7. Unset behavior
8. Global and static references

### Integration Tests
1. Complex reference scenarios
2. Performance benchmarks
3. Memory usage validation
4. Edge cases and error conditions

## Implementation Order

1. **Week 1**: Basic reference assignment (`$b = &$a`)
2. **Week 2**: Function parameter references
3. **Week 3**: Return by reference
4. **Week 4**: Array and object references
5. **Week 5**: Foreach and advanced features
6. **Week 6**: Testing and optimization

## Technical Decisions

### Reference Implementation Model

**Option 1: Direct Pointer Model** (Chosen)
- References store direct pointers to target values
- Simple and efficient for basic cases
- May need special handling for complex scenarios

**Option 2: Reference ID Model**
- Each reference gets a unique ID
- All variables with same ID share value
- More complex but handles all cases uniformly

### Memory Management

- Use Go's garbage collection for basic cleanup
- Implement reference counting for circular reference detection
- Consider weak references for special cases

## Success Criteria

1. All PHP reference test cases pass
2. Performance within 10% of native PHP for reference operations
3. Memory usage comparable to native PHP
4. No memory leaks in reference handling
5. Full compatibility with PHP 8.0+ reference semantics

## Status

- [x] Analysis complete
- [x] Specification written
- [ ] Parser implementation
- [ ] Compiler implementation
- [ ] VM implementation
- [ ] Testing
- [ ] Optimization