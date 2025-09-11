package compiler

import (
	"github.com/wudi/hey/compiler/vm"
)

// GetVMFunctions converts CompilerFunction map to vm.Function map for VM compatibility
func (c *Compiler) GetVMFunctions() map[string]*vm.Function {
	vmFunctions := make(map[string]*vm.Function)
	for name, compilerFunc := range c.functions {
		vmFunc := &vm.Function{
			Name:         compilerFunc.Name,
			Instructions: compilerFunc.Instructions,
			Constants:    compilerFunc.Constants,
			Parameters:   convertToVMParameters(compilerFunc.Parameters),
			IsVariadic:   compilerFunc.IsVariadic,
			IsGenerator:  compilerFunc.IsGenerator,
		}
		vmFunctions[name] = vmFunc
	}
	return vmFunctions
}

// GetVMClasses converts CompilerClass map to vm.Class map for VM compatibility
func (c *Compiler) GetVMClasses() map[string]*vm.Class {
	vmClasses := make(map[string]*vm.Class)
	for name, compilerClass := range c.classes {
		vmClass := &vm.Class{
			Name:       compilerClass.Name,
			Parent:     compilerClass.Parent,
			Properties: convertToVMProperties(compilerClass.Properties),
			Methods:    convertToVMFunctions(compilerClass.Methods),
			Constants:  convertToVMClassConstants(compilerClass.Constants),
			IsAbstract: compilerClass.IsAbstract,
			IsFinal:    compilerClass.IsFinal,
		}
		vmClasses[name] = vmClass
	}
	return vmClasses
}

// Helper conversion functions
func convertToVMParameters(compilerParams []CompilerParameter) []vm.Parameter {
	vmParams := make([]vm.Parameter, len(compilerParams))
	for i, param := range compilerParams {
		vmParams[i] = vm.Parameter{
			Name:         param.Name,
			Type:         param.Type,
			IsReference:  param.IsReference,
			HasDefault:   param.HasDefault,
			DefaultValue: param.DefaultValue,
		}
	}
	return vmParams
}

func convertToVMProperties(compilerProps map[string]*CompilerProperty) map[string]*vm.Property {
	vmProps := make(map[string]*vm.Property)
	for name, prop := range compilerProps {
		vmProps[name] = &vm.Property{
			Name:         prop.Name,
			Type:         prop.Type,
			Visibility:   prop.Visibility,
			IsStatic:     prop.IsStatic,
			DefaultValue: prop.DefaultValue,
		}
	}
	return vmProps
}

func convertToVMFunctions(compilerFuncs map[string]*CompilerFunction) map[string]*vm.Function {
	vmFuncs := make(map[string]*vm.Function)
	for name, fn := range compilerFuncs {
		vmFuncs[name] = &vm.Function{
			Name:         fn.Name,
			Instructions: fn.Instructions,
			Constants:    fn.Constants,
			Parameters:   convertToVMParameters(fn.Parameters),
			IsVariadic:   fn.IsVariadic,
			IsGenerator:  fn.IsGenerator,
		}
	}
	return vmFuncs
}

func convertToVMClassConstants(compilerConsts map[string]*CompilerClassConstant) map[string]*vm.ClassConstant {
	vmConsts := make(map[string]*vm.ClassConstant)
	for name, constant := range compilerConsts {
		vmConsts[name] = &vm.ClassConstant{
			Name:       constant.Name,
			Value:      constant.Value,
			Visibility: constant.Visibility,
			Type:       constant.Type,
			IsFinal:    constant.IsFinal,
			IsAbstract: constant.IsAbstract,
		}
	}
	return vmConsts
}