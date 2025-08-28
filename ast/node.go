package ast

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/wudi/php-parser/lexer"
)

// Node 表示抽象语法树中的节点接口
type Node interface {
	// GetKind 返回节点的AST Kind类型
	GetKind() ASTKind
	// GetPosition 返回节点在源代码中的位置
	GetPosition() lexer.Position
	// GetAttributes 返回节点的属性
	GetAttributes() map[string]interface{}
	// GetLineNo 返回行号
	GetLineNo() uint32
	// GetChildren 返回子节点
	GetChildren() []Node
	// String 返回节点的字符串表示
	String() string
	// ToJSON 转换为 JSON 表示
	ToJSON() ([]byte, error)
	// Accept 接受访问者
	Accept(visitor Visitor)
}

// Statement 表示语句节点
type Statement interface {
	Node
	statementNode()
}

// Expression 表示表达式节点
type Expression interface {
	Node
	expressionNode()
}

// Identifier 表示标识符节点
type Identifier interface {
	Node
	identifierNode()
}

// BaseNode 基础节点，提供公共字段和方法
type BaseNode struct {
	Kind       ASTKind                    `json:"kind"`
	Position   lexer.Position             `json:"position"`
	Attributes map[string]interface{}     `json:"attributes,omitempty"`
	LineNo     uint32                     `json:"lineno"`
}

// GetKind 返回节点的AST Kind类型
func (b *BaseNode) GetKind() ASTKind {
	return b.Kind
}

// GetPosition 返回节点位置
func (b *BaseNode) GetPosition() lexer.Position {
	return b.Position
}

// GetAttributes 返回节点属性
func (b *BaseNode) GetAttributes() map[string]interface{} {
	if b.Attributes == nil {
		b.Attributes = make(map[string]interface{})
	}
	return b.Attributes
}

// GetLineNo 返回行号
func (b *BaseNode) GetLineNo() uint32 {
	return b.LineNo
}

// GetChildren 返回子节点 - 默认实现，具体类型需要重写
func (b *BaseNode) GetChildren() []Node {
	return nil
}

// ToJSON 转换为 JSON - 这个方法在每个具体类型中应该被重写
func (b *BaseNode) ToJSON() ([]byte, error) {
	return json.MarshalIndent(b, "", "  ")
}

// String 返回节点的字符串表示 - 默认实现，具体类型需要重写
func (b *BaseNode) String() string {
	return b.Kind.String()
}

// Accept 接受访问者 - 默认实现，具体类型需要重写
func (b *BaseNode) Accept(visitor Visitor) {
	visitor.Visit(b)
}

// ============= PROGRAM NODE =============

// Program 表示整个 PHP 程序
type Program struct {
	BaseNode
	Statements []Statement `json:"statements"`
}

// GetChildren 返回子节点
func (p *Program) GetChildren() []Node {
	var children []Node
	for _, stmt := range p.Statements {
		children = append(children, stmt)
	}
	return children
}

// Accept 接受访问者
func (p *Program) Accept(visitor Visitor) {
	if visitor.Visit(p) {
		for _, stmt := range p.Statements {
			stmt.Accept(visitor)
		}
	}
}

// String 返回字符串表示
func (p *Program) String() string {
	var stmts []string
	for _, stmt := range p.Statements {
		stmts = append(stmts, stmt.String())
	}
	return strings.Join(stmts, "\n")
}

// ============= BASIC EXPRESSION NODES =============

// IntegerLiteral 表示整数字面量
type IntegerLiteral struct {
	BaseNode
	Value int64 `json:"value"`
}

func (i *IntegerLiteral) GetChildren() []Node { return nil }
func (i *IntegerLiteral) String() string { return fmt.Sprintf("%d", i.Value) }
func (i *IntegerLiteral) expressionNode() {}

// FloatLiteral 表示浮点数字面量
type FloatLiteral struct {
	BaseNode
	Value float64 `json:"value"`
}

func (f *FloatLiteral) GetChildren() []Node { return nil }
func (f *FloatLiteral) String() string { return fmt.Sprintf("%f", f.Value) }
func (f *FloatLiteral) expressionNode() {}

// StringLiteral 表示字符串字面量
type StringLiteral struct {
	BaseNode
	Value string `json:"value"`
}

func (s *StringLiteral) GetChildren() []Node { return nil }
func (s *StringLiteral) String() string { return fmt.Sprintf("'%s'", s.Value) }
func (s *StringLiteral) expressionNode() {}

// BooleanLiteral 表示布尔字面量
type BooleanLiteral struct {
	BaseNode
	Value bool `json:"value"`
}

func (b *BooleanLiteral) GetChildren() []Node { return nil }
func (b *BooleanLiteral) String() string { 
	if b.Value {
		return "true"
	}
	return "false"
}
func (b *BooleanLiteral) expressionNode() {}

// NullLiteral 表示 null 字面量
type NullLiteral struct {
	BaseNode
}

func (n *NullLiteral) GetChildren() []Node { return nil }
func (n *NullLiteral) String() string { return "null" }
func (n *NullLiteral) expressionNode() {}

// Variable 表示变量
type Variable struct {
	BaseNode
	Name string `json:"name"`
}

func (v *Variable) GetChildren() []Node { return nil }
func (v *Variable) String() string { return v.Name }
func (v *Variable) expressionNode() {}

