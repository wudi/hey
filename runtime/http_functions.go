package runtime

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func GetHTTPFunctions() []*registry.Function {
	return []*registry.Function{
		{
			Name:      "header",
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 1 {
					return nil, fmt.Errorf("header() expects at least 1 parameter, %d given", len(args))
				}

				header := args[0].ToString()
				replace := true
				responseCode := 0

				if len(args) >= 2 {
					replace = args[1].ToBool()
				}

				if len(args) >= 3 {
					responseCode = int(args[2].ToInt())
				}

				httpCtx := ctx.GetHTTPContext()
				if httpCtx == nil {
					return values.NewNull(), nil
				}

				parts := strings.SplitN(header, ":", 2)
				if len(parts) != 2 {
					return values.NewNull(), nil
				}

				name := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				if err := httpCtx.AddHeader(name, value, replace); err != nil {
					return values.NewBool(false), nil
				}

				if responseCode > 0 {
					httpCtx.SetResponseCode(responseCode)
				}

				return values.NewNull(), nil
			},
		},
		{
			Name:      "header_remove",
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				httpCtx := ctx.GetHTTPContext()
				if httpCtx == nil {
					return values.NewNull(), nil
				}

				if len(args) == 0 {
					ctx.ResetHTTPContext()
					return values.NewNull(), nil
				}

				name := args[0].ToString()
				ctx.RemoveHTTPHeader(name)

				return values.NewNull(), nil
			},
		},
		{
			Name:      "headers_list",
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				httpCtx := ctx.GetHTTPContext()
				if httpCtx == nil {
					return values.NewArray(), nil
				}

				headersList := httpCtx.GetHeadersList()
				arr := values.NewArray()
				for _, header := range headersList {
					arr.ArraySet(nil, values.NewString(header))
				}

				return arr, nil
			},
		},
		{
			Name:      "headers_sent",
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				httpCtx := ctx.GetHTTPContext()
				if httpCtx == nil {
					return values.NewBool(false), nil
				}

				sent, location := httpCtx.AreHeadersSent()

				if len(args) >= 1 && args[0].IsReference() {
					ref := args[0].Data.(*values.Reference)
					parts := strings.SplitN(location, ":", 2)
					if len(parts) > 0 {
						ref.Target = values.NewString(parts[0])
					}
				}

				if len(args) >= 2 && args[1].IsReference() {
					ref := args[1].Data.(*values.Reference)
					parts := strings.SplitN(location, ":", 2)
					if len(parts) > 1 {
						ref.Target = values.NewString(parts[1])
					}
				}

				return values.NewBool(sent), nil
			},
		},
		{
			Name:      "http_response_code",
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				httpCtx := ctx.GetHTTPContext()
				if httpCtx == nil {
					return values.NewBool(false), nil
				}

				currentCode := httpCtx.GetResponseCode()

				if len(args) > 0 {
					newCode := int(args[0].ToInt())
					if err := httpCtx.SetResponseCode(newCode); err != nil {
						return values.NewBool(false), nil
					}
					return values.NewInt(int64(currentCode)), nil
				}

				return values.NewInt(int64(currentCode)), nil
			},
		},
		{
			Name:      "setcookie",
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				return setcookieImpl(ctx, args, false)
			},
		},
		{
			Name:      "setrawcookie",
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				return setcookieImpl(ctx, args, true)
			},
		},
		{
			Name:      "getallheaders",
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				httpCtx := ctx.GetHTTPContext()
				if httpCtx == nil {
					return values.NewArray(), nil
				}

				headers := httpCtx.GetRequestHeaders()
				arr := values.NewArray()

				for name, value := range headers {
					arr.ArraySet(values.NewString(name), values.NewString(value))
				}

				return arr, nil
			},
		},
		{
			Name: "parse_str",
			Parameters: []*registry.Parameter{
				{Name: "string", Type: "string"},
				{Name: "result", Type: "", IsReference: true},
			},
			ReturnType: "void",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 {
					return nil, fmt.Errorf("parse_str() expects exactly 2 parameters, %d given", len(args))
				}

				str := args[0].ToString()

				// Parse the query string using Go's url.ParseQuery
				// This handles URL decoding and multiple values automatically
				parsed, err := url.ParseQuery(str)
				if err != nil {
					// On parse error, just return without setting anything
					return values.NewNull(), nil
				}

				result := values.NewArray()

				// Process each key-value pair
				for key, vals := range parsed {
					// Handle PHP array syntax: foo[] or foo[bar]
					if strings.Contains(key, "[") {
						// This is an array-style key
						parseArrayKey(result, key, vals)
					} else {
						// Simple key
						if len(vals) == 1 {
							result.ArraySet(values.NewString(key), values.NewString(vals[0]))
						} else if len(vals) > 1 {
							// Multiple values for same key (rare)
							arr := values.NewArray()
							for _, v := range vals {
								arr.ArraySet(nil, values.NewString(v))
							}
							result.ArraySet(values.NewString(key), arr)
						}
					}
				}

				// Set the reference parameter (second parameter)
				if args[1].IsReference() {
					ref := args[1].Data.(*values.Reference)
					*ref.Target = *result
				}

				return values.NewNull(), nil
			},
		},
	{
		Name: "wp_parse_str",
		Parameters: []*registry.Parameter{
			{Name: "string", Type: "string"},
			{Name: "result", Type: "", IsReference: true},
		},
		ReturnType: "void",
		MinArgs:    2,
		MaxArgs:    2,
		IsBuiltin:  true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			// wp_parse_str is a WordPress wrapper around parse_str
			// For now, we just call parse_str without applying the filter

			if len(args) < 2 {
				return nil, fmt.Errorf("wp_parse_str() expects exactly 2 parameters, %d given", len(args))
			}

			str := args[0].ToString()

			// Parse the query string using Go's url.ParseQuery
			parsed, err := url.ParseQuery(str)
			if err != nil {
				// On parse error, just return without setting anything
				return values.NewNull(), nil
			}

			result := values.NewArray()

			// Process each key-value pair
			for key, vals := range parsed {
				// Handle PHP array syntax: foo[] or foo[bar]
				if strings.Contains(key, "[") {
					// This is an array-style key
					parseArrayKey(result, key, vals)
				} else {
					// Simple key
					if len(vals) == 1 {
						result.ArraySet(values.NewString(key), values.NewString(vals[0]))
					} else if len(vals) > 1 {
						// Multiple values for same key -> array
						arr := values.NewArray()
						for i, v := range vals {
							arr.ArraySet(values.NewInt(int64(i)), values.NewString(v))
						}
						result.ArraySet(values.NewString(key), arr)
					}
				}
			}

			// Set the reference parameter (second parameter)
			if args[1].IsReference() {
				ref := args[1].Data.(*values.Reference)
				*ref.Target = *result
			}

			// TODO: Apply 'wp_parse_str' filter if apply_filters is available

			return values.NewNull(), nil
		},
	},
	}
}

