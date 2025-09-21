<?php

echo "=== Interfaces Test ===\n";

interface Drawable {
    public function draw();
}

class Circle implements Drawable {
    private $color = "black";

    public function draw() {
        return "Drawing a " . $this->color . " circle";
    }

    public function setColor($color) {
        $this->color = $color;
    }
}

$circle = new Circle();
$circle->setColor("red");
echo $circle->draw() . "\n";

echo "Test completed.\n";