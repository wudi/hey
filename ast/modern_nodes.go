package ast

import (
	"fmt"
	"strings"
)

// ============= MODERN PHP AST NODES (PHP 8.0+) =============

// AttributableStatement interface for statements that can have attributes
type AttributableStatement interface {
	Statement
	SetAttributes(AttributeList)
}

// AttributableDeclaration interface for declarations that can have attributes
type AttributableDeclaration interface {
	Declaration
	SetAttributes(AttributeList)
}

// AttributableClassMember interface for class members that can have attributes
type AttributableClassMember interface {
	ClassMember
	SetAttributes(AttributeList)
}

// AttributablePropertyHook interface for property hooks that can have attributes
type AttributablePropertyHook interface {
	Node
	SetAttributes(AttributeList)
}

// Declaration represents a declaration node
type Declaration interface {
	Node
	declarationNode()
}

// Type represents a type node
type Type interface {
	Node
	typeNode()
}

// ClassMember represents a class member node
type ClassMember interface {
	Node
	classMemberNode()
}

// AttributeList represents a list of attributes
type AttributeList interface {
	Node
	attributeListNode()
}

// ============= ATTRIBUTES (PHP 8.0+) =============

// AttributeListExpression represents #[Attribute] syntax
type AttributeListExpression struct {
	BaseNode
	Groups []*AttributeGroup `json:"groups"`
}

func (a *AttributeListExpression) GetChildren() []Node {
	var children []Node
	for _, group := range a.Groups {
		children = append(children, group)
	}
	return children
}

func (a *AttributeListExpression) String() string {
	var parts []string
	for _, group := range a.Groups {
		parts = append(parts, group.String())
	}
	return strings.Join(parts, " ")
}

func (a *AttributeListExpression) attributeListNode() {}

// AttributeGroup represents a group of attributes
type AttributeGroup struct {
	BaseNode
	Attributes []*AttributeExpression `json:"attributes"`
}

func (a *AttributeGroup) GetChildren() []Node {
	var children []Node
	for _, attr := range a.Attributes {
		children = append(children, attr)
	}
	return children
}

func (a *AttributeGroup) String() string {
	var attrs []string
	for _, attr := range a.Attributes {
		attrs = append(attrs, attr.String())
	}
	return fmt.Sprintf("#[%s]", strings.Join(attrs, ", "))
}

// AttributeExpression represents a single attribute
type AttributeExpression struct {
	BaseNode
	Name      Expression   `json:"name"`
	Arguments []Expression `json:"arguments,omitempty"`
}

func (a *AttributeExpression) GetChildren() []Node {
	children := []Node{a.Name}
	for _, arg := range a.Arguments {
		children = append(children, arg)
	}
	return children
}

func (a *AttributeExpression) String() string {
	if len(a.Arguments) > 0 {
		var args []string
		for _, arg := range a.Arguments {
			args = append(args, arg.String())
		}
		return fmt.Sprintf("%s(%s)", a.Name.String(), strings.Join(args, ", "))
	}
	return a.Name.String()
}

func (a *AttributeExpression) expressionNode() {}

// ============= MATCH EXPRESSION (PHP 8.0+) =============

// MatchExpression represents match expression
type MatchExpression struct {
	BaseNode
	Condition Expression   `json:"condition"`
	Arms      []*MatchArm `json:"arms"`
}

func (m *MatchExpression) GetChildren() []Node {
	children := []Node{m.Condition}
	for _, arm := range m.Arms {
		children = append(children, arm)
	}
	return children
}

func (m *MatchExpression) String() string {
	var arms []string
	for _, arm := range m.Arms {
		arms = append(arms, arm.String())
	}
	return fmt.Sprintf("match(%s) { %s }", m.Condition.String(), strings.Join(arms, ", "))
}

func (m *MatchExpression) expressionNode() {}

// MatchArm represents a single match arm
type MatchArm struct {
	BaseNode
	Conditions []Expression `json:"conditions,omitempty"` // nil for default
	Expression Expression   `json:"expression"`
}

