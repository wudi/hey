package testutils

import (
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wudi/php-parser/ast"
)

// ASTAssertions 提供AST特定的断言
type ASTAssertions struct {
	t *testing.T
}

// NewASTAssertions 创建AST断言工具
func NewASTAssertions(t *testing.T) *ASTAssertions {
	return &ASTAssertions{t: t}
}

// AssertVariable 断言变量节点
func (a *ASTAssertions) AssertVariable(node ast.Node, expectedName string) *ast.Variable {
	a.t.Helper()
	variable, ok := node.(*ast.Variable)
	require.True(a.t, ok, "Node should be Variable, got %T", node)
	assert.Equal(a.t, expectedName, variable.Name, "Variable name mismatch")
	return variable
}

// AssertStringLiteral 断言字符串字面量
func (a *ASTAssertions) AssertStringLiteral(node ast.Node, expectedValue, expectedRaw string) *ast.StringLiteral {
	a.t.Helper()
	stringLit, ok := node.(*ast.StringLiteral)
	require.True(a.t, ok, "Node should be StringLiteral, got %T", node)
	assert.Equal(a.t, expectedValue, stringLit.Value, "String value mismatch")
	assert.Equal(a.t, expectedRaw, stringLit.Raw, "String raw mismatch")
	return stringLit
}

// AssertNumberLiteral 断言数字字面量
func (a *ASTAssertions) AssertNumberLiteral(node ast.Node, expectedValue string) *ast.NumberLiteral {
	a.t.Helper()
	numLit, ok := node.(*ast.NumberLiteral)
	require.True(a.t, ok, "Node should be NumberLiteral, got %T", node)
	assert.Equal(a.t, expectedValue, numLit.Value, "Number value mismatch")
	return numLit
}

// AssertAssignment 断言赋值表达式
func (a *ASTAssertions) AssertAssignment(node ast.Node, expectedOperator string) *ast.AssignmentExpression {
	a.t.Helper()
	assignment, ok := node.(*ast.AssignmentExpression)
	require.True(a.t, ok, "Node should be AssignmentExpression, got %T", node)
	assert.Equal(a.t, expectedOperator, assignment.Operator, "Assignment operator mismatch")
	return assignment
}

// AssertBinaryExpression 断言二元表达式
func (a *ASTAssertions) AssertBinaryExpression(node ast.Node, expectedOperator string) *ast.BinaryExpression {
	a.t.Helper()
	binExpr, ok := node.(*ast.BinaryExpression)
	require.True(a.t, ok, "Node should be BinaryExpression, got %T", node)
	assert.Equal(a.t, expectedOperator, binExpr.Operator, "Binary operator mismatch")
	return binExpr
}

// AssertExpressionStatement 断言表达式语句
func (a *ASTAssertions) AssertExpressionStatement(stmt ast.Statement) *ast.ExpressionStatement {
	a.t.Helper()
	exprStmt, ok := stmt.(*ast.ExpressionStatement)
	require.True(a.t, ok, "Statement should be ExpressionStatement, got %T", stmt)
	return exprStmt
}

// AssertEchoStatement 断言echo语句
func (a *ASTAssertions) AssertEchoStatement(stmt ast.Statement, expectedArgCount int) *ast.EchoStatement {
	a.t.Helper()
	echoStmt, ok := stmt.(*ast.EchoStatement)
	require.True(a.t, ok, "Statement should be EchoStatement, got %T", stmt)
	assert.Len(a.t, echoStmt.Arguments.Arguments, expectedArgCount, "Echo argument count mismatch")
	return echoStmt
}

// AssertClass 断言类表达式
func (a *ASTAssertions) AssertClass(node ast.Node, expectedName string) *ast.ClassExpression {
	a.t.Helper()
	classExpr, ok := node.(*ast.ClassExpression)
	require.True(a.t, ok, "Node should be ClassExpression, got %T", node)
	
	nameIdent, ok := classExpr.Name.(*ast.IdentifierNode)
	require.True(a.t, ok, "Class name should be IdentifierNode, got %T", classExpr.Name)
	assert.Equal(a.t, expectedName, nameIdent.Name, "Class name mismatch")
	return classExpr
}

// AssertIdentifier 断言标识符节点
func (a *ASTAssertions) AssertIdentifier(node ast.Node, expectedName string) *ast.IdentifierNode {
	a.t.Helper()
	ident, ok := node.(*ast.IdentifierNode)
	require.True(a.t, ok, "Node should be IdentifierNode, got %T", node)
	assert.Equal(a.t, expectedName, ident.Name, "Identifier name mismatch")
	return ident
}

// AssertProgramBody 断言程序体
func (a *ASTAssertions) AssertProgramBody(program *ast.Program, expectedLength int) []ast.Statement {
	a.t.Helper()
	require.NotNil(a.t, program, "Program should not be nil")
	assert.Len(a.t, program.Body, expectedLength, "Program body length mismatch")
	return program.Body
}

// ValidationFunc 验证函数类型
type ValidationFunc func(*TestContext)

// ValidateVariable 创建变量验证函数
func ValidateVariable(name string) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
		assertions.AssertVariable(assignment.Left, name)
	}
}

// ValidateStringAssignment 创建字符串赋值验证函数
func ValidateStringAssignment(varName, expectedValue, expectedRaw string) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
		assertions.AssertVariable(assignment.Left, varName)
		assertions.AssertStringLiteral(assignment.Right, expectedValue, expectedRaw)
	}
}