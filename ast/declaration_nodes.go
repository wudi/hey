package ast

import (
	"fmt"
	"strings"
)

// ============= DECLARATION NODES =============

// DeclarationStatement wraps a declaration as a statement
type DeclarationStatement struct {
	BaseNode
	Declaration interface{} `json:"declaration"`
}

func (d *DeclarationStatement) GetChildren() []Node {
	if node, ok := d.Declaration.(Node); ok {
		return []Node{node}
	}
	return nil
}

func (d *DeclarationStatement) String() string {
	return fmt.Sprintf("%v", d.Declaration)
}

func (d *DeclarationStatement) statementNode() {}
func (d *DeclarationStatement) declarationNode() {}

// ============= FUNCTION DECLARATIONS =============

// Parameter represents a function/method parameter
type Parameter struct {
	BaseNode
	Name         string        `json:"name"`
	Type         Type          `json:"type,omitempty"`
	DefaultValue Expression    `json:"default_value,omitempty"`
	IsReference  bool          `json:"is_reference,omitempty"`
	IsVariadic   bool          `json:"is_variadic,omitempty"`
	Modifiers    []string      `json:"modifiers,omitempty"` // For constructor property promotion
	Attributes   AttributeList `json:"attributes,omitempty"`
}

func (p *Parameter) GetChildren() []Node {
	var children []Node
	if p.Type != nil {
		children = append(children, p.Type)
	}
	if p.DefaultValue != nil {
		children = append(children, p.DefaultValue)
	}
	return children
}

func (p *Parameter) String() string {
	var parts []string
	
	if len(p.Modifiers) > 0 {
		parts = append(parts, strings.Join(p.Modifiers, " "))
	}
	
	if p.Type != nil {
		parts = append(parts, p.Type.String())
	}
	
	if p.IsReference {
		parts = append(parts, "&")
	}
	
	if p.IsVariadic {
		parts = append(parts, "...")
	}
	
	parts = append(parts, p.Name)
	
	result := strings.Join(parts, " ")
	
	if p.DefaultValue != nil {
		result += " = " + p.DefaultValue.String()
	}
	
	return result
}

// UseVariable represents a variable in closure use clause
type UseVariable struct {
	BaseNode
	Name        string `json:"name"`
	IsReference bool   `json:"is_reference,omitempty"`
}

func (u *UseVariable) GetChildren() []Node { return nil }

func (u *UseVariable) String() string {
	if u.IsReference {
		return "&" + u.Name
	}
	return u.Name
}

// ============= CLASS DECLARATIONS =============

// ClassDeclaration represents a class declaration
type ClassDeclaration struct {
	BaseNode
	Name       string        `json:"name"`
	Modifiers  []string      `json:"modifiers,omitempty"` // abstract, final, readonly
	Extends    Expression    `json:"extends,omitempty"`
	Implements []Expression  `json:"implements,omitempty"`
	Members    []ClassMember `json:"members"`
	Attributes AttributeList `json:"attributes,omitempty"`
}

func (c *ClassDeclaration) GetChildren() []Node {
	var children []Node
	if c.Extends != nil {
		children = append(children, c.Extends)
	}
	for _, impl := range c.Implements {
		children = append(children, impl)
	}
	for _, member := range c.Members {
		children = append(children, member)
	}
	return children
}

func (c *ClassDeclaration) String() string {
	var parts []string
	
	if len(c.Modifiers) > 0 {
		parts = append(parts, strings.Join(c.Modifiers, " "))
	}
	
	parts = append(parts, "class", c.Name)
	
	if c.Extends != nil {
		parts = append(parts, "extends", c.Extends.String())
	}
	
	if len(c.Implements) > 0 {
		var impls []string
		for _, impl := range c.Implements {
			impls = append(impls, impl.String())
		}
		parts = append(parts, "implements", strings.Join(impls, ", "))
	}
	
	return strings.Join(parts, " ")
}

func (c *ClassDeclaration) declarationNode() {}
func (c *ClassDeclaration) SetAttributes(attrs AttributeList) { c.Attributes = attrs }

