package spl

import (
	"fmt"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// ObjectStorageEntry represents an entry in the object storage
type ObjectStorageEntry struct {
	Object *values.Value
	Data   *values.Value
}

// GetSplObjectStorageClass returns the SplObjectStorage class descriptor
func GetSplObjectStorageClass() *registry.ClassDescriptor {
	// Constructor
	constructorImpl := &registry.Function{
		Name:      "__construct",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 {
				return values.NewNull(), fmt.Errorf("SplObjectStorage::__construct() expects at least 1 argument, %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("__construct called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if obj.Properties == nil {
				obj.Properties = make(map[string]*values.Value)
			}

			// Initialize storage as empty slice
			storage := []*ObjectStorageEntry{}
			obj.Properties["__storage"] = &values.Value{
				Type: values.TypeArray,
				Data: storage,
			}
			obj.Properties["__position"] = values.NewInt(0)

			return thisObj, nil
		},
		Parameters: []*registry.Parameter{},
	}

	// attach() method
	attachImpl := &registry.Function{
		Name:      "attach",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 2 {
				return values.NewNull(), fmt.Errorf("attach expects at least 2 arguments")
			}
			if !args[0].IsObject() {
				return values.NewNull(), fmt.Errorf("attach called on non-object")
			}
			if !args[1].IsObject() {
				return values.NewNull(), fmt.Errorf("attach expects first parameter to be an object")
			}

			thisObj := args[0]
			objectToAttach := args[1]
			var data *values.Value = values.NewNull()
			if len(args) > 2 {
				data = args[2]
			}

			obj := thisObj.Data.(*values.Object)
			if obj.Properties == nil {
				obj.Properties = make(map[string]*values.Value)
			}

			// Get storage
			storageVal, ok := obj.Properties["__storage"]
			if !ok || storageVal == nil {
				return values.NewNull(), fmt.Errorf("storage not initialized")
			}

			storage, ok := storageVal.Data.([]*ObjectStorageEntry)
			if !ok {
				return values.NewNull(), fmt.Errorf("invalid storage type")
			}

			// Check if object already exists
			for i, entry := range storage {
				if entry.Object == objectToAttach {
					// Update existing entry
					storage[i].Data = data
					return values.NewNull(), nil
				}
			}

			// Add new entry
			entry := &ObjectStorageEntry{
				Object: objectToAttach,
				Data:   data,
			}
			storage = append(storage, entry)

			// Update storage
			obj.Properties["__storage"] = &values.Value{
				Type: values.TypeArray,
				Data: storage,
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "object", Type: "object"},
			{Name: "info", Type: "mixed", HasDefault: true, DefaultValue: values.NewNull()},
		},
	}

	// contains() method
	containsImpl := &registry.Function{
		Name:      "contains",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 2 {
				return values.NewBool(false), nil
			}
			if !args[0].IsObject() {
				return values.NewBool(false), nil
			}
			if !args[1].IsObject() {
				return values.NewBool(false), nil
			}

			thisObj := args[0]
			objectToCheck := args[1]

			obj := thisObj.Data.(*values.Object)
			storageVal, ok := obj.Properties["__storage"]
			if !ok || storageVal == nil {
				return values.NewBool(false), nil
			}

			storage, ok := storageVal.Data.([]*ObjectStorageEntry)
			if !ok {
				return values.NewBool(false), nil
			}

			// Search for object
			for i, entry := range storage {
				if entry.Object == objectToCheck {
					// Set current position to found object
					obj.Properties["__position"] = values.NewInt(int64(i))
					return values.NewBool(true), nil
				}
			}

			return values.NewBool(false), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "object", Type: "object"},
		},
	}

	// detach() method
	detachImpl := &registry.Function{
		Name:      "detach",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 2 {
				return values.NewNull(), fmt.Errorf("detach expects 2 arguments")
			}
			if !args[0].IsObject() {
				return values.NewNull(), fmt.Errorf("detach called on non-object")
			}
			if !args[1].IsObject() {
				return values.NewNull(), fmt.Errorf("detach expects first parameter to be an object")
			}

			thisObj := args[0]
			objectToDetach := args[1]

			obj := thisObj.Data.(*values.Object)
			storageVal, ok := obj.Properties["__storage"]
			if !ok || storageVal == nil {
				return values.NewNull(), nil
			}

			storage, ok := storageVal.Data.([]*ObjectStorageEntry)
			if !ok {
				return values.NewNull(), nil
			}

			// Find and remove object
			for i, entry := range storage {
				if entry.Object == objectToDetach {
					// Remove entry
					storage = append(storage[:i], storage[i+1:]...)
					obj.Properties["__storage"] = &values.Value{
						Type: values.TypeArray,
						Data: storage,
					}
					break
				}
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "object", Type: "object"},
		},
	}

	// count() method
	countImpl := &registry.Function{
		Name:      "count",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewInt(0), nil
			}

			obj := args[0].Data.(*values.Object)
			storageVal, ok := obj.Properties["__storage"]
			if !ok || storageVal == nil {
				return values.NewInt(0), nil
			}

			storage, ok := storageVal.Data.([]*ObjectStorageEntry)
			if !ok {
				return values.NewInt(0), nil
			}

			return values.NewInt(int64(len(storage))), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// getInfo() method
	getInfoImpl := &registry.Function{
		Name:      "getInfo",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewNull(), nil
			}

			obj := args[0].Data.(*values.Object)
			storageVal, ok := obj.Properties["__storage"]
			if !ok || storageVal == nil {
				return values.NewNull(), nil
			}

			storage, ok := storageVal.Data.([]*ObjectStorageEntry)
			if !ok {
				return values.NewNull(), nil
			}

			// Get current position
			position := int64(0)
			if pos, ok := obj.Properties["__position"]; ok && pos != nil {
				position = pos.ToInt()
			}

			if position >= 0 && int(position) < len(storage) {
				return storage[position].Data, nil
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// setInfo() method
	setInfoImpl := &registry.Function{
		Name:      "setInfo",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 2 {
				return values.NewNull(), fmt.Errorf("setInfo expects 2 arguments")
			}
			if !args[0].IsObject() {
				return values.NewNull(), fmt.Errorf("setInfo called on non-object")
			}

			thisObj := args[0]
			newData := args[1]

			obj := thisObj.Data.(*values.Object)
			storageVal, ok := obj.Properties["__storage"]
			if !ok || storageVal == nil {
				return values.NewNull(), nil
			}

			storage, ok := storageVal.Data.([]*ObjectStorageEntry)
			if !ok {
				return values.NewNull(), nil
			}

			// Get current position
			position := int64(0)
			if pos, ok := obj.Properties["__position"]; ok && pos != nil {
				position = pos.ToInt()
			}

			if position >= 0 && int(position) < len(storage) {
				storage[position].Data = newData
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "info", Type: "mixed"},
		},
	}

	// Iterator methods

	// current() method
	currentImpl := &registry.Function{
		Name:      "current",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewNull(), nil
			}

			obj := args[0].Data.(*values.Object)
			storageVal, ok := obj.Properties["__storage"]
			if !ok || storageVal == nil {
				return values.NewNull(), nil
			}

			storage, ok := storageVal.Data.([]*ObjectStorageEntry)
			if !ok {
				return values.NewNull(), nil
			}

			// Get current position
			position := int64(0)
			if pos, ok := obj.Properties["__position"]; ok && pos != nil {
				position = pos.ToInt()
			}

			if position >= 0 && int(position) < len(storage) {
				return storage[position].Object, nil
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// key() method
	keyImpl := &registry.Function{
		Name:      "key",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewNull(), nil
			}

			obj := args[0].Data.(*values.Object)
			if pos, ok := obj.Properties["__position"]; ok && pos != nil {
				return pos, nil
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// next() method
	nextImpl := &registry.Function{
		Name:      "next",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewNull(), nil
			}

			obj := args[0].Data.(*values.Object)
			position := int64(0)
			if pos, ok := obj.Properties["__position"]; ok && pos != nil {
				position = pos.ToInt()
			}

			obj.Properties["__position"] = values.NewInt(position + 1)
			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// rewind() method
	rewindImpl := &registry.Function{
		Name:      "rewind",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewNull(), nil
			}

			obj := args[0].Data.(*values.Object)
			obj.Properties["__position"] = values.NewInt(0)
			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// valid() method
	validImpl := &registry.Function{
		Name:      "valid",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewBool(false), nil
			}

			obj := args[0].Data.(*values.Object)
			storageVal, ok := obj.Properties["__storage"]
			if !ok || storageVal == nil {
				return values.NewBool(false), nil
			}

			storage, ok := storageVal.Data.([]*ObjectStorageEntry)
			if !ok {
				return values.NewBool(false), nil
			}

			// Get current position
			position := int64(0)
			if pos, ok := obj.Properties["__position"]; ok && pos != nil {
				position = pos.ToInt()
			}

			return values.NewBool(position >= 0 && int(position) < len(storage)), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// Create method descriptors
	methods := map[string]*registry.MethodDescriptor{
		"__construct": {
			Name:           "__construct",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(constructorImpl),
		},
		"attach": {
			Name:       "attach",
			Visibility: "public",
			Parameters: []*registry.ParameterDescriptor{
				{Name: "object", Type: "object"},
				{Name: "info", Type: "mixed", HasDefault: true, DefaultValue: values.NewNull()},
			},
			Implementation: NewBuiltinMethodImpl(attachImpl),
		},
		"contains": {
			Name:       "contains",
			Visibility: "public",
			Parameters: []*registry.ParameterDescriptor{
				{Name: "object", Type: "object"},
			},
			Implementation: NewBuiltinMethodImpl(containsImpl),
		},
		"detach": {
			Name:       "detach",
			Visibility: "public",
			Parameters: []*registry.ParameterDescriptor{
				{Name: "object", Type: "object"},
			},
			Implementation: NewBuiltinMethodImpl(detachImpl),
		},
		"count": {
			Name:           "count",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(countImpl),
		},
		"getInfo": {
			Name:           "getInfo",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(getInfoImpl),
		},
		"setInfo": {
			Name:       "setInfo",
			Visibility: "public",
			Parameters: []*registry.ParameterDescriptor{
				{Name: "info", Type: "mixed"},
			},
			Implementation: NewBuiltinMethodImpl(setInfoImpl),
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
	}

	return &registry.ClassDescriptor{
		Name:       "SplObjectStorage",
		Parent:     "",
		Interfaces: []string{"Iterator", "Countable"},
		Properties: make(map[string]*registry.PropertyDescriptor),
		Methods:    methods,
		Constants:  make(map[string]*registry.ConstantDescriptor),
	}
}