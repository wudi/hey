<?php

echo "=== Interface instanceof Test ===\n";

interface Drawable {
    public function draw();
}

interface Colorable {
    public function setColor($color);
}

class Circle implements Drawable, Colorable {
    private $color = "black";

    public function draw() {
        return "Drawing a " . $this->color . " circle";
    }

    public function setColor($color) {
        $this->color = $color;
    }
}

$circle = new Circle();

echo "Circle instanceof Circle: " . (($circle instanceof Circle) ? "true" : "false") . "\n";
echo "Circle instanceof Drawable: " . (($circle instanceof Drawable) ? "true" : "false") . "\n";
echo "Circle instanceof Colorable: " . (($circle instanceof Colorable) ? "true" : "false") . "\n";

echo "Test completed.\n";