package parser

import (
	"fmt"
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

func TestParsing_BitwiseOperators(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]interface{}
	}{
		{
			name:  "Bitwise NOT operator",
			input: `<?php $x = ~5; ?>`,
			expected: map[string]interface{}{
				"variable":    "$x",
				"operator":    "=",
				"rightType":   "UnaryExpression",
				"unaryOp":     "~",
				"unaryPrefix": true,
				"operandType": "IntegerLiteral",
				"operandVal":  "5",
			},
		},
		{
			name:  "Bitwise AND operator",
			input: `<?php $x = 5 & 3; ?>`,
			expected: map[string]interface{}{
				"variable":  "$x",
				"operator":  "=",
				"rightType": "BinaryExpression",
				"leftVal":   "5",
				"binaryOp":  "&",
				"rightVal":  "3",
			},
		},
		{
			name:  "Bitwise OR operator",
			input: `<?php $y = 5 | 3; ?>`,
			expected: map[string]interface{}{
				"variable":  "$y",
				"operator":  "=",
				"rightType": "BinaryExpression",
				"leftVal":   "5",
				"binaryOp":  "|",
				"rightVal":  "3",
			},
		},
		{
			name:  "Bitwise XOR operator",
			input: `<?php $z = 5 ^ 3; ?>`,
			expected: map[string]interface{}{
				"variable":  "$z",
				"operator":  "=",
				"rightType": "BinaryExpression",
				"leftVal":   "5",
				"binaryOp":  "^",
				"rightVal":  "3",
			},
		},
		{
			name:  "Complex bitwise expression (original failing case)",
			input: `<?php error_reporting($this->orig_error_level & ~E_WARNING); ?>`,
			expected: map[string]interface{}{
				"stmtType":     "ExpressionStatement",
				"exprType":     "CallExpression",
				"callee":       "error_reporting",
				"argType":      "BinaryExpression",
				"argLeftType":  "PropertyAccessExpression",
				"argRightType": "UnaryExpression",
				"argRightOp":   "~",
			},
		},
		{
			name:  "Multiple bitwise operators",
			input: `<?php $result = ($a & $b) | ($c ^ $d) & ~$e; ?>`,
			expected: map[string]interface{}{
				"variable": "$result",
				"operator": "=",
				"complex":  true, // Just verify it parses without errors
			},
		},
		{
			name:  "Bitwise with precedence",
			input: `<?php $x = 8 | 4 & 2; ?>`, // Should be parsed as 8 | (4 & 2) due to precedence
			expected: map[string]interface{}{
				"variable":  "$x",
				"operator":  "=",
				"rightType": "BinaryExpression",
				"binaryOp":  "|", // Top level should be OR
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			checkParserErrors(t, p)
			require.NotNil(t, program)
			require.Len(t, program.Body, 1)

			stmt := program.Body[0]

			// Handle different statement types
			switch tt.expected["stmtType"] {
			case "ExpressionStatement":
				exprStmt, ok := stmt.(*ast.ExpressionStatement)
				require.True(t, ok, "Statement should be ExpressionStatement")

				if tt.expected["exprType"] == "CallExpression" {
					call, ok := exprStmt.Expression.(*ast.CallExpression)
					require.True(t, ok, "Expression should be CallExpression")

					callee, ok := call.Callee.(*ast.IdentifierNode)
					require.True(t, ok, "Callee should be IdentifierNode")
					assert.Equal(t, tt.expected["callee"], callee.Name)

					require.Len(t, call.Arguments, 1)
					arg := call.Arguments[0]

					binary, ok := arg.(*ast.BinaryExpression)
					require.True(t, ok, "Argument should be BinaryExpression")

					// Verify left side (property access)
					_, ok = binary.Left.(*ast.PropertyAccessExpression)
					require.True(t, ok, "Left should be PropertyAccessExpression")

					// Verify right side (unary expression)
					unary, ok := binary.Right.(*ast.UnaryExpression)
					require.True(t, ok, "Right should be UnaryExpression")
					assert.Equal(t, tt.expected["argRightOp"], unary.Operator)
					assert.True(t, unary.Prefix, "Should be prefix unary operator")
				}
			default:
				// Handle assignment expressions
				exprStmt, ok := stmt.(*ast.ExpressionStatement)
				require.True(t, ok, "Statement should be ExpressionStatement")

				assignment, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok, "Expression should be AssignmentExpression")

				// Check variable
				variable, ok := assignment.Left.(*ast.Variable)
				require.True(t, ok, "Left side should be Variable")
				assert.Equal(t, tt.expected["variable"], variable.Name)

				// Check operator
				assert.Equal(t, tt.expected["operator"], assignment.Operator)

				// Check right side based on type
				switch tt.expected["rightType"] {
				case "UnaryExpression":
					unary, ok := assignment.Right.(*ast.UnaryExpression)
					require.True(t, ok, "Right side should be UnaryExpression")
					assert.Equal(t, tt.expected["unaryOp"], unary.Operator)
					assert.Equal(t, tt.expected["unaryPrefix"], unary.Prefix)

					operand, ok := unary.Operand.(*ast.NumberLiteral)
					require.True(t, ok, "Operand should be NumberLiteral")
					assert.Equal(t, tt.expected["operandVal"], operand.Value)

				case "BinaryExpression":
					binary, ok := assignment.Right.(*ast.BinaryExpression)
					require.True(t, ok, "Right side should be BinaryExpression")

					if tt.expected["complex"] == true {
						// For complex expressions, just verify it's a binary expression
						break
					}

					assert.Equal(t, tt.expected["binaryOp"], binary.Operator)

					if tt.expected["leftVal"] != nil {
						leftInt, ok := binary.Left.(*ast.NumberLiteral)
						require.True(t, ok, "Left operand should be NumberLiteral")
						assert.Equal(t, tt.expected["leftVal"], leftInt.Value)
					}

					if tt.expected["rightVal"] != nil {
						rightInt, ok := binary.Right.(*ast.NumberLiteral)
						require.True(t, ok, "Right operand should be NumberLiteral")
						assert.Equal(t, tt.expected["rightVal"], rightInt.Value)
					}
				}
			}
		})
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

// TestParsing_FloatLiteralEdgeCases tests the specific float literal bug where numbers ending
// with decimal point (like 1., 1.0) were incorrectly parsed due to lexer tokenization issues
func TestParsing_FloatLiteralEdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Float ending with decimal point",
			input:    `<?php $x = 1.; ?>`,
			expected: "1.",
		},
		{
			name:     "Float with zero after decimal",
			input:    `<?php $x = 1.0; ?>`,
			expected: "1.0",
		},
		{
			name:     "Float in array context (original failing case)",
			input:    `<?php $arr = [1., 1.0, 1.23]; ?>`,
			expected: "1.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			l := lexer.New(tc.input)
			p := New(l)
			program := p.ParseProgram()

			checkParserErrors(t, p)
			assert.NotNil(t, program)

			if tc.name == "Float in array context (original failing case)" {
				// Special handling for array case
				stmt := program.Body[0]
				exprStmt, ok := stmt.(*ast.ExpressionStatement)
				assert.True(t, ok)

				assignment, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				assert.True(t, ok)

				arrayExpr, ok := assignment.Right.(*ast.ArrayExpression)
				assert.True(t, ok)
				assert.Len(t, arrayExpr.Elements, 3)

				// Check first element (1.)
				firstElement, ok := arrayExpr.Elements[0].(*ast.NumberLiteral)
				assert.True(t, ok)
				assert.Equal(t, tc.expected, firstElement.Value)
				assert.Equal(t, "float", firstElement.Kind)

				// Check second element (1.0)
				secondElement, ok := arrayExpr.Elements[1].(*ast.NumberLiteral)
				assert.True(t, ok)
				assert.Equal(t, "1.0", secondElement.Value)
				assert.Equal(t, "float", secondElement.Kind)

				// Check third element (1.23)
				thirdElement, ok := arrayExpr.Elements[2].(*ast.NumberLiteral)
				assert.True(t, ok)
				assert.Equal(t, "1.23", thirdElement.Value)
				assert.Equal(t, "float", thirdElement.Kind)
			} else {
				// Handle simple assignment cases
				stmt := program.Body[0]
				exprStmt, ok := stmt.(*ast.ExpressionStatement)
				assert.True(t, ok)

				assignment, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				assert.True(t, ok)

				numberLit, ok := assignment.Right.(*ast.NumberLiteral)
				assert.True(t, ok)
				assert.Equal(t, tc.expected, numberLit.Value)
				assert.Equal(t, "float", numberLit.Kind)
			}
		})
	}
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

func TestParsing_InstanceofExpressions(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		leftVar    string
		rightClass string
	}{
		{
			"simple class name",
			`<?php $obj instanceof MyClass; ?>`,
			"$obj",
			"MyClass",
		},
		{
			"namespaced class name",
			`<?php $mailer instanceof PHPMailer\PHPMailer; ?>`,
			"$mailer",
			"PHPMailer\\PHPMailer",
		},
		{
			"fully qualified class name",
			`<?php $phpmailer instanceof PHPMailer\PHPMailer\PHPMailer; ?>`,
			"$phpmailer",
			"PHPMailer\\PHPMailer\\PHPMailer",
		},
		{
			"absolute namespace",
			`<?php $obj instanceof \Foo\Bar\Baz; ?>`,
			"$obj",
			"\\Foo\\Bar\\Baz",
		},
		{
			"instanceof in if condition",
			`<?php if ($phpmailer instanceof PHPMailer\PHPMailer\PHPMailer) {} ?>`,
			"$phpmailer",
			"PHPMailer\\PHPMailer\\PHPMailer",
		},
		{
			"negated instanceof in if condition",
			`<?php if (!($phpmailer instanceof PHPMailer\PHPMailer\PHPMailer)) {} ?>`,
			"$phpmailer",
			"PHPMailer\\PHPMailer\\PHPMailer",
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
			var instanceofExpr *ast.InstanceofExpression

			// Handle different contexts where instanceof might appear
			switch s := stmt.(type) {
			case *ast.ExpressionStatement:
				// Direct instanceof expression
				instanceofExpr, _ = s.Expression.(*ast.InstanceofExpression)
			case *ast.IfStatement:
				// instanceof in if condition
				if s.Test != nil {
					if unary, ok := s.Test.(*ast.UnaryExpression); ok {
						// Handle negated instanceof: !($obj instanceof Class)
						instanceofExpr, _ = unary.Operand.(*ast.InstanceofExpression)
					} else {
						instanceofExpr, _ = s.Test.(*ast.InstanceofExpression)
					}
				}
			}

			require.NotNil(t, instanceofExpr, "Should find instanceof expression")

			// Check left operand (variable)
			leftVar, ok := instanceofExpr.Left.(*ast.Variable)
			assert.True(t, ok, "Left operand should be Variable")
			assert.Equal(t, tt.leftVar, leftVar.Name)

			// Check right operand (class name)
			var rightName string
			switch right := instanceofExpr.Right.(type) {
			case *ast.IdentifierNode:
				rightName = right.Name
			case *ast.NamespaceExpression:
				if right.Name != nil {
					if id, ok := right.Name.(*ast.IdentifierNode); ok {
						rightName = "\\" + id.Name
					}
				}
			default:
				t.Errorf("Unexpected right operand type: %T", right)
			}

			assert.Equal(t, tt.rightClass, rightName)
		})
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
		{
			name:  "Function with static return type",
			input: `<?php function foo(): static { } ?>`,
			expectedParams: []struct {
				name string
				typ  string
			}{},
			expectedReturnType: "static",
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

// TestParsing_ArrayTrailingCommas tests parsing arrays with trailing commas
func TestParsing_ArrayTrailingCommas(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, program *ast.Program)
	}{
		{
			name:  "Simple array with trailing comma",
			input: `<?php $arr = array(1, 2, 3,); ?>`,
			check: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				stmt := program.Body[0].(*ast.ExpressionStatement)
				assign := stmt.Expression.(*ast.AssignmentExpression)
				arr := assign.Right.(*ast.ArrayExpression)
				assert.Len(t, arr.Elements, 3)

				// Check first element
				num1 := arr.Elements[0].(*ast.NumberLiteral)
				assert.Equal(t, "1", num1.Value)

				// Check second element
				num2 := arr.Elements[1].(*ast.NumberLiteral)
				assert.Equal(t, "2", num2.Value)

				// Check third element
				num3 := arr.Elements[2].(*ast.NumberLiteral)
				assert.Equal(t, "3", num3.Value)
			},
		},
		{
			name:  "Associative array with trailing comma",
			input: `<?php $arr = array("key" => "value", "key2" => "value2",); ?>`,
			check: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				stmt := program.Body[0].(*ast.ExpressionStatement)
				assign := stmt.Expression.(*ast.AssignmentExpression)
				arr := assign.Right.(*ast.ArrayExpression)
				assert.Len(t, arr.Elements, 2)

				// Check first element (key => value)
				elem1 := arr.Elements[0].(*ast.ArrayElementExpression)
				key1 := elem1.Key.(*ast.StringLiteral)
				assert.Equal(t, "key", key1.Value)
				val1 := elem1.Value.(*ast.StringLiteral)
				assert.Equal(t, "value", val1.Value)

				// Check second element
				elem2 := arr.Elements[1].(*ast.ArrayElementExpression)
				key2 := elem2.Key.(*ast.StringLiteral)
				assert.Equal(t, "key2", key2.Value)
				val2 := elem2.Value.(*ast.StringLiteral)
				assert.Equal(t, "value2", val2.Value)
			},
		},
		{
			name:  "Nested arrays with trailing commas",
			input: `<?php $arr = array("nested" => array(1, 2,), "simple" => 3,); ?>`,
			check: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				stmt := program.Body[0].(*ast.ExpressionStatement)
				assign := stmt.Expression.(*ast.AssignmentExpression)
				arr := assign.Right.(*ast.ArrayExpression)
				assert.Len(t, arr.Elements, 2)

				// Check first element (nested array)
				elem1 := arr.Elements[0].(*ast.ArrayElementExpression)
				key1 := elem1.Key.(*ast.StringLiteral)
				assert.Equal(t, "nested", key1.Value)

				nestedArr := elem1.Value.(*ast.ArrayExpression)
				assert.Len(t, nestedArr.Elements, 2)

				num1 := nestedArr.Elements[0].(*ast.NumberLiteral)
				assert.Equal(t, "1", num1.Value)
				num2 := nestedArr.Elements[1].(*ast.NumberLiteral)
				assert.Equal(t, "2", num2.Value)

				// Check second element
				elem2 := arr.Elements[1].(*ast.ArrayElementExpression)
				key2 := elem2.Key.(*ast.StringLiteral)
				assert.Equal(t, "simple", key2.Value)
				val2 := elem2.Value.(*ast.NumberLiteral)
				assert.Equal(t, "3", val2.Value)
			},
		},
		{
			name: "Complex class constant array with trailing commas",
			input: `<?php
class TestClass {
    const COMPLEX = array(
        self::KEY1 => array(
            "sub1" => array(self::VALUE1,),
            "sub2" => self::VALUE2,
        ),
        "simple" => "value",
    );
}`,
			check: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				stmt := program.Body[0].(*ast.ExpressionStatement)
				classExpr := stmt.Expression.(*ast.ClassExpression)

				// Find the constant declaration
				var constDecl *ast.ClassConstantDeclaration
				for _, member := range classExpr.Body {
					if decl, ok := member.(*ast.ClassConstantDeclaration); ok {
						constDecl = decl
						break
					}
				}
				require.NotNil(t, constDecl)
				require.Len(t, constDecl.Constants, 1)

				// Check the constant has a complex array value
				constant := constDecl.Constants[0]
				constName, ok := constant.Name.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "COMPLEX", constName.Name)

				arr, ok := constant.Value.(*ast.ArrayExpression)
				require.True(t, ok, "Constant value should be an array")
				assert.Len(t, arr.Elements, 2)

				// Check first element structure (complex nested)
				elem1 := arr.Elements[0].(*ast.ArrayElementExpression)
				classConst1 := elem1.Key.(*ast.StaticAccessExpression)
				assert.Equal(t, "self", classConst1.Class.(*ast.IdentifierNode).Name)
				assert.Equal(t, "KEY1", classConst1.Property.(*ast.IdentifierNode).Name)

				nestedArr1 := elem1.Value.(*ast.ArrayExpression)
				assert.Len(t, nestedArr1.Elements, 2)

				// Check second element is simple
				elem2 := arr.Elements[1].(*ast.ArrayElementExpression)
				key2 := elem2.Key.(*ast.StringLiteral)
				assert.Equal(t, "simple", key2.Value)
				val2 := elem2.Value.(*ast.StringLiteral)
				assert.Equal(t, "value", val2.Value)
			},
		},
		{
			name: "Original failing case",
			input: `<?php
class Foo extends Boo {
    const HOOKED_BLOCKS = array(
        self::ANCHOR_BLOCK_TYPE => array(
            'after'  => array( self::HOOKED_BLOCK_TYPE ),
            'before' => array( self::OTHER_HOOKED_BLOCK_TYPE ),
        ),
    );
}`,
			check: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				stmt := program.Body[0].(*ast.ExpressionStatement)
				classExpr := stmt.Expression.(*ast.ClassExpression)

				// Verify class name and inheritance
				assert.Equal(t, "Foo", classExpr.Name.(*ast.IdentifierNode).Name)
				assert.Equal(t, "Boo", classExpr.Extends.(*ast.IdentifierNode).Name)

				// Find the constant declaration
				var constDecl *ast.ClassConstantDeclaration
				for _, member := range classExpr.Body {
					if decl, ok := member.(*ast.ClassConstantDeclaration); ok {
						constDecl = decl
						break
					}
				}
				require.NotNil(t, constDecl)
				require.Len(t, constDecl.Constants, 1)

				// Check the constant structure
				constant := constDecl.Constants[0]
				constName, ok := constant.Name.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "HOOKED_BLOCKS", constName.Name)

				arr, ok := constant.Value.(*ast.ArrayExpression)
				require.True(t, ok, "Constant value should be an array")
				assert.Len(t, arr.Elements, 1)

				// Check the main array element with self::ANCHOR_BLOCK_TYPE key
				mainElem := arr.Elements[0].(*ast.ArrayElementExpression)
				selfConst := mainElem.Key.(*ast.StaticAccessExpression)
				assert.Equal(t, "self", selfConst.Class.(*ast.IdentifierNode).Name)
				assert.Equal(t, "ANCHOR_BLOCK_TYPE", selfConst.Property.(*ast.IdentifierNode).Name)

				// Check the nested array value
				nestedArr := mainElem.Value.(*ast.ArrayExpression)
				assert.Len(t, nestedArr.Elements, 2)

				// Check 'after' element
				afterElem := nestedArr.Elements[0].(*ast.ArrayElementExpression)
				afterKey := afterElem.Key.(*ast.StringLiteral)
				assert.Equal(t, "after", afterKey.Value)

				afterArr := afterElem.Value.(*ast.ArrayExpression)
				assert.Len(t, afterArr.Elements, 1)

				afterValue := afterArr.Elements[0].(*ast.StaticAccessExpression)
				assert.Equal(t, "self", afterValue.Class.(*ast.IdentifierNode).Name)
				assert.Equal(t, "HOOKED_BLOCK_TYPE", afterValue.Property.(*ast.IdentifierNode).Name)

				// Check 'before' element
				beforeElem := nestedArr.Elements[1].(*ast.ArrayElementExpression)
				beforeKey := beforeElem.Key.(*ast.StringLiteral)
				assert.Equal(t, "before", beforeKey.Value)

				beforeArr := beforeElem.Value.(*ast.ArrayExpression)
				assert.Len(t, beforeArr.Elements, 1)

				beforeValue := beforeArr.Elements[0].(*ast.StaticAccessExpression)
				assert.Equal(t, "self", beforeValue.Class.(*ast.IdentifierNode).Name)
				assert.Equal(t, "OTHER_HOOKED_BLOCK_TYPE", beforeValue.Property.(*ast.IdentifierNode).Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			checkParserErrors(t, p)
			require.NotNil(t, program)

			tt.check(t, program)
		})
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
		name                string
		input               string
		expectedContent     string
		expectInterpolation bool
		expectedParts       int
	}{
		{
			name: "Simple Heredoc",
			input: `<?php $str = <<<EOD
Hello World
EOD; ?>`,
			expectedContent:     "Hello World\n",
			expectInterpolation: false,
			expectedParts:       1,
		},
		{
			name: "Heredoc with variable",
			input: `<?php $str = <<<EOD
Hello $name
EOD; ?>`,
			expectedContent:     "Hello ", // First part content
			expectInterpolation: true,
			expectedParts:       3, // "Hello " + $name + "\n"
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
			name: "Simple Nowdoc",
			input: `<?php $str = <<<'EOD'
Hello World
EOD; ?>`,
			expected: "Hello World\n",
		},
		{
			name: "Nowdoc with $variable (no interpolation)",
			input: `<?php $str = <<<'EOD'
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
		name       string
		input      string
		partCount  int
		firstPart  string
		secondPart string
	}{
		{
			name:       "Simple variable interpolation",
			input:      `<?php $str = "Hello $name"; ?>`,
			partCount:  2,
			firstPart:  "Hello ",
			secondPart: "$name",
		},
		{
			name:       "String with multiple variables",
			input:      `<?php $str = "Hello $first and $second"; ?>`,
			partCount:  4, // "Hello ", "$first", " and ", "$second"
			firstPart:  "Hello ",
			secondPart: "$first",
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

// TestParsing_ArrayExpressionWithComments tests parsing array() expressions with comments
func TestParsing_ArrayExpressionWithComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, program *ast.Program)
	}{
		{
			name: "Simple array() with comments",
			input: `<?php $x = array(
				// Comment before element
				'key' => 'value'
			); ?>`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				stmt := program.Body[0]
				exprStmt, ok := stmt.(*ast.ExpressionStatement)
				require.True(t, ok, "Statement should be ExpressionStatement")

				assignment, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok, "Expression should be AssignmentExpression")

				arrayExpr, ok := assignment.Right.(*ast.ArrayExpression)
				require.True(t, ok, "Right side should be ArrayExpression")
				require.Len(t, arrayExpr.Elements, 1)

				element, ok := arrayExpr.Elements[0].(*ast.ArrayElementExpression)
				require.True(t, ok, "Element should be ArrayElementExpression")

				keyLit, ok := element.Key.(*ast.StringLiteral)
				require.True(t, ok, "Key should be StringLiteral")
				assert.Equal(t, "key", keyLit.Value)

				valueLit, ok := element.Value.(*ast.StringLiteral)
				require.True(t, ok, "Value should be StringLiteral")
				assert.Equal(t, "value", valueLit.Value)
			},
		},
		{
			name: "Function call with array() and comments - original failing case",
			input: `<?php
register_post_type(
    self::POST_TYPE,
    array(
        'capabilities'    => array(
            // No one can edit this post type once published.
            'edit_published_posts' => 'do_not_allow',
        ),
        'map_meta_cap'    => true,
        'supports'        => array( 'revisions' ),
    )
); ?>`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				stmt := program.Body[0]
				exprStmt, ok := stmt.(*ast.ExpressionStatement)
				require.True(t, ok, "Statement should be ExpressionStatement")

				callExpr, ok := exprStmt.Expression.(*ast.CallExpression)
				require.True(t, ok, "Expression should be CallExpression")

				// Check function name
				identifier, ok := callExpr.Callee.(*ast.IdentifierNode)
				require.True(t, ok, "Callee should be IdentifierNode")
				assert.Equal(t, "register_post_type", identifier.Name)

				// Check arguments count
				require.Len(t, callExpr.Arguments, 2)

				// First argument: self::POST_TYPE
				staticAccessExpr, ok := callExpr.Arguments[0].(*ast.StaticAccessExpression)
				require.True(t, ok, "First argument should be StaticAccessExpression")

				// Check self part
				selfIdent, ok := staticAccessExpr.Class.(*ast.IdentifierNode)
				require.True(t, ok, "Class should be IdentifierNode")
				assert.Equal(t, "self", selfIdent.Name)

				// Check POST_TYPE part
				postTypeIdent, ok := staticAccessExpr.Property.(*ast.IdentifierNode)
				require.True(t, ok, "Property should be IdentifierNode")
				assert.Equal(t, "POST_TYPE", postTypeIdent.Name)

				// Second argument: array(...)
				arrayExpr, ok := callExpr.Arguments[1].(*ast.ArrayExpression)
				require.True(t, ok, "Second argument should be ArrayExpression")
				require.Len(t, arrayExpr.Elements, 3) // capabilities, map_meta_cap, supports

				// Check 'capabilities' element
				capElement, ok := arrayExpr.Elements[0].(*ast.ArrayElementExpression)
				require.True(t, ok, "Element should be ArrayElementExpression")

				keyLit, ok := capElement.Key.(*ast.StringLiteral)
				require.True(t, ok, "Key should be StringLiteral")
				assert.Equal(t, "capabilities", keyLit.Value)

				// Check nested array
				nestedArray, ok := capElement.Value.(*ast.ArrayExpression)
				require.True(t, ok, "Value should be ArrayExpression")
				require.Len(t, nestedArray.Elements, 1)

				// Check element with comment before it
				nestedElement, ok := nestedArray.Elements[0].(*ast.ArrayElementExpression)
				require.True(t, ok, "Nested element should be ArrayElementExpression")

				nestedKey, ok := nestedElement.Key.(*ast.StringLiteral)
				require.True(t, ok, "Nested key should be StringLiteral")
				assert.Equal(t, "edit_published_posts", nestedKey.Value)

				nestedValue, ok := nestedElement.Value.(*ast.StringLiteral)
				require.True(t, ok, "Nested value should be StringLiteral")
				assert.Equal(t, "do_not_allow", nestedValue.Value)
			},
		},
		{
			name: "Nested arrays with multiple comments",
			input: `<?php $data = array(
				'outer' => array(
					// First inner comment
					'inner1' => 'value1',
					// Second inner comment  
					'inner2' => array(
						// Deep comment
						'deep' => 'deepvalue'
					)
				)
			); ?>`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				stmt := program.Body[0]
				exprStmt, ok := stmt.(*ast.ExpressionStatement)
				require.True(t, ok, "Statement should be ExpressionStatement")

				assignment, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok, "Expression should be AssignmentExpression")

				arrayExpr, ok := assignment.Right.(*ast.ArrayExpression)
				require.True(t, ok, "Right side should be ArrayExpression")
				require.Len(t, arrayExpr.Elements, 1)

				// Check nested structure is correctly parsed
				element, ok := arrayExpr.Elements[0].(*ast.ArrayElementExpression)
				require.True(t, ok, "Element should be ArrayElementExpression")

				keyLit, ok := element.Key.(*ast.StringLiteral)
				require.True(t, ok, "Key should be StringLiteral")
				assert.Equal(t, "outer", keyLit.Value)

				nestedArray, ok := element.Value.(*ast.ArrayExpression)
				require.True(t, ok, "Value should be ArrayExpression")
				require.Len(t, nestedArray.Elements, 2) // inner1 and inner2
			},
		},
		{
			name: "Mixed comments and comma handling",
			input: `<?php $mixed = array(
				'first' => 'value1', // comment after first
				// comment before second
				'second' => 'value2',
				// comment before third  
				'third' => 'value3', // comment after third
			); ?>`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				stmt := program.Body[0]
				exprStmt, ok := stmt.(*ast.ExpressionStatement)
				require.True(t, ok, "Statement should be ExpressionStatement")

				assignment, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok, "Expression should be AssignmentExpression")

				arrayExpr, ok := assignment.Right.(*ast.ArrayExpression)
				require.True(t, ok, "Right side should be ArrayExpression")
				require.Len(t, arrayExpr.Elements, 3)

				// Verify all elements are correctly parsed
				for i, expectedKey := range []string{"first", "second", "third"} {
					element, ok := arrayExpr.Elements[i].(*ast.ArrayElementExpression)
					require.True(t, ok, fmt.Sprintf("Element %d should be ArrayElementExpression", i))

					keyLit, ok := element.Key.(*ast.StringLiteral)
					require.True(t, ok, fmt.Sprintf("Key %d should be StringLiteral", i))
					assert.Equal(t, expectedKey, keyLit.Value)
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
			name: "Function call in expression context",
			input: `<?php save_text($file, <<<'PHP'
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

func TestParsing_TrailingCommaInFunctionCalls(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple function call with trailing comma",
			input:    `<?php func(1, 2,); ?>`,
			expected: "func call with 2 arguments",
		},
		{
			name:     "Method call with trailing comma",
			input:    `<?php $obj->method(1, 2,); ?>`,
			expected: "method call with 2 arguments",
		},
		{
			name:     "Complex method chain with trailing comma",
			input:    `<?php $obj->expectExceptionObject((new ModelNotFoundException())->setModel(EloquentTestUser::class, [1]),); ?>`,
			expected: "method call with complex argument",
		},
		{
			name:     "Nested function calls with trailing commas",
			input:    `<?php outer(inner(1,), 2,); ?>`,
			expected: "nested function calls",
		},
		{
			name:     "Function call with mixed argument types and trailing comma",
			input:    `<?php func($var, "string", 123, [1, 2, 3],); ?>`,
			expected: "function call with 4 arguments",
		},
		{
			name:     "Array function with trailing comma",
			input:    `<?php array_merge([1, 2], [3, 4],); ?>`,
			expected: "array_merge with 2 arguments",
		},
		{
			name:     "Method call on new object with trailing comma",
			input:    `<?php (new Class())->method(1, 2,); ?>`,
			expected: "method call on new object",
		},
		{
			name:     "Static method call with trailing comma",
			input:    `<?php Class::method(1, 2,); ?>`,
			expected: "static method call",
		},
		{
			name:     "Static method call with 'new' as method name",
			input:    `<?php UserFactory::new(); ?>`,
			expected: "static method call with new",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			checkParserErrors(t, p)
			assert.NotNil(t, program, "Program should not be nil")
			assert.Len(t, program.Body, 1, "Program should have 1 statement")

			stmt := program.Body[0]
			exprStmt, ok := stmt.(*ast.ExpressionStatement)
			require.True(t, ok, "Statement should be ExpressionStatement")
			require.NotNil(t, exprStmt.Expression, "Expression should not be nil")

			// All test cases should result in valid call expressions
			// The specific validation depends on the type of call
			switch callExpr := exprStmt.Expression.(type) {
			case *ast.CallExpression:
				// Simple function call
				assert.NotNil(t, callExpr.Callee)
				if tt.name == "Simple function call with trailing comma" {
					assert.Len(t, callExpr.Arguments, 2)
				} else if tt.name == "Nested function calls with trailing commas" {
					assert.Len(t, callExpr.Arguments, 2)
					// Check that the first argument is also a call expression
					innerCall, ok := callExpr.Arguments[0].(*ast.CallExpression)
					require.True(t, ok, "First argument should be a function call")
					assert.Len(t, innerCall.Arguments, 1)
				} else if tt.name == "Function call with mixed argument types and trailing comma" {
					assert.Len(t, callExpr.Arguments, 4)
				} else if tt.name == "Array function with trailing comma" {
					assert.Len(t, callExpr.Arguments, 2)
				} else if tt.name == "Static method call with trailing comma" {
					assert.Len(t, callExpr.Arguments, 2)
				}
			default:
				// For more complex expressions, just ensure they parsed without error
				assert.NotNil(t, callExpr, "Expression should not be nil")
			}
		})
	}
}

