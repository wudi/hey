<?php
// 测试 Match 表达式
$result = match ($x) {
    1 => 'one',
    2, 3 => 'two or three',
    default => 'other'
};

$value = match (true) {
    $a < $b => 'a is smaller',
    $a > $b => 'a is bigger',
    default => 'equal'
};
?>