func (m *MatchArm) GetChildren() []Node {
	var children []Node
	for _, cond := range m.Conditions {
		children = append(children, cond)
	}
	children = append(children, m.Expression)
	return children
}

func (m *MatchArm) String() string {
	if len(m.Conditions) == 0 {
		return fmt.Sprintf("default => %s", m.Expression.String())
	}
	var conds []string
	for _, cond := range m.Conditions {
		conds = append(conds, cond.String())
	}
	return fmt.Sprintf("%s => %s", strings.Join(conds, ", "), m.Expression.String())
}

// ============= ENUM (PHP 8.1+) =============

// EnumDeclaration represents an enum declaration
type EnumDeclaration struct {
	BaseNode
	Name        string        `json:"name"`
	BackingType Type          `json:"backing_type,omitempty"`
	Implements  []Expression  `json:"implements,omitempty"`
	Members     []ClassMember `json:"members"`
	Attributes  AttributeList `json:"attributes,omitempty"`
}

func (e *EnumDeclaration) GetChildren() []Node {
	var children []Node
	if e.BackingType != nil {
		children = append(children, e.BackingType)
	}
	for _, impl := range e.Implements {
		children = append(children, impl)
	}
	for _, member := range e.Members {
		children = append(children, member)
	}
	return children
}

func (e *EnumDeclaration) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("enum %s", e.Name))
	if e.BackingType != nil {
		parts = append(parts, fmt.Sprintf(": %s", e.BackingType.String()))
	}
	if len(e.Implements) > 0 {
		var impls []string
		for _, impl := range e.Implements {
			impls = append(impls, impl.String())
		}
		parts = append(parts, fmt.Sprintf("implements %s", strings.Join(impls, ", ")))
	}
	return strings.Join(parts, " ")
}

func (e *EnumDeclaration) declarationNode() {}
func (e *EnumDeclaration) SetAttributes(attrs AttributeList) { e.Attributes = attrs }

// EnumCase represents an enum case
type EnumCase struct {
	BaseNode
	Name       string        `json:"name"`
	Value      Expression    `json:"value,omitempty"`
	Attributes AttributeList `json:"attributes,omitempty"`
}

func (e *EnumCase) GetChildren() []Node {
	if e.Value != nil {
		return []Node{e.Value}
	}
	return nil
}

func (e *EnumCase) String() string {
	if e.Value != nil {
		return fmt.Sprintf("case %s = %s;", e.Name, e.Value.String())
	}
	return fmt.Sprintf("case %s;", e.Name)
}

func (e *EnumCase) classMemberNode() {}
func (e *EnumCase) SetAttributes(attrs AttributeList) { e.Attributes = attrs }

// ============= TYPES (PHP 7.0+, Enhanced in 8.0+) =============

// NullableType represents a nullable type
type NullableType struct {
	BaseNode
	Type Type `json:"type"`
}

func (n *NullableType) GetChildren() []Node {
	return []Node{n.Type}
}

func (n *NullableType) String() string {
	return fmt.Sprintf("?%s", n.Type.String())
}

func (n *NullableType) typeNode() {}

// UnionType represents a union type (PHP 8.0+)
type UnionType struct {
	BaseNode
	Types []Type `json:"types"`
}

func (u *UnionType) GetChildren() []Node {
	var children []Node
	for _, t := range u.Types {
		children = append(children, t)
	}
	return children
}

func (u *UnionType) String() string {
	var types []string
	for _, t := range u.Types {
		types = append(types, t.String())
	}
	return strings.Join(types, "|")
}

func (u *UnionType) typeNode() {}

// IntersectionType represents an intersection type (PHP 8.1+)
type IntersectionType struct {
	BaseNode
	Types []Type `json:"types"`
}

func (i *IntersectionType) GetChildren() []Node {
	var children []Node
	for _, t := range i.Types {
		children = append(children, t)
	}
	return children
}

