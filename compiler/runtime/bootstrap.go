package runtime

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/wudi/php-parser/compiler/values"
)

// Bootstrap initializes the runtime with all built-in entities
func Bootstrap() error {
	Initialize()

	if err := registerBuiltinConstants(); err != nil {
		return fmt.Errorf("failed to register built-in constants: %v", err)
	}

	if err := registerBuiltinVariables(); err != nil {
		return fmt.Errorf("failed to register built-in variables: %v", err)
	}

	if err := registerBuiltinFunctions(); err != nil {
		return fmt.Errorf("failed to register built-in functions: %v", err)
	}

	if err := registerBuiltinClasses(); err != nil {
		return fmt.Errorf("failed to register built-in classes: %v", err)
	}

	return nil
}

// registerBuiltinConstants registers all PHP built-in constants
func registerBuiltinConstants() error {
	constants := map[string]*values.Value{
		// PHP Version constants
		"PHP_VERSION":         values.NewString("8.4.0"),
		"PHP_MAJOR_VERSION":   values.NewInt(8),
		"PHP_MINOR_VERSION":   values.NewInt(4),
		"PHP_RELEASE_VERSION": values.NewInt(0),
		"PHP_VERSION_ID":      values.NewInt(80400),
		"PHP_EXTRA_VERSION":   values.NewString(""),

		// Boolean constants
		"TRUE":  values.NewBool(true),
		"FALSE": values.NewBool(false),
		"NULL":  values.NewNull(),

		// Math constants
		"PHP_INT_MAX":   values.NewInt(int64(math.MaxInt64)),
		"PHP_INT_MIN":   values.NewInt(int64(math.MinInt64)),
		"PHP_FLOAT_MAX": values.NewFloat(math.MaxFloat64),
		"PHP_FLOAT_MIN": values.NewFloat(math.SmallestNonzeroFloat64),
		"M_PI":          values.NewFloat(math.Pi),
		"M_E":           values.NewFloat(math.E),
		"M_LOG2E":       values.NewFloat(math.Log2E),
		"M_LOG10E":      values.NewFloat(math.Log10E),
		"M_LN2":         values.NewFloat(math.Ln2),
		"M_LN10":        values.NewFloat(math.Ln10),
		"INF":           values.NewFloat(math.Inf(1)),
		"NAN":           values.NewFloat(math.NaN()),

		// System constants
		"PHP_OS":              values.NewString("Linux"),
		"PHP_OS_FAMILY":       values.NewString("Linux"),
		"PHP_SAPI":            values.NewString("cli"),
		"DIRECTORY_SEPARATOR": values.NewString("/"),
		"PATH_SEPARATOR":      values.NewString(":"),
		"PHP_EOL":             values.NewString("\n"),

		// Error constants
		"E_ERROR":           values.NewInt(1),
		"E_WARNING":         values.NewInt(2),
		"E_PARSE":           values.NewInt(4),
		"E_NOTICE":          values.NewInt(8),
		"E_CORE_ERROR":      values.NewInt(16),
		"E_CORE_WARNING":    values.NewInt(32),
		"E_COMPILE_ERROR":   values.NewInt(64),
		"E_COMPILE_WARNING": values.NewInt(128),
		"E_USER_ERROR":      values.NewInt(256),
		"E_USER_WARNING":    values.NewInt(512),
		"E_USER_NOTICE":     values.NewInt(1024),
		"E_ALL":             values.NewInt(32767),

		// Case constants
		"CASE_LOWER": values.NewInt(0),
		"CASE_UPPER": values.NewInt(1),

		// Sort constants
		"SORT_REGULAR":       values.NewInt(0),
		"SORT_NUMERIC":       values.NewInt(1),
		"SORT_STRING":        values.NewInt(2),
		"SORT_LOCALE_STRING": values.NewInt(5),
		"SORT_NATURAL":       values.NewInt(6),
		"SORT_FLAG_CASE":     values.NewInt(8),
		"SORT_ASC":           values.NewInt(4),
		"SORT_DESC":          values.NewInt(3),
	}

	for name, value := range constants {
		if err := RegisterBuiltinConstant(name, value); err != nil {
			return err
		}
	}

	return nil
}

