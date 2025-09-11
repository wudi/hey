package stdlib

import (
	"fmt"
	"strings"

	"github.com/wudi/hey/compiler/values"
	"github.com/wudi/hey/compiler/vm"
)

// initClasses initializes built-in PHP classes
func (stdlib *StandardLibrary) initClasses() {
	// Exception class hierarchy
	stdlib.initExceptionClasses()

	// stdClass - the default PHP object
	stdlib.initStdClass()

	// DateTime classes
	stdlib.initDateTimeClasses()

	// Reflection classes
	stdlib.initReflectionClasses()
}

// initExceptionClasses initializes the PHP exception class hierarchy
func (stdlib *StandardLibrary) initExceptionClasses() {
	// Throwable interface (PHP 7+)
	throwable := &Class{
		Name:       "Throwable",
		Parent:     "",
		Properties: make(map[string]*Property),
		Methods:    make(map[string]*Method),
		Constants:  make(map[string]*values.Value),
		IsAbstract: true,
		IsFinal:    false,
	}

	throwable.Methods["getMessage"] = &Method{
		Name:       "getMessage",
		Visibility: "public",
		IsStatic:   false,
		IsAbstract: true,
		IsFinal:    false,
		Parameters: []Parameter{},
		Handler:    nil, // Abstract method
	}

	throwable.Methods["getCode"] = &Method{
		Name:       "getCode",
		Visibility: "public",
		IsStatic:   false,
		IsAbstract: true,
		IsFinal:    false,
		Parameters: []Parameter{},
		Handler:    nil, // Abstract method
	}

	throwable.Methods["getFile"] = &Method{
		Name:       "getFile",
		Visibility: "public",
		IsStatic:   false,
		IsAbstract: true,
		IsFinal:    false,
		Parameters: []Parameter{},
		Handler:    nil, // Abstract method
	}

	throwable.Methods["getLine"] = &Method{
		Name:       "getLine",
		Visibility: "public",
		IsStatic:   false,
		IsAbstract: true,
		IsFinal:    false,
		Parameters: []Parameter{},
		Handler:    nil, // Abstract method
	}

	throwable.Methods["getTrace"] = &Method{
		Name:       "getTrace",
		Visibility: "public",
		IsStatic:   false,
		IsAbstract: true,
		IsFinal:    false,
		Parameters: []Parameter{},
		Handler:    nil, // Abstract method
	}

	throwable.Methods["getTraceAsString"] = &Method{
		Name:       "getTraceAsString",
		Visibility: "public",
		IsStatic:   false,
		IsAbstract: true,
		IsFinal:    false,
		Parameters: []Parameter{},
		Handler:    nil, // Abstract method
	}

	throwable.Methods["getPrevious"] = &Method{
		Name:       "getPrevious",
		Visibility: "public",
		IsStatic:   false,
		IsAbstract: true,
		IsFinal:    false,
		Parameters: []Parameter{},
		Handler:    nil, // Abstract method
	}

	throwable.Methods["__toString"] = &Method{
		Name:       "__toString",
		Visibility: "public",
		IsStatic:   false,
		IsAbstract: true,
		IsFinal:    false,
		Parameters: []Parameter{},
		Handler:    nil, // Abstract method
	}

	stdlib.Classes["Throwable"] = throwable

	// Exception class
	exception := &Class{
		Name:       "Exception",
		Parent:     "Throwable",
		Properties: make(map[string]*Property),
		Methods:    make(map[string]*Method),
		Constants:  make(map[string]*values.Value),
		IsAbstract: false,
		IsFinal:    false,
	}

	// Exception properties
	exception.Properties["message"] = &Property{
		Name:         "message",
		Visibility:   "protected",
		IsStatic:     false,
		Type:         "string",
		DefaultValue: values.NewString(""),
	}

	exception.Properties["code"] = &Property{
		Name:         "code",
		Visibility:   "protected",
		IsStatic:     false,
		Type:         "int",
		DefaultValue: values.NewInt(0),
	}

	exception.Properties["file"] = &Property{
		Name:         "file",
		Visibility:   "protected",
		IsStatic:     false,
		Type:         "string",
		DefaultValue: values.NewString(""),
	}

	exception.Properties["line"] = &Property{
		Name:         "line",
		Visibility:   "protected",
		IsStatic:     false,
		Type:         "int",
		DefaultValue: values.NewInt(0),
	}

	exception.Properties["trace"] = &Property{
		Name:         "trace",
		Visibility:   "private",
		IsStatic:     false,
		Type:         "array",
		DefaultValue: values.NewArray(),
	}

	exception.Properties["previous"] = &Property{
		Name:         "previous",
		Visibility:   "private",
		IsStatic:     false,
		Type:         "Throwable",
		DefaultValue: values.NewNull(),
	}

	// Exception methods
	exception.Methods["__construct"] = &Method{
		Name:       "__construct",
		Visibility: "public",
		IsStatic:   false,
		IsAbstract: false,
		IsFinal:    false,
		Parameters: []Parameter{
			{Name: "message", Type: "string", IsReference: false, HasDefault: true, DefaultValue: values.NewString("")},
			{Name: "code", Type: "int", IsReference: false, HasDefault: true, DefaultValue: values.NewInt(0)},
			{Name: "previous", Type: "Throwable", IsReference: false, HasDefault: true, DefaultValue: values.NewNull()},
		},
		Handler: exceptionConstructHandler,
	}

	exception.Methods["getMessage"] = &Method{
		Name:       "getMessage",
		Visibility: "public",
		IsStatic:   false,
		IsAbstract: false,
		IsFinal:    true,
		Parameters: []Parameter{},
		Handler:    exceptionGetMessageHandler,
	}

	exception.Methods["getCode"] = &Method{
		Name:       "getCode",
		Visibility: "public",
		IsStatic:   false,
		IsAbstract: false,
		IsFinal:    true,
		Parameters: []Parameter{},
		Handler:    exceptionGetCodeHandler,
	}

	exception.Methods["getFile"] = &Method{
		Name:       "getFile",
		Visibility: "public",
		IsStatic:   false,
		IsAbstract: false,
		IsFinal:    true,
		Parameters: []Parameter{},
		Handler:    exceptionGetFileHandler,
	}

	exception.Methods["getLine"] = &Method{
		Name:       "getLine",
		Visibility: "public",
		IsStatic:   false,
		IsAbstract: false,
		IsFinal:    true,
		Parameters: []Parameter{},
		Handler:    exceptionGetLineHandler,
	}

	exception.Methods["getTrace"] = &Method{
		Name:       "getTrace",
		Visibility: "public",
		IsStatic:   false,
		IsAbstract: false,
		IsFinal:    true,
		Parameters: []Parameter{},
		Handler:    exceptionGetTraceHandler,
	}

	exception.Methods["getTraceAsString"] = &Method{
		Name:       "getTraceAsString",
		Visibility: "public",
		IsStatic:   false,
		IsAbstract: false,
		IsFinal:    true,
		Parameters: []Parameter{},
		Handler:    exceptionGetTraceAsStringHandler,
	}

	exception.Methods["getPrevious"] = &Method{
		Name:       "getPrevious",
		Visibility: "public",
		IsStatic:   false,
		IsAbstract: false,
		IsFinal:    true,
		Parameters: []Parameter{},
		Handler:    exceptionGetPreviousHandler,
	}

	exception.Methods["__toString"] = &Method{
		Name:       "__toString",
		Visibility: "public",
		IsStatic:   false,
		IsAbstract: false,
		IsFinal:    false,
		Parameters: []Parameter{},
		Handler:    exceptionToStringHandler,
	}

	stdlib.Classes["Exception"] = exception

	// Error class (PHP 7+)
	error := &Class{
		Name:       "Error",
		Parent:     "Throwable",
		Properties: exception.Properties, // Copy from Exception
		Methods:    make(map[string]*Method),
		Constants:  make(map[string]*values.Value),
		IsAbstract: false,
		IsFinal:    false,
	}

	// Copy methods from Exception
	for name, method := range exception.Methods {
		error.Methods[name] = method
	}

	stdlib.Classes["Error"] = error

	// RuntimeException
	runtimeException := &Class{
		Name:       "RuntimeException",
		Parent:     "Exception",
		Properties: exception.Properties,
		Methods:    exception.Methods,
		Constants:  make(map[string]*values.Value),
		IsAbstract: false,
		IsFinal:    false,
	}

	stdlib.Classes["RuntimeException"] = runtimeException

	// InvalidArgumentException
	invalidArgumentException := &Class{
		Name:       "InvalidArgumentException",
		Parent:     "Exception",
		Properties: exception.Properties,
		Methods:    exception.Methods,
		Constants:  make(map[string]*values.Value),
		IsAbstract: false,
		IsFinal:    false,
	}

	stdlib.Classes["InvalidArgumentException"] = invalidArgumentException

	// TypeError (PHP 7+)
	typeError := &Class{
		Name:       "TypeError",
		Parent:     "Error",
		Properties: error.Properties,
		Methods:    error.Methods,
		Constants:  make(map[string]*values.Value),
		IsAbstract: false,
		IsFinal:    false,
	}

	stdlib.Classes["TypeError"] = typeError
}

