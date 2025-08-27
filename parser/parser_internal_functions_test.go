package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/lexer"
)

func TestParsing_InternalFunctions(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(*testing.T, *ast.Program)
	}{
		{
			name:  "Isset with single variable",
			input: `<?php isset($x);`,
			check: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				exprStmt := program.Body[0].(*ast.ExpressionStatement)
				issetExpr := exprStmt.Expression.(*ast.IssetExpression)
				assert.Len(t, issetExpr.Arguments, 1)
				varExpr := issetExpr.Arguments[0].(*ast.Variable)
				assert.Equal(t, "$x", varExpr.Name)
			},
		},
		{
			name:  "Isset with multiple variables (should be AND connected)",
			input: `<?php isset($x, $y, $z);`,
			check: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				exprStmt := program.Body[0].(*ast.ExpressionStatement)
				
				// The result should be: isset($x) && isset($y) && isset($z)
				// So the top level should be a binary expression with &&
				binaryExpr := exprStmt.Expression.(*ast.BinaryExpression)
				assert.Equal(t, "&&", binaryExpr.Operator)
				
				// Left side should be another binary expression for isset($x) && isset($y)
				leftBinary := binaryExpr.Left.(*ast.BinaryExpression)
				assert.Equal(t, "&&", leftBinary.Operator)
				
				// Check the innermost isset($x)
				issetX := leftBinary.Left.(*ast.IssetExpression)
				assert.Len(t, issetX.Arguments, 1)
				varX := issetX.Arguments[0].(*ast.Variable)
				assert.Equal(t, "$x", varX.Name)
				
				// Check isset($y)
				issetY := leftBinary.Right.(*ast.IssetExpression)
				assert.Len(t, issetY.Arguments, 1)
				varY := issetY.Arguments[0].(*ast.Variable)
				assert.Equal(t, "$y", varY.Name)
				
				// Check isset($z) on the right side
				issetZ := binaryExpr.Right.(*ast.IssetExpression)
				assert.Len(t, issetZ.Arguments, 1)
				varZ := issetZ.Arguments[0].(*ast.Variable)
				assert.Equal(t, "$z", varZ.Name)
			},
		},
		{
			name:  "Empty function",
			input: `<?php empty($var);`,
			check: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				exprStmt := program.Body[0].(*ast.ExpressionStatement)
				emptyExpr := exprStmt.Expression.(*ast.EmptyExpression)
				varExpr := emptyExpr.Expression.(*ast.Variable)
				assert.Equal(t, "$var", varExpr.Name)
			},
		},
		{
			name:  "Include statement",
			input: `<?php include 'file.php';`,
			check: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				exprStmt := program.Body[0].(*ast.ExpressionStatement)
				includeExpr := exprStmt.Expression.(*ast.IncludeOrEvalExpression)
				assert.Equal(t, lexer.T_INCLUDE, includeExpr.Type)
				stringExpr := includeExpr.Expr.(*ast.StringLiteral)
				assert.Equal(t, "file.php", stringExpr.Value)
			},
		},
		{
			name:  "Require_once statement",
			input: `<?php require_once "config.php";`,
			check: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				exprStmt := program.Body[0].(*ast.ExpressionStatement)
				requireExpr := exprStmt.Expression.(*ast.IncludeOrEvalExpression)
				assert.Equal(t, lexer.T_REQUIRE_ONCE, requireExpr.Type)
				stringExpr := requireExpr.Expr.(*ast.StringLiteral)
				assert.Equal(t, "config.php", stringExpr.Value)
			},
		},
		{
			name:  "Eval statement",
			input: `<?php eval($code);`,
			check: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				exprStmt := program.Body[0].(*ast.ExpressionStatement)
				evalExpr := exprStmt.Expression.(*ast.EvalExpression)
				varExpr := evalExpr.Argument.(*ast.Variable)
				assert.Equal(t, "$code", varExpr.Name)
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

func TestParsing_SpaceshipOperator(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(*testing.T, *ast.Program)
	}{
		{
			name:  "Simple spaceship comparison",
			input: `<?php $result = $a <=> $b;`,
			check: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				exprStmt := program.Body[0].(*ast.ExpressionStatement)
				assignExpr := exprStmt.Expression.(*ast.AssignmentExpression)
				
				// Right side should be spaceship expression
				spaceshipExpr := assignExpr.Right.(*ast.BinaryExpression)
				assert.Equal(t, "<=>", spaceshipExpr.Operator)
				
				// Check left and right operands
				leftVar := spaceshipExpr.Left.(*ast.Variable)
				assert.Equal(t, "$a", leftVar.Name)
				
				rightVar := spaceshipExpr.Right.(*ast.Variable)
				assert.Equal(t, "$b", rightVar.Name)
			},
		},
		{
			name:  "Spaceship with numbers",
			input: `<?php $result = 1 <=> 2;`,
			check: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				exprStmt := program.Body[0].(*ast.ExpressionStatement)
				assignExpr := exprStmt.Expression.(*ast.AssignmentExpression)
				
				spaceshipExpr := assignExpr.Right.(*ast.BinaryExpression)
				assert.Equal(t, "<=>", spaceshipExpr.Operator)
				
				leftNum := spaceshipExpr.Left.(*ast.NumberLiteral)
				assert.Equal(t, "1", leftNum.Value)
				assert.Equal(t, "integer", leftNum.Kind)
				
				rightNum := spaceshipExpr.Right.(*ast.NumberLiteral)
				assert.Equal(t, "2", rightNum.Value)
				assert.Equal(t, "integer", rightNum.Kind)
			},
		},
		{
			name:  "Complex spaceship expression",
			input: `<?php $result = ($a + 1) <=> ($b * 2);`,
			check: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				exprStmt := program.Body[0].(*ast.ExpressionStatement)
				assignExpr := exprStmt.Expression.(*ast.AssignmentExpression)
				
				spaceshipExpr := assignExpr.Right.(*ast.BinaryExpression)
				assert.Equal(t, "<=>", spaceshipExpr.Operator)
				
				// Left side should be a binary expression ($a + 1)
				leftExpr := spaceshipExpr.Left.(*ast.BinaryExpression)
				assert.Equal(t, "+", leftExpr.Operator)
				
				// Right side should be a binary expression ($b * 2)
				rightExpr := spaceshipExpr.Right.(*ast.BinaryExpression)
				assert.Equal(t, "*", rightExpr.Operator)
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