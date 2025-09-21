<?php

echo "=== Abstract Classes Test ===\n";

abstract class AbstractShape {
    abstract public function getArea();

    public function describe() {
        return "I am a shape";
    }
}

class Rectangle extends AbstractShape {
    private $width;
    private $height;

    public function __construct($w, $h) {
        $this->width = $w;
        $this->height = $h;
    }

    public function getArea() {
        return $this->width * $this->height;
    }
}

$rect = new Rectangle(5, 10);
echo "Rectangle area: " . $rect->getArea() . "\n";
echo "Description: " . $rect->describe() . "\n";

echo "Test completed.\n";