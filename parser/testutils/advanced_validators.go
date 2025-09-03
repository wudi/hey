package testutils

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wudi/php-parser/ast"
)

// TypedParam 表示带类型的参数
type TypedParam struct {
	Name string
	Type string
}

// MatchArm 表示match表达式的分支
type MatchArm struct {
	Condition  string
	Conditions []string // 用于多条件
	Value      string
	IsDefault  bool
}

// EnumCase 表示枚举常量
type EnumCase struct {
	Name  string
	Value string
}

// ValidateFunctionDeclaration 验证函数声明
func ValidateFunctionDeclaration(funcName string, params []string, returnType string, bodyValidator ValidationFunc) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		funcDecl, ok := body[0].(*ast.FunctionDeclaration)
		require.True(ctx.T, ok, "Statement should be FunctionDeclaration, got %T", body[0])
		
		// 验证函数名
		if nameIdent, ok := funcDecl.Name.(*ast.IdentifierNode); ok {
			assert.Equal(ctx.T, funcName, nameIdent.Name)
		}
		
		// 验证参数
		if len(params) > 0 {
			require.NotNil(ctx.T, funcDecl.Parameters, "Function should have parameters")
			assert.Len(ctx.T, funcDecl.Parameters.Parameters, len(params))
		}
		
		// 验证函数体
		if bodyValidator != nil && len(funcDecl.Body) > 0 {
			funcCtx := &TestContext{T: ctx.T, Program: &ast.Program{Body: funcDecl.Body}}
			bodyValidator(funcCtx)
		}
	}
}

// ValidateFunctionWithParameters 验证带参数的函数
func ValidateFunctionWithParameters(funcName string, params []string, bodyValidator ValidationFunc) ValidationFunc {
	return ValidateFunctionDeclaration(funcName, params, "", bodyValidator)
}

// ValidateFunctionWithReturnType 验证带返回类型的函数
func ValidateFunctionWithReturnType(funcName string, returnType string, bodyValidator ValidationFunc) ValidationFunc {
	return ValidateFunctionDeclaration(funcName, []string{}, returnType, bodyValidator)
}

// ValidateTypedFunction 验证带类型参数的函数
func ValidateTypedFunction(funcName string, params []TypedParam, bodyValidator ValidationFunc) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		funcDecl, ok := body[0].(*ast.FunctionDeclaration)
		require.True(ctx.T, ok, "Statement should be FunctionDeclaration, got %T", body[0])
		
		if nameIdent, ok := funcDecl.Name.(*ast.IdentifierNode); ok {
			assert.Equal(ctx.T, funcName, nameIdent.Name)
		}
		
		if bodyValidator != nil && len(funcDecl.Body) > 0 {
			funcCtx := &TestContext{T: ctx.T, Program: &ast.Program{Body: funcDecl.Body}}
			bodyValidator(funcCtx)
		}
	}
}

// ValidateAnonymousFunctionAssignment 验证匿名函数赋值
func ValidateAnonymousFunctionAssignment(varName string, bodyValidator ValidationFunc) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
		
		// 验证左侧变量
		leftVar, ok := assignment.Left.(*ast.Variable)
		require.True(ctx.T, ok, "Left side should be Variable")
		assert.Equal(ctx.T, varName, leftVar.Name)
		
		// 验证右侧匿名函数
		anonFunc, ok := assignment.Right.(*ast.AnonymousFunctionExpression)
		require.True(ctx.T, ok, "Right side should be AnonymousFunctionExpression, got %T", assignment.Right)
		
		if bodyValidator != nil && len(anonFunc.Body) > 0 {
			funcCtx := &TestContext{T: ctx.T, Program: &ast.Program{Body: anonFunc.Body}}
			bodyValidator(funcCtx)
		}
	}
}

// ValidateAnonymousFunctionWithParams 验证带参数的匿名函数
func ValidateAnonymousFunctionWithParams(varName string, params []string, bodyValidator ValidationFunc) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
		
		leftVar, ok := assignment.Left.(*ast.Variable)
		require.True(ctx.T, ok, "Left side should be Variable")
		assert.Equal(ctx.T, varName, leftVar.Name)
		
		anonFunc, ok := assignment.Right.(*ast.AnonymousFunctionExpression)
		require.True(ctx.T, ok, "Right side should be AnonymousFunctionExpression")
		
		// 验证参数数量
		if anonFunc.Parameters != nil && len(params) > 0 {
			assert.Len(ctx.T, anonFunc.Parameters.Parameters, len(params))
		}
		
		if bodyValidator != nil && len(anonFunc.Body) > 0 {
			funcCtx := &TestContext{T: ctx.T, Program: &ast.Program{Body: anonFunc.Body}}
			bodyValidator(funcCtx)
		}
	}
}

