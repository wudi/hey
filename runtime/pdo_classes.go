package runtime

import (
	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetPDOClassDescriptors returns PDO class descriptors for registration
func GetPDOClassDescriptors() []*registry.ClassDescriptor {
	return []*registry.ClassDescriptor{
		{
			Name:       "PDO",
			Parent:     "",
			Interfaces: []string{},
			Traits:     []string{},
			Properties: make(map[string]*registry.PropertyDescriptor),
			Methods:    pdoMethodDescriptors(),
			Constants:  pdoConstantDescriptors(),
			IsAbstract: false,
			IsFinal:    false,
		},
		{
			Name:       "PDOStatement",
			Parent:     "",
			Interfaces: []string{},
			Traits:     []string{},
			Properties: pdoStatementPropertyDescriptors(),
			Methods:    pdoStatementMethodDescriptors(),
			Constants:  make(map[string]*registry.ConstantDescriptor),
			IsAbstract: false,
			IsFinal:    false,
		},
		{
			Name:       "PDOException",
			Parent:     "Exception",
			Interfaces: []string{},
			Traits:     []string{},
			Properties: map[string]*registry.PropertyDescriptor{
				"errorInfo": {
					Name:         "errorInfo",
					Visibility:   "public",
					IsStatic:     false,
					DefaultValue: values.NewNull(),
				},
			},
			Methods:    make(map[string]*registry.MethodDescriptor),
			Constants:  make(map[string]*registry.ConstantDescriptor),
			IsAbstract: false,
			IsFinal:    false,
		},
	}
}

// pdoMethodDescriptors returns method descriptors for PDO class
func pdoMethodDescriptors() map[string]*registry.MethodDescriptor {
	return map[string]*registry.MethodDescriptor{
		"__construct": newPDOMethod("__construct",
			[]registry.ParameterDescriptor{
				{Name: "dsn", Type: "string"},
				{Name: "username", Type: "string", HasDefault: true, DefaultValue: values.NewNull()},
				{Name: "password", Type: "string", HasDefault: true, DefaultValue: values.NewNull()},
				{Name: "options", Type: "array", HasDefault: true, DefaultValue: values.NewNull()},
			},
			"", pdoConstruct),
		"prepare": newPDOMethod("prepare",
			[]registry.ParameterDescriptor{
				{Name: "query", Type: "string"},
				{Name: "options", Type: "array", HasDefault: true, DefaultValue: values.NewArray()},
			},
			"PDOStatement|false", pdoPrepare),
		"query": newPDOMethod("query",
			[]registry.ParameterDescriptor{
				{Name: "query", Type: "string"},
			},
			"PDOStatement|false", pdoQuery),
		"exec": newPDOMethod("exec",
			[]registry.ParameterDescriptor{
				{Name: "statement", Type: "string"},
			},
			"int|false", pdoExec),
		"lastInsertId": newPDOMethod("lastInsertId",
			[]registry.ParameterDescriptor{
				{Name: "name", Type: "string", HasDefault: true, DefaultValue: values.NewNull()},
			},
			"string|false", pdoLastInsertId),
		"beginTransaction": newPDOMethod("beginTransaction", []registry.ParameterDescriptor{}, "bool", pdoBeginTransaction),
		"commit":           newPDOMethod("commit", []registry.ParameterDescriptor{}, "bool", pdoCommit),
		"rollBack":         newPDOMethod("rollBack", []registry.ParameterDescriptor{}, "bool", pdoRollBack),
		"inTransaction":    newPDOMethod("inTransaction", []registry.ParameterDescriptor{}, "bool", pdoInTransaction),
		"getAttribute": newPDOMethod("getAttribute",
			[]registry.ParameterDescriptor{
				{Name: "attribute", Type: "int"},
			},
			"mixed", pdoGetAttribute),
		"setAttribute": newPDOMethod("setAttribute",
			[]registry.ParameterDescriptor{
				{Name: "attribute", Type: "int"},
				{Name: "value", Type: "mixed"},
			},
			"bool", pdoSetAttribute),
		"errorCode": newPDOMethod("errorCode", []registry.ParameterDescriptor{}, "string|null", pdoErrorCode),
		"errorInfo": newPDOMethod("errorInfo", []registry.ParameterDescriptor{}, "array", pdoErrorInfo),
	}
}

// pdoStatementPropertyDescriptors returns property descriptors for PDOStatement class
func pdoStatementPropertyDescriptors() map[string]*registry.PropertyDescriptor {
	return map[string]*registry.PropertyDescriptor{
		"queryString": {
			Name:         "queryString",
			Visibility:   "public",
			IsStatic:     false,
			DefaultValue: values.NewString(""),
		},
	}
}

// pdoStatementMethodDescriptors returns method descriptors for PDOStatement class
func pdoStatementMethodDescriptors() map[string]*registry.MethodDescriptor {
	return map[string]*registry.MethodDescriptor{
		"execute": newPDOMethod("execute",
			[]registry.ParameterDescriptor{
				{Name: "params", Type: "array", HasDefault: true, DefaultValue: values.NewNull()},
			},
			"bool", pdoStmtExecute),
		"fetch": newPDOMethod("fetch",
			[]registry.ParameterDescriptor{
				{Name: "mode", Type: "int", HasDefault: true, DefaultValue: values.NewInt(4)},
			},
			"mixed", pdoStmtFetch),
		"fetchAll": newPDOMethod("fetchAll",
			[]registry.ParameterDescriptor{
				{Name: "mode", Type: "int", HasDefault: true, DefaultValue: values.NewInt(4)},
			},
			"array", pdoStmtFetchAll),
		"fetchColumn": newPDOMethod("fetchColumn",
			[]registry.ParameterDescriptor{
				{Name: "column", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
			},
			"mixed", pdoStmtFetchColumn),
		"rowCount":    newPDOMethod("rowCount", []registry.ParameterDescriptor{}, "int", pdoStmtRowCount),
		"columnCount": newPDOMethod("columnCount", []registry.ParameterDescriptor{}, "int", pdoStmtColumnCount),
		"bindValue": newPDOMethod("bindValue",
			[]registry.ParameterDescriptor{
				{Name: "param", Type: "mixed"},
				{Name: "value", Type: "mixed"},
				{Name: "type", Type: "int", HasDefault: true, DefaultValue: values.NewInt(2)},
			},
			"bool", pdoStmtBindValue),
		"bindParam": newPDOMethod("bindParam",
			[]registry.ParameterDescriptor{
				{Name: "param", Type: "mixed"},
				{Name: "var", Type: "mixed"},
				{Name: "type", Type: "int", HasDefault: true, DefaultValue: values.NewInt(2)},
			},
			"bool", pdoStmtBindParam),
		"closeCursor": newPDOMethod("closeCursor", []registry.ParameterDescriptor{}, "bool", pdoStmtCloseCursor),
		"errorCode":   newPDOMethod("errorCode", []registry.ParameterDescriptor{}, "string|null", pdoStmtErrorCode),
		"errorInfo":   newPDOMethod("errorInfo", []registry.ParameterDescriptor{}, "array", pdoStmtErrorInfo),
	}
}

// pdoConstantDescriptors returns constant descriptors for PDO class
func pdoConstantDescriptors() map[string]*registry.ConstantDescriptor {
	constants := make(map[string]*registry.ConstantDescriptor)
	pdoConstants := getPDOConstants()
	for name, value := range pdoConstants {
		constants[name] = &registry.ConstantDescriptor{
			Name:  name,
			Value: value,
		}
	}
	return constants
}
