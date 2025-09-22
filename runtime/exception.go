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
		getWaitGroupClass(),
	}
}

func getWaitGroupClass() *registry.ClassDescriptor {
	methods := map[string]*registry.MethodDescriptor{
		"__construct": {
			Name:       "__construct",
			Visibility: "public",
			Parameters: []*registry.ParameterDescriptor{},
		},
		"Add": {
			Name:       "Add",
			Visibility: "public",
			Parameters: []*registry.ParameterDescriptor{{Name: "delta", Type: "int"}},
		},
		"Done": {
			Name:       "Done",
			Visibility: "public",
			Parameters: []*registry.ParameterDescriptor{},
		},
		"Wait": {
			Name:       "Wait",
			Visibility: "public",
			Parameters: []*registry.ParameterDescriptor{},
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