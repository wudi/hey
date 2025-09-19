package vm

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/wudi/hey/compiler/ast"
	"github.com/wudi/hey/opcodes"
	"github.com/wudi/hey/registry"
	runtime2 "github.com/wudi/hey/runtime"
	"github.com/wudi/hey/values"
)

// DebugLevel controls the verbosity of runtime diagnostics collected.
type DebugLevel int

const (
	DebugLevelNone DebugLevel = iota
	DebugLevelBasic
	DebugLevelDetailed
)

// CompilerCallback is used by include/require instructions to compile and
// execute additional source files on demand.
type CompilerCallback func(ctx *ExecutionContext, program *ast.Program, filePath string, isRequired bool) (*values.Value, error)

// HotSpot describes an instruction pointer that was executed frequently.
type HotSpot struct {
	IP    int
	Count int
}

// VirtualMachine is the bytecode interpreter that executes compiled PHP
// instructions.
type VirtualMachine struct {
	debugLevel DebugLevel

	breakpoints map[int]struct{}
	watchVars   map[string]struct{}

	profile *profileState

	advancedProfiling bool
	DebugMode         bool

	CompilerCallback CompilerCallback

	mu          sync.Mutex
	lastContext *ExecutionContext
}

var (
	globalClassMu sync.RWMutex
	globalClasses = make(map[string]*registry.Class)
)

func storeGlobalClass(name string, class *registry.Class) *registry.Class {
	if class == nil {
		return nil
	}
	key := strings.ToLower(name)
	globalClassMu.Lock()
	defer globalClassMu.Unlock()
	if existing, ok := globalClasses[key]; ok && existing != nil {
		mergeClassDefinitions(existing, class)
		return existing
	}
	globalClasses[key] = class
	return class
}

func getGlobalClass(name string) *registry.Class {
	globalClassMu.RLock()
	defer globalClassMu.RUnlock()
	if class, ok := globalClasses[strings.ToLower(name)]; ok {
		return class
	}
	return nil
}

func mergeClassDefinitions(target, source *registry.Class) {
	if target == nil || source == nil {
		return
	}
	if source.Properties != nil {
		if target.Properties == nil {
			target.Properties = make(map[string]*registry.Property)
		}
		for name, prop := range source.Properties {
			target.Properties[name] = prop
		}
	}
	if source.Constants != nil {
		if target.Constants == nil {
			target.Constants = make(map[string]*registry.ClassConstant)
		}
		for name, constant := range source.Constants {
			target.Constants[name] = constant
		}
	}
	if source.Methods != nil {
		if target.Methods == nil {
			target.Methods = make(map[string]*registry.Function)
		}
		for name, method := range source.Methods {
			target.Methods[name] = method
		}
	}
}

// NewVirtualMachine constructs a VM with basic instrumentation disabled.
func NewVirtualMachine() *VirtualMachine {
	return &VirtualMachine{
		debugLevel:  DebugLevelNone,
		breakpoints: make(map[int]struct{}),
		watchVars:   make(map[string]struct{}),
		profile:     newProfileState(),
		DebugMode:   false,
	}
}

// NewVirtualMachineWithProfiling constructs a VM with the specified debug level.
func NewVirtualMachineWithProfiling(level DebugLevel) *VirtualMachine {
	vm := NewVirtualMachine()
	vm.debugLevel = level
	return vm
}

func (vm *VirtualMachine) EnableAdvancedProfiling() {
	vm.advancedProfiling = true
}

func (vm *VirtualMachine) SetBreakpoint(ip int) {
	vm.breakpoints[ip] = struct{}{}
}

func (vm *VirtualMachine) WatchVariable(name string) {
	if name != "" {
		vm.watchVars[name] = struct{}{}
	}
}

