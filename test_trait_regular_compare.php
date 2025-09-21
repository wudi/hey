<?php

trait TestTrait {
    public function traitMethod($p1, $p2 = "default") {
        echo "In trait method:\n";
        echo "  P1: ";
        var_dump($p1);
        echo "  P2: ";
        var_dump($p2);
        return "trait result";
    }
}

class TestClass {
    use TestTrait;

    public function regularMethod($p1, $p2 = "default") {
        echo "In regular method:\n";
        echo "  P1: ";
        var_dump($p1);
        echo "  P2: ";
        var_dump($p2);
        return "regular result";
    }
}

echo "=== Comparing trait vs regular methods ===\n";
$obj = new TestClass();

echo "\n1. Regular method:\n";
$r1 = $obj->regularMethod("first", "second");
echo "Result: $r1\n";

echo "\n2. Trait method:\n";
$r2 = $obj->traitMethod("first", "second");
echo "Result: $r2\n";