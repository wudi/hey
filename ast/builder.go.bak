package ast

import (
	"github.com/wudi/php-parser/lexer"
)

// ASTBuilder AST构建器，用于创建符合PHP官方结构的AST节点
type ASTBuilder struct {
	// 可以在这里添加构建器的状态和配置
}

// NewASTBuilder 创建新的AST构建器
func NewASTBuilder() *ASTBuilder {
	return &ASTBuilder{}
}

// CreateZval 创建zval节点（用于字面值）
func (b *ASTBuilder) CreateZval(pos lexer.Position, value interface{}) Node {
	switch v := value.(type) {
	case string:
		return NewStringLiteral(pos, v, "\""+v+"\"")
	case int, int32, int64:
		return NewNumberLiteral(pos, "integer", "integer")
	case float32, float64:
		return NewNumberLiteral(pos, "float", "float")
	case bool:
		return NewBooleanLiteral(pos, v)
	case nil:
		return NewNullLiteral(pos)
	default:
		return NewNullLiteral(pos)
	}
}

// CreateVar 创建变量节点
func (b *ASTBuilder) CreateVar(pos lexer.Position, name string) Node {
	return NewVariable(pos, name)
}

// CreateBinaryOp 创建二元操作节点
func (b *ASTBuilder) CreateBinaryOp(pos lexer.Position, left, right Node, operator string) Node {
	leftExpr, ok1 := left.(Expression)
	rightExpr, ok2 := right.(Expression)
	if !ok1 || !ok2 {
		return nil
	}
	
	return NewBinaryExpression(pos, leftExpr, operator, rightExpr)
}

// CreateAssign 创建赋值节点
func (b *ASTBuilder) CreateAssign(pos lexer.Position, left, right Node) Node {
	leftExpr, ok1 := left.(Expression)
	rightExpr, ok2 := right.(Expression)
	if !ok1 || !ok2 {
		return nil
	}
	
	return NewAssignmentExpression(pos, leftExpr, "=", rightExpr)
}

// CreateUnaryOp 创建一元操作节点
func (b *ASTBuilder) CreateUnaryOp(pos lexer.Position, operand Node, operator string, prefix bool) Node {
	operandExpr, ok := operand.(Expression)
	if !ok {
		return nil
	}
	
	return NewUnaryExpression(pos, operator, operandExpr, prefix)
}

// CreateCall 创建函数调用节点
func (b *ASTBuilder) CreateCall(pos lexer.Position, callee Node, args []Node) Node {
	calleeExpr, ok := callee.(Expression)
	if !ok {
		return nil
	}
	
	call := NewCallExpression(pos, calleeExpr)
	for _, arg := range args {
		if argExpr, ok := arg.(Expression); ok {
			call.Arguments = append(call.Arguments, argExpr)
		}
	}
	
	return call
}

// CreateArray 创建数组节点
func (b *ASTBuilder) CreateArray(pos lexer.Position, elements []Node) Node {
	array := NewArrayExpression(pos)
	for _, elem := range elements {
		if elemExpr, ok := elem.(Expression); ok {
			array.Elements = append(array.Elements, elemExpr)
		}
	}
	
	return array
}

// CreateEcho 创建echo语句节点
func (b *ASTBuilder) CreateEcho(pos lexer.Position, args []Node) Node {
	echo := NewEchoStatement(pos)
	for _, arg := range args {
		if argExpr, ok := arg.(Expression); ok {
			echo.Arguments = append(echo.Arguments, argExpr)
		}
	}
	
	return echo
}

// CreateReturn 创建return语句节点
func (b *ASTBuilder) CreateReturn(pos lexer.Position, arg Node) Node {
	if arg == nil {
		return NewReturnStatement(pos, nil)
	}
	
	if argExpr, ok := arg.(Expression); ok {
		return NewReturnStatement(pos, argExpr)
	}
	
	return NewReturnStatement(pos, nil)
}

