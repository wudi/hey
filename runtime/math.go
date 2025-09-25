package runtime

import (
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetMathFunctions returns math-related PHP functions
func GetMathFunctions() []*registry.Function {
	return []*registry.Function{
		{
			Name:       "abs",
			Parameters: []*registry.Parameter{{Name: "number", Type: "mixed"}},
			ReturnType: "number",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewInt(0), nil
				}

				if args[0].IsInt() {
					val := args[0].ToInt()
					if val < 0 {
						return values.NewInt(-val), nil
					}
					return values.NewInt(val), nil
				}

				if args[0].IsFloat() {
					val := args[0].ToFloat()
					return values.NewFloat(math.Abs(val)), nil
				}

				// Try to convert string to number
				str := args[0].ToString()
				if val, err := strconv.ParseFloat(str, 64); err == nil {
					return values.NewFloat(math.Abs(val)), nil
				}

				return values.NewInt(0), nil
			},
		},
		{
			Name:       "round",
			Parameters: []*registry.Parameter{
				{Name: "number", Type: "mixed"},
			},
			ReturnType: "number",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewFloat(0), nil
				}

				val := args[0].ToFloat()
				precision := 0
				if len(args) > 1 {
					precision = int(args[1].ToInt())
				}

				// Round to specified precision
				multiplier := math.Pow(10, float64(precision))
				rounded := math.Round(val*multiplier) / multiplier

				// Return int if precision is 0 and result is whole number
				if precision == 0 && rounded == math.Trunc(rounded) {
					return values.NewInt(int64(rounded)), nil
				}

				return values.NewFloat(rounded), nil
			},
		},
		{
			Name:       "decoct",
			Parameters: []*registry.Parameter{{Name: "number", Type: "mixed"}},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewString("0"), nil
				}

				// Convert to integer first
				var num int64
				if args[0].IsInt() {
					num = args[0].ToInt()
				} else if args[0].IsFloat() {
					num = int64(args[0].ToFloat())
				} else if args[0].IsString() {
					val, err := strconv.ParseFloat(args[0].ToString(), 64)
					if err != nil {
						return values.NewString("0"), nil
					}
					num = int64(val)
				} else {
					return values.NewString("0"), nil
				}

				// Convert to octal string (without 0 prefix)
				return values.NewString(strconv.FormatInt(num, 8)), nil
			},
		},

		// max - Find maximum value
		{
			Name:       "max",
			Parameters: []*registry.Parameter{{Name: "value", Type: "mixed"}},
			ReturnType: "mixed",
			MinArgs:    1,
			MaxArgs:    999,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return values.NewNull(), nil
				}

				// If single argument is an array, find max of array elements
				if len(args) == 1 && args[0].IsArray() {
					arr := args[0]
					arrayData := arr.Data.(*values.Array)
					if len(arrayData.Elements) == 0 {
						return values.NewNull(), nil
					}

					var maxVal *values.Value
					for _, val := range arrayData.Elements {
						if maxVal == nil || compareValuesForMath(val, maxVal) > 0 {
							maxVal = val
						}
					}
					return maxVal, nil
				}

				// Find max among all arguments
				maxVal := args[0]
				for i := 1; i < len(args); i++ {
					if compareValuesForMath(args[i], maxVal) > 0 {
						maxVal = args[i]
					}
				}
				return maxVal, nil
			},
		},

		// min - Find minimum value
		{
			Name:       "min",
			Parameters: []*registry.Parameter{{Name: "value", Type: "mixed"}},
			ReturnType: "mixed",
			MinArgs:    1,
			MaxArgs:    999,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return values.NewNull(), nil
				}

				// If single argument is an array, find min of array elements
				if len(args) == 1 && args[0].IsArray() {
					arr := args[0]
					arrayData := arr.Data.(*values.Array)
					if len(arrayData.Elements) == 0 {
						return values.NewNull(), nil
					}

					var minVal *values.Value
					for _, val := range arrayData.Elements {
						if minVal == nil || compareValuesForMath(val, minVal) < 0 {
							minVal = val
						}
					}
					return minVal, nil
				}

				// Find min among all arguments
				minVal := args[0]
				for i := 1; i < len(args); i++ {
					if compareValuesForMath(args[i], minVal) < 0 {
						minVal = args[i]
					}
				}
				return minVal, nil
			},
		},

		// ceil - Round up
		{
			Name:       "ceil",
			Parameters: []*registry.Parameter{{Name: "value", Type: "mixed"}},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewInt(0), nil
				}

				val := args[0].ToFloat()
				return values.NewInt(int64(math.Ceil(val))), nil
			},
		},

		// floor - Round down
		{
			Name:       "floor",
			Parameters: []*registry.Parameter{{Name: "value", Type: "mixed"}},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewInt(0), nil
				}

				val := args[0].ToFloat()
				return values.NewInt(int64(math.Floor(val))), nil
			},
		},

		// rand - Generate random number
		{
			Name:       "rand",
			Parameters: []*registry.Parameter{
				{Name: "min", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
				{Name: "max", Type: "int", HasDefault: true, DefaultValue: values.NewInt(32767)},
			},
			ReturnType: "int",
			MinArgs:    0,
			MaxArgs:    2,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				min := int64(0)
				max := int64(32767) // PHP default

				if len(args) >= 1 {
					min = args[0].ToInt()
				}
				if len(args) >= 2 {
					max = args[1].ToInt()
				}

				if min > max {
					min, max = max, min
				}

				result := min + rand.Int63n(max-min+1)
				return values.NewInt(result), nil
			},
		},

		// mt_rand - Generate random number (Mersenne Twister)
		{
			Name:       "mt_rand",
			Parameters: []*registry.Parameter{
				{Name: "min", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
				{Name: "max", Type: "int", HasDefault: true, DefaultValue: values.NewInt(2147483647)},
			},
			ReturnType: "int",
			MinArgs:    0,
			MaxArgs:    2,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				min := int64(0)
				max := int64(2147483647) // PHP mt_rand default

				if len(args) >= 1 {
					min = args[0].ToInt()
				}
				if len(args) >= 2 {
					max = args[1].ToInt()
				}

				if min > max {
					min, max = max, min
				}

				result := min + rand.Int63n(max-min+1)
				return values.NewInt(result), nil
			},
		},

		// sin - Sine
		{
			Name:       "sin",
			Parameters: []*registry.Parameter{{Name: "arg", Type: "float"}},
			ReturnType: "float",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewFloat(0), nil
				}
				val := args[0].ToFloat()
				return values.NewFloat(math.Sin(val)), nil
			},
		},

		// cos - Cosine
		{
			Name:       "cos",
			Parameters: []*registry.Parameter{{Name: "arg", Type: "float"}},
			ReturnType: "float",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewFloat(1), nil
				}
				val := args[0].ToFloat()
				return values.NewFloat(math.Cos(val)), nil
			},
		},

		// tan - Tangent
		{
			Name:       "tan",
			Parameters: []*registry.Parameter{{Name: "arg", Type: "float"}},
			ReturnType: "float",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewFloat(0), nil
				}
				val := args[0].ToFloat()
				return values.NewFloat(math.Tan(val)), nil
			},
		},

		// sqrt - Square root
		{
			Name:       "sqrt",
			Parameters: []*registry.Parameter{{Name: "arg", Type: "float"}},
			ReturnType: "float",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewFloat(0), nil
				}
				val := args[0].ToFloat()
				return values.NewFloat(math.Sqrt(val)), nil
			},
		},

		// pow - Power
		{
			Name:       "pow",
			Parameters: []*registry.Parameter{
				{Name: "base", Type: "mixed"},
				{Name: "exp", Type: "mixed"},
			},
			ReturnType: "number",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 {
					return values.NewInt(0), nil
				}
				base := args[0].ToFloat()
				exp := args[1].ToFloat()
				result := math.Pow(base, exp)

				// Return int if result is a whole number
				if result == math.Trunc(result) && result >= -9223372036854775808 && result <= 9223372036854775807 {
					return values.NewInt(int64(result)), nil
				}

				return values.NewFloat(result), nil
			},
		},

		// log - Natural logarithm
		{
			Name:       "log",
			Parameters: []*registry.Parameter{{Name: "arg", Type: "float"}},
			ReturnType: "float",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewFloat(0), nil
				}
				val := args[0].ToFloat()
				return values.NewFloat(math.Log(val)), nil
			},
		},

		// exp - Exponential function
		{
			Name:       "exp",
			Parameters: []*registry.Parameter{{Name: "arg", Type: "float"}},
			ReturnType: "float",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewFloat(1), nil
				}
				val := args[0].ToFloat()
				return values.NewFloat(math.Exp(val)), nil
			},
		},
	}
}

// Initialize random seed once
func init() {
	rand.Seed(time.Now().UnixNano())
}

// compareValuesForMath compares two values for max/min operations
// Returns: >0 if a > b, <0 if a < b, 0 if a == b
func compareValuesForMath(a, b *values.Value) int {
	// Handle null values
	if a.IsNull() && b.IsNull() {
		return 0
	}
	if a.IsNull() {
		return -1
	}
	if b.IsNull() {
		return 1
	}

	// If both are strings, compare lexicographically
	if a.IsString() && b.IsString() {
		aStr := a.ToString()
		bStr := b.ToString()
		if aStr < bStr {
			return -1
		} else if aStr > bStr {
			return 1
		}
		return 0
	}

	// Convert to numbers for numeric comparison
	aFloat := a.ToFloat()
	bFloat := b.ToFloat()

	if aFloat < bFloat {
		return -1
	} else if aFloat > bFloat {
		return 1
	}
	return 0
}