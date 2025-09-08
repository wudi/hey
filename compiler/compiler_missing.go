package compiler

import (
	"fmt"
	"strings"

	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/compiler/opcodes"
	"github.com/wudi/php-parser/compiler/values"
	"github.com/wudi/php-parser/compiler/vm"
	"github.com/wudi/php-parser/lexer"
)

// PrintStatement compilation - prints value and discards return value
func (c *Compiler) compilePrintStatement(stmt *ast.PrintStatement) error {
	// PrintStatement has Arguments, not Expression
	if stmt.Arguments != nil && len(stmt.Arguments.Arguments) > 0 {
		// Compile the first argument to print
		if err := c.compileNode(stmt.Arguments.Arguments[0]); err != nil {
			return err
		}

		// Emit ECHO instruction to print the value
		arg := c.nextTemp - 1
		c.emit(opcodes.OP_ECHO,
			opcodes.IS_TMP_VAR, arg,
			opcodes.IS_UNUSED, 0,
			opcodes.IS_UNUSED, 0)
	}

	// Print always returns 1 - allocate temp for result
	result := c.allocateTemp()
	c.emit(opcodes.OP_QM_ASSIGN,
		opcodes.IS_CONST, c.addConstant(values.NewInt(1)),
		opcodes.IS_UNUSED, 0,
		opcodes.IS_TMP_VAR, result)

	return nil
}

// PrintExpression compilation - prints value and returns 1
func (c *Compiler) compilePrintExpression(expr *ast.PrintExpression) error {
	// Compile the expression to print
	if err := c.compileNode(expr.Expression); err != nil {
		return err
	}

	// Emit ECHO instruction to print the value
	arg := c.nextTemp - 1
	c.emit(opcodes.OP_ECHO,
		opcodes.IS_TMP_VAR, arg,
		opcodes.IS_UNUSED, 0,
		opcodes.IS_UNUSED, 0)

	// Print always returns 1
	result := c.allocateTemp()
	c.emit(opcodes.OP_QM_ASSIGN,
		opcodes.IS_CONST, c.addConstant(values.NewInt(1)),
		opcodes.IS_UNUSED, 0,
		opcodes.IS_TMP_VAR, result)

	return nil
}

// CloneExpression compilation
func (c *Compiler) compileCloneExpression(expr *ast.CloneExpression) error {
	// Compile the object to clone
	if err := c.compileNode(expr.Object); err != nil {
		return err
	}

	// Emit CLONE instruction
	result := c.allocateTemp()
	arg := c.nextTemp - 1
	c.emit(opcodes.OP_CLONE,
		opcodes.IS_TMP_VAR, arg,
		opcodes.IS_UNUSED, 0,
		opcodes.IS_TMP_VAR, result)

	return nil
}

// InstanceofExpression compilation
func (c *Compiler) compileInstanceofExpression(expr *ast.InstanceofExpression) error {
	// Compile the left expression (object)
	if err := c.compileNode(expr.Left); err != nil {
		return err
	}

	leftOperand := c.nextTemp - 1

	// Compile the right expression (class)
	if err := c.compileNode(expr.Right); err != nil {
		return err
	}

	rightOperand := c.nextTemp - 1
	result := c.allocateTemp()

	// Emit INSTANCEOF instruction
	c.emit(opcodes.OP_INSTANCEOF,
		opcodes.IS_TMP_VAR, leftOperand,
		opcodes.IS_TMP_VAR, rightOperand,
		opcodes.IS_TMP_VAR, result)

	return nil
}

// CastExpression compilation
func (c *Compiler) compileCastExpression(expr *ast.CastExpression) error {
	// Compile the operand to cast
	if err := c.compileNode(expr.Operand); err != nil {
		return err
	}

	arg := c.nextTemp - 1
	result := c.allocateTemp()

	// Use specific cast opcodes
	var castOp opcodes.Opcode
	switch expr.CastType {
	case "(int)", "(integer)":
		castOp = opcodes.OP_CAST_LONG
	case "(bool)", "(boolean)":
		castOp = opcodes.OP_CAST_BOOL
	case "(float)", "(double)", "(real)":
		castOp = opcodes.OP_CAST_DOUBLE
	case "(string)":
		castOp = opcodes.OP_CAST_STRING
	case "(array)":
		castOp = opcodes.OP_CAST_ARRAY
	case "(object)":
		castOp = opcodes.OP_CAST_OBJECT
	case "(unset)":
		// For (unset) cast, assign null directly
		c.emit(opcodes.OP_QM_ASSIGN,
			opcodes.IS_CONST, c.addConstant(values.NewNull()),
			opcodes.IS_UNUSED, 0,
			opcodes.IS_TMP_VAR, result)
		return nil
	default:
		return fmt.Errorf("unsupported cast type: %s", expr.CastType)
	}

	// Emit specific cast instruction
	c.emit(castOp,
		opcodes.IS_TMP_VAR, arg,
		opcodes.IS_UNUSED, 0,
		opcodes.IS_TMP_VAR, result)

	return nil
}

// ErrorSuppressionExpression compilation (@expr)
func (c *Compiler) compileErrorSuppressionExpression(expr *ast.ErrorSuppressionExpression) error {
	// For now, just compile the inner expression
	// In a full implementation, we'd emit BEGIN_SILENCE/END_SILENCE opcodes
	return c.compileNode(expr.Expression)
}

// EmptyExpression compilation
func (c *Compiler) compileEmptyExpression(expr *ast.EmptyExpression) error {
	// Compile the expression to check
	if err := c.compileNode(expr.Expression); err != nil {
		return err
	}

	arg := c.nextTemp - 1
	result := c.allocateTemp()

	// Emit ISSET_ISEMPTY_VAR instruction (extended_value=1 for empty check)
	c.emit(opcodes.OP_ISSET_ISEMPTY_VAR,
		opcodes.IS_TMP_VAR, arg,
		opcodes.IS_UNUSED, 0,
		opcodes.IS_TMP_VAR, result)

	return nil
}

