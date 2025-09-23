package vm

import (
	"errors"
	"fmt"

	"github.com/wudi/hey/opcodes"
)

// Pre-defined VM error types for consistent error handling
var (
	// Operand errors
	ErrConstantOutOfRange   = errors.New("constant index out of range")
	ErrInvalidOperandType   = errors.New("invalid operand type")
	ErrOperandNotWritable   = errors.New("operand type not writable")
	ErrUnsupportedOperand   = errors.New("unsupported operand type")

	// Instruction errors
	ErrOpcodeNotImplemented = errors.New("opcode not implemented")
	ErrInvalidInstruction   = errors.New("invalid instruction")
	ErrInstructionFailed    = errors.New("instruction execution failed")

	// Arithmetic errors
	ErrDivisionByZero    = errors.New("division by zero")
	ErrModuloByZero     = errors.New("modulo by zero")
	ErrInvalidArithmetic = errors.New("invalid arithmetic operation")

	// Variable errors
	ErrVariableNotFound   = errors.New("variable not found")
	ErrGlobalNotFound     = errors.New("global variable not found")
	ErrInvalidVariableName = errors.New("invalid variable name")

	// Class errors
	ErrClassNotFound      = errors.New("class not found")
	ErrMethodNotFound     = errors.New("method not found")
	ErrPropertyNotFound   = errors.New("property not found")
	ErrAbstractClass      = errors.New("cannot instantiate abstract class")
	ErrInvalidClassContext = errors.New("invalid class context")

	// Function errors
	ErrFunctionNotFound   = errors.New("function not found")
	ErrInvalidArguments   = errors.New("invalid function arguments")
	ErrCallStackEmpty     = errors.New("call stack is empty")

	// Exception errors
	ErrNoException        = errors.New("no pending exception")
	ErrUncaughtException  = errors.New("uncaught exception")
	ErrExceptionMatch     = errors.New("exception type mismatch")

	// Context errors
	ErrNilContext         = errors.New("nil execution context")
	ErrHaltedExecution    = errors.New("execution halted")
	ErrInvalidState       = errors.New("invalid execution state")
)

// VMError wraps errors with additional context information
type VMError struct {
	Type     error         // The base error type
	Message  string        // Additional context message
	Context  string        // Where the error occurred
	Frame    *CallFrame    // Current call frame (optional)
	Opcode   opcodes.Opcode // Current opcode (optional)
	IP       int           // Instruction pointer (optional)
}

// Error implements the error interface
func (e *VMError) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("vm error in %s: %s: %s", e.Context, e.Type.Error(), e.Message)
	}
	if e.Message != "" {
		return fmt.Sprintf("vm error: %s: %s", e.Type.Error(), e.Message)
	}
	return fmt.Sprintf("vm error: %s", e.Type.Error())
}

// Unwrap allows error unwrapping for errors.Is and errors.As
func (e *VMError) Unwrap() error {
	return e.Type
}

// Is implements error comparison for errors.Is
func (e *VMError) Is(target error) bool {
	return errors.Is(e.Type, target)
}

// ErrorContext provides additional context information
type ErrorContext struct {
	Function string
	File     string
	Line     int
	Frame    *CallFrame
	Opcode   opcodes.Opcode
	IP       int
}

// NewVMError creates a new VM error with context
func NewVMError(baseError error, message string, args ...interface{}) *VMError {
	return &VMError{
		Type:    baseError,
		Message: fmt.Sprintf(message, args...),
	}
}

// NewVMErrorWithContext creates a new VM error with context information
func NewVMErrorWithContext(baseError error, context *ErrorContext, message string, args ...interface{}) *VMError {
	vmErr := &VMError{
		Type:    baseError,
		Message: fmt.Sprintf(message, args...),
	}

	if context != nil {
		if context.Function != "" {
			vmErr.Context = context.Function
		}
		vmErr.Frame = context.Frame
		vmErr.Opcode = context.Opcode
		vmErr.IP = context.IP
	}

	return vmErr
}

