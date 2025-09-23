package runtime

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

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

// compilePhpRegex compiles a PHP-style regex pattern
func compilePhpRegex(pattern string) (*regexp.Regexp, error) {
	clearRegexError()

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
					return values.NewNull(), nil
				}

				result := regex.ReplaceAllString(subject, replacement)
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
	}
}