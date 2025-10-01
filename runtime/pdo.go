package runtime

import (
	"fmt"

	"github.com/wudi/hey/pkg/pdo"
	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// pdoConstruct implements new PDO($dsn, $username, $password, $options)
// args[0] = $this, args[1] = dsn, args[2] = username, args[3] = password, args[4] = options
func pdoConstruct(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("PDO::__construct() expects at least 1 parameter")
	}

	thisObj := args[0]
	dsn := args[1].Data.(string)
	username := ""
	password := ""

	if len(args) > 2 && args[2].Type != values.TypeNull {
		username = args[2].Data.(string)
	}
	if len(args) > 3 && args[3].Type != values.TypeNull {
		password = args[3].Data.(string)
	}

	// Parse DSN to get driver
	dsnInfo, err := pdo.ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("invalid DSN: %v", err)
	}

	// Get driver
	driver, ok := pdo.GetDriver(dsnInfo.Driver)
	if !ok {
		return nil, fmt.Errorf("could not find driver: %s", dsnInfo.Driver)
	}

	// Open connection
	conn, err := driver.Open(dsn)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %v", err)
	}

	// Driver-specific connection setup
	switch dsnInfo.Driver {
	case "mysql":
		if mysqlConn, ok := conn.(*pdo.MySQLConn); ok {
			if err := mysqlConn.Connect(username, password); err != nil {
				return nil, fmt.Errorf("connection failed: %v", err)
			}
		}
	case "sqlite":
		if sqliteConn, ok := conn.(*pdo.SQLiteConn); ok {
			if err := sqliteConn.Connect(); err != nil {
				return nil, fmt.Errorf("connection failed: %v", err)
			}
		}
	case "pgsql":
		if pgsqlConn, ok := conn.(*pdo.PgSQLConn); ok {
			if err := pgsqlConn.Connect(username, password); err != nil {
				return nil, fmt.Errorf("connection failed: %v", err)
			}
		}
	}

	// Store connection in object properties
	obj := thisObj.Data.(*values.Object)
	if obj.Properties == nil {
		obj.Properties = make(map[string]*values.Value)
	}

	obj.Properties["__pdo_conn"] = values.NewResource(conn)
	obj.Properties["__pdo_driver"] = values.NewString(dsnInfo.Driver)
	obj.Properties["__pdo_in_tx"] = values.NewBool(false)
	obj.Properties["__pdo_tx"] = values.NewNull()

	return values.NewNull(), nil
}

// pdoPrepare implements $pdo->prepare($query)
// args[0] = $this, args[1] = query, args[2] = options
func pdoPrepare(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 2 {
		return values.NewBool(false), fmt.Errorf("PDO::prepare() expects at least 1 parameter")
	}

	thisObj := args[0]
	query := args[1].Data.(string)

	// Get connection from object properties
	obj := thisObj.Data.(*values.Object)
	connVal, ok := obj.Properties["__pdo_conn"]
	if !ok {
		return values.NewBool(false), fmt.Errorf("invalid PDO object: no connection")
	}

	conn := connVal.Data.(pdo.Conn)

	stmt, err := conn.Prepare(query)
	if err != nil {
		return values.NewBool(false), nil
	}

	// Create PDOStatement object
	stmtObj := values.NewObject("PDOStatement")
	stmtObjData := stmtObj.Data.(*values.Object)
	stmtObjData.Properties["queryString"] = values.NewString(query)
	stmtObjData.Properties["__pdo_stmt"] = values.NewResource(stmt)
	stmtObjData.Properties["__pdo_rows"] = values.NewNull()

	return stmtObj, nil
}

// pdoQuery implements $pdo->query($query)
// args[0] = $this, args[1] = query
func pdoQuery(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 2 {
		return values.NewBool(false), fmt.Errorf("PDO::query() expects at least 1 parameter")
	}

	thisObj := args[0]
	query := args[1].Data.(string)

	// Get connection
	obj := thisObj.Data.(*values.Object)

	var rows pdo.Rows
	var err error

	// Check if we're in a transaction
	inTxVal, hasInTx := obj.Properties["__pdo_in_tx"]
	inTx := hasInTx && inTxVal.Type == values.TypeBool && inTxVal.Data.(bool)

	if inTx {
		txVal := obj.Properties["__pdo_tx"]
		if txVal.Type != values.TypeNull {
			tx := txVal.Data.(pdo.Tx)
			rows, err = tx.Query(query)
		} else {
			return values.NewBool(false), fmt.Errorf("invalid transaction state")
		}
	} else {
		connVal, ok := obj.Properties["__pdo_conn"]
		if !ok {
			return values.NewBool(false), fmt.Errorf("invalid PDO object")
		}
		conn := connVal.Data.(pdo.Conn)
		rows, err = conn.Query(query)
	}

	if err != nil {
		return values.NewBool(false), nil
	}

	if rows == nil {
		return values.NewBool(false), fmt.Errorf("query failed to return rows")
	}

	// Create PDOStatement object with active result set
	stmtObj := values.NewObject("PDOStatement")
	stmtObjData := stmtObj.Data.(*values.Object)
	stmtObjData.Properties["queryString"] = values.NewString(query)
	stmtObjData.Properties["__pdo_rows"] = values.NewResource(rows)

	return stmtObj, nil
}

