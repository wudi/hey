package lexer

import "fmt"

// TokenType 表示 PHP Token 类型，与 PHP 官方保持一致
type TokenType int

// Position 表示 Token 在源代码中的位置
type Position struct {
	Line   int // 行号（从1开始）
	Column int // 列号（从1开始）
	Offset int // 字节偏移（从0开始）
}

// Token 表示一个词法单元
type Token struct {
	Type     TokenType // Token 类型
	Value    string    // Token 字符串值
	Position Position  // Token 位置
}

// String 返回 Token 的字符串表示
func (t Token) String() string {
	return fmt.Sprintf("Token{Type: %s, Value: %q, Pos: %d:%d}",
		TokenNames[t.Type], t.Value, t.Position.Line, t.Position.Column)
}

// PHP Token 类型常量，与 PHP 8.4 官方保持一致
const (
	// 特殊 Token
	T_UNKNOWN TokenType = iota
	T_EOF

	// PHP 官方 Token 常量（按 PHP 源码中的值）
	T_LNUMBER                  TokenType = 260 // 整数
	T_DNUMBER                  TokenType = 261 // 浮点数
	T_STRING                   TokenType = 262 // 标识符
	T_NAME_FULLY_QUALIFIED     TokenType = 263 // \Foo\Bar
	T_NAME_RELATIVE            TokenType = 264 // namespace\Foo\Bar
	T_NAME_QUALIFIED           TokenType = 265 // Foo\Bar
	T_VARIABLE                 TokenType = 266 // $var
	T_INLINE_HTML              TokenType = 267 // HTML 代码
	T_ENCAPSED_AND_WHITESPACE  TokenType = 268 // 字符串中的内容
	T_CONSTANT_ENCAPSED_STRING TokenType = 269 // 字符串常量
	T_STRING_VARNAME           TokenType = 270 // 字符串中的变量名
	T_NUM_STRING               TokenType = 271 // 数字字符串

	// 语言结构
	T_INCLUDE      TokenType = 272
	T_INCLUDE_ONCE TokenType = 273
	T_EVAL         TokenType = 274
	T_REQUIRE      TokenType = 275
	T_REQUIRE_ONCE TokenType = 276

	// 逻辑操作符
	T_LOGICAL_OR  TokenType = 277 // or
	T_LOGICAL_XOR TokenType = 278 // xor
	T_LOGICAL_AND TokenType = 279 // and
	T_PRINT       TokenType = 280

	// 生成器
	T_YIELD      TokenType = 281
	T_YIELD_FROM TokenType = 282

	// 类型检查
	T_INSTANCEOF TokenType = 283

	// 对象操作
	T_NEW   TokenType = 284
	T_CLONE TokenType = 285

	// 退出
	T_EXIT TokenType = 286

	// 控制结构
	T_IF         TokenType = 287
	T_ELSEIF     TokenType = 288
	T_ELSE       TokenType = 289
	T_ENDIF      TokenType = 290
	T_ECHO       TokenType = 291
	T_DO         TokenType = 292
	T_WHILE      TokenType = 293
	T_ENDWHILE   TokenType = 294
	T_FOR        TokenType = 295
	T_ENDFOR     TokenType = 296
	T_FOREACH    TokenType = 297
	T_ENDFOREACH TokenType = 298
	T_DECLARE    TokenType = 299
	T_ENDDECLARE TokenType = 300
	T_AS         TokenType = 301
	T_SWITCH     TokenType = 302
	T_ENDSWITCH  TokenType = 303
	T_CASE       TokenType = 304
	T_DEFAULT    TokenType = 305
	T_MATCH      TokenType = 306
	T_BREAK      TokenType = 307
	T_CONTINUE   TokenType = 308
	T_GOTO       TokenType = 309
	T_FUNCTION   TokenType = 310
	T_FN         TokenType = 311
	T_CONST      TokenType = 312
	T_RETURN     TokenType = 313
	T_TRY        TokenType = 314
	T_CATCH      TokenType = 315
	T_FINALLY    TokenType = 316
	T_THROW      TokenType = 317
	T_USE        TokenType = 318
	T_INSTEADOF  TokenType = 319
	T_GLOBAL     TokenType = 320
	T_STATIC     TokenType = 321
	T_ABSTRACT   TokenType = 322
	T_FINAL      TokenType = 323
	T_PRIVATE    TokenType = 324
	T_PROTECTED  TokenType = 325
	T_PUBLIC     TokenType = 326
	// 新的可见性修饰符 (PHP 8.4)
	T_PRIVATE_SET   TokenType = 327 // private(set)
	T_PROTECTED_SET TokenType = 328 // protected(set)
	T_PUBLIC_SET    TokenType = 329 // public(set)
	T_READONLY      TokenType = 330
	T_VAR           TokenType = 331

	// 类相关
	T_UNSET         TokenType = 332
	T_ISSET         TokenType = 333
	T_EMPTY         TokenType = 334
	T_HALT_COMPILER TokenType = 335
	T_CLASS         TokenType = 336
	T_TRAIT         TokenType = 337
	T_INTERFACE     TokenType = 338
	T_ENUM          TokenType = 339
	T_EXTENDS       TokenType = 340
	T_IMPLEMENTS    TokenType = 341
	T_LIST          TokenType = 342
	T_ARRAY         TokenType = 343

	// 类型相关
	T_CALLABLE   TokenType = 344
	T_LINE       TokenType = 345
	T_FILE       TokenType = 346
	T_DIR        TokenType = 347
	T_CLASS_C    TokenType = 348 // __CLASS__
	T_TRAIT_C    TokenType = 349 // __TRAIT__
	T_METHOD_C   TokenType = 350 // __METHOD__
	T_FUNC_C     TokenType = 351 // __FUNCTION__
	T_NS_C       TokenType = 352 // __NAMESPACE__
	T_PROPERTY_C TokenType = 353 // __PROPERTY__ (PHP 8.4)

	// 注释
	T_COMMENT     TokenType = 354
	T_DOC_COMMENT TokenType = 355
	
	// 属性钩子 (Property Hooks) - PHP 8.4
	T_GET TokenType = 414
	T_SET TokenType = 415

	// 开放和关闭标签
	T_OPEN_TAG           TokenType = 356 // <?php
	T_OPEN_TAG_WITH_ECHO TokenType = 357 // <?=
	T_CLOSE_TAG          TokenType = 358 // ?>

	// 空白字符
	T_WHITESPACE               TokenType = 359
	T_START_HEREDOC            TokenType = 360
	T_END_HEREDOC              TokenType = 361
	T_DOLLAR_OPEN_CURLY_BRACES TokenType = 362 // ${
	T_CURLY_OPEN               TokenType = 363 // {$

	// 字符串插值
	T_PAAMAYIM_NEKUDOTAYIM TokenType = 364 // ::
	T_NAMESPACE            TokenType = 365
	T_NS_SEPARATOR         TokenType = 366 // \

	// 数组访问
	T_ELLIPSIS TokenType = 367 // ...

	// 比较操作符
	T_IS_EQUAL            TokenType = 368 // ==
	T_IS_NOT_EQUAL        TokenType = 369 // !=
	T_IS_IDENTICAL        TokenType = 370 // ===
	T_IS_NOT_IDENTICAL    TokenType = 371 // !==
	T_IS_SMALLER_OR_EQUAL TokenType = 372 // <=
	T_IS_GREATER_OR_EQUAL TokenType = 373 // >=
	T_SPACESHIP           TokenType = 374 // <=>

	// 赋值操作符
	T_PLUS_EQUAL     TokenType = 375 // +=
	T_MINUS_EQUAL    TokenType = 376 // -=
	T_MUL_EQUAL      TokenType = 377 // *=
	T_DIV_EQUAL      TokenType = 378 // /=
	T_CONCAT_EQUAL   TokenType = 379 // .=
	T_MOD_EQUAL      TokenType = 380 // %=
	T_AND_EQUAL      TokenType = 381 // &=
	T_OR_EQUAL       TokenType = 382 // |=
	T_XOR_EQUAL      TokenType = 383 // ^=
	T_SL_EQUAL       TokenType = 384 // <<=
	T_SR_EQUAL       TokenType = 385 // >>=
	T_COALESCE_EQUAL TokenType = 386 // ??=

	// 增减操作符
	T_INC TokenType = 387 // ++
	T_DEC TokenType = 388 // --

	// 对象操作符 (正确位置)
	T_OBJECT_OPERATOR          TokenType = 389 // ->
	T_NULLSAFE_OBJECT_OPERATOR TokenType = 390 // ?->
	T_DOUBLE_ARROW             TokenType = 391 // =>

	// 逻辑操作符
	T_BOOLEAN_OR  TokenType = 392 // ||
	T_BOOLEAN_AND TokenType = 393 // &&

	// NULL 合并
	T_COALESCE TokenType = 394 // ??

	// 位移操作符 (正确位置)
	T_SL TokenType = 395 // <<
	T_SR TokenType = 396 // >>

	// 属性
	T_ATTRIBUTE TokenType = 397 // #[

	// 类型声明
	T_INT_CAST    TokenType = 398 // (int)
	T_DOUBLE_CAST TokenType = 399 // (double)
	T_STRING_CAST TokenType = 400 // (string)
	T_ARRAY_CAST  TokenType = 401 // (array)
	T_OBJECT_CAST TokenType = 402 // (object)
	T_BOOL_CAST   TokenType = 403 // (bool)
	T_UNSET_CAST  TokenType = 404 // (unset)
	T_VOID_CAST   TokenType = 405 // (void) - PHP 8.4

	// 更多操作符
	T_POW       TokenType = 406 // **
	T_POW_EQUAL TokenType = 407 // **=

	// 上下文敏感的 & 操作符
	T_AMPERSAND_FOLLOWED_BY_VAR_OR_VARARG     TokenType = 408
	T_AMPERSAND_NOT_FOLLOWED_BY_VAR_OR_VARARG TokenType = 409

	// Nowdoc 支持
	T_NOWDOC TokenType = 410 // <<<'IDENTIFIER'

	// 管道操作符 (PHP 8.4)
	T_PIPE TokenType = 411 // |>

	// 其他
	T_BAD_CHARACTER TokenType = 412
	T_CLOSE_TAG_ALT TokenType = 413 // 替代关闭标签

	// 单个字符 token（为了完整性）
	TOKEN_SEMICOLON   TokenType = 1000 + ';'  // ;
	TOKEN_COMMA       TokenType = 1000 + ','  // ,
	TOKEN_DOT         TokenType = 1000 + '.'  // .
	TOKEN_LBRACE      TokenType = 1000 + '{'  // {
	TOKEN_RBRACE      TokenType = 1000 + '}'  // }
	TOKEN_LPAREN      TokenType = 1000 + '('  // (
	TOKEN_RPAREN      TokenType = 1000 + ')'  // )
	TOKEN_LBRACKET    TokenType = 1000 + '['  // [
	TOKEN_RBRACKET    TokenType = 1000 + ']'  // ]
	TOKEN_PLUS        TokenType = 1000 + '+'  // +
	TOKEN_MINUS       TokenType = 1000 + '-'  // -
	TOKEN_MULTIPLY    TokenType = 1000 + '*'  // *
	TOKEN_DIVIDE      TokenType = 1000 + '/'  // /
	TOKEN_MODULO      TokenType = 1000 + '%'  // %
	TOKEN_AMPERSAND   TokenType = 1000 + '&'  // &
	TOKEN_PIPE        TokenType = 1000 + '|'  // |
	TOKEN_CARET       TokenType = 1000 + '^'  // ^
	TOKEN_TILDE       TokenType = 1000 + '~'  // ~
	TOKEN_LT          TokenType = 1000 + '<'  // <
	TOKEN_GT          TokenType = 1000 + '>'  // >
	TOKEN_EQUAL       TokenType = 1000 + '='  // =
	TOKEN_EXCLAMATION TokenType = 1000 + '!'  // !
	TOKEN_QUESTION    TokenType = 1000 + '?'  // ?
	TOKEN_COLON       TokenType = 1000 + ':'  // :
	TOKEN_AT          TokenType = 1000 + '@'  // @
	TOKEN_DOLLAR      TokenType = 1000 + '$'  // $
	TOKEN_BACKSLASH   TokenType = 1000 + '\\' // \
	TOKEN_QUOTE       TokenType = 1000 + '"'  // "
)

