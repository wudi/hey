# Hey-Codex Development Task Specification

## Overview

This document provides a comprehensive specification for upcoming development tasks in the Hey-Codex PHP interpreter project. Tasks are prioritized by criticality and impact on PHP compatibility.

## Task Priority Classification

- **CRITICAL**: Breaks existing functionality, blocks development
- **HIGH**: Essential for PHP compatibility, major missing features
- **MEDIUM**: Important improvements, performance optimizations
- **LOW**: Nice-to-have features, code quality improvements

## Development Tasks

### 1. Fix Broken Runtime Tests 🚨 CRITICAL
**Status**: Pending
**Priority**: CRITICAL
**File**: `/runtime/array_test.go`
**Issue**: `mockBuiltinContext` missing `GetHTTPContext()` method implementation
**Impact**: All runtime tests are currently broken, blocking development

**Acceptance Criteria**:
- [ ] Update `mockBuiltinContext` to implement complete `BuiltinCallContext` interface
- [ ] All runtime tests pass successfully
- [ ] No test compilation errors

**Dependencies**: None
**Estimate**: 30 minutes

### 2. Fix Broken SPL Tests 🚨 CRITICAL
**Status**: Pending
**Priority**: CRITICAL
**File**: `/runtime/spl/` (multiple test files)
**Issue**: SPL tests failing to build due to interface changes
**Impact**: SPL functionality cannot be properly tested

**Acceptance Criteria**:
- [ ] All SPL tests compile successfully
- [ ] Update test mocks to match current interfaces
- [ ] All existing SPL tests pass

**Dependencies**: Task #1 (runtime test fixes)
**Estimate**: 45 minutes

### 3. Implement SPL Observer Pattern 🔥 HIGH
**Status**: Pending
**Priority**: HIGH
**File**: `/runtime/spl/observer_pattern.go` (new)
**Issue**: Missing `SplObserver` and `SplSubject` interfaces
**Impact**: Foundation for PHP design patterns

**Acceptance Criteria**:
- [ ] Implement `SplObserver` interface with `update()` method
- [ ] Implement `SplSubject` interface with `attach()`, `detach()`, `notify()` methods
- [ ] Add comprehensive tests
- [ ] Update SPL function registry

**Dependencies**: Task #2 (SPL test fixes)
**Estimate**: 2 hours

### 4. Implement SplHeap Data Structures 🔥 HIGH
**Status**: Pending
**Priority**: HIGH
**Files**: `/runtime/spl/spl_heap.go`, `/runtime/spl/spl_max_heap.go`, `/runtime/spl/spl_min_heap.go`, `/runtime/spl/spl_priority_queue.go`
**Issue**: Missing essential heap data structures
**Impact**: Core data structures for algorithm implementations

**Components**:
- `SplHeap` (abstract base)
- `SplMaxHeap` (max heap implementation)
- `SplMinHeap` (min heap implementation)
- `SplPriorityQueue` (priority queue with heap backing)

**Acceptance Criteria**:
- [ ] Implement all four heap classes with proper inheritance
- [ ] Support standard heap operations (insert, extract, top, count, isEmpty)
- [ ] Maintain heap property invariants
- [ ] Add comprehensive tests for all operations
- [ ] Register classes in SPL registry

**Dependencies**: Task #3 (Observer pattern for consistency)
**Estimate**: 4 hours

### 5. Implement File System Iterators 📁 HIGH
**Status**: Pending
**Priority**: HIGH
**Files**: Multiple files in `/runtime/spl/`
**Issue**: Missing critical file iteration capabilities
**Impact**: Essential for file operations and directory traversal

**Components**:
- `DirectoryIterator` - Simple directory iteration
- `FilesystemIterator` - Enhanced directory iteration with more options
- `GlobIterator` - Pattern-based file matching
- `RecursiveDirectoryIterator` - Recursive directory traversal

**Acceptance Criteria**:
- [ ] Implement all iterator classes with proper interfaces
- [ ] Support standard iterator methods (current, key, next, rewind, valid)
- [ ] Handle file system errors gracefully
- [ ] Add comprehensive tests including edge cases
- [ ] Ensure cross-platform compatibility

