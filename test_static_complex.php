<?php

class BaseClass {
    protected static $count = 0;

    public static function getCount() {
        return self::$count;
    }

    public static function increment() {
        self::$count++;
        return self::getCount();
    }
}

class ChildClass extends BaseClass {
    public static function doubleIncrement() {
        parent::increment();
        parent::increment();
        return parent::getCount();
    }

    public static function testSelf() {
        return self::doubleIncrement();
    }
}

echo "Testing static method inheritance:\n";
echo "Initial count: " . BaseClass::getCount() . "\n";
echo "After increment: " . BaseClass::increment() . "\n";
echo "Child double increment: " . ChildClass::doubleIncrement() . "\n";
echo "Child test self: " . ChildClass::testSelf() . "\n";
echo "Final count: " . BaseClass::getCount() . "\n";