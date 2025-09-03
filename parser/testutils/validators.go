package testutils

import (
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wudi/php-parser/ast"
)

// 高级验证函数集合，用于复杂的AST结构验证

// ValidateEcho 验证echo语句
func ValidateEcho(expectedArgCount int, argValidators ...func(ast.Node, *testing.T)) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		echoStmt := assertions.AssertEchoStatement(body[0], expectedArgCount)
		
		// 验证每个参数
		for i, validator := range argValidators {
			if i < len(echoStmt.Arguments.Arguments) {
				validator(echoStmt.Arguments.Arguments[i], ctx.T)
			}
		}
	}
}

// ValidateStringArg 创建字符串参数验证器
func ValidateStringArg(expectedValue, expectedRaw string) func(ast.Node, *testing.T) {
	return func(node ast.Node, t *testing.T) {
		assertions := NewASTAssertions(t)
		assertions.AssertStringLiteral(node, expectedValue, expectedRaw)
	}
}

// ValidateNumberArg 创建数字参数验证器
func ValidateNumberArg(expectedValue string) func(ast.Node, *testing.T) {
	return func(node ast.Node, t *testing.T) {
		assertions := NewASTAssertions(t)
		assertions.AssertNumberLiteral(node, expectedValue)
	}
}

// ValidateBinaryOperation 验证二元运算表达式
func ValidateBinaryOperation(operator string, leftValidator, rightValidator func(ast.Node, *testing.T)) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
		binExpr := assertions.AssertBinaryExpression(assignment.Right, operator)
		
		if leftValidator != nil {
			leftValidator(binExpr.Left, ctx.T)
		}
		if rightValidator != nil {
			rightValidator(binExpr.Right, ctx.T)
		}
	}
}

// ValidateFunction 验证函数声明
func ValidateFunction(expectedName string, expectedParamCount int, validators ...func(*ast.FunctionDeclaration, *testing.T)) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		funcDecl, ok := body[0].(*ast.FunctionDeclaration)
		assert.True(ctx.T, ok, "Statement should be FunctionDeclaration, got %T", body[0])
		
		nameIdent, ok := funcDecl.Name.(*ast.IdentifierNode)
		assert.True(ctx.T, ok, "Function name should be IdentifierNode")
		assert.Equal(ctx.T, expectedName, nameIdent.Name)
		
		if funcDecl.Parameters != nil {
			assert.Len(ctx.T, funcDecl.Parameters.Parameters, expectedParamCount)
		} else {
			assert.Equal(ctx.T, 0, expectedParamCount, "Expected no parameters but Parameters is nil")
		}
		
		// 运行自定义验证器
		for _, validator := range validators {
			validator(funcDecl, ctx.T)
		}
	}
}

// ValidateClass 验证类声明
func ValidateClass(expectedName string, validators ...func(*ast.ClassExpression, *testing.T)) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		classExpr := assertions.AssertClass(exprStmt.Expression, expectedName)
		
		// 运行自定义验证器
		for _, validator := range validators {
			validator(classExpr, ctx.T)
		}
	}
}

// ValidateClassMethod 创建类方法验证器
func ValidateClassMethod(expectedName, expectedVisibility string) func(*ast.ClassExpression, *testing.T) {
	return func(classExpr *ast.ClassExpression, t *testing.T) {
		// 查找方法
		var foundMethod *ast.FunctionDeclaration
		for _, stmt := range classExpr.Body {
			if funcDecl, ok := stmt.(*ast.FunctionDeclaration); ok {
				if nameIdent, ok := funcDecl.Name.(*ast.IdentifierNode); ok && nameIdent.Name == expectedName {
					foundMethod = funcDecl
					break
				}
			}
		}
		
		assert.NotNil(t, foundMethod, "Method %s not found in class", expectedName)
		if foundMethod != nil {
			assert.Equal(t, expectedVisibility, foundMethod.Visibility, "Method visibility mismatch")
		}
	}
}

// ValidateClassConstant 创建类常量验证器
func ValidateClassConstant(expectedName string, expectedVisibility string) func(*ast.ClassExpression, *testing.T) {
	return func(classExpr *ast.ClassExpression, t *testing.T) {
		// 查找常量声明
		var foundConstant *ast.ClassConstantDeclaration
		for _, stmt := range classExpr.Body {
			if constDecl, ok := stmt.(*ast.ClassConstantDeclaration); ok {
				for _, constant := range constDecl.Constants {
					if nameIdent, ok := constant.Name.(*ast.IdentifierNode); ok && nameIdent.Name == expectedName {
						foundConstant = constDecl
						break
					}
				}
				if foundConstant != nil {
					break
				}
			}
		}
		
		assert.NotNil(t, foundConstant, "Constant %s not found in class", expectedName)
		if foundConstant != nil {
			assert.Equal(t, expectedVisibility, foundConstant.Visibility, "Constant visibility mismatch")
		}
	}
}

// ValidateProperty 创建类属性验证器
func ValidateProperty(expectedName string, expectedVisibility string) func(*ast.ClassExpression, *testing.T) {
	return func(classExpr *ast.ClassExpression, t *testing.T) {
		// 查找属性声明
		var foundProperty *ast.PropertyDeclaration
		for _, stmt := range classExpr.Body {
			if propDecl, ok := stmt.(*ast.PropertyDeclaration); ok {
				// PropertyDeclaration 有单个Name字段，不是Properties数组
				if propDecl.Name == expectedName {
					foundProperty = propDecl
					break
				}
			}
		}
		
		assert.NotNil(t, foundProperty, "Property %s not found in class", expectedName)
		if foundProperty != nil {
			assert.Equal(t, expectedVisibility, foundProperty.Visibility, "Property visibility mismatch")
		}
	}
}

// ValidateControlFlow 验证控制流语句
func ValidateControlFlow(expectedType string, validators ...func(ast.Statement, *testing.T)) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		stmt := body[0]
		
		switch expectedType {
		case "if":
			_, ok := stmt.(*ast.IfStatement)
			assert.True(ctx.T, ok, "Statement should be IfStatement")
		case "while":
			_, ok := stmt.(*ast.WhileStatement)
			assert.True(ctx.T, ok, "Statement should be WhileStatement")
		case "for":
			_, ok := stmt.(*ast.ForStatement)
			assert.True(ctx.T, ok, "Statement should be ForStatement")
		case "foreach":
			_, ok := stmt.(*ast.ForeachStatement)
			assert.True(ctx.T, ok, "Statement should be ForeachStatement")
		case "switch":
			_, ok := stmt.(*ast.SwitchStatement)
			assert.True(ctx.T, ok, "Statement should be SwitchStatement")
		case "try":
			_, ok := stmt.(*ast.TryStatement)
			assert.True(ctx.T, ok, "Statement should be TryStatement")
		}
		
		// 运行自定义验证器
		for _, validator := range validators {
			validator(stmt, ctx.T)
		}
	}
}