// Identifier 表示标识符
type IdentifierNode struct {
	BaseNode
	Value string `json:"value"`
}

func (i *IdentifierNode) GetChildren() []Node { return nil }
func (i *IdentifierNode) String() string { return i.Value }
func (i *IdentifierNode) expressionNode() {}
func (i *IdentifierNode) identifierNode() {}

// ============= EXPRESSION NODES =============

// BinaryExpression 表示二元表达式
type BinaryExpression struct {
	BaseNode
	Left     Expression `json:"left"`
	Operator string     `json:"operator"`
	Right    Expression `json:"right"`
}

func (b *BinaryExpression) GetChildren() []Node {
	return []Node{b.Left, b.Right}
}

func (b *BinaryExpression) String() string {
	return fmt.Sprintf("(%s %s %s)", b.Left.String(), b.Operator, b.Right.String())
}

func (b *BinaryExpression) expressionNode() {}

// UnaryExpression 表示一元表达式
type UnaryExpression struct {
	BaseNode
	Operator string     `json:"operator"`
	Right    Expression `json:"right"`
}

func (u *UnaryExpression) GetChildren() []Node {
	return []Node{u.Right}
}

func (u *UnaryExpression) String() string {
	return fmt.Sprintf("%s%s", u.Operator, u.Right.String())
}

func (u *UnaryExpression) expressionNode() {}

// PostfixExpression 表示后缀表达式
type PostfixExpression struct {
	BaseNode
	Left     Expression `json:"left"`
	Operator string     `json:"operator"`
}

func (p *PostfixExpression) GetChildren() []Node {
	return []Node{p.Left}
}

func (p *PostfixExpression) String() string {
	return fmt.Sprintf("%s%s", p.Left.String(), p.Operator)
}

func (p *PostfixExpression) expressionNode() {}

// AssignmentExpression 表示赋值表达式
type AssignmentExpression struct {
	BaseNode
	Left        Expression `json:"left"`
	Operator    string     `json:"operator"`
	Right       Expression `json:"right"`
	IsReference bool       `json:"is_reference,omitempty"`
}

func (a *AssignmentExpression) GetChildren() []Node {
	return []Node{a.Left, a.Right}
}

func (a *AssignmentExpression) String() string {
	ref := ""
	if a.IsReference {
		ref = "&"
	}
	return fmt.Sprintf("%s %s %s%s", a.Left.String(), a.Operator, ref, a.Right.String())
}

func (a *AssignmentExpression) expressionNode() {}

// TernaryExpression 表示三元运算符
type TernaryExpression struct {
	BaseNode
	Condition Expression `json:"condition"`
	TrueExp   Expression `json:"true_expression,omitempty"`
	FalseExp  Expression `json:"false_expression"`
}

func (t *TernaryExpression) GetChildren() []Node {
	children := []Node{t.Condition}
	if t.TrueExp != nil {
		children = append(children, t.TrueExp)
	}
	children = append(children, t.FalseExp)
	return children
}

func (t *TernaryExpression) String() string {
	if t.TrueExp != nil {
		return fmt.Sprintf("%s ? %s : %s", t.Condition.String(), t.TrueExp.String(), t.FalseExp.String())
	}
	return fmt.Sprintf("%s ?: %s", t.Condition.String(), t.FalseExp.String())
}

func (t *TernaryExpression) expressionNode() {}

// ArrayExpression 表示数组表达式
type ArrayExpression struct {
	BaseNode
	Elements []*ArrayElement `json:"elements"`
	IsShort  bool            `json:"is_short"`
}

func (a *ArrayExpression) GetChildren() []Node {
	var children []Node
	for _, elem := range a.Elements {
		children = append(children, elem)
	}
	return children
}

func (a *ArrayExpression) String() string {
	var elems []string
	for _, elem := range a.Elements {
		elems = append(elems, elem.String())
	}
	if a.IsShort {
		return fmt.Sprintf("[%s]", strings.Join(elems, ", "))
	}
	return fmt.Sprintf("array(%s)", strings.Join(elems, ", "))
}

func (a *ArrayExpression) expressionNode() {}

// ArrayElement 表示数组元素
type ArrayElement struct {
	BaseNode
	Key         Expression `json:"key,omitempty"`
	Value       Expression `json:"value"`
	IsReference bool       `json:"is_reference,omitempty"`
	IsUnpack    bool       `json:"is_unpack,omitempty"`
}

func (a *ArrayElement) GetChildren() []Node {
	var children []Node
	if a.Key != nil {
		children = append(children, a.Key)
	}
	children = append(children, a.Value)
	return children
}

func (a *ArrayElement) String() string {
	if a.IsUnpack {
		return fmt.Sprintf("...%s", a.Value.String())
	}
	ref := ""
	if a.IsReference {
		ref = "&"
	}
	if a.Key != nil {
		return fmt.Sprintf("%s => %s%s", a.Key.String(), ref, a.Value.String())
	}
	return fmt.Sprintf("%s%s", ref, a.Value.String())
}

// ============= STATEMENT NODES =============

// ExpressionStatement 表示表达式语句
type ExpressionStatement struct {
	BaseNode
	Expression Expression `json:"expression"`
}

func (e *ExpressionStatement) GetChildren() []Node {
	return []Node{e.Expression}
}