// registerBuiltinVariables registers all PHP built-in variables
func registerBuiltinVariables() error {
	// Create $_SERVER superglobal
	server := createServerArray()
	if err := GlobalRegistry.RegisterVariable("_SERVER", server, true, ""); err != nil {
		return err
	}

	// Create other superglobals
	superglobals := map[string]*values.Value{
		"_GET":                 values.NewArray(),
		"_POST":                values.NewArray(),
		"_FILES":               values.NewArray(),
		"_COOKIE":              values.NewArray(),
		"_SESSION":             values.NewArray(),
		"_REQUEST":             values.NewArray(),
		"_ENV":                 createEnvArray(),
		"GLOBALS":              values.NewArray(),
		"argc":                 values.NewInt(0),
		"http_response_header": values.NewArray(),
	}

	for name, value := range superglobals {
		if err := GlobalRegistry.RegisterVariable(name, value, true, ""); err != nil {
			return err
		}
	}

	// Create $argv
	argv := values.NewArray()
	argv.ArraySet(values.NewInt(0), values.NewString("php-parser"))
	if err := GlobalRegistry.RegisterVariable("argv", argv, true, ""); err != nil {
		return err
	}

	return nil
}

// createServerArray creates the $_SERVER superglobal array
func createServerArray() *values.Value {
	server := values.NewArray()

	// Basic server information
	serverVars := map[string]string{
		"SERVER_SOFTWARE":       "PHP-Parser/1.0",
		"SERVER_NAME":           "localhost",
		"SERVER_ADDR":           "127.0.0.1",
		"SERVER_PORT":           "80",
		"REMOTE_ADDR":           "127.0.0.1",
		"DOCUMENT_ROOT":         "/",
		"SERVER_ADMIN":          "admin@localhost",
		"SCRIPT_FILENAME":       "/index.php",
		"REMOTE_PORT":           "0",
		"GATEWAY_INTERFACE":     "CGI/1.1",
		"SERVER_PROTOCOL":       "HTTP/1.1",
		"REQUEST_METHOD":        "GET",
		"QUERY_STRING":          "",
		"HTTP_ACCEPT":           "*/*",
		"HTTP_HOST":             "localhost",
		"HTTP_USER_AGENT":       "PHP-Parser/1.0",
		"REQUEST_URI":           "/",
		"SCRIPT_NAME":           "/index.php",
		"PHP_SELF":              "/index.php",
		"PHP_CLI_PROCESS_TITLE": "php",
		"_":                     "/usr/bin/php",
	}

	for key, val := range serverVars {
		server.ArraySet(values.NewString(key), values.NewString(val))
	}

	// Add time-based values
	now := time.Now()
	server.ArraySet(values.NewString("REQUEST_TIME"), values.NewInt(now.Unix()))
	server.ArraySet(values.NewString("REQUEST_TIME_FLOAT"), values.NewFloat(float64(now.UnixNano())/1e9))

	return server
}

// createEnvArray creates the $_ENV superglobal array
func createEnvArray() *values.Value {
	env := values.NewArray()

	// Common environment variables
	envVars := map[string]string{
		"PATH":   "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		"HOME":   "/home/user",
		"USER":   "user",
		"SHELL":  "/bin/bash",
		"LANG":   "en_US.UTF-8",
		"TERM":   "xterm",
		"PWD":    "/",
		"TMPDIR": "/tmp",
	}

	for key, val := range envVars {
		env.ArraySet(values.NewString(key), values.NewString(val))
	}

	return env
}

