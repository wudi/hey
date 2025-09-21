package vm

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/wudi/hey/compiler/lexer"
	"github.com/wudi/hey/compiler/parser"
	"github.com/wudi/hey/opcodes"
	"github.com/wudi/hey/registry"
	runtime2 "github.com/wudi/hey/runtime"
	"github.com/wudi/hey/values"
)

type foreachIterator struct {
	keys      []*values.Value
	values    []*values.Value
	index     int
	generator *values.Value // For Generator objects
	isFirst   bool         // For generators, track if this is the first iteration
}

func decodeOperand(inst *opcodes.Instruction, idx int) (opcodes.OpType, uint32) {
	switch idx {
	case 1:
		return opcodes.DecodeOpType1(inst.OpType1), inst.Op1
	case 2:
		return opcodes.DecodeOpType2(inst.OpType1), inst.Op2
	default:
		panic("invalid operand index")
	}
}

func decodeResult(inst *opcodes.Instruction) (opcodes.OpType, uint32) {
	return opcodes.DecodeResultType(inst.OpType2), inst.Result
}

func (vm *VirtualMachine) readOperand(ctx *ExecutionContext, frame *CallFrame, opType opcodes.OpType, operand uint32) (*values.Value, error) {
	switch opType {
	case opcodes.IS_UNUSED:
		return values.NewNull(), nil
	case opcodes.IS_CONST:
		if int(operand) >= len(frame.Constants) {
			return nil, fmt.Errorf("constant index %d out of range", operand)
		}
		return frame.Constants[operand], nil
	case opcodes.IS_TMP_VAR:
		return frame.getTemp(operand), nil
	case opcodes.IS_VAR, opcodes.IS_CV:
		return frame.getLocal(operand), nil
	default:
		return nil, fmt.Errorf("unsupported operand type %d", opType)
	}
}

func (vm *VirtualMachine) writeOperand(ctx *ExecutionContext, frame *CallFrame, opType opcodes.OpType, operand uint32, value *values.Value) error {
	switch opType {
	case opcodes.IS_UNUSED:
		return nil
	case opcodes.IS_TMP_VAR:
		frame.setTemp(operand, value)
		ctx.setTemporary(operand, value)
		return nil
	case opcodes.IS_VAR, opcodes.IS_CV:
		frame.setLocal(operand, value)
		if globalName, ok := frame.globalSlotName(operand); ok {
			ctx.bindGlobalValue(globalName, value)
		}
		ctx.recordAssignment(frame, operand, value)
		if name, ok := frame.SlotNames[operand]; ok {
			if _, watched := vm.watchVars[name]; watched {
				vm.recordDebug(ctx, fmt.Sprintf("watch %s = %s", name, value.String()))
			}
			ctx.setVariable(name, value)
		}
		return nil
	case opcodes.IS_CONST:
		if frame == nil || int(operand) >= len(frame.Constants) {
			return fmt.Errorf("constant index %d out of range", operand)
		}
		constVal := frame.Constants[operand]
		if ctx != nil && ctx.currentClass != nil && constVal.IsString() {
			propName := constVal.ToString()
			if prop, ok := ctx.currentClass.Properties[propName]; ok {
				prop.Default = copyValue(value)
				if prop.IsStatic {
					if ctx.currentClass.StaticProps == nil {
						ctx.currentClass.StaticProps = make(map[string]*values.Value)
					}
					ctx.currentClass.StaticProps[propName] = copyValue(value)
				}
				return nil
			}
		}
		return fmt.Errorf("cannot write to operand type %d", opType)
	default:
		return fmt.Errorf("cannot write to operand type %d", opType)
	}
}

func copyValue(val *values.Value) *values.Value {
	if val == nil {
		return values.NewNull()
	}
	switch val.Type {
	case values.TypeNull:
		return values.NewNull()
	case values.TypeBool:
		return values.NewBool(val.Data.(bool))
	case values.TypeInt:
		return values.NewInt(val.Data.(int64))
	case values.TypeFloat:
		return values.NewFloat(val.Data.(float64))
	case values.TypeString:
		return values.NewString(val.Data.(string))
	default:
		return val
	}
}

func (vm *VirtualMachine) writableOperand(frame *CallFrame, opType opcodes.OpType, operand uint32) (*values.Value, error) {
	switch opType {
	case opcodes.IS_VAR, opcodes.IS_CV:
		return frame.ensureLocal(operand), nil
	case opcodes.IS_TMP_VAR:
		return frame.ensureTemp(operand), nil
	case opcodes.IS_UNUSED:
		return values.NewNull(), nil
	default:
		return nil, fmt.Errorf("operand type %d not writable", opType)
	}
}

func ensureArrayValue(val *values.Value) *values.Value {
	if val == nil {
		return values.NewArray()
	}
	actual := val
	if actual.IsReference() {
		actual = actual.Deref()
	}
	if !actual.IsArray() {
		newArr := values.NewArray()
		*actual = *newArr
	}
	return actual
}

func ensureArrayElement(arr *values.Value, key *values.Value) *values.Value {
	if arr == nil {
		arr = values.NewArray()
	}
	if arr.IsReference() {
		arr = arr.Deref()
	}
	actual := arr.Data.(*values.Array)

	if key == nil || key.IsNull() {
		index := actual.NextIndex
		if elem, exists := actual.Elements[index]; exists {
			return elem
		}
		null := values.NewNull()
		actual.Elements[index] = null
		actual.NextIndex++
		return null
	}

	mapKey := normalizeArrayKey(key)
	if elem, exists := actual.Elements[mapKey]; exists {
		return elem
	}
	null := values.NewNull()
	actual.Elements[mapKey] = null
	if idx, ok := asInt64(mapKey); ok && idx >= actual.NextIndex {
		actual.NextIndex = idx + 1
	}
	return null
}

func normalizeArrayKey(key *values.Value) interface{} {
	if key == nil {
		return ""
	}
	val := key.Deref()
	switch val.Type {
	case values.TypeNull:
		return ""
	case values.TypeBool:
		if val.Data.(bool) {
			return int64(1)
		}
		return int64(0)
	case values.TypeInt:
		return val.Data.(int64)
	case values.TypeFloat:
		return int64(val.Data.(float64))
	case values.TypeString:
		s := val.Data.(string)
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			return i
		}
		return s
	default:
		return val.ToString()
	}
}

func resolveRuntimeClassName(ctx *ExecutionContext, frame *CallFrame, raw string) (string, error) {
	if raw == "" {
		return raw, nil
	}
	lower := strings.ToLower(raw)
	switch lower {
	case "self", "static":
		if frame != nil && frame.ClassName != "" {
			return frame.ClassName, nil
		}
		if ctx != nil && ctx.currentClass != nil {
			return ctx.currentClass.Name, nil
		}
		return "", fmt.Errorf("cannot resolve %s outside class context", raw)
	case "parent":
		var current string
		if frame != nil && frame.ClassName != "" {
			current = frame.ClassName
		} else if ctx != nil && ctx.currentClass != nil {
			current = ctx.currentClass.Name
		}
		if current == "" {
			return "", fmt.Errorf("cannot resolve parent outside class context")
		}
		cls := ctx.ensureClass(current)
		if cls == nil || cls.Parent == "" {
			return "", fmt.Errorf("class %s has no parent", current)
		}
		return cls.Parent, nil
	default:
		return raw, nil
	}
}

func lookupClassConstantValue(ctx *ExecutionContext, cls *classRuntime, name string) *values.Value {
	if cls == nil {
		return nil
	}
	if cls.Constants != nil {
		if val, ok := cls.Constants[name]; ok {
			return copyValue(val)
		}
	}
	if cls.Descriptor != nil && cls.Descriptor.Constants != nil {
		if val, ok := cls.Descriptor.Constants[name]; ok && val.Value != nil {
			return copyValue(val.Value)
		}
	}
	if cls.Parent != "" {
		parent := ctx.ensureClass(cls.Parent)
		return lookupClassConstantValue(ctx, parent, name)
	}
	return nil
}

func resolveClassMethod(ctx *ExecutionContext, cls *classRuntime, method string) *registry.Function {
	if cls == nil {
		return nil
	}
	methodLower := strings.ToLower(method)
	if cls.Descriptor != nil && cls.Descriptor.Methods != nil {
		if fn, ok := cls.Descriptor.Methods[method]; ok {
			return fn
		}
		if fn, ok := cls.Descriptor.Methods[methodLower]; ok {
			return fn
		}
		// Fallback: linear scan for case-insensitive match if map uses mixed keys
		for name, fn := range cls.Descriptor.Methods {
			if strings.ToLower(name) == methodLower {
				return fn
			}
		}
	}
	if cls.Parent != "" {
		parent := ctx.ensureClass(cls.Parent)
		return resolveClassMethod(ctx, parent, method)
	}
	return nil
}

// resolveClassMethodWithDefiningClass returns both the method and the class where it's defined
func resolveClassMethodWithDefiningClass(ctx *ExecutionContext, cls *classRuntime, method string) (*registry.Function, *classRuntime) {
	if cls == nil {
		return nil, nil
	}
	methodLower := strings.ToLower(method)
	if cls.Descriptor != nil && cls.Descriptor.Methods != nil {
		if fn, ok := cls.Descriptor.Methods[method]; ok {
			return fn, cls
		}
		if fn, ok := cls.Descriptor.Methods[methodLower]; ok {
			return fn, cls
		}
		// Fallback: linear scan for case-insensitive match if map uses mixed keys
		for name, fn := range cls.Descriptor.Methods {
			if strings.ToLower(name) == methodLower {
				return fn, cls
			}
		}
	}
	if cls.Parent != "" {
		parent := ctx.ensureClass(cls.Parent)
		return resolveClassMethodWithDefiningClass(ctx, parent, method)
	}
	return nil, nil
}


func instantiateObject(ctx *ExecutionContext, className string) (*values.Value, error) {
	cls := ctx.ensureClass(className)
	if cls != nil && cls.Descriptor != nil {
		// Check if the class is abstract
		if cls.Descriptor.IsAbstract {
			return nil, fmt.Errorf("cannot instantiate abstract class %s", className)
		}
	}

	obj := values.NewObject(className)
	if cls != nil {
		// Initialize properties from both runtime and descriptor
		if obj.Data.(*values.Object).Properties == nil {
			obj.Data.(*values.Object).Properties = make(map[string]*values.Value)
		}

		// Initialize from runtime properties
		if cls.Properties != nil {
			for name, prop := range cls.Properties {
				if prop.IsStatic {
					continue
				}
				obj.Data.(*values.Object).Properties[name] = copyValue(prop.Default)
			}
		}

		// Initialize from descriptor properties (for builtin classes like Exception)
		if cls.Descriptor != nil && cls.Descriptor.Properties != nil {
			for name, propDesc := range cls.Descriptor.Properties {
				if propDesc.IsStatic {
					continue
				}
				// Only set if not already set by runtime properties
				if _, exists := obj.Data.(*values.Object).Properties[name]; !exists {
					obj.Data.(*values.Object).Properties[name] = copyValue(propDesc.DefaultValue)
				}
			}
		}
	}
	return obj, nil
}

func (vm *VirtualMachine) execAssignException(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	// Get the variable slot to assign the exception to
	opType1, op1 := decodeOperand(inst, 1)
	if opType1 != opcodes.IS_CV {
		return false, fmt.Errorf("ASSIGN_EXCEPTION requires CV operand")
	}

	// Get the pending exception from the current frame
	if frame.pendingException == nil {
		return false, fmt.Errorf("no pending exception to assign")
	}


	// Assign the exception to the specified variable slot
	if err := vm.writeOperand(ctx, frame, opType1, op1, frame.pendingException); err != nil {
		return false, err
	}

	return true, nil
}