// 控制流验证函数

// ValidateIfStatement 验证if语句
func ValidateIfStatement(testValidator ValidationFunc, consequentValidators ...ValidationFunc) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		// 支持两种类型if语句
		if ifStmt, ok := body[0].(*ast.IfStatement); ok {
			// 常规if语句
			// 验证条件
			if testValidator != nil {
				testCtx := &TestContext{
					T: ctx.T,
					Program: &ast.Program{Body: []ast.Statement{&ast.ExpressionStatement{Expression: ifStmt.Test}}},
				}
				testValidator(testCtx)
			}
			
			// 验证consequent块
			for i, validator := range consequentValidators {
				if i < len(ifStmt.Consequent) {
					stmtCtx := &TestContext{
						T: ctx.T,
						Program: &ast.Program{Body: []ast.Statement{ifStmt.Consequent[i]}},
					}
					validator(stmtCtx)
				}
			}
		} else if altIfStmt, ok := body[0].(*ast.AlternativeIfStatement); ok {
			// 替代if语句
			// 验证条件
			if testValidator != nil {
				testCtx := &TestContext{
					T: ctx.T,
					Program: &ast.Program{Body: []ast.Statement{&ast.ExpressionStatement{Expression: altIfStmt.Condition}}},
				}
				testValidator(testCtx)
			}
			
			// 验证consequent块
			for i, validator := range consequentValidators {
				if i < len(altIfStmt.Then) {
					stmtCtx := &TestContext{
						T: ctx.T,
						Program: &ast.Program{Body: []ast.Statement{altIfStmt.Then[i]}},
					}
					validator(stmtCtx)
				}
			}
		} else {
			assert.Fail(ctx.T, "Statement should be IfStatement or AlternativeIfStatement, got %T", body[0])
		}
	}
}

// ValidateIfElseStatement 验证if-else语句
func ValidateIfElseStatement(testValidator ValidationFunc, consequentValidator ValidationFunc, alternateValidator ValidationFunc) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		ifStmt, ok := body[0].(*ast.IfStatement)
		assert.True(ctx.T, ok, "Statement should be IfStatement, got %T", body[0])
		
		// 验证条件
		if testValidator != nil {
			testCtx := &TestContext{
				T: ctx.T,
				Program: &ast.Program{Body: []ast.Statement{&ast.ExpressionStatement{Expression: ifStmt.Test}}},
			}
			testValidator(testCtx)
		}
		
		// 验证consequent
		if consequentValidator != nil && len(ifStmt.Consequent) > 0 {
			stmtCtx := &TestContext{
				T: ctx.T,
				Program: &ast.Program{Body: []ast.Statement{ifStmt.Consequent[0]}},
			}
			consequentValidator(stmtCtx)
		}
		
		// 验证alternate
		if alternateValidator != nil && len(ifStmt.Alternate) > 0 {
			stmtCtx := &TestContext{
				T: ctx.T,
				Program: &ast.Program{Body: []ast.Statement{ifStmt.Alternate[0]}},
			}
			alternateValidator(stmtCtx)
		}
	}
}

// ValidateWhileStatement 验证while语句
func ValidateWhileStatement(testValidator ValidationFunc, bodyValidators ...ValidationFunc) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		// 支持两种类型while语句
		if whileStmt, ok := body[0].(*ast.WhileStatement); ok {
			// 常规while语句
			// 验证条件
			if testValidator != nil {
				testCtx := &TestContext{
					T: ctx.T,
					Program: &ast.Program{Body: []ast.Statement{&ast.ExpressionStatement{Expression: whileStmt.Test}}},
				}
				testValidator(testCtx)
			}
			
			// 验证循环体
			for i, validator := range bodyValidators {
				if i < len(whileStmt.Body) {
					stmtCtx := &TestContext{
						T: ctx.T,
						Program: &ast.Program{Body: []ast.Statement{whileStmt.Body[i]}},
					}
					validator(stmtCtx)
				}
			}
		} else if altWhileStmt, ok := body[0].(*ast.AlternativeWhileStatement); ok {
			// 替代while语句
			// 验证条件
			if testValidator != nil {
				testCtx := &TestContext{
					T: ctx.T,
					Program: &ast.Program{Body: []ast.Statement{&ast.ExpressionStatement{Expression: altWhileStmt.Condition}}},
				}
				testValidator(testCtx)
			}
			
			// 验证循环体
			for i, validator := range bodyValidators {
				if i < len(altWhileStmt.Body) {
					stmtCtx := &TestContext{
						T: ctx.T,
						Program: &ast.Program{Body: []ast.Statement{altWhileStmt.Body[i]}},
					}
					validator(stmtCtx)
				}
			}
		} else {
			assert.Fail(ctx.T, "Statement should be WhileStatement or AlternativeWhileStatement, got %T", body[0])
		}
	}
}

