package handler

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/wudi/hey/values"
	"github.com/wudi/hey/vm"
)

func SetupCGIVariables(vmCtx *vm.ExecutionContext, params map[string]string, stdin []byte) {
	server := values.NewArray()
	for k, v := range params {
		server.ArraySet(values.NewString(k), values.NewString(v))
	}

	if argc, ok := params["argc"]; ok {
		server.ArraySet(values.NewString("argc"), values.NewString(argc))
	}
	if argv, ok := params["argv"]; ok {
		server.ArraySet(values.NewString("argv"), values.NewString(argv))
	}

	vmCtx.GlobalVars.Store("$_SERVER", server)

	if qs, ok := params["QUERY_STRING"]; ok && qs != "" {
		vmCtx.GlobalVars.Store("$_GET", parseQueryString(qs))
	} else {
		vmCtx.GlobalVars.Store("$_GET", values.NewArray())
	}

	if method, ok := params["REQUEST_METHOD"]; ok && (method == "POST" || method == "PUT" || method == "PATCH") {
		contentType := params["CONTENT_TYPE"]
		vmCtx.GlobalVars.Store("$_POST", parsePostData(stdin, contentType))
	} else {
		vmCtx.GlobalVars.Store("$_POST", values.NewArray())
	}

	if cookie, ok := params["HTTP_COOKIE"]; ok && cookie != "" {
		vmCtx.GlobalVars.Store("$_COOKIE", parseCookies(cookie))
	} else {
		vmCtx.GlobalVars.Store("$_COOKIE", values.NewArray())
	}

	requestArr := values.NewArray()
	if get, ok := vmCtx.GlobalVars.Load("$_GET"); ok {
		if getArr, ok := get.(*values.Value); ok && getArr.IsArray() {
			arr := getArr.Data.(*values.Array)
			for k, v := range arr.Elements {
				requestArr.ArraySet(convertToValue(k), v)
			}
		}
	}
	if post, ok := vmCtx.GlobalVars.Load("$_POST"); ok {
		if postArr, ok := post.(*values.Value); ok && postArr.IsArray() {
			arr := postArr.Data.(*values.Array)
			for k, v := range arr.Elements {
				requestArr.ArraySet(convertToValue(k), v)
			}
		}
	}
	vmCtx.GlobalVars.Store("$_REQUEST", requestArr)

	vmCtx.GlobalVars.Store("$_FILES", values.NewArray())

	vmCtx.GlobalVars.Store("$_ENV", values.NewArray())
}

func parseQueryString(qs string) *values.Value {
	arr := values.NewArray()
	parsed, err := url.ParseQuery(qs)
	if err != nil {
		return arr
	}

	for k, vals := range parsed {
		if len(vals) == 1 {
			arr.ArraySet(values.NewString(k), values.NewString(vals[0]))
		} else {
			subArr := values.NewArray()
			for _, v := range vals {
				subArr.ArraySet(nil, values.NewString(v))
			}
			arr.ArraySet(values.NewString(k), subArr)
		}
	}

	return arr
}

func parsePostData(data []byte, contentType string) *values.Value {
	arr := values.NewArray()

	if !strings.Contains(contentType, "application/x-www-form-urlencoded") {
		return arr
	}

	parsed, err := url.ParseQuery(string(data))
	if err != nil {
		return arr
	}

	for k, vals := range parsed {
		if len(vals) == 1 {
			arr.ArraySet(values.NewString(k), values.NewString(vals[0]))
		} else {
			subArr := values.NewArray()
			for _, v := range vals {
				subArr.ArraySet(nil, values.NewString(v))
			}
			arr.ArraySet(values.NewString(k), subArr)
		}
	}

	return arr
}

func parseCookies(cookieHeader string) *values.Value {
	arr := values.NewArray()

	cookies := strings.Split(cookieHeader, ";")
	for _, cookie := range cookies {
		cookie = strings.TrimSpace(cookie)
		parts := strings.SplitN(cookie, "=", 2)
		if len(parts) == 2 {
			name := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if decodedValue, err := url.QueryUnescape(value); err == nil {
				arr.ArraySet(values.NewString(name), values.NewString(decodedValue))
			} else {
				arr.ArraySet(values.NewString(name), values.NewString(value))
			}
		}
	}

	return arr
}

func convertToValue(key interface{}) *values.Value {
	switch k := key.(type) {
	case int64:
		return values.NewInt(k)
	case string:
		return values.NewString(k)
	default:
		return values.NewString(fmt.Sprintf("%v", k))
	}
}