func (i *IntersectionType) String() string {
	var types []string
	for _, t := range i.Types {
		types = append(types, t.String())
	}
	return strings.Join(types, "&")
}

func (i *IntersectionType) typeNode() {}

// NamedType represents a named type
type NamedType struct {
	BaseNode
	Name Expression `json:"name"`
}

func (n *NamedType) GetChildren() []Node {
	return []Node{n.Name}
}

func (n *NamedType) String() string {
	return n.Name.String()
}

func (n *NamedType) typeNode() {}

// ArrayType represents array type hint
type ArrayType struct {
	BaseNode
}

func (a *ArrayType) GetChildren() []Node { return nil }
func (a *ArrayType) String() string { return "array" }
func (a *ArrayType) typeNode() {}

// CallableType represents callable type hint
type CallableType struct {
	BaseNode
}

func (c *CallableType) GetChildren() []Node { return nil }
func (c *CallableType) String() string { return "callable" }
func (c *CallableType) typeNode() {}

// ScalarType represents scalar types (int, float, string, bool)
type ScalarType struct {
	BaseNode
	Type string `json:"type"`
}

func (s *ScalarType) GetChildren() []Node { return nil }
func (s *ScalarType) String() string { return s.Type }
func (s *ScalarType) typeNode() {}

// StaticType represents static return type
type StaticType struct {
	BaseNode
}

func (s *StaticType) GetChildren() []Node { return nil }
func (s *StaticType) String() string { return "static" }
func (s *StaticType) typeNode() {}

// ============= PROPERTY HOOKS (PHP 8.4+) =============

// PropertyHook represents a property hook (get/set)
type PropertyHook struct {
	BaseNode
	Name             string        `json:"name"` // "get" or "set"
	Parameters       []*Parameter  `json:"parameters,omitempty"`
	Body             Statement     `json:"body,omitempty"`
	Expression       Expression    `json:"expression,omitempty"`
	Modifiers        []string      `json:"modifiers,omitempty"`
	ReturnsReference bool          `json:"returns_reference,omitempty"`
	Attributes       AttributeList `json:"attributes,omitempty"`
}

func (p *PropertyHook) GetChildren() []Node {
	var children []Node
	for _, param := range p.Parameters {
		children = append(children, param)
	}
	if p.Body != nil {
		children = append(children, p.Body)
	}
	if p.Expression != nil {
		children = append(children, p.Expression)
	}
	return children
}

func (p *PropertyHook) String() string {
	return fmt.Sprintf("%s hook", p.Name)
}

func (p *PropertyHook) SetAttributes(attrs AttributeList) { p.Attributes = attrs }

// ============= NAMED ARGUMENTS (PHP 8.0+) =============

// NamedArgument represents a named argument in function call
type NamedArgument struct {
	BaseNode
	Name  string     `json:"name"`
	Value Expression `json:"value"`
}

func (n *NamedArgument) GetChildren() []Node {
	return []Node{n.Value}
}

func (n *NamedArgument) String() string {
	return fmt.Sprintf("%s: %s", n.Name, n.Value.String())
}

func (n *NamedArgument) expressionNode() {}

// ============= THROW EXPRESSION (PHP 8.0+) =============

// ThrowExpression represents throw as expression (not statement)
type ThrowExpression struct {
	BaseNode
	Expression Expression `json:"expression"`
}

func (t *ThrowExpression) GetChildren() []Node {
	return []Node{t.Expression}
}

func (t *ThrowExpression) String() string {
	return fmt.Sprintf("throw %s", t.Expression.String())
}

func (t *ThrowExpression) expressionNode() {}

// ============= NULLSAFE OPERATOR (PHP 8.0+) =============

// NullsafeMemberAccessExpression represents ?-> operator
type NullsafeMemberAccessExpression struct {
	BaseNode
	Object   Expression `json:"object"`
	Property Expression `json:"property"`
}

func (n *NullsafeMemberAccessExpression) GetChildren() []Node {
	return []Node{n.Object, n.Property}
}

