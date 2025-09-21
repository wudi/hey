<?php

function test_normal($a, $b, $c) {
    echo "Arguments: $a, $b, $c\n";
    return $a + $b + $c;
}

echo "Testing normal function:\n";
$result = test_normal(1, 2, 3);
echo "Result: $result\n";