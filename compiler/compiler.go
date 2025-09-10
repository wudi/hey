package compiler

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/compiler/opcodes"
	"github.com/wudi/php-parser/compiler/registry"
	"github.com/wudi/php-parser/compiler/values"
	"github.com/wudi/php-parser/compiler/vm"
	"github.com/wudi/php-parser/lexer"
)

// ForwardJump represents a jump that needs to be resolved later
type ForwardJump struct {
	instructionIndex int
	opType           opcodes.OpType
	operand          int // 1 for Op1, 2 for Op2
}

// Compiler compiles AST to bytecode
type Compiler struct {
	instructions     []opcodes.Instruction
	constants        []*values.Value
	scopes           []*Scope
	labels           map[string]int
	forwardJumps     map[string][]ForwardJump
	nextTemp         uint32
	nextLabel        int
	nextAnonFunction int // Counter for anonymous functions
	functions        map[string]*vm.Function
	classes          map[string]*vm.Class
	interfaces       map[string]*vm.Interface
	traits           map[string]*vm.Trait
	currentClass     *vm.Class // Current class being compiled
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
	// Initialize the unified registry for compilation
	registry.Initialize()

	return &Compiler{
		instructions:     make([]opcodes.Instruction, 0),
		constants:        make([]*values.Value, 0),
		scopes:           make([]*Scope, 0),
		labels:           make(map[string]int),
		forwardJumps:     make(map[string][]ForwardJump),
		nextTemp:         1000, // Start temp vars at 1000 to avoid conflicts
		nextLabel:        0,
		nextAnonFunction: 0, // Start anonymous function counter at 0
		functions:        make(map[string]*vm.Function),
		classes:          make(map[string]*vm.Class),
		interfaces:       make(map[string]*vm.Interface),
		traits:           make(map[string]*vm.Trait),
		currentClass:     nil,
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
	case *ast.AnonymousFunctionExpression:
		return c.compileAnonymousFunction(n)
	case *ast.IncludeOrEvalExpression:
		return c.compileIncludeOrEval(n)
	case *ast.PrintExpression:
		return c.compilePrintExpression(n)
	case *ast.CloneExpression:
		return c.compileCloneExpression(n)
	case *ast.InstanceofExpression:
		return c.compileInstanceofExpression(n)
	case *ast.CastExpression:
		return c.compileCastExpression(n)
	case *ast.ErrorSuppressionExpression:
		return c.compileErrorSuppressionExpression(n)
	case *ast.EmptyExpression:
		return c.compileEmptyExpression(n)
	case *ast.ExitExpression:
		return c.compileExitExpression(n)
	case *ast.IssetExpression:
		return c.compileIssetExpression(n)
	case *ast.ListExpression:
		return c.compileListExpression(n)
	case *ast.EvalExpression:
		return c.compileEvalExpression(n)
	case *ast.YieldExpression:
		return c.compileYieldExpression(n)
	case *ast.YieldFromExpression:
		return c.compileYieldFromExpression(n)
	case *ast.ThrowExpression:
		return c.compileThrowExpression(n)
	case *ast.MagicConstantExpression:
		return c.compileMagicConstantExpression(n)
	case *ast.NamespaceNameExpression:
		return c.compileNamespaceNameExpression(n)
	case *ast.NullsafePropertyAccessExpression:
		return c.compileNullsafePropertyAccessExpression(n)
	case *ast.NullsafeMethodCallExpression:
		return c.compileNullsafeMethodCallExpression(n)
	case *ast.ShellExecExpression:
		return c.compileShellExecExpression(n)
	case *ast.CommaExpression:
		return c.compileCommaExpression(n)
	case *ast.SpreadExpression:
		return c.compileSpreadExpression(n)
	case *ast.ArrowFunctionExpression:
		return c.compileArrowFunctionExpression(n)
	case *ast.FirstClassCallable:
		return c.compileFirstClassCallable(n)
	case *ast.ArrayElementExpression:
		return c.compileArrayElement(n)

	// Statements
	case *ast.ExpressionStatement:
		return c.compileExpressionStatement(n)
	case *ast.EchoStatement:
		return c.compileEcho(n)
	case *ast.PrintStatement:
		return c.compilePrintStatement(n)
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
	case *ast.GlobalStatement:
		return c.compileGlobalStatement(n)
	case *ast.StaticStatement:
		return c.compileStaticStatement(n)
	case *ast.UnsetStatement:
		return c.compileUnsetStatement(n)
	case *ast.DoWhileStatement:
		return c.compileDoWhileStatement(n)
	case *ast.GotoStatement:
		return c.compileGotoStatement(n)
	case *ast.LabelStatement:
		return c.compileLabelStatement(n)
	case *ast.HaltCompilerStatement:
		return c.compileHaltCompilerStatement(n)
	case *ast.DeclareStatement:
		return c.compileDeclareStatement(n)
	case *ast.NamespaceStatement:
		return c.compileNamespaceStatement(n)
	case *ast.UseStatement:
		return c.compileUseStatement(n)
	case *ast.AlternativeIfStatement:
		return c.compileAlternativeIfStatement(n)
	case *ast.AlternativeWhileStatement:
		return c.compileAlternativeWhileStatement(n)
	case *ast.AlternativeForStatement:
		return c.compileAlternativeForStatement(n)
	case *ast.AlternativeForeachStatement:
		return c.compileAlternativeForeachStatement(n)

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
	case *ast.InterfaceDeclaration:
		return c.compileInterfaceDeclaration(n)
	case *ast.TraitDeclaration:
		return c.compileTraitDeclaration(n)
	case *ast.EnumDeclaration:
		return c.compileEnumDeclaration(n)
	case *ast.UseTraitStatement:
		return c.compileUseTraitStatement(n)
	case *ast.HookedPropertyDeclaration:
		return c.compileHookedPropertyDeclaration(n)

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
	leftResult := c.nextTemp - 1 // The last allocated temp contains the left result

	// Compile right operand
	err = c.compileNode(expr.Right)
	if err != nil {
		return err
	}
	rightResult := c.nextTemp - 1 // The last allocated temp contains the right result

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

	// Check if it's a property access (like $obj->property)
	if propAccess, ok := expr.Operand.(*ast.PropertyAccessExpression); ok {
		// Handle object property increment/decrement
		return c.compilePropertyIncDec(expr, propAccess)
	}

	// Check if it's an array access (like $array[$key])
	if arrayAccess, ok := expr.Operand.(*ast.ArrayAccessExpression); ok {
		// Handle array element increment/decrement
		return c.compileArrayIncDec(expr, arrayAccess)
	}

	return fmt.Errorf("increment/decrement can only be applied to variables, static properties, object properties, or array elements")
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
	// For static access increment/decrement like self::$counter++

	// Compile the class name (self, static, parent, or class name)
	var classOperandType opcodes.OpType
	var classOperand uint32

	switch class := staticAccess.Class.(type) {
	case *ast.IdentifierNode:
		className := class.Name
		classOperand = c.addConstant(values.NewString(className))
		classOperandType = opcodes.IS_CONST
	case *ast.Variable:
		className := class.Name
		if className == "self" || className == "static" || className == "parent" {
			classOperand = c.addConstant(values.NewString(className))
			classOperandType = opcodes.IS_CONST
		} else {
			return fmt.Errorf("unsupported variable in static access: %s", className)
		}
	default:
		return fmt.Errorf("unsupported class expression in static access increment")
	}

	// Get property name
	propName, ok := staticAccess.Property.(*ast.Variable)
	if !ok {
		return fmt.Errorf("expected variable property in static access increment")
	}
	propOperand := c.addConstant(values.NewString(propName.Name))

	// Allocate result for the post-increment operation
	result := c.allocateTemp()

	// For post-increment: first read the current value, then increment
	if expr.Operator == "++" {
		if expr.Prefix {
			// Pre-increment: ++self::$counter
			c.emit(opcodes.OP_PRE_INC,
				classOperandType, classOperand,
				opcodes.IS_CONST, propOperand,
				opcodes.IS_TMP_VAR, result)
		} else {
			// Post-increment: self::$counter++
			c.emit(opcodes.OP_POST_INC,
				classOperandType, classOperand,
				opcodes.IS_CONST, propOperand,
				opcodes.IS_TMP_VAR, result)
		}
	} else if expr.Operator == "--" {
		if expr.Prefix {
			// Pre-decrement: --self::$counter
			c.emit(opcodes.OP_PRE_DEC,
				classOperandType, classOperand,
				opcodes.IS_CONST, propOperand,
				opcodes.IS_TMP_VAR, result)
		} else {
			// Post-decrement: self::$counter--
			c.emit(opcodes.OP_POST_DEC,
				classOperandType, classOperand,
				opcodes.IS_CONST, propOperand,
				opcodes.IS_TMP_VAR, result)
		}
	} else {
		return fmt.Errorf("unsupported operator in static access increment: %s", expr.Operator)
	}

	return nil
}

func (c *Compiler) compilePropertyIncDec(expr *ast.UnaryExpression, propAccess *ast.PropertyAccessExpression) error {
	// Special handling for $this
	var objectOpType opcodes.OpType
	var objectOperand uint32

	if variable, ok := propAccess.Object.(*ast.Variable); ok && variable.Name == "$this" {
		// Use $this directly from slot 0 (when in class context)
		objectOpType = opcodes.IS_VAR
		objectOperand = 0 // $this is always in slot 0
	} else {
		// Compile object expression normally
		err := c.compileNode(propAccess.Object)
		if err != nil {
			return fmt.Errorf("failed to compile object expression in property increment: %w", err)
		}
		// Get the result from the previous compilation - it's in nextTemp-1
		objectOpType = opcodes.IS_TMP_VAR
		objectOperand = c.nextTemp - 1
	}

	// Handle property name - can be identifier or expression
	var propOpType opcodes.OpType
	var propOperand uint32

	if ident, ok := propAccess.Property.(*ast.IdentifierNode); ok {
		// Property is a simple identifier like "prop" in $obj->prop
		propOperand = c.addConstant(values.NewString(ident.Name))
		propOpType = opcodes.IS_CONST
	} else {
		// Property is an expression - compile it and use result
		err := c.compileNode(propAccess.Property)
		if err != nil {
			return fmt.Errorf("failed to compile property expression in property increment: %w", err)
		}
		// Use the result from property compilation
		propOperand = c.nextTemp - 1
		propOpType = opcodes.IS_TMP_VAR
	}

	// Use FETCH_OBJ_RW to get read-write access to the property
	propertyResult := c.allocateTemp()
	c.emit(opcodes.OP_FETCH_OBJ_RW, objectOpType, objectOperand, propOpType, propOperand, opcodes.IS_TMP_VAR, propertyResult)

	// Create constant 1 for increment/decrement
	oneConstant := c.addConstant(values.NewInt(1))

	// For post-increment, save the current value before modifying
	var currentVal uint32
	if expr.Operator == "++" || expr.Operator == "--" {
		if expr.Position.String() != "" { // Check if it's post-increment (this is a simplification)
			currentVal = c.allocateTemp()
			c.emit(opcodes.OP_QM_ASSIGN, opcodes.IS_TMP_VAR, propertyResult, 0, 0, opcodes.IS_TMP_VAR, currentVal)
		}
	}

	// Calculate new value
	newVal := c.allocateTemp()
	if expr.Operator == "++" {
		c.emit(opcodes.OP_ADD, opcodes.IS_TMP_VAR, propertyResult, opcodes.IS_CONST, oneConstant, opcodes.IS_TMP_VAR, newVal)
	} else { // "--"
		c.emit(opcodes.OP_SUB, opcodes.IS_TMP_VAR, propertyResult, opcodes.IS_CONST, oneConstant, opcodes.IS_TMP_VAR, newVal)
	}

	// Write the new value back to the property
	c.emit(opcodes.OP_ASSIGN_OBJ, objectOpType, objectOperand, propOpType, propOperand, opcodes.IS_TMP_VAR, newVal)

	// For pre-increment, the result is the new value; for post-increment, it's the old value
	result := c.allocateTemp()
	if expr.Position.String() != "" { // Post-increment/decrement
		c.emit(opcodes.OP_QM_ASSIGN, opcodes.IS_TMP_VAR, currentVal, 0, 0, opcodes.IS_TMP_VAR, result)
	} else { // Pre-increment/decrement
		c.emit(opcodes.OP_QM_ASSIGN, opcodes.IS_TMP_VAR, newVal, 0, 0, opcodes.IS_TMP_VAR, result)
	}

	return nil
}

func (c *Compiler) compileArrayIncDec(expr *ast.UnaryExpression, arrayAccess *ast.ArrayAccessExpression) error {
	// Compile array expression
	err := c.compileNode(arrayAccess.Array)
	if err != nil {
		return fmt.Errorf("failed to compile array expression in array increment: %w", err)
	}
	arrayResult := c.allocateTemp()
	c.emitMove(arrayResult)

	// Compile index expression
	if arrayAccess.Index == nil {
		return fmt.Errorf("array access requires an index for increment/decrement")
	}
	err = c.compileNode(*arrayAccess.Index)
	if err != nil {
		return fmt.Errorf("failed to compile index expression in array increment: %w", err)
	}
	indexResult := c.allocateTemp()
	c.emitMove(indexResult)

	// Use FETCH_DIM_RW to get read-write access to the array element
	elementResult := c.allocateTemp()
	c.emit(opcodes.OP_FETCH_DIM_RW, opcodes.IS_TMP_VAR, arrayResult, opcodes.IS_TMP_VAR, indexResult, opcodes.IS_TMP_VAR, elementResult)

	// Create constant 1 for increment/decrement
	oneConstant := c.addConstant(values.NewInt(1))

	// For post-increment, save the current value before modifying
	var currentVal uint32
	if expr.Operator == "++" || expr.Operator == "--" {
		if expr.Position.String() != "" { // Check if it's post-increment (this is a simplification)
			currentVal = c.allocateTemp()
			c.emit(opcodes.OP_QM_ASSIGN, opcodes.IS_TMP_VAR, elementResult, 0, 0, opcodes.IS_TMP_VAR, currentVal)
		}
	}

	// Calculate new value
	newVal := c.allocateTemp()
	if expr.Operator == "++" {
		c.emit(opcodes.OP_ADD, opcodes.IS_TMP_VAR, elementResult, opcodes.IS_CONST, oneConstant, opcodes.IS_TMP_VAR, newVal)
	} else { // "--"
		c.emit(opcodes.OP_SUB, opcodes.IS_TMP_VAR, elementResult, opcodes.IS_CONST, oneConstant, opcodes.IS_TMP_VAR, newVal)
	}

	// Write new value back to array element
	c.emit(opcodes.OP_ASSIGN_DIM, opcodes.IS_TMP_VAR, arrayResult, opcodes.IS_TMP_VAR, indexResult, opcodes.IS_TMP_VAR, newVal)

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
		} else if opcode == opcodes.OP_ASSIGN_OP {
			// Compound assignment: $var += value, $var *= value, etc.
			// Uses OP_ASSIGN_OP with operation type in Reserved field
			operationType := c.getOperationTypeForAssignmentOperator(expr.Operator)
			c.emitReserved(opcode, opcodes.IS_VAR, varSlot, opcodes.IS_TMP_VAR, valueResult, opcodes.IS_VAR, varSlot, operationType)
		} else {
			// Other assignment types (QM_ASSIGN, etc.)
			c.emit(opcode, opcodes.IS_VAR, varSlot, opcodes.IS_TMP_VAR, valueResult, opcodes.IS_VAR, varSlot)
		}
	} else if arrayAccess, ok := expr.Left.(*ast.ArrayAccessExpression); ok {
		// Handle array assignment: $arr[index] = value or nested assignments like $arr[key1][key2] = value
		return c.compileArrayAssignment(arrayAccess, valueResult)
	} else if listExpr, ok := expr.Left.(*ast.ListExpression); ok {
		// Handle list assignment: list($a, $b, $c) = $array
		return c.compileListAssignmentFromValue(listExpr, valueResult)
	} else if propAccess, ok := expr.Left.(*ast.PropertyAccessExpression); ok {
		// Handle property assignment: $obj->prop = value
		return c.compilePropertyAssignment(propAccess, valueResult)
	}
	return nil
}

