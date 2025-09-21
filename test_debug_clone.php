<?php

class SimpleClass {
    public $value = 10;
}

$original = new SimpleClass();
var_dump($original);
echo "About to clone...\n";
// Simple assignment test first
$copy = $original;
var_dump($copy);
echo "Assignment worked\n";