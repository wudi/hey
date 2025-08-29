package ast

// Visitor 访问者接口，用于遍历AST节点
type Visitor interface {
	// Visit 访问节点，返回是否继续遍历子节点
	Visit(node Node) bool
}

// VisitorFunc 函数类型的访问者
type VisitorFunc func(node Node) bool

// Visit 实现Visitor接口
func (f VisitorFunc) Visit(node Node) bool {
	return f(node)
}

// Walk 深度优先遍历AST节点
func Walk(visitor Visitor, node Node) {
	if node == nil {
		return
	}

	if !visitor.Visit(node) {
		return
	}

	// 根据节点类型遍历子节点
	switch n := node.(type) {
	case *Program:
		for _, stmt := range n.Body {
			Walk(visitor, stmt)
		}
	case *EchoStatement:
		for _, arg := range n.Arguments {
			Walk(visitor, arg)
		}
	case *PrintStatement:
		for _, arg := range n.Arguments {
			Walk(visitor, arg)
		}
	case *ExpressionStatement:
		Walk(visitor, n.Expression)
	case *AssignmentExpression:
		Walk(visitor, n.Left)
		Walk(visitor, n.Right)
	case *BinaryExpression:
		Walk(visitor, n.Left)
		Walk(visitor, n.Right)
	case *UnaryExpression:
		Walk(visitor, n.Operand)
	case *ArrayExpression:
		for _, elem := range n.Elements {
			Walk(visitor, elem)
		}
	case *IfStatement:
		Walk(visitor, n.Test)
		for _, stmt := range n.Consequent {
			Walk(visitor, stmt)
		}
		for _, stmt := range n.Alternate {
			Walk(visitor, stmt)
		}
	case *WhileStatement:
		Walk(visitor, n.Test)
		for _, stmt := range n.Body {
			Walk(visitor, stmt)
		}
	case *ForStatement:
		Walk(visitor, n.Init)
		Walk(visitor, n.Test)
		Walk(visitor, n.Update)
		for _, stmt := range n.Body {
			Walk(visitor, stmt)
		}
	case *FunctionDeclaration:
		Walk(visitor, n.Name)
		for _, param := range n.Parameters {
			Walk(visitor, param.DefaultValue)
		}
		for _, stmt := range n.Body {
			Walk(visitor, stmt)
		}
	case *ReturnStatement:
		Walk(visitor, n.Argument)
	case *BlockStatement:
		for _, stmt := range n.Body {
			Walk(visitor, stmt)
		}
	case *CallExpression:
		Walk(visitor, n.Callee)
		for _, arg := range n.Arguments {
			Walk(visitor, arg)
		}
	// 对于叶子节点（Variable, StringLiteral, NumberLiteral等），不需要进一步遍历
	}
}

// WalkFunc 使用函数作为访问者遍历AST
func WalkFunc(node Node, fn func(Node) bool) {
	Walk(VisitorFunc(fn), node)
}

// Inspector 检查器接口，用于在遍历过程中检查和收集信息
type Inspector interface {
	Inspect(node Node) bool
}

// InspectorFunc 函数类型的检查器
type InspectorFunc func(node Node) bool

// Inspect 实现Inspector接口
func (f InspectorFunc) Inspect(node Node) bool {
	return f(node)
}

// Inspect 使用检查器遍历AST
func Inspect(node Node, inspector Inspector) {
	Walk(VisitorFunc(inspector.Inspect), node)
}

// InspectFunc 使用函数作为检查器遍历AST
func InspectFunc(node Node, fn func(Node) bool) {
	Inspect(node, InspectorFunc(fn))
}

// Transformer 转换器接口，用于修改AST节点
type Transformer interface {
	Transform(node Node) Node
}

// TransformerFunc 函数类型的转换器
type TransformerFunc func(node Node) Node

// Transform 实现Transformer接口
func (f TransformerFunc) Transform(node Node) Node {
	return f(node)
}

