package runtime

import (
	"container/list"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// PCRE error constants
const (
	PREG_NO_ERROR              = 0
	PREG_INTERNAL_ERROR        = 1
	PREG_BACKTRACK_LIMIT_ERROR = 2
	PREG_RECURSION_LIMIT_ERROR = 3
	PREG_BAD_UTF8_ERROR        = 4
	PREG_BAD_UTF8_OFFSET_ERROR = 5
	PREG_JIT_STACKLIMIT_ERROR  = 6
)

// Global error state for regex operations
var (
	regexErrorMutex sync.RWMutex
	lastRegexError  int    = PREG_NO_ERROR
	lastErrorMsg    string = ""
)

// Regex cache configuration
const (
	DefaultCacheMaxSize = 1000   // Maximum number of cached patterns
	DefaultCacheTTL     = 5 * time.Minute // Cache entry time-to-live
)

// cachedRegex represents a cached compiled regex with metadata
type cachedRegex struct {
	regex       *regexp.Regexp
	compiledAt  time.Time
	accessCount int64
	lastAccess  time.Time
}

// regexCache implements an LRU cache with TTL for compiled regex patterns
type regexCache struct {
	mutex    sync.RWMutex
	cache    map[string]*list.Element
	lruList  *list.List
	maxSize  int
	ttl      time.Duration

	// Statistics
	hits     int64
	misses   int64
	evictions int64
}

// cacheEntry represents an entry in the LRU list
type cacheEntry struct {
	pattern string
	cached  *cachedRegex
}

// Global regex cache instance
var (
	globalRegexCache *regexCache
	cacheOnce       sync.Once
)

// getRegexCache returns the singleton regex cache instance
func getRegexCache() *regexCache {
	cacheOnce.Do(func() {
		globalRegexCache = &regexCache{
			cache:   make(map[string]*list.Element),
			lruList: list.New(),
			maxSize: DefaultCacheMaxSize,
			ttl:     DefaultCacheTTL,
		}
	})
	return globalRegexCache
}

// get retrieves a compiled regex from the cache
func (c *regexCache) get(pattern string) (*regexp.Regexp, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	element, exists := c.cache[pattern]
	if !exists {
		c.misses++
		return nil, false
	}

	entry := element.Value.(*cacheEntry)
	cached := entry.cached

	// Check if entry has expired
	if time.Since(cached.compiledAt) > c.ttl {
		c.removeElement(element)
		c.evictions++
		c.misses++
		return nil, false
	}

	// Update access statistics and move to front (most recently used)
	cached.accessCount++
	cached.lastAccess = time.Now()
	c.lruList.MoveToFront(element)
	c.hits++

	return cached.regex, true
}

// put stores a compiled regex in the cache
func (c *regexCache) put(pattern string, regex *regexp.Regexp) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Check if pattern already exists (update case)
	if element, exists := c.cache[pattern]; exists {
		entry := element.Value.(*cacheEntry)
		entry.cached.regex = regex
		entry.cached.compiledAt = time.Now()
		entry.cached.accessCount = 1
		entry.cached.lastAccess = time.Now()
		c.lruList.MoveToFront(element)
		return
	}

	// Create new cache entry
	cached := &cachedRegex{
		regex:       regex,
		compiledAt:  time.Now(),
		accessCount: 1,
		lastAccess:  time.Now(),
	}

	entry := &cacheEntry{
		pattern: pattern,
		cached:  cached,
	}

	element := c.lruList.PushFront(entry)
	c.cache[pattern] = element

	// Evict oldest entries if cache exceeds max size
	for c.lruList.Len() > c.maxSize {
		oldest := c.lruList.Back()
		if oldest != nil {
			c.removeElement(oldest)
			c.evictions++
		}
	}
}

// removeElement removes an element from both the cache map and LRU list
func (c *regexCache) removeElement(element *list.Element) {
	entry := element.Value.(*cacheEntry)
	delete(c.cache, entry.pattern)
	c.lruList.Remove(element)
}

// clear removes all entries from the cache
func (c *regexCache) clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache = make(map[string]*list.Element)
	c.lruList = list.New()
	c.hits = 0
	c.misses = 0
	c.evictions = 0
}

