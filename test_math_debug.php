<?php

class MathTest {
    public $value = 0;

    public function setValue($n) {
        echo "Setting value to " . $n . "\n";
        $this->value = $n;
        return $this;
    }

    public function add($n) {
        echo "Before add: value=" . $this->value . ", adding=" . $n . "\n";
        $this->value += $n;
        echo "After add: value=" . $this->value . "\n";
        return $this;
    }

    public function multiply($n) {
        echo "Before multiply: value=" . $this->value . ", multiplying=" . $n . "\n";
        $this->value *= $n;
        echo "After multiply: value=" . $this->value . "\n";
        return $this;
    }
}

$obj = new MathTest();
echo "=== Math Debug Test ===\n";

echo "Step by step:\n";
$obj->setValue(5);
$obj->multiply(2);
echo "Expected: 10, Got: " . $obj->value . "\n";