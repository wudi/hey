package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.NotNil(t, param1.Type)
	assert.Equal(t, "string", param1.Type.String())
	assert.Nil(t, param1.DefaultValue)

	// 第二个参数: ?array $env = null
	param2 := funcDecl.Parameters[1]
	assert.Equal(t, "$env", param2.Name)
	assert.NotNil(t, param2.Type)
	assert.Equal(t, "?array", param2.Type.String())
	assert.NotNil(t, param2.DefaultValue)

	// 第三个参数: ?string $stdin = null
	param3 := funcDecl.Parameters[2]
	assert.Equal(t, "$stdin", param3.Name)
	assert.NotNil(t, param3.Type)
	assert.Equal(t, "?string", param3.Type.String())
	assert.NotNil(t, param3.DefaultValue)

	// 第四个参数: bool $captureStdIn = true
	param4 := funcDecl.Parameters[3]
	assert.Equal(t, "$captureStdIn", param4.Name)
	assert.NotNil(t, param4.Type)
	assert.Equal(t, "bool", param4.Type.String())
	assert.NotNil(t, param4.DefaultValue)

	// 第五个参数: bool $captureStdErr = true
	param5 := funcDecl.Parameters[4]
	assert.Equal(t, "$captureStdErr", param5.Name)
	assert.NotNil(t, param5.Type)
	assert.Equal(t, "bool", param5.Type.String())
	assert.NotNil(t, param5.DefaultValue)

	// 检查函数体为空
	assert.Len(t, funcDecl.Body, 0)
}

// TestParsing_FunctionReturnTypes tests various type tokens in function return types and parameters
func TestParsing_FunctionReturnTypes(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedParams []struct {
			name string
			typ  string
		}
		expectedReturnType string
	}{
		{
			name:  "Function with array parameters and array return type",
			input: `<?php function foo(array $ar1, array $ar2, bool $is_reg, array $w): array { } ?>`,
			expectedParams: []struct {
				name string
				typ  string
			}{
				{"$ar1", "array"},
				{"$ar2", "array"},
				{"$is_reg", "bool"},
				{"$w", "array"},
			},
			expectedReturnType: "array",
		},
		{
			name:  "Function with callable parameter and callable return type",
			input: `<?php function test(callable $cb, int $x): callable { } ?>`,
			expectedParams: []struct {
				name string
				typ  string
			}{
				{"$cb", "callable"},
				{"$x", "int"},
			},
			expectedReturnType: "callable",
		},
		{
			name:  "Function with mixed types and string return type",
			input: `<?php function process(string $name, array $data, callable $callback): string { } ?>`,
			expectedParams: []struct {
				name string
				typ  string
			}{
				{"$name", "string"},
				{"$data", "array"},
				{"$callback", "callable"},
			},
			expectedReturnType: "string",
		},
		{
			name:  "Function with nullable types",
			input: `<?php function nullable(?array $data, ?callable $cb): ?array { } ?>`,
			expectedParams: []struct {
				name string
				typ  string
			}{
				{"$data", "?array"},
				{"$cb", "?callable"},
			},
			expectedReturnType: "?array",
		},
		{
			name:  "Function with all scalar types",
			input: `<?php function allTypes(int $i, float $f, string $s, bool $b): void { } ?>`,
			expectedParams: []struct {
				name string
				typ  string
			}{
				{"$i", "int"},
				{"$f", "float"},
				{"$s", "string"},
				{"$b", "bool"},
			},
			expectedReturnType: "void",
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
			funcDecl, ok := stmt.(*ast.FunctionDeclaration)
			assert.True(t, ok, "Statement should be FunctionDeclaration")

			// Check parameters
			assert.Len(t, funcDecl.Parameters, len(tt.expectedParams))
			for i, expected := range tt.expectedParams {
				assert.Equal(t, expected.name, funcDecl.Parameters[i].Name, "Parameter %d name mismatch", i)
				if expected.typ != "" {
					assert.NotNil(t, funcDecl.Parameters[i].Type, "Parameter %d should have a type", i)
					assert.Equal(t, expected.typ, funcDecl.Parameters[i].Type.String(), "Parameter %d type mismatch", i)
				} else {
					assert.Nil(t, funcDecl.Parameters[i].Type, "Parameter %d should not have a type", i)
				}
			}

			// Check return type
			if tt.expectedReturnType != "" {
				assert.NotNil(t, funcDecl.ReturnType, "Function should have a return type")
				assert.Equal(t, tt.expectedReturnType, funcDecl.ReturnType.String(), "Return type mismatch")
			} else {
				assert.Nil(t, funcDecl.ReturnType, "Function should not have a return type")
			}
		})
	}
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
		name             string
		input            string
		expectedContent  string
		expectInterpolation bool
		expectedParts    int
	}{
		{
			name:             "Simple Heredoc",
			input:            `<?php $str = <<<EOD
Hello World
EOD; ?>`,
			expectedContent:  "Hello World\n",
			expectInterpolation: false,
			expectedParts:    1,
		},
		{
			name:             "Heredoc with variable",
			input:            `<?php $str = <<<EOD
Hello $name
EOD; ?>`,
			expectedContent:  "Hello ", // First part content
			expectInterpolation: true,
			expectedParts:    3, // "Hello " + $name + "\n"
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

			if tt.expectInterpolation {
				// Should be InterpolatedStringExpression for variable interpolation
				interpolatedStr, ok := assignment.Right.(*ast.InterpolatedStringExpression)
				assert.True(t, ok, "Right side should be InterpolatedStringExpression for heredoc with variables")
				assert.Len(t, interpolatedStr.Parts, tt.expectedParts)
				
				// Check first part is string literal with expected content
				if len(interpolatedStr.Parts) > 0 {
					firstPart, ok := interpolatedStr.Parts[0].(*ast.StringLiteral)
					assert.True(t, ok, "First part should be StringLiteral")
					assert.Equal(t, tt.expectedContent, firstPart.Value)
				}
			} else {
				// Should be StringLiteral for simple heredoc
				stringLit, ok := assignment.Right.(*ast.StringLiteral)
				assert.True(t, ok, "Right side should be StringLiteral for simple heredoc")
				assert.Equal(t, tt.expectedContent, stringLit.Value)
			}
		})
	}
}

