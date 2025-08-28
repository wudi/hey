// +build !legacy

package ast

// This file is intentionally empty when not using legacy mode.
// The legacy node definitions in node.go are disabled to prevent conflicts
// with the new modern node definitions in modern_nodes.go and declaration_nodes.go.

// To enable legacy mode, build with: go build -tags legacy
// This will include the original node.go definitions.