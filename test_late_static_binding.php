<?php

class ParentClass {
    public static function who() {
        echo __CLASS__ . "\n";
    }

    public static function test() {
        self::who();   // Always ParentClass
        static::who(); // Late static binding
    }
}

class ChildClass extends ParentClass {
    public static function who() {
        echo __CLASS__ . "\n";
    }
}

echo "=== Testing Late Static Binding ===\n";
echo "ParentClass::test():\n";
ParentClass::test();

echo "\nChildClass::test():\n";
ChildClass::test();