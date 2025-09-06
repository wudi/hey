package values

import (
	"fmt"
	"strconv"
	"strings"
)

// ValueType represents the type of a PHP value
type ValueType byte

const (
	TypeNull ValueType = iota
	TypeBool
	TypeInt
	TypeFloat
	TypeString
	TypeArray
	TypeObject
	TypeResource
	TypeReference
	TypeCallable
)

// Value represents a PHP runtime value
type Value struct {
	Type ValueType
	Data interface{}
}

// PHP array implementation
type Array struct {
	Elements  map[interface{}]*Value // key -> value
	NextIndex int64                  // for auto-incrementing indices
	IsIndexed bool                   // optimization hint
}

// PHP object implementation
type Object struct {
	ClassName  string
	Properties map[string]*Value
	Methods    map[string]interface{} // function pointers
}

// Reference wrapper for pass-by-reference
type Reference struct {
	Target *Value
}

// Closure represents a PHP closure/anonymous function
type Closure struct {
	Function  interface{}       // Pointer to VM function or compiled function
	BoundVars map[string]*Value // Variables captured via 'use' clause
	Name      string            // Optional name for debugging
}

// Constructors for different value types

func NewNull() *Value {
	return &Value{Type: TypeNull, Data: nil}
}

func NewBool(b bool) *Value {
	return &Value{Type: TypeBool, Data: b}
}

func NewInt(i int64) *Value {
	return &Value{Type: TypeInt, Data: i}
}

func NewFloat(f float64) *Value {
	return &Value{Type: TypeFloat, Data: f}
}

func NewString(s string) *Value {
	return &Value{Type: TypeString, Data: s}
}

func NewArray() *Value {
	return &Value{
		Type: TypeArray,
		Data: &Array{
			Elements:  make(map[interface{}]*Value),
			NextIndex: 0,
			IsIndexed: true,
		},
	}
}

func NewObject(className string) *Value {
	return &Value{
		Type: TypeObject,
		Data: &Object{
			ClassName:  className,
			Properties: make(map[string]*Value),
			Methods:    make(map[string]interface{}),
		},
	}
}

func NewReference(target *Value) *Value {
	return &Value{
		Type: TypeReference,
		Data: &Reference{Target: target},
	}
}

func NewClosure(function interface{}, boundVars map[string]*Value, name string) *Value {
	if boundVars == nil {
		boundVars = make(map[string]*Value)
	}
	return &Value{
		Type: TypeCallable,
		Data: &Closure{
			Function:  function,
			BoundVars: boundVars,
			Name:      name,
		},
	}
}

// Type checking methods

func (v *Value) IsNull() bool {
	return v.Type == TypeNull
}

func (v *Value) IsBool() bool {
	return v.Type == TypeBool
}

func (v *Value) IsInt() bool {
	return v.Type == TypeInt
}

func (v *Value) IsFloat() bool {
	return v.Type == TypeFloat
}

func (v *Value) IsNumeric() bool {
	return v.Type == TypeInt || v.Type == TypeFloat
}

func (v *Value) IsNumericString() bool {
	if v.Type != TypeString {
		return false
	}
	s := strings.TrimSpace(v.Data.(string))
	if s == "" {
		return true // Empty string is considered numeric (converts to 0)
	}
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func (v *Value) IsString() bool {
	return v.Type == TypeString
}

func (v *Value) IsArray() bool {
	return v.Type == TypeArray
}

func (v *Value) IsObject() bool {
	return v.Type == TypeObject
}

func (v *Value) IsReference() bool {
	return v.Type == TypeReference
}

func (v *Value) IsClosure() bool {
	return v.Type == TypeCallable && v.Data != nil
}

// Dereferencing for references
func (v *Value) Deref() *Value {
	if v.Type == TypeReference {
		ref := v.Data.(*Reference)
		return ref.Target.Deref() // Handle chained references
	}
	return v
}

// Type conversion methods following PHP semantics

func (v *Value) ToBool() bool {
	switch v.Type {
	case TypeNull:
		return false
	case TypeBool:
		return v.Data.(bool)
	case TypeInt:
		return v.Data.(int64) != 0
	case TypeFloat:
		f := v.Data.(float64)
		return f != 0.0 && !isNaN(f)
	case TypeString:
		s := v.Data.(string)
		return s != "" && s != "0"
	case TypeArray:
		arr := v.Data.(*Array)
		return len(arr.Elements) > 0
	case TypeObject:
		return true // Objects are always truthy
	case TypeReference:
		return v.Deref().ToBool()
	default:
		return false
	}
}

func (v *Value) ToInt() int64 {
	switch v.Type {
	case TypeNull:
		return 0
	case TypeBool:
		if v.Data.(bool) {
			return 1
		}
		return 0
	case TypeInt:
		return v.Data.(int64)
	case TypeFloat:
		return int64(v.Data.(float64))
	case TypeString:
		s := strings.TrimSpace(v.Data.(string))
		if s == "" {
			return 0
		}
		// PHP-style string to int conversion
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			return i
		}
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return int64(f)
		}
		return 0
	case TypeArray:
		arr := v.Data.(*Array)
		return int64(len(arr.Elements))
	case TypeReference:
		return v.Deref().ToInt()
	default:
		return 0
	}
}

