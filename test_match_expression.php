<?php

// Test match expression (PHP 8.0 feature)

// Test 1: Basic match expression
$value = 2;
$result = match($value) {
    1 => "one",
    2 => "two",
    3 => "three",
    default => "other"
};
echo "Basic match: $result\n";

// Test 2: Multiple conditions per arm
$value = 'b';
$result = match($value) {
    'a', 'b', 'c' => "first group",
    'd', 'e', 'f' => "second group",
    default => "other"
};
echo "Multiple conditions: $result\n";

// Test 3: Expression evaluation
$x = 5;
$result = match($x + 3) {
    6 => "six",
    7 => "seven",
    8 => "eight",
    default => "other"
};
echo "Expression match: $result\n";

// Test 4: Without default (should work if matched)
$status = 200;
$message = match($status) {
    200 => "OK",
    404 => "Not Found",
    500 => "Server Error"
};
echo "Status match: $message\n";

// Test 5: More complex example
function getValueType($value) {
    return match(gettype($value)) {
        'integer' => "It's an integer",
        'double' => "It's a float",
        'string' => "It's a string",
        'boolean' => "It's a boolean",
        default => "Unknown type"
    };
}

echo getValueType(42) . "\n";
echo getValueType("hello") . "\n";
echo getValueType(3.14) . "\n";