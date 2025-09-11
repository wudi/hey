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

// CodeMetrics 代码度量结构
type CodeMetrics struct {
	LinesOfCode          int
	FunctionCount        int
	ClassCount           int
	MethodCount          int
	PropertyCount        int
	VariableCount        int
	ComplexityScore      int
	MaxNestingDepth      int
	CyclomaticComplexity int
}

// CodeAnalyzer 代码分析器
type CodeAnalyzer struct {
	Metrics           *CodeMetrics
	Functions         []*FunctionInfo
	Classes           []*ClassInfo
	Variables         map[string]int
	Issues            []AnalysisIssue
	currentDepth      int
	currentComplexity int
}

type FunctionInfo struct {
	Name           string
	ParameterCount int
	ReturnType     string
	Visibility     string
	LineCount      int
	Complexity     int
}

type ClassInfo struct {
	Name        string
	MethodCount int
	Properties  []string
	Methods     []*FunctionInfo
	Visibility  map[string]int // public, private, protected counts
}

type AnalysisIssue struct {
	Type       string
	Severity   string
	Message    string
	Location   string
	Suggestion string
}

func NewCodeAnalyzer() *CodeAnalyzer {
	return &CodeAnalyzer{
		Metrics:   &CodeMetrics{},
		Functions: make([]*FunctionInfo, 0),
		Classes:   make([]*ClassInfo, 0),
		Variables: make(map[string]int),
		Issues:    make([]AnalysisIssue, 0),
	}
}

func (ca *CodeAnalyzer) Visit(node ast2.Node) bool {
	switch n := node.(type) {
	case *ast2.FunctionDeclaration:
		ca.analyzeFunctionDeclaration(n)
	case *ast2.ClassExpression:
		ca.analyzeClassDeclaration(n)
	case *ast2.Variable:
		ca.analyzeVariable(n)
	case *ast2.IfStatement:
		ca.analyzeControlStructure("if")
	case *ast2.WhileStatement:
		ca.analyzeControlStructure("while")
	case *ast2.ForStatement:
		ca.analyzeControlStructure("for")
	case *ast2.BinaryExpression:
		ca.analyzeBinaryExpression(n)
	}

	// 跟踪嵌套深度
	ca.currentDepth++
	if ca.currentDepth > ca.Metrics.MaxNestingDepth {
		ca.Metrics.MaxNestingDepth = ca.currentDepth
	}

	defer func() {
		ca.currentDepth--
	}()

	return true
}

func (ca *CodeAnalyzer) analyzeFunctionDeclaration(fn *ast2.FunctionDeclaration) {
	ca.Metrics.FunctionCount++

	funcInfo := &FunctionInfo{
		Complexity: 1, // Base complexity
	}

	if identifier, ok := fn.Name.(*ast2.IdentifierNode); ok {
		funcInfo.Name = identifier.Name
	}

	if fn.Parameters != nil {
		funcInfo.ParameterCount = len(fn.Parameters.Parameters)
	}

	funcInfo.Visibility = fn.Visibility

	// 检查函数复杂度
	if funcInfo.ParameterCount > 5 {
		ca.addIssue("complexity", "warning",
			fmt.Sprintf("Function '%s' has too many parameters (%d)", funcInfo.Name, funcInfo.ParameterCount),
			funcInfo.Name, "Consider reducing parameter count or using parameter objects")
	}

	ca.Functions = append(ca.Functions, funcInfo)
}

func (ca *CodeAnalyzer) analyzeClassDeclaration(class *ast2.ClassExpression) {
	ca.Metrics.ClassCount++

	classInfo := &ClassInfo{
		Properties: make([]string, 0),
		Methods:    make([]*FunctionInfo, 0),
		Visibility: make(map[string]int),
	}

	if identifier, ok := class.Name.(*ast2.IdentifierNode); ok {
		classInfo.Name = identifier.Name
	}

	ca.Classes = append(ca.Classes, classInfo)
}

func (ca *CodeAnalyzer) analyzeVariable(variable *ast2.Variable) {
	ca.Variables[variable.Name]++
	ca.Metrics.VariableCount++
}

func (ca *CodeAnalyzer) analyzeControlStructure(structType string) {
	ca.currentComplexity++
	ca.Metrics.CyclomaticComplexity++
}

func (ca *CodeAnalyzer) analyzeBinaryExpression(expr *ast2.BinaryExpression) {
	// 增加复杂度分数，基于操作符类型
	if expr.Operator == "&&" || expr.Operator == "||" {
		ca.currentComplexity++
	}
}

