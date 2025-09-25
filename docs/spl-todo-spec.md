# SPL (Standard PHP Library) Implementation TODO Specification

## Overview
This document provides a detailed implementation plan for missing SPL components in Hey-Codex, prioritized by importance and usage patterns in real-world PHP applications.

## Current Status Summary
✅ **Completed (23/36 major components - 64%)**:
- Core Interfaces: Iterator, ArrayAccess, Countable, IteratorAggregate, OuterIterator, RecursiveIterator, SeekableIterator
- All SPL Exceptions (13 exception classes)
- Data Structures: SplDoublyLinkedList, SplStack, SplQueue, SplFixedArray, ArrayObject, SplObjectStorage
- Basic Iterators: ArrayIterator
- Advanced Iterators: AppendIterator, CachingIterator, CallbackFilterIterator, EmptyIterator, FilterIterator, InfiniteIterator, IteratorIterator, LimitIterator, MultipleIterator, NoRewindIterator, RecursiveArrayIterator, RecursiveIteratorIterator, RegexIterator
- File Handling: SplFileInfo
- SPL Functions: iterator_apply(), iterator_count(), iterator_to_array(), spl_classes(), spl_object_hash(), spl_object_id(), class_implements(), class_parents(), class_uses()

❌ **Missing (13 major components - 36%)**:

## Phase 1: High Priority Missing Components (Critical for PHP Compatibility)

### 1.1 Observer Pattern Interfaces
**Status**: ❌ Not Implemented
**Priority**: High (Design Pattern Foundation)

#### SplObserver Interface
```php
interface SplObserver {
    public function update(SplSubject $subject): void;
}
```

#### SplSubject Interface
```php
interface SplSubject {
    public function attach(SplObserver $observer): void;
    public function detach(SplObserver $observer): void;
    public function notify(): void;
}
```

**Implementation Tasks**:
- [ ] Create `observer_pattern.go` with interface definitions
- [ ] Add registration in `spl.go`
- [ ] Create comprehensive unit tests with PHP validation
- [ ] Document usage patterns

### 1.2 SplHeap Family (Data Structure Foundation)
**Status**: ❌ Not Implemented
**Priority**: High (Essential Data Structures)

#### SplHeap (Abstract Base Class)
```php
abstract class SplHeap implements Iterator, Countable {
    abstract protected function compare(mixed $value1, mixed $value2): int;
    public function count(): int;
    public function current(): mixed;
    public function extract(): mixed;
    public function insert(mixed $value): bool;
    public function isEmpty(): bool;
    public function key(): int;
    public function next(): void;
    public function recoverFromCorruption(): bool;
    public function rewind(): void;
    public function top(): mixed;
    public function valid(): bool;
}
```

#### SplMaxHeap extends SplHeap
```php
class SplMaxHeap extends SplHeap {
    protected function compare(mixed $value1, mixed $value2): int;
}
```

#### SplMinHeap extends SplHeap
```php
class SplMinHeap extends SplHeap {
    protected function compare(mixed $value1, mixed $value2): int;
}
```

#### SplPriorityQueue extends SplHeap
```php
class SplPriorityQueue extends SplHeap {
    public function compare(mixed $priority1, mixed $priority2): int;
    public function current(): mixed;
    public function extract(): mixed;
    public function insert(mixed $value, mixed $priority): bool;
    public function recoverFromCorruption(): bool;
    public function setExtractFlags(int $flags): int;
}
```

**Implementation Tasks**:
- [ ] Create `spl_heap.go` with abstract base class
- [ ] Create `spl_max_heap.go` implementing max heap behavior
- [ ] Create `spl_min_heap.go` implementing min heap behavior
- [ ] Create `spl_priority_queue.go` with priority queue functionality
- [ ] Implement heap data structure using Go slices
- [ ] Add comprehensive unit tests for all heap operations
- [ ] Create PHP validation scripts for heap behavior
- [ ] Document performance characteristics

### 1.3 File System Iterators
**Status**: ❌ Not Implemented
**Priority**: High (Essential for File Operations)

#### DirectoryIterator
```php
class DirectoryIterator implements SeekableIterator {
    public function __construct(string $directory);
    public function current(): DirectoryIterator;
    public function getATime(): int;
    public function getBasename(string $suffix = ""): string;
    public function getCTime(): int;
    public function getExtension(): string;
    public function getFilename(): string;
    public function getGroup(): int;
    public function getInode(): int;
    public function getMTime(): int;
    public function getOwner(): int;
    public function getPath(): string;
    public function getPathname(): string;
    public function getPerms(): int;
    public function getSize(): int;
    public function getType(): string;
    public function isDir(): bool;
    public function isDot(): bool;
    public function isExecutable(): bool;
    public function isFile(): bool;
    public function isLink(): bool;
    public function isReadable(): bool;
    public function isWritable(): bool;
    public function key(): int;
    public function next(): void;
    public function rewind(): void;
    public function seek(int $offset): void;
    public function valid(): bool;
}
```

