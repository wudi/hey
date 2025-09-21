<?php

echo "=== Simple getcwd() test ===\n";

// Test 1: Get current working directory
$cwd = getcwd();
echo "Current directory: " . $cwd . "\n";

// Test 2: Verify it returns a string
echo "Is string: " . (is_string($cwd) ? 'true' : 'false') . "\n";

// Test 3: Verify it's not empty
echo "Not empty: " . (strlen($cwd) > 0 ? 'true' : 'false') . "\n";

// Test 4: Verify it starts with / (absolute path on Unix)
$first_char = substr($cwd, 0, 1);
echo "First character: '" . $first_char . "'\n";
echo "Is absolute path: " . ($first_char === '/' ? 'true' : 'false') . "\n";

// Test 5: Multiple calls return same result
$cwd1 = getcwd();
$cwd2 = getcwd();
echo "Consistent: " . ($cwd1 === $cwd2 ? 'true' : 'false') . "\n";

// Test 6: Return type
echo "Type: " . gettype(getcwd()) . "\n";

echo "=== Test completed ===\n";

?>