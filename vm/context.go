package vm

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/wudi/hey/opcodes"
	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// ExecutionContext carries the mutable state associated with executing a single
// PHP script (and its nested function calls) inside the virtual machine.
type ExecutionContext struct {
	// Use sync.Map for thread-safe concurrent access without locking issues
	GlobalVars    *sync.Map // map[string]*values.Value
	IncludedFiles *sync.Map // map[string]bool
	Variables     *sync.Map // map[string]*values.Value
	Temporaries   *sync.Map // map[uint32]*values.Value
	ClassTable    *sync.Map // map[string]*classRuntime

	// User symbols can use regular maps with RWMutex since they're mostly read-only after setup
	userSymbolsMu    sync.RWMutex
	UserFunctions    map[string]*registry.Function
	UserClasses      map[string]*registry.Class
	UserInterfaces   map[string]*registry.Interface
	UserTraits       map[string]*registry.Trait

	// Mutex for stack and frame operations
	frameMu sync.Mutex

	Stack []*values.Value

	OutputWriter io.Writer
	OutputBufferStack *OutputBufferStack

	Halted bool
	ExitCode int

	CallStack []*CallFrame

	Constants    []*values.Value
	currentClass *classRuntime

	debugLog []string

	// Error reporting level for @ operator support
	ErrorReportingLevel int

	// Execution timeout support with Go context
	ctx            context.Context
	cancel         context.CancelFunc
	maxExecutionTime time.Duration
	timeoutMu        sync.RWMutex
}

// NewExecutionContext constructs a fresh execution context with sane defaults.
func NewExecutionContext() *ExecutionContext {
	ctx, cancel := context.WithCancel(context.Background())
	baseWriter := os.Stdout
	ec := &ExecutionContext{
		Stack:            make([]*values.Value, 0, 16),
		OutputWriter:     baseWriter,
		GlobalVars:       &sync.Map{},
		IncludedFiles:    &sync.Map{},
		Variables:        &sync.Map{},
		Temporaries:      &sync.Map{},
		ClassTable:       &sync.Map{},
		CallStack:        make([]*CallFrame, 0, 8),
		UserFunctions:    make(map[string]*registry.Function),
		UserClasses:      make(map[string]*registry.Class),
		UserInterfaces:   make(map[string]*registry.Interface),
		UserTraits:       make(map[string]*registry.Trait),
		debugLog:         make([]string, 0, 64),
		ErrorReportingLevel: 1, // Default: show errors (1 = on, 0 = off/silenced)
		ctx:              ctx,
		cancel:           cancel,
		maxExecutionTime: 0, // 0 means unlimited (default PHP behavior)
	}
	// Initialize output buffer stack
	ec.OutputBufferStack = NewOutputBufferStack(baseWriter)
	// Set the output writer to use the buffer stack
	ec.OutputWriter = ec.OutputBufferStack
	return ec
}

