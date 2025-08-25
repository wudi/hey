package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yourname/php-parser/ast"
	"github.com/yourname/php-parser/lexer"
)

func TestParsing_VariableDeclaration(t *testing.T) {
	input := `<?php $name = "John"; ?>`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	checkParserErrors(t, p)
	assert.NotNil(t, program)
	assert.Len(t, program.Body, 1)

	stmt := program.Body[0]
	exprStmt, ok := stmt.(*ast.ExpressionStatement)
	assert.True(t, ok, "Statement should be ExpressionStatement")

	assignment, ok := exprStmt.Expression.(*ast.AssignmentExpression)
	assert.True(t, ok, "Expression should be AssignmentExpression")

	// 检查左侧变量
	variable, ok := assignment.Left.(*ast.Variable)
	assert.True(t, ok, "Left side should be Variable")
	assert.Equal(t, "$name", variable.Name)

	// 检查操作符
	assert.Equal(t, "=", assignment.Operator)

	// 检查右侧字符串字面量
	stringLit, ok := assignment.Right.(*ast.StringLiteral)
	assert.True(t, ok, "Right side should be StringLiteral")
	assert.Equal(t, "John", stringLit.Value)
	assert.Equal(t, `"John"`, stringLit.Raw)
}

func TestParsing_EchoStatement(t *testing.T) {
	input := `<?php echo "Hello, World!"; ?>`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	checkParserErrors(t, p)
	assert.NotNil(t, program)
	assert.Len(t, program.Body, 1)

	stmt := program.Body[0]
	echoStmt, ok := stmt.(*ast.EchoStatement)
	assert.True(t, ok, "Statement should be EchoStatement")

	assert.Len(t, echoStmt.Arguments, 1)
	
	stringLit, ok := echoStmt.Arguments[0].(*ast.StringLiteral)
	assert.True(t, ok, "Argument should be StringLiteral")
	assert.Equal(t, "Hello, World!", stringLit.Value)
}

func TestParsing_EchoMultipleArguments(t *testing.T) {
	input := `<?php echo "Hello", " ", "World!"; ?>`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	checkParserErrors(t, p)
	assert.NotNil(t, program)
	assert.Len(t, program.Body, 1)

	stmt := program.Body[0]
	echoStmt, ok := stmt.(*ast.EchoStatement)
	assert.True(t, ok, "Statement should be EchoStatement")

	assert.Len(t, echoStmt.Arguments, 3)

	values := []string{"Hello", " ", "World!"}
	for i, arg := range echoStmt.Arguments {
		stringLit, ok := arg.(*ast.StringLiteral)
		assert.True(t, ok, "Argument should be StringLiteral")
		assert.Equal(t, values[i], stringLit.Value)
	}
}

func TestParsing_IntegerLiterals(t *testing.T) {
	input := `<?php $x = 123; ?>`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	checkParserErrors(t, p)
	assert.NotNil(t, program)
	assert.Len(t, program.Body, 1)

	stmt := program.Body[0]
	exprStmt, ok := stmt.(*ast.ExpressionStatement)
	assert.True(t, ok, "Statement should be ExpressionStatement")

	assignment, ok := exprStmt.Expression.(*ast.AssignmentExpression)
	assert.True(t, ok, "Expression should be AssignmentExpression")

	numberLit, ok := assignment.Right.(*ast.NumberLiteral)
	assert.True(t, ok, "Right side should be NumberLiteral")
	assert.Equal(t, "123", numberLit.Value)
	assert.Equal(t, "integer", numberLit.Kind)
}

func TestParsing_FloatLiterals(t *testing.T) {
	input := `<?php $pi = 3.14; ?>`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	checkParserErrors(t, p)
	assert.NotNil(t, program)

	stmt := program.Body[0]
	exprStmt, ok := stmt.(*ast.ExpressionStatement)
	assert.True(t, ok, "Statement should be ExpressionStatement")

	assignment, ok := exprStmt.Expression.(*ast.AssignmentExpression)
	assert.True(t, ok, "Expression should be AssignmentExpression")

	numberLit, ok := assignment.Right.(*ast.NumberLiteral)
	assert.True(t, ok, "Right side should be NumberLiteral")
	assert.Equal(t, "3.14", numberLit.Value)
	assert.Equal(t, "float", numberLit.Kind)
}