func (vm *VirtualMachine) execExceptionMatch(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	// Get the exception type name to match
	opType1, op1 := decodeOperand(inst, 1)
	typeName, err := vm.readOperand(ctx, frame, opType1, op1)
	if err != nil {
		return false, err
	}

	// Check if exception matches the type
	var isMatch bool
	if frame.pendingException == nil {
		// No exception, return false
		isMatch = false
	} else {
		isMatch = vm.exceptionMatchesType(frame.pendingException, typeName.ToString())
	}

	// Store result in the specified target location (like comparison operations)
	result := values.NewBool(isMatch)
	resType, resSlot := decodeResult(inst)
	if err := vm.writeOperand(ctx, frame, resType, resSlot, result); err != nil {
		return false, err
	}

	return true, nil
}

func (vm *VirtualMachine) execClearException(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	// Clear the pending exception when entering a catch block
	frame.pendingException = nil
	return true, nil
}

func (vm *VirtualMachine) execRethrow(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	// Re-throw the current pending exception
	if frame.pendingException != nil {
		return vm.raiseException(ctx, frame, frame.pendingException)
	}
	// If no exception to rethrow, continue
	return true, nil
}

func (vm *VirtualMachine) exceptionMatchesType(exception *values.Value, typeName string) bool {
	// Check if the exception object matches the given type name
	if exception == nil || exception.Type != values.TypeObject {
		return false
	}

	obj := exception.Data.(*values.Object)
	if obj == nil {
		return false
	}

	// Check for exact match or inheritance chain
	if registry.GlobalRegistry != nil {
		return registry.GlobalRegistry.IsInstanceOf(obj.ClassName, typeName)
	}
	return false
}

func (vm *VirtualMachine) execInstanceof(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	// Get the object to check (left operand)
	opType1, op1 := decodeOperand(inst, 1)
	objectVal, err := vm.readOperand(ctx, frame, opType1, op1)
	if err != nil {
		return false, err
	}

	// Get the class name to check against (right operand)
	opType2, op2 := decodeOperand(inst, 2)
	classVal, err := vm.readOperand(ctx, frame, opType2, op2)
	if err != nil {
		return false, err
	}

	targetClassName := classVal.ToString()
	isMatch := false

	// Check if the object is an instance of the specified class
	if objectVal != nil && objectVal.IsObject() {
		obj := objectVal.Data.(*values.Object)

		// Exact class name matching
		if strings.EqualFold(obj.ClassName, targetClassName) {
			isMatch = true
		} else {
			// Check inheritance hierarchy
			currentClass := ctx.ensureClass(obj.ClassName)
			for currentClass != nil && !isMatch {
				if strings.EqualFold(currentClass.Name, targetClassName) {
					isMatch = true
					break
				}
				// Check parent class
				if currentClass.Parent != "" {
					currentClass = ctx.ensureClass(currentClass.Parent)
				} else {
					break
				}
			}

			// Check interfaces implemented by the class
			if !isMatch && currentClass != nil && currentClass.Descriptor != nil {
				for _, iface := range currentClass.Descriptor.Interfaces {
					if strings.EqualFold(iface, targetClassName) {
						isMatch = true
						break
					}
				}
			}
		}
	}
	// Non-objects are never instances of classes

	// Store the result
	resultType, resultSlot := decodeResult(inst)
	result := values.NewBool(isMatch)
	if err := vm.writeOperand(ctx, frame, resultType, resultSlot, result); err != nil {
		return false, err
	}

	return true, nil
}

func resolveThis(frame *CallFrame, val *values.Value) *values.Value {
	if val != nil && val.IsObject() {
		return val
	}
	if frame != nil && frame.This != nil {
		return frame.This
	}
	return val
}

func (vm *VirtualMachine) execInitClassTable(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	opType1, op1 := decodeOperand(inst, 1)
	if opType1 != opcodes.IS_CONST {
		return false, fmt.Errorf("INIT_CLASS_TABLE requires class name constant")
	}
	if int(op1) >= len(frame.Constants) {
		return false, fmt.Errorf("class name constant %d out of range", op1)
	}
	className := frame.Constants[op1].ToString()
	cls := ctx.ensureClass(className)
	if cls != nil && cls.Descriptor == nil {
		if def, ok := ctx.UserClasses[strings.ToLower(className)]; ok {
			cls.Descriptor = def
		}
	}
	return true, nil
}

func (vm *VirtualMachine) execSetCurrentClass(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	opType1, op1 := decodeOperand(inst, 1)
	if opType1 != opcodes.IS_CONST {
		return false, fmt.Errorf("SET_CURRENT_CLASS requires class name constant")
	}
	if int(op1) >= len(frame.Constants) {
		return false, fmt.Errorf("class name constant %d out of range", op1)
	}
	className := frame.Constants[op1].ToString()
	cls := ctx.ensureClass(className)
	if cls != nil && cls.Descriptor == nil {
		if def, ok := ctx.UserClasses[strings.ToLower(className)]; ok {
			cls.Descriptor = def
		}
	}
	ctx.currentClass = cls
	return true, nil
}

func (vm *VirtualMachine) execDeclareProperty(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	classType, classOp := decodeOperand(inst, 1)
	propType, propOp := decodeOperand(inst, 2)
	metaType, metaOp := decodeResult(inst)
	if classType != opcodes.IS_CONST || propType != opcodes.IS_CONST || metaType != opcodes.IS_CONST {
		return false, fmt.Errorf("DECLARE_PROPERTY expects constant operands")
	}
	if int(classOp) >= len(frame.Constants) || int(propOp) >= len(frame.Constants) || int(metaOp) >= len(frame.Constants) {
		return false, fmt.Errorf("DECLARE_PROPERTY constant operand out of range")
	}
	className := frame.Constants[classOp].ToString()
	propName := frame.Constants[propOp].ToString()
	metadata := frame.Constants[metaOp]
	builder := ctx.ensureClass(className)
	if builder.Properties == nil {
		builder.Properties = make(map[string]*propertyRuntime)
	}
	visibility := "public"
	isStatic := false
	defaultValue := values.NewNull()
	if metadata != nil && metadata.IsArray() {
		metaArr := metadata.Data.(*values.Array)
		if v, ok := metaArr.Elements["visibility"]; ok && v != nil {
			visibility = v.ToString()
		}
		if v, ok := metaArr.Elements["static"]; ok && v != nil {
			isStatic = v.ToBool()
		}
		if v, ok := metaArr.Elements["defaultValue"]; ok && v != nil {
			defaultValue = copyValue(v)
		}
	}
	prop := &propertyRuntime{
		Visibility: visibility,
		IsStatic:   isStatic,
		Default:    copyValue(defaultValue),
	}
	builder.Properties[propName] = prop
	if isStatic {
		if builder.StaticProps == nil {
			builder.StaticProps = make(map[string]*values.Value)
		}
		builder.StaticProps[propName] = copyValue(defaultValue)
	}
	return true, nil
}

func (vm *VirtualMachine) execClearCurrentClass(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	ctx.currentClass = nil
	return true, nil
}

func (vm *VirtualMachine) execDeclareClass(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	opType1, op1 := decodeOperand(inst, 1)
	if opType1 != opcodes.IS_CONST {
		return false, fmt.Errorf("DECLARE_CLASS requires class name constant")
	}
	if int(op1) >= len(frame.Constants) {
		return false, fmt.Errorf("class name constant %d out of range", op1)
	}
	className := frame.Constants[op1].ToString()
	cls := ctx.ensureClass(className)
	if cls != nil && cls.Descriptor == nil {
		if def, ok := ctx.UserClasses[strings.ToLower(className)]; ok {
			cls.Descriptor = def
		}
	}
	if cls != nil && cls.StaticProps == nil {
		cls.StaticProps = make(map[string]*values.Value)
	}
	return true, nil
}

func (vm *VirtualMachine) execSetClassParent(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	classType, classOp := decodeOperand(inst, 1)
	parentType, parentOp := decodeOperand(inst, 2)
	if classType != opcodes.IS_CONST || parentType != opcodes.IS_CONST {
		return false, fmt.Errorf("SET_CLASS_PARENT expects constant operands")
	}
	if int(classOp) >= len(frame.Constants) || int(parentOp) >= len(frame.Constants) {
		return false, fmt.Errorf("SET_CLASS_PARENT constant operand out of range")
	}
	childName := frame.Constants[classOp].ToString()
	parentName := frame.Constants[parentOp].ToString()
	child := ctx.ensureClass(childName)
	child.Parent = parentName
	if child.Descriptor != nil {
		child.Descriptor.Parent = parentName
	}
	parent := ctx.ensureClass(parentName)
	inheritClassMetadata(child, parent)
	return true, nil
}

func (vm *VirtualMachine) execDeclareConstant(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	classType, classOp := decodeOperand(inst, 1)
	nameType, nameOp := decodeOperand(inst, 2)
	valueType, valueOp := decodeResult(inst)
	if classType != opcodes.IS_CONST || nameType != opcodes.IS_CONST || valueType != opcodes.IS_CONST {
		return false, fmt.Errorf("DECLARE_CONSTANT expects constant operands")
	}
	if int(classOp) >= len(frame.Constants) || int(nameOp) >= len(frame.Constants) || int(valueOp) >= len(frame.Constants) {
		return false, fmt.Errorf("DECLARE_CONSTANT constant operand out of range")
	}
	className := frame.Constants[classOp].ToString()
	constName := frame.Constants[nameOp].ToString()
	constVal := copyValue(frame.Constants[valueOp])
	cls := ctx.ensureClass(className)
	if cls.Constants == nil {
		cls.Constants = make(map[string]*values.Value)
	}
	cls.Constants[constName] = constVal
	if cls.Descriptor != nil {
		if cls.Descriptor.Constants == nil {
			cls.Descriptor.Constants = make(map[string]*registry.ClassConstant)
		}
		if existing, ok := cls.Descriptor.Constants[constName]; ok {
			existing.Value = copyValue(constVal)
		} else {
			cls.Descriptor.Constants[constName] = &registry.ClassConstant{
				Name:  constName,
				Value: copyValue(constVal),
			}
		}
	}
	if registry.GlobalRegistry != nil && cls.Descriptor != nil {
		if desc := descriptorFromClass(cls.Descriptor); desc != nil {
			_ = registry.GlobalRegistry.RegisterClass(desc)
		}
	}
	return true, nil
}

func (vm *VirtualMachine) execAssignObj(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	objType, objOp := decodeOperand(inst, 1)
	objVal, err := vm.readOperand(ctx, frame, objType, objOp)
	if err != nil {
		return false, err
	}
	objVal = resolveThis(frame, objVal)
	if objVal == nil || !objVal.IsObject() {
		return false, fmt.Errorf("ASSIGN_OBJ target is not object")
	}
	propType, propOp := decodeOperand(inst, 2)
	propVal, err := vm.readOperand(ctx, frame, propType, propOp)
	if err != nil {
		return false, err
	}
	propName := propVal.ToString()
	valueType, valueOp := decodeResult(inst)
	value, err := vm.readOperand(ctx, frame, valueType, valueOp)
	if err != nil {
		return false, err
	}
	obj := objVal.Data.(*values.Object)
	obj.Properties[propName] = copyValue(value)
	return true, nil
}

