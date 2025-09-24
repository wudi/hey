package runtime

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// TestDateTimeFunctions tests all date/time related functions
func TestDateTimeFunctions(t *testing.T) {
	functions := GetDateTimeFunctions()
	functionMap := make(map[string]*registry.Function)
	for _, fn := range functions {
		functionMap[fn.Name] = fn
	}

	t.Run("date", func(t *testing.T) {
		fn := functionMap["date"]
		if fn == nil {
			t.Fatal("date function not found")
		}

		tests := []struct {
			name      string
			args      []*values.Value
			expected  func(string) bool
			shouldErr bool
		}{
			{
				name: "Y-m-d format",
				args: []*values.Value{values.NewString("Y-m-d")},
				expected: func(result string) bool {
					// Should match YYYY-MM-DD format
					parts := strings.Split(result, "-")
					return len(parts) == 3 && len(parts[0]) == 4 && len(parts[1]) == 2 && len(parts[2]) == 2
				},
			},
			{
				name: "Y-m-d H:i:s format",
				args: []*values.Value{values.NewString("Y-m-d H:i:s")},
				expected: func(result string) bool {
					// Should match YYYY-MM-DD HH:II:SS format
					return len(result) == 19 && strings.Contains(result, " ")
				},
			},
			{
				name: "with specific timestamp",
				args: []*values.Value{values.NewString("Y-m-d H:i:s"), values.NewInt(1640995200)},
				expected: func(result string) bool {
					return result == "2022-01-01 00:00:00"
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)

				if tt.shouldErr {
					if err == nil {
						t.Error("expected error but got none")
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

				resultStr := result.ToString()
				if !tt.expected(resultStr) {
					t.Errorf("unexpected result: %s", resultStr)
				}
			})
		}
	})

	t.Run("gmdate", func(t *testing.T) {
		fn := functionMap["gmdate"]
		if fn == nil {
			t.Fatal("gmdate function not found")
		}

		// Test with known timestamp
		result, err := fn.Builtin(nil, []*values.Value{
			values.NewString("Y-m-d H:i:s"),
			values.NewInt(1640995200),
		})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		if result == nil {
			t.Error("result is nil")
			return
		}

		expected := "2022-01-01 00:00:00"
		if result.ToString() != expected {
			t.Errorf("expected %s, got %s", expected, result.ToString())
		}
	})

	t.Run("mktime", func(t *testing.T) {
		fn := functionMap["mktime"]
		if fn == nil {
			t.Fatal("mktime function not found")
		}

		// Test mktime(14, 30, 45, 6, 15, 2023)
		result, err := fn.Builtin(nil, []*values.Value{
			values.NewInt(14), // hour
			values.NewInt(30), // minute
			values.NewInt(45), // second
			values.NewInt(6),  // month
			values.NewInt(15), // day
			values.NewInt(2023), // year
		})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		if result == nil {
			t.Error("result is nil")
			return
		}

		// Verify the timestamp corresponds to the expected date
		timestamp := result.ToInt()
		expectedTime := time.Date(2023, 6, 15, 14, 30, 45, 0, time.Local)
		expectedTimestamp := expectedTime.Unix()

		if timestamp != expectedTimestamp {
			t.Errorf("expected timestamp %d, got %d", expectedTimestamp, timestamp)
		}
	})

	t.Run("gmmktime", func(t *testing.T) {
		fn := functionMap["gmmktime"]
		if fn == nil {
			t.Fatal("gmmktime function not found")
		}

		// Test gmmktime(14, 30, 45, 6, 15, 2023)
		result, err := fn.Builtin(nil, []*values.Value{
			values.NewInt(14), // hour
			values.NewInt(30), // minute
			values.NewInt(45), // second
			values.NewInt(6),  // month
			values.NewInt(15), // day
			values.NewInt(2023), // year
		})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		if result == nil {
			t.Error("result is nil")
			return
		}

		// Verify the timestamp corresponds to the expected UTC date
		timestamp := result.ToInt()
		expectedTime := time.Date(2023, 6, 15, 14, 30, 45, 0, time.UTC)
		expectedTimestamp := expectedTime.Unix()

		if timestamp != expectedTimestamp {
			t.Errorf("expected timestamp %d, got %d", expectedTimestamp, timestamp)
		}
	})

	t.Run("checkdate", func(t *testing.T) {
		fn := functionMap["checkdate"]
		if fn == nil {
			t.Fatal("checkdate function not found")
		}

		tests := []struct {
			name     string
			month    int
			day      int
			year     int
			expected bool
		}{
			{"valid date", 12, 31, 2023, true},
			{"invalid leap year", 2, 29, 2023, false},
			{"valid leap year", 2, 29, 2024, true},
			{"invalid month", 13, 1, 2023, false},
			{"invalid day", 6, 31, 2023, false}, // June has 30 days
			{"valid date in June", 6, 30, 2023, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, []*values.Value{
					values.NewInt(int64(tt.month)),
					values.NewInt(int64(tt.day)),
					values.NewInt(int64(tt.year)),
				})

				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				if result == nil {
					t.Error("result is nil")
					return
				}

				if result.ToBool() != tt.expected {
					t.Errorf("expected %t, got %t", tt.expected, result.ToBool())
				}
			})
		}
	})

	t.Run("getdate", func(t *testing.T) {
		fn := functionMap["getdate"]
		if fn == nil {
			t.Fatal("getdate function not found")
		}

		// Test with known timestamp (2022-01-01 00:00:00 UTC = Saturday)
		result, err := fn.Builtin(nil, []*values.Value{values.NewInt(1640995200)})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		if result == nil {
			t.Error("result is nil")
			return
		}

		if !result.IsArray() {
			t.Error("expected array result")
			return
		}

		arr := result

		// Check required keys
		requiredKeys := []string{"seconds", "minutes", "hours", "mday", "wday", "mon", "year", "yday", "weekday", "month", "0"}
		for _, key := range requiredKeys {
			val := arr.ArrayGet(values.NewString(key))
			if val == nil || val.IsNull() {
				t.Errorf("missing key: %s", key)
			}
		}

		// Check specific values for known timestamp
		val := arr.ArrayGet(values.NewString("year"))
		if val != nil && !val.IsNull() {
			if val.ToInt() != 2022 {
				t.Errorf("expected year 2022, got %d", val.ToInt())
			}
		}

		val = arr.ArrayGet(values.NewString("mon"))
		if val != nil && !val.IsNull() {
			if val.ToInt() != 1 {
				t.Errorf("expected month 1, got %d", val.ToInt())
			}
		}

		val = arr.ArrayGet(values.NewString("mday"))
		if val != nil && !val.IsNull() {
			if val.ToInt() != 1 {
				t.Errorf("expected day 1, got %d", val.ToInt())
			}
		}
	})

	t.Run("localtime", func(t *testing.T) {
		fn := functionMap["localtime"]
		if fn == nil {
			t.Fatal("localtime function not found")
		}

		// Test indexed array (default)
		result, err := fn.Builtin(nil, []*values.Value{values.NewInt(1640995200)})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		if result == nil {
			t.Error("result is nil")
			return
		}

		if !result.IsArray() {
			t.Error("expected array result")
			return
		}

		// Test associative array
		result2, err := fn.Builtin(nil, []*values.Value{
			values.NewInt(1640995200),
			values.NewBool(true),
		})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		if result2 == nil {
			t.Error("result2 is nil")
			return
		}

		if !result2.IsArray() {
			t.Error("expected array result2")
			return
		}

		arr := result2

		// Check associative keys
		requiredKeys := []string{"tm_sec", "tm_min", "tm_hour", "tm_mday", "tm_mon", "tm_year", "tm_wday", "tm_yday", "tm_isdst"}
		for _, key := range requiredKeys {
			val := arr.ArrayGet(values.NewString(key))
			if val == nil || val.IsNull() {
				t.Errorf("missing associative key: %s", key)
			}
		}
	})

	t.Run("strtotime", func(t *testing.T) {
		fn := functionMap["strtotime"]
		if fn == nil {
			t.Fatal("strtotime function not found")
		}

		tests := []struct {
			name      string
			timeStr   string
			baseTime  *values.Value
			expected  func(int64) bool
			shouldErr bool
		}{
			{
				name:    "ISO format",
				timeStr: "2023-06-15 14:30:45",
				expected: func(result int64) bool {
					expected := time.Date(2023, 6, 15, 14, 30, 45, 0, time.Local).Unix()
					return result == expected
				},
			},
			{
				name:    "now keyword",
				timeStr: "now",
				expected: func(result int64) bool {
					now := time.Now().Unix()
					return result >= now-1 && result <= now+1 // Allow 1 second tolerance
				},
			},
			{
				name:    "relative time",
				timeStr: "+1 day",
				expected: func(result int64) bool {
					base := time.Now()
					expected := base.AddDate(0, 0, 1).Unix()
					return result >= expected-1 && result <= expected+1
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				args := []*values.Value{values.NewString(tt.timeStr)}
				if tt.baseTime != nil {
					args = append(args, tt.baseTime)
				}

				result, err := fn.Builtin(nil, args)

				if tt.shouldErr {
					if err == nil {
						t.Error("expected error but got none")
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

				if result.IsBool() && !result.ToBool() {
					// Function returned false (parse failure)
					t.Errorf("strtotime returned false for input: %s", tt.timeStr)
					return
				}

				timestamp := result.ToInt()
				if !tt.expected(timestamp) {
					t.Errorf("unexpected timestamp %d for input: %s", timestamp, tt.timeStr)
				}
			})
		}
	})
}

// TestDateTimeFormatting tests the formatDateTime function
func TestDateTimeFormatting(t *testing.T) {
	// Test with a known time: 2023-06-15 14:30:45
	testTime := time.Date(2023, 6, 15, 14, 30, 45, 0, time.UTC)

	tests := []struct {
		format   string
		expected string
	}{
		{"Y-m-d", "2023-06-15"},
		{"Y-m-d H:i:s", "2023-06-15 14:30:45"},
		{"l, F j, Y", "Thursday, June 15, 2023"},
		{"H:i:s", "14:30:45"},
		{"g:i A", "2:30 PM"},
		{"j/n/Y", "15/6/2023"},
		{"U", strconv.FormatInt(testTime.Unix(), 10)},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			result, err := formatDateTime(tt.format, testTime)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("format %s: expected %s, got %s", tt.format, tt.expected, result)
			}
		})
	}
}

