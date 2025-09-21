<?php

class Test {
    private $value = 42;

    public function getValue() {
        echo "Inside getValue, value is: " . $this->value . "
";
        return $this->value;
    }

    public function setValue($newValue) {
        echo "Setting value to: " . $newValue . "
";
        $this->value = $newValue;
        echo "Value after setting: " . $this->value . "
";
        return $this;
    }
}

$test = new Test();
echo "1. getValue(): " . $test->getValue() . "
";
$test->setValue(100);
echo "2. getValue() after setValue: " . $test->getValue() . "
";
