<?php

trait ReturnTrait {
    public function returnString() {
        return "string value";
    }

    public function returnThis() {
        return $this;
    }

    public function returnParam($param) {
        return $param;
    }
}

class UsesReturnTrait {
    use ReturnTrait;

    public $property = "test";
}

echo "=== Testing return values in trait methods ===\n";
$obj = new UsesReturnTrait();

echo "Return string: " . $obj->returnString() . "\n";
echo "Return param: " . $obj->returnParam("param value") . "\n";
echo "Return this: ";
var_dump($obj->returnThis());