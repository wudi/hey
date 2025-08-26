<?php

// Test all previously problematic tokens

// Reference operator
&$variable;

// eval() expression  
eval('echo "test";');

// Class declaration
class TestClass {
    // Visibility modifiers
    public $publicVar;
    private $privateVar;
    protected $protectedVar;
}

// Static access
::staticMethod();
TestClass::staticProperty;

// Switch case/default
switch ($x) {
    case 1:
        break;
    default:
        break;
}

// Const declaration
const MY_CONST = 42;

?>