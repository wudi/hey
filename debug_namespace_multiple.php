<?php
namespace FirstNamespace;

class TestClass {
    public function hello() {
        return "Hello from FirstNamespace";
    }
}

echo "Creating first class\n";
$obj1 = new TestClass();
echo $obj1->hello() . "\n";

namespace SecondNamespace;

class TestClass {
    public function hello() {
        return "Hello from SecondNamespace";
    }
}

echo "Creating second class\n";
$obj2 = new TestClass();
echo $obj2->hello() . "\n";