package stdlib

import (
	"fmt"
	"strings"
	"time"

	"github.com/wudi/php-parser/compiler/registry"
	"github.com/wudi/php-parser/compiler/values"
)

// InitializeUnifiedClasses initializes all built-in classes using the unified registry
func InitializeUnifiedClasses() error {
	// Initialize registry if not already done
	if registry.GlobalRegistry == nil {
		registry.Initialize()
	}

	// Initialize exception classes
	if err := initUnifiedExceptionClasses(); err != nil {
		return fmt.Errorf("failed to initialize exception classes: %v", err)
	}

	// Initialize DateTime classes
	if err := initUnifiedDateTimeClasses(); err != nil {
		return fmt.Errorf("failed to initialize DateTime classes: %v", err)
	}

	// Initialize Reflection classes
	if err := initUnifiedReflectionClasses(); err != nil {
		return fmt.Errorf("failed to initialize Reflection classes: %v", err)
	}

	return nil
}

// initUnifiedExceptionClasses initializes the exception class hierarchy using unified registry
func initUnifiedExceptionClasses() error {
	// Throwable interface (abstract class in unified system)
	_, err := registry.NewClass("Throwable").
		Abstract().
		BuiltinClass("core").
		AddMethod("getMessage", "public").Abstract().Done().
		AddMethod("getCode", "public").Abstract().Done().
		AddMethod("getFile", "public").Abstract().Done().
		AddMethod("getLine", "public").Abstract().Done().
		AddMethod("getTrace", "public").Abstract().Done().
		AddMethod("getTraceAsString", "public").Abstract().Done().
		AddMethod("getPrevious", "public").Abstract().Done().
		AddMethod("__toString", "public").Abstract().Done().
		BuildAndRegister()

	if err != nil {
		return err
	}

	// Exception class (inherits from Throwable)
	_, err = registry.NewClass("Exception").
		Extends("Throwable").
		BuiltinClass("core").
		AddProperty("message", "protected", "string", values.NewString("")).
		AddProperty("code", "protected", "int", values.NewInt(0)).
		AddProperty("file", "protected", "string", values.NewString("")).
		AddProperty("line", "protected", "int", values.NewInt(0)).
		AddProperty("trace", "private", "array", values.NewArray()).
		AddProperty("previous", "private", "Throwable", values.NewNull()).
		Constructor().
		WithOptionalParameter("message", "string", values.NewString("")).
		WithOptionalParameter("code", "int", values.NewInt(0)).
		WithOptionalParameter("previous", "Throwable", values.NewNull()).
		WithNativeHandler(unifiedExceptionConstructHandler).
		Done().
		AddMethod("getMessage", "public").Final().
		WithNativeHandler(unifiedExceptionGetMessageHandler).
		Done().
		AddMethod("getCode", "public").Final().
		WithNativeHandler(unifiedExceptionGetCodeHandler).
		Done().
		AddMethod("getFile", "public").Final().
		WithNativeHandler(unifiedExceptionGetFileHandler).
		Done().
		AddMethod("getLine", "public").Final().
		WithNativeHandler(unifiedExceptionGetLineHandler).
		Done().
		AddMethod("getTrace", "public").Final().
		WithNativeHandler(unifiedExceptionGetTraceHandler).
		Done().
		AddMethod("getTraceAsString", "public").Final().
		WithNativeHandler(unifiedExceptionGetTraceAsStringHandler).
		Done().
		AddMethod("getPrevious", "public").Final().
		WithNativeHandler(unifiedExceptionGetPreviousHandler).
		Done().
		AddMethod("__toString", "public").
		WithNativeHandler(unifiedExceptionToStringHandler).
		Done().
		BuildAndRegister()

	if err != nil {
		return err
	}

	// Error class (PHP 7+)
	_, err = registry.NewClass("Error").
		Extends("Throwable").
		BuiltinClass("core").
		AddProperty("message", "protected", "string", values.NewString("")).
		AddProperty("code", "protected", "int", values.NewInt(0)).
		AddProperty("file", "protected", "string", values.NewString("")).
		AddProperty("line", "protected", "int", values.NewInt(0)).
		Constructor().
		WithOptionalParameter("message", "string", values.NewString("")).
		WithOptionalParameter("code", "int", values.NewInt(0)).
		WithNativeHandler(unifiedExceptionConstructHandler). // Reuse exception handler
		Done().
		AddMethod("getMessage", "public").Final().
		WithNativeHandler(unifiedExceptionGetMessageHandler).
		Done().
		AddMethod("getCode", "public").Final().
		WithNativeHandler(unifiedExceptionGetCodeHandler).
		Done().
		BuildAndRegister()

	if err != nil {
		return err
	}

	// RuntimeException
	_, err = registry.NewClass("RuntimeException").
		Extends("Exception").
		BuiltinClass("core").
		BuildAndRegister()

	if err != nil {
		return err
	}

	// InvalidArgumentException
	_, err = registry.NewClass("InvalidArgumentException").
		Extends("Exception").
		BuiltinClass("core").
		BuildAndRegister()

	if err != nil {
		return err
	}

	// TypeError
	_, err = registry.NewClass("TypeError").
		Extends("Error").
		BuiltinClass("core").
		BuildAndRegister()

	return err
}

