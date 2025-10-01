package runtime

import (
	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// mysqliResultMethodDescriptors returns method descriptors for mysqli_result class
func mysqliResultMethodDescriptors() map[string]*registry.MethodDescriptor {
	return map[string]*registry.MethodDescriptor{
		"close":             newMySQLiMethod("close", []registry.ParameterDescriptor{}, "void", mysqliResultClose),
		"free":              newMySQLiMethod("free", []registry.ParameterDescriptor{}, "void", mysqliResultFree),
		"free_result":       newMySQLiMethod("free_result", []registry.ParameterDescriptor{}, "void", mysqliResultFreeResult),
		"data_seek":         newMySQLiMethod("data_seek", []registry.ParameterDescriptor{{Name: "offset", Type: "int"}}, "bool", mysqliResultDataSeek),
		"fetch_all":         newMySQLiMethod("fetch_all", []registry.ParameterDescriptor{{Name: "mode", Type: "int", HasDefault: true, DefaultValue: values.NewInt(1)}}, "array", mysqliResultFetchAll),
		"fetch_array":       newMySQLiMethod("fetch_array", []registry.ParameterDescriptor{{Name: "mode", Type: "int", HasDefault: true, DefaultValue: values.NewInt(3)}}, "array|null", mysqliResultFetchArray),
		"fetch_assoc":       newMySQLiMethod("fetch_assoc", []registry.ParameterDescriptor{}, "array|null", mysqliResultFetchAssoc),
		"fetch_row":         newMySQLiMethod("fetch_row", []registry.ParameterDescriptor{}, "array|null", mysqliResultFetchRow),
		"fetch_object":      newMySQLiMethod("fetch_object", []registry.ParameterDescriptor{{Name: "class", Type: "string", HasDefault: true, DefaultValue: values.NewString("stdClass")}, {Name: "constructor_args", Type: "array", HasDefault: true, DefaultValue: values.NewArray()}}, "object|null", mysqliResultFetchObject),
		"fetch_field":       newMySQLiMethod("fetch_field", []registry.ParameterDescriptor{}, "object|false", mysqliResultFetchField),
		"fetch_field_direct": newMySQLiMethod("fetch_field_direct", []registry.ParameterDescriptor{{Name: "index", Type: "int"}}, "object|false", mysqliResultFetchFieldDirect),
		"fetch_fields":      newMySQLiMethod("fetch_fields", []registry.ParameterDescriptor{}, "array", mysqliResultFetchFields),
		"field_seek":        newMySQLiMethod("field_seek", []registry.ParameterDescriptor{{Name: "index", Type: "int"}}, "bool", mysqliResultFieldSeek),
	}
}

// mysqliResultPropertyDescriptors returns property descriptors for mysqli_result class
func mysqliResultPropertyDescriptors() map[string]*registry.PropertyDescriptor {
	return map[string]*registry.PropertyDescriptor{
		"current_field": {
			Name:         "current_field",
			Visibility:   "public",
			IsStatic:     false,
			DefaultValue: values.NewInt(0),
		},
		"field_count": {
			Name:         "field_count",
			Visibility:   "public",
			IsStatic:     false,
			DefaultValue: values.NewInt(0),
		},
		"lengths": {
			Name:         "lengths",
			Visibility:   "public",
			IsStatic:     false,
			DefaultValue: values.NewNull(),
		},
		"num_rows": {
			Name:         "num_rows",
			Visibility:   "public",
			IsStatic:     false,
			DefaultValue: values.NewInt(0),
		},
	}
}

// close(), free(), free_result() - All three are aliases for freeing result memory
func mysqliResultClose(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) == 0 {
		return values.NewNull(), nil
	}

	// Extract result from $this
	_, ok := extractMySQLiResult(args[0])
	if !ok {
		return values.NewNull(), nil
	}

	// Stub: Free result memory (no-op in our implementation)
	return values.NewNull(), nil
}

func mysqliResultFree(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	return mysqliResultClose(ctx, args)
}

func mysqliResultFreeResult(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	return mysqliResultClose(ctx, args)
}

// data_seek() - Seek to row offset
func mysqliResultDataSeek(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 2 {
		return values.NewBool(false), nil
	}

	result, ok := extractMySQLiResult(args[0])
	if !ok {
		return values.NewBool(false), nil
	}

	offset := int(args[1].ToInt())

	// Call existing procedural function logic
	if offset >= 0 && offset < len(result.Rows) {
		result.CurrentRow = offset
		return values.NewBool(true), nil
	}

	return values.NewBool(false), nil
}

// fetch_all() - Fetch all rows
func mysqliResultFetchAll(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) == 0 {
		return values.NewNull(), nil
	}

	result, ok := extractMySQLiResult(args[0])
	if !ok {
		return values.NewNull(), nil
	}

	// Create resource wrapper for procedural function
	resultResource := values.NewResource(result)

	// Prepare args for procedural function
	procArgs := []*values.Value{resultResource}
	if len(args) > 1 {
		procArgs = append(procArgs, args[1])
	} else {
		procArgs = append(procArgs, values.NewInt(1)) // MYSQLI_NUM
	}

	// Call existing procedural function from mysqli_functions.go
	// Note: This is a stub implementation, would need full implementation
	return values.NewArray(), nil
}

