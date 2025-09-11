package testutils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ErrorTestBuilder 错误测试构建器
type ErrorTestBuilder struct {
	parserFactory ParserFactory
}

// NewErrorTestBuilder 创建错误测试构建器
func NewErrorTestBuilder(parserFactory ParserFactory) *ErrorTestBuilder {
	return &ErrorTestBuilder{parserFactory: parserFactory}
}

// ExpectError 验证解析错误
func (b *ErrorTestBuilder) ExpectError(t *testing.T, source string, expectedErrorSubstring string) {
	t.Helper()

	builder := NewParserTestBuilder(b.parserFactory).WithStrictMode(false)

	builder.Test(t, source, func(ctx *TestContext) {
		errors := ctx.Parser.Errors()
		assert.NotEmpty(t, errors, "Expected parsing errors but got none")

		if len(errors) > 0 {
			found := false
			for _, err := range errors {
				if strings.Contains(err, expectedErrorSubstring) {
					found = true
					break
				}
			}
			assert.True(t, found, "Expected error containing '%s' but got errors: %v", expectedErrorSubstring, errors)
		}
	})
}

// ExpectNoError 验证无解析错误
func (b *ErrorTestBuilder) ExpectNoError(t *testing.T, source string) {
	t.Helper()

	builder := NewParserTestBuilder(b.parserFactory).WithStrictMode(true)

	builder.Test(t, source, func(ctx *TestContext) {
		errors := ctx.Parser.Errors()
		assert.Empty(t, errors, "Expected no parsing errors but got: %v", errors)
	})
}

// ExpectRecovery 验证错误恢复（有错误但仍能解析部分内容）
func (b *ErrorTestBuilder) ExpectRecovery(t *testing.T, source string, expectedStatementCount int) {
	t.Helper()

	builder := NewParserTestBuilder(b.parserFactory).WithStrictMode(false)

	builder.Test(t, source, func(ctx *TestContext) {
		errors := ctx.Parser.Errors()
		assert.NotEmpty(t, errors, "Expected parsing errors but got none")

		// 验证仍然解析了一些内容
		if ctx.Program != nil {
			assert.Len(t, ctx.Program.Body, expectedStatementCount,
				"Expected %d statements despite errors", expectedStatementCount)
		}
	})
}

// ErrorTestSuite 错误测试套件
type ErrorTestSuite struct {
	name    string
	builder *ErrorTestBuilder
	tests   []ErrorTestCase
}

// ErrorTestCase 错误测试用例
type ErrorTestCase struct {
	Name         string
	Source       string
	ExpectError  bool
	ErrorKeyword string
	RecoverCount int // 错误恢复时期望的语句数量
}

// NewErrorTestSuite 创建错误测试套件
func NewErrorTestSuite(name string, parserFactory ParserFactory) *ErrorTestSuite {
	return &ErrorTestSuite{
		name:    name,
		builder: NewErrorTestBuilder(parserFactory),
	}
}

// Add 添加错误测试用例
func (s *ErrorTestSuite) Add(testCase ErrorTestCase) *ErrorTestSuite {
	s.tests = append(s.tests, testCase)
	return s
}

// AddError 添加期望错误的测试
func (s *ErrorTestSuite) AddError(name, source, errorKeyword string) *ErrorTestSuite {
	return s.Add(ErrorTestCase{
		Name:         name,
		Source:       source,
		ExpectError:  true,
		ErrorKeyword: errorKeyword,
	})
}

// AddSuccess 添加期望成功的测试
func (s *ErrorTestSuite) AddSuccess(name, source string) *ErrorTestSuite {
	return s.Add(ErrorTestCase{
		Name:        name,
		Source:      source,
		ExpectError: false,
	})
}

// AddRecovery 添加错误恢复测试
func (s *ErrorTestSuite) AddRecovery(name, source string, recoverCount int) *ErrorTestSuite {
	return s.Add(ErrorTestCase{
		Name:         name,
		Source:       source,
		ExpectError:  true,
		RecoverCount: recoverCount,
	})
}

// Run 运行错误测试套件
func (s *ErrorTestSuite) Run(t *testing.T) {
	for _, test := range s.tests {
		t.Run(test.Name, func(t *testing.T) {
			if test.ExpectError {
				if test.RecoverCount > 0 {
					s.builder.ExpectRecovery(t, test.Source, test.RecoverCount)
				} else if test.ErrorKeyword != "" {
					s.builder.ExpectError(t, test.Source, test.ErrorKeyword)
				} else {
					// 只期望有错误，不关心具体内容
					s.builder.ExpectError(t, test.Source, "")
				}
			} else {
				s.builder.ExpectNoError(t, test.Source)
			}
		})
	}
}

// ValidateErrorMessage 创建错误消息验证函数
func ValidateErrorMessage(expectedKeywords ...string) ValidationFunc {
	return func(ctx *TestContext) {
		errors := ctx.Parser.Errors()
		assert.NotEmpty(ctx.T, errors, "Expected parsing errors but got none")

		for _, keyword := range expectedKeywords {
			found := false
			for _, err := range errors {
				if strings.Contains(err, keyword) {
					found = true
					break
				}
			}
			assert.True(ctx.T, found, "Expected error containing '%s' but got errors: %v", keyword, errors)
		}
	}
}

// ValidateNoErrors 验证无错误的验证函数
func ValidateNoErrors() ValidationFunc {
	return func(ctx *TestContext) {
		errors := ctx.Parser.Errors()
		assert.Empty(ctx.T, errors, "Expected no parsing errors but got: %v", errors)
	}
}