// ValidateForStatement 验证for语句
func ValidateForStatement(initValidator, testValidator, updateValidator, bodyValidator ValidationFunc) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		// 支持两种类型for语句
		if forStmt, ok := body[0].(*ast.ForStatement); ok {
			// 常规for语句
			// 验证初始化
			if initValidator != nil && forStmt.Init != nil {
				initCtx := &TestContext{
					T: ctx.T,
					Program: &ast.Program{Body: []ast.Statement{&ast.ExpressionStatement{Expression: forStmt.Init}}},
				}
				initValidator(initCtx)
			}
			
			// 验证条件
			if testValidator != nil && forStmt.Test != nil {
				testCtx := &TestContext{
					T: ctx.T,
					Program: &ast.Program{Body: []ast.Statement{&ast.ExpressionStatement{Expression: forStmt.Test}}},
				}
				testValidator(testCtx)
			}
			
			// 验证更新
			if updateValidator != nil && forStmt.Update != nil {
				updateCtx := &TestContext{
					T: ctx.T,
					Program: &ast.Program{Body: []ast.Statement{&ast.ExpressionStatement{Expression: forStmt.Update}}},
				}
				updateValidator(updateCtx)
			}
			
			// 验证循环体
			if bodyValidator != nil && len(forStmt.Body) > 0 {
				bodyCtx := &TestContext{
					T: ctx.T,
					Program: &ast.Program{Body: []ast.Statement{forStmt.Body[0]}},
				}
				bodyValidator(bodyCtx)
			}
		} else if altForStmt, ok := body[0].(*ast.AlternativeForStatement); ok {
			// 替代for语句
			// 验证初始化
			if initValidator != nil && len(altForStmt.Init) > 0 && altForStmt.Init[0] != nil {
				initCtx := &TestContext{
					T: ctx.T,
					Program: &ast.Program{Body: []ast.Statement{&ast.ExpressionStatement{Expression: altForStmt.Init[0]}}},
				}
				initValidator(initCtx)
			}
			
			// 验证条件
			if testValidator != nil && len(altForStmt.Condition) > 0 && altForStmt.Condition[0] != nil {
				testCtx := &TestContext{
					T: ctx.T,
					Program: &ast.Program{Body: []ast.Statement{&ast.ExpressionStatement{Expression: altForStmt.Condition[0]}}},
				}
				testValidator(testCtx)
			}
			
			// 验证更新
			if updateValidator != nil && len(altForStmt.Update) > 0 && altForStmt.Update[0] != nil {
				updateCtx := &TestContext{
					T: ctx.T,
					Program: &ast.Program{Body: []ast.Statement{&ast.ExpressionStatement{Expression: altForStmt.Update[0]}}},
				}
				updateValidator(updateCtx)
			}
			
			// 验证循环体
			if bodyValidator != nil && len(altForStmt.Body) > 0 {
				bodyCtx := &TestContext{
					T: ctx.T,
					Program: &ast.Program{Body: []ast.Statement{altForStmt.Body[0]}},
				}
				bodyValidator(bodyCtx)
			}
		} else {
			assert.Fail(ctx.T, "Statement should be ForStatement or AlternativeForStatement, got %T", body[0])
		}
	}
}

// ValidateForeachStatement 验证foreach语句
func ValidateForeachStatement(iterableVar, keyVar, valueVar string, bodyValidators ...ValidationFunc) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		// 检查是否是常规foreach或替代foreach
		if foreachStmt, ok := body[0].(*ast.ForeachStatement); ok {
			// 常规foreach - Body是单个Statement
			// 验证可迭代变量
			if iterableVar != "" {
				iterableVariable, ok := foreachStmt.Iterable.(*ast.Variable)
				assert.True(ctx.T, ok, "Iterable should be Variable")
				assert.Equal(ctx.T, iterableVar, iterableVariable.Name)
			}
			
			// 验证值变量
			if valueVar != "" {
				valueVariable, ok := foreachStmt.Value.(*ast.Variable)
				assert.True(ctx.T, ok, "Value should be Variable")
				assert.Equal(ctx.T, valueVar, valueVariable.Name)
			}
			
			// 验证键变量（如果存在）
			if keyVar != "" && foreachStmt.Key != nil {
				keyVariable, ok := foreachStmt.Key.(*ast.Variable)
				assert.True(ctx.T, ok, "Key should be Variable")
				assert.Equal(ctx.T, keyVar, keyVariable.Name)
			}
			
			// 验证循环体（单个Statement）
			if len(bodyValidators) > 0 && foreachStmt.Body != nil {
				stmtCtx := &TestContext{
					T: ctx.T,
					Program: &ast.Program{Body: []ast.Statement{foreachStmt.Body}},
				}
				bodyValidators[0](stmtCtx)
			}
		} else if altForeachStmt, ok := body[0].(*ast.AlternativeForeachStatement); ok {
			// 替代foreach - Body是[]Statement
			// 验证可迭代变量
			if iterableVar != "" {
				iterableVariable, ok := altForeachStmt.Iterable.(*ast.Variable)
				assert.True(ctx.T, ok, "Iterable should be Variable")
				assert.Equal(ctx.T, iterableVar, iterableVariable.Name)
			}
			
			// 验证值变量
			if valueVar != "" {
				valueVariable, ok := altForeachStmt.Value.(*ast.Variable)
				assert.True(ctx.T, ok, "Value should be Variable")
				assert.Equal(ctx.T, valueVar, valueVariable.Name)
			}
			
			// 验证键变量（如果存在）
			if keyVar != "" && altForeachStmt.Key != nil {
				keyVariable, ok := altForeachStmt.Key.(*ast.Variable)
				assert.True(ctx.T, ok, "Key should be Variable")
				assert.Equal(ctx.T, keyVar, keyVariable.Name)
			}
			
			// 验证循环体（[]Statement）
			for i, validator := range bodyValidators {
				if i < len(altForeachStmt.Body) {
					stmtCtx := &TestContext{
						T: ctx.T,
						Program: &ast.Program{Body: []ast.Statement{altForeachStmt.Body[i]}},
					}
					validator(stmtCtx)
				}
			}
		} else {
			assert.Fail(ctx.T, "Statement should be ForeachStatement or AlternativeForeachStatement, got %T", body[0])
		}
	}
}

// 类型定义
type CatchClause struct {
	ExceptionType string
	VariableName  string
	BodyValidator ValidationFunc
}

type SwitchCase struct {
	Value      string
	Validators []ValidationFunc
	IsDefault  bool
}

// ValidateTryCatchStatement 验证try-catch语句
func ValidateTryCatchStatement(tryValidator ValidationFunc, catchValidator CatchClause) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		tryStmt, ok := body[0].(*ast.TryStatement)
		assert.True(ctx.T, ok, "Statement should be TryStatement, got %T", body[0])
		
		// 验证try块
		if tryValidator != nil && len(tryStmt.Body) > 0 {
			tryCtx := &TestContext{
				T: ctx.T,
				Program: &ast.Program{Body: []ast.Statement{tryStmt.Body[0]}},
			}
			tryValidator(tryCtx)
		}
		
		// 验证catch块
		if len(tryStmt.CatchClauses) > 0 && catchValidator.BodyValidator != nil {
			catchClause := tryStmt.CatchClauses[0]
			
			// 验证异常类型 - 使用Types[0]而不是Type
			if catchValidator.ExceptionType != "" && len(catchClause.Types) > 0 {
				typeIdent, ok := catchClause.Types[0].(*ast.IdentifierNode)
				assert.True(ctx.T, ok, "Exception type should be IdentifierNode")
				assert.Equal(ctx.T, catchValidator.ExceptionType, typeIdent.Name)
			}
			
			// 验证异常变量 - 使用Parameter而不是Variable
			if catchValidator.VariableName != "" {
				catchVar, ok := catchClause.Parameter.(*ast.Variable)
				assert.True(ctx.T, ok, "Exception parameter should be Variable")
				assert.Equal(ctx.T, catchValidator.VariableName, catchVar.Name)
			}
			
			// 验证catch体
			if len(catchClause.Body) > 0 {
				catchCtx := &TestContext{
					T: ctx.T,
					Program: &ast.Program{Body: []ast.Statement{catchClause.Body[0]}},
				}
				catchValidator.BodyValidator(catchCtx)
			}
		}
	}
}

