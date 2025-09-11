package compiler

import (
	"fmt"
	
	"github.com/wudi/hey/compiler/ast"
	"github.com/wudi/hey/compiler/opcodes"
	"github.com/wudi/hey/compiler/values"
)

// ContextCompiler is the new context-based compiler that removes VM dependencies
type ContextCompilerFixed struct {
	// The compiler itself is stateless - all state is in the context
}

// NewContextCompilerFixed creates a new context-based compiler
func NewContextCompilerFixed() *ContextCompilerFixed {
	return &ContextCompilerFixed{}
}

// Compile compiles an AST node using the provided context
func (c *ContextCompilerFixed) Compile(ctx *CompileContext, node ast.Node) error {
	if ctx == nil {
		return fmt.Errorf("compilation context cannot be nil")
	}
	
	if node == nil {
		return nil
	}
	
	// Compile the node using the context
	err := c.compileNode(ctx, node)
	if err != nil {
		return err
	}
	
	// Add final return if needed (only for global scope)
	if ctx.IsGlobalScope() {
		if len(ctx.Instructions) == 0 || ctx.Instructions[len(ctx.Instructions)-1].Opcode != opcodes.OP_RETURN {
			nullConstant := ctx.AddConstant(values.NewNull())
			ctx.EmitInstruction(opcodes.OP_RETURN, opcodes.IS_CONST, nullConstant, opcodes.IS_UNUSED, 0, opcodes.IS_UNUSED, 0)
		}
	}
	
	return nil
}

// compileNode is the main dispatch method for compiling different AST node types
func (c *ContextCompilerFixed) compileNode(ctx *CompileContext, node ast.Node) error {
	if node == nil {
		return nil
	}
	
	switch n := node.(type) {
	// Expressions
	case *ast.BinaryExpression:
		return c.compileBinaryOp(ctx, n)
	case *ast.UnaryExpression:
		return c.compileUnaryOp(ctx, n)
	case *ast.AssignmentExpression:
		return c.compileAssignment(ctx, n)
	case *ast.IdentifierNode:
		return c.compileIdentifier(ctx, n)
	case *ast.NumberLiteral:
		return c.compileNumber(ctx, n)
	case *ast.StringLiteral:
		return c.compileString(ctx, n)
	case *ast.BooleanLiteral:
		return c.compileBool(ctx, n)
	case *ast.NullLiteral:
		return c.compileNull(ctx, n)
	case *ast.ArrayExpression:
		return c.compileArray(ctx, n)
	case *ast.PropertyAccessExpression:
		return c.compilePropertyAccess(ctx, n)
	case *ast.MethodCallExpression:
		return c.compileMethodCall(ctx, n)
	case *ast.CallExpression:
		return c.compileFunctionCall(ctx, n)
	case *ast.AnonymousFunctionExpression:
		return c.compileAnonymousFunction(ctx, n)
	case *ast.Variable:
		return c.compileVariable(ctx, n)
		
	// Statements
	case *ast.ExpressionStatement:
		return c.compileExpressionStatement(ctx, n)
	case *ast.IfStatement:
		return c.compileIf(ctx, n)
	case *ast.WhileStatement:
		return c.compileWhile(ctx, n)
	case *ast.ForStatement:
		return c.compileForLoop(ctx, n)
	case *ast.ForeachStatement:
		return c.compileForeachLoop(ctx, n)
	case *ast.SwitchStatement:
		return c.compileSwitch(ctx, n)
	case *ast.BreakStatement:
		return c.compileBreak(ctx, n)
	case *ast.ContinueStatement:
		return c.compileContinue(ctx, n)
	case *ast.ReturnStatement:
		return c.compileReturn(ctx, n)
	case *ast.BlockStatement:
		return c.compileBlock(ctx, n)
	case *ast.EchoStatement:
		return c.compileEcho(ctx, n)
	case *ast.PrintStatement:
		return c.compilePrint(ctx, n)
		
	// Declarations
	case *ast.FunctionDeclaration:
		return c.compileFunctionDeclaration(ctx, n)
	case *ast.ClassExpression:
		return c.compileClassDeclaration(ctx, n)
	case *ast.InterfaceDeclaration:
		return c.compileInterfaceDeclaration(ctx, n)
	case *ast.TraitDeclaration:
		return c.compileTraitDeclaration(ctx, n)
	case *ast.PropertyDeclaration:
		return c.compilePropertyDeclaration(ctx, n)
		
	// Handle Program nodes
	case *ast.Program:
		return c.compileProgram(ctx, n)
	
	// Handle statements that implement Statement interface
	case ast.Statement:
		// Try to handle as generic statement
		return fmt.Errorf("unhandled statement type: %T", n)
	
	// Handle expressions that implement Expression interface
	case ast.Expression:
		return fmt.Errorf("unhandled expression type: %T", n)
	
	default:
		return fmt.Errorf("unsupported node type: %T", n)
	}
}

// Basic compilation methods for literals and identifiers

func (c *ContextCompilerFixed) compileIdentifier(ctx *CompileContext, node *ast.IdentifierNode) error {
	varName := node.Name
	
	// Try to find variable in context chain
	if slot, exists := ctx.GetVariable(varName); exists {
		// Emit fetch variable instruction
		resultVar := ctx.GetNextTemp()
		ctx.EmitInstruction(opcodes.OP_FETCH_R, opcodes.IS_VAR, slot, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, resultVar)
		return nil
	}
	
	// Variable not found - could be undefined or a function name
	return fmt.Errorf("undefined variable: %s", varName)
}

func (c *ContextCompilerFixed) compileNumber(ctx *CompileContext, node *ast.NumberLiteral) error {
	var constantIndex uint32
	if node.Kind == ast.FloatKind {
		constantIndex = ctx.AddConstant(values.NewFloat(node.FloatValue))
	} else {
		constantIndex = ctx.AddConstant(values.NewInt(node.IntValue))
	}
	resultVar := ctx.GetNextTemp()
	ctx.EmitInstruction(opcodes.OP_FETCH_CONSTANT, opcodes.IS_CONST, constantIndex, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, resultVar)
	return nil
}