// ExitExpression compilation (exit/die)
func (c *Compiler) compileExitExpression(expr *ast.ExitExpression) error {
	var exitCode uint32 = 0
	var exitType opcodes.OpType = opcodes.IS_CONST

	// If there's an argument, compile it as the exit code
	if expr.Argument != nil {
		if err := c.compileNode(expr.Argument); err != nil {
			return err
		}
		exitCode = c.nextTemp - 1
		exitType = opcodes.IS_TMP_VAR
	} else {
		// Default exit code is 0
		exitCode = c.addConstant(values.NewInt(0))
	}

	// Emit EXIT instruction
	c.emit(opcodes.OP_EXIT,
		exitType, exitCode,
		opcodes.IS_UNUSED, 0,
		opcodes.IS_UNUSED, 0)

	return nil
}

// IssetExpression compilation
func (c *Compiler) compileIssetExpression(expr *ast.IssetExpression) error {
	result := c.allocateTemp()

	// If no arguments, isset() returns true
	if expr.Arguments == nil || len(expr.Arguments.Arguments) == 0 {
		c.emit(opcodes.OP_QM_ASSIGN,
			opcodes.IS_CONST, c.addConstant(values.NewBool(true)),
			opcodes.IS_UNUSED, 0,
			opcodes.IS_TMP_VAR, result)
		return nil
	}

	// For each argument, check if it's set WITHOUT evaluating it
	for i, e := range expr.Arguments.Arguments {
		tempResult := c.allocateTemp()

		if err := c.compileIssetVariable(e, tempResult); err != nil {
			return err
		}

		// If this is not the first expression, AND with previous result
		if i > 0 {
			prevResult := result
			result = c.allocateTemp()
			c.emit(opcodes.OP_BOOLEAN_AND,
				opcodes.IS_TMP_VAR, prevResult,
				opcodes.IS_TMP_VAR, tempResult,
				opcodes.IS_TMP_VAR, result)
		} else {
			result = tempResult
		}
	}

	// CRITICAL: Set nextTemp to point after our result so that subsequent
	// operations see our result as the "most recent temporary"
	c.nextTemp = result + 1

	return nil
}

// compileIssetVariable handles isset() for different types of variables
func (c *Compiler) compileIssetVariable(expr ast.Expression, result uint32) error {
	switch v := expr.(type) {
	case *ast.Variable:
		// Simple variable: isset($var) - check if variable exists in symbol table
		varSlot := c.getVariableSlot(v.Name)
		c.emit(opcodes.OP_ISSET_ISEMPTY_VAR,
			opcodes.IS_CV, varSlot,
			opcodes.IS_UNUSED, 0,
			opcodes.IS_TMP_VAR, result)
		return nil

	case *ast.ArrayAccessExpression:
		// Array element: isset($arr[key]) - check if array element exists

		// Handle the index first
		if v.Index == nil {
			// isset($arr[]) is invalid
			return fmt.Errorf("cannot use [] for isset check")
		}

		// Compile the index expression
		if err := c.compileNode(*v.Index); err != nil {
			return err
		}
		indexOperand := c.nextTemp - 1
		indexType := opcodes.IS_TMP_VAR

		// Handle the array - for isset, we need to work directly with the variable, not a copy
		var arrayOperand uint32
		var arrayType opcodes.OpType

		if variable, ok := v.Array.(*ast.Variable); ok {
			// Direct variable access: isset($arr[key])
			arrayOperand = c.getVariableSlot(variable.Name)
			arrayType = opcodes.IS_CV
		} else {
			// Complex expression: isset($obj->arr[key]) - compile and use temporary
			if err := c.compileNode(v.Array); err != nil {
				return err
			}
			arrayOperand = c.nextTemp - 1
			arrayType = opcodes.IS_TMP_VAR
		}

		// Emit FETCH_DIM_IS instruction for isset check
		c.emit(opcodes.OP_FETCH_DIM_IS,
			arrayType, arrayOperand,
			indexType, indexOperand,
			opcodes.IS_TMP_VAR, result)
		return nil

	case *ast.PropertyAccessExpression:
		// Object property: isset($obj->prop) - check if object property exists
		// Compile the object expression
		if err := c.compileNode(v.Object); err != nil {
			return err
		}
		objectOperand := c.nextTemp - 1

		// Handle the property
		var propOperand uint32
		var propType opcodes.OpType

		switch prop := v.Property.(type) {
		case *ast.IdentifierNode:
			// Static property name
			propOperand = c.addConstant(values.NewString(prop.Name))
			propType = opcodes.IS_CONST
		default:
			// Dynamic property name
			if err := c.compileNode(prop); err != nil {
				return err
			}
			propOperand = c.nextTemp - 1
			propType = opcodes.IS_TMP_VAR
		}

		// Emit FETCH_OBJ_IS instruction for isset check
		c.emit(opcodes.OP_FETCH_OBJ_IS,
			opcodes.IS_TMP_VAR, objectOperand,
			propType, propOperand,
			opcodes.IS_TMP_VAR, result)
		return nil

	case *ast.StaticPropertyAccessExpression:
		// Static property: isset(Class::$prop) - check if static property exists
		// Compile class name
		var classOperand uint32
		var classType opcodes.OpType

		switch class := v.Class.(type) {
		case *ast.IdentifierNode:
			// Static class name
			classOperand = c.addConstant(values.NewString(class.Name))
			classType = opcodes.IS_CONST
		default:
			// Dynamic class name
			if err := c.compileNode(class); err != nil {
				return err
			}
			classOperand = c.nextTemp - 1
			classType = opcodes.IS_TMP_VAR
		}

		// Compile property name
		var propOperand uint32
		var propType opcodes.OpType

		switch prop := v.Property.(type) {
		case *ast.Variable:
			// Remove the $ prefix if present
			propName := prop.Name
			if strings.HasPrefix(propName, "$") {
				propName = propName[1:]
			}
			propOperand = c.addConstant(values.NewString(propName))
			propType = opcodes.IS_CONST
		default:
			if err := c.compileNode(prop); err != nil {
				return err
			}
			propOperand = c.nextTemp - 1
			propType = opcodes.IS_TMP_VAR
		}

		// Emit FETCH_STATIC_PROP_IS instruction for isset check
		c.emit(opcodes.OP_FETCH_STATIC_PROP_IS,
			classType, classOperand,
			propType, propOperand,
			opcodes.IS_TMP_VAR, result)
		return nil

	default:
		return fmt.Errorf("isset() can only be used with variables, array elements, or object properties")
	}
}

