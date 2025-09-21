<?php

// Test named arguments functionality

// Test 1: Simple function with named arguments
function greet(string $name, string $greeting = "Hello", string $punctuation = "!") {
    return $greeting . " " . $name . $punctuation;
}

// Test different combinations
echo greet("World") . "\n";                                    // Positional only
echo greet(name: "Alice") . "\n";                              // Only named (rest default)
echo greet(name: "Bob", greeting: "Hi") . "\n";                // Mixed with defaults
echo greet(punctuation: ".", name: "Charlie", greeting: "Hey") . "\n";  // All named, different order

// Test 2: Function with required parameters
function calculate(int $a, int $b, string $operation = "add") {
    switch ($operation) {
        case "add": return $a + $b;
        case "multiply": return $a * $b;
        case "subtract": return $a - $b;
        default: return 0;
    }
}

echo calculate(5, 3) . "\n";                           // Positional
echo calculate(a: 10, b: 2, operation: "multiply") . "\n";  // All named
echo calculate(b: 7, a: 3, operation: "subtract") . "\n";   // Named in different order

// Test 3: Mixing positional and named arguments
echo calculate(5, b: 10, operation: "add") . "\n";     // Mixed: positional then named