package stdlib

import (
	"github.com/wudi/php-parser/compiler/values"
)

// initConstants initializes built-in PHP constants
func (stdlib *StandardLibrary) initConstants() {
	// Core constants
	stdlib.Constants["PHP_VERSION"] = values.NewString("8.4.0")
	stdlib.Constants["PHP_MAJOR_VERSION"] = values.NewInt(8)
	stdlib.Constants["PHP_MINOR_VERSION"] = values.NewInt(4)
	stdlib.Constants["PHP_RELEASE_VERSION"] = values.NewInt(0)
	stdlib.Constants["PHP_VERSION_ID"] = values.NewInt(80400)
	stdlib.Constants["PHP_EXTRA_VERSION"] = values.NewString("")

	// Zend constants
	stdlib.Constants["ZEND_VERSION"] = values.NewString("4.4.0")
	stdlib.Constants["ZEND_DEBUG"] = values.NewInt(0)
	stdlib.Constants["ZEND_THREAD_SAFE"] = values.NewInt(0)

	// System constants
	stdlib.Constants["PHP_OS"] = values.NewString("Linux")
	stdlib.Constants["PHP_OS_FAMILY"] = values.NewString("Linux")
	stdlib.Constants["PHP_SAPI"] = values.NewString("cli")
	stdlib.Constants["PHP_EOL"] = values.NewString("\n")
	stdlib.Constants["PHP_INT_MAX"] = values.NewInt(9223372036854775807)
	stdlib.Constants["PHP_INT_MIN"] = values.NewInt(-9223372036854775808)
	stdlib.Constants["PHP_INT_SIZE"] = values.NewInt(8)
	stdlib.Constants["PHP_FLOAT_MAX"] = values.NewFloat(1.7976931348623157e+308)
	stdlib.Constants["PHP_FLOAT_MIN"] = values.NewFloat(2.2250738585072014e-308)
	stdlib.Constants["PHP_FLOAT_DIG"] = values.NewInt(15)

	// Directory constants
	stdlib.Constants["DIRECTORY_SEPARATOR"] = values.NewString("/")
	stdlib.Constants["PATH_SEPARATOR"] = values.NewString(":")

	// File constants
	stdlib.Constants["SEEK_SET"] = values.NewInt(0)
	stdlib.Constants["SEEK_CUR"] = values.NewInt(1)
	stdlib.Constants["SEEK_END"] = values.NewInt(2)

	// Lock constants
	stdlib.Constants["LOCK_SH"] = values.NewInt(1)
	stdlib.Constants["LOCK_EX"] = values.NewInt(2)
	stdlib.Constants["LOCK_UN"] = values.NewInt(3)
	stdlib.Constants["LOCK_NB"] = values.NewInt(4)

	// Math constants
	stdlib.Constants["M_PI"] = values.NewFloat(3.14159265358979323846)
	stdlib.Constants["M_E"] = values.NewFloat(2.71828182845904523536)
	stdlib.Constants["M_LOG2E"] = values.NewFloat(1.44269504088896340736)
	stdlib.Constants["M_LOG10E"] = values.NewFloat(0.434294481903251827651)
	stdlib.Constants["M_LN2"] = values.NewFloat(0.693147180559945309417)
	stdlib.Constants["M_LN10"] = values.NewFloat(2.30258509299404568402)
	stdlib.Constants["M_PI_2"] = values.NewFloat(1.57079632679489661923)
	stdlib.Constants["M_PI_4"] = values.NewFloat(0.785398163397448309616)
	stdlib.Constants["M_1_PI"] = values.NewFloat(0.318309886183790671538)
	stdlib.Constants["M_2_PI"] = values.NewFloat(0.636619772367581343076)
	stdlib.Constants["M_SQRTPI"] = values.NewFloat(1.77245385090551602730)
	stdlib.Constants["M_2_SQRTPI"] = values.NewFloat(1.12837916709551257390)
	stdlib.Constants["M_SQRT2"] = values.NewFloat(1.41421356237309504880)
	stdlib.Constants["M_SQRT3"] = values.NewFloat(1.73205080756887729353)
	stdlib.Constants["M_SQRT1_2"] = values.NewFloat(0.707106781186547524401)
	stdlib.Constants["M_LNPI"] = values.NewFloat(1.14472988584940017414)
	stdlib.Constants["M_EULER"] = values.NewFloat(0.577215664901532860607)

	// Boolean constants
	stdlib.Constants["TRUE"] = values.NewBool(true)
	stdlib.Constants["FALSE"] = values.NewBool(false)
	stdlib.Constants["NULL"] = values.NewNull()

	// Error constants
	stdlib.Constants["E_ERROR"] = values.NewInt(1)
	stdlib.Constants["E_WARNING"] = values.NewInt(2)
	stdlib.Constants["E_PARSE"] = values.NewInt(4)
	stdlib.Constants["E_NOTICE"] = values.NewInt(8)
	stdlib.Constants["E_CORE_ERROR"] = values.NewInt(16)
	stdlib.Constants["E_CORE_WARNING"] = values.NewInt(32)
	stdlib.Constants["E_COMPILE_ERROR"] = values.NewInt(64)
	stdlib.Constants["E_COMPILE_WARNING"] = values.NewInt(128)
	stdlib.Constants["E_USER_ERROR"] = values.NewInt(256)
	stdlib.Constants["E_USER_WARNING"] = values.NewInt(512)
	stdlib.Constants["E_USER_NOTICE"] = values.NewInt(1024)
	stdlib.Constants["E_STRICT"] = values.NewInt(2048)
	stdlib.Constants["E_RECOVERABLE_ERROR"] = values.NewInt(4096)
	stdlib.Constants["E_DEPRECATED"] = values.NewInt(8192)
	stdlib.Constants["E_USER_DEPRECATED"] = values.NewInt(16384)
	stdlib.Constants["E_ALL"] = values.NewInt(32767)

	// Connection constants
	stdlib.Constants["CONNECTION_ABORTED"] = values.NewInt(1)
	stdlib.Constants["CONNECTION_NORMAL"] = values.NewInt(0)
	stdlib.Constants["CONNECTION_TIMEOUT"] = values.NewInt(2)

	// Stream constants
	stdlib.Constants["STREAM_FILTER_READ"] = values.NewInt(1)
	stdlib.Constants["STREAM_FILTER_WRITE"] = values.NewInt(2)
	stdlib.Constants["STREAM_FILTER_ALL"] = values.NewInt(3)

	// Case constants
	stdlib.Constants["CASE_LOWER"] = values.NewInt(0)
	stdlib.Constants["CASE_UPPER"] = values.NewInt(1)

	// Count constants
	stdlib.Constants["COUNT_NORMAL"] = values.NewInt(0)
	stdlib.Constants["COUNT_RECURSIVE"] = values.NewInt(1)

	// Sort constants
	stdlib.Constants["SORT_REGULAR"] = values.NewInt(0)
	stdlib.Constants["SORT_NUMERIC"] = values.NewInt(1)
	stdlib.Constants["SORT_STRING"] = values.NewInt(2)
	stdlib.Constants["SORT_LOCALE_STRING"] = values.NewInt(5)
	stdlib.Constants["SORT_NATURAL"] = values.NewInt(6)
	stdlib.Constants["SORT_FLAG_CASE"] = values.NewInt(8)

	// Array constants
	stdlib.Constants["ARRAY_FILTER_USE_BOTH"] = values.NewInt(1)
	stdlib.Constants["ARRAY_FILTER_USE_KEY"] = values.NewInt(2)

	// Extract constants
	stdlib.Constants["EXTR_OVERWRITE"] = values.NewInt(0)
	stdlib.Constants["EXTR_SKIP"] = values.NewInt(1)
	stdlib.Constants["EXTR_PREFIX_SAME"] = values.NewInt(2)
	stdlib.Constants["EXTR_PREFIX_ALL"] = values.NewInt(3)
	stdlib.Constants["EXTR_PREFIX_INVALID"] = values.NewInt(4)
	stdlib.Constants["EXTR_PREFIX_IF_EXISTS"] = values.NewInt(5)
	stdlib.Constants["EXTR_IF_EXISTS"] = values.NewInt(6)

	// String constants
	stdlib.Constants["STR_PAD_LEFT"] = values.NewInt(0)
	stdlib.Constants["STR_PAD_RIGHT"] = values.NewInt(1)
	stdlib.Constants["STR_PAD_BOTH"] = values.NewInt(2)

	// HTML constants
	stdlib.Constants["ENT_COMPAT"] = values.NewInt(2)
	stdlib.Constants["ENT_QUOTES"] = values.NewInt(3)
	stdlib.Constants["ENT_NOQUOTES"] = values.NewInt(0)
	stdlib.Constants["ENT_HTML401"] = values.NewInt(0)
	stdlib.Constants["ENT_XML1"] = values.NewInt(16)
	stdlib.Constants["ENT_XHTML"] = values.NewInt(32)
	stdlib.Constants["ENT_HTML5"] = values.NewInt(48)

	// Password constants
	stdlib.Constants["PASSWORD_DEFAULT"] = values.NewString("2y")
	stdlib.Constants["PASSWORD_BCRYPT"] = values.NewString("2y")
	stdlib.Constants["PASSWORD_ARGON2I"] = values.NewString("argon2i")
	stdlib.Constants["PASSWORD_ARGON2ID"] = values.NewString("argon2id")

	// Filter constants
	stdlib.Constants["FILTER_VALIDATE_INT"] = values.NewInt(257)
	stdlib.Constants["FILTER_VALIDATE_BOOLEAN"] = values.NewInt(258)
	stdlib.Constants["FILTER_VALIDATE_FLOAT"] = values.NewInt(259)
	stdlib.Constants["FILTER_VALIDATE_REGEXP"] = values.NewInt(272)
	stdlib.Constants["FILTER_VALIDATE_URL"] = values.NewInt(273)
	stdlib.Constants["FILTER_VALIDATE_EMAIL"] = values.NewInt(274)
	stdlib.Constants["FILTER_VALIDATE_IP"] = values.NewInt(275)

	// JSON constants
	stdlib.Constants["JSON_ERROR_NONE"] = values.NewInt(0)
	stdlib.Constants["JSON_ERROR_DEPTH"] = values.NewInt(1)
	stdlib.Constants["JSON_ERROR_STATE_MISMATCH"] = values.NewInt(2)
	stdlib.Constants["JSON_ERROR_CTRL_CHAR"] = values.NewInt(3)
	stdlib.Constants["JSON_ERROR_SYNTAX"] = values.NewInt(4)
	stdlib.Constants["JSON_ERROR_UTF8"] = values.NewInt(5)
	stdlib.Constants["JSON_THROW_ON_ERROR"] = values.NewInt(4194304)

	// PCRE constants
	stdlib.Constants["PREG_PATTERN_ORDER"] = values.NewInt(1)
	stdlib.Constants["PREG_SET_ORDER"] = values.NewInt(2)
	stdlib.Constants["PREG_OFFSET_CAPTURE"] = values.NewInt(256)

	// Date constants
	stdlib.Constants["DATE_ATOM"] = values.NewString("Y-m-d\\TH:i:sP")
	stdlib.Constants["DATE_COOKIE"] = values.NewString("l, d-M-Y H:i:s T")
	stdlib.Constants["DATE_ISO8601"] = values.NewString("Y-m-d\\TH:i:sO")
	stdlib.Constants["DATE_RFC822"] = values.NewString("D, d M y H:i:s O")
	stdlib.Constants["DATE_RFC850"] = values.NewString("l, d-M-y H:i:s T")
	stdlib.Constants["DATE_RFC1036"] = values.NewString("D, d M y H:i:s O")
	stdlib.Constants["DATE_RFC1123"] = values.NewString("D, d M Y H:i:s O")
	stdlib.Constants["DATE_RFC7231"] = values.NewString("D, d M Y H:i:s \\G\\M\\T")
	stdlib.Constants["DATE_RFC2822"] = values.NewString("D, d M Y H:i:s O")
	stdlib.Constants["DATE_RFC3339"] = values.NewString("Y-m-d\\TH:i:sP")
	stdlib.Constants["DATE_RFC3339_EXTENDED"] = values.NewString("Y-m-d\\TH:i:s.vP")
	stdlib.Constants["DATE_RSS"] = values.NewString("D, d M Y H:i:s O")
	stdlib.Constants["DATE_W3C"] = values.NewString("Y-m-d\\TH:i:sP")
}