// ValidateTryCatchFinallyStatement 验证try-catch-finally语句
func ValidateTryCatchFinallyStatement(tryValidator ValidationFunc, catchValidator CatchClause, finallyValidator ValidationFunc) ValidationFunc {
	return func(ctx *TestContext) {
		// 首先执行try-catch验证
		tryCatchValidator := ValidateTryCatchStatement(tryValidator, catchValidator)
		tryCatchValidator(ctx)
		
		// 然后验证finally块
		if finallyValidator != nil {
			assertions := NewASTAssertions(ctx.T)
			body := assertions.AssertProgramBody(ctx.Program, 1)
			
			tryStmt, ok := body[0].(*ast.TryStatement)
			assert.True(ctx.T, ok, "Statement should be TryStatement")
			
			// 使用FinallyBlock而不是Finally.Body
			if len(tryStmt.FinallyBlock) > 0 {
				finallyCtx := &TestContext{
					T: ctx.T,
					Program: &ast.Program{Body: []ast.Statement{tryStmt.FinallyBlock[0]}},
				}
				finallyValidator(finallyCtx)
			}
		}
	}
}

// ValidateTryMultipleCatchStatement 验证多个catch子句的try语句
func ValidateTryMultipleCatchStatement(tryValidator ValidationFunc, catchValidators []CatchClause) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		tryStmt, ok := body[0].(*ast.TryStatement)
		assert.True(ctx.T, ok, "Statement should be TryStatement, got %T", body[0])
		
		// 验证try块
		if tryValidator != nil && len(tryStmt.Body) > 0 {
			tryCtx := &TestContext{
				T: ctx.T,
				Program: &ast.Program{Body: []ast.Statement{tryStmt.Body[0]}},
			}
			tryValidator(tryCtx)
		}
		
		// 验证所有catch子句
		assert.Len(ctx.T, tryStmt.CatchClauses, len(catchValidators), "Catch clause count mismatch")
		
		for i, catchValidator := range catchValidators {
			if i < len(tryStmt.CatchClauses) {
				catchClause := tryStmt.CatchClauses[i]
				
				// 验证异常类型 - 使用Types[0]而不是Type
				if catchValidator.ExceptionType != "" && len(catchClause.Types) > 0 {
					typeIdent, ok := catchClause.Types[0].(*ast.IdentifierNode)
					assert.True(ctx.T, ok, "Exception type should be IdentifierNode")
					assert.Equal(ctx.T, catchValidator.ExceptionType, typeIdent.Name)
				}
				
				// 验证异常变量
				if catchValidator.VariableName != "" {
					catchVar, ok := catchClause.Parameter.(*ast.Variable)
					assert.True(ctx.T, ok, "Exception variable should be Variable")
					assert.Equal(ctx.T, catchValidator.VariableName, catchVar.Name)
				}
				
				// 验证catch体
				if catchValidator.BodyValidator != nil && len(catchClause.Body) > 0 {
					catchCtx := &TestContext{
						T: ctx.T,
						Program: &ast.Program{Body: []ast.Statement{catchClause.Body[0]}},
					}
					catchValidator.BodyValidator(catchCtx)
				}
			}
		}
	}
}

// ValidateSwitchStatement 验证switch语句
func ValidateSwitchStatement(discriminantVar string, cases []SwitchCase) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		switchStmt, ok := body[0].(*ast.SwitchStatement)
		assert.True(ctx.T, ok, "Statement should be SwitchStatement, got %T", body[0])
		
		// 验证discriminant
		if discriminantVar != "" {
			discriminantVariable, ok := switchStmt.Discriminant.(*ast.Variable)
			assert.True(ctx.T, ok, "Discriminant should be Variable")
			assert.Equal(ctx.T, discriminantVar, discriminantVariable.Name)
		}
		
		// 验证case子句
		assert.Len(ctx.T, switchStmt.Cases, len(cases), "Switch case count mismatch")
		
		for i, expectedCase := range cases {
			if i < len(switchStmt.Cases) {
				caseStmt := switchStmt.Cases[i]
				
				if expectedCase.IsDefault {
					assert.Nil(ctx.T, caseStmt.Test, "Default case should have nil test")
				} else {
					// 验证case值
					assert.NotNil(ctx.T, caseStmt.Test, "Case should have test value")
				}
				
				// 验证case体
				for j, validator := range expectedCase.Validators {
					if j < len(caseStmt.Body) {
						caseCtx := &TestContext{
							T: ctx.T,
							Program: &ast.Program{Body: []ast.Statement{caseStmt.Body[j]}},
						}
						validator(caseCtx)
					}
				}
			}
		}
	}
}

// 辅助验证函数

// ValidateBinaryExpression 创建二元表达式验证器
func ValidateBinaryExpression(leftVar, operator, rightValue string) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		binExpr := assertions.AssertBinaryExpression(exprStmt.Expression, operator)
		
		// 验证左侧变量
		if leftVar != "" {
			leftVariable, ok := binExpr.Left.(*ast.Variable)
			assert.True(ctx.T, ok, "Left operand should be Variable")
			assert.Equal(ctx.T, leftVar, leftVariable.Name)
		}
		
		// 验证右侧值
		if rightValue != "" {
			// 尝试作为数字
			if rightNum, ok := binExpr.Right.(*ast.NumberLiteral); ok {
				assert.Equal(ctx.T, rightValue, rightNum.Value)
			} else if rightStr, ok := binExpr.Right.(*ast.StringLiteral); ok {
				assert.Equal(ctx.T, rightValue, rightStr.Raw)
			}
		}
	}
}