// Execute runs the provided bytecode inside the supplied execution context.
func (vm *VirtualMachine) Execute(ctx *ExecutionContext, instructions []*opcodes.Instruction, constants []*values.Value, functions map[string]*registry.Function, classes map[string]*registry.Class) error {
	if ctx == nil {
		return errors.New("nil execution context")
	}
	if err := runtime2.Bootstrap(); err != nil {
		return fmt.Errorf("runtime bootstrap failed: %w", err)
	}

	if ctx.OutputWriter == nil {
		ctx.OutputWriter = os.Stdout
	}

	if vm.DebugMode && vm.debugLevel == DebugLevelNone {
		vm.debugLevel = DebugLevelDetailed
	}

	// Merge user-defined functions/classes into the context symbol tables.
	if ctx.UserFunctions == nil {
		ctx.UserFunctions = make(map[string]*registry.Function)
	}
	for name, fn := range functions {
		ctx.UserFunctions[strings.ToLower(name)] = fn
	}

	if ctx.UserClasses == nil {
		ctx.UserClasses = make(map[string]*registry.Class)
	}
	mergedClasses := make(map[string]*registry.Class, len(classes))
	for name, class := range classes {
		merged := storeGlobalClass(name, class)
		lower := strings.ToLower(name)
		ctx.UserClasses[lower] = merged
		mergedClasses[name] = merged
	}
	if registry.GlobalRegistry != nil {
		for _, class := range mergedClasses {
			if desc := descriptorFromClass(class); desc != nil {
				_ = registry.GlobalRegistry.RegisterClass(desc)
			}
		}
	}

	mainFrame := newCallFrame("{main}", nil, instructions, constants)
	ctx.pushFrame(mainFrame)

	vm.mu.Lock()
	vm.lastContext = ctx
	vm.mu.Unlock()

	return vm.run(ctx)
}

func (vm *VirtualMachine) run(ctx *ExecutionContext) error {
	for {
		frame := ctx.currentFrame()
		if frame == nil {
			ctx.Halted = true
			return nil
		}

		if frame.IP < 0 || frame.IP >= len(frame.Instructions) {
			// Implicit return null when reaching the end of the instruction stream.
			if err := vm.handleReturn(ctx, values.NewNull()); err != nil {
				return err
			}
			continue
		}

		inst := frame.Instructions[frame.IP]
		vm.profile.observe(frame.IP, inst.Opcode)

		if vm.debugLevel != DebugLevelNone {
			if _, ok := vm.breakpoints[frame.IP]; ok {
				vm.recordDebug(ctx, fmt.Sprintf("breakpoint hit at %d (%s)", frame.IP, inst.Opcode))
			}
		}

		advance, err := vm.executeInstruction(ctx, frame, inst)
		if err != nil {
			return vm.decorateError(frame, inst, err)
		}

		if ctx.Halted {
			return nil
		}

		if advance {
			frame.IP++
		}
	}
}

func (vm *VirtualMachine) decorateError(frame *CallFrame, inst *opcodes.Instruction, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("vm error at ip=%d opcode=%s: %w", frame.IP, inst.Opcode, err)
}