// stats returns cache statistics
func (c *regexCache) stats() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	total := c.hits + c.misses
	hitRate := float64(0)
	if total > 0 {
		hitRate = float64(c.hits) / float64(total)
	}

	return map[string]interface{}{
		"size":        c.lruList.Len(),
		"maxSize":     c.maxSize,
		"hits":        c.hits,
		"misses":      c.misses,
		"evictions":   c.evictions,
		"hitRate":     hitRate,
		"ttl":         c.ttl,
	}
}

// configure allows runtime configuration of cache parameters
func (c *regexCache) configure(maxSize int, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.maxSize = maxSize
	c.ttl = ttl

	// Evict excess entries if new max size is smaller
	for c.lruList.Len() > c.maxSize {
		oldest := c.lruList.Back()
		if oldest != nil {
			c.removeElement(oldest)
			c.evictions++
		}
	}
}

// cleanupExpired removes expired entries from the cache
func (c *regexCache) cleanupExpired() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	removed := 0
	now := time.Now()

	// Walk from back to front (oldest to newest)
	for element := c.lruList.Back(); element != nil; {
		entry := element.Value.(*cacheEntry)
		if now.Sub(entry.cached.compiledAt) > c.ttl {
			next := element.Prev()
			c.removeElement(element)
			removed++
			element = next
		} else {
			// Since entries are ordered by recency, we can break early
			break
		}
	}

	return removed
}

// setRegexError sets the last regex error
func setRegexError(errorCode int, message string) {
	regexErrorMutex.Lock()
	defer regexErrorMutex.Unlock()
	lastRegexError = errorCode
	lastErrorMsg = message
}

// getRegexError gets the last regex error
func getRegexError() (int, string) {
	regexErrorMutex.RLock()
	defer regexErrorMutex.RUnlock()
	return lastRegexError, lastErrorMsg
}

// clearRegexError clears the last regex error
func clearRegexError() {
	regexErrorMutex.Lock()
	defer regexErrorMutex.Unlock()
	lastRegexError = PREG_NO_ERROR
	lastErrorMsg = ""
}

// parsePhpPattern parses PHP regex pattern with delimiters and flags
func parsePhpPattern(pattern string) (string, string, error) {
	if len(pattern) < 2 {
		return "", "", fmt.Errorf("invalid pattern: too short")
	}

	// Find delimiter (first character)
	delimiter := pattern[0:1]

	// Find closing delimiter
	lastPos := strings.LastIndex(pattern[1:], delimiter)
	if lastPos == -1 {
		return "", "", fmt.Errorf("no ending delimiter '%s' found", delimiter)
	}

	// Extract pattern and flags
	actualPattern := pattern[1 : lastPos+1]
	flags := ""
	if len(pattern) > lastPos+2 {
		flags = pattern[lastPos+2:]
	}

	return actualPattern, flags, nil
}

// convertPhpFlags converts PHP regex flags to Go regex syntax
func convertPhpFlags(pattern string, flags string) (string, error) {
	var goPattern strings.Builder

	// Handle case-insensitive flag
	if strings.Contains(flags, "i") {
		goPattern.WriteString("(?i)")
	}

	// Handle multiline flag
	if strings.Contains(flags, "m") {
		goPattern.WriteString("(?m)")
	}

	// Handle dot-all flag (. matches newlines)
	if strings.Contains(flags, "s") {
		goPattern.WriteString("(?s)")
	}

	goPattern.WriteString(pattern)
	return goPattern.String(), nil
}

// compilePhpRegex compiles a PHP-style regex pattern with caching
func compilePhpRegex(pattern string) (*regexp.Regexp, error) {
	clearRegexError()

	// Try to get from cache first
	cache := getRegexCache()
	if cachedRegex, found := cache.get(pattern); found {
		return cachedRegex, nil
	}

	// Cache miss - compile the pattern
	actualPattern, flags, err := parsePhpPattern(pattern)
	if err != nil {
		setRegexError(PREG_INTERNAL_ERROR, err.Error())
		return nil, err
	}

	goPattern, err := convertPhpFlags(actualPattern, flags)
	if err != nil {
		setRegexError(PREG_INTERNAL_ERROR, err.Error())
		return nil, err
	}

	regex, err := regexp.Compile(goPattern)
	if err != nil {
		setRegexError(PREG_INTERNAL_ERROR, err.Error())
		return nil, err
	}

	// Store in cache for future use
	cache.put(pattern, regex)

	return regex, nil
}