func (n *NullsafeMemberAccessExpression) String() string {
	return fmt.Sprintf("%s?->%s", n.Object.String(), n.Property.String())
}

func (n *NullsafeMemberAccessExpression) expressionNode() {}

// ============= PIPE OPERATOR (PHP 8.4+) =============

// PipeExpression represents |> operator
type PipeExpression struct {
	BaseNode
	Left  Expression `json:"left"`
	Right Expression `json:"right"`
}

func (p *PipeExpression) GetChildren() []Node {
	return []Node{p.Left, p.Right}
}

func (p *PipeExpression) String() string {
	return fmt.Sprintf("%s |> %s", p.Left.String(), p.Right.String())
}

func (p *PipeExpression) expressionNode() {}

// ============= ARROW FUNCTIONS (PHP 7.4+) =============

// ArrowFunctionExpression represents fn() => expr
type ArrowFunctionExpression struct {
	BaseNode
	Parameters       []*Parameter  `json:"parameters"`
	ReturnType       Type          `json:"return_type,omitempty"`
	Expression       Expression    `json:"expression"`
	ReturnsReference bool          `json:"returns_reference,omitempty"`
	IsStatic         bool          `json:"is_static,omitempty"`
	Attributes       AttributeList `json:"attributes,omitempty"`
}

func (a *ArrowFunctionExpression) GetChildren() []Node {
	var children []Node
	for _, param := range a.Parameters {
		children = append(children, param)
	}
	if a.ReturnType != nil {
		children = append(children, a.ReturnType)
	}
	children = append(children, a.Expression)
	return children
}

func (a *ArrowFunctionExpression) String() string {
	var params []string
	for _, p := range a.Parameters {
		params = append(params, p.String())
	}
	ret := ""
	if a.ReturnType != nil {
		ret = fmt.Sprintf(": %s", a.ReturnType.String())
	}
	return fmt.Sprintf("fn(%s)%s => %s", strings.Join(params, ", "), ret, a.Expression.String())
}

func (a *ArrowFunctionExpression) expressionNode() {}
func (a *ArrowFunctionExpression) SetAttributes(attrs AttributeList) { a.Attributes = attrs }

// ArrowFunctionDeclaration for top-level arrow functions
type ArrowFunctionDeclaration struct {
	BaseNode
	Parameters       []*Parameter  `json:"parameters"`
	ReturnType       Type          `json:"return_type,omitempty"`
	Expression       Expression    `json:"expression"`
	ReturnsReference bool          `json:"returns_reference,omitempty"`
	Attributes       AttributeList `json:"attributes,omitempty"`
}

func (a *ArrowFunctionDeclaration) GetChildren() []Node {
	var children []Node
	for _, param := range a.Parameters {
		children = append(children, param)
	}
	if a.ReturnType != nil {
		children = append(children, a.ReturnType)
	}
	children = append(children, a.Expression)
	return children
}

func (a *ArrowFunctionDeclaration) String() string {
	var params []string
	for _, p := range a.Parameters {
		params = append(params, p.String())
	}
	ret := ""
	if a.ReturnType != nil {
		ret = fmt.Sprintf(": %s", a.ReturnType.String())
	}
	return fmt.Sprintf("fn(%s)%s => %s", strings.Join(params, ", "), ret, a.Expression.String())
}

func (a *ArrowFunctionDeclaration) declarationNode() {}
func (a *ArrowFunctionDeclaration) SetAttributes(attrs AttributeList) { a.Attributes = attrs }

// ============= ADDITIONAL EXPRESSION NODES =============

// UnpackExpression represents ... operator
type UnpackExpression struct {
	BaseNode
	Expression Expression `json:"expression"`
}

func (u *UnpackExpression) GetChildren() []Node {
	return []Node{u.Expression}
}

func (u *UnpackExpression) String() string {
	return fmt.Sprintf("...%s", u.Expression.String())
}

func (u *UnpackExpression) expressionNode() {}