func (vm *VirtualMachine) execAssignObjOp(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	// Read object (operand 1)
	objType, objOp := decodeOperand(inst, 1)
	objVal, err := vm.readOperand(ctx, frame, objType, objOp)
	if err != nil {
		return false, err
	}
	objVal = resolveThis(frame, objVal)
	if objVal == nil || !objVal.IsObject() {
		return false, fmt.Errorf("ASSIGN_OBJ_OP target is not object")
	}

	// Read property name (operand 2)
	propType, propOp := decodeOperand(inst, 2)
	propVal, err := vm.readOperand(ctx, frame, propType, propOp)
	if err != nil {
		return false, err
	}
	propName := propVal.ToString()

	// Read right-hand side value (result operand)
	valueType, valueOp := decodeResult(inst)
	right, err := vm.readOperand(ctx, frame, valueType, valueOp)
	if err != nil {
		return false, err
	}

	// Get current property value (left operand)
	obj := objVal.Data.(*values.Object)
	left, exists := obj.Properties[propName]
	if !exists {
		left = values.NewNull()
	}

	// Perform the operation based on inst.Reserved
	var result *values.Value
	switch inst.Reserved {
	case 1: // +=
		result = left.Add(right)
	case 2: // -=
		result = left.Subtract(right)
	case 3: // *=
		result = left.Multiply(right)
	case 4: // /=
		result = left.Divide(right)
	case 5: // %=
		result = left.Modulo(right)
	case 6: // **=
		result = left.Power(right)
	case 8: // .=
		result = left.Concat(right)
	case 9: // &=
		result = values.NewInt(left.ToInt() & right.ToInt())
	case 10: // |=
		result = values.NewInt(left.ToInt() | right.ToInt())
	case 11: // ^=
		result = values.NewInt(left.ToInt() ^ right.ToInt())
	case 12: // <<=
		result = values.NewInt(left.ToInt() << right.ToInt())
	case 13: // >>=
		result = values.NewInt(left.ToInt() >> right.ToInt())
	default:
		return false, fmt.Errorf("unsupported object assignment op code %d", inst.Reserved)
	}

	// Store the result back to the object property
	obj.Properties[propName] = copyValue(result)
	return true, nil
}

func (vm *VirtualMachine) execFetchObj(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	objType, objOp := decodeOperand(inst, 1)
	objVal, err := vm.readOperand(ctx, frame, objType, objOp)
	if err != nil {
		return false, err
	}
	objVal = resolveThis(frame, objVal)
	if objVal == nil || !objVal.IsObject() {
		return false, fmt.Errorf("FETCH_OBJ_R target is not object")
	}
	propType, propOp := decodeOperand(inst, 2)
	propVal, err := vm.readOperand(ctx, frame, propType, propOp)
	if err != nil {
		return false, err
	}
	propName := propVal.ToString()
	obj := objVal.Data.(*values.Object)
	if val, exists := obj.Properties[propName]; exists {
		resType, resSlot := decodeResult(inst)
		return vm.writeOperand(ctx, frame, resType, resSlot, copyValue(val)) == nil, nil
	}
	if cls, ok := ctx.getClass(obj.ClassName); ok {
		if prop, ok := cls.Properties[propName]; ok && !prop.IsStatic {
			obj.Properties[propName] = copyValue(prop.Default)
			resType, resSlot := decodeResult(inst)
			return vm.writeOperand(ctx, frame, resType, resSlot, copyValue(prop.Default)) == nil, nil
		}
	}
	resType, resSlot := decodeResult(inst)
	return vm.writeOperand(ctx, frame, resType, resSlot, values.NewNull()) == nil, nil
}

func (vm *VirtualMachine) execFetchObjIs(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	objType, objOp := decodeOperand(inst, 1)
	objVal, err := vm.readOperand(ctx, frame, objType, objOp)
	if err != nil {
		return false, err
	}
	objVal = resolveThis(frame, objVal)
	if objVal == nil || !objVal.IsObject() {
		return false, fmt.Errorf("FETCH_OBJ_IS target is not object")
	}
	propType, propOp := decodeOperand(inst, 2)
	propVal, err := vm.readOperand(ctx, frame, propType, propOp)
	if err != nil {
		return false, err
	}
	propName := propVal.ToString()
	obj := objVal.Data.(*values.Object)
	exists := false
	if val, ok := obj.Properties[propName]; ok {
		exists = !val.Deref().IsNull()
	} else if cls, ok := ctx.getClass(obj.ClassName); ok {
		if prop, ok := cls.Properties[propName]; ok && !prop.IsStatic {
			exists = prop.Default != nil && !prop.Default.Deref().IsNull()
		}
	}
	resType, resSlot := decodeResult(inst)
	if err := vm.writeOperand(ctx, frame, resType, resSlot, values.NewBool(exists)); err != nil {
		return false, err
	}
	return true, nil
}

func (vm *VirtualMachine) execFetchStaticProp(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	classType, classOp := decodeOperand(inst, 1)
	classVal, err := vm.readOperand(ctx, frame, classType, classOp)
	if err != nil {
		return false, err
	}
	className, err := resolveRuntimeClassName(ctx, frame, classVal.ToString())
	if err != nil {
		return false, err
	}
	propType, propOp := decodeOperand(inst, 2)
	propVal, err := vm.readOperand(ctx, frame, propType, propOp)
	if err != nil {
		return false, err
	}
	propName := propVal.ToString()
	propName = strings.TrimPrefix(propName, "$")
	cls := ctx.ensureClass(className)
	var result *values.Value
	if cls != nil {
		if cls.StaticProps != nil {
			if val, exists := cls.StaticProps[propName]; exists {
				result = copyValue(val)
			}
		}
		if result == nil {
			if prop, exists := cls.Properties[propName]; exists && prop.IsStatic {
				if cls.StaticProps == nil {
					cls.StaticProps = make(map[string]*values.Value)
				}
				if val, ok := cls.StaticProps[propName]; ok {
					result = copyValue(val)
				} else {
					defaultVal := copyValue(prop.Default)
					cls.StaticProps[propName] = defaultVal
					result = copyValue(prop.Default)
				}
			}
		}
	}
	if result == nil {
		result = values.NewNull()
	}
	resType, resSlot := decodeResult(inst)
	return vm.writeOperand(ctx, frame, resType, resSlot, result) == nil, nil
}

func (vm *VirtualMachine) execFetchStaticPropWrite(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	classType, classOp := decodeOperand(inst, 1)
	classVal, err := vm.readOperand(ctx, frame, classType, classOp)
	if err != nil {
		return false, err
	}
	className, err := resolveRuntimeClassName(ctx, frame, classVal.ToString())
	if err != nil {
		return false, err
	}
	propType, propOp := decodeOperand(inst, 2)
	propVal, err := vm.readOperand(ctx, frame, propType, propOp)
	if err != nil {
		return false, err
	}
	propName := strings.TrimPrefix(propVal.ToString(), "$")
	cls := ctx.ensureClass(className)
	if cls == nil {
		return false, fmt.Errorf("class %s not found", className)
	}
	if cls.StaticProps == nil {
		cls.StaticProps = make(map[string]*values.Value)
	}
	slot, exists := cls.StaticProps[propName]
	if !exists {
		if prop, ok := cls.Properties[propName]; ok && prop.IsStatic {
			slot = copyValue(prop.Default)
		} else {
			slot = values.NewNull()
		}
		cls.StaticProps[propName] = slot
	}
	resType, resSlot := decodeResult(inst)
	value, err := vm.readOperand(ctx, frame, resType, resSlot)
	if err != nil {
		return false, err
	}
	assigned := copyValue(value)
	*slot = *assigned
	cls.StaticProps[propName] = slot
	if err := vm.writeOperand(ctx, frame, resType, resSlot, slot); err != nil {
		return false, err
	}
	return true, nil
}

func (vm *VirtualMachine) execFetchClassConstant(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	classType, classOp := decodeOperand(inst, 1)
	classVal, err := vm.readOperand(ctx, frame, classType, classOp)
	if err != nil {
		return false, err
	}
	if classVal != nil {
		classVal = classVal.Deref()
	}
	var rawClass string
	if classVal != nil && classVal.IsObject() {
		rawClass = classVal.Data.(*values.Object).ClassName
	} else if classVal != nil {
		rawClass = classVal.ToString()
	}
	if strings.HasPrefix(rawClass, "$") {
		lookup := rawClass[1:]
		candidates := []string{lookup, rawClass}
		resolved := false
		for _, candidate := range candidates {
			if slot, ok := frame.slotByName(candidate); ok {
				val := frame.getLocal(slot)
				if val != nil {
					val = val.Deref()
					if val.IsObject() {
						rawClass = val.Data.(*values.Object).ClassName
					} else {
						rawClass = val.ToString()
					}
					resolved = true
					break
				}
			}
		}
		if !resolved && ctx.Variables != nil {
			for _, candidate := range candidates {
				if val, ok := ctx.Variables[candidate]; ok && val != nil {
					val = val.Deref()
					if val.IsObject() {
						rawClass = val.Data.(*values.Object).ClassName
					} else {
						rawClass = val.ToString()
					}
					resolved = true
					break
				}
			}
		}
	}
	if rawClass == "" {
		return false, fmt.Errorf("invalid class reference for constant fetch")
	}
	className, err := resolveRuntimeClassName(ctx, frame, rawClass)
	if err != nil {
		return false, err
	}
	constType, constOp := decodeOperand(inst, 2)
	constVal, err := vm.readOperand(ctx, frame, constType, constOp)
	if err != nil {
		return false, err
	}
	if constVal != nil {
		constVal = constVal.Deref()
	}
	constName := constVal.ToString()
	cls := ctx.ensureClass(className)
	if cls == nil {
		return false, fmt.Errorf("class %s not found", className)
	}
	result := lookupClassConstantValue(ctx, cls, constName)
	if result == nil {
		return false, fmt.Errorf("undefined class constant %s::%s", className, constName)
	}
	resType, resSlot := decodeResult(inst)
	if err := vm.writeOperand(ctx, frame, resType, resSlot, result); err != nil {
		return false, err
	}
	return true, nil
}

func (vm *VirtualMachine) execFetchLateStaticConstant(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	classType, classOp := decodeOperand(inst, 1)
	classVal, err := vm.readOperand(ctx, frame, classType, classOp)
	if err != nil {
		return false, err
	}
	if classVal != nil {
		classVal = classVal.Deref()
	}

	var rawClass string
	if classVal != nil {
		rawClass = classVal.ToString()
	}

	// For late static binding, "static" should resolve to the calling class
	var className string
	if rawClass == "static" {
		// Use the calling class from the call frame for late static binding
		if frame.CallingClass != "" {
			className = frame.CallingClass
		} else {
			// Fallback to current class if no calling class is set
			className = frame.ClassName
		}
	} else {
		// For non-static references, resolve normally
		var err error
		className, err = resolveRuntimeClassName(ctx, frame, rawClass)
		if err != nil {
			return false, err
		}
	}

	if className == "" {
		return false, fmt.Errorf("invalid class reference for late static constant fetch")
	}

	constType, constOp := decodeOperand(inst, 2)
	constVal, err := vm.readOperand(ctx, frame, constType, constOp)
	if err != nil {
		return false, err
	}
	if constVal != nil {
		constVal = constVal.Deref()
	}
	constName := constVal.ToString()

	cls := ctx.ensureClass(className)
	if cls == nil {
		return false, fmt.Errorf("class %s not found", className)
	}

	result := lookupClassConstantValue(ctx, cls, constName)
	if result == nil {
		return false, fmt.Errorf("undefined class constant %s::%s", className, constName)
	}

	resType, resSlot := decodeResult(inst)
	if err := vm.writeOperand(ctx, frame, resType, resSlot, result); err != nil {
		return false, err
	}
	return true, nil
}

func makeKeyValue(key interface{}) *values.Value {
	switch k := key.(type) {
	case string:
		return values.NewString(k)
	case int:
		return values.NewInt(int64(k))
	case int64:
		return values.NewInt(k)
	case uint32:
		return values.NewInt(int64(k))
	case uint64:
		return values.NewInt(int64(k))
	default:
		return values.NewString(fmt.Sprintf("%v", k))
	}
}