// ListExpression compilation (array destructuring)
func (c *Compiler) compileListExpression(expr *ast.ListExpression) error {
	// list() is typically used in assignment context
	// For now, we'll handle it as a pass-through
	// Full implementation would involve FETCH_LIST_R opcodes
	return fmt.Errorf("list expressions not yet fully implemented")
}

// EvalExpression compilation
func (c *Compiler) compileEvalExpression(expr *ast.EvalExpression) error {
	// Compile the code to evaluate
	if err := c.compileNode(expr.Argument); err != nil {
		return err
	}

	arg := c.nextTemp - 1
	result := c.allocateTemp()

	// Emit EVAL instruction
	c.emit(opcodes.OP_EVAL,
		opcodes.IS_TMP_VAR, arg,
		opcodes.IS_UNUSED, 0,
		opcodes.IS_TMP_VAR, result)

	return nil
}

// YieldExpression compilation
func (c *Compiler) compileYieldExpression(expr *ast.YieldExpression) error {
	var keyOperand uint32 = 0
	var keyType opcodes.OpType = opcodes.IS_UNUSED
	var valueOperand uint32 = 0
	var valueType opcodes.OpType = opcodes.IS_UNUSED

	// Compile key if present
	if expr.Key != nil {
		if err := c.compileNode(expr.Key); err != nil {
			return err
		}
		keyOperand = c.nextTemp - 1
		keyType = opcodes.IS_TMP_VAR
	}

	// Compile value if present
	if expr.Value != nil {
		if err := c.compileNode(expr.Value); err != nil {
			return err
		}
		valueOperand = c.nextTemp - 1
		valueType = opcodes.IS_TMP_VAR
	}

	result := c.allocateTemp()

	// Emit YIELD instruction
	c.emit(opcodes.OP_YIELD,
		keyType, keyOperand,
		valueType, valueOperand,
		opcodes.IS_TMP_VAR, result)

	return nil
}

// YieldFromExpression compilation
func (c *Compiler) compileYieldFromExpression(expr *ast.YieldFromExpression) error {
	// Compile the expression to yield from
	if err := c.compileNode(expr.Expression); err != nil {
		return err
	}

	arg := c.nextTemp - 1
	result := c.allocateTemp()

	// Emit YIELD_FROM instruction
	c.emit(opcodes.OP_YIELD_FROM,
		opcodes.IS_TMP_VAR, arg,
		opcodes.IS_UNUSED, 0,
		opcodes.IS_TMP_VAR, result)

	return nil
}

// ThrowExpression compilation (throw as expression)
func (c *Compiler) compileThrowExpression(expr *ast.ThrowExpression) error {
	// Compile the exception to throw
	if err := c.compileNode(expr.Argument); err != nil {
		return err
	}

	arg := c.nextTemp - 1

	// Emit THROW instruction
	c.emit(opcodes.OP_THROW,
		opcodes.IS_TMP_VAR, arg,
		opcodes.IS_UNUSED, 0,
		opcodes.IS_UNUSED, 0)

	return nil
}

// MagicConstantExpression compilation (__FILE__, __LINE__, etc.)
func (c *Compiler) compileMagicConstantExpression(expr *ast.MagicConstantExpression) error {
	result := c.allocateTemp()

	// Get magic constant value based on type
	var constValue *values.Value
	switch expr.TokenType {
	case lexer.T_FILE:
		constValue = values.NewString("<unknown>") // Would need file context
	case lexer.T_LINE:
		constValue = values.NewInt(int64(expr.GetLineNo()))
	case lexer.T_DIR:
		constValue = values.NewString(".")
	case lexer.T_FUNCTION:
		constValue = values.NewString("")
	case lexer.T_CLASS:
		constValue = values.NewString("")
	case lexer.T_NAMESPACE:
		constValue = values.NewString("")
	case lexer.T_TRAIT:
		constValue = values.NewString("")
	default:
		constValue = values.NewString("")
	}

	// Emit assignment of magic constant
	c.emit(opcodes.OP_QM_ASSIGN,
		opcodes.IS_CONST, c.addConstant(constValue),
		opcodes.IS_UNUSED, 0,
		opcodes.IS_TMP_VAR, result)

	return nil
}

// NamespaceNameExpression compilation
func (c *Compiler) compileNamespaceNameExpression(expr *ast.NamespaceNameExpression) error {
	result := c.allocateTemp()

	// Build full namespace name
	namespaceName := strings.Join(expr.Parts, "\\")

	// Emit assignment of namespace name
	c.emit(opcodes.OP_QM_ASSIGN,
		opcodes.IS_CONST, c.addConstant(values.NewString(namespaceName)),
		opcodes.IS_UNUSED, 0,
		opcodes.IS_TMP_VAR, result)

	return nil
}

// NullsafePropertyAccessExpression compilation (object?->property)
func (c *Compiler) compileNullsafePropertyAccessExpression(expr *ast.NullsafePropertyAccessExpression) error {
	// Compile the object expression
	if err := c.compileNode(expr.Object); err != nil {
		return err
	}

	objectOperand := c.nextTemp - 1
	result := c.allocateTemp()

	// Compile property name
	var propOperand uint32
	var propType opcodes.OpType

	switch prop := expr.Property.(type) {
	case *ast.IdentifierNode:
		propOperand = c.addConstant(values.NewString(prop.Name))
		propType = opcodes.IS_CONST
	default:
		if err := c.compileNode(prop); err != nil {
			return err
		}
		propOperand = c.nextTemp - 1
		propType = opcodes.IS_TMP_VAR
	}

	// Emit nullsafe property fetch (for now, same as regular)
	c.emit(opcodes.OP_FETCH_OBJ_R,
		opcodes.IS_TMP_VAR, objectOperand,
		propType, propOperand,
		opcodes.IS_TMP_VAR, result)

	return nil
}

