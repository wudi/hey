package compiler

import (
	"fmt"
	"strings"

	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/compiler/opcodes"
	"github.com/wudi/php-parser/compiler/values"
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

	// For each argument, check if it's set
	for i, e := range expr.Arguments.Arguments {
		if err := c.compileNode(e); err != nil {
			return err
		}

		arg := c.nextTemp - 1
		tempResult := c.allocateTemp()

		c.emit(opcodes.OP_ISSET_ISEMPTY_VAR,
			opcodes.IS_TMP_VAR, arg,
			opcodes.IS_UNUSED, 0,
			opcodes.IS_TMP_VAR, tempResult)

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

	return nil
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
	// For now, just return an error - spread expressions need context-specific handling
	return fmt.Errorf("spread expressions not yet implemented")
}

// ArrowFunctionExpression compilation (fn() => expr)
func (c *Compiler) compileArrowFunctionExpression(expr *ast.ArrowFunctionExpression) error {
	// Similar to anonymous function but simpler
	return fmt.Errorf("arrow functions not yet implemented")
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
	// Simplified unset implementation for now
	return fmt.Errorf("unset statements not yet implemented")
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
	return fmt.Errorf("goto statements not yet implemented")
}

func (c *Compiler) compileLabelStatement(stmt *ast.LabelStatement) error {
	return fmt.Errorf("label statements not yet implemented")
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
	return fmt.Errorf("declare statements not yet implemented")
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
	// Alternative if syntax not yet implemented
	return fmt.Errorf("alternative if statements not yet implemented")
}

func (c *Compiler) compileAlternativeWhileStatement(stmt *ast.AlternativeWhileStatement) error {
	// Alternative while syntax not yet implemented
	return fmt.Errorf("alternative while statements not yet implemented")
}

func (c *Compiler) compileAlternativeForStatement(stmt *ast.AlternativeForStatement) error {
	return fmt.Errorf("alternative for statements not yet fully implemented")
}

func (c *Compiler) compileAlternativeForeachStatement(stmt *ast.AlternativeForeachStatement) error {
	return fmt.Errorf("alternative foreach statements not yet fully implemented")
}

// Declaration implementations
func (c *Compiler) compileInterfaceDeclaration(decl *ast.InterfaceDeclaration) error {
	return fmt.Errorf("interface declarations not yet implemented")
}

func (c *Compiler) compileTraitDeclaration(decl *ast.TraitDeclaration) error {
	return fmt.Errorf("trait declarations not yet implemented")
}

func (c *Compiler) compileEnumDeclaration(decl *ast.EnumDeclaration) error {
	return fmt.Errorf("enum declarations not yet implemented")
}

func (c *Compiler) compileUseTraitStatement(stmt *ast.UseTraitStatement) error {
	return fmt.Errorf("use trait statements not yet implemented")
}

func (c *Compiler) compileHookedPropertyDeclaration(decl *ast.HookedPropertyDeclaration) error {
	return fmt.Errorf("hooked property declarations not yet implemented")
}
