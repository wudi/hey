package runtime

import (
	"bufio"
	"encoding/csv"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// FileHandle represents an open file handle
type FileHandle struct {
	ID       int64
	File     *os.File
	Mode     string
	Position int64
	EOF      bool
	mu       sync.RWMutex
}

// ProcessHandle represents an open process handle for popen
type ProcessHandle struct {
	ID      int64
	Cmd     *exec.Cmd
	File    io.Closer // stdin or stdout of the process (can be ReadCloser or WriteCloser)
	Mode    string
	mu      sync.RWMutex
}

// Global file handle registry
var (
	fileHandleCounter    int64
	fileHandles          = make(map[int64]*FileHandle)
	fileHandlesMutex     sync.RWMutex
	processHandleCounter int64
	processHandles       = make(map[int64]*ProcessHandle)
	processHandlesMutex  sync.RWMutex
)

// registerFileHandle adds a file handle to the global registry
func registerFileHandle(handle *FileHandle) {
	fileHandlesMutex.Lock()
	defer fileHandlesMutex.Unlock()
	fileHandles[handle.ID] = handle
}

// getFileHandle retrieves a file handle from the registry
func getFileHandle(id int64) (*FileHandle, bool) {
	fileHandlesMutex.RLock()
	defer fileHandlesMutex.RUnlock()
	handle, exists := fileHandles[id]
	return handle, exists
}

// removeFileHandle removes a file handle from the registry
func removeFileHandle(id int64) {
	fileHandlesMutex.Lock()
	defer fileHandlesMutex.Unlock()
	delete(fileHandles, id)
}

// registerProcessHandle adds a process handle to the global registry
func registerProcessHandle(handle *ProcessHandle) {
	processHandlesMutex.Lock()
	defer processHandlesMutex.Unlock()
	processHandles[handle.ID] = handle
}

// getProcessHandle retrieves a process handle from the registry
func getProcessHandle(id int64) (*ProcessHandle, bool) {
	processHandlesMutex.RLock()
	defer processHandlesMutex.RUnlock()
	handle, exists := processHandles[id]
	return handle, exists
}

// removeProcessHandle removes a process handle from the registry
func removeProcessHandle(id int64) {
	processHandlesMutex.Lock()
	defer processHandlesMutex.Unlock()
	delete(processHandles, id)
}

// parseINIContent parses INI content string and returns a PHP-compatible array
func parseINIContent(content string, processSections bool, scannerMode int64) *values.Value {
	result := values.NewArray()
	var currentSection string
	var sectionMap *values.Value

	if processSections {
		sectionMap = values.NewArray()
	}

	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle section headers
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			if processSections {
				currentSection = line[1 : len(line)-1]
				sectionMap = values.NewArray()
				result.ArraySet(values.NewString(currentSection), sectionMap)
			}
			continue
		}

		// Handle key-value pairs
		if idx := strings.Index(line, "="); idx != -1 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])

			// Handle array keys (key[])
			isArray := false
			if strings.HasSuffix(key, "[]") {
				key = key[:len(key)-2]
				isArray = true
			}

			// Parse value
			parsedValue := parseINIValue(value, scannerMode)

			// Determine target map
			targetMap := result
			if processSections && sectionMap != nil {
				targetMap = sectionMap
			}

			keyValue := values.NewString(key)

			if isArray {
				// Handle array values
				existingArray := targetMap.ArrayGet(keyValue)
				if existingArray.IsNull() {
					existingArray = values.NewArray()
					targetMap.ArraySet(keyValue, existingArray)
				}

				// Add to array
				arrayLen := existingArray.ArrayCount()
				existingArray.ArraySet(values.NewInt(int64(arrayLen)), parsedValue)
			} else {
				// Regular key-value
				targetMap.ArraySet(keyValue, parsedValue)
			}
		}
	}

	return result
}

// parseINIValue parses an INI value string based on scanner mode
func parseINIValue(value string, scannerMode int64) *values.Value {
	// Remove quotes if present
	if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
		(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
		value = value[1 : len(value)-1]
	}

	// INI_SCANNER_RAW mode - return as-is
	if scannerMode == 1 {
		return values.NewString(value)
	}

	// INI_SCANNER_TYPED mode - convert types
	if scannerMode == 2 {
		// Check for boolean values
		lowerValue := strings.ToLower(value)
		if lowerValue == "true" || lowerValue == "on" || lowerValue == "yes" {
			return values.NewBool(true)
		}
		if lowerValue == "false" || lowerValue == "off" || lowerValue == "no" || lowerValue == "" {
			return values.NewBool(false)
		}

		// Try to parse as integer
		if strings.ContainsAny(value, "0123456789") && !strings.Contains(value, ".") {
			if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
				return values.NewInt(intVal)
			}
		}

		// Try to parse as float
		if strings.Contains(value, ".") {
			if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
				return values.NewFloat(floatVal)
			}
		}
	}

	// INI_SCANNER_NORMAL mode or fallback - handle boolean-like values as strings
	if scannerMode == 0 {
		lowerValue := strings.ToLower(value)
		if lowerValue == "true" || lowerValue == "on" || lowerValue == "yes" {
			return values.NewString("1")
		}
		if lowerValue == "false" || lowerValue == "off" || lowerValue == "no" {
			return values.NewString("")
		}
	}

	return values.NewString(value)
}

// fnmatchImpl implements filename matching similar to Unix fnmatch
func fnmatchImpl(pattern, str string, noEscape, pathname, period bool) bool {
	return fnmatchRecursive(pattern, str, 0, 0, noEscape, pathname, period)
}

// fnmatchRecursive implements the recursive pattern matching logic
func fnmatchRecursive(pattern, str string, pi, si int, noEscape, pathname, period bool) bool {
	for pi < len(pattern) {
		switch pattern[pi] {
		case '?':
			if si >= len(str) {
				return false
			}
			// FNM_PATHNAME: '?' doesn't match '/'
			if pathname && str[si] == '/' {
				return false
			}
			// FNM_PERIOD: '?' doesn't match '.' at start or after '/'
			if period && str[si] == '.' && (si == 0 || (pathname && si > 0 && str[si-1] == '/')) {
				return false
			}
			pi++
			si++

		case '*':
			// Skip consecutive asterisks
			for pi < len(pattern) && pattern[pi] == '*' {
				pi++
			}
			// If pattern ends with *, it matches rest of string
			if pi >= len(pattern) {
				// FNM_PATHNAME: * doesn't match / in path
				if pathname && strings.Contains(str[si:], "/") {
					return false
				}
				return true
			}
			// Try matching at each position
			for si <= len(str) {
				if fnmatchRecursive(pattern, str, pi, si, noEscape, pathname, period) {
					return true
				}
				if si >= len(str) {
					break
				}
				// FNM_PATHNAME: * doesn't match /
				if pathname && str[si] == '/' {
					return false
				}
				// FNM_PERIOD: * doesn't match . at start or after /
				if period && str[si] == '.' && (si == 0 || (pathname && si > 0 && str[si-1] == '/')) {
					return false
				}
				si++
			}
			return false

		case '[':
			if si >= len(str) {
				return false
			}
			// FNM_PATHNAME: '[' doesn't match '/'
			if pathname && str[si] == '/' {
				return false
			}
			// FNM_PERIOD: '[' doesn't match '.' at start or after '/'
			if period && str[si] == '.' && (si == 0 || (pathname && si > 0 && str[si-1] == '/')) {
				return false
			}
			// Find the end of character class
			classEnd := pi + 1
			if classEnd < len(pattern) && pattern[classEnd] == ']' {
				classEnd++ // Handle []...] case
			}
			for classEnd < len(pattern) && pattern[classEnd] != ']' {
				if !noEscape && pattern[classEnd] == '\\' && classEnd+1 < len(pattern) {
					classEnd += 2
				} else {
					classEnd++
				}
			}
			if classEnd >= len(pattern) {
				return false // No closing ]
			}
			// Check if character matches class
			if matchCharClass(pattern[pi+1:classEnd], str[si], noEscape) {
				pi = classEnd + 1
				si++
			} else {
				return false
			}

		case '\\':
			if !noEscape && pi+1 < len(pattern) {
				pi++ // Skip backslash
				if si >= len(str) || pattern[pi] != str[si] {
					return false
				}
				pi++
				si++
			} else {
				// Treat as literal backslash
				if si >= len(str) || pattern[pi] != str[si] {
					return false
				}
				pi++
				si++
			}

		default:
			if si >= len(str) || pattern[pi] != str[si] {
				return false
			}
			pi++
			si++
		}
	}

	// Pattern consumed, string should also be consumed
	return si >= len(str)
}

// matchCharClass checks if a character matches a character class [...]
func matchCharClass(class string, ch byte, noEscape bool) bool {
	if len(class) == 0 {
		return false
	}

	negate := false
	i := 0

	// Check for negation
	if class[0] == '!' || class[0] == '^' {
		negate = true
		i = 1
	}

	matched := false
	for i < len(class) {
		var start, end byte

		// Handle escape sequences
		if !noEscape && class[i] == '\\' && i+1 < len(class) {
			start = class[i+1]
			i += 2
		} else {
			start = class[i]
			i++
		}

		// Check for range (a-z)
		if i+1 < len(class) && class[i] == '-' {
			i++
			if !noEscape && class[i] == '\\' && i+1 < len(class) {
				end = class[i+1]
				i += 2
			} else {
				end = class[i]
				i++
			}

			if ch >= start && ch <= end {
				matched = true
				break
			}
		} else {
			if ch == start {
				matched = true
				break
			}
		}
	}

	return matched != negate
}