// compilePropertyAssignment handles property assignments like $obj->prop = value
func (c *Compiler) compilePropertyAssignment(propAccess *ast.PropertyAccessExpression, valueResult uint32) error {
	// Special handling for $this
	var objectOpType opcodes.OpType
	var objectOperand uint32

	if variable, ok := propAccess.Object.(*ast.Variable); ok && variable.Name == "$this" {
		// Use $this directly from slot 0
		objectOpType = opcodes.IS_VAR
		objectOperand = 0 // $this is always in slot 0
	} else {
		// Compile object expression normally
		err := c.compileNode(propAccess.Object)
		if err != nil {
			return err
		}
		// Get the result from the previous compilation
		objectOpType = opcodes.IS_TMP_VAR
		objectOperand = c.nextTemp - 1
	}

	// Handle property name as string literal
	var propConstant uint32
	if ident, ok := propAccess.Property.(*ast.IdentifierNode); ok {
		// Property is a simple identifier like "value" in $this->value
		propConstant = c.addConstant(values.NewString(ident.Name))
		c.emit(opcodes.OP_ASSIGN_OBJ, objectOpType, objectOperand, opcodes.IS_CONST, propConstant, opcodes.IS_TMP_VAR, valueResult)
	} else {
		// Property is an expression - compile it normally
		err := c.compileNode(propAccess.Property)
		if err != nil {
			return err
		}
		propResult := c.nextTemp - 1
		c.emit(opcodes.OP_ASSIGN_OBJ, objectOpType, objectOperand, opcodes.IS_TMP_VAR, propResult, opcodes.IS_TMP_VAR, valueResult)
	}

	return nil
}

