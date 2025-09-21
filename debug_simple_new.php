<?php
namespace TestNamespace;

class SimpleClass {
    public function hello() {
        return "Hello";
    }
}

// Just try to create an object - no method call
$obj = new SimpleClass();
echo "Object created successfully\n";