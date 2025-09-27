package vm

import (
	"fmt"
	"strings"
	"sync"

	"github.com/wudi/hey/registry"
)

type HTTPContext struct {
	mu             sync.RWMutex
	headers        []registry.HTTPHeader
	responseCode   int
	headersSent    bool
	headersSentAt  string
	requestHeaders map[string]string
}


func NewHTTPContext() *HTTPContext {
	return &HTTPContext{
		headers:        make([]registry.HTTPHeader, 0),
		responseCode:   200,
		headersSent:    false,
		requestHeaders: make(map[string]string),
	}
}

func (h *HTTPContext) AddHeader(name, value string, replace bool) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.headersSent {
		return fmt.Errorf("Cannot modify header information - headers already sent%s", h.getHeadersSentLocation())
	}

	nameLower := strings.ToLower(name)

	if replace {
		newHeaders := make([]registry.HTTPHeader, 0)
		for _, header := range h.headers {
			if strings.ToLower(header.Name) != nameLower {
				newHeaders = append(newHeaders, header)
			}
		}
		newHeaders = append(newHeaders, registry.HTTPHeader{Name: name, Value: value})
		h.headers = newHeaders
	} else {
		h.headers = append(h.headers, registry.HTTPHeader{Name: name, Value: value})
	}

	return nil
}

func (h *HTTPContext) SetResponseCode(code int) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.headersSent {
		return fmt.Errorf("Cannot set response code - headers already sent%s", h.getHeadersSentLocation())
	}

	h.responseCode = code
	return nil
}

func (h *HTTPContext) GetResponseCode() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.responseCode
}

func (h *HTTPContext) GetHeaders() []registry.HTTPHeader {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]registry.HTTPHeader, len(h.headers))
	copy(result, h.headers)
	return result
}

func (h *HTTPContext) GetHeadersList() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]string, len(h.headers))
	for i, header := range h.headers {
		result[i] = fmt.Sprintf("%s: %s", header.Name, header.Value)
	}
	return result
}

func (h *HTTPContext) MarkHeadersSent(location string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.headersSent = true
	h.headersSentAt = location
}

func (h *HTTPContext) AreHeadersSent() (bool, string) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.headersSent, h.headersSentAt
}

func (h *HTTPContext) getHeadersSentLocation() string {
	if h.headersSentAt != "" {
		return fmt.Sprintf(" (output started at %s)", h.headersSentAt)
	}
	return ""
}

func (h *HTTPContext) SetRequestHeaders(headers map[string]string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.requestHeaders = headers
}

func (h *HTTPContext) GetRequestHeaders() map[string]string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make(map[string]string)
	for k, v := range h.requestHeaders {
		result[k] = v
	}
	return result
}

func (h *HTTPContext) FormatHeadersForFastCGI() string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var sb strings.Builder

	if h.responseCode != 200 {
		sb.WriteString(fmt.Sprintf("Status: %d\r\n", h.responseCode))
	}

	for _, header := range h.headers {
		sb.WriteString(fmt.Sprintf("%s: %s\r\n", header.Name, header.Value))
	}

	sb.WriteString("\r\n")
	return sb.String()
}