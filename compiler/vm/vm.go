package vm

import (
	"fmt"
	"runtime"

	"github.com/wudi/php-parser/compiler/opcodes"
	"github.com/wudi/php-parser/compiler/values"
)

// ExecutionContext represents the runtime execution state
type ExecutionContext struct {
	// Bytecode execution
	Instructions []opcodes.Instruction
	IP           int // Instruction pointer
	
	// Runtime stacks
	Stack         []*values.Value
	SP            int // Stack pointer
	MaxStackSize  int
	
	// Variable storage
	Variables     map[uint32]*values.Value // Variable slots
	Constants     []*values.Value          // Constant pool
	Temporaries   map[uint32]*values.Value // Temporary variables
	
	// Function call stack
	CallStack     []CallFrame
	
	// Global state
	GlobalVars    map[string]*values.Value
	Functions     map[string]*Function
	Classes       map[string]*Class
	
	// Error handling
	ExceptionStack []Exception
	
	// Execution control
	Halted        bool
	ExitCode      int
}

// CallFrame represents a function call frame
type CallFrame struct {
	Function      *Function
	ReturnIP      int
	Variables     map[uint32]*values.Value
	ThisObject    *values.Value
	Arguments     []*values.Value
}

// Function represents a compiled PHP function
type Function struct {
	Name          string
	Instructions  []opcodes.Instruction
	Constants     []*values.Value
	Parameters    []Parameter
	IsVariadic    bool
	IsGenerator   bool
}

// Parameter represents a function parameter
type Parameter struct {
	Name          string
	Type          string
	IsReference   bool
	HasDefault    bool
	DefaultValue  *values.Value
}

// Class represents a compiled PHP class
type Class struct {
	Name          string
	ParentClass   string
	Properties    map[string]*Property
	Methods       map[string]*Function
	Constants     map[string]*values.Value
	IsAbstract    bool
	IsFinal       bool
}

// Property represents a class property
type Property struct {
	Name          string
	Type          string
	Visibility    string // public, private, protected
	IsStatic      bool
	DefaultValue  *values.Value
}

// Exception represents a runtime exception
type Exception struct {
	Value         *values.Value
	File          string
	Line          int
	Trace         []string
}

// VirtualMachine is the PHP bytecode virtual machine
type VirtualMachine struct {
	StackSize     int
	MemoryLimit   int64
	TimeLimit     int
	DebugMode     bool
}

// NewVirtualMachine creates a new VM instance
func NewVirtualMachine() *VirtualMachine {
	return &VirtualMachine{
		StackSize:   10000,
		MemoryLimit: 128 * 1024 * 1024, // 128MB
		TimeLimit:   30,                 // 30 seconds
		DebugMode:   false,
	}
}

// NewExecutionContext creates a new execution context
func NewExecutionContext() *ExecutionContext {
	return &ExecutionContext{
		Stack:        make([]*values.Value, 1000),
		SP:           -1,
		MaxStackSize: 1000,
		Variables:    make(map[uint32]*values.Value),
		Temporaries:  make(map[uint32]*values.Value),
		CallStack:    make([]CallFrame, 0),
		GlobalVars:   make(map[string]*values.Value),
		Functions:    make(map[string]*Function),
		Classes:      make(map[string]*Class),
		ExceptionStack: make([]Exception, 0),
		Halted:       false,
		ExitCode:     0,
	}
}

// Execute runs bytecode instructions in the given context
func (vm *VirtualMachine) Execute(ctx *ExecutionContext, instructions []opcodes.Instruction, constants []*values.Value) error {
	ctx.Instructions = instructions
	ctx.Constants = constants
	ctx.IP = 0
	
	// Main execution loop with computed goto optimization
	for ctx.IP < len(ctx.Instructions) && !ctx.Halted {
		if vm.DebugMode {
			vm.debugInstruction(ctx)
		}
		
		inst := ctx.Instructions[ctx.IP]
		
		err := vm.executeInstruction(ctx, &inst)
		if err != nil {
			return err
		}
		
		// Prevent infinite loops in debug mode
		if vm.DebugMode && ctx.IP > 1000000 {
			return fmt.Errorf("execution limit exceeded (possible infinite loop)")
		}
	}
	
	return nil
}