func (e *ExpressionStatement) String() string {
	return e.Expression.String() + ";"
}

func (e *ExpressionStatement) statementNode() {}

// BlockStatement 表示块语句
type BlockStatement struct {
	BaseNode
	Statements []Statement `json:"statements"`
}

func (b *BlockStatement) GetChildren() []Node {
	var children []Node
	for _, stmt := range b.Statements {
		children = append(children, stmt)
	}
	return children
}

func (b *BlockStatement) String() string {
	var stmts []string
	for _, stmt := range b.Statements {
		stmts = append(stmts, stmt.String())
	}
	return fmt.Sprintf("{\n%s\n}", strings.Join(stmts, "\n"))
}

func (b *BlockStatement) statementNode() {}

// IfStatement 表示 if 语句
type IfStatement struct {
	BaseNode
	Condition        Expression        `json:"condition"`
	ThenStatement    Statement         `json:"then_statement"`
	ElseIfStatements []*ElseIfStatement `json:"elseif_statements,omitempty"`
	ElseStatement    Statement         `json:"else_statement,omitempty"`
	IsAlternative    bool              `json:"is_alternative,omitempty"`
}

func (i *IfStatement) GetChildren() []Node {
	children := []Node{i.Condition, i.ThenStatement}
	for _, elseIf := range i.ElseIfStatements {
		children = append(children, elseIf)
	}
	if i.ElseStatement != nil {
		children = append(children, i.ElseStatement)
	}
	return children
}

func (i *IfStatement) String() string {
	result := fmt.Sprintf("if (%s) ", i.Condition.String())
	if i.IsAlternative {
		result += ": " + i.ThenStatement.String()
		for _, elseIf := range i.ElseIfStatements {
			result += " " + elseIf.String()
		}
		if i.ElseStatement != nil {
			result += " else: " + i.ElseStatement.String()
		}
		result += " endif;"
	} else {
		result += i.ThenStatement.String()
		for _, elseIf := range i.ElseIfStatements {
			result += " " + elseIf.String()
		}
		if i.ElseStatement != nil {
			result += " else " + i.ElseStatement.String()
		}
	}
	return result
}

func (i *IfStatement) statementNode() {}

// ElseIfStatement 表示 elseif 语句
type ElseIfStatement struct {
	BaseNode
	Condition Expression `json:"condition"`
	Body      Statement  `json:"body"`
}

func (e *ElseIfStatement) GetChildren() []Node {
	return []Node{e.Condition, e.Body}
}

func (e *ElseIfStatement) String() string {
	return fmt.Sprintf("elseif (%s) %s", e.Condition.String(), e.Body.String())
}

// ============= OTHER ESSENTIAL NODES =============

// FunctionDeclaration 表示函数声明
type FunctionDeclaration struct {
	BaseNode
	Name             string       `json:"name"`
	Parameters       []*Parameter `json:"parameters"`
	ReturnType       Type         `json:"return_type,omitempty"`
	Body             Statement    `json:"body,omitempty"`
	ReturnsReference bool         `json:"returns_reference,omitempty"`
	Visibility       string       `json:"visibility,omitempty"`
}

func (f *FunctionDeclaration) GetChildren() []Node {
	var children []Node
	for _, param := range f.Parameters {
		children = append(children, param)
	}
	if f.ReturnType != nil {
		children = append(children, f.ReturnType)
	}
	if f.Body != nil {
		children = append(children, f.Body)
	}
	return children
}

func (f *FunctionDeclaration) String() string {
	var parts []string
	if f.Visibility != "" {
		parts = append(parts, f.Visibility)
	}
	parts = append(parts, "function")
	if f.ReturnsReference {
		parts = append(parts, "&")
	}
	parts = append(parts, f.Name)
	
	var params []string
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}
	parts = append(parts, fmt.Sprintf("(%s)", strings.Join(params, ", ")))
	
	if f.ReturnType != nil {
		parts = append(parts, ":", f.ReturnType.String())
	}
	
	return strings.Join(parts, " ")
}

func (f *FunctionDeclaration) statementNode() {}
func (f *FunctionDeclaration) declarationNode() {}

// Additional essential statement nodes...

// ReturnStatement 表示 return 语句
type ReturnStatement struct {
	BaseNode
	Value Expression `json:"value,omitempty"`
}

func (r *ReturnStatement) GetChildren() []Node {
	if r.Value != nil {
		return []Node{r.Value}
	}
	return nil
}

func (r *ReturnStatement) String() string {
	if r.Value != nil {
		return fmt.Sprintf("return %s;", r.Value.String())
	}
	return "return;"
}

func (r *ReturnStatement) statementNode() {}

// ListElement represents an element in a list() assignment
type ListElement struct {
	BaseNode
	Variable Expression `json:"variable,omitempty"`
}

func (l *ListElement) GetChildren() []Node {
	if l.Variable != nil {
		return []Node{l.Variable}
	}
	return nil
}

func (l *ListElement) String() string {
	if l.Variable != nil {
		return l.Variable.String()
	}
	return ""
}

// CaseStatement represents a case clause in a switch statement
type CaseStatement struct {
	BaseNode
	Value      Expression  `json:"value,omitempty"` // nil for default case
	Statements []Statement `json:"statements"`
	IsDefault  bool        `json:"is_default,omitempty"`
}

