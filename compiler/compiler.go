package compiler

import (
	"fmt"
	"strconv"

	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/compiler/opcodes"
	"github.com/wudi/php-parser/compiler/values"
)

// ForwardJump represents a jump that needs to be resolved later
type ForwardJump struct {
	instructionIndex int
	opType           opcodes.OpType
	operand          int // 1 for Op1, 2 for Op2
}

// Compiler compiles AST to bytecode
type Compiler struct {
	instructions    []opcodes.Instruction
	constants       []*values.Value
	scopes          []*Scope
	labels          map[string]int
	forwardJumps    map[string][]ForwardJump
	nextTemp        uint32
	nextLabel       int
}

// Scope represents a compilation scope (function, block, etc.)
type Scope struct {
	variables     map[string]uint32 // variable name -> slot
	parent        *Scope
	nextSlot      uint32
	isFunction    bool
	breakLabel    string
	continueLabel string
}

// NewCompiler creates a new bytecode compiler
func NewCompiler() *Compiler {
	return &Compiler{
		instructions: make([]opcodes.Instruction, 0),
		constants:    make([]*values.Value, 0),
		scopes:       make([]*Scope, 0),
		labels:       make(map[string]int),
		forwardJumps: make(map[string][]ForwardJump),
		nextTemp:     1000, // Start temp vars at 1000 to avoid conflicts
		nextLabel:    0,
	}
}

// Compile compiles an AST node to bytecode
func (c *Compiler) Compile(node ast.Node) error {
	// Create global scope
	c.pushScope(false)

	err := c.compileNode(node)
	if err != nil {
		return err
	}

	// Add final return if needed
	if len(c.instructions) == 0 || c.instructions[len(c.instructions)-1].Opcode != opcodes.OP_RETURN {
		c.emit(opcodes.OP_RETURN, opcodes.IS_CONST, c.addConstant(values.NewNull()), 0, 0, 0, 0)
	}

	c.popScope()
	return nil
}

// GetBytecode returns the compiled bytecode
func (c *Compiler) GetBytecode() []opcodes.Instruction {
	return c.instructions
}

// GetConstants returns the constant pool
func (c *Compiler) GetConstants() []*values.Value {
	return c.constants
}

// Main compilation dispatcher
func (c *Compiler) compileNode(node ast.Node) error {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	// Expressions
	case *ast.BinaryExpression:
		return c.compileBinaryOp(n)
	case *ast.UnaryExpression:
		return c.compileUnaryOp(n)
	case *ast.AssignmentExpression:
		return c.compileAssign(n)
	case *ast.Variable:
		return c.compileVariable(n)
	case *ast.IdentifierNode:
		return c.compileIdentifier(n)
	case *ast.NumberLiteral:
		return c.compileNumberLiteral(n)
	case *ast.StringLiteral:
		return c.compileStringLiteral(n)
	case *ast.BooleanLiteral:
		return c.compileBooleanLiteral(n)
	case *ast.NullLiteral:
		return c.compileNullLiteral(n)
	case *ast.ArrayExpression:
		return c.compileArray(n)
	case *ast.ArrayAccessExpression:
		return c.compileArrayAccess(n)
	case *ast.PropertyAccessExpression:
		return c.compilePropertyAccess(n)
	case *ast.CallExpression:
		return c.compileFunctionCall(n)
	case *ast.MethodCallExpression:
		return c.compileMethodCall(n)
	case *ast.TernaryExpression:
		return c.compileTernary(n)
	case *ast.CoalesceExpression:
		return c.compileCoalesce(n)
	case *ast.MatchExpression:
		return c.compileMatch(n)

	// Statements
	case *ast.ExpressionStatement:
		return c.compileExpressionStatement(n)
	case *ast.EchoStatement:
		return c.compileEcho(n)
	case *ast.ReturnStatement:
		return c.compileReturn(n)
	case *ast.IfStatement:
		return c.compileIf(n)
	case *ast.WhileStatement:
		return c.compileWhile(n)
	case *ast.ForStatement:
		return c.compileFor(n)
	case *ast.ForeachStatement:
		return c.compileForeach(n)
	case *ast.SwitchStatement:
		return c.compileSwitch(n)
	case *ast.BreakStatement:
		return c.compileBreak(n)
	case *ast.ContinueStatement:
		return c.compileContinue(n)
	case *ast.TryStatement:
		return c.compileTry(n)
	case *ast.ThrowStatement:
		return c.compileThrow(n)
	case *ast.BlockStatement:
		return c.compileBlock(n)

	// Declarations
	case *ast.FunctionDeclaration:
		return c.compileFunctionDeclaration(n)
	case *ast.AnonymousClass:
		return c.compileClassDeclaration(n)
	case *ast.PropertyDeclaration:
		return c.compilePropertyDeclaration(n)
	case *ast.ClassConstantDeclaration:
		return c.compileClassConstant(n)

	// Program node
	case *ast.Program:
		return c.compileProgram(n)

	default:
		return fmt.Errorf("unsupported AST node type: %T", n)
	}
}

