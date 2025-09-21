<?php

echo "=== Testing Trim Functions ===\n";

// Test basic trim
$text = "  hello world  ";
echo "Original: '" . $text . "'\n";
echo "trim: '" . trim($text) . "'\n";
echo "ltrim: '" . ltrim($text) . "'\n";
echo "rtrim: '" . rtrim($text) . "'\n";

// Test with different whitespace
$messy = "\t\n  hello  \r\n";
echo "\nMessy string: '" . $messy . "'\n";
echo "trim: '" . trim($messy) . "'\n";

// Test with custom characters
$punctuated = "...hello world!!!";
echo "\nWith punctuation: '" . $punctuated . "'\n";
echo "trim('.!'): '" . trim($punctuated, '.!') . "'\n";

echo "\n=== Trim test completed ===\n";

?>