<?php
echo "Starting test\n";

try {
    echo "In try block\n";
    throw new Exception("Test exception");
    echo "After throw (should not see this)\n";
} catch (Exception $e) {
    echo "Caught exception: " . $e->getMessage() . "\n";
}

echo "Test completed\n";