// TestTimeStringParsing tests the parseTimeString function
func TestTimeStringParsing(t *testing.T) {
	baseTime := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		timeStr   string
		expected  func(time.Time) bool
		shouldErr bool
	}{
		{
			name:    "now",
			timeStr: "now",
			expected: func(result time.Time) bool {
				now := time.Now()
				diff := result.Sub(now).Abs()
				return diff < time.Second
			},
		},
		{
			name:    "ISO format",
			timeStr: "2023-06-15 14:30:45",
			expected: func(result time.Time) bool {
				expected := time.Date(2023, 6, 15, 14, 30, 45, 0, time.UTC)
				return result.Equal(expected)
			},
		},
		{
			name:    "+1 day",
			timeStr: "+1 day",
			expected: func(result time.Time) bool {
				expected := baseTime.AddDate(0, 0, 1)
				return result.Equal(expected)
			},
		},
		{
			name:    "next Monday",
			timeStr: "next Monday",
			expected: func(result time.Time) bool {
				return result.Weekday() == time.Monday && result.After(baseTime)
			},
		},
		{
			name:      "invalid format",
			timeStr:   "invalid time string",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseTimeString(tt.timeStr, baseTime)

			if tt.shouldErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if !tt.expected(result) {
				t.Errorf("unexpected result for %s: %v", tt.timeStr, result)
			}
		})
	}
}

