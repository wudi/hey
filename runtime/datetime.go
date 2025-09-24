package runtime

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"regexp"
	"sync"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// Global timezone setting
var (
	defaultTimezoneMutex sync.RWMutex
	defaultTimezone     = "UTC"
)

// GetDateTimeFunctions returns all date/time related PHP functions
func GetDateTimeFunctions() []*registry.Function {
	return []*registry.Function{
		// Basic date functions
		{
			Name: "date",
			Parameters: []*registry.Parameter{
				{Name: "format", Type: "string"},
				{Name: "timestamp", Type: "int", HasDefault: true, DefaultValue: values.NewNull()},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				format := args[0].ToString()

				var t time.Time
				if len(args) > 1 && !args[1].IsNull() {
					timestamp := args[1].ToInt()
					t = time.Unix(timestamp, 0)
				} else {
					t = time.Now()
				}

				formatted, err := formatDateTime(format, t)
				if err != nil {
					return nil, err
				}
				return values.NewString(formatted), nil
			},
		},
		{
			Name: "gmdate",
			Parameters: []*registry.Parameter{
				{Name: "format", Type: "string"},
				{Name: "timestamp", Type: "int", HasDefault: true, DefaultValue: values.NewNull()},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				format := args[0].ToString()

				var t time.Time
				if len(args) > 1 && !args[1].IsNull() {
					timestamp := args[1].ToInt()
					t = time.Unix(timestamp, 0).UTC()
				} else {
					t = time.Now().UTC()
				}

				formatted, err := formatDateTime(format, t)
				if err != nil {
					return nil, err
				}
				return values.NewString(formatted), nil
			},
		},
		{
			Name: "mktime",
			Parameters: []*registry.Parameter{
				{Name: "hour", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
				{Name: "minute", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
				{Name: "second", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
				{Name: "month", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
				{Name: "day", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
				{Name: "year", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
			},
			ReturnType: "int|false",
			MinArgs:    0,
			MaxArgs:    6,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				now := time.Now()

				// Default values
				hour := now.Hour()
				minute := now.Minute()
				second := now.Second()
				month := int(now.Month())
				day := now.Day()
				year := now.Year()

				// Apply provided arguments
				if len(args) > 0 && !args[0].IsNull() {
					hour = int(args[0].ToInt())
				}
				if len(args) > 1 && !args[1].IsNull() {
					minute = int(args[1].ToInt())
				}
				if len(args) > 2 && !args[2].IsNull() {
					second = int(args[2].ToInt())
				}
				if len(args) > 3 && !args[3].IsNull() {
					month = int(args[3].ToInt())
				}
				if len(args) > 4 && !args[4].IsNull() {
					day = int(args[4].ToInt())
				}
				if len(args) > 5 && !args[5].IsNull() {
					year = int(args[5].ToInt())
				}

				// Create time and get Unix timestamp
				t := time.Date(year, time.Month(month), day, hour, minute, second, 0, time.Local)
				return values.NewInt(t.Unix()), nil
			},
		},
		{
			Name: "gmmktime",
			Parameters: []*registry.Parameter{
				{Name: "hour", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
				{Name: "minute", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
				{Name: "second", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
				{Name: "month", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
				{Name: "day", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
				{Name: "year", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
			},
			ReturnType: "int|false",
			MinArgs:    0,
			MaxArgs:    6,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				now := time.Now().UTC()

				// Default values
				hour := now.Hour()
				minute := now.Minute()
				second := now.Second()
				month := int(now.Month())
				day := now.Day()
				year := now.Year()

				// Apply provided arguments
				if len(args) > 0 && !args[0].IsNull() {
					hour = int(args[0].ToInt())
				}
				if len(args) > 1 && !args[1].IsNull() {
					minute = int(args[1].ToInt())
				}
				if len(args) > 2 && !args[2].IsNull() {
					second = int(args[2].ToInt())
				}
				if len(args) > 3 && !args[3].IsNull() {
					month = int(args[3].ToInt())
				}
				if len(args) > 4 && !args[4].IsNull() {
					day = int(args[4].ToInt())
				}
				if len(args) > 5 && !args[5].IsNull() {
					year = int(args[5].ToInt())
				}

				// Create time and get Unix timestamp
				t := time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC)
				return values.NewInt(t.Unix()), nil
			},
		},
		{
			Name: "checkdate",
			Parameters: []*registry.Parameter{
				{Name: "month", Type: "int"},
				{Name: "day", Type: "int"},
				{Name: "year", Type: "int"},
			},
			ReturnType: "bool",
			MinArgs:    3,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				month := int(args[0].ToInt())
				day := int(args[1].ToInt())
				year := int(args[2].ToInt())

				// Check year range
				if year < 1 || year > 32767 {
					return values.NewBool(false), nil
				}

				// Check month range
				if month < 1 || month > 12 {
					return values.NewBool(false), nil
				}

				// Check day range
				if day < 1 {
					return values.NewBool(false), nil
				}

				// Get the last day of the month
				t := time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC)
				lastDay := t.Day()

				if day > lastDay {
					return values.NewBool(false), nil
				}

				return values.NewBool(true), nil
			},
		},
		{
			Name: "getdate",
			Parameters: []*registry.Parameter{
				{Name: "timestamp", Type: "int", HasDefault: true, DefaultValue: values.NewNull()},
			},
			ReturnType: "array",
			MinArgs:    0,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				var t time.Time
				if len(args) > 0 && !args[0].IsNull() {
					timestamp := args[0].ToInt()
					t = time.Unix(timestamp, 0)
				} else {
					t = time.Now()
				}

				result := values.NewArray()
				result.ArraySet(values.NewString("seconds"), values.NewInt(int64(t.Second())))
				result.ArraySet(values.NewString("minutes"), values.NewInt(int64(t.Minute())))
				result.ArraySet(values.NewString("hours"), values.NewInt(int64(t.Hour())))
				result.ArraySet(values.NewString("mday"), values.NewInt(int64(t.Day())))
				result.ArraySet(values.NewString("wday"), values.NewInt(int64(t.Weekday())))
				result.ArraySet(values.NewString("mon"), values.NewInt(int64(t.Month())))
				result.ArraySet(values.NewString("year"), values.NewInt(int64(t.Year())))
				result.ArraySet(values.NewString("yday"), values.NewInt(int64(t.YearDay()-1))) // 0-indexed
				result.ArraySet(values.NewString("weekday"), values.NewString(t.Weekday().String()))
				result.ArraySet(values.NewString("month"), values.NewString(t.Month().String()))
				result.ArraySet(values.NewString("0"), values.NewInt(t.Unix()))

				return result, nil
			},
		},
		{
			Name: "localtime",
			Parameters: []*registry.Parameter{
				{Name: "timestamp", Type: "int", HasDefault: true, DefaultValue: values.NewNull()},
				{Name: "is_associative", Type: "bool", HasDefault: true, DefaultValue: values.NewBool(false)},
			},
			ReturnType: "array",
			MinArgs:    0,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				var t time.Time
				if len(args) > 0 && !args[0].IsNull() {
					timestamp := args[0].ToInt()
					t = time.Unix(timestamp, 0)
				} else {
					t = time.Now()
				}

				isAssociative := false
				if len(args) > 1 {
					isAssociative = args[1].ToBool()
				}

				result := values.NewArray()

				if isAssociative {
					result.ArraySet(values.NewString("tm_sec"), values.NewInt(int64(t.Second())))
					result.ArraySet(values.NewString("tm_min"), values.NewInt(int64(t.Minute())))
					result.ArraySet(values.NewString("tm_hour"), values.NewInt(int64(t.Hour())))
					result.ArraySet(values.NewString("tm_mday"), values.NewInt(int64(t.Day())))
					result.ArraySet(values.NewString("tm_mon"), values.NewInt(int64(t.Month()-1))) // 0-indexed
					result.ArraySet(values.NewString("tm_year"), values.NewInt(int64(t.Year()-1900))) // Years since 1900
					result.ArraySet(values.NewString("tm_wday"), values.NewInt(int64(t.Weekday())))
					result.ArraySet(values.NewString("tm_yday"), values.NewInt(int64(t.YearDay()-1))) // 0-indexed
					result.ArraySet(values.NewString("tm_isdst"), values.NewInt(0)) // DST info not available in Go
				} else {
					result.ArraySet(values.NewInt(0), values.NewInt(int64(t.Second())))
					result.ArraySet(values.NewInt(1), values.NewInt(int64(t.Minute())))
					result.ArraySet(values.NewInt(2), values.NewInt(int64(t.Hour())))
					result.ArraySet(values.NewInt(3), values.NewInt(int64(t.Day())))
					result.ArraySet(values.NewInt(4), values.NewInt(int64(t.Month()-1))) // 0-indexed
					result.ArraySet(values.NewInt(5), values.NewInt(int64(t.Year()-1900))) // Years since 1900
					result.ArraySet(values.NewInt(6), values.NewInt(int64(t.Weekday())))
					result.ArraySet(values.NewInt(7), values.NewInt(int64(t.YearDay()-1))) // 0-indexed
					result.ArraySet(values.NewInt(8), values.NewInt(0)) // DST info not available in Go
				}

				return result, nil
			},
		},
		{
			Name: "strtotime",
			Parameters: []*registry.Parameter{
				{Name: "time", Type: "string"},
				{Name: "now", Type: "int", HasDefault: true, DefaultValue: values.NewNull()},
			},
			ReturnType: "int|false",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				timeStr := args[0].ToString()

				var baseTime time.Time
				if len(args) > 1 && !args[1].IsNull() {
					timestamp := args[1].ToInt()
					baseTime = time.Unix(timestamp, 0)
				} else {
					baseTime = time.Now()
				}

				result, err := parseTimeString(timeStr, baseTime)
				if err != nil {
					return values.NewBool(false), nil
				}

				return values.NewInt(result.Unix()), nil
			},
		},
		{
			Name: "date_parse",
			Parameters: []*registry.Parameter{
				{Name: "date", Type: "string"},
			},
			ReturnType: "array",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				dateStr := args[0].ToString()

				result := values.NewArray()

				// Initialize all fields
				result.ArraySet(values.NewString("year"), values.NewBool(false))
				result.ArraySet(values.NewString("month"), values.NewBool(false))
				result.ArraySet(values.NewString("day"), values.NewBool(false))
				result.ArraySet(values.NewString("hour"), values.NewBool(false))
				result.ArraySet(values.NewString("minute"), values.NewBool(false))
				result.ArraySet(values.NewString("second"), values.NewBool(false))
				result.ArraySet(values.NewString("fraction"), values.NewFloat(0))
				result.ArraySet(values.NewString("warning_count"), values.NewInt(0))
				result.ArraySet(values.NewString("warnings"), values.NewArray())
				result.ArraySet(values.NewString("error_count"), values.NewInt(0))
				result.ArraySet(values.NewString("errors"), values.NewArray())
				result.ArraySet(values.NewString("is_localtime"), values.NewBool(false))

				// Try to parse the date
				parsedTime, err := parseTimeStringForParsing(dateStr)
				if err != nil {
					// Set error information
					errors := values.NewArray()
					errors.ArraySet(values.NewInt(0), values.NewString("The timezone could not be found in the database"))
					result.ArraySet(values.NewString("error_count"), values.NewInt(1))
					result.ArraySet(values.NewString("errors"), errors)
				} else {
					// Set parsed values
					result.ArraySet(values.NewString("year"), values.NewInt(int64(parsedTime.Year())))
					result.ArraySet(values.NewString("month"), values.NewInt(int64(parsedTime.Month())))
					result.ArraySet(values.NewString("day"), values.NewInt(int64(parsedTime.Day())))
					result.ArraySet(values.NewString("hour"), values.NewInt(int64(parsedTime.Hour())))
					result.ArraySet(values.NewString("minute"), values.NewInt(int64(parsedTime.Minute())))
					result.ArraySet(values.NewString("second"), values.NewInt(int64(parsedTime.Second())))
				}

				return result, nil
			},
		},
		{
			Name: "date_parse_from_format",
			Parameters: []*registry.Parameter{
				{Name: "format", Type: "string"},
				{Name: "date", Type: "string"},
			},
			ReturnType: "array",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				format := args[0].ToString()
				dateStr := args[1].ToString()

				result := values.NewArray()

				// Initialize all fields
				result.ArraySet(values.NewString("year"), values.NewBool(false))
				result.ArraySet(values.NewString("month"), values.NewBool(false))
				result.ArraySet(values.NewString("day"), values.NewBool(false))
				result.ArraySet(values.NewString("hour"), values.NewBool(false))
				result.ArraySet(values.NewString("minute"), values.NewBool(false))
				result.ArraySet(values.NewString("second"), values.NewBool(false))
				result.ArraySet(values.NewString("fraction"), values.NewFloat(0))
				result.ArraySet(values.NewString("warning_count"), values.NewInt(0))
				result.ArraySet(values.NewString("warnings"), values.NewArray())
				result.ArraySet(values.NewString("error_count"), values.NewInt(0))
				result.ArraySet(values.NewString("errors"), values.NewArray())
				result.ArraySet(values.NewString("is_localtime"), values.NewBool(false))

				// Convert PHP format to Go format
				goFormat, err := phpFormatToGoFormat(format)
				if err != nil {
					errors := values.NewArray()
					errors.ArraySet(values.NewInt(0), values.NewString("Invalid format string"))
					result.ArraySet(values.NewString("error_count"), values.NewInt(1))
					result.ArraySet(values.NewString("errors"), errors)
					return result, nil
				}

				// Try to parse the date
				parsedTime, err := time.Parse(goFormat, dateStr)
				if err != nil {
					errors := values.NewArray()
					errors.ArraySet(values.NewInt(0), values.NewString("Failed to parse date string"))
					result.ArraySet(values.NewString("error_count"), values.NewInt(1))
					result.ArraySet(values.NewString("errors"), errors)
				} else {
					// Set parsed values
					result.ArraySet(values.NewString("year"), values.NewInt(int64(parsedTime.Year())))
					result.ArraySet(values.NewString("month"), values.NewInt(int64(parsedTime.Month())))
					result.ArraySet(values.NewString("day"), values.NewInt(int64(parsedTime.Day())))
					result.ArraySet(values.NewString("hour"), values.NewInt(int64(parsedTime.Hour())))
					result.ArraySet(values.NewString("minute"), values.NewInt(int64(parsedTime.Minute())))
					result.ArraySet(values.NewString("second"), values.NewInt(int64(parsedTime.Second())))
				}

				return result, nil
			},
		},
		{
			Name: "gettimeofday",
			Parameters: []*registry.Parameter{
				{Name: "return_float", Type: "bool", HasDefault: true, DefaultValue: values.NewBool(false)},
			},
			ReturnType: "array|float",
			MinArgs:    0,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				returnFloat := false
				if len(args) > 0 {
					returnFloat = args[0].ToBool()
				}

				now := time.Now()

				if returnFloat {
					// Return as float: seconds.microseconds
					floatValue := float64(now.Unix()) + float64(now.Nanosecond())/1e9
					return values.NewFloat(floatValue), nil
				}

				// Return as array
				result := values.NewArray()
				result.ArraySet(values.NewString("sec"), values.NewInt(now.Unix()))
				result.ArraySet(values.NewString("usec"), values.NewInt(int64(now.Nanosecond()/1000)))
				result.ArraySet(values.NewString("minuteswest"), values.NewInt(0)) // UTC offset in minutes
				result.ArraySet(values.NewString("dsttime"), values.NewInt(0))     // DST flag

				return result, nil
			},
		},
		{
			Name: "idate",
			Parameters: []*registry.Parameter{
				{Name: "format", Type: "string"},
				{Name: "timestamp", Type: "int", HasDefault: true, DefaultValue: values.NewNull()},
			},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				format := args[0].ToString()

				var t time.Time
				if len(args) > 1 && !args[1].IsNull() {
					timestamp := args[1].ToInt()
					t = time.Unix(timestamp, 0)
				} else {
					t = time.Now()
				}

				if len(format) == 0 {
					return nil, fmt.Errorf("idate(): Format string cannot be empty")
				}

				formatChar := format[0]
				switch formatChar {
				case 'Y': // 4-digit year
					return values.NewInt(int64(t.Year())), nil
				case 'y': // 2-digit year
					return values.NewInt(int64(t.Year() % 100)), nil
				case 'm': // Month (01-12)
					return values.NewInt(int64(t.Month())), nil
				case 'n': // Month (1-12)
					return values.NewInt(int64(t.Month())), nil
				case 'd': // Day (01-31)
					return values.NewInt(int64(t.Day())), nil
				case 'j': // Day (1-31)
					return values.NewInt(int64(t.Day())), nil
				case 'H': // Hour (00-23)
					return values.NewInt(int64(t.Hour())), nil
				case 'G': // Hour (0-23)
					return values.NewInt(int64(t.Hour())), nil
				case 'h': // Hour (01-12)
					hour := t.Hour()
					if hour == 0 {
						hour = 12
					} else if hour > 12 {
						hour -= 12
					}
					return values.NewInt(int64(hour)), nil
				case 'g': // Hour (1-12)
					hour := t.Hour()
					if hour == 0 {
						hour = 12
					} else if hour > 12 {
						hour -= 12
					}
					return values.NewInt(int64(hour)), nil
				case 'i': // Minutes (00-59)
					return values.NewInt(int64(t.Minute())), nil
				case 's': // Seconds (00-59)
					return values.NewInt(int64(t.Second())), nil
				case 'U': // Unix timestamp
					return values.NewInt(t.Unix()), nil
				case 'w': // Day of week (0=Sunday)
					return values.NewInt(int64(t.Weekday())), nil
				case 'z': // Day of year (0-365)
					return values.NewInt(int64(t.YearDay() - 1)), nil
				case 'W': // Week number (ISO-8601)
					_, week := t.ISOWeek()
					return values.NewInt(int64(week)), nil
				case 't': // Number of days in month
					lastDay := time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, t.Location()).Day()
					return values.NewInt(int64(lastDay)), nil
				case 'L': // Leap year (1 if leap year, 0 otherwise)
					year := t.Year()
					if (year%4 == 0 && year%100 != 0) || year%400 == 0 {
						return values.NewInt(1), nil
					}
					return values.NewInt(0), nil
				default:
					return nil, fmt.Errorf("idate(): Unrecognized date format token")
				}
			},
		},
		{
			Name:       "date_default_timezone_get",
			Parameters: []*registry.Parameter{},
			ReturnType: "string",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				defaultTimezoneMutex.RLock()
				defer defaultTimezoneMutex.RUnlock()
				return values.NewString(defaultTimezone), nil
			},
		},
		{
			Name: "date_default_timezone_set",
			Parameters: []*registry.Parameter{
				{Name: "timezone_identifier", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				timezone := args[0].ToString()

				// Validate timezone
				_, err := time.LoadLocation(timezone)
				if err != nil {
					return values.NewBool(false), nil
				}

				defaultTimezoneMutex.Lock()
				defaultTimezone = timezone
				defaultTimezoneMutex.Unlock()

				return values.NewBool(true), nil
			},
		},
		{
			Name: "timezone_name_from_abbr",
			Parameters: []*registry.Parameter{
				{Name: "abbr", Type: "string"},
				{Name: "gmtOffset", Type: "int", HasDefault: true, DefaultValue: values.NewInt(-1)},
				{Name: "isdst", Type: "int", HasDefault: true, DefaultValue: values.NewInt(-1)},
			},
			ReturnType: "string|false",
			MinArgs:    1,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				abbr := strings.ToUpper(args[0].ToString())

				// Map common timezone abbreviations to timezone identifiers
				timezoneMap := map[string]string{
					"UTC":  "UTC",
					"GMT":  "UTC",
					"EST":  "America/New_York",
					"EDT":  "America/New_York",
					"CST":  "America/Chicago",
					"CDT":  "America/Chicago",
					"MST":  "America/Denver",
					"MDT":  "America/Denver",
					"PST":  "America/Los_Angeles",
					"PDT":  "America/Los_Angeles",
					"HST":  "Pacific/Honolulu",
					"AKST": "America/Anchorage",
					"AKDT": "America/Anchorage",
				}

				if timezone, exists := timezoneMap[abbr]; exists {
					return values.NewString(timezone), nil
				}

				return values.NewBool(false), nil
			},
		},
		{
			Name: "strftime",
			Parameters: []*registry.Parameter{
				{Name: "format", Type: "string"},
				{Name: "timestamp", Type: "int", HasDefault: true, DefaultValue: values.NewNull()},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				format := args[0].ToString()

				var t time.Time
				if len(args) > 1 && !args[1].IsNull() {
					timestamp := args[1].ToInt()
					t = time.Unix(timestamp, 0)
				} else {
					t = time.Now()
				}

				// Convert strftime format to Go format (simplified implementation)
				formatted := formatStrfTime(format, t)
				return values.NewString(formatted), nil
			},
		},
		{
			Name: "gmstrftime",
			Parameters: []*registry.Parameter{
				{Name: "format", Type: "string"},
				{Name: "timestamp", Type: "int", HasDefault: true, DefaultValue: values.NewNull()},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				format := args[0].ToString()

				var t time.Time
				if len(args) > 1 && !args[1].IsNull() {
					timestamp := args[1].ToInt()
					t = time.Unix(timestamp, 0).UTC()
				} else {
					t = time.Now().UTC()
				}

				// Convert strftime format to Go format (simplified implementation)
				formatted := formatStrfTime(format, t)
				return values.NewString(formatted), nil
			},
		},
		{
			Name: "strptime",
			Parameters: []*registry.Parameter{
				{Name: "date", Type: "string"},
				{Name: "format", Type: "string"},
			},
			ReturnType: "array",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				dateStr := args[0].ToString()
				format := args[1].ToString()

				// Parse the date string according to the strftime format
				parsedTime, err := parseStrptimeString(dateStr, format)
				if err != nil {
					return values.NewBool(false), nil
				}

				// Return array with parsed components
				result := values.NewArray()
				result.ArraySet(values.NewString("tm_sec"), values.NewInt(int64(parsedTime.Second())))
				result.ArraySet(values.NewString("tm_min"), values.NewInt(int64(parsedTime.Minute())))
				result.ArraySet(values.NewString("tm_hour"), values.NewInt(int64(parsedTime.Hour())))
				result.ArraySet(values.NewString("tm_mday"), values.NewInt(int64(parsedTime.Day())))
				result.ArraySet(values.NewString("tm_mon"), values.NewInt(int64(parsedTime.Month()-1))) // 0-based
				result.ArraySet(values.NewString("tm_year"), values.NewInt(int64(parsedTime.Year()-1900))) // Years since 1900
				result.ArraySet(values.NewString("tm_wday"), values.NewInt(int64(parsedTime.Weekday()))) // 0=Sunday
				result.ArraySet(values.NewString("tm_yday"), values.NewInt(int64(parsedTime.YearDay()-1))) // 0-based
				result.ArraySet(values.NewString("tm_isdst"), values.NewInt(-1)) // DST info unavailable

				return result, nil
			},
		},
	}
}

