package runtime

import (
	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// mysqliPropertyDescriptors returns property descriptors for mysqli class
// These are read-only properties that reflect the connection state
func mysqliPropertyDescriptors() map[string]*registry.PropertyDescriptor {
	return map[string]*registry.PropertyDescriptor{
		"affected_rows": {
			Name:         "affected_rows",
			Visibility:   "public",
			IsStatic:     false,
			DefaultValue: values.NewInt(0),
		},
		"client_info": {
			Name:         "client_info",
			Visibility:   "public",
			IsStatic:     false,
			DefaultValue: values.NewString("hey-mysqli-stub-1.0.0"),
		},
		"client_version": {
			Name:         "client_version",
			Visibility:   "public",
			IsStatic:     false,
			DefaultValue: values.NewInt(80030),
		},
		"connect_errno": {
			Name:         "connect_errno",
			Visibility:   "public",
			IsStatic:     false,
			DefaultValue: values.NewInt(0),
		},
		"connect_error": {
			Name:         "connect_error",
			Visibility:   "public",
			IsStatic:     false,
			DefaultValue: values.NewNull(),
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
		"host_info": {
			Name:         "host_info",
			Visibility:   "public",
			IsStatic:     false,
			DefaultValue: values.NewString("localhost via TCP/IP"),
		},
		"info": {
			Name:         "info",
			Visibility:   "public",
			IsStatic:     false,
			DefaultValue: values.NewNull(),
		},
		"insert_id": {
			Name:         "insert_id",
			Visibility:   "public",
			IsStatic:     false,
			DefaultValue: values.NewInt(0),
		},
		"protocol_version": {
			Name:         "protocol_version",
			Visibility:   "public",
			IsStatic:     false,
			DefaultValue: values.NewInt(10),
		},
		"server_info": {
			Name:         "server_info",
			Visibility:   "public",
			IsStatic:     false,
			DefaultValue: values.NewString("8.0.30"),
		},
		"server_version": {
			Name:         "server_version",
			Visibility:   "public",
			IsStatic:     false,
			DefaultValue: values.NewInt(80030),
		},
		"sqlstate": {
			Name:         "sqlstate",
			Visibility:   "public",
			IsStatic:     false,
			DefaultValue: values.NewString("00000"),
		},
		"thread_id": {
			Name:         "thread_id",
			Visibility:   "public",
			IsStatic:     false,
			DefaultValue: values.NewInt(1),
		},
		"warning_count": {
			Name:         "warning_count",
			Visibility:   "public",
			IsStatic:     false,
			DefaultValue: values.NewInt(0),
		},
	}
}

// mysqliMethodDescriptors returns method descriptors for mysqli class
func mysqliMethodDescriptors() map[string]*registry.MethodDescriptor {
	return map[string]*registry.MethodDescriptor{
		// Connection Methods
		"__construct": newMySQLiMethod("__construct",
			[]registry.ParameterDescriptor{
				{Name: "hostname", Type: "string", HasDefault: true, DefaultValue: values.NewString("localhost")},
				{Name: "username", Type: "string", HasDefault: true, DefaultValue: values.NewString("")},
				{Name: "password", Type: "string", HasDefault: true, DefaultValue: values.NewString("")},
				{Name: "database", Type: "string", HasDefault: true, DefaultValue: values.NewString("")},
				{Name: "port", Type: "int", HasDefault: true, DefaultValue: values.NewInt(3306)},
				{Name: "socket", Type: "string", HasDefault: true, DefaultValue: values.NewString("")},
			},
			"", mysqliConstruct),
		"connect": newMySQLiMethod("connect",
			[]registry.ParameterDescriptor{
				{Name: "hostname", Type: "string", HasDefault: true, DefaultValue: values.NewString("localhost")},
				{Name: "username", Type: "string", HasDefault: true, DefaultValue: values.NewString("")},
				{Name: "password", Type: "string", HasDefault: true, DefaultValue: values.NewString("")},
				{Name: "database", Type: "string", HasDefault: true, DefaultValue: values.NewString("")},
				{Name: "port", Type: "int", HasDefault: true, DefaultValue: values.NewInt(3306)},
				{Name: "socket", Type: "string", HasDefault: true, DefaultValue: values.NewString("")},
			},
			"", mysqliConnect),
		"real_connect": newMySQLiMethod("real_connect",
			[]registry.ParameterDescriptor{
				{Name: "hostname", Type: "string", HasDefault: true, DefaultValue: values.NewString("")},
				{Name: "username", Type: "string", HasDefault: true, DefaultValue: values.NewString("")},
				{Name: "password", Type: "string", HasDefault: true, DefaultValue: values.NewString("")},
				{Name: "database", Type: "string", HasDefault: true, DefaultValue: values.NewString("")},
				{Name: "port", Type: "int", HasDefault: true, DefaultValue: values.NewInt(3306)},
				{Name: "socket", Type: "string", HasDefault: true, DefaultValue: values.NewString("")},
				{Name: "flags", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
			},
			"bool", mysqliRealConnect),
		"init": newMySQLiMethod("init",
			[]registry.ParameterDescriptor{},
			"", mysqliInit),
		"close": newMySQLiMethod("close",
			[]registry.ParameterDescriptor{},
			"bool", mysqliClose),
		"change_user": newMySQLiMethod("change_user",
			[]registry.ParameterDescriptor{
				{Name: "username", Type: "string"},
				{Name: "password", Type: "string"},
				{Name: "database", Type: "string"},
			},
			"bool", mysqliChangeUser),
		"select_db": newMySQLiMethod("select_db",
			[]registry.ParameterDescriptor{
				{Name: "database", Type: "string"},
			},
			"bool", mysqliSelectDb),
		"ping": newMySQLiMethod("ping",
			[]registry.ParameterDescriptor{},
			"bool", mysqliPing),

		// Query Methods
		"query": newMySQLiMethod("query",
			[]registry.ParameterDescriptor{
				{Name: "query", Type: "string"},
				{Name: "result_mode", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
			},
			"mixed", mysqliQuery),
		"real_query": newMySQLiMethod("real_query",
			[]registry.ParameterDescriptor{
				{Name: "query", Type: "string"},
			},
			"bool", mysqliRealQuery),
		"multi_query": newMySQLiMethod("multi_query",
			[]registry.ParameterDescriptor{
				{Name: "query", Type: "string"},
			},
			"bool", mysqliMultiQuery),
		"prepare": newMySQLiMethod("prepare",
			[]registry.ParameterDescriptor{
				{Name: "query", Type: "string"},
			},
			"object", mysqliPrepare),
		"store_result": newMySQLiMethod("store_result",
			[]registry.ParameterDescriptor{
				{Name: "mode", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
			},
			"object", mysqliStoreResult),
		"use_result": newMySQLiMethod("use_result",
			[]registry.ParameterDescriptor{},
			"object", mysqliUseResult),
		"more_results": newMySQLiMethod("more_results",
			[]registry.ParameterDescriptor{},
			"bool", mysqliMoreResults),
		"next_result": newMySQLiMethod("next_result",
			[]registry.ParameterDescriptor{},
			"bool", mysqliNextResult),

		// Transaction Methods
		"autocommit": newMySQLiMethod("autocommit",
			[]registry.ParameterDescriptor{
				{Name: "enable", Type: "bool"},
			},
			"bool", mysqliAutocommit),
		"begin_transaction": newMySQLiMethod("begin_transaction",
			[]registry.ParameterDescriptor{
				{Name: "flags", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
				{Name: "name", Type: "string", HasDefault: true, DefaultValue: values.NewString("")},
			},
			"bool", mysqliBeginTransaction),
		"commit": newMySQLiMethod("commit",
			[]registry.ParameterDescriptor{
				{Name: "flags", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
				{Name: "name", Type: "string", HasDefault: true, DefaultValue: values.NewString("")},
			},
			"bool", mysqliCommit),
		"rollback": newMySQLiMethod("rollback",
			[]registry.ParameterDescriptor{
				{Name: "flags", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
				{Name: "name", Type: "string", HasDefault: true, DefaultValue: values.NewString("")},
			},
			"bool", mysqliRollback),

		// Info/Utility Methods
		"get_charset": newMySQLiMethod("get_charset",
			[]registry.ParameterDescriptor{},
			"object", mysqliGetCharset),
		"character_set_name": newMySQLiMethod("character_set_name",
			[]registry.ParameterDescriptor{},
			"string", mysqliCharacterSetName),
		"set_charset": newMySQLiMethod("set_charset",
			[]registry.ParameterDescriptor{
				{Name: "charset", Type: "string"},
			},
			"bool", mysqliSetCharset),
		"real_escape_string": newMySQLiMethod("real_escape_string",
			[]registry.ParameterDescriptor{
				{Name: "string", Type: "string"},
			},
			"string", mysqliRealEscapeString),
		"escape_string": newMySQLiMethod("escape_string",
			[]registry.ParameterDescriptor{
				{Name: "string", Type: "string"},
			},
			"string", mysqliEscapeString),
		"options": newMySQLiMethod("options",
			[]registry.ParameterDescriptor{
				{Name: "option", Type: "int"},
				{Name: "value", Type: "mixed"},
			},
			"bool", mysqliOptions),
		"stat": newMySQLiMethod("stat",
			[]registry.ParameterDescriptor{},
			"string", mysqliStat),
		"thread_id": newMySQLiMethod("thread_id",
			[]registry.ParameterDescriptor{},
			"int", mysqliThreadId),

		// Error/Info Getters
		"errno": newMySQLiMethod("errno",
			[]registry.ParameterDescriptor{},
			"int", mysqliErrno),
		"error": newMySQLiMethod("error",
			[]registry.ParameterDescriptor{},
			"string", mysqliError),
		"error_list": newMySQLiMethod("error_list",
			[]registry.ParameterDescriptor{},
			"array", mysqliErrorList),
		"sqlstate": newMySQLiMethod("sqlstate",
			[]registry.ParameterDescriptor{},
			"string", mysqliSqlstate),
		"info": newMySQLiMethod("info",
			[]registry.ParameterDescriptor{},
			"string", mysqliInfo),
		"warning_count": newMySQLiMethod("warning_count",
			[]registry.ParameterDescriptor{},
			"int", mysqliWarningCount),

		// Result/Statement Getters
		"affected_rows": newMySQLiMethod("affected_rows",
			[]registry.ParameterDescriptor{},
			"int", mysqliAffectedRows),
		"insert_id": newMySQLiMethod("insert_id",
			[]registry.ParameterDescriptor{},
			"int", mysqliInsertId),
		"field_count": newMySQLiMethod("field_count",
			[]registry.ParameterDescriptor{},
			"int", mysqliFieldCount),

		// Server Info Methods
		"get_client_info": newMySQLiMethod("get_client_info",
			[]registry.ParameterDescriptor{},
			"string", mysqliGetClientInfo),
		"get_host_info": newMySQLiMethod("get_host_info",
			[]registry.ParameterDescriptor{},
			"string", mysqliGetHostInfo),
		"get_proto_info": newMySQLiMethod("get_proto_info",
			[]registry.ParameterDescriptor{},
			"int", mysqliGetProtoInfo),
		"get_server_info": newMySQLiMethod("get_server_info",
			[]registry.ParameterDescriptor{},
			"string", mysqliGetServerInfo),
		"get_server_version": newMySQLiMethod("get_server_version",
			[]registry.ParameterDescriptor{},
			"int", mysqliGetServerVersion),

		// Property Getter Methods (11 methods)
		"get_affected_rows": newMySQLiMethod("get_affected_rows",
			[]registry.ParameterDescriptor{},
			"int", mysqliAffectedRows),
		"get_client_version": newMySQLiMethod("get_client_version",
			[]registry.ParameterDescriptor{},
			"int", func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				return values.NewInt(80030), nil
			}),
		"get_errno": newMySQLiMethod("get_errno",
			[]registry.ParameterDescriptor{},
			"int", mysqliErrno),
		"get_error": newMySQLiMethod("get_error",
			[]registry.ParameterDescriptor{},
			"string", mysqliError),
		"get_error_list": newMySQLiMethod("get_error_list",
			[]registry.ParameterDescriptor{},
			"array", mysqliErrorList),
		"get_field_count": newMySQLiMethod("get_field_count",
			[]registry.ParameterDescriptor{},
			"int", mysqliFieldCount),
		"get_info": newMySQLiMethod("get_info",
			[]registry.ParameterDescriptor{},
			"string|null", mysqliInfo),
		"get_insert_id": newMySQLiMethod("get_insert_id",
			[]registry.ParameterDescriptor{},
			"int", mysqliInsertId),
		"get_protocol_version": newMySQLiMethod("get_protocol_version",
			[]registry.ParameterDescriptor{},
			"int", mysqliGetProtoInfo),
		"get_sqlstate": newMySQLiMethod("get_sqlstate",
			[]registry.ParameterDescriptor{},
			"string", mysqliSqlstate),
		"get_thread_id": newMySQLiMethod("get_thread_id",
			[]registry.ParameterDescriptor{},
			"int", mysqliThreadId),
		"get_warning_count": newMySQLiMethod("get_warning_count",
			[]registry.ParameterDescriptor{},
			"int", mysqliWarningCount),

		// Advanced/Debug Methods (10 methods)
		"get_connection_stats": newMySQLiMethod("get_connection_stats",
			[]registry.ParameterDescriptor{},
			"array", mysqliGetConnectionStats),
		"get_warnings": newMySQLiMethod("get_warnings",
			[]registry.ParameterDescriptor{},
			"object|false", mysqliGetWarnings),
		"poll": newMySQLiMethod("poll",
			[]registry.ParameterDescriptor{
				{Name: "read", Type: "array", IsReference: true},
				{Name: "error", Type: "array", IsReference: true},
				{Name: "reject", Type: "array", IsReference: true},
				{Name: "seconds", Type: "int"},
				{Name: "microseconds", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
			},
			"int", mysqliPoll),
		"reap_async_query": newMySQLiMethod("reap_async_query",
			[]registry.ParameterDescriptor{},
			"mixed", mysqliReapAsyncQuery),
		"refresh": newMySQLiMethod("refresh",
			[]registry.ParameterDescriptor{
				{Name: "flags", Type: "int"},
			},
			"bool", mysqliRefresh),
		"ssl_set": newMySQLiMethod("ssl_set",
			[]registry.ParameterDescriptor{
				{Name: "key", Type: "string|null"},
				{Name: "cert", Type: "string|null"},
				{Name: "ca", Type: "string|null"},
				{Name: "capath", Type: "string|null"},
				{Name: "cipher", Type: "string|null"},
			},
			"bool", mysqliSslSet),
		"dump_debug_info": newMySQLiMethod("dump_debug_info",
			[]registry.ParameterDescriptor{},
			"bool", mysqliDumpDebugInfo),
		"debug": newMySQLiMethod("debug",
			[]registry.ParameterDescriptor{
				{Name: "options", Type: "string"},
			},
			"bool", mysqliDebug),
		"kill": newMySQLiMethod("kill",
			[]registry.ParameterDescriptor{
				{Name: "process_id", Type: "int"},
			},
			"bool", mysqliKill),
	}
}

// Connection Method Implementations

func mysqliConstruct(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewNull(), nil
	}

	thisObj := args[0]
	obj, ok := thisObj.Data.(*values.Object)
	if !ok {
		return values.NewNull(), nil
	}

	// Create connection
	conn := &MySQLiConnection{
		Host:     "localhost",
		Username: "",
		Password: "",
		Database: "",
		Port:     3306,
		Socket:   "",
		Connected: false,
	}

	if len(args) > 1 && args[1].Type != values.TypeNull {
		conn.Host = args[1].ToString()
	}
	if len(args) > 2 && args[2].Type != values.TypeNull {
		conn.Username = args[2].ToString()
	}
	if len(args) > 3 && args[3].Type != values.TypeNull {
		conn.Password = args[3].ToString()
	}
	if len(args) > 4 && args[4].Type != values.TypeNull {
		conn.Database = args[4].ToString()
	}
	if len(args) > 5 && args[5].Type != values.TypeNull {
		conn.Port = int(args[5].ToInt())
	}
	if len(args) > 6 && args[6].Type != values.TypeNull {
		conn.Socket = args[6].ToString()
	}

	// Attempt real connection
	RealMySQLiConnect(conn)

	// Store in object
	if obj.Properties == nil {
		obj.Properties = make(map[string]*values.Value)
	}
	obj.Properties["__mysqli_connection"] = values.NewResource(conn)

	return values.NewNull(), nil
}

func mysqliConnect(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	// Alias for __construct
	return mysqliConstruct(ctx, args)
}

func mysqliRealConnect(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewBool(false), nil
	}

	conn, ok := extractMySQLiConnection(args[0])
	if !ok {
		return values.NewBool(false), nil
	}

	if len(args) > 1 && args[1].Type != values.TypeNull {
		conn.Host = args[1].ToString()
	}
	if len(args) > 2 && args[2].Type != values.TypeNull {
		conn.Username = args[2].ToString()
	}
	if len(args) > 3 && args[3].Type != values.TypeNull {
		conn.Password = args[3].ToString()
	}
	if len(args) > 4 && args[4].Type != values.TypeNull {
		conn.Database = args[4].ToString()
	}
	if len(args) > 5 && args[5].Type != values.TypeNull {
		conn.Port = int(args[5].ToInt())
	}
	if len(args) > 6 && args[6].Type != values.TypeNull {
		conn.Socket = args[6].ToString()
	}

	conn.Connected = true
	return values.NewBool(true), nil
}

func mysqliInit(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewNull(), nil
	}

	thisObj := args[0]
	obj, ok := thisObj.Data.(*values.Object)
	if !ok {
		return values.NewNull(), nil
	}

	// Initialize connection object
	conn := &MySQLiConnection{
		Host:      "localhost",
		Username:  "",
		Password:  "",
		Database:  "",
		Port:      3306,
		Socket:    "",
		Connected: false,
	}

	if obj.Properties == nil {
		obj.Properties = make(map[string]*values.Value)
	}
	obj.Properties["__mysqli_connection"] = values.NewResource(conn)

	return values.NewNull(), nil
}

func mysqliClose(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewBool(false), nil
	}

	conn, ok := extractMySQLiConnection(args[0])
	if !ok {
		return values.NewBool(false), nil
	}

	RealMySQLiClose(conn)
	return values.NewBool(true), nil
}

func mysqliChangeUser(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 4 {
		return values.NewBool(false), nil
	}

	conn, ok := extractMySQLiConnection(args[0])
	if !ok {
		return values.NewBool(false), nil
	}

	conn.Username = args[1].ToString()
	conn.Password = args[2].ToString()
	conn.Database = args[3].ToString()

	return values.NewBool(true), nil
}

func mysqliSelectDb(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 2 {
		return values.NewBool(false), nil
	}

	conn, ok := extractMySQLiConnection(args[0])
	if !ok {
		return values.NewBool(false), nil
	}

	conn.Database = args[1].ToString()
	return values.NewBool(true), nil
}

func mysqliPing(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewBool(false), nil
	}

	_, ok := extractMySQLiConnection(args[0])
	if !ok {
		return values.NewBool(false), nil
	}

	return values.NewBool(true), nil
}

// Query Method Implementations

func mysqliQuery(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 2 {
		return values.NewBool(false), nil
	}

	thisObj := args[0]
	conn, ok := extractMySQLiConnection(thisObj)
	if !ok {
		return values.NewBool(false), nil
	}

	query := args[1].ToString()

	// Execute real query
	result, err := RealMySQLiQuery(conn, query)

	// Update mysqli object properties
	if obj, ok := thisObj.Data.(*values.Object); ok {
		if obj.Properties != nil {
			obj.Properties["errno"] = values.NewInt(int64(conn.ErrorNo))
			obj.Properties["error"] = values.NewString(conn.Error)
			obj.Properties["sqlstate"] = values.NewString(conn.SQLState)
			obj.Properties["affected_rows"] = values.NewInt(conn.AffectedRows)
			obj.Properties["insert_id"] = values.NewInt(conn.InsertID)
			obj.Properties["field_count"] = values.NewInt(int64(conn.FieldCount))
			obj.Properties["warning_count"] = values.NewInt(int64(conn.WarningCount))
			if conn.Info != "" {
				obj.Properties["info"] = values.NewString(conn.Info)
			}
		}
	}

	if err != nil {
		return values.NewBool(false), nil
	}

	// For non-SELECT queries (INSERT, UPDATE, DELETE), result is nil
	if result == nil {
		return values.NewBool(true), nil
	}

	return createMySQLiResultObject(ctx, result), nil
}

func mysqliRealQuery(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 2 {
		return values.NewBool(false), nil
	}

	_, ok := extractMySQLiConnection(args[0])
	if !ok {
		return values.NewBool(false), nil
	}

	// Stub: Always return true
	return values.NewBool(true), nil
}

func mysqliMultiQuery(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 2 {
		return values.NewBool(false), nil
	}

	_, ok := extractMySQLiConnection(args[0])
	if !ok {
		return values.NewBool(false), nil
	}

	// Stub: Always return true
	return values.NewBool(true), nil
}

func mysqliPrepare(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 2 {
		return values.NewBool(false), nil
	}

	conn, ok := extractMySQLiConnection(args[0])
	if !ok {
		return values.NewBool(false), nil
	}

	query := args[1].ToString()

	// Count placeholders
	paramCount := 0
	for i := 0; i < len(query); i++ {
		if query[i] == '?' {
			paramCount++
		}
	}

	stmt := &MySQLiStmt{
		Connection: conn,
		Query:      query,
		ParamCount: paramCount,
		FieldCount: 0,
		Params:     make([]*values.Value, paramCount),
	}

	return createMySQLiStmtObject(ctx, stmt), nil
}

func mysqliStoreResult(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewNull(), nil
	}

	_, ok := extractMySQLiConnection(args[0])
	if !ok {
		return values.NewNull(), nil
	}

	// Stub: Return empty result
	result := &MySQLiResult{
		NumRows:    0,
		FieldCount: 0,
		CurrentRow: 0,
		Rows:       make([]map[string]*values.Value, 0),
		Fields:     make([]MySQLiField, 0),
	}
	return createMySQLiResultObject(ctx, result), nil
}

func mysqliUseResult(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewNull(), nil
	}

	_, ok := extractMySQLiConnection(args[0])
	if !ok {
		return values.NewNull(), nil
	}

	// Stub: Return empty result
	result := &MySQLiResult{
		NumRows:    0,
		FieldCount: 0,
		CurrentRow: 0,
		Rows:       make([]map[string]*values.Value, 0),
		Fields:     make([]MySQLiField, 0),
	}
	return createMySQLiResultObject(ctx, result), nil
}

func mysqliMoreResults(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewBool(false), nil
	}

	_, ok := extractMySQLiConnection(args[0])
	if !ok {
		return values.NewBool(false), nil
	}

	return values.NewBool(false), nil
}

