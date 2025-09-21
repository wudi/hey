<?php

class RegularTest {
    public function normalMethod($param1, $param2) {
        echo "Param1: ";
        var_dump($param1);
        echo "Param2: ";
        var_dump($param2);
        return "normal result";
    }
}

$obj = new RegularTest();
echo "Regular method call:
";
$result = $obj->normalMethod("first", "second");
echo "Result: " . $result . "
";
