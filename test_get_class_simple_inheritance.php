<?php

echo "=== Testing get_class with Simple Inheritance ===\n";

// Test basic classes first
class SimpleParent {
    public function getMyClass() {
        return get_class($this);
    }
}

class SimpleChild extends SimpleParent {
}

// Test with objects
$parent = new SimpleParent();
echo "SimpleParent object class: " . get_class($parent) . "\n";

$child = new SimpleChild();
echo "SimpleChild object class: " . get_class($child) . "\n";

// Test calling method from parent in child context
echo "Child->getMyClass() (inherited method): " . $child->getMyClass() . "\n";

echo "=== Simple inheritance test completed ===\n";

?>