<?php

echo "=== Simple Diagnostic Test ===
";

class Test {
    public function simpleMethod($param) {
        echo "Parameter received: ";
        var_dump($param);
        return "result";
    }
    
    public function __call($method, $args) {
        echo "Magic method - Method: ";
        var_dump($method);
        echo "Magic method - Args: ";
        var_dump($args);
        return "magic";
    }
}

$obj = new Test();

echo "1. Call simpleMethod with string:
";
$result = $obj->simpleMethod("test");
echo "Result: " . $result . "

";

echo "2. Call undefined method:
";
$result2 = $obj->undefinedMethod("param");
echo "Result: " . $result2 . "
";