// formatStrfTime converts strftime format codes to formatted time string
func formatStrfTime(format string, t time.Time) string {
	result := ""

	for i := 0; i < len(format); i++ {
		if format[i] == '%' && i+1 < len(format) {
			i++ // Skip %
			switch format[i] {
			case 'Y': // 4-digit year
				result += fmt.Sprintf("%04d", t.Year())
			case 'y': // 2-digit year
				result += fmt.Sprintf("%02d", t.Year()%100)
			case 'm': // Month (01-12)
				result += fmt.Sprintf("%02d", int(t.Month()))
			case 'B': // Full month name
				result += t.Month().String()
			case 'b': // Short month name
				result += t.Month().String()[:3]
			case 'd': // Day (01-31)
				result += fmt.Sprintf("%02d", t.Day())
			case 'e': // Day (1-31) with space padding
				result += fmt.Sprintf("%2d", t.Day())
			case 'A': // Full weekday name
				result += t.Weekday().String()
			case 'a': // Short weekday name
				result += t.Weekday().String()[:3]
			case 'H': // Hour (00-23)
				result += fmt.Sprintf("%02d", t.Hour())
			case 'I': // Hour (01-12)
				hour := t.Hour()
				if hour == 0 {
					hour = 12
				} else if hour > 12 {
					hour -= 12
				}
				result += fmt.Sprintf("%02d", hour)
			case 'M': // Minute (00-59)
				result += fmt.Sprintf("%02d", t.Minute())
			case 'S': // Second (00-59)
				result += fmt.Sprintf("%02d", t.Second())
			case 'p': // AM/PM
				if t.Hour() < 12 {
					result += "AM"
				} else {
					result += "PM"
				}
			case 'w': // Weekday (0=Sunday)
				result += fmt.Sprintf("%d", int(t.Weekday()))
			case 'j': // Day of year (001-366)
				result += fmt.Sprintf("%03d", t.YearDay())
			case 'U': // Week number (00-53, Sunday as first day)
				_, week := t.ISOWeek()
				result += fmt.Sprintf("%02d", week)
			case 'W': // Week number (00-53, Monday as first day)
				_, week := t.ISOWeek()
				result += fmt.Sprintf("%02d", week)
			case 'c': // Complete date and time
				result += t.Format("Mon Jan 2 15:04:05 2006")
			case 'x': // Date representation
				result += t.Format("01/02/06")
			case 'X': // Time representation
				result += t.Format("15:04:05")
			case '%': // Literal %
				result += "%"
			default:
				// Unknown format, keep as-is
				result += "%" + string(format[i])
			}
		} else {
			result += string(format[i])
		}
	}

	return result
}