func (v *Value) ToFloat() float64 {
	switch v.Type {
	case TypeNull:
		return 0.0
	case TypeBool:
		if v.Data.(bool) {
			return 1.0
		}
		return 0.0
	case TypeInt:
		return float64(v.Data.(int64))
	case TypeFloat:
		return v.Data.(float64)
	case TypeString:
		s := strings.TrimSpace(v.Data.(string))
		if s == "" {
			return 0.0
		}
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return f
		}
		return 0.0
	case TypeArray:
		arr := v.Data.(*Array)
		return float64(len(arr.Elements))
	case TypeReference:
		return v.Deref().ToFloat()
	default:
		return 0.0
	}
}

func (v *Value) ToString() string {
	switch v.Type {
	case TypeNull:
		return ""
	case TypeBool:
		if v.Data.(bool) {
			return "1"
		}
		return ""
	case TypeInt:
		return strconv.FormatInt(v.Data.(int64), 10)
	case TypeFloat:
		return strconv.FormatFloat(v.Data.(float64), 'g', -1, 64)
	case TypeString:
		return v.Data.(string)
	case TypeArray:
		return "Array"
	case TypeObject:
		obj := v.Data.(*Object)
		return fmt.Sprintf("Object(%s)", obj.ClassName)
	case TypeReference:
		return v.Deref().ToString()
	default:
		return ""
	}
}

// Array operations

func (v *Value) ArrayGet(key *Value) *Value {
	if v.Type != TypeArray {
		return NewNull()
	}

	arr := v.Data.(*Array)
	keyValue := convertArrayKey(key)

	if val, exists := arr.Elements[keyValue]; exists {
		return val
	}
	return NewNull()
}

func (v *Value) ArraySet(key *Value, value *Value) {
	if v.Type != TypeArray {
		return
	}

	arr := v.Data.(*Array)

	if key == nil || key.IsNull() {
		// Auto-increment key
		arr.Elements[arr.NextIndex] = value
		arr.NextIndex++
	} else {
		keyValue := convertArrayKey(key)
		arr.Elements[keyValue] = value

		// Update next index for integer keys
		if key.IsInt() {
			intKey := key.ToInt()
			if intKey >= arr.NextIndex {
				arr.NextIndex = intKey + 1
			}
		}
	}
}

func (v *Value) ArrayUnset(key *Value) {
	if v.Type != TypeArray {
		return
	}

	arr := v.Data.(*Array)
	keyValue := convertArrayKey(key)
	delete(arr.Elements, keyValue)
}

func (v *Value) ArrayCount() int {
	if v.Type != TypeArray {
		return 0
	}
	arr := v.Data.(*Array)
	return len(arr.Elements)
}

// Closure operations

func (v *Value) ClosureGet() *Closure {
	if v.Type != TypeCallable {
		return nil
	}
	return v.Data.(*Closure)
}

// Object operations

func (v *Value) ObjectGet(property string) *Value {
	if v.Type != TypeObject {
		return NewNull()
	}

	obj := v.Data.(*Object)
	if val, exists := obj.Properties[property]; exists {
		return val
	}
	return NewNull()
}

func (v *Value) ObjectSet(property string, value *Value) {
	if v.Type != TypeObject {
		return
	}

	obj := v.Data.(*Object)
	obj.Properties[property] = value
}

func (v *Value) ObjectUnset(property string) {
	if v.Type != TypeObject {
		return
	}

	obj := v.Data.(*Object)
	delete(obj.Properties, property)
}

// PHP comparison operations following official semantics