func mysqliNextResult(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewBool(false), nil
	}

	_, ok := extractMySQLiConnection(args[0])
	if !ok {
		return values.NewBool(false), nil
	}

	return values.NewBool(false), nil
}

// Transaction Method Implementations

func mysqliAutocommit(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 2 {
		return values.NewBool(false), nil
	}

	_, ok := extractMySQLiConnection(args[0])
	if !ok {
		return values.NewBool(false), nil
	}

	return values.NewBool(true), nil
}

func mysqliBeginTransaction(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewBool(false), nil
	}

	_, ok := extractMySQLiConnection(args[0])
	if !ok {
		return values.NewBool(false), nil
	}

	return values.NewBool(true), nil
}

func mysqliCommit(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewBool(false), nil
	}

	_, ok := extractMySQLiConnection(args[0])
	if !ok {
		return values.NewBool(false), nil
	}

	return values.NewBool(true), nil
}

func mysqliRollback(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewBool(false), nil
	}

	_, ok := extractMySQLiConnection(args[0])
	if !ok {
		return values.NewBool(false), nil
	}

	return values.NewBool(true), nil
}

// Info/Utility Method Implementations

func mysqliGetCharset(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewNull(), nil
	}

	_, ok := extractMySQLiConnection(args[0])
	if !ok {
		return values.NewNull(), nil
	}

	// Stub: Return null
	return values.NewNull(), nil
}

