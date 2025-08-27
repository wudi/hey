package ast

// ASTKind 定义AST节点类型，与PHP官方zend_ast.h保持一致
type ASTKind uint16

const (
	// 特殊节点 - special nodes (bit 6 set)
	ASTZval         ASTKind = 64  // 1 << 6
	ASTConstant     ASTKind = 65
	ASTOpArray      ASTKind = 66
	ASTZNode        ASTKind = 67

	// 声明节点 - declaration nodes
	ASTFuncDecl     ASTKind = 68
	ASTClosure      ASTKind = 69
	ASTMethod       ASTKind = 70
	ASTClass        ASTKind = 71
	ASTArrowFunc    ASTKind = 72
	ASTPropertyHook ASTKind = 73
	ASTPropertyDecl ASTKind = 74

	// 列表节点 - list nodes (bit 7 set)
	ASTArgList          ASTKind = 128 // 1 << 7
	ASTArray            ASTKind = 129
	ASTEncapsList       ASTKind = 130
	ASTExprList         ASTKind = 131
	ASTStmtList         ASTKind = 132
	ASTIf               ASTKind = 133
	ASTSwitchList       ASTKind = 134
	ASTCatchList        ASTKind = 135
	ASTParamList        ASTKind = 136
	ASTClosureUses      ASTKind = 137
	ASTPropDecl         ASTKind = 138
	ASTConstDecl        ASTKind = 139
	ASTClassConstDecl   ASTKind = 140
	ASTNameList         ASTKind = 141
	ASTTraitAdaptations ASTKind = 142
	ASTUse              ASTKind = 143
	ASTTypeUnion        ASTKind = 144
	ASTTypeIntersection ASTKind = 145
	ASTAttributeList    ASTKind = 146
	ASTAttributeGroup   ASTKind = 147
	ASTMatchArmList     ASTKind = 148
	ASTModifierList     ASTKind = 149

	// 0子节点 - 0 child nodes (bits 8-15 = 0)
	ASTMagicConst      ASTKind = 0
	ASTType            ASTKind = 1
	ASTConstantClass   ASTKind = 2
	ASTCallableConvert ASTKind = 3

	// 1子节点 - 1 child node (bits 8-15 = 1)
	ASTVar                   ASTKind = 256 // 1 << 8
	ASTConst                 ASTKind = 257
	ASTUnpack                ASTKind = 258
	ASTUnaryPlus             ASTKind = 259
	ASTUnaryMinus            ASTKind = 260
	ASTCast                  ASTKind = 261
	ASTCastVoid              ASTKind = 262
	ASTEmpty                 ASTKind = 263
	ASTIsset                 ASTKind = 264
	ASTSilence               ASTKind = 265
	ASTShellExec             ASTKind = 266
	ASTPrint                 ASTKind = 267
	ASTIncludeOrEval         ASTKind = 268
	ASTUnaryOp               ASTKind = 269
	ASTPreInc                ASTKind = 270
	ASTPreDec                ASTKind = 271
	ASTPostInc               ASTKind = 272
	ASTPostDec               ASTKind = 273
	ASTYieldFrom             ASTKind = 274
	ASTClassName             ASTKind = 275
	ASTGlobal                ASTKind = 276
	ASTUnset                 ASTKind = 277
	ASTReturn                ASTKind = 278
	ASTLabel                 ASTKind = 279
	ASTRef                   ASTKind = 280
	ASTHaltCompiler          ASTKind = 281
	ASTEcho                  ASTKind = 282
	ASTThrow                 ASTKind = 283
	ASTGoto                  ASTKind = 284
	ASTBreak                 ASTKind = 285
	ASTContinue              ASTKind = 286
	ASTPropertyHookShortBody ASTKind = 287
	ASTClone                 ASTKind = 288
	ASTExit                  ASTKind = 289
	ASTList                  ASTKind = 290
	ASTAnonymousFunction     ASTKind = 291

	// 2子节点 - 2 child nodes (bits 8-15 = 2)
	ASTDim                       ASTKind = 512 // 2 << 8
	ASTProp                      ASTKind = 513
	ASTNullsafeProp              ASTKind = 514
	ASTStaticProp                ASTKind = 515
	ASTCall                      ASTKind = 516
	ASTClassConst                ASTKind = 517
	ASTAssign                    ASTKind = 518
	ASTAssignRef                 ASTKind = 519
	ASTAssignOp                  ASTKind = 520
	ASTBinaryOp                  ASTKind = 521
	ASTGreater                   ASTKind = 522
	ASTGreaterEqual              ASTKind = 523
	ASTAnd                       ASTKind = 524
	ASTOr                        ASTKind = 525
	ASTArrayElem                 ASTKind = 526
	ASTNew                       ASTKind = 527
	ASTInstanceof                ASTKind = 528
	ASTYield                     ASTKind = 529
	ASTCoalesce                  ASTKind = 530
	ASTAssignCoalesce            ASTKind = 531
	ASTStatic                    ASTKind = 532
	ASTWhile                     ASTKind = 533
	ASTDoWhile                   ASTKind = 534
	ASTIfElem                    ASTKind = 535
	ASTSwitch                    ASTKind = 536
	ASTSwitchCase                ASTKind = 537
	ASTDeclare                   ASTKind = 538
	ASTAltIf                     ASTKind = 539
	ASTAltWhile                  ASTKind = 540  
	ASTAltFor                    ASTKind = 541
	ASTAltForeach                ASTKind = 542
	ASTElseIf                    ASTKind = 543
	ASTUseTrait                  ASTKind = 544
	ASTTraitPrecedence           ASTKind = 545
	ASTMethodReference           ASTKind = 546
	ASTNamespace                 ASTKind = 547
	ASTNamespaceName             ASTKind = 548
	ASTUseElem                   ASTKind = 549
	ASTTraitAlias                ASTKind = 550
	ASTGroupUse                  ASTKind = 551
	ASTAttribute                 ASTKind = 552
	ASTMatch                     ASTKind = 553
	ASTMatchArm                  ASTKind = 554
	ASTNamedArg                  ASTKind = 555
	ASTParentPropertyHookCall    ASTKind = 556
	ASTPipe                      ASTKind = 557
	ASTInterface                 ASTKind = 558
	ASTTrait                     ASTKind = 559
	ASTEnum                      ASTKind = 560

	// 3子节点 - 3 child nodes (bits 8-15 = 3)
	ASTMethodCall         ASTKind = 768 // 3 << 8
	ASTNullsafeMethodCall ASTKind = 769
	ASTStaticCall         ASTKind = 770
	ASTConditional        ASTKind = 771
	ASTTry                ASTKind = 772
	ASTCatch              ASTKind = 773
	ASTPropGroup          ASTKind = 774
	ASTConstElem          ASTKind = 775
	ASTClassConstGroup    ASTKind = 776
	ASTConstEnumInit      ASTKind = 777
	ASTCase               ASTKind = 778
	ASTEvalExpression     ASTKind = 779 
	ASTVisibilityModifier ASTKind = 780

	// 4子节点 - 4 child nodes (bits 8-15 = 4)
	ASTFor      ASTKind = 1024 // 4 << 8
	ASTForeach  ASTKind = 1025
	ASTEnumCase ASTKind = 1026
	ASTPropElem ASTKind = 1027

	// 6子节点 - 6 child nodes (bits 8-15 = 6)
	ASTParam ASTKind = 1536 // 6 << 8
)

