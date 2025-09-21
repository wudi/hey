<?php
// Test edge cases for namespace implementation - Fixed for PHP compatibility

namespace {
    // Global namespace
    class GlobalClass {
        public function test() { return "Global"; }
    }
}

namespace FirstNamespace {

    // Case 2: Namespaced class
    class TestClass {
        public function test() { return "FirstNamespace"; }
    }

    // Case 3: Interface in namespace
    interface TestInterface {
        public function interfaceMethod();
    }

    // Case 4: Class implementing namespaced interface
    class InterfaceUser implements TestInterface {
        public function interfaceMethod() {
            return "Interface implementation";
        }
        public function test() { return "InterfaceUser"; }
    }

    // Test cases for FirstNamespace
    $global = new \GlobalClass();
    echo $global->test() . "\n";

    $local = new TestClass();
    echo $local->test() . "\n";

    $interface = new InterfaceUser();
    echo $interface->test() . "\n";
    echo $interface->interfaceMethod() . "\n";
}

namespace SecondNamespace {

    // Case 8: Same class name in different namespace
    class TestClass {
        public function test() { return "SecondNamespace"; }
    }

    // Case 9: Access classes from previous namespaces
    $first = new \FirstNamespace\TestClass();
    echo $first->test() . "\n";

    $second = new TestClass();
    echo $second->test() . "\n";

    $global2 = new \GlobalClass();
    echo $global2->test() . "\n";
}