// NullsafeMethodCallExpression compilation (object?->method())
func (c *Compiler) compileNullsafeMethodCallExpression(expr *ast.NullsafeMethodCallExpression) error {
	// Similar to regular method call but with nullsafe semantics
	// For now, compile as regular method call

	// Compile object
	if err := c.compileNode(expr.Object); err != nil {
		return err
	}
	objectOperand := c.nextTemp - 1

	// Compile method name
	var methodOperand uint32
	var methodType opcodes.OpType
	switch m := expr.Method.(type) {
	case *ast.IdentifierNode:
		methodOperand = c.addConstant(values.NewString(m.Name))
		methodType = opcodes.IS_CONST
	default:
		if err := c.compileNode(m); err != nil {
			return err
		}
		methodOperand = c.nextTemp - 1
		methodType = opcodes.IS_TMP_VAR
	}

	result := c.allocateTemp()

	// For now, treat nullsafe same as regular method call
	c.emit(opcodes.OP_DO_FCALL,
		opcodes.IS_TMP_VAR, objectOperand,
		methodType, methodOperand,
		opcodes.IS_TMP_VAR, result)

	return nil
}

// ShellExecExpression compilation (`command`)
func (c *Compiler) compileShellExecExpression(expr *ast.ShellExecExpression) error {
	// Simplified shell execution - not fully implemented
	return fmt.Errorf("shell execution expressions not yet implemented")
}

// CommaExpression compilation (expr1, expr2, expr3)
func (c *Compiler) compileCommaExpression(expr *ast.CommaExpression) error {
	// Compile all expressions, return value of last one
	for _, e := range expr.Expressions {
		if err := c.compileNode(e); err != nil {
			return err
		}
		// Last expression's result remains in the topmost temp
	}

	return nil
}

// SpreadExpression compilation (...$array)
func (c *Compiler) compileSpreadExpression(expr *ast.SpreadExpression) error {
	// Compile the argument to spread
	if err := c.compileNode(expr.Argument); err != nil {
		return err
	}

	// The argument should now be in the last temporary
	argTemp := c.nextTemp - 1
	result := c.allocateTemp()

	// SpreadExpression is context-dependent:
	// - In function calls, it becomes OP_SEND_UNPACK
	// - In array literals, it becomes OP_ADD_ARRAY_UNPACK
	// For now, we'll mark this as a spread and let the parent context handle it

	// Create a temporary marker for spread - the parent node will handle the actual unpacking
	c.emit(opcodes.OP_QM_ASSIGN,
		opcodes.IS_TMP_VAR, argTemp,
		opcodes.IS_UNUSED, 0,
		opcodes.IS_TMP_VAR, result)

	return nil
}

// ArrowFunctionExpression compilation (fn() => expr)
func (c *Compiler) compileArrowFunctionExpression(expr *ast.ArrowFunctionExpression) error {
	// Generate a unique name for the arrow function
	arrowName := fmt.Sprintf("__arrow_%d", len(c.functions))

	// Create new function
	function := &vm.Function{
		Name:         arrowName,
		Instructions: make([]opcodes.Instruction, 0),
		Constants:    make([]*values.Value, 0),
		Parameters:   make([]vm.Parameter, 0),
		IsVariadic:   false,
		IsGenerator:  false,
	}

	// Compile parameters
	if expr.Parameters != nil {
		for _, param := range expr.Parameters.Parameters {
			paramName := ""
			if nameNode, ok := param.Name.(*ast.IdentifierNode); ok {
				paramName = nameNode.Name
			} else {
				return fmt.Errorf("invalid parameter name type")
			}

			vmParam := vm.Parameter{
				Name:        paramName,
				IsReference: param.ByReference,
				HasDefault:  param.DefaultValue != nil,
			}

			// Handle parameter type
			if param.Type != nil {
				vmParam.Type = param.Type.String()
			}

			// Handle default value
			if param.DefaultValue != nil {
				vmParam.HasDefault = true
			}

			// Check for variadic
			if param.Variadic {
				function.IsVariadic = true
			}

			function.Parameters = append(function.Parameters, vmParam)
		}
	}

	// Store current compiler state
	savedInstructions := c.instructions
	savedConstants := c.constants
	savedNextTemp := c.nextTemp
	savedLabels := c.labels

	// Reset compiler state for function compilation
	c.instructions = make([]opcodes.Instruction, 0)
	c.constants = make([]*values.Value, 0)
	c.nextTemp = 100 // Start temporaries at 100 for function scope
	c.labels = make(map[string]int)

	// Compile the arrow function body (expression)
	if err := c.compileNode(expr.Body); err != nil {
		// Restore compiler state on error
		c.instructions = savedInstructions
		c.constants = savedConstants
		c.nextTemp = savedNextTemp
		c.labels = savedLabels
		return err
	}

	// Arrow functions automatically return their expression result
	if len(c.instructions) > 0 {
		// The result should be in the last temporary
		returnTemp := c.nextTemp - 1
		c.emit(opcodes.OP_RETURN,
			opcodes.IS_TMP_VAR, returnTemp,
			opcodes.IS_UNUSED, 0,
			opcodes.IS_UNUSED, 0)
	} else {
		// If no instructions, return null
		nullConst := c.addConstant(values.NewNull())
		c.emit(opcodes.OP_RETURN,
			opcodes.IS_CONST, nullConst,
			opcodes.IS_UNUSED, 0,
			opcodes.IS_UNUSED, 0)
	}

	// Store compiled function
	function.Instructions = c.instructions
	function.Constants = c.constants

	// Restore compiler state
	c.instructions = savedInstructions
	c.constants = savedConstants
	c.nextTemp = savedNextTemp
	c.labels = savedLabels

	// Add function to functions map
	if c.functions == nil {
		c.functions = make(map[string]*vm.Function)
	}
	c.functions[arrowName] = function

	// Create a closure for the arrow function
	result := c.allocateTemp()
	funcConst := c.addConstant(values.NewString(arrowName))

	c.emit(opcodes.OP_CREATE_CLOSURE,
		opcodes.IS_CONST, funcConst,
		opcodes.IS_UNUSED, 0,
		opcodes.IS_TMP_VAR, result)

	return nil
}

// FirstClassCallable compilation (strlen(...))
func (c *Compiler) compileFirstClassCallable(expr *ast.FirstClassCallable) error {
	// First-class callable creates a Closure object
	return fmt.Errorf("first-class callables not yet implemented")
}

// Statement implementations