// executeInstruction executes a single bytecode instruction
func (vm *VirtualMachine) executeInstruction(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	switch inst.Opcode {
	// Arithmetic operations
	case opcodes.OP_ADD:
		return vm.executeAdd(ctx, inst)
	case opcodes.OP_SUB:
		return vm.executeSub(ctx, inst)
	case opcodes.OP_MUL:
		return vm.executeMul(ctx, inst)
	case opcodes.OP_DIV:
		return vm.executeDiv(ctx, inst)
	case opcodes.OP_MOD:
		return vm.executeMod(ctx, inst)
	case opcodes.OP_POW:
		return vm.executePow(ctx, inst)
		
	// Unary operations
	case opcodes.OP_PLUS:
		return vm.executeUnaryPlus(ctx, inst)
	case opcodes.OP_MINUS:
		return vm.executeUnaryMinus(ctx, inst)
	case opcodes.OP_NOT:
		return vm.executeNot(ctx, inst)
	case opcodes.OP_BW_NOT:
		return vm.executeBitwiseNot(ctx, inst)
		
	// Comparison operations
	case opcodes.OP_IS_EQUAL:
		return vm.executeIsEqual(ctx, inst)
	case opcodes.OP_IS_NOT_EQUAL:
		return vm.executeIsNotEqual(ctx, inst)
	case opcodes.OP_IS_IDENTICAL:
		return vm.executeIsIdentical(ctx, inst)
	case opcodes.OP_IS_NOT_IDENTICAL:
		return vm.executeIsNotIdentical(ctx, inst)
	case opcodes.OP_IS_SMALLER:
		return vm.executeIsSmaller(ctx, inst)
	case opcodes.OP_IS_SMALLER_OR_EQUAL:
		return vm.executeIsSmallerOrEqual(ctx, inst)
	case opcodes.OP_IS_GREATER:
		return vm.executeIsGreater(ctx, inst)
	case opcodes.OP_IS_GREATER_OR_EQUAL:
		return vm.executeIsGreaterOrEqual(ctx, inst)
	case opcodes.OP_SPACESHIP:
		return vm.executeSpaceship(ctx, inst)
		
	// Logical operations
	case opcodes.OP_BOOLEAN_AND:
		return vm.executeBooleanAnd(ctx, inst)
	case opcodes.OP_BOOLEAN_OR:
		return vm.executeBooleanOr(ctx, inst)
	case opcodes.OP_LOGICAL_AND:
		return vm.executeBooleanAnd(ctx, inst)  // Same implementation as boolean AND
	case opcodes.OP_LOGICAL_OR:
		return vm.executeBooleanOr(ctx, inst)   // Same implementation as boolean OR
	case opcodes.OP_LOGICAL_XOR:
		return vm.executeBooleanXor(ctx, inst)  // New function needed
		
	// Bitwise operations
	case opcodes.OP_BW_AND:
		return vm.executeBitwiseAnd(ctx, inst)
	case opcodes.OP_BW_OR:
		return vm.executeBitwiseOr(ctx, inst)
	case opcodes.OP_BW_XOR:
		return vm.executeBitwiseXor(ctx, inst)
	case opcodes.OP_SL:
		return vm.executeShiftLeft(ctx, inst)
	case opcodes.OP_SR:
		return vm.executeShiftRight(ctx, inst)
		
	// Control flow
	case opcodes.OP_JMP:
		return vm.executeJump(ctx, inst)
	case opcodes.OP_JMPZ:
		return vm.executeJumpIfZero(ctx, inst)
	case opcodes.OP_JMPNZ:
		return vm.executeJumpIfNotZero(ctx, inst)
		
	// Variable operations
	case opcodes.OP_ASSIGN:
		return vm.executeAssign(ctx, inst)
	case opcodes.OP_FETCH_R:
		return vm.executeFetchRead(ctx, inst)
	case opcodes.OP_FETCH_W:
		return vm.executeFetchWrite(ctx, inst)
		
	// Array operations
	case opcodes.OP_INIT_ARRAY:
		return vm.executeInitArray(ctx, inst)
	case opcodes.OP_ADD_ARRAY_ELEMENT:
		return vm.executeAddArrayElement(ctx, inst)
	case opcodes.OP_FETCH_DIM_R:
		return vm.executeFetchDimRead(ctx, inst)
	case opcodes.OP_FETCH_DIM_W:
		return vm.executeFetchDimWrite(ctx, inst)
		
	// Object operations
	case opcodes.OP_FETCH_OBJ_R:
		return vm.executeFetchObjRead(ctx, inst)
	case opcodes.OP_FETCH_OBJ_W:
		return vm.executeFetchObjWrite(ctx, inst)
		
	// Function operations
	case opcodes.OP_INIT_FCALL:
		return vm.executeInitFunctionCall(ctx, inst)
	case opcodes.OP_SEND_VAL:
		return vm.executeSendValue(ctx, inst)
	case opcodes.OP_DO_FCALL:
		return vm.executeDoFunctionCall(ctx, inst)
		
	// Special operations
	case opcodes.OP_ECHO:
		return vm.executeEcho(ctx, inst)
	case opcodes.OP_RETURN:
		return vm.executeReturn(ctx, inst)
	case opcodes.OP_EXIT:
		return vm.executeExit(ctx, inst)
	case opcodes.OP_QM_ASSIGN:
		return vm.executeQuickAssign(ctx, inst)
		
	// String operations
	case opcodes.OP_CONCAT:
		return vm.executeConcat(ctx, inst)
		
	// No operation
	case opcodes.OP_NOP:
		ctx.IP++
		return nil
		
	default:
		return fmt.Errorf("unsupported opcode: %s", inst.Opcode.String())
	}
}

