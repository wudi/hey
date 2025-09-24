package vm

import (
	"strings"
	"testing"
	"time"

	"github.com/wudi/hey/registry"
	runtime2 "github.com/wudi/hey/runtime"
	"github.com/wudi/hey/values"
)

// TestExecutionContextSetTimeLimit tests basic SetTimeLimit functionality
func TestExecutionContextSetTimeLimit(t *testing.T) {
	tests := []struct {
		name           string
		seconds        int
		expectedResult bool
		expectedTime   time.Duration
	}{
		{
			name:           "Set positive limit",
			seconds:        30,
			expectedResult: true,
			expectedTime:   30 * time.Second,
		},
		{
			name:           "Set zero (unlimited)",
			seconds:        0,
			expectedResult: true,
			expectedTime:   0,
		},
		{
			name:           "Set negative (unlimited in PHP 8.0+)",
			seconds:        -1,
			expectedResult: true,
			expectedTime:   0,
		},
		{
			name:           "Large value",
			seconds:        3600,
			expectedResult: true,
			expectedTime:   3600 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewExecutionContext()

			result := ctx.SetTimeLimit(tt.seconds)
			if result != tt.expectedResult {
				t.Errorf("Expected result %v, got %v", tt.expectedResult, result)
			}

			actualDuration := ctx.GetMaxExecutionTimeAsDuration()
			if actualDuration != tt.expectedTime {
				t.Errorf("Expected max execution time %v, got %v", tt.expectedTime, actualDuration)
			}

			// Test GetMaxExecutionTime (seconds)
			expectedSeconds := tt.seconds
			if tt.seconds < 0 {
				expectedSeconds = 0 // Negative becomes unlimited (0)
			}
			actualSeconds := ctx.GetMaxExecutionTime()
			if actualSeconds != expectedSeconds {
				t.Errorf("Expected max execution time %d seconds, got %d", expectedSeconds, actualSeconds)
			}
		})
	}
}

// TestSetTimeLimitBuiltinFunction tests the actual set_time_limit PHP function
func TestSetTimeLimitBuiltinFunction(t *testing.T) {
	ctx := NewExecutionContext()
	builtinCtx := &builtinContext{
		vm:  NewVirtualMachine(),
		ctx: ctx,
	}

	systemFunctions := runtime2.GetSystemFunctions()
	var setTimeLimitFn *registry.Function
	for _, fn := range systemFunctions {
		if fn.Name == "set_time_limit" {
			setTimeLimitFn = fn
			break
		}
	}

	if setTimeLimitFn == nil {
		t.Fatal("set_time_limit function not found")
	}

	tests := []struct {
		name           string
		args           []*values.Value
		expectedResult bool
		expectedIni    string
	}{
		{
			name:           "Set 30 seconds",
			args:           []*values.Value{values.NewInt(30)},
			expectedResult: true,
			expectedIni:    "30",
		},
		{
			name:           "Set unlimited (0)",
			args:           []*values.Value{values.NewInt(0)},
			expectedResult: true,
			expectedIni:    "0",
		},
		{
			name:           "Set negative (unlimited)",
			args:           []*values.Value{values.NewInt(-1)},
			expectedResult: true,
			expectedIni:    "-1",
		},
		{
			name:           "String input '20'",
			args:           []*values.Value{values.NewString("20")},
			expectedResult: true,
			expectedIni:    "20",
		},
		{
			name:           "No arguments",
			args:           []*values.Value{},
			expectedResult: false,
			expectedIni:    "", // Won't check ini in this case
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := setTimeLimitFn.Builtin(builtinCtx, tt.args)
			if err != nil {
				t.Fatalf("set_time_limit failed: %v", err)
			}

			if result.ToBool() != tt.expectedResult {
				t.Errorf("Expected result %v, got %v", tt.expectedResult, result.ToBool())
			}

			if tt.expectedResult && tt.expectedIni != "" {
				// Test that ini_get returns the correct value
				iniFunctions := runtime2.GetIniFunctions()
				var iniGetFn *registry.Function
				for _, fn := range iniFunctions {
					if fn.Name == "ini_get" {
						iniGetFn = fn
						break
					}
				}

				if iniGetFn != nil {
					iniArgs := []*values.Value{values.NewString("max_execution_time")}
					iniResult, err := iniGetFn.Builtin(builtinCtx, iniArgs)
					if err != nil {
						t.Fatalf("ini_get failed: %v", err)
					}

					if iniResult.ToString() != tt.expectedIni {
						t.Errorf("Expected ini value %s, got %s", tt.expectedIni, iniResult.ToString())
					}
				}
			}
		})
	}
}

// TestTimeoutExecution tests that execution actually times out
func TestTimeoutExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping timeout test in short mode")
	}

	ctx := NewExecutionContext()

	// Set a very short timeout (1 second)
	success := ctx.SetTimeLimit(1)
	if !success {
		t.Fatal("Failed to set time limit")
	}

	// Wait longer than the timeout to trigger it
	time.Sleep(1100 * time.Millisecond)

	// Check that timeout occurred
	err := ctx.CheckTimeout()
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}

	if !strings.Contains(err.Error(), "Maximum execution time") {
		t.Errorf("Expected timeout error message, got: %v", err)
	}

	if !strings.Contains(err.Error(), "1 seconds") {
		t.Errorf("Expected timeout error to mention 1 second, got: %v", err)
	}
}

// TestTimeoutReset tests that timeout can be reset/extended
func TestTimeoutReset(t *testing.T) {
	ctx := NewExecutionContext()

	// Set initial timeout
	success := ctx.SetTimeLimit(2)
	if !success {
		t.Fatal("Failed to set initial time limit")
	}

	// Verify timeout is set
	if ctx.GetMaxExecutionTime() != 2 {
		t.Errorf("Expected max execution time 2, got %d", ctx.GetMaxExecutionTime())
	}

	// Reset to unlimited
	success = ctx.SetTimeLimit(0)
	if !success {
		t.Fatal("Failed to reset time limit to unlimited")
	}

	// Verify timeout is now unlimited
	if ctx.GetMaxExecutionTime() != 0 {
		t.Errorf("Expected unlimited execution time (0), got %d", ctx.GetMaxExecutionTime())
	}

	// Should not timeout now
	err := ctx.CheckTimeout()
	if err != nil {
		t.Errorf("Expected no timeout with unlimited time, got: %v", err)
	}
}

// TestCheckTimeoutBehavior tests CheckTimeout method behavior
func TestCheckTimeoutBehavior(t *testing.T) {
	ctx := NewExecutionContext()

	// With unlimited time, should never timeout
	err := ctx.CheckTimeout()
	if err != nil {
		t.Errorf("Expected no timeout with unlimited time, got: %v", err)
	}

	// Set a timeout but don't wait
	ctx.SetTimeLimit(10)
	err = ctx.CheckTimeout()
	if err != nil {
		t.Errorf("Expected no timeout before deadline, got: %v", err)
	}
}

// TestCancelExecution tests context cancellation
func TestCancelExecution(t *testing.T) {
	ctx := NewExecutionContext()

	// Set a timeout
	ctx.SetTimeLimit(10)

	// Cancel the context manually
	ctx.Cancel()

	// Check that context is cancelled
	err := ctx.CheckTimeout()
	if err == nil {
		t.Error("Expected error after cancelling context, got nil")
	}
}