#### FilesystemIterator extends DirectoryIterator
```php
class FilesystemIterator extends DirectoryIterator {
    public const CURRENT_MODE_MASK = 240;
    public const CURRENT_AS_PATHNAME = 32;
    public const CURRENT_AS_FILEINFO = 0;
    public const CURRENT_AS_SELF = 16;
    public const KEY_MODE_MASK = 3840;
    public const KEY_AS_PATHNAME = 0;
    public const KEY_AS_FILENAME = 256;
    public const FOLLOW_SYMLINKS = 16384;
    public const KEY_AS_HASH = 1024;
    public const NEW_CURRENT_AND_KEY = 256;
    public const OTHER_MODE_MASK = 28672;
    public const SKIP_DOTS = 4096;
    public const UNIX_PATHS = 8192;

    public function __construct(string $directory, int $flags = FilesystemIterator::KEY_AS_PATHNAME | FilesystemIterator::CURRENT_AS_FILEINFO | FilesystemIterator::SKIP_DOTS);
    public function current(): string|SplFileInfo|FilesystemIterator;
    public function getFlags(): int;
    public function key(): string;
    public function next(): void;
    public function rewind(): void;
    public function setFlags(int $flags): void;
}
```

#### GlobIterator extends FilesystemIterator
```php
class GlobIterator extends FilesystemIterator {
    public function __construct(string $pattern, int $flags = 0);
    public function count(): int;
}
```

#### RecursiveDirectoryIterator extends FilesystemIterator implements RecursiveIterator
```php
class RecursiveDirectoryIterator extends FilesystemIterator implements RecursiveIterator {
    public const FOLLOW_SYMLINKS = 16384;

    public function __construct(string $directory, int $flags = 0);
    public function getChildren(): RecursiveDirectoryIterator;
    public function getSubPath(): string;
    public function getSubPathname(): string;
    public function hasChildren(bool $allowLinks = false): bool;
    public function key(): string;
    public function next(): void;
    public function rewind(): void;
}
```

**Implementation Tasks**:
- [ ] Create `directory_iterator.go` with basic directory traversal
- [ ] Create `filesystem_iterator.go` extending DirectoryIterator with flags
- [ ] Create `glob_iterator.go` with glob pattern matching
- [ ] Create `recursive_directory_iterator.go` with recursive traversal
- [ ] Implement cross-platform file system operations
- [ ] Add comprehensive file system tests
- [ ] Create PHP validation scripts
- [ ] Handle permissions, symlinks, and special files

## Phase 2: Medium Priority Missing Components

### 2.1 Advanced File Handling
**Status**: ❌ Not Implemented
**Priority**: Medium (File I/O Operations)

#### SplFileObject extends SplFileInfo
```php
class SplFileObject extends SplFileInfo implements RecursiveIterator, SeekableIterator {
    public const DROP_NEW_LINE = 1;
    public const READ_AHEAD = 2;
    public const SKIP_EMPTY = 4;
    public const READ_CSV = 8;

    public function __construct(string $filename, string $mode = 'r', bool $useIncludePath = false, ?resource $context = null);
    public function current(): string|array|false;
    public function eof(): bool;
    public function fflush(): bool;
    public function fgetc(): string|false;
    public function fgetcsv(string $separator = ",", string $enclosure = "\"", string $escape = "\\"): array|false;
    public function fgets(): string;
    public function fgetss(string $allowable_tags = ?): string;
    public function flock(int $operation, int &$wouldBlock = null): bool;
    public function fpassthru(): int;
    public function fputcsv(array $fields, string $separator = ",", string $enclosure = "\"", string $escape = "\\"): int|false;
    public function fread(int $length): string|false;
    public function fscanf(string $format, mixed ...$vars): array|int|null;
    public function fseek(int $offset, int $whence = SEEK_SET): int;
    public function fstat(): array;
    public function ftell(): int|false;
    public function ftruncate(int $size): bool;
    public function fwrite(string $str, int $length = 0): int|false;
    public function getChildren(): ?RecursiveIterator;
    public function getCsvControl(): array;
    public function getFlags(): int;
    public function getMaxLineLen(): int;
    public function hasChildren(): bool;
    public function key(): int;
    public function next(): void;
    public function rewind(): void;
    public function seek(int $line): void;
    public function setCsvControl(string $separator = ",", string $enclosure = "\"", string $escape = "\\"): void;
    public function setFlags(int $flags): void;
    public function setMaxLineLen(int $maxLength): void;
    public function valid(): bool;
}
```

