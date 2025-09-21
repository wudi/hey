<?php

class PropertyTest {
    public $publicProp = "initial";
    private $privateProp = "private";

    public function testPropertyAccess() {
        echo "Initial public property: " . $this->publicProp . "\n";
        $this->publicProp = "modified";
        echo "Modified public property: " . $this->publicProp . "\n";

        echo "Private property: " . $this->privateProp . "\n";
        $this->privateProp = "changed";
        echo "Changed private property: " . $this->privateProp . "\n";

        return $this;
    }
}

$obj = new PropertyTest();
echo "=== Property Assignment Test ===\n";
$result = $obj->testPropertyAccess();
echo "Method returned object: ";
var_dump($result);