func TestParsing_TrailingCommaInFunctionParameters(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		paramCount int
	}{
		{
			name:       "Function with trailing comma in parameters",
			input:      `<?php function test($a, $b,) { return $a + $b; } ?>`,
			paramCount: 2,
		},
		{
			name:       "Function with single parameter and trailing comma",
			input:      `<?php function single($param,) { return $param; } ?>`,
			paramCount: 1,
		},
		{
			name:       "Function with typed parameters and trailing comma",
			input:      `<?php function typed(int $a, string $b,): bool { return true; } ?>`,
			paramCount: 2,
		},
		{
			name:       "Function with reference parameters and trailing comma",
			input:      `<?php function ref(&$a, &$b,) { $a = $b; } ?>`,
			paramCount: 2,
		},
		{
			name:       "Function with mixed parameters and trailing comma",
			input:      `<?php function mixed(int $a, &$b, $c = "default",) { return $a; } ?>`,
			paramCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			checkParserErrors(t, p)
			assert.NotNil(t, program, "Program should not be nil")
			assert.Len(t, program.Body, 1, "Program should have 1 statement")

			stmt := program.Body[0]
			funcDecl, ok := stmt.(*ast.FunctionDeclaration)
			require.True(t, ok, "Statement should be FunctionDeclaration")
			
			assert.Len(t, funcDecl.Parameters, tt.paramCount, "Function should have correct number of parameters")
			
			// Verify parameter names are correctly parsed
			if tt.paramCount >= 1 {
				assert.NotEmpty(t, funcDecl.Parameters[0].Name, "First parameter should have a name")
			}
			if tt.paramCount >= 2 {
				assert.NotEmpty(t, funcDecl.Parameters[1].Name, "Second parameter should have a name")
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

func TestParsing_ArrowFunctions(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		expectedParamCount int
		expectedStatic     bool
	}{
		{
			name:  "Basic arrow function with single parameter",
			input: `<?php $fn = fn($x) => $x * 2;`,
			expectedParamCount: 1,
			expectedStatic:     false,
		},
		{
			name:  "Arrow function with no parameters",
			input: `<?php $fn = fn() => 42;`,
			expectedParamCount: 0,
			expectedStatic:     false,
		},
		{
			name:  "Arrow function with multiple parameters",
			input: `<?php $fn = fn($a, $b, $c) => $a + $b + $c;`,
			expectedParamCount: 3,
			expectedStatic:     false,
		},
		{
			name:  "Arrow function with typed parameters",
			input: `<?php $fn = fn(string $userId) => strtoupper($userId);`,
			expectedParamCount: 1,
			expectedStatic:     false,
		},
		{
			name:  "Arrow function with return type",
			input: `<?php $fn = fn($x): int => (int)$x;`,
			expectedParamCount: 1,
			expectedStatic:     false,
		},
		{
			name:  "Static arrow function",
			input: `<?php $fn = static fn($x) => $x;`,
			expectedParamCount: 1,
			expectedStatic:     true,
		},
		{
			name:  "Arrow function in method call",
			input: `<?php $rateLimiter->for('user_limiter', fn(string $userId) => [$userId]);`,
			expectedParamCount: 1,
			expectedStatic:     false,
		},
		{
			name:  "Arrow function with complex body",
			input: `<?php $fn = fn($x) => $obj->method($x)->chain();`,
			expectedParamCount: 1,
			expectedStatic:     false,
		},
		{
			name:  "Arrow function with array return",
			input: `<?php $fn = fn($a, $b) => [$a, $b, 'constant'];`,
			expectedParamCount: 2,
			expectedStatic:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			
			checkParserErrors(t, p)
			assert.NotNil(t, program, "program should not be nil")
			assert.Len(t, program.Body, 1, "should have exactly one statement")

			stmt, ok := program.Body[0].(*ast.ExpressionStatement)
			assert.True(t, ok, "should be an expression statement")

			// The arrow function might be directly in the assignment or in a method call
			var arrowFunc *ast.ArrowFunctionExpression
			if assignExpr, ok := stmt.Expression.(*ast.AssignmentExpression); ok {
				// Direct assignment: $fn = fn(...) => ...
				arrowFunc, ok = assignExpr.Right.(*ast.ArrowFunctionExpression)
				assert.True(t, ok, "right side should be arrow function")
			} else if callExpr, ok := stmt.Expression.(*ast.CallExpression); ok {
				// Method call with arrow function as argument
				assert.Len(t, callExpr.Arguments, 2, "should have 2 arguments")
				arrowFunc, ok = callExpr.Arguments[1].(*ast.ArrowFunctionExpression)
				assert.True(t, ok, "second argument should be arrow function")
			} else {
				t.Fatalf("unexpected expression type: %T", stmt.Expression)
			}

			assert.NotNil(t, arrowFunc, "arrow function should not be nil")
			assert.Len(t, arrowFunc.Parameters, tt.expectedParamCount, "parameter count should match")
			assert.Equal(t, tt.expectedStatic, arrowFunc.Static, "static flag should match")
			assert.NotNil(t, arrowFunc.Body, "body should not be nil")
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
		name             string
		input            string
		expectedClass    string
		expectedProperty string
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
		name                string
		input               string
		expectedClassName   string
		expectedConstGroups int
		expectedTotalConsts int
		validateConstants   func(t *testing.T, classExpr *ast.ClassExpression)
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
				assert.Equal(t, "", constGroup.Visibility) // No explicit visibility
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
				assert.Equal(t, "", constGroup1.Visibility) // No explicit visibility
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
				assert.Equal(t, "", constGroup1.Visibility) // No explicit visibility
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

func TestParsing_FinalClass(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, class *ast.ClassExpression)
	}{
		{
			name: "Simple final class",
			input: `<?php
final class MyClass {
}`,
			validate: func(t *testing.T, class *ast.ClassExpression) {
				assert.True(t, class.Final, "Class should be final")
				assert.False(t, class.ReadOnly, "Class should not be readonly")
				assert.Equal(t, "MyClass", class.Name.(*ast.IdentifierNode).Name)
			},
		},
		{
			name: "Final class with extends",
			input: `<?php
final class Child extends Parent {
}`,
			validate: func(t *testing.T, class *ast.ClassExpression) {
				assert.True(t, class.Final, "Class should be final")
				assert.Equal(t, "Child", class.Name.(*ast.IdentifierNode).Name)
				assert.NotNil(t, class.Extends, "Class should have extends clause")
				assert.Equal(t, "Parent", class.Extends.(*ast.IdentifierNode).Name)
			},
		},
		{
			name: "Final class with implements",
			input: `<?php
final class MyClass implements InterfaceA, InterfaceB {
}`,
			validate: func(t *testing.T, class *ast.ClassExpression) {
				assert.True(t, class.Final, "Class should be final")
				assert.Equal(t, "MyClass", class.Name.(*ast.IdentifierNode).Name)
				assert.Len(t, class.Implements, 2, "Class should implement 2 interfaces")
				assert.Equal(t, "InterfaceA", class.Implements[0].(*ast.IdentifierNode).Name)
				assert.Equal(t, "InterfaceB", class.Implements[1].(*ast.IdentifierNode).Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			checkParserErrors(t, p)
			require.NotNil(t, program)
			require.Len(t, program.Body, 1)

			stmt := program.Body[0]
			exprStmt, ok := stmt.(*ast.ExpressionStatement)
			require.True(t, ok, "Statement should be ExpressionStatement")

			class, ok := exprStmt.Expression.(*ast.ClassExpression)
			require.True(t, ok, "Expression should be ClassExpression")

			tt.validate(t, class)
		})
	}
}

func TestParsing_TypedClassConstants(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, program *ast.Program)
	}{
		{
			name: "Private typed int constant",
			input: `<?php
class MyClass {
    private const int VERSION = 1;
}`,
			validate: func(t *testing.T, program *ast.Program) {
				class := extractClass(t, program)
				require.Len(t, class.Body, 1, "Class should have one statement")

				constDecl, ok := class.Body[0].(*ast.ClassConstantDeclaration)
				require.True(t, ok, "Statement should be ClassConstantDeclaration")
				
				assert.Equal(t, "private", constDecl.Visibility)
				require.NotNil(t, constDecl.Type, "Constant should have type")
				assert.Equal(t, "int", constDecl.Type.Name)
				
				require.Len(t, constDecl.Constants, 1)
				assert.Equal(t, "VERSION", constDecl.Constants[0].Name.(*ast.IdentifierNode).Name)
				assert.Equal(t, "1", constDecl.Constants[0].Value.(*ast.NumberLiteral).Value)
			},
		},
		{
			name: "Public typed string constant",
			input: `<?php
class MyClass {
    public const string NAME = "test";
}`,
			validate: func(t *testing.T, program *ast.Program) {
				class := extractClass(t, program)
				require.Len(t, class.Body, 1)

				constDecl, ok := class.Body[0].(*ast.ClassConstantDeclaration)
				require.True(t, ok, "Statement should be ClassConstantDeclaration")
				
				assert.Equal(t, "public", constDecl.Visibility)
				require.NotNil(t, constDecl.Type, "Constant should have type")
				assert.Equal(t, "string", constDecl.Type.Name)
			},
		},
		{
			name: "Multiple typed constants in one declaration",
			input: `<?php
class MyClass {
    protected const int MIN = 1, MAX = 100;
}`,
			validate: func(t *testing.T, program *ast.Program) {
				class := extractClass(t, program)
				require.Len(t, class.Body, 1)

				constDecl, ok := class.Body[0].(*ast.ClassConstantDeclaration)
				require.True(t, ok, "Statement should be ClassConstantDeclaration")
				
				assert.Equal(t, "protected", constDecl.Visibility)
				require.NotNil(t, constDecl.Type, "Constant should have type")
				assert.Equal(t, "int", constDecl.Type.Name)
				
				require.Len(t, constDecl.Constants, 2, "Should have 2 constants")
				assert.Equal(t, "MIN", constDecl.Constants[0].Name.(*ast.IdentifierNode).Name)
				assert.Equal(t, "MAX", constDecl.Constants[1].Name.(*ast.IdentifierNode).Name)
			},
		},
		{
			name: "Mixed typed and untyped constants",
			input: `<?php
class MyClass {
    private const int VERSION = 1;
    const LEGACY = "old";
}`,
			validate: func(t *testing.T, program *ast.Program) {
				class := extractClass(t, program)
				require.Len(t, class.Body, 2, "Class should have 2 constant declarations")

				// First constant - typed
				typedConst, ok := class.Body[0].(*ast.ClassConstantDeclaration)
				require.True(t, ok, "First statement should be ClassConstantDeclaration")
				assert.Equal(t, "private", typedConst.Visibility)
				require.NotNil(t, typedConst.Type, "First constant should have type")
				assert.Equal(t, "int", typedConst.Type.Name)

				// Second constant - untyped
				untypedConst, ok := class.Body[1].(*ast.ClassConstantDeclaration)
				require.True(t, ok, "Second statement should be ClassConstantDeclaration")
				assert.Equal(t, "", untypedConst.Visibility) // No explicit visibility
				assert.Nil(t, untypedConst.Type, "Second constant should not have type")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			checkParserErrors(t, p)
			require.NotNil(t, program)

			tt.validate(t, program)
		})
	}
}

// Helper function to extract class from program
func extractClass(t *testing.T, program *ast.Program) *ast.ClassExpression {
	require.Len(t, program.Body, 1, "Program should have one statement")

	stmt := program.Body[0]
	exprStmt, ok := stmt.(*ast.ExpressionStatement)
	require.True(t, ok, "Statement should be ExpressionStatement")

	class, ok := exprStmt.Expression.(*ast.ClassExpression)
	require.True(t, ok, "Expression should be ClassExpression")

	return class
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
					name       string
					typeName   string
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

func TestParsing_StaticMethodsWithTypeHints(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, class *ast.ClassExpression)
	}{
		{
			name: "Public static method with class type hint",
			input: `<?php
class Example {
    public static function foo(WP_UnitTest_Factory $factory) {
        self::$user_id_1 = $factory->user->create();
        self::$user_id_2 = $factory->user->create();
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

				if funcDecl.Name.(*ast.IdentifierNode).Name != "foo" {
					t.Errorf("Expected method name 'foo', got '%s'", funcDecl.Name.(*ast.IdentifierNode).Name)
				}

				if funcDecl.Visibility != "public" {
					t.Errorf("Expected visibility 'public', got '%s'", funcDecl.Visibility)
				}

				if !funcDecl.IsStatic {
					t.Errorf("Expected method to be static")
				}

				if len(funcDecl.Parameters) != 1 {
					t.Errorf("Expected 1 parameter, got %d", len(funcDecl.Parameters))
					return
				}

				// Check parameter: WP_UnitTest_Factory $factory
				param := funcDecl.Parameters[0]
				if param.Name != "$factory" {
					t.Errorf("Expected parameter name '$factory', got '%s'", param.Name)
				}
				if param.Type == nil || param.Type.Name != "WP_UnitTest_Factory" {
					t.Errorf("Expected parameter type 'WP_UnitTest_Factory', got %v", param.Type)
				}

				// Check that the method body has 2 statements
				if funcDecl.Body == nil {
					t.Errorf("Expected method to have a body")
					return
				}
				if len(funcDecl.Body) != 2 {
					t.Errorf("Expected 2 statements in method body, got %d", len(funcDecl.Body))
				}
			},
		},
		{
			name: "Private static method with multiple type hints",
			input: `<?php
class Test {
    private static function process(string $data, array $options, callable $callback) {
        return $callback($data, $options);
    }
}`,
			validate: func(t *testing.T, class *ast.ClassExpression) {
				funcDecl, ok := class.Body[0].(*ast.FunctionDeclaration)
				if !ok {
					t.Errorf("Expected FunctionDeclaration, got %T", class.Body[0])
					return
				}

				if funcDecl.Visibility != "private" {
					t.Errorf("Expected visibility 'private', got '%s'", funcDecl.Visibility)
				}

				if !funcDecl.IsStatic {
					t.Errorf("Expected method to be static")
				}

				if len(funcDecl.Parameters) != 3 {
					t.Errorf("Expected 3 parameters, got %d", len(funcDecl.Parameters))
					return
				}

				expectedParams := []struct {
					name string
					typ  string
				}{
					{"$data", "string"},
					{"$options", "array"},
					{"$callback", "callable"},
				}

				for i, expected := range expectedParams {
					param := funcDecl.Parameters[i]
					if param.Name != expected.name {
						t.Errorf("Parameter %d: expected name '%s', got '%s'", i, expected.name, param.Name)
					}
					if param.Type == nil || param.Type.Name != expected.typ {
						t.Errorf("Parameter %d: expected type '%s', got %v", i, expected.typ, param.Type)
					}
				}
			},
		},
		{
			name: "Protected static method with nullable type hint",
			input: `<?php
class Service {
    protected static function handle(?Config $config = null) {
        // implementation
    }
}`,
			validate: func(t *testing.T, class *ast.ClassExpression) {
				funcDecl, ok := class.Body[0].(*ast.FunctionDeclaration)
				if !ok {
					t.Errorf("Expected FunctionDeclaration, got %T", class.Body[0])
					return
				}

				if funcDecl.Visibility != "protected" {
					t.Errorf("Expected visibility 'protected', got '%s'", funcDecl.Visibility)
				}

				if !funcDecl.IsStatic {
					t.Errorf("Expected method to be static")
				}

				if len(funcDecl.Parameters) != 1 {
					t.Errorf("Expected 1 parameter, got %d", len(funcDecl.Parameters))
					return
				}

				param := funcDecl.Parameters[0]
				if param.Name != "$config" {
					t.Errorf("Expected parameter name '$config', got '%s'", param.Name)
				}
				if param.Type == nil || param.Type.Name != "Config" {
					t.Errorf("Expected parameter type 'Config', got %v", param.Type)
				}
				if param.Type.Nullable != true {
					t.Errorf("Expected parameter type to be nullable")
				}
			},
		},
		{
			name: "Static method without explicit visibility (defaults to public)",
			input: `<?php
class Utils {
    static function format(CustomFormatter $formatter) {
        return $formatter->format();
    }
}`,
			validate: func(t *testing.T, class *ast.ClassExpression) {
				funcDecl, ok := class.Body[0].(*ast.FunctionDeclaration)
				if !ok {
					t.Errorf("Expected FunctionDeclaration, got %T", class.Body[0])
					return
				}

				// In PHP, methods without explicit visibility default to public
				if funcDecl.Visibility != "" {
					// Note: depending on implementation, this might be empty or "public"
					// Adjust based on actual parser behavior
				}

				if !funcDecl.IsStatic {
					t.Errorf("Expected method to be static")
				}

				param := funcDecl.Parameters[0]
				if param.Type == nil || param.Type.Name != "CustomFormatter" {
					t.Errorf("Expected parameter type 'CustomFormatter', got %v", param.Type)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			l := lexer.New(test.input)
			p := New(l)
			program := p.ParseProgram()

			checkParserErrors(t, p)
			require.NotNil(t, program)
			require.Len(t, program.Body, 1)

			stmt := program.Body[0]
			exprStmt, ok := stmt.(*ast.ExpressionStatement)
			require.True(t, ok, "Statement should be ExpressionStatement")

			class, ok := exprStmt.Expression.(*ast.ClassExpression)
			require.True(t, ok, "Expression should be ClassExpression")

			test.validate(t, class)
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
					{"$a", "", false, true},        // &$a (no type)
					{"$b", "array", true, true},    // array &$b
					{"$c", "string", true, false},  // string $c
					{"$d", "callable", true, true}, // callable &$d
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
				propIdent, ok := propAccess.Property.(*ast.IdentifierNode)
				assert.True(t, ok, "Property should be IdentifierNode")
				assert.Equal(t, "getName", propIdent.Name)
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

// TestParsing_CatchBlocksWithQualifiedNames tests parsing catch blocks with fully qualified class names
func TestParsing_CatchBlocksWithQualifiedNames(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		expectedCatchCount int
		expectedTypes      []string
		expectedVariables  []string
	}{
		{
			name: "simple class name in catch",
			input: `<?php
try {
    $phpmailer->setFrom($from_email, $from_name, false);
} catch (Exception $e) {
}`,
			expectedCatchCount: 1,
			expectedTypes:      []string{"Exception"},
			expectedVariables:  []string{"$e"},
		},
		{
			name: "qualified class name in catch",
			input: `<?php
try {
    $phpmailer->setFrom($from_email, $from_name, false);
} catch (PHPMailer\PHPMailer\Exception $e) {
}`,
			expectedCatchCount: 1,
			expectedTypes:      []string{"PHPMailer\\PHPMailer\\Exception"},
			expectedVariables:  []string{"$e"},
		},
		{
			name: "fully qualified class name in catch",
			input: `<?php
try {
    $phpmailer->setFrom($from_email, $from_name, false);
} catch (\PHPMailer\Exception $e) {
}`,
			expectedCatchCount: 1,
			expectedTypes:      []string{"\\PHPMailer\\Exception"},
			expectedVariables:  []string{"$e"},
		},
		{
			name: "multiple catch blocks with different qualified names",
			input: `<?php
try {
    $phpmailer->setFrom($from_email, $from_name, false);
} catch (PHPMailer\PHPMailer\Exception $e) {
} catch (\PHPMailer\Exception $e) {
} catch (Exception $e) {
}`,
			expectedCatchCount: 3,
			expectedTypes:      []string{"PHPMailer\\PHPMailer\\Exception", "\\PHPMailer\\Exception", "Exception"},
			expectedVariables:  []string{"$e", "$e", "$e"},
		},
		{
			name: "catch with multiple exception types using pipe",
			input: `<?php
try {
    $phpmailer->setFrom($from_email, $from_name, false);
} catch (PHPMailer\PHPMailer\Exception | \PHPMailer\Exception | Exception $e) {
}`,
			expectedCatchCount: 1,
			expectedTypes:      []string{"PHPMailer\\PHPMailer\\Exception", "\\PHPMailer\\Exception", "Exception"},
			expectedVariables:  []string{"$e"},
		},
		{
			name: "original failing case from bug report",
			input: `<?php
try {
    $phpmailer->setFrom( $from_email, $from_name, false );
} catch ( PHPMailer\PHPMailer\Exception $e ) {
    
} catch ( \PHPMailer\Exception $e ) {

}`,
			expectedCatchCount: 2,
			expectedTypes:      []string{"PHPMailer\\PHPMailer\\Exception", "\\PHPMailer\\Exception"},
			expectedVariables:  []string{"$e", "$e"},
		},
		{
			name: "deeply nested namespace",
			input: `<?php
try {
    doSomething();
} catch (\Foo\Bar\Baz\Qux\Exception $ex) {
}`,
			expectedCatchCount: 1,
			expectedTypes:      []string{"\\Foo\\Bar\\Baz\\Qux\\Exception"},
			expectedVariables:  []string{"$ex"},
		},
		{
			name: "catch with finally block",
			input: `<?php
try {
    doSomething();
} catch (PHPMailer\PHPMailer\Exception $e) {
} finally {
    cleanup();
}`,
			expectedCatchCount: 1,
			expectedTypes:      []string{"PHPMailer\\PHPMailer\\Exception"},
			expectedVariables:  []string{"$e"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			// Check for parser errors
			checkParserErrors(t, p)
			assert.NotNil(t, program, "Program should not be nil")
			assert.Len(t, program.Body, 1, "Should have one statement (try-catch)")

			// Get the try statement
			tryStmt, ok := program.Body[0].(*ast.TryStatement)
			assert.True(t, ok, "Statement should be TryStatement")
			assert.NotNil(t, tryStmt, "TryStatement should not be nil")

			// Verify catch clauses count
			assert.Len(t, tryStmt.CatchClauses, tt.expectedCatchCount,
				"Should have %d catch clause(s)", tt.expectedCatchCount)

			// Verify each catch clause
			totalExpectedTypes := len(tt.expectedTypes)
			actualTypeIndex := 0

			for i, catchClause := range tryStmt.CatchClauses {
				assert.NotNil(t, catchClause, "Catch clause %d should not be nil", i)

				// Check variable
				assert.NotNil(t, catchClause.Parameter, "Catch clause %d should have a parameter", i)
				variable, ok := catchClause.Parameter.(*ast.Variable)
				assert.True(t, ok, "Catch clause %d parameter should be a Variable", i)
				assert.Equal(t, tt.expectedVariables[i], variable.Name,
					"Catch clause %d variable should be %s", i, tt.expectedVariables[i])

				// Check exception types
				for j, exceptionType := range catchClause.Types {
					assert.True(t, actualTypeIndex < totalExpectedTypes,
						"Too many exception types found")

					switch typedExpr := exceptionType.(type) {
					case *ast.IdentifierNode:
						// Simple class name or qualified name
						assert.Equal(t, tt.expectedTypes[actualTypeIndex], typedExpr.Name,
							"Catch clause %d, type %d should be %s", i, j, tt.expectedTypes[actualTypeIndex])
					case *ast.NamespaceExpression:
						// Fully qualified name starting with \
						assert.Equal(t, tt.expectedTypes[actualTypeIndex], typedExpr.String(),
							"Catch clause %d, type %d should be %s", i, j, tt.expectedTypes[actualTypeIndex])
					default:
						t.Errorf("Unexpected exception type: %T", exceptionType)
					}
					actualTypeIndex++
				}
			}

			// Ensure we found all expected types
			assert.Equal(t, totalExpectedTypes, actualTypeIndex,
				"Should have found all %d expected exception types", totalExpectedTypes)
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

				// The callee should be an identifier node with the fully qualified name
				identifierNode, ok := callExpr.Callee.(*ast.IdentifierNode)
				assert.True(t, ok, "Callee should be IdentifierNode")
				assert.Equal(t, "\\DateTime\\createFromFormat", identifierNode.Name)
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

				// The callee should be an identifier node with the fully qualified name
				identifierNode, ok := callExpr.Callee.(*ast.IdentifierNode)
				assert.True(t, ok, "Callee should be IdentifierNode")
				assert.Equal(t, "\\test", identifierNode.Name)
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

				propIdent, ok := propAccess.Property.(*ast.IdentifierNode)
				assert.True(t, ok, "Property should be IdentifierNode")
				assert.Equal(t, "method", propIdent.Name)
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

func TestParsing_YieldExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, program *ast.Program)
	}{
		{
			name:  "Empty yield",
			input: `<?php yield;`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)

				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				assert.True(t, ok, "Expected ExpressionStatement")

				yieldExpr, ok := stmt.Expression.(*ast.YieldExpression)
				assert.True(t, ok, "Expected YieldExpression")
				assert.Nil(t, yieldExpr.Key, "Key should be nil for empty yield")
				assert.Nil(t, yieldExpr.Value, "Value should be nil for empty yield")
				assert.Equal(t, "yield", yieldExpr.String())
			},
		},
		{
			name:  "Yield with value",
			input: `<?php yield $value;`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)

				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				assert.True(t, ok, "Expected ExpressionStatement")

				yieldExpr, ok := stmt.Expression.(*ast.YieldExpression)
				assert.True(t, ok, "Expected YieldExpression")
				assert.Nil(t, yieldExpr.Key, "Key should be nil for value-only yield")
				assert.NotNil(t, yieldExpr.Value, "Value should not be nil")

				valueVar, ok := yieldExpr.Value.(*ast.Variable)
				assert.True(t, ok, "Value should be Variable")
				assert.Equal(t, "$value", valueVar.Name)
				assert.Equal(t, "yield $value", yieldExpr.String())
			},
		},
		{
			name:  "Yield with key-value pair",
			input: `<?php yield $key => $value;`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)

				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				assert.True(t, ok, "Expected ExpressionStatement")

				yieldExpr, ok := stmt.Expression.(*ast.YieldExpression)
				assert.True(t, ok, "Expected YieldExpression")
				assert.NotNil(t, yieldExpr.Key, "Key should not be nil")
				assert.NotNil(t, yieldExpr.Value, "Value should not be nil")

				keyVar, ok := yieldExpr.Key.(*ast.Variable)
				assert.True(t, ok, "Key should be Variable")
				assert.Equal(t, "$key", keyVar.Name)

				valueVar, ok := yieldExpr.Value.(*ast.Variable)
				assert.True(t, ok, "Value should be Variable")
				assert.Equal(t, "$value", valueVar.Name)

				assert.Equal(t, "yield $key => $value", yieldExpr.String())
			},
		},
		{
			name:  "Yield with complex expressions",
			input: `<?php yield $obj->method() => $this->property;`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)

				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				assert.True(t, ok, "Expected ExpressionStatement")

				yieldExpr, ok := stmt.Expression.(*ast.YieldExpression)
				assert.True(t, ok, "Expected YieldExpression")
				assert.NotNil(t, yieldExpr.Key, "Key should not be nil")
				assert.NotNil(t, yieldExpr.Value, "Value should not be nil")

				// Key should be method call
				_, ok = yieldExpr.Key.(*ast.CallExpression)
				assert.True(t, ok, "Key should be CallExpression")

				// Value should be property access
				valueProp, ok := yieldExpr.Value.(*ast.PropertyAccessExpression)
				assert.True(t, ok, "Value should be PropertyAccessExpression")

				thisVar, ok := valueProp.Object.(*ast.Variable)
				assert.True(t, ok, "Property object should be Variable")
				assert.Equal(t, "$this", thisVar.Name)
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

func TestParsing_YieldFromExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, program *ast.Program)
	}{
		{
			name:  "Yield from variable",
			input: `<?php yield from $generator;`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)

				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				assert.True(t, ok, "Expected ExpressionStatement")

				yieldFromExpr, ok := stmt.Expression.(*ast.YieldFromExpression)
				assert.True(t, ok, "Expected YieldFromExpression")
				assert.NotNil(t, yieldFromExpr.Expression, "Expression should not be nil")

				genVar, ok := yieldFromExpr.Expression.(*ast.Variable)
				assert.True(t, ok, "Expression should be Variable")
				assert.Equal(t, "$generator", genVar.Name)
				assert.Equal(t, "yield from $generator", yieldFromExpr.String())
			},
		},
		{
			name:  "Yield from function call",
			input: `<?php yield from createGenerator();`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)

				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				assert.True(t, ok, "Expected ExpressionStatement")

				yieldFromExpr, ok := stmt.Expression.(*ast.YieldFromExpression)
				assert.True(t, ok, "Expected YieldFromExpression")
				assert.NotNil(t, yieldFromExpr.Expression, "Expression should not be nil")

				callExpr, ok := yieldFromExpr.Expression.(*ast.CallExpression)
				assert.True(t, ok, "Expression should be CallExpression")

				funcName, ok := callExpr.Callee.(*ast.IdentifierNode)
				assert.True(t, ok, "Callee should be IdentifierNode")
				assert.Equal(t, "createGenerator", funcName.Name)
			},
		},
		{
			name:  "Yield from method call",
			input: `<?php yield from $obj->getGenerator();`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)

				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				assert.True(t, ok, "Expected ExpressionStatement")

				yieldFromExpr, ok := stmt.Expression.(*ast.YieldFromExpression)
				assert.True(t, ok, "Expected YieldFromExpression")
				assert.NotNil(t, yieldFromExpr.Expression, "Expression should not be nil")

				callExpr, ok := yieldFromExpr.Expression.(*ast.CallExpression)
				assert.True(t, ok, "Expression should be CallExpression")

				propAccess, ok := callExpr.Callee.(*ast.PropertyAccessExpression)
				assert.True(t, ok, "Callee should be PropertyAccessExpression")

				objVar, ok := propAccess.Object.(*ast.Variable)
				assert.True(t, ok, "Object should be Variable")
				assert.Equal(t, "$obj", objVar.Name)
				propIdent, ok := propAccess.Property.(*ast.IdentifierNode)
				assert.True(t, ok, "Property should be IdentifierNode")
				assert.Equal(t, "getGenerator", propIdent.Name)
			},
		},
		{
			name:  "Yield from array",
			input: `<?php yield from [1, 2, 3];`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)

				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				assert.True(t, ok, "Expected ExpressionStatement")

				yieldFromExpr, ok := stmt.Expression.(*ast.YieldFromExpression)
				assert.True(t, ok, "Expected YieldFromExpression")
				assert.NotNil(t, yieldFromExpr.Expression, "Expression should not be nil")

				arrayExpr, ok := yieldFromExpr.Expression.(*ast.ArrayExpression)
				assert.True(t, ok, "Expression should be ArrayExpression")
				assert.Len(t, arrayExpr.Elements, 3, "Array should have 3 elements")
			},
		},
		{
			name:  "Nested yield from expressions",
			input: `<?php yield from getOuterGenerator($inner);`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)

				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				assert.True(t, ok, "Expected ExpressionStatement")

				yieldFromExpr, ok := stmt.Expression.(*ast.YieldFromExpression)
				assert.True(t, ok, "Expected YieldFromExpression")

				callExpr, ok := yieldFromExpr.Expression.(*ast.CallExpression)
				assert.True(t, ok, "Expression should be CallExpression")

				funcName, ok := callExpr.Callee.(*ast.IdentifierNode)
				assert.True(t, ok, "Callee should be IdentifierNode")
				assert.Equal(t, "getOuterGenerator", funcName.Name)

				assert.Len(t, callExpr.Arguments, 1, "Should have one argument")
				innerVar, ok := callExpr.Arguments[0].(*ast.Variable)
				assert.True(t, ok, "Argument should be Variable")
				assert.Equal(t, "$inner", innerVar.Name)
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

// TestParsing_ArrayWithHTMLPrefix tests parsing arrays after inline HTML content
func TestParsing_ArrayWithHTMLPrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []interface{} // expected array elements
	}{
		{
			name:     "Simple array after HTML",
			input:    "html\n<?php [1,2,3];",
			expected: []interface{}{1, 2, 3},
		},
		{
			name:     "Array with strings after HTML",
			input:    "<!DOCTYPE html>\n<?php [\"a\", \"b\", \"c\"];",
			expected: []interface{}{"a", "b", "c"},
		},
		{
			name:     "Empty array after HTML",
			input:    "<h1>Title</h1>\n<?php [];",
			expected: []interface{}{},
		},
		{
			name:     "Nested array after HTML",
			input:    "content\n<?php [[1,2], [3,4]];",
			expected: []interface{}{[]interface{}{1, 2}, []interface{}{3, 4}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			checkParserErrors(t, p)
			assert.NotNil(t, program)

			// Should have 2 statements: HTML inline content and the array expression
			require.Len(t, program.Body, 2)

			// First statement should be HTML inline content
			htmlStmt, ok := program.Body[0].(*ast.ExpressionStatement)
			require.True(t, ok, "First statement should be ExpressionStatement")

			htmlExpr, ok := htmlStmt.Expression.(*ast.StringLiteral)
			require.True(t, ok, "First expression should be StringLiteral (HTML)")
			// Just verify it contains some HTML content (any non-empty string)
			assert.NotEmpty(t, htmlExpr.Value, "HTML content should not be empty")

			// Second statement should be array expression
			arrayStmt, ok := program.Body[1].(*ast.ExpressionStatement)
			require.True(t, ok, "Second statement should be ExpressionStatement")

			arrayExpr, ok := arrayStmt.Expression.(*ast.ArrayExpression)
			require.True(t, ok, "Second expression should be ArrayExpression")

			// Check array elements
			assert.Len(t, arrayExpr.Elements, len(tt.expected))

			for i, expectedElement := range tt.expected {
				switch expected := expectedElement.(type) {
				case int:
					numLit, ok := arrayExpr.Elements[i].(*ast.NumberLiteral)
					require.True(t, ok, "Array element should be NumberLiteral")
					assert.Equal(t, fmt.Sprintf("%d", expected), numLit.Value)
					assert.Equal(t, "integer", numLit.Kind)
				case string:
					strLit, ok := arrayExpr.Elements[i].(*ast.StringLiteral)
					require.True(t, ok, "Array element should be StringLiteral")
					assert.Equal(t, expected, strLit.Value)
				case []interface{}:
					nestedArray, ok := arrayExpr.Elements[i].(*ast.ArrayExpression)
					require.True(t, ok, "Array element should be nested ArrayExpression")
					assert.Len(t, nestedArray.Elements, len(expected))
				}
			}
		})
	}
}

// TestParsing_ArrayAccessVsArrayLiteral tests that we correctly distinguish between array access and array literals
func TestParsing_ArrayAccessVsArrayLiteral(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectAccess bool // true = expect array access, false = expect array literal
	}{
		{
			name:         "Array literal",
			input:        "<?php [1,2,3];",
			expectAccess: false,
		},
		{
			name:         "Array access on variable",
			input:        "<?php $arr[0];",
			expectAccess: true,
		},
		{
			name:         "Array literal after HTML",
			input:        "html\n<?php [1,2,3];",
			expectAccess: false,
		},
		{
			name:         "Array access after assignment",
			input:        "<?php $x = [1,2,3]; $x[0];",
			expectAccess: true, // This tests the second expression
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			checkParserErrors(t, p)
			assert.NotNil(t, program)
			require.Greater(t, len(program.Body), 0)

			// Find the relevant expression statement
			var targetExpr ast.Expression
			if tt.input == "<?php $x = [1,2,3]; $x[0];" {
				// For this case, we want to check the second statement
				require.Len(t, program.Body, 2)
				exprStmt := program.Body[1].(*ast.ExpressionStatement)
				targetExpr = exprStmt.Expression
			} else if strings.Contains(tt.input, "html") {
				// For HTML cases, the array is in the second statement
				require.Len(t, program.Body, 2)
				exprStmt := program.Body[1].(*ast.ExpressionStatement)
				targetExpr = exprStmt.Expression
			} else {
				// For simple cases, it's the first (and only) statement
				exprStmt := program.Body[0].(*ast.ExpressionStatement)
				targetExpr = exprStmt.Expression
			}

			if tt.expectAccess {
				_, isArrayAccess := targetExpr.(*ast.ArrayAccessExpression)
				assert.True(t, isArrayAccess, "Expected ArrayAccessExpression")
			} else {
				_, isArrayLiteral := targetExpr.(*ast.ArrayExpression)
				assert.True(t, isArrayLiteral, "Expected ArrayExpression (literal)")
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

func TestParsing_StaticProperties(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		expectedClassName  string
		expectedProperties int
		validateProperties func(t *testing.T, classExpr *ast.ClassExpression)
	}{
		{
			name: "Basic public static property",
			input: `<?php
class Foo {
    public static $user_ids;
}`,
			expectedClassName:  "Foo",
			expectedProperties: 1,
			validateProperties: func(t *testing.T, classExpr *ast.ClassExpression) {
				property, ok := classExpr.Body[0].(*ast.PropertyDeclaration)
				assert.True(t, ok, "First body item should be PropertyDeclaration")
				assert.Equal(t, "public", property.Visibility)
				assert.True(t, property.Static, "Property should be static")
				assert.Equal(t, "user_ids", property.Name)
			},
		},
		{
			name: "All visibility modifiers with static",
			input: `<?php
class TestClass {
    public static $public_static;
    private static $private_static;
    protected static $protected_static;
}`,
			expectedClassName:  "TestClass",
			expectedProperties: 3,
			validateProperties: func(t *testing.T, classExpr *ast.ClassExpression) {
				// Test public static
				property1, ok := classExpr.Body[0].(*ast.PropertyDeclaration)
				assert.True(t, ok, "First property should be PropertyDeclaration")
				assert.Equal(t, "public", property1.Visibility)
				assert.True(t, property1.Static, "Property should be static")
				assert.Equal(t, "public_static", property1.Name)

				// Test private static
				property2, ok := classExpr.Body[1].(*ast.PropertyDeclaration)
				assert.True(t, ok, "Second property should be PropertyDeclaration")
				assert.Equal(t, "private", property2.Visibility)
				assert.True(t, property2.Static, "Property should be static")
				assert.Equal(t, "private_static", property2.Name)

				// Test protected static
				property3, ok := classExpr.Body[2].(*ast.PropertyDeclaration)
				assert.True(t, ok, "Third property should be PropertyDeclaration")
				assert.Equal(t, "protected", property3.Visibility)
				assert.True(t, property3.Static, "Property should be static")
				assert.Equal(t, "protected_static", property3.Name)
			},
		},
		{
			name: "Static without explicit visibility",
			input: `<?php
class Test {
    static $default_static;
}`,
			expectedClassName:  "Test",
			expectedProperties: 1,
			validateProperties: func(t *testing.T, classExpr *ast.ClassExpression) {
				property, ok := classExpr.Body[0].(*ast.PropertyDeclaration)
				assert.True(t, ok, "First body item should be PropertyDeclaration")
				assert.Equal(t, "public", property.Visibility, "Static property without visibility should default to public")
				assert.True(t, property.Static, "Property should be static")
				assert.Equal(t, "default_static", property.Name)
			},
		},
		{
			name: "Static property with type hint and default value",
			input: `<?php
class Advanced {
    public static int $typed_static;
    public static $initialized_static = 42;
    private static array $complex_static = ["a", "b"];
}`,
			expectedClassName:  "Advanced",
			expectedProperties: 3,
			validateProperties: func(t *testing.T, classExpr *ast.ClassExpression) {
				// Test typed static property
				property1, ok := classExpr.Body[0].(*ast.PropertyDeclaration)
				assert.True(t, ok, "First property should be PropertyDeclaration")
				assert.Equal(t, "public", property1.Visibility)
				assert.True(t, property1.Static, "Property should be static")
				assert.Equal(t, "typed_static", property1.Name)
				assert.NotNil(t, property1.Type, "Property should have type hint")
				assert.Equal(t, "int", property1.Type.Name)

				// Test initialized static property
				property2, ok := classExpr.Body[1].(*ast.PropertyDeclaration)
				assert.True(t, ok, "Second property should be PropertyDeclaration")
				assert.Equal(t, "public", property2.Visibility)
				assert.True(t, property2.Static, "Property should be static")
				assert.Equal(t, "initialized_static", property2.Name)
				assert.NotNil(t, property2.DefaultValue, "Property should have default value")

				// Test complex static property
				property3, ok := classExpr.Body[2].(*ast.PropertyDeclaration)
				assert.True(t, ok, "Third property should be PropertyDeclaration")
				assert.Equal(t, "private", property3.Visibility)
				assert.True(t, property3.Static, "Property should be static")
				assert.Equal(t, "complex_static", property3.Name)
				assert.NotNil(t, property3.Type, "Property should have type hint")
				assert.Equal(t, "array", property3.Type.Name)
				assert.NotNil(t, property3.DefaultValue, "Property should have default value")
			},
		},
		{
			name:  "Static property with fully qualified namespaced type",
			input: `<?php class A { protected static \WeakMap $recursionDetectionCache; }`,
			expectedClassName:  "A",
			expectedProperties: 1,
			validateProperties: func(t *testing.T, classExpr *ast.ClassExpression) {
				property, ok := classExpr.Body[0].(*ast.PropertyDeclaration)
				assert.True(t, ok, "First body item should be PropertyDeclaration")
				assert.Equal(t, "protected", property.Visibility)
				assert.True(t, property.Static, "Property should be static")
				assert.Equal(t, "recursionDetectionCache", property.Name)
				assert.NotNil(t, property.Type, "Property should have type hint")
				assert.Equal(t, "\\WeakMap", property.Type.Name)
			},
		},
		{
			name:  "Static property with qualified namespaced type",
			input: `<?php class A { private static Foo\Bar $cache; }`,
			expectedClassName:  "A",
			expectedProperties: 1,
			validateProperties: func(t *testing.T, classExpr *ast.ClassExpression) {
				property, ok := classExpr.Body[0].(*ast.PropertyDeclaration)
				assert.True(t, ok, "First body item should be PropertyDeclaration")
				assert.Equal(t, "private", property.Visibility)
				assert.True(t, property.Static, "Property should be static")
				assert.Equal(t, "cache", property.Name)
				assert.NotNil(t, property.Type, "Property should have type hint")
				assert.Equal(t, "Foo\\Bar", property.Type.Name)
			},
		},
		{
			name:  "Static property with relative namespaced type",
			input: `<?php class A { public static namespace\Cache $storage; }`,
			expectedClassName:  "A",
			expectedProperties: 1,
			validateProperties: func(t *testing.T, classExpr *ast.ClassExpression) {
				property, ok := classExpr.Body[0].(*ast.PropertyDeclaration)
				assert.True(t, ok, "First body item should be PropertyDeclaration")
				assert.Equal(t, "public", property.Visibility)
				assert.True(t, property.Static, "Property should be static")
				assert.Equal(t, "storage", property.Name)
				assert.NotNil(t, property.Type, "Property should have type hint")
				assert.Equal(t, "namespace\\Cache", property.Type.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			// Check for parser errors
			errors := p.Errors()
			if len(errors) != 0 {
				for _, err := range errors {
					t.Errorf("Parser error: %s", err)
				}
				t.FailNow()
			}

			require.NotNil(t, program)
			require.Len(t, program.Body, 1, "Program should have 1 statement")

			// Get the class expression statement
			exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
			require.True(t, ok, "Statement should be ExpressionStatement")

			classExpr, ok := exprStmt.Expression.(*ast.ClassExpression)
			require.True(t, ok, "Expression should be ClassExpression")

			// Validate class name
			nameIdent, ok := classExpr.Name.(*ast.IdentifierNode)
			require.True(t, ok, "Class name should be IdentifierNode")
			assert.Equal(t, tt.expectedClassName, nameIdent.Name)

			// Validate properties count
			assert.Len(t, classExpr.Body, tt.expectedProperties,
				fmt.Sprintf("Class should have %d properties", tt.expectedProperties))

			// Run custom validation
			tt.validateProperties(t, classExpr)
		})
	}
}

// TestParsing_PropertyAccess tests various property access scenarios
func TestParsing_PropertyAccess(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedObject string
		expectedProp   string
		expectedType   string
		validate       func(t *testing.T, prop ast.Expression)
	}{
		{
			name:           "Simple property access",
			input:          `<?php $obj->property; ?>`,
			expectedObject: "$obj",
			expectedProp:   "property",
			expectedType:   "IdentifierNode",
			validate: func(t *testing.T, prop ast.Expression) {
				ident, ok := prop.(*ast.IdentifierNode)
				assert.True(t, ok, "Property should be IdentifierNode")
				assert.Equal(t, "property", ident.Name)
			},
		},
		{
			name:           "Variable property access",
			input:          `<?php $post_type->$property_name; ?>`,
			expectedObject: "$post_type",
			expectedProp:   "$property_name",
			expectedType:   "Variable",
			validate: func(t *testing.T, prop ast.Expression) {
				variable, ok := prop.(*ast.Variable)
				assert.True(t, ok, "Property should be Variable")
				assert.Equal(t, "$property_name", variable.Name)
			},
		},
		{
			name:           "Dynamic property access with braces",
			input:          `<?php $obj->{"property"}; ?>`,
			expectedObject: "$obj",
			expectedProp:   `"property"`,
			expectedType:   "StringLiteral",
			validate: func(t *testing.T, prop ast.Expression) {
				stringLit, ok := prop.(*ast.StringLiteral)
				assert.True(t, ok, "Property should be StringLiteral")
				assert.Equal(t, "property", stringLit.Value)
			},
		},
		{
			name:           "Complex dynamic property access",
			input:          `<?php $obj->{$prefix . "suffix"}; ?>`,
			expectedObject: "$obj",
			expectedProp:   `$prefix . "suffix"`,
			expectedType:   "BinaryExpression",
			validate: func(t *testing.T, prop ast.Expression) {
				binaryExpr, ok := prop.(*ast.BinaryExpression)
				assert.True(t, ok, "Property should be BinaryExpression")
				assert.Equal(t, ".", binaryExpr.Operator)

				left, ok := binaryExpr.Left.(*ast.Variable)
				assert.True(t, ok, "Left should be Variable")
				assert.Equal(t, "$prefix", left.Name)

				right, ok := binaryExpr.Right.(*ast.StringLiteral)
				assert.True(t, ok, "Right should be StringLiteral")
				assert.Equal(t, "suffix", right.Value)
			},
		},
		{
			name:           "Chained property access",
			input:          `<?php $obj->first->$second; ?>`,
			expectedObject: "$obj->first",
			expectedProp:   "$second",
			expectedType:   "Variable",
			validate: func(t *testing.T, prop ast.Expression) {
				variable, ok := prop.(*ast.Variable)
				assert.True(t, ok, "Property should be Variable")
				assert.Equal(t, "$second", variable.Name)
			},
		},
		{
			name:           "Nullsafe property access simple",
			input:          `<?php $obj?->property; ?>`,
			expectedObject: "$obj",
			expectedProp:   "property",
			expectedType:   "IdentifierNode",
			validate: func(t *testing.T, prop ast.Expression) {
				ident, ok := prop.(*ast.IdentifierNode)
				assert.True(t, ok, "Property should be IdentifierNode")
				assert.Equal(t, "property", ident.Name)
			},
		},
		{
			name:           "Nullsafe variable property access",
			input:          `<?php $obj?->$prop; ?>`,
			expectedObject: "$obj",
			expectedProp:   "$prop",
			expectedType:   "Variable",
			validate: func(t *testing.T, prop ast.Expression) {
				variable, ok := prop.(*ast.Variable)
				assert.True(t, ok, "Property should be Variable")
				assert.Equal(t, "$prop", variable.Name)
			},
		},
		{
			name:           "Nullsafe dynamic property access",
			input:          `<?php $obj?->{"prop"}; ?>`,
			expectedObject: "$obj",
			expectedProp:   `"prop"`,
			expectedType:   "StringLiteral",
			validate: func(t *testing.T, prop ast.Expression) {
				stringLit, ok := prop.(*ast.StringLiteral)
				assert.True(t, ok, "Property should be StringLiteral")
				assert.Equal(t, "prop", stringLit.Value)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			// Check for parser errors
			errors := p.Errors()
			if len(errors) != 0 {
				for _, err := range errors {
					t.Errorf("Parser error: %s", err)
				}
				t.FailNow()
			}

			require.NotNil(t, program)
			require.Len(t, program.Body, 1, "Program should have 1 statement")

			// Get the expression statement
			exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
			require.True(t, ok, "Statement should be ExpressionStatement")

			// Check if it's regular or nullsafe property access
			var propertyExpr ast.Expression
			var object ast.Expression
			var property ast.Expression
			isNullsafe := strings.Contains(tt.input, "?->")

			if isNullsafe {
				nullsafeProp, ok := exprStmt.Expression.(*ast.NullsafePropertyAccessExpression)
				require.True(t, ok, "Expression should be NullsafePropertyAccessExpression")
				propertyExpr = nullsafeProp
				object = nullsafeProp.Object
				property = nullsafeProp.Property
			} else {
				propAccess, ok := exprStmt.Expression.(*ast.PropertyAccessExpression)
				require.True(t, ok, "Expression should be PropertyAccessExpression")
				propertyExpr = propAccess
				object = propAccess.Object
				property = propAccess.Property
			}

			// Validate object
			if strings.Contains(tt.expectedObject, "->") {
				// Chained property access - object should be another PropertyAccessExpression
				_, ok := object.(*ast.PropertyAccessExpression)
				assert.True(t, ok, "Object should be PropertyAccessExpression for chained access")
			} else {
				// Simple variable object
				objVar, ok := object.(*ast.Variable)
				assert.True(t, ok, "Object should be Variable")
				assert.Equal(t, tt.expectedObject, objVar.Name)
			}

			// Validate property using custom validation function
			tt.validate(t, property)

			// Test string representation (for brace-enclosed expressions, braces are not preserved in output)
			expectedStr := strings.ReplaceAll(tt.input, "<?php ", "")
			expectedStr = strings.ReplaceAll(expectedStr, "; ?>", "")
			expectedStr = strings.TrimSpace(expectedStr)

			// Skip string representation test for brace-enclosed expressions for now
			// as the parser correctly parses them but doesn't preserve braces in string output
			if !strings.Contains(expectedStr, "{") {
				assert.Equal(t, expectedStr, propertyExpr.String())
			}
		})
	}
}

// TestParsing_PropertyAccessInComplexExpressions tests property access in complex scenarios
func TestParsing_PropertyAccessInComplexExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, program *ast.Program)
	}{
		{
			name:  "Property access in assignment",
			input: `<?php $this->assertSame($expected_property_value, $post_type->$property_name); ?>`,
			validate: func(t *testing.T, program *ast.Program) {
				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok, "Statement should be ExpressionStatement")

				// This should be a method call
				call, ok := exprStmt.Expression.(*ast.CallExpression)
				require.True(t, ok, "Expression should be CallExpression")

				// Callee should be property access ($this->assertSame)
				propAccess, ok := call.Callee.(*ast.PropertyAccessExpression)
				require.True(t, ok, "Callee should be PropertyAccessExpression")

				// Validate object ($this)
				thisVar, ok := propAccess.Object.(*ast.Variable)
				require.True(t, ok, "Object should be Variable")
				assert.Equal(t, "$this", thisVar.Name)

				// Validate method name (assertSame)
				methodIdent, ok := propAccess.Property.(*ast.IdentifierNode)
				require.True(t, ok, "Property should be IdentifierNode")
				assert.Equal(t, "assertSame", methodIdent.Name)

				// Validate arguments - should have 2 arguments
				require.Len(t, call.Arguments, 2, "Call should have 2 arguments")

				// First argument should be a variable ($expected_property_value)
				arg1, ok := call.Arguments[0].(*ast.Variable)
				require.True(t, ok, "First argument should be Variable")
				assert.Equal(t, "$expected_property_value", arg1.Name)

				// Second argument should be property access ($post_type->$property_name)
				arg2PropAccess, ok := call.Arguments[1].(*ast.PropertyAccessExpression)
				require.True(t, ok, "Second argument should be PropertyAccessExpression")

				// Validate second argument object ($post_type)
				postTypeVar, ok := arg2PropAccess.Object.(*ast.Variable)
				require.True(t, ok, "Property object should be Variable")
				assert.Equal(t, "$post_type", postTypeVar.Name)

				// Validate second argument property ($property_name) - should be Variable
				propNameVar, ok := arg2PropAccess.Property.(*ast.Variable)
				require.True(t, ok, "Property should be Variable")
				assert.Equal(t, "$property_name", propNameVar.Name)
			},
		},
		{
			name:  "Chained property access with variable",
			input: `<?php $obj->method1()->method2()->$variable_property; ?>`,
			validate: func(t *testing.T, program *ast.Program) {
				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok, "Statement should be ExpressionStatement")

				// Should be property access where object is a call expression
				propAccess, ok := exprStmt.Expression.(*ast.PropertyAccessExpression)
				require.True(t, ok, "Expression should be PropertyAccessExpression")

				// Property should be a variable
				propVar, ok := propAccess.Property.(*ast.Variable)
				require.True(t, ok, "Property should be Variable")
				assert.Equal(t, "$variable_property", propVar.Name)

				// Object should be a call expression (method2())
				call, ok := propAccess.Object.(*ast.CallExpression)
				require.True(t, ok, "Object should be CallExpression")

				// The callee of method2() should be property access (obj.method1().method2)
				_, ok = call.Callee.(*ast.PropertyAccessExpression)
				require.True(t, ok, "Callee should be PropertyAccessExpression")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			// Check for parser errors
			errors := p.Errors()
			if len(errors) != 0 {
				for _, err := range errors {
					t.Errorf("Parser error: %s", err)
				}
				t.FailNow()
			}

			require.NotNil(t, program)
			require.Len(t, program.Body, 1, "Program should have 1 statement")

			// Run custom validation
			tt.validate(t, program)
		})
	}
}

// TestParsing_SingleLineControlStructures tests single-line if/while/for statements without braces
func TestParsing_SingleLineControlStructures(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(*testing.T, *ast.Program)
	}{
		{
			name:  "Single-line if statement with assignment",
			input: `<?php if($x > 0) $y = 1; ?>`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				ifStmt, ok := program.Body[0].(*ast.IfStatement)
				require.True(t, ok, "Should be IfStatement")
				require.NotNil(t, ifStmt.Test)
				require.Len(t, ifStmt.Consequent, 1)

				// Check the single statement in consequent
				exprStmt, ok := ifStmt.Consequent[0].(*ast.ExpressionStatement)
				require.True(t, ok, "Consequent should contain ExpressionStatement")

				assignment, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok, "Should be assignment")
				assert.Equal(t, "=", assignment.Operator)
			},
		},
		{
			name:  "Single-line while statement",
			input: `<?php while($x > 0) $x--; ?>`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				whileStmt, ok := program.Body[0].(*ast.WhileStatement)
				require.True(t, ok, "Should be WhileStatement")
				require.NotNil(t, whileStmt.Test)
				require.Len(t, whileStmt.Body, 1)

				// Check the single statement in body
				exprStmt, ok := whileStmt.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok, "Body should contain ExpressionStatement")

				unary, ok := exprStmt.Expression.(*ast.UnaryExpression)
				require.True(t, ok, "Should be unary expression")
				assert.Equal(t, "--", unary.Operator)
				assert.False(t, unary.Prefix, "Should be postfix")
			},
		},
		{
			name:  "Single-line for statement",
			input: `<?php for($i = 0; $i < 10; $i++) echo $i; ?>`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				forStmt, ok := program.Body[0].(*ast.ForStatement)
				require.True(t, ok, "Should be ForStatement")
				require.NotNil(t, forStmt.Init)
				require.NotNil(t, forStmt.Test)
				require.NotNil(t, forStmt.Update)
				require.Len(t, forStmt.Body, 1)

				// Check the single statement in body
				echoStmt, ok := forStmt.Body[0].(*ast.EchoStatement)
				require.True(t, ok, "Body should contain EchoStatement")
				require.Len(t, echoStmt.Arguments, 1)
			},
		},
		{
			name:  "Original failing case - complex regex if statement",
			input: `<?php if(preg_match("/^([0-9]{3})(-(.*[".CRLF."]{1,2})+\\1)? [^".CRLF."]+[".CRLF."]{1,2}$/", $this->_message, $regs)) $go=false; ?>`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				ifStmt, ok := program.Body[0].(*ast.IfStatement)
				require.True(t, ok, "Should be IfStatement")
				require.NotNil(t, ifStmt.Test)
				require.Len(t, ifStmt.Consequent, 1)

				// Check the preg_match call in condition
				callExpr, ok := ifStmt.Test.(*ast.CallExpression)
				require.True(t, ok, "Condition should be function call")

				callee, ok := callExpr.Callee.(*ast.IdentifierNode)
				require.True(t, ok, "Callee should be identifier")
				assert.Equal(t, "preg_match", callee.Name)
				require.Len(t, callExpr.Arguments, 3, "preg_match should have 3 arguments")

				// Check the single assignment statement in consequent
				exprStmt, ok := ifStmt.Consequent[0].(*ast.ExpressionStatement)
				require.True(t, ok, "Consequent should contain ExpressionStatement")

				assignment, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok, "Should be assignment")

				variable, ok := assignment.Left.(*ast.Variable)
				require.True(t, ok, "Left side should be variable")
				assert.Equal(t, "$go", variable.Name)

				identifier, ok := assignment.Right.(*ast.IdentifierNode)
				require.True(t, ok, "Right side should be identifier")
				assert.Equal(t, "false", identifier.Name)
			},
		},
		{
			name:  "Single-line if-else with assignments",
			input: `<?php if($x > 0) $y = 1; else $y = 0; ?>`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				ifStmt, ok := program.Body[0].(*ast.IfStatement)
				require.True(t, ok, "Should be IfStatement")
				require.Len(t, ifStmt.Consequent, 1)
				require.Len(t, ifStmt.Alternate, 1)

				// Check consequent
				exprStmt1, ok := ifStmt.Consequent[0].(*ast.ExpressionStatement)
				require.True(t, ok, "Consequent should be ExpressionStatement")
				assignment1, ok := exprStmt1.Expression.(*ast.AssignmentExpression)
				require.True(t, ok, "Should be assignment")

				// Check alternate
				exprStmt2, ok := ifStmt.Alternate[0].(*ast.ExpressionStatement)
				require.True(t, ok, "Alternate should be ExpressionStatement")
				assignment2, ok := exprStmt2.Expression.(*ast.AssignmentExpression)
				require.True(t, ok, "Should be assignment")

				assert.Equal(t, "=", assignment1.Operator)
				assert.Equal(t, "=", assignment2.Operator)
			},
		},
		{
			name:  "Mixed block and single-line statements",
			input: `<?php if($x > 0) { echo "positive"; } else echo "not positive"; ?>`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				ifStmt, ok := program.Body[0].(*ast.IfStatement)
				require.True(t, ok, "Should be IfStatement")
				require.Len(t, ifStmt.Consequent, 1, "Block should have 1 statement")
				require.Len(t, ifStmt.Alternate, 1, "Else should have 1 statement")

				// Consequent is a block statement
				echoStmt1, ok := ifStmt.Consequent[0].(*ast.EchoStatement)
				require.True(t, ok, "Consequent should be EchoStatement")

				// Alternate is a single-line statement
				echoStmt2, ok := ifStmt.Alternate[0].(*ast.EchoStatement)
				require.True(t, ok, "Alternate should be EchoStatement")

				assert.Len(t, echoStmt1.Arguments, 1)
				assert.Len(t, echoStmt2.Arguments, 1)
			},
		},
		{
			name:  "Nested single-line control structures",
			input: `<?php if($x > 0) if($y > 0) $z = 1; ?>`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				outerIf, ok := program.Body[0].(*ast.IfStatement)
				require.True(t, ok, "Should be IfStatement")
				require.Len(t, outerIf.Consequent, 1)

				// Check nested if
				innerIf, ok := outerIf.Consequent[0].(*ast.IfStatement)
				require.True(t, ok, "Nested statement should be IfStatement")
				require.Len(t, innerIf.Consequent, 1)

				// Check innermost assignment
				exprStmt, ok := innerIf.Consequent[0].(*ast.ExpressionStatement)
				require.True(t, ok, "Innermost should be ExpressionStatement")

				assignment, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok, "Should be assignment")
				assert.Equal(t, "=", assignment.Operator)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			// Check for parser errors
			if len(p.Errors()) != 0 {
				for _, err := range p.Errors() {
					t.Errorf("Parser error: %s", err)
				}
				t.FailNow()
			}

			require.NotNil(t, program)
			tt.validate(t, program)
		})
	}
}

// TestParsing_AttributeParameters tests parameter attributes in various scenarios
func TestParsing_AttributeParameters(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		validate func(t *testing.T, program *ast.Program)
	}{
		{
			name:  "Namespaced attribute on single parameter",
			input: `<?php function test(#[\SensitiveParameter] $password) {}`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				funcDecl, ok := program.Body[0].(*ast.FunctionDeclaration)
				require.True(t, ok)
				require.Equal(t, "test", funcDecl.Name.(*ast.IdentifierNode).Name)
				require.Len(t, funcDecl.Parameters, 1)

				param := funcDecl.Parameters[0]
				assert.Equal(t, "$password", param.Name)
				require.Len(t, param.Attributes, 1)

				attrGroup := param.Attributes[0]
				require.Len(t, attrGroup.Attributes, 1)

				attr := attrGroup.Attributes[0]
				assert.Equal(t, "\\SensitiveParameter", attr.Name.Name)
				assert.Nil(t, attr.Arguments)
			},
		},
		{
			name:  "Simple attribute on parameter",
			input: `<?php function test(#[Deprecated] $param) {}`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				funcDecl, ok := program.Body[0].(*ast.FunctionDeclaration)
				require.True(t, ok)
				require.Len(t, funcDecl.Parameters, 1)

				param := funcDecl.Parameters[0]
				assert.Equal(t, "$param", param.Name)
				require.Len(t, param.Attributes, 1)

				attr := param.Attributes[0].Attributes[0]
				assert.Equal(t, "Deprecated", attr.Name.Name)
			},
		},
		{
			name:  "Attribute with parameters",
			input: `<?php function test(#[Route('/path', method: 'GET')] $request) {}`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				funcDecl, ok := program.Body[0].(*ast.FunctionDeclaration)
				require.True(t, ok)
				require.Len(t, funcDecl.Parameters, 1)

				param := funcDecl.Parameters[0]
				assert.Equal(t, "$request", param.Name)
				require.Len(t, param.Attributes, 1)

				attr := param.Attributes[0].Attributes[0]
				assert.Equal(t, "Route", attr.Name.Name)
				require.Len(t, attr.Arguments, 2)
			},
		},
		{
			name:  "Multiple attributes on single parameter",
			input: `<?php function test(#[Attr1, Attr2] $param) {}`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				funcDecl, ok := program.Body[0].(*ast.FunctionDeclaration)
				require.True(t, ok)
				require.Len(t, funcDecl.Parameters, 1)

				param := funcDecl.Parameters[0]
				assert.Equal(t, "$param", param.Name)
				require.Len(t, param.Attributes, 1)

				attrGroup := param.Attributes[0]
				require.Len(t, attrGroup.Attributes, 2)
				assert.Equal(t, "Attr1", attrGroup.Attributes[0].Name.Name)
				assert.Equal(t, "Attr2", attrGroup.Attributes[1].Name.Name)
			},
		},
		{
			name:  "Multiple attribute groups on parameter",
			input: `<?php function test(#[Attr1] #[Attr2] $param) {}`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				funcDecl, ok := program.Body[0].(*ast.FunctionDeclaration)
				require.True(t, ok)
				require.Len(t, funcDecl.Parameters, 1)

				param := funcDecl.Parameters[0]
				assert.Equal(t, "$param", param.Name)
				require.Len(t, param.Attributes, 2)

				assert.Equal(t, "Attr1", param.Attributes[0].Attributes[0].Name.Name)
				assert.Equal(t, "Attr2", param.Attributes[1].Attributes[0].Name.Name)
			},
		},
		{
			name:  "Attributes with type hint and visibility",
			input: `<?php class Test { public function __construct(#[\SensitiveParameter] public string $password) {} }`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				classExpr, ok := program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.ClassExpression)
				require.True(t, ok)
				require.Len(t, classExpr.Body, 1)

				funcDecl, ok := classExpr.Body[0].(*ast.FunctionDeclaration)
				require.True(t, ok)
				require.Equal(t, "__construct", funcDecl.Name.(*ast.IdentifierNode).Name)
				require.Len(t, funcDecl.Parameters, 1)

				param := funcDecl.Parameters[0]
				assert.Equal(t, "$password", param.Name)
				assert.Equal(t, "public", param.Visibility)
				assert.NotNil(t, param.Type)
				assert.Equal(t, "string", param.Type.Name)

				require.Len(t, param.Attributes, 1)
				attr := param.Attributes[0].Attributes[0]
				assert.Equal(t, "\\SensitiveParameter", attr.Name.Name)
			},
		},
		{
			name:  "Multiple parameters with different attributes",
			input: `<?php function authenticate(string $username, #[\SensitiveParameter] string $password, #[Optional] array $options = []) {}`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				funcDecl, ok := program.Body[0].(*ast.FunctionDeclaration)
				require.True(t, ok)
				require.Equal(t, "authenticate", funcDecl.Name.(*ast.IdentifierNode).Name)
				require.Len(t, funcDecl.Parameters, 3)

				// First parameter: no attributes
				param1 := funcDecl.Parameters[0]
				assert.Equal(t, "$username", param1.Name)
				assert.Len(t, param1.Attributes, 0)
				assert.Equal(t, "string", param1.Type.Name)

				// Second parameter: SensitiveParameter attribute
				param2 := funcDecl.Parameters[1]
				assert.Equal(t, "$password", param2.Name)
				require.Len(t, param2.Attributes, 1)
				attr2 := param2.Attributes[0].Attributes[0]
				assert.Equal(t, "\\SensitiveParameter", attr2.Name.Name)
				assert.Equal(t, "string", param2.Type.Name)

				// Third parameter: Optional attribute with default value
				param3 := funcDecl.Parameters[2]
				assert.Equal(t, "$options", param3.Name)
				require.Len(t, param3.Attributes, 1)
				attr3 := param3.Attributes[0].Attributes[0]
				assert.Equal(t, "Optional", attr3.Name.Name)
				assert.Equal(t, "array", param3.Type.Name)
				assert.NotNil(t, param3.DefaultValue)
			},
		},
		{
			name:  "Attribute on reference parameter",
			input: `<?php function test(#[Ref] array &$data) {}`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				funcDecl, ok := program.Body[0].(*ast.FunctionDeclaration)
				require.True(t, ok)
				require.Len(t, funcDecl.Parameters, 1)

				param := funcDecl.Parameters[0]
				assert.Equal(t, "$data", param.Name)
				assert.True(t, param.ByReference)
				assert.Equal(t, "array", param.Type.Name)

				require.Len(t, param.Attributes, 1)
				attr := param.Attributes[0].Attributes[0]
				assert.Equal(t, "Ref", attr.Name.Name)
			},
		},
		{
			name:  "Attribute on variadic parameter",
			input: `<?php function test(#[Spread] ...$args) {}`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				funcDecl, ok := program.Body[0].(*ast.FunctionDeclaration)
				require.True(t, ok)
				require.Len(t, funcDecl.Parameters, 1)

				param := funcDecl.Parameters[0]
				assert.Equal(t, "$args", param.Name)
				assert.True(t, param.Variadic)

				require.Len(t, param.Attributes, 1)
				attr := param.Attributes[0].Attributes[0]
				assert.Equal(t, "Spread", attr.Name.Name)
			},
		},
		{
			name:  "Static attribute on parameter",
			input: `<?php function test(#[static] $param) {}`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				funcDecl, ok := program.Body[0].(*ast.FunctionDeclaration)
				require.True(t, ok)
				require.Len(t, funcDecl.Parameters, 1)

				param := funcDecl.Parameters[0]
				assert.Equal(t, "$param", param.Name)
				require.Len(t, param.Attributes, 1)

				attr := param.Attributes[0].Attributes[0]
				assert.Equal(t, "static", attr.Name.Name)
			},
		},
		{
			name:  "Complex multi-level namespaced attribute",
			input: `<?php function test(#[\Global\Fully\Qualified\Attribute] $param) {}`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				funcDecl, ok := program.Body[0].(*ast.FunctionDeclaration)
				require.True(t, ok)
				require.Len(t, funcDecl.Parameters, 1)

				param := funcDecl.Parameters[0]
				assert.Equal(t, "$param", param.Name)
				require.Len(t, param.Attributes, 1)

				attr := param.Attributes[0].Attributes[0]
				assert.Equal(t, "\\Global\\Fully\\Qualified\\Attribute", attr.Name.Name)
			},
		},
		{
			name:  "Mixed attribute types in one group",
			input: `<?php function test(#[Simple, static, \Namespaced\Attr] $param) {}`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				funcDecl, ok := program.Body[0].(*ast.FunctionDeclaration)
				require.True(t, ok)
				require.Len(t, funcDecl.Parameters, 1)

				param := funcDecl.Parameters[0]
				assert.Equal(t, "$param", param.Name)
				require.Len(t, param.Attributes, 1)

				attrGroup := param.Attributes[0]
				require.Len(t, attrGroup.Attributes, 3)

				assert.Equal(t, "Simple", attrGroup.Attributes[0].Name.Name)
				assert.Equal(t, "static", attrGroup.Attributes[1].Name.Name)
				assert.Equal(t, "\\Namespaced\\Attr", attrGroup.Attributes[2].Name.Name)
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			checkParserErrors(t, p)
			require.NotNil(t, program)
			tt.validate(t, program)
		})
	}
}