func TestParsing_HeredocWithComplexInterpolation(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedParts    int
		expectedVarNames []string
	}{
		{
			name: "Heredoc with curly brace interpolation",
			input: `<?php $sh_script = <<<SH
#!/bin/sh
{$exported_environment}
case "$1" in
"gdb")
    gdb --args {$orig_cmd}
    ;;
*)
    {$orig_cmd}
    ;;
esac
SH; ?>`,
			expectedParts:    7, // Text, var, text, var, text, var, text
			expectedVarNames: []string{"$exported_environment", "$orig_cmd", "$orig_cmd"},
		},
		{
			name: "Heredoc with mixed interpolation styles",
			input: `<?php $template = <<<HTML
<div class="user">
    <h1>Hello $username</h1>
    <p>Your ID is: {$user_id}</p>
    <p>Email: {$email}</p>
</div>
HTML; ?>`,
			expectedParts:    7, // text, var, text, var, text, var, text
			expectedVarNames: []string{"$username", "$user_id", "$email"},
		},
		{
			name: "Heredoc with only curly brace interpolation",
			input: `<?php $cmd = <<<CMD
{$binary} --option {$value}
CMD; ?>`,
			expectedParts:    4, // var, text, var, newline
			expectedVarNames: []string{"$binary", "$value"},
		},
		{
			name: "Heredoc with no interpolation",
			input: `<?php $static = <<<TEXT
This is plain text with no variables
TEXT; ?>`,
			expectedParts:    1,
			expectedVarNames: []string{},
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

			if len(tt.expectedVarNames) > 0 {
				// Should be InterpolatedStringExpression for heredoc with variables
				interpolatedStr, ok := assignment.Right.(*ast.InterpolatedStringExpression)
				assert.True(t, ok, "Right side should be InterpolatedStringExpression for heredoc with variables")
				assert.Len(t, interpolatedStr.Parts, tt.expectedParts)

				// Count and verify variables
				varCount := 0
				for _, part := range interpolatedStr.Parts {
					if variable, ok := part.(*ast.Variable); ok {
						assert.Contains(t, tt.expectedVarNames, variable.Name, "Variable name should be in expected list")
						varCount++
					}
				}
				assert.Equal(t, len(tt.expectedVarNames), varCount, "Should find all expected variables")
			} else {
				// Should be StringLiteral for heredoc without variables
				stringLit, ok := assignment.Right.(*ast.StringLiteral)
				assert.True(t, ok, "Right side should be StringLiteral for heredoc without variables")
				assert.NotEmpty(t, stringLit.Value)
			}
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

func TestParsing_AnonymousFunctions(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedParams []struct {
			name string
			typ  string
		}
		expectUseClause bool
		useVarCount     int
	}{
		{
			name: "Anonymous function with typed parameters and use clause",
			input: `<?php foo(function (int $errno, string $errstr) use ($abc): bool {
    return true;
}); ?>`,
			expectedParams: []struct {
				name string
				typ  string
			}{
				{"$errno", "int"},
				{"$errstr", "string"},
			},
			expectUseClause: true,
			useVarCount:     1,
		},
		{
			name: "Anonymous function with simple parameters",
			input: `<?php $callback = function ($x, $y) {
    return $x + $y;
}; ?>`,
			expectedParams: []struct {
				name string
				typ  string
			}{
				{"$x", ""},
				{"$y", ""},
			},
			expectUseClause: false,
			useVarCount:     0,
		},
		{
			name: "Anonymous function with mixed typed parameters",
			input: `<?php $fn = function (string $name, $value, array $options) use ($config) {
    return $config[$name] ?? $value;
}; ?>`,
			expectedParams: []struct {
				name string
				typ  string
			}{
				{"$name", "string"},
				{"$value", ""},
				{"$options", "array"},
			},
			expectUseClause: true,
			useVarCount:     1,
		},
		{
			name: "Anonymous function with no parameters",
			input: `<?php $empty = function () {
    echo "Hello";
}; ?>`,
			expectedParams: []struct {
				name string
				typ  string
			}{},
			expectUseClause: false,
			useVarCount:     0,
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

			var anonFunc *ast.AnonymousFunctionExpression
			
			// Check if it's a direct assignment or function call
			if assignment, ok := exprStmt.Expression.(*ast.AssignmentExpression); ok {
				// Direct assignment: $var = function(...) {...}
				anonFunc, ok = assignment.Right.(*ast.AnonymousFunctionExpression)
				assert.True(t, ok, "Right side should be AnonymousFunctionExpression")
			} else if callExpr, ok := exprStmt.Expression.(*ast.CallExpression); ok {
				// Function call: foo(function(...) {...})
				assert.NotNil(t, callExpr.Arguments)
				assert.Len(t, callExpr.Arguments, 1)
				anonFunc, ok = callExpr.Arguments[0].(*ast.AnonymousFunctionExpression)
				assert.True(t, ok, "Argument should be AnonymousFunctionExpression")
			} else {
				t.Fatal("Expression should be either AssignmentExpression or CallExpression")
			}

			// Verify parameters
			assert.Len(t, anonFunc.Parameters, len(tt.expectedParams))
			for i, expectedParam := range tt.expectedParams {
				assert.Equal(t, expectedParam.name, anonFunc.Parameters[i].Name)
				if expectedParam.typ != "" {
					assert.NotNil(t, anonFunc.Parameters[i].Type, "Parameter %d should have a type", i)
					assert.Equal(t, expectedParam.typ, anonFunc.Parameters[i].Type.String())
				} else {
					assert.Nil(t, anonFunc.Parameters[i].Type, "Parameter %d should not have a type", i)
				}
			}

			// Verify use clause
			if tt.expectUseClause {
				assert.NotNil(t, anonFunc.UseClause)
				assert.Len(t, anonFunc.UseClause, tt.useVarCount)
			} else {
				// UseClause can be nil or empty
				if anonFunc.UseClause != nil {
					assert.Len(t, anonFunc.UseClause, 0)
				}
			}

			// Verify function body exists
			assert.NotNil(t, anonFunc.Body)
			assert.Greater(t, len(anonFunc.Body), 0, "Function should have a body")
		})
	}
}

func TestParsing_AnonymousFunctionReferenceUse(t *testing.T) {
	tests := []struct {
		name                  string
		input                 string
		expectedUseClauses    int
		expectedReferenceVars []string
		expectedNormalVars    []string
	}{
		{
			name: "Anonymous function with reference variable in use clause",
			input: `<?php
$ini = preg_replace_callback("{ENV:(\S+)}", function ($m) use (&$skip) { }, $ini);`,
			expectedUseClauses:    1,
			expectedReferenceVars: []string{"$skip"},
			expectedNormalVars:    []string{},
		},
		{
			name: "Anonymous function with mixed use clause (normal and reference vars)",
			input: `<?php
$result = function ($x) use ($normal, &$reference) {
    return $x + $normal + $reference;
};`,
			expectedUseClauses:    2,
			expectedReferenceVars: []string{"$reference"},
			expectedNormalVars:    []string{"$normal"},
		},
		{
			name: "Anonymous function with multiple reference variables",
			input: `<?php
$callback = function () use (&$ref1, &$ref2, &$ref3) {
    return $ref1 + $ref2 + $ref3;
};`,
			expectedUseClauses:    3,
			expectedReferenceVars: []string{"$ref1", "$ref2", "$ref3"},
			expectedNormalVars:    []string{},
		},
		{
			name: "Anonymous function with complex mixed use clause",
			input: `<?php
$complex = function ($param) use ($var1, &$ref1, $var2, &$ref2) {
    return $param;
};`,
			expectedUseClauses:    4,
			expectedReferenceVars: []string{"$ref1", "$ref2"},
			expectedNormalVars:    []string{"$var1", "$var2"},
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

			var anonFunc *ast.AnonymousFunctionExpression

			// Navigate to find the anonymous function
			if assignment, ok := exprStmt.Expression.(*ast.AssignmentExpression); ok {
				// Direct assignment: $var = function(...) {...}
				if callExpr, ok := assignment.Right.(*ast.CallExpression); ok {
					// preg_replace_callback case
					assert.Len(t, callExpr.Arguments, 3)
					anonFunc, ok = callExpr.Arguments[1].(*ast.AnonymousFunctionExpression)
					assert.True(t, ok, "Second argument should be AnonymousFunctionExpression")
				} else {
					anonFunc, ok = assignment.Right.(*ast.AnonymousFunctionExpression)
					assert.True(t, ok, "Right side should be AnonymousFunctionExpression")
				}
			}

			assert.NotNil(t, anonFunc)
			assert.NotNil(t, anonFunc.UseClause)
			assert.Len(t, anonFunc.UseClause, tt.expectedUseClauses)

			// Count reference and normal variables
			foundReferenceVars := []string{}
			foundNormalVars := []string{}

			for _, useVar := range anonFunc.UseClause {
				if unaryExpr, ok := useVar.(*ast.UnaryExpression); ok {
					// This is a reference variable (&$var)
					assert.Equal(t, "&", unaryExpr.Operator)
					assert.True(t, unaryExpr.Prefix)
					if varExpr, ok := unaryExpr.Operand.(*ast.Variable); ok {
						foundReferenceVars = append(foundReferenceVars, varExpr.Name)
					}
				} else if varExpr, ok := useVar.(*ast.Variable); ok {
					// This is a normal variable ($var)
					foundNormalVars = append(foundNormalVars, varExpr.Name)
				}
			}

			assert.ElementsMatch(t, tt.expectedReferenceVars, foundReferenceVars, "Reference variables should match")
			assert.ElementsMatch(t, tt.expectedNormalVars, foundNormalVars, "Normal variables should match")
		})
	}
}

func TestParsing_ClassDeclarations(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		expectedClassName  string
		expectedProperties int
	}{
		{
			name: "Simple class with typed properties",
			input: `<?php
class JUnit {
    private bool $enabled = true;
    private $fp = null;
    private array $suites = [];
}`,
			expectedClassName:  "JUnit",
			expectedProperties: 3,
		},
		{
			name: "Class with complex property default value",
			input: `<?php
class TestClass {
    private array $rootSuite = self::EMPTY_SUITE + ["name" => "php"];
}`,
			expectedClassName:  "TestClass",
			expectedProperties: 1,
		},
		{
			name: "Class with all visibility modifiers",
			input: `<?php
class VisibilityTest {
    public string $publicProp = "public";
    protected int $protectedProp = 42;
    private bool $privateProp = false;
}`,
			expectedClassName:  "VisibilityTest",
			expectedProperties: 3,
		},
		{
			name: "Class with mixed typed and untyped properties",
			input: `<?php
class MixedProps {
    private bool $typed = true;
    private $untyped = null;
    public array $arrayProp = [];
    protected $mixed = "test";
}`,
			expectedClassName:  "MixedProps",
			expectedProperties: 4,
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

			classExpr, ok := exprStmt.Expression.(*ast.ClassExpression)
			assert.True(t, ok, "Expression should be ClassExpression")

			// Check class name
			nameIdent, ok := classExpr.Name.(*ast.IdentifierNode)
			assert.True(t, ok, "Class name should be IdentifierNode")
			assert.Equal(t, tt.expectedClassName, nameIdent.Name)

			// Check properties count
			assert.Len(t, classExpr.Body, tt.expectedProperties)

			// Verify all body items are PropertyDeclaration
			for i, bodyStmt := range classExpr.Body {
				propDecl, ok := bodyStmt.(*ast.PropertyDeclaration)
				assert.True(t, ok, "Class body item %d should be PropertyDeclaration", i)
				
				// Check that visibility is set
				assert.Contains(t, []string{"public", "protected", "private"}, propDecl.Visibility, 
					"Property %d should have valid visibility", i)
				
				// Check that property name is set and doesn't start with $
				assert.NotEmpty(t, propDecl.Name)
				assert.False(t, strings.HasPrefix(propDecl.Name, "$"), 
					"Property name should not start with $")
			}
		})
	}
}

