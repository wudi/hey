<?php

class MagicMethods {
    public function __call($method, $args) {
        echo "__call: Method '$method' called with:\n";
        var_dump($args);
        return "instance result";
    }

    public static function __callStatic($method, $args) {
        echo "__callStatic: Method '$method' called with:\n";
        var_dump($args);
        return "static result";
    }
}

echo "=== Testing __call ===\n";
$obj = new MagicMethods();
$result1 = $obj->undefinedInstance("param1", "param2");
echo "Result: $result1\n\n";

echo "=== Testing __callStatic ===\n";
$result2 = MagicMethods::undefinedStatic("static1", "static2");
echo "Result: $result2\n";