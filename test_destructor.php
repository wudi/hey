<?php

class DestructorTest {
    public $name;

    public function __construct($name) {
        $this->name = $name;
        echo "Constructing: {$this->name}\n";
    }

    public function __destruct() {
        echo "Destructing: {$this->name}\n";
    }
}

echo "=== Testing Destructor ===\n";
$obj = new DestructorTest("Test Object");
echo "Object created\n";
unset($obj);
echo "Object unset\n";

$obj2 = new DestructorTest("Object 2");
echo "Script ending...\n";
// $obj2 should be destructed at script end