// registerBuiltinFunctions registers all PHP built-in functions
func registerBuiltinFunctions() error {
	functions := []*FunctionDescriptor{
		// String functions
		{
			Name:    "strlen",
			Handler: strlenHandler,
			Parameters: []ParameterDescriptor{
				{Name: "string", Type: "string"},
			},
			MinArgs: 1,
			MaxArgs: 1,
		},
		{
			Name:    "substr",
			Handler: substrHandler,
			Parameters: []ParameterDescriptor{
				{Name: "string", Type: "string"},
				{Name: "start", Type: "int"},
				{Name: "length", Type: "int", HasDefault: true, DefaultValue: values.NewNull()},
			},
			MinArgs: 2,
			MaxArgs: 3,
		},

		// Array functions
		{
			Name:    "count",
			Handler: countHandler,
			Parameters: []ParameterDescriptor{
				{Name: "array", Type: "array|countable"},
				{Name: "mode", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
			},
			MinArgs: 1,
			MaxArgs: 2,
		},

		// Type checking functions
		{
			Name:    "is_string",
			Handler: isStringHandler,
			Parameters: []ParameterDescriptor{
				{Name: "value", Type: "mixed"},
			},
			MinArgs: 1,
			MaxArgs: 1,
		},
		{
			Name:    "is_int",
			Handler: isIntHandler,
			Parameters: []ParameterDescriptor{
				{Name: "value", Type: "mixed"},
			},
			MinArgs: 1,
			MaxArgs: 1,
		},
		{
			Name:    "is_array",
			Handler: isArrayHandler,
			Parameters: []ParameterDescriptor{
				{Name: "value", Type: "mixed"},
			},
			MinArgs: 1,
			MaxArgs: 1,
		},
		{
			Name:    "is_null",
			Handler: isNullHandler,
			Parameters: []ParameterDescriptor{
				{Name: "value", Type: "mixed"},
			},
			MinArgs: 1,
			MaxArgs: 1,
		},

		// Math functions
		{
			Name:    "abs",
			Handler: absHandler,
			Parameters: []ParameterDescriptor{
				{Name: "number", Type: "int|float"},
			},
			MinArgs: 1,
			MaxArgs: 1,
		},
		{
			Name:    "max",
			Handler: maxHandler,
			Parameters: []ParameterDescriptor{
				{Name: "values", Type: "mixed"},
			},
			IsVariadic: true,
			MinArgs:    1,
			MaxArgs:    -1,
		},
		{
			Name:    "min",
			Handler: minHandler,
			Parameters: []ParameterDescriptor{
				{Name: "values", Type: "mixed"},
			},
			IsVariadic: true,
			MinArgs:    1,
			MaxArgs:    -1,
		},

		// Output functions
		{
			Name:       "var_dump",
			Handler:    varDumpHandler,
			IsVariadic: true,
			MinArgs:    1,
			MaxArgs:    -1,
		},
		{
			Name:    "print_r",
			Handler: printRHandler,
			Parameters: []ParameterDescriptor{
				{Name: "value", Type: "mixed"},
				{Name: "return", Type: "bool", HasDefault: true, DefaultValue: values.NewBool(false)},
			},
			MinArgs: 1,
			MaxArgs: 2,
		},

		// String functions
		{
			Name:    "str_repeat",
			Handler: strRepeatHandler,
			Parameters: []ParameterDescriptor{
				{Name: "input", Type: "string"},
				{Name: "multiplier", Type: "int"},
			},
			MinArgs: 2,
			MaxArgs: 2,
		},

		// Introspection functions
		{
			Name:    "function_exists",
			Handler: functionExistsHandler,
			Parameters: []ParameterDescriptor{
				{Name: "function_name", Type: "string"},
			},
			MinArgs: 1,
			MaxArgs: 1,
		},
		{
			Name:    "class_exists",
			Handler: classExistsHandler,
			Parameters: []ParameterDescriptor{
				{Name: "class_name", Type: "string"},
			},
			MinArgs: 1,
			MaxArgs: 1,
		},
		{
			Name:    "method_exists",
			Handler: methodExistsHandler,
			Parameters: []ParameterDescriptor{
				{Name: "object_or_class", Type: "string|object"},
				{Name: "method", Type: "string"},
			},
			MinArgs: 2,
			MaxArgs: 2,
		},

		// Concurrency functions
		{
			Name:    "go",
			Handler: goHandler,
			Parameters: []ParameterDescriptor{
				{Name: "closure", Type: "callable"},
			},
			IsVariadic: true,
			MinArgs:    1,
			MaxArgs:    -1,
		},
		{
			Name:    "sleep",
			Handler: sleepHandler,
			Parameters: []ParameterDescriptor{
				{Name: "seconds", Type: "int"},
			},
			IsVariadic: false,
			MinArgs:    1,
			MaxArgs:    -1,
		},
		{
			Name:       "time",
			Handler:    timeHandler,
			IsVariadic: false,
			MinArgs:    0,
			MaxArgs:    -1,
		},
		{
			Name:    "microtime",
			Handler: microtimeHandler,
			Parameters: []ParameterDescriptor{
				{Name: "as_float", Type: "bool"},
			},
			IsVariadic: false,
			MinArgs:    0,
			MaxArgs:    -1,
		},
	}

	for _, desc := range functions {
		if err := RegisterBuiltinFunction(desc); err != nil {
			return err
		}
	}

	return nil
}

// registerBuiltinClasses registers all PHP built-in classes
func registerBuiltinClasses() error {
	// Exception class hierarchy
	exceptionClass := &ClassDescriptor{
		Name: "Exception",
		Properties: map[string]*PropertyDescriptor{
			"message": {
				Name:         "message",
				Type:         "string",
				Visibility:   "protected",
				DefaultValue: values.NewString(""),
			},
			"code": {
				Name:         "code",
				Type:         "int",
				Visibility:   "protected",
				DefaultValue: values.NewInt(0),
			},
			"file": {
				Name:       "file",
				Type:       "string",
				Visibility: "protected",
			},
			"line": {
				Name:       "line",
				Type:       "int",
				Visibility: "protected",
			},
		},
		Methods: map[string]*MethodDescriptor{
			"__construct": {
				Name:       "__construct",
				Visibility: "public",
				Parameters: []ParameterDescriptor{
					{Name: "message", Type: "string", HasDefault: true, DefaultValue: values.NewString("")},
					{Name: "code", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
				},
			},
			"getMessage": {
				Name:       "getMessage",
				Visibility: "public",
			},
			"getCode": {
				Name:       "getCode",
				Visibility: "public",
			},
		},
	}

	if err := GlobalRegistry.RegisterClass(exceptionClass); err != nil {
		return err
	}

	// stdClass
	stdClass := &ClassDescriptor{
		Name:       "stdClass",
		Properties: make(map[string]*PropertyDescriptor),
		Methods:    make(map[string]*MethodDescriptor),
	}

	if err := GlobalRegistry.RegisterClass(stdClass); err != nil {
		return err
	}

	// WaitGroup class for concurrency
	waitGroupClass := &ClassDescriptor{
		Name:       "WaitGroup",
		Properties: make(map[string]*PropertyDescriptor),
		Methods: map[string]*MethodDescriptor{
			"__construct": {
				Name:       "__construct",
				Visibility: "public",
				Handler:    waitGroupConstructHandler,
			},
			"Add": {
				Name:       "Add",
				Visibility: "public",
				Parameters: []ParameterDescriptor{
					{Name: "delta", Type: "int"},
				},
				Handler: waitGroupAddHandler,
			},
			"Done": {
				Name:       "Done",
				Visibility: "public",
				Handler:    waitGroupDoneHandler,
			},
			"Wait": {
				Name:       "Wait",
				Visibility: "public",
				Handler:    waitGroupWaitHandler,
			},
		},
	}

	if err := GlobalRegistry.RegisterClass(waitGroupClass); err != nil {
		return err
	}

	return nil
}

// Function handlers
func strlenHandler(ctx ExecutionContext, args []*values.Value) (*values.Value, error) {
	return values.NewInt(int64(len(args[0].ToString()))), nil
}

func substrHandler(ctx ExecutionContext, args []*values.Value) (*values.Value, error) {
	str := args[0].ToString()
	start := int(args[1].ToInt())

	if start < 0 {
		start = len(str) + start
	}
	if start < 0 || start >= len(str) {
		return values.NewString(""), nil
	}

	if len(args) == 3 && !args[2].IsNull() {
		length := int(args[2].ToInt())
		if length <= 0 {
			return values.NewString(""), nil
		}
		end := start + length
		if end > len(str) {
			end = len(str)
		}
		return values.NewString(str[start:end]), nil
	}

	return values.NewString(str[start:]), nil
}

func countHandler(ctx ExecutionContext, args []*values.Value) (*values.Value, error) {
	value := args[0]
	if value.IsArray() {
		return values.NewInt(int64(value.ArrayCount())), nil
	}
	if value.IsNull() {
		return values.NewInt(0), nil
	}
	return values.NewInt(1), nil
}

func isStringHandler(ctx ExecutionContext, args []*values.Value) (*values.Value, error) {
	return values.NewBool(args[0].IsString()), nil
}

func isIntHandler(ctx ExecutionContext, args []*values.Value) (*values.Value, error) {
	return values.NewBool(args[0].IsInt()), nil
}

func isArrayHandler(ctx ExecutionContext, args []*values.Value) (*values.Value, error) {
	return values.NewBool(args[0].IsArray()), nil
}

func isNullHandler(ctx ExecutionContext, args []*values.Value) (*values.Value, error) {
	return values.NewBool(args[0].IsNull()), nil
}

func absHandler(ctx ExecutionContext, args []*values.Value) (*values.Value, error) {
	value := args[0]
	if value.IsInt() {
		n := value.ToInt()
		if n < 0 {
			return values.NewInt(-n), nil
		}
		return values.NewInt(n), nil
	}
	if value.IsFloat() {
		f := value.ToFloat()
		return values.NewFloat(math.Abs(f)), nil
	}
	// Convert to number first
	if value.IsString() {
		str := value.ToString()
		if i, err := strconv.ParseInt(str, 10, 64); err == nil {
			if i < 0 {
				return values.NewInt(-i), nil
			}
			return values.NewInt(i), nil
		}
		if f, err := strconv.ParseFloat(str, 64); err == nil {
			return values.NewFloat(math.Abs(f)), nil
		}
	}
	return values.NewInt(0), nil
}

func maxHandler(ctx ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("max() expects at least 1 parameter, 0 given")
	}

	max := args[0]
	for i := 1; i < len(args); i++ {
		if compareValues(args[i], max) > 0 {
			max = args[i]
		}
	}

	return max, nil
}

func minHandler(ctx ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("min() expects at least 1 parameter, 0 given")
	}

	min := args[0]
	for i := 1; i < len(args); i++ {
		if compareValues(args[i], min) < 0 {
			min = args[i]
		}
	}

	return min, nil
}