// InterfaceDeclaration represents an interface declaration
type InterfaceDeclaration struct {
	BaseNode
	Name       string        `json:"name"`
	Extends    []Expression  `json:"extends,omitempty"`
	Members    []ClassMember `json:"members"`
	Attributes AttributeList `json:"attributes,omitempty"`
}

func (i *InterfaceDeclaration) GetChildren() []Node {
	var children []Node
	for _, ext := range i.Extends {
		children = append(children, ext)
	}
	for _, member := range i.Members {
		children = append(children, member)
	}
	return children
}

func (i *InterfaceDeclaration) String() string {
	var parts []string
	parts = append(parts, "interface", i.Name)
	
	if len(i.Extends) > 0 {
		var exts []string
		for _, ext := range i.Extends {
			exts = append(exts, ext.String())
		}
		parts = append(parts, "extends", strings.Join(exts, ", "))
	}
	
	return strings.Join(parts, " ")
}

func (i *InterfaceDeclaration) declarationNode() {}
func (i *InterfaceDeclaration) SetAttributes(attrs AttributeList) { i.Attributes = attrs }

// TraitDeclaration represents a trait declaration
type TraitDeclaration struct {
	BaseNode
	Name       string        `json:"name"`
	Members    []ClassMember `json:"members"`
	Attributes AttributeList `json:"attributes,omitempty"`
}

func (t *TraitDeclaration) GetChildren() []Node {
	var children []Node
	for _, member := range t.Members {
		children = append(children, member)
	}
	return children
}

func (t *TraitDeclaration) String() string {
	return fmt.Sprintf("trait %s", t.Name)
}

func (t *TraitDeclaration) declarationNode() {}
func (t *TraitDeclaration) SetAttributes(attrs AttributeList) { t.Attributes = attrs }

// ============= CLASS MEMBERS =============

// MethodDeclaration represents a method declaration
type MethodDeclaration struct {
	BaseNode
	Name             string        `json:"name"`
	Modifiers        []string      `json:"modifiers,omitempty"` // public, private, protected, static, abstract, final
	Parameters       []*Parameter  `json:"parameters"`
	ReturnType       Type          `json:"return_type,omitempty"`
	Body             Statement     `json:"body,omitempty"`
	ReturnsReference bool          `json:"returns_reference,omitempty"`
	Attributes       AttributeList `json:"attributes,omitempty"`
}

func (m *MethodDeclaration) GetChildren() []Node {
	var children []Node
	for _, param := range m.Parameters {
		children = append(children, param)
	}
	if m.ReturnType != nil {
		children = append(children, m.ReturnType)
	}
	if m.Body != nil {
		children = append(children, m.Body)
	}
	return children
}

func (m *MethodDeclaration) String() string {
	var parts []string
	
	if len(m.Modifiers) > 0 {
		parts = append(parts, strings.Join(m.Modifiers, " "))
	}
	
	parts = append(parts, "function")
	
	if m.ReturnsReference {
		parts = append(parts, "&")
	}
	
	parts = append(parts, m.Name)
	
	var params []string
	for _, p := range m.Parameters {
		params = append(params, p.String())
	}
	parts = append(parts, fmt.Sprintf("(%s)", strings.Join(params, ", ")))
	
	if m.ReturnType != nil {
		parts = append(parts, ":", m.ReturnType.String())
	}
	
	return strings.Join(parts, " ")
}

func (m *MethodDeclaration) classMemberNode() {}
func (m *MethodDeclaration) declarationNode() {} // Can also be a declaration
func (m *MethodDeclaration) SetAttributes(attrs AttributeList) { m.Attributes = attrs }

// PropertyDeclaration represents a property declaration
type PropertyDeclaration struct {
	BaseNode
	Name         string          `json:"name"`
	Type         Type            `json:"type,omitempty"`
	DefaultValue Expression      `json:"default_value,omitempty"`
	Modifiers    []string        `json:"modifiers,omitempty"` // public, private, protected, static, readonly
	Hooks        []*PropertyHook `json:"hooks,omitempty"`      // PHP 8.4+
	Attributes   AttributeList   `json:"attributes,omitempty"`
}