func (ca *CodeAnalyzer) addIssue(issueType, severity, message, location, suggestion string) {
	issue := AnalysisIssue{
		Type:       issueType,
		Severity:   severity,
		Message:    message,
		Location:   location,
		Suggestion: suggestion,
	}
	ca.Issues = append(ca.Issues, issue)
}

func (ca *CodeAnalyzer) generateReport() {
	fmt.Println("=== Code Analysis Report ===\n")

	// 基本度量
	fmt.Println("📊 Code Metrics:")
	fmt.Printf("  Lines of Code: %d\n", ca.Metrics.LinesOfCode)
	fmt.Printf("  Functions: %d\n", ca.Metrics.FunctionCount)
	fmt.Printf("  Classes: %d\n", ca.Metrics.ClassCount)
	fmt.Printf("  Variables: %d unique\n", len(ca.Variables))
	fmt.Printf("  Max Nesting Depth: %d\n", ca.Metrics.MaxNestingDepth)
	fmt.Printf("  Cyclomatic Complexity: %d\n", ca.Metrics.CyclomaticComplexity)
	fmt.Println()

	// 函数分析
	if len(ca.Functions) > 0 {
		fmt.Println("🔧 Function Analysis:")
		for i, fn := range ca.Functions {
			fmt.Printf("  %d. %s()", i+1, fn.Name)
			if fn.ParameterCount > 0 {
				fmt.Printf(" - %d parameters", fn.ParameterCount)
			}
			if fn.Visibility != "" {
				fmt.Printf(" - %s", fn.Visibility)
			}
			fmt.Println()
		}
		fmt.Println()
	}

	// 类分析
	if len(ca.Classes) > 0 {
		fmt.Println("🏛️  Class Analysis:")
		for i, class := range ca.Classes {
			fmt.Printf("  %d. %s\n", i+1, class.Name)
		}
		fmt.Println()
	}

	// 变量使用统计
	if len(ca.Variables) > 0 {
		fmt.Println("📝 Variable Usage:")
		for variable, count := range ca.Variables {
			fmt.Printf("  $%s: used %d times\n", variable, count)
		}
		fmt.Println()
	}

	// 问题报告
	if len(ca.Issues) > 0 {
		fmt.Println("⚠️  Issues Found:")
		for i, issue := range ca.Issues {
			fmt.Printf("  %d. [%s] %s: %s\n", i+1, strings.ToUpper(issue.Severity), issue.Type, issue.Message)
			if issue.Location != "" {
				fmt.Printf("     Location: %s\n", issue.Location)
			}
			if issue.Suggestion != "" {
				fmt.Printf("     Suggestion: %s\n", issue.Suggestion)
			}
		}
		fmt.Println()
	}

	// 总体评分
	ca.generateScoreCard()
}

func (ca *CodeAnalyzer) generateScoreCard() {
	fmt.Println("📈 Code Quality Score Card:")

	score := 100

	// 基于复杂度扣分
	if ca.Metrics.CyclomaticComplexity > 10 {
		score -= 10
		fmt.Println("  - High cyclomatic complexity (-10)")
	}

	// 基于嵌套深度扣分
	if ca.Metrics.MaxNestingDepth > 5 {
		score -= 15
		fmt.Println("  - Deep nesting detected (-15)")
	}

	// 基于问题数量扣分
	if len(ca.Issues) > 0 {
		score -= len(ca.Issues) * 5
		fmt.Printf("  - %d issues found (-%d)\n", len(ca.Issues), len(ca.Issues)*5)
	}

	if score < 0 {
		score = 0
	}

	fmt.Printf("\n  Overall Score: %d/100 ", score)

	switch {
	case score >= 90:
		fmt.Println("🌟 Excellent")
	case score >= 80:
		fmt.Println("✅ Good")
	case score >= 70:
		fmt.Println("👍 Fair")
	case score >= 60:
		fmt.Println("⚠️  Needs Improvement")
	default:
		fmt.Println("❌ Poor")
	}
	fmt.Println()
}