// Arithmetic instruction implementations

func (vm *VirtualMachine) executeAdd(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	op1 := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	op2 := vm.getValue(ctx, inst.Op2, opcodes.DecodeOpType2(inst.OpType1))
	
	if op1 == nil || op2 == nil {
		return fmt.Errorf("null operand in ADD operation")
	}
	
	result := op1.Add(op2)
	vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), result)
	
	ctx.IP++
	return nil
}

func (vm *VirtualMachine) executeSub(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	op1 := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	op2 := vm.getValue(ctx, inst.Op2, opcodes.DecodeOpType2(inst.OpType1))
	
	if op1 == nil || op2 == nil {
		return fmt.Errorf("null operand in SUB operation")
	}
	
	result := op1.Subtract(op2)
	vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), result)
	
	ctx.IP++
	return nil
}

func (vm *VirtualMachine) executeMul(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	op1 := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	op2 := vm.getValue(ctx, inst.Op2, opcodes.DecodeOpType2(inst.OpType1))
	
	if op1 == nil || op2 == nil {
		return fmt.Errorf("null operand in MUL operation")
	}
	
	result := op1.Multiply(op2)
	vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), result)
	
	ctx.IP++
	return nil
}

func (vm *VirtualMachine) executeDiv(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	op1 := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	op2 := vm.getValue(ctx, inst.Op2, opcodes.DecodeOpType2(inst.OpType1))
	
	if op1 == nil || op2 == nil {
		return fmt.Errorf("null operand in DIV operation")
	}
	
	// Check for division by zero
	if op2.ToFloat() == 0.0 {
		return fmt.Errorf("division by zero")
	}
	
	result := op1.Divide(op2)
	vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), result)
	
	ctx.IP++
	return nil
}

func (vm *VirtualMachine) executeMod(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	op1 := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	op2 := vm.getValue(ctx, inst.Op2, opcodes.DecodeOpType2(inst.OpType1))
	
	if op1 == nil || op2 == nil {
		return fmt.Errorf("null operand in MOD operation")
	}
	
	result := op1.Modulo(op2)
	vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), result)
	
	ctx.IP++
	return nil
}

func (vm *VirtualMachine) executePow(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	op1 := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	op2 := vm.getValue(ctx, inst.Op2, opcodes.DecodeOpType2(inst.OpType1))
	
	if op1 == nil || op2 == nil {
		return fmt.Errorf("null operand in POW operation")
	}
	
	result := op1.Power(op2)
	vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), result)
	
	ctx.IP++
	return nil
}

// Comparison instruction implementations

func (vm *VirtualMachine) executeIsEqual(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	op1 := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	op2 := vm.getValue(ctx, inst.Op2, opcodes.DecodeOpType2(inst.OpType1))
	
	result := values.NewBool(op1.Equal(op2))
	vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), result)
	
	ctx.IP++
	return nil
}

func (vm *VirtualMachine) executeIsNotEqual(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	op1 := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	op2 := vm.getValue(ctx, inst.Op2, opcodes.DecodeOpType2(inst.OpType1))
	
	result := values.NewBool(!op1.Equal(op2))
	vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), result)
	
	ctx.IP++
	return nil
}

func (vm *VirtualMachine) executeIsIdentical(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	op1 := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	op2 := vm.getValue(ctx, inst.Op2, opcodes.DecodeOpType2(inst.OpType1))
	
	result := values.NewBool(op1.Identical(op2))
	vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), result)
	
	ctx.IP++
	return nil
}

