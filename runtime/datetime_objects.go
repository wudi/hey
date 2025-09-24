package runtime

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetDateTimeObjectFunctions returns DateTime object-like functions implemented as procedural functions
// This provides the essential DateTime functionality without complex object system integration
func GetDateTimeObjectFunctions() []*registry.Function {
	return []*registry.Function{
		// DateTime_create - factory function that mimics "new DateTime()"
		{
			Name: "DateTime_create",
			Parameters: []*registry.Parameter{
				{Name: "datetime", Type: "string", HasDefault: true, DefaultValue: values.NewString("now")},
				{Name: "timezone", Type: "string", HasDefault: true, DefaultValue: values.NewString("UTC")},
			},
			ReturnType: "array",
			MinArgs:    0,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				dateTimeStr := "now"
				if len(args) > 0 && !args[0].IsNull() {
					dateTimeStr = args[0].ToString()
				}

				timezoneStr := "UTC"
				if len(args) > 1 && !args[1].IsNull() {
					timezoneStr = args[1].ToString()
				}

				// Parse the datetime
				var t time.Time
				var err error

				if dateTimeStr == "now" {
					t = time.Now()
				} else {
					t, err = parseTimeStringForParsing(dateTimeStr)
					if err != nil {
						return nil, fmt.Errorf("DateTime_create(): Failed to parse time string '%s'", dateTimeStr)
					}
				}

				// Load timezone
				loc, err := time.LoadLocation(timezoneStr)
				if err != nil {
					loc = time.UTC
				}
				t = t.In(loc)

				// Return array representing DateTime object
				result := values.NewArray()
				result.ArraySet(values.NewString("_timestamp"), values.NewInt(t.Unix()))
				result.ArraySet(values.NewString("_timezone"), values.NewString(timezoneStr))
				result.ArraySet(values.NewString("_type"), values.NewString("DateTime"))

				return result, nil
			},
		},
		// DateTime_format - format a DateTime object
		{
			Name: "DateTime_format",
			Parameters: []*registry.Parameter{
				{Name: "datetime_obj", Type: "array"},
				{Name: "format", Type: "string"},
			},
			ReturnType: "string",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				dtObj := args[0]
				format := args[1].ToString()

				if !dtObj.IsArray() {
					return nil, fmt.Errorf("DateTime_format(): First argument must be a DateTime object")
				}

				// Extract timestamp
				timestampVal := dtObj.ArrayGet(values.NewString("_timestamp"))
				if timestampVal == nil || timestampVal.IsNull() {
					return nil, fmt.Errorf("DateTime_format(): Invalid DateTime object")
				}

				timestamp := timestampVal.ToInt()
				timezoneVal := dtObj.ArrayGet(values.NewString("_timezone"))
				timezoneStr := "UTC"
				if timezoneVal != nil && !timezoneVal.IsNull() {
					timezoneStr = timezoneVal.ToString()
				}

				// Load timezone and create time
				loc, err := time.LoadLocation(timezoneStr)
				if err != nil {
					loc = time.UTC
				}

				t := time.Unix(timestamp, 0).In(loc)
				formatted, err := formatDateTime(format, t)
				if err != nil {
					return nil, err
				}

				return values.NewString(formatted), nil
			},
		},
		// DateTime_getTimestamp - get Unix timestamp from DateTime object
		{
			Name: "DateTime_getTimestamp",
			Parameters: []*registry.Parameter{
				{Name: "datetime_obj", Type: "array"},
			},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				dtObj := args[0]

				if !dtObj.IsArray() {
					return nil, fmt.Errorf("DateTime_getTimestamp(): Argument must be a DateTime object")
				}

				timestampVal := dtObj.ArrayGet(values.NewString("_timestamp"))
				if timestampVal == nil || timestampVal.IsNull() {
					return nil, fmt.Errorf("DateTime_getTimestamp(): Invalid DateTime object")
				}

				return timestampVal, nil
			},
		},
		// DateTime_setTimestamp - set Unix timestamp on DateTime object
		{
			Name: "DateTime_setTimestamp",
			Parameters: []*registry.Parameter{
				{Name: "datetime_obj", Type: "array"},
				{Name: "unixtimestamp", Type: "int"},
			},
			ReturnType: "array",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				dtObj := args[0]
				timestamp := args[1].ToInt()

				if !dtObj.IsArray() {
					return nil, fmt.Errorf("DateTime_setTimestamp(): First argument must be a DateTime object")
				}

				// Update the timestamp
				dtObj.ArraySet(values.NewString("_timestamp"), values.NewInt(timestamp))

				return dtObj, nil
			},
		},
		// DateTime_modify - modify DateTime object with relative time string
		{
			Name: "DateTime_modify",
			Parameters: []*registry.Parameter{
				{Name: "datetime_obj", Type: "array"},
				{Name: "modifier", Type: "string"},
			},
			ReturnType: "array|false",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				dtObj := args[0]
				modifier := args[1].ToString()

				if !dtObj.IsArray() {
					return nil, fmt.Errorf("DateTime_modify(): First argument must be a DateTime object")
				}

				// Get current timestamp
				timestampVal := dtObj.ArrayGet(values.NewString("_timestamp"))
				if timestampVal == nil || timestampVal.IsNull() {
					return values.NewBool(false), nil
				}

				currentTime := time.Unix(timestampVal.ToInt(), 0)

				// Parse and apply the modifier
				newTime, err := parseTimeString(modifier, currentTime)
				if err != nil {
					return values.NewBool(false), nil
				}

				// Update timestamp
				dtObj.ArraySet(values.NewString("_timestamp"), values.NewInt(newTime.Unix()))

				return dtObj, nil
			},
		},
		// DateTime_createFromFormat - create DateTime from format string
		{
			Name: "DateTime_createFromFormat",
			Parameters: []*registry.Parameter{
				{Name: "format", Type: "string"},
				{Name: "datetime", Type: "string"},
				{Name: "timezone", Type: "string", HasDefault: true, DefaultValue: values.NewString("UTC")},
			},
			ReturnType: "array|false",
			MinArgs:    2,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				format := args[0].ToString()
				dateTimeStr := args[1].ToString()

				timezoneStr := "UTC"
				if len(args) > 2 && !args[2].IsNull() {
					timezoneStr = args[2].ToString()
				}

				// Convert PHP format to Go format
				goFormat, err := phpFormatToGoFormat(format)
				if err != nil {
					return values.NewBool(false), nil
				}

				// Parse the datetime
				parsedTime, err := time.Parse(goFormat, dateTimeStr)
				if err != nil {
					return values.NewBool(false), nil
				}

				// Load timezone
				loc, err := time.LoadLocation(timezoneStr)
				if err != nil {
					loc = time.UTC
				}
				parsedTime = parsedTime.In(loc)

				// Create DateTime object
				result := values.NewArray()
				result.ArraySet(values.NewString("_timestamp"), values.NewInt(parsedTime.Unix()))
				result.ArraySet(values.NewString("_timezone"), values.NewString(timezoneStr))
				result.ArraySet(values.NewString("_type"), values.NewString("DateTime"))

				return result, nil
			},
		},
		// DateInterval_create - create DateInterval from ISO string
		{
			Name: "DateInterval_create",
			Parameters: []*registry.Parameter{
				{Name: "interval_spec", Type: "string"},
			},
			ReturnType: "array",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				intervalSpec := args[0].ToString()

				// Parse ISO 8601 interval specification
				interval, err := parseISOInterval(intervalSpec)
				if err != nil {
					return nil, fmt.Errorf("DateInterval_create(): Unknown or bad format (%s)", intervalSpec)
				}

				// Create DateInterval object
				result := values.NewArray()
				result.ArraySet(values.NewString("_type"), values.NewString("DateInterval"))
				result.ArraySet(values.NewString("y"), values.NewInt(int64(interval.Years)))
				result.ArraySet(values.NewString("m"), values.NewInt(int64(interval.Months)))
				result.ArraySet(values.NewString("d"), values.NewInt(int64(interval.Days)))
				result.ArraySet(values.NewString("h"), values.NewInt(int64(interval.Hours)))
				result.ArraySet(values.NewString("i"), values.NewInt(int64(interval.Minutes)))
				result.ArraySet(values.NewString("s"), values.NewInt(int64(interval.Seconds)))
				result.ArraySet(values.NewString("invert"), values.NewBool(interval.Invert))

				return result, nil
			},
		},
		// DateInterval_format - format DateInterval
		{
			Name: "DateInterval_format",
			Parameters: []*registry.Parameter{
				{Name: "interval_obj", Type: "array"},
				{Name: "format", Type: "string"},
			},
			ReturnType: "string",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				intervalObj := args[0]
				format := args[1].ToString()

				if !intervalObj.IsArray() {
					return nil, fmt.Errorf("DateInterval_format(): First argument must be a DateInterval object")
				}

				// Extract interval data
				interval := &DateIntervalData{
					Years:   int(intervalObj.ArrayGet(values.NewString("y")).ToInt()),
					Months:  int(intervalObj.ArrayGet(values.NewString("m")).ToInt()),
					Days:    int(intervalObj.ArrayGet(values.NewString("d")).ToInt()),
					Hours:   int(intervalObj.ArrayGet(values.NewString("h")).ToInt()),
					Minutes: int(intervalObj.ArrayGet(values.NewString("i")).ToInt()),
					Seconds: int(intervalObj.ArrayGet(values.NewString("s")).ToInt()),
					Invert:  intervalObj.ArrayGet(values.NewString("invert")).ToBool(),
				}

				formatted := formatInterval(format, interval)
				return values.NewString(formatted), nil
			},
		},
		// DateTime_add - add DateInterval to DateTime
		{
			Name: "DateTime_add",
			Parameters: []*registry.Parameter{
				{Name: "datetime_obj", Type: "array"},
				{Name: "interval_obj", Type: "array"},
			},
			ReturnType: "array",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				dtObj := args[0]
				intervalObj := args[1]

				if !dtObj.IsArray() {
					return nil, fmt.Errorf("DateTime_add(): First argument must be a DateTime object")
				}
				if !intervalObj.IsArray() {
					return nil, fmt.Errorf("DateTime_add(): Second argument must be a DateInterval object")
				}

				// Get current timestamp
				timestampVal := dtObj.ArrayGet(values.NewString("_timestamp"))
				if timestampVal == nil || timestampVal.IsNull() {
					return nil, fmt.Errorf("DateTime_add(): Invalid DateTime object")
				}

				currentTime := time.Unix(timestampVal.ToInt(), 0)

				// Extract interval values
				years := int(intervalObj.ArrayGet(values.NewString("y")).ToInt())
				months := int(intervalObj.ArrayGet(values.NewString("m")).ToInt())
				days := int(intervalObj.ArrayGet(values.NewString("d")).ToInt())
				hours := int(intervalObj.ArrayGet(values.NewString("h")).ToInt())
				minutes := int(intervalObj.ArrayGet(values.NewString("i")).ToInt())
				seconds := int(intervalObj.ArrayGet(values.NewString("s")).ToInt())

				// Apply interval
				newTime := currentTime.AddDate(years, months, days)
				newTime = newTime.Add(time.Duration(hours)*time.Hour +
					time.Duration(minutes)*time.Minute +
					time.Duration(seconds)*time.Second)

				// Update timestamp
				dtObj.ArraySet(values.NewString("_timestamp"), values.NewInt(newTime.Unix()))

				return dtObj, nil
			},
		},
		// DateTime_diff - calculate difference between two DateTime objects
		{
			Name: "DateTime_diff",
			Parameters: []*registry.Parameter{
				{Name: "datetime1", Type: "array"},
				{Name: "datetime2", Type: "array"},
			},
			ReturnType: "array",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				dt1 := args[0]
				dt2 := args[1]

				if !dt1.IsArray() || !dt2.IsArray() {
					return nil, fmt.Errorf("DateTime_diff(): Both arguments must be DateTime objects")
				}

				// Get timestamps
				ts1 := dt1.ArrayGet(values.NewString("_timestamp")).ToInt()
				ts2 := dt2.ArrayGet(values.NewString("_timestamp")).ToInt()

				// Calculate difference in seconds
				diff := ts2 - ts1
				isNegative := diff < 0
				if isNegative {
					diff = -diff
				}

				// Convert to components
				days := diff / 86400
				remaining := diff % 86400
				hours := remaining / 3600
				remaining = remaining % 3600
				minutes := remaining / 60
				seconds := remaining % 60

				// Create DateInterval result
				result := values.NewArray()
				result.ArraySet(values.NewString("_type"), values.NewString("DateInterval"))
				result.ArraySet(values.NewString("y"), values.NewInt(0)) // Simplified
				result.ArraySet(values.NewString("m"), values.NewInt(0)) // Simplified
				result.ArraySet(values.NewString("d"), values.NewInt(days))
				result.ArraySet(values.NewString("h"), values.NewInt(hours))
				result.ArraySet(values.NewString("i"), values.NewInt(minutes))
				result.ArraySet(values.NewString("s"), values.NewInt(seconds))
				result.ArraySet(values.NewString("days"), values.NewInt(days)) // Total days
				result.ArraySet(values.NewString("invert"), values.NewBool(isNegative))

				return result, nil
			},
		},
	}
}

