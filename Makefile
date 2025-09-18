# Hey PHP Interpreter Makefile

# Binary name
BINARY_NAME=hey

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build directory
BUILD_DIR=build

# Main package path
MAIN_PATH=./cmd/hey

# Build flags
LDFLAGS=-ldflags "-s -w"

# Default target
.PHONY: all
all: build

# Build the main binary
.PHONY: build
build:
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(LDFLAGS) $(MAIN_PATH)

# Build all demo binaries
.PHONY: build-all
build-all: build
	$(GOBUILD) -o $(BUILD_DIR)/php-parser ./cmd/php-parser
	$(GOBUILD) -o $(BUILD_DIR)/vm-demo ./cmd/vm-demo
	$(GOBUILD) -o $(BUILD_DIR)/bytecode-demo ./cmd/bytecode-demo

# Run all tests
.PHONY: test
test:
	$(GOTEST) -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	$(GOTEST) -v -cover ./...

# Run tests for specific package
.PHONY: test-parser
test-parser:
	$(GOTEST) -v ./compiler/parser

.PHONY: test-lexer
test-lexer:
	$(GOTEST) -v ./compiler/lexer

.PHONY: test-vm
test-vm:
	$(GOTEST) -v ./vm

.PHONY: test-compiler
test-compiler:
	$(GOTEST) -v ./compiler

# Clean build artifacts
.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)

# Install dependencies
.PHONY: deps
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Run the interpreter with a test command
.PHONY: run
run: build
	./$(BINARY_NAME) -r 'echo "Hello from Hey!";'

# Build and install to GOPATH/bin
.PHONY: install
install:
	$(GOCMD) install $(MAIN_PATH)

# Format code
.PHONY: fmt
fmt:
	$(GOCMD) fmt ./...

# Run linter
.PHONY: lint
lint:
	golangci-lint run

# Run vet
.PHONY: vet
vet:
	$(GOCMD) vet ./...

# Quick build without optimizations (faster for development)
.PHONY: dev
dev:
	$(GOBUILD) -o $(BINARY_NAME) $(MAIN_PATH)

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  make build         - Build the main hey binary"
	@echo "  make build-all     - Build all binaries including demos"
	@echo "  make test          - Run all tests"
	@echo "  make test-coverage - Run tests with coverage"
	@echo "  make test-parser   - Run parser tests only"
	@echo "  make test-lexer    - Run lexer tests only"
	@echo "  make test-vm       - Run VM tests only"
	@echo "  make test-compiler - Run compiler tests only"
	@echo "  make clean         - Remove build artifacts"
	@echo "  make deps          - Download and tidy dependencies"
	@echo "  make run           - Build and run with test command"
	@echo "  make install       - Install to GOPATH/bin"
	@echo "  make fmt           - Format code"
	@echo "  make lint          - Run linter (requires golangci-lint)"
	@echo "  make vet           - Run go vet"
	@echo "  make dev           - Quick build for development"
	@echo "  make help          - Show this help message"