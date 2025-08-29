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
	T_EOF TokenType = 0 /* end of file (T_EOF) */

	T_LNUMBER TokenType = iota + 126 /* "integer number (T_LNUMBER)"  */

	T_DNUMBER                                 /* "floating-point number (T_DNUMBER)"  */
	T_STRING                                  /* "identifier (T_STRING)"  */
	T_VARIABLE                                /* "variable (T_VARIABLE)"  */
	T_INLINE_HTML                             /* T_INLINE_HTML  */
	T_ENCAPSED_AND_WHITESPACE                 /* "quoted-string and whitespace (T_ENCAPSED_AND_WHITESPACE)"  */
	T_CONSTANT_ENCAPSED_STRING                /* "quoted-string (T_CONSTANT_ENCAPSED_STRING)"  */
	T_STRING_VARNAME                          /* "variable name (T_STRING_VARNAME)"  */
	T_NUM_STRING                              /* "number (T_NUM_STRING)"  */
	T_INCLUDE                                 /* "include (T_INCLUDE)"  */
	T_INCLUDE_ONCE                            /* "include_once (T_INCLUDE_ONCE)"  */
	T_EVAL                                    /* "eval (T_EVAL)"  */
	T_REQUIRE                                 /* "require (T_REQUIRE)"  */
	T_REQUIRE_ONCE                            /* "require_once (T_REQUIRE_ONCE)"  */
	T_LOGICAL_OR                              /* "or (T_LOGICAL_OR)"  */
	T_LOGICAL_XOR                             /* "xor (T_LOGICAL_XOR)"  */
	T_LOGICAL_AND                             /* "and (T_LOGICAL_AND)"  */
	T_PRINT                                   /* "print (T_PRINT)"  */
	T_YIELD                                   /* "yield (T_YIELD)"  */
	T_YIELD_FROM                              /* "yield from (T_YIELD_FROM)"  */
	T_PLUS_EQUAL                              /* "+= (T_PLUS_EQUAL)"  */
	T_MINUS_EQUAL                             /* "-= (T_MINUS_EQUAL)"  */
	T_MUL_EQUAL                               /* "*= (T_MUL_EQUAL)"  */
	T_DIV_EQUAL                               /* "/= (T_DIV_EQUAL)"  */
	T_CONCAT_EQUAL                            /* ".= (T_CONCAT_EQUAL)"  */
	T_MOD_EQUAL                               /* "%= (T_MOD_EQUAL)"  */
	T_AND_EQUAL                               /* "&= (T_AND_EQUAL)"  */
	T_OR_EQUAL                                /* "|= (T_OR_EQUAL)"  */
	T_XOR_EQUAL                               /* "^= (T_XOR_EQUAL)"  */
	T_SL_EQUAL                                /* "<<= (T_SL_EQUAL)"  */
	T_SR_EQUAL                                /* ">>= (T_SR_EQUAL)"  */
	T_COALESCE_EQUAL                          /* "??= (T_COALESCE_EQUAL)"  */
	T_BOOLEAN_OR                              /* "|| (T_BOOLEAN_OR)"  */
	T_BOOLEAN_AND                             /* "&& (T_BOOLEAN_AND)"  */
	T_IS_EQUAL                                /* "== (T_IS_EQUAL)"  */
	T_IS_NOT_EQUAL                            /* "!= (T_IS_NOT_EQUAL)"  */
	T_IS_IDENTICAL                            /* "=== (T_IS_IDENTICAL)"  */
	T_IS_NOT_IDENTICAL                        /* "!== (T_IS_NOT_IDENTICAL)"  */
	T_IS_SMALLER_OR_EQUAL                     /* "<= (T_IS_SMALLER_OR_EQUAL)"  */
	T_IS_GREATER_OR_EQUAL                     /* ">= (T_IS_GREATER_OR_EQUAL)"  */
	T_SPACESHIP                               /* "<=> (T_SPACESHIP)"  */
	T_SL                                      /* "<< (T_SL)"  */
	T_SR                                      /* ">> (T_SR)"  */
	T_INSTANCEOF                              /* "instanceof (T_INSTANCEOF)"  */
	T_INC                                     /* "++ (T_INC)"  */
	T_DEC                                     /* "-- (T_DEC)"  */
	T_INT_CAST                                /* "(int) (T_INT_CAST)"  */
	T_DOUBLE_CAST                             /* "(double) (T_DOUBLE_CAST)"  */
	T_STRING_CAST                             /* "(string) (T_STRING_CAST)"  */
	T_ARRAY_CAST                              /* "(array) (T_ARRAY_CAST)"  */
	T_OBJECT_CAST                             /* "(object) (T_OBJECT_CAST)"  */
	T_BOOL_CAST                               /* "(bool) (T_BOOL_CAST)"  */
	T_UNSET_CAST                              /* "(unset) (T_UNSET_CAST)"  */
	T_VOID_CAST                               /* "(void) (T_VOID_CAST)"  */
	T_NEW                                     /* "new (T_NEW)"  */
	T_CLONE                                   /* "clone (T_CLONE)"  */
	T_EXIT                                    /* "exit (T_EXIT)"  */
	T_IF                                      /* "if (T_IF)"  */
	T_ELSEIF                                  /* "elseif (T_ELSEIF)"  */
	T_ELSE                                    /* "else (T_ELSE)"  */
	T_ENDIF                                   /* "endif (T_ENDIF)"  */
	T_ECHO                                    /* "echo (T_ECHO)"  */
	T_DO                                      /* "do (T_DO)"  */
	T_WHILE                                   /* "while (T_WHILE)"  */
	T_ENDWHILE                                /* "endwhile (T_ENDWHILE)"  */
	T_FOR                                     /* "for (T_FOR)"  */
	T_ENDFOR                                  /* "endfor (T_ENDFOR)"  */
	T_FOREACH                                 /* "foreach (T_FOREACH)"  */
	T_ENDFOREACH                              /* "endforeach (T_ENDFOREACH)"  */
	T_DECLARE                                 /* "declare (T_DECLARE)"  */
	T_ENDDECLARE                              /* "enddeclare (T_ENDDECLARE)"  */
	T_AS                                      /* "as (T_AS)"  */
	T_SWITCH                                  /* "switch (T_SWITCH)"  */
	T_ENDSWITCH                               /* "endswitch (T_ENDSWITCH)"  */
	T_CASE                                    /* "case (T_CASE)"  */
	T_DEFAULT                                 /* "default (T_DEFAULT)"  */
	T_MATCH                                   /* "match (T_MATCH)"  */
	T_BREAK                                   /* "break (T_BREAK)"  */
	T_CONTINUE                                /* "continue (T_CONTINUE)"  */
	T_GOTO                                    /* "goto (T_GOTO)"  */
	T_FUNCTION                                /* "function (T_FUNCTION)"  */
	T_FN                                      /* "fn (T_FN)"  */
	T_CONST                                   /* "const (T_CONST)"  */
	T_RETURN                                  /* "return (T_RETURN)"  */
	T_TRY                                     /* "try (T_TRY)"  */
	T_CATCH                                   /* "catch (T_CATCH)"  */
	T_FINALLY                                 /* "finally (T_FINALLY)"  */
	T_THROW                                   /* "throw (T_THROW)"  */
	T_USE                                     /* "use (T_USE)"  */
	T_INSTEADOF                               /* "insteadof (T_INSTEADOF)"  */
	T_GLOBAL                                  /* "global (T_GLOBAL)"  */
	T_STATIC                                  /* "static (T_STATIC)"  */
	T_ABSTRACT                                /* "abstract (T_ABSTRACT)"  */
	T_FINAL                                   /* "final (T_FINAL)"  */
	T_PRIVATE                                 /* "private (T_PRIVATE)"  */
	T_PROTECTED                               /* "protected (T_PROTECTED)"  */
	T_PUBLIC                                  /* "public (T_PUBLIC)"  */
	T_PRIVATE_SET                             /* "private(set) (T_PRIVATE_SET)"  */
	T_PROTECTED_SET                           /* "protected(set) (T_PROTECTED_SET)"  */
	T_PUBLIC_SET                              /* "public(set) (T_PUBLIC_SET)"  */
	T_READONLY                                /* "readonly (T_READONLY)"  */
	T_VAR                                     /* "var (T_VAR)"  */
	T_UNSET                                   /* "unset (T_UNSET)"  */
	T_ISSET                                   /* "isset (T_ISSET)"  */
	T_EMPTY                                   /* "empty (T_EMPTY)"  */
	T_HALT_COMPILER                           /* "__halt_compiler (T_HALT_COMPILER)"  */
	T_CLASS                                   /* "class (T_CLASS)"  */
	T_TRAIT                                   /* "trait (T_TRAIT)"  */
	T_INTERFACE                               /* "interface (T_INTERFACE)"  */
	T_ENUM                                    /* "enum (T_ENUM)"  */
	T_EXTENDS                                 /* "extends (T_EXTENDS)"  */
	T_IMPLEMENTS                              /* "implements (T_IMPLEMENTS)"  */
	T_OBJECT_OPERATOR                         /* "-> (T_OBJECT_OPERATOR)"  */
	T_NULLSAFE_OBJECT_OPERATOR                /* "?-> (T_NULLSAFE_OBJECT_OPERATOR)"  */
	T_DOUBLE_ARROW                            /* "=> (T_DOUBLE_ARROW)"  */
	T_LIST                                    /* "list (T_LIST)"  */
	T_ARRAY                                   /* "array (T_ARRAY)"  */
	T_CALLABLE                                /* "callable (T_CALLABLE)"  */
	T_LINE                                    /* "__LINE__ (T_LINE)"  */
	T_FILE                                    /* "__FILE__ (T_FILE)"  */
	T_DIR                                     /* "__DIR__ (T_DIR)"  */
	T_CLASS_C                                 /* "__CLASS__ (T_CLASS_C)"  */
	T_TRAIT_C                                 /* "__TRAIT__ (T_TRAIT_C)"  */
	T_METHOD_C                                /* "__METHOD__ (T_METHOD_C)"  */
	T_FUNC_C                                  /* "__FUNCTION__ (T_FUNC_C)"  */
	T_COMMENT                                 /* "comment (T_COMMENT)"  */
	T_DOC_COMMENT                             /* "doc comment (T_DOC_COMMENT)"  */
	T_OPEN_TAG                                /* "open tag (T_OPEN_TAG)"  */
	T_OPEN_TAG_WITH_ECHO                      /* "open tag with echo (T_OPEN_TAG_WITH_ECHO)"  */
	T_CLOSE_TAG                               /* "close tag (T_CLOSE_TAG)"  */
	T_WHITESPACE                              /* "whitespace (T_WHITESPACE)"  */
	T_START_HEREDOC                           /* "heredoc start (T_START_HEREDOC)"  */
	T_END_HEREDOC                             /* "heredoc end (T_END_HEREDOC)"  */
	T_DOLLAR_OPEN_CURLY_BRACES                /* "${ (T_DOLLAR_OPEN_CURLY_BRACES)"  */
	T_CURLY_OPEN                              /* "{$ (T_CURLY_OPEN)"  */
	T_PAAMAYIM_NEKUDOTAYIM                    /* ":: (T_PAAMAYIM_NEKUDOTAYIM)"  */
	T_NAMESPACE                               /* "namespace (T_NAMESPACE)"  */
	T_PROPERTY_C                              /* "__PROPERTY__ (T_PROPERTY_C)"  */
	T_NS_C                                    /* "__NAMESPACE__ (T_NS_C)"  */
	T_ATTRIBUTE                               /* "#[ (T_ATTRIBUTE)"  */
	T_NS_SEPARATOR                            /* "\\ (T_NS_SEPARATOR)"  */
	T_ELLIPSIS                                /* "... (T_ELLIPSIS)"  */
	T_COALESCE                                /* "?? (T_COALESCE)"  */
	T_POW                                     /* "** (T_POW)"  */
	T_POW_EQUAL                               /* "**= (T_POW_EQUAL)"  */
	T_BAD_CHARACTER                           /* "invalid character (T_BAD_CHARACTER)"  */
	T_AMPERSAND_FOLLOWED_BY_VAR_OR_VARARG     /* "&$ or &... (T_AMPERSAND_FOLLOWED_BY_VAR_OR_VARARG)"  */
	T_AMPERSAND_NOT_FOLLOWED_BY_VAR_OR_VARARG /* "& not followed by $ or ... (T_AMPERSAND_NOT_FOLLOWED_BY_VAR_OR_VARARG)"  */
	T_NAME_FULLY_QUALIFIED                    /* "name fully qualified (T_NAME_FULLY_QUALIFIED)"  */
	T_NAME_RELATIVE                           /* "namespace-relative name (T_NAME_RELATIVE)"  */
	T_NAME_QUALIFIED                          /* "namespaced name (T_NAME_QUALIFIED)"  */

	// 单个字符 token（为了完整性）
	TOKEN_SEMICOLON   TokenType = ';'  // ;
	TOKEN_COMMA       TokenType = ','  // ,
	TOKEN_DOT         TokenType = '.'  // .
	TOKEN_LBRACE      TokenType = '{'  // {
	TOKEN_RBRACE      TokenType = '}'  // }
	TOKEN_LPAREN      TokenType = '('  // (
	TOKEN_RPAREN      TokenType = ')'  // )
	TOKEN_LBRACKET    TokenType = '['  // [
	TOKEN_RBRACKET    TokenType = ']'  // ]
	TOKEN_PLUS        TokenType = '+'  // +
	TOKEN_MINUS       TokenType = '-'  // -
	TOKEN_MULTIPLY    TokenType = '*'  // *
	TOKEN_DIVIDE      TokenType = '/'  // /
	TOKEN_MODULO      TokenType = '%'  // %
	TOKEN_AMPERSAND   TokenType = '&'  // &
	TOKEN_PIPE        TokenType = '|'  // |
	TOKEN_CARET       TokenType = '^'  // ^
	TOKEN_TILDE       TokenType = '~'  // ~
	TOKEN_LT          TokenType = '<'  // <
	TOKEN_GT          TokenType = '>'  // >
	TOKEN_EQUAL       TokenType = '='  // =
	TOKEN_EXCLAMATION TokenType = '!'  // !
	TOKEN_QUESTION    TokenType = '?'  // ?
	TOKEN_COLON       TokenType = ':'  // :
	TOKEN_AT          TokenType = '@'  // @
	TOKEN_DOLLAR      TokenType = '$'  // $
	TOKEN_BACKSLASH   TokenType = '\\' // \
	TOKEN_QUOTE       TokenType = '"'  // "
	TOKEN_BACKTICK    TokenType = '`'  // `
)

