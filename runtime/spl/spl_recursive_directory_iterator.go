package spl

import (
	"fmt"
	"os"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetRecursiveDirectoryIteratorClass returns the RecursiveDirectoryIterator class descriptor
func GetRecursiveDirectoryIteratorClass() *registry.ClassDescriptor {
	// Constructor
	constructorImpl := &registry.Function{
		Name:      "__construct",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]

			if len(args) < 2 {
				return nil, fmt.Errorf("RecursiveDirectoryIterator::__construct() expects at least 1 parameter, 0 given")
			}

			path := args[1].ToString()
			if path == "" {
				return nil, fmt.Errorf("RecursiveDirectoryIterator::__construct(): Path cannot be empty")
			}

			// Check if path exists and is a directory
			info, err := os.Stat(path)
			if err != nil {
				if os.IsNotExist(err) {
					return nil, fmt.Errorf("RecursiveDirectoryIterator::__construct(%s): Failed to open directory: No such file or directory", path)
				}
				return nil, fmt.Errorf("RecursiveDirectoryIterator::__construct(%s): Failed to open directory: %v", path, err)
			}

			if !info.IsDir() {
				return nil, fmt.Errorf("RecursiveDirectoryIterator::__construct(%s): Failed to open directory: Not a directory", path)
			}

			// Default flags
			flags := int64(0)
			if len(args) >= 3 {
				flags = args[2].ToInt()
			}

			// Read directory entries
			file, err := os.Open(path)
			if err != nil {
				return nil, fmt.Errorf("RecursiveDirectoryIterator::__construct(%s): Failed to open directory: %v", path, err)
			}
			defer file.Close()

			entries, err := file.Readdir(-1)
			if err != nil {
				return nil, fmt.Errorf("RecursiveDirectoryIterator::__construct(%s): Failed to read directory: %v", path, err)
			}

			// Create iterator data compatible with FilesystemIterator
			var allEntries []os.FileInfo

			// Add . and .. entries if SKIP_DOTS is not set
			if (flags & 4096) == 0 { // SKIP_DOTS = 4096
				dotEntry := &fakeFileInfo{name: ".", isDir: true}
				dotDotEntry := &fakeFileInfo{name: "..", isDir: true}
				allEntries = make([]os.FileInfo, 0, len(entries)+2)
				allEntries = append(allEntries, dotEntry, dotDotEntry)
				allEntries = append(allEntries, entries...)
			} else {
				allEntries = entries
			}

			iteratorData := &RecursiveDirectoryIteratorData{
				path:       path,
				entries:    allEntries,
				currentIdx: 0,
				flags:      flags,
			}

			objData := thisObj.Data.(*values.Object)
			objData.Properties["_iteratorData"] = &values.Value{
				Type: values.TypeResource,
				Data: iteratorData,
			}

			// Also store in _iterator format for compatibility with FilesystemIterator methods
			// We need to check what the actual FilesystemIterator struct looks like
			// For now, let's not store the _iterator property to avoid conflicts
			// The parent class methods will be handled by our own overrides

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "path", Type: "string"},
			{Name: "flags", Type: "int", DefaultValue: values.NewInt(0)},
		},
	}

	// Get parent methods from FilesystemIterator
	parentClass := GetFilesystemIteratorClass()
	methods := make(map[string]*registry.MethodDescriptor)

	// Copy all parent methods
	for name, method := range parentClass.Methods {
		if name != "__construct" {
			methods[name] = method
		}
	}

	// Override constructor
	methods["__construct"] = &registry.MethodDescriptor{
		Name:       "__construct",
		Visibility: "public",
		Parameters: []*registry.ParameterDescriptor{
			{Name: "path", Type: "string"},
			{Name: "flags", Type: "int"},
		},
		Implementation: NewBuiltinMethodImpl(constructorImpl),
	}

	// hasChildren method - implements RecursiveIterator
	hasChildrenImpl := &registry.Function{
		Name:      "hasChildren",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			objData := thisObj.Data.(*values.Object)

			iteratorData := objData.Properties["_iteratorData"].Data.(*RecursiveDirectoryIteratorData)

			if iteratorData.currentIdx < 0 || iteratorData.currentIdx >= len(iteratorData.entries) {
				return values.NewBool(false), nil
			}

			currentEntry := iteratorData.entries[iteratorData.currentIdx]

			// Only directories can have children, and skip '.' and '..' unless specifically requested
			if !currentEntry.IsDir() {
				return values.NewBool(false), nil
			}

			// Special handling for '.' and '..' - they never have children
			if currentEntry.Name() == "." || currentEntry.Name() == ".." {
				return values.NewBool(false), nil
			}

			return values.NewBool(true), nil
		},
	}

	// getChildren method - implements RecursiveIterator
	getChildrenImpl := &registry.Function{
		Name:      "getChildren",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			objData := thisObj.Data.(*values.Object)

			iteratorData := objData.Properties["_iteratorData"].Data.(*RecursiveDirectoryIteratorData)

			if iteratorData.currentIdx < 0 || iteratorData.currentIdx >= len(iteratorData.entries) {
				return nil, fmt.Errorf("RecursiveDirectoryIterator::getChildren(): Cannot get children - invalid position")
			}

			currentEntry := iteratorData.entries[iteratorData.currentIdx]

			if !currentEntry.IsDir() {
				return nil, fmt.Errorf("RecursiveDirectoryIterator::getChildren(): Cannot get children of non-directory")
			}

			// Build path to child directory
			childPath := iteratorData.path + "/" + currentEntry.Name()

			// Create new RecursiveDirectoryIterator for the child directory
			childObj := &values.Object{
				ClassName:  "RecursiveDirectoryIterator",
				Properties: make(map[string]*values.Value),
			}
			childThis := &values.Value{
				Type: values.TypeObject,
				Data: childObj,
			}

			// Initialize child iterator
			_, err := constructorImpl.Builtin(ctx, []*values.Value{childThis, values.NewString(childPath), values.NewInt(iteratorData.flags)})
			if err != nil {
				return nil, err
			}

			return childThis, nil
		},
	}

	// Override current method to return SplFileInfo objects
	currentImpl := &registry.Function{
		Name:      "current",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			objData := thisObj.Data.(*values.Object)

			iteratorData := objData.Properties["_iteratorData"].Data.(*RecursiveDirectoryIteratorData)

			if iteratorData.currentIdx < 0 || iteratorData.currentIdx >= len(iteratorData.entries) {
				return values.NewNull(), nil
			}

			currentEntry := iteratorData.entries[iteratorData.currentIdx]
			currentPath := iteratorData.path + "/" + currentEntry.Name()

			// Check flag to determine return type
			flags := iteratorData.flags

			// CURRENT_AS_PATHNAME flag (from FilesystemIterator)
			if flags&32 != 0 { // CURRENT_AS_PATHNAME = 32
				return values.NewString(currentPath), nil
			}

			// Default: return SplFileInfo object
			fileInfoObj := &values.Object{
				ClassName:  "SplFileInfo",
				Properties: make(map[string]*values.Value),
			}
			fileInfoObj.Properties["__filepath"] = values.NewString(currentPath)

			return &values.Value{
				Type: values.TypeObject,
				Data: fileInfoObj,
			}, nil
		},
	}

	// Override key method to handle flags
	keyImpl := &registry.Function{
		Name:      "key",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			objData := thisObj.Data.(*values.Object)

			iteratorData := objData.Properties["_iteratorData"].Data.(*RecursiveDirectoryIteratorData)

			if iteratorData.currentIdx < 0 || iteratorData.currentIdx >= len(iteratorData.entries) {
				return values.NewString(""), nil
			}

			currentEntry := iteratorData.entries[iteratorData.currentIdx]
			flags := iteratorData.flags

			// KEY_AS_FILENAME flag (from FilesystemIterator)
			if flags&16 != 0 { // KEY_AS_FILENAME = 16
				return values.NewString(currentEntry.Name()), nil
			}

			// Default: return full path
			currentPath := iteratorData.path + "/" + currentEntry.Name()
			return values.NewString(currentPath), nil
		},
	}

	// Override valid method to handle SKIP_DOTS flag
	validImpl := &registry.Function{
		Name:      "valid",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			objData := thisObj.Data.(*values.Object)

			iteratorData := objData.Properties["_iteratorData"].Data.(*RecursiveDirectoryIteratorData)

			// Skip invalid positions
			for iteratorData.currentIdx >= 0 && iteratorData.currentIdx < len(iteratorData.entries) {
				currentEntry := iteratorData.entries[iteratorData.currentIdx]

				// Check SKIP_DOTS flag
				if iteratorData.flags&4096 != 0 { // SKIP_DOTS = 4096
					if currentEntry.Name() == "." || currentEntry.Name() == ".." {
						iteratorData.currentIdx++
						continue
					}
				}

				return values.NewBool(true), nil
			}

			return values.NewBool(false), nil
		},
	}

	// Override next method to handle SKIP_DOTS flag
	nextImpl := &registry.Function{
		Name:      "next",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			objData := thisObj.Data.(*values.Object)

			iteratorData := objData.Properties["_iteratorData"].Data.(*RecursiveDirectoryIteratorData)
			iteratorData.currentIdx++

			// If SKIP_DOTS is set, skip dot entries automatically
			if iteratorData.flags&4096 != 0 { // SKIP_DOTS = 4096
				for iteratorData.currentIdx < len(iteratorData.entries) {
					currentEntry := iteratorData.entries[iteratorData.currentIdx]
					if currentEntry.Name() != "." && currentEntry.Name() != ".." {
						break
					}
					iteratorData.currentIdx++
				}
			}

			return values.NewNull(), nil
		},
	}

	// Override rewind method to handle SKIP_DOTS flag
	rewindImpl := &registry.Function{
		Name:      "rewind",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			objData := thisObj.Data.(*values.Object)

			iteratorData := objData.Properties["_iteratorData"].Data.(*RecursiveDirectoryIteratorData)
			iteratorData.currentIdx = 0

			// If SKIP_DOTS is set, skip initial dot entries
			if iteratorData.flags&4096 != 0 { // SKIP_DOTS = 4096
				for iteratorData.currentIdx < len(iteratorData.entries) {
					currentEntry := iteratorData.entries[iteratorData.currentIdx]
					if currentEntry.Name() != "." && currentEntry.Name() != ".." {
						break
					}
					iteratorData.currentIdx++
				}
			}

			return values.NewNull(), nil
		},
	}

	// Add new methods
	methods["hasChildren"] = &registry.MethodDescriptor{
		Name:           "hasChildren",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(hasChildrenImpl),
	}

	methods["getChildren"] = &registry.MethodDescriptor{
		Name:           "getChildren",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(getChildrenImpl),
	}

	// Override getFlags method to work with our data structure
	getFlagsImpl := &registry.Function{
		Name:      "getFlags",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			objData := thisObj.Data.(*values.Object)

			iteratorData := objData.Properties["_iteratorData"].Data.(*RecursiveDirectoryIteratorData)
			return values.NewInt(iteratorData.flags), nil
		},
	}

	// Override setFlags method to work with our data structure
	setFlagsImpl := &registry.Function{
		Name:      "setFlags",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("RecursiveDirectoryIterator::setFlags() expects 1 parameter, %d given", len(args)-1)
			}

			thisObj := args[0]
			objData := thisObj.Data.(*values.Object)

			iteratorData := objData.Properties["_iteratorData"].Data.(*RecursiveDirectoryIteratorData)
			iteratorData.flags = args[1].ToInt()

			return values.NewNull(), nil
		},
	}

	// Override methods for proper flag handling
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

	methods["valid"] = &registry.MethodDescriptor{
		Name:           "valid",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(validImpl),
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

	// Copy constants from parent and add RecursiveDirectoryIterator specific ones
	constants := make(map[string]*registry.ConstantDescriptor)
	for name, constant := range parentClass.Constants {
		constants[name] = constant
	}

	return &registry.ClassDescriptor{
		Name:       "RecursiveDirectoryIterator",
		Parent:     "FilesystemIterator",
		Interfaces: []string{"Iterator", "RecursiveIterator", "SeekableIterator"},
		Traits:     []string{},
		IsAbstract: false,
		IsFinal:    false,
		Methods:    methods,
		Properties: map[string]*registry.PropertyDescriptor{},
		Constants:  constants,
	}
}

// RecursiveDirectoryIteratorData holds the iterator state
type RecursiveDirectoryIteratorData struct {
	path       string
	entries    []os.FileInfo
	currentIdx int
	flags      int64
}

// Note: fakeFileInfo is already declared in spl_directory_iterator.go