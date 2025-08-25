package ast

import (
	"encoding/json"
	"strings"

	"github.com/wudi/php-parser/lexer"
)

// Node 表示抽象语法树中的节点接口
type Node interface {
	// GetType 返回节点类型
	GetType() string
	// GetPosition 返回节点在源代码中的位置
	GetPosition() lexer.Position
	// String 返回节点的字符串表示
	String() string
	// ToJSON 转换为 JSON 表示
	ToJSON() ([]byte, error)
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
	Type     string         `json:"type"`
	Position lexer.Position `json:"position"`
}

// GetType 返回节点类型
func (b *BaseNode) GetType() string {
	return b.Type
}

// GetPosition 返回节点位置
func (b *BaseNode) GetPosition() lexer.Position {
	return b.Position
}

// ToJSON 转换为 JSON - 这个方法在每个具体类型中应该被重写
func (b *BaseNode) ToJSON() ([]byte, error) {
	return json.MarshalIndent(b, "", "  ")
}

// Program 表示整个 PHP 程序
type Program struct {
	BaseNode
	Body []Statement `json:"body"`
}

// NewProgram 创建新的程序节点
func NewProgram(pos lexer.Position) *Program {
	return &Program{
		BaseNode: BaseNode{Type: "Program", Position: pos},
		Body:     make([]Statement, 0),
	}
}

func (p *Program) String() string {
	var out strings.Builder
	for _, stmt := range p.Body {
		out.WriteString(stmt.String())
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
		BaseNode:  BaseNode{Type: "EchoStatement", Position: pos},
		Arguments: make([]Expression, 0),
	}
}

func (e *EchoStatement) statementNode() {}

func (e *EchoStatement) String() string {
	var args []string
	for _, arg := range e.Arguments {
		args = append(args, arg.String())
	}
	return "echo " + strings.Join(args, ", ") + ";"
}

// ExpressionStatement 表达式语句
type ExpressionStatement struct {
	BaseNode
	Expression Expression `json:"expression"`
}

func NewExpressionStatement(pos lexer.Position, expr Expression) *ExpressionStatement {
	return &ExpressionStatement{
		BaseNode:   BaseNode{Type: "ExpressionStatement", Position: pos},
		Expression: expr,
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
		BaseNode: BaseNode{Type: "AssignmentExpression", Position: pos},
		Left:     left,
		Right:    right,
		Operator: operator,
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
		BaseNode: BaseNode{Type: "BinaryExpression", Position: pos},
		Left:     left,
		Right:    right,
		Operator: operator,
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

func NewUnaryExpression(pos lexer.Position, operator string, operand Expression, prefix bool) *UnaryExpression {
	return &UnaryExpression{
		BaseNode: BaseNode{Type: "UnaryExpression", Position: pos},
		Operator: operator,
		Operand:  operand,
		Prefix:   prefix,
	}
}

func (ue *UnaryExpression) expressionNode() {}

func (ue *UnaryExpression) String() string {
	if ue.Prefix {
		return ue.Operator + ue.Operand.String()
	}
	return ue.Operand.String() + ue.Operator
}

// Variable 变量
type Variable struct {
	BaseNode
	Name string `json:"name"`
}

func NewVariable(pos lexer.Position, name string) *Variable {
	return &Variable{
		BaseNode: BaseNode{Type: "Variable", Position: pos},
		Name:     name,
	}
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
		BaseNode: BaseNode{Type: "StringLiteral", Position: pos},
		Value:    value,
		Raw:      raw,
	}
}

func (sl *StringLiteral) expressionNode() {}

func (sl *StringLiteral) String() string {
	return sl.Raw
}

// NumberLiteral 数字字面量
type NumberLiteral struct {
	BaseNode
	Value string `json:"value"`
	Kind  string `json:"kind"` // "integer" or "float"
}

func NewNumberLiteral(pos lexer.Position, value, kind string) *NumberLiteral {
	return &NumberLiteral{
		BaseNode: BaseNode{Type: "NumberLiteral", Position: pos},
		Value:    value,
		Kind:     kind,
	}
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
		BaseNode: BaseNode{Type: "BooleanLiteral", Position: pos},
		Value:    value,
	}
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
		BaseNode: BaseNode{Type: "NullLiteral", Position: pos},
	}
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
		BaseNode: BaseNode{Type: "ArrayExpression", Position: pos},
		Elements: make([]Expression, 0),
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
		BaseNode:   BaseNode{Type: "IfStatement", Position: pos},
		Test:       test,
		Consequent: make([]Statement, 0),
		Alternate:  make([]Statement, 0),
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
		BaseNode: BaseNode{Type: "WhileStatement", Position: pos},
		Test:     test,
		Body:     make([]Statement, 0),
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
		BaseNode: BaseNode{Type: "ForStatement", Position: pos},
		Body:     make([]Statement, 0),
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
	Name       Identifier  `json:"name"`
	Parameters []Parameter `json:"parameters"`
	ReturnType string      `json:"returnType,omitempty"`
	Body       []Statement `json:"body"`
}

type Parameter struct {
	Name         string     `json:"name"`
	DefaultValue Expression `json:"defaultValue,omitempty"`
	Type         string     `json:"type,omitempty"`
}

func NewFunctionDeclaration(pos lexer.Position, name Identifier) *FunctionDeclaration {
	return &FunctionDeclaration{
		BaseNode:   BaseNode{Type: "FunctionDeclaration", Position: pos},
		Name:       name,
		Parameters: make([]Parameter, 0),
		Body:       make([]Statement, 0),
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
		out.WriteString("  " + stmt.String() + "\n")
	}
	out.WriteString("}")

	return out.String()
}

// IdentifierNode 标识符节点
type IdentifierNode struct {
	BaseNode
	Name string `json:"name"`
}

func NewIdentifierNode(pos lexer.Position, name string) *IdentifierNode {
	return &IdentifierNode{
		BaseNode: BaseNode{Type: "Identifier", Position: pos},
		Name:     name,
	}
}

func (i *IdentifierNode) expressionNode() {}
func (i *IdentifierNode) identifierNode() {}

func (i *IdentifierNode) String() string {
	return i.Name
}

// ReturnStatement return语句
type ReturnStatement struct {
	BaseNode
	Argument Expression `json:"argument,omitempty"`
}

func NewReturnStatement(pos lexer.Position, arg Expression) *ReturnStatement {
	return &ReturnStatement{
		BaseNode: BaseNode{Type: "ReturnStatement", Position: pos},
		Argument: arg,
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
		BaseNode: BaseNode{Type: "BreakStatement", Position: pos},
	}
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
		BaseNode: BaseNode{Type: "ContinueStatement", Position: pos},
	}
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
		BaseNode: BaseNode{Type: "BlockStatement", Position: pos},
		Body:     make([]Statement, 0),
	}
}

func (bs *BlockStatement) statementNode() {}

func (bs *BlockStatement) String() string {
	var out strings.Builder
	out.WriteString("{\n")
	for _, stmt := range bs.Body {
		out.WriteString("  " + stmt.String() + "\n")
	}
	out.WriteString("}")
	return out.String()
}

// DocBlockComment 文档块注释
type DocBlockComment struct {
	BaseNode
	Content string `json:"content"`
	Raw     string `json:"raw"`
}

func NewDocBlockComment(pos lexer.Position, content, raw string) *DocBlockComment {
	return &DocBlockComment{
		BaseNode: BaseNode{Type: "DocBlockComment", Position: pos},
		Content:  content,
		Raw:      raw,
	}
}

func (dbc *DocBlockComment) expressionNode() {}

func (dbc *DocBlockComment) String() string {
	return dbc.Raw
}
