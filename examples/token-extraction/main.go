//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"strings"

	"github.com/wudi/php-parser/lexer"
)

// TokenAnalyzer 分析token的统计信息
type TokenAnalyzer struct {
	TokenCounts    map[lexer.TokenType]int
	TotalTokens    int
	Keywords       []string
	Identifiers    []string
	StringLiterals []string
	Numbers        []string
}

func NewTokenAnalyzer() *TokenAnalyzer {
	return &TokenAnalyzer{
		TokenCounts:    make(map[lexer.TokenType]int),
		Keywords:       make([]string, 0),
		Identifiers:    make([]string, 0),
		StringLiterals: make([]string, 0),
		Numbers:        make([]string, 0),
	}
}

func (ta *TokenAnalyzer) Analyze(tok lexer.Token) {
	ta.TokenCounts[tok.Type]++
	ta.TotalTokens++

	switch tok.Type {
	case lexer.T_STRING:
		// Check if it's a keyword or identifier
		if isKeyword(tok.Value) {
			ta.Keywords = append(ta.Keywords, tok.Value)
		} else {
			ta.Identifiers = append(ta.Identifiers, tok.Value)
		}
	case lexer.T_CONSTANT_ENCAPSED_STRING:
		ta.StringLiterals = append(ta.StringLiterals, tok.Value)
	case lexer.T_LNUMBER, lexer.T_DNUMBER:
		ta.Numbers = append(ta.Numbers, tok.Value)
	}
}

// isKeyword 检查是否为PHP关键字
func isKeyword(literal string) bool {
	keywords := map[string]bool{
		"class": true, "function": true, "if": true, "else": true, "while": true,
		"for": true, "foreach": true, "return": true, "echo": true, "print": true,
		"var": true, "public": true, "private": true, "protected": true, "static": true,
		"abstract": true, "final": true, "interface": true, "implements": true, "extends": true,
		"new": true, "this": true, "parent": true, "self": true, "true": true, "false": true,
		"null": true, "array": true, "object": true, "string": true, "int": true, "float": true,
		"bool": true, "try": true, "catch": true, "finally": true, "throw": true,
		"namespace": true, "use": true, "const": true, "global": true, "isset": true, "empty": true,
	}
	return keywords[strings.ToLower(literal)]
}

