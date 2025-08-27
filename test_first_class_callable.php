<?php

// Function first-class callables
$func = strlen(...);
$func2 = trim(...);

// Object method first-class callables
$obj = new stdClass();
$method = $obj->method(...);
$method2 = $obj->toString(...);

// Nullsafe object method first-class callables
$method3 = $obj?->method(...);

// Static method first-class callables
$static = MyClass::staticMethod(...);
$static2 = self::method(...);
$static3 = parent::method(...);
$static4 = static::method(...);

// Variable function first-class callables
$funcName = 'strlen';
$variableFunc = $funcName(...);

// Property access method first-class callables
$prop = $obj->prop(...);

class TestClass {
    public function testMethod() {
        // Self method first-class callable
        $self = $this->method(...);
        $selfStatic = self::staticMethod(...);
        $parent = parent::method(...);
        $static = static::method(...);
        
        return [$self, $selfStatic, $parent, $static];
    }
    
    public static function staticMethod() {
        return 'static';
    }
}