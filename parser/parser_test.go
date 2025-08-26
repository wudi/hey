package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/lexer"
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

func TestParsing_TypedParameters(t *testing.T) {
	input := `<?php 
function func_name(
    string $commandline,
    ?array $env = null,
    ?string $stdin = null,
    bool $captureStdIn = true,
    bool $captureStdErr = true
) {

} ?>`

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
	assert.Equal(t, "func_name", identifier.Name)

	// 检查参数
	assert.Len(t, funcDecl.Parameters, 5)

	// 第一个参数: string $commandline
	param1 := funcDecl.Parameters[0]
	assert.Equal(t, "$commandline", param1.Name)
	assert.Equal(t, "string", param1.Type)
	assert.Nil(t, param1.DefaultValue)

	// 第二个参数: ?array $env = null
	param2 := funcDecl.Parameters[1]
	assert.Equal(t, "$env", param2.Name)
	assert.Equal(t, "?array", param2.Type)
	assert.NotNil(t, param2.DefaultValue)

	// 第三个参数: ?string $stdin = null
	param3 := funcDecl.Parameters[2]
	assert.Equal(t, "$stdin", param3.Name)
	assert.Equal(t, "?string", param3.Type)
	assert.NotNil(t, param3.DefaultValue)

	// 第四个参数: bool $captureStdIn = true
	param4 := funcDecl.Parameters[3]
	assert.Equal(t, "$captureStdIn", param4.Name)
	assert.Equal(t, "bool", param4.Type)
	assert.NotNil(t, param4.DefaultValue)

	// 第五个参数: bool $captureStdErr = true
	param5 := funcDecl.Parameters[4]
	assert.Equal(t, "$captureStdErr", param5.Name)
	assert.Equal(t, "bool", param5.Type)
	assert.NotNil(t, param5.DefaultValue)

	// 检查函数体为空
	assert.Len(t, funcDecl.Body, 0)
}

func TestParsing_BitwiseOperations(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		operator string
	}{
		{
			name:     "Bitwise right shift",
			input:    `<?php $result = $value >> 2; ?>`,
			operator: ">>",
		},
		{
			name:     "Bitwise left shift", 
			input:    `<?php $result = $value << 3; ?>`,
			operator: "<<",
		},
		{
			name:     "Bitwise AND",
			input:    `<?php $result = $value & 0xFF; ?>`,
			operator: "&",
		},
		{
			name:     "Bitwise OR",
			input:    `<?php $result = $value | 0b1010; ?>`,
			operator: "|",
		},
		{
			name:     "Complex bitwise expression",
			input:    `<?php (($stat["exitcode"] >> 28) & 0b1111) === 0b1100; ?>`,
			operator: "===",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			checkParserErrors(t, p)
			assert.NotNil(t, program)
			assert.Len(t, program.Body, 1)

			stmt := program.Body[0]
			exprStmt, ok := stmt.(*ast.ExpressionStatement)
			assert.True(t, ok, "Statement should be ExpressionStatement")

			// For complex bitwise expression, check the top-level operator
			if tt.name == "Complex bitwise expression" {
				binaryExpr, ok := exprStmt.Expression.(*ast.BinaryExpression)
				assert.True(t, ok, "Expression should be BinaryExpression")
				assert.Equal(t, tt.operator, binaryExpr.Operator)
				
				// Check left side is a bitwise AND operation
				leftExpr, ok := binaryExpr.Left.(*ast.BinaryExpression)
				assert.True(t, ok, "Left side should be BinaryExpression")
				assert.Equal(t, "&", leftExpr.Operator)
				
				// Check the left side of bitwise AND is a right shift operation
				leftLeftExpr, ok := leftExpr.Left.(*ast.BinaryExpression)
				assert.True(t, ok, "Left left should be BinaryExpression")
				assert.Equal(t, ">>", leftLeftExpr.Operator)
			} else {
				// For simple expressions, check assignment or standalone expression
				if assignment, ok := exprStmt.Expression.(*ast.AssignmentExpression); ok {
					binaryExpr, ok := assignment.Right.(*ast.BinaryExpression)
					assert.True(t, ok, "Right side should be BinaryExpression")
					assert.Equal(t, tt.operator, binaryExpr.Operator)
				} else {
					binaryExpr, ok := exprStmt.Expression.(*ast.BinaryExpression)
					assert.True(t, ok, "Expression should be BinaryExpression")
					assert.Equal(t, tt.operator, binaryExpr.Operator)
				}
			}
		})
	}
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

