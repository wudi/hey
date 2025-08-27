package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/lexer"
)

func TestParsing_PipeOperator(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(*testing.T, *ast.Program)
	}{
		{
			name:  "Simple pipe operation",
			input: `<?php $result = $input |> strtoupper;`,
			check: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				exprStmt := program.Body[0].(*ast.ExpressionStatement)
				assignExpr := exprStmt.Expression.(*ast.AssignmentExpression)
				
				// Right side should be pipe expression
				pipeExpr := assignExpr.Right.(*ast.BinaryExpression)
				assert.Equal(t, "|>", pipeExpr.Operator)
				
				// Check left and right operands
				leftVar := pipeExpr.Left.(*ast.Variable)
				assert.Equal(t, "$input", leftVar.Name)
				
				rightFunc := pipeExpr.Right.(*ast.IdentifierNode)
				assert.Equal(t, "strtoupper", rightFunc.Name)
			},
		},
		{
			name:  "Chained pipe operations",
			input: `<?php $result = $input |> strtoupper |> trim;`,
			check: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				exprStmt := program.Body[0].(*ast.ExpressionStatement)
				assignExpr := exprStmt.Expression.(*ast.AssignmentExpression)
				
				// Should be left-associative: (($input |> strtoupper) |> trim)
				outerPipe := assignExpr.Right.(*ast.BinaryExpression)
				assert.Equal(t, "|>", outerPipe.Operator)
				
				// Left side should be another pipe expression
				innerPipe := outerPipe.Left.(*ast.BinaryExpression)
				assert.Equal(t, "|>", innerPipe.Operator)
				
				// Check the innermost operands
				inputVar := innerPipe.Left.(*ast.Variable)
				assert.Equal(t, "$input", inputVar.Name)
				
				strToUpperFunc := innerPipe.Right.(*ast.IdentifierNode)
				assert.Equal(t, "strtoupper", strToUpperFunc.Name)
				
				// Check the final function
				trimFunc := outerPipe.Right.(*ast.IdentifierNode)
				assert.Equal(t, "trim", trimFunc.Name)
			},
		},
		{
			name:  "Pipe with function call",
			input: `<?php $result = $input |> trim(_);`,
			check: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				exprStmt := program.Body[0].(*ast.ExpressionStatement)
				assignExpr := exprStmt.Expression.(*ast.AssignmentExpression)
				
				pipeExpr := assignExpr.Right.(*ast.BinaryExpression)
				assert.Equal(t, "|>", pipeExpr.Operator)
				
				// Left side should be variable
				leftVar := pipeExpr.Left.(*ast.Variable)
				assert.Equal(t, "$input", leftVar.Name)
				
				// Right side should be function call
				rightCall := pipeExpr.Right.(*ast.CallExpression)
				funcName := rightCall.Callee.(*ast.IdentifierNode)
				assert.Equal(t, "trim", funcName.Name)
				assert.Len(t, rightCall.Arguments, 1)
				
				// Argument should be placeholder _
				argIdent := rightCall.Arguments[0].(*ast.IdentifierNode)
				assert.Equal(t, "_", argIdent.Name)
			},
		},
		{
			name:  "Complex pipe chain with function calls",
			input: `<?php $result = $data |> array_filter(_) |> array_values(_);`,
			check: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				exprStmt := program.Body[0].(*ast.ExpressionStatement)
				assignExpr := exprStmt.Expression.(*ast.AssignmentExpression)
				
				// Should be chained pipe operations
				outerPipe := assignExpr.Right.(*ast.BinaryExpression)
				assert.Equal(t, "|>", outerPipe.Operator)
				
				innerPipe := outerPipe.Left.(*ast.BinaryExpression)
				assert.Equal(t, "|>", innerPipe.Operator)
				
				// Check the data variable at the start
				dataVar := innerPipe.Left.(*ast.Variable)
				assert.Equal(t, "$data", dataVar.Name)
				
				// Check array_filter call
				arrayFilterCall := innerPipe.Right.(*ast.CallExpression)
				filterFunc := arrayFilterCall.Callee.(*ast.IdentifierNode)
				assert.Equal(t, "array_filter", filterFunc.Name)
				
				// Check array_values call
				arrayValuesCall := outerPipe.Right.(*ast.CallExpression)
				valuesFunc := arrayValuesCall.Callee.(*ast.IdentifierNode)
				assert.Equal(t, "array_values", valuesFunc.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			// Check for parser errors
			if len(p.Errors()) > 0 {
				t.Errorf("Parser errors: %v", p.Errors())
				return
			}

			assert.NotNil(t, program)
			tt.check(t, program)
		})
	}
}