func (c *CaseStatement) GetChildren() []Node {
	var children []Node
	if c.Value != nil {
		children = append(children, c.Value)
	}
	for _, stmt := range c.Statements {
		children = append(children, stmt)
	}
	return children
}

func (c *CaseStatement) String() string {
	if c.IsDefault {
		return "default:"
	}
	return fmt.Sprintf("case %s:", c.Value.String())
}

func (c *CaseStatement) statementNode() {}

// ErrorSuppressionExpression represents @ operator
type ErrorSuppressionExpression struct {
	BaseNode
	Expression Expression `json:"expression"`
}

func (e *ErrorSuppressionExpression) GetChildren() []Node {
	return []Node{e.Expression}
}

func (e *ErrorSuppressionExpression) String() string {
	return fmt.Sprintf("@%s", e.Expression.String())
}

func (e *ErrorSuppressionExpression) expressionNode() {}

// ReferenceExpression represents & operator for references
type ReferenceExpression struct {
	BaseNode
	Expression Expression `json:"expression"`
}

func (r *ReferenceExpression) GetChildren() []Node {
	return []Node{r.Expression}
}

func (r *ReferenceExpression) String() string {
	return fmt.Sprintf("&%s", r.Expression.String())
}

func (r *ReferenceExpression) expressionNode() {}

// MemberAccessExpression represents object member access (-> and ?->)
type MemberAccessExpression struct {
	BaseNode
	Object   Expression `json:"object"`
	Property Expression `json:"property"`
	IsNullsafe bool      `json:"is_nullsafe,omitempty"`
}

func (m *MemberAccessExpression) GetChildren() []Node {
	return []Node{m.Object, m.Property}
}

func (m *MemberAccessExpression) String() string {
	if m.IsNullsafe {
		return fmt.Sprintf("%s?->%s", m.Object.String(), m.Property.String())
	}
	return fmt.Sprintf("%s->%s", m.Object.String(), m.Property.String())
}

func (m *MemberAccessExpression) expressionNode() {}

// StaticMemberAccessExpression represents static member access (::)
type StaticMemberAccessExpression struct {
	BaseNode
	Class  Expression `json:"class"`
	Member Expression `json:"member"`
}

func (s *StaticMemberAccessExpression) GetChildren() []Node {
	return []Node{s.Class, s.Member}
}

func (s *StaticMemberAccessExpression) String() string {
	return fmt.Sprintf("%s::%s", s.Class.String(), s.Member.String())
}

func (s *StaticMemberAccessExpression) expressionNode() {}

// ArrayAccessExpression represents array access []
type ArrayAccessExpression struct {
	BaseNode
	Array Expression `json:"array"`
	Index Expression `json:"index,omitempty"`
}

func (a *ArrayAccessExpression) GetChildren() []Node {
	if a.Index != nil {
		return []Node{a.Array, a.Index}
	}
	return []Node{a.Array}
}

func (a *ArrayAccessExpression) String() string {
	if a.Index != nil {
		return fmt.Sprintf("%s[%s]", a.Array.String(), a.Index.String())
	}
	return fmt.Sprintf("%s[]", a.Array.String())
}

func (a *ArrayAccessExpression) expressionNode() {}

// AnonymousFunctionExpression represents anonymous function (closure)
type AnonymousFunctionExpression struct {
	BaseNode
	Parameters       []*Parameter  `json:"parameters"`
	UseVariables     []*UseVariable `json:"use_variables,omitempty"`
	ReturnType       Type          `json:"return_type,omitempty"`
	Body             Statement     `json:"body"`
	ReturnsReference bool          `json:"returns_reference,omitempty"`
	IsStatic         bool          `json:"is_static,omitempty"`
}

func (a *AnonymousFunctionExpression) GetChildren() []Node {
	var children []Node
	for _, param := range a.Parameters {
		children = append(children, param)
	}
	for _, useVar := range a.UseVariables {
		children = append(children, useVar)
	}
	if a.ReturnType != nil {
		children = append(children, a.ReturnType)
	}
	children = append(children, a.Body)
	return children
}

func (a *AnonymousFunctionExpression) String() string {
	var parts []string
	if a.IsStatic {
		parts = append(parts, "static")
	}
	parts = append(parts, "function")
	if a.ReturnsReference {
		parts = append(parts, "&")
	}
	
	var params []string
	for _, p := range a.Parameters {
		params = append(params, p.String())
	}
	parts = append(parts, fmt.Sprintf("(%s)", strings.Join(params, ", ")))
	
	if len(a.UseVariables) > 0 {
		var useVars []string
		for _, u := range a.UseVariables {
			useVars = append(useVars, u.String())
		}
		parts = append(parts, fmt.Sprintf("use (%s)", strings.Join(useVars, ", ")))
	}
	
	if a.ReturnType != nil {
		parts = append(parts, ":", a.ReturnType.String())
	}
	
	return strings.Join(parts, " ")
}

func (a *AnonymousFunctionExpression) expressionNode() {}

// CastExpression represents type casting
type CastExpression struct {
	BaseNode
	Type       string     `json:"type"`
	Expression Expression `json:"expression"`
}

func (c *CastExpression) GetChildren() []Node {
	return []Node{c.Expression}
}

