package runtime

import (
	"fmt"
	"time"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetTimeFunctions returns time-related PHP functions
func GetTimeFunctions() []*registry.Function {
	return []*registry.Function{
		{
			Name:       "time",
			Parameters: []*registry.Parameter{},
			ReturnType: "int",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Return current Unix timestamp
				now := time.Now().Unix()
				return values.NewInt(now), nil
			},
		},
		{
			Name: "microtime",
			Parameters: []*registry.Parameter{
				{Name: "get_as_float", Type: "bool", HasDefault: true, DefaultValue: values.NewBool(false)},
			},
			ReturnType: "string|float",
			MinArgs:    0,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				now := time.Now()
				seconds := now.Unix()
				microseconds := now.Nanosecond() / 1000

				// Determine if we should return as float
				getAsFloat := false
				if len(args) > 0 {
					getAsFloat = args[0].ToBool()
				}

				if getAsFloat {
					// Return as float: seconds.microseconds
					floatValue := float64(seconds) + float64(microseconds)/1000000.0
					return values.NewFloat(floatValue), nil
				} else {
					// Return as string: "0.microseconds seconds"
					microStr := fmt.Sprintf("0.%08d %d", microseconds, seconds)
					return values.NewString(microStr), nil
				}
			},
		},
		{
			Name: "time_nanosleep",
			Parameters: []*registry.Parameter{
				{Name: "seconds", Type: "int"},
				{Name: "nanoseconds", Type: "int"},
			},
			ReturnType: "bool|array",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) != 2 {
					return nil, fmt.Errorf("time_nanosleep() expects exactly 2 parameters, %d given", len(args))
				}

				seconds := args[0].ToInt()
				nanoseconds := args[1].ToInt()

				// Validate arguments
				if seconds < 0 {
					return nil, fmt.Errorf("ValueError: time_nanosleep(): Argument #1 ($seconds) must be greater than or equal to 0")
				}
				if nanoseconds < 0 {
					return nil, fmt.Errorf("ValueError: time_nanosleep(): Argument #2 ($nanoseconds) must be greater than or equal to 0")
				}
				if nanoseconds >= 1000000000 {
					return nil, fmt.Errorf("ValueError: Nanoseconds was not in the range 0 to 999 999 999 or seconds was negative")
				}

				// Sleep for the specified duration
				if seconds > 0 || nanoseconds > 0 {
					duration := time.Duration(seconds)*time.Second + time.Duration(nanoseconds)*time.Nanosecond
					time.Sleep(duration)
				}

				// Return true on success
				return values.NewBool(true), nil
			},
		},
		{
			Name: "time_sleep_until",
			Parameters: []*registry.Parameter{
				{Name: "timestamp", Type: "float"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("time_sleep_until() expects exactly 1 parameter, %d given", len(args))
				}

				targetTimestamp := args[0].ToFloat()
				now := time.Now()
				currentTimestamp := float64(now.Unix()) + float64(now.Nanosecond())/1e9

				// If target time is in the past, return false
				if targetTimestamp <= currentTimestamp {
					return values.NewBool(false), nil
				}

				// Calculate sleep duration
				sleepDuration := time.Duration((targetTimestamp - currentTimestamp) * 1e9)

				// Sleep until the target time
				time.Sleep(sleepDuration)

				return values.NewBool(true), nil
			},
		},
		{
			Name: "sleep",
			Parameters: []*registry.Parameter{
				{Name: "seconds", Type: "int"},
			},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return nil, fmt.Errorf("sleep() expects exactly 1 parameter, 0 given")
				}

				seconds := args[0].ToInt()
				if seconds < 0 {
					return nil, fmt.Errorf("ValueError: sleep(): Argument #1 ($seconds) must be greater than or equal to 0")
				}

				// Sleep for the specified number of seconds
				if seconds > 0 {
					time.Sleep(time.Duration(seconds) * time.Second)
				}

				// Return 0 on success (PHP behavior)
				return values.NewInt(0), nil
			},
		},
		{
			Name: "usleep",
			Parameters: []*registry.Parameter{
				{Name: "microseconds", Type: "int"},
			},
			ReturnType: "void",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return nil, fmt.Errorf("usleep() expects exactly 1 parameter, 0 given")
				}

				microseconds := args[0].ToInt()
				if microseconds < 0 {
					return nil, fmt.Errorf("ValueError: usleep(): Argument #1 ($microseconds) must be greater than or equal to 0")
				}

				// Sleep for the specified number of microseconds
				if microseconds > 0 {
					time.Sleep(time.Duration(microseconds) * time.Microsecond)
				}

				// Return null (PHP behavior for void functions)
				return values.NewNull(), nil
			},
		},
	}
}