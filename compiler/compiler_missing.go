package compiler

import (
	"fmt"
	"strconv"
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
	// Allocate temporary variable to store the previous error reporting level
	previousErrorLevel := c.allocateTemp()

	// Emit BEGIN_SILENCE instruction - saves current error reporting level to temp var
	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_UNUSED, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	beginInst := opcodes.Instruction{
		Opcode:  opcodes.OP_BEGIN_SILENCE,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0,
		Op2:     0,
		Result:  previousErrorLevel,
	}
	c.instructions = append(c.instructions, beginInst)

	// Compile the inner expression (errors will be suppressed during execution)
	if err := c.compileNode(expr.Expression); err != nil {
		return err
	}

	// Emit END_SILENCE instruction - restores previous error reporting level
	op1Type, op2Type = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_UNUSED)
	endInst := opcodes.Instruction{
		Opcode:  opcodes.OP_END_SILENCE,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     previousErrorLevel, // Use the saved error level from BEGIN_SILENCE
		Op2:     0,
		Result:  0,
	}
	c.instructions = append(c.instructions, endInst)

	return nil
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
	// list() expressions are used for array destructuring
	// This function is called when list() appears on the left side of an assignment
	// The actual list assignment logic is handled in assignment compilation

	// For a standalone list expression (which is unusual), we create an empty array
	// Most commonly, this would be called from assignment compilation
	result := c.allocateTemp()

	// Create an empty array to represent the list structure
	c.emit(opcodes.OP_INIT_ARRAY,
		opcodes.IS_UNUSED, 0,
		opcodes.IS_UNUSED, 0,
		opcodes.IS_TMP_VAR, result)

	// Add each element of the list as array elements (for structural representation)
	for i, element := range expr.Elements {
		if element != nil {
			// Compile the element (usually a variable)
			if err := c.compileNode(element); err != nil {
				return err
			}

			elementTemp := c.nextTemp - 1
			indexConst := c.addConstant(values.NewInt(int64(i)))

			// Add to array
			c.emit(opcodes.OP_ADD_ARRAY_ELEMENT,
				opcodes.IS_TMP_VAR, elementTemp,
				opcodes.IS_CONST, indexConst,
				opcodes.IS_TMP_VAR, result)
		} else {
			// Handle empty elements in list (like list($a, , $c))
			indexConst := c.addConstant(values.NewInt(int64(i)))
			nullConst := c.addConstant(values.NewNull())

			c.emit(opcodes.OP_ADD_ARRAY_ELEMENT,
				opcodes.IS_CONST, nullConst,
				opcodes.IS_CONST, indexConst,
				opcodes.IS_TMP_VAR, result)
		}
	}

	return nil
}

// compileListAssignmentFromValue handles list assignment from a temporary value: list($a, $b, $c) = $array
func (c *Compiler) compileListAssignmentFromValue(listExpr *ast.ListExpression, sourceTemp uint32) error {
	// For each element in the list, fetch from the source array and assign
	for i, element := range listExpr.Elements {
		if element == nil {
			// Skip empty elements
			continue
		}

		// Create index constant
		indexConst := c.addConstant(values.NewInt(int64(i)))

		// Check if this is a nested list
		if nestedList, ok := element.(*ast.ListExpression); ok {
			// Handle nested list assignment recursively
			// First fetch the sub-array
			fetchTemp := c.allocateTemp()
			c.emit(opcodes.OP_FETCH_LIST_R,
				opcodes.IS_TMP_VAR, sourceTemp,
				opcodes.IS_CONST, indexConst,
				opcodes.IS_TMP_VAR, fetchTemp)

			// Recursively handle nested list
			if err := c.compileListAssignmentFromValue(nestedList, fetchTemp); err != nil {
				return err
			}
		} else {
			// Regular variable assignment
			fetchTemp := c.allocateTemp()

			// Fetch the element from the source array
			c.emit(opcodes.OP_FETCH_LIST_R,
				opcodes.IS_TMP_VAR, sourceTemp,
				opcodes.IS_CONST, indexConst,
				opcodes.IS_TMP_VAR, fetchTemp)

			// Assign to the variable
			if variable, ok := element.(*ast.Variable); ok {
				varSlot := c.getVariableSlot(variable.Name)
				c.emit(opcodes.OP_ASSIGN,
					opcodes.IS_TMP_VAR, fetchTemp,
					opcodes.IS_UNUSED, 0,
					opcodes.IS_CV, varSlot)
			} else {
				// Handle more complex left-hand side expressions
				return fmt.Errorf("complex list assignment targets not fully implemented")
			}
		}
	}

	return nil
}