func (vm *VirtualMachine) executeIsNotIdentical(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	op1 := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	op2 := vm.getValue(ctx, inst.Op2, opcodes.DecodeOpType2(inst.OpType1))
	
	result := values.NewBool(!op1.Identical(op2))
	vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), result)
	
	ctx.IP++
	return nil
}

func (vm *VirtualMachine) executeIsSmaller(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	op1 := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	op2 := vm.getValue(ctx, inst.Op2, opcodes.DecodeOpType2(inst.OpType1))
	
	result := values.NewBool(op1.Compare(op2) < 0)
	vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), result)
	
	ctx.IP++
	return nil
}

func (vm *VirtualMachine) executeIsSmallerOrEqual(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	op1 := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	op2 := vm.getValue(ctx, inst.Op2, opcodes.DecodeOpType2(inst.OpType1))
	
	result := values.NewBool(op1.Compare(op2) <= 0)
	vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), result)
	
	ctx.IP++
	return nil
}

func (vm *VirtualMachine) executeIsGreater(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	op1 := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	op2 := vm.getValue(ctx, inst.Op2, opcodes.DecodeOpType2(inst.OpType1))
	
	result := values.NewBool(op1.Compare(op2) > 0)
	vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), result)
	
	ctx.IP++
	return nil
}

func (vm *VirtualMachine) executeIsGreaterOrEqual(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	op1 := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	op2 := vm.getValue(ctx, inst.Op2, opcodes.DecodeOpType2(inst.OpType1))
	
	result := values.NewBool(op1.Compare(op2) >= 0)
	vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), result)
	
	ctx.IP++
	return nil
}

func (vm *VirtualMachine) executeSpaceship(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	op1 := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	op2 := vm.getValue(ctx, inst.Op2, opcodes.DecodeOpType2(inst.OpType1))
	
	result := values.NewInt(int64(op1.Compare(op2)))
	vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), result)
	
	ctx.IP++
	return nil
}

// Control flow instructions

func (vm *VirtualMachine) executeJump(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	target := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	if target == nil || !target.IsInt() {
		return fmt.Errorf("invalid jump target")
	}
	
	ctx.IP = int(target.ToInt())
	return nil
}

func (vm *VirtualMachine) executeJumpIfZero(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	condition := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	
	if !condition.ToBool() {
		target := vm.getValue(ctx, inst.Op2, opcodes.DecodeOpType2(inst.OpType1))
		if target == nil || !target.IsInt() {
			return fmt.Errorf("invalid jump target")
		}
		ctx.IP = int(target.ToInt())
	} else {
		ctx.IP++
	}
	
	return nil
}

func (vm *VirtualMachine) executeJumpIfNotZero(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	condition := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	
	if condition.ToBool() {
		target := vm.getValue(ctx, inst.Op2, opcodes.DecodeOpType2(inst.OpType1))
		if target == nil || !target.IsInt() {
			return fmt.Errorf("invalid jump target")
		}
		ctx.IP = int(target.ToInt())
	} else {
		ctx.IP++
	}
	
	return nil
}

// Variable operations

func (vm *VirtualMachine) executeAssign(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	value := vm.getValue(ctx, inst.Op2, opcodes.DecodeOpType2(inst.OpType1))
	vm.setValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1), value)
	
	// Also set as result
	vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), value)
	
	ctx.IP++
	return nil
}

func (vm *VirtualMachine) executeFetchRead(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	value := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), value)
	
	ctx.IP++
	return nil
}

func (vm *VirtualMachine) executeFetchWrite(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	// For write operations, we typically return a reference
	// This is a simplified implementation
	value := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), value)
	
	ctx.IP++
	return nil
}

// Array operations

func (vm *VirtualMachine) executeInitArray(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	array := values.NewArray()
	vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), array)
	
	ctx.IP++
	return nil
}

func (vm *VirtualMachine) executeAddArrayElement(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	array := vm.getValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2))
	
	if !array.IsArray() {
		return fmt.Errorf("trying to add element to non-array")
	}
	
	var key *values.Value
	if opcodes.DecodeOpType1(inst.OpType1) != opcodes.IS_UNUSED {
		key = vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	}
	
	value := vm.getValue(ctx, inst.Op2, opcodes.DecodeOpType2(inst.OpType1))
	array.ArraySet(key, value)
	
	ctx.IP++
	return nil
}

