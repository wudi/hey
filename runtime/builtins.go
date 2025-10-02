package runtime

import (
	"math"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/runtime/spl"
	"github.com/wudi/hey/values"
)

var builtinClassStubs = map[string]map[string]struct{}{
	"stdclass": {},
	"exception": {
		"getmessage":       {},
		"getcode":          {},
		"getfile":          {},
		"getline":          {},
		"gettrace":         {},
		"gettraceasstring": {},
	},
}

// GetAllBuiltinFunctions returns all builtin functions from all modules
func GetAllBuiltinFunctions() []*registry.Function {
	var functions []*registry.Function

	// Add functions from each module
	functions = append(functions, GetArrayFunctions()...)
	functions = append(functions, GetStringFunctions()...)
	functions = append(functions, GetRegexFunctions()...)
	functions = append(functions, GetRegexCacheFunctions()...)
	functions = append(functions, GetTypeFunctions()...)
	functions = append(functions, GetEncodingFunctions()...)
	functions = append(functions, GetFilesystemFunctions()...)
	functions = append(functions, GetSystemFunctions()...)
	functions = append(functions, GetTimeFunctions()...)
	functions = append(functions, GetDateTimeFunctions()...)
	functions = append(functions, GetDateTimeObjectFunctions()...)
	functions = append(functions, GetMathFunctions()...)
	functions = append(functions, GetOutputFunctions()...)
	functions = append(functions, GetReflectionFunctions()...)
	functions = append(functions, GetVariableFunctions()...)
	functions = append(functions, GetConcurrencyFunctions()...)
	functions = append(functions, GetIniFunctions()...)
	functions = append(functions, GetCtypeFunctions()...)
	functions = append(functions, GetErrorFunctions()...)
	functions = append(functions, GetFunctionFunctions()...)
	functions = append(functions, GetWindowsFunctions()...)
	functions = append(functions, GetAssertFunctions()...)
	functions = append(functions, GetHTTPFunctions()...)
	functions = append(functions, GetMySQLiFunctions()...)
	functions = append(functions, GetMySQLiAdvancedFunctions()...)
	functions = append(functions, GetMySQLiStmtFunctions()...)

	return functions
}

// GetAllBuiltinClasses returns all builtin classes from all modules
func GetAllBuiltinClasses() []*registry.ClassDescriptor {
	var classes []*registry.ClassDescriptor

	// Add stdClass - PHP's generic object class
	stdClass := &registry.ClassDescriptor{
		Name:       "stdClass",
		Parent:     "",
		Interfaces: []string{},
		Traits:     []string{},
		Methods:    make(map[string]*registry.MethodDescriptor),
		Properties: make(map[string]*registry.PropertyDescriptor),
		Constants:  make(map[string]*registry.ConstantDescriptor),
		IsAbstract: false,
		IsFinal:    false,
	}
	classes = append(classes, stdClass)

	// Add classes from exception module
	classes = append(classes, GetClasses()...)

	// Add classes from iterator module
	classes = append(classes, GetIteratorClasses()...)

	// Add classes from concurrency module
	classes = append(classes, GetConcurrencyClasses()...)

	// Add classes from SPL module
	classes = append(classes, spl.GetSplClasses()...)

	// Add MySQLi classes
	classes = append(classes, GetMySQLiClasses()...)

	// Add PDO classes
	classes = append(classes, GetPDOClassDescriptors()...)

	return classes
}

// GetAllBuiltinInterfaces returns all builtin interfaces from all modules
func GetAllBuiltinInterfaces() []*registry.Interface {
	var interfaces []*registry.Interface

	// Add interfaces from iterator module
	interfaces = append(interfaces, GetInterfaces()...)

	// Add interfaces from SPL module
	interfaces = append(interfaces, spl.GetSplInterfaces()...)

	return interfaces
}