// compileListAssignment handles list assignment: list($a, $b, $c) = $array
func (c *Compiler) compileListAssignment(listExpr *ast.ListExpression, sourceNode *ast.Variable) error {
	// This is the main logic for list assignment
	// Get the source array
	if err := c.compileNode(sourceNode); err != nil {
		return err
	}

	sourceTemp := c.nextTemp - 1

	// For each element in the list, fetch from the source array and assign
	for i, element := range listExpr.Elements {
		if element == nil {
			// Skip empty elements
			continue
		}

		// Create index constant
		indexConst := c.addConstant(values.NewInt(int64(i)))

		// Check if this is a nested list
		if nestedList, ok := element.(*ast.ListExpression); ok {
			// Handle nested list assignment recursively
			// First fetch the sub-array
			fetchTemp := c.allocateTemp()
			c.emit(opcodes.OP_FETCH_LIST_R,
				opcodes.IS_TMP_VAR, sourceTemp,
				opcodes.IS_CONST, indexConst,
				opcodes.IS_TMP_VAR, fetchTemp)

			// Create a temporary variable node for the fetched value
			tempVarNode := &ast.Variable{
				BaseNode: ast.BaseNode{
					Kind: ast.ASTVar,
				},
				Name: fmt.Sprintf("__list_temp_%d", fetchTemp),
			}

			// Recursively handle nested list
			if err := c.compileListAssignment(nestedList, tempVarNode); err != nil {
				return err
			}
		} else {
			// Regular variable assignment
			fetchTemp := c.allocateTemp()

			// Fetch the element from the source array
			c.emit(opcodes.OP_FETCH_LIST_R,
				opcodes.IS_TMP_VAR, sourceTemp,
				opcodes.IS_CONST, indexConst,
				opcodes.IS_TMP_VAR, fetchTemp)

			// Assign to the variable
			if variable, ok := element.(*ast.Variable); ok {
				varSlot := c.getVariableSlot(variable.Name)
				c.emit(opcodes.OP_ASSIGN,
					opcodes.IS_TMP_VAR, fetchTemp,
					opcodes.IS_UNUSED, 0,
					opcodes.IS_CV, varSlot)
			} else {
				// Handle more complex left-hand side expressions
				return fmt.Errorf("complex list assignment targets not fully implemented")
			}
		}
	}

	return nil
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
	// Compile declare statement: declare(directive=value) or declare(directive=value) { ... }
	// Common directives: strict_types, ticks, encoding

	// Process each declaration and handle specific directives
	for _, decl := range stmt.Declarations {
		if assignment, ok := decl.(*ast.AssignmentExpression); ok {
			// Extract directive name and value
			var directiveName string

			// Get directive name (left side)
			if variable, ok := assignment.Left.(*ast.Variable); ok {
				directiveName = strings.TrimPrefix(variable.Name, "$")
			} else if identifier, ok := assignment.Left.(*ast.IdentifierNode); ok {
				directiveName = identifier.Name
			} else {
				return fmt.Errorf("invalid declare directive format")
			}

			// Compile directive value (right side)
			if err := c.compileNode(assignment.Right); err != nil {
				return err
			}
			valueTemp := c.nextTemp - 1

			// Handle specific directives
			switch directiveName {
			case "strict_types":
				// strict_types=1 enables strict type checking for the current file
				// This affects how type declarations are enforced
				c.emitDeclareDirective("strict_types", valueTemp)

			case "ticks":
				// ticks=N causes tick events to be emitted every N statements
				// Used for debugging and profiling
				c.emitDeclareDirective("ticks", valueTemp)

			case "encoding":
				// encoding="UTF-8" declares the encoding of the source file
				// Affects string handling in some cases
				c.emitDeclareDirective("encoding", valueTemp)

			default:
				// Unknown directive - emit generic declare
				directiveNameConst := c.addConstant(values.NewString(directiveName))
				c.emit(opcodes.OP_DECLARE,
					opcodes.IS_CONST, directiveNameConst,
					opcodes.IS_TMP_VAR, valueTemp,
					opcodes.IS_UNUSED, 0)
			}
		} else {
			// Handle other declaration types
			if err := c.compileNode(decl); err != nil {
				return err
			}
		}
	}

	// Compile body if present (for declare blocks: declare() { ... })
	if stmt.Body != nil {
		// Create new scope for declare block
		c.pushScope(false) // false = not a function scope
		defer c.popScope()

		for _, bodyStmt := range stmt.Body {
			if err := c.compileNode(bodyStmt); err != nil {
				return err
			}
		}
	}

	return nil
}