func varDumpHandler(ctx ExecutionContext, args []*values.Value) (*values.Value, error) {
	for _, arg := range args {
		ctx.WriteOutput(formatVarDump(arg, 0) + "\n")
	}
	return values.NewNull(), nil
}

func printRHandler(ctx ExecutionContext, args []*values.Value) (*values.Value, error) {
	output := formatPrintR(args[0], 0)
	if len(args) > 1 && args[1].ToBool() {
		return values.NewString(output), nil
	}
	fmt.Print(output)
	return values.NewBool(true), nil
}

// Helper functions
func compareValues(a, b *values.Value) int {
	if a.IsInt() && b.IsInt() {
		ai, bi := a.ToInt(), b.ToInt()
		if ai < bi {
			return -1
		} else if ai > bi {
			return 1
		}
		return 0
	}

	if (a.IsInt() || a.IsFloat()) && (b.IsInt() || b.IsFloat()) {
		af, bf := a.ToFloat(), b.ToFloat()
		if af < bf {
			return -1
		} else if af > bf {
			return 1
		}
		return 0
	}

	// String comparison
	as, bs := a.ToString(), b.ToString()
	return strings.Compare(as, bs)
}

func formatVarDump(value *values.Value, indent int) string {
	prefix := strings.Repeat("  ", indent)

	switch value.Type {
	case values.TypeString:
		return fmt.Sprintf("%sstring(%d) \"%s\"", prefix, len(value.ToString()), value.ToString())
	case values.TypeInt:
		return fmt.Sprintf("%sint(%d)", prefix, value.ToInt())
	case values.TypeFloat:
		return fmt.Sprintf("%sfloat(%g)", prefix, value.ToFloat())
	case values.TypeBool:
		if value.ToBool() {
			return fmt.Sprintf("%sbool(true)", prefix)
		}
		return fmt.Sprintf("%sbool(false)", prefix)
	case values.TypeNull:
		return fmt.Sprintf("%sNULL", prefix)
	case values.TypeArray:
		result := fmt.Sprintf("%sarray(%d) {\n", prefix, value.ArrayCount())
		if value.ArrayCount() > 0 {
			// Access the array data directly
			if arr, ok := value.Data.(*values.Array); ok {
				for key, val := range arr.Elements {
					keyStr := fmt.Sprintf("  [%v]=>", key)
					result += fmt.Sprintf("%s%s\n", prefix, keyStr)
					result += formatVarDump(val, indent+1) + "\n"
				}
			}
		}
		result += prefix + "}"
		return result
	default:
		return fmt.Sprintf("%s%s", prefix, value.ToString())
	}
}