**Dependencies**: Task #4 (SPL foundation)
**Estimate**: 6 hours

### 6. Complete Generator Implementation ⚡ MEDIUM
**Status**: Pending
**Priority**: MEDIUM
**File**: `/compiler/compiler.go`, `/vm/instructions.go`
**Issue**: Incomplete yield/resume logic for PHP generators
**Impact**: Modern PHP feature for memory-efficient iteration

**Acceptance Criteria**:
- [ ] Implement proper yield opcode handling
- [ ] Support generator resume/send/throw operations
- [ ] Handle generator state management
- [ ] Add comprehensive tests
- [ ] Update compiler to emit correct bytecode

**Dependencies**: Task #5 (iterator foundation)
**Estimate**: 3 hours

### 7. Implement preg_replace() Function 🔍 MEDIUM
**Status**: Pending
**Priority**: MEDIUM
**File**: `/runtime/regex.go`
**Issue**: Last missing regex function to complete the set
**Impact**: Completes regex function implementation

**Acceptance Criteria**:
- [ ] Implement `preg_replace()` with all parameter variants
- [ ] Support replacement patterns and callbacks
- [ ] Handle limit and count parameters
- [ ] Add comprehensive tests with real PHP validation
- [ ] Update documentation as complete

**Dependencies**: None (can be done in parallel)
**Estimate**: 2 hours

### 8. Complete Trait System 🎭 MEDIUM
**Status**: Pending
**Priority**: MEDIUM
**File**: `/compiler/compiler.go` (line 7147)
**Issue**: Missing trait adaptation handling (precedence and alias rules)
**Impact**: Complete PHP trait system support

**Acceptance Criteria**:
- [ ] Handle trait method precedence resolution
- [ ] Implement trait method aliasing
- [ ] Support trait conflict resolution
- [ ] Add comprehensive tests for trait adaptations
- [ ] Update compiler to handle all trait syntax

**Dependencies**: Task #6 (advanced language features)
**Estimate**: 4 hours

### 9. Implement Regex Pattern Caching 🚀 MEDIUM
**Status**: Pending
**Priority**: MEDIUM
**File**: `/runtime/regex.go`
**Issue**: Patterns recompiled on each use
**Impact**: Significant performance improvement for regex operations

**Acceptance Criteria**:
- [ ] Implement LRU cache for compiled regex patterns
- [ ] Configurable cache size
- [ ] Thread-safe cache operations
- [ ] Performance benchmarks showing improvement
- [ ] Memory usage monitoring

**Dependencies**: Task #7 (regex completion)
**Estimate**: 2 hours

### 10. Fix include_once and require_once Opcodes 📥 MEDIUM
**Status**: Pending
**Priority**: MEDIUM
**File**: `/compiler/compiler.go`, `/vm/instructions.go`
**Issue**: Opcodes not yet implemented, test skipped
**Impact**: Complete include system implementation

**Acceptance Criteria**:
- [ ] Implement missing opcodes for include_once/require_once
- [ ] Update compiler to emit correct bytecode
- [ ] Handle file tracking for "once" semantics
- [ ] Add comprehensive tests
- [ ] Unskip related tests

**Dependencies**: Task #8 (language features)
**Estimate**: 2 hours

## Implementation Order

1. **Phase 1: Critical Fixes** (Tasks 1-2) - Fix broken tests first
2. **Phase 2: SPL Foundation** (Tasks 3-5) - Build SPL infrastructure
3. **Phase 3: Language Features** (Tasks 6, 8, 10) - Complete language support
4. **Phase 4: Optimizations** (Tasks 7, 9) - Performance and completion

## Testing Strategy

- All tasks must include comprehensive test coverage
- Use TDD approach where applicable
- Validate against real PHP behavior
- Include edge case testing
- Performance testing for optimization tasks

## Success Metrics

- All tests passing
- No build failures
- Performance benchmarks (where applicable)
- PHP compatibility validation
- Code coverage maintenance

## Notes

- Tasks are designed to be completed sequentially within phases
- Each task includes specific acceptance criteria
- Dependencies are clearly marked to avoid blocking issues
- Time estimates are conservative and include testing