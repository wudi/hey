package compiler

import (
	"fmt"

	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/compiler/opcodes"
	"github.com/wudi/php-parser/compiler/values"
)

// SimpleCompiler is a simplified version that works with existing AST types
type SimpleCompiler struct {
	instructions []opcodes.Instruction
	constants    []*values.Value
	nextTemp     uint32
}

// NewSimpleCompiler creates a new simple bytecode compiler
func NewSimpleCompiler() *SimpleCompiler {
	return &SimpleCompiler{
		instructions: make([]opcodes.Instruction, 0),
		constants:    make([]*values.Value, 0),
		nextTemp:     1000,
	}
}

// CompileNode compiles a single AST node to bytecode
func (c *SimpleCompiler) CompileNode(node ast.Node) error {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	// Basic expressions that exist in the current AST
	case *ast.BinaryExpression:
		return c.compileBinaryExpression(n)
	case *ast.UnaryExpression:
		return c.compileUnaryExpression(n)
	case *ast.AssignmentExpression:
		return c.compileAssignmentExpression(n)
	case *ast.Variable:
		return c.compileVariable(n)
	case *ast.NumberLiteral:
		return c.compileNumberLiteral(n)
	case *ast.StringLiteral:
		return c.compileStringLiteral(n)
	case *ast.BooleanLiteral:
		return c.compileBooleanLiteral(n)
	case *ast.NullLiteral:
		return c.compileNullLiteral(n)
	case *ast.IdentifierNode:
		return c.compileIdentifier(n)

	// Statements
	case *ast.ExpressionStatement:
		return c.compileExpressionStatement(n)
	case *ast.EchoStatement:
		return c.compileEchoStatement(n)

	// Program
	case *ast.Program:
		return c.compileProgram(n)

	default:
		// For unsupported node types, just ignore for now
		return fmt.Errorf("unsupported AST node type: %T", n)
	}
}

// Compile methods for existing AST types

func (c *SimpleCompiler) compileBinaryExpression(expr *ast.BinaryExpression) error {
	// Compile left operand
	err := c.CompileNode(expr.Left)
	if err != nil {
		return err
	}
	// The result of left operand should be in the last allocated temp
	leftResult := c.nextTemp - 1

	// Compile right operand
	err = c.CompileNode(expr.Right)
	if err != nil {
		return err
	}
	// The result of right operand should be in the last allocated temp
	rightResult := c.nextTemp - 1

	// Generate operation
	result := c.allocateTemp()
	opcode := c.getOpcodeForOperator(expr.Operator)
	// emit(opcode, op1Type, op1, op2Type, op2, resultType, result)
	c.emit(opcode, opcodes.IS_TMP_VAR, leftResult, opcodes.IS_TMP_VAR, rightResult, opcodes.IS_TMP_VAR, result)

	return nil
}

func (c *SimpleCompiler) compileUnaryExpression(expr *ast.UnaryExpression) error {
	err := c.CompileNode(expr.Operand)
	if err != nil {
		return err
	}

	// The result of operand should be in the last allocated temp
	operandResult := c.nextTemp - 1
	result := c.allocateTemp()
	opcode := c.getOpcodeForUnaryOperator(expr.Operator)
	// emit(opcode, op1Type, op1, op2Type, op2, resultType, result)
	c.emit(opcode, opcodes.IS_TMP_VAR, operandResult, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, result)

	return nil
}

func (c *SimpleCompiler) compileAssignmentExpression(expr *ast.AssignmentExpression) error {
	// Compile right-hand side first
	err := c.CompileNode(expr.Right)
	if err != nil {
		return err
	}

	// The result of right-hand side should be in the last allocated temp
	valueResult := c.nextTemp - 1

	// For now, assume left side is always a variable
	if variable, ok := expr.Left.(*ast.Variable); ok {
		varSlot := c.getVariableSlot(variable.Name)
		
		// Get the appropriate assignment opcode based on operator
		opcode := c.getOpcodeForAssignmentOperator(expr.Operator)
		
		if opcode == opcodes.OP_ASSIGN {
			// Simple assignment: $var = value
			// emit(opcode, op1Type, op1, op2Type, op2, resultType, result)
			c.emit(opcodes.OP_ASSIGN, opcodes.IS_TMP_VAR, valueResult, opcodes.IS_UNUSED, 0, opcodes.IS_VAR, varSlot)
		} else {
			// Compound assignment: $var += value, $var *= value, etc.
			// These need special handling as they read and write the variable
			// emit(opcode, op1Type, op1, op2Type, op2, resultType, result)
			c.emit(opcode, opcodes.IS_VAR, varSlot, opcodes.IS_TMP_VAR, valueResult, opcodes.IS_VAR, varSlot)
		}
	}

	return nil
}

