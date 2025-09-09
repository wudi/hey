package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wudi/php-parser/lexer"
)

func TestProgram_String(t *testing.T) {
	pos := lexer.Position{Line: 1, Column: 0, Offset: 0}
	program := NewProgram(pos)

	// 添加一个 echo 语句
	echo := NewEchoStatement(pos)
	echo.Arguments = NewArgumentList(pos, []Expression{NewStringLiteral(pos, "Hello, World!", `"Hello, World!"`)})
	program.Body = append(program.Body, echo)

	expected := `echo "Hello, World!";` + "\n"
	assert.Equal(t, expected, program.String())
}

func TestEchoStatement_String(t *testing.T) {
	pos := lexer.Position{Line: 1, Column: 0, Offset: 0}
	echo := NewEchoStatement(pos)

	// 测试单个参数
	echo.Arguments = NewArgumentList(pos, []Expression{NewStringLiteral(pos, "Hello", `"Hello"`)})
	assert.Equal(t, `echo "Hello";`, echo.String())

	// 测试多个参数
	echo.Arguments = NewArgumentList(pos, []Expression{NewStringLiteral(pos, "Hello", `"Hello"`), NewVariable(pos, "$name")})
	assert.Equal(t, `echo "Hello", $name;`, echo.String())
}

func TestAssignmentExpression_String(t *testing.T) {
	pos := lexer.Position{Line: 1, Column: 0, Offset: 0}

	variable := NewVariable(pos, "$name")
	value := NewStringLiteral(pos, "John", `"John"`)
	assignment := NewAssignmentExpression(pos, variable, "=", value)

	expected := `$name = "John"`
	assert.Equal(t, expected, assignment.String())
}

func TestBinaryExpression_String(t *testing.T) {
	pos := lexer.Position{Line: 1, Column: 0, Offset: 0}

	left := NewNumberLiteral(pos, "5", IntegerKind)
	right := NewNumberLiteral(pos, "3", IntegerKind)
	expr := NewBinaryExpression(pos, left, "+", right)

	expected := "(5 + 3)"
	assert.Equal(t, expected, expr.String())
}

func TestUnaryExpression_String(t *testing.T) {
	pos := lexer.Position{Line: 1, Column: 0, Offset: 0}

	// 前缀递增
	variable := NewVariable(pos, "$i")
	prefixExpr := NewUnaryExpression(pos, "++", variable, true)
	assert.Equal(t, "++$i", prefixExpr.String())

	// 后缀递增
	postfixExpr := NewUnaryExpression(pos, "++", variable, false)
	assert.Equal(t, "$i++", postfixExpr.String())
}

func TestVariable_String(t *testing.T) {
	pos := lexer.Position{Line: 1, Column: 0, Offset: 0}
	variable := NewVariable(pos, "$name")

	assert.Equal(t, "$name", variable.String())
	assert.Equal(t, ASTVar, variable.GetKind())
}

func TestStringLiteral_String(t *testing.T) {
	pos := lexer.Position{Line: 1, Column: 0, Offset: 0}
	str := NewStringLiteral(pos, "Hello, World!", `"Hello, World!"`)

	assert.Equal(t, `"Hello, World!"`, str.String())
	assert.Equal(t, ASTZval, str.GetKind())
}

func TestNumberLiteral_String(t *testing.T) {
	pos := lexer.Position{Line: 1, Column: 0, Offset: 0}

	// 整数
	intNum := NewNumberLiteral(pos, "123", IntegerKind)
	assert.Equal(t, "123", intNum.String())
	assert.Equal(t, ASTZval, intNum.GetKind())

	// 浮点数
	floatNum := NewNumberLiteral(pos, "3.14", FloatKind)
	assert.Equal(t, "3.14", floatNum.String())
}

func TestBooleanLiteral_String(t *testing.T) {
	pos := lexer.Position{Line: 1, Column: 0, Offset: 0}

	trueLiteral := NewBooleanLiteral(pos, true)
	assert.Equal(t, "true", trueLiteral.String())

	falseLiteral := NewBooleanLiteral(pos, false)
	assert.Equal(t, "false", falseLiteral.String())
}

func TestNullLiteral_String(t *testing.T) {
	pos := lexer.Position{Line: 1, Column: 0, Offset: 0}
	nullLit := NewNullLiteral(pos)

	assert.Equal(t, "null", nullLit.String())
	assert.Equal(t, ASTZval, nullLit.GetKind())
}