func (vm *VirtualMachine) executeFetchDimRead(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	array := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	key := vm.getValue(ctx, inst.Op2, opcodes.DecodeOpType2(inst.OpType1))
	
	var result *values.Value
	if array.IsArray() {
		result = array.ArrayGet(key)
	} else {
		result = values.NewNull()
	}
	
	vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), result)
	
	ctx.IP++
	return nil
}

func (vm *VirtualMachine) executeFetchDimWrite(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	// This is a simplified implementation
	return vm.executeFetchDimRead(ctx, inst)
}

// Object operations

func (vm *VirtualMachine) executeFetchObjRead(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	object := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	property := vm.getValue(ctx, inst.Op2, opcodes.DecodeOpType2(inst.OpType1))
	
	var result *values.Value
	if object.IsObject() && property.IsString() {
		result = object.ObjectGet(property.ToString())
	} else {
		result = values.NewNull()
	}
	
	vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), result)
	
	ctx.IP++
	return nil
}

func (vm *VirtualMachine) executeFetchObjWrite(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	// Simplified implementation
	return vm.executeFetchObjRead(ctx, inst)
}

// Special operations

func (vm *VirtualMachine) executeEcho(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	value := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	fmt.Print(value.ToString())
	
	ctx.IP++
	return nil
}

func (vm *VirtualMachine) executeReturn(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	// For now, just halt execution
	// In a full implementation, this would handle function returns properly
	ctx.Halted = true
	return nil
}

func (vm *VirtualMachine) executeExit(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	if opcodes.DecodeOpType1(inst.OpType1) != opcodes.IS_UNUSED {
		exitCode := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
		ctx.ExitCode = int(exitCode.ToInt())
	}
	ctx.Halted = true
	return nil
}

func (vm *VirtualMachine) executeQuickAssign(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	value := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), value)
	
	ctx.IP++
	return nil
}

func (vm *VirtualMachine) executeConcat(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	op1 := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	op2 := vm.getValue(ctx, inst.Op2, opcodes.DecodeOpType2(inst.OpType1))
	
	result := op1.Concat(op2)
	vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), result)
	
	ctx.IP++
	return nil
}

// Unary operations

func (vm *VirtualMachine) executeUnaryPlus(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	operand := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	
	var result *values.Value
	if operand.IsFloat() {
		result = values.NewFloat(operand.ToFloat())
	} else {
		result = values.NewInt(operand.ToInt())
	}
	
	vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), result)
	ctx.IP++
	return nil
}

func (vm *VirtualMachine) executeUnaryMinus(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	operand := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	
	var result *values.Value
	if operand.IsFloat() {
		result = values.NewFloat(-operand.ToFloat())
	} else {
		result = values.NewInt(-operand.ToInt())
	}
	
	vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), result)
	ctx.IP++
	return nil
}

func (vm *VirtualMachine) executeNot(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	operand := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	result := values.NewBool(!operand.ToBool())
	
	vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), result)
	ctx.IP++
	return nil
}

func (vm *VirtualMachine) executeBitwiseNot(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	operand := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	result := values.NewInt(^operand.ToInt())
	
	vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), result)
	ctx.IP++
	return nil
}

// Logical operations

func (vm *VirtualMachine) executeBooleanAnd(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	op1 := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	op2 := vm.getValue(ctx, inst.Op2, opcodes.DecodeOpType2(inst.OpType1))
	
	result := values.NewBool(op1.ToBool() && op2.ToBool())
	vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), result)
	
	ctx.IP++
	return nil
}

func (vm *VirtualMachine) executeBooleanOr(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	op1 := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	op2 := vm.getValue(ctx, inst.Op2, opcodes.DecodeOpType2(inst.OpType1))
	
	result := values.NewBool(op1.ToBool() || op2.ToBool())
	vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), result)
	
	ctx.IP++
	return nil
}

func (vm *VirtualMachine) executeBooleanXor(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	op1 := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	op2 := vm.getValue(ctx, inst.Op2, opcodes.DecodeOpType2(inst.OpType1))
	
	// XOR: true if exactly one operand is true
	bool1 := op1.ToBool()
	bool2 := op2.ToBool()
	result := values.NewBool((bool1 && !bool2) || (!bool1 && bool2))
	vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), result)
	
	ctx.IP++
	return nil
}

// Bitwise operations

func (vm *VirtualMachine) executeBitwiseAnd(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	op1 := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	op2 := vm.getValue(ctx, inst.Op2, opcodes.DecodeOpType2(inst.OpType1))
	
	result := values.NewInt(op1.ToInt() & op2.ToInt())
	vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), result)
	
	ctx.IP++
	return nil
}

