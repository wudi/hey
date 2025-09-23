package vm

import (
	"errors"
	"testing"

	"github.com/wudi/hey/opcodes"
)

func TestVMError_Error(t *testing.T) {
	tests := []struct {
		name     string
		vmError  *VMError
		expected string
	}{
		{
			name: "basic error",
			vmError: &VMError{
				Type:    ErrDivisionByZero,
				Message: "",
			},
			expected: "vm error: division by zero",
		},
		{
			name: "error with message",
			vmError: &VMError{
				Type:    ErrConstantOutOfRange,
				Message: "index 5, max 3",
			},
			expected: "vm error: constant index out of range: index 5, max 3",
		},
		{
			name: "error with context",
			vmError: &VMError{
				Type:    ErrVariableNotFound,
				Message: "variable $x",
				Context: "testFunction",
			},
			expected: "vm error in testFunction: variable not found: variable $x",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.vmError.Error(); got != tt.expected {
				t.Errorf("VMError.Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestVMError_Unwrap(t *testing.T) {
	vmErr := &VMError{
		Type:    ErrDivisionByZero,
		Message: "test message",
	}

	if unwrapped := vmErr.Unwrap(); unwrapped != ErrDivisionByZero {
		t.Errorf("VMError.Unwrap() = %v, want %v", unwrapped, ErrDivisionByZero)
	}
}

func TestVMError_Is(t *testing.T) {
	vmErr := &VMError{
		Type:    ErrDivisionByZero,
		Message: "test message",
	}

	if !vmErr.Is(ErrDivisionByZero) {
		t.Errorf("VMError.Is() should return true for ErrDivisionByZero")
	}

	if vmErr.Is(ErrModuloByZero) {
		t.Errorf("VMError.Is() should return false for ErrModuloByZero")
	}

	// Test with errors.Is
	if !errors.Is(vmErr, ErrDivisionByZero) {
		t.Errorf("errors.Is() should return true for ErrDivisionByZero")
	}
}

func TestNewVMError(t *testing.T) {
	err := NewVMError(ErrConstantOutOfRange, "index %d, max %d", 5, 3)

	if err.Type != ErrConstantOutOfRange {
		t.Errorf("NewVMError() Type = %v, want %v", err.Type, ErrConstantOutOfRange)
	}

	expectedMessage := "index 5, max 3"
	if err.Message != expectedMessage {
		t.Errorf("NewVMError() Message = %q, want %q", err.Message, expectedMessage)
	}
}

func TestNewVMErrorWithContext(t *testing.T) {
	frame := &CallFrame{FunctionName: "testFunc", IP: 42}
	context := &ErrorContext{
		Function: "testFunction",
		Frame:    frame,
		Opcode:   opcodes.OP_ADD,
		IP:       42,
	}

	err := NewVMErrorWithContext(ErrInvalidOperandType, context, "operand %d", 1)

	if err.Context != "testFunction" {
		t.Errorf("NewVMErrorWithContext() Context = %q, want %q", err.Context, "testFunction")
	}
	if err.Frame != frame {
		t.Errorf("NewVMErrorWithContext() Frame mismatch")
	}
	if err.Opcode != opcodes.OP_ADD {
		t.Errorf("NewVMErrorWithContext() Opcode = %v, want %v", err.Opcode, opcodes.OP_ADD)
	}
	if err.IP != 42 {
		t.Errorf("NewVMErrorWithContext() IP = %d, want %d", err.IP, 42)
	}
}

func TestWrapError(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := WrapError(originalErr, ErrInstructionFailed, "context %s", "test")

	if wrappedErr.Type != ErrInstructionFailed {
		t.Errorf("WrapError() Type = %v, want %v", wrappedErr.Type, ErrInstructionFailed)
	}

	expectedMessage := "context test: original error"
	if wrappedErr.Message != expectedMessage {
		t.Errorf("WrapError() Message = %q, want %q", wrappedErr.Message, expectedMessage)
	}

	// Test wrapping nil error
	if wrapped := WrapError(nil, ErrInstructionFailed, "test"); wrapped != nil {
		t.Errorf("WrapError() with nil error should return nil")
	}
}

func TestErrorHelperFunctions(t *testing.T) {
	// Test NewConstantError
	constErr := NewConstantError(5, 3)
	if !errors.Is(constErr, ErrConstantOutOfRange) {
		t.Errorf("NewConstantError() should create ErrConstantOutOfRange")
	}

	// Test NewOperandError
	opErr := NewOperandError(opcodes.IS_CONST, "test operation")
	if !errors.Is(opErr, ErrInvalidOperandType) {
		t.Errorf("NewOperandError() should create ErrInvalidOperandType")
	}

	// Test NewDivisionByZeroError
	divErr := NewDivisionByZeroError()
	if !errors.Is(divErr, ErrDivisionByZero) {
		t.Errorf("NewDivisionByZeroError() should create ErrDivisionByZero")
	}

	// Test NewModuloByZeroError
	modErr := NewModuloByZeroError()
	if !errors.Is(modErr, ErrModuloByZero) {
		t.Errorf("NewModuloByZeroError() should create ErrModuloByZero")
	}

	// Test NewOpcodeError
	opcodeErr := NewOpcodeError(opcodes.OP_NOP)
	if !errors.Is(opcodeErr, ErrOpcodeNotImplemented) {
		t.Errorf("NewOpcodeError() should create ErrOpcodeNotImplemented")
	}

	// Test NewVariableError
	varErr := NewVariableError("$x")
	if !errors.Is(varErr, ErrVariableNotFound) {
		t.Errorf("NewVariableError() should create ErrVariableNotFound")
	}

	// Test NewClassError
	classErr := NewClassError("TestClass", "instantiation")
	if !errors.Is(classErr, ErrClassNotFound) {
		t.Errorf("NewClassError() should create ErrClassNotFound")
	}

	// Test NewMethodError
	methodErr := NewMethodError("TestClass", "testMethod")
	if !errors.Is(methodErr, ErrMethodNotFound) {
		t.Errorf("NewMethodError() should create ErrMethodNotFound")
	}

	// Test NewFunctionError
	funcErr := NewFunctionError("testFunc", "not found")
	if !errors.Is(funcErr, ErrFunctionNotFound) {
		t.Errorf("NewFunctionError() should create ErrFunctionNotFound")
	}
}

func TestDecorateError(t *testing.T) {
	originalErr := errors.New("original error")
	frame := &CallFrame{FunctionName: "testFunc", IP: 42}
	inst := &opcodes.Instruction{Opcode: opcodes.OP_ADD}

	decoratedErr := DecorateError(originalErr, frame, inst)

	vmErr, ok := decoratedErr.(*VMError)
	if !ok {
		t.Fatalf("DecorateError() should return VMError")
	}

	if vmErr.Frame != frame {
		t.Errorf("DecorateError() Frame mismatch")
	}
	if vmErr.IP != 42 {
		t.Errorf("DecorateError() IP = %d, want %d", vmErr.IP, 42)
	}
	if vmErr.Opcode != opcodes.OP_ADD {
		t.Errorf("DecorateError() Opcode = %v, want %v", vmErr.Opcode, opcodes.OP_ADD)
	}
	if vmErr.Context != "testFunc" {
		t.Errorf("DecorateError() Context = %q, want %q", vmErr.Context, "testFunc")
	}

	// Test decorating nil error
	if decorated := DecorateError(nil, frame, inst); decorated != nil {
		t.Errorf("DecorateError() with nil error should return nil")
	}

	// Test decorating existing VMError
	existingVMErr := NewVMError(ErrDivisionByZero, "test")
	decoratedVM := DecorateError(existingVMErr, frame, inst)

	if decoratedVMErr, ok := decoratedVM.(*VMError); ok {
		if decoratedVMErr.Frame != frame {
			t.Errorf("DecorateError() on VMError should preserve frame")
		}
	} else {
		t.Errorf("DecorateError() on VMError should return VMError")
	}
}

func TestIsVMError(t *testing.T) {
	vmErr := NewVMError(ErrDivisionByZero, "test")
	regularErr := errors.New("regular error")

	if !IsVMError(vmErr) {
		t.Errorf("IsVMError() should return true for VMError")
	}

	if IsVMError(regularErr) {
		t.Errorf("IsVMError() should return false for regular error")
	}

	if IsVMError(nil) {
		t.Errorf("IsVMError() should return false for nil")
	}
}

func TestGetVMError(t *testing.T) {
	vmErr := NewVMError(ErrDivisionByZero, "test")
	regularErr := errors.New("regular error")

	if retrieved := GetVMError(vmErr); retrieved != vmErr {
		t.Errorf("GetVMError() should return the same VMError")
	}

	if retrieved := GetVMError(regularErr); retrieved != nil {
		t.Errorf("GetVMError() should return nil for regular error")
	}

	if retrieved := GetVMError(nil); retrieved != nil {
		t.Errorf("GetVMError() should return nil for nil error")
	}
}