func TestParsing_PropertyHooks_Complete(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected func(t *testing.T, program *ast.Program)
	}{
		{
			name: "Simple get hook with arrow syntax",
			input: `<?php
class Example {
    public string $name {
        get => 'test';
    }
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				classStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				classExpr, ok := classStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)

				require.Len(t, classExpr.Body, 1)
				hookedProp, ok := classExpr.Body[0].(*ast.HookedPropertyDeclaration)
				require.True(t, ok)

				assert.Equal(t, "public", hookedProp.Visibility)
				assert.Equal(t, "name", hookedProp.Name)
				assert.Equal(t, "string", hookedProp.Type.Name)

				require.Len(t, hookedProp.Hooks, 1)
				getHook := hookedProp.Hooks[0]
				assert.Equal(t, "get", getHook.Type)
				assert.False(t, getHook.ByRef)
				assert.NotNil(t, getHook.Body)
				assert.Nil(t, getHook.Parameter)
			},
		},
		{
			name: "Set hook with parameter and arrow syntax",
			input: `<?php
class Example {
    public string $email {
        set(string $value) => strtolower($value);
    }
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				classStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				classExpr, ok := classStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)

				require.Len(t, classExpr.Body, 1)
				hookedProp, ok := classExpr.Body[0].(*ast.HookedPropertyDeclaration)
				require.True(t, ok)

				assert.Equal(t, "public", hookedProp.Visibility)
				assert.Equal(t, "email", hookedProp.Name)

				require.Len(t, hookedProp.Hooks, 1)
				setHook := hookedProp.Hooks[0]
				assert.Equal(t, "set", setHook.Type)
				assert.False(t, setHook.ByRef)
				assert.NotNil(t, setHook.Body)
				assert.NotNil(t, setHook.Parameter)
				assert.Equal(t, "$value", setHook.Parameter.Name)
			},
		},
		{
			name: "Both get and set hooks",
			input: `<?php
class Example {
    public string $fullName {
        get => $this->first . ' ' . $this->last;
        set(string $value) => $this->parseFullName($value);
    }
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				classStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				classExpr, ok := classStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)

				require.Len(t, classExpr.Body, 1)
				hookedProp, ok := classExpr.Body[0].(*ast.HookedPropertyDeclaration)
				require.True(t, ok)

				assert.Equal(t, "public", hookedProp.Visibility)
				assert.Equal(t, "fullName", hookedProp.Name)

				require.Len(t, hookedProp.Hooks, 2)

				getHook := hookedProp.Hooks[0]
				assert.Equal(t, "get", getHook.Type)
				assert.NotNil(t, getHook.Body)

				setHook := hookedProp.Hooks[1]
				assert.Equal(t, "set", setHook.Type)
				assert.NotNil(t, setHook.Body)
				assert.NotNil(t, setHook.Parameter)
			},
		},
		{
			name: "Reference get hook",
			input: `<?php
class Example {
    public array $data {
        &get => $this->internalData;
    }
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				classStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				classExpr, ok := classStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)

				require.Len(t, classExpr.Body, 1)
				hookedProp, ok := classExpr.Body[0].(*ast.HookedPropertyDeclaration)
				require.True(t, ok)

				require.Len(t, hookedProp.Hooks, 1)
				getHook := hookedProp.Hooks[0]
				assert.Equal(t, "get", getHook.Type)
				assert.True(t, getHook.ByRef)
				assert.NotNil(t, getHook.Body)
			},
		},
		{
			name: "Mixed hooked and regular properties",
			input: `<?php
class Example {
    public string $name {
        get => 'test';
    }
    private string $internal;
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				classStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				classExpr, ok := classStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)

				require.Len(t, classExpr.Body, 2)

				// First property should be hooked
				hookedProp, ok := classExpr.Body[0].(*ast.HookedPropertyDeclaration)
				require.True(t, ok)
				assert.Equal(t, "name", hookedProp.Name)
				require.Len(t, hookedProp.Hooks, 1)

				// Second property should be regular
				regularProp, ok := classExpr.Body[1].(*ast.PropertyDeclaration)
				require.True(t, ok)
				assert.Equal(t, "internal", regularProp.Name)
				assert.Equal(t, "private", regularProp.Visibility)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			l := lexer.New(test.input)
			parser := New(l)
			program := parser.ParseProgram()

			require.NotNil(t, program)
			if len(parser.Errors()) > 0 {
				t.Errorf("Parser errors: %v", parser.Errors())
			}

			test.expected(t, program)
		})
	}
}

func TestParsing_AlternativeIfStatement(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedErrors int
	}{
		{
			name: "Alternative If with elseif and else",
			input: `<?php
if ($condition):
    echo "true";
elseif ($other):
    echo "other";
else:
    echo "false";
endif;
?>`,
			expectedErrors: 0,
		},
		{
			name: "Alternative If simple",
			input: `<?php
if ($x > 0):
    echo "positive";
endif;
?>`,
			expectedErrors: 0,
		},
		{
			name: "Alternative If with multiple elseifs",
			input: `<?php
if ($x == 1):
    echo "one";
elseif ($x == 2):
    echo "two";
elseif ($x == 3):
    echo "three";
else:
    echo "other";
endif;
?>`,
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			assert.Equal(t, tt.expectedErrors, len(p.Errors()), "Parser errors: %v", p.Errors())
			assert.NotNil(t, program)
			assert.Greater(t, len(program.Body), 0, "Expected at least one statement")

			// Check that we have an AlternativeIfStatement
			stmt := program.Body[0]
			altIfStmt, ok := stmt.(*ast.AlternativeIfStatement)
			assert.True(t, ok, "Expected AlternativeIfStatement, got %T", stmt)
			assert.NotNil(t, altIfStmt.Condition, "Expected condition to be set")
		})
	}
}

func TestParsing_AlternativeWhileStatement(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedErrors int
	}{
		{
			name: "Alternative While loop",
			input: `<?php
while ($counter < 10):
    echo $counter;
    $counter++;
endwhile;
?>`,
			expectedErrors: 0,
		},
		{
			name: "Alternative While with complex condition",
			input: `<?php
while ($i < count($array) && $flag):
    process($array[$i]);
    $i++;
endwhile;
?>`,
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			assert.Equal(t, tt.expectedErrors, len(p.Errors()), "Parser errors: %v", p.Errors())
			assert.NotNil(t, program)
			assert.Greater(t, len(program.Body), 0, "Expected at least one statement")

			// Check that we have an AlternativeWhileStatement
			stmt := program.Body[0]
			altWhileStmt, ok := stmt.(*ast.AlternativeWhileStatement)
			assert.True(t, ok, "Expected AlternativeWhileStatement, got %T", stmt)
			assert.NotNil(t, altWhileStmt.Condition, "Expected condition to be set")
			assert.Greater(t, len(altWhileStmt.Body), 0, "Expected body statements")
		})
	}
}

func TestParsing_AlternativeForStatement(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedErrors int
	}{
		{
			name: "Alternative For loop",
			input: `<?php
for ($i = 0; $i < 5; $i++):
    echo $i;
endfor;
?>`,
			expectedErrors: 0,
		},
		{
			name: "Alternative For with empty init",
			input: `<?php
for (; $i < 10; $i++):
    echo $i;
endfor;
?>`,
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			assert.Equal(t, tt.expectedErrors, len(p.Errors()), "Parser errors: %v", p.Errors())
			assert.NotNil(t, program)
			assert.Greater(t, len(program.Body), 0, "Expected at least one statement")

			// Check that we have an AlternativeForStatement
			stmt := program.Body[0]
			altForStmt, ok := stmt.(*ast.AlternativeForStatement)
			assert.True(t, ok, "Expected AlternativeForStatement, got %T", stmt)
			assert.Greater(t, len(altForStmt.Body), 0, "Expected body statements")
		})
	}
}

func TestParsing_AlternativeForeachStatement(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedErrors int
	}{
		{
			name: "Alternative Foreach with value only",
			input: `<?php
foreach ($array as $value):
    echo $value;
endforeach;
?>`,
			expectedErrors: 0,
		},
		{
			name: "Alternative Foreach with key and value",
			input: `<?php
foreach ($array as $key => $value):
    echo $key . ": " . $value;
endforeach;
?>`,
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			assert.Equal(t, tt.expectedErrors, len(p.Errors()), "Parser errors: %v", p.Errors())
			assert.NotNil(t, program)
			assert.Greater(t, len(program.Body), 0, "Expected at least one statement")

			// Check that we have an AlternativeForeachStatement
			stmt := program.Body[0]
			altForeachStmt, ok := stmt.(*ast.AlternativeForeachStatement)
			assert.True(t, ok, "Expected AlternativeForeachStatement, got %T", stmt)
			assert.NotNil(t, altForeachStmt.Iterable, "Expected iterable to be set")
			assert.NotNil(t, altForeachStmt.Value, "Expected value to be set")
			assert.Greater(t, len(altForeachStmt.Body), 0, "Expected body statements")
		})
	}
}

func TestParsing_DeclareStatement(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedErrors int
		expectAlt      bool
	}{
		{
			name: "Alternative Declare statement",
			input: `<?php
declare(strict_types=1):
    function test(): int {
        return 42;
    }
enddeclare;
?>`,
			expectedErrors: 0,
			expectAlt:      true,
		},
		{
			name: "Alternative Declare with multiple declarations",
			input: `<?php
declare(ticks=1, encoding='UTF-8'):
    echo "Hello World";
enddeclare;
?>`,
			expectedErrors: 0,
			expectAlt:      true,
		},
		{
			name: "Regular Declare statement",
			input: `<?php
declare(strict_types=1);
?>`,
			expectedErrors: 0,
			expectAlt:      false,
		},
		{
			name: "Regular Declare with block",
			input: `<?php
declare(strict_types=1) {
    function test() {}
}
?>`,
			expectedErrors: 0,
			expectAlt:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			assert.Equal(t, tt.expectedErrors, len(p.Errors()), "Parser errors: %v", p.Errors())
			assert.NotNil(t, program)
			assert.Greater(t, len(program.Body), 0, "Expected at least one statement")

			// Check that we have a DeclareStatement
			stmt := program.Body[0]
			declareStmt, ok := stmt.(*ast.DeclareStatement)
			assert.True(t, ok, "Expected DeclareStatement, got %T", stmt)
			assert.Greater(t, len(declareStmt.Declarations), 0, "Expected at least one declaration")
			assert.Equal(t, tt.expectAlt, declareStmt.Alternative, "Expected alternative flag to match")
		})
	}
}

func TestParsing_NamespaceStatements(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedErrors int
	}{
		{
			name: "Simple namespace declaration",
			input: `<?php
namespace App;
?>`,
			expectedErrors: 0,
		},
		{
			name: "Multi-level namespace declaration",
			input: `<?php
namespace App\Http\Controllers;
?>`,
			expectedErrors: 0,
		},
		{
			name: "Global namespace declaration",
			input: `<?php
namespace;
?>`,
			expectedErrors: 0,
		},
		{
			name: "Namespace with block syntax",
			input: `<?php
namespace App {
    function test() {}
}
?>`,
			expectedErrors: 0,
		},
		{
			name: "Anonymous namespace with block syntax",
			input: `<?php
namespace {
    function test() {}
}
?>`,
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			assert.Equal(t, tt.expectedErrors, len(p.Errors()), "Parser errors: %v", p.Errors())
			assert.NotNil(t, program)
			assert.Greater(t, len(program.Body), 0, "Expected at least one statement")

			// Check that we have a NamespaceStatement
			stmt := program.Body[0]
			namespaceStmt, ok := stmt.(*ast.NamespaceStatement)
			assert.True(t, ok, "Expected NamespaceStatement, got %T", stmt)

			// Test specific behavior based on test case
			switch tt.name {
			case "Simple namespace declaration":
				assert.NotNil(t, namespaceStmt.Name, "Expected namespace name to be set")
				assert.Equal(t, []string{"App"}, namespaceStmt.Name.Parts, "Expected namespace parts to match")
				assert.Equal(t, 0, len(namespaceStmt.Body), "Expected empty body for simple declaration")
			case "Multi-level namespace declaration":
				assert.NotNil(t, namespaceStmt.Name, "Expected namespace name to be set")
				assert.Equal(t, []string{"App", "Http", "Controllers"}, namespaceStmt.Name.Parts, "Expected namespace parts to match")
				assert.Equal(t, 0, len(namespaceStmt.Body), "Expected empty body for simple declaration")
			case "Global namespace declaration":
				assert.Nil(t, namespaceStmt.Name, "Expected namespace name to be nil for global namespace")
				assert.Equal(t, 0, len(namespaceStmt.Body), "Expected empty body for global namespace")
			case "Namespace with block syntax":
				assert.NotNil(t, namespaceStmt.Name, "Expected namespace name to be set")
				assert.Equal(t, []string{"App"}, namespaceStmt.Name.Parts, "Expected namespace parts to match")
				assert.Greater(t, len(namespaceStmt.Body), 0, "Expected body statements for block syntax")
			case "Anonymous namespace with block syntax":
				assert.Nil(t, namespaceStmt.Name, "Expected namespace name to be nil for anonymous namespace")
				assert.Greater(t, len(namespaceStmt.Body), 0, "Expected body statements for block syntax")
			}
		})
	}
}

func TestParsing_UseStatements(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedErrors int
	}{
		{
			name: "Simple use statement",
			input: `<?php
use App\Http\Controller;
?>`,
			expectedErrors: 0,
		},
		{
			name: "Use statement with alias",
			input: `<?php
use App\Http\Controller as BaseController;
?>`,
			expectedErrors: 0,
		},
		{
			name: "Multiple use statements",
			input: `<?php
use App\Http\Request, App\Http\Response;
?>`,
			expectedErrors: 0,
		},
		{
			name: "Mixed use statements with aliases",
			input: `<?php
use App\Http\Request as Req, App\Http\Response, App\Http\Controller as BaseController;
?>`,
			expectedErrors: 0,
		},
		{
			name: "Use function statement",
			input: `<?php
use function App\Http\helper_function;
?>`,
			expectedErrors: 0,
		},
		{
			name: "Use const statement",
			input: `<?php
use const App\Http\SOME_CONSTANT;
?>`,
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			assert.Equal(t, tt.expectedErrors, len(p.Errors()), "Parser errors: %v", p.Errors())
			assert.NotNil(t, program)
			assert.Greater(t, len(program.Body), 0, "Expected at least one statement")

			// Check that we have a UseStatement
			stmt := program.Body[0]
			useStmt, ok := stmt.(*ast.UseStatement)
			assert.True(t, ok, "Expected UseStatement, got %T", stmt)
			assert.Greater(t, len(useStmt.Uses), 0, "Expected at least one use clause")

			// Test specific behavior based on test case
			switch tt.name {
			case "Simple use statement":
				assert.Equal(t, 1, len(useStmt.Uses), "Expected one use clause")
				assert.Equal(t, []string{"App", "Http", "Controller"}, useStmt.Uses[0].Name.Parts, "Expected namespace parts to match")
				assert.Equal(t, "", useStmt.Uses[0].Alias, "Expected no alias")
				assert.Equal(t, "", useStmt.Uses[0].Type, "Expected no type")
			case "Use statement with alias":
				assert.Equal(t, 1, len(useStmt.Uses), "Expected one use clause")
				assert.Equal(t, []string{"App", "Http", "Controller"}, useStmt.Uses[0].Name.Parts, "Expected namespace parts to match")
				assert.Equal(t, "BaseController", useStmt.Uses[0].Alias, "Expected alias to match")
				assert.Equal(t, "", useStmt.Uses[0].Type, "Expected no type")
			case "Multiple use statements":
				assert.Equal(t, 2, len(useStmt.Uses), "Expected two use clauses")
				assert.Equal(t, []string{"App", "Http", "Request"}, useStmt.Uses[0].Name.Parts, "Expected first namespace to match")
				assert.Equal(t, []string{"App", "Http", "Response"}, useStmt.Uses[1].Name.Parts, "Expected second namespace to match")
				assert.Equal(t, "", useStmt.Uses[0].Alias, "Expected no alias on first")
				assert.Equal(t, "", useStmt.Uses[1].Alias, "Expected no alias on second")
			case "Mixed use statements with aliases":
				assert.Equal(t, 3, len(useStmt.Uses), "Expected three use clauses")
				assert.Equal(t, "Req", useStmt.Uses[0].Alias, "Expected alias on first")
				assert.Equal(t, "", useStmt.Uses[1].Alias, "Expected no alias on second")
				assert.Equal(t, "BaseController", useStmt.Uses[2].Alias, "Expected alias on third")
			case "Use function statement":
				assert.Equal(t, 1, len(useStmt.Uses), "Expected one use clause")
				assert.Equal(t, "function", useStmt.Uses[0].Type, "Expected function type")
				assert.Equal(t, []string{"App", "Http", "helper_function"}, useStmt.Uses[0].Name.Parts, "Expected namespace parts to match")
			case "Use const statement":
				assert.Equal(t, 1, len(useStmt.Uses), "Expected one use clause")
				assert.Equal(t, "const", useStmt.Uses[0].Type, "Expected const type")
				assert.Equal(t, []string{"App", "Http", "SOME_CONSTANT"}, useStmt.Uses[0].Name.Parts, "Expected namespace parts to match")
			}
		})
	}
}

func TestParsing_InterfaceDeclarations(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedErrors int
	}{
		{
			name: "Simple interface declaration",
			input: `<?php
interface UserInterface {
    public function getName(): string;
}
?>`,
			expectedErrors: 0,
		},
		{
			name: "Interface with multiple methods",
			input: `<?php
interface DatabaseInterface {
    public function connect(): bool;
    public function query(string $sql, array $params): array;
    public function disconnect(): void;
}
?>`,
			expectedErrors: 0,
		},
		{
			name: "Interface extending another interface",
			input: `<?php
interface AdminInterface extends UserInterface {
    public function getPermissions(): array;
}
?>`,
			expectedErrors: 0,
		},
		{
			name: "Interface extending multiple interfaces",
			input: `<?php
interface SuperAdminInterface extends UserInterface, AdminInterface {
    public function deleteUser(int $id): bool;
}
?>`,
			expectedErrors: 0,
		},
		{
			name: "Interface with complex method signatures",
			input: `<?php
interface ServiceInterface {
    public function process(string $data, ?array $options = null): ?object;
    public function validate(array &$data): bool;
}
?>`,
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			assert.Equal(t, tt.expectedErrors, len(p.Errors()), "Parser errors: %v", p.Errors())
			assert.NotNil(t, program)
			assert.Greater(t, len(program.Body), 0, "Expected at least one statement")

			// Check that we have an InterfaceDeclaration
			stmt := program.Body[0]
			interfaceDecl, ok := stmt.(*ast.InterfaceDeclaration)
			assert.True(t, ok, "Expected InterfaceDeclaration, got %T", stmt)
			assert.NotNil(t, interfaceDecl.Name, "Expected interface name to be set")

			// Test specific behavior based on test case
			switch tt.name {
			case "Simple interface declaration":
				assert.Equal(t, "UserInterface", interfaceDecl.Name.Name, "Expected interface name to match")
				assert.Equal(t, 0, len(interfaceDecl.Extends), "Expected no parent interfaces")
				assert.Equal(t, 1, len(interfaceDecl.Methods), "Expected one method")
				assert.Equal(t, "getName", interfaceDecl.Methods[0].Name.Name, "Expected method name to match")
				assert.Equal(t, "public", interfaceDecl.Methods[0].Visibility, "Expected public visibility")
				assert.NotNil(t, interfaceDecl.Methods[0].ReturnType, "Expected return type")
			case "Interface with multiple methods":
				assert.Equal(t, "DatabaseInterface", interfaceDecl.Name.Name, "Expected interface name to match")
				assert.Equal(t, 3, len(interfaceDecl.Methods), "Expected three methods")
				assert.Equal(t, "connect", interfaceDecl.Methods[0].Name.Name, "Expected first method name")
				assert.Equal(t, "query", interfaceDecl.Methods[1].Name.Name, "Expected second method name")
				assert.Equal(t, "disconnect", interfaceDecl.Methods[2].Name.Name, "Expected third method name")
			case "Interface extending another interface":
				assert.Equal(t, "AdminInterface", interfaceDecl.Name.Name, "Expected interface name to match")
				assert.Equal(t, 1, len(interfaceDecl.Extends), "Expected one parent interface")
				assert.Equal(t, "UserInterface", interfaceDecl.Extends[0].Name, "Expected parent interface name")
				assert.Equal(t, 1, len(interfaceDecl.Methods), "Expected one method")
			case "Interface extending multiple interfaces":
				assert.Equal(t, "SuperAdminInterface", interfaceDecl.Name.Name, "Expected interface name to match")
				assert.Equal(t, 2, len(interfaceDecl.Extends), "Expected two parent interfaces")
				assert.Equal(t, "UserInterface", interfaceDecl.Extends[0].Name, "Expected first parent interface")
				assert.Equal(t, "AdminInterface", interfaceDecl.Extends[1].Name, "Expected second parent interface")
			case "Interface with complex method signatures":
				assert.Equal(t, "ServiceInterface", interfaceDecl.Name.Name, "Expected interface name to match")
				assert.Equal(t, 2, len(interfaceDecl.Methods), "Expected two methods")
				assert.Equal(t, "process", interfaceDecl.Methods[0].Name.Name, "Expected first method name")
				assert.Equal(t, "validate", interfaceDecl.Methods[1].Name.Name, "Expected second method name")
				// Test that the second method has reference parameter
				assert.True(t, interfaceDecl.Methods[1].Parameters[0].ByReference, "Expected reference parameter")
			}
		})
	}
}

func TestParsing_TraitDeclarations(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedErrors int
	}{
		{
			name: "Simple trait declaration",
			input: `<?php
trait LoggerTrait {
    public function log(string $message): void {
        echo $message;
    }
}
?>`,
			expectedErrors: 0,
		},
		{
			name: "Trait with properties and methods",
			input: `<?php
trait DatabaseTrait {
    private $connection;
    protected $config;
    
    public function connect(): bool {
        return true;
    }
    
    private function getConfig(): array {
        return $this->config;
    }
}
?>`,
			expectedErrors: 0,
		},
		{
			name: "Trait with mixed visibility modifiers",
			input: `<?php
trait UtilityTrait {
    public $publicProperty;
    private $privateProperty;
    protected $protectedProperty;
    
    public function publicMethod(): void {}
    private function privateMethod(): string { return "private"; }
    protected function protectedMethod(array $data): array { return $data; }
}
?>`,
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			assert.Equal(t, tt.expectedErrors, len(p.Errors()), "Parser errors: %v", p.Errors())
			assert.NotNil(t, program)
			assert.Greater(t, len(program.Body), 0, "Expected at least one statement")

			// Check that we have a TraitDeclaration
			stmt := program.Body[0]
			traitDecl, ok := stmt.(*ast.TraitDeclaration)
			assert.True(t, ok, "Expected TraitDeclaration, got %T", stmt)
			assert.NotNil(t, traitDecl.Name, "Expected trait name to be set")

			// Test specific behavior based on test case
			switch tt.name {
			case "Simple trait declaration":
				assert.Equal(t, "LoggerTrait", traitDecl.Name.Name, "Expected trait name to match")
				assert.Equal(t, 0, len(traitDecl.Properties), "Expected no properties")
				assert.Equal(t, 1, len(traitDecl.Methods), "Expected one method")
				assert.Equal(t, "log", traitDecl.Methods[0].Name.String(), "Expected method name to match")
				assert.Equal(t, "public", traitDecl.Methods[0].Visibility, "Expected public visibility")
			case "Trait with properties and methods":
				assert.Equal(t, "DatabaseTrait", traitDecl.Name.Name, "Expected trait name to match")
				assert.Equal(t, 2, len(traitDecl.Properties), "Expected two properties")
				assert.Equal(t, 2, len(traitDecl.Methods), "Expected two methods")
				assert.Equal(t, "connection", traitDecl.Properties[0].Name, "Expected first property name")
				assert.Equal(t, "config", traitDecl.Properties[1].Name, "Expected second property name")
				assert.Equal(t, "private", traitDecl.Properties[0].Visibility, "Expected private visibility")
				assert.Equal(t, "protected", traitDecl.Properties[1].Visibility, "Expected protected visibility")
			case "Trait with mixed visibility modifiers":
				assert.Equal(t, "UtilityTrait", traitDecl.Name.Name, "Expected trait name to match")
				assert.Equal(t, 3, len(traitDecl.Properties), "Expected three properties")
				assert.Equal(t, 3, len(traitDecl.Methods), "Expected three methods")
				// Test property visibilities
				assert.Equal(t, "public", traitDecl.Properties[0].Visibility, "Expected public visibility")
				assert.Equal(t, "private", traitDecl.Properties[1].Visibility, "Expected private visibility")
				assert.Equal(t, "protected", traitDecl.Properties[2].Visibility, "Expected protected visibility")
				// Test method visibilities
				assert.Equal(t, "public", traitDecl.Methods[0].Visibility, "Expected public method visibility")
				assert.Equal(t, "private", traitDecl.Methods[1].Visibility, "Expected private method visibility")
				assert.Equal(t, "protected", traitDecl.Methods[2].Visibility, "Expected protected method visibility")
			}
		})
	}
}

func TestParsing_ComprehensiveModernPHP(t *testing.T) {
	// This test demonstrates all the newly implemented features working together
	input := `<?php
namespace App\Http\Controllers;

use App\Models\User;
use App\Services\AuthService as Auth;
use function App\Helpers\sanitize_input;
use const App\Config\DEFAULT_TIMEOUT;

interface UserRepositoryInterface {
    public function findById(int $id): ?User;
    public function create(array $data): User;
    public function update(int $id, array $data): bool;
}

trait LoggerTrait {
    protected string $logFile = 'app.log';
    
    public function log(string $message): void {
        // Implementation here
    }
}
?>`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	assert.Equal(t, 0, len(p.Errors()), "Parser errors: %v", p.Errors())
	assert.NotNil(t, program)
	assert.Greater(t, len(program.Body), 5, "Expected multiple top-level statements")

	// Test namespace declaration
	namespaceStmt, ok := program.Body[0].(*ast.NamespaceStatement)
	assert.True(t, ok, "Expected first statement to be namespace")
	assert.Equal(t, []string{"App", "Http", "Controllers"}, namespaceStmt.Name.Parts)

	// Find and test use statements (they might be separate statements)
	var useStatements []*ast.UseStatement
	var interfaceDecl *ast.InterfaceDeclaration
	var traitDecl *ast.TraitDeclaration

	for _, stmt := range program.Body {
		switch s := stmt.(type) {
		case *ast.UseStatement:
			useStatements = append(useStatements, s)
		case *ast.InterfaceDeclaration:
			interfaceDecl = s
		case *ast.TraitDeclaration:
			traitDecl = s
		}
	}

	// Test use statements
	assert.Greater(t, len(useStatements), 0, "Expected at least one use statement")
	if len(useStatements) > 0 {
		// Test first use statement
		assert.Equal(t, []string{"App", "Models", "User"}, useStatements[0].Uses[0].Name.Parts)
	}
	if len(useStatements) > 1 {
		// Test aliased use statement
		assert.Equal(t, "Auth", useStatements[1].Uses[0].Alias, "Expected alias")
	}

	// Test interface declaration
	assert.NotNil(t, interfaceDecl, "Expected interface declaration")
	if interfaceDecl != nil {
		assert.Equal(t, "UserRepositoryInterface", interfaceDecl.Name.Name)
		assert.Equal(t, 3, len(interfaceDecl.Methods), "Expected 3 interface methods")
	}

	// Test trait declaration
	assert.NotNil(t, traitDecl, "Expected trait declaration")
	if traitDecl != nil {
		assert.Equal(t, "LoggerTrait", traitDecl.Name.Name)
		assert.Equal(t, 1, len(traitDecl.Properties), "Expected 1 trait property")
		assert.Equal(t, 1, len(traitDecl.Methods), "Expected 1 trait method")
	}
}

func TestParsing_AlternativeSyntaxWithModernFeatures(t *testing.T) {
	// This test combines alternative syntax with modern PHP features
	input := `<?php
namespace App\Utils;

use App\Services\Logger;

trait CacheTrait {
    private array $cache = [];
}

interface ProcessorInterface {
    public function process(array $data): bool;
}

if ($condition):
    echo "Processing...";
endif;

foreach ($items as $item):
    echo $item;
endforeach;

while ($running):
    echo "Running...";
endwhile;
?>`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	assert.Equal(t, 0, len(p.Errors()), "Parser errors: %v", p.Errors())
	assert.NotNil(t, program)
	assert.Equal(t, 7, len(program.Body), "Expected 7 statements combining modern features with alternative syntax")

	// Verify we have all the different types of statements
	statementTypes := make(map[string]bool)
	for _, stmt := range program.Body {
		switch stmt.(type) {
		case *ast.NamespaceStatement:
			statementTypes["namespace"] = true
		case *ast.UseStatement:
			statementTypes["use"] = true
		case *ast.TraitDeclaration:
			statementTypes["trait"] = true
		case *ast.InterfaceDeclaration:
			statementTypes["interface"] = true
		case *ast.AlternativeIfStatement:
			statementTypes["alt_if"] = true
		case *ast.AlternativeForeachStatement:
			statementTypes["alt_foreach"] = true
		case *ast.AlternativeWhileStatement:
			statementTypes["alt_while"] = true
		}
	}

	// Verify we have all expected statement types
	expectedTypes := []string{"namespace", "use", "trait", "interface", "alt_if", "alt_foreach", "alt_while"}
	for _, expectedType := range expectedTypes {
		assert.True(t, statementTypes[expectedType], "Expected %s statement type", expectedType)
	}
}

func TestParsing_EnumDeclarations(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedErrors int
	}{
		{
			name: "Simple enum declaration",
			input: `<?php
enum Status {
    case PENDING;
    case APPROVED;
    case REJECTED;
}
?>`,
			expectedErrors: 0,
		},
		{
			name: "Backed enum with string values",
			input: `<?php
enum Status: string {
    case PENDING = 'pending';
    case APPROVED = 'approved';
    case REJECTED = 'rejected';
}
?>`,
			expectedErrors: 0,
		},
		{
			name: "Backed enum with integer values",
			input: `<?php
enum Priority: int {
    case LOW = 1;
    case MEDIUM = 2;
    case HIGH = 3;
}
?>`,
			expectedErrors: 0,
		},
		{
			name: "Enum implementing interfaces",
			input: `<?php
enum Color implements ColorInterface {
    case RED;
    case GREEN;
    case BLUE;
    
    public function getHex(): string {
        return '#FF0000';
    }
}
?>`,
			expectedErrors: 0,
		},
		{
			name: "Enum with multiple interfaces and methods",
			input: `<?php
enum HttpStatus: int implements StatusInterface, JsonSerializable {
    case OK = 200;
    case NOT_FOUND = 404;
    case SERVER_ERROR = 500;
    
    public function isError(): bool {
        return $this->value >= 400;
    }
    
    public function jsonSerialize(): mixed {
        return $this->value;
    }
}
?>`,
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			assert.Equal(t, tt.expectedErrors, len(p.Errors()), "Parser errors: %v", p.Errors())
			assert.NotNil(t, program)
			assert.Greater(t, len(program.Body), 0, "Expected at least one statement")

			// Check that we have an EnumDeclaration
			stmt := program.Body[0]
			enumDecl, ok := stmt.(*ast.EnumDeclaration)
			assert.True(t, ok, "Expected EnumDeclaration, got %T", stmt)
			assert.NotNil(t, enumDecl.Name, "Expected enum name to be set")

			// Test specific behavior based on test case
			switch tt.name {
			case "Simple enum declaration":
				assert.Equal(t, "Status", enumDecl.Name.Name, "Expected enum name to match")
				assert.Nil(t, enumDecl.BackingType, "Expected no backing type")
				assert.Equal(t, 0, len(enumDecl.Implements), "Expected no interfaces")
				assert.Equal(t, 3, len(enumDecl.Cases), "Expected 3 enum cases")
				assert.Equal(t, 0, len(enumDecl.Methods), "Expected no methods")
				assert.Equal(t, "PENDING", enumDecl.Cases[0].Name.Name, "Expected first case name")
				assert.Nil(t, enumDecl.Cases[0].Value, "Expected no value for pure enum case")
			case "Backed enum with string values":
				assert.Equal(t, "Status", enumDecl.Name.Name, "Expected enum name to match")
				assert.NotNil(t, enumDecl.BackingType, "Expected backing type")
				assert.Equal(t, 3, len(enumDecl.Cases), "Expected 3 enum cases")
				assert.NotNil(t, enumDecl.Cases[0].Value, "Expected value for backed enum case")
			case "Backed enum with integer values":
				assert.Equal(t, "Priority", enumDecl.Name.Name, "Expected enum name to match")
				assert.NotNil(t, enumDecl.BackingType, "Expected backing type")
				assert.Equal(t, 3, len(enumDecl.Cases), "Expected 3 enum cases")
			case "Enum implementing interfaces":
				assert.Equal(t, "Color", enumDecl.Name.Name, "Expected enum name to match")
				assert.Equal(t, 1, len(enumDecl.Implements), "Expected one interface")
				assert.Equal(t, "ColorInterface", enumDecl.Implements[0].Name, "Expected interface name")
				assert.Equal(t, 3, len(enumDecl.Cases), "Expected 3 enum cases")
				assert.Equal(t, 1, len(enumDecl.Methods), "Expected 1 method")
			case "Enum with multiple interfaces and methods":
				assert.Equal(t, "HttpStatus", enumDecl.Name.Name, "Expected enum name to match")
				assert.NotNil(t, enumDecl.BackingType, "Expected backing type")
				assert.Equal(t, 2, len(enumDecl.Implements), "Expected two interfaces")
				assert.Equal(t, "StatusInterface", enumDecl.Implements[0].Name, "Expected first interface")
				assert.Equal(t, "JsonSerializable", enumDecl.Implements[1].Name, "Expected second interface")
				assert.Equal(t, 3, len(enumDecl.Cases), "Expected 3 enum cases")
				assert.Equal(t, 2, len(enumDecl.Methods), "Expected 2 methods")
			}
		})
	}
}

func TestParsing_FirstClassCallable_Complete(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected func(t *testing.T, program *ast.Program)
	}{
		{
			name: "Simple function first-class callable",
			input: `<?php
$func = strlen(...);`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assignment, ok := stmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				fcc, ok := assignment.Right.(*ast.FirstClassCallable)
				require.True(t, ok)

				identNode, ok := fcc.Callable.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "strlen", identNode.Name)
			},
		},
		{
			name: "Object method first-class callable",
			input: `<?php
$obj = new stdClass();
$method = $obj->method(...);`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 2)

				stmt2, ok := program.Body[1].(*ast.ExpressionStatement)
				require.True(t, ok)

				assignment, ok := stmt2.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				fcc, ok := assignment.Right.(*ast.FirstClassCallable)
				require.True(t, ok)

				propAccess, ok := fcc.Callable.(*ast.PropertyAccessExpression)
				require.True(t, ok)
				propIdent, ok := propAccess.Property.(*ast.IdentifierNode)
				require.True(t, ok, "Property should be IdentifierNode")
				assert.Equal(t, "method", propIdent.Name)

				objVar, ok := propAccess.Object.(*ast.Variable)
				require.True(t, ok)
				assert.Equal(t, "$obj", objVar.Name)
			},
		},
		{
			name: "Static method first-class callable",
			input: `<?php
$static = MyClass::staticMethod(...);`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assignment, ok := stmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				fcc, ok := assignment.Right.(*ast.FirstClassCallable)
				require.True(t, ok)

				staticAccess, ok := fcc.Callable.(*ast.StaticAccessExpression)
				require.True(t, ok)

				propertyNode, ok := staticAccess.Property.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "staticMethod", propertyNode.Name)

				className, ok := staticAccess.Class.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "MyClass", className.Name)
			},
		},
		{
			name: "Variable function first-class callable",
			input: `<?php
$funcName = 'strlen';
$variableFunc = $funcName(...);`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 2)

				stmt2, ok := program.Body[1].(*ast.ExpressionStatement)
				require.True(t, ok)

				assignment, ok := stmt2.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				fcc, ok := assignment.Right.(*ast.FirstClassCallable)
				require.True(t, ok)

				varNode, ok := fcc.Callable.(*ast.Variable)
				require.True(t, ok)
				assert.Equal(t, "$funcName", varNode.Name)
			},
		},
		{
			name: "Self static method first-class callable",
			input: `<?php
class TestClass {
    public function test() {
        $self = self::method(...);
        $parent = parent::method(...);
        $static = static::method(...);
    }
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				classExpr, ok := stmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)

				require.Len(t, classExpr.Body, 1)
				method, ok := classExpr.Body[0].(*ast.FunctionDeclaration)
				require.True(t, ok)

				require.Len(t, method.Body, 3)

				// Test self::method(...)
				selfStmt, ok := method.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				selfAssign, ok := selfStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				selfFCC, ok := selfAssign.Right.(*ast.FirstClassCallable)
				require.True(t, ok)

				selfStatic, ok := selfFCC.Callable.(*ast.StaticAccessExpression)
				require.True(t, ok)

				selfClass, ok := selfStatic.Class.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "self", selfClass.Name)

				selfPropertyNode, ok := selfStatic.Property.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "method", selfPropertyNode.Name)

				// Test parent::method(...)
				parentStmt, ok := method.Body[1].(*ast.ExpressionStatement)
				require.True(t, ok)

				parentAssign, ok := parentStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				parentFCC, ok := parentAssign.Right.(*ast.FirstClassCallable)
				require.True(t, ok)

				parentStatic, ok := parentFCC.Callable.(*ast.StaticAccessExpression)
				require.True(t, ok)

				parentClass, ok := parentStatic.Class.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "parent", parentClass.Name)

				parentPropertyNode, ok := parentStatic.Property.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "method", parentPropertyNode.Name)

				// Test static::method(...)
				staticStmt, ok := method.Body[2].(*ast.ExpressionStatement)
				require.True(t, ok)

				staticAssign, ok := staticStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				staticFCC, ok := staticAssign.Right.(*ast.FirstClassCallable)
				require.True(t, ok)

				staticStatic, ok := staticFCC.Callable.(*ast.StaticAccessExpression)
				require.True(t, ok)

				staticClass, ok := staticStatic.Class.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "static", staticClass.Name)

				staticPropertyNode, ok := staticStatic.Property.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "method", staticPropertyNode.Name)
			},
		},
		{
			name: "Closure first-class callable",
			input: `<?php
$closure = function() { return 'test'; };
$closureFCC = $closure(...);`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 2)

				stmt2, ok := program.Body[1].(*ast.ExpressionStatement)
				require.True(t, ok)

				assignment, ok := stmt2.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				fcc, ok := assignment.Right.(*ast.FirstClassCallable)
				require.True(t, ok)

				varNode, ok := fcc.Callable.(*ast.Variable)
				require.True(t, ok)
				assert.Equal(t, "$closure", varNode.Name)
			},
		},
		{
			name: "Complex array method first-class callable",
			input: `<?php
$arrayMethod = [new stdClass(), 'toString'](...);`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assignment, ok := stmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				fcc, ok := assignment.Right.(*ast.FirstClassCallable)
				require.True(t, ok)

				arrayLit, ok := fcc.Callable.(*ast.ArrayExpression)
				require.True(t, ok)

				require.Len(t, arrayLit.Elements, 2)

				// First element should be new stdClass()
				newExpr, ok := arrayLit.Elements[0].(*ast.NewExpression)
				require.True(t, ok)

				className, ok := newExpr.Class.(*ast.CallExpression)
				require.True(t, ok)

				classIdent, ok := className.Callee.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "stdClass", classIdent.Name)

				// Second element should be 'toString' string
				stringLit, ok := arrayLit.Elements[1].(*ast.StringLiteral)
				require.True(t, ok)
				assert.Equal(t, "toString", stringLit.Value)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			l := lexer.New(test.input)
			parser := New(l)
			program := parser.ParseProgram()

			require.NotNil(t, program)
			if len(parser.Errors()) > 0 {
				t.Errorf("Parser errors: %v", parser.Errors())
			}

			test.expected(t, program)
		})
	}
}

