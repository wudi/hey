package spl

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetSplFileInfoClass returns the SplFileInfo class descriptor
func GetSplFileInfoClass() *registry.ClassDescriptor {
	// Constructor
	constructorImpl := &registry.Function{
		Name:      "__construct",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 {
				return values.NewNull(), fmt.Errorf("SplFileInfo::__construct() expects at least 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("__construct called on non-object")
			}

			// Get filename parameter
			var filename *values.Value
			if len(args) > 1 {
				filename = args[1]
			} else {
				// Use default value
				filename = values.NewString("")
			}

			if !filename.IsString() {
				return values.NewNull(), fmt.Errorf("SplFileInfo::__construct() expects parameter 1 to be string")
			}

			// Note: In the future, we should validate that filename is not empty
			// For now, allowing empty filename to work around VM parameter passing issue
			// if filename.ToString() == "" {
			//     return values.NewNull(), fmt.Errorf("SplFileInfo::__construct() expects filename to be non-empty")
			// }

			obj := thisObj.Data.(*values.Object)
			if obj.Properties == nil {
				obj.Properties = make(map[string]*values.Value)
			}

			// Store the filepath
			obj.Properties["__filepath"] = filename

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "filename", Type: "string"},
		},
	}

	// getFilename() method
	getFilenameImpl := &registry.Function{
		Name:      "getFilename",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("SplFileInfo::getFilename() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("getFilename called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if filePathVal, ok := obj.Properties["__filepath"]; ok && filePathVal.IsString() {
				path := filePathVal.ToString()
				return values.NewString(filepath.Base(path)), nil
			}

			return values.NewString(""), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// getBasename() method
	getBasenameImpl := &registry.Function{
		Name:      "getBasename",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || len(args) > 2 {
				return values.NewNull(), fmt.Errorf("SplFileInfo::getBasename() expects 1 or 2 arguments, %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("getBasename called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if filePathVal, ok := obj.Properties["__filepath"]; ok && filePathVal.IsString() {
				path := filePathVal.ToString()
				basename := filepath.Base(path)

				// If suffix is provided, remove it
				if len(args) == 2 && args[1].IsString() {
					suffix := args[1].ToString()
					if suffix != "" && strings.HasSuffix(basename, suffix) {
						basename = strings.TrimSuffix(basename, suffix)
					}
				}

				return values.NewString(basename), nil
			}

			return values.NewString(""), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "suffix", Type: "string", HasDefault: true, DefaultValue: values.NewString("")},
		},
	}

	// getExtension() method
	getExtensionImpl := &registry.Function{
		Name:      "getExtension",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("SplFileInfo::getExtension() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("getExtension called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if filePathVal, ok := obj.Properties["__filepath"]; ok && filePathVal.IsString() {
				path := filePathVal.ToString()
				ext := filepath.Ext(path)
				if ext != "" && ext[0] == '.' {
					ext = ext[1:] // Remove the dot
				}
				return values.NewString(ext), nil
			}

			return values.NewString(""), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// getPath() method
	getPathImpl := &registry.Function{
		Name:      "getPath",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("SplFileInfo::getPath() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("getPath called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if filePathVal, ok := obj.Properties["__filepath"]; ok && filePathVal.IsString() {
				path := filePathVal.ToString()
				return values.NewString(filepath.Dir(path)), nil
			}

			return values.NewString(""), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// getPathname() method
	getPathnameImpl := &registry.Function{
		Name:      "getPathname",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("SplFileInfo::getPathname() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("getPathname called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if filePathVal, ok := obj.Properties["__filepath"]; ok && filePathVal.IsString() {
				return filePathVal, nil
			}

			return values.NewString(""), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// getRealPath() method
	getRealPathImpl := &registry.Function{
		Name:      "getRealPath",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("SplFileInfo::getRealPath() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("getRealPath called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if filePathVal, ok := obj.Properties["__filepath"]; ok && filePathVal.IsString() {
				path := filePathVal.ToString()
				realpath, err := filepath.Abs(path)
				if err != nil {
					return values.NewBool(false), nil
				}

				// Check if file exists
				if _, err := os.Stat(realpath); os.IsNotExist(err) {
					return values.NewBool(false), nil
				}

				return values.NewString(realpath), nil
			}

			return values.NewBool(false), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// isFile() method
	isFileImpl := &registry.Function{
		Name:      "isFile",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("SplFileInfo::isFile() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("isFile called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if filePathVal, ok := obj.Properties["__filepath"]; ok && filePathVal.IsString() {
				path := filePathVal.ToString()
				info, err := os.Stat(path)
				if err != nil {
					return values.NewBool(false), nil
				}
				return values.NewBool(info.Mode().IsRegular()), nil
			}

			return values.NewBool(false), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// isDir() method
	isDirImpl := &registry.Function{
		Name:      "isDir",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("SplFileInfo::isDir() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("isDir called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if filePathVal, ok := obj.Properties["__filepath"]; ok && filePathVal.IsString() {
				path := filePathVal.ToString()
				info, err := os.Stat(path)
				if err != nil {
					return values.NewBool(false), nil
				}
				return values.NewBool(info.IsDir()), nil
			}

			return values.NewBool(false), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// isReadable() method
	isReadableImpl := &registry.Function{
		Name:      "isReadable",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("SplFileInfo::isReadable() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("isReadable called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if filePathVal, ok := obj.Properties["__filepath"]; ok && filePathVal.IsString() {
				path := filePathVal.ToString()
				// Try to open for reading
				file, err := os.Open(path)
				if err != nil {
					return values.NewBool(false), nil
				}
				file.Close()
				return values.NewBool(true), nil
			}

			return values.NewBool(false), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// isWritable() method
	isWritableImpl := &registry.Function{
		Name:      "isWritable",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("SplFileInfo::isWritable() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("isWritable called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if filePathVal, ok := obj.Properties["__filepath"]; ok && filePathVal.IsString() {
				path := filePathVal.ToString()

				// Check if file exists
				if _, err := os.Stat(path); os.IsNotExist(err) {
					// File doesn't exist, check parent directory
					dir := filepath.Dir(path)
					info, err := os.Stat(dir)
					if err != nil {
						return values.NewBool(false), nil
					}
					// Check if directory is writable (simplified)
					return values.NewBool(info.Mode().Perm()&0200 != 0), nil
				}

				// File exists, try to open for writing
				file, err := os.OpenFile(path, os.O_WRONLY, 0)
				if err != nil {
					return values.NewBool(false), nil
				}
				file.Close()
				return values.NewBool(true), nil
			}

			return values.NewBool(false), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// getSize() method
	getSizeImpl := &registry.Function{
		Name:      "getSize",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("SplFileInfo::getSize() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("getSize called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if filePathVal, ok := obj.Properties["__filepath"]; ok && filePathVal.IsString() {
				path := filePathVal.ToString()
				info, err := os.Stat(path)
				if err != nil {
					return values.NewInt(0), nil
				}
				return values.NewInt(info.Size()), nil
			}

			return values.NewInt(0), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// Create method descriptors
	methods := map[string]*registry.MethodDescriptor{
		"__construct": {
			Name:       "__construct",
			Visibility: "public",
			Parameters: []*registry.ParameterDescriptor{
				{Name: "filename", Type: "string", HasDefault: true, DefaultValue: values.NewString("")},
			},
			Implementation: NewBuiltinMethodImpl(constructorImpl),
		},
		"getFilename": {
			Name:           "getFilename",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(getFilenameImpl),
		},
		"getBasename": {
			Name:       "getBasename",
			Visibility: "public",
			Parameters: []*registry.ParameterDescriptor{
				{Name: "suffix", Type: "string", HasDefault: true, DefaultValue: values.NewString("")},
			},
			Implementation: NewBuiltinMethodImpl(getBasenameImpl),
		},
		"getExtension": {
			Name:           "getExtension",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(getExtensionImpl),
		},
		"getPath": {
			Name:           "getPath",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(getPathImpl),
		},
		"getPathname": {
			Name:           "getPathname",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(getPathnameImpl),
		},
		"getRealPath": {
			Name:           "getRealPath",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(getRealPathImpl),
		},
		"isFile": {
			Name:           "isFile",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(isFileImpl),
		},
		"isDir": {
			Name:           "isDir",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(isDirImpl),
		},
		"isReadable": {
			Name:           "isReadable",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(isReadableImpl),
		},
		"isWritable": {
			Name:           "isWritable",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(isWritableImpl),
		},
		"getSize": {
			Name:           "getSize",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(getSizeImpl),
		},
	}

	return &registry.ClassDescriptor{
		Name:       "SplFileInfo",
		Methods:    methods,
		Constants:  make(map[string]*registry.ConstantDescriptor),
		Interfaces: []string{},
	}
}