package testutils

import (
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/lexer"
)

// TestContext 提供测试上下文和工具
type TestContext struct {
	T       *testing.T
	Parser  ParserInterface
	Lexer   *lexer.Lexer
	Program *ast.Program
	Config  *TestConfig
}

// TestConfig 测试配置选项
type TestConfig struct {
	StrictMode  bool // 严格模式：解析错误时失败测试
	ValidateAST bool // 验证AST结构
}

// ParserFactory 解析器工厂函数类型
type ParserFactory func(*lexer.Lexer) ParserInterface

// ParserTestBuilder 解析器测试构建器
type ParserTestBuilder struct {
	config        *TestConfig
	setup         []func(*TestContext)
	parserFactory ParserFactory
}

// NewParserTestBuilder 创建解析器测试构建器
// parserFactory 参数用于创建解析器实例
func NewParserTestBuilder(parserFactory ParserFactory) *ParserTestBuilder {
	return &ParserTestBuilder{
		config: &TestConfig{
			StrictMode:  true,
			ValidateAST: true,
		},
		parserFactory: parserFactory,
	}
}

// WithConfig 设置测试配置
func (b *ParserTestBuilder) WithConfig(config *TestConfig) *ParserTestBuilder {
	b.config = config
	return b
}

// WithStrictMode 设置严格模式
func (b *ParserTestBuilder) WithStrictMode(strict bool) *ParserTestBuilder {
	b.config.StrictMode = strict
	return b
}

// WithSetup 添加设置函数
func (b *ParserTestBuilder) WithSetup(setup func(*TestContext)) *ParserTestBuilder {
	b.setup = append(b.setup, setup)
	return b
}

// Test 执行测试
func (b *ParserTestBuilder) Test(t *testing.T, source string, validator func(*TestContext)) {
	t.Helper()
	
	ctx := &TestContext{
		T:      t,
		Lexer:  lexer.New(source),
		Config: b.config,
	}
	
	ctx.Parser = b.parserFactory(ctx.Lexer)
	
	// 执行设置函数
	for _, setup := range b.setup {
		setup(ctx)
	}
	
	// 解析程序
	ctx.Program = ctx.Parser.ParseProgram()
	
	// 错误检查
	if b.config.StrictMode {
		CheckParserErrors(t, ctx.Parser)
	}
	
	// AST验证
	if b.config.ValidateAST {
		require.NotNil(t, ctx.Program, "Program should not be nil")
	}
	
	// 执行验证函数
	if validator != nil {
		validator(ctx)
	}
}

// TestTableDriven 执行表驱动测试
func (b *ParserTestBuilder) TestTableDriven(t *testing.T, tests []struct {
	Name      string
	Source    string
	Validator func(*TestContext)
}) {
	t.Helper()
	
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			b.Test(t, tt.Source, tt.Validator)
		})
	}
}

// ExpectProgram 验证程序体长度
func ExpectProgram(expectedBodyLen int) func(*TestContext) {
	return func(ctx *TestContext) {
		assert.Len(ctx.T, ctx.Program.Body, expectedBodyLen)
	}
}

// ExpectNoErrors 验证无解析错误
func ExpectNoErrors() func(*TestContext) {
	return func(ctx *TestContext) {
		errors := ctx.Parser.Errors()
		assert.Empty(ctx.T, errors, "Expected no parsing errors but got: %v", errors)
	}
}