func TestParsing_StaticAccessExpressions(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		expectedClass     string
		expectedProperty  string
	}{
		{
			name:             "Simple static access",
			input:            `<?php $x = self::CONSTANT;`,
			expectedClass:    "self",
			expectedProperty: "CONSTANT",
		},
		{
			name:             "Class name static access", 
			input:            `<?php $x = MyClass::STATIC_VAR;`,
			expectedClass:    "MyClass",
			expectedProperty: "STATIC_VAR",
		},
		{
			name:             "Static access in complex expression",
			input:            `<?php $result = Parent::VALUE + 10;`,
			expectedClass:    "Parent",
			expectedProperty: "VALUE",
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

			assignExpr, ok := exprStmt.Expression.(*ast.AssignmentExpression)
			assert.True(t, ok, "Expression should be AssignmentExpression")

			var staticAccess *ast.StaticAccessExpression
			
			// Handle both direct static access and static access in binary expressions
			if sa, ok := assignExpr.Right.(*ast.StaticAccessExpression); ok {
				staticAccess = sa
			} else if binExpr, ok := assignExpr.Right.(*ast.BinaryExpression); ok {
				sa, ok := binExpr.Left.(*ast.StaticAccessExpression)
				assert.True(t, ok, "Left side of binary expression should be StaticAccessExpression")
				staticAccess = sa
			}

			assert.NotNil(t, staticAccess, "Should find StaticAccessExpression")

			// Check class name
			classIdent, ok := staticAccess.Class.(*ast.IdentifierNode)
			assert.True(t, ok, "Class should be IdentifierNode")
			assert.Equal(t, tt.expectedClass, classIdent.Name)

			// Check property name
			propIdent, ok := staticAccess.Property.(*ast.IdentifierNode)
			assert.True(t, ok, "Property should be IdentifierNode")
			assert.Equal(t, tt.expectedProperty, propIdent.Name)
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

func TestParsing_ClassConstants(t *testing.T) {
	tests := []struct {
		name                 string
		input                string
		expectedClassName    string
		expectedConstGroups  int
		expectedTotalConsts  int
		validateConstants    func(t *testing.T, classExpr *ast.ClassExpression)
	}{
		{
			name: "Private class constant with array value",
			input: `<?php
class JUnit {
    private const EMPTY_SUITE = [
        'test_total' => 0,
        'test_pass' => 0,
        'files' => [],
    ];
}`,
			expectedClassName:   "JUnit",
			expectedConstGroups: 1,
			expectedTotalConsts: 1,
			validateConstants: func(t *testing.T, classExpr *ast.ClassExpression) {
				constGroup, ok := classExpr.Body[0].(*ast.ClassConstantDeclaration)
				assert.True(t, ok, "First body item should be ClassConstantDeclaration")
				assert.Equal(t, "private", constGroup.Visibility)
				assert.Len(t, constGroup.Constants, 1)
				
				constant := constGroup.Constants[0]
				nameIdent, ok := constant.Name.(*ast.IdentifierNode)
				assert.True(t, ok, "Constant name should be IdentifierNode")
				assert.Equal(t, "EMPTY_SUITE", nameIdent.Name)
				
				arrayExpr, ok := constant.Value.(*ast.ArrayExpression)
				assert.True(t, ok, "Constant value should be ArrayExpression")
				assert.Len(t, arrayExpr.Elements, 3)
			},
		},
		{
			name: "Multiple constants in single declaration",
			input: `<?php
class TestConsts {
    const A = 1, B = 2, C = "hello";
}`,
			expectedClassName:   "TestConsts",
			expectedConstGroups: 1,
			expectedTotalConsts: 3,
			validateConstants: func(t *testing.T, classExpr *ast.ClassExpression) {
				constGroup, ok := classExpr.Body[0].(*ast.ClassConstantDeclaration)
				assert.True(t, ok, "First body item should be ClassConstantDeclaration")
				assert.Equal(t, "public", constGroup.Visibility) // Default visibility
				assert.Len(t, constGroup.Constants, 3)
				
				// Check first constant: A = 1
				nameA, ok := constGroup.Constants[0].Name.(*ast.IdentifierNode)
				assert.True(t, ok, "Constant A name should be IdentifierNode")
				assert.Equal(t, "A", nameA.Name)
				
				// Check second constant: B = 2
				nameB, ok := constGroup.Constants[1].Name.(*ast.IdentifierNode)
				assert.True(t, ok, "Constant B name should be IdentifierNode")
				assert.Equal(t, "B", nameB.Name)
				
				// Check third constant: C = "hello"
				nameC, ok := constGroup.Constants[2].Name.(*ast.IdentifierNode)
				assert.True(t, ok, "Constant C name should be IdentifierNode")
				assert.Equal(t, "C", nameC.Name)
			},
		},
		{
			name: "Different visibility modifiers",
			input: `<?php
class VisibilityConsts {
    public const PUB = "public";
    protected const PROT = true;
    private const PRIV = null;
}`,
			expectedClassName:   "VisibilityConsts",
			expectedConstGroups: 3,
			expectedTotalConsts: 3,
			validateConstants: func(t *testing.T, classExpr *ast.ClassExpression) {
				// Check public constant
				pubGroup, ok := classExpr.Body[0].(*ast.ClassConstantDeclaration)
				assert.True(t, ok, "First body item should be ClassConstantDeclaration")
				assert.Equal(t, "public", pubGroup.Visibility)
				assert.Len(t, pubGroup.Constants, 1)
				pubName, ok := pubGroup.Constants[0].Name.(*ast.IdentifierNode)
				assert.True(t, ok)
				assert.Equal(t, "PUB", pubName.Name)
				
				// Check protected constant
				protGroup, ok := classExpr.Body[1].(*ast.ClassConstantDeclaration)
				assert.True(t, ok, "Second body item should be ClassConstantDeclaration")
				assert.Equal(t, "protected", protGroup.Visibility)
				assert.Len(t, protGroup.Constants, 1)
				protName, ok := protGroup.Constants[0].Name.(*ast.IdentifierNode)
				assert.True(t, ok)
				assert.Equal(t, "PROT", protName.Name)
				
				// Check private constant
				privGroup, ok := classExpr.Body[2].(*ast.ClassConstantDeclaration)
				assert.True(t, ok, "Third body item should be ClassConstantDeclaration")
				assert.Equal(t, "private", privGroup.Visibility)
				assert.Len(t, privGroup.Constants, 1)
				privName, ok := privGroup.Constants[0].Name.(*ast.IdentifierNode)
				assert.True(t, ok)
				assert.Equal(t, "PRIV", privName.Name)
			},
		},
		{
			name: "Mixed constants and properties",
			input: `<?php
class Mixed {
    const VERSION = "1.0";
    private $property = "value";
    protected const DEBUG = true;
}`,
			expectedClassName:   "Mixed",
			expectedConstGroups: 2, // VERSION and DEBUG constant groups
			expectedTotalConsts: 2,
			validateConstants: func(t *testing.T, classExpr *ast.ClassExpression) {
				assert.Len(t, classExpr.Body, 3) // 2 const groups + 1 property
				
				// Check first constant
				constGroup1, ok := classExpr.Body[0].(*ast.ClassConstantDeclaration)
				assert.True(t, ok, "First body item should be ClassConstantDeclaration")
				assert.Equal(t, "public", constGroup1.Visibility)
				versionName, ok := constGroup1.Constants[0].Name.(*ast.IdentifierNode)
				assert.True(t, ok)
				assert.Equal(t, "VERSION", versionName.Name)
				
				// Check property
				propDecl, ok := classExpr.Body[1].(*ast.PropertyDeclaration)
				assert.True(t, ok, "Second body item should be PropertyDeclaration")
				assert.Equal(t, "private", propDecl.Visibility)
				assert.Equal(t, "property", propDecl.Name)
				
				// Check second constant
				constGroup2, ok := classExpr.Body[2].(*ast.ClassConstantDeclaration)
				assert.True(t, ok, "Third body item should be ClassConstantDeclaration")
				assert.Equal(t, "protected", constGroup2.Visibility)
				debugName, ok := constGroup2.Constants[0].Name.(*ast.IdentifierNode)
				assert.True(t, ok)
				assert.Equal(t, "DEBUG", debugName.Name)
			},
		},
		{
			name: "Complex constant values",
			input: `<?php
class ComplexConsts {
    const ARRAY_CONST = [1, "two", true, null];
    private const NESTED_ARRAY = [
        'key1' => ['nested' => true],
        'key2' => 42,
    ];
}`,
			expectedClassName:   "ComplexConsts",
			expectedConstGroups: 2,
			expectedTotalConsts: 2,
			validateConstants: func(t *testing.T, classExpr *ast.ClassExpression) {
				// Check first constant with simple array
				constGroup1, ok := classExpr.Body[0].(*ast.ClassConstantDeclaration)
				assert.True(t, ok)
				assert.Equal(t, "public", constGroup1.Visibility)
				arrayName, ok := constGroup1.Constants[0].Name.(*ast.IdentifierNode)
				assert.True(t, ok)
				assert.Equal(t, "ARRAY_CONST", arrayName.Name)
				arrayValue, ok := constGroup1.Constants[0].Value.(*ast.ArrayExpression)
				assert.True(t, ok)
				assert.Len(t, arrayValue.Elements, 4)
				
				// Check second constant with nested array
				constGroup2, ok := classExpr.Body[1].(*ast.ClassConstantDeclaration)
				assert.True(t, ok)
				assert.Equal(t, "private", constGroup2.Visibility)
				nestedName, ok := constGroup2.Constants[0].Name.(*ast.IdentifierNode)
				assert.True(t, ok)
				assert.Equal(t, "NESTED_ARRAY", nestedName.Name)
				nestedValue, ok := constGroup2.Constants[0].Value.(*ast.ArrayExpression)
				assert.True(t, ok)
				assert.Len(t, nestedValue.Elements, 2)
			},
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

			classExpr, ok := exprStmt.Expression.(*ast.ClassExpression)
			assert.True(t, ok, "Expression should be ClassExpression")

			// Check class name
			nameIdent, ok := classExpr.Name.(*ast.IdentifierNode)
			assert.True(t, ok, "Class name should be IdentifierNode")
			assert.Equal(t, tt.expectedClassName, nameIdent.Name)

			// Count constant groups and total constants
			constGroups := 0
			totalConsts := 0
			for _, bodyStmt := range classExpr.Body {
				if constGroup, ok := bodyStmt.(*ast.ClassConstantDeclaration); ok {
					constGroups++
					totalConsts += len(constGroup.Constants)
				}
			}
			
			assert.Equal(t, tt.expectedConstGroups, constGroups, "Expected %d constant groups", tt.expectedConstGroups)
			assert.Equal(t, tt.expectedTotalConsts, totalConsts, "Expected %d total constants", tt.expectedTotalConsts)

			// Run custom validation
			if tt.validateConstants != nil {
				tt.validateConstants(t, classExpr)
			}
		})
	}
}

func TestParsing_ClassMethodsWithVisibility(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, class *ast.ClassExpression)
	}{
		{
			name: "Public constructor with type hints",
			input: `<?php
class JUnit {
    public function __construct(array $env, int $workerID) {
    }
}`,
			validate: func(t *testing.T, class *ast.ClassExpression) {
				if len(class.Body) != 1 {
					t.Errorf("Expected 1 method, got %d", len(class.Body))
					return
				}

				funcDecl, ok := class.Body[0].(*ast.FunctionDeclaration)
				if !ok {
					t.Errorf("Expected FunctionDeclaration, got %T", class.Body[0])
					return
				}

				if funcDecl.Name.(*ast.IdentifierNode).Name != "__construct" {
					t.Errorf("Expected method name '__construct', got '%s'", funcDecl.Name.(*ast.IdentifierNode).Name)
				}

				if funcDecl.Visibility != "public" {
					t.Errorf("Expected visibility 'public', got '%s'", funcDecl.Visibility)
				}

				if len(funcDecl.Parameters) != 2 {
					t.Errorf("Expected 2 parameters, got %d", len(funcDecl.Parameters))
					return
				}

				// Check first parameter: array $env
				param1 := funcDecl.Parameters[0]
				if param1.Name != "$env" {
					t.Errorf("Expected first parameter name '$env', got '%s'", param1.Name)
				}
				if param1.Type == nil || param1.Type.Name != "array" {
					t.Errorf("Expected first parameter type 'array', got %v", param1.Type)
				}

				// Check second parameter: int $workerID  
				param2 := funcDecl.Parameters[1]
				if param2.Name != "$workerID" {
					t.Errorf("Expected second parameter name '$workerID', got '%s'", param2.Name)
				}
				if param2.Type == nil || param2.Type.Name != "int" {
					t.Errorf("Expected second parameter type 'int', got %v", param2.Type)
				}
			},
		},
		{
			name: "All visibility modifiers",
			input: `<?php
class Test {
    private function privateMethod(string $x) {
    }
    
    protected function protectedMethod(int $y) {
    }
    
    public function publicMethod(array $z) {
    }
}`,
			validate: func(t *testing.T, class *ast.ClassExpression) {
				if len(class.Body) != 3 {
					t.Errorf("Expected 3 methods, got %d", len(class.Body))
					return
				}

				expectedMethods := []struct {
					name       string
					visibility string
					paramType  string
					paramName  string
				}{
					{"privateMethod", "private", "string", "$x"},
					{"protectedMethod", "protected", "int", "$y"},
					{"publicMethod", "public", "array", "$z"},
				}

				for i, expected := range expectedMethods {
					funcDecl, ok := class.Body[i].(*ast.FunctionDeclaration)
					if !ok {
						t.Errorf("Expected FunctionDeclaration at index %d, got %T", i, class.Body[i])
						continue
					}

					if funcDecl.Name.(*ast.IdentifierNode).Name != expected.name {
						t.Errorf("Method %d: expected name '%s', got '%s'", i, expected.name, funcDecl.Name.(*ast.IdentifierNode).Name)
					}

					if funcDecl.Visibility != expected.visibility {
						t.Errorf("Method %d: expected visibility '%s', got '%s'", i, expected.visibility, funcDecl.Visibility)
					}

					if len(funcDecl.Parameters) != 1 {
						t.Errorf("Method %d: expected 1 parameter, got %d", i, len(funcDecl.Parameters))
						continue
					}

					param := funcDecl.Parameters[0]
					if param.Name != expected.paramName {
						t.Errorf("Method %d: expected parameter name '%s', got '%s'", i, expected.paramName, param.Name)
					}
					if param.Type == nil || param.Type.Name != expected.paramType {
						t.Errorf("Method %d: expected parameter type '%s', got %v", i, expected.paramType, param.Type)
					}
				}
			},
		},
		{
			name: "Method without visibility (defaults)",
			input: `<?php
class Test {
    function defaultMethod(bool $flag) {
    }
}`,
			validate: func(t *testing.T, class *ast.ClassExpression) {
				if len(class.Body) != 1 {
					t.Errorf("Expected 1 method, got %d", len(class.Body))
					return
				}

				funcDecl, ok := class.Body[0].(*ast.FunctionDeclaration)
				if !ok {
					t.Errorf("Expected FunctionDeclaration, got %T", class.Body[0])
					return
				}

				if funcDecl.Name.(*ast.IdentifierNode).Name != "defaultMethod" {
					t.Errorf("Expected method name 'defaultMethod', got '%s'", funcDecl.Name.(*ast.IdentifierNode).Name)
				}

				// Method without explicit visibility should have empty visibility string
				if funcDecl.Visibility != "" {
					t.Errorf("Expected empty visibility for method without modifier, got '%s'", funcDecl.Visibility)
				}
			},
		},
		{
			name: "Complex parameter types",
			input: `<?php
class Complex {
    public function handle(array $config, callable $callback, string $message = "default") {
    }
}`,
			validate: func(t *testing.T, class *ast.ClassExpression) {
				if len(class.Body) != 1 {
					t.Errorf("Expected 1 method, got %d", len(class.Body))
					return
				}

				funcDecl, ok := class.Body[0].(*ast.FunctionDeclaration)
				if !ok {
					t.Errorf("Expected FunctionDeclaration, got %T", class.Body[0])
					return
				}

				if len(funcDecl.Parameters) != 3 {
					t.Errorf("Expected 3 parameters, got %d", len(funcDecl.Parameters))
					return
				}

				expectedParams := []struct {
					name     string
					typeName string
					hasDefault bool
				}{
					{"$config", "array", false},
					{"$callback", "callable", false},
					{"$message", "string", true},
				}

				for i, expected := range expectedParams {
					param := funcDecl.Parameters[i]
					if param.Name != expected.name {
						t.Errorf("Parameter %d: expected name '%s', got '%s'", i, expected.name, param.Name)
					}
					if param.Type == nil || param.Type.Name != expected.typeName {
						t.Errorf("Parameter %d: expected type '%s', got %v", i, expected.typeName, param.Type)
					}
					hasDefault := param.DefaultValue != nil
					if hasDefault != expected.hasDefault {
						t.Errorf("Parameter %d: expected default value present=%t, got=%t", i, expected.hasDefault, hasDefault)
					}
				}
			},
		},
		{
			name: "Mixed class members (constants, properties, methods)",
			input: `<?php
class Mixed {
    const STATUS = "active";
    
    private $property = "value";
    
    public function method(int $x) {
        return $x * 2;
    }
}`,
			validate: func(t *testing.T, class *ast.ClassExpression) {
				if len(class.Body) != 3 {
					t.Errorf("Expected 3 class members, got %d", len(class.Body))
					return
				}

				// Check that we have a method with visibility
				var methodFound bool
				for _, member := range class.Body {
					if funcDecl, ok := member.(*ast.FunctionDeclaration); ok {
						methodFound = true
						if funcDecl.Visibility != "public" {
							t.Errorf("Expected method visibility 'public', got '%s'", funcDecl.Visibility)
						}
						if funcDecl.Name.(*ast.IdentifierNode).Name != "method" {
							t.Errorf("Expected method name 'method', got '%s'", funcDecl.Name.(*ast.IdentifierNode).Name)
						}
					}
				}

				if !methodFound {
					t.Error("Expected to find a method declaration")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Body) != 1 {
				t.Errorf("Expected 1 statement, got %d", len(program.Body))
				return
			}

			stmt, ok := program.Body[0].(*ast.ExpressionStatement)
			if !ok {
				t.Errorf("Expected ExpressionStatement, got %T", program.Body[0])
				return
			}

			class, ok := stmt.Expression.(*ast.ClassExpression)
			if !ok {
				t.Errorf("Expected ClassDeclaration, got %T", stmt.Expression)
				return
			}

			tt.validate(t, class)
		})
	}
}

