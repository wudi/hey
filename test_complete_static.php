<?php

class A {
    const NAME = "A";
    const TYPE = "BaseClass";

    public static function whoami() {
        echo "self::NAME = " . self::NAME . "\n";
        echo "static::NAME = " . static::NAME . "\n";
        echo "self::TYPE = " . self::TYPE . "\n";
        echo "static::TYPE = " . static::TYPE . "\n";
    }
}

class B extends A {
    const NAME = "B";
    // TYPE is inherited from A
}

class C extends A {
    const NAME = "C";
    const TYPE = "DerivedClass";
}

echo "=== A::whoami() ===\n";
A::whoami();

echo "\n=== B::whoami() ===\n";
B::whoami();

echo "\n=== C::whoami() ===\n";
C::whoami();