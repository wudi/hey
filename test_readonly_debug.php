<?php

class Test {
    public readonly string $prop;

    public function __construct() {
        echo "Constructor: About to set readonly property\n";
        $this->prop = "initial";
        echo "Constructor: Readonly property set\n";
    }
}

echo "Creating object...\n";
$obj = new Test();
echo "Object created successfully\n";

?>