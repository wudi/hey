package vm

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/wudi/hey/opcodes"
	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// ExecutionContext carries the mutable state associated with executing a single
// PHP script (and its nested function calls) inside the virtual machine.
type ExecutionContext struct {
	mu sync.Mutex

	Stack []*values.Value

	OutputWriter io.Writer

	GlobalVars map[string]*values.Value

	IncludedFiles map[string]bool

	Halted bool

	CallStack []*CallFrame

	UserFunctions map[string]*registry.Function
	UserClasses   map[string]*registry.Class

	Constants    []*values.Value
	Variables    map[string]*values.Value
	Temporaries  map[uint32]*values.Value
	ClassTable   map[string]*classRuntime
	currentClass *classRuntime

	debugLog []string
}

// NewExecutionContext constructs a fresh execution context with sane defaults.
func NewExecutionContext() *ExecutionContext {
	return &ExecutionContext{
		Stack:         make([]*values.Value, 0, 16),
		OutputWriter:  os.Stdout,
		GlobalVars:    make(map[string]*values.Value),
		IncludedFiles: make(map[string]bool),
		CallStack:     make([]*CallFrame, 0, 8),
		UserFunctions: make(map[string]*registry.Function),
		UserClasses:   make(map[string]*registry.Class),
		Variables:     make(map[string]*values.Value),
		Temporaries:   make(map[uint32]*values.Value),
		ClassTable:    make(map[string]*classRuntime),
		debugLog:      make([]string, 0, 64),
	}
}

// SetOutputWriter allows callers to redirect the script output stream.
func (ctx *ExecutionContext) SetOutputWriter(w io.Writer) {
	if w == nil {
		return
	}
	ctx.OutputWriter = w
}

type exceptionHandler struct {
	catchIP   int
	finallyIP int
}

type classRuntime struct {
	Name        string
	Parent      string
	Properties  map[string]*propertyRuntime
	StaticProps map[string]*values.Value
	Constants   map[string]*values.Value
	Descriptor  *registry.Class
}

type propertyRuntime struct {
	Visibility string
	IsStatic   bool
	Default    *values.Value
}

func clonePropertyRuntime(src *propertyRuntime) *propertyRuntime {
	if src == nil {
		return nil
	}
	return &propertyRuntime{
		Visibility: src.Visibility,
		IsStatic:   src.IsStatic,
		Default:    copyValue(src.Default),
	}
}

// pushFrame adds a new call frame to the call stack.
func (ctx *ExecutionContext) pushFrame(frame *CallFrame) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.CallStack = append(ctx.CallStack, frame)
}

// popFrame removes and returns the current call frame. Returns nil when the
// stack is empty.
func (ctx *ExecutionContext) popFrame() *CallFrame {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	if len(ctx.CallStack) == 0 {
		return nil
	}
	idx := len(ctx.CallStack) - 1
	frame := ctx.CallStack[idx]
	ctx.CallStack = ctx.CallStack[:idx]
	return frame
}

// currentFrame returns the actively executing call frame.
func (ctx *ExecutionContext) currentFrame() *CallFrame {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	if len(ctx.CallStack) == 0 {
		return nil
	}
	return ctx.CallStack[len(ctx.CallStack)-1]
}

// appendDebugRecord records an entry for later inspection via debug reports.
func (ctx *ExecutionContext) appendDebugRecord(record string) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.debugLog = append(ctx.debugLog, record)
}

// drainDebugRecords returns the accumulated debug log.
func (ctx *ExecutionContext) drainDebugRecords() []string {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	out := make([]string, len(ctx.debugLog))
	copy(out, ctx.debugLog)
	return out
}

