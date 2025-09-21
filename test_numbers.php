<?php

class DebugClass {
    public function testWithNumbers($n1, $n2) {
        echo "N1: " . $n1 . " (should be 42)
";
        echo "N2: " . $n2 . " (should be 99)
";
    }
}

$obj = new DebugClass();
echo "Testing with numbers:
";
$obj->testWithNumbers(42, 99);
