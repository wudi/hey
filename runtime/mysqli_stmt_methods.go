package runtime

import (
	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// mysqliStmtPropertyDescriptors returns property descriptors for mysqli_stmt class
func mysqliStmtPropertyDescriptors() map[string]*registry.PropertyDescriptor {
	return map[string]*registry.PropertyDescriptor{
		"affected_rows": {
			Name:         "affected_rows",
			Visibility:   "public",
			IsStatic:     false,
			DefaultValue: values.NewInt(0),
		},
		"errno": {
			Name:         "errno",
			Visibility:   "public",
			IsStatic:     false,
			DefaultValue: values.NewInt(0),
		},
		"error": {
			Name:         "error",
			Visibility:   "public",
			IsStatic:     false,
			DefaultValue: values.NewString(""),
		},
		"error_list": {
			Name:         "error_list",
			Visibility:   "public",
			IsStatic:     false,
			DefaultValue: values.NewArray(),
		},
		"field_count": {
			Name:         "field_count",
			Visibility:   "public",
			IsStatic:     false,
			DefaultValue: values.NewInt(0),
		},
		"insert_id": {
			Name:         "insert_id",
			Visibility:   "public",
			IsStatic:     false,
			DefaultValue: values.NewInt(0),
		},
		"num_rows": {
			Name:         "num_rows",
			Visibility:   "public",
			IsStatic:     false,
			DefaultValue: values.NewInt(0),
		},
		"param_count": {
			Name:         "param_count",
			Visibility:   "public",
			IsStatic:     false,
			DefaultValue: values.NewInt(0),
		},
		"sqlstate": {
			Name:         "sqlstate",
			Visibility:   "public",
			IsStatic:     false,
			DefaultValue: values.NewString("00000"),
		},
	}
}

// mysqliStmtMethodDescriptors returns method descriptors for mysqli_stmt class
func mysqliStmtMethodDescriptors() map[string]*registry.MethodDescriptor {
	return map[string]*registry.MethodDescriptor{
		// Attribute methods
		"attr_get": newMySQLiMethod("attr_get",
			[]registry.ParameterDescriptor{
				{Name: "attr", Type: "int"},
			},
			"int", mysqliStmtAttrGet),
		"attr_set": newMySQLiMethod("attr_set",
			[]registry.ParameterDescriptor{
				{Name: "attr", Type: "int"},
				{Name: "value", Type: "int"},
			},
			"bool", mysqliStmtAttrSet),

		// Parameter binding methods
		"bind_param": newMySQLiMethod("bind_param",
			[]registry.ParameterDescriptor{
				{Name: "types", Type: "string"},
				// Variadic parameters
			},
			"bool", mysqliStmtBindParam),
		"bind_result": newMySQLiMethod("bind_result",
			[]registry.ParameterDescriptor{
				// Variadic parameters
			},
			"bool", mysqliStmtBindResult),

		// Statement control methods
		"close": newMySQLiMethod("close",
			[]registry.ParameterDescriptor{},
			"bool", mysqliStmtClose),
		"data_seek": newMySQLiMethod("data_seek",
			[]registry.ParameterDescriptor{
				{Name: "offset", Type: "int"},
			},
			"void", mysqliStmtDataSeek),
		"execute": newMySQLiMethod("execute",
			[]registry.ParameterDescriptor{},
			"bool", mysqliStmtExecute),
		"fetch": newMySQLiMethod("fetch",
			[]registry.ParameterDescriptor{},
			"bool", mysqliStmtFetch),
		"free_result": newMySQLiMethod("free_result",
			[]registry.ParameterDescriptor{},
			"void", mysqliStmtFreeResult),
		"get_result": newMySQLiMethod("get_result",
			[]registry.ParameterDescriptor{},
			"object", mysqliStmtGetResult),
		"get_warnings": newMySQLiMethod("get_warnings",
			[]registry.ParameterDescriptor{},
			"object", mysqliStmtGetWarnings),
		"more_results": newMySQLiMethod("more_results",
			[]registry.ParameterDescriptor{},
			"bool", mysqliStmtMoreResults),
		"next_result": newMySQLiMethod("next_result",
			[]registry.ParameterDescriptor{},
			"bool", mysqliStmtNextResult),
		"prepare": newMySQLiMethod("prepare",
			[]registry.ParameterDescriptor{
				{Name: "query", Type: "string"},
			},
			"bool", mysqliStmtPrepare),
		"reset": newMySQLiMethod("reset",
			[]registry.ParameterDescriptor{},
			"bool", mysqliStmtReset),
		"result_metadata": newMySQLiMethod("result_metadata",
			[]registry.ParameterDescriptor{},
			"object", mysqliStmtResultMetadata),
		"send_long_data": newMySQLiMethod("send_long_data",
			[]registry.ParameterDescriptor{
				{Name: "param_num", Type: "int"},
				{Name: "data", Type: "string"},
			},
			"bool", mysqliStmtSendLongData),
		"store_result": newMySQLiMethod("store_result",
			[]registry.ParameterDescriptor{},
			"bool", mysqliStmtStoreResult),
	}
}

// Method implementations - these wrap the procedural functions

func mysqliStmtAttrGet(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 2 {
		return values.NewInt(0), nil
	}

	stmt, ok := extractMySQLiStmt(args[0])
	if !ok {
		return values.NewInt(0), nil
	}

	// Create resource value for procedural function
	stmtResource := values.NewResource(stmt)

	// Call procedural function mysqli_stmt_attr_get
	procArgs := []*values.Value{stmtResource, args[1]}
	for _, fn := range GetMySQLiStmtFunctions() {
		if fn.Name == "mysqli_stmt_attr_get" {
			return fn.Builtin(nil, procArgs)
		}
	}

	return values.NewInt(0), nil
}

func mysqliStmtAttrSet(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 3 {
		return values.NewBool(false), nil
	}

	stmt, ok := extractMySQLiStmt(args[0])
	if !ok {
		return values.NewBool(false), nil
	}

	stmtResource := values.NewResource(stmt)
	procArgs := []*values.Value{stmtResource, args[1], args[2]}

	for _, fn := range GetMySQLiStmtFunctions() {
		if fn.Name == "mysqli_stmt_attr_set" {
			return fn.Builtin(nil, procArgs)
		}
	}

	return values.NewBool(false), nil
}

func mysqliStmtBindParam(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 2 {
		return values.NewBool(false), nil
	}

	stmt, ok := extractMySQLiStmt(args[0])
	if !ok {
		return values.NewBool(false), nil
	}

	stmtResource := values.NewResource(stmt)
	procArgs := make([]*values.Value, len(args))
	procArgs[0] = stmtResource
	copy(procArgs[1:], args[1:])

	for _, fn := range GetMySQLiStmtFunctions() {
		if fn.Name == "mysqli_stmt_bind_param" {
			return fn.Builtin(nil, procArgs)
		}
	}

	return values.NewBool(false), nil
}

func mysqliStmtBindResult(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewBool(false), nil
	}

	stmt, ok := extractMySQLiStmt(args[0])
	if !ok {
		return values.NewBool(false), nil
	}

	stmtResource := values.NewResource(stmt)
	procArgs := make([]*values.Value, len(args))
	procArgs[0] = stmtResource
	copy(procArgs[1:], args[1:])

	for _, fn := range GetMySQLiStmtFunctions() {
		if fn.Name == "mysqli_stmt_bind_result" {
			return fn.Builtin(nil, procArgs)
		}
	}

	return values.NewBool(false), nil
}