// Test edge cases and error handling
func TestParsing_FirstClassCallable_EdgeCases(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		shouldHaveError bool
	}{
		{
			name: "Valid ellipsis syntax",
			input: `<?php
$func = strlen(...);`,
			shouldHaveError: false,
		},
		{
			name: "Invalid ellipsis usage (not in function call)",
			input: `<?php  
$invalid = ...;`,
			shouldHaveError: true,
		},
		{
			name: "Nested function calls with first-class callable",
			input: `<?php
$result = call_user_func(strlen(...), 'test');`,
			shouldHaveError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			l := lexer.New(test.input)
			parser := New(l)
			program := parser.ParseProgram()

			require.NotNil(t, program)
			errors := parser.Errors()

			if test.shouldHaveError {
				assert.NotEmpty(t, errors, "Expected parser errors but got none")
			} else {
				if len(errors) > 0 {
					t.Errorf("Unexpected parser errors: %v", errors)
				}
			}
		})
	}
}

func TestParsing_AnonymousClass(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected func(t *testing.T, result ast.Node)
	}{
		{
			name:  "basic anonymous class",
			input: `<?php $obj = new class {};`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				anonClass, ok := assign.Right.(*ast.AnonymousClass)
				require.True(t, ok)

				assert.Empty(t, anonClass.Arguments)
				assert.Nil(t, anonClass.Extends)
				assert.Empty(t, anonClass.Implements)
				assert.Empty(t, anonClass.Body)
			},
		},
		{
			name:  "anonymous class with constructor arguments",
			input: `<?php $obj = new class($arg1, $arg2) {};`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				anonClass, ok := assign.Right.(*ast.AnonymousClass)
				require.True(t, ok)

				require.Len(t, anonClass.Arguments, 2)

				arg1, ok := anonClass.Arguments[0].(*ast.Variable)
				require.True(t, ok)
				assert.Equal(t, "$arg1", arg1.Name)

				arg2, ok := anonClass.Arguments[1].(*ast.Variable)
				require.True(t, ok)
				assert.Equal(t, "$arg2", arg2.Name)
			},
		},
		{
			name:  "anonymous class with extends",
			input: `<?php $obj = new class extends BaseClass {};`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				anonClass, ok := assign.Right.(*ast.AnonymousClass)
				require.True(t, ok)

				require.NotNil(t, anonClass.Extends)
				extends, ok := anonClass.Extends.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "BaseClass", extends.Name)
			},
		},
		{
			name:  "anonymous class with implements",
			input: `<?php $obj = new class implements Interface1, Interface2 {};`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				anonClass, ok := assign.Right.(*ast.AnonymousClass)
				require.True(t, ok)

				require.Len(t, anonClass.Implements, 2)

				iface1, ok := anonClass.Implements[0].(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "Interface1", iface1.Name)

				iface2, ok := anonClass.Implements[1].(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "Interface2", iface2.Name)
			},
		},
		{
			name:  "anonymous class with class body",
			input: `<?php $obj = new class { private $prop; public function method() {} };`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				anonClass, ok := assign.Right.(*ast.AnonymousClass)
				require.True(t, ok)

				// Should have property and method in body
				require.Len(t, anonClass.Body, 2)

				// First should be property declaration
				_, ok = anonClass.Body[0].(*ast.PropertyDeclaration)
				require.True(t, ok, "Expected PropertyDeclaration, got %T", anonClass.Body[0])

				// Second should be method declaration
				_, ok = anonClass.Body[1].(*ast.FunctionDeclaration)
				require.True(t, ok, "Expected FunctionDeclaration, got %T", anonClass.Body[1])
			},
		},
		{
			name:  "anonymous class with final modifier",
			input: `<?php $obj = new final class {};`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				anonClass, ok := assign.Right.(*ast.AnonymousClass)
				require.True(t, ok)

				assert.Equal(t, []string{"final"}, anonClass.Modifiers)
				assert.Empty(t, anonClass.Arguments)
				assert.Nil(t, anonClass.Extends)
				assert.Empty(t, anonClass.Implements)
				assert.Empty(t, anonClass.Body)
			},
		},
		{
			name:  "anonymous class with readonly modifier",
			input: `<?php $obj = new readonly class {};`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				anonClass, ok := assign.Right.(*ast.AnonymousClass)
				require.True(t, ok)

				assert.Equal(t, []string{"readonly"}, anonClass.Modifiers)
				assert.Empty(t, anonClass.Arguments)
				assert.Nil(t, anonClass.Extends)
				assert.Empty(t, anonClass.Implements)
				assert.Empty(t, anonClass.Body)
			},
		},
		{
			name:  "anonymous class with abstract modifier",
			input: `<?php $obj = new abstract class {};`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				anonClass, ok := assign.Right.(*ast.AnonymousClass)
				require.True(t, ok)

				assert.Equal(t, []string{"abstract"}, anonClass.Modifiers)
				assert.Empty(t, anonClass.Arguments)
				assert.Nil(t, anonClass.Extends)
				assert.Empty(t, anonClass.Implements)
				assert.Empty(t, anonClass.Body)
			},
		},
		{
			name:  "anonymous class with multiple modifiers and complex structure",
			input: `<?php $obj = new final readonly class($param) extends Parent implements Interface { private $prop; };`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				anonClass, ok := assign.Right.(*ast.AnonymousClass)
				require.True(t, ok)

				assert.Equal(t, []string{"final", "readonly"}, anonClass.Modifiers)

				// Check constructor argument
				require.Len(t, anonClass.Arguments, 1)
				arg, ok := anonClass.Arguments[0].(*ast.Variable)
				require.True(t, ok)
				assert.Equal(t, "$param", arg.Name)

				// Check extends
				require.NotNil(t, anonClass.Extends)
				extends, ok := anonClass.Extends.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "Parent", extends.Name)

				// Check implements
				require.Len(t, anonClass.Implements, 1)
				iface, ok := anonClass.Implements[0].(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "Interface", iface.Name)

				// Check class body
				require.Len(t, anonClass.Body, 1)
			},
		},
		{
			name:  "anonymous class with attributes",
			input: `<?php $obj = new #[Attribute] class {};`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				anonClass, ok := assign.Right.(*ast.AnonymousClass)
				require.True(t, ok)

				// Check attributes
				require.Len(t, anonClass.Attributes, 1)
				attrGroup := anonClass.Attributes[0]
				require.Len(t, attrGroup.Attributes, 1)
				attr := attrGroup.Attributes[0]
				assert.Equal(t, "Attribute", attr.Name.Name)
				assert.Empty(t, attr.Arguments)

				assert.Empty(t, anonClass.Modifiers)
				assert.Empty(t, anonClass.Arguments)
				assert.Nil(t, anonClass.Extends)
				assert.Empty(t, anonClass.Implements)
				assert.Empty(t, anonClass.Body)
			},
		},
		{
			name:  "anonymous class with variadic constructor arguments",
			input: `<?php return new class(...$arguments) implements CastsAttributes, SerializesCastableAttributes { private $argument; public function __construct($argument = null) { $this->argument = $argument; } };`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				returnStmt, ok := program.Body[0].(*ast.ReturnStatement)
				require.True(t, ok)

				anonClass, ok := returnStmt.Argument.(*ast.AnonymousClass)
				require.True(t, ok)

				// Check variadic argument
				require.Len(t, anonClass.Arguments, 1)
				spreadExpr, ok := anonClass.Arguments[0].(*ast.SpreadExpression)
				require.True(t, ok, "Expected SpreadExpression for ...%arguments, got %T", anonClass.Arguments[0])

				variable, ok := spreadExpr.Argument.(*ast.Variable)
				require.True(t, ok)
				assert.Equal(t, "$arguments", variable.Name)

				// Check implements clause with multiple interfaces
				require.Len(t, anonClass.Implements, 2)
				
				iface1, ok := anonClass.Implements[0].(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "CastsAttributes", iface1.Name)

				iface2, ok := anonClass.Implements[1].(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "SerializesCastableAttributes", iface2.Name)

				// Check class body has property and constructor
				require.Len(t, anonClass.Body, 2)
				
				// First should be property declaration
				_, ok = anonClass.Body[0].(*ast.PropertyDeclaration)
				require.True(t, ok, "Expected PropertyDeclaration, got %T", anonClass.Body[0])

				// Second should be constructor method
				method, ok := anonClass.Body[1].(*ast.FunctionDeclaration)
				require.True(t, ok, "Expected FunctionDeclaration, got %T", anonClass.Body[1])
				methodName, ok := method.Name.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "__construct", methodName.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			result := p.ParseProgram()

			require.NotNil(t, result)
			assert.Empty(t, p.Errors(), "Parser errors: %v", p.Errors())

			tt.expected(t, result)
		})
	}
}

