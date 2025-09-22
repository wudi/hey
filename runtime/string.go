package runtime

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"
	"hash/crc32"
	"math"
	"math/rand"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetStringFunctions returns all string-related PHP functions
func GetStringFunctions() []*registry.Function {
	return []*registry.Function{
		{
			Name:       "strlen",
			Parameters: []*registry.Parameter{{Name: "str", Type: "string"}},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewInt(0), nil
				}
				return values.NewInt(int64(len(args[0].ToString()))), nil
			},
		},
		{
			Name: "strpos",
			Parameters: []*registry.Parameter{
				{Name: "haystack", Type: "string"},
				{Name: "needle", Type: "string"},
			},
			ReturnType: "int|false",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 {
					return values.NewBool(false), nil
				}
				haystack := args[0].ToString()
				needle := args[1].ToString()
				idx := strings.Index(haystack, needle)
				if idx == -1 {
					return values.NewBool(false), nil
				}
				return values.NewInt(int64(idx)), nil
			},
		},
		{
			Name: "substr",
			Parameters: []*registry.Parameter{
				{Name: "string", Type: "string"},
				{Name: "offset", Type: "int"},
				{Name: "length", Type: "int", HasDefault: true, DefaultValue: values.NewNull()},
			},
			ReturnType: "string|false",
			MinArgs:    2,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 {
					return values.NewBool(false), nil
				}

				str := args[0].ToString()
				offset := int(args[1].ToInt())
				strLen := len(str)

				// Handle negative offset
				if offset < 0 {
					offset = strLen + offset
					if offset < 0 {
						offset = 0
					}
				}

				// If offset is beyond string length, return false
				if offset >= strLen {
					return values.NewBool(false), nil
				}

				// Determine length
				var length int
				if len(args) >= 3 && !args[2].IsNull() {
					length = int(args[2].ToInt())
				} else {
					length = strLen - offset
				}

				// Handle negative length
				if length < 0 {
					endPos := strLen + length
					if endPos <= offset {
						return values.NewString(""), nil
					}
					length = endPos - offset
				}

				// Ensure we don't go past the end of the string
				if offset+length > strLen {
					length = strLen - offset
				}

				// Extract substring
				if offset < strLen && length > 0 {
					end := offset + length
					if end > strLen {
						end = strLen
					}
					return values.NewString(str[offset:end]), nil
				}

				return values.NewString(""), nil
			},
		},
		{
			Name: "str_repeat",
			Parameters: []*registry.Parameter{
				{Name: "input", Type: "string"},
				{Name: "multiplier", Type: "int"},
			},
			ReturnType: "string",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 {
					return values.NewString(""), nil
				}
				str := args[0].ToString()
				times := int(args[1].ToInt())
				if times <= 0 {
					return values.NewString(""), nil
				}
				return values.NewString(strings.Repeat(str, times)), nil
			},
		},
		{
			Name: "str_replace",
			Parameters: []*registry.Parameter{
				{Name: "search", Type: "mixed"},
				{Name: "replace", Type: "mixed"},
				{Name: "subject", Type: "mixed"},
			},
			ReturnType: "mixed",
			MinArgs:    3,
			MaxArgs:    4,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 3 {
					return values.NewString(""), nil
				}

				search := args[0].ToString()
				replace := args[1].ToString()
				subject := args[2].ToString()

				// PHP str_replace performs simple string replacement
				result := strings.ReplaceAll(subject, search, replace)
				return values.NewString(result), nil
			},
		},
		{
			Name:       "strtolower",
			Parameters: []*registry.Parameter{{Name: "string", Type: "string"}},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewString(""), nil
				}
				return values.NewString(strings.ToLower(args[0].ToString())), nil
			},
		},
		{
			Name:       "strtoupper",
			Parameters: []*registry.Parameter{{Name: "string", Type: "string"}},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewString(""), nil
				}
				return values.NewString(strings.ToUpper(args[0].ToString())), nil
			},
		},
		{
			Name:       "trim",
			Parameters: []*registry.Parameter{{Name: "string", Type: "string"}},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewString(""), nil
				}

				str := args[0].ToString()
				// PHP trim removes whitespace by default
				if len(args) == 1 {
					result := strings.TrimSpace(str)
					return values.NewString(result), nil
				}

				// Custom character mask (simplified version)
				mask := args[1].ToString()
				result := strings.Trim(str, mask)
				return values.NewString(result), nil
			},
		},
		{
			Name:       "ltrim",
			Parameters: []*registry.Parameter{{Name: "string", Type: "string"}},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewString(""), nil
				}

				str := args[0].ToString()
				// Left trim whitespace by default
				if len(args) == 1 {
					result := strings.TrimLeftFunc(str, func(r rune) bool {
						return r == ' ' || r == '\t' || r == '\n' || r == '\r'
					})
					return values.NewString(result), nil
				}

				// Custom character mask
				mask := args[1].ToString()
				result := strings.TrimLeft(str, mask)
				return values.NewString(result), nil
			},
		},
		{
			Name:       "rtrim",
			Parameters: []*registry.Parameter{{Name: "string", Type: "string"}},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewString(""), nil
				}

				str := args[0].ToString()
				// Right trim whitespace by default
				if len(args) == 1 {
					result := strings.TrimRightFunc(str, func(r rune) bool {
						return r == ' ' || r == '\t' || r == '\n' || r == '\r'
					})
					return values.NewString(result), nil
				}

				// Custom character mask
				mask := args[1].ToString()
				result := strings.TrimRight(str, mask)
				return values.NewString(result), nil
			},
		},
		{
			Name: "explode",
			Parameters: []*registry.Parameter{
				{Name: "delimiter", Type: "string"},
				{Name: "string", Type: "string"},
			},
			ReturnType: "array",
			MinArgs:    2,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 {
					return values.NewArray(), nil
				}

				delimiter := args[0].ToString()
				str := args[1].ToString()

				// Split the string
				parts := strings.Split(str, delimiter)

				// Convert to PHP array
				result := values.NewArray()
				resultArr := result.Data.(*values.Array)
				for i, part := range parts {
					resultArr.Elements[int64(i)] = values.NewString(part)
				}

				return result, nil
			},
		},
		{
			Name: "implode",
			Parameters: []*registry.Parameter{
				{Name: "separator", Type: "string"},
				{Name: "array", Type: "array"},
			},
			ReturnType: "string",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[1] == nil || !args[1].IsArray() {
					return values.NewString(""), nil
				}

				separator := args[0].ToString()
				arr := args[1].Data.(*values.Array)

				// Convert array elements to strings and join
				var parts []string
				for _, value := range arr.Elements {
					if value != nil {
						parts = append(parts, value.ToString())
					}
				}

				result := strings.Join(parts, separator)
				return values.NewString(result), nil
			},
		},
		{
			Name: "sprintf",
			Parameters: []*registry.Parameter{
				{Name: "format", Type: "string"},
			},
			ReturnType: "string",
			IsVariadic: true,
			MinArgs:    1,
			MaxArgs:    -1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewString(""), nil
				}

				format := args[0].ToString()
				var goArgs []interface{}

				// Convert PHP values to Go interface{} for fmt.Sprintf
				for i := 1; i < len(args); i++ {
					if args[i] == nil {
						goArgs = append(goArgs, nil)
					} else if args[i].IsInt() {
						goArgs = append(goArgs, args[i].ToInt())
					} else if args[i].IsFloat() {
						goArgs = append(goArgs, args[i].ToFloat())
					} else if args[i].IsBool() {
						goArgs = append(goArgs, args[i].ToBool())
					} else {
						goArgs = append(goArgs, args[i].ToString())
					}
				}

				// Use Go's fmt.Sprintf which supports similar format specifiers
				result := fmt.Sprintf(format, goArgs...)
				return values.NewString(result), nil
			},
		},
		{
			Name: "strrpos",
			Parameters: []*registry.Parameter{
				{Name: "haystack", Type: "string"},
				{Name: "needle", Type: "string"},
			},
			ReturnType: "int|false",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 {
					return values.NewBool(false), nil
				}
				haystack := args[0].ToString()
				needle := args[1].ToString()

				// Special case: empty needle returns length of haystack
				if needle == "" {
					return values.NewInt(int64(len(haystack))), nil
				}

				// Find last occurrence
				idx := strings.LastIndex(haystack, needle)
				if idx == -1 {
					return values.NewBool(false), nil
				}
				return values.NewInt(int64(idx)), nil
			},
		},
		{
			Name: "stripos",
			Parameters: []*registry.Parameter{
				{Name: "haystack", Type: "string"},
				{Name: "needle", Type: "string"},
			},
			ReturnType: "int|false",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 {
					return values.NewBool(false), nil
				}
				haystack := strings.ToLower(args[0].ToString())
				needle := strings.ToLower(args[1].ToString())

				if needle == "" {
					return values.NewBool(false), nil
				}

				idx := strings.Index(haystack, needle)
				if idx == -1 {
					return values.NewBool(false), nil
				}
				return values.NewInt(int64(idx)), nil
			},
		},
		{
			Name: "substr_count",
			Parameters: []*registry.Parameter{
				{Name: "haystack", Type: "string"},
				{Name: "needle", Type: "string"},
			},
			ReturnType: "int",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 {
					return values.NewInt(0), nil
				}
				haystack := args[0].ToString()
				needle := args[1].ToString()

				if needle == "" || haystack == "" {
					return values.NewInt(0), nil
				}

				// Count non-overlapping occurrences
				count := 0
				start := 0
				for {
					idx := strings.Index(haystack[start:], needle)
					if idx == -1 {
						break
					}
					count++
					start += idx + len(needle)
					if start >= len(haystack) {
						break
					}
				}
				return values.NewInt(int64(count)), nil
			},
		},
		{
			Name: "ucfirst",
			Parameters: []*registry.Parameter{
				{Name: "str", Type: "string"},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewString(""), nil
				}
				str := args[0].ToString()

				if str == "" {
					return values.NewString(""), nil
				}

				// Convert first character to uppercase
				result := strings.ToUpper(str[:1]) + str[1:]
				return values.NewString(result), nil
			},
		},
		{
			Name: "lcfirst",
			Parameters: []*registry.Parameter{
				{Name: "str", Type: "string"},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewString(""), nil
				}
				str := args[0].ToString()

				if str == "" {
					return values.NewString(""), nil
				}

				// Convert first character to lowercase
				result := strings.ToLower(str[:1]) + str[1:]
				return values.NewString(result), nil
			},
		},
		{
			Name: "ucwords",
			Parameters: []*registry.Parameter{
				{Name: "str", Type: "string"},
				{Name: "delimiters", Type: "string", HasDefault: true, DefaultValue: values.NewString(" \t\r\n\f\v")},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewString(""), nil
				}
				str := args[0].ToString()

				if str == "" {
					return values.NewString(""), nil
				}

				delimiters := " \t\r\n\f\v"
				if len(args) > 1 && args[1] != nil {
					delimiters = args[1].ToString()
				}

				// Convert to runes for proper Unicode handling
				runes := []rune(str)
				result := make([]rune, len(runes))
				copy(result, runes)

				// First character is always uppercased if it's a letter
				if len(result) > 0 {
					result[0] = []rune(strings.ToUpper(string(result[0])))[0]
				}

				// Uppercase character after any delimiter
				for i := 1; i < len(result); i++ {
					if strings.ContainsRune(delimiters, result[i-1]) {
						result[i] = []rune(strings.ToUpper(string(result[i])))[0]
					}
				}

				return values.NewString(string(result)), nil
			},
		},
		{
			Name: "str_ireplace",
			Parameters: []*registry.Parameter{
				{Name: "search", Type: "mixed"},
				{Name: "replace", Type: "mixed"},
				{Name: "subject", Type: "mixed"},
			},
			ReturnType: "mixed",
			MinArgs:    3,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 3 {
					return values.NewString(""), nil
				}

				search := args[0].ToString()
				replace := args[1].ToString()
				subject := args[2].ToString()

				// Case-insensitive replacement
				// Use a simple approach: find all occurrences in lowercase, then replace in original
				lowerSubject := strings.ToLower(subject)
				lowerSearch := strings.ToLower(search)

				if search == "" {
					return values.NewString(subject), nil
				}

				result := ""
				start := 0

				for {
					idx := strings.Index(lowerSubject[start:], lowerSearch)
					if idx == -1 {
						result += subject[start:]
						break
					}

					actualIdx := start + idx
					result += subject[start:actualIdx] + replace
					start = actualIdx + len(search)

					if start >= len(subject) {
						break
					}
				}

				return values.NewString(result), nil
			},
		},
		{
			Name: "strcmp",
			Parameters: []*registry.Parameter{
				{Name: "str1", Type: "string"},
				{Name: "str2", Type: "string"},
			},
			ReturnType: "int",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 {
					return values.NewInt(0), nil
				}

				str1 := args[0].ToString()
				str2 := args[1].ToString()

				if str1 == str2 {
					return values.NewInt(0), nil
				} else if str1 < str2 {
					return values.NewInt(-1), nil
				} else {
					return values.NewInt(1), nil
				}
			},
		},
		{
			Name: "strcasecmp",
			Parameters: []*registry.Parameter{
				{Name: "str1", Type: "string"},
				{Name: "str2", Type: "string"},
			},
			ReturnType: "int",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 {
					return values.NewInt(0), nil
				}

				str1 := strings.ToLower(args[0].ToString())
				str2 := strings.ToLower(args[1].ToString())

				if str1 == str2 {
					return values.NewInt(0), nil
				} else if str1 < str2 {
					return values.NewInt(-1), nil
				} else {
					return values.NewInt(1), nil
				}
			},
		},
		{
			Name: "str_pad",
			Parameters: []*registry.Parameter{
				{Name: "input", Type: "string"},
				{Name: "pad_length", Type: "int"},
				{Name: "pad_string", Type: "string", HasDefault: true, DefaultValue: values.NewString(" ")},
				{Name: "pad_type", Type: "int", HasDefault: true, DefaultValue: values.NewInt(1)}, // STR_PAD_RIGHT
			},
			ReturnType: "string",
			MinArgs:    2,
			MaxArgs:    4,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 {
					return values.NewString(""), nil
				}

				input := args[0].ToString()
				padLength := int(args[1].ToInt())

				// If already long enough, return original
				if len(input) >= padLength {
					return values.NewString(input), nil
				}

				padString := " "
				if len(args) > 2 && args[2] != nil {
					padString = args[2].ToString()
					if padString == "" {
						padString = " "
					}
				}

				padType := int64(1) // STR_PAD_RIGHT
				if len(args) > 3 && args[3] != nil {
					padType = args[3].ToInt()
				}

				totalPadding := padLength - len(input)

				// Generate padding string
				padNeeded := ""
				for len(padNeeded) < totalPadding {
					padNeeded += padString
				}
				padNeeded = padNeeded[:totalPadding]

				var result string
				switch padType {
				case 0: // STR_PAD_LEFT
					result = padNeeded + input
				case 2: // STR_PAD_BOTH
					leftPad := totalPadding / 2
					rightPad := totalPadding - leftPad

					leftPadStr := ""
					for len(leftPadStr) < leftPad {
						leftPadStr += padString
					}
					leftPadStr = leftPadStr[:leftPad]

					rightPadStr := ""
					for len(rightPadStr) < rightPad {
						rightPadStr += padString
					}
					rightPadStr = rightPadStr[:rightPad]

					result = leftPadStr + input + rightPadStr
				default: // STR_PAD_RIGHT (1)
					result = input + padNeeded
				}

				return values.NewString(result), nil
			},
		},
		{
			Name: "strrev",
			Parameters: []*registry.Parameter{
				{Name: "string", Type: "string"},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewString(""), nil
				}

				str := args[0].ToString()
				if str == "" {
					return values.NewString(""), nil
				}

				// Convert to runes for proper Unicode handling
				runes := []rune(str)

				// Reverse the runes
				for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
					runes[i], runes[j] = runes[j], runes[i]
				}

				return values.NewString(string(runes)), nil
			},
		},
		{
			Name: "strstr",
			Parameters: []*registry.Parameter{
				{Name: "haystack", Type: "string"},
				{Name: "needle", Type: "string"},
				{Name: "before_needle", Type: "bool", DefaultValue: values.NewBool(false)},
			},
			ReturnType: "string|false",
			MinArgs:    2,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 {
					return values.NewBool(false), nil
				}

				haystack := args[0].ToString()
				needle := args[1].ToString()
				beforeNeedle := false
				if len(args) >= 3 && args[2] != nil {
					beforeNeedle = args[2].ToBool()
				}

				// Handle empty needle - PHP returns entire haystack
				if needle == "" {
					if beforeNeedle {
						return values.NewString(""), nil
					}
					return values.NewString(haystack), nil
				}

				// Handle empty haystack
				if haystack == "" {
					return values.NewBool(false), nil
				}

				// Find the first occurrence
				idx := strings.Index(haystack, needle)
				if idx == -1 {
					return values.NewBool(false), nil
				}

				if beforeNeedle {
					// Return the part before the needle
					return values.NewString(haystack[:idx]), nil
				} else {
					// Return the part from the needle to the end
					return values.NewString(haystack[idx:]), nil
				}
			},
		},
		{
			Name: "strchr",
			Parameters: []*registry.Parameter{
				{Name: "haystack", Type: "string"},
				{Name: "needle", Type: "string"},
				{Name: "before_needle", Type: "bool", DefaultValue: values.NewBool(false)},
			},
			ReturnType: "string|false",
			MinArgs:    2,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 {
					return values.NewBool(false), nil
				}

				haystack := args[0].ToString()
				needle := args[1].ToString()
				beforeNeedle := false
				if len(args) >= 3 && args[2] != nil {
					beforeNeedle = args[2].ToBool()
				}

				// Handle empty needle - PHP returns entire haystack
				if needle == "" {
					if beforeNeedle {
						return values.NewString(""), nil
					}
					return values.NewString(haystack), nil
				}

				// Handle empty haystack
				if haystack == "" {
					return values.NewBool(false), nil
				}

				// Find the first occurrence
				idx := strings.Index(haystack, needle)
				if idx == -1 {
					return values.NewBool(false), nil
				}

				if beforeNeedle {
					// Return the part before the needle
					return values.NewString(haystack[:idx]), nil
				} else {
					// Return the part from the needle to the end
					return values.NewString(haystack[idx:]), nil
				}
			},
		},
		{
			Name: "strrchr",
			Parameters: []*registry.Parameter{
				{Name: "haystack", Type: "string"},
				{Name: "needle", Type: "string"},
			},
			ReturnType: "string|false",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 {
					return values.NewBool(false), nil
				}

				haystack := args[0].ToString()
				needle := args[1].ToString()

				// Handle empty needle
				if needle == "" {
					return values.NewBool(false), nil
				}

				// Handle empty haystack
				if haystack == "" {
					return values.NewBool(false), nil
				}

				// strrchr only uses the first character of needle
				searchChar := string([]rune(needle)[0])

				// Find the last occurrence of the character
				idx := strings.LastIndex(haystack, searchChar)
				if idx == -1 {
					return values.NewBool(false), nil
				}

				// Return substring from the found position to the end
				return values.NewString(haystack[idx:]), nil
			},
		},
		{
			Name: "strtr",
			Parameters: []*registry.Parameter{
				{Name: "str", Type: "string"},
				{Name: "from", Type: "string"},
				{Name: "to", Type: "string"},
			},
			ReturnType: "string",
			MinArgs:    3,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 3 {
					return values.NewString(""), nil
				}

				str := args[0].ToString()
				from := args[1].ToString()
				to := args[2].ToString()

				// Handle empty string
				if str == "" {
					return values.NewString(""), nil
				}

				// Handle empty from or to
				if from == "" || to == "" {
					return values.NewString(str), nil
				}

				// Build character translation map
				translationMap := make(map[rune]rune)
				fromRunes := []rune(from)
				toRunes := []rune(to)

				// Map characters from 'from' to 'to'
				// If 'to' is shorter, ignore extra characters in 'from'
				minLen := len(fromRunes)
				if len(toRunes) < minLen {
					minLen = len(toRunes)
				}

				for i := 0; i < minLen; i++ {
					translationMap[fromRunes[i]] = toRunes[i]
				}

				// Apply translation
				result := make([]rune, 0, len(str))
				for _, r := range str {
					if replacement, exists := translationMap[r]; exists {
						result = append(result, replacement)
					} else {
						result = append(result, r)
					}
				}

				return values.NewString(string(result)), nil
			},
		},
		{
			Name: "str_split",
			Parameters: []*registry.Parameter{
				{Name: "string", Type: "string"},
				{Name: "length", Type: "int", DefaultValue: values.NewInt(1)},
			},
			ReturnType: "array",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 1 {
					return values.NewArray(), nil
				}

				str := args[0].ToString()
				chunkSize := int64(1)

				if len(args) >= 2 && args[1] != nil {
					chunkSize = args[1].ToInt()
				}

				// Validate chunk size
				if chunkSize <= 0 {
					return nil, fmt.Errorf("str_split(): Argument #2 ($length) must be greater than 0")
				}

				// Handle empty string
				if str == "" {
					return values.NewArray(), nil
				}

				// Create result array
				result := values.NewArray()
				resultArr := result.Data.(*values.Array)

				// Split string into chunks
				strRunes := []rune(str)
				index := int64(0)

				for i := 0; i < len(strRunes); i += int(chunkSize) {
					end := i + int(chunkSize)
					if end > len(strRunes) {
						end = len(strRunes)
					}

					chunk := string(strRunes[i:end])
					resultArr.Elements[index] = values.NewString(chunk)
					index++
				}

				return result, nil
			},
		},
		{
			Name: "chunk_split",
			Parameters: []*registry.Parameter{
				{Name: "string", Type: "string"},
				{Name: "length", Type: "int", DefaultValue: values.NewInt(76)},
				{Name: "ending", Type: "string", DefaultValue: values.NewString("\r\n")},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 1 {
					return values.NewString(""), nil
				}

				str := args[0].ToString()
				chunkLength := int64(76)
				ending := "\r\n"

				if len(args) >= 2 && args[1] != nil {
					chunkLength = args[1].ToInt()
				}

				if len(args) >= 3 && args[2] != nil {
					ending = args[2].ToString()
				}

				// Validate chunk length
				if chunkLength <= 0 {
					return nil, fmt.Errorf("chunk_split(): Argument #2 ($length) must be greater than 0")
				}

				// Handle empty ending - return original string
				if ending == "" {
					return values.NewString(str), nil
				}

				// Handle empty string
				if str == "" {
					return values.NewString(ending), nil
				}

				// Split string into chunks and add endings
				var result strings.Builder
				strRunes := []rune(str)

				for i := 0; i < len(strRunes); i += int(chunkLength) {
					end := i + int(chunkLength)
					if end > len(strRunes) {
						end = len(strRunes)
					}

					chunk := string(strRunes[i:end])
					result.WriteString(chunk)
					result.WriteString(ending)
				}

				return values.NewString(result.String()), nil
			},
		},
		{
			Name: "stristr",
			Parameters: []*registry.Parameter{
				{Name: "haystack", Type: "string"},
				{Name: "needle", Type: "string"},
				{Name: "before_needle", Type: "bool", DefaultValue: values.NewBool(false)},
			},
			ReturnType: "string|false",
			MinArgs:    2,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 {
					return values.NewBool(false), nil
				}

				haystack := args[0].ToString()
				needle := args[1].ToString()
				beforeNeedle := false
				if len(args) >= 3 && args[2] != nil {
					beforeNeedle = args[2].ToBool()
				}

				// Handle empty needle - PHP returns entire haystack
				if needle == "" {
					if beforeNeedle {
						return values.NewString(""), nil
					}
					return values.NewString(haystack), nil
				}

				// Handle empty haystack
				if haystack == "" {
					return values.NewBool(false), nil
				}

				// Find the first occurrence (case-insensitive)
				lowerHaystack := strings.ToLower(haystack)
				lowerNeedle := strings.ToLower(needle)
				idx := strings.Index(lowerHaystack, lowerNeedle)
				if idx == -1 {
					return values.NewBool(false), nil
				}

				if beforeNeedle {
					// Return the part before the needle (original case)
					return values.NewString(haystack[:idx]), nil
				} else {
					// Return the part from the needle to the end (original case)
					return values.NewString(haystack[idx:]), nil
				}
			},
		},
		{
			Name: "strripos",
			Parameters: []*registry.Parameter{
				{Name: "haystack", Type: "string"},
				{Name: "needle", Type: "string"},
			},
			ReturnType: "int|false",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 {
					return values.NewBool(false), nil
				}

				haystack := args[0].ToString()
				needle := args[1].ToString()

				// Handle empty needle - PHP returns haystack length
				if needle == "" {
					return values.NewInt(int64(len(haystack))), nil
				}

				// Handle empty haystack
				if haystack == "" {
					return values.NewBool(false), nil
				}

				// Find the last occurrence (case-insensitive)
				lowerHaystack := strings.ToLower(haystack)
				lowerNeedle := strings.ToLower(needle)
				idx := strings.LastIndex(lowerHaystack, lowerNeedle)
				if idx == -1 {
					return values.NewBool(false), nil
				}

				return values.NewInt(int64(idx)), nil
			},
		},
		{
			Name: "substr_replace",
			Parameters: []*registry.Parameter{
				{Name: "string", Type: "string"},
				{Name: "replacement", Type: "string"},
				{Name: "start", Type: "int"},
				{Name: "length", Type: "int", DefaultValue: nil},
			},
			ReturnType: "string",
			MinArgs:    3,
			MaxArgs:    4,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 3 {
					return values.NewString(""), nil
				}

				str := args[0].ToString()
				replacement := args[1].ToString()
				start := args[2].ToInt()

				strLen := int64(len(str))

				// Handle length parameter
				var length int64
				hasLength := len(args) >= 4 && args[3] != nil
				if hasLength {
					length = args[3].ToInt()
				} else {
					// If no length provided, replace to end of string
					length = strLen
				}

				// Handle negative start
				if start < 0 {
					start = strLen + start
					if start < 0 {
						start = 0
					}
				}

				// Handle negative length
				if hasLength && length < 0 {
					// Negative length means "length characters from the end"
					length = strLen - start + length
					if length < 0 {
						length = 0
					}
				}

				// Handle start beyond string length
				if start > strLen {
					// Append replacement at the end
					return values.NewString(str + replacement), nil
				}

				// Calculate end position
				end := start + length
				if end > strLen {
					end = strLen
				}

				// Build result: before + replacement + after
				var result strings.Builder

				// Add part before start
				if start > 0 {
					result.WriteString(str[:start])
				}

				// Add replacement
				result.WriteString(replacement)

				// Add part after end
				if end < strLen {
					result.WriteString(str[end:])
				}

				return values.NewString(result.String()), nil
			},
		},
		{
			Name: "strncmp",
			Parameters: []*registry.Parameter{
				{Name: "str1", Type: "string"},
				{Name: "str2", Type: "string"},
				{Name: "len", Type: "int"},
			},
			ReturnType: "int",
			MinArgs:    3,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 3 {
					return values.NewInt(0), nil
				}

				str1 := args[0].ToString()
				str2 := args[1].ToString()
				length := args[2].ToInt()

				// Handle zero or negative length
				if length <= 0 {
					return values.NewInt(0), nil
				}

				// Get string lengths
				str1Len := int64(len(str1))
				str2Len := int64(len(str2))

				// Calculate the actual comparison length
				compareLen := length
				if compareLen > str1Len {
					compareLen = str1Len
				}
				if compareLen > str2Len {
					compareLen = str2Len
				}

				// Extract substrings for comparison
				sub1 := str1[:compareLen]
				sub2 := str2[:compareLen]

				// Perform binary comparison of the substrings
				if sub1 < sub2 {
					return values.NewInt(-1), nil
				} else if sub1 > sub2 {
					return values.NewInt(1), nil
				}

				// If the compared parts are equal, check lengths
				// Only consider length difference if we're comparing beyond the shorter string
				if compareLen < length {
					// One or both strings are shorter than the requested length
					if str1Len < str2Len {
						return values.NewInt(-1), nil
					} else if str1Len > str2Len {
						return values.NewInt(1), nil
					}
				}

				return values.NewInt(0), nil
			},
		},
		{
			Name: "strncasecmp",
			Parameters: []*registry.Parameter{
				{Name: "str1", Type: "string"},
				{Name: "str2", Type: "string"},
				{Name: "len", Type: "int"},
			},
			ReturnType: "int",
			MinArgs:    3,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				str1 := args[0].Data.(string)
				str2 := args[1].Data.(string)
				length := int(args[2].Data.(int64))

				if length < 0 {
					return values.NewInt(0), nil
				}

				if length == 0 {
					return values.NewInt(0), nil
				}

				// Convert to lowercase for case-insensitive comparison
				lowerStr1 := strings.ToLower(str1)
				lowerStr2 := strings.ToLower(str2)

				str1Runes := []rune(lowerStr1)
				str2Runes := []rune(lowerStr2)
				str1Len := len(str1Runes)
				str2Len := len(str2Runes)

				// Calculate how many characters to actually compare
				compareLen := length
				if str1Len < compareLen {
					compareLen = str1Len
				}
				if str2Len < compareLen {
					compareLen = str2Len
				}

				// Extract substrings to compare
				sub1 := string(str1Runes[:compareLen])
				sub2 := string(str2Runes[:compareLen])

				// Perform binary comparison of the substrings
				if sub1 < sub2 {
					return values.NewInt(-1), nil
				} else if sub1 > sub2 {
					return values.NewInt(1), nil
				}

				// If the compared parts are equal, check lengths
				// Only consider length difference if we're comparing beyond the shorter string
				if compareLen < length {
					// One or both strings are shorter than the requested length
					if str1Len < str2Len {
						return values.NewInt(-1), nil
					} else if str1Len > str2Len {
						return values.NewInt(1), nil
					}
				}

				return values.NewInt(0), nil
			},
		},
		{
			Name: "str_contains",
			Parameters: []*registry.Parameter{
				{Name: "haystack", Type: "string"},
				{Name: "needle", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				haystack := args[0].Data.(string)
				needle := args[1].Data.(string)

				// PHP behavior: empty needle is always found
				if needle == "" {
					return values.NewBool(true), nil
				}

				// Check if haystack contains needle (case-sensitive)
				contains := strings.Contains(haystack, needle)
				return values.NewBool(contains), nil
			},
		},
		{
			Name: "str_starts_with",
			Parameters: []*registry.Parameter{
				{Name: "haystack", Type: "string"},
				{Name: "needle", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				haystack := args[0].Data.(string)
				needle := args[1].Data.(string)

				// PHP behavior: empty needle always matches
				if needle == "" {
					return values.NewBool(true), nil
				}

				// Check if haystack starts with needle (case-sensitive)
				startsWith := strings.HasPrefix(haystack, needle)
				return values.NewBool(startsWith), nil
			},
		},
		{
			Name: "str_ends_with",
			Parameters: []*registry.Parameter{
				{Name: "haystack", Type: "string"},
				{Name: "needle", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				haystack := args[0].Data.(string)
				needle := args[1].Data.(string)

				// PHP behavior: empty needle always matches
				if needle == "" {
					return values.NewBool(true), nil
				}

				// Check if haystack ends with needle (case-sensitive)
				endsWith := strings.HasSuffix(haystack, needle)
				return values.NewBool(endsWith), nil
			},
		},
		{
			Name: "str_word_count",
			Parameters: []*registry.Parameter{
				{Name: "string", Type: "string"},
				{Name: "format", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
				{Name: "charlist", Type: "string", HasDefault: true, DefaultValue: values.NewString("")},
			},
			ReturnType: "int|array",
			MinArgs:    1,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				str := args[0].Data.(string)
				format := int64(0)
				if len(args) > 1 {
					format = args[1].Data.(int64)
				}
				charlist := ""
				if len(args) > 2 {
					charlist = args[2].Data.(string)
				}

				// Convert charlist to a map for quick lookup
				extraChars := make(map[rune]bool)
				for _, char := range charlist {
					extraChars[char] = true
				}

				// Function to check if a character is a word character
				isWordChar := func(r rune) bool {
					return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || extraChars[r]
				}

				// Parse words from the string
				words := []string{}
				wordPositions := []int{}

				runes := []rune(str)
				i := 0
				for i < len(runes) {
					// Skip non-word characters
					for i < len(runes) && !isWordChar(runes[i]) {
						i++
					}

					if i >= len(runes) {
						break
					}

					// Found start of a word
					wordStart := i
					wordStartPos := 0

					// Calculate byte position of word start
					for j := 0; j < wordStart; j++ {
						wordStartPos += len(string(runes[j]))
					}

					// Continue until end of word
					wordRunes := []rune{}
					for i < len(runes) && isWordChar(runes[i]) {
						wordRunes = append(wordRunes, runes[i])
						i++
					}

					// Check if the word contains at least one letter
					hasLetter := false
					for _, r := range wordRunes {
						if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
							hasLetter = true
							break
						}
					}

					if hasLetter {
						words = append(words, string(wordRunes))
						wordPositions = append(wordPositions, wordStartPos)
					}
				}

				// Return based on format
				switch format {
				case 0:
					// Return word count
					return values.NewInt(int64(len(words))), nil
				case 1:
					// Return array of words (indexed from 0)
					result := values.NewArray()
					arr := result.Data.(*values.Array)
					for i, word := range words {
						arr.Elements[int64(i)] = values.NewString(word)
					}
					return result, nil
				case 2:
					// Return associative array with positions as keys
					result := values.NewArray()
					arr := result.Data.(*values.Array)
					for i, word := range words {
						pos := wordPositions[i]
						arr.Elements[int64(pos)] = values.NewString(word)
					}
					return result, nil
				default:
					return values.NewInt(int64(len(words))), nil
				}
			},
		},
		{
			Name: "htmlspecialchars",
			Parameters: []*registry.Parameter{
				{Name: "string", Type: "string"},
				{Name: "flags", Type: "int", HasDefault: true, DefaultValue: values.NewInt(3)}, // ENT_QUOTES is default
				{Name: "encoding", Type: "string", HasDefault: true, DefaultValue: values.NewString("UTF-8")},
				{Name: "double_encode", Type: "bool", HasDefault: true, DefaultValue: values.NewBool(true)},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    4,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				str := args[0].Data.(string)
				flags := int64(3) // ENT_QUOTES default
				if len(args) > 1 {
					flags = args[1].Data.(int64)
				}
				// encoding parameter - we'll assume UTF-8 for now
				doubleEncode := true
				if len(args) > 3 {
					doubleEncode = args[3].Data.(bool)
				}

				// PHP constants:
				// ENT_COMPAT = 2 (default) - convert double quotes only
				// ENT_QUOTES = 3 - convert both double and single quotes
				// ENT_NOQUOTES = 0 - don't convert quotes

				result := str

				// If double_encode is false, we need to avoid double-encoding existing entities
				if !doubleEncode {
					// First, temporarily replace existing entities with placeholders
					result = strings.ReplaceAll(result, "&amp;", "\x00AMP\x00")
					result = strings.ReplaceAll(result, "&lt;", "\x00LT\x00")
					result = strings.ReplaceAll(result, "&gt;", "\x00GT\x00")
					result = strings.ReplaceAll(result, "&quot;", "\x00QUOT\x00")
					result = strings.ReplaceAll(result, "&#039;", "\x00APOS\x00")
				}

				// Convert basic HTML special characters
				result = strings.ReplaceAll(result, "&", "&amp;")
				result = strings.ReplaceAll(result, "<", "&lt;")
				result = strings.ReplaceAll(result, ">", "&gt;")

				// Handle quotes based on flags
				switch flags {
				case 0: // ENT_NOQUOTES - don't convert quotes
					// Do nothing for quotes
				case 2: // ENT_COMPAT - convert double quotes only
					result = strings.ReplaceAll(result, "\"", "&quot;")
					// Single quotes are NOT converted in ENT_COMPAT mode
				case 3: // ENT_QUOTES - convert both double and single quotes (default)
					result = strings.ReplaceAll(result, "\"", "&quot;")
					result = strings.ReplaceAll(result, "'", "&#039;")
				default:
					// Default to ENT_QUOTES behavior
					result = strings.ReplaceAll(result, "\"", "&quot;")
					result = strings.ReplaceAll(result, "'", "&#039;")
				}

				// If double_encode is false, restore the original entities
				if !doubleEncode {
					result = strings.ReplaceAll(result, "\x00AMP\x00", "&amp;")
					result = strings.ReplaceAll(result, "\x00LT\x00", "&lt;")
					result = strings.ReplaceAll(result, "\x00GT\x00", "&gt;")
					result = strings.ReplaceAll(result, "\x00QUOT\x00", "&quot;")
					result = strings.ReplaceAll(result, "\x00APOS\x00", "&#039;")
				}

				return values.NewString(result), nil
			},
		},
		{
			Name: "urlencode",
			Parameters: []*registry.Parameter{
				{Name: "str", Type: "string"},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				str := args[0].Data.(string)

				// PHP urlencode follows application/x-www-form-urlencoded format
				// Safe characters: a-z A-Z 0-9 - _ .
				// Space becomes +
				// Everything else becomes %XX

				var result strings.Builder
				for _, r := range str {
					switch {
					case r == ' ':
						result.WriteByte('+')
					case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9'):
						result.WriteRune(r)
					case r == '-' || r == '_' || r == '.':
						result.WriteRune(r)
					default:
						// Convert to UTF-8 bytes and percent-encode each byte
						utf8Bytes := []byte(string(r))
						for _, b := range utf8Bytes {
							result.WriteString(fmt.Sprintf("%%%.2X", b))
						}
					}
				}

				return values.NewString(result.String()), nil
			},
		},

		// urldecode - Decode URL encoded string
		{
			Name: "urldecode",
			Parameters: []*registry.Parameter{
				{Name: "str", Type: "string"},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				str := args[0].Data.(string)

				var result strings.Builder
				runes := []rune(str)

				for i := 0; i < len(runes); i++ {
					char := runes[i]

					if char == '+' {
						// Convert plus signs to spaces (application/x-www-form-urlencoded)
						result.WriteRune(' ')
					} else if char == '%' && i+2 < len(runes) {
						// Handle percent-encoded sequences
						hex1 := runes[i+1]
						hex2 := runes[i+2]

						// Check if both characters are valid hex digits
						if isHexDigit(hex1) && isHexDigit(hex2) {
							// Convert hex to byte value
							byte1 := hexToNibble(hex1)
							byte2 := hexToNibble(hex2)
							byteVal := (byte1 << 4) | byte2

							// Collect this byte for UTF-8 reconstruction
							var utf8Bytes []byte
							utf8Bytes = append(utf8Bytes, byteVal)
							i += 2 // Skip the two hex digits

							// Look ahead for continuation bytes if this is a UTF-8 start byte
							if byteVal >= 0x80 {
								// Determine how many continuation bytes we need
								continuationCount := 0
								if byteVal >= 0xC0 && byteVal < 0xE0 {
									continuationCount = 1 // 2-byte UTF-8
								} else if byteVal >= 0xE0 && byteVal < 0xF0 {
									continuationCount = 2 // 3-byte UTF-8
								} else if byteVal >= 0xF0 && byteVal < 0xF8 {
									continuationCount = 3 // 4-byte UTF-8
								}

								// Collect continuation bytes
								for j := 0; j < continuationCount && i+3 < len(runes); j++ {
									if i+3 < len(runes) && runes[i+1] == '%' {
										nextHex1 := runes[i+2]
										nextHex2 := runes[i+3]
										if isHexDigit(nextHex1) && isHexDigit(nextHex2) {
											nextByte1 := hexToNibble(nextHex1)
											nextByte2 := hexToNibble(nextHex2)
											nextByteVal := (nextByte1 << 4) | nextByte2

											// Check if this is a valid UTF-8 continuation byte (0x80-0xBF)
											if nextByteVal >= 0x80 && nextByteVal <= 0xBF {
												utf8Bytes = append(utf8Bytes, nextByteVal)
												i += 3 // Skip %XX
											} else {
												break // Not a continuation byte
											}
										} else {
											break // Invalid hex
										}
									} else {
										break // No more percent sequences
									}
								}
							}

							// Convert collected bytes to UTF-8 string
							if utf8.Valid(utf8Bytes) {
								result.Write(utf8Bytes)
							} else {
								// If not valid UTF-8, just write the first byte as-is
								result.WriteByte(utf8Bytes[0])
								// Put back the continuation bytes we consumed
								for k := 1; k < len(utf8Bytes); k++ {
									i -= 3 // Back up to reprocess as individual bytes
								}
							}
						} else {
							// Invalid hex digits, keep the percent sign as-is
							result.WriteRune(char)
						}
					} else {
						// Regular character, keep as-is
						result.WriteRune(char)
					}
				}

				return values.NewString(result.String()), nil
			},
		},

		// base64_encode - Encode string with base64
		{
			Name: "base64_encode",
			Parameters: []*registry.Parameter{
				{Name: "str", Type: "string"},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				str := args[0].Data.(string)

				// Convert string to bytes (UTF-8 encoding)
				data := []byte(str)

				// Use Go's standard base64 encoding
				encoded := base64.StdEncoding.EncodeToString(data)

				return values.NewString(encoded), nil
			},
		},

		// base64_decode - Decode base64 data
		{
			Name: "base64_decode",
			Parameters: []*registry.Parameter{
				{Name: "str", Type: "string"},
				{Name: "strict", Type: "bool", HasDefault: true, DefaultValue: values.NewBool(false)},
			},
			ReturnType: "mixed",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				str := args[0].Data.(string)

				strict := false
				if len(args) > 1 {
					strict = args[1].Data.(bool)
				}

				// Use Go's standard base64 decoding
				decoded, err := base64.StdEncoding.DecodeString(str)

				if err != nil {
					if strict {
						// In strict mode, return false for invalid input
						return values.NewBool(false), nil
					} else {
						// In non-strict mode, attempt to decode what we can
						// First clean up the input by removing invalid characters
						cleanStr := ""
						for _, char := range str {
							if (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') ||
							   (char >= '0' && char <= '9') || char == '+' || char == '/' || char == '=' {
								cleanStr += string(char)
							}
						}

						// Try to fix padding issues
						if len(cleanStr) > 0 {
							// Remove existing padding
							cleanStr = strings.TrimRight(cleanStr, "=")

							// Add correct padding
							switch len(cleanStr) % 4 {
							case 2:
								cleanStr += "=="
							case 3:
								cleanStr += "="
							}
						}

						// Try decoding the fixed string
						decoded, err = base64.StdEncoding.DecodeString(cleanStr)
						if err != nil {
							// If still invalid, return empty string
							return values.NewString(""), nil
						}
					}
				}

				// Convert decoded bytes back to string (preserving UTF-8)
				return values.NewString(string(decoded)), nil
			},
		},

		// addslashes - Quote string with slashes
		{
			Name: "addslashes",
			Parameters: []*registry.Parameter{
				{Name: "str", Type: "string"},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				str := args[0].Data.(string)

				var result strings.Builder
				for _, char := range str {
					switch char {
					case '\'':
						result.WriteString("\\'")
					case '"':
						result.WriteString("\\\"")
					case '\\':
						result.WriteString("\\\\")
					case '\x00': // null byte
						result.WriteString("\\0")
					default:
						result.WriteRune(char)
					}
				}

				return values.NewString(result.String()), nil
			},
		},

		// stripslashes - Remove slashes from string
		{
			Name: "stripslashes",
			Parameters: []*registry.Parameter{
				{Name: "str", Type: "string"},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				str := args[0].Data.(string)

				var result strings.Builder
				runes := []rune(str)

				for i := 0; i < len(runes); i++ {
					if runes[i] == '\\' {
						if i+1 < len(runes) {
							nextChar := runes[i+1]
							switch nextChar {
							case '\'':
								result.WriteRune('\'')
								i++ // Skip the next character
							case '"':
								result.WriteRune('"')
								i++ // Skip the next character
							case '\\':
								result.WriteRune('\\')
								i++ // Skip the next character
							case '0':
								result.WriteByte(0) // null byte
								i++ // Skip the next character
							default:
								// For any other character after backslash, remove the backslash
								// but keep the character (orphaned backslash)
								result.WriteRune(nextChar)
								i++ // Skip the next character
							}
						} else {
							// Trailing backslash - just remove it (don't add to result)
						}
					} else {
						result.WriteRune(runes[i])
					}
				}

				return values.NewString(result.String()), nil
			},
		},

		// md5 - Calculate MD5 hash
		{
			Name: "md5",
			Parameters: []*registry.Parameter{
				{Name: "str", Type: "string"},
				{Name: "binary", Type: "bool", HasDefault: true, DefaultValue: values.NewBool(false)},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				str := args[0].Data.(string)

				binary := false
				if len(args) > 1 {
					binary = args[1].Data.(bool)
				}

				// Calculate MD5 hash
				hasher := md5.New()
				hasher.Write([]byte(str))
				hashBytes := hasher.Sum(nil)

				if binary {
					// Return raw binary (16 bytes)
					return values.NewString(string(hashBytes)), nil
				} else {
					// Return lowercase hexadecimal string (32 characters)
					return values.NewString(hex.EncodeToString(hashBytes)), nil
				}
			},
		},

		// sha1 - Calculate SHA1 hash
		{
			Name: "sha1",
			Parameters: []*registry.Parameter{
				{Name: "str", Type: "string"},
				{Name: "binary", Type: "bool", HasDefault: true, DefaultValue: values.NewBool(false)},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				str := args[0].Data.(string)

				binary := false
				if len(args) > 1 {
					binary = args[1].Data.(bool)
				}

				// Calculate SHA1 hash
				hasher := sha1.New()
				hasher.Write([]byte(str))
				hashBytes := hasher.Sum(nil)

				if binary {
					// Return raw binary (20 bytes)
					return values.NewString(string(hashBytes)), nil
				} else {
					// Return lowercase hexadecimal string (40 characters)
					return values.NewString(hex.EncodeToString(hashBytes)), nil
				}
			},
		},

		// number_format() - Format a number with grouped thousands
		{
			Name: "number_format",
			Parameters: []*registry.Parameter{
				{Name: "number", Type: "mixed"},
				{Name: "decimals", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
				{Name: "decimal_point", Type: "string", HasDefault: true, DefaultValue: values.NewString(".")},
				{Name: "thousands_separator", Type: "string", HasDefault: true, DefaultValue: values.NewString(",")},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    4,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Get the number value and convert to float64
				var numVal float64
				switch args[0].Type {
				case values.TypeInt:
					numVal = float64(args[0].Data.(int64))
				case values.TypeFloat:
					numVal = args[0].Data.(float64)
				case values.TypeString:
					var err error
					numVal, err = strconv.ParseFloat(args[0].Data.(string), 64)
					if err != nil {
						return values.NewString("0"), nil // PHP returns "0" for invalid numbers
					}
				default:
					return values.NewString("0"), nil
				}

				// Get optional parameters
				decimals := 0
				if len(args) > 1 {
					decimals = int(args[1].Data.(int64))
				}

				decimalPoint := "."
				if len(args) > 2 {
					decimalPoint = args[2].Data.(string)
				}

				thousandsSep := ","
				if len(args) > 3 {
					thousandsSep = args[3].Data.(string)
				}

				// Handle negative numbers
				negative := numVal < 0
				if negative {
					numVal = -numVal
				}

				// Round to specified decimal places
				multiplier := math.Pow(10, float64(decimals))
				numVal = math.Round(numVal*multiplier) / multiplier

				// Split into integer and decimal parts
				intPart := int64(numVal)
				decPart := numVal - float64(intPart)

				// Format integer part with thousands separators
				intStr := strconv.FormatInt(intPart, 10)
				if thousandsSep != "" && len(intStr) > 3 {
					// Add thousands separators from right to left
					var result []rune
					for i, r := range intStr {
						if i > 0 && (len(intStr)-i)%3 == 0 {
							result = append(result, []rune(thousandsSep)...)
						}
						result = append(result, r)
					}
					intStr = string(result)
				}

				// Format decimal part if needed
				var decStr string
				if decimals > 0 {
					// Convert decimal part to string with specified precision
					decPart = math.Round(decPart*multiplier) / multiplier
					decStr = fmt.Sprintf("%."+strconv.Itoa(decimals)+"f", decPart)
					// Remove the "0." prefix
					if len(decStr) > 2 {
						decStr = decStr[2:]
					} else {
						decStr = strings.Repeat("0", decimals)
					}
				}

				// Combine parts
				result := intStr
				if decimals > 0 {
					result += decimalPoint + decStr
				}

				// Add negative sign if needed
				if negative {
					result = "-" + result
				}

				return values.NewString(result), nil
			},
		},

		// htmlentities() - Convert all applicable characters to HTML entities
		{
			Name: "htmlentities",
			Parameters: []*registry.Parameter{
				{Name: "string", Type: "string"},
				{Name: "flags", Type: "int", HasDefault: true, DefaultValue: values.NewInt(3)}, // ENT_QUOTES default
				{Name: "encoding", Type: "string", HasDefault: true, DefaultValue: values.NewString("UTF-8")},
				{Name: "double_encode", Type: "bool", HasDefault: true, DefaultValue: values.NewBool(true)},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    4,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				str := args[0].Data.(string)

				// Get optional parameters
				flags := int64(3) // ENT_QUOTES default
				if len(args) > 1 {
					flags = args[1].Data.(int64)
				}

				// encoding := "UTF-8" // Not used in basic implementation
				if len(args) > 2 {
					_ = args[2].Data.(string) // encoding parameter (ignored for now)
				}

				doubleEncode := true
				if len(args) > 3 {
					doubleEncode = args[3].Data.(bool)
				}

				// Use proper double encoding handling
				result := processHTMLEntities(str, flags, doubleEncode)
				return values.NewString(result), nil
			},
		},

		// nl2br() - Insert HTML line breaks before all newlines
		{
			Name: "nl2br",
			Parameters: []*registry.Parameter{
				{Name: "string", Type: "string"},
				{Name: "is_xhtml", Type: "bool", HasDefault: true, DefaultValue: values.NewBool(true)},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				str := args[0].Data.(string)

				// Get XHTML parameter (default is true for XHTML mode)
				isXHTML := true
				if len(args) > 1 {
					isXHTML = args[1].Data.(bool)
				}

				// Choose the appropriate BR tag
				brTag := "<br />"
				if !isXHTML {
					brTag = "<br>"
				}

				// Convert string to runes for proper Unicode handling
				runes := []rune(str)
				var result strings.Builder
				result.Grow(len(str) * 2) // Pre-allocate for efficiency

				for i, r := range runes {
					if r == '\r' {
						// Handle \r\n (Windows) and \r (Mac) newlines
						if i+1 < len(runes) && runes[i+1] == '\n' {
							// \r\n sequence - add BR before \r\n
							result.WriteString(brTag)
							result.WriteRune(r) // Write \r
						} else {
							// Standalone \r - add BR before \r
							result.WriteString(brTag)
							result.WriteRune(r) // Write \r
						}
					} else if r == '\n' {
						// Handle \n (Unix) newlines
						// Check if this \n is part of \r\n sequence
						if i > 0 && runes[i-1] == '\r' {
							// This \n is part of \r\n, just write it
							result.WriteRune(r)
						} else {
							// Standalone \n - add BR before \n
							result.WriteString(brTag)
							result.WriteRune(r)
						}
					} else {
						// Regular character
						result.WriteRune(r)
					}
				}

				return values.NewString(result.String()), nil
			},
		},

		// str_rot13() - Apply ROT13 transformation to alphabetic characters
		{
			Name: "str_rot13",
			Parameters: []*registry.Parameter{
				{Name: "str", Type: "string"},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				str := args[0].Data.(string)

				// Convert string to runes for proper Unicode handling
				runes := []rune(str)
				var result strings.Builder
				result.Grow(len(str)) // Pre-allocate for efficiency

				for _, r := range runes {
					// Apply ROT13 transformation
					transformed := rot13Char(r)
					result.WriteRune(transformed)
				}

				return values.NewString(result.String()), nil
			},
		},

		// wordwrap() - Wrap text to specified line lengths
		{
			Name: "wordwrap",
			Parameters: []*registry.Parameter{
				{Name: "str", Type: "string"},
				{Name: "width", Type: "int", HasDefault: true, DefaultValue: values.NewInt(75)},
				{Name: "break", Type: "string", HasDefault: true, DefaultValue: values.NewString("\n")},
				{Name: "cut", Type: "bool", HasDefault: true, DefaultValue: values.NewBool(false)},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    4,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				str := args[0].Data.(string)

				// Get optional parameters
				width := int64(75)
				if len(args) > 1 {
					width = args[1].Data.(int64)
				}

				breakStr := "\n"
				if len(args) > 2 {
					breakStr = args[2].Data.(string)
				}

				cut := false
				if len(args) > 3 {
					cut = args[3].Data.(bool)
				}

				// Handle special width cases
				if width <= 0 {
					width = 1 // Treat as word boundaries
				}

				result := wordwrapText(str, int(width), breakStr, cut)
				return values.NewString(result), nil
			},
		},
		// html_entity_decode() - Decode HTML entities
		{
			Name: "html_entity_decode",
			Parameters: []*registry.Parameter{
				{Name: "string", Type: "string"},
				{Name: "flags", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)}, // ENT_COMPAT is default (0)
				{Name: "encoding", Type: "string", HasDefault: true, DefaultValue: values.NewString("UTF-8")},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				str := args[0].Data.(string)

				// Get optional parameters
				flags := int64(0) // ENT_COMPAT
				if len(args) > 1 {
					flags = args[1].Data.(int64)
				}

				encoding := "UTF-8"
				if len(args) > 2 {
					encoding = args[2].Data.(string)
				}
				_ = encoding // Ignore encoding for now, assume UTF-8

				result := htmlEntityDecodeString(str, flags)
				return values.NewString(result), nil
			},
		},
		// printf() - Output formatted string and return number of characters printed
		{
			Name: "printf",
			Parameters: []*registry.Parameter{
				{Name: "format", Type: "string"},
				// Additional arguments are handled dynamically
			},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    -1, // Variable arguments
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return values.NewInt(0), nil
				}

				format := args[0].ToString()
				var goArgs []interface{}

				// Convert PHP values to Go interface{} for fmt.Sprintf (same as sprintf)
				for i := 1; i < len(args); i++ {
					if args[i] == nil {
						goArgs = append(goArgs, nil)
					} else if args[i].IsInt() {
						goArgs = append(goArgs, args[i].ToInt())
					} else if args[i].IsFloat() {
						goArgs = append(goArgs, args[i].ToFloat())
					} else if args[i].IsBool() {
						goArgs = append(goArgs, args[i].ToBool())
					} else {
						goArgs = append(goArgs, args[i].ToString())
					}
				}

				// Use Go's fmt.Sprintf to format (same as sprintf)
				formatted := fmt.Sprintf(format, goArgs...)

				// Note: In a real implementation, this would write to stdout
				// For testing and simplicity, we skip the actual output
				// The formatted string would be output here

				// Return the number of characters that would be output
				return values.NewInt(int64(len(formatted))), nil
			},
		},
		// rawurlencode() - URL encode according to RFC 3986
		{
			Name: "rawurlencode",
			Parameters: []*registry.Parameter{
				{Name: "str", Type: "string"},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				str := args[0].Data.(string)
				result := rawUrlEncode(str)
				return values.NewString(result), nil
			},
		},
		// rawurldecode() - Decode URL encoded string according to RFC 3986
		{
			Name: "rawurldecode",
			Parameters: []*registry.Parameter{
				{Name: "str", Type: "string"},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				str := args[0].Data.(string)
				result := rawUrlDecode(str)
				return values.NewString(result), nil
			},
		},
		// crc32() - Calculate CRC32 checksum
		{
			Name: "crc32",
			Parameters: []*registry.Parameter{
				{Name: "str", Type: "string"},
			},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				str := args[0].Data.(string)

				// Calculate CRC32 using IEEE polynomial (standard CRC32)
				checksum := crc32.ChecksumIEEE([]byte(str))

				// PHP returns CRC32 as unsigned 32-bit value stored in int64
				// to handle values that exceed signed 32-bit range
				result := int64(checksum)

				return values.NewInt(result), nil
			},
		},
		{
			Name: "quotemeta",
			Parameters: []*registry.Parameter{
				{Name: "str", Type: "string"},
			},
			ReturnType: "string",
			MinArgs: 1, MaxArgs: 1, IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				str := args[0].Data.(string)
				result := quotemetaString(str)
				return values.NewString(result), nil
			},
		},
		{
			Name: "sscanf",
			Parameters: []*registry.Parameter{
				{Name: "str", Type: "string"},
				{Name: "format", Type: "string"},
			},
			ReturnType: "array",
			MinArgs: 2, MaxArgs: 2, IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				str := args[0].Data.(string)
				format := args[1].Data.(string)
				result := sscanfParse(str, format)
				return result, nil
			},
		},
		{
			Name: "str_shuffle",
			Parameters: []*registry.Parameter{
				{Name: "str", Type: "string"},
			},
			ReturnType: "string",
			MinArgs: 1, MaxArgs: 1, IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				str := args[0].Data.(string)
				shuffled := shuffleString(str)
				return values.NewString(shuffled), nil
			},
		},
		{
			Name: "parse_str",
			Parameters: []*registry.Parameter{
				{Name: "str", Type: "string"},
			},
			ReturnType: "array",
			MinArgs: 1, MaxArgs: 1, IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				str := args[0].Data.(string)
				result := parseQueryString(str)
				return result, nil
			},
		},
		{
			Name: "similar_text",
			Parameters: []*registry.Parameter{
				{Name: "first", Type: "string"},
				{Name: "second", Type: "string"},
			},
			ReturnType: "int",
			MinArgs: 2, MaxArgs: 2, IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				first := args[0].Data.(string)
				second := args[1].Data.(string)
				similarity := calculateSimilarText(first, second)
				return values.NewInt(int64(similarity)), nil
			},
		},
		{
			Name: "levenshtein",
			Parameters: []*registry.Parameter{
				{Name: "str1", Type: "string"},
				{Name: "str2", Type: "string"},
			},
			ReturnType: "int",
			MinArgs: 2, MaxArgs: 2, IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				str1 := args[0].Data.(string)
				str2 := args[1].Data.(string)
				distance := calculateLevenshteinDistance(str1, str2)
				return values.NewInt(int64(distance)), nil
			},
		},
		{
			Name: "hash",
			Parameters: []*registry.Parameter{
				{Name: "algo", Type: "string"},
				{Name: "data", Type: "string"},
			},
			ReturnType: "string",
			MinArgs: 2, MaxArgs: 2, IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				algo := args[0].Data.(string)
				data := args[1].Data.(string)
				result, err := calculateHash(algo, data)
				if err != nil {
					return nil, err
				}
				return values.NewString(result), nil
			},
		},
		{
			Name: "money_format",
			Parameters: []*registry.Parameter{
				{Name: "format", Type: "string"},
				{Name: "number", Type: "float"},
			},
			ReturnType: "string",
			MinArgs: 2, MaxArgs: 2, IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				format := args[0].Data.(string)
				number := args[1].ToFloat()
				result, err := formatMoney(format, number)
				if err != nil {
					return nil, err
				}
				return values.NewString(result), nil
			},
		},
		{
			Name: "mb_strlen",
			Parameters: []*registry.Parameter{
				{Name: "str", Type: "string"},
			},
			ReturnType: "int",
			MinArgs: 1, MaxArgs: 1, IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				str := args[0].Data.(string)
				length := utf8.RuneCountInString(str)
				return values.NewInt(int64(length)), nil
			},
		},
		{
			Name: "mb_substr",
			Parameters: []*registry.Parameter{
				{Name: "str", Type: "string"},
				{Name: "start", Type: "int"},
				{Name: "length", Type: "int", DefaultValue: values.NewNull()},
			},
			ReturnType: "string",
			MinArgs: 2, MaxArgs: 3, IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				str := args[0].Data.(string)
				start := int(args[1].Data.(int64))

				// Convert string to runes for proper Unicode handling
				runes := []rune(str)
				strLen := len(runes)

				// Handle negative start position
				if start < 0 {
					start = strLen + start
				}

				// If start is still negative or beyond string, return empty string
				if start < 0 || start >= strLen {
					return values.NewString(""), nil
				}

				// Determine end position
				var end int
				if len(args) == 3 && !args[2].IsNull() {
					length := int(args[2].Data.(int64))
					if length < 0 {
						// Negative length: from start to (end - |length|)
						end = strLen + length
						if end <= start {
							return values.NewString(""), nil
						}
					} else if length == 0 {
						return values.NewString(""), nil
					} else {
						// Positive length
						end = start + length
					}
				} else {
					// No length specified, go to end of string
					end = strLen
				}

				// Ensure end doesn't exceed string length
				if end > strLen {
					end = strLen
				}

				// Extract substring using rune slice
				if start >= end {
					return values.NewString(""), nil
				}

				result := string(runes[start:end])
				return values.NewString(result), nil
			},
		},
		{
			Name: "mb_strtolower",
			Parameters: []*registry.Parameter{
				{Name: "str", Type: "string"},
			},
			ReturnType: "string",
			MinArgs: 1, MaxArgs: 1, IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				str := args[0].Data.(string)

				// Convert string to runes for proper Unicode handling
				runes := []rune(str)
				result := make([]rune, len(runes))

				// Convert each rune to lowercase using Unicode rules
				for i, r := range runes {
					result[i] = unicode.ToLower(r)
				}

				return values.NewString(string(result)), nil
			},
		},
	}
}