// Expression compilation methods

func (c *Compiler) compileBinaryOp(expr *ast.BinaryExpression) error {
	// Save current temp counter
	startTemp := c.nextTemp

	// Compile left operand
	err := c.compileNode(expr.Left)
	if err != nil {
		return err
	}
	leftResult := startTemp

	// Compile right operand  
	err = c.compileNode(expr.Right)
	if err != nil {
		return err
	}
	rightResult := startTemp + 1

	// Generate operation
	result := c.allocateTemp()
	opcode := c.getOpcodeForBinaryOperator(expr.Operator)
	c.emit(opcode, opcodes.IS_TMP_VAR, leftResult, opcodes.IS_TMP_VAR, rightResult, opcodes.IS_TMP_VAR, result)

	return nil
}

func (c *Compiler) compileUnaryOp(expr *ast.UnaryExpression) error {
	// Handle increment/decrement operations specially
	if expr.Operator == "++" || expr.Operator == "--" {
		return c.compileIncrementDecrement(expr)
	}

	// Handle regular unary operations
	err := c.compileNode(expr.Operand)
	if err != nil {
		return err
	}

	operandResult := c.allocateTemp()
	c.emitMove(operandResult)

	result := c.allocateTemp()
	opcode := c.getOpcodeForUnaryOperator(expr.Operator)
	c.emit(opcode, opcodes.IS_TMP_VAR, result, opcodes.IS_TMP_VAR, operandResult, 0, 0)

	return nil
}

func (c *Compiler) compileIncrementDecrement(expr *ast.UnaryExpression) error {
	// Increment/decrement only works on variables
	variable, ok := expr.Operand.(*ast.Variable)
	if !ok {
		return fmt.Errorf("increment/decrement can only be applied to variables")
	}

	varSlot := c.getVariableSlot(variable.Name)
	
	// Read current value from variable
	currentVal := c.allocateTemp()
	c.emit(opcodes.OP_FETCH_R, opcodes.IS_VAR, varSlot, 0, 0, opcodes.IS_TMP_VAR, currentVal)

	// Create constant 1 for increment/decrement
	oneConstant := c.addConstant(values.NewInt(1))
	
	// Calculate new value
	newVal := c.allocateTemp()
	if expr.Operator == "++" {
		c.emit(opcodes.OP_ADD, opcodes.IS_TMP_VAR, currentVal, opcodes.IS_CONST, oneConstant, opcodes.IS_TMP_VAR, newVal)
	} else { // "--"
		c.emit(opcodes.OP_SUB, opcodes.IS_TMP_VAR, currentVal, opcodes.IS_CONST, oneConstant, opcodes.IS_TMP_VAR, newVal)
	}

	// Write new value back to variable
	c.emit(opcodes.OP_ASSIGN, opcodes.IS_TMP_VAR, newVal, opcodes.IS_UNUSED, 0, opcodes.IS_VAR, varSlot)

	// Expression result handling: 
	// For standalone increment statements, we don't need to preserve the return value
	// The variable has been modified, which is the primary effect
	return nil
}

