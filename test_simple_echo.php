<?php

class SimpleTest {
    public function simpleEcho($param) {
        echo "Parameter is: " . $param . "
";
        return $param;
    }
}

$obj = new SimpleTest();
echo "Calling simpleEcho with hello:
";
$result = $obj->simpleEcho("hello");
echo "Returned: " . $result . "
";