func asInt64(key interface{}) (int64, bool) {
	switch k := key.(type) {
	case int:
		return int64(k), true
	case int8:
		return int64(k), true
	case int16:
		return int64(k), true
	case int32:
		return int64(k), true
	case int64:
		return k, true
	case uint:
		return int64(k), true
	case uint8:
		return int64(k), true
	case uint16:
		return int64(k), true
	case uint32:
		return int64(k), true
	case uint64:
		return int64(k), true
	default:
		return 0, false
	}
}

func (vm *VirtualMachine) execAssign(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction, isVarAssign bool) (bool, error) {
	opType1, op1 := decodeOperand(inst, 1)
	value, err := vm.readOperand(ctx, frame, opType1, op1)
	if err != nil {
		return false, err
	}
	if isVarAssign {
		value = copyValue(value)
	}
	resType, resSlot := decodeResult(inst)
	if err := vm.writeOperand(ctx, frame, resType, resSlot, value); err != nil {
		return false, err
	}
	return true, nil
}

func (vm *VirtualMachine) execAssignRef(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	opType1, op1 := decodeOperand(inst, 1)
	val, err := vm.readOperand(ctx, frame, opType1, op1)
	if err != nil {
		return false, err
	}
	ref := values.NewReference(val)
	resType, resSlot := decodeResult(inst)
	if err := vm.writeOperand(ctx, frame, resType, resSlot, ref); err != nil {
		return false, err
	}
	return true, nil
}

func (vm *VirtualMachine) execAssignOp(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	opType1, op1 := decodeOperand(inst, 1)
	left, err := vm.readOperand(ctx, frame, opType1, op1)
	if err != nil {
		return false, err
	}

	opType2, op2 := decodeOperand(inst, 2)
	right, err := vm.readOperand(ctx, frame, opType2, op2)
	if err != nil {
		return false, err
	}

	var result *values.Value
	switch inst.Reserved {
	case 1:
		result = left.Add(right)
	case 2:
		result = left.Subtract(right)
	case 3:
		result = left.Multiply(right)
	case 4:
		result = left.Divide(right)
	case 5:
		result = left.Modulo(right)
	case 6:
		result = left.Power(right)
	case 8:
		result = left.Concat(right)
	case 9:
		result = values.NewInt(left.ToInt() & right.ToInt())
	case 10:
		result = values.NewInt(left.ToInt() | right.ToInt())
	case 11:
		result = values.NewInt(left.ToInt() ^ right.ToInt())
	case 12:
		result = values.NewInt(left.ToInt() << right.ToInt())
	case 13:
		result = values.NewInt(left.ToInt() >> right.ToInt())
	default:
		return false, fmt.Errorf("unsupported assignment op code %d", inst.Reserved)
	}

	resType, resSlot := decodeResult(inst)
	if err := vm.writeOperand(ctx, frame, resType, resSlot, result); err != nil {
		return false, err
	}
	return true, nil
}

func (vm *VirtualMachine) execArithmetic(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	opType1, op1 := decodeOperand(inst, 1)
	left, err := vm.readOperand(ctx, frame, opType1, op1)
	if err != nil {
		return false, err
	}

	opType2, op2 := decodeOperand(inst, 2)
	right, err := vm.readOperand(ctx, frame, opType2, op2)
	if err != nil {
		return false, err
	}

	var result *values.Value
	switch inst.Opcode {
	case opcodes.OP_ADD:
		result = left.Add(right)
	case opcodes.OP_SUB:
		result = left.Subtract(right)
	case opcodes.OP_MUL:
		result = left.Multiply(right)
	case opcodes.OP_DIV:
		result = left.Divide(right)
	case opcodes.OP_MOD:
		result = left.Modulo(right)
	case opcodes.OP_POW:
		result = left.Power(right)
	default:
		return false, fmt.Errorf("unsupported arithmetic opcode %s", inst.Opcode)
	}

	resType, resSlot := decodeResult(inst)
	if err := vm.writeOperand(ctx, frame, resType, resSlot, result); err != nil {
		return false, err
	}
	return true, nil
}

func (vm *VirtualMachine) execUnary(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	opType1, op1 := decodeOperand(inst, 1)
	val, err := vm.readOperand(ctx, frame, opType1, op1)
	if err != nil {
		return false, err
	}

	var result *values.Value
	switch inst.Opcode {
	case opcodes.OP_PLUS:
		// Unary plus returns numeric value unchanged
		if val.IsFloat() {
			result = values.NewFloat(val.ToFloat())
		} else {
			result = values.NewInt(val.ToInt())
		}
	case opcodes.OP_MINUS:
		if val.IsFloat() {
			result = values.NewFloat(-val.ToFloat())
		} else {
			result = values.NewInt(-val.ToInt())
		}
	case opcodes.OP_NOT:
		result = values.NewBool(!val.ToBool())
	default:
		return false, fmt.Errorf("unsupported unary opcode %s", inst.Opcode)
	}

	resType, resSlot := decodeResult(inst)
	if err := vm.writeOperand(ctx, frame, resType, resSlot, result); err != nil {
		return false, err
	}
	return true, nil
}

func (vm *VirtualMachine) execBitwise(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	opType1, op1 := decodeOperand(inst, 1)
	left, err := vm.readOperand(ctx, frame, opType1, op1)
	if err != nil {
		return false, err
	}

	opType2, op2 := decodeOperand(inst, 2)
	right, err := vm.readOperand(ctx, frame, opType2, op2)
	if err != nil {
		return false, err
	}

	leftInt := left.ToInt()
	rightInt := right.ToInt()

	var result int64
	switch inst.Opcode {
	case opcodes.OP_BW_AND:
		result = leftInt & rightInt
	case opcodes.OP_BW_OR:
		result = leftInt | rightInt
	case opcodes.OP_BW_XOR:
		result = leftInt ^ rightInt
	case opcodes.OP_SL:
		result = leftInt << rightInt
	case opcodes.OP_SR:
		result = leftInt >> rightInt
	default:
		return false, fmt.Errorf("unsupported bitwise opcode %s", inst.Opcode)
	}

	resType, resSlot := decodeResult(inst)
	if err := vm.writeOperand(ctx, frame, resType, resSlot, values.NewInt(result)); err != nil {
		return false, err
	}
	return true, nil
}

func (vm *VirtualMachine) execConcat(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	opType1, op1 := decodeOperand(inst, 1)
	left, err := vm.readOperand(ctx, frame, opType1, op1)
	if err != nil {
		return false, err
	}
	opType2, op2 := decodeOperand(inst, 2)
	right, err := vm.readOperand(ctx, frame, opType2, op2)
	if err != nil {
		return false, err
	}
	result := left.Concat(right)
	resType, resSlot := decodeResult(inst)
	if err := vm.writeOperand(ctx, frame, resType, resSlot, result); err != nil {
		return false, err
	}
	return true, nil
}


func (vm *VirtualMachine) execComparison(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	opType1, op1 := decodeOperand(inst, 1)
	left, err := vm.readOperand(ctx, frame, opType1, op1)
	if err != nil {
		return false, err
	}
	opType2, op2 := decodeOperand(inst, 2)
	right, err := vm.readOperand(ctx, frame, opType2, op2)
	if err != nil {
		return false, err
	}

	var result *values.Value
	switch inst.Opcode {
	case opcodes.OP_IS_EQUAL:
		result = values.NewBool(left.Equal(right))
	case opcodes.OP_IS_NOT_EQUAL:
		result = values.NewBool(!left.Equal(right))
	case opcodes.OP_IS_IDENTICAL:
		result = values.NewBool(left.Identical(right))
	case opcodes.OP_IS_NOT_IDENTICAL:
		result = values.NewBool(!left.Identical(right))
	case opcodes.OP_IS_SMALLER:
		result = values.NewBool(left.Compare(right) < 0)
	case opcodes.OP_IS_SMALLER_OR_EQUAL:
		result = values.NewBool(left.Compare(right) <= 0)
	case opcodes.OP_IS_GREATER:
		result = values.NewBool(left.Compare(right) > 0)
	case opcodes.OP_IS_GREATER_OR_EQUAL:
		result = values.NewBool(left.Compare(right) >= 0)
	case opcodes.OP_SPACESHIP:
		result = values.NewInt(int64(left.Compare(right)))
	default:
		return false, fmt.Errorf("unsupported comparison opcode %s", inst.Opcode)
	}

	resType, resSlot := decodeResult(inst)
	if err := vm.writeOperand(ctx, frame, resType, resSlot, result); err != nil {
		return false, err
	}
	return true, nil
}

func (vm *VirtualMachine) execBoolean(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	opType1, op1 := decodeOperand(inst, 1)
	left, err := vm.readOperand(ctx, frame, opType1, op1)
	if err != nil {
		return false, err
	}
	opType2, op2 := decodeOperand(inst, 2)
	right, err := vm.readOperand(ctx, frame, opType2, op2)
	if err != nil {
		return false, err
	}

	var result bool
	switch inst.Opcode {
	case opcodes.OP_BOOLEAN_AND:
		result = left.ToBool() && right.ToBool()
	case opcodes.OP_BOOLEAN_OR:
		result = left.ToBool() || right.ToBool()
	default:
		return false, fmt.Errorf("unsupported boolean opcode %s", inst.Opcode)
	}

	resType, resSlot := decodeResult(inst)
	if err := vm.writeOperand(ctx, frame, resType, resSlot, values.NewBool(result)); err != nil {
		return false, err
	}
	return true, nil
}

func (vm *VirtualMachine) execLogical(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	opType1, op1 := decodeOperand(inst, 1)
	left, err := vm.readOperand(ctx, frame, opType1, op1)
	if err != nil {
		return false, err
	}
	opType2, op2 := decodeOperand(inst, 2)
	right, err := vm.readOperand(ctx, frame, opType2, op2)
	if err != nil {
		return false, err
	}

	var result bool
	switch inst.Opcode {
	case opcodes.OP_LOGICAL_AND:
		result = left.ToBool() && right.ToBool()
	case opcodes.OP_LOGICAL_OR:
		result = left.ToBool() || right.ToBool()
	case opcodes.OP_LOGICAL_XOR:
		result = left.ToBool() != right.ToBool()
	default:
		return false, fmt.Errorf("unsupported logical opcode %s", inst.Opcode)
	}

	resType, resSlot := decodeResult(inst)
	if err := vm.writeOperand(ctx, frame, resType, resSlot, values.NewBool(result)); err != nil {
		return false, err
	}
	return true, nil
}

func (vm *VirtualMachine) execIncDec(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	opType1, op1 := decodeOperand(inst, 1)
	target, err := vm.readOperand(ctx, frame, opType1, op1)
	if err != nil {
		return false, err
	}

	var updated *values.Value
	switch inst.Opcode {
	case opcodes.OP_PRE_INC, opcodes.OP_POST_INC:
		updated = target.Add(values.NewInt(1))
	case opcodes.OP_PRE_DEC, opcodes.OP_POST_DEC:
		updated = target.Subtract(values.NewInt(1))
	default:
		return false, fmt.Errorf("unsupported inc/dec opcode %s", inst.Opcode)
	}

	resType, resSlot := decodeResult(inst)

	// Post operators return original value, pre operators return updated value.
	var writeVal *values.Value
	var resultVal *values.Value
	if inst.Opcode == opcodes.OP_POST_INC || inst.Opcode == opcodes.OP_POST_DEC {
		writeVal = updated
		resultVal = target
	} else {
		writeVal = updated
		resultVal = updated
	}

	if err := vm.writeOperand(ctx, frame, opType1, op1, writeVal); err != nil {
		return false, err
	}
	if err := vm.writeOperand(ctx, frame, resType, resSlot, resultVal); err != nil {
		return false, err
	}

	return true, nil
}