// compileArrayAssignment handles array assignments including nested ones
// For $arr[key1][key2] = value, this recursively processes the array access chain
func (c *Compiler) compileArrayAssignment(arrayAccess *ast.ArrayAccessExpression, valueResult uint32) error {
	// Check if the array part is a simple variable or another array access
	if baseVar, ok := arrayAccess.Array.(*ast.Variable); ok {
		// Simple case: $arr[index] = value
		arraySlot := c.getVariableSlot(baseVar.Name)

		if arrayAccess.Index == nil {
			// Array append: $arr[] = value
			c.emit(opcodes.OP_ADD_ARRAY_ELEMENT, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, valueResult, opcodes.IS_VAR, arraySlot)
		} else {
			// Array index assignment: $arr[index] = value
			err := c.compileNode(*arrayAccess.Index)
			if err != nil {
				return err
			}
			indexResult := c.nextTemp - 1

			// Emit ASSIGN_DIM instruction: array[index] = value
			// We'll use a different approach: put value in Op2, key gets compiled separately
			c.emit(opcodes.OP_ASSIGN_DIM, opcodes.IS_VAR, arraySlot, opcodes.IS_TMP_VAR, indexResult, opcodes.IS_TMP_VAR, valueResult)
		}
	} else if nestedArrayAccess, ok := arrayAccess.Array.(*ast.ArrayAccessExpression); ok {
		// Nested case: $arr[key1][key2] = value
		// First, ensure the nested array structure exists by fetching for write access

		// Compile the nested array access for write (this will create intermediate arrays if needed)
		err := c.compileFetchDimWrite(nestedArrayAccess)
		if err != nil {
			return err
		}
		nestedArrayResult := c.nextTemp - 1

		if arrayAccess.Index == nil {
			// Nested array append: $arr[key1][] = value
			c.emit(opcodes.OP_ADD_ARRAY_ELEMENT, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, valueResult, opcodes.IS_TMP_VAR, nestedArrayResult)
		} else {
			// Nested array index assignment: $arr[key1][key2] = value
			err := c.compileNode(*arrayAccess.Index)
			if err != nil {
				return err
			}
			indexResult := c.nextTemp - 1

			// Emit ASSIGN_DIM instruction with the nested array
			c.emit(opcodes.OP_ASSIGN_DIM, opcodes.IS_TMP_VAR, nestedArrayResult, opcodes.IS_TMP_VAR, indexResult, opcodes.IS_TMP_VAR, valueResult)
		}
	} else if propAccess, ok := arrayAccess.Array.(*ast.PropertyAccessExpression); ok {
		// Property access case: $obj->property[index] = value
		// First compile the property access to get the array
		err := c.compilePropertyAccess(propAccess)
		if err != nil {
			return err
		}
		propResult := c.nextTemp - 1

		if arrayAccess.Index == nil {
			// Property array append: $obj->property[] = value
			c.emit(opcodes.OP_ADD_ARRAY_ELEMENT, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, valueResult, opcodes.IS_TMP_VAR, propResult)
		} else {
			// Property array index assignment: $obj->property[index] = value
			err := c.compileNode(*arrayAccess.Index)
			if err != nil {
				return err
			}
			indexResult := c.nextTemp - 1

			// Emit ASSIGN_DIM instruction with the property access result
			c.emit(opcodes.OP_ASSIGN_DIM, opcodes.IS_TMP_VAR, propResult, opcodes.IS_TMP_VAR, indexResult, opcodes.IS_TMP_VAR, valueResult)
		}
	} else {
		return fmt.Errorf("unsupported array expression type in assignment: %T", arrayAccess.Array)
	}

	return nil
}

// compileFetchDimWrite compiles an array access for write operations, ensuring intermediate arrays exist
func (c *Compiler) compileFetchDimWrite(arrayAccess *ast.ArrayAccessExpression) error {
	if baseVar, ok := arrayAccess.Array.(*ast.Variable); ok {
		// Simple case: fetch $arr[index] for writing
		arraySlot := c.getVariableSlot(baseVar.Name)

		if arrayAccess.Index == nil {
			return fmt.Errorf("cannot fetch array with empty index for writing")
		}

		err := c.compileNode(*arrayAccess.Index)
		if err != nil {
			return err
		}
		indexResult := c.nextTemp - 1
		resultTemp := c.allocateTemp()

		// Emit FETCH_DIM_W instruction to get writable reference to array element
		c.emit(opcodes.OP_FETCH_DIM_W, opcodes.IS_VAR, arraySlot, opcodes.IS_TMP_VAR, indexResult, opcodes.IS_TMP_VAR, resultTemp)

		return nil
	} else if nestedArrayAccess, ok := arrayAccess.Array.(*ast.ArrayAccessExpression); ok {
		// Nested case: recursively fetch nested array structure for writing
		err := c.compileFetchDimWrite(nestedArrayAccess)
		if err != nil {
			return err
		}
		nestedResult := c.nextTemp - 1

		if arrayAccess.Index == nil {
			return fmt.Errorf("cannot fetch array with empty index for writing")
		}

		err = c.compileNode(*arrayAccess.Index)
		if err != nil {
			return err
		}
		indexResult := c.nextTemp - 1
		resultTemp := c.allocateTemp()

		// Fetch from the nested array result
		c.emit(opcodes.OP_FETCH_DIM_W, opcodes.IS_TMP_VAR, nestedResult, opcodes.IS_TMP_VAR, indexResult, opcodes.IS_TMP_VAR, resultTemp)

		return nil
	} else if propAccess, ok := arrayAccess.Array.(*ast.PropertyAccessExpression); ok {
		// Property access case: $obj->property[index] for writing
		// First compile the property access to get the array
		err := c.compilePropertyAccess(propAccess)
		if err != nil {
			return err
		}
		propResult := c.nextTemp - 1

		if arrayAccess.Index == nil {
			return fmt.Errorf("cannot fetch array with empty index for writing")
		}

		err = c.compileNode(*arrayAccess.Index)
		if err != nil {
			return err
		}
		indexResult := c.nextTemp - 1
		resultTemp := c.allocateTemp()

		// Fetch from the property access result for writing
		c.emit(opcodes.OP_FETCH_DIM_W, opcodes.IS_TMP_VAR, propResult, opcodes.IS_TMP_VAR, indexResult, opcodes.IS_TMP_VAR, resultTemp)

		return nil
	} else {
		return fmt.Errorf("unsupported array expression type: %T", arrayAccess.Array)
	}
}

func (c *Compiler) getOpcodeForAssignmentOperator(operator string) opcodes.Opcode {
	switch operator {
	// Simple assignment
	case "=":
		return opcodes.OP_ASSIGN
	case "=&":
		return opcodes.OP_ASSIGN_REF

	// All compound assignments use OP_ASSIGN_OP
	case "+=", "-=", "*=", "/=", "%=", "**=", ".=", "&=", "|=", "^=", "<<=", ">>=":
		return opcodes.OP_ASSIGN_OP

	// Null coalescing assignment
	case "??=":
		return opcodes.OP_QM_ASSIGN

	default:
		return opcodes.OP_ASSIGN
	}
}