func TestParsing_TypedReferenceParameters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, program *ast.Program)
	}{
		{
			name: "Single typed reference parameter",
			input: `<?php
function test(array &$x) {
}`,
			validate: func(t *testing.T, program *ast.Program) {
				if len(program.Body) != 1 {
					t.Errorf("Expected 1 statement, got %d", len(program.Body))
					return
				}

				funcDecl, ok := program.Body[0].(*ast.FunctionDeclaration)
				if !ok {
					t.Errorf("Expected FunctionDeclaration, got %T", program.Body[0])
					return
				}

				if len(funcDecl.Parameters) != 1 {
					t.Errorf("Expected 1 parameter, got %d", len(funcDecl.Parameters))
					return
				}

				param := funcDecl.Parameters[0]
				if param.Name != "$x" {
					t.Errorf("Expected parameter name '$x', got '%s'", param.Name)
				}
				if param.Type == nil || param.Type.Name != "array" {
					t.Errorf("Expected parameter type 'array', got %v", param.Type)
				}
				if !param.ByReference {
					t.Error("Expected parameter to be by reference")
				}
			},
		},
		{
			name: "Multiple typed reference parameters",
			input: `<?php
function mergeSuites(array &$dest, array $source): void {
}`,
			validate: func(t *testing.T, program *ast.Program) {
				funcDecl, ok := program.Body[0].(*ast.FunctionDeclaration)
				if !ok {
					t.Errorf("Expected FunctionDeclaration, got %T", program.Body[0])
					return
				}

				if len(funcDecl.Parameters) != 2 {
					t.Errorf("Expected 2 parameters, got %d", len(funcDecl.Parameters))
					return
				}

				// Check first parameter (reference)
				param1 := funcDecl.Parameters[0]
				if param1.Name != "$dest" {
					t.Errorf("Expected first parameter name '$dest', got '%s'", param1.Name)
				}
				if param1.Type == nil || param1.Type.Name != "array" {
					t.Errorf("Expected first parameter type 'array', got %v", param1.Type)
				}
				if !param1.ByReference {
					t.Error("Expected first parameter to be by reference")
				}

				// Check second parameter (non-reference)
				param2 := funcDecl.Parameters[1]
				if param2.Name != "$source" {
					t.Errorf("Expected second parameter name '$source', got '%s'", param2.Name)
				}
				if param2.Type == nil || param2.Type.Name != "array" {
					t.Errorf("Expected second parameter type 'array', got %v", param2.Type)
				}
				if param2.ByReference {
					t.Error("Expected second parameter to not be by reference")
				}

				// Check return type
				if funcDecl.ReturnType == nil || funcDecl.ReturnType.Name != "void" {
					t.Errorf("Expected return type 'void', got %v", funcDecl.ReturnType)
				}
			},
		},
		{
			name: "Mixed reference and non-reference parameters",
			input: `<?php
function test(&$a, array &$b, string $c, callable &$d): bool {
}`,
			validate: func(t *testing.T, program *ast.Program) {
				funcDecl, ok := program.Body[0].(*ast.FunctionDeclaration)
				if !ok {
					t.Errorf("Expected FunctionDeclaration, got %T", program.Body[0])
					return
				}

				if len(funcDecl.Parameters) != 4 {
					t.Errorf("Expected 4 parameters, got %d", len(funcDecl.Parameters))
					return
				}

				expectedParams := []struct {
					name        string
					typeName    string
					hasType     bool
					byReference bool
				}{
					{"$a", "", false, true},           // &$a (no type)
					{"$b", "array", true, true},       // array &$b
					{"$c", "string", true, false},     // string $c
					{"$d", "callable", true, true},    // callable &$d
				}

				for i, expected := range expectedParams {
					param := funcDecl.Parameters[i]
					if param.Name != expected.name {
						t.Errorf("Parameter %d: expected name '%s', got '%s'", i, expected.name, param.Name)
					}

					hasType := param.Type != nil
					if hasType != expected.hasType {
						t.Errorf("Parameter %d: expected hasType=%t, got=%t", i, expected.hasType, hasType)
					}

					if hasType && param.Type.Name != expected.typeName {
						t.Errorf("Parameter %d: expected type '%s', got '%s'", i, expected.typeName, param.Type.Name)
					}

					if param.ByReference != expected.byReference {
						t.Errorf("Parameter %d: expected byReference=%t, got=%t", i, expected.byReference, param.ByReference)
					}
				}
			},
		},
		{
			name: "Nullable typed reference parameters",
			input: `<?php
function test(?array &$a, ?string &$b): void {
}`,
			validate: func(t *testing.T, program *ast.Program) {
				funcDecl, ok := program.Body[0].(*ast.FunctionDeclaration)
				if !ok {
					t.Errorf("Expected FunctionDeclaration, got %T", program.Body[0])
					return
				}

				if len(funcDecl.Parameters) != 2 {
					t.Errorf("Expected 2 parameters, got %d", len(funcDecl.Parameters))
					return
				}

				// Check first parameter: ?array &$a
				param1 := funcDecl.Parameters[0]
				if param1.Name != "$a" {
					t.Errorf("Expected first parameter name '$a', got '%s'", param1.Name)
				}
				if param1.Type == nil {
					t.Error("Expected first parameter to have type")
				} else {
					if param1.Type.Name != "array" {
						t.Errorf("Expected first parameter type 'array', got '%s'", param1.Type.Name)
					}
					if !param1.Type.Nullable {
						t.Error("Expected first parameter type to be nullable")
					}
				}
				if !param1.ByReference {
					t.Error("Expected first parameter to be by reference")
				}

				// Check second parameter: ?string &$b
				param2 := funcDecl.Parameters[1]
				if param2.Name != "$b" {
					t.Errorf("Expected second parameter name '$b', got '%s'", param2.Name)
				}
				if param2.Type == nil {
					t.Error("Expected second parameter to have type")
				} else {
					if param2.Type.Name != "string" {
						t.Errorf("Expected second parameter type 'string', got '%s'", param2.Type.Name)
					}
					if !param2.Type.Nullable {
						t.Error("Expected second parameter type to be nullable")
					}
				}
				if !param2.ByReference {
					t.Error("Expected second parameter to be by reference")
				}
			},
		},
		{
			name: "Union type reference parameters",
			input: `<?php
function test(array|string &$x): void {
}`,
			validate: func(t *testing.T, program *ast.Program) {
				funcDecl, ok := program.Body[0].(*ast.FunctionDeclaration)
				if !ok {
					t.Errorf("Expected FunctionDeclaration, got %T", program.Body[0])
					return
				}

				if len(funcDecl.Parameters) != 1 {
					t.Errorf("Expected 1 parameter, got %d", len(funcDecl.Parameters))
					return
				}

				param := funcDecl.Parameters[0]
				if param.Name != "$x" {
					t.Errorf("Expected parameter name '$x', got '%s'", param.Name)
				}

				if param.Type == nil {
					t.Error("Expected parameter to have type")
					return
				}

				// Check if it's a union type
				if param.Type.UnionTypes == nil || len(param.Type.UnionTypes) != 2 {
					t.Errorf("Expected union type with 2 types, got %v", param.Type)
					return
				}

				// Check union type components
				if param.Type.UnionTypes[0].Name != "array" {
					t.Errorf("Expected first union type 'array', got '%s'", param.Type.UnionTypes[0].Name)
				}
				if param.Type.UnionTypes[1].Name != "string" {
					t.Errorf("Expected second union type 'string', got '%s'", param.Type.UnionTypes[1].Name)
				}

				if !param.ByReference {
					t.Error("Expected parameter to be by reference")
				}
			},
		},
		{
			name: "Class method with typed reference parameters",
			input: `<?php
class TestClass {
    private function mergeSuites(array &$dest, array $source): void {
    }
}`,
			validate: func(t *testing.T, program *ast.Program) {
				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				if !ok {
					t.Errorf("Expected ExpressionStatement, got %T", program.Body[0])
					return
				}

				class, ok := stmt.Expression.(*ast.ClassExpression)
				if !ok {
					t.Errorf("Expected ClassExpression, got %T", stmt.Expression)
					return
				}

				if len(class.Body) != 1 {
					t.Errorf("Expected 1 method, got %d", len(class.Body))
					return
				}

				funcDecl, ok := class.Body[0].(*ast.FunctionDeclaration)
				if !ok {
					t.Errorf("Expected FunctionDeclaration, got %T", class.Body[0])
					return
				}

				if funcDecl.Visibility != "private" {
					t.Errorf("Expected method visibility 'private', got '%s'", funcDecl.Visibility)
				}

				if len(funcDecl.Parameters) != 2 {
					t.Errorf("Expected 2 parameters, got %d", len(funcDecl.Parameters))
					return
				}

				// Verify the original failing case: array &$dest, array $source
				param1 := funcDecl.Parameters[0]
				if param1.Name != "$dest" || param1.Type.Name != "array" || !param1.ByReference {
					t.Errorf("First parameter incorrect: name=%s, type=%v, byRef=%t", param1.Name, param1.Type, param1.ByReference)
				}

				param2 := funcDecl.Parameters[1]
				if param2.Name != "$source" || param2.Type.Name != "array" || param2.ByReference {
					t.Errorf("Second parameter incorrect: name=%s, type=%v, byRef=%t", param2.Name, param2.Type, param2.ByReference)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			tt.validate(t, program)
		})
	}
}