func TestArrayExpression_String(t *testing.T) {
	pos := lexer.Position{Line: 1, Column: 0, Offset: 0}
	arr := NewArrayExpression(pos)

	// 空数组
	assert.Equal(t, "[]", arr.String())

	// 有元素的数组
	arr.Elements = append(arr.Elements, NewNumberLiteral(pos, "1", IntegerKind))
	arr.Elements = append(arr.Elements, NewNumberLiteral(pos, "2", IntegerKind))
	arr.Elements = append(arr.Elements, NewNumberLiteral(pos, "3", IntegerKind))

	assert.Equal(t, "[1, 2, 3]", arr.String())
}

func TestIfStatement_String(t *testing.T) {
	pos := lexer.Position{Line: 1, Column: 0, Offset: 0}

	// 创建条件表达式
	variable := NewVariable(pos, "$x")
	value := NewNumberLiteral(pos, "5", IntegerKind)
	condition := NewBinaryExpression(pos, variable, ">", value)

	ifStmt := NewIfStatement(pos, condition)

	// 添加 consequent 语句
	echo := NewEchoStatement(pos)
	echo.Arguments = NewArgumentList(pos, []Expression{NewStringLiteral(pos, "x is greater than 5", `"x is greater than 5"`)})
	ifStmt.Consequent = append(ifStmt.Consequent, echo)

	expected := `if (($x > 5)) {
  echo "x is greater than 5";
}`
	assert.Equal(t, expected, ifStmt.String())
}

func TestWhileStatement_String(t *testing.T) {
	pos := lexer.Position{Line: 1, Column: 0, Offset: 0}

	// 创建条件表达式
	variable := NewVariable(pos, "$i")
	value := NewNumberLiteral(pos, "10", IntegerKind)
	condition := NewBinaryExpression(pos, variable, "<", value)

	whileStmt := NewWhileStatement(pos, condition)

	// 添加循环体
	increment := NewUnaryExpression(pos, "++", variable, false)
	exprStmt := NewExpressionStatement(pos, increment)
	whileStmt.Body = append(whileStmt.Body, exprStmt)

	expected := `while (($i < 10)) {
  $i++;
}`
	assert.Equal(t, expected, whileStmt.String())
}

func TestFunctionDeclaration_String(t *testing.T) {
	pos := lexer.Position{Line: 1, Column: 0, Offset: 0}

	funcName := NewIdentifierNode(pos, "greet")
	funcDecl := NewFunctionDeclaration(pos, funcName)

	// 添加参数
	paramName := NewIdentifierNode(pos, "$name")
	param := NewParameterNode(pos, paramName)
	funcDecl.Parameters = NewParameterList(pos, []*ParameterNode{param})

	// 添加函数体
	echo := NewEchoStatement(pos)
	echo.Arguments = NewArgumentList(pos, []Expression{NewStringLiteral(pos, "Hello, ", `"Hello, "`), NewVariable(pos, "$name")})
	funcDecl.Body = append(funcDecl.Body, echo)

	expected := `function greet($name) {
  echo "Hello, ", $name;
}`
	assert.Equal(t, expected, funcDecl.String())
}

func TestReturnStatement_String(t *testing.T) {
	pos := lexer.Position{Line: 1, Column: 0, Offset: 0}

	// 有返回值的 return
	value := NewNumberLiteral(pos, "42", IntegerKind)
	returnStmt := NewReturnStatement(pos, value)
	assert.Equal(t, "return 42;", returnStmt.String())

	// 无返回值的 return
	emptyReturn := NewReturnStatement(pos, nil)
	assert.Equal(t, "return;", emptyReturn.String())
}

func TestBreakContinueStatement_String(t *testing.T) {
	pos := lexer.Position{Line: 1, Column: 0, Offset: 0}

	breakStmt := NewBreakStatement(pos)
	assert.Equal(t, "break;", breakStmt.String())

	continueStmt := NewContinueStatement(pos)
	assert.Equal(t, "continue;", continueStmt.String())
}

func TestNodeJSON(t *testing.T) {
	pos := lexer.Position{Line: 1, Column: 0, Offset: 0}
	variable := NewVariable(pos, "$name")

	json, err := variable.ToJSON()
	assert.NoError(t, err)
	assert.Contains(t, string(json), `"kind": 256`) // ASTVar = 256
	assert.Contains(t, string(json), `"name": "$name"`)
}