// initStdClass initializes the standard PHP object class
func (stdlib *StandardLibrary) initStdClass() {
	stdClass := &Class{
		Name:       "stdClass",
		Parent:     "",
		Properties: make(map[string]*Property),
		Methods:    make(map[string]*Method),
		Constants:  make(map[string]*values.Value),
		IsAbstract: false,
		IsFinal:    false,
	}

	stdlib.Classes["stdClass"] = stdClass
}

// initDateTimeClasses initializes DateTime-related classes
func (stdlib *StandardLibrary) initDateTimeClasses() {
	// DateTime class
	dateTime := &Class{
		Name:       "DateTime",
		Parent:     "",
		Properties: make(map[string]*Property),
		Methods:    make(map[string]*Method),
		Constants:  make(map[string]*values.Value),
		IsAbstract: false,
		IsFinal:    false,
	}

	// DateTime constants
	dateTime.Constants["ATOM"] = values.NewString("Y-m-d\\TH:i:sP")
	dateTime.Constants["COOKIE"] = values.NewString("l, d-M-Y H:i:s T")
	dateTime.Constants["ISO8601"] = values.NewString("Y-m-d\\TH:i:sO")
	dateTime.Constants["RFC822"] = values.NewString("D, d M y H:i:s O")
	dateTime.Constants["RFC850"] = values.NewString("l, d-M-y H:i:s T")
	dateTime.Constants["RFC1036"] = values.NewString("D, d M y H:i:s O")
	dateTime.Constants["RFC1123"] = values.NewString("D, d M Y H:i:s O")
	dateTime.Constants["RFC7231"] = values.NewString("D, d M Y H:i:s \\G\\M\\T")
	dateTime.Constants["RFC2822"] = values.NewString("D, d M Y H:i:s O")
	dateTime.Constants["RFC3339"] = values.NewString("Y-m-d\\TH:i:sP")
	dateTime.Constants["RFC3339_EXTENDED"] = values.NewString("Y-m-d\\TH:i:s.vP")
	dateTime.Constants["RSS"] = values.NewString("D, d M Y H:i:s O")
	dateTime.Constants["W3C"] = values.NewString("Y-m-d\\TH:i:sP")

	// DateTime methods
	dateTime.Methods["__construct"] = &Method{
		Name:       "__construct",
		Visibility: "public",
		IsStatic:   false,
		IsAbstract: false,
		IsFinal:    false,
		Parameters: []Parameter{
			{Name: "datetime", Type: "string", IsReference: false, HasDefault: true, DefaultValue: values.NewString("now")},
			{Name: "timezone", Type: "DateTimeZone", IsReference: false, HasDefault: true, DefaultValue: values.NewNull()},
		},
		Handler: dateTimeConstructHandler,
	}

	dateTime.Methods["format"] = &Method{
		Name:       "format",
		Visibility: "public",
		IsStatic:   false,
		IsAbstract: false,
		IsFinal:    false,
		Parameters: []Parameter{
			{Name: "format", Type: "string", IsReference: false, HasDefault: false},
		},
		Handler: dateTimeFormatHandler,
	}

	dateTime.Methods["getTimestamp"] = &Method{
		Name:       "getTimestamp",
		Visibility: "public",
		IsStatic:   false,
		IsAbstract: false,
		IsFinal:    false,
		Parameters: []Parameter{},
		Handler:    dateTimeGetTimestampHandler,
	}

	stdlib.Classes["DateTime"] = dateTime

	// DateTimeZone class
	dateTimeZone := &Class{
		Name:       "DateTimeZone",
		Parent:     "",
		Properties: make(map[string]*Property),
		Methods:    make(map[string]*Method),
		Constants:  make(map[string]*values.Value),
		IsAbstract: false,
		IsFinal:    false,
	}

	dateTimeZone.Methods["__construct"] = &Method{
		Name:       "__construct",
		Visibility: "public",
		IsStatic:   false,
		IsAbstract: false,
		IsFinal:    false,
		Parameters: []Parameter{
			{Name: "timezone", Type: "string", IsReference: false, HasDefault: false},
		},
		Handler: dateTimeZoneConstructHandler,
	}

	dateTimeZone.Methods["getName"] = &Method{
		Name:       "getName",
		Visibility: "public",
		IsStatic:   false,
		IsAbstract: false,
		IsFinal:    false,
		Parameters: []Parameter{},
		Handler:    dateTimeZoneGetNameHandler,
	}

	stdlib.Classes["DateTimeZone"] = dateTimeZone
}