// CreateIf 创建if语句节点
func (b *ASTBuilder) CreateIf(pos lexer.Position, test Node, consequent []Node, alternate []Node) Node {
	testExpr, ok := test.(Expression)
	if !ok {
		return nil
	}
	
	ifStmt := NewIfStatement(pos, testExpr)
	
	for _, stmt := range consequent {
		if s, ok := stmt.(Statement); ok {
			ifStmt.Consequent = append(ifStmt.Consequent, s)
		}
	}
	
	for _, stmt := range alternate {
		if s, ok := stmt.(Statement); ok {
			ifStmt.Alternate = append(ifStmt.Alternate, s)
		}
	}
	
	return ifStmt
}

// CreateWhile 创建while语句节点
func (b *ASTBuilder) CreateWhile(pos lexer.Position, test Node, body []Node) Node {
	testExpr, ok := test.(Expression)
	if !ok {
		return nil
	}
	
	whileStmt := NewWhileStatement(pos, testExpr)
	
	for _, stmt := range body {
		if s, ok := stmt.(Statement); ok {
			whileStmt.Body = append(whileStmt.Body, s)
		}
	}
	
	return whileStmt
}

// CreateFor 创建for语句节点
func (b *ASTBuilder) CreateFor(pos lexer.Position, init, test, update Node, body []Node) Node {
	forStmt := NewForStatement(pos)
	
	if init != nil {
		if initExpr, ok := init.(Expression); ok {
			forStmt.Init = initExpr
		}
	}
	
	if test != nil {
		if testExpr, ok := test.(Expression); ok {
			forStmt.Test = testExpr
		}
	}
	
	if update != nil {
		if updateExpr, ok := update.(Expression); ok {
			forStmt.Update = updateExpr
		}
	}
	
	for _, stmt := range body {
		if s, ok := stmt.(Statement); ok {
			forStmt.Body = append(forStmt.Body, s)
		}
	}
	
	return forStmt
}

// CreateFuncDecl 创建函数声明节点
func (b *ASTBuilder) CreateFuncDecl(pos lexer.Position, name Node, params []Parameter, body []Node) Node {
	nameId, ok := name.(Identifier)
	if !ok {
		return nil
	}
	
	funcDecl := NewFunctionDeclaration(pos, nameId)
	funcDecl.Parameters = params
	
	for _, stmt := range body {
		if s, ok := stmt.(Statement); ok {
			funcDecl.Body = append(funcDecl.Body, s)
		}
	}
	
	return funcDecl
}

// CreateStmtList 创建语句列表节点
func (b *ASTBuilder) CreateStmtList(pos lexer.Position, stmts []Node) Node {
	if len(stmts) == 1 {
		// 如果只有一个语句，直接返回
		return stmts[0]
	}
	
	program := NewProgram(pos)
	for _, stmt := range stmts {
		if s, ok := stmt.(Statement); ok {
			program.Body = append(program.Body, s)
		}
	}
	
	return program
}

// CreateExprList 创建表达式列表节点
func (b *ASTBuilder) CreateExprList(pos lexer.Position, exprs []Node) Node {
	if len(exprs) == 1 {
		// 如果只有一个表达式，直接返回
		return exprs[0]
	}
	
	// 创建一个包含多个表达式的复合节点
	// 这里可以根据需要实现更复杂的逻辑
	return exprs[0] // 简单实现
}

// CreateConst 创建常量节点
func (b *ASTBuilder) CreateConst(pos lexer.Position, name string) Node {
	return NewIdentifierNode(pos, name)
}

// CreateBreak 创建break语句节点
func (b *ASTBuilder) CreateBreak(pos lexer.Position) Node {
	return NewBreakStatement(pos)
}

// CreateContinue 创建continue语句节点
func (b *ASTBuilder) CreateContinue(pos lexer.Position) Node {
	return NewContinueStatement(pos)
}

// CreateExpressionStatement 创建表达式语句节点
func (b *ASTBuilder) CreateExpressionStatement(pos lexer.Position, expr Node) Node {
	if exprNode, ok := expr.(Expression); ok {
		return NewExpressionStatement(pos, exprNode)
	}
	return nil
}

// CreateBlock 创建代码块节点
func (b *ASTBuilder) CreateBlock(pos lexer.Position, stmts []Node) Node {
	block := NewBlockStatement(pos)
	for _, stmt := range stmts {
		if s, ok := stmt.(Statement); ok {
			block.Body = append(block.Body, s)
		}
	}
	return block
}