<?php

class TestClass {
    public function getName() {
        return __CLASS__;
    }
}

echo "=== Testing __CLASS__ ===\n";
$obj = new TestClass();
echo "Class name: " . $obj->getName() . "\n";