// initUnifiedDateTimeClasses initializes DateTime classes using unified registry
func initUnifiedDateTimeClasses() error {
	// DateTime class
	_, err := registry.NewClass("DateTime").
		BuiltinClass("date").
		AddConstant("ATOM", values.NewString("Y-m-d\\TH:i:sP")).
		AddConstant("COOKIE", values.NewString("l, d-M-Y H:i:s T")).
		AddConstant("ISO8601", values.NewString("Y-m-d\\TH:i:sO")).
		AddConstant("RFC822", values.NewString("D, d M y H:i:s O")).
		AddConstant("RFC850", values.NewString("l, d-M-y H:i:s T")).
		AddConstant("RFC1036", values.NewString("D, d M y H:i:s O")).
		AddConstant("RFC1123", values.NewString("D, d M Y H:i:s O")).
		AddConstant("RFC7231", values.NewString("D, d M Y H:i:s \\G\\M\\T")).
		AddConstant("RFC2822", values.NewString("D, d M Y H:i:s O")).
		AddConstant("RFC3339", values.NewString("Y-m-d\\TH:i:sP")).
		AddConstant("RFC3339_EXTENDED", values.NewString("Y-m-d\\TH:i:s.vP")).
		AddConstant("RSS", values.NewString("D, d M Y H:i:s O")).
		AddConstant("W3C", values.NewString("Y-m-d\\TH:i:sP")).
		Constructor().
		WithOptionalParameter("datetime", "string", values.NewString("now")).
		WithOptionalParameter("timezone", "DateTimeZone", values.NewNull()).
		WithNativeHandler(unifiedDateTimeConstructHandler).
		Done().
		AddMethod("format", "public").
		WithParameter("format", "string").
		WithNativeHandler(unifiedDateTimeFormatHandler).
		Done().
		AddMethod("getTimestamp", "public").
		WithNativeHandler(unifiedDateTimeGetTimestampHandler).
		Done().
		BuildAndRegister()

	if err != nil {
		return err
	}

	// DateTimeZone class
	_, err = registry.NewClass("DateTimeZone").
		BuiltinClass("date").
		Constructor().
		WithParameter("timezone", "string").
		WithNativeHandler(unifiedDateTimeZoneConstructHandler).
		Done().
		AddMethod("getName", "public").
		WithNativeHandler(unifiedDateTimeZoneGetNameHandler).
		Done().
		BuildAndRegister()

	return err
}

// initUnifiedReflectionClasses initializes Reflection classes using unified registry
func initUnifiedReflectionClasses() error {
	// ReflectionClass
	_, err := registry.NewClass("ReflectionClass").
		BuiltinClass("reflection").
		AddProperty("name", "public", "string", values.NewString("")).
		Constructor().
		WithParameter("objectOrClass", "mixed").
		WithNativeHandler(unifiedReflectionClassConstructHandler).
		Done().
		AddMethod("getName", "public").
		WithNativeHandler(unifiedReflectionClassGetNameHandler).
		Done().
		AddMethod("getParentClass", "public").
		WithNativeHandler(unifiedReflectionClassGetParentClassHandler).
		Done().
		AddMethod("hasMethod", "public").
		WithParameter("name", "string").
		WithNativeHandler(unifiedReflectionClassHasMethodHandler).
		Done().
		BuildAndRegister()

	return err
}

