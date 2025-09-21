<?php
echo "Starting test\n";

try {
    echo "In try block\n";
    echo "Normal execution\n";
} catch (Exception $e) {
    echo "Should not reach here\n";
}

echo "Test completed\n";