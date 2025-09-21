<?php

echo "=== Basic OOP Test ===\n";

// 1. Basic Class and Object
class SimpleClass {
    public $property = "value";

    public function method() {
        return "method called";
    }
}
$obj = new SimpleClass();
echo "Property: " . $obj->property . "\n";
echo "Method: " . $obj->method() . "\n";

echo "Test completed.\n";