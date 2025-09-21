<?php

// Test match expression features that should work

// Test 1: Complex expressions in conditions
$x = 10;
$result = match($x) {
    5 + 5 => "calculated ten",
    20 - 10 => "also ten",
    default => "not ten"
};
echo "Complex condition: $result\n";

// Test 2: Variables as conditions
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

// Test 3: Return value from match
function processValue($val) {
    return match($val) {
        0 => "zero",
        1 => "one",
        default => "many"
    };
}

echo "Function return: " . processValue(0) . "\n";
echo "Function return: " . processValue(5) . "\n";

// Test 4: Nested match
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

// Test 5: String matching
$color = "blue";
$mood = match($color) {
    "red", "orange" => "warm",
    "blue", "green" => "cool",
    "black", "white" => "neutral",
    default => "unknown"
};
echo "Color mood: $mood\n";