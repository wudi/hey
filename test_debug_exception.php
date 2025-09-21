<?php
echo "Starting debug test\n";

try {
    echo "In try block\n";
    throw new Exception("Debug exception");
} catch (Exception $e) {
    echo "In catch block\n";
}

echo "Debug test completed\n";