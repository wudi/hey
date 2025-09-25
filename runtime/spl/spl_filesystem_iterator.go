package spl

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// FilesystemIterator constants
const (
	CURRENT_AS_PATHNAME = 32
	CURRENT_AS_FILEINFO = 0
	CURRENT_AS_SELF     = 16
	CURRENT_MODE_MASK   = 240
	KEY_AS_PATHNAME     = 0
	KEY_AS_FILENAME     = 256
	FOLLOW_SYMLINKS     = 512
	KEY_MODE_MASK       = 3840
	NEW_CURRENT_AND_KEY = 288
	SKIP_DOTS           = 4096
	UNIX_PATHS          = 8192
)

// FilesystemIterator represents a filesystem iterator
type FilesystemIterator struct {
	path     string
	entries  []os.DirEntry
	position int
	flags    int
}

// GetFilesystemIteratorClass returns the FilesystemIterator class descriptor
func GetFilesystemIteratorClass() *registry.ClassDescriptor {
	// Constructor
	constructorImpl := &registry.Function{
		Name:      "__construct",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("FilesystemIterator::__construct() expects 1 parameter, %d given", len(args)-1)
			}

			thisObj := args[0]
			pathArg := args[1]
			path := pathArg.ToString()

			// Default flags - skip dots by default (unlike DirectoryIterator)
			flags := SKIP_DOTS
			if len(args) >= 3 {
				flags = int(args[2].ToInt())
			}

			// Check if directory exists
			info, err := os.Stat(path)
			if err != nil {
				return nil, fmt.Errorf("FilesystemIterator::__construct(%s): Failed to open directory: %v", path, err)
			}
			if !info.IsDir() {
				return nil, fmt.Errorf("FilesystemIterator::__construct(%s): Not a directory", path)
			}

			// Read directory entries
			file, err := os.Open(path)
			if err != nil {
				return nil, fmt.Errorf("FilesystemIterator::__construct(%s): Failed to open directory: %v", path, err)
			}
			defer file.Close()

			entries, err := file.ReadDir(-1)
			if err != nil {
				return nil, fmt.Errorf("FilesystemIterator::__construct(%s): Failed to read directory: %v", path, err)
			}

			var allEntries []os.DirEntry

			// Add . and .. entries if SKIP_DOTS is not set
			if (flags & SKIP_DOTS) == 0 {
				dotEntry := &fakeFSDirEntry{name: ".", isDir: true}
				dotDotEntry := &fakeFSDirEntry{name: "..", isDir: true}
				allEntries = make([]os.DirEntry, 0, len(entries)+2)
				allEntries = append(allEntries, dotEntry, dotDotEntry)
				allEntries = append(allEntries, entries...)
			} else {
				allEntries = entries
			}

			// Create iterator
			iterator := &FilesystemIterator{
				path:     path,
				entries:  allEntries,
				position: 0,
				flags:    flags,
			}

			objData := thisObj.Data.(*values.Object)
			objData.Properties["_iterator"] = &values.Value{
				Type: values.TypeResource,
				Data: iterator,
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "path", Type: "string"},
			{Name: "flags", Type: "int", DefaultValue: values.NewInt(SKIP_DOTS)},
		},
	}

	// current() method - returns different types based on flags
	currentImpl := &registry.Function{
		Name:      "current",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			iterator, err := getFilesystemIteratorFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			if iterator.position < 0 || iterator.position >= len(iterator.entries) {
				return values.NewNull(), nil
			}

			entry := iterator.entries[iterator.position]
			currentMode := iterator.flags & CURRENT_MODE_MASK

			switch currentMode {
			case CURRENT_AS_PATHNAME:
				fullPath := filepath.Join(iterator.path, entry.Name())
				return values.NewString(fullPath), nil
			case CURRENT_AS_SELF:
				// Return the iterator itself
				return thisObj, nil
			default: // CURRENT_AS_FILEINFO (default)
				// Return the current DirectoryIterator object itself (like PHP)
				return thisObj, nil
			}
		},
		Parameters: []*registry.Parameter{},
	}

	// key() method - returns different types based on flags
	keyImpl := &registry.Function{
		Name:      "key",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			iterator, err := getFilesystemIteratorFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			if iterator.position < 0 || iterator.position >= len(iterator.entries) {
				return values.NewNull(), nil
			}

			entry := iterator.entries[iterator.position]
			keyMode := iterator.flags & KEY_MODE_MASK

			switch keyMode {
			case KEY_AS_FILENAME:
				return values.NewString(entry.Name()), nil
			default: // KEY_AS_PATHNAME (default)
				fullPath := filepath.Join(iterator.path, entry.Name())
				return values.NewString(fullPath), nil
			}
		},
		Parameters: []*registry.Parameter{},
	}

	// getFlags() method
	getFlagsImpl := &registry.Function{
		Name:      "getFlags",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			iterator, err := getFilesystemIteratorFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			return values.NewInt(int64(iterator.flags)), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// setFlags() method
	setFlagsImpl := &registry.Function{
		Name:      "setFlags",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("FilesystemIterator::setFlags() expects 1 parameter, %d given", len(args)-1)
			}

			thisObj := args[0]
			flags := int(args[1].ToInt())

			iterator, err := getFilesystemIteratorFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			iterator.flags = flags
			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "flags", Type: "int"},
		},
	}

	// Create methods map
	methods := make(map[string]*registry.MethodDescriptor)

	// Implement our own versions of DirectoryIterator methods for FilesystemIterator
	// getFilename() method
	getFilenameImpl := &registry.Function{
		Name:      "getFilename",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			iterator, err := getFilesystemIteratorFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			if iterator.position < 0 || iterator.position >= len(iterator.entries) {
				return values.NewString(""), nil
			}

			return values.NewString(iterator.entries[iterator.position].Name()), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// getPathname() method
	getPathnameImpl := &registry.Function{
		Name:      "getPathname",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			iterator, err := getFilesystemIteratorFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			if iterator.position < 0 || iterator.position >= len(iterator.entries) {
				return values.NewString(""), nil
			}

			fullPath := filepath.Join(iterator.path, iterator.entries[iterator.position].Name())
			return values.NewString(fullPath), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// getSize() method
	getSizeImpl := &registry.Function{
		Name:      "getSize",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			iterator, err := getFilesystemIteratorFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			if iterator.position < 0 || iterator.position >= len(iterator.entries) {
				return values.NewInt(0), nil
			}

			fullPath := filepath.Join(iterator.path, iterator.entries[iterator.position].Name())
			info, err := os.Stat(fullPath)
			if err != nil {
				return values.NewInt(0), nil
			}

			return values.NewInt(info.Size()), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// getType() method
	getTypeImpl := &registry.Function{
		Name:      "getType",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			iterator, err := getFilesystemIteratorFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			if iterator.position < 0 || iterator.position >= len(iterator.entries) {
				return values.NewString("unknown"), nil
			}

			entry := iterator.entries[iterator.position]
			if entry.IsDir() {
				return values.NewString("dir"), nil
			}
			return values.NewString("file"), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// getPerms() method
	getPermsImpl := &registry.Function{
		Name:      "getPerms",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			iterator, err := getFilesystemIteratorFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			if iterator.position < 0 || iterator.position >= len(iterator.entries) {
				return values.NewInt(0), nil
			}

			fullPath := filepath.Join(iterator.path, iterator.entries[iterator.position].Name())
			info, err := os.Stat(fullPath)
			if err != nil {
				return values.NewInt(0), nil
			}

			return values.NewInt(int64(info.Mode())), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// isDir() method
	isDirImpl := &registry.Function{
		Name:      "isDir",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			iterator, err := getFilesystemIteratorFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			if iterator.position < 0 || iterator.position >= len(iterator.entries) {
				return values.NewBool(false), nil
			}

			return values.NewBool(iterator.entries[iterator.position].IsDir()), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// isFile() method
	isFileImpl := &registry.Function{
		Name:      "isFile",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			iterator, err := getFilesystemIteratorFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			if iterator.position < 0 || iterator.position >= len(iterator.entries) {
				return values.NewBool(false), nil
			}

			return values.NewBool(!iterator.entries[iterator.position].IsDir()), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// isDot() method
	isDotImpl := &registry.Function{
		Name:      "isDot",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			iterator, err := getFilesystemIteratorFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			if iterator.position < 0 || iterator.position >= len(iterator.entries) {
				return values.NewBool(false), nil
			}

			name := iterator.entries[iterator.position].Name()
			return values.NewBool(name == "." || name == ".."), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// isReadable() method
	isReadableImpl := &registry.Function{
		Name:      "isReadable",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			iterator, err := getFilesystemIteratorFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			if iterator.position < 0 || iterator.position >= len(iterator.entries) {
				return values.NewBool(false), nil
			}

			fullPath := filepath.Join(iterator.path, iterator.entries[iterator.position].Name())
			_, err = os.Open(fullPath)
			if err != nil {
				return values.NewBool(false), nil
			}
			return values.NewBool(true), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// We need to implement our own iterator methods that work with FilesystemIterator
	// next() method
	nextImpl := &registry.Function{
		Name:      "next",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			iterator, err := getFilesystemIteratorFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			iterator.position++
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
			iterator, err := getFilesystemIteratorFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			iterator.position = 0
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
			iterator, err := getFilesystemIteratorFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			return values.NewBool(iterator.position >= 0 && iterator.position < len(iterator.entries)), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// Override specific methods
	methods["__construct"] = &registry.MethodDescriptor{
		Name:       "__construct",
		Visibility: "public",
		Parameters: []*registry.ParameterDescriptor{
			{Name: "path", Type: "string"},
			{Name: "flags", Type: "int"},
		},
		Implementation: NewBuiltinMethodImpl(constructorImpl),
	}
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
	methods["getFilename"] = &registry.MethodDescriptor{
		Name:           "getFilename",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(getFilenameImpl),
	}
	methods["getPathname"] = &registry.MethodDescriptor{
		Name:           "getPathname",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(getPathnameImpl),
	}
	methods["getSize"] = &registry.MethodDescriptor{
		Name:           "getSize",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(getSizeImpl),
	}
	methods["getType"] = &registry.MethodDescriptor{
		Name:           "getType",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(getTypeImpl),
	}
	methods["getPerms"] = &registry.MethodDescriptor{
		Name:           "getPerms",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(getPermsImpl),
	}
	methods["isDir"] = &registry.MethodDescriptor{
		Name:           "isDir",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(isDirImpl),
	}
	methods["isFile"] = &registry.MethodDescriptor{
		Name:           "isFile",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(isFileImpl),
	}
	methods["isDot"] = &registry.MethodDescriptor{
		Name:           "isDot",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(isDotImpl),
	}
	methods["isReadable"] = &registry.MethodDescriptor{
		Name:           "isReadable",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(isReadableImpl),
	}

	// Add constants
	constants := map[string]*registry.ConstantDescriptor{
		"CURRENT_AS_PATHNAME": {
			Name:       "CURRENT_AS_PATHNAME",
			Visibility: "public",
			Value:      values.NewInt(CURRENT_AS_PATHNAME),
			IsFinal:    true,
		},
		"CURRENT_AS_FILEINFO": {
			Name:       "CURRENT_AS_FILEINFO",
			Visibility: "public",
			Value:      values.NewInt(CURRENT_AS_FILEINFO),
			IsFinal:    true,
		},
		"CURRENT_AS_SELF": {
			Name:       "CURRENT_AS_SELF",
			Visibility: "public",
			Value:      values.NewInt(CURRENT_AS_SELF),
			IsFinal:    true,
		},
		"CURRENT_MODE_MASK": {
			Name:       "CURRENT_MODE_MASK",
			Visibility: "public",
			Value:      values.NewInt(CURRENT_MODE_MASK),
			IsFinal:    true,
		},
		"KEY_AS_PATHNAME": {
			Name:       "KEY_AS_PATHNAME",
			Visibility: "public",
			Value:      values.NewInt(KEY_AS_PATHNAME),
			IsFinal:    true,
		},
		"KEY_AS_FILENAME": {
			Name:       "KEY_AS_FILENAME",
			Visibility: "public",
			Value:      values.NewInt(KEY_AS_FILENAME),
			IsFinal:    true,
		},
		"FOLLOW_SYMLINKS": {
			Name:       "FOLLOW_SYMLINKS",
			Visibility: "public",
			Value:      values.NewInt(FOLLOW_SYMLINKS),
			IsFinal:    true,
		},
		"KEY_MODE_MASK": {
			Name:       "KEY_MODE_MASK",
			Visibility: "public",
			Value:      values.NewInt(KEY_MODE_MASK),
			IsFinal:    true,
		},
		"NEW_CURRENT_AND_KEY": {
			Name:       "NEW_CURRENT_AND_KEY",
			Visibility: "public",
			Value:      values.NewInt(NEW_CURRENT_AND_KEY),
			IsFinal:    true,
		},
		"SKIP_DOTS": {
			Name:       "SKIP_DOTS",
			Visibility: "public",
			Value:      values.NewInt(SKIP_DOTS),
			IsFinal:    true,
		},
		"UNIX_PATHS": {
			Name:       "UNIX_PATHS",
			Visibility: "public",
			Value:      values.NewInt(UNIX_PATHS),
			IsFinal:    true,
		},
	}

	return &registry.ClassDescriptor{
		Name:       "FilesystemIterator",
		Parent:     "DirectoryIterator",
		Interfaces: []string{"Iterator"},
		Traits:     []string{},
		IsAbstract: false,
		IsFinal:    false,
		Methods:    methods,
		Properties: map[string]*registry.PropertyDescriptor{},
		Constants:  constants,
	}
}

// Helper function to get the filesystem iterator from object properties
func getFilesystemIteratorFromObject(obj *values.Value) (*FilesystemIterator, error) {
	if obj.Type != values.TypeObject {
		return nil, fmt.Errorf("expected object, got %s", obj.Type)
	}

	objData := obj.Data.(*values.Object)
	iteratorValue, exists := objData.Properties["_iterator"]
	if !exists {
		return nil, fmt.Errorf("filesystem iterator not initialized")
	}

	if iteratorValue.Type != values.TypeResource {
		return nil, fmt.Errorf("invalid iterator property type")
	}

	iterator, ok := iteratorValue.Data.(*FilesystemIterator)
	if !ok {
		return nil, fmt.Errorf("invalid iterator property data")
	}

	return iterator, nil
}

// fakeFSDirEntry implements os.DirEntry for . and .. entries
type fakeFSDirEntry struct {
	name  string
	isDir bool
}

func (f *fakeFSDirEntry) Name() string {
	return f.name
}

func (f *fakeFSDirEntry) IsDir() bool {
	return f.isDir
}

func (f *fakeFSDirEntry) Type() fs.FileMode {
	if f.isDir {
		return fs.ModeDir
	}
	return 0
}

func (f *fakeFSDirEntry) Info() (fs.FileInfo, error) {
	return &fakeFSFileInfo{name: f.name, isDir: f.isDir}, nil
}

// fakeFSFileInfo implements fs.FileInfo for . and .. entries
type fakeFSFileInfo struct {
	name  string
	isDir bool
}

func (f *fakeFSFileInfo) Name() string {
	return f.name
}

func (f *fakeFSFileInfo) Size() int64 {
	return 4096 // Typical directory size
}

func (f *fakeFSFileInfo) Mode() fs.FileMode {
	if f.isDir {
		return fs.ModeDir | 0755
	}
	return 0644
}

func (f *fakeFSFileInfo) ModTime() time.Time {
	return time.Now()
}

func (f *fakeFSFileInfo) IsDir() bool {
	return f.isDir
}

func (f *fakeFSFileInfo) Sys() interface{} {
	return nil
}