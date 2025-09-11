package testutils

// GlobalConfig 全局测试配置
type GlobalConfig struct {
	DefaultStrictMode bool
	EnableDebug       bool
	TestDataDir       string
	GoldenFileDir     string
	BenchmarkEnabled  bool
	CoverageEnabled   bool
}

// DefaultGlobalConfig 默认全局配置
var DefaultGlobalConfig = &GlobalConfig{
	DefaultStrictMode: true,
	EnableDebug:       false,
	TestDataDir:       "testdata",
	GoldenFileDir:     "testdata/golden",
	BenchmarkEnabled:  true,
	CoverageEnabled:   true,
}

// SetGlobalConfig 设置全局配置
func SetGlobalConfig(config *GlobalConfig) {
	DefaultGlobalConfig = config
}

// GetGlobalConfig 获取全局配置
func GetGlobalConfig() *GlobalConfig {
	return DefaultGlobalConfig
}

// TestCategory 测试分类常量
const (
	CategoryBasic       = "basic"
	CategoryExpressions = "expressions"
	CategoryStatements  = "statements"
	CategoryClasses     = "classes"
	CategoryFunctions   = "functions"
	CategoryControlFlow = "control_flow"
	CategoryErrors      = "errors"
	CategoryPHP8        = "php8"
	CategoryPHP81       = "php81"
	CategoryPHP82       = "php82"
	CategoryPHP83       = "php83"
	CategoryPHP84       = "php84"
)

// TestTags 测试标签常量
const (
	TagSmoke        = "smoke"
	TagRegression   = "regression"
	TagSlow         = "slow"
	TagExperimental = "experimental"
	TagBenchmark    = "benchmark"
	TagIntegration  = "integration"
	TagUnit         = "unit"
	TagError        = "error"
)

// IsTestEnabled 检查测试是否启用
func IsTestEnabled(tags []string) bool {
	// 如果没有标签，默认启用
	if len(tags) == 0 {
		return true
	}

	// 检查是否有实验性标签，如果有且未开启调试模式，则跳过
	for _, tag := range tags {
		if tag == TagExperimental && !DefaultGlobalConfig.EnableDebug {
			return false
		}
		if tag == TagSlow && !DefaultGlobalConfig.BenchmarkEnabled {
			return false
		}
	}

	return true
}