// quotemetaString implements the quotemeta() function logic
// Escapes regex metacharacters with backslashes: . ^ $ * + ? [ ] ( ) \
// Does NOT escape: { } |
func quotemetaString(str string) string {
	if str == "" {
		return str
	}

	var result strings.Builder
	result.Grow(len(str) * 2) // Pre-allocate, assume worst case

	for _, char := range str {
		switch char {
		case '.', '^', '$', '*', '+', '?', '[', ']', '(', ')', '\\':
			result.WriteByte('\\')
			result.WriteRune(char)
		default:
			result.WriteRune(char)
		}
	}

	return result.String()
}

// shuffleString implements the str_shuffle() function logic
func shuffleString(str string) string {
	if str == "" {
		return str
	}

	// Convert to runes to handle Unicode characters properly
	runes := []rune(str)

	// Use current time as seed for better randomness
	rand.Seed(time.Now().UnixNano())

	// Fisher-Yates shuffle algorithm
	for i := len(runes) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}

// parseQueryString implements the parse_str() function logic
func parseQueryString(queryStr string) *values.Value {
	result := values.NewArray()
	resultArray := result.Data.(*values.Array)

	if queryStr == "" {
		return result
	}

	// Split by & and process each pair
	pairs := strings.Split(queryStr, "&")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		// Split by = to get key and value
		var key, value string
		if idx := strings.Index(pair, "="); idx >= 0 {
			key = pair[:idx]
			value = pair[idx+1:]
		} else {
			key = pair
			value = ""
		}

		// Skip empty keys
		if key == "" {
			continue
		}

		// URL decode key and value
		key = urlDecode(key)
		value = urlDecode(value)

		// Handle array notation (simplified)
		if strings.HasSuffix(key, "[]") {
			// Simple array: key[] = value
			arrayKey := key[:len(key)-2]
			if existing, exists := resultArray.Elements[arrayKey]; exists && existing.Type == values.TypeArray {
				// Append to existing array
				existingArray := existing.Data.(*values.Array)
				nextIndex := existingArray.NextIndex
				existingArray.Elements[nextIndex] = values.NewString(value)
				existingArray.NextIndex++
			} else {
				// Create new array
				newArray := values.NewArray()
				newArrayData := newArray.Data.(*values.Array)
				newArrayData.Elements[int64(0)] = values.NewString(value)
				newArrayData.NextIndex = 1
				resultArray.Elements[arrayKey] = newArray
			}
		} else {
			// Simple key-value pair
			resultArray.Elements[key] = values.NewString(value)
		}
	}

	return result
}