// parseTimeStringForParsing is a specialized version for date_parse
func parseTimeStringForParsing(timeStr string) (time.Time, error) {
	timeStr = strings.TrimSpace(timeStr)

	// Common formats for parsing
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02",
		"15:04:05",
		"01/02/2006",
		"01/02/2006 15:04:05",
		"Jan 02, 2006",
		"January 02, 2006",
		"Jan 02, 2006 15:04:05",
		"January 02, 2006 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time string: %s", timeStr)
}

// phpFormatToGoFormat converts PHP date format to Go time format
func phpFormatToGoFormat(phpFormat string) (string, error) {
	result := ""
	for i := 0; i < len(phpFormat); i++ {
		char := phpFormat[i]
		switch char {
		case 'd': // Day with leading zeros (01-31)
			result += "02"
		case 'j': // Day without leading zeros (1-31)
			result += "2"
		case 'm': // Month with leading zeros (01-12)
			result += "01"
		case 'n': // Month without leading zeros (1-12)
			result += "1"
		case 'Y': // 4-digit year
			result += "2006"
		case 'y': // 2-digit year
			result += "06"
		case 'H': // Hour in 24h format with leading zeros (00-23)
			result += "15"
		case 'G': // Hour in 24h format without leading zeros (0-23)
			result += "15" // Go doesn't have this format, use 15
		case 'i': // Minutes with leading zeros (00-59)
			result += "04"
		case 's': // Seconds with leading zeros (00-59)
			result += "05"
		default:
			result += string(char)
		}
	}
	return result, nil
}