// TokenNames 提供 Token 类型到名称的映射
var TokenNames = map[TokenType]string{
	T_UNKNOWN:                             "T_UNKNOWN",
	T_EOF:                                 "T_EOF",
	T_LNUMBER:                             "T_LNUMBER",
	T_DNUMBER:                             "T_DNUMBER",
	T_STRING:                              "T_STRING",
	T_NAME_FULLY_QUALIFIED:                "T_NAME_FULLY_QUALIFIED",
	T_NAME_RELATIVE:                       "T_NAME_RELATIVE",
	T_NAME_QUALIFIED:                      "T_NAME_QUALIFIED",
	T_VARIABLE:                            "T_VARIABLE",
	T_INLINE_HTML:                         "T_INLINE_HTML",
	T_ENCAPSED_AND_WHITESPACE:             "T_ENCAPSED_AND_WHITESPACE",
	T_CONSTANT_ENCAPSED_STRING:            "T_CONSTANT_ENCAPSED_STRING",
	T_STRING_VARNAME:                      "T_STRING_VARNAME",
	T_NUM_STRING:                          "T_NUM_STRING",
	T_INCLUDE:                             "T_INCLUDE",
	T_INCLUDE_ONCE:                        "T_INCLUDE_ONCE",
	T_EVAL:                                "T_EVAL",
	T_REQUIRE:                             "T_REQUIRE",
	T_REQUIRE_ONCE:                        "T_REQUIRE_ONCE",
	T_LOGICAL_OR:                          "T_LOGICAL_OR",
	T_LOGICAL_XOR:                         "T_LOGICAL_XOR",
	T_LOGICAL_AND:                         "T_LOGICAL_AND",
	T_PRINT:                               "T_PRINT",
	T_YIELD:                               "T_YIELD",
	T_YIELD_FROM:                          "T_YIELD_FROM",
	T_INSTANCEOF:                          "T_INSTANCEOF",
	T_NEW:                                 "T_NEW",
	T_CLONE:                               "T_CLONE",
	T_EXIT:                                "T_EXIT",
	T_IF:                                  "T_IF",
	T_ELSEIF:                              "T_ELSEIF",
	T_ELSE:                                "T_ELSE",
	T_ENDIF:                               "T_ENDIF",
	T_ECHO:                                "T_ECHO",
	T_DO:                                  "T_DO",
	T_WHILE:                               "T_WHILE",
	T_ENDWHILE:                            "T_ENDWHILE",
	T_FOR:                                 "T_FOR",
	T_ENDFOR:                              "T_ENDFOR",
	T_FOREACH:                             "T_FOREACH",
	T_ENDFOREACH:                          "T_ENDFOREACH",
	T_DECLARE:                             "T_DECLARE",
	T_ENDDECLARE:                          "T_ENDDECLARE",
	T_AS:                                  "T_AS",
	T_SWITCH:                              "T_SWITCH",
	T_ENDSWITCH:                           "T_ENDSWITCH",
	T_CASE:                                "T_CASE",
	T_DEFAULT:                             "T_DEFAULT",
	T_MATCH:                               "T_MATCH",
	T_BREAK:                               "T_BREAK",
	T_CONTINUE:                            "T_CONTINUE",
	T_GOTO:                                "T_GOTO",
	T_FUNCTION:                            "T_FUNCTION",
	T_FN:                                  "T_FN",
	T_CONST:                               "T_CONST",
	T_RETURN:                              "T_RETURN",
	T_TRY:                                 "T_TRY",
	T_CATCH:                               "T_CATCH",
	T_FINALLY:                             "T_FINALLY",
	T_THROW:                               "T_THROW",
	T_USE:                                 "T_USE",
	T_INSTEADOF:                           "T_INSTEADOF",
	T_GLOBAL:                              "T_GLOBAL",
	T_STATIC:                              "T_STATIC",
	T_ABSTRACT:                            "T_ABSTRACT",
	T_FINAL:                               "T_FINAL",
	T_PRIVATE:                             "T_PRIVATE",
	T_PROTECTED:                           "T_PROTECTED",
	T_PUBLIC:                              "T_PUBLIC",
	T_PRIVATE_SET:                         "T_PRIVATE_SET",
	T_PROTECTED_SET:                       "T_PROTECTED_SET",
	T_PUBLIC_SET:                          "T_PUBLIC_SET",
	T_READONLY:                            "T_READONLY",
	T_VAR:                                 "T_VAR",
	T_UNSET:                               "T_UNSET",
	T_ISSET:                               "T_ISSET",
	T_EMPTY:                               "T_EMPTY",
	T_HALT_COMPILER:                       "T_HALT_COMPILER",
	T_CLASS:                               "T_CLASS",
	T_TRAIT:                               "T_TRAIT",
	T_INTERFACE:                           "T_INTERFACE",
	T_ENUM:                                "T_ENUM",
	T_EXTENDS:                             "T_EXTENDS",
	T_IMPLEMENTS:                          "T_IMPLEMENTS",
	T_OBJECT_OPERATOR:                     "T_OBJECT_OPERATOR",
	T_NULLSAFE_OBJECT_OPERATOR:            "T_NULLSAFE_OBJECT_OPERATOR",
	T_DOUBLE_ARROW:                        "T_DOUBLE_ARROW",
	T_LIST:                                "T_LIST",
	T_ARRAY:                               "T_ARRAY",
	T_CALLABLE:                            "T_CALLABLE",
	T_LINE:                                "T_LINE",
	T_FILE:                                "T_FILE",
	T_DIR:                                 "T_DIR",
	T_CLASS_C:                             "T_CLASS_C",
	T_TRAIT_C:                             "T_TRAIT_C",
	T_METHOD_C:                            "T_METHOD_C",
	T_FUNC_C:                              "T_FUNC_C",
	T_NS_C:                                "T_NS_C",
	T_PROPERTY_C:                          "T_PROPERTY_C",
	T_COMMENT:                             "T_COMMENT",
	T_DOC_COMMENT:                         "T_DOC_COMMENT",
	T_OPEN_TAG:                            "T_OPEN_TAG",
	T_OPEN_TAG_WITH_ECHO:                  "T_OPEN_TAG_WITH_ECHO",
	T_CLOSE_TAG:                           "T_CLOSE_TAG",
	T_WHITESPACE:                          "T_WHITESPACE",
	T_START_HEREDOC:                       "T_START_HEREDOC",
	T_END_HEREDOC:                         "T_END_HEREDOC",
	T_DOLLAR_OPEN_CURLY_BRACES:            "T_DOLLAR_OPEN_CURLY_BRACES",
	T_CURLY_OPEN:                          "T_CURLY_OPEN",
	T_PAAMAYIM_NEKUDOTAYIM:                "T_PAAMAYIM_NEKUDOTAYIM",
	T_NAMESPACE:                           "T_NAMESPACE",
	T_NS_SEPARATOR:                        "T_NS_SEPARATOR",
	T_ELLIPSIS:                            "T_ELLIPSIS",
	T_IS_EQUAL:                            "T_IS_EQUAL",
	T_IS_NOT_EQUAL:                        "T_IS_NOT_EQUAL",
	T_IS_IDENTICAL:                        "T_IS_IDENTICAL",
	T_IS_NOT_IDENTICAL:                    "T_IS_NOT_IDENTICAL",
	T_IS_SMALLER_OR_EQUAL:                 "T_IS_SMALLER_OR_EQUAL",
	T_IS_GREATER_OR_EQUAL:                 "T_IS_GREATER_OR_EQUAL",
	T_SPACESHIP:                           "T_SPACESHIP",
	T_PLUS_EQUAL:                          "T_PLUS_EQUAL",
	T_MINUS_EQUAL:                         "T_MINUS_EQUAL",
	T_MUL_EQUAL:                           "T_MUL_EQUAL",
	T_DIV_EQUAL:                           "T_DIV_EQUAL",
	T_CONCAT_EQUAL:                        "T_CONCAT_EQUAL",
	T_MOD_EQUAL:                           "T_MOD_EQUAL",
	T_AND_EQUAL:                           "T_AND_EQUAL",
	T_OR_EQUAL:                            "T_OR_EQUAL",
	T_XOR_EQUAL:                           "T_XOR_EQUAL",
	T_SL_EQUAL:                            "T_SL_EQUAL",
	T_SR_EQUAL:                            "T_SR_EQUAL",
	T_COALESCE_EQUAL:                      "T_COALESCE_EQUAL",
	T_INC:                                 "T_INC",
	T_DEC:                                 "T_DEC",
	T_BOOLEAN_OR:                          "T_BOOLEAN_OR",
	T_BOOLEAN_AND:                         "T_BOOLEAN_AND",
	T_COALESCE:                            "T_COALESCE",
	T_SL:                                  "T_SL",
	T_SR:                                  "T_SR",
	T_ATTRIBUTE:                           "T_ATTRIBUTE",
	T_INT_CAST:                            "T_INT_CAST",
	T_DOUBLE_CAST:                         "T_DOUBLE_CAST",
	T_STRING_CAST:                         "T_STRING_CAST",
	T_ARRAY_CAST:                          "T_ARRAY_CAST",
	T_OBJECT_CAST:                         "T_OBJECT_CAST",
	T_BOOL_CAST:                           "T_BOOL_CAST",
	T_UNSET_CAST:                          "T_UNSET_CAST",
	T_VOID_CAST:                           "T_VOID_CAST",
	T_POW:                                 "T_POW",
	T_POW_EQUAL:                           "T_POW_EQUAL",
	T_AMPERSAND_FOLLOWED_BY_VAR_OR_VARARG: "T_AMPERSAND_FOLLOWED_BY_VAR_OR_VARARG",
	T_AMPERSAND_NOT_FOLLOWED_BY_VAR_OR_VARARG: "T_AMPERSAND_NOT_FOLLOWED_BY_VAR_OR_VARARG",
	T_NOWDOC:        "T_NOWDOC",
	T_PIPE:          "T_PIPE",
	T_BAD_CHARACTER: "T_BAD_CHARACTER",
	T_CLOSE_TAG_ALT: "T_CLOSE_TAG_ALT",

	// 单字符 token
	TOKEN_SEMICOLON:   ";",
	TOKEN_COMMA:       ",",
	TOKEN_DOT:         ".",
	TOKEN_LBRACE:      "{",
	TOKEN_RBRACE:      "}",
	TOKEN_LPAREN:      "(",
	TOKEN_RPAREN:      ")",
	TOKEN_LBRACKET:    "[",
	TOKEN_RBRACKET:    "]",
	TOKEN_PLUS:        "+",
	TOKEN_MINUS:       "-",
	TOKEN_MULTIPLY:    "*",
	TOKEN_DIVIDE:      "/",
	TOKEN_MODULO:      "%",
	TOKEN_AMPERSAND:   "&",
	TOKEN_PIPE:        "|",
	TOKEN_CARET:       "^",
	TOKEN_TILDE:       "~",
	TOKEN_LT:          "<",
	TOKEN_GT:          ">",
	TOKEN_EQUAL:       "=",
	TOKEN_EXCLAMATION: "!",
	TOKEN_QUESTION:    "?",
	TOKEN_COLON:       ":",
	TOKEN_AT:          "@",
	TOKEN_DOLLAR:      "$",
	TOKEN_BACKSLASH:   "\\",
	TOKEN_QUOTE:       "\"",
}