func (c *CastExpression) String() string {
	return fmt.Sprintf("(%s) %s", c.Type, c.Expression.String())
}

func (c *CastExpression) expressionNode() {}

// NewExpression represents object instantiation
type NewExpression struct {
	BaseNode
	Class     Expression   `json:"class"`
	Arguments []Expression `json:"arguments,omitempty"`
}

func (n *NewExpression) GetChildren() []Node {
	children := []Node{n.Class}
	for _, arg := range n.Arguments {
		children = append(children, arg)
	}
	return children
}

func (n *NewExpression) String() string {
	if len(n.Arguments) > 0 {
		var args []string
		for _, arg := range n.Arguments {
			args = append(args, arg.String())
		}
		return fmt.Sprintf("new %s(%s)", n.Class.String(), strings.Join(args, ", "))
	}
	return fmt.Sprintf("new %s", n.Class.String())
}

func (n *NewExpression) expressionNode() {}

// CloneExpression represents object cloning
type CloneExpression struct {
	BaseNode
	Expression Expression `json:"expression"`
}

func (c *CloneExpression) GetChildren() []Node {
	return []Node{c.Expression}
}

func (c *CloneExpression) String() string {
	return fmt.Sprintf("clone %s", c.Expression.String())
}

func (c *CloneExpression) expressionNode() {}

// IssetExpression represents isset() language construct
type IssetExpression struct {
	BaseNode
	Variables []Expression `json:"variables"`
}

func (i *IssetExpression) GetChildren() []Node {
	var children []Node
	for _, v := range i.Variables {
		children = append(children, v)
	}
	return children
}

func (i *IssetExpression) String() string {
	var vars []string
	for _, v := range i.Variables {
		vars = append(vars, v.String())
	}
	return fmt.Sprintf("isset(%s)", strings.Join(vars, ", "))
}

func (i *IssetExpression) expressionNode() {}

// EmptyExpression represents empty() language construct
type EmptyExpression struct {
	BaseNode
	Variable Expression `json:"variable"`
}

func (e *EmptyExpression) GetChildren() []Node {
	return []Node{e.Variable}
}

func (e *EmptyExpression) String() string {
	return fmt.Sprintf("empty(%s)", e.Variable.String())
}

func (e *EmptyExpression) expressionNode() {}

// ListExpression represents list() language construct
type ListExpression struct {
	BaseNode
	Elements []*ListElement `json:"elements"`
}

func (l *ListExpression) GetChildren() []Node {
	var children []Node
	for _, elem := range l.Elements {
		children = append(children, elem)
	}
	return children
}

func (l *ListExpression) String() string {
	var elems []string
	for _, elem := range l.Elements {
		elems = append(elems, elem.String())
	}
	return fmt.Sprintf("list(%s)", strings.Join(elems, ", "))
}

func (l *ListExpression) expressionNode() {}

// ExitExpression represents exit() or die() language construct
type ExitExpression struct {
	BaseNode
	Expression Expression `json:"expression,omitempty"`
}

func (e *ExitExpression) GetChildren() []Node {
	if e.Expression != nil {
		return []Node{e.Expression}
	}
	return nil
}

func (e *ExitExpression) String() string {
	if e.Expression != nil {
		return fmt.Sprintf("exit(%s)", e.Expression.String())
	}
	return "exit()"
}

func (e *ExitExpression) expressionNode() {}

// EvalExpression represents eval() language construct
type EvalExpression struct {
	BaseNode
	Code Expression `json:"code"`
}

func (e *EvalExpression) GetChildren() []Node {
	return []Node{e.Code}
}

func (e *EvalExpression) String() string {
	return fmt.Sprintf("eval(%s)", e.Code.String())
}

func (e *EvalExpression) expressionNode() {}

// PrintExpression represents print language construct
type PrintExpression struct {
	BaseNode
	Expression Expression `json:"expression"`
}

func (p *PrintExpression) GetChildren() []Node {
	return []Node{p.Expression}
}

func (p *PrintExpression) String() string {
	return fmt.Sprintf("print %s", p.Expression.String())
}

func (p *PrintExpression) expressionNode() {}

// IncludeExpression represents include/require constructs
type IncludeExpression struct {
	BaseNode
	Type       string     `json:"type"` // "include", "include_once", "require", "require_once"
	Expression Expression `json:"expression"`
}

func (i *IncludeExpression) GetChildren() []Node {
	return []Node{i.Expression}
}

func (i *IncludeExpression) String() string {
	return fmt.Sprintf("%s %s", i.Type, i.Expression.String())
}

func (i *IncludeExpression) expressionNode() {}

// YieldExpression represents yield expression
type YieldExpression struct {
	BaseNode
	Key   Expression `json:"key,omitempty"`
	Value Expression `json:"value,omitempty"`
}

func (y *YieldExpression) GetChildren() []Node {
	var children []Node
	if y.Key != nil {
		children = append(children, y.Key)
	}
	if y.Value != nil {
		children = append(children, y.Value)
	}
	return children
}

func (y *YieldExpression) String() string {
	if y.Key != nil && y.Value != nil {
		return fmt.Sprintf("yield %s => %s", y.Key.String(), y.Value.String())
	}
	if y.Value != nil {
		return fmt.Sprintf("yield %s", y.Value.String())
	}
	return "yield"
}

func (y *YieldExpression) expressionNode() {}

