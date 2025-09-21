<?php

echo "=== Testing Reflection Functions ===\n";

// Test get_class
class TestClass {
    public function test() {
        return "test";
    }
}

$obj = new TestClass();
echo "get_class: " . get_class($obj) . "\n";

// Test func_num_args (basic)
function test_args($a, $b, $c = null) {
    return func_num_args();
}
echo "func_num_args: " . test_args(1, 2) . "\n";

// Test is_subclass_of
class Parent {}
class Child extends Parent {}
echo "is_subclass_of: " . (class_exists('Child') ? 'Child class exists' : 'Child class missing') . "\n";

echo "=== Test completed ===\n";

?>