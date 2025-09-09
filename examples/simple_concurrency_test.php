<?php
// Simple concurrency test that works with current parser

echo "=== Simple Concurrency Test ===\n";

// Test 1: Basic function registration
echo "Testing go() function registration...\n";
if (function_exists('go')) {
    echo "go() function is registered!\n";
} else {
    echo "ERROR: go() function not found\n";
}

// Test 2: WaitGroup class instantiation
echo "Testing WaitGroup class...\n";
$wg = new WaitGroup();
echo "WaitGroup created: " . $wg . "\n";

// Test 3: Simple closure (without use clause for now)
echo "Testing simple closure...\n";
$closure = function() {
    return "Hello from closure!";
};

echo "Closure created and callable: " . (is_callable($closure) ? "YES" : "NO") . "\n";

echo "=== Test Complete ===\n";
?>