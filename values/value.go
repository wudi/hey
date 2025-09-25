package values

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
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
	TypeGoroutine
	TypeWaitGroup
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
	Destructed bool                   // flag to prevent multiple destructor calls
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

// Goroutine represents a running goroutine
type Goroutine struct {
	ID       int64             // Unique goroutine identifier
	Function *Closure          // The closure to execute
	UseVars  map[string]*Value // Variables captured with use()
	Status   string            // running, completed, error
	Result   *Value            // Return value when completed
	Error    error             // Error if execution failed
	Done     chan struct{}     // Channel to signal completion
}

// WaitGroup represents a wait group for synchronizing goroutines
type WaitGroup struct {
	counter  int64
	waitChan chan struct{}
	mu       sync.Mutex
	done     bool
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

func NewResource(data interface{}) *Value {
	return &Value{
		Type: TypeResource,
		Data: data,
	}
}

func NewCallable(closure *Closure) *Value {
	return &Value{
		Type: TypeCallable,
		Data: closure,
	}
}

// Global goroutine ID counter
var goroutineIDCounter int64

func NewGoroutine(closure *Closure, useVars map[string]*Value) *Value {
	if useVars == nil {
		useVars = make(map[string]*Value)
	}
	return &Value{
		Type: TypeGoroutine,
		Data: &Goroutine{
			ID:       atomic.AddInt64(&goroutineIDCounter, 1),
			Function: closure,
			UseVars:  useVars,
			Status:   "running",
			Result:   NewNull(),
			Error:    nil,
			Done:     make(chan struct{}),
		},
	}
}

func NewWaitGroup() *Value {
	return &Value{
		Type: TypeWaitGroup,
		Data: &WaitGroup{
			counter:  0,
			waitChan: make(chan struct{}),
			mu:       sync.Mutex{},
			done:     false,
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

func (v *Value) IsResource() bool {
	return v.Type == TypeResource
}

func (v *Value) IsReference() bool {
	return v.Type == TypeReference
}

func (v *Value) IsClosure() bool {
	return v.Type == TypeCallable && v.Data != nil
}

func (v *Value) IsCallable() bool {
	return v.Type == TypeCallable
}

func (v *Value) IsGoroutine() bool {
	return v.Type == TypeGoroutine
}

func (v *Value) IsWaitGroup() bool {
	return v.Type == TypeWaitGroup
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

// phpStringToInt implements PHP's string to integer conversion rules
// It parses the leading numeric part of a string, following PHP's exact behavior:
// - Leading whitespace is ignored
// - Parses optional sign (+/-)
// - Parses digits and decimal point
// - Stops at first non-numeric character
// - Returns integer part of the parsed number
func phpStringToInt(s string) int64 {
	if s == "" {
		return 0
	}

	// Skip leading whitespace
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r') {
		i++
	}

	if i >= len(s) {
		return 0 // Empty after trimming whitespace
	}

	// Parse sign
	sign := int64(1)
	if s[i] == '+' || s[i] == '-' {
		if s[i] == '-' {
			sign = -1
		}
		i++
	}

	if i >= len(s) {
		return 0 // Only sign, no digits
	}

	// Parse digits and decimal point
	var intPart int64
	var fracPart int64
	var fracDivisor int64 = 1
	inFraction := false

	for i < len(s) {
		ch := s[i]
		if ch >= '0' && ch <= '9' {
			digit := int64(ch - '0')
			if inFraction {
				fracPart = fracPart*10 + digit
				fracDivisor *= 10
			} else {
				// Check for overflow
				if intPart > (9223372036854775807-digit)/10 {
					// Would overflow, stop parsing
					break
				}
				intPart = intPart*10 + digit
			}
		} else if ch == '.' && !inFraction {
			inFraction = true
		} else if ch == 'e' || ch == 'E' {
			// For scientific notation like "1.23e2", we parse up to 'e' and ignore the rest
			break
		} else {
			// Any other character stops parsing
			break
		}
		i++
	}

	// Convert to integer (truncate decimal part, don't round)
	result := intPart
	return sign * result
}

// phpStringToFloat implements PHP's string to float conversion rules
// Similar to phpStringToInt but returns float64 and handles scientific notation
func phpStringToFloat(s string) float64 {
	if s == "" {
		return 0.0
	}

	// Skip leading whitespace
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r') {
		i++
	}

	if i >= len(s) {
		return 0.0
	}

	// Parse sign
	sign := 1.0
	if s[i] == '+' || s[i] == '-' {
		if s[i] == '-' {
			sign = -1.0
		}
		i++
	}

	if i >= len(s) {
		return 0.0
	}

	// Find the end of the numeric part (including decimals and scientific notation)
	start := i
	hasDecimal := false
	hasExponent := false

	for i < len(s) {
		ch := s[i]
		if ch >= '0' && ch <= '9' {
			// Digit is always valid
		} else if ch == '.' && !hasDecimal && !hasExponent {
			hasDecimal = true
		} else if (ch == 'e' || ch == 'E') && !hasExponent && i > start {
			hasExponent = true
			// Look ahead to see if there's a sign after 'e'
			if i+1 < len(s) && (s[i+1] == '+' || s[i+1] == '-') {
				i++ // Skip the sign
			}
		} else {
			// Any other character stops parsing
			break
		}
		i++
	}

	// Extract the numeric part
	numericPart := s[start:i]
	if numericPart == "" {
		return 0.0
	}

	// Try to parse as float
	if f, err := strconv.ParseFloat(numericPart, 64); err == nil {
		return sign * f
	}

	return 0.0
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
		return phpStringToInt(v.Data.(string))
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
		return phpStringToFloat(v.Data.(string))
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
	case TypeGoroutine:
		gor := v.Data.(*Goroutine)
		return fmt.Sprintf("Goroutine(#%d, %s)", gor.ID, gor.Status)
	case TypeWaitGroup:
		return "WaitGroup"
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

// VarDump renders the value following PHP's var_dump formatting rules.
func (v *Value) VarDump() string {
	var b strings.Builder
	visited := make(map[*Array]bool)
	v.appendVarDump(&b, 0, visited)
	return b.String()
}

// PrintR renders the value following PHP's print_r formatting rules.
func (v *Value) PrintR() string {
	var b strings.Builder
	visited := make(map[*Array]bool)
	v.appendPrintR(&b, 0, visited, false)
	return b.String()
}

func (v *Value) appendVarDump(b *strings.Builder, indent int, visited map[*Array]bool) {
	ind := strings.Repeat(" ", indent)
	switch v.Type {
	case TypeNull:
		b.WriteString(ind + "NULL\n")
	case TypeBool:
		if v.Data.(bool) {
			b.WriteString(ind + "bool(true)\n")
		} else {
			b.WriteString(ind + "bool(false)\n")
		}
	case TypeInt:
		b.WriteString(fmt.Sprintf("%sint(%d)\n", ind, v.Data.(int64)))
	case TypeFloat:
		b.WriteString(fmt.Sprintf("%sfloat(%s)\n", ind, strconv.FormatFloat(v.Data.(float64), 'g', -1, 64)))
	case TypeString:
		s := v.Data.(string)
		b.WriteString(fmt.Sprintf("%sstring(%d) %q\n", ind, len(s), s))
	case TypeArray:
		v.appendArrayVarDump(b, indent, visited)
	case TypeObject:
		v.appendObjectVarDump(b, indent, visited)
	case TypeReference:
		v.Deref().appendVarDump(b, indent, visited)
	case TypeCallable:
		b.WriteString(ind + "object(Closure)#1 (0) {}\n")
	case TypeResource:
		b.WriteString(ind + "resource(0) of type (unknown)\n")
	default:
		b.WriteString(ind + v.Type.String() + "\n")
	}
}

func (v *Value) appendArrayVarDump(b *strings.Builder, indent int, visited map[*Array]bool) {
	arr := v.Data.(*Array)
	ind := strings.Repeat(" ", indent)
	if visited[arr] {
		b.WriteString(ind + "*RECURSION*\n")
		return
	}
	visited[arr] = true
	defer delete(visited, arr)

	keys := make([]interface{}, 0, len(arr.Elements))
	for k := range arr.Elements {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return compareArrayKeys(keys[i], keys[j])
	})

	b.WriteString(fmt.Sprintf("%sarray(%d) {\n", ind, len(arr.Elements)))
	for _, key := range keys {
		keyStr := formatArrayKey(key)
		b.WriteString(fmt.Sprintf("%s  [%s]=>\n", ind, keyStr))
		val := arr.Elements[key]
		if val == nil {
			b.WriteString(strings.Repeat(" ", indent+2) + "NULL\n")
			continue
		}
		val.appendVarDump(b, indent+2, visited)
	}
	b.WriteString(ind + "}\n")
}

func (v *Value) appendObjectVarDump(b *strings.Builder, indent int, visited map[*Array]bool) {
	obj := v.Data.(*Object)
	ind := strings.Repeat(" ", indent)
	propKeys := make([]string, 0, len(obj.Properties))
	for name := range obj.Properties {
		propKeys = append(propKeys, name)
	}
	sort.Strings(propKeys)
	b.WriteString(fmt.Sprintf("%sobject(%s)#1 (%d) {\n", ind, obj.ClassName, len(obj.Properties)))
	for _, name := range propKeys {
		b.WriteString(fmt.Sprintf("%s  [\"%s\"]=>\n", ind, name))
		val := obj.Properties[name]
		if val == nil {
			b.WriteString(strings.Repeat(" ", indent+2) + "NULL\n")
		} else {
			val.appendVarDump(b, indent+2, visited)
		}
	}
	b.WriteString(ind + "}\n")
}

func (v *Value) appendPrintR(b *strings.Builder, indent int, visited map[*Array]bool, isArrayValue bool) {
	switch v.Type {
	case TypeNull:
		// PHP print_r outputs nothing for NULL
		b.WriteString("")
	case TypeBool:
		if v.Data.(bool) {
			b.WriteString("1")
		} else {
			b.WriteString("")
		}
	case TypeInt:
		b.WriteString(fmt.Sprintf("%d", v.Data.(int64)))
	case TypeFloat:
		// Format float to match PHP's print_r output
		f := v.Data.(float64)
		s := formatFloatForPrintR(f)
		b.WriteString(s)
	case TypeString:
		b.WriteString(v.Data.(string))
	case TypeArray:
		v.appendArrayPrintR(b, indent, visited)
	case TypeObject:
		v.appendObjectPrintR(b, indent, visited)
	case TypeReference:
		v.Deref().appendPrintR(b, indent, visited, isArrayValue)
	case TypeResource:
		// PHP outputs "Resource id #N" for resources
		b.WriteString("Resource id #5")
	case TypeCallable:
		b.WriteString("Closure Object\n(\n)\n")
	default:
		b.WriteString(v.Type.String())
	}
}

func (v *Value) appendArrayPrintR(b *strings.Builder, indent int, visited map[*Array]bool) {
	arr := v.Data.(*Array)

	// Check for recursion first, but don't print the header
	if visited[arr] {
		// For recursion, print special format
		b.WriteString("Array\n")
		b.WriteString(" *RECURSION*")
		return
	}

	b.WriteString("Array\n")

	ind := strings.Repeat(" ", indent*4)
	b.WriteString(ind + "(\n")

	visited[arr] = true
	defer delete(visited, arr)

	// Sort keys to ensure consistent output
	keys := make([]interface{}, 0, len(arr.Elements))
	for k := range arr.Elements {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return compareArrayKeys(keys[i], keys[j])
	})

	nextInd := strings.Repeat(" ", (indent+1)*4)
	for _, key := range keys {
		val := arr.Elements[key]

		// Format the key
		keyStr := formatPrintRKey(key)
		b.WriteString(fmt.Sprintf("%s[%s] => ", nextInd, keyStr))

		if val == nil {
			b.WriteString("\n")
			continue
		}

		// For arrays and objects, we need special formatting
		if val.Type == TypeArray || val.Type == TypeObject {
			val.appendPrintR(b, indent+2, visited, true)
			// Add blank line after nested arrays/objects
			b.WriteString("\n")
		} else {
			val.appendPrintR(b, 0, visited, true)
			b.WriteString("\n")
		}
	}

	b.WriteString(ind + ")\n")
}

func (v *Value) appendObjectPrintR(b *strings.Builder, indent int, visited map[*Array]bool) {
	obj := v.Data.(*Object)

	// Check for object recursion
	// For simplicity, we'll use the array visited map for objects too
	arrKey := &Array{} // Dummy key for the object in the visited map
	if _, exists := visited[arrKey]; exists {
		b.WriteString(fmt.Sprintf("%s Object\n", obj.ClassName))
		b.WriteString(" *RECURSION*")
		return
	}

	b.WriteString(fmt.Sprintf("%s Object\n", obj.ClassName))

	ind := strings.Repeat(" ", indent*4)
	b.WriteString(ind + "(\n")

	// Sort property keys
	propKeys := make([]string, 0, len(obj.Properties))
	for name := range obj.Properties {
		propKeys = append(propKeys, name)
	}
	sort.Strings(propKeys)

	nextInd := strings.Repeat(" ", (indent+1)*4)
	for _, name := range propKeys {
		val := obj.Properties[name]

		// Format property name with visibility modifiers
		// Check if the property has visibility markers
		formattedName := name
		if strings.Contains(name, ":private") || strings.Contains(name, ":protected") {
			// Property already has visibility formatting
			formattedName = name
		} else {
			// Check property visibility (this would need metadata from the class definition)
			// For now, we'll look for naming patterns or assume public
			// PHP's internal representation stores visibility info with the property
			// We could extend the Object struct to store visibility metadata
			formattedName = name
		}

		b.WriteString(fmt.Sprintf("%s[%s] => ", nextInd, formattedName))

		if val == nil {
			b.WriteString("\n")
		} else if val.Type == TypeArray || val.Type == TypeObject {
			val.appendPrintR(b, indent+2, visited, true)
		} else {
			val.appendPrintR(b, 0, visited, true)
			b.WriteString("\n")
		}
	}

	b.WriteString(ind + ")\n")
}

func formatFloatForPrintR(f float64) string {
	// Special case for -0
	if f == 0 && math.Signbit(f) {
		return "-0"
	}

	// Check if the number should be in scientific notation
	absVal := math.Abs(f)

	// PHP uses specific thresholds: < 1e-4 uses scientific, >= 1e10 shows as integer if possible
	if absVal != 0 && absVal < 1e-4 {
		// Use scientific notation like PHP
		s := strconv.FormatFloat(f, 'E', -1, 64)
		// PHP uses uppercase E and formats like 1.0E-5 not 1E-5
		// Ensure at least one decimal place
		if !strings.Contains(s, ".") {
			parts := strings.Split(s, "E")
			if len(parts) == 2 {
				s = parts[0] + ".0E" + parts[1]
			}
		}

		// Remove leading zeros from exponent (E-05 -> E-5, E+05 -> E+5)
		if idx := strings.Index(s, "E"); idx != -1 {
			exp := s[idx+1:]
			sign := ""
			if exp[0] == '+' || exp[0] == '-' {
				sign = string(exp[0])
				exp = exp[1:]
			}
			// Remove leading zeros
			exp = strings.TrimLeft(exp, "0")
			if exp == "" {
				exp = "0"
			}
			s = s[:idx+1] + sign + exp
		}

		return s
	}

	// For large numbers >= 1e10, show as integer if it's a whole number
	if absVal >= 1e10 && f == float64(int64(f)) {
		return fmt.Sprintf("%d", int64(f))
	}

	// Regular float formatting
	s := strconv.FormatFloat(f, 'g', -1, 64)
	return s
}

func formatPrintRKey(key interface{}) string {
	switch k := key.(type) {
	case string:
		return k
	case int:
		return fmt.Sprintf("%d", k)
	case int8:
		return fmt.Sprintf("%d", k)
	case int16:
		return fmt.Sprintf("%d", k)
	case int32:
		return fmt.Sprintf("%d", k)
	case int64:
		return fmt.Sprintf("%d", k)
	case uint:
		return fmt.Sprintf("%d", k)
	case uint8:
		return fmt.Sprintf("%d", k)
	case uint16:
		return fmt.Sprintf("%d", k)
	case uint32:
		return fmt.Sprintf("%d", k)
	case uint64:
		return fmt.Sprintf("%d", k)
	default:
		return fmt.Sprintf("%v", key)
	}
}

func compareArrayKeys(a, b interface{}) bool {
	k1, kind1 := arrayKeySortValue(a)
	k2, kind2 := arrayKeySortValue(b)
	if kind1 != kind2 {
		return kind1 < kind2
	}
	if kind1 == 0 {
		return k1 < k2
	}
	if kind1 == 1 {
		s1 := a.(string)
		s2 := b.(string)
		return s1 < s2
	}
	return fmt.Sprint(a) < fmt.Sprint(b)
}

func arrayKeySortValue(key interface{}) (int64, int) {
	switch k := key.(type) {
	case int:
		return int64(k), 0
	case int8:
		return int64(k), 0
	case int16:
		return int64(k), 0
	case int32:
		return int64(k), 0
	case int64:
		return k, 0
	case uint:
		return int64(k), 0
	case uint8:
		return int64(k), 0
	case uint16:
		return int64(k), 0
	case uint32:
		return int64(k), 0
	case uint64:
		return int64(k), 0
	case string:
		return 0, 1
	default:
		return 0, 2
	}
}

func formatArrayKey(key interface{}) string {
	switch k := key.(type) {
	case string:
		return fmt.Sprintf("\"%s\"", k)
	case int:
		return fmt.Sprintf("%d", k)
	case int8:
		return fmt.Sprintf("%d", k)
	case int16:
		return fmt.Sprintf("%d", k)
	case int32:
		return fmt.Sprintf("%d", k)
	case int64:
		return fmt.Sprintf("%d", k)
	case uint:
		return fmt.Sprintf("%d", k)
	case uint8:
		return fmt.Sprintf("%d", k)
	case uint16:
		return fmt.Sprintf("%d", k)
	case uint32:
		return fmt.Sprintf("%d", k)
	case uint64:
		return fmt.Sprintf("%d", k)
	default:
		return fmt.Sprintf("\"%v\"", k)
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
	case TypeGoroutine:
		return "goroutine"
	case TypeWaitGroup:
		return "waitgroup"
	default:
		return "unknown"
	}
}

func (v *Value) TypeName() string {
	return v.Type.String()
}

// WaitGroup methods
func (v *Value) WaitGroupAdd(delta int64) error {
	if v.Type != TypeWaitGroup {
		return fmt.Errorf("WaitGroup.Add() called on non-WaitGroup value")
	}

	wg := v.Data.(*WaitGroup)
	wg.mu.Lock()
	defer wg.mu.Unlock()

	if wg.done {
		return fmt.Errorf("WaitGroup is already done")
	}

	wg.counter += delta
	if wg.counter < 0 {
		return fmt.Errorf("WaitGroup counter cannot be negative")
	}

	if wg.counter == 0 && wg.waitChan != nil {
		close(wg.waitChan)
		wg.waitChan = nil
		wg.done = true
	}

	return nil
}

func (v *Value) WaitGroupDone() error {
	return v.WaitGroupAdd(-1)
}

func (v *Value) WaitGroupWait() error {
	if v.Type != TypeWaitGroup {
		return fmt.Errorf("WaitGroup.Wait() called on non-WaitGroup value")
	}

	wg := v.Data.(*WaitGroup)
	wg.mu.Lock()

	if wg.done {
		wg.mu.Unlock()
		return nil
	}

	if wg.counter == 0 {
		wg.mu.Unlock()
		return nil
	}

	waitChan := wg.waitChan
	wg.mu.Unlock()

	if waitChan != nil {
		<-waitChan
	}

	return nil
}