// urlDecode decodes URL-encoded strings (like PHP's parse_str)
func urlDecode(s string) string {
	// First replace + with spaces (application/x-www-form-urlencoded format)
	s = strings.ReplaceAll(s, "+", " ")

	// Then URL decode
	decoded, err := url.QueryUnescape(s)
	if err != nil {
		return s // Return original if decoding fails
	}
	return decoded
}

// calculateSimilarText implements the similar_text() algorithm
// This is a recursive algorithm that finds the longest common subsequence
func calculateSimilarText(first, second string) int {
	if first == "" || second == "" {
		return 0
	}

	if first == second {
		return len(first)
	}

	return similarTextRecursive(first, second, len(first), len(second))
}

// similarTextRecursive implements the core recursive algorithm used by PHP's similar_text
func similarTextRecursive(first, second string, firstLen, secondLen int) int {
	if firstLen == 0 || secondLen == 0 {
		return 0
	}

	var max, pos1, pos2 int

	// Find the longest common substring
	for i := 0; i < firstLen; i++ {
		for j := 0; j < secondLen; j++ {
			// Count matching characters starting from position i, j
			l := 0
			for k := 0; i+k < firstLen && j+k < secondLen && first[i+k] == second[j+k]; k++ {
				l++
			}

			// Keep track of the longest match
			if l > max {
				max = l
				pos1 = i
				pos2 = j
			}
		}
	}

	if max == 0 {
		return 0
	}

	// Recursively calculate similarity for the parts before and after the common substring
	before := similarTextRecursive(first[:pos1], second[:pos2], pos1, pos2)
	after := similarTextRecursive(first[pos1+max:], second[pos2+max:], firstLen-pos1-max, secondLen-pos2-max)

	return max + before + after
}