// ValidatePostfixExpression 创建后缀表达式验证器 
func ValidatePostfixExpression(varName, operator string) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		unaryExpr, ok := exprStmt.Expression.(*ast.UnaryExpression)
		assert.True(ctx.T, ok, "Expression should be PostfixExpression, got %T", exprStmt.Expression)
		
		assert.Equal(ctx.T, operator, unaryExpr.Operator)
		assert.False(ctx.T, unaryExpr.Prefix, "Should be postfix unary expression")
		
		if varName != "" {
			operandVar, ok := unaryExpr.Operand.(*ast.Variable)
			assert.True(ctx.T, ok, "Operand should be Variable")
			assert.Equal(ctx.T, varName, operandVar.Name)
		}
	}
}


// ValidateEchoVariable 创建echo变量验证器
func ValidateEchoVariable(varName string) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		echoStmt := assertions.AssertEchoStatement(body[0], 1)
		variable, ok := echoStmt.Arguments.Arguments[0].(*ast.Variable)
		assert.True(ctx.T, ok, "Echo argument should be Variable")
		assert.Equal(ctx.T, varName, variable.Name)
	}
}

// ValidateEchoArgs 创建echo语句验证器
func ValidateEchoArgs(args []string) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		echoStmt := assertions.AssertEchoStatement(body[0], len(args))
		
		for i, expectedArg := range args {
			if i < len(echoStmt.Arguments.Arguments) {
				arg := echoStmt.Arguments.Arguments[i]
				
				if stringLit, ok := arg.(*ast.StringLiteral); ok {
					assert.Equal(ctx.T, expectedArg, stringLit.Raw)
				} else if variable, ok := arg.(*ast.Variable); ok {
					assert.Equal(ctx.T, expectedArg, variable.Name)
				}
			}
		}
	}
}

// ValidateFunctionCall 创建函数调用验证器
func ValidateFunctionCall(funcName string, args ...string) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		callExpr, ok := exprStmt.Expression.(*ast.CallExpression)
		assert.True(ctx.T, ok, "Expression should be CallExpression, got %T", exprStmt.Expression)
		
		funcIdent, ok := callExpr.Callee.(*ast.IdentifierNode)
		assert.True(ctx.T, ok, "Callee should be IdentifierNode")
		assert.Equal(ctx.T, funcName, funcIdent.Name)
		
		if callExpr.Arguments != nil {
			assert.Len(ctx.T, callExpr.Arguments.Arguments, len(args))
		} else {
			assert.Equal(ctx.T, 0, len(args), "Expected no arguments but got %d", len(args))
		}
	}
}

// ValidateAssignmentExpression 创建赋值验证器
func ValidateAssignmentExpression(varName, value string) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
		
		// 验证左侧变量
		if varName != "" {
			leftVar, ok := assignment.Left.(*ast.Variable)
			assert.True(ctx.T, ok, "Left side should be Variable")
			assert.Equal(ctx.T, varName, leftVar.Name)
		}
		
		// 验证右侧值
		if value != "" {
			// 尝试作为数字
			if numberLit, ok := assignment.Right.(*ast.NumberLiteral); ok {
				assert.Equal(ctx.T, value, numberLit.Value)
			} else if stringLit, ok := assignment.Right.(*ast.StringLiteral); ok {
				assert.Equal(ctx.T, value, stringLit.Raw)
			} else if callExpr, ok := assignment.Right.(*ast.CallExpression); ok {
				// 函数调用情况
				if funcIdent, ok := callExpr.Callee.(*ast.IdentifierNode); ok {
					assert.Equal(ctx.T, value, funcIdent.Name+"()")
				}
			}
		}
	}
}

// ValidateBreakStatement 创建break语句验证器
func ValidateBreakStatement() ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		_, ok := body[0].(*ast.BreakStatement)
		assert.True(ctx.T, ok, "Statement should be BreakStatement, got %T", body[0])
	}
}

// ValidateCatchClause 创建catch子句验证器
func ValidateCatchClause(exceptionType, varName string, bodyValidator ValidationFunc) CatchClause {
	return CatchClause{
		ExceptionType: exceptionType,
		VariableName:  varName,
		BodyValidator: bodyValidator,
	}
}

// ValidateSwitchCase 创建switch case验证器
func ValidateSwitchCase(value string, validators ...ValidationFunc) SwitchCase {
	return SwitchCase{
		Value:      value,
		Validators: validators,
		IsDefault:  false,
	}
}

// ValidateDefaultCase 创建default case验证器
func ValidateDefaultCase(validators ...ValidationFunc) SwitchCase {
	return SwitchCase{
		Validators: validators,
		IsDefault:  true,
	}
}

// ValidateVariable 创建变量验证器
func ValidateVariableExpression(varName string) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		variable, ok := exprStmt.Expression.(*ast.Variable)
		assert.True(ctx.T, ok, "Expression should be Variable, got %T", exprStmt.Expression)
		assert.Equal(ctx.T, varName, variable.Name)
	}
}

// 数组和字符串验证函数

// ArrayElement 表示数组元素
type ArrayElement struct {
	Key       string // 数组键，空则为索引数组
	Value     string // 数组值
	IsNumeric bool   // 是否为数值
}

// StringInterpolation 表示字符串插值
type StringInterpolation struct {
	Parts []InterpolationPart
}

// InterpolationPart 表示插值部分
type InterpolationPart struct {
	Text       string // 字符串文本
	Variable   string // 变量名
	HasBraces  bool   // 是否有花括号
}