// initReflectionClasses initializes Reflection API classes
func (stdlib *StandardLibrary) initReflectionClasses() {
	// ReflectionClass
	reflectionClass := &Class{
		Name:       "ReflectionClass",
		Parent:     "",
		Properties: make(map[string]*Property),
		Methods:    make(map[string]*Method),
		Constants:  make(map[string]*values.Value),
		IsAbstract: false,
		IsFinal:    false,
	}

	reflectionClass.Properties["name"] = &Property{
		Name:         "name",
		Visibility:   "public",
		IsStatic:     false,
		Type:         "string",
		DefaultValue: values.NewString(""),
	}

	reflectionClass.Methods["__construct"] = &Method{
		Name:       "__construct",
		Visibility: "public",
		IsStatic:   false,
		IsAbstract: false,
		IsFinal:    false,
		Parameters: []Parameter{
			{Name: "objectOrClass", Type: "mixed", IsReference: false, HasDefault: false},
		},
		Handler: reflectionClassConstructHandler,
	}

	reflectionClass.Methods["getName"] = &Method{
		Name:       "getName",
		Visibility: "public",
		IsStatic:   false,
		IsAbstract: false,
		IsFinal:    false,
		Parameters: []Parameter{},
		Handler:    reflectionClassGetNameHandler,
	}

	reflectionClass.Methods["getParentClass"] = &Method{
		Name:       "getParentClass",
		Visibility: "public",
		IsStatic:   false,
		IsAbstract: false,
		IsFinal:    false,
		Parameters: []Parameter{},
		Handler:    reflectionClassGetParentClassHandler,
	}

	reflectionClass.Methods["hasMethod"] = &Method{
		Name:       "hasMethod",
		Visibility: "public",
		IsStatic:   false,
		IsAbstract: false,
		IsFinal:    false,
		Parameters: []Parameter{
			{Name: "name", Type: "string", IsReference: false, HasDefault: false},
		},
		Handler: reflectionClassHasMethodHandler,
	}

	stdlib.Classes["ReflectionClass"] = reflectionClass
}

