<?php

echo "=== Testing get_class with Anonymous Class Inheritance ===\n";

class BaseClass {
    public function getMyClass() {
        return get_class($this);
    }
}

// Test anonymous class extending base class
$anon = new class extends BaseClass {
    public function getAnonClass() {
        return get_class($this);
    }
};

echo "Anonymous class: " . get_class($anon) . "\n";
echo "Anonymous->getMyClass() (inherited): " . $anon->getMyClass() . "\n";
echo "Anonymous->getAnonClass() (own method): " . $anon->getAnonClass() . "\n";

// Test that both methods return the same anonymous class name
$class1 = $anon->getMyClass();
$class2 = $anon->getAnonClass();
echo "Both methods return same class: " . ($class1 === $class2 ? 'YES' : 'NO') . "\n";

echo "=== Anonymous inheritance test completed ===\n";

?>