func mysqliStmtClose(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewBool(false), nil
	}

	stmt, ok := extractMySQLiStmt(args[0])
	if !ok {
		return values.NewBool(false), nil
	}

	stmtResource := values.NewResource(stmt)
	procArgs := []*values.Value{stmtResource}

	for _, fn := range GetMySQLiStmtFunctions() {
		if fn.Name == "mysqli_stmt_close" {
			return fn.Builtin(nil, procArgs)
		}
	}

	return values.NewBool(false), nil
}

func mysqliStmtDataSeek(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 2 {
		return values.NewNull(), nil
	}

	stmt, ok := extractMySQLiStmt(args[0])
	if !ok {
		return values.NewNull(), nil
	}

	stmtResource := values.NewResource(stmt)
	procArgs := []*values.Value{stmtResource, args[1]}

	for _, fn := range GetMySQLiStmtFunctions() {
		if fn.Name == "mysqli_stmt_data_seek" {
			return fn.Builtin(nil, procArgs)
		}
	}

	return values.NewNull(), nil
}

func mysqliStmtExecute(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewBool(false), nil
	}

	stmt, ok := extractMySQLiStmt(args[0])
	if !ok {
		return values.NewBool(false), nil
	}

	stmtResource := values.NewResource(stmt)
	procArgs := []*values.Value{stmtResource}

	for _, fn := range GetMySQLiStmtFunctions() {
		if fn.Name == "mysqli_stmt_execute" {
			return fn.Builtin(nil, procArgs)
		}
	}

	return values.NewBool(false), nil
}

func mysqliStmtFetch(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewNull(), nil
	}

	stmt, ok := extractMySQLiStmt(args[0])
	if !ok {
		return values.NewNull(), nil
	}

	stmtResource := values.NewResource(stmt)
	procArgs := []*values.Value{stmtResource}

	for _, fn := range GetMySQLiStmtFunctions() {
		if fn.Name == "mysqli_stmt_fetch" {
			return fn.Builtin(nil, procArgs)
		}
	}

	return values.NewNull(), nil
}

