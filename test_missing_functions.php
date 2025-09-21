<?php

echo "=== Testing Common PHP Functions ===\n";

// Test class reflection functions
class TestClass {
    public function test() {
        return "test";
    }
}

try {
    echo "Testing get_class: ";
    $obj = new TestClass();
    echo get_class($obj) . "\n";
} catch (Error $e) {
    echo "NOT IMPLEMENTED: " . $e->getMessage() . "\n";
}

try {
    echo "Testing func_num_args: ";
    function test_args($a, $b, $c = null) {
        return func_num_args();
    }
    echo test_args(1, 2) . "\n";
} catch (Error $e) {
    echo "NOT IMPLEMENTED: " . $e->getMessage() . "\n";
}

try {
    echo "Testing func_get_args: ";
    function test_get_args($a, $b) {
        return implode(',', func_get_args());
    }
    echo test_get_args('hello', 'world', 'extra') . "\n";
} catch (Error $e) {
    echo "NOT IMPLEMENTED: " . $e->getMessage() . "\n";
}

try {
    echo "Testing is_subclass_of: ";
    class Parent {}
    class Child extends Parent {}
    echo is_subclass_of('Child', 'Parent') ? 'true' : 'false';
    echo "\n";
} catch (Error $e) {
    echo "NOT IMPLEMENTED: " . $e->getMessage() . "\n";
}

echo "=== Test completed ===\n";

?>