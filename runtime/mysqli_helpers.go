package runtime

import (
	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/runtime/spl"
	"github.com/wudi/hey/values"
)

// newMySQLiMethod is a helper to create method descriptors with implicit $this parameter
func newMySQLiMethod(name string, params []registry.ParameterDescriptor, returnType string, handler registry.BuiltinImplementation) *registry.MethodDescriptor {
	// Prepend $this parameter for internal use
	fullParams := make([]registry.ParameterDescriptor, 0, len(params)+1)
	fullParams = append(fullParams, registry.ParameterDescriptor{
		Name: "this",
		Type: "object",
	})
	fullParams = append(fullParams, params...)

	return &registry.MethodDescriptor{
		Name:       name,
		Visibility: "public",
		IsStatic:   false,
		IsAbstract: false,
		IsFinal:    false,
		IsVariadic: false,
		Parameters: convertToMySQLiParamPointers(params),
		Implementation: spl.NewBuiltinMethodImpl(&registry.Function{
			Name:       name,
			IsBuiltin:  true,
			Builtin:    handler,
			Parameters: convertMySQLiParamDescriptors(fullParams),
		}),
	}
}

// convertMySQLiParamDescriptors converts ParameterDescriptors to Parameters
func convertMySQLiParamDescriptors(params []registry.ParameterDescriptor) []*registry.Parameter {
	result := make([]*registry.Parameter, len(params))
	for i, p := range params {
		result[i] = &registry.Parameter{
			Name:         p.Name,
			Type:         p.Type,
			IsReference:  p.IsReference,
			HasDefault:   p.HasDefault,
			DefaultValue: p.DefaultValue,
		}
	}
	return result
}

// convertToMySQLiParamPointers converts slice of ParameterDescriptor to pointers
func convertToMySQLiParamPointers(params []registry.ParameterDescriptor) []*registry.ParameterDescriptor {
	result := make([]*registry.ParameterDescriptor, len(params))
	for i := range params {
		result[i] = &params[i]
	}
	return result
}

// extractMySQLiConnection extracts MySQLiConnection from $this object or resource
func extractMySQLiConnection(thisObj *values.Value) (*MySQLiConnection, bool) {
	if thisObj == nil {
		return nil, false
	}

	// Handle direct resource (procedural style)
	if thisObj.Type == values.TypeResource {
		conn, ok := thisObj.Data.(*MySQLiConnection)
		return conn, ok
	}

	// Handle object wrapper (OOP style)
	if thisObj.Type != values.TypeObject {
		return nil, false
	}

	obj, ok := thisObj.Data.(*values.Object)
	if !ok {
		return nil, false
	}

	// Look for __mysqli_connection property
	connVal, ok := obj.Properties["__mysqli_connection"]
	if !ok || connVal.Type != values.TypeResource {
		return nil, false
	}

	conn, ok := connVal.Data.(*MySQLiConnection)
	return conn, ok
}

// extractMySQLiResult extracts MySQLiResult from $this object
func extractMySQLiResult(thisObj *values.Value) (*MySQLiResult, bool) {
	if thisObj == nil || thisObj.Type != values.TypeObject {
		return nil, false
	}

	obj, ok := thisObj.Data.(*values.Object)
	if !ok {
		return nil, false
	}

	// Look for __mysqli_result property
	resVal, ok := obj.Properties["__mysqli_result"]
	if !ok || resVal.Type != values.TypeResource {
		return nil, false
	}

	result, ok := resVal.Data.(*MySQLiResult)
	return result, ok
}

// extractMySQLiStmt extracts MySQLiStmt from $this object or resource
func extractMySQLiStmt(thisObj *values.Value) (*MySQLiStmt, bool) {
	if thisObj == nil {
		return nil, false
	}

	// Handle direct resource (procedural style)
	if thisObj.Type == values.TypeResource {
		stmt, ok := thisObj.Data.(*MySQLiStmt)
		return stmt, ok
	}

	// Handle object wrapper (OOP style)
	if thisObj.Type != values.TypeObject {
		return nil, false
	}

	obj, ok := thisObj.Data.(*values.Object)
	if !ok {
		return nil, false
	}

	// Look for __mysqli_stmt property
	stmtVal, ok := obj.Properties["__mysqli_stmt"]
	if !ok || stmtVal.Type != values.TypeResource {
		return nil, false
	}

	stmt, ok := stmtVal.Data.(*MySQLiStmt)
	return stmt, ok
}

// createMySQLiResultObject creates a mysqli_result object wrapping a MySQLiResult
func createMySQLiResultObject(ctx registry.BuiltinCallContext, result *MySQLiResult) *values.Value {
	// Create mysqli_result object
	obj := &values.Object{
		ClassName:  "mysqli_result",
		Properties: make(map[string]*values.Value),
	}

	// Store the result resource inside the object
	obj.Properties["__mysqli_result"] = values.NewResource(result)

	// Set property values from result
	obj.Properties["current_field"] = values.NewInt(0)
	obj.Properties["field_count"] = values.NewInt(int64(result.FieldCount))
	obj.Properties["lengths"] = values.NewNull()
	obj.Properties["num_rows"] = values.NewInt(result.NumRows)

	return &values.Value{
		Type: values.TypeObject,
		Data: obj,
	}
}

// createMySQLiStmtObject creates a mysqli_stmt object wrapping a MySQLiStmt
func createMySQLiStmtObject(ctx registry.BuiltinCallContext, stmt *MySQLiStmt) *values.Value {
	obj := &values.Object{
		ClassName:  "mysqli_stmt",
		Properties: make(map[string]*values.Value),
	}

	obj.Properties["__mysqli_stmt"] = values.NewResource(stmt)
	obj.Properties["affected_rows"] = values.NewInt(stmt.AffectedRows)
	obj.Properties["errno"] = values.NewInt(int64(stmt.ErrorNo))
	obj.Properties["error"] = values.NewString(stmt.Error)
	obj.Properties["error_list"] = values.NewArray()
	obj.Properties["field_count"] = values.NewInt(int64(stmt.FieldCount))
	obj.Properties["insert_id"] = values.NewInt(stmt.InsertID)
	obj.Properties["num_rows"] = values.NewInt(0)
	obj.Properties["param_count"] = values.NewInt(int64(stmt.ParamCount))
	obj.Properties["sqlstate"] = values.NewString(stmt.SQLState)

	return &values.Value{
		Type: values.TypeObject,
		Data: obj,
	}
}