func mysqliCharacterSetName(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewString(""), nil
	}

	_, ok := extractMySQLiConnection(args[0])
	if !ok {
		return values.NewString(""), nil
	}

	return values.NewString("utf8mb4"), nil
}

func mysqliSetCharset(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 2 {
		return values.NewBool(false), nil
	}

	_, ok := extractMySQLiConnection(args[0])
	if !ok {
		return values.NewBool(false), nil
	}

	return values.NewBool(true), nil
}

func mysqliRealEscapeString(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 2 {
		return values.NewString(""), nil
	}

	_, ok := extractMySQLiConnection(args[0])
	if !ok {
		return values.NewString(""), nil
	}

	str := args[1].ToString()
	str = mysqliReplaceAll(str, "\\", "\\\\")
	str = mysqliReplaceAll(str, "'", "\\'")
	str = mysqliReplaceAll(str, "\"", "\\\"")
	str = mysqliReplaceAll(str, "\n", "\\n")
	str = mysqliReplaceAll(str, "\r", "\\r")
	str = mysqliReplaceAll(str, "\x00", "\\0")
	str = mysqliReplaceAll(str, "\x1a", "\\Z")

	return values.NewString(str), nil
}

func mysqliEscapeString(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	// Alias for real_escape_string
	return mysqliRealEscapeString(ctx, args)
}