func (vm *VirtualMachine) execJump(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	opType1, op1 := decodeOperand(inst, 1)
	targetVal, err := vm.readOperand(ctx, frame, opType1, op1)
	if err != nil {
		return false, err
	}
	target := int(targetVal.ToInt())
	if handler := frame.peekExceptionHandler(); handler != nil {
		if handler.finallyIP == 0 && handler.catchIP != 0 && target > handler.catchIP {
			frame.popExceptionHandler()
		}
	}
	frame.IP = target
	return false, nil
}

func (vm *VirtualMachine) execBitwiseNot(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	opType1, op1 := decodeOperand(inst, 1)
	target, err := vm.readOperand(ctx, frame, opType1, op1)
	if err != nil {
		return false, err
	}
	result := ^target.ToInt()

	resType, resSlot := decodeResult(inst)
	if err := vm.writeOperand(ctx, frame, resType, resSlot, values.NewInt(result)); err != nil {
		return false, err
	}
	return true, nil
}

func (vm *VirtualMachine) execConditionalJump(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	opType1, op1 := decodeOperand(inst, 1)
	cond, err := vm.readOperand(ctx, frame, opType1, op1)
	if err != nil {
		return false, err
	}


	opType2, op2 := decodeOperand(inst, 2)
	targetVal, err := vm.readOperand(ctx, frame, opType2, op2)
	if err != nil {
		return false, err
	}

	// PHP-style truthiness check
	jump := false
	switch cond.Type {
	case values.TypeNull:
		jump = false
	case values.TypeBool:
		jump = cond.Data.(bool)
	case values.TypeInt:
		jump = cond.Data.(int64) != 0
	case values.TypeFloat:
		jump = cond.Data.(float64) != 0.0
	case values.TypeString:
		str := cond.Data.(string)
		jump = str != "" && str != "0"
	default:
		jump = true // Objects, arrays, etc. are truthy
	}

	if inst.Opcode == opcodes.OP_JMPZ {
		jump = !jump
	}

	if jump {
		frame.IP = int(targetVal.ToInt())
		return false, nil
	}
	return true, nil
}

func (vm *VirtualMachine) execFetch(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	opType1, op1 := decodeOperand(inst, 1)
	val, err := vm.readOperand(ctx, frame, opType1, op1)
	if err != nil {
		return false, err
	}
	resType, resSlot := decodeResult(inst)
	if err := vm.writeOperand(ctx, frame, resType, resSlot, copyValue(val)); err != nil {
		return false, err
	}
	return true, nil
}

func (vm *VirtualMachine) execFetchDynamic(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	opType1, op1 := decodeOperand(inst, 1)
	nameVal, err := vm.readOperand(ctx, frame, opType1, op1)
	if err != nil {
		return false, err
	}
	name := nameVal.ToString()
	candidates := make(map[string]struct{})
	addCandidate := func(n string) {
		if n == "" {
			return
		}
		candidates[n] = struct{}{}
	}
	addCandidate(name)
	if !strings.HasPrefix(name, "$") {
		addCandidate("$" + name)
	}
	if sanitized := sanitizeVariableName(name); sanitized != "" {
		addCandidate(sanitized)
		addCandidate("$" + sanitized)
	}
	var val *values.Value
	found := false
	for candidate := range candidates {
		if slot, ok := frame.slotByName(candidate); ok {
			val = frame.getLocal(slot)
			found = true
			break
		}
	}
	if !found && ctx.Variables != nil {
		for candidate := range candidates {
			if v, exists := ctx.Variables[candidate]; exists {
				val = v
				found = true
				break
			}
		}
	}
	if !found {
		resType, resSlot := decodeResult(inst)
		if err := vm.writeOperand(ctx, frame, resType, resSlot, values.NewNull()); err != nil {
			return false, err
		}
		return true, nil
	}
	if val != nil {
		val = val.Deref()
	}
	resType, resSlot := decodeResult(inst)
	if err := vm.writeOperand(ctx, frame, resType, resSlot, copyValue(val)); err != nil {
		return false, err
	}
	return true, nil
}

func (vm *VirtualMachine) execBindVarName(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	opType1, op1 := decodeOperand(inst, 1)
	if opType1 != opcodes.IS_VAR && opType1 != opcodes.IS_CV {
		return false, fmt.Errorf("bind var name requires variable operand")
	}
	opType2, op2 := decodeOperand(inst, 2)
	if opType2 != opcodes.IS_CONST {
		return false, fmt.Errorf("bind var name requires constant name")
	}
	if int(op2) >= len(frame.Constants) {
		return false, fmt.Errorf("constant index %d out of range", op2)
	}
	name := frame.Constants[op2].ToString()
	frame.bindSlotName(op1, name)
	localVal, exists := frame.getLocalWithStatus(op1)
	if frame.Function == nil {
		frame.bindGlobalSlot(op1, name)
		if !exists {
			localVal = ctx.ensureGlobal(name)
			exists = true
			frame.setLocal(op1, localVal)
		}
	}
	if val, ok := ctx.Variables[name]; ok && !exists {
		localVal = val
		exists = true
		frame.setLocal(op1, val)
	}
	if frame.Function == nil {
		ctx.bindGlobalValue(name, frame.getLocal(op1))
	}
	ctx.setVariable(name, frame.getLocal(op1))
	return true, nil
}

func (vm *VirtualMachine) execBindGlobal(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	if ctx == nil || frame == nil {
		return false, fmt.Errorf("bind global requires valid execution context")
	}
	nameType, nameOp := decodeOperand(inst, 1)
	if nameType != opcodes.IS_CONST {
		return false, fmt.Errorf("BIND_GLOBAL expects constant operand")
	}
	if int(nameOp) >= len(frame.Constants) {
		return false, fmt.Errorf("constant index %d out of range", nameOp)
	}
	name := frame.Constants[nameOp].ToString()
	resType, resSlot := decodeResult(inst)
	if resType != opcodes.IS_VAR && resType != opcodes.IS_CV {
		return false, fmt.Errorf("BIND_GLOBAL result must target variable slot")
	}
	frame.bindSlotName(resSlot, name)
	frame.bindGlobalSlot(resSlot, name)
	globalVal := ctx.ensureGlobal(name)
	frame.setLocal(resSlot, globalVal)
	ctx.bindGlobalValue(name, globalVal)
	ctx.setVariable(name, globalVal)
	return true, nil
}

func (vm *VirtualMachine) execDeclareFunction(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	// Functions are registered with the VM before execution begins, so this is a no-op.
	return true, nil
}

func (vm *VirtualMachine) execFeReset(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	opType1, op1 := decodeOperand(inst, 1)
	iterable, err := vm.readOperand(ctx, frame, opType1, op1)
	if err != nil {
		return false, err
	}

	resType, resSlot := decodeResult(inst)
	iteratorID := op1
	if resType != opcodes.IS_UNUSED {
		iteratorID = resSlot
	}
	if frame.Iterators == nil {
		frame.Iterators = make(map[uint32]*foreachIterator)
	}

	iterator := &foreachIterator{
		keys:      make([]*values.Value, 0),
		values:    make([]*values.Value, 0),
		index:     0,
		generator: nil,
		isFirst:   true,
	}

	if iterable != nil && iterable.IsArray() {
		arr := iterable.Data.(*values.Array)
		intKeys := make([]interface{}, 0)
		stringKeys := make([]interface{}, 0)
		otherKeys := make([]interface{}, 0)
		for key := range arr.Elements {
			if _, ok := asInt64(key); ok {
				intKeys = append(intKeys, key)
				continue
			}
			if _, ok := key.(string); ok {
				stringKeys = append(stringKeys, key)
				continue
			}
			otherKeys = append(otherKeys, key)
		}

		sort.Slice(intKeys, func(i, j int) bool {
			iVal, _ := asInt64(intKeys[i])
			jVal, _ := asInt64(intKeys[j])
			return iVal < jVal
		})
		sort.Slice(stringKeys, func(i, j int) bool {
			return stringKeys[i].(string) < stringKeys[j].(string)
		})

		orderedKeys := append(intKeys, stringKeys...)
		orderedKeys = append(orderedKeys, otherKeys...)

		for _, key := range orderedKeys {
			if val, ok := arr.Elements[key]; ok {
				iterator.keys = append(iterator.keys, makeKeyValue(key))
				iterator.values = append(iterator.values, copyValue(val))
			}
		}
	} else if iterable != nil && iterable.IsObject() {
		obj := iterable.Data.(*values.Object)
		// Check if this is a Generator object
		if obj.ClassName == "Generator" {
			// Store the generator object for use in FE_FETCH
			iterator.generator = iterable
		} else {
			// Handle other object types that might be iterable
			// TODO: Add support for other Iterator interface implementations
		}
	}

	frame.Iterators[iteratorID] = iterator
	if resType != opcodes.IS_UNUSED {
		if err := vm.writeOperand(ctx, frame, resType, resSlot, values.NewNull()); err != nil {
			return false, err
		}
	}

	return true, nil
}

func (vm *VirtualMachine) execFeFetch(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	opType1, op1 := decodeOperand(inst, 1)
	_ = opType1 // operand type currently not used beyond validation

	iterator := frame.Iterators[op1]

	resType, resSlot := decodeResult(inst)
	opType2, op2 := decodeOperand(inst, 2)

	nextValue := values.NewNull()
	nextKey := values.NewNull()

	if iterator != nil && iterator.generator != nil {
		// Handle Generator iteration
		obj := iterator.generator.Data.(*values.Object)

		// Get generator
		if genVal, ok := obj.Properties["__channel_generator"]; ok {
			if gen, ok := genVal.Data.(*runtime2.Generator); ok {
				if iterator.isFirst {
					iterator.isFirst = false
					// Start the generator and get first value
					if gen.Next() {
						nextValue = gen.Current()
						nextKey = gen.Key()
					} else {
						nextValue = values.NewNull()
						nextKey = values.NewNull()
					}
				} else {
					// Advance to next value
					if gen.Next() {
						nextValue = gen.Current()
						nextKey = gen.Key()
					} else {
						nextValue = values.NewNull()
						nextKey = values.NewNull()
					}
				}
			}
		}
	} else if iterator != nil && iterator.index < len(iterator.values) {
		// Handle array iteration
		nextValue = copyValue(iterator.values[iterator.index])
		if iterator.index < len(iterator.keys) {
			nextKey = copyValue(iterator.keys[iterator.index])
		}
		iterator.index++
	} else {
		// Exhausted iterator
		nextValue = values.NewNull()
		nextKey = values.NewNull()
	}

	if opType2 != opcodes.IS_UNUSED {
		if err := vm.writeOperand(ctx, frame, opType2, op2, nextKey); err != nil {
			return false, err
		}
	}

	if resType != opcodes.IS_UNUSED {
		if err := vm.writeOperand(ctx, frame, resType, resSlot, nextValue); err != nil {
			return false, err
		}
	}

	return true, nil
}

func (vm *VirtualMachine) execFeFree(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	opType1, op1 := decodeOperand(inst, 1)
	_ = opType1
	if frame.Iterators != nil {
		delete(frame.Iterators, op1)
	}
	return true, nil
}

func (vm *VirtualMachine) execInitArray(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	resType, resSlot := decodeResult(inst)
	arr := values.NewArray()
	vm.profile.recordAlloc(1)
	if err := vm.writeOperand(ctx, frame, resType, resSlot, arr); err != nil {
		return false, err
	}
	return true, nil
}

