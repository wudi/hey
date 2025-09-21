<?php

echo "=== Testing Type Check Functions ===\n";

$obj = new stdClass();

// Test which type functions are missing
$tests = [
    'is_object' => $obj,
    'is_null' => null,
    'is_bool' => true,
    'is_int' => 42,
    'is_float' => 3.14,
    'is_numeric' => '123',
    'empty' => '',
    'isset' => $obj
];

foreach ($tests as $func => $value) {
    try {
        if ($func === 'isset') {
            echo "$func: " . (isset($value) ? "true" : "false") . "\n";
        } else if ($func === 'empty') {
            echo "$func: " . (empty($value) ? "true" : "false") . "\n";
        } else {
            echo "$func: " . ($func($value) ? "true" : "false") . "\n";
        }
    } catch (Error $e) {
        echo "$func: NOT IMPLEMENTED\n";
    }
}

echo "=== Test completed ===\n";

?>