// CallFrame houses the interpreter state necessary to execute a function body
// (user or builtin) including locals, temporaries, and the instruction pointer.
type CallFrame struct {
	Function     *registry.Function
	FunctionName string
	ClassName    string

	Instructions []*opcodes.Instruction
	Constants    []*values.Value

	IP int

	Locals           map[uint32]*values.Value
	TempVars         map[uint32]*values.Value
	SlotNames        map[uint32]string
	NameSlots        map[string]uint32
	GlobalSlots      map[uint32]string
	Iterators        map[uint32]*foreachIterator
	exHandlers       []*exceptionHandler
	pendingException *values.Value

	pendingCalls []*PendingCall

	ReturnTarget operandTarget

	This *values.Value
}

// operandTarget identifies where a return value should be written when a call
// completes. It mirrors the opcode operand encoding.
type operandTarget struct {
	opType opcodes.OpType
	slot   uint32
	valid  bool
}

// PendingCall captures intermediate state between INIT_FCALL/SEND/DO_FCALL.
type PendingCall struct {
	Callee      *values.Value
	Function    *registry.Function
	ClosureName string
	Args        []*values.Value
	Result      operandTarget
	Method      bool
	Static      bool
	This        *values.Value
	ClassName   string
	MethodName  string
}

// newCallFrame constructs an initialized call frame.
func newCallFrame(name string, fn *registry.Function, instructions []*opcodes.Instruction, constants []*values.Value) *CallFrame {
	return &CallFrame{
		Function:     fn,
		FunctionName: name,
		Instructions: instructions,
		Constants:    constants,
		IP:           0,
		Locals:       make(map[uint32]*values.Value),
		TempVars:     make(map[uint32]*values.Value),
		SlotNames:    make(map[uint32]string),
		NameSlots:    make(map[string]uint32),
		GlobalSlots:  make(map[uint32]string),
		Iterators:    make(map[uint32]*foreachIterator),
		exHandlers:   make([]*exceptionHandler, 0, 4),
		pendingCalls: make([]*PendingCall, 0, 4),
	}
}

func (f *CallFrame) setLocal(slot uint32, val *values.Value) {
	f.Locals[slot] = val
}

func (f *CallFrame) getLocal(slot uint32) *values.Value {
	if val, ok := f.Locals[slot]; ok {
		return val
	}
	return values.NewNull()
}

func (f *CallFrame) getLocalWithStatus(slot uint32) (*values.Value, bool) {
	if val, ok := f.Locals[slot]; ok {
		return val, true
	}
	return values.NewNull(), false
}

func (f *CallFrame) ensureLocal(slot uint32) *values.Value {
	if val, ok := f.Locals[slot]; ok {
		return val
	}
	null := values.NewNull()
	f.Locals[slot] = null
	return null
}

func (f *CallFrame) unsetLocal(slot uint32) {
	delete(f.Locals, slot)
}

func (f *CallFrame) setTemp(slot uint32, val *values.Value) {
	f.TempVars[slot] = val
}

func (f *CallFrame) getTemp(slot uint32) *values.Value {
	if val, ok := f.TempVars[slot]; ok {
		return val
	}
	return values.NewNull()
}

func (f *CallFrame) ensureTemp(slot uint32) *values.Value {
	if val, ok := f.TempVars[slot]; ok {
		return val
	}
	null := values.NewNull()
	f.TempVars[slot] = null
	return null
}

func (f *CallFrame) bindSlotName(slot uint32, name string) {
	if name == "" {
		return
	}
	f.SlotNames[slot] = name
	f.NameSlots[name] = slot
	if sanitized := sanitizeVariableName(name); sanitized != "" && sanitized != name {
		f.NameSlots[sanitized] = slot
	}
}

func (f *CallFrame) bindGlobalSlot(slot uint32, name string) {
	if name == "" {
		return
	}
	if f.GlobalSlots == nil {
		f.GlobalSlots = make(map[uint32]string)
	}
	f.GlobalSlots[slot] = name
}

func (f *CallFrame) unbindGlobalSlot(slot uint32) {
	if f.GlobalSlots == nil {
		return
	}
	delete(f.GlobalSlots, slot)
}