func formatPrintR(value *values.Value, indent int) string {
	prefix := strings.Repeat("    ", indent)

	switch value.Type {
	case values.TypeArray:
		if value.ArrayCount() == 0 {
			return "Array\n(\n)\n"
		}
		result := "Array\n" + prefix + "(\n"
		// Simple array formatting - would need more complex logic for full implementation
		result += prefix + ")\n"
		return result
	default:
		return value.ToString() + "\n"
	}
}

// strRepeatHandler implements the str_repeat function
func strRepeatHandler(ctx ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) != 2 {
		return values.NewNull(), fmt.Errorf("str_repeat() expects exactly 2 parameters, %d given", len(args))
	}

	input := args[0].ToString()
	multiplier := int(args[1].ToInt())

	if multiplier < 0 {
		return values.NewString(""), nil
	}

	result := strings.Repeat(input, multiplier)
	return values.NewString(result), nil
}

// functionExistsHandler implements the function_exists function
func functionExistsHandler(ctx ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) != 1 {
		return values.NewNull(), fmt.Errorf("function_exists() expects exactly 1 parameter, %d given", len(args))
	}

	functionName := args[0].ToString()

	// Check if the function exists (both built-in and user-defined)
	exists := ctx.HasFunction(functionName)

	return values.NewBool(exists), nil
}