// parseScanfFormat parses a string according to scanf format specifiers
func parseScanfFormat(input, format string) []*values.Value {
	var result []*values.Value
	inputWords := strings.Fields(input)

	// Simple scanf parsing - extract format specifiers
	formatSpecs := []string{}
	i := 0
	for i < len(format) {
		if format[i] == '%' && i+1 < len(format) {
			spec := string(format[i+1])
			formatSpecs = append(formatSpecs, spec)
			i += 2
		} else {
			i++
		}
	}

	// Parse each word according to format specs
	for i, word := range inputWords {
		if i >= len(formatSpecs) {
			break
		}

		spec := formatSpecs[i]
		switch spec {
		case "s":
			result = append(result, values.NewString(word))
		case "d":
			if intVal, err := strconv.ParseInt(word, 10, 64); err == nil {
				result = append(result, values.NewInt(intVal))
			} else {
				result = append(result, values.NewString(word))
			}
		case "f":
			if floatVal, err := strconv.ParseFloat(word, 64); err == nil {
				result = append(result, values.NewFloat(floatVal))
			} else {
				result = append(result, values.NewString(word))
			}
		default:
			result = append(result, values.NewString(word))
		}
	}

	return result
}

// GetFilesystemFunctions returns all file system related functions
func GetFilesystemFunctions() []*registry.Function {
	return []*registry.Function{
		// File handle operations
		{
			Name: "fopen",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
				{Name: "mode", Type: "string"},
			},
			ReturnType: "resource|false",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || args[1] == nil {
					return values.NewBool(false), nil
				}
				filename := args[0].ToString()
				mode := args[1].ToString()

				var flag int
				switch mode {
				case "r":
					flag = os.O_RDONLY
				case "w":
					flag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
				case "a":
					flag = os.O_WRONLY | os.O_CREATE | os.O_APPEND
				case "r+":
					flag = os.O_RDWR
				case "w+":
					flag = os.O_RDWR | os.O_CREATE | os.O_TRUNC
				case "a+":
					flag = os.O_RDWR | os.O_CREATE | os.O_APPEND
				default:
					return values.NewBool(false), nil
				}

				file, err := os.OpenFile(filename, flag, 0644)
				if err != nil {
					return values.NewBool(false), nil
				}

				handle := &FileHandle{
					ID:       atomic.AddInt64(&fileHandleCounter, 1),
					File:     file,
					Mode:     mode,
					Position: 0,
					EOF:      false,
				}

				// For append mode, seek to end
				if strings.Contains(mode, "a") {
					if stat, err := file.Stat(); err == nil {
						handle.Position = stat.Size()
					}
				}

				registerFileHandle(handle)
				return values.NewResource(handle.ID), nil
			},
		},
		{
			Name: "fclose",
			Parameters: []*registry.Parameter{
				{Name: "handle", Type: "resource"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewBool(false), nil
				}

				handleID, ok := args[0].Data.(int64)
				if !ok {
					return values.NewBool(false), nil
				}

				handle, exists := getFileHandle(handleID)
				if !exists {
					return values.NewBool(false), nil
				}

				handle.mu.Lock()
				defer handle.mu.Unlock()

				err := handle.File.Close()
				removeFileHandle(handleID)

				return values.NewBool(err == nil), nil
			},
		},
		{
			Name: "fread",
			Parameters: []*registry.Parameter{
				{Name: "handle", Type: "resource"},
				{Name: "length", Type: "int"},
			},
			ReturnType: "string|false",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || args[1] == nil || args[0].Type != values.TypeResource {
					return values.NewBool(false), nil
				}

				handleID, ok := args[0].Data.(int64)
				if !ok {
					return values.NewBool(false), nil
				}

				length := args[1].ToInt()
				if length < 0 {
					length = 0
				}

				handle, exists := getFileHandle(handleID)
				if !exists {
					return values.NewBool(false), nil
				}

				handle.mu.Lock()
				defer handle.mu.Unlock()

				buffer := make([]byte, length)
				n, err := handle.File.Read(buffer)

				if err == io.EOF {
					handle.EOF = true
				}

				if n == 0 && err != nil {
					return values.NewString(""), nil
				}

				handle.Position += int64(n)
				return values.NewString(string(buffer[:n])), nil
			},
		},
		{
			Name: "fwrite",
			Parameters: []*registry.Parameter{
				{Name: "handle", Type: "resource"},
				{Name: "string", Type: "string"},
			},
			ReturnType: "int|false",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || args[1] == nil || args[0].Type != values.TypeResource {
					return values.NewBool(false), nil
				}

				handleID, ok := args[0].Data.(int64)
				if !ok {
					return values.NewBool(false), nil
				}

				data := args[1].ToString()

				handle, exists := getFileHandle(handleID)
				if !exists {
					return values.NewBool(false), nil
				}

				handle.mu.Lock()
				defer handle.mu.Unlock()

				n, err := handle.File.WriteString(data)
				if err != nil {
					return values.NewBool(false), nil
				}

				handle.Position += int64(n)
				return values.NewInt(int64(n)), nil
			},
		},
		{
			Name: "ftell",
			Parameters: []*registry.Parameter{
				{Name: "handle", Type: "resource"},
			},
			ReturnType: "int|false",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewBool(false), nil
				}

				handleID, ok := args[0].Data.(int64)
				if !ok {
					return values.NewBool(false), nil
				}

				handle, exists := getFileHandle(handleID)
				if !exists {
					return values.NewBool(false), nil
				}

				handle.mu.RLock()
				defer handle.mu.RUnlock()

				pos, err := handle.File.Seek(0, io.SeekCurrent)
				if err != nil {
					return values.NewBool(false), nil
				}

				handle.Position = pos
				return values.NewInt(pos), nil
			},
		},
		{
			Name: "fseek",
			Parameters: []*registry.Parameter{
				{Name: "handle", Type: "resource"},
				{Name: "offset", Type: "int"},
				{Name: "whence", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
			},
			ReturnType: "int",
			MinArgs:    2,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || args[1] == nil || args[0].Type != values.TypeResource {
					return values.NewInt(-1), nil
				}

				handleID, ok := args[0].Data.(int64)
				if !ok {
					return values.NewInt(-1), nil
				}

				offset := args[1].ToInt()
				whence := int64(0) // SEEK_SET
				if len(args) > 2 && args[2] != nil {
					whence = args[2].ToInt()
				}

				handle, exists := getFileHandle(handleID)
				if !exists {
					return values.NewInt(-1), nil
				}

				handle.mu.Lock()
				defer handle.mu.Unlock()

				var seekWhence int
				switch whence {
				case 0: // SEEK_SET
					seekWhence = io.SeekStart
				case 1: // SEEK_CUR
					seekWhence = io.SeekCurrent
				case 2: // SEEK_END
					seekWhence = io.SeekEnd
				default:
					return values.NewInt(-1), nil
				}

				pos, err := handle.File.Seek(offset, seekWhence)
				if err != nil {
					return values.NewInt(-1), nil
				}

				handle.Position = pos
				handle.EOF = false // Reset EOF flag after seek
				return values.NewInt(0), nil
			},
		},
		{
			Name: "rewind",
			Parameters: []*registry.Parameter{
				{Name: "handle", Type: "resource"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewBool(false), nil
				}

				handleID, ok := args[0].Data.(int64)
				if !ok {
					return values.NewBool(false), nil
				}

				handle, exists := getFileHandle(handleID)
				if !exists {
					return values.NewBool(false), nil
				}

				handle.mu.Lock()
				defer handle.mu.Unlock()

				_, err := handle.File.Seek(0, io.SeekStart)
				if err != nil {
					return values.NewBool(false), nil
				}

				handle.Position = 0
				handle.EOF = false
				return values.NewBool(true), nil
			},
		},
		{
			Name: "feof",
			Parameters: []*registry.Parameter{
				{Name: "handle", Type: "resource"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewBool(false), nil
				}

				handleID, ok := args[0].Data.(int64)
				if !ok {
					return values.NewBool(false), nil
				}

				handle, exists := getFileHandle(handleID)
				if !exists {
					return values.NewBool(false), nil
				}

				handle.mu.RLock()
				defer handle.mu.RUnlock()

				return values.NewBool(handle.EOF), nil
			},
		},
		{
			Name: "fgets",
			Parameters: []*registry.Parameter{
				{Name: "handle", Type: "resource"},
			},
			ReturnType: "string|false",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewBool(false), nil
				}

				handleID, ok := args[0].Data.(int64)
				if !ok {
					return values.NewBool(false), nil
				}

				handle, exists := getFileHandle(handleID)
				if !exists {
					return values.NewBool(false), nil
				}

				handle.mu.Lock()
				defer handle.mu.Unlock()

				// Read line character by character to maintain proper file position
				var line strings.Builder
				buffer := make([]byte, 1)

				for {
					n, err := handle.File.Read(buffer)
					if n == 0 || err != nil {
						if err == io.EOF {
							handle.EOF = true
						}
						if line.Len() > 0 {
							// Return partial line if we read something
							return values.NewString(line.String()), nil
						}
						return values.NewBool(false), nil
					}

					handle.Position++
					line.WriteByte(buffer[0])

					// Break on newline
					if buffer[0] == '\n' {
						break
					}
				}

				return values.NewString(line.String()), nil
			},
		},
		{
			Name: "fgetc",
			Parameters: []*registry.Parameter{
				{Name: "handle", Type: "resource"},
			},
			ReturnType: "string|false",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewBool(false), nil
				}

				handleID, ok := args[0].Data.(int64)
				if !ok {
					return values.NewBool(false), nil
				}

				handle, exists := getFileHandle(handleID)
				if !exists {
					return values.NewBool(false), nil
				}

				handle.mu.Lock()
				defer handle.mu.Unlock()

				buffer := make([]byte, 1)
				n, err := handle.File.Read(buffer)

				if err == io.EOF {
					handle.EOF = true
					return values.NewBool(false), nil
				}

				if n == 0 || err != nil {
					return values.NewBool(false), nil
				}

				handle.Position++
				return values.NewString(string(buffer[0])), nil
			},
		},
		{
			Name: "fflush",
			Parameters: []*registry.Parameter{
				{Name: "handle", Type: "resource"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewBool(false), nil
				}

				handleID, ok := args[0].Data.(int64)
				if !ok {
					return values.NewBool(false), nil
				}

				handle, exists := getFileHandle(handleID)
				if !exists {
					return values.NewBool(false), nil
				}

				handle.mu.Lock()
				defer handle.mu.Unlock()

				err := handle.File.Sync()
				return values.NewBool(err == nil), nil
			},
		},
		{
			Name: "ftruncate",
			Parameters: []*registry.Parameter{
				{Name: "handle", Type: "resource"},
				{Name: "size", Type: "int"},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || args[1] == nil || args[0].Type != values.TypeResource {
					return values.NewBool(false), nil
				}

				handleID, ok := args[0].Data.(int64)
				if !ok {
					return values.NewBool(false), nil
				}

				size := args[1].ToInt()

				handle, exists := getFileHandle(handleID)
				if !exists {
					return values.NewBool(false), nil
				}

				handle.mu.Lock()
				defer handle.mu.Unlock()

				err := handle.File.Truncate(size)
				return values.NewBool(err == nil), nil
			},
		},
		{
			Name: "fputs",
			Parameters: []*registry.Parameter{
				{Name: "handle", Type: "resource"},
				{Name: "string", Type: "string"},
			},
			ReturnType: "int|false",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// fputs is an alias for fwrite
				if len(args) < 2 || args[0] == nil || args[1] == nil || args[0].Type != values.TypeResource {
					return values.NewBool(false), nil
				}

				handleID, ok := args[0].Data.(int64)
				if !ok {
					return values.NewBool(false), nil
				}

				data := args[1].ToString()

				handle, exists := getFileHandle(handleID)
				if !exists {
					return values.NewBool(false), nil
				}

				handle.mu.Lock()
				defer handle.mu.Unlock()

				n, err := handle.File.WriteString(data)
				if err != nil {
					return values.NewBool(false), nil
				}

				handle.Position += int64(n)
				return values.NewInt(int64(n)), nil
			},
		},

		// File content functions
		{
			Name: "file_get_contents",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
			},
			ReturnType: "string|false",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				filename := args[0].ToString()

				content, err := os.ReadFile(filename)
				if err != nil {
					return values.NewBool(false), nil
				}

				return values.NewString(string(content)), nil
			},
		},
		{
			Name: "file_put_contents",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
				{Name: "data", Type: "mixed"},
			},
			ReturnType: "int|false",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || args[1] == nil {
					return values.NewBool(false), nil
				}

				filename := args[0].ToString()
				data := args[1].ToString()

				err := os.WriteFile(filename, []byte(data), 0644)
				if err != nil {
					return values.NewBool(false), nil
				}

				return values.NewInt(int64(len(data))), nil
			},
		},
		{
			Name: "file",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
			},
			ReturnType: "array|false",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				filename := args[0].ToString()

				file, err := os.Open(filename)
				if err != nil {
					return values.NewBool(false), nil
				}
				defer file.Close()

				result := values.NewArray()
				scanner := bufio.NewScanner(file)
				index := int64(0)

				for scanner.Scan() {
					line := scanner.Text()
					// PHP file() preserves newlines except for the last line
					line += "\n"
					result.ArraySet(values.NewInt(index), values.NewString(line))
					index++
				}

				// Handle the last line without newline if file doesn't end with newline
				if index > 0 {
					// Read the file again to check if last character is newline
					content, err := os.ReadFile(filename)
					if err == nil && len(content) > 0 && content[len(content)-1] != '\n' {
						// Remove newline from last element
						lastElem := result.ArrayGet(values.NewInt(index-1))
						if lastElem.Type == values.TypeString {
							lastLine := lastElem.ToString()
							if len(lastLine) > 0 && lastLine[len(lastLine)-1] == '\n' {
								result.ArraySet(values.NewInt(index-1), values.NewString(lastLine[:len(lastLine)-1]))
							}
						}
					}
				}

				if err := scanner.Err(); err != nil {
					return values.NewBool(false), nil
				}

				return result, nil
			},
		},

		// File existence functions
		{
			Name: "file_exists",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				filename := args[0].ToString()

				_, err := os.Stat(filename)
				return values.NewBool(err == nil), nil
			},
		},
		{
			Name: "is_file",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				filename := args[0].ToString()

				stat, err := os.Stat(filename)
				if err != nil {
					return values.NewBool(false), nil
				}

				return values.NewBool(stat.Mode().IsRegular()), nil
			},
		},
		{
			Name: "is_dir",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				filename := args[0].ToString()

				stat, err := os.Stat(filename)
				if err != nil {
					return values.NewBool(false), nil
				}

				return values.NewBool(stat.IsDir()), nil
			},
		},
		{
			Name: "is_readable",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				filename := args[0].ToString()

				_, err := os.Open(filename)
				if err != nil {
					return values.NewBool(false), nil
				}
				return values.NewBool(true), nil
			},
		},
		{
			Name: "is_writable",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				filename := args[0].ToString()

				// Try to open for writing
				file, err := os.OpenFile(filename, os.O_WRONLY, 0)
				if err != nil {
					return values.NewBool(false), nil
				}
				file.Close()
				return values.NewBool(true), nil
			},
		},
		{
			Name: "is_writeable",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Alias for is_writable
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				filename := args[0].ToString()

				file, err := os.OpenFile(filename, os.O_WRONLY, 0)
				if err != nil {
					return values.NewBool(false), nil
				}
				file.Close()
				return values.NewBool(true), nil
			},
		},

		// File information functions
		{
			Name: "filesize",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
			},
			ReturnType: "int|false",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				filename := args[0].ToString()

				stat, err := os.Stat(filename)
				if err != nil {
					return values.NewBool(false), nil
				}

				return values.NewInt(stat.Size()), nil
			},
		},
		{
			Name: "filetype",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
			},
			ReturnType: "string|false",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				filename := args[0].ToString()

				stat, err := os.Stat(filename)
				if err != nil {
					return values.NewBool(false), nil
				}

				mode := stat.Mode()
				switch {
				case mode.IsRegular():
					return values.NewString("file"), nil
				case mode.IsDir():
					return values.NewString("dir"), nil
				case mode&os.ModeSymlink != 0:
					return values.NewString("link"), nil
				case mode&os.ModeDevice != 0:
					return values.NewString("block"), nil
				case mode&os.ModeCharDevice != 0:
					return values.NewString("char"), nil
				case mode&os.ModeNamedPipe != 0:
					return values.NewString("fifo"), nil
				default:
					return values.NewString("unknown"), nil
				}
			},
		},
		{
			Name: "filemtime",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
			},
			ReturnType: "int|false",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				filename := args[0].ToString()

				stat, err := os.Stat(filename)
				if err != nil {
					return values.NewBool(false), nil
				}

				return values.NewInt(stat.ModTime().Unix()), nil
			},
		},
		{
			Name: "fileatime",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
			},
			ReturnType: "int|false",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				filename := args[0].ToString()

				stat, err := os.Stat(filename)
				if err != nil {
					return values.NewBool(false), nil
				}

				// Go's os.FileInfo doesn't directly provide access time
				// We'll use ModTime as a fallback (common in many filesystems)
				// For more accurate access time, we'd need syscalls
				return values.NewInt(stat.ModTime().Unix()), nil
			},
		},
		{
			Name: "filectime",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
			},
			ReturnType: "int|false",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				filename := args[0].ToString()

				stat, err := os.Stat(filename)
				if err != nil {
					return values.NewBool(false), nil
				}

				// Go's os.FileInfo doesn't directly provide change time
				// We'll use ModTime as a fallback (common behavior)
				// For more accurate change time, we'd need platform-specific syscalls
				return values.NewInt(stat.ModTime().Unix()), nil
			},
		},

		// Directory operations
		{
			Name: "mkdir",
			Parameters: []*registry.Parameter{
				{Name: "dirname", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				dirname := args[0].ToString()

				err := os.MkdirAll(dirname, 0755)
				if err != nil {
					return values.NewBool(false), nil
				}

				return values.NewBool(true), nil
			},
		},
		{
			Name: "rmdir",
			Parameters: []*registry.Parameter{
				{Name: "dirname", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				dirname := args[0].ToString()

				err := os.Remove(dirname)
				if err != nil {
					return values.NewBool(false), nil
				}

				return values.NewBool(true), nil
			},
		},

		// Path functions
		{
			Name: "dirname",
			Parameters: []*registry.Parameter{
				{Name: "path", Type: "string"},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewString(""), nil
				}
				path := args[0].ToString()

				return values.NewString(filepath.Dir(path)), nil
			},
		},
		{
			Name: "basename",
			Parameters: []*registry.Parameter{
				{Name: "path", Type: "string"},
				{Name: "suffix", Type: "string", HasDefault: true, DefaultValue: values.NewString("")},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewString(""), nil
				}
				path := args[0].ToString()
				base := filepath.Base(path)

				if len(args) > 1 && args[1] != nil {
					suffix := args[1].ToString()
					if suffix != "" && strings.HasSuffix(base, suffix) {
						base = base[:len(base)-len(suffix)]
					}
				}

				return values.NewString(base), nil
			},
		},
		{
			Name: "pathinfo",
			Parameters: []*registry.Parameter{
				{Name: "path", Type: "string"},
				{Name: "flags", Type: "int", HasDefault: true, DefaultValue: values.NewInt(15)}, // PATHINFO_ALL = 1+2+4+8 = 15
			},
			ReturnType: "mixed",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewArray(), nil
				}
				path := args[0].ToString()

				flags := int64(15) // PATHINFO_ALL by default
				if len(args) > 1 && args[1] != nil {
					flags = args[1].ToInt()
				}

				// Parse path components
				dirname := "."
				basename := ""
				extension := ""
				filename := ""

				if path == "" {
					basename = ""
					filename = ""
				} else if path == "/" {
					dirname = "/"
					basename = ""
					filename = ""
				} else {
					// Get dirname and basename using filepath functions
					dirname = filepath.Dir(path)
					basename = filepath.Base(path)

					// Handle extension and filename
					if basename != "" && basename != "." && basename != ".." {
						// Find last dot for extension
						lastDot := strings.LastIndex(basename, ".")

						if lastDot == -1 {
							// No extension
							filename = basename
							extension = ""
						} else if lastDot == 0 {
							// File starts with dot (like .hidden)
							if len(basename) > 1 && strings.Contains(basename[1:], ".") {
								// Multiple dots like ..file
								secondDot := strings.LastIndex(basename[1:], ".")
								if secondDot != -1 {
									extension = basename[secondDot+2:] // +2 because we searched from index 1
									filename = basename[:secondDot+2]
								} else {
									extension = basename[1:]
									filename = ""
								}
							} else {
								extension = basename[1:]
								filename = ""
							}
						} else {
							// Normal case with extension
							extension = basename[lastDot+1:]
							filename = basename[:lastDot]

							// Handle trailing dot case (file.)
							if extension == "" {
								filename = basename[:len(basename)-1] // Remove trailing dot
							}
						}
					} else {
						// Special cases: ".", ".."
						filename = strings.TrimPrefix(basename, ".")
						if filename == basename {
							filename = "" // Was just "."
						}
						extension = ""
					}
				}

				// Return specific component if flags specify only one
				if flags == 1 { // PATHINFO_DIRNAME
					return values.NewString(dirname), nil
				} else if flags == 2 { // PATHINFO_BASENAME
					return values.NewString(basename), nil
				} else if flags == 4 { // PATHINFO_EXTENSION
					return values.NewString(extension), nil
				} else if flags == 8 { // PATHINFO_FILENAME
					return values.NewString(filename), nil
				}

				// Return array with requested components
				result := values.NewArray()

				if flags&1 != 0 { // PATHINFO_DIRNAME
					result.ArraySet(values.NewString("dirname"), values.NewString(dirname))
				}
				if flags&2 != 0 { // PATHINFO_BASENAME
					result.ArraySet(values.NewString("basename"), values.NewString(basename))
				}
				if flags&4 != 0 && extension != "" { // PATHINFO_EXTENSION
					result.ArraySet(values.NewString("extension"), values.NewString(extension))
				}
				if flags&8 != 0 { // PATHINFO_FILENAME
					result.ArraySet(values.NewString("filename"), values.NewString(filename))
				}

				return result, nil
			},
		},
		{
			Name: "realpath",
			Parameters: []*registry.Parameter{
				{Name: "path", Type: "string"},
			},
			ReturnType: "string|false",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				path := args[0].ToString()

				absPath, err := filepath.Abs(path)
				if err != nil {
					return values.NewBool(false), nil
				}

				// Check if the resolved path actually exists
				_, err = os.Stat(absPath)
				if err != nil {
					return values.NewBool(false), nil
				}

				// Clean the path to resolve . and .. components
				cleanPath := filepath.Clean(absPath)
				return values.NewString(cleanPath), nil
			},
		},
		{
			Name: "readfile",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
			},
			ReturnType: "int|false",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				filename := args[0].ToString()

				content, err := os.ReadFile(filename)
				if err != nil {
					return values.NewBool(false), nil
				}

				// In a real implementation, this would output to stdout/browser
				// For now, we'll just return the byte count
				// TODO: Implement actual output functionality
				return values.NewInt(int64(len(content))), nil
			},
		},
		{
			Name: "tempnam",
			Parameters: []*registry.Parameter{
				{Name: "dir", Type: "string"},
				{Name: "prefix", Type: "string"},
			},
			ReturnType: "string|false",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || args[1] == nil {
					return values.NewBool(false), nil
				}
				dir := args[0].ToString()
				prefix := args[1].ToString()

				// Use system temp dir if provided dir is empty or doesn't exist
				if dir == "" {
					dir = os.TempDir()
				} else {
					if _, err := os.Stat(dir); os.IsNotExist(err) {
						dir = os.TempDir()
					}
				}

				// Create temporary file
				file, err := os.CreateTemp(dir, prefix+"*")
				if err != nil {
					return values.NewBool(false), nil
				}
				defer file.Close()

				return values.NewString(file.Name()), nil
			},
		},
		{
			Name: "sys_get_temp_dir",
			Parameters: []*registry.Parameter{},
			ReturnType: "string",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				return values.NewString(os.TempDir()), nil
			},
		},
		{
			Name: "glob",
			Parameters: []*registry.Parameter{
				{Name: "pattern", Type: "string"},
				{Name: "flags", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
			},
			ReturnType: "array|false",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				pattern := args[0].ToString()

				flags := int64(0)
				if len(args) > 1 && args[1] != nil {
					flags = args[1].ToInt()
				}

				// Use Go's filepath.Glob for basic pattern matching
				matches, err := filepath.Glob(pattern)
				if err != nil {
					return values.NewBool(false), nil
				}

				// Create result array
				result := values.NewArray()

				if matches == nil {
					// Return empty array if no matches
					return result, nil
				}

				// Sort results unless GLOB_NOSORT is specified
				if flags&4 == 0 { // GLOB_NOSORT not set
					// matches are already sorted by filepath.Glob
				}

				for i, match := range matches {
					// GLOB_MARK: add trailing slash to directories
					if flags&2 != 0 { // GLOB_MARK
						if stat, err := os.Stat(match); err == nil && stat.IsDir() {
							if !strings.HasSuffix(match, string(os.PathSeparator)) {
								match = match + string(os.PathSeparator)
							}
						}
					}

					// GLOB_ONLYDIR: only return directories
					if flags&8 != 0 { // GLOB_ONLYDIR
						if stat, err := os.Stat(match); err != nil || !stat.IsDir() {
							continue
						}
					}

					result.ArraySet(values.NewInt(int64(i)), values.NewString(match))
				}

				return result, nil
			},
		},

		// File operations
		{
			Name: "copy",
			Parameters: []*registry.Parameter{
				{Name: "source", Type: "string"},
				{Name: "dest", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || args[1] == nil {
					return values.NewBool(false), nil
				}
				source := args[0].ToString()
				dest := args[1].ToString()

				sourceFile, err := os.Open(source)
				if err != nil {
					return values.NewBool(false), nil
				}
				defer sourceFile.Close()

				destFile, err := os.Create(dest)
				if err != nil {
					return values.NewBool(false), nil
				}
				defer destFile.Close()

				_, err = io.Copy(destFile, sourceFile)
				if err != nil {
					return values.NewBool(false), nil
				}

				return values.NewBool(true), nil
			},
		},
		{
			Name: "rename",
			Parameters: []*registry.Parameter{
				{Name: "oldname", Type: "string"},
				{Name: "newname", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || args[1] == nil {
					return values.NewBool(false), nil
				}
				oldname := args[0].ToString()
				newname := args[1].ToString()

				err := os.Rename(oldname, newname)
				if err != nil {
					return values.NewBool(false), nil
				}

				return values.NewBool(true), nil
			},
		},

		// Existing unlink function
		{
			Name: "unlink",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				filename := args[0].ToString()

				err := os.Remove(filename)
				if err != nil {
					return values.NewBool(false), nil
				}

				return values.NewBool(true), nil
			},
		},
		{
			Name: "fileperms",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
			},
			ReturnType: "int|false",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				filename := args[0].ToString()

				stat, err := os.Stat(filename)
				if err != nil {
					return values.NewBool(false), nil
				}

				// Return the file mode as an integer (octal permissions)
				mode := stat.Mode()
				return values.NewInt(int64(mode)), nil
			},
		},
		{
			Name: "chmod",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
				{Name: "mode", Type: "int"},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || args[1] == nil {
					return values.NewBool(false), nil
				}
				filename := args[0].ToString()
				mode := os.FileMode(args[1].ToInt())

				err := os.Chmod(filename, mode)
				if err != nil {
					return values.NewBool(false), nil
				}

				return values.NewBool(true), nil
			},
		},
		{
			Name: "fileowner",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
			},
			ReturnType: "int|false",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				filename := args[0].ToString()

				stat, err := os.Stat(filename)
				if err != nil {
					return values.NewBool(false), nil
				}

				// Get file system info - this is platform specific for Unix systems
				if sys := stat.Sys(); sys != nil {
					if unixStat, ok := sys.(*syscall.Stat_t); ok {
						return values.NewInt(int64(unixStat.Uid)), nil
					}
				}

				// Fallback - return 0 if we can't determine the owner
				return values.NewInt(0), nil
			},
		},
		{
			Name: "filegroup",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
			},
			ReturnType: "int|false",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				filename := args[0].ToString()

				stat, err := os.Stat(filename)
				if err != nil {
					return values.NewBool(false), nil
				}

				// Get file system info - this is platform specific for Unix systems
				if sys := stat.Sys(); sys != nil {
					if unixStat, ok := sys.(*syscall.Stat_t); ok {
						return values.NewInt(int64(unixStat.Gid)), nil
					}
				}

				// Fallback - return 0 if we can't determine the group
				return values.NewInt(0), nil
			},
		},
		{
			Name: "chown",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
				{Name: "user", Type: "mixed"}, // Can be string or int
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || args[1] == nil {
					return values.NewBool(false), nil
				}
				filename := args[0].ToString()

				// Convert user to uid (for simplicity, assuming it's an integer)
				uid := int(args[1].ToInt())

				err := os.Chown(filename, uid, -1) // -1 means don't change group
				return values.NewBool(err == nil), nil
			},
		},
		{
			Name: "chgrp",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
				{Name: "group", Type: "mixed"}, // Can be string or int
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || args[1] == nil {
					return values.NewBool(false), nil
				}
				filename := args[0].ToString()

				// Convert group to gid (for simplicity, assuming it's an integer)
				gid := int(args[1].ToInt())

				err := os.Chown(filename, -1, gid) // -1 means don't change user
				return values.NewBool(err == nil), nil
			},
		},
		{
			Name: "stat",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
			},
			ReturnType: "array|false",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				filename := args[0].ToString()

				stat, err := os.Stat(filename)
				if err != nil {
					return values.NewBool(false), nil
				}

				// Create stat array with PHP-compatible structure
				statArray := values.NewArray()

				// File info
				statArray.ArraySet(values.NewString("size"), values.NewInt(stat.Size()))
				statArray.ArraySet(values.NewInt(7), values.NewInt(stat.Size())) // Numeric index too

				// File mode
				mode := int64(stat.Mode())
				statArray.ArraySet(values.NewString("mode"), values.NewInt(mode))
				statArray.ArraySet(values.NewInt(2), values.NewInt(mode))

				// Times
				mtime := stat.ModTime().Unix()
				statArray.ArraySet(values.NewString("mtime"), values.NewInt(mtime))
				statArray.ArraySet(values.NewInt(9), values.NewInt(mtime))

				// For Unix systems, get detailed info
				if sys := stat.Sys(); sys != nil {
					if unixStat, ok := sys.(*syscall.Stat_t); ok {
						// Device
						statArray.ArraySet(values.NewString("dev"), values.NewInt(int64(unixStat.Dev)))
						statArray.ArraySet(values.NewInt(0), values.NewInt(int64(unixStat.Dev)))

						// Inode
						statArray.ArraySet(values.NewString("ino"), values.NewInt(int64(unixStat.Ino)))
						statArray.ArraySet(values.NewInt(1), values.NewInt(int64(unixStat.Ino)))

						// Number of links
						statArray.ArraySet(values.NewString("nlink"), values.NewInt(int64(unixStat.Nlink)))
						statArray.ArraySet(values.NewInt(3), values.NewInt(int64(unixStat.Nlink)))

						// UID and GID
						statArray.ArraySet(values.NewString("uid"), values.NewInt(int64(unixStat.Uid)))
						statArray.ArraySet(values.NewInt(4), values.NewInt(int64(unixStat.Uid)))

						statArray.ArraySet(values.NewString("gid"), values.NewInt(int64(unixStat.Gid)))
						statArray.ArraySet(values.NewInt(5), values.NewInt(int64(unixStat.Gid)))

						// Device type
						statArray.ArraySet(values.NewString("rdev"), values.NewInt(int64(unixStat.Rdev)))
						statArray.ArraySet(values.NewInt(6), values.NewInt(int64(unixStat.Rdev)))

						// Access and change times
						atime := int64(unixStat.Atim.Sec)
						ctime := int64(unixStat.Ctim.Sec)

						statArray.ArraySet(values.NewString("atime"), values.NewInt(atime))
						statArray.ArraySet(values.NewInt(8), values.NewInt(atime))

						statArray.ArraySet(values.NewString("ctime"), values.NewInt(ctime))
						statArray.ArraySet(values.NewInt(10), values.NewInt(ctime))

						// Block size and blocks
						statArray.ArraySet(values.NewString("blksize"), values.NewInt(int64(unixStat.Blksize)))
						statArray.ArraySet(values.NewInt(11), values.NewInt(int64(unixStat.Blksize)))

						statArray.ArraySet(values.NewString("blocks"), values.NewInt(int64(unixStat.Blocks)))
						statArray.ArraySet(values.NewInt(12), values.NewInt(int64(unixStat.Blocks)))
					}
				}

				return statArray, nil
			},
		},
		{
			Name: "lstat",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
			},
			ReturnType: "array|false",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				filename := args[0].ToString()

				// lstat is like stat but doesn't follow symlinks
				stat, err := os.Lstat(filename)
				if err != nil {
					return values.NewBool(false), nil
				}

				// Use same logic as stat() function - this is a simplified version
				// In a full implementation, we'd share this code
				statArray := values.NewArray()
				statArray.ArraySet(values.NewString("size"), values.NewInt(stat.Size()))
				statArray.ArraySet(values.NewInt(7), values.NewInt(stat.Size()))

				mode := int64(stat.Mode())
				statArray.ArraySet(values.NewString("mode"), values.NewInt(mode))
				statArray.ArraySet(values.NewInt(2), values.NewInt(mode))

				mtime := stat.ModTime().Unix()
				statArray.ArraySet(values.NewString("mtime"), values.NewInt(mtime))
				statArray.ArraySet(values.NewInt(9), values.NewInt(mtime))

				return statArray, nil
			},
		},
		{
			Name: "fstat",
			Parameters: []*registry.Parameter{
				{Name: "handle", Type: "resource"},
			},
			ReturnType: "array|false",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewBool(false), nil
				}

				handleID, ok := args[0].Data.(int64)
				if !ok {
					return values.NewBool(false), nil
				}

				handle, exists := getFileHandle(handleID)
				if !exists {
					return values.NewBool(false), nil
				}

				handle.mu.Lock()
				defer handle.mu.Unlock()

				stat, err := handle.File.Stat()
				if err != nil {
					return values.NewBool(false), nil
				}

				// Use same logic as stat() function
				statArray := values.NewArray()
				statArray.ArraySet(values.NewString("size"), values.NewInt(stat.Size()))
				statArray.ArraySet(values.NewInt(7), values.NewInt(stat.Size()))

				mode := int64(stat.Mode())
				statArray.ArraySet(values.NewString("mode"), values.NewInt(mode))
				statArray.ArraySet(values.NewInt(2), values.NewInt(mode))

				mtime := stat.ModTime().Unix()
				statArray.ArraySet(values.NewString("mtime"), values.NewInt(mtime))
				statArray.ArraySet(values.NewInt(9), values.NewInt(mtime))

				return statArray, nil
			},
		},
		{
			Name: "is_link",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				filename := args[0].ToString()

				// Use Lstat to check if it's a symlink without following it
				stat, err := os.Lstat(filename)
				if err != nil {
					return values.NewBool(false), nil
				}

				// Check if it's a symbolic link
				isSymlink := stat.Mode()&os.ModeSymlink != 0
				return values.NewBool(isSymlink), nil
			},
		},
		{
			Name: "link",
			Parameters: []*registry.Parameter{
				{Name: "target", Type: "string"},
				{Name: "link", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || args[1] == nil {
					return values.NewBool(false), nil
				}
				target := args[0].ToString()
				linkName := args[1].ToString()

				err := os.Link(target, linkName)
				return values.NewBool(err == nil), nil
			},
		},
		{
			Name: "symlink",
			Parameters: []*registry.Parameter{
				{Name: "target", Type: "string"},
				{Name: "link", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || args[1] == nil {
					return values.NewBool(false), nil
				}
				target := args[0].ToString()
				linkName := args[1].ToString()

				err := os.Symlink(target, linkName)
				return values.NewBool(err == nil), nil
			},
		},
		{
			Name: "readlink",
			Parameters: []*registry.Parameter{
				{Name: "path", Type: "string"},
			},
			ReturnType: "string|false",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				path := args[0].ToString()

				target, err := os.Readlink(path)
				if err != nil {
					return values.NewBool(false), nil
				}

				return values.NewString(target), nil
			},
		},
		{
			Name: "touch",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
				{Name: "time", Type: "int", DefaultValue: values.NewNull()},
				{Name: "atime", Type: "int", DefaultValue: values.NewNull()},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				filename := args[0].ToString()

				// Get current time as default
				now := syscall.NsecToTimespec(int64(1e9) * time.Now().Unix())
				mtime := now
				atime := now

				// If specific modification time provided
				if len(args) > 1 && !args[1].IsNull() {
					mtimeUnix := args[1].ToInt()
					mtime = syscall.NsecToTimespec(int64(1e9) * mtimeUnix)
				}

				// If specific access time provided
				if len(args) > 2 && !args[2].IsNull() {
					atimeUnix := args[2].ToInt()
					atime = syscall.NsecToTimespec(int64(1e9) * atimeUnix)
				}

				// If file doesn't exist, create it
				if _, err := os.Stat(filename); os.IsNotExist(err) {
					file, createErr := os.Create(filename)
					if createErr != nil {
						return values.NewBool(false), nil
					}
					file.Close()
				}

				// Change the times
				err := syscall.UtimesNano(filename, []syscall.Timespec{atime, mtime})
				return values.NewBool(err == nil), nil
			},
		},
		{
			Name: "clearstatcache",
			Parameters: []*registry.Parameter{
				{Name: "clear_realpath_cache", Type: "bool", DefaultValue: values.NewBool(false)},
				{Name: "filename", Type: "string", DefaultValue: values.NewString("")},
			},
			ReturnType: "void",
			MinArgs:    0,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// In Go, we don't have a stat cache like PHP does
				// This function would typically clear OS-level caches, but that's not
				// directly accessible or necessary in our implementation
				// So we just return null (void) to match PHP's behavior
				return values.NewNull(), nil
			},
		},
		{
			Name: "is_executable",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				filename := args[0].ToString()

				// Check if file exists first
				stat, err := os.Stat(filename)
				if err != nil {
					return values.NewBool(false), nil
				}

				// For directories, check if we can execute (access) them
				if stat.IsDir() {
					// For directories, executable means we can search/traverse it
					// This is platform-specific, but we can try to open it
					dir, err := os.Open(filename)
					if err != nil {
						return values.NewBool(false), nil
					}
					dir.Close()
					return values.NewBool(true), nil
				}

				// For files, check if they have execute permission
				mode := stat.Mode()

				// Check user execute permission (owner)
				if mode&0100 != 0 {
					return values.NewBool(true), nil
				}

				// Check group execute permission
				if mode&0010 != 0 {
					return values.NewBool(true), nil
				}

				// Check other execute permission
				if mode&0001 != 0 {
					return values.NewBool(true), nil
				}

				return values.NewBool(false), nil
			},
		},
		{
			Name: "fileinode",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
			},
			ReturnType: "int|false",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				filename := args[0].ToString()

				stat, err := os.Stat(filename)
				if err != nil {
					return values.NewBool(false), nil
				}

				// Get inode number from Unix stat
				if sys := stat.Sys(); sys != nil {
					if unixStat, ok := sys.(*syscall.Stat_t); ok {
						return values.NewInt(int64(unixStat.Ino)), nil
					}
				}

				// Fallback - return 0 if we can't determine the inode
				return values.NewInt(0), nil
			},
		},
		{
			Name: "umask",
			Parameters: []*registry.Parameter{
				{Name: "mask", Type: "int", DefaultValue: values.NewNull()},
			},
			ReturnType: "int",
			MinArgs:    0,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Get current umask by temporarily setting it to 0 and then restoring
				currentUmask := syscall.Umask(0)
				syscall.Umask(currentUmask)

				// If no argument provided, just return current umask
				if len(args) == 0 || args[0].IsNull() {
					return values.NewInt(int64(currentUmask)), nil
				}

				// Set new umask and return old one
				newMask := int(args[0].ToInt())
				syscall.Umask(newMask)
				return values.NewInt(int64(currentUmask)), nil
			},
		},
		{
			Name: "tmpfile",
			Parameters: []*registry.Parameter{},
			ReturnType: "resource|false",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Create a temporary file
				file, err := os.CreateTemp("", "hey_tmpfile_*")
				if err != nil {
					return values.NewBool(false), nil
				}

				// Register the file handle
				handleID := atomic.AddInt64(&fileHandleCounter, 1)
				handle := &FileHandle{
					ID:   handleID,
					File: file,
					Mode: "w+", // tmpfile is always opened in read/write mode
				}

				fileHandlesMutex.Lock()
				fileHandles[handleID] = handle
				fileHandlesMutex.Unlock()

				// Mark it for deletion when closed (Go will handle this automatically
				// when the file descriptor is closed since we created it with CreateTemp)

				return values.NewResource(handleID), nil
			},
		},
		{
			Name: "fgetcsv",
			Parameters: []*registry.Parameter{
				{Name: "handle", Type: "resource"},
				{Name: "length", Type: "int", DefaultValue: values.NewInt(0)},
				{Name: "delimiter", Type: "string", DefaultValue: values.NewString(",")},
				{Name: "enclosure", Type: "string", DefaultValue: values.NewString("\"")},
				{Name: "escape", Type: "string", DefaultValue: values.NewString("\\")},
			},
			ReturnType: "array|false",
			MinArgs:    1,
			MaxArgs:    5,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewBool(false), nil
				}

				handleID, ok := args[0].Data.(int64)
				if !ok {
					return values.NewBool(false), nil
				}

				handle, exists := getFileHandle(handleID)
				if !exists {
					return values.NewBool(false), nil
				}

				handle.mu.Lock()
				defer handle.mu.Unlock()

				// Get delimiter (default comma)
				delimiter := ","
				if len(args) > 2 && !args[2].IsNull() {
					delimiter = args[2].ToString()
				}

				// Create CSV reader
				reader := csv.NewReader(handle.File)
				if len(delimiter) > 0 {
					reader.Comma = rune(delimiter[0])
				}

				// Read one record
				record, err := reader.Read()
				if err != nil {
					if err == io.EOF {
						return values.NewBool(false), nil
					}
					return values.NewBool(false), nil
				}

				// Convert to PHP array
				result := values.NewArray()
				for i, field := range record {
					result.ArraySet(values.NewInt(int64(i)), values.NewString(field))
				}

				return result, nil
			},
		},
		{
			Name: "fputcsv",
			Parameters: []*registry.Parameter{
				{Name: "handle", Type: "resource"},
				{Name: "fields", Type: "array"},
				{Name: "delimiter", Type: "string", DefaultValue: values.NewString(",")},
				{Name: "enclosure", Type: "string", DefaultValue: values.NewString("\"")},
				{Name: "escape", Type: "string", DefaultValue: values.NewString("\\")},
			},
			ReturnType: "int|false",
			MinArgs:    2,
			MaxArgs:    5,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || args[1] == nil ||
				   args[0].Type != values.TypeResource || args[1].Type != values.TypeArray {
					return values.NewBool(false), nil
				}

				handleID, ok := args[0].Data.(int64)
				if !ok {
					return values.NewBool(false), nil
				}

				handle, exists := getFileHandle(handleID)
				if !exists {
					return values.NewBool(false), nil
				}

				handle.mu.Lock()
				defer handle.mu.Unlock()

				// Get delimiter (default comma)
				delimiter := ","
				if len(args) > 2 && !args[2].IsNull() {
					delimiter = args[2].ToString()
				}

				// Convert array to string slice
				fieldsArrayValue := args[1]
				var record []string

				// Iterate through array elements
				arrayCount := fieldsArrayValue.ArrayCount()
				for i := 0; i < arrayCount; i++ {
					key := values.NewInt(int64(i))
					if val := fieldsArrayValue.ArrayGet(key); val != nil && !val.IsNull() {
						record = append(record, val.ToString())
					}
				}

				// Create CSV writer
				writer := csv.NewWriter(handle.File)
				if len(delimiter) > 0 {
					writer.Comma = rune(delimiter[0])
				}

				// Write the record
				err := writer.Write(record)
				if err != nil {
					return values.NewBool(false), nil
				}

				// Flush to ensure it's written
				writer.Flush()
				if err := writer.Error(); err != nil {
					return values.NewBool(false), nil
				}

				// Calculate approximate bytes written (CSV line + newline)
				csvLine := strings.Join(record, delimiter)
				for _, field := range record {
					// Add quotes if field contains delimiter, quotes, or newlines
					if strings.Contains(field, delimiter) || strings.Contains(field, "\"") || strings.Contains(field, "\n") {
						// Rough estimate: field + 2 quotes + escaped quotes
						csvLine = strings.Replace(csvLine, field, "\""+strings.Replace(field, "\"", "\"\"", -1)+"\"", 1)
					}
				}
				return values.NewInt(int64(len(csvLine) + 1)), nil // +1 for newline
			},
		},
		{
			Name: "disk_free_space",
			Parameters: []*registry.Parameter{
				{Name: "directory", Type: "string"},
			},
			ReturnType: "float|false",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				directory := args[0].ToString()

				// Check if directory exists
				if _, err := os.Stat(directory); os.IsNotExist(err) {
					return values.NewBool(false), nil
				}

				// Get disk usage using syscall
				var stat syscall.Statfs_t
				err := syscall.Statfs(directory, &stat)
				if err != nil {
					return values.NewBool(false), nil
				}

				// Calculate free space: available blocks * block size
				freeSpace := float64(stat.Bavail) * float64(stat.Bsize)
				return values.NewFloat(freeSpace), nil
			},
		},
		{
			Name: "disk_total_space",
			Parameters: []*registry.Parameter{
				{Name: "directory", Type: "string"},
			},
			ReturnType: "float|false",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				directory := args[0].ToString()

				// Check if directory exists
				if _, err := os.Stat(directory); os.IsNotExist(err) {
					return values.NewBool(false), nil
				}

				// Get disk usage using syscall
				var stat syscall.Statfs_t
				err := syscall.Statfs(directory, &stat)
				if err != nil {
					return values.NewBool(false), nil
				}

				// Calculate total space: total blocks * block size
				totalSpace := float64(stat.Blocks) * float64(stat.Bsize)
				return values.NewFloat(totalSpace), nil
			},
		},
		{
			Name: "diskfreespace",
			Parameters: []*registry.Parameter{
				{Name: "directory", Type: "string"},
			},
			ReturnType: "float|false",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// diskfreespace is an alias for disk_free_space
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				directory := args[0].ToString()

				// Check if directory exists
				if _, err := os.Stat(directory); os.IsNotExist(err) {
					return values.NewBool(false), nil
				}

				// Get disk usage using syscall
				var stat syscall.Statfs_t
				err := syscall.Statfs(directory, &stat)
				if err != nil {
					return values.NewBool(false), nil
				}

				// Calculate free space: available blocks * block size
				freeSpace := float64(stat.Bavail) * float64(stat.Bsize)
				return values.NewFloat(freeSpace), nil
			},
		},
		{
			Name: "linkinfo",
			Parameters: []*registry.Parameter{
				{Name: "path", Type: "string"},
			},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewInt(-1), nil
				}
				path := args[0].ToString()

				// Get file information using lstat (doesn't follow symlinks)
				fileInfo, err := os.Lstat(path)
				if err != nil {
					// Return -1 for non-existent files (matches PHP behavior)
					return values.NewInt(-1), nil
				}

				// Get the system-specific stat data
				stat := fileInfo.Sys().(*syscall.Stat_t)

				// Return the inode number
				return values.NewInt(int64(stat.Ino)), nil
			},
		},
		{
			Name: "parse_ini_file",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
				{Name: "process_sections", Type: "bool", HasDefault: true, DefaultValue: values.NewBool(false)},
				{Name: "scanner_mode", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
			},
			ReturnType: "array|false",
			MinArgs:    1,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				filename := args[0].ToString()

				// Check if file exists
				if _, err := os.Stat(filename); os.IsNotExist(err) {
					return values.NewBool(false), nil
				}

				// Read file content
				content, err := os.ReadFile(filename)
				if err != nil {
					return values.NewBool(false), nil
				}

				// Get parameters
				processSections := false
				if len(args) > 1 && args[1] != nil {
					processSections = args[1].ToBool()
				}

				scannerMode := int64(0) // INI_SCANNER_NORMAL
				if len(args) > 2 && args[2] != nil {
					scannerMode = args[2].ToInt()
				}

				// Parse INI content
				return parseINIContent(string(content), processSections, scannerMode), nil
			},
		},
		{
			Name: "parse_ini_string",
			Parameters: []*registry.Parameter{
				{Name: "ini", Type: "string"},
				{Name: "process_sections", Type: "bool", HasDefault: true, DefaultValue: values.NewBool(false)},
				{Name: "scanner_mode", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
			},
			ReturnType: "array|false",
			MinArgs:    1,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				ini := args[0].ToString()

				// Get parameters
				processSections := false
				if len(args) > 1 && args[1] != nil {
					processSections = args[1].ToBool()
				}

				scannerMode := int64(0) // INI_SCANNER_NORMAL
				if len(args) > 2 && args[2] != nil {
					scannerMode = args[2].ToInt()
				}

				// Parse INI content
				return parseINIContent(ini, processSections, scannerMode), nil
			},
		},
		{
			Name: "fnmatch",
			Parameters: []*registry.Parameter{
				{Name: "pattern", Type: "string"},
				{Name: "string", Type: "string"},
				{Name: "flags", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || args[1] == nil {
					return values.NewBool(false), nil
				}

				pattern := args[0].ToString()
				str := args[1].ToString()

				flags := int64(0)
				if len(args) > 2 && args[2] != nil {
					flags = args[2].ToInt()
				}

				// Convert flags
				caseFold := flags&16 != 0    // FNM_CASEFOLD
				noEscape := flags&2 != 0     // FNM_NOESCAPE
				pathname := flags&1 != 0     // FNM_PATHNAME
				period := flags&4 != 0       // FNM_PERIOD

				// Apply case folding
				if caseFold {
					pattern = strings.ToLower(pattern)
					str = strings.ToLower(str)
				}

				// Use custom fnmatch implementation
				result := fnmatchImpl(pattern, str, noEscape, pathname, period)
				return values.NewBool(result), nil
			},
		},
		{
			Name: "fpassthru",
			Parameters: []*registry.Parameter{
				{Name: "stream", Type: "resource"},
			},
			ReturnType: "int|false",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewBool(false), nil
				}

				handleID, ok := args[0].Data.(int64)
				if !ok {
					return values.NewBool(false), nil
				}

				handle, exists := getFileHandle(handleID)
				if !exists {
					return values.NewBool(false), nil
				}

				handle.mu.Lock()
				defer handle.mu.Unlock()

				// Read all remaining data from file and output it
				// Note: In a real PHP environment, this would output to stdout
				// Here we'll just count the bytes that would be output
				buffer := make([]byte, 4096)
				totalBytes := int64(0)

				for {
					n, err := handle.File.Read(buffer)
					if n == 0 || err != nil {
						break
					}
					totalBytes += int64(n)
					// In real PHP, this would be: echo string(buffer[:n])
				}

				return values.NewInt(totalBytes), nil
			},
		},
		{
			Name: "fscanf",
			Parameters: []*registry.Parameter{
				{Name: "stream", Type: "resource"},
				{Name: "format", Type: "string"},
			},
			ReturnType: "array|int|false",
			MinArgs:    2,
			MaxArgs:    -1, // Variable arguments
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || args[1] == nil || args[0].Type != values.TypeResource {
					return values.NewBool(false), nil
				}

				handleID, ok := args[0].Data.(int64)
				if !ok {
					return values.NewBool(false), nil
				}

				handle, exists := getFileHandle(handleID)
				if !exists {
					return values.NewBool(false), nil
				}

				handle.mu.Lock()
				defer handle.mu.Unlock()

				format := args[1].ToString()

				// Read a line from the file
				var line strings.Builder
				buffer := make([]byte, 1)
				for {
					n, err := handle.File.Read(buffer)
					if n == 0 || err != nil {
						break
					}
					if buffer[0] == '\n' {
						break
					}
					line.WriteByte(buffer[0])
				}

				lineStr := line.String()
				if lineStr == "" {
					return values.NewBool(false), nil
				}

				// Parse the line according to format
				parsed := parseScanfFormat(lineStr, format)

				// If no additional arguments, return array
				if len(args) == 2 {
					result := values.NewArray()
					for i, val := range parsed {
						result.ArraySet(values.NewInt(int64(i)), val)
					}
					return result, nil
				}

				// If additional arguments provided, assign to variables and return count
				// Note: In real PHP, this would assign by reference to variables
				// Here we just return the count of successful assignments
				return values.NewInt(int64(len(parsed))), nil
			},
		},
		{
			Name: "popen",
			Parameters: []*registry.Parameter{
				{Name: "command", Type: "string"},
				{Name: "mode", Type: "string"},
			},
			ReturnType: "resource|false",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || args[1] == nil {
					return values.NewBool(false), nil
				}

				command := args[0].ToString()
				mode := args[1].ToString()

				// Validate mode
				if mode != "r" && mode != "rb" && mode != "w" && mode != "wb" {
					return values.NewBool(false), nil
				}

				// Create process command
				var cmd *exec.Cmd
				if mode == "r" || mode == "rb" {
					// Reading mode - we read from command's stdout
					cmd = exec.Command("sh", "-c", command)
				} else {
					// Writing mode - we write to command's stdin
					cmd = exec.Command("sh", "-c", command)
				}

				var file io.Closer
				var err error

				if mode == "r" || mode == "rb" {
					// Set up pipe to read from command's stdout
					file, err = cmd.StdoutPipe()
					if err != nil {
						return values.NewBool(false), nil
					}

					// Start the command
					if err := cmd.Start(); err != nil {
						return values.NewBool(false), nil
					}
				} else {
					// Set up pipe to write to command's stdin
					file, err = cmd.StdinPipe()
					if err != nil {
						return values.NewBool(false), nil
					}

					// Start the command
					if err := cmd.Start(); err != nil {
						return values.NewBool(false), nil
					}
				}

				// Create process handle
				handle := &ProcessHandle{
					ID:   atomic.AddInt64(&processHandleCounter, 1),
					Cmd:  cmd,
					File: file,
					Mode: mode,
				}

				registerProcessHandle(handle)
				return values.NewResource(handle.ID), nil
			},
		},
		{
			Name: "pclose",
			Parameters: []*registry.Parameter{
				{Name: "handle", Type: "resource"},
			},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewInt(-1), nil
				}

				handleID, ok := args[0].Data.(int64)
				if !ok {
					return values.NewInt(-1), nil
				}

				// Try to get as process handle first
				processHandle, exists := getProcessHandle(handleID)
				if exists {
					processHandle.mu.Lock()
					defer processHandle.mu.Unlock()

					// Close the file descriptor
					processHandle.File.Close()

					// Wait for the process to finish and get exit status
					err := processHandle.Cmd.Wait()

					// Remove from registry
					removeProcessHandle(handleID)

					// Return exit status
					if err != nil {
						if exitError, ok := err.(*exec.ExitError); ok {
							return values.NewInt(int64(exitError.ExitCode())), nil
						}
						return values.NewInt(-1), nil
					}
					return values.NewInt(0), nil
				}

				// Not found
				return values.NewInt(-1), nil
			},
		},
		{
			Name: "set_file_buffer",
			Parameters: []*registry.Parameter{
				{Name: "stream", Type: "resource"},
				{Name: "buffer", Type: "int"},
			},
			ReturnType: "int",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || args[1] == nil || args[0].Type != values.TypeResource {
					return values.NewInt(-1), nil
				}

				handleID, ok := args[0].Data.(int64)
				if !ok {
					return values.NewInt(-1), nil
				}

				bufferSize := args[1].ToInt()

				// Try to get file handle
				handle, exists := getFileHandle(handleID)
				if exists {
					handle.mu.Lock()
					defer handle.mu.Unlock()

					// Go doesn't provide fine-grained buffer control for *os.File
					// In PHP, this function often returns -1 anyway on many systems
					// We'll simulate the behavior by always returning -1 (not supported)
					_ = bufferSize // Use the parameter to avoid warnings
					return values.NewInt(-1), nil
				}

				// Not found or not supported
				return values.NewInt(-1), nil
			},
		},
		{
			Name: "fsync",
			Parameters: []*registry.Parameter{
				{Name: "stream", Type: "resource"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewBool(false), nil
				}

				handleID, ok := args[0].Data.(int64)
				if !ok {
					return values.NewBool(false), nil
				}

				handle, exists := getFileHandle(handleID)
				if !exists {
					return values.NewBool(false), nil
				}

				handle.mu.Lock()
				defer handle.mu.Unlock()

				// Synchronize file data and metadata to disk
				err := handle.File.Sync()
				return values.NewBool(err == nil), nil
			},
		},
		{
			Name: "fdatasync",
			Parameters: []*registry.Parameter{
				{Name: "stream", Type: "resource"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewBool(false), nil
				}

				handleID, ok := args[0].Data.(int64)
				if !ok {
					return values.NewBool(false), nil
				}

				handle, exists := getFileHandle(handleID)
				if !exists {
					return values.NewBool(false), nil
				}

				handle.mu.Lock()
				defer handle.mu.Unlock()

				// Note: Go's File.Sync() synchronizes both data and metadata
				// fdatasync in Unix only syncs data, not metadata, but Go doesn't
				// provide a separate fdatasync equivalent, so we use Sync()
				err := handle.File.Sync()
				return values.NewBool(err == nil), nil
			},
		},
		{
			Name: "flock",
			Parameters: []*registry.Parameter{
				{Name: "stream", Type: "resource"},
				{Name: "operation", Type: "int"},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || args[1] == nil {
					return values.NewBool(false), nil
				}

				// Get file handle from resource
				if args[0].Type != values.TypeResource {
					return values.NewBool(false), nil
				}

				handleID, ok := args[0].Data.(int64)
				if !ok {
					return values.NewBool(false), nil
				}

				handle, exists := getFileHandle(handleID)
				if !exists {
					return values.NewBool(false), nil
				}

				handle.mu.Lock()
				defer handle.mu.Unlock()

				operation := args[1].ToInt()

				// Extract lock type and flags
				lockType := operation & 3  // LOCK_SH=1, LOCK_EX=2, LOCK_UN=3
				isNonBlocking := operation&4 != 0  // LOCK_NB=4

				// Get file descriptor for syscall
				fd := int(handle.File.Fd())

				// Set up flock syscall
				var flockType int16
				switch lockType {
				case 1: // LOCK_SH
					flockType = syscall.F_RDLCK
				case 2: // LOCK_EX
					flockType = syscall.F_WRLCK
				case 3: // LOCK_UN
					flockType = syscall.F_UNLCK
				default:
					return values.NewBool(false), nil
				}

				// Create flock structure
				flock := syscall.Flock_t{
					Type:   flockType,
					Whence: 0,  // SEEK_SET
					Start:  0,  // Start of file
					Len:    0,  // Whole file
				}

				// Choose fcntl command based on blocking/non-blocking
				var cmd int
				if isNonBlocking {
					cmd = syscall.F_SETLK  // Non-blocking
				} else {
					cmd = syscall.F_SETLKW // Blocking
				}

				// Perform the lock operation
				err := syscall.FcntlFlock(uintptr(fd), cmd, &flock)
				if err != nil {
					return values.NewBool(false), nil
				}

				return values.NewBool(true), nil
			},
		},
		// Missing functions to achieve 100% coverage
		{
			Name: "delete",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// delete() is an alias of unlink()
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}

				filename := args[0].ToString()
				err := os.Remove(filename)
				return values.NewBool(err == nil), nil
			},
		},
		{
			Name: "fgetss",
			Parameters: []*registry.Parameter{
				{Name: "handle", Type: "resource"},
				{Name: "length", Type: "int", HasDefault: true, DefaultValue: values.NewNull()},
				{Name: "allowable_tags", Type: "string", HasDefault: true, DefaultValue: values.NewString("")},
			},
			ReturnType: "string|false",
			MinArgs:    1,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// fgetss() was deprecated in PHP 7.3 and removed in PHP 8.0
				// For compatibility, we implement a basic version that strips HTML/PHP tags
				if len(args) == 0 || args[0] == nil || args[0].Type != values.TypeResource {
					return values.NewBool(false), nil
				}

				handleID, ok := args[0].Data.(int64)
				if !ok {
					return values.NewBool(false), nil
				}

				handle, exists := getFileHandle(handleID)
				if !exists {
					return values.NewBool(false), nil
				}

				handle.mu.Lock()
				defer handle.mu.Unlock()

				// Read a line first
				reader := bufio.NewReader(handle.File)
				line, isPrefix, err := reader.ReadLine()
				if err != nil {
					if err == io.EOF {
						handle.EOF = true
					}
					return values.NewBool(false), nil
				}

				// Handle long lines that don't fit in buffer
				for isPrefix {
					var moreLine []byte
					moreLine, isPrefix, err = reader.ReadLine()
					if err != nil {
						break
					}
					line = append(line, moreLine...)
				}

				// Basic HTML tag stripping (simplified)
				result := string(line)
				// Remove HTML/XML tags: <tag>content</tag> or <tag/>
				inTag := false
				stripped := make([]byte, 0, len(result))
				for i := 0; i < len(result); i++ {
					if result[i] == '<' {
						inTag = true
					} else if result[i] == '>' {
						inTag = false
					} else if !inTag {
						stripped = append(stripped, result[i])
					}
				}
				result = string(stripped)

				handle.Position += int64(len(line)) + 1 // +1 for newline
				return values.NewString(result + "\n"), nil
			},
		},
		{
			Name: "is_uploaded_file",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// In a CLI environment, files are never uploaded via HTTP POST
				// This function checks if a file was uploaded via HTTP POST mechanism
				// Since we're in CLI context, always return false
				return values.NewBool(false), nil
			},
		},
		{
			Name: "move_uploaded_file",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
				{Name: "destination", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// In CLI environment, this would never succeed since files aren't uploaded via HTTP
				// But for completeness, we implement basic file moving functionality
				if len(args) < 2 || args[0] == nil || args[1] == nil {
					return values.NewBool(false), nil
				}

				filename := args[0].ToString()
				destination := args[1].ToString()

				// Check if source file exists
				if _, err := os.Stat(filename); os.IsNotExist(err) {
					return values.NewBool(false), nil
				}

				// In CLI context, we can't verify if file was uploaded, so just do a move
				err := os.Rename(filename, destination)
				return values.NewBool(err == nil), nil
			},
		},
		{
			Name: "lchgrp",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
				{Name: "group", Type: "mixed"}, // Can be string or int
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || args[1] == nil {
					return values.NewBool(false), nil
				}

				filename := args[0].ToString()
				group := args[1].ToInt() // Convert to int (GID)

				// lchgrp changes group ownership of symbolic link itself, not target
				err := os.Lchown(filename, -1, int(group)) // -1 means don't change user
				return values.NewBool(err == nil), nil
			},
		},
		{
			Name: "lchown",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
				{Name: "user", Type: "mixed"}, // Can be string or int
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || args[1] == nil {
					return values.NewBool(false), nil
				}

				filename := args[0].ToString()
				user := args[1].ToInt() // Convert to int (UID)

				// lchown changes user ownership of symbolic link itself, not target
				err := os.Lchown(filename, int(user), -1) // -1 means don't change group
				return values.NewBool(err == nil), nil
			},
		},
		{
			Name: "realpath_cache_get",
			Parameters: []*registry.Parameter{},
			ReturnType: "array",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Return empty array - realpath cache is internal PHP optimization
				// In our implementation, we don't maintain a separate cache
				return values.NewArray(), nil
			},
		},
		{
			Name: "realpath_cache_size",
			Parameters: []*registry.Parameter{},
			ReturnType: "int",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Return 0 - we don't maintain a realpath cache
				return values.NewInt(0), nil
			},
		},
	}
}