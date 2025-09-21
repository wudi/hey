<?php

class Resource {
    private $id;

    public function __construct($id) {
        $this->id = $id;
        echo "Resource {$this->id} acquired\n";
    }

    public function __destruct() {
        echo "Resource {$this->id} released\n";
    }
}

echo "=== Complex Destructor Test ===\n";

// Test 1: Normal scope
{
    $r1 = new Resource(1);
    echo "In scope\n";
}
echo "After scope\n";

// Test 2: Assignment and unset
$r2 = new Resource(2);
$r3 = $r2; // Same object, different variable
unset($r2);
echo "After unset r2\n";
unset($r3);
echo "After unset r3\n";

// Test 3: Function scope
function testFunction() {
    $r4 = new Resource(4);
    echo "In function\n";
    return $r4;
}

$r5 = testFunction();
echo "After function\n";

echo "Script ending...\n";