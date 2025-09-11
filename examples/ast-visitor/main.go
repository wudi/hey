//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"strings"

	ast2 "github.com/wudi/hey/compiler/ast"
	"github.com/wudi/hey/compiler/lexer"
	"github.com/wudi/hey/compiler/parser"
)

// VariableCollector 收集所有变量名
type VariableCollector struct {
	Variables []string
}

func NewVariableCollector() *VariableCollector {
	return &VariableCollector{
		Variables: make([]string, 0),
	}
}

func (vc *VariableCollector) Visit(node ast2.Node) bool {
	if variable, ok := node.(*ast2.Variable); ok {
		vc.Variables = append(vc.Variables, variable.Name)
	}
	return true // 继续遍历子节点
}

// FunctionCollector 收集所有函数声明
type FunctionCollector struct {
	Functions []*ast2.FunctionDeclaration
}

func NewFunctionCollector() *FunctionCollector {
	return &FunctionCollector{
		Functions: make([]*ast2.FunctionDeclaration, 0),
	}
}

func (fc *FunctionCollector) Visit(node ast2.Node) bool {
	if funcDecl, ok := node.(*ast2.FunctionDeclaration); ok {
		fc.Functions = append(fc.Functions, funcDecl)
	}
	return true
}

// DepthCounter 计算AST的最大深度
type DepthCounter struct {
	MaxDepth     int
	currentDepth int
}

func NewDepthCounter() *DepthCounter {
	return &DepthCounter{}
}

func (dc *DepthCounter) Visit(node ast2.Node) bool {
	dc.currentDepth++
	if dc.currentDepth > dc.MaxDepth {
		dc.MaxDepth = dc.currentDepth
	}

	// 遍历子节点后需要减少深度
	defer func() {
		dc.currentDepth--
	}()

	return true
}

func main() {
	// Sample PHP code with various constructs
	phpCode := `<?php
$username = "john";
$password = "secret123";
$isLoggedIn = false;

function authenticate($user, $pass) {
    global $username, $password;
    if ($user === $username && $pass === $password) {
        return true;
    }
    return false;
}

function getUserInfo($userId) {
    $userData = array(
        'id' => $userId,
        'name' => 'John Doe',
        'email' => 'john@example.com'
    );
    return $userData;
}

class User {
    private $id;
    private $name;
    
    public function __construct($id, $name) {
        $this->id = $id;
        $this->name = $name;
    }
    
    public function getName() {
        return $this->name;
    }
}

$user = new User(1, "John");
echo $user->getName();
?>`

	// Parse the code
	l := lexer.New(phpCode)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Println("Parsing errors:")
		for _, err := range p.Errors() {
			fmt.Printf("  %s\n", err)
		}
		return
	}

	fmt.Println("=== AST Visitor Examples ===\n")

	// Example 1: Collect all variables using custom visitor
	fmt.Println("1. Collecting all variables:")
	variableCollector := NewVariableCollector()
	ast2.Walk(variableCollector, program)

	// Remove duplicates
	uniqueVars := removeDuplicates(variableCollector.Variables)
	fmt.Printf("   Found %d unique variables: %s\n\n", len(uniqueVars), strings.Join(uniqueVars, ", "))

	// Example 2: Collect all function declarations
	fmt.Println("2. Collecting all function declarations:")
	functionCollector := NewFunctionCollector()
	ast2.Walk(functionCollector, program)

	for i, fn := range functionCollector.Functions {
		funcName := ""
		if identifier, ok := fn.Name.(*ast2.IdentifierNode); ok {
			funcName = identifier.Name
		}

		paramCount := 0
		if fn.Parameters != nil {
			paramCount = len(fn.Parameters.Parameters)
		}

		fmt.Printf("   %d. Function '%s' with %d parameters\n", i+1, funcName, paramCount)
	}
	fmt.Println()

	// Example 3: Calculate AST depth
	fmt.Println("3. Calculating AST depth:")
	depthCounter := NewDepthCounter()
	ast2.Walk(depthCounter, program)
	fmt.Printf("   Maximum AST depth: %d\n\n", depthCounter.MaxDepth)

	// Example 4: Using built-in visitor functions
	fmt.Println("4. Using built-in visitor functions:")

	// Find all string literals
	stringLiterals := ast2.FindAllFunc(program, func(node ast2.Node) bool {
		_, ok := node.(*ast2.StringLiteral)
		return ok
	})
	fmt.Printf("   Found %d string literals\n", len(stringLiterals))

	// Find all binary expressions
	binaryExpressions := ast2.FindAllFunc(program, func(node ast2.Node) bool {
		_, ok := node.(*ast2.BinaryExpression)
		return ok
	})
	fmt.Printf("   Found %d binary expressions\n", len(binaryExpressions))

	// Count assignment expressions
	assignmentCount := ast2.CountFunc(program, func(node ast2.Node) bool {
		_, ok := node.(*ast2.AssignmentExpression)
		return ok
	})
	fmt.Printf("   Found %d assignment expressions\n\n", assignmentCount)

	// Example 5: Using WalkFunc for simple operations
	fmt.Println("5. Using WalkFunc for node type analysis:")
	nodeTypes := make(map[string]int)

	ast2.WalkFunc(program, func(node ast2.Node) bool {
		nodeType := fmt.Sprintf("%T", node)
		// Remove package prefix for cleaner output
		if idx := strings.LastIndex(nodeType, "."); idx != -1 {
			nodeType = nodeType[idx+1:]
		}
		nodeTypes[nodeType]++
		return true
	})

	fmt.Println("   Node type statistics:")
	for nodeType, count := range nodeTypes {
		fmt.Printf("     %s: %d\n", nodeType, count)
	}
}

// Helper function to remove duplicates from string slice
func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}