// Transform 应用转换器修改AST
func Transform(node Node, transformer Transformer) Node {
	if node == nil {
		return nil
	}

	// 先转换子节点
	switch n := node.(type) {
	case *Program:
		for i, stmt := range n.Body {
			n.Body[i] = Transform(stmt, transformer).(Statement)
		}
	case *EchoStatement:
		for i, arg := range n.Arguments {
			n.Arguments[i] = Transform(arg, transformer).(Expression)
		}
	case *PrintStatement:
		for i, arg := range n.Arguments {
			n.Arguments[i] = Transform(arg, transformer).(Expression)
		}
	case *ExpressionStatement:
		n.Expression = Transform(n.Expression, transformer).(Expression)
	case *AssignmentExpression:
		n.Left = Transform(n.Left, transformer).(Expression)
		n.Right = Transform(n.Right, transformer).(Expression)
	case *BinaryExpression:
		n.Left = Transform(n.Left, transformer).(Expression)
		n.Right = Transform(n.Right, transformer).(Expression)
	case *UnaryExpression:
		n.Operand = Transform(n.Operand, transformer).(Expression)
	case *ArrayExpression:
		for i, elem := range n.Elements {
			n.Elements[i] = Transform(elem, transformer).(Expression)
		}
	case *IfStatement:
		n.Test = Transform(n.Test, transformer).(Expression)
		for i, stmt := range n.Consequent {
			n.Consequent[i] = Transform(stmt, transformer).(Statement)
		}
		for i, stmt := range n.Alternate {
			n.Alternate[i] = Transform(stmt, transformer).(Statement)
		}
	case *WhileStatement:
		n.Test = Transform(n.Test, transformer).(Expression)
		for i, stmt := range n.Body {
			n.Body[i] = Transform(stmt, transformer).(Statement)
		}
	case *ForStatement:
		if n.Init != nil {
			n.Init = Transform(n.Init, transformer).(Expression)
		}
		if n.Test != nil {
			n.Test = Transform(n.Test, transformer).(Expression)
		}
		if n.Update != nil {
			n.Update = Transform(n.Update, transformer).(Expression)
		}
		for i, stmt := range n.Body {
			n.Body[i] = Transform(stmt, transformer).(Statement)
		}
	case *FunctionDeclaration:
		n.Name = Transform(n.Name, transformer).(Identifier)
		for i, param := range n.Parameters {
			if param.DefaultValue != nil {
				n.Parameters[i].DefaultValue = Transform(param.DefaultValue, transformer).(Expression)
			}
		}
		for i, stmt := range n.Body {
			n.Body[i] = Transform(stmt, transformer).(Statement)
		}
	case *ReturnStatement:
		if n.Argument != nil {
			n.Argument = Transform(n.Argument, transformer).(Expression)
		}
	case *BlockStatement:
		for i, stmt := range n.Body {
			n.Body[i] = Transform(stmt, transformer).(Statement)
		}
	case *CallExpression:
		n.Callee = Transform(n.Callee, transformer).(Expression)
		for i, arg := range n.Arguments {
			n.Arguments[i] = Transform(arg, transformer).(Expression)
		}
	}

	// 然后转换当前节点
	return transformer.Transform(node)
}

// TransformFunc 使用函数作为转换器修改AST
func TransformFunc(node Node, fn func(Node) Node) Node {
	return Transform(node, TransformerFunc(fn))
}

// Filter 过滤器接口，用于过滤AST节点
type Filter interface {
	Filter(node Node) bool
}

// FilterFunc 函数类型的过滤器
type FilterFunc func(node Node) bool

// Filter 实现Filter接口
func (f FilterFunc) Filter(node Node) bool {
	return f(node)
}

// FindAll 查找所有满足条件的节点
func FindAll(node Node, filter Filter) []Node {
	var results []Node
	Walk(VisitorFunc(func(n Node) bool {
		if filter.Filter(n) {
			results = append(results, n)
		}
		return true
	}), node)
	return results
}

// FindAllFunc 使用函数作为过滤器查找节点
func FindAllFunc(node Node, fn func(Node) bool) []Node {
	return FindAll(node, FilterFunc(fn))
}

// FindFirst 查找第一个满足条件的节点
func FindFirst(node Node, filter Filter) Node {
	var result Node
	Walk(VisitorFunc(func(n Node) bool {
		if result == nil && filter.Filter(n) {
			result = n
			return false // 找到第一个后停止遍历
		}
		return result == nil // 如果还没找到继续遍历
	}), node)
	return result
}

// FindFirstFunc 使用函数作为过滤器查找第一个节点
func FindFirstFunc(node Node, fn func(Node) bool) Node {
	return FindFirst(node, FilterFunc(fn))
}

// Count 计算满足条件的节点数量
func Count(node Node, filter Filter) int {
	count := 0
	Walk(VisitorFunc(func(n Node) bool {
		if filter.Filter(n) {
			count++
		}
		return true
	}), node)
	return count
}

// CountFunc 使用函数作为过滤器计算节点数量
func CountFunc(node Node, fn func(Node) bool) int {
	return Count(node, FilterFunc(fn))
}