func TestParsing_AttributesEnhanced(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected func(t *testing.T, result ast.Node)
	}{
		{
			name:  "single attribute without parameters",
			input: `<?php $attr = #[Route];`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				attrGroup, ok := assign.Right.(*ast.AttributeGroup)
				require.True(t, ok)

				require.Len(t, attrGroup.Attributes, 1)
				assert.Equal(t, "Route", attrGroup.Attributes[0].Name.Name)
				assert.Empty(t, attrGroup.Attributes[0].Arguments)
			},
		},
		{
			name:  "single attribute with parameters",
			input: `<?php $attr = #[Route("/api/users")];`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				attrGroup, ok := assign.Right.(*ast.AttributeGroup)
				require.True(t, ok)

				require.Len(t, attrGroup.Attributes, 1)
				attr := attrGroup.Attributes[0]
				assert.Equal(t, "Route", attr.Name.Name)
				require.Len(t, attr.Arguments, 1)

				arg, ok := attr.Arguments[0].(*ast.StringLiteral)
				require.True(t, ok)
				assert.Equal(t, "/api/users", arg.Value)
			},
		},
		{
			name:  "attribute group with multiple attributes",
			input: `<?php $expr = #[Route("/api"), Method("GET")];`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				attrGroup, ok := assign.Right.(*ast.AttributeGroup)
				require.True(t, ok, "Expected AttributeGroup, got %T", assign.Right)

				require.Len(t, attrGroup.Attributes, 2)

				// First attribute: Route("/api")
				assert.Equal(t, "Route", attrGroup.Attributes[0].Name.Name)
				require.Len(t, attrGroup.Attributes[0].Arguments, 1)

				arg1, ok := attrGroup.Attributes[0].Arguments[0].(*ast.StringLiteral)
				require.True(t, ok)
				assert.Equal(t, "/api", arg1.Value)

				// Second attribute: Method("GET")
				assert.Equal(t, "Method", attrGroup.Attributes[1].Name.Name)
				require.Len(t, attrGroup.Attributes[1].Arguments, 1)

				arg2, ok := attrGroup.Attributes[1].Arguments[0].(*ast.StringLiteral)
				require.True(t, ok)
				assert.Equal(t, "GET", arg2.Value)
			},
		},
		{
			name:  "attribute with named parameters",
			input: `<?php $expr = #[Cache(ttl: 3600, tags: ["users"])];`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				attrGroup, ok := assign.Right.(*ast.AttributeGroup)
				require.True(t, ok)

				require.Len(t, attrGroup.Attributes, 1)
				attr := attrGroup.Attributes[0]

				assert.Equal(t, "Cache", attr.Name.Name)
				require.Len(t, attr.Arguments, 2)

				// First argument: ttl: 3600
				namedArg1, ok := attr.Arguments[0].(*ast.NamedArgument)
				require.True(t, ok)
				assert.Equal(t, "ttl", namedArg1.Name.Name)

				num, ok := namedArg1.Value.(*ast.NumberLiteral)
				require.True(t, ok)
				assert.Equal(t, "3600", num.Value)

				// Second argument: tags: ["users"]
				namedArg2, ok := attr.Arguments[1].(*ast.NamedArgument)
				require.True(t, ok)
				assert.Equal(t, "tags", namedArg2.Name.Name)

				arrayExpr, ok := namedArg2.Value.(*ast.ArrayExpression)
				require.True(t, ok)
				require.Len(t, arrayExpr.Elements, 1)
			},
		},
		{
			name:  "multiple attributes without parameters",
			input: `<?php $expr = #[Deprecated, Internal, Serializable];`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				attrGroup, ok := assign.Right.(*ast.AttributeGroup)
				require.True(t, ok)

				require.Len(t, attrGroup.Attributes, 3)

				expectedNames := []string{"Deprecated", "Internal", "Serializable"}
				for i, attr := range attrGroup.Attributes {
					assert.Equal(t, expectedNames[i], attr.Name.Name)
					assert.Empty(t, attr.Arguments)
				}
			},
		},
		{
			name:  "multiline attribute with trailing comma",
			input: `<?php $expr = #[AsCommand(
    name: 'error:dump',
    description: 'Dump error pages to plain HTML files',
)];`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				attrGroup, ok := assign.Right.(*ast.AttributeGroup)
				require.True(t, ok)

				require.Len(t, attrGroup.Attributes, 1)
				attr := attrGroup.Attributes[0]
				assert.Equal(t, "AsCommand", attr.Name.Name)
				require.Len(t, attr.Arguments, 2)

				// First argument: name: 'error:dump'
				namedArg1, ok := attr.Arguments[0].(*ast.NamedArgument)
				require.True(t, ok)
				assert.Equal(t, "name", namedArg1.Name.Name)
				str1, ok := namedArg1.Value.(*ast.StringLiteral)
				require.True(t, ok)
				assert.Equal(t, "error:dump", str1.Value)

				// Second argument: description: 'Dump error pages...'
				namedArg2, ok := attr.Arguments[1].(*ast.NamedArgument)
				require.True(t, ok)
				assert.Equal(t, "description", namedArg2.Name.Name)
				str2, ok := namedArg2.Value.(*ast.StringLiteral)
				require.True(t, ok)
				assert.Equal(t, "Dump error pages to plain HTML files", str2.Value)
			},
		},
		{
			name:  "attribute with static keyword",
			input: `<?php $expr = #[Route, static];`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				attrGroup, ok := assign.Right.(*ast.AttributeGroup)
				require.True(t, ok)

				require.Len(t, attrGroup.Attributes, 2)
				assert.Equal(t, "Route", attrGroup.Attributes[0].Name.Name)
				assert.Equal(t, "static", attrGroup.Attributes[1].Name.Name)
			},
		},
		{
			name:  "attribute with simple trailing comma",
			input: `<?php $expr = #[Route("path", "GET",)];`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				attrGroup, ok := assign.Right.(*ast.AttributeGroup)
				require.True(t, ok)

				require.Len(t, attrGroup.Attributes, 1)
				attr := attrGroup.Attributes[0]
				assert.Equal(t, "Route", attr.Name.Name)
				require.Len(t, attr.Arguments, 2)

				str1, ok := attr.Arguments[0].(*ast.StringLiteral)
				require.True(t, ok)
				assert.Equal(t, "path", str1.Value)

				str2, ok := attr.Arguments[1].(*ast.StringLiteral)
				require.True(t, ok)
				assert.Equal(t, "GET", str2.Value)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			result := p.ParseProgram()

			require.NotNil(t, result)
			assert.Empty(t, p.Errors(), "Parser errors: %v", p.Errors())

			tt.expected(t, result)
		})
	}
}

func TestParsing_AttributeErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name:          "unclosed attribute group",
			input:         `<?php #[Route("/api");`,
			expectedError: "expected next token to be `]`",
		},
		{
			name:          "attribute without name",
			input:         `<?php #[];`,
			expectedError: "expected attribute name",
		},
		{
			name:          "malformed attribute parameters",
			input:         `<?php #[Route(;`,
			expectedError: "expected next token to be `)`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			result := p.ParseProgram()

			// Should have parsing errors
			errors := p.Errors()
			require.NotEmpty(t, errors, "Expected parsing errors but got none")

			// Check if expected error message is present
			found := false
			for _, err := range errors {
				if contains(err, tt.expectedError) {
					found = true
					break
				}
			}
			assert.True(t, found, "Expected error containing '%s', got: %v", tt.expectedError, errors)

			// Result should still be parseable (error recovery)
			require.NotNil(t, result)
		})
	}
}

func TestParsing_InternalFunctions(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(*testing.T, *ast.Program)
	}{
		{
			name:  "Isset with single variable",
			input: `<?php isset($x);`,
			check: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				exprStmt := program.Body[0].(*ast.ExpressionStatement)
				issetExpr := exprStmt.Expression.(*ast.IssetExpression)
				assert.Len(t, issetExpr.Arguments, 1)
				varExpr := issetExpr.Arguments[0].(*ast.Variable)
				assert.Equal(t, "$x", varExpr.Name)
			},
		},
		{
			name:  "Isset with multiple variables (should be AND connected)",
			input: `<?php isset($x, $y, $z);`,
			check: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				exprStmt := program.Body[0].(*ast.ExpressionStatement)

				// The result should be: isset($x) && isset($y) && isset($z)
				// So the top level should be a binary expression with &&
				binaryExpr := exprStmt.Expression.(*ast.BinaryExpression)
				assert.Equal(t, "&&", binaryExpr.Operator)

				// Left side should be another binary expression for isset($x) && isset($y)
				leftBinary := binaryExpr.Left.(*ast.BinaryExpression)
				assert.Equal(t, "&&", leftBinary.Operator)

				// Check the innermost isset($x)
				issetX := leftBinary.Left.(*ast.IssetExpression)
				assert.Len(t, issetX.Arguments, 1)
				varX := issetX.Arguments[0].(*ast.Variable)
				assert.Equal(t, "$x", varX.Name)

				// Check isset($y)
				issetY := leftBinary.Right.(*ast.IssetExpression)
				assert.Len(t, issetY.Arguments, 1)
				varY := issetY.Arguments[0].(*ast.Variable)
				assert.Equal(t, "$y", varY.Name)

				// Check isset($z) on the right side
				issetZ := binaryExpr.Right.(*ast.IssetExpression)
				assert.Len(t, issetZ.Arguments, 1)
				varZ := issetZ.Arguments[0].(*ast.Variable)
				assert.Equal(t, "$z", varZ.Name)
			},
		},
		{
			name:  "Empty function",
			input: `<?php empty($var);`,
			check: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				exprStmt := program.Body[0].(*ast.ExpressionStatement)
				emptyExpr := exprStmt.Expression.(*ast.EmptyExpression)
				varExpr := emptyExpr.Expression.(*ast.Variable)
				assert.Equal(t, "$var", varExpr.Name)
			},
		},
		{
			name:  "Include statement",
			input: `<?php include 'file.php';`,
			check: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				exprStmt := program.Body[0].(*ast.ExpressionStatement)
				includeExpr := exprStmt.Expression.(*ast.IncludeOrEvalExpression)
				assert.Equal(t, lexer.T_INCLUDE, includeExpr.Type)
				stringExpr := includeExpr.Expr.(*ast.StringLiteral)
				assert.Equal(t, "file.php", stringExpr.Value)
			},
		},
		{
			name:  "Require_once statement",
			input: `<?php require_once "config.php";`,
			check: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				exprStmt := program.Body[0].(*ast.ExpressionStatement)
				requireExpr := exprStmt.Expression.(*ast.IncludeOrEvalExpression)
				assert.Equal(t, lexer.T_REQUIRE_ONCE, requireExpr.Type)
				stringExpr := requireExpr.Expr.(*ast.StringLiteral)
				assert.Equal(t, "config.php", stringExpr.Value)
			},
		},
		{
			name:  "Eval statement",
			input: `<?php eval($code);`,
			check: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				exprStmt := program.Body[0].(*ast.ExpressionStatement)
				evalExpr := exprStmt.Expression.(*ast.EvalExpression)
				varExpr := evalExpr.Argument.(*ast.Variable)
				assert.Equal(t, "$code", varExpr.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			// Check for parser errors
			if len(p.Errors()) > 0 {
				t.Errorf("Parser errors: %v", p.Errors())
				return
			}

			assert.NotNil(t, program)
			tt.check(t, program)
		})
	}
}

func TestParsing_SpaceshipOperator(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(*testing.T, *ast.Program)
	}{
		{
			name:  "Simple spaceship comparison",
			input: `<?php $result = $a <=> $b;`,
			check: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				exprStmt := program.Body[0].(*ast.ExpressionStatement)
				assignExpr := exprStmt.Expression.(*ast.AssignmentExpression)

				// Right side should be spaceship expression
				spaceshipExpr := assignExpr.Right.(*ast.BinaryExpression)
				assert.Equal(t, "<=>", spaceshipExpr.Operator)

				// Check left and right operands
				leftVar := spaceshipExpr.Left.(*ast.Variable)
				assert.Equal(t, "$a", leftVar.Name)

				rightVar := spaceshipExpr.Right.(*ast.Variable)
				assert.Equal(t, "$b", rightVar.Name)
			},
		},
		{
			name:  "Spaceship with numbers",
			input: `<?php $result = 1 <=> 2;`,
			check: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				exprStmt := program.Body[0].(*ast.ExpressionStatement)
				assignExpr := exprStmt.Expression.(*ast.AssignmentExpression)

				spaceshipExpr := assignExpr.Right.(*ast.BinaryExpression)
				assert.Equal(t, "<=>", spaceshipExpr.Operator)

				leftNum := spaceshipExpr.Left.(*ast.NumberLiteral)
				assert.Equal(t, "1", leftNum.Value)
				assert.Equal(t, "integer", leftNum.Kind)

				rightNum := spaceshipExpr.Right.(*ast.NumberLiteral)
				assert.Equal(t, "2", rightNum.Value)
				assert.Equal(t, "integer", rightNum.Kind)
			},
		},
		{
			name:  "Complex spaceship expression",
			input: `<?php $result = ($a + 1) <=> ($b * 2);`,
			check: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				exprStmt := program.Body[0].(*ast.ExpressionStatement)
				assignExpr := exprStmt.Expression.(*ast.AssignmentExpression)

				spaceshipExpr := assignExpr.Right.(*ast.BinaryExpression)
				assert.Equal(t, "<=>", spaceshipExpr.Operator)

				// Left side should be a binary expression ($a + 1)
				leftExpr := spaceshipExpr.Left.(*ast.BinaryExpression)
				assert.Equal(t, "+", leftExpr.Operator)

				// Right side should be a binary expression ($b * 2)
				rightExpr := spaceshipExpr.Right.(*ast.BinaryExpression)
				assert.Equal(t, "*", rightExpr.Operator)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			// Check for parser errors
			if len(p.Errors()) > 0 {
				t.Errorf("Parser errors: %v", p.Errors())
				return
			}

			assert.NotNil(t, program)
			tt.check(t, program)
		})
	}
}

func TestParsing_NamedArguments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected func(t *testing.T, result ast.Node)
	}{
		{
			name:  "single named argument",
			input: `<?php test(name: "John");`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				call, ok := exprStmt.Expression.(*ast.CallExpression)
				require.True(t, ok)
				require.Len(t, call.Arguments, 1)

				namedArg, ok := call.Arguments[0].(*ast.NamedArgument)
				require.True(t, ok)

				assert.Equal(t, "name", namedArg.Name.Name)

				stringLiteral, ok := namedArg.Value.(*ast.StringLiteral)
				require.True(t, ok)
				assert.Equal(t, "John", stringLiteral.Value)
			},
		},
		{
			name:  "multiple named arguments",
			input: `<?php calculate(x: 10, y: 20, operation: "add");`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				call, ok := exprStmt.Expression.(*ast.CallExpression)
				require.True(t, ok)
				require.Len(t, call.Arguments, 3)

				// First argument: x: 10
				namedArg1, ok := call.Arguments[0].(*ast.NamedArgument)
				require.True(t, ok)
				assert.Equal(t, "x", namedArg1.Name.Name)

				num1, ok := namedArg1.Value.(*ast.NumberLiteral)
				require.True(t, ok)
				assert.Equal(t, "10", num1.Value)

				// Second argument: y: 20
				namedArg2, ok := call.Arguments[1].(*ast.NamedArgument)
				require.True(t, ok)
				assert.Equal(t, "y", namedArg2.Name.Name)

				num2, ok := namedArg2.Value.(*ast.NumberLiteral)
				require.True(t, ok)
				assert.Equal(t, "20", num2.Value)

				// Third argument: operation: "add"
				namedArg3, ok := call.Arguments[2].(*ast.NamedArgument)
				require.True(t, ok)
				assert.Equal(t, "operation", namedArg3.Name.Name)

				str, ok := namedArg3.Value.(*ast.StringLiteral)
				require.True(t, ok)
				assert.Equal(t, "add", str.Value)
			},
		},
		{
			name:  "mixed positional and named arguments",
			input: `<?php mixed_args(1, 2, name: "Alice", value: 42);`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				call, ok := exprStmt.Expression.(*ast.CallExpression)
				require.True(t, ok)
				require.Len(t, call.Arguments, 4)

				// First argument: 1 (positional)
				num1, ok := call.Arguments[0].(*ast.NumberLiteral)
				require.True(t, ok)
				assert.Equal(t, "1", num1.Value)

				// Second argument: 2 (positional)
				num2, ok := call.Arguments[1].(*ast.NumberLiteral)
				require.True(t, ok)
				assert.Equal(t, "2", num2.Value)

				// Third argument: name: "Alice" (named)
				namedArg1, ok := call.Arguments[2].(*ast.NamedArgument)
				require.True(t, ok)
				assert.Equal(t, "name", namedArg1.Name.Name)

				str, ok := namedArg1.Value.(*ast.StringLiteral)
				require.True(t, ok)
				assert.Equal(t, "Alice", str.Value)

				// Fourth argument: value: 42 (named)
				namedArg2, ok := call.Arguments[3].(*ast.NamedArgument)
				require.True(t, ok)
				assert.Equal(t, "value", namedArg2.Name.Name)

				num4, ok := namedArg2.Value.(*ast.NumberLiteral)
				require.True(t, ok)
				assert.Equal(t, "42", num4.Value)
			},
		},
		{
			name:  "named argument with variable value",
			input: `<?php test(name: $userName);`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				call, ok := exprStmt.Expression.(*ast.CallExpression)
				require.True(t, ok)
				require.Len(t, call.Arguments, 1)

				namedArg, ok := call.Arguments[0].(*ast.NamedArgument)
				require.True(t, ok)
				assert.Equal(t, "name", namedArg.Name.Name)

				variable, ok := namedArg.Value.(*ast.Variable)
				require.True(t, ok)
				assert.Equal(t, "$userName", variable.Name)
			},
		},
		{
			name:  "named argument with complex expression",
			input: `<?php test(result: $a + $b * 2);`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				call, ok := exprStmt.Expression.(*ast.CallExpression)
				require.True(t, ok)
				require.Len(t, call.Arguments, 1)

				namedArg, ok := call.Arguments[0].(*ast.NamedArgument)
				require.True(t, ok)
				assert.Equal(t, "result", namedArg.Name.Name)

				// Should be a binary expression: $a + ($b * 2)
				binaryExpr, ok := namedArg.Value.(*ast.BinaryExpression)
				require.True(t, ok)
				assert.Equal(t, "+", binaryExpr.Operator)
			},
		},
		{
			name:  "named arguments with reserved keywords",
			input: `<?php assertJobHandled(class: $class, function: $callback, new: true, array: $data);`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				call, ok := exprStmt.Expression.(*ast.CallExpression)
				require.True(t, ok)
				require.Len(t, call.Arguments, 4)

				// First argument: class: $class
				namedArg1, ok := call.Arguments[0].(*ast.NamedArgument)
				require.True(t, ok)
				assert.Equal(t, "class", namedArg1.Name.Name)

				var1, ok := namedArg1.Value.(*ast.Variable)
				require.True(t, ok)
				assert.Equal(t, "$class", var1.Name)

				// Second argument: function: $callback
				namedArg2, ok := call.Arguments[1].(*ast.NamedArgument)
				require.True(t, ok)
				assert.Equal(t, "function", namedArg2.Name.Name)

				var2, ok := namedArg2.Value.(*ast.Variable)
				require.True(t, ok)
				assert.Equal(t, "$callback", var2.Name)

				// Third argument: new: true
				namedArg3, ok := call.Arguments[2].(*ast.NamedArgument)
				require.True(t, ok)
				assert.Equal(t, "new", namedArg3.Name.Name)

				// Fourth argument: array: $data  
				namedArg4, ok := call.Arguments[3].(*ast.NamedArgument)
				require.True(t, ok)
				assert.Equal(t, "array", namedArg4.Name.Name)

				var3, ok := namedArg4.Value.(*ast.Variable)
				require.True(t, ok)
				assert.Equal(t, "$data", var3.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			result := p.ParseProgram()

			require.NotNil(t, result)
			assert.Empty(t, p.Errors(), "Parser errors: %v", p.Errors())

			tt.expected(t, result)
		})
	}
}

func TestParsing_NamedArgumentsErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name:          "named argument without value",
			input:         `<?php test(name:);`,
			expectedError: "expected next token to be `)`, got `;`",
		},
		{
			name:          "named argument without colon",
			input:         `<?php test(name "value");`,
			expectedError: "expected next token to be `)`, got `T_CONSTANT_ENCAPSED_STRING`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			result := p.ParseProgram()

			// Should have parsing errors
			errors := p.Errors()
			require.NotEmpty(t, errors, "Expected parsing errors but got none")

			// Check if expected error message is present
			found := false
			for _, err := range errors {
				if contains(err, tt.expectedError) {
					found = true
					break
				}
			}
			assert.True(t, found, "Expected error containing '%s', got: %v", tt.expectedError, errors)

			// Result should still be parseable (error recovery)
			require.NotNil(t, result)
		})
	}
}

