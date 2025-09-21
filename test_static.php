<?php

echo "=== Static Test ===\n";

class StaticClass {
    public static $staticProperty = "static value";

    public static function staticMethod() {
        return "static method called";
    }
}

echo "Static property: " . StaticClass::$staticProperty . "\n";
echo "Static method: " . StaticClass::staticMethod() . "\n";

echo "Test completed.\n";