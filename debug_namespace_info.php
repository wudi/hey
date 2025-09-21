<?php
namespace FirstNamespace;
echo "Current namespace should be FirstNamespace\n";

class TestClass {
    public function hello() {
        return "Hello from FirstNamespace";
    }
}

namespace SecondNamespace;
echo "Current namespace should be SecondNamespace\n";

class TestClass {
    public function hello() {
        return "Hello from SecondNamespace";
    }
}