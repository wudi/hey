<?php

echo "=== Simple getenv() test ===\n";

// Test 1: Set and get a variable
putenv("MY_TEST=hello");
echo "MY_TEST: " . getenv("MY_TEST") . "\n";

// Test 2: Get non-existent
$result = getenv("DOES_NOT_EXIST");
echo "Non-existent: " . ($result === false ? 'false' : 'not-false') . "\n";

// Test 3: Get all vars
$all = getenv();
echo "All is array: " . (is_array($all) ? 'true' : 'false') . "\n";
echo "Has MY_TEST: " . (isset($all['MY_TEST']) ? 'true' : 'false') . "\n";

// Test 4: Empty value
putenv("EMPTY=");
$empty = getenv("EMPTY");
echo "Empty is string: " . (is_string($empty) ? 'true' : 'false') . "\n";
echo "Empty length: " . strlen($empty) . "\n";

echo "=== Test completed ===\n";

?>