<?php

class ChainTest {
    public $value = 0;

    public function add($n) {
        echo "Adding " . $n . " to " . $this->value . "\n";
        $this->value += $n;
        return $this;
    }

    public function multiply($n) {
        echo "Multiplying " . $this->value . " by " . $n . "\n";
        $this->value *= $n;
        return $this;
    }

    public function getValue() {
        return $this->value;
    }
}

$obj = new ChainTest();
echo "=== Method Chaining Test ===\n";

echo "Testing method chaining:\n";
$result = $obj->add(5)->multiply(2)->add(3);
echo "Final result: " . $result->getValue() . "\n";
echo "Object value: " . $obj->value . "\n";