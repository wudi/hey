package runtime

import (
	"strings"
	"testing"
	"time"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// TestDateTimeObjectFunctions tests the DateTime object-like functions
func TestDateTimeObjectFunctions(t *testing.T) {
	functions := GetDateTimeObjectFunctions()
	functionMap := make(map[string]*registry.Function)
	for _, fn := range functions {
		functionMap[fn.Name] = fn
	}

	t.Run("DateTime_create", func(t *testing.T) {
		fn := functionMap["DateTime_create"]
		if fn == nil {
			t.Fatal("DateTime_create function not found")
		}

		// Test default (now) creation
		result, err := fn.Builtin(nil, []*values.Value{})
		if err != nil {
			t.Errorf("DateTime_create error: %v", err)
			return
		}

		if !result.IsArray() {
			t.Error("DateTime_create should return array")
			return
		}

		// Check required fields
		timestamp := result.ArrayGet(values.NewString("_timestamp"))
		timezone := result.ArrayGet(values.NewString("_timezone"))
		typeField := result.ArrayGet(values.NewString("_type"))

		if timestamp == nil || timestamp.IsNull() {
			t.Error("DateTime_create missing _timestamp")
		}
		if timezone == nil || timezone.IsNull() {
			t.Error("DateTime_create missing _timezone")
		}
		if typeField == nil || typeField.ToString() != "DateTime" {
			t.Error("DateTime_create missing or wrong _type")
		}

		// Test with specific date
		result2, err := fn.Builtin(nil, []*values.Value{
			values.NewString("2023-06-15 14:30:45"),
			values.NewString("UTC"),
		})

		if err != nil {
			t.Errorf("DateTime_create with args error: %v", err)
			return
		}

		timestamp2 := result2.ArrayGet(values.NewString("_timestamp"))
		expectedTimestamp := time.Date(2023, 6, 15, 14, 30, 45, 0, time.UTC).Unix()
		if timestamp2.ToInt() != expectedTimestamp {
			t.Errorf("expected timestamp %d, got %d", expectedTimestamp, timestamp2.ToInt())
		}
	})

	t.Run("DateTime_format", func(t *testing.T) {
		createFn := functionMap["DateTime_create"]
		formatFn := functionMap["DateTime_format"]
		if createFn == nil || formatFn == nil {
			t.Fatal("Required functions not found")
		}

		// Create DateTime object
		dtObj, err := createFn.Builtin(nil, []*values.Value{
			values.NewString("2023-06-15 14:30:45"),
		})
		if err != nil {
			t.Errorf("DateTime_create error: %v", err)
			return
		}

		// Format it
		result, err := formatFn.Builtin(nil, []*values.Value{
			dtObj,
			values.NewString("Y-m-d H:i:s"),
		})

		if err != nil {
			t.Errorf("DateTime_format error: %v", err)
			return
		}

		expected := "2023-06-15 14:30:45"
		if result.ToString() != expected {
			t.Errorf("expected %s, got %s", expected, result.ToString())
		}
	})

	t.Run("DateTime_getTimestamp", func(t *testing.T) {
		createFn := functionMap["DateTime_create"]
		getTimestampFn := functionMap["DateTime_getTimestamp"]
		if createFn == nil || getTimestampFn == nil {
			t.Fatal("Required functions not found")
		}

		// Create DateTime object
		dtObj, err := createFn.Builtin(nil, []*values.Value{
			values.NewString("2023-06-15 14:30:45"),
		})
		if err != nil {
			t.Errorf("DateTime_create error: %v", err)
			return
		}

		// Get timestamp
		result, err := getTimestampFn.Builtin(nil, []*values.Value{dtObj})
		if err != nil {
			t.Errorf("DateTime_getTimestamp error: %v", err)
			return
		}

		expectedTimestamp := time.Date(2023, 6, 15, 14, 30, 45, 0, time.UTC).Unix()
		if result.ToInt() != expectedTimestamp {
			t.Errorf("expected timestamp %d, got %d", expectedTimestamp, result.ToInt())
		}
	})

	t.Run("DateTime_setTimestamp", func(t *testing.T) {
		createFn := functionMap["DateTime_create"]
		setTimestampFn := functionMap["DateTime_setTimestamp"]
		formatFn := functionMap["DateTime_format"]
		if createFn == nil || setTimestampFn == nil || formatFn == nil {
			t.Fatal("Required functions not found")
		}

		// Create DateTime object
		dtObj, err := createFn.Builtin(nil, []*values.Value{})
		if err != nil {
			t.Errorf("DateTime_create error: %v", err)
			return
		}

		// Set timestamp
		testTimestamp := int64(1640995200) // 2022-01-01 00:00:00 UTC
		result, err := setTimestampFn.Builtin(nil, []*values.Value{
			dtObj,
			values.NewInt(testTimestamp),
		})

		if err != nil {
			t.Errorf("DateTime_setTimestamp error: %v", err)
			return
		}

		// Verify by formatting
		formatted, err := formatFn.Builtin(nil, []*values.Value{
			result,
			values.NewString("Y-m-d H:i:s"),
		})

		if err != nil {
			t.Errorf("DateTime_format after set error: %v", err)
			return
		}

		expected := "2022-01-01 00:00:00"
		if formatted.ToString() != expected {
			t.Errorf("after setTimestamp: expected %s, got %s", expected, formatted.ToString())
		}
	})

	t.Run("DateTime_createFromFormat", func(t *testing.T) {
		fn := functionMap["DateTime_createFromFormat"]
		formatFn := functionMap["DateTime_format"]
		if fn == nil || formatFn == nil {
			t.Fatal("Required functions not found")
		}

		// Create from format
		result, err := fn.Builtin(nil, []*values.Value{
			values.NewString("d/m/Y H:i"),
			values.NewString("15/06/2023 14:30"),
		})

		if err != nil {
			t.Errorf("DateTime_createFromFormat error: %v", err)
			return
		}

		if result.IsBool() && !result.ToBool() {
			t.Error("DateTime_createFromFormat returned false")
			return
		}

		// Format the result
		formatted, err := formatFn.Builtin(nil, []*values.Value{
			result,
			values.NewString("Y-m-d H:i:s"),
		})

		if err != nil {
			t.Errorf("DateTime_format error: %v", err)
			return
		}

		expected := "2023-06-15 14:30:00"
		if formatted.ToString() != expected {
			t.Errorf("createFromFormat: expected %s, got %s", expected, formatted.ToString())
		}
	})

	t.Run("DateInterval_create", func(t *testing.T) {
		fn := functionMap["DateInterval_create"]
		if fn == nil {
			t.Fatal("DateInterval_create function not found")
		}

		// Test P1Y2M3DT4H5M6S
		result, err := fn.Builtin(nil, []*values.Value{
			values.NewString("P1Y2M3DT4H5M6S"),
		})

		if err != nil {
			t.Errorf("DateInterval_create error: %v", err)
			return
		}

		if !result.IsArray() {
			t.Error("DateInterval_create should return array")
			return
		}

		// Check values
		years := result.ArrayGet(values.NewString("y"))
		months := result.ArrayGet(values.NewString("m"))
		days := result.ArrayGet(values.NewString("d"))

		if years.ToInt() != 1 {
			t.Errorf("expected 1 year, got %d", years.ToInt())
		}
		if months.ToInt() != 2 {
			t.Errorf("expected 2 months, got %d", months.ToInt())
		}
		if days.ToInt() != 3 {
			t.Errorf("expected 3 days, got %d", days.ToInt())
		}
	})

	t.Run("DateInterval_format", func(t *testing.T) {
		createFn := functionMap["DateInterval_create"]
		formatFn := functionMap["DateInterval_format"]
		if createFn == nil || formatFn == nil {
			t.Fatal("Required functions not found")
		}

		// Create interval
		intervalObj, err := createFn.Builtin(nil, []*values.Value{
			values.NewString("P1Y2M3DT4H5M6S"),
		})

		if err != nil {
			t.Errorf("DateInterval_create error: %v", err)
			return
		}

		// Format interval
		result, err := formatFn.Builtin(nil, []*values.Value{
			intervalObj,
			values.NewString("%y years, %m months, %d days, %h hours, %i minutes, %s seconds"),
		})

		if err != nil {
			t.Errorf("DateInterval_format error: %v", err)
			return
		}

		expected := "1 years, 2 months, 3 days, 4 hours, 5 minutes, 6 seconds"
		if result.ToString() != expected {
			t.Errorf("expected %s, got %s", expected, result.ToString())
		}
	})

	t.Run("DateTime_add", func(t *testing.T) {
		createDateFn := functionMap["DateTime_create"]
		createIntervalFn := functionMap["DateInterval_create"]
		addFn := functionMap["DateTime_add"]
		formatFn := functionMap["DateTime_format"]

		if createDateFn == nil || createIntervalFn == nil || addFn == nil || formatFn == nil {
			t.Fatal("Required functions not found")
		}

		// Create DateTime
		dtObj, err := createDateFn.Builtin(nil, []*values.Value{
			values.NewString("2023-06-15 14:30:45"),
		})
		if err != nil {
			t.Errorf("DateTime_create error: %v", err)
			return
		}

		// Create DateInterval
		intervalObj, err := createIntervalFn.Builtin(nil, []*values.Value{
			values.NewString("P1D"),
		})
		if err != nil {
			t.Errorf("DateInterval_create error: %v", err)
			return
		}

		// Add interval
		result, err := addFn.Builtin(nil, []*values.Value{dtObj, intervalObj})
		if err != nil {
			t.Errorf("DateTime_add error: %v", err)
			return
		}

		// Format result
		formatted, err := formatFn.Builtin(nil, []*values.Value{
			result,
			values.NewString("Y-m-d H:i:s"),
		})

		if err != nil {
			t.Errorf("DateTime_format error: %v", err)
			return
		}

		expected := "2023-06-16 14:30:45"
		if formatted.ToString() != expected {
			t.Errorf("after add: expected %s, got %s", expected, formatted.ToString())
		}
	})

	t.Run("DateTime_diff", func(t *testing.T) {
		createFn := functionMap["DateTime_create"]
		diffFn := functionMap["DateTime_diff"]
		intervalFormatFn := functionMap["DateInterval_format"]

		if createFn == nil || diffFn == nil || intervalFormatFn == nil {
			t.Fatal("Required functions not found")
		}

		// Create two DateTime objects
		dt1, err := createFn.Builtin(nil, []*values.Value{
			values.NewString("2023-06-15 14:30:45"),
		})
		if err != nil {
			t.Errorf("DateTime_create error: %v", err)
			return
		}

		dt2, err := createFn.Builtin(nil, []*values.Value{
			values.NewString("2023-06-20 18:45:30"),
		})
		if err != nil {
			t.Errorf("DateTime_create error: %v", err)
			return
		}

		// Calculate difference
		diffResult, err := diffFn.Builtin(nil, []*values.Value{dt1, dt2})
		if err != nil {
			t.Errorf("DateTime_diff error: %v", err)
			return
		}

		// Format the difference
		formatted, err := intervalFormatFn.Builtin(nil, []*values.Value{
			diffResult,
			values.NewString("%R%d days, %h hours, %i minutes, %s seconds"),
		})

		if err != nil {
			t.Errorf("DateInterval_format error: %v", err)
			return
		}

		// Should be positive difference with some days/hours
		result := formatted.ToString()
		if !strings.Contains(result, "+") || !strings.Contains(result, "days") {
			t.Errorf("unexpected diff format: %s", result)
		}
	})
}

// TestStrftimeFunctions tests the strftime functions
func TestStrftimeFunctions(t *testing.T) {
	functions := GetDateTimeFunctions()
	functionMap := make(map[string]*registry.Function)
	for _, fn := range functions {
		functionMap[fn.Name] = fn
	}

	t.Run("strftime", func(t *testing.T) {
		fn := functionMap["strftime"]
		if fn == nil {
			t.Fatal("strftime function not found")
		}

		// Test with known timestamp
		timestamp := int64(1640995200) // 2022-01-01 00:00:00 UTC
		result, err := fn.Builtin(nil, []*values.Value{
			values.NewString("%Y-%m-%d %H:%M:%S"),
			values.NewInt(timestamp),
		})

		if err != nil {
			t.Errorf("strftime error: %v", err)
			return
		}

		expected := "2022-01-01 00:00:00"
		if result.ToString() != expected {
			t.Errorf("strftime: expected %s, got %s", expected, result.ToString())
		}
	})

	t.Run("gmstrftime", func(t *testing.T) {
		fn := functionMap["gmstrftime"]
		if fn == nil {
			t.Fatal("gmstrftime function not found")
		}

		// Test with known timestamp
		timestamp := int64(1640995200) // 2022-01-01 00:00:00 UTC
		result, err := fn.Builtin(nil, []*values.Value{
			values.NewString("%Y-%m-%d %H:%M:%S"),
			values.NewInt(timestamp),
		})

		if err != nil {
			t.Errorf("gmstrftime error: %v", err)
			return
		}

		expected := "2022-01-01 00:00:00"
		if result.ToString() != expected {
			t.Errorf("gmstrftime: expected %s, got %s", expected, result.ToString())
		}
	})

	t.Run("strftime_formats", func(t *testing.T) {
		fn := functionMap["strftime"]
		if fn == nil {
			t.Fatal("strftime function not found")
		}

		timestamp := int64(1640995200) // 2022-01-01 00:00:00 UTC Saturday

		tests := []struct {
			format   string
			expected string
		}{
			{"%Y", "2022"},     // 4-digit year
			{"%y", "22"},       // 2-digit year
			{"%m", "01"},       // Month
			{"%d", "01"},       // Day
			{"%H", "00"},       // Hour
			{"%M", "00"},       // Minute
			{"%S", "00"},       // Second
			{"%w", "6"},        // Weekday (0=Sunday, 6=Saturday)
			{"%A", "Saturday"}, // Full weekday name
			{"%B", "January"},  // Full month name
		}

		for _, tt := range tests {
			t.Run("format_"+tt.format, func(t *testing.T) {
				result, err := fn.Builtin(nil, []*values.Value{
					values.NewString(tt.format),
					values.NewInt(timestamp),
				})

				if err != nil {
					t.Errorf("strftime error: %v", err)
					return
				}

				if result.ToString() != tt.expected {
					t.Errorf("format %s: expected %s, got %s", tt.format, tt.expected, result.ToString())
				}
			})
		}
	})
}