func (c *Compiler) getOperationTypeForAssignmentOperator(operator string) byte {
	// Return Zend Engine operation types that match PHP's implementation
	switch operator {
	case "+=":
		return 1 // ZEND_ADD
	case "-=":
		return 2 // ZEND_SUB
	case "*=":
		return 3 // ZEND_MUL
	case "/=":
		return 4 // ZEND_DIV
	case "%=":
		return 5 // ZEND_MOD
	case "**=":
		return 6 // ZEND_POW
	case ".=":
		return 8 // ZEND_CONCAT
	case "&=":
		return 9 // ZEND_BW_AND
	case "|=":
		return 10 // ZEND_BW_OR
	case "^=":
		return 11 // ZEND_BW_XOR
	case "<<=":
		return 12 // ZEND_SL
	case ">>=":
		return 13 // ZEND_SR
	default:
		return 0 // No operation
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

	if expr.Kind == ast.IntegerKind {
		// 使用预转换的整数值
		constant = c.addConstant(values.NewInt(expr.IntValue))
	} else if expr.Kind == ast.FloatKind {
		// 使用预转换的浮点值
		constant = c.addConstant(values.NewFloat(expr.FloatValue))
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
		} else if spreadExpr, ok := element.(*ast.SpreadExpression); ok {
			// Handle spread expression (...$array)
			err := c.compileNode(spreadExpr.Argument)
			if err != nil {
				return err
			}
			valueResult := c.allocateTemp()
			c.emitMove(valueResult)

			// Use OP_ADD_ARRAY_UNPACK for spread expressions
			c.emit(opcodes.OP_ADD_ARRAY_UNPACK, opcodes.IS_TMP_VAR, valueResult, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, result)
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
	// Special handling for $this
	var objectOpType opcodes.OpType
	var objectOperand uint32

	if variable, ok := expr.Object.(*ast.Variable); ok && variable.Name == "$this" {
		// Use $this directly from slot 0 (when in class context)
		objectOpType = opcodes.IS_VAR
		objectOperand = 0 // $this is always in slot 0
	} else {
		// Compile object expression normally
		err := c.compileNode(expr.Object)
		if err != nil {
			return err
		}
		// Get the result from the previous compilation - it's in nextTemp-1
		objectOpType = opcodes.IS_TMP_VAR
		objectOperand = c.nextTemp - 1
	}

	// Handle property name - can be identifier or expression
	var propOpType opcodes.OpType
	var propOperand uint32

	if ident, ok := expr.Property.(*ast.IdentifierNode); ok {
		// Property is a simple identifier like "prop" in $obj->prop
		propOperand = c.addConstant(values.NewString(ident.Name))
		propOpType = opcodes.IS_CONST
	} else {
		// Property is an expression - compile it and use result
		err := c.compileNode(expr.Property)
		if err != nil {
			return err
		}
		// Use the result from property compilation
		propOperand = c.nextTemp - 1
		propOpType = opcodes.IS_TMP_VAR
	}

	// Emit the FETCH_OBJ_R instruction
	result := c.allocateTemp()
	c.emit(opcodes.OP_FETCH_OBJ_R, objectOpType, objectOperand, propOpType, propOperand, opcodes.IS_TMP_VAR, result)

	return nil
}

func (c *Compiler) compileFunctionCall(expr *ast.CallExpression) error {
	// Check if this is a static method call (e.g., parent::__construct, Class::method)
	if staticAccess, ok := expr.Callee.(*ast.StaticAccessExpression); ok {
		return c.compileStaticMethodCall(expr, staticAccess)
	}

	// Compile callee expression for regular function calls
	err := c.compileNode(expr.Callee)
	if err != nil {
		return err
	}
	// Use the temp that was allocated by compileNode
	calleeResult := c.nextTemp - 1

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

func (c *Compiler) compileStaticMethodCall(callExpr *ast.CallExpression, staticAccess *ast.StaticAccessExpression) error {
	// Handle static method calls like parent::__construct(), Class::method(), etc.

	// Get class name
	var className string
	switch class := staticAccess.Class.(type) {
	case *ast.IdentifierNode:
		className = class.Name
	case *ast.Variable:
		// Handle parent::, self::, static::
		className = class.Name
	default:
		return fmt.Errorf("unsupported class expression in static method call: %T", staticAccess.Class)
	}

	// Get method name
	var methodName string
	if method, ok := staticAccess.Property.(*ast.IdentifierNode); ok {
		methodName = method.Name
	} else {
		return fmt.Errorf("unsupported method name type in static method call: %T", staticAccess.Property)
	}

	// Get number of arguments
	var numArgs uint32
	if callExpr.Arguments != nil {
		numArgs = uint32(len(callExpr.Arguments.Arguments))
	}

	// Initialize static method call
	classConstant := c.addConstant(values.NewString(className))
	methodConstant := c.addConstant(values.NewString(methodName))
	argCountConstant := c.addConstant(values.NewInt(int64(numArgs)))

	c.emit(opcodes.OP_INIT_STATIC_METHOD_CALL,
		opcodes.IS_CONST, classConstant,
		opcodes.IS_CONST, methodConstant,
		opcodes.IS_CONST, argCountConstant)

	// Compile and send arguments
	if callExpr.Arguments != nil {
		for i, arg := range callExpr.Arguments.Arguments {
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

	// Execute static method call
	result := c.allocateTemp()
	c.emit(opcodes.OP_STATIC_METHOD_CALL, opcodes.IS_UNUSED, 0, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, result)

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
	// Handle short ternary (expr ?: alternate) separately
	if expr.Consequent == nil {
		return c.compileShortTernary(expr)
	}

	// Step 1: Compile condition
	err := c.compileNode(expr.Test)
	if err != nil {
		return err
	}
	condResult := c.nextTemp - 1

	// Step 2: Allocate result temporary FIRST (like PHP does)
	result := c.allocateTemp()

	// Step 3: Create conditional jump to false branch
	jmpzLabel := c.generateLabel()
	c.emitJumpZ(opcodes.IS_TMP_VAR, condResult, jmpzLabel)

	// Step 4: Compile true branch and assign directly to result
	err = c.compileNode(expr.Consequent)
	if err != nil {
		return err
	}
	trueBranchResult := c.nextTemp - 1
	c.emit(opcodes.OP_QM_ASSIGN, opcodes.IS_TMP_VAR, trueBranchResult, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, result)

	// Step 5: Jump over false branch
	jmpLabel := c.generateLabel()
	c.emitJump(opcodes.OP_JMP, opcodes.IS_CONST, 0, jmpLabel)

	// Step 6: False branch - assign to SAME result temporary
	c.placeLabel(jmpzLabel)
	err = c.compileNode(expr.Alternate)
	if err != nil {
		return err
	}
	falseBranchResult := c.nextTemp - 1
	c.emit(opcodes.OP_QM_ASSIGN, opcodes.IS_TMP_VAR, falseBranchResult, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, result)

	// Step 7: Both branches converge here
	c.placeLabel(jmpLabel)

	// CRITICAL: Set nextTemp to point after our result so that subsequent
	// operations see our result as the "most recent temporary"
	c.nextTemp = result + 1

	return nil
}

func (c *Compiler) compileShortTernary(expr *ast.TernaryExpression) error {
	// Short ternary: expr ?: alternate
	// If expr is truthy, use expr; otherwise use alternate

	// Step 1: Compile the test expression
	err := c.compileNode(expr.Test)
	if err != nil {
		return err
	}
	testResult := c.nextTemp - 1

	// Step 2: Allocate result temporary FIRST
	result := c.allocateTemp()

	// Step 3: Create jump to alternate if test is false
	jmpzLabel := c.generateLabel()
	c.emitJumpZ(opcodes.IS_TMP_VAR, testResult, jmpzLabel)

	// Step 4: True case: use the test result itself, assign to result
	c.emit(opcodes.OP_QM_ASSIGN, opcodes.IS_TMP_VAR, testResult, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, result)

	// Step 5: Jump over alternate
	jmpLabel := c.generateLabel()
	c.emitJump(opcodes.OP_JMP, opcodes.IS_CONST, 0, jmpLabel)

	// Step 6: False case: compile alternate and assign to SAME result
	c.placeLabel(jmpzLabel)
	err = c.compileNode(expr.Alternate)
	if err != nil {
		return err
	}
	alternateResult := c.nextTemp - 1
	c.emit(opcodes.OP_QM_ASSIGN, opcodes.IS_TMP_VAR, alternateResult, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, result)

	// Step 7: Both paths converge
	c.placeLabel(jmpLabel)

	// CRITICAL: Set nextTemp to point after our result so that subsequent
	// operations see our result as the "most recent temporary"
	c.nextTemp = result + 1

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

func (c *Compiler) emitReserved(opcode opcodes.Opcode, op1Type opcodes.OpType, op1 uint32, op2Type opcodes.OpType, op2 uint32, resultType opcodes.OpType, result uint32, reserved byte) {
	opType1, opType2 := opcodes.EncodeOpTypes(op1Type, op2Type, resultType)

	instruction := opcodes.Instruction{
		Opcode:   opcode,
		OpType1:  opType1,
		OpType2:  opType2,
		Reserved: reserved,
		Op1:      op1,
		Op2:      op2,
		Result:   result,
	}

	c.instructions = append(c.instructions, instruction)
}

func (c *Compiler) emitWithTypes(opcode opcodes.Opcode, opType1 byte, opType2 byte, op1 uint32, op2 uint32, result uint32) {
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
	source := c.nextTemp - 2 // -1 for the target that was just allocated, -1 more for the actual source
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
	if resultTemp != c.nextTemp-1 {
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
	// For class methods, check within the current class; for global functions, check globally
	if c.currentClass != nil {
		// This is a class method - check for conflicts within the current class
		if _, exists := c.currentClass.Methods[funcName]; exists {
			return fmt.Errorf("function %s already declared", funcName)
		}
	} else {
		// This is a global function - check globally
		if _, exists := c.functions[funcName]; exists {
			return fmt.Errorf("function %s already declared", funcName)
		}
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

	// Store the function in the appropriate location
	if c.currentClass != nil {
		// Store as a class method
		c.currentClass.Methods[funcName] = function

		// Also register in unified registry if available
		if registry.GlobalRegistry != nil {
			// Get or create class in registry
			classDesc, err := registry.GlobalRegistry.GetClass(c.currentClass.Name)
			if err != nil {
				// Class doesn't exist in registry, create it
				classDesc = &registry.ClassDescriptor{
					Name:       c.currentClass.Name,
					Parent:     c.currentClass.Parent,
					IsAbstract: c.currentClass.IsAbstract,
					IsFinal:    c.currentClass.IsFinal,
					Properties: make(map[string]*registry.PropertyDescriptor),
					Methods:    make(map[string]*registry.MethodDescriptor),
					Constants:  make(map[string]*registry.ConstantDescriptor),
				}
				registry.GlobalRegistry.RegisterClass(classDesc)
			}

			// Convert VM parameters to registry parameters
			registryParams := make([]registry.ParameterDescriptor, len(function.Parameters))
			for i, param := range function.Parameters {
				registryParams[i] = registry.ParameterDescriptor{
					Name:         param.Name,
					Type:         param.Type,
					IsReference:  param.IsReference,
					HasDefault:   param.HasDefault,
					DefaultValue: param.DefaultValue,
				}
			}

			// Create method implementation using bytecode
			methodImpl := &registry.BytecodeMethodImpl{
				Instructions: function.Instructions,
				Constants:    function.Constants,
				LocalVars:    len(function.Parameters),
			}

			// Register method in class
			classDesc.Methods[funcName] = &registry.MethodDescriptor{
				Name:           funcName,
				Visibility:     "public", // TODO: Parse actual visibility from decl
				IsStatic:       false,    // TODO: Parse actual static modifier from decl
				IsAbstract:     false,
				IsFinal:        false,
				Parameters:     registryParams,
				Implementation: methodImpl,
				IsVariadic:     function.IsVariadic,
			}
		}
	} else {
		// Store as a global function
		c.functions[funcName] = function
	}

	// Restore compiler state
	c.popScope()
	c.instructions = oldInstructions
	c.constants = oldConstants

	// Emit function declaration instruction
	nameConstant := c.addConstant(values.NewString(funcName))
	c.emit(opcodes.OP_DECLARE_FUNCTION, opcodes.IS_CONST, nameConstant, 0, 0, 0, 0)

	return nil
}

func (c *Compiler) compileAnonymousFunction(expr *ast.AnonymousFunctionExpression) error {
	// Generate a unique name for the anonymous function using the counter
	anonName := fmt.Sprintf("__anonymous_%d", c.nextAnonFunction)
	c.nextAnonFunction++

	// Create new function
	function := &vm.Function{
		Name:         anonName,
		Instructions: make([]opcodes.Instruction, 0),
		Constants:    make([]*values.Value, 0),
		Parameters:   make([]vm.Parameter, 0),
		IsVariadic:   false,
		IsGenerator:  false,
	}

	// Compile parameters
	if expr.Parameters != nil {
		for _, param := range expr.Parameters.Parameters {
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
	if expr.Parameters != nil {
		for _, param := range expr.Parameters.Parameters {
			if nameNode, ok := param.Name.(*ast.IdentifierNode); ok {
				// Register parameter name in function scope
				c.getOrCreateVariable(nameNode.Name)
			}
		}
	}

	// Compile function body
	for _, stmt := range expr.Body {
		err := c.compileNode(stmt)
		if err != nil {
			c.popScope()
			c.instructions = oldInstructions
			c.constants = oldConstants
			return fmt.Errorf("error compiling anonymous function: %v", err)
		}
	}

	// Add implicit return if needed
	if len(c.instructions) == 0 || c.instructions[len(c.instructions)-1].Opcode != opcodes.OP_RETURN {
		c.emit(opcodes.OP_RETURN, opcodes.IS_CONST, c.addConstant(values.NewNull()), 0, 0, 0, 0)
	}

	// Store compiled function
	function.Instructions = c.instructions
	function.Constants = c.constants

	// Store the function
	c.functions[anonName] = function

	// Restore compiler state
	c.popScope()
	c.instructions = oldInstructions
	c.constants = oldConstants

	// Create closure at runtime
	functionConstant := c.addConstant(values.NewString(anonName))
	closureResult := c.allocateTemp()
	c.emit(opcodes.OP_CREATE_CLOSURE, opcodes.IS_CONST, functionConstant, 0, 0, opcodes.IS_TMP_VAR, closureResult)

	// Bind use variables
	if expr.UseClause != nil {
		for _, useVar := range expr.UseClause {
			if varExpr, ok := useVar.(*ast.Variable); ok {
				// Normal variable binding (by value)
				err := c.compileNode(varExpr)
				if err != nil {
					return fmt.Errorf("error compiling use variable %s: %v", varExpr.Name, err)
				}
				varResult := c.allocateTemp()
				c.emitMove(varResult)

				// Bind the variable to the closure
				varNameConstant := c.addConstant(values.NewString(varExpr.Name))
				c.emit(opcodes.OP_BIND_USE_VAR, opcodes.IS_TMP_VAR, closureResult, opcodes.IS_CONST, varNameConstant, opcodes.IS_TMP_VAR, varResult)
			} else if refExpr, ok := useVar.(*ast.UnaryExpression); ok && refExpr.Operator == "&" {
				// Reference variable binding (&$var)
				if varExpr, ok := refExpr.Operand.(*ast.Variable); ok {
					// Get the variable value from current scope
					err := c.compileNode(varExpr)
					if err != nil {
						return fmt.Errorf("error compiling reference use variable %s: %v", varExpr.Name, err)
					}
					varResult := c.allocateTemp()
					c.emitMove(varResult)

					// Bind the variable to the closure with reference flag
					varNameConstant := c.addConstant(values.NewString(varExpr.Name))
					opType1, opType2 := opcodes.EncodeOpTypesWithFlags(opcodes.IS_TMP_VAR, opcodes.IS_CONST, opcodes.IS_TMP_VAR, opcodes.EXT_FLAG_REFERENCE)
					c.emitWithTypes(opcodes.OP_BIND_USE_VAR, opType1, opType2, closureResult, varNameConstant, varResult)
				}
			}
		}
	}

	// Ensure the closure is available as the result of this expression compilation
	// The closure is in closureResult, but other parts of the compiler expect the result
	// to be in the most recently allocated temp. We need to make sure that's the case.
	// If closureResult is not the most recent temp, we need to move it.
	expectedResultTemp := c.nextTemp - 1
	if closureResult != expectedResultTemp {
		finalResult := c.allocateTemp()
		c.emit(opcodes.OP_QM_ASSIGN, opcodes.IS_TMP_VAR, closureResult, 0, 0, opcodes.IS_TMP_VAR, finalResult)
	}

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
		Name:       className,
		Parent:     "",
		Properties: make(map[string]*vm.Property),
		Methods:    make(map[string]*vm.Function),
		Constants:  make(map[string]*vm.ClassConstant),
		IsAbstract: false,
		IsFinal:    false,
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
			class.Parent = parent.Name
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

	// Set current class context in VM for anonymous classes
	c.emit(opcodes.OP_SET_CURRENT_CLASS, opcodes.IS_CONST, nameConstant, 0, 0, 0, 0)

	// Set parent class if exists
	if class.Parent != "" {
		parentConstant := c.addConstant(values.NewString(class.Parent))
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

	// Clear current class context in VM for anonymous classes
	c.emit(opcodes.OP_CLEAR_CURRENT_CLASS, 0, 0, 0, 0, 0, 0)

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
			if defVal.Kind == ast.IntegerKind {
				// 使用预转换的整数值
				defaultValue = values.NewInt(defVal.IntValue)
			} else if defVal.Kind == ast.FloatKind {
				// 使用预转换的浮点值
				defaultValue = values.NewFloat(defVal.FloatValue)
			}
		case *ast.StringLiteral:
			defaultValue = values.NewString(defVal.Value)
		case *ast.BooleanLiteral:
			defaultValue = values.NewBool(defVal.Value)
		case *ast.NullLiteral:
			defaultValue = values.NewNull()
		case *ast.ArrayExpression:
			// Try to evaluate the array expression as a constant array
			arrayValue, err := c.evaluateConstantArrayExpression(defVal)
			if err != nil {
				return fmt.Errorf("invalid array default value for property %s: %v", propName, err)
			}
			defaultValue = arrayValue
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

	// Create metadata as a serialized object with visibility, static flag, and default value
	metadata := values.NewArray()
	metadata.ArraySet(values.NewString("visibility"), values.NewString(decl.Visibility))
	metadata.ArraySet(values.NewString("static"), values.NewBool(decl.Static))
	if defaultValue != nil {
		metadata.ArraySet(values.NewString("defaultValue"), defaultValue)
	} else {
		metadata.ArraySet(values.NewString("defaultValue"), values.NewNull())
	}

	metadataConstant := c.addConstant(metadata)

	// Emit property declaration with complete metadata
	c.emit(opcodes.OP_DECLARE_PROPERTY,
		opcodes.IS_CONST, classNameConstant,
		opcodes.IS_CONST, propNameConstant,
		opcodes.IS_CONST, metadataConstant)

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

	// Determine visibility - default to public if not specified
	visibility := "public"
	if decl.Visibility != "" {
		visibility = decl.Visibility
	}

	// Validate visibility + modifier combinations (PHP rules)
	if visibility == "private" && decl.IsFinal {
		return fmt.Errorf("private constant cannot be final as it is not visible to other classes")
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

		// Evaluate constant value using a dedicated method
		constValue, err := c.evaluateClassConstantExpression(constDeclarator.Value)
		if err != nil {
			return fmt.Errorf("failed to evaluate constant %s: %v", constName, err)
		}

		// Type validation for PHP 8.3+ typed constants
		var typeHint string
		if decl.Type != nil {
			// Extract type hint from AST
			typeHint = decl.Type.Name
			// Validate constant value against type
			if err := c.validateConstantType(constValue, typeHint); err != nil {
				return fmt.Errorf("constant %s::%s type mismatch: %v", c.currentClass.Name, constName, err)
			}
		}

		// Create class constant with full metadata
		classConstant := &vm.ClassConstant{
			Name:       constName,
			Value:      constValue,
			Visibility: visibility,
			Type:       typeHint,
			IsFinal:    decl.IsFinal,
			IsAbstract: decl.IsAbstract,
		}

		// Add constant to current class
		c.currentClass.Constants[constName] = classConstant

		// Emit opcode to register the class constant in the runtime registry
		classNameConst := c.addConstant(values.NewString(c.currentClass.Name))
		constNameConst := c.addConstant(values.NewString(constName))
		constValueConst := c.addConstant(constValue)

		c.emit(opcodes.OP_DECLARE_CLASS_CONST, opcodes.IS_CONST, classNameConst, opcodes.IS_CONST, constNameConst, opcodes.IS_CONST, constValueConst)
	}

	return nil
}

// evaluateClassConstantExpression evaluates a class constant expression at compile time
func (c *Compiler) evaluateClassConstantExpression(expr ast.Expression) (*values.Value, error) {
	switch val := expr.(type) {
	case *ast.NumberLiteral:
		if val.Kind == ast.IntegerKind {
			// 使用预转换的整数值
			return values.NewInt(val.IntValue), nil
		} else if val.Kind == ast.FloatKind {
			// 使用预转换的浮点值
			return values.NewFloat(val.FloatValue), nil
		} else {
			return nil, fmt.Errorf("unsupported number kind: %s", val.Kind)
		}
	case *ast.StringLiteral:
		return values.NewString(val.Value), nil
	case *ast.BooleanLiteral:
		return values.NewBool(val.Value), nil
	case *ast.NullLiteral:
		return values.NewNull(), nil
	case *ast.ArrayExpression:
		// Create new array
		arrayValue := values.NewArray()
		arrayData := arrayValue.Data.(*values.Array)

		// Process each element
		for _, element := range val.Elements {
			if arrayElement, ok := element.(*ast.ArrayElementExpression); ok {
				// Handle key => value pairs
				var keyValue interface{}
				if arrayElement.Key != nil {
					// Evaluate the key
					keyVal, err := c.evaluateClassConstantExpression(arrayElement.Key)
					if err != nil {
						return nil, fmt.Errorf("failed to evaluate array key: %w", err)
					}
					// Convert key to appropriate type for array indexing
					switch keyVal.Type {
					case values.TypeString:
						keyValue = keyVal.Data.(string)
					case values.TypeInt:
						keyValue = keyVal.Data.(int64)
					default:
						return nil, fmt.Errorf("unsupported array key type: %s", keyVal.Type)
					}
				} else {
					// Auto-index
					keyValue = arrayData.NextIndex
					arrayData.NextIndex++
				}

				// Evaluate the value
				valueVal, err := c.evaluateClassConstantExpression(arrayElement.Value)
				if err != nil {
					return nil, fmt.Errorf("failed to evaluate array value: %w", err)
				}

				// Set the element
				arrayData.Elements[keyValue] = valueVal
			} else {
				// Direct element (not ArrayElementExpression) - use auto-indexing
				valueVal, err := c.evaluateClassConstantExpression(element)
				if err != nil {
					return nil, fmt.Errorf("failed to evaluate array element: %w", err)
				}

				arrayData.Elements[arrayData.NextIndex] = valueVal
				arrayData.NextIndex++
			}
		}

		return arrayValue, nil
	case *ast.IdentifierNode:
		// Handle simple constant references like true, false, null
		switch val.Name {
		case "true":
			return values.NewBool(true), nil
		case "false":
			return values.NewBool(false), nil
		case "null":
			return values.NewNull(), nil
		default:
			return nil, fmt.Errorf("undefined constant reference: %s", val.Name)
		}

	case *ast.BinaryExpression:
		// Evaluate both operands
		left, err := c.evaluateClassConstantExpression(val.Left)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate left operand: %w", err)
		}
		right, err := c.evaluateClassConstantExpression(val.Right)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate right operand: %w", err)
		}

		// Perform the binary operation based on the operator
		return c.evaluateBinaryOperation(val.Operator, left, right)

	case *ast.UnaryExpression:
		// Evaluate the operand
		operand, err := c.evaluateClassConstantExpression(val.Operand)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate unary operand: %w", err)
		}

		// Perform the unary operation
		return c.evaluateUnaryOperation(val.Operator, operand)

	case *ast.TernaryExpression:
		// Evaluate the condition
		condition, err := c.evaluateClassConstantExpression(val.Test)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate ternary condition: %w", err)
		}

		// Choose branch based on condition truth value
		if c.isTruthy(condition) {
			if val.Consequent != nil {
				return c.evaluateClassConstantExpression(val.Consequent)
			}
			// If no true expression, return condition (PHP behavior)
			return condition, nil
		} else {
			return c.evaluateClassConstantExpression(val.Alternate)
		}

	case *ast.CoalesceExpression:
		// Evaluate the left operand
		left, err := c.evaluateClassConstantExpression(val.Left)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate coalesce left operand: %w", err)
		}

		// If left is not null, return it
		if !left.IsNull() {
			return left, nil
		}

		// Otherwise evaluate and return the right operand
		return c.evaluateClassConstantExpression(val.Right)

	case *ast.ClassConstantAccessExpression:
		// Handle class constant access like ClassName::CONSTANT_NAME, self::CONSTANT, parent::CONSTANT
		return c.evaluateClassConstantAccess(val)

	default:
		return nil, fmt.Errorf("unsupported constant expression type: %T", expr)
	}
}

// evaluateBinaryOperation performs binary operations on constant values
func (c *Compiler) evaluateBinaryOperation(operator string, left, right *values.Value) (*values.Value, error) {
	switch operator {
	// Arithmetic operations
	case "+":
		return c.performArithmetic(left, right, func(a, b int64) int64 { return a + b }, func(a, b float64) float64 { return a + b })
	case "-":
		return c.performArithmetic(left, right, func(a, b int64) int64 { return a - b }, func(a, b float64) float64 { return a - b })
	case "*":
		return c.performArithmetic(left, right, func(a, b int64) int64 { return a * b }, func(a, b float64) float64 { return a * b })
	case "/":
		if (right.IsInt() && right.Data.(int64) == 0) || (right.IsFloat() && right.Data.(float64) == 0) {
			return nil, fmt.Errorf("division by zero")
		}
		return c.performArithmetic(left, right, func(a, b int64) int64 { return a / b }, func(a, b float64) float64 { return a / b })
	case "%":
		if !left.IsInt() || !right.IsInt() {
			return nil, fmt.Errorf("modulo operation requires integers")
		}
		if right.Data.(int64) == 0 {
			return nil, fmt.Errorf("modulo by zero")
		}
		return values.NewInt(left.Data.(int64) % right.Data.(int64)), nil
	case "**":
		if left.IsInt() && right.IsInt() {
			leftVal := left.Data.(int64)
			rightVal := right.Data.(int64)
			if rightVal < 0 {
				return values.NewFloat(math.Pow(float64(leftVal), float64(rightVal))), nil
			}
			result := int64(1)
			for i := int64(0); i < rightVal; i++ {
				result *= leftVal
			}
			return values.NewInt(result), nil
		}
		leftFloat := c.toFloat(left)
		rightFloat := c.toFloat(right)
		return values.NewFloat(math.Pow(leftFloat, rightFloat)), nil

	// String concatenation
	case ".":
		leftStr := c.toString(left)
		rightStr := c.toString(right)
		return values.NewString(leftStr + rightStr), nil

	// Comparison operations
	case "==":
		return values.NewBool(c.isEqual(left, right)), nil
	case "!=":
		return values.NewBool(!c.isEqual(left, right)), nil
	case "<":
		cmp, err := c.compare(left, right)
		if err != nil {
			return nil, err
		}
		return values.NewBool(cmp < 0), nil
	case "<=":
		cmp, err := c.compare(left, right)
		if err != nil {
			return nil, err
		}
		return values.NewBool(cmp <= 0), nil
	case ">":
		cmp, err := c.compare(left, right)
		if err != nil {
			return nil, err
		}
		return values.NewBool(cmp > 0), nil
	case ">=":
		cmp, err := c.compare(left, right)
		if err != nil {
			return nil, err
		}
		return values.NewBool(cmp >= 0), nil
	case "<=>":
		cmp, err := c.compare(left, right)
		if err != nil {
			return nil, err
		}
		return values.NewInt(int64(cmp)), nil

	// Logical operations
	case "&&":
		return values.NewBool(c.isTruthy(left) && c.isTruthy(right)), nil
	case "||":
		return values.NewBool(c.isTruthy(left) || c.isTruthy(right)), nil

	// Bitwise operations
	case "&":
		if !left.IsInt() || !right.IsInt() {
			return nil, fmt.Errorf("bitwise AND requires integers")
		}
		return values.NewInt(left.Data.(int64) & right.Data.(int64)), nil
	case "|":
		if !left.IsInt() || !right.IsInt() {
			return nil, fmt.Errorf("bitwise OR requires integers")
		}
		return values.NewInt(left.Data.(int64) | right.Data.(int64)), nil
	case "^":
		if !left.IsInt() || !right.IsInt() {
			return nil, fmt.Errorf("bitwise XOR requires integers")
		}
		return values.NewInt(left.Data.(int64) ^ right.Data.(int64)), nil
	case "<<":
		if !left.IsInt() || !right.IsInt() {
			return nil, fmt.Errorf("left shift requires integers")
		}
		return values.NewInt(left.Data.(int64) << uint(right.Data.(int64))), nil
	case ">>":
		if !left.IsInt() || !right.IsInt() {
			return nil, fmt.Errorf("right shift requires integers")
		}
		return values.NewInt(left.Data.(int64) >> uint(right.Data.(int64))), nil

	default:
		return nil, fmt.Errorf("unsupported binary operator: %s", operator)
	}
}

// evaluateUnaryOperation performs unary operations on constant values
func (c *Compiler) evaluateUnaryOperation(operator string, operand *values.Value) (*values.Value, error) {
	switch operator {
	case "+":
		if operand.IsInt() {
			return values.NewInt(operand.Data.(int64)), nil
		}
		if operand.IsFloat() {
			return values.NewFloat(operand.Data.(float64)), nil
		}
		return nil, fmt.Errorf("unary plus requires numeric operand")
	case "-":
		if operand.IsInt() {
			return values.NewInt(-operand.Data.(int64)), nil
		}
		if operand.IsFloat() {
			return values.NewFloat(-operand.Data.(float64)), nil
		}
		return nil, fmt.Errorf("unary minus requires numeric operand")
	case "!":
		return values.NewBool(!c.isTruthy(operand)), nil
	case "~":
		if !operand.IsInt() {
			return nil, fmt.Errorf("bitwise NOT requires integer operand")
		}
		return values.NewInt(^operand.Data.(int64)), nil
	default:
		return nil, fmt.Errorf("unsupported unary operator: %s", operator)
	}
}

// isTruthy determines the truth value of a constant expression (PHP truthiness rules)
func (c *Compiler) isTruthy(value *values.Value) bool {
	switch value.Type {
	case values.TypeBool:
		return value.Data.(bool)
	case values.TypeInt:
		return value.Data.(int64) != 0
	case values.TypeFloat:
		return value.Data.(float64) != 0.0
	case values.TypeString:
		str := value.Data.(string)
		return str != "" && str != "0"
	case values.TypeNull:
		return false
	case values.TypeArray:
		arr := value.Data.(*values.Array)
		return len(arr.Elements) > 0
	default:
		return true // Objects and other types are truthy
	}
}

// evaluateClassConstantAccess handles self::CONST, parent::CONST, ClassName::CONST
func (c *Compiler) evaluateClassConstantAccess(expr *ast.ClassConstantAccessExpression) (*values.Value, error) {
	// Get the constant name
	constName, ok := expr.Constant.(*ast.IdentifierNode)
	if !ok {
		return nil, fmt.Errorf("constant name must be an identifier")
	}

	// Determine which class to look in
	var targetClass *vm.Class

	if classExpr, ok := expr.Class.(*ast.IdentifierNode); ok {
		switch classExpr.Name {
		case "self":
			if c.currentClass == nil {
				return nil, fmt.Errorf("cannot use self:: outside of class context")
			}
			targetClass = c.currentClass
		case "parent":
			if c.currentClass == nil {
				return nil, fmt.Errorf("cannot use parent:: outside of class context")
			}
			// For now, we'll return an error if trying to access parent constants
			// In a full implementation, we'd need to resolve the parent class
			return nil, fmt.Errorf("parent:: constant access not yet implemented in constant expressions")
		default:
			// Named class constant access
			className := classExpr.Name
			// Look up the class - for now we'll only support constants from the current class
			if c.currentClass != nil && c.currentClass.Name == className {
				targetClass = c.currentClass
			} else {
				return nil, fmt.Errorf("class constant access to external class %s not yet implemented in constant expressions", className)
			}
		}
	} else {
		return nil, fmt.Errorf("unsupported class expression in constant access: %T", expr.Class)
	}

	// Look up the constant in the target class
	if constant, found := targetClass.Constants[constName.Name]; found {
		return constant.Value, nil
	}

	return nil, fmt.Errorf("undefined class constant %s::%s", targetClass.Name, constName.Name)
}

// Helper functions for arithmetic operations and type conversions

// performArithmetic performs arithmetic operations with automatic type promotion
func (c *Compiler) performArithmetic(left, right *values.Value, intOp func(int64, int64) int64, floatOp func(float64, float64) float64) (*values.Value, error) {
	if left.IsInt() && right.IsInt() {
		return values.NewInt(intOp(left.Data.(int64), right.Data.(int64))), nil
	}
	// At least one operand is float, promote both to float
	leftFloat := c.toFloat(left)
	rightFloat := c.toFloat(right)
	return values.NewFloat(floatOp(leftFloat, rightFloat)), nil
}

// toFloat converts a value to float64
func (c *Compiler) toFloat(value *values.Value) float64 {
	if value.IsInt() {
		return float64(value.Data.(int64))
	}
	if value.IsFloat() {
		return value.Data.(float64)
	}
	if value.IsString() {
		if f, err := strconv.ParseFloat(value.Data.(string), 64); err == nil {
			return f
		}
	}
	return 0.0
}

// toString converts a value to string
func (c *Compiler) toString(value *values.Value) string {
	switch value.Type {
	case values.TypeString:
		return value.Data.(string)
	case values.TypeInt:
		return strconv.FormatInt(value.Data.(int64), 10)
	case values.TypeFloat:
		return strconv.FormatFloat(value.Data.(float64), 'G', -1, 64)
	case values.TypeBool:
		if value.Data.(bool) {
			return "1"
		}
		return ""
	case values.TypeNull:
		return ""
	default:
		return ""
	}
}

// isEqual checks if two values are equal using PHP comparison rules
func (c *Compiler) isEqual(left, right *values.Value) bool {
	if left.Type == right.Type {
		switch left.Type {
		case values.TypeBool:
			return left.Data.(bool) == right.Data.(bool)
		case values.TypeInt:
			return left.Data.(int64) == right.Data.(int64)
		case values.TypeFloat:
			return left.Data.(float64) == right.Data.(float64)
		case values.TypeString:
			return left.Data.(string) == right.Data.(string)
		case values.TypeNull:
			return true // null == null
		}
	}

	// Type juggling for mixed comparisons
	if (left.IsInt() || left.IsFloat()) && (right.IsInt() || right.IsFloat()) {
		return c.toFloat(left) == c.toFloat(right)
	}

	// String to number comparisons
	if left.IsString() && (right.IsInt() || right.IsFloat()) {
		if f, err := strconv.ParseFloat(left.Data.(string), 64); err == nil {
			return f == c.toFloat(right)
		}
	}
	if right.IsString() && (left.IsInt() || left.IsFloat()) {
		if f, err := strconv.ParseFloat(right.Data.(string), 64); err == nil {
			return c.toFloat(left) == f
		}
	}

	return false
}

// compare compares two values and returns -1, 0, or 1
func (c *Compiler) compare(left, right *values.Value) (int, error) {
	// Same type comparisons
	if left.Type == right.Type {
		switch left.Type {
		case values.TypeInt:
			leftVal := left.Data.(int64)
			rightVal := right.Data.(int64)
			if leftVal < rightVal {
				return -1, nil
			} else if leftVal > rightVal {
				return 1, nil
			}
			return 0, nil
		case values.TypeFloat:
			leftVal := left.Data.(float64)
			rightVal := right.Data.(float64)
			if leftVal < rightVal {
				return -1, nil
			} else if leftVal > rightVal {
				return 1, nil
			}
			return 0, nil
		case values.TypeString:
			leftVal := left.Data.(string)
			rightVal := right.Data.(string)
			if leftVal < rightVal {
				return -1, nil
			} else if leftVal > rightVal {
				return 1, nil
			}
			return 0, nil
		case values.TypeBool:
			leftVal := left.Data.(bool)
			rightVal := right.Data.(bool)
			if !leftVal && rightVal {
				return -1, nil
			} else if leftVal && !rightVal {
				return 1, nil
			}
			return 0, nil
		}
	}

	// Numeric comparisons with type promotion
	if (left.IsInt() || left.IsFloat()) && (right.IsInt() || right.IsFloat()) {
		leftFloat := c.toFloat(left)
		rightFloat := c.toFloat(right)
		if leftFloat < rightFloat {
			return -1, nil
		} else if leftFloat > rightFloat {
			return 1, nil
		}
		return 0, nil
	}

	// For other mixed type comparisons, we could implement more PHP rules
	// but for now, we'll just return equal
	return 0, nil
}

// validateConstantType validates that a constant value matches its declared type
func (c *Compiler) validateConstantType(value *values.Value, typeHint string) error {
	if typeHint == "" {
		return nil // No type constraint
	}

	switch typeHint {
	case "string":
		if !value.IsString() {
			return fmt.Errorf("expected string, got %s", value.Type.String())
		}
	case "int":
		if !value.IsInt() {
			return fmt.Errorf("expected int, got %s", value.Type.String())
		}
	case "float":
		if !value.IsFloat() {
			return fmt.Errorf("expected float, got %s", value.Type.String())
		}
	case "bool":
		if !value.IsBool() {
			return fmt.Errorf("expected bool, got %s", value.Type.String())
		}
	case "array":
		if !value.IsArray() {
			return fmt.Errorf("expected array, got %s", value.Type.String())
		}
	case "null":
		if !value.IsNull() {
			return fmt.Errorf("expected null, got %s", value.Type.String())
		}
	default:
		// For other types (classes, mixed, etc.), we might need more complex validation
		return fmt.Errorf("unsupported type hint: %s", typeHint)
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
		Name:       className,
		Parent:     "",
		Properties: make(map[string]*vm.Property),
		Methods:    make(map[string]*vm.Function),
		Constants:  make(map[string]*vm.ClassConstant),
		IsAbstract: decl.Abstract,
		IsFinal:    decl.Final,
	}

	// Handle extends
	if decl.Extends != nil {
		if parent, ok := decl.Extends.(*ast.IdentifierNode); ok {
			class.Parent = parent.Name
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

	// Set current class context in VM
	c.emit(opcodes.OP_SET_CURRENT_CLASS, opcodes.IS_CONST, nameConstant, 0, 0, 0, 0)

	// Set parent class if exists
	if class.Parent != "" {
		parentConstant := c.addConstant(values.NewString(class.Parent))
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

	// Clear current class context in VM
	c.emit(opcodes.OP_CLEAR_CURRENT_CLASS, 0, 0, 0, 0, 0, 0)

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
		// But regular variables like $obj should be compiled dynamically
		className := class.Name
		if className == "self" || className == "static" || className == "parent" {
			classOperand = c.addConstant(values.NewString(className))
			classOperandType = opcodes.IS_CONST
		} else {
			// Regular variable - compile it dynamically
			err := c.compileNode(class)
			if err != nil {
				return fmt.Errorf("failed to compile class expression: %w", err)
			}
			classOperand = c.nextTemp - 1 // Last allocated temp contains the class name
			classOperandType = opcodes.IS_TMP_VAR
		}
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

// compileIncludeOrEval compiles include/require/eval expressions
func (c *Compiler) compileIncludeOrEval(expr *ast.IncludeOrEvalExpression) error {
	// Compile the argument expression (file path for include/require, code for eval)
	if err := c.compileNode(expr.Expr); err != nil {
		return err
	}

	// Get the result of the argument expression
	argOperand := c.nextTemp - 1

	// Allocate temporary for the result
	result := c.allocateTemp()

	// Emit the appropriate opcode based on the token type
	var opcode opcodes.Opcode
	switch expr.Type {
	case lexer.T_INCLUDE:
		opcode = opcodes.OP_INCLUDE
	case lexer.T_INCLUDE_ONCE:
		opcode = opcodes.OP_INCLUDE_ONCE
	case lexer.T_REQUIRE:
		opcode = opcodes.OP_REQUIRE
	case lexer.T_REQUIRE_ONCE:
		opcode = opcodes.OP_REQUIRE_ONCE
	case lexer.T_EVAL:
		// eval is not yet implemented, but we can add the skeleton
		return fmt.Errorf("eval() is not yet supported")
	default:
		return fmt.Errorf("unsupported include/require/eval type: %v", expr.Type)
	}

	// Emit the instruction: OPCODE result, arg
	c.emit(opcode,
		opcodes.IS_TMP_VAR, argOperand,
		opcodes.IS_UNUSED, 0,
		opcodes.IS_TMP_VAR, result)

	return nil
}

// compileArrayElement compiles array element expressions
// Note: ArrayElementExpression is typically handled within array compilation
func (c *Compiler) compileArrayElement(expr *ast.ArrayElementExpression) error {
	// ArrayElementExpression represents key => value pairs in arrays
	// In normal PHP usage, these are processed as part of array expressions
	// However, we'll provide a standalone compilation for completeness

	// If there's a key, compile it first
	if expr.Key != nil {
		if err := c.compileNode(expr.Key); err != nil {
			return err
		}
	}

	// Compile the value
	if expr.Value != nil {
		if err := c.compileNode(expr.Value); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("array element must have a value")
	}

	// The result is the value expression's result
	// (The key is not returned as a standalone value)
	return nil
}

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
	for _, staticVar := range stmt.Variables {
		// Get variable name
		variable, ok := staticVar.Variable.(*ast.Variable)
		if !ok {
			return fmt.Errorf("invalid static variable: expected variable")
		}

		varName := variable.Name

		// Allocate variable slot for this static variable
		varSlot := c.getOrCreateVariable(varName)

		// Compile default value if provided and emit BIND_STATIC instruction
		if staticVar.DefaultValue != nil {
			err := c.compileNode(staticVar.DefaultValue)
			if err != nil {
				return fmt.Errorf("failed to compile static variable default value: %w", err)
			}
			// The result is in the last allocated temp
			defaultSlot := c.nextTemp - 1

			// Emit BIND_STATIC with default value
			c.emit(opcodes.OP_BIND_STATIC,
				opcodes.IS_CV, varSlot, // Variable to bind
				opcodes.IS_CONST, c.addConstant(values.NewString(varName)), // Variable name
				opcodes.IS_TMP_VAR, defaultSlot) // Default value
		} else {
			// Emit BIND_STATIC without default value (will use null)
			c.emit(opcodes.OP_BIND_STATIC,
				opcodes.IS_CV, varSlot, // Variable to bind
				opcodes.IS_CONST, c.addConstant(values.NewString(varName)), // Variable name
				opcodes.IS_UNUSED, 0) // No default value
		}
	}

	return nil
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
					DefaultValue: nil,
				}

				// Handle default value evaluation
				if param.DefaultValue != nil {
					// Evaluate the default value expression at compile time
					defaultValue := c.evaluateConstantExpression(param.DefaultValue)
					if defaultValue != nil {
						vmParam.DefaultValue = defaultValue
					} else {
						// If we can't evaluate it at compile time, use null as fallback
						vmParam.DefaultValue = values.NewNull()
					}
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
		Name:       enumName,
		Parent:     "",
		Properties: make(map[string]*vm.Property),
		Methods:    make(map[string]*vm.Function),
		Constants:  make(map[string]*vm.ClassConstant),
		IsAbstract: false,
		IsFinal:    true, // Enums are final by default
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

		enumClass.Constants[caseName] = &vm.ClassConstant{
			Name:       caseName,
			Value:      caseValue,
			Visibility: "public",
			IsFinal:    true,
		}
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

			// Also register in unified registry if available
			if registry.GlobalRegistry != nil {
				// Get or create class in registry
				classDesc, err := registry.GlobalRegistry.GetClass(c.currentClass.Name)
				if err != nil {
					// Class doesn't exist in registry, create it
					classDesc = &registry.ClassDescriptor{
						Name:       c.currentClass.Name,
						Parent:     c.currentClass.Parent,
						IsAbstract: c.currentClass.IsAbstract,
						IsFinal:    c.currentClass.IsFinal,
						Properties: make(map[string]*registry.PropertyDescriptor),
						Methods:    make(map[string]*registry.MethodDescriptor),
						Constants:  make(map[string]*registry.ConstantDescriptor),
					}
					registry.GlobalRegistry.RegisterClass(classDesc)
				}

				// Convert VM parameters to registry parameters
				registryParams := make([]registry.ParameterDescriptor, len(classMethod.Parameters))
				for i, param := range classMethod.Parameters {
					registryParams[i] = registry.ParameterDescriptor{
						Name:         param.Name,
						Type:         param.Type,
						IsReference:  param.IsReference,
						HasDefault:   param.HasDefault,
						DefaultValue: param.DefaultValue,
					}
				}

				// Convert VM parameters to registry parameter info for default value support
				paramInfo := make([]registry.ParameterInfo, len(classMethod.Parameters))
				for i, param := range classMethod.Parameters {
					paramInfo[i] = registry.ParameterInfo{
						Name:         param.Name,
						HasDefault:   param.HasDefault,
						DefaultValue: param.DefaultValue,
						IsVariadic:   false, // Individual parameter is not variadic
					}
				}

				// Mark the last parameter as variadic if the method is variadic
				if len(paramInfo) > 0 && classMethod.IsVariadic {
					paramInfo[len(paramInfo)-1].IsVariadic = true
				}

				// Use the actual trait method bytecode instead of hardcoded implementations
				methodImpl := &registry.BytecodeMethodImpl{
					Instructions: classMethod.Instructions,
					Constants:    classMethod.Constants,
					LocalVars:    len(classMethod.Parameters), // Approximate local variable count
					Parameters:   paramInfo,
				}

				// Create method descriptor
				methodDesc := &registry.MethodDescriptor{
					Name:           methodName,
					Visibility:     "public", // Traits methods are typically public
					IsStatic:       false,    // Trait methods are typically not static
					IsAbstract:     false,
					IsFinal:        false,
					Parameters:     registryParams,
					Implementation: methodImpl,
					IsVariadic:     classMethod.IsVariadic,
				}

				// Register the trait method in the registry
				classDesc.Methods[methodName] = methodDesc
			}
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
		if e.Kind == ast.IntegerKind {
			// Use pre-converted integer value
			return values.NewInt(e.IntValue)
		} else if e.Kind == ast.FloatKind {
			// Use pre-converted float value
			return values.NewFloat(e.FloatValue)
		}
		// Fallback to parsing the string value
		if intVal, err := strconv.ParseInt(e.Value, 10, 64); err == nil {
			return values.NewInt(intVal)
		}
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

// evaluateConstantArrayExpression evaluates a constant array expression at compile time
// This function only accepts literal values, no expressions or function calls
func (c *Compiler) evaluateConstantArrayExpression(arrExpr *ast.ArrayExpression) (*values.Value, error) {
	// Create a new array value
	arrayValue := values.NewArray()
	arrayData := arrayValue.Data.(*values.Array)

	// Process each element
	for _, element := range arrExpr.Elements {
		switch elem := element.(type) {
		case *ast.ArrayElementExpression:
			// Handle key => value pairs and simple values
			var keyValue interface{}
			var valueValue *values.Value

			// Handle the key if present
			if elem.Key != nil {
				keyConst := c.evaluateConstantExpression(elem.Key)
				if keyConst == nil {
					return nil, fmt.Errorf("array keys must be constant expressions")
				}
				// Convert to appropriate Go type for map key
				switch keyConst.Type {
				case values.TypeString:
					keyValue = keyConst.Data.(string)
				case values.TypeInt:
					keyValue = keyConst.Data.(int64)
				case values.TypeFloat:
					// PHP converts float keys to integers
					keyValue = int64(keyConst.Data.(float64))
				case values.TypeBool:
					// PHP converts bool keys to integers
					if keyConst.Data.(bool) {
						keyValue = int64(1)
					} else {
						keyValue = int64(0)
					}
				default:
					return nil, fmt.Errorf("invalid array key type: %v", keyConst.Type)
				}
			}

			// Handle the value
			if elem.Value == nil {
				return nil, fmt.Errorf("array element must have a value")
			}

			// Check if the value is another array (nested array)
			if nestedArrayExpr, ok := elem.Value.(*ast.ArrayExpression); ok {
				// Recursively evaluate nested array
				nestedArrayValue, err := c.evaluateConstantArrayExpression(nestedArrayExpr)
				if err != nil {
					return nil, fmt.Errorf("error in nested array: %v", err)
				}
				valueValue = nestedArrayValue
			} else {
				// Evaluate as constant expression
				valueValue = c.evaluateConstantExpression(elem.Value)
				if valueValue == nil {
					return nil, fmt.Errorf("array values must be constant expressions")
				}
			}

			// Add to array
			if keyValue != nil {
				// Keyed element
				arrayData.Elements[keyValue] = valueValue
			} else {
				// Auto-indexed element - find next integer key
				nextIndex := int64(len(arrayData.Elements))
				for {
					if _, exists := arrayData.Elements[nextIndex]; !exists {
						break
					}
					nextIndex++
				}
				arrayData.Elements[nextIndex] = valueValue
			}

		default:
			// Direct element (not wrapped in ArrayElementExpression)
			var valueValue *values.Value

			// Check if it's a nested array
			if nestedArrayExpr, ok := elem.(*ast.ArrayExpression); ok {
				// Recursively evaluate nested array
				nestedArrayValue, err := c.evaluateConstantArrayExpression(nestedArrayExpr)
				if err != nil {
					return nil, fmt.Errorf("error in nested array: %v", err)
				}
				valueValue = nestedArrayValue
			} else {
				// Evaluate as constant expression
				valueValue = c.evaluateConstantExpression(elem)
				if valueValue == nil {
					return nil, fmt.Errorf("array values must be constant expressions")
				}
			}

			// Auto-indexed element
			nextIndex := int64(len(arrayData.Elements))
			for {
				if _, exists := arrayData.Elements[nextIndex]; !exists {
					break
				}
				nextIndex++
			}
			arrayData.Elements[nextIndex] = valueValue
		}
	}

	return arrayValue, nil
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