func mysqliOptions(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 3 {
		return values.NewBool(false), nil
	}

	_, ok := extractMySQLiConnection(args[0])
	if !ok {
		return values.NewBool(false), nil
	}

	return values.NewBool(true), nil
}

func mysqliStat(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewString(""), nil
	}

	_, ok := extractMySQLiConnection(args[0])
	if !ok {
		return values.NewString(""), nil
	}

	return values.NewString("Uptime: 0  Threads: 1  Questions: 0  Slow queries: 0  Opens: 0  Flush tables: 0  Open tables: 0"), nil
}

func mysqliThreadId(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewInt(0), nil
	}

	_, ok := extractMySQLiConnection(args[0])
	if !ok {
		return values.NewInt(0), nil
	}

	return values.NewInt(1), nil
}

// Error/Info Getter Implementations

func mysqliErrno(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewInt(0), nil
	}

	conn, ok := extractMySQLiConnection(args[0])
	if !ok {
		return values.NewInt(0), nil
	}

	return values.NewInt(int64(conn.ErrorNo)), nil
}

func mysqliError(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewString(""), nil
	}

	conn, ok := extractMySQLiConnection(args[0])
	if !ok {
		return values.NewString(""), nil
	}

	return values.NewString(conn.Error), nil
}

