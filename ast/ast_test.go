package ast

import (
	"testing"

	"github.com/wudi/php-parser/lexer"
)

// TestASTKindConstants 测试AST Kind常量值
func TestASTKindConstants(t *testing.T) {
	tests := []struct {
		kind     ASTKind
		expected string
	}{
		{ASTZval, "ZVAL"},
		{ASTVar, "VAR"},
		{ASTBinaryOp, "BINARY_OP"},
		{ASTAssign, "ASSIGN"},
		{ASTEcho, "ECHO"},
		{ASTReturn, "RETURN"},
		{ASTIf, "IF"},
		{ASTWhile, "WHILE"},
		{ASTFor, "FOR"},
		{ASTFuncDecl, "FUNC_DECL"},
		{ASTArray, "ARRAY"},
		{ASTCall, "CALL"},
		{ASTStmtList, "STMT_LIST"},
	}

	for _, test := range tests {
		if test.kind.String() != test.expected {
			t.Errorf("Expected %s, got %s for kind %d", test.expected, test.kind.String(), test.kind)
		}
	}
}

// TestASTKindProperties 测试AST Kind属性检查方法
func TestASTKindProperties(t *testing.T) {
	// 测试特殊节点
	if !ASTZval.IsSpecial() {
		t.Error("ASTZval should be special")
	}
	if !ASTFuncDecl.IsDecl() {
		t.Error("ASTFuncDecl should be declaration")
	}

	// 测试列表节点
	if !ASTArray.IsList() {
		t.Error("ASTArray should be list")
	}
	if !ASTStmtList.IsList() {
		t.Error("ASTStmtList should be list")
	}

	// 测试子节点数量
	if ASTVar.GetNumChildren() != 1 {
		t.Errorf("ASTVar should have 1 child, got %d", ASTVar.GetNumChildren())
	}
	if ASTBinaryOp.GetNumChildren() != 2 {
		t.Errorf("ASTBinaryOp should have 2 children, got %d", ASTBinaryOp.GetNumChildren())
	}
}

// TestNodeCreation 测试节点创建
func TestNodeCreation(t *testing.T) {
	pos := lexer.Position{Line: 1, Column: 1, Offset: 0}

	// 测试变量节点
	variable := NewVariable(pos, "$test")
	if variable.GetKind() != ASTVar {
		t.Errorf("Expected ASTVar, got %s", variable.GetKind().String())
	}
	if variable.Name != "$test" {
		t.Errorf("Expected $test, got %s", variable.Name)
	}
	if variable.GetLineNo() != 1 {
		t.Errorf("Expected line 1, got %d", variable.GetLineNo())
	}

	// 测试字符串字面量
	str := NewStringLiteral(pos, "hello", "\"hello\"")
	if str.GetKind() != ASTZval {
		t.Errorf("Expected ASTZval, got %s", str.GetKind().String())
	}
	if str.Value != "hello" {
		t.Errorf("Expected hello, got %s", str.Value)
	}

	// 测试二元表达式
	left := NewVariable(pos, "$a")
	right := NewVariable(pos, "$b")
	binExpr := NewBinaryExpression(pos, left, "+", right)
	if binExpr.GetKind() != ASTBinaryOp {
		t.Errorf("Expected ASTBinaryOp, got %s", binExpr.GetKind().String())
	}
	if binExpr.Operator != "+" {
		t.Errorf("Expected +, got %s", binExpr.Operator)
	}

	// 测试子节点
	children := binExpr.GetChildren()
	if len(children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(children))
	}
}

// TestASTBuilder 测试AST构建器
func TestASTBuilder(t *testing.T) {
	pos := lexer.Position{Line: 1, Column: 1, Offset: 0}
	builder := NewASTBuilder()

	// 测试创建变量
	variable := builder.CreateVar(pos, "$test")
	if variable.GetKind() != ASTVar {
		t.Errorf("Expected ASTVar, got %s", variable.GetKind().String())
	}

	// 测试创建字面量
	str := builder.CreateZval(pos, "hello")
	if str.GetKind() != ASTZval {
		t.Errorf("Expected ASTZval, got %s", str.GetKind().String())
	}

	// 测试创建二元操作
	binOp := builder.CreateBinaryOp(pos, variable, str, "+")
	if binOp == nil {
		t.Error("Binary operation should not be nil")
	}
	if binOp.GetKind() != ASTBinaryOp {
		t.Errorf("Expected ASTBinaryOp, got %s", binOp.GetKind().String())
	}

	// 测试创建数组
	elements := []Node{variable, str}
	array := builder.CreateArray(pos, elements)
	if array.GetKind() != ASTArray {
		t.Errorf("Expected ASTArray, got %s", array.GetKind().String())
	}
}

