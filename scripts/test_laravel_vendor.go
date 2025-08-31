package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser"
)

type ParseResult struct {
	File    string
	Success bool
	Error   string
	Tokens  int
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run test_laravel_vendor.go <directory>")
		os.Exit(1)
	}

	directory := os.Args[1]
	fmt.Printf("Parsing all PHP files in: %s\n", directory)
	fmt.Println("=" + strings.Repeat("=", 50))

	var phpFiles []string
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".php") {
			phpFiles = append(phpFiles, path)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d PHP files to parse\n", len(phpFiles))
	fmt.Println()

	startTime := time.Now()
	var results []ParseResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Use a semaphore to limit concurrent parsing
	sem := make(chan struct{}, 10)

	for _, file := range phpFiles {
		wg.Add(1)
		go func(filename string) {
			defer wg.Done()
			sem <- struct{}{} // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			result := parseFile(filename)
			
			mu.Lock()
			results = append(results, result)
			if !result.Success {
				fmt.Printf("FAIL: %s - %s\n", result.File, result.Error)
			} else {
				fmt.Printf("OK:   %s (%d tokens)\n", result.File, result.Tokens)
			}
			mu.Unlock()
		}(file)
	}

	wg.Wait()
	elapsed := time.Since(startTime)

	fmt.Println()
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Printf("Results Summary:\n")
	fmt.Printf("Total files: %d\n", len(results))

	var successful, failed int
	var totalTokens int
	var failedFiles []ParseResult

	for _, result := range results {
		if result.Success {
			successful++
			totalTokens += result.Tokens
		} else {
			failed++
			failedFiles = append(failedFiles, result)
		}
	}

	fmt.Printf("Successful: %d\n", successful)
	fmt.Printf("Failed: %d\n", failed)
	fmt.Printf("Total tokens parsed: %d\n", totalTokens)
	fmt.Printf("Elapsed time: %v\n", elapsed)
	
	if failed > 0 {
		fmt.Printf("\nFailed files:\n")
		for _, result := range failedFiles {
			fmt.Printf("  %s: %s\n", result.File, result.Error)
		}
	}

	if failed > 0 {
		os.Exit(1)
	}
}

func parseFile(filename string) ParseResult {
	content, err := os.ReadFile(filename)
	if err != nil {
		return ParseResult{
			File:    filename,
			Success: false,
			Error:   fmt.Sprintf("read error: %v", err),
		}
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	
	// Parse the file
	_ = p.ParseProgram()
	
	if len(p.Errors()) > 0 {
		// Get first error for reporting
		firstError := p.Errors()[0]
		return ParseResult{
			File:    filename,
			Success: false,
			Error:   fmt.Sprintf("parse error: %s", firstError),
		}
	}

	// Count tokens for statistics
	l2 := lexer.New(string(content))
	tokenCount := 0
	for {
		tok := l2.NextToken()
		if tok.Type == lexer.T_EOF {
			break
		}
		tokenCount++
	}

	return ParseResult{
		File:    filename,
		Success: true,
		Tokens:  tokenCount,
	}
}