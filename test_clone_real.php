<?php

class SimpleClass {
    public $value = 10;
}

$original = new SimpleClass();
echo "Original:
";
var_dump($original);

echo "About to clone...
";
$clone = clone $original;
echo "Clone:
";
var_dump($clone);

echo "Modifying clone...
";
$clone->value = 20;

echo "Original after clone modification:
";
var_dump($original);
echo "Clone after modification:
";
var_dump($clone);
