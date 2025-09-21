<?php

class Test {
    public $name;

    public function __construct($name) {
        echo "Constructor called with: " . $name . "\n";
        $this->name = $name;
        echo "Assignment completed\n";
    }
}

$obj = new Test("hello");
echo "Final name: " . $obj->name . "\n";