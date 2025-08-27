<?php

// Simple function first-class callable
$func = strlen(...);

// Object method first-class callable  
$obj = new stdClass();
$method = $obj->method(...);

// Static method first-class callable
$static = MyClass::staticMethod(...);

// Variable function first-class callable
$funcName = 'strlen';
$variableFunc = $funcName(...);