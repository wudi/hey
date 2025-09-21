<?php

// Test 1: Parameter type declarations
function testStringParam(string $str): string {
    return "Got: " . $str;
}

// Test 2: Class type declarations
class TestClass {
    public function getName(): string {
        return "TestClass";
    }
}

function testClassParam(TestClass $obj): string {
    return $obj->getName();
}

// Test 3: Property type declarations (PHP 7.4+)
class TypedProperties {
    public string $name;
    public int $age;

    public function __construct(string $name, int $age) {
        $this->name = $name;
        $this->age = $age;
    }
}

// Test 4: Return type declarations
function getNumber(): int {
    return 42;
}

function getArray(): array {
    return [1, 2, 3];
}

// Test usage
echo testStringParam("hello") . "\n";

$obj = new TestClass();
echo testClassParam($obj) . "\n";

$typed = new TypedProperties("John", 30);
echo "Name: " . $typed->name . ", Age: " . $typed->age . "\n";

echo "Number: " . getNumber() . "\n";

$arr = getArray();
echo "Array length: " . count($arr) . "\n";