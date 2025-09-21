<?php

echo "=== Instanceof Test ===\n";

class SimpleClass {
    public $property = "value";
}

$obj = new SimpleClass();
echo "SimpleClass instanceof SimpleClass: " . (($obj instanceof SimpleClass) ? "true" : "false") . "\n";

echo "Test completed.\n";