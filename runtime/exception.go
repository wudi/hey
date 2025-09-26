package runtime

import (
	"fmt"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)


// GetClasses returns all exception-related class definitions
func GetClasses() []*registry.ClassDescriptor {
	return []*registry.ClassDescriptor{
		getExceptionClass(),
		getErrorExceptionClass(),
		getLogicExceptionClass(),
		getRuntimeExceptionClass(),
		getInvalidArgumentExceptionClass(),
		getBadMethodCallExceptionClass(),
		getBadFunctionCallExceptionClass(),
		getDomainExceptionClass(),
		getLengthExceptionClass(),
		getOutOfRangeExceptionClass(),
		getOutOfBoundsExceptionClass(),
		getOverflowExceptionClass(),
		getRangeExceptionClass(),
		getUnderflowExceptionClass(),
		getUnexpectedValueExceptionClass(),
		getJsonExceptionClass(),
		getErrorClass(),
		getTypeErrorClass(),
		getParseErrorClass(),
		getArithmeticErrorClass(),
		getAssertionErrorClass(),
		getDivisionByZeroErrorClass(),
		getValueErrorClass(),
		getArgumentCountErrorClass(),
		getUnhandledMatchErrorClass(),
		getWaitGroupClass(),
	}
}

func getWaitGroupClass() *registry.ClassDescriptor {
	// Create method implementations that properly call WaitGroup value methods
	constructorImpl := &registry.Function{
		Name:      "__construct",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 {
				return values.NewNull(), nil
			}
			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("__construct called on non-object")
			}

			// Initialize the object with a WaitGroup value as internal data
			obj := thisObj.Data.(*values.Object)
			if obj.Properties == nil {
				obj.Properties = make(map[string]*values.Value)
			}
			// Store a WaitGroup value as internal property
			wgVal := values.NewWaitGroup()
			if wgVal == nil {
				return values.NewNull(), fmt.Errorf("failed to create WaitGroup value")
			}
			obj.Properties["__waitgroup_internal"] = wgVal

			return values.NewNull(), nil
		},
	}

	addImpl := &registry.Function{
		Name:      "Add",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("Add() expects 2 arguments (this, delta)")
			}
			thisObj := args[0]
			deltaArg := args[1]

			if !thisObj.IsObject() {
				return nil, fmt.Errorf("Add() called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			wgVal, ok := obj.Properties["__waitgroup_internal"]
			if !ok {
				return nil, fmt.Errorf("WaitGroup not properly initialized")
			}

			delta := deltaArg.ToInt()
			err := wgVal.WaitGroupAdd(delta)
			if err != nil {
				return nil, err
			}

			return values.NewNull(), nil
		},
	}

	doneImpl := &registry.Function{
		Name:      "Done",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("Done() expects 1 argument (this)")
			}
			thisObj := args[0]

			if !thisObj.IsObject() {
				return nil, fmt.Errorf("Done() called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			wgVal, ok := obj.Properties["__waitgroup_internal"]
			if !ok {
				return nil, fmt.Errorf("WaitGroup not properly initialized")
			}

			err := wgVal.WaitGroupDone()
			if err != nil {
				return nil, err
			}

			return values.NewNull(), nil
		},
	}

	waitImpl := &registry.Function{
		Name:      "Wait",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("Wait() expects 1 argument (this)")
			}
			thisObj := args[0]

			if !thisObj.IsObject() {
				return nil, fmt.Errorf("Wait() called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			wgVal, ok := obj.Properties["__waitgroup_internal"]
			if !ok {
				return nil, fmt.Errorf("WaitGroup not properly initialized")
			}

			err := wgVal.WaitGroupWait()
			if err != nil {
				return nil, err
			}

			return values.NewNull(), nil
		},
	}

	methods := map[string]*registry.MethodDescriptor{
		"__construct": {
			Name:           "__construct",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(constructorImpl),
		},
		"Add": {
			Name:       "Add",
			Visibility: "public",
			Parameters: []*registry.ParameterDescriptor{{Name: "delta", Type: "int"}},
			Implementation: NewBuiltinMethodImpl(addImpl),
		},
		"Done": {
			Name:           "Done",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(doneImpl),
		},
		"Wait": {
			Name:           "Wait",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(waitImpl),
		},
	}

	return &registry.ClassDescriptor{
		Name:       "WaitGroup",
		Properties: make(map[string]*registry.PropertyDescriptor),
		Methods:    methods,
		Constants:  make(map[string]*registry.ConstantDescriptor),
	}
}