func (p *PropertyDeclaration) GetChildren() []Node {
	var children []Node
	if p.Type != nil {
		children = append(children, p.Type)
	}
	if p.DefaultValue != nil {
		children = append(children, p.DefaultValue)
	}
	for _, hook := range p.Hooks {
		children = append(children, hook)
	}
	return children
}

func (p *PropertyDeclaration) String() string {
	var parts []string
	
	if len(p.Modifiers) > 0 {
		parts = append(parts, strings.Join(p.Modifiers, " "))
	}
	
	if p.Type != nil {
		parts = append(parts, p.Type.String())
	}
	
	parts = append(parts, p.Name)
	
	result := strings.Join(parts, " ")
	
	if p.DefaultValue != nil {
		result += " = " + p.DefaultValue.String()
	}
	
	if len(p.Hooks) > 0 {
		result += " { ... }"
	} else {
		result += ";"
	}
	
	return result
}

func (p *PropertyDeclaration) classMemberNode() {}
func (p *PropertyDeclaration) SetAttributes(attrs AttributeList) { p.Attributes = attrs }

// ClassConstantDeclaration represents class constant declaration
type ClassConstantDeclaration struct {
	BaseNode
	Constants  []*ConstantClause `json:"constants"`
	Type       Type              `json:"type,omitempty"`
	Modifiers  []string          `json:"modifiers,omitempty"` // public, private, protected, final
	Attributes AttributeList     `json:"attributes,omitempty"`
}

func (c *ClassConstantDeclaration) GetChildren() []Node {
	var children []Node
	if c.Type != nil {
		children = append(children, c.Type)
	}
	for _, constant := range c.Constants {
		children = append(children, constant)
	}
	return children
}

func (c *ClassConstantDeclaration) String() string {
	var parts []string
	
	if len(c.Modifiers) > 0 {
		parts = append(parts, strings.Join(c.Modifiers, " "))
	}
	
	parts = append(parts, "const")
	
	if c.Type != nil {
		parts = append(parts, c.Type.String())
	}
	
	var consts []string
	for _, c := range c.Constants {
		consts = append(consts, c.String())
	}
	
	parts = append(parts, strings.Join(consts, ", "))
	
	return strings.Join(parts, " ") + ";"
}

func (c *ClassConstantDeclaration) classMemberNode() {}
func (c *ClassConstantDeclaration) declarationNode() {}
func (c *ClassConstantDeclaration) SetAttributes(attrs AttributeList) { c.Attributes = attrs }

// TraitUseClause represents use of traits in a class
type TraitUseClause struct {
	BaseNode
	Traits      []Expression         `json:"traits"`
	Adaptations []TraitAdaptation    `json:"adaptations,omitempty"`
}

func (t *TraitUseClause) GetChildren() []Node {
	var children []Node
	for _, trait := range t.Traits {
		children = append(children, trait)
	}
	for _, adaptation := range t.Adaptations {
		children = append(children, adaptation)
	}
	return children
}

func (t *TraitUseClause) String() string {
	var traits []string
	for _, trait := range t.Traits {
		traits = append(traits, trait.String())
	}
	return fmt.Sprintf("use %s", strings.Join(traits, ", "))
}

func (t *TraitUseClause) classMemberNode() {}

// TraitAdaptation interface for trait adaptations
type TraitAdaptation interface {
	Node
	traitAdaptationNode()
}

// TraitPrecedence represents trait precedence resolution
type TraitPrecedence struct {
	BaseNode
	Method     Expression   `json:"method"`
	InsteadOf  []Expression `json:"instead_of"`
}

func (t *TraitPrecedence) GetChildren() []Node {
	children := []Node{t.Method}
	for _, instead := range t.InsteadOf {
		children = append(children, instead)
	}
	return children
}

func (t *TraitPrecedence) String() string {
	var instead []string
	for _, i := range t.InsteadOf {
		instead = append(instead, i.String())
	}
	return fmt.Sprintf("%s insteadof %s", t.Method.String(), strings.Join(instead, ", "))
}

