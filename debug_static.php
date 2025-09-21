<?php

class A {
    const NAME = "A";

    public static function test() {
        echo "A::test() called\n";
        echo "self::NAME = " . self::NAME . "\n";
        echo "static::NAME = " . static::NAME . "\n";
    }
}

class B extends A {
    const NAME = "B";
}

echo "Testing B::test():\n";
B::test();