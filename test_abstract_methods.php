<?php

abstract class AbstractShape {
    protected $name;

    public function __construct($name) {
        $this->name = $name;
    }

    // Abstract method
    abstract public function calculateArea();

    // Concrete method
    public function getName() {
        return $this->name;
    }
}

class Circle extends AbstractShape {
    private $radius;

    public function __construct($radius) {
        parent::__construct("Circle");
        $this->radius = $radius;
    }

    public function calculateArea() {
        return 3.14159 * $this->radius * $this->radius;
    }
}

class Square extends AbstractShape {
    private $side;

    public function __construct($side) {
        parent::__construct("Square");
        $this->side = $side;
    }

    public function calculateArea() {
        return $this->side * $this->side;
    }
}

echo "=== Testing Abstract Classes ===\n";

$circle = new Circle(5);
echo $circle->getName() . " area: " . $circle->calculateArea() . "\n";

$square = new Square(4);
echo $square->getName() . " area: " . $square->calculateArea() . "\n";

// This should fail - can't instantiate abstract class
// $shape = new AbstractShape("Test");