// GlobalStatement compilation
func (c *Compiler) compileGlobalStatement(stmt *ast.GlobalStatement) error {
	// Make variables global - simplified for now
	for _, variable := range stmt.Variables {
		if v, ok := variable.(*ast.Variable); ok {
			varID := c.addConstant(values.NewString(v.Name))
			varIndex := c.getVariableSlot(v.Name)
			c.emit(opcodes.OP_BIND_GLOBAL,
				opcodes.IS_CONST, varID,
				opcodes.IS_UNUSED, 0,
				opcodes.IS_CV, varIndex)
		}
	}
	return nil
}

// StaticStatement compilation
func (c *Compiler) compileStaticStatement(stmt *ast.StaticStatement) error {
	// Static variables not yet fully implemented
	return fmt.Errorf("static variable declarations not yet implemented")
}

// UnsetStatement compilation
func (c *Compiler) compileUnsetStatement(stmt *ast.UnsetStatement) error {
	// Unset statement can have multiple variables: unset($a, $b, $c)
	for _, variable := range stmt.Variables {
		if err := c.compileUnsetVariable(variable); err != nil {
			return err
		}
	}
	return nil
}

// compileUnsetVariable handles unsetting different types of variables
func (c *Compiler) compileUnsetVariable(expr ast.Expression) error {
	switch v := expr.(type) {
	case *ast.Variable:
		// Simple variable: unset($var)
		if v.Name == "$this" {
			return fmt.Errorf("cannot unset $this")
		}

		varSlot := c.getVariableSlot(v.Name)
		c.emit(opcodes.OP_UNSET_VAR,
			opcodes.IS_CV, varSlot,
			opcodes.IS_UNUSED, 0,
			opcodes.IS_UNUSED, 0)
		return nil

	case *ast.ArrayAccessExpression:
		// Array dimension: unset($arr[key])
		return c.compileUnsetArrayDim(v)

	case *ast.PropertyAccessExpression:
		// Object property: unset($obj->prop)
		return c.compileUnsetObjectProp(v)

	case *ast.StaticPropertyAccessExpression:
		// Static property: unset(Class::$prop)
		return c.compileUnsetStaticProp(v)

	default:
		return fmt.Errorf("cannot unset expression of type %T", expr)
	}
}

// compileUnsetArrayDim handles array dimension unset: unset($arr[key])
func (c *Compiler) compileUnsetArrayDim(expr *ast.ArrayAccessExpression) error {
	// Handle the index
	var indexOperand uint32
	var indexType opcodes.OpType

	if expr.Index == nil {
		// Cannot use [] for unsetting
		return fmt.Errorf("cannot use [] for unsetting")
	}

	// Compile the index expression
	if err := c.compileNode(*expr.Index); err != nil {
		return err
	}
	indexOperand = c.nextTemp - 1
	indexType = opcodes.IS_TMP_VAR

	// Handle the array - for unset, we need to work directly with the variable, not a copy
	var arrayOperand uint32
	var arrayType opcodes.OpType

	if variable, ok := expr.Array.(*ast.Variable); ok {
		// Direct variable access: unset($arr[key])
		arrayOperand = c.getVariableSlot(variable.Name)
		arrayType = opcodes.IS_CV
	} else {
		// Complex expression: unset($obj->arr[key]) - compile and use temporary
		if err := c.compileNode(expr.Array); err != nil {
			return err
		}
		arrayOperand = c.nextTemp - 1
		arrayType = opcodes.IS_TMP_VAR
	}

	// Emit UNSET_DIM instruction
	c.emit(opcodes.OP_FETCH_DIM_UNSET,
		arrayType, arrayOperand,
		indexType, indexOperand,
		opcodes.IS_UNUSED, 0)

	return nil
}

// compileUnsetObjectProp handles object property unset: unset($obj->prop)
func (c *Compiler) compileUnsetObjectProp(expr *ast.PropertyAccessExpression) error {
	// Compile the object expression
	if err := c.compileNode(expr.Object); err != nil {
		return err
	}
	objectOperand := c.nextTemp - 1

	// Handle the property
	var propOperand uint32
	var propType opcodes.OpType

	switch prop := expr.Property.(type) {
	case *ast.IdentifierNode:
		// Static property name
		propOperand = c.addConstant(values.NewString(prop.Name))
		propType = opcodes.IS_CONST
	default:
		// Dynamic property name
		if err := c.compileNode(prop); err != nil {
			return err
		}
		propOperand = c.nextTemp - 1
		propType = opcodes.IS_TMP_VAR
	}

	// Emit UNSET_OBJ instruction
	c.emit(opcodes.OP_FETCH_OBJ_UNSET,
		opcodes.IS_TMP_VAR, objectOperand,
		propType, propOperand,
		opcodes.IS_UNUSED, 0)

	return nil
}

// compileUnsetStaticProp handles static property unset: unset(Class::$prop)
func (c *Compiler) compileUnsetStaticProp(expr *ast.StaticPropertyAccessExpression) error {
	// Compile class name
	var classOperand uint32
	var classType opcodes.OpType

	switch class := expr.Class.(type) {
	case *ast.IdentifierNode:
		// Static class name
		classOperand = c.addConstant(values.NewString(class.Name))
		classType = opcodes.IS_CONST
	default:
		// Dynamic class name
		if err := c.compileNode(class); err != nil {
			return err
		}
		classOperand = c.nextTemp - 1
		classType = opcodes.IS_TMP_VAR
	}

	// Compile property name
	var propOperand uint32
	var propType opcodes.OpType

	switch prop := expr.Property.(type) {
	case *ast.Variable:
		// Remove the $ prefix if present
		propName := prop.Name
		if strings.HasPrefix(propName, "$") {
			propName = propName[1:]
		}
		propOperand = c.addConstant(values.NewString(propName))
		propType = opcodes.IS_CONST
	default:
		if err := c.compileNode(prop); err != nil {
			return err
		}
		propOperand = c.nextTemp - 1
		propType = opcodes.IS_TMP_VAR
	}

	// Emit UNSET_STATIC_PROP instruction
	c.emit(opcodes.OP_FETCH_STATIC_PROP_UNSET,
		classType, classOperand,
		propType, propOperand,
		opcodes.IS_UNUSED, 0)

	return nil
}

