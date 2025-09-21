<?php
// Start in global namespace
class GlobalClass {
    public function test() {
        return "Global";
    }
}

namespace TestNamespace;

// Now we're in TestNamespace
class TestClass {
    public function test() {
        return "TestNamespace";
    }
}

// Create objects
$global = new \GlobalClass();  // Fully qualified
$test = new TestClass();       // Current namespace

echo $global->test() . "\n";
echo $test->test() . "\n";