func (vm *VirtualMachine) executeInstruction(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	switch inst.Opcode {
	case opcodes.OP_NOP:
		return true, nil
	case opcodes.OP_QM_ASSIGN:
		return vm.execAssign(ctx, frame, inst, false)
	case opcodes.OP_ASSIGN:
		return vm.execAssign(ctx, frame, inst, true)
	case opcodes.OP_ASSIGN_REF:
		return vm.execAssignRef(ctx, frame, inst)
	case opcodes.OP_ASSIGN_OP:
		return vm.execAssignOp(ctx, frame, inst)
	case opcodes.OP_ASSIGN_DIM:
		return vm.execAssignDim(ctx, frame, inst)
	case opcodes.OP_THROW:
		return vm.execThrow(ctx, frame, inst)
	case opcodes.OP_PLUS, opcodes.OP_MINUS, opcodes.OP_NOT:
		return vm.execUnary(ctx, frame, inst)
	case opcodes.OP_ADD, opcodes.OP_SUB, opcodes.OP_MUL, opcodes.OP_DIV, opcodes.OP_MOD, opcodes.OP_POW:
		return vm.execArithmetic(ctx, frame, inst)
	case opcodes.OP_BW_AND, opcodes.OP_BW_OR, opcodes.OP_BW_XOR, opcodes.OP_SL, opcodes.OP_SR:
		return vm.execBitwise(ctx, frame, inst)
	case opcodes.OP_CONCAT:
		return vm.execConcat(ctx, frame, inst)
	case opcodes.OP_NEW:
		return vm.execNew(ctx, frame, inst)
	case opcodes.OP_IS_EQUAL, opcodes.OP_IS_NOT_EQUAL, opcodes.OP_IS_IDENTICAL, opcodes.OP_IS_NOT_IDENTICAL,
		opcodes.OP_IS_SMALLER, opcodes.OP_IS_SMALLER_OR_EQUAL, opcodes.OP_IS_GREATER, opcodes.OP_IS_GREATER_OR_EQUAL,
		opcodes.OP_SPACESHIP:
		return vm.execComparison(ctx, frame, inst)
	case opcodes.OP_BOOLEAN_AND, opcodes.OP_BOOLEAN_OR:
		return vm.execBoolean(ctx, frame, inst)
	case opcodes.OP_LOGICAL_AND, opcodes.OP_LOGICAL_OR, opcodes.OP_LOGICAL_XOR:
		return vm.execLogical(ctx, frame, inst)
	case opcodes.OP_PRE_INC, opcodes.OP_PRE_DEC, opcodes.OP_POST_INC, opcodes.OP_POST_DEC:
		return vm.execIncDec(ctx, frame, inst)
	case opcodes.OP_BW_NOT:
		return vm.execBitwiseNot(ctx, frame, inst)
	case opcodes.OP_JMP:
		return vm.execJump(ctx, frame, inst)
	case opcodes.OP_JMPZ, opcodes.OP_JMPNZ:
		return vm.execConditionalJump(ctx, frame, inst)
	case opcodes.OP_FETCH_R:
		return vm.execFetch(ctx, frame, inst)
	case opcodes.OP_FETCH_R_DYNAMIC:
		return vm.execFetchDynamic(ctx, frame, inst)
	case opcodes.OP_FETCH_W, opcodes.OP_FETCH_RW:
		return vm.execFetch(ctx, frame, inst)
	case opcodes.OP_BIND_GLOBAL:
		return vm.execBindGlobal(ctx, frame, inst)
	case opcodes.OP_BIND_VAR_NAME:
		return vm.execBindVarName(ctx, frame, inst)
	case opcodes.OP_DECLARE_FUNCTION:
		return vm.execDeclareFunction(ctx, frame, inst)
	case opcodes.OP_INIT_ARRAY:
		return vm.execInitArray(ctx, frame, inst)
	case opcodes.OP_ADD_ARRAY_ELEMENT:
		return vm.execAddArrayElement(ctx, frame, inst)
	case opcodes.OP_ADD_ARRAY_UNPACK:
		return vm.execArrayUnpack(ctx, frame, inst)
	case opcodes.OP_INIT_CLASS_TABLE:
		return vm.execInitClassTable(ctx, frame, inst)
	case opcodes.OP_SET_CURRENT_CLASS:
		return vm.execSetCurrentClass(ctx, frame, inst)
	case opcodes.OP_SET_CLASS_PARENT:
		return vm.execSetClassParent(ctx, frame, inst)
	case opcodes.OP_DECLARE_PROPERTY:
		return vm.execDeclareProperty(ctx, frame, inst)
	case opcodes.OP_CLEAR_CURRENT_CLASS:
		return vm.execClearCurrentClass(ctx, frame, inst)
	case opcodes.OP_DECLARE_CLASS:
		return vm.execDeclareClass(ctx, frame, inst)
	case opcodes.OP_DECLARE_CONSTANT:
		return vm.execDeclareConstant(ctx, frame, inst)
	case opcodes.OP_FETCH_DIM_R:
		return vm.execFetchDim(ctx, frame, inst)
	case opcodes.OP_FETCH_DIM_IS:
		return vm.execFetchDimIs(ctx, frame, inst)
	case opcodes.OP_FETCH_DIM_UNSET:
		return vm.execFetchDimUnset(ctx, frame, inst)
	case opcodes.OP_FETCH_DIM_W:
		return vm.execFetchDimWrite(ctx, frame, inst, false)
	case opcodes.OP_FETCH_DIM_RW:
		return vm.execFetchDimWrite(ctx, frame, inst, true)
	case opcodes.OP_ASSIGN_OBJ:
		return vm.execAssignObj(ctx, frame, inst)
	case opcodes.OP_FETCH_OBJ_R:
		return vm.execFetchObj(ctx, frame, inst)
	case opcodes.OP_FETCH_OBJ_IS:
		return vm.execFetchObjIs(ctx, frame, inst)
	case opcodes.OP_FETCH_STATIC_PROP_R:
		return vm.execFetchStaticProp(ctx, frame, inst)
	case opcodes.OP_FETCH_STATIC_PROP_W:
		return vm.execFetchStaticPropWrite(ctx, frame, inst)
	case opcodes.OP_CATCH:
		return vm.execCatch(ctx, frame, inst)
	case opcodes.OP_FINALLY:
		return vm.execFinally(ctx, frame, inst)
	case opcodes.OP_UNSET_VAR:
		return vm.execUnsetVar(ctx, frame, inst)
	case opcodes.OP_ISSET_ISEMPTY_VAR:
		return vm.execIssetIsEmptyVar(ctx, frame, inst)
	case opcodes.OP_ECHO:
		return vm.execEcho(ctx, frame, inst)
	case opcodes.OP_FE_RESET:
		return vm.execFeReset(ctx, frame, inst)
	case opcodes.OP_FE_FETCH:
		return vm.execFeFetch(ctx, frame, inst)
	case opcodes.OP_FE_FREE:
		return vm.execFeFree(ctx, frame, inst)
	case opcodes.OP_INIT_FCALL, opcodes.OP_INIT_FCALL_BY_NAME:
		return vm.execInitFCall(ctx, frame, inst)
	case opcodes.OP_INIT_METHOD_CALL:
		return vm.execInitMethodCall(ctx, frame, inst)
	case opcodes.OP_INIT_STATIC_METHOD_CALL:
		return vm.execInitStaticMethodCall(ctx, frame, inst)
	case opcodes.OP_SEND_VAL, opcodes.OP_SEND_VAR, opcodes.OP_SEND_REF:
		return vm.execSendArg(ctx, frame, inst)
	case opcodes.OP_DO_FCALL, opcodes.OP_DO_UCALL, opcodes.OP_DO_ICALL, opcodes.OP_DO_FCALL_BY_NAME:
		return vm.execDoFCall(ctx, frame, inst)
	case opcodes.OP_RETURN, opcodes.OP_RETURN_BY_REF:
		return vm.execReturn(ctx, frame, inst)
	case opcodes.OP_CREATE_CLOSURE:
		return vm.execCreateClosure(ctx, frame, inst)
	case opcodes.OP_BIND_USE_VAR:
		return vm.execBindUseVar(ctx, frame, inst)
	case opcodes.OP_INCLUDE, opcodes.OP_INCLUDE_ONCE, opcodes.OP_REQUIRE, opcodes.OP_REQUIRE_ONCE:
		return vm.execInclude(ctx, frame, inst)
	case opcodes.OP_FETCH_CLASS_CONSTANT:
		return vm.execFetchClassConstant(ctx, frame, inst)
	case opcodes.OP_ASSIGN_EXCEPTION:
		return vm.execAssignException(ctx, frame, inst)
	case opcodes.OP_INSTANCEOF:
		return vm.execInstanceof(ctx, frame, inst)
	case opcodes.OP_YIELD:
		return vm.execYield(ctx, frame, inst)
	case opcodes.OP_YIELD_FROM:
		return vm.execYieldFrom(ctx, frame, inst)
	default:
		return false, fmt.Errorf("opcode %s not implemented", inst.Opcode)
	}
}

