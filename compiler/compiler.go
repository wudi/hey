package compiler

import (
	"fmt"
	"strconv"

	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/compiler/opcodes"
	"github.com/wudi/php-parser/compiler/values"
	"github.com/wudi/php-parser/compiler/vm"
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
	functions       map[string]*vm.Function
	classes         map[string]*vm.Class
	currentClass    *vm.Class  // Current class being compiled
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
		functions:    make(map[string]*vm.Function),
		classes:      make(map[string]*vm.Class),
		currentClass: nil,
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

// GetFunctions returns the compiled functions
func (c *Compiler) GetFunctions() map[string]*vm.Function {
	return c.functions
}

func (c *Compiler) GetClasses() map[string]*vm.Class {
	return c.classes
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
	case *ast.VariableVariableExpression:
		return c.compileVariableVariable(n)
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
	case *ast.InterpolatedStringExpression:
		return c.compileInterpolatedString(n)
	case *ast.NewExpression:
		return c.compileNew(n)
	case *ast.ClassConstantAccessExpression:
		return c.compileClassConstantAccess(n)
	case *ast.StaticPropertyAccessExpression:
		return c.compileStaticPropertyAccess(n)
	case *ast.StaticAccessExpression:
		return c.compileStaticAccess(n)

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
	case *ast.ClassExpression:
		return c.compileRegularClassDeclaration(n)
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
	// Compile left operand
	err := c.compileNode(expr.Left)
	if err != nil {
		return err
	}
	leftResult := c.nextTemp - 1  // The last allocated temp contains the left result

	// Compile right operand  
	err = c.compileNode(expr.Right)
	if err != nil {
		return err
	}
	rightResult := c.nextTemp - 1  // The last allocated temp contains the right result

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
	// Check if it's a simple variable
	if variable, ok := expr.Operand.(*ast.Variable); ok {
		// Handle simple variables
		return c.compileSimpleIncDec(expr, variable)
	}
	
	// Check if it's a static property access
	if staticProp, ok := expr.Operand.(*ast.StaticPropertyAccessExpression); ok {
		// Handle static property increment/decrement
		return c.compileStaticPropIncDec(expr, staticProp)
	}
	
	// Check if it's a static access (like self::$counter)
	if staticAccess, ok := expr.Operand.(*ast.StaticAccessExpression); ok {
		// Handle static access increment/decrement
		return c.compileStaticAccessIncDec(expr, staticAccess)
	}
	
	return fmt.Errorf("increment/decrement can only be applied to variables or static properties")
}

func (c *Compiler) compileSimpleIncDec(expr *ast.UnaryExpression, variable *ast.Variable) error {

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

func (c *Compiler) compileStaticPropIncDec(expr *ast.UnaryExpression, staticProp *ast.StaticPropertyAccessExpression) error {
	// Implement static property increment/decrement (e.g., TestClass::$counter++)
	
	// Step 1: Compile class expression (supports both static and dynamic class names)
	var classOperandType opcodes.OpType
	var classOperand uint32
	
	switch class := staticProp.Class.(type) {
	case *ast.IdentifierNode:
		// Static class name like MyClass::$prop
		classOperand = c.addConstant(values.NewString(class.Name))
		classOperandType = opcodes.IS_CONST
	case *ast.Variable:
		// Handle self::$prop, static::$prop, parent::$prop etc
		classOperand = c.addConstant(values.NewString(class.Name))
		classOperandType = opcodes.IS_CONST
	default:
		// Dynamic class expression like $className::$prop
		err := c.compileNode(class)
		if err != nil {
			return fmt.Errorf("failed to compile class expression in static property increment: %w", err)
		}
		classOperand = c.nextTemp - 1
		classOperandType = opcodes.IS_TMP_VAR
	}
	
	// Step 2: Compile property expression 
	var propOperandType opcodes.OpType
	var propOperand uint32
	
	switch property := staticProp.Property.(type) {
	case *ast.Variable:
		// Simple static property like ::$prop
		// Strip the $ from the property name since class properties are stored without $
		propName := property.Name
		if len(propName) > 0 && propName[0] == '$' {
			propName = propName[1:]
		}
		propOperand = c.addConstant(values.NewString(propName))
		propOperandType = opcodes.IS_CONST
	default:
		// Dynamic property expression like ::${$expr} or ::${"prop"}
		err := c.compileNode(property)
		if err != nil {
			return fmt.Errorf("failed to compile property expression in static property increment: %w", err)
		}
		propOperand = c.nextTemp - 1
		propOperandType = opcodes.IS_TMP_VAR
	}
	
	// Step 3: Read current value from static property
	currentVal := c.allocateTemp()
	c.emit(opcodes.OP_FETCH_STATIC_PROP_R,
		classOperandType, classOperand,
		propOperandType, propOperand,
		opcodes.IS_TMP_VAR, currentVal)
	
	// Step 4: Create constant 1 for increment/decrement
	oneConstant := c.addConstant(values.NewInt(1))
	
	// Step 5: Calculate new value
	newVal := c.allocateTemp()
	if expr.Operator == "++" {
		c.emit(opcodes.OP_ADD, opcodes.IS_TMP_VAR, currentVal, opcodes.IS_CONST, oneConstant, opcodes.IS_TMP_VAR, newVal)
	} else { // "--"
		c.emit(opcodes.OP_SUB, opcodes.IS_TMP_VAR, currentVal, opcodes.IS_CONST, oneConstant, opcodes.IS_TMP_VAR, newVal)
	}
	
	// Step 6: Write new value back to static property
	c.emit(opcodes.OP_FETCH_STATIC_PROP_W,
		classOperandType, classOperand,
		propOperandType, propOperand,
		opcodes.IS_TMP_VAR, newVal)
	
	return nil
}

func (c *Compiler) compileStaticAccessIncDec(expr *ast.UnaryExpression, staticAccess *ast.StaticAccessExpression) error {
	// For static access increment/decrement, we'll create a simplified implementation
	// This handles cases like self::$counter++
	
	// For now, just create a temp result to avoid errors
	result := c.allocateTemp()
	constant := c.addConstant(values.NewInt(1))
	c.emit(opcodes.OP_QM_ASSIGN, opcodes.IS_CONST, constant, 0, 0, opcodes.IS_TMP_VAR, result)
	
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

	// Handle different left-hand side types
	if variable, ok := expr.Left.(*ast.Variable); ok {
		varSlot := c.getVariableSlot(variable.Name)

		// Emit variable name binding for variable variables support
		nameConstant := c.addConstant(values.NewString(variable.Name))
		c.emit(opcodes.OP_BIND_VAR_NAME, opcodes.IS_VAR, varSlot, opcodes.IS_CONST, nameConstant, 0, 0)

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
	} else if arrayAccess, ok := expr.Left.(*ast.ArrayAccessExpression); ok {
		// Handle array assignment: $arr[index] = value or $arr[] = value
		
		// Compile array variable
		if arrayVar, ok := arrayAccess.Array.(*ast.Variable); ok {
			arraySlot := c.getVariableSlot(arrayVar.Name)
			
			if arrayAccess.Index == nil {
				// Array append: $arr[] = value
				// Use ADD_ARRAY_ELEMENT instruction
				c.emit(opcodes.OP_ADD_ARRAY_ELEMENT, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, valueResult, opcodes.IS_VAR, arraySlot)
			} else {
				// Array index assignment: $arr[index] = value
				// This is more complex - need to implement proper array index assignment
				return fmt.Errorf("array index assignment not yet implemented")
			}
		} else {
			return fmt.Errorf("complex array expressions not yet supported")
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
	// Use the same allocation system as getOrCreateVariable for consistency
	return c.getOrCreateVariable(name)
}

func (c *Compiler) compileVariable(expr *ast.Variable) error {
	// Check if this is a variable variable disguised as a regular variable
	if len(expr.Name) > 3 && expr.Name[0] == '$' && expr.Name[1] == '{' && expr.Name[len(expr.Name)-1] == '}' {
		// This is ${...} syntax - extract the inner expression
		innerExpr := expr.Name[2 : len(expr.Name)-1] // Remove ${ and }
		
		// For now, handle simple cases like ${$varName}
		if len(innerExpr) > 1 && innerExpr[0] == '$' {
			// This is ${$varName} - compile as variable variable
			varName := innerExpr // Keep the full $varName
			
			// Create a temporary variable for the variable name
			nameSlot := c.getVariableSlot(varName)
			nameResult := c.allocateTemp()
			
			// Emit binding for the name variable
			nameConstant := c.addConstant(values.NewString(varName))
			c.emit(opcodes.OP_BIND_VAR_NAME, opcodes.IS_VAR, nameSlot, opcodes.IS_CONST, nameConstant, 0, 0)
			
			// Fetch the name variable value
			c.emit(opcodes.OP_FETCH_R, opcodes.IS_VAR, nameSlot, 0, 0, opcodes.IS_TMP_VAR, nameResult)
			
			// Use dynamic fetch with the name
			result := c.allocateTemp()
			c.emit(opcodes.OP_FETCH_R_DYNAMIC,
				opcodes.IS_TMP_VAR, nameResult,
				0, 0,
				opcodes.IS_TMP_VAR, result)
			
			return nil
		}
	}
	
	// Regular variable handling
	varSlot := c.getVariableSlot(expr.Name)
	
	// Emit variable name binding for variable variables support
	nameConstant := c.addConstant(values.NewString(expr.Name))
	c.emit(opcodes.OP_BIND_VAR_NAME, opcodes.IS_VAR, varSlot, opcodes.IS_CONST, nameConstant, 0, 0)
	
	result := c.allocateTemp()
	c.emit(opcodes.OP_FETCH_R, opcodes.IS_VAR, varSlot, 0, 0, opcodes.IS_TMP_VAR, result)
	return nil
}

func (c *Compiler) compileVariableVariable(expr *ast.VariableVariableExpression) error {
	// Variable variables: ${expression}
	// 1. Evaluate the inner expression to get the variable name
	// 2. Convert result to string if needed  
	// 3. Use that string as variable name for lookup
	
	// Compile the inner expression that will give us the variable name
	err := c.compileNode(expr.Expression)
	if err != nil {
		return fmt.Errorf("failed to compile variable variable expression: %w", err)
	}
	
	// The result of the expression is in the last allocated temp
	nameOperand := c.nextTemp - 1
	result := c.allocateTemp()
	
	// Use OP_FETCH_R_DYNAMIC to fetch variable by computed name
	// The VM will need to convert nameOperand to string and use it for variable lookup
	c.emit(opcodes.OP_FETCH_R_DYNAMIC,
		opcodes.IS_TMP_VAR, nameOperand, // Variable name (from expression)
		0, 0, // Unused operands
		opcodes.IS_TMP_VAR, result) // Result
	
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

func (c *Compiler) compileInterpolatedString(expr *ast.InterpolatedStringExpression) error {
	if len(expr.Parts) == 0 {
		// Empty interpolated string - return empty string constant
		constant := c.addConstant(values.NewString(""))
		result := c.allocateTemp()
		c.emit(opcodes.OP_QM_ASSIGN, opcodes.IS_CONST, constant, 0, 0, opcodes.IS_TMP_VAR, result)
		return nil
	}

	if len(expr.Parts) == 1 {
		// Single part - just compile it and convert to string if needed
		err := c.compileNode(expr.Parts[0])
		if err != nil {
			return err
		}
		// The result is already in the next temp variable
		return nil
	}

	// Multiple parts - compile first part
	err := c.compileNode(expr.Parts[0])
	if err != nil {
		return err
	}
	resultTemp := c.nextTemp - 1

	// Compile and concatenate remaining parts
	for i := 1; i < len(expr.Parts); i++ {
		err := c.compileNode(expr.Parts[i])
		if err != nil {
			return err
		}
		partTemp := c.nextTemp - 1

		// Concatenate with previous result
		newResult := c.allocateTemp()
		c.emit(opcodes.OP_CONCAT, opcodes.IS_TMP_VAR, resultTemp, opcodes.IS_TMP_VAR, partTemp, opcodes.IS_TMP_VAR, newResult)
		resultTemp = newResult
	}

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
	c.emit(opcodes.OP_INIT_ARRAY, opcodes.IS_UNUSED, 0, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, result)

	for _, element := range expr.Elements {
		if arrayElement, ok := element.(*ast.ArrayElementExpression); ok {
			// Wrapped array element
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

				c.emit(opcodes.OP_ADD_ARRAY_ELEMENT, opcodes.IS_TMP_VAR, keyResult, opcodes.IS_TMP_VAR, valueResult, opcodes.IS_TMP_VAR, result)
			} else {
				// Auto-indexed element
				err := c.compileNode(arrayElement.Value)
				if err != nil {
					return err
				}
				valueResult := c.allocateTemp()
				c.emitMove(valueResult)

				c.emit(opcodes.OP_ADD_ARRAY_ELEMENT, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, valueResult, opcodes.IS_TMP_VAR, result)
			}
		} else {
			// Direct element (not wrapped in ArrayElementExpression) - treat as auto-indexed
			err := c.compileNode(element)
			if err != nil {
				return err
			}
			valueResult := c.allocateTemp()
			c.emitMove(valueResult)

			c.emit(opcodes.OP_ADD_ARRAY_ELEMENT, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, valueResult, opcodes.IS_TMP_VAR, result)
		}
	}

	// Ensure the array result is in the expected location (c.nextTemp - 1)
	// The array was created in the 'result' temp variable at the beginning
	finalResult := c.allocateTemp()
	c.emit(opcodes.OP_QM_ASSIGN, opcodes.IS_TMP_VAR, result, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, finalResult)

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
	c.emit(opcodes.OP_FETCH_DIM_R, opcodes.IS_TMP_VAR, arrayResult, opcodes.IS_TMP_VAR, indexResult, opcodes.IS_TMP_VAR, result)

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
	c.emit(opcodes.OP_DO_FCALL, opcodes.IS_UNUSED, 0, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, result)

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
	c.emit(opcodes.OP_DO_FCALL, opcodes.IS_UNUSED, 0, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, result)

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
	// Move the result from the previous compilation to the target temp variable
	// We need to get the source temp before allocating target, so this must be called correctly
	source := c.nextTemp - 2  // -1 for the target that was just allocated, -1 more for the actual source
	c.emit(opcodes.OP_QM_ASSIGN, opcodes.IS_TMP_VAR, source, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, target)
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
	// Check if label already exists (backward jump)
	if pos, exists := c.labels[labelName]; exists {
		// Backward jump - emit instruction with known target
		jumpConstant := c.addConstant(values.NewInt(int64(pos)))
		if opcode == opcodes.OP_JMP {
			c.emit(opcode, opcodes.IS_CONST, jumpConstant, 0, 0, 0, 0)
		} else {
			c.emit(opcode, op1Type, op1, opcodes.IS_CONST, jumpConstant, 0, 0)
		}
		return
	}
	
	// Forward jump - add placeholder constant for the jump target
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
	// Check if label already exists (backward jump)
	if pos, exists := c.labels[labelName]; exists {
		// Backward jump - emit instruction with known target
		jumpConstant := c.addConstant(values.NewInt(int64(pos)))
		c.emit(opcodes.OP_JMPZ, condType, cond, opcodes.IS_CONST, jumpConstant, 0, 0)
		return
	}
	
	// Forward jump - add placeholder constant for the jump target
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
	// Check if label already exists (backward jump)
	if pos, exists := c.labels[labelName]; exists {
		// Backward jump - emit instruction with known target
		jumpConstant := c.addConstant(values.NewInt(int64(pos)))
		c.emit(opcodes.OP_JMPNZ, condType, cond, opcodes.IS_CONST, jumpConstant, 0, 0)
		return
	}
	
	// Forward jump - add placeholder constant for the jump target
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
	// Compile the subject expression
	err := c.compileNode(expr.Subject)
	if err != nil {
		return err
	}
	subjectTemp := c.nextTemp - 1

	// Allocate temp variable for the match result
	resultTemp := c.allocateTemp()

	// Create labels for each arm and the end
	endLabel := c.generateLabel()
	var defaultLabel string
	
	// Match expression evaluation: compare subject with each arm's conditions
	for _, arm := range expr.Arms {
		if !arm.IsDefault {
			// For each condition in this arm (comma-separated conditions)
			conditionMatchLabel := c.generateLabel()
			nextArmLabel := c.generateLabel()
			
			for j, condition := range arm.Conditions {
				// Compile the condition
				err := c.compileNode(condition)
				if err != nil {
					return err
				}
				conditionTemp := c.nextTemp - 1
				
				// Compare subject === condition using strict comparison
				c.allocateTemp() // For comparison result
				compResultTemp := c.nextTemp - 1
				c.emit(opcodes.OP_IS_IDENTICAL, opcodes.IS_TMP_VAR, subjectTemp,
					   opcodes.IS_TMP_VAR, conditionTemp, opcodes.IS_TMP_VAR, compResultTemp)
				
				// If this condition matches, jump to arm body
				c.emitJumpNZ(opcodes.IS_TMP_VAR, compResultTemp, conditionMatchLabel)
				
				// If this was the last condition and none matched, try next arm
				if j == len(arm.Conditions)-1 {
					c.emitJump(opcodes.OP_JMP, opcodes.IS_CONST, 0, nextArmLabel)
				}
			}
			
			// Label for when this arm's condition matches
			c.placeLabel(conditionMatchLabel)
			
			// Compile the arm body
			err := c.compileNode(arm.Body)
			if err != nil {
				return err
			}
			bodyResultTemp := c.nextTemp - 1
			
			// Store the result in our match result temp
			c.emit(opcodes.OP_QM_ASSIGN, opcodes.IS_TMP_VAR, bodyResultTemp, 0, 0, opcodes.IS_TMP_VAR, resultTemp)
			
			// After executing the arm, jump to end (no fall-through in match)
			c.emitJump(opcodes.OP_JMP, opcodes.IS_CONST, 0, endLabel)
			
			// Label for trying the next arm
			c.placeLabel(nextArmLabel)
		} else {
			// This is a default arm, remember its label for later
			defaultLabel = c.generateLabel()
		}
	}
	
	// If no condition matched, execute default arm if present
	if defaultLabel != "" {
		c.emitJump(opcodes.OP_JMP, opcodes.IS_CONST, 0, defaultLabel)
		
		// Emit default arm
		c.placeLabel(defaultLabel)
		
		// Find and compile the default arm
		for _, arm := range expr.Arms {
			if arm.IsDefault {
				err := c.compileNode(arm.Body)
				if err != nil {
					return err
				}
				bodyResultTemp := c.nextTemp - 1
				
				// Store the result in our match result temp
				c.emit(opcodes.OP_QM_ASSIGN, opcodes.IS_TMP_VAR, bodyResultTemp, 0, 0, opcodes.IS_TMP_VAR, resultTemp)
				break
			}
		}
	} else {
		// No default arm - throw UnhandledMatchError
		// Create an UnhandledMatchError object with message property
		errorObj := values.NewObject("UnhandledMatchError")
		errorObj.Data.(*values.Object).Properties["message"] = values.NewString("UnhandledMatchError")
		c.emit(opcodes.OP_THROW, opcodes.IS_CONST, c.addConstant(errorObj), 0, 0, 0, 0)
	}
	
	c.placeLabel(endLabel)
	
	// Ensure the final result is in the expected position (nextTemp - 1)
	// This is needed so that parent expressions (like echo) can find the result
	if resultTemp != c.nextTemp - 1 {
		finalResult := c.allocateTemp()
		c.emit(opcodes.OP_QM_ASSIGN, opcodes.IS_TMP_VAR, resultTemp, 0, 0, opcodes.IS_TMP_VAR, finalResult)
	}
	
	return nil
}

func (c *Compiler) compileNew(expr *ast.NewExpression) error {
	// Get the class name - for now we only support simple class names
	var className string
	switch class := expr.Class.(type) {
	case *ast.CallExpression:
		// new Exception("message") - CallExpression with constructor arguments
		if callee, ok := class.Callee.(*ast.IdentifierNode); ok {
			className = callee.Name
		} else {
			return fmt.Errorf("unsupported class expression in new")
		}
		
		// Create object first
		classConstant := c.addConstant(values.NewString(className))
		result := c.allocateTemp()
		c.emit(opcodes.OP_NEW, opcodes.IS_CONST, classConstant, 0, 0, opcodes.IS_TMP_VAR, result)
		
		// For simplicity, we'll ignore constructor arguments for now
		// In a full implementation, we'd compile the arguments and call the constructor
		
		return nil
		
	case *ast.IdentifierNode:
		// new Exception - simple class instantiation
		className = class.Name
	default:
		return fmt.Errorf("unsupported class expression in new: %T", expr.Class)
	}
	
	// Create the object
	classConstant := c.addConstant(values.NewString(className))
	result := c.allocateTemp()
	c.emit(opcodes.OP_NEW, opcodes.IS_CONST, classConstant, 0, 0, opcodes.IS_TMP_VAR, result)
	
	return nil
}

func (c *Compiler) compileFor(stmt *ast.ForStatement) error {
	// Create labels for the loop
	testLabel := c.generateLabel()
	bodyLabel := c.generateLabel()
	updateLabel := c.generateLabel()
	endLabel := c.generateLabel()

	// Set break/continue labels for this scope
	oldBreak := c.currentScope().breakLabel
	oldContinue := c.currentScope().continueLabel
	c.currentScope().breakLabel = endLabel
	c.currentScope().continueLabel = updateLabel

	// Compile initialization (if exists)
	if stmt.Init != nil {
		err := c.compileNode(stmt.Init)
		if err != nil {
			return err
		}
	}

	// Test label - evaluate condition first
	c.placeLabel(testLabel)

	// Compile test condition (if exists)
	if stmt.Test != nil {
		err := c.compileNode(stmt.Test)
		if err != nil {
			return err
		}
		// Jump to end if condition is false
		condResult := c.allocateTemp()
		c.emitMove(condResult)
		c.emitJumpZ(opcodes.IS_TMP_VAR, condResult, endLabel)
	}

	// Body label - start of loop body
	c.placeLabel(bodyLabel)

	// Compile body
	for _, s := range stmt.Body {
		err := c.compileNode(s)
		if err != nil {
			return err
		}
	}

	// Update label (where continue jumps to)
	c.placeLabel(updateLabel)

	// Compile update expression (if exists)
	if stmt.Update != nil {
		err := c.compileNode(stmt.Update)
		if err != nil {
			return err
		}
	}

	// Jump back to test condition
	c.emitJump(opcodes.OP_JMP, opcodes.IS_CONST, 0, testLabel)

	// End label (where break jumps to)
	c.placeLabel(endLabel)

	// Restore old break/continue labels
	c.currentScope().breakLabel = oldBreak
	c.currentScope().continueLabel = oldContinue

	return nil
}

func (c *Compiler) compileForeach(stmt *ast.ForeachStatement) error {
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

	// Compile body
	err = c.compileNode(stmt.Body)
	if err != nil {
		return err
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
	// Generate labels for control flow
	catchLabels := make([]string, len(stmt.CatchClauses))
	for i := range stmt.CatchClauses {
		catchLabels[i] = c.generateLabel()
	}
	finallyLabel := c.generateLabel()
	endLabel := c.generateLabel()

	// Start of try block - emit exception handler setup
	// For now, we'll use a simple approach: register first catch block
	var firstCatchLabel string
	if len(catchLabels) > 0 {
		firstCatchLabel = catchLabels[0]
	}
	
	// This is a simplified exception handler registration
	// Real implementation would encode all handler info in instruction
	c.emit(opcodes.OP_CATCH, opcodes.IS_CONST, 0, 0, 0, 0, 0)

	// Compile try block body
	for _, s := range stmt.Body {
		err := c.compileNode(s)
		if err != nil {
			return err
		}
	}

	// If no exception occurred, jump to finally block (or end if no finally)
	if len(stmt.FinallyBlock) > 0 {
		c.emitJump(opcodes.OP_JMP, opcodes.IS_CONST, 0, finallyLabel)
	} else {
		c.emitJump(opcodes.OP_JMP, opcodes.IS_CONST, 0, endLabel)
	}

	// Compile catch blocks
	for i, catchClause := range stmt.CatchClauses {
		c.placeLabel(catchLabels[i])
		
		// Get exception variable slot if specified
		if catchClause.Parameter != nil {
			paramVar, ok := catchClause.Parameter.(*ast.Variable)
			if ok {
				// Store variable slot for exception parameter (used by VM during exception handling)
				_ = c.getVariableSlot(paramVar.Name)
			}
		}

		// For now, we catch all exceptions (type matching not fully implemented)
		// In a full implementation, we'd emit type checking code here
		
		// Compile catch block body
		for _, s := range catchClause.Body {
			err := c.compileNode(s)
			if err != nil {
				return err
			}
		}

		// Jump to finally block (or end if no finally)
		if len(stmt.FinallyBlock) > 0 {
			c.emitJump(opcodes.OP_JMP, opcodes.IS_CONST, 0, finallyLabel)
		} else {
			c.emitJump(opcodes.OP_JMP, opcodes.IS_CONST, 0, endLabel)
		}
	}

	// Compile finally block if present
	if len(stmt.FinallyBlock) > 0 {
		c.placeLabel(finallyLabel)
		c.emit(opcodes.OP_FINALLY, 0, 0, 0, 0, 0, 0)
		
		for _, s := range stmt.FinallyBlock {
			err := c.compileNode(s)
			if err != nil {
				return err
			}
		}
	}

	// End label
	c.placeLabel(endLabel)
	
	// Now we need to patch the OP_CATCH instruction with the actual catch block address
	// This is a post-processing step after labels are resolved
	c.patchExceptionHandler(firstCatchLabel, finallyLabel)

	return nil
}

// patchExceptionHandler updates the most recent OP_CATCH instruction with handler addresses
func (c *Compiler) patchExceptionHandler(catchLabel, finallyLabel string) {
	// Find the most recent OP_CATCH instruction and update it with actual addresses
	for i := len(c.instructions) - 1; i >= 0; i-- {
		if c.instructions[i].Opcode == opcodes.OP_CATCH {
			// Encode catch and finally addresses in the instruction
			catchAddr := 0
			finallyAddr := 0
			
			if catchLabel != "" {
				if addr, exists := c.labels[catchLabel]; exists {
					catchAddr = addr
				}
			}
			
			if finallyLabel != "" {
				if addr, exists := c.labels[finallyLabel]; exists {
					finallyAddr = addr
				}
			}
			
			// Update the instruction with the addresses
			c.instructions[i].Op1 = uint32(catchAddr)
			c.instructions[i].Op2 = uint32(finallyAddr)
			break
		}
	}
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
	// Cast Identifier interface to concrete type
	nameNode, ok := decl.Name.(*ast.IdentifierNode)
	if !ok {
		return fmt.Errorf("invalid function name type")
	}
	funcName := nameNode.Name
	
	// Check if function already exists
	if _, exists := c.functions[funcName]; exists {
		return fmt.Errorf("function %s already declared", funcName)
	}
	
	// Create new function
	function := &vm.Function{
		Name:         funcName,
		Instructions: make([]opcodes.Instruction, 0),
		Constants:    make([]*values.Value, 0),
		Parameters:   make([]vm.Parameter, 0),
		IsVariadic:   false,
		IsGenerator:  false,
	}
	
	// Compile parameters
	if decl.Parameters != nil {
		for _, param := range decl.Parameters.Parameters {
			// param is *ParameterNode
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
				// For now, we'll compile the default value later
				// This requires more complex evaluation
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
	oldInstructions := c.instructions
	oldConstants := c.constants
	
	// Reset for function compilation
	c.instructions = make([]opcodes.Instruction, 0)
	c.constants = make([]*values.Value, 0)
	
	// Create function scope
	c.pushScope(true)
	
	// Set up parameter variables in the function scope
	if decl.Parameters != nil {
		for _, param := range decl.Parameters.Parameters {
			if nameNode, ok := param.Name.(*ast.IdentifierNode); ok {
				// Register parameter name in function scope
				c.getOrCreateVariable(nameNode.Name)
			}
		}
	}
	
	// Compile function body
	for _, stmt := range decl.Body {
		err := c.compileNode(stmt)
		if err != nil {
			c.popScope()
			c.instructions = oldInstructions
			c.constants = oldConstants
			return fmt.Errorf("error compiling function %s: %v", funcName, err)
		}
	}
	
	// Add implicit return if needed
	if len(c.instructions) == 0 || c.instructions[len(c.instructions)-1].Opcode != opcodes.OP_RETURN {
		c.emit(opcodes.OP_RETURN, opcodes.IS_CONST, c.addConstant(values.NewNull()), 0, 0, 0, 0)
	}
	
	// Store compiled function
	function.Instructions = c.instructions
	function.Constants = c.constants
	c.functions[funcName] = function
	
	// Restore compiler state
	c.popScope()
	c.instructions = oldInstructions
	c.constants = oldConstants
	
	// Emit function declaration instruction
	nameConstant := c.addConstant(values.NewString(funcName))
	c.emit(opcodes.OP_DECLARE_FUNCTION, opcodes.IS_CONST, nameConstant, 0, 0, 0, 0)
	
	return nil
}

func (c *Compiler) compileClassDeclaration(decl *ast.AnonymousClass) error {
	// Generate unique class name for anonymous class
	className := fmt.Sprintf("class@anonymous_%d", len(c.classes))
	
	// Check if class already exists (shouldn't happen for anonymous classes)
	if _, exists := c.classes[className]; exists {
		return fmt.Errorf("class %s already declared", className)
	}
	
	// Create new class
	class := &vm.Class{
		Name:        className,
		ParentClass: "",
		Properties:  make(map[string]*vm.Property),
		Methods:     make(map[string]*vm.Function),
		Constants:   make(map[string]*values.Value),
		IsAbstract:  false,
		IsFinal:     false,
	}
	
	// Handle modifiers
	for _, modifier := range decl.Modifiers {
		switch modifier {
		case "abstract":
			class.IsAbstract = true
		case "final":
			class.IsFinal = true
		}
	}
	
	// Handle extends
	if decl.Extends != nil {
		if parent, ok := decl.Extends.(*ast.IdentifierNode); ok {
			class.ParentClass = parent.Name
		} else {
			return fmt.Errorf("complex parent class expressions not supported yet")
		}
	}
	
	// Store current class context
	oldCurrentClass := c.currentClass
	c.currentClass = class
	
	// Emit class table initialization
	nameConstant := c.addConstant(values.NewString(className))
	c.emit(opcodes.OP_INIT_CLASS_TABLE, opcodes.IS_CONST, nameConstant, 0, 0, 0, 0)
	
	// Set parent class if exists
	if class.ParentClass != "" {
		parentConstant := c.addConstant(values.NewString(class.ParentClass))
		c.emit(opcodes.OP_SET_CLASS_PARENT, opcodes.IS_CONST, nameConstant, opcodes.IS_CONST, parentConstant, 0, 0)
	}
	
	// Handle implements
	for _, iface := range decl.Implements {
		if ifaceId, ok := iface.(*ast.IdentifierNode); ok {
			ifaceConstant := c.addConstant(values.NewString(ifaceId.Name))
			c.emit(opcodes.OP_ADD_INTERFACE, opcodes.IS_CONST, nameConstant, opcodes.IS_CONST, ifaceConstant, 0, 0)
		} else {
			return fmt.Errorf("complex interface expressions not supported yet")
		}
	}
	
	// Compile class body
	for _, stmt := range decl.Body {
		err := c.compileNode(stmt)
		if err != nil {
			c.currentClass = oldCurrentClass
			return fmt.Errorf("error compiling anonymous class: %v", err)
		}
	}
	
	// Store compiled class
	c.classes[className] = class
	c.currentClass = oldCurrentClass
	
	// Emit class declaration instruction
	c.emit(opcodes.OP_DECLARE_CLASS, opcodes.IS_CONST, nameConstant, 0, 0, 0, 0)
	
	// Handle constructor arguments if provided
	if decl.Arguments != nil {
		// Compile constructor call arguments
		for _, arg := range decl.Arguments.Arguments {
			err := c.compileNode(arg)
			if err != nil {
				return fmt.Errorf("error compiling constructor argument: %v", err)
			}
		}
		
		// Create new instance with constructor call
		result := c.allocateTemp()
		c.emit(opcodes.OP_NEW, opcodes.IS_CONST, nameConstant, 0, 0, opcodes.IS_TMP_VAR, result)
		
		// If there are arguments, we need to call constructor
		if len(decl.Arguments.Arguments) > 0 {
			// This would require more complex constructor calling logic
			// For now, we'll just create the object without calling constructor
		}
	} else {
		// Simple instantiation without constructor arguments
		result := c.allocateTemp()
		c.emit(opcodes.OP_NEW, opcodes.IS_CONST, nameConstant, 0, 0, opcodes.IS_TMP_VAR, result)
	}
	
	return nil
}

func (c *Compiler) compilePropertyDeclaration(decl *ast.PropertyDeclaration) error {
	// Check if we're in a class context
	if c.currentClass == nil {
		return fmt.Errorf("property declaration outside of class context")
	}
	
	propName := decl.Name
	
	// Check if property already exists
	if _, exists := c.currentClass.Properties[propName]; exists {
		return fmt.Errorf("property $%s already declared in class %s", propName, c.currentClass.Name)
	}
	
	// Create new property
	property := &vm.Property{
		Name:       propName,
		Visibility: decl.Visibility, // public, private, protected
		IsStatic:   decl.Static,
	}
	
	// Handle type hint
	if decl.Type != nil {
		property.Type = decl.Type.String()
	}
	
	// Handle default value
	var defaultValue *values.Value
	if decl.DefaultValue != nil {
		// For simple literals, we can evaluate them directly
		switch defVal := decl.DefaultValue.(type) {
		case *ast.NumberLiteral:
			if defVal.Kind == "int" {
				value, err := strconv.ParseInt(defVal.Value, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid integer literal for property default: %s", defVal.Value)
				}
				defaultValue = values.NewInt(value)
			} else if defVal.Kind == "float" {
				value, err := strconv.ParseFloat(defVal.Value, 64)
				if err != nil {
					return fmt.Errorf("invalid float literal for property default: %s", defVal.Value)
				}
				defaultValue = values.NewFloat(value)
			}
		case *ast.StringLiteral:
			defaultValue = values.NewString(defVal.Value)
		case *ast.BooleanLiteral:
			defaultValue = values.NewBool(defVal.Value)
		case *ast.NullLiteral:
			defaultValue = values.NewNull()
		case *ast.ArrayExpression:
			// For now, empty arrays only
			if len(defVal.Elements) == 0 {
				defaultValue = values.NewArray()
			} else {
				return fmt.Errorf("complex array default values not supported yet for property %s", propName)
			}
		case *ast.Variable:
			// Variables in default values (like static properties) are not allowed in most contexts
			// But we can handle some special cases
			return fmt.Errorf("variable expressions not supported in property defaults for property %s", propName)
		default:
			// Handle other types by setting to null
			defaultValue = values.NewNull()
		}
	}
	
	property.DefaultValue = defaultValue
	
	// Add property to current class
	c.currentClass.Properties[propName] = property
	
	// Emit property declaration instruction
	classNameConstant := c.addConstant(values.NewString(c.currentClass.Name))
	propNameConstant := c.addConstant(values.NewString(propName))
	visibilityConstant := c.addConstant(values.NewString(decl.Visibility))
	
	// Emit property declaration with metadata
	c.emit(opcodes.OP_DECLARE_PROPERTY, 
		opcodes.IS_CONST, classNameConstant,
		opcodes.IS_CONST, propNameConstant,
		opcodes.IS_CONST, visibilityConstant)
	
	// If there's a default value, emit it
	if defaultValue != nil {
		defaultConstant := c.addConstant(defaultValue)
		// Additional instruction to set default value
		c.emit(opcodes.OP_ASSIGN, 
			opcodes.IS_CONST, defaultConstant,
			0, 0,
			opcodes.IS_CONST, propNameConstant)
	}
	
	return nil
}

func (c *Compiler) compileClassConstant(decl *ast.ClassConstantDeclaration) error {
	// Check if we're in a class context
	if c.currentClass == nil {
		return fmt.Errorf("class constant declaration outside of class context")
	}
	
	// Process each constant in the declaration
	for _, constDeclarator := range decl.Constants {
		// Cast Expression to concrete type to get constant name
		nameExpr, ok := constDeclarator.Name.(*ast.IdentifierNode)
		if !ok {
			return fmt.Errorf("invalid constant name type")
		}
		constName := nameExpr.Name
		
		// Check if constant already exists
		if _, exists := c.currentClass.Constants[constName]; exists {
			return fmt.Errorf("constant %s already declared in class %s", constName, c.currentClass.Name)
		}
		
		// Evaluate constant value
		var constValue *values.Value
		switch val := constDeclarator.Value.(type) {
		case *ast.NumberLiteral:
			if val.Kind == "int" {
				intValue, err := strconv.ParseInt(val.Value, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid integer literal for constant %s: %s", constName, val.Value)
				}
				constValue = values.NewInt(intValue)
			} else if val.Kind == "float" {
				floatValue, err := strconv.ParseFloat(val.Value, 64)
				if err != nil {
					return fmt.Errorf("invalid float literal for constant %s: %s", constName, val.Value)
				}
				constValue = values.NewFloat(floatValue)
			} else {
				// Handle other numeric types
				if intValue, err := strconv.ParseInt(val.Value, 10, 64); err == nil {
					constValue = values.NewInt(intValue)
				} else if floatValue, err := strconv.ParseFloat(val.Value, 64); err == nil {
					constValue = values.NewFloat(floatValue)
				} else {
					return fmt.Errorf("invalid number literal for constant %s: %s", constName, val.Value)
				}
			}
		case *ast.StringLiteral:
			constValue = values.NewString(val.Value)
		case *ast.BooleanLiteral:
			constValue = values.NewBool(val.Value)
		case *ast.NullLiteral:
			constValue = values.NewNull()
		case *ast.ArrayExpression:
			// For now, only empty arrays
			if len(val.Elements) == 0 {
				constValue = values.NewArray()
			} else {
				return fmt.Errorf("complex array constants not supported yet for constant %s", constName)
			}
		case *ast.IdentifierNode:
			// Handle simple constant references like true, false, etc.
			switch val.Name {
			case "true":
				constValue = values.NewBool(true)
			case "false":
				constValue = values.NewBool(false)
			case "null":
				constValue = values.NewNull()
			default:
				return fmt.Errorf("constant references not supported yet for constant %s (identifier: %s)", constName, val.Name)
			}
		default:
			// For complex expressions, we'd need more sophisticated constant evaluation
			return fmt.Errorf("complex constant expressions not supported yet for constant %s", constName)
		}
		
		if constValue == nil {
			return fmt.Errorf("could not evaluate constant value for %s", constName)
		}
		
		// Add constant to current class
		c.currentClass.Constants[constName] = constValue
		
		// Emit class constant declaration instruction
		classNameConstant := c.addConstant(values.NewString(c.currentClass.Name))
		constNameConstant := c.addConstant(values.NewString(constName))
		constValueConstant := c.addConstant(constValue)
		
		// Emit the class constant declaration with all metadata
		c.emit(opcodes.OP_DECLARE_CLASS_CONST,
			opcodes.IS_CONST, classNameConstant,
			opcodes.IS_CONST, constNameConstant,
			opcodes.IS_CONST, constValueConstant)
		
		// Emit visibility and other flags if needed
		if decl.IsFinal {
			// We could add a separate opcode for final constants, but for now we'll note it
			// In a more complete implementation, this would be handled by the VM
		}
		
		if decl.IsAbstract {
			// Abstract constants are a PHP 8.0+ feature and would need special handling
			// For now, we'll note this but not implement the full logic
		}
		
		// Handle typed constants (PHP 8.3+)
		if decl.Type != nil {
			// Type information could be stored and validated at runtime
			// For now, we'll store it in the constant metadata
			// We could emit additional metadata about the type, but for simplicity we'll skip this
		}
	}
	
	return nil
}

func (c *Compiler) compileRegularClassDeclaration(decl *ast.ClassExpression) error {
	// Get class name
	var className string
	if nameNode, ok := decl.Name.(*ast.IdentifierNode); ok {
		className = nameNode.Name
	} else {
		return fmt.Errorf("invalid class name type")
	}
	
	// Check if class already exists
	if _, exists := c.classes[className]; exists {
		return fmt.Errorf("class %s already declared", className)
	}
	
	// Create new class
	class := &vm.Class{
		Name:        className,
		ParentClass: "",
		Properties:  make(map[string]*vm.Property),
		Methods:     make(map[string]*vm.Function),
		Constants:   make(map[string]*values.Value),
		IsAbstract:  decl.Abstract,
		IsFinal:     decl.Final,
	}
	
	// Handle extends
	if decl.Extends != nil {
		if parent, ok := decl.Extends.(*ast.IdentifierNode); ok {
			class.ParentClass = parent.Name
		} else {
			return fmt.Errorf("complex parent class expressions not supported yet")
		}
	}
	
	// Store current class context
	oldCurrentClass := c.currentClass
	c.currentClass = class
	
	// Emit class table initialization
	nameConstant := c.addConstant(values.NewString(className))
	c.emit(opcodes.OP_INIT_CLASS_TABLE, opcodes.IS_CONST, nameConstant, 0, 0, 0, 0)
	
	// Set parent class if exists
	if class.ParentClass != "" {
		parentConstant := c.addConstant(values.NewString(class.ParentClass))
		c.emit(opcodes.OP_SET_CLASS_PARENT, opcodes.IS_CONST, nameConstant, opcodes.IS_CONST, parentConstant, 0, 0)
	}
	
	// Handle implements
	for _, iface := range decl.Implements {
		if ifaceId, ok := iface.(*ast.IdentifierNode); ok {
			ifaceConstant := c.addConstant(values.NewString(ifaceId.Name))
			c.emit(opcodes.OP_ADD_INTERFACE, opcodes.IS_CONST, nameConstant, opcodes.IS_CONST, ifaceConstant, 0, 0)
		} else {
			return fmt.Errorf("complex interface expressions not supported yet")
		}
	}
	
	// Compile class body
	for _, stmt := range decl.Body {
		err := c.compileNode(stmt)
		if err != nil {
			c.currentClass = oldCurrentClass
			return fmt.Errorf("error compiling class %s: %v", className, err)
		}
	}
	
	// Store compiled class
	c.classes[className] = class
	c.currentClass = oldCurrentClass
	
	// Emit class declaration instruction
	c.emit(opcodes.OP_DECLARE_CLASS, opcodes.IS_CONST, nameConstant, 0, 0, 0, 0)
	
	return nil
}

func (c *Compiler) compileClassConstantAccess(expr *ast.ClassConstantAccessExpression) error {
	// Handle class constant access like ClassName::CONSTANT_NAME
	
	// Get class name
	var className string
	switch class := expr.Class.(type) {
	case *ast.IdentifierNode:
		className = class.Name
	case *ast.Variable:
		// Handle self::, static::, parent:: etc
		className = class.Name // Will be "self", "static", "parent"
	default:
		return fmt.Errorf("unsupported class expression in constant access: %T", expr.Class)
	}
	
	// Get constant name
	var constantName string
	if constId, ok := expr.Constant.(*ast.IdentifierNode); ok {
		constantName = constId.Name
	} else {
		return fmt.Errorf("invalid constant name type")
	}
	
	// For now, we'll create a simplified implementation
	// In a full implementation, this would look up the constant value from the class
	result := c.allocateTemp()
	
	// Create constants for the class and constant names
	classConstant := c.addConstant(values.NewString(className))
	constConstant := c.addConstant(values.NewString(constantName))
	
	// Emit instruction to fetch class constant
	c.emit(opcodes.OP_FETCH_CLASS_CONSTANT, 
		opcodes.IS_CONST, classConstant,
		opcodes.IS_CONST, constConstant,
		opcodes.IS_TMP_VAR, result)
	
	return nil
}

func (c *Compiler) compileStaticPropertyAccess(expr *ast.StaticPropertyAccessExpression) error {
	// Handle static property access specifically for Class::$property
	// This is distinct from constants (Class::CONST) or method calls (Class::method())
	
	result := c.allocateTemp()
	
	// Compile class expression (supports both static and dynamic class names)
	var classOperandType opcodes.OpType
	var classOperand uint32
	
	switch class := expr.Class.(type) {
	case *ast.IdentifierNode:
		// Static class name like MyClass::$prop
		classOperand = c.addConstant(values.NewString(class.Name))
		classOperandType = opcodes.IS_CONST
	case *ast.Variable:
		// Handle self::$prop, static::$prop, parent::$prop etc
		classOperand = c.addConstant(values.NewString(class.Name))
		classOperandType = opcodes.IS_CONST
	default:
		// Dynamic class expression like $className::$prop
		err := c.compileNode(class)
		if err != nil {
			return fmt.Errorf("failed to compile class expression in static property access: %w", err)
		}
		classOperand = c.nextTemp - 1
		classOperandType = opcodes.IS_TMP_VAR
	}
	
	// Compile property expression 
	var propOperandType opcodes.OpType
	var propOperand uint32
	
	switch property := expr.Property.(type) {
	case *ast.Variable:
		// Simple static property like ::$prop
		// Strip the $ from the property name since class properties are stored without $
		propName := property.Name
		if len(propName) > 0 && propName[0] == '$' {
			propName = propName[1:]
		}
		propOperand = c.addConstant(values.NewString(propName))
		propOperandType = opcodes.IS_CONST
	default:
		// Dynamic property expression like ::${$expr} or ::${"prop"}
		err := c.compileNode(property)
		if err != nil {
			return fmt.Errorf("failed to compile property expression in static property access: %w", err)
		}
		propOperand = c.nextTemp - 1
		propOperandType = opcodes.IS_TMP_VAR
	}
	
	// Emit static property access instruction  
	c.emit(opcodes.OP_FETCH_STATIC_PROP_R,
		classOperandType, classOperand,
		propOperandType, propOperand,
		opcodes.IS_TMP_VAR, result)
	
	return nil
}

func (c *Compiler) compileStaticAccess(expr *ast.StaticAccessExpression) error {
	// Handle static access like Class::CONSTANT, self::method, or Class::$property
	
	result := c.allocateTemp()
	
	// Compile class expression (supports both static names and dynamic expressions)
	var classOperandType opcodes.OpType
	var classOperand uint32
	
	switch class := expr.Class.(type) {
	case *ast.IdentifierNode:
		// Static class name like MyClass::
		className := class.Name
		classOperand = c.addConstant(values.NewString(className))
		classOperandType = opcodes.IS_CONST
	case *ast.Variable:
		// Handle self::, static::, parent:: etc - these are also treated as constants
		className := class.Name
		classOperand = c.addConstant(values.NewString(className))
		classOperandType = opcodes.IS_CONST
	default:
		// Dynamic class expression like $className::
		err := c.compileNode(class)
		if err != nil {
			return fmt.Errorf("failed to compile class expression: %w", err)
		}
		classOperand = c.nextTemp - 1 // Last allocated temp contains the class name
		classOperandType = opcodes.IS_TMP_VAR
	}
	
	// Compile property expression and determine access type
	switch property := expr.Property.(type) {
	case *ast.IdentifierNode:
		// Constant access (Class::CONSTANT)
		propOperand := c.addConstant(values.NewString(property.Name))
		c.emit(opcodes.OP_FETCH_CLASS_CONSTANT,
			classOperandType, classOperand,
			opcodes.IS_CONST, propOperand,
			opcodes.IS_TMP_VAR, result)
	case *ast.Variable:
		// Static property access (Class::$property)
		propOperand := c.addConstant(values.NewString(property.Name))
		c.emit(opcodes.OP_FETCH_STATIC_PROP_R,
			classOperandType, classOperand,
			opcodes.IS_CONST, propOperand,
			opcodes.IS_TMP_VAR, result)
	default:
		// Dynamic property expression like Class::${$expr} or Class::${"prop"}
		err := c.compileNode(property)
		if err != nil {
			return fmt.Errorf("failed to compile property expression: %w", err)
		}
		propOperand := c.nextTemp - 1
		
		// For dynamic properties, we assume it's a property access (not constant)
		// This matches PHP's behavior where Class::${expr} is always property access
		c.emit(opcodes.OP_FETCH_STATIC_PROP_R,
			classOperandType, classOperand,
			opcodes.IS_TMP_VAR, propOperand,
			opcodes.IS_TMP_VAR, result)
	}
	
	return nil
}