// TestVisitorPattern 测试访问者模式
func TestVisitorPattern(t *testing.T) {
	pos := lexer.Position{Line: 1, Column: 1, Offset: 0}

	// 创建一个简单的AST
	variable := NewVariable(pos, "$test")
	str := NewStringLiteral(pos, "hello", "\"hello\"")
	binExpr := NewBinaryExpression(pos, variable, "+", str)
	echo := NewEchoStatement(pos)
	echo.Arguments = NewArgumentList(pos, []Expression{binExpr})

	program := NewProgram(pos)
	program.Body = append(program.Body, echo)

	// 测试访问者模式 - 计数节点
	nodeCount := 0
	Walk(VisitorFunc(func(node Node) bool {
		nodeCount++
		return true
	}), program)

	expectedCount := 6 // Program, EchoStatement, ArgumentList, BinaryExpression, Variable, StringLiteral
	if nodeCount != expectedCount {
		t.Errorf("Expected %d nodes, got %d", expectedCount, nodeCount)
	}

	// 测试查找特定节点类型
	variables := FindAllFunc(program, func(node Node) bool {
		return node.GetKind() == ASTVar
	})
	if len(variables) != 1 {
		t.Errorf("Expected 1 variable, got %d", len(variables))
	}

	// 测试查找第一个节点
	firstVar := FindFirstFunc(program, func(node Node) bool {
		return node.GetKind() == ASTVar
	})
	if firstVar == nil {
		t.Error("Should find first variable")
	}
	if v, ok := firstVar.(*Variable); !ok || v.Name != "$test" {
		t.Error("First variable should be $test")
	}

	// 测试计数
	varCount := CountFunc(program, func(node Node) bool {
		return node.GetKind() == ASTVar
	})
	if varCount != 1 {
		t.Errorf("Expected 1 variable count, got %d", varCount)
	}
}

// TestTransform 测试AST转换
func TestTransform(t *testing.T) {
	pos := lexer.Position{Line: 1, Column: 1, Offset: 0}

	// 创建一个变量节点
	variable := NewVariable(pos, "$old_name")

	// 使用转换器重命名变量
	transformed := TransformFunc(variable, func(node Node) Node {
		if v, ok := node.(*Variable); ok && v.Name == "$old_name" {
			return NewVariable(pos, "$new_name")
		}
		return node
	})

	// 检查转换结果
	if v, ok := transformed.(*Variable); !ok {
		t.Error("Transformed node should be Variable")
	} else if v.Name != "$new_name" {
		t.Errorf("Expected $new_name, got %s", v.Name)
	}
}

// TestNodeAttributes 测试节点属性
func TestNodeAttributes(t *testing.T) {
	pos := lexer.Position{Line: 5, Column: 10, Offset: 50}
	variable := NewVariable(pos, "$test")

	// 测试基本属性
	if variable.GetLineNo() != 5 {
		t.Errorf("Expected line 5, got %d", variable.GetLineNo())
	}

	// 测试位置
	position := variable.GetPosition()
	if position.Line != 5 || position.Column != 10 {
		t.Errorf("Expected line 5 column 10, got line %d column %d", position.Line, position.Column)
	}

	// 测试属性字典
	attrs := variable.GetAttributes()
	attrs["custom"] = "value"
	if attrs["custom"] != "value" {
		t.Error("Should be able to set custom attributes")
	}
}

// TestComplexAST 测试复杂AST结构
func TestComplexAST(t *testing.T) {
	pos := lexer.Position{Line: 1, Column: 1, Offset: 0}
	builder := NewASTBuilder()

	// 构建 if ($x > 0) { echo $x; }
	variable := builder.CreateVar(pos, "$x")
	zero := builder.CreateZval(pos, 0)
	condition := builder.CreateBinaryOp(pos, variable, zero, ">")

	echoVar := builder.CreateVar(pos, "$x")
	echoStmt := builder.CreateEcho(pos, []Node{echoVar})

	ifStmt := builder.CreateIf(pos, condition, []Node{echoStmt}, nil)

	// 验证结构
	if ifStmt.GetKind() != ASTIf {
		t.Error("Should create IF statement")
	}

	// 验证访问者能正确遍历
	nodeTypes := make(map[ASTKind]int)
	Walk(VisitorFunc(func(node Node) bool {
		nodeTypes[node.GetKind()]++
		return true
	}), ifStmt)

	expectedTypes := map[ASTKind]int{
		ASTIf:       1,
		ASTBinaryOp: 1,
		ASTVar:      2, // $x 出现两次
		ASTZval:     1, // 数字 0
		ASTEcho:     1,
	}

	for kind, expected := range expectedTypes {
		if nodeTypes[kind] != expected {
			t.Errorf("Expected %d %s nodes, got %d", expected, kind.String(), nodeTypes[kind])
		}
	}
}

// BenchmarkNodeCreation 基准测试节点创建
func BenchmarkNodeCreation(b *testing.B) {
	pos := lexer.Position{Line: 1, Column: 1, Offset: 0}

	b.Run("Variable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			NewVariable(pos, "$test")
		}
	})

	b.Run("BinaryExpression", func(b *testing.B) {
		left := NewVariable(pos, "$a")
		right := NewVariable(pos, "$b")
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			NewBinaryExpression(pos, left, "+", right)
		}
	})
}

// BenchmarkWalk 基准测试AST遍历
func BenchmarkWalk(b *testing.B) {
	pos := lexer.Position{Line: 1, Column: 1, Offset: 0}
	builder := NewASTBuilder()

	// 创建一个复杂的AST
	program := NewProgram(pos)
	for i := 0; i < 100; i++ {
		variable := builder.CreateVar(pos, "$test")
		str := builder.CreateZval(pos, "hello")
		binExpr := builder.CreateBinaryOp(pos, variable, str, "+")
		echo := builder.CreateEcho(pos, []Node{binExpr})
		program.Body = append(program.Body, echo.(Statement))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count := 0
		Walk(VisitorFunc(func(node Node) bool {
			count++
			return true
		}), program)
	}
}
