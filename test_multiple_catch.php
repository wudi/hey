<?php
class CustomException extends Exception {}

echo "Testing multiple catch blocks\n";

try {
    throw new Exception("Standard exception");
} catch (CustomException $e) {
    echo "Caught CustomException\n";
} catch (Exception $e) {
    echo "Caught Exception\n";
}

echo "Test completed\n";