func TestParsing_HeredocStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple Heredoc",
			input:    `<?php $str = <<<EOD
Hello World
EOD; ?>`,
			expected: "<<<EOD\nHello World\nEOD",
		},
		{
			name:     "Heredoc with variable",
			input:    `<?php $str = <<<EOD
Hello $name
EOD; ?>`,
			expected: "<<<EOD\nHello $name\nEOD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			stringLit, ok := assignment.Right.(*ast.StringLiteral)
			assert.True(t, ok, "Right side should be StringLiteral")
			assert.Equal(t, tt.expected, stringLit.Raw)
		})
	}
}

func TestParsing_NowdocStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple Nowdoc",
			input:    `<?php $str = <<<'EOD'
Hello World
EOD; ?>`,
			expected: "Hello World\n",
		},
		{
			name:     "Nowdoc with $variable (no interpolation)",
			input:    `<?php $str = <<<'EOD'
Hello $name
EOD; ?>`,
			expected: "Hello $name\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			stringLit, ok := assignment.Right.(*ast.StringLiteral)
			assert.True(t, ok, "Right side should be StringLiteral")
			assert.Equal(t, tt.expected, stringLit.Value)
		})
	}
}

func TestParsing_StringInterpolation(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		partCount   int
		firstPart   string
		secondPart  string
	}{
		{
			name:        "Simple variable interpolation",
			input:       `<?php $str = "Hello $name"; ?>`,
			partCount:   2,
			firstPart:   "Hello ",
			secondPart:  "$name",
		},
		{
			name:        "String with multiple variables",
			input:       `<?php $str = "Hello $first and $second"; ?>`,
			partCount:   4,  // "Hello ", "$first", " and ", "$second"
			firstPart:   "Hello ",
			secondPart:  "$first",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			// 检查是否是插值字符串
			if interpolatedStr, ok := assignment.Right.(*ast.InterpolatedStringExpression); ok {
				assert.Len(t, interpolatedStr.Parts, tt.partCount)
				
				// 检查第一部分
				if len(interpolatedStr.Parts) > 0 {
					if stringPart, ok := interpolatedStr.Parts[0].(*ast.StringLiteral); ok {
						assert.Equal(t, tt.firstPart, stringPart.Value)
					}
				}
				
				// 检查第二部分
				if len(interpolatedStr.Parts) > 1 {
					if len(tt.secondPart) > 0 {
						if tt.secondPart == "$name" {
							if varPart, ok := interpolatedStr.Parts[1].(*ast.Variable); ok {
								assert.Equal(t, tt.secondPart, varPart.Name)
							}
						} else {
							if stringPart, ok := interpolatedStr.Parts[1].(*ast.StringLiteral); ok {
								assert.Equal(t, tt.secondPart, stringPart.Value)
							}
						}
					}
				}
			} else {
				// 如果不是插值字符串，应该是普通字符串
				stringLit, ok := assignment.Right.(*ast.StringLiteral)
				assert.True(t, ok, "Right side should be StringLiteral or InterpolatedString")
				_ = stringLit
			}
		})
	}
}

func TestParsing_ComplexExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		operator string
	}{
		{
			name:     "Identical comparison",
			input:    `<?php $result = $a === $b; ?>`,
			operator: "===",
		},
		{
			name:     "Not identical comparison",
			input:    `<?php $result = $a !== $b; ?>`,
			operator: "!==",
		},
		{
			name:     "Less than or equal",
			input:    `<?php $result = $a <= $b; ?>`,
			operator: "<=",
		},
		{
			name:     "Greater than or equal",
			input:    `<?php $result = $a >= $b; ?>`,
			operator: ">=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
		})
	}
}

func TestParsing_ArraySyntax(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		isShort  bool
		expected int
	}{
		{
			name:     "Array function syntax",
			input:    `<?php $arr = array(1, 2, 3); ?>`,
			isShort:  false,
			expected: 3,
		},
		{
			name:     "Short array syntax",
			input:    `<?php $arr = [1, 2, 3]; ?>`,
			isShort:  true,
			expected: 3,
		},
		{
			name:     "Associative array",
			input:    `<?php $arr = ["name" => "John", "age" => 30]; ?>`,
			isShort:  true,
			expected: 2,
		},
		{
			name:     "Empty array",
			input:    `<?php $arr = []; ?>`,
			isShort:  true,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			arrayExpr, ok := assignment.Right.(*ast.ArrayExpression)
			assert.True(t, ok, "Right side should be ArrayExpression")
			assert.Len(t, arrayExpr.Elements, tt.expected)
		})
	}
}

