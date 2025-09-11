package ast

// ASTKind 定义AST节点类型，严格遵循PHP官方zend_ast.h定义
type ASTKind uint16

const (
	// 特殊节点 - special nodes (bit 6 set)
	ASTZval     ASTKind = 64 // 1 << 6 = ZEND_AST_ZVAL
	ASTConstant ASTKind = 65 // ZEND_AST_CONSTANT
	ASTZNode    ASTKind = 66 // ZEND_AST_ZNODE

	// 声明节点 - declaration nodes
	ASTFuncDecl     ASTKind = 67 // ZEND_AST_FUNC_DECL
	ASTClosure      ASTKind = 68 // ZEND_AST_CLOSURE
	ASTMethod       ASTKind = 69 // ZEND_AST_METHOD
	ASTClass        ASTKind = 70 // ZEND_AST_CLASS (also used for interface, trait, enum)
	ASTArrowFunc    ASTKind = 71 // ZEND_AST_ARROW_FUNC
	ASTPropertyHook ASTKind = 72 // ZEND_AST_PROPERTY_HOOK

	// 列表节点 - list nodes (bit 7 set)
	ASTArgList          ASTKind = 128 // 1 << 7 = ZEND_AST_ARG_LIST
	ASTArray            ASTKind = 129 // ZEND_AST_ARRAY
	ASTEncapsList       ASTKind = 130 // ZEND_AST_ENCAPS_LIST
	ASTExprList         ASTKind = 131 // ZEND_AST_EXPR_LIST
	ASTStmtList         ASTKind = 132 // ZEND_AST_STMT_LIST
	ASTIf               ASTKind = 133 // ZEND_AST_IF
	ASTSwitchList       ASTKind = 134 // ZEND_AST_SWITCH_LIST
	ASTCatchList        ASTKind = 135 // ZEND_AST_CATCH_LIST
	ASTParamList        ASTKind = 136 // ZEND_AST_PARAM_LIST
	ASTClosureUses      ASTKind = 137 // ZEND_AST_CLOSURE_USES
	ASTPropDecl         ASTKind = 138 // ZEND_AST_PROP_DECL
	ASTConstDecl        ASTKind = 139 // ZEND_AST_CONST_DECL
	ASTClassConstDecl   ASTKind = 140 // ZEND_AST_CLASS_CONST_DECL
	ASTNameList         ASTKind = 141 // ZEND_AST_NAME_LIST
	ASTTraitAdaptations ASTKind = 142 // ZEND_AST_TRAIT_ADAPTATIONS
	ASTUse              ASTKind = 143 // ZEND_AST_USE
	ASTTypeUnion        ASTKind = 144 // ZEND_AST_TYPE_UNION
	ASTTypeIntersection ASTKind = 145 // ZEND_AST_TYPE_INTERSECTION
	ASTAttributeList    ASTKind = 146 // ZEND_AST_ATTRIBUTE_LIST
	ASTAttributeGroup   ASTKind = 147 // ZEND_AST_ATTRIBUTE_GROUP
	ASTMatchArmList     ASTKind = 148 // ZEND_AST_MATCH_ARM_LIST
	ASTModifierList     ASTKind = 149 // ZEND_AST_MODIFIER_LIST

	// 0子节点 - 0 child nodes (bits 8-15 = 0)
	ASTMagicConst      ASTKind = 0 // ZEND_AST_MAGIC_CONST
	ASTType            ASTKind = 1 // ZEND_AST_TYPE
	ASTConstantClass   ASTKind = 2 // ZEND_AST_CONSTANT_CLASS
	ASTCallableConvert ASTKind = 3 // ZEND_AST_CALLABLE_CONVERT

	// 1子节点 - 1 child node (bits 8-15 = 1)
	ASTVar                   ASTKind = 256 // 1 << 8 = ZEND_AST_VAR
	ASTConst                 ASTKind = 257 // ZEND_AST_CONST
	ASTUnpack                ASTKind = 258 // ZEND_AST_UNPACK
	ASTUnaryPlus             ASTKind = 259 // ZEND_AST_UNARY_PLUS
	ASTUnaryMinus            ASTKind = 260 // ZEND_AST_UNARY_MINUS
	ASTCast                  ASTKind = 261 // ZEND_AST_CAST
	ASTEmpty                 ASTKind = 262 // ZEND_AST_EMPTY
	ASTIsset                 ASTKind = 263 // ZEND_AST_ISSET
	ASTSilence               ASTKind = 264 // ZEND_AST_SILENCE
	ASTShellExec             ASTKind = 265 // ZEND_AST_SHELL_EXEC
	ASTClone                 ASTKind = 266 // ZEND_AST_CLONE
	ASTExit                  ASTKind = 267 // ZEND_AST_EXIT
	ASTPrint                 ASTKind = 268 // ZEND_AST_PRINT
	ASTIncludeOrEval         ASTKind = 269 // ZEND_AST_INCLUDE_OR_EVAL
	ASTUnaryOp               ASTKind = 270 // ZEND_AST_UNARY_OP
	ASTPreInc                ASTKind = 271 // ZEND_AST_PRE_INC
	ASTPreDec                ASTKind = 272 // ZEND_AST_PRE_DEC
	ASTPostInc               ASTKind = 273 // ZEND_AST_POST_INC
	ASTPostDec               ASTKind = 274 // ZEND_AST_POST_DEC
	ASTYieldFrom             ASTKind = 275 // ZEND_AST_YIELD_FROM
	ASTClassName             ASTKind = 276 // ZEND_AST_CLASS_NAME
	ASTGlobal                ASTKind = 277 // ZEND_AST_GLOBAL
	ASTUnset                 ASTKind = 278 // ZEND_AST_UNSET
	ASTReturn                ASTKind = 279 // ZEND_AST_RETURN
	ASTLabel                 ASTKind = 280 // ZEND_AST_LABEL
	ASTRef                   ASTKind = 281 // ZEND_AST_REF
	ASTHaltCompiler          ASTKind = 282 // ZEND_AST_HALT_COMPILER
	ASTEcho                  ASTKind = 283 // ZEND_AST_ECHO
	ASTThrow                 ASTKind = 284 // ZEND_AST_THROW
	ASTGoto                  ASTKind = 285 // ZEND_AST_GOTO
	ASTBreak                 ASTKind = 286 // ZEND_AST_BREAK
	ASTContinue              ASTKind = 287 // ZEND_AST_CONTINUE
	ASTPropertyHookShortBody ASTKind = 288 // ZEND_AST_PROPERTY_HOOK_SHORT_BODY

	// 2子节点 - 2 child nodes (bits 8-15 = 2)
	ASTDim                    ASTKind = 512 // 2 << 8 = ZEND_AST_DIM
	ASTProp                   ASTKind = 513 // ZEND_AST_PROP
	ASTNullsafeProp           ASTKind = 514 // ZEND_AST_NULLSAFE_PROP
	ASTStaticProp             ASTKind = 515 // ZEND_AST_STATIC_PROP
	ASTCall                   ASTKind = 516 // ZEND_AST_CALL
	ASTClassConst             ASTKind = 517 // ZEND_AST_CLASS_CONST
	ASTAssign                 ASTKind = 518 // ZEND_AST_ASSIGN
	ASTAssignRef              ASTKind = 519 // ZEND_AST_ASSIGN_REF
	ASTAssignOp               ASTKind = 520 // ZEND_AST_ASSIGN_OP
	ASTBinaryOp               ASTKind = 521 // ZEND_AST_BINARY_OP
	ASTGreater                ASTKind = 522 // ZEND_AST_GREATER
	ASTGreaterEqual           ASTKind = 523 // ZEND_AST_GREATER_EQUAL
	ASTAnd                    ASTKind = 524 // ZEND_AST_AND
	ASTOr                     ASTKind = 525 // ZEND_AST_OR
	ASTArrayElem              ASTKind = 526 // ZEND_AST_ARRAY_ELEM
	ASTNew                    ASTKind = 527 // ZEND_AST_NEW
	ASTInstanceof             ASTKind = 528 // ZEND_AST_INSTANCEOF
	ASTYield                  ASTKind = 529 // ZEND_AST_YIELD
	ASTCoalesce               ASTKind = 530 // ZEND_AST_COALESCE
	ASTAssignCoalesce         ASTKind = 531 // ZEND_AST_ASSIGN_COALESCE
	ASTStatic                 ASTKind = 532 // ZEND_AST_STATIC
	ASTWhile                  ASTKind = 533 // ZEND_AST_WHILE
	ASTDoWhile                ASTKind = 534 // ZEND_AST_DO_WHILE
	ASTIfElem                 ASTKind = 535 // ZEND_AST_IF_ELEM
	ASTSwitch                 ASTKind = 536 // ZEND_AST_SWITCH
	ASTSwitchCase             ASTKind = 537 // ZEND_AST_SWITCH_CASE
	ASTDeclare                ASTKind = 538 // ZEND_AST_DECLARE
	ASTUseTrait               ASTKind = 539 // ZEND_AST_USE_TRAIT
	ASTTraitPrecedence        ASTKind = 540 // ZEND_AST_TRAIT_PRECEDENCE
	ASTMethodReference        ASTKind = 541 // ZEND_AST_METHOD_REFERENCE
	ASTNamespace              ASTKind = 542 // ZEND_AST_NAMESPACE
	ASTUseElem                ASTKind = 543 // ZEND_AST_USE_ELEM
	ASTTraitAlias             ASTKind = 544 // ZEND_AST_TRAIT_ALIAS
	ASTGroupUse               ASTKind = 545 // ZEND_AST_GROUP_USE
	ASTAttribute              ASTKind = 546 // ZEND_AST_ATTRIBUTE
	ASTMatch                  ASTKind = 547 // ZEND_AST_MATCH
	ASTMatchArm               ASTKind = 548 // ZEND_AST_MATCH_ARM
	ASTNamedArg               ASTKind = 549 // ZEND_AST_NAMED_ARG
	ASTParentPropertyHookCall ASTKind = 550 // ZEND_AST_PARENT_PROPERTY_HOOK_CALL

	// 3子节点 - 3 child nodes (bits 8-15 = 3)
	ASTMethodCall         ASTKind = 768 // 3 << 8 = ZEND_AST_METHOD_CALL
	ASTNullsafeMethodCall ASTKind = 769 // ZEND_AST_NULLSAFE_METHOD_CALL
	ASTStaticCall         ASTKind = 770 // ZEND_AST_STATIC_CALL
	ASTConditional        ASTKind = 771 // ZEND_AST_CONDITIONAL
	ASTTry                ASTKind = 772 // ZEND_AST_TRY
	ASTCatch              ASTKind = 773 // ZEND_AST_CATCH
	ASTPropGroup          ASTKind = 774 // ZEND_AST_PROP_GROUP
	ASTConstElem          ASTKind = 775 // ZEND_AST_CONST_ELEM
	ASTClassConstGroup    ASTKind = 776 // ZEND_AST_CLASS_CONST_GROUP
	ASTConstEnumInit      ASTKind = 777 // ZEND_AST_CONST_ENUM_INIT

	// 4子节点 - 4 child nodes (bits 8-15 = 4)
	ASTFor      ASTKind = 1024 // 4 << 8 = ZEND_AST_FOR
	ASTForeach  ASTKind = 1025 // ZEND_AST_FOREACH
	ASTEnumCase ASTKind = 1026 // ZEND_AST_ENUM_CASE
	ASTPropElem ASTKind = 1027 // ZEND_AST_PROP_ELEM

	// 6子节点 - 6 child nodes (bits 8-15 = 6)
	ASTParam ASTKind = 1536 // 6 << 8 = ZEND_AST_PARAM
)

// String 返回AST节点类型的字符串表示
func (k ASTKind) String() string {
	switch k {
	// 特殊节点
	case ASTZval:
		return "ZVAL"
	case ASTConstant:
		return "CONSTANT"
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
	case ASTEmpty:
		return "EMPTY"
	case ASTIsset:
		return "ISSET"
	case ASTSilence:
		return "SILENCE"
	case ASTShellExec:
		return "SHELL_EXEC"
	case ASTClone:
		return "CLONE"
	case ASTExit:
		return "EXIT"
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
	case ASTUseTrait:
		return "USE_TRAIT"
	case ASTTraitPrecedence:
		return "TRAIT_PRECEDENCE"
	case ASTMethodReference:
		return "METHOD_REFERENCE"
	case ASTNamespace:
		return "NAMESPACE"
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