// YieldFromExpression represents yield from expression
type YieldFromExpression struct {
	BaseNode
	Expression Expression `json:"expression"`
}

func (y *YieldFromExpression) GetChildren() []Node {
	return []Node{y.Expression}
}

func (y *YieldFromExpression) String() string {
	return fmt.Sprintf("yield from %s", y.Expression.String())
}

func (y *YieldFromExpression) expressionNode() {}

// InterpolatedStringExpression represents string interpolation
type InterpolatedStringExpression struct {
	BaseNode
	Parts []Expression `json:"parts"`
}

func (i *InterpolatedStringExpression) GetChildren() []Node {
	var children []Node
	for _, part := range i.Parts {
		children = append(children, part)
	}
	return children
}

func (i *InterpolatedStringExpression) String() string {
	var parts []string
	for _, part := range i.Parts {
		parts = append(parts, part.String())
	}
	return fmt.Sprintf("\"%s\"", strings.Join(parts, ""))
}

func (i *InterpolatedStringExpression) expressionNode() {}

// HeredocExpression represents heredoc strings
type HeredocExpression struct {
	BaseNode
	Label string        `json:"label"`
	Parts []Expression  `json:"parts"`
}

func (h *HeredocExpression) GetChildren() []Node {
	var children []Node
	for _, part := range h.Parts {
		children = append(children, part)
	}
	return children
}

func (h *HeredocExpression) String() string {
	return fmt.Sprintf("<<<EOD\\n...\\nEOD")
}

func (h *HeredocExpression) expressionNode() {}

// NowdocExpression represents nowdoc strings
type NowdocExpression struct {
	BaseNode
	Label   string `json:"label"`
	Content string `json:"content"`
}

func (n *NowdocExpression) GetChildren() []Node {
	return nil
}

func (n *NowdocExpression) String() string {
	return fmt.Sprintf("<<<'%s'\\n%s\\n%s", n.Label, n.Content, n.Label)
}

func (n *NowdocExpression) expressionNode() {}

// ShellExecExpression represents shell execution (backticks)
type ShellExecExpression struct {
	BaseNode
	Command Expression `json:"command"`
}

func (s *ShellExecExpression) GetChildren() []Node {
	return []Node{s.Command}
}

func (s *ShellExecExpression) String() string {
	return fmt.Sprintf("`%s`", s.Command.String())
}

func (s *ShellExecExpression) expressionNode() {}

// MagicConstantExpression represents magic constants (__FILE__, __LINE__, etc.)
type MagicConstantExpression struct {
	BaseNode
	Name string `json:"name"`
}

func (m *MagicConstantExpression) GetChildren() []Node {
	return nil
}

func (m *MagicConstantExpression) String() string {
	return m.Name
}

func (m *MagicConstantExpression) expressionNode() {}

// FunctionCallExpression represents function calls
type FunctionCallExpression struct {
	BaseNode
	Function  Expression   `json:"function"`
	Arguments []Expression `json:"arguments"`
}

func (f *FunctionCallExpression) GetChildren() []Node {
	children := []Node{f.Function}
	for _, arg := range f.Arguments {
		children = append(children, arg)
	}
	return children
}

func (f *FunctionCallExpression) String() string {
	var args []string
	for _, arg := range f.Arguments {
		args = append(args, arg.String())
	}
	return fmt.Sprintf("%s(%s)", f.Function.String(), strings.Join(args, ", "))
}

func (f *FunctionCallExpression) expressionNode() {}

// LabelStatement represents a goto label
type LabelStatement struct {
	BaseNode
	Name string `json:"name"`
}

func (l *LabelStatement) GetChildren() []Node {
	return nil
}

func (l *LabelStatement) String() string {
	return fmt.Sprintf("%s:", l.Name)
}

func (l *LabelStatement) statementNode() {}

// WhileStatement while循环语句
type WhileStatement struct {
	BaseNode
	Condition Expression `json:"condition"`
	Body      Statement  `json:"body"`
}

func (s *WhileStatement) GetChildren() []Node {
	children := []Node{}
	if s.Condition != nil {
		children = append(children, s.Condition)
	}
	if s.Body != nil {
		children = append(children, s.Body)
	}
	return children
}

func (s *WhileStatement) String() string {
	return fmt.Sprintf("while (%s) %s", s.Condition.String(), s.Body.String())
}

func (s *WhileStatement) statementNode() {}

// ForStatement for循环语句
type ForStatement struct {
	BaseNode
	Init      []Expression `json:"init"`
	Condition Expression   `json:"condition"`
	Update    []Expression `json:"update"`
	Body      Statement    `json:"body"`
}

func (s *ForStatement) GetChildren() []Node {
	children := []Node{}
	for _, init := range s.Init {
		if init != nil {
			children = append(children, init)
		}
	}
	if s.Condition != nil {
		children = append(children, s.Condition)
	}
	for _, update := range s.Update {
		if update != nil {
			children = append(children, update)
		}
	}
	if s.Body != nil {
		children = append(children, s.Body)
	}
	return children
}

func (s *ForStatement) String() string {
	return "for"
}

func (s *ForStatement) statementNode() {}

// ForeachStatement foreach循环语句
type ForeachStatement struct {
	BaseNode
	Expression Expression `json:"expression"`
	Key        Expression `json:"key"`
	Value      Expression `json:"value"`
	Body       Statement  `json:"body"`
}

