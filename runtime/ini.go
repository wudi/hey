package runtime

import (
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// IniSetting represents a single configuration setting
type IniSetting struct {
	Name         string // Configuration name
	GlobalValue  string // Current global value
	LocalValue   string // Current local value (usually same as global for our implementation)
	OriginalValue string // Original/default value for ini_restore
	Access       int64  // Access level (bitmask: 1=user, 2=perdir, 4=system, 7=all)
}

// IniStorage manages all ini configuration settings
type IniStorage struct {
	mu       sync.RWMutex
	settings map[string]*IniSetting
}

var (
	globalIniStorage     *IniStorage
	iniStorageOnce       sync.Once
)

// getIniStorage returns the global ini storage instance
func getIniStorage() *IniStorage {
	iniStorageOnce.Do(func() {
		globalIniStorage = &IniStorage{
			settings: make(map[string]*IniSetting),
		}
		// Initialize with some common PHP settings
		initializeDefaultSettings()
	})
	return globalIniStorage
}

// initializeDefaultSettings populates default PHP ini settings
func initializeDefaultSettings() {
	defaultSettings := map[string]*IniSetting{
		"display_errors": {
			Name: "display_errors",
			GlobalValue: "",
			LocalValue: "",
			OriginalValue: "",
			Access: 7, // PHP_INI_ALL
		},
		"memory_limit": {
			Name: "memory_limit",
			GlobalValue: "-1",
			LocalValue: "-1",
			OriginalValue: "-1",
			Access: 7, // PHP_INI_ALL
		},
		"max_execution_time": {
			Name: "max_execution_time",
			GlobalValue: "30",
			LocalValue: "30",
			OriginalValue: "30",
			Access: 7, // PHP_INI_ALL
		},
		"error_reporting": {
			Name: "error_reporting",
			GlobalValue: "22527",
			LocalValue: "22527",
			OriginalValue: "22527",
			Access: 7, // PHP_INI_ALL
		},
		"default_charset": {
			Name: "default_charset",
			GlobalValue: "UTF-8",
			LocalValue: "UTF-8",
			OriginalValue: "UTF-8",
			Access: 7, // PHP_INI_ALL
		},
		"allow_url_fopen": {
			Name: "allow_url_fopen",
			GlobalValue: "1",
			LocalValue: "1",
			OriginalValue: "1",
			Access: 4, // PHP_INI_SYSTEM
		},
		"allow_url_include": {
			Name: "allow_url_include",
			GlobalValue: "",
			LocalValue: "",
			OriginalValue: "",
			Access: 4, // PHP_INI_SYSTEM
		},
		"arg_separator.input": {
			Name: "arg_separator.input",
			GlobalValue: "&",
			LocalValue: "&",
			OriginalValue: "&",
			Access: 6, // PHP_INI_PERDIR
		},
		"arg_separator.output": {
			Name: "arg_separator.output",
			GlobalValue: "&",
			LocalValue: "&",
			OriginalValue: "&",
			Access: 7, // PHP_INI_ALL
		},
		"assert.active": {
			Name: "assert.active",
			GlobalValue: "1",
			LocalValue: "1",
			OriginalValue: "1",
			Access: 7, // PHP_INI_ALL
		},
	}

	for name, setting := range defaultSettings {
		globalIniStorage.settings[name] = setting
	}
}

// GetIniFunctions returns all ini-related PHP functions
func GetIniFunctions() []*registry.Function {
	return []*registry.Function{
		{
			Name:       "ini_get",
			Parameters: []*registry.Parameter{{Name: "option", Type: "string"}},
			ReturnType: "string|false",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}

				optionName := args[0].ToString()
				storage := getIniStorage()

				storage.mu.RLock()
				setting, exists := storage.settings[optionName]
				storage.mu.RUnlock()

				if !exists {
					return values.NewBool(false), nil
				}

				return values.NewString(setting.LocalValue), nil
			},
		},
		{
			Name: "ini_set",
			Parameters: []*registry.Parameter{
				{Name: "option", Type: "string"},
				{Name: "value", Type: "mixed"},
			},
			ReturnType: "string|false",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || args[1] == nil {
					return values.NewBool(false), nil
				}

				optionName := args[0].ToString()
				newValue := args[1].ToString()
				storage := getIniStorage()

				storage.mu.Lock()
				defer storage.mu.Unlock()

				setting, exists := storage.settings[optionName]
				if !exists {
					return values.NewBool(false), nil
				}

				// Store old value to return
				oldValue := setting.LocalValue

				// Update the setting
				setting.GlobalValue = newValue
				setting.LocalValue = newValue

				return values.NewString(oldValue), nil
			},
		},
		{
			Name:       "ini_restore",
			Parameters: []*registry.Parameter{{Name: "option", Type: "string"}},
			ReturnType: "void",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewNull(), nil
				}

				optionName := args[0].ToString()
				storage := getIniStorage()

				storage.mu.Lock()
				defer storage.mu.Unlock()

				setting, exists := storage.settings[optionName]
				if exists {
					setting.GlobalValue = setting.OriginalValue
					setting.LocalValue = setting.OriginalValue
				}

				return values.NewNull(), nil
			},
		},
		{
			Name: "ini_get_all",
			Parameters: []*registry.Parameter{
				{Name: "extension", Type: "string", HasDefault: true, DefaultValue: values.NewNull()},
				{Name: "details", Type: "bool", HasDefault: true, DefaultValue: values.NewBool(true)},
			},
			ReturnType: "array|false",
			MinArgs:    0,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				storage := getIniStorage()

				var extensionFilter string
				if len(args) > 0 && args[0] != nil && !args[0].IsNull() {
					extensionFilter = args[0].ToString()
					// For simplicity, we don't implement extension filtering
					// Return false for any extension filter (matches PHP behavior for unknown extensions)
					if extensionFilter != "" {
						return values.NewBool(false), nil
					}
				}

				details := true
				if len(args) > 1 && args[1] != nil {
					details = args[1].ToBool()
				}

				result := values.NewArray()

				storage.mu.RLock()
				defer storage.mu.RUnlock()

				for name, setting := range storage.settings {
					if details {
						// Return detailed array with global_value, local_value, access
						settingArray := values.NewArray()
						settingArray.ArraySet(values.NewString("global_value"), values.NewString(setting.GlobalValue))
						settingArray.ArraySet(values.NewString("local_value"), values.NewString(setting.LocalValue))
						settingArray.ArraySet(values.NewString("access"), values.NewInt(setting.Access))
						result.ArraySet(values.NewString(name), settingArray)
					} else {
						// Return just the current value
						result.ArraySet(values.NewString(name), values.NewString(setting.LocalValue))
					}
				}

				return result, nil
			},
		},
		{
			Name:       "ini_parse_quantity",
			Parameters: []*registry.Parameter{{Name: "shorthand", Type: "string"}},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewInt(0), nil
				}

				shorthand := strings.TrimSpace(args[0].ToString())
				if shorthand == "" {
					return values.NewInt(0), nil
				}

				// Parse quantity with K/M/G suffixes
				return parseQuantity(shorthand)
			},
		},
	}
}

// parseQuantity parses a size string like "128M" into bytes
func parseQuantity(shorthand string) (*values.Value, error) {
	if shorthand == "" {
		return values.NewInt(0), nil
	}

	// Remove whitespace
	shorthand = strings.TrimSpace(shorthand)

	// Regular expression to match number followed by optional suffix
	re := regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*([KMGkmg])?$`)
	matches := re.FindStringSubmatch(shorthand)

	if len(matches) == 0 {
		// If no valid number found, return 0 (matches PHP behavior)
		return values.NewInt(0), nil
	}

	// Parse the numeric part
	numStr := matches[1]
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return values.NewInt(0), nil
	}

	// Apply suffix multiplier
	suffix := strings.ToUpper(matches[2])
	switch suffix {
	case "K":
		num *= 1024
	case "M":
		num *= 1024 * 1024
	case "G":
		num *= 1024 * 1024 * 1024
	}

	return values.NewInt(int64(num)), nil
}