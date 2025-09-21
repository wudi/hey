<?php

echo "=== Testing Math Functions ===\n";

// Test abs function
echo "abs(-5): " . abs(-5) . "\n";
echo "abs(5): " . abs(5) . "\n";
echo "abs(-3.7): " . abs(-3.7) . "\n";
echo "abs('--4.5'): " . abs('-4.5') . "\n";

// Test round function
echo "\nround(3.7): " . round(3.7) . "\n";
echo "round(3.2): " . round(3.2) . "\n";
echo "round(3.14159, 2): " . round(3.14159, 2) . "\n";
echo "round(1234.5678, 1): " . round(1234.5678, 1) . "\n";
echo "round(123.456, 0): " . round(123.456, 0) . "\n";

// Test edge cases
echo "\nEdge cases:\n";
echo "abs(0): " . abs(0) . "\n";
echo "round(0.5): " . round(0.5) . "\n";
echo "round(-0.5): " . round(-0.5) . "\n";

echo "\n=== Math test completed ===\n";

?>