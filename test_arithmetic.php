<?php

echo "=== Arithmetic Operations Test ===\n";

$a = 5;
$b = 2;

echo "a = " . $a . "\n";
echo "b = " . $b . "\n";

echo "a + b = " . ($a + $b) . "\n";
echo "a * b = " . ($a * $b) . "\n";
echo "a *= b: ";
$a *= $b;
echo $a . "\n";

echo "\nTesting with property:\n";
class Test { public $val = 5; }
$obj = new Test();
echo "obj->val = " . $obj->val . "\n";
$obj->val *= 2;
echo "obj->val *= 2: " . $obj->val . "\n";