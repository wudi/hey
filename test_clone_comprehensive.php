<?php

class ComplexClone {
    public $id;
    public $name;
    public $counter = 0;
    private static $nextId = 1;

    public function __construct($name) {
        $this->id = self::$nextId++;
        $this->name = $name;
        echo "Constructed: ID={$this->id}, Name={$this->name}\n";
    }

    public function __clone() {
        echo "Cloning object ID={$this->id}\n";
        // Give clone a new ID
        $this->id = self::$nextId++;
        // Append (clone) to name
        $this->name = $this->name . " (clone)";
        // Reset counter for clone
        $this->counter = 0;
        echo "Clone created: ID={$this->id}, Name={$this->name}\n";
    }

    public function increment() {
        $this->counter++;
    }

    public function display() {
        echo "Object: ID={$this->id}, Name={$this->name}, Counter={$this->counter}\n";
    }
}

echo "=== Comprehensive Clone Test ===\n\n";

$obj1 = new ComplexClone("Original");
$obj1->increment();
$obj1->increment();

echo "\nBefore cloning:\n";
$obj1->display();

echo "\nCloning...\n";
$obj2 = clone $obj1;

echo "\nAfter cloning:\n";
echo "Original: ";
$obj1->display();
echo "Clone: ";
$obj2->display();

echo "\nModifying clone:\n";
$obj2->increment();

echo "\nFinal state:\n";
echo "Original: ";
$obj1->display();
echo "Clone: ";
$obj2->display();