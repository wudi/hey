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

// DirectoryIterator represents a directory iterator
type DirectoryIterator struct {
	path    string
	entries []os.DirEntry
	position int
}

// GetDirectoryIteratorClass returns the DirectoryIterator class descriptor
func GetDirectoryIteratorClass() *registry.ClassDescriptor {
	// Constructor
	constructorImpl := &registry.Function{
		Name:      "__construct",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("DirectoryIterator::__construct() expects 1 parameter, %d given", len(args)-1)
			}

			thisObj := args[0]
			pathArg := args[1]
			path := pathArg.ToString()

			// Check if directory exists
			info, err := os.Stat(path)
			if err != nil {
				return nil, fmt.Errorf("DirectoryIterator::__construct(%s): Failed to open directory: %v", path, err)
			}
			if !info.IsDir() {
				return nil, fmt.Errorf("DirectoryIterator::__construct(%s): Not a directory", path)
			}

			// Read directory entries
			file, err := os.Open(path)
			if err != nil {
				return nil, fmt.Errorf("DirectoryIterator::__construct(%s): Failed to open directory: %v", path, err)
			}
			defer file.Close()

			entries, err := file.ReadDir(-1)
			if err != nil {
				return nil, fmt.Errorf("DirectoryIterator::__construct(%s): Failed to read directory: %v", path, err)
			}

			// Add . and .. entries like PHP does
			dotEntry := &fakeDirEntry{name: ".", isDir: true}
			dotDotEntry := &fakeDirEntry{name: "..", isDir: true}

			allEntries := make([]os.DirEntry, 0, len(entries)+2)
			allEntries = append(allEntries, dotEntry, dotDotEntry)
			allEntries = append(allEntries, entries...)

			// Create iterator
			iterator := &DirectoryIterator{
				path:     path,
				entries:  allEntries,
				position: 0,
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
		},
	}

	// current() method
	currentImpl := &registry.Function{
		Name:      "current",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			iterator, err := getDirectoryIteratorFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			if iterator.position < 0 || iterator.position >= len(iterator.entries) {
				return values.NewNull(), nil
			}

			// Return the current DirectoryIterator object itself (like PHP)
			return thisObj, nil
		},
		Parameters: []*registry.Parameter{},
	}

	// getFilename() method
	getFilenameImpl := &registry.Function{
		Name:      "getFilename",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			iterator, err := getDirectoryIteratorFromObject(thisObj)
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
			iterator, err := getDirectoryIteratorFromObject(thisObj)
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
			iterator, err := getDirectoryIteratorFromObject(thisObj)
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
			iterator, err := getDirectoryIteratorFromObject(thisObj)
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
			iterator, err := getDirectoryIteratorFromObject(thisObj)
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
			iterator, err := getDirectoryIteratorFromObject(thisObj)
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
			iterator, err := getDirectoryIteratorFromObject(thisObj)
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
			iterator, err := getDirectoryIteratorFromObject(thisObj)
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
			iterator, err := getDirectoryIteratorFromObject(thisObj)
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

	// key() method
	keyImpl := &registry.Function{
		Name:      "key",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			iterator, err := getDirectoryIteratorFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			return values.NewInt(int64(iterator.position)), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// next() method
	nextImpl := &registry.Function{
		Name:      "next",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			iterator, err := getDirectoryIteratorFromObject(thisObj)
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
			iterator, err := getDirectoryIteratorFromObject(thisObj)
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
			iterator, err := getDirectoryIteratorFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			return values.NewBool(iterator.position >= 0 && iterator.position < len(iterator.entries)), nil
		},
		Parameters: []*registry.Parameter{},
	}

	methods := map[string]*registry.MethodDescriptor{
		"__construct": {
			Name:       "__construct",
			Visibility: "public",
			Parameters: []*registry.ParameterDescriptor{
				{Name: "path", Type: "string"},
			},
			Implementation: NewBuiltinMethodImpl(constructorImpl),
		},
		"current": {
			Name:           "current",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(currentImpl),
		},
		"getFilename": {
			Name:           "getFilename",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(getFilenameImpl),
		},
		"getPathname": {
			Name:           "getPathname",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(getPathnameImpl),
		},
		"getSize": {
			Name:           "getSize",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(getSizeImpl),
		},
		"getType": {
			Name:           "getType",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(getTypeImpl),
		},
		"getPerms": {
			Name:           "getPerms",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(getPermsImpl),
		},
		"isDir": {
			Name:           "isDir",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(isDirImpl),
		},
		"isFile": {
			Name:           "isFile",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(isFileImpl),
		},
		"isDot": {
			Name:           "isDot",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(isDotImpl),
		},
		"isReadable": {
			Name:           "isReadable",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(isReadableImpl),
		},
		"key": {
			Name:           "key",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(keyImpl),
		},
		"next": {
			Name:           "next",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(nextImpl),
		},
		"rewind": {
			Name:           "rewind",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(rewindImpl),
		},
		"valid": {
			Name:           "valid",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(validImpl),
		},
	}

	return &registry.ClassDescriptor{
		Name:       "DirectoryIterator",
		Parent:     "SplFileInfo",
		Interfaces: []string{"Iterator"},
		Traits:     []string{},
		IsAbstract: false,
		IsFinal:    false,
		Methods:    methods,
		Properties: map[string]*registry.PropertyDescriptor{},
		Constants:  map[string]*registry.ConstantDescriptor{},
	}
}

// Helper function to get the directory iterator from object properties
func getDirectoryIteratorFromObject(obj *values.Value) (*DirectoryIterator, error) {
	if obj.Type != values.TypeObject {
		return nil, fmt.Errorf("expected object, got %s", obj.Type)
	}

	objData := obj.Data.(*values.Object)
	iteratorValue, exists := objData.Properties["_iterator"]
	if !exists {
		return nil, fmt.Errorf("directory iterator not initialized")
	}

	if iteratorValue.Type != values.TypeResource {
		return nil, fmt.Errorf("invalid iterator property type")
	}

	iterator, ok := iteratorValue.Data.(*DirectoryIterator)
	if !ok {
		return nil, fmt.Errorf("invalid iterator property data")
	}

	return iterator, nil
}

// fakeDirEntry implements os.DirEntry for . and .. entries
type fakeDirEntry struct {
	name  string
	isDir bool
}

func (f *fakeDirEntry) Name() string {
	return f.name
}

func (f *fakeDirEntry) IsDir() bool {
	return f.isDir
}

func (f *fakeDirEntry) Type() fs.FileMode {
	if f.isDir {
		return fs.ModeDir
	}
	return 0
}

func (f *fakeDirEntry) Info() (fs.FileInfo, error) {
	return &fakeFileInfo{name: f.name, isDir: f.isDir}, nil
}

// fakeFileInfo implements fs.FileInfo for . and .. entries
type fakeFileInfo struct {
	name  string
	isDir bool
}

func (f *fakeFileInfo) Name() string {
	return f.name
}

func (f *fakeFileInfo) Size() int64 {
	return 4096 // Typical directory size
}

func (f *fakeFileInfo) Mode() fs.FileMode {
	if f.isDir {
		return fs.ModeDir | 0755
	}
	return 0644
}

func (f *fakeFileInfo) ModTime() time.Time {
	return time.Now()
}

func (f *fakeFileInfo) IsDir() bool {
	return f.isDir
}

func (f *fakeFileInfo) Sys() interface{} {
	return nil
}