func (c *Compiler) compileAssign(expr *ast.AssignmentExpression) error {
	// Compile right-hand side first
	err := c.compileNode(expr.Right)
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

func (c *Compiler) getOpcodeForAssignmentOperator(operator string) opcodes.Opcode {
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

func (c *Compiler) getVariableSlot(name string) uint32 {
	// Simplified variable management - in a real implementation,
	// this would use proper scope management
	// Use a hash of the name to get a consistent slot
	hash := uint32(0)
	for _, r := range name {
		hash = hash*31 + uint32(r)
	}
	return hash % 100 // Keep slots within reasonable range
}

func (c *Compiler) compileVariable(expr *ast.Variable) error {
	varSlot := c.getVariableSlot(expr.Name)
	result := c.allocateTemp()
	c.emit(opcodes.OP_FETCH_R, opcodes.IS_VAR, varSlot, 0, 0, opcodes.IS_TMP_VAR, result)
	return nil
}

func (c *Compiler) compileIdentifier(expr *ast.IdentifierNode) error {
	// Handle special literal keywords
	var constant uint32
	
	switch expr.Name {
	case "null":
		constant = c.addConstant(values.NewNull())
	case "true":
		constant = c.addConstant(values.NewBool(true))
	case "false":
		constant = c.addConstant(values.NewBool(false))
	default:
		// Identifiers are typically constants or function names
		constant = c.addConstant(values.NewString(expr.Name))
	}
	
	result := c.allocateTemp()
	c.emit(opcodes.OP_QM_ASSIGN, opcodes.IS_CONST, constant, 0, 0, opcodes.IS_TMP_VAR, result)
	return nil
}

func (c *Compiler) compileNumberLiteral(expr *ast.NumberLiteral) error {
	var constant uint32
	
	if expr.Kind == "integer" {
		value, err := strconv.ParseInt(expr.Value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid integer literal: %s", expr.Value)
		}
		constant = c.addConstant(values.NewInt(value))
	} else if expr.Kind == "float" {
		value, err := strconv.ParseFloat(expr.Value, 64)
		if err != nil {
			return fmt.Errorf("invalid float literal: %s", expr.Value)
		}
		constant = c.addConstant(values.NewFloat(value))
	} else {
		return fmt.Errorf("unsupported number kind: %s", expr.Kind)
	}

	result := c.allocateTemp()
	c.emit(opcodes.OP_QM_ASSIGN, opcodes.IS_CONST, constant, 0, 0, opcodes.IS_TMP_VAR, result)
	return nil
}

func (c *Compiler) compileStringLiteral(expr *ast.StringLiteral) error {
	constant := c.addConstant(values.NewString(expr.Value))
	result := c.allocateTemp()
	c.emit(opcodes.OP_QM_ASSIGN, opcodes.IS_CONST, constant, 0, 0, opcodes.IS_TMP_VAR, result)
	return nil
}

func (c *Compiler) compileBooleanLiteral(expr *ast.BooleanLiteral) error {
	constant := c.addConstant(values.NewBool(expr.Value))
	result := c.allocateTemp()
	c.emit(opcodes.OP_QM_ASSIGN, opcodes.IS_CONST, constant, 0, 0, opcodes.IS_TMP_VAR, result)
	return nil
}

func (c *Compiler) compileNullLiteral(expr *ast.NullLiteral) error {
	constant := c.addConstant(values.NewNull())
	result := c.allocateTemp()
	c.emit(opcodes.OP_QM_ASSIGN, opcodes.IS_CONST, constant, 0, 0, opcodes.IS_TMP_VAR, result)
	return nil
}

func (c *Compiler) compileArray(expr *ast.ArrayExpression) error {
	result := c.allocateTemp()
	c.emit(opcodes.OP_INIT_ARRAY, opcodes.IS_TMP_VAR, result, 0, 0, 0, 0)

	for _, element := range expr.Elements {
		arrayElement, ok := element.(*ast.ArrayElementExpression)
		if !ok {
			return fmt.Errorf("invalid array element type: %T", element)
		}
		if arrayElement.Key != nil {
			// Keyed element
			err := c.compileNode(arrayElement.Key)
			if err != nil {
				return err
			}
			keyResult := c.allocateTemp()
			c.emitMove(keyResult)

			err = c.compileNode(arrayElement.Value)
			if err != nil {
				return err
			}
			valueResult := c.allocateTemp()
			c.emitMove(valueResult)

			c.emit(opcodes.OP_ADD_ARRAY_ELEMENT, opcodes.IS_TMP_VAR, result, opcodes.IS_TMP_VAR, keyResult, opcodes.IS_TMP_VAR, valueResult)
		} else {
			// Auto-indexed element
			err := c.compileNode(arrayElement.Value)
			if err != nil {
				return err
			}
			valueResult := c.allocateTemp()
			c.emitMove(valueResult)

			c.emit(opcodes.OP_ADD_ARRAY_ELEMENT, opcodes.IS_TMP_VAR, result, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, valueResult)
		}
	}

	return nil
}

func (c *Compiler) compileArrayAccess(expr *ast.ArrayAccessExpression) error {
	// Compile array expression
	err := c.compileNode(expr.Array)
	if err != nil {
		return err
	}
	arrayResult := c.allocateTemp()
	c.emitMove(arrayResult)

	// Compile index expression
	if expr.Index == nil {
		return fmt.Errorf("array access requires an index")
	}
	err = c.compileNode(*expr.Index)
	if err != nil {
		return err
	}
	indexResult := c.allocateTemp()
	c.emitMove(indexResult)

	result := c.allocateTemp()
	c.emit(opcodes.OP_FETCH_DIM_R, opcodes.IS_TMP_VAR, result, opcodes.IS_TMP_VAR, arrayResult, opcodes.IS_TMP_VAR, indexResult)

	return nil
}

func (c *Compiler) compilePropertyAccess(expr *ast.PropertyAccessExpression) error {
	// Compile object expression
	err := c.compileNode(expr.Object)
	if err != nil {
		return err
	}
	objectResult := c.allocateTemp()
	c.emitMove(objectResult)

	// Compile property expression
	err = c.compileNode(expr.Property)
	if err != nil {
		return err
	}
	propResult := c.allocateTemp()
	c.emitMove(propResult)
	result := c.allocateTemp()
	c.emit(opcodes.OP_FETCH_OBJ_R, opcodes.IS_TMP_VAR, result, opcodes.IS_TMP_VAR, objectResult, opcodes.IS_TMP_VAR, propResult)

	return nil
}

func (c *Compiler) compileFunctionCall(expr *ast.CallExpression) error {
	// Compile callee expression
	err := c.compileNode(expr.Callee)
	if err != nil {
		return err
	}
	calleeResult := c.allocateTemp()
	c.emitMove(calleeResult)
	
	// Get number of arguments
	var numArgs uint32
	if expr.Arguments != nil {
		numArgs = uint32(len(expr.Arguments.Arguments))
	} else {
		numArgs = 0
	}

	c.emit(opcodes.OP_INIT_FCALL, opcodes.IS_TMP_VAR, calleeResult, opcodes.IS_CONST, c.addConstant(values.NewInt(int64(numArgs))), 0, 0)

	// Compile and send arguments
	if expr.Arguments != nil {
		for i, arg := range expr.Arguments.Arguments {
			err := c.compileNode(arg)
			if err != nil {
				return err
			}
			argResult := c.allocateTemp()
			c.emitMove(argResult)

			argNum := c.addConstant(values.NewInt(int64(i)))
			c.emit(opcodes.OP_SEND_VAL, opcodes.IS_CONST, argNum, opcodes.IS_TMP_VAR, argResult, 0, 0)
		}
	}

	// Execute call
	result := c.allocateTemp()
	c.emit(opcodes.OP_DO_FCALL, opcodes.IS_TMP_VAR, result, 0, 0, 0, 0)

	return nil
}

func (c *Compiler) compileMethodCall(expr *ast.MethodCallExpression) error {
	// Compile object
	err := c.compileNode(expr.Object)
	if err != nil {
		return err
	}
	objectResult := c.allocateTemp()
	c.emitMove(objectResult)

	// Compile method name
	err = c.compileNode(expr.Method)
	if err != nil {
		return err
	}
	methodResult := c.allocateTemp()
	c.emitMove(methodResult)
	
	// Get number of arguments
	var numArgs uint32
	if expr.Arguments != nil {
		numArgs = uint32(len(expr.Arguments.Arguments))
	} else {
		numArgs = 0
	}

	c.emit(opcodes.OP_INIT_METHOD_CALL, opcodes.IS_TMP_VAR, objectResult, opcodes.IS_TMP_VAR, methodResult, opcodes.IS_CONST, c.addConstant(values.NewInt(int64(numArgs))))

	// Compile and send arguments
	if expr.Arguments != nil {
		for i, arg := range expr.Arguments.Arguments {
			err := c.compileNode(arg)
			if err != nil {
				return err
			}
			argResult := c.allocateTemp()
			c.emitMove(argResult)

			argNum := c.addConstant(values.NewInt(int64(i)))
			c.emit(opcodes.OP_SEND_VAL, opcodes.IS_CONST, argNum, opcodes.IS_TMP_VAR, argResult, 0, 0)
		}
	}

	// Execute call
	result := c.allocateTemp()
	c.emit(opcodes.OP_DO_FCALL, opcodes.IS_TMP_VAR, result, 0, 0, 0, 0)

	return nil
}

func (c *Compiler) compileTernary(expr *ast.TernaryExpression) error {
	// Compile condition
	err := c.compileNode(expr.Test)
	if err != nil {
		return err
	}
	condResult := c.allocateTemp()
	c.emitMove(condResult)

	// Jump labels
	elseLabel := c.generateLabel()
	endLabel := c.generateLabel()

	// Jump to else if condition is false
	c.emitJumpZ(opcodes.IS_TMP_VAR, condResult, elseLabel)

	// Compile true branch
	if expr.Consequent != nil {
		err = c.compileNode(expr.Consequent)
	} else {
		// Short ternary - use condition result
		err = c.compileNode(expr.Test)
	}
	if err != nil {
		return err
	}
	trueResult := c.allocateTemp()
	c.emitMove(trueResult)

	// Jump to end
	c.emitJump(opcodes.OP_JMP, opcodes.IS_CONST, 0, endLabel)

	// Else branch
	c.placeLabel(elseLabel)
	err = c.compileNode(expr.Alternate)
	if err != nil {
		return err
	}
	falseResult := c.allocateTemp()
	c.emitMove(falseResult)

	// End label
	c.placeLabel(endLabel)

	// Result assignment (this is simplified - real implementation would be more complex)
	result := c.allocateTemp()
	c.emit(opcodes.OP_QM_ASSIGN, opcodes.IS_TMP_VAR, result, opcodes.IS_TMP_VAR, trueResult, 0, 0)

	return nil
}

// Statement compilation methods

func (c *Compiler) compileExpressionStatement(stmt *ast.ExpressionStatement) error {
	return c.compileNode(stmt.Expression)
}

func (c *Compiler) compileEcho(stmt *ast.EchoStatement) error {
	if stmt.Arguments != nil {
		for _, expr := range stmt.Arguments.Arguments {
			err := c.compileNode(expr)
			if err != nil {
				return err
			}
			// Use the most recently allocated temp as the result
			result := c.nextTemp - 1
			c.emit(opcodes.OP_ECHO, opcodes.IS_TMP_VAR, result, 0, 0, 0, 0)
		}
	}
	return nil
}

func (c *Compiler) compileReturn(stmt *ast.ReturnStatement) error {
	if stmt.Argument != nil {
		err := c.compileNode(stmt.Argument)
		if err != nil {
			return err
		}
		result := c.allocateTemp()
		c.emitMove(result)
		c.emit(opcodes.OP_RETURN, opcodes.IS_TMP_VAR, result, 0, 0, 0, 0)
	} else {
		nullConstant := c.addConstant(values.NewNull())
		c.emit(opcodes.OP_RETURN, opcodes.IS_CONST, nullConstant, 0, 0, 0, 0)
	}
	return nil
}

func (c *Compiler) compileIf(stmt *ast.IfStatement) error {
	// Compile condition
	err := c.compileNode(stmt.Test)
	if err != nil {
		return err
	}
	// The comparison result should be in the most recently allocated temp
	condResult := c.nextTemp - 1

	// Generate labels
	elseLabel := c.generateLabel()
	endLabel := c.generateLabel()

	// Jump to else if condition is false
	c.emitJumpZ(opcodes.IS_TMP_VAR, condResult, elseLabel)

	// Compile consequence
	for _, s := range stmt.Consequent {
		err = c.compileNode(s)
		if err != nil {
			return err
		}
	}

	// Jump to end
	c.emitJump(opcodes.OP_JMP, opcodes.IS_CONST, 0, endLabel)

	// Else branch
	c.placeLabel(elseLabel)
	if len(stmt.Alternate) > 0 {
		for _, s := range stmt.Alternate {
			err = c.compileNode(s)
			if err != nil {
				return err
			}
		}
	}

	// End label
	c.placeLabel(endLabel)

	return nil
}

func (c *Compiler) compileWhile(stmt *ast.WhileStatement) error {
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
	err := c.compileNode(stmt.Test)
	if err != nil {
		return err
	}
	condResult := c.allocateTemp()
	c.emitMove(condResult)

	// Jump to end if condition is false
	c.emitJumpZ(opcodes.IS_TMP_VAR, condResult, endLabel)

	// Compile body
	for _, s := range stmt.Body {
		err = c.compileNode(s)
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

func (c *Compiler) compileProgram(program *ast.Program) error {
	for _, stmt := range program.Body {
		err := c.compileNode(stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

// Helper methods

func (c *Compiler) emit(opcode opcodes.Opcode, op1Type opcodes.OpType, op1 uint32, op2Type opcodes.OpType, op2 uint32, resultType opcodes.OpType, result uint32) {
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

func (c *Compiler) emitMove(target uint32) {
	// This is a placeholder for moving the top of stack to target
	// In a real implementation, this would be more sophisticated
}

func (c *Compiler) addConstant(value *values.Value) uint32 {
	c.constants = append(c.constants, value)
	return uint32(len(c.constants) - 1)
}

func (c *Compiler) allocateTemp() uint32 {
	temp := c.nextTemp
	c.nextTemp++
	return temp
}

func (c *Compiler) getLastAllocatedTemp() uint32 {
	return c.nextTemp - 1
}

func (c *Compiler) generateLabel() string {
	label := fmt.Sprintf("L%d", c.nextLabel)
	c.nextLabel++
	return label
}

func (c *Compiler) addLabel(name string) uint32 {
	// Check if label is already placed
	if pos, exists := c.labels[name]; exists {
		return uint32(pos)
	}
	
	// Return a placeholder value (we'll use the label name as a constant)
	// This will be resolved when the label is placed
	return uint32(0xFFFF) // Placeholder value that will be patched
}

func (c *Compiler) addForwardJump(instructionIndex int, labelName string, operand int) {
	jump := ForwardJump{
		instructionIndex: instructionIndex,
		opType:           opcodes.IS_CONST,
		operand:          operand,
	}
	c.forwardJumps[labelName] = append(c.forwardJumps[labelName], jump)
}

func (c *Compiler) placeLabel(name string) {
	pos := len(c.instructions)
	c.labels[name] = pos
	
	// Resolve all forward jumps to this label
	if jumps, exists := c.forwardJumps[name]; exists {
		for _, jump := range jumps {
			if jump.operand == 0 {
				// Update constant value (for new jump system)
				constantIndex := jump.instructionIndex
				c.constants[constantIndex] = values.NewInt(int64(pos))
			} else {
				// Update instruction operand (for old jump system)
				instruction := &c.instructions[jump.instructionIndex]
				if jump.operand == 1 {
					instruction.Op1 = uint32(pos)
				} else if jump.operand == 2 {
					instruction.Op2 = uint32(pos)
				}
			}
		}
		delete(c.forwardJumps, name)
	}
}

// Helper function to emit unconditional jump with forward reference
func (c *Compiler) emitJump(opcode opcodes.Opcode, op1Type opcodes.OpType, op1 uint32, labelName string) {
	// Add a placeholder constant for the jump target
	jumpConstant := c.addConstant(values.NewInt(0)) // Will be updated later
	
	// For unconditional jumps, the target goes in Op1
	// For conditional jumps, condition is Op1 and target is Op2
	if opcode == opcodes.OP_JMP {
		c.emit(opcode, opcodes.IS_CONST, jumpConstant, 0, 0, 0, 0)
	} else {
		c.emit(opcode, op1Type, op1, opcodes.IS_CONST, jumpConstant, 0, 0)
	}
	
	// Record forward jump for later resolution - need to update the constant, not the instruction
	jump := ForwardJump{
		instructionIndex: int(jumpConstant), // Store constant index instead of instruction index
		opType:           opcodes.IS_CONST,
		operand:          0, // Special marker for constant update
	}
	c.forwardJumps[labelName] = append(c.forwardJumps[labelName], jump)
}

// Helper function to emit conditional jump with forward reference  
func (c *Compiler) emitJumpZ(condType opcodes.OpType, cond uint32, labelName string) {
	// Add a placeholder constant for the jump target
	jumpConstant := c.addConstant(values.NewInt(0)) // Will be updated later
	
	// Emit instruction with constant reference for label
	c.emit(opcodes.OP_JMPZ, condType, cond, opcodes.IS_CONST, jumpConstant, 0, 0)
	
	// Record forward jump for later resolution - need to update the constant, not the instruction
	jump := ForwardJump{
		instructionIndex: int(jumpConstant), // Store constant index instead of instruction index
		opType:           opcodes.IS_CONST,
		operand:          0, // Special marker for constant update
	}
	c.forwardJumps[labelName] = append(c.forwardJumps[labelName], jump)
}

func (c *Compiler) emitJumpNZ(condType opcodes.OpType, cond uint32, labelName string) {
	// Add a placeholder constant for the jump target
	jumpConstant := c.addConstant(values.NewInt(0)) // Will be updated later
	
	// Emit instruction with constant reference for label
	c.emit(opcodes.OP_JMPNZ, condType, cond, opcodes.IS_CONST, jumpConstant, 0, 0)
	
	// Record forward jump for later resolution - need to update the constant, not the instruction
	jump := ForwardJump{
		instructionIndex: int(jumpConstant), // Store constant index instead of instruction index
		opType:           opcodes.IS_CONST,
		operand:          0, // Special marker for constant update
	}
	c.forwardJumps[labelName] = append(c.forwardJumps[labelName], jump)
}

func (c *Compiler) pushScope(isFunction bool) {
	scope := &Scope{
		variables:  make(map[string]uint32),
		parent:     c.currentScope(),
		nextSlot:   0,
		isFunction: isFunction,
	}
	c.scopes = append(c.scopes, scope)
}

func (c *Compiler) popScope() {
	if len(c.scopes) > 0 {
		c.scopes = c.scopes[:len(c.scopes)-1]
	}
}

func (c *Compiler) currentScope() *Scope {
	if len(c.scopes) == 0 {
		return nil
	}
	return c.scopes[len(c.scopes)-1]
}

func (c *Compiler) getOrCreateVariable(name string) uint32 {
	scope := c.currentScope()
	if scope == nil {
		// Create global scope
		c.pushScope(false)
		scope = c.currentScope()
	}

	if slot, exists := scope.variables[name]; exists {
		return slot
	}

	slot := scope.nextSlot
	scope.variables[name] = slot
	scope.nextSlot++
	return slot
}

// Operator mapping methods

func (c *Compiler) getOpcodeForBinaryOperator(operator string) opcodes.Opcode {
	switch operator {
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
	case ".":
		return opcodes.OP_CONCAT
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
	default:
		return opcodes.OP_NOP
	}
}

func (c *Compiler) getOpcodeForUnaryOperator(operator string) opcodes.Opcode {
	switch operator {
	case "+":
		return opcodes.OP_PLUS
	case "-":
		return opcodes.OP_MINUS
	case "!":
		return opcodes.OP_NOT
	case "~":
		return opcodes.OP_BW_NOT
	case "++":
		return opcodes.OP_PRE_INC
	case "--":
		return opcodes.OP_PRE_DEC
	default:
		return opcodes.OP_NOP
	}
}

// Placeholder implementations for missing methods

func (c *Compiler) compileCoalesce(expr *ast.CoalesceExpression) error {
	// Compile left operand
	err := c.compileNode(expr.Left)
	if err != nil {
		return err
	}
	leftResult := c.nextTemp - 1

	// Generate labels for control flow
	rightLabel := c.generateLabel()
	endLabel := c.generateLabel()

	// Check if left operand is null - if null, jump to right operand
	nullConstant := c.addConstant(values.NewNull())
	compResult := c.allocateTemp()
	
	// Compare left with null (using identical comparison for precise null check)
	c.emit(opcodes.OP_IS_IDENTICAL, opcodes.IS_TMP_VAR, leftResult, opcodes.IS_CONST, nullConstant, opcodes.IS_TMP_VAR, compResult)
	
	// If left is null (comparison is true), jump to evaluate right operand
	c.emitJumpNZ(opcodes.IS_TMP_VAR, compResult, rightLabel)
	
	// Left is not null - we need to ensure the result is in the final temp position
	// We'll allocate a temp for the result and copy the left value into it
	result := c.allocateTemp()
	c.emit(opcodes.OP_QM_ASSIGN, opcodes.IS_TMP_VAR, leftResult, 0, 0, opcodes.IS_TMP_VAR, result)
	c.emitJump(opcodes.OP_JMP, opcodes.IS_CONST, 0, endLabel)
	
	// Right operand evaluation (when left is null)
	c.placeLabel(rightLabel)
	err = c.compileNode(expr.Right)
	if err != nil {
		return err
	}
	rightResult := c.nextTemp - 1
	
	// Since we already allocated result temp above, we need to ensure both branches use the same temp
	// Copy right result to the same result temp we allocated above
	c.emit(opcodes.OP_QM_ASSIGN, opcodes.IS_TMP_VAR, rightResult, 0, 0, opcodes.IS_TMP_VAR, result)
	
	// End label
	c.placeLabel(endLabel)
	
	// Ensure the result is in the final temp position for echo to find it
	// This handles the issue where compiling the right expression allocates additional temps
	finalResult := c.allocateTemp()
	c.emit(opcodes.OP_QM_ASSIGN, opcodes.IS_TMP_VAR, result, 0, 0, opcodes.IS_TMP_VAR, finalResult)

	return nil
}

func (c *Compiler) compileMatch(expr *ast.MatchExpression) error {
	// TODO: Implement match expression
	return fmt.Errorf("match expression not implemented")
}

func (c *Compiler) compileFor(stmt *ast.ForStatement) error {
	// TODO: Implement for loop
	return fmt.Errorf("for loop not implemented")
}

func (c *Compiler) compileForeach(stmt *ast.ForeachStatement) error {
	// TODO: Implement foreach loop
	return fmt.Errorf("foreach loop not implemented")
}

func (c *Compiler) compileSwitch(stmt *ast.SwitchStatement) error {
	// Compile the discriminant (the switch expression)
	err := c.compileNode(stmt.Discriminant)
	if err != nil {
		return err
	}
	discriminantTemp := c.nextTemp - 1

	// Create labels for the switch
	endLabel := c.generateLabel()
	var defaultLabel string
	var caseLabels []string
	
	// Generate labels for each case
	for i := 0; i < len(stmt.Cases); i++ {
		if stmt.Cases[i].Test == nil {
			// Default case
			defaultLabel = c.generateLabel()
			caseLabels = append(caseLabels, defaultLabel)
		} else {
			caseLabels = append(caseLabels, c.generateLabel())
		}
	}
	
	// Push new scope for break statements
	c.pushScope(false)
	c.currentScope().breakLabel = endLabel

	// Sequential case comparisons - evaluate each case and jump if equal
	for i, switchCase := range stmt.Cases {
		if switchCase.Test != nil {
			// Compile the case value
			err := c.compileNode(switchCase.Test)
			if err != nil {
				c.popScope()
				return err
			}
			caseValueTemp := c.nextTemp - 1
			
			// Compare discriminant == case value using loose comparison
			c.allocateTemp() // For comparison result
			compResultTemp := c.nextTemp - 1
			c.emit(opcodes.OP_IS_EQUAL, opcodes.IS_TMP_VAR, discriminantTemp, 
				   opcodes.IS_TMP_VAR, caseValueTemp, opcodes.IS_TMP_VAR, compResultTemp)
			
			// Jump to case if equal
			c.emitJumpNZ(opcodes.IS_TMP_VAR, compResultTemp, caseLabels[i])
		}
	}
	
	// If no case matched, jump to default (if exists) or end
	if defaultLabel != "" {
		c.emitJump(opcodes.OP_JMP, opcodes.IS_CONST, 0, defaultLabel)
	} else {
		c.emitJump(opcodes.OP_JMP, opcodes.IS_CONST, 0, endLabel)
	}
	
	// Emit case bodies
	for i, switchCase := range stmt.Cases {
		c.placeLabel(caseLabels[i])
		
		// Compile case body
		for _, stmt := range switchCase.Body {
			err := c.compileNode(stmt)
			if err != nil {
				c.popScope()
				return err
			}
		}
		// Fall-through behavior - no automatic jump to end
	}
	
	c.placeLabel(endLabel)
	c.popScope()
	
	return nil
}

func (c *Compiler) compileBreak(stmt *ast.BreakStatement) error {
	scope := c.currentScope()
	if scope == nil || scope.breakLabel == "" {
		return fmt.Errorf("break statement not in loop")
	}
	c.emitJump(opcodes.OP_JMP, opcodes.IS_CONST, 0, scope.breakLabel)
	return nil
}

func (c *Compiler) compileContinue(stmt *ast.ContinueStatement) error {
	scope := c.currentScope()
	if scope == nil || scope.continueLabel == "" {
		return fmt.Errorf("continue statement not in loop")
	}
	c.emitJump(opcodes.OP_JMP, opcodes.IS_CONST, 0, scope.continueLabel)
	return nil
}

func (c *Compiler) compileTry(stmt *ast.TryStatement) error {
	// TODO: Implement try-catch statement
	return fmt.Errorf("try-catch statement not implemented")
}

func (c *Compiler) compileThrow(stmt *ast.ThrowStatement) error {
	err := c.compileNode(stmt.Argument)
	if err != nil {
		return err
	}
	result := c.allocateTemp()
	c.emitMove(result)
	c.emit(opcodes.OP_THROW, opcodes.IS_TMP_VAR, result, 0, 0, 0, 0)
	return nil
}

func (c *Compiler) compileBlock(stmt *ast.BlockStatement) error {
	for _, s := range stmt.Body {
		err := c.compileNode(s)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiler) compileFunctionDeclaration(decl *ast.FunctionDeclaration) error {
	// TODO: Implement function declaration
	return fmt.Errorf("function declaration not implemented")
}

func (c *Compiler) compileClassDeclaration(decl *ast.AnonymousClass) error {
	// TODO: Implement class declaration
	return fmt.Errorf("class declaration not implemented")
}

func (c *Compiler) compilePropertyDeclaration(decl *ast.PropertyDeclaration) error {
	// TODO: Implement property declaration
	return fmt.Errorf("property declaration not implemented")
}

func (c *Compiler) compileClassConstant(decl *ast.ClassConstantDeclaration) error {
	// TODO: Implement class constant declaration
	return fmt.Errorf("class constant declaration not implemented")
}
