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

// PropertyAccessExpression 属性访问表达式
type PropertyAccessExpression struct {
	BaseNode
	Object   Expression `json:"object"`
	Property Expression `json:"property"`
}

func NewPropertyAccessExpression(pos lexer.Position, object, property Expression) *PropertyAccessExpression {
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

func (p *PropertyAccessExpression) GetChildren() []Node {
	children := make([]Node, 0, 2)
	if p.Object != nil {
		children = append(children, p.Object)
	}
	if p.Property != nil {
		children = append(children, p.Property)
	}
	return children
}

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
	obj := ""
	if p.Object != nil {
		obj = p.Object.String()
	}
	prop := ""
	if p.Property != nil {
		prop = p.Property.String()
	}
	return obj + "->" + prop
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