func (f *CallFrame) globalSlotName(slot uint32) (string, bool) {
	if f.GlobalSlots == nil {
		return "", false
	}
	name, ok := f.GlobalSlots[slot]
	return name, ok
}

func (f *CallFrame) slotByName(name string) (uint32, bool) {
	slot, ok := f.NameSlots[name]
	return slot, ok
}

// setReturnTarget configures where the next return value should be written.
func (f *CallFrame) setReturnTarget(opType opcodes.OpType, slot uint32) {
	f.ReturnTarget = operandTarget{opType: opType, slot: slot, valid: opType != opcodes.IS_UNUSED}
}

// resetReturnTarget clears any pending return target information.
func (f *CallFrame) resetReturnTarget() {
	f.ReturnTarget.valid = false
}

func (f *CallFrame) pushPendingCall(call *PendingCall) {
	f.pendingCalls = append(f.pendingCalls, call)
}

func (f *CallFrame) popPendingCall() *PendingCall {
	if len(f.pendingCalls) == 0 {
		return nil
	}
	idx := len(f.pendingCalls) - 1
	call := f.pendingCalls[idx]
	f.pendingCalls = f.pendingCalls[:idx]
	return call
}

func (f *CallFrame) currentPendingCall() *PendingCall {
	if len(f.pendingCalls) == 0 {
		return nil
	}
	return f.pendingCalls[len(f.pendingCalls)-1]
}

func (f *CallFrame) pushExceptionHandler(handler *exceptionHandler) {
	f.exHandlers = append(f.exHandlers, handler)
}

func (f *CallFrame) popExceptionHandler() *exceptionHandler {
	if len(f.exHandlers) == 0 {
		return nil
	}
	idx := len(f.exHandlers) - 1
	h := f.exHandlers[idx]
	f.exHandlers = f.exHandlers[:idx]
	return h
}

func (f *CallFrame) peekExceptionHandler() *exceptionHandler {
	if len(f.exHandlers) == 0 {
		return nil
	}
	return f.exHandlers[len(f.exHandlers)-1]
}

func (f *CallFrame) cloneConstants() []*values.Value {
	if len(f.Constants) == 0 {
		return nil
	}
	out := make([]*values.Value, len(f.Constants))
	copy(out, f.Constants)
	return out
}

func classKey(name string) string {
	return strings.ToLower(name)
}

func sanitizeVariableName(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return trimmed
	}
	if strings.HasPrefix(trimmed, "$") {
		trimmed = trimmed[1:]
	}
	if len(trimmed) > 2 && strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
		inner := strings.TrimSpace(trimmed[1 : len(trimmed)-1])
		trimmed = strings.Trim(inner, "\"'")
	}
	return trimmed
}

func globalNameVariants(name string) []string {
	variants := make([]string, 0, 3)
	seen := make(map[string]struct{}, 3)
	add := func(candidate string) {
		if candidate == "" {
			return
		}
		if _, ok := seen[candidate]; ok {
			return
		}
		seen[candidate] = struct{}{}
		variants = append(variants, candidate)
	}

	add(name)
	sanitized := sanitizeVariableName(name)
	if sanitized != "" {
		add(sanitized)
		add("$" + sanitized)
	}

	return variants
}

