<?php

echo "=== Testing define() Function ===\n";

// Test 1: Basic constant definition
define('TEST_CONSTANT', 'Hello World');
echo "Basic constant: " . TEST_CONSTANT . "\n";

// Test 2: Numeric constant
define('PI_VALUE', 3.14159);
echo "Numeric constant: " . PI_VALUE . "\n";

// Test 3: Boolean constants
define('IS_TRUE', true);
define('IS_FALSE', false);
echo "Boolean true: " . (IS_TRUE ? 'true' : 'false') . "\n";
echo "Boolean false: " . (IS_FALSE ? 'true' : 'false') . "\n";

// Test 4: Case-sensitive by default
define('Case_Sensitive', 'lower');
// This should be undefined and cause an error in strict mode, but let's test differently
if (defined('CASE_SENSITIVE')) {
    echo "Case sensitive failed - should not be defined\n";
} else {
    echo "Case sensitive works - CASE_SENSITIVE not defined\n";
}

// Test 5: Case-insensitive constant (deprecated, but test the parameter)
$case_result = define('CASE_INSENSITIVE', 'works', true);
echo "Case insensitive: " . CASE_INSENSITIVE . "\n";
echo "Case insensitive define result: " . ($case_result ? 'true' : 'false') . "\n";

// Test 6: Array constant (PHP 5.6+)
define('ARRAY_CONST', array('a', 'b', 'c'));
echo "Array constant: " . implode(',', ARRAY_CONST) . "\n";

// Test 7: Return value tests
$result1 = define('NEW_CONST', 'value');
echo "Define return value: " . ($result1 ? 'true' : 'false') . "\n";

// Test 8: Redefinition attempt (should return false)
$result2 = define('NEW_CONST', 'different');
echo "Redefinition return: " . ($result2 ? 'true' : 'false') . "\n";

// Test 9: defined() function
echo "TEST_CONSTANT defined: " . (defined('TEST_CONSTANT') ? 'true' : 'false') . "\n";
echo "UNDEFINED_CONST defined: " . (defined('UNDEFINED_CONST') ? 'true' : 'false') . "\n";

// Test 10: null constant
define('NULL_CONST', null);
echo "Null constant: " . (NULL_CONST === null ? 'null' : 'not null') . "\n";

echo "=== define() test completed ===\n";

?>