#### SplTempFileObject extends SplFileObject
```php
class SplTempFileObject extends SplFileObject {
    public function __construct(int $maxMemory = 2 * 1024 * 1024);
}
```

**Implementation Tasks**:
- [ ] Create `spl_file_object.go` with file I/O operations
- [ ] Create `spl_temp_file_object.go` with temporary file handling
- [ ] Implement CSV parsing functionality
- [ ] Add file locking mechanisms
- [ ] Create comprehensive file I/O tests
- [ ] Handle binary vs text mode differences
- [ ] Cross-platform file operations

### 2.2 Advanced Recursive Iterators
**Status**: ❌ Not Implemented
**Priority**: Medium (Specialized Use Cases)

#### RecursiveCachingIterator extends CachingIterator implements RecursiveIterator
```php
class RecursiveCachingIterator extends CachingIterator implements RecursiveIterator {
    public function getChildren(): RecursiveCachingIterator;
    public function hasChildren(): bool;
}
```

#### RecursiveCallbackFilterIterator extends CallbackFilterIterator implements RecursiveIterator
```php
class RecursiveCallbackFilterIterator extends CallbackFilterIterator implements RecursiveIterator {
    public function getChildren(): RecursiveCallbackFilterIterator;
    public function hasChildren(): bool;
}
```

#### RecursiveFilterIterator extends FilterIterator implements RecursiveIterator
```php
abstract class RecursiveFilterIterator extends FilterIterator implements RecursiveIterator {
    public function getChildren(): ?RecursiveFilterIterator;
    public function hasChildren(): bool;
}
```

#### RecursiveRegexIterator extends RegexIterator implements RecursiveIterator
```php
class RecursiveRegexIterator extends RegexIterator implements RecursiveIterator {
    public function getChildren(): RecursiveRegexIterator;
    public function hasChildren(): bool;
}
```

#### RecursiveTreeIterator extends RecursiveIteratorIterator
```php
class RecursiveTreeIterator extends RecursiveIteratorIterator {
    public const BYPASS_CURRENT = 4;
    public const BYPASS_KEY = 8;
    public const PREFIX_LEFT = 0;
    public const PREFIX_MID_HAS_NEXT = 1;
    public const PREFIX_MID_LAST = 2;
    public const PREFIX_END_HAS_NEXT = 3;
    public const PREFIX_END_LAST = 4;
    public const PREFIX_RIGHT = 5;

    public function __construct(RecursiveIterator|IteratorAggregate $iterator, int $flags = RecursiveTreeIterator::BYPASS_KEY, int $cachingIteratorFlags = CachingIterator::CALL_TOSTRING, int $mode = RecursiveIteratorIterator::SELF_FIRST);
    public function current(): string;
    public function getEntry(): string;
    public function getPostfix(): string;
    public function getPrefix(): string;
    public function key(): string;
    public function next(): void;
    public function rewind(): void;
    public function setPrefixPart(int $part, string $value): void;
    public function valid(): bool;
}
```

#### ParentIterator extends FilterIterator implements RecursiveIterator
```php
class ParentIterator extends FilterIterator implements RecursiveIterator {
    public function accept(): bool;
    public function getChildren(): ParentIterator;
    public function hasChildren(): bool;
}
```

**Implementation Tasks**:
- [ ] Create recursive versions of existing filtering iterators
- [ ] Create `recursive_tree_iterator.go` with tree display functionality
- [ ] Create `parent_iterator.go` filtering for parent elements
- [ ] Add comprehensive recursive iterator tests
- [ ] Ensure proper inheritance hierarchy

## Phase 3: Low Priority Missing Components

### 3.1 SPL Autoloader Functions
**Status**: ❌ Not Implemented
**Priority**: Low (Autoloading System)

```php
function spl_autoload(string $className, ?string $fileExtensions = null): void;
function spl_autoload_call(string $className): bool;
function spl_autoload_extensions(?string $fileExtensions = null): string;
function spl_autoload_functions(): array;
function spl_autoload_register(?callable $autoloadFunction = null, bool $throw = true, bool $prepend = false): bool;
function spl_autoload_unregister(callable $autoloadFunction): bool;
```

**Implementation Tasks**:
- [ ] Create `spl_autoload.go` with autoloading functions
- [ ] Implement autoloader registry system
- [ ] Add file extension handling
- [ ] Create autoloading tests
- [ ] Integrate with Hey-Codex class loading system

## Testing Strategy

### TDD Implementation Approach
For each missing component:

