<?php

class MagicClass {
    public function __call($method, $args) {
        return "Called undefined method: $method with args: " . implode(", ", $args);
    }

    public static function __callStatic($method, $args) {
        return "Called undefined static method: $method with args: " . implode(", ", $args);
    }
}

$magic = new MagicClass();
echo "Instance call: " . $magic->undefinedMethod("arg1", "arg2") . "
";
echo "Static call: " . MagicClass::undefinedStaticMethod("arg3", "arg4") . "
";