func mysqliErrorList(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewArray(), nil
	}

	_, ok := extractMySQLiConnection(args[0])
	if !ok {
		return values.NewArray(), nil
	}

	return values.NewArray(), nil
}

func mysqliSqlstate(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewString("00000"), nil
	}

	conn, ok := extractMySQLiConnection(args[0])
	if !ok {
		return values.NewString("00000"), nil
	}

	return values.NewString(conn.SQLState), nil
}

func mysqliInfo(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewNull(), nil
	}

	conn, ok := extractMySQLiConnection(args[0])
	if !ok {
		return values.NewNull(), nil
	}

	return values.NewString(conn.Info), nil
}

func mysqliWarningCount(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewInt(0), nil
	}

	conn, ok := extractMySQLiConnection(args[0])
	if !ok {
		return values.NewInt(0), nil
	}

	return values.NewInt(int64(conn.WarningCount)), nil
}

// Result/Statement Getter Implementations

func mysqliAffectedRows(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewInt(-1), nil
	}

	conn, ok := extractMySQLiConnection(args[0])
	if !ok {
		return values.NewInt(-1), nil
	}

	return values.NewInt(conn.AffectedRows), nil
}

func mysqliInsertId(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewInt(0), nil
	}

	conn, ok := extractMySQLiConnection(args[0])
	if !ok {
		return values.NewInt(0), nil
	}

	return values.NewInt(conn.InsertID), nil
}

