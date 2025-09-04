package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/lexer"
)

func TestParsing_AttributesOnClasses(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedErrors int
		validate       func(t *testing.T, classExpr *ast.ClassExpression)
	}{
		{
			name: "Multiple attributes on readonly class",
			input: `<?php
#[AsyncListener]
#[Listener]
readonly class KnowledgeBaseFragmentSyncSubscriber implements ListenerInterface
{
}
?>`,
			expectedErrors: 0,
			validate: func(t *testing.T, classExpr *ast.ClassExpression) {
				assert.Equal(t, "KnowledgeBaseFragmentSyncSubscriber", getClassName(t, classExpr))
				assert.True(t, classExpr.ReadOnly, "Expected class to be readonly")
				assert.Equal(t, 1, len(classExpr.Implements), "Expected one interface")
				assert.Equal(t, "ListenerInterface", getInterfaceName(t, classExpr.Implements[0]))

				// Check attributes
				assert.Equal(t, 2, len(classExpr.Attributes), "Expected two attribute groups")

				// First attribute group - AsyncListener
				attr1 := classExpr.Attributes[0]
				assert.Equal(t, 1, len(attr1.Attributes), "Expected one attribute in first group")
				assert.Equal(t, "AsyncListener", getAttributeName(t, attr1.Attributes[0]))
				assert.Nil(t, attr1.Attributes[0].Arguments, "Expected no arguments")

				// Second attribute group - Listener
				attr2 := classExpr.Attributes[1]
				assert.Equal(t, 1, len(attr2.Attributes), "Expected one attribute in second group")
				assert.Equal(t, "Listener", getAttributeName(t, attr2.Attributes[0]))
				assert.Nil(t, attr2.Attributes[0].Arguments, "Expected no arguments")
			},
		},
		{
			name: "Multiple attributes on abstract class",
			input: `<?php
#[Entity]
#[Repository('user')]
abstract class AbstractClass
{
}
?>`,
			expectedErrors: 0,
			validate: func(t *testing.T, classExpr *ast.ClassExpression) {
				assert.Equal(t, "AbstractClass", getClassName(t, classExpr))
				assert.True(t, classExpr.Abstract, "Expected class to be abstract")

				// Check attributes
				assert.Equal(t, 2, len(classExpr.Attributes), "Expected two attribute groups")

				// First attribute - Entity
				attr1 := classExpr.Attributes[0]
				assert.Equal(t, "Entity", getAttributeName(t, attr1.Attributes[0]))

				// Second attribute - Repository with argument
				attr2 := classExpr.Attributes[1]
				assert.Equal(t, "Repository", getAttributeName(t, attr2.Attributes[0]))
				assert.NotNil(t, attr2.Attributes[0].Arguments, "Expected arguments for Repository")
			},
		},
		{
			name: "Multiple attributes on final class",
			input: `<?php
#[Controller]
#[Route('/api')]
final class FinalClass
{
}
?>`,
			expectedErrors: 0,
			validate: func(t *testing.T, classExpr *ast.ClassExpression) {
				assert.Equal(t, "FinalClass", getClassName(t, classExpr))
				assert.True(t, classExpr.Final, "Expected class to be final")

				// Check attributes
				assert.Equal(t, 2, len(classExpr.Attributes), "Expected two attribute groups")
				assert.Equal(t, "Controller", getAttributeName(t, classExpr.Attributes[0].Attributes[0]))
				assert.Equal(t, "Route", getAttributeName(t, classExpr.Attributes[1].Attributes[0]))
			},
		},
		{
			name: "Multiple attributes on regular class",
			input: `<?php
#[Service]
#[Tagged('logger')]
class RegularClass
{
}
?>`,
			expectedErrors: 0,
			validate: func(t *testing.T, classExpr *ast.ClassExpression) {
				assert.Equal(t, "RegularClass", getClassName(t, classExpr))
				assert.False(t, classExpr.ReadOnly, "Expected class not to be readonly")
				assert.False(t, classExpr.Abstract, "Expected class not to be abstract")
				assert.False(t, classExpr.Final, "Expected class not to be final")

				// Check attributes
				assert.Equal(t, 2, len(classExpr.Attributes), "Expected two attribute groups")
				assert.Equal(t, "Service", getAttributeName(t, classExpr.Attributes[0].Attributes[0]))
				assert.Equal(t, "Tagged", getAttributeName(t, classExpr.Attributes[1].Attributes[0]))
			},
		},
		{
			name: "Single attribute with multiple entries",
			input: `<?php
#[Component, Service, Tagged('multi')]
class MultiAttributeClass
{
}
?>`,
			expectedErrors: 0,
			validate: func(t *testing.T, classExpr *ast.ClassExpression) {
				assert.Equal(t, "MultiAttributeClass", getClassName(t, classExpr))

				// Check attributes - should be 1 group with 3 attributes
				assert.Equal(t, 1, len(classExpr.Attributes), "Expected one attribute group")

				attrGroup := classExpr.Attributes[0]
				assert.Equal(t, 3, len(attrGroup.Attributes), "Expected three attributes in group")

				assert.Equal(t, "Component", getAttributeName(t, attrGroup.Attributes[0]))
				assert.Equal(t, "Service", getAttributeName(t, attrGroup.Attributes[1]))
				assert.Equal(t, "Tagged", getAttributeName(t, attrGroup.Attributes[2]))
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

			// Extract the class from ExpressionStatement
			stmt := program.Body[0]
			exprStmt, ok := stmt.(*ast.ExpressionStatement)
			assert.True(t, ok, "Expected ExpressionStatement, got %T", stmt)

			classExpr, ok := exprStmt.Expression.(*ast.ClassExpression)
			assert.True(t, ok, "Expected ClassExpression, got %T", exprStmt.Expression)

			// Run specific validation
			if tt.validate != nil {
				tt.validate(t, classExpr)
			}
		})
	}
}

// Helper function to get attribute name from Attribute
func getAttributeName(t *testing.T, attr *ast.Attribute) string {
	return attr.Name.Name
}
