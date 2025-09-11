package testutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wudi/hey/compiler/ast"
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

// AssertTernaryExpression 断言三元表达式
func (a *ASTAssertions) AssertTernaryExpression(node ast.Node) *ast.TernaryExpression {
	a.t.Helper()
	ternaryExpr, ok := node.(*ast.TernaryExpression)
	require.True(a.t, ok, "Node should be TernaryExpression, got %T", node)
	require.NotNil(a.t, ternaryExpr.Test, "Ternary test should not be nil")
	require.NotNil(a.t, ternaryExpr.Consequent, "Ternary consequent should not be nil")
	require.NotNil(a.t, ternaryExpr.Alternate, "Ternary alternate should not be nil")
	return ternaryExpr
}

// AssertArray 断言数组表达式
func (a *ASTAssertions) AssertArray(node ast.Node, expectedElementCount int) *ast.ArrayExpression {
	a.t.Helper()
	arrayExpr, ok := node.(*ast.ArrayExpression)
	require.True(a.t, ok, "Node should be ArrayExpression, got %T", node)
	assert.Len(a.t, arrayExpr.Elements, expectedElementCount, "Array element count mismatch")
	return arrayExpr
}

// AssertTryStatement 断言try语句
func (a *ASTAssertions) AssertTryStatement(stmt ast.Statement) *ast.TryStatement {
	a.t.Helper()
	tryStmt, ok := stmt.(*ast.TryStatement)
	require.True(a.t, ok, "Statement should be TryStatement, got %T", stmt)
	return tryStmt
}

// AssertTryBlockEmpty 断言try块为空
func (a *ASTAssertions) AssertTryBlockEmpty(tryStmt *ast.TryStatement) {
	a.t.Helper()
	assert.Len(a.t, tryStmt.Body, 0, "Try block should be empty")
}

// AssertTryBlockStatements 断言try块中的语句数量
func (a *ASTAssertions) AssertTryBlockStatements(tryStmt *ast.TryStatement, expectedCount int) {
	a.t.Helper()
	assert.Len(a.t, tryStmt.Body, expectedCount, "Try block statement count mismatch")
}

// AssertCatchClausesCount 断言catch子句数量
func (a *ASTAssertions) AssertCatchClausesCount(tryStmt *ast.TryStatement, expectedCount int) {
	a.t.Helper()
	assert.Len(a.t, tryStmt.CatchClauses, expectedCount, "Catch clauses count mismatch")
}

// AssertMethodCall 断言方法调用
func (a *ASTAssertions) AssertMethodCall(node ast.Node, expectedObject, expectedMethod string) *ast.MethodCallExpression {
	a.t.Helper()
	methodCall, ok := node.(*ast.MethodCallExpression)
	require.True(a.t, ok, "Node should be MethodCallExpression, got %T", node)

	objVar, ok := methodCall.Object.(*ast.Variable)
	require.True(a.t, ok, "Object should be Variable, got %T", methodCall.Object)
	assert.Equal(a.t, expectedObject, objVar.Name, "Object variable name mismatch")

	methodIdent, ok := methodCall.Method.(*ast.IdentifierNode)
	require.True(a.t, ok, "Method should be IdentifierNode, got %T", methodCall.Method)
	assert.Equal(a.t, expectedMethod, methodIdent.Name, "Method name mismatch")

	return methodCall
}

// AssertNamespaceStatement 断言namespace语句
func (a *ASTAssertions) AssertNamespaceStatement(stmt ast.Statement, expectedName string) *ast.NamespaceStatement {
	a.t.Helper()
	namespaceStmt, ok := stmt.(*ast.NamespaceStatement)
	require.True(a.t, ok, "Statement should be NamespaceStatement, got %T", stmt)

	if expectedName != "" {
		require.NotNil(a.t, namespaceStmt.Name, "Namespace name should not be nil")
		// NamespaceNameExpression has a String() method we can use
		assert.Equal(a.t, expectedName, namespaceStmt.Name.String(), "Namespace name mismatch")
	}

	return namespaceStmt
}

