<?php

echo "=== Testing Advanced PHP Features ===\n";

// Test file operations
try {
    echo "file_get_contents: " . (function_exists('file_get_contents') ? "EXISTS" : "NOT IMPLEMENTED") . "\n";
} catch (Error $e) {
    echo "file_get_contents NOT IMPLEMENTED\n";
}

// Test regular expressions
try {
    $result = preg_match('/test/', 'testing');
    echo "preg_match: " . $result . "\n";
} catch (Error $e) {
    echo "preg_match NOT IMPLEMENTED\n";
}

// Test array functions
try {
    $arr = [3, 1, 4, 1, 5];
    sort($arr);
    echo "sort: " . implode(',', $arr) . "\n";
} catch (Error $e) {
    echo "sort NOT IMPLEMENTED\n";
}

try {
    echo "array_merge: " . count(array_merge([1,2], [3,4])) . "\n";
} catch (Error $e) {
    echo "array_merge NOT IMPLEMENTED\n";
}

// Test sprintf
try {
    echo "sprintf: " . sprintf("Hello %s, you are %d years old", "John", 25) . "\n";
} catch (Error $e) {
    echo "sprintf NOT IMPLEMENTED\n";
}

echo "=== Test completed ===\n";

?>