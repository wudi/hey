<?php

class Animal {
    public $name;

    public function __construct($name) {
        echo "Animal constructor called\n";
        $this->name = $name;
        echo "Animal constructor completed\n";
    }
}

class Dog extends Animal {
    public function __construct($name) {
        echo "Dog constructor called\n";
        parent::__construct($name);
        echo "Dog constructor completed\n";
    }
}

$dog = new Dog("Buddy");
echo "Dog name: " . $dog->name . "\n";