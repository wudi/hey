package testutils

// ValidateAssignment 验证赋值表达式，支持指定操作符
func ValidateAssignment(varName, operator string) ValidationFunc {
	return func(ctx *TestContext) {
		assertions := NewASTAssertions(ctx.T)
		body := assertions.AssertProgramBody(ctx.Program, 1)
		
		exprStmt := assertions.AssertExpressionStatement(body[0])
		assignment := assertions.AssertAssignment(exprStmt.Expression, operator)
		assertions.AssertVariable(assignment.Left, varName)
	}
}

// ValidateBasicAssignment 验证基础赋值 (=)
func ValidateBasicAssignment(varName string) ValidationFunc {
	return ValidateAssignment(varName, "=")
}

// ValidateAdditionAssignment 验证加法赋值 (+=)
func ValidateAdditionAssignment(varName string) ValidationFunc {
	return ValidateAssignment(varName, "+=")
}

// ValidateSubtractionAssignment 验证减法赋值 (-=)
func ValidateSubtractionAssignment(varName string) ValidationFunc {
	return ValidateAssignment(varName, "-=")
}

// ValidateMultiplicationAssignment 验证乘法赋值 (*=)
func ValidateMultiplicationAssignment(varName string) ValidationFunc {
	return ValidateAssignment(varName, "*=")
}

// ValidateDivisionAssignment 验证除法赋值 (/=)
func ValidateDivisionAssignment(varName string) ValidationFunc {
	return ValidateAssignment(varName, "/=")
}

// ValidateModuloAssignment 验证模运算赋值 (%=)
func ValidateModuloAssignment(varName string) ValidationFunc {
	return ValidateAssignment(varName, "%=")
}

// ValidatePowerAssignment 验证幂运算赋值 (**=)
func ValidatePowerAssignment(varName string) ValidationFunc {
	return ValidateAssignment(varName, "**=")
}

// ValidateConcatenationAssignment 验证字符串连接赋值 (.=)
func ValidateConcatenationAssignment(varName string) ValidationFunc {
	return ValidateAssignment(varName, ".=")
}

// ValidateCoalesceAssignment 验证null合并赋值 (??=)
func ValidateCoalesceAssignment(varName string) ValidationFunc {
	return ValidateAssignment(varName, "??=")
}

// ValidateBitwiseAndAssignment 验证按位与赋值 (&=)
func ValidateBitwiseAndAssignment(varName string) ValidationFunc {
	return ValidateAssignment(varName, "&=")
}

// ValidateBitwiseOrAssignment 验证按位或赋值 (|=)
func ValidateBitwiseOrAssignment(varName string) ValidationFunc {
	return ValidateAssignment(varName, "|=")
}

// ValidateBitwiseXorAssignment 验证按位异或赋值 (^=)
func ValidateBitwiseXorAssignment(varName string) ValidationFunc {
	return ValidateAssignment(varName, "^=")
}

// ValidateLeftShiftAssignment 验证左移赋值 (<<=)
func ValidateLeftShiftAssignment(varName string) ValidationFunc {
	return ValidateAssignment(varName, "<<=")
}

// ValidateRightShiftAssignment 验证右移赋值 (>>=)
func ValidateRightShiftAssignment(varName string) ValidationFunc {
	return ValidateAssignment(varName, ">>=")
}

// ValidateVariableAssignment 验证变量赋值的别名，避免与已有函数冲突  
func ValidateVariableAssignment(varName string) ValidationFunc {
	return ValidateBasicAssignment(varName)
}