func getExceptionClass() *registry.ClassDescriptor {
	// Create builtin method implementations for Exception class
	constructorImpl := &registry.Function{
		Name:      "__construct",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			// Constructor is called after object creation, modify the object's properties
			// The 'this' object is passed as first argument in method calls
			if len(args) < 1 {
				return values.NewNull(), nil
			}
			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("__construct called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if obj.Properties == nil {
				obj.Properties = make(map[string]*values.Value)
			}

			// Set message (arg[1] if present)
			message := ""
			if len(args) > 1 && args[1] != nil {
				message = args[1].ToString()
			}
			obj.Properties["message"] = values.NewString(message)

			// Set code (arg[2] if present)
			code := int64(0)
			if len(args) > 2 && args[2] != nil {
				code = args[2].ToInt()
			}
			obj.Properties["code"] = values.NewInt(code)

			// Set file and line (simplified - would need actual source tracking)
			obj.Properties["file"] = values.NewString("")
			obj.Properties["line"] = values.NewInt(0)

			// Constructor should not return a value (void), but return the object for now
			return thisObj, nil
		},
		Parameters: []*registry.Parameter{
			{Name: "message", Type: "string", HasDefault: true, DefaultValue: values.NewString("")},
			{Name: "code", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
			{Name: "previous", Type: "?Throwable", HasDefault: true, DefaultValue: values.NewNull()},
		},
	}

	getMessageImpl := &registry.Function{
		Name:      "getMessage",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewString(""), nil
			}
			obj := args[0].Data.(*values.Object)
			if obj.Properties != nil {
				if msg, ok := obj.Properties["message"]; ok {
					return msg, nil
				}
			}
			return values.NewString(""), nil
		},
		Parameters: []*registry.Parameter{},
	}

	getCodeImpl := &registry.Function{
		Name:      "getCode",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewInt(0), nil
			}
			obj := args[0].Data.(*values.Object)
			if obj.Properties != nil {
				if code, ok := obj.Properties["code"]; ok {
					return code, nil
				}
			}
			return values.NewInt(0), nil
		},
		Parameters: []*registry.Parameter{},
	}

	getFileImpl := &registry.Function{
		Name:      "getFile",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewString(""), nil
			}
			obj := args[0].Data.(*values.Object)
			if obj.Properties != nil {
				if file, ok := obj.Properties["file"]; ok {
					return file, nil
				}
			}
			return values.NewString(""), nil
		},
		Parameters: []*registry.Parameter{},
	}

	getLineImpl := &registry.Function{
		Name:      "getLine",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewInt(0), nil
			}
			obj := args[0].Data.(*values.Object)
			if obj.Properties != nil {
				if line, ok := obj.Properties["line"]; ok {
					return line, nil
				}
			}
			return values.NewInt(0), nil
		},
		Parameters: []*registry.Parameter{},
	}

	getTraceImpl := &registry.Function{
		Name:      "getTrace",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			// Return empty array for now - would need full stack trace implementation
			return values.NewArray(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	getTraceAsStringImpl := &registry.Function{
		Name:      "getTraceAsString",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			// Return empty string for now - would need full stack trace implementation
			return values.NewString(""), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// Create method descriptors that point to the builtin implementations
	methods := map[string]*registry.MethodDescriptor{
		"__construct": {
			Name:       "__construct",
			Visibility: "public",
			Parameters: []*registry.ParameterDescriptor{
				{Name: "message", Type: "string", HasDefault: true, DefaultValue: values.NewString("")},
				{Name: "code", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
				{Name: "previous", Type: "?Throwable", HasDefault: true, DefaultValue: values.NewNull()},
			},
			Implementation: &BuiltinMethodImpl{function: constructorImpl},
		},
		"getMessage": {
			Name:           "getMessage",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: &BuiltinMethodImpl{function: getMessageImpl},
		},
		"getCode": {
			Name:           "getCode",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: &BuiltinMethodImpl{function: getCodeImpl},
		},
		"getFile": {
			Name:           "getFile",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: &BuiltinMethodImpl{function: getFileImpl},
		},
		"getLine": {
			Name:           "getLine",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: &BuiltinMethodImpl{function: getLineImpl},
		},
		"getTrace": {
			Name:           "getTrace",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: &BuiltinMethodImpl{function: getTraceImpl},
		},
		"getTraceAsString": {
			Name:           "getTraceAsString",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: &BuiltinMethodImpl{function: getTraceAsStringImpl},
		},
	}

	// Define class properties
	properties := map[string]*registry.PropertyDescriptor{
		"message": {
			Name:         "message",
			Visibility:   "protected",
			Type:         "string",
			DefaultValue: values.NewString(""),
		},
		"code": {
			Name:         "code",
			Visibility:   "protected",
			Type:         "int",
			DefaultValue: values.NewInt(0),
		},
		"file": {
			Name:         "file",
			Visibility:   "protected",
			Type:         "string",
			DefaultValue: values.NewString(""),
		},
		"line": {
			Name:         "line",
			Visibility:   "protected",
			Type:         "int",
			DefaultValue: values.NewInt(0),
		},
	}

	return &registry.ClassDescriptor{
		Name:       "Exception",
		Properties: properties,
		Methods:    methods,
		Constants:  make(map[string]*registry.ConstantDescriptor),
	}
}

// Helper function to create simple exception classes that inherit from Exception
func createSimpleExceptionClass(name, parent string) *registry.ClassDescriptor {
	return &registry.ClassDescriptor{
		Name:       name,
		Parent:     parent,
		Properties: make(map[string]*registry.PropertyDescriptor),
		Methods:    make(map[string]*registry.MethodDescriptor),
		Constants:  make(map[string]*registry.ConstantDescriptor),
	}
}

// Logic exceptions
func getLogicExceptionClass() *registry.ClassDescriptor {
	return createSimpleExceptionClass("LogicException", "Exception")
}

func getRuntimeExceptionClass() *registry.ClassDescriptor {
	return createSimpleExceptionClass("RuntimeException", "Exception")
}

// LogicException subclasses
func getInvalidArgumentExceptionClass() *registry.ClassDescriptor {
	return createSimpleExceptionClass("InvalidArgumentException", "LogicException")
}

func getBadMethodCallExceptionClass() *registry.ClassDescriptor {
	return createSimpleExceptionClass("BadMethodCallException", "LogicException")
}

func getBadFunctionCallExceptionClass() *registry.ClassDescriptor {
	return createSimpleExceptionClass("BadFunctionCallException", "LogicException")
}

func getDomainExceptionClass() *registry.ClassDescriptor {
	return createSimpleExceptionClass("DomainException", "LogicException")
}

func getOutOfRangeExceptionClass() *registry.ClassDescriptor {
	return createSimpleExceptionClass("OutOfRangeException", "LogicException")
}

// RuntimeException subclasses
func getOutOfBoundsExceptionClass() *registry.ClassDescriptor {
	return createSimpleExceptionClass("OutOfBoundsException", "RuntimeException")
}

func getOverflowExceptionClass() *registry.ClassDescriptor {
	return createSimpleExceptionClass("OverflowException", "RuntimeException")
}

func getRangeExceptionClass() *registry.ClassDescriptor {
	return createSimpleExceptionClass("RangeException", "RuntimeException")
}

func getUnderflowExceptionClass() *registry.ClassDescriptor {
	return createSimpleExceptionClass("UnderflowException", "RuntimeException")
}

func getUnexpectedValueExceptionClass() *registry.ClassDescriptor {
	return createSimpleExceptionClass("UnexpectedValueException", "RuntimeException")
}

// Error classes (PHP 7+)
func getTypeErrorClass() *registry.ClassDescriptor {
	return createSimpleExceptionClass("TypeError", "Error")
}

func getArgumentCountErrorClass() *registry.ClassDescriptor {
	return createSimpleExceptionClass("ArgumentCountError", "TypeError")
}

// ErrorException class (extends Exception)
func getErrorExceptionClass() *registry.ClassDescriptor {
	return createSimpleExceptionClass("ErrorException", "Exception")
}

// Base Error class (PHP 7+)
func getErrorClass() *registry.ClassDescriptor {
	return createSimpleExceptionClass("Error", "Exception")
}

// ParseError class (PHP 7+)
func getParseErrorClass() *registry.ClassDescriptor {
	return createSimpleExceptionClass("ParseError", "Error")
}

// ArithmeticError class (PHP 7+)
func getArithmeticErrorClass() *registry.ClassDescriptor {
	return createSimpleExceptionClass("ArithmeticError", "Error")
}

// DivisionByZeroError class (PHP 7+)
func getDivisionByZeroErrorClass() *registry.ClassDescriptor {
	return createSimpleExceptionClass("DivisionByZeroError", "ArithmeticError")
}

// LengthException class (extends LogicException)
func getLengthExceptionClass() *registry.ClassDescriptor {
	return createSimpleExceptionClass("LengthException", "LogicException")
}

// AssertionError class (PHP 7+)
func getAssertionErrorClass() *registry.ClassDescriptor {
	return createSimpleExceptionClass("AssertionError", "Error")
}

// ValueError class (PHP 8+)
func getValueErrorClass() *registry.ClassDescriptor {
	return createSimpleExceptionClass("ValueError", "Error")
}

// UnhandledMatchError class (PHP 8+)
func getUnhandledMatchErrorClass() *registry.ClassDescriptor {
	return createSimpleExceptionClass("UnhandledMatchError", "Error")
}

// JsonException class (PHP 7.3+)
func getJsonExceptionClass() *registry.ClassDescriptor {
	return createSimpleExceptionClass("JsonException", "Exception")
}