// TestRemainingDateTimeFunctions tests the newly implemented functions
func TestRemainingDateTimeFunctions(t *testing.T) {
	functions := GetDateTimeFunctions()
	functionMap := make(map[string]*registry.Function)
	for _, fn := range functions {
		functionMap[fn.Name] = fn
	}

	t.Run("date_parse", func(t *testing.T) {
		fn := functionMap["date_parse"]
		if fn == nil {
			t.Fatal("date_parse function not found")
		}

		tests := []struct {
			name      string
			input     string
			shouldErr bool
		}{
			{"valid ISO date", "2023-06-15 14:30:45", false},
			{"invalid date", "invalid date", true},
			{"ISO date only", "2023-06-15", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, []*values.Value{values.NewString(tt.input)})

				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				if result == nil {
					t.Error("result is nil")
					return
				}

				if !result.IsArray() {
					t.Error("expected array result")
					return
				}

				// Check required keys
				requiredKeys := []string{"year", "month", "day", "hour", "minute", "second", "fraction", "warning_count", "warnings", "error_count", "errors", "is_localtime"}
				for _, key := range requiredKeys {
					val := result.ArrayGet(values.NewString(key))
					if val == nil {
						t.Errorf("missing key: %s", key)
					}
				}

				errorCount := result.ArrayGet(values.NewString("error_count"))
				if tt.shouldErr {
					if errorCount.ToInt() == 0 {
						t.Error("expected parsing error but got none")
					}
				} else {
					if errorCount.ToInt() > 0 {
						t.Error("unexpected parsing error")
					}
				}
			})
		}
	})

	t.Run("date_parse_from_format", func(t *testing.T) {
		fn := functionMap["date_parse_from_format"]
		if fn == nil {
			t.Fatal("date_parse_from_format function not found")
		}

		result, err := fn.Builtin(nil, []*values.Value{
			values.NewString("d/m/Y H:i"),
			values.NewString("15/06/2023 14:30"),
		})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		if result == nil {
			t.Error("result is nil")
			return
		}

		if !result.IsArray() {
			t.Error("expected array result")
			return
		}

		// Check that parsing succeeded
		errorCount := result.ArrayGet(values.NewString("error_count"))
		if errorCount.ToInt() > 0 {
			t.Error("unexpected parsing error")
		}

		// Check parsed values
		year := result.ArrayGet(values.NewString("year"))
		if year != nil && !year.IsNull() && year.ToInt() != 2023 {
			t.Errorf("expected year 2023, got %d", year.ToInt())
		}
	})

	t.Run("gettimeofday", func(t *testing.T) {
		fn := functionMap["gettimeofday"]
		if fn == nil {
			t.Fatal("gettimeofday function not found")
		}

		// Test array return
		result, err := fn.Builtin(nil, []*values.Value{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		if !result.IsArray() {
			t.Error("expected array result")
			return
		}

		// Check required keys
		requiredKeys := []string{"sec", "usec", "minuteswest", "dsttime"}
		for _, key := range requiredKeys {
			val := result.ArrayGet(values.NewString(key))
			if val == nil || val.IsNull() {
				t.Errorf("missing key: %s", key)
			}
		}

		// Test float return
		result2, err := fn.Builtin(nil, []*values.Value{values.NewBool(true)})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		if !result2.IsFloat() {
			t.Error("expected float result when return_float is true")
		}
	})

	t.Run("idate", func(t *testing.T) {
		fn := functionMap["idate"]
		if fn == nil {
			t.Fatal("idate function not found")
		}

		timestamp := int64(1640995200) // 2022-01-01 00:00:00 UTC

		tests := []struct {
			format   string
			expected int64
		}{
			{"Y", 2022},
			{"m", 1},
			{"d", 1},
			{"H", 0},
			{"i", 0},
			{"s", 0},
			{"w", 6}, // Saturday
			{"z", 0}, // Day 0 of year
		}

		for _, tt := range tests {
			t.Run(fmt.Sprintf("format_%s", tt.format), func(t *testing.T) {
				result, err := fn.Builtin(nil, []*values.Value{
					values.NewString(tt.format),
					values.NewInt(timestamp),
				})

				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				if result == nil {
					t.Error("result is nil")
					return
				}

				if result.ToInt() != tt.expected {
					t.Errorf("format %s: expected %d, got %d", tt.format, tt.expected, result.ToInt())
				}
			})
		}
	})

	t.Run("date_default_timezone_get", func(t *testing.T) {
		fn := functionMap["date_default_timezone_get"]
		if fn == nil {
			t.Fatal("date_default_timezone_get function not found")
		}

		result, err := fn.Builtin(nil, []*values.Value{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		if result == nil {
			t.Error("result is nil")
			return
		}

		if !result.IsString() {
			t.Error("expected string result")
			return
		}

		// Should return default timezone (UTC)
		if result.ToString() != "UTC" {
			t.Errorf("expected UTC, got %s", result.ToString())
		}
	})

	t.Run("date_default_timezone_set", func(t *testing.T) {
		fn := functionMap["date_default_timezone_set"]
		if fn == nil {
			t.Fatal("date_default_timezone_set function not found")
		}

		getFn := functionMap["date_default_timezone_get"]
		if getFn == nil {
			t.Fatal("date_default_timezone_get function not found")
		}

		// Test setting valid timezone
		result, err := fn.Builtin(nil, []*values.Value{values.NewString("America/New_York")})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		if !result.ToBool() {
			t.Error("expected true for valid timezone")
		}

		// Verify timezone was set
		getResult, err := getFn.Builtin(nil, []*values.Value{})
		if err != nil {
			t.Errorf("unexpected error getting timezone: %v", err)
			return
		}

		if getResult.ToString() != "America/New_York" {
			t.Errorf("expected America/New_York, got %s", getResult.ToString())
		}

		// Test setting invalid timezone
		result2, err := fn.Builtin(nil, []*values.Value{values.NewString("Invalid/Timezone")})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		if result2.ToBool() {
			t.Error("expected false for invalid timezone")
		}

		// Reset to UTC
		fn.Builtin(nil, []*values.Value{values.NewString("UTC")})
	})

	t.Run("timezone_name_from_abbr", func(t *testing.T) {
		fn := functionMap["timezone_name_from_abbr"]
		if fn == nil {
			t.Fatal("timezone_name_from_abbr function not found")
		}

		tests := []struct {
			abbr     string
			expected string
		}{
			{"EST", "America/New_York"},
			{"PST", "America/Los_Angeles"},
			{"UTC", "UTC"},
			{"GMT", "UTC"},
			{"INVALID", ""}, // Should return false
		}

		for _, tt := range tests {
			t.Run(fmt.Sprintf("abbr_%s", tt.abbr), func(t *testing.T) {
				result, err := fn.Builtin(nil, []*values.Value{values.NewString(tt.abbr)})

				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				if result == nil {
					t.Error("result is nil")
					return
				}

				if tt.expected == "" {
					// Should return false for invalid abbreviations
					if result.ToBool() {
						t.Errorf("expected false for invalid abbreviation %s", tt.abbr)
					}
				} else {
					if result.ToString() != tt.expected {
						t.Errorf("abbr %s: expected %s, got %s", tt.abbr, tt.expected, result.ToString())
					}
				}
			})
		}
	})

	t.Run("strptime", func(t *testing.T) {
		fn := functionMap["strptime"]
		if fn == nil {
			t.Fatal("strptime function not found")
		}

		// Test basic date parsing
		result, err := fn.Builtin(nil, []*values.Value{
			values.NewString("2022-01-15 14:30:45"),
			values.NewString("%Y-%m-%d %H:%M:%S"),
		})

		if err != nil {
			t.Errorf("strptime error: %v", err)
			return
		}

		if !result.IsArray() {
			t.Error("strptime should return array")
			return
		}

		// Check parsed values
		tmYear := result.ArrayGet(values.NewString("tm_year"))
		tmMon := result.ArrayGet(values.NewString("tm_mon"))
		tmMday := result.ArrayGet(values.NewString("tm_mday"))
		tmHour := result.ArrayGet(values.NewString("tm_hour"))
		tmMin := result.ArrayGet(values.NewString("tm_min"))
		tmSec := result.ArrayGet(values.NewString("tm_sec"))

		if tmYear.ToInt() != 122 { // 2022 - 1900
			t.Errorf("expected tm_year 122, got %d", tmYear.ToInt())
		}
		if tmMon.ToInt() != 0 { // January = 0
			t.Errorf("expected tm_mon 0, got %d", tmMon.ToInt())
		}
		if tmMday.ToInt() != 15 {
			t.Errorf("expected tm_mday 15, got %d", tmMday.ToInt())
		}
		if tmHour.ToInt() != 14 {
			t.Errorf("expected tm_hour 14, got %d", tmHour.ToInt())
		}
		if tmMin.ToInt() != 30 {
			t.Errorf("expected tm_min 30, got %d", tmMin.ToInt())
		}
		if tmSec.ToInt() != 45 {
			t.Errorf("expected tm_sec 45, got %d", tmSec.ToInt())
		}

		// Test invalid date
		result2, err := fn.Builtin(nil, []*values.Value{
			values.NewString("invalid date"),
			values.NewString("%Y-%m-%d"),
		})

		if err != nil {
			t.Errorf("strptime should not error on invalid input: %v", err)
			return
		}

		if !result2.IsBool() || result2.ToBool() {
			t.Error("strptime should return false for invalid date")
		}
	})
}