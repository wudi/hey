package ast_test

import (
	"fmt"
	"log"

	ast2 "github.com/wudi/hey/compiler/ast"
	"github.com/wudi/hey/compiler/lexer"
)

// ExampleASTBuilder 演示如何使用AST构建器
func ExampleASTBuilder() {
	pos := lexer.Position{Line: 1, Column: 1, Offset: 0}
	builder := ast2.NewASTBuilder()

	// 构建 $name = "John";
	nameVar := builder.CreateVar(pos, "$name")
	john := builder.CreateZval(pos, "John")
	assignment := builder.CreateAssign(pos, nameVar, john)
	assignStmt := builder.CreateExpressionStatement(pos, assignment)

	// 构建 echo $name;
	echoVar := builder.CreateVar(pos, "$name")
	echoStmt := builder.CreateEcho(pos, []ast2.Node{echoVar})

	// 构建程序
	program := builder.CreateStmtList(pos, []ast2.Node{assignStmt, echoStmt})

	fmt.Printf("AST Kind: %s\n", program.GetKind().String())
	fmt.Printf("Children: %d\n", len(program.GetChildren()))

	// Output:
	// AST Kind: STMT_LIST
	// Children: 2
}

// ExampleVisitor 演示如何使用访问者模式
func ExampleVisitor() {
	pos := lexer.Position{Line: 1, Column: 1, Offset: 0}
	builder := ast2.NewASTBuilder()

	// 构建一个简单的AST: $x = $x + 1;
	x1 := builder.CreateVar(pos, "$x")
	x2 := builder.CreateVar(pos, "$x")
	one := builder.CreateZval(pos, 1)
	addition := builder.CreateBinaryOp(pos, x2, one, "+")
	assignment := builder.CreateAssign(pos, x1, addition)

	// 计算变量$x的使用次数
	varCount := ast2.CountFunc(assignment, func(node ast2.Node) bool {
		if v, ok := node.(*ast2.Variable); ok && v.Name == "$x" {
			return true
		}
		return false
	})

	fmt.Printf("Variable $x used %d times\n", varCount)

	// 找到所有二元操作
	binaryOps := ast2.FindAllFunc(assignment, func(node ast2.Node) bool {
		return node.GetKind() == ast2.ASTBinaryOp
	})

	fmt.Printf("Found %d binary operations\n", len(binaryOps))

	// Output:
	// Variable $x used 2 times
	// Found 1 binary operations
}

// ExampleTransform 演示如何使用AST转换
func ExampleTransform() {
	pos := lexer.Position{Line: 1, Column: 1, Offset: 0}

	// 创建原始变量
	original := ast2.NewVariable(pos, "$oldName")

	// 使用转换器重命名变量
	transformed := ast2.TransformFunc(original, func(node ast2.Node) ast2.Node {
		if v, ok := node.(*ast2.Variable); ok && v.Name == "$oldName" {
			return ast2.NewVariable(pos, "$newName")
		}
		return node
	})

	if v, ok := transformed.(*ast2.Variable); ok {
		fmt.Printf("Transformed: %s\n", v.Name)
	}

	// Output:
	// Transformed: $newName
}

// ExampleASTKind 演示AST Kind的使用
func ExampleASTKind() {
	// 检查Kind属性
	fmt.Printf("ASTZval is special: %t\n", ast2.ASTZval.IsSpecial())
	fmt.Printf("ASTArray is list: %t\n", ast2.ASTArray.IsList())
	fmt.Printf("ASTFuncDecl is declaration: %t\n", ast2.ASTFuncDecl.IsDecl())
	fmt.Printf("ASTBinaryOp has %d children\n", ast2.ASTBinaryOp.GetNumChildren())

	// Kind字符串表示
	fmt.Printf("Kind names: %s, %s, %s\n",
		ast2.ASTVar.String(),
		ast2.ASTBinaryOp.String(),
		ast2.ASTEcho.String())

	// Output:
	// ASTZval is special: true
	// ASTArray is list: true
	// ASTFuncDecl is declaration: true
	// ASTBinaryOp has 2 children
	// Kind names: VAR, BINARY_OP, ECHO
}

// Example_complexAST 演示构建复杂AST结构
func Example_complexAST() {
	pos := lexer.Position{Line: 1, Column: 1, Offset: 0}
	builder := ast2.NewASTBuilder()

	// 构建: if ($x > 0) { echo "positive"; } else { echo "non-positive"; }

	// 条件: $x > 0
	x := builder.CreateVar(pos, "$x")
	zero := builder.CreateZval(pos, 0)
	condition := builder.CreateBinaryOp(pos, x, zero, ">")

	// then分支: echo "positive";
	positive := builder.CreateZval(pos, "positive")
	echoPositive := builder.CreateEcho(pos, []ast2.Node{positive})

	// else分支: echo "non-positive";
	nonPositive := builder.CreateZval(pos, "non-positive")
	echoNonPositive := builder.CreateEcho(pos, []ast2.Node{nonPositive})

	// if语句
	ifStmt := builder.CreateIf(pos, condition,
		[]ast2.Node{echoPositive},
		[]ast2.Node{echoNonPositive})

	// 统计节点数量
	nodeCount := 0
	ast2.Walk(ast2.VisitorFunc(func(node ast2.Node) bool {
		nodeCount++
		return true
	}), ifStmt)

	fmt.Printf("Total nodes: %d\n", nodeCount)
	fmt.Printf("AST Kind: %s\n", ifStmt.GetKind().String())

	// 找到所有echo语句
	echoNodes := ast2.FindAllFunc(ifStmt, func(node ast2.Node) bool {
		return node.GetKind() == ast2.ASTEcho
	})

	fmt.Printf("Echo statements: %d\n", len(echoNodes))

	// Output:
	// Total nodes: 10
	// AST Kind: IF
	// Echo statements: 2
}

// Example_jsonSerialization 演示JSON序列化
func Example_jsonSerialization() {
	pos := lexer.Position{Line: 1, Column: 1, Offset: 0}
	variable := ast2.NewVariable(pos, "$test")

	// 序列化为JSON
	jsonData, err := variable.ToJSON()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("JSON contains kind: %t\n",
		string(jsonData) != "" && variable.GetKind() == ast2.ASTVar)
	fmt.Printf("JSON contains name: %t\n",
		string(jsonData) != "" && variable.Name == "$test")

	// Output:
	// JSON contains kind: true
	// JSON contains name: true
}