// parseArrayKey handles PHP array syntax in query strings like foo[] or foo[bar] or foo[bar][baz]
func parseArrayKey(result *values.Value, key string, vals []string) {
	// Find the first '[' to split base key from array keys
	openBracket := strings.Index(key, "[")
	if openBracket == -1 {
		// No brackets, shouldn't happen but handle it
		if len(vals) > 0 {
			result.ArraySet(values.NewString(key), values.NewString(vals[0]))
		}
		return
	}

	baseKey := key[:openBracket]
	arrayPart := key[openBracket:]

	// Get or create the base array
	baseVal := result.ArrayGet(values.NewString(baseKey))
	if baseVal == nil || baseVal.Type != values.TypeArray {
		baseVal = values.NewArray()
		result.ArraySet(values.NewString(baseKey), baseVal)
	}

	// Parse array indices: [] or [key] or [key1][key2]...
	indices := parseArrayIndices(arrayPart)

	// Set the value(s) at the array path
	if len(indices) == 0 {
		// Just foo[], append all values
		for _, v := range vals {
			baseVal.ArraySet(nil, values.NewString(v))
		}
	} else if len(indices) == 1 && indices[0] == "" {
		// foo[], append all values
		for _, v := range vals {
			baseVal.ArraySet(nil, values.NewString(v))
		}
	} else {
		// foo[key] or foo[key1][key2]...
		setNestedArrayValue(baseVal, indices, vals)
	}
}