// DateIntervalData represents interval data
type DateIntervalData struct {
	Years   int
	Months  int
	Days    int
	Hours   int
	Minutes int
	Seconds int
	Invert  bool // Whether the interval is negative
}

// parseISOInterval parses ISO 8601 duration format (P1Y2M3DT4H5M6S)
func parseISOInterval(spec string) (*DateIntervalData, error) {
	if len(spec) == 0 || spec[0] != 'P' {
		return nil, fmt.Errorf("invalid interval format")
	}

	spec = spec[1:] // Remove 'P'
	interval := &DateIntervalData{}

	// Split at 'T' if present
	timePart := ""
	if tIndex := strings.Index(spec, "T"); tIndex >= 0 {
		timePart = spec[tIndex+1:]
		spec = spec[:tIndex]
	}

	// Parse date part
	if spec != "" {
		if err := parseIntervalDatePart(spec, interval); err != nil {
			return nil, err
		}
	}

	// Parse time part
	if timePart != "" {
		if err := parseIntervalTimePart(timePart, interval); err != nil {
			return nil, err
		}
	}

	return interval, nil
}

func parseIntervalDatePart(part string, interval *DateIntervalData) error {
	i := 0
	for i < len(part) {
		// Extract number
		numStart := i
		for i < len(part) && (part[i] >= '0' && part[i] <= '9') {
			i++
		}

		if i == numStart {
			return fmt.Errorf("invalid interval format")
		}

		numStr := part[numStart:i]
		num, err := strconv.Atoi(numStr)
		if err != nil {
			return err
		}

		// Extract unit
		if i >= len(part) {
			return fmt.Errorf("invalid interval format")
		}

		unit := part[i]
		i++

		switch unit {
		case 'Y':
			interval.Years = num
		case 'M':
			interval.Months = num
		case 'D':
			interval.Days = num
		default:
			return fmt.Errorf("invalid date unit: %c", unit)
		}
	}

	return nil
}

