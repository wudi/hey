<?php

class ThisTest {
    public $property = "test property";
    
    public function testThis($param) {
        echo "Parameter: " . $param . "
";
        echo "This property: " . $this->property . "
";
        return $this;
    }
}

$obj = new ThisTest();
$result = $obj->testThis("hello");
echo "Method returned: ";
var_dump($result);