func (v *Value) Equal(other *Value) bool {
	// Handle references
	v = v.Deref()
	other = other.Deref()

	// Type coercion rules for ==
	if v.Type == other.Type {
		return v.identical(other)
	}

	// NULL comparisons
	if v.IsNull() || other.IsNull() {
		return v.IsNull() && other.IsNull()
	}

	// Boolean comparisons
	if v.IsBool() || other.IsBool() {
		return v.ToBool() == other.ToBool()
	}

	// Numeric comparisons
	if v.IsNumeric() && other.IsNumeric() {
		if v.IsFloat() || other.IsFloat() {
			return v.ToFloat() == other.ToFloat()
		}
		return v.ToInt() == other.ToInt()
	}

	// String/numeric comparisons - only for numeric strings
	if (v.IsNumericString() && other.IsNumeric()) || (v.IsNumeric() && other.IsNumericString()) {
		return v.ToFloat() == other.ToFloat()
	}

	// String comparisons
	if v.IsString() && other.IsString() {
		return v.ToString() == other.ToString()
	}

	// Array comparisons
	if v.IsArray() && other.IsArray() {
		return v.arrayEqual(other)
	}

	return false
}

func (v *Value) Identical(other *Value) bool {
	// Handle references
	v = v.Deref()
	other = other.Deref()

	if v.Type != other.Type {
		return false
	}

	return v.identical(other)
}

func (v *Value) identical(other *Value) bool {
	switch v.Type {
	case TypeNull:
		return true
	case TypeBool:
		return v.Data.(bool) == other.Data.(bool)
	case TypeInt:
		return v.Data.(int64) == other.Data.(int64)
	case TypeFloat:
		return v.Data.(float64) == other.Data.(float64)
	case TypeString:
		return v.Data.(string) == other.Data.(string)
	case TypeArray:
		return v.arrayIdentical(other)
	case TypeObject:
		// Object identity comparison (same instance)
		return v.Data == other.Data
	default:
		return false
	}
}

func (v *Value) Compare(other *Value) int {
	// Returns -1, 0, or 1 for <, ==, > respectively
	v = v.Deref()
	other = other.Deref()

	// NULL comparisons
	if v.IsNull() && other.IsNull() {
		return 0
	}
	if v.IsNull() {
		return -1
	}
	if other.IsNull() {
		return 1
	}

	// Numeric comparisons
	if v.IsNumeric() && other.IsNumeric() {
		if v.IsFloat() || other.IsFloat() {
			vf, of := v.ToFloat(), other.ToFloat()
			if vf < of {
				return -1
			} else if vf > of {
				return 1
			}
			return 0
		}
		vi, oi := v.ToInt(), other.ToInt()
		if vi < oi {
			return -1
		} else if vi > oi {
			return 1
		}
		return 0
	}

	// String comparisons
	vs, os := v.ToString(), other.ToString()
	if vs < os {
		return -1
	} else if vs > os {
		return 1
	}
	return 0
}

// Arithmetic operations

func (v *Value) Add(other *Value) *Value {
	// Handle array union for arrays
	if v.IsArray() && other.IsArray() {
		return v.arrayUnion(other)
	}

	// Numeric addition
	if v.IsNumeric() && other.IsNumeric() {
		if v.IsFloat() || other.IsFloat() {
			return NewFloat(v.ToFloat() + other.ToFloat())
		}
		return NewInt(v.ToInt() + other.ToInt())
	}

	// Convert to numeric
	if v.IsFloat() || other.IsFloat() {
		return NewFloat(v.ToFloat() + other.ToFloat())
	}
	return NewInt(v.ToInt() + other.ToInt())
}

func (v *Value) Subtract(other *Value) *Value {
	if v.IsFloat() || other.IsFloat() {
		return NewFloat(v.ToFloat() - other.ToFloat())
	}
	return NewInt(v.ToInt() - other.ToInt())
}

func (v *Value) Multiply(other *Value) *Value {
	if v.IsFloat() || other.IsFloat() {
		return NewFloat(v.ToFloat() * other.ToFloat())
	}
	return NewInt(v.ToInt() * other.ToInt())
}

func (v *Value) Divide(other *Value) *Value {
	otherFloat := other.ToFloat()
	if otherFloat == 0.0 {
		// Division by zero - return infinity or handle error
		return NewFloat(v.ToFloat() / otherFloat) // Go handles inf/-inf
	}

	vFloat := v.ToFloat()
	result := vFloat / otherFloat

	// Return integer if result is whole number and both operands were integers
	if v.IsInt() && other.IsInt() && result == float64(int64(result)) {
		return NewInt(int64(result))
	}

	return NewFloat(result)
}