func mysqliFieldCount(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewInt(0), nil
	}

	conn, ok := extractMySQLiConnection(args[0])
	if !ok {
		return values.NewInt(0), nil
	}

	return values.NewInt(int64(conn.FieldCount)), nil
}

// Server Info Method Implementations

func mysqliGetClientInfo(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	return values.NewString("hey-mysqli-stub-1.0.0"), nil
}

func mysqliGetHostInfo(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return values.NewString(""), nil
	}

	conn, ok := extractMySQLiConnection(args[0])
	if !ok {
		return values.NewString("localhost via TCP/IP"), nil
	}

	return values.NewString(conn.Host + " via TCP/IP"), nil
}

func mysqliGetProtoInfo(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	return values.NewInt(10), nil
}

func mysqliGetServerInfo(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	return values.NewString("8.0.30"), nil
}

func mysqliGetServerVersion(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	return values.NewInt(80030), nil
}

// Advanced/Debug Method Stub Implementations

func mysqliGetConnectionStats(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	return values.NewArray(), nil
}

func mysqliGetWarnings(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	return values.NewBool(false), nil
}

func mysqliPoll(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	return values.NewInt(0), nil
}

func mysqliReapAsyncQuery(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	return values.NewBool(false), nil
}

func mysqliRefresh(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	return values.NewBool(true), nil
}

func mysqliSslSet(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	return values.NewBool(true), nil
}

func mysqliDumpDebugInfo(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	return values.NewBool(true), nil
}

func mysqliDebug(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	return values.NewBool(true), nil
}

func mysqliKill(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	return values.NewBool(true), nil
}