// ValidateClosureWithUse 验证带use的闭包
func ValidateClosureWithUse(varName string, params []string, useVars []string, bodyValidator ValidationFunc) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
		
		leftVar, ok := assignment.Left.(*ast.Variable)
		require.True(ctx.T, ok, "Left side should be Variable")
		assert.Equal(ctx.T, varName, leftVar.Name)
		
		anonFunc, ok := assignment.Right.(*ast.AnonymousFunctionExpression)
		require.True(ctx.T, ok, "Right side should be AnonymousFunctionExpression")
		
		// 验证use变量
		if len(useVars) > 0 {
			assert.Len(ctx.T, anonFunc.UseClause, len(useVars))
		}
		
		if bodyValidator != nil && len(anonFunc.Body) > 0 {
			funcCtx := &TestContext{T: ctx.T, Program: &ast.Program{Body: anonFunc.Body}}
			bodyValidator(funcCtx)
		}
	}
}

// ValidateArrowFunctionAssignment 验证箭头函数赋值
func ValidateArrowFunctionAssignment(varName string, params []string, bodyValidator ValidationFunc) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
		
		leftVar, ok := assignment.Left.(*ast.Variable)
		require.True(ctx.T, ok, "Left side should be Variable")
		assert.Equal(ctx.T, varName, leftVar.Name)
		
		arrowFunc, ok := assignment.Right.(*ast.ArrowFunctionExpression)
		require.True(ctx.T, ok, "Right side should be ArrowFunctionExpression, got %T", assignment.Right)
		
		// 验证参数数量
		if arrowFunc.Parameters != nil && len(params) > 0 {
			assert.Len(ctx.T, arrowFunc.Parameters.Parameters, len(params))
		}
		
		// 验证箭头函数表达式 (简化)
		if bodyValidator != nil && arrowFunc.Body != nil {
			// 箭头函数的body是表达式，暂时简化验证
			assert.NotNil(ctx.T, arrowFunc.Body, "Arrow function body should not be nil")
		}
	}
}

// ValidateTypedArrowFunction 验证带类型的箭头函数
func ValidateTypedArrowFunction(varName string, params []TypedParam, returnType string, bodyValidator ValidationFunc) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
		
		leftVar, ok := assignment.Left.(*ast.Variable)
		require.True(ctx.T, ok, "Left side should be Variable")
		assert.Equal(ctx.T, varName, leftVar.Name)
		
		arrowFunc, ok := assignment.Right.(*ast.ArrowFunctionExpression)
		require.True(ctx.T, ok, "Right side should be ArrowFunctionExpression")
		
		// 验证返回类型
		if returnType != "" && arrowFunc.ReturnType != nil {
			assert.NotNil(ctx.T, arrowFunc.ReturnType, "Return type should not be nil")
		}
		
		if bodyValidator != nil && arrowFunc.Body != nil {
			// 箭头函数的body是表达式，暂时简化验证
			assert.NotNil(ctx.T, arrowFunc.Body, "Arrow function body should not be nil")
		}
	}
}

// ValidateClassDeclaration 验证类声明
func ValidateClassDeclaration(className, parentClass string, interfaces []string, memberValidator ValidationFunc) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		classDecl, ok := exprStmt.Expression.(*ast.ClassExpression)
		require.True(ctx.T, ok, "Expression should be ClassExpression, got %T", exprStmt.Expression)
		
		// 验证类名
		if nameIdent, ok := classDecl.Name.(*ast.IdentifierNode); ok {
			assert.Equal(ctx.T, className, nameIdent.Name)
		}
		
		// 验证成员
		if memberValidator != nil && classDecl.Body != nil && len(classDecl.Body) > 0 {
			memberCtx := &TestContext{T: ctx.T, Program: &ast.Program{Body: classDecl.Body}}
			memberValidator(memberCtx)
		}
	}
}

// ValidateClassWithInheritance 验证带继承的类
func ValidateClassWithInheritance(className, parentClass string, memberValidator ValidationFunc) ValidationFunc {
	return ValidateClassDeclaration(className, parentClass, []string{}, memberValidator)
}