func (s *ForeachStatement) GetChildren() []Node {
	children := []Node{}
	if s.Expression != nil {
		children = append(children, s.Expression)
	}
	if s.Key != nil {
		children = append(children, s.Key)
	}
	if s.Value != nil {
		children = append(children, s.Value)
	}
	if s.Body != nil {
		children = append(children, s.Body)
	}
	return children
}

func (s *ForeachStatement) String() string {
	return "foreach"
}

func (s *ForeachStatement) statementNode() {}

// DoWhileStatement do-while循环语句
type DoWhileStatement struct {
	BaseNode
	Body      Statement  `json:"body"`
	Condition Expression `json:"condition"`
}

func (s *DoWhileStatement) GetChildren() []Node {
	children := []Node{}
	if s.Body != nil {
		children = append(children, s.Body)
	}
	if s.Condition != nil {
		children = append(children, s.Condition)
	}
	return children
}

func (s *DoWhileStatement) String() string {
	return "do-while"
}

func (s *DoWhileStatement) statementNode() {}

// SwitchStatement switch语句
type SwitchStatement struct {
	BaseNode
	Expression Expression  `json:"expression"`
	Cases      []Statement `json:"cases"`
}

func (s *SwitchStatement) GetChildren() []Node {
	children := []Node{}
	if s.Expression != nil {
		children = append(children, s.Expression)
	}
	for _, c := range s.Cases {
		if c != nil {
			children = append(children, c)
		}
	}
	return children
}

func (s *SwitchStatement) String() string {
	return "switch"
}

func (s *SwitchStatement) statementNode() {}

// BreakStatement break语句
type BreakStatement struct {
	BaseNode
	Level Expression `json:"level"`
}

func (s *BreakStatement) GetChildren() []Node {
	if s.Level != nil {
		return []Node{s.Level}
	}
	return []Node{}
}

func (s *BreakStatement) String() string {
	if s.Level != nil {
		return fmt.Sprintf("break %s", s.Level.String())
	}
	return "break"
}

func (s *BreakStatement) statementNode() {}

// Constructor functions for AST nodes

// NewVariable creates a new variable node
func NewVariable(pos lexer.Position, name string) *Variable {
	return &Variable{
		BaseNode: BaseNode{Position: pos},
		Name:     name,
	}
}

// NewStringLiteral creates a new string literal node
func NewStringLiteral(pos lexer.Position, value, raw string) *StringLiteral {
	return &StringLiteral{
		BaseNode: BaseNode{Position: pos},
		Value:    value,
	}
}

// NewBinaryExpression creates a new binary expression node
func NewBinaryExpression(pos lexer.Position, left Expression, operator string, right Expression) *BinaryExpression {
	return &BinaryExpression{
		BaseNode: BaseNode{Position: pos},
		Left:     left,
		Operator: operator,
		Right:    right,
	}
}

// NewIntegerLiteral creates a new integer literal node
func NewIntegerLiteral(pos lexer.Position, value int64) *IntegerLiteral {
	return &IntegerLiteral{
		BaseNode: BaseNode{Position: pos},
		Value:    value,
	}
}

// NewFloatLiteral creates a new float literal node
func NewFloatLiteral(pos lexer.Position, value float64) *FloatLiteral {
	return &FloatLiteral{
		BaseNode: BaseNode{Position: pos},
		Value:    value,
	}
}

// NewBooleanLiteral creates a new boolean literal node
func NewBooleanLiteral(pos lexer.Position, value bool) *BooleanLiteral {
	return &BooleanLiteral{
		BaseNode: BaseNode{Position: pos},
		Value:    value,
	}
}

// NewNullLiteral creates a new null literal node
func NewNullLiteral(pos lexer.Position) *NullLiteral {
	return &NullLiteral{
		BaseNode: BaseNode{Position: pos},
	}
}

// NewEchoStatement creates a new echo statement node
func NewEchoStatement(pos lexer.Position, exprs []Expression) *EchoStatement {
	return &EchoStatement{
		BaseNode:    BaseNode{Position: pos},
		Expressions: exprs,
	}
}

// EchoStatement echo语句
type EchoStatement struct {
	BaseNode
	Expressions []Expression `json:"expressions"`
}

func (s *EchoStatement) GetChildren() []Node {
	children := []Node{}
	for _, expr := range s.Expressions {
		if expr != nil {
			children = append(children, expr)
		}
	}
	return children
}

func (s *EchoStatement) String() string {
	return "echo"
}

func (s *EchoStatement) statementNode() {}

// ASTBuilder AST构建器，用于创建符合PHP官方结构的AST节点
type ASTBuilder struct {
	// 可以在这里添加构建器的状态和配置
}

// NewASTBuilder 创建新的AST构建器
func NewASTBuilder() *ASTBuilder {
	return &ASTBuilder{}
}

// CreateVar 创建变量节点
func (b *ASTBuilder) CreateVar(pos lexer.Position, name string) Node {
	return NewVariable(pos, name)
}

