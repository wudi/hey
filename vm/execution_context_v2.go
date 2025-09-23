package vm

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// ExecutionContextV2 is the refactored execution context with separated concerns
type ExecutionContextV2 struct {
	// Separated managers for different concerns
	Variables *VariableManager
	Classes   *ClassManager
	CallStack *CallStackManager

	// File inclusion tracking
	IncludedFiles *sync.Map // map[string]bool

	// User symbols can use regular maps with RWMutex since they're mostly read-only after setup
	userSymbolsMu    sync.RWMutex
	UserFunctions    map[string]*registry.Function
	UserClasses      map[string]*registry.Class
	UserInterfaces   map[string]*registry.Interface
	UserTraits       map[string]*registry.Trait

	// Execution state
	Stack        []*values.Value
	OutputWriter io.Writer
	Halted       bool
	ExitCode     int
	Constants    []*values.Value

	// Debug support
	debugMu  sync.Mutex
	debugLog []string
}

// NewExecutionContextV2 constructs a fresh execution context with separated concerns
func NewExecutionContextV2() *ExecutionContextV2 {
	return &ExecutionContextV2{
		Variables:      NewVariableManager(),
		Classes:        NewClassManager(),
		CallStack:      NewCallStackManager(),
		IncludedFiles:  &sync.Map{},
		Stack:          make([]*values.Value, 0, 16),
		OutputWriter:   os.Stdout,
		UserFunctions:  make(map[string]*registry.Function),
		UserClasses:    make(map[string]*registry.Class),
		UserInterfaces: make(map[string]*registry.Interface),
		UserTraits:     make(map[string]*registry.Trait),
		debugLog:       make([]string, 0, 64),
	}
}

// SetOutputWriter allows callers to redirect the script output stream
func (ctx *ExecutionContextV2) SetOutputWriter(w io.Writer) {
	if w == nil {
		return
	}
	ctx.OutputWriter = w
}

// Delegation methods for backward compatibility

// pushFrame adds a new call frame to the call stack
func (ctx *ExecutionContextV2) pushFrame(frame *CallFrame) {
	ctx.CallStack.PushFrame(frame)
}

// popFrame removes and returns the current call frame
func (ctx *ExecutionContextV2) popFrame() *CallFrame {
	return ctx.CallStack.PopFrame()
}

// currentFrame returns the actively executing call frame
func (ctx *ExecutionContextV2) currentFrame() *CallFrame {
	return ctx.CallStack.CurrentFrame()
}

// appendDebugRecord records an entry for later inspection via debug reports
func (ctx *ExecutionContextV2) appendDebugRecord(record string) {
	ctx.debugMu.Lock()
	defer ctx.debugMu.Unlock()
	ctx.debugLog = append(ctx.debugLog, record)
}

// drainDebugRecords returns the accumulated debug log
func (ctx *ExecutionContextV2) drainDebugRecords() []string {
	ctx.debugMu.Lock()
	defer ctx.debugMu.Unlock()
	out := make([]string, len(ctx.debugLog))
	copy(out, ctx.debugLog)
	return out
}

// ensureClass ensures a class exists, creating it if necessary
func (ctx *ExecutionContextV2) ensureClass(name string) *classRuntime {
	return ctx.Classes.EnsureClass(name, ctx.UserClasses)
}

// getClass retrieves a class runtime
func (ctx *ExecutionContextV2) getClass(name string) (*classRuntime, bool) {
	return ctx.Classes.GetClass(name)
}

// ensureGlobal ensures a global variable exists
func (ctx *ExecutionContextV2) ensureGlobal(name string) *values.Value {
	return ctx.Variables.EnsureGlobal(name)
}

// setVariable sets a variable
func (ctx *ExecutionContextV2) setVariable(name string, value *values.Value) {
	ctx.Variables.SetVariable(name, value)
}

// unsetVariable removes a variable
func (ctx *ExecutionContextV2) unsetVariable(name string) {
	ctx.Variables.UnsetVariable(name)
}