func (vm *VirtualMachine) execAddArrayElement(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	resType, resSlot := decodeResult(inst)
	arrVal, err := vm.writableOperand(frame, resType, resSlot)
	if err != nil {
		return false, err
	}
	arrVal = ensureArrayValue(arrVal)
	arr := arrVal.Data.(*values.Array)

	opType1, op1 := decodeOperand(inst, 1)
	var key *values.Value
	if opType1 != opcodes.IS_UNUSED {
		key, err = vm.readOperand(ctx, frame, opType1, op1)
		if err != nil {
			return false, err
		}
	}
	opType2, op2 := decodeOperand(inst, 2)
	val, err := vm.readOperand(ctx, frame, opType2, op2)
	if err != nil {
		return false, err
	}

	if key == nil || key.IsNull() || opType1 == opcodes.IS_UNUSED {
		arr.Elements[arr.NextIndex] = copyValue(val)
		arr.NextIndex++
	} else {
		arr.Elements[keyInterface(key)] = copyValue(val)
	}

	return true, nil
}

func (vm *VirtualMachine) execArrayUnpack(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	resType, resSlot := decodeResult(inst)
	targetArr, err := vm.readOperand(ctx, frame, resType, resSlot)
	if err != nil {
		return false, err
	}
	if !targetArr.IsArray() {
		return false, errors.New("target is not an array")
	}
	srcType, srcOp := decodeOperand(inst, 1)
	srcVal, err := vm.readOperand(ctx, frame, srcType, srcOp)
	if err != nil {
		return false, err
	}
	if !srcVal.IsArray() {
		return true, nil
	}
	dest := targetArr.Data.(*values.Array)
	src := srcVal.Data.(*values.Array)
	for key, val := range src.Elements {
		dest.Elements[key] = copyValue(val)
	}
	if src.NextIndex > dest.NextIndex {
		dest.NextIndex = src.NextIndex
	}
	return true, nil
}

func (vm *VirtualMachine) execFetchDim(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	arrType, arrOp := decodeOperand(inst, 1)
	arrVal, err := vm.readOperand(ctx, frame, arrType, arrOp)
	if err != nil {
		return false, err
	}
	if arrVal == nil || !arrVal.IsArray() {
		resType, resSlot := decodeResult(inst)
		if err := vm.writeOperand(ctx, frame, resType, resSlot, values.NewNull()); err != nil {
			return false, err
		}
		return true, nil
	}

	keyType, keyOp := decodeOperand(inst, 2)
	keyVal, err := vm.readOperand(ctx, frame, keyType, keyOp)
	if err != nil {
		return false, err
	}

	result := arrVal.ArrayGet(keyVal)
	resType, resSlot := decodeResult(inst)
	if err := vm.writeOperand(ctx, frame, resType, resSlot, copyValue(result)); err != nil {
		return false, err
	}
	return true, nil
}

func (vm *VirtualMachine) execFetchDimIs(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	arrType, arrOp := decodeOperand(inst, 1)
	arrVal, err := vm.readOperand(ctx, frame, arrType, arrOp)
	if err != nil {
		return false, err
	}

	keyType, keyOp := decodeOperand(inst, 2)
	keyVal, err := vm.readOperand(ctx, frame, keyType, keyOp)
	if err != nil {
		return false, err
	}

	exists := false
	if arrVal != nil && arrVal.IsArray() {
		key := keyInterface(keyVal)
		_, exists = arrVal.Data.(*values.Array).Elements[key]
	}

	resType, resSlot := decodeResult(inst)
	if err := vm.writeOperand(ctx, frame, resType, resSlot, values.NewBool(exists)); err != nil {
		return false, err
	}
	return true, nil
}

func (vm *VirtualMachine) execFetchDimUnset(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	arrType, arrOp := decodeOperand(inst, 1)
	arrVal, err := vm.readOperand(ctx, frame, arrType, arrOp)
	if err != nil {
		return false, err
	}

	keyType, keyOp := decodeOperand(inst, 2)
	keyVal, err := vm.readOperand(ctx, frame, keyType, keyOp)
	if err != nil {
		return false, err
	}

	if arrVal != nil && arrVal.IsArray() {
		arrVal.ArrayUnset(keyVal)
	}

	return true, nil
}

func (vm *VirtualMachine) execFetchDimWrite(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction, ensureReadable bool) (bool, error) {
	arrType, arrOp := decodeOperand(inst, 1)
	baseVal, err := vm.writableOperand(frame, arrType, arrOp)
	if err != nil {
		return false, err
	}
	baseVal = ensureArrayValue(baseVal)

	keyType, keyOp := decodeOperand(inst, 2)
	keyVal, err := vm.readOperand(ctx, frame, keyType, keyOp)
	if err != nil {
		return false, err
	}

	elem := ensureArrayElement(baseVal, keyVal)
	if ensureReadable && elem.IsReference() {
		elem = elem.Deref()
	}

	resType, resSlot := decodeResult(inst)
	if err := vm.writeOperand(ctx, frame, resType, resSlot, elem); err != nil {
		return false, err
	}
	return true, nil
}

func (vm *VirtualMachine) execAssignDim(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	arrType, arrOp := decodeOperand(inst, 1)
	baseVal, err := vm.writableOperand(frame, arrType, arrOp)
	if err != nil {
		return false, err
	}
	baseVal = ensureArrayValue(baseVal)

	keyType, keyOp := decodeOperand(inst, 2)
	keyVal, err := vm.readOperand(ctx, frame, keyType, keyOp)
	if err != nil {
		return false, err
	}

	elem := ensureArrayElement(baseVal, keyVal)

	resType, resSlot := decodeResult(inst)
	value, err := vm.readOperand(ctx, frame, resType, resSlot)
	if err != nil {
		return false, err
	}
	assigned := copyValue(value)
	*elem = *assigned

	if err := vm.writeOperand(ctx, frame, resType, resSlot, assigned); err != nil {
		return false, err
	}
	return true, nil
}

func (vm *VirtualMachine) execCatch(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	frame.pushExceptionHandler(&exceptionHandler{
		catchIP:   int(inst.Op1),
		finallyIP: int(inst.Op2),
	})
	return true, nil
}

func (vm *VirtualMachine) execFinally(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	if handler := frame.peekExceptionHandler(); handler != nil && handler.finallyIP == frame.IP {
		frame.popExceptionHandler()
	}
	return true, nil
}

func (vm *VirtualMachine) execThrow(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	opType1, op1 := decodeOperand(inst, 1)
	val, err := vm.readOperand(ctx, frame, opType1, op1)
	if err != nil {
		return false, err
	}
	return vm.raiseException(ctx, frame, copyValue(val))
}

func (vm *VirtualMachine) execNew(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	opType1, op1 := decodeOperand(inst, 1)
	classVal, err := vm.readOperand(ctx, frame, opType1, op1)
	if err != nil {
		return false, err
	}
	className, err := resolveRuntimeClassName(ctx, frame, classVal.ToString())
	if err != nil {
		return false, err
	}
	obj, err := instantiateObject(ctx, className)
	if err != nil {
		return false, err
	}
	resType, resSlot := decodeResult(inst)
	if err := vm.writeOperand(ctx, frame, resType, resSlot, obj); err != nil {
		return false, err
	}
	return true, nil
}

func (vm *VirtualMachine) execUnsetVar(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	opType1, op1 := decodeOperand(inst, 1)
	switch opType1 {
	case opcodes.IS_VAR, opcodes.IS_CV:
		name, _ := frame.SlotNames[op1]

		// Get the value before unsetting to check if it's an object
		val, exists := frame.getLocalWithStatus(op1)
		if exists && val != nil && val.IsObject() {
			// Call destructor on the object before unsetting
			if err := vm.callDestructor(ctx, val); err != nil {
				return false, fmt.Errorf("error calling destructor: %v", err)
			}
		}

		frame.unsetLocal(op1)
		if globalName, ok := frame.globalSlotName(op1); ok {
			ctx.unsetGlobal(globalName)
			frame.unbindGlobalSlot(op1)
		}
		if name != "" {
			ctx.unsetVariable(name)
			if _, watched := vm.watchVars[name]; watched {
				vm.recordDebug(ctx, fmt.Sprintf("unset %s", name))
			}
		}
		ctx.recordAssignment(frame, op1, values.NewNull())
		return true, nil
	default:
		return false, fmt.Errorf("unset requires variable operand, got %d", opType1)
	}
}

func (vm *VirtualMachine) execIssetIsEmptyVar(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	opType1, op1 := decodeOperand(inst, 1)

	var result bool
	switch opType1 {
	case opcodes.IS_VAR, opcodes.IS_CV:
		val, exists := frame.getLocalWithStatus(op1)
		if !exists || val == nil {
			result = false
			break
		}
		deref := val.Deref()
		result = !deref.IsNull()
	case opcodes.IS_TMP_VAR, opcodes.IS_CONST:
		fallthrough
	default:
		val, err := vm.readOperand(ctx, frame, opType1, op1)
		if err != nil {
			return false, err
		}
		if val == nil {
			result = true
			break
		}
		deref := val.Deref()
		result = !deref.ToBool()
	}

	resType, resSlot := decodeResult(inst)
	if err := vm.writeOperand(ctx, frame, resType, resSlot, values.NewBool(result)); err != nil {
		return false, err
	}
	return true, nil
}

func (vm *VirtualMachine) execEcho(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	opType1, op1 := decodeOperand(inst, 1)
	val, err := vm.readOperand(ctx, frame, opType1, op1)
	if err != nil {
		return false, err
	}
	if ctx.OutputWriter == nil {
		return false, fmt.Errorf("no output writer configured")
	}
	_, err = fmt.Fprint(ctx.OutputWriter, val.ToString())
	return true, err
}

func (vm *VirtualMachine) execInitFCall(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	opType1, op1 := decodeOperand(inst, 1)
	callee, err := vm.readOperand(ctx, frame, opType1, op1)
	if err != nil {
		return false, err
	}

	pending := &PendingCall{
		Callee: callee,
		Args:   make([]*values.Value, 0),
	}

	if callee.IsCallable() {
		closure := callee.ClosureGet()
		if closure != nil {
			if fn, ok := closure.Function.(*registry.Function); ok {
				pending.Function = fn
				pending.ClosureName = fn.Name
			}
		}
	} else {
		name := callee.ToString()
		pending.ClosureName = name
		if fn, ok := ctx.UserFunctions[strings.ToLower(name)]; ok {
			pending.Function = fn
		} else if registry.GlobalRegistry != nil {
			if fn, ok := registry.GlobalRegistry.GetFunction(name); ok {
				pending.Function = fn
			}
		}
		if pending.Function == nil {
			// Check if this is an object with __invoke method
			if callee.IsObject() {
				obj := callee.Data.(*values.Object)
				class := ctx.ensureClass(obj.ClassName)
				if class != nil && class.Descriptor != nil {
					if invokeMethod, hasInvoke := class.Descriptor.Methods["__invoke"]; hasInvoke {
						// Convert to method call to __invoke
						pending.Function = invokeMethod
						pending.ClosureName = "__invoke"
						pending.Method = true
						pending.This = callee
						pending.ClassName = obj.ClassName
						pending.MethodName = "__invoke"
					} else {
						return false, fmt.Errorf("function name must be a string")
					}
				} else {
					return false, fmt.Errorf("undefined function %s", name)
				}
			} else {
				return false, fmt.Errorf("undefined function %s", name)
			}
		}
	}

	frame.pushPendingCall(pending)
	return true, nil
}