// CreateZval 创建字面量节点
func (b *ASTBuilder) CreateZval(pos lexer.Position, value interface{}) Node {
	switch v := value.(type) {
	case string:
		return NewStringLiteral(pos, v, "")
	case int:
		return NewIntegerLiteral(pos, int64(v))
	case int64:
		return NewIntegerLiteral(pos, v)
	case float32:
		return NewFloatLiteral(pos, float64(v))
	case float64:
		return NewFloatLiteral(pos, v)
	case bool:
		return NewBooleanLiteral(pos, v)
	case nil:
		return NewNullLiteral(pos)
	default:
		return NewNullLiteral(pos)
	}
}

// CreateBinaryOp 创建二元操作节点
func (b *ASTBuilder) CreateBinaryOp(pos lexer.Position, left, right Node, operator string) Node {
	leftExpr, ok1 := left.(Expression)
	rightExpr, ok2 := right.(Expression)
	if !ok1 || !ok2 {
		return nil
	}
	return NewBinaryExpression(pos, leftExpr, operator, rightExpr)
}

// CreateArray 创建数组节点
func (b *ASTBuilder) CreateArray(pos lexer.Position, elements []Node) Node {
	var arrayElements []*ArrayElement
	for _, elem := range elements {
		if expr, ok := elem.(Expression); ok {
			arrayElement := &ArrayElement{
				BaseNode: BaseNode{Position: pos, Kind: ASTArrayElem},
				Value:    expr,
			}
			arrayElements = append(arrayElements, arrayElement)
		}
	}
	return &ArrayExpression{
		BaseNode: BaseNode{Position: pos, Kind: ASTArray},
		Elements: arrayElements,
	}
}

// DeclareStatement declare语句
type DeclareStatement struct {
	BaseNode
	Directives []Expression `json:"directives"`
	Body       Statement    `json:"body"`
}

func (s *DeclareStatement) GetChildren() []Node {
	children := []Node{}
	for _, directive := range s.Directives {
		if directive != nil {
			children = append(children, directive)
		}
	}
	if s.Body != nil {
		children = append(children, s.Body)
	}
	return children
}

func (s *DeclareStatement) String() string {
	return "declare"
}

func (s *DeclareStatement) statementNode() {}

// HaltCompilerStatement __halt_compiler语句
type HaltCompilerStatement struct {
	BaseNode
	Data string `json:"data"`
}

func (s *HaltCompilerStatement) GetChildren() []Node {
	return []Node{}
}

func (s *HaltCompilerStatement) String() string {
	return "__halt_compiler()"
}

func (s *HaltCompilerStatement) statementNode() {}

// ContinueStatement continue语句
type ContinueStatement struct {
	BaseNode
	Level Expression `json:"level"`
}

func (s *ContinueStatement) GetChildren() []Node {
	if s.Level != nil {
		return []Node{s.Level}
	}
	return []Node{}
}

func (s *ContinueStatement) String() string {
	if s.Level != nil {
		return fmt.Sprintf("continue %s", s.Level.String())
	}
	return "continue"
}

func (s *ContinueStatement) statementNode() {}

// GotoStatement goto语句
type GotoStatement struct {
	BaseNode
	Label string `json:"label"`
}

func (s *GotoStatement) GetChildren() []Node {
	return []Node{}
}

func (s *GotoStatement) String() string {
	return fmt.Sprintf("goto %s", s.Label)
}

func (s *GotoStatement) statementNode() {}

// ThrowStatement throw语句
type ThrowStatement struct {
	BaseNode
	Expression Expression `json:"expression"`
}

func (s *ThrowStatement) GetChildren() []Node {
	if s.Expression != nil {
		return []Node{s.Expression}
	}
	return []Node{}
}

func (s *ThrowStatement) String() string {
	return "throw"
}

func (s *ThrowStatement) statementNode() {}

// TryStatement try-catch-finally语句
type TryStatement struct {
	BaseNode
	TryBlock      Statement       `json:"try_block"`
	CatchClauses  []*CatchClause  `json:"catch_clauses"`
	FinallyClause *FinallyClause  `json:"finally_clause"`
}

func (s *TryStatement) GetChildren() []Node {
	children := []Node{}
	if s.TryBlock != nil {
		children = append(children, s.TryBlock)
	}
	for _, catch := range s.CatchClauses {
		if catch != nil {
			children = append(children, catch)
		}
	}
	if s.FinallyClause != nil {
		children = append(children, s.FinallyClause)
	}
	return children
}

func (s *TryStatement) String() string {
	return "try"
}

func (s *TryStatement) statementNode() {}

// CatchClause catch子句
type CatchClause struct {
	BaseNode
	ExceptionTypes []Expression `json:"exception_types"`
	Variable       Expression   `json:"variable"`
	Body          Statement    `json:"body"`
}

func (c *CatchClause) GetChildren() []Node {
	children := []Node{}
	for _, t := range c.ExceptionTypes {
		if t != nil {
			children = append(children, t)
		}
	}
	if c.Variable != nil {
		children = append(children, c.Variable)
	}
	if c.Body != nil {
		children = append(children, c.Body)
	}
	return children
}

func (c *CatchClause) String() string {
	return "catch"
}

// FinallyClause finally子句
type FinallyClause struct {
	BaseNode
	Body Statement `json:"body"`
}

func (c *FinallyClause) GetChildren() []Node {
	if c.Body != nil {
		return []Node{c.Body}
	}
	return []Node{}
}

func (c *FinallyClause) String() string {
	return "finally"
}