// DoWhileStatement compilation
func (c *Compiler) compileDoWhileStatement(stmt *ast.DoWhileStatement) error {
	// Create labels
	startLabel := c.generateLabel()
	endLabel := c.generateLabel()

	// Set break/continue labels for this scope
	oldBreak := c.currentScope().breakLabel
	oldContinue := c.currentScope().continueLabel
	c.currentScope().breakLabel = endLabel
	c.currentScope().continueLabel = startLabel

	// Start of loop body
	c.placeLabel(startLabel)

	// Compile body
	if err := c.compileNode(stmt.Body); err != nil {
		return err
	}

	// Compile condition
	if err := c.compileNode(stmt.Condition); err != nil {
		return err
	}

	// Get condition result
	condResult := c.allocateTemp()
	c.emitMove(condResult)

	// Jump back to start if condition is true
	c.emitJumpNZ(opcodes.IS_TMP_VAR, condResult, startLabel)

	// End label
	c.placeLabel(endLabel)

	// Restore labels
	c.currentScope().breakLabel = oldBreak
	c.currentScope().continueLabel = oldContinue

	return nil
}

// Placeholder implementations for complex features
func (c *Compiler) compileGotoStatement(stmt *ast.GotoStatement) error {
	// Get the target label name
	var labelName string
	if ident, ok := stmt.Label.(*ast.IdentifierNode); ok {
		labelName = ident.Name
	} else {
		return fmt.Errorf("goto statement requires identifier label")
	}

	// Emit unconditional jump to the label
	// The label resolution will be handled by the existing label system
	c.emitJump(opcodes.OP_JMP, opcodes.IS_CONST, 0, labelName)

	return nil
}

func (c *Compiler) compileLabelStatement(stmt *ast.LabelStatement) error {
	// Get the label name
	var labelName string
	if ident, ok := stmt.Name.(*ast.IdentifierNode); ok {
		labelName = ident.Name
	} else {
		return fmt.Errorf("label statement requires identifier name")
	}

	// Place the label at the current position
	// This will resolve any forward jumps to this label
	c.placeLabel(labelName)

	return nil
}

func (c *Compiler) compileHaltCompilerStatement(stmt *ast.HaltCompilerStatement) error {
	// __halt_compiler() stops execution
	c.emit(opcodes.OP_EXIT,
		opcodes.IS_CONST, c.addConstant(values.NewInt(0)),
		opcodes.IS_UNUSED, 0,
		opcodes.IS_UNUSED, 0)
	return nil
}

