<?php

class TestParams {
    public function oneParam($p1) {
        echo "One param method - P1: ";
        var_dump($p1);
    }
    
    public function twoParams($p1, $p2) {
        echo "Two param method - P1: ";
        var_dump($p1);
        echo "Two param method - P2: ";
        var_dump($p2);
    }
}

$obj = new TestParams();

echo "Calling method with one parameter:
";
$obj->oneParam("single");

echo "
Calling method with two parameters:
";
$obj->twoParams("first", "second");
