<?php

echo "=== Testing sprintf Function ===\n";

// Test basic string formatting
$result1 = sprintf("Hello %s!", "World");
echo "Basic string: " . $result1 . "\n";

// Test with integer
$result2 = sprintf("Number: %d", 42);
echo "Integer: " . $result2 . "\n";

// Test with float
$result3 = sprintf("Float: %.2f", 3.14159);
echo "Float: " . $result3 . "\n";

// Test with multiple arguments
$result4 = sprintf("Hello %s, you are %d years old and have %.2f dollars", "John", 25, 123.45);
echo "Multiple args: " . $result4 . "\n";

// Test percent escape
$result5 = sprintf("This is 100%% complete");
echo "Percent: " . $result5 . "\n";

echo "=== sprintf test completed ===\n";

?>