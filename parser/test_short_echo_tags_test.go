package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/lexer"
)

func TestParsing_ShortEchoTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, program *ast.Program)
	}{
		{
			name:  "simple short echo with string",
			input: `<?= "hello" ?>`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				stmt, ok := program.Body[0].(*ast.EchoStatement)
				require.True(t, ok, "Statement should be EchoStatement")
				require.Len(t, stmt.Arguments, 1)
				
				stringLit, ok := stmt.Arguments[0].(*ast.StringLiteral)
				require.True(t, ok, "Argument should be StringLiteral")
				assert.Equal(t, "hello", stringLit.Value)
			},
		},
		{
			name:  "short echo with variable",
			input: `<?= $name ?>`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				stmt, ok := program.Body[0].(*ast.EchoStatement)
				require.True(t, ok, "Statement should be EchoStatement")
				require.Len(t, stmt.Arguments, 1)
				
				variable, ok := stmt.Arguments[0].(*ast.Variable)
				require.True(t, ok, "Argument should be Variable")
				assert.Equal(t, "$name", variable.Name)
			},
		},
		{
			name:  "short echo with object property",
			input: `<?= $this->charset ?>`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				stmt, ok := program.Body[0].(*ast.EchoStatement)
				require.True(t, ok, "Statement should be EchoStatement")
				require.Len(t, stmt.Arguments, 1)
				
				objAccess, ok := stmt.Arguments[0].(*ast.PropertyAccessExpression)
				require.True(t, ok, "Argument should be PropertyAccessExpression")
				
				variable, ok := objAccess.Object.(*ast.Variable)
				require.True(t, ok, "Object should be Variable")
				assert.Equal(t, "$this", variable.Name)
				
				property, ok := objAccess.Property.(*ast.IdentifierNode)
				require.True(t, ok, "Property should be IdentifierNode")
				assert.Equal(t, "charset", property.Name)
			},
		},
		{
			name:  "short echo with semicolon",
			input: `<?= "hello"; ?>`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				stmt, ok := program.Body[0].(*ast.EchoStatement)
				require.True(t, ok, "Statement should be EchoStatement")
				require.Len(t, stmt.Arguments, 1)
				
				stringLit, ok := stmt.Arguments[0].(*ast.StringLiteral)
				require.True(t, ok, "Argument should be StringLiteral")
				assert.Equal(t, "hello", stringLit.Value)
			},
		},
		{
			name:  "short echo with complex expression",
			input: `<?= $user->getName() . " says hello"; ?>`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				stmt, ok := program.Body[0].(*ast.EchoStatement)
				require.True(t, ok, "Statement should be EchoStatement")
				require.Len(t, stmt.Arguments, 1)
				
				// Should be a binary expression for string concatenation
				binExpr, ok := stmt.Arguments[0].(*ast.BinaryExpression)
				require.True(t, ok, "Argument should be BinaryExpression")
				assert.Equal(t, ".", binExpr.Operator)
			},
		},
		{
			name:  "short echo mixed with HTML",
			input: `<meta charset="<?= $this->charset; ?>" />`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 3)
				
				// Second statement should be the Short echo
				echoStmt, ok := program.Body[1].(*ast.EchoStatement)
				require.True(t, ok, "Second statement should be EchoStatement")
				require.Len(t, echoStmt.Arguments, 1)
				
				// Verify it's parsing the property access correctly
				objAccess, ok := echoStmt.Arguments[0].(*ast.PropertyAccessExpression)
				require.True(t, ok, "Echo argument should be PropertyAccessExpression")
				
				variable, ok := objAccess.Object.(*ast.Variable)
				require.True(t, ok, "Object should be Variable")
				assert.Equal(t, "$this", variable.Name)
			},
		},
		{
			name:  "short echo with number",
			input: `<?= 42 ?>`,
			validate: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				stmt, ok := program.Body[0].(*ast.EchoStatement)
				require.True(t, ok, "Statement should be EchoStatement")
				require.Len(t, stmt.Arguments, 1)
				
				intLit, ok := stmt.Arguments[0].(*ast.NumberLiteral)
				require.True(t, ok, "Argument should be NumberLiteral")
				assert.Equal(t, "42", intLit.Value)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			
			checkParserErrors(t, p)
			require.NotNil(t, program)
			tt.validate(t, program)
		})
	}
}