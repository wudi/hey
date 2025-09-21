<?php

echo "=== Testing get_class with Inheritance ===\n";

// Test basic inheritance
class Parent {
    public function getMyClass() {
        return get_class($this);
    }

    public function getParentClass() {
        return get_class();
    }
}

class Child extends Parent {
    public function getChildClass() {
        return get_class($this);
    }
}

// Test with objects
$parent = new Parent();
$child = new Child();

echo "Parent object class: " . get_class($parent) . "\n";
echo "Child object class: " . get_class($child) . "\n";

// Test calling from methods
echo "Parent->getMyClass(): " . $parent->getMyClass() . "\n";
echo "Child->getMyClass(): " . $child->getMyClass() . "\n";
echo "Child->getChildClass(): " . $child->getChildClass() . "\n";

// Test with anonymous classes
$anon_parent = new class extends Parent {
    public function getAnonClass() {
        return get_class($this);
    }
};

echo "Anonymous class extending Parent: " . get_class($anon_parent) . "\n";
echo "Anonymous->getMyClass(): " . $anon_parent->getMyClass() . "\n";
echo "Anonymous->getAnonClass(): " . $anon_parent->getAnonClass() . "\n";

// Test static context (should work with object parameter)
class StaticTest {
    public static function getClassStatic($obj) {
        return get_class($obj);
    }
}

echo "StaticTest::getClassStatic(child): " . StaticTest::getClassStatic($child) . "\n";

echo "=== All inheritance tests completed ===\n";

?>