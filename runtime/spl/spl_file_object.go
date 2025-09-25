package spl

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// SplFileObject constants
const (
	DROP_NEW_LINE = 1
	READ_AHEAD    = 2
	SKIP_EMPTY    = 4
	READ_CSV      = 8
)

// SplFileObjectData holds the file object state
type SplFileObjectData struct {
	file         *os.File
	scanner      *bufio.Scanner
	lines        []string
	currentLine  int
	flags        int
	csvControl   []string // [separator, enclosure, escape]
	maxLineLen   int
	eof          bool
	filePath     string
}

// GetSplFileObjectClass returns the SplFileObject class descriptor
func GetSplFileObjectClass() *registry.ClassDescriptor {
	// Constructor
	constructorImpl := &registry.Function{
		Name:      "__construct",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("SplFileObject::__construct() expects at least 1 parameter, %d given", len(args)-1)
			}

			thisObj := args[0]
			filenameArg := args[1]
			filename := filenameArg.ToString()

			// Default mode is 'r'
			mode := "r"
			if len(args) >= 3 {
				mode = args[2].ToString()
			}

			// Open the file
			file, err := os.OpenFile(filename, getModeFlags(mode), 0666)
			if err != nil {
				return nil, fmt.Errorf("SplFileObject::__construct(%s): Failed to open file: %v", filename, err)
			}

			// Read all lines for iteration
			var lines []string
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				lines = append(lines, scanner.Text()+"\n")
			}

			// Add empty line at end if file doesn't end with newline
			if len(lines) > 0 && !strings.HasSuffix(lines[len(lines)-1], "\n") {
				lines[len(lines)-1] += "\n"
			}

			// Always add one more empty line (PHP behavior)
			lines = append(lines, "")

			// Reset file position
			file.Seek(0, 0)

			// Create file object data
			fileData := &SplFileObjectData{
				file:        file,
				scanner:     bufio.NewScanner(file),
				lines:       lines,
				currentLine: 0,
				flags:       0,
				csvControl:  []string{",", "\"", "\\"},
				maxLineLen:  0,
				eof:         false,
				filePath:    filename,
			}

			objData := thisObj.Data.(*values.Object)
			objData.Properties["_fileData"] = &values.Value{
				Type: values.TypeResource,
				Data: fileData,
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "filename", Type: "string"},
			{Name: "mode", Type: "string", DefaultValue: values.NewString("r")},
		},
	}

	// current() method
	currentImpl := &registry.Function{
		Name:      "current",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			fileData, err := getFileDataFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			if fileData.currentLine >= len(fileData.lines) {
				return values.NewString(""), nil
			}

			line := fileData.lines[fileData.currentLine]

			// Handle flags
			if (fileData.flags & READ_CSV) != 0 {
				// Parse as CSV
				reader := csv.NewReader(strings.NewReader(line))
				reader.Comma = rune(fileData.csvControl[0][0])
				// Note: Go's csv.Reader doesn't have Quote field, it uses the default quote character
				record, err := reader.Read()
				if err != nil && err != io.EOF {
					return values.NewString(line), nil
				}
				if err == io.EOF && len(record) == 0 {
					return values.NewNull(), nil
				}

				// Return array for CSV
				arr := values.NewArray()
				for i, field := range record {
					arr.ArraySet(values.NewInt(int64(i)), values.NewString(field))
				}
				return arr, nil
			}

			if (fileData.flags & DROP_NEW_LINE) != 0 {
				line = strings.TrimSuffix(line, "\n")
				line = strings.TrimSuffix(line, "\r")
			}

			return values.NewString(line), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// key() method
	keyImpl := &registry.Function{
		Name:      "key",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			fileData, err := getFileDataFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			return values.NewInt(int64(fileData.currentLine)), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// next() method
	nextImpl := &registry.Function{
		Name:      "next",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			fileData, err := getFileDataFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			fileData.currentLine++

			// Handle SKIP_EMPTY flag
			if (fileData.flags & SKIP_EMPTY) != 0 {
				for fileData.currentLine < len(fileData.lines) {
					line := strings.TrimSpace(fileData.lines[fileData.currentLine])
					if line != "" {
						break
					}
					fileData.currentLine++
				}
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// rewind() method
	rewindImpl := &registry.Function{
		Name:      "rewind",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			fileData, err := getFileDataFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			fileData.currentLine = 0
			fileData.file.Seek(0, 0)
			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// valid() method
	validImpl := &registry.Function{
		Name:      "valid",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			fileData, err := getFileDataFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			return values.NewBool(fileData.currentLine < len(fileData.lines)), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// seek() method
	seekImpl := &registry.Function{
		Name:      "seek",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("SplFileObject::seek() expects 1 parameter, %d given", len(args)-1)
			}

			thisObj := args[0]
			line := int(args[1].ToInt())

			fileData, err := getFileDataFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			if line < 0 || line >= len(fileData.lines) {
				return nil, fmt.Errorf("SplFileObject::seek(): Invalid line number %d", line)
			}

			fileData.currentLine = line
			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "line", Type: "int"},
		},
	}

	// eof() method
	eofImpl := &registry.Function{
		Name:      "eof",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			fileData, err := getFileDataFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			return values.NewBool(fileData.currentLine >= len(fileData.lines)), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// fgets() method
	fgetsImpl := &registry.Function{
		Name:      "fgets",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			fileData, err := getFileDataFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			if fileData.currentLine >= len(fileData.lines) {
				return values.NewString(""), nil
			}

			line := fileData.lines[fileData.currentLine]
			fileData.currentLine++
			return values.NewString(line), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// fread() method
	freadImpl := &registry.Function{
		Name:      "fread",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("SplFileObject::fread() expects 1 parameter, %d given", len(args)-1)
			}

			thisObj := args[0]
			length := int(args[1].ToInt())

			fileData, err := getFileDataFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			buffer := make([]byte, length)
			n, err := fileData.file.Read(buffer)
			if err != nil && err != io.EOF {
				return values.NewString(""), nil
			}

			return values.NewString(string(buffer[:n])), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "length", Type: "int"},
		},
	}

	// fwrite() method
	fwriteImpl := &registry.Function{
		Name:      "fwrite",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("SplFileObject::fwrite() expects 1 parameter, %d given", len(args)-1)
			}

			thisObj := args[0]
			data := args[1].ToString()

			length := len(data)
			if len(args) >= 3 && args[2].ToInt() > 0 {
				length = int(args[2].ToInt())
				if length > len(data) {
					length = len(data)
				}
			}

			fileData, err := getFileDataFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			n, err := fileData.file.Write([]byte(data[:length]))
			if err != nil {
				return values.NewInt(0), nil
			}

			return values.NewInt(int64(n)), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "str", Type: "string"},
			{Name: "length", Type: "int", DefaultValue: values.NewInt(0)},
		},
	}

	// ftell() method
	ftellImpl := &registry.Function{
		Name:      "ftell",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			fileData, err := getFileDataFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			pos, err := fileData.file.Seek(0, 1) // Get current position
			if err != nil {
				return values.NewInt(-1), nil
			}

			return values.NewInt(pos), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// fseek() method
	fseekImpl := &registry.Function{
		Name:      "fseek",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("SplFileObject::fseek() expects 1 parameter, %d given", len(args)-1)
			}

			thisObj := args[0]
			offset := args[1].ToInt()

			whence := 0 // SEEK_SET
			if len(args) >= 3 {
				whence = int(args[2].ToInt())
			}

			fileData, err := getFileDataFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			_, err = fileData.file.Seek(offset, whence)
			if err != nil {
				return values.NewInt(-1), nil
			}

			return values.NewInt(0), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "offset", Type: "int"},
			{Name: "whence", Type: "int", DefaultValue: values.NewInt(0)},
		},
	}

	// getFlags() method
	getFlagsImpl := &registry.Function{
		Name:      "getFlags",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			fileData, err := getFileDataFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			return values.NewInt(int64(fileData.flags)), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// setFlags() method
	setFlagsImpl := &registry.Function{
		Name:      "setFlags",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("SplFileObject::setFlags() expects 1 parameter, %d given", len(args)-1)
			}

			thisObj := args[0]
			flags := int(args[1].ToInt())

			fileData, err := getFileDataFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			fileData.flags = flags
			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "flags", Type: "int"},
		},
	}

	methods := make(map[string]*registry.MethodDescriptor)

	// Implement SplFileInfo methods that work with our file data
	getFilenameImpl := &registry.Function{
		Name:      "getFilename",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			fileData, err := getFileDataFromObject(thisObj)
			if err != nil {
				return nil, err
			}
			return values.NewString(filepath.Base(fileData.filePath)), nil
		},
		Parameters: []*registry.Parameter{},
	}

	getBasenameImpl := &registry.Function{
		Name:      "getBasename",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			fileData, err := getFileDataFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			basename := filepath.Base(fileData.filePath)
			if len(args) >= 2 {
				suffix := args[1].ToString()
				if suffix != "" && strings.HasSuffix(basename, suffix) {
					basename = basename[:len(basename)-len(suffix)]
				}
			}
			return values.NewString(basename), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "suffix", Type: "string", DefaultValue: values.NewString("")},
		},
	}

	getPathImpl := &registry.Function{
		Name:      "getPath",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			fileData, err := getFileDataFromObject(thisObj)
			if err != nil {
				return nil, err
			}
			return values.NewString(filepath.Dir(fileData.filePath)), nil
		},
		Parameters: []*registry.Parameter{},
	}

	getPathnameImpl := &registry.Function{
		Name:      "getPathname",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			fileData, err := getFileDataFromObject(thisObj)
			if err != nil {
				return nil, err
			}
			return values.NewString(fileData.filePath), nil
		},
		Parameters: []*registry.Parameter{},
	}

	getExtensionImpl := &registry.Function{
		Name:      "getExtension",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			fileData, err := getFileDataFromObject(thisObj)
			if err != nil {
				return nil, err
			}
			ext := filepath.Ext(fileData.filePath)
			if ext != "" {
				ext = ext[1:] // Remove the dot
			}
			return values.NewString(ext), nil
		},
		Parameters: []*registry.Parameter{},
	}

	getSizeImpl := &registry.Function{
		Name:      "getSize",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			fileData, err := getFileDataFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			info, err := os.Stat(fileData.filePath)
			if err != nil {
				return values.NewInt(0), nil
			}
			return values.NewInt(info.Size()), nil
		},
		Parameters: []*registry.Parameter{},
	}

	isFileImpl := &registry.Function{
		Name:      "isFile",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			fileData, err := getFileDataFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			info, err := os.Stat(fileData.filePath)
			if err != nil {
				return values.NewBool(false), nil
			}
			return values.NewBool(info.Mode().IsRegular()), nil
		},
		Parameters: []*registry.Parameter{},
	}

	isDirImpl := &registry.Function{
		Name:      "isDir",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			fileData, err := getFileDataFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			info, err := os.Stat(fileData.filePath)
			if err != nil {
				return values.NewBool(false), nil
			}
			return values.NewBool(info.IsDir()), nil
		},
		Parameters: []*registry.Parameter{},
	}

	isReadableImpl := &registry.Function{
		Name:      "isReadable",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			fileData, err := getFileDataFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			_, err = os.Open(fileData.filePath)
			return values.NewBool(err == nil), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// Override parent methods with our implementations
	methods["__construct"] = &registry.MethodDescriptor{
		Name:       "__construct",
		Visibility: "public",
		Parameters: []*registry.ParameterDescriptor{
			{Name: "filename", Type: "string"},
			{Name: "mode", Type: "string"},
		},
		Implementation: NewBuiltinMethodImpl(constructorImpl),
	}

	// Add iterator methods
	methods["current"] = &registry.MethodDescriptor{
		Name:           "current",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(currentImpl),
	}
	methods["key"] = &registry.MethodDescriptor{
		Name:           "key",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(keyImpl),
	}
	methods["next"] = &registry.MethodDescriptor{
		Name:           "next",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(nextImpl),
	}
	methods["rewind"] = &registry.MethodDescriptor{
		Name:           "rewind",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(rewindImpl),
	}
	methods["valid"] = &registry.MethodDescriptor{
		Name:           "valid",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(validImpl),
	}
	methods["seek"] = &registry.MethodDescriptor{
		Name:       "seek",
		Visibility: "public",
		Parameters: []*registry.ParameterDescriptor{
			{Name: "line", Type: "int"},
		},
		Implementation: NewBuiltinMethodImpl(seekImpl),
	}

	// Add file operation methods
	methods["eof"] = &registry.MethodDescriptor{
		Name:           "eof",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(eofImpl),
	}
	methods["fgets"] = &registry.MethodDescriptor{
		Name:           "fgets",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(fgetsImpl),
	}
	methods["fread"] = &registry.MethodDescriptor{
		Name:       "fread",
		Visibility: "public",
		Parameters: []*registry.ParameterDescriptor{
			{Name: "length", Type: "int"},
		},
		Implementation: NewBuiltinMethodImpl(freadImpl),
	}
	methods["fwrite"] = &registry.MethodDescriptor{
		Name:       "fwrite",
		Visibility: "public",
		Parameters: []*registry.ParameterDescriptor{
			{Name: "str", Type: "string"},
			{Name: "length", Type: "int"},
		},
		Implementation: NewBuiltinMethodImpl(fwriteImpl),
	}
	methods["ftell"] = &registry.MethodDescriptor{
		Name:           "ftell",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(ftellImpl),
	}
	methods["fseek"] = &registry.MethodDescriptor{
		Name:       "fseek",
		Visibility: "public",
		Parameters: []*registry.ParameterDescriptor{
			{Name: "offset", Type: "int"},
			{Name: "whence", Type: "int"},
		},
		Implementation: NewBuiltinMethodImpl(fseekImpl),
	}
	methods["getFlags"] = &registry.MethodDescriptor{
		Name:           "getFlags",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(getFlagsImpl),
	}
	methods["setFlags"] = &registry.MethodDescriptor{
		Name:       "setFlags",
		Visibility: "public",
		Parameters: []*registry.ParameterDescriptor{
			{Name: "flags", Type: "int"},
		},
		Implementation: NewBuiltinMethodImpl(setFlagsImpl),
	}

	// Add SplFileInfo methods
	methods["getFilename"] = &registry.MethodDescriptor{
		Name:           "getFilename",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(getFilenameImpl),
	}
	methods["getBasename"] = &registry.MethodDescriptor{
		Name:       "getBasename",
		Visibility: "public",
		Parameters: []*registry.ParameterDescriptor{
			{Name: "suffix", Type: "string"},
		},
		Implementation: NewBuiltinMethodImpl(getBasenameImpl),
	}
	methods["getPath"] = &registry.MethodDescriptor{
		Name:           "getPath",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(getPathImpl),
	}
	methods["getPathname"] = &registry.MethodDescriptor{
		Name:           "getPathname",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(getPathnameImpl),
	}
	methods["getExtension"] = &registry.MethodDescriptor{
		Name:           "getExtension",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(getExtensionImpl),
	}
	methods["getSize"] = &registry.MethodDescriptor{
		Name:           "getSize",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(getSizeImpl),
	}
	methods["isFile"] = &registry.MethodDescriptor{
		Name:           "isFile",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(isFileImpl),
	}
	methods["isDir"] = &registry.MethodDescriptor{
		Name:           "isDir",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(isDirImpl),
	}
	methods["isReadable"] = &registry.MethodDescriptor{
		Name:           "isReadable",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(isReadableImpl),
	}

	// Add constants
	constants := map[string]*registry.ConstantDescriptor{
		"DROP_NEW_LINE": {
			Name:       "DROP_NEW_LINE",
			Visibility: "public",
			Value:      values.NewInt(DROP_NEW_LINE),
			IsFinal:    true,
		},
		"READ_AHEAD": {
			Name:       "READ_AHEAD",
			Visibility: "public",
			Value:      values.NewInt(READ_AHEAD),
			IsFinal:    true,
		},
		"SKIP_EMPTY": {
			Name:       "SKIP_EMPTY",
			Visibility: "public",
			Value:      values.NewInt(SKIP_EMPTY),
			IsFinal:    true,
		},
		"READ_CSV": {
			Name:       "READ_CSV",
			Visibility: "public",
			Value:      values.NewInt(READ_CSV),
			IsFinal:    true,
		},
	}

	return &registry.ClassDescriptor{
		Name:       "SplFileObject",
		Parent:     "SplFileInfo",
		Interfaces: []string{"Iterator", "RecursiveIterator", "SeekableIterator"},
		Traits:     []string{},
		IsAbstract: false,
		IsFinal:    false,
		Methods:    methods,
		Properties: map[string]*registry.PropertyDescriptor{},
		Constants:  constants,
	}
}