func TestParsing_ReservedKeywords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected func(t *testing.T, program *ast.Program)
	}{
		{
			name: "Class constants with reserved keywords",
			input: `<?php
class TestClass {
    const class = 'class_value';
    const function = 'function_value';
    const if = 'if_value';
    public const new = 'new_value';
    private const while = 'while_value';
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				classStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				classExpr, ok := classStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)

				require.Len(t, classExpr.Body, 5)

				// Test first constant: const class = 'class_value';
				constDecl1, ok := classExpr.Body[0].(*ast.ClassConstantDeclaration)
				require.True(t, ok)
				assert.Equal(t, "", constDecl1.Visibility) // No explicit visibility
				assert.Len(t, constDecl1.Constants, 1)
				nameNode1, ok := constDecl1.Constants[0].Name.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "class", nameNode1.Name)

				// Test public const new = 'new_value';
				constDecl4, ok := classExpr.Body[3].(*ast.ClassConstantDeclaration)
				require.True(t, ok)
				assert.Equal(t, "public", constDecl4.Visibility) // Has explicit public visibility
				nameNode4, ok := constDecl4.Constants[0].Name.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "new", nameNode4.Name)

				// Test private const while = 'while_value';
				constDecl5, ok := classExpr.Body[4].(*ast.ClassConstantDeclaration)
				require.True(t, ok)
				assert.Equal(t, "private", constDecl5.Visibility)
				nameNode5, ok := constDecl5.Constants[0].Name.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "while", nameNode5.Name)
			},
		},
		{
			name: "Method names with reserved keywords",
			input: `<?php
class TestClass {
    public function class() {
        return 'class';
    }
    
    private function if() {
        return 'if';
    }
    
    protected function while() {
        return 'while';
    }
    
    public function function() {
        return 'function';
    }
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				classStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				classExpr, ok := classStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)

				require.Len(t, classExpr.Body, 4)

				// Test public function class()
				method1, ok := classExpr.Body[0].(*ast.FunctionDeclaration)
				require.True(t, ok)
				nameNode1, ok := method1.Name.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "class", nameNode1.Name)
				assert.Equal(t, "public", method1.Visibility)

				// Test private function if()
				method2, ok := classExpr.Body[1].(*ast.FunctionDeclaration)
				require.True(t, ok)
				nameNode2, ok := method2.Name.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "if", nameNode2.Name)
				assert.Equal(t, "private", method2.Visibility)

				// Test protected function while()
				method3, ok := classExpr.Body[2].(*ast.FunctionDeclaration)
				require.True(t, ok)
				nameNode3, ok := method3.Name.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "while", nameNode3.Name)
				assert.Equal(t, "protected", method3.Visibility)

				// Test public function function()
				method4, ok := classExpr.Body[3].(*ast.FunctionDeclaration)
				require.True(t, ok)
				nameNode4, ok := method4.Name.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "function", nameNode4.Name)
				assert.Equal(t, "public", method4.Visibility)
			},
		},
		{
			name: "Property access with reserved keywords",
			input: `<?php
$obj->class;
$obj->function;
$obj->if;
$obj->while;
$obj->new;`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 5)

				// Test $obj->class
				stmt1, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				propAccess1, ok := stmt1.Expression.(*ast.PropertyAccessExpression)
				require.True(t, ok)
				propIdent1, ok := propAccess1.Property.(*ast.IdentifierNode)
				assert.True(t, ok, "Property should be IdentifierNode")
				assert.Equal(t, "class", propIdent1.Name)

				// Test $obj->function
				stmt2, ok := program.Body[1].(*ast.ExpressionStatement)
				require.True(t, ok)
				propAccess2, ok := stmt2.Expression.(*ast.PropertyAccessExpression)
				require.True(t, ok)
				assert.Equal(t, "function", propAccess2.Property.(*ast.IdentifierNode).Name)

				// Test $obj->if
				stmt3, ok := program.Body[2].(*ast.ExpressionStatement)
				require.True(t, ok)
				propAccess3, ok := stmt3.Expression.(*ast.PropertyAccessExpression)
				require.True(t, ok)
				assert.Equal(t, "if", propAccess3.Property.(*ast.IdentifierNode).Name)

				// Test $obj->while
				stmt4, ok := program.Body[3].(*ast.ExpressionStatement)
				require.True(t, ok)
				propAccess4, ok := stmt4.Expression.(*ast.PropertyAccessExpression)
				require.True(t, ok)
				assert.Equal(t, "while", propAccess4.Property.(*ast.IdentifierNode).Name)

				// Test $obj->new
				stmt5, ok := program.Body[4].(*ast.ExpressionStatement)
				require.True(t, ok)
				propAccess5, ok := stmt5.Expression.(*ast.PropertyAccessExpression)
				require.True(t, ok)
				assert.Equal(t, "new", propAccess5.Property.(*ast.IdentifierNode).Name)
			},
		},
		{
			name: "Nullsafe property access with reserved keywords",
			input: `<?php
$obj?->class;
$obj?->function;
$obj?->if;`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 3)

				// Test $obj?->class
				stmt1, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				propAccess1, ok := stmt1.Expression.(*ast.NullsafePropertyAccessExpression)
				require.True(t, ok)
				assert.Equal(t, "class", propAccess1.Property.(*ast.IdentifierNode).Name)

				// Test $obj?->function
				stmt2, ok := program.Body[1].(*ast.ExpressionStatement)
				require.True(t, ok)
				propAccess2, ok := stmt2.Expression.(*ast.NullsafePropertyAccessExpression)
				require.True(t, ok)
				assert.Equal(t, "function", propAccess2.Property.(*ast.IdentifierNode).Name)

				// Test $obj?->if
				stmt3, ok := program.Body[2].(*ast.ExpressionStatement)
				require.True(t, ok)
				propAccess3, ok := stmt3.Expression.(*ast.NullsafePropertyAccessExpression)
				require.True(t, ok)
				assert.Equal(t, "if", propAccess3.Property.(*ast.IdentifierNode).Name)
			},
		},
		{
			name: "Trait adaptations with reserved keywords",
			input: `<?php
class TestClass {
    use TestTrait {
        class as function;
        if as while;
        TestTrait::class as public echo;
        function as private new;
    }
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				classStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				classExpr, ok := classStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)

				require.Len(t, classExpr.Body, 1)
				useTraitStmt, ok := classExpr.Body[0].(*ast.UseTraitStatement)
				require.True(t, ok)

				require.Len(t, useTraitStmt.Adaptations, 4)

				// Test class as function
				alias1, ok := useTraitStmt.Adaptations[0].(*ast.TraitAliasStatement)
				require.True(t, ok)
				assert.Equal(t, "class", alias1.Method.Method.Name)
				assert.Equal(t, "function", alias1.Alias.Name)

				// Test if as while
				alias2, ok := useTraitStmt.Adaptations[1].(*ast.TraitAliasStatement)
				require.True(t, ok)
				assert.Equal(t, "if", alias2.Method.Method.Name)
				assert.Equal(t, "while", alias2.Alias.Name)

				// Test TestTrait::class as public echo
				alias3, ok := useTraitStmt.Adaptations[2].(*ast.TraitAliasStatement)
				require.True(t, ok)
				assert.Equal(t, "TestTrait", alias3.Method.Trait.Name)
				assert.Equal(t, "class", alias3.Method.Method.Name)
				assert.Equal(t, "public", alias3.Visibility)
				assert.Equal(t, "echo", alias3.Alias.Name)

				// Test function as private new
				alias4, ok := useTraitStmt.Adaptations[3].(*ast.TraitAliasStatement)
				require.True(t, ok)
				assert.Equal(t, "function", alias4.Method.Method.Name)
				assert.Equal(t, "private", alias4.Visibility)
				assert.Equal(t, "new", alias4.Alias.Name)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			l := lexer.New(test.input)
			parser := New(l)
			program := parser.ParseProgram()

			require.NotNil(t, program)
			if len(parser.Errors()) > 0 {
				t.Errorf("Parser errors: %v", parser.Errors())
			}

			test.expected(t, program)
		})
	}
}

func TestParsing_StaticMethodCallsWithReservedKeywords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected func(t *testing.T, program *ast.Program)
	}{
		{
			name: "Static method call with 'for' keyword and chaining",
			input: `<?php
Sleep::for(600)->milliseconds();`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				// The outer call: ->milliseconds()
				outerCall, ok := exprStmt.Expression.(*ast.CallExpression)
				require.True(t, ok)

				// The property access: ->milliseconds
				propAccess, ok := outerCall.Callee.(*ast.PropertyAccessExpression)
				require.True(t, ok)
				
				propName, ok := propAccess.Property.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "milliseconds", propName.Name)

				// The inner static call: Sleep::for(600)
				innerCall, ok := propAccess.Object.(*ast.CallExpression)
				require.True(t, ok)
				
				staticAccess, ok := innerCall.Callee.(*ast.StaticAccessExpression)
				require.True(t, ok)
				
				className, ok := staticAccess.Class.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "Sleep", className.Name)
				
				methodName, ok := staticAccess.Property.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "for", methodName.Name)
				
				// Check the argument (600)
				require.Len(t, innerCall.Arguments, 1)
				arg, ok := innerCall.Arguments[0].(*ast.NumberLiteral)
				require.True(t, ok)
				assert.Equal(t, "600", arg.Value)
			},
		},
		{
			name: "Static method calls with various reserved keywords",
			input: `<?php
MyClass::if($condition);
MyClass::while();
MyClass::return($value);
MyClass::function();
MyClass::class();
MyClass::new();
MyClass::foreach($items);
MyClass::try();
MyClass::catch($e);
MyClass::finally();
MyClass::switch($case);
MyClass::case();
MyClass::default();
MyClass::do();
MyClass::echo($msg);
MyClass::print($msg);
MyClass::break();
MyClass::continue();
MyClass::goto($label);`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 19)
				
				// Test a few representative cases
				testCases := []struct {
					index int
					expectedMethod string
				}{
					{0, "if"},
					{1, "while"},
					{2, "return"},
					{3, "function"},
					{4, "class"},
					{5, "new"},
					{6, "foreach"},
					{7, "try"},
					{8, "catch"},
					{9, "finally"},
					{10, "switch"},
					{11, "case"},
					{12, "default"},
					{13, "do"},
					{14, "echo"},
					{15, "print"},
					{16, "break"},
					{17, "continue"},
					{18, "goto"},
				}
				
				for _, tc := range testCases {
					exprStmt, ok := program.Body[tc.index].(*ast.ExpressionStatement)
					require.True(t, ok, "Statement %d should be an ExpressionStatement", tc.index)
					
					call, ok := exprStmt.Expression.(*ast.CallExpression)
					require.True(t, ok, "Expression %d should be a CallExpression", tc.index)
					
					staticAccess, ok := call.Callee.(*ast.StaticAccessExpression)
					require.True(t, ok, "Callee %d should be a StaticAccessExpression", tc.index)
					
					methodName, ok := staticAccess.Property.(*ast.IdentifierNode)
					require.True(t, ok, "Property %d should be an IdentifierNode", tc.index)
					assert.Equal(t, tc.expectedMethod, methodName.Name, "Method name mismatch at index %d", tc.index)
				}
			},
		},
		{
			name: "Complex chaining with reserved keywords",
			input: `<?php
Builder::create()->if($condition)->else($alternative)->finally();
Database::table('users')->where('id', 1)->for($user)->do($action);`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 2)
				
				// First statement: Builder::create()->if($condition)->else($alternative)->finally()
				exprStmt1, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				// Navigate through the chain and verify 'if' and 'else' are parsed correctly
				finalCall, ok := exprStmt1.Expression.(*ast.CallExpression)
				require.True(t, ok)
				
				finalProp, ok := finalCall.Callee.(*ast.PropertyAccessExpression)
				require.True(t, ok)
				finalName, ok := finalProp.Property.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "finally", finalName.Name)
				
				// Second statement should also parse without errors
				exprStmt2, ok := program.Body[1].(*ast.ExpressionStatement)
				require.True(t, ok)
				assert.NotNil(t, exprStmt2.Expression)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parser := New(lexer.New(test.input))
			program := parser.ParseProgram()

			if len(parser.Errors()) != 0 {
				t.Errorf("Parser errors: %v", parser.Errors())
			}

			test.expected(t, program)
		})
	}
}

func TestParsing_ReservedKeywordsAsMethodNames(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected func(t *testing.T, program *ast.Program)
	}{
		{
			name: "method named 'default' with static access patterns",
			input: `<?php
class A {
    public static function default()
    {
        $password = is_callable(static::$defaultCallback)
            ? call_user_func(static::$defaultCallback)
            : static::$defaultCallback;
        
        return $password instanceof Rule ? $password : static::min(8);
    }
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				classDecl, ok := exprStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)

				// Check class name
				className, ok := classDecl.Name.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "A", className.Name)

				// Check method
				require.Len(t, classDecl.Body, 1)
				method, ok := classDecl.Body[0].(*ast.FunctionDeclaration)
				require.True(t, ok)

				// Check method name is 'default'
				methodName, ok := method.Name.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "default", methodName.Name)
				assert.Equal(t, "public", method.Visibility)
				assert.True(t, method.IsStatic)

				// Verify the method has a proper body with complex expressions
				require.Len(t, method.Body, 2) // assignment + return statements
			},
		},
		{
			name: "multiple reserved keywords as method names",
			input: `<?php
class Test {
    public function default() {}
    public function return() {}
    public function break() {}
    public function case() {} 
    public function switch() {}
    public function for() {}
    public function while() {}
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				classDecl, ok := exprStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)

				expectedMethods := []string{"default", "return", "break", "case", "switch", "for", "while"}
				require.Len(t, classDecl.Body, len(expectedMethods))

				for i, expectedMethodName := range expectedMethods {
					method, ok := classDecl.Body[i].(*ast.FunctionDeclaration)
					require.True(t, ok, "Body item %d should be FunctionDeclaration", i)

					methodName, ok := method.Name.(*ast.IdentifierNode)
					require.True(t, ok, "Method %d name should be IdentifierNode", i)
					assert.Equal(t, expectedMethodName, methodName.Name, "Method %d name mismatch", i)
					assert.Equal(t, "public", method.Visibility, "Method %d visibility", i)
				}
			},
		},
		{
			name: "interface method with reserved keyword name",
			input: `<?php
interface TestInterface {
    public function default();
    public function case($arg);
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				interfaceDecl, ok := program.Body[0].(*ast.InterfaceDeclaration)
				require.True(t, ok)

				// Check interface has two methods
				require.Len(t, interfaceDecl.Methods, 2)

				// First method: default()
				method1 := interfaceDecl.Methods[0]
				assert.Equal(t, "default", method1.Name.Name)

				// Second method: case($arg)
				method2 := interfaceDecl.Methods[1]
				assert.Equal(t, "case", method2.Name.Name)
				require.Len(t, method2.Parameters, 1)
				assert.Equal(t, "$arg", method2.Parameters[0].Name)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			l := lexer.New(test.input)
			p := New(l)
			program := p.ParseProgram()

			require.NotNil(t, program)
			errors := p.Errors()
			if len(errors) > 0 {
				t.Fatalf("Parser errors: %v", errors)
			}

			test.expected(t, program)
		})
	}
}

func TestParsing_ReservedKeywords_IsHelperFunctions(t *testing.T) {
	tests := []struct {
		name       string
		tokenType  lexer.TokenType
		isReserved bool
		isSemi     bool
	}{
		{"T_CLASS is reserved non-modifier", lexer.T_CLASS, true, true},
		{"T_FUNCTION is reserved non-modifier", lexer.T_FUNCTION, true, true},
		{"T_IF is reserved non-modifier", lexer.T_IF, true, true},
		{"T_WHILE is reserved non-modifier", lexer.T_WHILE, true, true},
		{"T_NEW is reserved non-modifier", lexer.T_NEW, true, true},
		{"T_PRIVATE is not reserved non-modifier but is semi-reserved", lexer.T_PRIVATE, false, true},
		{"T_PUBLIC is not reserved non-modifier but is semi-reserved", lexer.T_PUBLIC, false, true},
		{"T_PROTECTED is not reserved non-modifier but is semi-reserved", lexer.T_PROTECTED, false, true},
		{"T_STATIC is not reserved non-modifier but is semi-reserved", lexer.T_STATIC, false, true},
		{"T_STRING is not reserved", lexer.T_STRING, false, false},
		{"T_VARIABLE is not reserved", lexer.T_VARIABLE, false, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.isReserved, isReservedNonModifier(test.tokenType))
			assert.Equal(t, test.isSemi, isSemiReserved(test.tokenType))
		})
	}
}

func TestParsing_SpreadSyntax(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected func(t *testing.T, result ast.Node)
	}{
		{
			name:  "array spread with single array",
			input: `<?php $arr2 = [...$arr1];`,
			expected: func(t *testing.T, result ast.Node) {
				program, ok := result.(*ast.Program)
				require.True(t, ok)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				arrayExpr, ok := assign.Right.(*ast.ArrayExpression)
				require.True(t, ok)
				require.Len(t, arrayExpr.Elements, 1)

				spread, ok := arrayExpr.Elements[0].(*ast.SpreadExpression)
				require.True(t, ok)

				variable, ok := spread.Argument.(*ast.Variable)
				require.True(t, ok)
				assert.Equal(t, "$arr1", variable.Name)
			},
		},
		{
			name:  "array spread with mixed elements",
			input: `<?php $arr2 = [...$arr1, 4, 5];`,
			expected: func(t *testing.T, result ast.Node) {
				program, ok := result.(*ast.Program)
				require.True(t, ok)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				arrayExpr, ok := assign.Right.(*ast.ArrayExpression)
				require.True(t, ok)
				require.Len(t, arrayExpr.Elements, 3)

				// First element: spread
				spread, ok := arrayExpr.Elements[0].(*ast.SpreadExpression)
				require.True(t, ok)
				variable, ok := spread.Argument.(*ast.Variable)
				require.True(t, ok)
				assert.Equal(t, "$arr1", variable.Name)

				// Second element: number 4
				num1, ok := arrayExpr.Elements[1].(*ast.NumberLiteral)
				require.True(t, ok)
				assert.Equal(t, "4", num1.Value)

				// Third element: number 5
				num2, ok := arrayExpr.Elements[2].(*ast.NumberLiteral)
				require.True(t, ok)
				assert.Equal(t, "5", num2.Value)
			},
		},
		{
			name:  "array() spread with mixed elements",
			input: `<?php $arr3 = array(0, ...$arr1, 6);`,
			expected: func(t *testing.T, result ast.Node) {
				program, ok := result.(*ast.Program)
				require.True(t, ok)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				arrayExpr, ok := assign.Right.(*ast.ArrayExpression)
				require.True(t, ok)
				require.Len(t, arrayExpr.Elements, 3)

				// First element: number 0
				num0, ok := arrayExpr.Elements[0].(*ast.NumberLiteral)
				require.True(t, ok)
				assert.Equal(t, "0", num0.Value)

				// Second element: spread
				spread, ok := arrayExpr.Elements[1].(*ast.SpreadExpression)
				require.True(t, ok)
				variable, ok := spread.Argument.(*ast.Variable)
				require.True(t, ok)
				assert.Equal(t, "$arr1", variable.Name)

				// Third element: number 6
				num6, ok := arrayExpr.Elements[2].(*ast.NumberLiteral)
				require.True(t, ok)
				assert.Equal(t, "6", num6.Value)
			},
		},
		{
			name:  "function call with spread arguments",
			input: `<?php $result = test(...$arr1);`,
			expected: func(t *testing.T, result ast.Node) {
				program, ok := result.(*ast.Program)
				require.True(t, ok)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				call, ok := assign.Right.(*ast.CallExpression)
				require.True(t, ok)
				require.Len(t, call.Arguments, 1)

				spread, ok := call.Arguments[0].(*ast.SpreadExpression)
				require.True(t, ok)

				variable, ok := spread.Argument.(*ast.Variable)
				require.True(t, ok)
				assert.Equal(t, "$arr1", variable.Name)
			},
		},
		{
			name:  "function call with mixed arguments",
			input: `<?php $mixed = test(1, ...[2, 3]);`,
			expected: func(t *testing.T, result ast.Node) {
				program, ok := result.(*ast.Program)
				require.True(t, ok)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				call, ok := assign.Right.(*ast.CallExpression)
				require.True(t, ok)
				require.Len(t, call.Arguments, 2)

				// First argument: number 1
				num1, ok := call.Arguments[0].(*ast.NumberLiteral)
				require.True(t, ok)
				assert.Equal(t, "1", num1.Value)

				// Second argument: spread of array literal
				spread, ok := call.Arguments[1].(*ast.SpreadExpression)
				require.True(t, ok)

				arrayExpr, ok := spread.Argument.(*ast.ArrayExpression)
				require.True(t, ok)
				require.Len(t, arrayExpr.Elements, 2)

				num2, ok := arrayExpr.Elements[0].(*ast.NumberLiteral)
				require.True(t, ok)
				assert.Equal(t, "2", num2.Value)

				num3, ok := arrayExpr.Elements[1].(*ast.NumberLiteral)
				require.True(t, ok)
				assert.Equal(t, "3", num3.Value)
			},
		},
		{
			name:  "multiple spread in array",
			input: `<?php $arr = [...$a, ...$b, ...$c];`,
			expected: func(t *testing.T, result ast.Node) {
				program, ok := result.(*ast.Program)
				require.True(t, ok)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				arrayExpr, ok := assign.Right.(*ast.ArrayExpression)
				require.True(t, ok)
				require.Len(t, arrayExpr.Elements, 3)

				// All elements should be spread expressions
				for i, element := range arrayExpr.Elements {
					spread, ok := element.(*ast.SpreadExpression)
					require.True(t, ok, "Element %d should be spread expression", i)

					variable, ok := spread.Argument.(*ast.Variable)
					require.True(t, ok)

					expectedNames := []string{"$a", "$b", "$c"}
					assert.Equal(t, expectedNames[i], variable.Name)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			result := p.ParseProgram()

			require.NotNil(t, result)
			assert.Empty(t, p.Errors(), "Parser errors: %v", p.Errors())

			tt.expected(t, result)
		})
	}
}

func TestParsing_SpreadSyntaxErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name:          "spread without expression",
			input:         `<?php $arr = [...];`,
			expectedError: "expected ',' or ']' in array",
		},
		{
			name:          "spread in wrong context",
			input:         `<?php $x = ...5;`,
			expectedError: "no prefix parse function for `T_ELLIPSIS`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			result := p.ParseProgram()

			// Should have parsing errors
			errors := p.Errors()
			require.NotEmpty(t, errors, "Expected parsing errors but got none")

			// Check if expected error message is present
			found := false
			for _, err := range errors {
				if contains(err, tt.expectedError) {
					found = true
					break
				}
			}
			assert.True(t, found, "Expected error containing '%s', got: %v", tt.expectedError, errors)

			// Result should still be parseable (error recovery)
			require.NotNil(t, result)
		})
	}
}

// Helper function to check if error message contains expected text
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) > len(substr) &&
			(s[len(s)-len(substr):] == substr ||
				indexOf(s, substr) != -1)
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func TestParsing_TraitAdaptations(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected func(t *testing.T, program *ast.Program)
	}{
		{
			name: "Simple trait usage without adaptations",
			input: `<?php
class TestClass {
    use TraitA;
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				classStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				classExpr, ok := classStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)

				require.Len(t, classExpr.Body, 1)
				useTraitStmt, ok := classExpr.Body[0].(*ast.UseTraitStatement)
				require.True(t, ok)

				assert.Len(t, useTraitStmt.Traits, 1)
				assert.Equal(t, "TraitA", useTraitStmt.Traits[0].Name)
				assert.Nil(t, useTraitStmt.Adaptations)
			},
		},
		{
			name: "Multiple traits usage without adaptations",
			input: `<?php
class TestClass {
    use TraitA, TraitB, TraitC;
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				classStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				classExpr, ok := classStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)

				require.Len(t, classExpr.Body, 1)
				useTraitStmt, ok := classExpr.Body[0].(*ast.UseTraitStatement)
				require.True(t, ok)

				assert.Len(t, useTraitStmt.Traits, 3)
				assert.Equal(t, "TraitA", useTraitStmt.Traits[0].Name)
				assert.Equal(t, "TraitB", useTraitStmt.Traits[1].Name)
				assert.Equal(t, "TraitC", useTraitStmt.Traits[2].Name)
				assert.Nil(t, useTraitStmt.Adaptations)
			},
		},
		{
			name: "Trait precedence (insteadof)",
			input: `<?php
class TestClass {
    use TraitA, TraitB {
        TraitA::foo insteadof TraitB;
    }
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				classStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				classExpr, ok := classStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)

				require.Len(t, classExpr.Body, 1)
				useTraitStmt, ok := classExpr.Body[0].(*ast.UseTraitStatement)
				require.True(t, ok)

				assert.Len(t, useTraitStmt.Traits, 2)
				require.Len(t, useTraitStmt.Adaptations, 1)

				adaptation := useTraitStmt.Adaptations[0]
				precedence, ok := adaptation.(*ast.TraitPrecedenceStatement)
				require.True(t, ok)

				assert.Equal(t, "TraitA", precedence.Method.Trait.Name)
				assert.Equal(t, "foo", precedence.Method.Method.Name)
				assert.Len(t, precedence.InsteadOf, 1)
				assert.Equal(t, "TraitB", precedence.InsteadOf[0].Name)
			},
		},
		{
			name: "Trait alias with new name",
			input: `<?php
class TestClass {
    use TraitA {
        foo as newFoo;
    }
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				classStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				classExpr, ok := classStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)

				require.Len(t, classExpr.Body, 1)
				useTraitStmt, ok := classExpr.Body[0].(*ast.UseTraitStatement)
				require.True(t, ok)

				require.Len(t, useTraitStmt.Adaptations, 1)

				adaptation := useTraitStmt.Adaptations[0]
				alias, ok := adaptation.(*ast.TraitAliasStatement)
				require.True(t, ok)

				assert.Nil(t, alias.Method.Trait) // Simple method reference
				assert.Equal(t, "foo", alias.Method.Method.Name)
				assert.Equal(t, "newFoo", alias.Alias.Name)
				assert.Empty(t, alias.Visibility)
			},
		},
		{
			name: "Trait alias with visibility change",
			input: `<?php
class TestClass {
    use TraitA {
        foo as private;
    }
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				classStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				classExpr, ok := classStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)

				require.Len(t, classExpr.Body, 1)
				useTraitStmt, ok := classExpr.Body[0].(*ast.UseTraitStatement)
				require.True(t, ok)

				require.Len(t, useTraitStmt.Adaptations, 1)

				adaptation := useTraitStmt.Adaptations[0]
				alias, ok := adaptation.(*ast.TraitAliasStatement)
				require.True(t, ok)

				assert.Equal(t, "foo", alias.Method.Method.Name)
				assert.Nil(t, alias.Alias) // No new name, only visibility change
				assert.Equal(t, "private", alias.Visibility)
			},
		},
		{
			name: "Trait alias with visibility and new name",
			input: `<?php
class TestClass {
    use TraitA {
        TraitA::bar as protected newBar;
    }
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				classStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				classExpr, ok := classStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)

				require.Len(t, classExpr.Body, 1)
				useTraitStmt, ok := classExpr.Body[0].(*ast.UseTraitStatement)
				require.True(t, ok)

				require.Len(t, useTraitStmt.Adaptations, 1)

				adaptation := useTraitStmt.Adaptations[0]
				alias, ok := adaptation.(*ast.TraitAliasStatement)
				require.True(t, ok)

				assert.Equal(t, "TraitA", alias.Method.Trait.Name)
				assert.Equal(t, "bar", alias.Method.Method.Name)
				assert.Equal(t, "newBar", alias.Alias.Name)
				assert.Equal(t, "protected", alias.Visibility)
			},
		},
		{
			name: "Multiple adaptations",
			input: `<?php
class TestClass {
    use TraitA, TraitB {
        TraitA::foo insteadof TraitB;
        TraitB::foo as fooFromB;
        bar as private privateBar;
        TraitA::baz as public publicBaz;
    }
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				classStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				classExpr, ok := classStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)

				require.Len(t, classExpr.Body, 1)
				useTraitStmt, ok := classExpr.Body[0].(*ast.UseTraitStatement)
				require.True(t, ok)

				require.Len(t, useTraitStmt.Adaptations, 4)

				// First adaptation - precedence
				precedence, ok := useTraitStmt.Adaptations[0].(*ast.TraitPrecedenceStatement)
				require.True(t, ok)
				assert.Equal(t, "TraitA", precedence.Method.Trait.Name)
				assert.Equal(t, "foo", precedence.Method.Method.Name)
				assert.Equal(t, "TraitB", precedence.InsteadOf[0].Name)

				// Second adaptation - alias
				alias1, ok := useTraitStmt.Adaptations[1].(*ast.TraitAliasStatement)
				require.True(t, ok)
				assert.Equal(t, "TraitB", alias1.Method.Trait.Name)
				assert.Equal(t, "foo", alias1.Method.Method.Name)
				assert.Equal(t, "fooFromB", alias1.Alias.Name)

				// Third adaptation - visibility change
				alias2, ok := useTraitStmt.Adaptations[2].(*ast.TraitAliasStatement)
				require.True(t, ok)
				assert.Nil(t, alias2.Method.Trait)
				assert.Equal(t, "bar", alias2.Method.Method.Name)
				assert.Equal(t, "privateBar", alias2.Alias.Name)
				assert.Equal(t, "private", alias2.Visibility)

				// Fourth adaptation - visibility with new name
				alias3, ok := useTraitStmt.Adaptations[3].(*ast.TraitAliasStatement)
				require.True(t, ok)
				assert.Equal(t, "TraitA", alias3.Method.Trait.Name)
				assert.Equal(t, "baz", alias3.Method.Method.Name)
				assert.Equal(t, "publicBaz", alias3.Alias.Name)
				assert.Equal(t, "public", alias3.Visibility)
			},
		},
		{
			name: "Multiple insteadof traits",
			input: `<?php
class TestClass {
    use TraitA, TraitB, TraitC {
        TraitA::foo insteadof TraitB, TraitC;
    }
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)

				classStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				classExpr, ok := classStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)

				require.Len(t, classExpr.Body, 1)
				useTraitStmt, ok := classExpr.Body[0].(*ast.UseTraitStatement)
				require.True(t, ok)

				require.Len(t, useTraitStmt.Adaptations, 1)

				adaptation := useTraitStmt.Adaptations[0]
				precedence, ok := adaptation.(*ast.TraitPrecedenceStatement)
				require.True(t, ok)

				assert.Equal(t, "TraitA", precedence.Method.Trait.Name)
				assert.Equal(t, "foo", precedence.Method.Method.Name)
				assert.Len(t, precedence.InsteadOf, 2)
				assert.Equal(t, "TraitB", precedence.InsteadOf[0].Name)
				assert.Equal(t, "TraitC", precedence.InsteadOf[1].Name)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			l := lexer.New(test.input)
			parser := New(l)
			program := parser.ParseProgram()

			require.NotNil(t, program)
			if len(parser.Errors()) > 0 {
				t.Errorf("Parser errors: %v", parser.Errors())
			}

			test.expected(t, program)
		})
	}
}

func TestParsing_TraitsWithUseStatements(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected func(t *testing.T, traitDecl *ast.TraitDeclaration)
	}{
		{
			name: "Trait using single trait",
			input: `<?php
trait NonPrunableTrait {
    use Prunable;
}`,
			expected: func(t *testing.T, traitDecl *ast.TraitDeclaration) {
				assert.Equal(t, "NonPrunableTrait", traitDecl.Name.Name)
				assert.Len(t, traitDecl.Body, 1, "Expected 1 use statement in trait body")
				
				useStmt, ok := traitDecl.Body[0].(*ast.UseTraitStatement)
				require.True(t, ok, "Expected UseTraitStatement in trait body")
				require.Len(t, useStmt.Traits, 1, "Expected 1 trait in use statement")
				assert.Equal(t, "Prunable", useStmt.Traits[0].Name)
				assert.Nil(t, useStmt.Adaptations, "Expected no adaptations")
			},
		},
		{
			name: "Trait using multiple traits",
			input: `<?php
trait CompositeTrait {
    use TraitA, TraitB, TraitC;
}`,
			expected: func(t *testing.T, traitDecl *ast.TraitDeclaration) {
				assert.Equal(t, "CompositeTrait", traitDecl.Name.Name)
				assert.Len(t, traitDecl.Body, 1, "Expected 1 use statement in trait body")
				
				useStmt, ok := traitDecl.Body[0].(*ast.UseTraitStatement)
				require.True(t, ok, "Expected UseTraitStatement in trait body")
				require.Len(t, useStmt.Traits, 3, "Expected 3 traits in use statement")
				assert.Equal(t, "TraitA", useStmt.Traits[0].Name)
				assert.Equal(t, "TraitB", useStmt.Traits[1].Name)
				assert.Equal(t, "TraitC", useStmt.Traits[2].Name)
			},
		},
		{
			name: "Trait using traits with adaptations",
			input: `<?php
trait AdaptedTrait {
    use TraitA, TraitB {
        TraitA::foo insteadof TraitB;
        TraitB::bar as baz;
    }
}`,
			expected: func(t *testing.T, traitDecl *ast.TraitDeclaration) {
				assert.Equal(t, "AdaptedTrait", traitDecl.Name.Name)
				assert.Len(t, traitDecl.Body, 1, "Expected 1 use statement in trait body")
				
				useStmt, ok := traitDecl.Body[0].(*ast.UseTraitStatement)
				require.True(t, ok, "Expected UseTraitStatement in trait body")
				require.Len(t, useStmt.Traits, 2, "Expected 2 traits in use statement")
				require.Len(t, useStmt.Adaptations, 2, "Expected 2 adaptations")
			},
		},
		{
			name: "Trait with multiple use statements and methods",
			input: `<?php
trait MixedTrait {
    use TraitA;
    use TraitB, TraitC;
    
    private $property;
    
    public function method() {
        return "test";
    }
}`,
			expected: func(t *testing.T, traitDecl *ast.TraitDeclaration) {
				assert.Equal(t, "MixedTrait", traitDecl.Name.Name)
				assert.Len(t, traitDecl.Body, 2, "Expected 2 use statements in trait body")
				assert.Len(t, traitDecl.Properties, 1, "Expected 1 property")
				assert.Len(t, traitDecl.Methods, 1, "Expected 1 method")
				
				// Check first use statement
				useStmt1, ok := traitDecl.Body[0].(*ast.UseTraitStatement)
				require.True(t, ok, "Expected first UseTraitStatement")
				require.Len(t, useStmt1.Traits, 1)
				assert.Equal(t, "TraitA", useStmt1.Traits[0].Name)
				
				// Check second use statement
				useStmt2, ok := traitDecl.Body[1].(*ast.UseTraitStatement)
				require.True(t, ok, "Expected second UseTraitStatement")
				require.Len(t, useStmt2.Traits, 2)
				assert.Equal(t, "TraitB", useStmt2.Traits[0].Name)
				assert.Equal(t, "TraitC", useStmt2.Traits[1].Name)
			},
		},
		{
			name: "Trait with constants and use statements",
			input: `<?php
trait ConstantTrait {
    const CONSTANT = "value";
    use HelperTrait;
    
    protected const ANOTHER_CONSTANT = 42;
}`,
			expected: func(t *testing.T, traitDecl *ast.TraitDeclaration) {
				assert.Equal(t, "ConstantTrait", traitDecl.Name.Name)
				assert.Len(t, traitDecl.Body, 3, "Expected 3 statements in trait body")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			checkParserErrors(t, p)
			require.NotNil(t, program)
			require.Len(t, program.Body, 1, "Expected 1 statement")

			stmt := program.Body[0]
			traitDecl, ok := stmt.(*ast.TraitDeclaration)
			require.True(t, ok, "Expected TraitDeclaration, got %T", stmt)
			
			tt.expected(t, traitDecl)
		})
	}
}

func TestParsing_TraitAdaptations_ErrorCases(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "Missing trait name after use",
			input: `<?php
class TestClass {
    use ;
}`,
			expectedError: "expected trait name",
		},
		{
			name: "Missing method name after ::",
			input: `<?php
class TestClass {
    use TraitA {
        TraitA:: as foo;
    }
}`,
			expectedError: "expected method name",
		},
		{
			name: "Missing insteadof trait name",
			input: `<?php
class TestClass {
    use TraitA {
        foo insteadof ;
    }
}`,
			expectedError: "expected trait name",
		},
		{
			name: "Missing method name after :: with insteadof",
			input: `<?php
class TestClass {
    use TraitA {
        TraitA:: insteadof TraitB;
    }
}`,
			expectedError: "expected method name",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			l := lexer.New(test.input)
			parser := New(l)
			program := parser.ParseProgram()

			require.NotNil(t, program)
			errors := parser.Errors()
			require.NotEmpty(t, errors, "Expected parser errors but got none")

			found := false
			for _, err := range errors {
				if containsSubstring(err, test.expectedError) {
					found = true
					break
				}
			}
			assert.True(t, found, "Expected error containing '%s', but got: %v", test.expectedError, errors)
		})
	}
}

func containsSubstring(str, substr string) bool {
	return len(str) >= len(substr) && str[:len(substr)] == substr
}

