package runtime

import (
	"github.com/wudi/hey/registry"
)

// GetMySQLiClasses returns simplified MySQLi class definitions
// Full OOP implementation would require more extensive class system support
func GetMySQLiClasses() []*registry.ClassDescriptor {
	return []*registry.ClassDescriptor{
		// mysqli class - Main database connection class
		{
			Name:       "mysqli",
			Parent:     "",
			Interfaces: []string{},
			Traits:     []string{},
			Properties: mysqliPropertyDescriptors(),
			Methods:    mysqliMethodDescriptors(),
			Constants:  make(map[string]*registry.ConstantDescriptor),
			IsAbstract: false,
			IsFinal:    false,
		},
		// mysqli_result class - Result set class
		{
			Name:       "mysqli_result",
			Parent:     "",
			Interfaces: []string{},
			Traits:     []string{},
			Properties: mysqliResultPropertyDescriptors(),
			Methods:    mysqliResultMethodDescriptors(),
			Constants:  make(map[string]*registry.ConstantDescriptor),
			IsAbstract: false,
			IsFinal:    false,
		},
		// mysqli_stmt class - Prepared statement class
		{
			Name:       "mysqli_stmt",
			Parent:     "",
			Interfaces: []string{},
			Traits:     []string{},
			Properties: mysqliStmtPropertyDescriptors(),
			Methods:    mysqliStmtMethodDescriptors(),
			Constants:  make(map[string]*registry.ConstantDescriptor),
			IsAbstract: false,
			IsFinal:    false,
		},
		// mysqli_driver class - Driver information
		{
			Name:       "mysqli_driver",
			Parent:     "",
			Interfaces: []string{},
			Traits:     []string{},
			Properties: make(map[string]*registry.PropertyDescriptor),
			Methods:    make(map[string]*registry.MethodDescriptor),
			Constants:  make(map[string]*registry.ConstantDescriptor),
			IsAbstract: false,
			IsFinal:    false,
		},
		// mysqli_warning class - Warning information
		{
			Name:       "mysqli_warning",
			Parent:     "",
			Interfaces: []string{},
			Traits:     []string{},
			Properties: make(map[string]*registry.PropertyDescriptor),
			Methods:    make(map[string]*registry.MethodDescriptor),
			Constants:  make(map[string]*registry.ConstantDescriptor),
			IsAbstract: false,
			IsFinal:    false,
		},
		// mysqli_sql_exception class - Exception class
		{
			Name:       "mysqli_sql_exception",
			Parent:     "RuntimeException",
			Interfaces: []string{},
			Traits:     []string{},
			Properties: make(map[string]*registry.PropertyDescriptor),
			Methods:    make(map[string]*registry.MethodDescriptor),
			Constants:  make(map[string]*registry.ConstantDescriptor),
			IsAbstract: false,
			IsFinal:    false,
		},
	}
}
