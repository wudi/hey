<?php

function testFunction($param1, $param2) {
    echo "Function Param1: ";
    var_dump($param1);
    echo "Function Param2: ";
    var_dump($param2);
    return "function result";
}

echo "Testing function call:
";
$result = testFunction("arg1", "arg2");
echo "Result: " . $result . "
";