// sscanfParse implements the sscanf() function logic
// This is a simplified implementation supporting basic format specifiers
func sscanfParse(str, format string) *values.Value {
	// Handle empty string special case
	if strings.TrimSpace(str) == "" {
		return values.NewNull()
	}

	// Parse format string to extract specifiers
	specifiers := parseFormatSpecifiers(format)
	if len(specifiers) == 0 {
		return values.NewNull()
	}

	// Try matching with progressively shorter patterns to handle partial matches
	matches, matchCount := findBestScanfMatch(str, specifiers)
	hasAnyMatch := matchCount > 0

	// Convert matches to appropriate types
	result := values.NewArray()
	resultArray := result.Data.(*values.Array)
	matchIndex := 1 // Skip full match at index 0
	outputIndex := int64(0)

	// Count non-suppressed specifiers to create proper array size
	nonSuppressedCount := 0
	for _, spec := range specifiers {
		if !spec.suppress {
			nonSuppressedCount++
		}
	}

	for _, spec := range specifiers {
		if spec.suppress {
			// Skip suppressed assignments
			if hasAnyMatch && matchIndex < len(matches) {
				matchIndex++
			}
			continue
		}

		var val *values.Value
		if hasAnyMatch && matchIndex < len(matches) && matches[matchIndex] != "" {
			val = convertScanfValue(matches[matchIndex], spec.specifier)
		} else {
			val = values.NewNull()
		}

		resultArray.Elements[outputIndex] = val
		outputIndex++
		if hasAnyMatch {
			matchIndex++
		}
	}

	resultArray.NextIndex = outputIndex
	return result
}