func (vm *VirtualMachine) execInitMethodCall(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	opType1, op1 := decodeOperand(inst, 1)
	objectVal, err := vm.readOperand(ctx, frame, opType1, op1)
	if err != nil {
		return false, err
	}
	if objectVal == nil || !objectVal.IsObject() {
		return false, fmt.Errorf("method call on non-object")
	}
	opType2, op2 := decodeOperand(inst, 2)
	methodVal, err := vm.readOperand(ctx, frame, opType2, op2)
	if err != nil {
		return false, err
	}
	methodRaw := methodVal.ToString()
	className := objectVal.Data.(*values.Object).ClassName
	cls := ctx.ensureClass(className)
	targetFn := resolveClassMethod(ctx, cls, methodRaw)
	if targetFn == nil {
		// Check if __call magic method exists
		callMethod := resolveClassMethod(ctx, cls, "__call")
		if callMethod != nil {
			// Use __call as the target, passing method name and arguments
			pending := &PendingCall{
				Callee:        objectVal,
				Function:      callMethod,
				ClosureName:   "__call",
				Args:          []*values.Value{values.NewString(methodRaw)}, // First arg is method name
				Method:        true,
				This:          objectVal,
				ClassName:     cls.Name,
				MethodName:    "__call",
				IsMagicMethod: true, // Mark as magic method for special argument handling
			}
			frame.pushPendingCall(pending)
			return true, nil
		}
		return false, fmt.Errorf("undefined method %s::%s", className, methodRaw)
	}
	pending := &PendingCall{
		Callee:      objectVal,
		Function:    targetFn,
		ClosureName: methodRaw,
		Args:        make([]*values.Value, 0),
		Method:      true,
		This:        objectVal,
		ClassName:   cls.Name,
		MethodName:  methodRaw,
	}
	frame.pushPendingCall(pending)
	return true, nil
}

func (vm *VirtualMachine) execInitStaticMethodCall(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	classType, classOp := decodeOperand(inst, 1)
	classVal, err := vm.readOperand(ctx, frame, classType, classOp)
	if err != nil {
		return false, err
	}
	originalClassName := classVal.ToString()
	className, err := resolveRuntimeClassName(ctx, frame, originalClassName)
	if err != nil {
		return false, err
	}
	methodType, methodOp := decodeOperand(inst, 2)
	methodVal, err := vm.readOperand(ctx, frame, methodType, methodOp)
	if err != nil {
		return false, err
	}
	methodNameRaw := methodVal.ToString()
	cls := ctx.ensureClass(className)
	if cls == nil {
		return false, fmt.Errorf("undefined class %s", className)
	}
	targetFn, definingClass := resolveClassMethodWithDefiningClass(ctx, cls, methodNameRaw)
	if targetFn == nil {
		// Check if __callStatic magic method exists
		callStaticMethod, callStaticDefining := resolveClassMethodWithDefiningClass(ctx, cls, "__callStatic")
		if callStaticMethod != nil {
			// Use __callStatic as the target, passing method name and arguments
			pending := &PendingCall{
				Function:      callStaticMethod,
				ClosureName:   "__callStatic",
				Args:          []*values.Value{values.NewString(methodNameRaw)}, // First arg is method name
				Method:        true,
				Static:        true,
				ClassName:     callStaticDefining.Name, // Use the class where __callStatic is defined
				CallingClass:  originalClassName, // Set calling class for late static binding
				MethodName:    "__callStatic",
				IsMagicMethod: true, // Mark as magic method for special argument handling
				This:          frame.This, // Pass current $this for parent:: calls and constructors
			}
			frame.pushPendingCall(pending)
			return true, nil
		}
		return false, fmt.Errorf("undefined method %s::%s", className, methodNameRaw)
	}
	pending := &PendingCall{
		Function:     targetFn,
		ClosureName:  methodNameRaw,
		Args:         make([]*values.Value, 0),
		Method:       true,
		Static:       true,
		ClassName:    definingClass.Name, // Use the class where the method is actually defined
		CallingClass: originalClassName, // Set calling class for late static binding
		MethodName:   methodNameRaw,
		This:         frame.This, // Pass current $this for parent:: calls and constructors
	}
	frame.pushPendingCall(pending)
	return true, nil
}

func (vm *VirtualMachine) execSendArg(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	pending := frame.currentPendingCall()
	if pending == nil {
		return false, errors.New("argument sent without pending call")
	}
	opType2, op2 := decodeOperand(inst, 2)
	val, err := vm.readOperand(ctx, frame, opType2, op2)
	if err != nil {
		return false, err
	}
	pending.Args = append(pending.Args, copyValue(val))
	return true, nil
}

func (vm *VirtualMachine) execDoFCall(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	pending := frame.popPendingCall()
	if pending == nil {
		return false, errors.New("call without INIT_FCALL")
	}

	resultType, resultSlot := decodeResult(inst)
	frame.ReturnTarget = operandTarget{opType: resultType, slot: resultSlot, valid: resultType != opcodes.IS_UNUSED}

	if pending.Function == nil || pending.Function.Builtin != nil && pending.Function.IsBuiltin {
		// Builtin function invocation
		fn := pending.Function
		if fn == nil && pending.Callee != nil && pending.Callee.IsCallable() {
			closure := pending.Callee.ClosureGet()
			if closure != nil {
				if runtimeFn, ok := closure.Function.(*registry.Function); ok {
					fn = runtimeFn
				}
			}
		}
		if fn == nil || fn.Builtin == nil {
			return false, fmt.Errorf("callable %s not resolved", pending.ClosureName)
		}
		ctxBuiltin := &builtinContext{vm: vm, ctx: ctx}
		args := pending.Args
		// For method calls, prepend the 'this' object as the first argument
		if pending.Method && pending.This != nil {
			args = append([]*values.Value{pending.This}, pending.Args...)
		}
		ret, err := fn.Builtin(ctxBuiltin, args)
		if err != nil {
			return false, err
		}

		// For constructors (__construct), ignore the return value and use the 'this' object instead
		if frame.ReturnTarget.valid {
			var valueToStore *values.Value = ret
			if pending.Method && fn.Name == "__construct" && pending.This != nil {
				// Constructor call - use the 'this' object, not the return value
				valueToStore = pending.This
			}
			if err := vm.writeOperand(ctx, frame, resultType, resultSlot, valueToStore); err != nil {
				return false, err
			}
		}
		frame.resetReturnTarget()
		return true, nil
	}

	// Handle generator functions
	if pending.Function.IsGenerator {
		// Create generator
		gen := runtime2.NewGenerator(pending.Function, pending.Args, vm)

		// Create Generator object
		generatorObj := &values.Object{
			ClassName:  "Generator",
			Properties: make(map[string]*values.Value),
		}
		generatorObj.Properties["__channel_generator"] = &values.Value{
			Type: values.TypeResource,
			Data: gen,
		}
		// Store function name for debugging/reflection
		generatorObj.Properties["function"] = values.NewString(pending.Function.Name)

		generator := &values.Value{
			Type: values.TypeObject,
			Data: generatorObj,
		}

		// Return generator object
		if frame.ReturnTarget.valid {
			if err := vm.writeOperand(ctx, frame, resultType, resultSlot, generator); err != nil {
				return false, err
			}
		}
		frame.resetReturnTarget()
		return true, nil
	}

	child := newCallFrame(pending.Function.Name, pending.Function, pending.Function.Instructions, pending.Function.Constants)
	child.ClassName = pending.ClassName
	child.CallingClass = pending.CallingClass

	// Handle magic method argument restructuring
	if pending.IsMagicMethod && len(pending.Args) > 1 {
		// For magic methods, convert: [methodName, arg1, arg2, ...]
		// to: [methodName, [arg1, arg2, ...]]
		methodName := pending.Args[0]
		actualArgs := pending.Args[1:] // All arguments except the method name

		// Create an array containing all the actual arguments
		argsArray := values.NewArray()
		for i, arg := range actualArgs {
			argsArray.ArraySet(values.NewInt(int64(i)), arg)
		}

		// Replace the arguments with: [methodName, argsArray]
		pending.Args = []*values.Value{methodName, argsArray}
	}

	// Handle both $this and parameters in the same slot range
	// This is a compromise solution for the compiler's slot expectations

	// Bind parameters first (starting from slot 0)
	for i, param := range pending.Function.Parameters {
		var arg *values.Value

		// Handle variadic parameter (last parameter when function is variadic)
		if pending.Function.IsVariadic && i == len(pending.Function.Parameters)-1 {
			// This is the variadic parameter - collect all remaining arguments into an array
			variadicArray := values.NewArray()
			startIndex := i // Start collecting from current parameter index
			for argIndex := startIndex; argIndex < len(pending.Args); argIndex++ {
				variadicArray.ArraySet(values.NewInt(int64(argIndex-startIndex)), copyValue(pending.Args[argIndex]))
			}
			arg = variadicArray
		} else if i < len(pending.Args) {
			arg = copyValue(pending.Args[i])
		} else if param.HasDefault {
			arg = copyValue(param.DefaultValue)
		} else {
			arg = values.NewNull()
		}
		child.setLocal(uint32(i), arg)
		child.bindSlotName(uint32(i), param.Name)
	}

	if pending.Method {
		// Put $this in the next available slot after parameters
		thisSlot := uint32(len(pending.Function.Parameters))
		child.bindSlotName(thisSlot, "this")
		if pending.This != nil {
			child.setLocal(thisSlot, pending.This)
			child.This = pending.This
		} else {
			child.setLocal(thisSlot, values.NewNull())
		}
	}

	ctx.pushFrame(child)
	return false, nil
}

func (vm *VirtualMachine) execReturn(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	opType1, op1 := decodeOperand(inst, 1)
	val, err := vm.readOperand(ctx, frame, opType1, op1)
	if err != nil {
		return false, err
	}
	if err := vm.handleReturn(ctx, copyValue(val)); err != nil {
		return false, err
	}
	return false, nil
}

func (vm *VirtualMachine) execCreateClosure(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	opType1, op1 := decodeOperand(inst, 1)
	if opType1 != opcodes.IS_CONST {
		return false, fmt.Errorf("closure creation requires const operand")
	}
	if int(op1) >= len(frame.Constants) {
		return false, fmt.Errorf("constant index %d out of range", op1)
	}
	name := frame.Constants[op1].ToString()

	var targetFn *registry.Function
	if fn, ok := ctx.UserFunctions[strings.ToLower(name)]; ok {
		targetFn = fn
	} else if registry.GlobalRegistry != nil {
		if fn, ok := registry.GlobalRegistry.GetFunction(name); ok {
			targetFn = fn
		}
	}

	if targetFn == nil {
		return false, fmt.Errorf("unknown function %s for closure", name)
	}

	closure := values.NewClosure(targetFn, nil, name)
	resType, resSlot := decodeResult(inst)
	if err := vm.writeOperand(ctx, frame, resType, resSlot, closure); err != nil {
		return false, err
	}
	return true, nil
}

func (vm *VirtualMachine) execBindUseVar(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	// For now closures capture values eagerly during creation, so this is a no-op.
	return true, nil
}

