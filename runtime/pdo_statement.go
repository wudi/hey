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
			idx := 1
			for _, val := range params.Elements {
				stmt.BindValue(idx, val, int(pdo.ParamStr))
				idx++
			}
		}
	}

	// Execute statement
	if hasStmt && stmtVal.Type == values.TypeResource {
		stmt := stmtVal.Data.(pdo.Stmt)

		// Prepared statement execution
		rows, err := stmt.Query()
		if err != nil {
			return values.NewBool(false), nil
		}

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
	return values.NewNull(), nil
}

// pdoStmtErrorInfo implements $stmt->errorInfo()
// args[0] = $this
func pdoStmtErrorInfo(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	return values.NewArray(), nil
}

// Helper functions for data conversion

func convertMapToArray(m map[string]*values.Value) *values.Value {
	arr := values.NewArray()
	for key, val := range m {
		arr.ArraySet(values.NewString(key), val)
	}
	return arr
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
