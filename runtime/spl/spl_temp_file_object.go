package spl

import (
	"fmt"
	"io/ioutil"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetSplTempFileObjectClass returns the SplTempFileObject class descriptor
func GetSplTempFileObjectClass() *registry.ClassDescriptor {
	// Constructor
	constructorImpl := &registry.Function{
		Name:      "__construct",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]

			// Default max memory is 2MB
			maxMemory := int64(2 * 1024 * 1024)
			if len(args) >= 2 {
				maxMemory = args[1].ToInt()
			}

			// Create a temporary file
			tempFile, err := ioutil.TempFile("", "spl_temp_file_*")
			if err != nil {
				return nil, fmt.Errorf("SplTempFileObject::__construct(): Failed to create temporary file: %v", err)
			}

			// Read all lines (initially empty)
			var lines []string
			lines = append(lines, "") // Add empty line like SplFileObject

			// Create file object data
			fileData := &SplFileObjectData{
				file:        tempFile,
				lines:       lines,
				currentLine: 0,
				flags:       0,
				csvControl:  []string{",", "\"", "\\"},
				maxLineLen:  0,
				eof:         false,
				filePath:    tempFile.Name(),
			}

			objData := thisObj.Data.(*values.Object)
			objData.Properties["_fileData"] = &values.Value{
				Type: values.TypeResource,
				Data: fileData,
			}

			// Store max memory for reference
			objData.Properties["_maxMemory"] = values.NewInt(maxMemory)

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "maxMemory", Type: "int", DefaultValue: values.NewInt(2 * 1024 * 1024)},
		},
	}

	// Get parent methods from SplFileObject
	parentClass := GetSplFileObjectClass()
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
			{Name: "maxMemory", Type: "int"},
		},
		Implementation: NewBuiltinMethodImpl(constructorImpl),
	}

	// Copy constants from parent
	constants := make(map[string]*registry.ConstantDescriptor)
	for name, constant := range parentClass.Constants {
		constants[name] = constant
	}

	return &registry.ClassDescriptor{
		Name:       "SplTempFileObject",
		Parent:     "SplFileObject",
		Interfaces: []string{"Iterator", "RecursiveIterator", "SeekableIterator"},
		Traits:     []string{},
		IsAbstract: false,
		IsFinal:    false,
		Methods:    methods,
		Properties: map[string]*registry.PropertyDescriptor{},
		Constants:  constants,
	}
}