package opcodes

import "fmt"

// Opcode represents a bytecode instruction type
type Opcode byte

// Arithmetic Operations (0-19)
const (
	OP_NOP Opcode = iota // No operation

	// Basic arithmetic
	OP_ADD // ADD result, op1, op2
	OP_SUB // SUB result, op1, op2
	OP_MUL // MUL result, op1, op2
	OP_DIV // DIV result, op1, op2
	OP_MOD // MOD result, op1, op2
	OP_POW // POW result, op1, op2 (**)

	// Unary operations
	OP_PLUS  // +value
	OP_MINUS // -value
	OP_NOT   // !value
	OP_BW_NOT // ~value (bitwise not)

	// Increment/Decrement
	OP_PRE_INC  // ++$var
	OP_PRE_DEC  // --$var
	OP_POST_INC // $var++
	OP_POST_DEC // $var--

	// Bitwise operations
	OP_BW_AND // & (bitwise and)
	OP_BW_OR  // | (bitwise or)
	OP_BW_XOR // ^ (bitwise xor)
	OP_SL     // << (shift left)
	OP_SR     // >> (shift right)
)

// Comparison Operations (20-39)
const (
	OP_IS_EQUAL Opcode = iota + 20
	OP_IS_NOT_EQUAL
	OP_IS_IDENTICAL     // ===
	OP_IS_NOT_IDENTICAL // !==
	OP_IS_SMALLER
	OP_IS_SMALLER_OR_EQUAL
	OP_IS_GREATER
	OP_IS_GREATER_OR_EQUAL
	OP_SPACESHIP // <=> (spaceship operator)

	// Type checking
	OP_INSTANCEOF

	// Logical operations
	OP_BOOLEAN_AND // &&
	OP_BOOLEAN_OR  // ||
	OP_LOGICAL_AND // and
	OP_LOGICAL_OR  // or
	OP_LOGICAL_XOR // xor
)

// Control Flow (40-59)
const (
	OP_JMP Opcode = iota + 40
	OP_JMPZ    // Jump if zero (false)
	OP_JMPNZ   // Jump if not zero (true)
	OP_JMPZ_EX // Jump if zero with extended info
	OP_JMPNZ_EX

	// Switch operations
	OP_SWITCH_LONG   // Integer switch case
	OP_SWITCH_STRING // String switch case

	// Exception handling
	OP_THROW
	OP_CATCH
	OP_FINALLY

	// Loop operations
	OP_FE_RESET // foreach reset
	OP_FE_FETCH // foreach fetch
)

// Variable Operations (60-89)
const (
	OP_ASSIGN Opcode = iota + 60
	OP_ASSIGN_REF // =& reference assignment

	// Assignment operators
	OP_ASSIGN_ADD // +=
	OP_ASSIGN_SUB // -=
	OP_ASSIGN_MUL // *=
	OP_ASSIGN_DIV // /=
	OP_ASSIGN_MOD // %=
	OP_ASSIGN_POW // **=
	OP_ASSIGN_CONCAT // .=
	OP_ASSIGN_BW_OR  // |=
	OP_ASSIGN_BW_AND // &=
	OP_ASSIGN_BW_XOR // ^=
	OP_ASSIGN_SL     // <<=
	OP_ASSIGN_SR     // >>=
	OP_ASSIGN_COALESCE // ??=

	// Variable fetching
	OP_FETCH_R      // Read variable
	OP_FETCH_W      // Write variable
	OP_FETCH_RW     // Read-write variable
	OP_FETCH_IS     // isset() check
	OP_FETCH_UNSET  // unset() operation

	// Array operations
	OP_FETCH_DIM_R     // $var[key] read
	OP_FETCH_DIM_W     // $var[key] write  
	OP_FETCH_DIM_RW    // $var[key] read-write
	OP_FETCH_DIM_IS    // isset($var[key])
	OP_FETCH_DIM_UNSET // unset($var[key])

	// Object operations
	OP_FETCH_OBJ_R     // $obj->prop read
	OP_FETCH_OBJ_W     // $obj->prop write
	OP_FETCH_OBJ_RW    // $obj->prop read-write
	OP_FETCH_OBJ_IS    // isset($obj->prop)
	OP_FETCH_OBJ_UNSET // unset($obj->prop)
)

// Function Operations (90-109)
const (
	OP_INIT_FCALL Opcode = iota + 90
	OP_INIT_FCALL_BY_NAME
	OP_INIT_METHOD_CALL
	OP_INIT_STATIC_METHOD_CALL

	// Argument passing
	OP_SEND_VAL    // Send value argument
	OP_SEND_VAR    // Send variable argument  
	OP_SEND_REF    // Send reference argument
	OP_SEND_UNPACK // Send unpacked arguments (...$args)

	// Function calls
	OP_DO_FCALL        // Execute function call
	OP_DO_ICALL        // Execute internal function call
	OP_DO_UCALL        // Execute user function call
	OP_DO_FCALL_BY_NAME // Execute function call by name

	// Returns
	OP_RETURN       // return value
	OP_RETURN_BY_REF // return by reference
	OP_GENERATOR_RETURN // generator return

	// Yield
	OP_YIELD      // yield value
	OP_YIELD_FROM // yield from
)

