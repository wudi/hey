<?php

function test_variadic(...$args) {
    echo "Arguments received: " . count($args) . "\n";
    foreach ($args as $i => $arg) {
        echo "Arg $i: $arg\n";
    }
    return array_sum($args);
}

echo "Testing variadic function:\n";
$result = test_variadic(1, 2, 3);
echo "Result: $result\n";