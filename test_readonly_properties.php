<?php

// Test 1: Basic readonly property
class BasicReadonly {
    public readonly string $id;

    public function __construct(string $id) {
        $this->id = $id;  // First assignment should work
    }

    public function tryModify() {
        $this->id = "new_id";  // This should fail
    }
}

echo "=== Test 1: Basic readonly property ===\n";
$obj = new BasicReadonly("test123");
echo "Initial ID: " . $obj->id . "\n";

try {
    $obj->tryModify();
    echo "ERROR: Should have failed!\n";
} catch (Exception $e) {
    echo "SUCCESS: " . $e->getMessage() . "\n";
}

// Test 2: Constructor property promotion with readonly
class PromotedReadonly {
    public function __construct(
        public readonly string $name,
        public readonly int $value
    ) {}

    public function tryModify() {
        $this->name = "modified";  // Should fail
    }
}

echo "\n=== Test 2: Readonly promoted properties ===\n";
$obj2 = new PromotedReadonly("test", 42);
echo "Name: " . $obj2->name . ", Value: " . $obj2->value . "\n";

try {
    $obj2->tryModify();
    echo "ERROR: Should have failed!\n";
} catch (Exception $e) {
    echo "SUCCESS: " . $e->getMessage() . "\n";
}

// Test 3: Mix of readonly and regular properties
class MixedProperties {
    public readonly string $readonly_prop;
    public string $regular_prop;

    public function __construct() {
        $this->readonly_prop = "readonly_value";
        $this->regular_prop = "regular_value";
    }

    public function modifyRegular() {
        $this->regular_prop = "modified_regular";  // Should work
    }

    public function modifyReadonly() {
        $this->readonly_prop = "modified_readonly";  // Should fail
    }
}

echo "\n=== Test 3: Mixed properties ===\n";
$obj3 = new MixedProperties();
echo "Before: readonly=" . $obj3->readonly_prop . ", regular=" . $obj3->regular_prop . "\n";

$obj3->modifyRegular();
echo "After modifying regular: " . $obj3->regular_prop . "\n";

try {
    $obj3->modifyReadonly();
    echo "ERROR: Should have failed!\n";
} catch (Exception $e) {
    echo "SUCCESS: " . $e->getMessage() . "\n";
}

// Test 4: Compound assignment on readonly property
class CompoundAssignTest {
    public readonly int $counter;

    public function __construct() {
        $this->counter = 10;
    }

    public function increment() {
        $this->counter += 1;  // Should fail
    }
}

echo "\n=== Test 4: Compound assignment on readonly ===\n";
$obj4 = new CompoundAssignTest();
echo "Initial counter: " . $obj4->counter . "\n";

try {
    $obj4->increment();
    echo "ERROR: Should have failed!\n";
} catch (Exception $e) {
    echo "SUCCESS: " . $e->getMessage() . "\n";
}

echo "\n=== All tests completed ===\n";
?>