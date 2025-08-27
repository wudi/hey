<?php

// Array callback syntax
$callback1 = [$obj, 'method'];
$callback2 = [self::class, 'staticMethod']; 
$callback3 = ['MyClass', 'staticMethod'];

// String callback syntax
$callback4 = 'strlen';
$callback5 = 'MyClass::staticMethod';

// Callable objects
$callable = new class {
    public function __invoke() {
        return 'invoked';
    }
};

// Closure
$closure = function() { return 'closure'; };
$closureFCC = $closure(...);

// Test edge cases with first-class callables
$arrayMethod = [new stdClass(), 'toString'](...);
$complexStatic = \My\Namespace\MyClass::method(...);
$chainedMethod = $obj->getCallback()(...);  // This might not work yet

// Anonymous class methods
$anon = new class {
    public function method() { return 'anon'; }
};
$anonFCC = $anon->method(...);

// Inherited methods
class Parent {
    public function parentMethod() { return 'parent'; }
}

class Child extends Parent {
    public function testCallables() {
        $parent = parent::parentMethod(...);
        $self = self::parentMethod(...);
        $static = static::parentMethod(...);
        return [$parent, $self, $static];
    }
}