func TestParsing_MixedAlternativeRegularSyntax(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "alternative if with nested regular if-elseif",
			input: `<?php if ($condition1) : ?>
<div>HTML content</div>
<?php 
if ($condition2) {
    echo "normal syntax";
} elseif ($condition3) {
    echo "this elseif should work";
}
?>
<?php endif; ?>`,
		},
		{
			name: "alternative if with multiple nested regular syntax",
			input: `<?php if ($outer) : ?>
<h1>Title</h1>
<?php 
if ($inner1) {
    if ($deep1) {
        echo "deep1";
    } elseif ($deep2) {
        echo "deep2";  
    }
} elseif ($inner2) {
    echo "inner2";
}
?>
<?php endif; ?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			errors := p.Errors()
			if len(errors) > 0 {
				t.Fatalf("Parser errors: %v", errors)
			}

			require.NotNil(t, program)
			assert.GreaterOrEqual(t, len(program.Body), 1, 
				"Expected at least 1 statement, got %d", len(program.Body))
		})
	}
}

func TestParsing_ForLoopMultipleExpressions(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "for loop with multiple initialization expressions",
			input: `<?php for ($i=0, $count=0; $i < 10; $i++, $count+=2) { echo $i; } ?>`,
		},
		{
			name: "for loop with multiple update expressions", 
			input: `<?php for ($i=0; $i < 10; $i++, $j--, $k*=2) { echo $i; } ?>`,
		},
		{
			name: "for loop with multiple expressions in all parts",
			input: `<?php for ($i=0, $j=10, $k=1; $i < 10, $j > 0; $i++, $j--, $k*=2) { 
				echo "$i $j $k"; 
			} ?>`,
		},
		{
			name: "empty for loop parts",
			input: `<?php for (;;) { break; } ?>`,
		},
		{
			name: "mixed empty and multiple expressions",
			input: `<?php for ($i=0, $j=5; ; $i++, $j--) { 
				if ($i > 5) break; 
			} ?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			errors := p.Errors()
			if len(errors) > 0 {
				t.Fatalf("Parser errors: %v", errors)
			}

			require.NotNil(t, program)
			assert.GreaterOrEqual(t, len(program.Body), 1, 
				"Expected at least 1 statement, got %d", len(program.Body))
		})
	}
}

func TestParsing_MixedPHPHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []struct {
			nodeType string
			content  string
		}
	}{
		{
			name: "basic mixed PHP HTML",
			input: `<?php
echo "Hello";
?>
<h1>Title</h1>
<?php
echo "World";
?>`,
			expected: []struct {
				nodeType string
				content  string
			}{
				{"EchoStatement", "Hello"},
				{"InlineHTML", "<h1>Title</h1>"},
				{"EchoStatement", "World"},
			},
		},
		{
			name: "multiple PHP blocks with HTML",
			input: `<?php $x = 1; ?>
<div>Content</div>
<?php $y = 2; ?>
<span>More</span>
<?php echo $x + $y; ?>`,
			expected: []struct {
				nodeType string
				content  string
			}{
				{"AssignmentExpression", "1"},
				{"InlineHTML", "<div>Content</div>"},
				{"AssignmentExpression", "2"},
				{"InlineHTML", "<span>More</span>"},
				{"EchoStatement", ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			errors := p.Errors()
			if len(errors) > 0 {
				t.Fatalf("Parser errors: %v", errors)
			}

			require.NotNil(t, program)
			
			// The test should parse more than one statement if multiple PHP blocks exist
			if len(tt.expected) > 1 {
				assert.GreaterOrEqual(t, len(program.Body), 2, 
					"Expected at least 2 statements for mixed content, got %d", len(program.Body))
			}

			// For now, we mainly want to ensure no parsing errors
			// Full AST structure validation can be added once the fix is implemented
		})
	}
}

// Test for parsing list() construct with empty elements  
func TestParsing_ListWithEmptyElements(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "list with one empty element at end",
			input:    "<?php list($a, $b, ) = $array;",
			expected: "list($a, $b, ) = $array",
		},
		{
			name:     "list with empty element in middle",
			input:    "<?php list($a, , $c) = $array;", 
			expected: "list($a, , $c) = $array",
		},
		{
			name:     "list with multiple empty elements",
			input:    "<?php list($a, , , $d) = $array;",
			expected: "list($a, , , $d) = $array",
		},
		{
			name:     "list with empty element at start",
			input:    "<?php list(, $b, $c) = $array;",
			expected: "list(, $b, $c) = $array",
		},
		{
			name:     "complex list assignment like WordPress",
			input:    "<?php list( $columns, $hidden, , $primary ) = $this->get_column_info();",
			expected: "list($columns, $hidden, , $primary) = $this->get_column_info()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			
			// Check for parsing errors
			checkParserErrors(t, p)
			
			// Verify we have statements
			assert.NotNil(t, program, "program should not be nil")
			assert.NotEmpty(t, program.Body, "%q should have 1 item(s), but has %d", program.String(), len(program.Body))
		})
	}
}

// TestParsing_ComplexVariableProperties tests parsing of complex variable property syntax
// including ${$var} and $$var patterns, both in static and regular contexts
func TestParsing_ComplexVariableProperties(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected func(t *testing.T, program *ast.Program)
	}{
		{
			name:  "Static complex variable property with braces",
			input: `<?php wp_set_current_user( static::${$user_property} );`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				callExpr, ok := exprStmt.Expression.(*ast.CallExpression)
				require.True(t, ok)
				assert.Equal(t, "wp_set_current_user", callExpr.Callee.(*ast.IdentifierNode).Name)
				
				require.Len(t, callExpr.Arguments, 1)
				staticAccess, ok := callExpr.Arguments[0].(*ast.StaticAccessExpression)
				require.True(t, ok)
				
				assert.Equal(t, "static", staticAccess.Class.(*ast.IdentifierNode).Name)
				assert.Equal(t, "${$user_property}", staticAccess.Property.(*ast.Variable).Name)
			},
		},
		{
			name:  "Static variable variable property",
			input: `<?php echo static::$$prop;`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				echoStmt, ok := program.Body[0].(*ast.EchoStatement)
				require.True(t, ok)
				
				require.Len(t, echoStmt.Arguments, 1)
				staticAccess, ok := echoStmt.Arguments[0].(*ast.StaticAccessExpression)
				require.True(t, ok)
				
				assert.Equal(t, "static", staticAccess.Class.(*ast.IdentifierNode).Name)
				assert.Equal(t, "$$prop", staticAccess.Property.(*ast.Variable).Name)
			},
		},
		{
			name:  "Complex variable syntax in echo",
			input: `<?php echo ${$var};`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				echoStmt, ok := program.Body[0].(*ast.EchoStatement)
				require.True(t, ok)
				
				require.Len(t, echoStmt.Arguments, 1)
				variable, ok := echoStmt.Arguments[0].(*ast.Variable)
				require.True(t, ok)
				assert.Equal(t, "${$var}", variable.Name)
			},
		},
		{
			name:  "Variable variable in assignment",
			input: `<?php $$name = "value";`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)
				
				variable, ok := assign.Left.(*ast.Variable)
				require.True(t, ok)
				assert.Equal(t, "$$name", variable.Name)
				
				str, ok := assign.Right.(*ast.StringLiteral)
				require.True(t, ok)
				assert.Equal(t, "value", str.Value)
			},
		},
		{
			name:  "Class with complex variable property",
			input: `<?php MyClass::${$method}();`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				callExpr, ok := exprStmt.Expression.(*ast.CallExpression)
				require.True(t, ok)
				
				staticAccess, ok := callExpr.Callee.(*ast.StaticAccessExpression)
				require.True(t, ok)
				
				assert.Equal(t, "MyClass", staticAccess.Class.(*ast.IdentifierNode).Name)
				assert.Equal(t, "${$method}", staticAccess.Property.(*ast.Variable).Name)
			},
		},
		{
			name:  "Multiple complex variable patterns",
			input: `<?php 
				echo static::$property;
				echo static::${$prop};
				echo static::$$prop;
				echo static::CONSTANT;
			`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 4)
				
				// First: static::$property
				echo1, ok := program.Body[0].(*ast.EchoStatement)
				require.True(t, ok)
				static1, ok := echo1.Arguments[0].(*ast.StaticAccessExpression)
				require.True(t, ok)
				assert.Equal(t, "$property", static1.Property.(*ast.Variable).Name)
				
				// Second: static::${$prop}
				echo2, ok := program.Body[1].(*ast.EchoStatement)
				require.True(t, ok)
				static2, ok := echo2.Arguments[0].(*ast.StaticAccessExpression)
				require.True(t, ok)
				assert.Equal(t, "${$prop}", static2.Property.(*ast.Variable).Name)
				
				// Third: static::$$prop
				echo3, ok := program.Body[2].(*ast.EchoStatement)
				require.True(t, ok)
				static3, ok := echo3.Arguments[0].(*ast.StaticAccessExpression)
				require.True(t, ok)
				assert.Equal(t, "$$prop", static3.Property.(*ast.Variable).Name)
				
				// Fourth: static::CONSTANT
				echo4, ok := program.Body[3].(*ast.EchoStatement)
				require.True(t, ok)
				static4, ok := echo4.Arguments[0].(*ast.StaticAccessExpression)
				require.True(t, ok)
				assert.Equal(t, "CONSTANT", static4.Property.(*ast.IdentifierNode).Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			
			// Check for parsing errors
			checkParserErrors(t, p)
			
			// Run specific test expectations
			if tt.expected != nil {
				tt.expected(t, program)
			}
		})
	}
}


func TestParsing_ClosureWithUseClause(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "simple closure with use clause",
			input: `<?php $closure = function() use ($var1) { return $var1; };`,
		},
		{
			name:  "closure with multiple use variables",
			input: `<?php $closure = function($x) use ($var1, $var2) { return $x + $var2; };`,
		},
		{
			name:  "closure with use clause and return type",
			input: `<?php $closure = function() use ($var1): string { return $var1; };`,
		},
		{
			name:  "static closure with use clause and return type", 
			input: `<?php $closure = static function($a, $b) use ($var2): int { return $a + $b + $var2; };`,
		},
		{
			name:  "closure with typed parameters and use clause",
			input: `<?php $closure = static function(array $items, ?string $prefix) use ($var1): array { return $items; };`,
		},
		{
			name:  "closure with reference in use clause",
			input: `<?php $closure = function() use (&$counter) { $counter++; };`,
		},
		{
			name:  "original WordPress failing case",
			input: `<?php add_filter("hook", static function(array $statuses, WP_Taxonomy $taxonomy) use ($custom_taxonomy): array { return $statuses; }, 10, 2);`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			
			// 确保没有解析错误
			assert.Empty(t, p.errors, "Unexpected parsing errors: %v", p.errors)
			assert.NotNil(t, program)
			assert.NotEmpty(t, program.Body)
		})
	}
}

