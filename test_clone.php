<?php

echo "=== Clone Test ===\n";

class Cloneable {
    public $value = 10;

    public function __clone() {
        $this->value = 20;
        echo "Object was cloned\n";
    }
}

$original = new Cloneable();
$clone = clone $original;
echo "Original value: " . $original->value . "\n";
echo "Clone value: " . $clone->value . "\n";

echo "Test completed.\n";