// formatSpecifier represents a scanf format specifier
type formatSpecifier struct {
	specifier rune // d, s, f, c, x, o
	width     int  // field width
	suppress  bool // assignment suppression (*)
}

// parseFormatSpecifiers extracts format specifiers from format string
func parseFormatSpecifiers(format string) []formatSpecifier {
	var specs []formatSpecifier
	runes := []rune(format)

	for i := 0; i < len(runes); i++ {
		if runes[i] == '%' && i+1 < len(runes) {
			spec := formatSpecifier{}
			i++ // skip %

			// Check for assignment suppression
			if i < len(runes) && runes[i] == '*' {
				spec.suppress = true
				i++
			}

			// Parse width
			widthStart := i
			for i < len(runes) && unicode.IsDigit(runes[i]) {
				i++
			}
			if i > widthStart {
				if width, err := strconv.Atoi(string(runes[widthStart:i])); err == nil {
					spec.width = width
				}
			}

			// Get specifier character
			if i < len(runes) {
				spec.specifier = runes[i]
				specs = append(specs, spec)
			}
		}
	}

	return specs
}

// buildScanfRegex builds a regex pattern from format string and specifiers
func buildScanfRegex(format string, specs []formatSpecifier) string {
	pattern := regexp.QuoteMeta(format)
	specIndex := 0

	// Replace %... patterns with appropriate regex groups
	for specIndex < len(specs) {
		spec := specs[specIndex]
		specIndex++

		// Build the original pattern to replace
		originalPattern := "%"
		if spec.suppress {
			originalPattern += "\\*"
		}
		if spec.width > 0 {
			originalPattern += strconv.Itoa(spec.width)
		}
		originalPattern += string(spec.specifier)

		// Build replacement regex based on specifier
		var replacement string
		switch spec.specifier {
		case 'd':
			if spec.width > 0 {
				replacement = fmt.Sprintf("(-?\\d{1,%d})", spec.width)
			} else {
				replacement = "(-?\\d+)"
			}
		case 'f':
			replacement = "(-?\\d+(?:\\.\\d+)?)"
		case 's':
			if spec.width > 0 {
				replacement = fmt.Sprintf("(\\S{1,%d})", spec.width)
			} else {
				replacement = "(\\S+)"
			}
		case 'c':
			replacement = "(.)"
		case 'x':
			if spec.width > 0 {
				replacement = fmt.Sprintf("([0-9a-fA-F]{1,%d})", spec.width)
			} else {
				replacement = "([0-9a-fA-F]+)"
			}
		case 'o':
			if spec.width > 0 {
				replacement = fmt.Sprintf("([0-7]{1,%d})", spec.width)
			} else {
				replacement = "([0-7]+)"
			}
		default:
			replacement = "(\\S+)" // fallback
		}

		// Replace first occurrence
		pattern = strings.Replace(pattern, originalPattern, replacement, 1)
	}

	// Handle whitespace in pattern - replace literal spaces with flexible whitespace
	pattern = regexp.MustCompile(`\s+`).ReplaceAllString(pattern, `\\s*`)

	// Make the entire pattern match partial input by making remainder optional
	// This allows for partial matches like "123" matching "%d %d" for the first part
	pattern = "^" + pattern

	return pattern
}