func (v *Value) Modulo(other *Value) *Value {
	otherInt := other.ToInt()
	if otherInt == 0 {
		return NewInt(0) // PHP returns 0 for modulo by zero
	}
	return NewInt(v.ToInt() % otherInt)
}

func (v *Value) Power(other *Value) *Value {
	base := v.ToFloat()
	exp := other.ToFloat()
	result := pow(base, exp)

	// Return integer if possible
	if result == float64(int64(result)) && result >= -9223372036854775808 && result <= 9223372036854775807 {
		return NewInt(int64(result))
	}
	return NewFloat(result)
}

// String operations

func (v *Value) Concat(other *Value) *Value {
	return NewString(v.ToString() + other.ToString())
}

// Helper functions

func convertArrayKey(key *Value) interface{} {
	key = key.Deref()
	switch key.Type {
	case TypeNull:
		return ""
	case TypeBool:
		if key.Data.(bool) {
			return int64(1)
		}
		return int64(0)
	case TypeInt:
		return key.Data.(int64)
	case TypeFloat:
		return int64(key.Data.(float64))
	case TypeString:
		s := key.Data.(string)
		// Try to convert to integer if it's a valid integer string
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			return i
		}
		return s
	default:
		return key.ToString()
	}
}

func (v *Value) arrayEqual(other *Value) bool {
	arr1 := v.Data.(*Array)
	arr2 := other.Data.(*Array)

	if len(arr1.Elements) != len(arr2.Elements) {
		return false
	}

	for key, val1 := range arr1.Elements {
		val2, exists := arr2.Elements[key]
		if !exists || !val1.Equal(val2) {
			return false
		}
	}
	return true
}

func (v *Value) arrayIdentical(other *Value) bool {
	arr1 := v.Data.(*Array)
	arr2 := other.Data.(*Array)

	if len(arr1.Elements) != len(arr2.Elements) {
		return false
	}

	for key, val1 := range arr1.Elements {
		val2, exists := arr2.Elements[key]
		if !exists || !val1.Identical(val2) {
			return false
		}
	}
	return true
}

func (v *Value) arrayUnion(other *Value) *Value {
	result := NewArray()
	resultArr := result.Data.(*Array)

	// Copy from left operand first
	arr1 := v.Data.(*Array)
	for key, val := range arr1.Elements {
		resultArr.Elements[key] = val
	}

	// Add from right operand (don't overwrite existing keys)
	arr2 := other.Data.(*Array)
	for key, val := range arr2.Elements {
		if _, exists := resultArr.Elements[key]; !exists {
			resultArr.Elements[key] = val
		}
	}

	return result
}

// Utility functions

func isNaN(f float64) bool {
	return f != f
}

func pow(base, exp float64) float64 {
	// Simple power implementation
	if exp == 0 {
		return 1
	}
	if exp == 1 {
		return base
	}
	if exp < 0 {
		return 1.0 / pow(base, -exp)
	}

	result := 1.0
	for exp > 0 {
		if int64(exp)%2 == 1 {
			result *= base
		}
		base *= base
		exp = exp / 2
	}
	return result
}

// String representation for debugging

func (v *Value) String() string {
	switch v.Type {
	case TypeNull:
		return "null"
	case TypeBool:
		if v.Data.(bool) {
			return "true"
		}
		return "false"
	case TypeInt:
		return fmt.Sprintf("int(%d)", v.Data.(int64))
	case TypeFloat:
		return fmt.Sprintf("float(%g)", v.Data.(float64))
	case TypeString:
		return fmt.Sprintf("string(%q)", v.Data.(string))
	case TypeArray:
		arr := v.Data.(*Array)
		return fmt.Sprintf("array[%d]", len(arr.Elements))
	case TypeObject:
		obj := v.Data.(*Object)
		return fmt.Sprintf("object(%s)", obj.ClassName)
	case TypeReference:
		return fmt.Sprintf("&%s", v.Deref().String())
	default:
		return "unknown"
	}
}

func (vt ValueType) String() string {
	switch vt {
	case TypeNull:
		return "null"
	case TypeBool:
		return "bool"
	case TypeInt:
		return "int"
	case TypeFloat:
		return "float"
	case TypeString:
		return "string"
	case TypeArray:
		return "array"
	case TypeObject:
		return "object"
	case TypeResource:
		return "resource"
	case TypeReference:
		return "reference"
	case TypeCallable:
		return "callable"
	default:
		return "unknown"
	}
}
