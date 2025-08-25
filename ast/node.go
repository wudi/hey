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

// GetChildren 返回子节点
func (fd *FunctionDeclaration) GetChildren() []Node {
	children := []Node{fd.Name}
	for _, param := range fd.Parameters {
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
		out.WriteString("  " + stmt.String() + "\n")
	}
	out.WriteString("}")
	return out.String()
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