func TestParsing_BinaryExpressions(t *testing.T) {
	tests := []struct {
		input    string
		operator string
	}{
		{`<?php $result = 5 + 3; ?>`, "+"},
		{`<?php $result = 5 - 3; ?>`, "-"},
		{`<?php $result = 5 * 3; ?>`, "*"},
		{`<?php $result = 5 / 3; ?>`, "/"},
		{`<?php $result = 5 % 3; ?>`, "%"},
		{`<?php $result = 5 == 3; ?>`, "=="},
		{`<?php $result = 5 != 3; ?>`, "!="},
		{`<?php $result = 5 < 3; ?>`, "<"},
		{`<?php $result = 5 > 3; ?>`, ">"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()

		checkParserErrors(t, p)
		assert.NotNil(t, program)
		assert.Len(t, program.Body, 1)

		stmt := program.Body[0]
		exprStmt, ok := stmt.(*ast.ExpressionStatement)
		assert.True(t, ok, "Statement should be ExpressionStatement")

		assignment, ok := exprStmt.Expression.(*ast.AssignmentExpression)
		assert.True(t, ok, "Expression should be AssignmentExpression")

		binaryExpr, ok := assignment.Right.(*ast.BinaryExpression)
		assert.True(t, ok, "Right side should be BinaryExpression")

		assert.Equal(t, tt.operator, binaryExpr.Operator)

		// 检查操作数
		leftNum, ok := binaryExpr.Left.(*ast.NumberLiteral)
		assert.True(t, ok, "Left operand should be NumberLiteral")
		assert.Equal(t, "5", leftNum.Value)

		rightNum, ok := binaryExpr.Right.(*ast.NumberLiteral)
		assert.True(t, ok, "Right operand should be NumberLiteral")
		assert.Equal(t, "3", rightNum.Value)
	}
}

func TestParsing_PrefixExpressions(t *testing.T) {
	tests := []struct {
		input    string
		operator string
	}{
		{`<?php $x = -5; ?>`, "-"},
		{`<?php $x = !true; ?>`, "!"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()

		checkParserErrors(t, p)
		assert.NotNil(t, program)
		assert.Len(t, program.Body, 1)

		stmt := program.Body[0]
		exprStmt, ok := stmt.(*ast.ExpressionStatement)
		assert.True(t, ok, "Statement should be ExpressionStatement")

		assignment, ok := exprStmt.Expression.(*ast.AssignmentExpression)
		assert.True(t, ok, "Expression should be AssignmentExpression")

		unaryExpr, ok := assignment.Right.(*ast.UnaryExpression)
		assert.True(t, ok, "Right side should be UnaryExpression")

		assert.Equal(t, tt.operator, unaryExpr.Operator)
		assert.True(t, unaryExpr.Prefix)
	}
}

func TestParsing_IfStatement(t *testing.T) {
	input := `<?php if ($x > 5) { echo "big"; } ?>`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	checkParserErrors(t, p)
	assert.NotNil(t, program)
	assert.Len(t, program.Body, 1)

	stmt := program.Body[0]
	ifStmt, ok := stmt.(*ast.IfStatement)
	assert.True(t, ok, "Statement should be IfStatement")

	// 检查条件
	binaryExpr, ok := ifStmt.Test.(*ast.BinaryExpression)
	assert.True(t, ok, "Condition should be BinaryExpression")
	assert.Equal(t, ">", binaryExpr.Operator)

	// 检查 consequent 
	assert.Len(t, ifStmt.Consequent, 1)
	echoStmt, ok := ifStmt.Consequent[0].(*ast.EchoStatement)
	assert.True(t, ok, "Consequent should contain EchoStatement")

	assert.Len(t, echoStmt.Arguments, 1)
	stringLit, ok := echoStmt.Arguments[0].(*ast.StringLiteral)
	assert.True(t, ok, "Echo argument should be StringLiteral")
	assert.Equal(t, "big", stringLit.Value)
}

func TestParsing_IfElseStatement(t *testing.T) {
	input := `<?php if ($x > 5) { echo "big"; } else { echo "small"; } ?>`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	checkParserErrors(t, p)
	assert.NotNil(t, program)
	assert.Len(t, program.Body, 1)

	stmt := program.Body[0]
	ifStmt, ok := stmt.(*ast.IfStatement)
	assert.True(t, ok, "Statement should be IfStatement")

	// 检查 consequent
	assert.Len(t, ifStmt.Consequent, 1)
	
	// 检查 alternate
	assert.Len(t, ifStmt.Alternate, 1)
	echoStmt, ok := ifStmt.Alternate[0].(*ast.EchoStatement)
	assert.True(t, ok, "Alternate should contain EchoStatement")

	stringLit, ok := echoStmt.Arguments[0].(*ast.StringLiteral)
	assert.True(t, ok, "Echo argument should be StringLiteral")
	assert.Equal(t, "small", stringLit.Value)
}

func TestParsing_WhileStatement(t *testing.T) {
	input := `<?php while ($i < 10) { $i++; } ?>`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	checkParserErrors(t, p)
	assert.NotNil(t, program)
	assert.Len(t, program.Body, 1)

	stmt := program.Body[0]
	whileStmt, ok := stmt.(*ast.WhileStatement)
	assert.True(t, ok, "Statement should be WhileStatement")

	// 检查条件
	binaryExpr, ok := whileStmt.Test.(*ast.BinaryExpression)
	assert.True(t, ok, "Condition should be BinaryExpression")
	assert.Equal(t, "<", binaryExpr.Operator)

	// 检查循环体
	assert.Len(t, whileStmt.Body, 1)
	exprStmt, ok := whileStmt.Body[0].(*ast.ExpressionStatement)
	assert.True(t, ok, "Body should contain ExpressionStatement")

	unaryExpr, ok := exprStmt.Expression.(*ast.UnaryExpression)
	assert.True(t, ok, "Expression should be UnaryExpression")
	assert.Equal(t, "++", unaryExpr.Operator)
	assert.False(t, unaryExpr.Prefix) // $i++ 是后缀
}

func TestParsing_ForStatement(t *testing.T) {
	input := `<?php for ($i = 0; $i < 10; $i++) { echo $i; } ?>`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	checkParserErrors(t, p)
	assert.NotNil(t, program)
	assert.Len(t, program.Body, 1)

	stmt := program.Body[0]
	forStmt, ok := stmt.(*ast.ForStatement)
	assert.True(t, ok, "Statement should be ForStatement")

	// 检查初始化
	assert.NotNil(t, forStmt.Init)
	initAssign, ok := forStmt.Init.(*ast.AssignmentExpression)
	assert.True(t, ok, "Init should be AssignmentExpression")
	
	variable, ok := initAssign.Left.(*ast.Variable)
	assert.True(t, ok, "Init left should be Variable")
	assert.Equal(t, "$i", variable.Name)

	// 检查条件
	assert.NotNil(t, forStmt.Test)
	testBinary, ok := forStmt.Test.(*ast.BinaryExpression)
	assert.True(t, ok, "Test should be BinaryExpression")
	assert.Equal(t, "<", testBinary.Operator)

	// 检查更新
	assert.NotNil(t, forStmt.Update)
	updateUnary, ok := forStmt.Update.(*ast.UnaryExpression)
	assert.True(t, ok, "Update should be UnaryExpression")
	assert.Equal(t, "++", updateUnary.Operator)

	// 检查循环体
	assert.Len(t, forStmt.Body, 1)
	_, ok = forStmt.Body[0].(*ast.EchoStatement)
	assert.True(t, ok, "Body should contain EchoStatement")
}

func TestParsing_FunctionDeclaration(t *testing.T) {
	input := `<?php function greet($name) { echo "Hello, ", $name; } ?>`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	checkParserErrors(t, p)
	assert.NotNil(t, program)
	assert.Len(t, program.Body, 1)

	stmt := program.Body[0]
	funcDecl, ok := stmt.(*ast.FunctionDeclaration)
	assert.True(t, ok, "Statement should be FunctionDeclaration")

	// 检查函数名
	identifier, ok := funcDecl.Name.(*ast.IdentifierNode)
	assert.True(t, ok, "Function name should be IdentifierNode")
	assert.Equal(t, "greet", identifier.Name)

	// 检查参数
	assert.Len(t, funcDecl.Parameters, 1)
	assert.Equal(t, "$name", funcDecl.Parameters[0].Name)

	// 检查函数体
	assert.Len(t, funcDecl.Body, 1)
	echoStmt, ok := funcDecl.Body[0].(*ast.EchoStatement)
	assert.True(t, ok, "Function body should contain EchoStatement")

	assert.Len(t, echoStmt.Arguments, 2)
}

func TestParsing_ReturnStatement(t *testing.T) {
	input := `<?php function add($a, $b) { return $a + $b; } ?>`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	checkParserErrors(t, p)
	assert.NotNil(t, program)
	assert.Len(t, program.Body, 1)

	stmt := program.Body[0]
	funcDecl, ok := stmt.(*ast.FunctionDeclaration)
	assert.True(t, ok, "Statement should be FunctionDeclaration")

	assert.Len(t, funcDecl.Body, 1)
	returnStmt, ok := funcDecl.Body[0].(*ast.ReturnStatement)
	assert.True(t, ok, "Function body should contain ReturnStatement")

	assert.NotNil(t, returnStmt.Argument)
	binaryExpr, ok := returnStmt.Argument.(*ast.BinaryExpression)
	assert.True(t, ok, "Return argument should be BinaryExpression")
	assert.Equal(t, "+", binaryExpr.Operator)
}

func TestParsing_ArrayExpression(t *testing.T) {
	input := `<?php $arr = array(1, 2, 3); ?>`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	checkParserErrors(t, p)
	assert.NotNil(t, program)
	assert.Len(t, program.Body, 1)

	stmt := program.Body[0]
	exprStmt, ok := stmt.(*ast.ExpressionStatement)
	assert.True(t, ok, "Statement should be ExpressionStatement")

	assignment, ok := exprStmt.Expression.(*ast.AssignmentExpression)
	assert.True(t, ok, "Expression should be AssignmentExpression")

	arrayExpr, ok := assignment.Right.(*ast.ArrayExpression)
	assert.True(t, ok, "Right side should be ArrayExpression")

	assert.Len(t, arrayExpr.Elements, 3)

	for i, element := range arrayExpr.Elements {
		numberLit, ok := element.(*ast.NumberLiteral)
		assert.True(t, ok, "Array element should be NumberLiteral")
		expected := []string{"1", "2", "3"}
		assert.Equal(t, expected[i], numberLit.Value)
	}
}

func TestParsing_GroupedExpression(t *testing.T) {
	input := `<?php $result = (5 + 3) * 2; ?>`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	checkParserErrors(t, p)
	assert.NotNil(t, program)
	assert.Len(t, program.Body, 1)

	stmt := program.Body[0]
	exprStmt, ok := stmt.(*ast.ExpressionStatement)
	assert.True(t, ok, "Statement should be ExpressionStatement")

	assignment, ok := exprStmt.Expression.(*ast.AssignmentExpression)
	assert.True(t, ok, "Expression should be AssignmentExpression")

	binaryExpr, ok := assignment.Right.(*ast.BinaryExpression)
	assert.True(t, ok, "Right side should be BinaryExpression")
	assert.Equal(t, "*", binaryExpr.Operator)

	// 左侧应该是 (5 + 3)
	leftBinary, ok := binaryExpr.Left.(*ast.BinaryExpression)
	assert.True(t, ok, "Left side should be BinaryExpression")
	assert.Equal(t, "+", leftBinary.Operator)
}

func TestParsing_OperatorPrecedence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`<?php $x = 1 + 2 * 3; ?>`, "(1 + (2 * 3))"},
		{`<?php $x = (1 + 2) * 3; ?>`, "((1 + 2) * 3)"},
		{`<?php $x = 1 == 2 + 3; ?>`, "(1 == (2 + 3))"},
		{`<?php $x = 1 + 2 == 3; ?>`, "((1 + 2) == 3)"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()

		checkParserErrors(t, p)
		assert.NotNil(t, program)
		assert.Len(t, program.Body, 1)

		stmt := program.Body[0]
		exprStmt, ok := stmt.(*ast.ExpressionStatement)
		assert.True(t, ok, "Statement should be ExpressionStatement")

		assignment, ok := exprStmt.Expression.(*ast.AssignmentExpression)
		assert.True(t, ok, "Expression should be AssignmentExpression")

		actual := assignment.Right.String()
		assert.Equal(t, tt.expected, actual, "Input: %s", tt.input)
	}
}

// 辅助函数

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}