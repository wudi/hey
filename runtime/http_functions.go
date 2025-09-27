package runtime

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func GetHTTPFunctions() []*registry.Function {
	return []*registry.Function{
		{
			Name:      "header",
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 1 {
					return nil, fmt.Errorf("header() expects at least 1 parameter, %d given", len(args))
				}

				header := args[0].ToString()
				replace := true
				responseCode := 0

				if len(args) >= 2 {
					replace = args[1].ToBool()
				}

				if len(args) >= 3 {
					responseCode = int(args[2].ToInt())
				}

				httpCtx := ctx.GetHTTPContext()
				if httpCtx == nil {
					return values.NewNull(), nil
				}

				parts := strings.SplitN(header, ":", 2)
				if len(parts) != 2 {
					return values.NewNull(), nil
				}

				name := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				if err := httpCtx.AddHeader(name, value, replace); err != nil {
					return values.NewBool(false), nil
				}

				if responseCode > 0 {
					httpCtx.SetResponseCode(responseCode)
				}

				return values.NewNull(), nil
			},
		},
		{
			Name:      "header_remove",
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				httpCtx := ctx.GetHTTPContext()
				if httpCtx == nil {
					return values.NewNull(), nil
				}

				if len(args) == 0 {
					ctx.ResetHTTPContext()
					return values.NewNull(), nil
				}

				name := args[0].ToString()
				ctx.RemoveHTTPHeader(name)

				return values.NewNull(), nil
			},
		},
		{
			Name:      "headers_list",
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				httpCtx := ctx.GetHTTPContext()
				if httpCtx == nil {
					return values.NewArray(), nil
				}

				headersList := httpCtx.GetHeadersList()
				arr := values.NewArray()
				for _, header := range headersList {
					arr.ArraySet(nil, values.NewString(header))
				}

				return arr, nil
			},
		},
		{
			Name:      "headers_sent",
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				httpCtx := ctx.GetHTTPContext()
				if httpCtx == nil {
					return values.NewBool(false), nil
				}

				sent, location := httpCtx.AreHeadersSent()

				if len(args) >= 1 && args[0].IsReference() {
					ref := args[0].Data.(*values.Reference)
					parts := strings.SplitN(location, ":", 2)
					if len(parts) > 0 {
						ref.Target = values.NewString(parts[0])
					}
				}

				if len(args) >= 2 && args[1].IsReference() {
					ref := args[1].Data.(*values.Reference)
					parts := strings.SplitN(location, ":", 2)
					if len(parts) > 1 {
						ref.Target = values.NewString(parts[1])
					}
				}

				return values.NewBool(sent), nil
			},
		},
		{
			Name:      "http_response_code",
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				httpCtx := ctx.GetHTTPContext()
				if httpCtx == nil {
					return values.NewBool(false), nil
				}

				currentCode := httpCtx.GetResponseCode()

				if len(args) > 0 {
					newCode := int(args[0].ToInt())
					if err := httpCtx.SetResponseCode(newCode); err != nil {
						return values.NewBool(false), nil
					}
					return values.NewInt(int64(currentCode)), nil
				}

				return values.NewInt(int64(currentCode)), nil
			},
		},
		{
			Name:      "setcookie",
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				return setcookieImpl(ctx, args, false)
			},
		},
		{
			Name:      "setrawcookie",
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				return setcookieImpl(ctx, args, true)
			},
		},
		{
			Name:      "getallheaders",
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				httpCtx := ctx.GetHTTPContext()
				if httpCtx == nil {
					return values.NewArray(), nil
				}

				headers := httpCtx.GetRequestHeaders()
				arr := values.NewArray()

				for name, value := range headers {
					arr.ArraySet(values.NewString(name), values.NewString(value))
				}

				return arr, nil
			},
		},
	}
}

func setcookieImpl(ctx registry.BuiltinCallContext, args []*values.Value, raw bool) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("setcookie() expects at least 1 parameter, %d given", len(args))
	}

	httpCtx := ctx.GetHTTPContext()
	if httpCtx == nil {
		return values.NewBool(false), nil
	}

	name := args[0].ToString()
	value := ""
	expires := int64(0)
	path := ""
	domain := ""
	secure := false
	httponly := false

	if len(args) >= 2 {
		value = args[1].ToString()
	}
	if len(args) >= 3 {
		expires = args[2].ToInt()
	}
	if len(args) >= 4 {
		path = args[3].ToString()
	}
	if len(args) >= 5 {
		domain = args[4].ToString()
	}
	if len(args) >= 6 {
		secure = args[5].ToBool()
	}
	if len(args) >= 7 {
		httponly = args[6].ToBool()
	}

	var cookieValue string
	if raw {
		cookieValue = value
	} else {
		cookieValue = url.QueryEscape(value)
	}

	cookie := fmt.Sprintf("%s=%s", name, cookieValue)

	if expires > 0 {
		expiresTime := time.Unix(expires, 0).UTC()
		cookie += fmt.Sprintf("; Expires=%s", expiresTime.Format(time.RFC1123))
		cookie += fmt.Sprintf("; Max-Age=%d", expires-time.Now().Unix())
	}

	if path != "" {
		cookie += fmt.Sprintf("; Path=%s", path)
	}

	if domain != "" {
		cookie += fmt.Sprintf("; Domain=%s", domain)
	}

	if secure {
		cookie += "; Secure"
	}

	if httponly {
		cookie += "; HttpOnly"
	}

	if err := httpCtx.AddHeader("Set-Cookie", cookie, false); err != nil {
		return values.NewBool(false), nil
	}

	return values.NewBool(true), nil
}