func mysqliStmtFreeResult(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewNull(), nil
	}

	stmt, ok := extractMySQLiStmt(args[0])
	if !ok {
		return values.NewNull(), nil
	}

	stmtResource := values.NewResource(stmt)
	procArgs := []*values.Value{stmtResource}

	for _, fn := range GetMySQLiStmtFunctions() {
		if fn.Name == "mysqli_stmt_free_result" {
			return fn.Builtin(nil, procArgs)
		}
	}

	return values.NewNull(), nil
}

func mysqliStmtGetResult(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewBool(false), nil
	}

	stmt, ok := extractMySQLiStmt(args[0])
	if !ok {
		return values.NewBool(false), nil
	}

	stmtResource := values.NewResource(stmt)
	procArgs := []*values.Value{stmtResource}

	for _, fn := range GetMySQLiStmtFunctions() {
		if fn.Name == "mysqli_stmt_get_result" {
			return fn.Builtin(nil, procArgs)
		}
	}

	return values.NewBool(false), nil
}

func mysqliStmtGetWarnings(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	// Stub: Return false (warnings not implemented)
	return values.NewBool(false), nil
}

func mysqliStmtMoreResults(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewBool(false), nil
	}

	stmt, ok := extractMySQLiStmt(args[0])
	if !ok {
		return values.NewBool(false), nil
	}

	stmtResource := values.NewResource(stmt)
	procArgs := []*values.Value{stmtResource}

	for _, fn := range GetMySQLiStmtFunctions() {
		if fn.Name == "mysqli_stmt_more_results" {
			return fn.Builtin(nil, procArgs)
		}
	}

	return values.NewBool(false), nil
}

func mysqliStmtNextResult(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewBool(false), nil
	}

	stmt, ok := extractMySQLiStmt(args[0])
	if !ok {
		return values.NewBool(false), nil
	}

	stmtResource := values.NewResource(stmt)
	procArgs := []*values.Value{stmtResource}

	for _, fn := range GetMySQLiStmtFunctions() {
		if fn.Name == "mysqli_stmt_next_result" {
			return fn.Builtin(nil, procArgs)
		}
	}

	return values.NewBool(false), nil
}

func mysqliStmtPrepare(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 2 {
		return values.NewBool(false), nil
	}

	stmt, ok := extractMySQLiStmt(args[0])
	if !ok {
		return values.NewBool(false), nil
	}

	stmtResource := values.NewResource(stmt)
	procArgs := []*values.Value{stmtResource, args[1]}

	for _, fn := range GetMySQLiStmtFunctions() {
		if fn.Name == "mysqli_stmt_prepare" {
			return fn.Builtin(nil, procArgs)
		}
	}

	return values.NewBool(false), nil
}

func mysqliStmtReset(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewBool(false), nil
	}

	stmt, ok := extractMySQLiStmt(args[0])
	if !ok {
		return values.NewBool(false), nil
	}

	stmtResource := values.NewResource(stmt)
	procArgs := []*values.Value{stmtResource}

	for _, fn := range GetMySQLiStmtFunctions() {
		if fn.Name == "mysqli_stmt_reset" {
			return fn.Builtin(nil, procArgs)
		}
	}

	return values.NewBool(false), nil
}

func mysqliStmtResultMetadata(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewBool(false), nil
	}

	stmt, ok := extractMySQLiStmt(args[0])
	if !ok {
		return values.NewBool(false), nil
	}

	stmtResource := values.NewResource(stmt)
	procArgs := []*values.Value{stmtResource}

	for _, fn := range GetMySQLiStmtFunctions() {
		if fn.Name == "mysqli_stmt_result_metadata" {
			return fn.Builtin(nil, procArgs)
		}
	}

	return values.NewBool(false), nil
}

func mysqliStmtSendLongData(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 3 {
		return values.NewBool(false), nil
	}

	stmt, ok := extractMySQLiStmt(args[0])
	if !ok {
		return values.NewBool(false), nil
	}

	stmtResource := values.NewResource(stmt)
	procArgs := []*values.Value{stmtResource, args[1], args[2]}

	for _, fn := range GetMySQLiStmtFunctions() {
		if fn.Name == "mysqli_stmt_send_long_data" {
			return fn.Builtin(nil, procArgs)
		}
	}

	return values.NewBool(false), nil
}

func mysqliStmtStoreResult(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewBool(false), nil
	}

	stmt, ok := extractMySQLiStmt(args[0])
	if !ok {
		return values.NewBool(false), nil
	}

	stmtResource := values.NewResource(stmt)
	procArgs := []*values.Value{stmtResource}

	for _, fn := range GetMySQLiStmtFunctions() {
		if fn.Name == "mysqli_stmt_store_result" {
			return fn.Builtin(nil, procArgs)
		}
	}

	return values.NewBool(false), nil
}