// GetRegexFunctions returns all regex-related PHP functions
func GetRegexFunctions() []*registry.Function {
	return []*registry.Function{
		{
			Name: "preg_match",
			Parameters: []*registry.Parameter{
				{Name: "pattern", Type: "string"},
				{Name: "subject", Type: "string"},
				{Name: "matches", Type: "array", IsReference: true, HasDefault: true, DefaultValue: values.NewNull()},
				{Name: "flags", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
				{Name: "offset", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
			},
			ReturnType: "int|false",
			MinArgs:    2,
			MaxArgs:    5,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 {
					return values.NewBool(false), nil
				}

				pattern := args[0].ToString()
				subject := args[1].ToString()

				regex, err := compilePhpRegex(pattern)
				if err != nil {
					return values.NewBool(false), nil
				}

				matches := regex.FindStringSubmatch(subject)
				if matches == nil {
					return values.NewInt(0), nil
				}

				// If matches parameter is provided, populate it
				if len(args) > 2 {
					var targetValue *values.Value

					// Handle nil or undefined reference parameters
					if args[2] == nil {
						// Create new array for undefined reference parameter
						newArray := values.NewArray()
						args[2] = newArray
						targetValue = newArray
					} else if args[2].Type == values.TypeReference {
						ref := args[2].Data.(*values.Reference)
						if ref.Target == nil || ref.Target.Type != values.TypeArray {
							// Convert the existing null value to an array in-place
							if ref.Target == nil {
								ref.Target = values.NewArray()
							} else {
								// Transform the existing value to array type
								ref.Target.Type = values.TypeArray
								ref.Target.Data = &values.Array{
									Elements:  make(map[interface{}]*values.Value),
									NextIndex: 0,
									IsIndexed: true,
								}
							}
							targetValue = ref.Target
						} else {
							targetValue = ref.Target
						}
					} else {
						// Direct value - ensure it's an array
						if args[2].Type != values.TypeArray {
							newArray := values.NewArray()
							args[2] = newArray
							targetValue = newArray
						} else {
							targetValue = args[2]
						}
					}

					// Clear existing array and populate with matches
					arr := targetValue.Data.(*values.Array)
					// Clear existing elements
					arr.Elements = make(map[interface{}]*values.Value)

					// Trim trailing empty strings to match PHP behavior
					// PHP omits unmatched optional capture groups from the end
					trimmedMatches := matches
					for len(trimmedMatches) > 1 && trimmedMatches[len(trimmedMatches)-1] == "" {
						trimmedMatches = trimmedMatches[:len(trimmedMatches)-1]
					}

					// Populate with trimmed matches
					for i, match := range trimmedMatches {
						arr.Elements[int64(i)] = values.NewString(match)
					}
					arr.NextIndex = int64(len(trimmedMatches))
				}

				return values.NewInt(1), nil
			},
		},
		{
			Name: "preg_match_all",
			Parameters: []*registry.Parameter{
				{Name: "pattern", Type: "string"},
				{Name: "subject", Type: "string"},
				{Name: "matches", Type: "array", IsReference: true, HasDefault: true, DefaultValue: values.NewNull()},
				{Name: "flags", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
				{Name: "offset", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
			},
			ReturnType: "int|false",
			MinArgs:    2,
			MaxArgs:    5,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 {
					return values.NewBool(false), nil
				}

				pattern := args[0].ToString()
				subject := args[1].ToString()

				regex, err := compilePhpRegex(pattern)
				if err != nil {
					return values.NewBool(false), nil
				}

				allMatches := regex.FindAllStringSubmatch(subject, -1)
				if allMatches == nil {
					allMatches = [][]string{} // Set to empty slice instead of nil
				}

				// If matches parameter is provided, populate it
				if len(args) > 2 {
					var targetValue *values.Value

					// Handle nil or undefined reference parameters
					if args[2] == nil {
						// Create new array for undefined reference parameter
						newArray := values.NewArray()
						args[2] = newArray
						targetValue = newArray
					} else if args[2].Type == values.TypeReference {
						ref := args[2].Data.(*values.Reference)
						if ref.Target == nil || ref.Target.Type != values.TypeArray {
							// Convert the existing null value to an array in-place
							if ref.Target == nil {
								ref.Target = values.NewArray()
							} else {
								// Transform the existing value to array type
								ref.Target.Type = values.TypeArray
								ref.Target.Data = &values.Array{
									Elements:  make(map[interface{}]*values.Value),
									NextIndex: 0,
									IsIndexed: true,
								}
							}
							targetValue = ref.Target
						} else {
							targetValue = ref.Target
						}
					} else {
						// Direct value - ensure it's an array
						if args[2].Type != values.TypeArray {
							newArray := values.NewArray()
							args[2] = newArray
							targetValue = newArray
						} else {
							targetValue = args[2]
						}
					}

					// Clear existing array and populate with matches in PHP format
					arr := targetValue.Data.(*values.Array)
					arr.Elements = make(map[interface{}]*values.Value)

					if len(allMatches) > 0 {
						// Figure out how many capture groups we have
						numGroups := len(allMatches[0])

						// Create sub-arrays for each capture group
						for groupIndex := 0; groupIndex < numGroups; groupIndex++ {
							groupArray := values.NewArray()
							groupArr := groupArray.Data.(*values.Array)

							// Populate this group's matches across all match sets
							for matchIndex, match := range allMatches {
								if groupIndex < len(match) {
									groupArr.Elements[int64(matchIndex)] = values.NewString(match[groupIndex])
								}
							}
							groupArr.NextIndex = int64(len(allMatches))

							// Add this group array to the main matches array
							arr.Elements[int64(groupIndex)] = groupArray
						}
						arr.NextIndex = int64(numGroups)
					} else {
						// No matches found - create empty array at index 0
						emptyArray := values.NewArray()
						arr.Elements[int64(0)] = emptyArray
						arr.NextIndex = 1
					}
				}

				return values.NewInt(int64(len(allMatches))), nil
			},
		},
		{
			Name: "preg_replace",
			Parameters: []*registry.Parameter{
				{Name: "pattern", Type: "string|array"},
				{Name: "replacement", Type: "string|array"},
				{Name: "subject", Type: "string|array"},
				{Name: "limit", Type: "int", HasDefault: true, DefaultValue: values.NewInt(-1)},
				{Name: "count", Type: "int", IsReference: true, HasDefault: true, DefaultValue: values.NewNull()},
			},
			ReturnType: "string|array|null",
			MinArgs:    3,
			MaxArgs:    5,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 3 {
					return values.NewNull(), nil
				}

				pattern := args[0].ToString()
				replacement := args[1].ToString()
				subject := args[2].ToString()

				regex, err := compilePhpRegex(pattern)
				if err != nil {
					setRegexError(PREG_INTERNAL_ERROR, fmt.Sprintf("Invalid regex pattern: %s", pattern))
					return values.NewNull(), nil
				}

				// Convert PHP-style backreferences to Go-style
				// Convert $1, $2, etc. to ${1}, ${2}, etc. for Go regex
				goReplacement := convertPhpBackreferences(replacement)

				result := regex.ReplaceAllString(subject, goReplacement)
				clearRegexError()
				return values.NewString(result), nil
			},
		},
		{
			Name: "preg_split",
			Parameters: []*registry.Parameter{
				{Name: "pattern", Type: "string"},
				{Name: "subject", Type: "string"},
				{Name: "limit", Type: "int", HasDefault: true, DefaultValue: values.NewInt(-1)},
				{Name: "flags", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
			},
			ReturnType: "array|false",
			MinArgs:    2,
			MaxArgs:    4,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 {
					return values.NewBool(false), nil
				}

				pattern := args[0].ToString()
				subject := args[1].ToString()
				limit := int(-1)
				if len(args) > 2 {
					limit = int(args[2].ToInt())
				}

				regex, err := compilePhpRegex(pattern)
				if err != nil {
					return values.NewBool(false), nil
				}

				parts := regex.Split(subject, limit)
				result := values.NewArray()
				arr := result.Data.(*values.Array)
				for i, part := range parts {
					arr.Elements[int64(i)] = values.NewString(part)
				}
				arr.NextIndex = int64(len(parts))

				return result, nil
			},
		},
		{
			Name: "preg_quote",
			Parameters: []*registry.Parameter{
				{Name: "str", Type: "string"},
				{Name: "delimiter", Type: "string", HasDefault: true, DefaultValue: values.NewNull()},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return values.NewString(""), nil
				}

				str := args[0].ToString()
				delimiter := ""
				if len(args) > 1 && args[1] != nil {
					delimiter = args[1].ToString()
				}

				// Quote regex metacharacters
				quoted := regexp.QuoteMeta(str)

				// Quote delimiter if provided
				if delimiter != "" && len(delimiter) > 0 {
					delim := string(delimiter[0])
					quoted = strings.ReplaceAll(quoted, delim, "\\"+delim)
				}

				return values.NewString(quoted), nil
			},
		},
		{
			Name: "preg_grep",
			Parameters: []*registry.Parameter{
				{Name: "pattern", Type: "string"},
				{Name: "array", Type: "array"},
				{Name: "flags", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
			},
			ReturnType: "array|false",
			MinArgs:    2,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 {
					return values.NewBool(false), nil
				}

				pattern := args[0].ToString()
				inputArray := args[1]

				if inputArray.Type != values.TypeArray {
					return values.NewBool(false), nil
				}

				regex, err := compilePhpRegex(pattern)
				if err != nil {
					return values.NewBool(false), nil
				}

				result := values.NewArray()
				resultArr := result.Data.(*values.Array)
				inputArr := inputArray.Data.(*values.Array)

				for key, val := range inputArr.Elements {
					strVal := val.ToString()
					if regex.MatchString(strVal) {
						resultArr.Elements[key] = val
					}
				}

				return result, nil
			},
		},
		{
			Name: "preg_last_error",
			Parameters: []*registry.Parameter{},
			ReturnType: "int",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				errorCode, _ := getRegexError()
				return values.NewInt(int64(errorCode)), nil
			},
		},
		{
			Name: "preg_last_error_msg",
			Parameters: []*registry.Parameter{},
			ReturnType: "string",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				_, errorMsg := getRegexError()
				return values.NewString(errorMsg), nil
			},
		},
		{
			Name: "preg_filter",
			Parameters: []*registry.Parameter{
				{Name: "pattern", Type: "string|array"},
				{Name: "replacement", Type: "string|array"},
				{Name: "subject", Type: "string|array"},
				{Name: "limit", Type: "int", HasDefault: true, DefaultValue: values.NewInt(-1)},
				{Name: "count", Type: "int", IsReference: true, HasDefault: true, DefaultValue: values.NewNull()},
			},
			ReturnType: "string|array|null",
			MinArgs:    3,
			MaxArgs:    5,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 3 {
					return values.NewNull(), nil
				}

				pattern := args[0].ToString()
				replacement := args[1].ToString()

				regex, err := compilePhpRegex(pattern)
				if err != nil {
					return values.NewNull(), nil
				}

				// Handle different subject types
				subject := args[2]
				if subject.Type == values.TypeString {
					// For strings, preg_filter acts like preg_replace
					subjectStr := subject.ToString()
					result := regex.ReplaceAllString(subjectStr, replacement)
					return values.NewString(result), nil
				} else if subject.Type == values.TypeArray {
					// For arrays, filter out non-matching elements
					inputArr := subject.Data.(*values.Array)
					result := values.NewArray()
					resultArr := result.Data.(*values.Array)

					for key, val := range inputArr.Elements {
						strVal := val.ToString()
						if regex.MatchString(strVal) {
							// Element matches, apply replacement and include in result
							replaced := regex.ReplaceAllString(strVal, replacement)
							resultArr.Elements[key] = values.NewString(replaced)
						}
						// Non-matching elements are filtered out (not included in result)
					}

					return result, nil
				}

				// Unsupported type
				return values.NewNull(), nil
			},
		},
		{
			Name: "preg_replace_callback",
			Parameters: []*registry.Parameter{
				{Name: "pattern", Type: "string|array"},
				{Name: "callback", Type: "callable"},
				{Name: "subject", Type: "string|array"},
				{Name: "limit", Type: "int", HasDefault: true, DefaultValue: values.NewInt(-1)},
				{Name: "count", Type: "int", IsReference: true, HasDefault: true, DefaultValue: values.NewNull()},
			},
			ReturnType: "string|array|null",
			MinArgs:    3,
			MaxArgs:    5,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 3 {
					return values.NewNull(), nil
				}

				clearRegexError() // Clear any previous errors

				pattern := args[0].ToString()
				callback := args[1]
				subject := args[2]
				limit := int(-1)
				if len(args) > 3 && !args[3].IsNull() {
					limit = int(args[3].ToInt())
				}

				// Compile the regex pattern
				regex, err := compilePhpRegex(pattern)
				if err != nil {
					setRegexError(PREG_INTERNAL_ERROR, fmt.Sprintf("Invalid regex pattern: %s", err.Error()))
					return values.NewNull(), nil
				}

				// Handle string subject
				if subject.Type == values.TypeString {
					subjectStr := subject.ToString()
					replacementCount := 0

					// Find all matches with capture groups
					allMatches := regex.FindAllStringSubmatch(subjectStr, limit)
					if len(allMatches) == 0 {
						return values.NewString(subjectStr), nil // No matches, return original
					}

					// Process each match and build result string manually
					result := subjectStr
					offset := 0

					for _, matchData := range allMatches {
						if len(matchData) == 0 {
							continue
						}

						// Check limit
						if limit >= 0 && replacementCount >= limit {
							break
						}
						replacementCount++

						// Create matches array (similar to preg_match format)
						matches := values.NewArray()
						matchesArr := matches.Data.(*values.Array)

						// Add all capture groups (including full match at index 0)
						for i, submatch := range matchData {
							matchesArr.Elements[int64(i)] = values.NewString(submatch)
						}
						matchesArr.NextIndex = int64(len(matchData))

						// Call the callback with matches array
						// For builtin string functions, pass just the matched string instead of the full array
						var callArgs []*values.Value
						if callback.Type == values.TypeString {
							funcName := callback.ToString()
							// Check if it's a common string builtin that expects a single string
							stringBuiltins := []string{"strtoupper", "strtolower", "ucfirst", "lcfirst", "trim", "ltrim", "rtrim"}
							isStringBuiltin := false
							for _, builtin := range stringBuiltins {
								if funcName == builtin {
									isStringBuiltin = true
									break
								}
							}
							if isStringBuiltin {
								// Pass just the matched string (matches[0])
								callArgs = []*values.Value{values.NewString(matchData[0])}
							} else {
								// Pass the full matches array for user-defined functions
								callArgs = []*values.Value{matches}
							}
						} else {
							// For closures and other callbacks, pass the matches array
							callArgs = []*values.Value{matches}
						}

						callbackResult, err := callbackInvoker(ctx, callback, callArgs)
						if err != nil {
							setRegexError(PREG_INTERNAL_ERROR, fmt.Sprintf("Callback error: %s", err.Error()))
							continue // Skip this match on error
						}

						var replacement string
						if callbackResult != nil {
							replacement = callbackResult.ToString()
						} else {
							replacement = matchData[0] // Use original match if callback returns null
						}

						// Replace the first occurrence of the full match in the result
						fullMatch := matchData[0]
						if idx := strings.Index(result[offset:], fullMatch); idx != -1 {
							actualIdx := offset + idx
							result = result[:actualIdx] + replacement + result[actualIdx+len(fullMatch):]
							offset = actualIdx + len(replacement)
						}
					}

					// Set count reference if provided
					// TODO: Implement proper reference parameter handling
					_ = replacementCount // Avoid unused variable warning for now

					return values.NewString(result), nil
				}

				// Handle array subject
				if subject.Type == values.TypeArray {
					inputArr := subject.Data.(*values.Array)
					result := values.NewArray()
					resultArr := result.Data.(*values.Array)
					totalReplacements := 0

					for key, val := range inputArr.Elements {
						if val == nil {
							continue
						}

						valStr := val.ToString()
						elementReplacements := 0

						// Process each array element
						elementResult := regex.ReplaceAllStringFunc(valStr, func(match string) string {
							// Check global limit
							if limit >= 0 && totalReplacements >= limit {
								return match
							}
							elementReplacements++
							totalReplacements++

							// Create matches array
							matches := values.NewArray()
							matchesArr := matches.Data.(*values.Array)
							matchesArr.Elements[int64(0)] = values.NewString(match)
							matchesArr.NextIndex = 1

							// Call the callback
							callArgs := []*values.Value{matches}
							result, err := callbackInvoker(ctx, callback, callArgs)
							if err != nil {
								setRegexError(PREG_INTERNAL_ERROR, fmt.Sprintf("Callback error: %s", err.Error()))
								return match
							}

							if result == nil {
								return match
							}
							return result.ToString()
						})

						resultArr.Elements[key] = values.NewString(elementResult)
					}

					// Set count reference if provided
					// TODO: Implement proper reference parameter handling
					_ = totalReplacements // Avoid unused variable warning for now

					return result, nil
				}

				// Unsupported type
				return values.NewNull(), nil
			},
		},
	}
}

