<?php

class Test {
    public readonly string $prop;

    public function __construct() {
        echo "Constructor: Setting prop\n";
        $this->prop = "constructor_value";
        echo "Constructor: prop value is now: " . $this->prop . "\n";
    }
}

echo "Creating object...\n";
$obj = new Test();
echo "After constructor, prop value: " . $obj->prop . "\n";

?>