func (vm *VirtualMachine) execInclude(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	if vm.CompilerCallback == nil {
		return false, errors.New("include executed without compiler callback")
	}

	opType1, op1 := decodeOperand(inst, 1)
	pathVal, err := vm.readOperand(ctx, frame, opType1, op1)
	if err != nil {
		return false, err
	}
	path := filepath.Clean(pathVal.ToString())
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}

	once := inst.Opcode == opcodes.OP_INCLUDE_ONCE || inst.Opcode == opcodes.OP_REQUIRE_ONCE
	if once {
		if ctx.IncludedFiles[path] {
			resType, resSlot := decodeResult(inst)
			if err := vm.writeOperand(ctx, frame, resType, resSlot, values.NewInt(1)); err != nil {
				return false, err
			}
			return true, nil
		}
	}

	source, err := os.ReadFile(path)
	if err != nil {
		if inst.Opcode == opcodes.OP_REQUIRE || inst.Opcode == opcodes.OP_REQUIRE_ONCE {
			errMsg := err.Error()
			if strings.Contains(errMsg, "no such file or directory") {
				errMsg = strings.Replace(errMsg, "no such file or directory", "No such file or directory", 1)
			}
			return false, fmt.Errorf("require: failed to read %s: %s", path, errMsg)
		}
		// include returns false on failure; mimic by returning false value
		resType, resSlot := decodeResult(inst)
		if err := vm.writeOperand(ctx, frame, resType, resSlot, values.NewBool(false)); err != nil {
			return false, err
		}
		return true, nil
	}

	lex := lexer.New(string(source))
	prs := parser.New(lex)
	program := prs.ParseProgram()
	if program == nil {
		return false, fmt.Errorf("failed to parse included file %s", path)
	}

	resultVal, err := vm.CompilerCallback(ctx, program, path, inst.Opcode == opcodes.OP_REQUIRE || inst.Opcode == opcodes.OP_REQUIRE_ONCE)
	if err != nil {
		return false, err
	}
	if resultVal == nil {
		resultVal = values.NewInt(1)
	}

	if once {
		ctx.IncludedFiles[path] = true
	}

	resType, resSlot := decodeResult(inst)
	if err := vm.writeOperand(ctx, frame, resType, resSlot, resultVal); err != nil {
		return false, err
	}

	return true, nil
}

func keyInterface(val *values.Value) interface{} {
	if val == nil {
		return nil
	}
	switch val.Type {
	case values.TypeInt:
		return val.ToInt()
	case values.TypeString:
		return val.ToString()
	default:
		return val.ToString()
	}
}

func (vm *VirtualMachine) execDeclareInterface(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	opType1, op1 := decodeOperand(inst, 1)
	if opType1 != opcodes.IS_CONST {
		return false, fmt.Errorf("DECLARE_INTERFACE requires interface name constant")
	}
	if int(op1) >= len(frame.Constants) {
		return false, fmt.Errorf("interface name constant %d out of range", op1)
	}
	_ = frame.Constants[op1].ToString() // interfaceName for future use

	// Store interface declaration in context (interfaces are compile-time constructs)
	// In PHP, interfaces are primarily used for type checking and contracts
	// They don't need runtime state like classes do

	return true, nil
}

func (vm *VirtualMachine) execAddInterface(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	// Get class name (first operand)
	classType, classOp := decodeOperand(inst, 1)
	interfaceType, interfaceOp := decodeOperand(inst, 2)

	if classType != opcodes.IS_CONST || interfaceType != opcodes.IS_CONST {
		return false, fmt.Errorf("ADD_INTERFACE expects constant operands")
	}

	if int(classOp) >= len(frame.Constants) || int(interfaceOp) >= len(frame.Constants) {
		return false, fmt.Errorf("ADD_INTERFACE constant operand out of range")
	}

	className := frame.Constants[classOp].ToString()
	interfaceName := frame.Constants[interfaceOp].ToString()

	// Ensure the class exists
	class := ctx.ensureClass(className)
	if class == nil {
		return false, fmt.Errorf("class %s not found for interface implementation", className)
	}

	// Add interface to class (for future interface checking)
	if class.Descriptor != nil && class.Descriptor.Interfaces == nil {
		class.Descriptor.Interfaces = make([]string, 0)
	}
	if class.Descriptor != nil {
		class.Descriptor.Interfaces = append(class.Descriptor.Interfaces, interfaceName)
	}

	return true, nil
}

func (vm *VirtualMachine) execDeclareTrait(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	opType1, op1 := decodeOperand(inst, 1)
	if opType1 != opcodes.IS_CONST {
		return false, fmt.Errorf("DECLARE_TRAIT requires trait name constant")
	}
	if int(op1) >= len(frame.Constants) {
		return false, fmt.Errorf("trait name constant %d out of range", op1)
	}
	_ = frame.Constants[op1].ToString() // traitName for future use

	// Store trait declaration in context (traits are compile-time constructs)
	// In PHP, traits provide a method of code reuse similar to mixins
	// They don't need runtime state like classes do

	return true, nil
}

func (vm *VirtualMachine) execUseTrait(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	// Get trait name (first operand)
	traitType, traitOp := decodeOperand(inst, 1)

	if traitType != opcodes.IS_CONST {
		return false, fmt.Errorf("USE_TRAIT expects trait name constant")
	}

	if int(traitOp) >= len(frame.Constants) {
		return false, fmt.Errorf("USE_TRAIT trait name constant out of range")
	}

	_ = frame.Constants[traitOp].ToString() // traitName for future use

	// This opcode is executed within class declaration context
	// The trait methods and properties have already been copied by the compiler
	// We just need to record the trait usage for runtime reflection if needed

	// For now, this is mainly for tracking purposes
	// The actual trait method and property copying is done at compile time

	return true, nil
}

func (vm *VirtualMachine) execClone(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	// Get the object to clone
	opType1, op1 := decodeOperand(inst, 1)
	objectVal, err := vm.readOperand(ctx, frame, opType1, op1)
	if err != nil {
		return false, err
	}

	var clonedObj *values.Value

	if objectVal == nil || !objectVal.IsObject() {
		return false, fmt.Errorf("__clone method called on non-object")
	}

	// Get the original object
	originalObj := objectVal.Data.(*values.Object)

	// Create a shallow copy of the object
	clonedObj = values.NewObject(originalObj.ClassName)
	clonedObjData := clonedObj.Data.(*values.Object)

	// Copy properties from original to clone
	for name, prop := range originalObj.Properties {
		// Shallow copy - properties reference the same values
		clonedObjData.Properties[name] = prop
	}

	// Check if the class has a __clone magic method and call it
	class := ctx.ensureClass(originalObj.ClassName)
	if class != nil && class.Descriptor != nil {
		if cloneMethod, hasClone := class.Descriptor.Methods["__clone"]; hasClone {
			// Prepare to call __clone method on the cloned object
			child := newCallFrame("__clone", cloneMethod, cloneMethod.Instructions, cloneMethod.Constants)
			child.ClassName = originalObj.ClassName

			// __clone has no parameters, so $this goes in slot 0
			child.bindSlotName(0, "$this")
			child.setLocal(0, clonedObj)
			child.This = clonedObj

			// Push frame and execute
			ctx.pushFrame(child)

			// Execute the __clone method instructions
			for child.IP < len(child.Instructions) {
				instruction := child.Instructions[child.IP]
				child.IP++

				// Special handling for RETURN in __clone
				if instruction.Opcode == opcodes.OP_RETURN {
					// Pop the frame and continue with clone operation
					ctx.popFrame()
					break
				}

				// Execute the instruction
				_, err := vm.executeInstruction(ctx, child, instruction)
				if err != nil {
					ctx.popFrame()
					return false, fmt.Errorf("error in __clone: %v", err)
				}
			}

			// Ensure frame is popped if we reached the end without a return
			if ctx.currentFrame() == child {
				ctx.popFrame()
			}
		}
	}

	// Store the result
	resultType, resultSlot := decodeResult(inst)
	if err := vm.writeOperand(ctx, frame, resultType, resultSlot, clonedObj); err != nil {
		return false, err
	}

	return true, nil
}

// callDestructor calls the __destruct method on an object if it exists
func (vm *VirtualMachine) callDestructor(ctx *ExecutionContext, objectVal *values.Value) error {
	if objectVal == nil || !objectVal.IsObject() {
		return nil // Not an object, nothing to destruct
	}

	obj := objectVal.Data.(*values.Object)
	// Check if already destructed to prevent multiple calls
	if obj.Destructed {
		return nil
	}

	class := ctx.ensureClass(obj.ClassName)
	if class == nil || class.Descriptor == nil {
		return nil // No class descriptor
	}

	// Mark object as destructed immediately to prevent multiple calls
	obj.Destructed = true

	// Check if the class has a __destruct magic method
	if destructMethod, hasDestruct := class.Descriptor.Methods["__destruct"]; hasDestruct {
		// Create a new frame for the __destruct method
		child := newCallFrame("__destruct", destructMethod, destructMethod.Instructions, destructMethod.Constants)
		child.ClassName = obj.ClassName

		// __destruct has no parameters, so $this goes in slot 0
		child.bindSlotName(0, "$this")
		child.setLocal(0, objectVal)
		child.This = objectVal

		// Push frame and execute
		ctx.pushFrame(child)

		// Execute the __destruct method instructions
		for child.IP < len(child.Instructions) {
			instruction := child.Instructions[child.IP]
			child.IP++

			// Special handling for RETURN in __destruct
			if instruction.Opcode == opcodes.OP_RETURN {
				// Pop the frame and continue
				ctx.popFrame()
				return nil
			}

			// Execute the instruction
			_, err := vm.executeInstruction(ctx, child, instruction)
			if err != nil {
				ctx.popFrame()
				return fmt.Errorf("error in __destruct: %v", err)
			}
		}

		// Ensure frame is popped if we reached the end without a return
		if ctx.currentFrame() == child {
			ctx.popFrame()
		}
	}

	return nil
}

// CallAllDestructors calls destructors on all objects in the execution context
func (vm *VirtualMachine) execCast(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (bool, error) {
	// Get the value to cast
	valueType, valueOp := decodeOperand(inst, 1)
	value, err := vm.readOperand(ctx, frame, valueType, valueOp)
	if err != nil {
		return false, err
	}

	// Get the result target
	resultType, resultOp := decodeResult(inst)

	var result *values.Value

	// Cast based on the opcode
	switch inst.Opcode {
	case opcodes.OP_CAST_BOOL:
		result = values.NewBool(value.ToBool())
	case opcodes.OP_CAST_LONG:
		result = values.NewInt(value.ToInt())
	case opcodes.OP_CAST_DOUBLE:
		result = values.NewFloat(value.ToFloat())
	case opcodes.OP_CAST_STRING:
		result = values.NewString(value.ToString())
	case opcodes.OP_CAST_ARRAY:
		if value.Type == values.TypeArray {
			result = value
		} else {
			// Convert to single-element array
			arr := values.NewArray()
			arr.ArraySet(values.NewNull(), value)
			result = arr
		}
	case opcodes.OP_CAST_OBJECT:
		if value.Type == values.TypeObject {
			result = value
		} else {
			// Create stdClass object with single property
			obj := values.NewObject("stdClass")
			obj.Data.(*values.Object).Properties["scalar"] = value
			result = obj
		}
	default:
		return false, fmt.Errorf("unsupported cast opcode: %s", inst.Opcode)
	}

	// Store the result
	err = vm.writeOperand(ctx, frame, resultType, resultOp, result)
	return err == nil, err
}

// This should be called at script end to clean up remaining objects
func (vm *VirtualMachine) CallAllDestructors(ctx *ExecutionContext) {
	// Call destructors on objects in all frames
	for frameIndex := len(ctx.CallStack) - 1; frameIndex >= 0; frameIndex-- {
		frame := ctx.CallStack[frameIndex]
		for slot := uint32(0); slot < uint32(len(frame.Locals)); slot++ {
			if val, exists := frame.getLocalWithStatus(slot); exists && val != nil && val.IsObject() {
				// Call destructor (ignore errors at script end)
				vm.callDestructor(ctx, val)
			}
		}
	}

	// Call destructors on global variables
	for _, val := range ctx.GlobalVars {
		if val != nil && val.IsObject() {
			// Call destructor (ignore errors at script end)
			vm.callDestructor(ctx, val)
		}
	}

	// Call destructors on other variables in the context
	for _, val := range ctx.Variables {
		if val != nil && val.IsObject() {
			// Call destructor (ignore errors at script end)
			vm.callDestructor(ctx, val)
		}
	}
}
