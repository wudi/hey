package runtime

import (
	"fmt"

	"github.com/wudi/hey/pkg/pdo"
	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// pdoStmtExecute implements $stmt->execute($params)
// args[0] = $this, args[1] = params
func pdoStmtExecute(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	thisObj := args[0]

	// Get statement
	obj := thisObj.Data.(*values.Object)
	stmtVal, hasStmt := obj.Properties["__pdo_stmt"]
	rowsVal, hasRows := obj.Properties["__pdo_rows"]

	// Bind parameters from array if provided
	if len(args) > 1 && args[1].Type == values.TypeArray {
		params := args[1].Data.(*values.Array)
		if hasStmt && stmtVal.Type == values.TypeResource {
			stmt := stmtVal.Data.(pdo.Stmt)

			// Check if we have a parameter map (for named parameters)
			paramMapVal, hasParamMap := obj.Properties["__pdo_param_map"]
			var paramMap map[string]int

			if hasParamMap && paramMapVal.Type == values.TypeArray {
				// Extract parameter map
				paramMap = make(map[string]int)
				paramMapArr := paramMapVal.Data.(*values.Array)
				for keyIface, val := range paramMapArr.Elements {
					if keyStr, ok := keyIface.(string); ok {
						paramPos := int(val.Data.(int64))
						paramMap[keyStr] = paramPos
					}
				}
			}

			// Bind parameters
			// If params.Elements keys are strings (named parameters), use paramMap
			// Otherwise, use positional binding
			idx := 1
			for keyIface, val := range params.Elements {
				switch k := keyIface.(type) {
				case string:
					// Named parameter
					if paramMap != nil {
						if pos, ok := paramMap[k]; ok {
							stmt.BindValue(pos, val, int(pdo.ParamStr))
						}
					}
				case int64:
					// Positional parameter (numeric key)
					stmt.BindValue(int(k)+1, val, int(pdo.ParamStr))
				default:
					// Fallback: treat as positional
					stmt.BindValue(idx, val, int(pdo.ParamStr))
					idx++
				}
			}
		}
	}

	// Initialize error state if not present
	if _, ok := obj.Properties["__pdo_stmt_error_code"]; !ok {
		obj.Properties["__pdo_stmt_error_code"] = values.NewString("00000")
		obj.Properties["__pdo_stmt_error_info"] = values.NewNull()
	}

	// Execute statement
	if hasStmt && stmtVal.Type == values.TypeResource {
		stmt := stmtVal.Data.(pdo.Stmt)

		// Prepared statement execution
		rows, err := stmt.Query()
		if err != nil {
			// Set error state
			sqlState := extractSQLState(err)
			obj.Properties["__pdo_stmt_error_code"] = values.NewString(sqlState)

			errorInfo := values.NewArray()
			errorInfo.ArraySet(values.NewInt(0), values.NewString(sqlState))
			errorInfo.ArraySet(values.NewInt(1), values.NewInt(1))
			errorInfo.ArraySet(values.NewInt(2), values.NewString(err.Error()))
			obj.Properties["__pdo_stmt_error_info"] = errorInfo

			return values.NewBool(false), nil
		}

		// Clear error state on success
		obj.Properties["__pdo_stmt_error_code"] = values.NewString("00000")
		obj.Properties["__pdo_stmt_error_info"] = values.NewNull()

		obj.Properties["__pdo_rows"] = values.NewResource(rows)
		return values.NewBool(true), nil
	}

	// If we already have rows (from query()), just return true
	if hasRows && rowsVal.Type == values.TypeResource {
		return values.NewBool(true), nil
	}

	return values.NewBool(false), nil
}

// pdoStmtFetch implements $stmt->fetch($mode)
// args[0] = $this, args[1] = mode
func pdoStmtFetch(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	thisObj := args[0]

	// Get rows
	obj := thisObj.Data.(*values.Object)
	rowsVal, ok := obj.Properties["__pdo_rows"]
	if !ok || rowsVal.Type != values.TypeResource {
		return values.NewBool(false), nil
	}

	rows := rowsVal.Data.(pdo.Rows)

	// Get fetch mode
	mode := pdo.FetchBoth
	if len(args) > 1 {
		mode = pdo.FetchMode(args[1].Data.(int64))
	}

	// Fetch based on mode
	switch mode {
	case pdo.FetchAssoc:
		row, err := rows.FetchAssoc()
		if err != nil || row == nil {
			return values.NewBool(false), nil
		}
		return convertMapToArray(row), nil

	case pdo.FetchNum:
		row, err := rows.FetchNum()
		if err != nil || row == nil {
			return values.NewBool(false), nil
		}
		return convertSliceToArray(row), nil

	case pdo.FetchBoth:
		assoc, numeric, err := rows.FetchBoth()
		if err != nil || assoc == nil {
			return values.NewBool(false), nil
		}
		return mergeAssocAndNumeric(assoc, numeric), nil

	case pdo.FetchObj:
		row, err := rows.FetchAssoc()
		if err != nil || row == nil {
			return values.NewBool(false), nil
		}
		return convertMapToObject(row), nil

	default:
		return values.NewBool(false), fmt.Errorf("unsupported fetch mode: %d", mode)
	}
}

// pdoStmtFetchAll implements $stmt->fetchAll($mode)
// args[0] = $this, args[1] = mode
func pdoStmtFetchAll(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	thisObj := args[0]

	// Get rows
	obj := thisObj.Data.(*values.Object)
	rowsVal, ok := obj.Properties["__pdo_rows"]
	if !ok || rowsVal.Type != values.TypeResource {
		return values.NewArray(), nil
	}

	rows := rowsVal.Data.(pdo.Rows)

	// Get fetch mode
	mode := pdo.FetchBoth
	if len(args) > 1 {
		mode = pdo.FetchMode(args[1].Data.(int64))
	}

	result := values.NewArray()

	// Fetch all rows
	for {
		var row *values.Value
		var err error

		switch mode {
		case pdo.FetchAssoc:
			rowMap, fetchErr := rows.FetchAssoc()
			if fetchErr != nil || rowMap == nil {
				return result, nil
			}
			row = convertMapToArray(rowMap)

		case pdo.FetchNum:
			rowSlice, fetchErr := rows.FetchNum()
			if fetchErr != nil || rowSlice == nil {
				return result, nil
			}
			row = convertSliceToArray(rowSlice)

		case pdo.FetchBoth:
			assoc, numeric, fetchErr := rows.FetchBoth()
			if fetchErr != nil || assoc == nil {
				return result, nil
			}
			row = mergeAssocAndNumeric(assoc, numeric)

		case pdo.FetchObj:
			rowMap, fetchErr := rows.FetchAssoc()
			if fetchErr != nil || rowMap == nil {
				return result, nil
			}
			row = convertMapToObject(rowMap)

		default:
			return result, fmt.Errorf("unsupported fetch mode: %d", mode)
		}

		if row == nil {
			break
		}

		result.ArraySet(nil, row) // Append to array
		err = rows.Err()
		if err != nil {
			break
		}
	}

	return result, nil
}

// pdoStmtFetchColumn implements $stmt->fetchColumn($column)
// args[0] = $this, args[1] = column
func pdoStmtFetchColumn(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	thisObj := args[0]

	// Get rows
	obj := thisObj.Data.(*values.Object)
	rowsVal, ok := obj.Properties["__pdo_rows"]
	if !ok || rowsVal.Type != values.TypeResource {
		return values.NewBool(false), nil
	}

	rows := rowsVal.Data.(pdo.Rows)

	columnIndex := 0
	if len(args) > 1 {
		columnIndex = int(args[1].Data.(int64))
	}

	// Fetch numeric row
	row, err := rows.FetchNum()
	if err != nil || row == nil {
		return values.NewBool(false), nil
	}

	if columnIndex >= len(row) {
		return values.NewBool(false), nil
	}

	return row[columnIndex], nil
}

// pdoStmtRowCount implements $stmt->rowCount()
// args[0] = $this
func pdoStmtRowCount(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	thisObj := args[0]

	// Get statement
	obj := thisObj.Data.(*values.Object)
	stmtVal, ok := obj.Properties["__pdo_stmt"]
	if !ok || stmtVal.Type != values.TypeResource {
		return values.NewInt(0), nil
	}

	stmt := stmtVal.Data.(pdo.Stmt)
	return values.NewInt(stmt.RowCount()), nil
}

// pdoStmtBindValue implements $stmt->bindValue($param, $value, $type)
// args[0] = $this, args[1] = param, args[2] = value, args[3] = type
func pdoStmtBindValue(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 3 {
		return values.NewBool(false), fmt.Errorf("PDOStatement::bindValue() expects at least 2 parameters")
	}

	thisObj := args[0]

	// Get statement
	obj := thisObj.Data.(*values.Object)
	stmtVal, ok := obj.Properties["__pdo_stmt"]
	if !ok || stmtVal.Type != values.TypeResource {
		return values.NewBool(false), fmt.Errorf("invalid PDOStatement object")
	}

	stmt := stmtVal.Data.(pdo.Stmt)

	// Get parameter identifier (int or string)
	var param interface{}
	if args[1].Type == values.TypeInt {
		param = int(args[1].Data.(int64))
	} else {
		param = args[1].Data.(string)
	}

	value := args[2]

	// Get parameter type
	paramType := int(pdo.ParamStr)
	if len(args) > 3 {
		paramType = int(args[3].Data.(int64))
	}

	// Bind the value
	if err := stmt.BindValue(param, value, paramType); err != nil {
		return values.NewBool(false), nil
	}

	return values.NewBool(true), nil
}

// pdoStmtBindParam implements $stmt->bindParam($param, &$var, $type)
// args[0] = $this, args[1] = param, args[2] = var, args[3] = type
func pdoStmtBindParam(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	// For now, bindParam behaves like bindValue
	// TODO: Implement reference binding properly
	return pdoStmtBindValue(ctx, args)
}

// pdoStmtCloseCursor implements $stmt->closeCursor()
// args[0] = $this
func pdoStmtCloseCursor(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	thisObj := args[0]

	// Get rows
	obj := thisObj.Data.(*values.Object)
	rowsVal, ok := obj.Properties["__pdo_rows"]
	if ok && rowsVal.Type == values.TypeResource {
		rows := rowsVal.Data.(pdo.Rows)
		rows.Close()
		obj.Properties["__pdo_rows"] = values.NewNull()
	}

	return values.NewBool(true), nil
}

// pdoStmtColumnCount implements $stmt->columnCount()
// args[0] = $this
func pdoStmtColumnCount(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	thisObj := args[0]

	// Get rows
	obj := thisObj.Data.(*values.Object)
	rowsVal, ok := obj.Properties["__pdo_rows"]
	if !ok || rowsVal.Type != values.TypeResource {
		return values.NewInt(0), nil
	}

	rows := rowsVal.Data.(pdo.Rows)
	cols, err := rows.Columns()
	if err != nil {
		return values.NewInt(0), nil
	}

	return values.NewInt(int64(len(cols))), nil
}

// pdoStmtErrorCode implements $stmt->errorCode()
// args[0] = $this
func pdoStmtErrorCode(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	thisObj := args[0]
	obj := thisObj.Data.(*values.Object)

	errorCode, ok := obj.Properties["__pdo_stmt_error_code"]
	if !ok || errorCode.Type != values.TypeString {
		return values.NewNull(), nil
	}

	return errorCode, nil
}

// pdoStmtErrorInfo implements $stmt->errorInfo()
// args[0] = $this
func pdoStmtErrorInfo(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	thisObj := args[0]
	obj := thisObj.Data.(*values.Object)

	errorInfo, ok := obj.Properties["__pdo_stmt_error_info"]
	if !ok || errorInfo.Type == values.TypeNull {
		// Return default error info array
		arr := values.NewArray()
		arr.ArraySet(values.NewInt(0), values.NewString("00000"))
		arr.ArraySet(values.NewInt(1), values.NewNull())
		arr.ArraySet(values.NewInt(2), values.NewNull())
		return arr, nil
	}

	return errorInfo, nil
}

// Helper functions for data conversion

func convertMapToArray(m map[string]*values.Value) *values.Value {
	arr := values.NewArray()
	for key, val := range m {
		arr.ArraySet(values.NewString(key), val)
	}
	return arr
}

func convertMapToObject(m map[string]*values.Value) *values.Value {
	// Create a stdClass object
	obj := values.NewObject("stdClass")
	objData := obj.Data.(*values.Object)

	// Initialize properties if needed
	if objData.Properties == nil {
		objData.Properties = make(map[string]*values.Value)
	}

	// Set properties from map
	for key, val := range m {
		objData.Properties[key] = val
	}

	return obj
}

func convertSliceToArray(s []*values.Value) *values.Value {
	arr := values.NewArray()
	for _, val := range s {
		arr.ArraySet(nil, val) // nil key means append
	}
	return arr
}

func mergeAssocAndNumeric(assoc map[string]*values.Value, numeric []*values.Value) *values.Value {
	arr := values.NewArray()

	// Add numeric indices
	for i, val := range numeric {
		arr.ArraySet(values.NewInt(int64(i)), val)
	}

	// Add associative keys
	for key, val := range assoc {
		arr.ArraySet(values.NewString(key), val)
	}

	return arr
}
