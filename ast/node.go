package ast

import (
	"encoding/json"
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

// Program 表示整个 PHP 程序
type Program struct {
	BaseNode
	Body []Statement `json:"body"`
}

// NewProgram 创建新的程序节点
func NewProgram(pos lexer.Position) *Program {
	return &Program{
		BaseNode: BaseNode{
			Kind:     ASTStmtList,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Body: make([]Statement, 0),
	}
}

// GetChildren 返回子节点
func (p *Program) GetChildren() []Node {
	children := make([]Node, len(p.Body))
	for i, stmt := range p.Body {
		children[i] = stmt
	}
	return children
}

// Accept 接受访问者
func (p *Program) Accept(visitor Visitor) {
	if visitor.Visit(p) {
		for _, stmt := range p.Body {
			stmt.Accept(visitor)
		}
	}
}

func (p *Program) String() string {
	var out strings.Builder
	for _, stmt := range p.Body {
		if stmt != nil {
			out.WriteString(stmt.String())
		}
	}
	return out.String()
}

// Echo 语句
type EchoStatement struct {
	BaseNode
	Arguments []Expression `json:"arguments"`
}

func NewEchoStatement(pos lexer.Position) *EchoStatement {
	return &EchoStatement{
		BaseNode: BaseNode{
			Kind:     ASTEcho,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Arguments: make([]Expression, 0),
	}
}

// GetChildren 返回子节点
func (e *EchoStatement) GetChildren() []Node {
	children := make([]Node, len(e.Arguments))
	for i, arg := range e.Arguments {
		children[i] = arg
	}
	return children
}

// Accept 接受访问者
func (e *EchoStatement) Accept(visitor Visitor) {
	if visitor.Visit(e) {
		for _, arg := range e.Arguments {
			arg.Accept(visitor)
		}
	}
}

func (e *EchoStatement) statementNode() {}

func (e *EchoStatement) String() string {
	var args []string
	for _, arg := range e.Arguments {
		if arg != nil {
			args = append(args, arg.String())
		}
	}
	return "echo " + strings.Join(args, ", ") + ";"
}

// NamespaceStatement 表示命名空间声明语句
type NamespaceStatement struct {
	BaseNode
	Name *NamespaceNameExpression `json:"name,omitempty"` // nil for global namespace
	Body []Statement              `json:"body,omitempty"`  // 可选的命名空间主体
}

func NewNamespaceStatement(pos lexer.Position, name *NamespaceNameExpression) *NamespaceStatement {
	return &NamespaceStatement{
		BaseNode: BaseNode{
			Kind:     ASTNamespace,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Name: name,
		Body: make([]Statement, 0),
	}
}

// GetChildren 返回子节点
func (n *NamespaceStatement) GetChildren() []Node {
	children := make([]Node, 0)
	if n.Name != nil {
		children = append(children, n.Name)
	}
	for _, stmt := range n.Body {
		children = append(children, stmt)
	}
	return children
}

// Accept 接受访问者
func (n *NamespaceStatement) Accept(visitor Visitor) {
	if visitor.Visit(n) {
		if n.Name != nil {
			n.Name.Accept(visitor)
		}
		for _, stmt := range n.Body {
			stmt.Accept(visitor)
		}
	}
}

func (n *NamespaceStatement) statementNode() {}

func (n *NamespaceStatement) String() string {
	if n.Name == nil {
		return "namespace;"
	}
	result := "namespace " + n.Name.String()
	if len(n.Body) > 0 {
		result += " {\n"
		for _, stmt := range n.Body {
			result += "  " + stmt.String() + "\n"
		}
		result += "}"
	} else {
		result += ";"
	}
	return result
}

// UseStatement 表示 use 语句（导入语句）
type UseStatement struct {
	BaseNode
	Uses []UseClause `json:"uses"` // 支持多个use子句，如 use A, B, C;
}

type UseClause struct {
	Name  *NamespaceNameExpression `json:"name"`  // 导入的名称
	Alias string                   `json:"alias"` // 别名 (如 use Foo as Bar)
	Type  string                   `json:"type"`  // 类型: "class", "function", "const" 或空字符串
}

func NewUseStatement(pos lexer.Position) *UseStatement {
	return &UseStatement{
		BaseNode: BaseNode{
			Kind:     ASTUse,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Uses: make([]UseClause, 0),
	}
}

// GetChildren 返回子节点
func (u *UseStatement) GetChildren() []Node {
	children := make([]Node, 0)
	for _, use := range u.Uses {
		if use.Name != nil {
			children = append(children, use.Name)
		}
	}
	return children
}

// Accept 接受访问者
func (u *UseStatement) Accept(visitor Visitor) {
	if visitor.Visit(u) {
		for _, use := range u.Uses {
			if use.Name != nil {
				use.Name.Accept(visitor)
			}
		}
	}
}

func (u *UseStatement) statementNode() {}

func (u *UseStatement) String() string {
	result := "use "
	clauses := make([]string, len(u.Uses))
	for i, clause := range u.Uses {
		clauseStr := ""
		if clause.Type != "" {
			clauseStr += clause.Type + " "
		}
		clauseStr += clause.Name.String()
		if clause.Alias != "" {
			clauseStr += " as " + clause.Alias
		}
		clauses[i] = clauseStr
	}
	result += strings.Join(clauses, ", ") + ";"
	return result
}

// InterfaceDeclaration 表示接口声明
type InterfaceDeclaration struct {
	BaseNode
	Name       *IdentifierNode   `json:"name"`                // 接口名称
	Extends    []*IdentifierNode `json:"extends,omitempty"`   // 继承的接口
	Methods    []*InterfaceMethod `json:"methods"`             // 接口方法
}

// InterfaceMethod 表示接口方法声明
type InterfaceMethod struct {
	Name       *IdentifierNode `json:"name"`                // 方法名称
	Parameters []Parameter     `json:"parameters"`          // 参数列表
	ReturnType *TypeHint       `json:"returnType,omitempty"` // 返回类型
	Visibility string          `json:"visibility"`          // 可见性 (通常是 public)
}

func NewInterfaceDeclaration(pos lexer.Position, name *IdentifierNode) *InterfaceDeclaration {
	return &InterfaceDeclaration{
		BaseNode: BaseNode{
			Kind:     ASTInterface,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Name:    name,
		Extends: make([]*IdentifierNode, 0),
		Methods: make([]*InterfaceMethod, 0),
	}
}

// GetChildren 返回子节点
func (i *InterfaceDeclaration) GetChildren() []Node {
	children := make([]Node, 0)
	if i.Name != nil {
		children = append(children, i.Name)
	}
	for _, extend := range i.Extends {
		children = append(children, extend)
	}
	// 接口方法不直接实现 Node 接口，所以不添加到子节点中
	return children
}

// Accept 接受访问者
func (i *InterfaceDeclaration) Accept(visitor Visitor) {
	if visitor.Visit(i) {
		if i.Name != nil {
			i.Name.Accept(visitor)
		}
		for _, extend := range i.Extends {
			extend.Accept(visitor)
		}
	}
}

func (i *InterfaceDeclaration) statementNode() {}

func (i *InterfaceDeclaration) String() string {
	result := "interface " + i.Name.String()
	
	if len(i.Extends) > 0 {
		result += " extends "
		extendNames := make([]string, len(i.Extends))
		for idx, extend := range i.Extends {
			extendNames[idx] = extend.String()
		}
		result += strings.Join(extendNames, ", ")
	}
	
	result += " {\n"
	for _, method := range i.Methods {
		result += "  " + method.Visibility + " function " + method.Name.String() + "("
		paramStrs := make([]string, len(method.Parameters))
		for idx, param := range method.Parameters {
			paramStr := ""
			if param.Type != nil {
				paramStr += param.Type.String() + " "
			}
			if param.ByReference {
				paramStr += "&"
			}
			paramStr += "$" + param.Name
			paramStrs[idx] = paramStr
		}
		result += strings.Join(paramStrs, ", ")
		result += ")"
		if method.ReturnType != nil {
			result += ": " + method.ReturnType.String()
		}
		result += ";\n"
	}
	result += "}"
	return result
}

// TraitDeclaration 表示 trait 声明
type TraitDeclaration struct {
	BaseNode
	Name       *IdentifierNode       `json:"name"`       // trait 名称
	Properties []*PropertyDeclaration `json:"properties"` // trait 属性
	Methods    []*FunctionDeclaration `json:"methods"`    // trait 方法
}

func NewTraitDeclaration(pos lexer.Position, name *IdentifierNode) *TraitDeclaration {
	return &TraitDeclaration{
		BaseNode: BaseNode{
			Kind:     ASTTrait,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Name:       name,
		Properties: make([]*PropertyDeclaration, 0),
		Methods:    make([]*FunctionDeclaration, 0),
	}
}

// GetChildren 返回子节点
func (t *TraitDeclaration) GetChildren() []Node {
	children := make([]Node, 0)
	if t.Name != nil {
		children = append(children, t.Name)
	}
	for _, prop := range t.Properties {
		children = append(children, prop)
	}
	for _, method := range t.Methods {
		children = append(children, method)
	}
	return children
}

// Accept 接受访问者
func (t *TraitDeclaration) Accept(visitor Visitor) {
	if visitor.Visit(t) {
		if t.Name != nil {
			t.Name.Accept(visitor)
		}
		for _, prop := range t.Properties {
			prop.Accept(visitor)
		}
		for _, method := range t.Methods {
			method.Accept(visitor)
		}
	}
}

func (t *TraitDeclaration) statementNode() {}

func (t *TraitDeclaration) String() string {
	result := "trait " + t.Name.String() + " {\n"
	
	// 添加属性
	for _, prop := range t.Properties {
		result += "  " + prop.String() + ";\n"
	}
	
	// 添加方法
	for _, method := range t.Methods {
		result += "  " + method.String() + "\n"
	}
	
	result += "}"
	return result
}

// EnumDeclaration 表示 enum 声明 (PHP 8.1+)
type EnumDeclaration struct {
	BaseNode
	Name        *IdentifierNode   `json:"name"`                  // enum 名称
	BackingType *TypeHint         `json:"backingType,omitempty"` // 可选的支撑类型 (string, int)
	Implements  []*IdentifierNode `json:"implements,omitempty"`  // 实现的接口
	Cases       []*EnumCase       `json:"cases"`                 // enum 案例
	Methods     []*FunctionDeclaration `json:"methods,omitempty"` // enum 方法
}

// EnumCase 表示 enum 案例
type EnumCase struct {
	Name  *IdentifierNode `json:"name"`            // 案例名称
	Value Expression      `json:"value,omitempty"` // 可选的值（对于支撑枚举）
}

func NewEnumDeclaration(pos lexer.Position, name *IdentifierNode) *EnumDeclaration {
	return &EnumDeclaration{
		BaseNode: BaseNode{
			Kind:     ASTEnum,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Name:       name,
		Implements: make([]*IdentifierNode, 0),
		Cases:      make([]*EnumCase, 0),
		Methods:    make([]*FunctionDeclaration, 0),
	}
}

func NewEnumCase(name *IdentifierNode, value Expression) *EnumCase {
	return &EnumCase{
		Name:  name,
		Value: value,
	}
}

// GetChildren 返回子节点
func (e *EnumDeclaration) GetChildren() []Node {
	children := make([]Node, 0)
	if e.Name != nil {
		children = append(children, e.Name)
	}
	if e.BackingType != nil {
		children = append(children, e.BackingType)
	}
	for _, impl := range e.Implements {
		children = append(children, impl)
	}
	// EnumCase 不直接实现 Node 接口
	for _, method := range e.Methods {
		children = append(children, method)
	}
	return children
}

// Accept 接受访问者
func (e *EnumDeclaration) Accept(visitor Visitor) {
	if visitor.Visit(e) {
		if e.Name != nil {
			e.Name.Accept(visitor)
		}
		if e.BackingType != nil {
			e.BackingType.Accept(visitor)
		}
		for _, impl := range e.Implements {
			impl.Accept(visitor)
		}
		for _, method := range e.Methods {
			method.Accept(visitor)
		}
	}
}

func (e *EnumDeclaration) statementNode() {}

func (e *EnumDeclaration) String() string {
	result := "enum " + e.Name.String()
	
	// 添加支撑类型
	if e.BackingType != nil {
		result += ": " + e.BackingType.String()
	}
	
	// 添加接口实现
	if len(e.Implements) > 0 {
		result += " implements "
		implNames := make([]string, len(e.Implements))
		for i, impl := range e.Implements {
			implNames[i] = impl.String()
		}
		result += strings.Join(implNames, ", ")
	}
	
	result += " {\n"
	
	// 添加案例
	for _, enumCase := range e.Cases {
		result += "  case " + enumCase.Name.String()
		if enumCase.Value != nil {
			result += " = " + enumCase.Value.String()
		}
		result += ";\n"
	}
	
	// 添加方法
	if len(e.Methods) > 0 && len(e.Cases) > 0 {
		result += "\n" // 在案例和方法之间添加空行
	}
	for _, method := range e.Methods {
		result += "  " + method.String() + "\n"
	}
	
	result += "}"
	return result
}

// PropertyAccessExpression 表示属性访问表达式 ($obj->property)
type PropertyAccessExpression struct {
	BaseNode
	Object   Expression      `json:"object"`   // 对象表达式
	Property *IdentifierNode `json:"property"` // 属性名称
}

func NewPropertyAccessExpression(pos lexer.Position, object Expression, property *IdentifierNode) *PropertyAccessExpression {
	return &PropertyAccessExpression{
		BaseNode: BaseNode{
			Kind:     ASTProp,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Object:   object,
		Property: property,
	}
}

// GetChildren 返回子节点
func (p *PropertyAccessExpression) GetChildren() []Node {
	children := make([]Node, 0)
	if p.Object != nil {
		children = append(children, p.Object)
	}
	if p.Property != nil {
		children = append(children, p.Property)
	}
	return children
}

// Accept 接受访问者
func (p *PropertyAccessExpression) Accept(visitor Visitor) {
	if visitor.Visit(p) {
		if p.Object != nil {
			p.Object.Accept(visitor)
		}
		if p.Property != nil {
			p.Property.Accept(visitor)
		}
	}
}

func (p *PropertyAccessExpression) expressionNode() {}

func (p *PropertyAccessExpression) String() string {
	return p.Object.String() + "->" + p.Property.String()
}

// NullsafePropertyAccessExpression 表示空安全属性访问表达式 ($obj?->property)
type NullsafePropertyAccessExpression struct {
	BaseNode
	Object   Expression      `json:"object"`   // 对象表达式
	Property *IdentifierNode `json:"property"` // 属性名称
}

func NewNullsafePropertyAccessExpression(pos lexer.Position, object Expression, property *IdentifierNode) *NullsafePropertyAccessExpression {
	return &NullsafePropertyAccessExpression{
		BaseNode: BaseNode{
			Kind:     ASTNullsafeProp,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Object:   object,
		Property: property,
	}
}

// GetChildren 返回子节点
func (n *NullsafePropertyAccessExpression) GetChildren() []Node {
	children := make([]Node, 0)
	if n.Object != nil {
		children = append(children, n.Object)
	}
	if n.Property != nil {
		children = append(children, n.Property)
	}
	return children
}

// Accept 接受访问者
func (n *NullsafePropertyAccessExpression) Accept(visitor Visitor) {
	if visitor.Visit(n) {
		if n.Object != nil {
			n.Object.Accept(visitor)
		}
		if n.Property != nil {
			n.Property.Accept(visitor)
		}
	}
}

func (n *NullsafePropertyAccessExpression) expressionNode() {}

func (n *NullsafePropertyAccessExpression) String() string {
	return n.Object.String() + "?->" + n.Property.String()
}

// ExpressionStatement 表达式语句
type ExpressionStatement struct {
	BaseNode
	Expression Expression `json:"expression"`
}

func NewExpressionStatement(pos lexer.Position, expr Expression) *ExpressionStatement {
	return &ExpressionStatement{
		BaseNode: BaseNode{
			Kind:     ASTExprList,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Expression: expr,
	}
}

// GetChildren 返回子节点
func (es *ExpressionStatement) GetChildren() []Node {
	if es.Expression != nil {
		return []Node{es.Expression}
	}
	return nil
}

// Accept 接受访问者
func (es *ExpressionStatement) Accept(visitor Visitor) {
	if visitor.Visit(es) && es.Expression != nil {
		es.Expression.Accept(visitor)
	}
}

func (es *ExpressionStatement) statementNode() {}

func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String() + ";"
	}
	return ""
}

// AssignmentExpression 赋值表达式
type AssignmentExpression struct {
	BaseNode
	Left     Expression `json:"left"`
	Right    Expression `json:"right"`
	Operator string     `json:"operator"`
}

func NewAssignmentExpression(pos lexer.Position, left Expression, operator string, right Expression) *AssignmentExpression {
	return &AssignmentExpression{
		BaseNode: BaseNode{
			Kind:     ASTAssign,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Left:     left,
		Right:    right,
		Operator: operator,
	}
}

// GetChildren 返回子节点
func (ae *AssignmentExpression) GetChildren() []Node {
	return []Node{ae.Left, ae.Right}
}

// Accept 接受访问者
func (ae *AssignmentExpression) Accept(visitor Visitor) {
	if visitor.Visit(ae) {
		ae.Left.Accept(visitor)
		ae.Right.Accept(visitor)
	}
}

func (ae *AssignmentExpression) expressionNode() {}

func (ae *AssignmentExpression) String() string {
	return ae.Left.String() + " " + ae.Operator + " " + ae.Right.String()
}

// BinaryExpression 二元表达式
type BinaryExpression struct {
	BaseNode
	Left     Expression `json:"left"`
	Right    Expression `json:"right"`
	Operator string     `json:"operator"`
}

func NewBinaryExpression(pos lexer.Position, left Expression, operator string, right Expression) *BinaryExpression {
	return &BinaryExpression{
		BaseNode: BaseNode{
			Kind:     ASTBinaryOp,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Left:     left,
		Right:    right,
		Operator: operator,
	}
}

// GetChildren 返回子节点
func (be *BinaryExpression) GetChildren() []Node {
	return []Node{be.Left, be.Right}
}

// Accept 接受访问者
func (be *BinaryExpression) Accept(visitor Visitor) {
	if visitor.Visit(be) {
		be.Left.Accept(visitor)
		be.Right.Accept(visitor)
	}
}

func (be *BinaryExpression) expressionNode() {}

func (be *BinaryExpression) String() string {
	return "(" + be.Left.String() + " " + be.Operator + " " + be.Right.String() + ")"
}

// UnaryExpression 一元表达式
type UnaryExpression struct {
	BaseNode
	Operator string     `json:"operator"`
	Operand  Expression `json:"operand"`
	Prefix   bool       `json:"prefix"` // true for ++$a, false for $a++
}

// CastExpression 类型转换表达式节点
type CastExpression struct {
	BaseNode
	CastType string     `json:"castType"`
	Operand  Expression `json:"operand"`
}

func NewUnaryExpression(pos lexer.Position, operator string, operand Expression, prefix bool) *UnaryExpression {
	kind := ASTUnaryOp
	if operator == "++" {
		if prefix {
			kind = ASTPreInc
		} else {
			kind = ASTPostInc
		}
	} else if operator == "--" {
		if prefix {
			kind = ASTPreDec
		} else {
			kind = ASTPostDec
		}
	}

	return &UnaryExpression{
		BaseNode: BaseNode{
			Kind:     kind,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Operator: operator,
		Operand:  operand,
		Prefix:   prefix,
	}
}

// NewCastExpression 创建类型转换表达式
func NewCastExpression(pos lexer.Position, castType string, operand Expression) *CastExpression {
	return &CastExpression{
		BaseNode: BaseNode{
			Kind:     ASTCast,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		CastType: castType,
		Operand:  operand,
	}
}

// GetChildren 返回子节点
func (ue *UnaryExpression) GetChildren() []Node {
	return []Node{ue.Operand}
}

// Accept 接受访问者
func (ue *UnaryExpression) Accept(visitor Visitor) {
	if visitor.Visit(ue) {
		ue.Operand.Accept(visitor)
	}
}

func (ue *UnaryExpression) expressionNode() {}

func (ue *UnaryExpression) String() string {
	if ue.Prefix {
		return ue.Operator + ue.Operand.String()
	}
	return ue.Operand.String() + ue.Operator
}

// GetChildren 返回子节点
func (ce *CastExpression) GetChildren() []Node {
	return []Node{ce.Operand}
}

// Accept 接受访问者
func (ce *CastExpression) Accept(visitor Visitor) {
	if visitor.Visit(ce) {
		ce.Operand.Accept(visitor)
	}
}

func (ce *CastExpression) expressionNode() {}

func (ce *CastExpression) String() string {
	return ce.CastType + " " + ce.Operand.String()
}

// Variable 变量
type Variable struct {
	BaseNode
	Name string `json:"name"`
}

func NewVariable(pos lexer.Position, name string) *Variable {
	return &Variable{
		BaseNode: BaseNode{
			Kind:     ASTVar,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Name: name,
	}
}

// GetChildren 返回子节点
func (v *Variable) GetChildren() []Node {
	return nil // 变量是叶子节点
}

// Accept 接受访问者
func (v *Variable) Accept(visitor Visitor) {
	visitor.Visit(v)
}

func (v *Variable) expressionNode() {}
func (v *Variable) identifierNode() {}

func (v *Variable) String() string {
	return v.Name
}

func (v *Variable) ToJSON() ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

// StringLiteral 字符串字面量
type StringLiteral struct {
	BaseNode
	Value string `json:"value"`
	Raw   string `json:"raw"` // 原始字符串（包含引号）
}

func NewStringLiteral(pos lexer.Position, value, raw string) *StringLiteral {
	return &StringLiteral{
		BaseNode: BaseNode{
			Kind:     ASTZval,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Value: value,
		Raw:   raw,
	}
}

// GetChildren 返回子节点
func (sl *StringLiteral) GetChildren() []Node {
	return nil // 字符串字面量是叶子节点
}

// Accept 接受访问者
func (sl *StringLiteral) Accept(visitor Visitor) {
	visitor.Visit(sl)
}

func (sl *StringLiteral) expressionNode() {}

func (sl *StringLiteral) String() string {
	return sl.Raw
}

// InterpolatedStringExpression 字符串插值表达式
type InterpolatedStringExpression struct {
	BaseNode
	Parts []Expression `json:"parts"` // 字符串的各个部分
}

func NewInterpolatedStringExpression(pos lexer.Position, parts []Expression) *InterpolatedStringExpression {
	return &InterpolatedStringExpression{
		BaseNode: BaseNode{
			Kind:     ASTEncapsList,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Parts: parts,
	}
}

// GetChildren 返回子节点
func (ise *InterpolatedStringExpression) GetChildren() []Node {
	nodes := make([]Node, len(ise.Parts))
	for i, part := range ise.Parts {
		nodes[i] = part
	}
	return nodes
}

// Accept 接受访问者
func (ise *InterpolatedStringExpression) Accept(visitor Visitor) {
	visitor.Visit(ise)
}

func (ise *InterpolatedStringExpression) expressionNode() {}

func (ise *InterpolatedStringExpression) String() string {
	var parts []string
	for _, part := range ise.Parts {
		parts = append(parts, part.String())
	}
	return `"` + strings.Join(parts, "") + `"`
}

// NumberLiteral 数字字面量
type NumberLiteral struct {
	BaseNode
	Value string `json:"value"`
	Kind  string `json:"kind"` // "integer" or "float"
}

func NewNumberLiteral(pos lexer.Position, value, kind string) *NumberLiteral {
	return &NumberLiteral{
		BaseNode: BaseNode{
			Kind:     ASTZval,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Value: value,
		Kind:  kind,
	}
}

// GetChildren 返回子节点
func (nl *NumberLiteral) GetChildren() []Node {
	return nil // 数字字面量是叶子节点
}

// Accept 接受访问者
func (nl *NumberLiteral) Accept(visitor Visitor) {
	visitor.Visit(nl)
}

func (nl *NumberLiteral) expressionNode() {}

func (nl *NumberLiteral) String() string {
	return nl.Value
}

// BooleanLiteral 布尔字面量
type BooleanLiteral struct {
	BaseNode
	Value bool `json:"value"`
}

func NewBooleanLiteral(pos lexer.Position, value bool) *BooleanLiteral {
	return &BooleanLiteral{
		BaseNode: BaseNode{
			Kind:     ASTZval,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Value: value,
	}
}

// GetChildren 返回子节点
func (bl *BooleanLiteral) GetChildren() []Node {
	return nil // 布尔字面量是叶子节点
}

// Accept 接受访问者
func (bl *BooleanLiteral) Accept(visitor Visitor) {
	visitor.Visit(bl)
}

func (bl *BooleanLiteral) expressionNode() {}

func (bl *BooleanLiteral) String() string {
	if bl.Value {
		return "true"
	}
	return "false"
}

// NullLiteral null字面量
type NullLiteral struct {
	BaseNode
}

func NewNullLiteral(pos lexer.Position) *NullLiteral {
	return &NullLiteral{
		BaseNode: BaseNode{
			Kind:     ASTZval,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
	}
}

// GetChildren 返回子节点
func (nl *NullLiteral) GetChildren() []Node {
	return nil // null字面量是叶子节点
}

// Accept 接受访问者
func (nl *NullLiteral) Accept(visitor Visitor) {
	visitor.Visit(nl)
}

func (nl *NullLiteral) expressionNode() {}

func (nl *NullLiteral) String() string {
	return "null"
}

// ArrayExpression 数组表达式
type ArrayExpression struct {
	BaseNode
	Elements []Expression `json:"elements"`
}

func NewArrayExpression(pos lexer.Position) *ArrayExpression {
	return &ArrayExpression{
		BaseNode: BaseNode{
			Kind:     ASTArray,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Elements: make([]Expression, 0),
	}
}

// GetChildren 返回子节点
func (ae *ArrayExpression) GetChildren() []Node {
	children := make([]Node, len(ae.Elements))
	for i, elem := range ae.Elements {
		children[i] = elem
	}
	return children
}

// Accept 接受访问者
func (ae *ArrayExpression) Accept(visitor Visitor) {
	if visitor.Visit(ae) {
		for _, elem := range ae.Elements {
			if elem != nil {
				elem.Accept(visitor)
			}
		}
	}
}

func (ae *ArrayExpression) expressionNode() {}

func (ae *ArrayExpression) String() string {
	var elements []string
	for _, elem := range ae.Elements {
		if elem != nil {
			elements = append(elements, elem.String())
		}
	}
	return "[" + strings.Join(elements, ", ") + "]"
}

// IfStatement if语句
type IfStatement struct {
	BaseNode
	Test       Expression  `json:"test"`
	Consequent []Statement `json:"consequent"`
	Alternate  []Statement `json:"alternate,omitempty"`
}

func NewIfStatement(pos lexer.Position, test Expression) *IfStatement {
	return &IfStatement{
		BaseNode: BaseNode{
			Kind:     ASTIf,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Test:       test,
		Consequent: make([]Statement, 0),
		Alternate:  make([]Statement, 0),
	}
}

// GetChildren 返回子节点
func (is *IfStatement) GetChildren() []Node {
	children := []Node{is.Test}
	for _, stmt := range is.Consequent {
		children = append(children, stmt)
	}
	for _, stmt := range is.Alternate {
		children = append(children, stmt)
	}
	return children
}

// Accept 接受访问者
func (is *IfStatement) Accept(visitor Visitor) {
	if visitor.Visit(is) {
		is.Test.Accept(visitor)
		for _, stmt := range is.Consequent {
			stmt.Accept(visitor)
		}
		for _, stmt := range is.Alternate {
			stmt.Accept(visitor)
		}
	}
}

func (is *IfStatement) statementNode() {}

func (is *IfStatement) String() string {
	var out strings.Builder
	out.WriteString("if (")
	out.WriteString(is.Test.String())
	out.WriteString(") {\n")

	for _, stmt := range is.Consequent {
		out.WriteString("  " + stmt.String() + "\n")
	}
	out.WriteString("}")

	if len(is.Alternate) > 0 {
		out.WriteString(" else {\n")
		for _, stmt := range is.Alternate {
			out.WriteString("  " + stmt.String() + "\n")
		}
		out.WriteString("}")
	}

	return out.String()
}

// WhileStatement while语句
type WhileStatement struct {
	BaseNode
	Test Expression  `json:"test"`
	Body []Statement `json:"body"`
}

func NewWhileStatement(pos lexer.Position, test Expression) *WhileStatement {
	return &WhileStatement{
		BaseNode: BaseNode{
			Kind:     ASTWhile,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Test: test,
		Body: make([]Statement, 0),
	}
}

// GetChildren 返回子节点
func (ws *WhileStatement) GetChildren() []Node {
	children := []Node{ws.Test}
	for _, stmt := range ws.Body {
		children = append(children, stmt)
	}
	return children
}

// Accept 接受访问者
func (ws *WhileStatement) Accept(visitor Visitor) {
	if visitor.Visit(ws) {
		ws.Test.Accept(visitor)
		for _, stmt := range ws.Body {
			stmt.Accept(visitor)
		}
	}
}

func (ws *WhileStatement) statementNode() {}

func (ws *WhileStatement) String() string {
	var out strings.Builder
	out.WriteString("while (")
	out.WriteString(ws.Test.String())
	out.WriteString(") {\n")

	for _, stmt := range ws.Body {
		out.WriteString("  " + stmt.String() + "\n")
	}
	out.WriteString("}")

	return out.String()
}

// ForStatement for语句
type ForStatement struct {
	BaseNode
	Init   Expression  `json:"init,omitempty"`
	Test   Expression  `json:"test,omitempty"`
	Update Expression  `json:"update,omitempty"`
	Body   []Statement `json:"body"`
}

func NewForStatement(pos lexer.Position) *ForStatement {
	return &ForStatement{
		BaseNode: BaseNode{
			Kind:     ASTFor,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Body: make([]Statement, 0),
	}
}

// GetChildren 返回子节点
func (fs *ForStatement) GetChildren() []Node {
	children := make([]Node, 0)
	if fs.Init != nil {
		children = append(children, fs.Init)
	}
	if fs.Test != nil {
		children = append(children, fs.Test)
	}
	if fs.Update != nil {
		children = append(children, fs.Update)
	}
	for _, stmt := range fs.Body {
		children = append(children, stmt)
	}
	return children
}

// Accept 接受访问者
func (fs *ForStatement) Accept(visitor Visitor) {
	if visitor.Visit(fs) {
		if fs.Init != nil {
			fs.Init.Accept(visitor)
		}
		if fs.Test != nil {
			fs.Test.Accept(visitor)
		}
		if fs.Update != nil {
			fs.Update.Accept(visitor)
		}
		for _, stmt := range fs.Body {
			stmt.Accept(visitor)
		}
	}
}

func (fs *ForStatement) statementNode() {}

func (fs *ForStatement) String() string {
	var out strings.Builder
	out.WriteString("for (")

	if fs.Init != nil {
		out.WriteString(fs.Init.String())
	}
	out.WriteString("; ")

	if fs.Test != nil {
		out.WriteString(fs.Test.String())
	}
	out.WriteString("; ")

	if fs.Update != nil {
		out.WriteString(fs.Update.String())
	}

	out.WriteString(") {\n")
	for _, stmt := range fs.Body {
		out.WriteString("  " + stmt.String() + "\n")
	}
	out.WriteString("}")

	return out.String()
}

// FunctionDeclaration 函数声明
type FunctionDeclaration struct {
	BaseNode
	Name         Identifier  `json:"name"`
	Parameters   []Parameter `json:"parameters"`
	ReturnType   *TypeHint   `json:"returnType,omitempty"`
	Body         []Statement `json:"body"`
	ByReference  bool        `json:"byReference,omitempty"`   // function &foo()
	Visibility   string      `json:"visibility,omitempty"`   // public, private, protected (for class methods)
}

type Parameter struct {
	Name         string     `json:"name"`
	DefaultValue Expression `json:"defaultValue,omitempty"`
	Type         *TypeHint  `json:"type,omitempty"`
	ByReference  bool       `json:"byReference,omitempty"`   // &$param
	Variadic     bool       `json:"variadic,omitempty"`      // ...$params
	Visibility   string     `json:"visibility,omitempty"`   // public, private, protected
	ReadOnly     bool       `json:"readOnly,omitempty"`     // readonly
}

// TypeHint represents a PHP type hint
type TypeHint struct {
	BaseNode
	Name         string      `json:"name"`                  // Simple type name
	Nullable     bool        `json:"nullable,omitempty"`    // ?Type
	UnionTypes   []*TypeHint `json:"unionTypes,omitempty"`  // Type1|Type2
	IntersectionTypes []*TypeHint `json:"intersectionTypes,omitempty"` // Type1&Type2
}

func NewFunctionDeclaration(pos lexer.Position, name Identifier) *FunctionDeclaration {
	return &FunctionDeclaration{
		BaseNode: BaseNode{
			Kind:     ASTFuncDecl,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Name:       name,
		Parameters: make([]Parameter, 0),
		Body:       make([]Statement, 0),
	}
}

// NewTypeHint creates a new type hint
func NewTypeHint(pos lexer.Position, name string) *TypeHint {
	return &TypeHint{
		BaseNode: BaseNode{
			Kind:     ASTType,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Name: name,
	}
}

// NewSimpleTypeHint creates a simple type hint (possibly nullable)
func NewSimpleTypeHint(pos lexer.Position, name string, nullable bool) *TypeHint {
	return &TypeHint{
		BaseNode: BaseNode{
			Kind:     ASTType,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Name:     name,
		Nullable: nullable,
	}
}

// NewUnionTypeHint creates a union type hint (Type1|Type2)
func NewUnionTypeHint(pos lexer.Position, types []*TypeHint) *TypeHint {
	return &TypeHint{
		BaseNode: BaseNode{
			Kind:     ASTTypeUnion,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		UnionTypes: types,
	}
}

// NewIntersectionTypeHint creates an intersection type hint (Type1&Type2)
func NewIntersectionTypeHint(pos lexer.Position, types []*TypeHint) *TypeHint {
	return &TypeHint{
		BaseNode: BaseNode{
			Kind:     ASTTypeIntersection,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		IntersectionTypes: types,
	}
}

// TypeHint AST interface implementations
func (th *TypeHint) GetChildren() []Node {
	var children []Node
	for _, t := range th.UnionTypes {
		children = append(children, t)
	}
	for _, t := range th.IntersectionTypes {
		children = append(children, t)
	}
	return children
}

func (th *TypeHint) Accept(visitor Visitor) {
	visitor.Visit(th)
	for _, child := range th.GetChildren() {
		child.Accept(visitor)
	}
}

func (th *TypeHint) String() string {
	if len(th.UnionTypes) > 0 {
		var types []string
		for _, t := range th.UnionTypes {
			types = append(types, t.String())
		}
		result := strings.Join(types, "|")
		if th.Nullable {
			result = "?" + result
		}
		return result
	}
	if len(th.IntersectionTypes) > 0 {
		var types []string
		for _, t := range th.IntersectionTypes {
			types = append(types, t.String())
		}
		return strings.Join(types, "&")
	}
	result := th.Name
	if th.Nullable {
		result = "?" + result
	}
	return result
}

// GetChildren 返回子节点
func (fd *FunctionDeclaration) GetChildren() []Node {
	children := []Node{fd.Name}
	if fd.ReturnType != nil {
		children = append(children, fd.ReturnType)
	}
	for _, param := range fd.Parameters {
		if param.Type != nil {
			children = append(children, param.Type)
		}
		if param.DefaultValue != nil {
			children = append(children, param.DefaultValue)
		}
	}
	for _, stmt := range fd.Body {
		children = append(children, stmt)
	}
	return children
}

// Accept 接受访问者
func (fd *FunctionDeclaration) Accept(visitor Visitor) {
	if visitor.Visit(fd) {
		fd.Name.Accept(visitor)
		for _, param := range fd.Parameters {
			if param.DefaultValue != nil {
				param.DefaultValue.Accept(visitor)
			}
		}
		for _, stmt := range fd.Body {
			stmt.Accept(visitor)
		}
	}
}

func (fd *FunctionDeclaration) statementNode() {}

func (fd *FunctionDeclaration) String() string {
	var out strings.Builder
	out.WriteString("function ")
	if fd.Name != nil {
		out.WriteString(fd.Name.String())
	}
	out.WriteString("(")

	var params []string
	for _, param := range fd.Parameters {
		paramStr := param.Name
		if param.DefaultValue != nil {
			paramStr += " = " + param.DefaultValue.String()
		}
		params = append(params, paramStr)
	}
	out.WriteString(strings.Join(params, ", "))

	out.WriteString(") {\n")
	for _, stmt := range fd.Body {
		if stmt != nil {
			out.WriteString("  " + stmt.String() + "\n")
		}
	}
	out.WriteString("}")

	return out.String()
}

// ArrowFunctionExpression 箭头函数表达式 (PHP 7.4+)
type ArrowFunctionExpression struct {
	BaseNode
	Parameters []Parameter `json:"parameters"`
	ReturnType *TypeHint   `json:"returnType,omitempty"`
	Body       Expression  `json:"body"`
	Static     bool        `json:"static,omitempty"`
}

func NewArrowFunctionExpression(pos lexer.Position, parameters []Parameter, returnType *TypeHint, body Expression, static bool) *ArrowFunctionExpression {
	return &ArrowFunctionExpression{
		BaseNode: BaseNode{
			Kind:     ASTArrowFunc,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Parameters: parameters,
		ReturnType: returnType,
		Body:       body,
		Static:     static,
	}
}

// GetChildren 返回子节点
func (af *ArrowFunctionExpression) GetChildren() []Node {
	children := make([]Node, 0)
	if af.ReturnType != nil {
		children = append(children, af.ReturnType)
	}
	for _, param := range af.Parameters {
		if param.Type != nil {
			children = append(children, param.Type)
		}
		if param.DefaultValue != nil {
			children = append(children, param.DefaultValue)
		}
	}
	if af.Body != nil {
		children = append(children, af.Body)
	}
	return children
}

// Accept 接受访问者
func (af *ArrowFunctionExpression) Accept(visitor Visitor) {
	if visitor.Visit(af) {
		if af.ReturnType != nil {
			af.ReturnType.Accept(visitor)
		}
		for _, param := range af.Parameters {
			if param.DefaultValue != nil {
				param.DefaultValue.Accept(visitor)
			}
		}
		if af.Body != nil {
			af.Body.Accept(visitor)
		}
	}
}

func (af *ArrowFunctionExpression) expressionNode() {}

func (af *ArrowFunctionExpression) String() string {
	var out strings.Builder
	if af.Static {
		out.WriteString("static ")
	}
	out.WriteString("fn(")
	for i, param := range af.Parameters {
		if i > 0 {
			out.WriteString(", ")
		}
		if param.Type != nil {
			out.WriteString(param.Type.String() + " ")
		}
		if param.ByReference {
			out.WriteString("&")
		}
		out.WriteString("$" + param.Name)
		if param.DefaultValue != nil {
			out.WriteString(" = " + param.DefaultValue.String())
		}
	}
	out.WriteString(")")
	if af.ReturnType != nil {
		out.WriteString(": " + af.ReturnType.String())
	}
	out.WriteString(" => ")
	if af.Body != nil {
		out.WriteString(af.Body.String())
	}
	return out.String()
}

// IdentifierNode 标识符节点
type IdentifierNode struct {
	BaseNode
	Name string `json:"name"`
}

func NewIdentifierNode(pos lexer.Position, name string) *IdentifierNode {
	return &IdentifierNode{
		BaseNode: BaseNode{
			Kind:     ASTConst, // 标识符通常用作常量
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Name: name,
	}
}

// GetChildren 返回子节点
func (i *IdentifierNode) GetChildren() []Node {
	return nil // 标识符是叶子节点
}

// Accept 接受访问者
func (i *IdentifierNode) Accept(visitor Visitor) {
	visitor.Visit(i)
}

func (i *IdentifierNode) expressionNode() {}
func (i *IdentifierNode) identifierNode() {}

func (i *IdentifierNode) String() string {
	return i.Name
}

// NamespaceNameExpression 表示命名空间名称表达式
type NamespaceNameExpression struct {
	BaseNode
	Parts []string `json:"parts"` // 命名空间的各部分，如 ["Foo", "Bar"] 表示 Foo\Bar
}

func NewNamespaceNameExpression(pos lexer.Position, parts []string) *NamespaceNameExpression {
	return &NamespaceNameExpression{
		BaseNode: BaseNode{
			Kind:     ASTNamespaceName,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Parts: parts,
	}
}

// GetChildren 返回子节点
func (n *NamespaceNameExpression) GetChildren() []Node {
	return nil // 命名空间名称是叶子节点
}

// Accept 接受访问者
func (n *NamespaceNameExpression) Accept(visitor Visitor) {
	visitor.Visit(n)
}

func (n *NamespaceNameExpression) expressionNode() {}

func (n *NamespaceNameExpression) String() string {
	return strings.Join(n.Parts, "\\")
}

// PropertyDeclaration 属性声明语句
type PropertyDeclaration struct {
	BaseNode
	Visibility   string      `json:"visibility"`   // private, protected, public
	ReadOnly     bool        `json:"readOnly,omitempty"`     // readonly
	Type         *TypeHint   `json:"type,omitempty"`
	Name         string      `json:"name"`         // Property name without $
	DefaultValue Expression  `json:"defaultValue,omitempty"`
}

func NewPropertyDeclaration(pos lexer.Position, visibility, name string, readOnly bool, typeHint *TypeHint, defaultValue Expression) *PropertyDeclaration {
	return &PropertyDeclaration{
		BaseNode: BaseNode{
			Kind:     ASTPropertyDecl,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Visibility:   visibility,
		ReadOnly:     readOnly,
		Type:         typeHint,
		Name:         name,
		DefaultValue: defaultValue,
	}
}

func (pd *PropertyDeclaration) GetChildren() []Node {
	var children []Node
	if pd.Type != nil {
		children = append(children, pd.Type)
	}
	if pd.DefaultValue != nil {
		children = append(children, pd.DefaultValue)
	}
	return children
}

func (pd *PropertyDeclaration) Accept(visitor Visitor) {
	if visitor.Visit(pd) {
		if pd.Type != nil {
			pd.Type.Accept(visitor)
		}
		if pd.DefaultValue != nil {
			pd.DefaultValue.Accept(visitor)
		}
	}
}

func (pd *PropertyDeclaration) statementNode() {}

func (pd *PropertyDeclaration) String() string {
	result := pd.Visibility
	if pd.ReadOnly {
		result += " readonly"
	}
	if pd.Type != nil {
		result += " " + pd.Type.String()
	}
	result += " $" + pd.Name
	if pd.DefaultValue != nil {
		result += " = " + pd.DefaultValue.String()
	}
	return result
}

// ReturnStatement return语句
type ReturnStatement struct {
	BaseNode
	Argument Expression `json:"argument,omitempty"`
}

func NewReturnStatement(pos lexer.Position, arg Expression) *ReturnStatement {
	return &ReturnStatement{
		BaseNode: BaseNode{
			Kind:     ASTReturn,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Argument: arg,
	}
}

// GetChildren 返回子节点
func (rs *ReturnStatement) GetChildren() []Node {
	if rs.Argument != nil {
		return []Node{rs.Argument}
	}
	return nil
}

// Accept 接受访问者
func (rs *ReturnStatement) Accept(visitor Visitor) {
	if visitor.Visit(rs) && rs.Argument != nil {
		rs.Argument.Accept(visitor)
	}
}

func (rs *ReturnStatement) statementNode() {}

func (rs *ReturnStatement) String() string {
	if rs.Argument != nil {
		return "return " + rs.Argument.String() + ";"
	}
	return "return;"
}

// BreakStatement break语句
type BreakStatement struct {
	BaseNode
}

func NewBreakStatement(pos lexer.Position) *BreakStatement {
	return &BreakStatement{
		BaseNode: BaseNode{
			Kind:     ASTBreak,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
	}
}

// GetChildren 返回子节点
func (bs *BreakStatement) GetChildren() []Node {
	return nil // break语句是叶子节点
}

// Accept 接受访问者
func (bs *BreakStatement) Accept(visitor Visitor) {
	visitor.Visit(bs)
}

func (bs *BreakStatement) statementNode() {}

func (bs *BreakStatement) String() string {
	return "break;"
}

// ContinueStatement continue语句
type ContinueStatement struct {
	BaseNode
}

func NewContinueStatement(pos lexer.Position) *ContinueStatement {
	return &ContinueStatement{
		BaseNode: BaseNode{
			Kind:     ASTContinue,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
	}
}

// GetChildren 返回子节点
func (cs *ContinueStatement) GetChildren() []Node {
	return nil // continue语句是叶子节点
}

// Accept 接受访问者
func (cs *ContinueStatement) Accept(visitor Visitor) {
	visitor.Visit(cs)
}

func (cs *ContinueStatement) statementNode() {}

func (cs *ContinueStatement) String() string {
	return "continue;"
}

// BlockStatement 块语句
type BlockStatement struct {
	BaseNode
	Body []Statement `json:"body"`
}

func NewBlockStatement(pos lexer.Position) *BlockStatement {
	return &BlockStatement{
		BaseNode: BaseNode{
			Kind:     ASTStmtList,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Body: make([]Statement, 0),
	}
}

// GetChildren 返回子节点
func (bs *BlockStatement) GetChildren() []Node {
	children := make([]Node, len(bs.Body))
	for i, stmt := range bs.Body {
		children[i] = stmt
	}
	return children
}

// Accept 接受访问者
func (bs *BlockStatement) Accept(visitor Visitor) {
	if visitor.Visit(bs) {
		for _, stmt := range bs.Body {
			stmt.Accept(visitor)
		}
	}
}

func (bs *BlockStatement) statementNode() {}

func (bs *BlockStatement) String() string {
	var out strings.Builder
	out.WriteString("{\n")
	for _, stmt := range bs.Body {
		if stmt != nil {
			out.WriteString("  " + stmt.String() + "\n")
		}
	}
	out.WriteString("}")
	return out.String()
}

// GlobalStatement global 语句
type GlobalStatement struct {
	BaseNode
	Variables []Expression `json:"variables"`
}

func NewGlobalStatement(pos lexer.Position) *GlobalStatement {
	return &GlobalStatement{
		BaseNode: BaseNode{
			Kind:     ASTGlobal,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Variables: make([]Expression, 0),
	}
}

// GetChildren 返回子节点
func (gs *GlobalStatement) GetChildren() []Node {
	children := make([]Node, len(gs.Variables))
	for i, variable := range gs.Variables {
		children[i] = variable
	}
	return children
}

// Accept 接受访问者
func (gs *GlobalStatement) Accept(visitor Visitor) {
	if visitor.Visit(gs) {
		for _, variable := range gs.Variables {
			if variable != nil {
				variable.Accept(visitor)
			}
		}
	}
}

func (gs *GlobalStatement) statementNode() {}

func (gs *GlobalStatement) String() string {
	var vars []string
	for _, variable := range gs.Variables {
		if variable != nil {
			vars = append(vars, variable.String())
		}
	}
	return "global " + strings.Join(vars, ", ") + ";"
}

// DoWhileStatement do-while 循环语句
type DoWhileStatement struct {
	BaseNode
	Body      Statement  `json:"body"`
	Condition Expression `json:"condition"`
}

func NewDoWhileStatement(pos lexer.Position, body Statement, condition Expression) *DoWhileStatement {
	return &DoWhileStatement{
		BaseNode: BaseNode{
			Kind:     ASTDoWhile,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Body:      body,
		Condition: condition,
	}
}

func (dw *DoWhileStatement) GetChildren() []Node {
	children := make([]Node, 0, 2)
	if dw.Body != nil {
		children = append(children, dw.Body)
	}
	if dw.Condition != nil {
		children = append(children, dw.Condition)
	}
	return children
}

func (dw *DoWhileStatement) Accept(visitor Visitor) {
	if visitor.Visit(dw) {
		if dw.Body != nil {
			dw.Body.Accept(visitor)
		}
		if dw.Condition != nil {
			dw.Condition.Accept(visitor)
		}
	}
}

func (dw *DoWhileStatement) statementNode() {}

func (dw *DoWhileStatement) String() string {
	body := ""
	if dw.Body != nil {
		body = dw.Body.String()
	}
	condition := ""
	if dw.Condition != nil {
		condition = dw.Condition.String()
	}
	return "do " + body + " while (" + condition + ");"
}

// ForeachStatement foreach 循环语句
type ForeachStatement struct {
	BaseNode
	Iterable Expression `json:"iterable"`
	Key      Expression `json:"key,omitempty"`
	Value    Expression `json:"value"`
	Body     Statement  `json:"body"`
}

func NewForeachStatement(pos lexer.Position, iterable, key, value Expression, body Statement) *ForeachStatement {
	return &ForeachStatement{
		BaseNode: BaseNode{
			Kind:     ASTForeach,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Iterable: iterable,
		Key:      key,
		Value:    value,
		Body:     body,
	}
}

func (f *ForeachStatement) GetChildren() []Node {
	children := make([]Node, 0, 4)
	if f.Iterable != nil {
		children = append(children, f.Iterable)
	}
	if f.Key != nil {
		children = append(children, f.Key)
	}
	if f.Value != nil {
		children = append(children, f.Value)
	}
	if f.Body != nil {
		children = append(children, f.Body)
	}
	return children
}

func (f *ForeachStatement) Accept(visitor Visitor) {
	if visitor.Visit(f) {
		if f.Iterable != nil {
			f.Iterable.Accept(visitor)
		}
		if f.Key != nil {
			f.Key.Accept(visitor)
		}
		if f.Value != nil {
			f.Value.Accept(visitor)
		}
		if f.Body != nil {
			f.Body.Accept(visitor)
		}
	}
}

func (f *ForeachStatement) statementNode() {}

func (f *ForeachStatement) String() string {
	var result strings.Builder
	result.WriteString("foreach (")
	if f.Iterable != nil {
		result.WriteString(f.Iterable.String())
	}
	result.WriteString(" as ")
	if f.Key != nil {
		result.WriteString(f.Key.String())
		result.WriteString(" => ")
	}
	if f.Value != nil {
		result.WriteString(f.Value.String())
	}
	result.WriteString(") ")
	if f.Body != nil {
		result.WriteString(f.Body.String())
	}
	return result.String()
}

// SwitchStatement switch 语句
type SwitchStatement struct {
	BaseNode
	Discriminant Expression      `json:"discriminant"`
	Cases        []*SwitchCase   `json:"cases"`
}

type SwitchCase struct {
	BaseNode
	Test       Expression  `json:"test,omitempty"` // null for default case
	Body       []Statement `json:"body"`
}

func NewSwitchStatement(pos lexer.Position, discriminant Expression) *SwitchStatement {
	return &SwitchStatement{
		BaseNode: BaseNode{
			Kind:     ASTSwitch,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Discriminant: discriminant,
		Cases:        make([]*SwitchCase, 0),
	}
}

func NewSwitchCase(pos lexer.Position, test Expression) *SwitchCase {
	return &SwitchCase{
		BaseNode: BaseNode{
			Kind:     ASTSwitchCase,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Test: test,
		Body: make([]Statement, 0),
	}
}

func (s *SwitchStatement) GetChildren() []Node {
	children := make([]Node, 0, len(s.Cases)+1)
	if s.Discriminant != nil {
		children = append(children, s.Discriminant)
	}
	for _, c := range s.Cases {
		children = append(children, c)
	}
	return children
}

func (s *SwitchStatement) Accept(visitor Visitor) {
	if visitor.Visit(s) {
		if s.Discriminant != nil {
			s.Discriminant.Accept(visitor)
		}
		for _, c := range s.Cases {
			c.Accept(visitor)
		}
	}
}

func (s *SwitchStatement) statementNode() {}

func (s *SwitchStatement) String() string {
	var result strings.Builder
	result.WriteString("switch (")
	if s.Discriminant != nil {
		result.WriteString(s.Discriminant.String())
	}
	result.WriteString(") {\n")
	for _, c := range s.Cases {
		if c != nil {
			result.WriteString(c.String())
		}
	}
	result.WriteString("}")
	return result.String()
}

func (sc *SwitchCase) GetChildren() []Node {
	children := make([]Node, 0, len(sc.Body)+1)
	if sc.Test != nil {
		children = append(children, sc.Test)
	}
	for _, stmt := range sc.Body {
		children = append(children, stmt)
	}
	return children
}

func (sc *SwitchCase) Accept(visitor Visitor) {
	if visitor.Visit(sc) {
		if sc.Test != nil {
			sc.Test.Accept(visitor)
		}
		for _, stmt := range sc.Body {
			if stmt != nil {
				stmt.Accept(visitor)
			}
		}
	}
}

func (sc *SwitchCase) String() string {
	var result strings.Builder
	if sc.Test != nil {
		result.WriteString("  case ")
		result.WriteString(sc.Test.String())
		result.WriteString(":\n")
	} else {
		result.WriteString("  default:\n")
	}
	for _, stmt := range sc.Body {
		if stmt != nil {
			result.WriteString("    ")
			result.WriteString(stmt.String())
			result.WriteString("\n")
		}
	}
	return result.String()
}

// TryStatement try-catch-finally 语句
type TryStatement struct {
	BaseNode
	Body         []Statement   `json:"body"`
	CatchClauses []*CatchClause `json:"catchClauses,omitempty"`
	FinallyBlock []Statement   `json:"finallyBlock,omitempty"`
}

type CatchClause struct {
	BaseNode
	Types     []Expression `json:"types"`     // Exception types
	Parameter Expression   `json:"parameter"` // Exception variable
	Body      []Statement  `json:"body"`
}

func NewTryStatement(pos lexer.Position) *TryStatement {
	return &TryStatement{
		BaseNode: BaseNode{
			Kind:     ASTTry,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Body:         make([]Statement, 0),
		CatchClauses: make([]*CatchClause, 0),
	}
}

func NewCatchClause(pos lexer.Position, parameter Expression) *CatchClause {
	return &CatchClause{
		BaseNode: BaseNode{
			Kind:     ASTCatch,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Types:     make([]Expression, 0),
		Parameter: parameter,
		Body:      make([]Statement, 0),
	}
}

func (t *TryStatement) GetChildren() []Node {
	children := make([]Node, 0)
	for _, stmt := range t.Body {
		children = append(children, stmt)
	}
	for _, catch := range t.CatchClauses {
		children = append(children, catch)
	}
	for _, stmt := range t.FinallyBlock {
		children = append(children, stmt)
	}
	return children
}

func (t *TryStatement) Accept(visitor Visitor) {
	if visitor.Visit(t) {
		for _, stmt := range t.Body {
			if stmt != nil {
				stmt.Accept(visitor)
			}
		}
		for _, catch := range t.CatchClauses {
			if catch != nil {
				catch.Accept(visitor)
			}
		}
		for _, stmt := range t.FinallyBlock {
			if stmt != nil {
				stmt.Accept(visitor)
			}
		}
	}
}

func (t *TryStatement) statementNode() {}

func (t *TryStatement) String() string {
	var result strings.Builder
	result.WriteString("try {\n")
	for _, stmt := range t.Body {
		if stmt != nil {
			result.WriteString("  ")
			result.WriteString(stmt.String())
			result.WriteString("\n")
		}
	}
	result.WriteString("}")
	
	for _, catch := range t.CatchClauses {
		if catch != nil {
			result.WriteString(" ")
			result.WriteString(catch.String())
		}
	}
	
	if len(t.FinallyBlock) > 0 {
		result.WriteString(" finally {\n")
		for _, stmt := range t.FinallyBlock {
			if stmt != nil {
				result.WriteString("  ")
				result.WriteString(stmt.String())
				result.WriteString("\n")
			}
		}
		result.WriteString("}")
	}
	return result.String()
}

func (c *CatchClause) GetChildren() []Node {
	children := make([]Node, 0)
	for _, t := range c.Types {
		children = append(children, t)
	}
	if c.Parameter != nil {
		children = append(children, c.Parameter)
	}
	for _, stmt := range c.Body {
		children = append(children, stmt)
	}
	return children
}

func (c *CatchClause) Accept(visitor Visitor) {
	if visitor.Visit(c) {
		for _, t := range c.Types {
			if t != nil {
				t.Accept(visitor)
			}
		}
		if c.Parameter != nil {
			c.Parameter.Accept(visitor)
		}
		for _, stmt := range c.Body {
			if stmt != nil {
				stmt.Accept(visitor)
			}
		}
	}
}

func (c *CatchClause) String() string {
	var result strings.Builder
	result.WriteString("catch (")
	
	// Types
	var typeStrs []string
	for _, t := range c.Types {
		if t != nil {
			typeStrs = append(typeStrs, t.String())
		}
	}
	result.WriteString(strings.Join(typeStrs, "|"))
	
	if c.Parameter != nil {
		result.WriteString(" ")
		result.WriteString(c.Parameter.String())
	}
	result.WriteString(") {\n")
	
	for _, stmt := range c.Body {
		if stmt != nil {
			result.WriteString("  ")
			result.WriteString(stmt.String())
			result.WriteString("\n")
		}
	}
	result.WriteString("}")
	return result.String()
}

// ThrowStatement throw 语句
type ThrowStatement struct {
	BaseNode
	Argument Expression `json:"argument"`
}

func NewThrowStatement(pos lexer.Position, argument Expression) *ThrowStatement {
	return &ThrowStatement{
		BaseNode: BaseNode{
			Kind:     ASTThrow,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Argument: argument,
	}
}

func (t *ThrowStatement) GetChildren() []Node {
	if t.Argument != nil {
		return []Node{t.Argument}
	}
	return nil
}

func (t *ThrowStatement) Accept(visitor Visitor) {
	if visitor.Visit(t) && t.Argument != nil {
		t.Argument.Accept(visitor)
	}
}

func (t *ThrowStatement) statementNode() {}

func (t *ThrowStatement) String() string {
	arg := ""
	if t.Argument != nil {
		arg = t.Argument.String()
	}
	return "throw " + arg + ";"
}

// StaticStatement static 变量声明语句
type StaticStatement struct {
	BaseNode
	Variables []*StaticVariable `json:"variables"`
}

type StaticVariable struct {
	BaseNode
	Variable     Expression `json:"variable"`
	DefaultValue Expression `json:"defaultValue,omitempty"`
}

func NewStaticStatement(pos lexer.Position) *StaticStatement {
	return &StaticStatement{
		BaseNode: BaseNode{
			Kind:     ASTStatic,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Variables: make([]*StaticVariable, 0),
	}
}

func NewStaticVariable(pos lexer.Position, variable, defaultValue Expression) *StaticVariable {
	return &StaticVariable{
		BaseNode: BaseNode{
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Variable:     variable,
		DefaultValue: defaultValue,
	}
}

func (s *StaticStatement) GetChildren() []Node {
	children := make([]Node, len(s.Variables))
	for i, v := range s.Variables {
		children[i] = v
	}
	return children
}

func (s *StaticStatement) Accept(visitor Visitor) {
	if visitor.Visit(s) {
		for _, v := range s.Variables {
			if v != nil {
				v.Accept(visitor)
			}
		}
	}
}

func (s *StaticStatement) statementNode() {}

func (s *StaticStatement) String() string {
	var vars []string
	for _, v := range s.Variables {
		if v != nil {
			varStr := v.Variable.String()
			if v.DefaultValue != nil {
				varStr += " = " + v.DefaultValue.String()
			}
			vars = append(vars, varStr)
		}
	}
	return "static " + strings.Join(vars, ", ") + ";"
}

func (sv *StaticVariable) GetChildren() []Node {
	children := make([]Node, 0, 2)
	if sv.Variable != nil {
		children = append(children, sv.Variable)
	}
	if sv.DefaultValue != nil {
		children = append(children, sv.DefaultValue)
	}
	return children
}

func (sv *StaticVariable) Accept(visitor Visitor) {
	if visitor.Visit(sv) {
		if sv.Variable != nil {
			sv.Variable.Accept(visitor)
		}
		if sv.DefaultValue != nil {
			sv.DefaultValue.Accept(visitor)
		}
	}
}

// UnsetStatement unset 语句
type UnsetStatement struct {
	BaseNode
	Variables []Expression `json:"variables"`
}

func NewUnsetStatement(pos lexer.Position) *UnsetStatement {
	return &UnsetStatement{
		BaseNode: BaseNode{
			Kind:     ASTUnset,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Variables: make([]Expression, 0),
	}
}

func (u *UnsetStatement) GetChildren() []Node {
	children := make([]Node, len(u.Variables))
	for i, v := range u.Variables {
		children[i] = v
	}
	return children
}

func (u *UnsetStatement) Accept(visitor Visitor) {
	if visitor.Visit(u) {
		for _, v := range u.Variables {
			if v != nil {
				v.Accept(visitor)
			}
		}
	}
}

func (u *UnsetStatement) statementNode() {}

func (u *UnsetStatement) String() string {
	var vars []string
	for _, v := range u.Variables {
		if v != nil {
			vars = append(vars, v.String())
		}
	}
	return "unset(" + strings.Join(vars, ", ") + ");"
}

// GotoStatement goto 语句
type GotoStatement struct {
	BaseNode
	Label Expression `json:"label"`
}

func NewGotoStatement(pos lexer.Position, label Expression) *GotoStatement {
	return &GotoStatement{
		BaseNode: BaseNode{
			Kind:     ASTGoto,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Label: label,
	}
}

func (g *GotoStatement) GetChildren() []Node {
	if g.Label != nil {
		return []Node{g.Label}
	}
	return nil
}

func (g *GotoStatement) Accept(visitor Visitor) {
	if visitor.Visit(g) && g.Label != nil {
		g.Label.Accept(visitor)
	}
}

func (g *GotoStatement) statementNode() {}

func (g *GotoStatement) String() string {
	label := ""
	if g.Label != nil {
		label = g.Label.String()
	}
	return "goto " + label + ";"
}

// LabelStatement 标签语句
type LabelStatement struct {
	BaseNode
	Name Expression `json:"name"`
}

func NewLabelStatement(pos lexer.Position, name Expression) *LabelStatement {
	return &LabelStatement{
		BaseNode: BaseNode{
			Kind:     ASTLabel,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Name: name,
	}
}

func (l *LabelStatement) GetChildren() []Node {
	if l.Name != nil {
		return []Node{l.Name}
	}
	return nil
}

func (l *LabelStatement) Accept(visitor Visitor) {
	if visitor.Visit(l) && l.Name != nil {
		l.Name.Accept(visitor)
	}
}

func (l *LabelStatement) statementNode() {}

func (l *LabelStatement) String() string {
	name := ""
	if l.Name != nil {
		name = l.Name.String()
	}
	return name + ":"
}

// NewExpression new 表达式
type NewExpression struct {
	BaseNode
	Class     Expression   `json:"class"`
	Arguments []Expression `json:"arguments"`
}

func NewNewExpression(pos lexer.Position, class Expression) *NewExpression {
	return &NewExpression{
		BaseNode: BaseNode{
			Kind:     ASTNew,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Class:     class,
		Arguments: make([]Expression, 0),
	}
}

func (n *NewExpression) GetChildren() []Node {
	children := make([]Node, 0, len(n.Arguments)+1)
	if n.Class != nil {
		children = append(children, n.Class)
	}
	for _, arg := range n.Arguments {
		children = append(children, arg)
	}
	return children
}

func (n *NewExpression) Accept(visitor Visitor) {
	if visitor.Visit(n) {
		if n.Class != nil {
			n.Class.Accept(visitor)
		}
		for _, arg := range n.Arguments {
			if arg != nil {
				arg.Accept(visitor)
			}
		}
	}
}

func (n *NewExpression) expressionNode() {}

func (n *NewExpression) String() string {
	class := ""
	if n.Class != nil {
		class = n.Class.String()
	}
	
	var args []string
	for _, arg := range n.Arguments {
		if arg != nil {
			args = append(args, arg.String())
		}
	}
	
	if len(args) > 0 {
		return "new " + class + "(" + strings.Join(args, ", ") + ")"
	}
	return "new " + class
}

// CloneExpression clone 表达式
type CloneExpression struct {
	BaseNode
	Object Expression `json:"object"`
}

func NewCloneExpression(pos lexer.Position, object Expression) *CloneExpression {
	return &CloneExpression{
		BaseNode: BaseNode{
			Kind:     ASTClone,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Object: object,
	}
}

func (c *CloneExpression) GetChildren() []Node {
	if c.Object != nil {
		return []Node{c.Object}
	}
	return nil
}

func (c *CloneExpression) Accept(visitor Visitor) {
	if visitor.Visit(c) && c.Object != nil {
		c.Object.Accept(visitor)
	}
}

func (c *CloneExpression) expressionNode() {}

func (c *CloneExpression) String() string {
	obj := ""
	if c.Object != nil {
		obj = c.Object.String()
	}
	return "clone " + obj
}

// InstanceofExpression instanceof 表达式
type InstanceofExpression struct {
	BaseNode
	Left  Expression `json:"left"`
	Right Expression `json:"right"`
}

func NewInstanceofExpression(pos lexer.Position, left, right Expression) *InstanceofExpression {
	return &InstanceofExpression{
		BaseNode: BaseNode{
			Kind:     ASTInstanceof,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Left:  left,
		Right: right,
	}
}

func (i *InstanceofExpression) GetChildren() []Node {
	children := make([]Node, 0, 2)
	if i.Left != nil {
		children = append(children, i.Left)
	}
	if i.Right != nil {
		children = append(children, i.Right)
	}
	return children
}

func (i *InstanceofExpression) Accept(visitor Visitor) {
	if visitor.Visit(i) {
		if i.Left != nil {
			i.Left.Accept(visitor)
		}
		if i.Right != nil {
			i.Right.Accept(visitor)
		}
	}
}

func (i *InstanceofExpression) expressionNode() {}

func (i *InstanceofExpression) String() string {
	left := ""
	if i.Left != nil {
		left = i.Left.String()
	}
	right := ""
	if i.Right != nil {
		right = i.Right.String()
	}
	return left + " instanceof " + right
}

// MatchExpression 表示 match 表达式 (PHP 8.0+)
type MatchExpression struct {
	BaseNode
	Subject Expression  `json:"subject"` // 匹配的表达式
	Arms    []*MatchArm `json:"arms"`    // match arms
}

// MatchArm 表示 match 表达式的一个分支
type MatchArm struct {
	BaseNode
	Conditions []Expression `json:"conditions,omitempty"` // 条件列表，empty for default
	Body       Expression   `json:"body"`                 // 分支体
	IsDefault  bool         `json:"isDefault"`            // 是否为 default 分支
}

func NewMatchExpression(pos lexer.Position, subject Expression) *MatchExpression {
	return &MatchExpression{
		BaseNode: BaseNode{
			Kind:     ASTMatch,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Subject: subject,
		Arms:    make([]*MatchArm, 0),
	}
}

func NewMatchArm(pos lexer.Position, conditions []Expression, body Expression, isDefault bool) *MatchArm {
	return &MatchArm{
		BaseNode: BaseNode{
			Kind:     ASTMatchArm,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Conditions: conditions,
		Body:       body,
		IsDefault:  isDefault,
	}
}

func (m *MatchExpression) GetChildren() []Node {
	children := make([]Node, 0)
	if m.Subject != nil {
		children = append(children, m.Subject)
	}
	for _, arm := range m.Arms {
		children = append(children, arm)
	}
	return children
}

func (m *MatchExpression) Accept(visitor Visitor) {
	if visitor.Visit(m) {
		if m.Subject != nil {
			m.Subject.Accept(visitor)
		}
		for _, arm := range m.Arms {
			arm.Accept(visitor)
		}
	}
}

func (m *MatchExpression) expressionNode() {}

func (m *MatchExpression) String() string {
	result := "match (" + m.Subject.String() + ") {"
	for i, arm := range m.Arms {
		if i > 0 {
			result += ", "
		}
		result += arm.String()
	}
	result += "}"
	return result
}

func (ma *MatchArm) GetChildren() []Node {
	children := make([]Node, 0)
	for _, condition := range ma.Conditions {
		children = append(children, condition)
	}
	if ma.Body != nil {
		children = append(children, ma.Body)
	}
	return children
}

func (ma *MatchArm) Accept(visitor Visitor) {
	if visitor.Visit(ma) {
		for _, condition := range ma.Conditions {
			condition.Accept(visitor)
		}
		if ma.Body != nil {
			ma.Body.Accept(visitor)
		}
	}
}

func (ma *MatchArm) String() string {
	if ma.IsDefault {
		return "default => " + ma.Body.String()
	}
	
	result := ""
	for i, condition := range ma.Conditions {
		if i > 0 {
			result += ", "
		}
		result += condition.String()
	}
	result += " => " + ma.Body.String()
	return result
}

// NamedArgument 表示命名参数 (PHP 8.0+) - name: value
type NamedArgument struct {
	BaseNode
	Name  *IdentifierNode `json:"name"`  // 参数名
	Value Expression      `json:"value"` // 参数值
}

func NewNamedArgument(pos lexer.Position, name *IdentifierNode, value Expression) *NamedArgument {
	return &NamedArgument{
		BaseNode: BaseNode{
			Kind:     ASTNamedArg,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Name:  name,
		Value: value,
	}
}

func (na *NamedArgument) GetChildren() []Node {
	children := make([]Node, 0)
	if na.Name != nil {
		children = append(children, na.Name)
	}
	if na.Value != nil {
		children = append(children, na.Value)
	}
	return children
}

func (na *NamedArgument) Accept(visitor Visitor) {
	if visitor.Visit(na) {
		if na.Name != nil {
			na.Name.Accept(visitor)
		}
		if na.Value != nil {
			na.Value.Accept(visitor)
		}
	}
}

func (na *NamedArgument) expressionNode() {}

func (na *NamedArgument) String() string {
	return na.Name.String() + ": " + na.Value.String()
}

// Attribute 表示PHP 8.0属性/注解 - #[AttributeName(arguments)]
type Attribute struct {
	BaseNode
	Name      *IdentifierNode `json:"name"`      // 属性名称
	Arguments []Expression    `json:"arguments"` // 属性参数
}

func NewAttribute(pos lexer.Position, name *IdentifierNode, arguments []Expression) *Attribute {
	return &Attribute{
		BaseNode: BaseNode{
			Kind:     ASTAttribute,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Name:      name,
		Arguments: arguments,
	}
}

func (a *Attribute) GetChildren() []Node {
	children := make([]Node, 0)
	if a.Name != nil {
		children = append(children, a.Name)
	}
	for _, arg := range a.Arguments {
		children = append(children, arg)
	}
	return children
}

func (a *Attribute) Accept(visitor Visitor) {
	if visitor.Visit(a) {
		if a.Name != nil {
			a.Name.Accept(visitor)
		}
		for _, arg := range a.Arguments {
			arg.Accept(visitor)
		}
	}
}

func (a *Attribute) expressionNode() {}

func (a *Attribute) String() string {
	result := "#[" + a.Name.String()
	if len(a.Arguments) > 0 {
		result += "("
		for i, arg := range a.Arguments {
			if i > 0 {
				result += ", "
			}
			result += arg.String()
		}
		result += ")"
	}
	result += "]"
	return result
}

// AttributeGroup 表示属性组 - #[Attr1, Attr2, ...]
type AttributeGroup struct {
	BaseNode
	Attributes []*Attribute `json:"attributes"` // 组内的属性列表
}

func NewAttributeGroup(pos lexer.Position, attributes []*Attribute) *AttributeGroup {
	return &AttributeGroup{
		BaseNode: BaseNode{
			Kind:     ASTAttributeGroup,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Attributes: attributes,
	}
}

func (ag *AttributeGroup) GetChildren() []Node {
	children := make([]Node, len(ag.Attributes))
	for i, attr := range ag.Attributes {
		children[i] = attr
	}
	return children
}

func (ag *AttributeGroup) Accept(visitor Visitor) {
	if visitor.Visit(ag) {
		for _, attr := range ag.Attributes {
			attr.Accept(visitor)
		}
	}
}

func (ag *AttributeGroup) expressionNode() {}

func (ag *AttributeGroup) String() string {
	result := "#["
	for i, attr := range ag.Attributes {
		if i > 0 {
			result += ", "
		}
		// 只输出属性名和参数，不包含外层的#[]
		result += attr.Name.String()
		if len(attr.Arguments) > 0 {
			result += "("
			for j, arg := range attr.Arguments {
				if j > 0 {
					result += ", "
				}
				result += arg.String()
			}
			result += ")"
		}
	}
	result += "]"
	return result
}

// AttributeList 表示多个属性组的列表
type AttributeList struct {
	BaseNode
	Groups []*AttributeGroup `json:"groups"` // 属性组列表
}

func NewAttributeList(pos lexer.Position, groups []*AttributeGroup) *AttributeList {
	return &AttributeList{
		BaseNode: BaseNode{
			Kind:     ASTAttributeList,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Groups: groups,
	}
}

func (al *AttributeList) GetChildren() []Node {
	children := make([]Node, len(al.Groups))
	for i, group := range al.Groups {
		children[i] = group
	}
	return children
}

func (al *AttributeList) Accept(visitor Visitor) {
	if visitor.Visit(al) {
		for _, group := range al.Groups {
			group.Accept(visitor)
		}
	}
}

func (al *AttributeList) expressionNode() {}

func (al *AttributeList) String() string {
	result := ""
	for i, group := range al.Groups {
		if i > 0 {
			result += "\n"
		}
		result += group.String()
	}
	return result
}

// CallExpression 函数调用表达式
type CallExpression struct {
	BaseNode
	Callee    Expression   `json:"callee"`
	Arguments []Expression `json:"arguments"`
}

func NewCallExpression(pos lexer.Position, callee Expression) *CallExpression {
	return &CallExpression{
		BaseNode: BaseNode{
			Kind:     ASTCall,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Callee:    callee,
		Arguments: make([]Expression, 0),
	}
}

// GetChildren 返回子节点
func (ce *CallExpression) GetChildren() []Node {
	children := []Node{ce.Callee}
	for _, arg := range ce.Arguments {
		children = append(children, arg)
	}
	return children
}

// Accept 接受访问者
func (ce *CallExpression) Accept(visitor Visitor) {
	if visitor.Visit(ce) {
		ce.Callee.Accept(visitor)
		for _, arg := range ce.Arguments {
			if arg != nil {
				arg.Accept(visitor)
			}
		}
	}
}

func (ce *CallExpression) expressionNode() {}

func (ce *CallExpression) String() string {
	var args []string
	for _, arg := range ce.Arguments {
		if arg != nil {
			args = append(args, arg.String())
		}
	}
	return ce.Callee.String() + "(" + strings.Join(args, ", ") + ")"
}

// DocBlockComment 文档块注释
type DocBlockComment struct {
	BaseNode
	Content string `json:"content"`
	Raw     string `json:"raw"`
}

func NewDocBlockComment(pos lexer.Position, content, raw string) *DocBlockComment {
	return &DocBlockComment{
		BaseNode: BaseNode{
			Kind:     ASTZval, // 文档注释作为特殊值处理
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Content: content,
		Raw:     raw,
	}
}

// GetChildren 返回子节点
func (dbc *DocBlockComment) GetChildren() []Node {
	return nil // 文档注释是叶子节点
}

// Accept 接受访问者
func (dbc *DocBlockComment) Accept(visitor Visitor) {
	visitor.Visit(dbc)
}

func (dbc *DocBlockComment) expressionNode() {}

func (dbc *DocBlockComment) String() string {
	return dbc.Raw
}

// ErrorSuppressionExpression 表示错误抑制操作符 @
type ErrorSuppressionExpression struct {
	BaseNode
	Expression Expression `json:"expression"`
}

func NewErrorSuppressionExpression(pos lexer.Position, expr Expression) *ErrorSuppressionExpression {
	return &ErrorSuppressionExpression{
		BaseNode: BaseNode{
			Kind:     ASTSilence,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Expression: expr,
	}
}

// GetChildren 返回子节点
func (ese *ErrorSuppressionExpression) GetChildren() []Node {
	if ese.Expression == nil {
		return nil
	}
	return []Node{ese.Expression}
}

// Accept 接受访问者
func (ese *ErrorSuppressionExpression) Accept(visitor Visitor) {
	if visitor.Visit(ese) {
		if ese.Expression != nil {
			ese.Expression.Accept(visitor)
		}
	}
}

func (ese *ErrorSuppressionExpression) expressionNode() {}

func (ese *ErrorSuppressionExpression) String() string {
	if ese.Expression == nil {
		return "@"
	}
	return "@" + ese.Expression.String()
}

// CoalesceExpression 表示 null 合并操作符 ??
type CoalesceExpression struct {
	BaseNode
	Left  Expression `json:"left"`
	Right Expression `json:"right"`
}

func NewCoalesceExpression(pos lexer.Position, left Expression, right Expression) *CoalesceExpression {
	return &CoalesceExpression{
		BaseNode: BaseNode{
			Kind:     ASTCoalesce,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Left:  left,
		Right: right,
	}
}

// GetChildren 返回子节点
func (ce *CoalesceExpression) GetChildren() []Node {
	var children []Node
	if ce.Left != nil {
		children = append(children, ce.Left)
	}
	if ce.Right != nil {
		children = append(children, ce.Right)
	}
	return children
}

// Accept 接受访问者
func (ce *CoalesceExpression) Accept(visitor Visitor) {
	if visitor.Visit(ce) {
		if ce.Left != nil {
			ce.Left.Accept(visitor)
		}
		if ce.Right != nil {
			ce.Right.Accept(visitor)
		}
	}
}

func (ce *CoalesceExpression) expressionNode() {}

func (ce *CoalesceExpression) String() string {
	left := ""
	if ce.Left != nil {
		left = ce.Left.String()
	}
	right := ""
	if ce.Right != nil {
		right = ce.Right.String()
	}
	return left + " ?? " + right
}

// ArrayAccessExpression 表示数组访问表达式 []
type ArrayAccessExpression struct {
	BaseNode
	Array Expression  `json:"array"`
	Index *Expression `json:"index,omitempty"` // nil for empty []
}

func NewArrayAccessExpression(pos lexer.Position, array Expression, index *Expression) *ArrayAccessExpression {
	return &ArrayAccessExpression{
		BaseNode: BaseNode{
			Kind:     ASTDim,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Array: array,
		Index: index,
	}
}

// GetChildren 返回子节点
func (aae *ArrayAccessExpression) GetChildren() []Node {
	var children []Node
	if aae.Array != nil {
		children = append(children, aae.Array)
	}
	if aae.Index != nil && *aae.Index != nil {
		children = append(children, *aae.Index)
	}
	return children
}

// Accept 接受访问者
func (aae *ArrayAccessExpression) Accept(visitor Visitor) {
	if visitor.Visit(aae) {
		if aae.Array != nil {
			aae.Array.Accept(visitor)
		}
		if aae.Index != nil && *aae.Index != nil {
			(*aae.Index).Accept(visitor)
		}
	}
}

func (aae *ArrayAccessExpression) expressionNode() {}

func (aae *ArrayAccessExpression) String() string {
	array := ""
	if aae.Array != nil {
		array = aae.Array.String()
	}
	if aae.Index == nil || *aae.Index == nil {
		return array + "[]"
	}
	return array + "[" + (*aae.Index).String() + "]"
}

// EmptyExpression 表示 empty() 函数调用
type EmptyExpression struct {
	BaseNode
	Expression Expression `json:"expression"`
}

func NewEmptyExpression(pos lexer.Position, expr Expression) *EmptyExpression {
	return &EmptyExpression{
		BaseNode: BaseNode{
			Kind:     ASTEmpty,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Expression: expr,
	}
}

// GetChildren 返回子节点
func (ee *EmptyExpression) GetChildren() []Node {
	if ee.Expression == nil {
		return nil
	}
	return []Node{ee.Expression}
}

// Accept 接受访问者
func (ee *EmptyExpression) Accept(visitor Visitor) {
	if visitor.Visit(ee) {
		if ee.Expression != nil {
			ee.Expression.Accept(visitor)
		}
	}
}

func (ee *EmptyExpression) expressionNode() {}

func (ee *EmptyExpression) String() string {
	if ee.Expression == nil {
		return "empty()"
	}
	return "empty(" + ee.Expression.String() + ")"
}

// ============== 新增的AST节点类型 ==============

// ExitExpression 表示 exit/die 表达式
type ExitExpression struct {
	BaseNode
	Argument Expression `json:"argument,omitempty"`
}

func NewExitExpression(pos lexer.Position, argument Expression) *ExitExpression {
	return &ExitExpression{
		BaseNode: BaseNode{
			Kind:     ASTExit,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Argument: argument,
	}
}

func (ee *ExitExpression) GetChildren() []Node {
	if ee.Argument == nil {
		return nil
	}
	return []Node{ee.Argument}
}

func (ee *ExitExpression) Accept(visitor Visitor) {
	if visitor.Visit(ee) {
		if ee.Argument != nil {
			ee.Argument.Accept(visitor)
		}
	}
}

func (ee *ExitExpression) expressionNode() {}

func (ee *ExitExpression) String() string {
	if ee.Argument == nil {
		return "exit"
	}
	return "exit(" + ee.Argument.String() + ")"
}

// IssetExpression 表示 isset() 表达式
type IssetExpression struct {
	BaseNode
	Arguments []Expression `json:"arguments"`
}

func NewIssetExpression(pos lexer.Position, arguments []Expression) *IssetExpression {
	return &IssetExpression{
		BaseNode: BaseNode{
			Kind:     ASTIsset,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Arguments: arguments,
	}
}

func (ie *IssetExpression) GetChildren() []Node {
	children := make([]Node, len(ie.Arguments))
	for i, arg := range ie.Arguments {
		children[i] = arg
	}
	return children
}

func (ie *IssetExpression) Accept(visitor Visitor) {
	if visitor.Visit(ie) {
		for _, arg := range ie.Arguments {
			if arg != nil {
				arg.Accept(visitor)
			}
		}
	}
}

func (ie *IssetExpression) expressionNode() {}

func (ie *IssetExpression) String() string {
	args := make([]string, len(ie.Arguments))
	for i, arg := range ie.Arguments {
		if arg != nil {
			args[i] = arg.String()
		}
	}
	return "isset(" + strings.Join(args, ", ") + ")"
}

// ListExpression 表示 list() 表达式
type ListExpression struct {
	BaseNode
	Elements []Expression `json:"elements"`
}

func NewListExpression(pos lexer.Position, elements []Expression) *ListExpression {
	return &ListExpression{
		BaseNode: BaseNode{
			Kind:     ASTList,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Elements: elements,
	}
}

func (le *ListExpression) GetChildren() []Node {
	children := make([]Node, 0, len(le.Elements))
	for _, elem := range le.Elements {
		if elem != nil {
			children = append(children, elem)
		}
	}
	return children
}

func (le *ListExpression) Accept(visitor Visitor) {
	if visitor.Visit(le) {
		for _, elem := range le.Elements {
			if elem != nil {
				elem.Accept(visitor)
			}
		}
	}
}

func (le *ListExpression) expressionNode() {}

func (le *ListExpression) String() string {
	args := make([]string, len(le.Elements))
	for i, elem := range le.Elements {
		if elem != nil {
			args[i] = elem.String()
		} else {
			args[i] = ""
		}
	}
	return "list(" + strings.Join(args, ", ") + ")"
}

// AnonymousFunctionExpression 表示匿名函数表达式
type AnonymousFunctionExpression struct {
	BaseNode
	Parameters []Parameter  `json:"parameters"`
	Body       []Statement  `json:"body"`
	UseClause  []Expression `json:"useClause,omitempty"`
}

func NewAnonymousFunctionExpression(pos lexer.Position, parameters []Parameter, body []Statement, useClause []Expression) *AnonymousFunctionExpression {
	return &AnonymousFunctionExpression{
		BaseNode: BaseNode{
			Kind:     ASTAnonymousFunction,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Parameters: parameters,
		Body:       body,
		UseClause:  useClause,
	}
}

func (afe *AnonymousFunctionExpression) GetChildren() []Node {
	children := make([]Node, 0, len(afe.Body)+len(afe.UseClause))
	
	for _, stmt := range afe.Body {
		if stmt != nil {
			children = append(children, stmt)
		}
	}
	
	for _, use := range afe.UseClause {
		if use != nil {
			children = append(children, use)
		}
	}
	
	return children
}

func (afe *AnonymousFunctionExpression) Accept(visitor Visitor) {
	if visitor.Visit(afe) {
		for _, stmt := range afe.Body {
			if stmt != nil {
				stmt.Accept(visitor)
			}
		}
		for _, use := range afe.UseClause {
			if use != nil {
				use.Accept(visitor)
			}
		}
	}
}

func (afe *AnonymousFunctionExpression) expressionNode() {}

func (afe *AnonymousFunctionExpression) String() string {
	return "function"
}

// TernaryExpression 表示三元运算符表达式
type TernaryExpression struct {
	BaseNode
	Test       Expression `json:"test"`
	Consequent Expression `json:"consequent,omitempty"` // 短三元运算符时为 nil
	Alternate  Expression `json:"alternate"`
}

func NewTernaryExpression(pos lexer.Position, test, consequent, alternate Expression) *TernaryExpression {
	return &TernaryExpression{
		BaseNode: BaseNode{
			Kind:     ASTConditional,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Test:       test,
		Consequent: consequent,
		Alternate:  alternate,
	}
}

func (te *TernaryExpression) GetChildren() []Node {
	children := make([]Node, 0, 3)
	
	if te.Test != nil {
		children = append(children, te.Test)
	}
	if te.Consequent != nil {
		children = append(children, te.Consequent)
	}
	if te.Alternate != nil {
		children = append(children, te.Alternate)
	}
	
	return children
}

func (te *TernaryExpression) Accept(visitor Visitor) {
	if visitor.Visit(te) {
		if te.Test != nil {
			te.Test.Accept(visitor)
		}
		if te.Consequent != nil {
			te.Consequent.Accept(visitor)
		}
		if te.Alternate != nil {
			te.Alternate.Accept(visitor)
		}
	}
}

func (te *TernaryExpression) expressionNode() {}

func (te *TernaryExpression) String() string {
	if te.Consequent == nil {
		// 短三元运算符 ?:
		return te.Test.String() + " ?: " + te.Alternate.String()
	}
	return te.Test.String() + " ? " + te.Consequent.String() + " : " + te.Alternate.String()
}

// ArrayElementExpression 表示数组元素表达式 (key => value)
type ArrayElementExpression struct {
	BaseNode
	Key   Expression `json:"key"`
	Value Expression `json:"value"`
}

func NewArrayElementExpression(pos lexer.Position, key, value Expression) *ArrayElementExpression {
	return &ArrayElementExpression{
		BaseNode: BaseNode{
			Kind:     ASTArrayElem,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Key:   key,
		Value: value,
	}
}

func (aee *ArrayElementExpression) GetChildren() []Node {
	children := make([]Node, 0, 2)
	if aee.Key != nil {
		children = append(children, aee.Key)
	}
	if aee.Value != nil {
		children = append(children, aee.Value)
	}
	return children
}

func (aee *ArrayElementExpression) Accept(visitor Visitor) {
	if visitor.Visit(aee) {
		if aee.Key != nil {
			aee.Key.Accept(visitor)
		}
		if aee.Value != nil {
			aee.Value.Accept(visitor)
		}
	}
}

func (aee *ArrayElementExpression) expressionNode() {}

func (aee *ArrayElementExpression) String() string {
	if aee.Key == nil {
		return aee.Value.String()
	}
	return aee.Key.String() + " => " + aee.Value.String()
}

// CaseExpression 表示 switch 语句中的 case 表达式
type CaseExpression struct {
	BaseNode
	Test Expression `json:"test"` // nil for default case
}

func NewCaseExpression(pos lexer.Position, test Expression) *CaseExpression {
	return &CaseExpression{
		BaseNode: BaseNode{
			Kind:     ASTCase,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Test: test,
	}
}

func (ce *CaseExpression) GetChildren() []Node {
	var children []Node
	if ce.Test != nil {
		children = append(children, ce.Test)
	}
	return children
}

func (ce *CaseExpression) Accept(visitor Visitor) {
	if visitor.Visit(ce) {
		if ce.Test != nil {
			ce.Test.Accept(visitor)
		}
	}
}

func (ce *CaseExpression) expressionNode() {}

func (ce *CaseExpression) String() string {
	if ce.Test == nil {
		return "default"
	}
	return "case " + ce.Test.String()
}

// ClassExpression 表示类声明表达式
type ClassExpression struct {
	BaseNode
	Name       Expression   `json:"name"`
	ReadOnly   bool         `json:"readOnly,omitempty"`   // readonly class
	Extends    Expression   `json:"extends"`
	Implements []Expression `json:"implements"`
	Body       []Statement  `json:"body"`
}

func NewClassExpression(pos lexer.Position, name, extends Expression, implements []Expression, readOnly bool) *ClassExpression {
	return &ClassExpression{
		BaseNode: BaseNode{
			Kind:     ASTClass,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Name:       name,
		ReadOnly:   readOnly,
		Extends:    extends,
		Implements: implements,
		Body:       make([]Statement, 0),
	}
}

func (ce *ClassExpression) GetChildren() []Node {
	var children []Node
	if ce.Name != nil {
		children = append(children, ce.Name)
	}
	if ce.Extends != nil {
		children = append(children, ce.Extends)
	}
	for _, impl := range ce.Implements {
		if impl != nil {
			children = append(children, impl)
		}
	}
	for _, stmt := range ce.Body {
		if stmt != nil {
			children = append(children, stmt)
		}
	}
	return children
}

func (ce *ClassExpression) Accept(visitor Visitor) {
	if visitor.Visit(ce) {
		if ce.Name != nil {
			ce.Name.Accept(visitor)
		}
		if ce.Extends != nil {
			ce.Extends.Accept(visitor)
		}
		for _, impl := range ce.Implements {
			if impl != nil {
				impl.Accept(visitor)
			}
		}
		for _, stmt := range ce.Body {
			if stmt != nil {
				stmt.Accept(visitor)
			}
		}
	}
}

func (ce *ClassExpression) expressionNode() {}

func (ce *ClassExpression) String() string {
	result := ""
	if ce.ReadOnly {
		result = "readonly "
	}
	result += "class"
	if ce.Name != nil {
		result += " " + ce.Name.String()
	}
	if ce.Extends != nil {
		result += " extends " + ce.Extends.String()
	}
	return result
}

// ConstExpression 表示常量声明表达式
type ConstExpression struct {
	BaseNode
	Name  Expression `json:"name"`
	Value Expression `json:"value"`
}

func NewConstExpression(pos lexer.Position, name, value Expression) *ConstExpression {
	return &ConstExpression{
		BaseNode: BaseNode{
			Kind:     ASTConst,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Name:  name,
		Value: value,
	}
}

func (ce *ConstExpression) GetChildren() []Node {
	var children []Node
	if ce.Name != nil {
		children = append(children, ce.Name)
	}
	if ce.Value != nil {
		children = append(children, ce.Value)
	}
	return children
}

func (ce *ConstExpression) Accept(visitor Visitor) {
	if visitor.Visit(ce) {
		if ce.Name != nil {
			ce.Name.Accept(visitor)
		}
		if ce.Value != nil {
			ce.Value.Accept(visitor)
		}
	}
}

func (ce *ConstExpression) expressionNode() {}

func (ce *ConstExpression) String() string {
	result := "const"
	if ce.Name != nil {
		result += " " + ce.Name.String()
	}
	if ce.Value != nil {
		result += " = " + ce.Value.String()
	}
	return result
}

// ClassConstantDeclaration 类常量声明语句
type ClassConstantDeclaration struct {
	BaseNode
	Visibility string              `json:"visibility"`   // private, protected, public
	Constants  []ConstantDeclarator `json:"constants"`    // 支持一行声明多个常量
}

// ConstantDeclarator 单个常量声明
type ConstantDeclarator struct {
	BaseNode
	Name  Expression `json:"name"`
	Value Expression `json:"value"`
}

func NewConstantDeclarator(pos lexer.Position, name, value Expression) *ConstantDeclarator {
	return &ConstantDeclarator{
		BaseNode: BaseNode{
			Kind:     ASTConstElem,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Name:  name,
		Value: value,
	}
}

func (cd *ConstantDeclarator) GetChildren() []Node {
	var children []Node
	if cd.Name != nil {
		children = append(children, cd.Name)
	}
	if cd.Value != nil {
		children = append(children, cd.Value)
	}
	return children
}

func (cd *ConstantDeclarator) Accept(visitor Visitor) {
	if visitor.Visit(cd) {
		if cd.Name != nil {
			cd.Name.Accept(visitor)
		}
		if cd.Value != nil {
			cd.Value.Accept(visitor)
		}
	}
}

func (cd *ConstantDeclarator) String() string {
	result := ""
	if cd.Name != nil {
		result += cd.Name.String()
	}
	if cd.Value != nil {
		result += " = " + cd.Value.String()
	}
	return result
}

func NewClassConstantDeclaration(pos lexer.Position, visibility string, constants []ConstantDeclarator) *ClassConstantDeclaration {
	return &ClassConstantDeclaration{
		BaseNode: BaseNode{
			Kind:     ASTClassConstGroup,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Visibility: visibility,
		Constants:  constants,
	}
}

func (ccd *ClassConstantDeclaration) GetChildren() []Node {
	var children []Node
	for i := range ccd.Constants {
		children = append(children, &ccd.Constants[i])
	}
	return children
}

func (ccd *ClassConstantDeclaration) Accept(visitor Visitor) {
	if visitor.Visit(ccd) {
		for i := range ccd.Constants {
			ccd.Constants[i].Accept(visitor)
		}
	}
}

func (ccd *ClassConstantDeclaration) statementNode() {}

func (ccd *ClassConstantDeclaration) String() string {
	result := ccd.Visibility + " const "
	for i, constant := range ccd.Constants {
		if i > 0 {
			result += ", "
		}
		result += constant.String()
	}
	return result
}

// EvalExpression 表示 eval() 表达式
type EvalExpression struct {
	BaseNode
	Argument Expression `json:"argument"`
}

func NewEvalExpression(pos lexer.Position, argument Expression) *EvalExpression {
	return &EvalExpression{
		BaseNode: BaseNode{
			Kind:     ASTEvalExpression,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Argument: argument,
	}
}

func (ee *EvalExpression) GetChildren() []Node {
	var children []Node
	if ee.Argument != nil {
		children = append(children, ee.Argument)
	}
	return children
}

func (ee *EvalExpression) Accept(visitor Visitor) {
	if visitor.Visit(ee) {
		if ee.Argument != nil {
			ee.Argument.Accept(visitor)
		}
	}
}

func (ee *EvalExpression) expressionNode() {}

func (ee *EvalExpression) String() string {
	if ee.Argument != nil {
		return "eval(" + ee.Argument.String() + ")"
	}
	return "eval()"
}

// StaticAccessExpression 表示静态访问表达式 Class::method 或 Class::$property  
type StaticAccessExpression struct {
	BaseNode
	Class    Expression `json:"class"`
	Property Expression `json:"property"`
}

func NewStaticAccessExpression(pos lexer.Position, class, property Expression) *StaticAccessExpression {
	return &StaticAccessExpression{
		BaseNode: BaseNode{
			Kind:     ASTStaticCall,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Class:    class,
		Property: property,
	}
}

func (sae *StaticAccessExpression) GetChildren() []Node {
	var children []Node
	if sae.Class != nil {
		children = append(children, sae.Class)
	}
	if sae.Property != nil {
		children = append(children, sae.Property)
	}
	return children
}

func (sae *StaticAccessExpression) Accept(visitor Visitor) {
	if visitor.Visit(sae) {
		if sae.Class != nil {
			sae.Class.Accept(visitor)
		}
		if sae.Property != nil {
			sae.Property.Accept(visitor)
		}
	}
}

func (sae *StaticAccessExpression) expressionNode() {}

func (sae *StaticAccessExpression) String() string {
	var class string
	if sae.Class != nil {
		class = sae.Class.String()
	}
	
	var property string
	if sae.Property != nil {
		property = sae.Property.String()
	}
	
	return class + "::" + property
}

// VisibilityModifierExpression 表示可见性修饰符表达式
type VisibilityModifierExpression struct {
	BaseNode
	Modifier string `json:"modifier"` // public, private, protected
}

func NewVisibilityModifierExpression(pos lexer.Position, modifier string) *VisibilityModifierExpression {
	return &VisibilityModifierExpression{
		BaseNode: BaseNode{
			Kind:     ASTVisibilityModifier,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Modifier: modifier,
	}
}

func (vme *VisibilityModifierExpression) GetChildren() []Node {
	return []Node{} // No children for simple modifier
}

func (vme *VisibilityModifierExpression) Accept(visitor Visitor) {
	visitor.Visit(vme)
}

func (vme *VisibilityModifierExpression) expressionNode() {}

func (vme *VisibilityModifierExpression) String() string {
	return vme.Modifier
}

// IncludeOrEvalExpression 表示 include/require/eval 表达式
type IncludeOrEvalExpression struct {
	BaseNode
	Type lexer.TokenType // T_INCLUDE, T_INCLUDE_ONCE, T_REQUIRE, T_REQUIRE_ONCE, T_EVAL
	Expr Node            // 要包含的文件表达式
}

func NewIncludeOrEvalExpression(pos lexer.Position, tokenType lexer.TokenType, expr Node) *IncludeOrEvalExpression {
	return &IncludeOrEvalExpression{
		BaseNode: BaseNode{
			Kind:     ASTIncludeOrEval,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Type: tokenType,
		Expr: expr,
	}
}

func (ie *IncludeOrEvalExpression) GetChildren() []Node {
	if ie.Expr != nil {
		return []Node{ie.Expr}
	}
	return []Node{}
}

func (ie *IncludeOrEvalExpression) Accept(visitor Visitor) {
	if visitor.Visit(ie) && ie.Expr != nil {
		ie.Expr.Accept(visitor)
	}
}

func (ie *IncludeOrEvalExpression) expressionNode() {}

func (ie *IncludeOrEvalExpression) String() string {
	switch ie.Type {
	case lexer.T_INCLUDE:
		return "include " + ie.Expr.String()
	case lexer.T_INCLUDE_ONCE:
		return "include_once " + ie.Expr.String()
	case lexer.T_REQUIRE:
		return "require " + ie.Expr.String()
	case lexer.T_REQUIRE_ONCE:
		return "require_once " + ie.Expr.String()
	case lexer.T_EVAL:
		return "eval(" + ie.Expr.String() + ")"
	default:
		return "include_or_eval " + ie.Expr.String()
	}
}


// CloseTagExpression 表示 PHP 结束标签
type CloseTagExpression struct {
	BaseNode
	Content string // ?> 后面的内容（如果有的话）
}

func NewCloseTagExpression(pos lexer.Position, content string) *CloseTagExpression {
	return &CloseTagExpression{
		BaseNode: BaseNode{
			Kind:     ASTZval, // 使用 Zval 类型表示简单值
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Content: content,
	}
}

func (ct *CloseTagExpression) GetChildren() []Node {
	return []Node{}
}

func (ct *CloseTagExpression) Accept(visitor Visitor) {
	visitor.Visit(ct)
}

func (ct *CloseTagExpression) expressionNode() {}

func (ct *CloseTagExpression) String() string {
	if ct.Content != "" {
		return "?>" + ct.Content
	}
	return "?>"
}

// NamespaceExpression 表示命名空间表达式（以 \ 开始）
type NamespaceExpression struct {
	BaseNode
	Name Node // 命名空间名称
}

func NewNamespaceExpression(pos lexer.Position, name Node) *NamespaceExpression {
	return &NamespaceExpression{
		BaseNode: BaseNode{
			Kind:     ASTNamespace,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Name: name,
	}
}

func (ne *NamespaceExpression) GetChildren() []Node {
	if ne.Name != nil {
		return []Node{ne.Name}
	}
	return []Node{}
}

func (ne *NamespaceExpression) Accept(visitor Visitor) {
	if visitor.Visit(ne) && ne.Name != nil {
		ne.Name.Accept(visitor)
	}
}

func (ne *NamespaceExpression) expressionNode() {}

func (ne *NamespaceExpression) String() string {
	if ne.Name != nil {
		return "\\" + ne.Name.String()
	}
	return "\\"
}

// AlternativeIfStatement 表示替代语法的if语句 (if: ... endif;)
type AlternativeIfStatement struct {
	BaseNode
	Condition Expression   `json:"condition"`
	Then      []Statement  `json:"then"`
	ElseIfs   []*ElseIfClause `json:"elseIfs,omitempty"`
	Else      []Statement  `json:"else,omitempty"`
}

// ElseIfClause 表示elseif子句
type ElseIfClause struct {
	BaseNode
	Condition Expression  `json:"condition"`
	Body      []Statement `json:"body"`
}

func NewAlternativeIfStatement(pos lexer.Position, condition Expression) *AlternativeIfStatement {
	return &AlternativeIfStatement{
		BaseNode: BaseNode{
			Kind:     ASTAltIf,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Condition: condition,
		Then:      make([]Statement, 0),
		ElseIfs:   make([]*ElseIfClause, 0),
		Else:      make([]Statement, 0),
	}
}

func NewElseIfClause(pos lexer.Position, condition Expression) *ElseIfClause {
	return &ElseIfClause{
		BaseNode: BaseNode{
			Kind:     ASTElseIf,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Condition: condition,
		Body:      make([]Statement, 0),
	}
}

func (ais *AlternativeIfStatement) GetChildren() []Node {
	var children []Node
	if ais.Condition != nil {
		children = append(children, ais.Condition)
	}
	for _, stmt := range ais.Then {
		if stmt != nil {
			children = append(children, stmt)
		}
	}
	for _, elseif := range ais.ElseIfs {
		if elseif != nil {
			children = append(children, elseif)
		}
	}
	for _, stmt := range ais.Else {
		if stmt != nil {
			children = append(children, stmt)
		}
	}
	return children
}

func (ais *AlternativeIfStatement) Accept(visitor Visitor) {
	if visitor.Visit(ais) {
		if ais.Condition != nil {
			ais.Condition.Accept(visitor)
		}
		for _, stmt := range ais.Then {
			if stmt != nil {
				stmt.Accept(visitor)
			}
		}
		for _, elseif := range ais.ElseIfs {
			if elseif != nil {
				elseif.Accept(visitor)
			}
		}
		for _, stmt := range ais.Else {
			if stmt != nil {
				stmt.Accept(visitor)
			}
		}
	}
}

func (ais *AlternativeIfStatement) statementNode() {}

func (ais *AlternativeIfStatement) String() string {
	result := "if (" + ais.Condition.String() + "):"
	for _, stmt := range ais.Then {
		result += " " + stmt.String()
	}
	for _, elseif := range ais.ElseIfs {
		result += " elseif (" + elseif.Condition.String() + "):"
		for _, stmt := range elseif.Body {
			result += " " + stmt.String()
		}
	}
	if len(ais.Else) > 0 {
		result += " else:"
		for _, stmt := range ais.Else {
			result += " " + stmt.String()
		}
	}
	result += " endif;"
	return result
}

func (eic *ElseIfClause) GetChildren() []Node {
	var children []Node
	if eic.Condition != nil {
		children = append(children, eic.Condition)
	}
	for _, stmt := range eic.Body {
		if stmt != nil {
			children = append(children, stmt)
		}
	}
	return children
}

func (eic *ElseIfClause) Accept(visitor Visitor) {
	if visitor.Visit(eic) {
		if eic.Condition != nil {
			eic.Condition.Accept(visitor)
		}
		for _, stmt := range eic.Body {
			if stmt != nil {
				stmt.Accept(visitor)
			}
		}
	}
}

func (eic *ElseIfClause) statementNode() {}

func (eic *ElseIfClause) String() string {
	result := "elseif (" + eic.Condition.String() + "):"
	for _, stmt := range eic.Body {
		result += " " + stmt.String()
	}
	return result
}

// AlternativeWhileStatement 表示替代语法的while语句 (while: ... endwhile;)
type AlternativeWhileStatement struct {
	BaseNode
	Condition Expression  `json:"condition"`
	Body      []Statement `json:"body"`
}

func NewAlternativeWhileStatement(pos lexer.Position, condition Expression) *AlternativeWhileStatement {
	return &AlternativeWhileStatement{
		BaseNode: BaseNode{
			Kind:     ASTAltWhile,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Condition: condition,
		Body:      make([]Statement, 0),
	}
}

func (aws *AlternativeWhileStatement) GetChildren() []Node {
	var children []Node
	if aws.Condition != nil {
		children = append(children, aws.Condition)
	}
	for _, stmt := range aws.Body {
		if stmt != nil {
			children = append(children, stmt)
		}
	}
	return children
}

func (aws *AlternativeWhileStatement) Accept(visitor Visitor) {
	if visitor.Visit(aws) {
		if aws.Condition != nil {
			aws.Condition.Accept(visitor)
		}
		for _, stmt := range aws.Body {
			if stmt != nil {
				stmt.Accept(visitor)
			}
		}
	}
}

func (aws *AlternativeWhileStatement) statementNode() {}

func (aws *AlternativeWhileStatement) String() string {
	result := "while (" + aws.Condition.String() + "):"
	for _, stmt := range aws.Body {
		result += " " + stmt.String()
	}
	result += " endwhile;"
	return result
}

// AlternativeForStatement 表示替代语法的for语句 (for: ... endfor;)
type AlternativeForStatement struct {
	BaseNode
	Init      []Expression `json:"init,omitempty"`
	Condition []Expression `json:"condition,omitempty"`
	Update    []Expression `json:"update,omitempty"`
	Body      []Statement  `json:"body"`
}

func NewAlternativeForStatement(pos lexer.Position) *AlternativeForStatement {
	return &AlternativeForStatement{
		BaseNode: BaseNode{
			Kind:     ASTAltFor,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Init:      make([]Expression, 0),
		Condition: make([]Expression, 0),
		Update:    make([]Expression, 0),
		Body:      make([]Statement, 0),
	}
}

func (afs *AlternativeForStatement) GetChildren() []Node {
	var children []Node
	for _, expr := range afs.Init {
		if expr != nil {
			children = append(children, expr)
		}
	}
	for _, expr := range afs.Condition {
		if expr != nil {
			children = append(children, expr)
		}
	}
	for _, expr := range afs.Update {
		if expr != nil {
			children = append(children, expr)
		}
	}
	for _, stmt := range afs.Body {
		if stmt != nil {
			children = append(children, stmt)
		}
	}
	return children
}

func (afs *AlternativeForStatement) Accept(visitor Visitor) {
	if visitor.Visit(afs) {
		for _, expr := range afs.Init {
			if expr != nil {
				expr.Accept(visitor)
			}
		}
		for _, expr := range afs.Condition {
			if expr != nil {
				expr.Accept(visitor)
			}
		}
		for _, expr := range afs.Update {
			if expr != nil {
				expr.Accept(visitor)
			}
		}
		for _, stmt := range afs.Body {
			if stmt != nil {
				stmt.Accept(visitor)
			}
		}
	}
}

func (afs *AlternativeForStatement) statementNode() {}

func (afs *AlternativeForStatement) String() string {
	result := "for ("
	for i, expr := range afs.Init {
		if i > 0 {
			result += ", "
		}
		result += expr.String()
	}
	result += "; "
	for i, expr := range afs.Condition {
		if i > 0 {
			result += ", "
		}
		result += expr.String()
	}
	result += "; "
	for i, expr := range afs.Update {
		if i > 0 {
			result += ", "
		}
		result += expr.String()
	}
	result += "):"
	for _, stmt := range afs.Body {
		result += " " + stmt.String()
	}
	result += " endfor;"
	return result
}

// AlternativeForeachStatement 表示替代语法的foreach语句 (foreach: ... endforeach;)
type AlternativeForeachStatement struct {
	BaseNode
	Iterable Expression  `json:"iterable"`
	Key      Expression  `json:"key,omitempty"`
	Value    Expression  `json:"value"`
	Body     []Statement `json:"body"`
}

func NewAlternativeForeachStatement(pos lexer.Position, iterable, value Expression) *AlternativeForeachStatement {
	return &AlternativeForeachStatement{
		BaseNode: BaseNode{
			Kind:     ASTAltForeach,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Iterable: iterable,
		Value:    value,
		Body:     make([]Statement, 0),
	}
}

func (afs *AlternativeForeachStatement) GetChildren() []Node {
	var children []Node
	if afs.Iterable != nil {
		children = append(children, afs.Iterable)
	}
	if afs.Key != nil {
		children = append(children, afs.Key)
	}
	if afs.Value != nil {
		children = append(children, afs.Value)
	}
	for _, stmt := range afs.Body {
		if stmt != nil {
			children = append(children, stmt)
		}
	}
	return children
}

func (afs *AlternativeForeachStatement) Accept(visitor Visitor) {
	if visitor.Visit(afs) {
		if afs.Iterable != nil {
			afs.Iterable.Accept(visitor)
		}
		if afs.Key != nil {
			afs.Key.Accept(visitor)
		}
		if afs.Value != nil {
			afs.Value.Accept(visitor)
		}
		for _, stmt := range afs.Body {
			if stmt != nil {
				stmt.Accept(visitor)
			}
		}
	}
}

func (afs *AlternativeForeachStatement) statementNode() {}

func (afs *AlternativeForeachStatement) String() string {
	result := "foreach (" + afs.Iterable.String() + " as "
	if afs.Key != nil {
		result += afs.Key.String() + " => "
	}
	result += afs.Value.String() + "):"
	for _, stmt := range afs.Body {
		result += " " + stmt.String()
	}
	result += " endforeach;"
	return result
}

// DeclareStatement 表示declare语句
type DeclareStatement struct {
	BaseNode
	Declarations []Expression `json:"declarations"`
	Body         []Statement  `json:"body,omitempty"`
	Alternative  bool         `json:"alternative"` // true for declare(): ... enddeclare;
}

func NewDeclareStatement(pos lexer.Position, declarations []Expression, alternative bool) *DeclareStatement {
	return &DeclareStatement{
		BaseNode: BaseNode{
			Kind:     ASTDeclare,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Declarations: declarations,
		Body:         make([]Statement, 0),
		Alternative:  alternative,
	}
}

func (ds *DeclareStatement) GetChildren() []Node {
	var children []Node
	for _, decl := range ds.Declarations {
		if decl != nil {
			children = append(children, decl)
		}
	}
	for _, stmt := range ds.Body {
		if stmt != nil {
			children = append(children, stmt)
		}
	}
	return children
}

func (ds *DeclareStatement) Accept(visitor Visitor) {
	if visitor.Visit(ds) {
		for _, decl := range ds.Declarations {
			if decl != nil {
				decl.Accept(visitor)
			}
		}
		for _, stmt := range ds.Body {
			if stmt != nil {
				stmt.Accept(visitor)
			}
		}
	}
}

func (ds *DeclareStatement) statementNode() {}

func (ds *DeclareStatement) String() string {
	result := "declare ("
	for i, decl := range ds.Declarations {
		if i > 0 {
			result += ", "
		}
		result += decl.String()
	}
	result += ")"
	if ds.Alternative {
		result += ":"
		for _, stmt := range ds.Body {
			result += " " + stmt.String()
		}
		result += " enddeclare;"
	} else if len(ds.Body) > 0 {
		result += " {"
		for _, stmt := range ds.Body {
			result += " " + stmt.String()
		}
		result += " }"
	} else {
		result += ";"
	}
	return result
}

// FirstClassCallable 表示第一类可调用语法 function(...), $obj->method(...), Class::method(...)
type FirstClassCallable struct {
	BaseNode
	Callable Expression `json:"callable"` // 被调用的函数/方法/静态方法
}

func NewFirstClassCallable(pos lexer.Position, callable Expression) *FirstClassCallable {
	return &FirstClassCallable{
		BaseNode: BaseNode{
			Kind:     ASTFirstClassCallable,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Callable: callable,
	}
}

func (fcc *FirstClassCallable) GetChildren() []Node {
	var children []Node
	if fcc.Callable != nil {
		children = append(children, fcc.Callable)
	}
	return children
}

func (fcc *FirstClassCallable) Accept(visitor Visitor) {
	if fcc.Callable != nil {
		fcc.Callable.Accept(visitor)
	}
}

func (fcc *FirstClassCallable) expressionNode() {}

func (fcc *FirstClassCallable) String() string {
	if fcc.Callable != nil {
		return fcc.Callable.String() + "(...)"
	}
	return "(...)"
}

// AnonymousClass 表示匿名类表达式
type AnonymousClass struct {
	BaseNode
	Arguments  []Expression `json:"arguments,omitempty"`  // 构造函数参数
	Extends    Expression   `json:"extends,omitempty"`    // 继承的类
	Implements []Expression `json:"implements,omitempty"` // 实现的接口
	Body       []Statement  `json:"body"`                 // 类体
}

func NewAnonymousClass(pos lexer.Position, args []Expression, extends Expression, implements []Expression, body []Statement) *AnonymousClass {
	return &AnonymousClass{
		BaseNode: BaseNode{
			Kind:     ASTAnonymousClass,
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Arguments:  args,
		Extends:    extends,
		Implements: implements,
		Body:       body,
	}
}

// GetChildren 返回子节点
func (ac *AnonymousClass) GetChildren() []Node {
	children := make([]Node, 0)
	for _, arg := range ac.Arguments {
		children = append(children, arg)
	}
	if ac.Extends != nil {
		children = append(children, ac.Extends)
	}
	for _, impl := range ac.Implements {
		children = append(children, impl)
	}
	for _, stmt := range ac.Body {
		children = append(children, stmt)
	}
	return children
}

// Accept 接受访问者
func (ac *AnonymousClass) Accept(visitor Visitor) {
	if visitor.Visit(ac) {
		for _, arg := range ac.Arguments {
			arg.Accept(visitor)
		}
		if ac.Extends != nil {
			ac.Extends.Accept(visitor)
		}
		for _, impl := range ac.Implements {
			impl.Accept(visitor)
		}
		for _, stmt := range ac.Body {
			stmt.Accept(visitor)
		}
	}
}

func (ac *AnonymousClass) expressionNode() {}

func (ac *AnonymousClass) String() string {
	var out strings.Builder
	out.WriteString("new class")
	
	if len(ac.Arguments) > 0 {
		out.WriteString("(")
		for i, arg := range ac.Arguments {
			if i > 0 {
				out.WriteString(", ")
			}
			out.WriteString(arg.String())
		}
		out.WriteString(")")
	}
	
	if ac.Extends != nil {
		out.WriteString(" extends ")
		out.WriteString(ac.Extends.String())
	}
	
	if len(ac.Implements) > 0 {
		out.WriteString(" implements ")
		for i, impl := range ac.Implements {
			if i > 0 {
				out.WriteString(", ")
			}
			out.WriteString(impl.String())
		}
	}
	
	out.WriteString(" {\n")
	for _, stmt := range ac.Body {
		out.WriteString("  " + stmt.String() + "\n")
	}
	out.WriteString("}")
	
	return out.String()
}


// SpreadExpression 表示展开表达式 (...expr)
type SpreadExpression struct {
	BaseNode
	Argument Expression `json:"argument"`
}

func NewSpreadExpression(pos lexer.Position, argument Expression) *SpreadExpression {
	return &SpreadExpression{
		BaseNode: BaseNode{
			Kind:     ASTUnpack, // 使用PHP的ZEND_AST_UNPACK 
			Position: pos,
			LineNo:   uint32(pos.Line),
		},
		Argument: argument,
	}
}

func (se *SpreadExpression) GetChildren() []Node {
	if se.Argument != nil {
		return []Node{se.Argument}
	}
	return []Node{}
}

func (se *SpreadExpression) Accept(visitor Visitor) {
	visitor.Visit(se)
}

func (se *SpreadExpression) String() string {
	return "..." + se.Argument.String()
}

func (se *SpreadExpression) expressionNode() {}