func TestParsing_ArrayWithComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name: "Array with inline comment after element",
			input: `<?php $consts = [
    'E_USER_NOTICE',
    'E_STRICT', // TODO test
    'E_RECOVERABLE_ERROR'
]; ?>`,
			expected: []string{"E_USER_NOTICE", "E_STRICT", "E_RECOVERABLE_ERROR"},
		},
		{
			name: "Array with multiple inline comments",
			input: `<?php $consts = [
    'FIRST',    // First constant
    'SECOND',   // Second constant  
    'THIRD'     // Third constant
]; ?>`,
			expected: []string{"FIRST", "SECOND", "THIRD"},
		},
		{
			name: "Array with comments between elements",
			input: `<?php $values = [
    1,
    // This is a comment
    2,
    // Another comment
    3
]; ?>`,
			expected: []string{"1", "2", "3"},
		},
		{
			name: "Array with trailing comment",
			input: `<?php $data = [
    'item1',
    'item2',
    'item3', // Trailing comment
]; ?>`,
			expected: []string{"item1", "item2", "item3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			arrayExpr, ok := assignment.Right.(*ast.ArrayExpression)
			assert.True(t, ok, "Right side should be ArrayExpression")
			assert.Len(t, arrayExpr.Elements, len(tt.expected))

			for i, element := range arrayExpr.Elements {
				if numberLit, ok := element.(*ast.NumberLiteral); ok {
					assert.Equal(t, tt.expected[i], numberLit.Value)
				} else if stringLit, ok := element.(*ast.StringLiteral); ok {
					assert.Equal(t, tt.expected[i], stringLit.Value)
				} else {
					t.Errorf("Unexpected element type: %T", element)
				}
			}
		})
	}
}

func TestParsing_FunctionCalls(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		funcName string
		argCount int
	}{
		{
			name:     "Function call no args",
			input:    `<?php time(); ?>`,
			funcName: "time",
			argCount: 0,
		},
		{
			name:     "Function call with args",
			input:    `<?php strlen("hello"); ?>`,
			funcName: "strlen",
			argCount: 1,
		},
		{
			name:     "Function call multiple args",
			input:    `<?php substr("hello", 1, 3); ?>`,
			funcName: "substr",
			argCount: 3,
		},
		{
			name:     "Function call in expression context",
			input:    `<?php save_text($file, <<<'PHP'
test
PHP); ?>`,
			funcName: "save_text",
			argCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			checkParserErrors(t, p)
			assert.NotNil(t, program)
			assert.Len(t, program.Body, 1)

			stmt := program.Body[0]
			exprStmt, ok := stmt.(*ast.ExpressionStatement)
			assert.True(t, ok, "Statement should be ExpressionStatement")

			callExpr, ok := exprStmt.Expression.(*ast.CallExpression)
			assert.True(t, ok, "Expression should be CallExpression")

			// 检查函数名
			if identifier, ok := callExpr.Callee.(*ast.IdentifierNode); ok {
				assert.Equal(t, tt.funcName, identifier.Name)
			}

			// 检查参数数量
			if tt.argCount == 0 {
				assert.Nil(t, callExpr.Arguments)
			} else {
				assert.NotNil(t, callExpr.Arguments)
				assert.Len(t, callExpr.Arguments, tt.argCount)
			}
		})
	}
}

func TestParsing_ErrorCases(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError bool
	}{
		{
			name:          "Unclosed string",
			input:         `<?php $var = "unclosed string ?>`,
			expectedError: true,
		},
		{
			name:          "Unclosed array",
			input:         `<?php $arr = [1, 2, 3 ?>`,
			expectedError: true,
		},
		{
			name:          "Valid syntax",
			input:         `<?php $var = "test"; ?>`,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			errors := p.Errors()
			if tt.expectedError {
				assert.NotEmpty(t, errors, "Expected parser errors but got none")
			} else {
				assert.Empty(t, errors, "Expected no parser errors but got: %v", errors)
				assert.NotNil(t, program)
			}
		})
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
