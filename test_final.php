<?php

echo "=== Final Test ===\n";

class BaseWithFinal {
    final public function finalMethod() {
        return "This method cannot be overridden";
    }
}

final class FinalClass {
    public function method() {
        return "Final class method";
    }
}

$finalObj = new FinalClass();
echo "Final class: " . $finalObj->method() . "\n";
$baseObj = new BaseWithFinal();
echo "Final method: " . $baseObj->finalMethod() . "\n";

echo "Test completed.\n";