// ValidateArrayAssignment 验证数组赋值
func ValidateArrayAssignment(varName string, elements []ArrayElement) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
		
		// 验证左侧变量
		leftVar, ok := assignment.Left.(*ast.Variable)
		assert.True(ctx.T, ok, "Left side should be Variable")
		assert.Equal(ctx.T, varName, leftVar.Name)
		
		// 验证右侧数组
		arrayExpr, ok := assignment.Right.(*ast.ArrayExpression)
		assert.True(ctx.T, ok, "Right side should be ArrayExpression, got %T", assignment.Right)
		
		assert.Len(ctx.T, arrayExpr.Elements, len(elements), "Array element count mismatch")
		
		for i, expectedElement := range elements {
			if i >= len(arrayExpr.Elements) {
				break
			}
			
			element := arrayExpr.Elements[i]
			
			// 检查是否为键值对
			if arrayElement, ok := element.(*ast.ArrayElementExpression); ok {
				// 验证键
				if expectedElement.Key != "" {
					assert.NotNil(ctx.T, arrayElement.Key, "Array element should have key")
					if keyStr, ok := arrayElement.Key.(*ast.StringLiteral); ok {
						assert.Equal(ctx.T, expectedElement.Key, keyStr.Raw)
					} else if keyNum, ok := arrayElement.Key.(*ast.NumberLiteral); ok {
						assert.Equal(ctx.T, expectedElement.Key, keyNum.Value)
					}
				} else {
					assert.Nil(ctx.T, arrayElement.Key, "Array element should not have key")
				}
				
				// 验证值
				if expectedElement.IsNumeric {
					numVal, ok := arrayElement.Value.(*ast.NumberLiteral)
					assert.True(ctx.T, ok, "Array element value should be NumberLiteral")
					assert.Equal(ctx.T, expectedElement.Value, numVal.Value)
				} else {
					strVal, ok := arrayElement.Value.(*ast.StringLiteral)
					assert.True(ctx.T, ok, "Array element value should be StringLiteral")
					assert.Equal(ctx.T, expectedElement.Value, strVal.Raw)
				}
			} else {
				// 直接元素（非键值对）
				assert.Equal(ctx.T, "", expectedElement.Key, "Expected direct element but got key")
				
				if expectedElement.IsNumeric {
					numVal, ok := element.(*ast.NumberLiteral)
					assert.True(ctx.T, ok, "Array element should be NumberLiteral")
					assert.Equal(ctx.T, expectedElement.Value, numVal.Value)
				} else {
					strVal, ok := element.(*ast.StringLiteral)
					assert.True(ctx.T, ok, "Array element should be StringLiteral")
					assert.Equal(ctx.T, expectedElement.Value, strVal.Raw)
				}
			}
		}
	}
}

// ValidateHeredocAssignment 验证Heredoc赋值
func ValidateHeredocAssignment(varName, expectedValue string) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
		
		// 验证左侧变量
		leftVar, ok := assignment.Left.(*ast.Variable)
		assert.True(ctx.T, ok, "Left side should be Variable")
		assert.Equal(ctx.T, varName, leftVar.Name)
		
		// 验证右侧Heredoc
		stringLit, ok := assignment.Right.(*ast.StringLiteral)
		assert.True(ctx.T, ok, "Right side should be StringLiteral for Heredoc, got %T", assignment.Right)
		assert.Equal(ctx.T, expectedValue, stringLit.Value)
	}
}

// ValidateNowdocAssignment 验证Nowdoc赋值
func ValidateNowdocAssignment(varName, expectedValue string) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
		
		// 验证左侧变量
		leftVar, ok := assignment.Left.(*ast.Variable)
		assert.True(ctx.T, ok, "Left side should be Variable")
		assert.Equal(ctx.T, varName, leftVar.Name)
		
		// 验证右侧Nowdoc
		stringLit, ok := assignment.Right.(*ast.StringLiteral)
		assert.True(ctx.T, ok, "Right side should be StringLiteral for Nowdoc, got %T", assignment.Right)
		assert.Equal(ctx.T, expectedValue, stringLit.Value)
	}
}

// ValidateInterpolatedStringAssignment 验证插值字符串赋值
func ValidateInterpolatedStringAssignment(varName string, interpolation StringInterpolation) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
		
		// 验证左侧变量
		leftVar, ok := assignment.Left.(*ast.Variable)
		assert.True(ctx.T, ok, "Left side should be Variable")
		assert.Equal(ctx.T, varName, leftVar.Name)
		
		// 验证右侧插值字符串
		interpolatedStr, ok := assignment.Right.(*ast.InterpolatedStringExpression)
		assert.True(ctx.T, ok, "Right side should be InterpolatedStringExpression, got %T", assignment.Right)
		
		assert.Len(ctx.T, interpolatedStr.Parts, len(interpolation.Parts), "Interpolation parts count mismatch")
		
		for i, expectedPart := range interpolation.Parts {
			if i >= len(interpolatedStr.Parts) {
				break
			}
			
			part := interpolatedStr.Parts[i]
			
			if expectedPart.Text != "" {
				// 文本部分
				stringLit, ok := part.(*ast.StringLiteral)
				assert.True(ctx.T, ok, "Interpolation part %d should be StringLiteral", i)
				assert.Equal(ctx.T, expectedPart.Text, stringLit.Value)
			} else if expectedPart.Variable != "" {
				// 变量部分
				variable, ok := part.(*ast.Variable)
				assert.True(ctx.T, ok, "Interpolation part %d should be Variable", i)
				assert.Equal(ctx.T, expectedPart.Variable, variable.Name)
			}
		}
	}
}

// ValidateArrayAccess 验证数组访问
func ValidateArrayAccess(varName, arrayVar, index string) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
		
		// 验证左侧变量
		leftVar, ok := assignment.Left.(*ast.Variable)
		assert.True(ctx.T, ok, "Left side should be Variable")
		assert.Equal(ctx.T, varName, leftVar.Name)
		
		// 验证右侧数组访问
		accessExpr, ok := assignment.Right.(*ast.ArrayAccessExpression)
		assert.True(ctx.T, ok, "Right side should be ArrayAccessExpression, got %T", assignment.Right)
		
		// 验证数组变量
		arrVar, ok := accessExpr.Array.(*ast.Variable)
		assert.True(ctx.T, ok, "Array should be Variable")
		assert.Equal(ctx.T, arrayVar, arrVar.Name)
		
		// 验证索引
		if index[0] == '"' || index[0] == '\'' {
			// 字符串索引
			indexStr, ok := (*accessExpr.Index).(*ast.StringLiteral)
			assert.True(ctx.T, ok, "Index should be StringLiteral")
			assert.Equal(ctx.T, index, indexStr.Raw)
		} else {
			// 数值索引
			indexNum, ok := (*accessExpr.Index).(*ast.NumberLiteral)
			assert.True(ctx.T, ok, "Index should be NumberLiteral")
			assert.Equal(ctx.T, index, indexNum.Value)
		}
	}
}

