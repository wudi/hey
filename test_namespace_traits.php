<?php
namespace TestNamespace {

    trait TestTrait {
        public function traitMethod() {
            return "From TestNamespace trait";
        }
    }

    class TestClass {
        use TestTrait;

        public function test() {
            return "TestClass: " . $this->traitMethod();
        }
    }

    $obj = new TestClass();
    echo $obj->test() . "\n";
}

namespace AnotherNamespace {

    // Using trait from another namespace
    class UsingExternalTrait {
        use \TestNamespace\TestTrait;

        public function test() {
            return "AnotherNamespace: " . $this->traitMethod();
        }
    }

    $obj = new UsingExternalTrait();
    echo $obj->test() . "\n";
}