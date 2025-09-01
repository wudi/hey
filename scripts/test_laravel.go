// +build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser"
)

func main() {
	laravelDir := "/home/ubuntu/framework"
	
	var phpFiles []string
	totalFiles := 0
	successCount := 0
	errorCount := 0
	startTime := time.Now()
	
	// Find all PHP files
	err := filepath.Walk(laravelDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".php") {
			phpFiles = append(phpFiles, path)
		}
		return nil
	})
	
	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		return
	}
	
	fmt.Printf("Found %d PHP files in Laravel framework\n", len(phpFiles))
	fmt.Println("Starting parsing test...")
	
	var failedFiles []string
	
	for _, filePath := range phpFiles {
		totalFiles++
		
		// Read file content
		content, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("ERROR reading %s: %v\n", filePath, err)
			errorCount++
			failedFiles = append(failedFiles, filePath)
			continue
		}
		
		// Create lexer and parser
		l := lexer.New(string(content))
		p := parser.New(l)
		
		// Parse the file
		p.ParseProgram()
		errors := p.Errors()
		
		if len(errors) > 0 {
			fmt.Printf("PARSE ERROR in %s:\n", filePath)
			for _, parseErr := range errors {
				fmt.Printf("  %s\n", parseErr)
			}
			fmt.Println("")
			errorCount++
			failedFiles = append(failedFiles, filePath)
		} else {
			successCount++
			if totalFiles%100 == 0 {
				fmt.Printf("Progress: %d/%d files processed (%.1f%% success)\n", 
					totalFiles, len(phpFiles), float64(successCount)/float64(totalFiles)*100)
			}
		}
		
		// Stop on first error for detailed analysis
		if len(errors) > 0 {
			fmt.Printf("\nStopping on first error for analysis...\n")
			fmt.Printf("Failed file: %s\n", filePath)
			fmt.Printf("File size: %d bytes\n", len(content))
			
			// Show first 1500 characters of the file
			preview := string(content)
			if len(preview) > 1500 {
				preview = preview[:1500] + "..."
			}
			fmt.Printf("File content preview:\n%s\n", preview)
			break
		}
	}
	
	duration := time.Since(startTime)
	
	fmt.Printf("\n=== SUMMARY ===\n")
	fmt.Printf("Total files: %d\n", totalFiles)
	fmt.Printf("Successfully parsed: %d (%.1f%%)\n", successCount, float64(successCount)/float64(totalFiles)*100)
	fmt.Printf("Parse errors: %d (%.1f%%)\n", errorCount, float64(errorCount)/float64(totalFiles)*100)
	fmt.Printf("Duration: %v\n", duration)
	
	if len(failedFiles) > 0 {
		fmt.Printf("\nFirst failed file for analysis: %s\n", failedFiles[0])
		os.Exit(1)
	}
}