// ValidateChainedArrayAccess 验证链式数组访问
func ValidateChainedArrayAccess(varName, arrayVar string, indices []string) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
		
		// 验证左侧变量
		leftVar, ok := assignment.Left.(*ast.Variable)
		assert.True(ctx.T, ok, "Left side should be Variable")
		assert.Equal(ctx.T, varName, leftVar.Name)
		
		// 验证链式数组访问
		currentExpr := assignment.Right
		
		// 从外层向内层逐层解析
		for i := len(indices) - 1; i >= 0; i-- {
			accessExpr, ok := currentExpr.(*ast.ArrayAccessExpression)
			assert.True(ctx.T, ok, "Should be ArrayAccessExpression at level %d", i)
			
			// 验证索引
			index := indices[i]
			if index[0] == '"' || index[0] == '\'' {
				// 字符串索引
				indexStr, ok := (*accessExpr.Index).(*ast.StringLiteral)
				assert.True(ctx.T, ok, "Index should be StringLiteral at level %d", i)
				assert.Equal(ctx.T, index, indexStr.Raw)
			} else {
				// 数值索引
				indexNum, ok := (*accessExpr.Index).(*ast.NumberLiteral)
				assert.True(ctx.T, ok, "Index should be NumberLiteral at level %d", i)
				assert.Equal(ctx.T, index, indexNum.Value)
			}
			
			if i == 0 {
				// 最内层，应该是原始数组变量
				arrVar, ok := accessExpr.Array.(*ast.Variable)
				assert.True(ctx.T, ok, "Base array should be Variable")
				assert.Equal(ctx.T, arrayVar, arrVar.Name)
			} else {
				// 继续向内层
				currentExpr = accessExpr.Array
			}
		}
	}
}
// 表达式验证函数

// ValidatePrefixExpression 验证前缀表达式
func ValidatePrefixExpression(varName, operandVar, operator string) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
		
		// 验证左侧变量
		leftVar, ok := assignment.Left.(*ast.Variable)
		assert.True(ctx.T, ok, "Left side should be Variable")
		assert.Equal(ctx.T, varName, leftVar.Name)
		
		// 验证右侧前缀表达式
		unaryExpr, ok := assignment.Right.(*ast.UnaryExpression)
		assert.True(ctx.T, ok, "Right side should be UnaryExpression, got %T", assignment.Right)
		
		assert.Equal(ctx.T, operator, unaryExpr.Operator)
		assert.True(ctx.T, unaryExpr.Prefix, "Should be prefix unary expression")
		
		if operandVar != "" {
			operandVariable, ok := unaryExpr.Operand.(*ast.Variable)
			assert.True(ctx.T, ok, "Operand should be Variable")
			assert.Equal(ctx.T, operandVar, operandVariable.Name)
		}
	}
}

// ValidatePostfixAssignment 验证后缀表达式赋值
func ValidatePostfixAssignment(varName, operandVar, operator string) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
		
		// 验证左侧变量
		leftVar, ok := assignment.Left.(*ast.Variable)
		assert.True(ctx.T, ok, "Left side should be Variable")
		assert.Equal(ctx.T, varName, leftVar.Name)
		
		// 验证右侧后缀表达式
		unaryExpr, ok := assignment.Right.(*ast.UnaryExpression)
		assert.True(ctx.T, ok, "Right side should be UnaryExpression, got %T", assignment.Right)
		
		assert.Equal(ctx.T, operator, unaryExpr.Operator)
		assert.False(ctx.T, unaryExpr.Prefix, "Should be postfix unary expression")
		
		if operandVar != "" {
			operandVariable, ok := unaryExpr.Operand.(*ast.Variable)
			assert.True(ctx.T, ok, "Operand should be Variable")
			assert.Equal(ctx.T, operandVar, operandVariable.Name)
		}
	}
}

// ValidateBinaryAssignment 验证二元表达式赋值
func ValidateBinaryAssignment(varName, leftVar, operator, rightVar string) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
		
		// 验证左侧变量
		leftVariable, ok := assignment.Left.(*ast.Variable)
		assert.True(ctx.T, ok, "Left side should be Variable")
		assert.Equal(ctx.T, varName, leftVariable.Name)
		
		// 验证右侧二元表达式
		binExpr := assertions.AssertBinaryExpression(assignment.Right, operator)
		
		// 验证左操作数
		if leftVar != "" {
			leftOperand, ok := binExpr.Left.(*ast.Variable)
			assert.True(ctx.T, ok, "Left operand should be Variable")
			assert.Equal(ctx.T, leftVar, leftOperand.Name)
		}
		
		// 验证右操作数
		if rightVar != "" {
			rightOperand, ok := binExpr.Right.(*ast.Variable)
			assert.True(ctx.T, ok, "Right operand should be Variable")
			assert.Equal(ctx.T, rightVar, rightOperand.Name)
		}
	}
}

// ValidateCoalesceExpression 验证null合并表达式
func ValidateCoalesceExpression(varName, leftVar, rightVar string) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
		
		// 验证左侧变量
		leftVariable, ok := assignment.Left.(*ast.Variable)
		assert.True(ctx.T, ok, "Left side should be Variable")
		assert.Equal(ctx.T, varName, leftVariable.Name)
		
		// 验证右侧coalesce表达式
		coalesceExpr, ok := assignment.Right.(*ast.CoalesceExpression)
		assert.True(ctx.T, ok, "Right side should be CoalesceExpression, got %T", assignment.Right)
		
		// 验证左操作数
		if leftVar != "" {
			leftOperand, ok := coalesceExpr.Left.(*ast.Variable)
			assert.True(ctx.T, ok, "Left operand should be Variable")
			assert.Equal(ctx.T, leftVar, leftOperand.Name)
		}
		
		// 验证右操作数
		if rightVar != "" {
			rightOperand, ok := coalesceExpr.Right.(*ast.Variable)
			assert.True(ctx.T, ok, "Right operand should be Variable")
			assert.Equal(ctx.T, rightVar, rightOperand.Name)
		}
	}
}

// ValidateReturnStatement 验证return语句
func ValidateReturnStatement(expectedValue string) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		returnStmt, ok := body[0].(*ast.ReturnStatement)
		require.True(ctx.T, ok, "Statement should be ReturnStatement, got %T", body[0])
		
		if expectedValue != "" && returnStmt.Argument != nil {
			// 根据期望值的类型进行验证
			if expectedValue[0] == '"' { // 字符串字面量
				stringLit, ok := returnStmt.Argument.(*ast.StringLiteral)
				require.True(ctx.T, ok, "Return argument should be StringLiteral")
				assert.Equal(ctx.T, expectedValue, stringLit.Raw)
			} else if expectedValue[0] >= '0' && expectedValue[0] <= '9' { // 数字字面量
				numLit, ok := returnStmt.Argument.(*ast.NumberLiteral)
				require.True(ctx.T, ok, "Return argument should be NumberLiteral")
				assert.Equal(ctx.T, expectedValue, numLit.Value)
			} else if expectedValue[0] == '$' { // 变量
				variable, ok := returnStmt.Argument.(*ast.Variable)
				require.True(ctx.T, ok, "Return argument should be Variable")
				assert.Equal(ctx.T, expectedValue, variable.Name)
			}
		}
	}
}

