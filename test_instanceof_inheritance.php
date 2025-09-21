<?php

echo "=== Instanceof Inheritance Test ===\n";

class ParentClass {
    public function parentMethod() {
        return "parent";
    }
}

class ChildClass extends ParentClass {
    public function childMethod() {
        return "child";
    }
}

$parent = new ParentClass();
$child = new ChildClass();

echo "Parent instanceof ParentClass: " . (($parent instanceof ParentClass) ? "true" : "false") . "\n";
echo "Child instanceof ChildClass: " . (($child instanceof ChildClass) ? "true" : "false") . "\n";
echo "Child instanceof ParentClass: " . (($child instanceof ParentClass) ? "true" : "false") . "\n";
echo "Parent instanceof ChildClass: " . (($parent instanceof ChildClass) ? "true" : "false") . "\n";

echo "Test completed.\n";