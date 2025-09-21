<?php

echo "=== File Magic Constants Test ===\n";

echo "__FILE__: " . __FILE__ . "\n";
echo "__LINE__: " . __LINE__ . "\n";
echo "__DIR__: " . __DIR__ . "\n";

class TestClass {
    public function testMethod() {
        echo "In method - __FILE__: " . __FILE__ . "\n";
        echo "In method - __LINE__: " . __LINE__ . "\n";
        echo "In method - __DIR__: " . __DIR__ . "\n";
    }
}

function testFunction() {
    echo "In function - __FILE__: " . __FILE__ . "\n";
    echo "In function - __LINE__: " . __LINE__ . "\n";
    echo "In function - __DIR__: " . __DIR__ . "\n";
}

$obj = new TestClass();
$obj->testMethod();

testFunction();