// classExistsHandler implements the class_exists function
func classExistsHandler(ctx ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) != 1 {
		return values.NewNull(), fmt.Errorf("class_exists() expects exactly 1 parameter, %d given", len(args))
	}

	className := args[0].ToString()

	// Check if the class exists (both built-in and user-defined)
	exists := ctx.HasClass(className)

	return values.NewBool(exists), nil
}

// methodExistsHandler implements the method_exists function
func methodExistsHandler(ctx ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) != 2 {
		return values.NewNull(), fmt.Errorf("method_exists() expects exactly 2 parameters, %d given", len(args))
	}

	objectOrClass := args[0]
	methodName := args[1].ToString()

	var className string

	// Determine the class name based on the first argument
	if objectOrClass.IsString() {
		// First argument is a class name string
		className = objectOrClass.ToString()
	} else if objectOrClass.IsObject() {
		// First argument is an object - get its class name
		if obj, ok := objectOrClass.Data.(*values.Object); ok {
			className = obj.ClassName
		} else {
			return values.NewBool(false), nil
		}
	} else {
		// Invalid argument type
		return values.NewBool(false), nil
	}

	// Check if the class exists
	if !ctx.HasClass(className) {
		return values.NewBool(false), nil
	}

	// Use the execution context to check for the method across both registries
	methodExists := ctx.HasMethod(className, methodName)
	return values.NewBool(methodExists), nil
}

