package runtime

import (
	"testing"
	"time"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegexCaching(t *testing.T) {
	t.Run("Cache basic operations", func(t *testing.T) {
		// Clear cache to start fresh
		cache := getRegexCache()
		cache.clear()

		// Test compilation and caching
		pattern := "/test@\\w+\\.\\w+/"

		// First compilation - should be cache miss
		regex1, err := compilePhpRegex(pattern)
		require.NoError(t, err)
		require.NotNil(t, regex1)

		// Second compilation - should be cache hit
		regex2, err := compilePhpRegex(pattern)
		require.NoError(t, err)
		require.NotNil(t, regex2)

		// Should be the same regex object (cached)
		assert.Equal(t, regex1, regex2)

		// Check cache stats
		stats := cache.stats()
		assert.Equal(t, 1, stats["size"])
		assert.Equal(t, int64(1), stats["hits"])
		assert.Equal(t, int64(1), stats["misses"])
	})

	t.Run("Cache with multiple patterns", func(t *testing.T) {
		cache := getRegexCache()
		cache.clear()

		patterns := []string{
			"/\\d{4}-\\d{2}-\\d{2}/",  // Date pattern
			"/\\b\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\b/",  // IP address
			"/https?:\\/\\/[^\\s]+/",  // URL pattern
		}

		// Compile each pattern twice
		for _, pattern := range patterns {
			// First compilation (miss)
			regex1, err := compilePhpRegex(pattern)
			require.NoError(t, err)
			require.NotNil(t, regex1)

			// Second compilation (hit)
			regex2, err := compilePhpRegex(pattern)
			require.NoError(t, err)
			require.NotNil(t, regex2)

			assert.Equal(t, regex1, regex2)
		}

		// Check final stats
		stats := cache.stats()
		assert.Equal(t, 3, stats["size"])          // 3 patterns cached
		assert.Equal(t, int64(3), stats["hits"])   // 3 cache hits (second compilations)
		assert.Equal(t, int64(3), stats["misses"]) // 3 cache misses (first compilations)
	})

	t.Run("Cache TTL expiration", func(t *testing.T) {
		cache := getRegexCache()
		cache.clear()

		// Configure very short TTL for testing
		cache.configure(100, 100*time.Millisecond)

		pattern := "/test/"

		// First compilation
		regex1, err := compilePhpRegex(pattern)
		require.NoError(t, err)
		require.NotNil(t, regex1)

		// Immediate second compilation (should hit)
		regex2, err := compilePhpRegex(pattern)
		require.NoError(t, err)
		require.NotNil(t, regex2)
		assert.Equal(t, regex1, regex2)

		// Wait for TTL expiration
		time.Sleep(150 * time.Millisecond)

		// Third compilation after expiration (should miss)
		regex3, err := compilePhpRegex(pattern)
		require.NoError(t, err)
		require.NotNil(t, regex3)

		// Check stats
		stats := cache.stats()
		assert.Equal(t, int64(1), stats["hits"])   // One hit before expiration
		assert.Equal(t, int64(2), stats["misses"]) // Two misses (initial and after expiration)
	})

	t.Run("Cache size limit", func(t *testing.T) {
		cache := getRegexCache()
		cache.clear()

		// Configure small cache size
		cache.configure(2, 5*time.Minute)

		patterns := []string{
			"/pattern1/",
			"/pattern2/",
			"/pattern3/",  // This should evict pattern1
		}

		// Fill cache beyond limit
		for _, pattern := range patterns {
			_, err := compilePhpRegex(pattern)
			require.NoError(t, err)
		}

		// Cache should only hold 2 patterns
		stats := cache.stats()
		assert.Equal(t, 2, stats["size"])
		assert.Equal(t, int64(1), stats["evictions"]) // One pattern evicted
	})
}

// mockCacheContext for testing cache builtin functions
type mockCacheContext struct{}

func (m *mockCacheContext) SymbolRegistry() *registry.Registry                       { return nil }
func (m *mockCacheContext) WriteOutput(val *values.Value) error                      { return nil }
func (m *mockCacheContext) GetGlobal(name string) (*values.Value, bool)              { return nil, false }
func (m *mockCacheContext) SetGlobal(name string, val *values.Value)                 {}
func (m *mockCacheContext) LookupUserFunction(name string) (*registry.Function, bool) { return nil, false }
func (m *mockCacheContext) CallUserFunction(function *registry.Function, args []*values.Value) (*values.Value, error) { return nil, nil }
func (m *mockCacheContext) SimpleCallUserFunction(function *registry.Function, args []*values.Value) (*values.Value, error) { return nil, nil }
func (m *mockCacheContext) LookupUserClass(name string) (*registry.Class, bool)      { return nil, false }
func (m *mockCacheContext) Halt(exitCode int, message string) error                  { return nil }
func (m *mockCacheContext) GetExecutionContext() registry.ExecutionContextInterface  { return nil }
func (m *mockCacheContext) GetOutputBufferStack() registry.OutputBufferStackInterface { return nil }
func (m *mockCacheContext) GetCurrentFunctionArgCount() (int, error)                 { return 0, nil }
func (m *mockCacheContext) GetCurrentFunctionArg(index int) (*values.Value, error)   { return nil, nil }
func (m *mockCacheContext) GetCurrentFunctionArgs() ([]*values.Value, error)         { return nil, nil }
func (m *mockCacheContext) ThrowException(exception *values.Value) error { return nil }
func (m *mockCacheContext) GetHTTPContext() registry.HTTPContext { return nil }
func (m *mockCacheContext) ResetHTTPContext() {}
func (m *mockCacheContext) RemoveHTTPHeader(name string) {}

func TestRegexCacheFunctions(t *testing.T) {
	ctx := &mockCacheContext{}

	// Get cache functions
	functions := GetRegexCacheFunctions()
	require.Greater(t, len(functions), 0)

	var statsFunc, clearFunc, configureFunc, cleanupFunc *registry.Function
	for _, fn := range functions {
		switch fn.Name {
		case "preg_cache_stats":
			statsFunc = fn
		case "preg_cache_clear":
			clearFunc = fn
		case "preg_cache_configure":
			configureFunc = fn
		case "preg_cache_cleanup":
			cleanupFunc = fn
		}
	}

	t.Run("Cache stats function", func(t *testing.T) {
		require.NotNil(t, statsFunc)

		// Clear and populate cache
		cache := getRegexCache()
		cache.clear()
		compilePhpRegex("/test/")

		// Call stats function
		result, err := statsFunc.Builtin(ctx, []*values.Value{})
		require.NoError(t, err)
		require.NotNil(t, result)
		require.True(t, result.IsArray(), "Stats should return array")

		arr := result.Data.(*values.Array)
		require.Greater(t, len(arr.Elements), 0, "Stats array should not be empty")

		// Check for expected keys
		assert.Contains(t, arr.Elements, "size")
		assert.Contains(t, arr.Elements, "hits")
		assert.Contains(t, arr.Elements, "misses")
	})

	t.Run("Cache clear function", func(t *testing.T) {
		require.NotNil(t, clearFunc)

		// Populate cache
		compilePhpRegex("/test/")

		// Call clear function
		result, err := clearFunc.Builtin(ctx, []*values.Value{})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.ToBool(), "Clear should return true")

		// Verify cache is empty
		cache := getRegexCache()
		stats := cache.stats()
		assert.Equal(t, 0, stats["size"])
	})

	t.Run("Cache configure function", func(t *testing.T) {
		require.NotNil(t, configureFunc)

		args := []*values.Value{
			values.NewInt(500),  // max_size
			values.NewInt(120),  // ttl_seconds
		}

		result, err := configureFunc.Builtin(ctx, args)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.ToBool(), "Configure should return true")

		// Verify configuration was applied
		cache := getRegexCache()
		stats := cache.stats()
		assert.Equal(t, 500, stats["maxSize"])
		assert.Equal(t, 120*time.Second, stats["ttl"])
	})

	t.Run("Cache cleanup function", func(t *testing.T) {
		require.NotNil(t, cleanupFunc)

		result, err := cleanupFunc.Builtin(ctx, []*values.Value{})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsInt(), "Cleanup should return int")
	})
}