// pdoExec implements $pdo->exec($statement)
// args[0] = $this, args[1] = statement
func pdoExec(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 2 {
		return values.NewBool(false), fmt.Errorf("PDO::exec() expects 1 parameter")
	}

	thisObj := args[0]
	query := args[1].Data.(string)

	// Get connection
	obj := thisObj.Data.(*values.Object)

	var result pdo.Result
	var err error

	// Check if we're in a transaction
	inTxVal, hasInTx := obj.Properties["__pdo_in_tx"]
	inTx := hasInTx && inTxVal.Type == values.TypeBool && inTxVal.Data.(bool)

	if inTx {
		txVal := obj.Properties["__pdo_tx"]
		if txVal.Type != values.TypeNull {
			tx := txVal.Data.(pdo.Tx)
			result, err = tx.Exec(query)
		} else {
			return values.NewBool(false), fmt.Errorf("invalid transaction state")
		}
	} else {
		connVal, ok := obj.Properties["__pdo_conn"]
		if !ok {
			return values.NewBool(false), fmt.Errorf("invalid PDO object")
		}
		conn := connVal.Data.(pdo.Conn)
		result, err = conn.Exec(query)
	}

	if err != nil {
		return values.NewBool(false), nil
	}

	if result == nil {
		return values.NewBool(false), fmt.Errorf("exec failed to return result")
	}

	affected, _ := result.RowsAffected()
	return values.NewInt(affected), nil
}

// pdoLastInsertId implements $pdo->lastInsertId()
// args[0] = $this, args[1] = name (optional)
func pdoLastInsertId(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	thisObj := args[0]

	// Get connection
	obj := thisObj.Data.(*values.Object)
	connVal, ok := obj.Properties["__pdo_conn"]
	if !ok {
		return values.NewString("0"), fmt.Errorf("invalid PDO object")
	}

	conn := connVal.Data.(pdo.Conn)

	id, err := conn.LastInsertId()
	if err != nil {
		return values.NewString("0"), nil
	}

	return values.NewString(fmt.Sprintf("%d", id)), nil
}

// pdoBeginTransaction implements $pdo->beginTransaction()
// args[0] = $this
func pdoBeginTransaction(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	thisObj := args[0]

	// Get connection
	obj := thisObj.Data.(*values.Object)
	connVal, ok := obj.Properties["__pdo_conn"]
	if !ok {
		return values.NewBool(false), fmt.Errorf("invalid PDO object")
	}

	// Check if already in transaction
	if obj.Properties["__pdo_in_tx"].Data.(bool) {
		return values.NewBool(false), fmt.Errorf("already in transaction")
	}

	conn := connVal.Data.(pdo.Conn)

	tx, err := conn.Begin()
	if err != nil {
		return values.NewBool(false), nil
	}

	obj.Properties["__pdo_tx"] = values.NewResource(tx)
	obj.Properties["__pdo_in_tx"] = values.NewBool(true)

	return values.NewBool(true), nil
}

// pdoCommit implements $pdo->commit()
// args[0] = $this
func pdoCommit(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	thisObj := args[0]

	// Get transaction
	obj := thisObj.Data.(*values.Object)

	if !obj.Properties["__pdo_in_tx"].Data.(bool) {
		return values.NewBool(false), fmt.Errorf("no active transaction")
	}

	txVal := obj.Properties["__pdo_tx"]
	if txVal.Type == values.TypeNull {
		return values.NewBool(false), fmt.Errorf("no active transaction")
	}

	tx := txVal.Data.(pdo.Tx)

	if err := tx.Commit(); err != nil {
		return values.NewBool(false), nil
	}

	obj.Properties["__pdo_in_tx"] = values.NewBool(false)
	obj.Properties["__pdo_tx"] = values.NewNull()

	return values.NewBool(true), nil
}

// pdoRollBack implements $pdo->rollBack()
// args[0] = $this
func pdoRollBack(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	thisObj := args[0]

	// Get transaction
	obj := thisObj.Data.(*values.Object)

	if !obj.Properties["__pdo_in_tx"].Data.(bool) {
		return values.NewBool(false), fmt.Errorf("no active transaction")
	}

	txVal := obj.Properties["__pdo_tx"]
	if txVal.Type == values.TypeNull {
		return values.NewBool(false), fmt.Errorf("no active transaction")
	}

	tx := txVal.Data.(pdo.Tx)

	if err := tx.Rollback(); err != nil {
		return values.NewBool(false), nil
	}

	obj.Properties["__pdo_in_tx"] = values.NewBool(false)
	obj.Properties["__pdo_tx"] = values.NewNull()

	return values.NewBool(true), nil
}

// pdoInTransaction implements $pdo->inTransaction()
// args[0] = $this
func pdoInTransaction(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	thisObj := args[0]

	obj := thisObj.Data.(*values.Object)
	inTx := obj.Properties["__pdo_in_tx"]

	if inTx == nil {
		return values.NewBool(false), nil
	}

	return values.NewBool(inTx.Data.(bool)), nil
}

// Placeholder implementations for remaining methods
func pdoGetAttribute(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	return values.NewNull(), nil
}

func pdoSetAttribute(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	return values.NewBool(true), nil
}

func pdoErrorCode(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	return values.NewNull(), nil
}

func pdoErrorInfo(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	return values.NewArray(), nil
}
