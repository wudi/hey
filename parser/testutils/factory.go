package testutils

// DefaultParserFactory 默认解析器工厂函数类型
// 该函数需要在parser包中实现，避免循环依赖
var DefaultParserFactory ParserFactory

// NewTestBuilder 创建带默认解析器工厂的测试构建器
// 使用前需要在parser包中设置DefaultParserFactory
func NewTestBuilder() *ParserTestBuilder {
	if DefaultParserFactory == nil {
		panic("DefaultParserFactory not set. Call testutils.SetDefaultParserFactory() first")
	}
	return NewParserTestBuilder(DefaultParserFactory)
}

// SetDefaultParserFactory 设置默认解析器工厂
func SetDefaultParserFactory(factory ParserFactory) {
	DefaultParserFactory = factory
}

// NewQuickTestBuilder 创建快速测试构建器（用于parser包内部）
func NewQuickTestBuilder(factory ParserFactory) *ParserTestBuilder {
	return NewParserTestBuilder(factory)
}