func main() {
	// 复杂的PHP代码示例用于分析
	phpCode := `<?php
namespace App\Services;

use App\Models\User;
use App\Utils\Logger;

/**
 * User management service
 */
class UserService 
{
    private $logger;
    private $database;
    private static $instance = null;
    
    const MAX_USERS = 1000;
    const CACHE_TTL = 3600;
    
    public function __construct($database, Logger $logger) {
        $this->database = $database;
        $this->logger = $logger;
    }
    
    /**
     * Complex user validation with multiple conditions
     */
    public function validateUser($username, $password, $email, $age, $country, $preferences) {
        if (empty($username) || strlen($username) < 3) {
            return false;
        }
        
        if (empty($password) || strlen($password) < 8) {
            return false;
        }
        
        if (!filter_var($email, FILTER_VALIDATE_EMAIL)) {
            return false;
        }
        
        if ($age < 13 || $age > 120) {
            return false;
        }
        
        // Nested conditions increase complexity
        if ($country === 'US') {
            if ($age < 18) {
                if (!$this->hasParentalConsent($username)) {
                    return false;
                }
            }
        } else if ($country === 'EU') {
            if ($age < 16) {
                if (!$this->hasParentalConsent($username)) {
                    return false;
                }
            }
        }
        
        return true;
    }
    
    private function hasParentalConsent($username) {
        return $this->database->checkParentalConsent($username);
    }
    
    public function createUser($userData) {
        $user = new User();
        $user->username = $userData['username'];
        $user->email = $userData['email'];
        $user->created_at = time();
        
        if ($this->database->save($user)) {
            $this->logger->info("User created: " . $user->username);
            return $user;
        }
        
        return false;
    }
    
    public function deleteUser($userId) {
        $user = $this->database->find($userId);
        if ($user) {
            $this->database->delete($userId);
            $this->logger->info("User deleted: " . $userId);
            return true;
        }
        return false;
    }
}

// 函数级别的代码
function calculateUserScore($user, $activities, $timeframe) {
    $score = 0;
    $baseScore = 100;
    $activityWeight = 0.3;
    $timeWeight = 0.2;
    $bonusThreshold = 50;
    
    foreach ($activities as $activity) {
        if ($activity['type'] === 'login') {
            $score += 10 * $activityWeight;
        } else if ($activity['type'] === 'post') {
            $score += 20 * $activityWeight;
        } else if ($activity['type'] === 'comment') {
            $score += 15 * $activityWeight;
        }
        
        if ($activity['timestamp'] > $timeframe) {
            $score += $score * $timeWeight;
        }
    }
    
    if ($score > $bonusThreshold) {
        $score += $baseScore * 0.1;
    }
    
    return min($score, 1000);
}
?>`

	fmt.Println("=== PHP Code Analysis Example ===\n")

	// 解析代码
	l := lexer.New(phpCode)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Println("❌ Parsing errors found:")
		for _, err := range p.Errors() {
			fmt.Printf("  %s\n", err)
		}
		return
	}

	fmt.Println("✅ Code parsed successfully, starting analysis...\n")

	// 分析代码
	analyzer := NewCodeAnalyzer()
	analyzer.Metrics.LinesOfCode = strings.Count(phpCode, "\n") + 1
	ast2.Walk(analyzer, program)

	// 生成分析报告
	analyzer.generateReport()

	// 额外的分析示例
	fmt.Println("=== Additional Analysis Examples ===\n")

	// 查找特定模式
	fmt.Println("🔍 Pattern Analysis:")

	// 查找所有字符串字面量
	stringLiterals := ast2.FindAllFunc(program, func(node ast2.Node) bool {
		_, ok := node.(*ast2.StringLiteral)
		return ok
	})
	fmt.Printf("  String literals found: %d\n", len(stringLiterals))

	// 查找所有数组表达式
	arrayExpressions := ast2.FindAllFunc(program, func(node ast2.Node) bool {
		_, ok := node.(*ast2.ArrayExpression)
		return ok
	})
	fmt.Printf("  Array expressions found: %d\n", len(arrayExpressions))

	// 计算二进制表达式数量
	binaryExprCount := ast2.CountFunc(program, func(node ast2.Node) bool {
		_, ok := node.(*ast2.BinaryExpression)
		return ok
	})
	fmt.Printf("  Binary expressions: %d\n", binaryExprCount)

	// 计算赋值表达式数量
	assignmentCount := ast2.CountFunc(program, func(node ast2.Node) bool {
		_, ok := node.(*ast2.AssignmentExpression)
		return ok
	})
	fmt.Printf("  Assignment expressions: %d\n", assignmentCount)

	fmt.Println()

	// 建议报告
	fmt.Println("💡 Improvement Suggestions:")
	fmt.Println("  • Consider breaking down complex functions into smaller ones")
	fmt.Println("  • Use dependency injection for better testability")
	fmt.Println("  • Add type hints for better code documentation")
	fmt.Println("  • Consider using constants for magic numbers")
	fmt.Println("  • Implement proper error handling with exceptions")
}
