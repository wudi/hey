package compiler

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser"
)

// TestInterfaceDeclaration tests the compilation of interface declarations
func TestInterfaceDeclaration(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name: "Simple interface declaration",
			phpCode: `<?php
interface Drawable {
    public function draw();
}
?>`,
		},
		{
			name: "Interface with multiple methods",
			phpCode: `<?php
interface Shape {
    public function area();
    public function perimeter();
}
?>`,
		},
		{
			name: "Interface with extends",
			phpCode: `<?php
interface ColoredShape extends Drawable {
    public function setColor($color);
}
?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			// Verify interface was stored
			require.Greater(t, len(comp.interfaces), 0, "No interfaces were compiled")
		})
	}
}

// TestTraitDeclaration tests the compilation of trait declarations
func TestTraitDeclaration(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name: "Simple trait declaration",
			phpCode: `<?php
trait Loggable {
    public function log($message) {
        echo "LOG: " . $message . "\n";
    }
}
?>`,
		},
		{
			name: "Trait with property",
			phpCode: `<?php
trait Timestampable {
    private $timestamp;
    
    public function setTimestamp($time) {
        $this->timestamp = $time;
    }
}
?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			// Verify trait was stored
			require.Greater(t, len(comp.traits), 0, "No traits were compiled")
		})
	}
}

// TestTraitUsage tests the compilation of trait usage in classes
func TestTraitUsage(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name: "Class using trait",
			phpCode: `<?php
trait Loggable {
    public function log($message) {
        echo "LOG: " . $message;
    }
}

class User {
    use Loggable;
    
    public function getName() {
        return "user";
    }
}
?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			// Verify both trait and class were compiled
			require.Greater(t, len(comp.traits), 0, "No traits were compiled")
			require.Greater(t, len(comp.classes), 0, "No classes were compiled")
		})
	}
}

// TestInterfaceImplementation tests classes implementing interfaces
func TestInterfaceImplementation(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name: "Class implementing interface",
			phpCode: `<?php
interface Drawable {
    public function draw();
}

class Circle implements Drawable {
    public function draw() {
        echo "Drawing circle";
    }
}
?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			// Verify both interface and class were compiled
			require.Greater(t, len(comp.interfaces), 0, "No interfaces were compiled")
			require.Greater(t, len(comp.classes), 0, "No classes were compiled")
		})
	}
}

// TestEnumDeclaration tests the compilation of enum declarations
func TestEnumDeclaration(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name: "Simple enum declaration",
			phpCode: `<?php
enum Status {
    case PENDING;
    case APPROVED;
    case REJECTED;
}
?>`,
		},
		{
			name: "Backed enum with string values",
			phpCode: `<?php
enum Color: string {
    case RED = 'red';
    case GREEN = 'green';
    case BLUE = 'blue';
}
?>`,
		},
		{
			name: "Enum with method",
			phpCode: `<?php
enum Size {
    case SMALL;
    case MEDIUM;
    case LARGE;
    
    public function getLabel() {
        return match($this) {
            Size::SMALL => 'Small',
            Size::MEDIUM => 'Medium', 
            Size::LARGE => 'Large',
        };
    }
}
?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			// Verify enum was compiled as class
			require.Greater(t, len(comp.classes), 0, "No enums were compiled")
		})
	}
}