// convertPhpBackreferences converts PHP-style backreferences ($1, $2, etc.) to Go-style (${1}, ${2}, etc.)
func convertPhpBackreferences(replacement string) string {
	// Use a simple regex to find and replace $1, $2, etc. with ${1}, ${2}, etc.
	// This handles the most common PHP backreference syntax
	result := replacement

	// Convert $1-9 to ${1}-${9} (most common case)
	for i := 1; i <= 9; i++ {
		oldPattern := fmt.Sprintf("$%d", i)
		newPattern := fmt.Sprintf("${%d}", i)
		result = strings.ReplaceAll(result, oldPattern, newPattern)
	}

	// Also handle $0 (full match)
	result = strings.ReplaceAll(result, "$0", "${0}")

	return result
}

// GetRegexCacheFunctions returns regex cache management PHP functions
func GetRegexCacheFunctions() []*registry.Function {
	return []*registry.Function{
		{
			Name:       "preg_cache_stats",
			Parameters: []*registry.Parameter{},
			ReturnType: "array",
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				cache := getRegexCache()
				stats := cache.stats()

				// Convert stats to PHP array
				result := values.NewArray()
				resultArr := result.Data.(*values.Array)

				for key, value := range stats {
					var val *values.Value
					switch v := value.(type) {
					case int:
						val = values.NewInt(int64(v))
					case int64:
						val = values.NewInt(v)
					case float64:
						val = values.NewFloat(v)
					case time.Duration:
						val = values.NewString(v.String())
					default:
						val = values.NewString(fmt.Sprintf("%v", v))
					}
					resultArr.Elements[key] = val
				}

				return result, nil
			},
		},
		{
			Name:       "preg_cache_clear",
			Parameters: []*registry.Parameter{},
			ReturnType: "bool",
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				cache := getRegexCache()
				cache.clear()
				return values.NewBool(true), nil
			},
		},
		{
			Name: "preg_cache_configure",
			Parameters: []*registry.Parameter{
				{Name: "max_size", Type: "int", HasDefault: true, DefaultValue: values.NewInt(DefaultCacheMaxSize)},
				{Name: "ttl_seconds", Type: "int", HasDefault: true, DefaultValue: values.NewInt(int64(DefaultCacheTTL.Seconds()))},
			},
			ReturnType: "bool",
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				maxSize := DefaultCacheMaxSize
				ttlSeconds := int64(DefaultCacheTTL.Seconds())

				if len(args) > 0 {
					maxSize = int(args[0].ToInt())
				}
				if len(args) > 1 {
					ttlSeconds = args[1].ToInt()
				}

				// Validate parameters
				if maxSize <= 0 {
					return values.NewBool(false), fmt.Errorf("max_size must be greater than 0")
				}
				if ttlSeconds <= 0 {
					return values.NewBool(false), fmt.Errorf("ttl_seconds must be greater than 0")
				}

				cache := getRegexCache()
				cache.configure(maxSize, time.Duration(ttlSeconds)*time.Second)
				return values.NewBool(true), nil
			},
		},
		{
			Name:       "preg_cache_cleanup",
			Parameters: []*registry.Parameter{},
			ReturnType: "int",
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				cache := getRegexCache()
				removed := cache.cleanupExpired()
				return values.NewInt(int64(removed)), nil
			},
		},
	}
}