func (vm *VirtualMachine) executeBitwiseOr(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	op1 := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	op2 := vm.getValue(ctx, inst.Op2, opcodes.DecodeOpType2(inst.OpType1))
	
	result := values.NewInt(op1.ToInt() | op2.ToInt())
	vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), result)
	
	ctx.IP++
	return nil
}

func (vm *VirtualMachine) executeBitwiseXor(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	op1 := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	op2 := vm.getValue(ctx, inst.Op2, opcodes.DecodeOpType2(inst.OpType1))
	
	result := values.NewInt(op1.ToInt() ^ op2.ToInt())
	vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), result)
	
	ctx.IP++
	return nil
}

func (vm *VirtualMachine) executeShiftLeft(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	op1 := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	op2 := vm.getValue(ctx, inst.Op2, opcodes.DecodeOpType2(inst.OpType1))
	
	result := values.NewInt(op1.ToInt() << uint(op2.ToInt()))
	vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), result)
	
	ctx.IP++
	return nil
}

func (vm *VirtualMachine) executeShiftRight(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	op1 := vm.getValue(ctx, inst.Op1, opcodes.DecodeOpType1(inst.OpType1))
	op2 := vm.getValue(ctx, inst.Op2, opcodes.DecodeOpType2(inst.OpType1))
	
	result := values.NewInt(op1.ToInt() >> uint(op2.ToInt()))
	vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), result)
	
	ctx.IP++
	return nil
}

// Function call operations (simplified)

func (vm *VirtualMachine) executeInitFunctionCall(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	// This is a placeholder for function call initialization
	ctx.IP++
	return nil
}

func (vm *VirtualMachine) executeSendValue(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	// This is a placeholder for sending function arguments
	ctx.IP++
	return nil
}

func (vm *VirtualMachine) executeDoFunctionCall(ctx *ExecutionContext, inst *opcodes.Instruction) error {
	// This is a placeholder for executing function calls
	// In a real implementation, this would look up the function and call it
	result := values.NewNull()
	vm.setValue(ctx, inst.Result, opcodes.DecodeResultType(inst.OpType2), result)
	
	ctx.IP++
	return nil
}

// Helper methods

func (vm *VirtualMachine) getValue(ctx *ExecutionContext, operand uint32, opType opcodes.OpType) *values.Value {
	switch opType {
	case opcodes.IS_CONST:
		if int(operand) < len(ctx.Constants) {
			return ctx.Constants[operand]
		}
		return values.NewNull()
		
	case opcodes.IS_TMP_VAR:
		if val, exists := ctx.Temporaries[operand]; exists {
			return val
		}
		return values.NewNull()
		
	case opcodes.IS_VAR:
		if val, exists := ctx.Variables[operand]; exists {
			return val
		}
		return values.NewNull()
		
	case opcodes.IS_CV:
		// Compiled variables (cached lookups)
		if val, exists := ctx.Variables[operand]; exists {
			return val
		}
		return values.NewNull()
		
	default:
		return values.NewNull()
	}
}

func (vm *VirtualMachine) setValue(ctx *ExecutionContext, operand uint32, opType opcodes.OpType, value *values.Value) {
	switch opType {
	case opcodes.IS_TMP_VAR:
		ctx.Temporaries[operand] = value
		
	case opcodes.IS_VAR, opcodes.IS_CV:
		ctx.Variables[operand] = value
		
	// Constants cannot be set
	case opcodes.IS_CONST:
		// No-op
	}
}

// Debug support

func (vm *VirtualMachine) debugInstruction(ctx *ExecutionContext) {
	if ctx.IP >= len(ctx.Instructions) {
		return
	}
	
	inst := ctx.Instructions[ctx.IP]
	fmt.Printf("[%04d] %s\n", ctx.IP, inst.String())
	
	// Optionally print stack state
	if vm.DebugMode {
		fmt.Printf("       Stack: SP=%d, Size=%d\n", ctx.SP, len(ctx.Stack))
		fmt.Printf("       Temps: %d, Vars: %d\n", len(ctx.Temporaries), len(ctx.Variables))
	}
}

// Memory management

func (vm *VirtualMachine) checkMemoryLimit(ctx *ExecutionContext) error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	if int64(m.Alloc) > vm.MemoryLimit {
		return fmt.Errorf("memory limit exceeded: %d bytes", m.Alloc)
	}
	
	return nil
}