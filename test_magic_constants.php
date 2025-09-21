<?php

class TestClass {
    public function testMethod() {
        echo "__CLASS__ in method: " . __CLASS__ . "\n";
        echo "__METHOD__ in method: " . __METHOD__ . "\n";
        echo "__FUNCTION__ in method: " . __FUNCTION__ . "\n";
    }

    public static function staticMethod() {
        echo "__CLASS__ in static method: " . __CLASS__ . "\n";
        echo "__METHOD__ in static method: " . __METHOD__ . "\n";
        echo "__FUNCTION__ in static method: " . __FUNCTION__ . "\n";
    }
}

function globalFunction() {
    echo "__CLASS__ in global function: " . __CLASS__ . "\n";
    echo "__METHOD__ in global function: " . __METHOD__ . "\n";
    echo "__FUNCTION__ in global function: " . __FUNCTION__ . "\n";
}

echo "=== Magic Constants Test ===\n";

echo "=== Global context ===\n";
echo "__CLASS__ in global: " . __CLASS__ . "\n";
echo "__FUNCTION__ in global: " . __FUNCTION__ . "\n";

echo "\n=== Global function ===\n";
globalFunction();

echo "\n=== Instance method ===\n";
$obj = new TestClass();
$obj->testMethod();

echo "\n=== Static method ===\n";
TestClass::staticMethod();