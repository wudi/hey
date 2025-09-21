<?php
namespace MyNamespace;

class TestClass {
    public function test() {
        return "Hello from namespace";
    }
}

$obj = new TestClass();
echo $obj->test() . "\n";

namespace AnotherNamespace;

class TestClass {
    public function test() {
        return "Hello from another namespace";
    }
}

$obj = new TestClass();
echo $obj->test() . "\n";

// Test fully qualified names
$obj2 = new \MyNamespace\TestClass();
echo $obj2->test() . "\n";