package runtime

import (
	"fmt"
	"strings"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// MySQLi statement result structure (for stmt->get_result())
type MySQLiStmtResult struct {
	Rows       []map[string]*values.Value
	FieldCount int
	CurrentRow int
}

// GetMySQLiStmtFunctions returns all MySQLi prepared statement procedural functions
func GetMySQLiStmtFunctions() []*registry.Function {
	return []*registry.Function{
		// mysqli_stmt_init - Initialize a statement and return an object for use with mysqli_stmt_prepare
		{
			Name: "mysqli_stmt_init",
			Parameters: []*registry.Parameter{
				{Name: "link", Type: "object"},
			},
			ReturnType: "object",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}

				conn, ok := extractMySQLiConnection(args[0])
				if !ok {
					return values.NewBool(false), nil
				}

				stmt := &MySQLiStmt{
					Connection:   conn,
					Query:        "",
					ParamCount:   0,
					FieldCount:   0,
					AffectedRows: 0,
					InsertID:     0,
					ErrorNo:      0,
					Error:        "",
					SQLState:     "00000",
					Params:       make([]*values.Value, 0),
				}

				return values.NewResource(stmt), nil
			},
		},

		// mysqli_stmt_prepare - Prepare an SQL statement for execution
		{
			Name: "mysqli_stmt_prepare",
			Parameters: []*registry.Parameter{
				{Name: "stmt", Type: "object"},
				{Name: "query", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil {
					return values.NewBool(false), nil
				}

				stmt, ok := extractMySQLiStmt(args[0])
				if !ok {
					return values.NewBool(false), nil
				}

				query := args[1].ToString()
				stmt.Query = query

				// Count placeholders (?) in query
				paramCount := strings.Count(query, "?")
				stmt.ParamCount = paramCount
				stmt.Params = make([]*values.Value, paramCount)

				// Clear errors
				stmt.ErrorNo = 0
				stmt.Error = ""
				stmt.SQLState = "00000"

				return values.NewBool(true), nil
			},
		},

		// mysqli_stmt_bind_param - Bind variables to a prepared statement as parameters
		{
			Name: "mysqli_stmt_bind_param",
			Parameters: []*registry.Parameter{
				{Name: "stmt", Type: "object"},
				{Name: "types", Type: "string"},
				// Variadic parameters for bound variables
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    -1, // Variadic
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil {
					return values.NewBool(false), nil
				}

				stmt, ok := extractMySQLiStmt(args[0])
				if !ok {
					return values.NewBool(false), nil
				}

				types := args[1].ToString()
				paramCount := len(types)

				// Check parameter count matches
				if len(args)-2 != paramCount {
					stmt.ErrorNo = 2031
					stmt.Error = fmt.Sprintf("Parameter count mismatch: expected %d, got %d", paramCount, len(args)-2)
					return values.NewBool(false), nil
				}

				// Store parameters with type conversion
				stmt.Params = make([]*values.Value, paramCount)
				for i := 0; i < paramCount; i++ {
					paramValue := args[i+2]
					typeChar := types[i]

					switch typeChar {
					case 'i': // integer
						stmt.Params[i] = values.NewInt(paramValue.ToInt())
					case 'd': // double
						stmt.Params[i] = values.NewFloat(paramValue.ToFloat())
					case 's': // string
						stmt.Params[i] = values.NewString(paramValue.ToString())
					case 'b': // blob (string)
						stmt.Params[i] = values.NewString(paramValue.ToString())
					default:
						stmt.ErrorNo = 2031
						stmt.Error = fmt.Sprintf("Unknown type character: %c", typeChar)
						return values.NewBool(false), nil
					}
				}

				return values.NewBool(true), nil
			},
		},

		// mysqli_stmt_execute - Execute a prepared statement
		{
			Name: "mysqli_stmt_execute",
			Parameters: []*registry.Parameter{
				{Name: "stmt", Type: "object"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}

				stmt, ok := extractMySQLiStmt(args[0])
				if !ok {
					return values.NewBool(false), nil
				}

				// Build query by replacing placeholders with actual values
				query := stmt.Query
				for _, param := range stmt.Params {
					// Find first ? and replace with parameter value
					idx := strings.Index(query, "?")
					if idx == -1 {
						break
					}

					// Format parameter value for SQL
					var paramStr string
					switch param.Type {
					case values.TypeNull:
						paramStr = "NULL"
					case values.TypeInt:
						paramStr = fmt.Sprintf("%d", param.ToInt())
					case values.TypeFloat:
						paramStr = fmt.Sprintf("%f", param.ToFloat())
					case values.TypeString:
						// Escape string for SQL
						escaped := mysqliReplaceAll(param.ToString(), "'", "''")
						paramStr = fmt.Sprintf("'%s'", escaped)
					default:
						paramStr = fmt.Sprintf("'%s'", param.ToString())
					}

					query = query[:idx] + paramStr + query[idx+1:]
				}

				// Execute query using real database connection
				db, ok := GetRealConnection(stmt.Connection)
				if !ok {
					stmt.ErrorNo = 2006
					stmt.Error = "MySQL server has gone away"
					return values.NewBool(false), nil
				}

				// Try to execute as non-query first (INSERT, UPDATE, DELETE)
				result, err := db.Exec(query)
				if err != nil {
					stmt.ErrorNo = 1064
					stmt.Error = fmt.Sprintf("Execute error: %v", err)
					return values.NewBool(false), nil
				}

				// Update statement metadata
				affectedRows, _ := result.RowsAffected()
				stmt.AffectedRows = affectedRows
				stmt.Connection.AffectedRows = affectedRows

				lastID, _ := result.LastInsertId()
				stmt.InsertID = lastID
				stmt.Connection.InsertID = lastID

				stmt.ErrorNo = 0
				stmt.Error = ""

				return values.NewBool(true), nil
			},
		},

		// mysqli_stmt_bind_result - Bind variables to a prepared statement for result storage
		{
			Name: "mysqli_stmt_bind_result",
			Parameters: []*registry.Parameter{
				{Name: "stmt", Type: "object"},
				// Variadic parameters for bound result variables
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    -1, // Variadic
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Basic implementation
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}

				stmt, ok := extractMySQLiStmt(args[0])
				if !ok {
					return values.NewBool(false), nil
				}

				// Store reference count for validation
				stmt.FieldCount = len(args) - 1

				return values.NewBool(true), nil
			},
		},

		// mysqli_stmt_fetch - Fetch results from a prepared statement into bound variables
		{
			Name: "mysqli_stmt_fetch",
			Parameters: []*registry.Parameter{
				{Name: "stmt", Type: "object"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Return null to indicate no more rows
				if len(args) == 0 || args[0] == nil {
					return values.NewNull(), nil
				}

				_, ok := extractMySQLiStmt(args[0])
				if !ok {
					return values.NewNull(), nil
				}

				// In a real implementation, this would fetch the next row
				// and populate bound variables
				return values.NewNull(), nil
			},
		},

		// mysqli_stmt_close - Close a prepared statement
		{
			Name: "mysqli_stmt_close",
			Parameters: []*registry.Parameter{
				{Name: "stmt", Type: "object"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}

				stmt, ok := extractMySQLiStmt(args[0])
				if !ok {
					return values.NewBool(false), nil
				}

				// Clear statement data
				stmt.Query = ""
				stmt.Params = nil
				stmt.ParamCount = 0
				stmt.FieldCount = 0

				return values.NewBool(true), nil
			},
		},

		// mysqli_stmt_affected_rows - Return the number of rows affected by the last statement execution
		{
			Name: "mysqli_stmt_affected_rows",
			Parameters: []*registry.Parameter{
				{Name: "stmt", Type: "object"},
			},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewInt(-1), nil
				}

				stmt, ok := extractMySQLiStmt(args[0])
				if !ok {
					return values.NewInt(-1), nil
				}

				return values.NewInt(stmt.AffectedRows), nil
			},
		},

		// mysqli_stmt_attr_get - Get the value of a statement attribute
		{
			Name: "mysqli_stmt_attr_get",
			Parameters: []*registry.Parameter{
				{Name: "stmt", Type: "object"},
				{Name: "attr", Type: "int"},
			},
			ReturnType: "int",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil {
					return values.NewInt(0), nil
				}

				_, ok := extractMySQLiStmt(args[0])
				if !ok {
					return values.NewInt(0), nil
				}

				// Stub: Return 0 for all attributes
				return values.NewInt(0), nil
			},
		},

		// mysqli_stmt_attr_set - Set a statement attribute
		{
			Name: "mysqli_stmt_attr_set",
			Parameters: []*registry.Parameter{
				{Name: "stmt", Type: "object"},
				{Name: "attr", Type: "int"},
				{Name: "value", Type: "int"},
			},
			ReturnType: "bool",
			MinArgs:    3,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 3 || args[0] == nil {
					return values.NewBool(false), nil
				}

				_, ok := extractMySQLiStmt(args[0])
				if !ok {
					return values.NewBool(false), nil
				}

				// Stub: Always return true
				return values.NewBool(true), nil
			},
		},

		// mysqli_stmt_errno - Return the error code for the most recent statement call
		{
			Name: "mysqli_stmt_errno",
			Parameters: []*registry.Parameter{
				{Name: "stmt", Type: "object"},
			},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewInt(0), nil
				}

				stmt, ok := extractMySQLiStmt(args[0])
				if !ok {
					return values.NewInt(0), nil
				}

				return values.NewInt(int64(stmt.ErrorNo)), nil
			},
		},

		// mysqli_stmt_error - Return a string description of the last statement error
		{
			Name: "mysqli_stmt_error",
			Parameters: []*registry.Parameter{
				{Name: "stmt", Type: "object"},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewString(""), nil
				}

				stmt, ok := extractMySQLiStmt(args[0])
				if !ok {
					return values.NewString(""), nil
				}

				return values.NewString(stmt.Error), nil
			},
		},

		// mysqli_stmt_field_count - Return the number of columns in the statement result set
		{
			Name: "mysqli_stmt_field_count",
			Parameters: []*registry.Parameter{
				{Name: "stmt", Type: "object"},
			},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewInt(0), nil
				}

				stmt, ok := extractMySQLiStmt(args[0])
				if !ok {
					return values.NewInt(0), nil
				}

				return values.NewInt(int64(stmt.FieldCount)), nil
			},
		},

		// mysqli_stmt_free_result - Free stored result memory for the given statement handle
		{
			Name: "mysqli_stmt_free_result",
			Parameters: []*registry.Parameter{
				{Name: "stmt", Type: "object"},
			},
			ReturnType: "void",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewNull(), nil
				}

				_, ok := extractMySQLiStmt(args[0])
				if !ok {
					return values.NewNull(), nil
				}

				// Stub: Nothing to free in current implementation
				return values.NewNull(), nil
			},
		},

		// mysqli_stmt_get_result - Get a result set from a prepared statement
		{
			Name: "mysqli_stmt_get_result",
			Parameters: []*registry.Parameter{
				{Name: "stmt", Type: "object"},
			},
			ReturnType: "object",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}

				stmt, ok := extractMySQLiStmt(args[0])
				if !ok {
					return values.NewBool(false), nil
				}

				// Build query by replacing placeholders
				query := stmt.Query
				for _, param := range stmt.Params {
					idx := strings.Index(query, "?")
					if idx == -1 {
						break
					}

					var paramStr string
					switch param.Type {
					case values.TypeNull:
						paramStr = "NULL"
					case values.TypeInt:
						paramStr = fmt.Sprintf("%d", param.ToInt())
					case values.TypeFloat:
						paramStr = fmt.Sprintf("%f", param.ToFloat())
					case values.TypeString:
						escaped := mysqliReplaceAll(param.ToString(), "'", "''")
						paramStr = fmt.Sprintf("'%s'", escaped)
					default:
						paramStr = fmt.Sprintf("'%s'", param.ToString())
					}

					query = query[:idx] + paramStr + query[idx+1:]
				}

				// Execute query and return result
				result, err := RealMySQLiQuery(stmt.Connection, query)
				if err != nil {
					stmt.ErrorNo = 1064
					stmt.Error = fmt.Sprintf("Query error: %v", err)
					return values.NewBool(false), nil
				}

				stmt.FieldCount = result.FieldCount
				return values.NewResource(result), nil
			},
		},

		// mysqli_stmt_insert_id - Get the ID generated from the previous INSERT operation
		{
			Name: "mysqli_stmt_insert_id",
			Parameters: []*registry.Parameter{
				{Name: "stmt", Type: "object"},
			},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewInt(0), nil
				}

				stmt, ok := extractMySQLiStmt(args[0])
				if !ok {
					return values.NewInt(0), nil
				}

				return values.NewInt(stmt.InsertID), nil
			},
		},

		// mysqli_stmt_more_results - Check if there are more query results from a multiple query
		{
			Name: "mysqli_stmt_more_results",
			Parameters: []*registry.Parameter{
				{Name: "stmt", Type: "object"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}

				_, ok := extractMySQLiStmt(args[0])
				if !ok {
					return values.NewBool(false), nil
				}

				// Stub: Return false (no multiple result sets)
				return values.NewBool(false), nil
			},
		},

		// mysqli_stmt_next_result - Read the next result from a multiple query
		{
			Name: "mysqli_stmt_next_result",
			Parameters: []*registry.Parameter{
				{Name: "stmt", Type: "object"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}

				_, ok := extractMySQLiStmt(args[0])
				if !ok {
					return values.NewBool(false), nil
				}

				// Stub: Return false (no more results)
				return values.NewBool(false), nil
			},
		},

		// mysqli_stmt_num_rows - Return the number of rows in the statement result set
		{
			Name: "mysqli_stmt_num_rows",
			Parameters: []*registry.Parameter{
				{Name: "stmt", Type: "object"},
			},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewInt(0), nil
				}

				_, ok := extractMySQLiStmt(args[0])
				if !ok {
					return values.NewInt(0), nil
				}

				// Stub: Return 0 (would need stored result to know row count)
				return values.NewInt(0), nil
			},
		},

		// mysqli_stmt_param_count - Return the number of parameters in the given statement
		{
			Name: "mysqli_stmt_param_count",
			Parameters: []*registry.Parameter{
				{Name: "stmt", Type: "object"},
			},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewInt(0), nil
				}

				stmt, ok := extractMySQLiStmt(args[0])
				if !ok {
					return values.NewInt(0), nil
				}

				return values.NewInt(int64(stmt.ParamCount)), nil
			},
		},

		// mysqli_stmt_reset - Reset a prepared statement
		{
			Name: "mysqli_stmt_reset",
			Parameters: []*registry.Parameter{
				{Name: "stmt", Type: "object"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}

				stmt, ok := extractMySQLiStmt(args[0])
				if !ok {
					return values.NewBool(false), nil
				}

				// Reset statement state but keep query and param count
				stmt.AffectedRows = 0
				stmt.InsertID = 0
				stmt.ErrorNo = 0
				stmt.Error = ""
				stmt.SQLState = "00000"

				return values.NewBool(true), nil
			},
		},

		// mysqli_stmt_result_metadata - Return result set metadata from a prepared statement
		{
			Name: "mysqli_stmt_result_metadata",
			Parameters: []*registry.Parameter{
				{Name: "stmt", Type: "object"},
			},
			ReturnType: "object",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}

				_, ok := extractMySQLiStmt(args[0])
				if !ok {
					return values.NewBool(false), nil
				}

				// Stub: Return false (metadata not implemented)
				return values.NewBool(false), nil
			},
		},

		// mysqli_stmt_send_long_data - Send data in blocks
		{
			Name: "mysqli_stmt_send_long_data",
			Parameters: []*registry.Parameter{
				{Name: "stmt", Type: "object"},
				{Name: "param_num", Type: "int"},
				{Name: "data", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    3,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 3 || args[0] == nil {
					return values.NewBool(false), nil
				}

				_, ok := extractMySQLiStmt(args[0])
				if !ok {
					return values.NewBool(false), nil
				}

				// Stub: Always return true
				return values.NewBool(true), nil
			},
		},

		// mysqli_stmt_sqlstate - Return the SQLSTATE error from the previous statement operation
		{
			Name: "mysqli_stmt_sqlstate",
			Parameters: []*registry.Parameter{
				{Name: "stmt", Type: "object"},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewString("00000"), nil
				}

				stmt, ok := extractMySQLiStmt(args[0])
				if !ok {
					return values.NewString("00000"), nil
				}

				return values.NewString(stmt.SQLState), nil
			},
		},

		// mysqli_stmt_store_result - Transfer a result set from a prepared statement
		{
			Name: "mysqli_stmt_store_result",
			Parameters: []*registry.Parameter{
				{Name: "stmt", Type: "object"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}

				_, ok := extractMySQLiStmt(args[0])
				if !ok {
					return values.NewBool(false), nil
				}

				// Stub: Always return true (result would be buffered internally)
				return values.NewBool(true), nil
			},
		},

		// mysqli_stmt_data_seek - Seek to an arbitrary row in statement result set
		{
			Name: "mysqli_stmt_data_seek",
			Parameters: []*registry.Parameter{
				{Name: "stmt", Type: "object"},
				{Name: "offset", Type: "int"},
			},
			ReturnType: "void",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil {
					return values.NewNull(), nil
				}

				_, ok := extractMySQLiStmt(args[0])
				if !ok {
					return values.NewNull(), nil
				}

				// Stub: Nothing to do in current implementation
				return values.NewNull(), nil
			},
		},
	}
}
