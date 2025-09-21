<?php

class SimpleClass {
    public $value = 10;
}

$original = new SimpleClass();
$temp = $original;  // Simple assignment
var_dump($temp);
