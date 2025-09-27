package handler

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/wudi/hey/compiler"
	"github.com/wudi/hey/compiler/lexer"
	"github.com/wudi/hey/compiler/parser"
	"github.com/wudi/hey/pkg/fastcgi"
	"github.com/wudi/hey/runtime"
	"github.com/wudi/hey/vm"
	"github.com/wudi/hey/vmfactory"
)

type RequestHandler struct {
	vmFactory *vmfactory.VMFactory
}

func NewRequestHandler(vmFactory *vmfactory.VMFactory) *RequestHandler {
	return &RequestHandler{
		vmFactory: vmFactory,
	}
}

func (h *RequestHandler) HandleRequest(ctx context.Context, proto *fastcgi.Protocol, req *fastcgi.Request) error {
	scriptFile, ok := req.Params["SCRIPT_FILENAME"]
	if !ok || scriptFile == "" {
		return h.sendError(proto, req.ID, "SCRIPT_FILENAME not provided")
	}

	if _, err := os.Stat(scriptFile); os.IsNotExist(err) {
		return h.sendError(proto, req.ID, fmt.Sprintf("File not found: %s", scriptFile))
	}

	code, err := os.ReadFile(scriptFile)
	if err != nil {
		return h.sendError(proto, req.ID, fmt.Sprintf("Failed to read file: %v", err))
	}

	l := lexer.New(string(code))
	p := parser.New(l)
	prog := p.ParseProgram()

	if len(p.Errors()) != 0 {
		var errBuf bytes.Buffer
		for _, msg := range p.Errors() {
			errBuf.WriteString(msg)
			errBuf.WriteString("\n")
		}
		return h.sendError(proto, req.ID, errBuf.String())
	}

	comp := compiler.NewCompiler()
	comp.SetCurrentFile(scriptFile)
	if err := comp.Compile(prog); err != nil {
		return h.sendError(proto, req.ID, fmt.Sprintf("Compilation error: %v", err))
	}

	vmCtx := vm.NewExecutionContext()

	var outBuf bytes.Buffer
	vmCtx.OutputWriter = &outBuf

	SetupCGIVariables(vmCtx, req.Params, req.Stdin)

	extractRequestHeaders(vmCtx, req.Params)

	variables := runtime.GlobalVMIntegration.GetAllVariables()
	for name, value := range variables {
		vmCtx.GlobalVars.Store(name, value)
	}

	vmachine := h.vmFactory.CreateVM()

	err = vmachine.Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(),
		comp.Functions(), comp.Classes(), comp.Interfaces(), comp.Traits())

	vmachine.CallAllDestructors(vmCtx)

	var stderrBuf bytes.Buffer
	if err != nil {
		stderrBuf.WriteString(fmt.Sprintf("Runtime error: %v\n", err))
	}

	exitCode := 0
	if vmCtx.Halted {
		exitCode = vmCtx.ExitCode
	}

	httpHeaders := vmCtx.HTTPContext.FormatHeadersForFastCGI()

	var response bytes.Buffer
	response.WriteString(httpHeaders)
	response.Write(outBuf.Bytes())

	return proto.SendResponse(req.ID, response.Bytes(), stderrBuf.Bytes(), exitCode)
}

func extractRequestHeaders(vmCtx *vm.ExecutionContext, params map[string]string) {
	headers := make(map[string]string)

	for key, value := range params {
		if len(key) > 5 && key[:5] == "HTTP_" {
			headerName := key[5:]
			headerName = strings.ReplaceAll(headerName, "_", "-")
			headers[headerName] = value
		}
	}

	if contentType, ok := params["CONTENT_TYPE"]; ok {
		headers["Content-Type"] = contentType
	}
	if contentLength, ok := params["CONTENT_LENGTH"]; ok {
		headers["Content-Length"] = contentLength
	}

	vmCtx.HTTPContext.SetRequestHeaders(headers)
}

func (h *RequestHandler) sendError(proto *fastcgi.Protocol, requestID uint16, errMsg string) error {
	stderr := []byte(errMsg)
	stdout := []byte(fmt.Sprintf("Status: 500 Internal Server Error\r\nContent-Type: text/plain\r\n\r\n%s", errMsg))
	return proto.SendResponse(requestID, stdout, stderr, 1)
}