// goHandler implements the go() function for running closures in goroutines
// Signature: go($closure, $var1, $var2, ...) - variables are passed as additional arguments
func goHandler(ctx ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewNull(), fmt.Errorf("go() expects at least 1 parameter, %d given", len(args))
	}

	if !args[0].IsCallable() {
		return values.NewNull(), fmt.Errorf("go() expects a callable as first argument")
	}

	closure := args[0].ClosureGet()
	if closure == nil {
		return values.NewNull(), fmt.Errorf("invalid closure data")
	}

	// Create a map of captured variables from the additional arguments
	capturedVars := make(map[string]*values.Value)

	// If additional arguments are provided, store them as captured variables
	// The variables will be accessible in the goroutine context
	for i := 1; i < len(args); i++ {
		// Use generic names for the captured variables: var_0, var_1, etc.
		varName := fmt.Sprintf("var_%d", i-1)
		capturedVars[varName] = args[i]
	}

	// Also include any existing bound variables from the closure
	for name, value := range closure.BoundVars {
		capturedVars[name] = value
	}

	goroutine := values.NewGoroutine(closure, capturedVars)

	// Start the goroutine and actually execute the closure
	go func() {
		defer func() {
			if r := recover(); r != nil {
				gor := goroutine.Data.(*values.Goroutine)
				gor.Status = "error"
				if err, ok := r.(error); ok {
					gor.Error = err
				} else {
					gor.Error = fmt.Errorf("panic: %v", r)
				}
				close(gor.Done)
			}
		}()

		// Get the goroutine data
		gor := goroutine.Data.(*values.Goroutine)

		// Create a VM instance to execute the closure
		vm := newVirtualMachineForGoroutine()

		// Create a new execution context for the goroutine
		goroutineCtx := newExecutionContextForGoroutine()

		// Copy global variables from the parent context if available
		// For now, we'll copy from registry since we don't have direct access to VM context
		// In a full implementation, this would copy from the parent execution context

		// Add captured variables to the goroutine context
		for name, value := range gor.UseVars {
			// Store captured variables in the global scope of the goroutine
			goroutineCtx.GlobalVars[name] = value
		}

		// Execute the closure using the VM
		result, err := vm.ExecuteClosure(goroutineCtx, gor.Function, []*values.Value{})

		// Update goroutine status
		if err != nil {
			gor.Status = "error"
			gor.Error = err
		} else {
			gor.Status = "completed"
			gor.Result = result
		}

		// Signal completion
		close(gor.Done)
	}()

	return goroutine, nil
}

// Helper function to create a VM instance for goroutine execution
func newVirtualMachineForGoroutine() VMExecutor {
	// Import the vm package dynamically to avoid circular imports
	// This is a simple interface approach
	if vmFactory != nil {
		return vmFactory()
	}
	// Fallback: return a mock VM that doesn't actually execute
	return &mockVM{}
}

// Helper function to create an execution context for goroutine
func newExecutionContextForGoroutine() *GoroutineContext {
	if ctxFactory != nil {
		return ctxFactory()
	}
	// Fallback: return a basic execution context
	return &GoroutineContext{
		GlobalVars:      make(map[string]*values.Value),
		GlobalConstants: make(map[string]*values.Value),
		Functions:       make(map[string]*VMFunction),
		Variables:       make(map[uint32]*values.Value),
		Temporaries:     make(map[uint32]*values.Value),
	}
}