// String 返回AST节点类型的字符串表示
func (k ASTKind) String() string {
	switch k {
	// 特殊节点
	case ASTZval:
		return "ZVAL"
	case ASTConstant:
		return "CONSTANT"
	case ASTOpArray:
		return "OP_ARRAY"
	case ASTZNode:
		return "ZNODE"

	// 声明节点
	case ASTFuncDecl:
		return "FUNC_DECL"
	case ASTClosure:
		return "CLOSURE"
	case ASTMethod:
		return "METHOD"
	case ASTClass:
		return "CLASS"
	case ASTArrowFunc:
		return "ARROW_FUNC"
	case ASTPropertyHook:
		return "PROPERTY_HOOK"
	case ASTPropertyDecl:
		return "PROPERTY_DECL"

	// 列表节点
	case ASTArgList:
		return "ARG_LIST"
	case ASTArray:
		return "ARRAY"
	case ASTEncapsList:
		return "ENCAPS_LIST"
	case ASTExprList:
		return "EXPR_LIST"
	case ASTStmtList:
		return "STMT_LIST"
	case ASTIf:
		return "IF"
	case ASTSwitchList:
		return "SWITCH_LIST"
	case ASTCatchList:
		return "CATCH_LIST"
	case ASTParamList:
		return "PARAM_LIST"
	case ASTClosureUses:
		return "CLOSURE_USES"
	case ASTPropDecl:
		return "PROP_DECL"
	case ASTConstDecl:
		return "CONST_DECL"
	case ASTClassConstDecl:
		return "CLASS_CONST_DECL"
	case ASTNameList:
		return "NAME_LIST"
	case ASTTraitAdaptations:
		return "TRAIT_ADAPTATIONS"
	case ASTUse:
		return "USE"
	case ASTTypeUnion:
		return "TYPE_UNION"
	case ASTTypeIntersection:
		return "TYPE_INTERSECTION"
	case ASTAttributeList:
		return "ATTRIBUTE_LIST"
	case ASTAttributeGroup:
		return "ATTRIBUTE_GROUP"
	case ASTMatchArmList:
		return "MATCH_ARM_LIST"
	case ASTModifierList:
		return "MODIFIER_LIST"

	// 0子节点
	case ASTMagicConst:
		return "MAGIC_CONST"
	case ASTType:
		return "TYPE"
	case ASTConstantClass:
		return "CONSTANT_CLASS"
	case ASTCallableConvert:
		return "CALLABLE_CONVERT"

	// 1子节点
	case ASTVar:
		return "VAR"
	case ASTConst:
		return "CONST"
	case ASTUnpack:
		return "UNPACK"
	case ASTUnaryPlus:
		return "UNARY_PLUS"
	case ASTUnaryMinus:
		return "UNARY_MINUS"
	case ASTCast:
		return "CAST"
	case ASTCastVoid:
		return "CAST_VOID"
	case ASTEmpty:
		return "EMPTY"
	case ASTIsset:
		return "ISSET"
	case ASTSilence:
		return "SILENCE"
	case ASTShellExec:
		return "SHELL_EXEC"
	case ASTPrint:
		return "PRINT"
	case ASTIncludeOrEval:
		return "INCLUDE_OR_EVAL"
	case ASTUnaryOp:
		return "UNARY_OP"
	case ASTPreInc:
		return "PRE_INC"
	case ASTPreDec:
		return "PRE_DEC"
	case ASTPostInc:
		return "POST_INC"
	case ASTPostDec:
		return "POST_DEC"
	case ASTYieldFrom:
		return "YIELD_FROM"
	case ASTClassName:
		return "CLASS_NAME"
	case ASTGlobal:
		return "GLOBAL"
	case ASTUnset:
		return "UNSET"
	case ASTReturn:
		return "RETURN"
	case ASTLabel:
		return "LABEL"
	case ASTRef:
		return "REF"
	case ASTHaltCompiler:
		return "HALT_COMPILER"
	case ASTEcho:
		return "ECHO"
	case ASTThrow:
		return "THROW"
	case ASTGoto:
		return "GOTO"
	case ASTBreak:
		return "BREAK"
	case ASTContinue:
		return "CONTINUE"
	case ASTPropertyHookShortBody:
		return "PROPERTY_HOOK_SHORT_BODY"
	case ASTClone:
		return "CLONE"
	case ASTExit:
		return "EXIT"
	case ASTList:
		return "LIST"
	case ASTAnonymousFunction:
		return "ANONYMOUS_FUNCTION"

	// 2子节点
	case ASTDim:
		return "DIM"
	case ASTProp:
		return "PROP"
	case ASTNullsafeProp:
		return "NULLSAFE_PROP"
	case ASTStaticProp:
		return "STATIC_PROP"
	case ASTCall:
		return "CALL"
	case ASTClassConst:
		return "CLASS_CONST"
	case ASTAssign:
		return "ASSIGN"
	case ASTAssignRef:
		return "ASSIGN_REF"
	case ASTAssignOp:
		return "ASSIGN_OP"
	case ASTBinaryOp:
		return "BINARY_OP"
	case ASTGreater:
		return "GREATER"
	case ASTGreaterEqual:
		return "GREATER_EQUAL"
	case ASTAnd:
		return "AND"
	case ASTOr:
		return "OR"
	case ASTArrayElem:
		return "ARRAY_ELEM"
	case ASTNew:
		return "NEW"
	case ASTInstanceof:
		return "INSTANCEOF"
	case ASTYield:
		return "YIELD"
	case ASTCoalesce:
		return "COALESCE"
	case ASTAssignCoalesce:
		return "ASSIGN_COALESCE"
	case ASTStatic:
		return "STATIC"
	case ASTWhile:
		return "WHILE"
	case ASTDoWhile:
		return "DO_WHILE"
	case ASTIfElem:
		return "IF_ELEM"
	case ASTSwitch:
		return "SWITCH"
	case ASTSwitchCase:
		return "SWITCH_CASE"
	case ASTDeclare:
		return "DECLARE"
	case ASTAltIf:
		return "ALT_IF"
	case ASTAltWhile:
		return "ALT_WHILE"
	case ASTAltFor:
		return "ALT_FOR"
	case ASTAltForeach:
		return "ALT_FOREACH"
	case ASTElseIf:
		return "ELSEIF"
	case ASTUseTrait:
		return "USE_TRAIT"
	case ASTTraitPrecedence:
		return "TRAIT_PRECEDENCE"
	case ASTMethodReference:
		return "METHOD_REFERENCE"
	case ASTNamespace:
		return "NAMESPACE"
	case ASTNamespaceName:
		return "NAMESPACE_NAME"
	case ASTUseElem:
		return "USE_ELEM"
	case ASTTraitAlias:
		return "TRAIT_ALIAS"
	case ASTGroupUse:
		return "GROUP_USE"
	case ASTAttribute:
		return "ATTRIBUTE"
	case ASTMatch:
		return "MATCH"
	case ASTMatchArm:
		return "MATCH_ARM"
	case ASTNamedArg:
		return "NAMED_ARG"
	case ASTParentPropertyHookCall:
		return "PARENT_PROPERTY_HOOK_CALL"
	case ASTPipe:
		return "PIPE"
	case ASTInterface:
		return "INTERFACE"
	case ASTTrait:
		return "TRAIT"
	case ASTEnum:
		return "ENUM"

	// 3子节点
	case ASTMethodCall:
		return "METHOD_CALL"
	case ASTNullsafeMethodCall:
		return "NULLSAFE_METHOD_CALL"
	case ASTStaticCall:
		return "STATIC_CALL"
	case ASTConditional:
		return "CONDITIONAL"
	case ASTTry:
		return "TRY"
	case ASTCatch:
		return "CATCH"
	case ASTPropGroup:
		return "PROP_GROUP"
	case ASTConstElem:
		return "CONST_ELEM"
	case ASTClassConstGroup:
		return "CLASS_CONST_GROUP"
	case ASTConstEnumInit:
		return "CONST_ENUM_INIT"
	case ASTCase:
		return "CASE"
	case ASTEvalExpression:
		return "EVAL_EXPRESSION"
	case ASTVisibilityModifier:
		return "VISIBILITY_MODIFIER"

	// 4子节点
	case ASTFor:
		return "FOR"
	case ASTForeach:
		return "FOREACH"
	case ASTEnumCase:
		return "ENUM_CASE"
	case ASTPropElem:
		return "PROP_ELEM"

	// 6子节点
	case ASTParam:
		return "PARAM"

	default:
		return "UNKNOWN"
	}
}

// IsSpecial 检查是否为特殊节点
func (k ASTKind) IsSpecial() bool {
	return (uint16(k)>>6)&1 == 1
}

// IsList 检查是否为列表节点
func (k ASTKind) IsList() bool {
	return (uint16(k)>>7)&1 == 1
}

// IsDecl 检查是否为声明节点
func (k ASTKind) IsDecl() bool {
	return k.IsSpecial() && k >= ASTFuncDecl
}

// GetNumChildren 获取子节点数量（对于非特殊、非列表节点）
func (k ASTKind) GetNumChildren() uint32 {
	if k.IsList() || k.IsSpecial() {
		return 0 // 列表和特殊节点的子节点数量是动态的
	}
	return uint32(k) >> 8
}