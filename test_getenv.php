<?php

echo "=== Testing getenv() Function ===\n";

// Test 1: Get an existing environment variable
// Set a test variable first
putenv("TEST_VAR=hello_world");
$result = getenv("TEST_VAR");
echo "Get TEST_VAR: " . $result . "\n";

// Test 2: Get a non-existent variable (should return false)
$result = getenv("NON_EXISTENT_VAR");
echo "Get non-existent: " . ($result === false ? 'false' : $result) . "\n";

// Test 3: Get PATH variable (commonly exists)
$path = getenv("PATH");
echo "PATH exists: " . ($path !== false ? 'true' : 'false') . "\n";

// Test 4: Get HOME variable (commonly exists on Unix)
$home = getenv("HOME");
echo "HOME exists: " . ($home !== false ? 'true' : 'false') . "\n";

// Test 5: Set and get a variable with special characters
putenv("TEST_SPECIAL=value with spaces and special chars!");
$special = getenv("TEST_SPECIAL");
echo "Special chars: " . $special . "\n";

// Test 6: Set and get empty value
putenv("TEST_EMPTY=");
$empty = getenv("TEST_EMPTY");
echo "Empty value: '" . $empty . "' (type: " . gettype($empty) . ")\n";

// Test 7: Override an existing variable
putenv("TEST_VAR=new_value");
$overridden = getenv("TEST_VAR");
echo "Overridden value: " . $overridden . "\n";

// Test 8: Get all environment variables (no parameter)
$all_vars = getenv();
echo "All vars is array: " . (is_array($all_vars) ? 'true' : 'false') . "\n";
echo "All vars has TEST_VAR: " . (isset($all_vars['TEST_VAR']) ? 'true' : 'false') . "\n";

// Test 9: Case sensitivity
putenv("test_lowercase=lower");
putenv("TEST_LOWERCASE=UPPER");
$lower = getenv("test_lowercase");
$upper = getenv("TEST_LOWERCASE");
echo "Lowercase: " . $lower . "\n";
echo "Uppercase: " . $upper . "\n";

// Test 10: Numeric value
putenv("TEST_NUMBER=12345");
$number = getenv("TEST_NUMBER");
echo "Number value: " . $number . " (type: " . gettype($number) . ")\n";

echo "=== getenv() test completed ===\n";

?>