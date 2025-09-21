<?php

class Test {
    public readonly string $prop;

    public function __construct() {
        echo "Before assignment in constructor\n";
        $this->prop = "initial";
        echo "After assignment in constructor, value: " . $this->prop . "\n";
    }

    public function tryModify() {
        echo "Before modification attempt\n";
        $this->prop = "modified";
        echo "After modification, value: " . $this->prop . "\n";
    }
}

echo "Creating object...\n";
$obj = new Test();
echo "Object created successfully\n";

echo "Calling tryModify...\n";
$obj->tryModify();
echo "Done\n";

?>