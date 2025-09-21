<?php

trait SimpleTrait {
    public function traitMethod($param1, $param2 = "default") {
        echo "Trait method called\n";
        echo "Param1: ";
        var_dump($param1);
        echo "Param2: ";
        var_dump($param2);
        echo "This: ";
        var_dump($this);
        return "trait result";
    }
}

class UsesTrait {
    use SimpleTrait;

    public $prop = "property value";
}

echo "=== Testing trait method parameters ===\n";
$obj = new UsesTrait();
$result = $obj->traitMethod("first", "second");
echo "Result: $result\n";