package runtime

import (
	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/runtime/spl"
)

// newPDOMethod creates a registry.MethodDescriptor with builtin implementation
// Following SPL pattern: $this is passed as first parameter to handler
func newPDOMethod(name string, params []registry.ParameterDescriptor, returnType string, handler registry.BuiltinImplementation) *registry.MethodDescriptor {
	// Prepend $this parameter for internal use
	fullParams := make([]registry.ParameterDescriptor, 0, len(params)+1)
	fullParams = append(fullParams, registry.ParameterDescriptor{
		Name: "this",
		Type: "object",
	})
	fullParams = append(fullParams, params...)

	return &registry.MethodDescriptor{
		Name:       name,
		Visibility: "public",
		IsStatic:   false,
		IsAbstract: false,
		IsFinal:    false,
		IsVariadic: false,
		Parameters: convertToParamPointers(params),
		Implementation: spl.NewBuiltinMethodImpl(&registry.Function{
			Name:       name,
			IsBuiltin:  true,
			Builtin:    handler,
			Parameters: convertParamDescriptors(fullParams),
		}),
	}
}

// convertParamDescriptors converts ParameterDescriptors to Parameters
func convertParamDescriptors(params []registry.ParameterDescriptor) []*registry.Parameter {
	result := make([]*registry.Parameter, len(params))
	for i, p := range params {
		result[i] = &registry.Parameter{
			Name:         p.Name,
			Type:         p.Type,
			IsReference:  p.IsReference,
			HasDefault:   p.HasDefault,
			DefaultValue: p.DefaultValue,
		}
	}
	return result
}

// convertToParamPointers converts slice of ParameterDescriptor to pointers
func convertToParamPointers(params []registry.ParameterDescriptor) []*registry.ParameterDescriptor {
	result := make([]*registry.ParameterDescriptor, len(params))
	for i := range params {
		result[i] = &params[i]
	}
	return result
}