// Exception class method handlers

func exceptionConstructHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	// Constructor implementation would set properties
	return values.NewNull(), nil
}

func exceptionGetMessageHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	// Would retrieve message property from $this
	return values.NewString("Exception message"), nil
}

func exceptionGetCodeHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	// Would retrieve code property from $this
	return values.NewInt(0), nil
}

func exceptionGetFileHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	// Would retrieve file property from $this
	return values.NewString(__FILE__), nil
}

func exceptionGetLineHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	// Would retrieve line property from $this
	return values.NewInt(0), nil
}

func exceptionGetTraceHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	// Would return stack trace array
	return values.NewArray(), nil
}

func exceptionGetTraceAsStringHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	// Would return formatted stack trace string
	return values.NewString("Stack trace"), nil
}

func exceptionGetPreviousHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	// Would retrieve previous exception
	return values.NewNull(), nil
}

func exceptionToStringHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	// Would return formatted exception string
	return values.NewString("Exception"), nil
}

// DateTime class method handlers

func dateTimeConstructHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	// Constructor would parse datetime string and set internal timestamp
	return values.NewNull(), nil
}

func dateTimeFormatHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("DateTime::format() expects exactly 1 parameter, %d given", len(args))
	}

	format := args[0].ToString()
	// Simplified formatting - would use internal timestamp
	formatted := strings.ReplaceAll(format, "Y", "2024")
	formatted = strings.ReplaceAll(formatted, "m", "01")
	formatted = strings.ReplaceAll(formatted, "d", "01")
	formatted = strings.ReplaceAll(formatted, "H", "12")
	formatted = strings.ReplaceAll(formatted, "i", "00")
	formatted = strings.ReplaceAll(formatted, "s", "00")

	return values.NewString(formatted), nil
}

func dateTimeGetTimestampHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	// Would return internal timestamp
	return values.NewInt(1640995200), nil // 2022-01-01 00:00:00 UTC
}

// DateTimeZone class method handlers

func dateTimeZoneConstructHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("DateTimeZone::__construct() expects exactly 1 parameter, %d given", len(args))
	}

	// Would validate and store timezone
	return values.NewNull(), nil
}

func dateTimeZoneGetNameHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	// Would return stored timezone name
	return values.NewString("UTC"), nil
}

// ReflectionClass method handlers

func reflectionClassConstructHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("ReflectionClass::__construct() expects exactly 1 parameter, %d given", len(args))
	}

	// Would set up reflection for the given class/object
	return values.NewNull(), nil
}

func reflectionClassGetNameHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	// Would return the class name being reflected
	return values.NewString("stdClass"), nil
}

func reflectionClassGetParentClassHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	// Would return parent class reflection or false
	return values.NewBool(false), nil
}

func reflectionClassHasMethodHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("ReflectionClass::hasMethod() expects exactly 1 parameter, %d given", len(args))
	}

	// Would check if class has the specified method
	return values.NewBool(false), nil
}

// Helper function to get current file name (would be set during compilation)
const __FILE__ = "unknown"