func (ctx *ExecutionContext) ensureClass(name string) *classRuntime {
	key := classKey(name)
	if cls, ok := ctx.ClassTable[key]; ok {
		return cls
	}
	cls := &classRuntime{
		Name:        name,
		Properties:  make(map[string]*propertyRuntime),
		StaticProps: make(map[string]*values.Value),
		Constants:   make(map[string]*values.Value),
	}
	if def, ok := ctx.UserClasses[strings.ToLower(name)]; ok {
		populateRuntimeFromClassDef(cls, def)
	}
	if cls.Descriptor == nil && registry.GlobalRegistry != nil {
		if desc, err := registry.GlobalRegistry.GetClass(name); err == nil {
			if converted := classFromDescriptor(desc); converted != nil {
				populateRuntimeFromClassDef(cls, converted)
			}
		}
	}
	if cls.Descriptor == nil {
		if cached := getGlobalClass(name); cached != nil {
			populateRuntimeFromClassDef(cls, cached)
		}
	}
	ctx.ClassTable[key] = cls
	if cls.Parent != "" {
		if parent := ctx.ensureClass(cls.Parent); parent != nil {
			inheritClassMetadata(cls, parent)
		}
	}
	return cls
}

func (ctx *ExecutionContext) getClass(name string) (*classRuntime, bool) {
	cls, ok := ctx.ClassTable[classKey(name)]
	return cls, ok
}

func inheritClassMetadata(target, parent *classRuntime) {
	if target == nil || parent == nil {
		return
	}
	if target.Parent == "" {
		target.Parent = parent.Name
	}
	if parent.Properties != nil {
		for name, prop := range parent.Properties {
			if _, exists := target.Properties[name]; !exists {
				target.Properties[name] = clonePropertyRuntime(prop)
				if prop.IsStatic {
					if _, ok := target.StaticProps[name]; !ok {
						target.StaticProps[name] = copyValue(prop.Default)
					}
				}
			}
		}
	}
	if parent.Constants != nil {
		for name, val := range parent.Constants {
			if _, exists := target.Constants[name]; !exists {
				target.Constants[name] = copyValue(val)
			}
		}
	}
	if parent.StaticProps != nil {
		for name, val := range parent.StaticProps {
			if _, exists := target.StaticProps[name]; !exists {
				target.StaticProps[name] = copyValue(val)
			}
		}
	}
}

func populateRuntimeFromClassDef(target *classRuntime, def *registry.Class) {
	if target == nil || def == nil {
		return
	}
	target.Descriptor = def
	if def.Parent != "" {
		target.Parent = def.Parent
	}
	if def.Properties != nil {
		for propName, prop := range def.Properties {
			runtimeProp := &propertyRuntime{
				Visibility: prop.Visibility,
				IsStatic:   prop.IsStatic,
				Default:    copyValue(prop.DefaultValue),
			}
			target.Properties[propName] = runtimeProp
			if prop.IsStatic {
				target.StaticProps[propName] = copyValue(prop.DefaultValue)
			}
		}
	}
	if def.Constants != nil {
		for constName, constant := range def.Constants {
			target.Constants[constName] = copyValue(constant.Value)
		}
	}
}