// ValidateReturnVariable 验证return变量
func ValidateReturnVariable(varName string) ValidationFunc {
	return ValidateReturnStatement(varName)
}

// ValidateReturnBinaryExpression 验证return二元表达式
func ValidateReturnBinaryExpression(leftVar, operator, rightVar string) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		returnStmt, ok := body[0].(*ast.ReturnStatement)
		require.True(ctx.T, ok, "Statement should be ReturnStatement")
		
		binExpr := assertions.AssertBinaryExpression(returnStmt.Argument, operator)
		
		// 验证左操作数
		if leftVar != "" {
			if leftVar[0] == '$' {
				leftOperand, ok := binExpr.Left.(*ast.Variable)
				require.True(ctx.T, ok, "Left operand should be Variable")
				assert.Equal(ctx.T, leftVar, leftOperand.Name)
			}
		}
		
		// 验证右操作数
		if rightVar != "" {
			if rightVar[0] == '$' {
				rightOperand, ok := binExpr.Right.(*ast.Variable)
				require.True(ctx.T, ok, "Right operand should be Variable")
				assert.Equal(ctx.T, rightVar, rightOperand.Name)
			} else if rightVar[0] >= '0' && rightVar[0] <= '9' {
				rightOperand, ok := binExpr.Right.(*ast.NumberLiteral)
				require.True(ctx.T, ok, "Right operand should be NumberLiteral")
				assert.Equal(ctx.T, rightVar, rightOperand.Value)
			}
		}
	}
}

// ValidateReturnNull 验证return null
func ValidateReturnNull() ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		returnStmt, ok := body[0].(*ast.ReturnStatement)
		require.True(ctx.T, ok, "Statement should be ReturnStatement")
		
		// 检查是否是null
		if nullLit, ok := returnStmt.Argument.(*ast.NullLiteral); ok {
			assert.NotNil(ctx.T, nullLit, "Should be null literal")
		} else if ident, ok := returnStmt.Argument.(*ast.IdentifierNode); ok {
			assert.Equal(ctx.T, "null", ident.Name)
		}
	}
}

// ValidatePropertyDeclaration 验证属性声明
func ValidatePropertyDeclaration(visibility, varName, typeName, defaultValue string) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		propDecl, ok := body[0].(*ast.PropertyDeclaration)
		require.True(ctx.T, ok, "Statement should be PropertyDeclaration, got %T", body[0])
		
		// 验证可见性
		if visibility != "" {
			assert.Equal(ctx.T, visibility, propDecl.Visibility)
		}
		
		// 验证属性名
		if varName != "" {
			expectedName := varName
			if expectedName[0] == '$' {
				expectedName = expectedName[1:] // 去掉$前缀，PropertyDeclaration.Name不包含$
			}
			assert.Equal(ctx.T, expectedName, propDecl.Name)
		}
	}
}

// ValidateInstanceofExpression 验证instanceof表达式
func ValidateInstanceofExpression(varName, objectVar, className string) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
		
		// 验证左侧变量
		leftVar, ok := assignment.Left.(*ast.Variable)
		assert.True(ctx.T, ok, "Left side should be Variable")
		assert.Equal(ctx.T, varName, leftVar.Name)
		
		// 验证右侧instanceof表达式
		instanceofExpr, ok := assignment.Right.(*ast.InstanceofExpression)
		assert.True(ctx.T, ok, "Right side should be InstanceofExpression, got %T", assignment.Right)
		
		// 验证对象变量
		if objectVar != "" {
			objectVariable, ok := instanceofExpr.Left.(*ast.Variable)
			assert.True(ctx.T, ok, "Left operand should be Variable")
			assert.Equal(ctx.T, objectVar, objectVariable.Name)
		}
		
		// 验证类名
		if className != "" {
			classIdent, ok := instanceofExpr.Right.(*ast.IdentifierNode)
			assert.True(ctx.T, ok, "Right operand should be IdentifierNode")
			assert.Equal(ctx.T, className, classIdent.Name)
		}
	}
}

// ValidateTernaryExpression 验证三元表达式
func ValidateTernaryExpression(varName, conditionVar, trueVar, falseVar string) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
		
		// 验证左侧变量
		leftVar, ok := assignment.Left.(*ast.Variable)
		assert.True(ctx.T, ok, "Left side should be Variable")
		assert.Equal(ctx.T, varName, leftVar.Name)
		
		// 验证右侧三元表达式
		ternaryExpr, ok := assignment.Right.(*ast.TernaryExpression)
		assert.True(ctx.T, ok, "Right side should be TernaryExpression, got %T", assignment.Right)
		
		// 验证条件
		if conditionVar != "" {
			conditionVariable, ok := ternaryExpr.Test.(*ast.Variable)
			assert.True(ctx.T, ok, "Condition should be Variable")
			assert.Equal(ctx.T, conditionVar, conditionVariable.Name)
		}
		
		// 验证真值
		if trueVar != "" {
			trueVariable, ok := ternaryExpr.Consequent.(*ast.Variable)
			assert.True(ctx.T, ok, "True value should be Variable")
			assert.Equal(ctx.T, trueVar, trueVariable.Name)
		}
		
		// 验证假值
		if falseVar != "" {
			falseVariable, ok := ternaryExpr.Alternate.(*ast.Variable)
			assert.True(ctx.T, ok, "False value should be Variable")
			assert.Equal(ctx.T, falseVar, falseVariable.Name)
		}
	}
}

// ValidateAssignmentOperation 验证赋值操作
func ValidateAssignmentOperation(varName, operator, valueVar string) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		assignment := assertions.AssertAssignment(exprStmt.Expression, operator)
		
		// 验证左侧变量
		leftVar, ok := assignment.Left.(*ast.Variable)
		assert.True(ctx.T, ok, "Left side should be Variable")
		assert.Equal(ctx.T, varName, leftVar.Name)
		
		// 验证右侧值
		if valueVar != "" {
			rightVariable, ok := assignment.Right.(*ast.Variable)
			assert.True(ctx.T, ok, "Right side should be Variable")
			assert.Equal(ctx.T, valueVar, rightVariable.Name)
		}
	}
}
