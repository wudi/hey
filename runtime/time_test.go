package runtime

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// TestTimeFunctions tests all time-related functions
func TestTimeFunctions(t *testing.T) {
	functions := GetTimeFunctions()
	functionMap := make(map[string]*registry.Function)
	for _, fn := range functions {
		functionMap[fn.Name] = fn
	}

	t.Run("time", func(t *testing.T) {
		fn := functionMap["time"]
		if fn == nil {
			t.Fatal("time function not found")
		}

		// Test basic functionality
		before := time.Now().Unix()
		result, err := fn.Builtin(nil, []*values.Value{})
		after := time.Now().Unix()

		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		if result == nil {
			t.Error("result is nil")
			return
		}

		timestamp := result.ToInt()
		if timestamp < before || timestamp > after {
			t.Errorf("expected timestamp between %d and %d, got %d", before, after, timestamp)
		}
	})

	t.Run("microtime", func(t *testing.T) {
		fn := functionMap["microtime"]
		if fn == nil {
			t.Fatal("microtime function not found")
		}

		tests := []struct {
			name         string
			args         []*values.Value
			expectedType string
		}{
			{
				name:         "microtime() default (string)",
				args:         []*values.Value{},
				expectedType: "string",
			},
			{
				name:         "microtime(false) (string)",
				args:         []*values.Value{values.NewBool(false)},
				expectedType: "string",
			},
			{
				name:         "microtime(true) (float)",
				args:         []*values.Value{values.NewBool(true)},
				expectedType: "float",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				before := time.Now()
				result, err := fn.Builtin(nil, tt.args)
				after := time.Now()

				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				if result == nil {
					t.Error("result is nil")
					return
				}

				if tt.expectedType == "string" {
					if !result.IsString() {
						t.Errorf("expected string result, got %T", result.Data)
						return
					}

					// String format should be "0.xxxxxxxx timestamp"
					str := result.ToString()
					parts := strings.Split(str, " ")
					if len(parts) != 2 {
						t.Errorf("expected string format '0.xxxxxxxx timestamp', got %s", str)
						return
					}

					// Check microseconds part
					if !strings.HasPrefix(parts[0], "0.") {
						t.Errorf("expected microseconds to start with '0.', got %s", parts[0])
					}

					// Check timestamp part
					timestamp, err := strconv.ParseInt(parts[1], 10, 64)
					if err != nil {
						t.Errorf("invalid timestamp in string: %s", parts[1])
						return
					}

					beforeUnix := before.Unix()
					afterUnix := after.Unix()
					if timestamp < beforeUnix || timestamp > afterUnix {
						t.Errorf("timestamp %d not in expected range [%d, %d]", timestamp, beforeUnix, afterUnix)
					}
				} else {
					// Float
					if !result.IsFloat() {
						t.Errorf("expected float result, got %T", result.Data)
						return
					}

					floatVal := result.ToFloat()
					beforeFloat := float64(before.Unix()) + float64(before.Nanosecond())/1e9
					afterFloat := float64(after.Unix()) + float64(after.Nanosecond())/1e9

					// Add small tolerance for timing precision
					tolerance := 0.001 // 1ms tolerance
					if floatVal < beforeFloat-tolerance || floatVal > afterFloat+tolerance {
						t.Errorf("float value %f not in expected range [%f, %f] (tolerance: %f)", floatVal, beforeFloat-tolerance, afterFloat+tolerance, tolerance)
					}
				}
			})
		}
	})

	t.Run("time_nanosleep", func(t *testing.T) {
		fn := functionMap["time_nanosleep"]
		if fn == nil {
			t.Fatal("time_nanosleep function not found")
		}

		tests := []struct {
			name        string
			args        []*values.Value
			expectError bool
			errorMsg    string
		}{
			{
				name: "sleep 0 seconds 0 nanoseconds",
				args: []*values.Value{values.NewInt(0), values.NewInt(0)},
			},
			{
				name: "sleep 0 seconds 100000000 nanoseconds (0.1 sec)",
				args: []*values.Value{values.NewInt(0), values.NewInt(100000000)},
			},
			{
				name: "sleep 1 second 0 nanoseconds",
				args: []*values.Value{values.NewInt(1), values.NewInt(0)},
			},
			{
				name:        "negative seconds",
				args:        []*values.Value{values.NewInt(-1), values.NewInt(0)},
				expectError: true,
				errorMsg:    "ValueError: time_nanosleep(): Argument #1 ($seconds) must be greater than or equal to 0",
			},
			{
				name:        "negative nanoseconds",
				args:        []*values.Value{values.NewInt(0), values.NewInt(-1)},
				expectError: true,
				errorMsg:    "ValueError: time_nanosleep(): Argument #2 ($nanoseconds) must be greater than or equal to 0",
			},
			{
				name:        "nanoseconds >= 1000000000",
				args:        []*values.Value{values.NewInt(0), values.NewInt(1000000000)},
				expectError: true,
				errorMsg:    "ValueError: Nanoseconds was not in the range 0 to 999 999 999 or seconds was negative",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				start := time.Now()
				result, err := fn.Builtin(nil, tt.args)
				duration := time.Since(start)

				if tt.expectError {
					if err == nil {
						t.Errorf("expected error but got none")
						return
					}
					if err.Error() != tt.errorMsg {
						t.Errorf("expected error %q, got %q", tt.errorMsg, err.Error())
					}
					return
				}

				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				if result == nil {
					t.Error("result is nil")
					return
				}

				// Should return true on success
				if !result.ToBool() {
					t.Errorf("expected true return value, got %v", result)
				}

				// Check timing for non-zero sleep
				if tt.args[0].ToInt() > 0 || tt.args[1].ToInt() > 0 {
					expectedDuration := time.Duration(tt.args[0].ToInt())*time.Second + time.Duration(tt.args[1].ToInt())*time.Nanosecond
					tolerance := 50 * time.Millisecond
					if duration < expectedDuration-tolerance || duration > expectedDuration+tolerance {
						t.Errorf("expected sleep duration around %v, got %v", expectedDuration, duration)
					}
				}
			})
		}
	})

	t.Run("time_sleep_until", func(t *testing.T) {
		fn := functionMap["time_sleep_until"]
		if fn == nil {
			t.Fatal("time_sleep_until function not found")
		}

		t.Run("sleep until future time", func(t *testing.T) {
			now := time.Now()
			targetTime := float64(now.Unix()) + float64(now.Nanosecond())/1e9 + 0.5 // 0.5 second from now
			start := time.Now()
			result, err := fn.Builtin(nil, []*values.Value{values.NewFloat(targetTime)})
			duration := time.Since(start)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("result is nil")
				return
			}

			// Should return true on success
			if !result.ToBool() {
				t.Errorf("expected true return value, got %v", result)
			}

			// Check that we slept approximately 0.5 seconds
			tolerance := 100 * time.Millisecond
			expectedDuration := 500 * time.Millisecond
			if duration < expectedDuration-tolerance || duration > expectedDuration+tolerance {
				t.Errorf("expected sleep duration around %v, got %v", expectedDuration, duration)
			}
		})

		t.Run("sleep until past time", func(t *testing.T) {
			pastTime := float64(time.Now().Unix()) - 10.0 // 10 seconds ago
			result, err := fn.Builtin(nil, []*values.Value{values.NewFloat(pastTime)})

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("result is nil")
				return
			}

			// Should return false for past time
			if result.ToBool() {
				t.Errorf("expected false return value for past time, got %v", result)
			}
		})

		t.Run("sleep until current time", func(t *testing.T) {
			currentTime := float64(time.Now().Unix())
			result, err := fn.Builtin(nil, []*values.Value{values.NewFloat(currentTime)})

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("result is nil")
				return
			}

			// Should return false for current/past time
			if result.ToBool() {
				t.Errorf("expected false return value for current time, got %v", result)
			}
		})
	})

	t.Run("sleep", func(t *testing.T) {
		fn := functionMap["sleep"]
		if fn == nil {
			t.Fatal("sleep function not found")
		}

		tests := []struct {
			name        string
			args        []*values.Value
			expected    interface{}
			expectError bool
			errorMsg    string
		}{
			{
				name:     "sleep for 0 seconds",
				args:     []*values.Value{values.NewInt(0)},
				expected: int64(0),
			},
			{
				name:        "sleep with negative value",
				args:        []*values.Value{values.NewInt(-1)},
				expectError: true,
				errorMsg:    "ValueError: sleep(): Argument #1 ($seconds) must be greater than or equal to 0",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)

				if tt.expectError {
					if err == nil {
						t.Errorf("expected error but got none")
						return
					}
					if err.Error() != tt.errorMsg {
						t.Errorf("expected error %q, got %q", tt.errorMsg, err.Error())
					}
					return
				}

				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				if result == nil {
					t.Error("result is nil")
					return
				}

				// Check return value
				expected := tt.expected.(int64)
				if result.ToInt() != expected {
					t.Errorf("expected return value %d, got %d", expected, result.ToInt())
				}
			})
		}
	})

	t.Run("usleep", func(t *testing.T) {
		fn := functionMap["usleep"]
		if fn == nil {
			t.Fatal("usleep function not found")
		}

		tests := []struct {
			name        string
			args        []*values.Value
			expectError bool
			errorMsg    string
		}{
			{
				name: "usleep for 0 microseconds",
				args: []*values.Value{values.NewInt(0)},
			},
			{
				name:        "usleep with negative value",
				args:        []*values.Value{values.NewInt(-1)},
				expectError: true,
				errorMsg:    "ValueError: usleep(): Argument #1 ($microseconds) must be greater than or equal to 0",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)

				if tt.expectError {
					if err == nil {
						t.Errorf("expected error but got none")
						return
					}
					if err.Error() != tt.errorMsg {
						t.Errorf("expected error %q, got %q", tt.errorMsg, err.Error())
					}
					return
				}

				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				if result == nil {
					t.Error("result is nil")
					return
				}

				// usleep should return null
				if !result.IsNull() {
					t.Errorf("expected null return value, got %v", result)
				}
			})
		}
	})
}