func classFromDescriptor(desc *registry.ClassDescriptor) *registry.Class {
	if desc == nil {
		return nil
	}
	class := &registry.Class{
		Name:       desc.Name,
		Parent:     desc.Parent,
		Interfaces: desc.Interfaces,
		Traits:     desc.Traits,
		IsAbstract: desc.IsAbstract,
		IsFinal:    desc.IsFinal,
		Properties: make(map[string]*registry.Property),
		Methods:    make(map[string]*registry.Function),
		Constants:  make(map[string]*registry.ClassConstant),
	}
	if desc.Properties != nil {
		for name, prop := range desc.Properties {
			class.Properties[name] = &registry.Property{
				Name:         prop.Name,
				Visibility:   prop.Visibility,
				IsStatic:     prop.IsStatic,
				Type:         prop.Type,
				DefaultValue: copyValue(prop.DefaultValue),
			}
		}
	}
	if desc.Constants != nil {
		for name, constant := range desc.Constants {
			class.Constants[name] = &registry.ClassConstant{
				Name:       constant.Name,
				Value:      copyValue(constant.Value),
				Visibility: constant.Visibility,
				IsFinal:    constant.IsFinal,
			}
		}
	}
	if desc.Methods != nil {
		for name, method := range desc.Methods {
			fn := &registry.Function{
				Name:       method.Name,
				Parameters: make([]*registry.Parameter, 0, len(method.Parameters)),
				IsVariadic: method.IsVariadic,
			}
			for _, paramDesc := range method.Parameters {
				fn.Parameters = append(fn.Parameters, &registry.Parameter{
					Name:         paramDesc.Name,
					Type:         paramDesc.Type,
					IsReference:  paramDesc.IsReference,
					HasDefault:   paramDesc.HasDefault,
					DefaultValue: copyValue(paramDesc.DefaultValue),
				})
			}
			switch impl := method.Implementation.(type) {
			case *registry.BytecodeMethodImpl:
				if impl != nil {
					fn.Instructions = impl.Instructions
					fn.Constants = impl.Constants
				}
			case interface{ GetFunction() *registry.Function }:
				// Handle BuiltinMethodImpl
				if builtinFn := impl.GetFunction(); builtinFn != nil {
					fn.IsBuiltin = builtinFn.IsBuiltin
					fn.Builtin = builtinFn.Builtin
					fn.Handler = builtinFn.Handler
					fn.MinArgs = builtinFn.MinArgs
					fn.MaxArgs = builtinFn.MaxArgs
				}
			}
			fn.MinArgs = len(fn.Parameters)
			fn.MaxArgs = len(fn.Parameters)
			class.Methods[name] = fn
		}
	}
	return class
}

func descriptorFromClass(class *registry.Class) *registry.ClassDescriptor {
	if class == nil {
		return nil
	}
	desc := &registry.ClassDescriptor{
		Name:       class.Name,
		Parent:     class.Parent,
		Interfaces: class.Interfaces,
		Traits:     class.Traits,
		IsAbstract: class.IsAbstract,
		IsFinal:    class.IsFinal,
		Properties: make(map[string]*registry.PropertyDescriptor),
		Methods:    make(map[string]*registry.MethodDescriptor),
		Constants:  make(map[string]*registry.ConstantDescriptor),
	}
	if class.Properties != nil {
		for name, prop := range class.Properties {
			desc.Properties[name] = &registry.PropertyDescriptor{
				Name:         prop.Name,
				Visibility:   prop.Visibility,
				IsStatic:     prop.IsStatic,
				Type:         prop.Type,
				DefaultValue: copyValue(prop.DefaultValue),
			}
		}
	}
	if class.Constants != nil {
		for name, constant := range class.Constants {
			desc.Constants[name] = &registry.ConstantDescriptor{
				Name:       constant.Name,
				Visibility: constant.Visibility,
				Value:      copyValue(constant.Value),
				IsFinal:    constant.IsFinal,
			}
		}
	}
	if class.Methods != nil {
		for name, method := range class.Methods {
			if method == nil {
				continue
			}
			params := make([]*registry.ParameterDescriptor, 0, len(method.Parameters))
			paramInfos := make([]*registry.ParameterInfo, 0, len(method.Parameters))
			for _, param := range method.Parameters {
				params = append(params, &registry.ParameterDescriptor{
					Name:         param.Name,
					Type:         param.Type,
					IsReference:  param.IsReference,
					HasDefault:   param.HasDefault,
					DefaultValue: copyValue(param.DefaultValue),
				})
				paramInfos = append(paramInfos, &registry.ParameterInfo{
					Name:         param.Name,
					HasDefault:   param.HasDefault,
					DefaultValue: copyValue(param.DefaultValue),
					IsVariadic:   false,
				})
			}
			if method.IsVariadic && len(paramInfos) > 0 {
				paramInfos[len(paramInfos)-1].IsVariadic = true
			}
			impl := &registry.BytecodeMethodImpl{
				Instructions: method.Instructions,
				Constants:    method.Constants,
				Parameters:   paramInfos,
			}
			desc.Methods[name] = &registry.MethodDescriptor{
				Name:           method.Name,
				Visibility:     "public",
				IsStatic:       false,
				IsAbstract:     false,
				IsFinal:        false,
				IsVariadic:     method.IsVariadic,
				Parameters:     params,
				Implementation: impl,
			}
		}
	}
	return desc
}

