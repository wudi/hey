<?php

trait ThisTrait {
    public function getProperty() {
        return $this->property;
    }

    public function setProperty($value) {
        $this->property = $value;
        return $this;
    }
}

class UsesThisTrait {
    use ThisTrait;

    public $property = "initial value";
}

echo "=== Testing $this in trait methods ===\n";
$obj = new UsesThisTrait();

echo "Initial property: " . $obj->getProperty() . "\n";

$result = $obj->setProperty("new value");
echo "After set property: " . $obj->getProperty() . "\n";

echo "Method returned: ";
var_dump($result);