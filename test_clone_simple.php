<?php

class TestClone {
    public $value = 10;

    public function __clone() {
        echo "In __clone method\n";
        $this->value = 20;
        echo "Changed value to 20\n";
    }
}

echo "=== Testing __clone ===\n";
$obj1 = new TestClone();
echo "Original value: " . $obj1->value . "\n";

$obj2 = clone $obj1;
echo "Clone value: " . $obj2->value . "\n";
echo "Original still: " . $obj1->value . "\n";