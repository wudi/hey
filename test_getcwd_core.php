<?php

// Test getcwd basic functionality
$cwd = getcwd();

// Test 1: Returns a value
if ($cwd) {
    echo "PASS: getcwd returns a value\n";
} else {
    echo "FAIL: getcwd returns nothing\n";
}

// Test 2: Is string
if (is_string($cwd)) {
    echo "PASS: Returns string\n";
} else {
    echo "FAIL: Not a string\n";
}

// Test 3: Not empty
if (strlen($cwd) > 0) {
    echo "PASS: Not empty\n";
} else {
    echo "FAIL: Empty string\n";
}

// Test 4: Absolute path (starts with /)
if (substr($cwd, 0, 1) === '/') {
    echo "PASS: Absolute path\n";
} else {
    echo "FAIL: Not absolute path\n";
}

// Test 5: Multiple calls consistent
$cwd2 = getcwd();
if ($cwd === $cwd2) {
    echo "PASS: Consistent\n";
} else {
    echo "FAIL: Inconsistent\n";
}

echo "Directory: " . $cwd . "\n";

?>