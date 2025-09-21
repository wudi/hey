<?php

class SimpleClass {
    public $value = 10;
}

$original = new SimpleClass();
echo "Before clone: ";
var_dump($original->value);

$cloned = clone $original;
echo "After clone: ";
var_dump($cloned->value);