// formatDateTime formats a time.Time according to PHP date format string
func formatDateTime(format string, t time.Time) (string, error) {
	result := ""

	for i := 0; i < len(format); i++ {
		char := format[i]

		switch char {
		case 'Y': // 4-digit year
			result += strconv.Itoa(t.Year())
		case 'y': // 2-digit year
			result += fmt.Sprintf("%02d", t.Year()%100)
		case 'm': // Month with leading zeros (01-12)
			result += fmt.Sprintf("%02d", int(t.Month()))
		case 'n': // Month without leading zeros (1-12)
			result += strconv.Itoa(int(t.Month()))
		case 'F': // Full month name
			result += t.Month().String()
		case 'M': // Short month name
			result += t.Month().String()[:3]
		case 'd': // Day with leading zeros (01-31)
			result += fmt.Sprintf("%02d", t.Day())
		case 'j': // Day without leading zeros (1-31)
			result += strconv.Itoa(t.Day())
		case 'l': // Full day name
			result += t.Weekday().String()
		case 'D': // Short day name
			result += t.Weekday().String()[:3]
		case 'H': // Hour in 24h format with leading zeros (00-23)
			result += fmt.Sprintf("%02d", t.Hour())
		case 'G': // Hour in 24h format without leading zeros (0-23)
			result += strconv.Itoa(t.Hour())
		case 'h': // Hour in 12h format with leading zeros (01-12)
			hour := t.Hour()
			if hour == 0 {
				hour = 12
			} else if hour > 12 {
				hour -= 12
			}
			result += fmt.Sprintf("%02d", hour)
		case 'g': // Hour in 12h format without leading zeros (1-12)
			hour := t.Hour()
			if hour == 0 {
				hour = 12
			} else if hour > 12 {
				hour -= 12
			}
			result += strconv.Itoa(hour)
		case 'i': // Minutes with leading zeros (00-59)
			result += fmt.Sprintf("%02d", t.Minute())
		case 's': // Seconds with leading zeros (00-59)
			result += fmt.Sprintf("%02d", t.Second())
		case 'A': // AM/PM uppercase
			if t.Hour() < 12 {
				result += "AM"
			} else {
				result += "PM"
			}
		case 'a': // am/pm lowercase
			if t.Hour() < 12 {
				result += "am"
			} else {
				result += "pm"
			}
		case 'U': // Unix timestamp
			result += strconv.FormatInt(t.Unix(), 10)
		case 'w': // Day of week (0=Sunday, 6=Saturday)
			result += strconv.Itoa(int(t.Weekday()))
		case 'z': // Day of year (0-365)
			result += strconv.Itoa(t.YearDay() - 1)
		case 'W': // Week number (ISO-8601)
			year, week := t.ISOWeek()
			_ = year
			result += fmt.Sprintf("%02d", week)
		case 't': // Number of days in month
			// Get last day of month
			lastDay := time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, t.Location()).Day()
			result += strconv.Itoa(lastDay)
		case 'L': // Leap year (1 if leap year, 0 otherwise)
			year := t.Year()
			if (year%4 == 0 && year%100 != 0) || year%400 == 0 {
				result += "1"
			} else {
				result += "0"
			}
		case 'T': // Timezone abbreviation
			result += t.Format("MST")
		case 'O': // Difference to GMT in hours (+0200)
			result += t.Format("-0700")
		case 'P': // Difference to GMT with colon (+02:00)
			result += t.Format("-07:00")
		case '\\': // Escape next character
			if i+1 < len(format) {
				i++
				result += string(format[i])
			} else {
				result += string(char)
			}
		default:
			result += string(char)
		}
	}

	return result, nil
}