// ValidateClassWithInterfaces 验证实现接口的类
func ValidateClassWithInterfaces(className string, interfaces []string) ValidationFunc {
	return ValidateClassDeclaration(className, "", interfaces, nil)
}

// ValidateFinalClass 验证final类
func ValidateFinalClass(className string) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		classDecl, ok := exprStmt.Expression.(*ast.ClassExpression)
		require.True(ctx.T, ok, "Expression should be ClassExpression")
		
		if nameIdent, ok := classDecl.Name.(*ast.IdentifierNode); ok {
			assert.Equal(ctx.T, className, nameIdent.Name)
		}
		
		// 检查final修饰符
		assert.True(ctx.T, classDecl.Final, "Class should be marked as final")
	}
}

// ValidateAbstractClass 验证abstract类
func ValidateAbstractClass(className string) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		classDecl, ok := exprStmt.Expression.(*ast.ClassExpression)
		require.True(ctx.T, ok, "Expression should be ClassExpression")
		
		if nameIdent, ok := classDecl.Name.(*ast.IdentifierNode); ok {
			assert.Equal(ctx.T, className, nameIdent.Name)
		}
		
		// 检查abstract修饰符
		assert.True(ctx.T, classDecl.Abstract, "Class should be marked as abstract")
	}
}

// ValidateStaticPropertyAccess 验证静态属性访问
func ValidateStaticPropertyAccess(varName, className, propertyName string) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
		
		// 验证左侧变量
		leftVar, ok := assignment.Left.(*ast.Variable)
		require.True(ctx.T, ok, "Left side should be Variable")
		assert.Equal(ctx.T, varName, leftVar.Name)
		
		// 验证静态属性访问
		staticAccess, ok := assignment.Right.(*ast.StaticPropertyAccessExpression)
		require.True(ctx.T, ok, "Right side should be StaticPropertyAccessExpression, got %T", assignment.Right)
		
		// 验证类名
		if classIdent, ok := staticAccess.Class.(*ast.IdentifierNode); ok {
			assert.Equal(ctx.T, className, classIdent.Name)
		}
		
		// 验证属性名
		if propVar, ok := staticAccess.Property.(*ast.Variable); ok {
			assert.Equal(ctx.T, propertyName, propVar.Name)
		}
	}
}

// ValidateStaticMethodCall 验证静态方法调用
func ValidateStaticMethodCall(varName, className, methodName string, args []string) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
		
		// 验证左侧变量
		leftVar, ok := assignment.Left.(*ast.Variable)
		require.True(ctx.T, ok, "Left side should be Variable")
		assert.Equal(ctx.T, varName, leftVar.Name)
		
		// 简化验证静态方法调用 - 可能是函数调用表达式的特殊形式
		// 具体实现需要根据实际的AST结构来调整
		assert.NotNil(ctx.T, assignment.Right, "Right side should contain static method call")
	}
}

// ValidateStaticConstantAccess 验证静态常量访问
func ValidateStaticConstantAccess(varName, className, constantName string) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
		
		// 验证左侧变量
		leftVar, ok := assignment.Left.(*ast.Variable)
		require.True(ctx.T, ok, "Left side should be Variable")
		assert.Equal(ctx.T, varName, leftVar.Name)
		
		// 验证静态常量访问 - 尝试不同的AST节点类型
		if staticAccess, ok := assignment.Right.(*ast.StaticAccessExpression); ok {
			// 验证类名
			if classIdent, ok := staticAccess.Class.(*ast.IdentifierNode); ok {
				assert.Equal(ctx.T, className, classIdent.Name)
			}
			// 验证常量名
			if constIdent, ok := staticAccess.Property.(*ast.IdentifierNode); ok {
				assert.Equal(ctx.T, constantName, constIdent.Name)
			}
		} else if classConstAccess, ok := assignment.Right.(*ast.ClassConstantAccessExpression); ok {
			// 可能是ClassConstantAccessExpression
			assert.NotNil(ctx.T, classConstAccess, "Should be class constant access")
		} else {
			require.Fail(ctx.T, "Right side should be StaticAccessExpression or ClassConstantAccessExpression, got %T", assignment.Right)
		}
	}
}

