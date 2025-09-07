package jit

import (
	"fmt"
	"runtime"
	"syscall"
	"unsafe"
)

// ExecutableMemory represents executable memory region
type ExecutableMemory struct {
	Data []byte  // The actual memory region
	Size int     // Size of the memory region
	ptr  uintptr // Raw pointer to the memory
}

// AllocateExecutableMemory allocates executable memory using platform-specific system calls
func AllocateExecutableMemory(size int) (*ExecutableMemory, error) {
	switch runtime.GOOS {
	case "linux", "darwin":
		return allocateExecutableMemoryUnix(size)
	case "windows":
		return allocateExecutableMemoryWindows(size)
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// allocateExecutableMemoryUnix allocates executable memory on Unix-like systems using mmap
func allocateExecutableMemoryUnix(size int) (*ExecutableMemory, error) {
	// Round up to page size
	pageSize := syscall.Getpagesize()
	alignedSize := ((size + pageSize - 1) / pageSize) * pageSize

	// Use mmap to allocate memory with PROT_READ | PROT_WRITE | PROT_EXEC
	ptr, _, errno := syscall.Syscall6(
		syscall.SYS_MMAP,
		0,                    // addr
		uintptr(alignedSize), // length
		syscall.PROT_READ|syscall.PROT_WRITE|syscall.PROT_EXEC, // prot
		syscall.MAP_PRIVATE|syscall.MAP_ANONYMOUS,              // flags
		0, // fd
		0, // offset
	)

	if ptr == ^uintptr(0) || errno != 0 { // MAP_FAILED is typically (void*)-1
		return nil, fmt.Errorf("mmap failed: %v", errno)
	}

	// Create a Go slice backed by the allocated memory
	data := (*[1 << 30]byte)(unsafe.Pointer(ptr))[:alignedSize:alignedSize]

	return &ExecutableMemory{
		Data: data,
		Size: alignedSize,
		ptr:  ptr,
	}, nil
}

// allocateExecutableMemoryWindows allocates executable memory on Windows using VirtualAlloc
func allocateExecutableMemoryWindows(size int) (*ExecutableMemory, error) {
	// On Windows, we would use VirtualAlloc with PAGE_EXECUTE_READWRITE
	// For now, return an error as this requires Windows-specific syscalls
	return nil, fmt.Errorf("Windows executable memory allocation not yet implemented")
}

// Free releases the executable memory
func (em *ExecutableMemory) Free() error {
	if em.ptr == 0 {
		return nil // Already freed or never allocated
	}

	switch runtime.GOOS {
	case "linux", "darwin":
		return em.freeUnix()
	case "windows":
		return em.freeWindows()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// freeUnix releases memory on Unix-like systems using munmap
func (em *ExecutableMemory) freeUnix() error {
	_, _, errno := syscall.Syscall(
		syscall.SYS_MUNMAP,
		em.ptr,
		uintptr(em.Size),
		0,
	)

	if errno != 0 {
		return fmt.Errorf("munmap failed: %v", errno)
	}

	em.ptr = 0
	em.Data = nil
	return nil
}

// freeWindows releases memory on Windows
func (em *ExecutableMemory) freeWindows() error {
	// Would use VirtualFree on Windows
	return fmt.Errorf("Windows memory deallocation not yet implemented")
}

// WriteBytes writes bytes to the executable memory
func (em *ExecutableMemory) WriteBytes(offset int, data []byte) error {
	if offset+len(data) > len(em.Data) {
		return fmt.Errorf("write would exceed memory bounds")
	}

	copy(em.Data[offset:], data)
	return nil
}

// GetFunctionPointer returns a function pointer for the given offset
func (em *ExecutableMemory) GetFunctionPointer(offset int) uintptr {
	return em.ptr + uintptr(offset)
}

// ExecutionContext represents the execution context for JIT-compiled code
type JITExecutionContext struct {
	// PHP value stack for passing arguments and receiving return values
	ValueStack []*interface{} // Stack of PHP values
	StackPtr   int            // Stack pointer

	// Register state (for debugging and context switching)
	Registers map[string]uint64 // Register values

	// Error handling
	LastError error

	// VM callback functions
	VMCallbacks *VMCallbacks
}

// VMCallbacks provides callbacks to VM functionality from JIT code
type VMCallbacks struct {
	// Callback for unsupported operations that need VM fallback
	VMFallback func(opcode int, args []interface{}) (interface{}, error)

	// Memory management callbacks
	AllocateValue func() interface{}
	FreeValue     func(interface{})

	// Type conversion callbacks
	ToInt    func(interface{}) (int64, error)
	ToFloat  func(interface{}) (float64, error)
	ToString func(interface{}) (string, error)
	ToBool   func(interface{}) (bool, error)
}

// NewJITExecutionContext creates a new JIT execution context
func NewJITExecutionContext() *JITExecutionContext {
	return &JITExecutionContext{
		ValueStack:  make([]*interface{}, 256), // Pre-allocate stack
		StackPtr:    0,
		Registers:   make(map[string]uint64),
		VMCallbacks: &VMCallbacks{},
	}
}

// PushValue pushes a value onto the JIT execution stack
func (ctx *JITExecutionContext) PushValue(value *interface{}) error {
	if ctx.StackPtr >= len(ctx.ValueStack) {
		return fmt.Errorf("JIT stack overflow")
	}

	ctx.ValueStack[ctx.StackPtr] = value
	ctx.StackPtr++
	return nil
}

// PopValue pops a value from the JIT execution stack
func (ctx *JITExecutionContext) PopValue() (*interface{}, error) {
	if ctx.StackPtr <= 0 {
		return nil, fmt.Errorf("JIT stack underflow")
	}

	ctx.StackPtr--
	value := ctx.ValueStack[ctx.StackPtr]
	ctx.ValueStack[ctx.StackPtr] = nil // Clear reference
	return value, nil
}

// SetRegister sets a register value (for debugging/inspection)
func (ctx *JITExecutionContext) SetRegister(name string, value uint64) {
	ctx.Registers[name] = value
}

// GetRegister gets a register value (for debugging/inspection)
func (ctx *JITExecutionContext) GetRegister(name string) uint64 {
	return ctx.Registers[name]
}