// Unified method handlers

// Exception handlers - simplified for unified system
func unifiedExceptionConstructHandler(ctx registry.ExecutionContext, args []*values.Value) (*values.Value, error) {
	return values.NewNull(), nil
}

func unifiedExceptionGetMessageHandler(ctx registry.ExecutionContext, args []*values.Value) (*values.Value, error) {
	return values.NewString(""), nil
}

func unifiedExceptionGetCodeHandler(ctx registry.ExecutionContext, args []*values.Value) (*values.Value, error) {
	return values.NewInt(0), nil
}

func unifiedExceptionGetFileHandler(ctx registry.ExecutionContext, args []*values.Value) (*values.Value, error) {
	return values.NewString(""), nil
}

func unifiedExceptionGetLineHandler(ctx registry.ExecutionContext, args []*values.Value) (*values.Value, error) {
	return values.NewInt(0), nil
}

func unifiedExceptionGetTraceHandler(ctx registry.ExecutionContext, args []*values.Value) (*values.Value, error) {
	return values.NewArray(), nil
}

func unifiedExceptionGetTraceAsStringHandler(ctx registry.ExecutionContext, args []*values.Value) (*values.Value, error) {
	return values.NewString(""), nil
}

func unifiedExceptionGetPreviousHandler(ctx registry.ExecutionContext, args []*values.Value) (*values.Value, error) {
	return values.NewNull(), nil
}

func unifiedExceptionToStringHandler(ctx registry.ExecutionContext, args []*values.Value) (*values.Value, error) {
	return values.NewString("Exception"), nil
}

// DateTime handlers
func unifiedDateTimeConstructHandler(ctx registry.ExecutionContext, args []*values.Value) (*values.Value, error) {
	return values.NewNull(), nil
}

func unifiedDateTimeFormatHandler(ctx registry.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("DateTime::format() expects exactly 1 parameter, %d given", len(args))
	}

	format := args[0].ToString()
	now := time.Now()

	// Basic format replacements
	formatted := strings.ReplaceAll(format, "Y", fmt.Sprintf("%04d", now.Year()))
	formatted = strings.ReplaceAll(formatted, "m", fmt.Sprintf("%02d", int(now.Month())))
	formatted = strings.ReplaceAll(formatted, "d", fmt.Sprintf("%02d", now.Day()))
	formatted = strings.ReplaceAll(formatted, "H", fmt.Sprintf("%02d", now.Hour()))
	formatted = strings.ReplaceAll(formatted, "i", fmt.Sprintf("%02d", now.Minute()))
	formatted = strings.ReplaceAll(formatted, "s", fmt.Sprintf("%02d", now.Second()))

	return values.NewString(formatted), nil
}

func unifiedDateTimeGetTimestampHandler(ctx registry.ExecutionContext, args []*values.Value) (*values.Value, error) {
	return values.NewInt(time.Now().Unix()), nil
}

// DateTimeZone handlers
func unifiedDateTimeZoneConstructHandler(ctx registry.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("DateTimeZone::__construct() expects exactly 1 parameter, %d given", len(args))
	}
	return values.NewNull(), nil
}

func unifiedDateTimeZoneGetNameHandler(ctx registry.ExecutionContext, args []*values.Value) (*values.Value, error) {
	return values.NewString("UTC"), nil
}

// ReflectionClass handlers
func unifiedReflectionClassConstructHandler(ctx registry.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("ReflectionClass::__construct() expects exactly 1 parameter, %d given", len(args))
	}
	return values.NewNull(), nil
}

func unifiedReflectionClassGetNameHandler(ctx registry.ExecutionContext, args []*values.Value) (*values.Value, error) {
	return values.NewString("stdClass"), nil
}

func unifiedReflectionClassGetParentClassHandler(ctx registry.ExecutionContext, args []*values.Value) (*values.Value, error) {
	return values.NewBool(false), nil
}

func unifiedReflectionClassHasMethodHandler(ctx registry.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("ReflectionClass::hasMethod() expects exactly 1 parameter, %d given", len(args))
	}
	return values.NewBool(false), nil
}