// TestParsing_TryCatchWithStatements tests parsing try-catch blocks followed by statements
func TestParsing_TryCatchWithStatements(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, program *ast.Program)
	}{
		{
			name: "try-catch with assignment after",
			input: `<?php
try {
} catch (Exception $ex) {
}
$tested = $test->getName();`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 2)

				// Check try-catch statement
				tryStmt, ok := program.Body[0].(*ast.TryStatement)
				assert.True(t, ok, "First statement should be TryStatement")
				assert.Len(t, tryStmt.Body, 0, "Try block should be empty")
				assert.Len(t, tryStmt.CatchClauses, 1, "Should have one catch clause")

				catch := tryStmt.CatchClauses[0]
				assert.Len(t, catch.Types, 1, "Catch should have one exception type")
				assert.Equal(t, "Exception", catch.Types[0].(*ast.IdentifierNode).Name)
				assert.Equal(t, "$ex", catch.Parameter.(*ast.Variable).Name)
				assert.Len(t, catch.Body, 0, "Catch block should be empty")

				// Check assignment statement
				exprStmt, ok := program.Body[1].(*ast.ExpressionStatement)
				assert.True(t, ok, "Second statement should be ExpressionStatement")

				assignment, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				assert.True(t, ok, "Expression should be AssignmentExpression")
				assert.Equal(t, "=", assignment.Operator)

				// Check left side ($tested)
				leftVar, ok := assignment.Left.(*ast.Variable)
				assert.True(t, ok, "Left side should be Variable")
				assert.Equal(t, "$tested", leftVar.Name)

				// Check right side ($test->getName())
				callExpr, ok := assignment.Right.(*ast.CallExpression)
				assert.True(t, ok, "Right side should be CallExpression")

				propAccess, ok := callExpr.Callee.(*ast.PropertyAccessExpression)
				assert.True(t, ok, "Callee should be PropertyAccessExpression")

				objVar, ok := propAccess.Object.(*ast.Variable)
				assert.True(t, ok, "Object should be Variable")
				assert.Equal(t, "$test", objVar.Name)

				assert.NotNil(t, propAccess.Property, "Property should not be nil")
				assert.Equal(t, "getName", propAccess.Property.Name)
			},
		},
		{
			name: "try-catch with statements in blocks and after",
			input: `<?php
try {
    $x = 1;
} catch (Exception $e) {
    echo "Error";
}
$result = $obj->process();`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 2)

				// Check try-catch statement
				tryStmt, ok := program.Body[0].(*ast.TryStatement)
				assert.True(t, ok, "First statement should be TryStatement")
				assert.Len(t, tryStmt.Body, 1, "Try block should have one statement")
				assert.Len(t, tryStmt.CatchClauses, 1, "Should have one catch clause")

				// Check try block content
				_, ok = tryStmt.Body[0].(*ast.ExpressionStatement)
				assert.True(t, ok, "Try statement should be assignment")

				// Check catch block content  
				catch := tryStmt.CatchClauses[0]
				assert.Len(t, catch.Body, 1, "Catch block should have one statement")
				echoStmt, ok := catch.Body[0].(*ast.EchoStatement)
				assert.True(t, ok, "Catch statement should be echo")
				assert.Len(t, echoStmt.Arguments, 1)

				// Check statement after try-catch
				exprStmt, ok := program.Body[1].(*ast.ExpressionStatement)
				assert.True(t, ok, "Second statement should be ExpressionStatement")
				assignment, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				assert.True(t, ok, "Should be assignment after try-catch")
				assert.Equal(t, "$result", assignment.Left.(*ast.Variable).Name)
			},
		},
		{
			name: "try with multiple catch clauses and statements after",
			input: `<?php
try {
    throw new Exception();
} catch (InvalidArgumentException $e) {
    return false;
} catch (Exception $e) {
    log($e);
}
$success = true;
return $success;`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 3)

				// Check try-catch statement
				tryStmt, ok := program.Body[0].(*ast.TryStatement)
				assert.True(t, ok, "First statement should be TryStatement")
				assert.Len(t, tryStmt.CatchClauses, 2, "Should have two catch clauses")

				// Check first catch clause
				catch1 := tryStmt.CatchClauses[0]
				assert.Equal(t, "InvalidArgumentException", catch1.Types[0].(*ast.IdentifierNode).Name)
				assert.Equal(t, "$e", catch1.Parameter.(*ast.Variable).Name)

				// Check second catch clause
				catch2 := tryStmt.CatchClauses[1]
				assert.Equal(t, "Exception", catch2.Types[0].(*ast.IdentifierNode).Name)
				assert.Equal(t, "$e", catch2.Parameter.(*ast.Variable).Name)

				// Check statements after try-catch
				exprStmt1, ok := program.Body[1].(*ast.ExpressionStatement)
				assert.True(t, ok, "Second statement should be assignment")
				assignment := exprStmt1.Expression.(*ast.AssignmentExpression)
				assert.Equal(t, "$success", assignment.Left.(*ast.Variable).Name)

				returnStmt, ok := program.Body[2].(*ast.ReturnStatement)
				assert.True(t, ok, "Third statement should be return")
				returnVar, ok := returnStmt.Argument.(*ast.Variable)
				assert.True(t, ok, "Return value should be variable")
				assert.Equal(t, "$success", returnVar.Name)
			},
		},
		{
			name: "nested try-catch with statements",
			input: `<?php
try {
    try {
        $inner = doSomething();
    } catch (InnerException $e) {
        handle($e);
    }
} catch (OuterException $e) {
    cleanup();
}
$final = complete();`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 2)

				// Check outer try-catch
				outerTry, ok := program.Body[0].(*ast.TryStatement)
				assert.True(t, ok, "First statement should be TryStatement")
				assert.Len(t, outerTry.Body, 1, "Outer try should have one statement")
				assert.Len(t, outerTry.CatchClauses, 1, "Outer try should have one catch")

				// Check inner try-catch exists within outer try
				innerTry, ok := outerTry.Body[0].(*ast.TryStatement)
				assert.True(t, ok, "Inner statement should be TryStatement")
				assert.Len(t, innerTry.CatchClauses, 1, "Inner try should have one catch")

				// Check final statement after nested try-catch
				finalStmt, ok := program.Body[1].(*ast.ExpressionStatement)
				assert.True(t, ok, "Final statement should be assignment")
				assignment := finalStmt.Expression.(*ast.AssignmentExpression)
				assert.Equal(t, "$final", assignment.Left.(*ast.Variable).Name)
			},
		},
		{
			name: "empty try-catch followed by multiple statements", 
			input: `<?php
try {
} catch (Exception $ex) {
}
$a = 1;
$b = 2;
echo $a + $b;`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 4)

				// Check try-catch
				_, ok := program.Body[0].(*ast.TryStatement)
				assert.True(t, ok, "First statement should be TryStatement")

				// Check all following statements parse correctly
				for i := 1; i < 4; i++ {
					stmt := program.Body[i]
					assert.NotNil(t, stmt, "Statement %d should not be nil", i)
					
					if i < 3 {
						// First two should be assignments
						exprStmt, ok := stmt.(*ast.ExpressionStatement)
						assert.True(t, ok, "Statement %d should be ExpressionStatement", i)
						_, ok = exprStmt.Expression.(*ast.AssignmentExpression)
						assert.True(t, ok, "Statement %d should be assignment", i)
					} else {
						// Last should be echo
						_, ok := stmt.(*ast.EchoStatement)
						assert.True(t, ok, "Statement %d should be EchoStatement", i)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			assert.NotNil(t, program)
			tt.validate(t, program)
		})
	}
}

