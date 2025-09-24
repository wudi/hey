package runtime

import (
	"unicode"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetCtypeFunctions returns all character type checking functions
func GetCtypeFunctions() []*registry.Function {
	return []*registry.Function{
		{
			Name: "ctype_alnum",
			Parameters: []*registry.Parameter{
				{Name: "text", Type: "mixed"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				text := args[0].ToString()
				if text == "" {
					return values.NewBool(false), nil
				}

				for _, r := range text {
					if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
						return values.NewBool(false), nil
					}
				}
				return values.NewBool(true), nil
			},
		},
		{
			Name: "ctype_alpha",
			Parameters: []*registry.Parameter{
				{Name: "text", Type: "mixed"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				text := args[0].ToString()
				if text == "" {
					return values.NewBool(false), nil
				}

				for _, r := range text {
					if !unicode.IsLetter(r) {
						return values.NewBool(false), nil
					}
				}
				return values.NewBool(true), nil
			},
		},
		{
			Name: "ctype_digit",
			Parameters: []*registry.Parameter{
				{Name: "text", Type: "mixed"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				text := args[0].ToString()
				if text == "" {
					return values.NewBool(false), nil
				}

				for _, r := range text {
					if !unicode.IsDigit(r) {
						return values.NewBool(false), nil
					}
				}
				return values.NewBool(true), nil
			},
		},
		{
			Name: "ctype_lower",
			Parameters: []*registry.Parameter{
				{Name: "text", Type: "mixed"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				text := args[0].ToString()
				if text == "" {
					return values.NewBool(false), nil
				}

				for _, r := range text {
					if !unicode.IsLower(r) && unicode.IsLetter(r) {
						return values.NewBool(false), nil
					}
					if !unicode.IsLetter(r) {
						return values.NewBool(false), nil
					}
				}
				return values.NewBool(true), nil
			},
		},
		{
			Name: "ctype_upper",
			Parameters: []*registry.Parameter{
				{Name: "text", Type: "mixed"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				text := args[0].ToString()
				if text == "" {
					return values.NewBool(false), nil
				}

				for _, r := range text {
					if !unicode.IsUpper(r) && unicode.IsLetter(r) {
						return values.NewBool(false), nil
					}
					if !unicode.IsLetter(r) {
						return values.NewBool(false), nil
					}
				}
				return values.NewBool(true), nil
			},
		},
		{
			Name: "ctype_space",
			Parameters: []*registry.Parameter{
				{Name: "text", Type: "mixed"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				text := args[0].ToString()
				if text == "" {
					return values.NewBool(false), nil
				}

				for _, r := range text {
					if !unicode.IsSpace(r) {
						return values.NewBool(false), nil
					}
				}
				return values.NewBool(true), nil
			},
		},
	}
}