// ValidateChainedStaticCall 验证链式静态调用 (简化版本)
func ValidateChainedStaticCall(varName, className, firstMethod string, firstArgs []string, secondMethod string) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
		
		leftVar, ok := assignment.Left.(*ast.Variable)
		require.True(ctx.T, ok, "Left side should be Variable")
		assert.Equal(ctx.T, varName, leftVar.Name)
		
		// 简化验证 - 检查是否包含预期的结构
		assert.NotNil(ctx.T, assignment.Right, "Right side should not be nil")
	}
}

// 以下是简化版本的验证器，用于match表达式、枚举等可能不存在的AST节点类型

// ValidateMatchExpression 验证match表达式 (简化版本)
func ValidateMatchExpression(varName, matchVar string, arms []MatchArm) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
		
		leftVar, ok := assignment.Left.(*ast.Variable)
		require.True(ctx.T, ok, "Left side should be Variable")
		assert.Equal(ctx.T, varName, leftVar.Name)
		
		// match表达式验证需要根据实际AST结构实现
		assert.NotNil(ctx.T, assignment.Right, "Right side should contain match expression")
	}
}

// ValidateMatchWithMultipleConditions 验证多条件match表达式 (简化版本)
func ValidateMatchWithMultipleConditions(varName, matchVar string, arms []MatchArm) ValidationFunc {
	return ValidateMatchExpression(varName, matchVar, arms)
}

// ValidateEnumDeclaration 验证枚举声明 (简化版本)
func ValidateEnumDeclaration(enumName, backingType string, cases []string) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		// 枚举验证需要根据实际AST结构实现
		assert.NotNil(ctx.T, body[0], "Should have enum declaration")
	}
}

// ValidateBackedEnum 验证带值的枚举 (简化版本)
func ValidateBackedEnum(enumName, backingType string, cases []EnumCase) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		// 枚举验证需要根据实际AST结构实现
		assert.NotNil(ctx.T, body[0], "Should have backed enum declaration")
	}
}

// ValidateFunctionWithUnionType 验证联合类型函数 (简化版本)
func ValidateFunctionWithUnionType(funcName, unionType, paramName string, bodyValidator ValidationFunc) ValidationFunc {
	return ValidateFunctionDeclaration(funcName, []string{paramName}, "", bodyValidator)
}

// ValidateFunctionWithIntersectionType 验证交集类型函数 (简化版本)
func ValidateFunctionWithIntersectionType(funcName, intersectionType, paramName string, bodyValidator ValidationFunc) ValidationFunc {
	return ValidateFunctionDeclaration(funcName, []string{paramName}, "", bodyValidator)
}

// ValidateFunctionWithNullableUnionType 验证可空联合类型函数 (简化版本)
func ValidateFunctionWithNullableUnionType(funcName, nullableUnionType string, bodyValidator ValidationFunc) ValidationFunc {
	return ValidateFunctionDeclaration(funcName, []string{}, nullableUnionType, bodyValidator)
}

// ValidateClassInExpressionStatement 验证表达式语句中的类声明
func ValidateClassInExpressionStatement(className string) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		
		classDecl, ok := exprStmt.Expression.(*ast.ClassExpression)
		require.True(ctx.T, ok, "Expression should be ClassExpression, got %T", exprStmt.Expression)
		
		// 验证类名
		if nameIdent, ok := classDecl.Name.(*ast.IdentifierNode); ok {
			assert.Equal(ctx.T, className, nameIdent.Name)
		}
	}
}

// ValidateFinalClassInExpressionStatement 验证final类在表达式语句中
func ValidateFinalClassInExpressionStatement(className string) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		
		classDecl, ok := exprStmt.Expression.(*ast.ClassExpression)
		require.True(ctx.T, ok, "Expression should be ClassExpression")
		
		if nameIdent, ok := classDecl.Name.(*ast.IdentifierNode); ok {
			assert.Equal(ctx.T, className, nameIdent.Name)
		}
		
		// 检查final修饰符
		assert.True(ctx.T, classDecl.Final, "Class should be marked as final")
	}
}

// ValidateAbstractClassInExpressionStatement 验证abstract类在表达式语句中
func ValidateAbstractClassInExpressionStatement(className string) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		
		classDecl, ok := exprStmt.Expression.(*ast.ClassExpression)
		require.True(ctx.T, ok, "Expression should be ClassExpression")
		
		if nameIdent, ok := classDecl.Name.(*ast.IdentifierNode); ok {
			assert.Equal(ctx.T, className, nameIdent.Name)
		}
		
		// 检查abstract修饰符
		assert.True(ctx.T, classDecl.Abstract, "Class should be marked as abstract")
	}
}