func (c *SimpleCompiler) compileVariable(expr *ast.Variable) error {
	varSlot := c.getVariableSlot(expr.Name)
	result := c.allocateTemp()
	c.emit(opcodes.OP_FETCH_R, opcodes.IS_TMP_VAR, result, opcodes.IS_VAR, varSlot, 0, 0)
	return nil
}

func (c *SimpleCompiler) compileNumberLiteral(expr *ast.NumberLiteral) error {
	// Parse the string value to int64
	var val int64 = 0
	if expr.Value != "" {
		// Simple parsing - in reality you'd want proper number parsing
		for _, r := range expr.Value {
			if r >= '0' && r <= '9' {
				val = val*10 + int64(r-'0')
			}
		}
	}
	
	constant := c.addConstant(values.NewInt(val))
	result := c.allocateTemp()
	// emit(opcode, op1Type, op1, op2Type, op2, resultType, result)
	c.emit(opcodes.OP_QM_ASSIGN, opcodes.IS_CONST, constant, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, result)
	return nil
}

func (c *SimpleCompiler) compileStringLiteral(expr *ast.StringLiteral) error {
	constant := c.addConstant(values.NewString(expr.Value))
	result := c.allocateTemp()
	// emit(opcode, op1Type, op1, op2Type, op2, resultType, result)
	c.emit(opcodes.OP_QM_ASSIGN, opcodes.IS_CONST, constant, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, result)
	return nil
}

func (c *SimpleCompiler) compileBooleanLiteral(expr *ast.BooleanLiteral) error {
	constant := c.addConstant(values.NewBool(expr.Value))
	result := c.allocateTemp()
	// emit(opcode, op1Type, op1, op2Type, op2, resultType, result)
	c.emit(opcodes.OP_QM_ASSIGN, opcodes.IS_CONST, constant, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, result)
	return nil
}

func (c *SimpleCompiler) compileNullLiteral(expr *ast.NullLiteral) error {
	constant := c.addConstant(values.NewNull())
	result := c.allocateTemp()
	// emit(opcode, op1Type, op1, op2Type, op2, resultType, result)
	c.emit(opcodes.OP_QM_ASSIGN, opcodes.IS_CONST, constant, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, result)
	return nil
}

func (c *SimpleCompiler) compileIdentifier(expr *ast.IdentifierNode) error {
	// Handle PHP constants like true, false, null
	var constant *values.Value
	switch expr.Name {
	case "true":
		constant = values.NewBool(true)
	case "false":
		constant = values.NewBool(false)
	case "null":
		constant = values.NewNull()
	default:
		// For other identifiers, this might be a constant or class name
		// For now, treat as string literal
		constant = values.NewString(expr.Name)
	}
	
	constantIndex := c.addConstant(constant)
	result := c.allocateTemp()
	// emit(opcode, op1Type, op1, op2Type, op2, resultType, result)
	c.emit(opcodes.OP_QM_ASSIGN, opcodes.IS_CONST, constantIndex, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, result)
	return nil
}

func (c *SimpleCompiler) compileExpressionStatement(stmt *ast.ExpressionStatement) error {
	return c.CompileNode(stmt.Expression)
}

func (c *SimpleCompiler) compileEchoStatement(stmt *ast.EchoStatement) error {
	// EchoStatement has Arguments field containing ArgumentList
	if stmt.Arguments != nil && len(stmt.Arguments.Arguments) > 0 {
		// Compile each argument and emit ECHO for each
		for _, arg := range stmt.Arguments.Arguments {
			// Compile the argument expression
			err := c.CompileNode(arg)
			if err != nil {
				return err
			}
			
			// Get the result of the compiled expression
			// The last allocated temp should contain the result
			result := c.nextTemp - 1
			c.emit(opcodes.OP_ECHO, opcodes.IS_TMP_VAR, result, 0, 0, 0, 0)
		}
	}
	return nil
}

