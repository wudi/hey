<?php

echo "=== Testing Array Processing Functions ===\n";

// Test array_filter without callback (filter truthy values)
$numbers = [0, 1, 2, 0, 3, false, 4, null, 5];
echo "Original: " . implode(',', $numbers) . "\n";
$filtered = array_filter($numbers);
echo "Filtered: " . implode(',', $filtered) . "\n";

// Test array_values
$assoc = ['a' => 1, 'b' => 2, 'c' => 3];
echo "Associative keys: " . implode(',', array_keys($assoc)) . "\n";
$values_only = array_values($assoc);
echo "Values only: " . implode(',', $values_only) . "\n";

// Test count function
$test_array = [1, 2, 3, 4, 5];
echo "Count array: " . count($test_array) . "\n";
echo "Count string: " . count("hello") . "\n";
echo "Count null: " . count(null) . "\n";

echo "=== Array processing test completed ===\n";

?>