func (ctx *ExecutionContext) ensureGlobal(name string) *values.Value {
	for _, variant := range globalNameVariants(name) {
		if val, ok := ctx.GlobalVars[variant]; ok {
			ctx.bindGlobalValue(name, val)
			return val
		}
	}
	for _, variant := range globalNameVariants(name) {
		if ctx.Variables != nil {
			if val, ok := ctx.Variables[variant]; ok {
				ctx.bindGlobalValue(name, val)
				return val
			}
		}
	}
	null := values.NewNull()
	ctx.bindGlobalValue(name, null)
	return null
}

func (ctx *ExecutionContext) setVariable(name string, value *values.Value) {
	if ctx.Variables == nil {
		ctx.Variables = make(map[string]*values.Value)
	}
	ctx.Variables[name] = value
	if sanitized := sanitizeVariableName(name); sanitized != "" && sanitized != name {
		ctx.Variables[sanitized] = value
	}
}

func (ctx *ExecutionContext) unsetVariable(name string) {
	if ctx.Variables == nil {
		return
	}
	delete(ctx.Variables, name)
	if sanitized := sanitizeVariableName(name); sanitized != "" && sanitized != name {
		delete(ctx.Variables, sanitized)
	}
}

func (ctx *ExecutionContext) bindGlobalValue(name string, value *values.Value) {
	if ctx.GlobalVars == nil {
		ctx.GlobalVars = make(map[string]*values.Value)
	}
	variants := globalNameVariants(name)
	for _, variant := range variants {
		ctx.GlobalVars[variant] = value
	}
	ctx.updateGlobalBindings(variants, value)
}

func (ctx *ExecutionContext) unsetGlobal(name string) {
	if ctx.GlobalVars == nil {
		return
	}
	for _, variant := range globalNameVariants(name) {
		delete(ctx.GlobalVars, variant)
	}
}

func (ctx *ExecutionContext) updateGlobalBindings(names []string, value *values.Value) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	for _, frame := range ctx.CallStack {
		if frame == nil || len(frame.GlobalSlots) == 0 {
			continue
		}
		for slot, bound := range frame.GlobalSlots {
			for _, candidate := range names {
				if bound == candidate {
					frame.Locals[slot] = value
					break
				}
			}
		}
	}
}

func (ctx *ExecutionContext) setTemporary(slot uint32, value *values.Value) {
	if ctx.Temporaries == nil {
		ctx.Temporaries = make(map[uint32]*values.Value)
	}
	ctx.Temporaries[slot] = value
}

func (ctx *ExecutionContext) exportState(frame *CallFrame) {
	if frame == nil {
		return
	}
	ctx.Constants = frame.cloneConstants()
	ctx.Temporaries = make(map[uint32]*values.Value, len(frame.TempVars))
	for slot, val := range frame.TempVars {
		ctx.Temporaries[slot] = val
	}
	ctx.Variables = make(map[string]*values.Value, len(frame.Locals))
	for slot, val := range frame.Locals {
		if name, ok := frame.SlotNames[slot]; ok && name != "" {
			ctx.Variables[name] = val
		} else {
			ctx.Variables[fmt.Sprintf("$%d", slot)] = val
		}
	}
}

// recordAssignment is invoked whenever a local variable changes. It feeds the
// debug log and can later be extended to integrate with watch expressions.
func (ctx *ExecutionContext) recordAssignment(frame *CallFrame, slot uint32, value *values.Value) {
	if frame == nil {
		return
	}
	name, ok := frame.SlotNames[slot]
	if !ok {
		return
	}
	ctx.appendDebugRecord(fmt.Sprintf("assign %s = %s", name, value.String()))
}