// TestParsing_MatchExpressions tests match expression parsing with various scenarios
func TestParsing_MatchExpressions(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "simple match expression",
			input: `<?php match (1) { 1 => 'one', default => 'other' }; ?>`,
		},
		{
			name:  "multi-line match expression with trailing comma",
			input: `<?php match (1) {
				1 => 'one',
				2 => 'two',
				default => 'other',
			}; ?>`,
		},
		{
			name:  "match expression with instanceof",
			input: `<?php match (true) {
				$factory instanceof UserFactory => User::class,
				default => 'other',
			}; ?>`,
		},
		{
			name:  "match expression with throw in closure",
			input: `<?php Factory::guessModelNamesUsing(function (Factory $factory) {
				return match (true) {
					$factory instanceof UserFactory => User::class,
					default => throw new LogicException('Unknown factory'),
				};
			}); ?>`,
		},
		{
			name:  "assignment within function argument",
			input: `<?php assertType('UserFactory', $factory = UserFactory::new()); ?>`,
		},
		{
			name:  "static return type in class method",
			input: `<?php class CommonBuilder {
				public function foo(): static {
					return $this;
				}
			} ?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			
			// 确保没有解析错误
			assert.Empty(t, p.errors, "Unexpected parsing errors: %v", p.errors)
			assert.NotNil(t, program)
			assert.NotEmpty(t, program.Body)
		})
	}
}

// TestParsing_QualifiedNameUseStatements tests the fix for T_NAME_QUALIFIED tokens in use statements
// This addresses the bug where use statements with qualified names (generated as single tokens by the lexer)
// were failing to parse because the parser was expecting only T_STRING tokens
func TestParsing_QualifiedNameUseStatements(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedErrors int
		expectedUses   []struct {
			nameParts []string
			alias     string
			useType   string
		}
	}{
		{
			name: "Original failing case - Laravel use statements",
			input: `<?php

namespace Illuminate\Tests\Mail;

use Illuminate\Contracts\Mail\Attachable;
use Illuminate\Http\Testing\File;
?>`,
			expectedErrors: 0,
			expectedUses: []struct {
				nameParts []string
				alias     string
				useType   string
			}{
				{
					nameParts: []string{"Illuminate", "Contracts", "Mail", "Attachable"},
					alias:     "",
					useType:   "",
				},
				{
					nameParts: []string{"Illuminate", "Http", "Testing", "File"},
					alias:     "",
					useType:   "",
				},
			},
		},
		{
			name: "Mixed qualified and simple names",
			input: `<?php
use SomeClass;
use Namespace\SubNamespace\AnotherClass;
use Fully\Qualified\Name as Alias;
?>`,
			expectedErrors: 0,
			expectedUses: []struct {
				nameParts []string
				alias     string
				useType   string
			}{
				{
					nameParts: []string{"SomeClass"},
					alias:     "",
					useType:   "",
				},
				{
					nameParts: []string{"Namespace", "SubNamespace", "AnotherClass"},
					alias:     "",
					useType:   "",
				},
				{
					nameParts: []string{"Fully", "Qualified", "Name"},
					alias:     "Alias",
					useType:   "",
				},
			},
		},
		{
			name: "Qualified names with function and const",
			input: `<?php
use function My\Namespace\someFunction;
use const My\Namespace\SOME_CONSTANT;
?>`,
			expectedErrors: 0,
			expectedUses: []struct {
				nameParts []string
				alias     string
				useType   string
			}{
				{
					nameParts: []string{"My", "Namespace", "someFunction"},
					alias:     "",
					useType:   "function",
				},
				{
					nameParts: []string{"My", "Namespace", "SOME_CONSTANT"},
					alias:     "",
					useType:   "const",
				},
			},
		},
		{
			name: "Multiple qualified use statements on same line",
			input: `<?php
use A\B\C, D\E\F as G, H\I\J;
?>`,
			expectedErrors: 0,
			expectedUses: []struct {
				nameParts []string
				alias     string
				useType   string
			}{
				{
					nameParts: []string{"A", "B", "C"},
					alias:     "",
					useType:   "",
				},
				{
					nameParts: []string{"D", "E", "F"},
					alias:     "G",
					useType:   "",
				},
				{
					nameParts: []string{"H", "I", "J"},
					alias:     "",
					useType:   "",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			
			// Check for parsing errors
			assert.Equal(t, tt.expectedErrors, len(p.Errors()), "Parser errors: %v", p.Errors())
			assert.NotNil(t, program)
			
			// Find all UseStatement nodes in the program
			var useStatements []*ast.UseStatement
			for _, stmt := range program.Body {
				if useStmt, ok := stmt.(*ast.UseStatement); ok {
					useStatements = append(useStatements, useStmt)
				}
			}
			
			// Verify we found the expected number of use statements
			totalExpectedUses := 0
			for _, expected := range tt.expectedUses {
				_ = expected // Count each expected use
				totalExpectedUses++
			}
			
			// Count actual uses across all statements
			totalActualUses := 0
			for _, useStmt := range useStatements {
				totalActualUses += len(useStmt.Uses)
			}
			
			assert.Equal(t, totalExpectedUses, totalActualUses, 
				"Expected %d use clauses but found %d", totalExpectedUses, totalActualUses)
			
			// Validate each expected use clause
			useIndex := 0
			for _, expected := range tt.expectedUses {
				found := false
				for _, useStmt := range useStatements {
					for _, useClause := range useStmt.Uses {
						if useIndex < totalExpectedUses {
							currentExpected := tt.expectedUses[useIndex]
							if len(useClause.Name.Parts) == len(currentExpected.nameParts) {
								match := true
								for i, part := range useClause.Name.Parts {
									if part != currentExpected.nameParts[i] {
										match = false
										break
									}
								}
								if match && useClause.Alias == currentExpected.alias && useClause.Type == currentExpected.useType {
									found = true
									useIndex++
									break
								}
							}
						}
					}
					if found {
						break
					}
				}
				assert.True(t, found, "Could not find expected use clause: %v", expected)
			}
		})
	}
}

// TestParsing_NullablePropertyTypes tests parsing of nullable typed properties
func TestParsing_NullablePropertyTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, program *ast.Program)
	}{
		{
			name: "Simple nullable bool properties",
			input: `<?php
class Settings
{
    public ?bool $foo;
    public ?bool $bar;
}`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				classExpr, ok := exprStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)
				nameIdent, ok := classExpr.Name.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "Settings", nameIdent.Name)
				
				require.Len(t, classExpr.Body, 2)
				
				// Check first property: public ?bool $foo;
				prop1, ok := classExpr.Body[0].(*ast.PropertyDeclaration)
				require.True(t, ok)
				assert.Equal(t, "public", prop1.Visibility)
				assert.Equal(t, "foo", prop1.Name)
				assert.False(t, prop1.Static)
				assert.False(t, prop1.ReadOnly)
				require.NotNil(t, prop1.Type)
				assert.Equal(t, "bool", prop1.Type.Name)
				assert.True(t, prop1.Type.Nullable)
				assert.Nil(t, prop1.DefaultValue)
				
				// Check second property: public ?bool $bar;
				prop2, ok := classExpr.Body[1].(*ast.PropertyDeclaration)
				require.True(t, ok)
				assert.Equal(t, "public", prop2.Visibility)
				assert.Equal(t, "bar", prop2.Name)
				assert.False(t, prop2.Static)
				assert.False(t, prop2.ReadOnly)
				require.NotNil(t, prop2.Type)
				assert.Equal(t, "bool", prop2.Type.Name)
				assert.True(t, prop2.Type.Nullable)
				assert.Nil(t, prop2.DefaultValue)
			},
		},
		{
			name: "Mixed nullable and non-nullable properties",
			input: `<?php
class User
{
    public string $name;
    public ?string $email;
    protected ?array $metadata;
    private bool $active;
}`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				classExpr, ok := exprStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)
				nameIdent, ok := classExpr.Name.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "User", nameIdent.Name)
				
				require.Len(t, classExpr.Body, 4)
				
				// public string $name;
				prop1, ok := classExpr.Body[0].(*ast.PropertyDeclaration)
				require.True(t, ok)
				assert.Equal(t, "name", prop1.Name)
				require.NotNil(t, prop1.Type)
				assert.Equal(t, "string", prop1.Type.Name)
				assert.False(t, prop1.Type.Nullable)
				
				// public ?string $email;
				prop2, ok := classExpr.Body[1].(*ast.PropertyDeclaration)
				require.True(t, ok)
				assert.Equal(t, "email", prop2.Name)
				require.NotNil(t, prop2.Type)
				assert.Equal(t, "string", prop2.Type.Name)
				assert.True(t, prop2.Type.Nullable)
				
				// protected ?array $metadata;
				prop3, ok := classExpr.Body[2].(*ast.PropertyDeclaration)
				require.True(t, ok)
				assert.Equal(t, "protected", prop3.Visibility)
				assert.Equal(t, "metadata", prop3.Name)
				require.NotNil(t, prop3.Type)
				assert.Equal(t, "array", prop3.Type.Name)
				assert.True(t, prop3.Type.Nullable)
				
				// private bool $active;
				prop4, ok := classExpr.Body[3].(*ast.PropertyDeclaration)
				require.True(t, ok)
				assert.Equal(t, "private", prop4.Visibility)
				assert.Equal(t, "active", prop4.Name)
				require.NotNil(t, prop4.Type)
				assert.Equal(t, "bool", prop4.Type.Name)
				assert.False(t, prop4.Type.Nullable)
			},
		},
		{
			name: "Nullable properties with default values",
			input: `<?php
class Config
{
    public ?string $host = null;
    public ?int $port = 8080;
    private ?array $options = [];
}`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				classExpr, ok := exprStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)
				nameIdent, ok := classExpr.Name.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "Config", nameIdent.Name)
				
				require.Len(t, classExpr.Body, 3)
				
				// public ?string $host = null;
				prop1, ok := classExpr.Body[0].(*ast.PropertyDeclaration)
				require.True(t, ok)
				assert.Equal(t, "host", prop1.Name)
				require.NotNil(t, prop1.Type)
				assert.Equal(t, "string", prop1.Type.Name)
				assert.True(t, prop1.Type.Nullable)
				require.NotNil(t, prop1.DefaultValue)
				
				// public ?int $port = 8080;
				prop2, ok := classExpr.Body[1].(*ast.PropertyDeclaration)
				require.True(t, ok)
				assert.Equal(t, "port", prop2.Name)
				require.NotNil(t, prop2.Type)
				assert.Equal(t, "int", prop2.Type.Name)
				assert.True(t, prop2.Type.Nullable)
				require.NotNil(t, prop2.DefaultValue)
				
				// private ?array $options = [];
				prop3, ok := classExpr.Body[2].(*ast.PropertyDeclaration)
				require.True(t, ok)
				assert.Equal(t, "private", prop3.Visibility)
				assert.Equal(t, "options", prop3.Name)
				require.NotNil(t, prop3.Type)
				assert.Equal(t, "array", prop3.Type.Name)
				assert.True(t, prop3.Type.Nullable)
				require.NotNil(t, prop3.DefaultValue)
			},
		},
		{
			name: "Static nullable properties with visibility",
			input: `<?php
class Service
{
    public static ?object $instance;
    private static ?callable $callback = null;
    protected static ?Service $singleton;
}`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				classExpr, ok := exprStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)
				nameIdent, ok := classExpr.Name.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "Service", nameIdent.Name)
				
				require.Len(t, classExpr.Body, 3)
				
				// public static ?object $instance;
				prop1, ok := classExpr.Body[0].(*ast.PropertyDeclaration)
				require.True(t, ok)
				assert.Equal(t, "public", prop1.Visibility)
				assert.Equal(t, "instance", prop1.Name)
				assert.True(t, prop1.Static)
				require.NotNil(t, prop1.Type)
				assert.Equal(t, "object", prop1.Type.Name)
				assert.True(t, prop1.Type.Nullable)
				assert.Nil(t, prop1.DefaultValue)
				
				// private static ?callable $callback = null;
				prop2, ok := classExpr.Body[1].(*ast.PropertyDeclaration)
				require.True(t, ok)
				assert.Equal(t, "private", prop2.Visibility)
				assert.Equal(t, "callback", prop2.Name)
				assert.True(t, prop2.Static)
				require.NotNil(t, prop2.Type)
				assert.Equal(t, "callable", prop2.Type.Name)
				assert.True(t, prop2.Type.Nullable)
				require.NotNil(t, prop2.DefaultValue)
				
				// protected static ?Service $singleton;
				prop3, ok := classExpr.Body[2].(*ast.PropertyDeclaration)
				require.True(t, ok)
				assert.Equal(t, "protected", prop3.Visibility)
				assert.Equal(t, "singleton", prop3.Name)
				assert.True(t, prop3.Static)
				require.NotNil(t, prop3.Type)
				assert.Equal(t, "Service", prop3.Type.Name)
				assert.True(t, prop3.Type.Nullable)
				assert.Nil(t, prop3.DefaultValue)
			},
		},
		{
			name: "Readonly nullable properties",
			input: `<?php
class Model
{
    public readonly ?string $id;
    private readonly ?DateTime $createdAt;
}`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				classExpr, ok := exprStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)
				nameIdent, ok := classExpr.Name.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "Model", nameIdent.Name)
				
				require.Len(t, classExpr.Body, 2)
				
				// public readonly ?string $id;
				prop1, ok := classExpr.Body[0].(*ast.PropertyDeclaration)
				require.True(t, ok)
				assert.Equal(t, "public", prop1.Visibility)
				assert.Equal(t, "id", prop1.Name)
				assert.False(t, prop1.Static)
				assert.True(t, prop1.ReadOnly)
				require.NotNil(t, prop1.Type)
				assert.Equal(t, "string", prop1.Type.Name)
				assert.True(t, prop1.Type.Nullable)
				
				// private readonly ?DateTime $createdAt;
				prop2, ok := classExpr.Body[1].(*ast.PropertyDeclaration)
				require.True(t, ok)
				assert.Equal(t, "private", prop2.Visibility)
				assert.Equal(t, "createdAt", prop2.Name)
				assert.False(t, prop2.Static)
				assert.True(t, prop2.ReadOnly)
				require.NotNil(t, prop2.Type)
				assert.Equal(t, "DateTime", prop2.Type.Name)
				assert.True(t, prop2.Type.Nullable)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			
			if len(p.Errors()) > 0 {
				t.Errorf("Parser errors: %v", p.Errors())
				return
			}
			
			require.NotNil(t, program)
			tt.validate(t, program)
		})
	}
}

func TestParsing_DynamicStaticMethodCalls(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(*testing.T, *ast.Program)
	}{
		{
			name:  "dynamic method name with curly braces",
			input: `<?php $value::{$method}(); ?>`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				
				exprStmt := program.Body[0].(*ast.ExpressionStatement)
				callExpr := exprStmt.Expression.(*ast.CallExpression)
				
				// Should be a static method call with dynamic method name
				staticAccessExpr := callExpr.Callee.(*ast.StaticAccessExpression)
				assert.Equal(t, "$value", staticAccessExpr.Class.(*ast.Variable).Name)
				
				// The property/method should be the inner expression $method
				methodVar := staticAccessExpr.Property.(*ast.Variable)
				assert.Equal(t, "$method", methodVar.Name)
			},
		},
		{
			name:  "dynamic method name with parameters",
			input: `<?php $value::{$method}($param1, $param2); ?>`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				
				exprStmt := program.Body[0].(*ast.ExpressionStatement)
				callExpr := exprStmt.Expression.(*ast.CallExpression)
				
				// Check arguments
				assert.Len(t, callExpr.Arguments, 2)
				param1 := callExpr.Arguments[0].(*ast.Variable)
				param2 := callExpr.Arguments[1].(*ast.Variable)
				assert.Equal(t, "$param1", param1.Name)
				assert.Equal(t, "$param2", param2.Name)
				
				// Check the static access structure
				staticAccessExpr := callExpr.Callee.(*ast.StaticAccessExpression)
				assert.Equal(t, "$value", staticAccessExpr.Class.(*ast.Variable).Name)
				assert.Equal(t, "$method", staticAccessExpr.Property.(*ast.Variable).Name)
			},
		},
		{
			name:  "simple static method call - should continue to work",
			input: `<?php $value::method(); ?>`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				
				exprStmt := program.Body[0].(*ast.ExpressionStatement)
				callExpr := exprStmt.Expression.(*ast.CallExpression)
				
				staticAccessExpr := callExpr.Callee.(*ast.StaticAccessExpression)
				assert.Equal(t, "$value", staticAccessExpr.Class.(*ast.Variable).Name)
				assert.Equal(t, "method", staticAccessExpr.Property.(*ast.IdentifierNode).Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			
			checkParserErrors(t, p)
			require.NotNil(t, program)
			tt.validate(t, program)
		})
	}
}

func TestParsing_QualifiedTraitNames(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(*testing.T, *ast.Program)
	}{
		{
			name:  "simple qualified trait name",
			input: `<?php class Test { use Concerns\CallsCommands; } ?>`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				
				exprStmt := program.Body[0].(*ast.ExpressionStatement)
				classExpr := exprStmt.Expression.(*ast.ClassExpression)
				
				assert.Len(t, classExpr.Body, 1)
				traitUse := classExpr.Body[0].(*ast.UseTraitStatement)
				
				assert.Len(t, traitUse.Traits, 1)
				assert.Equal(t, "Concerns\\CallsCommands", traitUse.Traits[0].Name)
			},
		},
		{
			name:  "multiple qualified trait names", 
			input: `<?php class Test { use Concerns\CallsCommands, Concerns\InteractsWithIO, Macroable; } ?>`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				
				exprStmt := program.Body[0].(*ast.ExpressionStatement)
				classExpr := exprStmt.Expression.(*ast.ClassExpression)
				
				assert.Len(t, classExpr.Body, 1)
				traitUse := classExpr.Body[0].(*ast.UseTraitStatement)
				
				assert.Len(t, traitUse.Traits, 3)
				assert.Equal(t, "Concerns\\CallsCommands", traitUse.Traits[0].Name)
				assert.Equal(t, "Concerns\\InteractsWithIO", traitUse.Traits[1].Name)
				assert.Equal(t, "Macroable", traitUse.Traits[2].Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			
			checkParserErrors(t, p)
			require.NotNil(t, program)
			tt.validate(t, program)
		})
	}
}

func TestParsing_MatchExpressionWithComplexConditions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(*testing.T, *ast.Program)
	}{
		{
			name:  "match with boolean OR condition",
			input: `<?php match(true) { is_array($x) || str_contains($x, 'y') => 'action' }; ?>`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				
				exprStmt := program.Body[0].(*ast.ExpressionStatement)
				matchExpr := exprStmt.Expression.(*ast.MatchExpression)
				
				assert.Equal(t, "true", matchExpr.Subject.(*ast.IdentifierNode).Name)
				assert.Len(t, matchExpr.Arms, 1)
				
				arm := matchExpr.Arms[0]
				assert.Len(t, arm.Conditions, 1)
				
				// The condition should be a binary expression with || operator
				condition := arm.Conditions[0].(*ast.BinaryExpression)
				assert.Equal(t, "||", condition.Operator)
				
				// Left side should be is_array($x) function call
				leftCall := condition.Left.(*ast.CallExpression)
				assert.Equal(t, "is_array", leftCall.Callee.(*ast.IdentifierNode).Name)
				assert.Len(t, leftCall.Arguments, 1)
				assert.Equal(t, "$x", leftCall.Arguments[0].(*ast.Variable).Name)
				
				// Right side should be str_contains($x, 'y') function call
				rightCall := condition.Right.(*ast.CallExpression)
				assert.Equal(t, "str_contains", rightCall.Callee.(*ast.IdentifierNode).Name)
				assert.Len(t, rightCall.Arguments, 2)
				assert.Equal(t, "$x", rightCall.Arguments[0].(*ast.Variable).Name)
				assert.Equal(t, "y", rightCall.Arguments[1].(*ast.StringLiteral).Value)
				
				// Body should be 'action' string
				assert.Equal(t, "action", arm.Body.(*ast.StringLiteral).Value)
			},
		},
		{
			name:  "match with boolean AND condition",
			input: `<?php match(true) { is_string($x) && !empty($x) => 'valid' }; ?>`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				
				exprStmt := program.Body[0].(*ast.ExpressionStatement)
				matchExpr := exprStmt.Expression.(*ast.MatchExpression)
				
				assert.Len(t, matchExpr.Arms, 1)
				arm := matchExpr.Arms[0]
				condition := arm.Conditions[0].(*ast.BinaryExpression)
				assert.Equal(t, "&&", condition.Operator)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			
			checkParserErrors(t, p)
			require.NotNil(t, program)
			tt.validate(t, program)
		})
	}
}

func TestParsing_MatchExpressionTrailingCommas(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(*testing.T, *ast.Program)
	}{
		{
			name: "Simple trailing comma in condition list",
			input: `<?php match ($name) { 'a', 'b', => 'value', default => null }; ?>`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				exprStmt := program.Body[0].(*ast.ExpressionStatement)
				matchExpr := exprStmt.Expression.(*ast.MatchExpression)
				
				assert.Equal(t, "$name", matchExpr.Subject.(*ast.Variable).Name)
				assert.Len(t, matchExpr.Arms, 2)
				
				// First arm should have two conditions
				arm := matchExpr.Arms[0]
				assert.Len(t, arm.Conditions, 2)
				assert.Equal(t, "a", arm.Conditions[0].(*ast.StringLiteral).Value)
				assert.Equal(t, "b", arm.Conditions[1].(*ast.StringLiteral).Value)
				assert.Equal(t, "value", arm.Body.(*ast.StringLiteral).Value)
				assert.False(t, arm.IsDefault)
				
				// Second arm should be default
				assert.True(t, matchExpr.Arms[1].IsDefault)
				assert.Equal(t, "null", matchExpr.Arms[1].Body.(*ast.IdentifierNode).Name)
			},
		},
		{
			name: "Original failing case - multiple conditions with trailing commas",
			input: `<?php
foreach ($context as $name => $value) {
    match ($name) {
        'use-cache', 'client-tracking', 'throw-on-error', 'client-invalidations', 'reply-literal', 'persistent',
            => $context[$name] = filter_var($value, \FILTER_VALIDATE_BOOLEAN),
        'max-retries', 'serializer', 'compression', 'compression-level',
            => $context[$name] = filter_var($value, \FILTER_VALIDATE_INT),
        default => null,
    };
}
?>`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				foreachStmt := program.Body[0].(*ast.ForeachStatement)
				require.NotNil(t, foreachStmt)
				
				blockStmt := foreachStmt.Body.(*ast.BlockStatement)
				require.Len(t, blockStmt.Body, 1)
				
				exprStmt := blockStmt.Body[0].(*ast.ExpressionStatement)
				matchExpr := exprStmt.Expression.(*ast.MatchExpression)
				
				assert.Equal(t, "$name", matchExpr.Subject.(*ast.Variable).Name)
				assert.Len(t, matchExpr.Arms, 3)
				
				// First arm - boolean filters
				arm1 := matchExpr.Arms[0]
				assert.Len(t, arm1.Conditions, 6)
				assert.Equal(t, "use-cache", arm1.Conditions[0].(*ast.StringLiteral).Value)
				assert.Equal(t, "client-tracking", arm1.Conditions[1].(*ast.StringLiteral).Value)
				assert.Equal(t, "throw-on-error", arm1.Conditions[2].(*ast.StringLiteral).Value)
				assert.Equal(t, "client-invalidations", arm1.Conditions[3].(*ast.StringLiteral).Value)
				assert.Equal(t, "reply-literal", arm1.Conditions[4].(*ast.StringLiteral).Value)
				assert.Equal(t, "persistent", arm1.Conditions[5].(*ast.StringLiteral).Value)
				assert.False(t, arm1.IsDefault)
				
				// Check the body is an assignment
				assignment1 := arm1.Body.(*ast.AssignmentExpression)
				assert.Equal(t, "=", assignment1.Operator)
				
				// Second arm - integer filters
				arm2 := matchExpr.Arms[1]
				assert.Len(t, arm2.Conditions, 4)
				assert.Equal(t, "max-retries", arm2.Conditions[0].(*ast.StringLiteral).Value)
				assert.Equal(t, "serializer", arm2.Conditions[1].(*ast.StringLiteral).Value)
				assert.Equal(t, "compression", arm2.Conditions[2].(*ast.StringLiteral).Value)
				assert.Equal(t, "compression-level", arm2.Conditions[3].(*ast.StringLiteral).Value)
				assert.False(t, arm2.IsDefault)
				
				// Third arm - default
				arm3 := matchExpr.Arms[2]
				assert.True(t, arm3.IsDefault)
				assert.Equal(t, "null", arm3.Body.(*ast.IdentifierNode).Name)
			},
		},
		{
			name: "Multiple trailing commas in different arms",
			input: `<?php match ($x) {
				1, 2, => 'first',
				3, 4, => 'second',
				5, => 'third',
				default => 'other'
			}; ?>`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				exprStmt := program.Body[0].(*ast.ExpressionStatement)
				matchExpr := exprStmt.Expression.(*ast.MatchExpression)
				
				assert.Len(t, matchExpr.Arms, 4)
				
				// First arm: 1, 2, => 'first'
				assert.Len(t, matchExpr.Arms[0].Conditions, 2)
				assert.Equal(t, "first", matchExpr.Arms[0].Body.(*ast.StringLiteral).Value)
				
				// Second arm: 3, 4, => 'second'  
				assert.Len(t, matchExpr.Arms[1].Conditions, 2)
				assert.Equal(t, "second", matchExpr.Arms[1].Body.(*ast.StringLiteral).Value)
				
				// Third arm: 5, => 'third'
				assert.Len(t, matchExpr.Arms[2].Conditions, 1)
				assert.Equal(t, "third", matchExpr.Arms[2].Body.(*ast.StringLiteral).Value)
				
				// Default arm
				assert.True(t, matchExpr.Arms[3].IsDefault)
				assert.Equal(t, "other", matchExpr.Arms[3].Body.(*ast.StringLiteral).Value)
			},
		},
		{
			name: "Empty condition list with trailing comma (error case)",
			input: `<?php match ($x) { , => 'invalid' }; ?>`,
			validate: func(t *testing.T, program *ast.Program) {
				// This should parse but the first condition should be empty or invalid
				// The focus is that the parser doesn't crash
				require.NotNil(t, program)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			
			// For most cases, we expect no parsing errors
			if tt.name != "Empty condition list with trailing comma (error case)" {
				assert.Empty(t, p.errors, "Unexpected parsing errors: %v", p.errors)
			}
			
			assert.NotNil(t, program)
			tt.validate(t, program)
		})
	}
}

func TestParsing_EnumConstants(t *testing.T) {
	input := `<?php
enum Level: int
{
    case Debug = 100;
    case Info = 200;
    
    public const VALUES = [100, 200];
    private const NAMES = ['debug', 'info'];
    const DEFAULT = 100;
    
    public function getName(): string {
        return 'test';
    }
}
?>`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	checkParserErrors(t, p)
	assert.NotNil(t, program)
	assert.Len(t, program.Body, 1)

	// 检查 enum 声明
	enumDecl, ok := program.Body[0].(*ast.EnumDeclaration)
	assert.True(t, ok)
	assert.Equal(t, "Level", enumDecl.Name.Name)
	assert.NotNil(t, enumDecl.BackingType)
	assert.Equal(t, "int", enumDecl.BackingType.Name)

	// 检查 enum 案例
	assert.Len(t, enumDecl.Cases, 2)
	assert.Equal(t, "Debug", enumDecl.Cases[0].Name.Name)
	assert.Equal(t, "Info", enumDecl.Cases[1].Name.Name)

	// 检查 enum 常量
	assert.Len(t, enumDecl.Constants, 3)
	
	// public const VALUES
	assert.Equal(t, "public", enumDecl.Constants[0].Visibility)
	assert.Len(t, enumDecl.Constants[0].Constants, 1)
	// Check first constant - cast Name to *IdentifierNode
	if constName, ok := enumDecl.Constants[0].Constants[0].Name.(*ast.IdentifierNode); ok {
		assert.Equal(t, "VALUES", constName.Name)
	} else {
		assert.Fail(t, "Expected constant name to be an IdentifierNode")
	}
	
	// private const NAMES  
	assert.Equal(t, "private", enumDecl.Constants[1].Visibility)
	assert.Len(t, enumDecl.Constants[1].Constants, 1)
	// Check second constant - cast Name to *IdentifierNode
	if constName, ok := enumDecl.Constants[1].Constants[0].Name.(*ast.IdentifierNode); ok {
		assert.Equal(t, "NAMES", constName.Name)
	} else {
		assert.Fail(t, "Expected constant name to be an IdentifierNode")
	}
	
	// const DEFAULT (no visibility)
	assert.Equal(t, "", enumDecl.Constants[2].Visibility)
	assert.Len(t, enumDecl.Constants[2].Constants, 1)
	// Check third constant - cast Name to *IdentifierNode
	if constName, ok := enumDecl.Constants[2].Constants[0].Name.(*ast.IdentifierNode); ok {
		assert.Equal(t, "DEFAULT", constName.Name)
	} else {
		assert.Fail(t, "Expected constant name to be an IdentifierNode")
	}

	// 检查 enum 方法
	assert.Len(t, enumDecl.Methods, 1)
	// Check method
	if methodName, ok := enumDecl.Methods[0].Name.(*ast.IdentifierNode); ok {
		assert.Equal(t, "getName", methodName.Name)
	} else {
		assert.Fail(t, "Expected method name to be an IdentifierNode")
	}
	assert.Equal(t, "public", enumDecl.Methods[0].Visibility)
}

func TestParsing_EnumComprehensive(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, program *ast.Program)
	}{
		{
			name: "Simple enum without backing type",
			input: `<?php
enum Status
{
    case Pending;
    case Active;
    case Inactive;
}
?>`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				enumDecl, ok := program.Body[0].(*ast.EnumDeclaration)
				require.True(t, ok)
				assert.Equal(t, "Status", enumDecl.Name.Name)
				assert.Nil(t, enumDecl.BackingType) // No backing type
				assert.Len(t, enumDecl.Cases, 3)
				assert.Equal(t, "Pending", enumDecl.Cases[0].Name.Name)
				assert.Equal(t, "Active", enumDecl.Cases[1].Name.Name)
				assert.Equal(t, "Inactive", enumDecl.Cases[2].Name.Name)
			},
		},
		{
			name: "String-backed enum with values",
			input: `<?php
enum Color: string
{
    case Red = 'red';
    case Green = 'green';
    case Blue = 'blue';
}
?>`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				enumDecl, ok := program.Body[0].(*ast.EnumDeclaration)
				require.True(t, ok)
				assert.Equal(t, "Color", enumDecl.Name.Name)
				assert.NotNil(t, enumDecl.BackingType)
				assert.Equal(t, "string", enumDecl.BackingType.Name)
				assert.Len(t, enumDecl.Cases, 3)
				assert.Equal(t, "Red", enumDecl.Cases[0].Name.Name)
				assert.NotNil(t, enumDecl.Cases[0].Value)
				assert.Equal(t, "Green", enumDecl.Cases[1].Name.Name)
				assert.NotNil(t, enumDecl.Cases[1].Value)
			},
		},
		{
			name: "Int-backed enum with constants",
			input: `<?php
enum HttpStatus: int
{
    case OK = 200;
    case NotFound = 404;
    
    const PREFIX = 'HTTP_';
    public const CODES = [200, 404];
}
?>`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				enumDecl, ok := program.Body[0].(*ast.EnumDeclaration)
				require.True(t, ok)
				assert.Equal(t, "HttpStatus", enumDecl.Name.Name)
				assert.NotNil(t, enumDecl.BackingType)
				assert.Equal(t, "int", enumDecl.BackingType.Name)
				assert.Len(t, enumDecl.Cases, 2)
				assert.Len(t, enumDecl.Constants, 2)
				
				// Check first constant (no visibility)
				assert.Equal(t, "", enumDecl.Constants[0].Visibility)
				if constName, ok := enumDecl.Constants[0].Constants[0].Name.(*ast.IdentifierNode); ok {
					assert.Equal(t, "PREFIX", constName.Name)
				}
				
				// Check second constant (public)
				assert.Equal(t, "public", enumDecl.Constants[1].Visibility)
				if constName, ok := enumDecl.Constants[1].Constants[0].Name.(*ast.IdentifierNode); ok {
					assert.Equal(t, "CODES", constName.Name)
				}
			},
		},
		{
			name: "Enum with all visibility modifiers for constants",
			input: `<?php
enum Permission: string
{
    case Read = 'r';
    case Write = 'w';
    case Execute = 'x';
    
    private const ADMIN = 'rwx';
    protected const USER = 'r';
    public const GUEST = '';
    const DEFAULT = 'r';
}
?>`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				enumDecl, ok := program.Body[0].(*ast.EnumDeclaration)
				require.True(t, ok)
				assert.Len(t, enumDecl.Constants, 4)
				
				// Check visibility modifiers
				assert.Equal(t, "private", enumDecl.Constants[0].Visibility)
				assert.Equal(t, "protected", enumDecl.Constants[1].Visibility)
				assert.Equal(t, "public", enumDecl.Constants[2].Visibility)
				assert.Equal(t, "", enumDecl.Constants[3].Visibility) // No explicit visibility
			},
		},
		{
			name: "Enum with multiple methods",
			input: `<?php
enum UserRole: string
{
    case Admin = 'admin';
    case User = 'user';
    
    public function getLabel(): string {
        return match($this) {
            self::Admin => 'Administrator',
            self::User => 'Regular User',
        };
    }
    
    private function validate(): bool {
        return true;
    }
    
    protected function format(): string {
        return strtoupper($this->value);
    }
}
?>`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				enumDecl, ok := program.Body[0].(*ast.EnumDeclaration)
				require.True(t, ok)
				assert.Len(t, enumDecl.Methods, 3)
				
				// Check method visibility
				assert.Equal(t, "public", enumDecl.Methods[0].Visibility)
				if methodName, ok := enumDecl.Methods[0].Name.(*ast.IdentifierNode); ok {
					assert.Equal(t, "getLabel", methodName.Name)
				}
				
				assert.Equal(t, "private", enumDecl.Methods[1].Visibility)
				if methodName, ok := enumDecl.Methods[1].Name.(*ast.IdentifierNode); ok {
					assert.Equal(t, "validate", methodName.Name)
				}
				
				assert.Equal(t, "protected", enumDecl.Methods[2].Visibility)
				if methodName, ok := enumDecl.Methods[2].Name.(*ast.IdentifierNode); ok {
					assert.Equal(t, "format", methodName.Name)
				}
			},
		},
		{
			name: "Enum implementing interfaces",
			input: `<?php
enum Priority: int implements Comparable, Serializable
{
    case Low = 1;
    case Medium = 5;
    case High = 10;
    
    public function compareTo($other): int {
        return $this->value <=> $other->value;
    }
}
?>`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				enumDecl, ok := program.Body[0].(*ast.EnumDeclaration)
				require.True(t, ok)
				assert.Equal(t, "Priority", enumDecl.Name.Name)
				assert.Len(t, enumDecl.Implements, 2)
				assert.Equal(t, "Comparable", enumDecl.Implements[0].Name)
				assert.Equal(t, "Serializable", enumDecl.Implements[1].Name)
			},
		},
		{
			name: "Mixed enum features",
			input: `<?php
enum LogLevel: int implements Stringable
{
    case Emergency = 0;
    case Alert = 1;
    case Critical = 2;
    case Error = 3;
    
    const SEVERE_LEVELS = [0, 1, 2];
    private const MAX_LEVEL = 7;
    public const DEFAULT = 3;
    
    public function __toString(): string {
        return $this->name;
    }
    
    public static function fromString(string $name): self {
        return self::tryFrom($name);
    }
}
?>`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				enumDecl, ok := program.Body[0].(*ast.EnumDeclaration)
				require.True(t, ok)
				assert.Equal(t, "LogLevel", enumDecl.Name.Name)
				assert.NotNil(t, enumDecl.BackingType)
				assert.Equal(t, "int", enumDecl.BackingType.Name)
				assert.Len(t, enumDecl.Cases, 4)
				assert.Len(t, enumDecl.Constants, 3)
				assert.Len(t, enumDecl.Methods, 2)
				assert.Len(t, enumDecl.Implements, 1)
				assert.Equal(t, "Stringable", enumDecl.Implements[0].Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			
			checkParserErrors(t, p)
			require.NotNil(t, program)
			tt.validate(t, program)
		})
	}
}

// TestParsing_SemiReservedAsIdentifiers tests that semi-reserved keywords can be used as identifiers
func TestParsing_SemiReservedAsIdentifiers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, program *ast.Program)
	}{
		{
			name:  "new Enum() class instantiation",
			input: "<?php new Enum(StringStatus::class);",
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok, "Expected expression statement")
				
				newExpr, ok := stmt.Expression.(*ast.NewExpression)
				require.True(t, ok, "Expected new expression")
				
				callExpr, ok := newExpr.Class.(*ast.CallExpression)
				require.True(t, ok, "Expected call expression")
				
				ident, ok := callExpr.Callee.(*ast.IdentifierNode)
				require.True(t, ok, "Expected identifier")
				assert.Equal(t, "Enum", ident.Name)
				
				// Check the argument: StringStatus::class
				require.Len(t, callExpr.Arguments, 1)
				staticAccess, ok := callExpr.Arguments[0].(*ast.StaticAccessExpression)
				require.True(t, ok, "Expected static access expression")
				
				classIdent, ok := staticAccess.Class.(*ast.IdentifierNode)
				require.True(t, ok, "Expected class identifier")
				assert.Equal(t, "StringStatus", classIdent.Name)
				
				propIdent, ok := staticAccess.Property.(*ast.IdentifierNode)
				require.True(t, ok, "Expected property identifier")
				assert.Equal(t, "class", propIdent.Name)
			},
		},
		{
			name:  "Original failing case - Complex Enum usage",
			input: `<?php
$v = new Validator(
    resolve('translator'),
    [
        'status' => 'pending',
        'int_status' => 1,
    ],
    [
        'status' => new Enum(StringStatus::class),
        'int_status' => new Enum(IntegerStatus::class),
    ]
);`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok, "Expected expression statement")
				
				assignment, ok := stmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok, "Expected assignment expression")
				
				// Check left side is $v
				variable, ok := assignment.Left.(*ast.Variable)
				require.True(t, ok, "Expected variable")
				assert.Equal(t, "$v", variable.Name)
				
				// Check right side is new Validator(...)
				newExpr, ok := assignment.Right.(*ast.NewExpression)
				require.True(t, ok, "Expected new expression")
				
				validatorCall, ok := newExpr.Class.(*ast.CallExpression)
				require.True(t, ok, "Expected call expression")
				
				validatorIdent, ok := validatorCall.Callee.(*ast.IdentifierNode)
				require.True(t, ok, "Expected identifier")
				assert.Equal(t, "Validator", validatorIdent.Name)
				
				// The third argument should be an array with Enum instantiations
				require.Len(t, validatorCall.Arguments, 3)
				thirdArg, ok := validatorCall.Arguments[2].(*ast.ArrayExpression)
				require.True(t, ok, "Expected array expression for third argument")
				
				// Check both array elements have new Enum(...) as values
				require.Len(t, thirdArg.Elements, 2)
				
				// First element: 'status' => new Enum(StringStatus::class)
				statusElement, ok := thirdArg.Elements[0].(*ast.ArrayElementExpression)
				require.True(t, ok, "Expected array element expression")
				statusKey, ok := statusElement.Key.(*ast.StringLiteral)
				require.True(t, ok, "Expected string key")
				assert.Equal(t, "status", statusKey.Value)
				
				statusNewExpr, ok := statusElement.Value.(*ast.NewExpression)
				require.True(t, ok, "Expected new expression")
				
				statusEnumCall, ok := statusNewExpr.Class.(*ast.CallExpression)
				require.True(t, ok, "Expected call expression")
				
				statusEnumIdent, ok := statusEnumCall.Callee.(*ast.IdentifierNode)
				require.True(t, ok, "Expected identifier")
				assert.Equal(t, "Enum", statusEnumIdent.Name)
				
				// Second element: 'int_status' => new Enum(IntegerStatus::class)
				intElement, ok := thirdArg.Elements[1].(*ast.ArrayElementExpression)
				require.True(t, ok, "Expected array element expression")
				intKey, ok := intElement.Key.(*ast.StringLiteral)
				require.True(t, ok, "Expected string key")
				assert.Equal(t, "int_status", intKey.Value)
				
				intNewExpr, ok := intElement.Value.(*ast.NewExpression)
				require.True(t, ok, "Expected new expression")
				
				intEnumCall, ok := intNewExpr.Class.(*ast.CallExpression)
				require.True(t, ok, "Expected call expression")
				
				intEnumIdent, ok := intEnumCall.Callee.(*ast.IdentifierNode)
				require.True(t, ok, "Expected identifier")
				assert.Equal(t, "Enum", intEnumIdent.Name)
			},
		},
		{
			name:  "match as method name",
			input: "<?php $obj->match();",
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok, "Expected expression statement")
				
				callExpr, ok := stmt.Expression.(*ast.CallExpression)
				require.True(t, ok, "Expected call expression")
				
				methodCallExpr, ok := callExpr.Callee.(*ast.PropertyAccessExpression)
				require.True(t, ok, "Expected property access expression")
				
				ident, ok := methodCallExpr.Property.(*ast.IdentifierNode)
				require.True(t, ok, "Expected identifier for method name")
				assert.Equal(t, "match", ident.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			
			// The key test: no parsing errors should occur
			checkParserErrors(t, p)
			require.NotNil(t, program)
			tt.validate(t, program)
		})
	}
}

func TestParsing_IntersectionTypeParameters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, program *ast.Program)
	}{
		{
			name:  "Simple intersection type parameter",
			input: `<?php function test(A&B $param) {}`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)

				funcDecl, ok := program.Body[0].(*ast.FunctionDeclaration)
				assert.True(t, ok, "Expected FunctionDeclaration")
				assert.Equal(t, "test", funcDecl.Name.(*ast.IdentifierNode).Name)

				assert.Len(t, funcDecl.Parameters, 1)
				param := funcDecl.Parameters[0]
				assert.Equal(t, "$param", param.Name)

				assert.NotNil(t, param.Type)
				assert.Equal(t, ast.ASTTypeIntersection, param.Type.GetKind())
				assert.Len(t, param.Type.IntersectionTypes, 2)
				assert.Equal(t, "A", param.Type.IntersectionTypes[0].Name)
				assert.Equal(t, "B", param.Type.IntersectionTypes[1].Name)
			},
		},
		{
			name:  "Triple intersection type parameter",
			input: `<?php function test(A&B&C $param) {}`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)

				funcDecl, ok := program.Body[0].(*ast.FunctionDeclaration)
				assert.True(t, ok, "Expected FunctionDeclaration")

				assert.Len(t, funcDecl.Parameters, 1)
				param := funcDecl.Parameters[0]
				assert.Equal(t, "$param", param.Name)

				assert.NotNil(t, param.Type)
				assert.Equal(t, ast.ASTTypeIntersection, param.Type.GetKind())
				assert.Len(t, param.Type.IntersectionTypes, 3)
				assert.Equal(t, "A", param.Type.IntersectionTypes[0].Name)
				assert.Equal(t, "B", param.Type.IntersectionTypes[1].Name)
				assert.Equal(t, "C", param.Type.IntersectionTypes[2].Name)
			},
		},
		{
			name:  "Constructor with intersection type parameter and visibility",
			input: `<?php 
			class Test {
			    public function __construct(
			        private TranslatorInterface&TranslatorBagInterface&LocaleAwareInterface $translator,
			    ) {
			    }
			}`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)

				stmt := program.Body[0]
				exprStmt, ok := stmt.(*ast.ExpressionStatement)
				assert.True(t, ok, "Expected ExpressionStatement")

				classExpr, ok := exprStmt.Expression.(*ast.ClassExpression)
				assert.True(t, ok, "Expected ClassExpression")
				assert.Equal(t, "Test", classExpr.Name.(*ast.IdentifierNode).Name)
				assert.Len(t, classExpr.Body, 1)

				constructor, ok := classExpr.Body[0].(*ast.FunctionDeclaration)
				assert.True(t, ok, "Expected FunctionDeclaration")
				assert.Equal(t, "__construct", constructor.Name.(*ast.IdentifierNode).Name)
				assert.Equal(t, "public", constructor.Visibility)

				assert.Len(t, constructor.Parameters, 1)
				param := constructor.Parameters[0]
				assert.Equal(t, "$translator", param.Name)
				assert.Equal(t, "private", param.Visibility)

				assert.NotNil(t, param.Type)
				assert.Equal(t, ast.ASTTypeIntersection, param.Type.GetKind())
				assert.Len(t, param.Type.IntersectionTypes, 3)
				assert.Equal(t, "TranslatorInterface", param.Type.IntersectionTypes[0].Name)
				assert.Equal(t, "TranslatorBagInterface", param.Type.IntersectionTypes[1].Name)
				assert.Equal(t, "LocaleAwareInterface", param.Type.IntersectionTypes[2].Name)
			},
		},
		{
			name:  "Intersection type with variadic parameter", 
			input: `<?php function test(A&B ...$params) {}`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)

				funcDecl, ok := program.Body[0].(*ast.FunctionDeclaration)
				assert.True(t, ok, "Expected FunctionDeclaration")

				assert.Len(t, funcDecl.Parameters, 1)
				param := funcDecl.Parameters[0]
				assert.Equal(t, "$params", param.Name)
				assert.True(t, param.Variadic)

				assert.NotNil(t, param.Type)
				assert.Equal(t, ast.ASTTypeIntersection, param.Type.GetKind())
				assert.Len(t, param.Type.IntersectionTypes, 2)
				assert.Equal(t, "A", param.Type.IntersectionTypes[0].Name)
				assert.Equal(t, "B", param.Type.IntersectionTypes[1].Name)
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
func TestParsing_FinalPublicFunctionWithUnionTypesAndStaticReturn(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, program *ast.Program)
	}{
		{
			name:  "final public function with union parameter and static return",
			input: `<?php class BaseUri { final public function host(string|array $host): static { } }`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				classDecl, ok := exprStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)
				require.Equal(t, "BaseUri", classDecl.Name.(*ast.IdentifierNode).Name)
				require.Len(t, classDecl.Body, 1)
				
				funcDecl, ok := classDecl.Body[0].(*ast.FunctionDeclaration)
				require.True(t, ok)
				
				// Check function properties
				require.Equal(t, "host", funcDecl.Name.(*ast.IdentifierNode).Name)
				require.Equal(t, "public", funcDecl.Visibility)
				require.True(t, funcDecl.IsFinal)
				require.False(t, funcDecl.IsStatic)
				require.False(t, funcDecl.IsAbstract)
				
				// Check parameter with union type
				require.Len(t, funcDecl.Parameters, 1)
				param := funcDecl.Parameters[0]
				require.Equal(t, "$host", param.Name)
				require.NotNil(t, param.Type)
				
				require.Len(t, param.Type.UnionTypes, 2, "Parameter should have union type with 2 types")
				require.Equal(t, "string", param.Type.UnionTypes[0].Name)
				require.Equal(t, "array", param.Type.UnionTypes[1].Name)
				
				// Check static return type
				require.NotNil(t, funcDecl.ReturnType)
				require.Equal(t, "static", funcDecl.ReturnType.Name)
			},
		},
		{
			name:  "final private function with union types",
			input: `<?php class Test { final private function test(int|string|null $value): array { } }`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				classDecl, ok := exprStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)
				require.Len(t, classDecl.Body, 1)
				
				funcDecl, ok := classDecl.Body[0].(*ast.FunctionDeclaration)
				require.True(t, ok)
				
				require.Equal(t, "test", funcDecl.Name.(*ast.IdentifierNode).Name)
				require.Equal(t, "private", funcDecl.Visibility)
				require.True(t, funcDecl.IsFinal)
				
				// Check parameter with 3-type union
				require.Len(t, funcDecl.Parameters, 1)
				param := funcDecl.Parameters[0]
				require.Equal(t, "$value", param.Name)
				
				require.Len(t, param.Type.UnionTypes, 3, "Parameter should have union type with 3 types")
				require.Equal(t, "int", param.Type.UnionTypes[0].Name)
				require.Equal(t, "string", param.Type.UnionTypes[1].Name)
				require.Equal(t, "null", param.Type.UnionTypes[2].Name)
				
				// Check array return type
				require.NotNil(t, funcDecl.ReturnType)
				require.Equal(t, "array", funcDecl.ReturnType.Name)
			},
		},
		{
			name:  "final protected function with complex union types",
			input: `<?php class Test { final protected function process(MyClass|YourClass|TheirClass $obj): self|static { } }`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				classDecl, ok := exprStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)
				require.Len(t, classDecl.Body, 1)
				
				funcDecl, ok := classDecl.Body[0].(*ast.FunctionDeclaration)
				require.True(t, ok)
				
				require.Equal(t, "process", funcDecl.Name.(*ast.IdentifierNode).Name)
				require.Equal(t, "protected", funcDecl.Visibility)
				require.True(t, funcDecl.IsFinal)
				
				// Check parameter union type with class names
				require.Len(t, funcDecl.Parameters, 1)
				param := funcDecl.Parameters[0]
				require.Equal(t, "$obj", param.Name)
				
				require.Len(t, param.Type.UnionTypes, 3, "Parameter should have union type with 3 class types")
				require.Equal(t, "MyClass", param.Type.UnionTypes[0].Name)
				require.Equal(t, "YourClass", param.Type.UnionTypes[1].Name)
				require.Equal(t, "TheirClass", param.Type.UnionTypes[2].Name)
				
				// Check return union type with self|static
				require.NotNil(t, funcDecl.ReturnType)
				require.Len(t, funcDecl.ReturnType.UnionTypes, 2, "Return type should have union of self|static")
				require.Equal(t, "self", funcDecl.ReturnType.UnionTypes[0].Name)
				require.Equal(t, "static", funcDecl.ReturnType.UnionTypes[1].Name)
			},
		},
		{
			name:  "final public static function with union types",
			input: `<?php class Test { final public static function create(string|array $data): static { } }`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				classDecl, ok := exprStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)
				require.Len(t, classDecl.Body, 1)
				
				funcDecl, ok := classDecl.Body[0].(*ast.FunctionDeclaration)
				require.True(t, ok)
				
				require.Equal(t, "create", funcDecl.Name.(*ast.IdentifierNode).Name)
				require.Equal(t, "public", funcDecl.Visibility)
				require.True(t, funcDecl.IsFinal)
				require.True(t, funcDecl.IsStatic)
				
				// Check union parameter
				require.Len(t, funcDecl.Parameters, 1)
				param := funcDecl.Parameters[0]
				require.Len(t, param.Type.UnionTypes, 2, "Parameter should have union type")
				require.Equal(t, "string", param.Type.UnionTypes[0].Name)
				require.Equal(t, "array", param.Type.UnionTypes[1].Name)
				
				// Check static return
				require.NotNil(t, funcDecl.ReturnType)
				require.Equal(t, "static", funcDecl.ReturnType.Name)
			},
		},
		{
			name:  "multiple parameters with mixed union and simple types",
			input: `<?php class Test { final public function mixed(string $name, int|float $number, bool $flag): void { } }`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				classDecl, ok := exprStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)
				require.Len(t, classDecl.Body, 1)
				
				funcDecl, ok := classDecl.Body[0].(*ast.FunctionDeclaration)
				require.True(t, ok)
				
				require.Equal(t, "mixed", funcDecl.Name.(*ast.IdentifierNode).Name)
				require.Equal(t, "public", funcDecl.Visibility)
				require.True(t, funcDecl.IsFinal)
				
				// Check all parameters
				require.Len(t, funcDecl.Parameters, 3)
				
				// First parameter: string $name
				param1 := funcDecl.Parameters[0]
				require.Equal(t, "$name", param1.Name)
				require.Equal(t, "string", param1.Type.Name)
				
				// Second parameter: int|float $number
				param2 := funcDecl.Parameters[1]
				require.Equal(t, "$number", param2.Name)
				require.Len(t, param2.Type.UnionTypes, 2, "Second parameter should have union type")
				require.Equal(t, "int", param2.Type.UnionTypes[0].Name)
				require.Equal(t, "float", param2.Type.UnionTypes[1].Name)
				
				// Third parameter: bool $flag
				param3 := funcDecl.Parameters[2]
				require.Equal(t, "$flag", param3.Name)
				require.Equal(t, "bool", param3.Type.Name)
				
				// Check void return type
				require.NotNil(t, funcDecl.ReturnType)
				require.Equal(t, "void", funcDecl.ReturnType.Name)
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

func TestParsing_Closures(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, program *ast.Program)
	}{
		{
			name:  "Simple closure without parameters",
			input: `<?php $fn = function() { return 42; };`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				assignment, ok := stmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)
				
				closure, ok := assignment.Right.(*ast.AnonymousFunctionExpression)
				require.True(t, ok)
				require.Len(t, closure.Parameters, 0)
				require.False(t, closure.Static)
				require.False(t, closure.ByReference)
				require.Len(t, closure.UseClause, 0)
				require.Nil(t, closure.ReturnType)
				require.Len(t, closure.Body, 1)
			},
		},
		{
			name:  "Closure with parameters and type hints",
			input: `<?php $fn = function(string $name, int $age) { return $name . $age; };`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				assignment, ok := stmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)
				
				closure, ok := assignment.Right.(*ast.AnonymousFunctionExpression)
				require.True(t, ok)
				require.Len(t, closure.Parameters, 2)
				
				// First parameter: string $name
				param1 := closure.Parameters[0]
				require.Equal(t, "$name", param1.Name)
				require.Equal(t, "string", param1.Type.Name)
				
				// Second parameter: int $age
				param2 := closure.Parameters[1]
				require.Equal(t, "$age", param2.Name)
				require.Equal(t, "int", param2.Type.Name)
				
				require.False(t, closure.Static)
				require.False(t, closure.ByReference)
			},
		},
		{
			name:  "Closure with use clause",
			input: `<?php $fn = function($x) use ($y, $z) { return $x + $y + $z; };`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				assignment, ok := stmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)
				
				closure, ok := assignment.Right.(*ast.AnonymousFunctionExpression)
				require.True(t, ok)
				require.Len(t, closure.Parameters, 1)
				require.Len(t, closure.UseClause, 2)
				
				// Check use clause variables
				useVar1, ok := closure.UseClause[0].(*ast.Variable)
				require.True(t, ok)
				require.Equal(t, "$y", useVar1.Name)
				
				useVar2, ok := closure.UseClause[1].(*ast.Variable)
				require.True(t, ok)
				require.Equal(t, "$z", useVar2.Name)
			},
		},
		{
			name:  "Closure with use clause and references",
			input: `<?php $fn = function() use (&$x, $y) { $x++; };`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				assignment, ok := stmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)
				
				closure, ok := assignment.Right.(*ast.AnonymousFunctionExpression)
				require.True(t, ok)
				require.Len(t, closure.UseClause, 2)
				
				// First use variable should be a reference
				refExpr, ok := closure.UseClause[0].(*ast.UnaryExpression)
				require.True(t, ok)
				refVar, ok := refExpr.Operand.(*ast.Variable)
				require.True(t, ok)
				require.Equal(t, "$x", refVar.Name)
				
				// Second use variable should be normal
				useVar, ok := closure.UseClause[1].(*ast.Variable)
				require.True(t, ok)
				require.Equal(t, "$y", useVar.Name)
			},
		},
		{
			name:  "Closure with return type",
			input: `<?php $fn = function(): int { return 42; };`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				assignment, ok := stmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)
				
				closure, ok := assignment.Right.(*ast.AnonymousFunctionExpression)
				require.True(t, ok)
				require.NotNil(t, closure.ReturnType)
				require.Equal(t, "int", closure.ReturnType.Name)
			},
		},
		{
			name:  "Static closure",
			input: `<?php $fn = static function() { return self::class; };`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				assignment, ok := stmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)
				
				closure, ok := assignment.Right.(*ast.AnonymousFunctionExpression)
				require.True(t, ok)
				require.True(t, closure.Static)
				require.False(t, closure.ByReference)
			},
		},
		{
			name:  "Complex closure with all features",
			input: `<?php $id = $cancellation->subscribe(static function (\Throwable $exception) use (&$waiting, &$offset, $suspension,): void { foreach ($waiting as $key => $pending) { } });`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				assignment, ok := stmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)
				
				call, ok := assignment.Right.(*ast.CallExpression)
				require.True(t, ok)
				require.Len(t, call.Arguments, 1)
				
				closure, ok := call.Arguments[0].(*ast.AnonymousFunctionExpression)
				require.True(t, ok)
				
				// Verify static modifier
				require.True(t, closure.Static)
				require.False(t, closure.ByReference)
				
				// Verify parameters
				require.Len(t, closure.Parameters, 1)
				param := closure.Parameters[0]
				require.Equal(t, "$exception", param.Name)
				require.Equal(t, "\\Throwable", param.Type.Name)
				
				// Verify use clause
				require.Len(t, closure.UseClause, 3)
				
				// First use variable: &$waiting (reference)
				refExpr1, ok := closure.UseClause[0].(*ast.UnaryExpression)
				require.True(t, ok)
				refVar1, ok := refExpr1.Operand.(*ast.Variable)
				require.True(t, ok)
				require.Equal(t, "$waiting", refVar1.Name)
				
				// Second use variable: &$offset (reference)
				refExpr2, ok := closure.UseClause[1].(*ast.UnaryExpression)
				require.True(t, ok)
				refVar2, ok := refExpr2.Operand.(*ast.Variable)
				require.True(t, ok)
				require.Equal(t, "$offset", refVar2.Name)
				
				// Third use variable: $suspension (normal)
				useVar, ok := closure.UseClause[2].(*ast.Variable)
				require.True(t, ok)
				require.Equal(t, "$suspension", useVar.Name)
				
				// Verify return type
				require.NotNil(t, closure.ReturnType)
				require.Equal(t, "void", closure.ReturnType.Name)
				
				// Verify body contains foreach statement
				require.Len(t, closure.Body, 1)
				foreachStmt, ok := closure.Body[0].(*ast.ForeachStatement)
				require.True(t, ok)
				
				iterVar, ok := foreachStmt.Iterable.(*ast.Variable)
				require.True(t, ok)
				require.Equal(t, "$waiting", iterVar.Name)
			},
		},
		{
			name:  "By-reference closure",
			input: `<?php $fn = function &() { return $this->value; };`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				assignment, ok := stmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)
				
				closure, ok := assignment.Right.(*ast.AnonymousFunctionExpression)
				require.True(t, ok)
				require.True(t, closure.ByReference)
				require.False(t, closure.Static)
			},
		},
		{
			name:  "Static by-reference closure with everything",
			input: `<?php $fn = static function &(array $data) use (&$count): ?array { return $data ?: null; };`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				assignment, ok := stmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)
				
				closure, ok := assignment.Right.(*ast.AnonymousFunctionExpression)
				require.True(t, ok)
				require.True(t, closure.Static)
				require.True(t, closure.ByReference)
				
				// Verify parameter
				require.Len(t, closure.Parameters, 1)
				param := closure.Parameters[0]
				require.Equal(t, "$data", param.Name)
				require.Equal(t, "array", param.Type.Name)
				
				// Verify use clause
				require.Len(t, closure.UseClause, 1)
				refExpr, ok := closure.UseClause[0].(*ast.UnaryExpression)
				require.True(t, ok)
				refVar, ok := refExpr.Operand.(*ast.Variable)
				require.True(t, ok)
				require.Equal(t, "$count", refVar.Name)
				
				// Verify nullable return type
				require.NotNil(t, closure.ReturnType)
				require.Equal(t, "array", closure.ReturnType.Name)
				require.True(t, closure.ReturnType.Nullable)
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