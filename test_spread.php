<?php
$arr1 = [1, 2, 3];
$arr2 = [...$arr1, 4, 5];
$arr3 = array(0, ...$arr1, 6);

// 函数调用中的展开
function test($a, $b, $c) {
    return $a + $b + $c;
}

$result = test(...$arr1);
$mixed = test(1, ...[2, 3]);