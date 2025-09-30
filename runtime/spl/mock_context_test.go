package spl

import (
	"fmt"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// SharedMockContext for testing - used across all SPL tests
type SharedMockContext struct {
	registry *registry.Registry
}

func (m *SharedMockContext) WriteOutput(val *values.Value) error                      { return nil }
func (m *SharedMockContext) GetGlobal(name string) (*values.Value, bool)              { return nil, false }
func (m *SharedMockContext) SetGlobal(name string, val *values.Value)                 {}
func (m *SharedMockContext) SymbolRegistry() *registry.Registry                       { return m.registry }
func (m *SharedMockContext) LookupUserFunction(name string) (*registry.Function, bool) { return nil, false }
func (m *SharedMockContext) CallUserFunction(function *registry.Function, args []*values.Value) (*values.Value, error) {
	return nil, nil
}
func (m *SharedMockContext) SimpleCallUserFunction(function *registry.Function, args []*values.Value) (*values.Value, error) {
	return nil, nil
}
func (m *SharedMockContext) LookupUserClass(name string) (*registry.Class, bool)      { return nil, false }
func (m *SharedMockContext) Halt(exitCode int, message string) error                  { return nil }
func (m *SharedMockContext) GetExecutionContext() registry.ExecutionContextInterface  { return nil }
func (m *SharedMockContext) GetOutputBufferStack() registry.OutputBufferStackInterface { return nil }
func (m *SharedMockContext) GetCurrentFunctionArgCount() (int, error)                 { return 0, nil }
func (m *SharedMockContext) GetCurrentFunctionArg(index int) (*values.Value, error)   { return nil, nil }
func (m *SharedMockContext) GetCurrentFunctionArgs() ([]*values.Value, error)         { return nil, nil }
func (m *SharedMockContext) ThrowException(exception *values.Value) error { return fmt.Errorf("exception thrown in test mock: %v", exception) }

// HTTP methods - missing from original mocks
func (m *SharedMockContext) GetHTTPContext() registry.HTTPContext { return &SharedMockHTTPContext{} }
func (m *SharedMockContext) ResetHTTPContext() {}
func (m *SharedMockContext) RemoveHTTPHeader(name string) {}

// SharedMockHTTPContext is a simple mock for HTTPContext interface
type SharedMockHTTPContext struct{}

func (m *SharedMockHTTPContext) AddHeader(name, value string, replace bool) error { return nil }
func (m *SharedMockHTTPContext) SetResponseCode(code int) error { return nil }
func (m *SharedMockHTTPContext) GetResponseCode() int { return 200 }
func (m *SharedMockHTTPContext) GetHeaders() []registry.HTTPHeader { return nil }
func (m *SharedMockHTTPContext) GetHeadersList() []string { return nil }
func (m *SharedMockHTTPContext) MarkHeadersSent(location string) {}
func (m *SharedMockHTTPContext) AreHeadersSent() (bool, string) { return false, "" }
func (m *SharedMockHTTPContext) SetRequestHeaders(headers map[string]string) {}
func (m *SharedMockHTTPContext) GetRequestHeaders() map[string]string { return nil }
func (m *SharedMockHTTPContext) FormatHeadersForFastCGI() string { return "" }

// NewMockContext creates a new mock context for testing
func NewMockContext() *SharedMockContext {
	registry.Initialize()
	return &SharedMockContext{registry: registry.GlobalRegistry}
}