// findBestScanfMatch tries to find the best possible match for scanf
// Returns matches and the count of successful matches
func findBestScanfMatch(str string, specs []formatSpecifier) ([]string, int) {
	// First try the full pattern
	fullFormat := buildPartialFormat(specs)
	pattern := buildScanfRegex(fullFormat, specs)

	if re, err := regexp.Compile(pattern); err == nil {
		if matches := re.FindStringSubmatch(str); matches != nil && len(matches) >= 2 {
			return matches, len(specs)
		}
	}

	// If full pattern fails, try progressively shorter patterns
	for numSpecs := len(specs) - 1; numSpecs >= 1; numSpecs-- {
		partialSpecs := specs[:numSpecs]
		format := buildPartialFormat(partialSpecs)
		pattern := buildScanfRegex(format, partialSpecs)

		if re, err := regexp.Compile(pattern); err == nil {
			if matches := re.FindStringSubmatch(str); matches != nil && len(matches) >= 2 {
				return matches, numSpecs
			}
		}
	}
	return nil, 0
}

// buildPartialFormat builds a format string from partial specifiers
func buildPartialFormat(specs []formatSpecifier) string {
	var format strings.Builder
	for i, spec := range specs {
		if i > 0 {
			format.WriteByte(' ') // Add space between specifiers
		}
		format.WriteByte('%')
		if spec.suppress {
			format.WriteByte('*')
		}
		if spec.width > 0 {
			format.WriteString(strconv.Itoa(spec.width))
		}
		format.WriteRune(spec.specifier)
	}
	return format.String()
}

// convertScanfValue converts a string match to appropriate Value type
func convertScanfValue(match string, specifier rune) *values.Value {
	switch specifier {
	case 'd':
		if val, err := strconv.ParseInt(match, 10, 64); err == nil {
			return values.NewInt(val)
		}
		return values.NewNull()
	case 'f':
		if val, err := strconv.ParseFloat(match, 64); err == nil {
			return values.NewFloat(val)
		}
		return values.NewNull()
	case 's', 'c':
		return values.NewString(match)
	case 'x':
		if val, err := strconv.ParseInt(match, 16, 64); err == nil {
			return values.NewInt(val)
		}
		return values.NewNull()
	case 'o':
		if val, err := strconv.ParseInt(match, 8, 64); err == nil {
			return values.NewInt(val)
		}
		return values.NewNull()
	default:
		return values.NewString(match)
	}
}

// wordwrapText implements the word wrapping logic
func wordwrapText(text string, width int, breakStr string, cut bool) string {
	if text == "" {
		return text
	}

	// Split text by existing newlines first and preserve them
	lines := strings.Split(text, "\n")
	var wrappedLines []string

	for _, line := range lines {
		if line == "" {
			wrappedLines = append(wrappedLines, "")
			continue
		}
		wrappedLine := wordwrapLine(line, width, breakStr, cut)
		wrappedLines = append(wrappedLines, wrappedLine)
	}

	return strings.Join(wrappedLines, "\n")
}

// wordwrapLine wraps a single line using PHP-compatible logic
func wordwrapLine(line string, width int, breakStr string, cut bool) string {
	if len(line) <= width {
		return line
	}

	var result []string
	currentPos := 0

	for currentPos < len(line) {
		// If remaining text fits in width, take it all
		if len(line)-currentPos <= width {
			result = append(result, line[currentPos:])
			break
		}

		// Find the best break point within width
		endPos := currentPos + width
		breakPos := -1

		// Look backwards from endPos to find a space (not tab)
		for i := endPos; i > currentPos; i-- {
			if i < len(line) && line[i] == ' ' {
				breakPos = i
				break
			}
		}

		if breakPos > currentPos {
			// Found a good break point (space)
			result = append(result, line[currentPos:breakPos])
			// Skip the space
			currentPos = breakPos + 1
		} else if cut {
			// No space found, force cut at width
			result = append(result, line[currentPos:currentPos+width])
			currentPos += width
		} else {
			// No space found, look for next space (not tab)
			nextSpace := -1
			for i := currentPos + width; i < len(line); i++ {
				if line[i] == ' ' {
					nextSpace = i
					break
				}
			}

			if nextSpace >= 0 {
				result = append(result, line[currentPos:nextSpace])
				currentPos = nextSpace + 1
			} else {
				// No space found, take the rest
				result = append(result, line[currentPos:])
				break
			}
		}
	}

	return strings.Join(result, breakStr)
}

// rot13Char applies ROT13 transformation to a single character
func rot13Char(r rune) rune {
	// Only transform ASCII letters A-Z and a-z
	if r >= 'A' && r <= 'Z' {
		// Uppercase letters: A=65, Z=90
		// Shift by 13, wrapping around at Z
		return 'A' + (r-'A'+13)%26
	} else if r >= 'a' && r <= 'z' {
		// Lowercase letters: a=97, z=122
		// Shift by 13, wrapping around at z
		return 'a' + (r-'a'+13)%26
	}
	// Non-alphabetic characters remain unchanged
	return r
}

