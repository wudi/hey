<?php
// First class callable - 应该是 test(...)
$callable = test(...);

// Spread argument - 应该是 test(...$arr)  
$arr = [1, 2, 3];
$result = test(...$arr);