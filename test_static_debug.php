<?php

abstract class A {
    public static function test() {
        echo "In A::test() - __CLASS__ = " . __CLASS__ . "\n";
        echo "In A::test() - self::class = " . self::class . "\n";
        echo "In A::test() - static::class = " . static::class . "\n";
    }
}

class B extends A {
}

class C extends A {
}

echo "=== Debug Test ===\n";

echo "Calling A::test():\n";
A::test();

echo "\nCalling B::test():\n";
B::test();

echo "\nCalling C::test():\n";
C::test();