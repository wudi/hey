package spl

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetGlobIteratorClass returns the GlobIterator class descriptor
func GetGlobIteratorClass() *registry.ClassDescriptor {
	// Constructor
	constructorImpl := &registry.Function{
		Name:      "__construct",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]

			if len(args) < 2 {
				return nil, fmt.Errorf("GlobIterator::__construct() expects at least 1 parameter, 0 given")
			}

			pattern := args[1].ToString()
			if pattern == "" {
				return nil, fmt.Errorf("GlobIterator::__construct(): Argument #1 ($pattern) must not be empty")
			}

			// Default flags
			flags := int64(0)
			if len(args) >= 3 {
				flags = args[2].ToInt()
			}

			// Use Go's filepath.Glob to find matching files
			matches, err := filepath.Glob(pattern)
			if err != nil {
				return nil, fmt.Errorf("GlobIterator::__construct(): Invalid glob pattern: %v", err)
			}

			// Sort matches to ensure consistent ordering
			sort.Strings(matches)

			// Store the matches - we'll create SplFileInfo objects on demand

			// Create iterator data
			iteratorData := &GlobIteratorData{
				pattern:    pattern,
				matches:    matches,
				currentIdx: 0,
				flags:      flags,
			}

			objData := thisObj.Data.(*values.Object)
			objData.Properties["_iteratorData"] = &values.Value{
				Type: values.TypeResource,
				Data: iteratorData,
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "pattern", Type: "string"},
			{Name: "flags", Type: "int", DefaultValue: values.NewInt(0)},
		},
	}

	// count method
	countImpl := &registry.Function{
		Name:      "count",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			objData := thisObj.Data.(*values.Object)

			iteratorData := objData.Properties["_iteratorData"].Data.(*GlobIteratorData)
			return values.NewInt(int64(len(iteratorData.matches))), nil
		},
	}

	// current method - returns current SplFileInfo object
	currentImpl := &registry.Function{
		Name:      "current",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			objData := thisObj.Data.(*values.Object)

			iteratorData := objData.Properties["_iteratorData"].Data.(*GlobIteratorData)

			if iteratorData.currentIdx < 0 || iteratorData.currentIdx >= len(iteratorData.matches) {
				return values.NewNull(), nil
			}

			currentPath := iteratorData.matches[iteratorData.currentIdx]

			// Create SplFileInfo object
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

	// key method - returns current filename path
	keyImpl := &registry.Function{
		Name:      "key",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			objData := thisObj.Data.(*values.Object)

			iteratorData := objData.Properties["_iteratorData"].Data.(*GlobIteratorData)

			if iteratorData.currentIdx < 0 || iteratorData.currentIdx >= len(iteratorData.matches) {
				return values.NewNull(), nil
			}

			return values.NewString(iteratorData.matches[iteratorData.currentIdx]), nil
		},
	}

	// next method
	nextImpl := &registry.Function{
		Name:      "next",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			objData := thisObj.Data.(*values.Object)

			iteratorData := objData.Properties["_iteratorData"].Data.(*GlobIteratorData)
			iteratorData.currentIdx++

			return values.NewNull(), nil
		},
	}

	// rewind method
	rewindImpl := &registry.Function{
		Name:      "rewind",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			objData := thisObj.Data.(*values.Object)

			iteratorData := objData.Properties["_iteratorData"].Data.(*GlobIteratorData)
			iteratorData.currentIdx = 0

			return values.NewNull(), nil
		},
	}

	// valid method
	validImpl := &registry.Function{
		Name:      "valid",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			objData := thisObj.Data.(*values.Object)

			iteratorData := objData.Properties["_iteratorData"].Data.(*GlobIteratorData)

			valid := iteratorData.currentIdx >= 0 && iteratorData.currentIdx < len(iteratorData.matches)
			return values.NewBool(valid), nil
		},
	}

	// getFlags method (inherited from FilesystemIterator)
	getFlagsImpl := &registry.Function{
		Name:      "getFlags",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			objData := thisObj.Data.(*values.Object)

			iteratorData := objData.Properties["_iteratorData"].Data.(*GlobIteratorData)
			return values.NewInt(iteratorData.flags), nil
		},
	}

	// setFlags method (inherited from FilesystemIterator)
	setFlagsImpl := &registry.Function{
		Name:      "setFlags",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			objData := thisObj.Data.(*values.Object)

			if len(args) < 2 {
				return nil, fmt.Errorf("GlobIterator::setFlags() expects exactly 1 parameter, 0 given")
			}

			iteratorData := objData.Properties["_iteratorData"].Data.(*GlobIteratorData)
			iteratorData.flags = args[1].ToInt()

			return values.NewNull(), nil
		},
	}

	methods := map[string]*registry.MethodDescriptor{
		"__construct": {
			Name:           "__construct",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{
				{Name: "pattern", Type: "string"},
				{Name: "flags", Type: "int"},
			},
			Implementation: NewBuiltinMethodImpl(constructorImpl),
		},
		"count": {
			Name:           "count",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(countImpl),
		},
		"current": {
			Name:           "current",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(currentImpl),
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
		"getFlags": {
			Name:           "getFlags",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(getFlagsImpl),
		},
		"setFlags": {
			Name:           "setFlags",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{
				{Name: "flags", Type: "int"},
			},
			Implementation: NewBuiltinMethodImpl(setFlagsImpl),
		},
	}

	// Get FilesystemIterator constants
	filesystemIteratorClass := GetFilesystemIteratorClass()
	constants := make(map[string]*registry.ConstantDescriptor)
	for name, constant := range filesystemIteratorClass.Constants {
		constants[name] = constant
	}

	return &registry.ClassDescriptor{
		Name:       "GlobIterator",
		Parent:     "FilesystemIterator",
		Interfaces: []string{"Iterator", "SeekableIterator", "Countable"},
		Traits:     []string{},
		IsAbstract: false,
		IsFinal:    false,
		Methods:    methods,
		Properties: map[string]*registry.PropertyDescriptor{},
		Constants:  constants,
	}
}

// GlobIteratorData holds the iterator state
type GlobIteratorData struct {
	pattern    string
	matches    []string
	currentIdx int
	flags      int64
}