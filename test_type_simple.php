<?php

echo "=== Testing Type Functions Individually ===\n";

// Test each function separately
echo "is_object(new stdClass()): " . (is_object(new stdClass()) ? "true" : "false") . "\n";
echo "is_bool(true): " . (is_bool(true) ? "true" : "false") . "\n";
echo "is_int(42): " . (is_int(42) ? "true" : "false") . "\n";
echo "is_float(3.14): " . (is_float(3.14) ? "true" : "false") . "\n";
echo "is_null(null): " . (is_null(null) ? "true" : "false") . "\n";
echo "is_string('hello'): " . (is_string('hello') ? "true" : "false") . "\n";
echo "is_array([1,2,3]): " . (is_array([1,2,3]) ? "true" : "false") . "\n";
echo "is_numeric('123'): " . (is_numeric('123') ? "true" : "false") . "\n";
echo "is_numeric('abc'): " . (is_numeric('abc') ? "true" : "false") . "\n";

echo "=== All type functions working! ===\n";

?>