// parseArrayIndices extracts array indices from [key1][key2]... format
func parseArrayIndices(arrayPart string) []string {
	var indices []string
	current := ""
	inBracket := false

	for _, ch := range arrayPart {
		if ch == '[' {
			inBracket = true
			current = ""
		} else if ch == ']' {
			if inBracket {
				indices = append(indices, current)
				current = ""
			}
			inBracket = false
		} else if inBracket {
			current += string(ch)
		}
	}

	return indices
}

// setNestedArrayValue sets a value in a nested array structure
func setNestedArrayValue(arr *values.Value, indices []string, vals []string) {
	if len(indices) == 0 {
		// No more indices, set the values
		if len(vals) == 1 {
			// Should not happen at this level
			return
		}
		return
	}

	if len(indices) == 1 {
		// Last index, set the value(s)
		key := indices[0]
		if key == "" {
			// Empty index means append
			for _, v := range vals {
				arr.ArraySet(nil, values.NewString(v))
			}
		} else {
			// Specific index
			if len(vals) == 1 {
				arr.ArraySet(values.NewString(key), values.NewString(vals[0]))
			} else {
				// Multiple values, create array
				valArr := values.NewArray()
				for _, v := range vals {
					valArr.ArraySet(nil, values.NewString(v))
				}
				arr.ArraySet(values.NewString(key), valArr)
			}
		}
		return
	}

	// More indices to traverse
	key := indices[0]
	var nextArr *values.Value

	if key == "" {
		// foo[][bar] - create a new array element
		nextArr = values.NewArray()
		arr.ArraySet(nil, nextArr)
	} else {
		// foo[key][bar] - get or create array at key
		nextArr = arr.ArrayGet(values.NewString(key))
		if nextArr == nil || nextArr.Type != values.TypeArray {
			nextArr = values.NewArray()
			arr.ArraySet(values.NewString(key), nextArr)
		}
	}

	// Recurse
	setNestedArrayValue(nextArr, indices[1:], vals)
}

func setcookieImpl(ctx registry.BuiltinCallContext, args []*values.Value, raw bool) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("setcookie() expects at least 1 parameter, %d given", len(args))
	}

	httpCtx := ctx.GetHTTPContext()
	if httpCtx == nil {
		return values.NewBool(false), nil
	}

	name := args[0].ToString()
	value := ""
	expires := int64(0)
	path := ""
	domain := ""
	secure := false
	httponly := false

	if len(args) >= 2 {
		value = args[1].ToString()
	}
	if len(args) >= 3 {
		expires = args[2].ToInt()
	}
	if len(args) >= 4 {
		path = args[3].ToString()
	}
	if len(args) >= 5 {
		domain = args[4].ToString()
	}
	if len(args) >= 6 {
		secure = args[5].ToBool()
	}
	if len(args) >= 7 {
		httponly = args[6].ToBool()
	}

	var cookieValue string
	if raw {
		cookieValue = value
	} else {
		cookieValue = url.QueryEscape(value)
	}

	cookie := fmt.Sprintf("%s=%s", name, cookieValue)

	if expires > 0 {
		expiresTime := time.Unix(expires, 0).UTC()
		cookie += fmt.Sprintf("; Expires=%s", expiresTime.Format(time.RFC1123))
		cookie += fmt.Sprintf("; Max-Age=%d", expires-time.Now().Unix())
	}

	if path != "" {
		cookie += fmt.Sprintf("; Path=%s", path)
	}

	if domain != "" {
		cookie += fmt.Sprintf("; Domain=%s", domain)
	}

	if secure {
		cookie += "; Secure"
	}

	if httponly {
		cookie += "; HttpOnly"
	}

	if err := httpCtx.AddHeader("Set-Cookie", cookie, false); err != nil {
		return values.NewBool(false), nil
	}

	return values.NewBool(true), nil
}