// Array Operations (110-129)
const (
	OP_INIT_ARRAY Opcode = iota + 110
	OP_ADD_ARRAY_ELEMENT
	OP_ADD_ARRAY_UNPACK

	// Array functions
	OP_COUNT
	OP_IN_ARRAY

	// String operations
	OP_CONCAT
	OP_FAST_CONCAT // Optimized concatenation

	// Type casting
	OP_CAST_BOOL
	OP_CAST_LONG
	OP_CAST_DOUBLE
	OP_CAST_STRING
	OP_CAST_ARRAY
	OP_CAST_OBJECT
)

// Class Operations (130-159)
const (
	OP_NEW Opcode = iota + 130
	OP_CLONE
	OP_INIT_CTOR_CALL
	OP_CALL_CTOR

	// Property operations
	OP_FETCH_CLASS_CONSTANT
	OP_FETCH_STATIC_PROP_R
	OP_FETCH_STATIC_PROP_W
	OP_FETCH_STATIC_PROP_RW
	OP_FETCH_STATIC_PROP_IS
	OP_FETCH_STATIC_PROP_UNSET

	// Method calls
	OP_METHOD_CALL
	OP_STATIC_METHOD_CALL

	// Class checks
	OP_VERIFY_ABSTRACT_CLASS
	OP_VERIFY_RETURN_TYPE
)

// Special Operations (160-179)
const (
	OP_EXIT Opcode = iota + 160
	OP_ECHO
	OP_PRINT
	OP_INCLUDE
	OP_INCLUDE_ONCE  
	OP_REQUIRE
	OP_REQUIRE_ONCE
	OP_EVAL

	// Global operations
	OP_FETCH_GLOBALS
	OP_BIND_GLOBAL
	OP_UNSET_VAR
	OP_ISSET_ISEMPTY_VAR

	// Constants
	OP_FETCH_CONSTANT
	OP_DECLARE_CONST

	// Special values
	OP_QM_ASSIGN    // Ternary operator assignment
	OP_COALESCE     // ?? null coalescing
	OP_MATCH        // match expression
)

// Operand types for instruction encoding
type OpType byte

const (
	IS_UNUSED OpType = iota
	IS_CONST  // Constant
	IS_TMP_VAR // Temporary variable
	IS_VAR     // Variable
	IS_CV      // Compiled variable (cached)
)

// Instruction represents a single bytecode instruction
type Instruction struct {
	Opcode   Opcode // Instruction opcode
	OpType1  byte   // Op1 type (4 bits) + Op2 type (4 bits)
	OpType2  byte   // Result type (4 bits) + Extended info (4 bits)
	Reserved byte   // Reserved for alignment

	Op1    uint32 // First operand
	Op2    uint32 // Second operand  
	Result uint32 // Result location
}

// Helper functions for operand type encoding/decoding
func EncodeOpTypes(op1Type, op2Type, resultType OpType) (byte, byte) {
	opType1 := byte(op1Type)<<4 | byte(op2Type)
	opType2 := byte(resultType) << 4
	return opType1, opType2
}

func DecodeOpType1(encoded byte) OpType {
	return OpType(encoded >> 4)
}

func DecodeOpType2(encoded byte) OpType {
	return OpType(encoded & 0x0F)
}

func DecodeResultType(encoded byte) OpType {
	return OpType(encoded >> 4)
}