func parseIntervalTimePart(part string, interval *DateIntervalData) error {
	i := 0
	for i < len(part) {
		// Extract number
		numStart := i
		for i < len(part) && (part[i] >= '0' && part[i] <= '9') {
			i++
		}

		if i == numStart {
			return fmt.Errorf("invalid interval format")
		}

		numStr := part[numStart:i]
		num, err := strconv.Atoi(numStr)
		if err != nil {
			return err
		}

		// Extract unit
		if i >= len(part) {
			return fmt.Errorf("invalid interval format")
		}

		unit := part[i]
		i++

		switch unit {
		case 'H':
			interval.Hours = num
		case 'M':
			interval.Minutes = num
		case 'S':
			interval.Seconds = num
		default:
			return fmt.Errorf("invalid time unit: %c", unit)
		}
	}

	return nil
}

// formatInterval formats a DateInterval according to format string
func formatInterval(format string, interval *DateIntervalData) string {
	result := ""

	for i := 0; i < len(format); i++ {
		if format[i] == '%' && i+1 < len(format) {
			i++ // Skip %
			switch format[i] {
			case 'y', 'Y':
				result += strconv.Itoa(interval.Years)
			case 'm':
				result += strconv.Itoa(interval.Months)
			case 'd':
				result += strconv.Itoa(interval.Days)
			case 'h', 'H':
				result += strconv.Itoa(interval.Hours)
			case 'i', 'I':
				result += strconv.Itoa(interval.Minutes)
			case 's', 'S':
				result += strconv.Itoa(interval.Seconds)
			case 'a':
				// Total days - simplified calculation
				totalDays := interval.Days + (interval.Years * 365) + (interval.Months * 30)
				result += strconv.Itoa(totalDays)
			case 'R':
				if interval.Invert {
					result += "-"
				} else {
					result += "+"
				}
			default:
				result += "%" + string(format[i])
			}
		} else {
			result += string(format[i])
		}
	}

	return result
}