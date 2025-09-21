<?php

echo "=== Testing Common PHP Functions ===\n";

// Test string functions
try {
    echo "strlen: " . strlen("hello") . "\n";
} catch (Error $e) {
    echo "strlen NOT IMPLEMENTED\n";
}

try {
    echo "substr: " . substr("hello", 1, 3) . "\n";
} catch (Error $e) {
    echo "substr NOT IMPLEMENTED\n";
}

try {
    echo "str_replace: " . str_replace("world", "PHP", "hello world") . "\n";
} catch (Error $e) {
    echo "str_replace NOT IMPLEMENTED\n";
}

// Test array functions
try {
    $arr = [1, 2, 3];
    echo "array_push: ";
    array_push($arr, 4);
    echo count($arr) . "\n";
} catch (Error $e) {
    echo "array_push NOT IMPLEMENTED\n";
}

try {
    echo "in_array: " . (in_array(2, [1, 2, 3]) ? "true" : "false") . "\n";
} catch (Error $e) {
    echo "in_array NOT IMPLEMENTED\n";
}

try {
    echo "array_keys: " . count(array_keys(['a' => 1, 'b' => 2])) . "\n";
} catch (Error $e) {
    echo "array_keys NOT IMPLEMENTED\n";
}

echo "=== Test completed ===\n";

?>