// WrapError wraps an existing error with VM error context
func WrapError(err error, baseError error, context string, args ...interface{}) *VMError {
	if err == nil {
		return nil
	}

	// If it's already a VMError, preserve the chain
	if vmErr, ok := err.(*VMError); ok {
		return &VMError{
			Type:    baseError,
			Message: fmt.Sprintf(context, args...),
			Context: vmErr.Context,
			Frame:   vmErr.Frame,
			Opcode:  vmErr.Opcode,
			IP:      vmErr.IP,
		}
	}

	return &VMError{
		Type:    baseError,
		Message: fmt.Sprintf("%s: %v", fmt.Sprintf(context, args...), err),
	}
}

// Error helper functions for common scenarios

// NewConstantError creates an error for constant access issues
func NewConstantError(index uint32, maxIndex int) *VMError {
	return NewVMError(ErrConstantOutOfRange, "index %d, max %d", index, maxIndex)
}

// NewOperandError creates an error for operand type issues
func NewOperandError(opType opcodes.OpType, operation string) *VMError {
	return NewVMError(ErrInvalidOperandType, "%s with operand type %d", operation, opType)
}

// NewArithmeticError creates an error for arithmetic operations
func NewArithmeticError(operation string, operand1, operand2 interface{}) *VMError {
	return NewVMError(ErrInvalidArithmetic, "%s operation with %v and %v", operation, operand1, operand2)
}

// NewDivisionByZeroError creates a division by zero error
func NewDivisionByZeroError() *VMError {
	return NewVMError(ErrDivisionByZero, "")
}

// NewModuloByZeroError creates a modulo by zero error
func NewModuloByZeroError() *VMError {
	return NewVMError(ErrModuloByZero, "")
}

// NewOpcodeError creates an error for unsupported opcodes
func NewOpcodeError(opcode opcodes.Opcode) *VMError {
	return NewVMError(ErrOpcodeNotImplemented, "opcode %s", opcode)
}

// NewClassError creates an error for class-related issues
func NewClassError(className string, operation string) *VMError {
	return NewVMError(ErrClassNotFound, "class %s for %s", className, operation)
}

// NewMethodError creates an error for method-related issues
func NewMethodError(className, methodName string) *VMError {
	return NewVMError(ErrMethodNotFound, "method %s::%s", className, methodName)
}

// NewVariableError creates an error for variable access issues
func NewVariableError(varName string) *VMError {
	return NewVMError(ErrVariableNotFound, "variable %s", varName)
}

// NewFunctionError creates an error for function-related issues
func NewFunctionError(funcName string, reason string) *VMError {
	return NewVMError(ErrFunctionNotFound, "function %s: %s", funcName, reason)
}

// DecorateError adds frame and instruction context to an existing error
func DecorateError(err error, frame *CallFrame, inst *opcodes.Instruction) error {
	if err == nil {
		return nil
	}

	context := &ErrorContext{}
	if frame != nil {
		context.Function = frame.FunctionName
		context.Frame = frame
		context.IP = frame.IP
	}
	if inst != nil {
		context.Opcode = inst.Opcode
	}

	if vmErr, ok := err.(*VMError); ok {
		vmErr.Frame = frame
		vmErr.IP = frame.IP
		vmErr.Opcode = inst.Opcode
		if vmErr.Context == "" && frame != nil {
			vmErr.Context = frame.FunctionName
		}
		return vmErr
	}

	return NewVMErrorWithContext(ErrInstructionFailed, context, "%v", err)
}

// IsVMError checks if an error is a VM error
func IsVMError(err error) bool {
	_, ok := err.(*VMError)
	return ok
}

// GetVMError extracts VM error from an error chain
func GetVMError(err error) *VMError {
	if vmErr, ok := err.(*VMError); ok {
		return vmErr
	}
	return nil
}