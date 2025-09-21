<?php

abstract class Shape {
    abstract public function getArea();
    abstract public function getName();

    public function describe() {
        echo "This is a " . $this->getName() . " with area " . $this->getArea() . "\n";
    }
}

class Circle extends Shape {
    private $radius;

    public function __construct($radius) {
        $this->radius = $radius;
    }

    public function getArea() {
        return 3.14159 * $this->radius * $this->radius;
    }

    public function getName() {
        return "circle";
    }
}

class Rectangle extends Shape {
    private $width;
    private $height;

    public function __construct($width, $height) {
        $this->width = $width;
        $this->height = $height;
    }

    public function getArea() {
        return $this->width * $this->height;
    }

    // Missing getName() method - this should cause an error
}

// This should work - all abstract methods implemented
$circle = new Circle(5);
$circle->describe();

// This should fail - Rectangle doesn't implement getName()
$rectangle = new Rectangle(4, 6);
$rectangle->describe();