<?php

abstract class AbstractBase {
    abstract public function doSomething();

    public function concreteMethod() {
        echo "This is concrete\n";
    }
}

class ConcreteClass extends AbstractBase {
    public function doSomething() {
        echo "Doing something concrete\n";
    }
}

echo "=== Abstract Class Instantiation Test ===\n";

// This should work - concrete class
$obj = new ConcreteClass();
$obj->doSomething();
$obj->concreteMethod();

echo "Concrete class created successfully\n";

// This should fail - abstract class
try {
    $abstract = new AbstractBase();
    echo "ERROR: Abstract class was instantiated!\n";
} catch (Error $e) {
    echo "CORRECT: Cannot instantiate abstract class: " . $e->getMessage() . "\n";
}

echo "Test completed\n";