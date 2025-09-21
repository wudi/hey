<?php
namespace TestNamespace;

class SimpleClass {
    public function hello() {
        return "Hello";
    }
}

$obj = new SimpleClass();
echo $obj->hello() . "\n";