func TestParsing_IncludeAndRequireStatements(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, program *ast.Program)
	}{
		{
			name:  "require statement",
			input: `<?php require 'config.php';`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				assert.True(t, ok, "Statement should be ExpressionStatement")

				includeExpr, ok := exprStmt.Expression.(*ast.IncludeOrEvalExpression)
				assert.True(t, ok, "Expression should be IncludeOrEvalExpression")
				assert.Equal(t, lexer.T_REQUIRE, includeExpr.Type)
				
				stringLit, ok := includeExpr.Expr.(*ast.StringLiteral)
				assert.True(t, ok, "Included file should be StringLiteral")
				assert.Equal(t, "config.php", stringLit.Value)
			},
		},
		{
			name:  "require_once statement",
			input: `<?php require_once 'utils.php';`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				assert.True(t, ok, "Statement should be ExpressionStatement")

				includeExpr, ok := exprStmt.Expression.(*ast.IncludeOrEvalExpression)
				assert.True(t, ok, "Expression should be IncludeOrEvalExpression")
				assert.Equal(t, lexer.T_REQUIRE_ONCE, includeExpr.Type)
				
				stringLit, ok := includeExpr.Expr.(*ast.StringLiteral)
				assert.True(t, ok, "Included file should be StringLiteral")
				assert.Equal(t, "utils.php", stringLit.Value)
			},
		},
		{
			name:  "include and include_once statements",
			input: `<?php include 'header.php'; include_once 'footer.php';`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 2)

				// Check include
				exprStmt1, ok := program.Body[0].(*ast.ExpressionStatement)
				assert.True(t, ok, "First statement should be ExpressionStatement")
				includeExpr1, ok := exprStmt1.Expression.(*ast.IncludeOrEvalExpression)
				assert.True(t, ok, "Expression should be IncludeOrEvalExpression")
				assert.Equal(t, lexer.T_INCLUDE, includeExpr1.Type)
				
				// Check include_once
				exprStmt2, ok := program.Body[1].(*ast.ExpressionStatement)
				assert.True(t, ok, "Second statement should be ExpressionStatement")
				includeExpr2, ok := exprStmt2.Expression.(*ast.IncludeOrEvalExpression)
				assert.True(t, ok, "Expression should be IncludeOrEvalExpression")
				assert.Equal(t, lexer.T_INCLUDE_ONCE, includeExpr2.Type)
			},
		},
		{
			name:  "include with variable expression",
			input: `<?php require $config_file;`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				assert.True(t, ok, "Statement should be ExpressionStatement")

				includeExpr, ok := exprStmt.Expression.(*ast.IncludeOrEvalExpression)
				assert.True(t, ok, "Expression should be IncludeOrEvalExpression")
				assert.Equal(t, lexer.T_REQUIRE, includeExpr.Type)
				
				variable, ok := includeExpr.Expr.(*ast.Variable)
				assert.True(t, ok, "Included expression should be Variable")
				assert.Equal(t, "$config_file", variable.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			assert.NotNil(t, program)
			tt.validate(t, program)
		})
	}
}

