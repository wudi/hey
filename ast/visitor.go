package ast

// Visitor interface for AST traversal
type Visitor interface {
	Visit(node Node) bool
}

// Walk traverses the AST with a visitor
func Walk(v Visitor, node Node) {
	if v.Visit(node) {
		for _, child := range node.GetChildren() {
			Walk(v, child)
		}
	}
}