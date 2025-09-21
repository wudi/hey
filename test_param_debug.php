<?php

class ParamTest {
    public function testParams($param1, $param2 = "default") {
        echo "Param1: ";
        var_dump($param1);
        echo "Param2: ";
        var_dump($param2);
        echo "This object: ";
        var_dump($this);
    }
}

$obj = new ParamTest();
echo "=== Parameter Debug Test ===\n";
$obj->testParams("hello", "world");