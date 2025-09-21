<?php

class A {
    const NAME = "A";

    public static function test() {
        echo "self::NAME = " . self::NAME . "\n";
        echo "static::NAME = " . static::NAME . "\n";
    }
}

class B extends A {
    const NAME = "B";
}

echo "=== Static Constant Test ===\n";

echo "Calling A::test():\n";
A::test();

echo "\nCalling B::test():\n";
B::test();