<?php

echo "=== Readonly Properties Test ===\n";

// Test 1: Basic readonly property with constructor assignment
class ReadonlyTest {
    public readonly string $value;

    public function __construct(string $val) {
        $this->value = $val;
    }

    public function getValue() {
        return $this->value;
    }

    public function tryModify() {
        $this->value = "modified";
    }
}

echo "Test 1: Basic readonly property\n";
$obj1 = new ReadonlyTest("initial");
echo "Value: " . $obj1->getValue() . "\n";

try {
    $obj1->tryModify();
    echo "ERROR: Modification should have failed!\n";
} catch (Exception $e) {
    echo "SUCCESS: " . $e->getMessage() . "\n";
}

// Test 2: Constructor property promotion with readonly
class PromotedReadonly {
    public function __construct(
        public readonly string $name,
        public readonly int $age
    ) {}

    public function modify() {
        $this->name = "changed";
    }
}

echo "\nTest 2: Promoted readonly properties\n";
$obj2 = new PromotedReadonly("John", 25);
echo "Name: " . $obj2->name . ", Age: " . $obj2->age . "\n";

try {
    $obj2->modify();
    echo "ERROR: Modification should have failed!\n";
} catch (Exception $e) {
    echo "SUCCESS: " . $e->getMessage() . "\n";
}

echo "\n=== All tests completed ===\n";

?>