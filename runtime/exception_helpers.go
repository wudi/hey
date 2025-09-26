package runtime

import (
	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func CreateException(ctx registry.BuiltinCallContext, className, message string) *values.Value {
	classDesc, err := ctx.SymbolRegistry().GetClass(className)
	if err != nil || classDesc == nil {
		return nil
	}

	exceptionValue := values.NewObject(className)
	exceptionObj := exceptionValue.Data.(*values.Object)
	exceptionObj.Properties["message"] = values.NewString(message)
	exceptionObj.Properties["code"] = values.NewInt(0)
	exceptionObj.Properties["file"] = values.NewString("")
	exceptionObj.Properties["line"] = values.NewInt(0)

	return exceptionValue
}