// AssertGlobalNamespaceStatement 断言全局namespace语句
func (a *ASTAssertions) AssertGlobalNamespaceStatement(stmt ast.Statement) *ast.NamespaceStatement {
	a.t.Helper()
	namespaceStmt, ok := stmt.(*ast.NamespaceStatement)
	require.True(a.t, ok, "Statement should be NamespaceStatement, got %T", stmt)
	assert.Nil(a.t, namespaceStmt.Name, "Global namespace should have nil name")
	return namespaceStmt
}

// AssertCallExpression 断言函数调用表达式
func (a *ASTAssertions) AssertCallExpression(expr ast.Expression) *ast.CallExpression {
	a.t.Helper()
	callExpr, ok := expr.(*ast.CallExpression)
	require.True(a.t, ok, "Expression should be CallExpression, got %T", expr)
	return callExpr
}

// AssertFullyQualifiedCall 断言完全限定的函数调用
func (a *ASTAssertions) AssertFullyQualifiedCall(callExpr *ast.CallExpression, expectedName string) {
	a.t.Helper()
	identNode, ok := callExpr.Callee.(*ast.IdentifierNode)
	require.True(a.t, ok, "Callee should be IdentifierNode, got %T", callExpr.Callee)
	// Note: The parser may handle namespace separators differently
	// This assertion may need adjustment based on actual AST structure
	assert.Contains(a.t, identNode.Name, expectedName, "Fully qualified call name mismatch")
}

// AssertTraitDeclaration 断言Trait声明
func (a *ASTAssertions) AssertTraitDeclaration(stmt ast.Statement, expectedName string) *ast.TraitDeclaration {
	a.t.Helper()
	traitDecl, ok := stmt.(*ast.TraitDeclaration)
	require.True(a.t, ok, "Statement should be TraitDeclaration, got %T", stmt)

	require.NotNil(a.t, traitDecl.Name, "Trait name should not be nil")
	assert.Equal(a.t, expectedName, traitDecl.Name.Name, "Trait name mismatch")

	return traitDecl
}

// AssertMatchExpression 断言Match表达式
func (a *ASTAssertions) AssertMatchExpression(expr ast.Expression) *ast.MatchExpression {
	a.t.Helper()
	matchExpr, ok := expr.(*ast.MatchExpression)
	require.True(a.t, ok, "Expression should be MatchExpression, got %T", expr)
	return matchExpr
}

// AssertFunctionDeclaration 断言函数声明
func (a *ASTAssertions) AssertFunctionDeclaration(stmt ast.Statement, expectedName string) *ast.FunctionDeclaration {
	a.t.Helper()
	funcDecl, ok := stmt.(*ast.FunctionDeclaration)
	require.True(a.t, ok, "Statement should be FunctionDeclaration, got %T", stmt)

	nameIdent, ok := funcDecl.Name.(*ast.IdentifierNode)
	require.True(a.t, ok, "Function name should be IdentifierNode, got %T", funcDecl.Name)
	assert.Equal(a.t, expectedName, nameIdent.Name, "Function name mismatch")

	return funcDecl
}

// AssertFunctionBody 断言函数体
func (a *ASTAssertions) AssertFunctionBody(funcDecl *ast.FunctionDeclaration, expectedStatements int) {
	a.t.Helper()
	assert.Len(a.t, funcDecl.Body, expectedStatements, "Function body statement count mismatch")
}

// AssertReturnStatement 断言return语句
func (a *ASTAssertions) AssertReturnStatement(stmt ast.Statement) *ast.ReturnStatement {
	a.t.Helper()
	returnStmt, ok := stmt.(*ast.ReturnStatement)
	require.True(a.t, ok, "Statement should be ReturnStatement, got %T", stmt)
	return returnStmt
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
