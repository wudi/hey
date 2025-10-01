package runtime

import (
	"fmt"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// MySQLi connection structure (stored in resource)
type MySQLiConnection struct {
	Host     string
	Username string
	Password string
	Database string
	Port     int
	Socket   string
	Connected bool
	AffectedRows int64
	InsertID     int64
	ErrorNo      int
	Error        string
	SQLState     string
	FieldCount   int
	WarningCount int
	Info         string
}

// MySQLi result structure
type MySQLiResult struct {
	NumRows    int64
	FieldCount int
	CurrentRow int
	Rows       []map[string]*values.Value
	Fields     []MySQLiField
}

// MySQLi field structure
type MySQLiField struct {
	Name      string
	OrgName   string
	Table     string
	OrgTable  string
	Database  string
	Def       string
	MaxLength int
	Length    int
	Charsetnr int
	Flags     int
	Type      int
	Decimals  int
}

// MySQLi prepared statement structure
type MySQLiStmt struct {
	Connection *MySQLiConnection
	Query      string
	ParamCount int
	FieldCount int
	AffectedRows int64
	InsertID     int64
	ErrorNo      int
	Error        string
	SQLState     string
	Params       []*values.Value
}

// GetMySQLiFunctions returns all MySQLi procedural functions
func GetMySQLiFunctions() []*registry.Function {
	return []*registry.Function{
		// mysqli_affected_rows - Get the number of affected rows
		{
			Name: "mysqli_affected_rows",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
			},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewInt(-1), nil
				}
				if conn, ok := args[0].Data.(*MySQLiConnection); ok {
					return values.NewInt(conn.AffectedRows), nil
				}
				return values.NewInt(-1), nil
			},
		},

		// mysqli_autocommit - Turn autocommit on or off
		{
			Name: "mysqli_autocommit",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
				{Name: "enable", Type: "bool"},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewBool(false), nil
				}
				// Stub: Always return true
				return values.NewBool(true), nil
			},
		},

		// mysqli_begin_transaction - Start a transaction
		{
			Name: "mysqli_begin_transaction",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
				{Name: "flags", Type: "int", DefaultValue: values.NewInt(0)},
				{Name: "name", Type: "string", DefaultValue: values.NewString("")},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewBool(false), nil
				}
				// Stub: Always return true
				return values.NewBool(true), nil
			},
		},

		// mysqli_change_user - Change the user of the database connection
		{
			Name: "mysqli_change_user",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
				{Name: "username", Type: "string"},
				{Name: "password", Type: "string"},
				{Name: "database", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    4,
			MaxArgs:    4,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 4 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewBool(false), nil
				}
				if conn, ok := args[0].Data.(*MySQLiConnection); ok {
					conn.Username = args[1].ToString()
					conn.Password = args[2].ToString()
					conn.Database = args[3].ToString()
					return values.NewBool(true), nil
				}
				return values.NewBool(false), nil
			},
		},

		// mysqli_character_set_name - Return the default character set for the database connection
		{
			Name: "mysqli_character_set_name",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Return utf8mb4
				return values.NewString("utf8mb4"), nil
			},
		},

		// mysqli_close - Close a previously opened database connection
		{
			Name: "mysqli_close",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewBool(false), nil
				}
				if conn, ok := args[0].Data.(*MySQLiConnection); ok {
					// Close real database connection
					RealMySQLiClose(conn)
					return values.NewBool(true), nil
				}
				return values.NewBool(true), nil
			},
		},

		// mysqli_commit - Commit the current transaction
		{
			Name: "mysqli_commit",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
				{Name: "flags", Type: "int", DefaultValue: values.NewInt(0)},
				{Name: "name", Type: "string", DefaultValue: values.NewString("")},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewBool(false), nil
				}
				// Stub: Always return true
				return values.NewBool(true), nil
			},
		},

		// mysqli_connect - Open a new connection to the MySQL server
		{
			Name: "mysqli_connect",
			Parameters: []*registry.Parameter{
				{Name: "hostname", Type: "string", DefaultValue: values.NewString("localhost")},
				{Name: "username", Type: "string", DefaultValue: values.NewString("")},
				{Name: "password", Type: "string", DefaultValue: values.NewString("")},
				{Name: "database", Type: "string", DefaultValue: values.NewString("")},
				{Name: "port", Type: "int", DefaultValue: values.NewInt(3306)},
				{Name: "socket", Type: "string", DefaultValue: values.NewString("")},
			},
			ReturnType: "object",
			MinArgs:    0,
			MaxArgs:    6,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				conn := &MySQLiConnection{
					Host:     "localhost",
					Username: "",
					Password: "",
					Database: "",
					Port:     3306,
					Socket:   "",
					Connected: false,
					AffectedRows: 0,
					InsertID:     0,
					ErrorNo:      0,
					Error:        "",
					SQLState:     "00000",
				}

				if len(args) > 0 && args[0] != nil {
					conn.Host = args[0].ToString()
				}
				if len(args) > 1 && args[1] != nil {
					conn.Username = args[1].ToString()
				}
				if len(args) > 2 && args[2] != nil {
					conn.Password = args[2].ToString()
				}
				if len(args) > 3 && args[3] != nil {
					conn.Database = args[3].ToString()
				}
				if len(args) > 4 && args[4] != nil {
					conn.Port = int(args[4].ToInt())
				}
				if len(args) > 5 && args[5] != nil {
					conn.Socket = args[5].ToString()
				}

				// Establish real database connection
				if err := RealMySQLiConnect(conn); err != nil {
					// Return false on connection failure (PHP behavior)
					return values.NewBool(false), nil
				}

				return values.NewResource(conn), nil
			},
		},

		// mysqli_connect_errno - Return the error code from last connect call
		{
			Name:       "mysqli_connect_errno",
			Parameters: []*registry.Parameter{},
			ReturnType: "int",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				return values.NewInt(int64(GetLastConnectErrno())), nil
			},
		},

		// mysqli_connect_error - Return the error description from last connect call
		{
			Name:       "mysqli_connect_error",
			Parameters: []*registry.Parameter{},
			ReturnType: "string",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				return values.NewString(GetLastConnectError()), nil
			},
		},

		// mysqli_data_seek - Adjust the result pointer to an arbitrary row
		{
			Name: "mysqli_data_seek",
			Parameters: []*registry.Parameter{
				{Name: "result", Type: "object"},
				{Name: "offset", Type: "int"},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewBool(false), nil
				}
				if result, ok := args[0].Data.(*MySQLiResult); ok {
					offset := int(args[1].ToInt())
					if offset >= 0 && offset < len(result.Rows) {
						result.CurrentRow = offset
						return values.NewBool(true), nil
					}
				}
				return values.NewBool(false), nil
			},
		},

		// mysqli_errno - Return the error code for the most recent function call
		{
			Name: "mysqli_errno",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
			},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewInt(0), nil
				}
				if conn, ok := args[0].Data.(*MySQLiConnection); ok {
					return values.NewInt(int64(conn.ErrorNo)), nil
				}
				return values.NewInt(0), nil
			},
		},

		// mysqli_error - Return a description of the last error
		{
			Name: "mysqli_error",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewString(""), nil
				}
				if conn, ok := args[0].Data.(*MySQLiConnection); ok {
					return values.NewString(conn.Error), nil
				}
				return values.NewString(""), nil
			},
		},

		// mysqli_error_list - Return a list of errors from the last command executed
		{
			Name: "mysqli_error_list",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
			},
			ReturnType: "array",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Return empty array
				return values.NewArray(), nil
			},
		},

		// mysqli_fetch_all - Fetch all result rows as an associative array, a numeric array, or both
		{
			Name: "mysqli_fetch_all",
			Parameters: []*registry.Parameter{
				{Name: "result", Type: "object"},
				{Name: "mode", Type: "int", DefaultValue: values.NewInt(1)}, // MYSQLI_NUM
			},
			ReturnType: "array",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewNull(), nil
				}
				if result, ok := args[0].Data.(*MySQLiResult); ok {
					return values.NewArray(), fmt.Errorf("stub: fetch_all not fully implemented, rows: %d", len(result.Rows))
				}
				return values.NewNull(), nil
			},
		},

		// mysqli_fetch_array - Fetch the next row of a result set as an associative, a numeric array, or both
		{
			Name: "mysqli_fetch_array",
			Parameters: []*registry.Parameter{
				{Name: "result", Type: "object"},
				{Name: "mode", Type: "int", DefaultValue: values.NewInt(3)}, // MYSQLI_BOTH
			},
			ReturnType: "array",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewNull(), nil
				}

				result, ok := args[0].Data.(*MySQLiResult)
				if !ok {
					return values.NewNull(), nil
				}

				if result.CurrentRow >= len(result.Rows) {
					return values.NewNull(), nil
				}

				mode := int64(3) // MYSQLI_BOTH
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
			},
		},

		// mysqli_fetch_assoc - Fetch the next row of a result set as an associative array
		{
			Name: "mysqli_fetch_assoc",
			Parameters: []*registry.Parameter{
				{Name: "result", Type: "object"},
			},
			ReturnType: "array",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewNull(), nil
				}
				if result, ok := args[0].Data.(*MySQLiResult); ok {
					if result.CurrentRow < len(result.Rows) {
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
				}
				// No more rows
				return values.NewNull(), nil
			},
		},

		// mysqli_fetch_field - Fetch meta-data for a single field
		{
			Name: "mysqli_fetch_field",
			Parameters: []*registry.Parameter{
				{Name: "result", Type: "object"},
			},
			ReturnType: "object",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Return null
				return values.NewNull(), nil
			},
		},

		// mysqli_fetch_field_direct - Fetch meta-data for a single field
		{
			Name: "mysqli_fetch_field_direct",
			Parameters: []*registry.Parameter{
				{Name: "result", Type: "object"},
				{Name: "index", Type: "int"},
			},
			ReturnType: "object",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Return null
				return values.NewNull(), nil
			},
		},

		// mysqli_fetch_fields - Fetch meta-data for all fields
		{
			Name: "mysqli_fetch_fields",
			Parameters: []*registry.Parameter{
				{Name: "result", Type: "object"},
			},
			ReturnType: "array",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Return empty array
				return values.NewArray(), nil
			},
		},

		// mysqli_fetch_lengths - Return the lengths of the columns of the current row
		{
			Name: "mysqli_fetch_lengths",
			Parameters: []*registry.Parameter{
				{Name: "result", Type: "object"},
			},
			ReturnType: "array",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Return empty array
				return values.NewArray(), nil
			},
		},

		// mysqli_fetch_object - Fetch the next row of a result set as an object
		{
			Name: "mysqli_fetch_object",
			Parameters: []*registry.Parameter{
				{Name: "result", Type: "object"},
				{Name: "class", Type: "string", DefaultValue: values.NewString("stdClass")},
				{Name: "constructor_args", Type: "array", DefaultValue: values.NewArray()},
			},
			ReturnType: "object",
			MinArgs:    1,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewNull(), nil
				}

				// Extract MySQLiResult from resource or object
				var result *MySQLiResult
				if args[0].Type == values.TypeResource {
					r, ok := args[0].Data.(*MySQLiResult)
					if !ok {
						return values.NewNull(), nil
					}
					result = r
				} else {
					r, ok := extractMySQLiResult(args[0])
					if !ok {
						return values.NewNull(), nil
					}
					result = r
				}

				// Check if there are more rows
				if result.CurrentRow >= len(result.Rows) {
					return values.NewNull(), nil
				}

				// Get current row
				row := result.Rows[result.CurrentRow]
				result.CurrentRow++

				// Create stdClass object with row data as properties
				obj := &values.Object{
					ClassName:  "stdClass",
					Properties: make(map[string]*values.Value),
				}

				// Copy row data to object properties
				for key, val := range row {
					obj.Properties[key] = val
				}

				return &values.Value{
					Type: values.TypeObject,
					Data: obj,
				}, nil
			},
		},

		// mysqli_fetch_row - Fetch the next row of a result set as an enumerated array
		{
			Name: "mysqli_fetch_row",
			Parameters: []*registry.Parameter{
				{Name: "result", Type: "object"},
			},
			ReturnType: "array",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewNull(), nil
				}

				result, ok := args[0].Data.(*MySQLiResult)
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
			},
		},

		// mysqli_field_count - Return the number of columns for the most recent query
		{
			Name: "mysqli_field_count",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
			},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewInt(0), nil
				}
				if conn, ok := args[0].Data.(*MySQLiConnection); ok {
					return values.NewInt(int64(conn.FieldCount)), nil
				}
				return values.NewInt(0), nil
			},
		},

		// mysqli_field_seek - Set result pointer to a specified field offset
		{
			Name: "mysqli_field_seek",
			Parameters: []*registry.Parameter{
				{Name: "result", Type: "object"},
				{Name: "index", Type: "int"},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Always return true
				return values.NewBool(true), nil
			},
		},

		// mysqli_field_tell - Get current field offset of a result pointer
		{
			Name: "mysqli_field_tell",
			Parameters: []*registry.Parameter{
				{Name: "result", Type: "object"},
			},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Return 0
				return values.NewInt(0), nil
			},
		},

		// mysqli_free_result - Free the memory associated with a result
		{
			Name: "mysqli_free_result",
			Parameters: []*registry.Parameter{
				{Name: "result", Type: "object"},
			},
			ReturnType: "void",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Do nothing
				return values.NewNull(), nil
			},
		},

		// mysqli_get_charset - Get character set information
		{
			Name: "mysqli_get_charset",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
			},
			ReturnType: "object",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Return null
				return values.NewNull(), nil
			},
		},

		// mysqli_get_client_info - Get MySQL client info
		{
			Name: "mysqli_get_client_info",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object", DefaultValue: values.NewNull()},
			},
			ReturnType: "string",
			MinArgs:    0,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Return version string
				return values.NewString("hey-mysqli-stub-1.0.0"), nil
			},
		},

		// mysqli_get_client_version - Get MySQL client version as an integer
		{
			Name:       "mysqli_get_client_version",
			Parameters: []*registry.Parameter{},
			ReturnType: "int",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Return 80030 (8.0.30)
				return values.NewInt(80030), nil
			},
		},

		// mysqli_get_host_info - Get connection host information
		{
			Name: "mysqli_get_host_info",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewString(""), nil
				}
				if conn, ok := args[0].Data.(*MySQLiConnection); ok {
					return values.NewString(fmt.Sprintf("%s via TCP/IP", conn.Host)), nil
				}
				return values.NewString("localhost via TCP/IP"), nil
			},
		},

		// mysqli_get_proto_info - Get MySQL protocol information
		{
			Name: "mysqli_get_proto_info",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
			},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Return protocol version 10
				return values.NewInt(10), nil
			},
		},

		// mysqli_get_server_info - Get MySQL server version
		{
			Name: "mysqli_get_server_info",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Return server version
				return values.NewString("8.0.30"), nil
			},
		},

		// mysqli_get_server_version - Get MySQL server version as an integer
		{
			Name: "mysqli_get_server_version",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
			},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Return 80030 (8.0.30)
				return values.NewInt(80030), nil
			},
		},

		// mysqli_info - Get information about the most recently executed query
		{
			Name: "mysqli_info",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewNull(), nil
				}
				if conn, ok := args[0].Data.(*MySQLiConnection); ok {
					return values.NewString(conn.Info), nil
				}
				return values.NewNull(), nil
			},
		},

		// mysqli_init - Initialize MySQLi and return an object for use with mysqli_real_connect
		{
			Name:       "mysqli_init",
			Parameters: []*registry.Parameter{},
			ReturnType: "object",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				conn := &MySQLiConnection{
					Host:      "localhost",
					Username:  "",
					Password:  "",
					Database:  "",
					Port:      3306,
					Socket:    "",
					Connected: false,
					ErrorNo:   0,
					Error:     "",
					SQLState:  "00000",
				}
				return values.NewResource(conn), nil
			},
		},

		// mysqli_insert_id - Get the ID generated from the previous INSERT operation
		{
			Name: "mysqli_insert_id",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
			},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewInt(0), nil
				}
				if conn, ok := args[0].Data.(*MySQLiConnection); ok {
					return values.NewInt(conn.InsertID), nil
				}
				return values.NewInt(0), nil
			},
		},

		// mysqli_more_results - Check if there are more query results from a multi query
		{
			Name: "mysqli_more_results",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Return false
				return values.NewBool(false), nil
			},
		},

		// mysqli_multi_query - Perform one or more queries
		{
			Name: "mysqli_multi_query",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
				{Name: "query", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Return true
				return values.NewBool(true), nil
			},
		},

		// mysqli_next_result - Prepare next result from multi_query
		{
			Name: "mysqli_next_result",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Return false (no more results)
				return values.NewBool(false), nil
			},
		},

		// mysqli_num_fields - Get the number of fields in a result
		{
			Name: "mysqli_num_fields",
			Parameters: []*registry.Parameter{
				{Name: "result", Type: "object"},
			},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewInt(0), nil
				}
				if result, ok := args[0].Data.(*MySQLiResult); ok {
					return values.NewInt(int64(result.FieldCount)), nil
				}
				return values.NewInt(0), nil
			},
		},

		// mysqli_num_rows - Get the number of rows in a result
		{
			Name: "mysqli_num_rows",
			Parameters: []*registry.Parameter{
				{Name: "result", Type: "object"},
			},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewInt(0), nil
				}
				if result, ok := args[0].Data.(*MySQLiResult); ok {
					return values.NewInt(result.NumRows), nil
				}
				return values.NewInt(0), nil
			},
		},

		// mysqli_options - Set connection options
		{
			Name: "mysqli_options",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
				{Name: "option", Type: "int"},
				{Name: "value", Type: "mixed"},
			},
			ReturnType: "bool",
			MinArgs:    3,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Always return true
				return values.NewBool(true), nil
			},
		},

		// mysqli_ping - Ping a server connection or reconnect if there is no connection
		{
			Name: "mysqli_ping",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Always return true
				return values.NewBool(true), nil
			},
		},

		// mysqli_poll - Poll connections
		{
			Name: "mysqli_poll",
			Parameters: []*registry.Parameter{
				{Name: "read", Type: "array", IsReference: true},
				{Name: "error", Type: "array", IsReference: true},
				{Name: "reject", Type: "array", IsReference: true},
				{Name: "seconds", Type: "int"},
				{Name: "microseconds", Type: "int", DefaultValue: values.NewInt(0)},
			},
			ReturnType: "int",
			MinArgs:    4,
			MaxArgs:    5,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Return 0
				return values.NewInt(0), nil
			},
		},

		// mysqli_prepare - Prepare an SQL statement for execution
		{
			Name: "mysqli_prepare",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
				{Name: "query", Type: "string"},
			},
			ReturnType: "object",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewBool(false), nil
				}

				conn, ok := args[0].Data.(*MySQLiConnection)
				if !ok {
					return values.NewBool(false), nil
				}

				query := args[1].ToString()
				paramCount := 0
				for i := 0; i < len(query); i++ {
					if query[i] == '?' {
						paramCount++
					}
				}

				stmt := &MySQLiStmt{
					Connection:   conn,
					Query:        query,
					ParamCount:   paramCount,
					FieldCount:   0,
					AffectedRows: 0,
					InsertID:     0,
					ErrorNo:      0,
					Error:        "",
					SQLState:     "00000",
					Params:       make([]*values.Value, paramCount),
				}

				return values.NewResource(stmt), nil
			},
		},

		// mysqli_query - Perform a query on the database
		{
			Name: "mysqli_query",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
				{Name: "query", Type: "string"},
				{Name: "result_mode", Type: "int", DefaultValue: values.NewInt(0)},
			},
			ReturnType: "mixed",
			MinArgs:    2,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil {
					return values.NewBool(false), nil
				}

				conn, ok := extractMySQLiConnection(args[0])
				if !ok {
					return values.NewBool(false), nil
				}

				query := args[1].ToString()

				// Execute real query
				result, err := RealMySQLiQuery(conn, query)
				if err != nil {
					return values.NewBool(false), nil
				}

				// For non-SELECT queries (INSERT, UPDATE, DELETE), result is nil
				if result == nil {
					return values.NewBool(true), nil
				}

				// Return result resource for SELECT queries
				return values.NewResource(result), nil
			},
		},

		// mysqli_real_connect - Open a new connection to the MySQL server
		{
			Name: "mysqli_real_connect",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
				{Name: "hostname", Type: "string", DefaultValue: values.NewString("")},
				{Name: "username", Type: "string", DefaultValue: values.NewString("")},
				{Name: "password", Type: "string", DefaultValue: values.NewString("")},
				{Name: "database", Type: "string", DefaultValue: values.NewString("")},
				{Name: "port", Type: "int", DefaultValue: values.NewInt(3306)},
				{Name: "socket", Type: "string", DefaultValue: values.NewString("")},
				{Name: "flags", Type: "int", DefaultValue: values.NewInt(0)},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    8,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewBool(false), nil
				}

				if conn, ok := args[0].Data.(*MySQLiConnection); ok {
					if len(args) > 1 && args[1] != nil {
						conn.Host = args[1].ToString()
					}
					if len(args) > 2 && args[2] != nil {
						conn.Username = args[2].ToString()
					}
					if len(args) > 3 && args[3] != nil {
						conn.Password = args[3].ToString()
					}
					if len(args) > 4 && args[4] != nil {
						conn.Database = args[4].ToString()
					}
					if len(args) > 5 && args[5] != nil {
						conn.Port = int(args[5].ToInt())
					}
					if len(args) > 6 && args[6] != nil {
						conn.Socket = args[6].ToString()
					}
					conn.Connected = true
					return values.NewBool(true), nil
				}

				return values.NewBool(false), nil
			},
		},

		// mysqli_real_escape_string - Escape special characters in a string for use in an SQL statement
		{
			Name: "mysqli_real_escape_string",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
				{Name: "string", Type: "string"},
			},
			ReturnType: "string",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 {
					return values.NewString(""), nil
				}

				// Basic escaping
				str := args[1].ToString()
				str = mysqliReplaceAll(str, "\\", "\\\\")
				str = mysqliReplaceAll(str, "'", "\\'")
				str = mysqliReplaceAll(str, "\"", "\\\"")
				str = mysqliReplaceAll(str, "\n", "\\n")
				str = mysqliReplaceAll(str, "\r", "\\r")
				str = mysqliReplaceAll(str, "\x00", "\\0")
				str = mysqliReplaceAll(str, "\x1a", "\\Z")

				return values.NewString(str), nil
			},
		},

		// mysqli_real_query - Execute an SQL query
		{
			Name: "mysqli_real_query",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
				{Name: "query", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Always return true
				return values.NewBool(true), nil
			},
		},

		// mysqli_reap_async_query - Get result from async query
		{
			Name: "mysqli_reap_async_query",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
			},
			ReturnType: "mixed",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Return false
				return values.NewBool(false), nil
			},
		},

		// mysqli_rollback - Rollback the current transaction
		{
			Name: "mysqli_rollback",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
				{Name: "flags", Type: "int", DefaultValue: values.NewInt(0)},
				{Name: "name", Type: "string", DefaultValue: values.NewString("")},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Always return true
				return values.NewBool(true), nil
			},
		},

		// mysqli_select_db - Select the default database for database queries
		{
			Name: "mysqli_select_db",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
				{Name: "database", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewBool(false), nil
				}
				if conn, ok := args[0].Data.(*MySQLiConnection); ok {
					conn.Database = args[1].ToString()
					return values.NewBool(true), nil
				}
				return values.NewBool(false), nil
			},
		},

		// mysqli_set_charset - Set the default client character set
		{
			Name: "mysqli_set_charset",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
				{Name: "charset", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Always return true
				return values.NewBool(true), nil
			},
		},

		// mysqli_sqlstate - Return the SQLSTATE error from previous MySQL operation
		{
			Name: "mysqli_sqlstate",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewString("00000"), nil
				}
				if conn, ok := args[0].Data.(*MySQLiConnection); ok {
					return values.NewString(conn.SQLState), nil
				}
				return values.NewString("00000"), nil
			},
		},

		// mysqli_stat - Get current system status
		{
			Name: "mysqli_stat",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Return status string
				return values.NewString("Uptime: 0  Threads: 1  Questions: 0  Slow queries: 0  Opens: 0  Flush tables: 0  Open tables: 0"), nil
			},
		},

		// mysqli_store_result - Transfer a result set from the last query
		{
			Name: "mysqli_store_result",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
				{Name: "mode", Type: "int", DefaultValue: values.NewInt(0)},
			},
			ReturnType: "object",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Return empty result
				result := &MySQLiResult{
					NumRows:    0,
					FieldCount: 0,
					CurrentRow: 0,
					Rows:       make([]map[string]*values.Value, 0),
					Fields:     make([]MySQLiField, 0),
				}
				return values.NewResource(result), nil
			},
		},

		// mysqli_thread_id - Return the thread ID for the current connection
		{
			Name: "mysqli_thread_id",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
			},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Return thread ID 1
				return values.NewInt(1), nil
			},
		},

		// mysqli_thread_safe - Return whether thread safety is given or not
		{
			Name:       "mysqli_thread_safe",
			Parameters: []*registry.Parameter{},
			ReturnType: "bool",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Return true
				return values.NewBool(true), nil
			},
		},

		// mysqli_use_result - Initiate a result set retrieval
		{
			Name: "mysqli_use_result",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
			},
			ReturnType: "object",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Return empty result
				result := &MySQLiResult{
					NumRows:    0,
					FieldCount: 0,
					CurrentRow: 0,
					Rows:       make([]map[string]*values.Value, 0),
					Fields:     make([]MySQLiField, 0),
				}
				return values.NewResource(result), nil
			},
		},

		// mysqli_warning_count - Return the number of warnings from the last query for the given link
		{
			Name: "mysqli_warning_count",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
			},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewInt(0), nil
				}
				if conn, ok := args[0].Data.(*MySQLiConnection); ok {
					return values.NewInt(int64(conn.WarningCount)), nil
				}
				return values.NewInt(0), nil
			},
		},
	}
}

// Helper function for string replacement (mysqli version)
func mysqliReplaceAll(s, old, new string) string {
	result := ""
	for {
		idx := 0
		found := false
		for i := 0; i <= len(s)-len(old); i++ {
			if s[i:i+len(old)] == old {
				result += s[idx:i] + new
				idx = i + len(old)
				s = s[idx:]
				found = true
				break
			}
		}
		if !found {
			result += s
			break
		}
	}
	return result
}