// SetOutputWriter allows callers to redirect the script output stream.
func (ctx *ExecutionContext) SetOutputWriter(w io.Writer) {
	if w == nil {
		return
	}
	// Update the base writer of the output buffer stack
	if ctx.OutputBufferStack != nil {
		ctx.OutputBufferStack.baseWriter = w
		ctx.OutputWriter = ctx.OutputBufferStack
	} else {
		ctx.OutputWriter = w
	}
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

// copyClassRuntime creates a deep copy of a classRuntime for goroutine isolation
func copyClassRuntime(src *classRuntime) *classRuntime {
	if src == nil {
		return nil
	}

	dst := &classRuntime{
		Name:       src.Name,
		Parent:     src.Parent,
		Descriptor: src.Descriptor, // Descriptor can be shared as it's immutable
		Properties: make(map[string]*propertyRuntime),
		StaticProps: make(map[string]*values.Value),
		Constants:   make(map[string]*values.Value),
	}

	// Deep copy properties
	for k, v := range src.Properties {
		dst.Properties[k] = clonePropertyRuntime(v)
	}

	// Deep copy static properties
	for k, v := range src.StaticProps {
		dst.StaticProps[k] = copyValue(v)
	}

	// Deep copy constants
	for k, v := range src.Constants {
		dst.Constants[k] = copyValue(v)
	}

	return dst
}

// pushFrame adds a new call frame to the call stack.
func (ctx *ExecutionContext) pushFrame(frame *CallFrame) {
	ctx.frameMu.Lock()
	defer ctx.frameMu.Unlock()
	ctx.CallStack = append(ctx.CallStack, frame)
}

// popFrame removes and returns the current call frame. Returns nil when the
// stack is empty.
func (ctx *ExecutionContext) popFrame() *CallFrame {
	ctx.frameMu.Lock()
	defer ctx.frameMu.Unlock()
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
	ctx.frameMu.Lock()
	defer ctx.frameMu.Unlock()
	if len(ctx.CallStack) == 0 {
		return nil
	}
	return ctx.CallStack[len(ctx.CallStack)-1]
}

// appendDebugRecord records an entry for later inspection via debug reports.
func (ctx *ExecutionContext) appendDebugRecord(record string) {
	ctx.frameMu.Lock()
	defer ctx.frameMu.Unlock()
	ctx.debugLog = append(ctx.debugLog, record)
}

// drainDebugRecords returns the accumulated debug log.
func (ctx *ExecutionContext) drainDebugRecords() []string {
	ctx.frameMu.Lock()
	defer ctx.frameMu.Unlock()
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
	CallingClass string // For late static binding - the class that initiated the call

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

	// Generator context for generator functions
	Generator interface{}

	// Generator state
	generatorIndex int // Auto-incrementing index for generator keys
}

// SetGenerator sets the generator reference for this call frame
func (cf *CallFrame) SetGenerator(generator interface{}) {
	cf.Generator = generator
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
	ArgNames    []string        // Named argument names (empty string for positional args)
	Result      operandTarget
	Method      bool
	Static      bool
	This        *values.Value
	ClassName   string
	CallingClass string // For late static binding - the class that initiated the call
	MethodName  string
	IsMagicMethod bool  // Flag to indicate magic method calls (__call, __callStatic)
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

	// Try to load existing class first
	if val, ok := ctx.ClassTable.Load(key); ok {
		return val.(*classRuntime)
	}

	// Create new class
	cls := &classRuntime{
		Name:        name,
		Properties:  make(map[string]*propertyRuntime),
		StaticProps: make(map[string]*values.Value),
		Constants:   make(map[string]*values.Value),
	}

	// Get user class definition with proper locking
	ctx.userSymbolsMu.RLock()
	userClassDef, hasUserClass := ctx.UserClasses[strings.ToLower(name)]
	ctx.userSymbolsMu.RUnlock()

	if hasUserClass {
		populateRuntimeFromClassDef(cls, userClassDef)
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

	// Use LoadOrStore to handle race conditions
	if actual, loaded := ctx.ClassTable.LoadOrStore(key, cls); loaded {
		// Another goroutine created it first, use theirs
		cls = actual.(*classRuntime)
	}

	// Handle parent class inheritance (recursive call is safe now)
	if cls.Parent != "" {
		if parent := ctx.ensureClass(cls.Parent); parent != nil {
			inheritClassMetadata(cls, parent)
		}
	}

	return cls
}

func (ctx *ExecutionContext) getClass(name string) (*classRuntime, bool) {
	if val, ok := ctx.ClassTable.Load(classKey(name)); ok {
		return val.(*classRuntime), true
	}
	return nil, false
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
				IsReadonly:   prop.IsReadonly,
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
	// First check GlobalVars
	for _, variant := range globalNameVariants(name) {
		if val, ok := ctx.GlobalVars.Load(variant); ok {
			ctx.bindGlobalValue(name, val.(*values.Value))
			return val.(*values.Value)
		}
	}

	// Then check Variables
	for _, variant := range globalNameVariants(name) {
		if val, ok := ctx.Variables.Load(variant); ok {
			ctx.bindGlobalValue(name, val.(*values.Value))
			return val.(*values.Value)
		}
	}

	// Create new null value
	null := values.NewNull()
	ctx.bindGlobalValue(name, null)
	return null
}

func (ctx *ExecutionContext) setVariable(name string, value *values.Value) {
	ctx.Variables.Store(name, value)
	if sanitized := sanitizeVariableName(name); sanitized != "" && sanitized != name {
		ctx.Variables.Store(sanitized, value)
	}
}

func (ctx *ExecutionContext) unsetVariable(name string) {
	ctx.Variables.Delete(name)
	if sanitized := sanitizeVariableName(name); sanitized != "" && sanitized != name {
		ctx.Variables.Delete(sanitized)
	}
}

func (ctx *ExecutionContext) bindGlobalValue(name string, value *values.Value) {
	variants := globalNameVariants(name)
	for _, variant := range variants {
		ctx.GlobalVars.Store(variant, value)
	}
	ctx.updateGlobalBindings(variants, value)
}

func (ctx *ExecutionContext) unsetGlobal(name string) {
	for _, variant := range globalNameVariants(name) {
		ctx.GlobalVars.Delete(variant)
	}
}

func (ctx *ExecutionContext) updateGlobalBindings(names []string, value *values.Value) {
	ctx.frameMu.Lock()
	defer ctx.frameMu.Unlock()
	for _, frame := range ctx.CallStack {
		if frame == nil || len(frame.GlobalSlots) == 0 {
			continue
		}
		for slot, bound := range frame.GlobalSlots {
			for _, candidate := range names {
				if bound == candidate {
					// Check if the current local is a reference - if so, don't overwrite it
					currentLocal := frame.Locals[slot]
					if currentLocal != nil && currentLocal.IsReference() {
						// If the local is a reference, update its target instead of replacing it
						ref := currentLocal.Data.(*values.Reference)
						if !value.IsReference() {
							ref.Target.Type = value.Type
							ref.Target.Data = value.Data
						}
						// If value is also a reference, we could link them, but for now skip the update
					} else {
						// Normal global binding - replace the local value
						frame.Locals[slot] = value
					}
					break
				}
			}
		}
	}
}

func (ctx *ExecutionContext) setTemporary(slot uint32, value *values.Value) {
	ctx.Temporaries.Store(slot, value)
}

// GetGlobalVar safely retrieves a global variable
func (ctx *ExecutionContext) GetGlobalVar(name string) (*values.Value, bool) {
	if val, ok := ctx.GlobalVars.Load(name); ok {
		return val.(*values.Value), true
	}
	return nil, false
}

// SetGlobalVar safely sets a global variable
func (ctx *ExecutionContext) SetGlobalVar(name string, val *values.Value) {
	ctx.GlobalVars.Store(name, val)
}

// GetUserFunction safely retrieves a user function
func (ctx *ExecutionContext) GetUserFunction(name string) (*registry.Function, bool) {
	ctx.userSymbolsMu.RLock()
	defer ctx.userSymbolsMu.RUnlock()
	fn, ok := ctx.UserFunctions[name]
	return fn, ok
}

// GetUserClass safely retrieves a user class
func (ctx *ExecutionContext) GetUserClass(name string) (*registry.Class, bool) {
	ctx.userSymbolsMu.RLock()
	defer ctx.userSymbolsMu.RUnlock()
	cls, ok := ctx.UserClasses[name]
	return cls, ok
}

// IsFileIncluded safely checks if a file is included
func (ctx *ExecutionContext) IsFileIncluded(path string) bool {
	if val, ok := ctx.IncludedFiles.Load(path); ok {
		return val.(bool)
	}
	return false
}

// MarkFileIncluded safely marks a file as included
func (ctx *ExecutionContext) MarkFileIncluded(path string) {
	ctx.IncludedFiles.Store(path, true)
}

// GetVariable safely retrieves a variable
func (ctx *ExecutionContext) GetVariable(name string) (*values.Value, bool) {
	if val, ok := ctx.Variables.Load(name); ok {
		return val.(*values.Value), true
	}
	return nil, false
}

// GetTemporary safely retrieves a temporary variable
func (ctx *ExecutionContext) GetTemporary(slot uint32) (*values.Value, bool) {
	if val, ok := ctx.Temporaries.Load(slot); ok {
		return val.(*values.Value), true
	}
	return nil, false
}

func (ctx *ExecutionContext) exportState(frame *CallFrame) {
	if frame == nil {
		return
	}

	// Export constants (no locking needed for slice)
	ctx.Constants = frame.cloneConstants()

	// Clear and repopulate Temporaries
	ctx.Temporaries = &sync.Map{}
	for slot, val := range frame.TempVars {
		ctx.Temporaries.Store(slot, val)
	}

	// Clear and repopulate Variables
	ctx.Variables = &sync.Map{}
	for slot, val := range frame.Locals {
		if name, ok := frame.SlotNames[slot]; ok && name != "" {
			ctx.Variables.Store(name, val)
		} else {
			ctx.Variables.Store(fmt.Sprintf("$%d", slot), val)
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

// SetTimeLimit sets the maximum execution time for the script.
// If seconds is 0, the timeout is unlimited.
// If seconds is negative, it's treated as unlimited in PHP 8.0+
func (ctx *ExecutionContext) SetTimeLimit(seconds int) bool {
	ctx.timeoutMu.Lock()
	defer ctx.timeoutMu.Unlock()

	// Cancel the existing context
	if ctx.cancel != nil {
		ctx.cancel()
	}

	// Set the new max execution time
	if seconds <= 0 {
		// Unlimited execution time
		ctx.maxExecutionTime = 0
		ctx.ctx, ctx.cancel = context.WithCancel(context.Background())
	} else {
		// Set timeout
		ctx.maxExecutionTime = time.Duration(seconds) * time.Second
		ctx.ctx, ctx.cancel = context.WithTimeout(context.Background(), ctx.maxExecutionTime)
	}

	return true
}

// GetMaxExecutionTime returns the current max execution time in seconds.
// Returns 0 if unlimited.
func (ctx *ExecutionContext) GetMaxExecutionTime() int {
	ctx.timeoutMu.RLock()
	defer ctx.timeoutMu.RUnlock()

	if ctx.maxExecutionTime == 0 {
		return 0
	}
	return int(ctx.maxExecutionTime.Seconds())
}

// GetMaxExecutionTimeAsDuration returns the current max execution time as a Duration.
// Returns 0 if unlimited.
func (ctx *ExecutionContext) GetMaxExecutionTimeAsDuration() time.Duration {
	ctx.timeoutMu.RLock()
	defer ctx.timeoutMu.RUnlock()

	return ctx.maxExecutionTime
}

// CheckTimeout checks if the execution has timed out.
// Returns an error if timeout occurred.
func (ctx *ExecutionContext) CheckTimeout() error {
	ctx.timeoutMu.RLock()
	defer ctx.timeoutMu.RUnlock()

	if ctx.ctx == nil {
		return nil
	}

	select {
	case <-ctx.ctx.Done():
		if ctx.ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("Fatal error: Maximum execution time of %d seconds exceeded", int(ctx.maxExecutionTime.Seconds()))
		}
		return ctx.ctx.Err()
	default:
		return nil
	}
}

// Cancel cancels the execution context
func (ctx *ExecutionContext) Cancel() {
	ctx.timeoutMu.Lock()
	defer ctx.timeoutMu.Unlock()

	if ctx.cancel != nil {
		ctx.cancel()
	}
}