// CoalescingExpression represents ?? operator
type CoalescingExpression struct {
	BaseNode
	Left  Expression `json:"left"`
	Right Expression `json:"right"`
}

func (c *CoalescingExpression) GetChildren() []Node {
	return []Node{c.Left, c.Right}
}

func (c *CoalescingExpression) String() string {
	return fmt.Sprintf("%s ?? %s", c.Left.String(), c.Right.String())
}

func (c *CoalescingExpression) expressionNode() {}

// InstanceofExpression represents instanceof operator
type InstanceofExpression struct {
	BaseNode
	Expression Expression `json:"expression"`
	Class      Expression `json:"class"`
}

func (i *InstanceofExpression) GetChildren() []Node {
	return []Node{i.Expression, i.Class}
}

func (i *InstanceofExpression) String() string {
	return fmt.Sprintf("%s instanceof %s", i.Expression.String(), i.Class.String())
}

func (i *InstanceofExpression) expressionNode() {}

// ============= USE DECLARATIONS =============

// UseDeclaration represents a use statement
type UseDeclaration struct {
	BaseNode
	Type       string       `json:"type,omitempty"` // "function", "const", or ""
	Uses       []*UseClause `json:"uses"`
	Prefix     Expression   `json:"prefix,omitempty"` // For group use
	IsGroupUse bool         `json:"is_group_use,omitempty"`
}

func (u *UseDeclaration) GetChildren() []Node {
	var children []Node
	if u.Prefix != nil {
		children = append(children, u.Prefix)
	}
	for _, use := range u.Uses {
		children = append(children, use)
	}
	return children
}

func (u *UseDeclaration) String() string {
	var uses []string
	for _, use := range u.Uses {
		uses = append(uses, use.String())
	}
	useType := ""
	if u.Type != "" {
		useType = u.Type + " "
	}
	if u.IsGroupUse && u.Prefix != nil {
		return fmt.Sprintf("use %s%s\\{%s};", useType, u.Prefix.String(), strings.Join(uses, ", "))
	}
	return fmt.Sprintf("use %s%s;", useType, strings.Join(uses, ", "))
}

func (u *UseDeclaration) declarationNode() {}

// UseClause represents a single use clause
type UseClause struct {
	BaseNode
	Name  Expression `json:"name"`
	Alias Expression `json:"alias,omitempty"`
	Type  string     `json:"type,omitempty"` // For mixed group use
}

func (u *UseClause) GetChildren() []Node {
	children := []Node{u.Name}
	if u.Alias != nil {
		children = append(children, u.Alias)
	}
	return children
}

func (u *UseClause) String() string {
	if u.Alias != nil {
		return fmt.Sprintf("%s as %s", u.Name.String(), u.Alias.String())
	}
	return u.Name.String()
}

// ============= NAMESPACE =============

// NamespaceDeclaration represents a namespace declaration
type NamespaceDeclaration struct {
	BaseNode
	Name       Expression  `json:"name,omitempty"`
	Statements []Statement `json:"statements,omitempty"`
	IsBlock    bool        `json:"is_block"`
}

func (n *NamespaceDeclaration) GetChildren() []Node {
	var children []Node
	if n.Name != nil {
		children = append(children, n.Name)
	}
	for _, stmt := range n.Statements {
		children = append(children, stmt)
	}
	return children
}

func (n *NamespaceDeclaration) String() string {
	name := ""
	if n.Name != nil {
		name = " " + n.Name.String()
	}
	if n.IsBlock {
		return fmt.Sprintf("namespace%s { ... }", name)
	}
	return fmt.Sprintf("namespace%s;", name)
}

func (n *NamespaceDeclaration) declarationNode() {}

// NamespaceNameExpression represents a namespaced name
type NamespaceNameExpression struct {
	BaseNode
	Parts []string `json:"parts"`
}

func (n *NamespaceNameExpression) GetChildren() []Node { return nil }

func (n *NamespaceNameExpression) String() string {
	return strings.Join(n.Parts, "\\")
}

func (n *NamespaceNameExpression) expressionNode() {}