func (c *Compiler) compileDeclareStatement(stmt *ast.DeclareStatement) error {
	// Compile declare statement: declare(directive=value);
	// Common directives: strict_types, ticks, encoding

	// Process each declaration
	for _, decl := range stmt.Declarations {
		// For now, handle as assignment expressions
		// In a full implementation, these would set compiler/runtime flags
		if err := c.compileNode(decl); err != nil {
			return err
		}
	}

	// Emit DECLARE opcode with declarations count
	declCount := c.addConstant(values.NewInt(int64(len(stmt.Declarations))))
	c.emit(opcodes.OP_DECLARE,
		opcodes.IS_CONST, declCount,
		opcodes.IS_UNUSED, 0,
		opcodes.IS_UNUSED, 0)

	// Compile body if present (for declare blocks)
	if stmt.Body != nil {
		for _, bodyStmt := range stmt.Body {
			if err := c.compileNode(bodyStmt); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Compiler) compileNamespaceStatement(stmt *ast.NamespaceStatement) error {
	// Compile namespace body if present
	if stmt.Body != nil {
		for _, node := range stmt.Body {
			if err := c.compileNode(node); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Compiler) compileUseStatement(stmt *ast.UseStatement) error {
	// Use statements are handled at compile time
	// They affect name resolution but don't generate runtime code
	return nil
}

func (c *Compiler) compileAlternativeIfStatement(stmt *ast.AlternativeIfStatement) error {
	// Alternative if syntax: if (condition): ... endif;
	// This is functionally identical to regular if statements

	// Compile condition
	if err := c.compileNode(stmt.Condition); err != nil {
		return err
	}

	// Generate labels
	elseLabel := c.generateLabel()
	endLabel := c.generateLabel()

	// Get condition result and jump if false
	condResult := c.allocateTemp()
	c.emitMove(condResult)
	c.emitJumpZ(opcodes.IS_TMP_VAR, condResult, elseLabel)

	// Compile then block
	for _, thenStmt := range stmt.Then {
		if err := c.compileNode(thenStmt); err != nil {
			return err
		}
	}

	// Jump to end after then block
	c.emitJump(opcodes.OP_JMP, opcodes.IS_CONST, 0, endLabel)

	// Handle elseif clauses
	for _, elseif := range stmt.ElseIfs {
		c.placeLabel(elseLabel)
		elseLabel = c.generateLabel() // New label for next elseif/else

		// Compile elseif condition
		if err := c.compileNode(elseif.Condition); err != nil {
			return err
		}

		condResult := c.allocateTemp()
		c.emitMove(condResult)
		c.emitJumpZ(opcodes.IS_TMP_VAR, condResult, elseLabel)

		// Compile elseif body
		for _, elseifStmt := range elseif.Body {
			if err := c.compileNode(elseifStmt); err != nil {
				return err
			}
		}

		c.emitJump(opcodes.OP_JMP, opcodes.IS_CONST, 0, endLabel)
	}

	// Handle else clause
	c.placeLabel(elseLabel)
	if stmt.Else != nil {
		for _, elseStmt := range stmt.Else {
			if err := c.compileNode(elseStmt); err != nil {
				return err
			}
		}
	}

	// End label
	c.placeLabel(endLabel)

	return nil
}

func (c *Compiler) compileAlternativeWhileStatement(stmt *ast.AlternativeWhileStatement) error {
	// Alternative while syntax: while (condition): ... endwhile;
	// Functionally identical to regular while loops

	// Labels
	startLabel := c.generateLabel()
	endLabel := c.generateLabel()

	// Set break/continue labels for this scope
	oldBreak := c.currentScope().breakLabel
	oldContinue := c.currentScope().continueLabel
	c.currentScope().breakLabel = endLabel
	c.currentScope().continueLabel = startLabel

	// Start of loop
	c.placeLabel(startLabel)

	// Compile condition
	err := c.compileNode(stmt.Condition)
	if err != nil {
		return err
	}
	condResult := c.allocateTemp()
	c.emitMove(condResult)

	// Jump to end if condition is false
	c.emitJumpZ(opcodes.IS_TMP_VAR, condResult, endLabel)

	// Compile body
	for _, bodyStmt := range stmt.Body {
		err = c.compileNode(bodyStmt)
		if err != nil {
			return err
		}
	}

	// Jump back to start
	c.emitJump(opcodes.OP_JMP, opcodes.IS_CONST, 0, startLabel)

	// End label
	c.placeLabel(endLabel)

	// Restore labels
	c.currentScope().breakLabel = oldBreak
	c.currentScope().continueLabel = oldContinue

	return nil
}

func (c *Compiler) compileAlternativeForStatement(stmt *ast.AlternativeForStatement) error {
	// Alternative for syntax: for (init; condition; update): ... endfor;
	// Functionally identical to regular for loops

	// Labels
	startLabel := c.generateLabel()
	continueLabel := c.generateLabel()
	endLabel := c.generateLabel()

	// Set break/continue labels for this scope
	oldBreak := c.currentScope().breakLabel
	oldContinue := c.currentScope().continueLabel
	c.currentScope().breakLabel = endLabel
	c.currentScope().continueLabel = continueLabel

	// Compile initialization expressions
	for _, init := range stmt.Init {
		if err := c.compileNode(init); err != nil {
			return err
		}
	}

	// Start of loop
	c.placeLabel(startLabel)

	// Compile condition expressions (all must be true)
	if len(stmt.Condition) > 0 {
		for _, cond := range stmt.Condition {
			if err := c.compileNode(cond); err != nil {
				return err
			}
			condResult := c.allocateTemp()
			c.emitMove(condResult)

			// Jump to end if any condition is false
			c.emitJumpZ(opcodes.IS_TMP_VAR, condResult, endLabel)
		}
	}

	// Compile body
	for _, bodyStmt := range stmt.Body {
		if err := c.compileNode(bodyStmt); err != nil {
			return err
		}
	}

	// Continue label (for continue statements)
	c.placeLabel(continueLabel)

	// Compile update expressions
	for _, update := range stmt.Update {
		if err := c.compileNode(update); err != nil {
			return err
		}
	}

	// Jump back to start
	c.emitJump(opcodes.OP_JMP, opcodes.IS_CONST, 0, startLabel)

	// End label
	c.placeLabel(endLabel)

	// Restore labels
	c.currentScope().breakLabel = oldBreak
	c.currentScope().continueLabel = oldContinue

	return nil
}

func (c *Compiler) compileAlternativeForeachStatement(stmt *ast.AlternativeForeachStatement) error {
	// Alternative foreach syntax: foreach ($array as $value): ... endforeach;
	// This is more complex and requires iterator management like regular foreach

	// For now, provide a basic implementation that compiles but may not execute perfectly
	// A full implementation would need proper iterator support in the VM

	// Labels
	startLabel := c.generateLabel()
	endLabel := c.generateLabel()

	// Set break/continue labels for this scope
	oldBreak := c.currentScope().breakLabel
	oldContinue := c.currentScope().continueLabel
	c.currentScope().breakLabel = endLabel
	c.currentScope().continueLabel = startLabel

	// Compile iterable expression
	if err := c.compileNode(stmt.Iterable); err != nil {
		return err
	}

	// For a simplified implementation, we'll just compile the body once
	// In a full implementation, this would set up iteration over the array

	// Allocate iterator (simplified)
	iteratorVar := c.allocateTemp()
	c.emit(opcodes.OP_QM_ASSIGN,
		opcodes.IS_TMP_VAR, c.nextTemp-1, // The iterable result
		opcodes.IS_UNUSED, 0,
		opcodes.IS_TMP_VAR, iteratorVar)

	// Start iteration label
	c.placeLabel(startLabel)

	// For now, just execute the body once (simplified)
	// TODO: Implement proper iteration logic

	// Compile foreach body
	for _, bodyStmt := range stmt.Body {
		if err := c.compileNode(bodyStmt); err != nil {
			return err
		}
	}

	// End iteration (simplified - no actual iteration)
	c.placeLabel(endLabel)

	// Restore labels
	c.currentScope().breakLabel = oldBreak
	c.currentScope().continueLabel = oldContinue

	return nil
}

// Declaration implementations
func (c *Compiler) compileInterfaceDeclaration(decl *ast.InterfaceDeclaration) error {
	if decl.Name == nil {
		return fmt.Errorf("interface declaration missing name")
	}

	interfaceName := decl.Name.Name

	// Check if interface already exists
	if _, exists := c.interfaces[interfaceName]; exists {
		return fmt.Errorf("interface %s already declared", interfaceName)
	}

	// Create new interface
	iface := &vm.Interface{
		Name:    interfaceName,
		Methods: make(map[string]*vm.InterfaceMethod),
		Extends: make([]string, 0),
	}

	// Handle extends
	for _, parent := range decl.Extends {
		iface.Extends = append(iface.Extends, parent.Name)
	}

	// Add interface methods
	for _, method := range decl.Methods {
		if method.Name == nil {
			continue
		}

		methodName := method.Name.Name
		interfaceMethod := &vm.InterfaceMethod{
			Name:       methodName,
			Visibility: method.Visibility,
			Parameters: make([]*vm.Parameter, 0),
		}

		// Add parameters if present
		if method.Parameters != nil {
			for _, param := range method.Parameters.Parameters {
				if param.Name == nil {
					continue
				}

				var paramName string
				if ident, ok := param.Name.(*ast.IdentifierNode); ok {
					paramName = ident.Name
				} else {
					continue
				}

				vmParam := &vm.Parameter{
					Name:         paramName,
					Type:         "", // Type hints not fully implemented yet
					IsReference:  param.ByReference,
					HasDefault:   param.DefaultValue != nil,
					DefaultValue: nil, // TODO: evaluate default value
				}
				interfaceMethod.Parameters = append(interfaceMethod.Parameters, vmParam)
			}
		}

		iface.Methods[methodName] = interfaceMethod
	}

	// Store interface
	c.interfaces[interfaceName] = iface

	// Emit interface declaration opcode
	nameConstant := c.addConstant(values.NewString(interfaceName))
	c.emit(opcodes.OP_DECLARE_INTERFACE,
		opcodes.IS_CONST, nameConstant,
		opcodes.IS_UNUSED, 0,
		opcodes.IS_UNUSED, 0)

	return nil
}

func (c *Compiler) compileTraitDeclaration(decl *ast.TraitDeclaration) error {
	if decl.Name == nil {
		return fmt.Errorf("trait declaration missing name")
	}

	traitName := decl.Name.Name

	// Check if trait already exists
	if _, exists := c.traits[traitName]; exists {
		return fmt.Errorf("trait %s already declared", traitName)
	}

	// Create new trait
	trait := &vm.Trait{
		Name:       traitName,
		Properties: make(map[string]*vm.Property),
		Methods:    make(map[string]*vm.Function),
	}

	// Compile trait properties
	for _, prop := range decl.Properties {
		if err := c.compileTraitProperty(trait, prop); err != nil {
			return err
		}
	}

	// Compile trait methods
	for _, method := range decl.Methods {
		if err := c.compileTraitMethod(trait, method); err != nil {
			return err
		}
	}

	// Store trait
	c.traits[traitName] = trait

	// Emit trait declaration opcode
	nameConstant := c.addConstant(values.NewString(traitName))
	c.emit(opcodes.OP_DECLARE_TRAIT,
		opcodes.IS_CONST, nameConstant,
		opcodes.IS_UNUSED, 0,
		opcodes.IS_UNUSED, 0)

	return nil
}

func (c *Compiler) compileEnumDeclaration(decl *ast.EnumDeclaration) error {
	if decl.Name == nil {
		return fmt.Errorf("enum declaration missing name")
	}

	enumName := decl.Name.Name

	// Check if enum already exists (enums are stored as classes in the VM)
	if _, exists := c.classes[enumName]; exists {
		return fmt.Errorf("enum %s already declared", enumName)
	}

	// Create enum as a special class
	enumClass := &vm.Class{
		Name:        enumName,
		ParentClass: "",
		Properties:  make(map[string]*vm.Property),
		Methods:     make(map[string]*vm.Function),
		Constants:   make(map[string]*values.Value),
		IsAbstract:  false,
		IsFinal:     true, // Enums are final by default
	}

	// Add enum cases as constants
	for _, enumCase := range decl.Cases {
		if enumCase.Name == nil {
			continue
		}

		caseName := enumCase.Name.Name
		var caseValue *values.Value

		if enumCase.Value != nil {
			// Backed enum - evaluate the backing value
			// For now, simplified - assume it's a literal
			caseValue = values.NewString(caseName) // Simplified
		} else {
			// Pure enum - use the case name
			caseValue = values.NewString(caseName)
		}

		enumClass.Constants[caseName] = caseValue
	}

	// Add enum methods if any
	for _, method := range decl.Methods {
		if err := c.compileEnumMethod(enumClass, method); err != nil {
			return err
		}
	}

	// Store enum as class
	c.classes[enumName] = enumClass

	// Emit class declaration opcode (reuse existing class opcodes)
	nameConstant := c.addConstant(values.NewString(enumName))
	c.emit(opcodes.OP_DECLARE_CLASS,
		opcodes.IS_CONST, nameConstant,
		opcodes.IS_UNUSED, 0,
		opcodes.IS_UNUSED, 0)

	return nil
}

func (c *Compiler) compileUseTraitStatement(stmt *ast.UseTraitStatement) error {
	// Use trait statements are handled within class context
	if c.currentClass == nil {
		return fmt.Errorf("use trait statement outside of class context")
	}

	// For each trait being used
	for _, traitName := range stmt.Traits {
		if traitName == nil {
			continue
		}

		// Emit USE_TRAIT opcode
		traitConstant := c.addConstant(values.NewString(traitName.Name))
		c.emit(opcodes.OP_USE_TRAIT,
			opcodes.IS_CONST, traitConstant,
			opcodes.IS_UNUSED, 0,
			opcodes.IS_UNUSED, 0)
	}

	// TODO: Handle trait adaptations (precedence and alias rules)
	// This would require more complex logic to resolve method conflicts

	return nil
}

func (c *Compiler) compileHookedPropertyDeclaration(decl *ast.HookedPropertyDeclaration) error {
	return fmt.Errorf("hooked property declarations not yet implemented")
}

// Helper methods for trait compilation
func (c *Compiler) compileTraitProperty(trait *vm.Trait, prop *ast.PropertyDeclaration) error {
	// PropertyDeclaration contains a single property
	property := &vm.Property{
		Name:       prop.Name,
		Type:       "", // Type hints not fully implemented yet
		Visibility: prop.Visibility,
		IsStatic:   prop.Static,
	}

	trait.Properties[prop.Name] = property
	return nil
}

func (c *Compiler) compileTraitMethod(trait *vm.Trait, method *ast.FunctionDeclaration) error {
	// For now, simplified method compilation - just store the method info
	// Full method compilation would require compiling the method body
	if method.Name == nil {
		return nil
	}

	var methodName string
	if ident, ok := method.Name.(*ast.IdentifierNode); ok {
		methodName = ident.Name
	} else {
		return nil
	}

	// Create a simplified function for the trait
	// In a full implementation, we'd compile the method body here
	function := &vm.Function{
		Name:         methodName,
		Parameters:   make([]vm.Parameter, 0),
		Instructions: make([]opcodes.Instruction, 0), // Empty for now
		Constants:    make([]*values.Value, 0),
	}

	trait.Methods[methodName] = function
	return nil
}

// Helper method for enum compilation
func (c *Compiler) compileEnumMethod(enumClass *vm.Class, method *ast.FunctionDeclaration) error {
	// Similar to trait method compilation but store in enum class
	if method.Name == nil {
		return nil
	}

	var methodName string
	if ident, ok := method.Name.(*ast.IdentifierNode); ok {
		methodName = ident.Name
	} else {
		return nil
	}

	// Create a simplified function for the enum
	function := &vm.Function{
		Name:         methodName,
		Parameters:   make([]vm.Parameter, 0),
		Instructions: make([]opcodes.Instruction, 0),
		Constants:    make([]*values.Value, 0),
	}

	enumClass.Methods[methodName] = function
	return nil
}
