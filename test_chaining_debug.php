<?php

class Chainable {
    private $value = 0;

    public function add($n) {
        $this->value += $n;
        return $this;
    }

    public function multiply($n) {
        $this->value *= $n;
        return $this;
    }

    public function getValue() {
        return $this->value;
    }
}

$chain = new Chainable();
echo "Step by step:
";
$chain->add(5);
echo "After add(5): " . $chain->getValue() . "
";
$chain->multiply(3);  
echo "After multiply(3): " . $chain->getValue() . "
";

echo "
Method chaining:
";
$result = $chain->add(5)->multiply(3)->getValue();
echo "Chained result: " . $result . "
";