// Factory functions for VM and ExecutionContext - these will be set by the VM package
var (
	vmFactory  func() VMExecutor
	ctxFactory func() *GoroutineContext
)

// SetVMFactory sets the VM factory function (called by VM package init)
func SetVMFactory(factory func() VMExecutor) {
	vmFactory = factory
}

// SetContextFactory sets the GoroutineContext factory function (called by VM package init)
func SetContextFactory(factory func() *GoroutineContext) {
	ctxFactory = factory
}

// VMExecutor interface to avoid circular imports
type VMExecutor interface {
	ExecuteClosure(ctx *GoroutineContext, closure *values.Closure, args []*values.Value) (*values.Value, error)
}

// GoroutineContext represents the VM execution context for goroutines (simplified for interface)
type GoroutineContext struct {
	GlobalVars      map[string]*values.Value
	GlobalConstants map[string]*values.Value
	Functions       map[string]*VMFunction
	Variables       map[uint32]*values.Value
	Temporaries     map[uint32]*values.Value
}

// VMFunction represents a VM function (simplified for interface)
type VMFunction struct {
	Name         string
	Instructions []interface{} // Simplified
	Constants    []*values.Value
}

// Mock VM implementation for cases where real VM is not available
type mockVM struct{}

func (m *mockVM) ExecuteClosure(ctx *GoroutineContext, closure *values.Closure, args []*values.Value) (*values.Value, error) {
	// Simple mock execution - just return a success value
	return values.NewString("mock execution completed"), nil
}

// WaitGroup method handlers
func waitGroupConstructHandler(ctx ExecutionContext, args []*values.Value) (*values.Value, error) {
	return values.NewWaitGroup(), nil
}

func waitGroupAddHandler(ctx ExecutionContext, args []*values.Value) (*values.Value, error) {
	// This should be called on a WaitGroup instance
	// For now, return error - proper implementation would need object context
	return values.NewNull(), fmt.Errorf("WaitGroup.Add() method handler not fully implemented")
}

func waitGroupDoneHandler(ctx ExecutionContext, args []*values.Value) (*values.Value, error) {
	// This should be called on a WaitGroup instance
	return values.NewNull(), fmt.Errorf("WaitGroup.Done() method handler not fully implemented")
}

func waitGroupWaitHandler(ctx ExecutionContext, args []*values.Value) (*values.Value, error) {
	// This should be called on a WaitGroup instance
	return values.NewNull(), fmt.Errorf("WaitGroup.Wait() method handler not fully implemented")
}

func sleepHandler(ctx ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) != 1 {
		return values.NewNull(), fmt.Errorf("sleep() expects exactly 1 parameter, %d given", len(args))
	}

	seconds := int(args[0].ToInt())
	if seconds < 0 {
		return values.NewNull(), fmt.Errorf("sleep() expects a non-negative integer")
	}

	time.Sleep(time.Duration(seconds) * time.Second)
	return values.NewInt(0), nil
}

func timeHandler(ctx ExecutionContext, args []*values.Value) (*values.Value, error) {
	now := time.Now()
	return values.NewInt(now.Unix()), nil
}

func microtimeHandler(ctx ExecutionContext, args []*values.Value) (*values.Value, error) {
	asFloat := false
	if len(args) == 1 {
		asFloat = args[0].ToBool()
	} else if len(args) > 1 {
		return values.NewNull(), fmt.Errorf("microtime() expects at most 1 parameter, %d given", len(args))
	}

	now := time.Now()
	if asFloat {
		return values.NewFloat(float64(now.UnixNano()) / 1e9), nil
	}

	sec := now.Unix()
	usec := now.Nanosecond() / 1000
	return values.NewString(fmt.Sprintf("%d %d", usec, sec)), nil
}
