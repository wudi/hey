package stdlib

import (
	"github.com/wudi/hey/compiler/values"
)

// initVariables initializes built-in PHP variables
func (stdlib *StandardLibrary) initVariables() {
	// Initialize $_SERVER superglobal
	stdlib.Variables["_SERVER"] = createServerArray()

	// Initialize $_GET superglobal
	stdlib.Variables["_GET"] = values.NewArray()

	// Initialize $_POST superglobal
	stdlib.Variables["_POST"] = values.NewArray()

	// Initialize $_FILES superglobal
	stdlib.Variables["_FILES"] = values.NewArray()

	// Initialize $_COOKIE superglobal
	stdlib.Variables["_COOKIE"] = values.NewArray()

	// Initialize $_SESSION superglobal
	stdlib.Variables["_SESSION"] = values.NewArray()

	// Initialize $_REQUEST superglobal
	stdlib.Variables["_REQUEST"] = values.NewArray()

	// Initialize $_ENV superglobal
	stdlib.Variables["_ENV"] = createEnvArray()

	// Initialize $GLOBALS superglobal
	stdlib.Variables["GLOBALS"] = values.NewArray()

	// Initialize $argc (command line argument count)
	stdlib.Variables["argc"] = values.NewInt(0)

	// Initialize $argv (command line arguments)
	argv := values.NewArray()
	argv.ArraySet(values.NewInt(0), values.NewString("hey"))
	stdlib.Variables["argv"] = argv

	// Initialize $http_response_header
	stdlib.Variables["http_response_header"] = values.NewArray()
}

// createServerArray creates the $_SERVER superglobal array
func createServerArray() *values.Value {
	server := values.NewArray()

	// Basic server information
	server.ArraySet(values.NewString("SERVER_SOFTWARE"), values.NewString("Hey"))
	server.ArraySet(values.NewString("SERVER_NAME"), values.NewString("localhost"))
	server.ArraySet(values.NewString("SERVER_ADDR"), values.NewString("127.0.0.1"))
	server.ArraySet(values.NewString("SERVER_PORT"), values.NewString("80"))
	server.ArraySet(values.NewString("REMOTE_ADDR"), values.NewString("127.0.0.1"))
	server.ArraySet(values.NewString("DOCUMENT_ROOT"), values.NewString("/"))
	server.ArraySet(values.NewString("SERVER_ADMIN"), values.NewString("admin@localhost"))
	server.ArraySet(values.NewString("SCRIPT_FILENAME"), values.NewString("/index.php"))
	server.ArraySet(values.NewString("REMOTE_PORT"), values.NewString("0"))
	server.ArraySet(values.NewString("GATEWAY_INTERFACE"), values.NewString("CGI/1.1"))
	server.ArraySet(values.NewString("SERVER_PROTOCOL"), values.NewString("HTTP/1.1"))
	server.ArraySet(values.NewString("REQUEST_METHOD"), values.NewString("GET"))
	server.ArraySet(values.NewString("REQUEST_TIME"), values.NewInt(0))
	server.ArraySet(values.NewString("REQUEST_TIME_FLOAT"), values.NewFloat(0.0))
	server.ArraySet(values.NewString("QUERY_STRING"), values.NewString(""))
	server.ArraySet(values.NewString("HTTP_ACCEPT"), values.NewString("*/*"))
	server.ArraySet(values.NewString("HTTP_ACCEPT_CHARSET"), values.NewString(""))
	server.ArraySet(values.NewString("HTTP_ACCEPT_ENCODING"), values.NewString(""))
	server.ArraySet(values.NewString("HTTP_ACCEPT_LANGUAGE"), values.NewString(""))
	server.ArraySet(values.NewString("HTTP_CONNECTION"), values.NewString(""))
	server.ArraySet(values.NewString("HTTP_HOST"), values.NewString("localhost"))
	server.ArraySet(values.NewString("HTTP_REFERER"), values.NewString(""))
	server.ArraySet(values.NewString("HTTP_USER_AGENT"), values.NewString("Hey"))
	server.ArraySet(values.NewString("HTTPS"), values.NewString(""))
	server.ArraySet(values.NewString("REQUEST_URI"), values.NewString("/"))
	server.ArraySet(values.NewString("SCRIPT_NAME"), values.NewString("/index.php"))
	server.ArraySet(values.NewString("PATH_INFO"), values.NewString(""))
	server.ArraySet(values.NewString("PATH_TRANSLATED"), values.NewString(""))
	server.ArraySet(values.NewString("PHP_SELF"), values.NewString("/index.php"))

	// CLI-specific values
	server.ArraySet(values.NewString("PHP_CLI_PROCESS_TITLE"), values.NewString("php"))
	server.ArraySet(values.NewString("_"), values.NewString("/usr/bin/php"))

	return server
}

// createEnvArray creates the $_ENV superglobal array
func createEnvArray() *values.Value {
	env := values.NewArray()

	// Common environment variables
	env.ArraySet(values.NewString("PATH"), values.NewString("/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"))
	env.ArraySet(values.NewString("HOME"), values.NewString("/home/user"))
	env.ArraySet(values.NewString("USER"), values.NewString("user"))
	env.ArraySet(values.NewString("SHELL"), values.NewString("/bin/bash"))
	env.ArraySet(values.NewString("LANG"), values.NewString("en_US.UTF-8"))
	env.ArraySet(values.NewString("TERM"), values.NewString("xterm"))
	env.ArraySet(values.NewString("PWD"), values.NewString("/"))
	env.ArraySet(values.NewString("TMPDIR"), values.NewString("/tmp"))

	return env
}

// GetSuperGlobal returns a superglobal variable by name
func (stdlib *StandardLibrary) GetSuperGlobal(name string) *values.Value {
	switch name {
	case "_SERVER", "_GET", "_POST", "_FILES", "_COOKIE", "_SESSION", "_REQUEST", "_ENV", "GLOBALS":
		if val, exists := stdlib.Variables[name]; exists {
			return val
		}
		return values.NewArray()
	default:
		if val, exists := stdlib.Variables[name]; exists {
			return val
		}
		return values.NewNull()
	}
}

// SetSuperGlobal sets a superglobal variable
func (stdlib *StandardLibrary) SetSuperGlobal(name string, value *values.Value) {
	stdlib.Variables[name] = value
}

// IsSuperGlobal checks if a variable name is a superglobal
func IsSuperGlobal(name string) bool {
	switch name {
	case "_SERVER", "_GET", "_POST", "_FILES", "_COOKIE", "_SESSION", "_REQUEST", "_ENV", "GLOBALS", "argc", "argv":
		return true
	default:
		return false
	}
}
