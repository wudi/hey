package spl

import (
	"fmt"
	"regexp"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// RegexIterator mode constants (matching PHP constants)
const (
	REGEX_ITERATOR_MATCH       = 0
	REGEX_ITERATOR_GET_MATCH   = 1
	REGEX_ITERATOR_ALL_MATCHES = 2
	REGEX_ITERATOR_SPLIT       = 3
	REGEX_ITERATOR_REPLACE     = 4
)

// RegexIterator flag constants
const (
	REGEX_ITERATOR_USE_KEY = 1
)

// GetRegexIteratorClass returns the RegexIterator class descriptor
func GetRegexIteratorClass() *registry.ClassDescriptor {
	// Constructor
	constructorImpl := &registry.Function{
		Name:      "__construct",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 {
				return values.NewNull(), fmt.Errorf("RegexIterator::__construct() expects at least 1 argument")
			}

			thisObj := args[0]

			// Handle VM parameter passing issue - make parameters optional
			var iterator *values.Value = values.NewNull()
			var regex *values.Value = values.NewString("")
			var mode *values.Value = values.NewInt(REGEX_ITERATOR_MATCH)
			var flags *values.Value = values.NewInt(0)
			var preg_flags *values.Value = values.NewInt(0)

			if len(args) > 1 && !args[1].IsNull() {
				iterator = args[1]
			}
			if len(args) > 2 && !args[2].IsNull() {
				regex = args[2]
			}
			if len(args) > 3 && !args[3].IsNull() {
				mode = args[3]
			}
			if len(args) > 4 && !args[4].IsNull() {
				flags = args[4]
			}
			if len(args) > 5 && !args[5].IsNull() {
				preg_flags = args[5]
			}

			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("__construct called on non-object")
			}

			if !iterator.IsNull() && !iterator.IsObject() {
				return values.NewNull(), fmt.Errorf("Iterator must be an object")
			}

			obj := thisObj.Data.(*values.Object)
			if obj.Properties == nil {
				obj.Properties = make(map[string]*values.Value)
			}

			// Store parameters
			obj.Properties["__iterator"] = iterator
			obj.Properties["__regex"] = regex
			obj.Properties["__mode"] = mode
			obj.Properties["__flags"] = flags
			obj.Properties["__preg_flags"] = preg_flags

			// Compile regex pattern
			regexStr := regex.ToString()
			if regexStr != "" {
				// Remove PHP delimiters and extract pattern
				pattern := regexStr
				if len(regexStr) > 2 && regexStr[0] == '/' {
					// Find last / and extract pattern
					lastSlash := -1
					for i := len(regexStr) - 1; i > 0; i-- {
						if regexStr[i] == '/' {
							lastSlash = i
							break
						}
					}
					if lastSlash > 0 {
						pattern = regexStr[1:lastSlash]
					}
				}

				compiledRegex, err := regexp.Compile(pattern)
				if err != nil {
					return values.NewNull(), fmt.Errorf("Invalid regex pattern: %v", err)
				}
				obj.Properties["__compiled_regex"] = &values.Value{
					Type: values.TypeResource,
					Data: compiledRegex,
				}
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "iterator", Type: "Iterator", HasDefault: true, DefaultValue: values.NewNull()},
			{Name: "regex", Type: "string", HasDefault: true, DefaultValue: values.NewString("")},
			{Name: "mode", Type: "int", HasDefault: true, DefaultValue: values.NewInt(REGEX_ITERATOR_MATCH)},
			{Name: "flags", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
			{Name: "preg_flags", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
		},
	}

	// accept() method - implements the filtering logic
	acceptImpl := &registry.Function{
		Name:      "accept",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("accept() expects exactly 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewBool(false), fmt.Errorf("accept() called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			iterator := obj.Properties["__iterator"]
			compiledRegex := obj.Properties["__compiled_regex"]
			flags := obj.Properties["__flags"]

			if iterator == nil || iterator.IsNull() || compiledRegex == nil {
				return values.NewBool(false), nil
			}

			regex := compiledRegex.Data.(*regexp.Regexp)

			// Get current value or key to test against regex
			var testValue *values.Value
			useKey := false
			if flags != nil && flags.IsInt() && (flags.ToInt()&REGEX_ITERATOR_USE_KEY) != 0 {
				useKey = true
			}

			if useKey {
				// Test against key
				keyResult, err := callIteratorMethod(ctx, iterator, "key", []*values.Value{iterator})
				if err != nil {
					return values.NewBool(false), nil
				}
				testValue = keyResult
			} else {
				// Test against value (default)
				currentResult, err := callIteratorMethod(ctx, iterator, "current", []*values.Value{iterator})
				if err != nil {
					return values.NewBool(false), nil
				}
				testValue = currentResult
			}

			if testValue == nil {
				return values.NewBool(false), nil
			}

			// Apply regex to test value
			testStr := testValue.ToString()
			matches := regex.MatchString(testStr)

			return values.NewBool(matches), nil
		},
	}

	// current() method - transforms the value based on mode
	currentImpl := &registry.Function{
		Name:      "current",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("current() expects exactly 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("current() called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			iterator := obj.Properties["__iterator"]
			compiledRegex := obj.Properties["__compiled_regex"]
			mode := obj.Properties["__mode"]

			if iterator == nil || iterator.IsNull() {
				return values.NewNull(), nil
			}

			// Get current value from inner iterator
			currentResult, err := callIteratorMethod(ctx, iterator, "current", []*values.Value{iterator})
			if err != nil {
				return values.NewNull(), err
			}

			if currentResult == nil {
				return values.NewNull(), nil
			}

			// Apply mode transformation
			modeValue := REGEX_ITERATOR_MATCH
			if mode != nil && mode.IsInt() {
				modeValue = int(mode.ToInt())
			}

			switch modeValue {
			case REGEX_ITERATOR_MATCH:
				// Return original value (default mode)
				return currentResult, nil

			case REGEX_ITERATOR_GET_MATCH:
				// Return array of matches
				if compiledRegex != nil {
					regex := compiledRegex.Data.(*regexp.Regexp)
					testStr := currentResult.ToString()
					matches := regex.FindStringSubmatch(testStr)

					result := values.NewArray()
					for i, match := range matches {
						result.ArraySet(values.NewInt(int64(i)), values.NewString(match))
					}
					return result, nil
				}
				return values.NewArray(), nil

			case REGEX_ITERATOR_ALL_MATCHES:
				// Return array of all matches
				if compiledRegex != nil {
					regex := compiledRegex.Data.(*regexp.Regexp)
					testStr := currentResult.ToString()
					allMatches := regex.FindAllStringSubmatch(testStr, -1)

					result := values.NewArray()
					for i, matchGroup := range allMatches {
						groupArray := values.NewArray()
						for j, match := range matchGroup {
							groupArray.ArraySet(values.NewInt(int64(j)), values.NewString(match))
						}
						result.ArraySet(values.NewInt(int64(i)), groupArray)
					}
					return result, nil
				}
				return values.NewArray(), nil

			case REGEX_ITERATOR_SPLIT:
				// Split by regex
				if compiledRegex != nil {
					regex := compiledRegex.Data.(*regexp.Regexp)
					testStr := currentResult.ToString()
					parts := regex.Split(testStr, -1)

					result := values.NewArray()
					for i, part := range parts {
						result.ArraySet(values.NewInt(int64(i)), values.NewString(part))
					}
					return result, nil
				}
				return values.NewArray(), nil

			case REGEX_ITERATOR_REPLACE:
				// Not implemented - would need replacement string
				return currentResult, nil

			default:
				return currentResult, nil
			}
		},
	}

	// Delegate other iterator methods to FilterIterator behavior
	// key(), next(), rewind(), valid() methods
	keyImpl := &registry.Function{
		Name:      "key",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("key() expects exactly 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("key() called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			iterator := obj.Properties["__iterator"]

			if iterator == nil || iterator.IsNull() {
				return values.NewNull(), nil
			}

			// Delegate to inner iterator's key method
			return callIteratorMethod(ctx, iterator, "key", []*values.Value{iterator})
		},
	}

	nextImpl := &registry.Function{
		Name:      "next",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("next() expects exactly 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("next() called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			iterator := obj.Properties["__iterator"]

			if iterator == nil || iterator.IsNull() {
				return values.NewNull(), nil
			}

			// Use FilterIterator logic: advance until we find a match or reach end
			for {
				// Move to next element
				_, err := callIteratorMethod(ctx, iterator, "next", []*values.Value{iterator})
				if err != nil {
					return values.NewNull(), err
				}

				// Check if still valid
				validResult, err := callIteratorMethod(ctx, iterator, "valid", []*values.Value{iterator})
				if err != nil || !validResult.ToBool() {
					break
				}

				// Check if current element matches filter
				acceptResult, err := acceptImpl.Builtin(ctx, []*values.Value{thisObj})
				if err == nil && acceptResult.ToBool() {
					break
				}
			}

			return values.NewNull(), nil
		},
	}

	rewindImpl := &registry.Function{
		Name:      "rewind",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("rewind() expects exactly 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("rewind() called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			iterator := obj.Properties["__iterator"]

			if iterator == nil || iterator.IsNull() {
				return values.NewNull(), nil
			}

			// Rewind inner iterator
			_, err := callIteratorMethod(ctx, iterator, "rewind", []*values.Value{iterator})
			if err != nil {
				return values.NewNull(), err
			}

			// Use FilterIterator logic: find first matching element
			validResult, err := callIteratorMethod(ctx, iterator, "valid", []*values.Value{iterator})
			if err == nil && validResult.ToBool() {
				// Check if current element matches filter
				acceptResult, err := acceptImpl.Builtin(ctx, []*values.Value{thisObj})
				if err != nil || !acceptResult.ToBool() {
					// Current doesn't match, advance to next matching element
					nextImpl.Builtin(ctx, []*values.Value{thisObj})
				}
			}

			return values.NewNull(), nil
		},
	}

	validImpl := &registry.Function{
		Name:      "valid",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("valid() expects exactly 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewBool(false), fmt.Errorf("valid() called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			iterator := obj.Properties["__iterator"]

			if iterator == nil || iterator.IsNull() {
				return values.NewBool(false), nil
			}

			// Delegate to inner iterator's valid method
			return callIteratorMethod(ctx, iterator, "valid", []*values.Value{iterator})
		},
	}

	// getInnerIterator() method - implements OuterIterator
	getInnerIteratorImpl := &registry.Function{
		Name:      "getInnerIterator",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("getInnerIterator() expects exactly 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("getInnerIterator() called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			iterator := obj.Properties["__iterator"]

			if iterator == nil {
				return values.NewNull(), nil
			}

			return iterator, nil
		},
	}

	return &registry.ClassDescriptor{
		Name:   "RegexIterator",
		Parent: "FilterIterator", // Extends FilterIterator
		Methods: map[string]*registry.MethodDescriptor{
			"__construct": {
				Name:       "__construct",
				Visibility: "public",
				Parameters: []*registry.ParameterDescriptor{
					{Name: "iterator", Type: "Iterator", HasDefault: true, DefaultValue: values.NewNull()},
					{Name: "regex", Type: "string", HasDefault: true, DefaultValue: values.NewString("")},
					{Name: "mode", Type: "int", HasDefault: true, DefaultValue: values.NewInt(REGEX_ITERATOR_MATCH)},
					{Name: "flags", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
					{Name: "preg_flags", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
				},
				Implementation: NewBuiltinMethodImpl(constructorImpl),
			},
			"accept": {
				Name:           "accept",
				Visibility:     "public",
				Parameters:     []*registry.ParameterDescriptor{},
				Implementation: NewBuiltinMethodImpl(acceptImpl),
			},
			"current": {
				Name:           "current",
				Visibility:     "public",
				Parameters:     []*registry.ParameterDescriptor{},
				Implementation: NewBuiltinMethodImpl(currentImpl),
			},
			"key": {
				Name:           "key",
				Visibility:     "public",
				Parameters:     []*registry.ParameterDescriptor{},
				Implementation: NewBuiltinMethodImpl(keyImpl),
			},
			"next": {
				Name:           "next",
				Visibility:     "public",
				Parameters:     []*registry.ParameterDescriptor{},
				Implementation: NewBuiltinMethodImpl(nextImpl),
			},
			"rewind": {
				Name:           "rewind",
				Visibility:     "public",
				Parameters:     []*registry.ParameterDescriptor{},
				Implementation: NewBuiltinMethodImpl(rewindImpl),
			},
			"valid": {
				Name:           "valid",
				Visibility:     "public",
				Parameters:     []*registry.ParameterDescriptor{},
				Implementation: NewBuiltinMethodImpl(validImpl),
			},
			"getInnerIterator": {
				Name:           "getInnerIterator",
				Visibility:     "public",
				Parameters:     []*registry.ParameterDescriptor{},
				Implementation: NewBuiltinMethodImpl(getInnerIteratorImpl),
			},
		},
		Interfaces: []string{"OuterIterator", "Iterator"},
	}
}