<?php

echo "=== Simple define() test ===\n";

// Test 1: Define a constant
echo "1. Defining MY_CONST\n";
$result = define('MY_CONST', 'Hello');
echo "Result: " . ($result ? 'true' : 'false') . "\n";

// Test 2: Check if it's defined
echo "2. Checking if MY_CONST is defined\n";
$isDefined = defined('MY_CONST');
echo "Defined: " . ($isDefined ? 'true' : 'false') . "\n";

// Test 3: Try to use the constant
echo "3. Using the constant\n";
echo "Value: " . MY_CONST . "\n";

echo "=== Test completed ===\n";

?>