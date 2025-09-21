<?php

abstract class A {
    public static function who() {
        echo __CLASS__ . "\n";
    }

    public static function test() {
        // self:: refers to the class where the method is defined (A)
        echo "self::who() = ";
        self::who();

        // static:: refers to the class that called the method (late static binding)
        echo "static::who() = ";
        static::who();
    }
}

class B extends A {
    public static function who() {
        echo __CLASS__ . "\n";
    }
}

class C extends A {
    public static function who() {
        echo __CLASS__ . "\n";
    }
}

echo "=== Late Static Binding Test ===\n";

echo "\nCalling A::test():\n";
A::test();

echo "\nCalling B::test():\n";
B::test();

echo "\nCalling C::test():\n";
C::test();

echo "\n=== Direct calls for comparison ===\n";
echo "A::who() = ";
A::who();

echo "B::who() = ";
B::who();

echo "C::who() = ";
C::who();