<?php

echo "=== Complete Generator Test Suite ===\n";

// Test 1: Basic generator
function simpleGenerator() {
    yield 1;
    yield 2;
    yield 3;
}

echo "Test 1: Basic generator\n";
foreach (simpleGenerator() as $value) {
    echo "  Value: $value\n";
}

// Test 2: Generator with keys
function keyValueGenerator() {
    yield 'a' => 10;
    yield 'b' => 20;
    yield 'c' => 30;
}

echo "\nTest 2: Generator with keys\n";
foreach (keyValueGenerator() as $key => $value) {
    echo "  Key: $key, Value: $value\n";
}

// Test 3: Yield from array
function arrayYieldFrom() {
    echo "  Before array delegation\n";
    yield from [100, 200, 300];
    echo "  After array delegation\n";
}

echo "\nTest 3: Yield from array\n";
foreach (arrayYieldFrom() as $key => $value) {
    echo "  Array[$key] = $value\n";
}

// Test 4: Yield from generator
function innerGen() {
    yield 'inner1';
    yield 'inner2';
}

function outerGen() {
    echo "  Before generator delegation\n";
    yield from innerGen();
    echo "  After generator delegation\n";
}

echo "\nTest 4: Yield from generator\n";
foreach (outerGen() as $value) {
    echo "  Delegated: $value\n";
}

// Test 5: Complex yield from with mixed content
function complexGenerator() {
    yield 'start';
    yield from [10, 20];
    yield 'middle';
    yield from ['x', 'y'];
    yield 'end';
}

echo "\nTest 5: Complex yield from\n";
foreach (complexGenerator() as $key => $value) {
    echo "  Mixed[$key] = $value\n";
}

echo "\n=== All generator tests completed ===\n";

?>