// TokenNames 提供 Token 类型到名称的映射
var TokenNames = map[TokenType]string{
	T_EOF:                                 "T_EOF",
	T_LNUMBER:                             "T_LNUMBER",
	T_DNUMBER:                             "T_DNUMBER",
	T_STRING:                              "T_STRING",
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
	T_NAME_FULLY_QUALIFIED:                "T_NAME_FULLY_QUALIFIED",
	T_NAME_RELATIVE:                       "T_NAME_RELATIVE",
	T_NAME_QUALIFIED:                      "T_NAME_QUALIFIED",
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
	T_BAD_CHARACTER: "T_BAD_CHARACTER",

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
	TOKEN_BACKTICK:    "`",
}

// Keywords 定义 PHP 关键字到 Token 类型的映射
var Keywords = map[string]TokenType{
	"include":         T_INCLUDE,
	"include_once":    T_INCLUDE_ONCE,
	"eval":            T_EVAL,
	"require":         T_REQUIRE,
	"require_once":    T_REQUIRE_ONCE,
	"or":              T_LOGICAL_OR,
	"xor":             T_LOGICAL_XOR,
	"and":             T_LOGICAL_AND,
	"print":           T_PRINT,
	"yield":           T_YIELD,
	"instanceof":      T_INSTANCEOF,
	"new":             T_NEW,
	"clone":           T_CLONE,
	"exit":            T_EXIT,
	"die":             T_EXIT,
	"if":              T_IF,
	"elseif":          T_ELSEIF,
	"else":            T_ELSE,
	"endif":           T_ENDIF,
	"echo":            T_ECHO,
	"do":              T_DO,
	"while":           T_WHILE,
	"endwhile":        T_ENDWHILE,
	"for":             T_FOR,
	"endfor":          T_ENDFOR,
	"foreach":         T_FOREACH,
	"endforeach":      T_ENDFOREACH,
	"declare":         T_DECLARE,
	"enddeclare":      T_ENDDECLARE,
	"as":              T_AS,
	"switch":          T_SWITCH,
	"endswitch":       T_ENDSWITCH,
	"case":            T_CASE,
	"default":         T_DEFAULT,
	"match":           T_MATCH,
	"break":           T_BREAK,
	"continue":        T_CONTINUE,
	"goto":            T_GOTO,
	"function":        T_FUNCTION,
	"fn":              T_FN,
	"const":           T_CONST,
	"return":          T_RETURN,
	"try":             T_TRY,
	"catch":           T_CATCH,
	"finally":         T_FINALLY,
	"throw":           T_THROW,
	"use":             T_USE,
	"insteadof":       T_INSTEADOF,
	"global":          T_GLOBAL,
	"static":          T_STATIC,
	"abstract":        T_ABSTRACT,
	"final":           T_FINAL,
	"private":         T_PRIVATE,
	"protected":       T_PROTECTED,
	"public":          T_PUBLIC,
	"readonly":        T_READONLY,
	"var":             T_VAR,
	"unset":           T_UNSET,
	"isset":           T_ISSET,
	"empty":           T_EMPTY,
	"class":           T_CLASS,
	"trait":           T_TRAIT,
	"interface":       T_INTERFACE,
	"enum":            T_ENUM,
	"extends":         T_EXTENDS,
	"implements":      T_IMPLEMENTS,
	"list":            T_LIST,
	"array":           T_ARRAY,
	"callable":        T_CALLABLE,
	"namespace":       T_NAMESPACE,
	"__halt_compiler": T_HALT_COMPILER,
	// 属性钩子关键字 (Property Hooks) - PHP 8.4
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

// String 返回 Position 的字符串表示
func (p Position) String() string {
	return fmt.Sprintf("Line %d, Column %d, Offset %d", p.Line, p.Column, p.Offset)
}