// parseTimeString parses various time string formats
func parseTimeString(timeStr string, baseTime time.Time) (time.Time, error) {
	timeStr = strings.TrimSpace(timeStr)

	// Handle special keywords
	switch strings.ToLower(timeStr) {
	case "now":
		return time.Now(), nil
	case "today":
		return time.Date(baseTime.Year(), baseTime.Month(), baseTime.Day(), 0, 0, 0, 0, baseTime.Location()), nil
	case "tomorrow":
		return baseTime.AddDate(0, 0, 1), nil
	case "yesterday":
		return baseTime.AddDate(0, 0, -1), nil
	}

	// Handle relative time strings
	if strings.HasPrefix(timeStr, "+") || strings.HasPrefix(timeStr, "-") {
		return parseRelativeTime(timeStr, baseTime)
	}

	// Handle "next" keywords
	if strings.HasPrefix(strings.ToLower(timeStr), "next ") {
		return parseNext(timeStr, baseTime)
	}

	// Handle ISO format dates
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02",
		"15:04:05",
		"01/02/2006",
		"01/02/2006 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time string: %s", timeStr)
}

// parseRelativeTime handles "+1 day", "-2 hours", etc.
func parseRelativeTime(timeStr string, baseTime time.Time) (time.Time, error) {
	// Simple regex for relative time
	re := regexp.MustCompile(`([+-]?\d+)\s+(second|minute|hour|day|week|month|year)s?`)
	matches := re.FindStringSubmatch(timeStr)

	if len(matches) != 3 {
		return time.Time{}, fmt.Errorf("invalid relative time format")
	}

	amount, err := strconv.Atoi(matches[1])
	if err != nil {
		return time.Time{}, err
	}

	unit := matches[2]

	switch unit {
	case "second":
		return baseTime.Add(time.Duration(amount) * time.Second), nil
	case "minute":
		return baseTime.Add(time.Duration(amount) * time.Minute), nil
	case "hour":
		return baseTime.Add(time.Duration(amount) * time.Hour), nil
	case "day":
		return baseTime.AddDate(0, 0, amount), nil
	case "week":
		return baseTime.AddDate(0, 0, amount*7), nil
	case "month":
		return baseTime.AddDate(0, amount, 0), nil
	case "year":
		return baseTime.AddDate(amount, 0, 0), nil
	}

	return time.Time{}, fmt.Errorf("unknown time unit: %s", unit)
}