// getHTMLEntity returns the HTML entity for a rune, or the rune itself if no entity exists
func getHTMLEntity(r rune, flags int64, doubleEncode bool) string {
	// Check if we need to avoid double encoding
	if !doubleEncode {
		// Simple check for existing entities (this is a basic implementation)
		runeStr := string(r)
		if r == '&' && len(runeStr) > 1 {
			// This is a simplified check - a full implementation would parse entities
			return runeStr
		}
	}

	// Handle quote flags
	// ENT_NOQUOTES = 0, ENT_COMPAT = 2, ENT_QUOTES = 3
	switch r {
	case '"':
		if flags == 0 { // ENT_NOQUOTES
			return `"`
		}
		return "&quot;"
	case '\'':
		if flags == 0 || flags == 2 { // ENT_NOQUOTES or ENT_COMPAT
			return "'"
		}
		return "&#039;"
	}

	// Check HTML entity map
	if entity, exists := htmlEntityMap[r]; exists {
		return entity
	}

	// Return the original character if no entity exists
	return string(r)
}

// processHTMLEntities handles the full string conversion with proper double encoding logic
func processHTMLEntities(str string, flags int64, doubleEncode bool) string {
	if doubleEncode {
		// Simple case: convert everything
		runes := []rune(str)
		var result strings.Builder
		result.Grow(len(str) * 2)

		for _, r := range runes {
			entity := getHTMLEntitySimple(r, flags)
			result.WriteString(entity)
		}
		return result.String()
	}

	// Complex case: avoid double encoding existing entities
	runes := []rune(str)
	var result strings.Builder
	result.Grow(len(str) * 2)

	for i, r := range runes {
		if r == '&' {
			// Look ahead to see if this might be an existing entity
			remaining := string(runes[i:])
			if isExistingEntity(remaining) {
				// It's already an entity, keep it as-is
				result.WriteRune(r)
			} else {
				// Not an entity, encode it
				result.WriteString("&amp;")
			}
		} else {
			entity := getHTMLEntitySimple(r, flags)
			result.WriteString(entity)
		}
	}
	return result.String()
}

// getHTMLEntitySimple handles entity conversion without double encoding logic
func getHTMLEntitySimple(r rune, flags int64) string {
	// Handle quote flags
	switch r {
	case '"':
		if flags == 0 { // ENT_NOQUOTES
			return `"`
		}
		return "&quot;"
	case '\'':
		if flags == 0 || flags == 2 { // ENT_NOQUOTES or ENT_COMPAT
			return "'"
		}
		return "&#039;"
	case '&':
		return "&amp;"
	}

	// Check HTML entity map
	if entity, exists := htmlEntityMap[r]; exists {
		return entity
	}

	// Return the original character if no entity exists
	return string(r)
}

// isExistingEntity checks if the string starts with a known HTML entity
func isExistingEntity(str string) bool {
	if len(str) < 4 { // Minimum entity length is &lt;
		return false
	}

	// Look for common entity patterns
	commonEntities := []string{"&lt;", "&gt;", "&amp;", "&quot;", "&#039;"}
	for _, entity := range commonEntities {
		if strings.HasPrefix(str, entity) {
			return true
		}
	}
	return false
}

// htmlEntityMap contains the mapping of Unicode characters to HTML entities
var htmlEntityMap = map[rune]string{
	// Basic HTML entities
	'<':  "&lt;",
	'>':  "&gt;",
	'&':  "&amp;",

	// Latin-1 Supplement (160-255)
	'\u00A0': "&nbsp;",   // Non-breaking space
	'\u00A1': "&iexcl;",  // ¡
	'\u00A2': "&cent;",   // ¢
	'\u00A3': "&pound;",  // £
	'\u00A4': "&curren;", // ¤
	'\u00A5': "&yen;",    // ¥
	'\u00A6': "&brvbar;", // ¦
	'\u00A7': "&sect;",   // §
	'\u00A8': "&uml;",    // ¨
	'\u00A9': "&copy;",   // ©
	'\u00AA': "&ordf;",   // ª
	'\u00AB': "&laquo;",  // «
	'\u00AC': "&not;",    // ¬
	'\u00AD': "&shy;",    // Soft hyphen
	'\u00AE': "&reg;",    // ®
	'\u00AF': "&macr;",   // ¯
	'\u00B0': "&deg;",    // °
	'\u00B1': "&plusmn;", // ±
	'\u00B2': "&sup2;",   // ²
	'\u00B3': "&sup3;",   // ³
	'\u00B4': "&acute;",  // ´
	'\u00B5': "&micro;",  // µ
	'\u00B6': "&para;",   // ¶
	'\u00B7': "&middot;", // ·
	'\u00B8': "&cedil;",  // ¸
	'\u00B9': "&sup1;",   // ¹
	'\u00BA': "&ordm;",   // º
	'\u00BB': "&raquo;",  // »
	'\u00BC': "&frac14;", // ¼
	'\u00BD': "&frac12;", // ½
	'\u00BE': "&frac34;", // ¾
	'\u00BF': "&iquest;", // ¿

	// Latin-1 uppercase letters
	'\u00C0': "&Agrave;", // À
	'\u00C1': "&Aacute;", // Á
	'\u00C2': "&Acirc;",  // Â
	'\u00C3': "&Atilde;", // Ã
	'\u00C4': "&Auml;",   // Ä
	'\u00C5': "&Aring;",  // Å
	'\u00C6': "&AElig;",  // Æ
	'\u00C7': "&Ccedil;", // Ç
	'\u00C8': "&Egrave;", // È
	'\u00C9': "&Eacute;", // É
	'\u00CA': "&Ecirc;",  // Ê
	'\u00CB': "&Euml;",   // Ë
	'\u00CC': "&Igrave;", // Ì
	'\u00CD': "&Iacute;", // Í
	'\u00CE': "&Icirc;",  // Î
	'\u00CF': "&Iuml;",   // Ï
	'\u00D0': "&ETH;",    // Ð
	'\u00D1': "&Ntilde;", // Ñ
	'\u00D2': "&Ograve;", // Ò
	'\u00D3': "&Oacute;", // Ó
	'\u00D4': "&Ocirc;",  // Ô
	'\u00D5': "&Otilde;", // Õ
	'\u00D6': "&Ouml;",   // Ö
	'\u00D7': "&times;",  // ×
	'\u00D8': "&Oslash;", // Ø
	'\u00D9': "&Ugrave;", // Ù
	'\u00DA': "&Uacute;", // Ú
	'\u00DB': "&Ucirc;",  // Û
	'\u00DC': "&Uuml;",   // Ü
	'\u00DD': "&Yacute;", // Ý
	'\u00DE': "&THORN;",  // Þ
	'\u00DF': "&szlig;",  // ß

	// Latin-1 lowercase letters
	'\u00E0': "&agrave;", // à
	'\u00E1': "&aacute;", // á
	'\u00E2': "&acirc;",  // â
	'\u00E3': "&atilde;", // ã
	'\u00E4': "&auml;",   // ä
	'\u00E5': "&aring;",  // å
	'\u00E6': "&aelig;",  // æ
	'\u00E7': "&ccedil;", // ç
	'\u00E8': "&egrave;", // è
	'\u00E9': "&eacute;", // é
	'\u00EA': "&ecirc;",  // ê
	'\u00EB': "&euml;",   // ë
	'\u00EC': "&igrave;", // ì
	'\u00ED': "&iacute;", // í
	'\u00EE': "&icirc;",  // î
	'\u00EF': "&iuml;",   // ï
	'\u00F0': "&eth;",    // ð
	'\u00F1': "&ntilde;", // ñ
	'\u00F2': "&ograve;", // ò
	'\u00F3': "&oacute;", // ó
	'\u00F4': "&ocirc;",  // ô
	'\u00F5': "&otilde;", // õ
	'\u00F6': "&ouml;",   // ö
	'\u00F7': "&divide;", // ÷
	'\u00F8': "&oslash;", // ø
	'\u00F9': "&ugrave;", // ù
	'\u00FA': "&uacute;", // ú
	'\u00FB': "&ucirc;",  // û
	'\u00FC': "&uuml;",   // ü
	'\u00FD': "&yacute;", // ý
	'\u00FE': "&thorn;",  // þ
	'\u00FF': "&yuml;",   // ÿ

	// Mathematical operators
	'\u2212': "&minus;",  // −
	'\u2264': "&le;",     // ≤
	'\u2265': "&ge;",     // ≥
	'\u2260': "&ne;",     // ≠
	'\u2248': "&asymp;",  // ≈
	'\u221E': "&infin;",  // ∞
	'\u2200': "&forall;", // ∀
	'\u2202': "&part;",   // ∂
	'\u2203': "&exist;",  // ∃
	'\u2207': "&nabla;",  // ∇
	'\u2208': "&isin;",   // ∈
	'\u2209': "&notin;",  // ∉
	'\u220B': "&ni;",     // ∋
	'\u220F': "&prod;",   // ∏
	'\u2211': "&sum;",    // ∑

	// Greek letters lowercase
	'\u03B1': "&alpha;",   // α
	'\u03B2': "&beta;",    // β
	'\u03B3': "&gamma;",   // γ
	'\u03B4': "&delta;",   // δ
	'\u03B5': "&epsilon;", // ε
	'\u03B6': "&zeta;",    // ζ
	'\u03B7': "&eta;",     // η
	'\u03B8': "&theta;",   // θ
	'\u03B9': "&iota;",    // ι
	'\u03BA': "&kappa;",   // κ
	'\u03BB': "&lambda;",  // λ
	'\u03BC': "&mu;",      // μ
	'\u03BD': "&nu;",      // ν
	'\u03BE': "&xi;",      // ξ
	'\u03BF': "&omicron;", // ο
	'\u03C0': "&pi;",      // π
	'\u03C1': "&rho;",     // ρ
	'\u03C2': "&sigmaf;",  // ς
	'\u03C3': "&sigma;",   // σ
	'\u03C4': "&tau;",     // τ
	'\u03C5': "&upsilon;", // υ
	'\u03C6': "&phi;",     // φ
	'\u03C7': "&chi;",     // χ
	'\u03C8': "&psi;",     // ψ
	'\u03C9': "&omega;",   // ω

	// Greek letters uppercase
	'\u0391': "&Alpha;",   // Α
	'\u0392': "&Beta;",    // Β
	'\u0393': "&Gamma;",   // Γ
	'\u0394': "&Delta;",   // Δ
	'\u0395': "&Epsilon;", // Ε
	'\u0396': "&Zeta;",    // Ζ
	'\u0397': "&Eta;",     // Η
	'\u0398': "&Theta;",   // Θ
	'\u0399': "&Iota;",    // Ι
	'\u039A': "&Kappa;",   // Κ
	'\u039B': "&Lambda;",  // Λ
	'\u039C': "&Mu;",      // Μ
	'\u039D': "&Nu;",      // Ν
	'\u039E': "&Xi;",      // Ξ
	'\u039F': "&Omicron;", // Ο
	'\u03A0': "&Pi;",      // Π
	'\u03A1': "&Rho;",     // Ρ
	'\u03A3': "&Sigma;",   // Σ
	'\u03A4': "&Tau;",     // Τ
	'\u03A5': "&Upsilon;", // Υ
	'\u03A6': "&Phi;",     // Φ
	'\u03A7': "&Chi;",     // Χ
	'\u03A8': "&Psi;",     // Ψ
	'\u03A9': "&Omega;",   // Ω

	// Special punctuation and symbols
	'\u2020': "&dagger;",  // †
	'\u2021': "&Dagger;",  // ‡
	'\u2022': "&bull;",    // •
	'\u2026': "&hellip;",  // …
	'\u2030': "&permil;",  // ‰
	'\u2032': "&prime;",   // ′
	'\u2033': "&Prime;",   // ″
	'\u2039': "&lsaquo;",  // ‹
	'\u203A': "&rsaquo;",  // ›
	'\u20AC': "&euro;",    // €
	'\u2122': "&trade;",   // ™

	// Card suits
	'\u2660': "&spades;", // ♠
	'\u2663': "&clubs;",  // ♣
	'\u2665': "&hearts;", // ♥
	'\u2666': "&diams;",  // ♦
}

// Helper function to check if a rune is a valid hex digit
func isHexDigit(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'A' && r <= 'F') || (r >= 'a' && r <= 'f')
}

// Helper function to convert hex digit to nibble value
func hexToNibble(r rune) byte {
	if r >= '0' && r <= '9' {
		return byte(r - '0')
	} else if r >= 'A' && r <= 'F' {
		return byte(r - 'A' + 10)
	} else if r >= 'a' && r <= 'f' {
		return byte(r - 'a' + 10)
	}
	return 0
}

