<?php

echo "=== Magic Methods Test ===\n";

class MagicClass {
    private $data = [];

    public function __get($name) {
        return $this->data[$name] ?? null;
    }

    public function __set($name, $value) {
        $this->data[$name] = $value;
    }

    public function __toString() {
        return "MagicClass object";
    }

    public function __invoke($param) {
        return "Invoked with: " . $param;
    }
}

$magic = new MagicClass();
$magic->property = "dynamic value";
echo "__get/__set: " . $magic->property . "\n";
echo "__toString: " . $magic . "\n";
echo "__invoke: " . $magic("test") . "\n";

echo "Test completed.\n";