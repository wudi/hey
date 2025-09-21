<?php

echo "=== Anonymous Class Test ===\n";

// Test 1: Simple anonymous class
$obj = new class {
    public function greet() {
        return "Hello from anonymous class";
    }
};

echo "Test 1: " . $obj->greet() . "\n";

// Test 2: Anonymous class with constructor
$obj2 = new class("World") {
    private $name;

    public function __construct($name) {
        $this->name = $name;
    }

    public function sayHello() {
        return "Hello, " . $this->name . "!";
    }
};

echo "Test 2: " . $obj2->sayHello() . "\n";

// Test 3: Anonymous class with inheritance
class BaseClass {
    protected $value = "base";

    public function getValue() {
        return $this->value;
    }
}

$obj3 = new class extends BaseClass {
    public function getValue() {
        return "extended: " . parent::getValue();
    }
};

echo "Test 3: " . $obj3->getValue() . "\n";

echo "=== Anonymous class tests completed ===\n";

?>