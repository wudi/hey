<?php

class MagicClass {
    public function __call($method, $args) {
        echo "Method: ";
        var_dump($method);
        echo "Args: ";
        var_dump($args);
        return "Called method: " . $method;
    }
}

$magic = new MagicClass();
echo $magic->testMethod() . "
";