func (vm *VirtualMachine) handleReturn(ctx *ExecutionContext, value *values.Value) error {
	completed := ctx.popFrame()
	if completed == nil {
		ctx.Halted = true
		return nil
	}

	caller := ctx.currentFrame()
	if caller == nil {
		// top-level script
		ctx.exportState(completed)
		ctx.Stack = append(ctx.Stack, value)
		ctx.Halted = true
		return nil
	}

	target := caller.ReturnTarget
	if target.valid {
		if err := vm.writeOperand(ctx, caller, target.opType, target.slot, value); err != nil {
			return err
		}
		caller.resetReturnTarget()
	}

	// Advance caller beyond call instruction.
	caller.IP++
	return nil
}

// GetPerformanceReport renders a summary of the collected profiling data.
func (vm *VirtualMachine) GetPerformanceReport() string {
	return vm.profile.render()
}

// GetDebugReport returns the debug log accumulated by the active execution
// context. The caller is expected to have completed execution before invoking
// this helper.
func (vm *VirtualMachine) GetDebugReport() string {
	vm.mu.Lock()
	ctx := vm.lastContext
	vm.mu.Unlock()

	lines := vm.profile.debugRecords()
	if ctx != nil {
		ctxLines := ctx.drainDebugRecords()
		if len(ctxLines) > 0 {
			lines = append(lines, ctxLines...)
		}
	}
	return strings.Join(lines, "\n")
}

