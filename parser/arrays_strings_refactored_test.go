package parser

import (
	"testing"

	"github.com/wudi/php-parser/parser/testutils"
)

// TestRefactored_ArrayExpressions 重构后的数组表达式测试
func TestRefactored_ArrayExpressions(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("ArrayExpressions", createParserFactory())

	// 基础数组测试
	suite.AddSimple("basic_array",
		`<?php $arr = [1, 2, 3]; ?>`,
		testutils.ValidateArrayAssignment("$arr", []testutils.ArrayElement{
			{Value: "1", IsNumeric: true},
			{Value: "2", IsNumeric: true},
			{Value: "3", IsNumeric: true},
		}))

	// 带键的关联数组
	suite.AddSimple("associative_array",
		`<?php $arr = ["key1" => "value1", "key2" => "value2"]; ?>`,
		testutils.ValidateArrayAssignment("$arr", []testutils.ArrayElement{
			{Key: `"key1"`, Value: `"value1"`},
			{Key: `"key2"`, Value: `"value2"`},
		}))

	// 混合数组
	suite.AddSimple("mixed_array",
		`<?php $arr = [1, "key" => "value", 2]; ?>`,
		testutils.ValidateArrayAssignment("$arr", []testutils.ArrayElement{
			{Value: "1", IsNumeric: true},
			{Key: `"key"`, Value: `"value"`},
			{Value: "2", IsNumeric: true},
		}))

	// 带尾逗号的数组
	suite.AddSimple("array_trailing_comma",
		`<?php $arr = [1, 2, 3,]; ?>`,
		testutils.ValidateArrayAssignment("$arr", []testutils.ArrayElement{
			{Value: "1", IsNumeric: true},
			{Value: "2", IsNumeric: true},
			{Value: "3", IsNumeric: true},
		}))

	// array()函数语法
	suite.AddSimple("array_function_syntax",
		`<?php $arr = array(1, 2, 3); ?>`,
		testutils.ValidateArrayAssignment("$arr", []testutils.ArrayElement{
			{Value: "1", IsNumeric: true},
			{Value: "2", IsNumeric: true},
			{Value: "3", IsNumeric: true},
		}))

	suite.Run(t)
}

// TestRefactored_StringLiterals 重构后的字符串字面量测试
func TestRefactored_StringLiterals(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("StringLiterals", createParserFactory())

	// 基础字符串赋值
	suite.AddStringAssignment("basic_string", "$str", "Hello", `"Hello"`)

	// 单引号字符串
	suite.AddStringAssignment("single_quote_string", "$str", "World", `'World'`)

	// 空字符串
	suite.AddStringAssignment("empty_string", "$str", "", `""`)

	suite.Run(t)
}

// TestRefactored_HeredocNowdoc 重构后的Heredoc和Nowdoc测试
func TestRefactored_HeredocNowdoc(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("HeredocNowdoc", createParserFactory())

	// 简单Heredoc - 修正尾部换行符
	suite.AddSimple("simple_heredoc",
		`<?php $str = <<<EOD
Hello World
EOD; ?>`,
		testutils.ValidateHeredocAssignment("$str", "Hello World\n"))

	// 简单Nowdoc - 修正尾部换行符
	suite.AddSimple("simple_nowdoc",
		`<?php $str = <<<'EOD'
Hello World
EOD; ?>`,
		testutils.ValidateNowdocAssignment("$str", "Hello World\n"))

	// 简单Heredoc无插值测试
	suite.AddSimple("heredoc_no_interpolation",
		`<?php $str = <<<EOD
Hello John
EOD; ?>`,
		testutils.ValidateHeredocAssignment("$str", "Hello John\n"))

	suite.Run(t)
}

// TestRefactored_ArrayAccess 重构后的数组访问测试
func TestRefactored_ArrayAccess(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("ArrayAccess", createParserFactory())

	// 数组元素访问
	suite.AddSimple("array_element_access",
		`<?php $value = $arr[0]; ?>`,
		testutils.ValidateArrayAccess("$value", "$arr", "0"))

	// 关联数组访问
	suite.AddSimple("associative_array_access",
		`<?php $value = $arr["key"]; ?>`,
		testutils.ValidateArrayAccess("$value", "$arr", `"key"`))

	// 多维数组访问
	suite.AddSimple("multi_dimensional_access",
		`<?php $value = $arr[0][1]; ?>`,
		testutils.ValidateChainedArrayAccess("$value", "$arr", []string{"0", "1"}))

	suite.Run(t)
}
