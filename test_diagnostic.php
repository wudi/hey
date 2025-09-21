<?php

echo "=== Diagnostic Test for Argument Passing ===
";

// Test 1: Simple method with one parameter
class DiagnosticClass {
    public function testMethod($param) {
        echo "Received parameter type: " . gettype($param) . "
";
        echo "Parameter value: ";
        var_dump($param);
        return "method result";
    }
    
    public function __call($method, $args) {
        echo "Magic method called with:
";
        echo "Method name type: " . gettype($method) . "
";
        echo "Method name: ";
        var_dump($method);
        echo "Args type: " . gettype($args) . "
";
        echo "Args: ";
        var_dump($args);
        return "magic result";
    }
}

$obj = new DiagnosticClass();

echo "1. Regular method call:
";
echo "Result: " . $obj->testMethod("hello") . "

";

echo "2. Magic method call:
";
echo "Result: " . $obj->undefinedMethod("world") . "

";
