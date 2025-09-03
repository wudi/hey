package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/lexer"
)

func TestParsing_InterfaceMethodsWithReferenceReturn(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedErrors int
		validate       func(t *testing.T, interfaceDecl *ast.InterfaceDeclaration)
	}{
		{
			name: "Interface with reference return method",
			input: `<?php
interface EntityInterface {
    public function &get(string $field): mixed;
}
?>`,
			expectedErrors: 0,
			validate: func(t *testing.T, interfaceDecl *ast.InterfaceDeclaration) {
				assert.Equal(t, "EntityInterface", interfaceDecl.Name.Name)
				assert.Equal(t, 1, len(interfaceDecl.Methods))
				
				method := interfaceDecl.Methods[0]
				assert.Equal(t, "get", method.Name.Name)
				assert.Equal(t, "public", method.Visibility)
				assert.True(t, method.ByReference, "Expected method to have reference return")
				assert.Equal(t, 1, len(method.Parameters.Parameters))
				assert.Equal(t, "$field", method.Parameters.Parameters[0].Name.(*ast.IdentifierNode).Name)
				assert.NotNil(t, method.ReturnType)
				assert.Equal(t, "mixed", method.ReturnType.Name)
			},
		},
		{
			name: "Interface extending multiple interfaces with reference return",
			input: `<?php
interface EntityInterface extends ArrayAccess, JsonSerializable, Stringable {
    public function &get(string $field): mixed;
}
?>`,
			expectedErrors: 0,
			validate: func(t *testing.T, interfaceDecl *ast.InterfaceDeclaration) {
				assert.Equal(t, "EntityInterface", interfaceDecl.Name.Name)
				assert.Equal(t, 3, len(interfaceDecl.Extends))
				assert.Equal(t, "ArrayAccess", interfaceDecl.Extends[0].Name)
				assert.Equal(t, "JsonSerializable", interfaceDecl.Extends[1].Name)
				assert.Equal(t, "Stringable", interfaceDecl.Extends[2].Name)
				
				assert.Equal(t, 1, len(interfaceDecl.Methods))
				method := interfaceDecl.Methods[0]
				assert.True(t, method.ByReference, "Expected method to have reference return")
			},
		},
		{
			name: "Interface with mixed reference and non-reference methods",
			input: `<?php
interface DataInterface {
    public function &getByRef(): array;
    public function getNormal(): string;
    function &getDefault();
}
?>`,
			expectedErrors: 0,
			validate: func(t *testing.T, interfaceDecl *ast.InterfaceDeclaration) {
				assert.Equal(t, "DataInterface", interfaceDecl.Name.Name)
				assert.Equal(t, 3, len(interfaceDecl.Methods))
				
				// First method: with reference
				assert.Equal(t, "getByRef", interfaceDecl.Methods[0].Name.Name)
				assert.True(t, interfaceDecl.Methods[0].ByReference)
				assert.Equal(t, "array", interfaceDecl.Methods[0].ReturnType.Name)
				
				// Second method: without reference
				assert.Equal(t, "getNormal", interfaceDecl.Methods[1].Name.Name)
				assert.False(t, interfaceDecl.Methods[1].ByReference)
				assert.Equal(t, "string", interfaceDecl.Methods[1].ReturnType.Name)
				
				// Third method: with reference, no explicit visibility
				assert.Equal(t, "getDefault", interfaceDecl.Methods[2].Name.Name)
				assert.True(t, interfaceDecl.Methods[2].ByReference)
				assert.Equal(t, "public", interfaceDecl.Methods[2].Visibility) // Default to public
			},
		},
		{
			name: "Interface with complex parameter types and reference return",
			input: `<?php
interface ProcessorInterface {
    public function &process(array &$data, ?string $key = null): mixed;
    function &transform(int|string $value): array|object;
}
?>`,
			expectedErrors: 0,
			validate: func(t *testing.T, interfaceDecl *ast.InterfaceDeclaration) {
				assert.Equal(t, "ProcessorInterface", interfaceDecl.Name.Name)
				assert.Equal(t, 2, len(interfaceDecl.Methods))
				
				// First method
				method1 := interfaceDecl.Methods[0]
				assert.Equal(t, "process", method1.Name.Name)
				assert.True(t, method1.ByReference)
				assert.Equal(t, 2, len(method1.Parameters.Parameters))
				assert.Equal(t, "$data", method1.Parameters.Parameters[0].Name.(*ast.IdentifierNode).Name)
				assert.True(t, method1.Parameters.Parameters[0].ByReference)
				assert.Equal(t, "$key", method1.Parameters.Parameters[1].Name.(*ast.IdentifierNode).Name)
				assert.NotNil(t, method1.Parameters.Parameters[1].Type)
				assert.True(t, method1.Parameters.Parameters[1].Type.Nullable)
				assert.NotNil(t, method1.Parameters.Parameters[1].DefaultValue)
				
				// Second method
				method2 := interfaceDecl.Methods[1]
				assert.Equal(t, "transform", method2.Name.Name)
				assert.True(t, method2.ByReference)
				if method2.Parameters != nil {
					assert.Equal(t, 1, len(method2.Parameters.Parameters))
				} else {
					assert.Equal(t, 0, 1) // This will fail if expected 1 but got nil
				}
			},
		},
		{
			name: "Interface with only reference return methods",
			input: `<?php
interface ReferenceInterface {
    function &getData(): array;
    function &getObject(): object;
    function &getString(): string;
}
?>`,
			expectedErrors: 0,
			validate: func(t *testing.T, interfaceDecl *ast.InterfaceDeclaration) {
				assert.Equal(t, "ReferenceInterface", interfaceDecl.Name.Name)
				assert.Equal(t, 3, len(interfaceDecl.Methods))
				
				// All methods should have reference return
				for _, method := range interfaceDecl.Methods {
					assert.True(t, method.ByReference, "Expected all methods to have reference return")
					assert.NotNil(t, method.ReturnType, "Expected all methods to have return type")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			assert.Equal(t, tt.expectedErrors, len(p.Errors()), "Parser errors: %v", p.Errors())
			assert.NotNil(t, program)
			assert.Greater(t, len(program.Body), 0, "Expected at least one statement")

			// Check that we have an InterfaceDeclaration
			stmt := program.Body[0]
			interfaceDecl, ok := stmt.(*ast.InterfaceDeclaration)
			assert.True(t, ok, "Expected InterfaceDeclaration, got %T", stmt)
			assert.NotNil(t, interfaceDecl.Name, "Expected interface name to be set")

			// Run specific validation
			if tt.validate != nil {
				tt.validate(t, interfaceDecl)
			}
		})
	}
}