func (c *ContextCompilerFixed) compileString(ctx *CompileContext, node *ast.StringLiteral) error {
	constantIndex := ctx.AddConstant(values.NewString(node.Value))
	resultVar := ctx.GetNextTemp()
	ctx.EmitInstruction(opcodes.OP_FETCH_CONSTANT, opcodes.IS_CONST, constantIndex, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, resultVar)
	return nil
}

func (c *ContextCompilerFixed) compileBool(ctx *CompileContext, node *ast.BooleanLiteral) error {
	constantIndex := ctx.AddConstant(values.NewBool(node.Value))
	resultVar := ctx.GetNextTemp()
	ctx.EmitInstruction(opcodes.OP_FETCH_CONSTANT, opcodes.IS_CONST, constantIndex, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, resultVar)
	return nil
}

func (c *ContextCompilerFixed) compileNull(ctx *CompileContext, node *ast.NullLiteral) error {
	constantIndex := ctx.AddConstant(values.NewNull())
	resultVar := ctx.GetNextTemp()
	ctx.EmitInstruction(opcodes.OP_FETCH_CONSTANT, opcodes.IS_CONST, constantIndex, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, resultVar)
	return nil
}

func (c *ContextCompilerFixed) compileVariable(ctx *CompileContext, node *ast.Variable) error {
	if node.Name == "" {
		return fmt.Errorf("variable must have a name")
	}
	
	varName := node.Name
	
	// Try to find variable in context chain
	if slot, exists := ctx.GetVariable(varName); exists {
		resultVar := ctx.GetNextTemp()
		ctx.EmitInstruction(opcodes.OP_FETCH_R, opcodes.IS_VAR, slot, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, resultVar)
		return nil
	}
	
	// Variable not found
	return fmt.Errorf("undefined variable: %s", varName)
}

// Assignment operations

func (c *ContextCompilerFixed) compileAssignment(ctx *CompileContext, node *ast.AssignmentExpression) error {
	// Compile the right-hand side expression first
	err := c.compileNode(ctx, node.Right)
	if err != nil {
		return err
	}
	
	// Handle left-hand side (currently only simple variable assignment)
	if identifierNode, ok := node.Left.(*ast.IdentifierNode); ok {
		varName := identifierNode.Name
		slot := ctx.GetOrCreateVariable(varName)
		
		// Emit assignment instruction
		ctx.EmitInstruction(opcodes.OP_ASSIGN, opcodes.IS_TMP_VAR, ctx.GetNextTemp()-1, opcodes.IS_UNUSED, 0, opcodes.IS_VAR, slot)
		return nil
	}
	
	return fmt.Errorf("unsupported assignment target: %T", node.Left)
}

// Binary and unary operations

func (c *ContextCompilerFixed) compileBinaryOp(ctx *CompileContext, node *ast.BinaryExpression) error {
	// Compile left operand
	err := c.compileNode(ctx, node.Left)
	if err != nil {
		return err
	}
	
	// Compile right operand
	err = c.compileNode(ctx, node.Right)
	if err != nil {
		return err
	}
	
	// Emit the appropriate binary operation instruction
	var opcode opcodes.Opcode
	switch node.Operator {
	case "+":
		opcode = opcodes.OP_ADD
	case "-":
		opcode = opcodes.OP_SUB
	case "*":
		opcode = opcodes.OP_MUL
	case "/":
		opcode = opcodes.OP_DIV
	case "%":
		opcode = opcodes.OP_MOD
	case "==":
		opcode = opcodes.OP_IS_EQUAL
	case "!=":
		opcode = opcodes.OP_IS_NOT_EQUAL
	case "<":
		opcode = opcodes.OP_IS_SMALLER
	case "<=":
		opcode = opcodes.OP_IS_SMALLER_OR_EQUAL
	case ">":
		opcode = opcodes.OP_IS_GREATER
	case ">=":
		opcode = opcodes.OP_IS_GREATER_OR_EQUAL
	case "&&":
		opcode = opcodes.OP_BOOLEAN_AND
	case "||":
		opcode = opcodes.OP_BOOLEAN_OR
	case ".":
		opcode = opcodes.OP_CONCAT
	default:
		return fmt.Errorf("unsupported binary operator: %s", node.Operator)
	}
	
	resultVar := ctx.GetNextTemp()
	ctx.EmitInstruction(opcode, opcodes.IS_UNUSED, 0, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, resultVar)
	return nil
}

func (c *ContextCompilerFixed) compileUnaryOp(ctx *CompileContext, node *ast.UnaryExpression) error {
	// Compile operand first
	err := c.compileNode(ctx, node.Operand)
	if err != nil {
		return err
	}
	
	// Emit the appropriate unary operation instruction
	var opcode opcodes.Opcode
	switch node.Operator {
	case "-":
		opcode = opcodes.OP_MINUS
	case "+":
		// Unary plus is essentially a no-op
		return nil
	case "!":
		opcode = opcodes.OP_NOT
	default:
		return fmt.Errorf("unsupported unary operator: %s", node.Operator)
	}
	
	resultVar := ctx.GetNextTemp()
	ctx.EmitInstruction(opcode, opcodes.IS_UNUSED, 0, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, resultVar)
	return nil
}

// Program and statement compilation

