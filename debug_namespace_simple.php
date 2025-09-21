<?php
namespace TestNamespace;

class SimpleClass {
    public function hello() {
        return "Hello from TestNamespace";
    }
}

$obj = new SimpleClass();
echo $obj->hello();