<?php

// Test if the issue is with ALL method parameters or just the first one

class TestClass {
    public function zeroParams() {
        echo "Zero params method called
";
        return "zero result";
    }
    
    public function oneParam($p1) {
        echo "One param - received: ";
        var_dump($p1);
        return "one result";
    }
    
    public function threeParams($p1, $p2, $p3) {
        echo "Three params:
";
        echo "P1: "; var_dump($p1);
        echo "P2: "; var_dump($p2);  
        echo "P3: "; var_dump($p3);
        return "three result";
    }
}

$obj = new TestClass();

echo "=== Zero parameters ===
";
echo $obj->zeroParams() . "

";

echo "=== One parameter ===
";
echo $obj->oneParam("single") . "

";

echo "=== Three parameters ===
";
echo $obj->threeParams("first", "second", "third") . "

";
