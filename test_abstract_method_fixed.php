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

    // Now implementing getName() method
    public function getName() {
        return "rectangle";
    }
}

// This should work - all abstract methods implemented
$circle = new Circle(5);
$circle->describe();

// This should now work too - Rectangle now implements getName()
$rectangle = new Rectangle(4, 6);
$rectangle->describe();