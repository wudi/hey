package runtime

import (
	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetI18nFunctions returns WordPress internationalization (i18n) functions
// These are stub implementations that return the original text without translation
func GetI18nFunctions() []*registry.Function {
	return []*registry.Function{
		// __() - Retrieve the translation of text
		{
			Name: "__",
			Parameters: []*registry.Parameter{
				{Name: "text", Type: "string"},
				{Name: "domain", Type: "string", DefaultValue: values.NewString("default")},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewString(""), nil
				}
				// For now, just return the original text without translation
				return values.NewString(args[0].ToString()), nil
			},
		},
		// _e() - Display translated text
		{
			Name: "_e",
			Parameters: []*registry.Parameter{
				{Name: "text", Type: "string"},
				{Name: "domain", Type: "string", DefaultValue: values.NewString("default")},
			},
			ReturnType: "void",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewNull(), nil
				}
				// Echo the translated text
				text := args[0].ToString()
				ctx.WriteOutput(values.NewString(text))
				return values.NewNull(), nil
			},
		},
		// _x() - Retrieve translated string with gettext context
		{
			Name: "_x",
			Parameters: []*registry.Parameter{
				{Name: "text", Type: "string"},
				{Name: "context", Type: "string"},
				{Name: "domain", Type: "string", DefaultValue: values.NewString("default")},
			},
			ReturnType: "string",
			MinArgs:    2,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewString(""), nil
				}
				// Return the original text without translation
				return values.NewString(args[0].ToString()), nil
			},
		},
		// _n() - Retrieve translated singular/plural form
		{
			Name: "_n",
			Parameters: []*registry.Parameter{
				{Name: "single", Type: "string"},
				{Name: "plural", Type: "string"},
				{Name: "number", Type: "int"},
				{Name: "domain", Type: "string", DefaultValue: values.NewString("default")},
			},
			ReturnType: "string",
			MinArgs:    3,
			MaxArgs:    4,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 3 {
					return values.NewString(""), nil
				}
				// Simple English plural logic: 1 = singular, else plural
				number := args[2].ToInt()
				if number == 1 {
					return values.NewString(args[0].ToString()), nil
				}
				return values.NewString(args[1].ToString()), nil
			},
		},
		// esc_html() - Escape HTML
		{
			Name: "esc_html",
			Parameters: []*registry.Parameter{
				{Name: "text", Type: "string"},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewString(""), nil
				}
				// Use htmlspecialchars-like escaping
				text := args[0].ToString()
				text = escapeHTML(text)
				return values.NewString(text), nil
			},
		},
		// esc_attr() - Escape for HTML attributes
		{
			Name: "esc_attr",
			Parameters: []*registry.Parameter{
				{Name: "text", Type: "string"},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewString(""), nil
				}
				text := args[0].ToString()
				text = escapeHTML(text)
				return values.NewString(text), nil
			},
		},
		// esc_url() - Escape for URLs
		{
			Name: "esc_url",
			Parameters: []*registry.Parameter{
				{Name: "url", Type: "string"},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewString(""), nil
				}
				// For now, just return the URL as-is
				// A proper implementation would validate and sanitize
				return values.NewString(args[0].ToString()), nil
			},
		},
		// wp_fix_server_vars() - Standardize $_SERVER variables
		{
			Name: "wp_fix_server_vars",
			Parameters: []*registry.Parameter{},
			ReturnType: "void",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Do nothing for now
				return values.NewNull(), nil
			},
		},
		// wp_load_translations_early() - Load translation files early
		{
			Name: "wp_load_translations_early",
			Parameters: []*registry.Parameter{},
			ReturnType: "void",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Do nothing for now
				return values.NewNull(), nil
			},
		},
	}
}

// escapeHTML performs HTML escaping similar to PHP's htmlspecialchars
func escapeHTML(text string) string {
	replacements := map[string]string{
		"&":  "&amp;",
		"<":  "&lt;",
		">":  "&gt;",
		"\"": "&quot;",
		"'":  "&#039;",
	}

	result := text
	for old, new := range replacements {
		result = replaceAll(result, old, new)
	}
	return result
}

// replaceAll is a helper to replace all occurrences
func replaceAll(s, old, new string) string {
	result := ""
	for len(s) > 0 {
		if len(s) >= len(old) && s[:len(old)] == old {
			result += new
			s = s[len(old):]
		} else {
			result += s[:1]
			s = s[1:]
		}
	}
	return result
}
