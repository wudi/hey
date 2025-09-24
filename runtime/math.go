package runtime

import (
	"math"
	"strconv"

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
	}
}