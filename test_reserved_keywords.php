<?php

class TestClass {
    // Class constants with reserved keywords
    const class = 'class_value';
    const function = 'function_value';
    const if = 'if_value';
    const new = 'new_value';
    public const private = 'private_value';
    
    // Methods with reserved keywords as names
    public function class() {
        return 'class method';
    }
    
    public function function() {
        return 'function method';
    }
    
    public function if() {
        return 'if method';
    }
    
    public function new() {
        return 'new method';
    }
    
    private function const() {
        return 'const method';
    }
    
    protected function echo() {
        return 'echo method';
    }
    
    public function while() {
        return 'while method';
    }
}

$obj = new TestClass();

// Property access with reserved keywords
$result1 = $obj->class();
$result2 = $obj->function();  
$result3 = $obj->if();
$result4 = $obj->new();

// Nullsafe property access with reserved keywords  
$result5 = $obj?->class();
$result6 = $obj?->echo();

// Object property access with reserved keywords (dynamic properties)
$obj->class = 'dynamic_class';
$obj->function = 'dynamic_function';
$obj->if = 'dynamic_if';
$val1 = $obj->class;
$val2 = $obj->function;
$val3 = $obj->if;

trait TestTrait {
    public function class() {
        return 'trait class';
    }
    
    public function if() {
        return 'trait if';
    }
}

class TestTraitUsage {
    use TestTrait {
        // Trait aliases with reserved keywords
        class as function;
        if as while;
        TestTrait::class as public echo;
    }
}