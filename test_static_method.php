<?php

class StaticTest {
    public static function staticMethod($p1, $p2) {
        echo "Static P1: ";
        var_dump($p1);
        echo "Static P2: ";
        var_dump($p2);
        return "static result";
    }
}

echo "Testing static method:
";
$result = StaticTest::staticMethod("static1", "static2");
echo "Result: " . $result . "
";