// parseNext handles "next Monday", "next week", etc.
func parseNext(timeStr string, baseTime time.Time) (time.Time, error) {
	parts := strings.Fields(strings.ToLower(timeStr))
	if len(parts) < 2 || parts[0] != "next" {
		return time.Time{}, fmt.Errorf("invalid next format")
	}

	target := parts[1]

	weekdays := map[string]time.Weekday{
		"sunday":    time.Sunday,
		"monday":    time.Monday,
		"tuesday":   time.Tuesday,
		"wednesday": time.Wednesday,
		"thursday":  time.Thursday,
		"friday":    time.Friday,
		"saturday":  time.Saturday,
	}

	if weekday, exists := weekdays[target]; exists {
		// Find next occurrence of this weekday
		days := int(weekday - baseTime.Weekday())
		if days <= 0 {
			days += 7
		}
		return baseTime.AddDate(0, 0, days), nil
	}

	switch target {
	case "week":
		return baseTime.AddDate(0, 0, 7), nil
	case "month":
		return baseTime.AddDate(0, 1, 0), nil
	case "year":
		return baseTime.AddDate(1, 0, 0), nil
	}

	return time.Time{}, fmt.Errorf("unknown next target: %s", target)
}

// parseStrptimeString parses a date string according to strftime format
func parseStrptimeString(dateStr, format string) (time.Time, error) {
	// Convert strftime format to Go time format
	goFormat := convertStrftimeToGoFormat(format)

	// Try parsing with the converted format
	parsedTime, err := time.Parse(goFormat, dateStr)
	if err != nil {
		return time.Time{}, err
	}

	return parsedTime, nil
}

// convertStrftimeToGoFormat converts strftime format to Go time format
func convertStrftimeToGoFormat(strftimeFormat string) string {
	// Map of strftime codes to Go time format
	replacements := map[string]string{
		"%Y": "2006",     // 4-digit year
		"%y": "06",       // 2-digit year
		"%m": "01",       // Month (01-12)
		"%B": "January",  // Full month name
		"%b": "Jan",      // Short month name
		"%d": "02",       // Day of month (01-31)
		"%e": "_2",       // Day of month, space padded
		"%H": "15",       // Hour (00-23)
		"%I": "03",       // Hour (01-12)
		"%M": "04",       // Minute (00-59)
		"%S": "05",       // Second (00-60)
		"%p": "PM",       // AM/PM
		"%A": "Monday",   // Full weekday name
		"%a": "Mon",      // Short weekday name
		"%w": "0",        // Weekday (0=Sunday)
		"%z": "-0700",    // Timezone offset
		"%Z": "MST",      // Timezone abbreviation
	}

	result := strftimeFormat
	for strftime, goFormat := range replacements {
		result = strings.ReplaceAll(result, strftime, goFormat)
	}

	return result
}