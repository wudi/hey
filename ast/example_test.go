package ast_test

import (
	"fmt"
	"log"

	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/lexer"
)

// ExampleASTBuilder 演示如何使用AST构建器
func ExampleASTBuilder() {
	pos := lexer.Position{Line: 1, Column: 1, Offset: 0}
	builder := ast.NewASTBuilder()

	// 构建 $name = "John";
	nameVar := builder.CreateVar(pos, "$name")
	john := builder.CreateZval(pos, "John")
	assignment := builder.CreateAssign(pos, nameVar, john)
	assignStmt := builder.CreateExpressionStatement(pos, assignment)

	// 构建 echo $name;
	echoVar := builder.CreateVar(pos, "$name")
	echoStmt := builder.CreateEcho(pos, []ast.Node{echoVar})

	// 构建程序
	program := builder.CreateStmtList(pos, []ast.Node{assignStmt, echoStmt})

	fmt.Printf("AST Kind: %s\n", program.GetKind().String())
	fmt.Printf("Children: %d\n", len(program.GetChildren()))
	
	// Output:
	// AST Kind: STMT_LIST
	// Children: 2
}

// ExampleVisitor 演示如何使用访问者模式
func ExampleVisitor() {
	pos := lexer.Position{Line: 1, Column: 1, Offset: 0}
	builder := ast.NewASTBuilder()

	// 构建一个简单的AST: $x = $x + 1;
	x1 := builder.CreateVar(pos, "$x")
	x2 := builder.CreateVar(pos, "$x")
	one := builder.CreateZval(pos, 1)
	addition := builder.CreateBinaryOp(pos, x2, one, "+")
	assignment := builder.CreateAssign(pos, x1, addition)

	// 计算变量$x的使用次数
	varCount := ast.CountFunc(assignment, func(node ast.Node) bool {
		if v, ok := node.(*ast.Variable); ok && v.Name == "$x" {
			return true
		}
		return false
	})

	fmt.Printf("Variable $x used %d times\n", varCount)

	// 找到所有二元操作
	binaryOps := ast.FindAllFunc(assignment, func(node ast.Node) bool {
		return node.GetKind() == ast.ASTBinaryOp
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
	original := ast.NewVariable(pos, "$oldName")

	// 使用转换器重命名变量
	transformed := ast.TransformFunc(original, func(node ast.Node) ast.Node {
		if v, ok := node.(*ast.Variable); ok && v.Name == "$oldName" {
			return ast.NewVariable(pos, "$newName")
		}
		return node
	})

	if v, ok := transformed.(*ast.Variable); ok {
		fmt.Printf("Transformed: %s\n", v.Name)
	}

	// Output:
	// Transformed: $newName
}

// ExampleASTKind 演示AST Kind的使用
func ExampleASTKind() {
	// 检查Kind属性
	fmt.Printf("ASTZval is special: %t\n", ast.ASTZval.IsSpecial())
	fmt.Printf("ASTArray is list: %t\n", ast.ASTArray.IsList())
	fmt.Printf("ASTFuncDecl is declaration: %t\n", ast.ASTFuncDecl.IsDecl())
	fmt.Printf("ASTBinaryOp has %d children\n", ast.ASTBinaryOp.GetNumChildren())

	// Kind字符串表示
	fmt.Printf("Kind names: %s, %s, %s\n", 
		ast.ASTVar.String(), 
		ast.ASTBinaryOp.String(), 
		ast.ASTEcho.String())

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
	builder := ast.NewASTBuilder()

	// 构建: if ($x > 0) { echo "positive"; } else { echo "non-positive"; }
	
	// 条件: $x > 0
	x := builder.CreateVar(pos, "$x")
	zero := builder.CreateZval(pos, 0)
	condition := builder.CreateBinaryOp(pos, x, zero, ">")

	// then分支: echo "positive";
	positive := builder.CreateZval(pos, "positive")
	echoPositive := builder.CreateEcho(pos, []ast.Node{positive})

	// else分支: echo "non-positive";
	nonPositive := builder.CreateZval(pos, "non-positive")
	echoNonPositive := builder.CreateEcho(pos, []ast.Node{nonPositive})

	// if语句
	ifStmt := builder.CreateIf(pos, condition, 
		[]ast.Node{echoPositive}, 
		[]ast.Node{echoNonPositive})

	// 统计节点数量
	nodeCount := 0
	ast.Walk(ast.VisitorFunc(func(node ast.Node) bool {
		nodeCount++
		return true
	}), ifStmt)

	fmt.Printf("Total nodes: %d\n", nodeCount)
	fmt.Printf("AST Kind: %s\n", ifStmt.GetKind().String())

	// 找到所有echo语句
	echoNodes := ast.FindAllFunc(ifStmt, func(node ast.Node) bool {
		return node.GetKind() == ast.ASTEcho
	})

	fmt.Printf("Echo statements: %d\n", len(echoNodes))

	// Output:
	// Total nodes: 8
	// AST Kind: IF
	// Echo statements: 2
}

// Example_jsonSerialization 演示JSON序列化
func Example_jsonSerialization() {
	pos := lexer.Position{Line: 1, Column: 1, Offset: 0}
	variable := ast.NewVariable(pos, "$test")

	// 序列化为JSON
	jsonData, err := variable.ToJSON()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("JSON contains kind: %t\n", 
		string(jsonData) != "" && variable.GetKind() == ast.ASTVar)
	fmt.Printf("JSON contains name: %t\n", 
		string(jsonData) != "" && variable.Name == "$test")

	// Output:
	// JSON contains kind: true
	// JSON contains name: true
}