// emitDeclareDirective emits a declare directive with proper opcode
func (c *Compiler) emitDeclareDirective(directive string, valueTemp uint32) {
	directiveNameConst := c.addConstant(values.NewString(directive))

	switch directive {
	case "ticks":
		// For ticks, we need special handling to set up tick callbacks
		c.emit(opcodes.OP_TICKS,
			opcodes.IS_TMP_VAR, valueTemp,
			opcodes.IS_UNUSED, 0,
			opcodes.IS_UNUSED, 0)

	default:
		// Generic declare opcode
		c.emit(opcodes.OP_DECLARE,
			opcodes.IS_CONST, directiveNameConst,
			opcodes.IS_TMP_VAR, valueTemp,
			opcodes.IS_UNUSED, 0)
	}
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
	// This should work exactly like regular foreach, just with different body structure

	// Create labels for the foreach loop
	startLabel := c.generateLabel()
	endLabel := c.generateLabel()
	continueLabel := c.generateLabel()

	// Set break/continue labels for this scope
	oldBreak := c.currentScope().breakLabel
	oldContinue := c.currentScope().continueLabel
	c.currentScope().breakLabel = endLabel
	c.currentScope().continueLabel = continueLabel

	// Compile the iterable expression
	err := c.compileNode(stmt.Iterable)
	if err != nil {
		return err
	}
	iterableTemp := c.nextTemp - 1

	// Initialize foreach iterator
	iteratorTemp := c.allocateTemp()
	c.emit(opcodes.OP_FE_RESET, opcodes.IS_TMP_VAR, iterableTemp, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, iteratorTemp)

	// Start of loop
	c.placeLabel(startLabel)

	// Fetch next element - this will set values or null if no more elements
	valueTemp := c.allocateTemp()
	var keyTemp uint32 = 0
	if stmt.Key != nil {
		keyTemp = c.allocateTemp()
		c.emit(opcodes.OP_FE_FETCH, opcodes.IS_TMP_VAR, iteratorTemp, opcodes.IS_TMP_VAR, keyTemp, opcodes.IS_TMP_VAR, valueTemp)
	} else {
		c.emit(opcodes.OP_FE_FETCH, opcodes.IS_TMP_VAR, iteratorTemp, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, valueTemp)
	}

	// Check if value is null (end of iteration)
	nullCheckTemp := c.allocateTemp()
	nullConstant := c.addConstant(values.NewNull())
	c.emit(opcodes.OP_IS_IDENTICAL, opcodes.IS_TMP_VAR, valueTemp, opcodes.IS_CONST, nullConstant, opcodes.IS_TMP_VAR, nullCheckTemp)
	c.emitJumpNZ(opcodes.IS_TMP_VAR, nullCheckTemp, endLabel)

	// Assign key to key variable if present
	if stmt.Key != nil {
		if keyVar, ok := stmt.Key.(*ast.Variable); ok {
			keySlot := c.getOrCreateVariable(keyVar.Name)
			c.emit(opcodes.OP_ASSIGN, opcodes.IS_TMP_VAR, keyTemp, opcodes.IS_UNUSED, 0, opcodes.IS_VAR, keySlot)
		}
	}

	// Assign value to value variable
	if valueVar, ok := stmt.Value.(*ast.Variable); ok {
		valueSlot := c.getOrCreateVariable(valueVar.Name)
		c.emit(opcodes.OP_ASSIGN, opcodes.IS_TMP_VAR, valueTemp, opcodes.IS_UNUSED, 0, opcodes.IS_VAR, valueSlot)
	}

	// Compile body statements (this is the main difference from regular foreach)
	for _, bodyStmt := range stmt.Body {
		err = c.compileNode(bodyStmt)
		if err != nil {
			return err
		}
	}

	// Continue label (where continue jumps to)
	c.placeLabel(continueLabel)

	// Jump back to start for next iteration
	c.emitJump(opcodes.OP_JMP, opcodes.IS_CONST, 0, startLabel)

	// End label
	c.placeLabel(endLabel)

	// Restore old break/continue labels
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

		// Find the trait
		traitNameStr := traitName.Name
		trait, exists := c.traits[traitNameStr]
		if !exists {
			return fmt.Errorf("trait %s not found", traitNameStr)
		}

		// Copy trait methods into current class
		// This implements the PHP behavior where trait methods become class methods
		for methodName, traitMethod := range trait.Methods {
			// Check for method conflicts
			if _, exists := c.currentClass.Methods[methodName]; exists {
				// In a full implementation, we would handle method precedence rules here
				// For now, we'll allow the trait method to override
			}

			// Create a copy of the trait method for the class
			classMethod := &vm.Function{
				Name:         traitMethod.Name,
				Instructions: make([]opcodes.Instruction, len(traitMethod.Instructions)),
				Constants:    make([]*values.Value, len(traitMethod.Constants)),
				Parameters:   make([]vm.Parameter, len(traitMethod.Parameters)),
				IsVariadic:   traitMethod.IsVariadic,
				IsGenerator:  traitMethod.IsGenerator,
			}

			// Deep copy instructions
			copy(classMethod.Instructions, traitMethod.Instructions)

			// Deep copy constants
			copy(classMethod.Constants, traitMethod.Constants)

			// Deep copy parameters
			copy(classMethod.Parameters, traitMethod.Parameters)

			// Add the method to the current class
			c.currentClass.Methods[methodName] = classMethod
		}

		// Copy trait properties into current class
		for propName, traitProp := range trait.Properties {
			// Check for property conflicts
			if _, exists := c.currentClass.Properties[propName]; exists {
				// In a full implementation, we would handle property conflicts
				// For now, we'll allow the trait property to override
			}

			// Create a copy of the trait property for the class
			classProp := &vm.Property{
				Name:       traitProp.Name,
				Visibility: traitProp.Visibility,
				IsStatic:   traitProp.IsStatic,
				Type:       traitProp.Type,
			}

			// Add the property to the current class
			c.currentClass.Properties[propName] = classProp
		}

		// Emit USE_TRAIT opcode for runtime tracking
		traitConstant := c.addConstant(values.NewString(traitNameStr))
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

// evaluateConstantExpression attempts to evaluate a constant expression at compile time
// This is a simplified implementation that handles basic literals
func (c *Compiler) evaluateConstantExpression(expr ast.Node) *values.Value {
	switch e := expr.(type) {
	case *ast.StringLiteral:
		return values.NewString(e.Value)
	case *ast.NumberLiteral:
		// Try integer first
		if intVal, err := strconv.ParseInt(e.Value, 10, 64); err == nil {
			return values.NewInt(intVal)
		}
		// Fall back to float
		if floatVal, err := strconv.ParseFloat(e.Value, 64); err == nil {
			return values.NewFloat(floatVal)
		}
		// If we can't parse it, use 0
		return values.NewInt(0)
	case *ast.BooleanLiteral:
		return values.NewBool(e.Value)
	case *ast.NullLiteral:
		return values.NewNull()
	default:
		// For complex expressions, we can't evaluate at compile time
		return nil
	}
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
	// Validate method name
	if method.Name == nil {
		return fmt.Errorf("trait method missing name")
	}

	nameNode, ok := method.Name.(*ast.IdentifierNode)
	if !ok {
		return fmt.Errorf("invalid trait method name type")
	}
	methodName := nameNode.Name

	// Check if method already exists in trait
	if _, exists := trait.Methods[methodName]; exists {
		return fmt.Errorf("method %s already declared in trait %s", methodName, trait.Name)
	}

	// Create new function for the trait method
	function := &vm.Function{
		Name:         methodName,
		Instructions: make([]opcodes.Instruction, 0),
		Constants:    make([]*values.Value, 0),
		Parameters:   make([]vm.Parameter, 0),
		IsVariadic:   false,
		IsGenerator:  false,
	}

	// Compile parameters
	if method.Parameters != nil {
		for _, param := range method.Parameters.Parameters {
			paramName := ""
			if nameNode, ok := param.Name.(*ast.IdentifierNode); ok {
				paramName = nameNode.Name
			} else {
				return fmt.Errorf("invalid parameter name type in trait method %s", methodName)
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
				// Compile the default value expression
				// For now, we'll evaluate simple default values at compile time
				defaultValue := c.evaluateConstantExpression(param.DefaultValue)
				if defaultValue != nil {
					vmParam.DefaultValue = defaultValue
				} else {
					// If we can't evaluate it at compile time, use null as fallback
					vmParam.DefaultValue = values.NewNull()
				}
			}

			// Check for variadic
			if param.Variadic {
				function.IsVariadic = true
			}

			function.Parameters = append(function.Parameters, vmParam)
		}
	}

	// Store current compiler state to restore later
	oldInstructions := c.instructions
	oldConstants := c.constants
	oldCurrentClass := c.currentClass

	// Reset for trait method compilation
	c.instructions = make([]opcodes.Instruction, 0)
	c.constants = make([]*values.Value, 0)
	// Note: we don't set c.currentClass for traits as trait methods are different

	// Create method scope
	c.pushScope(true)

	// Add implicit $this variable FIRST (to ensure it gets slot 0)
	thisSlot := c.getOrCreateVariable("this")
	_ = thisSlot // Ensure it's slot 0

	// Set up parameter variables in the method scope (they get slots 1, 2, 3, ...)
	if method.Parameters != nil {
		for _, param := range method.Parameters.Parameters {
			if nameNode, ok := param.Name.(*ast.IdentifierNode); ok {
				// Register parameter name in method scope
				c.getOrCreateVariable(nameNode.Name)
			}
		}
	}

	// Compile method body
	for _, stmt := range method.Body {
		err := c.compileNode(stmt)
		if err != nil {
			// Restore compiler state on error
			c.popScope()
			c.instructions = oldInstructions
			c.constants = oldConstants
			c.currentClass = oldCurrentClass
			return fmt.Errorf("error compiling trait method %s in trait %s: %v", methodName, trait.Name, err)
		}
	}

	// Add implicit return if needed
	if len(c.instructions) == 0 || c.instructions[len(c.instructions)-1].Opcode != opcodes.OP_RETURN {
		c.emit(opcodes.OP_RETURN, opcodes.IS_CONST, c.addConstant(values.NewNull()), 0, 0, 0, 0)
	}

	// Store compiled method
	function.Instructions = c.instructions
	function.Constants = c.constants

	// Store the method in the trait
	trait.Methods[methodName] = function

	// Restore compiler state
	c.popScope()
	c.instructions = oldInstructions
	c.constants = oldConstants
	c.currentClass = oldCurrentClass

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