func (t *TraitPrecedence) traitAdaptationNode() {}

// TraitAlias represents trait method aliasing
type TraitAlias struct {
	BaseNode
	Method    Expression `json:"method"`
	Alias     string     `json:"alias,omitempty"`
	Modifiers []string   `json:"modifiers,omitempty"`
}

func (t *TraitAlias) GetChildren() []Node {
	return []Node{t.Method}
}

func (t *TraitAlias) String() string {
	var parts []string
	parts = append(parts, t.Method.String(), "as")
	
	if len(t.Modifiers) > 0 {
		parts = append(parts, strings.Join(t.Modifiers, " "))
	}
	
	if t.Alias != "" {
		parts = append(parts, t.Alias)
	}
	
	return strings.Join(parts, " ")
}

func (t *TraitAlias) traitAdaptationNode() {}

// ============= CONSTANT DECLARATIONS =============

// ConstantDeclaration represents a constant declaration
type ConstantDeclaration struct {
	BaseNode
	Constants []*ConstantClause `json:"constants"`
}

func (c *ConstantDeclaration) GetChildren() []Node {
	var children []Node
	for _, constant := range c.Constants {
		children = append(children, constant)
	}
	return children
}

func (c *ConstantDeclaration) String() string {
	var consts []string
	for _, c := range c.Constants {
		consts = append(consts, c.String())
	}
	return fmt.Sprintf("const %s;", strings.Join(consts, ", "))
}

func (c *ConstantDeclaration) declarationNode() {}

// ConstantClause represents a single constant clause
type ConstantClause struct {
	BaseNode
	Name  string     `json:"name"`
	Value Expression `json:"value"`
}

func (c *ConstantClause) GetChildren() []Node {
	return []Node{c.Value}
}

func (c *ConstantClause) String() string {
	return fmt.Sprintf("%s = %s", c.Name, c.Value.String())
}

// ============= ANONYMOUS CLASS =============

// AnonymousClassExpression represents an anonymous class
type AnonymousClassExpression struct {
	BaseNode
	Modifiers       []string      `json:"modifiers,omitempty"`
	ConstructorArgs []Expression  `json:"constructor_args,omitempty"`
	Extends         Expression    `json:"extends,omitempty"`
	Implements      []Expression  `json:"implements,omitempty"`
	Members         []ClassMember `json:"members"`
	Attributes      AttributeList `json:"attributes,omitempty"`
}

func (a *AnonymousClassExpression) GetChildren() []Node {
	var children []Node
	for _, arg := range a.ConstructorArgs {
		children = append(children, arg)
	}
	if a.Extends != nil {
		children = append(children, a.Extends)
	}
	for _, impl := range a.Implements {
		children = append(children, impl)
	}
	for _, member := range a.Members {
		children = append(children, member)
	}
	return children
}

func (a *AnonymousClassExpression) String() string {
	var parts []string
	parts = append(parts, "new")
	
	if len(a.Modifiers) > 0 {
		parts = append(parts, strings.Join(a.Modifiers, " "))
	}
	
	parts = append(parts, "class")
	
	if len(a.ConstructorArgs) > 0 {
		var args []string
		for _, arg := range a.ConstructorArgs {
			args = append(args, arg.String())
		}
		parts = append(parts, fmt.Sprintf("(%s)", strings.Join(args, ", ")))
	}
	
	if a.Extends != nil {
		parts = append(parts, "extends", a.Extends.String())
	}
	
	if len(a.Implements) > 0 {
		var impls []string
		for _, impl := range a.Implements {
			impls = append(impls, impl.String())
		}
		parts = append(parts, "implements", strings.Join(impls, ", "))
	}
	
	parts = append(parts, "{...}")
	
	return strings.Join(parts, " ")
}

func (a *AnonymousClassExpression) expressionNode() {}
func (a *AnonymousClassExpression) SetAttributes(attrs AttributeList) { a.Attributes = attrs }