func (c *ContextCompilerFixed) compileProgram(ctx *CompileContext, node *ast.Program) error {
	for _, stmt := range node.Body {
		if err := c.compileNode(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func (c *ContextCompilerFixed) compileExpressionStatement(ctx *CompileContext, node *ast.ExpressionStatement) error {
	return c.compileNode(ctx, node.Expression)
}

func (c *ContextCompilerFixed) compileBlock(ctx *CompileContext, node *ast.BlockStatement) error {
	blockCtx := ctx.NewChildContext(ScopeBlock)
	
	if node.Body != nil {
		for _, stmt := range node.Body {
			if err := c.compileNode(blockCtx, stmt); err != nil {
				return err
			}
		}
	}
	
	// Merge block instructions back to parent context
	ctx.Instructions = append(ctx.Instructions, blockCtx.Instructions...)
	return nil
}

func (c *ContextCompilerFixed) compileEcho(ctx *CompileContext, node *ast.EchoStatement) error {
	if node.Arguments != nil {
		for _, arg := range node.Arguments.Arguments {
			if err := c.compileNode(ctx, arg); err != nil {
				return err
			}
			ctx.EmitInstruction(opcodes.OP_ECHO, opcodes.IS_TMP_VAR, ctx.GetNextTemp()-1, opcodes.IS_UNUSED, 0, opcodes.IS_UNUSED, 0)
		}
	}
	return nil
}

func (c *ContextCompilerFixed) compilePrint(ctx *CompileContext, node *ast.PrintStatement) error {
	if node.Arguments != nil {
		for _, arg := range node.Arguments.Arguments {
			if err := c.compileNode(ctx, arg); err != nil {
				return err
			}
			ctx.EmitInstruction(opcodes.OP_PRINT, opcodes.IS_TMP_VAR, ctx.GetNextTemp()-1, opcodes.IS_UNUSED, 0, opcodes.IS_UNUSED, 0)
		}
	}
	return nil
}

// Control flow statements

func (c *ContextCompilerFixed) compileIf(ctx *CompileContext, node *ast.IfStatement) error {
	// Compile condition
	if err := c.compileNode(ctx, node.Test); err != nil {
		return err
	}
	conditionVar := ctx.GetNextTemp() - 1
	
	// Create labels
	elseLabel := ctx.GetNextLabel()
	endLabel := ctx.GetNextLabel()
	
	// Jump to else if condition is false
	ctx.EmitInstruction(opcodes.OP_JMPZ, opcodes.IS_TMP_VAR, conditionVar, opcodes.IS_CONST, 0, opcodes.IS_UNUSED, 0)
	ctx.AddForwardJump(elseLabel, len(ctx.Instructions)-1, opcodes.OP_JMPZ, 2)
	
	// Compile then block
	if node.Consequent != nil {
		for _, stmt := range node.Consequent {
			if err := c.compileNode(ctx, stmt); err != nil {
				return err
			}
		}
	}
	
	// Jump to end
	ctx.EmitInstruction(opcodes.OP_JMP, opcodes.IS_CONST, 0, opcodes.IS_UNUSED, 0, opcodes.IS_UNUSED, 0)
	ctx.AddForwardJump(endLabel, len(ctx.Instructions)-1, opcodes.OP_JMP, 1)
	
	// Place else label
	ctx.PlaceLabel(elseLabel)
	
	// Compile alternate block if present  
	if node.Alternate != nil {
		for _, stmt := range node.Alternate {
			if err := c.compileNode(ctx, stmt); err != nil {
				return err
			}
		}
	}
	
	// Place end label
	ctx.PlaceLabel(endLabel)
	
	// Resolve forward jumps
	c.resolveForwardJumps(ctx, elseLabel)
	c.resolveForwardJumps(ctx, endLabel)
	
	return nil
}

func (c *ContextCompilerFixed) compileWhile(ctx *CompileContext, node *ast.WhileStatement) error {
	startLabel := ctx.GetNextLabel()
	endLabel := ctx.GetNextLabel()
	
	oldBreakLabel := ctx.BreakLabel
	oldContinueLabel := ctx.ContinueLabel
	ctx.BreakLabel = endLabel
	ctx.ContinueLabel = startLabel
	
	ctx.PlaceLabel(startLabel)
	
	if err := c.compileNode(ctx, node.Test); err != nil {
		return err
	}
	conditionVar := ctx.GetNextTemp() - 1
	
	ctx.EmitInstruction(opcodes.OP_JMPZ, opcodes.IS_TMP_VAR, conditionVar, opcodes.IS_CONST, 0, opcodes.IS_UNUSED, 0)
	ctx.AddForwardJump(endLabel, len(ctx.Instructions)-1, opcodes.OP_JMPZ, 2)
	
	if node.Body != nil {
		for _, stmt := range node.Body {
			if err := c.compileNode(ctx, stmt); err != nil {
				return err
			}
		}
	}
	
	startPos, _ := ctx.GetLabelPosition(startLabel)
	ctx.EmitInstruction(opcodes.OP_JMP, opcodes.IS_CONST, uint32(startPos), opcodes.IS_UNUSED, 0, opcodes.IS_UNUSED, 0)
	
	ctx.PlaceLabel(endLabel)
	c.resolveForwardJumps(ctx, endLabel)
	
	ctx.BreakLabel = oldBreakLabel
	ctx.ContinueLabel = oldContinueLabel
	
	return nil
}

func (c *ContextCompilerFixed) compileForLoop(ctx *CompileContext, node *ast.ForStatement) error {
	forCtx := ctx.NewChildContext(ScopeBlock)
	
	// Compile initialization - single expression
	if node.Init != nil {
		if err := c.compileNode(forCtx, node.Init); err != nil {
			return err
		}
	}
	
	startLabel := forCtx.GetNextLabel()
	continueLabel := forCtx.GetNextLabel()
	endLabel := forCtx.GetNextLabel()
	
	oldBreakLabel := forCtx.BreakLabel
	oldContinueLabel := forCtx.ContinueLabel
	forCtx.BreakLabel = endLabel
	forCtx.ContinueLabel = continueLabel
	
	forCtx.PlaceLabel(startLabel)
	
	// Compile condition - single expression
	if node.Test != nil {
		if err := c.compileNode(forCtx, node.Test); err != nil {
			return err
		}
		conditionVar := forCtx.GetNextTemp() - 1
		
		forCtx.EmitInstruction(opcodes.OP_JMPZ, opcodes.IS_TMP_VAR, conditionVar, opcodes.IS_CONST, 0, opcodes.IS_UNUSED, 0)
		forCtx.AddForwardJump(endLabel, len(forCtx.Instructions)-1, opcodes.OP_JMPZ, 2)
	}
	
	// Compile body
	if node.Body != nil {
		for _, stmt := range node.Body {
			if err := c.compileNode(forCtx, stmt); err != nil {
				return err
			}
		}
	}
	
	forCtx.PlaceLabel(continueLabel)
	
	// Compile update - single expression
	if node.Update != nil {
		if err := c.compileNode(forCtx, node.Update); err != nil {
			return err
		}
	}
	
	startPos, _ := forCtx.GetLabelPosition(startLabel)
	forCtx.EmitInstruction(opcodes.OP_JMP, opcodes.IS_CONST, uint32(startPos), opcodes.IS_UNUSED, 0, opcodes.IS_UNUSED, 0)
	
	forCtx.PlaceLabel(endLabel)
	c.resolveForwardJumps(forCtx, endLabel)
	
	forCtx.BreakLabel = oldBreakLabel
	forCtx.ContinueLabel = oldContinueLabel
	
	ctx.Instructions = append(ctx.Instructions, forCtx.Instructions...)
	return nil
}

func (c *ContextCompilerFixed) compileForeachLoop(ctx *CompileContext, node *ast.ForeachStatement) error {
	// Compile iterable expression
	if err := c.compileNode(ctx, node.Iterable); err != nil {
		return err
	}
	iterableVar := ctx.GetNextTemp() - 1
	
	startLabel := ctx.GetNextLabel()
	endLabel := ctx.GetNextLabel()
	
	oldBreakLabel := ctx.BreakLabel
	oldContinueLabel := ctx.ContinueLabel
	ctx.BreakLabel = endLabel
	ctx.ContinueLabel = startLabel
	
	ctx.EmitInstruction(opcodes.OP_FE_RESET, opcodes.IS_TMP_VAR, iterableVar, opcodes.IS_UNUSED, 0, opcodes.IS_UNUSED, 0)
	
	ctx.PlaceLabel(startLabel)
	
	// Fetch next element
	valueVar := ctx.GetOrCreateVariable("__foreach_value")
	var keyVar uint32
	if node.Key != nil {
		if keyId, ok := node.Key.(*ast.IdentifierNode); ok {
			keyVar = ctx.GetOrCreateVariable(keyId.Name)
		}
	}
	
	ctx.EmitInstruction(opcodes.OP_FE_FETCH, opcodes.IS_TMP_VAR, iterableVar, opcodes.IS_VAR, valueVar, opcodes.IS_VAR, keyVar)
	ctx.AddForwardJump(endLabel, len(ctx.Instructions)-1, opcodes.OP_FE_FETCH, 2)
	
	// Assign value to loop variable
	if valueId, ok := node.Value.(*ast.IdentifierNode); ok {
		targetVar := ctx.GetOrCreateVariable(valueId.Name)
		ctx.EmitInstruction(opcodes.OP_ASSIGN, opcodes.IS_VAR, valueVar, opcodes.IS_UNUSED, 0, opcodes.IS_VAR, targetVar)
	}
	
	// Compile body - single statement
	if node.Body != nil {
		if err := c.compileNode(ctx, node.Body); err != nil {
			return err
		}
	}
	
	startPos, _ := ctx.GetLabelPosition(startLabel)
	ctx.EmitInstruction(opcodes.OP_JMP, opcodes.IS_CONST, uint32(startPos), opcodes.IS_UNUSED, 0, opcodes.IS_UNUSED, 0)
	
	ctx.PlaceLabel(endLabel)
	ctx.EmitInstruction(opcodes.OP_FE_FREE, opcodes.IS_TMP_VAR, iterableVar, opcodes.IS_UNUSED, 0, opcodes.IS_UNUSED, 0)
	
	c.resolveForwardJumps(ctx, endLabel)
	
	ctx.BreakLabel = oldBreakLabel
	ctx.ContinueLabel = oldContinueLabel
	
	return nil
}

func (c *ContextCompilerFixed) compileSwitch(ctx *CompileContext, node *ast.SwitchStatement) error {
	if err := c.compileNode(ctx, node.Discriminant); err != nil {
		return err
	}
	discriminantVar := ctx.GetNextTemp() - 1
	
	endLabel := ctx.GetNextLabel()
	oldBreakLabel := ctx.BreakLabel
	ctx.BreakLabel = endLabel
	
	blockCtx := ctx.NewChildContext(ScopeBlock)
	blockCtx.BreakLabel = endLabel
	
	for i, switchCase := range node.Cases {
		if switchCase.Test != nil {
			err := c.compileNode(blockCtx, switchCase.Test)
			if err != nil {
				ctx.BreakLabel = oldBreakLabel
				return err
			}
			caseValueVar := blockCtx.GetNextTemp() - 1
			
			blockCtx.EmitInstruction(opcodes.OP_IS_EQUAL, opcodes.IS_TMP_VAR, discriminantVar, opcodes.IS_TMP_VAR, caseValueVar, opcodes.IS_TMP_VAR, blockCtx.GetNextTemp())
			comparisonVar := blockCtx.GetNextTemp() - 1
			
			caseBodyLabel := blockCtx.GetNextLabel()
			nextCaseLabel := blockCtx.GetNextLabel()
			
			blockCtx.EmitInstruction(opcodes.OP_JMPNZ, opcodes.IS_TMP_VAR, comparisonVar, opcodes.IS_CONST, 0, opcodes.IS_UNUSED, 0)
			blockCtx.AddForwardJump(caseBodyLabel, len(blockCtx.Instructions)-1, opcodes.OP_JMPNZ, 2)
			
			blockCtx.EmitInstruction(opcodes.OP_JMP, opcodes.IS_CONST, 0, opcodes.IS_UNUSED, 0, opcodes.IS_UNUSED, 0)
			blockCtx.AddForwardJump(nextCaseLabel, len(blockCtx.Instructions)-1, opcodes.OP_JMP, 1)
			
			blockCtx.PlaceLabel(caseBodyLabel)
		}
		
		for _, stmt := range switchCase.Body {
			err := c.compileNode(blockCtx, stmt)
			if err != nil {
				ctx.BreakLabel = oldBreakLabel
				return err
			}
		}
		
		if i < len(node.Cases)-1 {
			nextCaseLabel := fmt.Sprintf("next_case_%d", i)
			blockCtx.PlaceLabel(nextCaseLabel)
			c.resolveForwardJumps(blockCtx, nextCaseLabel)
		}
	}
	
	blockCtx.PlaceLabel(endLabel)
	c.resolveForwardJumps(blockCtx, endLabel)
	
	ctx.Instructions = append(ctx.Instructions, blockCtx.Instructions...)
	ctx.BreakLabel = oldBreakLabel
	
	return nil
}

func (c *ContextCompilerFixed) compileBreak(ctx *CompileContext, node *ast.BreakStatement) error {
	if ctx.BreakLabel == "" {
		return fmt.Errorf("break statement not within a loop or switch")
	}
	
	ctx.EmitInstruction(opcodes.OP_JMP, opcodes.IS_CONST, 0, opcodes.IS_UNUSED, 0, opcodes.IS_UNUSED, 0)
	ctx.AddForwardJump(ctx.BreakLabel, len(ctx.Instructions)-1, opcodes.OP_JMP, 1)
	
	return nil
}

func (c *ContextCompilerFixed) compileContinue(ctx *CompileContext, node *ast.ContinueStatement) error {
	if ctx.ContinueLabel == "" {
		return fmt.Errorf("continue statement not within a loop")
	}
	
	ctx.EmitInstruction(opcodes.OP_JMP, opcodes.IS_CONST, 0, opcodes.IS_UNUSED, 0, opcodes.IS_UNUSED, 0)
	ctx.AddForwardJump(ctx.ContinueLabel, len(ctx.Instructions)-1, opcodes.OP_JMP, 1)
	
	return nil
}

func (c *ContextCompilerFixed) compileReturn(ctx *CompileContext, node *ast.ReturnStatement) error {
	if node.Argument != nil {
		if err := c.compileNode(ctx, node.Argument); err != nil {
			return err
		}
		valueVar := ctx.GetNextTemp() - 1
		ctx.EmitInstruction(opcodes.OP_RETURN, opcodes.IS_TMP_VAR, valueVar, opcodes.IS_UNUSED, 0, opcodes.IS_UNUSED, 0)
	} else {
		nullConstant := ctx.AddConstant(values.NewNull())
		ctx.EmitInstruction(opcodes.OP_RETURN, opcodes.IS_CONST, nullConstant, opcodes.IS_UNUSED, 0, opcodes.IS_UNUSED, 0)
	}
	return nil
}

// Function and class declarations

func (c *ContextCompilerFixed) compileFunctionDeclaration(ctx *CompileContext, node *ast.FunctionDeclaration) error {
	var funcName string
	if nameNode, ok := node.Name.(*ast.IdentifierNode); ok {
		funcName = nameNode.Name
	} else {
		return fmt.Errorf("function declaration must have an identifier name")
	}
	
	functionCtx := ctx.NewChildContext(ScopeFunction)
	
	compilerFunc := &CompilerFunction{
		Name:        funcName,
		Parameters:  make([]CompilerParameter, 0),
		IsVariadic:  false,
		IsGenerator: false,
		IsAnonymous: false,
	}
	
	functionCtx.SetCurrentFunction(compilerFunc)
	
	if node.Parameters != nil {
		for _, param := range node.Parameters.Parameters {
			if paramName, ok := param.Name.(*ast.IdentifierNode); ok {
				functionCtx.GetOrCreateVariable(paramName.Name)
				
				compilerParam := CompilerParameter{
					Name:        paramName.Name,
					Type:        "",
					IsReference: false,
					HasDefault:  param.DefaultValue != nil,
				}
				
				if param.DefaultValue != nil {
					compilerParam.DefaultValue = values.NewNull()
				}
				
				compilerFunc.Parameters = append(compilerFunc.Parameters, compilerParam)
			}
		}
	}
	
	if node.Body != nil {
		for _, stmt := range node.Body {
			if err := c.compileNode(functionCtx, stmt); err != nil {
				return fmt.Errorf("error compiling function %s: %v", funcName, err)
			}
		}
	}
	
	if len(functionCtx.Instructions) == 0 || functionCtx.Instructions[len(functionCtx.Instructions)-1].Opcode != opcodes.OP_RETURN {
		nullConstant := functionCtx.AddConstant(values.NewNull())
		functionCtx.EmitInstruction(opcodes.OP_RETURN, opcodes.IS_CONST, nullConstant, opcodes.IS_UNUSED, 0, opcodes.IS_UNUSED, 0)
	}
	
	compilerFunc.Instructions = functionCtx.Instructions
	compilerFunc.Constants = functionCtx.Constants
	
	ctx.AddFunction(funcName, compilerFunc)
	
	nameConstant := ctx.AddConstant(values.NewString(funcName))
	ctx.EmitInstruction(opcodes.OP_DECLARE_FUNCTION, opcodes.IS_CONST, nameConstant, opcodes.IS_UNUSED, 0, opcodes.IS_UNUSED, 0)
	
	return nil
}

func (c *ContextCompilerFixed) compileClassDeclaration(ctx *CompileContext, node *ast.ClassExpression) error {
	var className string
	if nameNode, ok := node.Name.(*ast.IdentifierNode); ok {
		className = nameNode.Name
	} else {
		return fmt.Errorf("class declaration must have an identifier name")
	}
	
	compilerClass := &CompilerClass{
		Name:       className,
		Properties: make(map[string]*CompilerProperty),
		Methods:    make(map[string]*CompilerFunction),
		Constants:  make(map[string]*CompilerClassConstant),
		IsAbstract: false,
		IsFinal:    false,
	}
	
	if node.Extends != nil {
		if parentName, ok := node.Extends.(*ast.IdentifierNode); ok {
			compilerClass.Parent = parentName.Name
		}
	}
	
	oldCurrentClass := ctx.CurrentClass
	ctx.SetCurrentClass(compilerClass)
	
	classCtx := ctx.NewChildContext(ScopeClass)
	classCtx.SetCurrentClass(compilerClass)
	
	if node.Body != nil {
		for _, member := range node.Body {
			switch m := member.(type) {
			case *ast.PropertyDeclaration:
				if err := c.compilePropertyDeclaration(classCtx, m); err != nil {
					ctx.SetCurrentClass(oldCurrentClass)
					return fmt.Errorf("error compiling property in class %s: %v", className, err)
				}
			case *ast.FunctionDeclaration:
				if err := c.compileMethodDeclaration(classCtx, m); err != nil {
					ctx.SetCurrentClass(oldCurrentClass)
					return fmt.Errorf("error compiling method in class %s: %v", className, err)
				}
			default:
				if err := c.compileNode(classCtx, m); err != nil {
					ctx.SetCurrentClass(oldCurrentClass)
					return fmt.Errorf("error compiling class member in %s: %v", className, err)
				}
			}
		}
	}
	
	ctx.AddClass(className, compilerClass)
	
	nameConstant := ctx.AddConstant(values.NewString(className))
	ctx.EmitInstruction(opcodes.OP_DECLARE_CLASS, opcodes.IS_CONST, nameConstant, opcodes.IS_UNUSED, 0, opcodes.IS_UNUSED, 0)
	
	ctx.SetCurrentClass(oldCurrentClass)
	return nil
}

func (c *ContextCompilerFixed) compilePropertyDeclaration(ctx *CompileContext, node *ast.PropertyDeclaration) error {
	if ctx.CurrentClass == nil {
		return fmt.Errorf("property declaration outside of class")
	}
	
	property := &CompilerProperty{
		Name:       node.Name,
		Visibility: node.Visibility,
		IsStatic:   node.Static,
	}
	
	if node.DefaultValue != nil {
		property.DefaultValue = values.NewNull()
	}
	
	ctx.CurrentClass.Properties[node.Name] = property
	return nil
}

func (c *ContextCompilerFixed) compileMethodDeclaration(ctx *CompileContext, node *ast.FunctionDeclaration) error {
	if ctx.CurrentClass == nil {
		return fmt.Errorf("method declaration outside of class")
	}
	
	var methodName string
	if nameNode, ok := node.Name.(*ast.IdentifierNode); ok {
		methodName = nameNode.Name
	} else {
		return fmt.Errorf("method declaration must have an identifier name")
	}
	
	methodCtx := ctx.NewChildContext(ScopeMethod)
	methodCtx.SetCurrentClass(ctx.CurrentClass)
	
	thisSlot := methodCtx.GetOrCreateVariable("this")
	_ = thisSlot
	
	compilerFunc := &CompilerFunction{
		Name:        methodName,
		Parameters:  make([]CompilerParameter, 0),
		IsVariadic:  false,
		IsGenerator: false,
		IsAnonymous: false,
	}
	
	methodCtx.SetCurrentFunction(compilerFunc)
	
	if node.Parameters != nil {
		for _, param := range node.Parameters.Parameters {
			if paramName, ok := param.Name.(*ast.IdentifierNode); ok {
				methodCtx.GetOrCreateVariable(paramName.Name)
				
				compilerParam := CompilerParameter{
					Name:        paramName.Name,
					Type:        "",
					IsReference: false,
					HasDefault:  param.DefaultValue != nil,
				}
				
				if param.DefaultValue != nil {
					compilerParam.DefaultValue = values.NewNull()
				}
				
				compilerFunc.Parameters = append(compilerFunc.Parameters, compilerParam)
			}
		}
	}
	
	if node.Body != nil {
		for _, stmt := range node.Body {
			if err := c.compileNode(methodCtx, stmt); err != nil {
				return fmt.Errorf("error compiling method %s: %v", methodName, err)
			}
		}
	}
	
	if len(methodCtx.Instructions) == 0 || methodCtx.Instructions[len(methodCtx.Instructions)-1].Opcode != opcodes.OP_RETURN {
		nullConstant := methodCtx.AddConstant(values.NewNull())
		methodCtx.EmitInstruction(opcodes.OP_RETURN, opcodes.IS_CONST, nullConstant, opcodes.IS_UNUSED, 0, opcodes.IS_UNUSED, 0)
	}
	
	compilerFunc.Instructions = methodCtx.Instructions
	compilerFunc.Constants = methodCtx.Constants
	
	ctx.CurrentClass.Methods[methodName] = compilerFunc
	
	return nil
}

func (c *ContextCompilerFixed) compileInterfaceDeclaration(ctx *CompileContext, node *ast.InterfaceDeclaration) error {
	interfaceName := node.Name.Name
	
	compilerInterface := &CompilerInterface{
		Name:    interfaceName,
		Methods: make(map[string]*CompilerInterfaceMethod),
		Extends: make([]string, 0),
	}
	
	if node.Extends != nil {
		for _, extend := range node.Extends {
			compilerInterface.Extends = append(compilerInterface.Extends, extend.Name)
		}
	}
	
	if node.Methods != nil {
		for _, method := range node.Methods {
			interfaceMethod := &CompilerInterfaceMethod{
				Name:       method.Name.Name,
				Visibility: "public",
				Parameters: make([]*CompilerParameter, 0),
			}
			
			if method.Parameters != nil {
				for _, param := range method.Parameters.Parameters {
					paramName, ok := param.Name.(*ast.IdentifierNode)
					if !ok {
						continue
					}
					interfaceParam := &CompilerParameter{
						Name:        paramName.Name,
						Type:        "",
						IsReference: false,
						HasDefault:  param.DefaultValue != nil,
					}
					
					if param.DefaultValue != nil {
						interfaceParam.DefaultValue = values.NewNull()
					}
					
					interfaceMethod.Parameters = append(interfaceMethod.Parameters, interfaceParam)
				}
			}
			
			compilerInterface.Methods[method.Name.Name] = interfaceMethod
		}
	}
	
	ctx.AddInterface(interfaceName, compilerInterface)
	
	nameConstant := ctx.AddConstant(values.NewString(interfaceName))
	ctx.EmitInstruction(opcodes.OP_DECLARE_INTERFACE, opcodes.IS_CONST, nameConstant, opcodes.IS_UNUSED, 0, opcodes.IS_UNUSED, 0)
	
	return nil
}

func (c *ContextCompilerFixed) compileTraitDeclaration(ctx *CompileContext, node *ast.TraitDeclaration) error {
	traitName := node.Name.Name
	
	compilerTrait := &CompilerTrait{
		Name:       traitName,
		Properties: make(map[string]*CompilerProperty),
		Methods:    make(map[string]*CompilerFunction),
	}
	
	oldCurrentClass := ctx.CurrentClass
	tempClass := &CompilerClass{
		Name:       traitName,
		Properties: make(map[string]*CompilerProperty),
		Methods:    make(map[string]*CompilerFunction),
		Constants:  make(map[string]*CompilerClassConstant),
	}
	ctx.SetCurrentClass(tempClass)
	
	if node.Body != nil {
		for _, member := range node.Body {
			switch m := member.(type) {
			case *ast.PropertyDeclaration:
				if err := c.compilePropertyDeclaration(ctx, m); err != nil {
					ctx.SetCurrentClass(oldCurrentClass)
					return fmt.Errorf("error compiling property in trait %s: %v", traitName, err)
				}
				for name, prop := range tempClass.Properties {
					compilerTrait.Properties[name] = prop
				}
			case *ast.FunctionDeclaration:
				if err := c.compileMethodDeclaration(ctx, m); err != nil {
					ctx.SetCurrentClass(oldCurrentClass)
					return fmt.Errorf("error compiling method in trait %s: %v", traitName, err)
				}
				for name, method := range tempClass.Methods {
					compilerTrait.Methods[name] = method
				}
			default:
				if err := c.compileNode(ctx, m); err != nil {
					ctx.SetCurrentClass(oldCurrentClass)
					return fmt.Errorf("error compiling trait member in %s: %v", traitName, err)
				}
			}
		}
	}
	
	ctx.AddTrait(traitName, compilerTrait)
	
	nameConstant := ctx.AddConstant(values.NewString(traitName))
	ctx.EmitInstruction(opcodes.OP_DECLARE_TRAIT, opcodes.IS_CONST, nameConstant, opcodes.IS_UNUSED, 0, opcodes.IS_UNUSED, 0)
	
	ctx.SetCurrentClass(oldCurrentClass)
	return nil
}

// Collection and function calls

func (c *ContextCompilerFixed) compileArray(ctx *CompileContext, node *ast.ArrayExpression) error {
	arrayVar := ctx.GetNextTemp()
	ctx.EmitInstruction(opcodes.OP_INIT_ARRAY, opcodes.IS_UNUSED, 0, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, arrayVar)
	
	if node.Elements != nil {
		for _, element := range node.Elements {
			if element == nil {
				continue
			}
			
			switch elem := element.(type) {
			case *ast.ArrayElementExpression:
				if err := c.compileNode(ctx, elem.Value); err != nil {
					return err
				}
				valueVar := ctx.GetNextTemp() - 1
				
				if elem.Key != nil {
					if err := c.compileNode(ctx, elem.Key); err != nil {
						return err
					}
					keyVar := ctx.GetNextTemp() - 1
					ctx.EmitInstruction(opcodes.OP_ADD_ARRAY_ELEMENT, opcodes.IS_TMP_VAR, valueVar, opcodes.IS_TMP_VAR, keyVar, opcodes.IS_TMP_VAR, arrayVar)
				} else {
					ctx.EmitInstruction(opcodes.OP_ADD_ARRAY_ELEMENT, opcodes.IS_TMP_VAR, valueVar, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, arrayVar)
				}
			default:
				if err := c.compileNode(ctx, elem); err != nil {
					return err
				}
				valueVar := ctx.GetNextTemp() - 1
				ctx.EmitInstruction(opcodes.OP_ADD_ARRAY_ELEMENT, opcodes.IS_TMP_VAR, valueVar, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, arrayVar)
			}
		}
	}
	
	return nil
}

func (c *ContextCompilerFixed) compilePropertyAccess(ctx *CompileContext, node *ast.PropertyAccessExpression) error {
	if err := c.compileNode(ctx, node.Object); err != nil {
		return err
	}
	objectVar := ctx.GetNextTemp() - 1
	
	if err := c.compileNode(ctx, node.Property); err != nil {
		return err
	}
	propertyVar := ctx.GetNextTemp() - 1
	
	resultVar := ctx.GetNextTemp()
	ctx.EmitInstruction(opcodes.OP_FETCH_OBJ_R, opcodes.IS_TMP_VAR, objectVar, opcodes.IS_TMP_VAR, propertyVar, opcodes.IS_TMP_VAR, resultVar)
	
	return nil
}

func (c *ContextCompilerFixed) compileMethodCall(ctx *CompileContext, node *ast.MethodCallExpression) error {
	if err := c.compileNode(ctx, node.Object); err != nil {
		return err
	}
	objectVar := ctx.GetNextTemp() - 1
	
	if err := c.compileNode(ctx, node.Method); err != nil {
		return err
	}
	methodVar := ctx.GetNextTemp() - 1
	
	ctx.EmitInstruction(opcodes.OP_INIT_METHOD_CALL, opcodes.IS_TMP_VAR, objectVar, opcodes.IS_TMP_VAR, methodVar, opcodes.IS_UNUSED, 0)
	
	if node.Arguments != nil {
		for i, arg := range node.Arguments.Arguments {
			if err := c.compileNode(ctx, arg); err != nil {
				return err
			}
			argVar := ctx.GetNextTemp() - 1
			argNum := uint32(i)
			ctx.EmitInstruction(opcodes.OP_SEND_VAL, opcodes.IS_TMP_VAR, argVar, opcodes.IS_CONST, argNum, opcodes.IS_UNUSED, 0)
		}
	}
	
	resultVar := ctx.GetNextTemp()
	ctx.EmitInstruction(opcodes.OP_DO_FCALL, opcodes.IS_UNUSED, 0, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, resultVar)
	
	return nil
}

func (c *ContextCompilerFixed) compileFunctionCall(ctx *CompileContext, node *ast.CallExpression) error {
	if err := c.compileNode(ctx, node.Callee); err != nil {
		return err
	}
	functionVar := ctx.GetNextTemp() - 1
	
	ctx.EmitInstruction(opcodes.OP_INIT_FCALL, opcodes.IS_TMP_VAR, functionVar, opcodes.IS_UNUSED, 0, opcodes.IS_UNUSED, 0)
	
	if node.Arguments != nil {
		for i, arg := range node.Arguments.Arguments {
			if err := c.compileNode(ctx, arg); err != nil {
				return err
			}
			argVar := ctx.GetNextTemp() - 1
			argNum := uint32(i)
			ctx.EmitInstruction(opcodes.OP_SEND_VAL, opcodes.IS_TMP_VAR, argVar, opcodes.IS_CONST, argNum, opcodes.IS_UNUSED, 0)
		}
	}
	
	resultVar := ctx.GetNextTemp()
	ctx.EmitInstruction(opcodes.OP_DO_FCALL, opcodes.IS_UNUSED, 0, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, resultVar)
	
	return nil
}

func (c *ContextCompilerFixed) compileAnonymousFunction(ctx *CompileContext, node *ast.AnonymousFunctionExpression) error {
	functionCtx := ctx.NewChildContext(ScopeFunction)
	
	anonName := fmt.Sprintf("__lambda_%d", len(ctx.Functions))
	compilerFunc := &CompilerFunction{
		Name:        anonName,
		Parameters:  make([]CompilerParameter, 0),
		IsVariadic:  false,
		IsGenerator: false,
		IsAnonymous: true,
	}
	
	functionCtx.SetCurrentFunction(compilerFunc)
	
	if node.Parameters != nil {
		for _, param := range node.Parameters.Parameters {
			if paramName, ok := param.Name.(*ast.IdentifierNode); ok {
				functionCtx.GetOrCreateVariable(paramName.Name)
				
				compilerParam := CompilerParameter{
					Name:        paramName.Name,
					Type:        "", 
					IsReference: false,
					HasDefault:  param.DefaultValue != nil,
				}
				
				if param.DefaultValue != nil {
					compilerParam.DefaultValue = values.NewNull()
				}
				
				compilerFunc.Parameters = append(compilerFunc.Parameters, compilerParam)
			}
		}
	}
	
	if node.Body != nil {
		for _, stmt := range node.Body {
			if err := c.compileNode(functionCtx, stmt); err != nil {
				return fmt.Errorf("error compiling anonymous function: %v", err)
			}
		}
	}
	
	if len(functionCtx.Instructions) == 0 || functionCtx.Instructions[len(functionCtx.Instructions)-1].Opcode != opcodes.OP_RETURN {
		nullConstant := functionCtx.AddConstant(values.NewNull())
		functionCtx.EmitInstruction(opcodes.OP_RETURN, opcodes.IS_CONST, nullConstant, opcodes.IS_UNUSED, 0, opcodes.IS_UNUSED, 0)
	}
	
	compilerFunc.Instructions = functionCtx.Instructions
	compilerFunc.Constants = functionCtx.Constants
	
	ctx.AddFunction(anonName, compilerFunc)
	
	functionConstant := ctx.AddConstant(values.NewString(anonName))
	resultVar := ctx.GetNextTemp()
	ctx.EmitInstruction(opcodes.OP_CREATE_CLOSURE, opcodes.IS_CONST, functionConstant, opcodes.IS_UNUSED, 0, opcodes.IS_TMP_VAR, resultVar)
	
	return nil
}

// Helper method to resolve forward jumps
func (c *ContextCompilerFixed) resolveForwardJumps(ctx *CompileContext, label string) {
	if jumps, exists := ctx.ForwardJumps[label]; exists {
		labelPos, labelExists := ctx.GetLabelPosition(label)
		if labelExists {
			for _, jump := range jumps {
				if jump.instructionIndex < len(ctx.Instructions) {
					instruction := &ctx.Instructions[jump.instructionIndex]
					switch jump.operand {
					case 1:
						instruction.Op1 = uint32(labelPos)
					case 2:
						instruction.Op2 = uint32(labelPos)
					}
				}
			}
		}
		delete(ctx.ForwardJumps, label)
	}
}