func TestParsing_StaticDeclarations(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, program *ast.Program)
	}{
		{
			name:  "static variable declaration",
			input: `<?php static $counter = 0;`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				
				staticStmt, ok := program.Body[0].(*ast.StaticStatement)
				assert.True(t, ok, "Statement should be StaticStatement")
				assert.Len(t, staticStmt.Variables, 1, "Should have one static variable")
				
				staticVar := staticStmt.Variables[0]
				variable, ok := staticVar.Variable.(*ast.Variable)
				assert.True(t, ok, "Variable should be Variable type")
				assert.Equal(t, "$counter", variable.Name)
				
				defaultValue, ok := staticVar.DefaultValue.(*ast.NumberLiteral)
				assert.True(t, ok, "Default value should be NumberLiteral")
				assert.Equal(t, "0", defaultValue.Value)
			},
		},
		{
			name:  "static as identifier in expression context",
			input: `<?php $x = static;`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				
				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				assert.True(t, ok, "Statement should be ExpressionStatement")
				
				assignExpr, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				assert.True(t, ok, "Expression should be AssignmentExpression")
				
				staticIdent, ok := assignExpr.Right.(*ast.IdentifierNode)
				assert.True(t, ok, "Right should be IdentifierNode")
				assert.Equal(t, "static", staticIdent.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			assert.NotNil(t, program)
			tt.validate(t, program)
		})
	}
}

func TestParsing_AbstractKeyword(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, program *ast.Program)
	}{
		{
			name:  "abstract class declaration",
			input: `<?php abstract class BaseClass {}`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 2)
				
				// First should be abstract identifier
				exprStmt1, ok := program.Body[0].(*ast.ExpressionStatement)
				assert.True(t, ok, "First statement should be ExpressionStatement")
				abstractIdent, ok := exprStmt1.Expression.(*ast.IdentifierNode)
				assert.True(t, ok, "Expression should be IdentifierNode")
				assert.Equal(t, "abstract", abstractIdent.Name)
				
				// Second should be class declaration
				exprStmt2, ok := program.Body[1].(*ast.ExpressionStatement)
				assert.True(t, ok, "Second statement should be ExpressionStatement")
				classExpr, ok := exprStmt2.Expression.(*ast.ClassExpression)
				assert.True(t, ok, "Expression should be ClassExpression")
				nameIdent, ok := classExpr.Name.(*ast.IdentifierNode)
				assert.True(t, ok, "Class name should be IdentifierNode")
				assert.Equal(t, "BaseClass", nameIdent.Name)
			},
		},
		{
			name:  "abstract as identifier in expression",
			input: `<?php $type = abstract;`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				
				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				assert.True(t, ok, "Statement should be ExpressionStatement")
				
				assignExpr, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				assert.True(t, ok, "Expression should be AssignmentExpression")
				
				abstractIdent, ok := assignExpr.Right.(*ast.IdentifierNode)
				assert.True(t, ok, "Right should be IdentifierNode")
				assert.Equal(t, "abstract", abstractIdent.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			assert.NotNil(t, program)
			tt.validate(t, program)
		})
	}
}

func TestParsing_NamespaceSeparator(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, program *ast.Program)
	}{
		{
			name:  "fully qualified namespace call",
			input: `<?php \DateTime\createFromFormat();`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				
				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				assert.True(t, ok, "Statement should be ExpressionStatement")
				
				callExpr, ok := exprStmt.Expression.(*ast.CallExpression)
				assert.True(t, ok, "Expression should be CallExpression")
				
				// The callee should be a namespace expression
				namespaceExpr, ok := callExpr.Callee.(*ast.NamespaceExpression)
				assert.True(t, ok, "Callee should be NamespaceExpression")
				
				// Check the namespace name contains the full path
				nameIdent, ok := namespaceExpr.Name.(*ast.IdentifierNode)
				assert.True(t, ok, "Name should be IdentifierNode")
				assert.Equal(t, "DateTime\\createFromFormat", nameIdent.Name)
			},
		},
		{
			name:  "single leading backslash",
			input: `<?php \test();`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				
				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				assert.True(t, ok, "Statement should be ExpressionStatement")
				
				callExpr, ok := exprStmt.Expression.(*ast.CallExpression)
				assert.True(t, ok, "Expression should be CallExpression")
				
				namespaceExpr, ok := callExpr.Callee.(*ast.NamespaceExpression)
				assert.True(t, ok, "Callee should be NamespaceExpression")
				
				nameIdent, ok := namespaceExpr.Name.(*ast.IdentifierNode)
				assert.True(t, ok, "Name should be IdentifierNode")
				assert.Equal(t, "test", nameIdent.Name)
			},
		},
		{
			name:  "bare backslash",
			input: `<?php $x = \;`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				
				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				assert.True(t, ok, "Statement should be ExpressionStatement")
				
				assignExpr, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				assert.True(t, ok, "Expression should be AssignmentExpression")
				
				namespaceExpr, ok := assignExpr.Right.(*ast.NamespaceExpression)
				assert.True(t, ok, "Right should be NamespaceExpression")
				assert.Nil(t, namespaceExpr.Name, "Bare backslash should have nil name")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			assert.NotNil(t, program)
			tt.validate(t, program)
		})
	}
}