// Helper function to get file data from object properties
func getFileDataFromObject(obj *values.Value) (*SplFileObjectData, error) {
	if obj.Type != values.TypeObject {
		return nil, fmt.Errorf("expected object, got %s", obj.Type)
	}

	objData := obj.Data.(*values.Object)
	fileDataValue, exists := objData.Properties["_fileData"]
	if !exists {
		return nil, fmt.Errorf("file data not initialized")
	}

	if fileDataValue.Type != values.TypeResource {
		return nil, fmt.Errorf("invalid file data property type")
	}

	fileData, ok := fileDataValue.Data.(*SplFileObjectData)
	if !ok {
		return nil, fmt.Errorf("invalid file data property data")
	}

	return fileData, nil
}

// getModeFlags converts PHP file mode to Go flags
func getModeFlags(mode string) int {
	switch mode {
	case "r":
		return os.O_RDONLY
	case "r+":
		return os.O_RDWR
	case "w":
		return os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	case "w+":
		return os.O_RDWR | os.O_CREATE | os.O_TRUNC
	case "a":
		return os.O_WRONLY | os.O_CREATE | os.O_APPEND
	case "a+":
		return os.O_RDWR | os.O_CREATE | os.O_APPEND
	case "x":
		return os.O_WRONLY | os.O_CREATE | os.O_EXCL
	case "x+":
		return os.O_RDWR | os.O_CREATE | os.O_EXCL
	case "c":
		return os.O_WRONLY | os.O_CREATE
	case "c+":
		return os.O_RDWR | os.O_CREATE
	default:
		return os.O_RDONLY
	}
}