func main() {
	// Sample PHP code with various token types
	phpCode := `<?php
namespace App\Controllers;

use App\Models\User;
use App\Services\AuthService;

/**
 * User controller class
 */
class UserController 
{
    private $authService;
    private static $instance = null;
    
    const MAX_LOGIN_ATTEMPTS = 3;
    const SESSION_TIMEOUT = 3600;
    
    public function __construct(AuthService $authService) {
        $this->authService = $authService;
    }
    
    /**
     * Authenticate user
     * @param string $username
     * @param string $password
     * @return bool|array
     */
    public function login($username, $password) {
        if (empty($username) || empty($password)) {
            return false;
        }
        
        $user = User::findByUsername($username);
        
        if (!$user || !password_verify($password, $user->password_hash)) {
            return array('error' => 'Invalid credentials');
        }
        
        $sessionData = array(
            'user_id' => $user->id,
            'username' => $user->username,
            'login_time' => time(),
            'expires_at' => time() + self::SESSION_TIMEOUT
        );
        
        return $sessionData;
    }
    
    public function logout() {
        session_destroy();
        header('Location: /login');
        exit(0);
    }
}
?>`

	fmt.Println("=== Token Extraction Examples ===\n")

	// Create lexer
	l := lexer.New(phpCode)
	analyzer := NewTokenAnalyzer()

	fmt.Println("1. Extracting all tokens:")
	fmt.Printf("%-25s %-15s %-10s %s\n", "Token Type", "Literal", "Position", "Line:Column")
	fmt.Println(strings.Repeat("-", 75))

	// Extract and display all tokens
	for {
		tok := l.NextToken()
		if tok.Type == lexer.T_EOF {
			break
		}

		// Skip whitespace and comments for cleaner output
		if tok.Type == lexer.T_WHITESPACE ||
			tok.Type == lexer.T_COMMENT ||
			tok.Type == lexer.T_DOC_COMMENT {
			analyzer.Analyze(tok)
			continue
		}

		analyzer.Analyze(tok)

		// Format literal for display (truncate long strings)
		value := tok.Value
		if len(value) > 20 {
			value = value[:17] + "..."
		}

		fmt.Printf("%-25s %-15s %-10d %d:%d\n",
			tok.Type.String(),
			value,
			tok.Position.Offset,
			tok.Position.Line,
			tok.Position.Column)
	}

	fmt.Println()

	// Display analysis results
	fmt.Println("2. Token Analysis Results:")
	fmt.Printf("   Total tokens processed: %d\n", analyzer.TotalTokens)
	fmt.Printf("   Unique token types: %d\n\n", len(analyzer.TokenCounts))

	// Show token type statistics
	fmt.Println("3. Token Type Statistics:")
	for tokenType, count := range analyzer.TokenCounts {
		if count > 0 {
			fmt.Printf("   %-25s: %d\n", tokenType.String(), count)
		}
	}
	fmt.Println()

	// Show collected identifiers and literals
	fmt.Println("4. Extracted Content:")

	if len(analyzer.Keywords) > 0 {
		uniqueKeywords := removeDuplicates(analyzer.Keywords)
		fmt.Printf("   Keywords (%d unique): %s\n",
			len(uniqueKeywords), strings.Join(uniqueKeywords, ", "))
	}

	if len(analyzer.Identifiers) > 0 {
		uniqueIdentifiers := removeDuplicates(analyzer.Identifiers)
		fmt.Printf("   Identifiers (%d unique): %s\n",
			len(uniqueIdentifiers), strings.Join(uniqueIdentifiers, ", "))
	}

	if len(analyzer.StringLiterals) > 0 {
		fmt.Printf("   String literals (%d): %s\n",
			len(analyzer.StringLiterals), strings.Join(analyzer.StringLiterals, ", "))
	}

	if len(analyzer.Numbers) > 0 {
		fmt.Printf("   Numbers (%d): %s\n",
			len(analyzer.Numbers), strings.Join(analyzer.Numbers, ", "))
	}
	fmt.Println()

	// Example of filtering specific token types
	fmt.Println("5. Filtering Specific Token Types:")
	fmt.Printf("   Variables: %d\n", analyzer.TokenCounts[lexer.T_VARIABLE])
	fmt.Printf("   Functions: %d\n", analyzer.TokenCounts[lexer.T_FUNCTION])
	fmt.Printf("   Classes: %d\n", analyzer.TokenCounts[lexer.T_CLASS])
	fmt.Printf("   Operators: %d\n", countOperators(analyzer.TokenCounts))
	fmt.Printf("   Operators: %d\n", countOperators(analyzer.TokenCounts))

	// Show lexer state transitions example
	fmt.Println("\n6. Demonstrating String Interpolation Tokens:")
	interpolatedCode := `<?php
$name = "World";
echo "Hello, $name! Today is {$date}";
?>`

	l2 := lexer.New(interpolatedCode)
	fmt.Println("   Tokens in string interpolation:")
	for {
		tok := l2.NextToken()
		if tok.Type == lexer.T_EOF {
			break
		}
		if tok.Type == lexer.T_WHITESPACE {
			continue
		}

		fmt.Printf("     %s: '%s'\n", tok.Type.String(), tok.Value)
	}
}

// countOperators 计算操作符token的数量
func countOperators(tokenCounts map[lexer.TokenType]int) int {
	operatorTokens := []lexer.TokenType{
		lexer.T_PLUS_EQUAL, lexer.T_MINUS_EQUAL, lexer.T_MUL_EQUAL, lexer.T_DIV_EQUAL,
		lexer.T_IS_EQUAL, lexer.T_IS_NOT_EQUAL, lexer.T_IS_IDENTICAL, lexer.T_IS_NOT_IDENTICAL,
		lexer.T_IS_SMALLER_OR_EQUAL, lexer.T_IS_GREATER_OR_EQUAL,
		lexer.T_CONCAT_EQUAL, lexer.T_COALESCE_EQUAL,
	}

	count := 0
	for _, tokenType := range operatorTokens {
		count += tokenCounts[tokenType]
	}
	return count
}

// removeDuplicates 移除字符串切片中的重复项
func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}
