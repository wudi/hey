<?php

echo "=== Parent and Self Test ===\n";

class ParentWithMethod {
    public function method() {
        return "parent version";
    }
}

class ChildWithParent extends ParentWithMethod {
    public function method() {
        return "child version";
    }

    public function callParent() {
        return parent::method();
    }

    public function callSelf() {
        return self::method();
    }
}

$childObj = new ChildWithParent();
echo "Child method: " . $childObj->method() . "\n";
echo "Parent method via parent::: " . $childObj->callParent() . "\n";
echo "Self method via self::: " . $childObj->callSelf() . "\n";

echo "Test completed.\n";