func (vm *VirtualMachine) GetHotSpots(n int) []HotSpot {
	return vm.profile.hotSpots(n)
}

func (vm *VirtualMachine) GetMemoryStats() (allocs int, frees int) {
	return vm.profile.allocs, vm.profile.frees
}

func (vm *VirtualMachine) recordDebug(ctx *ExecutionContext, message string) {
	if ctx != nil {
		ctx.appendDebugRecord(message)
	}
	vm.profile.addDebug(message)
}

func (vm *VirtualMachine) raiseException(ctx *ExecutionContext, frame *CallFrame, value *values.Value) (bool, error) {

	for {
		if frame == nil {
			return false, fmt.Errorf("uncaught exception: %s", value.ToString())
		}
		if handler := frame.popExceptionHandler(); handler != nil {
			if handler.catchIP > 0 {
				frame.IP = handler.catchIP
				frame.pendingException = value
				return false, nil
			}
			if handler.finallyIP > 0 {
				frame.pendingException = value
				frame.IP = handler.finallyIP
				return false, nil
			}
			continue
		}
		ctx.popFrame()
		frame = ctx.currentFrame()
	}
}

func (vm *VirtualMachine) execYield(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	// Extract yield operands
	keyType := opcodes.DecodeOpType1(inst.OpType1)
	valueType := opcodes.DecodeOpType2(inst.OpType1)

	// Get the generator context from call frame
	generator := frame.Generator
	if generator == nil {
		return false, fmt.Errorf("yield called outside generator context")
	}

	// Get key value (if any)
	var keyValue *values.Value
	if keyType != opcodes.IS_UNUSED {
		var err error
		keyValue, err = vm.readOperand(ctx, frame, keyType, inst.Op1)
		if err != nil {
			return false, fmt.Errorf("error getting yield key: %v", err)
		}
	} else {
		// Auto-increment key for generators without explicit keys
		keyValue = values.NewInt(int64(frame.generatorIndex))
		frame.generatorIndex++
	}

	// Get value
	var yieldValue *values.Value
	if valueType != opcodes.IS_UNUSED {
		var err error
		yieldValue, err = vm.readOperand(ctx, frame, valueType, inst.Op2)
		if err != nil {
			return false, fmt.Errorf("error getting yield value: %v", err)
		}
	} else {
		yieldValue = values.NewNull()
	}

	// Use the Yield method to store key and value
	gen := generator.(*runtime2.Generator)
	gen.Yield(keyValue, yieldValue)

	// Store result if needed (yield expression result)
	resultType := opcodes.DecodeResultType(inst.OpType2)
	if resultType != opcodes.IS_UNUSED {
		// For now, yield result is always null (sent value)
		// In full PHP implementation, this would be the value sent via send() method
		result := yieldValue
		err := vm.writeOperand(ctx, frame, resultType, inst.Result, result)
		if err != nil {
			return false, fmt.Errorf("error storing yield result: %v", err)
		}
	}

	return true, nil
}

func (vm *VirtualMachine) execYieldFrom(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	// TODO: Implement yield from - delegates to another generator
	return false, fmt.Errorf("yield from not yet implemented")
}

// CreateExecutionContext creates a new execution context for generator execution
func (vm *VirtualMachine) CreateExecutionContext() interface{} {
	return NewExecutionContext()
}

// CreateCallFrame creates a new call frame for generator function execution
func (vm *VirtualMachine) CreateCallFrame(fn *registry.Function, args []*values.Value) interface{} {
	frame := newCallFrame(fn.Name, fn, fn.Instructions, fn.Constants)

	// Set up arguments in the frame's locals
	for i, arg := range args {
		if i < len(fn.Parameters) {
			// Map argument to parameter slot
			slot := uint32(i)
			frame.Locals[slot] = arg
		}
	}

	return frame
}

