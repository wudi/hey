<?php

// Test match expression edge cases

// Test 1: Unhandled match (should throw error)
$value = 99;
try {
    $result = match($value) {
        1 => "one",
        2 => "two",
        3 => "three"
        // No default - should throw UnhandledMatchError
    };
    echo "Should not reach here\n";
} catch (Error $e) {
    echo "Caught error: UnhandledMatchError\n";
}

// Test 2: Complex expressions in conditions
$x = 10;
$result = match($x) {
    5 + 5 => "calculated ten",
    20 - 10 => "also ten",
    default => "not ten"
};
echo "Complex condition: $result\n";

// Test 3: Variables as conditions
$target = 42;
$a = 40;
$b = 41;
$c = 42;
$result = match($target) {
    $a => "forty",
    $b => "forty-one",
    $c => "forty-two",
    default => "other"
};
echo "Variable conditions: $result\n";

// Test 4: Return value from match
function processValue($val) {
    return match($val) {
        0 => "zero",
        1 => "one",
        default => "many"
    };
}

echo "Function return: " . processValue(0) . "\n";
echo "Function return: " . processValue(5) . "\n";

// Test 5: Nested match (if supported)
$type = "number";
$value = 2;
$result = match($type) {
    "number" => match($value) {
        1 => "one",
        2 => "two",
        default => "other number"
    },
    "string" => "text value",
    default => "unknown"
};
echo "Nested match: $result\n";