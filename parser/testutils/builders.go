package testutils

import (
	"testing"
	
	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/lexer"
)

// TestSuiteBuilder 测试套件构建器
type TestSuiteBuilder struct {
	name          string
	tests         []TestCase
	parserFactory ParserFactory
	config        *TestConfig
}

// TestCase 标准测试用例结构
type TestCase struct {
	Name      string
	Source    string
	Validator ValidationFunc
	Skip      bool
	Only      bool
	Tags      []string
}

// NewTestSuiteBuilder 创建测试套件构建器
func NewTestSuiteBuilder(name string, parserFactory ParserFactory) *TestSuiteBuilder {
	return &TestSuiteBuilder{
		name:          name,
		parserFactory: parserFactory,
		config: &TestConfig{
			StrictMode:  true,
			ValidateAST: true,
		},
	}
}

// WithConfig 设置配置
func (b *TestSuiteBuilder) WithConfig(config *TestConfig) *TestSuiteBuilder {
	b.config = config
	return b
}

// Add 添加测试用例
func (b *TestSuiteBuilder) Add(testCase TestCase) *TestSuiteBuilder {
	b.tests = append(b.tests, testCase)
	return b
}

// AddSimple 添加简单测试用例
func (b *TestSuiteBuilder) AddSimple(name, source string, validator ValidationFunc) *TestSuiteBuilder {
	return b.Add(TestCase{
		Name:      name,
		Source:    source,
		Validator: validator,
	})
}

// AddStringAssignment 添加字符串赋值测试
func (b *TestSuiteBuilder) AddStringAssignment(name, varName, value, raw string) *TestSuiteBuilder {
	return b.AddSimple(name,
		"<?php "+varName+" = "+raw+"; ?>",
		ValidateStringAssignment(varName, value, raw))
}

// AddVariableAssignment 添加变量赋值测试
func (b *TestSuiteBuilder) AddVariableAssignment(name, varName, valueSource string) *TestSuiteBuilder {
	return b.AddSimple(name,
		"<?php "+varName+" = "+valueSource+"; ?>",
		ValidateVariable(varName))
}

// AddEcho 添加echo测试
func (b *TestSuiteBuilder) AddEcho(name string, args []string, validators ...func(ast.Node, *testing.T)) *TestSuiteBuilder {
	argsStr := ""
	for i, arg := range args {
		if i > 0 {
			argsStr += ", "
		}
		argsStr += arg
	}
	
	return b.AddSimple(name,
		"<?php echo "+argsStr+"; ?>",
		ValidateEcho(len(args), validators...))
}

// AddFunction 添加函数测试
func (b *TestSuiteBuilder) AddFunction(name, funcName string, params []string, validators ...func(*ast.FunctionDeclaration, *testing.T)) *TestSuiteBuilder {
	paramsStr := ""
	for i, param := range params {
		if i > 0 {
			paramsStr += ", "
		}
		paramsStr += param
	}
	
	return b.AddSimple(name,
		"<?php function "+funcName+"("+paramsStr+") {} ?>",
		ValidateFunction(funcName, len(params), validators...))
}

// AddClass 添加类测试
func (b *TestSuiteBuilder) AddClass(name, className, classBody string, validators ...func(*ast.ClassExpression, *testing.T)) *TestSuiteBuilder {
	return b.AddSimple(name,
		"<?php class "+className+" { "+classBody+" } ?>",
		ValidateClass(className, validators...))
}

// AddControlFlow 添加控制流测试
func (b *TestSuiteBuilder) AddControlFlow(name, flowType, source string, validators ...func(ast.Statement, *testing.T)) *TestSuiteBuilder {
	return b.AddSimple(name,
		"<?php "+source+" ?>",
		ValidateControlFlow(flowType, validators...))
}

// Skip 标记跳过测试
func (b *TestSuiteBuilder) Skip(testName string) *TestSuiteBuilder {
	for i, test := range b.tests {
		if test.Name == testName {
			b.tests[i].Skip = true
			break
		}
	}
	return b
}

// Only 仅运行特定测试
func (b *TestSuiteBuilder) Only(testName string) *TestSuiteBuilder {
	for i, test := range b.tests {
		if test.Name == testName {
			b.tests[i].Only = true
		}
	}
	return b
}

// Run 执行测试套件
func (b *TestSuiteBuilder) Run(t *testing.T) {
	// 检查是否有Only标记的测试
	hasOnly := false
	for _, test := range b.tests {
		if test.Only {
			hasOnly = true
			break
		}
	}
	
	builder := NewParserTestBuilder(b.parserFactory).WithConfig(b.config)
	
	for _, test := range b.tests {
		// 如果有Only标记，只运行标记的测试
		if hasOnly && !test.Only {
			continue
		}
		
		// 跳过标记的测试
		if test.Skip {
			t.Run(test.Name, func(t *testing.T) {
				t.Skip("Test marked as skip")
			})
			continue
		}
		
		t.Run(test.Name, func(t *testing.T) {
			builder.Test(t, test.Source, test.Validator)
		})
	}
}

// BenchmarkBuilder 性能测试构建器
type BenchmarkBuilder struct {
	name          string
	tests         []BenchmarkCase
	parserFactory ParserFactory
}

// BenchmarkCase 性能测试用例
type BenchmarkCase struct {
	Name   string
	Source string
}

// NewBenchmarkBuilder 创建性能测试构建器
func NewBenchmarkBuilder(name string, parserFactory ParserFactory) *BenchmarkBuilder {
	return &BenchmarkBuilder{
		name:          name,
		parserFactory: parserFactory,
	}
}

// Add 添加性能测试
func (b *BenchmarkBuilder) Add(name, source string) *BenchmarkBuilder {
	b.tests = append(b.tests, BenchmarkCase{Name: name, Source: source})
	return b
}

// Run 运行性能测试
func (b *BenchmarkBuilder) Run(bench *testing.B) {
	for _, test := range b.tests {
		bench.Run(test.Name, func(innerB *testing.B) {
			innerB.ResetTimer()
			for i := 0; i < innerB.N; i++ {
				ctx := &TestContext{
					T:      &testing.T{}, // 临时的T，性能测试不需要断言
					Lexer:  lexer.New(test.Source),
					Config: &TestConfig{StrictMode: false, ValidateAST: false},
				}
				ctx.Parser = b.parserFactory(ctx.Lexer)
				ctx.Program = ctx.Parser.ParseProgram()
			}
		})
	}
}