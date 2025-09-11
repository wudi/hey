package compiler

import (
	"fmt"

	"github.com/wudi/hey/opcodes"
	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// CompileContext represents the compilation context for a scope
// It contains all intermediate compilation state and has a parent chain for scoping
type CompileContext struct {
	// Parent context for scope chain (nil for global context)
	Parent *CompileContext

	// Scope-specific compilation state
	Variables  map[string]uint32              // variable name -> slot
	Constants  []*values.Value                // constant pool for this context
	Functions  map[string]*registry.Function  // functions defined in this scope
	Classes    map[string]*registry.Class     // classes defined in this scope
	Interfaces map[string]*registry.Interface // interfaces defined in this scope
	Traits     map[string]*registry.Trait     // traits defined in this scope

	// Compilation state
	Instructions []*opcodes.Instruction   // bytecode instructions for this context
	Labels       map[string]int           // label name -> instruction index
	ForwardJumps map[string][]ForwardJump // unresolved forward jumps

	// Scope metadata
	ScopeType ScopeType // type of scope (global, function, class, block)
	NextSlot  uint32    // next available variable slot
	NextTemp  uint32    // next temporary variable counter
	NextLabel int       // next label counter

	// Control flow labels for break/continue
	BreakLabel    string
	ContinueLabel string

	// Current compilation context
	CurrentClass    *registry.Class    // currently compiling class (nil if not in class)
	CurrentFunction *registry.Function // currently compiling function (nil if not in function)
}

// ScopeType represents the type of compilation scope
type ScopeType int

const (
	ScopeGlobal ScopeType = iota
	ScopeFunction
	ScopeClass
	ScopeBlock
	ScopeMethod
)

// NewCompileContext creates a new compilation context with optional parent
func NewCompileContext(parent *CompileContext) *CompileContext {
	ctx := &CompileContext{
		Parent:       parent,
		Variables:    make(map[string]uint32),
		Constants:    make([]*values.Value, 0),
		Functions:    make(map[string]*registry.Function),
		Classes:      make(map[string]*registry.Class),
		Interfaces:   make(map[string]*registry.Interface),
		Traits:       make(map[string]*registry.Trait),
		Instructions: make([]*opcodes.Instruction, 0),
		Labels:       make(map[string]int),
		ForwardJumps: make(map[string][]ForwardJump),
		ScopeType:    ScopeBlock, // default to block scope
		NextSlot:     0,
		NextTemp:     1000, // start temp vars at 1000 to avoid conflicts
		NextLabel:    0,
	}

	// If no parent, this is the global context
	if parent == nil {
		ctx.ScopeType = ScopeGlobal
	}

	return ctx
}

// GetVariable looks up a variable in the current context or parent chain
func (ctx *CompileContext) GetVariable(name string) (uint32, bool) {
	// Check current context first
	if slot, exists := ctx.Variables[name]; exists {
		return slot, true
	}

	// Check parent chain
	if ctx.Parent != nil {
		return ctx.Parent.GetVariable(name)
	}

	return 0, false
}

// GetOrCreateVariable gets an existing variable or creates a new one in current context
func (ctx *CompileContext) GetOrCreateVariable(name string) uint32 {
	// Check if variable exists in current context
	if slot, exists := ctx.Variables[name]; exists {
		return slot
	}

	// Create new variable in current context
	slot := ctx.NextSlot
	ctx.Variables[name] = slot
	ctx.NextSlot++
	return slot
}

// GetFunction looks up a function in the current context or parent chain
func (ctx *CompileContext) GetFunction(name string) (*registry.Function, bool) {
	// Check current context first
	if fn, exists := ctx.Functions[name]; exists {
		return fn, true
	}

	// Check parent chain
	if ctx.Parent != nil {
		return ctx.Parent.GetFunction(name)
	}

	return nil, false
}

// GetClass looks up a class in the current context or parent chain
func (ctx *CompileContext) GetClass(name string) (*registry.Class, bool) {
	// Check current context first
	if class, exists := ctx.Classes[name]; exists {
		return class, true
	}

	// Check parent chain
	if ctx.Parent != nil {
		return ctx.Parent.GetClass(name)
	}

	return nil, false
}

// GetInterface looks up an interface in the current context or parent chain
func (ctx *CompileContext) GetInterface(name string) (*registry.Interface, bool) {
	// Check current context first
	if iface, exists := ctx.Interfaces[name]; exists {
		return iface, true
	}

	// Check parent chain
	if ctx.Parent != nil {
		return ctx.Parent.GetInterface(name)
	}

	return nil, false
}

// GetTrait looks up a trait in the current context or parent chain
func (ctx *CompileContext) GetTrait(name string) (*registry.Trait, bool) {
	// Check current context first
	if trait, exists := ctx.Traits[name]; exists {
		return trait, true
	}

	// Check parent chain
	if ctx.Parent != nil {
		return ctx.Parent.GetTrait(name)
	}

	return nil, false
}

// AddConstant adds a constant to the context and returns its index
func (ctx *CompileContext) AddConstant(value *values.Value) uint32 {
	ctx.Constants = append(ctx.Constants, value)
	return uint32(len(ctx.Constants) - 1)
}

// EmitInstruction adds an instruction to the current context
func (ctx *CompileContext) EmitInstruction(opcode opcodes.Opcode, op1Type opcodes.OpType, op1 uint32, op2Type opcodes.OpType, op2 uint32, resultType opcodes.OpType, result uint32) {
	opType1, opType2 := opcodes.EncodeOpTypes(op1Type, op2Type, resultType)
	instruction := &opcodes.Instruction{
		Opcode:  opcode,
		OpType1: opType1,
		OpType2: opType2,
		Op1:     op1,
		Op2:     op2,
		Result:  result,
	}
	ctx.Instructions = append(ctx.Instructions, instruction)
}

// GetNextTemp returns the next temporary variable counter and increments it
func (ctx *CompileContext) GetNextTemp() uint32 {
	temp := ctx.NextTemp
	ctx.NextTemp++
	return temp
}

// GetNextLabel returns the next label counter and increments it
func (ctx *CompileContext) GetNextLabel() string {
	label := ctx.NextLabel
	ctx.NextLabel++
	return fmt.Sprintf("L%d", label)
}

// PlaceLabel sets a label at the current instruction position
func (ctx *CompileContext) PlaceLabel(label string) {
	ctx.Labels[label] = len(ctx.Instructions)
}

// GetLabelPosition returns the instruction index for a label
func (ctx *CompileContext) GetLabelPosition(label string) (int, bool) {
	pos, exists := ctx.Labels[label]
	return pos, exists
}

// AddForwardJump adds a forward jump that needs to be resolved later
func (ctx *CompileContext) AddForwardJump(label string, instructionIndex int, opcode opcodes.Opcode, operand int) {
	if ctx.ForwardJumps[label] == nil {
		ctx.ForwardJumps[label] = make([]ForwardJump, 0)
	}
	ctx.ForwardJumps[label] = append(ctx.ForwardJumps[label], ForwardJump{
		instructionIndex: instructionIndex,
		opType:           opcodes.OpType(opcode),
		operand:          operand,
	})
}

// SetCurrentClass sets the current class being compiled
func (ctx *CompileContext) SetCurrentClass(class *registry.Class) {
	ctx.CurrentClass = class
}

// SetCurrentFunction sets the current function being compiled
func (ctx *CompileContext) SetCurrentFunction(function *registry.Function) {
	ctx.CurrentFunction = function
}

// GetRootContext returns the root (global) context
func (ctx *CompileContext) GetRootContext() *CompileContext {
	current := ctx
	for current.Parent != nil {
		current = current.Parent
	}
	return current
}

// IsGlobalScope returns true if this is the global scope
func (ctx *CompileContext) IsGlobalScope() bool {
	return ctx.ScopeType == ScopeGlobal || ctx.Parent == nil
}

// IsFunctionScope returns true if this is a function scope
func (ctx *CompileContext) IsFunctionScope() bool {
	return ctx.ScopeType == ScopeFunction || ctx.ScopeType == ScopeMethod
}

// NewChildContext creates a new child context with the specified scope type
func (ctx *CompileContext) NewChildContext(scopeType ScopeType) *CompileContext {
	child := NewCompileContext(ctx)
	child.ScopeType = scopeType
	return child
}

// AddFunction adds a function to the current context
func (ctx *CompileContext) AddFunction(name string, function *registry.Function) {
	ctx.Functions[name] = function
}

// AddClass adds a class to the current context
func (ctx *CompileContext) AddClass(name string, class *registry.Class) {
	ctx.Classes[name] = class
}

// AddInterface adds an interface to the current context
func (ctx *CompileContext) AddInterface(name string, iface *registry.Interface) {
	ctx.Interfaces[name] = iface
}

// AddTrait adds a trait to the current context
func (ctx *CompileContext) AddTrait(name string, trait *registry.Trait) {
	ctx.Traits[name] = trait
}

// GetAllConstants returns all constants from this context and its parents
func (ctx *CompileContext) GetAllConstants() []*values.Value {
	var allConstants []*values.Value

	// Collect constants from parent chain (bottom-up)
	if ctx.Parent != nil {
		allConstants = append(allConstants, ctx.Parent.GetAllConstants()...)
	}

	// Add constants from current context
	allConstants = append(allConstants, ctx.Constants...)

	return allConstants
}

// GetAllInstructions returns all instructions from this context and its parents
func (ctx *CompileContext) GetAllInstructions() []*opcodes.Instruction {
	var allInstructions []*opcodes.Instruction

	// Collect instructions from parent chain (bottom-up)
	if ctx.Parent != nil {
		allInstructions = append(allInstructions, ctx.Parent.GetAllInstructions()...)
	}

	// Add instructions from current context
	allInstructions = append(allInstructions, ctx.Instructions...)

	return allInstructions
}
