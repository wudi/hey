<?php

abstract class A {
    public static function test() {
        echo "Frame info: class=" . __CLASS__ . "\n";
    }

    public static function callStatic() {
        echo "static call from A\n";
    }
}

class B extends A {
    public static function callStatic() {
        echo "static call from B\n";
    }
}

echo "=== Simple Static Test ===\n";

echo "Calling A::test():\n";
A::test();

echo "\nCalling B::test():\n";
B::test();

echo "\nDirect static calls:\n";
echo "A::callStatic(): ";
A::callStatic();
echo "B::callStatic(): ";
B::callStatic();