func (c *SimpleCompiler) compileProgram(program *ast.Program) error {
	// Program has Body field, not Statements
	for _, stmt := range program.Body {
		err := c.CompileNode(stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

// Helper methods

func (c *SimpleCompiler) emit(opcode opcodes.Opcode, op1Type opcodes.OpType, op1 uint32, op2Type opcodes.OpType, op2 uint32, resultType opcodes.OpType, result uint32) {
	opType1, opType2 := opcodes.EncodeOpTypes(op1Type, op2Type, resultType)

	instruction := opcodes.Instruction{
		Opcode:  opcode,
		OpType1: opType1,
		OpType2: opType2,
		Op1:     op1,
		Op2:     op2,
		Result:  result,
	}

	c.instructions = append(c.instructions, instruction)
}

func (c *SimpleCompiler) addConstant(value *values.Value) uint32 {
	c.constants = append(c.constants, value)
	return uint32(len(c.constants) - 1)
}

func (c *SimpleCompiler) allocateTemp() uint32 {
	temp := c.nextTemp
	c.nextTemp++
	return temp
}

func (c *SimpleCompiler) getVariableSlot(name string) uint32 {
	// Simplified variable management - in a real implementation,
	// this would use proper scope management
	return uint32(len(name)) // Just use string length as slot for demo
}

func (c *SimpleCompiler) getOpcodeForOperator(operator string) opcodes.Opcode {
	switch operator {
	// Arithmetic operators
	case "+":
		return opcodes.OP_ADD
	case "-":
		return opcodes.OP_SUB
	case "*":
		return opcodes.OP_MUL
	case "/":
		return opcodes.OP_DIV
	case "%":
		return opcodes.OP_MOD
	case "**":
		return opcodes.OP_POW
	
	// String operators
	case ".":
		return opcodes.OP_CONCAT
	
	// Comparison operators
	case "==":
		return opcodes.OP_IS_EQUAL
	case "!=", "<>":
		return opcodes.OP_IS_NOT_EQUAL
	case "===":
		return opcodes.OP_IS_IDENTICAL
	case "!==":
		return opcodes.OP_IS_NOT_IDENTICAL
	case "<":
		return opcodes.OP_IS_SMALLER
	case "<=":
		return opcodes.OP_IS_SMALLER_OR_EQUAL
	case ">":
		return opcodes.OP_IS_GREATER
	case ">=":
		return opcodes.OP_IS_GREATER_OR_EQUAL
	case "<=>":
		return opcodes.OP_SPACESHIP
	
	// Logical operators
	case "&&":
		return opcodes.OP_BOOLEAN_AND
	case "||":
		return opcodes.OP_BOOLEAN_OR
	case "and":
		return opcodes.OP_LOGICAL_AND
	case "or":
		return opcodes.OP_LOGICAL_OR
	case "xor":
		return opcodes.OP_LOGICAL_XOR
	
	// Bitwise operators
	case "&":
		return opcodes.OP_BW_AND
	case "|":
		return opcodes.OP_BW_OR
	case "^":
		return opcodes.OP_BW_XOR
	case "<<":
		return opcodes.OP_SL
	case ">>":
		return opcodes.OP_SR
		
	// Instance check
	case "instanceof":
		return opcodes.OP_INSTANCEOF
	
	default:
		return opcodes.OP_NOP
	}
}

func (c *SimpleCompiler) getOpcodeForUnaryOperator(operator string) opcodes.Opcode {
	switch operator {
	// Unary arithmetic
	case "+":
		return opcodes.OP_PLUS
	case "-":
		return opcodes.OP_MINUS
		
	// Logical operators
	case "!":
		return opcodes.OP_NOT
		
	// Bitwise operators
	case "~":
		return opcodes.OP_BW_NOT
		
	// Increment/Decrement operators
	case "++":
		return opcodes.OP_PRE_INC  // This will need context to determine pre vs post
	case "--":
		return opcodes.OP_PRE_DEC  // This will need context to determine pre vs post
		
	default:
		return opcodes.OP_NOP
	}
}

func (c *SimpleCompiler) getOpcodeForAssignmentOperator(operator string) opcodes.Opcode {
	switch operator {
	// Simple assignment
	case "=":
		return opcodes.OP_ASSIGN
	case "=&":
		return opcodes.OP_ASSIGN_REF
		
	// Arithmetic compound assignments
	case "+=":
		return opcodes.OP_ASSIGN_ADD
	case "-=":
		return opcodes.OP_ASSIGN_SUB
	case "*=":
		return opcodes.OP_ASSIGN_MUL
	case "/=":
		return opcodes.OP_ASSIGN_DIV
	case "%=":
		return opcodes.OP_ASSIGN_MOD
	case "**=":
		return opcodes.OP_ASSIGN_POW
		
	// String assignment
	case ".=":
		return opcodes.OP_ASSIGN_CONCAT
		
	// Bitwise compound assignments
	case "&=":
		return opcodes.OP_ASSIGN_BW_AND
	case "|=":
		return opcodes.OP_ASSIGN_BW_OR
	case "^=":
		return opcodes.OP_ASSIGN_BW_XOR
	case "<<=":
		return opcodes.OP_ASSIGN_SL
	case ">>=":
		return opcodes.OP_ASSIGN_SR
		
	// Null coalescing assignment
	case "??=":
		return opcodes.OP_ASSIGN_COALESCE
		
	default:
		return opcodes.OP_ASSIGN
	}
}

// GetBytecode returns the compiled bytecode
func (c *SimpleCompiler) GetBytecode() []opcodes.Instruction {
	return c.instructions
}

// GetConstants returns the constant pool
func (c *SimpleCompiler) GetConstants() []*values.Value {
	return c.constants
}