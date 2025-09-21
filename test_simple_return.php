<?php

class Test {
    private $value = 42;

    public function getValue() {
        return $this->value;
    }
}

$test = new Test();
echo "Value: " . $test->getValue() . "
";