// GetAllBuiltinConstants returns all builtin constants
func GetAllBuiltinConstants() []*registry.ConstantDescriptor {
	constants := []*registry.ConstantDescriptor{
		{
			Name:  "CASE_LOWER",
			Value: values.NewInt(0),
		},
		{
			Name:  "CASE_UPPER",
			Value: values.NewInt(1),
		},
		{
			Name:  "SORT_REGULAR",
			Value: values.NewInt(0),
		},
		{
			Name:  "SORT_NUMERIC",
			Value: values.NewInt(1),
		},
		{
			Name:  "SORT_STRING",
			Value: values.NewInt(2),
		},
		{
			Name:  "SORT_DESC",
			Value: values.NewInt(3),
		},
		{
			Name:  "SORT_ASC",
			Value: values.NewInt(4),
		},
		{
			Name:  "SORT_LOCALE_STRING",
			Value: values.NewInt(5),
		},
		{
			Name:  "SORT_NATURAL",
			Value: values.NewInt(6),
		},
		{
			Name:  "SORT_FLAG_CASE",
			Value: values.NewInt(8),
		},
		// Error handling constants
		{
			Name:  "E_ERROR",
			Value: values.NewInt(1),
		},
		{
			Name:  "E_WARNING",
			Value: values.NewInt(2),
		},
		{
			Name:  "E_PARSE",
			Value: values.NewInt(4),
		},
		{
			Name:  "E_NOTICE",
			Value: values.NewInt(8),
		},
		{
			Name:  "E_CORE_ERROR",
			Value: values.NewInt(16),
		},
		{
			Name:  "E_CORE_WARNING",
			Value: values.NewInt(32),
		},
		{
			Name:  "E_COMPILE_ERROR",
			Value: values.NewInt(64),
		},
		{
			Name:  "E_COMPILE_WARNING",
			Value: values.NewInt(128),
		},
		{
			Name:  "E_USER_ERROR",
			Value: values.NewInt(256),
		},
		{
			Name:  "E_USER_WARNING",
			Value: values.NewInt(512),
		},
		{
			Name:  "E_USER_NOTICE",
			Value: values.NewInt(1024),
		},
		{
			Name:  "E_STRICT",
			Value: values.NewInt(2048),
		},
		{
			Name:  "E_RECOVERABLE_ERROR",
			Value: values.NewInt(4096),
		},
		{
			Name:  "E_DEPRECATED",
			Value: values.NewInt(8192),
		},
		{
			Name:  "E_USER_DEPRECATED",
			Value: values.NewInt(16384),
		},
		{
			Name:  "E_ALL",
			Value: values.NewInt(30719),
		},

		// Filesystem constants
		// File seek constants
		{
			Name:  "SEEK_SET",
			Value: values.NewInt(0),
		},
		{
			Name:  "SEEK_CUR",
			Value: values.NewInt(1),
		},
		{
			Name:  "SEEK_END",
			Value: values.NewInt(2),
		},

		// File lock constants
		{
			Name:  "LOCK_SH",
			Value: values.NewInt(1),
		},
		{
			Name:  "LOCK_EX",
			Value: values.NewInt(2),
		},
		{
			Name:  "LOCK_UN",
			Value: values.NewInt(3),
		},
		{
			Name:  "LOCK_NB",
			Value: values.NewInt(4),
		},

		// File() function flags
		{
			Name:  "FILE_USE_INCLUDE_PATH",
			Value: values.NewInt(1),
		},
		{
			Name:  "FILE_NO_DEFAULT_CONTEXT",
			Value: values.NewInt(16),
		},
		{
			Name:  "FILE_APPEND",
			Value: values.NewInt(8),
		},
		{
			Name:  "FILE_IGNORE_NEW_LINES",
			Value: values.NewInt(2),
		},
		{
			Name:  "FILE_SKIP_EMPTY_LINES",
			Value: values.NewInt(4),
		},
		{
			Name:  "FILE_BINARY",
			Value: values.NewInt(0),
		},
		{
			Name:  "FILE_TEXT",
			Value: values.NewInt(0),
		},

		// Pathinfo constants
		{
			Name:  "PATHINFO_ALL",
			Value: values.NewInt(15), // 1+2+4+8
		},
		{
			Name:  "PATHINFO_DIRNAME",
			Value: values.NewInt(1),
		},
		{
			Name:  "PATHINFO_BASENAME",
			Value: values.NewInt(2),
		},
		{
			Name:  "PATHINFO_EXTENSION",
			Value: values.NewInt(4),
		},
		{
			Name:  "PATHINFO_FILENAME",
			Value: values.NewInt(8),
		},

		// Glob constants
		{
			Name:  "GLOB_AVAILABLE_FLAGS",
			Value: values.NewInt(0),
		},
		{
			Name:  "GLOB_BRACE",
			Value: values.NewInt(1024),
		},
		{
			Name:  "GLOB_ERR",
			Value: values.NewInt(1),
		},
		{
			Name:  "GLOB_MARK",
			Value: values.NewInt(2),
		},
		{
			Name:  "GLOB_NOCHECK",
			Value: values.NewInt(16),
		},
		{
			Name:  "GLOB_NOESCAPE",
			Value: values.NewInt(64),
		},
		{
			Name:  "GLOB_NOSORT",
			Value: values.NewInt(4),
		},
		{
			Name:  "GLOB_ONLYDIR",
			Value: values.NewInt(8),
		},

		// INI scanner modes
		{
			Name:  "INI_SCANNER_NORMAL",
			Value: values.NewInt(0),
		},
		{
			Name:  "INI_SCANNER_RAW",
			Value: values.NewInt(1),
		},
		{
			Name:  "INI_SCANNER_TYPED",
			Value: values.NewInt(2),
		},

		// fnmatch flags
		{
			Name:  "FNM_NOESCAPE",
			Value: values.NewInt(1),
		},
		{
			Name:  "FNM_PATHNAME",
			Value: values.NewInt(2),
		},
		{
			Name:  "FNM_PERIOD",
			Value: values.NewInt(4),
		},
		{
			Name:  "FNM_CASEFOLD",
			Value: values.NewInt(16),
		},

		// Upload error constants
		{
			Name:  "UPLOAD_ERR_OK",
			Value: values.NewInt(0),
		},
		{
			Name:  "UPLOAD_ERR_INI_SIZE",
			Value: values.NewInt(1),
		},
		{
			Name:  "UPLOAD_ERR_FORM_SIZE",
			Value: values.NewInt(2),
		},
		{
			Name:  "UPLOAD_ERR_PARTIAL",
			Value: values.NewInt(3),
		},
		{
			Name:  "UPLOAD_ERR_NO_FILE",
			Value: values.NewInt(4),
		},
		{
			Name:  "UPLOAD_ERR_NO_TMP_DIR",
			Value: values.NewInt(6),
		},
		{
			Name:  "UPLOAD_ERR_CANT_WRITE",
			Value: values.NewInt(7),
		},
		{
			Name:  "UPLOAD_ERR_EXTENSION",
			Value: values.NewInt(8),
		},

		// Array filter constants
		{
			Name:  "ARRAY_FILTER_USE_KEY",
			Value: values.NewInt(2),
		},
		{
			Name:  "ARRAY_FILTER_USE_BOTH",
			Value: values.NewInt(1),
		},

		// Count constants
		{
			Name:  "COUNT_NORMAL",
			Value: values.NewInt(0),
		},
		{
			Name:  "COUNT_RECURSIVE",
			Value: values.NewInt(1),
		},

		// Extract constants
		{
			Name:  "EXTR_OVERWRITE",
			Value: values.NewInt(0),
		},
		{
			Name:  "EXTR_SKIP",
			Value: values.NewInt(1),
		},
		{
			Name:  "EXTR_PREFIX_SAME",
			Value: values.NewInt(2),
		},
		{
			Name:  "EXTR_PREFIX_ALL",
			Value: values.NewInt(3),
		},
		{
			Name:  "EXTR_PREFIX_INVALID",
			Value: values.NewInt(4),
		},
		{
			Name:  "EXTR_PREFIX_IF_EXISTS",
			Value: values.NewInt(5),
		},
		{
			Name:  "EXTR_IF_EXISTS",
			Value: values.NewInt(6),
		},
		{
			Name:  "EXTR_REFS",
			Value: values.NewInt(256),
		},

		// PHP core constants
		{
			Name:  "PHP_EOL",
			Value: values.NewString("\n"),
		},
		{
			Name:  "PHP_VERSION",
			Value: values.NewString("8.0.30"),
		},
		{
			Name:  "PHP_OS",
			Value: values.NewString("Linux"),
		},
		{
			Name:  "PHP_OS_FAMILY",
			Value: values.NewString("Linux"),
		},
		{
			Name:  "PHP_SAPI",
			Value: values.NewString("cli"),
		},
		{
			Name:  "PHP_BINARY",
			Value: values.NewString("/usr/local/bin/hey"),
		},
		{
			Name:  "DIRECTORY_SEPARATOR",
			Value: values.NewString("/"),
		},
		{
			Name:  "PATH_SEPARATOR",
			Value: values.NewString(":"),
		},

		// Assert constants
		{
			Name:  "ASSERT_ACTIVE",
			Value: values.NewInt(1),
		},
		{
			Name:  "ASSERT_WARNING",
			Value: values.NewInt(4),
		},
		{
			Name:  "ASSERT_BAIL",
			Value: values.NewInt(3),
		},

		// JSON constants
		{
			Name:  "JSON_THROW_ON_ERROR",
			Value: values.NewInt(4194304),
		},

		// Mathematical constants
		{
			Name:  "M_PI",
			Value: values.NewFloat(math.Pi),
		},
		{
			Name:  "M_E",
			Value: values.NewFloat(math.E),
		},
	}

	// Add MySQLi constants
	for _, c := range GetMySQLiConstants() {
		constants = append(constants, &registry.ConstantDescriptor{
			Name:  c.Name,
			Value: c.Value,
		})
	}

	return constants
}

// Legacy variable for backwards compatibility
var builtinFunctionSpecs = GetAllBuiltinFunctions()

// BuiltinMethodImpl represents a builtin method implementation
type BuiltinMethodImpl struct {
	function *registry.Function
}

func NewBuiltinMethodImpl(function *registry.Function) *BuiltinMethodImpl {
	return &BuiltinMethodImpl{function: function}
}

func (b *BuiltinMethodImpl) ImplementationKind() string { return "builtin" }

func (b *BuiltinMethodImpl) GetFunction() *registry.Function {
	return b.function
}