// htmlEntityDecodeString decodes HTML entities in a string
func htmlEntityDecodeString(str string, flags int64) string {
	if str == "" {
		return str
	}

	var result strings.Builder
	result.Grow(len(str))

	runes := []rune(str)
	for i := 0; i < len(runes); i++ {
		if runes[i] == '&' {
			// Look for entity
			entityEnd := -1
			for j := i + 1; j < len(runes) && j < i+10; j++ { // Max entity length around 10
				if runes[j] == ';' {
					entityEnd = j
					break
				}
				if runes[j] == '&' || runes[j] == ' ' {
					break // Invalid entity
				}
			}

			if entityEnd > i+1 {
				entity := string(runes[i:entityEnd+1])
				decoded := decodeHTMLEntity(entity, flags)
				if decoded != entity {
					// Successfully decoded
					result.WriteString(decoded)
					i = entityEnd
					continue
				}
			}
		}

		// Not an entity or failed to decode, keep original character
		result.WriteRune(runes[i])
	}

	return result.String()
}

// decodeHTMLEntity decodes a single HTML entity
func decodeHTMLEntity(entity string, flags int64) string {
	if len(entity) < 3 || entity[0] != '&' || entity[len(entity)-1] != ';' {
		return entity
	}

	entityBody := entity[1 : len(entity)-1]

	// Handle numeric entities
	if len(entityBody) > 0 && entityBody[0] == '#' {
		if len(entityBody) == 1 {
			return entity
		}

		numStr := entityBody[1:]
		var codePoint int64
		var err error

		if len(numStr) > 1 && (numStr[0] == 'x' || numStr[0] == 'X') {
			// Hexadecimal entity
			codePoint, err = strconv.ParseInt(numStr[1:], 16, 32)
		} else {
			// Decimal entity
			codePoint, err = strconv.ParseInt(numStr, 10, 32)
		}

		if err == nil && codePoint >= 0 && codePoint <= 0x10FFFF {
			return string(rune(codePoint))
		}
		return entity
	}

	// Handle named entities
	switch entityBody {
	case "amp":
		return "&"
	case "lt":
		return "<"
	case "gt":
		return ">"
	case "quot":
		return "\""
	case "apos":
		// apos is only decoded with ENT_QUOTES flag (not in PHP's default ENT_COMPAT)
		return entity // Default behavior doesn't decode apos
	case "nbsp":
		return "\u00a0"
	case "copy":
		return "©"
	case "reg":
		return "®"
	case "trade":
		return "™"
	case "euro":
		return "€"
	case "hellip":
		return "…"
	case "mdash":
		return "—"
	case "ndash":
		return "–"
	case "laquo":
		return "«"
	case "raquo":
		return "»"
	case "aacute":
		return "á"
	case "agrave":
		return "à"
	case "acirc":
		return "â"
	case "atilde":
		return "ã"
	case "auml":
		return "ä"
	case "aring":
		return "å"
	case "ccedil":
		return "ç"
	case "szlig":
		return "ß"
	}

	// Check against the reverse entity map
	if char, exists := htmlDecodeMap[entityBody]; exists {
		return string(char)
	}

	// Unknown entity, return as-is
	return entity
}

// htmlDecodeMap maps entity names to their Unicode characters (reverse of htmlEntityMap)
var htmlDecodeMap = map[string]rune{
	// This would be populated with the reverse mapping of htmlEntityMap
	// For now, we handle the most common ones in the switch statement above
}

// rawUrlEncode encodes a string according to RFC 3986
func rawUrlEncode(str string) string {
	if str == "" {
		return str
	}

	var result strings.Builder
	result.Grow(len(str) * 2) // Estimate for growth

	for _, b := range []byte(str) {
		if shouldEncodeRFC3986(b) {
			result.WriteString(fmt.Sprintf("%%%02X", b))
		} else {
			result.WriteByte(b)
		}
	}

	return result.String()
}

// shouldEncodeRFC3986 determines if a byte should be percent-encoded according to RFC 3986
func shouldEncodeRFC3986(b byte) bool {
	// RFC 3986 unreserved characters: A-Z a-z 0-9 - . _ ~
	if (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9') {
		return false
	}
	if b == '-' || b == '.' || b == '_' || b == '~' {
		return false
	}
	// Everything else should be encoded
	return true
}

// rawUrlDecode decodes a URL encoded string according to RFC 3986
func rawUrlDecode(str string) string {
	if str == "" {
		return str
	}

	var result strings.Builder
	result.Grow(len(str)) // Decoded string is typically shorter

	for i := 0; i < len(str); i++ {
		if str[i] == '%' && i+2 < len(str) {
			// Try to decode the percent-encoded sequence
			hex1, hex2 := str[i+1], str[i+2]
			if isValidHexDigit(hex1) && isValidHexDigit(hex2) {
				// Convert hex digits to byte value
				val := hexDigitValue(hex1)*16 + hexDigitValue(hex2)
				result.WriteByte(val)
				i += 2 // Skip the hex digits
			} else {
				// Invalid hex sequence, keep the percent sign
				result.WriteByte(str[i])
			}
		} else {
			// Regular character (including lone % at end)
			result.WriteByte(str[i])
		}
	}

	return result.String()
}

// isValidHexDigit checks if a character is a valid hexadecimal digit
func isValidHexDigit(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f')
}

// hexDigitValue converts a hex digit character to its numeric value
func hexDigitValue(c byte) byte {
	if c >= '0' && c <= '9' {
		return c - '0'
	} else if c >= 'A' && c <= 'F' {
		return c - 'A' + 10
	} else if c >= 'a' && c <= 'f' {
		return c - 'a' + 10
	}
	return 0 // Should not happen if isValidHexDigit was called first
}

// calculateLevenshteinDistance implements the Levenshtein distance algorithm
// Returns the minimum number of single-character edits (insertions, deletions, or substitutions)
// required to change one string into the other
// Note: PHP's levenshtein() works on byte level, not Unicode character level
func calculateLevenshteinDistance(str1, str2 string) int {
	// Convert strings to bytes for PHP compatibility
	bytes1 := []byte(str1)
	bytes2 := []byte(str2)

	len1 := len(bytes1)
	len2 := len(bytes2)

	// Special cases
	if len1 == 0 {
		return len2
	}
	if len2 == 0 {
		return len1
	}

	// Create a matrix to store distances
	// dp[i][j] represents the distance between the first i characters of str1 and first j characters of str2
	dp := make([][]int, len1+1)
	for i := range dp {
		dp[i] = make([]int, len2+1)
	}

	// Initialize base cases
	// Distance from empty string to any prefix
	for i := 0; i <= len1; i++ {
		dp[i][0] = i
	}
	for j := 0; j <= len2; j++ {
		dp[0][j] = j
	}

	// Fill the matrix using dynamic programming
	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			// Cost of substitution (0 if bytes are same, 1 if different)
			substitutionCost := 1
			if bytes1[i-1] == bytes2[j-1] {
				substitutionCost = 0
			}

			// Calculate minimum of three operations:
			// 1. Deletion: dp[i-1][j] + 1
			// 2. Insertion: dp[i][j-1] + 1
			// 3. Substitution: dp[i-1][j-1] + substitutionCost
			dp[i][j] = min(
				dp[i-1][j]+1,           // deletion
				dp[i][j-1]+1,           // insertion
				dp[i-1][j-1]+substitutionCost, // substitution
			)
		}
	}

	return dp[len1][len2]
}

// min returns the minimum of three integers
func min(a, b, c int) int {
	if a <= b && a <= c {
		return a
	}
	if b <= c {
		return b
	}
	return c
}

// calculateHash implements the hash() function with support for multiple algorithms
func calculateHash(algo, data string) (string, error) {
	var hasher hash.Hash

	// Convert algorithm name to lowercase for case-insensitive matching
	algo = strings.ToLower(algo)

	// Select the appropriate hash algorithm
	switch algo {
	case "md5":
		hasher = md5.New()
	case "sha1":
		hasher = sha1.New()
	case "sha224":
		hasher = sha256.New224()
	case "sha256":
		hasher = sha256.New()
	case "sha384":
		hasher = sha512.New384()
	case "sha512":
		hasher = sha512.New()
	case "sha512/224":
		hasher = sha512.New512_224()
	case "sha512/256":
		hasher = sha512.New512_256()
	default:
		return "", fmt.Errorf("hash(): Unknown hashing algorithm: %s", algo)
	}

	// Hash the data
	hasher.Write([]byte(data))
	hashBytes := hasher.Sum(nil)

	// Return as lowercase hexadecimal string
	return hex.EncodeToString(hashBytes), nil
}

// formatMoney implements the money_format() function
// Provides basic money formatting with US locale defaults
func formatMoney(format string, number float64) (string, error) {
	// Parse the format string to extract components
	formatSpec, err := parseMoneyFormat(format)
	if err != nil {
		return "", err
	}

	// Apply precision (rounding)
	var rounded float64
	if formatSpec.Precision >= 0 {
		rounded = math.Round(number*math.Pow(10, float64(formatSpec.Precision))) / math.Pow(10, float64(formatSpec.Precision))
	} else {
		// No precision specified, round to nearest integer
		rounded = math.Round(number)
	}

	// Handle negative numbers
	isNegative := rounded < 0
	if isNegative {
		rounded = -rounded
	}

	// Format the number with precision
	var formattedNumber string
	if formatSpec.Precision >= 0 {
		formattedNumber = fmt.Sprintf("%."+strconv.Itoa(formatSpec.Precision)+"f", rounded)
	} else {
		// No precision specified, round to nearest integer
		formattedNumber = fmt.Sprintf("%.0f", math.Round(rounded))
	}

	// Add thousands separators (commas)
	formattedNumber = addThousandsSeparators(formattedNumber)

	// Add currency symbol based on format type
	var result string
	switch formatSpec.Type {
	case 'n': // national format
		if isNegative {
			result = "-$" + formattedNumber
		} else {
			result = "$" + formattedNumber
		}
	case 'i': // international format
		if isNegative {
			result = "-USD " + formattedNumber
		} else {
			result = "USD " + formattedNumber
		}
	default:
		return "", fmt.Errorf("money_format: unsupported format type '%c'", formatSpec.Type)
	}

	// Apply width and alignment
	if formatSpec.Width > 0 {
		if formatSpec.LeftAlign {
			// Left align with padding on the right
			result = result + strings.Repeat(" ", formatSpec.Width-len(result))
		} else {
			// Right align with padding on the left
			if len(result) < formatSpec.Width {
				padding := formatSpec.Width - len(result)
				result = strings.Repeat(" ", padding) + result
			}
		}
	}

	return result, nil
}

// moneyFormatSpec holds parsed format specification
type moneyFormatSpec struct {
	Width     int
	Precision int
	Type      rune
	LeftAlign bool
}

// parseMoneyFormat parses a money_format format string
func parseMoneyFormat(format string) (*moneyFormatSpec, error) {
	spec := &moneyFormatSpec{
		Width:     0,
		Precision: -1, // Default: no precision specified
		Type:      'n', // Default: national
		LeftAlign: false,
	}

	if len(format) == 0 {
		return nil, fmt.Errorf("money_format: empty format string")
	}

	i := 0
	if format[i] != '%' {
		return nil, fmt.Errorf("money_format: format must start with %%")
	}
	i++

	// Parse flags
	for i < len(format) {
		switch format[i] {
		case '-':
			spec.LeftAlign = true
			i++
		case '0':
			// Zero padding - for now we'll treat as regular padding
			i++
		default:
			goto parseWidth
		}
	}

parseWidth:
	// Parse width
	widthStart := i
	for i < len(format) && format[i] >= '0' && format[i] <= '9' {
		i++
	}
	if i > widthStart {
		width, err := strconv.Atoi(format[widthStart:i])
		if err == nil {
			spec.Width = width
		}
	}

	// Parse precision
	if i < len(format) && format[i] == '.' {
		i++
		precisionStart := i
		for i < len(format) && format[i] >= '0' && format[i] <= '9' {
			i++
		}
		if i > precisionStart {
			precision, err := strconv.Atoi(format[precisionStart:i])
			if err == nil {
				spec.Precision = precision
			}
		} else {
			spec.Precision = 0 // "." with no digits means 0 precision
		}
	}

	// Parse type
	if i < len(format) {
		spec.Type = rune(format[i])
		if spec.Type != 'n' && spec.Type != 'i' {
			return nil, fmt.Errorf("money_format: unsupported format type '%c'", spec.Type)
		}
	}

	return spec, nil
}

// addThousandsSeparators adds commas to separate thousands
func addThousandsSeparators(s string) string {
	// Find decimal point if it exists
	parts := strings.Split(s, ".")
	intPart := parts[0]

	// Add commas to integer part
	result := ""
	for i, digit := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			result += ","
		}
		result += string(digit)
	}

	// Add decimal part back if it exists
	if len(parts) > 1 {
		result += "." + parts[1]
	}

	return result
}