// Keywords 定义 PHP 关键字到 Token 类型的映射
var Keywords = map[string]TokenType{
	"include":      T_INCLUDE,
	"include_once": T_INCLUDE_ONCE,
	"eval":         T_EVAL,
	"require":      T_REQUIRE,
	"require_once": T_REQUIRE_ONCE,
	"or":           T_LOGICAL_OR,
	"xor":          T_LOGICAL_XOR,
	"and":          T_LOGICAL_AND,
	"print":        T_PRINT,
	"yield":        T_YIELD,
	"instanceof":   T_INSTANCEOF,
	"new":          T_NEW,
	"clone":        T_CLONE,
	"exit":         T_EXIT,
	"die":          T_EXIT,
	"if":           T_IF,
	"elseif":       T_ELSEIF,
	"else":         T_ELSE,
	"endif":        T_ENDIF,
	"echo":         T_ECHO,
	"do":           T_DO,
	"while":        T_WHILE,
	"endwhile":     T_ENDWHILE,
	"for":          T_FOR,
	"endfor":       T_ENDFOR,
	"foreach":      T_FOREACH,
	"endforeach":   T_ENDFOREACH,
	"declare":      T_DECLARE,
	"enddeclare":   T_ENDDECLARE,
	"as":           T_AS,
	"switch":       T_SWITCH,
	"endswitch":    T_ENDSWITCH,
	"case":         T_CASE,
	"default":      T_DEFAULT,
	"match":        T_MATCH,
	"break":        T_BREAK,
	"continue":     T_CONTINUE,
	"goto":         T_GOTO,
	"function":     T_FUNCTION,
	"fn":           T_FN,
	"const":        T_CONST,
	"return":       T_RETURN,
	"try":          T_TRY,
	"catch":        T_CATCH,
	"finally":      T_FINALLY,
	"throw":        T_THROW,
	"use":          T_USE,
	"insteadof":    T_INSTEADOF,
	"global":       T_GLOBAL,
	"static":       T_STATIC,
	"abstract":     T_ABSTRACT,
	"final":        T_FINAL,
	"private":      T_PRIVATE,
	"protected":    T_PROTECTED,
	"public":       T_PUBLIC,
	"readonly":     T_READONLY,
	"var":          T_VAR,
	"unset":        T_UNSET,
	"isset":        T_ISSET,
	"empty":        T_EMPTY,
	"class":        T_CLASS,
	"trait":        T_TRAIT,
	"interface":    T_INTERFACE,
	"enum":         T_ENUM,
	"extends":      T_EXTENDS,
	"implements":   T_IMPLEMENTS,
	"list":         T_LIST,
	"array":        T_ARRAY,
	"callable":     T_CALLABLE,
	"namespace":    T_NAMESPACE,
	// 属性钩子关键字 (Property Hooks) - PHP 8.4
	"get":          T_GET,
	"set":          T_SET,
}

// IsKeyword 检查给定字符串是否为 PHP 关键字
func IsKeyword(s string) (TokenType, bool) {
	tokenType, exists := Keywords[s]
	return tokenType, exists
}

// String 返回 TokenType 的字符串表示
func (t TokenType) String() string {
	if name, ok := TokenNames[t]; ok {
		return name
	}
	return fmt.Sprintf("UNKNOWN(%d)", t)
}
