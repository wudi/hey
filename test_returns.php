<?php

class ReturnTest {
    public function returnSimpleValue() {
        return "simple";
    }
    
    public function returnNumber() {
        return 42;
    }
    
    public function returnThis() {
        return $this;
    }
}

$obj = new ReturnTest();
echo "Return simple: " . $obj->returnSimpleValue() . "
";
echo "Return number: " . $obj->returnNumber() . "
";
echo "Return this: ";
var_dump($obj->returnThis());