1. **Red Phase** - Create failing test
   - Write PHP validation script using native PHP
   - Capture expected behavior and edge cases
   - Create Go unit test that fails

2. **Green Phase** - Implement minimum code
   - Implement just enough to make tests pass
   - Focus on correct behavior over optimization

3. **Refactor Phase** - Optimize and clean
   - Improve performance while keeping tests green
   - Follow Go idioms and Hey-Codex patterns
   - Add comprehensive error handling

### Test Categories
- **Unit Tests**: Individual component behavior
- **Integration Tests**: Component interaction
- **PHP Validation**: Behavior matches native PHP
- **Performance Tests**: Acceptable overhead vs native
- **Edge Case Tests**: Error conditions and boundaries

### PHP Validation Scripts
Each component needs validation against native PHP:

```php
// Example validation script pattern
<?php
// Test SplMaxHeap behavior
$heap = new SplMaxHeap();
$heap->insert(1);
$heap->insert(3);
$heap->insert(2);

echo "Top: " . $heap->top() . "\n";        // Should output: 3
echo "Extract: " . $heap->extract() . "\n"; // Should output: 3
echo "Count: " . $heap->count() . "\n";     // Should output: 2

// Test edge cases
try {
    $empty = new SplMaxHeap();
    $empty->top(); // Should throw RuntimeException
} catch (RuntimeException $e) {
    echo "Expected exception: " . $e->getMessage() . "\n";
}
?>
```

## Implementation Order

### Phase 1 (Critical) - Immediate Priority
1. **SplObserver/SplSubject** - Observer pattern foundation
2. **SplHeap Family** - Essential data structures
3. **DirectoryIterator** - Basic file system traversal
4. **FilesystemIterator** - Enhanced directory operations

### Phase 2 (Important) - Next Sprint
5. **SplFileObject** - File I/O operations
6. **RecursiveDirectoryIterator** - Recursive file system
7. **GlobIterator** - Pattern-based file matching
8. **SplTempFileObject** - Temporary file handling

### Phase 3 (Nice to Have) - Future Iterations
9. **RecursiveCachingIterator** - Recursive caching
10. **RecursiveCallbackFilterIterator** - Recursive filtering
11. **RecursiveTreeIterator** - Tree visualization
12. **SPL Autoload Functions** - Autoloading system

## Success Criteria

### Functional Requirements
- ✅ All tests pass (unit, integration, PHP validation)
- ✅ Behavior matches native PHP exactly
- ✅ Proper error handling and exceptions
- ✅ Thread-safe implementations where applicable

### Performance Requirements
- ✅ Maximum 2x overhead vs native PHP operations
- ✅ Memory usage comparable to native structures
- ✅ O(n) complexity preserved for all operations

### Quality Requirements
- ✅ 100% test coverage for public APIs
- ✅ Comprehensive edge case testing
- ✅ Clear documentation and examples
- ✅ Follows Hey-Codex coding standards

## File Structure

```
runtime/spl/
├── observer_pattern.go           # SplObserver/SplSubject interfaces
├── observer_pattern_test.go
├── spl_heap.go                   # Abstract SplHeap base class
├── spl_heap_test.go
├── spl_max_heap.go               # SplMaxHeap implementation
├── spl_max_heap_test.go
├── spl_min_heap.go               # SplMinHeap implementation
├── spl_min_heap_test.go
├── spl_priority_queue.go         # SplPriorityQueue implementation
├── spl_priority_queue_test.go
├── directory_iterator.go         # DirectoryIterator implementation
├── directory_iterator_test.go
├── filesystem_iterator.go        # FilesystemIterator implementation
├── filesystem_iterator_test.go
├── glob_iterator.go              # GlobIterator implementation
├── glob_iterator_test.go
├── recursive_directory_iterator.go    # RecursiveDirectoryIterator
├── recursive_directory_iterator_test.go
├── spl_file_object.go            # SplFileObject implementation
├── spl_file_object_test.go
├── spl_temp_file_object.go       # SplTempFileObject implementation
├── spl_temp_file_object_test.go
├── recursive_*_iterator.go       # Remaining recursive iterators
├── recursive_*_iterator_test.go
├── spl_autoload.go              # Autoload functions
├── spl_autoload_test.go
└── validation/                   # PHP validation scripts
    ├── test_observer_pattern.php
    ├── test_spl_heap.php
    ├── test_directory_iterator.php
    ├── test_filesystem_iterator.php
    ├── test_spl_file_object.php
    └── ...
```

This specification provides a comprehensive roadmap for completing the SPL implementation in Hey-Codex, prioritized by real-world usage and PHP compatibility requirements.