<?php

echo "=== Testing Array Functions ===\n";

// Test array_merge
$arr1 = [1, 2, 3];
$arr2 = [4, 5, 6];
$merged = array_merge($arr1, $arr2);
echo "array_merge([1,2,3], [4,5,6]): " . implode(',', $merged) . "\n";

// Test with multiple arrays
$arr3 = ['a', 'b'];
$merged2 = array_merge($arr1, $arr2, $arr3);
echo "array_merge three arrays: " . implode(',', $merged2) . "\n";

// Test sort
$unsorted = [3, 1, 4, 1, 5, 9, 2, 6];
echo "Before sort: " . implode(',', $unsorted) . "\n";
sort($unsorted);
echo "After sort: " . implode(',', $unsorted) . "\n";

// Test sort with strings
$words = ['zebra', 'apple', 'banana', 'cherry'];
echo "Before sort (strings): " . implode(',', $words) . "\n";
sort($words);
echo "After sort (strings): " . implode(',', $words) . "\n";

echo "=== Array functions test completed ===\n";

?>