// bindGlobalValue binds a global variable
func (ctx *ExecutionContextV2) bindGlobalValue(name string, value *values.Value) {
	ctx.Variables.BindGlobalValue(name, value)
	// Update call stack bindings
	ctx.CallStack.UpdateGlobalBindings(globalNameVariants(name), value)
}

// unsetGlobal removes a global variable
func (ctx *ExecutionContextV2) unsetGlobal(name string) {
	ctx.Variables.UnsetGlobal(name)
}

// setTemporary sets a temporary variable
func (ctx *ExecutionContextV2) setTemporary(slot uint32, value *values.Value) {
	ctx.Variables.SetTemporary(slot, value)
}

// GetGlobalVar safely retrieves a global variable
func (ctx *ExecutionContextV2) GetGlobalVar(name string) (*values.Value, bool) {
	return ctx.Variables.GetGlobalVar(name)
}

// SetGlobalVar safely sets a global variable
func (ctx *ExecutionContextV2) SetGlobalVar(name string, val *values.Value) {
	ctx.Variables.SetGlobalVar(name, val)
}

// GetUserFunction safely retrieves a user function
func (ctx *ExecutionContextV2) GetUserFunction(name string) (*registry.Function, bool) {
	ctx.userSymbolsMu.RLock()
	defer ctx.userSymbolsMu.RUnlock()
	fn, ok := ctx.UserFunctions[name]
	return fn, ok
}

// GetUserClass safely retrieves a user class
func (ctx *ExecutionContextV2) GetUserClass(name string) (*registry.Class, bool) {
	ctx.userSymbolsMu.RLock()
	defer ctx.userSymbolsMu.RUnlock()
	cls, ok := ctx.UserClasses[name]
	return cls, ok
}

// IsFileIncluded safely checks if a file is included
func (ctx *ExecutionContextV2) IsFileIncluded(path string) bool {
	if val, ok := ctx.IncludedFiles.Load(path); ok {
		return val.(bool)
	}
	return false
}

// MarkFileIncluded safely marks a file as included
func (ctx *ExecutionContextV2) MarkFileIncluded(path string) {
	ctx.IncludedFiles.Store(path, true)
}

// GetVariable safely retrieves a variable
func (ctx *ExecutionContextV2) GetVariable(name string) (*values.Value, bool) {
	return ctx.Variables.GetVariable(name)
}

// GetTemporary safely retrieves a temporary variable
func (ctx *ExecutionContextV2) GetTemporary(slot uint32) (*values.Value, bool) {
	return ctx.Variables.GetTemporary(slot)
}

// exportState exports frame state back to context
func (ctx *ExecutionContextV2) exportState(frame *CallFrame) {
	if frame == nil {
		return
	}

	// Export constants (no locking needed for slice)
	ctx.Constants = frame.cloneConstants()

	// Clear and repopulate Temporaries
	ctx.Variables.Temporaries = &sync.Map{}
	for slot, val := range frame.TempVars {
		ctx.Variables.Temporaries.Store(slot, val)
	}

	// Clear and repopulate Variables
	ctx.Variables.Variables = &sync.Map{}
	for slot, val := range frame.Locals {
		if name, ok := frame.SlotNames[slot]; ok && name != "" {
			ctx.Variables.Variables.Store(name, val)
		} else {
			ctx.Variables.Variables.Store(fmt.Sprintf("$%d", slot), val)
		}
	}
}

// recordAssignment is invoked whenever a local variable changes
func (ctx *ExecutionContextV2) recordAssignment(frame *CallFrame, slot uint32, value *values.Value) {
	if frame == nil {
		return
	}
	name, ok := frame.SlotNames[slot]
	if !ok {
		return
	}
	ctx.appendDebugRecord(fmt.Sprintf("assign %s = %s", name, value.String()))
}

// Access to embedded fields for compatibility
func (ctx *ExecutionContextV2) GetCurrentClass() *classRuntime {
	return ctx.Classes.GetCurrentClass()
}

func (ctx *ExecutionContextV2) SetCurrentClass(cls *classRuntime) {
	ctx.Classes.SetCurrentClass(cls)
}

func (ctx *ExecutionContextV2) ClearCurrentClass() {
	ctx.Classes.ClearCurrentClass()
}