// String representations for debugging
var opcodeNames = map[Opcode]string{
	OP_NOP: "NOP",
	
	// Arithmetic
	OP_ADD: "ADD",
	OP_SUB: "SUB", 
	OP_MUL: "MUL",
	OP_DIV: "DIV",
	OP_MOD: "MOD",
	OP_POW: "POW",
	
	// Unary
	OP_PLUS:   "PLUS",
	OP_MINUS:  "MINUS",
	OP_NOT:    "NOT",
	OP_BW_NOT: "BW_NOT",
	
	// Increment/Decrement
	OP_PRE_INC:  "PRE_INC",
	OP_PRE_DEC:  "PRE_DEC", 
	OP_POST_INC: "POST_INC",
	OP_POST_DEC: "POST_DEC",
	
	// Bitwise
	OP_BW_AND: "BW_AND",
	OP_BW_OR:  "BW_OR",
	OP_BW_XOR: "BW_XOR",
	OP_SL:     "SL",
	OP_SR:     "SR",
	
	// Comparison
	OP_IS_EQUAL:            "IS_EQUAL",
	OP_IS_NOT_EQUAL:        "IS_NOT_EQUAL",
	OP_IS_IDENTICAL:        "IS_IDENTICAL",
	OP_IS_NOT_IDENTICAL:    "IS_NOT_IDENTICAL",
	OP_IS_SMALLER:          "IS_SMALLER",
	OP_IS_SMALLER_OR_EQUAL: "IS_SMALLER_OR_EQUAL",
	OP_IS_GREATER:          "IS_GREATER",
	OP_IS_GREATER_OR_EQUAL: "IS_GREATER_OR_EQUAL",
	OP_SPACESHIP:           "SPACESHIP",
	
	OP_INSTANCEOF: "INSTANCEOF",
	
	// Logical
	OP_BOOLEAN_AND: "BOOLEAN_AND",
	OP_BOOLEAN_OR:  "BOOLEAN_OR",
	OP_LOGICAL_AND: "LOGICAL_AND",
	OP_LOGICAL_OR:  "LOGICAL_OR",
	OP_LOGICAL_XOR: "LOGICAL_XOR",
	
	// Control flow
	OP_JMP:         "JMP",
	OP_JMPZ:        "JMPZ",
	OP_JMPNZ:       "JMPNZ",
	OP_JMPZ_EX:     "JMPZ_EX",
	OP_JMPNZ_EX:    "JMPNZ_EX",
	OP_SWITCH_LONG: "SWITCH_LONG",
	OP_SWITCH_STRING: "SWITCH_STRING",
	
	// Exception handling
	OP_THROW:   "THROW",
	OP_CATCH:   "CATCH",
	OP_FINALLY: "FINALLY",
	
	// Variables
	OP_ASSIGN:     "ASSIGN",
	OP_ASSIGN_REF: "ASSIGN_REF",
	
	// Assignment operators
	OP_ASSIGN_ADD:      "ASSIGN_ADD",
	OP_ASSIGN_SUB:      "ASSIGN_SUB", 
	OP_ASSIGN_MUL:      "ASSIGN_MUL",
	OP_ASSIGN_DIV:      "ASSIGN_DIV",
	OP_ASSIGN_MOD:      "ASSIGN_MOD",
	OP_ASSIGN_POW:      "ASSIGN_POW",
	OP_ASSIGN_CONCAT:   "ASSIGN_CONCAT",
	OP_ASSIGN_BW_OR:    "ASSIGN_BW_OR",
	OP_ASSIGN_BW_AND:   "ASSIGN_BW_AND",
	OP_ASSIGN_BW_XOR:   "ASSIGN_BW_XOR",
	OP_ASSIGN_SL:       "ASSIGN_SL",
	OP_ASSIGN_SR:       "ASSIGN_SR",
	OP_ASSIGN_COALESCE: "ASSIGN_COALESCE",
	
	// Variable operations
	OP_FETCH_R:     "FETCH_R",
	OP_FETCH_W:     "FETCH_W",
	OP_FETCH_RW:    "FETCH_RW", 
	OP_FETCH_IS:    "FETCH_IS",
	OP_FETCH_UNSET: "FETCH_UNSET",
	
	// Array operations
	OP_FETCH_DIM_R:     "FETCH_DIM_R",
	OP_FETCH_DIM_W:     "FETCH_DIM_W",
	OP_FETCH_DIM_RW:    "FETCH_DIM_RW",
	OP_FETCH_DIM_IS:    "FETCH_DIM_IS",
	OP_FETCH_DIM_UNSET: "FETCH_DIM_UNSET",
	
	// Object operations  
	OP_FETCH_OBJ_R:     "FETCH_OBJ_R",
	OP_FETCH_OBJ_W:     "FETCH_OBJ_W",
	OP_FETCH_OBJ_RW:    "FETCH_OBJ_RW",
	OP_FETCH_OBJ_IS:    "FETCH_OBJ_IS",
	OP_FETCH_OBJ_UNSET: "FETCH_OBJ_UNSET",
	
	// Functions
	OP_INIT_FCALL:             "INIT_FCALL",
	OP_INIT_FCALL_BY_NAME:     "INIT_FCALL_BY_NAME",
	OP_INIT_METHOD_CALL:       "INIT_METHOD_CALL",
	OP_INIT_STATIC_METHOD_CALL: "INIT_STATIC_METHOD_CALL",
	
	OP_SEND_VAL:    "SEND_VAL",
	OP_SEND_VAR:    "SEND_VAR",
	OP_SEND_REF:    "SEND_REF",
	OP_SEND_UNPACK: "SEND_UNPACK",
	
	OP_DO_FCALL:        "DO_FCALL",
	OP_DO_ICALL:        "DO_ICALL",
	OP_DO_UCALL:        "DO_UCALL",
	OP_DO_FCALL_BY_NAME: "DO_FCALL_BY_NAME",
	
	OP_RETURN:           "RETURN",
	OP_RETURN_BY_REF:    "RETURN_BY_REF",
	OP_GENERATOR_RETURN: "GENERATOR_RETURN",
	
	OP_YIELD:      "YIELD",
	OP_YIELD_FROM: "YIELD_FROM",
	
	// Arrays
	OP_INIT_ARRAY:        "INIT_ARRAY",
	OP_ADD_ARRAY_ELEMENT: "ADD_ARRAY_ELEMENT",
	OP_ADD_ARRAY_UNPACK:  "ADD_ARRAY_UNPACK",
	
	OP_COUNT:    "COUNT",
	OP_IN_ARRAY: "IN_ARRAY",
	
	// Strings
	OP_CONCAT:      "CONCAT",
	OP_FAST_CONCAT: "FAST_CONCAT",
	
	// Casting
	OP_CAST_BOOL:   "CAST_BOOL",
	OP_CAST_LONG:   "CAST_LONG",
	OP_CAST_DOUBLE: "CAST_DOUBLE",
	OP_CAST_STRING: "CAST_STRING",
	OP_CAST_ARRAY:  "CAST_ARRAY",
	OP_CAST_OBJECT: "CAST_OBJECT",
	
	// Classes
	OP_NEW:            "NEW",
	OP_CLONE:          "CLONE", 
	OP_INIT_CTOR_CALL: "INIT_CTOR_CALL",
	OP_CALL_CTOR:      "CALL_CTOR",
	
	OP_FETCH_CLASS_CONSTANT:    "FETCH_CLASS_CONSTANT",
	OP_FETCH_STATIC_PROP_R:     "FETCH_STATIC_PROP_R",
	OP_FETCH_STATIC_PROP_W:     "FETCH_STATIC_PROP_W",
	OP_FETCH_STATIC_PROP_RW:    "FETCH_STATIC_PROP_RW",
	OP_FETCH_STATIC_PROP_IS:    "FETCH_STATIC_PROP_IS",
	OP_FETCH_STATIC_PROP_UNSET: "FETCH_STATIC_PROP_UNSET",
	
	OP_METHOD_CALL:        "METHOD_CALL",
	OP_STATIC_METHOD_CALL: "STATIC_METHOD_CALL",
	
	OP_VERIFY_ABSTRACT_CLASS: "VERIFY_ABSTRACT_CLASS",
	OP_VERIFY_RETURN_TYPE:    "VERIFY_RETURN_TYPE",
	
	// Special
	OP_EXIT:         "EXIT",
	OP_ECHO:         "ECHO",
	OP_PRINT:        "PRINT",
	OP_INCLUDE:      "INCLUDE",
	OP_INCLUDE_ONCE: "INCLUDE_ONCE",
	OP_REQUIRE:      "REQUIRE",
	OP_REQUIRE_ONCE: "REQUIRE_ONCE",
	OP_EVAL:         "EVAL",
	
	OP_FETCH_GLOBALS:      "FETCH_GLOBALS",
	OP_BIND_GLOBAL:        "BIND_GLOBAL",
	OP_UNSET_VAR:          "UNSET_VAR",
	OP_ISSET_ISEMPTY_VAR:  "ISSET_ISEMPTY_VAR",
	
	OP_FETCH_CONSTANT: "FETCH_CONSTANT",
	OP_DECLARE_CONST:  "DECLARE_CONST",
	
	OP_QM_ASSIGN: "QM_ASSIGN",
	OP_COALESCE:  "COALESCE",
	OP_MATCH:     "MATCH",
}

func (op Opcode) String() string {
	if name, exists := opcodeNames[op]; exists {
		return name
	}
	return "UNKNOWN"
}

func (inst *Instruction) String() string {
	op1Type := DecodeOpType1(inst.OpType1)
	op2Type := DecodeOpType2(inst.OpType1)
	resultType := DecodeResultType(inst.OpType2)
	
	return fmt.Sprintf("%s %s:%d, %s:%d, %s:%d",
		inst.Opcode.String(),
		op1Type.String(), inst.Op1,
		op2Type.String(), inst.Op2,
		resultType.String(), inst.Result,
	)
}

func (ot OpType) String() string {
	switch ot {
	case IS_UNUSED:
		return "UNUSED"
	case IS_CONST:
		return "CONST"
	case IS_TMP_VAR:
		return "TMP"
	case IS_VAR:
		return "VAR"
	case IS_CV:
		return "CV"
	default:
		return "UNKNOWN"
	}
}