// ExecuteFunction executes a function in the given context and frame
func (vm *VirtualMachine) ExecuteFunction(ctxInterface, frameInterface interface{}) error {
	ctx, ok := ctxInterface.(*ExecutionContext)
	if !ok {
		return fmt.Errorf("invalid execution context type")
	}

	frame, ok := frameInterface.(*CallFrame)
	if !ok {
		return fmt.Errorf("invalid call frame type")
	}

	// Push frame onto call stack
	ctx.CallStack = append(ctx.CallStack, frame)
	defer func() {
		// Pop frame when done
		if len(ctx.CallStack) > 0 {
			ctx.CallStack = ctx.CallStack[:len(ctx.CallStack)-1]
		}
	}()

	// Execute function instructions
	for frame.IP < len(frame.Instructions) {
		inst := frame.Instructions[frame.IP]

		continued, err := vm.executeInstruction(ctx, frame, inst)
		if err != nil {
			return err
		}

		if !continued {
			// Function returned or yielded
			break
		}

		frame.IP++
	}

	return nil
}

// ExecuteUntilYield executes a function until first yield or completion
func (vm *VirtualMachine) ExecuteUntilYield(ctxInterface, frameInterface interface{}) (bool, error) {
	ctx, ok := ctxInterface.(*ExecutionContext)
	if !ok {
		return false, fmt.Errorf("invalid execution context type")
	}

	frame, ok := frameInterface.(*CallFrame)
	if !ok {
		return false, fmt.Errorf("invalid call frame type")
	}

	// Push frame onto call stack
	ctx.CallStack = append(ctx.CallStack, frame)
	defer func() {
		// Pop frame when done
		if len(ctx.CallStack) > 0 {
			ctx.CallStack = ctx.CallStack[:len(ctx.CallStack)-1]
		}
	}()

	// Execute function instructions until yield or completion
	for frame.IP < len(frame.Instructions) {
		inst := frame.Instructions[frame.IP]

		// Check if this is a yield instruction
		if inst.Opcode == opcodes.OP_YIELD {
			// Execute the yield instruction
			_, err := vm.executeInstruction(ctx, frame, inst)
			if err != nil {
				return false, err
			}
			// Yield suspends execution - don't advance IP
			// The IP will be advanced when resuming
			return true, nil // true = yielded
		}

		continued, err := vm.executeInstruction(ctx, frame, inst)
		if err != nil {
			return false, err
		}

		if continued {
			// Normal instruction - advance IP
			frame.IP++
		} else {
			// Instruction handled IP manually (jump, return, etc.)
			// Check if this was a return instruction
			if inst.Opcode == opcodes.OP_RETURN || inst.Opcode == opcodes.OP_RETURN_BY_REF || inst.Opcode == opcodes.OP_GENERATOR_RETURN {
				// Function returned
				return false, nil // false = completed
			}
			// Otherwise it was a jump - continue execution from new IP
		}
	}

	return false, nil // false = completed
}

// ResumeFromYield resumes generator execution from suspended state
func (vm *VirtualMachine) ResumeFromYield(ctxInterface, frameInterface interface{}) (bool, error) {
	ctx, ok := ctxInterface.(*ExecutionContext)
	if !ok {
		return false, fmt.Errorf("invalid execution context type")
	}

	frame, ok := frameInterface.(*CallFrame)
	if !ok {
		return false, fmt.Errorf("invalid call frame type")
	}

	// Push frame onto call stack
	ctx.CallStack = append(ctx.CallStack, frame)
	defer func() {
		// Pop frame when done
		if len(ctx.CallStack) > 0 {
			ctx.CallStack = ctx.CallStack[:len(ctx.CallStack)-1]
		}
	}()

	// Advance IP past the yield instruction
	frame.IP++

	// Continue execution until next yield or completion
	for frame.IP < len(frame.Instructions) {
		inst := frame.Instructions[frame.IP]

		// Check if this is a yield instruction
		if inst.Opcode == opcodes.OP_YIELD {
			// Execute the yield instruction
			_, err := vm.executeInstruction(ctx, frame, inst)
			if err != nil {
				return false, err
			}
			// Yield suspends execution - don't advance IP
			return true, nil // true = yielded
		}

		continued, err := vm.executeInstruction(ctx, frame, inst)
		if err != nil {
			return false, err
		}

		if continued {
			// Normal instruction - advance IP
			frame.IP++
		} else {
			// Instruction handled IP manually (jump, return, etc.)
			// Check if this was a return instruction by seeing if we're at end
			if inst.Opcode == opcodes.OP_RETURN || inst.Opcode == opcodes.OP_RETURN_BY_REF || inst.Opcode == opcodes.OP_GENERATOR_RETURN {
				// Function returned
				return false, nil // false = completed
			}
			// Otherwise it was a jump - continue execution from new IP
		}
	}

	return false, nil // false = completed
}