func TestParsing_Attributes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, program *ast.Program)
	}{
		{
			name:  "Simple attribute without parameters",
			input: `<?php #[Route]`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				
				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				assert.True(t, ok, "Expected ExpressionStatement")
				
				attrGroup, ok := stmt.Expression.(*ast.AttributeGroup)
				assert.True(t, ok, "Expected AttributeGroup")
				require.Len(t, attrGroup.Attributes, 1)
				attr := attrGroup.Attributes[0]
				assert.Equal(t, "Route", attr.Name.Name)
				assert.Empty(t, attr.Arguments)
			},
		},
		{
			name:  "Attribute with string parameter",
			input: `<?php #[Route("/api/users")]`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				
				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				assert.True(t, ok, "Expected ExpressionStatement")
				
				attrGroup, ok := stmt.Expression.(*ast.AttributeGroup)
				assert.True(t, ok, "Expected AttributeGroup")
				require.Len(t, attrGroup.Attributes, 1)
				attr := attrGroup.Attributes[0]
				assert.Equal(t, "Route", attr.Name.Name)
				assert.Len(t, attr.Arguments, 1)
				
				strLit, ok := attr.Arguments[0].(*ast.StringLiteral)
				assert.True(t, ok, "Expected StringLiteral")
				assert.Equal(t, "/api/users", strLit.Value)
			},
		},
		{
			name:  "Attribute with multiple parameters",
			input: `<?php #[Route("/api/users", "GET")]`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				
				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				assert.True(t, ok, "Expected ExpressionStatement")
				
				attrGroup, ok := stmt.Expression.(*ast.AttributeGroup)
				assert.True(t, ok, "Expected AttributeGroup")
				require.Len(t, attrGroup.Attributes, 1)
				attr := attrGroup.Attributes[0]
				assert.Equal(t, "Route", attr.Name.Name)
				assert.Len(t, attr.Arguments, 2)
				
				pathArg, ok := attr.Arguments[0].(*ast.StringLiteral)
				assert.True(t, ok, "Expected StringLiteral for path")
				assert.Equal(t, "/api/users", pathArg.Value)
				
				methodArg, ok := attr.Arguments[1].(*ast.StringLiteral)
				assert.True(t, ok, "Expected StringLiteral for method")
				assert.Equal(t, "GET", methodArg.Value)
			},
		},
		{
			name:  "Attribute with complex arguments",
			input: `<?php #[Cache(timeout: 3600)]`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				
				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				assert.True(t, ok, "Expected ExpressionStatement")
				
				attrGroup, ok := stmt.Expression.(*ast.AttributeGroup)
				assert.True(t, ok, "Expected AttributeGroup")
				require.Len(t, attrGroup.Attributes, 1)
				attr := attrGroup.Attributes[0]
				assert.Equal(t, "Cache", attr.Name.Name)
				assert.Len(t, attr.Arguments, 1)
				
				namedArg, ok := attr.Arguments[0].(*ast.NamedArgument)
				assert.True(t, ok, "Expected NamedArgument")
				assert.Equal(t, "timeout", namedArg.Name.Name)
				
				numLit, ok := namedArg.Value.(*ast.NumberLiteral)
				assert.True(t, ok, "Expected NumberLiteral")
				assert.Equal(t, "3600", numLit.Value)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			
			checkParserErrors(t, p)

			assert.NotNil(t, program)
			tt.validate(t, program)
		})
	}
}

func TestParsing_IntersectionTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, program *ast.Program)
	}{
		{
			name:  "Function with intersection type return",
			input: `<?php function test(): Type1&Type2 {}`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				
				funcDecl, ok := program.Body[0].(*ast.FunctionDeclaration)
				assert.True(t, ok, "Expected FunctionDeclaration")
				assert.Equal(t, "test", funcDecl.Name.(*ast.IdentifierNode).Name)
				
				assert.NotNil(t, funcDecl.ReturnType)
				assert.Equal(t, ast.ASTTypeIntersection, funcDecl.ReturnType.GetKind())
				assert.Len(t, funcDecl.ReturnType.IntersectionTypes, 2)
				assert.Equal(t, "Type1", funcDecl.ReturnType.IntersectionTypes[0].Name)
				assert.Equal(t, "Type2", funcDecl.ReturnType.IntersectionTypes[1].Name)
			},
		},
		{
			name:  "Class method with intersection type return",
			input: `<?php class Test { public function method(): Interface1&Interface2 {} }`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				
				classExpr, ok := program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.ClassExpression)
				assert.True(t, ok, "Expected ClassExpression")
				assert.Equal(t, "Test", classExpr.Name.(*ast.IdentifierNode).Name)
				assert.Len(t, classExpr.Body, 1)
				
				methodDecl, ok := classExpr.Body[0].(*ast.FunctionDeclaration)
				assert.True(t, ok, "Expected FunctionDeclaration")
				assert.Equal(t, "method", methodDecl.Name.(*ast.IdentifierNode).Name)
				assert.Equal(t, "public", methodDecl.Visibility)
				
				assert.NotNil(t, methodDecl.ReturnType)
				assert.Equal(t, ast.ASTTypeIntersection, methodDecl.ReturnType.GetKind())
				assert.Len(t, methodDecl.ReturnType.IntersectionTypes, 2)
				assert.Equal(t, "Interface1", methodDecl.ReturnType.IntersectionTypes[0].Name)
				assert.Equal(t, "Interface2", methodDecl.ReturnType.IntersectionTypes[1].Name)
			},
		},
		{
			name:  "Multiple intersection types",
			input: `<?php function complex(): A&B&C {}`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				
				funcDecl, ok := program.Body[0].(*ast.FunctionDeclaration)
				assert.True(t, ok, "Expected FunctionDeclaration")
				assert.Equal(t, "complex", funcDecl.Name.(*ast.IdentifierNode).Name)
				
				assert.NotNil(t, funcDecl.ReturnType)
				assert.Equal(t, ast.ASTTypeIntersection, funcDecl.ReturnType.GetKind())
				assert.Len(t, funcDecl.ReturnType.IntersectionTypes, 3)
				assert.Equal(t, "A", funcDecl.ReturnType.IntersectionTypes[0].Name)
				assert.Equal(t, "B", funcDecl.ReturnType.IntersectionTypes[1].Name)
				assert.Equal(t, "C", funcDecl.ReturnType.IntersectionTypes[2].Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			
			checkParserErrors(t, p)

			assert.NotNil(t, program)
			tt.validate(t, program)
		})
	}
}

func TestParsing_FirstClassCallable(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, program *ast.Program)
	}{
		{
			name:  "Simple function first-class callable",
			input: `<?php $func = strlen(...);`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				
				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				assert.True(t, ok, "Expected ExpressionStatement")
				
				assign, ok := stmt.Expression.(*ast.AssignmentExpression)
				assert.True(t, ok, "Expected AssignmentExpression")
				
				fcc, ok := assign.Right.(*ast.FirstClassCallable)
				assert.True(t, ok, "Expected FirstClassCallable")
				
				ident, ok := fcc.Callable.(*ast.IdentifierNode)
				assert.True(t, ok, "Expected IdentifierNode")
				assert.Equal(t, "strlen", ident.Name)
			},
		},
		{
			name:  "Object method first-class callable",
			input: `<?php $method = $obj->method(...);`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				
				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				assert.True(t, ok, "Expected ExpressionStatement")
				
				assign, ok := stmt.Expression.(*ast.AssignmentExpression)
				assert.True(t, ok, "Expected AssignmentExpression")
				
				fcc, ok := assign.Right.(*ast.FirstClassCallable)
				assert.True(t, ok, "Expected FirstClassCallable")
				
				propAccess, ok := fcc.Callable.(*ast.PropertyAccessExpression)
				assert.True(t, ok, "Expected PropertyAccessExpression")
				
				objVar, ok := propAccess.Object.(*ast.Variable)
				assert.True(t, ok, "Expected Variable")
				assert.Equal(t, "$obj", objVar.Name)
				
				assert.Equal(t, "method", propAccess.Property.Name)
			},
		},
		{
			name:  "Static method first-class callable",
			input: `<?php $staticMethod = MyClass::method(...);`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				
				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				assert.True(t, ok, "Expected ExpressionStatement")
				
				assign, ok := stmt.Expression.(*ast.AssignmentExpression)
				assert.True(t, ok, "Expected AssignmentExpression")
				
				fcc, ok := assign.Right.(*ast.FirstClassCallable)
				assert.True(t, ok, "Expected FirstClassCallable")
				
				staticAccess, ok := fcc.Callable.(*ast.StaticAccessExpression)
				assert.True(t, ok, "Expected StaticAccessExpression")
				
				className, ok := staticAccess.Class.(*ast.IdentifierNode)
				assert.True(t, ok, "Expected IdentifierNode")
				assert.Equal(t, "MyClass", className.Name)
				
				methodName, ok := staticAccess.Property.(*ast.IdentifierNode)
				assert.True(t, ok, "Expected IdentifierNode")
				assert.Equal(t, "method", methodName.Name)
			},
		},
		{
			name:  "Built-in function first-class callable",
			input: `<?php $builtin = array_map(...);`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				
				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				assert.True(t, ok, "Expected ExpressionStatement")
				
				assign, ok := stmt.Expression.(*ast.AssignmentExpression)
				assert.True(t, ok, "Expected AssignmentExpression")
				
				fcc, ok := assign.Right.(*ast.FirstClassCallable)
				assert.True(t, ok, "Expected FirstClassCallable")
				
				ident, ok := fcc.Callable.(*ast.IdentifierNode)
				assert.True(t, ok, "Expected IdentifierNode")
				assert.Equal(t, "array_map", ident.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			
			checkParserErrors(t, p)

			assert.NotNil(t, program)
			tt.validate(t, program)
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
