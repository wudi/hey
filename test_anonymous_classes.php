<?php
// Test anonymous classes
$obj = new class {
    public function test() {
        return "Hello from anonymous class";
    }
};

echo $obj->test() . "\n";