// fetch_array() - Fetch row as array
func mysqliResultFetchArray(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) == 0 {
		return values.NewNull(), nil
	}

	result, ok := extractMySQLiResult(args[0])
	if !ok {
		return values.NewNull(), nil
	}

	if result.CurrentRow >= len(result.Rows) {
		return values.NewNull(), nil
	}

	mode := int64(3) // MYSQLI_BOTH (default)
	if len(args) > 1 && args[1] != nil {
		mode = args[1].ToInt()
	}

	row := result.Rows[result.CurrentRow]
	result.CurrentRow++

	// Mode 1: MYSQLI_ASSOC (associative array only)
	// Mode 2: MYSQLI_NUM (numeric array only)
	// Mode 3: MYSQLI_BOTH (both associative and numeric)
	arr := values.NewArray()
	arrData := arr.Data.(*values.Array)

	if mode == 1 {
		// Associative only
		for key, val := range row {
			arrData.Elements[key] = val
		}
		arrData.IsIndexed = false
		return arr, nil
	} else if mode == 2 {
		// Numeric array
		idx := int64(0)
		for _, val := range row {
			arrData.Elements[idx] = val
			idx++
		}
		arrData.NextIndex = idx
		arrData.IsIndexed = true
		return arr, nil
	} else {
		// Both
		idx := int64(0)
		for key, val := range row {
			arrData.Elements[key] = val
			arrData.Elements[idx] = val
			idx++
		}
		arrData.NextIndex = idx
		arrData.IsIndexed = false
		return arr, nil
	}
}

// fetch_assoc() - Fetch row as associative array
func mysqliResultFetchAssoc(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) == 0 {
		return values.NewNull(), nil
	}

	result, ok := extractMySQLiResult(args[0])
	if !ok {
		return values.NewNull(), nil
	}

	if result.CurrentRow >= len(result.Rows) {
		return values.NewNull(), nil
	}

	row := result.Rows[result.CurrentRow]
	result.CurrentRow++

	// Convert map[string]*values.Value to proper Array
	arr := values.NewArray()
	arrData := arr.Data.(*values.Array)
	for key, val := range row {
		arrData.Elements[key] = val
	}
	arrData.IsIndexed = false // associative array

	return arr, nil
}

// fetch_row() - Fetch row as numeric array
func mysqliResultFetchRow(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) == 0 {
		return values.NewNull(), nil
	}

	result, ok := extractMySQLiResult(args[0])
	if !ok {
		return values.NewNull(), nil
	}

	if result.CurrentRow >= len(result.Rows) {
		return values.NewNull(), nil
	}

	row := result.Rows[result.CurrentRow]
	result.CurrentRow++

	// Convert to numeric array
	arr := values.NewArray()
	arrData := arr.Data.(*values.Array)
	idx := int64(0)
	for _, val := range row {
		arrData.Elements[idx] = val
		idx++
	}
	arrData.NextIndex = idx
	arrData.IsIndexed = true

	return arr, nil
}

// fetch_object() - Fetch row as object
func mysqliResultFetchObject(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) == 0 {
		return values.NewNull(), nil
	}

	result, ok := extractMySQLiResult(args[0])
	if !ok {
		return values.NewNull(), nil
	}

	if result.CurrentRow >= len(result.Rows) {
		return values.NewNull(), nil
	}

	row := result.Rows[result.CurrentRow]
	result.CurrentRow++

	// Create stdClass object with row data as properties
	obj := values.NewObject("stdClass")
	objData := obj.Data.(*values.Object)
	for key, val := range row {
		objData.Properties[key] = val
	}

	return obj, nil
}

// fetch_field() - Fetch field metadata
func mysqliResultFetchField(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) == 0 {
		return values.NewBool(false), nil
	}

	result, ok := extractMySQLiResult(args[0])
	if !ok {
		return values.NewBool(false), nil
	}

	// Stub: Return false (not implemented)
	// Would need to track current field position and return field metadata
	_ = result
	return values.NewBool(false), nil
}

// fetch_field_direct() - Fetch specific field metadata
func mysqliResultFetchFieldDirect(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 2 {
		return values.NewBool(false), nil
	}

	result, ok := extractMySQLiResult(args[0])
	if !ok {
		return values.NewBool(false), nil
	}

	// Stub: Return false (not implemented)
	_ = result
	return values.NewBool(false), nil
}

// fetch_fields() - Fetch all fields metadata
func mysqliResultFetchFields(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) == 0 {
		return values.NewArray(), nil
	}

	result, ok := extractMySQLiResult(args[0])
	if !ok {
		return values.NewArray(), nil
	}

	// Stub: Return empty array
	_ = result
	return values.NewArray(), nil
}

// field_seek() - Seek to field
func mysqliResultFieldSeek(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 2 {
		return values.NewBool(false), nil
	}

	result, ok := extractMySQLiResult(args[0])
	if !ok {
		return values.NewBool(false), nil
	}

	// Stub: Always return true (field position tracking not implemented)
	_ = result
	return values.NewBool(true), nil
}
